package scdeps

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/config"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scexec"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scexec/scmutationexec"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

type JobRegistry interface {
	MakeJobID() jobspb.JobID
	CreateJobWithTxn(ctx context.Context, record jobs.Record, jobID jobspb.JobID, txn *kv.Txn) (*jobs.Job, error)
	UpdateJobWithTxn(
		ctx context.Context, jobID jobspb.JobID, txn *kv.Txn, useReadLock bool, updateFunc jobs.UpdateFn,
	) error
	CheckPausepoint(name string) error
}

func NewExecutorDependencies(
	codec keys.SQLCodec,
	sessionData *sessiondata.SessionData,
	txn *kv.Txn,
	user security.SQLUsername,
	descsCollection *descs.Collection,
	jobRegistry JobRegistry,
	backfiller scexec.Backfiller,
	backfillTracker scexec.BackfillTracker,
	backfillFlusher scexec.PeriodicProgressFlusher,
	indexValidator scexec.IndexValidator,
	clock scmutationexec.Clock,
	commentUpdaterFactory scexec.DescriptorMetadataUpdaterFactory,
	eventLogger scexec.EventLogger,
	kvTrace bool,
	schemaChangerJobID jobspb.JobID,
	statements []string,
) scexec.Dependencies {
	__antithesis_instrumentation__.Notify(580624)
	return &execDeps{
		txnDeps: txnDeps{
			txn:                txn,
			codec:              codec,
			descsCollection:    descsCollection,
			jobRegistry:        jobRegistry,
			indexValidator:     indexValidator,
			eventLogger:        eventLogger,
			schemaChangerJobID: schemaChangerJobID,
			kvTrace:            kvTrace,
		},
		backfiller:              backfiller,
		backfillTracker:         backfillTracker,
		commentUpdaterFactory:   commentUpdaterFactory,
		periodicProgressFlusher: backfillFlusher,
		statements:              statements,
		user:                    user,
		sessionData:             sessionData,
		clock:                   clock,
	}
}

type txnDeps struct {
	txn                *kv.Txn
	codec              keys.SQLCodec
	descsCollection    *descs.Collection
	jobRegistry        JobRegistry
	createdJobs        []jobspb.JobID
	indexValidator     scexec.IndexValidator
	eventLogger        scexec.EventLogger
	deletedDescriptors catalog.DescriptorIDSet
	schemaChangerJobID jobspb.JobID
	kvTrace            bool
}

func (d *txnDeps) UpdateSchemaChangeJob(
	ctx context.Context, id jobspb.JobID, callback scexec.JobUpdateCallback,
) error {
	__antithesis_instrumentation__.Notify(580625)
	const useReadLock = false
	return d.jobRegistry.UpdateJobWithTxn(ctx, id, d.txn, useReadLock, func(
		txn *kv.Txn, md jobs.JobMetadata, ju *jobs.JobUpdater,
	) error {
		__antithesis_instrumentation__.Notify(580626)
		setNonCancelable := func() {
			__antithesis_instrumentation__.Notify(580628)
			payload := *md.Payload
			if !payload.Noncancelable {
				__antithesis_instrumentation__.Notify(580629)
				payload.Noncancelable = true
				ju.UpdatePayload(&payload)
			} else {
				__antithesis_instrumentation__.Notify(580630)
			}
		}
		__antithesis_instrumentation__.Notify(580627)
		return callback(md, ju.UpdateProgress, setNonCancelable)
	})
}

var _ scexec.Catalog = (*txnDeps)(nil)

func (d *txnDeps) MustReadImmutableDescriptors(
	ctx context.Context, ids ...descpb.ID,
) ([]catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(580631)
	flags := tree.CommonLookupFlags{
		Required:       true,
		RequireMutable: false,
		AvoidLeased:    true,
		IncludeOffline: true,
		IncludeDropped: true,
	}
	return d.descsCollection.GetImmutableDescriptorsByID(ctx, d.txn, flags, ids...)
}

func (d *txnDeps) GetFullyQualifiedName(ctx context.Context, id descpb.ID) (string, error) {
	__antithesis_instrumentation__.Notify(580632)
	objectDesc, err := d.descsCollection.GetImmutableDescriptorByID(ctx,
		d.txn,
		id,
		tree.CommonLookupFlags{
			Required:       true,
			IncludeDropped: true,
			AvoidLeased:    true,
		})
	if err != nil {
		__antithesis_instrumentation__.Notify(580635)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(580636)
	}
	__antithesis_instrumentation__.Notify(580633)

	if objectDesc.DescriptorType() != catalog.Database && func() bool {
		__antithesis_instrumentation__.Notify(580637)
		return objectDesc.DescriptorType() != catalog.Schema == true
	}() == true {
		__antithesis_instrumentation__.Notify(580638)
		_, databaseDesc, err := d.descsCollection.GetImmutableDatabaseByID(ctx,
			d.txn,
			objectDesc.GetParentID(),
			tree.CommonLookupFlags{
				IncludeDropped: true,
				Required:       true,
				AvoidLeased:    true,
			})
		if err != nil {
			__antithesis_instrumentation__.Notify(580641)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(580642)
		}
		__antithesis_instrumentation__.Notify(580639)
		schemaDesc, err := d.descsCollection.GetImmutableSchemaByID(ctx, d.txn, objectDesc.GetParentSchemaID(),
			tree.SchemaLookupFlags{
				Required:       true,
				IncludeDropped: true,
				AvoidLeased:    true,
			})
		if err != nil {
			__antithesis_instrumentation__.Notify(580643)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(580644)
		}
		__antithesis_instrumentation__.Notify(580640)
		name := tree.MakeTableNameWithSchema(tree.Name(databaseDesc.GetName()),
			tree.Name(schemaDesc.GetName()),
			tree.Name(objectDesc.GetName()))
		return name.FQString(), nil
	} else {
		__antithesis_instrumentation__.Notify(580645)
		if objectDesc.DescriptorType() == catalog.Database {
			__antithesis_instrumentation__.Notify(580646)
			return objectDesc.GetName(), nil
		} else {
			__antithesis_instrumentation__.Notify(580647)
			if objectDesc.DescriptorType() == catalog.Schema {
				__antithesis_instrumentation__.Notify(580648)
				_, databaseDesc, err := d.descsCollection.GetImmutableDatabaseByID(ctx,
					d.txn,
					objectDesc.GetParentID(),
					tree.CommonLookupFlags{
						IncludeDropped: true,
						Required:       true,
						AvoidLeased:    true,
					})
				if err != nil {
					__antithesis_instrumentation__.Notify(580650)
					return "", err
				} else {
					__antithesis_instrumentation__.Notify(580651)
				}
				__antithesis_instrumentation__.Notify(580649)
				return fmt.Sprintf("%s.%s", databaseDesc.GetName(), objectDesc.GetName()), nil
			} else {
				__antithesis_instrumentation__.Notify(580652)
			}
		}
	}
	__antithesis_instrumentation__.Notify(580634)
	return "", errors.Newf("unknown descriptor type : %s\n", objectDesc.DescriptorType())
}

func (d *txnDeps) AddSyntheticDescriptor(desc catalog.Descriptor) {
	__antithesis_instrumentation__.Notify(580653)
	d.descsCollection.AddSyntheticDescriptor(desc)
}

func (d *txnDeps) RemoveSyntheticDescriptor(id descpb.ID) {
	__antithesis_instrumentation__.Notify(580654)
	d.descsCollection.RemoveSyntheticDescriptor(id)
}

func (d *txnDeps) MustReadMutableDescriptor(
	ctx context.Context, id descpb.ID,
) (catalog.MutableDescriptor, error) {
	__antithesis_instrumentation__.Notify(580655)
	return d.descsCollection.GetMutableDescriptorByID(ctx, d.txn, id)
}

func (d *txnDeps) NewCatalogChangeBatcher() scexec.CatalogChangeBatcher {
	__antithesis_instrumentation__.Notify(580656)
	return &catalogChangeBatcher{
		txnDeps: d,
		batch:   d.txn.NewBatch(),
	}
}

type catalogChangeBatcher struct {
	*txnDeps
	batch *kv.Batch
}

var _ scexec.CatalogChangeBatcher = (*catalogChangeBatcher)(nil)

func (b *catalogChangeBatcher) CreateOrUpdateDescriptor(
	ctx context.Context, desc catalog.MutableDescriptor,
) error {
	__antithesis_instrumentation__.Notify(580657)
	return b.descsCollection.WriteDescToBatch(ctx, b.kvTrace, desc, b.batch)
}

func (b *catalogChangeBatcher) DeleteName(
	ctx context.Context, nameInfo descpb.NameInfo, id descpb.ID,
) error {
	__antithesis_instrumentation__.Notify(580658)
	marshalledKey := catalogkeys.EncodeNameKey(b.codec, nameInfo)
	if b.kvTrace {
		__antithesis_instrumentation__.Notify(580660)
		log.VEventf(ctx, 2, "Del %s", marshalledKey)
	} else {
		__antithesis_instrumentation__.Notify(580661)
	}
	__antithesis_instrumentation__.Notify(580659)
	b.batch.Del(marshalledKey)
	return nil
}

func (b *catalogChangeBatcher) DeleteDescriptor(ctx context.Context, id descpb.ID) error {
	__antithesis_instrumentation__.Notify(580662)
	marshalledKey := catalogkeys.MakeDescMetadataKey(b.codec, id)
	b.batch.Del(marshalledKey)
	if b.kvTrace {
		__antithesis_instrumentation__.Notify(580664)
		log.VEventf(ctx, 2, "Del %s", marshalledKey)
	} else {
		__antithesis_instrumentation__.Notify(580665)
	}
	__antithesis_instrumentation__.Notify(580663)
	b.deletedDescriptors.Add(id)
	b.descsCollection.AddDeletedDescriptor(id)
	return nil
}

func (b *catalogChangeBatcher) DeleteZoneConfig(ctx context.Context, id descpb.ID) error {
	__antithesis_instrumentation__.Notify(580666)
	zoneKeyPrefix := config.MakeZoneKeyPrefix(b.codec, id)
	if b.kvTrace {
		__antithesis_instrumentation__.Notify(580668)
		log.VEventf(ctx, 2, "DelRange %s", zoneKeyPrefix)
	} else {
		__antithesis_instrumentation__.Notify(580669)
	}
	__antithesis_instrumentation__.Notify(580667)
	b.batch.DelRange(zoneKeyPrefix, zoneKeyPrefix.PrefixEnd(), false)
	return nil
}

func (b *catalogChangeBatcher) ValidateAndRun(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(580670)
	if err := b.descsCollection.ValidateUncommittedDescriptors(ctx, b.txn); err != nil {
		__antithesis_instrumentation__.Notify(580673)
		return err
	} else {
		__antithesis_instrumentation__.Notify(580674)
	}
	__antithesis_instrumentation__.Notify(580671)
	if err := b.txn.Run(ctx, b.batch); err != nil {
		__antithesis_instrumentation__.Notify(580675)
		return errors.Wrap(err, "writing descriptors")
	} else {
		__antithesis_instrumentation__.Notify(580676)
	}
	__antithesis_instrumentation__.Notify(580672)
	return nil
}

var _ scexec.TransactionalJobRegistry = (*txnDeps)(nil)

func (d *txnDeps) MakeJobID() jobspb.JobID {
	__antithesis_instrumentation__.Notify(580677)
	return d.jobRegistry.MakeJobID()
}

func (d *txnDeps) CheckPausepoint(name string) error {
	__antithesis_instrumentation__.Notify(580678)
	return d.jobRegistry.CheckPausepoint(name)
}

func (d *txnDeps) SchemaChangerJobID() jobspb.JobID {
	__antithesis_instrumentation__.Notify(580679)
	if d.schemaChangerJobID == 0 {
		__antithesis_instrumentation__.Notify(580681)
		d.schemaChangerJobID = d.jobRegistry.MakeJobID()
	} else {
		__antithesis_instrumentation__.Notify(580682)
	}
	__antithesis_instrumentation__.Notify(580680)
	return d.schemaChangerJobID
}

func (d *txnDeps) CreateJob(ctx context.Context, record jobs.Record) error {
	__antithesis_instrumentation__.Notify(580683)
	if _, err := d.jobRegistry.CreateJobWithTxn(ctx, record, record.JobID, d.txn); err != nil {
		__antithesis_instrumentation__.Notify(580685)
		return err
	} else {
		__antithesis_instrumentation__.Notify(580686)
	}
	__antithesis_instrumentation__.Notify(580684)
	d.createdJobs = append(d.createdJobs, record.JobID)
	return nil
}

func (d *txnDeps) CreatedJobs() []jobspb.JobID {
	__antithesis_instrumentation__.Notify(580687)
	return d.createdJobs
}

var _ scexec.IndexSpanSplitter = (*txnDeps)(nil)

func (d *txnDeps) MaybeSplitIndexSpans(
	ctx context.Context, table catalog.TableDescriptor, indexToBackfill catalog.Index,
) error {
	__antithesis_instrumentation__.Notify(580688)

	if !d.codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(580690)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(580691)
	}
	__antithesis_instrumentation__.Notify(580689)

	span := table.IndexSpan(d.codec, indexToBackfill.GetID())
	const backfillSplitExpiration = time.Hour
	expirationTime := d.txn.DB().Clock().Now().Add(backfillSplitExpiration.Nanoseconds(), 0)
	return d.txn.DB().AdminSplit(ctx, span.Key, expirationTime)
}

func (d *txnDeps) GetResumeSpans(
	ctx context.Context, tableID descpb.ID, indexID descpb.IndexID,
) ([]roachpb.Span, error) {
	__antithesis_instrumentation__.Notify(580692)
	table, err := d.descsCollection.GetImmutableTableByID(ctx, d.txn, tableID, tree.ObjectLookupFlags{
		CommonLookupFlags: tree.CommonLookupFlags{
			Required:    true,
			AvoidLeased: true,
		},
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(580694)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(580695)
	}
	__antithesis_instrumentation__.Notify(580693)
	return []roachpb.Span{table.IndexSpan(d.codec, indexID)}, nil
}

func (d *txnDeps) SetResumeSpans(
	ctx context.Context, tableID descpb.ID, indexID descpb.IndexID, total, done []roachpb.Span,
) error {
	__antithesis_instrumentation__.Notify(580696)
	panic("implement me")
}

type execDeps struct {
	txnDeps
	clock                   scmutationexec.Clock
	commentUpdaterFactory   scexec.DescriptorMetadataUpdaterFactory
	backfiller              scexec.Backfiller
	backfillTracker         scexec.BackfillTracker
	periodicProgressFlusher scexec.PeriodicProgressFlusher
	statements              []string
	user                    security.SQLUsername
	sessionData             *sessiondata.SessionData
}

func (d *execDeps) Clock() scmutationexec.Clock {
	__antithesis_instrumentation__.Notify(580697)
	return d.clock
}

var _ scexec.Dependencies = (*execDeps)(nil)

func (d *execDeps) Catalog() scexec.Catalog {
	__antithesis_instrumentation__.Notify(580698)
	return d
}

func (d *execDeps) IndexBackfiller() scexec.Backfiller {
	__antithesis_instrumentation__.Notify(580699)
	return d.backfiller
}

func (d *execDeps) BackfillProgressTracker() scexec.BackfillTracker {
	__antithesis_instrumentation__.Notify(580700)
	return d.backfillTracker
}

func (d *execDeps) PeriodicProgressFlusher() scexec.PeriodicProgressFlusher {
	__antithesis_instrumentation__.Notify(580701)
	return d.periodicProgressFlusher
}

func (d *execDeps) IndexValidator() scexec.IndexValidator {
	__antithesis_instrumentation__.Notify(580702)
	return d.indexValidator
}

func (d *execDeps) IndexSpanSplitter() scexec.IndexSpanSplitter {
	__antithesis_instrumentation__.Notify(580703)
	return d
}

func (d *execDeps) TransactionalJobRegistry() scexec.TransactionalJobRegistry {
	__antithesis_instrumentation__.Notify(580704)
	return d
}

func (d *execDeps) Statements() []string {
	__antithesis_instrumentation__.Notify(580705)
	return d.statements
}

func (d *execDeps) User() security.SQLUsername {
	__antithesis_instrumentation__.Notify(580706)
	return d.user
}

func (d *execDeps) DescriptorMetadataUpdater(ctx context.Context) scexec.DescriptorMetadataUpdater {
	__antithesis_instrumentation__.Notify(580707)
	return d.commentUpdaterFactory.NewMetadataUpdater(ctx, d.txn, d.sessionData)
}

type EventLoggerFactory = func(*kv.Txn) scexec.EventLogger

func (d *execDeps) EventLogger() scexec.EventLogger {
	__antithesis_instrumentation__.Notify(580708)
	return d.eventLogger
}

func NewNoOpBackfillTracker(codec keys.SQLCodec) scexec.BackfillTracker {
	__antithesis_instrumentation__.Notify(580709)
	return noopBackfillProgress{codec: codec}
}

type noopBackfillProgress struct {
	codec keys.SQLCodec
}

func (n noopBackfillProgress) FlushCheckpoint(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(580710)
	return nil
}

func (n noopBackfillProgress) FlushFractionCompleted(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(580711)
	return nil
}

func (n noopBackfillProgress) GetBackfillProgress(
	ctx context.Context, b scexec.Backfill,
) (scexec.BackfillProgress, error) {
	__antithesis_instrumentation__.Notify(580712)
	key := n.codec.IndexPrefix(uint32(b.TableID), uint32(b.SourceIndexID))
	return scexec.BackfillProgress{
		Backfill: b,
		CompletedSpans: []roachpb.Span{
			{Key: key, EndKey: key.PrefixEnd()},
		},
	}, nil
}

func (n noopBackfillProgress) SetBackfillProgress(
	ctx context.Context, progress scexec.BackfillProgress,
) error {
	__antithesis_instrumentation__.Notify(580713)
	return nil
}

type noopPeriodicProgressFlusher struct {
}

func NewNoopPeriodicProgressFlusher() scexec.PeriodicProgressFlusher {
	__antithesis_instrumentation__.Notify(580714)
	return noopPeriodicProgressFlusher{}
}

func (n noopPeriodicProgressFlusher) StartPeriodicUpdates(
	ctx context.Context, tracker scexec.BackfillProgressFlusher,
) (stop func() error) {
	__antithesis_instrumentation__.Notify(580715)
	return func() error { __antithesis_instrumentation__.Notify(580716); return nil }
}

type constantClock struct {
	ts time.Time
}

func NewConstantClock(ts time.Time) scmutationexec.Clock {
	__antithesis_instrumentation__.Notify(580717)
	return constantClock{ts: ts}
}

func (c constantClock) ApproximateTime() time.Time {
	__antithesis_instrumentation__.Notify(580718)
	return c.ts
}
