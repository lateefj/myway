package mw

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/lateefj/mctest"
	_ "github.com/mattn/go-sqlite3" // sqlite3 rocks!
)

const (
	testDBUrlPath   = "/api/name/:year/:name"
	testMultiDBBase = "/tmp/testmyway"
	testMultiTable  = "test_multi_table"
)

func dbDir() string {
	path := os.Getenv("DB_DIR")
	if path == "" {
		return testMultiDBBase
	}
	return path
}

func dbFullPath(key string) string {
	return fmt.Sprintf("%s/%s.db", dbDir(), key)
}

var dbMap map[string]*sql.DB

func init() {
	dbMap = make(map[string]*sql.DB)
}

// Fun with locks
var dbMapLock sync.RWMutex

func FindDB(key string) (*sql.DB, error) {
	var err error
	dbMapLock.RLock()
	db, exists := dbMap[key]
	dbMapLock.RUnlock()
	if exists {
		log.Printf("Found databse in dbMap with key %s", key)
		return db, err
	}
	os.MkdirAll(dbDir(), 0777)
	path := dbFullPath(key)
	db, err = sql.Open("sqlite3", path)
	if err != nil {
		log.Printf("WTF on open path %s.... error: %s", path, err)
		return nil, err
	}
	log.Printf("Opened database file %s", path)
	db.SetMaxOpenConns(1) // SQLite is single threaded :(
	dbMapLock.Lock()
	dbMap[key] = db
	dbMapLock.Unlock()
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
	db, err := FindDB(year)
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

func yearTableSize(year, tName string) (int, error) {
	var size int
	db, err := FindDB(year)
	if err != nil {
		return -1, err
	}
	err = db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s;", tName)).Scan(&size)
	if err != nil {
		log.Printf("Error query count %s", err)
		return -1, err
	}
	log.Printf("For year %s size is %d", year, size)
	return size, nil
}

func cleanUpYears(years []int) {
	for _, year := range years {
		os.RemoveAll(dbFullPath(fmt.Sprintf("%d", year)))
	}
}
func createTestDB(t *testing.T, year, createSQL, tName string) {
	db, err := FindDB(year)
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("For database year %s running createSQL %s", year, createSQL)
	_, err = db.Exec(createSQL)
	if err != nil {
		t.Fatalf("%s\n Failed to create table %s", createSQL, err)
	}
	s, err := yearTableSize(year, tName)
	if err != nil {
		t.Fatal(err)
	}
	if s != 0 {
		t.Fatalf("Expected 0 rows in table but has %d", s)
	}
}

func TestFindDB(t *testing.T) {
	years := []int{2, 4, 8, 16}
	defer cleanUpYears(years)
	for _, year := range years {
		ys := fmt.Sprintf("%d", year)
		createTestDB(t, ys, "CREATE TABLE IF NOT EXISTS test_table(year INT, i INT);", "test_table")
		db, err := FindDB(ys)
		if err != nil {
			t.Error(err)
		}
		for i := 0; i < year; i++ {
			_, err = db.Exec(fmt.Sprintf("INSERT INTO test_table VALUES(%d, %d)", year, i))
			if err != nil {
				t.Error(err)
			}
		}
	}
	for _, year := range years {
		ys := fmt.Sprintf("%d", year)
		size, err := yearTableSize(ys, "test_table")
		if err != nil {
			t.Error(err)
		}
		if year != size {
			t.Errorf("Expected size of %d but it is %d", year, size)
		}
	}
}

func TestMultiTxHandler(t *testing.T) {
	y := 1970
	year := fmt.Sprintf("%d", y)
	createSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s(year INT, x INT);", testMultiTable)
	createTestDB(t, year, createSQL, testMultiTable)
	//defer cleanUpYears([]int{y})
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/name/%s/Epic", year), nil)
	resp := mctest.NewMockTestResponse(t)
	MultiTxHandler(func(tx *sql.Tx, w http.ResponseWriter, r *http.Request) error {
		log.Printf("Calling insert now...")
		sql := fmt.Sprintf("INSERT INTO %s VALUES(%d, 1);", testMultiTable, y)
		_, err := tx.Exec(sql)
		if err != nil {
			t.Fatalf("Insert failure error: %s \n%s", err, sql)
		}
		return err
	})(resp, req)
	s, err := yearTableSize(year, testMultiTable)
	if err != nil {
		t.Fatal(err)
	}
	if s != 1 {
		t.Fatalf("Expected 1 row in table but has %d", s)
	}
	MultiTxHandler(func(tx *sql.Tx, w http.ResponseWriter, r *http.Request) error {
		_, err := tx.Exec(fmt.Sprintf("INSERT INTO %s VALUES(1, 1);", testMultiTable))
		if err != nil {
			t.Errorf("Failed to exec %s", err)
		}
		return errors.New("Failed expect automagic rollback please")
	})(resp, req)
	// Should be same number of rows
	s, err = yearTableSize(year, testMultiTable)
	if err != nil {
		t.Fatal(err)
	}
	if s != 1 {
		t.Fatalf("Expected 1 row in table but has %d", s)
	}
}
