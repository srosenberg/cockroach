package startupmigrations

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/bootstrap"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/startupmigrations/leasemanager"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

var (
	leaseDuration        = time.Minute
	leaseRefreshInterval = leaseDuration / 5
)

type MigrationManagerTestingKnobs struct {
	DisableBackfillMigrations bool
	AfterJobMigration         func()

	AlwaysRunJobMigration bool

	AfterEnsureMigrations func()
}

func (*MigrationManagerTestingKnobs) ModuleTestingKnobs() {
	__antithesis_instrumentation__.Notify(633308)
}

var _ base.ModuleTestingKnobs = &MigrationManagerTestingKnobs{}

var backwardCompatibleMigrations = []migrationDescriptor{
	{

		name: "default UniqueID to uuid_v4 in system.eventlog",
	},
	{

		name: "create system.jobs table",
	},
	{

		name: "create system.settings table",
	},
	{

		name:   "enable diagnostics reporting",
		workFn: optInToDiagnosticsStatReporting,
	},
	{

		name: "establish conservative dependencies for views #17280 #17269 #17306",
	},
	{

		name: "create system.sessions table",
	},
	{

		name:        "populate initial version cluster setting table entry",
		workFn:      populateVersionSetting,
		clusterWide: true,
	},
	{

		name:             "create system.table_statistics table",
		newDescriptorIDs: staticIDs(keys.TableStatisticsTableID),
	},
	{

		name:   "add root user",
		workFn: addRootUser,
	},
	{

		name:             "create system.locations table",
		newDescriptorIDs: staticIDs(keys.LocationsTableID),
	},
	{

		name: "add default .meta and .liveness zone configs",
	},
	{

		name:             "create system.role_members table",
		newDescriptorIDs: staticIDs(keys.RoleMembersTableID),
	},
	{

		name:   "add system.users isRole column and create admin role",
		workFn: addAdminRole,
	},
	{

		name: "grant superuser privileges on all objects to the admin role",
	},
	{

		name:   "make root a member of the admin role",
		workFn: addRootToAdminRole,
	},
	{

		name: "upgrade table descs to interleaved format version",
	},
	{

		name: "remove cluster setting `kv.gc.batch_size`",
	},
	{

		name: "remove cluster setting `kv.transaction.max_intents`",
	},
	{

		name: "add default system.jobs zone config",
	},
	{

		name:   "initialize cluster.secret",
		workFn: initializeClusterSecret,
	},
	{

		name: "ensure admin role privileges in all descriptors",
	},
	{

		name: "repeat: ensure admin role privileges in all descriptors",
	},
	{

		name:   "disallow public user or role name",
		workFn: disallowPublicUserOrRole,
	},
	{

		name:             "create default databases",
		workFn:           createDefaultDbs,
		newDescriptorIDs: databaseIDs(catalogkeys.DefaultDatabaseName, catalogkeys.PgDatabaseName),
	},
	{

		name: "add progress to system.jobs",
	},
	{

		name: "create system.comment table",
	},
	{

		name: "create system.replication_constraint_stats table",
	},
	{

		name: "create system.replication_critical_localities table",
	},
	{

		name: "create system.reports_meta table",
	},
	{

		name: "create system.replication_stats table",
	},
	{

		name:        "propagate the ts purge interval to the new setting names",
		workFn:      retireOldTsPurgeIntervalSettings,
		clusterWide: true,
	},
	{

		name:   "update system.locations with default location data",
		workFn: updateSystemLocationData,
	},
	{

		name: "change reports fields from timestamp to timestamptz",
	},
	{

		name: "create system.protected_ts_meta table",
	},
	{

		name: "create system.protected_ts_records table",
	},
	{

		name: "create new system.namespace table v2",
	},
	{

		name: "migrate system.namespace_deprecated entries into system.namespace",
	},
	{

		name: "create system.role_options table",
	},
	{

		name: "create statement_diagnostics_requests, statement_diagnostics and " +
			"system.statement_bundle_chunks tables",
	},
	{

		name: "add CREATEROLE privilege to admin/root",
	},
	{

		name: "add created_by columns to system.jobs",
	},
	{

		name:             "create new system.scheduled_jobs table",
		newDescriptorIDs: staticIDs(keys.ScheduledJobsTableID),
	},
	{

		name: "add new sqlliveness table and claim columns to system.jobs",
	},
	{

		name: "create new system.tenants table",

		includedInBootstrap: roachpb.Version{Major: 20, Minor: 2},
		newDescriptorIDs:    staticIDs(keys.TenantsTableID),
	},
	{

		name: "alter scheduled jobs",
	},
	{

		name:   "add CREATELOGIN privilege to roles with CREATEROLE",
		workFn: extendCreateRoleWithCreateLogin,
	},
	{

		name: "mark non-terminal schema change jobs with a pre-20.1 format version as failed",
	},
}

func staticIDs(
	ids ...descpb.ID,
) func(ctx context.Context, db DB, codec keys.SQLCodec) ([]descpb.ID, error) {
	__antithesis_instrumentation__.Notify(633309)
	return func(ctx context.Context, db DB, codec keys.SQLCodec) ([]descpb.ID, error) {
		__antithesis_instrumentation__.Notify(633310)
		return ids, nil
	}
}

func databaseIDs(
	names ...string,
) func(ctx context.Context, db DB, codec keys.SQLCodec) ([]descpb.ID, error) {
	__antithesis_instrumentation__.Notify(633311)
	return func(ctx context.Context, db DB, codec keys.SQLCodec) ([]descpb.ID, error) {
		__antithesis_instrumentation__.Notify(633312)
		var ids []descpb.ID
		for _, name := range names {
			__antithesis_instrumentation__.Notify(633314)
			kv, err := db.Get(ctx, catalogkeys.MakeDatabaseNameKey(codec, name))
			if err != nil {
				__antithesis_instrumentation__.Notify(633316)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(633317)
			}
			__antithesis_instrumentation__.Notify(633315)
			ids = append(ids, descpb.ID(kv.ValueInt()))
		}
		__antithesis_instrumentation__.Notify(633313)
		return ids, nil
	}
}

type migrationDescriptor struct {
	name string

	workFn func(context.Context, runner) error

	includedInBootstrap roachpb.Version

	doesBackfill bool

	clusterWide bool

	newDescriptorIDs func(ctx context.Context, db DB, codec keys.SQLCodec) ([]descpb.ID, error)
}

func init() {

	names := make(map[string]struct{}, len(backwardCompatibleMigrations))
	for _, migration := range backwardCompatibleMigrations {
		name := migration.name
		if _, ok := names[name]; ok {
			log.Fatalf(context.Background(), "duplicate sql migration %q", name)
		}
		names[name] = struct{}{}
	}
}

type runner struct {
	db          DB
	codec       keys.SQLCodec
	sqlExecutor *sql.InternalExecutor
	settings    *cluster.Settings
}

func (r runner) execAsRoot(ctx context.Context, opName, stmt string, qargs ...interface{}) error {
	__antithesis_instrumentation__.Notify(633318)
	_, err := r.sqlExecutor.ExecEx(ctx, opName, nil,
		sessiondata.InternalExecutorOverride{
			User: security.RootUserName(),
		},
		stmt, qargs...)
	return err
}

func (r runner) execAsRootWithRetry(
	ctx context.Context, opName string, stmt string, qargs ...interface{},
) error {
	__antithesis_instrumentation__.Notify(633319)

	var err error
	for retry := retry.Start(retry.Options{MaxRetries: 5}); retry.Next(); {
		__antithesis_instrumentation__.Notify(633321)
		err := r.execAsRoot(ctx, opName, stmt, qargs...)
		if err == nil {
			__antithesis_instrumentation__.Notify(633323)
			break
		} else {
			__antithesis_instrumentation__.Notify(633324)
		}
		__antithesis_instrumentation__.Notify(633322)
		log.Warningf(ctx, "failed to run %s: %v", stmt, err)
	}
	__antithesis_instrumentation__.Notify(633320)
	return err
}

type leaseManager interface {
	AcquireLease(ctx context.Context, key roachpb.Key) (*leasemanager.Lease, error)
	ExtendLease(ctx context.Context, l *leasemanager.Lease) error
	ReleaseLease(ctx context.Context, l *leasemanager.Lease) error
	TimeRemaining(l *leasemanager.Lease) time.Duration
}

type DB interface {
	Scan(ctx context.Context, begin, end interface{}, maxRows int64) ([]kv.KeyValue, error)
	Get(ctx context.Context, key interface{}) (kv.KeyValue, error)
	Put(ctx context.Context, key, value interface{}) error
	Txn(ctx context.Context, retryable func(ctx context.Context, txn *kv.Txn) error) error
}

type Manager struct {
	stopper      *stop.Stopper
	leaseManager leaseManager
	db           DB
	codec        keys.SQLCodec
	sqlExecutor  *sql.InternalExecutor
	testingKnobs MigrationManagerTestingKnobs
	settings     *cluster.Settings
	jobRegistry  *jobs.Registry
}

func NewManager(
	stopper *stop.Stopper,
	db *kv.DB,
	codec keys.SQLCodec,
	executor *sql.InternalExecutor,
	clock *hlc.Clock,
	testingKnobs MigrationManagerTestingKnobs,
	clientID string,
	settings *cluster.Settings,
	registry *jobs.Registry,
) *Manager {
	__antithesis_instrumentation__.Notify(633325)
	opts := leasemanager.Options{
		ClientID:      clientID,
		LeaseDuration: leaseDuration,
	}
	return &Manager{
		stopper:      stopper,
		leaseManager: leasemanager.New(db, clock, opts),
		db:           db,
		codec:        codec,
		sqlExecutor:  executor,
		testingKnobs: testingKnobs,
		settings:     settings,
		jobRegistry:  registry,
	}
}

func ExpectedDescriptorIDs(
	ctx context.Context,
	db DB,
	codec keys.SQLCodec,
	defaultZoneConfig *zonepb.ZoneConfig,
	defaultSystemZoneConfig *zonepb.ZoneConfig,
) (descpb.IDs, error) {
	__antithesis_instrumentation__.Notify(633326)
	completedMigrations, err := getCompletedMigrations(ctx, db, codec)
	if err != nil {
		__antithesis_instrumentation__.Notify(633329)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(633330)
	}
	__antithesis_instrumentation__.Notify(633327)
	descriptorIDs := bootstrap.MakeMetadataSchema(codec, defaultZoneConfig, defaultSystemZoneConfig).DescriptorIDs()
	for _, migration := range backwardCompatibleMigrations {
		__antithesis_instrumentation__.Notify(633331)

		if migration.newDescriptorIDs == nil || func() bool {
			__antithesis_instrumentation__.Notify(633333)
			return (migration.includedInBootstrap != roachpb.Version{}) == true
		}() == true {
			__antithesis_instrumentation__.Notify(633334)
			continue
		} else {
			__antithesis_instrumentation__.Notify(633335)
		}
		__antithesis_instrumentation__.Notify(633332)
		if _, ok := completedMigrations[string(migrationKey(codec, migration))]; ok {
			__antithesis_instrumentation__.Notify(633336)
			newIDs, err := migration.newDescriptorIDs(ctx, db, codec)
			if err != nil {
				__antithesis_instrumentation__.Notify(633338)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(633339)
			}
			__antithesis_instrumentation__.Notify(633337)
			descriptorIDs = append(descriptorIDs, newIDs...)
		} else {
			__antithesis_instrumentation__.Notify(633340)
		}
	}
	__antithesis_instrumentation__.Notify(633328)
	sort.Sort(descriptorIDs)
	return descriptorIDs, nil
}

func (m *Manager) EnsureMigrations(ctx context.Context, bootstrapVersion roachpb.Version) error {
	__antithesis_instrumentation__.Notify(633341)
	if m.testingKnobs.AfterEnsureMigrations != nil {
		__antithesis_instrumentation__.Notify(633353)
		defer m.testingKnobs.AfterEnsureMigrations()
	} else {
		__antithesis_instrumentation__.Notify(633354)
	}
	__antithesis_instrumentation__.Notify(633342)

	completedMigrations, err := getCompletedMigrations(ctx, m.db, m.codec)
	if err != nil {
		__antithesis_instrumentation__.Notify(633355)
		return err
	} else {
		__antithesis_instrumentation__.Notify(633356)
	}
	__antithesis_instrumentation__.Notify(633343)
	allMigrationsCompleted := true
	for _, migration := range backwardCompatibleMigrations {
		__antithesis_instrumentation__.Notify(633357)
		if !m.shouldRunMigration(migration, bootstrapVersion) {
			__antithesis_instrumentation__.Notify(633360)
			continue
		} else {
			__antithesis_instrumentation__.Notify(633361)
		}
		__antithesis_instrumentation__.Notify(633358)
		if m.testingKnobs.DisableBackfillMigrations && func() bool {
			__antithesis_instrumentation__.Notify(633362)
			return migration.doesBackfill == true
		}() == true {
			__antithesis_instrumentation__.Notify(633363)
			log.Infof(ctx, "ignoring migrations after (and including) %s due to testing knob",
				migration.name)
			break
		} else {
			__antithesis_instrumentation__.Notify(633364)
		}
		__antithesis_instrumentation__.Notify(633359)
		key := migrationKey(m.codec, migration)
		if _, ok := completedMigrations[string(key)]; !ok {
			__antithesis_instrumentation__.Notify(633365)
			allMigrationsCompleted = false
		} else {
			__antithesis_instrumentation__.Notify(633366)
		}
	}
	__antithesis_instrumentation__.Notify(633344)
	if allMigrationsCompleted {
		__antithesis_instrumentation__.Notify(633367)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(633368)
	}
	__antithesis_instrumentation__.Notify(633345)

	var lease *leasemanager.Lease
	if log.V(1) {
		__antithesis_instrumentation__.Notify(633369)
		log.Info(ctx, "trying to acquire lease")
	} else {
		__antithesis_instrumentation__.Notify(633370)
	}
	__antithesis_instrumentation__.Notify(633346)
	for r := retry.StartWithCtx(ctx, base.DefaultRetryOptions()); r.Next(); {
		__antithesis_instrumentation__.Notify(633371)
		lease, err = m.leaseManager.AcquireLease(ctx, m.codec.MigrationLeaseKey())
		if err == nil {
			__antithesis_instrumentation__.Notify(633373)
			break
		} else {
			__antithesis_instrumentation__.Notify(633374)
		}
		__antithesis_instrumentation__.Notify(633372)
		log.Infof(ctx, "failed attempt to acquire migration lease: %s", err)
	}
	__antithesis_instrumentation__.Notify(633347)
	if err != nil {
		__antithesis_instrumentation__.Notify(633375)
		return errors.Wrapf(err, "failed to acquire lease for running necessary migrations")
	} else {
		__antithesis_instrumentation__.Notify(633376)
	}
	__antithesis_instrumentation__.Notify(633348)

	done := make(chan interface{}, 1)
	defer func() {
		__antithesis_instrumentation__.Notify(633377)
		done <- nil
		if log.V(1) {
			__antithesis_instrumentation__.Notify(633379)
			log.Info(ctx, "trying to release the lease")
		} else {
			__antithesis_instrumentation__.Notify(633380)
		}
		__antithesis_instrumentation__.Notify(633378)
		if err := m.leaseManager.ReleaseLease(ctx, lease); err != nil {
			__antithesis_instrumentation__.Notify(633381)
			log.Errorf(ctx, "failed to release migration lease: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(633382)
		}
	}()
	__antithesis_instrumentation__.Notify(633349)
	if err := m.stopper.RunAsyncTask(ctx, "migrations.Manager: lease watcher",
		func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(633383)
			select {
			case <-done:
				__antithesis_instrumentation__.Notify(633384)
				return
			case <-time.After(leaseRefreshInterval):
				__antithesis_instrumentation__.Notify(633385)
				if err := m.leaseManager.ExtendLease(ctx, lease); err != nil {
					__antithesis_instrumentation__.Notify(633387)
					log.Warningf(ctx, "unable to extend ownership of expiration lease: %s", err)
				} else {
					__antithesis_instrumentation__.Notify(633388)
				}
				__antithesis_instrumentation__.Notify(633386)
				if m.leaseManager.TimeRemaining(lease) < leaseRefreshInterval {
					__antithesis_instrumentation__.Notify(633389)

					select {
					case <-done:
						__antithesis_instrumentation__.Notify(633390)
						return
					default:
						__antithesis_instrumentation__.Notify(633391)

						log.Fatal(ctx, "not enough time left on migration lease, terminating for safety")
					}
				} else {
					__antithesis_instrumentation__.Notify(633392)
				}
			}
		}); err != nil {
		__antithesis_instrumentation__.Notify(633393)
		return err
	} else {
		__antithesis_instrumentation__.Notify(633394)
	}
	__antithesis_instrumentation__.Notify(633350)

	completedMigrations, err = getCompletedMigrations(ctx, m.db, m.codec)
	if err != nil {
		__antithesis_instrumentation__.Notify(633395)
		return err
	} else {
		__antithesis_instrumentation__.Notify(633396)
	}
	__antithesis_instrumentation__.Notify(633351)

	startTime := timeutil.Now().String()
	r := runner{
		db:          m.db,
		codec:       m.codec,
		sqlExecutor: m.sqlExecutor,
		settings:    m.settings,
	}
	for _, migration := range backwardCompatibleMigrations {
		__antithesis_instrumentation__.Notify(633397)
		if !m.shouldRunMigration(migration, bootstrapVersion) {
			__antithesis_instrumentation__.Notify(633403)
			continue
		} else {
			__antithesis_instrumentation__.Notify(633404)
		}
		__antithesis_instrumentation__.Notify(633398)

		key := migrationKey(m.codec, migration)
		if _, ok := completedMigrations[string(key)]; ok {
			__antithesis_instrumentation__.Notify(633405)
			continue
		} else {
			__antithesis_instrumentation__.Notify(633406)
		}
		__antithesis_instrumentation__.Notify(633399)

		if m.testingKnobs.DisableBackfillMigrations && func() bool {
			__antithesis_instrumentation__.Notify(633407)
			return migration.doesBackfill == true
		}() == true {
			__antithesis_instrumentation__.Notify(633408)
			log.Infof(ctx, "ignoring migrations after (and including) %s due to testing knob",
				migration.name)
			break
		} else {
			__antithesis_instrumentation__.Notify(633409)
		}
		__antithesis_instrumentation__.Notify(633400)

		if log.V(1) {
			__antithesis_instrumentation__.Notify(633410)
			log.Infof(ctx, "running migration %q", migration.name)
		} else {
			__antithesis_instrumentation__.Notify(633411)
		}
		__antithesis_instrumentation__.Notify(633401)
		if err := migration.workFn(ctx, r); err != nil {
			__antithesis_instrumentation__.Notify(633412)
			return errors.Wrapf(err, "failed to run migration %q", migration.name)
		} else {
			__antithesis_instrumentation__.Notify(633413)
		}
		__antithesis_instrumentation__.Notify(633402)

		log.VEventf(ctx, 1, "persisting record of completing migration %s", migration.name)
		if err := m.db.Put(ctx, key, startTime); err != nil {
			__antithesis_instrumentation__.Notify(633414)
			return errors.Wrapf(err, "failed to persist record of completing migration %q",
				migration.name)
		} else {
			__antithesis_instrumentation__.Notify(633415)
		}
	}
	__antithesis_instrumentation__.Notify(633352)

	return nil
}

func (m *Manager) shouldRunMigration(
	migration migrationDescriptor, bootstrapVersion roachpb.Version,
) bool {
	__antithesis_instrumentation__.Notify(633416)
	if migration.workFn == nil {
		__antithesis_instrumentation__.Notify(633420)

		return false
	} else {
		__antithesis_instrumentation__.Notify(633421)
	}
	__antithesis_instrumentation__.Notify(633417)
	minVersion := migration.includedInBootstrap
	if minVersion != (roachpb.Version{}) && func() bool {
		__antithesis_instrumentation__.Notify(633422)
		return !bootstrapVersion.Less(minVersion) == true
	}() == true {
		__antithesis_instrumentation__.Notify(633423)

		return false
	} else {
		__antithesis_instrumentation__.Notify(633424)
	}
	__antithesis_instrumentation__.Notify(633418)
	if migration.clusterWide && func() bool {
		__antithesis_instrumentation__.Notify(633425)
		return !m.codec.ForSystemTenant() == true
	}() == true {
		__antithesis_instrumentation__.Notify(633426)

		return false
	} else {
		__antithesis_instrumentation__.Notify(633427)
	}
	__antithesis_instrumentation__.Notify(633419)
	return true
}

func getCompletedMigrations(
	ctx context.Context, db DB, codec keys.SQLCodec,
) (map[string]struct{}, error) {
	__antithesis_instrumentation__.Notify(633428)
	if log.V(1) {
		__antithesis_instrumentation__.Notify(633432)
		log.Info(ctx, "trying to get the list of completed migrations")
	} else {
		__antithesis_instrumentation__.Notify(633433)
	}
	__antithesis_instrumentation__.Notify(633429)
	prefix := codec.MigrationKeyPrefix()
	keyvals, err := db.Scan(ctx, prefix, prefix.PrefixEnd(), 0)
	if err != nil {
		__antithesis_instrumentation__.Notify(633434)
		return nil, errors.Wrapf(err, "failed to get list of completed migrations")
	} else {
		__antithesis_instrumentation__.Notify(633435)
	}
	__antithesis_instrumentation__.Notify(633430)
	completedMigrations := make(map[string]struct{})
	for _, keyval := range keyvals {
		__antithesis_instrumentation__.Notify(633436)
		completedMigrations[string(keyval.Key)] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(633431)
	return completedMigrations, nil
}

func migrationKey(codec keys.SQLCodec, migration migrationDescriptor) roachpb.Key {
	__antithesis_instrumentation__.Notify(633437)
	return append(codec.MigrationKeyPrefix(), roachpb.RKey(migration.name)...)
}

func extendCreateRoleWithCreateLogin(ctx context.Context, r runner) error {
	__antithesis_instrumentation__.Notify(633438)

	const upsertCreateRoleStmt = `
     UPSERT INTO system.role_options (username, option, value)
        SELECT username, 'CREATELOGIN', NULL
          FROM system.role_options
         WHERE option = 'CREATEROLE'
     `
	return r.execAsRootWithRetry(ctx,
		"add CREATELOGIN where a role already has CREATEROLE",
		upsertCreateRoleStmt)
}

var SettingsDefaultOverrides = map[string]string{
	"diagnostics.reporting.enabled": "true",
	"cluster.secret":                "<random>",
}

func optInToDiagnosticsStatReporting(ctx context.Context, r runner) error {
	__antithesis_instrumentation__.Notify(633439)

	if cluster.TelemetryOptOut() {
		__antithesis_instrumentation__.Notify(633441)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(633442)
	}
	__antithesis_instrumentation__.Notify(633440)
	return r.execAsRootWithRetry(ctx, "optInToDiagnosticsStatReporting",
		`SET CLUSTER SETTING diagnostics.reporting.enabled = true`)
}

func initializeClusterSecret(ctx context.Context, r runner) error {
	__antithesis_instrumentation__.Notify(633443)
	return r.execAsRootWithRetry(
		ctx, "initializeClusterSecret",
		`SET CLUSTER SETTING cluster.secret = gen_random_uuid()::STRING`,
	)
}

func populateVersionSetting(ctx context.Context, r runner) error {
	__antithesis_instrumentation__.Notify(633444)
	var v roachpb.Version
	if err := r.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(633450)
		return txn.GetProto(ctx, keys.BootstrapVersionKey, &v)
	}); err != nil {
		__antithesis_instrumentation__.Notify(633451)
		return err
	} else {
		__antithesis_instrumentation__.Notify(633452)
	}
	__antithesis_instrumentation__.Notify(633445)
	if v == (roachpb.Version{}) {
		__antithesis_instrumentation__.Notify(633453)

		v = clusterversion.TestingBinaryMinSupportedVersion
	} else {
		__antithesis_instrumentation__.Notify(633454)
	}
	__antithesis_instrumentation__.Notify(633446)

	b, err := protoutil.Marshal(&clusterversion.ClusterVersion{Version: v})
	if err != nil {
		__antithesis_instrumentation__.Notify(633455)
		return errors.Wrap(err, "while marshaling version")
	} else {
		__antithesis_instrumentation__.Notify(633456)
	}
	__antithesis_instrumentation__.Notify(633447)

	if err := r.execAsRoot(
		ctx,
		"insert-setting",
		fmt.Sprintf(`INSERT INTO system.settings (name, value, "lastUpdated", "valueType") VALUES ('version', x'%x', now(), 'm') ON CONFLICT(name) DO NOTHING`, b),
	); err != nil {
		__antithesis_instrumentation__.Notify(633457)
		return err
	} else {
		__antithesis_instrumentation__.Notify(633458)
	}
	__antithesis_instrumentation__.Notify(633448)

	if err := r.execAsRootWithRetry(
		ctx, "set-setting", "SET CLUSTER SETTING version = $1", v.String(),
	); err != nil {
		__antithesis_instrumentation__.Notify(633459)
		return err
	} else {
		__antithesis_instrumentation__.Notify(633460)
	}
	__antithesis_instrumentation__.Notify(633449)
	return nil
}

func addRootUser(ctx context.Context, r runner) error {
	__antithesis_instrumentation__.Notify(633461)

	const upsertRootStmt = `
	        UPSERT INTO system.users (username, "hashedPassword", "isRole") VALUES ($1, '', false)
	        `
	return r.execAsRootWithRetry(ctx, "addRootUser", upsertRootStmt, security.RootUser)
}

func addAdminRole(ctx context.Context, r runner) error {
	__antithesis_instrumentation__.Notify(633462)

	const upsertAdminStmt = `
          UPSERT INTO system.users (username, "hashedPassword", "isRole") VALUES ($1, '', true)
          `
	return r.execAsRootWithRetry(ctx, "addAdminRole", upsertAdminStmt, security.AdminRole)
}

func addRootToAdminRole(ctx context.Context, r runner) error {
	__antithesis_instrumentation__.Notify(633463)

	const upsertAdminStmt = `
          UPSERT INTO system.role_members ("role", "member", "isAdmin") VALUES ($1, $2, true)
          `
	return r.execAsRootWithRetry(
		ctx, "addRootToAdminRole", upsertAdminStmt, security.AdminRole, security.RootUser)
}

func disallowPublicUserOrRole(ctx context.Context, r runner) error {
	__antithesis_instrumentation__.Notify(633464)

	const selectPublicStmt = `
          SELECT username, "isRole" from system.users WHERE username = $1
          `

	for retry := retry.Start(retry.Options{MaxRetries: 5}); retry.Next(); {
		__antithesis_instrumentation__.Notify(633466)
		row, err := r.sqlExecutor.QueryRowEx(
			ctx, "disallowPublicUserOrRole", nil,
			sessiondata.InternalExecutorOverride{
				User: security.RootUserName(),
			},
			selectPublicStmt, security.PublicRole,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(633471)
			continue
		} else {
			__antithesis_instrumentation__.Notify(633472)
		}
		__antithesis_instrumentation__.Notify(633467)
		if row == nil {
			__antithesis_instrumentation__.Notify(633473)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(633474)
		}
		__antithesis_instrumentation__.Notify(633468)

		isRole, ok := tree.AsDBool(row[1])
		if !ok {
			__antithesis_instrumentation__.Notify(633475)
			log.Fatalf(ctx, "expected 'isRole' column of system.users to be of type bool, got %v", row)
		} else {
			__antithesis_instrumentation__.Notify(633476)
		}
		__antithesis_instrumentation__.Notify(633469)

		if isRole {
			__antithesis_instrumentation__.Notify(633477)
			return fmt.Errorf(`found a role named %s which is now a reserved name. Please drop the role `+
				`(DROP ROLE %s) using a previous version of CockroachDB and try again`,
				security.PublicRole, security.PublicRole)
		} else {
			__antithesis_instrumentation__.Notify(633478)
		}
		__antithesis_instrumentation__.Notify(633470)
		return fmt.Errorf(`found a user named %s which is now a reserved name. Please drop the role `+
			`(DROP USER %s) using a previous version of CockroachDB and try again`,
			security.PublicRole, security.PublicRole)
	}
	__antithesis_instrumentation__.Notify(633465)
	return nil
}

func createDefaultDbs(ctx context.Context, r runner) error {
	__antithesis_instrumentation__.Notify(633479)

	const createDbStmt = `CREATE DATABASE IF NOT EXISTS "%s"`

	var err error
	for retry := retry.Start(retry.Options{MaxRetries: 5}); retry.Next(); {
		__antithesis_instrumentation__.Notify(633481)
		for _, dbName := range []string{catalogkeys.DefaultDatabaseName, catalogkeys.PgDatabaseName} {
			__antithesis_instrumentation__.Notify(633483)
			stmt := fmt.Sprintf(createDbStmt, dbName)
			err = r.execAsRoot(ctx, "create-default-DB", stmt)
			if err != nil {
				__antithesis_instrumentation__.Notify(633484)
				log.Warningf(ctx, "failed attempt to add database %q: %s", dbName, err)
				break
			} else {
				__antithesis_instrumentation__.Notify(633485)
			}
		}
		__antithesis_instrumentation__.Notify(633482)
		if err == nil {
			__antithesis_instrumentation__.Notify(633486)
			break
		} else {
			__antithesis_instrumentation__.Notify(633487)
		}
	}
	__antithesis_instrumentation__.Notify(633480)
	return err
}

func retireOldTsPurgeIntervalSettings(ctx context.Context, r runner) error {
	__antithesis_instrumentation__.Notify(633488)

	if err := r.execAsRoot(ctx, "copy-setting", `
INSERT INTO system.settings (name, value, "lastUpdated", "valueType")
   SELECT 'timeseries.storage.resolution_10s.ttl', value, "lastUpdated", "valueType"
     FROM system.settings WHERE name = 'timeseries.storage.10s_resolution_ttl'
ON CONFLICT (name) DO NOTHING`,
	); err != nil {
		__antithesis_instrumentation__.Notify(633491)
		return err
	} else {
		__antithesis_instrumentation__.Notify(633492)
	}
	__antithesis_instrumentation__.Notify(633489)

	if err := r.execAsRoot(ctx, "copy-setting", `
INSERT INTO system.settings (name, value, "lastUpdated", "valueType")
   SELECT 'timeseries.storage.resolution_30m.ttl', value, "lastUpdated", "valueType"
     FROM system.settings WHERE name = 'timeseries.storage.30m_resolution_ttl'
ON CONFLICT (name) DO NOTHING`,
	); err != nil {
		__antithesis_instrumentation__.Notify(633493)
		return err
	} else {
		__antithesis_instrumentation__.Notify(633494)
	}
	__antithesis_instrumentation__.Notify(633490)

	return nil
}

func updateSystemLocationData(ctx context.Context, r runner) error {
	__antithesis_instrumentation__.Notify(633495)

	row, err := r.sqlExecutor.QueryRowEx(ctx, "update-system-locations",
		nil,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		`SELECT count(*) FROM system.locations`)
	if err != nil {
		__antithesis_instrumentation__.Notify(633500)
		return err
	} else {
		__antithesis_instrumentation__.Notify(633501)
	}
	__antithesis_instrumentation__.Notify(633496)
	if row == nil {
		__antithesis_instrumentation__.Notify(633502)
		return errors.New("failed to update system locations")
	} else {
		__antithesis_instrumentation__.Notify(633503)
	}
	__antithesis_instrumentation__.Notify(633497)
	count := int(tree.MustBeDInt(row[0]))
	if count != 0 {
		__antithesis_instrumentation__.Notify(633504)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(633505)
	}
	__antithesis_instrumentation__.Notify(633498)

	for _, loc := range roachpb.DefaultLocationInformation {
		__antithesis_instrumentation__.Notify(633506)
		stmt := `UPSERT INTO system.locations VALUES ($1, $2, $3, $4)`
		tier := loc.Locality.Tiers[0]
		if err := r.execAsRoot(ctx, "update-system-locations",
			stmt, tier.Key, tier.Value, loc.Latitude, loc.Longitude,
		); err != nil {
			__antithesis_instrumentation__.Notify(633507)
			return err
		} else {
			__antithesis_instrumentation__.Notify(633508)
		}
	}
	__antithesis_instrumentation__.Notify(633499)
	return nil
}
