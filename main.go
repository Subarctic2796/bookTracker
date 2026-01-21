package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v3"
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
	Title    string
	Author   string
	ISBN     string
	State    BookState
	Started  time.Time
	Finished time.Time
	Took     time.Time
	Genres   []string
}

func (b *Book) String() string {
	return fmt.Sprintf(
		"Title   : %s\nAuthor  : %s\nISBN    : %s\nState   : %s\nStarted : %s\nFinished: %s\nTook    : %s\nGenres  : %s",
		b.Title, b.Author, b.ISBN, b.State,
		b.Started, b.Finished, b.Took,
		strings.Join(b.Genres, " "))
}

var defaultAction = func(_ context.Context, c *cli.Command) error {
	lenght := c.Args().Len()
	padding := 1 + int(math.Log10(float64(lenght)))
	for ix := range lenght {
		fmt.Printf("arg %*d: '%v'\n", padding, ix, c.Args().Get(ix))
	}
	return nil
}

var addFlags = []cli.Flag{
	&cli.StringFlag{
		Name:     "title",
		Aliases:  []string{"t"},
		Usage:    "the `title` of the book",
		Required: true,
	},
	&cli.StringFlag{
		Name:     "author",
		Aliases:  []string{"a"},
		Usage:    "the name of the `author`",
		Required: true,
	},
	&cli.StringFlag{
		Name:  "isbn",
		Usage: "the `ISBN` number of the book",
	},
	&cli.StringFlag{
		Name:        "state",
		Aliases:     []string{"st"},
		Value:       "none",
		Usage:       "the `state` of the book, must be one of 'none' 'reading' 'finished' 'tbr' 'dnf'",
		DefaultText: "none",
		Action: func(ctx context.Context, c *cli.Command, s string) error {
			switch strings.ToLower(s) {
			case "none", "reading", "finished", "tbr", "dnf":
				return nil
			default:
				return fmt.Errorf(
					"'%s' is not a valid state for a book, state must be one of 'none' 'reading' 'finished' 'tbr' 'dnf'",
					s)
			}
		},
	},
	&cli.TimestampFlag{
		Name:    "start",
		Aliases: []string{"s"},
		Usage:   "the `date` you started the book",
		Value:   time.Now(),
		Config: cli.TimestampConfig{
			Timezone: time.Local,
			Layouts:  []string{"2006-01-02T15:04:05"},
		},
		DefaultText: "now",
	},
	&cli.TimestampFlag{
		Name:    "finished",
		Aliases: []string{"f"},
		Usage:   "the `date` you finished the book",
		Value:   time.Now(),
		Config: cli.TimestampConfig{
			Timezone: time.Local,
			Layouts:  []string{"2006-01-02T15:04:05"},
		},
		DefaultText: "now",
	},
	&cli.StringFlag{
		Name:    "genres",
		Aliases: []string{"g"},
		Usage:   "a list of comma separated genres `genre1,genre2`",
	},
}

func main() {
	cmd := &cli.Command{
		Name:  "bookTracker",
		Usage: "use to track the books you read",
		Commands: []*cli.Command{
			{
				Name:  "add",
				Usage: "add a new book",
				Flags: addFlags,
				Action: func(ctx context.Context, c *cli.Command) error {
					fmt.Printf("args: %s\n", c.Args())
					fmt.Printf("author: %s\n", c.String("author"))
					fmt.Printf("start: %s\n", c.Timestamp("start"))
					return nil
				},
			},
			{
				Name:   "list",
				Usage:  "list out all of the books in the database",
				Action: defaultAction,
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
