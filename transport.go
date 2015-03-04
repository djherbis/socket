package socket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

const Connection = "connection"
const Disconnection = "disconnection"
const Disconnect = "disconnect"

type Packet interface {
	Transport() Transport
	Namespace() string
	Socket() string
	Event() string
	DecodeArgs(args ...interface{})
}

type Transport interface {
	Request() *http.Request
	NextPacket() (Packet, error)
	Send(string, string, string, ...interface{}) error
}

type wsTransport struct {
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
	var f *jsonInFrame
	if err := ws.conn.ReadJSON(&f); err != nil {
		return nil, err
	}
	return f.toPacket(ws), nil
}

func (ws *wsTransport) Send(ns, socket, event string, args ...interface{}) error {
	return ws.conn.WriteJSON(&jsonOutFrame{
		Namespace: ns,
		Socket:    socket,
		Event:     event,
		Args:      args,
	})
}

type packet struct {
	transport Transport
	namespace string
	socket    string
	event     string
	args      [][]byte
	decode    func([]byte, interface{}) error
}

func (p *packet) Transport() Transport { return p.transport }
func (p *packet) Namespace() string    { return p.namespace }
func (p *packet) Socket() string       { return p.socket }
func (p *packet) Event() string        { return p.event }
func (p *packet) DecodeArgs(args ...interface{}) {
	for i, _ := range args {
		if i >= len(p.args) {
			return
		}
		if err := p.decode(p.args[i], &args[i]); err != nil {
			return
		}
	}
}

type jsonInFrame struct {
	Namespace string            `json:"namespace"`
	Socket    string            `json:"socket"`
	Event     string            `json:"event"`
	Args      []json.RawMessage `json:"args"`
}

func (f *jsonInFrame) toPacket(t Transport) Packet {
	args := make([][]byte, len(f.Args))
	for i, jarg := range f.Args {
		args[i] = jarg
	}

	return &packet{
		transport: t,
		namespace: f.Namespace,
		socket:    f.Socket,
		event:     f.Event,
		args:      args,
		decode:    json.Unmarshal,
	}
}

type jsonOutFrame struct {
	Namespace string        `json:"namespace"`
	Socket    string        `json:"socket"`
	Event     string        `json:"event"`
	Args      []interface{} `json:"args"`
}
