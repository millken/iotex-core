package didauth

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/iotexproject/go-pkgs/crypto"
	"github.com/spf13/cobra"

	"github.com/iotexproject/iotex-core/db"
	"github.com/iotexproject/iotex-core/ioctl/output"
)

var _didGetCmd = &cobra.Command{
	Use:  "get",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		err := get(args)
		return output.PrintError(err)
	},
}

func get(args []string) error {
	d, err := db.CreateKVStore(db.DefaultConfig, didDBPath)
	if err != nil {
		return err
	}
	if err = d.Start(context.Background()); err != nil {
		return err
	}
	defer d.Stop(context.Background())

	addr := args[0]

	v, err := d.Get("did:io", []byte(addr))
	if err != nil {
		return err
	}
	private, _ := crypto.BytesToPrivateKey(v)
	doc := newDIDDoc()
	doc.ID = DIDPrefix + addr
	authentication := authenticationStruct{
		ID:           doc.ID + DIDOwner,
		Type:         DIDAuthType,
		Controller:   doc.ID,
		PublicKeyHex: fmt.Sprintf("%x", private.PublicKey().Bytes()),
	}
	doc.Authentication = append(doc.Authentication, authentication)
	msg, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return output.NewError(output.ConvertError, "", err)
	}
	fmt.Printf("%s\n", msg)
	fmt.Printf("publick key: %s\n", fmt.Sprintf("%x", private.PublicKey().Bytes()))
	fmt.Printf("private key: %s\n", fmt.Sprintf("%x", private.Bytes()))
	return nil
}
