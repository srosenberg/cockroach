package sqlsmith

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

type colRef struct {
	typ  *types.T
	item *tree.ColumnItem
}

func (c *colRef) typedExpr() tree.TypedExpr {
	__antithesis_instrumentation__.Notify(69860)
	return makeTypedExpr(c.item, c.typ)
}

type colRefs []*colRef

func (t colRefs) extend(refs ...*colRef) colRefs {
	__antithesis_instrumentation__.Notify(69861)
	ret := append(make(colRefs, 0, len(t)+len(refs)), t...)
	ret = append(ret, refs...)
	return ret
}

func (t colRefs) stripTableName() {
	__antithesis_instrumentation__.Notify(69862)
	for _, c := range t {
		__antithesis_instrumentation__.Notify(69863)
		c.item.TableName = nil
	}
}

func (s *Smither) canRecurse() bool {
	__antithesis_instrumentation__.Notify(69864)
	return s.complexity > s.rnd.Float64()
}

type Context struct {
	fnClass  tree.FunctionClass
	noWindow bool
}

var (
	emptyCtx   = Context{}
	groupByCtx = Context{fnClass: tree.AggregateClass}
	havingCtx  = Context{
		fnClass:  tree.AggregateClass,
		noWindow: true,
	}
	windowCtx = Context{fnClass: tree.WindowClass}
)
