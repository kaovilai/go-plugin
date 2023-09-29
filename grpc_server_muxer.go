package plugin

import (
	"errors"
	"net"

	"github.com/hashicorp/yamux"
)

var _ grpcMuxer = (*grpcServerMuxer)(nil)

type grpcServerMuxer struct {
	session *yamux.Session
	errCh   chan error
}

func newGRPCServerMuxer(ln net.Listener) *grpcServerMuxer {
	m := &grpcServerMuxer{
		errCh: make(chan error),
	}

	go func() {
		defer close(m.errCh)

		conn, err := ln.Accept()
		if err != nil {
			m.errCh <- err
			return
		}

		m.session, err = yamux.Server(conn, nil)
		if err != nil {
			m.errCh <- err
			return
		}
	}()

	return m
}

func (m *grpcServerMuxer) Session() (*yamux.Session, error) {
	select {
	case err := <-m.errCh:
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("connection not yet established")
	}

	// Should never happen.
	if m.session == nil {
		return nil, errors.New("no connection established and no error received")
	}

	return m.session, nil
}

func (m *grpcServerMuxer) Close() error {
	if m.session != nil {
		return m.session.Close()
	}

	return nil
}

// func (m *grpcServerMuxer) configureTLS(cfg *tls.Config) {
// 	m.underlyingLn = tls.NewListener(m.underlyingLn, cfg)
// }
