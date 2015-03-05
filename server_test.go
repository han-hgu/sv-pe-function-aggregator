package main

import (
	"encoding/binary"
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
	err := announceMulticast(s.MulticastAddr, 1111)
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

func announceMulticast(addr string, port uint32) error {
	a, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}
	c, err := net.DialUDP("udp", nil, a)
	if err != nil {
		return err
	}
	defer c.Close()
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, port)
	_, err = c.Write(b)
	return err
}
