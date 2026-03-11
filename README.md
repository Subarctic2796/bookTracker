# Book Tracker
A simple cli and web app(?) where I keep track of the books I have read

# Usage
```sh
git clone --depth=1 https://github.com/Subarctic2796/bookTracker.git
cd bookTracker
go build -o bookTracker src/*.go
./bookTracker --help
```
You can optionally add it to your path.
It will automatically create a new sqlite database at `$XDG_DATA_HOME/bookTracker/books.db` or `$HOME/.local/share/bookTracker/books.db`
You can change this by creating `$XDG_CONFIG_HOME/bookTracker/bookTracker.conf` or `$HOME/.config/bookTracker/bookTracker.conf`
and adding the line `db_path = /path/to/database`

# TODO
[ ] add sqlite
	[ ] add
	[x] finish
	[x] start
	[ ] list
	[ ] update
	[x] remove
	[ ]
[ ] add creating db to `$HOME/.local/share/bookTracker/books.db`
[ ] add checking config in `$HOME/.config/bookTracker`
[ ] add windows support

# Notes
SQL schema:
id|isbn|author|title|series|date started|date ended|reading status|genres

# Dev Notes
good omens isbn: 057504800X
rhythm of war isbn: 9780765326386

Openlibrary
[API docs](https://openlibrary.org/dev/docs/api/search)
[API url](https://openlibrary.org/search.json)

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
