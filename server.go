package socket

import (
	"net/http"
	"sync"
)

// Server handles creating Sockets from http connections
type Server struct {
	mu sync.RWMutex
	*namespace
	subspaces map[string]*namespace
}

// Of creates a new Namespace with "name", or returns the existing Namespace
// with name "name"
func (s *Server) Of(name string) Namespace {
	return s.ns(name)
}

func (s *Server) ns(name string) *namespace {
	s.mu.Lock()
	defer s.mu.Unlock()
	if nsp, ok := s.subspaces[name]; ok {
		return nsp
	}
	subns := newNamespace(name)
	s.subspaces[name] = subns
	return subns
}

// NewServer creates a new Server
func NewServer() *Server {
	s := &Server{
		namespace: newNamespace(""),
		subspaces: make(map[string]*namespace),
	}
	s.subspaces[""] = s.namespace
	s.subspaces["/"] = s.namespace
	return s
}

// ServeHttp handles http requests and converting them into Sockets
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t, err := newWSServer(w, r)
	if err != nil {
		return
	}

	sockets := make(map[string]*socket)

	for {
		p, err := t.NextPacket()
		if err != nil {
			break
		}

		ns := s.ns(p.Namespace())
		so, ok := sockets[p.Socket()]
		if !ok {
			so = newSocket(ns, p.Socket(), t)
			ns.addSocket(so)
			sockets[so.ID()] = so
		}

		ns.OnPacket(p)
		so.OnPacket(p)
	}

	for _, so := range sockets {
		so.disconnect()
		s.ns(so.Namespace()).removeSocket(so)
	}
}
