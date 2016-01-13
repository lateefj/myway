package mwapi

import (
	"encoding/json"
	"net/http"

	"log"
)

// JSONWrap ... Just sets the proper header to json
// Take a look at Jsonify it is a much better wrapper for most cases
func JSONWrap(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", "application/json")
		}
		fn(w, r)
	}
}

// JsonHandler ... Is the requirement for a function that will automatically jsonify the return
type JsonHandler func(http.ResponseWriter, *http.Request) (interface{}, error)

// Jsonify ... Automatically jsonify
func Jsonify(jfn JsonHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		JSONWrap(func(w http.ResponseWriter, r *http.Request) {

			resp, err := jfn(w, r)
			if err != nil {
				log.Printf("Error in json handler function %s", err)
				http.Error(w, "Sorry Internal server error", http.StatusInternalServerError)
				return
			}
			if resp == nil {
				log.Printf("Can't jsonify nil %s", err)
				http.Error(w, "Sorry Internal server error", http.StatusInternalServerError)
			}
			b, err := json.Marshal(resp)
			if err != nil {
				log.Printf("Error json marshaling %s", err)
				http.Error(w, "Sorry Internal server error", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(b)
		})(w, r)
	}
}
