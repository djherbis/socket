package socket

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"sync"
)

// ClientSocket creates a client-side Socket
type ClientSocket interface {

	// Id() is this sockets unique identifier
	Id() string

	// Namespace() is the namespace this socket is a part of
	Namespace() string

	// On will register an function to handle events sent
	// from the other end of the socket
	EventHandler

	// Emit will send an event to the other end of the socket
	Emitter
}

type clientSocket struct {
	mu sync.RWMutex
	id string
	ns string
	Handler
	Transport
	onDisconnect func()
}

func newClientSocket(id, ns string, t Transport) *clientSocket {
	return &clientSocket{
		id:        id,
		ns:        ns,
		Handler:   newHandler(),
		Transport: t,
	}
}

func (s *clientSocket) Id() string { return s.id }

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

	so := newClientSocket(id, ns, t)

	go func() {
		for {
			p, err := so.NextPacket()
			if err != nil {
				break
			}
			so.OnPacket(p)
		}
		so.disconnect()
	}()

	return so, nil
}

func (s *clientSocket) Namespace() string {
	return s.ns
}

func (s *clientSocket) Emit(event string, args ...interface{}) error {
	return s.Send(s.ns, s.Id(), event, args...)
}

func (s *clientSocket) On(event string, fn interface{}) error {
	switch event {
	case Disconnect:
		sfn, ok := fn.(func())
		if !ok {
			return errors.New("Disconnect takes a func of type func()")
		}

		s.mu.Lock()
		s.onDisconnect = sfn
		s.mu.Unlock()
		return nil

	default:
		return s.Handler.On(event, fn)
	}
}

func (s *clientSocket) disconnect() {
	s.mu.RLock()
	fn := s.onDisconnect
	s.mu.RUnlock()
	if fn != nil {
		fn()
	}
}

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
