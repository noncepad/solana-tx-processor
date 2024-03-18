package worker_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestWorker(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Create a new server instance
	server := &http.Server{
		Addr:    ":8080",
		Handler: http.HandlerFunc(handleRequest),
	}

	fmt.Println("Server listening on port 8080...")

	// Start the server
	go loopClose(ctx, server)
	go loopListen(server)

	err := send("http://localhost:8080", []byte("i am a transaction")) //cast string as byte array
	if err != nil {
		t.Fatal(err)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, client!")
	time.Sleep(10 * time.Second)
}

func loopClose(ctx context.Context, s *http.Server) {
	<-ctx.Done()
	s.Shutdown(context.Background())
}

func loopListen(s *http.Server) {
	if err := s.ListenAndServe(); err != nil {
		fmt.Println(err)
	}
}

type JsonRpcRequest struct {
	Jsonrpc string   `json name:"jsonrpc"`
	Id      int      `json name:"id"`
	Method  string   `json name:"method`
	Params  []string `json name:"params"`
}

func send(url string, data []byte) error {
	req := new(JsonRpcRequest)
	req.Jsonrpc = "2.0"
	req.Id = 1
	req.Method = "sendTransaction"
	req.Params = make([]string, 1)
	req.Params[0] = base64.StdEncoding.EncodeToString(data)
	jsonStr, err := json.Marshal(req)
	if err != nil {
		return err
	}
	req2, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return err
	}

	req2.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req2)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	// Read response body if needed
	// body, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println("response Body:", string(body))
	return nil
}
