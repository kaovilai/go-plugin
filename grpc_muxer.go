package plugin

import "net"

type grpcMuxer interface {
	Dial() (net.Conn, error)
	net.Listener
}
