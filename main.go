package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/noncepad/solana-tx-processor/proxy"
	"github.com/noncepad/solana-tx-processor/server"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	CMD_SERVER = "server"
	CMD_CLIENT = "client"
)

func run_client(ctx context.Context, args []string) error {
	if len(args) < 2 {
		return errors.New("not enough args: need listen url and proxy url")
	}
	config := new(proxy.Configuration)
	config.ListenUrl = args[0]
	config.ProxyUrl = args[1]
	return proxy.Run(ctx, config)
}

func run_server(ctx context.Context, args []string) error {
	log.Error("rs - 1")
	var present bool
	var err error
	config := new(server.Configuration)
	if len(args) < 2 {
		return errors.New("no worker count and listen url specified")
	}

	log.Error("rs - 2")
	config.WorkerCount, err = strconv.Atoi(args[0])
	if err != nil {
		return err
	}
	log.Error("rs - 3")
	log.SetLevel(log.DebugLevel)
	var l net.Listener
	url := args[1]
	if strings.HasPrefix(url, "unix://") {
		l, err = net.Listen("unix", strings.TrimPrefix(url, "unix://"))
	} else if strings.HasPrefix(url, "tcp://") {
		l, err = net.Listen("tcp", strings.TrimPrefix(url, "tcp://"))
	} else {
		err = fmt.Errorf("unknown url protocol %s", url)
	}
	if err != nil {
		return err
	}
	config.Rpc, present = os.LookupEnv("RPC_URL")
	if !present {
		return errors.New("no rpc url specified")
	}
	config.TxSenderRpc, present = os.LookupEnv("TX_RPC_URL")
	if !present {
		config.TxSenderRpc = config.Rpc
	}
	config.Ws, present = os.LookupEnv("WS_URL")
	if !present {
		return errors.New("no ws url specified")
	}
	s := grpc.NewServer()

	err = server.Run(ctx, config, s)
	if err != nil {
		return err
	}
	go loopClose(ctx, l, s)
	return s.Serve(l)
}

// Do ./txproc server 15 tcp://:50051
// Or ./txproc client tcp://127.0.0.1:8899 tcp://localhost:50051
func main() {
	if len(os.Args) < 1 {
		panic(errors.New("no argument specified"))
	}

	signalC := make(chan os.Signal, 1)
	signal.Notify(signalC, syscall.SIGTERM, syscall.SIGINT)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go loopSignal(ctx, cancel, signalC)
	var err error
	args := os.Args[1:]
	switch args[0] {
	case CMD_CLIENT:
		err = run_client(ctx, args[1:])
	case CMD_SERVER:
		err = run_server(ctx, args[1:])
	default:
		err = fmt.Errorf("unknown command %s", args[0])
	}
	if err != nil {
		fmt.Printf("failed: %s", err)
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

func loopClose(ctx context.Context, l net.Listener, s *grpc.Server) {
	<-ctx.Done()
	s.GracefulStop()
	l.Close()
}

func loopSignal(ctx context.Context, cancel context.CancelFunc, signalC <-chan os.Signal) {
	defer cancel()
	doneC := ctx.Done()
	select {
	case <-doneC:
	case s := <-signalC:
		os.Stderr.WriteString(fmt.Sprintf("%s\n", s.String()))
	}
}
