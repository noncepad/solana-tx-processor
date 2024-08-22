// Run a JSON RPC endpoint.  Forward sendTransaction calls to the TxProc gRPC connection.
package proxy

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"

	sgo "github.com/gagliardetto/solana-go"
	"github.com/noncepad/solana-tx-processor/util"
	pbt "github.com/noncepad/solpipe-market/go/proto/txproc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Configuration struct {
	ProxyUrl   string `json:"proxy"`
	ListenUrl  string `json:"listen"`
	proxyAddr  net.Addr
	listenAddr net.Addr
}

func (c *Configuration) Check() error {
	var err error
	c.proxyAddr, err = util.Parse(c.ProxyUrl)
	if err != nil {
		return err
	}
	c.listenAddr, err = util.Parse(c.ListenUrl)
	if err != nil {
		return err
	}
	return nil
}

func (c *Configuration) grpcurl() (string, error) {
	switch c.proxyAddr.Network() {
	case "unix":
		return fmt.Sprintf("%s://%s", c.proxyAddr.Network(), c.proxyAddr.String()), nil
	case "tcp":
		return c.proxyAddr.String(), nil
	default:
		return "", errors.New("unsupported protocol")
	}
}

type external struct {
	c            pbt.TransactionProcessingClient
	sharedBuffer grpc.SharedBufferPool
}

func Run(
	parentCtx context.Context,
	config *Configuration,
) error {
	err := config.Check()
	if err != nil {
		return err
	}
	var conn *grpc.ClientConn
	{
		url, err := config.grpcurl()
		if err != nil {
			return err
		}
		conn, err = grpc.Dial(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return err
		}
	}
	l, err := net.Listen(config.listenAddr.Network(), config.listenAddr.String())
	if err != nil {
		return err
	}
	e1 := external{
		c:            pbt.NewTransactionProcessingClient(conn),
		sharedBuffer: grpc.NewSharedBufferPool(),
	}

	server := &http.Server{
		Handler: e1,
	}
	go loopClose(parentCtx, server, l, conn)
	return server.Serve(l)
}

func loopClose(ctx context.Context, s *http.Server, l net.Listener, conn *grpc.ClientConn) {
	<-ctx.Done()
	conn.Close()
	s.Close()
	l.Close()
}

func (e1 external) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	jsonreq := new(RpcRequest)
	err := json.NewDecoder(r.Body).Decode(jsonreq)
	if err != nil {
		log.Printf("failed to parse: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var resp *RpcResponse
	switch jsonreq.Method {
	case "sendTransaction":
		resp, err = e1.do_tx(r.Context(), jsonreq.Id, jsonreq.Parameters)
	default:
		w.WriteHeader(http.StatusNotImplemented)
		return
	}
	if err != nil {
		log.Printf("error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Printf("error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// https://solana.com/docs/rpc/http/sendtransaction
func (e1 external) do_tx(ctx context.Context, reqId int, params []string) (*RpcResponse, error) {
	if len(params) < 1 {
		return nil, errors.New("blank transaction")
	}
	var simulate bool
	if 2 <= len(params) {
		switch params[1] {
		case "true":
			simulate = false
		case "false":
			simulate = true
		default:
			return nil, errors.New("unknown skipPreflight parameter")
		}
	} else {
		simulate = false
	}
	data, err := base64.StdEncoding.DecodeString(params[0])
	if err != nil {
		return nil, err
	}
	resp, err := e1.c.Broadcast(ctx, &pbt.BroadcastRequest{
		Transaction: data,
		Simulate:    simulate,
	})
	if err != nil {
		return nil, err
	}
	protosig := resp.GetSignature()
	if protosig == nil {
		return nil, errors.New("blank signature")
	}
	if len(protosig) != sgo.SignatureLength {
		return nil, fmt.Errorf("bad signature length: %d", len(protosig))
	}
	sig := sgo.SignatureFromBytes(protosig)
	ans := new(RpcResponse)
	ans.Id = reqId
	ans.JsonRpcVersion = "2.0"
	ans.Result = sig.String()
	return ans, nil
}

type RpcRequest struct {
	JsonRpcVersion string   `json:"jsonrpc"`
	Id             int      `json:"id"`
	Method         string   `json:"method"`
	Parameters     []string `json:"params"`
}

type RpcResponse struct {
	JsonRpcVersion string `json:"jsonrpc"`
	Id             int    `json:"id"`
	Result         string `json:"result"`
}
