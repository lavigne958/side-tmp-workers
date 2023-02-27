package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var (
	logger = log.Default()
)

func handleList(resposne http.ResponseWriter, request *http.Request) {
	logger.Println("Handler /list route")
}

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

	server := http.NewServeMux()
	server.HandleFunc("/list", handleList)

	err = http.ListenAndServe(":80", server)
	if err != nil {
		logger.Fatalln("Failed to server requests: ", err)
		os.Exit(1)
	}
}
