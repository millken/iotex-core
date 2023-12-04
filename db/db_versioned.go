// Copyright (c) 2021 IoTeX Foundation
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/iotexproject/go-pkgs/byteutil"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"

	"github.com/iotexproject/iotex-core/db/batch"
)

// vars
var (
	ErrDBCorruption     = errors.New("DB is corrupted")
	ErrInvalidKeyLength = errors.New("invalid key length")
	ErrInvalidWrite     = errors.New("invalid write attempt")
	_minKey             = []byte{0} // the minimum key, used to store namespace's metadata
)

type (
	// KvVersioned is a versioned key-value store, where each key can (but doesn't
	// have to) have multiple versions of value (corresponding to different heights
	// in a blockchain)
	//
	// Versioning is achieved by using (key + 8-byte version) as actual storage key
	// in the underlying DB. For buckets containing versioned keys, a metadata is
	// stored at the special key = []byte{0}. The metadata specifies the bucket's
	// name and the key length
	//
	// For each versioned key, the special location = key + []byte{0} stores the
	// key's version (as 8-byte big endian). If the location does not store a value,
	// the key has never been written. A zero value means the key has been deleted
	//
	// A namespace/bucket is considered versioned by default, user needs to use the
	// NonversionedNamespaceOption() to specify non-versioned namespaces at the time
	// of creating the versioned key-value store/DB
	//
	// Here's an example of a versioned DB which has 3 buckets:
	// 1. "mta" -- regular bucket storing metadata, key is not versioned
	// 2. "act" -- versioned namespace, key length = 20
	// 3. "stk" -- versioned namespace, key length = 8
	KvVersioned interface {
		KVStore

		// Version returns the key's most recent version
		Version(string, []byte) (uint64, error)

		// SetVersion sets the version before calling Put()
		SetVersion(uint64) KvVersioned

		// CreateVersionedNamespace creates a namespace to store versioned keys
		CreateVersionedNamespace(string, uint32) error
	}

	// BoltDBVersioned is KvVersioned implementation based on bolt DB
	BoltDBVersioned struct {
		*BoltDB
		version      uint64          // version for Get/Put()
		versioned    map[string]int  // buckets for versioned keys
		nonversioned map[string]bool // buckets for non-versioned keys
	}
)

// Option sets an option
type Option func(b *BoltDBVersioned)

// NonversionedNamespaceOption sets non-versioned namespace
func NonversionedNamespaceOption(ns ...string) Option {
	return func(b *BoltDBVersioned) {
		for _, v := range ns {
			b.nonversioned[v] = true
		}
	}
}

// VersionedNamespaceOption sets versioned namespace
func VersionedNamespaceOption(ns string, n int) Option {
	return func(b *BoltDBVersioned) {
		b.versioned[ns] = n
	}
}

// NewMemoryDBVersioned instantiates an in-memory DB with implements KvVersioned
// TODO: implement this
func NewMemoryDBVersioned() *BoltDBVersioned {
	b := &BoltDBVersioned{
		BoltDB:       nil,
		versioned:    make(map[string]int),
		nonversioned: make(map[string]bool),
	}

	return b
}

// NewBoltDBVersioned instantiates an BoltDB with implements KvVersioned
func NewBoltDBVersioned(cfg Config, opts ...Option) *BoltDBVersioned {
	b := &BoltDBVersioned{
		BoltDB:       NewBoltDB(cfg),
		versioned:    make(map[string]int),
		nonversioned: make(map[string]bool),
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

// Start starts the DB
func (b *BoltDBVersioned) Start(ctx context.Context) error {
	if err := b.BoltDB.Start(ctx); err != nil {
		return err
	}
	// verify non-versioned namespace
	for ns := range b.nonversioned {
		vn, err := b.checkVersionedNS(ns)
		if err != nil {
			return err
		}
		if vn != nil {
			return errors.Wrapf(ErrDBCorruption, "expect namespace %s to be non-versioned, but got versioned", ns)
		}
	}
	// verify initial versioned namespace
	buf := batch.NewBatch()
	for ns, n := range b.versioned {
		vn, err := b.checkVersionedNS(ns)
		if err != nil {
			return err
		}
		if vn == nil {
			// create the versioned namespace
			buf.Put(ns, _minKey, (&VersionedNamespace{
				Name:   ns,
				KeyLen: uint32(n),
			}).Serialize(), "failed to create metadata")
		}
	}
	if buf.Size() > 0 {
		return b.BoltDB.WriteBatch(buf)
	}
	return nil
}

// Put writes a <key, value> record
func (b *BoltDBVersioned) Put(ns string, key, value []byte) error {
	versioned, err := b.isVersioned(ns, key)
	if err != nil && err != ErrNotExist {
		return err
	}
	if versioned {
		buf := batch.NewBatch()
		if err == ErrNotExist {
			// namespace not yet created
			buf.Put(ns, _minKey, (&VersionedNamespace{
				Name:   ns,
				KeyLen: uint32(len(key)),
			}).Serialize(), "failed to create metadata")
		}
		buf.Put(ns, versionedKey(key, b.version), value, fmt.Sprintf("failed to put key %x", key))
		buf.Put(ns, append(key, 0), byteutil.Uint64ToBytesBigEndian(b.version), fmt.Sprintf("failed to put key %x's version", key))
		return b.BoltDB.WriteBatch(buf)
	}
	return b.BoltDB.Put(ns, key, value)
}

// Get retrieves the most recent version
func (b *BoltDBVersioned) Get(ns string, key []byte) ([]byte, error) {
	versioned, err := b.isVersioned(ns, key)
	if err != nil {
		return nil, err
	}
	if !versioned {
		return b.BoltDB.Get(ns, key)
	}
	return b.getVersion(ns, key)
}

// Version returns the key's most recent version
func (b *BoltDBVersioned) Version(ns string, key []byte) (uint64, error) {
	versioned, err := b.isVersioned(ns, key)
	if err != nil {
		return 0, err
	}
	if !versioned {
		_, err := b.BoltDB.Get(ns, key)
		return 0, err
	}
	v, err := b.BoltDB.Get(ns, append(key, 0))
	if err != nil {
		return 0, err
	}
	return byteutil.BytesToUint64BigEndian(v), nil
}

// SetVersion sets the version, should ONLY be called before a Put() operation
func (b *BoltDBVersioned) SetVersion(v uint64) KvVersioned {
	b.version = v
	return b
}

// CreateVersionedNamespace creates a namespace to store versioned keys
func (b *BoltDBVersioned) CreateVersionedNamespace(ns string, n uint32) error {
	if _, ok := b.nonversioned[ns]; ok {
		return errors.Wrapf(ErrInvalidWrite, "namespace %s is non-versioned", ns)
	}
	vn, err := b.checkVersionedNS(ns)
	if err != nil {
		return err
	}
	if vn != nil {
		return errors.Wrapf(ErrInvalidWrite, "namespace %s already exist", ns)
	}
	if err = b.BoltDB.Put(ns, _minKey, NewVersionedNamespace(ns, n).Serialize()); err != nil {
		return err
	}
	b.versioned[ns] = int(n)
	return nil
}

// WriteBatch commits a batch
func (b *BoltDBVersioned) WriteBatch(kvsb batch.KVStoreBatch) (err error) {
	if b.db == nil {
		return ErrDBNotStarted
	}

	kvsb.Lock()
	defer kvsb.Unlock()
	var (
		newVNS = make(map[string]int)
		cause  error
	)
	for c := uint8(0); c < b.config.NumRetries; c++ {
		err = b.db.Update(func(tx *bolt.Tx) error {
			for i := 0; i < kvsb.Size(); i++ {
				write, e := kvsb.Entry(i)
				if e != nil {
					return e
				}
				ns := write.Namespace()
				key := write.Key()
				switch write.WriteType() {
				case batch.Put:
					bucket, e := tx.CreateBucketIfNotExists([]byte(ns))
					if e != nil {
						return errors.Wrapf(e, write.Error())
					}
					val := write.Value()
					if bytes.Compare(_minKey, key) == 0 {
						if b.nonversioned[ns] || b.versioned[ns] > 0 || newVNS[ns] > 0 {
							return errors.Wrapf(ErrInvalidWrite, write.Error())
						}
						vns := VersionedNamespace{}
						if err := vns.Deserialize(val); err != nil {
							return errors.Wrap(ErrInvalidWrite, err.Error())
						}
						newVNS[ns] = int(vns.KeyLen)
					} else {
						n := b.versioned[ns]
						if n == 0 {
							n = newVNS[ns]
						}
						if n > 0 {
							if len(key) != n {
								return errors.Wrapf(ErrInvalidKeyLength, write.Error())
							}
							if e := bucket.Put(append(key, 0), byteutil.Uint64ToBytesBigEndian(b.version)); e != nil {
								return errors.Wrapf(e, write.Error())
							}
							key = append(key, byteutil.Uint64ToBytesBigEndian(b.version)...)
						}
					}
					if e := bucket.Put(key, val); e != nil {
						return errors.Wrapf(e, write.Error())
					}
				case batch.Delete:
					bucket := tx.Bucket([]byte(ns))
					if bucket == nil {
						continue
					}
					if e := bucket.Delete(key); e != nil {
						return errors.Wrapf(e, write.Error())
					}
				}
			}
			return nil
		})
		cause = errors.Cause(err)
		if err == nil || cause == ErrInvalidWrite || cause == ErrInvalidKeyLength {
			break
		}
	}
	if err == nil {
		for k, v := range newVNS {
			b.versioned[k] = v
		}
		return
	}
	if cause != ErrInvalidWrite && cause != ErrInvalidKeyLength {
		err = errors.Wrap(ErrIO, err.Error())
	}
	return
}

// getVersion retrieves the <k, v> at certain version
func (b *BoltDBVersioned) getVersion(ns string, key []byte) ([]byte, error) {
	key = versionedKey(key, b.version)
	var value []byte
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(ns))
		if bucket == nil {
			return errors.Wrapf(ErrNotExist, "bucket = %x doesn't exist", []byte(ns))
		}
		c := bucket.Cursor()
		k, v := c.Seek(key)
		if k == nil || bytes.Compare(k, key) == 1 {
			k, v = c.Prev()
			key = append(key[:len(key)-8], 0)
			if k == nil || bytes.Compare(k, key) <= 0 {
				// cursor is at the beginning of the bucket or smaller than minimum key
				return errors.Wrapf(ErrNotExist, "key = %x doesn't exist", key[:len(key)-1])
			}
		}
		value = make([]byte, len(v))
		copy(value, v)
		return nil
	})
	if err == nil {
		return value, nil
	}
	if errors.Cause(err) == ErrNotExist {
		return nil, err
	}
	return nil, errors.Wrap(ErrIO, err.Error())
}

func versionedKey(key []byte, v uint64) []byte {
	return append(key, byteutil.Uint64ToBytesBigEndian(v)...)
}

func (b *BoltDBVersioned) isVersioned(ns string, key []byte) (bool, error) {
	if _, ok := b.nonversioned[ns]; ok {
		return false, nil
	}
	if keyLen, ok := b.versioned[ns]; ok {
		if len(key) != keyLen {
			return true, errors.Wrapf(ErrInvalidKeyLength, "expecting %d, got %d", keyLen, len(key))
		}
		return true, nil

	}
	// check if the namespace already exist in DB
	vn, err := b.checkVersionedNS(ns)
	if err != nil {
		return false, err
	}
	if vn != nil {
		b.versioned[ns] = int(vn.KeyLen)
		return true, nil
	}
	// namespace not yet created
	return true, ErrNotExist
}

func (b *BoltDBVersioned) checkVersionedNS(ns string) (*VersionedNamespace, error) {
	data, err := b.BoltDB.Get(ns, _minKey)
	switch errors.Cause(err) {
	case nil:
		vn := VersionedNamespace{}
		if err := vn.Deserialize(data); err != nil {
			return nil, err
		}
		return &vn, nil
	case ErrNotExist, ErrBucketNotExist:
		return nil, nil
	default:
		return nil, err
	}
}

// ======================================
// funcs for VersionedNamespace object
// ======================================

// VersionedNamespace is the metadata for versioned namespace
type VersionedNamespace struct {
	Name   string `json:"name"`
	KeyLen uint32 `json:"keyLen"`
}

// NewVersionedNamespace returns a new instance of VersionedNamespace
func NewVersionedNamespace(ns string, n uint32) *VersionedNamespace {
	return &VersionedNamespace{
		Name:   ns,
		KeyLen: n,
	}
}

// Serialize to bytes
func (vn *VersionedNamespace) Serialize() []byte {
	return byteutil.Must(json.Marshal(vn))
}

// Deserialize from bytes
func (vn *VersionedNamespace) Deserialize(data []byte) error {
	return json.Unmarshal(data, vn)
}
