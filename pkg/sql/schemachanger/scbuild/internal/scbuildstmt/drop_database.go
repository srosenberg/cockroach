package scbuildstmt

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/catid"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

func DropDatabase(b BuildCtx, n *tree.DropDatabase) {
	__antithesis_instrumentation__.Notify(579994)
	elts := b.ResolveDatabase(n.Name, ResolveParams{
		IsExistenceOptional: n.IfExists,
		RequiredPrivilege:   privilege.DROP,
	})
	_, _, db := scpb.FindDatabase(elts)
	if db == nil {
		__antithesis_instrumentation__.Notify(580002)
		return
	} else {
		__antithesis_instrumentation__.Notify(580003)
	}
	__antithesis_instrumentation__.Notify(579995)
	if string(n.Name) == b.SessionData().Database && func() bool {
		__antithesis_instrumentation__.Notify(580004)
		return b.SessionData().SafeUpdates == true
	}() == true {
		__antithesis_instrumentation__.Notify(580005)
		panic(pgerror.DangerousStatementf("DROP DATABASE on current database"))
	} else {
		__antithesis_instrumentation__.Notify(580006)
	}
	__antithesis_instrumentation__.Notify(579996)
	b.IncrementSchemaChangeDropCounter("database")

	if n.DropBehavior == tree.DropCascade || func() bool {
		__antithesis_instrumentation__.Notify(580007)
		return (n.DropBehavior == tree.DropDefault && func() bool {
			__antithesis_instrumentation__.Notify(580008)
			return !b.SessionData().SafeUpdates == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(580009)
		dropCascadeDescriptor(b, db.DatabaseID)
		return
	} else {
		__antithesis_instrumentation__.Notify(580010)
	}
	__antithesis_instrumentation__.Notify(579997)

	if !dropRestrictDescriptor(b, db.DatabaseID) {
		__antithesis_instrumentation__.Notify(580011)
		return
	} else {
		__antithesis_instrumentation__.Notify(580012)
	}
	__antithesis_instrumentation__.Notify(579998)

	var publicSchemaID catid.DescID
	b.BackReferences(db.DatabaseID).ForEachElementStatus(func(_ scpb.Status, _ scpb.TargetStatus, e scpb.Element) {
		__antithesis_instrumentation__.Notify(580013)
		switch t := e.(type) {
		case *scpb.Schema:
			__antithesis_instrumentation__.Notify(580014)
			if t.IsPublic {
				__antithesis_instrumentation__.Notify(580015)
				publicSchemaID = t.SchemaID
			} else {
				__antithesis_instrumentation__.Notify(580016)
			}
		}
	})
	__antithesis_instrumentation__.Notify(579999)
	dropRestrictDescriptor(b, publicSchemaID)
	dbBackrefs := undroppedBackrefs(b, db.DatabaseID)
	publicSchemaBackrefs := undroppedBackrefs(b, publicSchemaID)
	if dbBackrefs.IsEmpty() && func() bool {
		__antithesis_instrumentation__.Notify(580017)
		return publicSchemaBackrefs.IsEmpty() == true
	}() == true {
		__antithesis_instrumentation__.Notify(580018)
		return
	} else {
		__antithesis_instrumentation__.Notify(580019)
	}
	__antithesis_instrumentation__.Notify(580000)

	if n.DropBehavior == tree.DropRestrict {
		__antithesis_instrumentation__.Notify(580020)
		panic(pgerror.Newf(pgcode.DependentObjectsStillExist,
			"database %q is not empty and RESTRICT was specified", simpleName(b, db.DatabaseID)))
	} else {
		__antithesis_instrumentation__.Notify(580021)
	}
	__antithesis_instrumentation__.Notify(580001)
	panic(pgerror.DangerousStatementf(
		"DROP DATABASE on non-empty database without explicit CASCADE"))
}
