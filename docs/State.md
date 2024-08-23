# Validator + Solpipe State Server

The validator refers to the Solana Rust default validator.  [Solpipe](https://solpipe.io/docs/) and [Safejar](https://safejar.io/docs/) CLI tools require access to a state management server.  The *state* server requires access to a *tape* server.  The *tape* server deduplicates chain state updates from the validator received via [the Yellowstone Geyser plugin](https://github.com/rpcpool/yellowstone-grpc).

![State server management](./docs/img/State.png)

## Validator Requirements

The *tape* server requires access to a Yellowstone geyser gRPC endpoint running on the *validator*.  This documentation assumes the endpoint is on port 50051.

The geyser plugin is written in Rust and must be compiled to the same version as the validator that is running.

[Noncepad provides a compiled Solana and Yellowstone grpc Debian package](https://solpipe.io/docs/getting-started/linux/) for convenience for those running on Debian Bookworm.

Install the geyser plugin using the Noncepad repository:

```bash
sudo apt-get install -y solana* yellowstone-grpc-geyser
```

* [our Yellowstone grpc code](https://github.com/noncepad/yellowstone-grpc) is forked from [RPC Pool (Triton's code)](https://github.com/rpcpool/yellowstone-grpc)

Create a geyser plugin configuration file at `/etc/solana/geyser-testnet.json` to look like:

```json
{
  "libpath": "/usr/lib/libyellowstone-grpc-geyser.so",
  "log": {
    "level": "info"
  },
  "grpc": {
    "address": "0.0.0.0:50051",
    "snapshot_plugin_channel_capacity": "1_000_000",
    "snapshot_client_channel_capacity": "50_000_000",
    "channel_capacity": "100_000",
    "unary_concurrency_limit": 100,
    "unary_disabled": false
  },
  "block_fail_action": "log"
}
```

Add this flag to the command that runs the validator:

```bash
--geyser-plugin-config /etc/solana/geyser-testnet.json
```

## Tape Server Requirements

| *Requirement* | *Minimum* |
| ---- | ---- |
| CPU | 4 core, >2.8 GHz |
| RAM | 75 GB |
| Disk | 200 GB |
| Listening Ports | `tcp://[::]:30040, tcp://[::]:30041` |

```bash
sudo apt-get install -y solpipe-state 
sudo mkdir -p /var/share/solpipe/testnet/tape
sudo chown -R solpipe:solpipe /var/share/solpipe
```

Set `/etc/default/tape-testnet` to be:

```bash
NETWORK=TESTNET
WORKING_DIR=/var/share/solpipe/testnet/tape
LISTEN_STREAM_URL=tcp://:30040
LISTEN_SNAPSHOT_URL=tcp://:30041
BUFFER_SIZE=100000
```

Set `/etc/systemd/system/tape-testnet.service` to be:

```systemd

[Unit]
Description=Solpipe Tape on Testnet
After=syslog.target network.target remote-fs.target nss-lookup.target
Wants=systuner.service
StartLimitIntervalSec=0

[Service]
Type=simple
ExecStartPre=/bin/sh -c 'rm /run/solpipe-testnet/pusher.sock || true'
ExecStart=/usr/bin/solpipe-state tape --work=${WORKING_DIR} --buffer=${BUFFER_SIZE} ${LISTEN_STREAM_URL} ${LISTEN_SNAPSHOT_URL} unix:///run/solpipe-testnet/pusher.sock 
RuntimeDirectory=solpipe
RuntimeDirectoryMode=0750
User=solpipe
Group=solpipe
LimitNOFILE=1000000
RestartSec=120
Restart=never
#Restart=on-failure


[Install]
WantedBy=multi-user.target
```

### Ingest from Validator

The tape server runs independently of the validator. Once the tape server is on, it has to be attached to the validator Geyser endpoint to start reading block updates.

The following events need to take place for ingesting to be successful:

1. `systemctl start solana-testnet` - start the validator
1. do `solana monitor` to see when the validator has successfully downloaded a snapshot.
1. Run the attach command below to have the tape server start downloading the snapshot over gRPC in order to get the entire Solana account state.
1. The tape server churns through the account state (72 GB for mainnet)
   * **WARNING** - the validator will not complete the boot process until the entire snapshot is processed over gRPC
1. The state server can now connect and parsing account state.

```bash
systemctl start solana-testnet
sleep 30
VALIDATOR_HOST=192.168.10.1 NETWORK=TESTNET sudo -u solpipe /usr/bin/solpipe-state attach geyser --buffer-size=100000000 unix:///run/solpipe-testnet/pusher.sock tcp://${VALIDATOR_HOST}:50051
```

* replace VALIDATOR_HOST with the ipv4 address of the validator
* the sleep time may be insufficient. The attach command will only be successful after the validator has downloaded a snapshot from a peer

This command exists once the gRPC connection has been made by the tape server to the validator Geyser endpoint.

## State Server

| *Requirement* | *Minimum* |
| ---- | ---- |
| CPU | 4 core, >2.8 GHz |
| RAM | 75 GB |
| Disk | 200 GB |
| Listening Ports | `tcp://[::]:30004` |

Install a state server systemd service at `/etc/systemd/system/state-testnet.service`:

```systemd
[Unit]
Description=Solpipe State Indexer
After=syslog.target network.target remote-fs.target nss-lookup.target
Wants=systuner.service
Requires=tape-testnet.service
StartLimitIntervalSec=0

[Service]
Type=simple
ExecStart=/tmp/bin/solpipe-indexer run -v 
RuntimeDirectory=solpipe-indexer
RuntimeDirectoryMode=0750
EnvironmentFile=/etc/default/state-localnet
User=solpiper
Group=solpiper
LimitNOFILE=1000000
RestartSec=120
Restart=never


[Install]
WantedBy=multi-user.target
```

Set default variables at `/etc/default/state-testnet`:

```bash
GOMAXPROCS=4
NETWORK=TESTNET
TAPE_URL=tcp://localhost:30040
SNAPSHOT_URL=tcp://localhost:30041
LISTEN_URL=tcp://:30104
WORKING_DIR=/var/share/solpiple/testnet/state
BUFFER_SIZE=100000
IGNORE_EMPTY_SNAPSHOT="1"
```
