package grpcmux

import (
	"context"
	"io"
	"net"
)

var _ net.Listener = (*serverBlockedListener)(nil)

// serverBlockedListener uses a yamux.Session to implement the net.Listener interface
// with one addition: it requires the client to "knock" first. We can't control
// order in which the gRPC servers will call Accept() on our multiplexed listener,
// but we do need to control which gRPC server accepts which connection, so we
// use a blocked listener and a knock from the client to select the gRPC server
// we dial to. The selection is based on the gRPC broker's connection ID, and is
// handled one layer higher in the client and server muxer structs.
type serverBlockedListener struct {
	addr     net.Addr
	acceptCh chan acceptResult
	doneCtx  context.Context
	cancel   func()
}

type acceptResult struct {
	conn net.Conn
	err  error
}

func newBlockedServerListener(ctx context.Context, cancel func(), addr net.Addr) *serverBlockedListener {
	return &serverBlockedListener{
		addr:     addr,
		acceptCh: make(chan acceptResult),
		doneCtx:  ctx,
		cancel:   cancel,
	}
}

func (b serverBlockedListener) Accept() (net.Conn, error) {
	select {
	case accept := <-b.acceptCh:
		return accept.conn, accept.err
	case <-b.doneCtx.Done():
		return nil, io.EOF
	}
}

func (b serverBlockedListener) Addr() net.Addr {
	return b.addr
}

func (b serverBlockedListener) Close() error {
	b.cancel()
	return nil
}
