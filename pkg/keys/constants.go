package keys

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
)

const (
	meta1PrefixByte  = roachpb.LocalMaxByte
	meta2PrefixByte  = '\x03'
	metaMaxByte      = '\x04'
	systemPrefixByte = metaMaxByte
	systemMaxByte    = '\x05'
	tenantPrefixByte = '\xfe'
)

const (
	appliedUnsafeReplicaRecoveryPrefix = "applied"
)

var (
	MinKey = roachpb.KeyMin

	MaxKey = roachpb.KeyMax

	LocalPrefix = roachpb.LocalPrefix

	LocalMax = roachpb.LocalMax

	localSuffixLength = 4

	LocalRangeIDPrefix = roachpb.RKey(makeKey(LocalPrefix, roachpb.Key("i")))

	LocalRangeIDReplicatedInfix = []byte("r")

	LocalAbortSpanSuffix = []byte("abc-")

	localRangeFrozenStatusSuffix = []byte("fzn-")

	LocalRangeGCThresholdSuffix = []byte("lgc-")

	LocalRangeAppliedStateSuffix = []byte("rask")

	_ = []byte("rftt")

	LocalRangeLeaseSuffix = []byte("rll-")

	LocalRangePriorReadSummarySuffix = []byte("rprs")

	LocalRangeVersionSuffix = []byte("rver")

	LocalRangeStatsLegacySuffix = []byte("stat")

	localTxnSpanGCThresholdSuffix = []byte("tst-")

	localRangeIDUnreplicatedInfix = []byte("u")

	LocalRangeTombstoneSuffix = []byte("rftb")

	LocalRaftHardStateSuffix = []byte("rfth")

	localRaftLastIndexSuffix = []byte("rfti")

	LocalRaftLogSuffix = []byte("rftl")

	LocalRaftReplicaIDSuffix = []byte("rftr")

	LocalRaftTruncatedStateSuffix = []byte("rftt")

	LocalRangeLastReplicaGCTimestampSuffix = []byte("rlrt")

	localRangeLastVerificationTimestampSuffix = []byte("rlvt")

	LocalRangePrefix = roachpb.Key(makeKey(LocalPrefix, roachpb.RKey("k")))
	LocalRangeMax    = LocalRangePrefix.PrefixEnd()

	LocalRangeProbeSuffix = roachpb.RKey("prbe")

	LocalQueueLastProcessedSuffix = roachpb.RKey("qlpt")

	LocalRangeDescriptorSuffix = roachpb.RKey("rdsc")

	LocalTransactionSuffix = roachpb.RKey("txn-")

	LocalStorePrefix = makeKey(LocalPrefix, roachpb.Key("s"))

	localStoreClusterVersionSuffix = []byte("cver")

	localStoreGossipSuffix = []byte("goss")

	localStoreHLCUpperBoundSuffix = []byte("hlcu")

	localStoreIdentSuffix = []byte("iden")

	localStoreUnsafeReplicaRecoverySuffix = makeKey([]byte("loqr"),
		[]byte(appliedUnsafeReplicaRecoveryPrefix))

	LocalStoreUnsafeReplicaRecoveryKeyMin = MakeStoreKey(localStoreUnsafeReplicaRecoverySuffix, nil)

	LocalStoreUnsafeReplicaRecoveryKeyMax = LocalStoreUnsafeReplicaRecoveryKeyMin.PrefixEnd()

	localStoreNodeTombstoneSuffix = []byte("ntmb")

	localStoreCachedSettingsSuffix = []byte("stng")

	LocalStoreCachedSettingsKeyMin = MakeStoreKey(localStoreCachedSettingsSuffix, nil)

	LocalStoreCachedSettingsKeyMax = LocalStoreCachedSettingsKeyMin.PrefixEnd()

	localStoreLastUpSuffix = []byte("uptm")

	localRemovedLeakedRaftEntriesSuffix = []byte("dlre")

	LocalRangeLockTablePrefix = roachpb.Key(makeKey(LocalPrefix, roachpb.RKey("z")))
	LockTableSingleKeyInfix   = []byte("k")

	LockTableSingleKeyStart = roachpb.Key(makeKey(LocalRangeLockTablePrefix, LockTableSingleKeyInfix))

	LockTableSingleKeyEnd = roachpb.Key(
		makeKey(LocalRangeLockTablePrefix, roachpb.Key(LockTableSingleKeyInfix).PrefixEnd()))

	MetaMin = Meta1Prefix

	MetaMax = roachpb.Key{metaMaxByte}

	Meta1Prefix = roachpb.Key{meta1PrefixByte}

	Meta1KeyMax = roachpb.Key(makeKey(Meta1Prefix, roachpb.RKeyMax))

	Meta2Prefix = roachpb.Key{meta2PrefixByte}

	Meta2KeyMax = roachpb.Key(makeKey(Meta2Prefix, roachpb.RKeyMax))

	SystemPrefix = roachpb.Key{systemPrefixByte}
	SystemMax    = roachpb.Key{systemMaxByte}

	NodeLivenessPrefix = roachpb.Key(makeKey(SystemPrefix, roachpb.RKey("\x00liveness-")))

	NodeLivenessKeyMax = NodeLivenessPrefix.PrefixEnd()

	BootstrapVersionKey = roachpb.Key(makeKey(SystemPrefix, roachpb.RKey("bootstrap-version")))

	descIDGenerator = roachpb.Key(makeKey(SystemPrefix, roachpb.RKey("desc-idgen")))

	NodeIDGenerator = roachpb.Key(makeKey(SystemPrefix, roachpb.RKey("node-idgen")))

	RangeIDGenerator = roachpb.Key(makeKey(SystemPrefix, roachpb.RKey("range-idgen")))

	StoreIDGenerator = roachpb.Key(makeKey(SystemPrefix, roachpb.RKey("store-idgen")))

	StatusPrefix = roachpb.Key(makeKey(SystemPrefix, roachpb.RKey("status-")))

	StatusNodePrefix = roachpb.Key(makeKey(StatusPrefix, roachpb.RKey("node-")))

	MigrationPrefix = roachpb.Key(makeKey(SystemPrefix, roachpb.RKey("system-version/")))

	MigrationLease = roachpb.Key(makeKey(MigrationPrefix, roachpb.RKey("lease")))

	TimeseriesPrefix = roachpb.Key(makeKey(SystemPrefix, roachpb.RKey("tsd")))

	TimeseriesKeyMax = TimeseriesPrefix.PrefixEnd()

	SystemSpanConfigPrefix = roachpb.Key(makeKey(SystemPrefix, roachpb.RKey("\xffsys-scfg")))

	SystemSpanConfigEntireKeyspace = roachpb.Key(makeKey(SystemSpanConfigPrefix, roachpb.RKey("host/all")))

	SystemSpanConfigHostOnTenantKeyspace = roachpb.Key(makeKey(SystemSpanConfigPrefix, roachpb.RKey("host/ten/")))

	SystemSpanConfigSecondaryTenantOnEntireKeyspace = roachpb.Key(makeKey(SystemSpanConfigPrefix, roachpb.RKey("ten/")))

	SystemSpanConfigKeyMax = SystemSpanConfigPrefix.PrefixEnd()

	TableDataMin = SystemSQLCodec.TablePrefix(0)

	TableDataMax = SystemSQLCodec.TablePrefix(math.MaxUint32).PrefixEnd()

	ScratchRangeMin = TableDataMax
	ScratchRangeMax = TenantPrefix

	SystemConfigSplitKey = []byte(TableDataMin)

	SystemConfigTableDataMax = SystemSQLCodec.TablePrefix(MaxSystemConfigDescID + 1)

	NamespaceTableMin = SystemSQLCodec.TablePrefix(NamespaceTableID)

	NamespaceTableMax = SystemSQLCodec.TablePrefix(NamespaceTableID + 1)

	TenantPrefix       = roachpb.Key{tenantPrefixByte}
	TenantTableDataMin = MakeTenantPrefix(roachpb.MinTenantID)
	TenantTableDataMax = MakeTenantPrefix(roachpb.MaxTenantID).PrefixEnd()
)

const (
	MaxSystemConfigDescID = 10

	MaxReservedDescID = 49

	RootNamespaceID = 0

	SystemDatabaseID = 1

	DeprecatedNamespaceTableID = 2
	DescriptorTableID          = 3
	UsersTableID               = 4
	ZonesTableID               = 5
	SettingsTableID            = 6
	DescIDSequenceID           = 7
	TenantsTableID             = 8

	ZonesTablePrimaryIndexID = 1
	ZonesTableConfigColumnID = 2
	ZonesTableConfigColFamID = 2

	DescriptorTablePrimaryKeyIndexID         = 1
	DescriptorTableDescriptorColID           = 2
	DescriptorTableDescriptorColFamID        = 2
	TenantsTablePrimaryKeyIndexID            = 1
	SpanConfigurationsTablePrimaryKeyIndexID = 1

	LeaseTableID                         = 11
	EventLogTableID                      = 12
	RangeEventTableID                    = 13
	UITableID                            = 14
	JobsTableID                          = 15
	MetaRangesID                         = 16
	SystemRangesID                       = 17
	TimeseriesRangesID                   = 18
	WebSessionsTableID                   = 19
	TableStatisticsTableID               = 20
	LocationsTableID                     = 21
	LivenessRangesID                     = 22
	RoleMembersTableID                   = 23
	CommentsTableID                      = 24
	ReplicationConstraintStatsTableID    = 25
	ReplicationCriticalLocalitiesTableID = 26
	ReplicationStatsTableID              = 27
	ReportsMetaTableID                   = 28

	PublicSchemaID = 29

	PublicSchemaIDForBackup = 29

	SystemPublicSchemaID = 29

	NamespaceTableID                    = 30
	ProtectedTimestampsMetaTableID      = 31
	ProtectedTimestampsRecordsTableID   = 32
	RoleOptionsTableID                  = 33
	StatementBundleChunksTableID        = 34
	StatementDiagnosticsRequestsTableID = 35
	StatementDiagnosticsTableID         = 36
	ScheduledJobsTableID                = 37
	TenantsRangesID                     = 38
	SqllivenessID                       = 39
	MigrationsID                        = 40
	JoinTokensTableID                   = 41
	StatementStatisticsTableID          = 42
	TransactionStatisticsTableID        = 43
	DatabaseRoleSettingsTableID         = 44
	TenantUsageTableID                  = 45
	SQLInstancesTableID                 = 46
	SpanConfigurationsTableID           = 47
)

type CommentType int

const (
	DatabaseCommentType CommentType = 0

	TableCommentType CommentType = 1

	ColumnCommentType CommentType = 2

	IndexCommentType CommentType = 3

	SchemaCommentType CommentType = 4

	ConstraintCommentType CommentType = 5
)

const (
	SequenceIndexID = 1

	SequenceColumnFamilyID = 0
)

var PseudoTableIDs = []uint32{
	MetaRangesID,
	SystemRangesID,
	TimeseriesRangesID,
	LivenessRangesID,
	PublicSchemaID,
	TenantsRangesID,
}

var MaxPseudoTableID = func() uint32 {
	__antithesis_instrumentation__.Notify(85212)
	var max uint32
	for _, id := range PseudoTableIDs {
		__antithesis_instrumentation__.Notify(85214)
		if max < id {
			__antithesis_instrumentation__.Notify(85215)
			max = id
		} else {
			__antithesis_instrumentation__.Notify(85216)
		}
	}
	__antithesis_instrumentation__.Notify(85213)
	return max
}()
