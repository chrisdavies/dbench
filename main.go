package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Task struct {
	Id     string
	Output string
	Status string
	Cmd    string
	Args   []string
}

func genId() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func simulate(fn func(action string, t *Task)) {
	t := &Task{
		Id:     genId(),
		Status: "pending",
		Cmd:    "example",
		Output: "",
		Args:   []string{"a", "b", "c"},
	}

	fn("new", t)
	t.Status = "processing"
	fn("status", t)
	for i := 0; i < 10; i++ {
		t.Output = strconv.Itoa(i) + "Some kind of output and whatever"
		fn("out", t)
	}
	fn("rm", t)
}

func logPanic(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func makeSQLSimulator() (*sql.DB, func(action string, t *Task)) {
	fmt.Println("Opening...")
	db, err := sql.Open("sqlite3", "./tmp.db?_timeout=5000&_journal=WAL&_sync=1")
	if err != nil {
		log.Panic(err)
	}

	fmt.Println("Migrations...")
	_, err = db.Exec(`
	  BEGIN;
		CREATE TABLE IF NOT EXISTS tasks (
			id      varchar(16) NOT NULL PRIMARY KEY,
			status  varchar(16) NOT NULL,
			cmd     varchar(16) NOT NULL,
			args    text NOT NULL,
			output  text NULL
		);
		CREATE INDEX IF NOT EXISTS task_status_idx ON tasks (status);
		COMMIT;
	`)
	logPanic(err)

	return db, func(action string, t *Task) {
		switch action {
		case "new":
			args, err := json.Marshal(t.Args)
			logPanic(err)
			_, err = db.Exec(`INSERT INTO tasks
				(id, status, cmd, args)
			VALUES
			  (@id, @status, @cmd, @args)
			`, sql.Named("id", t.Id), sql.Named("status", t.Status), sql.Named("cmd", t.Cmd), sql.Named("args", string(args)))
			logPanic(err)
		case "status":
			_, err = db.Exec(`UPDATE tasks SET status=@status WHERE id=@id`, sql.Named("id", t.Id), sql.Named("status", t.Status))
			logPanic(err)
		case "out":
			_, err = db.Exec(`UPDATE tasks SET output=@out WHERE id=@id`, sql.Named("id", t.Id), sql.Named("out", t.Output))
			logPanic(err)
		case "rm":
			_, err = db.Exec(`DELETE FROM tasks WHERE id=@id`, sql.Named("id", t.Id))
			logPanic(err)
		}
	}
}

func fileSimulator(action string, t *Task) {
	dir := "./tmp"
	if action == "rm" {
		logPanic(os.Remove(path.Join(dir, t.Id)))
		return
	}

	b, err := json.Marshal(t)
	if err != nil {
		logPanic(err)
	}

	filename := path.Join(dir, t.Id)
	tmpfile := filename + ".tmp"

	logPanic(ioutil.WriteFile(tmpfile, b, 0777))
	logPanic(os.Rename(tmpfile, filename))
}

func simulateN(name string, concurrency, num int, handler func(action string, t *Task)) {
	wg := new(sync.WaitGroup)
	start := time.Now()
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			for i := 0; i < num; i++ {
				simulate(handler)
			}
		}()
	}
	wg.Wait()
	fmt.Println(name, "ran", num*concurrency, "tasks in ", time.Since(start).String())
}

func main() {
	os.Mkdir("./tmp", 0777)
	db, sqlHandler := makeSQLSimulator()
	defer db.Close()
	simulateN("SQL", 100, 100, sqlHandler)
	simulateN("SQL", 100, 100, sqlHandler)
	simulateN("SQL", 100, 100, sqlHandler)
	simulateN("SQL", 100, 100, sqlHandler)
	simulateN("SQL", 100, 100, sqlHandler)
	simulateN("SQL", 100, 100, sqlHandler)
}
