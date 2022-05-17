package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"net/http"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/ts/tspb"
	"github.com/cockroachdb/cockroach/pkg/util/httputil"
)

type tsQueryType int

const (
	total tsQueryType = iota

	rate
)

type tsQuery struct {
	name      string
	queryType tsQueryType
}

func mustGetMetrics(
	t test.Test, adminURL string, start, end time.Time, tsQueries []tsQuery,
) tspb.TimeSeriesQueryResponse {
	__antithesis_instrumentation__.Notify(52108)
	response, err := getMetrics(adminURL, start, end, tsQueries)
	if err != nil {
		__antithesis_instrumentation__.Notify(52110)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(52111)
	}
	__antithesis_instrumentation__.Notify(52109)
	return response
}

func getMetrics(
	adminURL string, start, end time.Time, tsQueries []tsQuery,
) (tspb.TimeSeriesQueryResponse, error) {
	__antithesis_instrumentation__.Notify(52112)
	url := "http://" + adminURL + "/ts/query"
	queries := make([]tspb.Query, len(tsQueries))
	for i := 0; i < len(tsQueries); i++ {
		__antithesis_instrumentation__.Notify(52114)
		switch tsQueries[i].queryType {
		case total:
			__antithesis_instrumentation__.Notify(52115)
			queries[i] = tspb.Query{
				Name:             tsQueries[i].name,
				Downsampler:      tspb.TimeSeriesQueryAggregator_AVG.Enum(),
				SourceAggregator: tspb.TimeSeriesQueryAggregator_SUM.Enum(),
			}
		case rate:
			__antithesis_instrumentation__.Notify(52116)
			queries[i] = tspb.Query{
				Name:             tsQueries[i].name,
				Downsampler:      tspb.TimeSeriesQueryAggregator_AVG.Enum(),
				SourceAggregator: tspb.TimeSeriesQueryAggregator_SUM.Enum(),
				Derivative:       tspb.TimeSeriesQueryDerivative_NON_NEGATIVE_DERIVATIVE.Enum(),
			}
		default:
			__antithesis_instrumentation__.Notify(52117)
			panic("unexpected")
		}
	}
	__antithesis_instrumentation__.Notify(52113)
	request := tspb.TimeSeriesQueryRequest{
		StartNanos: start.UnixNano(),
		EndNanos:   end.UnixNano(),

		SampleNanos: (1 * time.Minute).Nanoseconds(),
		Queries:     queries,
	}
	var response tspb.TimeSeriesQueryResponse
	err := httputil.PostJSON(http.Client{Timeout: 500 * time.Millisecond}, url, &request, &response)
	return response, err

}

func verifyTxnPerSecond(
	ctx context.Context,
	c cluster.Cluster,
	t test.Test,
	adminNode option.NodeListOption,
	start, end time.Time,
	txnTarget, maxPercentTimeUnderTarget float64,
) {
	__antithesis_instrumentation__.Notify(52118)

	adminUIAddrs, err := c.ExternalAdminUIAddr(ctx, t.L(), adminNode)
	if err != nil {
		__antithesis_instrumentation__.Notify(52122)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(52123)
	}
	__antithesis_instrumentation__.Notify(52119)
	adminURL := adminUIAddrs[0]
	response := mustGetMetrics(t, adminURL, start, end, []tsQuery{
		{name: "cr.node.txn.commits", queryType: rate},
		{name: "cr.node.txn.commits", queryType: total},
	})

	perMinute := response.Results[0].Datapoints[2:]
	cumulative := response.Results[1].Datapoints[2:]

	totalTxns := cumulative[len(cumulative)-1].Value - cumulative[0].Value
	avgTxnPerSec := totalTxns / float64(end.Sub(start)/time.Second)

	if avgTxnPerSec < txnTarget {
		__antithesis_instrumentation__.Notify(52124)
		t.Fatalf("average txns per second %f was under target %f", avgTxnPerSec, txnTarget)
	} else {
		__antithesis_instrumentation__.Notify(52125)
		t.L().Printf("average txns per second: %f", avgTxnPerSec)
	}
	__antithesis_instrumentation__.Notify(52120)

	minutesBelowTarget := 0.0
	for _, dp := range perMinute {
		__antithesis_instrumentation__.Notify(52126)
		if dp.Value < txnTarget {
			__antithesis_instrumentation__.Notify(52127)
			minutesBelowTarget++
		} else {
			__antithesis_instrumentation__.Notify(52128)
		}
	}
	__antithesis_instrumentation__.Notify(52121)
	if perc := minutesBelowTarget / float64(len(perMinute)); perc > maxPercentTimeUnderTarget {
		__antithesis_instrumentation__.Notify(52129)
		t.Fatalf(
			"spent %f%% of time below target of %f txn/s, wanted no more than %f%%",
			perc*100, txnTarget, maxPercentTimeUnderTarget*100,
		)
	} else {
		__antithesis_instrumentation__.Notify(52130)
		t.L().Printf("spent %f%% of time below target of %f txn/s", perc*100, txnTarget)
	}
}

func verifyLookupsPerSec(
	ctx context.Context,
	c cluster.Cluster,
	t test.Test,
	adminNode option.NodeListOption,
	start, end time.Time,
	rangeLookupsTarget float64,
) {
	__antithesis_instrumentation__.Notify(52131)

	adminUIAddrs, err := c.ExternalAdminUIAddr(ctx, t.L(), adminNode)
	if err != nil {
		__antithesis_instrumentation__.Notify(52133)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(52134)
	}
	__antithesis_instrumentation__.Notify(52132)
	adminURL := adminUIAddrs[0]
	response := mustGetMetrics(t, adminURL, start, end, []tsQuery{
		{name: "cr.node.distsender.rangelookups", queryType: rate},
	})

	perMinute := response.Results[0].Datapoints[2:]

	for _, dp := range perMinute {
		__antithesis_instrumentation__.Notify(52135)
		if dp.Value > rangeLookupsTarget {
			__antithesis_instrumentation__.Notify(52136)
			t.Fatalf("Found minute interval with %f lookup/sec above target of %f lookup/sec\n", dp.Value, rangeLookupsTarget)
		} else {
			__antithesis_instrumentation__.Notify(52137)
			t.L().Printf("Found minute interval with %f lookup/sec\n", dp.Value)
		}
	}
}
