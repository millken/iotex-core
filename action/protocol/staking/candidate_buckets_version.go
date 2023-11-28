package staking

import (
	"github.com/iotexproject/iotex-core/db"
	"github.com/iotexproject/iotex-core/pkg/enc"
	"github.com/pkg/errors"
)

// KvVersioned is a kv store with version, which is used to store data with version
// TODO: move this interface to db package
type KvVersioned interface {
	db.KVStore

	// Version returns the key's most recent version
	Version(bucket string, key []byte) (uint64, error)

	// SetVersion sets the version before calling Get() and Put()
	SetVersion(uint64) KvVersioned

	// PutOnlyOnChange writes the value only when it changes from
	// last-time value
	PutOnlyOnChange(bucket string, key, value []byte) bool
}

var (
	ErrInvalidBucketKey = errors.New("invalid bucket key")
)

// VoteBucketKey is the key of vote bucket, which is used to store vote bucket
type VoteBucketKey struct {
	IsNative bool
	Index    uint64
}

// Serialize serializes vote bucket key
func (vbk *VoteBucketKey) Serialize() []byte {
	buf := make([]byte, 9)
	if vbk.IsNative {
		buf[0] = 0
	} else {
		buf[0] = 1
	}
	enc.MachineEndian.PutUint64(buf[1:], vbk.Index)
	return buf[:]
}

// Deserialize deserializes vote bucket key
func (vbk *VoteBucketKey) Deserialize(buf []byte) error {
	if len(buf) < 9 {
		return ErrInvalidBucketKey
	}
	vbk.IsNative = buf[0] == 0
	vbk.Index = enc.MachineEndian.Uint64(buf[1:])
	return nil
}
