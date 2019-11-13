// Copyright (c) 2019 IoTeX Foundation
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package db

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"testing"

	"github.com/iotexproject/go-pkgs/hash"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"

	"github.com/iotexproject/iotex-core/testutil"
)

func TestNewCountingIndex(t *testing.T) {
	require := require.New(t)

	_, err := NewCountingIndex(nil, 3, []byte{1}, 0)
	require.Equal(ErrInvalid, errors.Cause(err))
	_, err = NewCountingIndex(&bolt.DB{}, 3, nil, 0)
	require.Equal(ErrInvalid, errors.Cause(err))
	_, err = NewCountingIndex(&bolt.DB{}, 3, []byte{}, 0)
	require.Equal(ErrInvalid, errors.Cause(err))
}

func TestCountingIndex(t *testing.T) {
	testFunc := func(kv KVStore, t *testing.T) {
		require := require.New(t)

		require.NoError(kv.Start(context.Background()))
		defer func() {
			require.NoError(kv.Stop(context.Background()))
		}()

		bucket := []byte("test")
		_, err := kv.CountingIndex(bucket)
		require.Equal(ErrBucketNotExist, errors.Cause(err))

		index, err := kv.CreateCountingIndexNX(bucket)
		require.NoError(err)
		require.Equal(uint64(0), index.Size())

		// write 200 entries in batch mode
		for i := 0; i < 200; i++ {
			h := hash.Hash160b([]byte(strconv.Itoa(i)))
			require.NoError(index.Add(h[:], true))
		}
		// cannot Add() before Commit() in batch mode
		require.Equal(ErrInvalid, errors.Cause(index.Add([]byte{1}, false)))
		require.NoError(index.Commit())
		require.EqualValues(200, index.Size())
		// cannot get > size
		_, err = index.Get(index.Size())
		require.Equal(ErrNotExist, errors.Cause(err))
		k, err := index.Get(10)
		require.NoError(err)
		h := hash.Hash160b([]byte(strconv.Itoa(10)))
		require.Equal(h[:], k)
		index.Close()

		// re-open the bucket
		index, err = kv.CountingIndex(bucket)
		require.NoError(err)
		// write another 100 entries
		for i := 200; i < 300; i++ {
			h := hash.Hash160b([]byte(strconv.Itoa(i)))
			require.NoError(index.Add(h[:], false))
		}
		require.EqualValues(300, index.Size())

		_, err = index.Range(248, 0)
		require.Equal(ErrInvalid, errors.Cause(err))
		_, err = index.Range(248, 53)
		require.Equal(ErrInvalid, errors.Cause(err))

		// last key
		v, err := index.Range(299, 1)
		require.NoError(err)
		require.Equal(1, len(v))
		h = hash.Hash160b([]byte(strconv.Itoa(299)))
		require.Equal(h[:], v[0])

		// first 5 keys
		v, err = index.Range(0, 5)
		require.NoError(err)
		require.Equal(5, len(v))
		for i := range v {
			h := hash.Hash160b([]byte(strconv.Itoa(i)))
			require.Equal(h[:], v[i])
		}

		// last 40 keys
		v, err = index.Range(260, 40)
		require.NoError(err)
		require.Equal(40, len(v))
		for i := range v {
			h := hash.Hash160b([]byte(strconv.Itoa(260 + i)))
			require.Equal(h[:], v[i])
		}
		index.Close()

		// re-open the bucket, verify size = 300
		index1, err := kv.CountingIndex(bucket)
		require.NoError(err)
		require.EqualValues(300, index1.Size())

		// revert last 40 keys
		err = index1.Revert(0)
		require.Equal(ErrInvalid, errors.Cause(err))
		err = index1.Revert(index1.Size() + 1)
		require.Equal(ErrInvalid, errors.Cause(err))
		require.NoError(index1.Revert(40))
		require.EqualValues(260, index1.Size())

		// last 40 keys
		_, err = index1.Range(220, 41)
		require.Equal(ErrInvalid, errors.Cause(err))
		v, err = index1.Range(220, 40)
		require.NoError(err)
		require.Equal(40, len(v))
		for i := range v {
			h := hash.Hash160b([]byte(strconv.Itoa(220 + i)))
			require.Equal(h[:], v[i])
		}
	}

	t.Run("in-mem KVStore", func(t *testing.T) {
		testFunc(NewMemKVStore(), t)
	})

	path := "test-iterate.bolt"
	testFile, _ := ioutil.TempFile(os.TempDir(), path)
	testPath := testFile.Name()
	cfg.DbPath = testPath
	t.Run("Bolt DB", func(t *testing.T) {
		testutil.CleanupPath(t, testPath)
		defer testutil.CleanupPath(t, testPath)
		testFunc(NewBoltDB(cfg), t)
	})
}

const (
	Tenants = 10000
	Keys    = 200
)

func TestBulk(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	testFunc := func(kv KVStore, t *testing.T) {
		require := require.New(t)

		require.NoError(kv.Start(context.Background()))
		defer func() {
			require.NoError(kv.Stop(context.Background()))
		}()

		// create 10000 tenants
		for i := 0; i < Tenants; i++ {
			h := hash.Hash160b([]byte(strconv.Itoa(i)))
			tenant, err := kv.CreateCountingIndexNX(h[:])
			require.NoError(err)

			for i := 0; i < Keys; i++ {
				h := hash.Hash160b([]byte(strconv.Itoa(i)))
				require.NoError(tenant.Add(h[:], true))
			}
			require.NoError(tenant.Commit())
			tenant.Close()
			fmt.Printf("write tenant %d:\n", i)
		}
	}

	cfg.DbPath = "test-bulk.dat"
	t.Run("Bolt DB", func(t *testing.T) {
		testutil.CleanupPath(t, cfg.DbPath)
		testFunc(NewBoltDB(cfg), t)
	})
}

func TestCheckBulk(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	testFunc := func(kv KVStore, t *testing.T) {
		require := require.New(t)

		require.NoError(kv.Start(context.Background()))
		defer func() {
			require.NoError(kv.Stop(context.Background()))
		}()

		// verify 1000 tenants
		for i := 0; i < Tenants; i++ {
			h := hash.Hash160b([]byte(strconv.Itoa(i)))
			index, err := kv.CountingIndex(h[:])
			require.NoError(err)
			require.EqualValues(Keys, index.Size())

			value, err := index.Range(0, Keys)
			require.NoError(err)
			require.EqualValues(Keys, len(value))

			for i := range value {
				h := hash.Hash160b([]byte(strconv.Itoa(i)))
				require.Equal(h[:], value[i])
			}
			fmt.Printf("verify tenant: %d\n", i)
		}
	}

	cfg.DbPath = "test-bulk.dat"
	t.Run("Bolt DB", func(t *testing.T) {
		defer testutil.CleanupPath(t, cfg.DbPath)
		testFunc(NewBoltDB(cfg), t)
	})
}