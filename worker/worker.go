package worker

import (
	"context"

	"github.com/noncepad/worker-pool/pool"
)

// Request specifies the structure for job requests
type Request struct {
	Tx []byte
}

// Holds a list of strings representing files.
type Result struct {
	Success bool
}

// Implementation of a worker capable of processing Jobs.
type simpleWorker struct {
	ctx    context.Context    // context for cancellation
	cancel context.CancelFunc // call when cancelling the worker
	url    string
}

func Create(parentCtx context.Context, url string) (pool.Worker[Request, Result], error) {
	e1 := new(simpleWorker)
	e1.ctx, e1.cancel = context.WithCancel(parentCtx)
	e1.url = url
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
