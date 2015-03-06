package main

import (
	"encoding/binary"
	"net"
	"strconv"

	"github.com/golang/glog"
)

const maxDatagramSize = 32

// Discover starts a UDP server to listen for multicast packets for
// service discovery. When it learns about a new upstream server it
// keeps it in an internal list until a request to that server fails.
//
// The optional "ready" argument can be used to notify when this
// server is up and running. It is supposed to run on its own goroutine.
func (s *Server) Discover(ready ...chan struct{}) error {
	glog.V(1).Infoln("starting discovery server on", s.MulticastAddr)
	addr, err := net.ResolveUDPAddr("udp", s.MulticastAddr)
	if err != nil {
		return err
	}
	l, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		return err
	}
	l.SetReadBuffer(maxDatagramSize)
	b := make([]byte, 2)
	for {
		if ready != nil {
			close(ready[0])
			ready = nil
		}
		n, src, err := l.ReadFromUDP(b)
		if err != nil {
			return err
		}
		if n != 2 {
			glog.Errorf("received malformed UDP with %d bytes from %s: %v",
				n, src, b)
			continue
		}
		glog.V(2).Infof("received %d bytes UDP from %s: %v", n, src, b)
		port := binary.BigEndian.Uint16(b)
		host, _, err := net.SplitHostPort(src.String())
		if err != nil {
			return err
		}
		peer := host + ":" + strconv.Itoa(int(port))
		s.setUpstream(peer)
	}
}

// Discovered returns a channel where newly discovered peers are
// published as ip:port.
func (s *Server) Discovered() <-chan string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.usev == nil {
		s.usev = make(chan string, 1)
	}
	return s.usev
}

// MulticastPing sends a UDP multicast packet containing a port number
// encoded as uint16 in the payload. This is used to announce ourselves
// to other servers like this, which are listening for announcements
// using the Discover function.
func MulticastPing(addr string, port uint16) error {
	glog.V(2).Infof("sending multicast ping to %s with value %v", addr, port)
	a, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}
	c, err := net.DialUDP("udp", nil, a)
	if err != nil {
		return err
	}
	defer c.Close()
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, port)
	_, err = c.Write(b)
	return err
}
