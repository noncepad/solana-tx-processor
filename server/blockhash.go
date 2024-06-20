package server

import (
	"context"
	"log"
	"time"

	sgorpc "github.com/gagliardetto/solana-go/rpc"
	pbt "github.com/noncepad/solana-tx-processor/proto/txproc"
)

func (e1 *external) loopBlock(
	client *sgorpc.Client,
	updateInterval time.Duration,
) {
	defer e1.cancel()
	doneC := e1.ctx.Done()

out:
	for {
		select {
		case <-doneC:
			break out
		case <-time.After(updateInterval):
		}
		// changed from CommitmentConfirmed --> block not being recognized by the lead validator
		out, err := client.GetLatestBlockhash(e1.ctx, sgorpc.CommitmentFinalized)
		if err != nil {
			log.Printf("failed to fetch blockhash: %s", err)
			continue
		}
		e1.m.Lock()
		e1.hash = out.Value.Blockhash
		e1.m.Unlock()

		err = e1.rent.findRent(e1.ctx, client)
		if err != nil {
			log.Printf("failed to update rent: %s", err)
		}
	}
}

func (e1 *external) Blockhash(ctx context.Context, req *pbt.Empty) (resp *pbt.BlockhashResponse, err error) {

	e1.m.RLock()
	hash := e1.hash
	e1.m.RUnlock()
	resp = new(pbt.BlockhashResponse)
	resp.Hash = make([]byte, len(hash))
	copy(resp.Hash, hash[:])
	return resp, nil
}
