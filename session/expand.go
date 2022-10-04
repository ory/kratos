package session

// Expandable controls what fields to expand for projects.
type Expandable string

// Expandables is a list of Expandable values.
type Expandables []Expandable

// String returns a string representation of the Expandable.
func (e Expandable) String() string {
	return string(e)
}

// ToEager returns the fields used by pop's Eager command.
func (e Expandables) ToEager() []string {
	var s []string
	for _, e := range e {
		s = append(s, e.String())
	}
	return s
}

// Has returns true if the Expandable is in the list.
func (e Expandables) Has(search Expandable) bool {
	for _, e := range e {
		if e == search {
			return true
		}
	}
	return false
}

const (
	ExpandSessionDevices Expandable = "Devices"
)

// ExpandNothing expands nothing
var ExpandNothing []Expandable

// ExpandEverything expands all the fields of a project.
var ExpandEverything = Expandables{
	ExpandSessionDevices,
}
