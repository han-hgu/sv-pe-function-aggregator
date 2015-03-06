package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/golang/glog"
)

var Version = "tip"

func main() {
	laddr := flag.String("http_addr", ":8080", "address in form of ip:port to listen on for http")
	lmaddr := flag.String("multicast_addr", "224.0.0.1:8888", "address in form of ip:port to listen on for multicast")
	maddr := flag.String("multicast_ping", "", "address in form of ip:port to announce ourselves via multicast")
	mintvl := flag.Duration("multicast_interval", 30*time.Second, "interval between multicast pings")
	version := flag.Bool("version", false, "show version and exit")
	flag.Parse()
	if *version {
		fmt.Printf("Sandvine API Aggregator %s", Version)
		os.Exit(1)
	}
	if len(*maddr) > 0 && *maddr == *lmaddr {
		glog.Fatal("cannot announce ourselves to the same address that we listen on for multicast")
	}
	s := &Server{
		Addr:          *laddr,
		MulticastAddr: *lmaddr,
	}
	s.Handler = NewHandler(s)
	go func() { glog.Fatal(s.ListenAndServe()) }()
	go func() { glog.Fatal(s.Discover()) }()
	if len(*maddr) > 0 {
		go announce(*mintvl, *maddr, *laddr)
	}
	select {} // Block forever.
}

func announce(interval time.Duration, multicast_addr, http_addr string) {
	glog.Infof("sending announcements to %s every %s",
		multicast_addr, interval)
	_, port, err := net.SplitHostPort(http_addr)
	if err != nil {
		log.Fatal(err)
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		log.Fatal(err)
	}
	v := uint16(p)
	for {
		MulticastPing(multicast_addr, v)
		time.Sleep(interval)
	}
}
