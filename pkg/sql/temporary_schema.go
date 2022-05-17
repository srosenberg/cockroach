package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descidgen"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/resolver"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/uint128"
	"github.com/cockroachdb/errors"
	io_prometheus_client "github.com/prometheus/client_model/go"
)

var TempObjectCleanupInterval = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"sql.temp_object_cleaner.cleanup_interval",
	"how often to clean up orphaned temporary objects",
	30*time.Minute,
).WithPublic()

var TempObjectWaitInterval = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"sql.temp_object_cleaner.wait_interval",
	"how long after creation a temporary object will be cleaned up",
	30*time.Minute,
).WithPublic()

var (
	temporaryObjectCleanerActiveCleanersMetric = metric.Metadata{
		Name:        "sql.temp_object_cleaner.active_cleaners",
		Help:        "number of cleaner tasks currently running on this node",
		Measurement: "Count",
		Unit:        metric.Unit_COUNT,
		MetricType:  io_prometheus_client.MetricType_GAUGE,
	}
	temporaryObjectCleanerSchemasToDeleteMetric = metric.Metadata{
		Name:        "sql.temp_object_cleaner.schemas_to_delete",
		Help:        "number of schemas to be deleted by the temp object cleaner on this node",
		Measurement: "Count",
		Unit:        metric.Unit_COUNT,
		MetricType:  io_prometheus_client.MetricType_COUNTER,
	}
	temporaryObjectCleanerSchemasDeletionErrorMetric = metric.Metadata{
		Name:        "sql.temp_object_cleaner.schemas_deletion_error",
		Help:        "number of errored schema deletions by the temp object cleaner on this node",
		Measurement: "Count",
		Unit:        metric.Unit_COUNT,
		MetricType:  io_prometheus_client.MetricType_COUNTER,
	}
	temporaryObjectCleanerSchemasDeletionSuccessMetric = metric.Metadata{
		Name:        "sql.temp_object_cleaner.schemas_deletion_success",
		Help:        "number of successful schema deletions by the temp object cleaner on this node",
		Measurement: "Count",
		Unit:        metric.Unit_COUNT,
		MetricType:  io_prometheus_client.MetricType_COUNTER,
	}
)

const TemporarySchemaNameForRestorePrefix string = "pg_temp_0_"

func (p *planner) getOrCreateTemporarySchema(
	ctx context.Context, db catalog.DatabaseDescriptor,
) (catalog.SchemaDescriptor, error) {
	__antithesis_instrumentation__.Notify(627843)
	tempSchemaName := p.TemporarySchemaName()
	sc, err := p.Descriptors().GetMutableSchemaByName(ctx, p.txn, db, tempSchemaName, p.CommonLookupFlags(false))
	if sc != nil || func() bool {
		__antithesis_instrumentation__.Notify(627848)
		return err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(627849)
		return sc, err
	} else {
		__antithesis_instrumentation__.Notify(627850)
	}
	__antithesis_instrumentation__.Notify(627844)
	sKey := catalogkeys.NewNameKeyComponents(db.GetID(), keys.RootNamespaceID, tempSchemaName)

	id, err := descidgen.GenerateUniqueDescID(ctx, p.ExecCfg().DB, p.ExecCfg().Codec)
	if err != nil {
		__antithesis_instrumentation__.Notify(627851)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(627852)
	}
	__antithesis_instrumentation__.Notify(627845)
	if err := p.CreateSchemaNamespaceEntry(ctx, catalogkeys.EncodeNameKey(p.ExecCfg().Codec, sKey), id); err != nil {
		__antithesis_instrumentation__.Notify(627853)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(627854)
	}
	__antithesis_instrumentation__.Notify(627846)
	p.sessionDataMutatorIterator.applyOnEachMutator(func(m sessionDataMutator) {
		__antithesis_instrumentation__.Notify(627855)
		m.SetTemporarySchemaName(sKey.GetName())
		m.SetTemporarySchemaIDForDatabase(uint32(db.GetID()), uint32(id))
	})
	__antithesis_instrumentation__.Notify(627847)
	return p.Descriptors().GetImmutableSchemaByID(ctx, p.Txn(), id, p.CommonLookupFlags(true))
}

func (p *planner) CreateSchemaNamespaceEntry(
	ctx context.Context, schemaNameKey roachpb.Key, schemaID descpb.ID,
) error {
	__antithesis_instrumentation__.Notify(627856)
	if p.ExtendedEvalContext().Tracing.KVTracingEnabled() {
		__antithesis_instrumentation__.Notify(627858)
		log.VEventf(ctx, 2, "CPut %s -> %d", schemaNameKey, schemaID)
	} else {
		__antithesis_instrumentation__.Notify(627859)
	}
	__antithesis_instrumentation__.Notify(627857)

	b := &kv.Batch{}
	b.CPut(schemaNameKey, schemaID, nil)

	return p.txn.Run(ctx, b)
}

func temporarySchemaName(sessionID ClusterWideID) string {
	__antithesis_instrumentation__.Notify(627860)
	return fmt.Sprintf("pg_temp_%d_%d", sessionID.Hi, sessionID.Lo)
}

func temporarySchemaSessionID(scName string) (bool, ClusterWideID, error) {
	__antithesis_instrumentation__.Notify(627861)
	if !strings.HasPrefix(scName, "pg_temp_") {
		__antithesis_instrumentation__.Notify(627866)
		return false, ClusterWideID{}, nil
	} else {
		__antithesis_instrumentation__.Notify(627867)
	}
	__antithesis_instrumentation__.Notify(627862)
	parts := strings.Split(scName, "_")
	if len(parts) != 4 {
		__antithesis_instrumentation__.Notify(627868)
		return false, ClusterWideID{}, errors.Errorf("malformed temp schema name %s", scName)
	} else {
		__antithesis_instrumentation__.Notify(627869)
	}
	__antithesis_instrumentation__.Notify(627863)
	hi, err := strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		__antithesis_instrumentation__.Notify(627870)
		return false, ClusterWideID{}, err
	} else {
		__antithesis_instrumentation__.Notify(627871)
	}
	__antithesis_instrumentation__.Notify(627864)
	lo, err := strconv.ParseUint(parts[3], 10, 64)
	if err != nil {
		__antithesis_instrumentation__.Notify(627872)
		return false, ClusterWideID{}, err
	} else {
		__antithesis_instrumentation__.Notify(627873)
	}
	__antithesis_instrumentation__.Notify(627865)
	return true, ClusterWideID{uint128.Uint128{Hi: hi, Lo: lo}}, nil
}

func cleanupSessionTempObjects(
	ctx context.Context,
	settings *cluster.Settings,
	cf *descs.CollectionFactory,
	db *kv.DB,
	codec keys.SQLCodec,
	ie sqlutil.InternalExecutor,
	sessionID ClusterWideID,
) error {
	__antithesis_instrumentation__.Notify(627874)
	tempSchemaName := temporarySchemaName(sessionID)
	return cf.Txn(ctx, ie, db, func(ctx context.Context, txn *kv.Txn, descsCol *descs.Collection) error {
		__antithesis_instrumentation__.Notify(627875)

		allDbDescs, err := descsCol.GetAllDatabaseDescriptors(ctx, txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(627878)
			return err
		} else {
			__antithesis_instrumentation__.Notify(627879)
		}
		__antithesis_instrumentation__.Notify(627876)
		for _, dbDesc := range allDbDescs {
			__antithesis_instrumentation__.Notify(627880)
			if err := cleanupSchemaObjects(
				ctx,
				settings,
				txn,
				descsCol,
				codec,
				ie,
				dbDesc,
				tempSchemaName,
			); err != nil {
				__antithesis_instrumentation__.Notify(627882)
				return err
			} else {
				__antithesis_instrumentation__.Notify(627883)
			}
			__antithesis_instrumentation__.Notify(627881)

			key := catalogkeys.MakeSchemaNameKey(codec, dbDesc.GetID(), tempSchemaName)
			if err := txn.Del(ctx, key); err != nil {
				__antithesis_instrumentation__.Notify(627884)
				return err
			} else {
				__antithesis_instrumentation__.Notify(627885)
			}
		}
		__antithesis_instrumentation__.Notify(627877)
		return nil
	})
}

func cleanupSchemaObjects(
	ctx context.Context,
	settings *cluster.Settings,
	txn *kv.Txn,
	descsCol *descs.Collection,
	codec keys.SQLCodec,
	ie sqlutil.InternalExecutor,
	dbDesc catalog.DatabaseDescriptor,
	schemaName string,
) error {
	__antithesis_instrumentation__.Notify(627886)
	tbNames, tbIDs, err := descsCol.GetObjectNamesAndIDs(
		ctx,
		txn,
		dbDesc,
		schemaName,
		tree.DatabaseListFlags{CommonLookupFlags: tree.CommonLookupFlags{Required: false}},
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(627890)
		return err
	} else {
		__antithesis_instrumentation__.Notify(627891)
	}
	__antithesis_instrumentation__.Notify(627887)

	databaseIDToTempSchemaID := make(map[uint32]uint32)

	var tables descpb.IDs
	var views descpb.IDs
	var sequences descpb.IDs

	tblDescsByID := make(map[descpb.ID]catalog.TableDescriptor, len(tbNames))
	tblNamesByID := make(map[descpb.ID]tree.TableName, len(tbNames))
	for i, tbName := range tbNames {
		__antithesis_instrumentation__.Notify(627892)
		desc, err := descsCol.Direct().MustGetTableDescByID(ctx, txn, tbIDs[i])
		if err != nil {
			__antithesis_instrumentation__.Notify(627894)
			return err
		} else {
			__antithesis_instrumentation__.Notify(627895)
		}
		__antithesis_instrumentation__.Notify(627893)

		tblDescsByID[desc.GetID()] = desc
		tblNamesByID[desc.GetID()] = tbName

		databaseIDToTempSchemaID[uint32(desc.GetParentID())] = uint32(desc.GetParentSchemaID())

		if desc.IsSequence() && func() bool {
			__antithesis_instrumentation__.Notify(627896)
			return desc.GetSequenceOpts().SequenceOwner.OwnerColumnID == 0 == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(627897)
			return desc.GetSequenceOpts().SequenceOwner.OwnerTableID == 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(627898)
			sequences = append(sequences, desc.GetID())
		} else {
			__antithesis_instrumentation__.Notify(627899)
			if desc.GetViewQuery() != "" {
				__antithesis_instrumentation__.Notify(627900)
				views = append(views, desc.GetID())
			} else {
				__antithesis_instrumentation__.Notify(627901)
				if !desc.IsSequence() {
					__antithesis_instrumentation__.Notify(627902)
					tables = append(tables, desc.GetID())
				} else {
					__antithesis_instrumentation__.Notify(627903)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(627888)

	searchPath := sessiondata.DefaultSearchPathForUser(security.RootUserName()).WithTemporarySchemaName(schemaName)
	override := sessiondata.InternalExecutorOverride{
		SearchPath:               &searchPath,
		User:                     security.RootUserName(),
		DatabaseIDToTempSchemaID: databaseIDToTempSchemaID,
	}

	for _, toDelete := range []struct {
		typeName string

		ids descpb.IDs

		preHook func(descpb.ID) error
	}{

		{"VIEW", views, nil},
		{"TABLE", tables, nil},

		{
			"SEQUENCE",
			sequences,
			func(id descpb.ID) error {
				__antithesis_instrumentation__.Notify(627904)
				desc := tblDescsByID[id]

				return desc.ForeachDependedOnBy(func(d *descpb.TableDescriptor_Reference) error {
					__antithesis_instrumentation__.Notify(627905)

					if _, ok := tblDescsByID[d.ID]; ok {
						__antithesis_instrumentation__.Notify(627912)
						return nil
					} else {
						__antithesis_instrumentation__.Notify(627913)
					}
					__antithesis_instrumentation__.Notify(627906)
					dTableDesc, err := descsCol.Direct().MustGetTableDescByID(ctx, txn, d.ID)
					if err != nil {
						__antithesis_instrumentation__.Notify(627914)
						return err
					} else {
						__antithesis_instrumentation__.Notify(627915)
					}
					__antithesis_instrumentation__.Notify(627907)
					db, err := descsCol.Direct().MustGetDatabaseDescByID(ctx, txn, dTableDesc.GetParentID())
					if err != nil {
						__antithesis_instrumentation__.Notify(627916)
						return err
					} else {
						__antithesis_instrumentation__.Notify(627917)
					}
					__antithesis_instrumentation__.Notify(627908)
					schema, err := resolver.ResolveSchemaNameByID(
						ctx,
						txn,
						codec,
						db,
						dTableDesc.GetParentSchemaID(),
					)
					if err != nil {
						__antithesis_instrumentation__.Notify(627918)
						return err
					} else {
						__antithesis_instrumentation__.Notify(627919)
					}
					__antithesis_instrumentation__.Notify(627909)
					dependentColIDs := util.MakeFastIntSet()
					for _, colID := range d.ColumnIDs {
						__antithesis_instrumentation__.Notify(627920)
						dependentColIDs.Add(int(colID))
					}
					__antithesis_instrumentation__.Notify(627910)
					for _, col := range dTableDesc.PublicColumns() {
						__antithesis_instrumentation__.Notify(627921)
						if dependentColIDs.Contains(int(col.GetID())) {
							__antithesis_instrumentation__.Notify(627922)
							tbName := tree.MakeTableNameWithSchema(
								tree.Name(db.GetName()),
								tree.Name(schema),
								tree.Name(dTableDesc.GetName()),
							)
							_, err = ie.ExecEx(
								ctx,
								"delete-temp-dependent-col",
								txn,
								override,
								fmt.Sprintf(
									"ALTER TABLE %s ALTER COLUMN %s DROP DEFAULT",
									tbName.FQString(),
									tree.NameString(col.GetName()),
								),
							)
							if err != nil {
								__antithesis_instrumentation__.Notify(627923)
								return err
							} else {
								__antithesis_instrumentation__.Notify(627924)
							}
						} else {
							__antithesis_instrumentation__.Notify(627925)
						}
					}
					__antithesis_instrumentation__.Notify(627911)
					return nil
				})
			},
		},
	} {
		__antithesis_instrumentation__.Notify(627926)
		if len(toDelete.ids) > 0 {
			__antithesis_instrumentation__.Notify(627927)
			if toDelete.preHook != nil {
				__antithesis_instrumentation__.Notify(627930)
				for _, id := range toDelete.ids {
					__antithesis_instrumentation__.Notify(627931)
					if err := toDelete.preHook(id); err != nil {
						__antithesis_instrumentation__.Notify(627932)
						return err
					} else {
						__antithesis_instrumentation__.Notify(627933)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(627934)
			}
			__antithesis_instrumentation__.Notify(627928)

			var query strings.Builder
			query.WriteString("DROP ")
			query.WriteString(toDelete.typeName)

			for i, id := range toDelete.ids {
				__antithesis_instrumentation__.Notify(627935)
				tbName := tblNamesByID[id]
				if i != 0 {
					__antithesis_instrumentation__.Notify(627937)
					query.WriteString(",")
				} else {
					__antithesis_instrumentation__.Notify(627938)
				}
				__antithesis_instrumentation__.Notify(627936)
				query.WriteString(" ")
				query.WriteString(tbName.FQString())
			}
			__antithesis_instrumentation__.Notify(627929)
			query.WriteString(" CASCADE")
			_, err = ie.ExecEx(ctx, "delete-temp-"+toDelete.typeName, txn, override, query.String())
			if err != nil {
				__antithesis_instrumentation__.Notify(627939)
				return err
			} else {
				__antithesis_instrumentation__.Notify(627940)
			}
		} else {
			__antithesis_instrumentation__.Notify(627941)
		}
	}
	__antithesis_instrumentation__.Notify(627889)
	return nil
}

type isMeta1LeaseholderFunc func(context.Context, hlc.ClockTimestamp) (bool, error)

type TemporaryObjectCleaner struct {
	settings                         *cluster.Settings
	db                               *kv.DB
	codec                            keys.SQLCodec
	makeSessionBoundInternalExecutor sqlutil.SessionBoundInternalExecutorFactory

	statusServer           serverpb.SQLStatusServer
	isMeta1LeaseholderFunc isMeta1LeaseholderFunc
	testingKnobs           ExecutorTestingKnobs
	metrics                *temporaryObjectCleanerMetrics
	collectionFactory      *descs.CollectionFactory
}

type temporaryObjectCleanerMetrics struct {
	ActiveCleaners         *metric.Gauge
	SchemasToDelete        *metric.Counter
	SchemasDeletionError   *metric.Counter
	SchemasDeletionSuccess *metric.Counter
}

var _ metric.Struct = (*temporaryObjectCleanerMetrics)(nil)

func (m *temporaryObjectCleanerMetrics) MetricStruct() { __antithesis_instrumentation__.Notify(627942) }

func NewTemporaryObjectCleaner(
	settings *cluster.Settings,
	db *kv.DB,
	codec keys.SQLCodec,
	registry *metric.Registry,
	makeSessionBoundInternalExecutor sqlutil.SessionBoundInternalExecutorFactory,
	statusServer serverpb.SQLStatusServer,
	isMeta1LeaseholderFunc isMeta1LeaseholderFunc,
	testingKnobs ExecutorTestingKnobs,
	cf *descs.CollectionFactory,
) *TemporaryObjectCleaner {
	__antithesis_instrumentation__.Notify(627943)
	metrics := makeTemporaryObjectCleanerMetrics()
	registry.AddMetricStruct(metrics)
	return &TemporaryObjectCleaner{
		settings:                         settings,
		db:                               db,
		codec:                            codec,
		makeSessionBoundInternalExecutor: makeSessionBoundInternalExecutor,
		statusServer:                     statusServer,
		isMeta1LeaseholderFunc:           isMeta1LeaseholderFunc,
		testingKnobs:                     testingKnobs,
		metrics:                          metrics,
		collectionFactory:                cf,
	}
}

func makeTemporaryObjectCleanerMetrics() *temporaryObjectCleanerMetrics {
	__antithesis_instrumentation__.Notify(627944)
	return &temporaryObjectCleanerMetrics{
		ActiveCleaners:         metric.NewGauge(temporaryObjectCleanerActiveCleanersMetric),
		SchemasToDelete:        metric.NewCounter(temporaryObjectCleanerSchemasToDeleteMetric),
		SchemasDeletionError:   metric.NewCounter(temporaryObjectCleanerSchemasDeletionErrorMetric),
		SchemasDeletionSuccess: metric.NewCounter(temporaryObjectCleanerSchemasDeletionSuccessMetric),
	}
}

func (c *TemporaryObjectCleaner) doTemporaryObjectCleanup(
	ctx context.Context, closerCh <-chan struct{},
) error {
	__antithesis_instrumentation__.Notify(627945)
	defer log.Infof(ctx, "completed temporary object cleanup job")

	retryFunc := func(ctx context.Context, do func() error) error {
		__antithesis_instrumentation__.Notify(627954)
		return retry.WithMaxAttempts(
			ctx,
			retry.Options{
				InitialBackoff: 1 * time.Second,
				MaxBackoff:     1 * time.Minute,
				Multiplier:     2,
				Closer:         closerCh,
			},
			5,
			func() error {
				__antithesis_instrumentation__.Notify(627955)
				err := do()
				if err != nil {
					__antithesis_instrumentation__.Notify(627957)
					log.Warningf(ctx, "error during schema cleanup, retrying: %v", err)
				} else {
					__antithesis_instrumentation__.Notify(627958)
				}
				__antithesis_instrumentation__.Notify(627956)
				return err
			},
		)
	}
	__antithesis_instrumentation__.Notify(627946)

	if c.codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(627959)

		isLeaseHolder, err := c.isMeta1LeaseholderFunc(ctx, c.db.Clock().NowAsClockTimestamp())
		if err != nil {
			__antithesis_instrumentation__.Notify(627961)
			return err
		} else {
			__antithesis_instrumentation__.Notify(627962)
		}
		__antithesis_instrumentation__.Notify(627960)

		if !isLeaseHolder {
			__antithesis_instrumentation__.Notify(627963)
			log.Infof(ctx, "skipping temporary object cleanup run as it is not the leaseholder")
			return nil
		} else {
			__antithesis_instrumentation__.Notify(627964)
		}
	} else {
		__antithesis_instrumentation__.Notify(627965)
	}
	__antithesis_instrumentation__.Notify(627947)

	c.metrics.ActiveCleaners.Inc(1)
	defer c.metrics.ActiveCleaners.Dec(1)

	log.Infof(ctx, "running temporary object cleanup background job")

	txn := kv.NewTxn(ctx, c.db, 0)

	waitTimeForCreation := TempObjectWaitInterval.Get(&c.settings.SV)

	var allDbDescs []catalog.DatabaseDescriptor
	descsCol := c.collectionFactory.NewCollection(ctx, nil)
	if err := retryFunc(ctx, func() error {
		__antithesis_instrumentation__.Notify(627966)
		var err error
		allDbDescs, err = descsCol.GetAllDatabaseDescriptors(ctx, txn)
		return err
	}); err != nil {
		__antithesis_instrumentation__.Notify(627967)
		return err
	} else {
		__antithesis_instrumentation__.Notify(627968)
	}
	__antithesis_instrumentation__.Notify(627948)

	sessionIDs := make(map[ClusterWideID]struct{})
	for _, dbDesc := range allDbDescs {
		__antithesis_instrumentation__.Notify(627969)
		var schemaEntries map[descpb.ID]resolver.SchemaEntryForDB
		if err := retryFunc(ctx, func() error {
			__antithesis_instrumentation__.Notify(627971)
			var err error
			schemaEntries, err = resolver.GetForDatabase(ctx, txn, c.codec, dbDesc)
			return err
		}); err != nil {
			__antithesis_instrumentation__.Notify(627972)
			return err
		} else {
			__antithesis_instrumentation__.Notify(627973)
		}
		__antithesis_instrumentation__.Notify(627970)
		for _, scEntry := range schemaEntries {
			__antithesis_instrumentation__.Notify(627974)

			if !scEntry.Timestamp.Less(txn.ReadTimestamp().Add(-waitTimeForCreation.Nanoseconds(), 0)) {
				__antithesis_instrumentation__.Notify(627977)
				continue
			} else {
				__antithesis_instrumentation__.Notify(627978)
			}
			__antithesis_instrumentation__.Notify(627975)
			isTempSchema, sessionID, err := temporarySchemaSessionID(scEntry.Name)
			if err != nil {
				__antithesis_instrumentation__.Notify(627979)

				log.Warningf(ctx, "could not parse %q as temporary schema name", scEntry)
				continue
			} else {
				__antithesis_instrumentation__.Notify(627980)
			}
			__antithesis_instrumentation__.Notify(627976)
			if isTempSchema {
				__antithesis_instrumentation__.Notify(627981)
				sessionIDs[sessionID] = struct{}{}
			} else {
				__antithesis_instrumentation__.Notify(627982)
			}
		}
	}
	__antithesis_instrumentation__.Notify(627949)
	log.Infof(ctx, "found %d temporary schemas", len(sessionIDs))

	if len(sessionIDs) == 0 {
		__antithesis_instrumentation__.Notify(627983)
		log.Infof(ctx, "early exiting temporary schema cleaner as no temporary schemas were found")
		return nil
	} else {
		__antithesis_instrumentation__.Notify(627984)
	}
	__antithesis_instrumentation__.Notify(627950)

	var response *serverpb.ListSessionsResponse
	if err := retryFunc(ctx, func() error {
		__antithesis_instrumentation__.Notify(627985)
		var err error
		response, err = c.statusServer.ListSessions(
			ctx,
			&serverpb.ListSessionsRequest{},
		)
		if response != nil && func() bool {
			__antithesis_instrumentation__.Notify(627987)
			return len(response.Errors) > 0 == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(627988)
			return err == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(627989)
			return errors.Newf("fan out rpc failed with %s on node %d", response.Errors[0].Message, response.Errors[0].NodeID)
		} else {
			__antithesis_instrumentation__.Notify(627990)
		}
		__antithesis_instrumentation__.Notify(627986)
		return err
	}); err != nil {
		__antithesis_instrumentation__.Notify(627991)
		return err
	} else {
		__antithesis_instrumentation__.Notify(627992)
	}
	__antithesis_instrumentation__.Notify(627951)
	activeSessions := make(map[uint128.Uint128]struct{})
	for _, session := range response.Sessions {
		__antithesis_instrumentation__.Notify(627993)
		activeSessions[uint128.FromBytes(session.ID)] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(627952)

	ie := c.makeSessionBoundInternalExecutor(ctx, &sessiondata.SessionData{})
	for sessionID := range sessionIDs {
		__antithesis_instrumentation__.Notify(627994)
		if _, ok := activeSessions[sessionID.Uint128]; !ok {
			__antithesis_instrumentation__.Notify(627995)
			log.Eventf(ctx, "cleaning up temporary object for session %q", sessionID)
			c.metrics.SchemasToDelete.Inc(1)

			if err := retryFunc(ctx, func() error {
				__antithesis_instrumentation__.Notify(627996)
				return cleanupSessionTempObjects(
					ctx,
					c.settings,
					c.collectionFactory,
					c.db,
					c.codec,
					ie,
					sessionID,
				)
			}); err != nil {
				__antithesis_instrumentation__.Notify(627997)

				log.Warningf(ctx, "failed to clean temp objects under session %q: %v", sessionID, err)
				c.metrics.SchemasDeletionError.Inc(1)
			} else {
				__antithesis_instrumentation__.Notify(627998)
				c.metrics.SchemasDeletionSuccess.Inc(1)
				telemetry.Inc(sqltelemetry.TempObjectCleanerDeletionCounter)
			}
		} else {
			__antithesis_instrumentation__.Notify(627999)
			log.Eventf(ctx, "not cleaning up %q as session is still active", sessionID)
		}
	}
	__antithesis_instrumentation__.Notify(627953)

	return nil
}

func (c *TemporaryObjectCleaner) Start(ctx context.Context, stopper *stop.Stopper) {
	__antithesis_instrumentation__.Notify(628000)
	_ = stopper.RunAsyncTask(ctx, "object-cleaner", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(628001)
		nextTick := timeutil.Now()
		for {
			__antithesis_instrumentation__.Notify(628002)
			nextTickCh := time.After(nextTick.Sub(timeutil.Now()))
			if c.testingKnobs.TempObjectsCleanupCh != nil {
				__antithesis_instrumentation__.Notify(628006)
				nextTickCh = c.testingKnobs.TempObjectsCleanupCh
			} else {
				__antithesis_instrumentation__.Notify(628007)
			}
			__antithesis_instrumentation__.Notify(628003)

			select {
			case <-nextTickCh:
				__antithesis_instrumentation__.Notify(628008)
				if err := c.doTemporaryObjectCleanup(ctx, stopper.ShouldQuiesce()); err != nil {
					__antithesis_instrumentation__.Notify(628011)
					log.Warningf(ctx, "failed to clean temp objects: %v", err)
				} else {
					__antithesis_instrumentation__.Notify(628012)
				}
			case <-stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(628009)
				return
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(628010)
				return
			}
			__antithesis_instrumentation__.Notify(628004)
			if c.testingKnobs.OnTempObjectsCleanupDone != nil {
				__antithesis_instrumentation__.Notify(628013)
				c.testingKnobs.OnTempObjectsCleanupDone()
			} else {
				__antithesis_instrumentation__.Notify(628014)
			}
			__antithesis_instrumentation__.Notify(628005)
			nextTick = nextTick.Add(TempObjectCleanupInterval.Get(&c.settings.SV))
			log.Infof(ctx, "temporary object cleaner next scheduled to run at %s", nextTick)
		}
	})
}
