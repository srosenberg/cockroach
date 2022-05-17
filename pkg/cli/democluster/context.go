package democluster

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/cockroach/pkg/cli/clicfg"
	"github.com/cockroachdb/cockroach/pkg/workload"
)

type Context struct {
	CliCtx *clicfg.Context

	NumNodes int

	SQLPoolMemorySize int64

	CacheSize int64

	DisableTelemetry bool

	DisableLicenseAcquisition bool

	NoExampleDatabase bool

	RunWorkload bool

	WorkloadGenerator workload.Generator

	WorkloadMaxQPS int

	Localities DemoLocalityList

	GeoPartitionedReplicas bool

	SimulateLatency bool

	DefaultKeySize int

	DefaultCALifetime time.Duration

	DefaultCertLifetime time.Duration

	Insecure bool

	SQLPort int

	HTTPPort int

	ListeningURLFile string

	Multitenant bool
}

func (demoCtx *Context) IsInteractive() bool {
	__antithesis_instrumentation__.Notify(31941)
	return demoCtx.CliCtx != nil && func() bool {
		__antithesis_instrumentation__.Notify(31942)
		return demoCtx.CliCtx.IsInteractive == true
	}() == true
}
