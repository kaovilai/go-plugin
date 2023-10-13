package grpcmux

import (
	"context"
	"net"
)

type GRPCMuxer interface {
	Enabled() bool
	Listener(id uint32, listenForKnocksFn func(context.Context, uint32) error) (net.Listener, error)
	KnockAndDial(id uint32, knockFn func(uint32) error) (net.Conn, error)
	AcceptKnock(id uint32) error
	Close() error
}
