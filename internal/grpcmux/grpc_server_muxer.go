package grpcmux

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/yamux"
)

var _ GRPCMuxer = (*GRPCServerMuxer)(nil)

type GRPCServerMuxer struct {
	addr   net.Addr
	errCh  chan error
	logger hclog.Logger

	sess *yamux.Session
}

func NewGRPCServerMuxer(logger hclog.Logger, ln net.Listener) *GRPCServerMuxer {
	m := &GRPCServerMuxer{
		addr:   ln.Addr(),
		errCh:  make(chan error),
		logger: logger,
	}

	go m.acceptNewConnection(ln)

	return m
}

func (m *GRPCServerMuxer) acceptNewConnection(ln net.Listener) {
	defer close(m.errCh)

	m.logger.Debug("accepting initial connection")
	conn, err := ln.Accept()
	if err != nil {
		m.errCh <- err
		return
	}

	m.logger.Debug("initial server connection accepted", "addr", m.addr)
	m.sess, err = yamux.Server(conn, nil)
	if err != nil {
		m.errCh <- err
		return
	}
	m.logger.Debug("server session created")
	m.errCh <- nil
}

func (m *GRPCServerMuxer) session() (*yamux.Session, error) {
	select {
	case err := <-m.errCh:
		if err != nil {
			return nil, err
		}
	case <-time.After(2 * time.Second):
		return nil, errors.New("timed out waiting for connection to be established")
	}

	// Should never happen.
	if m.sess == nil {
		return nil, errors.New("no connection established and no error received")
	}

	return m.sess, nil
}

func (m *GRPCServerMuxer) Dial() (net.Conn, error) {
	sess, err := m.session()
	if err != nil {
		return nil, fmt.Errorf("error getting session for server Dial: %w", err)
	}

	m.logger.Debug("dialling new server stream...")
	stream, err := sess.OpenStream()
	if err != nil {
		return nil, fmt.Errorf("error dialling new server stream: %w", err)
	}

	m.logger.Debug("dialled new server stream", "id", stream.StreamID())
	return stream, nil
}

func (m *GRPCServerMuxer) Accept() (net.Conn, error) {
	sess, err := m.session()
	if err != nil {
		return nil, fmt.Errorf("error getting session for server Accept: %w", err)
	}

	m.logger.Debug("accepting new server stream...")
	stream, err := sess.AcceptStream()
	if err != nil {
		return nil, fmt.Errorf("error accepting new server stream: %w", err)
	}

	m.logger.Debug("accepted new server stream", "id", stream.StreamID())
	return stream, nil
}

func (m *GRPCServerMuxer) Addr() net.Addr {
	return m.addr
}

func (m *GRPCServerMuxer) Close() error {
	m.logger.Debug("closing server muxer")
	if m.sess != nil {
		return m.sess.Close()
	}

	return nil
}
