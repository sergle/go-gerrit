package main

import (
        "log"
        "time"
        "flag"
        // profile
        "net/http"
        _ "net/http/pprof"
        g "github.com/sergle/go-gerrit"
        "github.com/sergle/go-gerrit/change"
)

const ConfigFile = "gerrit.conf"

var (
    gerrit *g.Gerrit
    last_updated map[string]string
)

func process_change_list(list []*change.LongChange, verbose bool) {
        ready := 0
        updated := 0
        for _, ch := range list {

            if verbose {
                print_change(ch)
            }

            _, verified := gerrit.IsVerified(ch)

            // no jenkins mark - ignore
            if verified != 1 {
                if verbose {
                    log.Printf("SKIP: not verified")
                }
                continue
            }

            // merge conflict - ignore
            if !ch.Mergeable {
                if verbose {
                    log.Printf("SKIP: cannot be merged")
                }
                continue
            }

            all_review := true
            has_negative := false
            already_approved := false
            is_reviewer := false
            for _, p := range ch.Labels.CodeReview.All {
                if gerrit.IsMyself(p.Username) {
                    is_reviewer = true

                    if p.Value == 2 {
                        already_approved = true
                    }
                }

                if p.Value == -1 || p.Value == -2 || p.Value == 0 {
                    if p.Value == 0 {
                        if p.Username == ch.Owner.Username {
                            // owner itself
                            continue
                        } else if gerrit.IsCI(p.Username) {
                            // CI agent
                            continue
                        }
                    }

                    all_review = false
                    //break
                    if p.Value == -1 || p.Value == -2 {
                        has_negative = true
                    }
                }

                // +1, +2 is ok
            }

            if !is_reviewer {
                log.Printf("NOT IN REVIEW LIST")
                continue
            }

            if already_approved {
                if verbose {
                    log.Printf("SKIP: already +2")
                }
                continue
            }

            // not all reviewers argee
            if has_negative {
                if verbose {
                    log.Printf("SKIP: has negative review")
                }
                continue
            }

            // not all reviewers
            if !all_review {
                if verbose {
                    log.Printf("SKIP: not all reviewed")
                }
                continue
            }

            // print it
            ready++

            // if verbose - already printed
            if !verbose {
                print_change(ch)
            }

            msg := "Gerrit-Bot: auto-approve (all reviewers +1)"
            log.Printf("ALL OK: %s", msg)
            reply, err := gerrit.PostReview(ch.Id, msg, 2)
            if err != nil {
                log.Printf("Post failed: %s", err)
                return
            }
            log.Printf("Posted review: +%d", reply.Labels.CodeReview)
            updated++
        }

        log.Printf("Ready: %d, Reviewed: %d", ready, updated)
}

func print_change(change *change.LongChange) {

    log.Printf("Change ID: %s", change.Id)
    log.Printf("  Project: %s", change.Project)
    log.Printf("  Branch: %s", change.Branch)
    log.Printf("  Owner: %s", change.Owner.Username)
    log.Printf("  Subject: %s", change.Subject)
    log.Printf("  Updated: %s", change.Updated[0:16])

    ci_username, verified := gerrit.IsVerified(change)

    if ci_username != "" {
        log.Printf("  Verified: %s : %d", ci_username, verified)
    }

    for _, rv := range change.Labels.CodeReview.All {
        log.Printf("  Review: %s : %d", rv.Username, rv.Value)
    }
}

func get_change(id string, ch chan<- *change.LongChange) () {

    detail, err := gerrit.GetChange(id)
    if err != nil {
        log.Printf("Failed to fetch change: %s\n", id)
    }

    ch <- detail
    return
}

func dashboard(query_string string, project_list []string, verbose bool) {

    var change_list = make([]change.ShortChange, 0)

    for _, project := range project_list {
        list, err := gerrit.FetchChangeList(query_string + "+p:" + project)
        if err != nil {
            return
        }
        if verbose {
            log.Printf("Project %s has %d changes\n", project, len(list))
        }
        if len(list) > 0 {
            change_list = append(change_list, list...)
        }
    }

    log.Printf("Total %d changes\n", len(change_list))
    if len(change_list) == 0 {
        return
    }

    ch := make(chan *change.LongChange, len(change_list))

    existing_id := make(map[string]bool)

    cnt_updated := 0
    for _, change := range change_list {

        existing_id[ change.Id ] = true

        last_upd, ok := last_updated[ change.Id ]
        if ! ok {
            log.Printf("New change %s - %s\n", change.Id, change.Subject)
        } else if last_upd == change.Updated {
            if verbose {
                log.Printf("Change %s not updated\n", change.Id)
            }
            continue
        } else {
            log.Printf("Change %s updated at %s\n", change.Id, change.Updated)
        }

        // read details only for new or updated changes
        last_updated[ change.Id] = change.Updated
        cnt_updated++

        go get_change(change.Id, ch)
    }

    cnt_deleted := cleanup_deleted(existing_id)

    log.Printf("Updated: %d, Deleted: %d\n", cnt_updated, cnt_deleted)
    if cnt_updated == 0 {
        return
    }

    var processed = 0
    var ch_list = []*change.LongChange{}

Loop:
    for {
        select {
            case change := <-ch:
                processed++
                ch_list = append(ch_list, change)
                if processed == cnt_updated {
                    break Loop
                }
            }
    }

    process_change_list(ch_list, verbose)

    return
}

func cleanup_deleted(existing map[string]bool) int {
    cnt := 0
    for id := range last_updated {
        _, ok := existing[id]
        if ! ok {
            log.Printf("Change %s was deleted\n", id)
            delete(last_updated, id)
            cnt++
        }
    }

    return cnt
}

func parse_args() (bool, string) {
    var verbose = flag.Bool("v", false, "Verbose output")
    var conf_file = flag.String("f", ConfigFile, "Path to config file")
    flag.Parse()

    return *verbose, *conf_file
}

func main() {
    verbose, conf_file := parse_args()

    cfg, err := ReadConfig(conf_file)
    if err != nil {
        log.Fatalf("Error reading config %s - %s\n", conf_file, err)
    }


    profile := cfg.Robot.Profile

    ticker, err := time.ParseDuration(cfg.Robot.Interval)
    if err != nil {
        log.Fatalf("Failed to parse [Robot]Interval: %s", err)
    }

    gerrit = g.New(cfg.Gerrit.User, cfg.Gerrit.Password, cfg.Gerrit.Host, cfg.Gerrit.CI)
    last_updated = make(map[string]string)

    if profile != "" {
        // start profile server
        go func() {
            log.Println(http.ListenAndServe(profile, nil))
        }()
        log.Printf("Started profiling server on %s", profile)
    }

    project_list := cfg.Robot.Project
    //query := "?q=is:reviewer+status:open+-owner:self+p:{PROJECT}"
    query := "?q=is:reviewer+status:open+-owner:self"
    dashboard(query, project_list, verbose)

    ticker_chan := time.NewTicker(ticker).C

    for {
        select {
        case _ = <-ticker_chan:
            log.Println("Tick")
            dashboard(query, project_list, verbose)
        }
    }
}
