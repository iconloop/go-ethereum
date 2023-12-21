package vm

import (
	"math/big"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/params"
)

type frameHandler struct {
	FrameHandler
	callFn func(from ContractRef, to common.Address, input []byte, gas uint64, value *big.Int) (ret []byte, gasLeft uint64, err error)
}

func (f frameHandler) Call(from ContractRef, to common.Address, input []byte, gas uint64, value *big.Int) (ret []byte, gasLeft uint64, err error) {
	return f.callFn(from, to, input, gas, value)
}

func CanTransfer(db StateDB, addr common.Address, amount *big.Int) bool {
	return db.GetBalance(addr).Cmp(amount) >= 0
}

func Transfer(db StateDB, sender, recipient common.Address, amount *big.Int) {
	db.SubBalance(sender, amount)
	db.AddBalance(recipient, amount)
}

func asByteSl(opc ...OpCode) []byte {
	var bs []byte
	for _, op := range opc {
		bs = append(bs, byte(op))
	}
	return bs
}

func TestEVM_FrameHandler(t *testing.T) {
	blkCtx := BlockContext{
		CanTransfer: CanTransfer,
		Transfer:    Transfer,
		GetHash:     nil,
	}
	db := state.NewDatabase(rawdb.NewDatabase(memorydb.New()))
	stateDB, err := state.New(common.Hash{}, db, nil)
	assert.NoError(t, err)
	abort := atomic.Bool{}
	fh := frameHandler{}
	var callTarget common.Address
	fh.callFn = func(from ContractRef, to common.Address, input []byte, gas uint64, value *big.Int) (ret []byte, gasLeft uint64, err error) {
		callTarget = to
		return nil, 0, nil
	}
	evm := NewEVM2(blkCtx, TxContext{}, stateDB, params.MainnetChainConfig, Config{}, fh, &abort)
	assert.NotNil(t, evm)
	init := asByteSl(
		//
		// Constructor
		//

		// Size
		PUSH1, 26,
		// Offset
		PUSH1, 16,
		// DstOffset
		PUSH3, 0x2, 0, 0,
		CODECOPY,
		// RetSize
		PUSH1, 26,
		// RetOffset
		PUSH3, 0x2, 0, 0,
		RETURN,

		//
		// Deployed Code
		//

		// Mem[0] = 0
		PUSH1, 0x00,
		PUSH1, 0x00,
		MSTORE,

		// RetSize
		PUSH1, 0x00,
		// RetOffset
		PUSH1, 0x00,
		// ArgsSize
		PUSH1, 0x20,
		// ArgsOffset
		PUSH1, 0x00,
		// Value
		PUSH1, 0x05,
		// Address
		PUSH2, 0x1, 0x00,
		// Gas
		PUSH1, 0xff,
		CALL,

		// RetSize
		PUSH1, 0x00,
		// RetOffset
		PUSH1, 0x00,
		RETURN,
	)
	ret, addr, _, err := evm.Create(contractRef{}, init, 1_000_000, common.Big0)
	assert.NoError(t, err)
	assert.Equal(t, init[len(init)-26:], ret)
	ret, _, err = evm.Call(contractRef{}, addr, []byte{0x1, 0x2}, 1_000_000, common.Big0)
	assert.NoError(t, err)
	assert.Equal(t, common.BigToAddress(new(big.Int).SetUint64(0x100)), callTarget)
}
