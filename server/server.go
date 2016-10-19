package server

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
)

type Store interface {
	// Get returns the value for the given key.
	Get(key []byte) ([]byte, error)

	// Set sets the value for the given key, via distributed consensus.
	Set(key, value []byte) error

	// Delete removes the given key, via distributed consensus.
	Delete(key []byte) error

	// Join joins the node, reachable at addr, to the cluster.
	Join(addr []byte) error
}

type Server struct {
	addr string
	ln   net.Listener

	store Store
}

func New(addr string, store Store) *Server {
	return &Server{
		addr:  addr,
		store: store,
	}
}

// Start starts the service.
func (s *Server) Start() error {
	server := http.Server{
		Handler: s,
	}

	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.ln = ln

	http.Handle("/", s)

	go func() {
		err := server.Serve(s.ln)
		if err != nil {
			log.Fatalf("HTTP serve: %s", err)
		}
	}()

	return nil
}

// Close closes the service.
func (s *Server) Close() {
	s.ln.Close()
	return
}

// ServeHTTP allows Server to serve HTTP requests.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/join" {
		s.handleJoin(w, r)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (s *Server) handleJoin(w http.ResponseWriter, r *http.Request) {
	m := map[string]string{}
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(m) != 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	remoteAddr, ok := m["addr"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := s.store.Join([]byte(remoteAddr)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// Addr returns the address on which the Server is listening
func (s *Server) Addr() net.Addr {
	return s.ln.Addr()
}
