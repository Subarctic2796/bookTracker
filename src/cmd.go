package main

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/urfave/cli/v3"
)

var defaultAction = func(_ context.Context, c *cli.Command) error {
	fmt.Println("TODO:", c.Name)
	fmt.Println("isbn:", c.IsSet("isbn"))
	lenght := c.Args().Len()
	padding := 1 + int(math.Log10(float64(lenght)))
	for ix := range lenght {
		fmt.Printf("arg %*d: '%v'\n", padding, ix, c.Args().Get(ix))
	}
	return nil
}

var requireAuthorTitleOrISBN = func(c *cli.Command) error {
	// TODO: allow using ISBN number instead of author and title

	title := c.StringArg("title")
	if c.Bool("ISBN") {
		if validISBN(title) {
			return nil
		}
		return fmt.Errorf("'%s' is not a valid ISBN number", title)
	}
	authorNotSet, titleNotSet := c.StringArg("author") == "", title == ""
	if authorNotSet && titleNotSet {
		return fmt.Errorf("title and author must be provided or use '-I ISBN'")
	} else if titleNotSet {
		return fmt.Errorf("title must be provided or use '-I ISBN'")
	} else if authorNotSet {
		return fmt.Errorf("author must be provided or use '-I ISBN'")
	}
	return nil
}

var commonArgs = []cli.Argument{
	&cli.StringArg{Name: "title"},
	&cli.StringArg{Name: "author"},
}

var addFlags = []cli.Flag{
	&cli.StringFlag{
		Name:  "isbn",
		Usage: "the `ISBN` number of the book",
		Value: "",
	},
	&cli.StringFlag{
		Name:    "series",
		Aliases: []string{"se"},
		Usage:   "the name of the `series` the book belongs to",
		Value:   "",
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
	&cli.StringSliceFlag{
		Name:    "genres",
		Aliases: []string{"g"},
		Usage:   "a list of comma separated genres `genre1,genre2`",
		Value:   nil,
	},
}

var startFlags = []cli.Flag{
	&cli.StringFlag{
		Name:  "isbn",
		Usage: "the `ISBN` number of the book",
		Value: "",
	},
	&cli.StringFlag{
		Name:    "series",
		Aliases: []string{"se"},
		Usage:   "the name of the `series` the book belongs to",
		Value:   "",
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
	&cli.StringSliceFlag{
		Name:    "genres",
		Aliases: []string{"g"},
		Usage:   "a list of comma separated genres `genre1,genre2`",
		Value:   nil,
	},
}

var updateFlags = []cli.Flag{
	&cli.StringFlag{
		Name:  "isbn",
		Usage: "the `ISBN` number of the book",
		Value: "",
	},
	&cli.StringFlag{
		Name:    "series",
		Aliases: []string{"se"},
		Usage:   "the name of the `series` the book belongs to",
		Value:   "",
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
	&cli.StringSliceFlag{
		Name:    "genres",
		Aliases: []string{"g"},
		Usage:   "a list of comma separated genres `genre1,genre2`",
		Value:   nil,
	},
}

// TODO: at the moment we build a Book obj and then write it to the db
// do we want to maybe just write it to the db straight
var CMD = &cli.Command{
	Name:  "bookTracker",
	Usage: "track your books locally",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "ISBN",
			Aliases: []string{"I"},
			Usage:   "use ISBN instead of title and author pair",
		},
	},
	Commands: []*cli.Command{
		{
			Name:      "add",
			Usage:     "add a new book",
			Arguments: commonArgs,
			ArgsUsage: "[[title author]|ISBN]",
			Flags:     addFlags,
			Action: func(ctx context.Context, c *cli.Command) error {
				c.IsSet("isbn")
				if err := requireAuthorTitleOrISBN(c); err != nil {
					return err
				}
				fmt.Printf("args: %s\n", c.Args())
				fmt.Printf("start: %s\n", c.Timestamp("start"))
				fmt.Printf("genres: %s\n", c.StringSlice("genres"))
				return nil
			},
		},
		{
			Name:      "finish",
			Usage:     "finish a book that you started",
			ArgsUsage: "[[title author]|ISBN]",
			Arguments: commonArgs,
			Action:    defaultAction,
		},
		{
			Name:      "start",
			Usage:     "start a book",
			Arguments: commonArgs,
			ArgsUsage: "[[title author]|ISBN]",
			Flags:     startFlags,
			Action: func(ctx context.Context, c *cli.Command) error {
				// fmt.Printf("TODO: '%s'\n", c.Name)
				if err := requireAuthorTitleOrISBN(c); err != nil {
					return err
				}

				// for ix, i := range c.Flags {
				// 	fmt.Printf("%d: %s: %v\n", ix, i.Names(), i.Get())
				// }

				book := Book{
					ISBN:    c.String("isbn"),
					Author:  c.StringArg("author"),
					Title:   c.StringArg("title"),
					Series:  c.String("series"),
					Status:  BS_READING,
					Started: c.Timestamp("start"),
				}
				genres := c.StringSlice("genres")
				if genres != nil {
					book.Genres = genres
				}
				db := ctx.Value(myCtx{}).(ctxValues)[cv_db].(*sql.DB)
				query := "INSERT INTO books (isbn, author, title, series, date_started, status, genres) VALUES(?, ?, ?, ?, ?, ?, ?)"
				_, err := db.Exec(query, book.ISBN, book.Author, book.Title, book.Series, book.Started.Unix(), book.Status, strings.Join(book.Genres, ","))
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			Name:   "list",
			Usage:  "list out all of the books in the database",
			Action: defaultAction,
		},
		{
			Name:      "update",
			Usage:     "update info about a book",
			Arguments: commonArgs,
			ArgsUsage: "[[title author]|ISBN]",
			Flags:     updateFlags,
			Action:    defaultAction,
		},
	},
}
