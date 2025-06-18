package worker

import (
	"context"

	sgo "github.com/gagliardetto/solana-go"
	sgorpc "github.com/gagliardetto/solana-go/rpc"
	sgows "github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/noncepad/worker-pool/pool"
)

// Request specifies the structure for job requests
type Request struct {
	Tx       []byte
	Simulate bool
}

// Holds a list of strings representing files.
type Result struct {
	Signature sgo.Signature
	Slot      uint64
}

// Implementation of a worker capable of processing Jobs.
type simpleWorker struct {
	ctx       context.Context    // context for cancellation
	cancel    context.CancelFunc // call when cancelling the worker
	rpc       *sgorpc.Client
	txSendRpc *sgorpc.Client
	ws        *sgows.Client
}

func Create(parentCtx context.Context, rpcUrl string, txSendRpcUrl string, wsUrl string) (pool.Worker[Request, Result], error) {
	e1 := new(simpleWorker)
	e1.ctx, e1.cancel = context.WithCancel(parentCtx)
	e1.rpc = sgorpc.New(rpcUrl)
	e1.txSendRpc = sgorpc.New(txSendRpcUrl)
	var err error
	e1.ws, err = sgows.Connect(e1.ctx, wsUrl)
	if err != nil {
		return nil, err
	}
	return e1, nil
}

// Terminate the worker by cancelling its context
func (sw *simpleWorker) Close() error {
	sw.cancel()
	return sw.ctx.Err()
}

// When context is cancelled indicates and error with LoopError
func (sw *simpleWorker) CloseSignal() <-chan error {
	signalC := make(chan error, 1)
	go loopError(sw.ctx, signalC)
	return signalC
}

// Listens for the context's done signal and relays the context's error to the provided channel.
func loopError(ctx context.Context, errorC chan<- error) {
	<-ctx.Done()
	errorC <- ctx.Err()
}
