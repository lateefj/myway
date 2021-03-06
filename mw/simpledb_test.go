// This is just an example so only going to write one test Blah!
package mw

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/lateefj/mctest"
	_ "github.com/mattn/go-sqlite3" // sqlite3 rocks!
)

const (
	simpleTestTableName = "test_t"
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
	setupDB(db)
	AssignDB(db)
}

func tableSize() int64 {
	size := int64(0)
	db := CurrentDB()
	db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s;", simpleTestTableName)).Scan(&size)
	return size
}

// Contrived example but yeah this is the idea
func TestTxHandler(t *testing.T) {
	db := CurrentDB()
	setupDB(db)
	db.Exec(fmt.Sprintf("DELETE FROM %s;", simpleTestTableName))
	if tableSize() != 0 {
		t.Fatalf("Expected 0 rows in table but has %d", tableSize())
	}
	req, _ := http.NewRequest("GET", "/path/to/handler", nil)
	resp := mctest.NewMockTestResponse(t)

	TxHandler(func(tx *sql.Tx, w http.ResponseWriter, r *http.Request) error {
		log.Printf("Inserting row ....")
		_, err := tx.Exec(fmt.Sprintf("INSERT INTO %s VALUES(1, 1);", simpleTestTableName))
		if err != nil {
			t.Errorf("Failed to exec %s", err)
		}
		return err
	})(resp, req)
	if tableSize() != 1 {
		t.Errorf("Expected 1 rows in table but has %d", tableSize())
	}
	TxHandler(func(tx *sql.Tx, w http.ResponseWriter, r *http.Request) error {
		_, err := tx.Exec(fmt.Sprintf("INSERT INTO %s VALUES(1, 1);", simpleTestTableName))
		if err != nil {
			t.Errorf("Failed to exec %s", err)
		}
		return errors.New("Failed expect automagic rollback please")
	})(resp, req)
	// Should be same number of rows
	if tableSize() != 1 {
		t.Errorf("Expected 1 rows in table but has %d", tableSize())
	}

}
