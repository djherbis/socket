package socket

import (
	"net/http"
	"sync"
)

type Server struct {
	mu sync.RWMutex
	*namespace
	subspaces map[string]*namespace
}

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

func NewServer() *Server {
	s := &Server{
		namespace: newNamespace(""),
		subspaces: make(map[string]*namespace),
	}
	s.subspaces[""] = s.namespace
	s.subspaces["/"] = s.namespace
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t, err := newWSTransport(w, r)
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
			so = newSocket(ns, p.Socket(), p.Transport())
			ns.addSocket(so)
			sockets[so.Id()] = so
		}

		ns.OnPacket(p)
		so.OnPacket(p)
	}

	for _, so := range sockets {
		so.disconnect()
		s.ns(so.Namespace()).removeSocket(so)
	}
}
