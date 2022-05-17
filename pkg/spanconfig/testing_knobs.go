package spanconfig

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
)

type TestingKnobs struct {
	ManagerDisableJobCreation bool

	ManagerCheckJobInterceptor func()

	ManagerCreatedJobInterceptor func(interface{})

	ManagerAfterCheckedReconciliationJobExistsInterceptor func(exists bool)

	JobDisablePersistingCheckpoints bool

	JobDisableInternalRetry bool

	JobOverrideRetryOptions *retry.Options

	JobOnCheckpointInterceptor func() error

	KVSubscriberRangeFeedKnobs base.ModuleTestingKnobs

	StoreKVSubscriberOverride KVSubscriber

	KVAccessorPaginationInterceptor func()

	KVAccessorBatchSizeOverrideFn func() int

	KVAccessorPreCommitMinTSWaitInterceptor func()

	KVAccessorPostCommitDeadlineSetInterceptor func(*kv.Txn)

	SQLWatcherOnEventInterceptor func() error

	SQLWatcherCheckpointNoopsEveryDurationOverride time.Duration

	SplitterStepLogger func(string)

	ExcludeDroppedDescriptorsFromLookup bool

	ConfigureScratchRange bool

	ReconcilerInitialInterceptor func(startTS hlc.Timestamp)

	ProtectedTSReaderOverrideFn func(clock *hlc.Clock) ProtectedTSReader

	LimiterLimitOverride func() int64
}

func (t *TestingKnobs) ModuleTestingKnobs() { __antithesis_instrumentation__.Notify(242188) }

var _ base.ModuleTestingKnobs = (*TestingKnobs)(nil)
