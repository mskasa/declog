package decision

// Decision represents a parsed architectural decision record.
type Decision struct {
	ID     int
	Title  string
	Date   string
	Status string
	Author string
	File   string
}
