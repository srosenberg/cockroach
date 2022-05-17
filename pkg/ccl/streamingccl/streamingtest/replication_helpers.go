package streamingtest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/changefeedbase"
	"github.com/cockroachdb/cockroach/pkg/ccl/streamingccl"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/testutils/serverutils"
	"github.com/cockroachdb/cockroach/pkg/testutils/sqlutils"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/require"
)

type FeedPredicate func(message streamingccl.Event) bool

func KeyMatches(key roachpb.Key) FeedPredicate {
	__antithesis_instrumentation__.Notify(25629)
	return func(msg streamingccl.Event) bool {
		__antithesis_instrumentation__.Notify(25630)
		if msg.Type() != streamingccl.KVEvent {
			__antithesis_instrumentation__.Notify(25632)
			return false
		} else {
			__antithesis_instrumentation__.Notify(25633)
		}
		__antithesis_instrumentation__.Notify(25631)
		return bytes.Equal(key, msg.GetKV().Key)
	}
}

func ResolvedAtLeast(lo hlc.Timestamp) FeedPredicate {
	__antithesis_instrumentation__.Notify(25634)
	return func(msg streamingccl.Event) bool {
		__antithesis_instrumentation__.Notify(25635)
		if msg.Type() != streamingccl.CheckpointEvent {
			__antithesis_instrumentation__.Notify(25637)
			return false
		} else {
			__antithesis_instrumentation__.Notify(25638)
		}
		__antithesis_instrumentation__.Notify(25636)
		return lo.LessEq(*msg.GetResolved())
	}
}

func ReceivedNewGeneration() FeedPredicate {
	__antithesis_instrumentation__.Notify(25639)
	return func(msg streamingccl.Event) bool {
		__antithesis_instrumentation__.Notify(25640)
		return msg.Type() == streamingccl.GenerationEvent
	}
}

type FeedSource interface {
	Next() (streamingccl.Event, bool)

	Close(ctx context.Context)
}

type ReplicationFeed struct {
	t   *testing.T
	f   FeedSource
	msg streamingccl.Event
}

func MakeReplicationFeed(t *testing.T, f FeedSource) *ReplicationFeed {
	__antithesis_instrumentation__.Notify(25641)
	return &ReplicationFeed{
		t: t,
		f: f,
	}
}

func (rf *ReplicationFeed) ObserveKey(ctx context.Context, key roachpb.Key) roachpb.KeyValue {
	__antithesis_instrumentation__.Notify(25642)
	require.NoError(rf.t, rf.consumeUntil(ctx, KeyMatches(key)))
	return *rf.msg.GetKV()
}

func (rf *ReplicationFeed) ObserveResolved(ctx context.Context, lo hlc.Timestamp) hlc.Timestamp {
	__antithesis_instrumentation__.Notify(25643)
	require.NoError(rf.t, rf.consumeUntil(ctx, ResolvedAtLeast(lo)))
	return *rf.msg.GetResolved()
}

func (rf *ReplicationFeed) ObserveGeneration(ctx context.Context) bool {
	__antithesis_instrumentation__.Notify(25644)
	require.NoError(rf.t, rf.consumeUntil(ctx, ReceivedNewGeneration()))
	return true
}

func (rf *ReplicationFeed) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(25645)
	rf.f.Close(ctx)
}

func (rf *ReplicationFeed) consumeUntil(ctx context.Context, pred FeedPredicate) error {
	__antithesis_instrumentation__.Notify(25646)
	const maxWait = 20 * time.Second
	doneCh := make(chan struct{})
	mu := struct {
		syncutil.Mutex
		err error
	}{}
	defer close(doneCh)
	go func() {
		__antithesis_instrumentation__.Notify(25648)
		select {
		case <-time.After(maxWait):
			__antithesis_instrumentation__.Notify(25649)
			mu.Lock()
			mu.err = errors.New("test timed out")
			mu.Unlock()
			rf.f.Close(ctx)
		case <-doneCh:
			__antithesis_instrumentation__.Notify(25650)
		}
	}()
	__antithesis_instrumentation__.Notify(25647)

	rowCount := 0
	for {
		__antithesis_instrumentation__.Notify(25651)
		msg, haveMoreRows := rf.f.Next()
		if !haveMoreRows {
			__antithesis_instrumentation__.Notify(25653)

			mu.Lock()
			err := mu.err
			mu.Unlock()
			if err != nil {
				__antithesis_instrumentation__.Notify(25654)
				rf.t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(25655)
				rf.t.Fatalf("ran out of rows after processing %d rows", rowCount)
			}
		} else {
			__antithesis_instrumentation__.Notify(25656)
		}
		__antithesis_instrumentation__.Notify(25652)
		rowCount++

		require.NotNil(rf.t, msg)
		if pred(msg) {
			__antithesis_instrumentation__.Notify(25657)
			rf.msg = msg
			return nil
		} else {
			__antithesis_instrumentation__.Notify(25658)
		}
	}
}

type TenantState struct {
	ID roachpb.TenantID

	Codec keys.SQLCodec

	SQL *sqlutils.SQLRunner
}

type ReplicationHelper struct {
	SysServer serverutils.TestServerInterface

	SysDB *sqlutils.SQLRunner

	PGUrl url.URL

	Tenant TenantState
}

func NewReplicationHelper(
	t *testing.T, serverArgs base.TestServerArgs,
) (*ReplicationHelper, func()) {
	__antithesis_instrumentation__.Notify(25659)
	ctx := context.Background()

	s, db, _ := serverutils.StartServer(t, serverArgs)

	resetFreq := changefeedbase.TestingSetDefaultMinCheckpointFrequency(50 * time.Millisecond)

	_, err := db.Exec(`
SET CLUSTER SETTING kv.rangefeed.enabled = true;
SET CLUSTER SETTING kv.closed_timestamp.target_duration = '1s';
SET CLUSTER SETTING changefeed.experimental_poll_interval = '10ms';
SET CLUSTER SETTING sql.defaults.experimental_stream_replication.enabled = 'on';
`)
	require.NoError(t, err)

	tenantID := serverutils.TestTenantID()
	_, tenantConn := serverutils.StartTenant(t, s, base.TestTenantArgs{TenantID: tenantID})

	sink, cleanupSink := sqlutils.PGUrl(t, s.ServingSQLAddr(), t.Name(), url.User(security.RootUser))

	h := &ReplicationHelper{
		SysServer: s,
		SysDB:     sqlutils.MakeSQLRunner(db),
		PGUrl:     sink,
		Tenant: TenantState{
			ID:    tenantID,
			Codec: keys.MakeSQLCodec(tenantID),
			SQL:   sqlutils.MakeSQLRunner(tenantConn),
		},
	}

	return h, func() {
		__antithesis_instrumentation__.Notify(25660)
		cleanupSink()
		resetFreq()
		require.NoError(t, tenantConn.Close())
		s.Stopper().Stop(ctx)
	}
}
