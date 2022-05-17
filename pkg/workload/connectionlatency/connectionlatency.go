package connectionlatency

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"github.com/cockroachdb/cockroach/pkg/workload/histogram"
	"github.com/jackc/pgx/v4"
	"github.com/spf13/pflag"
)

type connectionLatency struct {
	flags     workload.Flags
	connFlags *workload.ConnFlags

	locality string
}

func init() {
	workload.Register(connectionLatencyMeta)
}

var connectionLatencyMeta = workload.Meta{
	Name:        `connectionlatency`,
	Description: `Testing Connection Latencies`,
	Version:     `1.0.0`,
	New: func() workload.Generator {
		__antithesis_instrumentation__.Notify(693916)
		c := &connectionLatency{}
		c.flags.FlagSet = pflag.NewFlagSet(`connectionlatency`, pflag.ContinueOnError)
		c.flags.StringVar(&c.locality, `locality`, ``, `Which locality is the workload running in? (east,west,central)`)
		c.connFlags = workload.NewConnFlags(&c.flags)
		return c
	},
}

func (connectionLatency) Meta() workload.Meta {
	__antithesis_instrumentation__.Notify(693917)
	return connectionLatencyMeta
}

func (c *connectionLatency) Flags() workload.Flags {
	__antithesis_instrumentation__.Notify(693918)
	return c.flags
}

func (connectionLatency) Tables() []workload.Table {
	__antithesis_instrumentation__.Notify(693919)
	return nil
}

func (c *connectionLatency) Ops(
	ctx context.Context, urls []string, reg *histogram.Registry,
) (workload.QueryLoad, error) {
	__antithesis_instrumentation__.Notify(693920)
	ql := workload.QueryLoad{}
	_, err := workload.SanitizeUrls(c, c.connFlags.DBOverride, urls)
	if err != nil {
		__antithesis_instrumentation__.Notify(693923)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(693924)
	}
	__antithesis_instrumentation__.Notify(693921)

	for _, url := range urls {
		__antithesis_instrumentation__.Notify(693925)
		op := &connectionOp{
			url:   url,
			hists: reg.GetHandle(),
		}

		conn, err := pgx.Connect(ctx, url)
		if err != nil {
			__antithesis_instrumentation__.Notify(693928)
			return workload.QueryLoad{}, err
		} else {
			__antithesis_instrumentation__.Notify(693929)
		}
		__antithesis_instrumentation__.Notify(693926)

		var locality string
		err = conn.QueryRow(ctx, "SHOW LOCALITY").Scan(&locality)
		if err != nil {
			__antithesis_instrumentation__.Notify(693930)
			return workload.QueryLoad{}, err
		} else {
			__antithesis_instrumentation__.Notify(693931)
		}
		__antithesis_instrumentation__.Notify(693927)

		localitySplit := strings.Split(locality, "zone=")
		locality = localitySplit[len(localitySplit)-1]

		op.connectFrom = c.locality
		op.connectTo = locality
		ql.WorkerFns = append(ql.WorkerFns, op.run)
	}
	__antithesis_instrumentation__.Notify(693922)
	return ql, nil
}

type connectionOp struct {
	url         string
	hists       *histogram.Histograms
	connectFrom string
	connectTo   string
}

func (o *connectionOp) run(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(693932)
	start := timeutil.Now()
	conn, err := pgx.Connect(ctx, o.url)
	if err != nil {
		__antithesis_instrumentation__.Notify(693936)
		return err
	} else {
		__antithesis_instrumentation__.Notify(693937)
	}
	__antithesis_instrumentation__.Notify(693933)
	defer func() {
		__antithesis_instrumentation__.Notify(693938)
		if err := conn.Close(ctx); err != nil {
			__antithesis_instrumentation__.Notify(693939)
			log.Warningf(ctx, "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(693940)
		}
	}()
	__antithesis_instrumentation__.Notify(693934)
	elapsed := timeutil.Since(start)
	o.hists.Get(fmt.Sprintf(`connect-from-%s-to-%s`, o.connectFrom, o.connectTo)).Record(elapsed)

	if _, err = conn.Exec(ctx, "SELECT 1"); err != nil {
		__antithesis_instrumentation__.Notify(693941)
		return err
	} else {
		__antithesis_instrumentation__.Notify(693942)
	}
	__antithesis_instrumentation__.Notify(693935)

	elapsed = timeutil.Since(start)
	o.hists.Get(`select`).Record(elapsed)
	return nil
}
