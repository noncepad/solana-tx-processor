package worker

import (
	"context"
	"errors"
	"fmt"

	sgo "github.com/gagliardetto/solana-go"
	sgorpc "github.com/gagliardetto/solana-go/rpc"
	"github.com/noncepad/worker-pool/pool"
)

func (sw *simpleWorker) Run(job pool.Job[Request]) (r Result, err error) {
	ctx := job.Ctx()
	payload := job.Payload()
	var slot uint64
	var sig sgo.Signature
	sig, slot, err = sw.send(ctx, payload.Tx, payload.Simulate)
	if err != nil {
		return
	}
	r = Result{Slot: slot, Signature: sig}

	return
}

type JsonRpcRequest struct {
	Jsonrpc string        `json name:"jsonrpc"`
	Id      int           `json name:"id"`
	Method  string        `json name:"method"`
	Params  []interface{} `json name:"params"`
}

func (sw *simpleWorker) send(ctx context.Context, data []byte, simulate bool) (sig sgo.Signature, slot uint64, err error) {
	if data == nil {
		err = errors.New("no data to send")
		return
	}

	sig, err = sw.rpc.SendRawTransactionWithOpts(ctx, data, sgorpc.TransactionOpts{
		Encoding:            sgo.EncodingBase64,
		SkipPreflight:       !simulate,
		PreflightCommitment: sgorpc.CommitmentProcessed,
	})
	if err != nil {
		return
	}
	sub, err := sw.ws.SignatureSubscribe(sig, sgorpc.CommitmentFinalized)
	if err != nil {
		return
	}
	defer sub.Unsubscribe()
	select {
	case <-ctx.Done():
		err = ctx.Err()
	case err = <-sub.Err():
	case result := <-sub.Response():
		if result.Value.Err != nil {
			err = fmt.Errorf("tx failed to confirm: %+v", result.Value.Err)
		} else {
			slot = result.Context.Slot
		}
	}
	return
}
