package grpcmux

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/yamux"
)

var _ GRPCMuxer = (*GRPCServerMuxer)(nil)

type GRPCServerMuxer struct {
	addr   net.Addr
	logger hclog.Logger

	sessionErrCh chan error
	sess         *yamux.Session

	rootMuxer *rootMuxer
}

func NewGRPCServerMuxer(logger hclog.Logger, ln net.Listener) *GRPCServerMuxer {
	m := &GRPCServerMuxer{
		addr:   ln.Addr(),
		logger: logger,

		sessionErrCh: make(chan error),
	}

	go m.acceptSession(ln)

	m.rootMuxer = newRootMuxer(logger, ln.Addr(), false, m.session)

	return m
}

// acceptSessionAndMuxAccept is responsible for establishing the yamux session,
// and then kicking off the acceptLoop function.
func (m *GRPCServerMuxer) acceptSession(ln net.Listener) {
	defer close(m.sessionErrCh)

	m.logger.Debug("accepting initial connection")
	conn, err := ln.Accept()
	if err != nil {
		m.sessionErrCh <- err
		return
	}

	m.logger.Debug("initial server connection accepted", "addr", m.addr)
	m.sess, err = yamux.Server(conn, nil)
	if err != nil {
		m.sessionErrCh <- err
		return
	}
	m.logger.Debug("server session created")
}

func (m *GRPCServerMuxer) session() (*yamux.Session, error) {
	select {
	case err := <-m.sessionErrCh:
		if err != nil {
			return nil, err
		}
	case <-time.After(5 * time.Second):
		return nil, errors.New("timed out waiting for connection to be established")
	}

	// Should never happen.
	if m.sess == nil {
		return nil, errors.New("no connection established and no error received")
	}

	return m.sess, nil
}

func (m *GRPCServerMuxer) MainListener() net.Listener {
	return m.rootMuxer
}

func (m *GRPCServerMuxer) Listener(id uint32, listenForKnocksFn func(context.Context, uint32) error) (net.Listener, error) {
	return m.rootMuxer.listener(id, listenForKnocksFn)
}

func (m *GRPCServerMuxer) KnockAndDial(id uint32, knockFn func(id uint32) error) (net.Conn, error) {
	return m.rootMuxer.knockAndDial(id, knockFn)
}

func (m *GRPCServerMuxer) AcceptKnock(id uint32) {
	m.rootMuxer.acceptKnock(id)
}

func (m *GRPCServerMuxer) Close() error {
	return m.rootMuxer.Close()
}
