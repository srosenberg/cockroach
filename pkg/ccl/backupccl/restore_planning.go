package backupccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/ccl/multiregionccl"
	"github.com/cockroachdb/cockroach/pkg/ccl/storageccl"
	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/featureflag"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/dbdesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descbuilder"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descidgen"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/multiregion"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/rewrite"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemadesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/systemschema"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgnotice"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/roleoption"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
)

const (
	restoreOptIntoDB                    = "into_db"
	restoreOptSkipMissingFKs            = "skip_missing_foreign_keys"
	restoreOptSkipMissingSequences      = "skip_missing_sequences"
	restoreOptSkipMissingSequenceOwners = "skip_missing_sequence_owners"
	restoreOptSkipMissingViews          = "skip_missing_views"
	restoreOptSkipLocalitiesCheck       = "skip_localities_check"
	restoreOptDebugPauseOn              = "debug_pause_on"
	restoreOptAsTenant                  = "tenant"

	restoreTempSystemDB = "crdb_temp_system"
)

var allowedDebugPauseOnValues = map[string]struct{}{
	"error": {},
}

var featureRestoreEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"feature.restore.enabled",
	"set to true to enable restore, false to disable; default is true",
	featureflag.FeatureFlagEnabledDefault,
).WithPublic()

func maybeFilterMissingViews(
	tablesByID map[descpb.ID]*tabledesc.Mutable,
	typesByID map[descpb.ID]*typedesc.Mutable,
	skipMissingViews bool,
) (map[descpb.ID]*tabledesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(11468)

	var hasValidViewDependencies func(desc *tabledesc.Mutable) bool
	hasValidViewDependencies = func(desc *tabledesc.Mutable) bool {
		__antithesis_instrumentation__.Notify(11471)
		if !desc.IsView() {
			__antithesis_instrumentation__.Notify(11475)
			return true
		} else {
			__antithesis_instrumentation__.Notify(11476)
		}
		__antithesis_instrumentation__.Notify(11472)
		for _, id := range desc.DependsOn {
			__antithesis_instrumentation__.Notify(11477)
			if depDesc, ok := tablesByID[id]; !ok || func() bool {
				__antithesis_instrumentation__.Notify(11478)
				return !hasValidViewDependencies(depDesc) == true
			}() == true {
				__antithesis_instrumentation__.Notify(11479)
				return false
			} else {
				__antithesis_instrumentation__.Notify(11480)
			}
		}
		__antithesis_instrumentation__.Notify(11473)
		for _, id := range desc.DependsOnTypes {
			__antithesis_instrumentation__.Notify(11481)
			if _, ok := typesByID[id]; !ok {
				__antithesis_instrumentation__.Notify(11482)
				return false
			} else {
				__antithesis_instrumentation__.Notify(11483)
			}
		}
		__antithesis_instrumentation__.Notify(11474)
		return true
	}
	__antithesis_instrumentation__.Notify(11469)

	filteredTablesByID := make(map[descpb.ID]*tabledesc.Mutable)
	for id, table := range tablesByID {
		__antithesis_instrumentation__.Notify(11484)
		if hasValidViewDependencies(table) {
			__antithesis_instrumentation__.Notify(11485)
			filteredTablesByID[id] = table
		} else {
			__antithesis_instrumentation__.Notify(11486)
			if !skipMissingViews {
				__antithesis_instrumentation__.Notify(11487)
				return nil, errors.Errorf(
					"cannot restore view %q without restoring referenced table (or %q option)",
					table.Name, restoreOptSkipMissingViews,
				)
			} else {
				__antithesis_instrumentation__.Notify(11488)
			}
		}
	}
	__antithesis_instrumentation__.Notify(11470)
	return filteredTablesByID, nil
}

func synthesizePGTempSchema(
	ctx context.Context, p sql.PlanHookState, schemaName string, dbID descpb.ID,
) (descpb.ID, error) {
	__antithesis_instrumentation__.Notify(11489)
	var synthesizedSchemaID descpb.ID
	err := sql.DescsTxn(ctx, p.ExecCfg(), func(ctx context.Context, txn *kv.Txn, col *descs.Collection) error {
		__antithesis_instrumentation__.Notify(11491)
		var err error
		schemaID, err := col.Direct().LookupSchemaID(ctx, txn, dbID, schemaName)
		if err != nil {
			__antithesis_instrumentation__.Notify(11495)
			return err
		} else {
			__antithesis_instrumentation__.Notify(11496)
		}
		__antithesis_instrumentation__.Notify(11492)
		if schemaID != descpb.InvalidID {
			__antithesis_instrumentation__.Notify(11497)
			return errors.Newf("attempted to synthesize temp schema during RESTORE but found"+
				" another schema already using the same schema key %s", schemaName)
		} else {
			__antithesis_instrumentation__.Notify(11498)
		}
		__antithesis_instrumentation__.Notify(11493)
		synthesizedSchemaID, err = descidgen.GenerateUniqueDescID(ctx, p.ExecCfg().DB, p.ExecCfg().Codec)
		if err != nil {
			__antithesis_instrumentation__.Notify(11499)
			return err
		} else {
			__antithesis_instrumentation__.Notify(11500)
		}
		__antithesis_instrumentation__.Notify(11494)
		return p.CreateSchemaNamespaceEntry(ctx, catalogkeys.MakeSchemaNameKey(p.ExecCfg().Codec, dbID, schemaName), synthesizedSchemaID)
	})
	__antithesis_instrumentation__.Notify(11490)

	return synthesizedSchemaID, err
}

func allocateDescriptorRewrites(
	ctx context.Context,
	p sql.PlanHookState,
	databasesByID map[descpb.ID]*dbdesc.Mutable,
	schemasByID map[descpb.ID]*schemadesc.Mutable,
	tablesByID map[descpb.ID]*tabledesc.Mutable,
	typesByID map[descpb.ID]*typedesc.Mutable,
	restoreDBs []catalog.DatabaseDescriptor,
	descriptorCoverage tree.DescriptorCoverage,
	opts tree.RestoreOptions,
	intoDB string,
	newDBName string,
	restoreSystemUsers bool,
) (jobspb.DescRewriteMap, error) {
	__antithesis_instrumentation__.Notify(11501)
	descriptorRewrites := make(jobspb.DescRewriteMap)

	restoreDBNames := make(map[string]catalog.DatabaseDescriptor, len(restoreDBs))
	for _, db := range restoreDBs {
		__antithesis_instrumentation__.Notify(11518)
		restoreDBNames[db.GetName()] = db
	}
	__antithesis_instrumentation__.Notify(11502)

	if len(restoreDBNames) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(11519)
		return intoDB != "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(11520)
		return nil, errors.Errorf("cannot use %q option when restoring database(s)", restoreOptIntoDB)
	} else {
		__antithesis_instrumentation__.Notify(11521)
	}
	__antithesis_instrumentation__.Notify(11503)

	var maxDescIDInBackup int64
	for _, table := range tablesByID {
		__antithesis_instrumentation__.Notify(11522)
		if int64(table.ID) > maxDescIDInBackup {
			__antithesis_instrumentation__.Notify(11526)
			maxDescIDInBackup = int64(table.ID)
		} else {
			__antithesis_instrumentation__.Notify(11527)
		}
		__antithesis_instrumentation__.Notify(11523)

		for i := range table.OutboundFKs {
			__antithesis_instrumentation__.Notify(11528)
			fk := &table.OutboundFKs[i]
			if _, ok := tablesByID[fk.ReferencedTableID]; !ok {
				__antithesis_instrumentation__.Notify(11529)
				if !opts.SkipMissingFKs {
					__antithesis_instrumentation__.Notify(11530)
					return nil, errors.Errorf(
						"cannot restore table %q without referenced table %d (or %q option)",
						table.Name, fk.ReferencedTableID, restoreOptSkipMissingFKs,
					)
				} else {
					__antithesis_instrumentation__.Notify(11531)
				}
			} else {
				__antithesis_instrumentation__.Notify(11532)
			}
		}
		__antithesis_instrumentation__.Notify(11524)

		for i := range table.Columns {
			__antithesis_instrumentation__.Notify(11533)
			col := &table.Columns[i]

			if col.Type.UserDefined() {
				__antithesis_instrumentation__.Notify(11536)

				id, err := typedesc.GetUserDefinedTypeDescID(col.Type)
				if err != nil {
					__antithesis_instrumentation__.Notify(11538)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(11539)
				}
				__antithesis_instrumentation__.Notify(11537)
				if _, ok := typesByID[id]; !ok {
					__antithesis_instrumentation__.Notify(11540)
					return nil, errors.Errorf(
						"cannot restore table %q without referenced type %d",
						table.Name,
						id,
					)
				} else {
					__antithesis_instrumentation__.Notify(11541)
				}
			} else {
				__antithesis_instrumentation__.Notify(11542)
			}
			__antithesis_instrumentation__.Notify(11534)
			for _, seqID := range col.UsesSequenceIds {
				__antithesis_instrumentation__.Notify(11543)
				if _, ok := tablesByID[seqID]; !ok {
					__antithesis_instrumentation__.Notify(11544)
					if !opts.SkipMissingSequences {
						__antithesis_instrumentation__.Notify(11545)
						return nil, errors.Errorf(
							"cannot restore table %q without referenced sequence %d (or %q option)",
							table.Name, seqID, restoreOptSkipMissingSequences,
						)
					} else {
						__antithesis_instrumentation__.Notify(11546)
					}
				} else {
					__antithesis_instrumentation__.Notify(11547)
				}
			}
			__antithesis_instrumentation__.Notify(11535)
			for _, seqID := range col.OwnsSequenceIds {
				__antithesis_instrumentation__.Notify(11548)
				if _, ok := tablesByID[seqID]; !ok {
					__antithesis_instrumentation__.Notify(11549)
					if !opts.SkipMissingSequenceOwners {
						__antithesis_instrumentation__.Notify(11550)
						return nil, errors.Errorf(
							"cannot restore table %q without referenced sequence %d (or %q option)",
							table.Name, seqID, restoreOptSkipMissingSequenceOwners)
					} else {
						__antithesis_instrumentation__.Notify(11551)
					}
				} else {
					__antithesis_instrumentation__.Notify(11552)
				}
			}
		}
		__antithesis_instrumentation__.Notify(11525)

		if table.IsSequence() && func() bool {
			__antithesis_instrumentation__.Notify(11553)
			return table.SequenceOpts.HasOwner() == true
		}() == true {
			__antithesis_instrumentation__.Notify(11554)
			if _, ok := tablesByID[table.SequenceOpts.SequenceOwner.OwnerTableID]; !ok {
				__antithesis_instrumentation__.Notify(11555)
				if !opts.SkipMissingSequenceOwners {
					__antithesis_instrumentation__.Notify(11556)
					return nil, errors.Errorf(
						"cannot restore sequence %q without referenced owner table %d (or %q option)",
						table.Name,
						table.SequenceOpts.SequenceOwner.OwnerTableID,
						restoreOptSkipMissingSequenceOwners,
					)
				} else {
					__antithesis_instrumentation__.Notify(11557)
				}
			} else {
				__antithesis_instrumentation__.Notify(11558)
			}
		} else {
			__antithesis_instrumentation__.Notify(11559)
		}
	}
	__antithesis_instrumentation__.Notify(11504)

	for _, database := range databasesByID {
		__antithesis_instrumentation__.Notify(11560)
		if int64(database.ID) > maxDescIDInBackup {
			__antithesis_instrumentation__.Notify(11561)
			maxDescIDInBackup = int64(database.ID)
		} else {
			__antithesis_instrumentation__.Notify(11562)
		}
	}
	__antithesis_instrumentation__.Notify(11505)

	for _, typ := range typesByID {
		__antithesis_instrumentation__.Notify(11563)
		if int64(typ.ID) > maxDescIDInBackup {
			__antithesis_instrumentation__.Notify(11564)
			maxDescIDInBackup = int64(typ.ID)
		} else {
			__antithesis_instrumentation__.Notify(11565)
		}
	}
	__antithesis_instrumentation__.Notify(11506)

	for _, sc := range schemasByID {
		__antithesis_instrumentation__.Notify(11566)
		if int64(sc.ID) > maxDescIDInBackup {
			__antithesis_instrumentation__.Notify(11567)
			maxDescIDInBackup = int64(sc.ID)
		} else {
			__antithesis_instrumentation__.Notify(11568)
		}
	}
	__antithesis_instrumentation__.Notify(11507)

	needsNewParentIDs := make(map[string][]descpb.ID)

	var tempSysDBID descpb.ID
	if descriptorCoverage == tree.AllDescriptors || func() bool {
		__antithesis_instrumentation__.Notify(11569)
		return restoreSystemUsers == true
	}() == true {
		__antithesis_instrumentation__.Notify(11570)
		var err error
		if descriptorCoverage == tree.AllDescriptors {
			__antithesis_instrumentation__.Notify(11575)

			if err = p.ExecCfg().DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
				__antithesis_instrumentation__.Notify(11577)
				v, err := txn.Get(ctx, p.ExecCfg().Codec.DescIDSequenceKey())
				if err != nil {
					__antithesis_instrumentation__.Notify(11580)
					return err
				} else {
					__antithesis_instrumentation__.Notify(11581)
				}
				__antithesis_instrumentation__.Notify(11578)
				newValue := maxDescIDInBackup + 1
				if newValue <= v.ValueInt() {
					__antithesis_instrumentation__.Notify(11582)

					newValue = v.ValueInt() + 1
				} else {
					__antithesis_instrumentation__.Notify(11583)
				}
				__antithesis_instrumentation__.Notify(11579)
				b := txn.NewBatch()

				b.Put(p.ExecCfg().Codec.DescIDSequenceKey(), newValue)
				return txn.Run(ctx, b)
			}); err != nil {
				__antithesis_instrumentation__.Notify(11584)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(11585)
			}
			__antithesis_instrumentation__.Notify(11576)
			tempSysDBID, err = descidgen.GenerateUniqueDescID(ctx, p.ExecCfg().DB, p.ExecCfg().Codec)
			if err != nil {
				__antithesis_instrumentation__.Notify(11586)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(11587)
			}
		} else {
			__antithesis_instrumentation__.Notify(11588)
			if restoreSystemUsers {
				__antithesis_instrumentation__.Notify(11589)
				tempSysDBID, err = descidgen.GenerateUniqueDescID(ctx, p.ExecCfg().DB, p.ExecCfg().Codec)
				if err != nil {
					__antithesis_instrumentation__.Notify(11590)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(11591)
				}
			} else {
				__antithesis_instrumentation__.Notify(11592)
			}
		}
		__antithesis_instrumentation__.Notify(11571)

		descriptorRewrites[tempSysDBID] = &jobspb.DescriptorRewrite{ID: tempSysDBID}
		for _, table := range tablesByID {
			__antithesis_instrumentation__.Notify(11593)
			if table.GetParentID() == systemschema.SystemDB.GetID() {
				__antithesis_instrumentation__.Notify(11594)
				descriptorRewrites[table.GetID()] = &jobspb.DescriptorRewrite{
					ParentID:       tempSysDBID,
					ParentSchemaID: keys.PublicSchemaIDForBackup,
				}
			} else {
				__antithesis_instrumentation__.Notify(11595)
			}
		}
		__antithesis_instrumentation__.Notify(11572)
		for _, sc := range schemasByID {
			__antithesis_instrumentation__.Notify(11596)
			if sc.GetParentID() == systemschema.SystemDB.GetID() {
				__antithesis_instrumentation__.Notify(11597)
				descriptorRewrites[sc.GetID()] = &jobspb.DescriptorRewrite{ParentID: tempSysDBID}
			} else {
				__antithesis_instrumentation__.Notify(11598)
			}
		}
		__antithesis_instrumentation__.Notify(11573)
		for _, typ := range typesByID {
			__antithesis_instrumentation__.Notify(11599)
			if typ.GetParentID() == systemschema.SystemDB.GetID() {
				__antithesis_instrumentation__.Notify(11600)
				descriptorRewrites[typ.GetID()] = &jobspb.DescriptorRewrite{
					ParentID:       tempSysDBID,
					ParentSchemaID: keys.PublicSchemaIDForBackup,
				}
			} else {
				__antithesis_instrumentation__.Notify(11601)
			}
		}
		__antithesis_instrumentation__.Notify(11574)

		haveSynthesizedTempSchema := make(map[descpb.ID]bool)
		var synthesizedTempSchemaCount int
		for _, table := range tablesByID {
			__antithesis_instrumentation__.Notify(11602)
			if table.IsTemporary() {
				__antithesis_instrumentation__.Notify(11603)
				if _, ok := haveSynthesizedTempSchema[table.GetParentSchemaID()]; !ok {
					__antithesis_instrumentation__.Notify(11604)
					var synthesizedSchemaID descpb.ID
					var err error

					schemaName := sql.TemporarySchemaNameForRestorePrefix +
						strconv.Itoa(synthesizedTempSchemaCount)
					synthesizedSchemaID, err = synthesizePGTempSchema(ctx, p, schemaName, table.GetParentID())
					if err != nil {
						__antithesis_instrumentation__.Notify(11606)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(11607)
					}
					__antithesis_instrumentation__.Notify(11605)

					descriptorRewrites[table.GetParentSchemaID()] = &jobspb.DescriptorRewrite{ID: synthesizedSchemaID}
					haveSynthesizedTempSchema[table.GetParentSchemaID()] = true
					synthesizedTempSchemaCount++
				} else {
					__antithesis_instrumentation__.Notify(11608)
				}
			} else {
				__antithesis_instrumentation__.Notify(11609)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(11610)
	}
	__antithesis_instrumentation__.Notify(11508)

	if err := sql.DescsTxn(ctx, p.ExecCfg(), func(ctx context.Context, txn *kv.Txn, col *descs.Collection) error {
		__antithesis_instrumentation__.Notify(11611)

		for name := range restoreDBNames {
			__antithesis_instrumentation__.Notify(11616)
			dbID, err := col.Direct().LookupDatabaseID(ctx, txn, name)
			if err != nil {
				__antithesis_instrumentation__.Notify(11618)
				return err
			} else {
				__antithesis_instrumentation__.Notify(11619)
			}
			__antithesis_instrumentation__.Notify(11617)
			if dbID != descpb.InvalidID {
				__antithesis_instrumentation__.Notify(11620)
				return errors.Errorf("database %q already exists", name)
			} else {
				__antithesis_instrumentation__.Notify(11621)
			}
		}
		__antithesis_instrumentation__.Notify(11612)

		for _, sc := range schemasByID {
			__antithesis_instrumentation__.Notify(11622)
			if _, ok := descriptorRewrites[sc.ID]; ok {
				__antithesis_instrumentation__.Notify(11625)
				continue
			} else {
				__antithesis_instrumentation__.Notify(11626)
			}
			__antithesis_instrumentation__.Notify(11623)

			targetDB, err := resolveTargetDB(ctx, txn, p, databasesByID, intoDB, descriptorCoverage, sc)
			if err != nil {
				__antithesis_instrumentation__.Notify(11627)
				return err
			} else {
				__antithesis_instrumentation__.Notify(11628)
			}
			__antithesis_instrumentation__.Notify(11624)

			if _, ok := restoreDBNames[targetDB]; ok {
				__antithesis_instrumentation__.Notify(11629)
				needsNewParentIDs[targetDB] = append(needsNewParentIDs[targetDB], sc.ID)
			} else {
				__antithesis_instrumentation__.Notify(11630)

				parentID, parentDB, err := getDatabaseIDAndDesc(ctx, txn, col, targetDB)
				if err != nil {
					__antithesis_instrumentation__.Notify(11634)
					return err
				} else {
					__antithesis_instrumentation__.Notify(11635)
				}
				__antithesis_instrumentation__.Notify(11631)
				if err := p.CheckPrivilege(ctx, parentDB, privilege.CREATE); err != nil {
					__antithesis_instrumentation__.Notify(11636)
					return err
				} else {
					__antithesis_instrumentation__.Notify(11637)
				}
				__antithesis_instrumentation__.Notify(11632)

				id, err := col.Direct().LookupSchemaID(ctx, txn, parentID, sc.Name)
				if err != nil {
					__antithesis_instrumentation__.Notify(11638)
					return err
				} else {
					__antithesis_instrumentation__.Notify(11639)
				}
				__antithesis_instrumentation__.Notify(11633)
				if id == descpb.InvalidID {
					__antithesis_instrumentation__.Notify(11640)

					descriptorRewrites[sc.ID] = &jobspb.DescriptorRewrite{ParentID: parentID}
				} else {
					__antithesis_instrumentation__.Notify(11641)

					desc, err := col.Direct().MustGetSchemaDescByID(ctx, txn, id)
					if err != nil {
						__antithesis_instrumentation__.Notify(11643)
						return err
					} else {
						__antithesis_instrumentation__.Notify(11644)
					}
					__antithesis_instrumentation__.Notify(11642)
					descriptorRewrites[sc.ID] = &jobspb.DescriptorRewrite{
						ParentID:   desc.GetParentID(),
						ID:         desc.GetID(),
						ToExisting: true,
					}
				}
			}
		}
		__antithesis_instrumentation__.Notify(11613)

		for _, table := range tablesByID {
			__antithesis_instrumentation__.Notify(11645)

			if _, ok := descriptorRewrites[table.ID]; ok {
				__antithesis_instrumentation__.Notify(11648)
				continue
			} else {
				__antithesis_instrumentation__.Notify(11649)
			}
			__antithesis_instrumentation__.Notify(11646)

			targetDB, err := resolveTargetDB(ctx, txn, p, databasesByID, intoDB, descriptorCoverage,
				table)
			if err != nil {
				__antithesis_instrumentation__.Notify(11650)
				return err
			} else {
				__antithesis_instrumentation__.Notify(11651)
			}
			__antithesis_instrumentation__.Notify(11647)

			if _, ok := restoreDBNames[targetDB]; ok {
				__antithesis_instrumentation__.Notify(11652)
				needsNewParentIDs[targetDB] = append(needsNewParentIDs[targetDB], table.ID)
			} else {
				__antithesis_instrumentation__.Notify(11653)
				if descriptorCoverage == tree.AllDescriptors {
					__antithesis_instrumentation__.Notify(11654)

					if targetDB != restoreTempSystemDB {
						__antithesis_instrumentation__.Notify(11655)
						descriptorRewrites[table.ID] = &jobspb.DescriptorRewrite{ParentID: table.ParentID}
					} else {
						__antithesis_instrumentation__.Notify(11656)
					}
				} else {
					__antithesis_instrumentation__.Notify(11657)
					var parentID descpb.ID
					{
						__antithesis_instrumentation__.Notify(11663)
						newParentID, err := col.Direct().LookupDatabaseID(ctx, txn, targetDB)
						if err != nil {
							__antithesis_instrumentation__.Notify(11666)
							return err
						} else {
							__antithesis_instrumentation__.Notify(11667)
						}
						__antithesis_instrumentation__.Notify(11664)
						if newParentID == descpb.InvalidID {
							__antithesis_instrumentation__.Notify(11668)
							return errors.Errorf("a database named %q needs to exist to restore table %q",
								targetDB, table.Name)
						} else {
							__antithesis_instrumentation__.Notify(11669)
						}
						__antithesis_instrumentation__.Notify(11665)
						parentID = newParentID
					}
					__antithesis_instrumentation__.Notify(11658)

					tableName := tree.NewUnqualifiedTableName(tree.Name(table.GetName()))
					err := col.Direct().CheckObjectCollision(ctx, txn, parentID, table.GetParentSchemaID(), tableName)
					if err != nil {
						__antithesis_instrumentation__.Notify(11670)
						return err
					} else {
						__antithesis_instrumentation__.Notify(11671)
					}
					__antithesis_instrumentation__.Notify(11659)

					parentDB, err := col.Direct().MustGetDatabaseDescByID(ctx, txn, parentID)
					if err != nil {
						__antithesis_instrumentation__.Notify(11672)
						return errors.Wrapf(err,
							"failed to lookup parent DB %d", errors.Safe(parentID))
					} else {
						__antithesis_instrumentation__.Notify(11673)
					}
					__antithesis_instrumentation__.Notify(11660)
					if err := p.CheckPrivilege(ctx, parentDB, privilege.CREATE); err != nil {
						__antithesis_instrumentation__.Notify(11674)
						return err
					} else {
						__antithesis_instrumentation__.Notify(11675)
					}
					__antithesis_instrumentation__.Notify(11661)

					if err := checkMultiRegionCompatible(ctx, txn, col, table, parentDB); err != nil {
						__antithesis_instrumentation__.Notify(11676)
						return pgerror.WithCandidateCode(err, pgcode.FeatureNotSupported)
					} else {
						__antithesis_instrumentation__.Notify(11677)
					}
					__antithesis_instrumentation__.Notify(11662)

					descriptorRewrites[table.ID] = &jobspb.DescriptorRewrite{ParentID: parentID}

					if table.GetParentSchemaID() == keys.PublicSchemaIDForBackup || func() bool {
						__antithesis_instrumentation__.Notify(11678)
						return table.GetParentSchemaID() == descpb.InvalidID == true
					}() == true {
						__antithesis_instrumentation__.Notify(11679)
						publicSchemaID := parentDB.GetSchemaID(tree.PublicSchema)
						descriptorRewrites[table.ID].ParentSchemaID = publicSchemaID
					} else {
						__antithesis_instrumentation__.Notify(11680)
					}
				}
			}
		}
		__antithesis_instrumentation__.Notify(11614)

		for _, typ := range typesByID {
			__antithesis_instrumentation__.Notify(11681)

			if _, ok := descriptorRewrites[typ.ID]; ok {
				__antithesis_instrumentation__.Notify(11684)
				continue
			} else {
				__antithesis_instrumentation__.Notify(11685)
			}
			__antithesis_instrumentation__.Notify(11682)

			targetDB, err := resolveTargetDB(ctx, txn, p, databasesByID, intoDB, descriptorCoverage, typ)
			if err != nil {
				__antithesis_instrumentation__.Notify(11686)
				return err
			} else {
				__antithesis_instrumentation__.Notify(11687)
			}
			__antithesis_instrumentation__.Notify(11683)

			if _, ok := restoreDBNames[targetDB]; ok {
				__antithesis_instrumentation__.Notify(11688)
				needsNewParentIDs[targetDB] = append(needsNewParentIDs[targetDB], typ.ID)
			} else {
				__antithesis_instrumentation__.Notify(11689)

				if typ.Kind == descpb.TypeDescriptor_ALIAS {
					__antithesis_instrumentation__.Notify(11697)
					continue
				} else {
					__antithesis_instrumentation__.Notify(11698)
				}
				__antithesis_instrumentation__.Notify(11690)

				parentID, err := col.Direct().LookupDatabaseID(ctx, txn, targetDB)
				if err != nil {
					__antithesis_instrumentation__.Notify(11699)
					return err
				} else {
					__antithesis_instrumentation__.Notify(11700)
				}
				__antithesis_instrumentation__.Notify(11691)
				if parentID == descpb.InvalidID {
					__antithesis_instrumentation__.Notify(11701)
					return errors.Errorf("a database named %q needs to exist to restore type %q",
						targetDB, typ.Name)
				} else {
					__antithesis_instrumentation__.Notify(11702)
				}
				__antithesis_instrumentation__.Notify(11692)

				parentDB, err := col.Direct().MustGetDatabaseDescByID(ctx, txn, parentID)
				if err != nil {
					__antithesis_instrumentation__.Notify(11703)
					return errors.Wrapf(err,
						"failed to lookup parent DB %d", errors.Safe(parentID))
				} else {
					__antithesis_instrumentation__.Notify(11704)
				}
				__antithesis_instrumentation__.Notify(11693)

				getParentSchemaID := func(typ *typedesc.Mutable) (parentSchemaID descpb.ID) {
					__antithesis_instrumentation__.Notify(11705)
					parentSchemaID = typ.GetParentSchemaID()

					if rewrite, ok := descriptorRewrites[parentSchemaID]; ok && func() bool {
						__antithesis_instrumentation__.Notify(11707)
						return rewrite.ID != 0 == true
					}() == true {
						__antithesis_instrumentation__.Notify(11708)
						parentSchemaID = rewrite.ID
					} else {
						__antithesis_instrumentation__.Notify(11709)
					}
					__antithesis_instrumentation__.Notify(11706)
					return
				}
				__antithesis_instrumentation__.Notify(11694)
				desc, err := col.Direct().GetDescriptorCollidingWithObject(
					ctx,
					txn,
					parentID,
					getParentSchemaID(typ),
					typ.Name,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(11710)
					return err
				} else {
					__antithesis_instrumentation__.Notify(11711)
				}
				__antithesis_instrumentation__.Notify(11695)
				if desc == nil {
					__antithesis_instrumentation__.Notify(11712)

					if err := p.CheckPrivilege(ctx, parentDB, privilege.CREATE); err != nil {
						__antithesis_instrumentation__.Notify(11715)
						return err
					} else {
						__antithesis_instrumentation__.Notify(11716)
					}
					__antithesis_instrumentation__.Notify(11713)

					descriptorRewrites[typ.ID] = &jobspb.DescriptorRewrite{ParentID: parentID}

					arrTyp := typesByID[typ.ArrayTypeID]
					typeName := tree.NewUnqualifiedTypeName(arrTyp.GetName())
					err = col.Direct().CheckObjectCollision(ctx, txn, parentID, getParentSchemaID(typ), typeName)
					if err != nil {
						__antithesis_instrumentation__.Notify(11717)
						return errors.Wrapf(err, "name collision for %q's array type", typ.Name)
					} else {
						__antithesis_instrumentation__.Notify(11718)
					}
					__antithesis_instrumentation__.Notify(11714)

					descriptorRewrites[arrTyp.ID] = &jobspb.DescriptorRewrite{ParentID: parentID}
				} else {
					__antithesis_instrumentation__.Notify(11719)

					existingType, isType := desc.(catalog.TypeDescriptor)
					if !isType {
						__antithesis_instrumentation__.Notify(11722)
						return sqlerrors.MakeObjectAlreadyExistsError(desc.DescriptorProto(), typ.Name)
					} else {
						__antithesis_instrumentation__.Notify(11723)
					}
					__antithesis_instrumentation__.Notify(11720)

					if err := typ.IsCompatibleWith(existingType); err != nil {
						__antithesis_instrumentation__.Notify(11724)
						return errors.Wrapf(
							err,
							"%q is not compatible with type %q existing in cluster",
							existingType.GetName(),
							existingType.GetName(),
						)
					} else {
						__antithesis_instrumentation__.Notify(11725)
					}
					__antithesis_instrumentation__.Notify(11721)

					descriptorRewrites[typ.ID] = &jobspb.DescriptorRewrite{
						ParentID:   existingType.GetParentID(),
						ID:         existingType.GetID(),
						ToExisting: true,
					}
					descriptorRewrites[typ.ArrayTypeID] = &jobspb.DescriptorRewrite{
						ParentID:   existingType.GetParentID(),
						ID:         existingType.GetArrayTypeID(),
						ToExisting: true,
					}
				}
				__antithesis_instrumentation__.Notify(11696)

				if typ.GetParentSchemaID() == keys.PublicSchemaIDForBackup || func() bool {
					__antithesis_instrumentation__.Notify(11726)
					return typ.GetParentSchemaID() == descpb.InvalidID == true
				}() == true {
					__antithesis_instrumentation__.Notify(11727)
					publicSchemaID := parentDB.GetSchemaID(tree.PublicSchema)
					descriptorRewrites[typ.ID].ParentSchemaID = publicSchemaID
					descriptorRewrites[typ.ArrayTypeID].ParentSchemaID = publicSchemaID
				} else {
					__antithesis_instrumentation__.Notify(11728)
				}
			}
		}
		__antithesis_instrumentation__.Notify(11615)

		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(11729)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(11730)
	}
	__antithesis_instrumentation__.Notify(11509)

	for _, db := range restoreDBs {
		__antithesis_instrumentation__.Notify(11731)
		var newID descpb.ID
		var err error
		if descriptorCoverage == tree.AllDescriptors {
			__antithesis_instrumentation__.Notify(11734)
			newID = db.GetID()
		} else {
			__antithesis_instrumentation__.Notify(11735)
			newID, err = descidgen.GenerateUniqueDescID(ctx, p.ExecCfg().DB, p.ExecCfg().Codec)
			if err != nil {
				__antithesis_instrumentation__.Notify(11736)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(11737)
			}
		}
		__antithesis_instrumentation__.Notify(11732)

		descriptorRewrites[db.GetID()] = &jobspb.DescriptorRewrite{ID: newID}

		if newDBName != "" {
			__antithesis_instrumentation__.Notify(11738)
			descriptorRewrites[db.GetID()].NewDBName = newDBName
		} else {
			__antithesis_instrumentation__.Notify(11739)
		}
		__antithesis_instrumentation__.Notify(11733)

		for _, objectID := range needsNewParentIDs[db.GetName()] {
			__antithesis_instrumentation__.Notify(11740)
			descriptorRewrites[objectID] = &jobspb.DescriptorRewrite{ParentID: newID}
		}
	}
	__antithesis_instrumentation__.Notify(11510)

	descriptorsToRemap := make([]catalog.Descriptor, 0, len(tablesByID))
	for _, table := range tablesByID {
		__antithesis_instrumentation__.Notify(11741)
		if descriptorCoverage == tree.AllDescriptors || func() bool {
			__antithesis_instrumentation__.Notify(11742)
			return restoreSystemUsers == true
		}() == true {
			__antithesis_instrumentation__.Notify(11743)
			if table.ParentID == systemschema.SystemDB.GetID() {
				__antithesis_instrumentation__.Notify(11744)

				descriptorsToRemap = append(descriptorsToRemap, table)
			} else {
				__antithesis_instrumentation__.Notify(11745)

				descriptorRewrites[table.ID].ID = table.ID
			}
		} else {
			__antithesis_instrumentation__.Notify(11746)
			descriptorsToRemap = append(descriptorsToRemap, table)
		}
	}
	__antithesis_instrumentation__.Notify(11511)

	for _, typ := range typesByID {
		__antithesis_instrumentation__.Notify(11747)
		if descriptorCoverage == tree.AllDescriptors {
			__antithesis_instrumentation__.Notify(11748)

			descriptorRewrites[typ.ID].ID = typ.ID
		} else {
			__antithesis_instrumentation__.Notify(11749)

			if !descriptorRewrites[typ.ID].ToExisting {
				__antithesis_instrumentation__.Notify(11750)
				descriptorsToRemap = append(descriptorsToRemap, typ)
			} else {
				__antithesis_instrumentation__.Notify(11751)
			}
		}
	}
	__antithesis_instrumentation__.Notify(11512)

	for _, sc := range schemasByID {
		__antithesis_instrumentation__.Notify(11752)
		if descriptorCoverage == tree.AllDescriptors {
			__antithesis_instrumentation__.Notify(11753)

			descriptorRewrites[sc.ID].ID = sc.ID
		} else {
			__antithesis_instrumentation__.Notify(11754)

			if !descriptorRewrites[sc.ID].ToExisting {
				__antithesis_instrumentation__.Notify(11755)
				descriptorsToRemap = append(descriptorsToRemap, sc)
			} else {
				__antithesis_instrumentation__.Notify(11756)
			}
		}
	}
	__antithesis_instrumentation__.Notify(11513)

	sort.Sort(catalog.Descriptors(descriptorsToRemap))

	for _, desc := range descriptorsToRemap {
		__antithesis_instrumentation__.Notify(11757)
		id, err := descidgen.GenerateUniqueDescID(ctx, p.ExecCfg().DB, p.ExecCfg().Codec)
		if err != nil {
			__antithesis_instrumentation__.Notify(11759)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(11760)
		}
		__antithesis_instrumentation__.Notify(11758)
		descriptorRewrites[desc.GetID()].ID = id
	}
	__antithesis_instrumentation__.Notify(11514)

	rewriteObject := func(desc catalog.Descriptor) {
		__antithesis_instrumentation__.Notify(11761)
		if descriptorRewrites[desc.GetID()].ParentSchemaID != descpb.InvalidID {
			__antithesis_instrumentation__.Notify(11764)

			return
		} else {
			__antithesis_instrumentation__.Notify(11765)
		}
		__antithesis_instrumentation__.Notify(11762)
		curSchemaID := desc.GetParentSchemaID()
		newSchemaID := curSchemaID
		if rw, ok := descriptorRewrites[curSchemaID]; ok {
			__antithesis_instrumentation__.Notify(11766)
			newSchemaID = rw.ID
		} else {
			__antithesis_instrumentation__.Notify(11767)
		}
		__antithesis_instrumentation__.Notify(11763)
		descriptorRewrites[desc.GetID()].ParentSchemaID = newSchemaID
	}
	__antithesis_instrumentation__.Notify(11515)
	for _, table := range tablesByID {
		__antithesis_instrumentation__.Notify(11768)
		rewriteObject(table)
	}
	__antithesis_instrumentation__.Notify(11516)
	for _, typ := range typesByID {
		__antithesis_instrumentation__.Notify(11769)
		rewriteObject(typ)
	}
	__antithesis_instrumentation__.Notify(11517)

	return descriptorRewrites, nil
}

func getDatabaseIDAndDesc(
	ctx context.Context, txn *kv.Txn, col *descs.Collection, targetDB string,
) (dbID descpb.ID, dbDesc catalog.DatabaseDescriptor, err error) {
	__antithesis_instrumentation__.Notify(11770)
	dbID, err = col.Direct().LookupDatabaseID(ctx, txn, targetDB)
	if err != nil {
		__antithesis_instrumentation__.Notify(11774)
		return 0, nil, err
	} else {
		__antithesis_instrumentation__.Notify(11775)
	}
	__antithesis_instrumentation__.Notify(11771)
	if dbID == descpb.InvalidID {
		__antithesis_instrumentation__.Notify(11776)
		return dbID, nil, errors.Errorf("a database named %q needs to exist", targetDB)
	} else {
		__antithesis_instrumentation__.Notify(11777)
	}
	__antithesis_instrumentation__.Notify(11772)

	dbDesc, err = col.Direct().MustGetDatabaseDescByID(ctx, txn, dbID)
	if err != nil {
		__antithesis_instrumentation__.Notify(11778)
		return 0, nil, errors.Wrapf(err,
			"failed to lookup parent DB %d", errors.Safe(dbID))
	} else {
		__antithesis_instrumentation__.Notify(11779)
	}
	__antithesis_instrumentation__.Notify(11773)
	return dbID, dbDesc, nil
}

func dropDefaultUserDBs(ctx context.Context, execCfg *sql.ExecutorConfig) error {
	__antithesis_instrumentation__.Notify(11780)
	return sql.DescsTxn(ctx, execCfg, func(ctx context.Context, txn *kv.Txn, col *descs.Collection) error {
		__antithesis_instrumentation__.Notify(11781)
		ie := execCfg.InternalExecutor
		_, err := ie.Exec(ctx, "drop-defaultdb", nil, "DROP DATABASE IF EXISTS defaultdb")
		if err != nil {
			__antithesis_instrumentation__.Notify(11784)
			return err
		} else {
			__antithesis_instrumentation__.Notify(11785)
		}
		__antithesis_instrumentation__.Notify(11782)

		_, err = ie.Exec(ctx, "drop-postgres", nil, "DROP DATABASE IF EXISTS postgres")
		if err != nil {
			__antithesis_instrumentation__.Notify(11786)
			return err
		} else {
			__antithesis_instrumentation__.Notify(11787)
		}
		__antithesis_instrumentation__.Notify(11783)
		return nil
	})
}

func resolveTargetDB(
	ctx context.Context,
	txn *kv.Txn,
	p sql.PlanHookState,
	databasesByID map[descpb.ID]*dbdesc.Mutable,
	intoDB string,
	descriptorCoverage tree.DescriptorCoverage,
	descriptor catalog.Descriptor,
) (string, error) {
	__antithesis_instrumentation__.Notify(11788)
	if intoDB != "" {
		__antithesis_instrumentation__.Notify(11792)
		return intoDB, nil
	} else {
		__antithesis_instrumentation__.Notify(11793)
	}
	__antithesis_instrumentation__.Notify(11789)

	if descriptorCoverage == tree.AllDescriptors && func() bool {
		__antithesis_instrumentation__.Notify(11794)
		return catalog.IsSystemDescriptor(descriptor) == true
	}() == true {
		__antithesis_instrumentation__.Notify(11795)
		var targetDB string
		if descriptor.GetParentID() == systemschema.SystemDB.GetID() {
			__antithesis_instrumentation__.Notify(11797)

			targetDB = restoreTempSystemDB
		} else {
			__antithesis_instrumentation__.Notify(11798)
		}
		__antithesis_instrumentation__.Notify(11796)
		return targetDB, nil
	} else {
		__antithesis_instrumentation__.Notify(11799)
	}
	__antithesis_instrumentation__.Notify(11790)

	database, ok := databasesByID[descriptor.GetParentID()]
	if !ok {
		__antithesis_instrumentation__.Notify(11800)
		return "", errors.Errorf("no database with ID %d in backup for object %q (%d)",
			descriptor.GetParentID(), descriptor.GetName(), descriptor.GetID())
	} else {
		__antithesis_instrumentation__.Notify(11801)
	}
	__antithesis_instrumentation__.Notify(11791)
	return database.Name, nil
}

func maybeUpgradeDescriptors(descs []catalog.Descriptor, skipFKsWithNoMatchingTable bool) error {
	__antithesis_instrumentation__.Notify(11802)
	for j, desc := range descs {
		__antithesis_instrumentation__.Notify(11804)
		var b catalog.DescriptorBuilder
		if tableDesc, isTable := desc.(catalog.TableDescriptor); isTable {
			__antithesis_instrumentation__.Notify(11809)
			b = tabledesc.NewBuilderForFKUpgrade(tableDesc.TableDesc(), skipFKsWithNoMatchingTable)
		} else {
			__antithesis_instrumentation__.Notify(11810)
			b = desc.NewBuilder()
		}
		__antithesis_instrumentation__.Notify(11805)
		if err := b.RunPostDeserializationChanges(); err != nil {
			__antithesis_instrumentation__.Notify(11811)
			return errors.NewAssertionErrorWithWrappedErrf(err, "error during RunPostDeserializationChanges")
		} else {
			__antithesis_instrumentation__.Notify(11812)
		}
		__antithesis_instrumentation__.Notify(11806)
		err := b.RunRestoreChanges(func(id descpb.ID) catalog.Descriptor {
			__antithesis_instrumentation__.Notify(11813)
			for _, d := range descs {
				__antithesis_instrumentation__.Notify(11815)
				if d.GetID() == id {
					__antithesis_instrumentation__.Notify(11816)
					return d
				} else {
					__antithesis_instrumentation__.Notify(11817)
				}
			}
			__antithesis_instrumentation__.Notify(11814)
			return nil
		})
		__antithesis_instrumentation__.Notify(11807)
		if err != nil {
			__antithesis_instrumentation__.Notify(11818)
			return err
		} else {
			__antithesis_instrumentation__.Notify(11819)
		}
		__antithesis_instrumentation__.Notify(11808)
		descs[j] = b.BuildExistingMutable()
	}
	__antithesis_instrumentation__.Notify(11803)
	return nil
}

func maybeUpgradeDescriptorsInBackupManifests(
	backupManifests []BackupManifest, skipFKsWithNoMatchingTable bool,
) error {
	__antithesis_instrumentation__.Notify(11820)
	if len(backupManifests) == 0 {
		__antithesis_instrumentation__.Notify(11825)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(11826)
	}
	__antithesis_instrumentation__.Notify(11821)
	descriptors := make([]catalog.Descriptor, 0, len(backupManifests[0].Descriptors))
	for _, backupManifest := range backupManifests {
		__antithesis_instrumentation__.Notify(11827)
		for _, pb := range backupManifest.Descriptors {
			__antithesis_instrumentation__.Notify(11828)
			descriptors = append(descriptors, descbuilder.NewBuilder(&pb).BuildExistingMutable())
		}
	}
	__antithesis_instrumentation__.Notify(11822)

	err := maybeUpgradeDescriptors(descriptors, skipFKsWithNoMatchingTable)
	if err != nil {
		__antithesis_instrumentation__.Notify(11829)
		return err
	} else {
		__antithesis_instrumentation__.Notify(11830)
	}
	__antithesis_instrumentation__.Notify(11823)

	k := 0
	for i := range backupManifests {
		__antithesis_instrumentation__.Notify(11831)
		manifest := &backupManifests[i]
		for j := range manifest.Descriptors {
			__antithesis_instrumentation__.Notify(11832)
			manifest.Descriptors[j] = *descriptors[k].DescriptorProto()
			k++
		}
	}
	__antithesis_instrumentation__.Notify(11824)
	return nil
}

func resolveOptionsForRestoreJobDescription(
	opts tree.RestoreOptions, intoDB string, newDBName string, kmsURIs []string, incFrom []string,
) (tree.RestoreOptions, error) {
	__antithesis_instrumentation__.Notify(11833)
	if opts.IsDefault() {
		__antithesis_instrumentation__.Notify(11840)
		return opts, nil
	} else {
		__antithesis_instrumentation__.Notify(11841)
	}
	__antithesis_instrumentation__.Notify(11834)

	newOpts := tree.RestoreOptions{
		SkipMissingFKs:            opts.SkipMissingFKs,
		SkipMissingSequences:      opts.SkipMissingSequences,
		SkipMissingSequenceOwners: opts.SkipMissingSequenceOwners,
		SkipMissingViews:          opts.SkipMissingViews,
		Detached:                  opts.Detached,
	}

	if opts.EncryptionPassphrase != nil {
		__antithesis_instrumentation__.Notify(11842)
		newOpts.EncryptionPassphrase = tree.NewDString("redacted")
	} else {
		__antithesis_instrumentation__.Notify(11843)
	}
	__antithesis_instrumentation__.Notify(11835)

	if opts.IntoDB != nil {
		__antithesis_instrumentation__.Notify(11844)
		newOpts.IntoDB = tree.NewDString(intoDB)
	} else {
		__antithesis_instrumentation__.Notify(11845)
	}
	__antithesis_instrumentation__.Notify(11836)

	if opts.NewDBName != nil {
		__antithesis_instrumentation__.Notify(11846)
		newOpts.NewDBName = tree.NewDString(newDBName)
	} else {
		__antithesis_instrumentation__.Notify(11847)
	}
	__antithesis_instrumentation__.Notify(11837)

	for _, uri := range kmsURIs {
		__antithesis_instrumentation__.Notify(11848)
		redactedURI, err := cloud.RedactKMSURI(uri)
		if err != nil {
			__antithesis_instrumentation__.Notify(11850)
			return tree.RestoreOptions{}, err
		} else {
			__antithesis_instrumentation__.Notify(11851)
		}
		__antithesis_instrumentation__.Notify(11849)
		newOpts.DecryptionKMSURI = append(newOpts.DecryptionKMSURI, tree.NewDString(redactedURI))
	}
	__antithesis_instrumentation__.Notify(11838)

	if opts.IncrementalStorage != nil {
		__antithesis_instrumentation__.Notify(11852)
		var err error
		newOpts.IncrementalStorage, err = sanitizeURIList(incFrom)
		if err != nil {
			__antithesis_instrumentation__.Notify(11853)
			return tree.RestoreOptions{}, err
		} else {
			__antithesis_instrumentation__.Notify(11854)
		}
	} else {
		__antithesis_instrumentation__.Notify(11855)
	}
	__antithesis_instrumentation__.Notify(11839)

	return newOpts, nil
}

func restoreJobDescription(
	p sql.PlanHookState,
	restore *tree.Restore,
	from [][]string,
	incFrom []string,
	opts tree.RestoreOptions,
	intoDB string,
	newDBName string,
	kmsURIs []string,
) (string, error) {
	__antithesis_instrumentation__.Notify(11856)
	r := &tree.Restore{
		SystemUsers:        restore.SystemUsers,
		DescriptorCoverage: restore.DescriptorCoverage,
		AsOf:               restore.AsOf,
		Targets:            restore.Targets,
		From:               make([]tree.StringOrPlaceholderOptList, len(restore.From)),
	}

	var options tree.RestoreOptions
	var err error
	if options, err = resolveOptionsForRestoreJobDescription(opts, intoDB, newDBName,
		kmsURIs, incFrom); err != nil {
		__antithesis_instrumentation__.Notify(11859)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(11860)
	}
	__antithesis_instrumentation__.Notify(11857)
	r.Options = options

	for i, backup := range from {
		__antithesis_instrumentation__.Notify(11861)
		r.From[i] = make(tree.StringOrPlaceholderOptList, len(backup))
		r.From[i], err = sanitizeURIList(backup)
		if err != nil {
			__antithesis_instrumentation__.Notify(11862)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(11863)
		}
	}
	__antithesis_instrumentation__.Notify(11858)

	ann := p.ExtendedEvalContext().Annotations
	return tree.AsStringWithFQNames(r, ann), nil
}

func restorePlanHook(
	ctx context.Context, stmt tree.Statement, p sql.PlanHookState,
) (sql.PlanHookRowFn, colinfo.ResultColumns, []sql.PlanNode, bool, error) {
	__antithesis_instrumentation__.Notify(11864)
	restoreStmt, ok := stmt.(*tree.Restore)
	if !ok {
		__antithesis_instrumentation__.Notify(11878)
		return nil, nil, nil, false, nil
	} else {
		__antithesis_instrumentation__.Notify(11879)
	}
	__antithesis_instrumentation__.Notify(11865)

	if err := featureflag.CheckEnabled(
		ctx,
		p.ExecCfg(),
		featureRestoreEnabled,
		"RESTORE",
	); err != nil {
		__antithesis_instrumentation__.Notify(11880)
		return nil, nil, nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(11881)
	}
	__antithesis_instrumentation__.Notify(11866)

	fromFns := make([]func() ([]string, error), len(restoreStmt.From))
	for i := range restoreStmt.From {
		__antithesis_instrumentation__.Notify(11882)
		fromFn, err := p.TypeAsStringArray(ctx, tree.Exprs(restoreStmt.From[i]), "RESTORE")
		if err != nil {
			__antithesis_instrumentation__.Notify(11884)
			return nil, nil, nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(11885)
		}
		__antithesis_instrumentation__.Notify(11883)
		fromFns[i] = fromFn
	}
	__antithesis_instrumentation__.Notify(11867)

	var pwFn func() (string, error)
	var err error
	if restoreStmt.Options.EncryptionPassphrase != nil {
		__antithesis_instrumentation__.Notify(11886)
		pwFn, err = p.TypeAsString(ctx, restoreStmt.Options.EncryptionPassphrase, "RESTORE")
		if err != nil {
			__antithesis_instrumentation__.Notify(11887)
			return nil, nil, nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(11888)
		}
	} else {
		__antithesis_instrumentation__.Notify(11889)
	}
	__antithesis_instrumentation__.Notify(11868)

	var kmsFn func() ([]string, error)
	if restoreStmt.Options.DecryptionKMSURI != nil {
		__antithesis_instrumentation__.Notify(11890)
		if restoreStmt.Options.EncryptionPassphrase != nil {
			__antithesis_instrumentation__.Notify(11892)
			return nil, nil, nil, false, errors.New("cannot have both encryption_passphrase and kms option set")
		} else {
			__antithesis_instrumentation__.Notify(11893)
		}
		__antithesis_instrumentation__.Notify(11891)
		kmsFn, err = p.TypeAsStringArray(ctx, tree.Exprs(restoreStmt.Options.DecryptionKMSURI),
			"RESTORE")
		if err != nil {
			__antithesis_instrumentation__.Notify(11894)
			return nil, nil, nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(11895)
		}
	} else {
		__antithesis_instrumentation__.Notify(11896)
	}
	__antithesis_instrumentation__.Notify(11869)

	var intoDBFn func() (string, error)
	if restoreStmt.Options.IntoDB != nil {
		__antithesis_instrumentation__.Notify(11897)
		if restoreStmt.SystemUsers {
			__antithesis_instrumentation__.Notify(11899)
			return nil, nil, nil, false, errors.New("cannot set into_db option when only restoring system users")
		} else {
			__antithesis_instrumentation__.Notify(11900)
		}
		__antithesis_instrumentation__.Notify(11898)
		intoDBFn, err = p.TypeAsString(ctx, restoreStmt.Options.IntoDB, "RESTORE")
		if err != nil {
			__antithesis_instrumentation__.Notify(11901)
			return nil, nil, nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(11902)
		}
	} else {
		__antithesis_instrumentation__.Notify(11903)
	}
	__antithesis_instrumentation__.Notify(11870)

	subdirFn := func() (string, error) { __antithesis_instrumentation__.Notify(11904); return "", nil }
	__antithesis_instrumentation__.Notify(11871)
	if restoreStmt.Subdir != nil {
		__antithesis_instrumentation__.Notify(11905)
		subdirFn, err = p.TypeAsString(ctx, restoreStmt.Subdir, "RESTORE")
		if err != nil {
			__antithesis_instrumentation__.Notify(11906)
			return nil, nil, nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(11907)
		}
	} else {
		__antithesis_instrumentation__.Notify(11908)

		p.BufferClientNotice(ctx,
			pgnotice.Newf("The `RESTORE FROM <backup>` syntax will be removed in a future release, please"+
				" switch over to using `RESTORE FROM <backup> IN <collection>` to restore a particular backup from a collection: %s",
				"https://www.cockroachlabs.com/docs/stable/restore.html#view-the-backup-subdirectories"))
	}
	__antithesis_instrumentation__.Notify(11872)

	var incStorageFn func() ([]string, error)
	if restoreStmt.Options.IncrementalStorage != nil {
		__antithesis_instrumentation__.Notify(11909)
		if restoreStmt.Subdir == nil {
			__antithesis_instrumentation__.Notify(11911)
			err = errors.New("incremental_location can only be used with the following" +
				" syntax: 'RESTORE [target] FROM [subdirectory] IN [destination]'")
			return nil, nil, nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(11912)
		}
		__antithesis_instrumentation__.Notify(11910)
		incStorageFn, err = p.TypeAsStringArray(ctx, tree.Exprs(restoreStmt.Options.IncrementalStorage),
			"RESTORE")
		if err != nil {
			__antithesis_instrumentation__.Notify(11913)
			return nil, nil, nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(11914)
		}
	} else {
		__antithesis_instrumentation__.Notify(11915)
	}
	__antithesis_instrumentation__.Notify(11873)

	var newDBNameFn func() (string, error)
	if restoreStmt.Options.NewDBName != nil {
		__antithesis_instrumentation__.Notify(11916)
		if restoreStmt.DescriptorCoverage == tree.AllDescriptors || func() bool {
			__antithesis_instrumentation__.Notify(11919)
			return len(restoreStmt.Targets.Databases) != 1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(11920)
			err = errors.New("new_db_name can only be used for RESTORE DATABASE with a single target" +
				" database")
			return nil, nil, nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(11921)
		}
		__antithesis_instrumentation__.Notify(11917)
		if restoreStmt.SystemUsers {
			__antithesis_instrumentation__.Notify(11922)
			return nil, nil, nil, false, errors.New("cannot set new_db_name option when only restoring system users")
		} else {
			__antithesis_instrumentation__.Notify(11923)
		}
		__antithesis_instrumentation__.Notify(11918)
		newDBNameFn, err = p.TypeAsString(ctx, restoreStmt.Options.NewDBName, "RESTORE")
		if err != nil {
			__antithesis_instrumentation__.Notify(11924)
			return nil, nil, nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(11925)
		}
	} else {
		__antithesis_instrumentation__.Notify(11926)
	}
	__antithesis_instrumentation__.Notify(11874)

	var newTenantIDFn func() (*roachpb.TenantID, error)
	if restoreStmt.Options.AsTenant != nil {
		__antithesis_instrumentation__.Notify(11927)
		if restoreStmt.DescriptorCoverage == tree.AllDescriptors || func() bool {
			__antithesis_instrumentation__.Notify(11930)
			return !restoreStmt.Targets.TenantID.IsSet() == true
		}() == true {
			__antithesis_instrumentation__.Notify(11931)
			err := errors.Errorf("%q can only be used when running RESTORE TENANT for a single tenant", restoreOptAsTenant)
			return nil, nil, nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(11932)
		}
		__antithesis_instrumentation__.Notify(11928)

		fn, err := p.TypeAsString(ctx, restoreStmt.Options.AsTenant, "RESTORE")
		if err != nil {
			__antithesis_instrumentation__.Notify(11933)
			return nil, nil, nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(11934)
		}
		__antithesis_instrumentation__.Notify(11929)
		newTenantIDFn = func() (*roachpb.TenantID, error) {
			__antithesis_instrumentation__.Notify(11935)
			s, err := fn()
			if err != nil {
				__antithesis_instrumentation__.Notify(11939)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(11940)
			}
			__antithesis_instrumentation__.Notify(11936)
			x, err := strconv.Atoi(s)
			if err != nil {
				__antithesis_instrumentation__.Notify(11941)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(11942)
			}
			__antithesis_instrumentation__.Notify(11937)
			if x < int(roachpb.MinTenantID.ToUint64()) {
				__antithesis_instrumentation__.Notify(11943)
				return nil, errors.New("invalid tenant ID value")
			} else {
				__antithesis_instrumentation__.Notify(11944)
			}
			__antithesis_instrumentation__.Notify(11938)
			id := roachpb.MakeTenantID(uint64(x))
			return &id, nil
		}
	} else {
		__antithesis_instrumentation__.Notify(11945)
	}
	__antithesis_instrumentation__.Notify(11875)

	fn := func(ctx context.Context, _ []sql.PlanNode, resultsCh chan<- tree.Datums) error {
		__antithesis_instrumentation__.Notify(11946)

		ctx, span := tracing.ChildSpan(ctx, stmt.StatementTag())
		defer span.Finish()

		if !(p.ExtendedEvalContext().TxnIsSingleStmt || func() bool {
			__antithesis_instrumentation__.Notify(11958)
			return restoreStmt.Options.Detached == true
		}() == true) {
			__antithesis_instrumentation__.Notify(11959)
			return errors.Errorf("RESTORE cannot be used inside a multi-statement transaction without DETACHED option")
		} else {
			__antithesis_instrumentation__.Notify(11960)
		}
		__antithesis_instrumentation__.Notify(11947)

		subdir, err := subdirFn()
		if err != nil {
			__antithesis_instrumentation__.Notify(11961)
			return err
		} else {
			__antithesis_instrumentation__.Notify(11962)
		}
		__antithesis_instrumentation__.Notify(11948)

		from := make([][]string, len(fromFns))
		for i := range fromFns {
			__antithesis_instrumentation__.Notify(11963)
			from[i], err = fromFns[i]()
			if err != nil {
				__antithesis_instrumentation__.Notify(11964)
				return err
			} else {
				__antithesis_instrumentation__.Notify(11965)
			}
		}
		__antithesis_instrumentation__.Notify(11949)

		if err := checkPrivilegesForRestore(ctx, restoreStmt, p, from); err != nil {
			__antithesis_instrumentation__.Notify(11966)
			return err
		} else {
			__antithesis_instrumentation__.Notify(11967)
		}
		__antithesis_instrumentation__.Notify(11950)

		var endTime hlc.Timestamp
		if restoreStmt.AsOf.Expr != nil {
			__antithesis_instrumentation__.Notify(11968)
			asOf, err := p.EvalAsOfTimestamp(ctx, restoreStmt.AsOf)
			if err != nil {
				__antithesis_instrumentation__.Notify(11970)
				return err
			} else {
				__antithesis_instrumentation__.Notify(11971)
			}
			__antithesis_instrumentation__.Notify(11969)
			endTime = asOf.Timestamp
		} else {
			__antithesis_instrumentation__.Notify(11972)
		}
		__antithesis_instrumentation__.Notify(11951)

		var passphrase string
		if pwFn != nil {
			__antithesis_instrumentation__.Notify(11973)
			passphrase, err = pwFn()
			if err != nil {
				__antithesis_instrumentation__.Notify(11974)
				return err
			} else {
				__antithesis_instrumentation__.Notify(11975)
			}
		} else {
			__antithesis_instrumentation__.Notify(11976)
		}
		__antithesis_instrumentation__.Notify(11952)

		var kms []string
		if kmsFn != nil {
			__antithesis_instrumentation__.Notify(11977)
			kms, err = kmsFn()
			if err != nil {
				__antithesis_instrumentation__.Notify(11978)
				return err
			} else {
				__antithesis_instrumentation__.Notify(11979)
			}
		} else {
			__antithesis_instrumentation__.Notify(11980)
		}
		__antithesis_instrumentation__.Notify(11953)

		var intoDB string
		if intoDBFn != nil {
			__antithesis_instrumentation__.Notify(11981)
			intoDB, err = intoDBFn()
			if err != nil {
				__antithesis_instrumentation__.Notify(11982)
				return err
			} else {
				__antithesis_instrumentation__.Notify(11983)
			}
		} else {
			__antithesis_instrumentation__.Notify(11984)
		}
		__antithesis_instrumentation__.Notify(11954)

		var newDBName string
		if newDBNameFn != nil {
			__antithesis_instrumentation__.Notify(11985)
			newDBName, err = newDBNameFn()
			if err != nil {
				__antithesis_instrumentation__.Notify(11986)
				return err
			} else {
				__antithesis_instrumentation__.Notify(11987)
			}
		} else {
			__antithesis_instrumentation__.Notify(11988)
		}
		__antithesis_instrumentation__.Notify(11955)
		var newTenantID *roachpb.TenantID
		if newTenantIDFn != nil {
			__antithesis_instrumentation__.Notify(11989)
			newTenantID, err = newTenantIDFn()
			if err != nil {
				__antithesis_instrumentation__.Notify(11990)
				return err
			} else {
				__antithesis_instrumentation__.Notify(11991)
			}
		} else {
			__antithesis_instrumentation__.Notify(11992)
		}
		__antithesis_instrumentation__.Notify(11956)

		var incFrom []string
		if incStorageFn != nil {
			__antithesis_instrumentation__.Notify(11993)
			incFrom, err = incStorageFn()
			if err != nil {
				__antithesis_instrumentation__.Notify(11994)
				return err
			} else {
				__antithesis_instrumentation__.Notify(11995)
			}
		} else {
			__antithesis_instrumentation__.Notify(11996)
		}
		__antithesis_instrumentation__.Notify(11957)

		return doRestorePlan(ctx, restoreStmt, p, from, incFrom, passphrase, kms, intoDB,
			newDBName, newTenantID, endTime, resultsCh, subdir)
	}
	__antithesis_instrumentation__.Notify(11876)

	if restoreStmt.Options.Detached {
		__antithesis_instrumentation__.Notify(11997)
		return fn, jobs.DetachedJobExecutionResultHeader, nil, false, nil
	} else {
		__antithesis_instrumentation__.Notify(11998)
	}
	__antithesis_instrumentation__.Notify(11877)
	return fn, jobs.BulkJobExecutionResultHeader, nil, false, nil
}

func checkPrivilegesForRestore(
	ctx context.Context, restoreStmt *tree.Restore, p sql.PlanHookState, from [][]string,
) error {
	__antithesis_instrumentation__.Notify(11999)
	hasAdmin, err := p.HasAdminRole(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(12007)
		return err
	} else {
		__antithesis_instrumentation__.Notify(12008)
	}
	__antithesis_instrumentation__.Notify(12000)
	if hasAdmin {
		__antithesis_instrumentation__.Notify(12009)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(12010)
	}
	__antithesis_instrumentation__.Notify(12001)

	if restoreStmt.DescriptorCoverage == tree.AllDescriptors {
		__antithesis_instrumentation__.Notify(12011)
		return pgerror.Newf(
			pgcode.InsufficientPrivilege,
			"only users with the admin role are allowed to restore full cluster backups")
	} else {
		__antithesis_instrumentation__.Notify(12012)
	}
	__antithesis_instrumentation__.Notify(12002)

	if restoreStmt.Targets.TenantID.IsSet() {
		__antithesis_instrumentation__.Notify(12013)
		return pgerror.Newf(
			pgcode.InsufficientPrivilege,
			"only users with the admin role can perform RESTORE TENANT")
	} else {
		__antithesis_instrumentation__.Notify(12014)
	}
	__antithesis_instrumentation__.Notify(12003)

	if len(restoreStmt.Targets.Databases) > 0 {
		__antithesis_instrumentation__.Notify(12015)
		hasCreateDB, err := p.HasRoleOption(ctx, roleoption.CREATEDB)
		if err != nil {
			__antithesis_instrumentation__.Notify(12017)
			return err
		} else {
			__antithesis_instrumentation__.Notify(12018)
		}
		__antithesis_instrumentation__.Notify(12016)
		if !hasCreateDB {
			__antithesis_instrumentation__.Notify(12019)
			return pgerror.Newf(
				pgcode.InsufficientPrivilege,
				"only users with the CREATEDB privilege can restore databases")
		} else {
			__antithesis_instrumentation__.Notify(12020)
		}
	} else {
		__antithesis_instrumentation__.Notify(12021)
	}
	__antithesis_instrumentation__.Notify(12004)
	if p.ExecCfg().ExternalIODirConfig.EnableNonAdminImplicitAndArbitraryOutbound {
		__antithesis_instrumentation__.Notify(12022)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(12023)
	}
	__antithesis_instrumentation__.Notify(12005)

	for i := range from {
		__antithesis_instrumentation__.Notify(12024)
		for j := range from[i] {
			__antithesis_instrumentation__.Notify(12025)
			conf, err := cloud.ExternalStorageConfFromURI(from[i][j], p.User())
			if err != nil {
				__antithesis_instrumentation__.Notify(12027)
				return err
			} else {
				__antithesis_instrumentation__.Notify(12028)
			}
			__antithesis_instrumentation__.Notify(12026)
			if !conf.AccessIsWithExplicitAuth() {
				__antithesis_instrumentation__.Notify(12029)
				return pgerror.Newf(
					pgcode.InsufficientPrivilege,
					"only users with the admin role are allowed to RESTORE from the specified %s URI",
					conf.Provider.String())
			} else {
				__antithesis_instrumentation__.Notify(12030)
			}
		}
	}
	__antithesis_instrumentation__.Notify(12006)
	return nil
}

func checkClusterRegions(
	ctx context.Context, p sql.PlanHookState, typesByID map[descpb.ID]*typedesc.Mutable,
) error {
	__antithesis_instrumentation__.Notify(12031)
	regionSet := make(map[catpb.RegionName]struct{})
	for _, typ := range typesByID {
		__antithesis_instrumentation__.Notify(12037)
		typeDesc := typedesc.NewBuilder(typ.TypeDesc()).BuildImmutableType()
		if typeDesc.GetKind() == descpb.TypeDescriptor_MULTIREGION_ENUM {
			__antithesis_instrumentation__.Notify(12038)
			regionNames, err := typeDesc.RegionNames()
			if err != nil {
				__antithesis_instrumentation__.Notify(12040)
				return err
			} else {
				__antithesis_instrumentation__.Notify(12041)
			}
			__antithesis_instrumentation__.Notify(12039)
			for _, region := range regionNames {
				__antithesis_instrumentation__.Notify(12042)
				if _, ok := regionSet[region]; !ok {
					__antithesis_instrumentation__.Notify(12043)
					regionSet[region] = struct{}{}
				} else {
					__antithesis_instrumentation__.Notify(12044)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(12045)
		}
	}
	__antithesis_instrumentation__.Notify(12032)

	if len(regionSet) == 0 {
		__antithesis_instrumentation__.Notify(12046)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(12047)
	}
	__antithesis_instrumentation__.Notify(12033)

	l, err := sql.GetLiveClusterRegions(ctx, p)
	if err != nil {
		__antithesis_instrumentation__.Notify(12048)
		return err
	} else {
		__antithesis_instrumentation__.Notify(12049)
	}
	__antithesis_instrumentation__.Notify(12034)

	missingRegions := make([]string, 0)
	for region := range regionSet {
		__antithesis_instrumentation__.Notify(12050)
		if !l.IsActive(region) {
			__antithesis_instrumentation__.Notify(12051)
			missingRegions = append(missingRegions, string(region))
		} else {
			__antithesis_instrumentation__.Notify(12052)
		}
	}
	__antithesis_instrumentation__.Notify(12035)

	if len(missingRegions) > 0 {
		__antithesis_instrumentation__.Notify(12053)

		sort.Strings(missingRegions)
		mismatchErr := errors.Newf("detected a mismatch in regions between the restore cluster and the backup cluster, "+
			"missing regions detected: %s.", strings.Join(missingRegions, ", "))
		hintsMsg := fmt.Sprintf("there are two ways you can resolve this issue: "+
			"1) update the cluster to which you're restoring to ensure that the regions present on the nodes' "+
			"--locality flags match those present in the backup image, or "+
			"2) restore with the %q option", restoreOptSkipLocalitiesCheck)
		return errors.WithHint(mismatchErr, hintsMsg)
	} else {
		__antithesis_instrumentation__.Notify(12054)
	}
	__antithesis_instrumentation__.Notify(12036)

	return nil
}

func doRestorePlan(
	ctx context.Context,
	restoreStmt *tree.Restore,
	p sql.PlanHookState,
	from [][]string,
	incFrom []string,
	passphrase string,
	kms []string,
	intoDB string,
	newDBName string,
	newTenantID *roachpb.TenantID,
	endTime hlc.Timestamp,
	resultsCh chan<- tree.Datums,
	subdir string,
) error {
	__antithesis_instrumentation__.Notify(12055)
	if len(from) == 0 || func() bool {
		__antithesis_instrumentation__.Notify(12101)
		return len(from[0]) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(12102)
		return errors.New("invalid base backup specified")
	} else {
		__antithesis_instrumentation__.Notify(12103)
	}
	__antithesis_instrumentation__.Notify(12056)

	if subdir != "" && func() bool {
		__antithesis_instrumentation__.Notify(12104)
		return len(from) != 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(12105)
		return errors.Errorf("RESTORE FROM ... IN can only by used against a single collection path (per-locality)")
	} else {
		__antithesis_instrumentation__.Notify(12106)
	}
	__antithesis_instrumentation__.Notify(12057)

	var fullyResolvedSubdir string

	if strings.EqualFold(subdir, latestFileName) {
		__antithesis_instrumentation__.Notify(12107)

		latest, err := readLatestFile(ctx, from[0][0], p.ExecCfg().DistSQLSrv.ExternalStorageFromURI, p.User())
		if err != nil {
			__antithesis_instrumentation__.Notify(12109)
			return err
		} else {
			__antithesis_instrumentation__.Notify(12110)
		}
		__antithesis_instrumentation__.Notify(12108)
		fullyResolvedSubdir = latest
	} else {
		__antithesis_instrumentation__.Notify(12111)
		fullyResolvedSubdir = subdir
	}
	__antithesis_instrumentation__.Notify(12058)

	fullyResolvedBaseDirectory, err := appendPaths(from[0][:], fullyResolvedSubdir)
	if err != nil {
		__antithesis_instrumentation__.Notify(12112)
		return err
	} else {
		__antithesis_instrumentation__.Notify(12113)
	}
	__antithesis_instrumentation__.Notify(12059)

	fullyResolvedIncrementalsDirectory, err := resolveIncrementalsBackupLocation(
		ctx,
		p.User(),
		p.ExecCfg(),
		incFrom,
		from[0],
		fullyResolvedSubdir,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(12114)
		if errors.Is(err, cloud.ErrListingUnsupported) {
			__antithesis_instrumentation__.Notify(12115)
			log.Warningf(ctx, "storage sink %v does not support listing, only resolving the base backup", incFrom)
		} else {
			__antithesis_instrumentation__.Notify(12116)
			return err
		}
	} else {
		__antithesis_instrumentation__.Notify(12117)
	}
	__antithesis_instrumentation__.Notify(12060)

	baseStores := make([]cloud.ExternalStorage, len(fullyResolvedBaseDirectory))
	for i := range fullyResolvedBaseDirectory {
		__antithesis_instrumentation__.Notify(12118)
		store, err := p.ExecCfg().DistSQLSrv.ExternalStorageFromURI(ctx, fullyResolvedBaseDirectory[i], p.User())
		if err != nil {
			__antithesis_instrumentation__.Notify(12120)
			return errors.Wrapf(err, "failed to open backup storage location")
		} else {
			__antithesis_instrumentation__.Notify(12121)
		}
		__antithesis_instrumentation__.Notify(12119)
		defer store.Close()
		baseStores[i] = store
	}
	__antithesis_instrumentation__.Notify(12061)

	var encryption *jobspb.BackupEncryptionOptions
	if restoreStmt.Options.EncryptionPassphrase != nil {
		__antithesis_instrumentation__.Notify(12122)
		opts, err := readEncryptionOptions(ctx, baseStores[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(12124)
			return err
		} else {
			__antithesis_instrumentation__.Notify(12125)
		}
		__antithesis_instrumentation__.Notify(12123)
		encryptionKey := storageccl.GenerateKey([]byte(passphrase), opts[0].Salt)
		encryption = &jobspb.BackupEncryptionOptions{
			Mode: jobspb.EncryptionMode_Passphrase,
			Key:  encryptionKey,
		}
	} else {
		__antithesis_instrumentation__.Notify(12126)
		if restoreStmt.Options.DecryptionKMSURI != nil {
			__antithesis_instrumentation__.Notify(12127)
			opts, err := readEncryptionOptions(ctx, baseStores[0])
			if err != nil {
				__antithesis_instrumentation__.Notify(12131)
				return err
			} else {
				__antithesis_instrumentation__.Notify(12132)
			}
			__antithesis_instrumentation__.Notify(12128)
			ioConf := baseStores[0].ExternalIOConf()

			var defaultKMSInfo *jobspb.BackupEncryptionOptions_KMSInfo
			for _, encFile := range opts {
				__antithesis_instrumentation__.Notify(12133)
				defaultKMSInfo, err = validateKMSURIsAgainstFullBackup(kms,
					newEncryptedDataKeyMapFromProtoMap(encFile.EncryptedDataKeyByKMSMasterKeyID), &backupKMSEnv{
						baseStores[0].Settings(),
						&ioConf,
					})
				if err == nil {
					__antithesis_instrumentation__.Notify(12134)
					break
				} else {
					__antithesis_instrumentation__.Notify(12135)
				}
			}
			__antithesis_instrumentation__.Notify(12129)
			if err != nil {
				__antithesis_instrumentation__.Notify(12136)
				return err
			} else {
				__antithesis_instrumentation__.Notify(12137)
			}
			__antithesis_instrumentation__.Notify(12130)
			encryption = &jobspb.BackupEncryptionOptions{
				Mode:    jobspb.EncryptionMode_KMS,
				KMSInfo: defaultKMSInfo,
			}
		} else {
			__antithesis_instrumentation__.Notify(12138)
		}
	}
	__antithesis_instrumentation__.Notify(12062)

	mem := p.ExecCfg().RootMemoryMonitor.MakeBoundAccount()
	defer mem.Close(ctx)

	var defaultURIs []string
	var mainBackupManifests []BackupManifest
	var localityInfo []jobspb.RestoreDetails_BackupLocalityInfo
	var memReserved int64
	mkStore := p.ExecCfg().DistSQLSrv.ExternalStorageFromURI
	if len(from) <= 1 {
		__antithesis_instrumentation__.Notify(12139)

		defaultURIs, mainBackupManifests, localityInfo, memReserved, err = resolveBackupManifests(
			ctx, &mem, baseStores, mkStore, fullyResolvedBaseDirectory,
			fullyResolvedIncrementalsDirectory, endTime, encryption, p.User(),
		)
	} else {
		__antithesis_instrumentation__.Notify(12140)

		defaultURIs, mainBackupManifests, localityInfo, memReserved, err = resolveBackupManifestsExplicitIncrementals(
			ctx, &mem, mkStore, from, endTime, encryption, p.User())
	}
	__antithesis_instrumentation__.Notify(12063)

	if err != nil {
		__antithesis_instrumentation__.Notify(12141)
		return err
	} else {
		__antithesis_instrumentation__.Notify(12142)
	}
	__antithesis_instrumentation__.Notify(12064)
	defer func() {
		__antithesis_instrumentation__.Notify(12143)
		mem.Shrink(ctx, memReserved)
	}()
	__antithesis_instrumentation__.Notify(12065)

	currentVersion := p.ExecCfg().Settings.Version.ActiveVersion(ctx)
	for i := range mainBackupManifests {
		__antithesis_instrumentation__.Notify(12144)
		if v := mainBackupManifests[i].ClusterVersion; v.Major != 0 {
			__antithesis_instrumentation__.Notify(12145)

			if currentVersion.Less(v) {
				__antithesis_instrumentation__.Notify(12146)
				return errors.Errorf("backup from version %s is newer than current version %s", v, currentVersion)
			} else {
				__antithesis_instrumentation__.Notify(12147)
			}
		} else {
			__antithesis_instrumentation__.Notify(12148)
		}
	}
	__antithesis_instrumentation__.Notify(12066)

	if restoreStmt.DescriptorCoverage == tree.AllDescriptors && func() bool {
		__antithesis_instrumentation__.Notify(12149)
		return mainBackupManifests[0].DescriptorCoverage == tree.RequestedDescriptors == true
	}() == true {
		__antithesis_instrumentation__.Notify(12150)
		return errors.Errorf("full cluster RESTORE can only be used on full cluster BACKUP files")
	} else {
		__antithesis_instrumentation__.Notify(12151)
	}
	__antithesis_instrumentation__.Notify(12067)

	if restoreStmt.DescriptorCoverage == tree.AllDescriptors {
		__antithesis_instrumentation__.Notify(12152)
		var allDescs []catalog.Descriptor
		if err := sql.DescsTxn(ctx, p.ExecCfg(), func(ctx context.Context, txn *kv.Txn, col *descs.Collection) (err error) {
			__antithesis_instrumentation__.Notify(12154)
			txn.SetDebugName("count-user-descs")
			all, err := col.GetAllDescriptors(ctx, txn)
			allDescs = all.OrderedDescriptors()
			return err
		}); err != nil {
			__antithesis_instrumentation__.Notify(12155)
			return errors.Wrap(err, "looking up user descriptors during restore")
		} else {
			__antithesis_instrumentation__.Notify(12156)
		}
		__antithesis_instrumentation__.Notify(12153)
		if allUserDescs := filteredUserCreatedDescriptors(allDescs); len(allUserDescs) > 0 {
			__antithesis_instrumentation__.Notify(12157)
			userDescriptorNames := make([]string, 0, 20)
			for i, desc := range allUserDescs {
				__antithesis_instrumentation__.Notify(12159)
				if i == 20 {
					__antithesis_instrumentation__.Notify(12161)
					userDescriptorNames = append(userDescriptorNames, "...")
					break
				} else {
					__antithesis_instrumentation__.Notify(12162)
				}
				__antithesis_instrumentation__.Notify(12160)
				userDescriptorNames = append(userDescriptorNames, desc.GetName())
			}
			__antithesis_instrumentation__.Notify(12158)
			return errors.Errorf(
				"full cluster restore can only be run on a cluster with no tables or databases but found %d descriptors: %s",
				len(allUserDescs), strings.Join(userDescriptorNames, ", "),
			)
		} else {
			__antithesis_instrumentation__.Notify(12163)
		}
	} else {
		__antithesis_instrumentation__.Notify(12164)
	}
	__antithesis_instrumentation__.Notify(12068)

	wasOffline := make(map[tableAndIndex]hlc.Timestamp)

	for _, m := range mainBackupManifests {
		__antithesis_instrumentation__.Notify(12165)
		spans := roachpb.Spans(m.Spans)
		for i := range m.Descriptors {
			__antithesis_instrumentation__.Notify(12166)
			table, _, _, _ := descpb.FromDescriptor(&m.Descriptors[i])
			if table == nil {
				__antithesis_instrumentation__.Notify(12169)
				continue
			} else {
				__antithesis_instrumentation__.Notify(12170)
			}
			__antithesis_instrumentation__.Notify(12167)
			index := table.GetPrimaryIndex()
			if len(index.Interleave.Ancestors) > 0 || func() bool {
				__antithesis_instrumentation__.Notify(12171)
				return len(index.InterleavedBy) > 0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(12172)
				return errors.Errorf("restoring interleaved tables is no longer allowed. table %s was found to be interleaved", table.Name)
			} else {
				__antithesis_instrumentation__.Notify(12173)
			}
			__antithesis_instrumentation__.Notify(12168)
			if err := catalog.ForEachNonDropIndex(
				tabledesc.NewBuilder(table).BuildImmutable().(catalog.TableDescriptor),
				func(index catalog.Index) error {
					__antithesis_instrumentation__.Notify(12174)
					if index.Adding() && func() bool {
						__antithesis_instrumentation__.Notify(12176)
						return spans.ContainsKey(p.ExecCfg().Codec.IndexPrefix(uint32(table.ID), uint32(index.GetID()))) == true
					}() == true {
						__antithesis_instrumentation__.Notify(12177)
						k := tableAndIndex{tableID: table.ID, indexID: index.GetID()}
						if _, ok := wasOffline[k]; !ok {
							__antithesis_instrumentation__.Notify(12178)
							wasOffline[k] = m.EndTime
						} else {
							__antithesis_instrumentation__.Notify(12179)
						}
					} else {
						__antithesis_instrumentation__.Notify(12180)
					}
					__antithesis_instrumentation__.Notify(12175)
					return nil
				}); err != nil {
				__antithesis_instrumentation__.Notify(12181)
				return err
			} else {
				__antithesis_instrumentation__.Notify(12182)
			}
		}
	}
	__antithesis_instrumentation__.Notify(12069)

	sqlDescs, restoreDBs, tenants, err := selectTargets(
		ctx, p, mainBackupManifests, restoreStmt.Targets, restoreStmt.DescriptorCoverage, endTime, restoreStmt.SystemUsers,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(12183)
		return errors.Wrap(err,
			"failed to resolve targets in the BACKUP location specified by the RESTORE stmt, "+
				"use SHOW BACKUP to find correct targets")
	} else {
		__antithesis_instrumentation__.Notify(12184)
	}
	__antithesis_instrumentation__.Notify(12070)

	var revalidateIndexes []jobspb.RestoreDetails_RevalidateIndex
	for _, desc := range sqlDescs {
		__antithesis_instrumentation__.Notify(12185)
		tbl, ok := desc.(catalog.TableDescriptor)
		if !ok {
			__antithesis_instrumentation__.Notify(12187)
			continue
		} else {
			__antithesis_instrumentation__.Notify(12188)
		}
		__antithesis_instrumentation__.Notify(12186)
		for _, idx := range tbl.ActiveIndexes() {
			__antithesis_instrumentation__.Notify(12189)
			if _, ok := wasOffline[tableAndIndex{tableID: desc.GetID(), indexID: idx.GetID()}]; ok {
				__antithesis_instrumentation__.Notify(12190)
				revalidateIndexes = append(revalidateIndexes, jobspb.RestoreDetails_RevalidateIndex{
					TableID: desc.GetID(), IndexID: idx.GetID(),
				})
			} else {
				__antithesis_instrumentation__.Notify(12191)
			}
		}
	}
	__antithesis_instrumentation__.Notify(12071)

	err = ensureMultiRegionDatabaseRestoreIsAllowed(p, restoreDBs)
	if err != nil {
		__antithesis_instrumentation__.Notify(12192)
		return err
	} else {
		__antithesis_instrumentation__.Notify(12193)
	}
	__antithesis_instrumentation__.Notify(12072)

	databaseModifiers, newTypeDescs, err := planDatabaseModifiersForRestore(ctx, p, sqlDescs, restoreDBs)
	if err != nil {
		__antithesis_instrumentation__.Notify(12194)
		return err
	} else {
		__antithesis_instrumentation__.Notify(12195)
	}
	__antithesis_instrumentation__.Notify(12073)

	sqlDescs = append(sqlDescs, newTypeDescs...)

	if err := maybeUpgradeDescriptors(sqlDescs, restoreStmt.Options.SkipMissingFKs); err != nil {
		__antithesis_instrumentation__.Notify(12196)
		return err
	} else {
		__antithesis_instrumentation__.Notify(12197)
	}
	__antithesis_instrumentation__.Notify(12074)

	if restoreStmt.Options.NewDBName != nil {
		__antithesis_instrumentation__.Notify(12198)
		if err := renameTargetDatabaseDescriptor(sqlDescs, restoreDBs, newDBName); err != nil {
			__antithesis_instrumentation__.Notify(12199)
			return err
		} else {
			__antithesis_instrumentation__.Notify(12200)
		}
	} else {
		__antithesis_instrumentation__.Notify(12201)
	}
	__antithesis_instrumentation__.Notify(12075)

	var oldTenantID *roachpb.TenantID
	if len(tenants) > 0 {
		__antithesis_instrumentation__.Notify(12202)
		if !p.ExecCfg().Codec.ForSystemTenant() {
			__antithesis_instrumentation__.Notify(12204)
			return pgerror.Newf(pgcode.InsufficientPrivilege, "only the system tenant can restore other tenants")
		} else {
			__antithesis_instrumentation__.Notify(12205)
		}
		__antithesis_instrumentation__.Notify(12203)
		if newTenantID != nil {
			__antithesis_instrumentation__.Notify(12206)
			if len(tenants) != 1 {
				__antithesis_instrumentation__.Notify(12210)
				return errors.Errorf("%q option can only be used when restoring a single tenant", restoreOptAsTenant)
			} else {
				__antithesis_instrumentation__.Notify(12211)
			}
			__antithesis_instrumentation__.Notify(12207)
			res, err := p.ExecCfg().InternalExecutor.QueryRow(
				ctx, "restore-lookup-tenant", p.ExtendedEvalContext().Txn,
				`SELECT active FROM system.tenants WHERE id = $1`, newTenantID.ToUint64(),
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(12212)
				return err
			} else {
				__antithesis_instrumentation__.Notify(12213)
			}
			__antithesis_instrumentation__.Notify(12208)
			if res != nil {
				__antithesis_instrumentation__.Notify(12214)
				return errors.Errorf("tenant %s already exists", newTenantID)
			} else {
				__antithesis_instrumentation__.Notify(12215)
			}
			__antithesis_instrumentation__.Notify(12209)
			old := roachpb.MakeTenantID(tenants[0].ID)
			tenants[0].ID = newTenantID.ToUint64()
			oldTenantID = &old
		} else {
			__antithesis_instrumentation__.Notify(12216)
			for _, i := range tenants {
				__antithesis_instrumentation__.Notify(12217)
				res, err := p.ExecCfg().InternalExecutor.QueryRow(
					ctx, "restore-lookup-tenant", p.ExtendedEvalContext().Txn,
					`SELECT active FROM system.tenants WHERE id = $1`, i.ID,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(12219)
					return err
				} else {
					__antithesis_instrumentation__.Notify(12220)
				}
				__antithesis_instrumentation__.Notify(12218)
				if res != nil {
					__antithesis_instrumentation__.Notify(12221)
					return errors.Errorf("tenant %d already exists", i.ID)
				} else {
					__antithesis_instrumentation__.Notify(12222)
				}
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(12223)
	}
	__antithesis_instrumentation__.Notify(12076)

	databasesByID := make(map[descpb.ID]*dbdesc.Mutable)
	schemasByID := make(map[descpb.ID]*schemadesc.Mutable)
	tablesByID := make(map[descpb.ID]*tabledesc.Mutable)
	typesByID := make(map[descpb.ID]*typedesc.Mutable)

	for _, desc := range sqlDescs {
		__antithesis_instrumentation__.Notify(12224)
		switch desc := desc.(type) {
		case *dbdesc.Mutable:
			__antithesis_instrumentation__.Notify(12225)
			databasesByID[desc.GetID()] = desc
		case *schemadesc.Mutable:
			__antithesis_instrumentation__.Notify(12226)
			schemasByID[desc.ID] = desc
		case *tabledesc.Mutable:
			__antithesis_instrumentation__.Notify(12227)
			tablesByID[desc.ID] = desc
		case *typedesc.Mutable:
			__antithesis_instrumentation__.Notify(12228)
			typesByID[desc.ID] = desc
		}
	}
	__antithesis_instrumentation__.Notify(12077)

	if !restoreStmt.Options.SkipLocalitiesCheck {
		__antithesis_instrumentation__.Notify(12229)
		if err := checkClusterRegions(ctx, p, typesByID); err != nil {
			__antithesis_instrumentation__.Notify(12230)
			return err
		} else {
			__antithesis_instrumentation__.Notify(12231)
		}
	} else {
		__antithesis_instrumentation__.Notify(12232)
	}
	__antithesis_instrumentation__.Notify(12078)

	var debugPauseOn string
	if restoreStmt.Options.DebugPauseOn != nil {
		__antithesis_instrumentation__.Notify(12233)
		pauseOnFn, err := p.TypeAsString(ctx, restoreStmt.Options.DebugPauseOn, "RESTORE")
		if err != nil {
			__antithesis_instrumentation__.Notify(12236)
			return err
		} else {
			__antithesis_instrumentation__.Notify(12237)
		}
		__antithesis_instrumentation__.Notify(12234)

		debugPauseOn, err = pauseOnFn()
		if err != nil {
			__antithesis_instrumentation__.Notify(12238)
			return err
		} else {
			__antithesis_instrumentation__.Notify(12239)
		}
		__antithesis_instrumentation__.Notify(12235)

		if _, ok := allowedDebugPauseOnValues[debugPauseOn]; len(debugPauseOn) > 0 && func() bool {
			__antithesis_instrumentation__.Notify(12240)
			return !ok == true
		}() == true {
			__antithesis_instrumentation__.Notify(12241)
			return errors.Newf("%s cannot be set with the value %s", restoreOptDebugPauseOn, debugPauseOn)
		} else {
			__antithesis_instrumentation__.Notify(12242)
		}
	} else {
		__antithesis_instrumentation__.Notify(12243)
	}
	__antithesis_instrumentation__.Notify(12079)

	filteredTablesByID, err := maybeFilterMissingViews(
		tablesByID,
		typesByID,
		restoreStmt.Options.SkipMissingViews)
	if err != nil {
		__antithesis_instrumentation__.Notify(12244)
		return err
	} else {
		__antithesis_instrumentation__.Notify(12245)
	}
	__antithesis_instrumentation__.Notify(12080)

	if restoreStmt.DescriptorCoverage == tree.AllDescriptors {
		__antithesis_instrumentation__.Notify(12246)
		if err := dropDefaultUserDBs(ctx, p.ExecCfg()); err != nil {
			__antithesis_instrumentation__.Notify(12247)
			return err
		} else {
			__antithesis_instrumentation__.Notify(12248)
		}
	} else {
		__antithesis_instrumentation__.Notify(12249)
	}
	__antithesis_instrumentation__.Notify(12081)

	descriptorRewrites, err := allocateDescriptorRewrites(
		ctx,
		p,
		databasesByID,
		schemasByID,
		filteredTablesByID,
		typesByID,
		restoreDBs,
		restoreStmt.DescriptorCoverage,
		restoreStmt.Options,
		intoDB,
		newDBName,
		restoreStmt.SystemUsers)
	if err != nil {
		__antithesis_instrumentation__.Notify(12250)
		return err
	} else {
		__antithesis_instrumentation__.Notify(12251)
	}
	__antithesis_instrumentation__.Notify(12082)
	var fromDescription [][]string
	if len(from) == 1 {
		__antithesis_instrumentation__.Notify(12252)
		fromDescription = [][]string{fullyResolvedBaseDirectory}
	} else {
		__antithesis_instrumentation__.Notify(12253)
		fromDescription = from
	}
	__antithesis_instrumentation__.Notify(12083)
	description, err := restoreJobDescription(
		p,
		restoreStmt,
		fromDescription,
		fullyResolvedIncrementalsDirectory,
		restoreStmt.Options,
		intoDB,
		newDBName,
		kms)
	if err != nil {
		__antithesis_instrumentation__.Notify(12254)
		return err
	} else {
		__antithesis_instrumentation__.Notify(12255)
	}
	__antithesis_instrumentation__.Notify(12084)

	var databases []*dbdesc.Mutable
	for i := range databasesByID {
		__antithesis_instrumentation__.Notify(12256)
		if _, ok := descriptorRewrites[i]; ok {
			__antithesis_instrumentation__.Notify(12257)
			databases = append(databases, databasesByID[i])
		} else {
			__antithesis_instrumentation__.Notify(12258)
		}
	}
	__antithesis_instrumentation__.Notify(12085)
	var schemas []*schemadesc.Mutable
	for i := range schemasByID {
		__antithesis_instrumentation__.Notify(12259)
		schemas = append(schemas, schemasByID[i])
	}
	__antithesis_instrumentation__.Notify(12086)
	var tables []*tabledesc.Mutable
	for _, desc := range filteredTablesByID {
		__antithesis_instrumentation__.Notify(12260)
		tables = append(tables, desc)
	}
	__antithesis_instrumentation__.Notify(12087)
	var types []*typedesc.Mutable
	for _, desc := range typesByID {
		__antithesis_instrumentation__.Notify(12261)
		types = append(types, desc)
	}
	__antithesis_instrumentation__.Notify(12088)

	if err := rewrite.TableDescs(tables, descriptorRewrites, intoDB); err != nil {
		__antithesis_instrumentation__.Notify(12262)
		return err
	} else {
		__antithesis_instrumentation__.Notify(12263)
	}
	__antithesis_instrumentation__.Notify(12089)
	if err := rewrite.DatabaseDescs(databases, descriptorRewrites); err != nil {
		__antithesis_instrumentation__.Notify(12264)
		return err
	} else {
		__antithesis_instrumentation__.Notify(12265)
	}
	__antithesis_instrumentation__.Notify(12090)
	if err := rewrite.SchemaDescs(schemas, descriptorRewrites); err != nil {
		__antithesis_instrumentation__.Notify(12266)
		return err
	} else {
		__antithesis_instrumentation__.Notify(12267)
	}
	__antithesis_instrumentation__.Notify(12091)
	if err := rewrite.TypeDescs(types, descriptorRewrites); err != nil {
		__antithesis_instrumentation__.Notify(12268)
		return err
	} else {
		__antithesis_instrumentation__.Notify(12269)
	}
	__antithesis_instrumentation__.Notify(12092)
	for i := range revalidateIndexes {
		__antithesis_instrumentation__.Notify(12270)
		revalidateIndexes[i].TableID = descriptorRewrites[revalidateIndexes[i].TableID].ID
	}
	__antithesis_instrumentation__.Notify(12093)

	collectTelemetry := func() {
		__antithesis_instrumentation__.Notify(12271)
		telemetry.Count("restore.total.started")
		if restoreStmt.DescriptorCoverage == tree.AllDescriptors {
			__antithesis_instrumentation__.Notify(12273)
			telemetry.Count("restore.full-cluster")
		} else {
			__antithesis_instrumentation__.Notify(12274)
		}
		__antithesis_instrumentation__.Notify(12272)
		if restoreStmt.Subdir == nil {
			__antithesis_instrumentation__.Notify(12275)
			telemetry.Count("restore.deprecated-subdir-syntax")
		} else {
			__antithesis_instrumentation__.Notify(12276)
			telemetry.Count("restore.collection")
		}
	}
	__antithesis_instrumentation__.Notify(12094)

	encodedTables := make([]*descpb.TableDescriptor, len(tables))
	for i, table := range tables {
		__antithesis_instrumentation__.Notify(12277)
		encodedTables[i] = table.TableDesc()
	}
	__antithesis_instrumentation__.Notify(12095)

	jr := jobs.Record{
		Description: description,
		Username:    p.User(),
		DescriptorIDs: func() (sqlDescIDs []descpb.ID) {
			__antithesis_instrumentation__.Notify(12278)
			for _, tableRewrite := range descriptorRewrites {
				__antithesis_instrumentation__.Notify(12280)
				sqlDescIDs = append(sqlDescIDs, tableRewrite.ID)
			}
			__antithesis_instrumentation__.Notify(12279)
			return sqlDescIDs
		}(),
		Details: jobspb.RestoreDetails{
			EndTime:            endTime,
			DescriptorRewrites: descriptorRewrites,
			URIs:               defaultURIs,
			BackupLocalityInfo: localityInfo,
			TableDescs:         encodedTables,
			Tenants:            tenants,
			OverrideDB:         intoDB,
			DescriptorCoverage: restoreStmt.DescriptorCoverage,
			Encryption:         encryption,
			RevalidateIndexes:  revalidateIndexes,
			DatabaseModifiers:  databaseModifiers,
			DebugPauseOn:       debugPauseOn,
			RestoreSystemUsers: restoreStmt.SystemUsers,
			PreRewriteTenantId: oldTenantID,
			Validation:         jobspb.RestoreValidation_DefaultRestore,
		},
		Progress: jobspb.RestoreProgress{},
	}
	__antithesis_instrumentation__.Notify(12096)

	if restoreStmt.Options.Detached {
		__antithesis_instrumentation__.Notify(12281)

		jobID := p.ExecCfg().JobRegistry.MakeJobID()
		_, err := p.ExecCfg().JobRegistry.CreateAdoptableJobWithTxn(
			ctx, jr, jobID, p.ExtendedEvalContext().Txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(12283)
			return err
		} else {
			__antithesis_instrumentation__.Notify(12284)
		}
		__antithesis_instrumentation__.Notify(12282)
		resultsCh <- tree.Datums{tree.NewDInt(tree.DInt(jobID))}
		collectTelemetry()
		return nil
	} else {
		__antithesis_instrumentation__.Notify(12285)
	}
	__antithesis_instrumentation__.Notify(12097)

	plannerTxn := p.ExtendedEvalContext().Txn

	var sj *jobs.StartableJob
	if err := func() (err error) {
		__antithesis_instrumentation__.Notify(12286)
		defer func() {
			__antithesis_instrumentation__.Notify(12289)
			if err == nil || func() bool {
				__antithesis_instrumentation__.Notify(12291)
				return sj == nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(12292)
				return
			} else {
				__antithesis_instrumentation__.Notify(12293)
			}
			__antithesis_instrumentation__.Notify(12290)
			if cleanupErr := sj.CleanupOnRollback(ctx); cleanupErr != nil {
				__antithesis_instrumentation__.Notify(12294)
				log.Errorf(ctx, "failed to cleanup job: %v", cleanupErr)
			} else {
				__antithesis_instrumentation__.Notify(12295)
			}
		}()
		__antithesis_instrumentation__.Notify(12287)
		jobID := p.ExecCfg().JobRegistry.MakeJobID()
		if err := p.ExecCfg().JobRegistry.CreateStartableJobWithTxn(ctx, &sj, jobID, plannerTxn, jr); err != nil {
			__antithesis_instrumentation__.Notify(12296)
			return err
		} else {
			__antithesis_instrumentation__.Notify(12297)
		}
		__antithesis_instrumentation__.Notify(12288)

		return plannerTxn.Commit(ctx)
	}(); err != nil {
		__antithesis_instrumentation__.Notify(12298)
		return err
	} else {
		__antithesis_instrumentation__.Notify(12299)
	}
	__antithesis_instrumentation__.Notify(12098)
	collectTelemetry()
	if err := sj.Start(ctx); err != nil {
		__antithesis_instrumentation__.Notify(12300)
		return err
	} else {
		__antithesis_instrumentation__.Notify(12301)
	}
	__antithesis_instrumentation__.Notify(12099)
	if err := sj.AwaitCompletion(ctx); err != nil {
		__antithesis_instrumentation__.Notify(12302)
		return err
	} else {
		__antithesis_instrumentation__.Notify(12303)
	}
	__antithesis_instrumentation__.Notify(12100)
	return sj.ReportExecutionResults(ctx, resultsCh)
}

func filteredUserCreatedDescriptors(
	allDescs []catalog.Descriptor,
) (userDescs []catalog.Descriptor) {
	__antithesis_instrumentation__.Notify(12304)
	defaultDBs := make(map[descpb.ID]catalog.DatabaseDescriptor, len(catalogkeys.DefaultUserDBs))
	{
		__antithesis_instrumentation__.Notify(12307)
		names := make(map[string]struct{}, len(catalogkeys.DefaultUserDBs))
		for _, name := range catalogkeys.DefaultUserDBs {
			__antithesis_instrumentation__.Notify(12309)
			names[name] = struct{}{}
		}
		__antithesis_instrumentation__.Notify(12308)
		for _, desc := range allDescs {
			__antithesis_instrumentation__.Notify(12310)
			if db, ok := desc.(catalog.DatabaseDescriptor); ok {
				__antithesis_instrumentation__.Notify(12311)
				if _, found := names[desc.GetName()]; found {
					__antithesis_instrumentation__.Notify(12312)
					defaultDBs[db.GetID()] = db
				} else {
					__antithesis_instrumentation__.Notify(12313)
				}
			} else {
				__antithesis_instrumentation__.Notify(12314)
			}
		}
	}
	__antithesis_instrumentation__.Notify(12305)

	userDescs = make([]catalog.Descriptor, 0, len(allDescs))
	for _, desc := range allDescs {
		__antithesis_instrumentation__.Notify(12315)
		if catalog.IsSystemDescriptor(desc) {
			__antithesis_instrumentation__.Notify(12319)

			continue
		} else {
			__antithesis_instrumentation__.Notify(12320)
		}
		__antithesis_instrumentation__.Notify(12316)
		if _, found := defaultDBs[desc.GetID()]; found {
			__antithesis_instrumentation__.Notify(12321)

			continue
		} else {
			__antithesis_instrumentation__.Notify(12322)
		}
		__antithesis_instrumentation__.Notify(12317)
		if sc, ok := desc.(catalog.SchemaDescriptor); ok && func() bool {
			__antithesis_instrumentation__.Notify(12323)
			return sc.GetName() == catconstants.PublicSchemaName == true
		}() == true {
			__antithesis_instrumentation__.Notify(12324)
			if _, found := defaultDBs[sc.GetParentID()]; found {
				__antithesis_instrumentation__.Notify(12325)

				continue
			} else {
				__antithesis_instrumentation__.Notify(12326)
			}
		} else {
			__antithesis_instrumentation__.Notify(12327)
		}
		__antithesis_instrumentation__.Notify(12318)
		userDescs = append(userDescs, desc)
	}
	__antithesis_instrumentation__.Notify(12306)
	return userDescs
}

func renameTargetDatabaseDescriptor(
	sqlDescs []catalog.Descriptor, restoreDBs []catalog.DatabaseDescriptor, newDBName string,
) error {
	__antithesis_instrumentation__.Notify(12328)
	if len(restoreDBs) != 1 {
		__antithesis_instrumentation__.Notify(12332)
		return errors.NewAssertionErrorWithWrappedErrf(errors.Newf(
			"expected restoreDBs to have a single entry but found %d entries when renaming the target"+
				" database", len(restoreDBs)), "assertion failed")
	} else {
		__antithesis_instrumentation__.Notify(12333)
	}
	__antithesis_instrumentation__.Notify(12329)

	for _, desc := range sqlDescs {
		__antithesis_instrumentation__.Notify(12334)

		db, isDB := desc.(*dbdesc.Mutable)
		if !isDB {
			__antithesis_instrumentation__.Notify(12336)
			continue
		} else {
			__antithesis_instrumentation__.Notify(12337)
		}
		__antithesis_instrumentation__.Notify(12335)
		db.SetName(newDBName)
	}
	__antithesis_instrumentation__.Notify(12330)
	db, ok := restoreDBs[0].(*dbdesc.Mutable)
	if !ok {
		__antithesis_instrumentation__.Notify(12338)
		return errors.NewAssertionErrorWithWrappedErrf(errors.Newf(
			"expected db desc but found %T", db), "assertion failed")
	} else {
		__antithesis_instrumentation__.Notify(12339)
	}
	__antithesis_instrumentation__.Notify(12331)
	db.SetName(newDBName)
	return nil
}

func ensureMultiRegionDatabaseRestoreIsAllowed(
	p sql.PlanHookState, restoreDBs []catalog.DatabaseDescriptor,
) error {
	__antithesis_instrumentation__.Notify(12340)
	if p.ExecCfg().Codec.ForSystemTenant() || func() bool {
		__antithesis_instrumentation__.Notify(12343)
		return sql.SecondaryTenantsMultiRegionAbstractionsEnabled.Get(&p.ExecCfg().Settings.SV) == true
	}() == true {
		__antithesis_instrumentation__.Notify(12344)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(12345)
	}
	__antithesis_instrumentation__.Notify(12341)
	for _, dbDesc := range restoreDBs {
		__antithesis_instrumentation__.Notify(12346)

		if dbDesc.IsMultiRegion() {
			__antithesis_instrumentation__.Notify(12347)
			return pgerror.Newf(
				pgcode.InvalidDatabaseDefinition,
				"setting %s disallows secondary tenant to restore a multi-region database",
				sql.SecondaryTenantsMultiRegionAbstractionsEnabledSettingName,
			)
		} else {
			__antithesis_instrumentation__.Notify(12348)
		}
	}
	__antithesis_instrumentation__.Notify(12342)

	return nil
}

func planDatabaseModifiersForRestore(
	ctx context.Context,
	p sql.PlanHookState,
	sqlDescs []catalog.Descriptor,
	restoreDBs []catalog.DatabaseDescriptor,
) (map[descpb.ID]*jobspb.RestoreDetails_DatabaseModifier, []catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(12349)
	databaseModifiers := make(map[descpb.ID]*jobspb.RestoreDetails_DatabaseModifier)
	defaultPrimaryRegion := catpb.RegionName(
		sql.DefaultPrimaryRegion.Get(&p.ExecCfg().Settings.SV),
	)
	if !p.ExecCfg().Codec.ForSystemTenant() && func() bool {
		__antithesis_instrumentation__.Notify(12359)
		return !sql.SecondaryTenantsMultiRegionAbstractionsEnabled.Get(&p.ExecCfg().Settings.SV) == true
	}() == true {
		__antithesis_instrumentation__.Notify(12360)

		return nil, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(12361)
	}
	__antithesis_instrumentation__.Notify(12350)
	if defaultPrimaryRegion == "" {
		__antithesis_instrumentation__.Notify(12362)
		return nil, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(12363)
	}
	__antithesis_instrumentation__.Notify(12351)
	if err := multiregionccl.CheckClusterSupportsMultiRegion(p.ExecCfg()); err != nil {
		__antithesis_instrumentation__.Notify(12364)
		return nil, nil, errors.WithHintf(
			err,
			"try disabling the default PRIMARY REGION by using RESET CLUSTER SETTING %s",
			sql.DefaultPrimaryRegionClusterSettingName,
		)
	} else {
		__antithesis_instrumentation__.Notify(12365)
	}
	__antithesis_instrumentation__.Notify(12352)

	l, err := sql.GetLiveClusterRegions(ctx, p)
	if err != nil {
		__antithesis_instrumentation__.Notify(12366)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(12367)
	}
	__antithesis_instrumentation__.Notify(12353)
	if err := sql.CheckClusterRegionIsLive(
		l,
		defaultPrimaryRegion,
	); err != nil {
		__antithesis_instrumentation__.Notify(12368)
		return nil, nil, errors.WithHintf(
			err,
			"set the default PRIMARY REGION to a region that exists (see SHOW REGIONS FROM CLUSTER) then using SET CLUSTER SETTING %s = 'region'",
			sql.DefaultPrimaryRegionClusterSettingName,
		)
	} else {
		__antithesis_instrumentation__.Notify(12369)
	}
	__antithesis_instrumentation__.Notify(12354)

	var maxSeenID descpb.ID
	for _, desc := range sqlDescs {
		__antithesis_instrumentation__.Notify(12370)
		if desc.GetID() > maxSeenID {
			__antithesis_instrumentation__.Notify(12371)
			maxSeenID = desc.GetID()
		} else {
			__antithesis_instrumentation__.Notify(12372)
		}
	}
	__antithesis_instrumentation__.Notify(12355)

	shouldRestoreDatabaseIDs := make(map[descpb.ID]struct{})
	for _, db := range restoreDBs {
		__antithesis_instrumentation__.Notify(12373)
		shouldRestoreDatabaseIDs[db.GetID()] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(12356)

	var extraDescs []catalog.Descriptor

	dbToPublicSchema := make(map[descpb.ID]catalog.SchemaDescriptor)
	for _, desc := range sqlDescs {
		__antithesis_instrumentation__.Notify(12374)
		sc, isSc := desc.(*schemadesc.Mutable)
		if !isSc {
			__antithesis_instrumentation__.Notify(12376)
			continue
		} else {
			__antithesis_instrumentation__.Notify(12377)
		}
		__antithesis_instrumentation__.Notify(12375)
		if sc.GetName() == tree.PublicSchema {
			__antithesis_instrumentation__.Notify(12378)
			dbToPublicSchema[sc.GetParentID()] = sc
		} else {
			__antithesis_instrumentation__.Notify(12379)
		}
	}
	__antithesis_instrumentation__.Notify(12357)
	for _, desc := range sqlDescs {
		__antithesis_instrumentation__.Notify(12380)

		db, isDB := desc.(*dbdesc.Mutable)
		if !isDB {
			__antithesis_instrumentation__.Notify(12389)
			continue
		} else {
			__antithesis_instrumentation__.Notify(12390)
		}
		__antithesis_instrumentation__.Notify(12381)

		if _, ok := shouldRestoreDatabaseIDs[db.GetID()]; !ok {
			__antithesis_instrumentation__.Notify(12391)
			continue
		} else {
			__antithesis_instrumentation__.Notify(12392)
		}
		__antithesis_instrumentation__.Notify(12382)

		if db.IsMultiRegion() {
			__antithesis_instrumentation__.Notify(12393)
			continue
		} else {
			__antithesis_instrumentation__.Notify(12394)
		}
		__antithesis_instrumentation__.Notify(12383)
		p.BufferClientNotice(
			ctx,
			errors.WithHintf(
				pgnotice.Newf(
					"setting the PRIMARY REGION as %s on database %s",
					defaultPrimaryRegion,
					db.GetName(),
				),
				"to change the default primary region, use SET CLUSTER SETTING %[1]s = 'region' "+
					"or use RESET CLUSTER SETTING %[1]s to disable this behavior",
				sql.DefaultPrimaryRegionClusterSettingName,
			),
		)

		regionEnumID := maxSeenID + 1
		regionEnumArrayID := maxSeenID + 2
		maxSeenID += 2

		sg, err := sql.TranslateSurvivalGoal(tree.SurvivalGoalDefault)
		if err != nil {
			__antithesis_instrumentation__.Notify(12395)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(12396)
		}
		__antithesis_instrumentation__.Notify(12384)
		regionConfig := multiregion.MakeRegionConfig(
			[]catpb.RegionName{defaultPrimaryRegion},
			defaultPrimaryRegion,
			sg,
			regionEnumID,
			descpb.DataPlacement_DEFAULT,
			nil,
			descpb.ZoneConfigExtensions{},
		)
		if err := multiregion.ValidateRegionConfig(regionConfig); err != nil {
			__antithesis_instrumentation__.Notify(12397)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(12398)
		}
		__antithesis_instrumentation__.Notify(12385)
		if err := db.SetInitialMultiRegionConfig(&regionConfig); err != nil {
			__antithesis_instrumentation__.Notify(12399)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(12400)
		}
		__antithesis_instrumentation__.Notify(12386)

		var sc catalog.SchemaDescriptor
		if db.HasPublicSchemaWithDescriptor() {
			__antithesis_instrumentation__.Notify(12401)
			sc = dbToPublicSchema[db.GetID()]
		} else {
			__antithesis_instrumentation__.Notify(12402)
			sc = schemadesc.GetPublicSchema()
		}
		__antithesis_instrumentation__.Notify(12387)
		regionEnum, regionArrayEnum, err := restoreCreateDefaultPrimaryRegionEnums(
			ctx,
			p,
			db,
			sc,
			regionConfig,
			regionEnumID,
			regionEnumArrayID,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(12403)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(12404)
		}
		__antithesis_instrumentation__.Notify(12388)

		extraDescs = append(extraDescs, regionEnum, regionArrayEnum)
		databaseModifiers[db.GetID()] = &jobspb.RestoreDetails_DatabaseModifier{
			ExtraTypeDescs: []*descpb.TypeDescriptor{
				&regionEnum.TypeDescriptor,
				&regionArrayEnum.TypeDescriptor,
			},
			RegionConfig: db.GetRegionConfig(),
		}
	}
	__antithesis_instrumentation__.Notify(12358)
	return databaseModifiers, extraDescs, nil
}

func restoreCreateDefaultPrimaryRegionEnums(
	ctx context.Context,
	p sql.PlanHookState,
	db *dbdesc.Mutable,
	sc catalog.SchemaDescriptor,
	regionConfig multiregion.RegionConfig,
	regionEnumID descpb.ID,
	regionEnumArrayID descpb.ID,
) (*typedesc.Mutable, *typedesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(12405)
	regionLabels := make(tree.EnumValueList, 0, len(regionConfig.Regions()))
	for _, regionName := range regionConfig.Regions() {
		__antithesis_instrumentation__.Notify(12409)
		regionLabels = append(regionLabels, tree.EnumValue(regionName))
	}
	__antithesis_instrumentation__.Notify(12406)
	regionEnum, err := sql.CreateEnumTypeDesc(
		p.RunParams(ctx),
		regionConfig.RegionEnumID(),
		regionLabels,
		db,
		sc,
		tree.NewQualifiedTypeName(db.GetName(), tree.PublicSchema, tree.RegionEnum),
		sql.EnumTypeMultiRegion,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(12410)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(12411)
	}
	__antithesis_instrumentation__.Notify(12407)
	regionEnum.ArrayTypeID = regionEnumArrayID
	regionArrayEnum, err := sql.CreateEnumArrayTypeDesc(
		p.RunParams(ctx),
		regionEnum,
		db,
		sc.GetID(),
		regionEnumArrayID,
		"_"+tree.RegionEnum,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(12412)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(12413)
	}
	__antithesis_instrumentation__.Notify(12408)
	return regionEnum, regionArrayEnum, nil
}

func init() {
	sql.AddPlanHook("restore", restorePlanHook)
}
