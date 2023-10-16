package grpcmux

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/yamux"
)

var _ GRPCMuxer = (*GRPCServerMuxer)(nil)
var _ net.Listener = (*GRPCServerMuxer)(nil)

type GRPCServerMuxer struct {
	addr   net.Addr
	logger hclog.Logger

	sessionErrCh chan error
	sess         *yamux.Session

	knockCh chan uint32

	acceptMutex    sync.Mutex
	acceptChannels map[uint32]chan acceptResult

	dialMutex sync.Mutex
}

func NewGRPCServerMuxer(logger hclog.Logger, ln net.Listener) *GRPCServerMuxer {
	m := &GRPCServerMuxer{
		addr:   ln.Addr(),
		logger: logger,

		sessionErrCh: make(chan error),

		knockCh:        make(chan uint32, 1),
		acceptChannels: make(map[uint32]chan acceptResult),
	}

	go m.acceptSession(ln)

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

// Accept accepts all incoming connections and routes them to the correct
// stream ID based on the most recent knock received.
func (m *GRPCServerMuxer) Accept() (net.Conn, error) {
	session, err := m.session()
	if err != nil {
		return nil, fmt.Errorf("error establishing yamux session: %w", err)
	}

	for {
		conn, acceptErr := session.Accept()

		select {
		case id := <-m.knockCh:
			m.acceptMutex.Lock()
			acceptCh, ok := m.acceptChannels[id]
			m.acceptMutex.Unlock()

			if !ok {
				if conn != nil {
					_ = conn.Close()
				}
				return nil, fmt.Errorf("received knock on ID %d that doesn't have a listener", id)
			}
			m.logger.Debug("sending conn to brokered listener", "id", id)
			acceptCh <- acceptResult{
				conn: conn,
				err:  acceptErr,
			}
		default:
			m.logger.Debug("sending conn to default listener")
			return conn, acceptErr
		}
	}
}

func (m *GRPCServerMuxer) Addr() net.Addr {
	return m.addr
}

func (m *GRPCServerMuxer) Close() error {
	session, err := m.session()
	if err != nil {
		return err
	}

	return session.Close()
}

func (m *GRPCServerMuxer) Enabled() bool {
	return m != nil
}

func (m *GRPCServerMuxer) Listener(id uint32, listenForKnocksFn func(context.Context, uint32) error) (net.Listener, error) {
	sess, err := m.session()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		err := listenForKnocksFn(ctx, id)
		if err != nil {
			m.logger.Error("error listening for knocks", "id", id, "error", err)
		}
	}()
	ln := newBlockedServerListener(ctx, cancel, sess.Addr())
	m.acceptMutex.Lock()
	m.acceptChannels[id] = ln.acceptCh
	m.acceptMutex.Unlock()

	return ln, nil
}

func (m *GRPCServerMuxer) KnockAndDial(id uint32, knockFn func(id uint32) error) (net.Conn, error) {
	sess, err := m.session()
	if err != nil {
		return nil, err
	}

	m.dialMutex.Lock()
	defer m.dialMutex.Unlock()

	// Tell the client the gRPC broker ID it should map the next stream to.
	err = knockFn(id)
	if err != nil {
		return nil, fmt.Errorf("failed to knock before dialling client: %w", err)
	}

	stream, err := sess.OpenStream()
	if err != nil {
		return nil, fmt.Errorf("error dialling new stream: %w", err)
	}

	return stream, nil
}

func (m *GRPCServerMuxer) AcceptKnock(id uint32) error {
	m.knockCh <- id
	return nil
}
