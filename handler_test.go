package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestCORS_OPTIONS(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/", func() http.HandlerFunc {
		f := func(w http.ResponseWriter, r *http.Request) {}
		return corsHandler(f, "GET")
	}())
	s := httptest.NewServer(mux)
	defer s.Close()
	req, err := http.NewRequest("OPTIONS", s.URL+"/", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected server response: %s", resp.Status)
	}
}

func TestCORS_MethodUnsupported(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/", func() http.HandlerFunc {
		f := func(w http.ResponseWriter, r *http.Request) {}
		return corsHandler(f, "GET")
	}())
	s := httptest.NewServer(mux)
	defer s.Close()
	req, err := http.NewRequest("POST", s.URL+"/", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("Unexpected server response: %s", resp.Status)
	}
}

func fakeTables(i int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string][]string{
			"table_names": {
				string('a' + i),
				string('b' + i),
				string('c' + i),
			},
		})
	}
}

func TestHandler_Tables(t *testing.T) {
	srv := new(Server)
	for i := 0; i < 5; i++ {
		mux := http.NewServeMux()
		mux.Handle("/tables", fakeTables(i*3))
		upstream := httptest.NewServer(mux)
		defer upstream.Close()
		u, err := url.Parse(upstream.URL)
		if err != nil {
			t.Fatal(err)
		}
		// Parse the upstream server's URL to extract the port.
		srv.setUpstream(u.Host) // Host should be ip:port.
	}
	handler := NewHandler(srv)
	s := httptest.NewServer(handler)
	defer s.Close()
	resp, err := http.Get(s.URL + "/tables")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected server response: %s", resp.Status)
	}
	var data []aggregateResponse
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 5 {
		t.Fatalf("Unexpected # of records. Want 5, have %d", len(data))
	}
	v, ok := data[0].Data.(map[string]interface{})
	if !ok {
		t.Fatalf("Unexpected data format: %#v", data[0].Data)
	}
	if _, ok = v["table_names"]; !ok {
		t.Fatalf("Missing table_names key: %#v", v)
	}
}

func TestHandler_Tables_BrokenUpstream(t *testing.T) {
	srv := new(Server)
	upstream := httptest.NewServer(http.NewServeMux())
	defer upstream.Close()
	u, err := url.Parse(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}
	// Parse the upstream server's URL to extract the port.
	srv.setUpstream(u.Host) // Host should be ip:port.
	handler := NewHandler(srv)
	s := httptest.NewServer(handler)
	defer s.Close()
	resp, err := http.Get(s.URL + "/tables")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected server response: %s", resp.Status)
	}
	if v := resp.Header.Get("Content-Length"); v != "3" {
		t.Fatalf("Unexpected Content-Length. Want 3, have %s", v)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b, []byte("[]\n")) {
		t.Fatalf("Unexpected response. Want []\\n, have %q", b)
	}
}

func TestHandler_Tables_NoUpstream(t *testing.T) {
	handler := NewHandler(new(Server))
	s := httptest.NewServer(handler)
	defer s.Close()
	resp, err := http.Get(s.URL + "/tables")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected server response: %s", resp.Status)
	}
	if v := resp.Header.Get("Content-Length"); v != "3" {
		t.Fatalf("Unexpected Content-Length. Want 3, have %s", v)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b, []byte("[]\n")) {
		t.Fatalf("Unexpected response. Want []\\n, have %q", b)
	}
}
