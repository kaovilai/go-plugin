package plugin

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/yamux"
)

var _ grpcMuxer = (*grpcServerMuxer)(nil)

type grpcServerMuxer struct {
	addr   net.Addr
	errCh  chan error
	logger hclog.Logger

	sess *yamux.Session
}

func newGRPCServerMuxer(logger hclog.Logger, ln net.Listener) *grpcServerMuxer {
	m := &grpcServerMuxer{
		addr:   ln.Addr(),
		errCh:  make(chan error),
		logger: logger,
	}

	go func() {
		defer close(m.errCh)

		logger.Debug("accepting initial connection")
		conn, err := ln.Accept()
		if err != nil {
			m.errCh <- err
			return
		}

		logger.Debug("initial server connection accepted", "addr", ln.Addr())
		m.sess, err = yamux.Server(conn, nil)
		if err != nil {
			m.errCh <- err
			return
		}
		logger.Debug("server session created")
	}()

	return m
}

func (m *grpcServerMuxer) session() (*yamux.Session, error) {
	select {
	case err := <-m.errCh:
		if err != nil {
			m.logger.Error("error from channel", "error", err)
			return nil, err
		}
	case <-time.After(2 * time.Second):
		m.logger.Error("timeout waiting for server session")
		return nil, errors.New("timed out waiting for connection to be established")
	}

	// Should never happen.
	if m.sess == nil {
		return nil, errors.New("no connection established and no error received")
	}

	return m.sess, nil
}

func (m *grpcServerMuxer) Dial() (net.Conn, error) {
	sess, err := m.session()
	if err != nil {
		return nil, fmt.Errorf("error getting session for server Dial: %w", err)
	}

	stream, err := sess.OpenStream()
	if err != nil {
		return nil, fmt.Errorf("error dialling new server stream: %w", err)
	}

	m.logger.Debug("dialled new server stream", "id", stream.StreamID())
	return stream, nil
}

func (m *grpcServerMuxer) Accept() (net.Conn, error) {
	sess, err := m.session()
	if err != nil {
		return nil, fmt.Errorf("error getting session for server Accept: %w", err)
	}

	stream, err := sess.AcceptStream()
	if err != nil {
		return nil, fmt.Errorf("error accepting new server stream: %w", err)
	}

	m.logger.Debug("accepted new server stream", "id", stream.StreamID())
	return stream, nil
}

func (m *grpcServerMuxer) Addr() net.Addr {
	return m.addr
}

func (m *grpcServerMuxer) Close() error {
	m.logger.Debug("closing server muxer")
	if m.sess != nil {
		return m.sess.Close()
	}

	return nil
}

// func (m *grpcServerMuxer) configureTLS(cfg *tls.Config) {
// 	m.underlyingLn = tls.NewListener(m.underlyingLn, cfg)
// }
