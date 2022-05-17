package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"strings"
)

func (ctx *FmtCtx) formatNodeOrHideConstants(n NodeFormatter) {
	__antithesis_instrumentation__.Notify(609935)
	if ctx.flags.HasFlags(FmtHideConstants) {
		__antithesis_instrumentation__.Notify(609937)
		switch v := n.(type) {
		case *ValuesClause:
			__antithesis_instrumentation__.Notify(609938)
			v.formatHideConstants(ctx)
			return
		case *Tuple:
			__antithesis_instrumentation__.Notify(609939)
			v.formatHideConstants(ctx)
			return
		case *Array:
			__antithesis_instrumentation__.Notify(609940)
			v.formatHideConstants(ctx)
			return
		case *Placeholder:
			__antithesis_instrumentation__.Notify(609941)

		case *StrVal:
			__antithesis_instrumentation__.Notify(609942)
			ctx.WriteString("'_'")
			return
		case Datum, Constant:
			__antithesis_instrumentation__.Notify(609943)
			ctx.WriteByte('_')
			return
		}
	} else {
		__antithesis_instrumentation__.Notify(609944)
	}
	__antithesis_instrumentation__.Notify(609936)
	n.Format(ctx)
}

func (node *ValuesClause) formatHideConstants(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609945)
	ctx.WriteString("VALUES (")
	node.Rows[0].formatHideConstants(ctx)
	ctx.WriteByte(')')
	if len(node.Rows) > 1 {
		__antithesis_instrumentation__.Notify(609946)
		ctx.Printf(", (%s)", arityString(len(node.Rows)-1))
	} else {
		__antithesis_instrumentation__.Notify(609947)
	}
}

func (node *Exprs) formatHideConstants(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609948)
	exprs := *node
	if len(exprs) < 2 {
		__antithesis_instrumentation__.Notify(609952)
		node.Format(ctx)
		return
	} else {
		__antithesis_instrumentation__.Notify(609953)
	}
	__antithesis_instrumentation__.Notify(609949)

	var i int
	for i = 0; i < len(exprs); i++ {
		__antithesis_instrumentation__.Notify(609954)
		switch exprs[i].(type) {
		case Datum, Constant, *Placeholder:
			__antithesis_instrumentation__.Notify(609956)
			continue
		}
		__antithesis_instrumentation__.Notify(609955)
		break
	}
	__antithesis_instrumentation__.Notify(609950)

	if i == len(exprs) {
		__antithesis_instrumentation__.Notify(609957)

		v2 := append(make(Exprs, 0, 3), exprs[:2]...)
		if len(exprs) > 2 {
			__antithesis_instrumentation__.Notify(609959)
			v2 = append(v2, arityIndicator(len(exprs)-2))
		} else {
			__antithesis_instrumentation__.Notify(609960)
		}
		__antithesis_instrumentation__.Notify(609958)
		v2.Format(ctx)
		return
	} else {
		__antithesis_instrumentation__.Notify(609961)
	}
	__antithesis_instrumentation__.Notify(609951)
	node.Format(ctx)
}

func (node *Tuple) formatHideConstants(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609962)
	if len(node.Exprs) < 2 {
		__antithesis_instrumentation__.Notify(609966)
		node.Format(ctx)
		return
	} else {
		__antithesis_instrumentation__.Notify(609967)
	}
	__antithesis_instrumentation__.Notify(609963)

	var i int
	for i = 0; i < len(node.Exprs); i++ {
		__antithesis_instrumentation__.Notify(609968)
		switch node.Exprs[i].(type) {
		case Datum, Constant, *Placeholder:
			__antithesis_instrumentation__.Notify(609970)
			continue
		}
		__antithesis_instrumentation__.Notify(609969)
		break
	}
	__antithesis_instrumentation__.Notify(609964)

	if i == len(node.Exprs) {
		__antithesis_instrumentation__.Notify(609971)

		v2 := *node
		v2.Exprs = append(make(Exprs, 0, 3), v2.Exprs[:2]...)
		if len(node.Exprs) > 2 {
			__antithesis_instrumentation__.Notify(609973)
			v2.Exprs = append(v2.Exprs, arityIndicator(len(node.Exprs)-2))
			if node.Labels != nil {
				__antithesis_instrumentation__.Notify(609974)
				v2.Labels = node.Labels[:2]
			} else {
				__antithesis_instrumentation__.Notify(609975)
			}
		} else {
			__antithesis_instrumentation__.Notify(609976)
		}
		__antithesis_instrumentation__.Notify(609972)
		v2.Format(ctx)
		return
	} else {
		__antithesis_instrumentation__.Notify(609977)
	}
	__antithesis_instrumentation__.Notify(609965)
	node.Format(ctx)
}

func (node *Array) formatHideConstants(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609978)
	if len(node.Exprs) < 2 {
		__antithesis_instrumentation__.Notify(609982)
		node.Format(ctx)
		return
	} else {
		__antithesis_instrumentation__.Notify(609983)
	}
	__antithesis_instrumentation__.Notify(609979)

	var i int
	for i = 0; i < len(node.Exprs); i++ {
		__antithesis_instrumentation__.Notify(609984)
		switch node.Exprs[i].(type) {
		case Datum, Constant, *Placeholder:
			__antithesis_instrumentation__.Notify(609986)
			continue
		}
		__antithesis_instrumentation__.Notify(609985)
		break
	}
	__antithesis_instrumentation__.Notify(609980)

	if i == len(node.Exprs) {
		__antithesis_instrumentation__.Notify(609987)

		v2 := *node
		v2.Exprs = append(make(Exprs, 0, 3), v2.Exprs[:2]...)
		if len(node.Exprs) > 2 {
			__antithesis_instrumentation__.Notify(609989)
			v2.Exprs = append(v2.Exprs, arityIndicator(len(node.Exprs)-2))
		} else {
			__antithesis_instrumentation__.Notify(609990)
		}
		__antithesis_instrumentation__.Notify(609988)
		v2.Format(ctx)
		return
	} else {
		__antithesis_instrumentation__.Notify(609991)
	}
	__antithesis_instrumentation__.Notify(609981)
	node.Format(ctx)
}

func arityIndicator(n int) Expr {
	__antithesis_instrumentation__.Notify(609992)
	return NewUnresolvedName(arityString(n))
}

func arityString(n int) string {
	__antithesis_instrumentation__.Notify(609993)
	var v int
	for v = 1; n >= 10; n /= 10 {
		__antithesis_instrumentation__.Notify(609995)
		v = v * 10
	}
	__antithesis_instrumentation__.Notify(609994)
	v = v * n
	return fmt.Sprintf("__more%d__", v)
}

func isArityIndicatorString(s string) bool {
	__antithesis_instrumentation__.Notify(609996)
	return strings.HasPrefix(s, "__more")
}
