package api

import (
    "fmt"
    "bufio"
    "io/ioutil"
    "net/http"
    "net/url"
    "errors"
    "bytes"

    "github.com/sergle/go-http-digest"
)

// API access
type API struct {
    User string
    Password string
    Host string
}

// GET
func (gerrit *API) Fetch_json(get_url *url.URL) ([]byte, error) {
    request, err := http.NewRequest("GET", get_url.String(), nil)
    if err != nil {
        fmt.Printf("NewRequest failed: %s\n", err)
        return nil, err
    }
    // need to re-assign to ensure that URL.Opaque part kept
    request.URL = get_url

    // HTTP digest authentication
    t := digest.NewTransport(gerrit.User, gerrit.Password)

    resp, err := t.RoundTrip(request)
    if err != nil {
            fmt.Printf("RoundTrip Failed: %s\n", err)
            return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        return nil, errors.New("Not OK")
    }

    contents, err := parse_response(resp)
    return contents, nil
}

// POST
func (gerrit *API) Post_json(post_url *url.URL, json []byte) ([]byte, error) {

    request, err := http.NewRequest("POST", post_url.String(), bytes.NewBuffer(json))
    if err != nil {
        return nil, err
    }

    // avoid escaping
    request.URL = post_url
    request.Header.Set("Content-Type", "application/json")

    t := digest.NewTransport(gerrit.User, gerrit.Password)

    resp, err := t.RoundTrip(request)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        return nil, errors.New("Not OK")
    }

    contents, err := parse_response(resp)
    return contents, err
}

// parse response - skip special marker and read all to []byte
func parse_response(resp *http.Response) ([]byte, error) {

    bufio := bufio.NewReader(resp.Body)
    // skip special marker:  )]}'
    _, err := bufio.ReadString('\n')
    if err != nil {
        return nil, err
    }

    contents, err := ioutil.ReadAll(bufio)
    if err != nil {
        return nil, err
    }

    return contents, nil
}
