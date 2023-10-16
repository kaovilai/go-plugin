package grpcmux

import (
	"github.com/hashicorp/yamux"
	"io"
	"net"
)

var _ net.Listener = (*blockedClientListener)(nil)

type blockedClientListener struct {
	session *yamux.Session
	waitCh  chan struct{}
	doneCh  <-chan struct{}
}

func newBlockedClientListener(session *yamux.Session, doneCh <-chan struct{}) *blockedClientListener {
	return &blockedClientListener{
		waitCh:  make(chan struct{}, 1),
		doneCh:  doneCh,
		session: session,
	}
}

func (b *blockedClientListener) Accept() (net.Conn, error) {
	select {
	case <-b.waitCh:
		return b.session.Accept()
	case <-b.doneCh:
		return nil, io.EOF
	}
}

func (b *blockedClientListener) Addr() net.Addr {
	return b.session.Addr()
}

func (b *blockedClientListener) Close() error {
	// We don't close the session, the client muxer is responsible for that.
	return nil
}

func (b *blockedClientListener) unblock() {
	b.waitCh <- struct{}{}
}
