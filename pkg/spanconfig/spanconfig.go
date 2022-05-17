package spanconfig

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/systemschema"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlliveness"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

type KVAccessor interface {
	GetSpanConfigRecords(ctx context.Context, targets []Target) ([]Record, error)

	GetAllSystemSpanConfigsThatApply(
		ctx context.Context, id roachpb.TenantID,
	) ([]roachpb.SpanConfig, error)

	UpdateSpanConfigRecords(
		ctx context.Context,
		toDelete []Target,
		toUpsert []Record,
		minCommitTS, maxCommitTS hlc.Timestamp,
	) error

	WithTxn(context.Context, *kv.Txn) KVAccessor
}

type KVSubscriber interface {
	StoreReader
	ProtectedTSReader

	LastUpdated() hlc.Timestamp
	Subscribe(func(ctx context.Context, updated roachpb.Span))
}

type SQLTranslator interface {
	Translate(ctx context.Context, ids descpb.IDs,
		generateSystemSpanConfigurations bool) ([]Record, hlc.Timestamp, error)
}

func FullTranslate(ctx context.Context, s SQLTranslator) ([]Record, hlc.Timestamp, error) {
	__antithesis_instrumentation__.Notify(240289)

	return s.Translate(ctx, descpb.IDs{keys.RootNamespaceID},
		true)
}

type SQLWatcherHandler func(context.Context, []SQLUpdate, hlc.Timestamp) error

type SQLWatcher interface {
	WatchForSQLUpdates(
		ctx context.Context,
		startTS hlc.Timestamp,
		handler SQLWatcherHandler,
	) error
}

type Reconciler interface {
	Reconcile(
		ctx context.Context,
		startTS hlc.Timestamp,
		session sqlliveness.Session,
		onCheckpoint func() error,
	) error

	Checkpoint() hlc.Timestamp
}

type Store interface {
	StoreWriter
	StoreReader

	ForEachOverlappingSpanConfig(
		context.Context, roachpb.Span, func(roachpb.Span, roachpb.SpanConfig) error,
	) error
}

type StoreWriter interface {
	Apply(ctx context.Context, dryrun bool, updates ...Update) (
		deleted []Target, added []Record,
	)
}

type StoreReader interface {
	NeedsSplit(ctx context.Context, start, end roachpb.RKey) bool
	ComputeSplitKey(ctx context.Context, start, end roachpb.RKey) roachpb.RKey
	GetSpanConfigForKey(ctx context.Context, key roachpb.RKey) (roachpb.SpanConfig, error)
}

type Limiter interface {
	ShouldLimit(ctx context.Context, txn *kv.Txn, delta int) (bool, error)
}

type Splitter interface {
	Splits(ctx context.Context, table catalog.TableDescriptor) (int, error)
}

func Delta(
	ctx context.Context, s Splitter, committed, uncommitted catalog.TableDescriptor,
) (int, error) {
	__antithesis_instrumentation__.Notify(240290)
	if committed == nil && func() bool {
		__antithesis_instrumentation__.Notify(240296)
		return uncommitted == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(240297)
		log.Fatalf(ctx, "unexpected: got two nil table descriptors")
	} else {
		__antithesis_instrumentation__.Notify(240298)
	}
	__antithesis_instrumentation__.Notify(240291)

	var nonNilDesc catalog.TableDescriptor
	if committed != nil {
		__antithesis_instrumentation__.Notify(240299)
		nonNilDesc = committed
	} else {
		__antithesis_instrumentation__.Notify(240300)
		nonNilDesc = uncommitted
	}
	__antithesis_instrumentation__.Notify(240292)
	if nonNilDesc.GetParentID() == systemschema.SystemDB.GetID() {
		__antithesis_instrumentation__.Notify(240301)
		return 0, nil
	} else {
		__antithesis_instrumentation__.Notify(240302)
	}
	__antithesis_instrumentation__.Notify(240293)

	uncommittedSplits, err := s.Splits(ctx, uncommitted)
	if err != nil {
		__antithesis_instrumentation__.Notify(240303)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(240304)
	}
	__antithesis_instrumentation__.Notify(240294)

	committedSplits, err := s.Splits(ctx, committed)
	if err != nil {
		__antithesis_instrumentation__.Notify(240305)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(240306)
	}
	__antithesis_instrumentation__.Notify(240295)

	delta := uncommittedSplits - committedSplits
	return delta, nil
}

type SQLUpdate struct {
	descriptorUpdate         DescriptorUpdate
	protectedTimestampUpdate ProtectedTimestampUpdate
}

func MakeDescriptorSQLUpdate(id descpb.ID, descType catalog.DescriptorType) SQLUpdate {
	__antithesis_instrumentation__.Notify(240307)
	return SQLUpdate{descriptorUpdate: DescriptorUpdate{
		ID:   id,
		Type: descType,
	}}
}

func (d *SQLUpdate) GetDescriptorUpdate() DescriptorUpdate {
	__antithesis_instrumentation__.Notify(240308)
	return d.descriptorUpdate
}

func (d *SQLUpdate) IsDescriptorUpdate() bool {
	__antithesis_instrumentation__.Notify(240309)
	return d.descriptorUpdate != DescriptorUpdate{}
}

func MakeTenantProtectedTimestampSQLUpdate(tenantID roachpb.TenantID) SQLUpdate {
	__antithesis_instrumentation__.Notify(240310)
	return SQLUpdate{protectedTimestampUpdate: ProtectedTimestampUpdate{TenantTarget: tenantID}}
}

func MakeClusterProtectedTimestampSQLUpdate() SQLUpdate {
	__antithesis_instrumentation__.Notify(240311)
	return SQLUpdate{protectedTimestampUpdate: ProtectedTimestampUpdate{ClusterTarget: true}}
}

func (d *SQLUpdate) GetProtectedTimestampUpdate() ProtectedTimestampUpdate {
	__antithesis_instrumentation__.Notify(240312)
	return d.protectedTimestampUpdate
}

func (d *SQLUpdate) IsProtectedTimestampUpdate() bool {
	__antithesis_instrumentation__.Notify(240313)
	return d.protectedTimestampUpdate != ProtectedTimestampUpdate{}
}

type DescriptorUpdate struct {
	ID descpb.ID

	Type catalog.DescriptorType
}

type ProtectedTimestampUpdate struct {
	ClusterTarget bool

	TenantTarget roachpb.TenantID
}

func (p *ProtectedTimestampUpdate) IsClusterUpdate() bool {
	__antithesis_instrumentation__.Notify(240314)
	return p.ClusterTarget
}

func (p *ProtectedTimestampUpdate) IsTenantsUpdate() bool {
	__antithesis_instrumentation__.Notify(240315)
	return !p.ClusterTarget
}

type Update Record

func Deletion(target Target) (Update, error) {
	__antithesis_instrumentation__.Notify(240316)
	record, err := MakeRecord(target, roachpb.SpanConfig{})
	if err != nil {
		__antithesis_instrumentation__.Notify(240318)
		return Update{}, err
	} else {
		__antithesis_instrumentation__.Notify(240319)
	}
	__antithesis_instrumentation__.Notify(240317)
	return Update(record), nil
}

func Addition(target Target, conf roachpb.SpanConfig) (Update, error) {
	__antithesis_instrumentation__.Notify(240320)
	record, err := MakeRecord(target, conf)
	if err != nil {
		__antithesis_instrumentation__.Notify(240322)
		return Update{}, err
	} else {
		__antithesis_instrumentation__.Notify(240323)
	}
	__antithesis_instrumentation__.Notify(240321)
	return Update(record), nil
}

func (u Update) Deletion() bool {
	__antithesis_instrumentation__.Notify(240324)
	return u.config.IsEmpty()
}

func (u Update) Addition() bool {
	__antithesis_instrumentation__.Notify(240325)
	return !u.Deletion()
}

func (u Update) GetTarget() Target {
	__antithesis_instrumentation__.Notify(240326)
	return u.target
}

func (u Update) GetConfig() roachpb.SpanConfig {
	__antithesis_instrumentation__.Notify(240327)
	return u.config
}

type ProtectedTSReader interface {
	GetProtectionTimestamps(ctx context.Context, sp roachpb.Span) (
		protectionTimestamps []hlc.Timestamp, asOf hlc.Timestamp, _ error,
	)
}

func EmptyProtectedTSReader(c *hlc.Clock) ProtectedTSReader {
	__antithesis_instrumentation__.Notify(240328)
	return (*emptyProtectedTSReader)(c)
}

type emptyProtectedTSReader hlc.Clock

func (r *emptyProtectedTSReader) GetProtectionTimestamps(
	context.Context, roachpb.Span,
) ([]hlc.Timestamp, hlc.Timestamp, error) {
	__antithesis_instrumentation__.Notify(240329)
	return nil, (*hlc.Clock)(r).Now(), nil
}
