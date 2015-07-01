package gerrit

import (
    "sort"
)

// ByUpdated implements sort.Interface for []Change based on
// the Updated field.
type ByUpdated []*LongChange
func (a ByUpdated) Len() int           { return len(a) }
func (a ByUpdated) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByUpdated) Less(i, j int) bool { return a[i].Updated > a[j].Updated }

// sort list of changes by Updated field (recently-updated first)
func (gerrit *Gerrit) SortChanges(list []*LongChange) {
    sort.Sort( ByUpdated(list) )
    return
}
