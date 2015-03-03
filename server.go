package socket

import "net/http"

type Server struct {
	Namespace
}

func NewServer() *Server {
	return &Server{
		Namespace: newNamespace(""),
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t, err := newWSTransport(w, r)
	if err != nil {
		return
	}

	sockets := newSocketSet()

	for {
		p, err := t.NextPacket()
		if err != nil {
			break
		}

		sockets.add(p)
		s.OnPacket(p)
	}

	sockets.forEach(func(namespace, socketId string) {
		s.OnPacket(&packet{
			transport: t,
			namespace: namespace,
			socket:    socketId,
			event:     Disconnection,
		})
	})
}

type socketSet struct {
	sockets map[string]map[string]struct{}
}

func newSocketSet() *socketSet {
	return &socketSet{sockets: make(map[string]map[string]struct{})}
}

func (s *socketSet) add(p Packet) {
	ns, ok := s.sockets[p.Namespace()]
	if !ok {
		ns = make(map[string]struct{})
		s.sockets[p.Namespace()] = ns
	}
	ns[p.Socket()] = struct{}{}
}

func (s *socketSet) forEach(fn func(string, string)) {
	for ns := range s.sockets {
		for so := range s.sockets[ns] {
			fn(ns, so)
		}
	}
}
