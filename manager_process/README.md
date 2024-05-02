# Manager Process
* faucet for arbitrum-sepolia https://www.alchemy.com/faucets/arbitrum-sepolia
## Setup python env
1. Create python virtual env: `python -m venv ./venv`
2. Activate venv: `source ./venv/bin/activate`
3. Install packages: `pip install -r ./nodelist-pyscript/requirements.txt`
4. Check current PSS status from epoch1 to epoch2 `python nodelist-pyscript/node_list.py -p 1 2`, it should returns 0.
5. Check current epoch `python nodelist-pyscript/node_list.py -e`, it should returns 1.
6. Check balance of owner `python nodelist-pyscript/node_list.py -bo`, hopefully more than 0. (owner's address: `0xaAbb43AF42c823b75d39673C99197cd2B10Fa3E1`)
6. To exit python virtual env: `deactivate`

## Python script
* owner's sk = `0x8d33ef20ec6519d7242aeee66e67d0771a794fce356a22ade91df0731efe99b8`
Addres = `0xaAbb43AF42c823b75d39673C99197cd2B10Fa3E1`
* owner can whitelist a node, update the epoch, set the epoch info

#### Balance / Address / Send Eth
* Check owner's balance: `python nodelist-pyscript/node_list.py -bo`
* Check balance of an address: `python nodelist-pyscript/node_list.py -b [addr]`
* Get address from a private key: `python nodelist-pyscript/node_list.py -afk [key]`
* Send v Eth form A to B: `python nodelist-pyscript/node_list.py -s -sk [A's private key] -to [B's address] -v [value in eth]`

#### Get epoch / Pss status
* Get current epoch:  `python nodelist-pyscript/node_list.py -e`
* Get epoch info: `python nodelist-pyscript/node_list.py -ef [epoch]`
* Check current PSS status from epoch1 to epoch2: `python nodelist-pyscript/node_list.py -p [epoch1] [epoch2]`

#### Set epoch / Pss status
* Change current epoch: `python nodelist-pyscript/node_list.py -ec [epoch]`
* Set epoch info: `python nodelist-pyscript/node_list.py -se [epoch] -n [N] -k [K] -t [T]` (Warning: this funcion will unregister nodes in the epoch)
* Set pss status from [epoch1] to [epoch2] to [1 or 0]: `python nodelist-pyscript/node_list.py -pc [epoch1] [epoch2] [1 or 0]`

#### whitelist
* Whitelist a node in an epoch: `python nodelist-pyscript/node_list.py -w [epoch] -a [address]` (address can be get from key using -afk, see Balance / Address)
* Check if a node is whitelisted in an epoch: `python nodelist-pyscript/node_list.py -iw [epoch] -a [address]`


## How to set up a committee
1. Prepare config files for all the nodes
2. Whitelist nodes for the epoch:  `python nodelist-pyscript/node_list.py -w [epoch] -a [address]`
3. Check if a node is whitelisted in an epoch: `python nodelist-pyscript/node_list.py -iw [epoch] -a [address]` 
3. Set the epoch info: `python nodelist-pyscript/node_list.py -se [epoch] -n [N] -k [K] -t [T]` (Warning: this funcion will unregister nodes in the epoch)
4. Fund all the nodes so they can registered themself at start. Send v Eth form A to B: `python nodelist-pyscript/node_list.py -s -sk [A's private key] -to [B's address] -v [value in eth]` **(0.002 eth shoud be enough. However, if unable to start the committee, it's most likely due to not enough balance. Try increase balance for nodes.)**
5. Start all the nodes as below

## Start a single manager/node processes
1. Build the node process `go build -o adkgNode main.go`. For testing: `go build -tags test -o adkgNode main.go`
2. Build the manager process `go build -o nodeManager ./manager_process`
3. Run the manager & node with default config in ./manager_process/new-node-config: `./nodeManager`

## Run multiple managers/nodes
1. Build the node process ``go build -o adkgNode main.go``. For testing: `go build -tags test -o adkgNode main.go`
2. Build the manager process ``go build -o nodeManager ./manager_process``
3. Set up multiple config directories and files (the default config dir is ``./manager_process/new-node-config``) e.g. ``mkdir ./manager_process/new-node-config2 && mkdir ./manager_process/new-node-config3``
4. Create/Copy config files for the newly created config dirs. Test config files can be found in `./local-setup-data`. `config.local.[1~3].json` are configs for nodes in epoch 1, `config.local.[4~6].json` are configs for nodes in epoch 2.
5. To start 3 nodes with different config dirs, open three terminals and run `./nodeManager`, `./nodeManager -config ./manager_process/new-node-config2`, and `./nodeManager -config ./manager_process/new-node-config3`.
6. Wait for all three nodes to output log `time="" level=info msg="started tendermint"...` 

Note: addresses in `config.local.[1~6].json` have been whitelisted and registered. If you want to use a new address, first whitelist it to the desired epoch. Then transfer enough funds to the address before starting a node with the address. (Whitelist command not yet created in script)

## Start PSS from epoch 1 to 2
1. Update config files in `./manager_process/new-node-config*` to test configs for nodes in epoch 2. (Change `config.local.[1~3].json` to `config.local.[4~6].json`.)
2. Use the Python script to change PSS status (change epoch1 -> epoch2 to running): `python nodelist-pyscript/node_list.py -pc 1 2 1`
3. manager process should start new nodes in the new committee.

## Change epoch
1. Use the Python script to change epoch to epoch 2: `python nodelist-pyscript/node_list.py -ec 2`
2. The manger process should stop the old node

## Reset contract for next testing
1. Change PSS status from epoch1 to epoch2 to 0: `python nodelist-pyscript/node_list.py -pc 1 2 0`
2. Change epoch back to epoch 1: `python nodelist-pyscript/node_list.py -ec 1`
3. Change the config in config dirs back to `config.local.[1~3].json`

## Set vscode for build tags:
1. open control panel: (command+shift+p / control+shift+p)
2. search "open user settings(JSON)"
3.  "gopls": {
        "buildFlags": [
            "-tags=test"
        ]
    }



## Start 1 process manually (directly run binary instead of through main)

To test how a correct startup could be and see what logging looks like. 

In `adkg`:
```
./node start --config local-setup-data/config.local.1.json --secret-config local-setup-data/config.local.1.json
```