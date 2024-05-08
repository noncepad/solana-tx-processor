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

	"github.com/noncepad/solana-tx-processor/server"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// ./txproc 15 tcp://:50051
func main() {
	var present bool
	var err error
	config := new(server.Configuration)
	if len(os.Args) < 3 {
		panic(errors.New("no worker count and listen url specified"))
	}

	config.WorkerCount, err = strconv.Atoi(os.Args[1])
	if err != nil {
		panic(err)
	}
	log.SetLevel(log.DebugLevel)
	var l net.Listener
	url := os.Args[2]
	if strings.HasPrefix(url, "unix://") {
		l, err = net.Listen("unix", strings.TrimPrefix(url, "unix://"))
	} else if strings.HasPrefix(url, "tcp://") {
		l, err = net.Listen("tcp", strings.TrimPrefix(url, "tcp://"))
	} else {
		err = fmt.Errorf("unknown url protocol %s", url)
	}
	if err != nil {
		panic(err)
	}
	config.Rpc, present = os.LookupEnv("RPC_URL")
	if !present {
		panic(errors.New("no rpc url specified"))
	}
	config.Ws, present = os.LookupEnv("WS_URL")
	if !present {
		panic(errors.New("no ws url specified"))
	}
	signalC := make(chan os.Signal, 1)
	signal.Notify(signalC, syscall.SIGTERM, syscall.SIGINT)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go loopSignal(ctx, cancel, signalC)
	s := grpc.NewServer()

	err = server.Run(ctx, config, s)
	if err != nil {
		panic(err)

	}
	go loopClose(ctx, l, s)
	err = s.Serve(l)
	if err != nil {
		panic(err)
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
