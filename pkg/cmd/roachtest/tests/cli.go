package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"reflect"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cli"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/stretchr/testify/require"
)

func runCLINodeStatus(ctx context.Context, t test.Test, c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(46611)
	c.Put(ctx, t.Cockroach(), "./cockroach")
	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.Range(1, 3))

	db := c.Conn(ctx, t.L(), 1)
	defer db.Close()

	err := WaitFor3XReplication(ctx, t, db)
	require.NoError(t, err)

	lastWords := func(s string) []string {
		__antithesis_instrumentation__.Notify(46615)
		var result []string
		s = cli.ElideInsecureDeprecationNotice(s)
		lines := strings.Split(s, "\n")
		for _, line := range lines {
			__antithesis_instrumentation__.Notify(46617)
			words := strings.Fields(line)
			if n := len(words); n > 0 {
				__antithesis_instrumentation__.Notify(46618)
				result = append(result, words[n-2]+" "+words[n-1])
			} else {
				__antithesis_instrumentation__.Notify(46619)
			}
		}
		__antithesis_instrumentation__.Notify(46616)
		return result
	}
	__antithesis_instrumentation__.Notify(46612)

	nodeStatus := func() (_ string, _ []string, err error) {
		__antithesis_instrumentation__.Notify(46620)
		result, err := c.RunWithDetailsSingleNode(ctx, t.L(), c.Node(1), "./cockroach node status --insecure -p {pgport:1}")
		if err != nil {
			__antithesis_instrumentation__.Notify(46622)
			return "", nil, err
		} else {
			__antithesis_instrumentation__.Notify(46623)
		}
		__antithesis_instrumentation__.Notify(46621)
		return result.Stdout, lastWords(result.Stdout), nil
	}

	{
		__antithesis_instrumentation__.Notify(46624)
		expected := []string{
			"is_available is_live",
			"true true",
			"true true",
			"true true",
		}
		raw, actual, err := nodeStatus()
		if err != nil {
			__antithesis_instrumentation__.Notify(46626)
			t.Fatalf("node status failed: %v\n%s", err, raw)
		} else {
			__antithesis_instrumentation__.Notify(46627)
		}
		__antithesis_instrumentation__.Notify(46625)
		if !reflect.DeepEqual(expected, actual) {
			__antithesis_instrumentation__.Notify(46628)
			t.Fatalf("expected %s, but found %s:\nfrom:\n%s", expected, actual, raw)
		} else {
			__antithesis_instrumentation__.Notify(46629)
		}
	}
	__antithesis_instrumentation__.Notify(46613)

	waitUntil := func(expected []string) {
		__antithesis_instrumentation__.Notify(46630)
		var (
			raw    string
			actual []string
			err    error
		)

		for i := 0; i < 20; i++ {
			__antithesis_instrumentation__.Notify(46632)
			if raw, actual, err = nodeStatus(); err != nil {
				__antithesis_instrumentation__.Notify(46634)
				t.L().Printf("node status failed: %v\n%s", err, raw)
			} else {
				__antithesis_instrumentation__.Notify(46635)
				if reflect.DeepEqual(expected, actual) {
					__antithesis_instrumentation__.Notify(46636)
					break
				} else {
					__antithesis_instrumentation__.Notify(46637)
				}
			}
			__antithesis_instrumentation__.Notify(46633)
			t.L().Printf("not done: %s vs %s\n", expected, actual)
			time.Sleep(time.Second)
		}
		__antithesis_instrumentation__.Notify(46631)
		if !reflect.DeepEqual(expected, actual) {
			__antithesis_instrumentation__.Notify(46638)
			t.Fatalf("expected %s, but found %s from:\n%s", expected, actual, raw)
		} else {
			__antithesis_instrumentation__.Notify(46639)
		}
	}
	__antithesis_instrumentation__.Notify(46614)

	c.Stop(ctx, t.L(), option.DefaultStopOpts(), c.Node(2))
	waitUntil([]string{
		"is_available is_live",
		"true true",
		"false false",
		"true true",
	})

	c.Stop(ctx, t.L(), option.DefaultStopOpts(), c.Node(3))
	waitUntil([]string{
		"is_available is_live",
		"false true",
		"false false",
		"false false",
	})

	c.Stop(ctx, t.L(), option.DefaultStopOpts(), c.Range(1, 2))
	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.Range(1, 2))

	waitUntil([]string{
		"is_available is_live",
		"true true",
		"true true",
		"false false",
	})

	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.Node(3))
}
