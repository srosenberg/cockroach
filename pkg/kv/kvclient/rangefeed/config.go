package rangefeed

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
)

type Option interface {
	set(*config)
}

type config struct {
	scanConfig
	retryOptions       retry.Options
	onInitialScanDone  OnInitialScanDone
	withInitialScan    bool
	onInitialScanError OnInitialScanError

	useRowTimestampInInitialScan bool

	withDiff             bool
	onUnrecoverableError OnUnrecoverableError
	onCheckpoint         OnCheckpoint
	onFrontierAdvance    OnFrontierAdvance
	onSSTable            OnSSTable
	extraPProfLabels     []string
}

type scanConfig struct {
	scanParallelism func() int

	targetScanBytes int64

	mon *mon.BytesMonitor

	onSpanDone OnScanCompleted

	retryBehavior ScanRetryBehavior
}

type optionFunc func(*config)

func (o optionFunc) set(c *config) { __antithesis_instrumentation__.Notify(89635); o(c) }

type OnInitialScanDone func(ctx context.Context)

type OnInitialScanError func(ctx context.Context, err error) (shouldFail bool)

type OnUnrecoverableError func(ctx context.Context, err error)

func WithInitialScan(f OnInitialScanDone) Option {
	__antithesis_instrumentation__.Notify(89636)
	return optionFunc(func(c *config) {
		__antithesis_instrumentation__.Notify(89637)
		c.withInitialScan = true
		c.onInitialScanDone = f
	})
}

func WithOnInitialScanError(f OnInitialScanError) Option {
	__antithesis_instrumentation__.Notify(89638)
	return optionFunc(func(c *config) {
		__antithesis_instrumentation__.Notify(89639)
		c.onInitialScanError = f
	})
}

func WithRowTimestampInInitialScan(shouldUse bool) Option {
	__antithesis_instrumentation__.Notify(89640)
	return optionFunc(func(c *config) {
		__antithesis_instrumentation__.Notify(89641)
		c.useRowTimestampInInitialScan = shouldUse
	})
}

func WithOnInternalError(f OnUnrecoverableError) Option {
	__antithesis_instrumentation__.Notify(89642)
	return optionFunc(func(c *config) {
		__antithesis_instrumentation__.Notify(89643)
		c.onUnrecoverableError = f
	})
}

func WithDiff(withDiff bool) Option {
	__antithesis_instrumentation__.Notify(89644)
	return optionFunc(func(c *config) {
		__antithesis_instrumentation__.Notify(89645)
		c.withDiff = withDiff
	})
}

func WithRetry(options retry.Options) Option {
	__antithesis_instrumentation__.Notify(89646)
	return optionFunc(func(c *config) {
		__antithesis_instrumentation__.Notify(89647)
		c.retryOptions = options
	})
}

type OnCheckpoint func(ctx context.Context, checkpoint *roachpb.RangeFeedCheckpoint)

func WithOnCheckpoint(f OnCheckpoint) Option {
	__antithesis_instrumentation__.Notify(89648)
	return optionFunc(func(c *config) {
		__antithesis_instrumentation__.Notify(89649)
		c.onCheckpoint = f
	})
}

type OnSSTable func(ctx context.Context, sst *roachpb.RangeFeedSSTable)

func WithOnSSTable(f OnSSTable) Option {
	__antithesis_instrumentation__.Notify(89650)
	return optionFunc(func(c *config) {
		__antithesis_instrumentation__.Notify(89651)
		c.onSSTable = f
	})
}

type OnFrontierAdvance func(ctx context.Context, timestamp hlc.Timestamp)

func WithOnFrontierAdvance(f OnFrontierAdvance) Option {
	__antithesis_instrumentation__.Notify(89652)
	return optionFunc(func(c *config) {
		__antithesis_instrumentation__.Notify(89653)
		c.onFrontierAdvance = f
	})
}

func initConfig(c *config, options []Option) {
	__antithesis_instrumentation__.Notify(89654)
	*c = config{}
	for _, o := range options {
		__antithesis_instrumentation__.Notify(89656)
		o.set(c)
	}
	__antithesis_instrumentation__.Notify(89655)

	if c.targetScanBytes == 0 {
		__antithesis_instrumentation__.Notify(89657)
		c.targetScanBytes = 1 << 19
	} else {
		__antithesis_instrumentation__.Notify(89658)
	}
}

func WithInitialScanParallelismFn(parallelismFn func() int) Option {
	__antithesis_instrumentation__.Notify(89659)
	return optionFunc(func(c *config) {
		__antithesis_instrumentation__.Notify(89660)
		c.scanParallelism = parallelismFn
	})
}

func WithTargetScanBytes(target int64) Option {
	__antithesis_instrumentation__.Notify(89661)
	return optionFunc(func(c *config) {
		__antithesis_instrumentation__.Notify(89662)
		c.targetScanBytes = target
	})
}

func WithMemoryMonitor(mon *mon.BytesMonitor) Option {
	__antithesis_instrumentation__.Notify(89663)
	return optionFunc(func(c *config) {
		__antithesis_instrumentation__.Notify(89664)
		c.mon = mon
	})
}

type OnScanCompleted func(ctx context.Context, sp roachpb.Span) error

func WithOnScanCompleted(fn OnScanCompleted) Option {
	__antithesis_instrumentation__.Notify(89665)
	return optionFunc(func(c *config) {
		__antithesis_instrumentation__.Notify(89666)
		c.onSpanDone = fn
	})
}

type ScanRetryBehavior int

const (
	ScanRetryAll ScanRetryBehavior = iota

	ScanRetryRemaining
)

func WithScanRetryBehavior(b ScanRetryBehavior) Option {
	__antithesis_instrumentation__.Notify(89667)
	return optionFunc(func(c *config) {
		__antithesis_instrumentation__.Notify(89668)
		c.retryBehavior = b
	})
}

func WithPProfLabel(key, value string) Option {
	__antithesis_instrumentation__.Notify(89669)
	return optionFunc(func(c *config) {
		__antithesis_instrumentation__.Notify(89670)
		c.extraPProfLabels = append(c.extraPProfLabels, key, value)
	})
}
