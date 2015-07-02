package gerrit

import (
    "github.com/sergle/go-gerrit/api"
    "github.com/sergle/go-gerrit/change"
    "github.com/sergle/go-gerrit/review"
)

type Gerrit struct {
    api *api.API
}

//----------- public api -------------

func New(user string, password string, host string, ci string) *Gerrit {
     api := &api.API{
                User:     user,
                Password: password,
                Host:     host,
                CI:       ci,
            }
    return &Gerrit{api: api}
}

// get list of changes according to the query string
func (gerrit *Gerrit) FetchChangeList(query_string string) ([]change.ShortChange, error) {
    return change.FetchList(gerrit.api, query_string)
}

// fetch detailed information about the change
func (gerrit *Gerrit) GetChange(id string) (*change.LongChange, error) {
    return change.Get(gerrit.api, id)
}

// get mark which set by our Continuous Integration tool
func (gerrit *Gerrit) IsVerified(chg *change.LongChange) (string, int8) {
    return change.IsVerified(gerrit.api, chg)
}

// post review message
func (gerrit *Gerrit) PostReview(id string, message string, mark int8) (*review.ReviewReply, error) {
    return review.Post(gerrit.api, id, message, mark)
}

// sort list of changes by Updated field (recently-updated first)
func (gerrit *Gerrit) SortChanges(list []*change.LongChange) {
    change.Sort(list)
    return
}

// is this username used for API access
func (gerrit *Gerrit) IsMyself(username string) bool {
    if username == gerrit.api.User {
        return true
    }
    return false
}

// it this username used by Continuous Integration tool
func (gerrit *Gerrit) IsCI(username string) bool {
    if username == gerrit.api.CI {
        return true
    }
    return false
}
