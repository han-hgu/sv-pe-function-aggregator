package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

// NewHandler creates and initializes an http.ServeMux that contains
// http handlers for endpoints that manage policy engine tables.
func NewHandler(srv *Server) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/tables", handleTables(srv))
	mux.Handle("/tables/", handleTableRows(srv))
	return mux
}

// aggregateResponse is an object used to aggregate responses from
// multiple policy engines into a single response.
type aggregateResponse struct {
	URL  string
	Data interface{}
}

// handleTables handles requests that return a list of tables from
// the policy engine.
//
// If no upstream servers are available it returns an empty JSON
// array. In case an upstream server fails to handle the request,
// we remove it from the list of available upstream servers until
// it announces itself again via multicast.
//
// The response from this handler is a JSON object that contains
// the URL of the upstream server being queries and its data.
//
// Requests to multiple upstream servers are executed concurrently.
func handleTables(srv *Server) http.HandlerFunc {
	f := func(w http.ResponseWriter, r *http.Request) {
		data := make(chan []interface{})
		tables := make(chan *aggregateResponse, 1)
		go func() {
			var d []interface{}
			for table := range tables {
				d = append(d, table)
			}
			data <- d
		}()
		srv.foreachUpstream(func(addr string) error {
			url := "http://" + addr + "/tables"
			data, err := getTables(url)
			if err != nil {
				return err
			}
			tables <- &aggregateResponse{
				URL:  url,
				Data: data,
			}
			return nil
		})
		close(tables)
		d := <-data
		w.Header().Set("Content-Type", "application/json")
		if d == nil {
			json.NewEncoder(w).Encode([]string{})
			return
		}
		json.NewEncoder(w).Encode(d)
	}
	return corsHandler(f, "GET")
}

// TODO(afiori): Make requests and aggregate responses.
func handleTableRows(srv *Server) http.HandlerFunc {
	f := func(w http.ResponseWriter, r *http.Request) {
		// Return 400 (Bad Request) if no table name is given.
		name := r.URL.Path[len("/tables/"):]
		if len(name) == 0 {
			s := http.StatusBadRequest
			http.Error(w, http.StatusText(s), s)
			return
		}
	}
	return corsHandler(f, "GET")
}

// corsHandler is an http handler that filters allowed request methods
// (verbs) and add CORS headers to the response.
//
// See http://en.wikipedia.org/wiki/Cross-origin_resource_sharing for details.
func corsHandler(f http.HandlerFunc, allow ...string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Method",
			strings.Join(allow, ", ")+", OPTIONS")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		for _, method := range allow {
			if r.Method == method {
				f(w, r)
				return
			}
		}
		w.Header().Set("Allow", strings.Join(allow, ", ")+", OPTIONS")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed)
	})
}
