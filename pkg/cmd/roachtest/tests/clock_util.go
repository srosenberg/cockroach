package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/logger"
)

func isAlive(db *gosql.DB, l *logger.Logger) bool {
	__antithesis_instrumentation__.Notify(46708)

	_ = db.Ping()
	if err := db.Ping(); err != nil {
		__antithesis_instrumentation__.Notify(46710)
		l.Printf("isAlive returned err=%v (%T)", err, err)
	} else {
		__antithesis_instrumentation__.Notify(46711)
		return true
	}
	__antithesis_instrumentation__.Notify(46709)
	return false
}

func dbUnixEpoch(db *gosql.DB) (float64, error) {
	__antithesis_instrumentation__.Notify(46712)
	var epoch float64
	if err := db.QueryRow("SELECT now()::DECIMAL").Scan(&epoch); err != nil {
		__antithesis_instrumentation__.Notify(46714)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(46715)
	}
	__antithesis_instrumentation__.Notify(46713)
	return epoch, nil
}

type offsetInjector struct {
	t        test.Test
	c        cluster.Cluster
	deployed bool
}

func (oi *offsetInjector) deploy(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(46716)
	if err := oi.c.RunE(ctx, oi.c.All(), "test -x ./bumptime"); err == nil {
		__antithesis_instrumentation__.Notify(46723)
		oi.deployed = true
		return nil
	} else {
		__antithesis_instrumentation__.Notify(46724)
	}
	__antithesis_instrumentation__.Notify(46717)

	if err := oi.c.Install(ctx, oi.t.L(), oi.c.All(), "ntp"); err != nil {
		__antithesis_instrumentation__.Notify(46725)
		return err
	} else {
		__antithesis_instrumentation__.Notify(46726)
	}
	__antithesis_instrumentation__.Notify(46718)
	if err := oi.c.Install(ctx, oi.t.L(), oi.c.All(), "gcc"); err != nil {
		__antithesis_instrumentation__.Notify(46727)
		return err
	} else {
		__antithesis_instrumentation__.Notify(46728)
	}
	__antithesis_instrumentation__.Notify(46719)
	if err := oi.c.RunE(ctx, oi.c.All(), "sudo", "service", "ntp", "stop"); err != nil {
		__antithesis_instrumentation__.Notify(46729)
		return err
	} else {
		__antithesis_instrumentation__.Notify(46730)
	}
	__antithesis_instrumentation__.Notify(46720)
	if err := oi.c.RunE(ctx, oi.c.All(),
		"curl", "--retry", "3", "--fail", "--show-error", "-kO",
		"https://raw.githubusercontent.com/cockroachdb/jepsen/master/cockroachdb/resources/bumptime.c",
	); err != nil {
		__antithesis_instrumentation__.Notify(46731)
		return err
	} else {
		__antithesis_instrumentation__.Notify(46732)
	}
	__antithesis_instrumentation__.Notify(46721)
	if err := oi.c.RunE(ctx, oi.c.All(),
		"gcc", "bumptime.c", "-o", "bumptime", "&&", "rm bumptime.c",
	); err != nil {
		__antithesis_instrumentation__.Notify(46733)
		return err
	} else {
		__antithesis_instrumentation__.Notify(46734)
	}
	__antithesis_instrumentation__.Notify(46722)
	oi.deployed = true
	return nil
}

func (oi *offsetInjector) offset(ctx context.Context, nodeID int, s time.Duration) {
	__antithesis_instrumentation__.Notify(46735)
	if !oi.deployed {
		__antithesis_instrumentation__.Notify(46737)
		oi.t.Fatal("Offset injector must be deployed before injecting a clock offset")
	} else {
		__antithesis_instrumentation__.Notify(46738)
	}
	__antithesis_instrumentation__.Notify(46736)

	oi.c.Run(
		ctx,
		oi.c.Node(nodeID),
		fmt.Sprintf("sudo ./bumptime %f", float64(s)/float64(time.Millisecond)),
	)
}

func (oi *offsetInjector) recover(ctx context.Context, nodeID int) {
	__antithesis_instrumentation__.Notify(46739)
	if !oi.deployed {
		__antithesis_instrumentation__.Notify(46741)
		oi.t.Fatal("Offset injector must be deployed before recovering from clock offsets")
	} else {
		__antithesis_instrumentation__.Notify(46742)
	}
	__antithesis_instrumentation__.Notify(46740)

	syncCmds := [][]string{
		{"sudo", "service", "ntp", "stop"},
		{"sudo", "ntpdate", "-u", "time.google.com"},
		{"sudo", "service", "ntp", "start"},
	}
	for _, cmd := range syncCmds {
		__antithesis_instrumentation__.Notify(46743)
		oi.c.Run(
			ctx,
			oi.c.Node(nodeID),
			cmd...,
		)
	}
}

func newOffsetInjector(t test.Test, c cluster.Cluster) *offsetInjector {
	__antithesis_instrumentation__.Notify(46744)
	return &offsetInjector{t: t, c: c}
}
