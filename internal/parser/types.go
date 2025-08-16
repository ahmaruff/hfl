package parser

type Entry struct {
	Date string `json:date`
	Body string `json:body`
}

type Journal struct {
	Entries []Entry
}
