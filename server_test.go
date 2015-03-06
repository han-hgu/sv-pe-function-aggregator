package main

import (
	"errors"
	"net"
	"testing"
	"time"
)

func TestServer_Discovery(t *testing.T) {
	s := &Server{
		MulticastAddr: "224.0.0.1:8888",
	}
	ready := make(chan struct{})
	go s.Discover(ready)
	<-ready
	err := MulticastPing(s.MulticastAddr, 1111)
	if err != nil {
		t.Fatal(err)
	}
	select {
	case peer := <-s.Discovered():
		_, port, err := net.SplitHostPort(peer)
		if err != nil {
			t.Fatal(err)
		}
		if port != "1111" {
			t.Fatalf("Unexpected port. Want 1111, have %s", port)
		}
	case <-time.After(time.Second):
		t.Fatal("No peer discovered")
	}
}

func TestServer_Upstream(t *testing.T) {
	s := &Server{}
	s.setUpstream("a")
	s.setUpstream("b")
	s.setUpstream("c")
	s.foreachUpstream(func(s string) error {
		if s == "a" {
			// This forces a call to s.delUpstream.
			return errors.New("failed")
		}
		return nil
	})
	if len(s.upstream) != 2 {
		t.Fatalf("Unexpected # of upstreams. Want 2, have %d", len(s.upstream))
	}
}
