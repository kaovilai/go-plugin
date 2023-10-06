package grpcmux

import (
	"context"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/yamux"
	"net"
)

var _ GRPCMuxer = (*GRPCClientMuxer)(nil)

type GRPCClientMuxer struct {
	logger hclog.Logger
	mux    *rootMuxer
}

func NewGRPCClientMuxer(logger hclog.Logger, addr net.Addr) (*GRPCClientMuxer, error) {
	// Eagerly establish the underlying connection as early as possible.
	logger.Debug("making new client mux initial connection")
	conn, err := net.Dial(addr.Network(), addr.String())
	if err != nil {
		return nil, err
	}
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		// Make sure to set keep alive so that the connection doesn't die
		_ = tcpConn.SetKeepAlive(true)
	}

	sess, err := yamux.Client(conn, nil)
	if err != nil {
		return nil, err
	}

	logger.Debug("client muxer connected", "addr", addr)
	m := &GRPCClientMuxer{
		logger: logger,
		mux: newRootMuxer(logger, addr, true, func() (*yamux.Session, error) {
			return sess, nil
		}),
	}

	return m, nil
}

func (m *GRPCClientMuxer) MainDial() (net.Conn, error) {
	return m.mux.dial()
}

func (m *GRPCClientMuxer) Listener(id uint32, listenForKnocksFn func(context.Context, uint32) error) (net.Listener, error) {
	m.logger.Debug("accepting new client stream...")
	return m.mux.listener(id, listenForKnocksFn)
}

func (m *GRPCClientMuxer) KnockAndDial(id uint32, knockFn func(id uint32) error) (net.Conn, error) {
	m.logger.Debug("dialling new client stream...")
	return m.mux.knockAndDial(id, knockFn)
}

func (m *GRPCClientMuxer) AcceptKnock(id uint32) {
	m.mux.acceptKnock(id)
}

func (m *GRPCClientMuxer) Close() error {
	m.logger.Debug("closing client muxer")
	return m.mux.Close()
}
