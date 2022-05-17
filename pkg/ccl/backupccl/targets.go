package backupccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/ccl/backupccl/backupresolver"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/dbdesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descbuilder"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/systemschema"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

func getRelevantDescChanges(
	ctx context.Context,
	execCfg *sql.ExecutorConfig,
	startTime, endTime hlc.Timestamp,
	descriptors []catalog.Descriptor,
	expanded []descpb.ID,
	priorIDs map[descpb.ID]descpb.ID,
	fullCluster bool,
) ([]BackupManifest_DescriptorRevision, error) {
	__antithesis_instrumentation__.Notify(13325)

	allChanges, err := getAllDescChanges(ctx, execCfg.Codec, execCfg.DB, startTime, endTime, priorIDs)
	if err != nil {
		__antithesis_instrumentation__.Notify(13336)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(13337)
	}
	__antithesis_instrumentation__.Notify(13326)

	if len(allChanges) == 0 {
		__antithesis_instrumentation__.Notify(13338)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(13339)
	}
	__antithesis_instrumentation__.Notify(13327)

	var interestingChanges []BackupManifest_DescriptorRevision

	interestingIDs := make(map[descpb.ID]struct{}, len(descriptors))

	systemTableIDsToExcludeFromBackup, err := GetSystemTableIDsToExcludeFromClusterBackup(ctx, execCfg)
	if err != nil {
		__antithesis_instrumentation__.Notify(13340)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(13341)
	}
	__antithesis_instrumentation__.Notify(13328)
	isExcludedDescriptor := func(id descpb.ID) bool {
		__antithesis_instrumentation__.Notify(13342)
		if _, isOptOutSystemTable := systemTableIDsToExcludeFromBackup[id]; id == keys.SystemDatabaseID || func() bool {
			__antithesis_instrumentation__.Notify(13344)
			return isOptOutSystemTable == true
		}() == true {
			__antithesis_instrumentation__.Notify(13345)
			return true
		} else {
			__antithesis_instrumentation__.Notify(13346)
		}
		__antithesis_instrumentation__.Notify(13343)
		return false
	}
	__antithesis_instrumentation__.Notify(13329)

	isInterestingID := func(id descpb.ID) bool {
		__antithesis_instrumentation__.Notify(13347)

		if fullCluster && func() bool {
			__antithesis_instrumentation__.Notify(13350)
			return !isExcludedDescriptor(id) == true
		}() == true {
			__antithesis_instrumentation__.Notify(13351)
			return true
		} else {
			__antithesis_instrumentation__.Notify(13352)
		}
		__antithesis_instrumentation__.Notify(13348)

		if _, ok := interestingIDs[id]; ok {
			__antithesis_instrumentation__.Notify(13353)
			return true
		} else {
			__antithesis_instrumentation__.Notify(13354)
		}
		__antithesis_instrumentation__.Notify(13349)
		return false
	}
	__antithesis_instrumentation__.Notify(13330)

	for _, i := range descriptors {
		__antithesis_instrumentation__.Notify(13355)
		interestingIDs[i.GetID()] = struct{}{}
		if table, isTable := i.(catalog.TableDescriptor); isTable {
			__antithesis_instrumentation__.Notify(13356)

			for j := table.GetReplacementOf().ID; j != descpb.InvalidID; j = priorIDs[j] {
				__antithesis_instrumentation__.Notify(13357)
				interestingIDs[j] = struct{}{}
			}
		} else {
			__antithesis_instrumentation__.Notify(13358)
		}
	}
	__antithesis_instrumentation__.Notify(13331)

	interestingParents := make(map[descpb.ID]struct{}, len(expanded))
	for _, i := range expanded {
		__antithesis_instrumentation__.Notify(13359)
		interestingParents[i] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(13332)

	if !startTime.IsEmpty() {
		__antithesis_instrumentation__.Notify(13360)
		starting, err := backupresolver.LoadAllDescs(ctx, execCfg, startTime)
		if err != nil {
			__antithesis_instrumentation__.Notify(13362)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(13363)
		}
		__antithesis_instrumentation__.Notify(13361)
		for _, i := range starting {
			__antithesis_instrumentation__.Notify(13364)
			switch desc := i.(type) {
			case catalog.TableDescriptor, catalog.TypeDescriptor, catalog.SchemaDescriptor:
				__antithesis_instrumentation__.Notify(13366)

				if _, ok := interestingParents[desc.GetParentID()]; ok {
					__antithesis_instrumentation__.Notify(13367)
					interestingIDs[desc.GetID()] = struct{}{}
				} else {
					__antithesis_instrumentation__.Notify(13368)
				}
			}
			__antithesis_instrumentation__.Notify(13365)
			if isInterestingID(i.GetID()) {
				__antithesis_instrumentation__.Notify(13369)
				desc := i

				initial := BackupManifest_DescriptorRevision{Time: startTime, ID: i.GetID(), Desc: desc.DescriptorProto()}
				interestingChanges = append(interestingChanges, initial)
			} else {
				__antithesis_instrumentation__.Notify(13370)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(13371)
	}
	__antithesis_instrumentation__.Notify(13333)

	for _, change := range allChanges {
		__antithesis_instrumentation__.Notify(13372)

		if isInterestingID(change.ID) {
			__antithesis_instrumentation__.Notify(13373)
			interestingChanges = append(interestingChanges, change)
		} else {
			__antithesis_instrumentation__.Notify(13374)
			if change.Desc != nil {
				__antithesis_instrumentation__.Notify(13375)
				desc := descbuilder.NewBuilder(change.Desc).BuildExistingMutable()
				switch desc := desc.(type) {
				case catalog.TableDescriptor, catalog.TypeDescriptor, catalog.SchemaDescriptor:
					__antithesis_instrumentation__.Notify(13376)
					if _, ok := interestingParents[desc.GetParentID()]; ok {
						__antithesis_instrumentation__.Notify(13377)
						interestingIDs[desc.GetID()] = struct{}{}
						interestingChanges = append(interestingChanges, change)
					} else {
						__antithesis_instrumentation__.Notify(13378)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(13379)
			}
		}
	}
	__antithesis_instrumentation__.Notify(13334)

	sort.Slice(interestingChanges, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(13380)
		return interestingChanges[i].Time.Less(interestingChanges[j].Time)
	})
	__antithesis_instrumentation__.Notify(13335)

	return interestingChanges, nil
}

func getAllDescChanges(
	ctx context.Context,
	codec keys.SQLCodec,
	db *kv.DB,
	startTime, endTime hlc.Timestamp,
	priorIDs map[descpb.ID]descpb.ID,
) ([]BackupManifest_DescriptorRevision, error) {
	__antithesis_instrumentation__.Notify(13381)
	startKey := codec.TablePrefix(keys.DescriptorTableID)
	endKey := startKey.PrefixEnd()

	allRevs, err := kvclient.GetAllRevisions(ctx, db, startKey, endKey, startTime, endTime)
	if err != nil {
		__antithesis_instrumentation__.Notify(13384)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(13385)
	}
	__antithesis_instrumentation__.Notify(13382)

	var res []BackupManifest_DescriptorRevision

	for _, revs := range allRevs {
		__antithesis_instrumentation__.Notify(13386)
		id, err := codec.DecodeDescMetadataID(revs.Key)
		if err != nil {
			__antithesis_instrumentation__.Notify(13388)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(13389)
		}
		__antithesis_instrumentation__.Notify(13387)
		for _, rev := range revs.Values {
			__antithesis_instrumentation__.Notify(13390)
			r := BackupManifest_DescriptorRevision{ID: descpb.ID(id), Time: rev.Timestamp}
			if len(rev.RawBytes) != 0 {
				__antithesis_instrumentation__.Notify(13392)
				var desc descpb.Descriptor
				if err := rev.GetProto(&desc); err != nil {
					__antithesis_instrumentation__.Notify(13394)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(13395)
				}
				__antithesis_instrumentation__.Notify(13393)
				r.Desc = &desc

				t, _, _, _ := descpb.FromDescriptorWithMVCCTimestamp(r.Desc, rev.Timestamp)
				if priorIDs != nil && func() bool {
					__antithesis_instrumentation__.Notify(13396)
					return t != nil == true
				}() == true && func() bool {
					__antithesis_instrumentation__.Notify(13397)
					return t.ReplacementOf.ID != descpb.InvalidID == true
				}() == true {
					__antithesis_instrumentation__.Notify(13398)
					priorIDs[t.ID] = t.ReplacementOf.ID
				} else {
					__antithesis_instrumentation__.Notify(13399)
				}
			} else {
				__antithesis_instrumentation__.Notify(13400)
			}
			__antithesis_instrumentation__.Notify(13391)
			res = append(res, r)
		}
	}
	__antithesis_instrumentation__.Notify(13383)
	return res, nil
}

func fullClusterTargets(
	allDescs []catalog.Descriptor,
) ([]catalog.Descriptor, []catalog.DatabaseDescriptor, error) {
	__antithesis_instrumentation__.Notify(13401)
	fullClusterDescs := make([]catalog.Descriptor, 0, len(allDescs))
	fullClusterDBs := make([]catalog.DatabaseDescriptor, 0)

	systemTablesToBackup := GetSystemTablesToIncludeInClusterBackup()

	for _, desc := range allDescs {
		__antithesis_instrumentation__.Notify(13403)

		if desc.Dropped() {
			__antithesis_instrumentation__.Notify(13405)
			continue
		} else {
			__antithesis_instrumentation__.Notify(13406)
		}
		__antithesis_instrumentation__.Notify(13404)
		switch desc := desc.(type) {
		case catalog.DatabaseDescriptor:
			__antithesis_instrumentation__.Notify(13407)
			dbDesc := dbdesc.NewBuilder(desc.DatabaseDesc()).BuildImmutableDatabase()
			fullClusterDescs = append(fullClusterDescs, desc)
			if dbDesc.GetID() != systemschema.SystemDB.GetID() {
				__antithesis_instrumentation__.Notify(13411)

				fullClusterDBs = append(fullClusterDBs, dbDesc)
			} else {
				__antithesis_instrumentation__.Notify(13412)
			}
		case catalog.TableDescriptor:
			__antithesis_instrumentation__.Notify(13408)
			if desc.GetParentID() == keys.SystemDatabaseID {
				__antithesis_instrumentation__.Notify(13413)

				if _, ok := systemTablesToBackup[desc.GetName()]; ok {
					__antithesis_instrumentation__.Notify(13414)
					fullClusterDescs = append(fullClusterDescs, desc)
				} else {
					__antithesis_instrumentation__.Notify(13415)
				}
			} else {
				__antithesis_instrumentation__.Notify(13416)
				fullClusterDescs = append(fullClusterDescs, desc)
			}
		case catalog.SchemaDescriptor:
			__antithesis_instrumentation__.Notify(13409)
			fullClusterDescs = append(fullClusterDescs, desc)
		case catalog.TypeDescriptor:
			__antithesis_instrumentation__.Notify(13410)
			fullClusterDescs = append(fullClusterDescs, desc)
		}
	}
	__antithesis_instrumentation__.Notify(13402)
	return fullClusterDescs, fullClusterDBs, nil
}

func fullClusterTargetsRestore(
	allDescs []catalog.Descriptor, lastBackupManifest BackupManifest,
) ([]catalog.Descriptor, []catalog.DatabaseDescriptor, []descpb.TenantInfoWithUsage, error) {
	__antithesis_instrumentation__.Notify(13417)
	fullClusterDescs, fullClusterDBs, err := fullClusterTargets(allDescs)
	var filteredDescs []catalog.Descriptor
	var filteredDBs []catalog.DatabaseDescriptor
	for _, desc := range fullClusterDescs {
		__antithesis_instrumentation__.Notify(13421)
		if desc.GetID() != keys.SystemDatabaseID {
			__antithesis_instrumentation__.Notify(13422)
			filteredDescs = append(filteredDescs, desc)
		} else {
			__antithesis_instrumentation__.Notify(13423)
		}
	}
	__antithesis_instrumentation__.Notify(13418)
	for _, desc := range fullClusterDBs {
		__antithesis_instrumentation__.Notify(13424)
		if desc.GetID() != keys.SystemDatabaseID {
			__antithesis_instrumentation__.Notify(13425)
			filteredDBs = append(filteredDBs, desc)
		} else {
			__antithesis_instrumentation__.Notify(13426)
		}
	}
	__antithesis_instrumentation__.Notify(13419)
	if err != nil {
		__antithesis_instrumentation__.Notify(13427)
		return nil, nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(13428)
	}
	__antithesis_instrumentation__.Notify(13420)
	return filteredDescs, filteredDBs, lastBackupManifest.GetTenants(), nil
}

func fullClusterTargetsBackup(
	ctx context.Context, execCfg *sql.ExecutorConfig, endTime hlc.Timestamp,
) ([]catalog.Descriptor, []descpb.ID, error) {
	__antithesis_instrumentation__.Notify(13429)
	allDescs, err := backupresolver.LoadAllDescs(ctx, execCfg, endTime)
	if err != nil {
		__antithesis_instrumentation__.Notify(13434)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(13435)
	}
	__antithesis_instrumentation__.Notify(13430)

	fullClusterDescs, fullClusterDBs, err := fullClusterTargets(allDescs)
	if err != nil {
		__antithesis_instrumentation__.Notify(13436)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(13437)
	}
	__antithesis_instrumentation__.Notify(13431)

	fullClusterDBIDs := make([]descpb.ID, 0)
	for _, desc := range fullClusterDBs {
		__antithesis_instrumentation__.Notify(13438)
		fullClusterDBIDs = append(fullClusterDBIDs, desc.GetID())
	}
	__antithesis_instrumentation__.Notify(13432)
	if len(fullClusterDescs) == 0 {
		__antithesis_instrumentation__.Notify(13439)
		return nil, nil, errors.New("no descriptors available to backup at selected time")
	} else {
		__antithesis_instrumentation__.Notify(13440)
	}
	__antithesis_instrumentation__.Notify(13433)
	return fullClusterDescs, fullClusterDBIDs, nil
}

func selectTargets(
	ctx context.Context,
	p sql.PlanHookState,
	backupManifests []BackupManifest,
	targets tree.TargetList,
	descriptorCoverage tree.DescriptorCoverage,
	asOf hlc.Timestamp,
	restoreSystemUsers bool,
) ([]catalog.Descriptor, []catalog.DatabaseDescriptor, []descpb.TenantInfoWithUsage, error) {
	__antithesis_instrumentation__.Notify(13441)
	allDescs, lastBackupManifest := loadSQLDescsFromBackupsAtTime(backupManifests, asOf)

	if descriptorCoverage == tree.AllDescriptors {
		__antithesis_instrumentation__.Notify(13448)
		return fullClusterTargetsRestore(allDescs, lastBackupManifest)
	} else {
		__antithesis_instrumentation__.Notify(13449)
	}
	__antithesis_instrumentation__.Notify(13442)

	if restoreSystemUsers {
		__antithesis_instrumentation__.Notify(13450)
		systemTables := make([]catalog.Descriptor, 0)
		var users catalog.Descriptor
		for _, desc := range allDescs {
			__antithesis_instrumentation__.Notify(13453)
			if desc.GetParentID() == systemschema.SystemDB.GetID() {
				__antithesis_instrumentation__.Notify(13454)
				switch desc.GetName() {
				case systemschema.UsersTable.GetName():
					__antithesis_instrumentation__.Notify(13455)
					users = desc
					systemTables = append(systemTables, desc)
				case systemschema.RoleMembersTable.GetName():
					__antithesis_instrumentation__.Notify(13456)
					systemTables = append(systemTables, desc)
				default:
					__antithesis_instrumentation__.Notify(13457)

				}
			} else {
				__antithesis_instrumentation__.Notify(13458)
			}
		}
		__antithesis_instrumentation__.Notify(13451)
		if users == nil {
			__antithesis_instrumentation__.Notify(13459)
			return nil, nil, nil, errors.Errorf("cannot restore system users as no system.users table in the backup")
		} else {
			__antithesis_instrumentation__.Notify(13460)
		}
		__antithesis_instrumentation__.Notify(13452)
		return systemTables, nil, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(13461)
	}
	__antithesis_instrumentation__.Notify(13443)

	if targets.TenantID.IsSet() {
		__antithesis_instrumentation__.Notify(13462)
		for _, tenant := range lastBackupManifest.GetTenants() {
			__antithesis_instrumentation__.Notify(13464)

			if tenant.ID == targets.TenantID.ToUint64() {
				__antithesis_instrumentation__.Notify(13465)
				return nil, nil, []descpb.TenantInfoWithUsage{tenant}, nil
			} else {
				__antithesis_instrumentation__.Notify(13466)
			}
		}
		__antithesis_instrumentation__.Notify(13463)
		return nil, nil, nil, errors.Errorf("tenant %d not in backup", targets.TenantID.ToUint64())
	} else {
		__antithesis_instrumentation__.Notify(13467)
	}
	__antithesis_instrumentation__.Notify(13444)

	matched, err := backupresolver.DescriptorsMatchingTargets(ctx,
		p.CurrentDatabase(), p.CurrentSearchPath(), allDescs, targets, asOf)
	if err != nil {
		__antithesis_instrumentation__.Notify(13468)
		return nil, nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(13469)
	}
	__antithesis_instrumentation__.Notify(13445)

	if len(matched.Descs) == 0 {
		__antithesis_instrumentation__.Notify(13470)
		return nil, nil, nil, errors.Errorf("no tables or databases matched the given targets: %s", tree.ErrString(&targets))
	} else {
		__antithesis_instrumentation__.Notify(13471)
	}
	__antithesis_instrumentation__.Notify(13446)

	if lastBackupManifest.FormatVersion >= BackupFormatDescriptorTrackingVersion {
		__antithesis_instrumentation__.Notify(13472)
		if err := matched.CheckExpansions(lastBackupManifest.CompleteDbs); err != nil {
			__antithesis_instrumentation__.Notify(13473)
			return nil, nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(13474)
		}
	} else {
		__antithesis_instrumentation__.Notify(13475)
	}
	__antithesis_instrumentation__.Notify(13447)

	return matched.Descs, matched.RequestedDBs, nil, nil
}

type EntryFiles []execinfrapb.RestoreFileSpec

type BackupTableEntry struct {
	Desc                 catalog.TableDescriptor
	Span                 roachpb.Span
	Files                []EntryFiles
	LastSchemaChangeTime hlc.Timestamp
}

func MakeBackupTableEntry(
	ctx context.Context,
	fullyQualifiedTableName string,
	backupManifests []BackupManifest,
	endTime hlc.Timestamp,
	user security.SQLUsername,
	backupCodec keys.SQLCodec,
) (BackupTableEntry, error) {
	__antithesis_instrumentation__.Notify(13476)
	var descName []string
	if descName = strings.Split(fullyQualifiedTableName, "."); len(descName) != 3 {
		__antithesis_instrumentation__.Notify(13486)
		return BackupTableEntry{}, errors.Newf("table name should be specified in format databaseName.schemaName.tableName")
	} else {
		__antithesis_instrumentation__.Notify(13487)
	}
	__antithesis_instrumentation__.Notify(13477)

	if !endTime.IsEmpty() {
		__antithesis_instrumentation__.Notify(13488)
		ind := -1
		for i, b := range backupManifests {
			__antithesis_instrumentation__.Notify(13491)
			if b.StartTime.Less(endTime) && func() bool {
				__antithesis_instrumentation__.Notify(13492)
				return endTime.LessEq(b.EndTime) == true
			}() == true {
				__antithesis_instrumentation__.Notify(13493)
				if endTime != b.EndTime && func() bool {
					__antithesis_instrumentation__.Notify(13495)
					return b.MVCCFilter != MVCCFilter_All == true
				}() == true {
					__antithesis_instrumentation__.Notify(13496)
					errorHints := "reading data for requested time requires that BACKUP was created with %q" +
						" or should specify the time to be an exact backup time, nearest backup time is %s"
					return BackupTableEntry{}, errors.WithHintf(
						errors.Newf("unknown read time: %s", timeutil.Unix(0, endTime.WallTime).UTC()),
						errorHints, backupOptRevisionHistory, timeutil.Unix(0, b.EndTime.WallTime).UTC(),
					)
				} else {
					__antithesis_instrumentation__.Notify(13497)
				}
				__antithesis_instrumentation__.Notify(13494)
				ind = i
				break
			} else {
				__antithesis_instrumentation__.Notify(13498)
			}
		}
		__antithesis_instrumentation__.Notify(13489)
		if ind == -1 {
			__antithesis_instrumentation__.Notify(13499)
			return BackupTableEntry{}, errors.Newf("supplied backups do not cover requested time %s", timeutil.Unix(0, endTime.WallTime).UTC())
		} else {
			__antithesis_instrumentation__.Notify(13500)
		}
		__antithesis_instrumentation__.Notify(13490)
		backupManifests = backupManifests[:ind+1]
	} else {
		__antithesis_instrumentation__.Notify(13501)
	}
	__antithesis_instrumentation__.Notify(13478)

	allDescs, _ := loadSQLDescsFromBackupsAtTime(backupManifests, endTime)
	resolver, err := backupresolver.NewDescriptorResolver(allDescs)
	if err != nil {
		__antithesis_instrumentation__.Notify(13502)
		return BackupTableEntry{}, errors.Wrapf(err, "creating a new resolver for all descriptors")
	} else {
		__antithesis_instrumentation__.Notify(13503)
	}
	__antithesis_instrumentation__.Notify(13479)

	found, _, desc, err := resolver.LookupObject(ctx, tree.ObjectLookupFlags{}, descName[0], descName[1], descName[2])
	if err != nil {
		__antithesis_instrumentation__.Notify(13504)
		return BackupTableEntry{}, errors.Wrapf(err, "looking up table %s", fullyQualifiedTableName)
	} else {
		__antithesis_instrumentation__.Notify(13505)
	}
	__antithesis_instrumentation__.Notify(13480)
	if !found {
		__antithesis_instrumentation__.Notify(13506)
		return BackupTableEntry{}, errors.Newf("table %s not found", fullyQualifiedTableName)
	} else {
		__antithesis_instrumentation__.Notify(13507)
	}
	__antithesis_instrumentation__.Notify(13481)
	tbMutable, ok := desc.(*tabledesc.Mutable)
	if !ok {
		__antithesis_instrumentation__.Notify(13508)
		return BackupTableEntry{}, errors.Newf("object %s not mutable", fullyQualifiedTableName)
	} else {
		__antithesis_instrumentation__.Notify(13509)
	}
	__antithesis_instrumentation__.Notify(13482)
	tbDesc, err := catalog.AsTableDescriptor(tbMutable)
	if err != nil {
		__antithesis_instrumentation__.Notify(13510)
		return BackupTableEntry{}, errors.Wrapf(err, "fetching table %s descriptor", fullyQualifiedTableName)
	} else {
		__antithesis_instrumentation__.Notify(13511)
	}
	__antithesis_instrumentation__.Notify(13483)

	tablePrimaryIndexSpan := tbDesc.PrimaryIndexSpan(backupCodec)

	if err := checkCoverage(ctx, []roachpb.Span{tablePrimaryIndexSpan}, backupManifests); err != nil {
		__antithesis_instrumentation__.Notify(13512)
		return BackupTableEntry{}, errors.Wrapf(err, "making spans for table %s", fullyQualifiedTableName)
	} else {
		__antithesis_instrumentation__.Notify(13513)
	}
	__antithesis_instrumentation__.Notify(13484)

	entry := makeSimpleImportSpans(
		[]roachpb.Span{tablePrimaryIndexSpan},
		backupManifests,
		nil,
		roachpb.Key{},
	)

	lastSchemaChangeTime := findLastSchemaChangeTime(backupManifests, tbDesc, endTime)

	backupTableEntry := BackupTableEntry{
		tbDesc,
		tablePrimaryIndexSpan,
		make([]EntryFiles, 0),
		lastSchemaChangeTime,
	}

	for _, e := range entry {
		__antithesis_instrumentation__.Notify(13514)
		backupTableEntry.Files = append(backupTableEntry.Files, e.Files)
	}
	__antithesis_instrumentation__.Notify(13485)

	return backupTableEntry, nil
}

func findLastSchemaChangeTime(
	backupManifests []BackupManifest, tbDesc catalog.TableDescriptor, endTime hlc.Timestamp,
) hlc.Timestamp {
	__antithesis_instrumentation__.Notify(13515)
	lastSchemaChangeTime := endTime
	for i := len(backupManifests) - 1; i >= 0; i-- {
		__antithesis_instrumentation__.Notify(13517)
		manifest := backupManifests[i]
		for j := len(manifest.DescriptorChanges) - 1; j >= 0; j-- {
			__antithesis_instrumentation__.Notify(13518)
			rev := manifest.DescriptorChanges[j]

			if endTime.LessEq(rev.Time) {
				__antithesis_instrumentation__.Notify(13520)
				continue
			} else {
				__antithesis_instrumentation__.Notify(13521)
			}
			__antithesis_instrumentation__.Notify(13519)

			if rev.ID == tbDesc.GetID() {
				__antithesis_instrumentation__.Notify(13522)
				d := descbuilder.NewBuilder(rev.Desc).BuildExistingMutable()
				revDesc, _ := catalog.AsTableDescriptor(d)
				if !reflect.DeepEqual(revDesc.PublicColumns(), tbDesc.PublicColumns()) {
					__antithesis_instrumentation__.Notify(13524)
					return lastSchemaChangeTime
				} else {
					__antithesis_instrumentation__.Notify(13525)
				}
				__antithesis_instrumentation__.Notify(13523)
				lastSchemaChangeTime = rev.Time
			} else {
				__antithesis_instrumentation__.Notify(13526)
			}
		}
	}
	__antithesis_instrumentation__.Notify(13516)
	return lastSchemaChangeTime
}

func checkMultiRegionCompatible(
	ctx context.Context,
	txn *kv.Txn,
	col *descs.Collection,
	table *tabledesc.Mutable,
	database catalog.DatabaseDescriptor,
) error {
	__antithesis_instrumentation__.Notify(13527)

	if !database.IsMultiRegion() && func() bool {
		__antithesis_instrumentation__.Notify(13534)
		return table.GetLocalityConfig() == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(13535)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(13536)
	}
	__antithesis_instrumentation__.Notify(13528)

	if database.IsMultiRegion() && func() bool {
		__antithesis_instrumentation__.Notify(13537)
		return table.GetLocalityConfig() == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(13538)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(13539)
	}
	__antithesis_instrumentation__.Notify(13529)

	if !database.IsMultiRegion() && func() bool {
		__antithesis_instrumentation__.Notify(13540)
		return table.GetLocalityConfig() != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(13541)
		return pgerror.Newf(pgcode.FeatureNotSupported,
			"cannot restore descriptor for multi-region table %s into non-multi-region database %s",
			table.GetName(),
			database.GetName(),
		)
	} else {
		__antithesis_instrumentation__.Notify(13542)
	}
	__antithesis_instrumentation__.Notify(13530)

	if table.IsLocalityGlobal() {
		__antithesis_instrumentation__.Notify(13543)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(13544)
	}
	__antithesis_instrumentation__.Notify(13531)

	if table.IsLocalityRegionalByTable() {
		__antithesis_instrumentation__.Notify(13545)
		regionName, _ := table.GetRegionalByTableRegion()
		if regionName == catpb.RegionName(tree.PrimaryRegionNotSpecifiedName) {
			__antithesis_instrumentation__.Notify(13550)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(13551)
		}
		__antithesis_instrumentation__.Notify(13546)

		regionEnumID := database.GetRegionConfig().RegionEnumID
		regionEnum, err := col.Direct().MustGetTypeDescByID(ctx, txn, regionEnumID)
		if err != nil {
			__antithesis_instrumentation__.Notify(13552)
			return err
		} else {
			__antithesis_instrumentation__.Notify(13553)
		}
		__antithesis_instrumentation__.Notify(13547)
		dbRegionNames, err := regionEnum.RegionNames()
		if err != nil {
			__antithesis_instrumentation__.Notify(13554)
			return err
		} else {
			__antithesis_instrumentation__.Notify(13555)
		}
		__antithesis_instrumentation__.Notify(13548)
		existingRegions := make([]string, len(dbRegionNames))
		for i, dbRegionName := range dbRegionNames {
			__antithesis_instrumentation__.Notify(13556)
			if dbRegionName == regionName {
				__antithesis_instrumentation__.Notify(13558)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(13559)
			}
			__antithesis_instrumentation__.Notify(13557)
			existingRegions[i] = fmt.Sprintf("%q", dbRegionName)
		}
		__antithesis_instrumentation__.Notify(13549)

		return errors.Newf(
			"cannot restore REGIONAL BY TABLE %s IN REGION %q (table ID: %d) into database %q; region %q not found in database regions %s",
			table.GetName(), regionName, table.GetID(),
			database.GetName(), regionName, strings.Join(existingRegions, ", "),
		)
	} else {
		__antithesis_instrumentation__.Notify(13560)
	}
	__antithesis_instrumentation__.Notify(13532)

	if table.IsLocalityRegionalByRow() {
		__antithesis_instrumentation__.Notify(13561)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(13562)
	}
	__antithesis_instrumentation__.Notify(13533)

	return errors.AssertionFailedf(
		"locality config of table %s (ID: %d) has locality %v which is unknown by RESTORE",
		table.GetName(), table.GetID(), table.GetLocalityConfig(),
	)
}
