package main

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path"

	_ "github.com/mattn/go-sqlite3"
)

type config struct {
	dbPath string
}

func ReadConfigFile(path string) (config, error) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		panic("TODO")
	}

	f, err := os.ReadFile(path)
	if err != nil {
		return config{}, err
	}

	dbPath := ""
	for line := range bytes.Lines(f) {
		parts := bytes.Split(bytes.TrimSpace(line), []byte{'='})
		key := bytes.TrimSpace(parts[0])
		value := bytes.TrimSpace(parts[1])
		if string(key) == "db_path" {
			dbPath = string(value)
			break
		}
	}

	if dbPath == "" {
		return config{}, fmt.Errorf("[malformed config file]: file must contain `db_path = /path/to/database`")
	}

	return config{dbPath}, nil
}

func GetConfig() (config, error) {
	xdg_config_home, err := os.UserConfigDir()
	if err != nil {
		return config{}, err
	}

	path := path.Join(xdg_config_home, "bookTracker", "bookTracker.conf")
	return ReadConfigFile(path)
}

func initDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(dbPath); errors.Is(err, os.ErrNotExist) {
		fmt.Println("INFO: 'books.db' does not exist, creating 'books.db'")
	}

	const CREATESCHEMAQUERY = `CREATE TABLE IF NOT EXISTS books (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		isbn TEXT,
		author TEXT NOT NULL,
		title TEXT NOT NULL,
		series TEXT,
		date_started INTEGER,
		date_finished INTEGER,
		status INTEGER NOT NULL DEFAULT 0,
		genres TEXT
	);`

	_, err = db.Exec(CREATESCHEMAQUERY)
	if err != nil {
		return nil, err
	}
	return db, nil
}

type myCtx struct{}

func main() {
	// config, err := getConfig()
	// if err != nil {
	// 	fmt.Fprintln(os.Stderr, err)
	// 	os.Exit(1)
	// }
	config := config{"books.db"}

	db, err := initDB(config.dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	mainCtx := myCtx{}
	ctx := context.WithValue(context.Background(), mainCtx, db)

	if err := CMD.Run(ctx, os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
