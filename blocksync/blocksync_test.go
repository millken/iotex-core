// Copyright (c) 2019 IoTeX Foundation
// This source code is provided 'as is' and no warranties are given as to title or non-infringement, merchantability
// or fitness for purpose and, to the extent permitted by law, all liability for your use of the code is disclaimed.
// This source code is governed by Apache License 2.0 that can be found in the LICENSE file.

package blocksync

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/iotexproject/go-pkgs/hash"
	"github.com/iotexproject/iotex-core/action/protocol"
	"github.com/iotexproject/iotex-core/action/protocol/account"
	accountutil "github.com/iotexproject/iotex-core/action/protocol/account/util"
	"github.com/iotexproject/iotex-core/action/protocol/rewarding"
	"github.com/iotexproject/iotex-core/action/protocol/rolldpos"
	"github.com/iotexproject/iotex-core/actpool"
	"github.com/iotexproject/iotex-core/blockchain"
	"github.com/iotexproject/iotex-core/blockchain/block"
	"github.com/iotexproject/iotex-core/blockchain/blockdao"
	"github.com/iotexproject/iotex-core/blockchain/genesis"
	"github.com/iotexproject/iotex-core/consensus"
	"github.com/iotexproject/iotex-core/db"
	"github.com/iotexproject/iotex-core/state/factory"
	"github.com/iotexproject/iotex-core/test/identityset"
	"github.com/iotexproject/iotex-core/test/mock/mock_blockchain"
	"github.com/iotexproject/iotex-core/test/mock/mock_blockdao"
	"github.com/iotexproject/iotex-core/test/mock/mock_blocksync"
	"github.com/iotexproject/iotex-core/test/mock/mock_consensus"
	"github.com/iotexproject/iotex-core/testutil"
	goproto "github.com/iotexproject/iotex-proto/golang"
	"github.com/iotexproject/iotex-proto/golang/iotexrpc"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

type testConfig struct {
	BlockSync Config
	Genesis   genesis.Genesis
	Chain     blockchain.Config
	ActPool   actpool.Config
}

func newBlockSyncerForTest(cfg Config, chain blockchain.Blockchain, dao blockdao.BlockDAO, cs consensus.Consensus) (*blockSyncer, error) {
	bs, err := NewBlockSyncer(cfg, chain.TipHeight,
		func(h uint64) (*block.Block, error) {
			return dao.GetBlockByHeight(h)
		},
		func(blk *block.Block) error {
			if err := cs.ValidateBlockFooter(blk); err != nil {
				return err
			}
			if err := chain.ValidateBlock(blk); err != nil {
				return err
			}
			if err := chain.CommitBlock(blk); err != nil {
				return err
			}
			cs.Calibrate(blk.Height())
			return nil
		},
		func() ([]peer.AddrInfo, error) {
			return []peer.AddrInfo{}, nil
		},
		func(context.Context, peer.AddrInfo, proto.Message) error {
			return nil
		},
		func(string) {
			return
		},
	)
	if err != nil {
		return nil, err
	}
	return bs.(*blockSyncer), nil
}

func TestNewBlockSyncer(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	ctrl := gomock.NewController(t)

	mBc := mock_blockchain.NewMockBlockchain(ctrl)
	// TipHeight return ERROR
	mBc.EXPECT().TipHeight().AnyTimes().Return(uint64(0))
	mBc.EXPECT().ChainID().AnyTimes().Return(blockchain.DefaultConfig.ID)
	blk := block.NewBlockDeprecated(
		uint32(123),
		uint64(0),
		hash.Hash256{},
		testutil.TimestampNow(),
		identityset.PrivateKey(27).PublicKey(),
		nil,
	)
	dao := mock_blockdao.NewMockBlockDAO(ctrl)
	dao.EXPECT().GetBlockByHeight(gomock.Any()).AnyTimes().Return(blk, nil)

	cfg, err := newTestConfig()
	require.NoError(err)

	cs := mock_consensus.NewMockConsensus(ctrl)

	bs, err := newBlockSyncerForTest(cfg.BlockSync, mBc, dao, cs)
	assert.Nil(err)
	assert.NotNil(bs)
}

func TestBlockSyncerStart(t *testing.T) {
	assert := assert.New(t)

	ctrl := gomock.NewController(t)

	ctx := context.Background()
	mBs := mock_blocksync.NewMockBlockSync(ctrl)
	mBs.EXPECT().Start(gomock.Any()).Times(1)
	assert.Nil(mBs.Start(ctx))
}

func TestBlockSyncerStop(t *testing.T) {
	assert := assert.New(t)

	ctrl := gomock.NewController(t)

	ctx := context.Background()
	mBs := mock_blocksync.NewMockBlockSync(ctrl)
	mBs.EXPECT().Stop(gomock.Any()).Times(1)
	assert.Nil(mBs.Stop(ctx))
}

func TestBlockSyncerProcessSyncRequest(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	ctrl := gomock.NewController(t)

	mBc := mock_blockchain.NewMockBlockchain(ctrl)
	mBc.EXPECT().ChainID().AnyTimes().Return(blockchain.DefaultConfig.ID)
	blk := block.NewBlockDeprecated(
		uint32(123),
		uint64(0),
		hash.Hash256{},
		testutil.TimestampNow(),
		identityset.PrivateKey(27).PublicKey(),
		nil,
	)
	dao := mock_blockdao.NewMockBlockDAO(ctrl)
	dao.EXPECT().GetBlockByHeight(gomock.Any()).AnyTimes().Return(blk, nil)
	mBc.EXPECT().TipHeight().AnyTimes().Return(uint64(0))
	cfg, err := newTestConfig()
	require.NoError(err)
	cs := mock_consensus.NewMockConsensus(ctrl)

	bs, err := newBlockSyncerForTest(cfg.BlockSync, mBc, dao, cs)
	assert.NoError(err)
	assert.NoError(bs.ProcessSyncRequest(context.Background(), peer.AddrInfo{}, 1, 1))
}

func TestBlockSyncerProcessSyncRequestError(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)

	cfg, err := newTestConfig()
	require.NoError(err)

	chain := mock_blockchain.NewMockBlockchain(ctrl)
	dao := mock_blockdao.NewMockBlockDAO(ctrl)
	dao.EXPECT().GetBlockByHeight(uint64(1)).Return(nil, errors.New("some error")).Times(1)
	chain.EXPECT().ChainID().Return(uint32(1)).AnyTimes()
	chain.EXPECT().TipHeight().Return(uint64(10)).Times(1)
	cs := mock_consensus.NewMockConsensus(ctrl)

	bs, err := newBlockSyncerForTest(cfg.BlockSync, chain, dao, cs)
	require.NoError(err)

	require.Error(bs.ProcessSyncRequest(context.Background(), peer.AddrInfo{}, 1, 5))
}

func TestBlockSyncerProcessBlockTipHeight(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)

	ctx := context.Background()
	cfg, err := newTestConfig()
	require.NoError(err)
	registry := protocol.NewRegistry()
	acc := account.NewProtocol(rewarding.DepositGas)
	require.NoError(acc.Register(registry))
	rp := rolldpos.NewProtocol(cfg.Genesis.NumCandidateDelegates, cfg.Genesis.NumDelegates, cfg.Genesis.NumSubEpochs)
	require.NoError(rp.Register(registry))
	factoryCfg := factory.GenerateConfig(cfg.Chain, cfg.Genesis)
	sf, err := factory.NewFactory(factoryCfg, db.NewMemKVStore(), factory.RegistryOption(registry))
	require.NoError(err)
	ap, err := actpool.NewActPool(cfg.Genesis, sf, cfg.ActPool)
	require.NotNil(ap)
	require.NoError(err)
	ap.AddActionEnvelopeValidators(protocol.NewGenericValidator(sf, accountutil.AccountState))
	dao := blockdao.NewBlockDAOInMemForTest([]blockdao.BlockIndexer{sf})
	chain := blockchain.NewBlockchain(
		cfg.Chain,
		cfg.Genesis,
		dao,
		factory.NewMinter(sf, ap),
		blockchain.BlockValidatorOption(block.NewValidator(sf, ap)),
	)
	require.NoError(chain.Start(ctx))
	require.NotNil(chain)
	cs := mock_consensus.NewMockConsensus(ctrl)
	cs.EXPECT().ValidateBlockFooter(gomock.Any()).Return(nil).Times(1)
	cs.EXPECT().Calibrate(uint64(1)).Times(1)

	bs, err := newBlockSyncerForTest(cfg.BlockSync, chain, dao, cs)
	require.NoError(err)

	defer func() {
		require.NoError(chain.Stop(ctx))
	}()

	h := chain.TipHeight()
	blk, err := chain.MintNewBlock(testutil.TimestampNow())
	require.NotNil(blk)
	require.NoError(err)
	ctx, err = chain.Context(ctx)
	require.NoError(err)

	peer := "peer1"

	require.NoError(bs.ProcessBlock(ctx, peer, blk))
	h2 := chain.TipHeight()
	assert.Equal(t, h+1, h2)

	// commit top
	require.NoError(bs.ProcessBlock(ctx, peer, blk))
	h3 := chain.TipHeight()
	assert.Equal(t, h+1, h3)

	// commit same block again
	require.NoError(bs.ProcessBlock(ctx, peer, blk))
	h4 := chain.TipHeight()
	assert.Equal(t, h3, h4)
}

func TestBlockSyncerProcessBlockOutOfOrder(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)

	ctx := context.Background()
	cfg, err := newTestConfig()
	require.NoError(err)
	registry := protocol.NewRegistry()
	acc := account.NewProtocol(rewarding.DepositGas)
	require.NoError(acc.Register(registry))
	rp := rolldpos.NewProtocol(cfg.Genesis.NumCandidateDelegates, cfg.Genesis.NumDelegates, cfg.Genesis.NumSubEpochs)
	require.NoError(rp.Register(registry))
	factoryCfg := factory.GenerateConfig(cfg.Chain, cfg.Genesis)
	sf, err := factory.NewFactory(factoryCfg, db.NewMemKVStore(), factory.RegistryOption(registry))
	require.NoError(err)
	ap1, err := actpool.NewActPool(cfg.Genesis, sf, cfg.ActPool)
	require.NotNil(ap1)
	require.NoError(err)
	ap1.AddActionEnvelopeValidators(protocol.NewGenericValidator(sf, accountutil.AccountState))
	dao := blockdao.NewBlockDAOInMemForTest([]blockdao.BlockIndexer{sf})
	chain1 := blockchain.NewBlockchain(
		cfg.Chain,
		cfg.Genesis,
		dao,
		factory.NewMinter(sf, ap1),
		blockchain.BlockValidatorOption(block.NewValidator(sf, ap1)),
	)
	require.NotNil(chain1)
	require.NoError(chain1.Start(ctx))
	cs1 := mock_consensus.NewMockConsensus(ctrl)
	cs1.EXPECT().ValidateBlockFooter(gomock.Any()).Return(nil).Times(3)
	cs1.EXPECT().Calibrate(gomock.Any()).Times(3)

	bs1, err := newBlockSyncerForTest(cfg.BlockSync, chain1, dao, cs1)
	require.NoError(err)
	registry2 := protocol.NewRegistry()
	require.NoError(acc.Register(registry2))
	require.NoError(rp.Register(registry2))
	sf2, err := factory.NewFactory(factoryCfg, db.NewMemKVStore(), factory.RegistryOption(registry2))
	require.NoError(err)
	ap2, err := actpool.NewActPool(cfg.Genesis, sf2, cfg.ActPool)
	require.NotNil(ap2)
	require.NoError(err)
	ap2.AddActionEnvelopeValidators(protocol.NewGenericValidator(sf2, accountutil.AccountState))
	dao2 := blockdao.NewBlockDAOInMemForTest([]blockdao.BlockIndexer{sf2})
	chain2 := blockchain.NewBlockchain(
		cfg.Chain,
		cfg.Genesis,
		dao2,
		factory.NewMinter(sf2, ap2),
		blockchain.BlockValidatorOption(block.NewValidator(sf2, ap2)),
	)
	require.NotNil(chain2)
	require.NoError(chain2.Start(ctx))
	cs2 := mock_consensus.NewMockConsensus(ctrl)
	cs2.EXPECT().ValidateBlockFooter(gomock.Any()).Return(nil).Times(3)
	cs2.EXPECT().Calibrate(gomock.Any()).Times(3)
	bs2, err := newBlockSyncerForTest(cfg.BlockSync, chain2, dao2, cs2)
	require.NoError(err)

	defer func() {
		require.NoError(chain1.Stop(ctx))
		require.NoError(chain2.Stop(ctx))
	}()

	// commit top
	ctx, err = chain1.Context(ctx)
	require.NoError(err)

	peer := "peer1"

	blk1, err := chain1.MintNewBlock(testutil.TimestampNow())
	require.NotNil(blk1)
	require.NoError(err)
	require.NoError(bs1.ProcessBlock(ctx, peer, blk1))
	blk2, err := chain1.MintNewBlock(testutil.TimestampNow())
	require.NotNil(blk2)
	require.NoError(err)
	require.NoError(bs1.ProcessBlock(ctx, peer, blk2))
	blk3, err := chain1.MintNewBlock(testutil.TimestampNow())
	require.NotNil(blk3)
	require.NoError(err)
	require.NoError(bs1.ProcessBlock(ctx, peer, blk3))
	h1 := chain1.TipHeight()
	assert.Equal(t, uint64(3), h1)

	require.NoError(bs2.ProcessBlock(ctx, peer, blk3))
	require.NoError(bs2.ProcessBlock(ctx, peer, blk2))
	require.NoError(bs2.ProcessBlock(ctx, peer, blk2))
	require.NoError(bs2.ProcessBlock(ctx, peer, blk1))
	h2 := chain2.TipHeight()
	assert.Equal(t, h1, h2)
}

func TestBlockSyncerProcessBlock(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)

	ctx := context.Background()
	cfg, err := newTestConfig()
	require.NoError(err)
	registry := protocol.NewRegistry()
	acc := account.NewProtocol(rewarding.DepositGas)
	require.NoError(acc.Register(registry))
	rolldposProtocol := rolldpos.NewProtocol(
		cfg.Genesis.NumCandidateDelegates,
		cfg.Genesis.NumDelegates,
		cfg.Genesis.NumSubEpochs,
	)
	require.NoError(rolldposProtocol.Register(registry))
	factoryCfg := factory.GenerateConfig(cfg.Chain, cfg.Genesis)
	sf, err := factory.NewFactory(factoryCfg, db.NewMemKVStore(), factory.RegistryOption(registry))
	require.NoError(err)
	ap1, err := actpool.NewActPool(cfg.Genesis, sf, cfg.ActPool)
	require.NotNil(ap1)
	require.NoError(err)
	ap1.AddActionEnvelopeValidators(protocol.NewGenericValidator(sf, accountutil.AccountState))
	dao := blockdao.NewBlockDAOInMemForTest([]blockdao.BlockIndexer{sf})
	chain1 := blockchain.NewBlockchain(
		cfg.Chain,
		cfg.Genesis,
		dao,
		factory.NewMinter(sf, ap1),
		blockchain.BlockValidatorOption(block.NewValidator(sf, ap1)),
	)
	require.NoError(chain1.Start(ctx))
	require.NotNil(chain1)
	cs1 := mock_consensus.NewMockConsensus(ctrl)
	cs1.EXPECT().ValidateBlockFooter(gomock.Any()).Return(nil).Times(3)
	cs1.EXPECT().Calibrate(gomock.Any()).Times(3)
	bs1, err := newBlockSyncerForTest(cfg.BlockSync, chain1, dao, cs1)
	require.NoError(err)
	registry2 := protocol.NewRegistry()
	require.NoError(acc.Register(registry2))
	require.NoError(rolldposProtocol.Register(registry2))
	sf2, err := factory.NewFactory(factoryCfg, db.NewMemKVStore(), factory.RegistryOption(registry2))
	require.NoError(err)
	ap2, err := actpool.NewActPool(cfg.Genesis, sf2, cfg.ActPool)
	require.NotNil(ap2)
	require.NoError(err)
	ap2.AddActionEnvelopeValidators(protocol.NewGenericValidator(sf2, accountutil.AccountState))
	dao2 := blockdao.NewBlockDAOInMemForTest([]blockdao.BlockIndexer{sf2})
	chain2 := blockchain.NewBlockchain(
		cfg.Chain,
		cfg.Genesis,
		dao2,
		factory.NewMinter(sf2, ap2),
		blockchain.BlockValidatorOption(block.NewValidator(sf2, ap2)),
	)
	require.NoError(chain2.Start(ctx))
	require.NotNil(chain2)
	cs2 := mock_consensus.NewMockConsensus(ctrl)
	cs2.EXPECT().ValidateBlockFooter(gomock.Any()).Return(nil).Times(3)
	cs2.EXPECT().Calibrate(gomock.Any()).Times(3)
	bs2, err := newBlockSyncerForTest(cfg.BlockSync, chain2, dao2, cs2)
	require.NoError(err)

	defer func() {
		require.NoError(chain1.Stop(ctx))
		require.NoError(chain2.Stop(ctx))
	}()

	ctx, err = chain1.Context(ctx)
	require.NoError(err)

	peer := "peer1"

	// commit top
	blk1, err := chain1.MintNewBlock(testutil.TimestampNow())
	require.NotNil(blk1)
	require.NoError(err)
	require.NoError(bs1.ProcessBlock(ctx, peer, blk1))
	blk2, err := chain1.MintNewBlock(testutil.TimestampNow())
	require.NotNil(blk2)
	require.NoError(err)
	require.NoError(bs1.ProcessBlock(ctx, peer, blk2))
	blk3, err := chain1.MintNewBlock(testutil.TimestampNow())
	require.NotNil(blk3)
	require.NoError(err)
	require.NoError(bs1.ProcessBlock(ctx, peer, blk3))
	h1 := chain1.TipHeight()
	assert.Equal(t, uint64(3), h1)

	require.NoError(bs2.ProcessBlock(ctx, peer, blk2))
	require.NoError(bs2.ProcessBlock(ctx, peer, blk3))
	require.NoError(bs2.ProcessBlock(ctx, peer, blk1))
	h2 := chain2.TipHeight()
	assert.Equal(t, h1, h2)
}

func TestBlockSyncerSync(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)

	ctx := context.Background()
	cfg, err := newTestConfig()
	require.NoError(err)
	registry := protocol.NewRegistry()
	rp := rolldpos.NewProtocol(cfg.Genesis.NumCandidateDelegates, cfg.Genesis.NumDelegates, cfg.Genesis.NumSubEpochs)
	require.NoError(rp.Register(registry))
	factoryCfg := factory.GenerateConfig(cfg.Chain, cfg.Genesis)
	sf, err := factory.NewFactory(factoryCfg, db.NewMemKVStore(), factory.RegistryOption(registry))
	require.NoError(err)
	ap, err := actpool.NewActPool(cfg.Genesis, sf, cfg.ActPool)
	require.NotNil(ap)
	require.NoError(err)
	ap.AddActionEnvelopeValidators(protocol.NewGenericValidator(sf, accountutil.AccountState))
	dao := blockdao.NewBlockDAOInMemForTest([]blockdao.BlockIndexer{sf})
	chain := blockchain.NewBlockchain(
		cfg.Chain,
		cfg.Genesis,
		dao,
		factory.NewMinter(sf, ap),
		blockchain.BlockValidatorOption(block.NewValidator(sf, ap)),
	)
	require.NoError(chain.Start(ctx))
	require.NotNil(chain)
	cs := mock_consensus.NewMockConsensus(ctrl)
	cs.EXPECT().ValidateBlockFooter(gomock.Any()).Return(nil).Times(2)
	cs.EXPECT().Calibrate(gomock.Any()).Times(2)

	bs, err := newBlockSyncerForTest(cfg.BlockSync, chain, dao, cs)
	require.NoError(err)
	require.NotNil(bs)
	require.NoError(bs.Start(ctx))
	time.Sleep(time.Millisecond << 7)

	defer func() {
		require.NoError(bs.Stop(ctx))
		require.NoError(chain.Stop(ctx))
	}()

	ctx, err = chain.Context(ctx)
	require.NoError(err)

	peer := "peer1"

	blk, err := chain.MintNewBlock(testutil.TimestampNow())
	require.NotNil(blk)
	require.NoError(err)
	require.NoError(bs.ProcessBlock(ctx, peer, blk))

	blk, err = chain.MintNewBlock(testutil.TimestampNow())
	require.NotNil(blk)
	require.NoError(err)
	require.NoError(bs.ProcessBlock(ctx, peer, blk))
	time.Sleep(time.Millisecond << 7)
}

func newTestConfig() (testConfig, error) {
	testTriePath, err := testutil.PathOfTempFile("trie")
	if err != nil {
		return testConfig{}, err
	}
	testDBPath, err := testutil.PathOfTempFile("db")
	if err != nil {
		return testConfig{}, err
	}
	defer func() {
		testutil.CleanupPath(testTriePath)
		testutil.CleanupPath(testDBPath)
	}()

	cfg := testConfig{
		BlockSync: DefaultConfig,
		Genesis:   genesis.Default,
		Chain:     blockchain.DefaultConfig,
		ActPool:   actpool.DefaultConfig,
	}
	cfg.Chain.TrieDBPath = testTriePath
	cfg.Chain.ChainDBPath = testDBPath
	cfg.BlockSync.Interval = 100 * time.Millisecond
	cfg.Genesis.EnableGravityChainVoting = false
	return cfg, nil
}

func TestDummyBlockSync(t *testing.T) {
	require := require.New(t)
	bs := NewDummyBlockSyncer()
	require.NoError(bs.Start(nil))
	require.NoError(bs.Stop(nil))
	require.NoError(bs.ProcessBlock(nil, "", nil))
	require.NoError(bs.ProcessSyncRequest(nil, peer.AddrInfo{}, 0, 0))
	require.Equal(bs.TargetHeight(), uint64(0))
	startingHeight, currentHeight, targetHeight, desc := bs.SyncStatus()
	require.Zero(startingHeight)
	require.Zero(currentHeight)
	require.Zero(targetHeight)
	require.Empty(desc)
}

type test1 struct {
	syncBlockIncrease uint64
	startingHeight    uint64
	targetHeight      uint64
	mu                sync.RWMutex
}

func (bs *test1) sync() {
	for i := 0; i >= 0; i++ {
		time.Sleep(time.Millisecond * 2)
		bs.mu.Lock()
		atomic.StoreUint64(&bs.syncBlockIncrease, uint64(i))
		bs.targetHeight = uint64(i)
		bs.startingHeight = uint64(i)
		bs.mu.Unlock()
	}
}
func (bs *test1) tipHeightHandler() uint64 {
	return 1
}
func (bs *test1) SyncStatus() (uint64, uint64, uint64, string) {
	// bs.mu.RLock()
	// defer bs.mu.RUnlock()
	var syncSpeedDesc string
	syncBlockIncrease := atomic.LoadUint64(&bs.syncBlockIncrease)
	switch {
	case syncBlockIncrease == 1:
		syncSpeedDesc = "synced to blockchain tip"
	default:
		syncSpeedDesc = fmt.Sprintf("sync in progress at %.1f blocks/sec", float64(syncBlockIncrease))
	}
	return bs.startingHeight, bs.tipHeightHandler(), bs.targetHeight, syncSpeedDesc
}

func (bs *test1) BuildReport() string {
	startingHeight, tipHeight, targetHeight, syncSpeedDesc := bs.SyncStatus()
	return fmt.Sprintf(
		"BlockSync startingHeight: %d, tipHeight: %d, targetHeight: %d, %s",
		startingHeight,
		tipHeight,
		targetHeight,
		syncSpeedDesc,
	)
}

func TestBlockSyncerBugIssue3889(t *testing.T) {
	bs := &test1{}
	go bs.sync()
	time.Sleep(time.Millisecond * 10)
	wait := sync.WaitGroup{}
	wait.Add(5)
	for i := 0; i < 5; i++ {
		go func(i int) {
			defer wait.Done()
			startingHeight, tipHeight, targetHeight, syncSpeedDesc := bs.SyncStatus()
			t.Logf("BlockSync startingHeight: %d, tipHeight: %d, targetHeight: %d, %s", startingHeight, tipHeight, targetHeight, syncSpeedDesc)
			report := bs.BuildReport()
			t.Log(report)
		}(i)
	}
	wait.Wait()
}

/*
RateLimitInterval = 1 sec, syncd 9600 blocks in 60 sec, 160 blocks/sec
RateLimitInterval = 2 sec, syncd 4800 blocks in 60 sec, 80 blocks/sec
*/
func TestBlockSync(t *testing.T) {
	//t.SkipNow()
	require := require.New(t)
	cfg := DefaultConfig
	cfg.Interval = 10 * time.Second
	cfg.BufferSize = 200
	cfg.MaxRepeat = 3
	cfg.RepeatDecayStep = 3
	cfg.RateLimitInterval = 2 * time.Second
	var tipHeight uint64
	chanProcessBlk := make(chan *block.Block)

	bs, err := NewBlockSyncer(cfg, func() uint64 {
		return atomic.LoadUint64(&tipHeight)
	}, nil, func(b *block.Block) error {
		//t.Logf("commit block %d", b.Height())
		atomic.StoreUint64(&tipHeight, b.Height())
		return nil
	}, func() ([]peer.AddrInfo, error) {
		return []peer.AddrInfo{{
			ID: peer.ID("test"),
		}}, nil
	}, func(ctx context.Context, ai peer.AddrInfo, m proto.Message) error {
		msgType, err := goproto.GetTypeFromRPCMsg(m)
		if err != nil {
			return err
		}
		switch msgType {
		case iotexrpc.MessageType_BLOCK_REQUEST:
			msgBody, err := proto.Marshal(m)
			if err != nil {
				return err
			}
			var req iotexrpc.BlockSync
			if err := proto.Unmarshal(msgBody, &req); err != nil {
				return err
			}
			// send block
			for i := req.Start; i <= req.End; i++ {
				blk, _ := block.NewTestingBuilder().
					SetHeight(i).SignAndBuild(identityset.PrivateKey(0))
				chanProcessBlk <- &blk
			}
		}
		return nil
	}, func(s string) {
		t.Logf("block p2p %s", s)
	})
	ctx := context.Background()
	require.NoError(err)
	require.NotNil(bs)
	go func(bs BlockSync) {
		for blk := range chanProcessBlk {
			bs.ProcessBlock(ctx, "test", blk)
		}
	}(bs)
	go func(bs BlockSync) {
		for {
			blk, _ := block.NewTestingBuilder().
				SetHeight(20000000).SignAndBuild(identityset.PrivateKey(0))
			bs.ProcessBlock(ctx, "test", &blk)
			time.Sleep(time.Second * 1)
		}
	}(bs)
	require.NoError(bs.Start(ctx))
	go func(bs BlockSync) {
		for {
			time.Sleep(cfg.Interval)
			t.Log(bs.BuildReport())
		}
	}(bs)
	defer func() {
		require.NoError(bs.Stop(ctx))
	}()

	time.Sleep(time.Second * 60)
}
