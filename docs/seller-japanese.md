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


### Solpipe でパイプラインインスタンスを作成する
bot ファイルに満足したら、次のコマンドでパイプラインインスタンスを作成します (インスタンスを作成したら、このファイルは調整できません)。
```bash
solpipe pipeline create ./bot.json --fee-payer=authorizer.json
```

## パイプラインをデーモン化
次に、ローカルのパイプライン設定をサーバーにコピーします。

ユーザーを作成します:
```bash
sudo useradd -r solpipe
sudo mkdir -p /var/share/solpipe/txproc
sudo chown -R solpipe:solpipe /var/share/solpipe
```
実行ファイルのインストール:

```bash
go install github.com/noncepad/solana-tx-processor@main
sudo install -m 0755 $(which solana-tx-processor) /usr/local/bin
```

Solpipe、Safejar、Solana-tx-processor の実行ファイルが揃っているはずです。

### Create systemd files

パイプラインをデーモン化するには、次の systemd ファイルが必要です
    * txproc.default
    * txproc-forwarder.service
    * txproc-bot.service
    * txproc-relay.service

### txproc.default
default ファイルには、サービスファイルを実行するために必要なすべての環境変数が格納されています。

    実ファイルパス: プレースホルダーのファイルパスを、ファイルの実際の場所に置き換えてください。

    RPC および WebSocket URL: URL を、サーバーが各接続タイプに使用している特定のポート番号に置き換えてください。

次のようになります。
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
このファイルは、トランザクション処理サービスを構成します。
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
このファイルは、ボットサービスを構成します。\
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
このファイルは、リレーサービスを構成します。 
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
### パイプラインプロセスの開始 
サービスファイルを有効にする:```bash
sudo systemctl enable txproc-bot.service
sudo systemctl enable txproc-relay.service
```

サービスを開始する:
```bash
sudo systemctl start txproc-bot.service
sudo systemctl start txproc-relay.service
```
サービスが実行されていることを確認する:
```bash
sudo systemctl status txproc-bot.service
sudo systemctl status txproc-relay.service
```
これで、パイプラインサーバーは起動時に自動的に起動し、バックグラウンドで実行し続けます。必要に応じて、systemctl コマンドを使用してサービスを管理できます (開始、停止、再起動)。



