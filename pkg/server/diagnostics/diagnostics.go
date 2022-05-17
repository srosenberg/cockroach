package diagnostics

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math/rand"
	"net/url"
	"strconv"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/diagnostics/diagnosticspb"
	"github.com/cockroachdb/cockroach/pkg/util/cloudinfo"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/system"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
)

var updatesURL *url.URL

const defaultUpdatesURL = `https://register.cockroachdb.com/api/clusters/updates`

var reportingURL *url.URL

const defaultReportingURL = `https://register.cockroachdb.com/api/clusters/report`

func init() {
	var err error
	updatesURL, err = url.Parse(
		envutil.EnvOrDefaultString("COCKROACH_UPDATE_CHECK_URL", defaultUpdatesURL),
	)
	if err != nil {
		panic(err)
	}
	reportingURL, err = url.Parse(
		envutil.EnvOrDefaultString("COCKROACH_USAGE_REPORT_URL", defaultReportingURL),
	)
	if err != nil {
		panic(err)
	}
}

type TestingKnobs struct {
	OverrideUpdatesURL **url.URL

	OverrideReportingURL **url.URL
}

type ClusterInfo struct {
	StorageClusterID uuid.UUID
	LogicalClusterID uuid.UUID
	TenantID         roachpb.TenantID
	IsInsecure       bool
	IsInternal       bool
}

func addInfoToURL(
	url *url.URL,
	clusterInfo *ClusterInfo,
	env *diagnosticspb.Environment,
	nodeID roachpb.NodeID,
	sqlInfo *diagnosticspb.SQLInstanceInfo,
) *url.URL {
	__antithesis_instrumentation__.Notify(190599)
	if url == nil {
		__antithesis_instrumentation__.Notify(190602)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(190603)
	}
	__antithesis_instrumentation__.Notify(190600)
	result := *url
	q := result.Query()

	if nodeID != 0 {
		__antithesis_instrumentation__.Notify(190604)
		q.Set("nodeid", strconv.Itoa(int(nodeID)))
	} else {
		__antithesis_instrumentation__.Notify(190605)
	}
	__antithesis_instrumentation__.Notify(190601)

	b := env.Build
	q.Set("sqlid", strconv.Itoa(int(sqlInfo.SQLInstanceID)))
	q.Set("uptime", strconv.Itoa(int(sqlInfo.Uptime)))
	q.Set("licensetype", env.LicenseType)
	q.Set("version", b.Tag)
	q.Set("platform", b.Platform)
	q.Set("uuid", clusterInfo.StorageClusterID.String())
	q.Set("logical_uuid", clusterInfo.LogicalClusterID.String())
	q.Set("tenantid", clusterInfo.TenantID.String())
	q.Set("insecure", strconv.FormatBool(clusterInfo.IsInsecure))
	q.Set("internal", strconv.FormatBool(clusterInfo.IsInternal))
	q.Set("buildchannel", b.Channel)
	q.Set("envchannel", b.EnvChannel)
	result.RawQuery = q.Encode()
	return &result
}

func addJitter(d time.Duration) time.Duration {
	__antithesis_instrumentation__.Notify(190606)
	const jitterSeconds = 120
	j := time.Duration(rand.Intn(jitterSeconds*2)-jitterSeconds) * time.Second
	return d + j
}

var populateMutex syncutil.Mutex

func populateHardwareInfo(ctx context.Context, e *diagnosticspb.Environment) {
	__antithesis_instrumentation__.Notify(190607)

	populateMutex.Lock()
	defer populateMutex.Unlock()

	if platform, family, version, err := host.PlatformInformation(); err == nil {
		__antithesis_instrumentation__.Notify(190613)
		e.Os.Family = family
		e.Os.Platform = platform
		e.Os.Version = version
	} else {
		__antithesis_instrumentation__.Notify(190614)
	}
	__antithesis_instrumentation__.Notify(190608)

	if virt, role, err := host.Virtualization(); err == nil && func() bool {
		__antithesis_instrumentation__.Notify(190615)
		return role == "guest" == true
	}() == true {
		__antithesis_instrumentation__.Notify(190616)
		e.Hardware.Virtualization = virt
	} else {
		__antithesis_instrumentation__.Notify(190617)
	}
	__antithesis_instrumentation__.Notify(190609)

	if m, err := mem.VirtualMemory(); err == nil {
		__antithesis_instrumentation__.Notify(190618)
		e.Hardware.Mem.Available = m.Available
		e.Hardware.Mem.Total = m.Total
	} else {
		__antithesis_instrumentation__.Notify(190619)
	}
	__antithesis_instrumentation__.Notify(190610)

	e.Hardware.Cpu.Numcpu = int32(system.NumCPU())
	if cpus, err := cpu.InfoWithContext(ctx); err == nil && func() bool {
		__antithesis_instrumentation__.Notify(190620)
		return len(cpus) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(190621)
		e.Hardware.Cpu.Sockets = int32(len(cpus))
		c := cpus[0]
		e.Hardware.Cpu.Cores = c.Cores
		e.Hardware.Cpu.Model = c.ModelName
		e.Hardware.Cpu.Mhz = float32(c.Mhz)
		e.Hardware.Cpu.Features = c.Flags
	} else {
		__antithesis_instrumentation__.Notify(190622)
	}
	__antithesis_instrumentation__.Notify(190611)

	if l, err := load.AvgWithContext(ctx); err == nil {
		__antithesis_instrumentation__.Notify(190623)
		e.Hardware.Loadavg15 = float32(l.Load15)
	} else {
		__antithesis_instrumentation__.Notify(190624)
	}
	__antithesis_instrumentation__.Notify(190612)

	e.Hardware.Provider, e.Hardware.InstanceClass = cloudinfo.GetInstanceClass(ctx)
	e.Topology.Provider, e.Topology.Region = cloudinfo.GetInstanceRegion(ctx)
}
