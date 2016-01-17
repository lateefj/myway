package mw

import (
	"database/sql"
	"log"
	"net/http"
)

// database/sql has a built in connection pool so we just need a single instance
var myDB *sql.DB

// AssignDB set the database connection to use
func AssignDB(db *sql.DB) {
	if myDB != nil { // Don't judge...
		myDB.Close()
	}
	myDB = db
}

// CurrentDB expose the database connection to external packages
func CurrentDB() *sql.DB {
	return myDB
}

// DBHandlerFunc defines a function type (yeah, Go is cool like that)
type DBHandlerFunc func(*sql.DB, http.ResponseWriter, *http.Request)

// DBHandler no magic here just takes a function type that expects a database connection (plus http.Handlerfunc)
// The important par tis it returns a regular http.Handlerfunc so it is compatable with the Go standard library and any library / framework that is worth a damn thing.
func DBHandler(fn DBHandlerFunc) http.HandlerFunc {
	// Make sure the connection is still good
	err := myDB.Ping()
	if err != nil {
		// TODO: Error handling code in a real production application
		log.Printf("Database connection failed %s", err)
	}

	// This returns a regular
	return func(w http.ResponseWriter, r *http.Request) {
		fn(myDB, w, r)
	}
}

// TxHandlerFunc most of the time just want a transaction
type TxHandlerFunc func(*sql.Tx, http.ResponseWriter, *http.Request) error

func TxWrapper(db *sql.DB, w http.ResponseWriter, r *http.Request, fn TxHandlerFunc) {

	tx, err := myDB.Begin()
	// Default behavior is to commit the transaction
	if err != nil {
		log.Printf("Transcation failed to begin %s\n", err)
	}
	txErr := fn(tx, w, r)
	if txErr == nil {
		tx.Commit()
		log.Printf("Successful commit")
	} else {
		tx.Rollback()
		log.Printf("Transaction rolling back because of error: %s\n", txErr)
	}
}

// TxHandler double down by using the wrapper we already built wrap
func TxHandler(fn TxHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		DBHandler(func(db *sql.DB, w http.ResponseWriter, r *http.Request) {
			TxWrapper(db, w, r, fn)
		})(w, r)
	}
}
