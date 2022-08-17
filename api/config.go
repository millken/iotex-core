package api

import (
	"github.com/iotexproject/iotex-core/pkg/tracer"
	"github.com/iotexproject/iotex-core/pkg/unit"
)

var (
	//DefaultConfig is the default config for the api
	DefaultConfig = Config{
		UseRDS:        false,
		GRPCPort:      14014,
		HTTPPort:      15014,
		WebSocketPort: 16014,
		TpsWindow:     10,
		GasStation: GasStation{
			SuggestBlockWindow: 20,
			DefaultGas:         uint64(unit.Qev),
			Percentile:         60,
		},
		RangeQueryLimit: 1000,
	}
)

// Config is the api service config
type Config struct {
	UseRDS          bool          `yaml:"useRDS"`
	GRPCPort        int           `yaml:"port"`
	HTTPPort        int           `yaml:"web3port"`
	WebSocketPort   int           `yaml:"webSocketPort"`
	RedisCacheURL   string        `yaml:"redisCacheURL"`
	TpsWindow       int           `yaml:"tpsWindow"`
	GasStation      GasStation    `yaml:"gasStation"`
	RangeQueryLimit uint64        `yaml:"rangeQueryLimit"`
	Tracer          tracer.Config `yaml:"tracer"`
}

// GasStation is the gas station config
type GasStation struct {
	SuggestBlockWindow int    `yaml:"suggestBlockWindow"`
	DefaultGas         uint64 `yaml:"defaultGas"`
	Percentile         int    `yaml:"Percentile"`
}
