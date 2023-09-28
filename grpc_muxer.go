package plugin

import (
	"crypto/tls"
	"net"
	"sync"

	"github.com/hashicorp/yamux"
)

var _ net.Listener = (*grpcMuxer)(nil)

type grpcMuxer struct {
	mtx          sync.Mutex
	underlyingLn net.Listener
	session      *yamux.Session
}

func (m *grpcMuxer) Accept() (net.Conn, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	// if m.session == nil {
	// 	conn, err := m.underlyingLn.Accept()
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	m.session, err = yamux.Server(conn, nil)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }

	conn, err := m.underlyingLn.Accept()
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (m *grpcMuxer) Close() error {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	if m.session != nil {
		return m.session.Close()
	}

	return m.underlyingLn.Close()
}

func (m *grpcMuxer) Addr() net.Addr {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	return m.underlyingLn.Addr()
}

func (m *grpcMuxer) configureTLS(cfg *tls.Config) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	m.underlyingLn = tls.NewListener(m.underlyingLn, cfg)
}

func (m *grpcMuxer) Session() *yamux.Session {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	return m.session
}
