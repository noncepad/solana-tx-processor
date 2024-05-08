package server

import (
	"context"
	"errors"
	"fmt"
	"sync"

	sgorpc "github.com/gagliardetto/solana-go/rpc"
	pbt "github.com/noncepad/solpipe-market/go/proto/txproc"
)

func (e1 *external) RentExemption(ctx context.Context, req *pbt.RentRequest) (resp *pbt.RentResponse, err error) {
	if req == nil {
		err = errors.New("blank request")
		return
	}
	resp = new(pbt.RentResponse)
	resp.Lamports = e1.rent.rent(req.Size)
	return
}

// a_0 + a_1 * size
type rentCalculator struct {
	m   *sync.RWMutex
	a_0 float64
	a_1 float64
}

func (rc *rentCalculator) rent(size uint64) uint64 {
	// put an upperbound on size
	if 10_000_000 < size {
		size = 10_000_000
	}
	rc.m.RLock()
	defer rc.m.RUnlock()
	return uint64(rc.a_0 + rc.a_1*float64(size))
}

const b_1 uint64 = 1_000

func (rc *rentCalculator) findRent(ctx context.Context, client *sgorpc.Client) error {

	// formula y=mx + b
	// x is the size in bytes

	g_0, err := client.GetMinimumBalanceForRentExemption(ctx, 0, sgorpc.CommitmentFinalized)
	if err != nil {
		return fmt.Errorf("failed to update rent - 1: %s", err)

	}

	f_0 := float64(g_0)
	g_1, err := client.GetMinimumBalanceForRentExemption(ctx, b_1, sgorpc.CommitmentFinalized)
	if err != nil {
		return fmt.Errorf("failed to update rent - 1: %s", err)
	}
	f_1 := float64(g_1)

	rc.m.Lock()
	defer rc.m.Unlock()
	rc.a_0 = f_0
	rc.a_1 = (f_1 - f_0) / float64(b_1)

	return nil
}
