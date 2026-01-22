package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type myCtx struct{}
type ctxValues map[any]any

func main() {
	mainCtx := myCtx{}
	ctx := context.WithValue(context.Background(), mainCtx, ctxValues{"hi": 1})

	db, err := sql.Open("sqlite3", "books.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// if err := CMD.Run(context.Background(), os.Args); err != nil {
	if err := CMD.Run(ctx, os.Args); err != nil {
		log.Fatal(err)
	}
}
