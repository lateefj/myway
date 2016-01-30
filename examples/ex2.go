package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/lateefj/myway/mw"
	_ "github.com/mattn/go-sqlite3" // sqlite3 rocks!
)

const (
	simpleTestTableName = "test_ex2"
)

func dbPath() string {
	path := os.Getenv("DB_PATH")
	if path == "" {
		return ":memory:"
	}
	return path
}

func setupDB(db *sql.DB) error {
	// Database initialization
	log.Printf("Setting up database %s", simpleTestTableName)
	_, err := db.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s(x INT, y INT);", simpleTestTableName))
	return err
}
func init() {
	// Database initialization
	db, err := sql.Open("sqlite3", dbPath())
	if err != nil {
		log.Fatalf("Failed to connect to sqlite3 database with path %s error: %s", dbPath(), err)
		return
	}
	db.SetMaxOpenConns(1)
	setupDB(db)
	mw.AssignDB(db)
}

func main() {
	http.HandleFunc("/", mw.TxHandler(func(tx *sql.Tx, w http.ResponseWriter, r *http.Request) error {
		_, err := tx.Exec(fmt.Sprintf("INSERT INTO %s VALUES(1, 1);", simpleTestTableName))
		if err != nil {
			log.Printf("Failed to exec %s\n", err)
			return err
		}
		var size int
		err = tx.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s;", simpleTestTableName)).Scan(&size)
		if err == nil {
			fmt.Fprintf(w, "%s Hello there size of table %s is %d", r.URL.String(), simpleTestTableName, size)
		}
		return err
	}))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
