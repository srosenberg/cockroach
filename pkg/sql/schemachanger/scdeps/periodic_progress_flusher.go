package scdeps

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/backfill"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scexec"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"golang.org/x/sync/errgroup"
)

func NewPeriodicProgressFlusher(
	checkpointIntervalFn func() time.Duration, fractionIntervalFn func() time.Duration,
) scexec.PeriodicProgressFlusher {
	__antithesis_instrumentation__.Notify(580732)
	return &periodicProgressFlusher{
		clock:              timeutil.DefaultTimeSource{},
		checkpointInterval: checkpointIntervalFn,
		fractionInterval:   fractionIntervalFn,
	}
}

func newPeriodicProgressFlusherForIndexBackfill(
	settings *cluster.Settings,
) scexec.PeriodicProgressFlusher {
	__antithesis_instrumentation__.Notify(580733)
	return NewPeriodicProgressFlusher(
		func() time.Duration {
			__antithesis_instrumentation__.Notify(580734)
			return backfill.IndexBackfillCheckpointInterval.Get(&settings.SV)

		},
		func() time.Duration {
			__antithesis_instrumentation__.Notify(580735)

			const fractionInterval = 10 * time.Second
			return fractionInterval
		},
	)
}

type periodicProgressFlusher struct {
	clock                                timeutil.TimeSource
	checkpointInterval, fractionInterval func() time.Duration
}

func (p *periodicProgressFlusher) StartPeriodicUpdates(
	ctx context.Context, tracker scexec.BackfillProgressFlusher,
) (stop func() error) {
	__antithesis_instrumentation__.Notify(580736)
	stopCh := make(chan struct{})
	runPeriodicWrite := func(
		ctx context.Context,
		write func(context.Context) error,
		interval func() time.Duration,
	) error {
		__antithesis_instrumentation__.Notify(580740)
		timer := p.clock.NewTimer()
		defer timer.Stop()
		for {
			__antithesis_instrumentation__.Notify(580741)
			timer.Reset(interval())
			select {
			case <-stopCh:
				__antithesis_instrumentation__.Notify(580742)
				return nil
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(580743)
				return ctx.Err()
			case <-timer.Ch():
				__antithesis_instrumentation__.Notify(580744)
				timer.MarkRead()
				if err := write(ctx); err != nil {
					__antithesis_instrumentation__.Notify(580745)
					return err
				} else {
					__antithesis_instrumentation__.Notify(580746)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(580737)
	var g errgroup.Group
	g.Go(func() error {
		__antithesis_instrumentation__.Notify(580747)
		return runPeriodicWrite(
			ctx, tracker.FlushFractionCompleted, p.fractionInterval)
	})
	__antithesis_instrumentation__.Notify(580738)
	g.Go(func() error {
		__antithesis_instrumentation__.Notify(580748)
		return runPeriodicWrite(
			ctx, tracker.FlushCheckpoint, p.checkpointInterval)
	})
	__antithesis_instrumentation__.Notify(580739)
	toClose := stopCh
	return func() error {
		__antithesis_instrumentation__.Notify(580749)
		if toClose != nil {
			__antithesis_instrumentation__.Notify(580751)
			close(toClose)
			toClose = nil
		} else {
			__antithesis_instrumentation__.Notify(580752)
		}
		__antithesis_instrumentation__.Notify(580750)
		return g.Wait()
	}
}
