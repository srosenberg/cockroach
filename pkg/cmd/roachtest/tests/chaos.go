package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

type ChaosTimer interface {
	Timing() (time.Duration, time.Duration)
}

type Periodic struct {
	Period, DownTime time.Duration
}

func (p Periodic) Timing() (time.Duration, time.Duration) {
	__antithesis_instrumentation__.Notify(46519)
	return p.Period, p.DownTime
}

type Chaos struct {
	Timer ChaosTimer

	Target func() option.NodeListOption

	Stopper <-chan time.Time

	DrainAndQuit bool

	ChaosEventCh chan ChaosEvent
}

type ChaosEventType uint64

const (
	ChaosEventTypePreShutdown ChaosEventType = iota

	ChaosEventTypeShutdownComplete

	ChaosEventTypePreStartup

	ChaosEventTypeStartupComplete

	ChaosEventTypeStart

	ChaosEventTypeEnd
)

type ChaosEvent struct {
	Type   ChaosEventType
	Target option.NodeListOption
	Time   time.Time
}

func (ch *Chaos) sendEvent(t ChaosEventType, target option.NodeListOption) {
	__antithesis_instrumentation__.Notify(46520)
	if ch.ChaosEventCh != nil {
		__antithesis_instrumentation__.Notify(46521)
		ch.ChaosEventCh <- ChaosEvent{
			Type:   t,
			Target: target,
			Time:   timeutil.Now(),
		}
	} else {
		__antithesis_instrumentation__.Notify(46522)
	}
}

func (ch *Chaos) Runner(
	c cluster.Cluster, t test.Test, m cluster.Monitor,
) func(context.Context) error {
	__antithesis_instrumentation__.Notify(46523)
	return func(ctx context.Context) (err error) {
		__antithesis_instrumentation__.Notify(46524)
		l, err := t.L().ChildLogger("CHAOS")
		if err != nil {
			__antithesis_instrumentation__.Notify(46528)
			return err
		} else {
			__antithesis_instrumentation__.Notify(46529)
		}
		__antithesis_instrumentation__.Notify(46525)
		defer func() {
			__antithesis_instrumentation__.Notify(46530)
			ch.sendEvent(ChaosEventTypeEnd, nil)
			if ch.ChaosEventCh != nil {
				__antithesis_instrumentation__.Notify(46532)
				close(ch.ChaosEventCh)
			} else {
				__antithesis_instrumentation__.Notify(46533)
			}
			__antithesis_instrumentation__.Notify(46531)
			l.Printf("chaos stopping: %v", err)
		}()
		__antithesis_instrumentation__.Notify(46526)
		t := timeutil.Timer{}
		{
			__antithesis_instrumentation__.Notify(46534)
			p, _ := ch.Timer.Timing()
			t.Reset(p)
		}
		__antithesis_instrumentation__.Notify(46527)
		ch.sendEvent(ChaosEventTypeStart, nil)
		for {
			__antithesis_instrumentation__.Notify(46535)
			select {
			case <-ch.Stopper:
				__antithesis_instrumentation__.Notify(46540)
				return nil
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(46541)
				return ctx.Err()
			case <-t.C:
				__antithesis_instrumentation__.Notify(46542)
				t.Read = true
			}
			__antithesis_instrumentation__.Notify(46536)

			period, downTime := ch.Timer.Timing()

			target := ch.Target()
			m.ExpectDeaths(int32(len(target)))

			ch.sendEvent(ChaosEventTypePreShutdown, target)
			if ch.DrainAndQuit {
				__antithesis_instrumentation__.Notify(46543)
				l.Printf("stopping and draining %v\n", target)
				stopOpts := option.DefaultStopOpts()
				stopOpts.RoachprodOpts.Sig = 15
				stopOpts.RoachtestOpts.Worker = true
				if err := c.StopE(ctx, l, stopOpts, target); err != nil {
					__antithesis_instrumentation__.Notify(46544)
					return errors.Wrapf(err, "could not stop node %s", target)
				} else {
					__antithesis_instrumentation__.Notify(46545)
				}
			} else {
				__antithesis_instrumentation__.Notify(46546)
				l.Printf("killing %v\n", target)
				stopOpts := option.DefaultStopOpts()
				stopOpts.RoachtestOpts.Worker = true
				if err := c.StopE(ctx, l, stopOpts, target); err != nil {
					__antithesis_instrumentation__.Notify(46547)
					return errors.Wrapf(err, "could not stop node %s", target)
				} else {
					__antithesis_instrumentation__.Notify(46548)
				}
			}
			__antithesis_instrumentation__.Notify(46537)
			ch.sendEvent(ChaosEventTypeShutdownComplete, target)

			select {
			case <-ch.Stopper:
				__antithesis_instrumentation__.Notify(46549)

				l.Printf("restarting %v (chaos is done)\n", target)
				startOpts := option.DefaultStartOpts()
				startOpts.RoachtestOpts.Worker = true
				if err := c.StartE(ctx, l, startOpts, install.MakeClusterSettings(), target); err != nil {
					__antithesis_instrumentation__.Notify(46554)
					return errors.Wrapf(err, "could not restart node %s", target)
				} else {
					__antithesis_instrumentation__.Notify(46555)
				}
				__antithesis_instrumentation__.Notify(46550)
				return nil
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(46551)

				l.Printf("restarting %v (chaos is done)\n", target)

				tCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				startOpts := option.DefaultStartOpts()
				startOpts.RoachtestOpts.Worker = true
				if err := c.StartE(tCtx, l, startOpts, install.MakeClusterSettings(), target); err != nil {
					__antithesis_instrumentation__.Notify(46556)
					return errors.Wrapf(err, "could not restart node %s", target)
				} else {
					__antithesis_instrumentation__.Notify(46557)
				}
				__antithesis_instrumentation__.Notify(46552)
				return ctx.Err()
			case <-time.After(downTime):
				__antithesis_instrumentation__.Notify(46553)
			}
			__antithesis_instrumentation__.Notify(46538)
			l.Printf("restarting %v after %s of downtime\n", target, downTime)
			t.Reset(period)
			ch.sendEvent(ChaosEventTypePreStartup, target)
			startOpts := option.DefaultStartOpts()
			startOpts.RoachtestOpts.Worker = true
			if err := c.StartE(ctx, l, startOpts, install.MakeClusterSettings(), target); err != nil {
				__antithesis_instrumentation__.Notify(46558)
				return errors.Wrapf(err, "could not restart node %s", target)
			} else {
				__antithesis_instrumentation__.Notify(46559)
			}
			__antithesis_instrumentation__.Notify(46539)
			ch.sendEvent(ChaosEventTypeStartupComplete, target)
		}
	}
}
