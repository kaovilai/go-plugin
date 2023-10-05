package grpcmux

import (
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/yamux"
	"net"
	"sync"
)

type muxer struct {
	logger hclog.Logger

	knockCh        chan int64
	acceptChannels map[int64]chan acceptResult

	// dialMutex ensures we only allow one ID to be knocked and dialled at a time.
	// This prevents data races if for some reason two gRPC connections decide
	// they need to establish a new connection at the same time.
	dialMutex sync.Mutex
}

// acceptLoop accepts all incoming connections and routes them to the correct
// stream ID based on the most recent knock received.
func (m *muxer) acceptLoop(sess *yamux.Session) {
	for {
		m.logger.Debug("accepting new stream...")
		stream, acceptErr := sess.AcceptStream()
		result := acceptResult{
			conn: stream,
			err:  acceptErr,
		}

		m.logger.Debug("accepted new stream", "id", stream.StreamID())
		select {
		case id := <-m.knockCh:
			acceptCh, ok := m.acceptChannels[id]
			if !ok {
				stream.Close()
				m.logger.Error("received knock on ID that doesn't have a listener", "id", id)
				continue
			}
			m.logger.Debug("sending conn to brokered listener", "id", id)
			acceptCh <- result
		default:
			acceptCh, ok := m.acceptChannels[0]
			if !ok {
				stream.Close()
				m.logger.Error("no default gRPC mux connection")
				continue
			}
			m.logger.Debug("sending conn to default listener")
			acceptCh <- result
		}

		// If we can't accept connections, we can't multiplex anything.
		if acceptErr != nil {
			return
		}
	}
}

func (m *muxer) knockAndDial(sess *yamux.Session, id uint32) (net.Conn, error) {
	m.dialMutex.Lock()
	defer m.dialMutex.Unlock()

	// Tell the client the gRPC broker ID it should map the next stream to.
	//err := knockFn(id)
	//if err != nil {
	//	return nil, fmt.Errorf("failed to knock before dialling client: %w", err)
	//}

	m.logger.Debug("dialling new server stream...")
	stream, err := sess.OpenStream()
	if err != nil {
		return nil, fmt.Errorf("error dialling new server stream: %w", err)
	}

	m.logger.Debug("dialled new server stream", "id", stream.StreamID())
	return stream, nil
}

func (m *muxer) listener(sess *yamux.Session, id int64) net.Listener {
	ln := newBlockedListener(sess.Addr())
	m.acceptChannels[id] = ln.acceptCh
	return ln
}
