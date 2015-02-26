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

	sockets := make(map[string]map[string]struct{})

	for {
		p, err := t.NextPacket()
		if err != nil {
			break
		}

		ns, ok := sockets[p.Namespace()]
		if !ok {
			ns = make(map[string]struct{})
			sockets[p.Namespace()] = ns
		}
		ns[p.Socket()] = struct{}{}

		s.OnPacket(p)
	}

	for ns := range sockets {
		for so := range sockets[ns] {
			s.OnPacket(&packet{
				transport: t,
				namespace: ns,
				socket:    so,
				event:     DISCONNECTION,
			})
		}
	}
}
