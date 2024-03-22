package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	sgo "github.com/gagliardetto/solana-go"
	sgorpc "github.com/gagliardetto/solana-go/rpc"
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
	m      *sync.RWMutex
	hash   sgo.Hash
	rent   *rentCalculator
}

func Run(
	parentCtx context.Context,
	config *Configuration,
	s *grpc.Server,
) error {

	client := sgorpc.New(config.Rpc)
	out, err := client.GetLatestBlockhash(parentCtx, sgorpc.CommitmentConfirmed)
	if err != nil {
		return fmt.Errorf("failed to fetch blockhash: %s", err)
	}
	rc := new(rentCalculator)
	rc.m = &sync.RWMutex{}
	err = rc.findRent(parentCtx, client)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(parentCtx)
	mgrTxSend, err := manager.Create[worker.Request, worker.Result](ctx, 1)
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
		mgrTxSend.Add(w)
	}
	e1 := new(external)
	e1.ctx = ctx
	e1.cancel = cancel
	e1.mgr = mgrTxSend
	e1.m = &sync.RWMutex{}
	e1.hash = out.Value.Blockhash
	e1.rent = rc
	pbt.RegisterTransactionProcessingServer(s, e1)
	go e1.loopBlock(client, 1*time.Minute)
	return nil
}
