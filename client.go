package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/golang/glog"
)

// getTables queries a remote web server and return a list of tables
// available in the policy engine of that server.
func getTables(url string) (map[string][]string, error) {
	glog.V(2).Infof("making request to upstream server %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errUnexpectedStatus
	}
	if resp.Header.Get("Content-Type") != "application/json" {
		return nil, errUnexpectedContentType
	}
	var m map[string][]string
	err = json.NewDecoder(resp.Body).Decode(&m)
	if err != nil {
		return nil, errUnexpectedResponse
	}
	if _, ok := m["table_names"]; !ok {
		return nil, errUnexpectedDocument
	}
	return m, nil
}

var (
	errUnexpectedStatus      = errors.New("unexpected status code")
	errUnexpectedContentType = errors.New("unexpected content type")
	errUnexpectedResponse    = errors.New("unexpected server response")
	errUnexpectedDocument    = errors.New("unexpected server document")
)
