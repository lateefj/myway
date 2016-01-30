package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"time"
)

const (
	RuntimeHeader = "Server-Response-Time"
)

// TimeWrap ... Simple wrapper to make printing out time a request takes
func TimeWrap(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rec := httptest.NewRecorder() // Need a recorder so the runtime header can write response time header out before
		s := time.Now()
		fn(rec, r) // Calls the http.HandlerFunc that is passed in
		e := time.Now()
		t := e.Sub(s) // Compute total time diff
		url := r.URL.String()
		total := (t.Seconds() * 1000) // Convert to millis
		// Copy headers over
		for k, v := range rec.Header() {
			w.Header()[k] = v
		}
		w.Header().Set(RuntimeHeader, fmt.Sprintf("%f", total)) // Write response time header
		w.Write(rec.Body.Bytes())
		log.Printf("%s: took %f\n", url, total)
	}
}

func main() {
	http.HandleFunc("/", TimeWrap(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello there.... %s", r.URL.String())
	}))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
