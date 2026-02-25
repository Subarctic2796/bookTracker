package main

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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
	Took     time.Duration
}

// make sure to reset b4 using
var CASER = cases.Title(language.Und)

func (b *Book) String() string {
	CASER.Reset()
	var sb strings.Builder

	fmt.Fprintf(&sb, "Title   : %s\n", CASER.String(b.Title))
	CASER.Reset()

	fmt.Fprintf(&sb, "Series  : %s\n", CASER.String(b.Series))
	CASER.Reset()

	fmt.Fprintf(&sb, "Author  : %s\n", CASER.String(b.Author))
	CASER.Reset()

	fmt.Fprintf(&sb, "ISBN    : %s\n", b.ISBN)
	fmt.Fprintf(&sb, "Status  : %s\n", b.Status)
	fmt.Fprintf(&sb, "Started : %s\n", b.Started)
	fmt.Fprintf(&sb, "Finished: %s\n", b.Finished)
	fmt.Fprintf(&sb, "Took    : %s\n", b.Took)
	fmt.Fprintf(&sb, "Genres  : %s", strings.Join(b.Genres, ", "))
	return sb.String()
}

func cleanISBN(isbn string) string {
	return strings.ReplaceAll(strings.ReplaceAll(isbn, "-", ""), " ", "")
}

func validISBN(isbn string) bool {
	if len(isbn) < 10 {
		return false
	}
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
