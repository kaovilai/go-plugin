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
	logger hclog.Logger

	sessionErrCh chan error
	sess         *yamux.Session

	muxCommon *muxer
}

func NewGRPCServerMuxer(logger hclog.Logger, ln net.Listener) *GRPCServerMuxer {
	m := &GRPCServerMuxer{
		addr:   ln.Addr(),
		logger: logger,

		sessionErrCh: make(chan error),
		muxCommon: &muxer{
			logger:         logger,
			knockCh:        make(chan int64),
			acceptChannels: make(map[int64]chan acceptResult),
		},
	}

	go m.acceptSessionAndMuxAccept(ln)

	return m
}

// acceptSessionAndMuxAccept is responsible for establishing the yamux session,
// and then kicking off the acceptLoop function.
func (m *GRPCServerMuxer) acceptSessionAndMuxAccept(ln net.Listener) {
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
	go m.muxCommon.acceptLoop(m.sess)
}

func (m *GRPCServerMuxer) session() (*yamux.Session, error) {
	select {
	case err := <-m.sessionErrCh:
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

func (m *GRPCServerMuxer) Dial(id uint32) (net.Conn, error) {
	sess, err := m.session()
	if err != nil {
		return nil, fmt.Errorf("error getting session for server Dial: %w", err)
	}

	return m.muxCommon.knockAndDial(sess, id)
}

func (m *GRPCServerMuxer) Close() error {
	m.logger.Debug("closing server muxer")
	if m.sess != nil {
		return m.sess.Close()
	}

	return nil
}

func (m *GRPCServerMuxer) Listener(id uint32) (net.Listener, error) {
	sess, err := m.session()
	if err != nil {
		return nil, err
	}

	return m.muxCommon.listener(sess, int64(id)), nil
}
