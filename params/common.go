package params

import (
	"errors"
	"github.com/ethereum/go-ethereum/common"

	"math/big"
)

var (
	baseUnit           = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	FbaseUnit          = new(big.Float).SetFloat64(float64(baseUnit.Int64()))
	Base               = new(big.Int).SetUint64(10000)
	InvalidFee         = big.NewInt(65535)
	RelayerAddress     = common.BytesToAddress([]byte("RelayerAddress"))
	HeaderStoreAddress = common.BytesToAddress([]byte("headerstoreAddress"))
	TxVerifyAddress    = common.BytesToAddress([]byte("txVerifyAddress"))
)

var RelayerGas = map[string]uint64{
	"getBalance":      450000,
	"register":        2400000,
	"append":          2400000,
	"withdraw":        2520000,
	"getPeriodHeight": 450000,
	"getRelayers":     450000,
}

var (
	CountInEpoch                       = 100
	MaxRedeemHeight             uint64 = 200
	NewEpochLength              uint64 = 200
	ElectionPoint               uint64 = 20
	FirstNewEpochID             uint64 = 1
	PowForkPoint                uint64 = 0
	ElectionMinLimitForRegister        = new(big.Int).Mul(big.NewInt(100000), big.NewInt(1e18))
	MinWorkEfficiency           uint64 = 1 //every relayer sync 1000 block at least
)

var (
	ErrInvalidParam      = errors.New("Invalid Param")
	ErrOverEpochID       = errors.New("Over epoch id")
	ErrNotSequential     = errors.New("epoch id not sequential")
	ErrInvalidEpochInfo  = errors.New("Invalid epoch info")
	ErrNotFoundEpoch     = errors.New("cann't found the epoch info")
	ErrInvalidRegister   = errors.New("Invalid register account")
	ErrMatchEpochID      = errors.New("wrong match epoch id in a reward block")
	ErrNotRegister       = errors.New("Not match the register account")
	ErrNotDelegation     = errors.New("Not match the account")
	ErrNotMatchEpochInfo = errors.New("the epoch info is not match with accounts")
	ErrNotElectionTime   = errors.New("not time to election the next relayer")
	ErrAmountOver        = errors.New("the amount more than register amount")
	ErrDelegationSelf    = errors.New("wrong")
	ErrRedeemAmount      = errors.New("wrong redeem amount")
	ErrForbidAddress     = errors.New("Forbidding Address")
)

const (
	// StateRegisterOnce can be election only once
	StateRegisterOnce uint8 = 1 << iota
	// StateResgisterAuto can be election in every epoch
	StateResgisterAuto
	// StateUnregister can be redeem real time (after MaxRedeemHeight block)
	StateUnregister
	// StateUnregistered flag the asset which is unregistered in the height is redeemed
	StateUnregistered
)
const (
	OpQueryRegister uint8 = 1 << iota
	OpQueryLocked
	OpQueryUnlocking
	OpQueryReward
	OpQueryFine
)

const (
	StateUnusedFlag    = 0xa0
	StateUsedFlag      = 0xa1
	StateSwitchingFlag = 0xa2
	StateRemovedFlag   = 0xa3
	StateAppendFlag    = 0xa4
	// health enter type
	TypeFixed  = 0xa1
	TypeWorked = 0xa2
	TypeBack   = 0xa3
)

const RelayerABIJSON = `[
  {
    "name": "Register",
    "inputs": [
      {
        "type": "address",
        "name": "from",
        "indexed": true
      },
      {
        "type": "uint256",
        "name": "value",
        "indexed": false
      }
    ],
    "anonymous": false,
    "type": "event"
  },
  {
    "name": "Withdraw",
    "inputs": [
      {
        "type": "address",
        "name": "from",
        "indexed": true
      },
      {
        "type": "uint256",
        "name": "value",
        "indexed": false
      }
    ],
    "anonymous": false,
    "type": "event"
  },
  {
    "name": "Unregister",
    "inputs": [
      {
        "type": "address",
        "name": "from",
        "indexed": true
      },
      {
        "type": "uint256",
        "name": "value",
        "indexed": false
      }
    ],
    "anonymous": false,
    "type": "event"
  },
  {
    "name": "Append",
    "inputs": [
      {
        "type": "address",
        "name": "from",
        "indexed": true
      },
      {
        "type": "uint256",
        "name": "value",
        "indexed": false
      }
    ],
    "anonymous": false,
    "type": "event"
  },
  {
    "name": "register",
    "outputs": [],
    "inputs": [
      {
        "type": "uint256",
        "name": "value"
      }
    ],
    "constant": false,
    "payable": false,
    "type": "function"
  },
  {
    "name": "append",
    "outputs": [],
    "inputs": [
      {
        "type": "uint256",
        "name": "value"
      }
    ],
    "constant": false,
    "payable": false,
    "type": "function"
  },
  {
    "name": "getRelayerBalance",
    "outputs": [
      {
        "type": "uint256",
        "unit": "wei",
        "name": "registered"
      },
      {
        "type": "uint256",
        "unit": "wei",
        "name": "unregistering"
      },
      {
        "type": "uint256",
        "unit": "wei",
        "name": "unregistered"
      }
    ],
    "inputs": [
      {
        "type": "address",
        "name": "owner"
      }
    ],
    "constant": true,
    "payable": false,
    "type": "function"
  },
  {
    "name": "withdraw",
    "outputs": [],
    "inputs": [
      {
        "type": "uint256",
        "unit": "wei",
        "name": "value"
      }
    ],
    "constant": false,
    "payable": false,
    "type": "function"
  },
  {
    "name": "unregister",
    "outputs": [],
    "inputs": [
      {
        "type": "uint256",
        "unit": "wei",
        "name": "value"
      }
    ],
    "constant": false,
    "payable": false,
    "type": "function"
  },
  {
    "name": "getPeriodHeight",
    "outputs": [
      {
        "type": "uint256",
        "name": "start"
      },
      {
        "type": "uint256",
        "name": "end"
      },
      {
        "type": "bool",
        "name": "relayer"
      }
    ],
    "inputs": [
      {
        "type": "address",
        "name": "owner"
      }
    ],
    "constant": true,
    "payable": false,
    "type": "function"
  },
  {
    "name": "getRelayer",
    "inputs": [
      {
        "type": "address",
        "name": "owner"
      }
    ],
    "outputs": [
      {
        "type": "bool",
        "name": "register"
      },
      {
        "type": "bool",
        "name": "relayer"
      },
      {
        "type": "uint256",
        "name": "epoch"
      }
    ],
    "constant": true,
    "payable": false,
    "type": "function"
  }
]`
