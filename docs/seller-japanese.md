## Tx-Processor パイプライン (販売者)

## セットアップ
この操作には、ターミナルコマンドとデーモン化されたプロセスの組み合わせが必要です。
仕組みは以下のとおりです。
1. 初期設定: システムの初期状態を設定するには、ターミナルで特定のコマンドを実行する必要があります。
2. デーモン化: 初期設定後、残りの操作はバックグラウンドで実行されるデーモン化されたプロセスによって処理されます。


### パイプラインオペレーターの要件

* Solpipe と SafeJar CLI。こちらからインストールしてください。[こちら](https://solpipe.io/docs/getting-started/linux/)からインストールしてください。
* サーバー


### パイプラインの初期化
Solana トランザクションプロセッサのマーケットプレイスIDを使って、**ローカル**でパイプラインを初期化します:
```bash
mkdir solpipe-txproc
cd solpipe-txproc
export TXPROC_MARKET_ID=HgsivZqrenp1835P4y8yLkF3dR2DKhN3AKiZ9sxCC5xH
solpipe pipeline init $xTXPROC_MARKET_ID . --create-jar --window=6h
```

> **create-jar フラグ**: --create-jar フラグを使ってパイプラインを初期化します。このプロセスは SafeJar、jar デリゲーションアカウント (SweepV2(Jar, Mint, 0) ルールで管理) を作成します。
> 詳細については、[Safejar.io](https://safejar.io/docs/concepts/)Safejar.io をご覧ください。 .

> **window  フラグ**: window パラメーターは時間間隔を表し、使用量容量を測定するために使用されます。Solpipe は、この定義された時間間隔内に収容できる呼び出しの数に基づいて、パイプラインの容量を決定します。

次のファイルが初期化されます。

* **authorizer.json** ファイル: ボットによって実行されるすべてのトランザクションに署名する責任のある秘密鍵が含まれています。

* **bot.json** ファイル: パイプラインボットを管理するパラメーターと条件を詳細に示し、オンチェーンの存在を管理します。

* **relay.json** ファイル: パイプラインと入札者の間で通信を円滑化します。
* **jar-owner.json** ファイル: Jar アカウントの秘密鍵を保持します。
* **usage.lua**ファイル: マーケットプレイスの使用量測定ツールとして機能し、課金目的でサービスの使用量を定量化します。このファイルは特定のフォーマットに従い、[Noncepad](https://docs.google.com/forms/d/1mcc3KsDuA-Lba30Q6mJ6T7aq8I2irrPboUWT9CoBse0/viewform?edit_requested=true)に相談せずに変更しないでください。

### 設定のカスタマイズ

パイプラインインスタンスを作成する前に、bot.json ファイルと relay.json ファイルを好みに合わせて調整できます。これらのファイルのカスタマイズ方法の詳細については、[bot.json の理解](https://solpipe.io/docs/pipeline/bot/)、[relay.json の理解](https://solpipe.io/docs/pipeline/relay/) を参照してください。

### オーソライザーの資金供給

次に、トランザクション手数料の署名のために、authorizer.json ファイルに約 0.3 Sol を入金します。


### Create a Pipeline Instance on Solpipe
Once you are satisfied with you bot file create your pipeline instance with the following command (you may not adjust this file once the instance is created):
```bash
solpipe pipeline create ./bot.json --fee-payer=authorizer.json
```

## Daemonize Pipeline
Next, copy local pipeline configs to your server.

Create your user:
```bash
sudo useradd -r solpipe
sudo mkdir -p /var/share/solpipe/txproc
sudo chown -R solpipe:solpipe /var/share/solpipe
```
Install executables:
```bash
go install github.com/noncepad/solana-tx-processor@main
sudo install -m 0755 $(which solana-tx-processor) /usr/local/bin
```

You should now have solpipe, safejar, and solana-tx-processor executables.

### Create systemd files

To daemonize your pipeline you will need the following systemd files:
    * txproc.default
    * txproc-forwarder.service
    * txproc-bot.service
    * txproc-relay.service

### txproc.default
The default file holds all environmental variables needed to run service files.

* Real file paths: Replace any placeholder file paths with the actual locations of your files.

* RPC and WebSocket URLs: Replace URLs with the specific port numbers your server is using for each connection type.

It should look something like this: 
```bash
# for bot and relay
NETWORK=MAINNET
AUTHORIZER=/etc/solpipe/txproc/authorizer.json
# for relay
RELAY=/etc/solpipe/txproc/relay.json
# for bot
BOT=/etc/solpipe/txproc/bot.json

# for the txproc server itself
WORKER_COUNT=15
RPC_URL=http://localhost:8899
WS_URL=ws://localhost:8900
```
### txproc-forwarder.service
This file will configure the transaction processing service.
```bash
[Unit]
Description=Solpipe Txproc Forwarder
After=network.target

[Service]
Type=simple
User=solpipe
Group=solpipe
WorkingDirectory=/var/share/solpipe/txproc
EnvironmentFile=/etc/default/txproc-relay
ExecStart=/usr/local/bin/solana-tx-processor server $WORKER_COUNT unix:///run/txproc/server.sock 
Restart=always
RestartSec=60
RuntimeDirectory=txproc
RuntimeDirectoryMode=0755

[Install]
WantedBy=multi-user.target #Target the service belongs to (e.g., multi-user.target).
```
### txproc-bot.service
This file will configure the bot service.
```bash
[Unit]
Description=Solpipe Txproc Bot Relay
After=network.target

[Service]
Type=simple
User=solpipe
Group=solpipe
WorkingDirectory=/var/share/solpipe/txproc
EnvironmentFile=/etc/default/txproc
ExecStart=/usr/bin/solpipe pipeline --fee-payer=$AUTHORIZER $RELAY  -v 
Restart=never

[Install]
WantedBy=multi-user.target #Target the service belongs to (e.g., multi-user.target).
```
### txproc-relay.service
This file will configure the relay service.
```bash
[Unit]
Description=Solpipe Txproc Relay
After=network.target
Requires=txproc-forwarder.service

[Service]
Type=simple
User=solpipe
Group=solpipe
WorkingDirectory=/var/share/solpipe/txproc
EnvironmentFile=/etc/default/txproc
ExecStart=/usr/bin/solpipe pipeline relay $RELAY $AUTHORIZER -v 
Restart=never

[Install]
WantedBy=multi-user.target #Target the service belongs to (e.g., multi-user.target).
```
### Start pipeline processes

Enable the service files:
```bash
sudo systemctl enable txproc-bot.service
sudo systemctl enable txproc-relay.service
```

Start the services:
```bash
sudo systemctl start txproc-bot.service
sudo systemctl start txproc-relay.service
```
Verify that your services are running:
```bash
sudo systemctl status txproc-bot.service
sudo systemctl status txproc-relay.service
```

Now, your pipeline server will start automatically on boot and continue running in the background. You can manage your services (start, stop, restart) using systemctl commands as needed.



