package vm

import (
	"math/big"
	"sync/atomic"

	"github.com/holiman/uint256"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

type FrameHandler interface {
	Call(from ContractRef, to common.Address, input []byte, gas uint64, value *big.Int) (ret []byte, gasLeft uint64, err error)
	CallCode(from ContractRef, to common.Address, input []byte, gas uint64, value *big.Int) (ret []byte, gasLeft uint64, err error)
	DelegateCall(from ContractRef, to common.Address, input []byte, gas uint64) (ret []byte, gasLeft uint64, err error)
	StaticCall(from ContractRef, to common.Address, input []byte, gas uint64) (ret []byte, gasLeft uint64, err error)
	Create(from ContractRef, code []byte, gas uint64, value *big.Int) (ret []byte, contractAddr common.Address, gasLeft uint64, err error)
	Create2(from ContractRef, code []byte, gas uint64, endowment *big.Int, salt *uint256.Int) (ret []byte, contractAddr common.Address, gasLeft uint64, err error)
}

func (evm *EVM) SetFrameHandler(fh FrameHandler) {
	evm.fh = fh
}

func (evm *EVM) SetAbortPtr(abortPtr *atomic.Bool) {
	evm.abort = abortPtr
}

func NewEVM2(blockCtx BlockContext, txCtx TxContext, statedb StateDB, chainConfig *params.ChainConfig, config Config, fh FrameHandler, abortPtr *atomic.Bool) *EVM {
	evm := NewEVM(blockCtx, txCtx, statedb, chainConfig, config)
	evm.SetFrameHandler(fh)
	evm.SetAbortPtr(abortPtr)
	return evm
}
