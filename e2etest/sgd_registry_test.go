package e2etest

import (
	"context"
	"encoding/hex"
	"math/big"
	"sync/atomic"
	"testing"
	"time"

	"github.com/iotexproject/go-pkgs/crypto"
	"github.com/iotexproject/iotex-address/address"
	"github.com/iotexproject/iotex-core/action"
	"github.com/iotexproject/iotex-core/action/protocol"
	"github.com/iotexproject/iotex-core/action/protocol/account"
	accountutil "github.com/iotexproject/iotex-core/action/protocol/account/util"
	"github.com/iotexproject/iotex-core/action/protocol/execution"
	"github.com/iotexproject/iotex-core/action/protocol/rewarding"
	"github.com/iotexproject/iotex-core/action/protocol/rolldpos"
	"github.com/iotexproject/iotex-core/actpool"
	"github.com/iotexproject/iotex-core/blockchain"
	"github.com/iotexproject/iotex-core/blockchain/block"
	"github.com/iotexproject/iotex-core/blockchain/blockdao"
	"github.com/iotexproject/iotex-core/blockchain/genesis"
	"github.com/iotexproject/iotex-core/blockindex"
	"github.com/iotexproject/iotex-core/config"
	"github.com/iotexproject/iotex-core/db"
	"github.com/iotexproject/iotex-core/state/factory"
	"github.com/stretchr/testify/require"
)

func TestSGDRegistry(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()
	cfg := config.Default
	cfg.Chain.EnableAsyncIndexWrite = false
	cfg.Genesis.EnableGravityChainVoting = false
	cfg.Genesis.InitBalanceMap[_executor] = "1000000000000000000000000000"
	registry := protocol.NewRegistry()
	acc := account.NewProtocol(rewarding.DepositGas)
	r.NoError(acc.Register(registry))
	rp := rolldpos.NewProtocol(cfg.Genesis.NumCandidateDelegates, cfg.Genesis.NumDelegates, cfg.Genesis.NumSubEpochs)
	r.NoError(rp.Register(registry))
	factoryCfg := factory.GenerateConfig(cfg.Chain, cfg.Genesis)
	sf, err := factory.NewFactory(factoryCfg, db.NewMemKVStore(), factory.RegistryOption(registry))
	r.NoError(err)
	genericValidator := protocol.NewGenericValidator(sf, accountutil.AccountState)
	ap, err := actpool.NewActPool(cfg.Genesis, sf, cfg.ActPool)
	r.NoError(err)
	ap.AddActionEnvelopeValidators(genericValidator)
	dao := blockdao.NewBlockDAOInMemForTest([]blockdao.BlockIndexer{sf})
	bc := blockchain.NewBlockchain(
		cfg.Chain,
		cfg.Genesis,
		dao,
		factory.NewMinter(sf, ap),
		blockchain.BlockValidatorOption(block.NewValidator(
			sf,
			genericValidator,
		)),
	)
	r.NotNil(bc)
	reward := rewarding.NewProtocol(cfg.Genesis.Rewarding)
	r.NoError(reward.Register(registry))

	ep := execution.NewProtocol(dao.GetBlockHash, rewarding.DepositGasWithSGD, nil)
	r.NoError(ep.Register(registry))
	r.NoError(bc.Start(ctx))
	ctx = genesis.WithGenesisContext(ctx, cfg.Genesis)
	r.NoError(sf.Start(ctx))
	defer r.NoError(bc.Stop(ctx))

	_execPriKey, _ := crypto.HexStringToPrivateKey(_executorPriKey)
	data, _ := hex.DecodeString("608060405234801561001057600080fd5b5061002d61002261003260201b60201c565b61003a60201b60201c565b6100fe565b600033905090565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050816000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055508173ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff167f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e060405160405180910390a35050565b61108f8061010d6000396000f3fe608060405234801561001057600080fd5b50600436106100885760003560e01c8063a0ee93181161005b578063a0ee931814610101578063c375c2ef1461011d578063d7e5fbf314610139578063f2fde38b1461015557610088565b806307f7aafb1461008d5780630ad1c2fa146100a9578063715018a6146100d95780638da5cb5b146100e3575b600080fd5b6100a760048036038101906100a29190610bb9565b610171565b005b6100c360048036038101906100be9190610bb9565b6102fa565b6040516100d09190610e93565b60405180910390f35b6100e16103c4565b005b6100eb6103d8565b6040516100f89190610d6f565b60405180910390f35b61011b60048036038101906101169190610bb9565b610401565b005b61013760048036038101906101329190610bb9565b610589565b005b610153600480360381019061014e9190610be2565b610720565b005b61016f600480360381019061016a9190610bb9565b6109a4565b005b610179610a28565b6000600160008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000209050600073ffffffffffffffffffffffffffffffffffffffff168160000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff161415610250576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161024790610dd3565b60405180910390fd5b8060000160149054906101000a900460ff16156102a2576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161029990610e73565b60405180910390fd5b60018160000160146101000a81548160ff0219169083151502179055507faf42961ad755cade79794d4122cb0afedc32bf55a0c716dd085fbee2afc6ac55826040516102ee9190610d6f565b60405180910390a15050565b610302610b72565b600160008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206040518060400160405290816000820160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020016000820160149054906101000a900460ff1615151515815250509050919050565b6103cc610a28565b6103d66000610aa6565b565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905090565b610409610a28565b6000600160008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000209050600073ffffffffffffffffffffffffffffffffffffffff168160000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1614156104e0576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016104d790610dd3565b60405180910390fd5b8060000160149054906101000a900460ff16610531576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161052890610df3565b60405180910390fd5b60008160000160146101000a81548160ff0219169083151502179055507f97fd609c50722d17887170f33e7e8ae86421f650c02b82d6fe1e4ac9ef2842d48260405161057d9190610d6f565b60405180910390a15050565b610591610a28565b6000600160008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000209050600073ffffffffffffffffffffffffffffffffffffffff168160000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff161415610668576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161065f90610dd3565b60405180910390fd5b600160008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600080820160006101000a81549073ffffffffffffffffffffffffffffffffffffffff02191690556000820160146101000a81549060ff021916905550507f8d30d41865a0b811b9545d879520d2dde9f4cc49e4241f486ad9752bc904b565826040516107149190610d6f565b60405180910390a15050565b600073ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff161415610790576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161078790610e53565b60405180910390fd5b600073ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff161415610800576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016107f790610e13565b60405180910390fd5b600160008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060000160149054906101000a900460ff1615610890576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161088790610e73565b60405180910390fd5b60405180604001604052808273ffffffffffffffffffffffffffffffffffffffff16815260200160001515815250600160008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008201518160000160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060208201518160000160146101000a81548160ff0219169083151502179055509050507f768fb430a0d4b201cb764ab221c316dd14d8babf2e4b2348e05964c6565318b68282604051610998929190610d8a565b60405180910390a15050565b6109ac610a28565b600073ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff161415610a1c576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610a1390610db3565b60405180910390fd5b610a2581610aa6565b50565b610a30610b6a565b73ffffffffffffffffffffffffffffffffffffffff16610a4e6103d8565b73ffffffffffffffffffffffffffffffffffffffff1614610aa4576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610a9b90610e33565b60405180910390fd5b565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050816000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055508173ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff167f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e060405160405180910390a35050565b600033905090565b6040518060400160405280600073ffffffffffffffffffffffffffffffffffffffff1681526020016000151581525090565b600081359050610bb381611042565b92915050565b600060208284031215610bcb57600080fd5b6000610bd984828501610ba4565b91505092915050565b60008060408385031215610bf557600080fd5b6000610c0385828601610ba4565b9250506020610c1485828601610ba4565b9150509250929050565b610c2781610ebf565b82525050565b610c3681610ebf565b82525050565b610c4581610ed1565b82525050565b6000610c58602683610eae565b9150610c6382610efd565b604082019050919050565b6000610c7b601a83610eae565b9150610c8682610f4c565b602082019050919050565b6000610c9e601883610eae565b9150610ca982610f75565b602082019050919050565b6000610cc1602083610eae565b9150610ccc82610f9e565b602082019050919050565b6000610ce4602083610eae565b9150610cef82610fc7565b602082019050919050565b6000610d07601f83610eae565b9150610d1282610ff0565b602082019050919050565b6000610d2a601c83610eae565b9150610d3582611019565b602082019050919050565b604082016000820151610d566000850182610c1e565b506020820151610d696020850182610c3c565b50505050565b6000602082019050610d846000830184610c2d565b92915050565b6000604082019050610d9f6000830185610c2d565b610dac6020830184610c2d565b9392505050565b60006020820190508181036000830152610dcc81610c4b565b9050919050565b60006020820190508181036000830152610dec81610c6e565b9050919050565b60006020820190508181036000830152610e0c81610c91565b9050919050565b60006020820190508181036000830152610e2c81610cb4565b9050919050565b60006020820190508181036000830152610e4c81610cd7565b9050919050565b60006020820190508181036000830152610e6c81610cfa565b9050919050565b60006020820190508181036000830152610e8c81610d1d565b9050919050565b6000604082019050610ea86000830184610d40565b92915050565b600082825260208201905092915050565b6000610eca82610edd565b9050919050565b60008115159050919050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b7f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160008201527f6464726573730000000000000000000000000000000000000000000000000000602082015250565b7f436f6e7472616374206973206e6f742072656769737465726564000000000000600082015250565b7f436f6e7472616374206973206e6f7420617070726f7665640000000000000000600082015250565b7f526563697069656e7420616464726573732063616e6e6f74206265207a65726f600082015250565b7f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572600082015250565b7f436f6e747261637420616464726573732063616e6e6f74206265207a65726f00600082015250565b7f436f6e747261637420697320616c726561647920617070726f76656400000000600082015250565b61104b81610ebf565b811461105657600080fd5b5056fea264697066735822122078ad584b1d2be93e1239d9997b2a4be465fb7c6edc9d5302878b3dd92587f49464736f6c63430008010033")
	fixedTime := time.Unix(cfg.Genesis.Timestamp, 0)
	nonce := uint64(0)
	exec, err := action.SignedExecution(action.EmptyAddress, _execPriKey, atomic.AddUint64(&nonce, 1), big.NewInt(0), 10000000, big.NewInt(9000000000000), data)
	r.NoError(err)
	deployHash, err := exec.Hash()
	r.NoError(err)
	r.NoError(ap.Add(context.Background(), exec))
	blk, err := bc.MintNewBlock(fixedTime)
	r.NoError(err)
	r.NoError(bc.CommitBlock(blk))
	receipt, err := dao.GetReceiptByActionHash(deployHash, 1)
	r.NoError(err)
	r.Equal(receipt.ContractAddress, "io1va03q4lcr608dr3nltwm64sfcz05czjuycsqgn")
	height, err := dao.Height()
	r.NoError(err)
	r.Equal(uint64(1), height)

	contractAddress := receipt.ContractAddress
	kvstore := db.NewMemKVStore()
	sgdRegistry := blockindex.NewSGDRegistry(contractAddress, kvstore, 0)
	registerAddress, err := address.FromHex("5b38da6a701c568545dcfcb03fcb875f56beddc4")
	r.NoError(err)
	receiverAddress, err := address.FromHex("78731d3ca6b7e34ac0f824c42a7cc18a495cabab")
	r.NoError(err)
	t.Run("registerContract", func(t *testing.T) {
		data, _ = hex.DecodeString("d7e5fbf30000000000000000000000005b38da6a701c568545dcfcb03fcb875f56beddc400000000000000000000000078731d3ca6b7e34ac0f824c42a7cc18a495cabab")
		exec, err = action.SignedExecution(contractAddress, _execPriKey, atomic.AddUint64(&nonce, 1), big.NewInt(0), 10000000, big.NewInt(9000000000000), data)
		r.NoError(err)
		r.NoError(ap.Add(context.Background(), exec))
		blk, err = bc.MintNewBlock(fixedTime)
		r.NoError(err)
		r.NoError(bc.CommitBlock(blk))
		height, err = dao.Height()
		r.NoError(err)
		r.Equal(uint64(2), height)

		ctx = genesis.WithGenesisContext(
			protocol.WithBlockchainCtx(
				protocol.WithRegistry(ctx, registry),
				protocol.BlockchainCtx{
					Tip: protocol.TipInfo{
						Height:    height,
						Hash:      blk.HashHeader(),
						Timestamp: blk.Timestamp(),
					},
				}),
			cfg.Genesis,
		)
		r.NoError(sgdRegistry.PutBlock(ctx, blk))
		receiver, percentage, isApproved, err := sgdRegistry.CheckContract(ctx, registerAddress.String())
		r.NoError(err)
		r.Equal(uint64(20), percentage)
		r.Equal(receiverAddress, receiver)
		r.False(isApproved)
	})
	t.Run("approveContract", func(t *testing.T) {
		data, _ = hex.DecodeString("07f7aafb0000000000000000000000005b38da6a701c568545dcfcb03fcb875f56beddc4")
		exec, err = action.SignedExecution(contractAddress, _execPriKey, atomic.AddUint64(&nonce, 1), big.NewInt(0), 10000000, big.NewInt(9000000000000), data)
		r.NoError(err)
		r.NoError(ap.Add(context.Background(), exec))
		blk, err = bc.MintNewBlock(fixedTime)
		r.NoError(err)
		r.NoError(bc.CommitBlock(blk))
		height, err = dao.Height()
		r.NoError(err)
		r.Equal(uint64(3), height)

		ctx = genesis.WithGenesisContext(
			protocol.WithBlockchainCtx(
				protocol.WithRegistry(ctx, registry),
				protocol.BlockchainCtx{
					Tip: protocol.TipInfo{
						Height:    height,
						Hash:      blk.HashHeader(),
						Timestamp: blk.Timestamp(),
					},
				}),
			cfg.Genesis,
		)
		r.NoError(sgdRegistry.PutBlock(ctx, blk))
		receiver, percentage, isApproved, err := sgdRegistry.CheckContract(ctx, registerAddress.String())
		r.NoError(err)
		r.Equal(receiverAddress, receiver)
		r.True(isApproved)
		r.Equal(uint64(20), percentage)
	})

	t.Run("disapproveContract", func(t *testing.T) {
		data, _ = hex.DecodeString("a0ee93180000000000000000000000005b38da6a701c568545dcfcb03fcb875f56beddc4")
		exec, err = action.SignedExecution(contractAddress, _execPriKey, atomic.AddUint64(&nonce, 1), big.NewInt(0), 10000000, big.NewInt(9000000000000), data)
		r.NoError(err)
		r.NoError(ap.Add(context.Background(), exec))
		blk, err = bc.MintNewBlock(fixedTime)
		r.NoError(err)
		r.NoError(bc.CommitBlock(blk))
		height, err = dao.Height()
		r.NoError(err)
		r.Equal(uint64(4), height)

		ctx = genesis.WithGenesisContext(
			protocol.WithBlockchainCtx(
				protocol.WithRegistry(ctx, registry),
				protocol.BlockchainCtx{
					Tip: protocol.TipInfo{
						Height:    height,
						Hash:      blk.HashHeader(),
						Timestamp: blk.Timestamp(),
					},
				}),
			cfg.Genesis,
		)
		r.NoError(sgdRegistry.PutBlock(ctx, blk))
		receiver, percentage, isApproved, err := sgdRegistry.CheckContract(ctx, registerAddress.String())
		r.NoError(err)
		r.Equal(receiverAddress, receiver)
		r.False(isApproved)
		r.Equal(uint64(20), percentage)
	})

	t.Run("removeContract", func(t *testing.T) {
		data, _ = hex.DecodeString("c375c2ef0000000000000000000000005b38da6a701c568545dcfcb03fcb875f56beddc4")
		exec, err = action.SignedExecution(contractAddress, _execPriKey, atomic.AddUint64(&nonce, 1), big.NewInt(0), 10000000, big.NewInt(9000000000000), data)
		r.NoError(err)
		r.NoError(ap.Add(context.Background(), exec))
		blk, err = bc.MintNewBlock(fixedTime)
		r.NoError(err)
		r.NoError(bc.CommitBlock(blk))
		height, err = dao.Height()
		r.NoError(err)
		r.Equal(uint64(5), height)

		ctx = genesis.WithGenesisContext(
			protocol.WithBlockchainCtx(
				protocol.WithRegistry(ctx, registry),
				protocol.BlockchainCtx{
					Tip: protocol.TipInfo{
						Height:    height,
						Hash:      blk.HashHeader(),
						Timestamp: blk.Timestamp(),
					},
				}),
			cfg.Genesis,
		)
		r.NoError(sgdRegistry.PutBlock(ctx, blk))
		receiver, percentage, isApproved, err := sgdRegistry.CheckContract(ctx, registerAddress.String())
		r.ErrorContains(err, "not exist in DB")
		r.Nil(receiver)
		r.False(isApproved)
		r.Equal(uint64(0), percentage)
	})
}
