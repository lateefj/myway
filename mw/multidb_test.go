package mw

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/lateefj/mctest"
	_ "github.com/mattn/go-sqlite3" // sqlite3 rocks!
)

const (
	testDBUrlPath   = "/api/name/:year/:name"
	testMultiDBBase = "/tmp/testmyway"
)

func dbDir() string {
	path := os.Getenv("DB_DIR")
	if path == "" {
		return testMultiDBBase
	}
	return path
}

var databaseDB map[string]*sql.DB

func findDB(key string) (*sql.DB, error) {
	if databaseDB == nil {
		databaseDB = make(map[string]*sql.DB)
	}
	var err error
	db, exists := databaseDB[key]
	if !exists {
		path := fmt.Sprintf("%s/%s.db", dbDir(), key)
		db, err = sql.Open("sqlite3", path)
		databaseDB[key] = db
		setupDB(db)
	}
	return db, err
}

type testDBConnect struct {
}

func (tdbc *testDBConnect) Open(w http.ResponseWriter, r *http.Request) (*sql.DB, error) {
	router := httprouter.New()
	router.GET(testDBUrlPath, func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	})
	_, params, _ := router.Lookup("GET", r.URL.Path)
	year := params.ByName("year")
	log.Printf("Ok rout path is %s and year is %s", r.URL.Path, year)
	if year == "" {
		return nil, errors.New("Failed to find year in path")
	}
	db, err := findDB(year)
	if err != nil {
		log.Printf("Failed to find database for year %s", year)
	}
	return db, err
}

var mytdbc DBConnect

func init() {
	os.MkdirAll(testMultiDBBase, 0777)

	mytdbc = &testDBConnect{}
	AssignDBConnect(mytdbc)
}

func yearTableSize(year string) int64 {
	size := int64(0)
	db, err := findDB(year)
	if err != nil {
		return int64(-1)
	}
	db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s;", tableName)).Scan(&size)
	return size
}
func cleanUpDBDir() {
	os.RemoveAll(testMultiDBBase)
}
func TestMultiTxHandler(t *testing.T) {
	//defer cleanUpDBDir()
	year := "1970"
	if yearTableSize(year) != 0 {
		t.Errorf("Expected 0 rows in table but has %d", tableSize())
	}
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/name/%s/Epic", year), nil)
	resp := mctest.NewMockTestResponse(t)
	MultiTxHandler(func(tx *sql.Tx, w http.ResponseWriter, r *http.Request) error {
		log.Printf("Calling insert now...")
		_, err := tx.Exec(fmt.Sprintf("INSERT INTO %s VALUES(1, 1);", tableName))
		if err != nil {
			t.Errorf("Failed to exec insert %s", err)
		}
		return err
	})(resp, req)
	if yearTableSize(year) != 1 {
		t.Errorf("Expected 1 row in table but has %d", yearTableSize(year))
	}
	MultiTxHandler(func(tx *sql.Tx, w http.ResponseWriter, r *http.Request) error {
		_, err := tx.Exec(fmt.Sprintf("INSERT INTO %s VALUES(1, 1);", tableName))
		if err != nil {
			t.Errorf("Failed to exec %s", err)
		}
		return errors.New("Failed expect automagic rollback please")
	})(resp, req)
	// Should be same number of rows
	if yearTableSize(year) != 1 {
		t.Errorf("Expected 1 rows in table but has %d", yearTableSize(year))
	}

}
