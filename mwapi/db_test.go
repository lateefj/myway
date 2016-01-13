// This is just an example so only going to write one test Blah!
package mwapi

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/lateefj/mctest"
)

const (
	tableName = "test_t"
)

func init() {
	db, _ := SetupDB()
	db.Exec(fmt.Sprintf("CREATE TABLE %s(x INT, y INT);", tableName))
}

func tableSize() int64 {
	size := int64(0)
	db, _ := SetupDB()
	db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s;", tableName)).Scan(&size)
	return size
}

// Contrived example but yeah this is the idea
func TestTxHandler(t *testing.T) {
	if tableSize() != 0 {
		t.Errorf("Expected 0 rows in table but has %d", tableSize())
	}
	req, _ := http.NewRequest("GET", "/path/to/handler", nil)
	resp := mctest.NewMockTestResponse(t)

	TxHandler(func(tx *sql.Tx, w http.ResponseWriter, r *http.Request) error {
		_, err := tx.Exec(fmt.Sprintf("INSERT INTO %s VALUES(1, 1);", tableName))
		if err != nil {
			t.Errorf("Failed to exec %s", err)
		}
		return err
	})(resp, req)
	if tableSize() != 1 {
		t.Errorf("Expected 1 rows in table but has %d", tableSize())
	}
	TxHandler(func(tx *sql.Tx, w http.ResponseWriter, r *http.Request) error {
		_, err := tx.Exec(fmt.Sprintf("INSERT INTO %s VALUES(1, 1);", tableName))
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
