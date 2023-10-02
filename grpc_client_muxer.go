package plugin

import (
	"fmt"
	"net"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/yamux"
)

var _ grpcMuxer = (*grpcClientMuxer)(nil)

type grpcClientMuxer struct {
	logger  hclog.Logger
	session *yamux.Session
}

func newGRPCClientMuxer(logger hclog.Logger, addr net.Addr) (*grpcClientMuxer, error) {
	logger.Debug("making new client mux initial connection")
	conn, err := netAddrDialer(addr)("unused", 0)
	if err != nil {
		return nil, err
	}

	sess, err := yamux.Client(conn, nil)
	if err != nil {
		return nil, err
	}

	logger.Debug("client connected", "addr", addr)
	return &grpcClientMuxer{
		logger:  logger,
		session: sess,
	}, nil
}

func (m *grpcClientMuxer) Dial() (net.Conn, error) {
	stream, err := m.session.OpenStream()
	if err != nil {
		return nil, fmt.Errorf("error dialling new client stream: %w", err)
	}

	m.logger.Debug("dialled new client stream", "id", stream.StreamID())
	return stream, nil
}

func (m *grpcClientMuxer) Accept() (net.Conn, error) {
	stream, err := m.session.AcceptStream()
	if err != nil {
		return nil, fmt.Errorf("error accepting new client stream: %w", err)
	}

	m.logger.Debug("accepted new client stream", "id", stream.StreamID())
	return stream, nil
}

func (m *grpcClientMuxer) Addr() net.Addr {
	return m.session.Addr()
}

func (m *grpcClientMuxer) Close() error {
	m.logger.Debug("closing client muxer")
	return m.session.Close()
}
