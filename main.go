package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type taskStatus string

type item struct {
	Id               uint64     `json:"id"`
	Name             string     `json:"name"`
	Organisation     string     `json:"organisation"`
	NrSlots          uint32     `json:"slots-total"`
	NrSlotsAvailable uint32     `json:"slots-available"`
	NrSlotsFilled    uint32     `json:"slots-filled"`
	NrApplicants     uint32     `json:"applicants"`
	Status           taskStatus `json:"status"`
	Picture          string     `json:"picture"`
	Ops              ops_member `json:"ops"`
}

type items struct {
	Items []item `json:"items"`
}

type ops_member struct {
	Id   uint64 `json:"id"`
	Name string `json:"name"`
}

type ops_task_assignee struct {
	TaskId uint64 `json:"task-id"`
	OpsId  uint64 `json:"ops-id"`
}

var (
	logger         = log.Default()
	db     *sql.DB = nil
	reset          = flag.Bool("reset-db", false, "reset the database at init")
)

const (
	TASK_TABLE_NAME      string = "tasks"
	OPS_TABLE_NAME       string = "ops_members"
	OPS_TASKS_TABLE_NAME string = "ops_tasks"

	STATUS_UPCOMING taskStatus = "upcoming"
	STATUS_ONGOING  taskStatus = "ongoing"
	STATUS_DONE     taskStatus = "done"
)

func handleList(resposne http.ResponseWriter, request *http.Request) {
	logger.Println("Handler /list route")

	if request.Method != http.MethodGet {
		msg := "/list only accepts GET requests"
		writeErrorResponse(resposne, 400, msg)
		return
	}

	if request.Header.Get("accept") != "application/json" {
		msg := "/list only accepts 'application/json'"
		writeErrorResponse(resposne, 400, msg)
		return
	}

	resposne.Header().Add("content-type", "application/json")

	stmt := fmt.Sprintf("select id, name, organisation, slots, available, filled, applicants, status, picture from %s", TASK_TABLE_NAME)

	query := request.URL.Query()
	if query.Has("status") {
		status_filter := query.Get("status")
		stmt = fmt.Sprintf("%s where status='%s'", stmt, status_filter)
	}

	// don't forget to add trailling ;
	stmt += ";"

	logger.Println(stmt)
	res, err := db.Query(stmt)
	if err != nil {
		msg := fmt.Sprintf("failed to get tasks list from DB: %v", err)
		writeErrorResponse(resposne, 503, msg)
		return
	}

	defer res.Close()

	items := items{[]item{}}
	for res.Next() {
		item := item{}
		err = res.Scan(
			&item.Id,
			&item.Name,
			&item.Organisation,
			&item.NrSlots,
			&item.NrSlotsAvailable,
			&item.NrSlotsFilled,
			&item.NrApplicants,
			&item.Status,
			&item.Picture,
		)
		if err != nil {
			msg := fmt.Sprintf("failed to extra item from database: %v", err)
			writeErrorResponse(resposne, 503, msg)
			return
		}

		ops_task_row := db.QueryRow(
			fmt.Sprintf(
				"select task_id, ops_id from %s where task_id=%d",
				OPS_TASKS_TABLE_NAME, item.Id,
			),
		)

		ops_task := ops_task_assignee{}
		if ops_task_row.Scan(&ops_task.TaskId, &ops_task.OpsId) == nil {
			logger.Println("found a matching ops: ", ops_task)
			ops_row := db.QueryRow(
				fmt.Sprintf(
					"select id, name from %s where id=%d",
					OPS_TABLE_NAME, ops_task.OpsId,
				),
			)

			ops := ops_member{}
			if ops_row.Scan(&ops.Id, &ops.Name) == nil {
				logger.Println("found ops: ", ops)
				item.Ops = ops
			}
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

func handleTaskAdd(response http.ResponseWriter, request *http.Request) {
	logger.Println("Add new task")

	if request.Method != http.MethodPost {
		msg := "/task/add only accepts POST requests"
		writeErrorResponse(response, 400, msg)
		return
	}

	if request.Header.Get("content-type") != "application/json" {
		msg := "/task/add only accepts 'applicaton/json' content-type"
		writeErrorResponse(response, 400, msg)
		return
	}

	if request.ContentLength <= 0 {
		msg := "/task/add request must be >0"
		writeErrorResponse(response, 409, msg)
		return
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		msg := fmt.Sprintf("failed to read full body of length: %d", request.ContentLength)
		writeErrorResponse(response, 503, msg)
		return
	}

	item := item{}
	json.Unmarshal(body, &item)

	logger.Println("add new object: ", item)
	stmt := fmt.Sprintf(
		"insert into %s(name, organisation, slots, available, filled, applicants, status, picture) "+
			"values ('%s', '%s', %d, %d, %d, %d, '%s', '%s')",
		TASK_TABLE_NAME, item.Name, item.Organisation,
		item.NrSlots, item.NrSlotsAvailable, item.NrSlotsFilled,
		item.NrApplicants, item.Status, item.Picture,
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

func handleOpsAdd(w http.ResponseWriter, req *http.Request) {
	logger.Println("add new ops memeber")

	if req.Method != http.MethodPost {
		msg := "/ops/add only accepts POST requests"
		writeErrorResponse(w, 400, msg)
		return
	}

	if req.Header.Get("content-type") != "application/json" {
		msg := "/ops/add only accepts 'applicaton/json' content-type"
		writeErrorResponse(w, 400, msg)
		return
	}

	if req.ContentLength <= 0 {
		msg := "/ops/add request must be >0"
		writeErrorResponse(w, 409, msg)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		msg := fmt.Sprintf("failed to read full body of length: %d", req.ContentLength)
		writeErrorResponse(w, 503, msg)
		return
	}

	ops := ops_member{}
	json.Unmarshal(body, &ops)

	logger.Println("add new ops member: ", ops)

	stmt := fmt.Sprintf(`
		insert into %s(name) values ("%s");
	`, OPS_TABLE_NAME, ops.Name)
	_, err = db.Exec(stmt)
	if err != nil {
		msg := fmt.Sprintf("failed to insert ops member: %v", err)
		writeErrorResponse(w, 503, msg)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("ok"))
}

func handleAssign(w http.ResponseWriter, req *http.Request) {
	logger.Println("add new ops memeber")

	if req.Method != http.MethodPost {
		msg := "/assign only accepts POST requests"
		writeErrorResponse(w, 400, msg)
		return
	}

	if req.Header.Get("content-type") != "application/json" {
		msg := "/assign only accepts 'applicaton/json' content-type"
		writeErrorResponse(w, 400, msg)
		return
	}

	if req.ContentLength <= 0 {
		msg := "/assign request must be >0"
		writeErrorResponse(w, 409, msg)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		msg := fmt.Sprintf("failed to read full body of length: %d", req.ContentLength)
		writeErrorResponse(w, 503, msg)
		return
	}

	ops_task := ops_task_assignee{}
	json.Unmarshal(body, &ops_task)

	logger.Println("assign ops ", ops_task.OpsId, " to task ", ops_task.TaskId)

	stmt := fmt.Sprintf(`
		insert into %s(task_id, ops_id)
		values (%d, %d)
	`, OPS_TASKS_TABLE_NAME, ops_task.TaskId, ops_task.OpsId)

	_, err = db.Exec(stmt)
	if err != nil {
		msg := fmt.Sprintf("failed to assign ops to task: %v", err)
		writeErrorResponse(w, 503, msg)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func initTables() error {
	logger.Println("Init database tables")

	if *reset {
		logger.Println("reset table")
		_, err := db.Exec(fmt.Sprintf("drop table if exists %s;", TASK_TABLE_NAME))
		if err != nil {
			logger.Fatalln("failed to drop table at init: ", err)
		}

		_, err = db.Exec(fmt.Sprintf("drop table if exists %s;", OPS_TABLE_NAME))
		if err != nil {
			logger.Fatalln("failed to drop table at init: ", err)
		}

		_, err = db.Exec(fmt.Sprintf("drop table if exists %s;", OPS_TASKS_TABLE_NAME))
		if err != nil {
			logger.Fatalln("failed to drop table at init: ", err)
		}
	}

	statement := fmt.Sprintf(
		`
		create table if not exists %s
			(
				id integer not null primary key autoincrement,
				name string,
				organisation string,
				slots integer,
				available integer,
				filled integer,
				applicants integer,
				status string,
				picture string
			);
		`,
		TASK_TABLE_NAME,
	)
	_, err := db.Exec(statement)
	if err != nil {
		return err
	}

	statement = fmt.Sprintf(`
		create table if not exists %s (
			id integer not null primary key autoincrement,
			name string
		);
	`, OPS_TABLE_NAME)
	_, err = db.Exec(statement)
	if err != nil {
		return err
	}

	statement = fmt.Sprintf(`
		create table if not exists %s (
			task_id integer not null,
			ops_id integer not null
		);
	`, OPS_TASKS_TABLE_NAME)
	_, err = db.Exec(statement)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	logger.Println("start server")
	flag.Parse()

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
	server.HandleFunc("/task/add", handleTaskAdd)
	server.HandleFunc("/ops/add", handleOpsAdd)
	server.HandleFunc("/assign", handleAssign)

	err = http.ListenAndServe(":80", server)
	if err != nil {
		logger.Fatalln("Failed to server requests: ", err)
	}
}
