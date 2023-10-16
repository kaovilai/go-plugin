package grpcmux

import (
	"net"
)

// GRPCMuxer enables multiple implementations of net.Listener to accept
// connections over a single "main" multiplexed net.Conn, and dial multiple
// client connections over the same multiplexed net.Conn.
//
// The first multiplexed connection is used to serve the gRPC broker's own
// control services: plugin.GRPCBroker, plugin.GRPCController, plugin.GRPCStdio.
//
// Clients must "knock" before dialling, to tell the server side that the
// next net.Conn should be accepted onto a specific stream ID. The knock is a
// bidirectional streaming message on the plugin.GRPCBroker service.
type GRPCMuxer interface {
	Enabled() bool
	Listener(id uint32, listenForKnocksFn func(uint32) error, doneCh <-chan struct{}) (net.Listener, error)
	KnockAndDial(id uint32, knockFn func(uint32) error) (net.Conn, error)
	AcceptKnock(id uint32) error
	Close() error
}
