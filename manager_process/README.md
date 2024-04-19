# Manager Process

## Setup python env
1. Create python virtual env: `python -m venv ./venv`
2. Activate venv: `source ./venv/bin/activate`
3. Install packages: `pip install -r ./nodelist-pyscript/requirements.txt`
4. Check current PSS status from epoch1 to epoch2 `python nodelist-pyscript/node_list.py -p 1 2`, it should returns 0.
5. Check current epoch `python nodelist-pyscript/node_list.py -e`, it should returns 1.
6. Check balance of owner `python nodelist-pyscript/node_list.py -b`, hopefully more than 0. (owner's address: `0xaAbb43AF42c823b75d39673C99197cd2B10Fa3E1`)
6. To exit python virtual env: `deactivate`

## Start a single manager/node processes
1. Build the node process `go build -o adkgNode main.go`
2. Build the manager process `go build -o nodeManager ./manager_process`
3. Run the manager & node with default config in ./manager_process/new-node-config: `./nodeManager`

## Run multiple managers/nodes
1. Build the node process ``go build -o adkgNode main.go``
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


## Start 1 process manually (directly run binary instead of through main)

To test how a correct startup could be and see what logging looks like. 

In `adkg`:
```
./node start --config local-setup-data/config.local.1.json --secret-config local-setup-data/config.local.1.json
```