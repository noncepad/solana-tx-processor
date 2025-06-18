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

	"git.noncepad.com/pkg/go-solpipe/p2p"
	sgo "git.noncepad.com/pkg/solana-go"
	"github.com/noncepad/solana-tx-processor/server"
	pbt "github.com/noncepad/solpipe-market/go/proto/txproc"
	log "github.com/sirupsen/logrus"
	gproxy "golang.org/x/net/proxy"
	"google.golang.org/grpc"
)

const (
	CMD_SERVER = "server"
)

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
	addr, err := p2p.ParseAddress(os.Getenv("TXPROC_URL"))
	if err != nil {
		return err
	}
	if addr.Network() != "solpipe" {
		return errors.New("must use solpipe protocol")
	}
	destination, err := sgo.PublicKeyFromBase58(addr.String())
	if err != nil {
		return err
	}
	fakeAdmin, err := sgo.NewRandomPrivateKey()
	if err != nil {
		return err
	}
	torD, err := localTorDialer()
	if err != nil {
		return err
	}
	conn, err := p2p.ConnectToClearNetOrFallbackToTor(ctx, destination, fakeAdmin, nil, torD, []grpc.DialOption{})
	if err != nil {
		return err
	}
	txprocClient := pbt.NewTransactionProcessingClient(conn)
	config.TxSenderRpc, present = os.LookupEnv("TX_RPC_URL")
	if !present {
		return errors.New("missing TX_RPC_URL")
	}
	s := grpc.NewServer()

	err = server.Run(ctx, config, s, txprocClient)
	if err != nil {
		return err
	}
	go loopClose(ctx, l, s)
	return s.Serve(l)
}

func localTorDialer() (gproxy.ContextDialer, error) {
	dialer, err := gproxy.SOCKS5("tcp", "localhost:9050", nil, gproxy.Direct)
	if err != nil {
		return nil, err
	}
	return &torDialer{dialer: dialer}, nil
}

type torDialer struct {
	dialer gproxy.Dialer
}

func (t *torDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return t.dialer.Dial(network, address)
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
