package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var Version = "tip"

func main() {
	addr := flag.String("addr", ":8080", "address in form of ip:port to listen on for http")
	mcaddr := flag.String("mcaddr", "224.0.0.1:8888", "address in form of ip:port to listen on for multicast")
	version := flag.Bool("version", false, "show version and exit")
	flag.Parse()
	if *version {
		fmt.Printf("Sandvine API Aggregator %s", Version)
		os.Exit(1)
	}
	s := &Server{
		Addr:          *addr,
		MulticastAddr: *mcaddr,
	}
	s.Handler = NewHandler(s)
	go func() { log.Fatal(s.ListenAndServe()) }()
	go func() { log.Fatal(s.Discover()) }()
	select {} // Block forever.
}
