// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package nodelist

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// NodeListDetails is an auto generated low-level Go binding around an user-defined struct.
type NodeListDetails struct {
	DeclaredIp         string
	Position           *big.Int
	PubKx              *big.Int
	PubKy              *big.Int
	TmP2PListenAddress string
	P2pListenAddress   string
}

// NodeListMetaData contains all meta data concerning the NodeList contract.
var NodeListMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"oldEpoch\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newEpoch\",\"type\":\"uint256\"}],\"name\":\"EpochChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"EpochCleared\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"EpochUpdate\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"publicKey\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"position\",\"type\":\"uint256\"}],\"name\":\"NodeListed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"status\",\"type\":\"uint256\"}],\"name\":\"PssStatusUpdate\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"whitelistAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"isAllowed\",\"type\":\"bool\"}],\"name\":\"WhitelistUpdate\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"bufferSize\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"clearAllEpoch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"currentEpoch\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"epochInfo\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"n\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"k\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"t\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"prevEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nextEpoch\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getCurrentEpochDetails\",\"outputs\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"declaredIp\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"position\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pubKx\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pubKy\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"tmP2PListenAddress\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"p2pListenAddress\",\"type\":\"string\"}],\"internalType\":\"structNodeList.Details[]\",\"name\":\"nodes\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"}],\"name\":\"getEpochInfo\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"n\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"k\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"t\",\"type\":\"uint256\"},{\"internalType\":\"address[]\",\"name\":\"nodeList\",\"type\":\"address[]\"},{\"internalType\":\"uint256\",\"name\":\"prevEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nextEpoch\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"nodeAddress\",\"type\":\"address\"}],\"name\":\"getNodeDetails\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"declaredIp\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"position\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"tmP2PListenAddress\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"p2pListenAddress\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"}],\"name\":\"getNodes\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"oldEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"newEpoch\",\"type\":\"uint256\"}],\"name\":\"getPssStatus\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"nodeAddress\",\"type\":\"address\"}],\"name\":\"isWhitelisted\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"declaredIp\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"pubKx\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pubKy\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"tmP2PListenAddress\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"p2pListenAddress\",\"type\":\"string\"}],\"name\":\"listNode\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"nodeDetails\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"declaredIp\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"position\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pubKx\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pubKy\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"tmP2PListenAddress\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"p2pListenAddress\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"nodeAddress\",\"type\":\"address\"}],\"name\":\"nodeRegistered\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"pssStatus\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"size\",\"type\":\"uint256\"}],\"name\":\"setBufferSize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newEpoch\",\"type\":\"uint256\"}],\"name\":\"setCurrentEpoch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"n\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"k\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"t\",\"type\":\"uint256\"},{\"internalType\":\"address[]\",\"name\":\"nodeList\",\"type\":\"address[]\"},{\"internalType\":\"uint256\",\"name\":\"prevEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nextEpoch\",\"type\":\"uint256\"}],\"name\":\"updateEpoch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"oldEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"newEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"status\",\"type\":\"uint256\"}],\"name\":\"updatePssStatus\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"nodeAddress\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"allowed\",\"type\":\"bool\"}],\"name\":\"updateWhitelist\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"whitelist\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// NodeListABI is the input ABI used to generate the binding from.
// Deprecated: Use NodeListMetaData.ABI instead.
var NodeListABI = NodeListMetaData.ABI

// NodeList is an auto generated Go binding around an Ethereum contract.
type NodeList struct {
	NodeListCaller     // Read-only binding to the contract
	NodeListTransactor // Write-only binding to the contract
	NodeListFilterer   // Log filterer for contract events
}

// NodeListCaller is an auto generated read-only Go binding around an Ethereum contract.
type NodeListCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// NodeListTransactor is an auto generated write-only Go binding around an Ethereum contract.
type NodeListTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// NodeListFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type NodeListFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// NodeListSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type NodeListSession struct {
	Contract     *NodeList         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// NodeListCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type NodeListCallerSession struct {
	Contract *NodeListCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// NodeListTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type NodeListTransactorSession struct {
	Contract     *NodeListTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// NodeListRaw is an auto generated low-level Go binding around an Ethereum contract.
type NodeListRaw struct {
	Contract *NodeList // Generic contract binding to access the raw methods on
}

// NodeListCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type NodeListCallerRaw struct {
	Contract *NodeListCaller // Generic read-only contract binding to access the raw methods on
}

// NodeListTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type NodeListTransactorRaw struct {
	Contract *NodeListTransactor // Generic write-only contract binding to access the raw methods on
}

// NewNodeList creates a new instance of NodeList, bound to a specific deployed contract.
func NewNodeList(address common.Address, backend bind.ContractBackend) (*NodeList, error) {
	contract, err := bindNodeList(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &NodeList{NodeListCaller: NodeListCaller{contract: contract}, NodeListTransactor: NodeListTransactor{contract: contract}, NodeListFilterer: NodeListFilterer{contract: contract}}, nil
}

// NewNodeListCaller creates a new read-only instance of NodeList, bound to a specific deployed contract.
func NewNodeListCaller(address common.Address, caller bind.ContractCaller) (*NodeListCaller, error) {
	contract, err := bindNodeList(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &NodeListCaller{contract: contract}, nil
}

// NewNodeListTransactor creates a new write-only instance of NodeList, bound to a specific deployed contract.
func NewNodeListTransactor(address common.Address, transactor bind.ContractTransactor) (*NodeListTransactor, error) {
	contract, err := bindNodeList(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &NodeListTransactor{contract: contract}, nil
}

// NewNodeListFilterer creates a new log filterer instance of NodeList, bound to a specific deployed contract.
func NewNodeListFilterer(address common.Address, filterer bind.ContractFilterer) (*NodeListFilterer, error) {
	contract, err := bindNodeList(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &NodeListFilterer{contract: contract}, nil
}

// bindNodeList binds a generic wrapper to an already deployed contract.
func bindNodeList(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(NodeListABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_NodeList *NodeListRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _NodeList.Contract.NodeListCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_NodeList *NodeListRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _NodeList.Contract.NodeListTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_NodeList *NodeListRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _NodeList.Contract.NodeListTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_NodeList *NodeListCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _NodeList.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_NodeList *NodeListTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _NodeList.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_NodeList *NodeListTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _NodeList.Contract.contract.Transact(opts, method, params...)
}

// BufferSize is a free data retrieval call binding the contract method 0x9c2c770b.
//
// Solidity: function bufferSize() view returns(uint256)
func (_NodeList *NodeListCaller) BufferSize(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _NodeList.contract.Call(opts, &out, "bufferSize")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BufferSize is a free data retrieval call binding the contract method 0x9c2c770b.
//
// Solidity: function bufferSize() view returns(uint256)
func (_NodeList *NodeListSession) BufferSize() (*big.Int, error) {
	return _NodeList.Contract.BufferSize(&_NodeList.CallOpts)
}

// BufferSize is a free data retrieval call binding the contract method 0x9c2c770b.
//
// Solidity: function bufferSize() view returns(uint256)
func (_NodeList *NodeListCallerSession) BufferSize() (*big.Int, error) {
	return _NodeList.Contract.BufferSize(&_NodeList.CallOpts)
}

// CurrentEpoch is a free data retrieval call binding the contract method 0x76671808.
//
// Solidity: function currentEpoch() view returns(uint256)
func (_NodeList *NodeListCaller) CurrentEpoch(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _NodeList.contract.Call(opts, &out, "currentEpoch")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CurrentEpoch is a free data retrieval call binding the contract method 0x76671808.
//
// Solidity: function currentEpoch() view returns(uint256)
func (_NodeList *NodeListSession) CurrentEpoch() (*big.Int, error) {
	return _NodeList.Contract.CurrentEpoch(&_NodeList.CallOpts)
}

// CurrentEpoch is a free data retrieval call binding the contract method 0x76671808.
//
// Solidity: function currentEpoch() view returns(uint256)
func (_NodeList *NodeListCallerSession) CurrentEpoch() (*big.Int, error) {
	return _NodeList.Contract.CurrentEpoch(&_NodeList.CallOpts)
}

// EpochInfo is a free data retrieval call binding the contract method 0x3894228e.
//
// Solidity: function epochInfo(uint256 ) view returns(uint256 id, uint256 n, uint256 k, uint256 t, uint256 prevEpoch, uint256 nextEpoch)
func (_NodeList *NodeListCaller) EpochInfo(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Id        *big.Int
	N         *big.Int
	K         *big.Int
	T         *big.Int
	PrevEpoch *big.Int
	NextEpoch *big.Int
}, error) {
	var out []interface{}
	err := _NodeList.contract.Call(opts, &out, "epochInfo", arg0)

	outstruct := new(struct {
		Id        *big.Int
		N         *big.Int
		K         *big.Int
		T         *big.Int
		PrevEpoch *big.Int
		NextEpoch *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Id = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.N = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.K = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.T = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.PrevEpoch = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.NextEpoch = *abi.ConvertType(out[5], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// EpochInfo is a free data retrieval call binding the contract method 0x3894228e.
//
// Solidity: function epochInfo(uint256 ) view returns(uint256 id, uint256 n, uint256 k, uint256 t, uint256 prevEpoch, uint256 nextEpoch)
func (_NodeList *NodeListSession) EpochInfo(arg0 *big.Int) (struct {
	Id        *big.Int
	N         *big.Int
	K         *big.Int
	T         *big.Int
	PrevEpoch *big.Int
	NextEpoch *big.Int
}, error) {
	return _NodeList.Contract.EpochInfo(&_NodeList.CallOpts, arg0)
}

// EpochInfo is a free data retrieval call binding the contract method 0x3894228e.
//
// Solidity: function epochInfo(uint256 ) view returns(uint256 id, uint256 n, uint256 k, uint256 t, uint256 prevEpoch, uint256 nextEpoch)
func (_NodeList *NodeListCallerSession) EpochInfo(arg0 *big.Int) (struct {
	Id        *big.Int
	N         *big.Int
	K         *big.Int
	T         *big.Int
	PrevEpoch *big.Int
	NextEpoch *big.Int
}, error) {
	return _NodeList.Contract.EpochInfo(&_NodeList.CallOpts, arg0)
}

// GetCurrentEpochDetails is a free data retrieval call binding the contract method 0x0c56e48f.
//
// Solidity: function getCurrentEpochDetails() view returns((string,uint256,uint256,uint256,string,string)[] nodes)
func (_NodeList *NodeListCaller) GetCurrentEpochDetails(opts *bind.CallOpts) ([]NodeListDetails, error) {
	var out []interface{}
	err := _NodeList.contract.Call(opts, &out, "getCurrentEpochDetails")

	if err != nil {
		return *new([]NodeListDetails), err
	}

	out0 := *abi.ConvertType(out[0], new([]NodeListDetails)).(*[]NodeListDetails)

	return out0, err

}

// GetCurrentEpochDetails is a free data retrieval call binding the contract method 0x0c56e48f.
//
// Solidity: function getCurrentEpochDetails() view returns((string,uint256,uint256,uint256,string,string)[] nodes)
func (_NodeList *NodeListSession) GetCurrentEpochDetails() ([]NodeListDetails, error) {
	return _NodeList.Contract.GetCurrentEpochDetails(&_NodeList.CallOpts)
}

// GetCurrentEpochDetails is a free data retrieval call binding the contract method 0x0c56e48f.
//
// Solidity: function getCurrentEpochDetails() view returns((string,uint256,uint256,uint256,string,string)[] nodes)
func (_NodeList *NodeListCallerSession) GetCurrentEpochDetails() ([]NodeListDetails, error) {
	return _NodeList.Contract.GetCurrentEpochDetails(&_NodeList.CallOpts)
}

// GetEpochInfo is a free data retrieval call binding the contract method 0x135022c2.
//
// Solidity: function getEpochInfo(uint256 epoch) view returns(uint256 id, uint256 n, uint256 k, uint256 t, address[] nodeList, uint256 prevEpoch, uint256 nextEpoch)
func (_NodeList *NodeListCaller) GetEpochInfo(opts *bind.CallOpts, epoch *big.Int) (struct {
	Id        *big.Int
	N         *big.Int
	K         *big.Int
	T         *big.Int
	NodeList  []common.Address
	PrevEpoch *big.Int
	NextEpoch *big.Int
}, error) {
	var out []interface{}
	err := _NodeList.contract.Call(opts, &out, "getEpochInfo", epoch)

	outstruct := new(struct {
		Id        *big.Int
		N         *big.Int
		K         *big.Int
		T         *big.Int
		NodeList  []common.Address
		PrevEpoch *big.Int
		NextEpoch *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Id = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.N = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.K = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.T = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.NodeList = *abi.ConvertType(out[4], new([]common.Address)).(*[]common.Address)
	outstruct.PrevEpoch = *abi.ConvertType(out[5], new(*big.Int)).(**big.Int)
	outstruct.NextEpoch = *abi.ConvertType(out[6], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// GetEpochInfo is a free data retrieval call binding the contract method 0x135022c2.
//
// Solidity: function getEpochInfo(uint256 epoch) view returns(uint256 id, uint256 n, uint256 k, uint256 t, address[] nodeList, uint256 prevEpoch, uint256 nextEpoch)
func (_NodeList *NodeListSession) GetEpochInfo(epoch *big.Int) (struct {
	Id        *big.Int
	N         *big.Int
	K         *big.Int
	T         *big.Int
	NodeList  []common.Address
	PrevEpoch *big.Int
	NextEpoch *big.Int
}, error) {
	return _NodeList.Contract.GetEpochInfo(&_NodeList.CallOpts, epoch)
}

// GetEpochInfo is a free data retrieval call binding the contract method 0x135022c2.
//
// Solidity: function getEpochInfo(uint256 epoch) view returns(uint256 id, uint256 n, uint256 k, uint256 t, address[] nodeList, uint256 prevEpoch, uint256 nextEpoch)
func (_NodeList *NodeListCallerSession) GetEpochInfo(epoch *big.Int) (struct {
	Id        *big.Int
	N         *big.Int
	K         *big.Int
	T         *big.Int
	NodeList  []common.Address
	PrevEpoch *big.Int
	NextEpoch *big.Int
}, error) {
	return _NodeList.Contract.GetEpochInfo(&_NodeList.CallOpts, epoch)
}

// GetNodeDetails is a free data retrieval call binding the contract method 0xbafb3581.
//
// Solidity: function getNodeDetails(address nodeAddress) view returns(string declaredIp, uint256 position, string tmP2PListenAddress, string p2pListenAddress)
func (_NodeList *NodeListCaller) GetNodeDetails(opts *bind.CallOpts, nodeAddress common.Address) (struct {
	DeclaredIp         string
	Position           *big.Int
	TmP2PListenAddress string
	P2pListenAddress   string
}, error) {
	var out []interface{}
	err := _NodeList.contract.Call(opts, &out, "getNodeDetails", nodeAddress)

	outstruct := new(struct {
		DeclaredIp         string
		Position           *big.Int
		TmP2PListenAddress string
		P2pListenAddress   string
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.DeclaredIp = *abi.ConvertType(out[0], new(string)).(*string)
	outstruct.Position = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.TmP2PListenAddress = *abi.ConvertType(out[2], new(string)).(*string)
	outstruct.P2pListenAddress = *abi.ConvertType(out[3], new(string)).(*string)

	return *outstruct, err

}

// GetNodeDetails is a free data retrieval call binding the contract method 0xbafb3581.
//
// Solidity: function getNodeDetails(address nodeAddress) view returns(string declaredIp, uint256 position, string tmP2PListenAddress, string p2pListenAddress)
func (_NodeList *NodeListSession) GetNodeDetails(nodeAddress common.Address) (struct {
	DeclaredIp         string
	Position           *big.Int
	TmP2PListenAddress string
	P2pListenAddress   string
}, error) {
	return _NodeList.Contract.GetNodeDetails(&_NodeList.CallOpts, nodeAddress)
}

// GetNodeDetails is a free data retrieval call binding the contract method 0xbafb3581.
//
// Solidity: function getNodeDetails(address nodeAddress) view returns(string declaredIp, uint256 position, string tmP2PListenAddress, string p2pListenAddress)
func (_NodeList *NodeListCallerSession) GetNodeDetails(nodeAddress common.Address) (struct {
	DeclaredIp         string
	Position           *big.Int
	TmP2PListenAddress string
	P2pListenAddress   string
}, error) {
	return _NodeList.Contract.GetNodeDetails(&_NodeList.CallOpts, nodeAddress)
}

// GetNodes is a free data retrieval call binding the contract method 0x47de074f.
//
// Solidity: function getNodes(uint256 epoch) view returns(address[])
func (_NodeList *NodeListCaller) GetNodes(opts *bind.CallOpts, epoch *big.Int) ([]common.Address, error) {
	var out []interface{}
	err := _NodeList.contract.Call(opts, &out, "getNodes", epoch)

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetNodes is a free data retrieval call binding the contract method 0x47de074f.
//
// Solidity: function getNodes(uint256 epoch) view returns(address[])
func (_NodeList *NodeListSession) GetNodes(epoch *big.Int) ([]common.Address, error) {
	return _NodeList.Contract.GetNodes(&_NodeList.CallOpts, epoch)
}

// GetNodes is a free data retrieval call binding the contract method 0x47de074f.
//
// Solidity: function getNodes(uint256 epoch) view returns(address[])
func (_NodeList *NodeListCallerSession) GetNodes(epoch *big.Int) ([]common.Address, error) {
	return _NodeList.Contract.GetNodes(&_NodeList.CallOpts, epoch)
}

// GetPssStatus is a free data retrieval call binding the contract method 0xc7aa8ff7.
//
// Solidity: function getPssStatus(uint256 oldEpoch, uint256 newEpoch) view returns(uint256)
func (_NodeList *NodeListCaller) GetPssStatus(opts *bind.CallOpts, oldEpoch *big.Int, newEpoch *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _NodeList.contract.Call(opts, &out, "getPssStatus", oldEpoch, newEpoch)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetPssStatus is a free data retrieval call binding the contract method 0xc7aa8ff7.
//
// Solidity: function getPssStatus(uint256 oldEpoch, uint256 newEpoch) view returns(uint256)
func (_NodeList *NodeListSession) GetPssStatus(oldEpoch *big.Int, newEpoch *big.Int) (*big.Int, error) {
	return _NodeList.Contract.GetPssStatus(&_NodeList.CallOpts, oldEpoch, newEpoch)
}

// GetPssStatus is a free data retrieval call binding the contract method 0xc7aa8ff7.
//
// Solidity: function getPssStatus(uint256 oldEpoch, uint256 newEpoch) view returns(uint256)
func (_NodeList *NodeListCallerSession) GetPssStatus(oldEpoch *big.Int, newEpoch *big.Int) (*big.Int, error) {
	return _NodeList.Contract.GetPssStatus(&_NodeList.CallOpts, oldEpoch, newEpoch)
}

// IsWhitelisted is a free data retrieval call binding the contract method 0x7d22c35c.
//
// Solidity: function isWhitelisted(uint256 epoch, address nodeAddress) view returns(bool)
func (_NodeList *NodeListCaller) IsWhitelisted(opts *bind.CallOpts, epoch *big.Int, nodeAddress common.Address) (bool, error) {
	var out []interface{}
	err := _NodeList.contract.Call(opts, &out, "isWhitelisted", epoch, nodeAddress)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// TODO this has to be actually added to the contract and generate a binding
// this is just for illustration and so the code compiles
func (_NodeList *NodeListCaller) CheckPss(opts *bind.CallOpts, epoch *big.Int) (bool, error) {
	return false, nil
}


// IsWhitelisted is a free data retrieval call binding the contract method 0x7d22c35c.
//
// Solidity: function isWhitelisted(uint256 epoch, address nodeAddress) view returns(bool)
func (_NodeList *NodeListSession) IsWhitelisted(epoch *big.Int, nodeAddress common.Address) (bool, error) {
	return _NodeList.Contract.IsWhitelisted(&_NodeList.CallOpts, epoch, nodeAddress)
}

// IsWhitelisted is a free data retrieval call binding the contract method 0x7d22c35c.
//
// Solidity: function isWhitelisted(uint256 epoch, address nodeAddress) view returns(bool)
func (_NodeList *NodeListCallerSession) IsWhitelisted(epoch *big.Int, nodeAddress common.Address) (bool, error) {
	return _NodeList.Contract.IsWhitelisted(&_NodeList.CallOpts, epoch, nodeAddress)
}

// NodeDetails is a free data retrieval call binding the contract method 0x859da85f.
//
// Solidity: function nodeDetails(address ) view returns(string declaredIp, uint256 position, uint256 pubKx, uint256 pubKy, string tmP2PListenAddress, string p2pListenAddress)
func (_NodeList *NodeListCaller) NodeDetails(opts *bind.CallOpts, arg0 common.Address) (struct {
	DeclaredIp         string
	Position           *big.Int
	PubKx              *big.Int
	PubKy              *big.Int
	TmP2PListenAddress string
	P2pListenAddress   string
}, error) {
	var out []interface{}
	err := _NodeList.contract.Call(opts, &out, "nodeDetails", arg0)

	outstruct := new(struct {
		DeclaredIp         string
		Position           *big.Int
		PubKx              *big.Int
		PubKy              *big.Int
		TmP2PListenAddress string
		P2pListenAddress   string
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.DeclaredIp = *abi.ConvertType(out[0], new(string)).(*string)
	outstruct.Position = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.PubKx = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.PubKy = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.TmP2PListenAddress = *abi.ConvertType(out[4], new(string)).(*string)
	outstruct.P2pListenAddress = *abi.ConvertType(out[5], new(string)).(*string)

	return *outstruct, err

}

// NodeDetails is a free data retrieval call binding the contract method 0x859da85f.
//
// Solidity: function nodeDetails(address ) view returns(string declaredIp, uint256 position, uint256 pubKx, uint256 pubKy, string tmP2PListenAddress, string p2pListenAddress)
func (_NodeList *NodeListSession) NodeDetails(arg0 common.Address) (struct {
	DeclaredIp         string
	Position           *big.Int
	PubKx              *big.Int
	PubKy              *big.Int
	TmP2PListenAddress string
	P2pListenAddress   string
}, error) {
	return _NodeList.Contract.NodeDetails(&_NodeList.CallOpts, arg0)
}

// NodeDetails is a free data retrieval call binding the contract method 0x859da85f.
//
// Solidity: function nodeDetails(address ) view returns(string declaredIp, uint256 position, uint256 pubKx, uint256 pubKy, string tmP2PListenAddress, string p2pListenAddress)
func (_NodeList *NodeListCallerSession) NodeDetails(arg0 common.Address) (struct {
	DeclaredIp         string
	Position           *big.Int
	PubKx              *big.Int
	PubKy              *big.Int
	TmP2PListenAddress string
	P2pListenAddress   string
}, error) {
	return _NodeList.Contract.NodeDetails(&_NodeList.CallOpts, arg0)
}

// NodeRegistered is a free data retrieval call binding the contract method 0x86470e9e.
//
// Solidity: function nodeRegistered(uint256 epoch, address nodeAddress) view returns(bool)
func (_NodeList *NodeListCaller) NodeRegistered(opts *bind.CallOpts, epoch *big.Int, nodeAddress common.Address) (bool, error) {
	var out []interface{}
	err := _NodeList.contract.Call(opts, &out, "nodeRegistered", epoch, nodeAddress)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// NodeRegistered is a free data retrieval call binding the contract method 0x86470e9e.
//
// Solidity: function nodeRegistered(uint256 epoch, address nodeAddress) view returns(bool)
func (_NodeList *NodeListSession) NodeRegistered(epoch *big.Int, nodeAddress common.Address) (bool, error) {
	return _NodeList.Contract.NodeRegistered(&_NodeList.CallOpts, epoch, nodeAddress)
}

// NodeRegistered is a free data retrieval call binding the contract method 0x86470e9e.
//
// Solidity: function nodeRegistered(uint256 epoch, address nodeAddress) view returns(bool)
func (_NodeList *NodeListCallerSession) NodeRegistered(epoch *big.Int, nodeAddress common.Address) (bool, error) {
	return _NodeList.Contract.NodeRegistered(&_NodeList.CallOpts, epoch, nodeAddress)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_NodeList *NodeListCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _NodeList.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_NodeList *NodeListSession) Owner() (common.Address, error) {
	return _NodeList.Contract.Owner(&_NodeList.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_NodeList *NodeListCallerSession) Owner() (common.Address, error) {
	return _NodeList.Contract.Owner(&_NodeList.CallOpts)
}

// PssStatus is a free data retrieval call binding the contract method 0x52fc47b4.
//
// Solidity: function pssStatus(uint256 , uint256 ) view returns(uint256)
func (_NodeList *NodeListCaller) PssStatus(opts *bind.CallOpts, arg0 *big.Int, arg1 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _NodeList.contract.Call(opts, &out, "pssStatus", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// PssStatus is a free data retrieval call binding the contract method 0x52fc47b4.
//
// Solidity: function pssStatus(uint256 , uint256 ) view returns(uint256)
func (_NodeList *NodeListSession) PssStatus(arg0 *big.Int, arg1 *big.Int) (*big.Int, error) {
	return _NodeList.Contract.PssStatus(&_NodeList.CallOpts, arg0, arg1)
}

// PssStatus is a free data retrieval call binding the contract method 0x52fc47b4.
//
// Solidity: function pssStatus(uint256 , uint256 ) view returns(uint256)
func (_NodeList *NodeListCallerSession) PssStatus(arg0 *big.Int, arg1 *big.Int) (*big.Int, error) {
	return _NodeList.Contract.PssStatus(&_NodeList.CallOpts, arg0, arg1)
}

// Whitelist is a free data retrieval call binding the contract method 0x4b25bfce.
//
// Solidity: function whitelist(uint256 , address ) view returns(bool)
func (_NodeList *NodeListCaller) Whitelist(opts *bind.CallOpts, arg0 *big.Int, arg1 common.Address) (bool, error) {
	var out []interface{}
	err := _NodeList.contract.Call(opts, &out, "whitelist", arg0, arg1)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Whitelist is a free data retrieval call binding the contract method 0x4b25bfce.
//
// Solidity: function whitelist(uint256 , address ) view returns(bool)
func (_NodeList *NodeListSession) Whitelist(arg0 *big.Int, arg1 common.Address) (bool, error) {
	return _NodeList.Contract.Whitelist(&_NodeList.CallOpts, arg0, arg1)
}

// Whitelist is a free data retrieval call binding the contract method 0x4b25bfce.
//
// Solidity: function whitelist(uint256 , address ) view returns(bool)
func (_NodeList *NodeListCallerSession) Whitelist(arg0 *big.Int, arg1 common.Address) (bool, error) {
	return _NodeList.Contract.Whitelist(&_NodeList.CallOpts, arg0, arg1)
}

// ClearAllEpoch is a paid mutator transaction binding the contract method 0xfad15163.
//
// Solidity: function clearAllEpoch() returns()
func (_NodeList *NodeListTransactor) ClearAllEpoch(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _NodeList.contract.Transact(opts, "clearAllEpoch")
}

// ClearAllEpoch is a paid mutator transaction binding the contract method 0xfad15163.
//
// Solidity: function clearAllEpoch() returns()
func (_NodeList *NodeListSession) ClearAllEpoch() (*types.Transaction, error) {
	return _NodeList.Contract.ClearAllEpoch(&_NodeList.TransactOpts)
}

// ClearAllEpoch is a paid mutator transaction binding the contract method 0xfad15163.
//
// Solidity: function clearAllEpoch() returns()
func (_NodeList *NodeListTransactorSession) ClearAllEpoch() (*types.Transaction, error) {
	return _NodeList.Contract.ClearAllEpoch(&_NodeList.TransactOpts)
}

// Initialize is a paid mutator transaction binding the contract method 0xfe4b84df.
//
// Solidity: function initialize(uint256 epoch) returns()
func (_NodeList *NodeListTransactor) Initialize(opts *bind.TransactOpts, epoch *big.Int) (*types.Transaction, error) {
	return _NodeList.contract.Transact(opts, "initialize", epoch)
}

// Initialize is a paid mutator transaction binding the contract method 0xfe4b84df.
//
// Solidity: function initialize(uint256 epoch) returns()
func (_NodeList *NodeListSession) Initialize(epoch *big.Int) (*types.Transaction, error) {
	return _NodeList.Contract.Initialize(&_NodeList.TransactOpts, epoch)
}

// Initialize is a paid mutator transaction binding the contract method 0xfe4b84df.
//
// Solidity: function initialize(uint256 epoch) returns()
func (_NodeList *NodeListTransactorSession) Initialize(epoch *big.Int) (*types.Transaction, error) {
	return _NodeList.Contract.Initialize(&_NodeList.TransactOpts, epoch)
}

// ListNode is a paid mutator transaction binding the contract method 0xbf2d6f81.
//
// Solidity: function listNode(uint256 epoch, string declaredIp, uint256 pubKx, uint256 pubKy, string tmP2PListenAddress, string p2pListenAddress) returns()
func (_NodeList *NodeListTransactor) ListNode(opts *bind.TransactOpts, epoch *big.Int, declaredIp string, pubKx *big.Int, pubKy *big.Int, tmP2PListenAddress string, p2pListenAddress string) (*types.Transaction, error) {
	return _NodeList.contract.Transact(opts, "listNode", epoch, declaredIp, pubKx, pubKy, tmP2PListenAddress, p2pListenAddress)
}

// ListNode is a paid mutator transaction binding the contract method 0xbf2d6f81.
//
// Solidity: function listNode(uint256 epoch, string declaredIp, uint256 pubKx, uint256 pubKy, string tmP2PListenAddress, string p2pListenAddress) returns()
func (_NodeList *NodeListSession) ListNode(epoch *big.Int, declaredIp string, pubKx *big.Int, pubKy *big.Int, tmP2PListenAddress string, p2pListenAddress string) (*types.Transaction, error) {
	return _NodeList.Contract.ListNode(&_NodeList.TransactOpts, epoch, declaredIp, pubKx, pubKy, tmP2PListenAddress, p2pListenAddress)
}

// ListNode is a paid mutator transaction binding the contract method 0xbf2d6f81.
//
// Solidity: function listNode(uint256 epoch, string declaredIp, uint256 pubKx, uint256 pubKy, string tmP2PListenAddress, string p2pListenAddress) returns()
func (_NodeList *NodeListTransactorSession) ListNode(epoch *big.Int, declaredIp string, pubKx *big.Int, pubKy *big.Int, tmP2PListenAddress string, p2pListenAddress string) (*types.Transaction, error) {
	return _NodeList.Contract.ListNode(&_NodeList.TransactOpts, epoch, declaredIp, pubKx, pubKy, tmP2PListenAddress, p2pListenAddress)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_NodeList *NodeListTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _NodeList.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_NodeList *NodeListSession) RenounceOwnership() (*types.Transaction, error) {
	return _NodeList.Contract.RenounceOwnership(&_NodeList.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_NodeList *NodeListTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _NodeList.Contract.RenounceOwnership(&_NodeList.TransactOpts)
}

// SetBufferSize is a paid mutator transaction binding the contract method 0x6401aa18.
//
// Solidity: function setBufferSize(uint256 size) returns()
func (_NodeList *NodeListTransactor) SetBufferSize(opts *bind.TransactOpts, size *big.Int) (*types.Transaction, error) {
	return _NodeList.contract.Transact(opts, "setBufferSize", size)
}

// SetBufferSize is a paid mutator transaction binding the contract method 0x6401aa18.
//
// Solidity: function setBufferSize(uint256 size) returns()
func (_NodeList *NodeListSession) SetBufferSize(size *big.Int) (*types.Transaction, error) {
	return _NodeList.Contract.SetBufferSize(&_NodeList.TransactOpts, size)
}

// SetBufferSize is a paid mutator transaction binding the contract method 0x6401aa18.
//
// Solidity: function setBufferSize(uint256 size) returns()
func (_NodeList *NodeListTransactorSession) SetBufferSize(size *big.Int) (*types.Transaction, error) {
	return _NodeList.Contract.SetBufferSize(&_NodeList.TransactOpts, size)
}

// SetCurrentEpoch is a paid mutator transaction binding the contract method 0x1dd6b9b1.
//
// Solidity: function setCurrentEpoch(uint256 newEpoch) returns()
func (_NodeList *NodeListTransactor) SetCurrentEpoch(opts *bind.TransactOpts, newEpoch *big.Int) (*types.Transaction, error) {
	return _NodeList.contract.Transact(opts, "setCurrentEpoch", newEpoch)
}

// SetCurrentEpoch is a paid mutator transaction binding the contract method 0x1dd6b9b1.
//
// Solidity: function setCurrentEpoch(uint256 newEpoch) returns()
func (_NodeList *NodeListSession) SetCurrentEpoch(newEpoch *big.Int) (*types.Transaction, error) {
	return _NodeList.Contract.SetCurrentEpoch(&_NodeList.TransactOpts, newEpoch)
}

// SetCurrentEpoch is a paid mutator transaction binding the contract method 0x1dd6b9b1.
//
// Solidity: function setCurrentEpoch(uint256 newEpoch) returns()
func (_NodeList *NodeListTransactorSession) SetCurrentEpoch(newEpoch *big.Int) (*types.Transaction, error) {
	return _NodeList.Contract.SetCurrentEpoch(&_NodeList.TransactOpts, newEpoch)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_NodeList *NodeListTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _NodeList.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_NodeList *NodeListSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _NodeList.Contract.TransferOwnership(&_NodeList.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_NodeList *NodeListTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _NodeList.Contract.TransferOwnership(&_NodeList.TransactOpts, newOwner)
}

// UpdateEpoch is a paid mutator transaction binding the contract method 0xae4df20c.
//
// Solidity: function updateEpoch(uint256 epoch, uint256 n, uint256 k, uint256 t, address[] nodeList, uint256 prevEpoch, uint256 nextEpoch) returns()
func (_NodeList *NodeListTransactor) UpdateEpoch(opts *bind.TransactOpts, epoch *big.Int, n *big.Int, k *big.Int, t *big.Int, nodeList []common.Address, prevEpoch *big.Int, nextEpoch *big.Int) (*types.Transaction, error) {
	return _NodeList.contract.Transact(opts, "updateEpoch", epoch, n, k, t, nodeList, prevEpoch, nextEpoch)
}

// UpdateEpoch is a paid mutator transaction binding the contract method 0xae4df20c.
//
// Solidity: function updateEpoch(uint256 epoch, uint256 n, uint256 k, uint256 t, address[] nodeList, uint256 prevEpoch, uint256 nextEpoch) returns()
func (_NodeList *NodeListSession) UpdateEpoch(epoch *big.Int, n *big.Int, k *big.Int, t *big.Int, nodeList []common.Address, prevEpoch *big.Int, nextEpoch *big.Int) (*types.Transaction, error) {
	return _NodeList.Contract.UpdateEpoch(&_NodeList.TransactOpts, epoch, n, k, t, nodeList, prevEpoch, nextEpoch)
}

// UpdateEpoch is a paid mutator transaction binding the contract method 0xae4df20c.
//
// Solidity: function updateEpoch(uint256 epoch, uint256 n, uint256 k, uint256 t, address[] nodeList, uint256 prevEpoch, uint256 nextEpoch) returns()
func (_NodeList *NodeListTransactorSession) UpdateEpoch(epoch *big.Int, n *big.Int, k *big.Int, t *big.Int, nodeList []common.Address, prevEpoch *big.Int, nextEpoch *big.Int) (*types.Transaction, error) {
	return _NodeList.Contract.UpdateEpoch(&_NodeList.TransactOpts, epoch, n, k, t, nodeList, prevEpoch, nextEpoch)
}

// UpdatePssStatus is a paid mutator transaction binding the contract method 0x6967ac51.
//
// Solidity: function updatePssStatus(uint256 oldEpoch, uint256 newEpoch, uint256 status) returns()
func (_NodeList *NodeListTransactor) UpdatePssStatus(opts *bind.TransactOpts, oldEpoch *big.Int, newEpoch *big.Int, status *big.Int) (*types.Transaction, error) {
	return _NodeList.contract.Transact(opts, "updatePssStatus", oldEpoch, newEpoch, status)
}

// UpdatePssStatus is a paid mutator transaction binding the contract method 0x6967ac51.
//
// Solidity: function updatePssStatus(uint256 oldEpoch, uint256 newEpoch, uint256 status) returns()
func (_NodeList *NodeListSession) UpdatePssStatus(oldEpoch *big.Int, newEpoch *big.Int, status *big.Int) (*types.Transaction, error) {
	return _NodeList.Contract.UpdatePssStatus(&_NodeList.TransactOpts, oldEpoch, newEpoch, status)
}

// UpdatePssStatus is a paid mutator transaction binding the contract method 0x6967ac51.
//
// Solidity: function updatePssStatus(uint256 oldEpoch, uint256 newEpoch, uint256 status) returns()
func (_NodeList *NodeListTransactorSession) UpdatePssStatus(oldEpoch *big.Int, newEpoch *big.Int, status *big.Int) (*types.Transaction, error) {
	return _NodeList.Contract.UpdatePssStatus(&_NodeList.TransactOpts, oldEpoch, newEpoch, status)
}

// UpdateWhitelist is a paid mutator transaction binding the contract method 0x3d4602a9.
//
// Solidity: function updateWhitelist(uint256 epoch, address nodeAddress, bool allowed) returns()
func (_NodeList *NodeListTransactor) UpdateWhitelist(opts *bind.TransactOpts, epoch *big.Int, nodeAddress common.Address, allowed bool) (*types.Transaction, error) {
	return _NodeList.contract.Transact(opts, "updateWhitelist", epoch, nodeAddress, allowed)
}

// UpdateWhitelist is a paid mutator transaction binding the contract method 0x3d4602a9.
//
// Solidity: function updateWhitelist(uint256 epoch, address nodeAddress, bool allowed) returns()
func (_NodeList *NodeListSession) UpdateWhitelist(epoch *big.Int, nodeAddress common.Address, allowed bool) (*types.Transaction, error) {
	return _NodeList.Contract.UpdateWhitelist(&_NodeList.TransactOpts, epoch, nodeAddress, allowed)
}

// UpdateWhitelist is a paid mutator transaction binding the contract method 0x3d4602a9.
//
// Solidity: function updateWhitelist(uint256 epoch, address nodeAddress, bool allowed) returns()
func (_NodeList *NodeListTransactorSession) UpdateWhitelist(epoch *big.Int, nodeAddress common.Address, allowed bool) (*types.Transaction, error) {
	return _NodeList.Contract.UpdateWhitelist(&_NodeList.TransactOpts, epoch, nodeAddress, allowed)
}

// NodeListEpochChangedIterator is returned from FilterEpochChanged and is used to iterate over the raw logs and unpacked data for EpochChanged events raised by the NodeList contract.
type NodeListEpochChangedIterator struct {
	Event *NodeListEpochChanged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *NodeListEpochChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(NodeListEpochChanged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(NodeListEpochChanged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *NodeListEpochChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *NodeListEpochChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// NodeListEpochChanged represents a EpochChanged event raised by the NodeList contract.
type NodeListEpochChanged struct {
	OldEpoch *big.Int
	NewEpoch *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterEpochChanged is a free log retrieval operation binding the contract event 0x528990bbb5369a7f6d5acab41233e32bddb4882673d0208805b59cbad0dc1ec8.
//
// Solidity: event EpochChanged(uint256 oldEpoch, uint256 newEpoch)
func (_NodeList *NodeListFilterer) FilterEpochChanged(opts *bind.FilterOpts) (*NodeListEpochChangedIterator, error) {

	logs, sub, err := _NodeList.contract.FilterLogs(opts, "EpochChanged")
	if err != nil {
		return nil, err
	}
	return &NodeListEpochChangedIterator{contract: _NodeList.contract, event: "EpochChanged", logs: logs, sub: sub}, nil
}

// WatchEpochChanged is a free log subscription operation binding the contract event 0x528990bbb5369a7f6d5acab41233e32bddb4882673d0208805b59cbad0dc1ec8.
//
// Solidity: event EpochChanged(uint256 oldEpoch, uint256 newEpoch)
func (_NodeList *NodeListFilterer) WatchEpochChanged(opts *bind.WatchOpts, sink chan<- *NodeListEpochChanged) (event.Subscription, error) {

	logs, sub, err := _NodeList.contract.WatchLogs(opts, "EpochChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(NodeListEpochChanged)
				if err := _NodeList.contract.UnpackLog(event, "EpochChanged", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseEpochChanged is a log parse operation binding the contract event 0x528990bbb5369a7f6d5acab41233e32bddb4882673d0208805b59cbad0dc1ec8.
//
// Solidity: event EpochChanged(uint256 oldEpoch, uint256 newEpoch)
func (_NodeList *NodeListFilterer) ParseEpochChanged(log types.Log) (*NodeListEpochChanged, error) {
	event := new(NodeListEpochChanged)
	if err := _NodeList.contract.UnpackLog(event, "EpochChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// NodeListEpochClearedIterator is returned from FilterEpochCleared and is used to iterate over the raw logs and unpacked data for EpochCleared events raised by the NodeList contract.
type NodeListEpochClearedIterator struct {
	Event *NodeListEpochCleared // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *NodeListEpochClearedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(NodeListEpochCleared)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(NodeListEpochCleared)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *NodeListEpochClearedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *NodeListEpochClearedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// NodeListEpochCleared represents a EpochCleared event raised by the NodeList contract.
type NodeListEpochCleared struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterEpochCleared is a free log retrieval operation binding the contract event 0x4edf392be08371c6e69fdf4ee9e6ac02191e17511208bbe30c0ef3b841f5c57f.
//
// Solidity: event EpochCleared()
func (_NodeList *NodeListFilterer) FilterEpochCleared(opts *bind.FilterOpts) (*NodeListEpochClearedIterator, error) {

	logs, sub, err := _NodeList.contract.FilterLogs(opts, "EpochCleared")
	if err != nil {
		return nil, err
	}
	return &NodeListEpochClearedIterator{contract: _NodeList.contract, event: "EpochCleared", logs: logs, sub: sub}, nil
}

// WatchEpochCleared is a free log subscription operation binding the contract event 0x4edf392be08371c6e69fdf4ee9e6ac02191e17511208bbe30c0ef3b841f5c57f.
//
// Solidity: event EpochCleared()
func (_NodeList *NodeListFilterer) WatchEpochCleared(opts *bind.WatchOpts, sink chan<- *NodeListEpochCleared) (event.Subscription, error) {

	logs, sub, err := _NodeList.contract.WatchLogs(opts, "EpochCleared")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(NodeListEpochCleared)
				if err := _NodeList.contract.UnpackLog(event, "EpochCleared", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseEpochCleared is a log parse operation binding the contract event 0x4edf392be08371c6e69fdf4ee9e6ac02191e17511208bbe30c0ef3b841f5c57f.
//
// Solidity: event EpochCleared()
func (_NodeList *NodeListFilterer) ParseEpochCleared(log types.Log) (*NodeListEpochCleared, error) {
	event := new(NodeListEpochCleared)
	if err := _NodeList.contract.UnpackLog(event, "EpochCleared", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// NodeListEpochUpdateIterator is returned from FilterEpochUpdate and is used to iterate over the raw logs and unpacked data for EpochUpdate events raised by the NodeList contract.
type NodeListEpochUpdateIterator struct {
	Event *NodeListEpochUpdate // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *NodeListEpochUpdateIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(NodeListEpochUpdate)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(NodeListEpochUpdate)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *NodeListEpochUpdateIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *NodeListEpochUpdateIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// NodeListEpochUpdate represents a EpochUpdate event raised by the NodeList contract.
type NodeListEpochUpdate struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterEpochUpdate is a free log retrieval operation binding the contract event 0x6607321c8a6d680eb30bf4ed3cf9f6d8263c1296b61c4e4230ab4a5d7f38cad4.
//
// Solidity: event EpochUpdate()
func (_NodeList *NodeListFilterer) FilterEpochUpdate(opts *bind.FilterOpts) (*NodeListEpochUpdateIterator, error) {

	logs, sub, err := _NodeList.contract.FilterLogs(opts, "EpochUpdate")
	if err != nil {
		return nil, err
	}
	return &NodeListEpochUpdateIterator{contract: _NodeList.contract, event: "EpochUpdate", logs: logs, sub: sub}, nil
}

// WatchEpochUpdate is a free log subscription operation binding the contract event 0x6607321c8a6d680eb30bf4ed3cf9f6d8263c1296b61c4e4230ab4a5d7f38cad4.
//
// Solidity: event EpochUpdate()
func (_NodeList *NodeListFilterer) WatchEpochUpdate(opts *bind.WatchOpts, sink chan<- *NodeListEpochUpdate) (event.Subscription, error) {

	logs, sub, err := _NodeList.contract.WatchLogs(opts, "EpochUpdate")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(NodeListEpochUpdate)
				if err := _NodeList.contract.UnpackLog(event, "EpochUpdate", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseEpochUpdate is a log parse operation binding the contract event 0x6607321c8a6d680eb30bf4ed3cf9f6d8263c1296b61c4e4230ab4a5d7f38cad4.
//
// Solidity: event EpochUpdate()
func (_NodeList *NodeListFilterer) ParseEpochUpdate(log types.Log) (*NodeListEpochUpdate, error) {
	event := new(NodeListEpochUpdate)
	if err := _NodeList.contract.UnpackLog(event, "EpochUpdate", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// NodeListInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the NodeList contract.
type NodeListInitializedIterator struct {
	Event *NodeListInitialized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *NodeListInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(NodeListInitialized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(NodeListInitialized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *NodeListInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *NodeListInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// NodeListInitialized represents a Initialized event raised by the NodeList contract.
type NodeListInitialized struct {
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_NodeList *NodeListFilterer) FilterInitialized(opts *bind.FilterOpts) (*NodeListInitializedIterator, error) {

	logs, sub, err := _NodeList.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &NodeListInitializedIterator{contract: _NodeList.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_NodeList *NodeListFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *NodeListInitialized) (event.Subscription, error) {

	logs, sub, err := _NodeList.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(NodeListInitialized)
				if err := _NodeList.contract.UnpackLog(event, "Initialized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseInitialized is a log parse operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_NodeList *NodeListFilterer) ParseInitialized(log types.Log) (*NodeListInitialized, error) {
	event := new(NodeListInitialized)
	if err := _NodeList.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// NodeListNodeListedIterator is returned from FilterNodeListed and is used to iterate over the raw logs and unpacked data for NodeListed events raised by the NodeList contract.
type NodeListNodeListedIterator struct {
	Event *NodeListNodeListed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *NodeListNodeListedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(NodeListNodeListed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(NodeListNodeListed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *NodeListNodeListedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *NodeListNodeListedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// NodeListNodeListed represents a NodeListed event raised by the NodeList contract.
type NodeListNodeListed struct {
	PublicKey common.Address
	Epoch     *big.Int
	Position  *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterNodeListed is a free log retrieval operation binding the contract event 0xe2f8adb0f494dc82ccf446c031763ef3762d6396d51664611ed89aac0117339e.
//
// Solidity: event NodeListed(address publicKey, uint256 epoch, uint256 position)
func (_NodeList *NodeListFilterer) FilterNodeListed(opts *bind.FilterOpts) (*NodeListNodeListedIterator, error) {

	logs, sub, err := _NodeList.contract.FilterLogs(opts, "NodeListed")
	if err != nil {
		return nil, err
	}
	return &NodeListNodeListedIterator{contract: _NodeList.contract, event: "NodeListed", logs: logs, sub: sub}, nil
}

// WatchNodeListed is a free log subscription operation binding the contract event 0xe2f8adb0f494dc82ccf446c031763ef3762d6396d51664611ed89aac0117339e.
//
// Solidity: event NodeListed(address publicKey, uint256 epoch, uint256 position)
func (_NodeList *NodeListFilterer) WatchNodeListed(opts *bind.WatchOpts, sink chan<- *NodeListNodeListed) (event.Subscription, error) {

	logs, sub, err := _NodeList.contract.WatchLogs(opts, "NodeListed")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(NodeListNodeListed)
				if err := _NodeList.contract.UnpackLog(event, "NodeListed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseNodeListed is a log parse operation binding the contract event 0xe2f8adb0f494dc82ccf446c031763ef3762d6396d51664611ed89aac0117339e.
//
// Solidity: event NodeListed(address publicKey, uint256 epoch, uint256 position)
func (_NodeList *NodeListFilterer) ParseNodeListed(log types.Log) (*NodeListNodeListed, error) {
	event := new(NodeListNodeListed)
	if err := _NodeList.contract.UnpackLog(event, "NodeListed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// NodeListOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the NodeList contract.
type NodeListOwnershipTransferredIterator struct {
	Event *NodeListOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *NodeListOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(NodeListOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(NodeListOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *NodeListOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *NodeListOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// NodeListOwnershipTransferred represents a OwnershipTransferred event raised by the NodeList contract.
type NodeListOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_NodeList *NodeListFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*NodeListOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _NodeList.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &NodeListOwnershipTransferredIterator{contract: _NodeList.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_NodeList *NodeListFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *NodeListOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _NodeList.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(NodeListOwnershipTransferred)
				if err := _NodeList.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_NodeList *NodeListFilterer) ParseOwnershipTransferred(log types.Log) (*NodeListOwnershipTransferred, error) {
	event := new(NodeListOwnershipTransferred)
	if err := _NodeList.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// NodeListPssStatusUpdateIterator is returned from FilterPssStatusUpdate and is used to iterate over the raw logs and unpacked data for PssStatusUpdate events raised by the NodeList contract.
type NodeListPssStatusUpdateIterator struct {
	Event *NodeListPssStatusUpdate // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *NodeListPssStatusUpdateIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(NodeListPssStatusUpdate)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(NodeListPssStatusUpdate)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *NodeListPssStatusUpdateIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *NodeListPssStatusUpdateIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// NodeListPssStatusUpdate represents a PssStatusUpdate event raised by the NodeList contract.
type NodeListPssStatusUpdate struct {
	Epoch  *big.Int
	Status *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterPssStatusUpdate is a free log retrieval operation binding the contract event 0x5e17f1d0525b9fecff300306ae6ae54583fadf25d127a4ded5e242a79453cbb8.
//
// Solidity: event PssStatusUpdate(uint256 indexed epoch, uint256 status)
func (_NodeList *NodeListFilterer) FilterPssStatusUpdate(opts *bind.FilterOpts, epoch []*big.Int) (*NodeListPssStatusUpdateIterator, error) {

	var epochRule []interface{}
	for _, epochItem := range epoch {
		epochRule = append(epochRule, epochItem)
	}

	logs, sub, err := _NodeList.contract.FilterLogs(opts, "PssStatusUpdate", epochRule)
	if err != nil {
		return nil, err
	}
	return &NodeListPssStatusUpdateIterator{contract: _NodeList.contract, event: "PssStatusUpdate", logs: logs, sub: sub}, nil
}

// WatchPssStatusUpdate is a free log subscription operation binding the contract event 0x5e17f1d0525b9fecff300306ae6ae54583fadf25d127a4ded5e242a79453cbb8.
//
// Solidity: event PssStatusUpdate(uint256 indexed epoch, uint256 status)
func (_NodeList *NodeListFilterer) WatchPssStatusUpdate(opts *bind.WatchOpts, sink chan<- *NodeListPssStatusUpdate, epoch []*big.Int) (event.Subscription, error) {

	var epochRule []interface{}
	for _, epochItem := range epoch {
		epochRule = append(epochRule, epochItem)
	}

	logs, sub, err := _NodeList.contract.WatchLogs(opts, "PssStatusUpdate", epochRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(NodeListPssStatusUpdate)
				if err := _NodeList.contract.UnpackLog(event, "PssStatusUpdate", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParsePssStatusUpdate is a log parse operation binding the contract event 0x5e17f1d0525b9fecff300306ae6ae54583fadf25d127a4ded5e242a79453cbb8.
//
// Solidity: event PssStatusUpdate(uint256 indexed epoch, uint256 status)
func (_NodeList *NodeListFilterer) ParsePssStatusUpdate(log types.Log) (*NodeListPssStatusUpdate, error) {
	event := new(NodeListPssStatusUpdate)
	if err := _NodeList.contract.UnpackLog(event, "PssStatusUpdate", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// NodeListWhitelistUpdateIterator is returned from FilterWhitelistUpdate and is used to iterate over the raw logs and unpacked data for WhitelistUpdate events raised by the NodeList contract.
type NodeListWhitelistUpdateIterator struct {
	Event *NodeListWhitelistUpdate // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *NodeListWhitelistUpdateIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(NodeListWhitelistUpdate)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(NodeListWhitelistUpdate)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *NodeListWhitelistUpdateIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *NodeListWhitelistUpdateIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// NodeListWhitelistUpdate represents a WhitelistUpdate event raised by the NodeList contract.
type NodeListWhitelistUpdate struct {
	Epoch            *big.Int
	WhitelistAddress common.Address
	IsAllowed        bool
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterWhitelistUpdate is a free log retrieval operation binding the contract event 0x426d5b4b9e01cc411bbf5246328bdcaaa9d416f6059b19526fa90c549056e44c.
//
// Solidity: event WhitelistUpdate(uint256 indexed epoch, address indexed whitelistAddress, bool isAllowed)
func (_NodeList *NodeListFilterer) FilterWhitelistUpdate(opts *bind.FilterOpts, epoch []*big.Int, whitelistAddress []common.Address) (*NodeListWhitelistUpdateIterator, error) {

	var epochRule []interface{}
	for _, epochItem := range epoch {
		epochRule = append(epochRule, epochItem)
	}
	var whitelistAddressRule []interface{}
	for _, whitelistAddressItem := range whitelistAddress {
		whitelistAddressRule = append(whitelistAddressRule, whitelistAddressItem)
	}

	logs, sub, err := _NodeList.contract.FilterLogs(opts, "WhitelistUpdate", epochRule, whitelistAddressRule)
	if err != nil {
		return nil, err
	}
	return &NodeListWhitelistUpdateIterator{contract: _NodeList.contract, event: "WhitelistUpdate", logs: logs, sub: sub}, nil
}

// WatchWhitelistUpdate is a free log subscription operation binding the contract event 0x426d5b4b9e01cc411bbf5246328bdcaaa9d416f6059b19526fa90c549056e44c.
//
// Solidity: event WhitelistUpdate(uint256 indexed epoch, address indexed whitelistAddress, bool isAllowed)
func (_NodeList *NodeListFilterer) WatchWhitelistUpdate(opts *bind.WatchOpts, sink chan<- *NodeListWhitelistUpdate, epoch []*big.Int, whitelistAddress []common.Address) (event.Subscription, error) {

	var epochRule []interface{}
	for _, epochItem := range epoch {
		epochRule = append(epochRule, epochItem)
	}
	var whitelistAddressRule []interface{}
	for _, whitelistAddressItem := range whitelistAddress {
		whitelistAddressRule = append(whitelistAddressRule, whitelistAddressItem)
	}

	logs, sub, err := _NodeList.contract.WatchLogs(opts, "WhitelistUpdate", epochRule, whitelistAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(NodeListWhitelistUpdate)
				if err := _NodeList.contract.UnpackLog(event, "WhitelistUpdate", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseWhitelistUpdate is a log parse operation binding the contract event 0x426d5b4b9e01cc411bbf5246328bdcaaa9d416f6059b19526fa90c549056e44c.
//
// Solidity: event WhitelistUpdate(uint256 indexed epoch, address indexed whitelistAddress, bool isAllowed)
func (_NodeList *NodeListFilterer) ParseWhitelistUpdate(log types.Log) (*NodeListWhitelistUpdate, error) {
	event := new(NodeListWhitelistUpdate)
	if err := _NodeList.contract.UnpackLog(event, "WhitelistUpdate", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
