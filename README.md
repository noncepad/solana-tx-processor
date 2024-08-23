# solana-tx-processor

This repository is for validators who want:

* [to sell transaction bandwidth](https://github.com/noncepad/solpipe-market/tree/main/txproc)
* [to run Solpipe State gRPC endpoint](https://github.com/noncepad/solpipe-market/tree/main/solpipe)

[See this page for State management](/docs/State.md)

## Transaction Processing

This is a:

1. gRPC <-- JSON RPC translating proxy to use in front of a validator and behind a pipeline reverse proxy
2. a JSON RPC -> gRPC translating proxy to use in front of a bidder proxy daemon.

## Install

```bash
go install github.com/noncepad/solana-tx-processor@main
```

## Run

### Validator Wrapper

Pipelines run this program in front of the validator.

```bash
RPC_URL=http://localhost:18899 WS_URL=ws://localhost:18900 solana-tx-processor server 15 tcp://:50051
```

* 15 is the worker count
* `tcp://:50051` is the address on which the grpc server will listen

### Json Rpc Endpoint

Bidders run this program in front of their proxy daemon.

```bash
solana-tx-processor client tcp://:8899 unix:///tmp/my-proxy-daemon.sock
```

* `tcp://:8899` is the listening address for the JSON RPC endpoint that only accepts `sendTransaction` calls.
* `unix:///tmp/my-proxy-daemon.sock` is the address that the proxy daemon is listening on
