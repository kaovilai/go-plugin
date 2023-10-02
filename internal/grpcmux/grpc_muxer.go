package grpcmux

import "net"

type GRPCMuxer interface {
	Dial() (net.Conn, error)
	net.Listener
}
