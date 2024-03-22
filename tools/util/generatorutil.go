// Copyright (c) 2019 IoTeX Foundation
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package util

import (
	"context"
	"math/big"
	"math/rand"
	"sync"

	"github.com/cenkalti/backoff"
	"github.com/iotexproject/iotex-core/action"
	"github.com/iotexproject/iotex-core/pkg/log"
	"github.com/iotexproject/iotex-proto/golang/iotexapi"
	"go.uber.org/zap"
)

type nounceManager struct {
	mu           sync.RWMutex
	pendingNonce map[string]uint64
}

// AccountManager is tbd
type AccountManager struct {
	AccountList        []*AddressKey
	nonceMng           nounceManager
	nonceProcessingMap sync.Map
}

type FeedT struct {
	Processing bool
	Time       int64
}

// NewAccountManager is tbd
func NewAccountManager(accounts []*AddressKey) *AccountManager {
	return &AccountManager{
		AccountList: accounts,
		nonceMng: nounceManager{
			pendingNonce: make(map[string]uint64),
		},
		nonceProcessingMap: sync.Map{},
	}
}

// Get is tbd
func (ac *AccountManager) Get(addr string) uint64 {
	ac.nonceMng.mu.RLock()
	defer ac.nonceMng.mu.RUnlock()
	return ac.nonceMng.pendingNonce[addr]
}

// GetAndInc is tbd
func (ac *AccountManager) GetAndInc(addr string) uint64 {
	var ret uint64
	ac.nonceMng.mu.Lock()
	defer ac.nonceMng.mu.Unlock()
	ret = ac.nonceMng.pendingNonce[addr]
	ac.nonceMng.pendingNonce[addr]++
	return ret
}

// GetAllAddr is tbd
func (ac *AccountManager) GetAllAddr() []string {
	var ret []string
	for _, v := range ac.AccountList {
		ret = append(ret, v.EncodedAddr)
	}
	return ret
}

// Set is tbd
func (ac *AccountManager) Set(addr string, val uint64) {
	ac.nonceMng.mu.Lock()
	defer ac.nonceMng.mu.Unlock()
	ac.nonceMng.pendingNonce[addr] = val
}

// UpdateNonce is tbd
func (ac *AccountManager) UpdateNonce(client iotexapi.APIServiceClient) error {
	// load the nonce and balance of addr
	for _, account := range ac.AccountList {
		addr := account.EncodedAddr
		err := backoff.Retry(func() error {
			acctDetails, err := client.GetAccount(context.Background(), &iotexapi.GetAccountRequest{Address: addr})
			if err != nil {
				return err
			}
			ac.Set(addr, acctDetails.GetAccountMeta().PendingNonce)
			return nil
		}, backoff.NewExponentialBackOff())
		if err != nil {
			log.L().Fatal("Failed to inject actions by APS",
				zap.Error(err),
				zap.String("addr", account.EncodedAddr))
			return err
		}
	}
	return nil
}

func (ac *AccountManager) NonceProcessingLoad(sender string) (FeedT, bool) {
	ft, ok := ac.nonceProcessingMap.Load(sender)
	return ft.(FeedT), ok
}

func (ac *AccountManager) NonceProcessingStore(sender string, ft FeedT) {
	ac.nonceProcessingMap.Store(sender, ft)
}

// ActionGenerator is tbd
func ActionGenerator(
	actionType int,
	accountManager *AccountManager,
	chainID uint32,
	transferGasLimit uint64,
	transferGasPrice *big.Int,
	executionGasLimit uint64,
	executionGasPrice *big.Int,
	contractAddr string,
	transferPayload, executionPayload []byte,
) (action.SealedEnvelope, error) {
	var (
		selp      action.SealedEnvelope
		err       error
		delegates []*AddressKey
	)
	for _, addr := range accountManager.AccountList {
		pm, ok := accountManager.NonceProcessingLoad(addr.EncodedAddr)
		if ok {
			if !pm.Processing {
				delegates = append(delegates, addr)
			}
		}
	}
	var (
		randNum   = rand.Intn(len(delegates))
		sender    = delegates[randNum]
		recipient = delegates[(randNum+1)%len(delegates)]
		nonce     = accountManager.GetAndInc(sender.EncodedAddr)
	)
	switch actionType {
	case 1:
		selp, _, err = createSignedTransfer(sender, recipient, big.NewInt(0), chainID, nonce, transferGasLimit, transferGasPrice, transferPayload)
	case 2:
		selp, _, err = createSignedExecution(sender, contractAddr, chainID, nonce, big.NewInt(0), executionGasLimit, executionGasPrice, executionPayload)
	case 3:
		if rand.Intn(2) == 0 {
			selp, _, err = createSignedTransfer(sender, recipient, big.NewInt(0), chainID, nonce, transferGasLimit, transferGasPrice, transferPayload)
		} else {
			selp, _, err = createSignedExecution(sender, contractAddr, chainID, nonce, big.NewInt(0), executionGasLimit, executionGasPrice, executionPayload)
		}
	}
	return selp, err
}
