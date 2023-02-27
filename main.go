package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type item struct {
	Id           uint64 `json:"id"`
	Name         string `json:"name"`
	Organisation string `json:"organisation"`
}

type items struct {
	Items []item `json:"items"`
}

var (
	logger         = log.Default()
	db     *sql.DB = nil
)

const (
	TASK_TABLE_NAME string = "tasks"
)

func handleList(resposne http.ResponseWriter, request *http.Request) {
	logger.Println("Handler /list route")

	if request.Method != http.MethodGet {
		msg := "/list only accepts GET requests"
		writeErrorResponse(resposne, 400, msg)
		return
	}

	if request.Header.Get("accepts") != "application/json" {
		msg := "/list only accepts 'application/json'"
		writeErrorResponse(resposne, 400, msg)
		return
	}

	resposne.Header().Add("content-type", "application/json")

	stmt := fmt.Sprintf("select id, name, organisation from %s;", TASK_TABLE_NAME)
	res, err := db.Query(stmt)
	if err != nil {
		msg := "failed to get tasks list from DB"
		writeErrorResponse(resposne, 503, msg)
		return
	}

	defer res.Close()

	items := items{}
	for res.Next() {
		item := item{}
		err = res.Scan(&item.Id, &item.Name, &item.Organisation)
		if err != nil {
			msg := fmt.Sprintf("failed to extra item from database: %v", err)
			writeErrorResponse(resposne, 503, msg)
			return
		}

		items.Items = append(items.Items, item)
	}

	buffer, _ := json.Marshal(&items)
	resposne.WriteHeader(200)
	resposne.Write(buffer)
}

func writeErrorResponse(w http.ResponseWriter, code int, msg string) {
	logger.Println(msg)
	w.WriteHeader(code)
	w.Write([]byte(msg))
}

func handleAdd(response http.ResponseWriter, request *http.Request) {
	logger.Println("Add new task")

	if request.Method != http.MethodPost {
		msg := "/add only accepts POST requests"
		writeErrorResponse(response, 400, msg)
		return
	}

	if request.Header.Get("content-type") != "application/json" {
		msg := "/add only accepts 'applicaton/json' content-type"
		writeErrorResponse(response, 400, msg)
		return
	}

	if request.ContentLength <= 0 {
		msg := "/add request must be >0"
		writeErrorResponse(response, 409, msg)
		return
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		msg := fmt.Sprintf("failed to read full body of length: %d", request.ContentLength)
		writeErrorResponse(response, 503, msg)
		return
	}

	item := item{}
	json.Unmarshal(body, &item)

	logger.Println("add new object: ", item)
	stmt := fmt.Sprintf(
		"insert into %s(name, organisation) "+
			"values ('%s', '%s')",
		TASK_TABLE_NAME, item.Name, item.Organisation,
	)
	logger.Println("exec: ", stmt)
	_, err = db.Exec(stmt)
	if err != nil {
		msg := fmt.Sprintf("failed to add task: %v", err)
		writeErrorResponse(response, 503, msg)
		return
	}

	response.WriteHeader(200)
	response.Write([]byte("OK"))
}

func initTables() error {
	logger.Println("Init database tables")

	statement := fmt.Sprintf(
		"create table if not exists %s (id integer not null primary key autoincrement,name string, organisation string);",
		TASK_TABLE_NAME,
	)
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
	}

	defer db.Close()

	res := db.QueryRow("SELECT SQLITE_VERSION()")
	if res.Err() != nil {
		logger.Fatalln("failed to open DB: ", res.Err())
	}
	var version string
	res.Scan(&version)

	logger.Println("Sqlite version: ", version)
	err = initTables()
	if err != nil {
		logger.Fatalln("failed to init tables: ", err)
	}

	server := http.NewServeMux()
	server.HandleFunc("/list", handleList)
	server.HandleFunc("/add", handleAdd)

	err = http.ListenAndServe(":80", server)
	if err != nil {
		logger.Fatalln("Failed to server requests: ", err)
	}
}
