package socket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

const Connection = "connection"
const Disconnection = "disconnection"
const Disconnect = "disconnect"

type Packet interface {
	Namespace() string
	Socket() string
	Event() string
	DecodeArgs(args ...interface{})
}

type Transport interface {
	Request() *http.Request
	NextPacket() (Packet, error)
	Send(string, string, string, ...interface{}) error
	Close() error
}

type wsTransport struct {
	mu      sync.Mutex
	request *http.Request
	conn    *websocket.Conn
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func splitHostNamespace(url string) (host, ns string) {
	res := strings.Split(url, "/")
	switch len(res) {
	case 0:
		host, ns = "localhost", ""

	case 1:
		host, ns = res[0], ""

	case 2:
		host, ns = res[0], fmt.Sprintf("/%s", res[1])
	}

	return host, ns
}

func newWSClient(host string) (Transport, error) {
	properUrl := fmt.Sprintf("ws://%s/socket", host)
	c, r, err := websocket.DefaultDialer.Dial(properUrl, nil)
	if err != nil {
		return nil, err
	}

	return &wsTransport{
		request: r.Request,
		conn:    c,
	}, nil
}

func newWSServer(w http.ResponseWriter, r *http.Request) (Transport, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	return &wsTransport{
		request: r,
		conn:    conn,
	}, nil
}

func (ws *wsTransport) Request() *http.Request {
	return ws.request
}

func (ws *wsTransport) NextPacket() (Packet, error) {
	var f inPacket
	f.decode = json.Unmarshal
	if err := ws.conn.ReadJSON(&f); err != nil {
		return nil, err
	}
	return &f, nil
}

func (ws *wsTransport) Close() error {
	return ws.conn.Close()
}

func (ws *wsTransport) Send(ns, socket, event string, args ...interface{}) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	return ws.conn.WriteJSON(&outPacket{
		N:    ns,
		S:    socket,
		E:    event,
		Args: args,
	})
}

type inPacket struct {
	N      string            `json:"namespace"`
	S      string            `json:"socket"`
	E      string            `json:"event"`
	Args   []json.RawMessage `json:"args"`
	decode func([]byte, interface{}) error
}

type outPacket struct {
	N    string        `json:"namespace"`
	S    string        `json:"socket"`
	E    string        `json:"event"`
	Args []interface{} `json:"args"`
}

func (p *inPacket) Namespace() string { return p.N }
func (p *inPacket) Socket() string    { return p.S }
func (p *inPacket) Event() string     { return p.E }
func (p *inPacket) DecodeArgs(args ...interface{}) {
	for i, _ := range args {
		if i >= len(p.Args) {
			return
		}
		if err := p.decode(p.Args[i], &args[i]); err != nil {
			return
		}
	}
}
