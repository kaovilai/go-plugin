package plugin

import (
	"github.com/hashicorp/yamux"
)

type grpcMuxer interface {
	Session() (*yamux.Session, error)
	Close() error
}
