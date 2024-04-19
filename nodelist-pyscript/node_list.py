import argparse
from web3 import Web3
from web3.middleware import geth_poa_middleware

abi = '[  {   "inputs": [],   "name": "InvalidInitialization",   "type": "error"  },  {   "inputs": [],   "name": "NotInitializing",   "type": "error"  },  {   "inputs": [    {     "internalType": "address",     "name": "owner",     "type": "address"    }   ],   "name": "OwnableInvalidOwner",   "type": "error"  },  {   "inputs": [    {     "internalType": "address",     "name": "account",     "type": "address"    }   ],   "name": "OwnableUnauthorizedAccount",   "type": "error"  },  {   "anonymous": false,   "inputs": [    {     "indexed": false,     "internalType": "uint256",     "name": "oldEpoch",     "type": "uint256"    },    {     "indexed": false,     "internalType": "uint256",     "name": "newEpoch",     "type": "uint256"    }   ],   "name": "EpochChanged",   "type": "event"  },  {   "anonymous": false,   "inputs": [    {     "indexed": false,     "internalType": "uint64",     "name": "version",     "type": "uint64"    }   ],   "name": "Initialized",   "type": "event"  },  {   "anonymous": false,   "inputs": [    {     "indexed": false,     "internalType": "address",     "name": "publicKey",     "type": "address"    },    {     "indexed": false,     "internalType": "uint256",     "name": "epoch",     "type": "uint256"    },    {     "indexed": false,     "internalType": "uint256",     "name": "position",     "type": "uint256"    }   ],   "name": "NodeListed",   "type": "event"  },  {   "anonymous": false,   "inputs": [    {     "indexed": true,     "internalType": "address",     "name": "previousOwner",     "type": "address"    },    {     "indexed": true,     "internalType": "address",     "name": "newOwner",     "type": "address"    }   ],   "name": "OwnershipTransferred",   "type": "event"  },  {   "inputs": [],   "name": "clearAllEpoch",   "outputs": [],   "stateMutability": "nonpayable",   "type": "function"  },  {   "inputs": [],   "name": "currentEpoch",   "outputs": [    {     "internalType": "uint256",     "name": "",     "type": "uint256"    }   ],   "stateMutability": "view",   "type": "function"  },  {   "inputs": [    {     "internalType": "uint256",     "name": "",     "type": "uint256"    }   ],   "name": "epochInfo",   "outputs": [    {     "internalType": "uint256",     "name": "id",     "type": "uint256"    },    {     "internalType": "uint256",     "name": "n",     "type": "uint256"    },    {     "internalType": "uint256",     "name": "k",     "type": "uint256"    },    {     "internalType": "uint256",     "name": "t",     "type": "uint256"    },    {     "internalType": "uint256",     "name": "prevEpoch",     "type": "uint256"    },    {     "internalType": "uint256",     "name": "nextEpoch",     "type": "uint256"    }   ],   "stateMutability": "view",   "type": "function"  },  {   "inputs": [],   "name": "getCurrentEpochDetails",   "outputs": [    {     "components": [      {       "internalType": "string",       "name": "declaredIp",       "type": "string"      },      {       "internalType": "uint256",       "name": "position",       "type": "uint256"      },      {       "internalType": "uint256",       "name": "pubKx",       "type": "uint256"      },      {       "internalType": "uint256",       "name": "pubKy",       "type": "uint256"      },      {       "internalType": "string",       "name": "tmP2PListenAddress",       "type": "string"      },      {       "internalType": "string",       "name": "p2pListenAddress",       "type": "string"      }     ],     "internalType": "struct NodeList.Details[]",     "name": "nodes",     "type": "tuple[]"    }   ],   "stateMutability": "view",   "type": "function"  },  {   "inputs": [    {     "internalType": "uint256",     "name": "epoch",     "type": "uint256"    }   ],   "name": "getEpochInfo",   "outputs": [    {     "internalType": "uint256",     "name": "id",     "type": "uint256"    },    {     "internalType": "uint256",     "name": "n",     "type": "uint256"    },    {     "internalType": "uint256",     "name": "k",     "type": "uint256"    },    {     "internalType": "uint256",     "name": "t",     "type": "uint256"    },    {     "internalType": "address[]",     "name": "nodeList",     "type": "address[]"    },    {     "internalType": "uint256",     "name": "prevEpoch",     "type": "uint256"    },    {     "internalType": "uint256",     "name": "nextEpoch",     "type": "uint256"    }   ],   "stateMutability": "view",   "type": "function"  },  {   "inputs": [    {     "internalType": "address",     "name": "nodeAddress",     "type": "address"    }   ],   "name": "getNodeDetails",   "outputs": [    {     "internalType": "string",     "name": "declaredIp",     "type": "string"    },    {     "internalType": "uint256",     "name": "position",     "type": "uint256"    },    {     "internalType": "string",     "name": "tmP2PListenAddress",     "type": "string"    },    {     "internalType": "string",     "name": "p2pListenAddress",     "type": "string"    }   ],   "stateMutability": "view",   "type": "function"  },  {   "inputs": [    {     "internalType": "uint256",     "name": "epoch",     "type": "uint256"    }   ],   "name": "getNodes",   "outputs": [    {     "internalType": "address[]",     "name": "",     "type": "address[]"    }   ],   "stateMutability": "view",   "type": "function"  },  {   "inputs": [    {     "internalType": "uint256",     "name": "oldEpoch",     "type": "uint256"    },    {     "internalType": "uint256",     "name": "newEpoch",     "type": "uint256"    }   ],   "name": "getPssStatus",   "outputs": [    {     "internalType": "uint256",     "name": "",     "type": "uint256"    }   ],   "stateMutability": "view",   "type": "function"  },  {   "inputs": [    {     "internalType": "uint256",     "name": "_epoch",     "type": "uint256"    }   ],   "name": "initialize",   "outputs": [],   "stateMutability": "nonpayable",   "type": "function"  },  {   "inputs": [    {     "internalType": "uint256",     "name": "epoch",     "type": "uint256"    },    {     "internalType": "address",     "name": "nodeAddress",     "type": "address"    }   ],   "name": "isWhitelisted",   "outputs": [    {     "internalType": "bool",     "name": "",     "type": "bool"    }   ],   "stateMutability": "view",   "type": "function"  },  {   "inputs": [    {     "internalType": "uint256",     "name": "epoch",     "type": "uint256"    },    {     "internalType": "string",     "name": "declaredIp",     "type": "string"    },    {     "internalType": "uint256",     "name": "pubKx",     "type": "uint256"    },    {     "internalType": "uint256",     "name": "pubKy",     "type": "uint256"    },    {     "internalType": "string",     "name": "tmP2PListenAddress",     "type": "string"    },    {     "internalType": "string",     "name": "p2pListenAddress",     "type": "string"    }   ],   "name": "listNode",   "outputs": [],   "stateMutability": "nonpayable",   "type": "function"  },  {   "inputs": [    {     "internalType": "address",     "name": "",     "type": "address"    }   ],   "name": "nodeDetails",   "outputs": [    {     "internalType": "string",     "name": "declaredIp",     "type": "string"    },    {     "internalType": "uint256",     "name": "position",     "type": "uint256"    },    {     "internalType": "uint256",     "name": "pubKx",     "type": "uint256"    },    {     "internalType": "uint256",     "name": "pubKy",     "type": "uint256"    },    {     "internalType": "string",     "name": "tmP2PListenAddress",     "type": "string"    },    {     "internalType": "string",     "name": "p2pListenAddress",     "type": "string"    }   ],   "stateMutability": "view",   "type": "function"  },  {   "inputs": [    {     "internalType": "uint256",     "name": "epoch",     "type": "uint256"    },    {     "internalType": "address",     "name": "nodeAddress",     "type": "address"    }   ],   "name": "nodeRegistered",   "outputs": [    {     "internalType": "bool",     "name": "",     "type": "bool"    }   ],   "stateMutability": "view",   "type": "function"  },  {   "inputs": [],   "name": "owner",   "outputs": [    {     "internalType": "address",     "name": "",     "type": "address"    }   ],   "stateMutability": "view",   "type": "function"  },  {   "inputs": [    {     "internalType": "uint256",     "name": "",     "type": "uint256"    },    {     "internalType": "uint256",     "name": "",     "type": "uint256"    }   ],   "name": "pssStatus",   "outputs": [    {     "internalType": "uint256",     "name": "",     "type": "uint256"    }   ],   "stateMutability": "view",   "type": "function"  },  {   "inputs": [],   "name": "renounceOwnership",   "outputs": [],   "stateMutability": "nonpayable",   "type": "function"  },  {   "inputs": [    {     "internalType": "uint256",     "name": "_newEpoch",     "type": "uint256"    }   ],   "name": "setCurrentEpoch",   "outputs": [],   "stateMutability": "nonpayable",   "type": "function"  },  {   "inputs": [    {     "internalType": "address",     "name": "newOwner",     "type": "address"    }   ],   "name": "transferOwnership",   "outputs": [],   "stateMutability": "nonpayable",   "type": "function"  },  {   "inputs": [    {     "internalType": "uint256",     "name": "epoch",     "type": "uint256"    },    {     "internalType": "uint256",     "name": "n",     "type": "uint256"    },    {     "internalType": "uint256",     "name": "k",     "type": "uint256"    },    {     "internalType": "uint256",     "name": "t",     "type": "uint256"    },    {     "internalType": "address[]",     "name": "nodeList",     "type": "address[]"    },    {     "internalType": "uint256",     "name": "prevEpoch",     "type": "uint256"    },    {     "internalType": "uint256",     "name": "nextEpoch",     "type": "uint256"    }   ],   "name": "updateEpoch",   "outputs": [],   "stateMutability": "nonpayable",   "type": "function"  },  {   "inputs": [    {     "internalType": "uint256",     "name": "oldEpoch",     "type": "uint256"    },    {     "internalType": "uint256",     "name": "newEpoch",     "type": "uint256"    },    {     "internalType": "uint256",     "name": "status",     "type": "uint256"    }   ],   "name": "updatePssStatus",   "outputs": [],   "stateMutability": "nonpayable",   "type": "function"  },  {   "inputs": [    {     "internalType": "uint256",     "name": "epoch",     "type": "uint256"    },    {     "internalType": "address",     "name": "nodeAddress",     "type": "address"    },    {     "internalType": "bool",     "name": "allowed",     "type": "bool"    }   ],   "name": "updateWhitelist",   "outputs": [],   "stateMutability": "nonpayable",   "type": "function"  },  {   "inputs": [    {     "internalType": "uint256",     "name": "",     "type": "uint256"    },    {     "internalType": "address",     "name": "",     "type": "address"    }   ],   "name": "whitelist",   "outputs": [    {     "internalType": "bool",     "name": "",     "type": "bool"    }   ],   "stateMutability": "view",   "type": "function"  } ]'
address = "0x7c20cB99e1F2CD1ECd1B425A51ff66D0f01E0Eda"
owner_sk = "0x8d33ef20ec6519d7242aeee66e67d0771a794fce356a22ade91df0731efe99b8"


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "-pc",
        "--PssStatusChange",
        help="change PSS status for epoch x, y to z",
        nargs="+",
        type=int,
    )
    parser.add_argument(
        "-p",
        "--GetPssStatus",
        help="get PSS status for epoch x, y",
        nargs="+",
        type=int,
    )
    parser.add_argument("-ec", "--EpochChange", help="change epoch", type=int)
    parser.add_argument("-e", "--GetEpoch", help="get epoch", action="store_true")
    parser.add_argument(
        "-b", "--CheckBalance", help="check balance of owner", action="store_true"
    )

    args = parser.parse_args()

    w3 = Web3(
        Web3.HTTPProvider("https://arbitrum-sepolia.blockpi.network/v1/rpc/public")
    )
    w3.middleware_onion.inject(geth_poa_middleware, layer=0)
    contract = w3.eth.contract(address=address, abi=abi)
    owner = w3.eth.account.from_key(owner_sk)

    if args.PssStatusChange:
        setPssStatus(
            args.PssStatusChange[0],
            args.PssStatusChange[1],
            args.PssStatusChange[2],
            contract,
            owner,
            w3,
        )

    if args.GetPssStatus:
        getPssStatus(contract, args.GetPssStatus[0], args.GetPssStatus[1])

    if args.EpochChange:
        setCurrentEpoch(args.EpochChange, owner, contract, w3)

    if args.GetEpoch:
        getCurrentEpoch(contract)

    if args.CheckBalance:
        print(w3.from_wei(w3.eth.get_balance(owner.address), "ether"))


def getCurrentEpoch(contract):
    print(contract.functions.currentEpoch().call())


def getPssStatus(contract, oldEpoch, newEpoch):
    print(contract.functions.pssStatus(oldEpoch, newEpoch).call())


def setPssStatus(oldEpoch, newEpoch, statusInt, contract, owner, w3):
    tx = contract.functions.updatePssStatus(
        oldEpoch, newEpoch, statusInt
    ).build_transaction(
        {
            "from": owner.address,
            "nonce": w3.eth.get_transaction_count(owner.address),
        }
    )
    signed_tx = w3.eth.account.sign_transaction(tx, private_key=owner.key)
    tx_hash = w3.eth.send_raw_transaction(signed_tx.rawTransaction)
    recp = w3.eth.wait_for_transaction_receipt(tx_hash)
    print("include in", recp.blockNumber)


def setCurrentEpoch(epoch, owner, contract, w3):
    tx = contract.functions.setCurrentEpoch(epoch).build_transaction(
        {
            "from": owner.address,
            "nonce": w3.eth.get_transaction_count(owner.address),
        }
    )
    signed_tx = w3.eth.account.sign_transaction(tx, private_key=owner.key)
    tx_hash = w3.eth.send_raw_transaction(signed_tx.rawTransaction)
    recp = w3.eth.wait_for_transaction_receipt(tx_hash)
    print("include in", recp.blockNumber)


if __name__ == "__main__":
    main()
