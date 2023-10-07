package grpcmux

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/yamux"
	"io"
	"net"
	"sync"
)

var _ net.Listener = (*rootMuxer)(nil)

type rootMuxer struct {
	logger    hclog.Logger
	addr      net.Addr
	isClient  bool
	sessionFn func() (*yamux.Session, error)

	knockCh        chan uint32
	acceptChannels map[uint32]chan acceptResult

	knockMutex sync.Mutex
	knockID    uint32

	// dialMutex ensures we only allow one ID to be knocked and dialled at a time.
	// This prevents data races if for some reason two gRPC connections decide
	// they need to establish a new connection at the same time.
	dialMutex sync.Mutex
}

func newRootMuxer(logger hclog.Logger, addr net.Addr, isClient bool, sessionFn func() (*yamux.Session, error)) *rootMuxer {
	m := &rootMuxer{
		logger:         logger,
		addr:           addr,
		isClient:       isClient,
		sessionFn:      sessionFn,
		knockCh:        make(chan uint32, 1),
		acceptChannels: make(map[uint32]chan acceptResult),
	}

	if isClient {
		go func() {
			_, err := m.Accept()
			if err != nil && err != io.EOF {
				m.logger.Error("received error in client mux Accept loop", "error", err)
			}
		}()
	}

	return m
}

// Accept accepts all incoming connections and routes them to the correct
// stream ID based on the most recent knock received.
func (m *rootMuxer) Accept() (net.Conn, error) {
	session, err := m.sessionFn()
	if err != nil {
		return nil, fmt.Errorf("error establishing yamux session: %w", err)
	}

	for {
		stream, acceptErr := session.AcceptStream()

		select {
		case id := <-m.knockCh:
			acceptCh, ok := m.acceptChannels[id]
			if !ok {
				if stream != nil {
					_ = stream.Close()
				}
				return nil, fmt.Errorf("received knock on ID %d that doesn't have a listener", id)
			}
			m.logger.Debug("sending conn to brokered listener", "client", m.isClient, "id", id)
			acceptCh <- acceptResult{
				conn: stream,
				err:  acceptErr,
			}
		default:
			if m.isClient {
				if acceptErr != nil {
					return nil, err
				}
				m.logger.Error("accepted new stream without a knock handshake")
				continue
			} else {
				m.logger.Debug("sending conn to default listener")
				return stream, acceptErr
			}
		}
	}
}

func (m *rootMuxer) Addr() net.Addr {
	return m.addr
}

func (m *rootMuxer) Close() error {
	session, err := m.sessionFn()
	if err != nil {
		return err
	}

	return session.Close()
}

func (m *rootMuxer) dial() (net.Conn, error) {
	sess, err := m.sessionFn()
	if err != nil {
		return nil, err
	}

	return sess.Open()
}

// acceptKnock sets the stream ID the next accepted connection will go to.
func (m *rootMuxer) acceptKnock(id uint32) {
	m.knockCh <- id
}

// knockAndDial makes an outgoing connection on a specific stream ID.
// It uses knockFn to sync with the server before it starts dialling.
func (m *rootMuxer) knockAndDial(id uint32, knockFn func(id uint32) error) (net.Conn, error) {
	sess, err := m.sessionFn()
	if err != nil {
		return nil, err
	}

	m.dialMutex.Lock()
	defer m.dialMutex.Unlock()

	// Tell the client the gRPC broker ID it should map the next stream to.
	err = knockFn(id)
	if err != nil {
		return nil, fmt.Errorf("failed to knock before dialling client: %w", err)
	}

	m.logger.Debug("dialling new stream...", "client", m.isClient)
	stream, err := sess.OpenStream()
	if err != nil {
		return nil, fmt.Errorf("error dialling new stream: %w", err)
	}

	m.logger.Debug("dialled new stream", "client", m.isClient, "id", stream.StreamID())
	return stream, nil
}

func (m *rootMuxer) listener(id uint32, listenForKnocksFn func(context.Context, uint32) error) (net.Listener, error) {
	sess, err := m.sessionFn()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		err := listenForKnocksFn(ctx, id)
		if err != nil {
			m.logger.Error("error listening for knocks", "client", m.isClient, "id", id, "error", err)
		}
	}()
	ln := newBlockedListener(ctx, cancel, sess.Addr())
	m.acceptChannels[id] = ln.acceptCh
	return ln, nil
}
