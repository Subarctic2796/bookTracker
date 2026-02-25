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
	title, author := c.StringArg("title"), c.StringArg("author")
	if c.Bool("ISBN") {
		if validISBN(title) {
			return nil
		}
		if author != "" {
			return fmt.Errorf("author must not be set if using ISBN mode")
		}
		return fmt.Errorf("'%s' is not a valid ISBN number", title)
	}
	authorNotSet, titleNotSet := author == "", title == ""
	if authorNotSet && titleNotSet {
		return fmt.Errorf("title and author must be provided or use '-I ISBN'")
	} else if titleNotSet {
		return fmt.Errorf("title must be provided or use '-I ISBN'")
	} else if authorNotSet {
		return fmt.Errorf("author must be provided or use '-I ISBN'")
	}
	return nil
}

// returns isbnSet, title, author, isbn, error in that order
// It zeros title and author if ISBN is set
func determineTitleAuthorISBNAndISBNisSet(c *cli.Command) (bool, string, string, string, error) {
	isbnSet := c.IsSet("ISBN")
	title, author := c.StringArg("title"), c.StringArg("author")
	if isbnSet {
		if !c.IsSet("isbn") {
			return isbnSet, "", "", cleanISBN(title), nil
		}
		localISBN := c.String("isbn")
		if cleanISBN(title) != cleanISBN(localISBN) {
			return isbnSet, "", "", "", fmt.Errorf(
				"isbn was set twice and they do not match: ISBN = '%s' isbn = '%s'",
				title, localISBN)
		}
	}
	if c.IsSet("isbn") && !validISBN(c.String("isbn")) {
		return isbnSet, "", "", "", fmt.Errorf("'%s' is not a valid ISBN number", c.String("isbn"))

	}
	return isbnSet, title, author, cleanISBN(c.String("isbn")), nil
}

func validStateAction(_ context.Context, c *cli.Command, s string) error {
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

var (
	isbnFlag = &cli.StringFlag{
		Name:  "isbn",
		Usage: "the `ISBN` number of the book",
		Value: "",
	}
	authorFlag = &cli.StringFlag{
		Name:    "author",
		Aliases: []string{"a"},
		Usage:   "the `author` who wrote the book",
		Action: func(ctx context.Context, c *cli.Command, s string) error {
			if s == "" {
				return fmt.Errorf("author can not be empty")
			}
			return nil
		},
	}
	titleFlag = &cli.StringFlag{
		Name:    "title",
		Aliases: []string{"t"},
		Usage:   "the `title` of the book",
		Action: func(ctx context.Context, c *cli.Command, s string) error {
			if s == "" {
				return fmt.Errorf("title can not be empty")
			}
			return nil
		},
	}
	seriesFlag = &cli.StringFlag{
		Name:    "series",
		Aliases: []string{"se"},
		Usage:   "the name of the `series` the book belongs to",
		Value:   "",
	}
	stateFlag = &cli.StringFlag{
		Name:        "state",
		Aliases:     []string{"st"},
		Value:       "none",
		Usage:       "the `state` of the book, must be one of 'none' 'reading' 'finished' 'tbr' 'dnf'",
		DefaultText: "none",
		Action:      validStateAction,
	}
	startedFlag = &cli.TimestampFlag{
		Name:    "started",
		Aliases: []string{"s"},
		Usage:   "the `date` you started the book",
		Value:   time.Now(),
		Config: cli.TimestampConfig{
			Timezone: time.Local,
			Layouts:  []string{"2006-01-02T15:04:05"},
		},
		DefaultText: "now",
	}
	finishedFlag = &cli.TimestampFlag{
		Name:    "finished",
		Aliases: []string{"f"},
		Usage:   "the `date` you finished the book",
		Value:   time.Now(),
		Config: cli.TimestampConfig{
			Timezone: time.Local,
			Layouts:  []string{"2006-01-02T15:04:05"},
		},
		DefaultText: "now",
	}
	genresFlag = &cli.StringSliceFlag{
		Name:    "genres",
		Aliases: []string{"g"},
		Usage:   "a list of comma separated genres `genre1,genre2`",
		Value:   nil,
	}
)

var addFlags = []cli.Flag{
	isbnFlag,
	seriesFlag,
	stateFlag,
	startedFlag,
	finishedFlag,
	genresFlag,
}

var startFlags = []cli.Flag{
	isbnFlag,
	seriesFlag,
	startedFlag,
	genresFlag,
}

var finishFlags = []cli.Flag{
	isbnFlag,
	finishedFlag,
	stateFlag,
}

var updateFlags = []cli.Flag{
	isbnFlag,
	seriesFlag,
	stateFlag,
	startedFlag,
	finishedFlag,
	genresFlag,
	authorFlag,
	titleFlag,
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
			Usage:   "switches to ISBN mode, where only the isbn number can be used and not a title author pair",
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
				fmt.Printf("started: %s\n", c.Timestamp("started"))
				fmt.Printf("genres: %s\n", c.StringSlice("genres"))
				return nil
			},
		},
		{
			Name:      "finish",
			Usage:     "finish a book that you started",
			Arguments: commonArgs,
			ArgsUsage: "[[title author]|ISBN]",
			Flags:     finishFlags,
			Action: func(ctx context.Context, c *cli.Command) error {
				if err := requireAuthorTitleOrISBN(c); err != nil {
					return err
				}

				isbnSet, title, author, isbn, err := determineTitleAuthorISBNAndISBNisSet(c)
				if err != nil {
					return err
				}

				state := BS_FINISHED
				if c.IsSet("state") {
					switch stateStr := c.String("state"); stateStr {
					case "none", "tbr", "reading":
						return fmt.Errorf("the status of a book you are finishing can not be '%s'", stateStr)
					case "finished":
						state = BS_FINISHED
					case "dnf":
						state = BS_DNF
					}
				}

				db := ctx.Value(myCtx{}).(*sql.DB)
				// check if the book already exists
				if isbnSet {
					QUERY := "SELECT EXISTS(SELECT 1 FROM books WHERE isbn = ?)"
					row := db.QueryRow(QUERY, isbn)
					if err != nil {
						return err
					}
					var exists int
					err = row.Scan(&exists)
					if err != nil {
						return err
					}

					if exists != 1 {
						return fmt.Errorf("book with isbn of '%s' does not exist", isbn)
					}

					QUERY = "UPDATE books SET status = ?, date_finished = ? WHERE isbn = ?"
					_, err = db.Exec(QUERY, state, c.Timestamp("finished").Unix(), isbn)
					if err != nil {
						return err
					}
				} else {
					QUERY := "SELECT EXISTS(SELECT 1 FROM books WHERE title = ? AND author = ?)"
					row := db.QueryRow(QUERY, strings.ToLower(title), strings.ToLower(author))
					var exists int
					err = row.Scan(&exists)
					if err != nil {
						return err
					}

					if exists != 1 {
						return fmt.Errorf("book with title: '%s' and author '%s' does not exist", title, author)
					}

					QUERY = "UPDATE books SET status = ?, date_finished = ? WHERE title = ? AND author = ?"
					_, err = db.Exec(QUERY, state, c.Timestamp("finished").Unix(), strings.ToLower(title), strings.ToLower(author))
					if err != nil {
						return err
					}
				}

				return nil
			},
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

				isbnSet, title, author, isbn, err := determineTitleAuthorISBNAndISBNisSet(c)
				if err != nil {
					return err
				}

				db := ctx.Value(myCtx{}).(*sql.DB)
				// check if the book already exists
				if isbnSet {
					const QUERY = "SELECT EXISTS(SELECT 1 FROM books WHERE isbn = ?)"
					row := db.QueryRow(QUERY, isbn)
					if err != nil {
						return err
					}
					var exists int
					err = row.Scan(&exists)
					if err != nil {
						return err
					}

					if exists == 1 {
						return fmt.Errorf("book with isbn of '%s' already exists", isbn)
					}
				} else {
					const QUERY = "SELECT EXISTS(SELECT 1 FROM books WHERE title = ? AND author = ?)"
					row := db.QueryRow(QUERY, strings.ToLower(title), strings.ToLower(author))
					var exists int
					err = row.Scan(&exists)
					if err != nil {
						return err
					}

					if exists == 1 {
						return fmt.Errorf("book with title: '%s' and author '%s' already exists", title, author)
					}
				}

				book := Book{
					ISBN:    isbn,
					Author:  strings.ToLower(author),
					Title:   strings.ToLower(title),
					Series:  strings.ToLower(c.String("series")),
					Status:  BS_READING,
					Started: c.Timestamp("started"),
				}

				genres := c.StringSlice("genres")
				if genres != nil {
					for ix, i := range genres {
						genres[ix] = strings.ToLower(i)
					}
					book.Genres = genres
				}

				const QUERY = "INSERT INTO books (isbn, author, title, series, date_started, status, genres) VALUES(?, ?, ?, ?, ?, ?, ?)"
				_, err = db.Exec(QUERY,
					book.ISBN, book.Author, book.Title, book.Series,
					book.Started.Unix(), book.Status, strings.Join(book.Genres, ","))
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			Name:  "list",
			Usage: "list out all of the books in the database",
			Action: func(ctx context.Context, c *cli.Command) error {
				db := ctx.Value(myCtx{}).(*sql.DB)

				const QUERY = "SELECT * FROM books"
				rows, err := db.Query(QUERY)
				if err != nil {
					return err
				}
				defer rows.Close()

				for rows.Next() {
					var id, status int
					var date_started, date_finished sql.NullInt64
					var isbn, title, author, series, genres string
					err = rows.Scan(&id, &isbn, &author, &title, &series, &date_started, &date_finished, &status, &genres)
					if err != nil {
						return err
					}

					started := time.Time{}
					if date_started.Valid {
						started = time.Unix(date_started.Int64, 0).Local()
					}
					finished := time.Time{}
					if date_finished.Valid {
						finished = time.Unix(date_finished.Int64, 0).Local()
					}

					book := Book{
						ISBN:     isbn,
						Author:   author,
						Title:    title,
						Series:   series,
						Started:  started,
						Finished: finished,
						Status:   BookState(status),
						Genres:   strings.Split(genres, ","),
						Took:     finished.Sub(started),
					}
					fmt.Println(book.String())
					fmt.Println()
				}

				return nil
			},
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
