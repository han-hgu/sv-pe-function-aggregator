package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
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
// the policy engine. If a specific table name is given at the
// end of the URI, then we return rows from that table.
func handleTables(srv *Server) http.HandlerFunc {
	f := func(w http.ResponseWriter, r *http.Request) {
		wg := &sync.WaitGroup{}
		datac := make(chan *aggregateResponse)
		n := 0
		srv.foreachUpstream(func(addr string) error {
			n++
			wg.Add(1)
			defer wg.Done()
			url := "http://" + addr + "/tables"
			errc := make(chan error)
			go func() {
				data, err := getTables(url)
				if err != nil {
					errc <- err
					return
				}
				close(errc)
				datac <- &aggregateResponse{
					URL:  url,
					Data: data,
				}
			}()
			// In case of error, we remove this upstream
			// server from the internal list.
			return <-errc
		})
		wg.Wait()
		if n == 0 {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]string{})
			return
		}
		var data []interface{}
	aggregation:
		for i := 0; i < n; i++ {
			select {
			case chunk := <-datac:
				data = append(data, chunk)
			default:
				break aggregation
			}
		}
		w.Header().Set("Content-Type", "application/json")
		if data == nil {
			json.NewEncoder(w).Encode([]string{})
			return
		}
		json.NewEncoder(w).Encode(data)
	}
	return corsHandler(f, "GET")
}

func handleTableRows(srv *Server) http.HandlerFunc {
	f := func(w http.ResponseWriter, r *http.Request) {
		// Return 400 (Bad Request) if no table name is given.
		name := r.URL.Path[len("/tables/"):]
		if len(name) == 0 {
			s := http.StatusBadRequest
			http.Error(w, http.StatusText(s), s)
			return
		}
		// TODO(afiori): Make requests and aggregate responses.
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
