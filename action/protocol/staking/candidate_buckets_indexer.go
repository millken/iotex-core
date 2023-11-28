// Copyright (c) 2020 IoTeX Foundation
// This source code is provided 'as is' and no warranties are given as to title or non-infringement, merchantability
// or fitness for purpose and, to the extent permitted by law, all liability for your use of the code is disclaimed.
// This source code is governed by Apache License 2.0 that can be found in the LICENSE file.

package staking

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	"github.com/iotexproject/iotex-proto/golang/iotextypes"

	"github.com/iotexproject/iotex-core/db"
	"github.com/iotexproject/iotex-core/db/batch"
	"github.com/iotexproject/iotex-core/pkg/util/byteutil"
)

const (
	// StakingCandidatesNamespace is a namespace to store candidates with epoch start height
	StakingCandidatesNamespace = "stakingCandidates"
	// StakingBucketsNamespace is a namespace to store vote buckets with epoch start height
	StakingBucketsNamespace = "stakingBuckets"
	// StakingMetaNamespace is a namespace to store metadata
	StakingMetaNamespace = "stakingMeta"
)

var (
	_candHeightKey   = []byte("cht")
	_bucketHeightKey = []byte("bht")
)

// CandidatesBucketsIndexer is an indexer to store candidates by given height
type CandidatesBucketsIndexer struct {
	latestCandidatesHeight uint64
	latestBucketsHeight    uint64
	kvStore                KvVersioned
}

// NewStakingCandidatesBucketsIndexer creates a new StakingCandidatesIndexer
func NewStakingCandidatesBucketsIndexer(kv KvVersioned) (*CandidatesBucketsIndexer, error) {
	if kv == nil {
		return nil, ErrMissingField
	}
	return &CandidatesBucketsIndexer{
		kvStore: kv,
	}, nil
}

// Start starts the indexer
func (cbi *CandidatesBucketsIndexer) Start(ctx context.Context) error {
	if err := cbi.kvStore.Start(ctx); err != nil {
		return err
	}
	ret, err := cbi.kvStore.Get(StakingMetaNamespace, _candHeightKey)
	switch errors.Cause(err) {
	case nil:
		cbi.latestCandidatesHeight = byteutil.BytesToUint64BigEndian(ret)
	case db.ErrNotExist:
		cbi.latestCandidatesHeight = 0
	default:
		return err
	}

	ret, err = cbi.kvStore.Get(StakingMetaNamespace, _bucketHeightKey)
	switch errors.Cause(err) {
	case nil:
		cbi.latestBucketsHeight = byteutil.BytesToUint64BigEndian(ret)
	case db.ErrNotExist:
		cbi.latestBucketsHeight = 0
	default:
		return err
	}

	return nil
}

// Stop stops the indexer
func (cbi *CandidatesBucketsIndexer) Stop(ctx context.Context) error {
	return cbi.kvStore.Stop(ctx)
}

// PutCandidates puts candidates into indexer
func (cbi *CandidatesBucketsIndexer) PutCandidates(height uint64, candidates CandidateList) error {
	for _, b := range candidates {
		v, err := b.Serialize()
		if err != nil {
			return err
		}
		cbi.kvStore.PutOnlyOnChange(StakingCandidatesNamespace, b.Owner.Bytes(), v)
	}
	cbi.latestCandidatesHeight = height
	return nil
}

// GetCandidates gets candidates from indexer given epoch start height
func (cbi *CandidatesBucketsIndexer) GetCandidates(height uint64, offset, limit uint32) (*iotextypes.CandidateListV2, uint64, error) {
	if height > cbi.latestCandidatesHeight {
		height = cbi.latestCandidatesHeight
	}
	candidateList := &iotextypes.CandidateListV2{}
	ret, err := getFromIndexer(cbi.kvStore, StakingCandidatesNamespace, height)
	cause := errors.Cause(err)
	if cause == db.ErrNotExist || cause == db.ErrBucketNotExist {
		return candidateList, height, nil
	}
	if err != nil {
		return nil, height, err
	}
	if err := proto.Unmarshal(ret, candidateList); err != nil {
		return nil, height, err
	}
	length := uint32(len(candidateList.Candidates))
	if offset >= length {
		return &iotextypes.CandidateListV2{}, height, nil
	}
	end := offset + limit
	if end > uint32(len(candidateList.Candidates)) {
		end = uint32(len(candidateList.Candidates))
	}
	candidateList.Candidates = candidateList.Candidates[offset:end]
	return candidateList, height, nil
}

// PutBuckets puts vote buckets into indexer
func (cbi *CandidatesBucketsIndexer) PutBuckets(height uint64, buckets []*VoteBucket) error {
	for _, b := range buckets {
		v, err := b.Serialize()
		if err != nil {
			return err
		}
		k := &VoteBucketKey{
			IsNative: b.isNative(),
			Index:    b.Index,
		}
		cbi.kvStore.PutOnlyOnChange(StakingBucketsNamespace, k.Serialize(), v)
	}

	cbi.latestBucketsHeight = height
	return nil
}

// GetBuckets gets vote buckets from indexer given epoch start height
func (cbi *CandidatesBucketsIndexer) GetBuckets(height uint64, offset, limit uint32) (*iotextypes.VoteBucketList, uint64, error) {
	if height > cbi.latestBucketsHeight {
		height = cbi.latestBucketsHeight
	}
	buckets := &iotextypes.VoteBucketList{}
	ret, err := getFromIndexer(cbi.kvStore, StakingBucketsNamespace, height)
	cause := errors.Cause(err)
	if cause == db.ErrNotExist || cause == db.ErrBucketNotExist {
		return buckets, height, nil
	}
	if err != nil {
		return nil, height, err
	}
	if err := proto.Unmarshal(ret, buckets); err != nil {
		return nil, height, err
	}
	length := uint32(len(buckets.Buckets))
	if offset >= length {
		return &iotextypes.VoteBucketList{}, height, nil
	}
	end := offset + limit
	if end > uint32(len(buckets.Buckets)) {
		end = uint32(len(buckets.Buckets))
	}
	buckets.Buckets = buckets.Buckets[offset:end]
	return buckets, height, nil
}

func (cbi *CandidatesBucketsIndexer) putToIndexer(ns string, height uint64, data []byte) error {
	var (
		heightKey []byte
	)
	switch ns {
	case StakingCandidatesNamespace:
		heightKey = _candHeightKey
	case StakingBucketsNamespace:
		heightKey = _bucketHeightKey
	default:
		return ErrTypeAssertion
	}

	heightBytes := byteutil.Uint64ToBytesBigEndian(height)

	// update latest height
	b := batch.NewBatch()
	b.Put(StakingMetaNamespace, heightKey, heightBytes, "failed to update indexer height")
	if err := cbi.kvStore.WriteBatch(b); err != nil {
		return err
	}
	return nil
}

func getFromIndexer(kv KvVersioned, ns string, height uint64) ([]byte, error) {
	b, err := kv.Get(ns, byteutil.Uint64ToBytesBigEndian(height))
	switch errors.Cause(err) {
	case nil:
		return b, nil
	// case db.ErrNotExist:
	// 	// height does not exist, fallback to previous height TODO: remove this fallback
	// 	return kv.SeekPrev([]byte(ns), height)
	default:
		return nil, err
	}
}
