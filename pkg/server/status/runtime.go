package status

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"os"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/cockroachdb/cockroach/pkg/build"
	"github.com/cockroachdb/cockroach/pkg/util/cgroups"
	"github.com/cockroachdb/cockroach/pkg/util/goschedstats"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/elastic/gosigar"
	"github.com/shirou/gopsutil/v3/net"
)

var (
	metaCgoCalls = metric.Metadata{
		Name:        "sys.cgocalls",
		Help:        "Total number of cgo calls",
		Measurement: "cgo Calls",
		Unit:        metric.Unit_COUNT,
	}
	metaGoroutines = metric.Metadata{
		Name:        "sys.goroutines",
		Help:        "Current number of goroutines",
		Measurement: "goroutines",
		Unit:        metric.Unit_COUNT,
	}
	metaRunnableGoroutinesPerCPU = metric.Metadata{
		Name:        "sys.runnable.goroutines.per.cpu",
		Help:        "Average number of goroutines that are waiting to run, normalized by number of cores",
		Measurement: "goroutines",
		Unit:        metric.Unit_COUNT,
	}
	metaGoAllocBytes = metric.Metadata{
		Name:        "sys.go.allocbytes",
		Help:        "Current bytes of memory allocated by go",
		Measurement: "Memory",
		Unit:        metric.Unit_BYTES,
	}
	metaGoTotalBytes = metric.Metadata{
		Name:        "sys.go.totalbytes",
		Help:        "Total bytes of memory allocated by go, but not released",
		Measurement: "Memory",
		Unit:        metric.Unit_BYTES,
	}
	metaCgoAllocBytes = metric.Metadata{
		Name:        "sys.cgo.allocbytes",
		Help:        "Current bytes of memory allocated by cgo",
		Measurement: "Memory",
		Unit:        metric.Unit_BYTES,
	}
	metaCgoTotalBytes = metric.Metadata{
		Name:        "sys.cgo.totalbytes",
		Help:        "Total bytes of memory allocated by cgo, but not released",
		Measurement: "Memory",
		Unit:        metric.Unit_BYTES,
	}
	metaGCCount = metric.Metadata{
		Name:        "sys.gc.count",
		Help:        "Total number of GC runs",
		Measurement: "GC Runs",
		Unit:        metric.Unit_COUNT,
	}
	metaGCPauseNS = metric.Metadata{
		Name:        "sys.gc.pause.ns",
		Help:        "Total GC pause",
		Measurement: "GC Pause",
		Unit:        metric.Unit_NANOSECONDS,
	}
	metaGCPausePercent = metric.Metadata{
		Name:        "sys.gc.pause.percent",
		Help:        "Current GC pause percentage",
		Measurement: "GC Pause",
		Unit:        metric.Unit_PERCENT,
	}
	metaCPUUserNS = metric.Metadata{
		Name:        "sys.cpu.user.ns",
		Help:        "Total user cpu time",
		Measurement: "CPU Time",
		Unit:        metric.Unit_NANOSECONDS,
	}
	metaCPUUserPercent = metric.Metadata{
		Name:        "sys.cpu.user.percent",
		Help:        "Current user cpu percentage",
		Measurement: "CPU Time",
		Unit:        metric.Unit_PERCENT,
	}
	metaCPUSysNS = metric.Metadata{
		Name:        "sys.cpu.sys.ns",
		Help:        "Total system cpu time",
		Measurement: "CPU Time",
		Unit:        metric.Unit_NANOSECONDS,
	}
	metaCPUSysPercent = metric.Metadata{
		Name:        "sys.cpu.sys.percent",
		Help:        "Current system cpu percentage",
		Measurement: "CPU Time",
		Unit:        metric.Unit_PERCENT,
	}
	metaCPUCombinedPercentNorm = metric.Metadata{
		Name:        "sys.cpu.combined.percent-normalized",
		Help:        "Current user+system cpu percentage, normalized 0-1 by number of cores",
		Measurement: "CPU Time",
		Unit:        metric.Unit_PERCENT,
	}
	metaCPUNowNS = metric.Metadata{
		Name:        "sys.cpu.now.ns",
		Help:        "Number of nanoseconds elapsed since January 1, 1970 UTC",
		Measurement: "CPU Time",
		Unit:        metric.Unit_NANOSECONDS,
	}
	metaRSSBytes = metric.Metadata{
		Name:        "sys.rss",
		Help:        "Current process RSS",
		Measurement: "RSS",
		Unit:        metric.Unit_BYTES,
	}
	metaFDOpen = metric.Metadata{
		Name:        "sys.fd.open",
		Help:        "Process open file descriptors",
		Measurement: "File Descriptors",
		Unit:        metric.Unit_COUNT,
	}
	metaFDSoftLimit = metric.Metadata{
		Name:        "sys.fd.softlimit",
		Help:        "Process open FD soft limit",
		Measurement: "File Descriptors",
		Unit:        metric.Unit_COUNT,
	}
	metaUptime = metric.Metadata{
		Name:        "sys.uptime",
		Help:        "Process uptime",
		Measurement: "Uptime",
		Unit:        metric.Unit_SECONDS,
	}

	metaHostDiskReadCount = metric.Metadata{
		Name:        "sys.host.disk.read.count",
		Unit:        metric.Unit_COUNT,
		Measurement: "Operations",
		Help:        "Disk read operations across all disks since this process started",
	}
	metaHostDiskReadBytes = metric.Metadata{
		Name:        "sys.host.disk.read.bytes",
		Unit:        metric.Unit_BYTES,
		Measurement: "Bytes",
		Help:        "Bytes read from all disks since this process started",
	}
	metaHostDiskReadTime = metric.Metadata{
		Name:        "sys.host.disk.read.time",
		Unit:        metric.Unit_NANOSECONDS,
		Measurement: "Time",
		Help:        "Time spent reading from all disks since this process started",
	}
	metaHostDiskWriteCount = metric.Metadata{
		Name:        "sys.host.disk.write.count",
		Unit:        metric.Unit_COUNT,
		Measurement: "Operations",
		Help:        "Disk write operations across all disks since this process started",
	}
	metaHostDiskWriteBytes = metric.Metadata{
		Name:        "sys.host.disk.write.bytes",
		Unit:        metric.Unit_BYTES,
		Measurement: "Bytes",
		Help:        "Bytes written to all disks since this process started",
	}
	metaHostDiskWriteTime = metric.Metadata{
		Name:        "sys.host.disk.write.time",
		Unit:        metric.Unit_NANOSECONDS,
		Measurement: "Time",
		Help:        "Time spent writing to all disks since this process started",
	}
	metaHostDiskIOTime = metric.Metadata{
		Name:        "sys.host.disk.io.time",
		Unit:        metric.Unit_NANOSECONDS,
		Measurement: "Time",
		Help:        "Time spent reading from or writing to all disks since this process started",
	}
	metaHostDiskWeightedIOTime = metric.Metadata{
		Name:        "sys.host.disk.weightedio.time",
		Unit:        metric.Unit_NANOSECONDS,
		Measurement: "Time",
		Help:        "Weighted time spent reading from or writing to to all disks since this process started",
	}
	metaHostIopsInProgress = metric.Metadata{
		Name:        "sys.host.disk.iopsinprogress",
		Unit:        metric.Unit_COUNT,
		Measurement: "Operations",
		Help:        "IO operations currently in progress on this host",
	}
	metaHostNetRecvBytes = metric.Metadata{
		Name:        "sys.host.net.recv.bytes",
		Unit:        metric.Unit_BYTES,
		Measurement: "Bytes",
		Help:        "Bytes received on all network interfaces since this process started",
	}
	metaHostNetRecvPackets = metric.Metadata{
		Name:        "sys.host.net.recv.packets",
		Unit:        metric.Unit_COUNT,
		Measurement: "Packets",
		Help:        "Packets received on all network interfaces since this process started",
	}
	metaHostNetSendBytes = metric.Metadata{
		Name:        "sys.host.net.send.bytes",
		Unit:        metric.Unit_BYTES,
		Measurement: "Bytes",
		Help:        "Bytes sent on all network interfaces since this process started",
	}
	metaHostNetSendPackets = metric.Metadata{
		Name:        "sys.host.net.send.packets",
		Unit:        metric.Unit_COUNT,
		Measurement: "Packets",
		Help:        "Packets sent on all network interfaces since this process started",
	}
)

var getCgoMemStats func(context.Context) (uint, uint, error)

type RuntimeStatSampler struct {
	clock *hlc.Clock

	startTimeNanos int64

	last struct {
		now         int64
		utime       int64
		stime       int64
		cgoCall     int64
		gcCount     int64
		gcPauseTime uint64
		disk        diskStats
		net         net.IOCountersStat
		runnableSum float64
	}

	initialDiskCounters diskStats
	initialNetCounters  net.IOCountersStat

	fdUsageNotImplemented bool

	CgoCalls                 *metric.Gauge
	Goroutines               *metric.Gauge
	RunnableGoroutinesPerCPU *metric.GaugeFloat64
	GoAllocBytes             *metric.Gauge
	GoTotalBytes             *metric.Gauge
	CgoAllocBytes            *metric.Gauge
	CgoTotalBytes            *metric.Gauge
	GcCount                  *metric.Gauge
	GcPauseNS                *metric.Gauge
	GcPausePercent           *metric.GaugeFloat64

	CPUUserNS              *metric.Gauge
	CPUUserPercent         *metric.GaugeFloat64
	CPUSysNS               *metric.Gauge
	CPUSysPercent          *metric.GaugeFloat64
	CPUCombinedPercentNorm *metric.GaugeFloat64
	CPUNowNS               *metric.Gauge

	RSSBytes *metric.Gauge

	FDOpen      *metric.Gauge
	FDSoftLimit *metric.Gauge

	HostDiskReadBytes      *metric.Gauge
	HostDiskReadCount      *metric.Gauge
	HostDiskReadTime       *metric.Gauge
	HostDiskWriteBytes     *metric.Gauge
	HostDiskWriteCount     *metric.Gauge
	HostDiskWriteTime      *metric.Gauge
	HostDiskIOTime         *metric.Gauge
	HostDiskWeightedIOTime *metric.Gauge
	IopsInProgress         *metric.Gauge
	HostNetRecvBytes       *metric.Gauge
	HostNetRecvPackets     *metric.Gauge
	HostNetSendBytes       *metric.Gauge
	HostNetSendPackets     *metric.Gauge

	Uptime         *metric.Gauge
	BuildTimestamp *metric.Gauge
}

func NewRuntimeStatSampler(ctx context.Context, clock *hlc.Clock) *RuntimeStatSampler {
	__antithesis_instrumentation__.Notify(235629)

	info := build.GetInfo()
	timestamp, err := info.Timestamp()
	if err != nil {
		__antithesis_instrumentation__.Notify(235633)

		log.Warningf(ctx, "could not parse build timestamp: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(235634)
	}
	__antithesis_instrumentation__.Notify(235630)

	metaBuildTimestamp := metric.Metadata{
		Name:        "build.timestamp",
		Help:        "Build information",
		Measurement: "Build Time",
		Unit:        metric.Unit_TIMESTAMP_SEC,
	}
	metaBuildTimestamp.AddLabel("tag", info.Tag)
	metaBuildTimestamp.AddLabel("go_version", info.GoVersion)

	buildTimestamp := metric.NewGauge(metaBuildTimestamp)
	buildTimestamp.Update(timestamp)

	diskCounters, err := getSummedDiskCounters(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(235635)
		log.Ops.Errorf(ctx, "could not get initial disk IO counters: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(235636)
	}
	__antithesis_instrumentation__.Notify(235631)
	netCounters, err := getSummedNetStats(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(235637)
		log.Ops.Errorf(ctx, "could not get initial disk IO counters: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(235638)
	}
	__antithesis_instrumentation__.Notify(235632)

	rsr := &RuntimeStatSampler{
		clock:                    clock,
		startTimeNanos:           clock.PhysicalNow(),
		initialNetCounters:       netCounters,
		initialDiskCounters:      diskCounters,
		CgoCalls:                 metric.NewGauge(metaCgoCalls),
		Goroutines:               metric.NewGauge(metaGoroutines),
		RunnableGoroutinesPerCPU: metric.NewGaugeFloat64(metaRunnableGoroutinesPerCPU),
		GoAllocBytes:             metric.NewGauge(metaGoAllocBytes),
		GoTotalBytes:             metric.NewGauge(metaGoTotalBytes),
		CgoAllocBytes:            metric.NewGauge(metaCgoAllocBytes),
		CgoTotalBytes:            metric.NewGauge(metaCgoTotalBytes),
		GcCount:                  metric.NewGauge(metaGCCount),
		GcPauseNS:                metric.NewGauge(metaGCPauseNS),
		GcPausePercent:           metric.NewGaugeFloat64(metaGCPausePercent),
		CPUUserNS:                metric.NewGauge(metaCPUUserNS),
		CPUUserPercent:           metric.NewGaugeFloat64(metaCPUUserPercent),
		CPUSysNS:                 metric.NewGauge(metaCPUSysNS),
		CPUSysPercent:            metric.NewGaugeFloat64(metaCPUSysPercent),
		CPUCombinedPercentNorm:   metric.NewGaugeFloat64(metaCPUCombinedPercentNorm),
		CPUNowNS:                 metric.NewGauge(metaCPUNowNS),
		RSSBytes:                 metric.NewGauge(metaRSSBytes),
		HostDiskReadBytes:        metric.NewGauge(metaHostDiskReadBytes),
		HostDiskReadCount:        metric.NewGauge(metaHostDiskReadCount),
		HostDiskReadTime:         metric.NewGauge(metaHostDiskReadTime),
		HostDiskWriteBytes:       metric.NewGauge(metaHostDiskWriteBytes),
		HostDiskWriteCount:       metric.NewGauge(metaHostDiskWriteCount),
		HostDiskWriteTime:        metric.NewGauge(metaHostDiskWriteTime),
		HostDiskIOTime:           metric.NewGauge(metaHostDiskIOTime),
		HostDiskWeightedIOTime:   metric.NewGauge(metaHostDiskWeightedIOTime),
		IopsInProgress:           metric.NewGauge(metaHostIopsInProgress),
		HostNetRecvBytes:         metric.NewGauge(metaHostNetRecvBytes),
		HostNetRecvPackets:       metric.NewGauge(metaHostNetRecvPackets),
		HostNetSendBytes:         metric.NewGauge(metaHostNetSendBytes),
		HostNetSendPackets:       metric.NewGauge(metaHostNetSendPackets),
		FDOpen:                   metric.NewGauge(metaFDOpen),
		FDSoftLimit:              metric.NewGauge(metaFDSoftLimit),
		Uptime:                   metric.NewGauge(metaUptime),
		BuildTimestamp:           buildTimestamp,
	}
	rsr.last.disk = rsr.initialDiskCounters
	rsr.last.net = rsr.initialNetCounters
	return rsr
}

type GoMemStats struct {
	runtime.MemStats

	Collected time.Time
}

type CGoMemStats struct {
	CGoAllocatedBytes uint64

	CGoTotalBytes uint64
}

func GetCGoMemStats(ctx context.Context) *CGoMemStats {
	__antithesis_instrumentation__.Notify(235639)
	var cgoAllocated, cgoTotal uint
	if getCgoMemStats != nil {
		__antithesis_instrumentation__.Notify(235641)
		var err error
		cgoAllocated, cgoTotal, err = getCgoMemStats(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(235642)
			log.Warningf(ctx, "problem fetching CGO memory stats: %s; CGO stats will be empty.", err)
		} else {
			__antithesis_instrumentation__.Notify(235643)
		}
	} else {
		__antithesis_instrumentation__.Notify(235644)
	}
	__antithesis_instrumentation__.Notify(235640)
	return &CGoMemStats{
		CGoAllocatedBytes: uint64(cgoAllocated),
		CGoTotalBytes:     uint64(cgoTotal),
	}
}

func (rsr *RuntimeStatSampler) SampleEnvironment(
	ctx context.Context, ms *GoMemStats, cs *CGoMemStats,
) {
	__antithesis_instrumentation__.Notify(235645)

	gc := &debug.GCStats{}
	debug.ReadGCStats(gc)

	numCgoCall := runtime.NumCgoCall()
	numGoroutine := runtime.NumGoroutine()

	pid := os.Getpid()
	mem := gosigar.ProcMem{}
	if err := mem.Get(pid); err != nil {
		__antithesis_instrumentation__.Notify(235651)
		log.Ops.Errorf(ctx, "unable to get mem usage: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(235652)
	}
	__antithesis_instrumentation__.Notify(235646)
	userTimeMillis, sysTimeMillis, err := GetCPUTime(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(235653)
		log.Ops.Errorf(ctx, "unable to get cpu usage: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(235654)
	}
	__antithesis_instrumentation__.Notify(235647)
	cgroupCPU, _ := cgroups.GetCgroupCPU()
	cpuShare := cgroupCPU.CPUShares()

	fds := gosigar.ProcFDUsage{}
	if err := fds.Get(pid); err != nil {
		__antithesis_instrumentation__.Notify(235655)
		if gosigar.IsNotImplemented(err) {
			__antithesis_instrumentation__.Notify(235656)
			if !rsr.fdUsageNotImplemented {
				__antithesis_instrumentation__.Notify(235657)
				rsr.fdUsageNotImplemented = true
				log.Ops.Warningf(ctx, "unable to get file descriptor usage (will not try again): %s", err)
			} else {
				__antithesis_instrumentation__.Notify(235658)
			}
		} else {
			__antithesis_instrumentation__.Notify(235659)
			log.Ops.Errorf(ctx, "unable to get file descriptor usage: %s", err)
		}
	} else {
		__antithesis_instrumentation__.Notify(235660)
	}
	__antithesis_instrumentation__.Notify(235648)

	var deltaDisk diskStats
	diskCounters, err := getSummedDiskCounters(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(235661)
		log.Ops.Warningf(ctx, "problem fetching disk stats: %s; disk stats will be empty.", err)
	} else {
		__antithesis_instrumentation__.Notify(235662)
		deltaDisk = diskCounters
		subtractDiskCounters(&deltaDisk, rsr.last.disk)
		rsr.last.disk = diskCounters
		subtractDiskCounters(&diskCounters, rsr.initialDiskCounters)

		rsr.HostDiskReadBytes.Update(diskCounters.readBytes)
		rsr.HostDiskReadCount.Update(diskCounters.readCount)
		rsr.HostDiskReadTime.Update(int64(diskCounters.readTime))
		rsr.HostDiskWriteBytes.Update(diskCounters.writeBytes)
		rsr.HostDiskWriteCount.Update(diskCounters.writeCount)
		rsr.HostDiskWriteTime.Update(int64(diskCounters.writeTime))
		rsr.HostDiskIOTime.Update(int64(diskCounters.ioTime))
		rsr.HostDiskWeightedIOTime.Update(int64(diskCounters.weightedIOTime))
		rsr.IopsInProgress.Update(diskCounters.iopsInProgress)
	}
	__antithesis_instrumentation__.Notify(235649)

	var deltaNet net.IOCountersStat
	netCounters, err := getSummedNetStats(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(235663)
		log.Ops.Warningf(ctx, "problem fetching net stats: %s; net stats will be empty.", err)
	} else {
		__antithesis_instrumentation__.Notify(235664)
		deltaNet = netCounters
		subtractNetworkCounters(&deltaNet, rsr.last.net)
		rsr.last.net = netCounters
		subtractNetworkCounters(&netCounters, rsr.initialNetCounters)

		rsr.HostNetSendBytes.Update(int64(netCounters.BytesSent))
		rsr.HostNetSendPackets.Update(int64(netCounters.PacketsSent))
		rsr.HostNetRecvBytes.Update(int64(netCounters.BytesRecv))
		rsr.HostNetRecvPackets.Update(int64(netCounters.PacketsRecv))
	}
	__antithesis_instrumentation__.Notify(235650)

	now := rsr.clock.PhysicalNow()
	dur := float64(now - rsr.last.now)

	utime := userTimeMillis * 1e6
	stime := sysTimeMillis * 1e6
	urate := float64(utime-rsr.last.utime) / dur
	srate := float64(stime-rsr.last.stime) / dur
	combinedNormalizedPerc := (srate + urate) / cpuShare
	gcPauseRatio := float64(uint64(gc.PauseTotal)-rsr.last.gcPauseTime) / dur
	runnableSum := goschedstats.CumulativeNormalizedRunnableGoroutines()

	runnableAvg := (runnableSum - rsr.last.runnableSum) * 1e9 / dur
	rsr.last.now = now
	rsr.last.utime = utime
	rsr.last.stime = stime
	rsr.last.gcPauseTime = uint64(gc.PauseTotal)
	rsr.last.runnableSum = runnableSum

	cgoRate := float64((numCgoCall-rsr.last.cgoCall)*int64(time.Second)) / dur
	goStatsStaleness := float32(timeutil.Since(ms.Collected)) / float32(time.Second)
	goTotal := ms.Sys - ms.HeapReleased

	stats := &eventpb.RuntimeStats{
		MemRSSBytes:       mem.Resident,
		GoroutineCount:    uint64(numGoroutine),
		MemStackSysBytes:  ms.StackSys,
		GoAllocBytes:      ms.HeapAlloc,
		GoTotalBytes:      goTotal,
		GoStatsStaleness:  goStatsStaleness,
		HeapFragmentBytes: ms.HeapInuse - ms.HeapAlloc,
		HeapReservedBytes: ms.HeapIdle - ms.HeapReleased,
		HeapReleasedBytes: ms.HeapReleased,
		CGoAllocBytes:     cs.CGoAllocatedBytes,
		CGoTotalBytes:     cs.CGoTotalBytes,
		CGoCallRate:       float32(cgoRate),
		CPUUserPercent:    float32(urate) * 100,
		CPUSysPercent:     float32(srate) * 100,
		GCPausePercent:    float32(gcPauseRatio) * 100,
		GCRunCount:        uint64(gc.NumGC),
		NetHostRecvBytes:  deltaNet.BytesRecv,
		NetHostSendBytes:  deltaNet.BytesSent,
	}

	logStats(ctx, stats)

	rsr.last.cgoCall = numCgoCall
	rsr.last.gcCount = gc.NumGC

	rsr.GoAllocBytes.Update(int64(ms.HeapAlloc))
	rsr.GoTotalBytes.Update(int64(goTotal))
	rsr.CgoCalls.Update(numCgoCall)
	rsr.Goroutines.Update(int64(numGoroutine))
	rsr.RunnableGoroutinesPerCPU.Update(runnableAvg)
	rsr.CgoAllocBytes.Update(int64(cs.CGoAllocatedBytes))
	rsr.CgoTotalBytes.Update(int64(cs.CGoTotalBytes))
	rsr.GcCount.Update(gc.NumGC)
	rsr.GcPauseNS.Update(int64(gc.PauseTotal))
	rsr.GcPausePercent.Update(gcPauseRatio)
	rsr.CPUUserNS.Update(utime)
	rsr.CPUUserPercent.Update(urate)
	rsr.CPUSysNS.Update(stime)
	rsr.CPUSysPercent.Update(srate)
	rsr.CPUCombinedPercentNorm.Update(combinedNormalizedPerc)
	rsr.CPUNowNS.Update(now)
	rsr.FDOpen.Update(int64(fds.Open))
	rsr.FDSoftLimit.Update(int64(fds.SoftLimit))
	rsr.RSSBytes.Update(int64(mem.Resident))
	rsr.Uptime.Update((now - rsr.startTimeNanos) / 1e9)
}

func (rsr *RuntimeStatSampler) GetCPUCombinedPercentNorm() float64 {
	__antithesis_instrumentation__.Notify(235665)
	return rsr.CPUCombinedPercentNorm.Value()
}

type diskStats struct {
	readBytes int64
	readCount int64

	readTime time.Duration

	writeBytes int64
	writeCount int64
	writeTime  time.Duration

	ioTime time.Duration

	weightedIOTime time.Duration

	iopsInProgress int64
}

func getSummedDiskCounters(ctx context.Context) (diskStats, error) {
	__antithesis_instrumentation__.Notify(235666)
	diskCounters, err := getDiskCounters(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(235668)
		return diskStats{}, err
	} else {
		__antithesis_instrumentation__.Notify(235669)
	}
	__antithesis_instrumentation__.Notify(235667)

	return sumDiskCounters(diskCounters), nil
}

func getSummedNetStats(ctx context.Context) (net.IOCountersStat, error) {
	__antithesis_instrumentation__.Notify(235670)
	netCounters, err := net.IOCountersWithContext(ctx, true)
	if err != nil {
		__antithesis_instrumentation__.Notify(235672)
		return net.IOCountersStat{}, err
	} else {
		__antithesis_instrumentation__.Notify(235673)
	}
	__antithesis_instrumentation__.Notify(235671)

	return sumNetworkCounters(netCounters), nil
}

func sumDiskCounters(disksStats []diskStats) diskStats {
	__antithesis_instrumentation__.Notify(235674)
	output := diskStats{}
	for _, stats := range disksStats {
		__antithesis_instrumentation__.Notify(235676)
		output.readBytes += stats.readBytes
		output.readCount += stats.readCount
		output.readTime += stats.readTime

		output.writeBytes += stats.writeBytes
		output.writeCount += stats.writeCount
		output.writeTime += stats.writeTime

		output.ioTime += stats.ioTime
		output.weightedIOTime += stats.weightedIOTime

		output.iopsInProgress += stats.iopsInProgress
	}
	__antithesis_instrumentation__.Notify(235675)
	return output
}

func subtractDiskCounters(from *diskStats, sub diskStats) {
	__antithesis_instrumentation__.Notify(235677)
	from.writeCount -= sub.writeCount
	from.writeBytes -= sub.writeBytes
	from.writeTime -= sub.writeTime

	from.readCount -= sub.readCount
	from.readBytes -= sub.readBytes
	from.readTime -= sub.readTime

	from.ioTime -= sub.ioTime
	from.weightedIOTime -= sub.weightedIOTime
}

func sumNetworkCounters(netCounters []net.IOCountersStat) net.IOCountersStat {
	__antithesis_instrumentation__.Notify(235678)
	output := net.IOCountersStat{}
	for _, counter := range netCounters {
		__antithesis_instrumentation__.Notify(235680)
		output.BytesRecv += counter.BytesRecv
		output.BytesSent += counter.BytesSent
		output.PacketsRecv += counter.PacketsRecv
		output.PacketsSent += counter.PacketsSent
	}
	__antithesis_instrumentation__.Notify(235679)
	return output
}

func subtractNetworkCounters(from *net.IOCountersStat, sub net.IOCountersStat) {
	__antithesis_instrumentation__.Notify(235681)
	from.BytesRecv -= sub.BytesRecv
	from.BytesSent -= sub.BytesSent
	from.PacketsRecv -= sub.PacketsRecv
	from.PacketsSent -= sub.PacketsSent
}

func GetCPUTime(ctx context.Context) (userTimeMillis, sysTimeMillis int64, err error) {
	__antithesis_instrumentation__.Notify(235682)
	pid := os.Getpid()
	cpuTime := gosigar.ProcTime{}
	if err := cpuTime.Get(pid); err != nil {
		__antithesis_instrumentation__.Notify(235684)
		return 0, 0, err
	} else {
		__antithesis_instrumentation__.Notify(235685)
	}
	__antithesis_instrumentation__.Notify(235683)
	return int64(cpuTime.User), int64(cpuTime.Sys), nil
}
