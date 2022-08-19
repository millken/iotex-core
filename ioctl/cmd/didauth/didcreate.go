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

// _didGenerateCmd represents the generate command
var _didCreateCmd = &cobra.Command{
	Use: "create",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		err := create()
		return output.PrintError(err)
	},
}

func create() error {
	d, err := db.CreateKVStore(db.DefaultConfig, didDBPath)
	if err != nil {
		return err
	}
	if err = d.Start(context.Background()); err != nil {
		return err
	}
	defer d.Stop(context.Background())

	private, _ := crypto.GenerateKey()

	addr := private.PublicKey().Address()
	if addr == nil {
		return output.NewError(output.ConvertError, "failed to convert public key into address", nil)
	}
	if err = d.Put("did:io", []byte(addr.Hex()), private.Bytes()); err != nil {
		return err
	}
	doc := newDIDDoc()
	doc.ID = DIDPrefix + addr.Hex()
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
	// fmt.Printf("did: did:io:%s\n", addr.Hex())
	// fmt.Printf("publick key: %s\n", fmt.Sprintf("%x", private.PublicKey().Bytes()))
	fmt.Printf("private key: %s\n", fmt.Sprintf("%x", private.Bytes()))
	return nil
}
