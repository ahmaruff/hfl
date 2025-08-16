package parser

type Entry struct {
	Date string
	Body string
}

type Journal struct {
	Entries []Entry
}
