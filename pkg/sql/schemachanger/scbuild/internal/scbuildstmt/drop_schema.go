package scbuildstmt

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/catid"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
)

func DropSchema(b BuildCtx, n *tree.DropSchema) {
	__antithesis_instrumentation__.Notify(580022)
	var toCheckBackrefs []catid.DescID
	for _, name := range n.Names {
		__antithesis_instrumentation__.Notify(580024)
		elts := b.ResolveSchema(name, ResolveParams{
			IsExistenceOptional: n.IfExists,
			RequiredPrivilege:   privilege.DROP,
		})
		_, _, sc := scpb.FindSchema(elts)
		if sc == nil {
			__antithesis_instrumentation__.Notify(580029)
			continue
		} else {
			__antithesis_instrumentation__.Notify(580030)
		}
		__antithesis_instrumentation__.Notify(580025)
		if sc.IsVirtual || func() bool {
			__antithesis_instrumentation__.Notify(580031)
			return sc.IsPublic == true
		}() == true {
			__antithesis_instrumentation__.Notify(580032)
			panic(pgerror.Newf(pgcode.InvalidSchemaName,
				"cannot drop schema %q", simpleName(b, sc.SchemaID)))
		} else {
			__antithesis_instrumentation__.Notify(580033)
		}
		__antithesis_instrumentation__.Notify(580026)
		if sc.IsTemporary {
			__antithesis_instrumentation__.Notify(580034)
			panic(scerrors.NotImplementedErrorf(n, "dropping a temporary schema"))
		} else {
			__antithesis_instrumentation__.Notify(580035)
		}
		__antithesis_instrumentation__.Notify(580027)
		if n.DropBehavior == tree.DropCascade {
			__antithesis_instrumentation__.Notify(580036)
			dropCascadeDescriptor(b, sc.SchemaID)
			toCheckBackrefs = append(toCheckBackrefs, sc.SchemaID)
		} else {
			__antithesis_instrumentation__.Notify(580037)
			if dropRestrictDescriptor(b, sc.SchemaID) {
				__antithesis_instrumentation__.Notify(580038)
				toCheckBackrefs = append(toCheckBackrefs, sc.SchemaID)
			} else {
				__antithesis_instrumentation__.Notify(580039)
			}
		}
		__antithesis_instrumentation__.Notify(580028)
		b.IncrementSubWorkID()
		b.IncrementSchemaChangeDropCounter("schema")
		b.IncrementUserDefinedSchemaCounter(sqltelemetry.UserDefinedSchemaDrop)
	}
	__antithesis_instrumentation__.Notify(580023)

	for _, schemaID := range toCheckBackrefs {
		__antithesis_instrumentation__.Notify(580040)
		if n.DropBehavior == tree.DropCascade {
			__antithesis_instrumentation__.Notify(580041)

			var objectIDs, typeIDs catalog.DescriptorIDSet
			scpb.ForEachObjectParent(b.BackReferences(schemaID), func(_ scpb.Status, _ scpb.TargetStatus, op *scpb.ObjectParent) {
				__antithesis_instrumentation__.Notify(580044)
				objectIDs.Add(op.ObjectID)
			})
			__antithesis_instrumentation__.Notify(580042)
			objectIDs.ForEach(func(id descpb.ID) {
				__antithesis_instrumentation__.Notify(580045)
				elts := b.QueryByID(id)
				if _, _, enum := scpb.FindEnumType(elts); enum != nil {
					__antithesis_instrumentation__.Notify(580046)
					typeIDs.Add(enum.TypeID)
				} else {
					__antithesis_instrumentation__.Notify(580047)
					if _, _, alias := scpb.FindAliasType(elts); alias != nil {
						__antithesis_instrumentation__.Notify(580048)
						typeIDs.Add(alias.TypeID)
					} else {
						__antithesis_instrumentation__.Notify(580049)
					}
				}
			})
			__antithesis_instrumentation__.Notify(580043)
			typeIDs.ForEach(func(id descpb.ID) {
				__antithesis_instrumentation__.Notify(580050)
				if dependentNames := dependentTypeNames(b, id); len(dependentNames) > 0 {
					__antithesis_instrumentation__.Notify(580051)
					panic(unimplemented.NewWithIssueDetailf(51480, "DROP TYPE CASCADE is not yet supported",
						"cannot drop type %q because other objects (%v) still depend on it",
						qualifiedName(b, id), dependentNames))
				} else {
					__antithesis_instrumentation__.Notify(580052)
				}
			})
		} else {
			__antithesis_instrumentation__.Notify(580053)
			backrefs := undroppedBackrefs(b, schemaID)
			if backrefs.IsEmpty() {
				__antithesis_instrumentation__.Notify(580055)
				continue
			} else {
				__antithesis_instrumentation__.Notify(580056)
			}
			__antithesis_instrumentation__.Notify(580054)
			panic(pgerror.Newf(pgcode.DependentObjectsStillExist,
				"schema %q is not empty and CASCADE was not specified", simpleName(b, schemaID)))
		}
	}
}
