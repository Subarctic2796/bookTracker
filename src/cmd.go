package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/urfave/cli/v3"
)

func defaultAction(_ context.Context, c *cli.Command) error {
	fmt.Println("TODO:", c.Name)
	fmt.Println("isbn:", c.IsSet("isbn"))
	length := c.Args().Len()
	padding := 1 + int(math.Log10(float64(length)))
	for ix := range length {
		fmt.Printf("arg %*d: '%v'\n", padding, ix, c.Args().Get(ix))
	}
	return nil
}

func requireAuthorTitleOrISBN(c *cli.Command) error {
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

func validStateAction(ctx context.Context, c *cli.Command, s string) error {
	switch strings.ToLower(s) {
	case "none", "reading", "finished", "tbr", "dnf":
		return nil
	default:
		return fmt.Errorf(
			"'%s' is not a valid state for a book, state must be one of 'none' 'reading' 'finished' 'tbr' 'dnf'",
			s)
	}
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
		Action:      validStateAction,
	},
	&cli.TimestampFlag{
		Name:    "started",
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
		Name:    "started",
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
		Name:    "author",
		Aliases: []string{"a"},
		Usage:   "the `author` who wrote the book",
		Action: func(ctx context.Context, c *cli.Command, s string) error {
			if s == "" {
				return fmt.Errorf("author can not be empty")
			}
			return nil
		},
	},
	&cli.StringFlag{
		Name:    "title",
		Aliases: []string{"t"},
		Usage:   "the `title` of the book",
		Action: func(ctx context.Context, c *cli.Command, s string) error {
			if s == "" {
				return fmt.Errorf("title can not be empty")
			}
			return nil
		},
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
		Action:      validStateAction,
	},
	&cli.TimestampFlag{
		Name:    "started",
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
		&cli.BoolFlag{
			Name:    "lookup",
			Aliases: []string{"L"},
			Usage:   "use openlibrary to look up details about a book and add those details to the database",
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
				if err := requireAuthorTitleOrISBN(c); err != nil {
					return err
				}

				isbnSet := c.IsSet("ISBN")
				isbn, title, author := "", c.StringArg("title"), c.StringArg("author")
				if isbnSet {
					if c.IsSet("isbn") {
						localISBN := c.String("isbn")
						if cleanISBN(title) != cleanISBN(localISBN) {
							return fmt.Errorf(
								"isbn was set twice and they do not match: ISBN = '%s' isbn = '%s'",
								title, localISBN)
						}
					}
					isbn = cleanISBN(title)
					title, author = "", ""
				}

				db := ctx.Value(myCtx{}).(*sql.DB)
				// check if the book has already exists
				if isbnSet {
					const QUERY = "SELECT isbn FROM books WHERE isbn LIKE ?"
					rows, err := db.Query(QUERY, isbn)
					if err != nil {
						return err
					}
					defer rows.Close()

					for rows.Next() {
						var isbn string
						err = rows.Scan(&isbn)
						if err != nil {
							return err
						}
						fmt.Println(isbn)
					}
				} else {
					return fmt.Errorf("TODO: look for author and title")
				}

				book := Book{
					ISBN:    isbn,
					Author:  author,
					Title:   title,
					Series:  c.String("series"),
					Status:  BS_READING,
					Started: c.Timestamp("start"),
				}

				genres := c.StringSlice("genres")
				if genres != nil {
					book.Genres = genres
				}

				const QUERY = "INSERT INTO books (isbn, author, title, series, date_started, status, genres) VALUES(?, ?, ?, ?, ?, ?, ?)"
				_, err := db.Exec(QUERY,
					book.ISBN, book.Author, book.Title, book.Series,
					book.Started.Unix(), book.Status, strings.Join(book.Genres, ","))
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
		{
			Name:      "remove",
			Usage:     "remove a book from the database",
			Arguments: commonArgs,
			ArgsUsage: "[[title author]|ISBN]",
			Action:    defaultAction,
		},
		{
			Name:      "search",
			Usage:     "lookup an ISBN number",
			Arguments: []cli.Argument{&cli.StringArg{Name: "isbn"}},
			Action: func(_ context.Context, c *cli.Command) error {
				isbn := c.StringArg("isbn")
				if !validISBN(isbn) {
					return fmt.Errorf("'%s' is an invalid isbn number", isbn)
				}

				fmt.Printf("searching '%s' on openlibrary\n", isbn)
				cleanISBN := strings.ReplaceAll(strings.ReplaceAll(isbn, "-", ""), " ", "")
				query := fmt.Sprintf("https://openlibrary.org/search.json?q=%s&fields=title,author_name", cleanISBN)
				resp, err := http.Get(query)
				if err != nil {
					panic(err)
				}
				fmt.Printf("successfully found '%s'\n", isbn)

				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return err
				}

				var json_ struct {
					Docs []struct {
						Author_names []string `json:"author_name"`
						Title        string   `json:"title"`
					} `json:"docs"`
				}
				if err := json.Unmarshal(body, &json_); err != nil {
					return err
				}

				docs0 := json_.Docs[0]
				fmt.Println("author:", docs0.Author_names)
				fmt.Println("title:", docs0.Title)
				return nil
			},
		},
	},
}
