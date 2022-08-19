package didauth

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/iotexproject/go-pkgs/hash"
	"github.com/iotexproject/iotex-core/ioctl/output"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var IssueCmd = &cobra.Command{
	Use: "issue",
}

var _IssueClaimCmd = &cobra.Command{
	Use:  "claim",
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		err := issueClaim(args)
		return output.PrintError(err)
	},
}

func issueClaim(args []string) error {
	pKey, err := getPrivateKey(args[0])
	if err != nil {
		return err
	}

	data, err := os.ReadFile(args[1])
	if err != nil {
		return err
	}
	shash := hash.Hash256b(data)
	sig, err := pKey.Sign(shash[:])
	if err != nil {
		return errors.Wrapf(err, "failed to sign data")
	}
	var vv map[string]interface{}
	if err = json.Unmarshal(data, &vv); err != nil {
		return errors.Wrapf(err, "failed to unmarshal data")
	}
	vv["proof"] = map[string]string{
		"type":           "Secp256k1",
		"signatureValue": fmt.Sprintf("%x", sig),
	}
	data, err = json.MarshalIndent(vv, "", "  ")
	if err != nil {
		return errors.Wrapf(err, "failed to marshal data")
	}
	fmt.Printf("%s\n", data)
	return nil
}

func init() {
	IssueCmd.AddCommand(_IssueClaimCmd)

}
