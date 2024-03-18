package worker_test

import (
	"bytes"
	"context"
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

func send(url string, data []byte) error {
	jsonStr := []byte(`{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "sendTransaction",
    "params": [
      "4hXTCkRzt9WyecNzV1XPgCDfGAZzQKNxLXgynz5QDuWWPSAZBZSHptvWRL3BjCvzUXRdKvHL2b7yGrRQcWyaqsaBCncVG7BFggS8w9snUts67BSh3EqKpXLUm5UMHfD7ZBe9GhARjbNQMLJ1QD3Spr6oMTBU6EhdB4RD8CP2xUxr2u3d6fos36PD98XS6oX8TQjLpsMwncs5DAMiD4nNnR8NBfyghGCWvCVifVwvA8B8TJxE1aiyiv2L429BCWfyzAme5sZW8rDb14NeCQHhZbtNqfXhcp2tAnaAT"
    ]
  }`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
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
