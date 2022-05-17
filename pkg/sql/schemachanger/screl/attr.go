package screl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"reflect"

	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/rel"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
)

type Attr int

func MustQuery(clauses ...rel.Clause) *rel.Query {
	__antithesis_instrumentation__.Notify(594981)
	q, err := rel.NewQuery(Schema, clauses...)
	if err != nil {
		__antithesis_instrumentation__.Notify(594983)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(594984)
	}
	__antithesis_instrumentation__.Notify(594982)
	return q
}

var elementProtoElementSelectors = func() (selectors []string) {
	__antithesis_instrumentation__.Notify(594985)
	elementProtoType := reflect.TypeOf((*scpb.ElementProto)(nil)).Elem()
	selectors = make([]string, elementProtoType.NumField())
	for i := 0; i < elementProtoType.NumField(); i++ {
		__antithesis_instrumentation__.Notify(594987)
		selectors[i] = elementProtoType.Field(i).Name
	}
	__antithesis_instrumentation__.Notify(594986)
	return selectors
}()

var _ rel.Attr = Attr(0)

const (
	_ Attr = iota

	DescID

	IndexID

	ColumnFamilyID

	ColumnID

	ConstraintID

	Name

	ReferencedDescID

	TargetStatus

	CurrentStatus

	Element

	Target

	AttrMax = iota - 1
)

var t = reflect.TypeOf

var elementSchemaOptions = []rel.SchemaOption{

	rel.AttrType(Element, t((*protoutil.Message)(nil)).Elem()),

	rel.EntityMapping(t((*scpb.Database)(nil)),
		rel.EntityAttr(DescID, "DatabaseID"),
	),
	rel.EntityMapping(t((*scpb.Schema)(nil)),
		rel.EntityAttr(DescID, "SchemaID"),
	),
	rel.EntityMapping(t((*scpb.AliasType)(nil)),
		rel.EntityAttr(DescID, "TypeID"),
	),
	rel.EntityMapping(t((*scpb.EnumType)(nil)),
		rel.EntityAttr(DescID, "TypeID"),
	),
	rel.EntityMapping(t((*scpb.View)(nil)),
		rel.EntityAttr(DescID, "ViewID"),
	),
	rel.EntityMapping(t((*scpb.Sequence)(nil)),
		rel.EntityAttr(DescID, "SequenceID"),
	),
	rel.EntityMapping(t((*scpb.Table)(nil)),
		rel.EntityAttr(DescID, "TableID"),
	),

	rel.EntityMapping(t((*scpb.ColumnFamily)(nil)),
		rel.EntityAttr(DescID, "TableID"),
		rel.EntityAttr(ColumnFamilyID, "FamilyID"),
		rel.EntityAttr(Name, "Name"),
	),
	rel.EntityMapping(t((*scpb.Column)(nil)),
		rel.EntityAttr(DescID, "TableID"),
		rel.EntityAttr(ColumnID, "ColumnID"),
	),
	rel.EntityMapping(t((*scpb.PrimaryIndex)(nil)),
		rel.EntityAttr(DescID, "TableID"),
		rel.EntityAttr(IndexID, "IndexID"),
	),
	rel.EntityMapping(t((*scpb.SecondaryIndex)(nil)),
		rel.EntityAttr(DescID, "TableID"),
		rel.EntityAttr(IndexID, "IndexID"),
	),
	rel.EntityMapping(t((*scpb.TemporaryIndex)(nil)),
		rel.EntityAttr(DescID, "TableID"),
		rel.EntityAttr(IndexID, "IndexID"),
	),
	rel.EntityMapping(t((*scpb.UniqueWithoutIndexConstraint)(nil)),
		rel.EntityAttr(DescID, "TableID"),
		rel.EntityAttr(ConstraintID, "ConstraintID"),
	),
	rel.EntityMapping(t((*scpb.CheckConstraint)(nil)),
		rel.EntityAttr(DescID, "TableID"),
		rel.EntityAttr(ConstraintID, "ConstraintID"),
	),
	rel.EntityMapping(t((*scpb.ForeignKeyConstraint)(nil)),
		rel.EntityAttr(DescID, "TableID"),
		rel.EntityAttr(ReferencedDescID, "ReferencedTableID"),
		rel.EntityAttr(ConstraintID, "ConstraintID"),
	),
	rel.EntityMapping(t((*scpb.RowLevelTTL)(nil)),
		rel.EntityAttr(DescID, "TableID"),
	),

	rel.EntityMapping(t((*scpb.TableLocalityGlobal)(nil)),
		rel.EntityAttr(DescID, "TableID"),
	),
	rel.EntityMapping(t((*scpb.TableLocalityPrimaryRegion)(nil)),
		rel.EntityAttr(DescID, "TableID"),
	),
	rel.EntityMapping(t((*scpb.TableLocalitySecondaryRegion)(nil)),
		rel.EntityAttr(DescID, "TableID"),
		rel.EntityAttr(ReferencedDescID, "RegionEnumTypeID"),
	),
	rel.EntityMapping(t((*scpb.TableLocalityRegionalByRow)(nil)),
		rel.EntityAttr(DescID, "TableID"),
	),

	rel.EntityMapping(t((*scpb.ColumnName)(nil)),
		rel.EntityAttr(DescID, "TableID"),
		rel.EntityAttr(ColumnID, "ColumnID"),
		rel.EntityAttr(Name, "Name"),
	),
	rel.EntityMapping(t((*scpb.ColumnType)(nil)),
		rel.EntityAttr(DescID, "TableID"),
		rel.EntityAttr(ColumnFamilyID, "FamilyID"),
		rel.EntityAttr(ColumnID, "ColumnID"),
	),
	rel.EntityMapping(t((*scpb.SequenceOwner)(nil)),
		rel.EntityAttr(DescID, "TableID"),
		rel.EntityAttr(ColumnID, "ColumnID"),
		rel.EntityAttr(ReferencedDescID, "SequenceID"),
	),
	rel.EntityMapping(t((*scpb.ColumnDefaultExpression)(nil)),
		rel.EntityAttr(DescID, "TableID"),
		rel.EntityAttr(ColumnID, "ColumnID"),
	),
	rel.EntityMapping(t((*scpb.ColumnOnUpdateExpression)(nil)),
		rel.EntityAttr(DescID, "TableID"),
		rel.EntityAttr(ColumnID, "ColumnID"),
	),

	rel.EntityMapping(t((*scpb.IndexName)(nil)),
		rel.EntityAttr(DescID, "TableID"),
		rel.EntityAttr(IndexID, "IndexID"),
		rel.EntityAttr(Name, "Name"),
	),
	rel.EntityMapping(t((*scpb.IndexPartitioning)(nil)),
		rel.EntityAttr(DescID, "TableID"),
		rel.EntityAttr(IndexID, "IndexID"),
	),
	rel.EntityMapping(t((*scpb.SecondaryIndexPartial)(nil)),
		rel.EntityAttr(DescID, "TableID"),
		rel.EntityAttr(IndexID, "IndexID"),
	),

	rel.EntityMapping(t((*scpb.ConstraintName)(nil)),
		rel.EntityAttr(DescID, "TableID"),
		rel.EntityAttr(ConstraintID, "ConstraintID"),
		rel.EntityAttr(Name, "Name"),
	),

	rel.EntityMapping(t((*scpb.Namespace)(nil)),
		rel.EntityAttr(DescID, "DescriptorID"),
		rel.EntityAttr(ReferencedDescID, "DatabaseID"),
		rel.EntityAttr(Name, "Name"),
	),
	rel.EntityMapping(t((*scpb.Owner)(nil)),
		rel.EntityAttr(DescID, "DescriptorID"),
	),
	rel.EntityMapping(t((*scpb.UserPrivileges)(nil)),
		rel.EntityAttr(DescID, "DescriptorID"),
		rel.EntityAttr(Name, "UserName"),
	),

	rel.EntityMapping(t((*scpb.DatabaseRegionConfig)(nil)),
		rel.EntityAttr(DescID, "DatabaseID"),
		rel.EntityAttr(ReferencedDescID, "RegionEnumTypeID"),
	),
	rel.EntityMapping(t((*scpb.DatabaseRoleSetting)(nil)),
		rel.EntityAttr(DescID, "DatabaseID"),
		rel.EntityAttr(Name, "RoleName"),
	),

	rel.EntityMapping(t((*scpb.SchemaParent)(nil)),
		rel.EntityAttr(DescID, "SchemaID"),
		rel.EntityAttr(ReferencedDescID, "ParentDatabaseID"),
	),
	rel.EntityMapping(t((*scpb.ObjectParent)(nil)),
		rel.EntityAttr(DescID, "ObjectID"),
		rel.EntityAttr(ReferencedDescID, "ParentSchemaID"),
	),

	rel.EntityMapping(t((*scpb.TableComment)(nil)),
		rel.EntityAttr(DescID, "TableID"),
	),
	rel.EntityMapping(t((*scpb.DatabaseComment)(nil)),
		rel.EntityAttr(DescID, "DatabaseID"),
	),
	rel.EntityMapping(t((*scpb.SchemaComment)(nil)),
		rel.EntityAttr(DescID, "SchemaID"),
	),
	rel.EntityMapping(t((*scpb.ColumnComment)(nil)),
		rel.EntityAttr(DescID, "TableID"),
		rel.EntityAttr(ColumnID, "ColumnID"),
	),
	rel.EntityMapping(t((*scpb.IndexComment)(nil)),
		rel.EntityAttr(DescID, "TableID"),
		rel.EntityAttr(IndexID, "IndexID"),
	),
	rel.EntityMapping(t((*scpb.ConstraintComment)(nil)),
		rel.EntityAttr(DescID, "TableID"),
		rel.EntityAttr(ConstraintID, "ConstraintID"),
	),
}

var Schema = rel.MustSchema("screl", append(
	elementSchemaOptions,
	rel.EntityMapping(t((*Node)(nil)),
		rel.EntityAttr(CurrentStatus, "CurrentStatus"),
		rel.EntityAttr(Target, "Target"),
	),
	rel.EntityMapping(t((*scpb.Target)(nil)),
		rel.EntityAttr(TargetStatus, "TargetStatus"),
		rel.EntityAttr(Element, elementProtoElementSelectors...),
	),
)...)

func JoinTarget(element, target rel.Var) rel.Clause {
	__antithesis_instrumentation__.Notify(594988)
	return rel.And(
		target.Type((*scpb.Target)(nil)),
		target.AttrEqVar(Element, element),
	)
}

func JoinTargetNode(element, target, node rel.Var) rel.Clause {
	__antithesis_instrumentation__.Notify(594989)
	return rel.And(
		JoinTarget(element, target),
		node.Type((*Node)(nil)),
		node.AttrEqVar(Target, target),
	)
}
