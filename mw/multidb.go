package mw

import (
	"database/sql"
	"log"
	"net/http"
)

// DBConnect is a simple interface that can look up a database based on the http request
type DBConnect interface {
	Open(http.ResponseWriter, *http.Request) (*sql.DB, error)
}

// myDBConnect the single instance
var myDBConnect DBConnect

// AssignDBConnect one time call
func AssignDBConnect(dbc DBConnect) {
	myDBConnect = dbc
}

// CurrentDBConnect return the existing way to get database connection
func CurrentDBConnect() DBConnect {
	return myDBConnect
}

// MultiDBHandler is a wrapper that will call Open on the DBConnect instance of the interface for every request
func MultiDBHandler(fn DBHandlerFunc) http.HandlerFunc {
	// This returns a regular
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := myDBConnect.Open(w, r)
		if err != nil {
			log.Printf("TODO: Error handling for database connection failure....")
			return
		}
		fn(c, w, r)
	}
}

// MultiTxHandler is the transaction version that just wrapps both MultiDBHandler and TxHandler. Hopefully the sum is greater than the parts
func MultiTxHandler(fn TxHandlerFunc) http.HandlerFunc {
	return MultiDBHandler(func(db *sql.DB, w http.ResponseWriter, r *http.Request) {
		TxWrapper(db, w, r, fn)
	})
}
