package bstore

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"sync/atomic"

	"github.com/iotexproject/iotex-address/address"
	"github.com/iotexproject/iotex-core/db"
)

var (
	kdb                                        db.KVStoreBasic
	kdbOpened, lastCommitHeight, pendingHeight uint64
	storeMap                                   = make(map[string]*big.Int)
)

var (
	currentHeight   uint64
	accountBalances = make(map[string]*big.Int)
)

func OpenDB(ctx context.Context, dbPath string) error {
	var err error
	kdb, err = db.CreateKVStore(db.DefaultConfig, dbPath)
	if err != nil {
		return err
	}
	if err := kdb.Start(ctx); err != nil {
		return err
	}
	atomic.StoreUint64(&kdbOpened, 1)
	return nil
}

func CloseDB() error {
	atomic.StoreUint64(&kdbOpened, 0)
	return kdb.Stop(context.Background())
}
func StoreAccountBalance(height uint64, addr address.Address, balance *big.Int) error {
	if height != currentHeight {
		if err := sqlStoreAccount(currentHeight, accountBalances); err != nil {
			return err
		}
		accountBalances = make(map[string]*big.Int)
		currentHeight = height
	}

	accountBalances[addr.String()] = balance
	fmt.Printf("StoreAccountBalance: %s, %d, %s\n", addr.String(), height, balance.String())
	return nil
}

func kvStoreAccount(height uint64, accounts map[string]*big.Int) error {
	if atomic.LoadUint64(&kdbOpened) == 0 || len(accounts) == 0 {
		return nil
	}
	for k, b := range accounts {
		if err := kdb.Put(k, []byte(strconv.FormatUint(height, 10)), b.Bytes()); err != nil {
			return err
		}
	}
	return nil
}
