package server

import (
	"context"

	pbt "github.com/noncepad/solpipe-market/go/proto/txproc"
)

func (e1 *external) RentExemption(ctx context.Context, req *pbt.RentRequest) (resp *pbt.RentResponse, err error) {
	resp, err = e1.txprocClient.RentExemption(ctx, req)
	return
}
