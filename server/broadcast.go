package server

import (
	"context"
	"errors"

	"github.com/noncepad/solana-tx-processor/worker"
	pbt "github.com/noncepad/solpipe-market/go/proto/txproc"
)

func (e1 *external) Broadcast(ctx context.Context, req *pbt.BroadcastRequest) (resp *pbt.TransactionResult, err error) {
	if req == nil {
		err = errors.New("blank request")
		return
	}
	if req.Transaction == nil {
		err = errors.New("blank transaction")
		return
	}
	if len(req.Transaction) == 0 {
		err = errors.New("blank transaction")
		return
	}
	var r worker.Result
	r, err = e1.mgr.Submit(ctx, worker.Request{
		Tx:       req.Transaction,
		Simulate: req.Simulate,
	})
	if err != nil {
		return
	}
	resp = new(pbt.TransactionResult)
	resp.Signature = make([]byte, len(r.Signature))
	copy(resp.Signature, r.Signature[:])
	resp.Slot = r.Slot
	return

}
