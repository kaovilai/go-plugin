package grpcmux

import (
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/yamux"
	"net"
)

var _ GRPCMuxer = (*GRPCClientMuxer)(nil)

type GRPCClientMuxer struct {
	logger hclog.Logger

	session *yamux.Session
	addr    net.Addr
}

func NewGRPCClientMuxer(logger hclog.Logger, addr net.Addr) (*GRPCClientMuxer, error) {
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
	return &GRPCClientMuxer{
		logger:  logger,
		session: sess,
	}, nil
}

func (m *GRPCClientMuxer) Dial() (net.Conn, error) {
	m.logger.Debug("dialling new client stream...")
	stream, err := m.session.OpenStream()
	if err != nil {
		return nil, fmt.Errorf("error dialling new client stream: %w", err)
	}

	m.logger.Debug("dialled new client stream", "id", stream.StreamID())
	return stream, nil
}

func (m *GRPCClientMuxer) Accept() (net.Conn, error) {
	m.logger.Debug("accepting new client stream...")
	stream, err := m.session.AcceptStream()
	if err != nil {
		return nil, fmt.Errorf("error accepting new client stream: %w", err)
	}

	m.logger.Debug("accepted new client stream", "id", stream.StreamID())
	return stream, nil
}

func (m *GRPCClientMuxer) Addr() net.Addr {
	return m.session.Addr()
}

func (m *GRPCClientMuxer) Close() error {
	m.logger.Debug("closing client muxer")
	return m.session.Close()
}
