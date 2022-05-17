package kvcoord

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
)

type TxnCoordSenderFactory struct {
	log.AmbientContext

	st                     *cluster.Settings
	wrapped                kv.Sender
	clock                  *hlc.Clock
	heartbeatInterval      time.Duration
	linearizable           bool
	stopper                *stop.Stopper
	metrics                TxnMetrics
	condensedIntentsEveryN log.EveryN

	testingKnobs ClientTestingKnobs
}

var _ kv.TxnSenderFactory = &TxnCoordSenderFactory{}

type TxnCoordSenderFactoryConfig struct {
	AmbientCtx log.AmbientContext

	Settings *cluster.Settings
	Clock    *hlc.Clock
	Stopper  *stop.Stopper

	HeartbeatInterval time.Duration
	Linearizable      bool
	Metrics           TxnMetrics

	TestingKnobs ClientTestingKnobs
}

func NewTxnCoordSenderFactory(
	cfg TxnCoordSenderFactoryConfig, wrapped kv.Sender,
) *TxnCoordSenderFactory {
	__antithesis_instrumentation__.Notify(88396)
	tcf := &TxnCoordSenderFactory{
		AmbientContext:         cfg.AmbientCtx,
		st:                     cfg.Settings,
		wrapped:                wrapped,
		clock:                  cfg.Clock,
		stopper:                cfg.Stopper,
		linearizable:           cfg.Linearizable,
		heartbeatInterval:      cfg.HeartbeatInterval,
		metrics:                cfg.Metrics,
		condensedIntentsEveryN: log.Every(time.Second),
		testingKnobs:           cfg.TestingKnobs,
	}
	if tcf.st == nil {
		__antithesis_instrumentation__.Notify(88400)
		tcf.st = cluster.MakeTestingClusterSettings()
	} else {
		__antithesis_instrumentation__.Notify(88401)
	}
	__antithesis_instrumentation__.Notify(88397)
	if tcf.heartbeatInterval == 0 {
		__antithesis_instrumentation__.Notify(88402)
		tcf.heartbeatInterval = base.DefaultTxnHeartbeatInterval
	} else {
		__antithesis_instrumentation__.Notify(88403)
	}
	__antithesis_instrumentation__.Notify(88398)
	if tcf.metrics == (TxnMetrics{}) {
		__antithesis_instrumentation__.Notify(88404)
		tcf.metrics = MakeTxnMetrics(metric.TestSampleInterval)
	} else {
		__antithesis_instrumentation__.Notify(88405)
	}
	__antithesis_instrumentation__.Notify(88399)
	return tcf
}

func (tcf *TxnCoordSenderFactory) RootTransactionalSender(
	txn *roachpb.Transaction, pri roachpb.UserPriority,
) kv.TxnSender {
	__antithesis_instrumentation__.Notify(88406)
	return newRootTxnCoordSender(tcf, txn, pri)
}

func (tcf *TxnCoordSenderFactory) LeafTransactionalSender(
	tis *roachpb.LeafTxnInputState,
) kv.TxnSender {
	__antithesis_instrumentation__.Notify(88407)
	return newLeafTxnCoordSender(tcf, tis)
}

func (tcf *TxnCoordSenderFactory) NonTransactionalSender() kv.Sender {
	__antithesis_instrumentation__.Notify(88408)
	return tcf.wrapped
}

func (tcf *TxnCoordSenderFactory) Metrics() TxnMetrics {
	__antithesis_instrumentation__.Notify(88409)
	return tcf.metrics
}
