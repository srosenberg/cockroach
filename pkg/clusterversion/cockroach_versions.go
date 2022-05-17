package clusterversion

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/roachpb"

type Key int

const (
	_ Key = iota - 1

	V21_2

	Start22_1

	TargetBytesAvoidExcess

	AvoidDrainingNames

	DrainingNamesMigration

	TraceIDDoesntImplyStructuredRecording

	AlterSystemTableStatisticsAddAvgSizeCol

	AlterSystemStmtDiagReqs

	MVCCAddSSTable

	InsertPublicSchemaNamespaceEntryOnRestore

	UnsplitRangesInAsyncGCJobs

	ValidateGrantOption

	PebbleFormatBlockPropertyCollector

	ProbeRequest

	SelectRPCsTakeTracingInfoInband

	PreSeedTenantSpanConfigs

	SeedTenantSpanConfigs

	PublicSchemasWithDescriptors

	EnsureSpanConfigReconciliation

	EnsureSpanConfigSubscription

	EnableSpanConfigStore

	ScanWholeRows

	SCRAMAuthentication

	UnsafeLossOfQuorumRecoveryRangeLog

	AlterSystemProtectedTimestampAddColumn

	EnableProtectedTimestampsForTenant

	DeleteCommentsWithDroppedIndexes

	RemoveIncompatibleDatabasePrivileges

	AddRaftAppliedIndexTermMigration

	PostAddRaftAppliedIndexTermMigration

	DontProposeWriteTimestampForLeaseTransfers

	TenantSettingsTable

	EnablePebbleFormatVersionBlockProperties

	DisableSystemConfigGossipTrigger

	MVCCIndexBackfiller

	EnableLeaseHolderRemoval

	BackupResolutionInJob

	LooselyCoupledRaftLogTruncation

	ChangefeedIdleness

	BackupDoesNotOverwriteLatestAndCheckpoint

	EnableDeclarativeSchemaChanger

	RowLevelTTL

	PebbleFormatSplitUserKeysMarked

	IncrementalBackupSubdir

	DateStyleIntervalStyleCastRewrite

	EnableNewStoreRebalancer

	ClusterLocksVirtualTable

	AutoStatsTableSettings

	ForecastStats

	SuperRegions

	EnableNewChangefeedOptions

	SpanCountTable

	PreSeedSpanCountTable

	SeedSpanCountTable
)

const TODOPreV21_2 = V21_2

var versionsSingleton = keyedVersions{
	{

		Key:     V21_2,
		Version: roachpb.Version{Major: 21, Minor: 2},
	},

	{
		Key:     Start22_1,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 2},
	},
	{
		Key:     TargetBytesAvoidExcess,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 4},
	},
	{
		Key:     AvoidDrainingNames,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 6},
	},
	{
		Key:     DrainingNamesMigration,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 8},
	},
	{
		Key:     TraceIDDoesntImplyStructuredRecording,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 10},
	},
	{
		Key:     AlterSystemTableStatisticsAddAvgSizeCol,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 12},
	},
	{
		Key:     AlterSystemStmtDiagReqs,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 14},
	},
	{
		Key:     MVCCAddSSTable,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 16},
	},
	{
		Key:     InsertPublicSchemaNamespaceEntryOnRestore,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 18},
	},
	{
		Key:     UnsplitRangesInAsyncGCJobs,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 20},
	},
	{
		Key:     ValidateGrantOption,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 22},
	},
	{
		Key:     PebbleFormatBlockPropertyCollector,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 24},
	},
	{
		Key:     ProbeRequest,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 26},
	},
	{
		Key:     SelectRPCsTakeTracingInfoInband,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 28},
	},
	{
		Key:     PreSeedTenantSpanConfigs,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 30},
	},
	{
		Key:     SeedTenantSpanConfigs,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 32},
	},
	{
		Key:     PublicSchemasWithDescriptors,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 34},
	},
	{
		Key:     EnsureSpanConfigReconciliation,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 36},
	},
	{
		Key:     EnsureSpanConfigSubscription,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 38},
	},
	{
		Key:     EnableSpanConfigStore,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 40},
	},
	{
		Key:     ScanWholeRows,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 42},
	},
	{
		Key:     SCRAMAuthentication,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 44},
	},
	{
		Key:     UnsafeLossOfQuorumRecoveryRangeLog,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 46},
	},
	{
		Key:     AlterSystemProtectedTimestampAddColumn,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 48},
	},
	{
		Key:     EnableProtectedTimestampsForTenant,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 50},
	},
	{
		Key:     DeleteCommentsWithDroppedIndexes,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 52},
	},
	{
		Key:     RemoveIncompatibleDatabasePrivileges,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 54},
	},
	{
		Key:     AddRaftAppliedIndexTermMigration,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 56},
	},
	{
		Key:     PostAddRaftAppliedIndexTermMigration,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 58},
	},
	{
		Key:     DontProposeWriteTimestampForLeaseTransfers,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 60},
	},
	{
		Key:     TenantSettingsTable,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 62},
	},
	{
		Key:     EnablePebbleFormatVersionBlockProperties,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 64},
	},
	{
		Key:     DisableSystemConfigGossipTrigger,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 66},
	},
	{
		Key:     MVCCIndexBackfiller,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 68},
	},
	{
		Key:     EnableLeaseHolderRemoval,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 70},
	},

	{
		Key:     BackupResolutionInJob,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 76},
	},

	{
		Key:     LooselyCoupledRaftLogTruncation,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 80},
	},
	{
		Key:     ChangefeedIdleness,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 82},
	},
	{
		Key:     BackupDoesNotOverwriteLatestAndCheckpoint,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 84},
	},
	{
		Key:     EnableDeclarativeSchemaChanger,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 86},
	},
	{
		Key:     RowLevelTTL,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 88},
	},
	{
		Key:     PebbleFormatSplitUserKeysMarked,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 90},
	},
	{
		Key:     IncrementalBackupSubdir,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 92},
	},
	{
		Key:     DateStyleIntervalStyleCastRewrite,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 94},
	},
	{
		Key:     EnableNewStoreRebalancer,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 96},
	},
	{
		Key:     ClusterLocksVirtualTable,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 98},
	},
	{
		Key:     AutoStatsTableSettings,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 100},
	},
	{
		Key:     ForecastStats,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 102},
	},
	{
		Key:     SuperRegions,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 104},
	},
	{
		Key:     EnableNewChangefeedOptions,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 106},
	},
	{
		Key:     SpanCountTable,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 108},
	},
	{
		Key:     PreSeedSpanCountTable,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 110},
	},
	{
		Key:     SeedSpanCountTable,
		Version: roachpb.Version{Major: 21, Minor: 2, Internal: 112},
	},
}

var (
	binaryMinSupportedVersion = ByKey(V21_2)

	binaryVersion = versionsSingleton[len(versionsSingleton)-1].Version
)

func init() {
	const isReleaseBranch = false
	if isReleaseBranch {
		if binaryVersion != ByKey(V21_2) {
			panic("unexpected cluster version greater than release's binary version")
		}
	}
}

func ByKey(key Key) roachpb.Version {
	__antithesis_instrumentation__.Notify(37214)
	return versionsSingleton.MustByKey(key)
}

func ListBetween(from, to ClusterVersion) []ClusterVersion {
	__antithesis_instrumentation__.Notify(37215)
	return listBetweenInternal(from, to, versionsSingleton)
}

func listBetweenInternal(from, to ClusterVersion, vs keyedVersions) []ClusterVersion {
	__antithesis_instrumentation__.Notify(37216)
	var cvs []ClusterVersion
	for _, keyedV := range vs {
		__antithesis_instrumentation__.Notify(37218)

		if from.Less(keyedV.Version) && func() bool {
			__antithesis_instrumentation__.Notify(37219)
			return keyedV.Version.LessEq(to.Version) == true
		}() == true {
			__antithesis_instrumentation__.Notify(37220)
			cvs = append(cvs, ClusterVersion{Version: keyedV.Version})
		} else {
			__antithesis_instrumentation__.Notify(37221)
		}
	}
	__antithesis_instrumentation__.Notify(37217)
	return cvs
}
