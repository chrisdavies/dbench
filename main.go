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
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func genId() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

// Task is a basic struct. Realistically, we'd be storing a more complex
// structure, but I think this will be sufficient for a basic test.
type Task struct {
	Id   string
	Cmd  string
	Args []string
}

func writeN(name string, num int, fn func(id string, js []byte) error) {
	start := time.Now()

	for i := 0; i < num; i++ {
		t := &Task{
			Id:   genId(),
			Cmd:  "example",
			Args: []string{"a", "b", "c"},
		}
		b, err := json.Marshal(t)
		if err != nil {
			log.Panic(err)
		}
		if err = fn(t.Id, b); err != nil {
			log.Panic(err)
		}
	}
	fmt.Println(name, "wrote", num, "tasks in ", time.Since(start).String())
}

func writeFiles(num int) {
	dir := "./tmp"
	// Don't really care about errors for testing... we'll know if it fails
	os.Mkdir(dir, 0777)
	writeN("Files", num, func(id string, js []byte) error {
		return ioutil.WriteFile(path.Join(dir, id), js, 0777)
	})
}

func writeSql(num int) {
	fmt.Println("Opening...")
	db, err := sql.Open("sqlite3", "./tmp.db?_timeout=5000&_journal=WAL&_sync=1")
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	fmt.Println("Migrations...")
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id    varchar(16) NOT NULL PRIMARY KEY,
			task  text NOT NULL
		)
	`)

	if err != nil {
		log.Panic(err)
	}

	w, err := db.Prepare(`
		INSERT INTO tasks
			(id, task)
		VALUES
			(@id, @task)
	`)
	if err != nil {
		log.Panic(err)
	}
	defer w.Close()

	writeN("SQLite", num, func(id string, js []byte) error {
		_, err := w.Exec(sql.Named("id", id), sql.Named("task", string(js)))
		return err
	})
}

func main() {
	writeSql(10000)
	writeFiles(10000)
	writeSql(10000)
	writeFiles(10000)
	writeSql(10000)
	writeFiles(10000)
}
