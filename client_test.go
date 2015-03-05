package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_GetTables(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string][]string{
			"table_names": {"a", "b", "c"},
		})
	})
	s := httptest.NewServer(mux)
	defer s.Close()
	m, err := getTables(s.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	v := m["table_names"]
	if v == nil {
		t.Fatalf("Missing table_names key: %#v", m)
	}
	if len(v) != 3 {
		t.Fatalf("Unexpected # of tables. Want 3, have %d", len(v))
	}
}

func TestClient_GetTables_UnexpectedStatus(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	s := httptest.NewServer(mux)
	defer s.Close()
	m, err := getTables(s.URL + "/")
	if err != errUnexpectedStatus {
		t.Fatalf("Expected error didn't occur. Got: %#v, %s", m, err)
	}
}

func TestClient_GetTables_UnexpectedContentType(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	s := httptest.NewServer(mux)
	defer s.Close()
	m, err := getTables(s.URL + "/")
	if err != errUnexpectedContentType {
		t.Fatalf("Expected error didn't occur. Got: %#v, %s", m, err)
	}
}

func TestClient_GetTables_UnexpectedResponse(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	})
	s := httptest.NewServer(mux)
	defer s.Close()
	m, err := getTables(s.URL + "/")
	if err != errUnexpectedResponse {
		t.Fatalf("Expected error didn't occur. Got: %#v, %s", m, err)
	}
}

func TestClient_GetTables_UnexpectedDocument(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string][]string{
			"foobar": {"123"},
		})
	})
	s := httptest.NewServer(mux)
	defer s.Close()
	m, err := getTables(s.URL + "/")
	if err != errUnexpectedDocument {
		t.Fatalf("Expected error didn't occur. Got: %#v, %s", m, err)
	}
}
