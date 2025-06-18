package server

import (
	"context"

	pbt "github.com/noncepad/solpipe-market/go/proto/txproc"
)

func (e1 *external) Blockhash(ctx context.Context, req *pbt.Empty) (resp *pbt.BlockhashResponse, err error) {
	resp, err = e1.txprocClient.Blockhash(ctx, req)
	return
}
