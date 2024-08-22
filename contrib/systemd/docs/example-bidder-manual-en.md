# Example Solana-Txproc: Solpipe Bidder Manual

The Solana Transaction Processor, or Tx-Proc, is a server designed to write to the Solana Blockchain network.

**pipeline id**: <6npGpXLNMCDt3uYuFonZZ6Zhr5jZM6S4o9qj3BdqT6qy>
**Validator stake**: <2 M>
**Period length**: 6h
**Average Payout per Bidder**: $1/period

## Pricing

The traditional route for monetizing API based bandwidth can be very complex. Solpipe’s approach of throttling offers a more transparent and cost-effective solution. Instead, pricing adjusts dynamically in response to supply/demand and available resources.

Solpipe replaces accounting, client authentication, and ingress networking.

Solpipe automates pricing based on server capacity, saving time and effort in determining pricing structures. Instead of directly setting prices, establish constraints on your capacity. Instead of paying per requests, set constraints on price and pay the “equilibrium price” dictated by demand and available resources.

## Run a Solpipe Bidder 

### Install 
Install Solpipe and Safejar CLI: https://solpipe.io/docs/getting-started/linux/

### Compiling to Your Server
 To initiate a bidder in a Solpipe marketplace, integrate the marketplace libraries into your code.
* download marketplace
* integrate

### Use the SafeJar program to operate automatic spending/bidding

With SafeJar, your funds are protected by customizable rule sets and constraints, mitigating risks associated with direct crypto transfers. This makes automatic spending by bots possible. Rules embedded directly to the Blockchain allows for secure, transparent transfers.

https://safejar.io
Solpipe CLI:

Also see Solpipe User Docs: https://solpipe.io/docs/bidder/bidder/

### Create a Bidder

```bash
mkdir solpipe-bidder
cd solpipe-bidder
```

#### Generate an Authorizer key

As the engineer tasked with configuring the bidder bot and conducting bids for software needs, you are required to create an authorizer public-private key pair. The following commands will help you generate these keys:
```bash
solana-keygen new -o ./authorizer.json
solana-keygen pubkey ./authorizer.json
```

#### Create a Rule-Set

* Create a rule set orrr (parse command)
Or generate a Solpipe compatible rule-set for the initialization of the 'Delegation Account'.
The Safejar 'Delegation Account' will be used create the bidding account and governs spending with a set of rules/constraints.
Generate the 'rule.set' file with the following command:
 ```bash
solpipe bidder rule-set <marketplace-id> <jar_id> <authorizer> <max-spend> <slot> <minBal>
``` 
----------------------
marketplace-id
: Specify the marketplace ID of the desired marketplace.

jar_id
: Specify the Jar ID created with [SafeJar](https://safejar.io). (To be created prior by the company's CFO/financial department)

authorizer
: Specify the authorizer.json file (private key) generated to to sign transactions.

max-spend
: Specify the max tokens that can be used by the delegation account per given slot (for the rate limiter rule).

slot
: Specify the duration of slots in which the delegation can spend the max-spend (for the rate limiter rule).

minBal
: Specify the minimum balance that remains in a bidder vault with a sweep (for the sweep rule).

--------------

Provide your company's CFO/financial department with this generated **rule.set file** to establish a **delegation account** using [SafeJar](https://safejar.io). This account will be used as the 'owner' of the bidder account.


there is more


## Start your Proxy Daemon

The bidder proxy is a daemonized server with a single grpc endpoint. The proxy will facilitate the bidding to the Solana Txproc Marketplace.

To start you bidder proxy execute the following:
```bash
 solpipe bidder proxy --listen=tcp://localhost:34043
 ```

**Record and save the generated bidder id.** 


## Add a Marketplace

Tell your proxy, which acts as an intermediary, to start connecting to the pipelines within the Solana Txproc Marketplace.

Execute the following:
```bash
solpipe bidder market <marketplace-id> <jar-id generated with Safejar> <file path to your rule.set file>
```
### Allocation
## Other

## Check you bidder status

Execute the following command to get a list of marketplaces connected to your proxy:
```bash
solpipe bidder status 
```

## View Your Bidder State

To view your bidder agent data with your bidder/agent id:

```bash
solpipe view agent <agent-id> [flags]
```
<details>
  <summary style="font-size: 20px; padding: 10px 20px; background-color: #d9ff00; border-radius: 5px; color: #280096; margin-bottom: 20px; width: max-content" >Help: Arguments and Flags</summary><p>
  
----------------
agent-id
: Specify your bidder ID
----------------
**[flags]**
show-bids
: To see list of all bids made by bidder include this boolean flag
-----------------

  </p>
</details>



### Other 
* check status
* check state








## Create a Blank Directory





