package colinfo

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

type NumResolutionResults int

const (
	NoResults NumResolutionResults = iota

	ExactlyOne

	MoreThanOne
)

type ColumnItemResolver interface {
	FindSourceMatchingName(
		ctx context.Context, tn tree.TableName,
	) (res NumResolutionResults, prefix *tree.TableName, srcMeta ColumnSourceMeta, err error)

	FindSourceProvidingColumn(
		ctx context.Context, col tree.Name,
	) (prefix *tree.TableName, srcMeta ColumnSourceMeta, colHint int, err error)

	Resolve(
		ctx context.Context, prefix *tree.TableName, srcMeta ColumnSourceMeta, colHint int, col tree.Name,
	) (ColumnResolutionResult, error)
}

type ColumnSourceMeta interface {
	ColumnSourceMeta()
}

type ColumnResolutionResult interface {
	ColumnResolutionResult()
}

func ResolveAllColumnsSelector(
	ctx context.Context, r ColumnItemResolver, a *tree.AllColumnsSelector,
) (srcName *tree.TableName, srcMeta ColumnSourceMeta, err error) {
	__antithesis_instrumentation__.Notify(251003)
	prefix := a.TableName.ToTableName()

	var res NumResolutionResults
	res, srcName, srcMeta, err = r.FindSourceMatchingName(ctx, prefix)
	if err != nil {
		__antithesis_instrumentation__.Notify(251007)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(251008)
	}
	__antithesis_instrumentation__.Notify(251004)
	if res == NoResults && func() bool {
		__antithesis_instrumentation__.Notify(251009)
		return a.TableName.NumParts == 2 == true
	}() == true {
		__antithesis_instrumentation__.Notify(251010)

		prefix.ExplicitCatalog = true
		prefix.CatalogName = prefix.SchemaName
		prefix.SchemaName = tree.PublicSchemaName
		res, srcName, srcMeta, err = r.FindSourceMatchingName(ctx, prefix)
		if err != nil {
			__antithesis_instrumentation__.Notify(251011)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(251012)
		}
	} else {
		__antithesis_instrumentation__.Notify(251013)
	}
	__antithesis_instrumentation__.Notify(251005)
	if res == NoResults {
		__antithesis_instrumentation__.Notify(251014)
		return nil, nil, pgerror.Newf(pgcode.UndefinedTable,
			"no data source matches pattern: %s", a)
	} else {
		__antithesis_instrumentation__.Notify(251015)
	}
	__antithesis_instrumentation__.Notify(251006)
	return srcName, srcMeta, nil
}

func ResolveColumnItem(
	ctx context.Context, r ColumnItemResolver, c *tree.ColumnItem,
) (ColumnResolutionResult, error) {
	__antithesis_instrumentation__.Notify(251016)
	colName := c.ColumnName
	if c.TableName == nil {
		__antithesis_instrumentation__.Notify(251021)

		srcName, srcMeta, cHint, err := r.FindSourceProvidingColumn(ctx, colName)
		if err != nil {
			__antithesis_instrumentation__.Notify(251023)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(251024)
		}
		__antithesis_instrumentation__.Notify(251022)
		return r.Resolve(ctx, srcName, srcMeta, cHint, colName)
	} else {
		__antithesis_instrumentation__.Notify(251025)
	}
	__antithesis_instrumentation__.Notify(251017)

	prefix := c.TableName.ToTableName()

	res, srcName, srcMeta, err := r.FindSourceMatchingName(ctx, prefix)
	if err != nil {
		__antithesis_instrumentation__.Notify(251026)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(251027)
	}
	__antithesis_instrumentation__.Notify(251018)
	if res == NoResults && func() bool {
		__antithesis_instrumentation__.Notify(251028)
		return c.TableName.NumParts == 2 == true
	}() == true {
		__antithesis_instrumentation__.Notify(251029)

		prefix.ExplicitCatalog = true
		prefix.CatalogName = prefix.SchemaName
		prefix.SchemaName = tree.PublicSchemaName
		res, srcName, srcMeta, err = r.FindSourceMatchingName(ctx, prefix)
		if err != nil {
			__antithesis_instrumentation__.Notify(251030)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(251031)
		}
	} else {
		__antithesis_instrumentation__.Notify(251032)
	}
	__antithesis_instrumentation__.Notify(251019)
	if res == NoResults {
		__antithesis_instrumentation__.Notify(251033)
		return nil, pgerror.Newf(pgcode.UndefinedTable,
			"no data source matches prefix: %s in this context", c.TableName)
	} else {
		__antithesis_instrumentation__.Notify(251034)
	}
	__antithesis_instrumentation__.Notify(251020)
	return r.Resolve(ctx, srcName, srcMeta, -1, colName)
}
