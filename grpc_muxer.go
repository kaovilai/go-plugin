package plugin

import (
	"crypto/tls"
	"errors"
	"net"
	"sync"

	"github.com/hashicorp/yamux"
)

var _ net.Listener = (*grpcMuxer)(nil)

type grpcMuxer struct {
	once         sync.Once
	underlyingLn net.Listener
	session      *yamux.Session
}

func (m *grpcMuxer) Accept() (net.Conn, error) {
	var onceErr error
	m.once.Do(func() {
		if m.underlyingLn == nil {
			return
		}

		var conn net.Conn
		conn, onceErr = m.underlyingLn.Accept()
		if onceErr != nil {
			return
		}

		m.session, onceErr = yamux.Server(conn, nil)
		if onceErr != nil {
			return
		}
	})
	if onceErr != nil {
		return nil, onceErr
	}

	conn, err := m.session.Accept()
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (m *grpcMuxer) Close() error {
	var err error
	// if m.session != nil {
	// 	errors.Join(err, m.session.Close())
	// }

	if m.underlyingLn != nil {
		errors.Join(err, m.underlyingLn.Close())
	}

	return err
}

func (m *grpcMuxer) Addr() net.Addr {
	if m.underlyingLn != nil {
		return m.underlyingLn.Addr()
	} else {
		return m.session.Addr()
	}
}

func (m *grpcMuxer) configureTLS(cfg *tls.Config) {
	m.underlyingLn = tls.NewListener(m.underlyingLn, cfg)
}

func (m *grpcMuxer) Dial() (net.Conn, error) {
	if m.session == nil {
		return nil, errors.New("connection not yet made")
	}

	return m.session.Open()
}
