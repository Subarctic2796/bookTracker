# Book Tracker
A simple cli and web app(?) where I keep track of the books I have read

# TODO
[ ] make work with just a txt
[ ] add sqlite

# Notes
SQL schema:
id|isbn|author|title|series|date started|date ended|time took|reading status

API docs: https://openlibrary.org/dev/docs/api/search
API url:  https://openlibrary.org/search.json

# Make API call
```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const OPEN_LIBRARY_URL = "https://openlibrary.org/search.json"

func main() {
	resp, err := http.Get(OPEN_LIBRARY_URL)
	if err != nil {
		panic(err)
	}
	body, err := io.ReadAll(resp.Body)
	// defer func() { _ = resp.Body.Close() }()
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	_ = json.Indent(&buf, body, "", "    ")
	fmt.Println(buf.String())
}
```

[google book api](https://www.googleapis.com/books/v1/volumes?q=)
