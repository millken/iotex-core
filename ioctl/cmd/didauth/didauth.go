package didauth

import (
	"github.com/spf13/cobra"

	"github.com/iotexproject/iotex-core/ioctl/config"
)

const (
	didDBPath = "did.db"
)

// Multi-language support
var (
	DIDCmdShorts = map[config.Language]string{
		config.English: "DID command",
		config.Chinese: "DID command",
	}
	_flagEndpoint = map[config.Language]string{
		config.English: "set endpoint for once",
		config.Chinese: "一次设置端点",
	}
	_flagInsecure = map[config.Language]string{
		config.English: "insecure connection for once",
		config.Chinese: "一次不安全连接",
	}
)

// DIDCmd represents the DID command
var DIDCmd = &cobra.Command{
	Use:   "didauth",
	Short: config.TranslateInLang(DIDCmdShorts, config.UILanguage),
}

func init() {
	DIDCmd.AddCommand(_didCreateCmd)
	DIDCmd.AddCommand(_didGetCmd)
	DIDCmd.AddCommand(IssueCmd)
	DIDCmd.PersistentFlags().StringVar(&config.ReadConfig.Endpoint, "endpoint",
		config.ReadConfig.Endpoint, config.TranslateInLang(_flagEndpoint, config.UILanguage))
	DIDCmd.PersistentFlags().BoolVar(&config.Insecure, "insecure", config.Insecure, config.TranslateInLang(_flagInsecure, config.UILanguage))
}
