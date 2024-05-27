# Documentation on Testing

The DPSS functionality can be tested in different ways:
- unit tests (available per message handler)
- integration tests (separate tests for dacss & old committee)
- end-to-batch reconstruction test
- testing with actual nodes

The last 2 options aim to test the full DPSS flow until the end of batch reconstruction.

This document gives more details about the end-to-batch reconstruction (integration) test and how to test with actual nodes.

## 1. Testing with Actual Nodes
- We implement the `PSSParticipant` interface for `PSSNode`(the actual node) inside `adkg/dpss/pss_node.go`.
- The Transport using the `MessageBroker` is implemented inside `adkg/dpss/pss_transport.go` for sending and receiving messages over p2p network.
- The `PssService` is implemented in `adkg/dpss/pss_service.go`. For the testing, we use `adkg/dpss/pss_test_service.go`.
- The `adkg/manager_process/main.go` manages the main process and child process for the start of DPSS.
- The `adkg/manager/manager_service.go` is for signaling the manager process regarding the current DPSS status.
- `adkg/nodelist-pyscript/node_list.py` The python-script for interacting with the contract.

### Test Flow(manager_process/README.md):
- **Setup**(run: `./test_setup1.sh`)
    - Create python virtual env: `python -m venv ./venv`
    - Activate venv: `source ./venv/bin/activate`
    - Install packages: `pip install -r ./nodelist-pyscript/requirements.txt`
    - Check current PSS status from epoch1 to epoch2 `python nodelist-pyscript/node_list.py -p 1 2`, it should returns 0.
    - Check current epoch `python nodelist-pyscript/node_list.py -e`, it should returns 1.
    - The nodes are already whitelisted for a particular epoch. If not, it can be whitelisted in an epoch: `python nodelist-pyscript/node_list.py -w [epoch] -a [address]`.
    - All the commands can be found in the `manager_process/README.md` readme file.

- **Running of DPSS** 
    - **Create the config files for the nodes**(run: `./test_setup2.sh`)
        - Build the node process ``go build -o adkgNode main.go``. For testing: `go build -tags test -o adkgNode main.go`
        - Build the manager process `go build -o nodeManager ./manager_process`
        - Set up multiple config directories and files (the default config dir is ``./manager_process/new-node-config``) e.g. ``mkdir ./manager_process/new-node-config2 && mkdir ./manager_process/new-node-config3``
        - Create/Copy config files for the newly created config dirs. Test config files can be found in `./local-setup-data`. `config.local.[1~4].json` are configs for nodes in epoch 1, `config.local.[5~8].json` are configs for nodes in epoch 2.
    -  **Start nodes with different config dirs** (run: `./test_setup3.sh` **NOTE:** this script opens different terminal for macOS and maynot work on linux)
        -  open four terminals and runs `./nodeManager`, `./nodeManager -config ./manager_process/new-node-config2`, `./nodeManager -config ./manager_process/new-node-config3` and `./nodeManager -config ./manager_process/new-node-config4`.
    - Wait for all four nodes to output log `time="" level=info msg="started tendermint"...`
    - **Update the config and change the PSS status**(run: `./test_step4.sh`)
        - Update config files in `./manager_process/new-node-config*` to test configs for nodes in epoch 2. (Change `config.local.[1~4].json` to `config.local.[5~8].json`.)
        - Change PSS status (change epoch1 -> epoch2 to running): `python nodelist-pyscript/node_list.py -pc 1 2 1`
        - manager process should start new nodes in the new committee.
- **Reset everything for next test**(run: `./test_setup_reset.sh`)
- **NOTE:**  you might also need to clear the temporary data in the machine tmp folder: `rm -r /tmp/keygen-data*`
    
---   

## 2. Local End-to-Batchreconstruction Testing

- The current local integration testing test the following:
    **DACSS -> MVBA -> BatchReconstruction**
- The test can be found in `/adkg/dpss/dpss_end_to_end_testing/end_to_batch_rec_test.go`
- The test creates the Integration test setup by implementing the `PSSParticipant` interface for the `IntegrationTestNode` and Mocking the transport for sending, receiving and broadcasting of messages.
- Test details:
    - creating the `TestSetUp` and mock `transport` of nodes of Old and New committee.
    - constructing (old)shares for testing using `GenerateSecretAndGetCommitmentAndShares`.
    - constructing InitMsg for each old node by generating the ephemeral keys and using the shares constructed in the previous step.
    - constructing an empty state for each old and new nodes for each acssRound.
    - Start of DPSS for each old nodes.

- Once all the messages are processes from all the handler after sufficient wait time, we check the following assertions:
    - DACSS
        - check that the RBC ended for all nodes
        - extract the shamir shares of the random scalar from each nodes(both old and new) state received from the OldNode[0] during DACSS.
        - reconstruct the shares and check that it is equal to the random scalar sent by the oldNode[0].
    - BatchReconstruction
        - Check each HimHandler invocation should send 1 preprocess message. So in total we expect n(=len(OldNodes)) many PreProcessMessages
        - check each PreprocessBatchRecMessage should send $ceil(B/(n-2t))$ InitRecHandlerMessages
        - check from each InitRecHandlerMessage we expect n_old PrivateRecHandlerMessages to be sent. so total PrivateRecHandlerMessages will be equal to `nrInitRecMessages*nOld`
        - check each old node should broadcast 1 PublicRecMessage per batch round. There are nrBatches=ceil(B/(n-2t)) batch rounds. so total number of broadcasted msg will be `nrBatches*nOld`