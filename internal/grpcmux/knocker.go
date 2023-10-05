package grpcmux

import "context"

var _ KnockerServer = (*knocker)(nil)

type knocker struct {
	UnimplementedKnockerServer

	knockCh chan int64
}

func (k knocker) Knock(ctx context.Context, req *KnockRequest) (*KnockResponse, error) {
	k.knockCh <- req.ServiceId
	return &KnockResponse{}, nil
}
