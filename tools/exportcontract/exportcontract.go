// Copyright (c) 2020 IoTeX Foundation
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/iotexproject/iotex-address/address"
	"github.com/iotexproject/iotex-antenna-go/v2/iotex"
	"github.com/iotexproject/iotex-election/contract"
	"github.com/iotexproject/iotex-proto/golang/iotexapi"

	"github.com/iotexproject/iotex-core/tools/exportcontract/contracts"
)

type BucketsResult struct {
	Count           *big.Int
	Indexes         []*big.Int
	StakeStartTimes []*big.Int
	StakeDurations  []*big.Int
	Decays          []bool
	StakedAmounts   []*big.Int
	CanNames        [][12]byte
	Owners          []common.Address
}

type Bucket struct {
	index         uint64
	startTime     time.Time
	duration      time.Duration
	decay         bool
	amount        *big.Int
	candidateName []byte
	owner         address.Address
}

// Registration defines a registration in contract
type Registration struct {
	name              string
	address           string
	operatorAddress   string
	rewardAddress     string
	selfStakingWeight uint64
}

func decodeAddress(data [][32]byte, num int) ([][]byte, error) {
	if len(data) != 2*num {
		return nil, errors.New("the length of address array is not as expected")
	}
	keys := [][]byte{}
	for i := 0; i < num; i++ {
		key := append(data[2*i][:], data[2*i+1][:9]...)
		keys = append(keys, key)
	}

	return keys, nil
}

func fetchCandidates(contractAddress common.Address, opts *bind.CallOpts) ([]*Registration, error) {
	var allCandidates []*Registration
	client, err := ethclient.Dial("https://mainnet.infura.io/v3/b355cae6fafc4302b106b937ee6c15af")
	if err != nil {
		return nil, err
	}
	limit := big.NewInt(255)
	startIndex := big.NewInt(1)
	for {
		caller, err := contract.NewRegisterCaller(contractAddress, client)
		if err != nil {
			return nil, err
		}
		var count *big.Int
		if count, err = caller.CandidateCount(opts); err != nil {
			return nil, err
		}
		if startIndex.Cmp(count) >= 0 {
			break
		}
		retval, err := caller.GetAllCandidates(opts, startIndex, limit)
		if err != nil {
			return nil, err
		}
		num := len(retval.Names)
		if len(retval.Addresses) != num {
			return nil, errors.New("invalid addresses from GetAllCandidates")
		}
		operatorPubKeys, err := decodeAddress(retval.IoOperatorAddr, num)
		if err != nil {
			return nil, err
		}
		rewardPubKeys, err := decodeAddress(retval.IoRewardAddr, num)
		if err != nil {
			return nil, err
		}
		candidates := make([]*Registration, num)
		for i := 0; i < num; i++ {
			candidates[i] = &Registration{
				name:              string(retval.Names[i][:]),
				address:           common.BytesToAddress(retval.Addresses[i][:]).String(),
				operatorAddress:   string(operatorPubKeys[i]),
				rewardAddress:     string(rewardPubKeys[i]),
				selfStakingWeight: retval.Weights[i].Uint64(),
			}
		}
		allCandidates = append(allCandidates, candidates...)
		if len(candidates) < int(limit.Int64()) {
			break
		}
		startIndex = new(big.Int).Add(startIndex, big.NewInt(int64(num)))
	}
	return allCandidates, nil
}

func fetchEtherumBuckets(contractAddress common.Address, opts *bind.CallOpts) ([]Bucket, error) {
	var (
		err        error
		prevBucket struct {
			CanName          [12]byte
			StakedAmount     *big.Int
			StakeDuration    *big.Int
			StakeStartTime   *big.Int
			NonDecay         bool
			UnstakeStartTime *big.Int
			BucketOwner      common.Address
			CreateTime       *big.Int
			Prev             *big.Int
			Next             *big.Int
		}
		buckets []Bucket
		retval  BucketsResult
	)
	client, err := ethclient.Dial("https://mainnet.infura.io/v3/b355cae6fafc4302b106b937ee6c15af")
	if err != nil {
		return nil, err
	}
	caller, err := contracts.NewStakingCaller(contractAddress, client)
	if err != nil {
		return nil, err
	}
	limit := big.NewInt(255)
	previousIndex := big.NewInt(0)
	for {
		fmt.Println(previousIndex)
		if prevBucket, err = caller.Buckets(opts, previousIndex); err == nil {
			if prevBucket.Next.Cmp(big.NewInt(0)) <= 0 {
				break
			}
			retval, err = caller.GetActiveBuckets(opts, previousIndex, limit)
			if err != nil {
				return nil, err
			}
			if retval.Count == nil || retval.Count.Cmp(big.NewInt(0)) == 0 || len(retval.Indexes) == 0 {
				fmt.Println("something goes wrong")
				os.Exit(0)
			}
			cnt := int64(0)
			for i, index := range retval.Indexes {
				if big.NewInt(0).Cmp(index) == 0 {
					break
				}
				cnt++
				owner, err := address.FromBytes(retval.Owners[i].Bytes())
				if err != nil {
					return nil, err
				}
				buckets = append(buckets, Bucket{
					index:         index.Uint64(),
					startTime:     time.Unix(retval.StakeStartTimes[i].Int64(), 0),
					duration:      time.Duration(retval.StakeDurations[i].Uint64()*24) * time.Hour,
					amount:        retval.StakedAmounts[i],
					owner:         owner,
					candidateName: retval.CanNames[i][:],
					decay:         retval.Decays[i],
				})
				if index.Cmp(previousIndex) > 0 {
					previousIndex = index
				}
			}
			// process result and append to buckets
			// update previousIndex
			if limit.Int64() > cnt {
				break
			}
		}
	}

	return buckets, nil
}

func fetchNativeBuckets(contractAddress string) ([]Bucket, error) {
	ca, err := address.FromString(contractAddress)
	if err != nil {
		return nil, err
	}
	ns, err := abi.JSON(strings.NewReader(contracts.NativeStakingABI))
	if err != nil {
		return nil, err
	}
	previousIndex := big.NewInt(0)
	limit := big.NewInt(255)
	var buckets []Bucket
	conn, err := iotex.NewDefaultGRPCConn("api.iotex.one:443")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := iotex.NewReadOnlyClient(iotexapi.NewAPIServiceClient(conn))
	c := client.ReadOnlyContract(ca, ns)
	retval := &BucketsResult{}
	for {
		fmt.Println(previousIndex)
		data, err := c.Read("getActivePyggs", previousIndex, limit).Call(context.Background())
		if err != nil {
			return nil, err
		}
		if err := ns.Unpack(retval, "getActivePyggs", data.Raw); err != nil {
			return nil, err
		}
		if retval.Count == nil || retval.Count.Cmp(big.NewInt(0)) == 0 || len(retval.Indexes) == 0 {
			fmt.Println("something goes wrong")
			os.Exit(0)
		}
		cnt := int64(0)
		for i, index := range retval.Indexes {
			if big.NewInt(0).Cmp(index) == 0 {
				break
			}
			cnt++
			owner, err := address.FromBytes(retval.Owners[i].Bytes())
			if err != nil {
				return nil, err
			}
			buckets = append(buckets, Bucket{
				index:         index.Uint64(),
				startTime:     time.Unix(retval.StakeStartTimes[i].Int64(), 0),
				duration:      time.Duration(retval.StakeDurations[i].Uint64()*24) * time.Hour,
				amount:        retval.StakedAmounts[i],
				owner:         owner,
				candidateName: retval.CanNames[i][:],
				decay:         retval.Decays[i],
			})
			if index.Cmp(previousIndex) > 0 {
				previousIndex = index
			}
		}
		// process result and append to buckets
		// update previousIndex
		if limit.Int64() > cnt {
			break
		}
	}
	return buckets, nil
}

func main() {
	height := uint64(9917100)
	opts := &bind.CallOpts{}
	if height != 0 {
		opts = &bind.CallOpts{BlockNumber: new(big.Int).SetUint64(height)}
	}
	// mainnet
	candidates, err := fetchCandidates(common.HexToAddress("0x95724986563028deb58f15c5fac19fa09304f32d"), opts)
	// testnet
	// candidates, err := fetchCandidates(common.HexToAddress("0x92adef0e5e0c2b4f64a1ac79823f7ad3bc1662c4"), opts)
	if err != nil {
		fmt.Println("failed to fetch candidates", err)
		os.Exit(1)
	}
	for _, candidate := range candidates {
		fmt.Printf("%s, %s, %s, %s\n", candidate.name, candidate.address, candidate.operatorAddress, candidate.rewardAddress)
	}
	// fetch buckets from ethereum
	// mainnet
	ebs, err := fetchEtherumBuckets(common.HexToAddress("0x87c9dbff0016af23f5b1ab9b8e072124ab729193"), opts)
	// testnet
	// ebs, err := fetchEtherumBuckets(common.HexToAddress("0x3bbe2346c40d34fc3f66ab059f75a6caece2c3b3"), opts)

	if err != nil {
		fmt.Println("failed to fetch ethereum staking buckets", err)
		os.Exit(1)
	}
	// fetch buckets from native staking contract 1
	// mainnet
	nbs1, err := fetchNativeBuckets("io1xpq62aw85uqzrccg9y5hnryv8ld2nkpycc3gza")
	// testnet
	// nbs1, err := fetchNativeBuckets("io1w97pslyg7qdayp8mfnffxkjkpapaf83wmmll2l")
	if err != nil {
		fmt.Println("failed to fetch native staking buckets", err)
		os.Exit(1)
	}
	/*
		// fetch buckets from native staking contract 2
		nbs2, err := fetchNativeBuckets()
		if err != nil {
			fmt.Println("failed to fetch native staking buckets", err)
			os.Exit(1)
		}
	*/
	for _, bucket := range ebs {
		fmt.Printf("%d, %s, %s, %s, %s, %s, %t\n", bucket.index, string(bucket.candidateName), bucket.owner, bucket.amount, bucket.startTime, bucket.duration, !bucket.decay)
	}
	for _, bucket := range nbs1 {
		fmt.Printf("%d, %s, %s, %s, %s, %s, %t\n", bucket.index+10000, string(bucket.candidateName), bucket.owner, bucket.amount, bucket.startTime, bucket.duration, !bucket.decay)
	}
	/*
		for _, bucket := range nbs2 {
			fmt.Printf("%d, %s, %s, %t, %s, %s, %s\n", bucket.index+20000, bucket.startTime, bucket.duration, bucket.decay, bucket.amount, hex.EncodeToString(bucket.candidateName), bucket.owner)
		}
	*/
}
