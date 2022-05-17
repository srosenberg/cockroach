package scbuildstmt

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemaexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/catid"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
)

func alterTableAddColumn(
	b BuildCtx, tn *tree.TableName, tbl *scpb.Table, t *tree.AlterTableAddColumn,
) {
	__antithesis_instrumentation__.Notify(579738)
	b.IncrementSchemaChangeAlterCounter("table", "add_column")
	d := t.ColumnDef

	{
		__antithesis_instrumentation__.Notify(579749)
		elts := b.ResolveColumn(tbl.TableID, d.Name, ResolveParams{
			IsExistenceOptional: true,
			RequiredPrivilege:   privilege.CREATE,
		})
		_, _, col := scpb.FindColumn(elts)
		if col != nil {
			__antithesis_instrumentation__.Notify(579750)
			if t.IfNotExists {
				__antithesis_instrumentation__.Notify(579752)
				return
			} else {
				__antithesis_instrumentation__.Notify(579753)
			}
			__antithesis_instrumentation__.Notify(579751)
			panic(sqlerrors.NewColumnAlreadyExistsError(string(d.Name), tn.Object()))
		} else {
			__antithesis_instrumentation__.Notify(579754)
		}
	}
	__antithesis_instrumentation__.Notify(579739)
	if d.IsSerial {
		__antithesis_instrumentation__.Notify(579755)
		panic(scerrors.NotImplementedErrorf(d, "contains serial data type"))
	} else {
		__antithesis_instrumentation__.Notify(579756)
	}
	__antithesis_instrumentation__.Notify(579740)
	if d.IsComputed() {
		__antithesis_instrumentation__.Notify(579757)
		d.Computed.Expr = schemaexpr.MaybeRewriteComputedColumn(d.Computed.Expr, b.SessionData())
	} else {
		__antithesis_instrumentation__.Notify(579758)
	}
	__antithesis_instrumentation__.Notify(579741)

	if d.Unique.IsUnique {
		__antithesis_instrumentation__.Notify(579759)
		panic(scerrors.NotImplementedErrorf(d, "contains unique constraint"))
	} else {
		__antithesis_instrumentation__.Notify(579760)
	}
	__antithesis_instrumentation__.Notify(579742)
	cdd, err := tabledesc.MakeColumnDefDescs(b, d, b.SemaCtx(), b.EvalCtx())
	if err != nil {
		__antithesis_instrumentation__.Notify(579761)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(579762)
	}
	__antithesis_instrumentation__.Notify(579743)
	desc := cdd.ColumnDescriptor
	spec := addColumnSpec{
		tbl: tbl,
		col: &scpb.Column{
			TableID:                 tbl.TableID,
			ColumnID:                b.NextTableColumnID(tbl),
			IsHidden:                desc.Hidden,
			IsInaccessible:          desc.Inaccessible,
			GeneratedAsIdentityType: desc.GeneratedAsIdentityType,
			PgAttributeNum:          desc.PGAttributeNum,
		},
	}
	if ptr := desc.GeneratedAsIdentitySequenceOption; ptr != nil {
		__antithesis_instrumentation__.Notify(579763)
		spec.col.GeneratedAsIdentitySequenceOption = *ptr
	} else {
		__antithesis_instrumentation__.Notify(579764)
	}
	__antithesis_instrumentation__.Notify(579744)
	spec.name = &scpb.ColumnName{
		TableID:  tbl.TableID,
		ColumnID: spec.col.ColumnID,
		Name:     string(d.Name),
	}
	spec.colType = &scpb.ColumnType{
		TableID:    tbl.TableID,
		ColumnID:   spec.col.ColumnID,
		IsNullable: desc.Nullable,
		IsVirtual:  desc.Virtual,
	}
	if desc.IsComputed() {
		__antithesis_instrumentation__.Notify(579765)
		expr, typ := b.ComputedColumnExpression(tbl, d)
		spec.colType.ComputeExpr = b.WrapExpression(expr)
		spec.colType.TypeT = typ
	} else {
		__antithesis_instrumentation__.Notify(579766)
		spec.colType.TypeT = b.ResolveTypeRef(d.Type)
	}
	__antithesis_instrumentation__.Notify(579745)
	if d.HasColumnFamily() {
		__antithesis_instrumentation__.Notify(579767)
		elts := b.QueryByID(tbl.TableID)
		var found bool
		scpb.ForEachColumnFamily(elts, func(_ scpb.Status, target scpb.TargetStatus, cf *scpb.ColumnFamily) {
			__antithesis_instrumentation__.Notify(579769)
			if target == scpb.ToPublic && func() bool {
				__antithesis_instrumentation__.Notify(579770)
				return cf.Name == string(d.Family.Name) == true
			}() == true {
				__antithesis_instrumentation__.Notify(579771)
				spec.colType.FamilyID = cf.FamilyID
				found = true
			} else {
				__antithesis_instrumentation__.Notify(579772)
			}
		})
		__antithesis_instrumentation__.Notify(579768)
		if !found {
			__antithesis_instrumentation__.Notify(579773)
			if !d.Family.Create {
				__antithesis_instrumentation__.Notify(579775)
				panic(errors.Errorf("unknown family %q", d.Family.Name))
			} else {
				__antithesis_instrumentation__.Notify(579776)
			}
			__antithesis_instrumentation__.Notify(579774)
			spec.fam = &scpb.ColumnFamily{
				TableID:  tbl.TableID,
				FamilyID: b.NextColumnFamilyID(tbl),
				Name:     string(d.Family.Name),
			}
			spec.colType.FamilyID = spec.fam.FamilyID
		} else {
			__antithesis_instrumentation__.Notify(579777)
			if d.Family.Create && func() bool {
				__antithesis_instrumentation__.Notify(579778)
				return !d.Family.IfNotExists == true
			}() == true {
				__antithesis_instrumentation__.Notify(579779)
				panic(errors.Errorf("family %q already exists", d.Family.Name))
			} else {
				__antithesis_instrumentation__.Notify(579780)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(579781)
	}
	__antithesis_instrumentation__.Notify(579746)
	if desc.HasDefault() {
		__antithesis_instrumentation__.Notify(579782)
		spec.def = &scpb.ColumnDefaultExpression{
			TableID:    tbl.TableID,
			ColumnID:   spec.col.ColumnID,
			Expression: *b.WrapExpression(cdd.DefaultExpr),
		}
	} else {
		__antithesis_instrumentation__.Notify(579783)
	}
	__antithesis_instrumentation__.Notify(579747)
	if desc.HasOnUpdate() {
		__antithesis_instrumentation__.Notify(579784)
		spec.onUpdate = &scpb.ColumnOnUpdateExpression{
			TableID:    tbl.TableID,
			ColumnID:   spec.col.ColumnID,
			Expression: *b.WrapExpression(cdd.OnUpdateExpr),
		}
	} else {
		__antithesis_instrumentation__.Notify(579785)
	}
	__antithesis_instrumentation__.Notify(579748)

	if newPrimary := addColumn(b, spec); newPrimary != nil {
		__antithesis_instrumentation__.Notify(579786)
		if idx := cdd.PrimaryKeyOrUniqueIndexDescriptor; idx != nil {
			__antithesis_instrumentation__.Notify(579787)
			idx.ID = b.NextTableIndexID(tbl)
			addSecondaryIndexTargetsForAddColumn(b, tbl, idx, newPrimary.SourceIndexID)
		} else {
			__antithesis_instrumentation__.Notify(579788)
		}
	} else {
		__antithesis_instrumentation__.Notify(579789)
	}
}

type addColumnSpec struct {
	tbl      *scpb.Table
	col      *scpb.Column
	fam      *scpb.ColumnFamily
	name     *scpb.ColumnName
	colType  *scpb.ColumnType
	def      *scpb.ColumnDefaultExpression
	onUpdate *scpb.ColumnOnUpdateExpression
	comment  *scpb.ColumnComment
}

func addColumn(b BuildCtx, spec addColumnSpec) (backing *scpb.PrimaryIndex) {
	__antithesis_instrumentation__.Notify(579790)
	b.Add(spec.col)
	if spec.fam != nil {
		__antithesis_instrumentation__.Notify(579806)
		b.Add(spec.fam)
	} else {
		__antithesis_instrumentation__.Notify(579807)
	}
	__antithesis_instrumentation__.Notify(579791)
	b.Add(spec.name)
	b.Add(spec.colType)
	if spec.def != nil {
		__antithesis_instrumentation__.Notify(579808)
		b.Add(spec.def)
	} else {
		__antithesis_instrumentation__.Notify(579809)
	}
	__antithesis_instrumentation__.Notify(579792)
	if spec.onUpdate != nil {
		__antithesis_instrumentation__.Notify(579810)
		b.Add(spec.onUpdate)
	} else {
		__antithesis_instrumentation__.Notify(579811)
	}
	__antithesis_instrumentation__.Notify(579793)
	if spec.comment != nil {
		__antithesis_instrumentation__.Notify(579812)
		b.Add(spec.comment)
	} else {
		__antithesis_instrumentation__.Notify(579813)
	}
	__antithesis_instrumentation__.Notify(579794)

	if spec.colType.IsVirtual {
		__antithesis_instrumentation__.Notify(579814)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(579815)
	}
	__antithesis_instrumentation__.Notify(579795)

	var existing, freshlyAdded *scpb.PrimaryIndex
	publicTargets := b.QueryByID(spec.tbl.TableID).Filter(
		func(_ scpb.Status, target scpb.TargetStatus, _ scpb.Element) bool {
			__antithesis_instrumentation__.Notify(579816)
			return target == scpb.ToPublic
		},
	)
	__antithesis_instrumentation__.Notify(579796)
	scpb.ForEachPrimaryIndex(publicTargets, func(status scpb.Status, _ scpb.TargetStatus, idx *scpb.PrimaryIndex) {
		__antithesis_instrumentation__.Notify(579817)
		existing = idx
		if status == scpb.Status_ABSENT {
			__antithesis_instrumentation__.Notify(579818)

			freshlyAdded = idx
		} else {
			__antithesis_instrumentation__.Notify(579819)
		}
	})
	__antithesis_instrumentation__.Notify(579797)
	if freshlyAdded != nil {
		__antithesis_instrumentation__.Notify(579820)

		freshlyAdded.StoringColumnIDs = append(freshlyAdded.StoringColumnIDs, spec.col.ColumnID)
		return freshlyAdded
	} else {
		__antithesis_instrumentation__.Notify(579821)
	}
	__antithesis_instrumentation__.Notify(579798)

	if existing == nil {
		__antithesis_instrumentation__.Notify(579822)

		panic(pgerror.Newf(pgcode.NoPrimaryKey, "missing active primary key"))
	} else {
		__antithesis_instrumentation__.Notify(579823)
	}
	__antithesis_instrumentation__.Notify(579799)

	b.Drop(existing)
	var existingName *scpb.IndexName
	var existingPartitioning *scpb.IndexPartitioning
	scpb.ForEachIndexName(publicTargets, func(_ scpb.Status, _ scpb.TargetStatus, name *scpb.IndexName) {
		__antithesis_instrumentation__.Notify(579824)
		if name.IndexID == existing.IndexID {
			__antithesis_instrumentation__.Notify(579825)
			existingName = name
		} else {
			__antithesis_instrumentation__.Notify(579826)
		}
	})
	__antithesis_instrumentation__.Notify(579800)
	scpb.ForEachIndexPartitioning(publicTargets, func(_ scpb.Status, _ scpb.TargetStatus, part *scpb.IndexPartitioning) {
		__antithesis_instrumentation__.Notify(579827)
		if part.IndexID == existing.IndexID {
			__antithesis_instrumentation__.Notify(579828)
			existingPartitioning = part
		} else {
			__antithesis_instrumentation__.Notify(579829)
		}
	})
	__antithesis_instrumentation__.Notify(579801)
	if existingPartitioning != nil {
		__antithesis_instrumentation__.Notify(579830)
		b.Drop(existingPartitioning)
	} else {
		__antithesis_instrumentation__.Notify(579831)
	}
	__antithesis_instrumentation__.Notify(579802)
	if existingName != nil {
		__antithesis_instrumentation__.Notify(579832)
		b.Drop(existingName)
	} else {
		__antithesis_instrumentation__.Notify(579833)
	}
	__antithesis_instrumentation__.Notify(579803)

	replacement := protoutil.Clone(existing).(*scpb.PrimaryIndex)
	replacement.IndexID = b.NextTableIndexID(spec.tbl)
	replacement.SourceIndexID = existing.IndexID
	replacement.StoringColumnIDs = append(replacement.StoringColumnIDs, spec.col.ColumnID)
	b.Add(replacement)
	if existingName != nil {
		__antithesis_instrumentation__.Notify(579834)
		updatedName := protoutil.Clone(existingName).(*scpb.IndexName)
		updatedName.IndexID = replacement.IndexID
		b.Add(updatedName)
	} else {
		__antithesis_instrumentation__.Notify(579835)
	}
	__antithesis_instrumentation__.Notify(579804)
	if existingPartitioning != nil {
		__antithesis_instrumentation__.Notify(579836)
		updatedPartitioning := protoutil.Clone(existingPartitioning).(*scpb.IndexPartitioning)
		updatedPartitioning.IndexID = replacement.IndexID
		b.Add(updatedPartitioning)
	} else {
		__antithesis_instrumentation__.Notify(579837)
	}
	__antithesis_instrumentation__.Notify(579805)
	return replacement
}

func addSecondaryIndexTargetsForAddColumn(
	b BuildCtx, tbl *scpb.Table, desc *descpb.IndexDescriptor, sourceID catid.IndexID,
) {
	__antithesis_instrumentation__.Notify(579838)
	index := scpb.Index{
		TableID:             tbl.TableID,
		IndexID:             desc.ID,
		KeyColumnIDs:        desc.KeyColumnIDs,
		KeyColumnDirections: make([]scpb.Index_Direction, len(desc.KeyColumnIDs)),
		KeySuffixColumnIDs:  desc.KeySuffixColumnIDs,
		StoringColumnIDs:    desc.StoreColumnIDs,
		CompositeColumnIDs:  desc.CompositeColumnIDs,
		IsUnique:            desc.Unique,
		IsInverted:          desc.Type == descpb.IndexDescriptor_INVERTED,
		SourceIndexID:       sourceID,
	}
	for i, dir := range desc.KeyColumnDirections {
		__antithesis_instrumentation__.Notify(579841)
		if dir == descpb.IndexDescriptor_DESC {
			__antithesis_instrumentation__.Notify(579842)
			index.KeyColumnDirections[i] = scpb.Index_DESC
		} else {
			__antithesis_instrumentation__.Notify(579843)
		}
	}
	__antithesis_instrumentation__.Notify(579839)
	if desc.Sharded.IsSharded {
		__antithesis_instrumentation__.Notify(579844)
		index.Sharding = &desc.Sharded
	} else {
		__antithesis_instrumentation__.Notify(579845)
	}
	__antithesis_instrumentation__.Notify(579840)
	b.Add(&scpb.SecondaryIndex{Index: index})
	b.Add(&scpb.IndexName{
		TableID: tbl.TableID,
		IndexID: index.IndexID,
		Name:    desc.Name,
	})
	if p := &desc.Partitioning; len(p.List)+len(p.Range) > 0 {
		__antithesis_instrumentation__.Notify(579846)
		b.Add(&scpb.IndexPartitioning{
			TableID:                tbl.TableID,
			IndexID:                index.IndexID,
			PartitioningDescriptor: *protoutil.Clone(p).(*catpb.PartitioningDescriptor),
		})
	} else {
		__antithesis_instrumentation__.Notify(579847)
	}
}
