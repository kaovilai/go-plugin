package grpcmux

import (
	"context"
	"github.com/hashicorp/yamux"
	"io"
	"net"
)

var _ net.Listener = (*blockedListener)(nil)

// blockedListener uses a yamux.Session to implement the net.Listener interface
// with one addition: it requires the client to "knock" first. We can't control
// order in which the gRPC servers will call Accept() on our multiplexed listener,
// but we do need to control which gRPC server accepts which connection, so we
// use a blocked listener and a knock from the client to select the gRPC server
// we dial to. The selection is based on the gRPC broker's connection ID, and is
// handled one layer higher in the client and server muxer structs.
type blockedListener struct {
	addr    net.Addr
	knockCh chan struct{}
	doneCtx context.Context
	cancel  func()
	sess    *yamux.Session
}

func newBlockedListener(session *yamux.Session) *blockedListener {
	ctx, cancel := context.WithCancel(context.Background())
	return &blockedListener{
		addr:    session.Addr(),
		knockCh: make(chan struct{}),
		doneCtx: ctx,
		cancel:  cancel,
		sess:    session,
	}
}

func (b blockedListener) Accept() (net.Conn, error) {
	select {
	case <-b.knockCh:
		// Unblock.
	case <-b.doneCtx.Done():
		return nil, io.EOF
	}

	stream, err := b.sess.AcceptStream()
	if err != nil {
		return nil, err
	}

	return stream, nil
}

func (b blockedListener) Addr() net.Addr {
	return b.addr
}

func (b blockedListener) Close() error {
	b.cancel()
	return nil
}
