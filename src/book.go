package main

import (
	"fmt"
	"strings"
	"time"
)

type BookState byte

const (
	BS_NONE BookState = iota
	BS_READING
	BS_FINISHED
	BS_TBR
	BS_DNF
)

func (s BookState) String() string {
	return [...]string{"NONE", "READING", "FINISHED", "TBR", "DNF"}[s]
}

type Book struct {
	ISBN     string
	Author   string
	Title    string
	Series   string
	State    BookState
	Started  time.Time
	Finished time.Time
	Took     time.Time
	Genres   []string
}

func (b *Book) String() string {
	return fmt.Sprintf(
		"Title   : %s\nSeries:  %s\nAuthor  : %s\nISBN    : %s\nState   : %s\nStarted : %s\nFinished: %s\nTook    : %s\nGenres  : %s",
		b.Title, b.Series, b.Author, b.ISBN, b.State,
		b.Started, b.Finished, b.Took,
		strings.Join(b.Genres, " "))
}
