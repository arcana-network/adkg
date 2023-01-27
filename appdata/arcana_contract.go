// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package appdata

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

// ArcanaMetaData contains all meta data concerning the Arcana contract.
var ArcanaMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"DeleteApp\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"did\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"}],\"name\":\"DownloadViaRuleSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"did\",\"type\":\"bytes32\"}],\"name\":\"addFile\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"aggregateLogin\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"appFiles\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"userVersion\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"appLevelControl\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"did\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"changeFileOwner\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"consumption\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"store\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"bandwidth\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"defaultLimit\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"store\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"bandwidth\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"delegators\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"deleteApp\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"did\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"ephemeralWallet\",\"type\":\"address\"}],\"name\":\"download\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"did\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"txHash\",\"type\":\"bytes32\"}],\"name\":\"downloadClose\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"appPermission\",\"type\":\"uint8\"},{\"internalType\":\"bool\",\"name\":\"add\",\"type\":\"bool\"}],\"name\":\"editAppPermission\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getAppConfig\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"grantAppPermission\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"factoryAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"relayer\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"aggregateLoginValue\",\"type\":\"bool\"},{\"internalType\":\"address\",\"name\":\"did\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"appConfigValue\",\"type\":\"bytes32\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"isActive\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"forwarder\",\"type\":\"address\"}],\"name\":\"isTrustedForwarder\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"limit\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"store\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"bandwidth\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"did\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"nftContract\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"chainId\",\"type\":\"uint256\"}],\"name\":\"linkNFT\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"did\",\"type\":\"bytes32\"}],\"name\":\"removeUserFile\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"revokeApp\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"appConfig\",\"type\":\"bytes32\"}],\"name\":\"setAppConfig\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"store\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"bandwidth\",\"type\":\"uint256\"}],\"name\":\"setAppLimit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"store\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"bandwidth\",\"type\":\"uint256\"}],\"name\":\"setDefaultLimit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bool\",\"name\":\"status\",\"type\":\"bool\"}],\"name\":\"setUnPartitioned\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"store\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"bandwidth\",\"type\":\"uint256\"}],\"name\":\"setUserLevelLimit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"toggleWalletType\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"txCounter\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"unpartitioned\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"internalType\":\"uint8\",\"name\":\"control\",\"type\":\"uint8\"},{\"internalType\":\"bool\",\"name\":\"add\",\"type\":\"bool\"}],\"name\":\"updateDelegator\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"did\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"ruleHash\",\"type\":\"bytes32\"}],\"name\":\"updateRuleSet\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"did\",\"type\":\"bytes32\"}],\"name\":\"uploadClose\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"did\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"fileSize\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"name\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"fileHash\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"storageNode\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"ephemeralAddress\",\"type\":\"address\"}],\"name\":\"uploadInit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"userAppPermission\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"userVersion\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"walletType\",\"outputs\":[{\"internalType\":\"enumArcana.WalletMode\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// ArcanaABI is the input ABI used to generate the binding from.
// Deprecated: Use ArcanaMetaData.ABI instead.
var ArcanaABI = ArcanaMetaData.ABI

// Arcana is an auto generated Go binding around an Ethereum contract.
type Arcana struct {
	ArcanaCaller     // Read-only binding to the contract
	ArcanaTransactor // Write-only binding to the contract
	ArcanaFilterer   // Log filterer for contract events
}

// ArcanaCaller is an auto generated read-only Go binding around an Ethereum contract.
type ArcanaCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ArcanaTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ArcanaTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ArcanaFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ArcanaFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ArcanaSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ArcanaSession struct {
	Contract     *Arcana           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ArcanaCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ArcanaCallerSession struct {
	Contract *ArcanaCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// ArcanaTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ArcanaTransactorSession struct {
	Contract     *ArcanaTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ArcanaRaw is an auto generated low-level Go binding around an Ethereum contract.
type ArcanaRaw struct {
	Contract *Arcana // Generic contract binding to access the raw methods on
}

// ArcanaCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ArcanaCallerRaw struct {
	Contract *ArcanaCaller // Generic read-only contract binding to access the raw methods on
}

// ArcanaTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ArcanaTransactorRaw struct {
	Contract *ArcanaTransactor // Generic write-only contract binding to access the raw methods on
}

// NewArcana creates a new instance of Arcana, bound to a specific deployed contract.
func NewArcana(address common.Address, backend bind.ContractBackend) (*Arcana, error) {
	contract, err := bindArcana(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Arcana{ArcanaCaller: ArcanaCaller{contract: contract}, ArcanaTransactor: ArcanaTransactor{contract: contract}, ArcanaFilterer: ArcanaFilterer{contract: contract}}, nil
}

// NewArcanaCaller creates a new read-only instance of Arcana, bound to a specific deployed contract.
func NewArcanaCaller(address common.Address, caller bind.ContractCaller) (*ArcanaCaller, error) {
	contract, err := bindArcana(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ArcanaCaller{contract: contract}, nil
}

// NewArcanaTransactor creates a new write-only instance of Arcana, bound to a specific deployed contract.
func NewArcanaTransactor(address common.Address, transactor bind.ContractTransactor) (*ArcanaTransactor, error) {
	contract, err := bindArcana(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ArcanaTransactor{contract: contract}, nil
}

// NewArcanaFilterer creates a new log filterer instance of Arcana, bound to a specific deployed contract.
func NewArcanaFilterer(address common.Address, filterer bind.ContractFilterer) (*ArcanaFilterer, error) {
	contract, err := bindArcana(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ArcanaFilterer{contract: contract}, nil
}

// bindArcana binds a generic wrapper to an already deployed contract.
func bindArcana(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ArcanaABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Arcana *ArcanaRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Arcana.Contract.ArcanaCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Arcana *ArcanaRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Arcana.Contract.ArcanaTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Arcana *ArcanaRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Arcana.Contract.ArcanaTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Arcana *ArcanaCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Arcana.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Arcana *ArcanaTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Arcana.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Arcana *ArcanaTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Arcana.Contract.contract.Transact(opts, method, params...)
}

// AggregateLogin is a free data retrieval call binding the contract method 0xc53be4c8.
//
// Solidity: function aggregateLogin() view returns(bool)
func (_Arcana *ArcanaCaller) AggregateLogin(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _Arcana.contract.Call(opts, &out, "aggregateLogin")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// AggregateLogin is a free data retrieval call binding the contract method 0xc53be4c8.
//
// Solidity: function aggregateLogin() view returns(bool)
func (_Arcana *ArcanaSession) AggregateLogin() (bool, error) {
	return _Arcana.Contract.AggregateLogin(&_Arcana.CallOpts)
}

// AggregateLogin is a free data retrieval call binding the contract method 0xc53be4c8.
//
// Solidity: function aggregateLogin() view returns(bool)
func (_Arcana *ArcanaCallerSession) AggregateLogin() (bool, error) {
	return _Arcana.Contract.AggregateLogin(&_Arcana.CallOpts)
}

// AppFiles is a free data retrieval call binding the contract method 0xb72f72b6.
//
// Solidity: function appFiles(bytes32 ) view returns(address owner, uint256 userVersion)
func (_Arcana *ArcanaCaller) AppFiles(opts *bind.CallOpts, arg0 [32]byte) (struct {
	Owner       common.Address
	UserVersion *big.Int
}, error) {
	var out []interface{}
	err := _Arcana.contract.Call(opts, &out, "appFiles", arg0)

	outstruct := new(struct {
		Owner       common.Address
		UserVersion *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Owner = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.UserVersion = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// AppFiles is a free data retrieval call binding the contract method 0xb72f72b6.
//
// Solidity: function appFiles(bytes32 ) view returns(address owner, uint256 userVersion)
func (_Arcana *ArcanaSession) AppFiles(arg0 [32]byte) (struct {
	Owner       common.Address
	UserVersion *big.Int
}, error) {
	return _Arcana.Contract.AppFiles(&_Arcana.CallOpts, arg0)
}

// AppFiles is a free data retrieval call binding the contract method 0xb72f72b6.
//
// Solidity: function appFiles(bytes32 ) view returns(address owner, uint256 userVersion)
func (_Arcana *ArcanaCallerSession) AppFiles(arg0 [32]byte) (struct {
	Owner       common.Address
	UserVersion *big.Int
}, error) {
	return _Arcana.Contract.AppFiles(&_Arcana.CallOpts, arg0)
}

// AppLevelControl is a free data retrieval call binding the contract method 0x5cdb2592.
//
// Solidity: function appLevelControl() view returns(uint8)
func (_Arcana *ArcanaCaller) AppLevelControl(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _Arcana.contract.Call(opts, &out, "appLevelControl")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// AppLevelControl is a free data retrieval call binding the contract method 0x5cdb2592.
//
// Solidity: function appLevelControl() view returns(uint8)
func (_Arcana *ArcanaSession) AppLevelControl() (uint8, error) {
	return _Arcana.Contract.AppLevelControl(&_Arcana.CallOpts)
}

// AppLevelControl is a free data retrieval call binding the contract method 0x5cdb2592.
//
// Solidity: function appLevelControl() view returns(uint8)
func (_Arcana *ArcanaCallerSession) AppLevelControl() (uint8, error) {
	return _Arcana.Contract.AppLevelControl(&_Arcana.CallOpts)
}

// Consumption is a free data retrieval call binding the contract method 0x633f7155.
//
// Solidity: function consumption(address ) view returns(uint256 store, uint256 bandwidth)
func (_Arcana *ArcanaCaller) Consumption(opts *bind.CallOpts, arg0 common.Address) (struct {
	Store     *big.Int
	Bandwidth *big.Int
}, error) {
	var out []interface{}
	err := _Arcana.contract.Call(opts, &out, "consumption", arg0)

	outstruct := new(struct {
		Store     *big.Int
		Bandwidth *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Store = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Bandwidth = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// Consumption is a free data retrieval call binding the contract method 0x633f7155.
//
// Solidity: function consumption(address ) view returns(uint256 store, uint256 bandwidth)
func (_Arcana *ArcanaSession) Consumption(arg0 common.Address) (struct {
	Store     *big.Int
	Bandwidth *big.Int
}, error) {
	return _Arcana.Contract.Consumption(&_Arcana.CallOpts, arg0)
}

// Consumption is a free data retrieval call binding the contract method 0x633f7155.
//
// Solidity: function consumption(address ) view returns(uint256 store, uint256 bandwidth)
func (_Arcana *ArcanaCallerSession) Consumption(arg0 common.Address) (struct {
	Store     *big.Int
	Bandwidth *big.Int
}, error) {
	return _Arcana.Contract.Consumption(&_Arcana.CallOpts, arg0)
}

// DefaultLimit is a free data retrieval call binding the contract method 0xe26b013b.
//
// Solidity: function defaultLimit() view returns(uint256 store, uint256 bandwidth)
func (_Arcana *ArcanaCaller) DefaultLimit(opts *bind.CallOpts) (struct {
	Store     *big.Int
	Bandwidth *big.Int
}, error) {
	var out []interface{}
	err := _Arcana.contract.Call(opts, &out, "defaultLimit")

	outstruct := new(struct {
		Store     *big.Int
		Bandwidth *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Store = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Bandwidth = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// DefaultLimit is a free data retrieval call binding the contract method 0xe26b013b.
//
// Solidity: function defaultLimit() view returns(uint256 store, uint256 bandwidth)
func (_Arcana *ArcanaSession) DefaultLimit() (struct {
	Store     *big.Int
	Bandwidth *big.Int
}, error) {
	return _Arcana.Contract.DefaultLimit(&_Arcana.CallOpts)
}

// DefaultLimit is a free data retrieval call binding the contract method 0xe26b013b.
//
// Solidity: function defaultLimit() view returns(uint256 store, uint256 bandwidth)
func (_Arcana *ArcanaCallerSession) DefaultLimit() (struct {
	Store     *big.Int
	Bandwidth *big.Int
}, error) {
	return _Arcana.Contract.DefaultLimit(&_Arcana.CallOpts)
}

// Delegators is a free data retrieval call binding the contract method 0x8d23fc61.
//
// Solidity: function delegators(address ) view returns(uint8)
func (_Arcana *ArcanaCaller) Delegators(opts *bind.CallOpts, arg0 common.Address) (uint8, error) {
	var out []interface{}
	err := _Arcana.contract.Call(opts, &out, "delegators", arg0)

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// Delegators is a free data retrieval call binding the contract method 0x8d23fc61.
//
// Solidity: function delegators(address ) view returns(uint8)
func (_Arcana *ArcanaSession) Delegators(arg0 common.Address) (uint8, error) {
	return _Arcana.Contract.Delegators(&_Arcana.CallOpts, arg0)
}

// Delegators is a free data retrieval call binding the contract method 0x8d23fc61.
//
// Solidity: function delegators(address ) view returns(uint8)
func (_Arcana *ArcanaCallerSession) Delegators(arg0 common.Address) (uint8, error) {
	return _Arcana.Contract.Delegators(&_Arcana.CallOpts, arg0)
}

// GetAppConfig is a free data retrieval call binding the contract method 0xf7664c72.
//
// Solidity: function getAppConfig() view returns(bytes32)
func (_Arcana *ArcanaCaller) GetAppConfig(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Arcana.contract.Call(opts, &out, "getAppConfig")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetAppConfig is a free data retrieval call binding the contract method 0xf7664c72.
//
// Solidity: function getAppConfig() view returns(bytes32)
func (_Arcana *ArcanaSession) GetAppConfig() ([32]byte, error) {
	return _Arcana.Contract.GetAppConfig(&_Arcana.CallOpts)
}

// GetAppConfig is a free data retrieval call binding the contract method 0xf7664c72.
//
// Solidity: function getAppConfig() view returns(bytes32)
func (_Arcana *ArcanaCallerSession) GetAppConfig() ([32]byte, error) {
	return _Arcana.Contract.GetAppConfig(&_Arcana.CallOpts)
}

// IsActive is a free data retrieval call binding the contract method 0x22f3e2d4.
//
// Solidity: function isActive() view returns(bool)
func (_Arcana *ArcanaCaller) IsActive(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _Arcana.contract.Call(opts, &out, "isActive")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsActive is a free data retrieval call binding the contract method 0x22f3e2d4.
//
// Solidity: function isActive() view returns(bool)
func (_Arcana *ArcanaSession) IsActive() (bool, error) {
	return _Arcana.Contract.IsActive(&_Arcana.CallOpts)
}

// IsActive is a free data retrieval call binding the contract method 0x22f3e2d4.
//
// Solidity: function isActive() view returns(bool)
func (_Arcana *ArcanaCallerSession) IsActive() (bool, error) {
	return _Arcana.Contract.IsActive(&_Arcana.CallOpts)
}

// IsTrustedForwarder is a free data retrieval call binding the contract method 0x572b6c05.
//
// Solidity: function isTrustedForwarder(address forwarder) view returns(bool)
func (_Arcana *ArcanaCaller) IsTrustedForwarder(opts *bind.CallOpts, forwarder common.Address) (bool, error) {
	var out []interface{}
	err := _Arcana.contract.Call(opts, &out, "isTrustedForwarder", forwarder)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsTrustedForwarder is a free data retrieval call binding the contract method 0x572b6c05.
//
// Solidity: function isTrustedForwarder(address forwarder) view returns(bool)
func (_Arcana *ArcanaSession) IsTrustedForwarder(forwarder common.Address) (bool, error) {
	return _Arcana.Contract.IsTrustedForwarder(&_Arcana.CallOpts, forwarder)
}

// IsTrustedForwarder is a free data retrieval call binding the contract method 0x572b6c05.
//
// Solidity: function isTrustedForwarder(address forwarder) view returns(bool)
func (_Arcana *ArcanaCallerSession) IsTrustedForwarder(forwarder common.Address) (bool, error) {
	return _Arcana.Contract.IsTrustedForwarder(&_Arcana.CallOpts, forwarder)
}

// Limit is a free data retrieval call binding the contract method 0xd8797262.
//
// Solidity: function limit(address ) view returns(uint256 store, uint256 bandwidth)
func (_Arcana *ArcanaCaller) Limit(opts *bind.CallOpts, arg0 common.Address) (struct {
	Store     *big.Int
	Bandwidth *big.Int
}, error) {
	var out []interface{}
	err := _Arcana.contract.Call(opts, &out, "limit", arg0)

	outstruct := new(struct {
		Store     *big.Int
		Bandwidth *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Store = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Bandwidth = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// Limit is a free data retrieval call binding the contract method 0xd8797262.
//
// Solidity: function limit(address ) view returns(uint256 store, uint256 bandwidth)
func (_Arcana *ArcanaSession) Limit(arg0 common.Address) (struct {
	Store     *big.Int
	Bandwidth *big.Int
}, error) {
	return _Arcana.Contract.Limit(&_Arcana.CallOpts, arg0)
}

// Limit is a free data retrieval call binding the contract method 0xd8797262.
//
// Solidity: function limit(address ) view returns(uint256 store, uint256 bandwidth)
func (_Arcana *ArcanaCallerSession) Limit(arg0 common.Address) (struct {
	Store     *big.Int
	Bandwidth *big.Int
}, error) {
	return _Arcana.Contract.Limit(&_Arcana.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Arcana *ArcanaCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Arcana.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Arcana *ArcanaSession) Owner() (common.Address, error) {
	return _Arcana.Contract.Owner(&_Arcana.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Arcana *ArcanaCallerSession) Owner() (common.Address, error) {
	return _Arcana.Contract.Owner(&_Arcana.CallOpts)
}

// TxCounter is a free data retrieval call binding the contract method 0x5531d119.
//
// Solidity: function txCounter(bytes32 ) view returns(bool)
func (_Arcana *ArcanaCaller) TxCounter(opts *bind.CallOpts, arg0 [32]byte) (bool, error) {
	var out []interface{}
	err := _Arcana.contract.Call(opts, &out, "txCounter", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// TxCounter is a free data retrieval call binding the contract method 0x5531d119.
//
// Solidity: function txCounter(bytes32 ) view returns(bool)
func (_Arcana *ArcanaSession) TxCounter(arg0 [32]byte) (bool, error) {
	return _Arcana.Contract.TxCounter(&_Arcana.CallOpts, arg0)
}

// TxCounter is a free data retrieval call binding the contract method 0x5531d119.
//
// Solidity: function txCounter(bytes32 ) view returns(bool)
func (_Arcana *ArcanaCallerSession) TxCounter(arg0 [32]byte) (bool, error) {
	return _Arcana.Contract.TxCounter(&_Arcana.CallOpts, arg0)
}

// Unpartitioned is a free data retrieval call binding the contract method 0xd699bfa8.
//
// Solidity: function unpartitioned() view returns(bool)
func (_Arcana *ArcanaCaller) Unpartitioned(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _Arcana.contract.Call(opts, &out, "unpartitioned")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Unpartitioned is a free data retrieval call binding the contract method 0xd699bfa8.
//
// Solidity: function unpartitioned() view returns(bool)
func (_Arcana *ArcanaSession) Unpartitioned() (bool, error) {
	return _Arcana.Contract.Unpartitioned(&_Arcana.CallOpts)
}

// Unpartitioned is a free data retrieval call binding the contract method 0xd699bfa8.
//
// Solidity: function unpartitioned() view returns(bool)
func (_Arcana *ArcanaCallerSession) Unpartitioned() (bool, error) {
	return _Arcana.Contract.Unpartitioned(&_Arcana.CallOpts)
}

// UserAppPermission is a free data retrieval call binding the contract method 0x79e04beb.
//
// Solidity: function userAppPermission(address ) view returns(uint8)
func (_Arcana *ArcanaCaller) UserAppPermission(opts *bind.CallOpts, arg0 common.Address) (uint8, error) {
	var out []interface{}
	err := _Arcana.contract.Call(opts, &out, "userAppPermission", arg0)

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// UserAppPermission is a free data retrieval call binding the contract method 0x79e04beb.
//
// Solidity: function userAppPermission(address ) view returns(uint8)
func (_Arcana *ArcanaSession) UserAppPermission(arg0 common.Address) (uint8, error) {
	return _Arcana.Contract.UserAppPermission(&_Arcana.CallOpts, arg0)
}

// UserAppPermission is a free data retrieval call binding the contract method 0x79e04beb.
//
// Solidity: function userAppPermission(address ) view returns(uint8)
func (_Arcana *ArcanaCallerSession) UserAppPermission(arg0 common.Address) (uint8, error) {
	return _Arcana.Contract.UserAppPermission(&_Arcana.CallOpts, arg0)
}

// UserVersion is a free data retrieval call binding the contract method 0xaea7d1c1.
//
// Solidity: function userVersion(address ) view returns(uint256)
func (_Arcana *ArcanaCaller) UserVersion(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Arcana.contract.Call(opts, &out, "userVersion", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// UserVersion is a free data retrieval call binding the contract method 0xaea7d1c1.
//
// Solidity: function userVersion(address ) view returns(uint256)
func (_Arcana *ArcanaSession) UserVersion(arg0 common.Address) (*big.Int, error) {
	return _Arcana.Contract.UserVersion(&_Arcana.CallOpts, arg0)
}

// UserVersion is a free data retrieval call binding the contract method 0xaea7d1c1.
//
// Solidity: function userVersion(address ) view returns(uint256)
func (_Arcana *ArcanaCallerSession) UserVersion(arg0 common.Address) (*big.Int, error) {
	return _Arcana.Contract.UserVersion(&_Arcana.CallOpts, arg0)
}

// WalletType is a free data retrieval call binding the contract method 0x5b648b0a.
//
// Solidity: function walletType() view returns(uint8)
func (_Arcana *ArcanaCaller) WalletType(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _Arcana.contract.Call(opts, &out, "walletType")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// WalletType is a free data retrieval call binding the contract method 0x5b648b0a.
//
// Solidity: function walletType() view returns(uint8)
func (_Arcana *ArcanaSession) WalletType() (uint8, error) {
	return _Arcana.Contract.WalletType(&_Arcana.CallOpts)
}

// WalletType is a free data retrieval call binding the contract method 0x5b648b0a.
//
// Solidity: function walletType() view returns(uint8)
func (_Arcana *ArcanaCallerSession) WalletType() (uint8, error) {
	return _Arcana.Contract.WalletType(&_Arcana.CallOpts)
}

// AddFile is a paid mutator transaction binding the contract method 0x6b318270.
//
// Solidity: function addFile(bytes32 did) returns()
func (_Arcana *ArcanaTransactor) AddFile(opts *bind.TransactOpts, did [32]byte) (*types.Transaction, error) {
	return _Arcana.contract.Transact(opts, "addFile", did)
}

// AddFile is a paid mutator transaction binding the contract method 0x6b318270.
//
// Solidity: function addFile(bytes32 did) returns()
func (_Arcana *ArcanaSession) AddFile(did [32]byte) (*types.Transaction, error) {
	return _Arcana.Contract.AddFile(&_Arcana.TransactOpts, did)
}

// AddFile is a paid mutator transaction binding the contract method 0x6b318270.
//
// Solidity: function addFile(bytes32 did) returns()
func (_Arcana *ArcanaTransactorSession) AddFile(did [32]byte) (*types.Transaction, error) {
	return _Arcana.Contract.AddFile(&_Arcana.TransactOpts, did)
}

// ChangeFileOwner is a paid mutator transaction binding the contract method 0xc5b26447.
//
// Solidity: function changeFileOwner(bytes32 did, address newOwner) returns()
func (_Arcana *ArcanaTransactor) ChangeFileOwner(opts *bind.TransactOpts, did [32]byte, newOwner common.Address) (*types.Transaction, error) {
	return _Arcana.contract.Transact(opts, "changeFileOwner", did, newOwner)
}

// ChangeFileOwner is a paid mutator transaction binding the contract method 0xc5b26447.
//
// Solidity: function changeFileOwner(bytes32 did, address newOwner) returns()
func (_Arcana *ArcanaSession) ChangeFileOwner(did [32]byte, newOwner common.Address) (*types.Transaction, error) {
	return _Arcana.Contract.ChangeFileOwner(&_Arcana.TransactOpts, did, newOwner)
}

// ChangeFileOwner is a paid mutator transaction binding the contract method 0xc5b26447.
//
// Solidity: function changeFileOwner(bytes32 did, address newOwner) returns()
func (_Arcana *ArcanaTransactorSession) ChangeFileOwner(did [32]byte, newOwner common.Address) (*types.Transaction, error) {
	return _Arcana.Contract.ChangeFileOwner(&_Arcana.TransactOpts, did, newOwner)
}

// DeleteApp is a paid mutator transaction binding the contract method 0x1cf93926.
//
// Solidity: function deleteApp() returns()
func (_Arcana *ArcanaTransactor) DeleteApp(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Arcana.contract.Transact(opts, "deleteApp")
}

// DeleteApp is a paid mutator transaction binding the contract method 0x1cf93926.
//
// Solidity: function deleteApp() returns()
func (_Arcana *ArcanaSession) DeleteApp() (*types.Transaction, error) {
	return _Arcana.Contract.DeleteApp(&_Arcana.TransactOpts)
}

// DeleteApp is a paid mutator transaction binding the contract method 0x1cf93926.
//
// Solidity: function deleteApp() returns()
func (_Arcana *ArcanaTransactorSession) DeleteApp() (*types.Transaction, error) {
	return _Arcana.Contract.DeleteApp(&_Arcana.TransactOpts)
}

// Download is a paid mutator transaction binding the contract method 0x268737de.
//
// Solidity: function download(bytes32 did, address ephemeralWallet) returns()
func (_Arcana *ArcanaTransactor) Download(opts *bind.TransactOpts, did [32]byte, ephemeralWallet common.Address) (*types.Transaction, error) {
	return _Arcana.contract.Transact(opts, "download", did, ephemeralWallet)
}

// Download is a paid mutator transaction binding the contract method 0x268737de.
//
// Solidity: function download(bytes32 did, address ephemeralWallet) returns()
func (_Arcana *ArcanaSession) Download(did [32]byte, ephemeralWallet common.Address) (*types.Transaction, error) {
	return _Arcana.Contract.Download(&_Arcana.TransactOpts, did, ephemeralWallet)
}

// Download is a paid mutator transaction binding the contract method 0x268737de.
//
// Solidity: function download(bytes32 did, address ephemeralWallet) returns()
func (_Arcana *ArcanaTransactorSession) Download(did [32]byte, ephemeralWallet common.Address) (*types.Transaction, error) {
	return _Arcana.Contract.Download(&_Arcana.TransactOpts, did, ephemeralWallet)
}

// DownloadClose is a paid mutator transaction binding the contract method 0x288cdf0a.
//
// Solidity: function downloadClose(bytes32 did, bytes32 txHash) returns()
func (_Arcana *ArcanaTransactor) DownloadClose(opts *bind.TransactOpts, did [32]byte, txHash [32]byte) (*types.Transaction, error) {
	return _Arcana.contract.Transact(opts, "downloadClose", did, txHash)
}

// DownloadClose is a paid mutator transaction binding the contract method 0x288cdf0a.
//
// Solidity: function downloadClose(bytes32 did, bytes32 txHash) returns()
func (_Arcana *ArcanaSession) DownloadClose(did [32]byte, txHash [32]byte) (*types.Transaction, error) {
	return _Arcana.Contract.DownloadClose(&_Arcana.TransactOpts, did, txHash)
}

// DownloadClose is a paid mutator transaction binding the contract method 0x288cdf0a.
//
// Solidity: function downloadClose(bytes32 did, bytes32 txHash) returns()
func (_Arcana *ArcanaTransactorSession) DownloadClose(did [32]byte, txHash [32]byte) (*types.Transaction, error) {
	return _Arcana.Contract.DownloadClose(&_Arcana.TransactOpts, did, txHash)
}

// EditAppPermission is a paid mutator transaction binding the contract method 0x3a87872a.
//
// Solidity: function editAppPermission(uint8 appPermission, bool add) returns()
func (_Arcana *ArcanaTransactor) EditAppPermission(opts *bind.TransactOpts, appPermission uint8, add bool) (*types.Transaction, error) {
	return _Arcana.contract.Transact(opts, "editAppPermission", appPermission, add)
}

// EditAppPermission is a paid mutator transaction binding the contract method 0x3a87872a.
//
// Solidity: function editAppPermission(uint8 appPermission, bool add) returns()
func (_Arcana *ArcanaSession) EditAppPermission(appPermission uint8, add bool) (*types.Transaction, error) {
	return _Arcana.Contract.EditAppPermission(&_Arcana.TransactOpts, appPermission, add)
}

// EditAppPermission is a paid mutator transaction binding the contract method 0x3a87872a.
//
// Solidity: function editAppPermission(uint8 appPermission, bool add) returns()
func (_Arcana *ArcanaTransactorSession) EditAppPermission(appPermission uint8, add bool) (*types.Transaction, error) {
	return _Arcana.Contract.EditAppPermission(&_Arcana.TransactOpts, appPermission, add)
}

// GrantAppPermission is a paid mutator transaction binding the contract method 0x8a91caf9.
//
// Solidity: function grantAppPermission() returns()
func (_Arcana *ArcanaTransactor) GrantAppPermission(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Arcana.contract.Transact(opts, "grantAppPermission")
}

// GrantAppPermission is a paid mutator transaction binding the contract method 0x8a91caf9.
//
// Solidity: function grantAppPermission() returns()
func (_Arcana *ArcanaSession) GrantAppPermission() (*types.Transaction, error) {
	return _Arcana.Contract.GrantAppPermission(&_Arcana.TransactOpts)
}

// GrantAppPermission is a paid mutator transaction binding the contract method 0x8a91caf9.
//
// Solidity: function grantAppPermission() returns()
func (_Arcana *ArcanaTransactorSession) GrantAppPermission() (*types.Transaction, error) {
	return _Arcana.Contract.GrantAppPermission(&_Arcana.TransactOpts)
}

// Initialize is a paid mutator transaction binding the contract method 0x4ceb0eea.
//
// Solidity: function initialize(address factoryAddress, address relayer, bool aggregateLoginValue, address did, bytes32 appConfigValue) returns()
func (_Arcana *ArcanaTransactor) Initialize(opts *bind.TransactOpts, factoryAddress common.Address, relayer common.Address, aggregateLoginValue bool, did common.Address, appConfigValue [32]byte) (*types.Transaction, error) {
	return _Arcana.contract.Transact(opts, "initialize", factoryAddress, relayer, aggregateLoginValue, did, appConfigValue)
}

// Initialize is a paid mutator transaction binding the contract method 0x4ceb0eea.
//
// Solidity: function initialize(address factoryAddress, address relayer, bool aggregateLoginValue, address did, bytes32 appConfigValue) returns()
func (_Arcana *ArcanaSession) Initialize(factoryAddress common.Address, relayer common.Address, aggregateLoginValue bool, did common.Address, appConfigValue [32]byte) (*types.Transaction, error) {
	return _Arcana.Contract.Initialize(&_Arcana.TransactOpts, factoryAddress, relayer, aggregateLoginValue, did, appConfigValue)
}

// Initialize is a paid mutator transaction binding the contract method 0x4ceb0eea.
//
// Solidity: function initialize(address factoryAddress, address relayer, bool aggregateLoginValue, address did, bytes32 appConfigValue) returns()
func (_Arcana *ArcanaTransactorSession) Initialize(factoryAddress common.Address, relayer common.Address, aggregateLoginValue bool, did common.Address, appConfigValue [32]byte) (*types.Transaction, error) {
	return _Arcana.Contract.Initialize(&_Arcana.TransactOpts, factoryAddress, relayer, aggregateLoginValue, did, appConfigValue)
}

// LinkNFT is a paid mutator transaction binding the contract method 0xbab43293.
//
// Solidity: function linkNFT(bytes32 did, uint256 tokenId, address nftContract, uint256 chainId) returns()
func (_Arcana *ArcanaTransactor) LinkNFT(opts *bind.TransactOpts, did [32]byte, tokenId *big.Int, nftContract common.Address, chainId *big.Int) (*types.Transaction, error) {
	return _Arcana.contract.Transact(opts, "linkNFT", did, tokenId, nftContract, chainId)
}

// LinkNFT is a paid mutator transaction binding the contract method 0xbab43293.
//
// Solidity: function linkNFT(bytes32 did, uint256 tokenId, address nftContract, uint256 chainId) returns()
func (_Arcana *ArcanaSession) LinkNFT(did [32]byte, tokenId *big.Int, nftContract common.Address, chainId *big.Int) (*types.Transaction, error) {
	return _Arcana.Contract.LinkNFT(&_Arcana.TransactOpts, did, tokenId, nftContract, chainId)
}

// LinkNFT is a paid mutator transaction binding the contract method 0xbab43293.
//
// Solidity: function linkNFT(bytes32 did, uint256 tokenId, address nftContract, uint256 chainId) returns()
func (_Arcana *ArcanaTransactorSession) LinkNFT(did [32]byte, tokenId *big.Int, nftContract common.Address, chainId *big.Int) (*types.Transaction, error) {
	return _Arcana.Contract.LinkNFT(&_Arcana.TransactOpts, did, tokenId, nftContract, chainId)
}

// RemoveUserFile is a paid mutator transaction binding the contract method 0x19fcc8f9.
//
// Solidity: function removeUserFile(bytes32 did) returns()
func (_Arcana *ArcanaTransactor) RemoveUserFile(opts *bind.TransactOpts, did [32]byte) (*types.Transaction, error) {
	return _Arcana.contract.Transact(opts, "removeUserFile", did)
}

// RemoveUserFile is a paid mutator transaction binding the contract method 0x19fcc8f9.
//
// Solidity: function removeUserFile(bytes32 did) returns()
func (_Arcana *ArcanaSession) RemoveUserFile(did [32]byte) (*types.Transaction, error) {
	return _Arcana.Contract.RemoveUserFile(&_Arcana.TransactOpts, did)
}

// RemoveUserFile is a paid mutator transaction binding the contract method 0x19fcc8f9.
//
// Solidity: function removeUserFile(bytes32 did) returns()
func (_Arcana *ArcanaTransactorSession) RemoveUserFile(did [32]byte) (*types.Transaction, error) {
	return _Arcana.Contract.RemoveUserFile(&_Arcana.TransactOpts, did)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Arcana *ArcanaTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Arcana.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Arcana *ArcanaSession) RenounceOwnership() (*types.Transaction, error) {
	return _Arcana.Contract.RenounceOwnership(&_Arcana.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Arcana *ArcanaTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Arcana.Contract.RenounceOwnership(&_Arcana.TransactOpts)
}

// RevokeApp is a paid mutator transaction binding the contract method 0x1807a311.
//
// Solidity: function revokeApp() returns()
func (_Arcana *ArcanaTransactor) RevokeApp(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Arcana.contract.Transact(opts, "revokeApp")
}

// RevokeApp is a paid mutator transaction binding the contract method 0x1807a311.
//
// Solidity: function revokeApp() returns()
func (_Arcana *ArcanaSession) RevokeApp() (*types.Transaction, error) {
	return _Arcana.Contract.RevokeApp(&_Arcana.TransactOpts)
}

// RevokeApp is a paid mutator transaction binding the contract method 0x1807a311.
//
// Solidity: function revokeApp() returns()
func (_Arcana *ArcanaTransactorSession) RevokeApp() (*types.Transaction, error) {
	return _Arcana.Contract.RevokeApp(&_Arcana.TransactOpts)
}

// SetAppConfig is a paid mutator transaction binding the contract method 0x54c0439d.
//
// Solidity: function setAppConfig(bytes32 appConfig) returns()
func (_Arcana *ArcanaTransactor) SetAppConfig(opts *bind.TransactOpts, appConfig [32]byte) (*types.Transaction, error) {
	return _Arcana.contract.Transact(opts, "setAppConfig", appConfig)
}

// SetAppConfig is a paid mutator transaction binding the contract method 0x54c0439d.
//
// Solidity: function setAppConfig(bytes32 appConfig) returns()
func (_Arcana *ArcanaSession) SetAppConfig(appConfig [32]byte) (*types.Transaction, error) {
	return _Arcana.Contract.SetAppConfig(&_Arcana.TransactOpts, appConfig)
}

// SetAppConfig is a paid mutator transaction binding the contract method 0x54c0439d.
//
// Solidity: function setAppConfig(bytes32 appConfig) returns()
func (_Arcana *ArcanaTransactorSession) SetAppConfig(appConfig [32]byte) (*types.Transaction, error) {
	return _Arcana.Contract.SetAppConfig(&_Arcana.TransactOpts, appConfig)
}

// SetAppLimit is a paid mutator transaction binding the contract method 0x20777435.
//
// Solidity: function setAppLimit(uint256 store, uint256 bandwidth) returns()
func (_Arcana *ArcanaTransactor) SetAppLimit(opts *bind.TransactOpts, store *big.Int, bandwidth *big.Int) (*types.Transaction, error) {
	return _Arcana.contract.Transact(opts, "setAppLimit", store, bandwidth)
}

// SetAppLimit is a paid mutator transaction binding the contract method 0x20777435.
//
// Solidity: function setAppLimit(uint256 store, uint256 bandwidth) returns()
func (_Arcana *ArcanaSession) SetAppLimit(store *big.Int, bandwidth *big.Int) (*types.Transaction, error) {
	return _Arcana.Contract.SetAppLimit(&_Arcana.TransactOpts, store, bandwidth)
}

// SetAppLimit is a paid mutator transaction binding the contract method 0x20777435.
//
// Solidity: function setAppLimit(uint256 store, uint256 bandwidth) returns()
func (_Arcana *ArcanaTransactorSession) SetAppLimit(store *big.Int, bandwidth *big.Int) (*types.Transaction, error) {
	return _Arcana.Contract.SetAppLimit(&_Arcana.TransactOpts, store, bandwidth)
}

// SetDefaultLimit is a paid mutator transaction binding the contract method 0x377dd46e.
//
// Solidity: function setDefaultLimit(uint256 store, uint256 bandwidth) returns()
func (_Arcana *ArcanaTransactor) SetDefaultLimit(opts *bind.TransactOpts, store *big.Int, bandwidth *big.Int) (*types.Transaction, error) {
	return _Arcana.contract.Transact(opts, "setDefaultLimit", store, bandwidth)
}

// SetDefaultLimit is a paid mutator transaction binding the contract method 0x377dd46e.
//
// Solidity: function setDefaultLimit(uint256 store, uint256 bandwidth) returns()
func (_Arcana *ArcanaSession) SetDefaultLimit(store *big.Int, bandwidth *big.Int) (*types.Transaction, error) {
	return _Arcana.Contract.SetDefaultLimit(&_Arcana.TransactOpts, store, bandwidth)
}

// SetDefaultLimit is a paid mutator transaction binding the contract method 0x377dd46e.
//
// Solidity: function setDefaultLimit(uint256 store, uint256 bandwidth) returns()
func (_Arcana *ArcanaTransactorSession) SetDefaultLimit(store *big.Int, bandwidth *big.Int) (*types.Transaction, error) {
	return _Arcana.Contract.SetDefaultLimit(&_Arcana.TransactOpts, store, bandwidth)
}

// SetUnPartitioned is a paid mutator transaction binding the contract method 0xb0fcc28d.
//
// Solidity: function setUnPartitioned(bool status) returns()
func (_Arcana *ArcanaTransactor) SetUnPartitioned(opts *bind.TransactOpts, status bool) (*types.Transaction, error) {
	return _Arcana.contract.Transact(opts, "setUnPartitioned", status)
}

// SetUnPartitioned is a paid mutator transaction binding the contract method 0xb0fcc28d.
//
// Solidity: function setUnPartitioned(bool status) returns()
func (_Arcana *ArcanaSession) SetUnPartitioned(status bool) (*types.Transaction, error) {
	return _Arcana.Contract.SetUnPartitioned(&_Arcana.TransactOpts, status)
}

// SetUnPartitioned is a paid mutator transaction binding the contract method 0xb0fcc28d.
//
// Solidity: function setUnPartitioned(bool status) returns()
func (_Arcana *ArcanaTransactorSession) SetUnPartitioned(status bool) (*types.Transaction, error) {
	return _Arcana.Contract.SetUnPartitioned(&_Arcana.TransactOpts, status)
}

// SetUserLevelLimit is a paid mutator transaction binding the contract method 0x6a29afdc.
//
// Solidity: function setUserLevelLimit(address user, uint256 store, uint256 bandwidth) returns()
func (_Arcana *ArcanaTransactor) SetUserLevelLimit(opts *bind.TransactOpts, user common.Address, store *big.Int, bandwidth *big.Int) (*types.Transaction, error) {
	return _Arcana.contract.Transact(opts, "setUserLevelLimit", user, store, bandwidth)
}

// SetUserLevelLimit is a paid mutator transaction binding the contract method 0x6a29afdc.
//
// Solidity: function setUserLevelLimit(address user, uint256 store, uint256 bandwidth) returns()
func (_Arcana *ArcanaSession) SetUserLevelLimit(user common.Address, store *big.Int, bandwidth *big.Int) (*types.Transaction, error) {
	return _Arcana.Contract.SetUserLevelLimit(&_Arcana.TransactOpts, user, store, bandwidth)
}

// SetUserLevelLimit is a paid mutator transaction binding the contract method 0x6a29afdc.
//
// Solidity: function setUserLevelLimit(address user, uint256 store, uint256 bandwidth) returns()
func (_Arcana *ArcanaTransactorSession) SetUserLevelLimit(user common.Address, store *big.Int, bandwidth *big.Int) (*types.Transaction, error) {
	return _Arcana.Contract.SetUserLevelLimit(&_Arcana.TransactOpts, user, store, bandwidth)
}

// ToggleWalletType is a paid mutator transaction binding the contract method 0xc9d869b4.
//
// Solidity: function toggleWalletType() returns()
func (_Arcana *ArcanaTransactor) ToggleWalletType(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Arcana.contract.Transact(opts, "toggleWalletType")
}

// ToggleWalletType is a paid mutator transaction binding the contract method 0xc9d869b4.
//
// Solidity: function toggleWalletType() returns()
func (_Arcana *ArcanaSession) ToggleWalletType() (*types.Transaction, error) {
	return _Arcana.Contract.ToggleWalletType(&_Arcana.TransactOpts)
}

// ToggleWalletType is a paid mutator transaction binding the contract method 0xc9d869b4.
//
// Solidity: function toggleWalletType() returns()
func (_Arcana *ArcanaTransactorSession) ToggleWalletType() (*types.Transaction, error) {
	return _Arcana.Contract.ToggleWalletType(&_Arcana.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Arcana *ArcanaTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Arcana.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Arcana *ArcanaSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Arcana.Contract.TransferOwnership(&_Arcana.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Arcana *ArcanaTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Arcana.Contract.TransferOwnership(&_Arcana.TransactOpts, newOwner)
}

// UpdateDelegator is a paid mutator transaction binding the contract method 0x88e8e531.
//
// Solidity: function updateDelegator(address delegator, uint8 control, bool add) returns()
func (_Arcana *ArcanaTransactor) UpdateDelegator(opts *bind.TransactOpts, delegator common.Address, control uint8, add bool) (*types.Transaction, error) {
	return _Arcana.contract.Transact(opts, "updateDelegator", delegator, control, add)
}

// UpdateDelegator is a paid mutator transaction binding the contract method 0x88e8e531.
//
// Solidity: function updateDelegator(address delegator, uint8 control, bool add) returns()
func (_Arcana *ArcanaSession) UpdateDelegator(delegator common.Address, control uint8, add bool) (*types.Transaction, error) {
	return _Arcana.Contract.UpdateDelegator(&_Arcana.TransactOpts, delegator, control, add)
}

// UpdateDelegator is a paid mutator transaction binding the contract method 0x88e8e531.
//
// Solidity: function updateDelegator(address delegator, uint8 control, bool add) returns()
func (_Arcana *ArcanaTransactorSession) UpdateDelegator(delegator common.Address, control uint8, add bool) (*types.Transaction, error) {
	return _Arcana.Contract.UpdateDelegator(&_Arcana.TransactOpts, delegator, control, add)
}

// UpdateRuleSet is a paid mutator transaction binding the contract method 0xb7abbeea.
//
// Solidity: function updateRuleSet(bytes32 did, bytes32 ruleHash) returns()
func (_Arcana *ArcanaTransactor) UpdateRuleSet(opts *bind.TransactOpts, did [32]byte, ruleHash [32]byte) (*types.Transaction, error) {
	return _Arcana.contract.Transact(opts, "updateRuleSet", did, ruleHash)
}

// UpdateRuleSet is a paid mutator transaction binding the contract method 0xb7abbeea.
//
// Solidity: function updateRuleSet(bytes32 did, bytes32 ruleHash) returns()
func (_Arcana *ArcanaSession) UpdateRuleSet(did [32]byte, ruleHash [32]byte) (*types.Transaction, error) {
	return _Arcana.Contract.UpdateRuleSet(&_Arcana.TransactOpts, did, ruleHash)
}

// UpdateRuleSet is a paid mutator transaction binding the contract method 0xb7abbeea.
//
// Solidity: function updateRuleSet(bytes32 did, bytes32 ruleHash) returns()
func (_Arcana *ArcanaTransactorSession) UpdateRuleSet(did [32]byte, ruleHash [32]byte) (*types.Transaction, error) {
	return _Arcana.Contract.UpdateRuleSet(&_Arcana.TransactOpts, did, ruleHash)
}

// UploadClose is a paid mutator transaction binding the contract method 0x63a15202.
//
// Solidity: function uploadClose(bytes32 did) returns()
func (_Arcana *ArcanaTransactor) UploadClose(opts *bind.TransactOpts, did [32]byte) (*types.Transaction, error) {
	return _Arcana.contract.Transact(opts, "uploadClose", did)
}

// UploadClose is a paid mutator transaction binding the contract method 0x63a15202.
//
// Solidity: function uploadClose(bytes32 did) returns()
func (_Arcana *ArcanaSession) UploadClose(did [32]byte) (*types.Transaction, error) {
	return _Arcana.Contract.UploadClose(&_Arcana.TransactOpts, did)
}

// UploadClose is a paid mutator transaction binding the contract method 0x63a15202.
//
// Solidity: function uploadClose(bytes32 did) returns()
func (_Arcana *ArcanaTransactorSession) UploadClose(did [32]byte) (*types.Transaction, error) {
	return _Arcana.Contract.UploadClose(&_Arcana.TransactOpts, did)
}

// UploadInit is a paid mutator transaction binding the contract method 0x9534e594.
//
// Solidity: function uploadInit(bytes32 did, uint256 fileSize, bytes32 name, bytes32 fileHash, address storageNode, address ephemeralAddress) returns()
func (_Arcana *ArcanaTransactor) UploadInit(opts *bind.TransactOpts, did [32]byte, fileSize *big.Int, name [32]byte, fileHash [32]byte, storageNode common.Address, ephemeralAddress common.Address) (*types.Transaction, error) {
	return _Arcana.contract.Transact(opts, "uploadInit", did, fileSize, name, fileHash, storageNode, ephemeralAddress)
}

// UploadInit is a paid mutator transaction binding the contract method 0x9534e594.
//
// Solidity: function uploadInit(bytes32 did, uint256 fileSize, bytes32 name, bytes32 fileHash, address storageNode, address ephemeralAddress) returns()
func (_Arcana *ArcanaSession) UploadInit(did [32]byte, fileSize *big.Int, name [32]byte, fileHash [32]byte, storageNode common.Address, ephemeralAddress common.Address) (*types.Transaction, error) {
	return _Arcana.Contract.UploadInit(&_Arcana.TransactOpts, did, fileSize, name, fileHash, storageNode, ephemeralAddress)
}

// UploadInit is a paid mutator transaction binding the contract method 0x9534e594.
//
// Solidity: function uploadInit(bytes32 did, uint256 fileSize, bytes32 name, bytes32 fileHash, address storageNode, address ephemeralAddress) returns()
func (_Arcana *ArcanaTransactorSession) UploadInit(did [32]byte, fileSize *big.Int, name [32]byte, fileHash [32]byte, storageNode common.Address, ephemeralAddress common.Address) (*types.Transaction, error) {
	return _Arcana.Contract.UploadInit(&_Arcana.TransactOpts, did, fileSize, name, fileHash, storageNode, ephemeralAddress)
}

// ArcanaDeleteAppIterator is returned from FilterDeleteApp and is used to iterate over the raw logs and unpacked data for DeleteApp events raised by the Arcana contract.
type ArcanaDeleteAppIterator struct {
	Event *ArcanaDeleteApp // Event containing the contract specifics and raw log

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
func (it *ArcanaDeleteAppIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ArcanaDeleteApp)
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
		it.Event = new(ArcanaDeleteApp)
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
func (it *ArcanaDeleteAppIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ArcanaDeleteAppIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ArcanaDeleteApp represents a DeleteApp event raised by the Arcana contract.
type ArcanaDeleteApp struct {
	Owner common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterDeleteApp is a free log retrieval operation binding the contract event 0xe8f8f4a14f262bd706c406540e1b99c2265e2ba250c414340f7525c3926958f2.
//
// Solidity: event DeleteApp(address owner)
func (_Arcana *ArcanaFilterer) FilterDeleteApp(opts *bind.FilterOpts) (*ArcanaDeleteAppIterator, error) {

	logs, sub, err := _Arcana.contract.FilterLogs(opts, "DeleteApp")
	if err != nil {
		return nil, err
	}
	return &ArcanaDeleteAppIterator{contract: _Arcana.contract, event: "DeleteApp", logs: logs, sub: sub}, nil
}

// WatchDeleteApp is a free log subscription operation binding the contract event 0xe8f8f4a14f262bd706c406540e1b99c2265e2ba250c414340f7525c3926958f2.
//
// Solidity: event DeleteApp(address owner)
func (_Arcana *ArcanaFilterer) WatchDeleteApp(opts *bind.WatchOpts, sink chan<- *ArcanaDeleteApp) (event.Subscription, error) {

	logs, sub, err := _Arcana.contract.WatchLogs(opts, "DeleteApp")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ArcanaDeleteApp)
				if err := _Arcana.contract.UnpackLog(event, "DeleteApp", log); err != nil {
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

// ParseDeleteApp is a log parse operation binding the contract event 0xe8f8f4a14f262bd706c406540e1b99c2265e2ba250c414340f7525c3926958f2.
//
// Solidity: event DeleteApp(address owner)
func (_Arcana *ArcanaFilterer) ParseDeleteApp(log types.Log) (*ArcanaDeleteApp, error) {
	event := new(ArcanaDeleteApp)
	if err := _Arcana.contract.UnpackLog(event, "DeleteApp", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ArcanaDownloadViaRuleSetIterator is returned from FilterDownloadViaRuleSet and is used to iterate over the raw logs and unpacked data for DownloadViaRuleSet events raised by the Arcana contract.
type ArcanaDownloadViaRuleSetIterator struct {
	Event *ArcanaDownloadViaRuleSet // Event containing the contract specifics and raw log

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
func (it *ArcanaDownloadViaRuleSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ArcanaDownloadViaRuleSet)
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
		it.Event = new(ArcanaDownloadViaRuleSet)
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
func (it *ArcanaDownloadViaRuleSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ArcanaDownloadViaRuleSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ArcanaDownloadViaRuleSet represents a DownloadViaRuleSet event raised by the Arcana contract.
type ArcanaDownloadViaRuleSet struct {
	Did  [32]byte
	User common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterDownloadViaRuleSet is a free log retrieval operation binding the contract event 0x2a395d2e297515f7afc1f72e10bb10e4349b4c5a816ac345df615bec4a34c0cd.
//
// Solidity: event DownloadViaRuleSet(bytes32 did, address user)
func (_Arcana *ArcanaFilterer) FilterDownloadViaRuleSet(opts *bind.FilterOpts) (*ArcanaDownloadViaRuleSetIterator, error) {

	logs, sub, err := _Arcana.contract.FilterLogs(opts, "DownloadViaRuleSet")
	if err != nil {
		return nil, err
	}
	return &ArcanaDownloadViaRuleSetIterator{contract: _Arcana.contract, event: "DownloadViaRuleSet", logs: logs, sub: sub}, nil
}

// WatchDownloadViaRuleSet is a free log subscription operation binding the contract event 0x2a395d2e297515f7afc1f72e10bb10e4349b4c5a816ac345df615bec4a34c0cd.
//
// Solidity: event DownloadViaRuleSet(bytes32 did, address user)
func (_Arcana *ArcanaFilterer) WatchDownloadViaRuleSet(opts *bind.WatchOpts, sink chan<- *ArcanaDownloadViaRuleSet) (event.Subscription, error) {

	logs, sub, err := _Arcana.contract.WatchLogs(opts, "DownloadViaRuleSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ArcanaDownloadViaRuleSet)
				if err := _Arcana.contract.UnpackLog(event, "DownloadViaRuleSet", log); err != nil {
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

// ParseDownloadViaRuleSet is a log parse operation binding the contract event 0x2a395d2e297515f7afc1f72e10bb10e4349b4c5a816ac345df615bec4a34c0cd.
//
// Solidity: event DownloadViaRuleSet(bytes32 did, address user)
func (_Arcana *ArcanaFilterer) ParseDownloadViaRuleSet(log types.Log) (*ArcanaDownloadViaRuleSet, error) {
	event := new(ArcanaDownloadViaRuleSet)
	if err := _Arcana.contract.UnpackLog(event, "DownloadViaRuleSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ArcanaInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the Arcana contract.
type ArcanaInitializedIterator struct {
	Event *ArcanaInitialized // Event containing the contract specifics and raw log

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
func (it *ArcanaInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ArcanaInitialized)
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
		it.Event = new(ArcanaInitialized)
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
func (it *ArcanaInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ArcanaInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ArcanaInitialized represents a Initialized event raised by the Arcana contract.
type ArcanaInitialized struct {
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_Arcana *ArcanaFilterer) FilterInitialized(opts *bind.FilterOpts) (*ArcanaInitializedIterator, error) {

	logs, sub, err := _Arcana.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &ArcanaInitializedIterator{contract: _Arcana.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_Arcana *ArcanaFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *ArcanaInitialized) (event.Subscription, error) {

	logs, sub, err := _Arcana.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ArcanaInitialized)
				if err := _Arcana.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_Arcana *ArcanaFilterer) ParseInitialized(log types.Log) (*ArcanaInitialized, error) {
	event := new(ArcanaInitialized)
	if err := _Arcana.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ArcanaOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Arcana contract.
type ArcanaOwnershipTransferredIterator struct {
	Event *ArcanaOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *ArcanaOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ArcanaOwnershipTransferred)
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
		it.Event = new(ArcanaOwnershipTransferred)
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
func (it *ArcanaOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ArcanaOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ArcanaOwnershipTransferred represents a OwnershipTransferred event raised by the Arcana contract.
type ArcanaOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Arcana *ArcanaFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*ArcanaOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Arcana.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &ArcanaOwnershipTransferredIterator{contract: _Arcana.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Arcana *ArcanaFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *ArcanaOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Arcana.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ArcanaOwnershipTransferred)
				if err := _Arcana.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_Arcana *ArcanaFilterer) ParseOwnershipTransferred(log types.Log) (*ArcanaOwnershipTransferred, error) {
	event := new(ArcanaOwnershipTransferred)
	if err := _Arcana.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
