## Tx-Processor Pipeline (Seller)

## Setup
This operation requires a combination of terminal commands and daemonized processes.
Here's how it works:
    
1. Initial Setup: You'll need to run specific commands in your terminal to configure the initial state of the system.

2. Daemonization: After the initial setup, the remaining operations will be handled by a daemonized process, running in the background.


## Pipeline Operator Requirements

* Solpipe and SafeJar CLI. Install [here](https://solpipe.io/docs/getting-started/linux/)
* Server


## Create Pipeline Locally

### Initialize pipeline
Initialize pipeline locally with the Solana tx processor marketplace id:
```bash
mkdir solpipe-txproc
cd solpipe-txproc
export TXPROC_MARKET_ID=HgsivZqrenp1835P4y8yLkF3dR2DKhN3AKiZ9sxCC5xH
solpipe pipeline init $xTXPROC_MARKET_ID . --create-jar --window=6h
```

> **create-jar flag**: Initialize your pipeline using the boolean ‘--create-jar’ flag. This process will create a SafeJar, jar delegation account (governed by SweepV2(Jar, Mint, 0) rule). 
>    For further details please visit [Safejar.io](https://safejar.io/docs/concepts/).

> **window flag**: The ‘window’ parameter represents a time interval , which is utilized to gauge usage capacity. Solpipe determines pipeline capacity based on the number of calls that can be accommodated within >this defined time window. 

The following files will initialize:

* The **authorizer.json** file, which contains the private key responsible for signing all transactions executed by the bot.

* The **bot.json** file, which details the parameters and conditions governing the pipeline bot and manages its on-chain existence.

* The **relay.json** file, which facilitates communication between your pipeline and those bidding.

* The **jar-owner.json** file holds the private key to your Jar Account.

* The **usage.lua** file serves as a metering tool for the marketplace, quantifying the usage of services for billing purposes. This file adheres to a specific format and should not be changed without consulting [Noncepad](https://docs.google.com/forms/d/1mcc3KsDuA-Lba30Q6mJ6T7aq8I2irrPboUWT9CoBse0/viewform?edit_requested=true).
 
### Customize Configs

The bot.json and relay.json files can be adjusted to suit your preferences before you create the pipeline instance.You can find more information on how to customize these files: [Understanding bot.json](https://solpipe.io/docs/pipeline/bot/), [Understanding relay.json](https://solpipe.io/docs/pipeline/relay/)

### Fund Authorizer

Next, fund the authorizer.json file with ~0.3 Sol for signing transaction fees.


### Create a Pipeline Instance on Solpipe
Once you are satisfied with you bot file create your pipeline instance with the following command (you may not adjust this file once the instance is created):
```bash
solpipe pipeline create ./bot.json --fee-payer=authorizer.json
```

## Daemonize Pipeline

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

Next, copy local pipeline configs to server.


 






### txproc.default
### txproc-forwarder.service
### txproc-bot.service
### txproc-relay.service





