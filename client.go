package socket

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/gorilla/websocket"
)

type ClientSocket interface {
	EventHandler
	Emitter
}

func New(url string) (ClientSocket, error) {
	res := strings.Split(url, "/")
	var host, ns string
	switch len(res) {
	case 0:
		host, ns = "localhost", ""

	case 1:
		host, ns = res[0], ""

	case 2:
		host, ns = res[0], fmt.Sprintf("/%s", res[1])
	}

	properUrl := fmt.Sprintf("ws://%s/socket", host)
	c, r, err := websocket.DefaultDialer.Dial(properUrl, nil)
	if err != nil {
		return nil, err
	}

	id, err := generateId()
	if err != nil {
		return nil, err
	}

	so := newSocket(&clientNS{ns: ns}, id, &wsTransport{
		request: r.Request,
		conn:    c,
	})

	go func() {
		for {
			p, err := so.t.NextPacket()
			if err != nil {
				break
			}
			so.OnPacket(p)
		}
		so.disconnect()
	}()

	return so, nil
}

type clientNS struct {
	ns string
}

func (ns *clientNS) Name() string          { return ns.ns }
func (ns *clientNS) Room(room string) Room { return nil }

func generateId() (string, error) {
	buf := bytes.NewBuffer(nil)
	enc := base64.NewEncoder(base64.StdEncoding, buf)
	_, err := io.CopyN(enc, rand.Reader, 32)
	if err != nil {
		return "", err
	}
	enc.Close()
	return string(buf.Bytes()), nil
}
