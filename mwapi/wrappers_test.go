package mwapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/lateefj/mctest"
)

func simpleHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("simpleHandler"))
}

func TestJSONWrap(t *testing.T) {
	req, _ := http.NewRequest("GET", "/path/to/handler", nil)
	resp := mctest.NewMockTestResponse(t)
	JSONWrap(simpleHandler)(resp, req)
	h := resp.Header().Get("Content-Type")
	if h == "" {
		t.Fatalf("Exepcted runtime header to be set but was not")
	}
	if h != "application/json" {
		t.Fatalf("Expected header to be 'application/json' however is %s", h)
	}
}

func TestJsonify(t *testing.T) {
	req, _ := http.NewRequest("GET", "/path/to/handler", nil)
	resp := mctest.NewMockTestResponse(t)

	data := &struct {
		X string
		Y string
	}{"foo", "bar"}
	Jsonify(func(w http.ResponseWriter, r *http.Request) (interface{}, error) {
		return data, nil
	})(resp, req)
	err := json.Unmarshal(resp.Bytes(), data)
	if err != nil {
		t.Fatalf("Error unmarshaling data %s", err)
	}
	if data.X != "foo" || data.Y != "bar" {
		t.Fatalf("Expected X to be 'foo' but is '%s' and Y to be 'bar' but is '%s'", data.X, data.Y)
	}
	resp = mctest.NewMockTestResponse(t)
	Jsonify(func(w http.ResponseWriter, r *http.Request) (interface{}, error) {
		return nil, errors.New("Test error")
	})(resp, req)
	resp.AssertCode(http.StatusInternalServerError)
	resp = mctest.NewMockTestResponse(t)
	Jsonify(func(w http.ResponseWriter, r *http.Request) (interface{}, error) {
		return nil, nil
	})(resp, req)
	resp.AssertCode(http.StatusOK)
}
