package gerrit

import (
    "encoding/json"
    "net/url"
)

type ReviewLabels struct {
    CodeReview int8 `json:"Code-Review"`
}

type ReviewPost struct {
    Message string `json:"message"`
    Labels *ReviewLabels `json:"labels"`
}

type ReviewReply struct {
    Labels *ReviewLabels `json:"labels"`
}

// post review
// POST /changes/{change-id}/revisions/{revision-id}/review
func (gerrit *Gerrit) PostReview(id string, message string, mark int8) (*ReviewReply, error) {

    change_url := url.URL{Scheme: "https", Host: gerrit.Host, Opaque: "/a/changes/" + id + "/revisions/current/review"}

    review := ReviewPost{
                Message: message,
                Labels: &ReviewLabels{CodeReview: mark},
            }

    json_str, json_err := json.Marshal(review)
    if json_err != nil {
        return nil, json_err
    }

    contents, err := gerrit.post_json(&change_url, json_str)
    if err != nil {
        return nil, err
    }

    var reply ReviewReply
    json_err = json.Unmarshal(contents, &reply)
    if json_err != nil {
        return nil, json_err
    }

    return &reply, nil
}

