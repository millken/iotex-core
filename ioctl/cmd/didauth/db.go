package didauth

import (
	"context"

	"github.com/iotexproject/go-pkgs/crypto"
	"github.com/iotexproject/iotex-core/db"
	"github.com/pkg/errors"
)

func getPrivateKey(did string) (crypto.PrivateKey, error) {
	d, err := db.CreateKVStore(db.DefaultConfig, didDBPath)
	if err != nil {
		return nil, err
	}
	if err = d.Start(context.Background()); err != nil {
		return nil, err
	}
	defer d.Stop(context.Background())

	v, err := d.Get("did:io", []byte(did))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get private key for did: %s", did)
	}
	return crypto.BytesToPrivateKey(v)
}
