package main

import (
	"context"
	"encoding/binary"

	"flag"
	"fmt"
	"log"
	"os"

	"github.com/iotexproject/go-pkgs/hash"
	"github.com/iotexproject/iotex-core/action/protocol/staking"
	"github.com/iotexproject/iotex-core/db"
	"github.com/iotexproject/iotex-proto/golang/iotextypes"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

var (
	_stakingDBV1Path string
	_stakingDBV2Path string
)

func init() {
	flag.StringVar(&_stakingDBV1Path, "staking-db-v1-path", "", "staking db v1 path")
	flag.StringVar(&_stakingDBV2Path, "staking-db-v2-path", "", "staking db v2 path")
	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "usage: stakingdb_refactor -staking-db-v1-path=[string] -staking-db-v2-path=[string]\n")
		flag.PrintDefaults()
		os.Exit(2)
	}
	flag.Parse()
}

func handleStakingCandidates(db1 *bolt.DB, db2 *staking.CandidatesBucketsIndexer) error {
	return db1.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(staking.StakingCandidatesNamespace))
		if b == nil {
			return errors.New("Bucket not found")
		}
		c := b.Cursor()
		candidates := &iotextypes.CandidateListV2{}
		for k, v := c.First(); k != nil; k, v = c.Next() {
			h := binary.BigEndian.Uint64(k)
			if h > 30000000 {
				continue
			}
			if err := proto.Unmarshal(v, candidates); err != nil {
				return errors.Wrapf(err, "failed to unmarshal candidate list at height %d, %x", h, v)
			}
			log.Printf("handleStakingCandidates height: %d, candidates: %d\n", h, len(candidates.GetCandidates()))
			db2.GetKVStore().SetVersion(h)
			if err := db2.PutCandidates(h, candidates); err != nil {
				return errors.Wrapf(err, "failed to put candidates at height %d", h)
			}
		}

		return nil
	})
}

func handleStakingBuckets(db1 *bolt.DB, db2 *staking.CandidatesBucketsIndexer) error {
	return db1.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(staking.StakingBucketsNamespace))
		if b == nil {
			return errors.New("Bucket not found")
		}
		c := b.Cursor()
		buckets := &iotextypes.VoteBucketList{}
		for k, v := c.First(); k != nil; k, v = c.Next() {
			h := binary.BigEndian.Uint64(k)
			if h > 30000000 {
				continue
			}
			if err := proto.Unmarshal(v, buckets); err != nil {
				return errors.Wrapf(err, "failed to unmarshal vote bucket list at height %d, %x", h, v)
			}
			log.Printf("handleStakingBuckets height: %d, buckets: %d\n", h, len(buckets.GetBuckets()))
			db2.GetKVStore().SetVersion(h)
			if err := db2.PutBuckets(h, buckets); err != nil {
				return errors.Wrapf(err, "failed to put buckets at height %d", h)
			}
		}

		return nil
	})

}

func main() {
	o := &bolt.Options{ReadOnly: true}
	db1, err := bolt.Open(_stakingDBV1Path, 0600, o)
	if err != nil {
		log.Fatal(err)
	}
	defer db1.Close()

	ctx := context.Background()
	cfg := db.DefaultConfig
	cfg.DbPath = _stakingDBV2Path
	db2, err := staking.NewStakingCandidatesBucketsIndexer(db.NewBoltDBVersioned(cfg, func(in []byte) []byte {
		h := hash.Hash160b(in)
		return h[:]
	}, db.NonversionedNamespaceOption(staking.StakingMetaNamespace)))
	if err != nil {
		log.Fatal(err)
	}
	if err := db2.Start(ctx); err != nil {
		log.Fatal(err)
	}
	defer db2.Stop(ctx)
	if err := handleStakingCandidates(db1, db2); err != nil {
		log.Fatal(err)
	}
	if err := handleStakingBuckets(db1, db2); err != nil {
		log.Fatal(err)
	}
}
