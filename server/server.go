package server

import (
	"context"
	"errors"

	pbt "github.com/noncepad/solana-tx-processor/proto/txproc"
	"github.com/noncepad/solana-tx-processor/worker"
	"github.com/noncepad/worker-pool/manager"
	"github.com/noncepad/worker-pool/pool"
	"google.golang.org/grpc"
)

type Configuration struct {
	Rpc         string `json:"rpc"`
	Ws          string `json:"ws"`
	WorkerCount int    `json:"count"`
}

type external struct {
	pbt.UnimplementedTransactionProcessingServer
	ctx    context.Context
	cancel context.CancelFunc
	mgr    pool.Manager[worker.Request, worker.Result]
}

func Run(
	parentCtx context.Context,
	config *Configuration,
	s *grpc.Server,
) error {
	ctx, cancel := context.WithCancel(parentCtx)
	mgr, err := manager.Create[worker.Request, worker.Result](ctx, 1)
	if err != nil {
		cancel()
		return err
	}

	for i := 0; i < config.WorkerCount; i++ {
		w, err := worker.Create(ctx, config.Rpc, config.Ws)
		if err != nil {
			cancel()
			return err
		}
		mgr.Add(w)
	}
	pbt.RegisterTransactionProcessingServer(s, external{
		ctx: ctx, cancel: cancel, mgr: mgr,
	})
	return nil
}

func (e1 external) Broadcast(ctx context.Context, req *pbt.BroadcastRequest) (resp *pbt.TransactionResult, err error) {
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
