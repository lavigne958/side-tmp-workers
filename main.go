package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var (
	logger = log.Default()
)

func main() {
	logger.Println("start server")
	db, err := sql.Open("sqlite3", "side.db")
	if err != nil {
		logger.Fatalln("failed to open db 'side.db'")
		os.Exit(1)
	}

	defer db.Close()

	res := db.QueryRow("SELECT SQLITE_VERSION()")
	if res.Err() != nil {
		logger.Fatalln("failed to open DB: ", res.Err())
		os.Exit(1)
	}
	var version string
	res.Scan(&version)

	logger.Println("Sqlite version: ", version)
}
