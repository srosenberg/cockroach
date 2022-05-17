package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"
)

func (p *planner) ResolveOIDFromString(
	ctx context.Context, resultType *types.T, toResolve *tree.DString,
) (*tree.DOid, error) {
	__antithesis_instrumentation__.Notify(566582)
	ie := p.ExecCfg().InternalExecutorFactory(ctx, p.SessionData())
	return resolveOID(
		ctx, p.Txn(),
		ie,
		resultType, toResolve,
	)
}

func (p *planner) ResolveOIDFromOID(
	ctx context.Context, resultType *types.T, toResolve *tree.DOid,
) (*tree.DOid, error) {
	__antithesis_instrumentation__.Notify(566583)
	ie := p.ExecCfg().InternalExecutorFactory(ctx, p.SessionData())
	return resolveOID(
		ctx, p.Txn(),
		ie,
		resultType, toResolve,
	)
}

func resolveOID(
	ctx context.Context,
	txn *kv.Txn,
	ie sqlutil.InternalExecutor,
	resultType *types.T,
	toResolve tree.Datum,
) (*tree.DOid, error) {
	__antithesis_instrumentation__.Notify(566584)
	info, ok := regTypeInfos[resultType.Oid()]
	if !ok {
		__antithesis_instrumentation__.Notify(566589)
		return nil, pgerror.Newf(
			pgcode.InvalidTextRepresentation,
			"invalid input syntax for type %s: %q",
			resultType,
			tree.AsStringWithFlags(toResolve, tree.FmtBareStrings),
		)
	} else {
		__antithesis_instrumentation__.Notify(566590)
	}
	__antithesis_instrumentation__.Notify(566585)
	queryCol := info.nameCol
	if _, isOid := toResolve.(*tree.DOid); isOid {
		__antithesis_instrumentation__.Notify(566591)
		queryCol = "oid"
	} else {
		__antithesis_instrumentation__.Notify(566592)
	}
	__antithesis_instrumentation__.Notify(566586)
	q := fmt.Sprintf(
		"SELECT %s.oid, %s FROM pg_catalog.%s WHERE %s = $1",
		info.tableName, info.nameCol, info.tableName, queryCol,
	)

	results, err := ie.QueryRowEx(ctx, "queryOid", txn,
		sessiondata.NoSessionDataOverride, q, toResolve)
	if err != nil {
		__antithesis_instrumentation__.Notify(566593)
		if errors.HasType(err, (*tree.MultipleResultsError)(nil)) {
			__antithesis_instrumentation__.Notify(566595)
			return nil, pgerror.Newf(pgcode.AmbiguousAlias,
				"more than one %s named %s", info.objName, toResolve)
		} else {
			__antithesis_instrumentation__.Notify(566596)
		}
		__antithesis_instrumentation__.Notify(566594)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(566597)
	}
	__antithesis_instrumentation__.Notify(566587)
	if results.Len() == 0 {
		__antithesis_instrumentation__.Notify(566598)
		return nil, pgerror.Newf(info.errType,
			"%s %s does not exist", info.objName, toResolve)
	} else {
		__antithesis_instrumentation__.Notify(566599)
	}
	__antithesis_instrumentation__.Notify(566588)
	return tree.NewDOidWithName(
		results[0].(*tree.DOid).DInt,
		resultType,
		tree.AsStringWithFlags(results[1], tree.FmtBareStrings),
	), nil
}

type regTypeInfo struct {
	tableName string

	nameCol string

	objName string

	errType pgcode.Code
}

var regTypeInfos = map[oid.Oid]regTypeInfo{
	oid.T_regclass:     {"pg_class", "relname", "relation", pgcode.UndefinedTable},
	oid.T_regnamespace: {"pg_namespace", "nspname", "namespace", pgcode.UndefinedObject},
	oid.T_regproc:      {"pg_proc", "proname", "function", pgcode.UndefinedFunction},
	oid.T_regprocedure: {"pg_proc", "proname", "function", pgcode.UndefinedFunction},
	oid.T_regrole:      {"pg_authid", "rolname", "role", pgcode.UndefinedObject},
	oid.T_regtype:      {"pg_type", "typname", "type", pgcode.UndefinedObject},
}
