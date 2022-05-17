package tabledesc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"sort"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catprivilege"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/multiregion"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemaexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/cockroach/pkg/util/interval"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"
)

func (desc *wrapper) ValidateTxnCommit(
	vea catalog.ValidationErrorAccumulator, _ catalog.ValidationDescGetter,
) {
	__antithesis_instrumentation__.Notify(271046)

	if !desc.HasPrimaryKey() {
		__antithesis_instrumentation__.Notify(271048)
		vea.Report(unimplemented.NewWithIssue(48026,
			"primary key dropped without subsequent addition of new primary key in same transaction"))
	} else {
		__antithesis_instrumentation__.Notify(271049)
	}
	__antithesis_instrumentation__.Notify(271047)

	if n := len(desc.Mutations); n > 0 && func() bool {
		__antithesis_instrumentation__.Notify(271050)
		return desc.GetDeclarativeSchemaChangerState() != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(271051)
		lastMutationID := desc.Mutations[n-1].MutationID
		if lastMutationID != desc.NextMutationID {
			__antithesis_instrumentation__.Notify(271052)
			vea.Report(errors.AssertionFailedf(
				"expected next mutation ID to be %d in table undergoing declarative schema change, found %d instead",
				lastMutationID, desc.NextMutationID))
		} else {
			__antithesis_instrumentation__.Notify(271053)
		}
	} else {
		__antithesis_instrumentation__.Notify(271054)
	}
}

func (desc *wrapper) GetReferencedDescIDs() (catalog.DescriptorIDSet, error) {
	__antithesis_instrumentation__.Notify(271055)
	ids := catalog.MakeDescriptorIDSet(desc.GetID(), desc.GetParentID())

	if desc.GetParentSchemaID() != keys.PublicSchemaID {
		__antithesis_instrumentation__.Notify(271065)
		ids.Add(desc.GetParentSchemaID())
	} else {
		__antithesis_instrumentation__.Notify(271066)
	}
	__antithesis_instrumentation__.Notify(271056)

	for _, fk := range desc.OutboundFKs {
		__antithesis_instrumentation__.Notify(271067)
		ids.Add(fk.ReferencedTableID)
	}
	__antithesis_instrumentation__.Notify(271057)
	for _, fk := range desc.InboundFKs {
		__antithesis_instrumentation__.Notify(271068)
		ids.Add(fk.OriginTableID)
	}
	__antithesis_instrumentation__.Notify(271058)

	for _, col := range desc.DeletableColumns() {
		__antithesis_instrumentation__.Notify(271069)
		children, err := typedesc.GetTypeDescriptorClosure(col.GetType())
		if err != nil {
			__antithesis_instrumentation__.Notify(271072)
			return catalog.DescriptorIDSet{}, err
		} else {
			__antithesis_instrumentation__.Notify(271073)
		}
		__antithesis_instrumentation__.Notify(271070)
		for id := range children {
			__antithesis_instrumentation__.Notify(271074)
			ids.Add(id)
		}
		__antithesis_instrumentation__.Notify(271071)
		for i := 0; i < col.NumUsesSequences(); i++ {
			__antithesis_instrumentation__.Notify(271075)
			ids.Add(col.GetUsesSequenceID(i))
		}
	}
	__antithesis_instrumentation__.Notify(271059)

	visitor := &tree.TypeCollectorVisitor{OIDs: make(map[oid.Oid]struct{})}
	_ = ForEachExprStringInTableDesc(desc, func(expr *string) error {
		__antithesis_instrumentation__.Notify(271076)
		if parsedExpr, err := parser.ParseExpr(*expr); err == nil {
			__antithesis_instrumentation__.Notify(271078)

			tree.WalkExpr(visitor, parsedExpr)
		} else {
			__antithesis_instrumentation__.Notify(271079)
		}
		__antithesis_instrumentation__.Notify(271077)
		return nil
	})
	__antithesis_instrumentation__.Notify(271060)

	for oid := range visitor.OIDs {
		__antithesis_instrumentation__.Notify(271080)
		if !types.IsOIDUserDefinedType(oid) {
			__antithesis_instrumentation__.Notify(271083)
			continue
		} else {
			__antithesis_instrumentation__.Notify(271084)
		}
		__antithesis_instrumentation__.Notify(271081)
		id, err := typedesc.UserDefinedTypeOIDToID(oid)
		if err != nil {
			__antithesis_instrumentation__.Notify(271085)
			return catalog.DescriptorIDSet{}, err
		} else {
			__antithesis_instrumentation__.Notify(271086)
		}
		__antithesis_instrumentation__.Notify(271082)
		ids.Add(id)
	}
	__antithesis_instrumentation__.Notify(271061)

	for _, id := range desc.GetDependsOn() {
		__antithesis_instrumentation__.Notify(271087)
		ids.Add(id)
	}
	__antithesis_instrumentation__.Notify(271062)
	for _, id := range desc.GetDependsOnTypes() {
		__antithesis_instrumentation__.Notify(271088)
		ids.Add(id)
	}
	__antithesis_instrumentation__.Notify(271063)
	for _, ref := range desc.GetDependedOnBy() {
		__antithesis_instrumentation__.Notify(271089)
		ids.Add(ref.ID)
	}
	__antithesis_instrumentation__.Notify(271064)

	return ids, nil
}

func (desc *wrapper) ValidateCrossReferences(
	vea catalog.ValidationErrorAccumulator, vdg catalog.ValidationDescGetter,
) {
	__antithesis_instrumentation__.Notify(271090)

	dbDesc, err := vdg.GetDatabaseDescriptor(desc.GetParentID())
	if err != nil {
		__antithesis_instrumentation__.Notify(271099)
		vea.Report(err)
	} else {
		__antithesis_instrumentation__.Notify(271100)
		if dbDesc.Dropped() {
			__antithesis_instrumentation__.Notify(271101)
			vea.Report(errors.AssertionFailedf("parent database %q (%d) is dropped",
				dbDesc.GetName(), dbDesc.GetID()))
		} else {
			__antithesis_instrumentation__.Notify(271102)
		}
	}
	__antithesis_instrumentation__.Notify(271091)

	if desc.GetParentSchemaID() != keys.PublicSchemaID && func() bool {
		__antithesis_instrumentation__.Notify(271103)
		return !desc.IsTemporary() == true
	}() == true {
		__antithesis_instrumentation__.Notify(271104)
		schemaDesc, err := vdg.GetSchemaDescriptor(desc.GetParentSchemaID())
		if err != nil {
			__antithesis_instrumentation__.Notify(271107)
			vea.Report(err)
		} else {
			__antithesis_instrumentation__.Notify(271108)
		}
		__antithesis_instrumentation__.Notify(271105)
		if schemaDesc != nil && func() bool {
			__antithesis_instrumentation__.Notify(271109)
			return dbDesc != nil == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(271110)
			return schemaDesc.GetParentID() != dbDesc.GetID() == true
		}() == true {
			__antithesis_instrumentation__.Notify(271111)
			vea.Report(errors.AssertionFailedf("parent schema %d is in different database %d",
				desc.GetParentSchemaID(), schemaDesc.GetParentID()))
		} else {
			__antithesis_instrumentation__.Notify(271112)
		}
		__antithesis_instrumentation__.Notify(271106)
		if schemaDesc != nil && func() bool {
			__antithesis_instrumentation__.Notify(271113)
			return schemaDesc.Dropped() == true
		}() == true {
			__antithesis_instrumentation__.Notify(271114)
			vea.Report(errors.AssertionFailedf("parent schema %q (%d) is dropped",
				schemaDesc.GetName(), schemaDesc.GetID()))
		} else {
			__antithesis_instrumentation__.Notify(271115)
		}
	} else {
		__antithesis_instrumentation__.Notify(271116)
	}
	__antithesis_instrumentation__.Notify(271092)

	if dbDesc != nil {
		__antithesis_instrumentation__.Notify(271117)

		typeIDs, _, err := desc.GetAllReferencedTypeIDs(dbDesc, vdg.GetTypeDescriptor)
		if err != nil {
			__antithesis_instrumentation__.Notify(271119)
			vea.Report(err)
		} else {
			__antithesis_instrumentation__.Notify(271120)
			for _, id := range typeIDs {
				__antithesis_instrumentation__.Notify(271121)
				_, err := vdg.GetTypeDescriptor(id)
				vea.Report(err)
			}
		}
		__antithesis_instrumentation__.Notify(271118)

		if err := multiregion.ValidateTableLocalityConfig(desc, dbDesc, vdg); err != nil {
			__antithesis_instrumentation__.Notify(271122)
			vea.Report(errors.Wrap(err, "invalid locality config"))
			return
		} else {
			__antithesis_instrumentation__.Notify(271123)
		}
	} else {
		__antithesis_instrumentation__.Notify(271124)
	}
	__antithesis_instrumentation__.Notify(271093)

	if desc.IsView() {
		__antithesis_instrumentation__.Notify(271125)
		for _, id := range desc.DependsOn {
			__antithesis_instrumentation__.Notify(271127)
			vea.Report(desc.validateOutboundTableRef(id, vdg))
		}
		__antithesis_instrumentation__.Notify(271126)
		for _, id := range desc.DependsOnTypes {
			__antithesis_instrumentation__.Notify(271128)
			vea.Report(desc.validateOutboundTypeRef(id, vdg))
		}
	} else {
		__antithesis_instrumentation__.Notify(271129)
	}
	__antithesis_instrumentation__.Notify(271094)

	for _, by := range desc.DependedOnBy {
		__antithesis_instrumentation__.Notify(271130)
		vea.Report(desc.validateInboundTableRef(by, vdg))
	}
	__antithesis_instrumentation__.Notify(271095)

	if desc.HasRowLevelTTL() {
		__antithesis_instrumentation__.Notify(271131)
		pk := desc.GetPrimaryIndex()
		if col, err := desc.FindColumnWithName(colinfo.TTLDefaultExpirationColumnName); err != nil {
			__antithesis_instrumentation__.Notify(271134)
			vea.Report(errors.Wrapf(err, "expected column %s", colinfo.TTLDefaultExpirationColumnName))
		} else {
			__antithesis_instrumentation__.Notify(271135)
			intervalExpr := desc.GetRowLevelTTL().DurationExpr
			expectedStr := `current_timestamp():::TIMESTAMPTZ + ` + string(intervalExpr)
			if col.GetDefaultExpr() != expectedStr {
				__antithesis_instrumentation__.Notify(271137)
				vea.Report(pgerror.Newf(
					pgcode.InvalidTableDefinition,
					"expected DEFAULT expression of %s to be %s",
					colinfo.TTLDefaultExpirationColumnName,
					expectedStr,
				))
			} else {
				__antithesis_instrumentation__.Notify(271138)
			}
			__antithesis_instrumentation__.Notify(271136)
			if col.GetOnUpdateExpr() != expectedStr {
				__antithesis_instrumentation__.Notify(271139)
				vea.Report(pgerror.Newf(
					pgcode.InvalidTableDefinition,
					"expected ON UPDATE expression of %s to be %s",
					colinfo.TTLDefaultExpirationColumnName,
					expectedStr,
				))
			} else {
				__antithesis_instrumentation__.Notify(271140)
			}
		}
		__antithesis_instrumentation__.Notify(271132)

		for i := 0; i < pk.NumKeyColumns(); i++ {
			__antithesis_instrumentation__.Notify(271141)
			dir := pk.GetKeyColumnDirection(i)
			if dir != descpb.IndexDescriptor_ASC {
				__antithesis_instrumentation__.Notify(271142)
				vea.Report(unimplemented.NewWithIssuef(
					76912,
					`non-ascending ordering on PRIMARY KEYs are not supported`,
				))
			} else {
				__antithesis_instrumentation__.Notify(271143)
			}
		}
		__antithesis_instrumentation__.Notify(271133)
		if len(desc.OutboundFKs) > 0 || func() bool {
			__antithesis_instrumentation__.Notify(271144)
			return len(desc.InboundFKs) > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(271145)
			vea.Report(unimplemented.NewWithIssuef(
				76407,
				`foreign keys to/from table with TTL "%s" are not permitted`,
				desc.Name,
			))
		} else {
			__antithesis_instrumentation__.Notify(271146)
		}
	} else {
		__antithesis_instrumentation__.Notify(271147)
	}
	__antithesis_instrumentation__.Notify(271096)

	for i := range desc.OutboundFKs {
		__antithesis_instrumentation__.Notify(271148)
		vea.Report(desc.validateOutboundFK(&desc.OutboundFKs[i], vdg))
	}
	__antithesis_instrumentation__.Notify(271097)
	for i := range desc.InboundFKs {
		__antithesis_instrumentation__.Notify(271149)
		vea.Report(desc.validateInboundFK(&desc.InboundFKs[i], vdg))
	}
	__antithesis_instrumentation__.Notify(271098)

	if desc.PartitionAllBy {
		__antithesis_instrumentation__.Notify(271150)
		for _, indexI := range desc.ActiveIndexes() {
			__antithesis_instrumentation__.Notify(271151)
			if !desc.matchingPartitionbyAll(indexI) {
				__antithesis_instrumentation__.Notify(271152)
				vea.Report(errors.AssertionFailedf(
					"table has PARTITION ALL BY defined, but index %s does not have matching PARTITION BY",
					indexI.GetName(),
				))
			} else {
				__antithesis_instrumentation__.Notify(271153)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(271154)
	}
}

func (desc *wrapper) validateOutboundTableRef(
	id descpb.ID, vdg catalog.ValidationDescGetter,
) error {
	__antithesis_instrumentation__.Notify(271155)
	referencedTable, err := vdg.GetTableDescriptor(id)
	if err != nil {
		__antithesis_instrumentation__.Notify(271159)
		return errors.NewAssertionErrorWithWrappedErrf(err, "invalid depends-on relation reference")
	} else {
		__antithesis_instrumentation__.Notify(271160)
	}
	__antithesis_instrumentation__.Notify(271156)
	if referencedTable.Dropped() {
		__antithesis_instrumentation__.Notify(271161)
		return errors.AssertionFailedf("depends-on relation %q (%d) is dropped",
			referencedTable.GetName(), referencedTable.GetID())
	} else {
		__antithesis_instrumentation__.Notify(271162)
	}
	__antithesis_instrumentation__.Notify(271157)
	for _, by := range referencedTable.TableDesc().DependedOnBy {
		__antithesis_instrumentation__.Notify(271163)
		if by.ID == desc.GetID() {
			__antithesis_instrumentation__.Notify(271164)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(271165)
		}
	}
	__antithesis_instrumentation__.Notify(271158)
	return errors.AssertionFailedf("depends-on relation %q (%d) has no corresponding depended-on-by back reference",
		referencedTable.GetName(), id)
}

func (desc *wrapper) validateOutboundTypeRef(id descpb.ID, vdg catalog.ValidationDescGetter) error {
	__antithesis_instrumentation__.Notify(271166)
	typ, err := vdg.GetTypeDescriptor(id)
	if err != nil {
		__antithesis_instrumentation__.Notify(271169)
		return errors.NewAssertionErrorWithWrappedErrf(err, "invalid depends-on type reference")
	} else {
		__antithesis_instrumentation__.Notify(271170)
	}
	__antithesis_instrumentation__.Notify(271167)
	if typ.Dropped() {
		__antithesis_instrumentation__.Notify(271171)
		return errors.AssertionFailedf("depends-on type %q (%d) is dropped",
			typ.GetName(), typ.GetID())
	} else {
		__antithesis_instrumentation__.Notify(271172)
	}
	__antithesis_instrumentation__.Notify(271168)

	return nil
}

func (desc *wrapper) validateInboundTableRef(
	by descpb.TableDescriptor_Reference, vdg catalog.ValidationDescGetter,
) error {
	__antithesis_instrumentation__.Notify(271173)
	backReferencedTable, err := vdg.GetTableDescriptor(by.ID)
	if err != nil {
		__antithesis_instrumentation__.Notify(271179)
		return errors.NewAssertionErrorWithWrappedErrf(err, "invalid depended-on-by relation back reference")
	} else {
		__antithesis_instrumentation__.Notify(271180)
	}
	__antithesis_instrumentation__.Notify(271174)
	if backReferencedTable.Dropped() {
		__antithesis_instrumentation__.Notify(271181)
		return errors.AssertionFailedf("depended-on-by relation %q (%d) is dropped",
			backReferencedTable.GetName(), backReferencedTable.GetID())
	} else {
		__antithesis_instrumentation__.Notify(271182)
	}
	__antithesis_instrumentation__.Notify(271175)
	if desc.IsSequence() {
		__antithesis_instrumentation__.Notify(271183)

		for _, colID := range by.ColumnIDs {
			__antithesis_instrumentation__.Notify(271184)
			col, _ := backReferencedTable.FindColumnWithID(colID)
			if col == nil {
				__antithesis_instrumentation__.Notify(271188)
				return errors.AssertionFailedf("depended-on-by relation %q (%d) does not have a column with ID %d",
					backReferencedTable.GetName(), by.ID, colID)
			} else {
				__antithesis_instrumentation__.Notify(271189)
			}
			__antithesis_instrumentation__.Notify(271185)
			var found bool
			for i := 0; i < col.NumUsesSequences(); i++ {
				__antithesis_instrumentation__.Notify(271190)
				if col.GetUsesSequenceID(i) == desc.GetID() {
					__antithesis_instrumentation__.Notify(271191)
					found = true
					break
				} else {
					__antithesis_instrumentation__.Notify(271192)
				}
			}
			__antithesis_instrumentation__.Notify(271186)
			if found {
				__antithesis_instrumentation__.Notify(271193)
				continue
			} else {
				__antithesis_instrumentation__.Notify(271194)
			}
			__antithesis_instrumentation__.Notify(271187)
			return errors.AssertionFailedf(
				"depended-on-by relation %q (%d) has no reference to this sequence in column %q (%d)",
				backReferencedTable.GetName(), by.ID, col.GetName(), col.GetID())
		}
	} else {
		__antithesis_instrumentation__.Notify(271195)
	}
	__antithesis_instrumentation__.Notify(271176)

	if !backReferencedTable.IsView() {
		__antithesis_instrumentation__.Notify(271196)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(271197)
	}
	__antithesis_instrumentation__.Notify(271177)
	for _, id := range backReferencedTable.TableDesc().DependsOn {
		__antithesis_instrumentation__.Notify(271198)
		if id == desc.GetID() {
			__antithesis_instrumentation__.Notify(271199)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(271200)
		}
	}
	__antithesis_instrumentation__.Notify(271178)
	return errors.AssertionFailedf("depended-on-by view %q (%d) has no corresponding depends-on forward reference",
		backReferencedTable.GetName(), by.ID)
}

func (desc *wrapper) validateOutboundFK(
	fk *descpb.ForeignKeyConstraint, vdg catalog.ValidationDescGetter,
) error {
	__antithesis_instrumentation__.Notify(271201)
	referencedTable, err := vdg.GetTableDescriptor(fk.ReferencedTableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(271206)
		return errors.Wrapf(err,
			"invalid foreign key: missing table=%d", fk.ReferencedTableID)
	} else {
		__antithesis_instrumentation__.Notify(271207)
	}
	__antithesis_instrumentation__.Notify(271202)
	if referencedTable.Dropped() {
		__antithesis_instrumentation__.Notify(271208)
		return errors.AssertionFailedf("referenced table %q (%d) is dropped",
			referencedTable.GetName(), referencedTable.GetID())
	} else {
		__antithesis_instrumentation__.Notify(271209)
	}
	__antithesis_instrumentation__.Notify(271203)
	found := false
	_ = referencedTable.ForeachInboundFK(func(backref *descpb.ForeignKeyConstraint) error {
		__antithesis_instrumentation__.Notify(271210)
		if !found && func() bool {
			__antithesis_instrumentation__.Notify(271212)
			return backref.OriginTableID == desc.ID == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(271213)
			return backref.Name == fk.Name == true
		}() == true {
			__antithesis_instrumentation__.Notify(271214)
			found = true
		} else {
			__antithesis_instrumentation__.Notify(271215)
		}
		__antithesis_instrumentation__.Notify(271211)
		return nil
	})
	__antithesis_instrumentation__.Notify(271204)
	if found {
		__antithesis_instrumentation__.Notify(271216)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(271217)
	}
	__antithesis_instrumentation__.Notify(271205)
	return errors.AssertionFailedf("missing fk back reference %q to %q from %q",
		fk.Name, desc.Name, referencedTable.GetName())
}

func (desc *wrapper) validateInboundFK(
	backref *descpb.ForeignKeyConstraint, vdg catalog.ValidationDescGetter,
) error {
	__antithesis_instrumentation__.Notify(271218)
	originTable, err := vdg.GetTableDescriptor(backref.OriginTableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(271223)
		return errors.Wrapf(err,
			"invalid foreign key backreference: missing table=%d", backref.OriginTableID)
	} else {
		__antithesis_instrumentation__.Notify(271224)
	}
	__antithesis_instrumentation__.Notify(271219)
	if originTable.Dropped() {
		__antithesis_instrumentation__.Notify(271225)
		return errors.AssertionFailedf("origin table %q (%d) is dropped",
			originTable.GetName(), originTable.GetID())
	} else {
		__antithesis_instrumentation__.Notify(271226)
	}
	__antithesis_instrumentation__.Notify(271220)
	found := false
	_ = originTable.ForeachOutboundFK(func(fk *descpb.ForeignKeyConstraint) error {
		__antithesis_instrumentation__.Notify(271227)
		if !found && func() bool {
			__antithesis_instrumentation__.Notify(271229)
			return fk.ReferencedTableID == desc.ID == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(271230)
			return fk.Name == backref.Name == true
		}() == true {
			__antithesis_instrumentation__.Notify(271231)
			found = true
		} else {
			__antithesis_instrumentation__.Notify(271232)
		}
		__antithesis_instrumentation__.Notify(271228)
		return nil
	})
	__antithesis_instrumentation__.Notify(271221)
	if found {
		__antithesis_instrumentation__.Notify(271233)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(271234)
	}
	__antithesis_instrumentation__.Notify(271222)
	return errors.AssertionFailedf("missing fk forward reference %q to %q from %q",
		backref.Name, desc.Name, originTable.GetName())
}

func (desc *wrapper) matchingPartitionbyAll(indexI catalog.Index) bool {
	__antithesis_instrumentation__.Notify(271235)
	primaryIndexPartitioning := desc.PrimaryIndex.KeyColumnIDs[:desc.PrimaryIndex.Partitioning.NumColumns]
	indexPartitioning := indexI.IndexDesc().KeyColumnIDs[:indexI.GetPartitioning().NumColumns()]
	if len(primaryIndexPartitioning) != len(indexPartitioning) {
		__antithesis_instrumentation__.Notify(271238)
		return false
	} else {
		__antithesis_instrumentation__.Notify(271239)
	}
	__antithesis_instrumentation__.Notify(271236)
	for i, id := range primaryIndexPartitioning {
		__antithesis_instrumentation__.Notify(271240)
		if id != indexPartitioning[i] {
			__antithesis_instrumentation__.Notify(271241)
			return false
		} else {
			__antithesis_instrumentation__.Notify(271242)
		}
	}
	__antithesis_instrumentation__.Notify(271237)
	return true
}

func validateMutation(m *descpb.DescriptorMutation) error {
	__antithesis_instrumentation__.Notify(271243)
	unSetEnums := m.State == descpb.DescriptorMutation_UNKNOWN || func() bool {
		__antithesis_instrumentation__.Notify(271246)
		return m.Direction == descpb.DescriptorMutation_NONE == true
	}() == true
	switch desc := m.Descriptor_.(type) {
	case *descpb.DescriptorMutation_Column:
		__antithesis_instrumentation__.Notify(271247)
		col := desc.Column
		if unSetEnums {
			__antithesis_instrumentation__.Notify(271255)
			return errors.AssertionFailedf(
				"mutation in state %s, direction %s, col %q, id %v",
				errors.Safe(m.State), errors.Safe(m.Direction), col.Name, errors.Safe(col.ID))
		} else {
			__antithesis_instrumentation__.Notify(271256)
		}
	case *descpb.DescriptorMutation_Index:
		__antithesis_instrumentation__.Notify(271248)
		if unSetEnums {
			__antithesis_instrumentation__.Notify(271257)
			idx := desc.Index
			return errors.AssertionFailedf(
				"mutation in state %s, direction %s, index %s, id %v",
				errors.Safe(m.State), errors.Safe(m.Direction), idx.Name, errors.Safe(idx.ID))
		} else {
			__antithesis_instrumentation__.Notify(271258)
		}
	case *descpb.DescriptorMutation_Constraint:
		__antithesis_instrumentation__.Notify(271249)
		if unSetEnums {
			__antithesis_instrumentation__.Notify(271259)
			return errors.AssertionFailedf(
				"mutation in state %s, direction %s, constraint %v",
				errors.Safe(m.State), errors.Safe(m.Direction), desc.Constraint.Name)
		} else {
			__antithesis_instrumentation__.Notify(271260)
		}
	case *descpb.DescriptorMutation_PrimaryKeySwap:
		__antithesis_instrumentation__.Notify(271250)
		if m.Direction == descpb.DescriptorMutation_NONE {
			__antithesis_instrumentation__.Notify(271261)
			return errors.AssertionFailedf(
				"primary key swap mutation in state %s, direction %s", errors.Safe(m.State), errors.Safe(m.Direction))
		} else {
			__antithesis_instrumentation__.Notify(271262)
		}
	case *descpb.DescriptorMutation_ComputedColumnSwap:
		__antithesis_instrumentation__.Notify(271251)
		if m.Direction == descpb.DescriptorMutation_NONE {
			__antithesis_instrumentation__.Notify(271263)
			return errors.AssertionFailedf(
				"computed column swap mutation in state %s, direction %s", errors.Safe(m.State), errors.Safe(m.Direction))
		} else {
			__antithesis_instrumentation__.Notify(271264)
		}
	case *descpb.DescriptorMutation_MaterializedViewRefresh:
		__antithesis_instrumentation__.Notify(271252)
		if m.Direction == descpb.DescriptorMutation_NONE {
			__antithesis_instrumentation__.Notify(271265)
			return errors.AssertionFailedf(
				"materialized view refresh mutation in state %s, direction %s", errors.Safe(m.State), errors.Safe(m.Direction))
		} else {
			__antithesis_instrumentation__.Notify(271266)
		}
	case *descpb.DescriptorMutation_ModifyRowLevelTTL:
		__antithesis_instrumentation__.Notify(271253)
		if m.Direction == descpb.DescriptorMutation_NONE {
			__antithesis_instrumentation__.Notify(271267)
			return errors.AssertionFailedf(
				"modify row level TTL mutation in state %s, direction %s", errors.Safe(m.State), errors.Safe(m.Direction))
		} else {
			__antithesis_instrumentation__.Notify(271268)
		}
	default:
		__antithesis_instrumentation__.Notify(271254)
		return errors.AssertionFailedf(
			"mutation in state %s, direction %s, and no column/index descriptor",
			errors.Safe(m.State), errors.Safe(m.Direction))
	}
	__antithesis_instrumentation__.Notify(271244)

	switch m.State {
	case descpb.DescriptorMutation_BACKFILLING:
		__antithesis_instrumentation__.Notify(271269)
		if _, ok := m.Descriptor_.(*descpb.DescriptorMutation_Index); !ok {
			__antithesis_instrumentation__.Notify(271272)
			return errors.AssertionFailedf("non-index mutation in state %s", errors.Safe(m.State))
		} else {
			__antithesis_instrumentation__.Notify(271273)
		}
	case descpb.DescriptorMutation_MERGING:
		__antithesis_instrumentation__.Notify(271270)
		if _, ok := m.Descriptor_.(*descpb.DescriptorMutation_Index); !ok {
			__antithesis_instrumentation__.Notify(271274)
			return errors.AssertionFailedf("non-index mutation in state %s", errors.Safe(m.State))
		} else {
			__antithesis_instrumentation__.Notify(271275)
		}
	default:
		__antithesis_instrumentation__.Notify(271271)
	}
	__antithesis_instrumentation__.Notify(271245)

	return nil
}

func (desc *wrapper) ValidateSelf(vea catalog.ValidationErrorAccumulator) {
	__antithesis_instrumentation__.Notify(271276)

	vea.Report(catalog.ValidateName(desc.Name, "table"))
	if desc.GetID() == descpb.InvalidID {
		__antithesis_instrumentation__.Notify(271296)
		vea.Report(errors.AssertionFailedf("invalid table ID %d", desc.GetID()))
	} else {
		__antithesis_instrumentation__.Notify(271297)
	}
	__antithesis_instrumentation__.Notify(271277)
	if desc.GetParentSchemaID() == descpb.InvalidID {
		__antithesis_instrumentation__.Notify(271298)
		vea.Report(errors.AssertionFailedf("invalid parent schema ID %d", desc.GetParentSchemaID()))
	} else {
		__antithesis_instrumentation__.Notify(271299)
	}
	__antithesis_instrumentation__.Notify(271278)

	if desc.GetParentID() == descpb.InvalidID && func() bool {
		__antithesis_instrumentation__.Notify(271300)
		return !desc.IsVirtualTable() == true
	}() == true {
		__antithesis_instrumentation__.Notify(271301)
		vea.Report(errors.AssertionFailedf("invalid parent ID %d", desc.GetParentID()))
	} else {
		__antithesis_instrumentation__.Notify(271302)
	}
	__antithesis_instrumentation__.Notify(271279)

	if desc.Privileges == nil {
		__antithesis_instrumentation__.Notify(271303)
		vea.Report(errors.AssertionFailedf("privileges not set"))
	} else {
		__antithesis_instrumentation__.Notify(271304)
		vea.Report(catprivilege.Validate(*desc.Privileges, desc, privilege.Table))
	}
	__antithesis_instrumentation__.Notify(271280)

	for _, ref := range desc.DependedOnBy {
		__antithesis_instrumentation__.Notify(271305)
		if ref.ID == descpb.InvalidID {
			__antithesis_instrumentation__.Notify(271307)
			vea.Report(errors.AssertionFailedf(
				"invalid relation ID %d in depended-on-by references",
				ref.ID))
		} else {
			__antithesis_instrumentation__.Notify(271308)
		}
		__antithesis_instrumentation__.Notify(271306)
		if len(ref.ColumnIDs) > catalog.MakeTableColSet(ref.ColumnIDs...).Len() {
			__antithesis_instrumentation__.Notify(271309)
			vea.Report(errors.AssertionFailedf("duplicate column IDs found in depended-on-by references: %v",
				ref.ColumnIDs))
		} else {
			__antithesis_instrumentation__.Notify(271310)
		}
	}
	__antithesis_instrumentation__.Notify(271281)

	if !desc.IsView() {
		__antithesis_instrumentation__.Notify(271311)
		if len(desc.DependsOn) > 0 {
			__antithesis_instrumentation__.Notify(271313)
			vea.Report(errors.AssertionFailedf(
				"has depends-on references despite not being a view"))
		} else {
			__antithesis_instrumentation__.Notify(271314)
		}
		__antithesis_instrumentation__.Notify(271312)
		if len(desc.DependsOnTypes) > 0 {
			__antithesis_instrumentation__.Notify(271315)
			vea.Report(errors.AssertionFailedf(
				"has depends-on-types references despite not being a view"))
		} else {
			__antithesis_instrumentation__.Notify(271316)
		}
	} else {
		__antithesis_instrumentation__.Notify(271317)
	}
	__antithesis_instrumentation__.Notify(271282)

	if desc.IsSequence() {
		__antithesis_instrumentation__.Notify(271318)
		return
	} else {
		__antithesis_instrumentation__.Notify(271319)
	}
	__antithesis_instrumentation__.Notify(271283)

	for _, ref := range desc.DependedOnBy {
		__antithesis_instrumentation__.Notify(271320)
		if ref.IndexID != 0 {
			__antithesis_instrumentation__.Notify(271322)
			if idx, _ := desc.FindIndexWithID(ref.IndexID); idx == nil {
				__antithesis_instrumentation__.Notify(271323)
				vea.Report(errors.AssertionFailedf(
					"index ID %d found in depended-on-by references, no such index in this relation",
					ref.IndexID))
			} else {
				__antithesis_instrumentation__.Notify(271324)
			}
		} else {
			__antithesis_instrumentation__.Notify(271325)
		}
		__antithesis_instrumentation__.Notify(271321)
		for _, colID := range ref.ColumnIDs {
			__antithesis_instrumentation__.Notify(271326)
			if col, _ := desc.FindColumnWithID(colID); col == nil {
				__antithesis_instrumentation__.Notify(271327)
				vea.Report(errors.AssertionFailedf(
					"column ID %d found in depended-on-by references, no such column in this relation",
					colID))
			} else {
				__antithesis_instrumentation__.Notify(271328)
			}
		}
	}
	__antithesis_instrumentation__.Notify(271284)

	if len(desc.Columns) == 0 {
		__antithesis_instrumentation__.Notify(271329)
		vea.Report(ErrMissingColumns)
		return
	} else {
		__antithesis_instrumentation__.Notify(271330)
	}
	__antithesis_instrumentation__.Notify(271285)

	columnNames := make(map[string]descpb.ColumnID, len(desc.Columns))
	columnIDs := make(map[descpb.ColumnID]*descpb.ColumnDescriptor, len(desc.Columns))
	if err := desc.validateColumns(columnNames, columnIDs); err != nil {
		__antithesis_instrumentation__.Notify(271331)
		vea.Report(err)
		return
	} else {
		__antithesis_instrumentation__.Notify(271332)
	}
	__antithesis_instrumentation__.Notify(271286)

	if desc.IsVirtualTable() {
		__antithesis_instrumentation__.Notify(271333)
		return
	} else {
		__antithesis_instrumentation__.Notify(271334)
	}
	__antithesis_instrumentation__.Notify(271287)

	if desc.GetFormatVersion() < descpb.InterleavedFormatVersion {
		__antithesis_instrumentation__.Notify(271335)
		vea.Report(errors.AssertionFailedf(
			"table is encoded using using version %d, but this client only supports version %d",
			desc.GetFormatVersion(), descpb.InterleavedFormatVersion))
		return
	} else {
		__antithesis_instrumentation__.Notify(271336)
	}
	__antithesis_instrumentation__.Notify(271288)

	if err := desc.CheckUniqueConstraints(); err != nil {
		__antithesis_instrumentation__.Notify(271337)
		vea.Report(err)
		return
	} else {
		__antithesis_instrumentation__.Notify(271338)
	}
	__antithesis_instrumentation__.Notify(271289)

	mutationsHaveErrs := false
	for _, m := range desc.Mutations {
		__antithesis_instrumentation__.Notify(271339)
		if err := validateMutation(&m); err != nil {
			__antithesis_instrumentation__.Notify(271341)
			vea.Report(err)
			mutationsHaveErrs = true
		} else {
			__antithesis_instrumentation__.Notify(271342)
		}
		__antithesis_instrumentation__.Notify(271340)
		switch desc := m.Descriptor_.(type) {
		case *descpb.DescriptorMutation_Column:
			__antithesis_instrumentation__.Notify(271343)
			col := desc.Column
			columnIDs[col.ID] = col
		}
	}
	__antithesis_instrumentation__.Notify(271290)

	if mutationsHaveErrs {
		__antithesis_instrumentation__.Notify(271344)
		return
	} else {
		__antithesis_instrumentation__.Notify(271345)
	}
	__antithesis_instrumentation__.Notify(271291)

	if desc.IsPhysicalTable() {
		__antithesis_instrumentation__.Notify(271346)
		newErrs := []error{
			desc.validateColumnFamilies(columnIDs),
			desc.validateCheckConstraints(columnIDs),
			desc.validateUniqueWithoutIndexConstraints(columnIDs),
			desc.validateTableIndexes(columnNames, vea),
			desc.validatePartitioning(),
		}
		hasErrs := false
		for _, err := range newErrs {
			__antithesis_instrumentation__.Notify(271349)
			if err != nil {
				__antithesis_instrumentation__.Notify(271350)
				vea.Report(err)
				hasErrs = true
			} else {
				__antithesis_instrumentation__.Notify(271351)
			}
		}
		__antithesis_instrumentation__.Notify(271347)
		if hasErrs {
			__antithesis_instrumentation__.Notify(271352)
			return
		} else {
			__antithesis_instrumentation__.Notify(271353)
		}
		__antithesis_instrumentation__.Notify(271348)
		desc.validateConstraintIDs(vea)
	} else {
		__antithesis_instrumentation__.Notify(271354)
	}
	__antithesis_instrumentation__.Notify(271292)

	var alterPKMutation descpb.MutationID
	var alterColumnTypeMutation descpb.MutationID
	var modifyTTLMutation descpb.MutationID
	var foundAlterPK bool
	var foundAlterColumnType bool
	var foundModifyTTL bool

	for _, m := range desc.Mutations {
		__antithesis_instrumentation__.Notify(271355)

		if foundAlterPK {
			__antithesis_instrumentation__.Notify(271361)
			if alterPKMutation == m.MutationID {
				__antithesis_instrumentation__.Notify(271363)
				vea.Report(unimplemented.NewWithIssue(
					45615,
					"cannot perform other schema changes in the same transaction as a primary key change"),
				)
			} else {
				__antithesis_instrumentation__.Notify(271364)
				vea.Report(unimplemented.NewWithIssue(
					45615,
					"cannot perform a schema change operation while a primary key change is in progress"),
				)
			}
			__antithesis_instrumentation__.Notify(271362)
			return
		} else {
			__antithesis_instrumentation__.Notify(271365)
		}
		__antithesis_instrumentation__.Notify(271356)
		if foundAlterColumnType {
			__antithesis_instrumentation__.Notify(271366)
			if alterColumnTypeMutation == m.MutationID {
				__antithesis_instrumentation__.Notify(271368)
				vea.Report(unimplemented.NewWithIssue(
					47137,
					"cannot perform other schema changes in the same transaction as an ALTER COLUMN TYPE schema change",
				))
			} else {
				__antithesis_instrumentation__.Notify(271369)
				vea.Report(unimplemented.NewWithIssue(
					47137,
					"cannot perform a schema change operation while an ALTER COLUMN TYPE schema change is in progress",
				))
			}
			__antithesis_instrumentation__.Notify(271367)
			return
		} else {
			__antithesis_instrumentation__.Notify(271370)
		}
		__antithesis_instrumentation__.Notify(271357)
		if foundModifyTTL {
			__antithesis_instrumentation__.Notify(271371)
			if modifyTTLMutation == m.MutationID {
				__antithesis_instrumentation__.Notify(271373)
				vea.Report(pgerror.Newf(
					pgcode.FeatureNotSupported,
					"cannot perform other schema changes in the same transaction as a TTL mutation",
				))
			} else {
				__antithesis_instrumentation__.Notify(271374)
				vea.Report(pgerror.Newf(
					pgcode.FeatureNotSupported,
					"cannot perform a schema change operation while a TTL change is in progress",
				))
			}
			__antithesis_instrumentation__.Notify(271372)
			return
		} else {
			__antithesis_instrumentation__.Notify(271375)
		}
		__antithesis_instrumentation__.Notify(271358)
		if m.GetPrimaryKeySwap() != nil {
			__antithesis_instrumentation__.Notify(271376)
			foundAlterPK = true
			alterPKMutation = m.MutationID
		} else {
			__antithesis_instrumentation__.Notify(271377)
		}
		__antithesis_instrumentation__.Notify(271359)
		if m.GetComputedColumnSwap() != nil {
			__antithesis_instrumentation__.Notify(271378)
			foundAlterColumnType = true
			alterColumnTypeMutation = m.MutationID
		} else {
			__antithesis_instrumentation__.Notify(271379)
		}
		__antithesis_instrumentation__.Notify(271360)
		if m.GetModifyRowLevelTTL() != nil {
			__antithesis_instrumentation__.Notify(271380)
			foundModifyTTL = true
			modifyTTLMutation = m.MutationID
		} else {
			__antithesis_instrumentation__.Notify(271381)
		}
	}
	__antithesis_instrumentation__.Notify(271293)

	if dscs := desc.DeclarativeSchemaChangerState; dscs != nil && func() bool {
		__antithesis_instrumentation__.Notify(271382)
		return len(desc.MutationJobs) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(271383)
		vea.Report(errors.AssertionFailedf(
			"invalid concurrent declarative schema change job %d and legacy schema change jobs %v",
			dscs.JobID, desc.MutationJobs))
	} else {
		__antithesis_instrumentation__.Notify(271384)
	}
	__antithesis_instrumentation__.Notify(271294)

	_ = ForEachExprStringInTableDesc(desc, func(expr *string) error {
		__antithesis_instrumentation__.Notify(271385)
		_, err := parser.ParseExpr(*expr)
		vea.Report(err)
		return nil
	})
	__antithesis_instrumentation__.Notify(271295)

	vea.Report(ValidateRowLevelTTL(desc.GetRowLevelTTL()))

	ValidateOnUpdate(desc, vea.Report)
}

func ValidateOnUpdate(desc catalog.TableDescriptor, errReportFn func(err error)) {
	__antithesis_instrumentation__.Notify(271386)
	var onUpdateCols catalog.TableColSet
	for _, col := range desc.AllColumns() {
		__antithesis_instrumentation__.Notify(271388)
		if col.HasOnUpdate() {
			__antithesis_instrumentation__.Notify(271389)
			onUpdateCols.Add(col.GetID())
		} else {
			__antithesis_instrumentation__.Notify(271390)
		}
	}
	__antithesis_instrumentation__.Notify(271387)

	_ = desc.ForeachOutboundFK(func(fk *descpb.ForeignKeyConstraint) error {
		__antithesis_instrumentation__.Notify(271391)
		if fk.OnUpdate == catpb.ForeignKeyAction_NO_ACTION || func() bool {
			__antithesis_instrumentation__.Notify(271394)
			return fk.OnUpdate == catpb.ForeignKeyAction_RESTRICT == true
		}() == true {
			__antithesis_instrumentation__.Notify(271395)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(271396)
		}
		__antithesis_instrumentation__.Notify(271392)
		for _, fkCol := range fk.OriginColumnIDs {
			__antithesis_instrumentation__.Notify(271397)
			if onUpdateCols.Contains(fkCol) {
				__antithesis_instrumentation__.Notify(271398)
				col, err := desc.FindColumnWithID(fkCol)
				if err != nil {
					__antithesis_instrumentation__.Notify(271400)
					return err
				} else {
					__antithesis_instrumentation__.Notify(271401)
				}
				__antithesis_instrumentation__.Notify(271399)
				errReportFn(pgerror.Newf(pgcode.InvalidTableDefinition,
					"cannot specify both ON UPDATE expression and a foreign key"+
						" ON UPDATE action for column %q",
					col.ColName(),
				))
			} else {
				__antithesis_instrumentation__.Notify(271402)
			}
		}
		__antithesis_instrumentation__.Notify(271393)
		return nil
	})
}

func (desc *wrapper) validateConstraintIDs(vea catalog.ValidationErrorAccumulator) {
	__antithesis_instrumentation__.Notify(271403)
	if !vea.IsActive(ConstraintIDsAddedToTableDescsVersion) {
		__antithesis_instrumentation__.Notify(271408)
		return
	} else {
		__antithesis_instrumentation__.Notify(271409)
	}
	__antithesis_instrumentation__.Notify(271404)
	if !desc.IsTable() {
		__antithesis_instrumentation__.Notify(271410)
		return
	} else {
		__antithesis_instrumentation__.Notify(271411)
	}
	__antithesis_instrumentation__.Notify(271405)
	constraints, err := desc.GetConstraintInfo()
	if err != nil {
		__antithesis_instrumentation__.Notify(271412)
		vea.Report(err)
		return
	} else {
		__antithesis_instrumentation__.Notify(271413)
	}
	__antithesis_instrumentation__.Notify(271406)

	orderedNames := make([]string, 0, len(constraints))
	for name := range constraints {
		__antithesis_instrumentation__.Notify(271414)
		orderedNames = append(orderedNames, name)
	}
	__antithesis_instrumentation__.Notify(271407)
	sort.Strings(orderedNames)
	for _, name := range orderedNames {
		__antithesis_instrumentation__.Notify(271415)
		constraint := constraints[name]
		if constraint.ConstraintID == 0 {
			__antithesis_instrumentation__.Notify(271416)
			vea.Report(errors.AssertionFailedf("constraint id was missing for constraint: %s with name %q",
				constraint.Kind,
				name))

		} else {
			__antithesis_instrumentation__.Notify(271417)
		}
	}
}

func (desc *wrapper) validateColumns(
	columnNames map[string]descpb.ColumnID, columnIDs map[descpb.ColumnID]*descpb.ColumnDescriptor,
) error {
	__antithesis_instrumentation__.Notify(271418)
	for _, column := range desc.NonDropColumns() {
		__antithesis_instrumentation__.Notify(271420)

		if err := catalog.ValidateName(column.GetName(), "column"); err != nil {
			__antithesis_instrumentation__.Notify(271432)
			return err
		} else {
			__antithesis_instrumentation__.Notify(271433)
		}
		__antithesis_instrumentation__.Notify(271421)
		if column.GetID() == 0 {
			__antithesis_instrumentation__.Notify(271434)
			return errors.AssertionFailedf("invalid column ID %d", errors.Safe(column.GetID()))
		} else {
			__antithesis_instrumentation__.Notify(271435)
		}
		__antithesis_instrumentation__.Notify(271422)

		if _, columnNameExists := columnNames[column.GetName()]; columnNameExists {
			__antithesis_instrumentation__.Notify(271436)
			for i := range desc.Columns {
				__antithesis_instrumentation__.Notify(271438)
				if desc.Columns[i].Name == column.GetName() {
					__antithesis_instrumentation__.Notify(271439)
					return pgerror.Newf(pgcode.DuplicateColumn,
						"duplicate column name: %q", column.GetName())
				} else {
					__antithesis_instrumentation__.Notify(271440)
				}
			}
			__antithesis_instrumentation__.Notify(271437)
			return pgerror.Newf(pgcode.DuplicateColumn,
				"duplicate: column %q in the middle of being added, not yet public", column.GetName())
		} else {
			__antithesis_instrumentation__.Notify(271441)
		}
		__antithesis_instrumentation__.Notify(271423)
		if colinfo.IsSystemColumnName(column.GetName()) {
			__antithesis_instrumentation__.Notify(271442)
			return pgerror.Newf(pgcode.DuplicateColumn,
				"column name %q conflicts with a system column name", column.GetName())
		} else {
			__antithesis_instrumentation__.Notify(271443)
		}
		__antithesis_instrumentation__.Notify(271424)
		columnNames[column.GetName()] = column.GetID()

		if other, ok := columnIDs[column.GetID()]; ok {
			__antithesis_instrumentation__.Notify(271444)
			return errors.Newf("column %q duplicate ID of column %q: %d",
				column.GetName(), other.Name, column.GetID())
		} else {
			__antithesis_instrumentation__.Notify(271445)
		}
		__antithesis_instrumentation__.Notify(271425)
		columnIDs[column.GetID()] = column.ColumnDesc()

		if column.GetID() >= desc.NextColumnID {
			__antithesis_instrumentation__.Notify(271446)
			return errors.AssertionFailedf("column %q invalid ID (%d) >= next column ID (%d)",
				column.GetName(), errors.Safe(column.GetID()), errors.Safe(desc.NextColumnID))
		} else {
			__antithesis_instrumentation__.Notify(271447)
		}
		__antithesis_instrumentation__.Notify(271426)

		if column.IsComputed() {
			__antithesis_instrumentation__.Notify(271448)

			expr, err := parser.ParseExpr(column.GetComputeExpr())
			if err != nil {
				__antithesis_instrumentation__.Notify(271451)
				return err
			} else {
				__antithesis_instrumentation__.Notify(271452)
			}
			__antithesis_instrumentation__.Notify(271449)
			valid, err := schemaexpr.HasValidColumnReferences(desc, expr)
			if err != nil {
				__antithesis_instrumentation__.Notify(271453)
				return err
			} else {
				__antithesis_instrumentation__.Notify(271454)
			}
			__antithesis_instrumentation__.Notify(271450)
			if !valid {
				__antithesis_instrumentation__.Notify(271455)
				return errors.Newf("computed column %q refers to unknown columns in expression: %s",
					column.GetName(), column.GetComputeExpr())
			} else {
				__antithesis_instrumentation__.Notify(271456)
			}
		} else {
			__antithesis_instrumentation__.Notify(271457)
			if column.IsVirtual() {
				__antithesis_instrumentation__.Notify(271458)
				return errors.Newf("virtual column %q is not computed", column.GetName())
			} else {
				__antithesis_instrumentation__.Notify(271459)
			}
		}
		__antithesis_instrumentation__.Notify(271427)

		if column.IsComputed() {
			__antithesis_instrumentation__.Notify(271460)
			if column.HasDefault() {
				__antithesis_instrumentation__.Notify(271462)
				return pgerror.Newf(pgcode.InvalidTableDefinition,
					"computed column %q cannot also have a DEFAULT expression",
					column.GetName(),
				)
			} else {
				__antithesis_instrumentation__.Notify(271463)
			}
			__antithesis_instrumentation__.Notify(271461)
			if column.HasOnUpdate() {
				__antithesis_instrumentation__.Notify(271464)
				return pgerror.Newf(pgcode.InvalidTableDefinition,
					"computed column %q cannot also have an ON UPDATE expression",
					column.GetName(),
				)
			} else {
				__antithesis_instrumentation__.Notify(271465)
			}
		} else {
			__antithesis_instrumentation__.Notify(271466)
		}
		__antithesis_instrumentation__.Notify(271428)

		if column.IsHidden() && func() bool {
			__antithesis_instrumentation__.Notify(271467)
			return column.IsInaccessible() == true
		}() == true {
			__antithesis_instrumentation__.Notify(271468)
			return errors.Newf("column %q cannot be hidden and inaccessible", column.GetName())
		} else {
			__antithesis_instrumentation__.Notify(271469)
		}
		__antithesis_instrumentation__.Notify(271429)

		if column.IsComputed() && func() bool {
			__antithesis_instrumentation__.Notify(271470)
			return column.IsGeneratedAsIdentity() == true
		}() == true {
			__antithesis_instrumentation__.Notify(271471)
			return errors.Newf("both generated identity and computed expression specified for column %q", column.GetName())
		} else {
			__antithesis_instrumentation__.Notify(271472)
		}
		__antithesis_instrumentation__.Notify(271430)

		if column.IsNullable() && func() bool {
			__antithesis_instrumentation__.Notify(271473)
			return column.IsGeneratedAsIdentity() == true
		}() == true {
			__antithesis_instrumentation__.Notify(271474)
			return errors.Newf("conflicting NULL/NOT NULL declarations for column %q", column.GetName())
		} else {
			__antithesis_instrumentation__.Notify(271475)
		}
		__antithesis_instrumentation__.Notify(271431)

		if column.HasOnUpdate() && func() bool {
			__antithesis_instrumentation__.Notify(271476)
			return column.IsGeneratedAsIdentity() == true
		}() == true {
			__antithesis_instrumentation__.Notify(271477)
			return errors.Newf("both generated identity and on update expression specified for column %q", column.GetName())
		} else {
			__antithesis_instrumentation__.Notify(271478)
		}
	}
	__antithesis_instrumentation__.Notify(271419)
	return nil
}

func (desc *wrapper) validateColumnFamilies(
	columnIDs map[descpb.ColumnID]*descpb.ColumnDescriptor,
) error {
	__antithesis_instrumentation__.Notify(271479)
	if len(desc.Families) < 1 {
		__antithesis_instrumentation__.Notify(271484)
		return errors.Newf("at least 1 column family must be specified")
	} else {
		__antithesis_instrumentation__.Notify(271485)
	}
	__antithesis_instrumentation__.Notify(271480)
	if desc.Families[0].ID != descpb.FamilyID(0) {
		__antithesis_instrumentation__.Notify(271486)
		return errors.Newf("the 0th family must have ID 0")
	} else {
		__antithesis_instrumentation__.Notify(271487)
	}
	__antithesis_instrumentation__.Notify(271481)

	familyNames := map[string]struct{}{}
	familyIDs := map[descpb.FamilyID]string{}
	colIDToFamilyID := map[descpb.ColumnID]descpb.FamilyID{}
	for i := range desc.Families {
		__antithesis_instrumentation__.Notify(271488)
		family := &desc.Families[i]
		if err := catalog.ValidateName(family.Name, "family"); err != nil {
			__antithesis_instrumentation__.Notify(271496)
			return err
		} else {
			__antithesis_instrumentation__.Notify(271497)
		}
		__antithesis_instrumentation__.Notify(271489)

		if i != 0 {
			__antithesis_instrumentation__.Notify(271498)
			prevFam := desc.Families[i-1]
			if family.ID < prevFam.ID {
				__antithesis_instrumentation__.Notify(271499)
				return errors.Newf(
					"family %s at index %d has id %d less than family %s at index %d with id %d",
					family.Name, i, family.ID, prevFam.Name, i-1, prevFam.ID)
			} else {
				__antithesis_instrumentation__.Notify(271500)
			}
		} else {
			__antithesis_instrumentation__.Notify(271501)
		}
		__antithesis_instrumentation__.Notify(271490)

		if _, ok := familyNames[family.Name]; ok {
			__antithesis_instrumentation__.Notify(271502)
			return errors.Newf("duplicate family name: %q", family.Name)
		} else {
			__antithesis_instrumentation__.Notify(271503)
		}
		__antithesis_instrumentation__.Notify(271491)
		familyNames[family.Name] = struct{}{}

		if other, ok := familyIDs[family.ID]; ok {
			__antithesis_instrumentation__.Notify(271504)
			return errors.Newf("family %q duplicate ID of family %q: %d",
				family.Name, other, family.ID)
		} else {
			__antithesis_instrumentation__.Notify(271505)
		}
		__antithesis_instrumentation__.Notify(271492)
		familyIDs[family.ID] = family.Name

		if family.ID >= desc.NextFamilyID {
			__antithesis_instrumentation__.Notify(271506)
			return errors.Newf("family %q invalid family ID (%d) > next family ID (%d)",
				family.Name, family.ID, desc.NextFamilyID)
		} else {
			__antithesis_instrumentation__.Notify(271507)
		}
		__antithesis_instrumentation__.Notify(271493)

		if len(family.ColumnIDs) != len(family.ColumnNames) {
			__antithesis_instrumentation__.Notify(271508)
			return errors.Newf("mismatched column ID size (%d) and name size (%d)",
				len(family.ColumnIDs), len(family.ColumnNames))
		} else {
			__antithesis_instrumentation__.Notify(271509)
		}
		__antithesis_instrumentation__.Notify(271494)

		for i, colID := range family.ColumnIDs {
			__antithesis_instrumentation__.Notify(271510)
			col, ok := columnIDs[colID]
			if !ok {
				__antithesis_instrumentation__.Notify(271513)
				return errors.Newf("family %q contains unknown column \"%d\"", family.Name, colID)
			} else {
				__antithesis_instrumentation__.Notify(271514)
			}
			__antithesis_instrumentation__.Notify(271511)
			if col.Name != family.ColumnNames[i] {
				__antithesis_instrumentation__.Notify(271515)
				return errors.Newf("family %q column %d should have name %q, but found name %q",
					family.Name, colID, col.Name, family.ColumnNames[i])
			} else {
				__antithesis_instrumentation__.Notify(271516)
			}
			__antithesis_instrumentation__.Notify(271512)
			if col.Virtual {
				__antithesis_instrumentation__.Notify(271517)
				return errors.Newf("virtual computed column %q cannot be part of a family", col.Name)
			} else {
				__antithesis_instrumentation__.Notify(271518)
			}
		}
		__antithesis_instrumentation__.Notify(271495)

		for _, colID := range family.ColumnIDs {
			__antithesis_instrumentation__.Notify(271519)
			if famID, ok := colIDToFamilyID[colID]; ok {
				__antithesis_instrumentation__.Notify(271521)
				return errors.Newf("column %d is in both family %d and %d", colID, famID, family.ID)
			} else {
				__antithesis_instrumentation__.Notify(271522)
			}
			__antithesis_instrumentation__.Notify(271520)
			colIDToFamilyID[colID] = family.ID
		}
	}
	__antithesis_instrumentation__.Notify(271482)
	for colID, colDesc := range columnIDs {
		__antithesis_instrumentation__.Notify(271523)
		if !colDesc.Virtual {
			__antithesis_instrumentation__.Notify(271524)
			if _, ok := colIDToFamilyID[colID]; !ok {
				__antithesis_instrumentation__.Notify(271525)
				return errors.Newf("column %q is not in any column family", colDesc.Name)
			} else {
				__antithesis_instrumentation__.Notify(271526)
			}
		} else {
			__antithesis_instrumentation__.Notify(271527)
		}
	}
	__antithesis_instrumentation__.Notify(271483)
	return nil
}

func (desc *wrapper) validateCheckConstraints(
	columnIDs map[descpb.ColumnID]*descpb.ColumnDescriptor,
) error {
	__antithesis_instrumentation__.Notify(271528)
	for _, chk := range desc.AllActiveAndInactiveChecks() {
		__antithesis_instrumentation__.Notify(271530)

		for _, colID := range chk.ColumnIDs {
			__antithesis_instrumentation__.Notify(271534)
			_, ok := columnIDs[colID]
			if !ok {
				__antithesis_instrumentation__.Notify(271535)
				return errors.Newf("check constraint %q contains unknown column \"%d\"", chk.Name, colID)
			} else {
				__antithesis_instrumentation__.Notify(271536)
			}
		}
		__antithesis_instrumentation__.Notify(271531)

		expr, err := parser.ParseExpr(chk.Expr)
		if err != nil {
			__antithesis_instrumentation__.Notify(271537)
			return err
		} else {
			__antithesis_instrumentation__.Notify(271538)
		}
		__antithesis_instrumentation__.Notify(271532)
		valid, err := schemaexpr.HasValidColumnReferences(desc, expr)
		if err != nil {
			__antithesis_instrumentation__.Notify(271539)
			return err
		} else {
			__antithesis_instrumentation__.Notify(271540)
		}
		__antithesis_instrumentation__.Notify(271533)
		if !valid {
			__antithesis_instrumentation__.Notify(271541)
			return errors.Newf("check constraint %q refers to unknown columns in expression: %s",
				chk.Name, chk.Expr)
		} else {
			__antithesis_instrumentation__.Notify(271542)
		}
	}
	__antithesis_instrumentation__.Notify(271529)
	return nil
}

func (desc *wrapper) validateUniqueWithoutIndexConstraints(
	columnIDs map[descpb.ColumnID]*descpb.ColumnDescriptor,
) error {
	__antithesis_instrumentation__.Notify(271543)
	for _, c := range desc.AllActiveAndInactiveUniqueWithoutIndexConstraints() {
		__antithesis_instrumentation__.Notify(271545)
		if err := catalog.ValidateName(c.Name, "unique without index constraint"); err != nil {
			__antithesis_instrumentation__.Notify(271549)
			return err
		} else {
			__antithesis_instrumentation__.Notify(271550)
		}
		__antithesis_instrumentation__.Notify(271546)

		if c.TableID != desc.ID {
			__antithesis_instrumentation__.Notify(271551)
			return errors.Newf(
				"TableID mismatch for unique without index constraint %q: \"%d\" doesn't match descriptor: \"%d\"",
				c.Name, c.TableID, desc.ID,
			)
		} else {
			__antithesis_instrumentation__.Notify(271552)
		}
		__antithesis_instrumentation__.Notify(271547)

		var seen util.FastIntSet
		for _, colID := range c.ColumnIDs {
			__antithesis_instrumentation__.Notify(271553)
			_, ok := columnIDs[colID]
			if !ok {
				__antithesis_instrumentation__.Notify(271556)
				return errors.Newf(
					"unique without index constraint %q contains unknown column \"%d\"", c.Name, colID,
				)
			} else {
				__antithesis_instrumentation__.Notify(271557)
			}
			__antithesis_instrumentation__.Notify(271554)
			if seen.Contains(int(colID)) {
				__antithesis_instrumentation__.Notify(271558)
				return errors.Newf(
					"unique without index constraint %q contains duplicate column \"%d\"", c.Name, colID,
				)
			} else {
				__antithesis_instrumentation__.Notify(271559)
			}
			__antithesis_instrumentation__.Notify(271555)
			seen.Add(int(colID))
		}
		__antithesis_instrumentation__.Notify(271548)

		if c.IsPartial() {
			__antithesis_instrumentation__.Notify(271560)
			expr, err := parser.ParseExpr(c.Predicate)
			if err != nil {
				__antithesis_instrumentation__.Notify(271563)
				return err
			} else {
				__antithesis_instrumentation__.Notify(271564)
			}
			__antithesis_instrumentation__.Notify(271561)
			valid, err := schemaexpr.HasValidColumnReferences(desc, expr)
			if err != nil {
				__antithesis_instrumentation__.Notify(271565)
				return err
			} else {
				__antithesis_instrumentation__.Notify(271566)
			}
			__antithesis_instrumentation__.Notify(271562)
			if !valid {
				__antithesis_instrumentation__.Notify(271567)
				return errors.Newf(
					"partial unique without index constraint %q refers to unknown columns in predicate: %s",
					c.Name,
					c.Predicate,
				)
			} else {
				__antithesis_instrumentation__.Notify(271568)
			}
		} else {
			__antithesis_instrumentation__.Notify(271569)
		}
	}
	__antithesis_instrumentation__.Notify(271544)

	return nil
}

func (desc *wrapper) validateTableIndexes(
	columnNames map[string]descpb.ColumnID, vea catalog.ValidationErrorAccumulator,
) error {
	__antithesis_instrumentation__.Notify(271570)
	if len(desc.PrimaryIndex.KeyColumnIDs) == 0 {
		__antithesis_instrumentation__.Notify(271575)
		return ErrMissingPrimaryKey
	} else {
		__antithesis_instrumentation__.Notify(271576)
	}
	__antithesis_instrumentation__.Notify(271571)

	columnsByID := make(map[descpb.ColumnID]catalog.Column)
	for _, col := range desc.DeletableColumns() {
		__antithesis_instrumentation__.Notify(271577)
		columnsByID[col.GetID()] = col
	}
	__antithesis_instrumentation__.Notify(271572)

	if !vea.IsActive(clusterversion.Start22_1) {
		__antithesis_instrumentation__.Notify(271578)

		for _, pkID := range desc.PrimaryIndex.KeyColumnIDs {
			__antithesis_instrumentation__.Notify(271579)
			if col := columnsByID[pkID]; col != nil && func() bool {
				__antithesis_instrumentation__.Notify(271580)
				return col.IsVirtual() == true
			}() == true {
				__antithesis_instrumentation__.Notify(271581)
				return errors.Newf("primary index column %q cannot be virtual", col.GetName())
			} else {
				__antithesis_instrumentation__.Notify(271582)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(271583)
	}
	__antithesis_instrumentation__.Notify(271573)

	indexNames := map[string]struct{}{}
	indexIDs := map[descpb.IndexID]string{}
	for _, idx := range desc.NonDropIndexes() {
		__antithesis_instrumentation__.Notify(271584)
		if err := catalog.ValidateName(idx.GetName(), "index"); err != nil {
			__antithesis_instrumentation__.Notify(271608)
			return err
		} else {
			__antithesis_instrumentation__.Notify(271609)
		}
		__antithesis_instrumentation__.Notify(271585)
		if idx.GetID() == 0 {
			__antithesis_instrumentation__.Notify(271610)
			return errors.Newf("invalid index ID %d", idx.GetID())
		} else {
			__antithesis_instrumentation__.Notify(271611)
		}
		__antithesis_instrumentation__.Notify(271586)

		if idx.IndexDesc().ForeignKey.IsSet() || func() bool {
			__antithesis_instrumentation__.Notify(271612)
			return len(idx.IndexDesc().ReferencedBy) > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(271613)
			return errors.AssertionFailedf("index %q contains deprecated foreign key representation", idx.GetName())
		} else {
			__antithesis_instrumentation__.Notify(271614)
		}
		__antithesis_instrumentation__.Notify(271587)

		if len(idx.IndexDesc().Interleave.Ancestors) > 0 || func() bool {
			__antithesis_instrumentation__.Notify(271615)
			return len(idx.IndexDesc().InterleavedBy) > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(271616)
			return errors.Newf("index is interleaved")
		} else {
			__antithesis_instrumentation__.Notify(271617)
		}
		__antithesis_instrumentation__.Notify(271588)

		if _, indexNameExists := indexNames[idx.GetName()]; indexNameExists {
			__antithesis_instrumentation__.Notify(271618)
			for i := range desc.Indexes {
				__antithesis_instrumentation__.Notify(271620)
				if desc.Indexes[i].Name == idx.GetName() {
					__antithesis_instrumentation__.Notify(271621)

					return errors.HandleAsAssertionFailure(errors.Newf("duplicate index name: %q", idx.GetName()))
				} else {
					__antithesis_instrumentation__.Notify(271622)
				}
			}
			__antithesis_instrumentation__.Notify(271619)

			return errors.HandleAsAssertionFailure(errors.Newf(
				"duplicate: index %q in the middle of being added, not yet public", idx.GetName()))
		} else {
			__antithesis_instrumentation__.Notify(271623)
		}
		__antithesis_instrumentation__.Notify(271589)
		indexNames[idx.GetName()] = struct{}{}

		if other, ok := indexIDs[idx.GetID()]; ok {
			__antithesis_instrumentation__.Notify(271624)
			return errors.Newf("index %q duplicate ID of index %q: %d",
				idx.GetName(), other, idx.GetID())
		} else {
			__antithesis_instrumentation__.Notify(271625)
		}
		__antithesis_instrumentation__.Notify(271590)
		indexIDs[idx.GetID()] = idx.GetName()

		if idx.GetID() >= desc.NextIndexID {
			__antithesis_instrumentation__.Notify(271626)
			return errors.Newf("index %q invalid index ID (%d) > next index ID (%d)",
				idx.GetName(), idx.GetID(), desc.NextIndexID)
		} else {
			__antithesis_instrumentation__.Notify(271627)
		}
		__antithesis_instrumentation__.Notify(271591)

		if len(idx.IndexDesc().KeyColumnIDs) != len(idx.IndexDesc().KeyColumnNames) {
			__antithesis_instrumentation__.Notify(271628)
			return errors.Newf("mismatched column IDs (%d) and names (%d)",
				len(idx.IndexDesc().KeyColumnIDs), len(idx.IndexDesc().KeyColumnNames))
		} else {
			__antithesis_instrumentation__.Notify(271629)
		}
		__antithesis_instrumentation__.Notify(271592)
		if len(idx.IndexDesc().KeyColumnIDs) != len(idx.IndexDesc().KeyColumnDirections) {
			__antithesis_instrumentation__.Notify(271630)
			return errors.Newf("mismatched column IDs (%d) and directions (%d)",
				len(idx.IndexDesc().KeyColumnIDs), len(idx.IndexDesc().KeyColumnDirections))
		} else {
			__antithesis_instrumentation__.Notify(271631)
		}
		__antithesis_instrumentation__.Notify(271593)

		if len(idx.IndexDesc().StoreColumnIDs) > len(idx.IndexDesc().StoreColumnNames) {
			__antithesis_instrumentation__.Notify(271632)
			return errors.Newf("mismatched STORING column IDs (%d) and names (%d)",
				len(idx.IndexDesc().StoreColumnIDs), len(idx.IndexDesc().StoreColumnNames))
		} else {
			__antithesis_instrumentation__.Notify(271633)
		}
		__antithesis_instrumentation__.Notify(271594)

		if len(idx.IndexDesc().KeyColumnIDs) == 0 {
			__antithesis_instrumentation__.Notify(271634)
			return errors.Newf("index %q must contain at least 1 column", idx.GetName())
		} else {
			__antithesis_instrumentation__.Notify(271635)
		}
		__antithesis_instrumentation__.Notify(271595)

		var validateIndexDup catalog.TableColSet
		for i, name := range idx.IndexDesc().KeyColumnNames {
			__antithesis_instrumentation__.Notify(271636)
			inIndexColID := idx.IndexDesc().KeyColumnIDs[i]
			colID, ok := columnNames[name]
			if !ok {
				__antithesis_instrumentation__.Notify(271640)
				return errors.Newf("index %q contains unknown column %q", idx.GetName(), name)
			} else {
				__antithesis_instrumentation__.Notify(271641)
			}
			__antithesis_instrumentation__.Notify(271637)
			if colID != inIndexColID {
				__antithesis_instrumentation__.Notify(271642)
				return errors.Newf("index %q column %q should have ID %d, but found ID %d",
					idx.GetName(), name, colID, inIndexColID)
			} else {
				__antithesis_instrumentation__.Notify(271643)
			}
			__antithesis_instrumentation__.Notify(271638)
			if validateIndexDup.Contains(colID) {
				__antithesis_instrumentation__.Notify(271644)
				col, _ := desc.FindColumnWithID(colID)
				if col.IsExpressionIndexColumn() {
					__antithesis_instrumentation__.Notify(271646)
					return pgerror.Newf(pgcode.FeatureNotSupported,
						"index %q contains duplicate expression %q",
						idx.GetName(), col.GetComputeExpr(),
					)
				} else {
					__antithesis_instrumentation__.Notify(271647)
				}
				__antithesis_instrumentation__.Notify(271645)
				return pgerror.Newf(pgcode.FeatureNotSupported,
					"index %q contains duplicate column %q",
					idx.GetName(), name,
				)
			} else {
				__antithesis_instrumentation__.Notify(271648)
			}
			__antithesis_instrumentation__.Notify(271639)
			validateIndexDup.Add(colID)
		}
		__antithesis_instrumentation__.Notify(271596)
		for i, colID := range idx.IndexDesc().StoreColumnIDs {
			__antithesis_instrumentation__.Notify(271649)
			inIndexColName := idx.IndexDesc().StoreColumnNames[i]
			col, exists := columnsByID[colID]
			if !exists {
				__antithesis_instrumentation__.Notify(271651)
				return errors.Newf("index %q contains stored column %q with unknown ID %d", idx.GetName(), inIndexColName, colID)
			} else {
				__antithesis_instrumentation__.Notify(271652)
			}
			__antithesis_instrumentation__.Notify(271650)
			if col.GetName() != inIndexColName {
				__antithesis_instrumentation__.Notify(271653)
				return errors.Newf("index %q stored column ID %d should have name %q, but found name %q",
					idx.GetName(), colID, col.ColName(), inIndexColName)
			} else {
				__antithesis_instrumentation__.Notify(271654)
			}
		}
		__antithesis_instrumentation__.Notify(271597)
		for _, colID := range idx.IndexDesc().KeySuffixColumnIDs {
			__antithesis_instrumentation__.Notify(271655)
			if _, exists := columnsByID[colID]; !exists {
				__antithesis_instrumentation__.Notify(271656)
				return errors.Newf("index %q key suffix column ID %d is invalid",
					idx.GetName(), colID)
			} else {
				__antithesis_instrumentation__.Notify(271657)
			}
		}
		__antithesis_instrumentation__.Notify(271598)

		if idx.IsSharded() {
			__antithesis_instrumentation__.Notify(271658)
			if err := desc.ensureShardedIndexNotComputed(idx.IndexDesc()); err != nil {
				__antithesis_instrumentation__.Notify(271660)
				return err
			} else {
				__antithesis_instrumentation__.Notify(271661)
			}
			__antithesis_instrumentation__.Notify(271659)
			if _, exists := columnNames[idx.GetSharded().Name]; !exists {
				__antithesis_instrumentation__.Notify(271662)
				return errors.Newf("index %q refers to non-existent shard column %q",
					idx.GetName(), idx.GetSharded().Name)
			} else {
				__antithesis_instrumentation__.Notify(271663)
			}
		} else {
			__antithesis_instrumentation__.Notify(271664)
		}
		__antithesis_instrumentation__.Notify(271599)
		if idx.IsPartial() {
			__antithesis_instrumentation__.Notify(271665)
			expr, err := parser.ParseExpr(idx.GetPredicate())
			if err != nil {
				__antithesis_instrumentation__.Notify(271668)
				return err
			} else {
				__antithesis_instrumentation__.Notify(271669)
			}
			__antithesis_instrumentation__.Notify(271666)
			valid, err := schemaexpr.HasValidColumnReferences(desc, expr)
			if err != nil {
				__antithesis_instrumentation__.Notify(271670)
				return err
			} else {
				__antithesis_instrumentation__.Notify(271671)
			}
			__antithesis_instrumentation__.Notify(271667)
			if !valid {
				__antithesis_instrumentation__.Notify(271672)
				return errors.Newf("partial index %q refers to unknown columns in predicate: %s",
					idx.GetName(), idx.GetPredicate())
			} else {
				__antithesis_instrumentation__.Notify(271673)
			}
		} else {
			__antithesis_instrumentation__.Notify(271674)
		}
		__antithesis_instrumentation__.Notify(271600)

		if !idx.IsMutation() {
			__antithesis_instrumentation__.Notify(271675)
			if idx.IndexDesc().UseDeletePreservingEncoding {
				__antithesis_instrumentation__.Notify(271676)
				return errors.Newf("public index %q is using the delete preserving encoding", idx.GetName())
			} else {
				__antithesis_instrumentation__.Notify(271677)
			}
		} else {
			__antithesis_instrumentation__.Notify(271678)
		}
		__antithesis_instrumentation__.Notify(271601)

		curPKColIDs := catalog.MakeTableColSet(desc.PrimaryIndex.KeyColumnIDs...)
		newPKColIDs := catalog.MakeTableColSet()
		for _, mut := range desc.Mutations {
			__antithesis_instrumentation__.Notify(271679)
			if mut.GetPrimaryKeySwap() != nil {
				__antithesis_instrumentation__.Notify(271680)
				newPKIdxID := mut.GetPrimaryKeySwap().NewPrimaryIndexId
				newPK, err := desc.FindIndexWithID(newPKIdxID)
				if err != nil {
					__antithesis_instrumentation__.Notify(271682)
					return err
				} else {
					__antithesis_instrumentation__.Notify(271683)
				}
				__antithesis_instrumentation__.Notify(271681)
				newPKColIDs.UnionWith(newPK.CollectKeyColumnIDs())
			} else {
				__antithesis_instrumentation__.Notify(271684)
			}
		}
		__antithesis_instrumentation__.Notify(271602)
		for _, colID := range idx.IndexDesc().KeySuffixColumnIDs {
			__antithesis_instrumentation__.Notify(271685)
			if !vea.IsActive(clusterversion.Start22_1) {
				__antithesis_instrumentation__.Notify(271691)
				if col := columnsByID[colID]; col != nil && func() bool {
					__antithesis_instrumentation__.Notify(271692)
					return col.IsVirtual() == true
				}() == true {
					__antithesis_instrumentation__.Notify(271693)
					return errors.Newf("index %q cannot store virtual column %d", idx.GetName(), colID)
				} else {
					__antithesis_instrumentation__.Notify(271694)
				}
			} else {
				__antithesis_instrumentation__.Notify(271695)
			}
			__antithesis_instrumentation__.Notify(271686)

			if _, ok := columnsByID[colID]; !ok {
				__antithesis_instrumentation__.Notify(271696)
				return errors.Newf("column %d does not exist in table %s", colID, desc.Name)
			} else {
				__antithesis_instrumentation__.Notify(271697)
			}
			__antithesis_instrumentation__.Notify(271687)
			col := columnsByID[colID]
			if !col.IsVirtual() {
				__antithesis_instrumentation__.Notify(271698)
				continue
			} else {
				__antithesis_instrumentation__.Notify(271699)
			}
			__antithesis_instrumentation__.Notify(271688)

			if newPKColIDs.Len() == 0 && func() bool {
				__antithesis_instrumentation__.Notify(271700)
				return curPKColIDs.Contains(colID) == true
			}() == true {
				__antithesis_instrumentation__.Notify(271701)
				continue
			} else {
				__antithesis_instrumentation__.Notify(271702)
			}
			__antithesis_instrumentation__.Notify(271689)

			isOldPKCol := !idx.Adding() && func() bool {
				__antithesis_instrumentation__.Notify(271703)
				return curPKColIDs.Contains(colID) == true
			}() == true
			isNewPKCol := idx.Adding() && func() bool {
				__antithesis_instrumentation__.Notify(271704)
				return newPKColIDs.Contains(colID) == true
			}() == true
			if newPKColIDs.Len() > 0 && func() bool {
				__antithesis_instrumentation__.Notify(271705)
				return (isOldPKCol || func() bool {
					__antithesis_instrumentation__.Notify(271706)
					return isNewPKCol == true
				}() == true) == true
			}() == true {
				__antithesis_instrumentation__.Notify(271707)
				continue
			} else {
				__antithesis_instrumentation__.Notify(271708)
			}
			__antithesis_instrumentation__.Notify(271690)

			return errors.Newf("index %q cannot store virtual column %q", idx.GetName(), col.GetName())
		}
		__antithesis_instrumentation__.Notify(271603)

		for i, colID := range idx.IndexDesc().StoreColumnIDs {
			__antithesis_instrumentation__.Notify(271709)
			if col := columnsByID[colID]; col != nil && func() bool {
				__antithesis_instrumentation__.Notify(271710)
				return col.IsVirtual() == true
			}() == true {
				__antithesis_instrumentation__.Notify(271711)
				return errors.Newf("index %q cannot store virtual column %q",
					idx.GetName(), idx.IndexDesc().StoreColumnNames[i])
			} else {
				__antithesis_instrumentation__.Notify(271712)
			}
		}
		__antithesis_instrumentation__.Notify(271604)
		if idx.Primary() {
			__antithesis_instrumentation__.Notify(271713)
			if idx.GetVersion() < descpb.PrimaryIndexWithStoredColumnsVersion {
				__antithesis_instrumentation__.Notify(271715)
				return errors.AssertionFailedf("primary index %q has invalid version %d, expected at least %d",
					idx.GetName(), idx.GetVersion(), descpb.PrimaryIndexWithStoredColumnsVersion)
			} else {
				__antithesis_instrumentation__.Notify(271716)
			}
			__antithesis_instrumentation__.Notify(271714)
			if idx.IndexDesc().EncodingType != descpb.PrimaryIndexEncoding {
				__antithesis_instrumentation__.Notify(271717)
				return errors.AssertionFailedf("primary index %q has invalid encoding type %d in proto, expected %d",
					idx.GetName(), idx.IndexDesc().EncodingType, descpb.PrimaryIndexEncoding)
			} else {
				__antithesis_instrumentation__.Notify(271718)
			}
		} else {
			__antithesis_instrumentation__.Notify(271719)
		}
		__antithesis_instrumentation__.Notify(271605)

		if idx.GetVersion() < descpb.StrictIndexColumnIDGuaranteesVersion {
			__antithesis_instrumentation__.Notify(271720)
			continue
		} else {
			__antithesis_instrumentation__.Notify(271721)
		}
		__antithesis_instrumentation__.Notify(271606)
		slices := []struct {
			name  string
			slice []descpb.ColumnID
		}{
			{"KeyColumnIDs", idx.IndexDesc().KeyColumnIDs},
			{"KeySuffixColumnIDs", idx.IndexDesc().KeySuffixColumnIDs},
			{"StoreColumnIDs", idx.IndexDesc().StoreColumnIDs},
		}
		allIDs := catalog.MakeTableColSet()
		sets := map[string]catalog.TableColSet{}
		for _, s := range slices {
			__antithesis_instrumentation__.Notify(271722)
			set := catalog.MakeTableColSet(s.slice...)
			sets[s.name] = set
			if set.Len() == 0 {
				__antithesis_instrumentation__.Notify(271726)
				continue
			} else {
				__antithesis_instrumentation__.Notify(271727)
			}
			__antithesis_instrumentation__.Notify(271723)
			if set.Ordered()[0] <= 0 {
				__antithesis_instrumentation__.Notify(271728)
				return errors.AssertionFailedf("index %q contains invalid column ID value %d in %s",
					idx.GetName(), set.Ordered()[0], s.name)
			} else {
				__antithesis_instrumentation__.Notify(271729)
			}
			__antithesis_instrumentation__.Notify(271724)
			if set.Len() < len(s.slice) {
				__antithesis_instrumentation__.Notify(271730)
				return errors.AssertionFailedf("index %q has duplicates in %s: %v",
					idx.GetName(), s.name, s.slice)
			} else {
				__antithesis_instrumentation__.Notify(271731)
			}
			__antithesis_instrumentation__.Notify(271725)
			allIDs.UnionWith(set)
		}
		__antithesis_instrumentation__.Notify(271607)
		foundIn := make([]string, 0, len(sets))
		for _, colID := range allIDs.Ordered() {
			__antithesis_instrumentation__.Notify(271732)
			foundIn = foundIn[:0]
			for _, s := range slices {
				__antithesis_instrumentation__.Notify(271734)
				set := sets[s.name]
				if set.Contains(colID) {
					__antithesis_instrumentation__.Notify(271735)
					foundIn = append(foundIn, s.name)
				} else {
					__antithesis_instrumentation__.Notify(271736)
				}
			}
			__antithesis_instrumentation__.Notify(271733)
			if len(foundIn) > 1 {
				__antithesis_instrumentation__.Notify(271737)
				return errors.AssertionFailedf("index %q has column ID %d present in: %v",
					idx.GetName(), colID, foundIn)
			} else {
				__antithesis_instrumentation__.Notify(271738)
			}
		}
	}
	__antithesis_instrumentation__.Notify(271574)
	return nil
}

func (desc *wrapper) ensureShardedIndexNotComputed(index *descpb.IndexDescriptor) error {
	__antithesis_instrumentation__.Notify(271739)
	for _, colName := range index.Sharded.ColumnNames {
		__antithesis_instrumentation__.Notify(271741)
		col, err := desc.FindColumnWithName(tree.Name(colName))
		if err != nil {
			__antithesis_instrumentation__.Notify(271743)
			return err
		} else {
			__antithesis_instrumentation__.Notify(271744)
		}
		__antithesis_instrumentation__.Notify(271742)
		if col.IsComputed() {
			__antithesis_instrumentation__.Notify(271745)
			return pgerror.Newf(pgcode.InvalidTableDefinition,
				"cannot create a sharded index on a computed column")
		} else {
			__antithesis_instrumentation__.Notify(271746)
		}
	}
	__antithesis_instrumentation__.Notify(271740)
	return nil
}

func (desc *wrapper) validatePartitioningDescriptor(
	a *tree.DatumAlloc,
	idx catalog.Index,
	part catalog.Partitioning,
	colOffset int,
	partitionNames map[string]string,
) error {
	__antithesis_instrumentation__.Notify(271747)
	if part.NumImplicitColumns() > part.NumColumns() {
		__antithesis_instrumentation__.Notify(271756)
		return errors.Newf(
			"cannot have implicit partitioning columns (%d) > partitioning columns (%d)",
			part.NumImplicitColumns(),
			part.NumColumns(),
		)
	} else {
		__antithesis_instrumentation__.Notify(271757)
	}
	__antithesis_instrumentation__.Notify(271748)
	if part.NumColumns() == 0 {
		__antithesis_instrumentation__.Notify(271758)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(271759)
	}
	__antithesis_instrumentation__.Notify(271749)

	fakePrefixDatums := make([]tree.Datum, colOffset)
	for i := range fakePrefixDatums {
		__antithesis_instrumentation__.Notify(271760)
		fakePrefixDatums[i] = tree.DNull
	}
	__antithesis_instrumentation__.Notify(271750)

	if part.NumLists() == 0 && func() bool {
		__antithesis_instrumentation__.Notify(271761)
		return part.NumRanges() == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(271762)
		return errors.Newf("at least one of LIST or RANGE partitioning must be used")
	} else {
		__antithesis_instrumentation__.Notify(271763)
	}
	__antithesis_instrumentation__.Notify(271751)
	if part.NumLists() > 0 && func() bool {
		__antithesis_instrumentation__.Notify(271764)
		return part.NumRanges() > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(271765)
		return errors.Newf("only one LIST or RANGE partitioning may used")
	} else {
		__antithesis_instrumentation__.Notify(271766)
	}

	{
		__antithesis_instrumentation__.Notify(271767)
		for i := colOffset; i < colOffset+part.NumColumns(); i++ {
			__antithesis_instrumentation__.Notify(271768)

			if i >= idx.NumKeyColumns() {
				__antithesis_instrumentation__.Notify(271771)
				continue
			} else {
				__antithesis_instrumentation__.Notify(271772)
			}
			__antithesis_instrumentation__.Notify(271769)
			col, err := desc.FindColumnWithID(idx.GetKeyColumnID(i))
			if err != nil {
				__antithesis_instrumentation__.Notify(271773)
				return err
			} else {
				__antithesis_instrumentation__.Notify(271774)
			}
			__antithesis_instrumentation__.Notify(271770)
			if col.GetType().UserDefined() && func() bool {
				__antithesis_instrumentation__.Notify(271775)
				return !col.GetType().IsHydrated() == true
			}() == true {
				__antithesis_instrumentation__.Notify(271776)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(271777)
			}
		}
	}
	__antithesis_instrumentation__.Notify(271752)

	checkName := func(name string) error {
		__antithesis_instrumentation__.Notify(271778)
		if len(name) == 0 {
			__antithesis_instrumentation__.Notify(271781)
			return errors.Newf("PARTITION name must be non-empty")
		} else {
			__antithesis_instrumentation__.Notify(271782)
		}
		__antithesis_instrumentation__.Notify(271779)
		if indexName, exists := partitionNames[name]; exists {
			__antithesis_instrumentation__.Notify(271783)
			if indexName == idx.GetName() {
				__antithesis_instrumentation__.Notify(271784)
				return errors.Newf("PARTITION %s: name must be unique (used twice in index %q)",
					name, indexName)
			} else {
				__antithesis_instrumentation__.Notify(271785)
			}
		} else {
			__antithesis_instrumentation__.Notify(271786)
		}
		__antithesis_instrumentation__.Notify(271780)
		partitionNames[name] = idx.GetName()
		return nil
	}
	__antithesis_instrumentation__.Notify(271753)

	codec := keys.SystemSQLCodec

	if part.NumLists() > 0 {
		__antithesis_instrumentation__.Notify(271787)
		listValues := make(map[string]struct{}, part.NumLists())
		err := part.ForEachList(func(name string, values [][]byte, subPartitioning catalog.Partitioning) error {
			__antithesis_instrumentation__.Notify(271789)
			if err := checkName(name); err != nil {
				__antithesis_instrumentation__.Notify(271793)
				return err
			} else {
				__antithesis_instrumentation__.Notify(271794)
			}
			__antithesis_instrumentation__.Notify(271790)

			if len(values) == 0 {
				__antithesis_instrumentation__.Notify(271795)
				return errors.Newf("PARTITION %s: must contain values", name)
			} else {
				__antithesis_instrumentation__.Notify(271796)
			}
			__antithesis_instrumentation__.Notify(271791)

			for _, valueEncBuf := range values {
				__antithesis_instrumentation__.Notify(271797)
				tuple, keyPrefix, err := rowenc.DecodePartitionTuple(
					a, codec, desc, idx, part, valueEncBuf, fakePrefixDatums)
				if err != nil {
					__antithesis_instrumentation__.Notify(271800)
					return errors.Wrapf(err, "PARTITION %s", name)
				} else {
					__antithesis_instrumentation__.Notify(271801)
				}
				__antithesis_instrumentation__.Notify(271798)
				if _, exists := listValues[string(keyPrefix)]; exists {
					__antithesis_instrumentation__.Notify(271802)
					return errors.Newf("%s cannot be present in more than one partition", tuple)
				} else {
					__antithesis_instrumentation__.Notify(271803)
				}
				__antithesis_instrumentation__.Notify(271799)
				listValues[string(keyPrefix)] = struct{}{}
			}
			__antithesis_instrumentation__.Notify(271792)

			newColOffset := colOffset + part.NumColumns()
			return desc.validatePartitioningDescriptor(
				a, idx, subPartitioning, newColOffset, partitionNames,
			)
		})
		__antithesis_instrumentation__.Notify(271788)
		if err != nil {
			__antithesis_instrumentation__.Notify(271804)
			return err
		} else {
			__antithesis_instrumentation__.Notify(271805)
		}
	} else {
		__antithesis_instrumentation__.Notify(271806)
	}
	__antithesis_instrumentation__.Notify(271754)

	if part.NumRanges() > 0 {
		__antithesis_instrumentation__.Notify(271807)
		tree := interval.NewTree(interval.ExclusiveOverlapper)
		err := part.ForEachRange(func(name string, from, to []byte) error {
			__antithesis_instrumentation__.Notify(271809)
			if err := checkName(name); err != nil {
				__antithesis_instrumentation__.Notify(271815)
				return err
			} else {
				__antithesis_instrumentation__.Notify(271816)
			}
			__antithesis_instrumentation__.Notify(271810)

			fromDatums, fromKey, err := rowenc.DecodePartitionTuple(
				a, codec, desc, idx, part, from, fakePrefixDatums)
			if err != nil {
				__antithesis_instrumentation__.Notify(271817)
				return errors.Wrapf(err, "PARTITION %s", name)
			} else {
				__antithesis_instrumentation__.Notify(271818)
			}
			__antithesis_instrumentation__.Notify(271811)
			toDatums, toKey, err := rowenc.DecodePartitionTuple(
				a, codec, desc, idx, part, to, fakePrefixDatums)
			if err != nil {
				__antithesis_instrumentation__.Notify(271819)
				return errors.Wrapf(err, "PARTITION %s", name)
			} else {
				__antithesis_instrumentation__.Notify(271820)
			}
			__antithesis_instrumentation__.Notify(271812)
			pi := partitionInterval{name, fromKey, toKey}
			if overlaps := tree.Get(pi.Range()); len(overlaps) > 0 {
				__antithesis_instrumentation__.Notify(271821)
				return errors.Newf("partitions %s and %s overlap",
					overlaps[0].(partitionInterval).name, name)
			} else {
				__antithesis_instrumentation__.Notify(271822)
			}
			__antithesis_instrumentation__.Notify(271813)
			if err := tree.Insert(pi, false); errors.Is(err, interval.ErrEmptyRange) {
				__antithesis_instrumentation__.Notify(271823)
				return errors.Newf("PARTITION %s: empty range: lower bound %s is equal to upper bound %s",
					name, fromDatums, toDatums)
			} else {
				__antithesis_instrumentation__.Notify(271824)
				if errors.Is(err, interval.ErrInvertedRange) {
					__antithesis_instrumentation__.Notify(271825)
					return errors.Newf("PARTITION %s: empty range: lower bound %s is greater than upper bound %s",
						name, fromDatums, toDatums)
				} else {
					__antithesis_instrumentation__.Notify(271826)
					if err != nil {
						__antithesis_instrumentation__.Notify(271827)
						return errors.Wrapf(err, "PARTITION %s", name)
					} else {
						__antithesis_instrumentation__.Notify(271828)
					}
				}
			}
			__antithesis_instrumentation__.Notify(271814)
			return nil
		})
		__antithesis_instrumentation__.Notify(271808)
		if err != nil {
			__antithesis_instrumentation__.Notify(271829)
			return err
		} else {
			__antithesis_instrumentation__.Notify(271830)
		}
	} else {
		__antithesis_instrumentation__.Notify(271831)
	}
	__antithesis_instrumentation__.Notify(271755)

	return nil
}

type partitionInterval struct {
	name  string
	start roachpb.Key
	end   roachpb.Key
}

var _ interval.Interface = partitionInterval{}

func (ps partitionInterval) ID() uintptr { __antithesis_instrumentation__.Notify(271832); return 0 }

func (ps partitionInterval) Range() interval.Range {
	__antithesis_instrumentation__.Notify(271833)
	return interval.Range{Start: []byte(ps.start), End: []byte(ps.end)}
}

func (desc *wrapper) validatePartitioning() error {
	__antithesis_instrumentation__.Notify(271834)
	partitionNames := make(map[string]string)

	a := &tree.DatumAlloc{}
	return catalog.ForEachNonDropIndex(desc, func(idx catalog.Index) error {
		__antithesis_instrumentation__.Notify(271835)
		return desc.validatePartitioningDescriptor(
			a, idx, idx.GetPartitioning(), 0, partitionNames,
		)
	})
}
