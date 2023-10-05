package grpcmux

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/yamux"
	"google.golang.org/grpc"
	"net"
)

var _ GRPCMuxer = (*GRPCClientMuxer)(nil)

type GRPCClientMuxer struct {
	logger  hclog.Logger
	session *yamux.Session
	mux     *muxer
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
	m := &GRPCClientMuxer{
		logger:  logger,
		session: sess,
		mux: &muxer{
			logger:         logger,
			acceptChannels: make(map[int64]chan acceptResult),
			knockCh:        make(chan int64),
		},
	}

	go m.mux.acceptLoop(sess)
	go m.startKnockerServer()

	return m, nil
}

func (m *GRPCClientMuxer) Dial(id uint32) (net.Conn, error) {
	m.logger.Debug("dialling new client stream...")
	return m.mux.knockAndDial(m.session, id)
}

func (m *GRPCClientMuxer) Listener(id uint32) (net.Listener, error) {
	m.logger.Debug("accepting new client stream...")
	return m.mux.listener(m.session, int64(id)), nil
}

func (m *GRPCClientMuxer) Close() error {
	m.logger.Debug("closing client muxer")
	return m.session.Close()
}

func (m *GRPCClientMuxer) startKnockerServer() {
	ln := m.mux.listener(m.session, -1)
	defer ln.Close()

	server := grpc.NewServer()
	RegisterKnockerServer(server, &knocker{
		knockCh: m.mux.knockCh,
	})

	if err := server.Serve(ln); err != nil {
		m.logger.Error("error serving knocker service", "error", err)
	}
}
