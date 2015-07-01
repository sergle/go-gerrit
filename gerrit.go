package gerrit

import (
    "encoding/json"
    "net/url"
)

const NotVerified = -100

// API access
type Gerrit struct {
    User string
    Password string
    Host string
    CI string
}

type ReviewPoint struct {
    Name string
    Mark int
}

type Change struct {
    Id string
    Project string
    Branch string
    Owner string
    Subject string
    Updated string
    Verified int
    Mergeable bool
    Review []ReviewPoint
}

type ChangeOwner struct {
    Name string `json:"name"`
}

// data returned by in list
type ShortChange struct {
    Id string `json:"id"`
    Project string `json:"project"`
    Branch string `json:"branch"`
    ChangeID string `json:"change_id"`
    Subject string `json:"subject"`
    Status string `json:"status"`
    Created string `json:"created"`
    Updated string `json:"updated"`
    Mergeable bool `json:"mergeable"`
    // insertions, deletions, _sortkey, _number
    Owner *ChangeOwner `json:"owner"`
}

type LongChange struct {
    Id string `json:"id"`
    Project string `json:"project"`
    Branch string `json:"branch"`
    ChangeID string `json:"change_id"`
    Subject string `json:"subject"`
    Status string `json:"status"`
    Created string `json:"created"`
    Updated string `json:"updated"`
    Mergeable bool `json:"mergeable"`
    Owner *Person `json:"owner"`
    Labels *ChangeLabels `json:"labels"`
    // permitted_labels
    // removable_reviewers []
    // messages []
}

type Person struct {
    Name string `json:"name"`
    Email string `json:"email"`
    Username string `json:"username"`
    Value int8 `json:"value,omitempty"`
}

type ChangeLabels struct {
    Verified struct {
        Approved *Person `json:"approved"`
        All []Person `json:"all"`
        // values: {}
        Values map[string]string `json:"values"`
        DefaultValue int8 `json:"default_value"`
    } `json:"verified"`
    CodeReview struct {
        Approved *Person `json:"approved"`
        All []Person `json:"all"`
        Values map[string]string `json:"values"`
        DefaultValue int8 `json:"default_value"`
    } `json:"Code-Review"`
}

//----------- public api -------------

func New(user string, password string, host string, ci string) *Gerrit {
    return &Gerrit{
                User:     user,
                Password: password,
                Host:     host,
                CI:       ci,
            }
}

// get list of changes according to the query string
func (gerrit *Gerrit) FetchChangeList(query_string string) ([]ShortChange, error) {
    list_url, _ := url.Parse("https://" + gerrit.Host + "/a/changes/" + query_string)

    contents, err := gerrit.fetch_json(list_url)
    if err != nil {
        return nil, err
    }

    var data []ShortChange
    err = json.Unmarshal(contents, &data)

    if err != nil {
        //fmt.Printf("JSON failed: %s\n", err)
        //fmt.Printf("JSON data is: %s\n", contents)
        return nil, err
    }

    return data, nil
}

// fetch detailed information about the change
func (gerrit *Gerrit) GetChange(id string) (*LongChange, error) {

    change_url := url.URL{Scheme: "https", Host: gerrit.Host, Opaque: "/a/changes/" + id + "/detail/"}

    contents, err := gerrit.fetch_json(&change_url)
    if err != nil {
        return nil, err
    }

    var data LongChange
    err = json.Unmarshal(contents, &data)
    if err != nil {
        return nil, err
    }

    return &data, nil
}

// get mark which set by our Continuous Integration tool
func (gerrit *Gerrit) IsVerified(change *LongChange) (int8) {
    var verified int8 = -100
    verified = NotVerified
    for _, p := range change.Labels.Verified.All {
        if p.Username == gerrit.CI {
            verified = p.Value
            break
        }
    }

    return verified
}
