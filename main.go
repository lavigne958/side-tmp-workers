package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var (
	logger         = log.Default()
	db     *sql.DB = nil
)

func handleList(resposne http.ResponseWriter, request *http.Request) {
	logger.Println("Handler /list route")
}

func initTables() error {
	logger.Println("Init database tables")

	statement := "create table if not exists tasks (id integer not null primary key, org_name string, );"
	_, err := db.Exec(statement)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	logger.Println("start server")
	var err error
	db, err = sql.Open("sqlite3", "side.db")
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
	err = initTables()
	if err != nil {
		logger.Fatalln("failed to init tables: ", err)
		os.Exit(1)
	}

	server := http.NewServeMux()
	server.HandleFunc("/list", handleList)

	err = http.ListenAndServe(":80", server)
	if err != nil {
		logger.Fatalln("Failed to server requests: ", err)
		os.Exit(1)
	}
}
