package plugin

import (
	"net"

	"github.com/hashicorp/yamux"
)

var _ grpcMuxer = (*grpcClientMuxer)(nil)

type grpcClientMuxer struct {
	session *yamux.Session
}

func newGRPCClientMuxer(addr net.Addr) (*grpcClientMuxer, error) {
	conn, err := net.Dial(addr.Network(), addr.String())
	if err != nil {
		return nil, err
	}

	sess, err := yamux.Client(conn, nil)
	if err != nil {
		return nil, err
	}

	return &grpcClientMuxer{
		session: sess,
	}, nil
}

func (m *grpcClientMuxer) Session() (*yamux.Session, error) {
	return m.session, nil
}

func (m *grpcClientMuxer) Close() error {
	return m.session.Close()
}
