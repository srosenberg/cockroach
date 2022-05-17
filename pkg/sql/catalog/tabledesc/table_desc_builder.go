package tabledesc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catprivilege"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
)

type TableDescriptorBuilder interface {
	catalog.DescriptorBuilder
	BuildImmutableTable() catalog.TableDescriptor
	BuildExistingMutableTable() *Mutable
	BuildCreatedMutableTable() *Mutable
}

type tableDescriptorBuilder struct {
	original                   *descpb.TableDescriptor
	maybeModified              *descpb.TableDescriptor
	changes                    catalog.PostDeserializationChanges
	skipFKsWithNoMatchingTable bool
	isUncommittedVersion       bool
}

var _ TableDescriptorBuilder = &tableDescriptorBuilder{}

func NewBuilder(desc *descpb.TableDescriptor) TableDescriptorBuilder {
	__antithesis_instrumentation__.Notify(270687)
	return newBuilder(desc, false,
		catalog.PostDeserializationChanges{})
}

func NewBuilderForFKUpgrade(
	desc *descpb.TableDescriptor, skipFKsWithNoMatchingTable bool,
) TableDescriptorBuilder {
	__antithesis_instrumentation__.Notify(270688)
	b := newBuilder(desc, false,
		catalog.PostDeserializationChanges{})
	b.skipFKsWithNoMatchingTable = skipFKsWithNoMatchingTable
	return b
}

func NewUnsafeImmutable(desc *descpb.TableDescriptor) catalog.TableDescriptor {
	__antithesis_instrumentation__.Notify(270689)
	b := tableDescriptorBuilder{original: desc}
	return b.BuildImmutableTable()
}

func newBuilder(
	desc *descpb.TableDescriptor,
	isUncommittedVersion bool,
	changes catalog.PostDeserializationChanges,
) *tableDescriptorBuilder {
	__antithesis_instrumentation__.Notify(270690)
	return &tableDescriptorBuilder{
		original:             protoutil.Clone(desc).(*descpb.TableDescriptor),
		isUncommittedVersion: isUncommittedVersion,
		changes:              changes,
	}
}

func (tdb *tableDescriptorBuilder) DescriptorType() catalog.DescriptorType {
	__antithesis_instrumentation__.Notify(270691)
	return catalog.Table
}

func (tdb *tableDescriptorBuilder) RunPostDeserializationChanges() error {
	__antithesis_instrumentation__.Notify(270692)
	var err error

	prevChanges := tdb.changes
	tdb.maybeModified = protoutil.Clone(tdb.original).(*descpb.TableDescriptor)
	tdb.changes, err = maybeFillInDescriptor(tdb.maybeModified)
	if err != nil {
		__antithesis_instrumentation__.Notify(270695)
		return err
	} else {
		__antithesis_instrumentation__.Notify(270696)
	}
	__antithesis_instrumentation__.Notify(270693)
	prevChanges.ForEach(func(change catalog.PostDeserializationChangeType) {
		__antithesis_instrumentation__.Notify(270697)
		tdb.changes.Add(change)
	})
	__antithesis_instrumentation__.Notify(270694)
	return nil
}

func (tdb *tableDescriptorBuilder) RunRestoreChanges(
	descLookupFn func(id descpb.ID) catalog.Descriptor,
) (err error) {
	__antithesis_instrumentation__.Notify(270698)
	upgradedFK, err := maybeUpgradeForeignKeyRepresentation(
		descLookupFn,
		tdb.skipFKsWithNoMatchingTable,
		tdb.maybeModified,
	)
	if upgradedFK {
		__antithesis_instrumentation__.Notify(270700)
		tdb.changes.Add(catalog.UpgradedForeignKeyRepresentation)
	} else {
		__antithesis_instrumentation__.Notify(270701)
	}
	__antithesis_instrumentation__.Notify(270699)
	return err
}

func (tdb *tableDescriptorBuilder) BuildImmutable() catalog.Descriptor {
	__antithesis_instrumentation__.Notify(270702)
	return tdb.BuildImmutableTable()
}

func (tdb *tableDescriptorBuilder) BuildImmutableTable() catalog.TableDescriptor {
	__antithesis_instrumentation__.Notify(270703)
	desc := tdb.maybeModified
	if desc == nil {
		__antithesis_instrumentation__.Notify(270705)
		desc = tdb.original
	} else {
		__antithesis_instrumentation__.Notify(270706)
	}
	__antithesis_instrumentation__.Notify(270704)
	imm := makeImmutable(desc)
	imm.changes = tdb.changes
	imm.isUncommittedVersion = tdb.isUncommittedVersion
	return imm
}

func (tdb *tableDescriptorBuilder) BuildExistingMutable() catalog.MutableDescriptor {
	__antithesis_instrumentation__.Notify(270707)
	return tdb.BuildExistingMutableTable()
}

func (tdb *tableDescriptorBuilder) BuildExistingMutableTable() *Mutable {
	__antithesis_instrumentation__.Notify(270708)
	if tdb.maybeModified == nil {
		__antithesis_instrumentation__.Notify(270710)
		tdb.maybeModified = protoutil.Clone(tdb.original).(*descpb.TableDescriptor)
	} else {
		__antithesis_instrumentation__.Notify(270711)
	}
	__antithesis_instrumentation__.Notify(270709)
	return &Mutable{
		wrapper: wrapper{
			TableDescriptor: *tdb.maybeModified,
			changes:         tdb.changes,
		},
		original: makeImmutable(tdb.original),
	}
}

func (tdb *tableDescriptorBuilder) BuildCreatedMutable() catalog.MutableDescriptor {
	__antithesis_instrumentation__.Notify(270712)
	return tdb.BuildCreatedMutableTable()
}

func (tdb *tableDescriptorBuilder) BuildCreatedMutableTable() *Mutable {
	__antithesis_instrumentation__.Notify(270713)
	desc := tdb.maybeModified
	if desc == nil {
		__antithesis_instrumentation__.Notify(270715)
		desc = tdb.original
	} else {
		__antithesis_instrumentation__.Notify(270716)
	}
	__antithesis_instrumentation__.Notify(270714)
	return &Mutable{
		wrapper: wrapper{
			TableDescriptor: *desc,
			changes:         tdb.changes,
		},
	}
}

func makeImmutable(tbl *descpb.TableDescriptor) *immutable {
	__antithesis_instrumentation__.Notify(270717)
	desc := immutable{wrapper: wrapper{TableDescriptor: *tbl}}
	desc.mutationCache = newMutationCache(desc.TableDesc())
	desc.indexCache = newIndexCache(desc.TableDesc(), desc.mutationCache)
	desc.columnCache = newColumnCache(desc.TableDesc(), desc.mutationCache)

	desc.allChecks = make([]descpb.TableDescriptor_CheckConstraint, len(tbl.Checks))
	for i, c := range tbl.Checks {
		__antithesis_instrumentation__.Notify(270719)
		desc.allChecks[i] = *c
	}
	__antithesis_instrumentation__.Notify(270718)

	return &desc
}

func maybeFillInDescriptor(
	desc *descpb.TableDescriptor,
) (changes catalog.PostDeserializationChanges, err error) {
	__antithesis_instrumentation__.Notify(270720)
	set := func(change catalog.PostDeserializationChangeType, cond bool) {
		__antithesis_instrumentation__.Notify(270726)
		if cond {
			__antithesis_instrumentation__.Notify(270727)
			changes.Add(change)
		} else {
			__antithesis_instrumentation__.Notify(270728)
		}
	}
	__antithesis_instrumentation__.Notify(270721)
	set(catalog.UpgradedFormatVersion, maybeUpgradeFormatVersion(desc))
	set(catalog.FixedIndexEncodingType, maybeFixPrimaryIndexEncoding(&desc.PrimaryIndex))
	set(catalog.UpgradedIndexFormatVersion, maybeUpgradePrimaryIndexFormatVersion(desc))
	for i := range desc.Indexes {
		__antithesis_instrumentation__.Notify(270729)
		idx := &desc.Indexes[i]
		set(catalog.UpgradedIndexFormatVersion,
			maybeUpgradeSecondaryIndexFormatVersion(idx))
	}
	__antithesis_instrumentation__.Notify(270722)
	for i := range desc.Mutations {
		__antithesis_instrumentation__.Notify(270730)
		if idx := desc.Mutations[i].GetIndex(); idx != nil {
			__antithesis_instrumentation__.Notify(270731)
			set(catalog.UpgradedIndexFormatVersion,
				maybeUpgradeSecondaryIndexFormatVersion(idx))
		} else {
			__antithesis_instrumentation__.Notify(270732)
		}
	}
	__antithesis_instrumentation__.Notify(270723)
	set(catalog.UpgradedNamespaceName, maybeUpgradeNamespaceName(desc))
	set(catalog.RemovedDefaultExprFromComputedColumn,
		maybeRemoveDefaultExprFromComputedColumns(desc))

	parentSchemaID := desc.GetUnexposedParentSchemaID()

	if parentSchemaID == descpb.InvalidID {
		__antithesis_instrumentation__.Notify(270733)
		parentSchemaID = keys.PublicSchemaID
	} else {
		__antithesis_instrumentation__.Notify(270734)
	}
	__antithesis_instrumentation__.Notify(270724)
	fixedPrivileges := catprivilege.MaybeFixPrivileges(
		&desc.Privileges,
		desc.GetParentID(),
		parentSchemaID,
		privilege.Table,
		desc.GetName(),
	)
	addedGrantOptions := catprivilege.MaybeUpdateGrantOptions(desc.Privileges)
	set(catalog.UpgradedPrivileges, fixedPrivileges || func() bool {
		__antithesis_instrumentation__.Notify(270735)
		return addedGrantOptions == true
	}() == true)
	set(catalog.RemovedDuplicateIDsInRefs, maybeRemoveDuplicateIDsInRefs(desc))
	set(catalog.AddedConstraintIDs, maybeAddConstraintIDs(desc))

	rewrittenCast, err := maybeRewriteCast(desc)
	if err != nil {
		__antithesis_instrumentation__.Notify(270736)
		return changes, err
	} else {
		__antithesis_instrumentation__.Notify(270737)
	}
	__antithesis_instrumentation__.Notify(270725)
	set(catalog.FixedDateStyleIntervalStyleCast, rewrittenCast)
	return changes, nil
}

func maybeRewriteCast(desc *descpb.TableDescriptor) (hasChanged bool, err error) {
	__antithesis_instrumentation__.Notify(270738)

	if desc.ParentID == keys.SystemDatabaseID {
		__antithesis_instrumentation__.Notify(270742)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(270743)
	}
	__antithesis_instrumentation__.Notify(270739)

	ctx := context.Background()
	var semaCtx tree.SemaContext
	semaCtx.IntervalStyleEnabled = true
	semaCtx.DateStyleEnabled = true
	hasChanged = false

	for i, col := range desc.Columns {
		__antithesis_instrumentation__.Notify(270744)
		if col.IsComputed() {
			__antithesis_instrumentation__.Notify(270745)
			expr, err := parser.ParseExpr(*col.ComputeExpr)
			if err != nil {
				__antithesis_instrumentation__.Notify(270748)
				return hasChanged, err
			} else {
				__antithesis_instrumentation__.Notify(270749)
			}
			__antithesis_instrumentation__.Notify(270746)
			newExpr, changed, err := ResolveCastForStyleUsingVisitor(
				ctx,
				&semaCtx,
				desc,
				expr,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(270750)
				return hasChanged, err
			} else {
				__antithesis_instrumentation__.Notify(270751)
			}
			__antithesis_instrumentation__.Notify(270747)
			if changed {
				__antithesis_instrumentation__.Notify(270752)
				hasChanged = true
				s := tree.Serialize(newExpr)
				desc.Columns[i].ComputeExpr = &s
			} else {
				__antithesis_instrumentation__.Notify(270753)
			}
		} else {
			__antithesis_instrumentation__.Notify(270754)
		}
	}
	__antithesis_instrumentation__.Notify(270740)

	for i, idx := range desc.Indexes {
		__antithesis_instrumentation__.Notify(270755)
		if idx.IsPartial() {
			__antithesis_instrumentation__.Notify(270756)
			expr, err := parser.ParseExpr(idx.Predicate)

			if err != nil {
				__antithesis_instrumentation__.Notify(270759)
				return hasChanged, err
			} else {
				__antithesis_instrumentation__.Notify(270760)
			}
			__antithesis_instrumentation__.Notify(270757)
			newExpr, changed, err := ResolveCastForStyleUsingVisitor(
				ctx,
				&semaCtx,
				desc,
				expr,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(270761)
				return hasChanged, err
			} else {
				__antithesis_instrumentation__.Notify(270762)
			}
			__antithesis_instrumentation__.Notify(270758)
			if changed {
				__antithesis_instrumentation__.Notify(270763)
				hasChanged = true
				s := tree.Serialize(newExpr)
				desc.Indexes[i].Predicate = s
			} else {
				__antithesis_instrumentation__.Notify(270764)
			}
		} else {
			__antithesis_instrumentation__.Notify(270765)
		}
	}
	__antithesis_instrumentation__.Notify(270741)
	return hasChanged, nil
}

func maybeRemoveDefaultExprFromComputedColumns(desc *descpb.TableDescriptor) (hasChanged bool) {
	__antithesis_instrumentation__.Notify(270766)
	doCol := func(col *descpb.ColumnDescriptor) {
		__antithesis_instrumentation__.Notify(270770)
		if col.IsComputed() && func() bool {
			__antithesis_instrumentation__.Notify(270771)
			return col.HasDefault() == true
		}() == true {
			__antithesis_instrumentation__.Notify(270772)
			col.DefaultExpr = nil
			hasChanged = true
		} else {
			__antithesis_instrumentation__.Notify(270773)
		}
	}
	__antithesis_instrumentation__.Notify(270767)

	for i := range desc.Columns {
		__antithesis_instrumentation__.Notify(270774)
		doCol(&desc.Columns[i])
	}
	__antithesis_instrumentation__.Notify(270768)
	for _, m := range desc.Mutations {
		__antithesis_instrumentation__.Notify(270775)
		if col := m.GetColumn(); col != nil && func() bool {
			__antithesis_instrumentation__.Notify(270776)
			return m.Direction != descpb.DescriptorMutation_DROP == true
		}() == true {
			__antithesis_instrumentation__.Notify(270777)
			doCol(col)
		} else {
			__antithesis_instrumentation__.Notify(270778)
		}
	}
	__antithesis_instrumentation__.Notify(270769)
	return hasChanged
}

func maybeUpgradeForeignKeyRepresentation(
	descLookupFn func(id descpb.ID) catalog.Descriptor,
	skipFKsWithNoMatchingTable bool,
	desc *descpb.TableDescriptor,
) (bool, error) {
	__antithesis_instrumentation__.Notify(270779)
	if desc.Dropped() {
		__antithesis_instrumentation__.Notify(270783)

		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(270784)
	}
	__antithesis_instrumentation__.Notify(270780)
	otherUnupgradedTables := make(map[descpb.ID]catalog.TableDescriptor)
	changed := false

	for i := range desc.Indexes {
		__antithesis_instrumentation__.Notify(270785)
		newChanged, err := maybeUpgradeForeignKeyRepOnIndex(
			descLookupFn, otherUnupgradedTables, desc, &desc.Indexes[i], skipFKsWithNoMatchingTable,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(270787)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(270788)
		}
		__antithesis_instrumentation__.Notify(270786)
		changed = changed || func() bool {
			__antithesis_instrumentation__.Notify(270789)
			return newChanged == true
		}() == true
	}
	__antithesis_instrumentation__.Notify(270781)
	newChanged, err := maybeUpgradeForeignKeyRepOnIndex(
		descLookupFn, otherUnupgradedTables, desc, &desc.PrimaryIndex, skipFKsWithNoMatchingTable,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(270790)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(270791)
	}
	__antithesis_instrumentation__.Notify(270782)
	changed = changed || func() bool {
		__antithesis_instrumentation__.Notify(270792)
		return newChanged == true
	}() == true

	return changed, nil
}

func maybeUpgradeForeignKeyRepOnIndex(
	descLookupFn func(id descpb.ID) catalog.Descriptor,
	otherUnupgradedTables map[descpb.ID]catalog.TableDescriptor,
	desc *descpb.TableDescriptor,
	idx *descpb.IndexDescriptor,
	skipFKsWithNoMatchingTable bool,
) (bool, error) {
	__antithesis_instrumentation__.Notify(270793)
	updateUnupgradedTablesMap := func(id descpb.ID) (err error) {
		__antithesis_instrumentation__.Notify(270797)
		defer func() {
			__antithesis_instrumentation__.Notify(270802)
			if errors.Is(err, catalog.ErrDescriptorNotFound) && func() bool {
				__antithesis_instrumentation__.Notify(270803)
				return skipFKsWithNoMatchingTable == true
			}() == true {
				__antithesis_instrumentation__.Notify(270804)
				err = nil
			} else {
				__antithesis_instrumentation__.Notify(270805)
			}
		}()
		__antithesis_instrumentation__.Notify(270798)
		if _, found := otherUnupgradedTables[id]; found {
			__antithesis_instrumentation__.Notify(270806)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(270807)
		}
		__antithesis_instrumentation__.Notify(270799)
		d := descLookupFn(id)
		if d == nil {
			__antithesis_instrumentation__.Notify(270808)
			return catalog.WrapTableDescRefErr(id, catalog.ErrDescriptorNotFound)
		} else {
			__antithesis_instrumentation__.Notify(270809)
		}
		__antithesis_instrumentation__.Notify(270800)
		tbl, ok := d.(catalog.TableDescriptor)
		if !ok {
			__antithesis_instrumentation__.Notify(270810)
			return catalog.WrapTableDescRefErr(id, catalog.ErrDescriptorNotFound)
		} else {
			__antithesis_instrumentation__.Notify(270811)
		}
		__antithesis_instrumentation__.Notify(270801)
		otherUnupgradedTables[id] = tbl
		return nil
	}
	__antithesis_instrumentation__.Notify(270794)

	var changed bool
	if idx.ForeignKey.IsSet() {
		__antithesis_instrumentation__.Notify(270812)
		ref := &idx.ForeignKey
		if err := updateUnupgradedTablesMap(ref.Table); err != nil {
			__antithesis_instrumentation__.Notify(270815)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(270816)
		}
		__antithesis_instrumentation__.Notify(270813)
		if tbl, ok := otherUnupgradedTables[ref.Table]; ok {
			__antithesis_instrumentation__.Notify(270817)
			referencedIndex, err := tbl.FindIndexWithID(ref.Index)
			if err != nil {
				__antithesis_instrumentation__.Notify(270819)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(270820)
			}
			__antithesis_instrumentation__.Notify(270818)
			numCols := ref.SharedPrefixLen
			outFK := descpb.ForeignKeyConstraint{
				OriginTableID:       desc.ID,
				OriginColumnIDs:     idx.KeyColumnIDs[:numCols],
				ReferencedTableID:   ref.Table,
				ReferencedColumnIDs: referencedIndex.IndexDesc().KeyColumnIDs[:numCols],
				Name:                ref.Name,
				Validity:            ref.Validity,
				OnDelete:            ref.OnDelete,
				OnUpdate:            ref.OnUpdate,
				Match:               ref.Match,
				ConstraintID:        desc.GetNextConstraintID(),
			}
			desc.NextConstraintID++
			desc.OutboundFKs = append(desc.OutboundFKs, outFK)
		} else {
			__antithesis_instrumentation__.Notify(270821)
		}
		__antithesis_instrumentation__.Notify(270814)
		changed = true
		idx.ForeignKey = descpb.ForeignKeyReference{}
	} else {
		__antithesis_instrumentation__.Notify(270822)
	}
	__antithesis_instrumentation__.Notify(270795)

	for refIdx := range idx.ReferencedBy {
		__antithesis_instrumentation__.Notify(270823)
		ref := &(idx.ReferencedBy[refIdx])
		if err := updateUnupgradedTablesMap(ref.Table); err != nil {
			__antithesis_instrumentation__.Notify(270826)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(270827)
		}
		__antithesis_instrumentation__.Notify(270824)
		if otherTable, ok := otherUnupgradedTables[ref.Table]; ok {
			__antithesis_instrumentation__.Notify(270828)
			originIndexI, err := otherTable.FindIndexWithID(ref.Index)
			if err != nil {
				__antithesis_instrumentation__.Notify(270831)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(270832)
			}
			__antithesis_instrumentation__.Notify(270829)
			originIndex := originIndexI.IndexDesc()

			var inFK descpb.ForeignKeyConstraint
			if !originIndex.ForeignKey.IsSet() {
				__antithesis_instrumentation__.Notify(270833)

				var forwardFK *descpb.ForeignKeyConstraint
				_ = otherTable.ForeachOutboundFK(func(otherFK *descpb.ForeignKeyConstraint) error {
					__antithesis_instrumentation__.Notify(270836)
					if forwardFK != nil {
						__antithesis_instrumentation__.Notify(270839)
						return nil
					} else {
						__antithesis_instrumentation__.Notify(270840)
					}
					__antithesis_instrumentation__.Notify(270837)

					if otherFK.ReferencedTableID == desc.ID && func() bool {
						__antithesis_instrumentation__.Notify(270841)
						return descpb.ColumnIDs(originIndex.KeyColumnIDs).HasPrefix(otherFK.OriginColumnIDs) == true
					}() == true {
						__antithesis_instrumentation__.Notify(270842)

						forwardFK = otherFK
					} else {
						__antithesis_instrumentation__.Notify(270843)
					}
					__antithesis_instrumentation__.Notify(270838)
					return nil
				})
				__antithesis_instrumentation__.Notify(270834)
				if forwardFK == nil {
					__antithesis_instrumentation__.Notify(270844)

					return false, errors.AssertionFailedf(
						"error finding foreign key on table %d for backref %+v",
						otherTable.GetID(), ref)
				} else {
					__antithesis_instrumentation__.Notify(270845)
				}
				__antithesis_instrumentation__.Notify(270835)
				inFK = descpb.ForeignKeyConstraint{
					OriginTableID:       ref.Table,
					OriginColumnIDs:     forwardFK.OriginColumnIDs,
					ReferencedTableID:   desc.ID,
					ReferencedColumnIDs: forwardFK.ReferencedColumnIDs,
					Name:                forwardFK.Name,
					Validity:            forwardFK.Validity,
					OnDelete:            forwardFK.OnDelete,
					OnUpdate:            forwardFK.OnUpdate,
					Match:               forwardFK.Match,
					ConstraintID:        desc.GetNextConstraintID(),
				}
			} else {
				__antithesis_instrumentation__.Notify(270846)

				numCols := originIndex.ForeignKey.SharedPrefixLen
				inFK = descpb.ForeignKeyConstraint{
					OriginTableID:       ref.Table,
					OriginColumnIDs:     originIndex.KeyColumnIDs[:numCols],
					ReferencedTableID:   desc.ID,
					ReferencedColumnIDs: idx.KeyColumnIDs[:numCols],
					Name:                originIndex.ForeignKey.Name,
					Validity:            originIndex.ForeignKey.Validity,
					OnDelete:            originIndex.ForeignKey.OnDelete,
					OnUpdate:            originIndex.ForeignKey.OnUpdate,
					Match:               originIndex.ForeignKey.Match,
					ConstraintID:        desc.GetNextConstraintID(),
				}
			}
			__antithesis_instrumentation__.Notify(270830)
			desc.NextConstraintID++
			desc.InboundFKs = append(desc.InboundFKs, inFK)
		} else {
			__antithesis_instrumentation__.Notify(270847)
		}
		__antithesis_instrumentation__.Notify(270825)
		changed = true
	}
	__antithesis_instrumentation__.Notify(270796)
	idx.ReferencedBy = nil
	return changed, nil
}

func maybeUpgradeFormatVersion(desc *descpb.TableDescriptor) (wasUpgraded bool) {
	__antithesis_instrumentation__.Notify(270848)
	for _, pair := range []struct {
		targetVersion descpb.FormatVersion
		upgradeFn     func(*descpb.TableDescriptor)
	}{
		{descpb.FamilyFormatVersion, upgradeToFamilyFormatVersion},
		{descpb.InterleavedFormatVersion, func(_ *descpb.TableDescriptor) { __antithesis_instrumentation__.Notify(270850) }},
	} {
		__antithesis_instrumentation__.Notify(270851)
		if desc.FormatVersion < pair.targetVersion {
			__antithesis_instrumentation__.Notify(270852)
			pair.upgradeFn(desc)
			desc.FormatVersion = pair.targetVersion
			wasUpgraded = true
		} else {
			__antithesis_instrumentation__.Notify(270853)
		}
	}
	__antithesis_instrumentation__.Notify(270849)
	return wasUpgraded
}

const FamilyPrimaryName = "primary"

func upgradeToFamilyFormatVersion(desc *descpb.TableDescriptor) {
	__antithesis_instrumentation__.Notify(270854)
	var primaryIndexColumnIDs catalog.TableColSet
	for _, colID := range desc.PrimaryIndex.KeyColumnIDs {
		__antithesis_instrumentation__.Notify(270858)
		primaryIndexColumnIDs.Add(colID)
	}
	__antithesis_instrumentation__.Notify(270855)

	desc.Families = []descpb.ColumnFamilyDescriptor{
		{ID: 0, Name: FamilyPrimaryName},
	}
	desc.NextFamilyID = desc.Families[0].ID + 1
	addFamilyForCol := func(col *descpb.ColumnDescriptor) {
		__antithesis_instrumentation__.Notify(270859)
		if primaryIndexColumnIDs.Contains(col.ID) {
			__antithesis_instrumentation__.Notify(270861)
			desc.Families[0].ColumnNames = append(desc.Families[0].ColumnNames, col.Name)
			desc.Families[0].ColumnIDs = append(desc.Families[0].ColumnIDs, col.ID)
			return
		} else {
			__antithesis_instrumentation__.Notify(270862)
		}
		__antithesis_instrumentation__.Notify(270860)
		colNames := []string{col.Name}
		family := descpb.ColumnFamilyDescriptor{
			ID:              descpb.FamilyID(col.ID),
			Name:            generatedFamilyName(descpb.FamilyID(col.ID), colNames),
			ColumnNames:     colNames,
			ColumnIDs:       []descpb.ColumnID{col.ID},
			DefaultColumnID: col.ID,
		}
		desc.Families = append(desc.Families, family)
		if family.ID >= desc.NextFamilyID {
			__antithesis_instrumentation__.Notify(270863)
			desc.NextFamilyID = family.ID + 1
		} else {
			__antithesis_instrumentation__.Notify(270864)
		}
	}
	__antithesis_instrumentation__.Notify(270856)

	for i := range desc.Columns {
		__antithesis_instrumentation__.Notify(270865)
		addFamilyForCol(&desc.Columns[i])
	}
	__antithesis_instrumentation__.Notify(270857)
	for i := range desc.Mutations {
		__antithesis_instrumentation__.Notify(270866)
		m := &desc.Mutations[i]
		if c := m.GetColumn(); c != nil {
			__antithesis_instrumentation__.Notify(270867)
			addFamilyForCol(c)
		} else {
			__antithesis_instrumentation__.Notify(270868)
		}
	}
}

func maybeUpgradePrimaryIndexFormatVersion(desc *descpb.TableDescriptor) (hasChanged bool) {
	__antithesis_instrumentation__.Notify(270869)

	desc.PrimaryIndex.EncodingType = descpb.PrimaryIndexEncoding

	switch desc.PrimaryIndex.Version {
	case descpb.PrimaryIndexWithStoredColumnsVersion:
		__antithesis_instrumentation__.Notify(270877)
		return false
	default:
		__antithesis_instrumentation__.Notify(270878)
		break
	}
	__antithesis_instrumentation__.Notify(270870)

	nonVirtualCols := make([]*descpb.ColumnDescriptor, 0, len(desc.Columns)+len(desc.Mutations))
	maybeAddCol := func(col *descpb.ColumnDescriptor) {
		__antithesis_instrumentation__.Notify(270879)
		if col == nil || func() bool {
			__antithesis_instrumentation__.Notify(270881)
			return col.Virtual == true
		}() == true {
			__antithesis_instrumentation__.Notify(270882)
			return
		} else {
			__antithesis_instrumentation__.Notify(270883)
		}
		__antithesis_instrumentation__.Notify(270880)
		nonVirtualCols = append(nonVirtualCols, col)
	}
	__antithesis_instrumentation__.Notify(270871)
	for i := range desc.Columns {
		__antithesis_instrumentation__.Notify(270884)
		maybeAddCol(&desc.Columns[i])
	}
	__antithesis_instrumentation__.Notify(270872)
	for _, m := range desc.Mutations {
		__antithesis_instrumentation__.Notify(270885)
		maybeAddCol(m.GetColumn())
	}
	__antithesis_instrumentation__.Notify(270873)

	newStoreColumnIDs := make([]descpb.ColumnID, 0, len(nonVirtualCols))
	newStoreColumnNames := make([]string, 0, len(nonVirtualCols))
	keyColIDs := catalog.TableColSet{}
	for _, colID := range desc.PrimaryIndex.KeyColumnIDs {
		__antithesis_instrumentation__.Notify(270886)
		keyColIDs.Add(colID)
	}
	__antithesis_instrumentation__.Notify(270874)
	for _, col := range nonVirtualCols {
		__antithesis_instrumentation__.Notify(270887)
		if keyColIDs.Contains(col.ID) {
			__antithesis_instrumentation__.Notify(270889)
			continue
		} else {
			__antithesis_instrumentation__.Notify(270890)
		}
		__antithesis_instrumentation__.Notify(270888)
		newStoreColumnIDs = append(newStoreColumnIDs, col.ID)
		newStoreColumnNames = append(newStoreColumnNames, col.Name)
	}
	__antithesis_instrumentation__.Notify(270875)
	if len(newStoreColumnIDs) == 0 {
		__antithesis_instrumentation__.Notify(270891)
		newStoreColumnIDs = nil
		newStoreColumnNames = nil
	} else {
		__antithesis_instrumentation__.Notify(270892)
	}
	__antithesis_instrumentation__.Notify(270876)
	desc.PrimaryIndex.StoreColumnIDs = newStoreColumnIDs
	desc.PrimaryIndex.StoreColumnNames = newStoreColumnNames
	desc.PrimaryIndex.Version = descpb.PrimaryIndexWithStoredColumnsVersion
	return true
}

func maybeUpgradeSecondaryIndexFormatVersion(idx *descpb.IndexDescriptor) (hasChanged bool) {
	__antithesis_instrumentation__.Notify(270893)
	switch idx.Version {
	case descpb.SecondaryIndexFamilyFormatVersion:
		__antithesis_instrumentation__.Notify(270897)
		if idx.Type == descpb.IndexDescriptor_INVERTED {
			__antithesis_instrumentation__.Notify(270900)
			return false
		} else {
			__antithesis_instrumentation__.Notify(270901)
		}
	case descpb.EmptyArraysInInvertedIndexesVersion:
		__antithesis_instrumentation__.Notify(270898)
		break
	default:
		__antithesis_instrumentation__.Notify(270899)
		return false
	}
	__antithesis_instrumentation__.Notify(270894)
	slice := make([]descpb.ColumnID, 0, len(idx.KeyColumnIDs)+len(idx.KeySuffixColumnIDs)+len(idx.StoreColumnIDs))
	slice = append(slice, idx.KeyColumnIDs...)
	slice = append(slice, idx.KeySuffixColumnIDs...)
	slice = append(slice, idx.StoreColumnIDs...)
	set := catalog.MakeTableColSet(slice...)
	if len(slice) != set.Len() {
		__antithesis_instrumentation__.Notify(270902)
		return false
	} else {
		__antithesis_instrumentation__.Notify(270903)
	}
	__antithesis_instrumentation__.Notify(270895)
	if set.Contains(0) {
		__antithesis_instrumentation__.Notify(270904)
		return false
	} else {
		__antithesis_instrumentation__.Notify(270905)
	}
	__antithesis_instrumentation__.Notify(270896)
	idx.Version = descpb.StrictIndexColumnIDGuaranteesVersion
	return true
}

func maybeUpgradeNamespaceName(d *descpb.TableDescriptor) (hasChanged bool) {
	__antithesis_instrumentation__.Notify(270906)
	if d.ID != keys.NamespaceTableID || func() bool {
		__antithesis_instrumentation__.Notify(270908)
		return d.Name != catconstants.PreMigrationNamespaceTableName == true
	}() == true {
		__antithesis_instrumentation__.Notify(270909)
		return false
	} else {
		__antithesis_instrumentation__.Notify(270910)
	}
	__antithesis_instrumentation__.Notify(270907)
	d.Name = string(catconstants.NamespaceTableName)
	return true
}

func maybeFixPrimaryIndexEncoding(idx *descpb.IndexDescriptor) (hasChanged bool) {
	__antithesis_instrumentation__.Notify(270911)
	if idx.EncodingType == descpb.PrimaryIndexEncoding {
		__antithesis_instrumentation__.Notify(270913)
		return false
	} else {
		__antithesis_instrumentation__.Notify(270914)
	}
	__antithesis_instrumentation__.Notify(270912)
	idx.EncodingType = descpb.PrimaryIndexEncoding
	return true
}

func maybeRemoveDuplicateIDsInRefs(d *descpb.TableDescriptor) (hasChanged bool) {
	__antithesis_instrumentation__.Notify(270915)

	if s := cleanedIDs(d.DependsOn); len(s) < len(d.DependsOn) {
		__antithesis_instrumentation__.Notify(270920)
		d.DependsOn = s
		hasChanged = true
	} else {
		__antithesis_instrumentation__.Notify(270921)
	}
	__antithesis_instrumentation__.Notify(270916)

	if s := cleanedIDs(d.DependsOnTypes); len(s) < len(d.DependsOnTypes) {
		__antithesis_instrumentation__.Notify(270922)
		d.DependsOnTypes = s
		hasChanged = true
	} else {
		__antithesis_instrumentation__.Notify(270923)
	}
	__antithesis_instrumentation__.Notify(270917)

	for i := range d.DependedOnBy {
		__antithesis_instrumentation__.Notify(270924)
		ref := &d.DependedOnBy[i]
		s := catalog.MakeTableColSet(ref.ColumnIDs...).Ordered()

		if len(s) > 1 && func() bool {
			__antithesis_instrumentation__.Notify(270926)
			return s[0] == 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(270927)
			s = s[1:]
		} else {
			__antithesis_instrumentation__.Notify(270928)
		}
		__antithesis_instrumentation__.Notify(270925)
		if len(s) < len(ref.ColumnIDs) {
			__antithesis_instrumentation__.Notify(270929)
			ref.ColumnIDs = s
			hasChanged = true
		} else {
			__antithesis_instrumentation__.Notify(270930)
		}
	}
	__antithesis_instrumentation__.Notify(270918)

	for i := range d.Columns {
		__antithesis_instrumentation__.Notify(270931)
		col := &d.Columns[i]
		if s := cleanedIDs(col.UsesSequenceIds); len(s) < len(col.UsesSequenceIds) {
			__antithesis_instrumentation__.Notify(270933)
			col.UsesSequenceIds = s
			hasChanged = true
		} else {
			__antithesis_instrumentation__.Notify(270934)
		}
		__antithesis_instrumentation__.Notify(270932)
		if s := cleanedIDs(col.OwnsSequenceIds); len(s) < len(col.OwnsSequenceIds) {
			__antithesis_instrumentation__.Notify(270935)
			col.OwnsSequenceIds = s
			hasChanged = true
		} else {
			__antithesis_instrumentation__.Notify(270936)
		}
	}
	__antithesis_instrumentation__.Notify(270919)
	return hasChanged
}

func cleanedIDs(input []descpb.ID) []descpb.ID {
	__antithesis_instrumentation__.Notify(270937)
	s := catalog.MakeDescriptorIDSet(input...).Ordered()
	if len(s) == 0 {
		__antithesis_instrumentation__.Notify(270939)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(270940)
	}
	__antithesis_instrumentation__.Notify(270938)
	return s
}

func maybeAddConstraintIDs(desc *descpb.TableDescriptor) (hasChanged bool) {
	__antithesis_instrumentation__.Notify(270941)

	if !desc.IsTable() {
		__antithesis_instrumentation__.Notify(270952)
		return false
	} else {
		__antithesis_instrumentation__.Notify(270953)
	}
	__antithesis_instrumentation__.Notify(270942)
	initialConstraintID := desc.NextConstraintID

	constraintIndexes := make(map[descpb.IndexID]*descpb.IndexDescriptor)
	if desc.NextConstraintID == 0 {
		__antithesis_instrumentation__.Notify(270954)
		desc.NextConstraintID = 1
	} else {
		__antithesis_instrumentation__.Notify(270955)
	}
	__antithesis_instrumentation__.Notify(270943)
	nextConstraintID := func() descpb.ConstraintID {
		__antithesis_instrumentation__.Notify(270956)
		id := desc.GetNextConstraintID()
		desc.NextConstraintID++
		return id
	}
	__antithesis_instrumentation__.Notify(270944)

	if desc.PrimaryIndex.ConstraintID == 0 {
		__antithesis_instrumentation__.Notify(270957)
		desc.PrimaryIndex.ConstraintID = nextConstraintID()
		constraintIndexes[desc.PrimaryIndex.ID] = &desc.PrimaryIndex
	} else {
		__antithesis_instrumentation__.Notify(270958)
	}
	__antithesis_instrumentation__.Notify(270945)
	for i := range desc.Indexes {
		__antithesis_instrumentation__.Notify(270959)
		idx := &desc.Indexes[i]
		if idx.Unique && func() bool {
			__antithesis_instrumentation__.Notify(270960)
			return idx.ConstraintID == 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(270961)
			idx.ConstraintID = nextConstraintID()
			constraintIndexes[idx.ID] = idx
		} else {
			__antithesis_instrumentation__.Notify(270962)
		}
	}
	__antithesis_instrumentation__.Notify(270946)
	for i := range desc.Checks {
		__antithesis_instrumentation__.Notify(270963)
		check := desc.Checks[i]
		if check.ConstraintID == 0 {
			__antithesis_instrumentation__.Notify(270964)
			check.ConstraintID = nextConstraintID()
		} else {
			__antithesis_instrumentation__.Notify(270965)
		}
	}
	__antithesis_instrumentation__.Notify(270947)
	for i := range desc.InboundFKs {
		__antithesis_instrumentation__.Notify(270966)
		fk := &desc.InboundFKs[i]
		if fk.ConstraintID == 0 {
			__antithesis_instrumentation__.Notify(270967)
			fk.ConstraintID = nextConstraintID()
		} else {
			__antithesis_instrumentation__.Notify(270968)
		}
	}
	__antithesis_instrumentation__.Notify(270948)
	for i := range desc.OutboundFKs {
		__antithesis_instrumentation__.Notify(270969)
		fk := &desc.OutboundFKs[i]
		if fk.ConstraintID == 0 {
			__antithesis_instrumentation__.Notify(270970)
			fk.ConstraintID = nextConstraintID()
		} else {
			__antithesis_instrumentation__.Notify(270971)
		}
	}
	__antithesis_instrumentation__.Notify(270949)
	for i := range desc.UniqueWithoutIndexConstraints {
		__antithesis_instrumentation__.Notify(270972)
		unique := desc.UniqueWithoutIndexConstraints[i]
		if unique.ConstraintID == 0 {
			__antithesis_instrumentation__.Notify(270973)
			unique.ConstraintID = nextConstraintID()
		} else {
			__antithesis_instrumentation__.Notify(270974)
		}
	}
	__antithesis_instrumentation__.Notify(270950)

	for _, mutation := range desc.GetMutations() {
		__antithesis_instrumentation__.Notify(270975)
		if idx := mutation.GetIndex(); idx != nil && func() bool {
			__antithesis_instrumentation__.Notify(270976)
			return idx.ConstraintID == 0 == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(270977)
			return mutation.Direction == descpb.DescriptorMutation_ADD == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(270978)
			return idx.Unique == true
		}() == true {
			__antithesis_instrumentation__.Notify(270979)
			idx.ConstraintID = nextConstraintID()
			constraintIndexes[idx.ID] = idx
		} else {
			__antithesis_instrumentation__.Notify(270980)
			if pkSwap := mutation.GetPrimaryKeySwap(); pkSwap != nil {
				__antithesis_instrumentation__.Notify(270981)
				for idx := range pkSwap.NewIndexes {
					__antithesis_instrumentation__.Notify(270982)
					oldIdx, firstOk := constraintIndexes[pkSwap.OldIndexes[idx]]
					newIdx := constraintIndexes[pkSwap.NewIndexes[idx]]
					if !firstOk {
						__antithesis_instrumentation__.Notify(270984)
						continue
					} else {
						__antithesis_instrumentation__.Notify(270985)
					}
					__antithesis_instrumentation__.Notify(270983)
					newIdx.ConstraintID = oldIdx.ConstraintID
				}
			} else {
				__antithesis_instrumentation__.Notify(270986)
				if constraint := mutation.GetConstraint(); constraint != nil {
					__antithesis_instrumentation__.Notify(270987)
					nextID := nextConstraintID()
					constraint.UniqueWithoutIndexConstraint.ConstraintID = nextID
					constraint.ForeignKey.ConstraintID = nextID
					constraint.Check.ConstraintID = nextID
				} else {
					__antithesis_instrumentation__.Notify(270988)
				}
			}
		}

	}
	__antithesis_instrumentation__.Notify(270951)
	return desc.NextConstraintID != initialConstraintID
}
