# End-to-Batchrec running multiprocess

Run DPSS with real nodes. Current (hacky) working
- old committee of 3 nodes
- trigger DPSS:
  - 1 share is added to the db of each old node
  - lastAssignedIndex for secp256k1 is hardcoded 1
  - at the process of each old node, a new node is started

Follow the steps very carefully to trigger this exact behaviour (especially with changing of configs before DPSS trigger). 

Current issue is something with the p2p connections. More explanation at the end. 

Note: chainId was added to config. Not sure if it's needed but I thought it could give issues with Tendermint otherwise. When everything works we can check.

## Test steps

### Setup

To be safe, remove `/tmp/keygen-datax` for x=1,..,6. 

```
# create correct bins
go build -o adkgNode main.go
go build -o nodeManager ./manager_process

# set pss flag to false
python3 nodelist-pyscript/node_list.py -pc 1 2 0
```

Set up 3 different configs in: 
- `new-node-config/config.test.1.json`
- `new-node-config2/config.test.2.json`
- `new-node-config3/config.test.3.json`

Note: Add "chainId": "test-net-1".

### Run initial nodes

Then run in 3 different terminals:
```
./nodeManager
./nodeManager -config ./manager_process/new-node-config2
./nodeManager -config ./manager_process/new-node-config3
```
Let the nodes connect.

### Add new configs & Trigger PSS flag

Before the next step, CHANGE THE CONFIGS. 
The same files, but different content. Make sure to add "chainId": "test-net-2". 

Then, change PSS flag:
```
python nodelist-pyscript/node_list.py -pc 1 2 1
```

## Info: Temp Hacks to make DPSS run

This contains explanation, no further instructions.

To kick off the PSS process, the old node should find some assigned shares in the database. This is hacked in for the moment:
1. when PSS is triggered 1 share is added to the db with index 0
2. secpShareNum is set to 1. (this can also be 0, but with 1 we can see in the logging 1 of the shares was not retrieved)

We'll see this type of logging:
```
13530 time="2024-04-26T16:05:44-06:00" level=error msg="unable to retrieve secp256k1 share of index 1: Share not found!"
13530 time="2024-04-26T16:05:44-06:00" level=info msg="Running DPSS" batch=0 type=secp256k1
```

Ideally we want to be able to nicely add shares (only in a test situation) + set the last assigned index somehow.

## Current Status

### 0429 - Alex
* The "failed to dial" problem below can be fixed by changing ip address in config to 127.0.0.1
* Encounter another problem when Sending pss p2p message: `level=error msg="failed to create stream" error="failed to negotiate protocol: protocols not supported: [dpss-1/]"` that is caused by #L37 in dpss/pss_node.go. Since we are creating P2P protocol with the node's epoch (`getPSSProtocolPrefix(epoch)`), old nodes and new nodes will have different protocol (`dpss-1/` and `dpss-2/`). For nodes to communicate through P2P, they need to shared the same protocol. Currently all nodes' protocol is set to "dpss-1/" to fixed this issue. It seems fine since the protocol shoudn't change between epochs.
* Adds "time.Sleep(10 * time.Second)" in dpss/pss_service.go #L178 to allow other honest nodes to finish creating PSSNode. Without this line, we might also get `level=error msg="failed to create stream" error="failed to negotiate protocol: protocols not supported: [dpss-1/]"`. This is caused by trying to send P2P message to other nodes while they are still setting up their PSSNodes. (Actually all the nodes seems to received the correct msg eventually dispite some failed attempts at first, because retry.Do{} is used when sending p2p message.) 
* The reasoning of "time.Sleep(10 * time.Second)" is that nodes only query pssStatus every 10 secs, so we should make sure all honest nodes have seen the pss status change before starting the PSS process.


## Problem

This works:
1. Old node starts
2. DPSS is triggered
3. For each old node: 1 test share is added to the db with index 0
4. For each old node PROCESS: a new node is created.
5. Old node sends InitMsg for the single share

Expected behaviour:
1. InitMsg is received by self
2. Then NewDualCommitteeACSSShareMessage is received by self
3. Then NewAcssProposeMessageround is broadcast to old & new committee
4. ... and then more but we don't get there yet

The problem is at the Broadcast at step 3. Why?
- DPSS batch is triggered:
`17240 time="2024-04-26T19:03:05-06:00" level=info msg="Running DPSS" batch=0 type=secp256k1`
- The logging shows old node self receives an `Echo` message. Which means it receives all previous messages as well. 
```
17240 time="2024-04-26T19:03:05-06:00" level=info msg="InitMessage: Process"
17240 time="2024-04-26T19:03:05-06:00" level=info msg="Echo received: Sender=3, Receiver=3"
17240 time="2024-04-26T19:03:05-06:00" level=error msg="DACSSEchoMessage: Process" Message="Ignore the echo message from self." SelfIdx=3 SenderIdx=3
17240 time="2024-04-26T19:03:05-06:00" level=info msg="Echo received: Sender=3, Receiver=3"
17240 time="2024-04-26T19:03:05-06:00" level=error msg="DACSSEchoMessage: Process" Message="Ignore the echo message from self." SelfIdx=3 SenderIdx=3
```
- The logging shows error of not actually being able to connect to the other nodes (of ANY committee). Errors printed below in next section.

### Where to look

The problem is NOT in DPSS specifically. 
In `epochNodesMonitor` in `chain_service` I added a send p2p message after the should have been "connected" to a peer. But this results in the same error. 

Solving this should tell us how to solve the broadcast issue. (Or immediately solve it)

Below are the errors are shown when broadcast fails. Use `export IPFS_LOGGING=DEBUG` to turn on debugging for the p2plib. 

```
17232 2024-04-26T19:03:08.766-0600      DEBUG   swarm2  swarm/limiter.go:201    [limiter] clearing all peer dials: 16Uiu2HAm87WC1Y1gNx8byRUMxdowRSEmH1mXEmcDmszC1QU5BbvR
17232 2024-04-26T19:03:08.768-0600      DEBUG   basichost       basic/basic_host.go:737 host 16Uiu2HAm2W4Ad8JCVxhsaDvpfw1yEbY6xE8PhaTkbZeXhWJsnzw5 dialing 16Uiu2HAmSWiDB6K42p6tTwNih5yDBbv2RoqdTjGJSP7iSpHDPABC
17232 2024-04-26T19:03:08.768-0600      DEBUG   swarm2  swarm/swarm_dial.go:239 dialing peer   {"from": "16Uiu2HAm2W4Ad8JCVxhsaDvpfw1yEbY6xE8PhaTkbZeXhWJsnzw5", "to": "16Uiu2HAmSWiDB6K42p6tTwNih5yDBbv2RoqdTjGJSP7iSpHDPABC"}
17232 2024-04-26T19:03:08.768-0600      DEBUG   swarm2  swarm/swarm_dial.go:277 network for 16Uiu2HAm2W4Ad8JCVxhsaDvpfw1yEbY6xE8PhaTkbZeXhWJsnzw5 finished dialing 16Uiu2HAmSWiDB6K42p6tTwNih5yDBbv2RoqdTjGJSP7iSpHDPABC
17232 time="2024-04-26T19:03:08-06:00" level=error msg="failed to create stream" error="failed to dial: failed to dial 16Uiu2HAmSWiDB6K42p6tTwNih5yDBbv2RoqdTjGJSP7iSpHDPABC: all dials failed\n  * [/ip4/192.167.10.13/tcp/1082] dial backoff"
17232 2024-04-26T19:03:08.768-0600      DEBUG   swarm2  swarm/limiter.go:201    [limiter] clearing all peer dials: 16Uiu2HAmSWiDB6K42p6tTwNih5yDBbv2RoqdTjGJSP7iSpHDPABC
17232 2024-04-26T19:03:08.777-0600      DEBUG   basichost       basic/basic_host.go:737 host 16Uiu2HAm2W4Ad8JCVxhsaDvpfw1yEbY6xE8PhaTkbZeXhWJsnzw5 dialing 16Uiu2HAm87WC1Y1gNx8byRUMxdowRSEmH1mXEmcDmszC1QU5BbvR
17232 2024-04-26T19:03:08.777-0600      DEBUG   swarm2  swarm/swarm_dial.go:239 dialing peer   {"from": "16Uiu2HAm2W4Ad8JCVxhsaDvpfw1yEbY6xE8PhaTkbZeXhWJsnzw5", "to": "16Uiu2HAm87WC1Y1gNx8byRUMxdowRSEmH1mXEmcDmszC1QU5BbvR"}
17232 2024-04-26T19:03:08.777-0600      DEBUG   swarm2  swarm/swarm_dial.go:277 network for 16Uiu2HAm2W4Ad8JCVxhsaDvpfw1yEbY6xE8PhaTkbZeXhWJsnzw5 finished dialing 16Uiu2HAm87WC1Y1gNx8byRUMxdowRSEmH1mXEmcDmszC1QU5BbvR
17232 time="2024-04-26T19:03:08-06:00" level=error msg="failed to create stream" error="failed to dial: failed to dial 16Uiu2HAm87WC1Y1gNx8byRUMxdowRSEmH1mXEmcDmszC1QU5BbvR: all dials failed\n  * [/ip4/192.167.10.12/tcp/1081] dial backoff"
17232 2024-04-26T19:03:08.777-0600      DEBUG   swarm2  swarm/limiter.go:201    [limiter] clearing all peer dials: 16Uiu2HAm87WC1Y1gNx8byRUMxdowRSEmH1mXEmcDmszC1QU5BbvR
17232 2024-04-26T19:03:08.779-0600      DEBUG   basichost       basic/basic_host.go:737 host 16Uiu2HAm2W4Ad8JCVxhsaDvpfw1yEbY6xE8PhaTkbZeXhWJsnzw5 dialing 16Uiu2HAmSWiDB6K42p6tTwNih5yDBbv2RoqdTjGJSP7iSpHDPABC
17232 2024-04-26T19:03:08.779-0600      DEBUG   swarm2  swarm/swarm_dial.go:239 dialing peer   {"from": "16Uiu2HAm2W4Ad8JCVxhsaDvpfw1yEbY6xE8PhaTkbZeXhWJsnzw5", "to": "16Uiu2HAmSWiDB6K42p6tTwNih5yDBbv2RoqdTjGJSP7iSpHDPABC"}
17232 2024-04-26T19:03:08.780-0600      DEBUG   swarm2  swarm/swarm_dial.go:277 network for 16Uiu2HAm2W4Ad8JCVxhsaDvpfw1yEbY6xE8PhaTkbZeXhWJsnzw5 finished dialing 16Uiu2HAmSWiDB6K42p6tTwNih5yDBbv2RoqdTjGJSP7iSpHDPABC
17232 time="2024-04-26T19:03:08-06:00" level=error msg="failed to create stream" error="failed to dial: failed to dial 16Uiu2HAmSWiDB6K42p6tTwNih5yDBbv2RoqdTjGJSP7iSpHDPABC: all dials failed\n  * [/ip4/192.167.10.13/tcp/1082] dial backoff"
17232 2024-04-26T19:03:08.780-0600      DEBUG   swarm2  swarm/limiter.go:201    [limiter] clearing all peer dials: 16Uiu2HAmSWiDB6K42p6tTwNih5yDBbv2RoqdTjGJSP7iSpHDPABC
17232 2024-04-26T19:03:08.785-0600      DEBUG   basichost       basic/basic_host.go:737 host 16Uiu2HAm2W4Ad8JCVxhsaDvpfw1yEbY6xE8PhaTkbZeXhWJsnzw5 dialing 16Uiu2HAmSWiDB6K42p6tTwNih5yDBbv2RoqdTjGJSP7iSpHDPABC
17232 2024-04-26T19:03:08.785-0600      DEBUG   swarm2  swarm/swarm_dial.go:239 dialing peer   {"from": "16Uiu2HAm2W4Ad8JCVxhsaDvpfw1yEbY6xE8PhaTkbZeXhWJsnzw5", "to": "16Uiu2HAmSWiDB6K42p6tTwNih5yDBbv2RoqdTjGJSP7iSpHDPABC"}
17232 2024-04-26T19:03:08.786-0600      DEBUG   swarm2  swarm/swarm_dial.go:277 network for 16Uiu2HAm2W4Ad8JCVxhsaDvpfw1yEbY6xE8PhaTkbZeXhWJsnzw5 finished dialing 16Uiu2HAmSWiDB6K42p6tTwNih5yDBbv2RoqdTjGJSP7iSpHDPABC
17232 2024-04-26T19:03:08.786-0600      DEBUG   swarm2  swarm/limiter.go:201    [limiter] clearing all peer dials: 16Uiu2HAmSWiDB6K42p6tTwNih5yDBbv2RoqdTjGJSP7iSpHDPABC
17232 time="2024-04-26T19:03:08-06:00" level=error msg="failed to create stream" error="failed to dial: failed to dial 16Uiu2HAmSWiDB6K42p6tTwNih5yDBbv2RoqdTjGJSP7iSpHDPABC: all dials failed\n  * [/ip4/192.167.10.13/tcp/1082] dial backoff"
17232 2024-04-26T19:03:08.794-0600      DEBUG   basichost       basic/basic_host.go:737 host 16Uiu2HAm2W4Ad8JCVxhsaDvpfw1yEbY6xE8PhaTkbZeXhWJsnzw5 dialing 16Uiu2HAm87WC1Y1gNx8byRUMxdowRSEmH1mXEmcDmszC1QU5BbvR
17232 2024-04-26T19:03:08.794-0600      DEBUG   swarm2  swarm/swarm_dial.go:239 dialing peer   {"from": "16Uiu2HAm2W4Ad8JCVxhsaDvpfw1yEbY6xE8PhaTkbZeXhWJsnzw5", "to": "16Uiu2HAm87WC1Y1gNx8byRUMxdowRSEmH1mXEmcDmszC1QU5BbvR"}
17232 2024-04-26T19:03:08.794-0600      DEBUG   swarm2  swarm/swarm_dial.go:277 network for 16Uiu2HAm2W4Ad8JCVxhsaDvpfw1yEbY6xE8PhaTkbZeXhWJsnzw5 finished dialing 16Uiu2HAm87WC1Y1gNx8byRUMxdowRSEmH1mXEmcDmszC1QU5BbvR
17232 2024-04-26T19:03:08.794-0600      DEBUG   swarm2  swarm/limiter.go:201    [limiter] clearing all peer dials: 16Uiu2HAm87WC1Y1gNx8byRUMxdowRSEmH1mXEmcDmszC1QU5BbvR
17232 time="2024-04-26T19:03:08-06:00" level=error msg="failed to create stream" error="failed to dial: failed to dial 16Uiu2HAm87WC1Y1gNx8byRUMxdowRSEmH1mXEmcDmszC1QU5BbvR: all dials failed\n  * [/ip4/192.167.10.12/tcp/1081] dial backoff"
17232 2024-04-26T19:03:08.795-0600      DEBUG   basichost       basic/basic_host.go:737 host 16Uiu2HAm2W4Ad8JCVxhsaDvpfw1yEbY6xE8PhaTkbZeXhWJsnzw5 dialing 16Uiu2HAmLJsKV6kn8Yq7S8kKdynNvJg6g1sPPWNPKGpDckXQ9QSD
17232 2024-04-26T19:03:08.795-0600      DEBUG   swarm2  swarm/swarm_dial.go:239 dialing peer   {"from": "16Uiu2HAm2W4Ad8JCVxhsaDvpfw1yEbY6xE8PhaTkbZeXhWJsnzw5", "to": "16Uiu2HAmLJsKV6kn8Yq7S8kKdynNvJg6g1sPPWNPKGpDckXQ9QSD"}
17232 2024-04-26T19:03:08.795-0600      DEBUG   swarm2  swarm/swarm_dial.go:277 network for 16Uiu2HAm2W4Ad8JCVxhsaDvpfw1yEbY6xE8PhaTkbZeXhWJsnzw5 finished dialing 16Uiu2HAmLJsKV6kn8Yq7S8kKdynNvJg6g1sPPWNPKGpDckXQ9QSD
17232 2024-04-26T19:03:08.795-0600      DEBUG   swarm2  swarm/limiter.go:201    [limiter] clearing all peer dials: 16Uiu2HAmLJsKV6kn8Yq7S8kKdynNvJg6g1sPPWNPKGpDckXQ9QSD
17232 time="2024-04-26T19:03:08-06:00" level=error msg="failed to create stream" error="failed to dial: failed to dial 16Uiu2HAmLJsKV6kn8Yq7S8kKdynNvJg6g1sPPWNPKGpDckXQ9QSD: all dials failed\n  * [/ip4/192.167.10.14/tcp/1083] dial backoff"
17232 2024-04-26T19:03:08.800-0600      DEBUG   basichost       basic/basic_host.go:737 host 16Uiu2HAm2W4Ad8JCVxhsaDvpfw1yEbY6xE8PhaTkbZeXhWJsnzw5 dialing 16Uiu2HAkymh6L495yRF9gF9usSrrG59XnR2DvMHEvtDi4zm85idq
17232 2024-04-26T19:03:08.800-0600      DEBUG   swarm2  swarm/swarm_dial.go:239 dialing peer   {"from": "16Uiu2HAm2W4Ad8JCVxhsaDvpfw1yEbY6xE8PhaTkbZeXhWJsnzw5", "to": "16Uiu2HAkymh6L495yRF9gF9usSrrG59XnR2DvMHEvtDi4zm85idq"}
17232 2024-04-26T19:03:08.800-0600      DEBUG   swarm2  swarm/swarm_dial.go:277 network for 16Uiu2HAm2W4Ad8JCVxhsaDvpfw1yEbY6xE8PhaTkbZeXhWJsnzw5 finished dialing 16Uiu2HAkymh6L495yRF9gF9usSrrG59XnR2DvMHEvtDi4zm85idq
17232 time="2024-04-26T19:03:08-06:00" level=error msg="failed to create stream" error="failed to dial: failed to dial 16Uiu2HAkymh6L495yRF9gF9usSrrG59XnR2DvMHEvtDi4zm85idq: all dials failed\n  * [/ip4/192.167.10.15/tcp/1084] dial backoff"
```