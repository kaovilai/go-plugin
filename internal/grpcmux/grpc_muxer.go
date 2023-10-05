package grpcmux

import "net"

type GRPCMuxer interface {
	Dial(id uint32) (net.Conn, error)
	Listener(id uint32) (net.Listener, error)
	Close() error
}
