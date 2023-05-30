package nodestats

import (
	"time"

	"github.com/iotexproject/iotex-core/pkg/lifecycle"
	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/memory"
)

// APIReport is the report of an RPC call
type APIReport struct {
	Method       string
	HandlingTime time.Duration
	Success      bool
}

type apiMethodStats struct {
	Successes          int
	Errors             int
	AvgTimeOfErrors    int64
	AvgTimeOfSuccesses int64
	MaxTimeOfError     int64
	MaxTimeOfSuccess   int64
	TotalSize          int64
}

// AvgSize returns the average size of the rpc call
func (m *apiMethodStats) AvgSize() int64 {
	if m.Successes+m.Errors == 0 {
		return 0
	}
	return m.TotalSize / int64(m.Successes+m.Errors)
}

// RPCLocalStats is the interface for getting RPC stats
type RPCLocalStats interface {
	ReportCall(report APIReport, size int64)
	BuildReport() string
}

// NodeStats is the interface for getting node stats
type NodeStats interface {
	lifecycle.StartStopper
}

// SytemStats is the interface for getting system stats
type SytemStats interface {
	GetMemory() (*memory.Stats, error)
	GetCPU() (*cpu.Stats, error)
	BuildReport() string
}
