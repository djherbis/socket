package socket

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type actionspace struct {
	actions map[string]*caller
}

func newActionSpace() *actionspace {
	return &actionspace{
		actions: make(map[string]*caller),
	}
}

type Socket struct {
	*actionspace
	conn *websocket.Conn
}

func newSocket(conn *websocket.Conn) *Socket {
	return &Socket{
		actionspace: newActionSpace(),
		conn:        conn,
	}
}

type Server struct {
	*actionspace
	onConnect    func(s *Socket)
	onDisconnect func(s *Socket)
}

func NewServer() *Server {
	return &Server{
		actionspace: newActionSpace(),
	}
}

type inFrame struct {
	Namespace string            `json:"namespace"`
	Event     string            `json:"event"`
	Args      []json.RawMessage `json:"args"`
}

type outFrame struct {
	Namespace string        `json:"namespace"`
	Event     string        `json:"event"`
	Args      []interface{} `json:"args"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	so := newSocket(conn)
	if s.onConnect != nil {
		s.onConnect(so)
	}

	for {
		var f *inFrame
		if err := conn.ReadJSON(&f); err != nil {
			if s.onDisconnect != nil {
				s.onDisconnect(so)
			}
			return
		}

		so.actionspace.on(f)
	}
}

func (s *Server) On(event string, fn interface{}) (err error) {
	switch event {
	case "connection":
		if sofn, ok := fn.(func(so *Socket)); ok {
			s.onConnect = sofn
		} else {
			err = errors.New("connection fn must meet: func(*socket.Socket)")
		}

	case "disconnection":
		if sofn, ok := fn.(func(so *Socket)); ok {
			s.onDisconnect = sofn
		} else {
			err = errors.New("disconnection fn must meet: func(*socket.Socket)")
		}

	default:
		err = s.actionspace.On(event, fn)
	}
	return err
}

func (s *actionspace) On(event string, fn interface{}) (err error) {
	s.actions[event], err = newCaller(fn, json.Unmarshal)
	return err
}

func (s *actionspace) on(f *inFrame) {
	if c, ok := s.actions[f.Event]; ok {
		c.Call(f.Args)
	}
}

func (s *Socket) Emit(event string, args ...interface{}) error {
	return s.conn.WriteJSON(&outFrame{
		Namespace: "test",
		Event:     event,
		Args:      args,
	})
}
