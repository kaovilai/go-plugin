package grpcmux

import (
	"fmt"
	"net"
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/yamux"
)

var _ GRPCMuxer = (*GRPCClientMuxer)(nil)

// GRPCClientMuxer implements the client (host) side of the gRPC broker's
// GRPCMuxer interface for multiplexing multiple gRPC broker connections over
// a single net.Conn.
//
// The client dials the initial net.Conn eagerly, and creates a yamux.Session
// as the implementation for multiplexing any additional connections.
//
// Each net.Listener returned from Listener will block until the client receives
// a knock that matches its gRPC broker stream ID. There is no default listener
// on the client, as it is a client for the gRPC broker's control services. (See
// GRPCServerMuxer for more details).
type GRPCClientMuxer struct {
	logger  hclog.Logger
	session *yamux.Session

	acceptMutex     sync.Mutex
	acceptListeners map[uint32]*blockedClientListener

	dialMutex sync.Mutex
}

func NewGRPCClientMuxer(logger hclog.Logger, addr net.Addr) (*GRPCClientMuxer, error) {
	// Eagerly establish the underlying connection as early as possible.
	logger.Debug("making new client mux initial connection", "addr", addr)
	conn, err := net.Dial(addr.Network(), addr.String())
	if err != nil {
		return nil, err
	}
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		// Make sure to set keep alive so that the connection doesn't die
		_ = tcpConn.SetKeepAlive(true)
	}

	cfg := yamux.DefaultConfig()
	cfg.Logger = logger.Named("yamux").StandardLogger(&hclog.StandardLoggerOptions{
		InferLevels: true,
	})
	sess, err := yamux.Client(conn, cfg)
	if err != nil {
		return nil, err
	}

	logger.Debug("client muxer connected", "addr", addr)
	m := &GRPCClientMuxer{
		logger:          logger,
		session:         sess,
		acceptListeners: make(map[uint32]*blockedClientListener),
	}

	return m, nil
}

func (m *GRPCClientMuxer) MainDial() (net.Conn, error) {
	m.dialMutex.Lock()
	defer m.dialMutex.Unlock()

	return m.session.Open()
}

func (m *GRPCClientMuxer) Enabled() bool {
	return m != nil
}

func (m *GRPCClientMuxer) Listener(id uint32, listenForKnocksFn func(uint32) error, doneCh <-chan struct{}) (net.Listener, error) {
	go func() {
		err := listenForKnocksFn(id)
		if err != nil {
			m.logger.Error("error listening for knocks", "id", id, "error", err)
		}
	}()
	ln := newBlockedClientListener(m.session, doneCh)

	m.acceptMutex.Lock()
	m.acceptListeners[id] = ln
	m.acceptMutex.Unlock()

	return ln, nil
}

func (m *GRPCClientMuxer) KnockAndDial(id uint32, knockFn func(id uint32) error) (net.Conn, error) {
	m.dialMutex.Lock()
	defer m.dialMutex.Unlock()

	// Tell the client the gRPC broker ID it should map the next stream to.
	err := knockFn(id)
	if err != nil {
		return nil, fmt.Errorf("failed to knock before dialling client: %w", err)
	}

	conn, err := m.session.Open()
	if err != nil {
		return nil, fmt.Errorf("error dialling new stream: %w", err)
	}

	return conn, nil
}

func (m *GRPCClientMuxer) AcceptKnock(id uint32) error {
	m.acceptMutex.Lock()
	defer m.acceptMutex.Unlock()

	ln, ok := m.acceptListeners[id]
	if !ok {
		return fmt.Errorf("no listener for id %d", id)
	}
	ln.unblock()
	return nil
}

func (m *GRPCClientMuxer) Close() error {
	return m.session.Close()
}
