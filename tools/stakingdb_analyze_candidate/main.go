package main

import (
	"encoding/binary"

	"flag"
	"fmt"
	"log"
	"os"

	"github.com/iotexproject/iotex-core/action/protocol/staking"
	"github.com/iotexproject/iotex-proto/golang/iotextypes"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

var (
	_stakingDBV1Path string
)

func init() {
	flag.StringVar(&_stakingDBV1Path, "db-path", "", "staking db v1 path")
	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "usage: stakingdb_analyze_candidate -db-path=[string] \n")
		flag.PrintDefaults()
		os.Exit(2)
	}
	flag.Parse()
}

type candidateTotal struct {
	changed     uint64
	noneChanged uint64
	last        *iotextypes.CandidateV2
}

func main() {
	o := &bolt.Options{ReadOnly: true}
	db1, err := bolt.Open(_stakingDBV1Path, 0600, o)
	if err != nil {
		log.Fatal(err)
	}
	defer db1.Close()
	cands := make(map[string]*iotextypes.CandidateV2)
	candTotal := make(map[string]*candidateTotal)
	db1.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(staking.StakingCandidatesNamespace))
		if b == nil {
			return errors.New("Bucket not found")
		}
		candidates := &iotextypes.CandidateListV2{}
		b.ForEach(func(k, v []byte) error {
			h := binary.BigEndian.Uint64(k)
			if h > 30000000 {
				return nil
			}
			if err := proto.Unmarshal(v, candidates); err != nil {
				return errors.Wrapf(err, "failed to unmarshal candidate list at height %d, %x", h, v)
			}
			//log.Printf("handleStakingCandidates height: %d, candidates: %d\n", h, len(candidates.GetCandidates()))
			for _, cand := range candidates.GetCandidates() {
				cands[cand.OwnerAddress] = cand
				if err != nil {
					return errors.Wrapf(err, "failed to marshal candidate at height %d", h)
				}
				if _, ok := candTotal[cand.OwnerAddress]; !ok {
					candTotal[cand.OwnerAddress] = &candidateTotal{}
				}
				if candTotal[cand.OwnerAddress].last == nil {
					candTotal[cand.OwnerAddress].last = cand
				} else {
					if proto.Equal(candTotal[cand.OwnerAddress].last, cand) {
						candTotal[cand.OwnerAddress].noneChanged++
					} else {
						candTotal[cand.OwnerAddress].changed++
					}
					candTotal[cand.OwnerAddress].last = cand
				}

			}
			return nil
		})

		return nil
	})
	// for addr, cand := range cands {
	// 	fmt.Printf("addr: %s %s, changed: %d, noneChanged: %d\n", cand.Name, addr, candTotal[addr].changed, candTotal[addr].noneChanged)
	// }
	for addr, total := range candTotal {
		fmt.Printf("%s,%s,%d,%d\n", cands[addr].Name, addr, total.changed, total.noneChanged)
	}
}
