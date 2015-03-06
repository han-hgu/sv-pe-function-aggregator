package main

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/golang/glog"
)

// Server is a specialized http server that also listens on a UDP multicast
// address and learn about upstream servers from there.
type Server struct {
	Addr          string // Address in form of ip:port to listen on.
	MulticastAddr string // Multicast address in form of ip:port to listen on.

	mu       sync.RWMutex        // Guards all the below.
	Handler  *http.ServeMux      // Our request multiplexer.
	upstream map[string]struct{} // Map of ip:port of upstream servers.
	usev     chan string         // Upstream server discovery events.
}

// ListenAndServe makes the server start accepting http connections.
func (s *Server) ListenAndServe() error {
	if glog.V(1) {
		glog.Infoln("starting http server on", s.Addr)
		return http.ListenAndServe(s.Addr, httpLog(s.Handler))
	}
	return http.ListenAndServe(s.Addr, s.Handler)
}

// setUpstream records the upstream server discovered via multicast
// and sends its address to the events channel.
func (s *Server) setUpstream(addr string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.upstream == nil {
		s.upstream = make(map[string]struct{})
	}
	if _, ok := s.upstream[addr]; !ok {
		glog.V(2).Infof("upstream server discovered: %s", addr)
		s.upstream[addr] = struct{}{}
	}
	// Notify without blocking.
	if s.usev == nil {
		s.usev = make(chan string, 1)
	}
	select {
	case s.usev <- addr:
	default:
	}
}

// delUpstream removes the given upstream from the internal list.
func (s *Server) delUpstream(addr string) {
	glog.V(2).Infof("upstream server deleted: %s", addr)
	s.mu.Lock()
	if s.upstream != nil {
		delete(s.upstream, addr)
	}
	s.mu.Unlock()
}

// upstreamList returns a list of all upstreams currently available.
func (s *Server) upstreamList() []string {
	s.mu.RLock()
	i := 0
	peers := make([]string, len(s.upstream))
	for addr := range s.upstream {
		peers[i] = addr
		i++
	}
	s.mu.RUnlock()
	return peers
}

// foreachUpstream loops over each upstream server calling f in its own
// goroutine. In case f returns an error, the upstream server is removed
// from the internal list.
func (s *Server) foreachUpstream(f func(addr string) error) {
	var err error
	var wg sync.WaitGroup
	for _, addr := range s.upstreamList() {
		wg.Add(1)
		go func(addr string) {
			if err = f(addr); err != nil {
				s.delUpstream(addr)
			}
			wg.Done()
		}(addr)
	}
	wg.Wait()
}

// httpLog logs http requests.
func httpLog(f http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := responseWriter{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		f.ServeHTTP(&resp, r)
		elapsed := time.Since(start)
		glog.Infof("%s %d %q %q %s %db %s",
			r.Proto,
			resp.status,
			r.Method,
			r.URL.Path,
			remoteIP(r),
			resp.bytes,
			elapsed,
		)
	})
}

// remoteIP returns the client's address without the port number.
func remoteIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// responseWriter is an http.ResponseWriter that records the returned
// status and bytes written to the client.
type responseWriter struct {
	http.ResponseWriter
	flusher http.Flusher
	status  int
	bytes   int
}

// Write implements the http.ResponseWriter interface.
func (w *responseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	if err != nil {
		return 0, err
	}
	w.bytes += n
	return n, nil
}

// WriteHeader implements the http.ResponseWriter interface.
func (w *responseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
