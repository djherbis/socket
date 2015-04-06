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

	// ID() is this sockets unique identifier
	ID() string

	// Namespace() is the namespace this socket is a part of
	Namespace() string

	// On will register an function to handle events sent
	// from the other end of the socket
	EventHandler

	// Emit will send an event to the other end of the socket
	Emitter

	// Close the underlying Transport
	Close() error
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

func (s *clientSocket) ID() string { return s.id }

func New(url string) (ClientSocket, error) {
	id, err := generateID()
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

	return so, so.Emit(Connection)
}

func (s *clientSocket) Namespace() string {
	return s.ns
}

func (s *clientSocket) Emit(event string, args ...interface{}) error {
	return s.Send(s.ns, s.ID(), event, args...)
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

func generateID() (string, error) {
	buf := bytes.NewBuffer(nil)
	enc := base64.NewEncoder(base64.URLEncoding, buf)
	_, err := io.CopyN(enc, rand.Reader, 32)
	if err != nil {
		return "", err
	}
	enc.Close()
	return string(buf.Bytes()), nil
}
