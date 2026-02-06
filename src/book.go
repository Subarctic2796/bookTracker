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
	Started  time.Time
	Finished time.Time
	Status   BookState
	Genres   []string
	Took     time.Time
}

func (b *Book) String() string {
	return fmt.Sprintf(
		"Title   : %s\nSeries:  %s\nAuthor  : %s\nISBN    : %s\nStatus  : %s\nStarted : %s\nFinished: %s\nTook    : %s\nGenres  : %s",
		b.Title, b.Series, b.Author, b.ISBN, b.Status,
		b.Started, b.Finished, b.Took,
		strings.Join(b.Genres, " "))
}

func cleanISBN(isbn string) string {
	return strings.ReplaceAll(strings.ReplaceAll(isbn, "-", ""), " ", "")
}

func validISBN(isbn string) bool {
	data := make([]int, 0, 13)
	for _, i := range isbn {
		if (i >= '0' && i <= '9') || (i == 'x' || i == 'X') {
			if i == 'x' || i == 'X' {
				data = append(data, 10)
			} else {
				data = append(data, int(i-'0'))
			}
		}
	}
	return validate10(data) || validate13(data)
}

func validate10(isbn []int) bool {
	if len(isbn) != 10 {
		return false
	}
	sum := 0
	for ix, i := range isbn {
		sum += (10 - ix) * i
	}
	return sum%11 == 0
}

func validate13(isbn []int) bool {
	if len(isbn) != 13 {
		return false
	}
	sum := 0
	for ix, i := range isbn {
		if ix%2 == 0 {
			sum += i
		} else {
			sum += 3 * i
		}
	}
	return sum%10 == 0
}
