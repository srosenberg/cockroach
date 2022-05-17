package scbuild

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strings"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemadesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemaexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/seqexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scbuild/internal/scbuildstmt"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scdecomp"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/screl"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/catid"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"
)

var _ scbuildstmt.BuilderState = (*builderState)(nil)

func (b *builderState) QueryByID(id catid.DescID) scbuildstmt.ElementResultSet {
	__antithesis_instrumentation__.Notify(579343)
	if id == catid.InvalidDescID {
		__antithesis_instrumentation__.Notify(579345)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(579346)
	}
	__antithesis_instrumentation__.Notify(579344)
	b.ensureDescriptor(id)
	return b.descCache[id].ers
}

func (b *builderState) Ensure(
	current scpb.Status, target scpb.TargetStatus, e scpb.Element, meta scpb.TargetMetadata,
) {
	__antithesis_instrumentation__.Notify(579347)
	if e == nil {
		__antithesis_instrumentation__.Notify(579349)
		panic(errors.AssertionFailedf("cannot define target for nil element"))
	} else {
		__antithesis_instrumentation__.Notify(579350)
	}
	__antithesis_instrumentation__.Notify(579348)
	id := screl.GetDescID(e)
	b.ensureDescriptor(id)
	c := b.descCache[id]
	key := screl.ElementString(e)
	if i, ok := c.elementIndexMap[key]; ok {
		__antithesis_instrumentation__.Notify(579351)
		es := &b.output[i]
		if !screl.EqualElements(es.element, e) {
			__antithesis_instrumentation__.Notify(579354)
			panic(errors.AssertionFailedf("element key %v does not match element: %s",
				key, screl.ElementString(es.element)))
		} else {
			__antithesis_instrumentation__.Notify(579355)
		}
		__antithesis_instrumentation__.Notify(579352)
		if current != scpb.Status_UNKNOWN {
			__antithesis_instrumentation__.Notify(579356)
			es.current = current
		} else {
			__antithesis_instrumentation__.Notify(579357)
		}
		__antithesis_instrumentation__.Notify(579353)
		es.target = target
		es.element = e
		es.metadata = meta
	} else {
		__antithesis_instrumentation__.Notify(579358)
		if current == scpb.Status_UNKNOWN {
			__antithesis_instrumentation__.Notify(579360)
			if target == scpb.ToAbsent {
				__antithesis_instrumentation__.Notify(579362)
				panic(errors.AssertionFailedf("element not found: %s", screl.ElementString(e)))
			} else {
				__antithesis_instrumentation__.Notify(579363)
			}
			__antithesis_instrumentation__.Notify(579361)
			current = scpb.Status_ABSENT
		} else {
			__antithesis_instrumentation__.Notify(579364)
		}
		__antithesis_instrumentation__.Notify(579359)
		c.ers.indexes = append(c.ers.indexes, len(b.output))
		c.elementIndexMap[key] = len(b.output)
		b.output = append(b.output, elementState{
			element:  e,
			target:   target,
			current:  current,
			metadata: meta,
		})
	}

}

func (b *builderState) ForEachElementStatus(
	fn func(current scpb.Status, target scpb.TargetStatus, e scpb.Element),
) {
	__antithesis_instrumentation__.Notify(579365)
	for _, es := range b.output {
		__antithesis_instrumentation__.Notify(579366)
		fn(es.current, es.target, es.element)
	}
}

var _ scbuildstmt.PrivilegeChecker = (*builderState)(nil)

func (b *builderState) HasOwnership(e scpb.Element) bool {
	__antithesis_instrumentation__.Notify(579367)
	if b.hasAdmin {
		__antithesis_instrumentation__.Notify(579369)
		return true
	} else {
		__antithesis_instrumentation__.Notify(579370)
	}
	__antithesis_instrumentation__.Notify(579368)
	id := screl.GetDescID(e)
	b.ensureDescriptor(id)
	return b.descCache[id].hasOwnership
}

func (b *builderState) mustOwn(id catid.DescID) {
	__antithesis_instrumentation__.Notify(579371)
	if b.hasAdmin {
		__antithesis_instrumentation__.Notify(579373)
		return
	} else {
		__antithesis_instrumentation__.Notify(579374)
	}
	__antithesis_instrumentation__.Notify(579372)
	b.ensureDescriptor(id)
	if c := b.descCache[id]; !c.hasOwnership {
		__antithesis_instrumentation__.Notify(579375)
		panic(pgerror.Newf(pgcode.InsufficientPrivilege,
			"must be owner of %s %s", c.desc.DescriptorType(), c.desc.GetName()))
	} else {
		__antithesis_instrumentation__.Notify(579376)
	}
}

func (b *builderState) CheckPrivilege(e scpb.Element, privilege privilege.Kind) {
	__antithesis_instrumentation__.Notify(579377)
	b.checkPrivilege(screl.GetDescID(e), privilege)
}

func (b *builderState) checkPrivilege(id catid.DescID, privilege privilege.Kind) {
	__antithesis_instrumentation__.Notify(579378)
	b.ensureDescriptor(id)
	c := b.descCache[id]
	if c.hasOwnership {
		__antithesis_instrumentation__.Notify(579381)
		return
	} else {
		__antithesis_instrumentation__.Notify(579382)
	}
	__antithesis_instrumentation__.Notify(579379)
	err, found := c.privileges[privilege]
	if !found {
		__antithesis_instrumentation__.Notify(579383)
		err = b.auth.CheckPrivilege(b.ctx, c.desc, privilege)
		c.privileges[privilege] = err
	} else {
		__antithesis_instrumentation__.Notify(579384)
	}
	__antithesis_instrumentation__.Notify(579380)
	if err != nil {
		__antithesis_instrumentation__.Notify(579385)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(579386)
	}
}

var _ scbuildstmt.TableHelpers = (*builderState)(nil)

func (b *builderState) NextTableColumnID(table *scpb.Table) (ret catid.ColumnID) {
	__antithesis_instrumentation__.Notify(579387)
	{
		__antithesis_instrumentation__.Notify(579390)
		b.ensureDescriptor(table.TableID)
		desc := b.descCache[table.TableID].desc
		tbl, ok := desc.(catalog.TableDescriptor)
		if !ok {
			__antithesis_instrumentation__.Notify(579392)
			panic(errors.AssertionFailedf("Expected table descriptor for ID %d, instead got %s",
				desc.GetID(), desc.DescriptorType()))
		} else {
			__antithesis_instrumentation__.Notify(579393)
		}
		__antithesis_instrumentation__.Notify(579391)
		ret = tbl.GetNextColumnID()
	}
	__antithesis_instrumentation__.Notify(579388)
	scpb.ForEachColumn(b, func(_ scpb.Status, _ scpb.TargetStatus, column *scpb.Column) {
		__antithesis_instrumentation__.Notify(579394)
		if column.TableID == table.TableID && func() bool {
			__antithesis_instrumentation__.Notify(579395)
			return column.ColumnID >= ret == true
		}() == true {
			__antithesis_instrumentation__.Notify(579396)
			ret = column.ColumnID + 1
		} else {
			__antithesis_instrumentation__.Notify(579397)
		}
	})
	__antithesis_instrumentation__.Notify(579389)
	return ret
}

func (b *builderState) NextColumnFamilyID(table *scpb.Table) (ret catid.FamilyID) {
	__antithesis_instrumentation__.Notify(579398)
	{
		__antithesis_instrumentation__.Notify(579401)
		b.ensureDescriptor(table.TableID)
		desc := b.descCache[table.TableID].desc
		tbl, ok := desc.(catalog.TableDescriptor)
		if !ok {
			__antithesis_instrumentation__.Notify(579403)
			panic(errors.AssertionFailedf("Expected table descriptor for ID %d, instead got %s",
				desc.GetID(), desc.DescriptorType()))
		} else {
			__antithesis_instrumentation__.Notify(579404)
		}
		__antithesis_instrumentation__.Notify(579402)
		ret = tbl.GetNextFamilyID()
	}
	__antithesis_instrumentation__.Notify(579399)
	scpb.ForEachColumnFamily(b, func(_ scpb.Status, _ scpb.TargetStatus, cf *scpb.ColumnFamily) {
		__antithesis_instrumentation__.Notify(579405)
		if cf.TableID == table.TableID && func() bool {
			__antithesis_instrumentation__.Notify(579406)
			return cf.FamilyID >= ret == true
		}() == true {
			__antithesis_instrumentation__.Notify(579407)
			ret = cf.FamilyID + 1
		} else {
			__antithesis_instrumentation__.Notify(579408)
		}
	})
	__antithesis_instrumentation__.Notify(579400)
	return ret
}

func (b *builderState) NextTableIndexID(table *scpb.Table) (ret catid.IndexID) {
	__antithesis_instrumentation__.Notify(579409)
	return b.nextIndexID(table.TableID)
}

func (b *builderState) NextViewIndexID(view *scpb.View) (ret catid.IndexID) {
	__antithesis_instrumentation__.Notify(579410)
	if !view.IsMaterialized {
		__antithesis_instrumentation__.Notify(579412)
		panic(errors.AssertionFailedf("expected materialized view: %s", screl.ElementString(view)))
	} else {
		__antithesis_instrumentation__.Notify(579413)
	}
	__antithesis_instrumentation__.Notify(579411)
	return b.nextIndexID(view.ViewID)
}

func (b *builderState) nextIndexID(id catid.DescID) (ret catid.IndexID) {
	__antithesis_instrumentation__.Notify(579414)
	{
		__antithesis_instrumentation__.Notify(579418)
		b.ensureDescriptor(id)
		desc := b.descCache[id].desc
		tbl, ok := desc.(catalog.TableDescriptor)
		if !ok {
			__antithesis_instrumentation__.Notify(579420)
			panic(errors.AssertionFailedf("Expected table descriptor for ID %d, instead got %s",
				desc.GetID(), desc.DescriptorType()))
		} else {
			__antithesis_instrumentation__.Notify(579421)
		}
		__antithesis_instrumentation__.Notify(579419)
		ret = tbl.GetNextIndexID()
	}
	__antithesis_instrumentation__.Notify(579415)
	scpb.ForEachPrimaryIndex(b, func(_ scpb.Status, _ scpb.TargetStatus, index *scpb.PrimaryIndex) {
		__antithesis_instrumentation__.Notify(579422)
		if index.TableID == id && func() bool {
			__antithesis_instrumentation__.Notify(579423)
			return index.IndexID >= ret == true
		}() == true {
			__antithesis_instrumentation__.Notify(579424)
			ret = index.IndexID + 1
		} else {
			__antithesis_instrumentation__.Notify(579425)
		}
	})
	__antithesis_instrumentation__.Notify(579416)
	scpb.ForEachSecondaryIndex(b, func(_ scpb.Status, _ scpb.TargetStatus, index *scpb.SecondaryIndex) {
		__antithesis_instrumentation__.Notify(579426)
		if index.TableID == id && func() bool {
			__antithesis_instrumentation__.Notify(579427)
			return index.IndexID >= ret == true
		}() == true {
			__antithesis_instrumentation__.Notify(579428)
			ret = index.IndexID + 1
		} else {
			__antithesis_instrumentation__.Notify(579429)
		}
	})
	__antithesis_instrumentation__.Notify(579417)
	return ret
}

func (b *builderState) SecondaryIndexPartitioningDescriptor(
	index *scpb.SecondaryIndex, partBy *tree.PartitionBy,
) catpb.PartitioningDescriptor {
	__antithesis_instrumentation__.Notify(579430)
	b.ensureDescriptor(index.TableID)
	desc := b.descCache[index.TableID].desc
	tbl, ok := desc.(catalog.TableDescriptor)
	if !ok {
		__antithesis_instrumentation__.Notify(579436)
		panic(errors.AssertionFailedf("Expected table descriptor for ID %d, instead got %s",
			desc.GetID(), desc.DescriptorType()))
	} else {
		__antithesis_instrumentation__.Notify(579437)
	}
	__antithesis_instrumentation__.Notify(579431)
	var oldNumImplicitColumns int
	scpb.ForEachIndexPartitioning(b, func(_ scpb.Status, _ scpb.TargetStatus, p *scpb.IndexPartitioning) {
		__antithesis_instrumentation__.Notify(579438)
		if p.TableID != index.TableID || func() bool {
			__antithesis_instrumentation__.Notify(579440)
			return p.IndexID != index.IndexID == true
		}() == true {
			__antithesis_instrumentation__.Notify(579441)
			return
		} else {
			__antithesis_instrumentation__.Notify(579442)
		}
		__antithesis_instrumentation__.Notify(579439)
		oldNumImplicitColumns = int(p.PartitioningDescriptor.NumImplicitColumns)
	})
	__antithesis_instrumentation__.Notify(579432)
	oldKeyColumnNames := make([]string, len(index.KeyColumnIDs))
	for i, colID := range index.KeyColumnIDs {
		__antithesis_instrumentation__.Notify(579443)
		scpb.ForEachColumnName(b, func(_ scpb.Status, _ scpb.TargetStatus, cn *scpb.ColumnName) {
			__antithesis_instrumentation__.Notify(579444)
			if cn.TableID != index.TableID && func() bool {
				__antithesis_instrumentation__.Notify(579446)
				return cn.ColumnID != colID == true
			}() == true {
				__antithesis_instrumentation__.Notify(579447)
				return
			} else {
				__antithesis_instrumentation__.Notify(579448)
			}
			__antithesis_instrumentation__.Notify(579445)
			oldKeyColumnNames[i] = cn.Name
		})
	}
	__antithesis_instrumentation__.Notify(579433)
	var allowedNewColumnNames []tree.Name
	scpb.ForEachColumnName(b, func(current scpb.Status, target scpb.TargetStatus, cn *scpb.ColumnName) {
		__antithesis_instrumentation__.Notify(579449)
		if cn.TableID != index.TableID && func() bool {
			__antithesis_instrumentation__.Notify(579450)
			return current != scpb.Status_PUBLIC == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(579451)
			return target == scpb.ToPublic == true
		}() == true {
			__antithesis_instrumentation__.Notify(579452)
			allowedNewColumnNames = append(allowedNewColumnNames, tree.Name(cn.Name))
		} else {
			__antithesis_instrumentation__.Notify(579453)
		}
	})
	__antithesis_instrumentation__.Notify(579434)
	_, ret, err := b.createPartCCL(
		b.ctx,
		b.clusterSettings,
		b.evalCtx,
		tbl.FindColumnWithName,
		oldNumImplicitColumns,
		oldKeyColumnNames,
		partBy,
		allowedNewColumnNames,
		true,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(579454)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(579455)
	}
	__antithesis_instrumentation__.Notify(579435)
	return ret
}

func (b *builderState) ResolveTypeRef(ref tree.ResolvableTypeReference) scpb.TypeT {
	__antithesis_instrumentation__.Notify(579456)
	toType, err := tree.ResolveType(b.ctx, ref, b.cr)
	if err != nil {
		__antithesis_instrumentation__.Notify(579458)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(579459)
	}
	__antithesis_instrumentation__.Notify(579457)
	return newTypeT(toType)
}

func newTypeT(t *types.T) scpb.TypeT {
	__antithesis_instrumentation__.Notify(579460)
	m, err := typedesc.GetTypeDescriptorClosure(t)
	if err != nil {
		__antithesis_instrumentation__.Notify(579463)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(579464)
	}
	__antithesis_instrumentation__.Notify(579461)
	var ids catalog.DescriptorIDSet
	for id := range m {
		__antithesis_instrumentation__.Notify(579465)
		ids.Add(id)
	}
	__antithesis_instrumentation__.Notify(579462)
	return scpb.TypeT{Type: t, ClosedTypeIDs: ids.Ordered()}
}

func (b *builderState) WrapExpression(expr tree.Expr) *scpb.Expression {
	__antithesis_instrumentation__.Notify(579466)
	if expr == nil {
		__antithesis_instrumentation__.Notify(579470)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(579471)
	}
	__antithesis_instrumentation__.Notify(579467)

	var typeIDs catalog.DescriptorIDSet
	{
		__antithesis_instrumentation__.Notify(579472)
		visitor := &tree.TypeCollectorVisitor{OIDs: make(map[oid.Oid]struct{})}
		tree.WalkExpr(visitor, expr)
		for oid := range visitor.OIDs {
			__antithesis_instrumentation__.Notify(579473)
			if !types.IsOIDUserDefinedType(oid) {
				__antithesis_instrumentation__.Notify(579478)
				continue
			} else {
				__antithesis_instrumentation__.Notify(579479)
			}
			__antithesis_instrumentation__.Notify(579474)
			id, err := typedesc.UserDefinedTypeOIDToID(oid)
			if err != nil {
				__antithesis_instrumentation__.Notify(579480)
				panic(err)
			} else {
				__antithesis_instrumentation__.Notify(579481)
			}
			__antithesis_instrumentation__.Notify(579475)
			b.ensureDescriptor(id)
			typ, err := catalog.AsTypeDescriptor(b.descCache[id].desc)
			if err != nil {
				__antithesis_instrumentation__.Notify(579482)
				panic(err)
			} else {
				__antithesis_instrumentation__.Notify(579483)
			}
			__antithesis_instrumentation__.Notify(579476)
			ids, err := typ.GetIDClosure()
			if err != nil {
				__antithesis_instrumentation__.Notify(579484)
				panic(err)
			} else {
				__antithesis_instrumentation__.Notify(579485)
			}
			__antithesis_instrumentation__.Notify(579477)
			for id = range ids {
				__antithesis_instrumentation__.Notify(579486)
				typeIDs.Add(id)
			}
		}
	}
	__antithesis_instrumentation__.Notify(579468)

	var seqIDs catalog.DescriptorIDSet
	{
		__antithesis_instrumentation__.Notify(579487)
		seqIdentifiers, err := seqexpr.GetUsedSequences(expr)
		if err != nil {
			__antithesis_instrumentation__.Notify(579490)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(579491)
		}
		__antithesis_instrumentation__.Notify(579488)
		seqNameToID := make(map[string]int64)
		for _, seqIdentifier := range seqIdentifiers {
			__antithesis_instrumentation__.Notify(579492)
			if seqIdentifier.IsByID() {
				__antithesis_instrumentation__.Notify(579495)
				seqIDs.Add(catid.DescID(seqIdentifier.SeqID))
				continue
			} else {
				__antithesis_instrumentation__.Notify(579496)
			}
			__antithesis_instrumentation__.Notify(579493)
			uqName, err := parser.ParseTableName(seqIdentifier.SeqName)
			if err != nil {
				__antithesis_instrumentation__.Notify(579497)
				panic(err)
			} else {
				__antithesis_instrumentation__.Notify(579498)
			}
			__antithesis_instrumentation__.Notify(579494)
			elts := b.ResolveSequence(uqName, scbuildstmt.ResolveParams{
				IsExistenceOptional: false,
				RequiredPrivilege:   privilege.SELECT,
			})
			_, _, seq := scpb.FindSequence(elts)
			seqNameToID[seqIdentifier.SeqName] = int64(seq.SequenceID)
			seqIDs.Add(seq.SequenceID)
		}
		__antithesis_instrumentation__.Notify(579489)
		if len(seqNameToID) > 0 {
			__antithesis_instrumentation__.Notify(579499)
			expr, err = seqexpr.ReplaceSequenceNamesWithIDs(expr, seqNameToID)
			if err != nil {
				__antithesis_instrumentation__.Notify(579500)
				panic(err)
			} else {
				__antithesis_instrumentation__.Notify(579501)
			}
		} else {
			__antithesis_instrumentation__.Notify(579502)
		}
	}
	__antithesis_instrumentation__.Notify(579469)
	ret := &scpb.Expression{
		Expr:            catpb.Expression(tree.Serialize(expr)),
		UsesSequenceIDs: seqIDs.Ordered(),
		UsesTypeIDs:     typeIDs.Ordered(),
	}
	return ret
}

func (b *builderState) ComputedColumnExpression(
	tbl *scpb.Table, d *tree.ColumnTableDef,
) (tree.Expr, scpb.TypeT) {
	__antithesis_instrumentation__.Notify(579503)
	_, _, ns := scpb.FindNamespace(b.QueryByID(tbl.TableID))
	tn := tree.MakeTableNameFromPrefix(b.NamePrefix(tbl), tree.Name(ns.Name))
	b.ensureDescriptor(tbl.TableID)

	expr, typ, err := schemaexpr.ValidateComputedColumnExpression(
		b.ctx,
		b.descCache[tbl.TableID].desc.(catalog.TableDescriptor),
		d,
		&tn,
		"computed column",
		b.semaCtx,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(579506)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(579507)
	}
	__antithesis_instrumentation__.Notify(579504)
	parsedExpr, err := parser.ParseExpr(expr)
	if err != nil {
		__antithesis_instrumentation__.Notify(579508)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(579509)
	}
	__antithesis_instrumentation__.Notify(579505)
	return parsedExpr, newTypeT(typ)
}

var _ scbuildstmt.ElementReferences = (*builderState)(nil)

func (b *builderState) ForwardReferences(e scpb.Element) scbuildstmt.ElementResultSet {
	__antithesis_instrumentation__.Notify(579510)
	ids := screl.AllDescIDs(e)
	var c int
	ids.ForEach(func(id descpb.ID) {
		__antithesis_instrumentation__.Notify(579513)
		b.ensureDescriptor(id)
		c = c + len(b.descCache[id].ers.indexes)
	})
	__antithesis_instrumentation__.Notify(579511)
	ret := &elementResultSet{b: b, indexes: make([]int, 0, c)}
	ids.ForEach(func(id descpb.ID) {
		__antithesis_instrumentation__.Notify(579514)
		ret.indexes = append(ret.indexes, b.descCache[id].ers.indexes...)
	})
	__antithesis_instrumentation__.Notify(579512)
	return ret
}

func (b *builderState) BackReferences(id catid.DescID) scbuildstmt.ElementResultSet {
	__antithesis_instrumentation__.Notify(579515)
	if id == catid.InvalidDescID {
		__antithesis_instrumentation__.Notify(579519)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(579520)
	}
	__antithesis_instrumentation__.Notify(579516)
	var ids catalog.DescriptorIDSet
	{
		__antithesis_instrumentation__.Notify(579521)
		b.ensureDescriptor(id)
		c := b.descCache[id]
		c.backrefs.ForEach(ids.Add)
		c.backrefs.ForEach(b.ensureDescriptor)
		for i := range b.output {
			__antithesis_instrumentation__.Notify(579523)
			es := &b.output[i]
			if es.current == scpb.Status_PUBLIC || func() bool {
				__antithesis_instrumentation__.Notify(579527)
				return es.target != scpb.ToPublic == true
			}() == true {
				__antithesis_instrumentation__.Notify(579528)
				continue
			} else {
				__antithesis_instrumentation__.Notify(579529)
			}
			__antithesis_instrumentation__.Notify(579524)
			descID := screl.GetDescID(es.element)
			if ids.Contains(descID) || func() bool {
				__antithesis_instrumentation__.Notify(579530)
				return descID == id == true
			}() == true {
				__antithesis_instrumentation__.Notify(579531)
				continue
			} else {
				__antithesis_instrumentation__.Notify(579532)
			}
			__antithesis_instrumentation__.Notify(579525)
			if !screl.ContainsDescID(es.element, id) {
				__antithesis_instrumentation__.Notify(579533)
				continue
			} else {
				__antithesis_instrumentation__.Notify(579534)
			}
			__antithesis_instrumentation__.Notify(579526)
			ids.Add(descID)
		}
		__antithesis_instrumentation__.Notify(579522)
		ids.ForEach(b.ensureDescriptor)
	}
	__antithesis_instrumentation__.Notify(579517)
	ret := &elementResultSet{b: b}
	ids.ForEach(func(id descpb.ID) {
		__antithesis_instrumentation__.Notify(579535)
		ret.indexes = append(ret.indexes, b.descCache[id].ers.indexes...)
	})
	__antithesis_instrumentation__.Notify(579518)
	return ret
}

var _ scbuildstmt.NameResolver = (*builderState)(nil)

func (b *builderState) NamePrefix(e scpb.Element) tree.ObjectNamePrefix {
	__antithesis_instrumentation__.Notify(579536)
	id := screl.GetDescID(e)
	b.ensureDescriptor(id)
	return b.descCache[id].prefix
}

func (b *builderState) ResolveDatabase(
	name tree.Name, p scbuildstmt.ResolveParams,
) scbuildstmt.ElementResultSet {
	__antithesis_instrumentation__.Notify(579537)
	db := b.cr.MayResolveDatabase(b.ctx, name)
	if db == nil {
		__antithesis_instrumentation__.Notify(579539)
		if p.IsExistenceOptional {
			__antithesis_instrumentation__.Notify(579542)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(579543)
		}
		__antithesis_instrumentation__.Notify(579540)
		if string(name) == "" {
			__antithesis_instrumentation__.Notify(579544)
			panic(pgerror.New(pgcode.Syntax, "empty database name"))
		} else {
			__antithesis_instrumentation__.Notify(579545)
		}
		__antithesis_instrumentation__.Notify(579541)
		panic(sqlerrors.NewUndefinedDatabaseError(name.String()))
	} else {
		__antithesis_instrumentation__.Notify(579546)
	}
	__antithesis_instrumentation__.Notify(579538)
	b.ensureDescriptor(db.GetID())
	b.checkPrivilege(db.GetID(), p.RequiredPrivilege)
	return b.descCache[db.GetID()].ers
}

func (b *builderState) ResolveSchema(
	name tree.ObjectNamePrefix, p scbuildstmt.ResolveParams,
) scbuildstmt.ElementResultSet {
	__antithesis_instrumentation__.Notify(579547)
	_, sc := b.cr.MayResolveSchema(b.ctx, name)
	if sc == nil {
		__antithesis_instrumentation__.Notify(579550)
		if p.IsExistenceOptional {
			__antithesis_instrumentation__.Notify(579552)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(579553)
		}
		__antithesis_instrumentation__.Notify(579551)
		panic(sqlerrors.NewUndefinedSchemaError(name.Schema()))
	} else {
		__antithesis_instrumentation__.Notify(579554)
	}
	__antithesis_instrumentation__.Notify(579548)
	switch sc.SchemaKind() {
	case catalog.SchemaPublic, catalog.SchemaVirtual, catalog.SchemaTemporary:
		__antithesis_instrumentation__.Notify(579555)
		panic(pgerror.Newf(pgcode.InsufficientPrivilege,
			"%s permission denied for schema %q", p.RequiredPrivilege.String(), name))
	case catalog.SchemaUserDefined:
		__antithesis_instrumentation__.Notify(579556)
		b.ensureDescriptor(sc.GetID())
		b.mustOwn(sc.GetID())
	default:
		__antithesis_instrumentation__.Notify(579557)
		panic(errors.AssertionFailedf("unknown schema kind %d", sc.SchemaKind()))
	}
	__antithesis_instrumentation__.Notify(579549)
	return b.descCache[sc.GetID()].ers
}

func (b *builderState) ResolveEnumType(
	name *tree.UnresolvedObjectName, p scbuildstmt.ResolveParams,
) scbuildstmt.ElementResultSet {
	__antithesis_instrumentation__.Notify(579558)
	prefix, typ := b.cr.MayResolveType(b.ctx, *name)
	if typ == nil {
		__antithesis_instrumentation__.Notify(579561)
		if p.IsExistenceOptional {
			__antithesis_instrumentation__.Notify(579563)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(579564)
		}
		__antithesis_instrumentation__.Notify(579562)
		panic(sqlerrors.NewUndefinedTypeError(name))
	} else {
		__antithesis_instrumentation__.Notify(579565)
	}
	__antithesis_instrumentation__.Notify(579559)
	switch typ.GetKind() {
	case descpb.TypeDescriptor_ALIAS:
		__antithesis_instrumentation__.Notify(579566)

		panic(pgerror.Newf(pgcode.DependentObjectsStillExist,
			"%q is an implicit array type and cannot be modified", typ.GetName()))
	case descpb.TypeDescriptor_MULTIREGION_ENUM:
		__antithesis_instrumentation__.Notify(579567)
		typeName := tree.MakeTypeNameWithPrefix(prefix.NamePrefix(), typ.GetName())

		panic(errors.WithHintf(pgerror.Newf(pgcode.DependentObjectsStillExist,
			"%q is a multi-region enum and cannot be modified directly", typeName.FQString()),
			"try ALTER DATABASE %s DROP REGION %s", prefix.Database.GetName(), typ.GetName()))
	case descpb.TypeDescriptor_ENUM:
		__antithesis_instrumentation__.Notify(579568)
		b.ensureDescriptor(typ.GetID())
		b.mustOwn(typ.GetID())
	case descpb.TypeDescriptor_TABLE_IMPLICIT_RECORD_TYPE:
		__antithesis_instrumentation__.Notify(579569)

		panic(pgerror.Newf(pgcode.DependentObjectsStillExist,
			"cannot modify table record type %q", typ.GetName()))
	default:
		__antithesis_instrumentation__.Notify(579570)
		panic(errors.AssertionFailedf("unknown type kind %s", typ.GetKind()))
	}
	__antithesis_instrumentation__.Notify(579560)
	return b.descCache[typ.GetID()].ers
}

func (b *builderState) ResolveRelation(
	name *tree.UnresolvedObjectName, p scbuildstmt.ResolveParams,
) scbuildstmt.ElementResultSet {
	__antithesis_instrumentation__.Notify(579571)
	c := b.resolveRelation(name, p)
	if c == nil {
		__antithesis_instrumentation__.Notify(579573)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(579574)
	}
	__antithesis_instrumentation__.Notify(579572)
	return c.ers
}

func (b *builderState) resolveRelation(
	name *tree.UnresolvedObjectName, p scbuildstmt.ResolveParams,
) *cachedDesc {
	__antithesis_instrumentation__.Notify(579575)
	prefix, rel := b.cr.MayResolveTable(b.ctx, *name)
	if rel == nil {
		__antithesis_instrumentation__.Notify(579584)
		if p.IsExistenceOptional {
			__antithesis_instrumentation__.Notify(579586)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(579587)
		}
		__antithesis_instrumentation__.Notify(579585)
		panic(sqlerrors.NewUndefinedRelationError(name))
	} else {
		__antithesis_instrumentation__.Notify(579588)
	}
	__antithesis_instrumentation__.Notify(579576)
	if rel.IsVirtualTable() {
		__antithesis_instrumentation__.Notify(579589)
		if prefix.Schema.GetName() == catconstants.PgCatalogName {
			__antithesis_instrumentation__.Notify(579591)
			panic(pgerror.Newf(pgcode.InsufficientPrivilege,
				"%s is a system catalog", tree.ErrNameString(rel.GetName())))
		} else {
			__antithesis_instrumentation__.Notify(579592)
		}
		__antithesis_instrumentation__.Notify(579590)
		panic(pgerror.Newf(pgcode.WrongObjectType,
			"%s is a virtual object and cannot be modified", tree.ErrNameString(rel.GetName())))
	} else {
		__antithesis_instrumentation__.Notify(579593)
	}
	__antithesis_instrumentation__.Notify(579577)
	if rel.IsTemporary() {
		__antithesis_instrumentation__.Notify(579594)
		panic(scerrors.NotImplementedErrorf(nil, "dropping a temporary table"))
	} else {
		__antithesis_instrumentation__.Notify(579595)
	}
	__antithesis_instrumentation__.Notify(579578)

	b.ensureDescriptor(rel.GetID())
	c := b.descCache[rel.GetID()]
	b.ensureDescriptor(rel.GetParentSchemaID())
	if b.descCache[rel.GetParentSchemaID()].hasOwnership {
		__antithesis_instrumentation__.Notify(579596)
		c.hasOwnership = true
		return c
	} else {
		__antithesis_instrumentation__.Notify(579597)
	}
	__antithesis_instrumentation__.Notify(579579)
	err, found := c.privileges[p.RequiredPrivilege]
	if !found {
		__antithesis_instrumentation__.Notify(579598)
		err = b.auth.CheckPrivilege(b.ctx, rel, p.RequiredPrivilege)
		c.privileges[p.RequiredPrivilege] = err
	} else {
		__antithesis_instrumentation__.Notify(579599)
	}
	__antithesis_instrumentation__.Notify(579580)
	if err == nil {
		__antithesis_instrumentation__.Notify(579600)
		return c
	} else {
		__antithesis_instrumentation__.Notify(579601)
	}
	__antithesis_instrumentation__.Notify(579581)
	if p.RequiredPrivilege != privilege.CREATE {
		__antithesis_instrumentation__.Notify(579602)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(579603)
	}
	__antithesis_instrumentation__.Notify(579582)
	relationType := "table"
	if rel.IsView() {
		__antithesis_instrumentation__.Notify(579604)
		relationType = "view"
	} else {
		__antithesis_instrumentation__.Notify(579605)
		if rel.IsSequence() {
			__antithesis_instrumentation__.Notify(579606)
			relationType = "sequence"
		} else {
			__antithesis_instrumentation__.Notify(579607)
		}
	}
	__antithesis_instrumentation__.Notify(579583)
	panic(pgerror.Newf(pgcode.InsufficientPrivilege,
		"must be owner of %s %s or have %s privilege on %s %s",
		relationType,
		tree.Name(rel.GetName()),
		p.RequiredPrivilege,
		relationType,
		tree.Name(rel.GetName())))
}

func (b *builderState) ResolveTable(
	name *tree.UnresolvedObjectName, p scbuildstmt.ResolveParams,
) scbuildstmt.ElementResultSet {
	__antithesis_instrumentation__.Notify(579608)
	c := b.resolveRelation(name, p)
	if c == nil {
		__antithesis_instrumentation__.Notify(579611)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(579612)
	}
	__antithesis_instrumentation__.Notify(579609)
	if rel := c.desc.(catalog.TableDescriptor); !rel.IsTable() {
		__antithesis_instrumentation__.Notify(579613)
		panic(pgerror.Newf(pgcode.WrongObjectType, "%q is not a table", rel.GetName()))
	} else {
		__antithesis_instrumentation__.Notify(579614)
	}
	__antithesis_instrumentation__.Notify(579610)
	return c.ers
}

func (b *builderState) ResolveSequence(
	name *tree.UnresolvedObjectName, p scbuildstmt.ResolveParams,
) scbuildstmt.ElementResultSet {
	__antithesis_instrumentation__.Notify(579615)
	c := b.resolveRelation(name, p)
	if c == nil {
		__antithesis_instrumentation__.Notify(579618)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(579619)
	}
	__antithesis_instrumentation__.Notify(579616)
	if rel := c.desc.(catalog.TableDescriptor); !rel.IsSequence() {
		__antithesis_instrumentation__.Notify(579620)
		panic(pgerror.Newf(pgcode.WrongObjectType, "%q is not a sequence", rel.GetName()))
	} else {
		__antithesis_instrumentation__.Notify(579621)
	}
	__antithesis_instrumentation__.Notify(579617)
	return c.ers
}

func (b *builderState) ResolveView(
	name *tree.UnresolvedObjectName, p scbuildstmt.ResolveParams,
) scbuildstmt.ElementResultSet {
	__antithesis_instrumentation__.Notify(579622)
	c := b.resolveRelation(name, p)
	if c == nil {
		__antithesis_instrumentation__.Notify(579625)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(579626)
	}
	__antithesis_instrumentation__.Notify(579623)
	if rel := c.desc.(catalog.TableDescriptor); !rel.IsView() {
		__antithesis_instrumentation__.Notify(579627)
		panic(pgerror.Newf(pgcode.WrongObjectType, "%q is not a view", rel.GetName()))
	} else {
		__antithesis_instrumentation__.Notify(579628)
	}
	__antithesis_instrumentation__.Notify(579624)
	return c.ers
}

func (b *builderState) ResolveIndex(
	relationID catid.DescID, indexName tree.Name, p scbuildstmt.ResolveParams,
) scbuildstmt.ElementResultSet {
	__antithesis_instrumentation__.Notify(579629)
	b.ensureDescriptor(relationID)
	c := b.descCache[relationID]
	rel := c.desc.(catalog.TableDescriptor)
	if !rel.IsPhysicalTable() || func() bool {
		__antithesis_instrumentation__.Notify(579634)
		return rel.IsSequence() == true
	}() == true {
		__antithesis_instrumentation__.Notify(579635)
		panic(pgerror.Newf(pgcode.WrongObjectType,
			"%q is not an indexable table or a materialized view", rel.GetName()))
	} else {
		__antithesis_instrumentation__.Notify(579636)
	}
	__antithesis_instrumentation__.Notify(579630)
	b.checkPrivilege(rel.GetID(), p.RequiredPrivilege)
	var indexID catid.IndexID
	scpb.ForEachIndexName(c.ers, func(_ scpb.Status, _ scpb.TargetStatus, e *scpb.IndexName) {
		__antithesis_instrumentation__.Notify(579637)
		if e.TableID == relationID && func() bool {
			__antithesis_instrumentation__.Notify(579638)
			return tree.Name(e.Name) == indexName == true
		}() == true {
			__antithesis_instrumentation__.Notify(579639)
			indexID = e.IndexID
		} else {
			__antithesis_instrumentation__.Notify(579640)
		}
	})
	__antithesis_instrumentation__.Notify(579631)
	if indexID == 0 && func() bool {
		__antithesis_instrumentation__.Notify(579641)
		return (indexName == "" || func() bool {
			__antithesis_instrumentation__.Notify(579642)
			return indexName == tabledesc.LegacyPrimaryKeyIndexName == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(579643)
		indexID = rel.GetPrimaryIndexID()
	} else {
		__antithesis_instrumentation__.Notify(579644)
	}
	__antithesis_instrumentation__.Notify(579632)
	if indexID == 0 {
		__antithesis_instrumentation__.Notify(579645)
		if p.IsExistenceOptional {
			__antithesis_instrumentation__.Notify(579647)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(579648)
		}
		__antithesis_instrumentation__.Notify(579646)
		panic(pgerror.Newf(pgcode.UndefinedObject,
			"index %q not found in relation %q", indexName, rel.GetName()))
	} else {
		__antithesis_instrumentation__.Notify(579649)
	}
	__antithesis_instrumentation__.Notify(579633)
	return c.ers.Filter(func(_ scpb.Status, _ scpb.TargetStatus, element scpb.Element) bool {
		__antithesis_instrumentation__.Notify(579650)
		idI, _ := screl.Schema.GetAttribute(screl.IndexID, element)
		return idI != nil && func() bool {
			__antithesis_instrumentation__.Notify(579651)
			return idI.(catid.IndexID) == indexID == true
		}() == true
	})
}

func (b *builderState) ResolveColumn(
	relationID catid.DescID, columnName tree.Name, p scbuildstmt.ResolveParams,
) scbuildstmt.ElementResultSet {
	__antithesis_instrumentation__.Notify(579652)
	b.ensureDescriptor(relationID)
	c := b.descCache[relationID]
	rel := c.desc.(catalog.TableDescriptor)
	b.checkPrivilege(rel.GetID(), p.RequiredPrivilege)
	var columnID catid.ColumnID
	scpb.ForEachColumnName(c.ers, func(status scpb.Status, _ scpb.TargetStatus, e *scpb.ColumnName) {
		__antithesis_instrumentation__.Notify(579655)
		if e.TableID == relationID && func() bool {
			__antithesis_instrumentation__.Notify(579656)
			return tree.Name(e.Name) == columnName == true
		}() == true {
			__antithesis_instrumentation__.Notify(579657)
			columnID = e.ColumnID
		} else {
			__antithesis_instrumentation__.Notify(579658)
		}
	})
	__antithesis_instrumentation__.Notify(579653)
	if columnID == 0 {
		__antithesis_instrumentation__.Notify(579659)
		if p.IsExistenceOptional {
			__antithesis_instrumentation__.Notify(579661)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(579662)
		}
		__antithesis_instrumentation__.Notify(579660)
		panic(pgerror.Newf(pgcode.UndefinedColumn,
			"column %q not found in relation %q", columnName, rel.GetName()))
	} else {
		__antithesis_instrumentation__.Notify(579663)
	}
	__antithesis_instrumentation__.Notify(579654)
	return c.ers.Filter(func(_ scpb.Status, _ scpb.TargetStatus, e scpb.Element) bool {
		__antithesis_instrumentation__.Notify(579664)
		idI, _ := screl.Schema.GetAttribute(screl.ColumnID, e)
		return idI != nil && func() bool {
			__antithesis_instrumentation__.Notify(579665)
			return idI.(catid.ColumnID) == columnID == true
		}() == true
	})
}

func (b *builderState) ensureDescriptor(id catid.DescID) {
	__antithesis_instrumentation__.Notify(579666)
	if _, found := b.descCache[id]; found {
		__antithesis_instrumentation__.Notify(579671)
		return
	} else {
		__antithesis_instrumentation__.Notify(579672)
	}
	__antithesis_instrumentation__.Notify(579667)
	c := &cachedDesc{
		desc:            b.readDescriptor(id),
		privileges:      make(map[privilege.Kind]error),
		hasOwnership:    b.hasAdmin,
		ers:             &elementResultSet{b: b},
		elementIndexMap: map[string]int{},
	}

	if !c.hasOwnership {
		__antithesis_instrumentation__.Notify(579673)
		var err error
		c.hasOwnership, err = b.auth.HasOwnership(b.ctx, c.desc)
		if err != nil {
			__antithesis_instrumentation__.Notify(579674)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(579675)
		}
	} else {
		__antithesis_instrumentation__.Notify(579676)
	}
	__antithesis_instrumentation__.Notify(579668)

	b.descCache[id] = c
	crossRefLookupFn := func(id catid.DescID) catalog.Descriptor {
		__antithesis_instrumentation__.Notify(579677)
		return b.readDescriptor(id)
	}
	__antithesis_instrumentation__.Notify(579669)
	visitorFn := func(status scpb.Status, e scpb.Element) {
		__antithesis_instrumentation__.Notify(579678)
		c.ers.indexes = append(c.ers.indexes, len(b.output))
		key := screl.ElementString(e)
		c.elementIndexMap[key] = len(b.output)
		b.output = append(b.output, elementState{
			element: e,
			target:  scpb.AsTargetStatus(status),
			current: status,
		})
	}
	__antithesis_instrumentation__.Notify(579670)
	c.backrefs = scdecomp.WalkDescriptor(c.desc, crossRefLookupFn, visitorFn)

	switch d := c.desc.(type) {
	case catalog.DatabaseDescriptor:
		__antithesis_instrumentation__.Notify(579679)
		if !d.HasPublicSchemaWithDescriptor() {
			__antithesis_instrumentation__.Notify(579683)
			panic(scerrors.NotImplementedErrorf(nil,
				"database %q (%d) with a descriptorless public schema",
				d.GetName(), d.GetID()))
		} else {
			__antithesis_instrumentation__.Notify(579684)
		}
		__antithesis_instrumentation__.Notify(579680)

		childSchemas := b.cr.MustGetSchemasForDatabase(b.ctx, d)
		for schemaID, schemaName := range childSchemas {
			__antithesis_instrumentation__.Notify(579685)
			c.backrefs.Add(schemaID)
			if strings.HasPrefix(schemaName, catconstants.PgTempSchemaName) {
				__antithesis_instrumentation__.Notify(579686)
				b.tempSchemas[schemaID] = schemadesc.NewTemporarySchema(schemaName, schemaID, d.GetID())
			} else {
				__antithesis_instrumentation__.Notify(579687)
			}
		}
	case catalog.SchemaDescriptor:
		__antithesis_instrumentation__.Notify(579681)
		b.ensureDescriptor(c.desc.GetParentID())
		db := b.descCache[c.desc.GetParentID()].desc
		c.prefix.CatalogName = tree.Name(db.GetName())
		c.prefix.ExplicitCatalog = true

		_, objectIDs := b.cr.ReadObjectNamesAndIDs(b.ctx, db.(catalog.DatabaseDescriptor), d)
		for _, objectID := range objectIDs {
			__antithesis_instrumentation__.Notify(579688)
			c.backrefs.Add(objectID)
		}
	default:
		__antithesis_instrumentation__.Notify(579682)
		b.ensureDescriptor(c.desc.GetParentID())
		db := b.descCache[c.desc.GetParentID()].desc
		c.prefix.CatalogName = tree.Name(db.GetName())
		c.prefix.ExplicitCatalog = true
		b.ensureDescriptor(c.desc.GetParentSchemaID())
		sc := b.descCache[c.desc.GetParentSchemaID()].desc
		c.prefix.SchemaName = tree.Name(sc.GetName())
		c.prefix.ExplicitSchema = true
	}
}

func (b *builderState) readDescriptor(id catid.DescID) catalog.Descriptor {
	__antithesis_instrumentation__.Notify(579689)
	if id == catid.InvalidDescID {
		__antithesis_instrumentation__.Notify(579693)
		panic(errors.AssertionFailedf("invalid descriptor ID %d", id))
	} else {
		__antithesis_instrumentation__.Notify(579694)
	}
	__antithesis_instrumentation__.Notify(579690)
	if id == keys.SystemPublicSchemaID || func() bool {
		__antithesis_instrumentation__.Notify(579695)
		return id == keys.PublicSchemaIDForBackup == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(579696)
		return id == keys.PublicSchemaID == true
	}() == true {
		__antithesis_instrumentation__.Notify(579697)
		panic(scerrors.NotImplementedErrorf(nil, "descriptorless public schema %d", id))
	} else {
		__antithesis_instrumentation__.Notify(579698)
	}
	__antithesis_instrumentation__.Notify(579691)
	if tempSchema := b.tempSchemas[id]; tempSchema != nil {
		__antithesis_instrumentation__.Notify(579699)
		return tempSchema
	} else {
		__antithesis_instrumentation__.Notify(579700)
	}
	__antithesis_instrumentation__.Notify(579692)
	return b.cr.MustReadDescriptor(b.ctx, id)
}

type elementResultSet struct {
	b       *builderState
	indexes []int
}

var _ scbuildstmt.ElementResultSet = &elementResultSet{}

func (ers *elementResultSet) ForEachElementStatus(
	fn func(current scpb.Status, target scpb.TargetStatus, element scpb.Element),
) {
	__antithesis_instrumentation__.Notify(579701)
	if ers.IsEmpty() {
		__antithesis_instrumentation__.Notify(579703)
		return
	} else {
		__antithesis_instrumentation__.Notify(579704)
	}
	__antithesis_instrumentation__.Notify(579702)
	for _, i := range ers.indexes {
		__antithesis_instrumentation__.Notify(579705)
		es := &ers.b.output[i]
		fn(es.current, es.target, es.element)
	}
}

func (ers *elementResultSet) IsEmpty() bool {
	__antithesis_instrumentation__.Notify(579706)
	return ers == nil || func() bool {
		__antithesis_instrumentation__.Notify(579707)
		return len(ers.indexes) == 0 == true
	}() == true
}

func (ers *elementResultSet) Filter(
	predicate func(current scpb.Status, target scpb.TargetStatus, e scpb.Element) bool,
) scbuildstmt.ElementResultSet {
	__antithesis_instrumentation__.Notify(579708)
	return ers.filter(predicate)
}

func (ers *elementResultSet) filter(
	predicate func(current scpb.Status, target scpb.TargetStatus, e scpb.Element) bool,
) *elementResultSet {
	__antithesis_instrumentation__.Notify(579709)
	if ers.IsEmpty() {
		__antithesis_instrumentation__.Notify(579713)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(579714)
	}
	__antithesis_instrumentation__.Notify(579710)
	ret := elementResultSet{b: ers.b}
	for _, i := range ers.indexes {
		__antithesis_instrumentation__.Notify(579715)
		es := &ers.b.output[i]
		if predicate(es.current, es.target, es.element) {
			__antithesis_instrumentation__.Notify(579716)
			ret.indexes = append(ret.indexes, i)
		} else {
			__antithesis_instrumentation__.Notify(579717)
		}
	}
	__antithesis_instrumentation__.Notify(579711)
	if len(ret.indexes) == 0 {
		__antithesis_instrumentation__.Notify(579718)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(579719)
	}
	__antithesis_instrumentation__.Notify(579712)
	return &ret
}
