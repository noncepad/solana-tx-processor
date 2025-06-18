package server

import (
	"context"

	"github.com/noncepad/solana-tx-processor/worker"
	pbt "github.com/noncepad/solpipe-market/go/proto/txproc"
	"github.com/noncepad/worker-pool/manager"
	"github.com/noncepad/worker-pool/meter"
	"github.com/noncepad/worker-pool/pool"
	"google.golang.org/grpc"
)

type Configuration struct {
	TxSenderRpc string `json:"txrpc"`
	WorkerCount int    `json:"count"`
}

type external struct {
	pbt.UnimplementedTransactionProcessingServer
	ctx          context.Context
	cancel       context.CancelFunc
	mgr          pool.Manager[worker.Request, worker.Result]
	txprocClient pbt.TransactionProcessingClient
}

func Run(
	parentCtx context.Context,
	config *Configuration,
	s *grpc.Server,
	txprocClient pbt.TransactionProcessingClient,
) error {
	ctx, cancel := context.WithCancel(parentCtx)
	mgrTxSend, err := manager.Create[worker.Request, worker.Result](ctx, 1, meter.Create(s))
	if err != nil {
		cancel()
		return err
	}

	for i := 0; i < config.WorkerCount; i++ {
		w, err := worker.Create(ctx, config.TxSenderRpc)
		if err != nil {
			cancel()
			return err
		}
		mgrTxSend.Add(w)
	}
	e1 := new(external)
	e1.ctx = ctx
	e1.cancel = cancel
	e1.mgr = mgrTxSend
	e1.txprocClient = txprocClient
	pbt.RegisterTransactionProcessingServer(s, e1)
	return nil
}
