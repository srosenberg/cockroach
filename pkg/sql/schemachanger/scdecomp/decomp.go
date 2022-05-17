package scdecomp

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/catid"
	"github.com/cockroachdb/cockroach/pkg/util/iterutil"
	"github.com/cockroachdb/errors"
)

type ElementVisitor func(status scpb.Status, element scpb.Element)

type walkCtx struct {
	desc                 catalog.Descriptor
	ev                   ElementVisitor
	lookupFn             func(id catid.DescID) catalog.Descriptor
	cachedTypeIDClosures map[catid.DescID]map[catid.DescID]struct{}
	backRefs             catalog.DescriptorIDSet
}

func WalkDescriptor(
	desc catalog.Descriptor, lookupFn func(id catid.DescID) catalog.Descriptor, ev ElementVisitor,
) (backRefs catalog.DescriptorIDSet) {
	__antithesis_instrumentation__.Notify(580261)
	w := walkCtx{
		desc:                 desc,
		ev:                   ev,
		lookupFn:             lookupFn,
		cachedTypeIDClosures: make(map[catid.DescID]map[catid.DescID]struct{}),
	}
	w.walkRoot()
	w.backRefs.Remove(catid.InvalidDescID)
	return w.backRefs
}

func (w *walkCtx) walkRoot() {
	__antithesis_instrumentation__.Notify(580262)

	w.ev(scpb.Status_PUBLIC, &scpb.Namespace{
		DatabaseID:   w.desc.GetParentID(),
		SchemaID:     w.desc.GetParentSchemaID(),
		DescriptorID: w.desc.GetID(),
		Name:         w.desc.GetName(),
	})
	privileges := w.desc.GetPrivileges()
	w.ev(scpb.Status_PUBLIC, &scpb.Owner{
		DescriptorID: w.desc.GetID(),
		Owner:        privileges.Owner().Normalized(),
	})
	for _, user := range privileges.Users {
		__antithesis_instrumentation__.Notify(580264)
		w.ev(scpb.Status_PUBLIC, &scpb.UserPrivileges{
			DescriptorID: w.desc.GetID(),
			UserName:     user.User().Normalized(),
			Privileges:   user.Privileges,
		})
	}
	__antithesis_instrumentation__.Notify(580263)

	switch d := w.desc.(type) {
	case catalog.DatabaseDescriptor:
		__antithesis_instrumentation__.Notify(580265)
		w.walkDatabase(d)
	case catalog.SchemaDescriptor:
		__antithesis_instrumentation__.Notify(580266)
		w.walkSchema(d)
	case catalog.TypeDescriptor:
		__antithesis_instrumentation__.Notify(580267)
		w.walkType(d)
	case catalog.TableDescriptor:
		__antithesis_instrumentation__.Notify(580268)
		w.walkRelation(d)
	default:
		__antithesis_instrumentation__.Notify(580269)
		panic(errors.AssertionFailedf("unexpected descriptor type %T: %+v",
			w.desc, w.desc))
	}
}

func (w *walkCtx) walkDatabase(db catalog.DatabaseDescriptor) {
	__antithesis_instrumentation__.Notify(580270)
	w.ev(descriptorStatus(db), &scpb.Database{DatabaseID: db.GetID()})

	w.ev(scpb.Status_PUBLIC, &scpb.DatabaseRoleSetting{
		DatabaseID: db.GetID(),
		RoleName:   scpb.PlaceHolderRoleName,
	})
	w.ev(scpb.Status_PUBLIC, &scpb.DatabaseComment{
		DatabaseID: db.GetID(),
		Comment:    scpb.PlaceHolderComment,
	})
	if db.IsMultiRegion() {
		__antithesis_instrumentation__.Notify(580272)
		w.ev(scpb.Status_PUBLIC, &scpb.DatabaseRegionConfig{
			DatabaseID:       db.GetID(),
			RegionEnumTypeID: db.GetRegionConfig().RegionEnumID,
		})
	} else {
		__antithesis_instrumentation__.Notify(580273)
	}
	__antithesis_instrumentation__.Notify(580271)
	_ = db.ForEachNonDroppedSchema(func(id descpb.ID, name string) error {
		__antithesis_instrumentation__.Notify(580274)
		w.backRefs.Add(id)
		return nil
	})
}

func (w *walkCtx) walkSchema(sc catalog.SchemaDescriptor) {
	__antithesis_instrumentation__.Notify(580275)
	w.ev(descriptorStatus(sc), &scpb.Schema{
		SchemaID:    sc.GetID(),
		IsPublic:    sc.GetName() == catconstants.PublicSchemaName,
		IsVirtual:   sc.SchemaKind() == catalog.SchemaVirtual,
		IsTemporary: sc.SchemaKind() == catalog.SchemaTemporary,
	})
	w.ev(scpb.Status_PUBLIC, &scpb.SchemaParent{
		SchemaID:         sc.GetID(),
		ParentDatabaseID: sc.GetParentID(),
	})

	w.ev(scpb.Status_PUBLIC, &scpb.SchemaComment{
		SchemaID: sc.GetID(),
		Comment:  scpb.PlaceHolderComment,
	})
}

func (w *walkCtx) walkType(typ catalog.TypeDescriptor) {
	__antithesis_instrumentation__.Notify(580276)
	switch typ.GetKind() {
	case descpb.TypeDescriptor_ALIAS:
		__antithesis_instrumentation__.Notify(580278)
		typeT, err := newTypeT(typ.TypeDesc().Alias)
		if err != nil {
			__antithesis_instrumentation__.Notify(580282)
			panic(errors.NewAssertionErrorWithWrappedErrf(err, "alias type %q (%d)",
				typ.GetName(), typ.GetID()))
		} else {
			__antithesis_instrumentation__.Notify(580283)
		}
		__antithesis_instrumentation__.Notify(580279)
		w.ev(descriptorStatus(typ), &scpb.AliasType{
			TypeID: typ.GetID(),
			TypeT:  *typeT,
		})
	case descpb.TypeDescriptor_ENUM, descpb.TypeDescriptor_MULTIREGION_ENUM:
		__antithesis_instrumentation__.Notify(580280)
		w.ev(descriptorStatus(typ), &scpb.EnumType{
			TypeID:        typ.GetID(),
			ArrayTypeID:   typ.GetArrayTypeID(),
			IsMultiRegion: typ.GetKind() == descpb.TypeDescriptor_MULTIREGION_ENUM,
		})
	default:
		__antithesis_instrumentation__.Notify(580281)
		panic(errors.AssertionFailedf("unsupported type kind %q", typ.GetKind()))
	}
	__antithesis_instrumentation__.Notify(580277)
	w.ev(scpb.Status_PUBLIC, &scpb.ObjectParent{
		ObjectID:       typ.GetID(),
		ParentSchemaID: typ.GetParentSchemaID(),
	})
	for i := 0; i < typ.NumReferencingDescriptors(); i++ {
		__antithesis_instrumentation__.Notify(580284)
		w.backRefs.Add(typ.GetReferencingDescriptorID(i))
	}
}

func (w *walkCtx) walkRelation(tbl catalog.TableDescriptor) {
	__antithesis_instrumentation__.Notify(580285)
	switch {
	case tbl.IsSequence():
		__antithesis_instrumentation__.Notify(580294)
		w.ev(descriptorStatus(tbl), &scpb.Sequence{
			SequenceID:  tbl.GetID(),
			IsTemporary: tbl.IsTemporary(),
		})
		if opts := tbl.GetSequenceOpts(); opts != nil {
			__antithesis_instrumentation__.Notify(580297)
			w.backRefs.Add(opts.SequenceOwner.OwnerTableID)
		} else {
			__antithesis_instrumentation__.Notify(580298)
		}
	case tbl.IsView():
		__antithesis_instrumentation__.Notify(580295)
		w.ev(descriptorStatus(tbl), &scpb.View{
			ViewID:          tbl.GetID(),
			UsesTypeIDs:     catalog.MakeDescriptorIDSet(tbl.GetDependsOnTypes()...).Ordered(),
			UsesRelationIDs: catalog.MakeDescriptorIDSet(tbl.GetDependsOn()...).Ordered(),
			IsTemporary:     tbl.IsTemporary(),
			IsMaterialized:  tbl.MaterializedView(),
		})
	default:
		__antithesis_instrumentation__.Notify(580296)
		w.ev(descriptorStatus(tbl), &scpb.Table{
			TableID:     tbl.GetID(),
			IsTemporary: tbl.IsTemporary(),
		})
	}
	__antithesis_instrumentation__.Notify(580286)

	w.ev(scpb.Status_PUBLIC, &scpb.ObjectParent{
		ObjectID:       tbl.GetID(),
		ParentSchemaID: tbl.GetParentSchemaID(),
	})
	if l := tbl.GetLocalityConfig(); l != nil {
		__antithesis_instrumentation__.Notify(580299)
		w.walkLocality(tbl, l)
	} else {
		__antithesis_instrumentation__.Notify(580300)
	}
	{
		__antithesis_instrumentation__.Notify(580301)

		w.ev(scpb.Status_PUBLIC, &scpb.TableComment{
			TableID: tbl.GetID(),
			Comment: scpb.PlaceHolderComment,
		})
	}
	__antithesis_instrumentation__.Notify(580287)
	if !tbl.IsSequence() {
		__antithesis_instrumentation__.Notify(580302)
		_ = tbl.ForeachFamily(func(family *descpb.ColumnFamilyDescriptor) error {
			__antithesis_instrumentation__.Notify(580304)
			w.ev(scpb.Status_PUBLIC, &scpb.ColumnFamily{
				TableID:  tbl.GetID(),
				FamilyID: family.ID,
				Name:     family.Name,
			})
			return nil
		})
		__antithesis_instrumentation__.Notify(580303)
		for _, col := range tbl.AllColumns() {
			__antithesis_instrumentation__.Notify(580305)
			if col.IsSystemColumn() {
				__antithesis_instrumentation__.Notify(580307)
				continue
			} else {
				__antithesis_instrumentation__.Notify(580308)
			}
			__antithesis_instrumentation__.Notify(580306)
			w.walkColumn(tbl, col)
		}
	} else {
		__antithesis_instrumentation__.Notify(580309)
	}
	__antithesis_instrumentation__.Notify(580288)
	if (tbl.IsTable() && func() bool {
		__antithesis_instrumentation__.Notify(580310)
		return !tbl.IsVirtualTable() == true
	}() == true) || func() bool {
		__antithesis_instrumentation__.Notify(580311)
		return tbl.MaterializedView() == true
	}() == true {
		__antithesis_instrumentation__.Notify(580312)
		for _, idx := range tbl.AllIndexes() {
			__antithesis_instrumentation__.Notify(580314)
			w.walkIndex(tbl, idx)
		}
		__antithesis_instrumentation__.Notify(580313)
		if ttl := tbl.GetRowLevelTTL(); ttl != nil {
			__antithesis_instrumentation__.Notify(580315)
			w.ev(scpb.Status_PUBLIC, &scpb.RowLevelTTL{
				TableID:     tbl.GetID(),
				RowLevelTTL: *ttl,
			})
		} else {
			__antithesis_instrumentation__.Notify(580316)
		}
	} else {
		__antithesis_instrumentation__.Notify(580317)
	}
	__antithesis_instrumentation__.Notify(580289)
	for _, c := range tbl.AllActiveAndInactiveUniqueWithoutIndexConstraints() {
		__antithesis_instrumentation__.Notify(580318)
		w.walkUniqueWithoutIndexConstraint(tbl, c)
	}
	__antithesis_instrumentation__.Notify(580290)
	for _, c := range tbl.AllActiveAndInactiveChecks() {
		__antithesis_instrumentation__.Notify(580319)
		w.walkCheckConstraint(tbl, c)
	}
	__antithesis_instrumentation__.Notify(580291)
	for _, c := range tbl.AllActiveAndInactiveForeignKeys() {
		__antithesis_instrumentation__.Notify(580320)
		w.walkForeignKeyConstraint(tbl, c)
	}
	__antithesis_instrumentation__.Notify(580292)

	_ = tbl.ForeachDependedOnBy(func(dep *descpb.TableDescriptor_Reference) error {
		__antithesis_instrumentation__.Notify(580321)
		w.backRefs.Add(dep.ID)
		return nil
	})
	__antithesis_instrumentation__.Notify(580293)
	_ = tbl.ForeachInboundFK(func(fk *descpb.ForeignKeyConstraint) error {
		__antithesis_instrumentation__.Notify(580322)
		w.backRefs.Add(fk.OriginTableID)
		return nil
	})
}

func (w *walkCtx) walkLocality(tbl catalog.TableDescriptor, l *catpb.LocalityConfig) {
	__antithesis_instrumentation__.Notify(580323)
	if g := l.GetGlobal(); g != nil {
		__antithesis_instrumentation__.Notify(580324)
		w.ev(scpb.Status_PUBLIC, &scpb.TableLocalityGlobal{
			TableID: tbl.GetID(),
		})
		return
	} else {
		__antithesis_instrumentation__.Notify(580325)
		if rbr := l.GetRegionalByRow(); rbr != nil {
			__antithesis_instrumentation__.Notify(580326)
			var as string
			if rbr.As != nil {
				__antithesis_instrumentation__.Notify(580328)
				as = *rbr.As
			} else {
				__antithesis_instrumentation__.Notify(580329)
			}
			__antithesis_instrumentation__.Notify(580327)
			w.ev(scpb.Status_PUBLIC, &scpb.TableLocalityRegionalByRow{
				TableID: tbl.GetID(),
				As:      as,
			})
		} else {
			__antithesis_instrumentation__.Notify(580330)
			if rbt := l.GetRegionalByTable(); rbt != nil {
				__antithesis_instrumentation__.Notify(580331)
				if rgn := rbt.Region; rgn != nil {
					__antithesis_instrumentation__.Notify(580332)
					parent := w.lookupFn(tbl.GetParentID())
					db, err := catalog.AsDatabaseDescriptor(parent)
					if err != nil {
						__antithesis_instrumentation__.Notify(580335)
						panic(err)
					} else {
						__antithesis_instrumentation__.Notify(580336)
					}
					__antithesis_instrumentation__.Notify(580333)
					id, err := db.MultiRegionEnumID()
					if err != nil {
						__antithesis_instrumentation__.Notify(580337)
						panic(err)
					} else {
						__antithesis_instrumentation__.Notify(580338)
					}
					__antithesis_instrumentation__.Notify(580334)
					w.ev(scpb.Status_PUBLIC, &scpb.TableLocalitySecondaryRegion{
						TableID:          tbl.GetID(),
						RegionName:       *rgn,
						RegionEnumTypeID: id,
					})
				} else {
					__antithesis_instrumentation__.Notify(580339)
					w.ev(scpb.Status_PUBLIC, &scpb.TableLocalityPrimaryRegion{TableID: tbl.GetID()})
				}
			} else {
				__antithesis_instrumentation__.Notify(580340)
			}
		}
	}
}

func (w *walkCtx) walkColumn(tbl catalog.TableDescriptor, col catalog.Column) {
	__antithesis_instrumentation__.Notify(580341)
	onErrPanic := func(err error) {
		__antithesis_instrumentation__.Notify(580346)
		if err == nil {
			__antithesis_instrumentation__.Notify(580348)
			return
		} else {
			__antithesis_instrumentation__.Notify(580349)
		}
		__antithesis_instrumentation__.Notify(580347)
		panic(errors.NewAssertionErrorWithWrappedErrf(err, "column %q in table %q (%d)",
			col.GetName(), tbl.GetName(), tbl.GetID()))
	}
	__antithesis_instrumentation__.Notify(580342)
	column := &scpb.Column{
		TableID:                           tbl.GetID(),
		ColumnID:                          col.GetID(),
		IsHidden:                          col.IsHidden(),
		IsInaccessible:                    col.IsInaccessible(),
		GeneratedAsIdentityType:           col.GetGeneratedAsIdentityType(),
		GeneratedAsIdentitySequenceOption: col.GetGeneratedAsIdentitySequenceOption(),
		PgAttributeNum:                    col.ColumnDesc().PGAttributeNum,
	}
	w.ev(maybeMutationStatus(col), column)
	w.ev(scpb.Status_PUBLIC, &scpb.ColumnName{
		TableID:  tbl.GetID(),
		ColumnID: col.GetID(),
		Name:     col.GetName(),
	})
	{
		__antithesis_instrumentation__.Notify(580350)
		columnType := &scpb.ColumnType{
			TableID:    tbl.GetID(),
			ColumnID:   col.GetID(),
			IsNullable: col.IsNullable(),
			IsVirtual:  col.IsVirtual(),
		}
		_ = tbl.ForeachFamily(func(family *descpb.ColumnFamilyDescriptor) error {
			__antithesis_instrumentation__.Notify(580353)
			if catalog.MakeTableColSet(family.ColumnIDs...).Contains(col.GetID()) {
				__antithesis_instrumentation__.Notify(580355)
				columnType.FamilyID = family.ID
				return iterutil.StopIteration()
			} else {
				__antithesis_instrumentation__.Notify(580356)
			}
			__antithesis_instrumentation__.Notify(580354)
			return nil
		})
		__antithesis_instrumentation__.Notify(580351)
		typeT, err := newTypeT(col.GetType())
		onErrPanic(err)
		columnType.TypeT = *typeT

		if col.IsComputed() {
			__antithesis_instrumentation__.Notify(580357)
			expr, err := w.newExpression(col.GetComputeExpr())
			onErrPanic(err)
			columnType.ComputeExpr = expr
		} else {
			__antithesis_instrumentation__.Notify(580358)
		}
		__antithesis_instrumentation__.Notify(580352)
		w.ev(scpb.Status_PUBLIC, columnType)
	}
	__antithesis_instrumentation__.Notify(580343)
	if col.HasDefault() {
		__antithesis_instrumentation__.Notify(580359)
		expr, err := w.newExpression(col.GetDefaultExpr())
		onErrPanic(err)
		w.ev(scpb.Status_PUBLIC, &scpb.ColumnDefaultExpression{
			TableID:    tbl.GetID(),
			ColumnID:   col.GetID(),
			Expression: *expr,
		})
	} else {
		__antithesis_instrumentation__.Notify(580360)
	}
	__antithesis_instrumentation__.Notify(580344)
	if col.HasOnUpdate() {
		__antithesis_instrumentation__.Notify(580361)
		expr, err := w.newExpression(col.GetOnUpdateExpr())
		onErrPanic(err)
		w.ev(scpb.Status_PUBLIC, &scpb.ColumnOnUpdateExpression{
			TableID:    tbl.GetID(),
			ColumnID:   col.GetID(),
			Expression: *expr,
		})
	} else {
		__antithesis_instrumentation__.Notify(580362)
	}
	__antithesis_instrumentation__.Notify(580345)

	w.ev(scpb.Status_PUBLIC, &scpb.ColumnComment{
		TableID:  tbl.GetID(),
		ColumnID: col.GetID(),
		Comment:  scpb.PlaceHolderComment,
	})
	owns := catalog.MakeDescriptorIDSet(col.ColumnDesc().OwnsSequenceIds...)
	owns.Remove(catid.InvalidDescID)
	owns.ForEach(func(id descpb.ID) {
		__antithesis_instrumentation__.Notify(580363)
		w.ev(scpb.Status_PUBLIC, &scpb.SequenceOwner{
			SequenceID: id,
			TableID:    tbl.GetID(),
			ColumnID:   col.GetID(),
		})
	})
}

func (w *walkCtx) walkIndex(tbl catalog.TableDescriptor, idx catalog.Index) {
	__antithesis_instrumentation__.Notify(580364)
	onErrPanic := func(err error) {
		__antithesis_instrumentation__.Notify(580366)
		if err == nil {
			__antithesis_instrumentation__.Notify(580368)
			return
		} else {
			__antithesis_instrumentation__.Notify(580369)
		}
		__antithesis_instrumentation__.Notify(580367)
		panic(errors.NewAssertionErrorWithWrappedErrf(err, "index %q in table %q (%d)",
			idx.GetName(), tbl.GetName(), tbl.GetID()))
	}
	{
		__antithesis_instrumentation__.Notify(580370)
		cpy := idx.IndexDescDeepCopy()
		index := scpb.Index{
			TableID:            tbl.GetID(),
			IndexID:            idx.GetID(),
			IsUnique:           idx.IsUnique(),
			KeyColumnIDs:       cpy.KeyColumnIDs,
			KeySuffixColumnIDs: cpy.KeySuffixColumnIDs,
			StoringColumnIDs:   cpy.StoreColumnIDs,
			CompositeColumnIDs: cpy.CompositeColumnIDs,
			IsInverted:         idx.GetType() == descpb.IndexDescriptor_INVERTED,
		}
		index.KeyColumnDirections = make([]scpb.Index_Direction, len(index.KeyColumnIDs))
		for i := 0; i < idx.NumKeyColumns(); i++ {
			__antithesis_instrumentation__.Notify(580374)
			if idx.GetKeyColumnDirection(i) == descpb.IndexDescriptor_DESC {
				__antithesis_instrumentation__.Notify(580375)
				index.KeyColumnDirections[i] = scpb.Index_DESC
			} else {
				__antithesis_instrumentation__.Notify(580376)
			}
		}
		__antithesis_instrumentation__.Notify(580371)
		if idx.IsSharded() {
			__antithesis_instrumentation__.Notify(580377)
			index.Sharding = &cpy.Sharded
		} else {
			__antithesis_instrumentation__.Notify(580378)
		}
		__antithesis_instrumentation__.Notify(580372)
		idxStatus := maybeMutationStatus(idx)
		if idx.GetEncodingType() == descpb.PrimaryIndexEncoding {
			__antithesis_instrumentation__.Notify(580379)
			w.ev(idxStatus, &scpb.PrimaryIndex{Index: index})
		} else {
			__antithesis_instrumentation__.Notify(580380)
			sec := &scpb.SecondaryIndex{Index: index}
			if idx.IsPartial() {
				__antithesis_instrumentation__.Notify(580382)
				pp, err := w.newExpression(idx.GetPredicate())
				onErrPanic(err)
				w.ev(scpb.Status_PUBLIC, &scpb.SecondaryIndexPartial{
					TableID:    index.TableID,
					IndexID:    index.IndexID,
					Expression: *pp,
				})
			} else {
				__antithesis_instrumentation__.Notify(580383)
			}
			__antithesis_instrumentation__.Notify(580381)
			w.ev(idxStatus, sec)
		}
		__antithesis_instrumentation__.Notify(580373)
		if p := idx.GetPartitioning(); p != nil && func() bool {
			__antithesis_instrumentation__.Notify(580384)
			return p.NumLists()+p.NumRanges() > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(580385)
			w.ev(scpb.Status_PUBLIC, &scpb.IndexPartitioning{
				TableID:                tbl.GetID(),
				IndexID:                idx.GetID(),
				PartitioningDescriptor: cpy.Partitioning,
			})
		} else {
			__antithesis_instrumentation__.Notify(580386)
		}
	}
	__antithesis_instrumentation__.Notify(580365)
	w.ev(scpb.Status_PUBLIC, &scpb.IndexName{
		TableID: tbl.GetID(),
		IndexID: idx.GetID(),
		Name:    idx.GetName(),
	})

	w.ev(scpb.Status_PUBLIC, &scpb.IndexComment{
		TableID: tbl.GetID(),
		IndexID: idx.GetID(),
		Comment: scpb.PlaceHolderComment,
	})
}

func (w *walkCtx) walkUniqueWithoutIndexConstraint(
	tbl catalog.TableDescriptor, c *descpb.UniqueWithoutIndexConstraint,
) {
	__antithesis_instrumentation__.Notify(580387)

	w.ev(scpb.Status_PUBLIC, &scpb.UniqueWithoutIndexConstraint{
		TableID:      tbl.GetID(),
		ConstraintID: c.ConstraintID,
		ColumnIDs:    catalog.MakeTableColSet(c.ColumnIDs...).Ordered(),
	})
	w.ev(scpb.Status_PUBLIC, &scpb.ConstraintName{
		TableID:      tbl.GetID(),
		ConstraintID: c.ConstraintID,
		Name:         c.Name,
	})

	w.ev(scpb.Status_PUBLIC, &scpb.ConstraintComment{
		TableID:      tbl.GetID(),
		ConstraintID: c.ConstraintID,
		Comment:      scpb.PlaceHolderComment,
	})
}

func (w *walkCtx) walkCheckConstraint(
	tbl catalog.TableDescriptor, c *descpb.TableDescriptor_CheckConstraint,
) {
	__antithesis_instrumentation__.Notify(580388)
	expr, err := w.newExpression(c.Expr)
	if err != nil {
		__antithesis_instrumentation__.Notify(580390)
		panic(errors.NewAssertionErrorWithWrappedErrf(err, "check constraint %q in table %q (%d)",
			c.Name, tbl.GetName(), tbl.GetID()))
	} else {
		__antithesis_instrumentation__.Notify(580391)
	}
	__antithesis_instrumentation__.Notify(580389)

	w.ev(scpb.Status_PUBLIC, &scpb.CheckConstraint{
		TableID:      tbl.GetID(),
		ConstraintID: c.ConstraintID,
		ColumnIDs:    catalog.MakeTableColSet(c.ColumnIDs...).Ordered(),
		Expression:   *expr,
	})
	w.ev(scpb.Status_PUBLIC, &scpb.ConstraintName{
		TableID:      tbl.GetID(),
		ConstraintID: c.ConstraintID,
		Name:         c.Name,
	})

	w.ev(scpb.Status_PUBLIC, &scpb.ConstraintComment{
		TableID:      tbl.GetID(),
		ConstraintID: c.ConstraintID,
		Comment:      scpb.PlaceHolderComment,
	})
}

func (w *walkCtx) walkForeignKeyConstraint(
	tbl catalog.TableDescriptor, c *descpb.ForeignKeyConstraint,
) {
	__antithesis_instrumentation__.Notify(580392)

	w.ev(scpb.Status_PUBLIC, &scpb.ForeignKeyConstraint{
		TableID:             tbl.GetID(),
		ConstraintID:        c.ConstraintID,
		ColumnIDs:           catalog.MakeTableColSet(c.OriginColumnIDs...).Ordered(),
		ReferencedTableID:   c.ReferencedTableID,
		ReferencedColumnIDs: catalog.MakeTableColSet(c.ReferencedColumnIDs...).Ordered(),
	})
	w.ev(scpb.Status_PUBLIC, &scpb.ConstraintName{
		TableID:      tbl.GetID(),
		ConstraintID: c.ConstraintID,
		Name:         c.Name,
	})

	w.ev(scpb.Status_PUBLIC, &scpb.ConstraintComment{
		TableID:      tbl.GetID(),
		ConstraintID: c.ConstraintID,
		Comment:      scpb.PlaceHolderComment,
	})
}
