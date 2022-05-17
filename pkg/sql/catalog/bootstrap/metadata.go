// Package bootstrap contains the metadata required to bootstrap the sql
// schema for a fresh cockroach cluster.
package bootstrap

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/systemschema"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/iterutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
)

type MetadataSchema struct {
	codec         keys.SQLCodec
	descs         []catalog.Descriptor
	otherSplitIDs []uint32
	otherKV       []roachpb.KeyValue
	ids           catalog.DescriptorIDSet
}

func MakeMetadataSchema(
	codec keys.SQLCodec,
	defaultZoneConfig *zonepb.ZoneConfig,
	defaultSystemZoneConfig *zonepb.ZoneConfig,
) MetadataSchema {
	__antithesis_instrumentation__.Notify(247223)
	ms := MetadataSchema{codec: codec}
	addSystemDatabaseToSchema(&ms, defaultZoneConfig, defaultSystemZoneConfig)
	return ms
}

const firstNonSystemDescriptorID = 100

func (ms *MetadataSchema) AddDescriptor(desc catalog.Descriptor) {
	__antithesis_instrumentation__.Notify(247224)
	switch id := desc.GetID(); id {
	case descpb.InvalidID:
		__antithesis_instrumentation__.Notify(247226)
		if _, isTable := desc.(catalog.TableDescriptor); !isTable {
			__antithesis_instrumentation__.Notify(247229)
			log.Fatalf(context.TODO(), "only system tables may have dynamic IDs, got %T for %s",
				desc, desc.GetName())
		} else {
			__antithesis_instrumentation__.Notify(247230)
		}
		__antithesis_instrumentation__.Notify(247227)
		mut := desc.NewBuilder().BuildCreatedMutable().(*tabledesc.Mutable)
		mut.ID = ms.allocateID()
		desc = mut.ImmutableCopy()
	default:
		__antithesis_instrumentation__.Notify(247228)
		if ms.ids.Contains(id) {
			__antithesis_instrumentation__.Notify(247231)
			log.Fatalf(context.TODO(), "adding descriptor with duplicate ID: %v", desc)
		} else {
			__antithesis_instrumentation__.Notify(247232)
		}
	}
	__antithesis_instrumentation__.Notify(247225)
	ms.descs = append(ms.descs, desc)
}

func (ms *MetadataSchema) AddDescriptorForSystemTenant(desc catalog.Descriptor) {
	__antithesis_instrumentation__.Notify(247233)
	if !ms.codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(247235)
		return
	} else {
		__antithesis_instrumentation__.Notify(247236)
	}
	__antithesis_instrumentation__.Notify(247234)
	ms.AddDescriptor(desc)
}

func (ms *MetadataSchema) AddDescriptorForNonSystemTenant(desc catalog.Descriptor) {
	__antithesis_instrumentation__.Notify(247237)
	if ms.codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(247239)
		return
	} else {
		__antithesis_instrumentation__.Notify(247240)
	}
	__antithesis_instrumentation__.Notify(247238)
	ms.AddDescriptor(desc)
}

func (ms MetadataSchema) ForEachCatalogDescriptor(fn func(desc catalog.Descriptor) error) error {
	__antithesis_instrumentation__.Notify(247241)
	for _, desc := range ms.descs {
		__antithesis_instrumentation__.Notify(247243)
		if err := fn(desc); err != nil {
			__antithesis_instrumentation__.Notify(247244)
			if iterutil.Done(err) {
				__antithesis_instrumentation__.Notify(247246)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(247247)
			}
			__antithesis_instrumentation__.Notify(247245)
			return err
		} else {
			__antithesis_instrumentation__.Notify(247248)
		}
	}
	__antithesis_instrumentation__.Notify(247242)
	return nil
}

func (ms *MetadataSchema) AddSplitIDs(id ...uint32) {
	__antithesis_instrumentation__.Notify(247249)
	ms.otherSplitIDs = append(ms.otherSplitIDs, id...)
}

func (ms MetadataSchema) SystemDescriptorCount() int {
	__antithesis_instrumentation__.Notify(247250)
	return len(ms.descs)
}

func (ms MetadataSchema) GetInitialValues() ([]roachpb.KeyValue, []roachpb.RKey) {
	__antithesis_instrumentation__.Notify(247251)
	var ret []roachpb.KeyValue
	var splits []roachpb.RKey
	add := func(key roachpb.Key, value roachpb.Value) {
		__antithesis_instrumentation__.Notify(247256)
		ret = append(ret, roachpb.KeyValue{Key: key, Value: value})
	}

	{
		__antithesis_instrumentation__.Notify(247257)
		value := roachpb.Value{}
		value.SetInt(int64(ms.FirstNonSystemDescriptorID()))
		add(ms.codec.DescIDSequenceKey(), value)
	}
	__antithesis_instrumentation__.Notify(247252)

	for _, desc := range ms.descs {
		__antithesis_instrumentation__.Notify(247258)

		nameValue := roachpb.Value{}
		nameValue.SetInt(int64(desc.GetID()))
		if desc.GetParentID() != keys.RootNamespaceID {
			__antithesis_instrumentation__.Notify(247261)
			add(catalogkeys.MakePublicObjectNameKey(ms.codec, desc.GetParentID(), desc.GetName()), nameValue)
		} else {
			__antithesis_instrumentation__.Notify(247262)

			add(catalogkeys.MakeDatabaseNameKey(ms.codec, desc.GetName()), nameValue)
			publicSchemaValue := roachpb.Value{}
			publicSchemaValue.SetInt(int64(keys.SystemPublicSchemaID))
			add(catalogkeys.MakeSchemaNameKey(ms.codec, desc.GetID(), tree.PublicSchema), publicSchemaValue)
		}
		__antithesis_instrumentation__.Notify(247259)

		descValue := roachpb.Value{}
		if err := descValue.SetProto(desc.DescriptorProto()); err != nil {
			__antithesis_instrumentation__.Notify(247263)
			log.Fatalf(context.TODO(), "could not marshal %v", desc)
		} else {
			__antithesis_instrumentation__.Notify(247264)
		}
		__antithesis_instrumentation__.Notify(247260)
		add(catalogkeys.MakeDescMetadataKey(ms.codec, desc.GetID()), descValue)
		if desc.GetID() > keys.MaxSystemConfigDescID {
			__antithesis_instrumentation__.Notify(247265)
			splits = append(splits, roachpb.RKey(ms.codec.TablePrefix(uint32(desc.GetID()))))
		} else {
			__antithesis_instrumentation__.Notify(247266)
		}
	}
	__antithesis_instrumentation__.Notify(247253)

	if ms.codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(247267)
		for _, id := range ms.otherSplitIDs {
			__antithesis_instrumentation__.Notify(247268)
			splits = append(splits, roachpb.RKey(ms.codec.TablePrefix(id)))
		}
	} else {
		__antithesis_instrumentation__.Notify(247269)
		splits = []roachpb.RKey{roachpb.RKey(ms.codec.TenantPrefix())}
	}
	__antithesis_instrumentation__.Notify(247254)

	ret = append(ret, ms.otherKV...)

	sort.Sort(roachpb.KeyValueByKey(ret))
	sort.Slice(splits, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(247270)
		return splits[i].Less(splits[j])
	})
	__antithesis_instrumentation__.Notify(247255)

	return ret, splits
}

func (ms MetadataSchema) DescriptorIDs() descpb.IDs {
	__antithesis_instrumentation__.Notify(247271)
	descriptorIDs := descpb.IDs{}
	for _, d := range ms.descs {
		__antithesis_instrumentation__.Notify(247273)
		descriptorIDs = append(descriptorIDs, d.GetID())
	}
	__antithesis_instrumentation__.Notify(247272)
	sort.Sort(descriptorIDs)
	return descriptorIDs
}

func (ms MetadataSchema) FirstNonSystemDescriptorID() descpb.ID {
	__antithesis_instrumentation__.Notify(247274)
	if next := ms.allocateID(); next > firstNonSystemDescriptorID {
		__antithesis_instrumentation__.Notify(247276)
		return next
	} else {
		__antithesis_instrumentation__.Notify(247277)
	}
	__antithesis_instrumentation__.Notify(247275)
	return firstNonSystemDescriptorID
}

func (ms MetadataSchema) allocateID() (nextID descpb.ID) {
	__antithesis_instrumentation__.Notify(247278)
	maxID := descpb.ID(keys.MaxReservedDescID)
	for _, d := range ms.descs {
		__antithesis_instrumentation__.Notify(247280)
		if d.GetID() > maxID {
			__antithesis_instrumentation__.Notify(247281)
			maxID = d.GetID()
		} else {
			__antithesis_instrumentation__.Notify(247282)
		}
	}
	__antithesis_instrumentation__.Notify(247279)
	return maxID + 1
}

func addSystemDescriptorsToSchema(target *MetadataSchema) {
	__antithesis_instrumentation__.Notify(247283)

	target.AddDescriptor(systemschema.SystemDB)

	target.AddDescriptor(systemschema.NamespaceTable)
	target.AddDescriptor(systemschema.DescriptorTable)
	target.AddDescriptor(systemschema.UsersTable)
	target.AddDescriptor(systemschema.ZonesTable)
	target.AddDescriptor(systemschema.SettingsTable)

	target.AddDescriptorForNonSystemTenant(systemschema.DescIDSequence)
	target.AddDescriptorForSystemTenant(systemschema.TenantsTable)

	target.AddDescriptor(systemschema.LeaseTable)
	target.AddDescriptor(systemschema.EventLogTable)
	target.AddDescriptor(systemschema.RangeEventTable)
	target.AddDescriptor(systemschema.UITable)
	target.AddDescriptor(systemschema.JobsTable)
	target.AddDescriptor(systemschema.WebSessionsTable)
	target.AddDescriptor(systemschema.RoleOptionsTable)

	target.AddDescriptor(systemschema.TableStatisticsTable)
	target.AddDescriptor(systemschema.LocationsTable)
	target.AddDescriptor(systemschema.RoleMembersTable)

	target.AddDescriptor(systemschema.CommentsTable)
	target.AddDescriptor(systemschema.ReportsMetaTable)
	target.AddDescriptor(systemschema.ReplicationConstraintStatsTable)
	target.AddDescriptor(systemschema.ReplicationStatsTable)
	target.AddDescriptor(systemschema.ReplicationCriticalLocalitiesTable)
	target.AddDescriptor(systemschema.ProtectedTimestampsMetaTable)
	target.AddDescriptor(systemschema.ProtectedTimestampsRecordsTable)

	target.AddDescriptor(systemschema.StatementBundleChunksTable)
	target.AddDescriptor(systemschema.StatementDiagnosticsRequestsTable)
	target.AddDescriptor(systemschema.StatementDiagnosticsTable)

	target.AddDescriptor(systemschema.ScheduledJobsTable)
	target.AddDescriptor(systemschema.SqllivenessTable)
	target.AddDescriptor(systemschema.MigrationsTable)

	target.AddDescriptor(systemschema.JoinTokensTable)

	target.AddDescriptor(systemschema.StatementStatisticsTable)
	target.AddDescriptor(systemschema.TransactionStatisticsTable)
	target.AddDescriptor(systemschema.DatabaseRoleSettingsTable)
	target.AddDescriptorForSystemTenant(systemschema.TenantUsageTable)
	target.AddDescriptor(systemschema.SQLInstancesTable)
	target.AddDescriptorForSystemTenant(systemschema.SpanConfigurationsTable)

	target.AddDescriptorForSystemTenant(systemschema.TenantSettingsTable)
	target.AddDescriptorForNonSystemTenant(systemschema.SpanCountTable)

}

func addSplitIDs(target *MetadataSchema) {
	__antithesis_instrumentation__.Notify(247284)
	target.AddSplitIDs(keys.PseudoTableIDs...)
}

func createZoneConfigKV(
	keyID int, codec keys.SQLCodec, zoneConfig *zonepb.ZoneConfig,
) roachpb.KeyValue {
	__antithesis_instrumentation__.Notify(247285)
	value := roachpb.Value{}
	if err := value.SetProto(zoneConfig); err != nil {
		__antithesis_instrumentation__.Notify(247287)
		panic(errors.NewAssertionErrorWithWrappedErrf(err, "could not marshal ZoneConfig for ID: %d", keyID))
	} else {
		__antithesis_instrumentation__.Notify(247288)
	}
	__antithesis_instrumentation__.Notify(247286)
	return roachpb.KeyValue{
		Key:   codec.ZoneKey(uint32(keyID)),
		Value: value,
	}
}

func InitialZoneConfigKVs(
	codec keys.SQLCodec,
	defaultZoneConfig *zonepb.ZoneConfig,
	defaultSystemZoneConfig *zonepb.ZoneConfig,
) (ret []roachpb.KeyValue) {
	__antithesis_instrumentation__.Notify(247289)

	ret = append(ret,
		createZoneConfigKV(keys.RootNamespaceID, codec, defaultZoneConfig))

	if !codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(247291)
		return ret
	} else {
		__antithesis_instrumentation__.Notify(247292)
	}
	__antithesis_instrumentation__.Notify(247290)

	systemZoneConf := defaultSystemZoneConfig
	metaRangeZoneConf := protoutil.Clone(defaultSystemZoneConfig).(*zonepb.ZoneConfig)
	livenessZoneConf := protoutil.Clone(defaultSystemZoneConfig).(*zonepb.ZoneConfig)

	metaRangeZoneConf.GC.TTLSeconds = 60 * 60
	ret = append(ret,
		createZoneConfigKV(keys.MetaRangesID, codec, metaRangeZoneConf))

	replicationConstraintStatsZoneConf := &zonepb.ZoneConfig{
		GC: &zonepb.GCPolicy{TTLSeconds: int32(systemschema.ReplicationConstraintStatsTableTTL.Seconds())},
	}
	replicationStatsZoneConf := &zonepb.ZoneConfig{
		GC: &zonepb.GCPolicy{TTLSeconds: int32(systemschema.ReplicationStatsTableTTL.Seconds())},
	}
	tenantUsageZoneConf := &zonepb.ZoneConfig{
		GC: &zonepb.GCPolicy{TTLSeconds: int32(systemschema.TenantUsageTableTTL.Seconds())},
	}

	livenessZoneConf.GC.TTLSeconds = 10 * 60
	ret = append(ret,
		createZoneConfigKV(keys.LivenessRangesID, codec, livenessZoneConf))
	ret = append(ret,
		createZoneConfigKV(keys.SystemRangesID, codec, systemZoneConf))
	ret = append(ret,
		createZoneConfigKV(keys.SystemDatabaseID, codec, systemZoneConf))
	ret = append(ret,
		createZoneConfigKV(keys.ReplicationConstraintStatsTableID, codec, replicationConstraintStatsZoneConf))
	ret = append(ret,
		createZoneConfigKV(keys.ReplicationStatsTableID, codec, replicationStatsZoneConf))
	ret = append(ret,
		createZoneConfigKV(keys.TenantUsageTableID, codec, tenantUsageZoneConf))

	return ret
}

func addZoneConfigKVsToSchema(
	target *MetadataSchema,
	defaultZoneConfig *zonepb.ZoneConfig,
	defaultSystemZoneConfig *zonepb.ZoneConfig,
) {
	__antithesis_instrumentation__.Notify(247293)
	kvs := InitialZoneConfigKVs(target.codec, defaultZoneConfig, defaultSystemZoneConfig)
	target.otherKV = append(target.otherKV, kvs...)
}

func addSystemDatabaseToSchema(
	target *MetadataSchema,
	defaultZoneConfig *zonepb.ZoneConfig,
	defaultSystemZoneConfig *zonepb.ZoneConfig,
) {
	__antithesis_instrumentation__.Notify(247294)
	addSystemDescriptorsToSchema(target)
	addSplitIDs(target)
	addZoneConfigKVsToSchema(target, defaultZoneConfig, defaultSystemZoneConfig)
}

func TestingMinUserDescID() uint32 {
	__antithesis_instrumentation__.Notify(247295)
	ms := MakeMetadataSchema(keys.SystemSQLCodec, zonepb.DefaultZoneConfigRef(), zonepb.DefaultSystemZoneConfigRef())
	return uint32(ms.FirstNonSystemDescriptorID())
}

func TestingMinNonDefaultUserDescID() uint32 {
	__antithesis_instrumentation__.Notify(247296)

	return TestingMinUserDescID() + uint32(len(catalogkeys.DefaultUserDBs))*2
}

func TestingUserDescID(offset uint32) uint32 {
	__antithesis_instrumentation__.Notify(247297)
	return TestingMinUserDescID() + offset
}

func TestingUserTableDataMin() roachpb.Key {
	__antithesis_instrumentation__.Notify(247298)
	return keys.SystemSQLCodec.TablePrefix(TestingUserDescID(0))
}
