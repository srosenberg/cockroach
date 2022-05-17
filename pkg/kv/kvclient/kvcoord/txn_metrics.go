package kvcoord

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
)

type TxnMetrics struct {
	Aborts          *metric.Counter
	Commits         *metric.Counter
	Commits1PC      *metric.Counter
	ParallelCommits *metric.Counter
	CommitWaits     *metric.Counter

	RefreshSuccess                *metric.Counter
	RefreshFail                   *metric.Counter
	RefreshFailWithCondensedSpans *metric.Counter
	RefreshMemoryLimitExceeded    *metric.Counter
	RefreshAutoRetries            *metric.Counter

	Durations *metric.Histogram

	TxnsWithCondensedIntents      *metric.Counter
	TxnsWithCondensedIntentsGauge *metric.Gauge
	TxnsRejectedByLockSpanBudget  *metric.Counter

	Restarts *metric.Histogram

	RestartsWriteTooOld            telemetry.CounterWithMetric
	RestartsWriteTooOldMulti       telemetry.CounterWithMetric
	RestartsSerializable           telemetry.CounterWithMetric
	RestartsAsyncWriteFailure      telemetry.CounterWithMetric
	RestartsCommitDeadlineExceeded telemetry.CounterWithMetric
	RestartsReadWithinUncertainty  telemetry.CounterWithMetric
	RestartsTxnAborted             telemetry.CounterWithMetric
	RestartsTxnPush                telemetry.CounterWithMetric
	RestartsUnknown                telemetry.CounterWithMetric

	RollbacksFailed      *metric.Counter
	AsyncRollbacksFailed *metric.Counter
}

var (
	metaAbortsRates = metric.Metadata{
		Name:        "txn.aborts",
		Help:        "Number of aborted KV transactions",
		Measurement: "KV Transactions",
		Unit:        metric.Unit_COUNT,
	}
	metaCommitsRates = metric.Metadata{
		Name:        "txn.commits",
		Help:        "Number of committed KV transactions (including 1PC)",
		Measurement: "KV Transactions",
		Unit:        metric.Unit_COUNT,
	}
	metaCommits1PCRates = metric.Metadata{
		Name:        "txn.commits1PC",
		Help:        "Number of KV transaction one-phase commit attempts",
		Measurement: "KV Transactions",
		Unit:        metric.Unit_COUNT,
	}
	metaParallelCommitsRates = metric.Metadata{
		Name:        "txn.parallelcommits",
		Help:        "Number of KV transaction parallel commit attempts",
		Measurement: "KV Transactions",
		Unit:        metric.Unit_COUNT,
	}
	metaCommitWaitCount = metric.Metadata{
		Name: "txn.commit_waits",
		Help: "Number of KV transactions that had to commit-wait on commit " +
			"in order to ensure linearizability. This generally happens to " +
			"transactions writing to global ranges.",
		Measurement: "KV Transactions",
		Unit:        metric.Unit_COUNT,
	}
	metaRefreshSuccess = metric.Metadata{
		Name: "txn.refresh.success",
		Help: "Number of successful transaction refreshes. A refresh may be " +
			"preemptive or reactive. A reactive refresh is performed after a " +
			"request throws an error because a refresh is needed for it to " +
			"succeed. In these cases, the request will be re-issued as an " +
			"auto-retry (see txn.refresh.auto_retries) after the refresh succeeds.",
		Measurement: "Refreshes",
		Unit:        metric.Unit_COUNT,
	}
	metaRefreshFail = metric.Metadata{
		Name:        "txn.refresh.fail",
		Help:        "Number of failed transaction refreshes",
		Measurement: "Refreshes",
		Unit:        metric.Unit_COUNT,
	}
	metaRefreshFailWithCondensedSpans = metric.Metadata{
		Name: "txn.refresh.fail_with_condensed_spans",
		Help: "Number of failed refreshes for transactions whose read " +
			"tracking lost fidelity because of condensing. Such a failure " +
			"could be a false conflict. Failures counted here are also counted " +
			"in txn.refresh.fail, and the respective transactions are also counted in " +
			"txn.refresh.memory_limit_exceeded.",
		Measurement: "Refreshes",
		Unit:        metric.Unit_COUNT,
	}
	metaRefreshMemoryLimitExceeded = metric.Metadata{
		Name: "txn.refresh.memory_limit_exceeded",
		Help: "Number of transaction which exceed the refresh span bytes limit, causing " +
			"their read spans to be condensed",
		Measurement: "Transactions",
		Unit:        metric.Unit_COUNT,
	}
	metaRefreshAutoRetries = metric.Metadata{
		Name:        "txn.refresh.auto_retries",
		Help:        "Number of request retries after successful refreshes",
		Measurement: "Retries",
		Unit:        metric.Unit_COUNT,
	}
	metaDurationsHistograms = metric.Metadata{
		Name:        "txn.durations",
		Help:        "KV transaction durations",
		Measurement: "KV Txn Duration",
		Unit:        metric.Unit_NANOSECONDS,
	}
	metaTxnsWithCondensedIntentSpans = metric.Metadata{
		Name: "txn.condensed_intent_spans",
		Help: "KV transactions that have exceeded their intent tracking " +
			"memory budget (kv.transaction.max_intents_bytes). See also " +
			"txn.condensed_intent_spans_gauge for a gauge of such transactions currently running.",
		Measurement: "KV Transactions",
		Unit:        metric.Unit_COUNT,
	}
	metaTxnsWithCondensedIntentSpansGauge = metric.Metadata{
		Name: "txn.condensed_intent_spans_gauge",
		Help: "KV transactions currently running that have exceeded their intent tracking " +
			"memory budget (kv.transaction.max_intents_bytes). See also txn.condensed_intent_spans " +
			"for a perpetual counter/rate.",
		Measurement: "KV Transactions",
		Unit:        metric.Unit_COUNT,
	}
	metaTxnsRejectedByLockSpanBudget = metric.Metadata{
		Name: "txn.condensed_intent_spans_rejected",
		Help: "KV transactions that have been aborted because they exceeded their intent tracking " +
			"memory budget (kv.transaction.max_intents_bytes). " +
			"Rejection is caused by kv.transaction.reject_over_max_intents_budget.",
		Measurement: "KV Transactions",
		Unit:        metric.Unit_COUNT,
	}

	metaRestartsHistogram = metric.Metadata{
		Name:        "txn.restarts",
		Help:        "Number of restarted KV transactions",
		Measurement: "KV Transactions",
		Unit:        metric.Unit_COUNT,
	}

	metaRestartsWriteTooOld = metric.Metadata{
		Name:        "txn.restarts.writetooold",
		Help:        "Number of restarts due to a concurrent writer committing first",
		Measurement: "Restarted Transactions",
		Unit:        metric.Unit_COUNT,
	}
	metaRestartsWriteTooOldMulti = metric.Metadata{
		Name:        "txn.restarts.writetoooldmulti",
		Help:        "Number of restarts due to multiple concurrent writers committing first",
		Measurement: "Restarted Transactions",
		Unit:        metric.Unit_COUNT,
	}
	metaRestartsSerializable = metric.Metadata{
		Name:        "txn.restarts.serializable",
		Help:        "Number of restarts due to a forwarded commit timestamp and isolation=SERIALIZABLE",
		Measurement: "Restarted Transactions",
		Unit:        metric.Unit_COUNT,
	}
	metaRestartsPossibleReplay = metric.Metadata{
		Name:        "txn.restarts.possiblereplay",
		Help:        "Number of restarts due to possible replays of command batches at the storage layer",
		Measurement: "Restarted Transactions",
		Unit:        metric.Unit_COUNT,
	}
	metaRestartsAsyncWriteFailure = metric.Metadata{
		Name:        "txn.restarts.asyncwritefailure",
		Help:        "Number of restarts due to async consensus writes that failed to leave intents",
		Measurement: "Restarted Transactions",
		Unit:        metric.Unit_COUNT,
	}
	metaRestartsCommitDeadlineExceeded = metric.Metadata{
		Name:        "txn.restarts.commitdeadlineexceeded",
		Help:        "Number of restarts due to a transaction exceeding its deadline",
		Measurement: "Restarted Transactions",
		Unit:        metric.Unit_COUNT,
	}
	metaRestartsReadWithinUncertainty = metric.Metadata{
		Name:        "txn.restarts.readwithinuncertainty",
		Help:        "Number of restarts due to reading a new value within the uncertainty interval",
		Measurement: "Restarted Transactions",
		Unit:        metric.Unit_COUNT,
	}
	metaRestartsTxnAborted = metric.Metadata{
		Name:        "txn.restarts.txnaborted",
		Help:        "Number of restarts due to an abort by a concurrent transaction (usually due to deadlock)",
		Measurement: "Restarted Transactions",
		Unit:        metric.Unit_COUNT,
	}

	metaRestartsTxnPush = metric.Metadata{
		Name:        "txn.restarts.txnpush",
		Help:        "Number of restarts due to a transaction push failure",
		Measurement: "Restarted Transactions",
		Unit:        metric.Unit_COUNT,
	}
	metaRestartsUnknown = metric.Metadata{
		Name:        "txn.restarts.unknown",
		Help:        "Number of restarts due to a unknown reasons",
		Measurement: "Restarted Transactions",
		Unit:        metric.Unit_COUNT,
	}
	metaRollbacksFailed = metric.Metadata{
		Name:        "txn.rollbacks.failed",
		Help:        "Number of KV transaction that failed to send final abort",
		Measurement: "KV Transactions",
		Unit:        metric.Unit_COUNT,
	}
	metaAsyncRollbacksFailed = metric.Metadata{
		Name:        "txn.rollbacks.async.failed",
		Help:        "Number of KV transaction that failed to send abort asynchronously which is not always retried",
		Measurement: "KV Transactions",
		Unit:        metric.Unit_COUNT,
	}
)

func MakeTxnMetrics(histogramWindow time.Duration) TxnMetrics {
	__antithesis_instrumentation__.Notify(89244)
	return TxnMetrics{
		Aborts:                         metric.NewCounter(metaAbortsRates),
		Commits:                        metric.NewCounter(metaCommitsRates),
		Commits1PC:                     metric.NewCounter(metaCommits1PCRates),
		ParallelCommits:                metric.NewCounter(metaParallelCommitsRates),
		CommitWaits:                    metric.NewCounter(metaCommitWaitCount),
		RefreshSuccess:                 metric.NewCounter(metaRefreshSuccess),
		RefreshFail:                    metric.NewCounter(metaRefreshFail),
		RefreshFailWithCondensedSpans:  metric.NewCounter(metaRefreshFailWithCondensedSpans),
		RefreshMemoryLimitExceeded:     metric.NewCounter(metaRefreshMemoryLimitExceeded),
		RefreshAutoRetries:             metric.NewCounter(metaRefreshAutoRetries),
		Durations:                      metric.NewLatency(metaDurationsHistograms, histogramWindow),
		TxnsWithCondensedIntents:       metric.NewCounter(metaTxnsWithCondensedIntentSpans),
		TxnsWithCondensedIntentsGauge:  metric.NewGauge(metaTxnsWithCondensedIntentSpansGauge),
		TxnsRejectedByLockSpanBudget:   metric.NewCounter(metaTxnsRejectedByLockSpanBudget),
		Restarts:                       metric.NewHistogram(metaRestartsHistogram, histogramWindow, 100, 3),
		RestartsWriteTooOld:            telemetry.NewCounterWithMetric(metaRestartsWriteTooOld),
		RestartsWriteTooOldMulti:       telemetry.NewCounterWithMetric(metaRestartsWriteTooOldMulti),
		RestartsSerializable:           telemetry.NewCounterWithMetric(metaRestartsSerializable),
		RestartsAsyncWriteFailure:      telemetry.NewCounterWithMetric(metaRestartsAsyncWriteFailure),
		RestartsCommitDeadlineExceeded: telemetry.NewCounterWithMetric(metaRestartsCommitDeadlineExceeded),
		RestartsReadWithinUncertainty:  telemetry.NewCounterWithMetric(metaRestartsReadWithinUncertainty),
		RestartsTxnAborted:             telemetry.NewCounterWithMetric(metaRestartsTxnAborted),
		RestartsTxnPush:                telemetry.NewCounterWithMetric(metaRestartsTxnPush),
		RestartsUnknown:                telemetry.NewCounterWithMetric(metaRestartsUnknown),
		RollbacksFailed:                metric.NewCounter(metaRollbacksFailed),
		AsyncRollbacksFailed:           metric.NewCounter(metaAsyncRollbacksFailed),
	}
}
