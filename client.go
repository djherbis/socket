package socket

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"io"
)

type ClientSocket interface {
	Id() string
	Namespace() string
	EventHandler
	Emitter
}

func New(url string) (ClientSocket, error) {
	id, err := generateId()
	if err != nil {
		return nil, err
	}

	host, ns := splitHostNamespace(url)
	t, err := newWSClient(host)
	if err != nil {
		return nil, err
	}

	so := newSocket(&clientNS{ns: ns}, id, t)

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
