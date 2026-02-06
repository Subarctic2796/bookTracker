package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func initDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "books.db")
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat("books.db"); errors.Is(err, os.ErrNotExist) {
		fmt.Println("INFO: 'books.db' does not exist, creating 'books.db'")
	}

	const CREATESCHEMAQUERY = `CREATE TABLE IF NOT EXISTS books (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		isbn TEXT,
		author TEXT NOT NULL,
		title TEXT NOT NULL,
		series TEXT,
		date_started INTEGER,
		date_finishied INTEGER,
		status INTEGER NOT NULL DEFAULT 0,
		genres TEXT
	);`

	_, err = db.Exec(CREATESCHEMAQUERY)
	if err != nil {
		return nil, err
	}
	return db, nil
}

type ctxValue byte

const (
	cv_db ctxValue = iota
	cv_CNT
)

type myCtx struct{}
type ctxValues [cv_CNT]any

func main() {
	db, err := initDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	mainCtx := myCtx{}
	ctx := context.WithValue(context.Background(), mainCtx, ctxValues{db})

	// if err := CMD.Run(context.Background(), os.Args); err != nil {
	if err := CMD.Run(ctx, os.Args); err != nil {
		log.Fatal(err)
	}
}
