# solana-tx-processor

This is a gRPC <--> JSON RPC translating proxy to use in front of a validator.

## Run

```bash
go build -o ./txproc github.com/noncepad/solana-tx-processor
RPC_URL=http://localhost:18899 WS_URL=ws://localhost:18900 ./txproc 15 tcp://:50051
```

