package grpcmux

import (
	"context"
	"github.com/hashicorp/yamux"
	"io"
	"net"
)

var _ net.Listener = (*blockedClientListener)(nil)

type blockedClientListener struct {
	session *yamux.Session
	waitCh  chan struct{}
	doneCtx context.Context
	cancel  func()
}

func newBlockedClientListener(ctx context.Context, cancel func(), session *yamux.Session) *blockedClientListener {
	return &blockedClientListener{
		waitCh:  make(chan struct{}, 1),
		session: session,
		doneCtx: ctx,
		cancel:  cancel,
	}
}

func (b *blockedClientListener) Accept() (net.Conn, error) {
	select {
	case <-b.waitCh:
		return b.session.Accept()
	case <-b.doneCtx.Done():
		return nil, io.EOF
	}
}

func (b *blockedClientListener) Addr() net.Addr {
	return b.session.Addr()
}

func (b *blockedClientListener) Close() error {
	b.cancel()
	// We don't close the session, the client muxer is responsible for that.
	return nil
}

func (b *blockedClientListener) unblock() {
	b.waitCh <- struct{}{}
}
