package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	"github.com/cockroachdb/errors"
)

type Visitor interface {
	VisitPre(expr Expr) (recurse bool, newExpr Expr)

	VisitPost(expr Expr) (newNode Expr)
}

func (expr *AndExpr) Walk(v Visitor) Expr {
	__antithesis_instrumentation__.Notify(615954)
	left, changedL := WalkExpr(v, expr.Left)
	right, changedR := WalkExpr(v, expr.Right)
	if changedL || func() bool {
		__antithesis_instrumentation__.Notify(615956)
		return changedR == true
	}() == true {
		__antithesis_instrumentation__.Notify(615957)
		exprCopy := *expr
		exprCopy.Left = left
		exprCopy.Right = right
		return &exprCopy
	} else {
		__antithesis_instrumentation__.Notify(615958)
	}
	__antithesis_instrumentation__.Notify(615955)
	return expr
}

func (expr *AnnotateTypeExpr) Walk(v Visitor) Expr {
	__antithesis_instrumentation__.Notify(615959)
	e, changed := WalkExpr(v, expr.Expr)
	if changed {
		__antithesis_instrumentation__.Notify(615961)
		exprCopy := *expr
		exprCopy.Expr = e
		return &exprCopy
	} else {
		__antithesis_instrumentation__.Notify(615962)
	}
	__antithesis_instrumentation__.Notify(615960)
	return expr
}

func (expr *BinaryExpr) Walk(v Visitor) Expr {
	__antithesis_instrumentation__.Notify(615963)
	left, changedL := WalkExpr(v, expr.Left)
	right, changedR := WalkExpr(v, expr.Right)
	if changedL || func() bool {
		__antithesis_instrumentation__.Notify(615965)
		return changedR == true
	}() == true {
		__antithesis_instrumentation__.Notify(615966)
		exprCopy := *expr
		exprCopy.Left = left
		exprCopy.Right = right
		return &exprCopy
	} else {
		__antithesis_instrumentation__.Notify(615967)
	}
	__antithesis_instrumentation__.Notify(615964)
	return expr
}

func (expr *CaseExpr) copyNode() *CaseExpr {
	__antithesis_instrumentation__.Notify(615968)
	exprCopy := *expr

	exprCopy.Whens = make([]*When, len(expr.Whens))
	for i, w := range expr.Whens {
		__antithesis_instrumentation__.Notify(615970)
		wCopy := *w
		exprCopy.Whens[i] = &wCopy
	}
	__antithesis_instrumentation__.Notify(615969)
	return &exprCopy
}

func (expr *CaseExpr) Walk(v Visitor) Expr {
	__antithesis_instrumentation__.Notify(615971)
	ret := expr

	if expr.Expr != nil {
		__antithesis_instrumentation__.Notify(615975)
		e, changed := WalkExpr(v, expr.Expr)
		if changed {
			__antithesis_instrumentation__.Notify(615976)
			ret = expr.copyNode()
			ret.Expr = e
		} else {
			__antithesis_instrumentation__.Notify(615977)
		}
	} else {
		__antithesis_instrumentation__.Notify(615978)
	}
	__antithesis_instrumentation__.Notify(615972)
	for i, w := range expr.Whens {
		__antithesis_instrumentation__.Notify(615979)
		cond, changedC := WalkExpr(v, w.Cond)
		val, changedV := WalkExpr(v, w.Val)
		if changedC || func() bool {
			__antithesis_instrumentation__.Notify(615980)
			return changedV == true
		}() == true {
			__antithesis_instrumentation__.Notify(615981)
			if ret == expr {
				__antithesis_instrumentation__.Notify(615983)
				ret = expr.copyNode()
			} else {
				__antithesis_instrumentation__.Notify(615984)
			}
			__antithesis_instrumentation__.Notify(615982)
			ret.Whens[i].Cond = cond
			ret.Whens[i].Val = val
		} else {
			__antithesis_instrumentation__.Notify(615985)
		}
	}
	__antithesis_instrumentation__.Notify(615973)
	if expr.Else != nil {
		__antithesis_instrumentation__.Notify(615986)
		e, changed := WalkExpr(v, expr.Else)
		if changed {
			__antithesis_instrumentation__.Notify(615987)
			if ret == expr {
				__antithesis_instrumentation__.Notify(615989)
				ret = expr.copyNode()
			} else {
				__antithesis_instrumentation__.Notify(615990)
			}
			__antithesis_instrumentation__.Notify(615988)
			ret.Else = e
		} else {
			__antithesis_instrumentation__.Notify(615991)
		}
	} else {
		__antithesis_instrumentation__.Notify(615992)
	}
	__antithesis_instrumentation__.Notify(615974)
	return ret
}

func (expr *CastExpr) Walk(v Visitor) Expr {
	__antithesis_instrumentation__.Notify(615993)
	e, changed := WalkExpr(v, expr.Expr)
	if changed {
		__antithesis_instrumentation__.Notify(615995)
		exprCopy := *expr
		exprCopy.Expr = e
		return &exprCopy
	} else {
		__antithesis_instrumentation__.Notify(615996)
	}
	__antithesis_instrumentation__.Notify(615994)
	return expr
}

func (expr *CollateExpr) Walk(v Visitor) Expr {
	__antithesis_instrumentation__.Notify(615997)
	e, changed := WalkExpr(v, expr.Expr)
	if changed {
		__antithesis_instrumentation__.Notify(615999)
		exprCopy := *expr
		exprCopy.Expr = e
		return &exprCopy
	} else {
		__antithesis_instrumentation__.Notify(616000)
	}
	__antithesis_instrumentation__.Notify(615998)
	return expr
}

func (expr *ColumnAccessExpr) Walk(v Visitor) Expr {
	__antithesis_instrumentation__.Notify(616001)
	e, changed := WalkExpr(v, expr.Expr)
	if changed {
		__antithesis_instrumentation__.Notify(616003)
		exprCopy := *expr
		exprCopy.Expr = e
		return &exprCopy
	} else {
		__antithesis_instrumentation__.Notify(616004)
	}
	__antithesis_instrumentation__.Notify(616002)
	return expr
}

func (expr *TupleStar) Walk(v Visitor) Expr {
	__antithesis_instrumentation__.Notify(616005)
	e, changed := WalkExpr(v, expr.Expr)
	if changed {
		__antithesis_instrumentation__.Notify(616007)
		exprCopy := *expr
		exprCopy.Expr = e
		return &exprCopy
	} else {
		__antithesis_instrumentation__.Notify(616008)
	}
	__antithesis_instrumentation__.Notify(616006)
	return expr
}

func (expr *CoalesceExpr) copyNode() *CoalesceExpr {
	__antithesis_instrumentation__.Notify(616009)
	exprCopy := *expr
	return &exprCopy
}

func (expr *CoalesceExpr) Walk(v Visitor) Expr {
	__antithesis_instrumentation__.Notify(616010)
	ret := expr
	exprs, changed := walkExprSlice(v, expr.Exprs)
	if changed {
		__antithesis_instrumentation__.Notify(616012)
		if ret == expr {
			__antithesis_instrumentation__.Notify(616014)
			ret = expr.copyNode()
		} else {
			__antithesis_instrumentation__.Notify(616015)
		}
		__antithesis_instrumentation__.Notify(616013)
		ret.Exprs = exprs
	} else {
		__antithesis_instrumentation__.Notify(616016)
	}
	__antithesis_instrumentation__.Notify(616011)
	return ret
}

func (expr *ComparisonExpr) Walk(v Visitor) Expr {
	__antithesis_instrumentation__.Notify(616017)
	left, changedL := WalkExpr(v, expr.Left)
	right, changedR := WalkExpr(v, expr.Right)
	if changedL || func() bool {
		__antithesis_instrumentation__.Notify(616019)
		return changedR == true
	}() == true {
		__antithesis_instrumentation__.Notify(616020)
		exprCopy := *expr
		exprCopy.Left = left
		exprCopy.Right = right
		return &exprCopy
	} else {
		__antithesis_instrumentation__.Notify(616021)
	}
	__antithesis_instrumentation__.Notify(616018)
	return expr
}

func (expr *FuncExpr) copyNode() *FuncExpr {
	__antithesis_instrumentation__.Notify(616022)
	exprCopy := *expr
	exprCopy.Exprs = append(Exprs(nil), exprCopy.Exprs...)
	return &exprCopy
}

func (node *WindowFrame) copyNode() *WindowFrame {
	__antithesis_instrumentation__.Notify(616023)
	nodeCopy := *node
	return &nodeCopy
}

func walkWindowFrame(v Visitor, frame *WindowFrame) (*WindowFrame, bool) {
	__antithesis_instrumentation__.Notify(616024)
	ret := frame
	if frame.Bounds.StartBound != nil {
		__antithesis_instrumentation__.Notify(616027)
		b, changed := walkWindowFrameBound(v, frame.Bounds.StartBound)
		if changed {
			__antithesis_instrumentation__.Notify(616028)
			if ret == frame {
				__antithesis_instrumentation__.Notify(616030)
				ret = frame.copyNode()
			} else {
				__antithesis_instrumentation__.Notify(616031)
			}
			__antithesis_instrumentation__.Notify(616029)
			ret.Bounds.StartBound = b
		} else {
			__antithesis_instrumentation__.Notify(616032)
		}
	} else {
		__antithesis_instrumentation__.Notify(616033)
	}
	__antithesis_instrumentation__.Notify(616025)
	if frame.Bounds.EndBound != nil {
		__antithesis_instrumentation__.Notify(616034)
		b, changed := walkWindowFrameBound(v, frame.Bounds.EndBound)
		if changed {
			__antithesis_instrumentation__.Notify(616035)
			if ret == frame {
				__antithesis_instrumentation__.Notify(616037)
				ret = frame.copyNode()
			} else {
				__antithesis_instrumentation__.Notify(616038)
			}
			__antithesis_instrumentation__.Notify(616036)
			ret.Bounds.EndBound = b
		} else {
			__antithesis_instrumentation__.Notify(616039)
		}
	} else {
		__antithesis_instrumentation__.Notify(616040)
	}
	__antithesis_instrumentation__.Notify(616026)
	return ret, ret != frame
}

func (node *WindowFrameBound) copyNode() *WindowFrameBound {
	__antithesis_instrumentation__.Notify(616041)
	nodeCopy := *node
	return &nodeCopy
}

func walkWindowFrameBound(v Visitor, bound *WindowFrameBound) (*WindowFrameBound, bool) {
	__antithesis_instrumentation__.Notify(616042)
	ret := bound
	if bound.HasOffset() {
		__antithesis_instrumentation__.Notify(616044)
		e, changed := WalkExpr(v, bound.OffsetExpr)
		if changed {
			__antithesis_instrumentation__.Notify(616045)
			if ret == bound {
				__antithesis_instrumentation__.Notify(616047)
				ret = bound.copyNode()
			} else {
				__antithesis_instrumentation__.Notify(616048)
			}
			__antithesis_instrumentation__.Notify(616046)
			ret.OffsetExpr = e
		} else {
			__antithesis_instrumentation__.Notify(616049)
		}
	} else {
		__antithesis_instrumentation__.Notify(616050)
	}
	__antithesis_instrumentation__.Notify(616043)
	return ret, ret != bound
}

func (node *WindowDef) copyNode() *WindowDef {
	__antithesis_instrumentation__.Notify(616051)
	nodeCopy := *node
	return &nodeCopy
}

func walkWindowDef(v Visitor, windowDef *WindowDef) (*WindowDef, bool) {
	__antithesis_instrumentation__.Notify(616052)
	ret := windowDef
	if len(windowDef.Partitions) > 0 {
		__antithesis_instrumentation__.Notify(616056)
		exprs, changed := walkExprSlice(v, windowDef.Partitions)
		if changed {
			__antithesis_instrumentation__.Notify(616057)
			if ret == windowDef {
				__antithesis_instrumentation__.Notify(616059)
				ret = windowDef.copyNode()
			} else {
				__antithesis_instrumentation__.Notify(616060)
			}
			__antithesis_instrumentation__.Notify(616058)
			ret.Partitions = exprs
		} else {
			__antithesis_instrumentation__.Notify(616061)
		}
	} else {
		__antithesis_instrumentation__.Notify(616062)
	}
	__antithesis_instrumentation__.Notify(616053)
	if len(windowDef.OrderBy) > 0 {
		__antithesis_instrumentation__.Notify(616063)
		order, changed := walkOrderBy(v, windowDef.OrderBy)
		if changed {
			__antithesis_instrumentation__.Notify(616064)
			if ret == windowDef {
				__antithesis_instrumentation__.Notify(616066)
				ret = windowDef.copyNode()
			} else {
				__antithesis_instrumentation__.Notify(616067)
			}
			__antithesis_instrumentation__.Notify(616065)
			ret.OrderBy = order
		} else {
			__antithesis_instrumentation__.Notify(616068)
		}
	} else {
		__antithesis_instrumentation__.Notify(616069)
	}
	__antithesis_instrumentation__.Notify(616054)
	if windowDef.Frame != nil {
		__antithesis_instrumentation__.Notify(616070)
		frame, changed := walkWindowFrame(v, windowDef.Frame)
		if changed {
			__antithesis_instrumentation__.Notify(616071)
			if ret == windowDef {
				__antithesis_instrumentation__.Notify(616073)
				ret = windowDef.copyNode()
			} else {
				__antithesis_instrumentation__.Notify(616074)
			}
			__antithesis_instrumentation__.Notify(616072)
			ret.Frame = frame
		} else {
			__antithesis_instrumentation__.Notify(616075)
		}
	} else {
		__antithesis_instrumentation__.Notify(616076)
	}
	__antithesis_instrumentation__.Notify(616055)

	return ret, ret != windowDef
}

func (expr *FuncExpr) Walk(v Visitor) Expr {
	__antithesis_instrumentation__.Notify(616077)
	ret := expr
	exprs, changed := walkExprSlice(v, expr.Exprs)
	if changed {
		__antithesis_instrumentation__.Notify(616081)
		if ret == expr {
			__antithesis_instrumentation__.Notify(616083)
			ret = expr.copyNode()
		} else {
			__antithesis_instrumentation__.Notify(616084)
		}
		__antithesis_instrumentation__.Notify(616082)
		ret.Exprs = exprs
	} else {
		__antithesis_instrumentation__.Notify(616085)
	}
	__antithesis_instrumentation__.Notify(616078)
	if expr.Filter != nil {
		__antithesis_instrumentation__.Notify(616086)
		e, changed := WalkExpr(v, expr.Filter)
		if changed {
			__antithesis_instrumentation__.Notify(616087)
			if ret == expr {
				__antithesis_instrumentation__.Notify(616089)
				ret = expr.copyNode()
			} else {
				__antithesis_instrumentation__.Notify(616090)
			}
			__antithesis_instrumentation__.Notify(616088)
			ret.Filter = e
		} else {
			__antithesis_instrumentation__.Notify(616091)
		}
	} else {
		__antithesis_instrumentation__.Notify(616092)
	}
	__antithesis_instrumentation__.Notify(616079)

	if expr.OrderBy != nil {
		__antithesis_instrumentation__.Notify(616093)
		order, changed := walkOrderBy(v, expr.OrderBy)
		if changed {
			__antithesis_instrumentation__.Notify(616094)
			if ret == expr {
				__antithesis_instrumentation__.Notify(616096)
				ret = expr.copyNode()
			} else {
				__antithesis_instrumentation__.Notify(616097)
			}
			__antithesis_instrumentation__.Notify(616095)
			ret.OrderBy = order
		} else {
			__antithesis_instrumentation__.Notify(616098)
		}
	} else {
		__antithesis_instrumentation__.Notify(616099)
	}
	__antithesis_instrumentation__.Notify(616080)
	return ret
}

func (expr *IfExpr) Walk(v Visitor) Expr {
	__antithesis_instrumentation__.Notify(616100)
	c, changedC := WalkExpr(v, expr.Cond)
	t, changedT := WalkExpr(v, expr.True)
	e, changedE := WalkExpr(v, expr.Else)
	if changedC || func() bool {
		__antithesis_instrumentation__.Notify(616102)
		return changedT == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(616103)
		return changedE == true
	}() == true {
		__antithesis_instrumentation__.Notify(616104)
		exprCopy := *expr
		exprCopy.Cond = c
		exprCopy.True = t
		exprCopy.Else = e
		return &exprCopy
	} else {
		__antithesis_instrumentation__.Notify(616105)
	}
	__antithesis_instrumentation__.Notify(616101)
	return expr
}

func (expr *IfErrExpr) Walk(v Visitor) Expr {
	__antithesis_instrumentation__.Notify(616106)
	c, changedC := WalkExpr(v, expr.Cond)
	t := expr.ErrCode
	changedEC := false
	if t != nil {
		__antithesis_instrumentation__.Notify(616110)
		t, changedEC = WalkExpr(v, expr.ErrCode)
	} else {
		__antithesis_instrumentation__.Notify(616111)
	}
	__antithesis_instrumentation__.Notify(616107)
	e := expr.Else
	changedE := false
	if e != nil {
		__antithesis_instrumentation__.Notify(616112)
		e, changedE = WalkExpr(v, expr.Else)
	} else {
		__antithesis_instrumentation__.Notify(616113)
	}
	__antithesis_instrumentation__.Notify(616108)
	if changedC || func() bool {
		__antithesis_instrumentation__.Notify(616114)
		return changedEC == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(616115)
		return changedE == true
	}() == true {
		__antithesis_instrumentation__.Notify(616116)
		exprCopy := *expr
		exprCopy.Cond = c
		exprCopy.ErrCode = t
		exprCopy.Else = e
		return &exprCopy
	} else {
		__antithesis_instrumentation__.Notify(616117)
	}
	__antithesis_instrumentation__.Notify(616109)
	return expr
}

func (expr *IndirectionExpr) copyNode() *IndirectionExpr {
	__antithesis_instrumentation__.Notify(616118)
	exprCopy := *expr
	exprCopy.Indirection = append(ArraySubscripts(nil), exprCopy.Indirection...)
	for i, t := range exprCopy.Indirection {
		__antithesis_instrumentation__.Notify(616120)
		subscriptCopy := *t
		exprCopy.Indirection[i] = &subscriptCopy
	}
	__antithesis_instrumentation__.Notify(616119)
	return &exprCopy
}

func (expr *IndirectionExpr) Walk(v Visitor) Expr {
	__antithesis_instrumentation__.Notify(616121)
	ret := expr

	e, changed := WalkExpr(v, expr.Expr)
	if changed {
		__antithesis_instrumentation__.Notify(616124)
		if ret == expr {
			__antithesis_instrumentation__.Notify(616126)
			ret = expr.copyNode()
		} else {
			__antithesis_instrumentation__.Notify(616127)
		}
		__antithesis_instrumentation__.Notify(616125)
		ret.Expr = e
	} else {
		__antithesis_instrumentation__.Notify(616128)
	}
	__antithesis_instrumentation__.Notify(616122)

	for i, t := range expr.Indirection {
		__antithesis_instrumentation__.Notify(616129)
		if t.Begin != nil {
			__antithesis_instrumentation__.Notify(616131)
			e, changed := WalkExpr(v, t.Begin)
			if changed {
				__antithesis_instrumentation__.Notify(616132)
				if ret == expr {
					__antithesis_instrumentation__.Notify(616134)
					ret = expr.copyNode()
				} else {
					__antithesis_instrumentation__.Notify(616135)
				}
				__antithesis_instrumentation__.Notify(616133)
				ret.Indirection[i].Begin = e
			} else {
				__antithesis_instrumentation__.Notify(616136)
			}
		} else {
			__antithesis_instrumentation__.Notify(616137)
		}
		__antithesis_instrumentation__.Notify(616130)
		if t.End != nil {
			__antithesis_instrumentation__.Notify(616138)
			e, changed := WalkExpr(v, t.End)
			if changed {
				__antithesis_instrumentation__.Notify(616139)
				if ret == expr {
					__antithesis_instrumentation__.Notify(616141)
					ret = expr.copyNode()
				} else {
					__antithesis_instrumentation__.Notify(616142)
				}
				__antithesis_instrumentation__.Notify(616140)
				ret.Indirection[i].End = e
			} else {
				__antithesis_instrumentation__.Notify(616143)
			}
		} else {
			__antithesis_instrumentation__.Notify(616144)
		}
	}
	__antithesis_instrumentation__.Notify(616123)

	return ret
}

func (expr *IsOfTypeExpr) Walk(v Visitor) Expr {
	__antithesis_instrumentation__.Notify(616145)
	e, changed := WalkExpr(v, expr.Expr)
	if changed {
		__antithesis_instrumentation__.Notify(616147)
		exprCopy := *expr
		exprCopy.Expr = e
		return &exprCopy
	} else {
		__antithesis_instrumentation__.Notify(616148)
	}
	__antithesis_instrumentation__.Notify(616146)
	return expr
}

func (expr *NotExpr) Walk(v Visitor) Expr {
	__antithesis_instrumentation__.Notify(616149)
	e, changed := WalkExpr(v, expr.Expr)
	if changed {
		__antithesis_instrumentation__.Notify(616151)
		exprCopy := *expr
		exprCopy.Expr = e
		return &exprCopy
	} else {
		__antithesis_instrumentation__.Notify(616152)
	}
	__antithesis_instrumentation__.Notify(616150)
	return expr
}

func (expr *IsNullExpr) Walk(v Visitor) Expr {
	__antithesis_instrumentation__.Notify(616153)
	e, changed := WalkExpr(v, expr.Expr)
	if changed {
		__antithesis_instrumentation__.Notify(616155)
		exprCopy := *expr
		exprCopy.Expr = e
		return &exprCopy
	} else {
		__antithesis_instrumentation__.Notify(616156)
	}
	__antithesis_instrumentation__.Notify(616154)
	return expr
}

func (expr *IsNotNullExpr) Walk(v Visitor) Expr {
	__antithesis_instrumentation__.Notify(616157)
	e, changed := WalkExpr(v, expr.Expr)
	if changed {
		__antithesis_instrumentation__.Notify(616159)
		exprCopy := *expr
		exprCopy.Expr = e
		return &exprCopy
	} else {
		__antithesis_instrumentation__.Notify(616160)
	}
	__antithesis_instrumentation__.Notify(616158)
	return expr
}

func (expr *NullIfExpr) Walk(v Visitor) Expr {
	__antithesis_instrumentation__.Notify(616161)
	e1, changed1 := WalkExpr(v, expr.Expr1)
	e2, changed2 := WalkExpr(v, expr.Expr2)
	if changed1 || func() bool {
		__antithesis_instrumentation__.Notify(616163)
		return changed2 == true
	}() == true {
		__antithesis_instrumentation__.Notify(616164)
		exprCopy := *expr
		exprCopy.Expr1 = e1
		exprCopy.Expr2 = e2
		return &exprCopy
	} else {
		__antithesis_instrumentation__.Notify(616165)
	}
	__antithesis_instrumentation__.Notify(616162)
	return expr
}

func (expr *OrExpr) Walk(v Visitor) Expr {
	__antithesis_instrumentation__.Notify(616166)
	left, changedL := WalkExpr(v, expr.Left)
	right, changedR := WalkExpr(v, expr.Right)
	if changedL || func() bool {
		__antithesis_instrumentation__.Notify(616168)
		return changedR == true
	}() == true {
		__antithesis_instrumentation__.Notify(616169)
		exprCopy := *expr
		exprCopy.Left = left
		exprCopy.Right = right
		return &exprCopy
	} else {
		__antithesis_instrumentation__.Notify(616170)
	}
	__antithesis_instrumentation__.Notify(616167)
	return expr
}

func (expr *ParenExpr) Walk(v Visitor) Expr {
	__antithesis_instrumentation__.Notify(616171)
	e, changed := WalkExpr(v, expr.Expr)
	if changed {
		__antithesis_instrumentation__.Notify(616173)
		exprCopy := *expr
		exprCopy.Expr = e
		return &exprCopy
	} else {
		__antithesis_instrumentation__.Notify(616174)
	}
	__antithesis_instrumentation__.Notify(616172)
	return expr
}

func (expr *RangeCond) Walk(v Visitor) Expr {
	__antithesis_instrumentation__.Notify(616175)
	l, changedL := WalkExpr(v, expr.Left)
	f, changedF := WalkExpr(v, expr.From)
	t, changedT := WalkExpr(v, expr.To)
	if changedL || func() bool {
		__antithesis_instrumentation__.Notify(616177)
		return changedF == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(616178)
		return changedT == true
	}() == true {
		__antithesis_instrumentation__.Notify(616179)
		exprCopy := *expr
		exprCopy.Left = l
		exprCopy.From = f
		exprCopy.To = t
		return &exprCopy
	} else {
		__antithesis_instrumentation__.Notify(616180)
	}
	__antithesis_instrumentation__.Notify(616176)
	return expr
}

func (expr *Subquery) Walk(v Visitor) Expr {
	__antithesis_instrumentation__.Notify(616181)
	sel, changed := walkStmt(v, expr.Select)
	if changed {
		__antithesis_instrumentation__.Notify(616183)
		exprCopy := *expr
		exprCopy.Select = sel.(SelectStatement)
		return &exprCopy
	} else {
		__antithesis_instrumentation__.Notify(616184)
	}
	__antithesis_instrumentation__.Notify(616182)
	return expr
}

func (expr *Subquery) WalkTableExpr(v Visitor) TableExpr {
	__antithesis_instrumentation__.Notify(616185)
	sel, changed := walkStmt(v, expr.Select)
	if changed {
		__antithesis_instrumentation__.Notify(616187)
		exprCopy := *expr
		exprCopy.Select = sel.(SelectStatement)
		return &exprCopy
	} else {
		__antithesis_instrumentation__.Notify(616188)
	}
	__antithesis_instrumentation__.Notify(616186)
	return expr
}

func (expr *AliasedTableExpr) WalkTableExpr(v Visitor) TableExpr {
	__antithesis_instrumentation__.Notify(616189)
	newExpr, changed := walkTableExpr(v, expr.Expr)
	if changed {
		__antithesis_instrumentation__.Notify(616191)
		exprCopy := *expr
		exprCopy.Expr = newExpr
		return &exprCopy
	} else {
		__antithesis_instrumentation__.Notify(616192)
	}
	__antithesis_instrumentation__.Notify(616190)
	return expr
}

func (expr *ParenTableExpr) WalkTableExpr(v Visitor) TableExpr {
	__antithesis_instrumentation__.Notify(616193)
	newExpr, changed := walkTableExpr(v, expr.Expr)
	if changed {
		__antithesis_instrumentation__.Notify(616195)
		exprCopy := *expr
		exprCopy.Expr = newExpr
		return &exprCopy
	} else {
		__antithesis_instrumentation__.Notify(616196)
	}
	__antithesis_instrumentation__.Notify(616194)
	return expr
}

func (expr *JoinTableExpr) WalkTableExpr(v Visitor) TableExpr {
	__antithesis_instrumentation__.Notify(616197)
	left, changedL := walkTableExpr(v, expr.Left)
	right, changedR := walkTableExpr(v, expr.Right)
	if changedL || func() bool {
		__antithesis_instrumentation__.Notify(616199)
		return changedR == true
	}() == true {
		__antithesis_instrumentation__.Notify(616200)
		exprCopy := *expr
		exprCopy.Left = left
		exprCopy.Right = right
		return &exprCopy
	} else {
		__antithesis_instrumentation__.Notify(616201)
	}
	__antithesis_instrumentation__.Notify(616198)
	return expr
}

func (expr *RowsFromExpr) copyNode() *RowsFromExpr {
	__antithesis_instrumentation__.Notify(616202)
	exprCopy := *expr
	exprCopy.Items = append(Exprs(nil), exprCopy.Items...)
	return &exprCopy
}

func (expr *RowsFromExpr) WalkTableExpr(v Visitor) TableExpr {
	__antithesis_instrumentation__.Notify(616203)
	ret := expr
	for i := range expr.Items {
		__antithesis_instrumentation__.Notify(616205)
		e, changed := WalkExpr(v, expr.Items[i])
		if changed {
			__antithesis_instrumentation__.Notify(616206)
			if ret == expr {
				__antithesis_instrumentation__.Notify(616208)
				ret = expr.copyNode()
			} else {
				__antithesis_instrumentation__.Notify(616209)
			}
			__antithesis_instrumentation__.Notify(616207)
			ret.Items[i] = e
		} else {
			__antithesis_instrumentation__.Notify(616210)
		}
	}
	__antithesis_instrumentation__.Notify(616204)
	return ret
}

func (expr *StatementSource) WalkTableExpr(v Visitor) TableExpr {
	__antithesis_instrumentation__.Notify(616211)
	s, changed := walkStmt(v, expr.Statement)
	if changed {
		__antithesis_instrumentation__.Notify(616213)
		exprCopy := *expr
		exprCopy.Statement = s
		return &exprCopy
	} else {
		__antithesis_instrumentation__.Notify(616214)
	}
	__antithesis_instrumentation__.Notify(616212)
	return expr
}

func (expr *TableName) WalkTableExpr(_ Visitor) TableExpr {
	__antithesis_instrumentation__.Notify(616215)
	return expr
}

func (expr *TableRef) WalkTableExpr(_ Visitor) TableExpr {
	__antithesis_instrumentation__.Notify(616216)
	return expr
}

func (expr *UnresolvedObjectName) WalkTableExpr(_ Visitor) TableExpr {
	__antithesis_instrumentation__.Notify(616217)
	return expr
}

func (expr *UnaryExpr) Walk(v Visitor) Expr {
	__antithesis_instrumentation__.Notify(616218)
	e, changed := WalkExpr(v, expr.Expr)
	if changed {
		__antithesis_instrumentation__.Notify(616220)
		exprCopy := *expr
		exprCopy.Expr = e
		return &exprCopy
	} else {
		__antithesis_instrumentation__.Notify(616221)
	}
	__antithesis_instrumentation__.Notify(616219)
	return expr
}

func walkExprSlice(v Visitor, slice []Expr) ([]Expr, bool) {
	__antithesis_instrumentation__.Notify(616222)
	copied := false
	for i := range slice {
		__antithesis_instrumentation__.Notify(616224)
		e, changed := WalkExpr(v, slice[i])
		if changed {
			__antithesis_instrumentation__.Notify(616225)
			if !copied {
				__antithesis_instrumentation__.Notify(616227)
				slice = append([]Expr(nil), slice...)
				copied = true
			} else {
				__antithesis_instrumentation__.Notify(616228)
			}
			__antithesis_instrumentation__.Notify(616226)
			slice[i] = e
		} else {
			__antithesis_instrumentation__.Notify(616229)
		}
	}
	__antithesis_instrumentation__.Notify(616223)
	return slice, copied
}

func walkKVOptions(v Visitor, opts KVOptions) (KVOptions, bool) {
	__antithesis_instrumentation__.Notify(616230)
	copied := false
	for i := range opts {
		__antithesis_instrumentation__.Notify(616232)
		if opts[i].Value == nil {
			__antithesis_instrumentation__.Notify(616234)
			continue
		} else {
			__antithesis_instrumentation__.Notify(616235)
		}
		__antithesis_instrumentation__.Notify(616233)
		e, changed := WalkExpr(v, opts[i].Value)
		if changed {
			__antithesis_instrumentation__.Notify(616236)
			if !copied {
				__antithesis_instrumentation__.Notify(616238)
				opts = append(KVOptions(nil), opts...)
				copied = true
			} else {
				__antithesis_instrumentation__.Notify(616239)
			}
			__antithesis_instrumentation__.Notify(616237)
			opts[i].Value = e
		} else {
			__antithesis_instrumentation__.Notify(616240)
		}
	}
	__antithesis_instrumentation__.Notify(616231)
	return opts, copied
}

func (expr *Tuple) Walk(v Visitor) Expr {
	__antithesis_instrumentation__.Notify(616241)
	exprs, changed := walkExprSlice(v, expr.Exprs)
	if changed {
		__antithesis_instrumentation__.Notify(616243)
		exprCopy := *expr
		exprCopy.Exprs = exprs
		return &exprCopy
	} else {
		__antithesis_instrumentation__.Notify(616244)
	}
	__antithesis_instrumentation__.Notify(616242)
	return expr
}

func (expr *Array) Walk(v Visitor) Expr {
	__antithesis_instrumentation__.Notify(616245)
	if exprs, changed := walkExprSlice(v, expr.Exprs); changed {
		__antithesis_instrumentation__.Notify(616247)
		exprCopy := *expr
		exprCopy.Exprs = exprs
		return &exprCopy
	} else {
		__antithesis_instrumentation__.Notify(616248)
	}
	__antithesis_instrumentation__.Notify(616246)
	return expr
}

func (expr *DVoid) Walk(_ Visitor) Expr { __antithesis_instrumentation__.Notify(616249); return expr }

func (expr *ArrayFlatten) Walk(v Visitor) Expr {
	__antithesis_instrumentation__.Notify(616250)
	if sq, changed := WalkExpr(v, expr.Subquery); changed {
		__antithesis_instrumentation__.Notify(616252)
		exprCopy := *expr
		exprCopy.Subquery = sq
		return &exprCopy
	} else {
		__antithesis_instrumentation__.Notify(616253)
	}
	__antithesis_instrumentation__.Notify(616251)
	return expr
}

func (expr UnqualifiedStar) Walk(_ Visitor) Expr {
	__antithesis_instrumentation__.Notify(616254)
	return expr
}

func (expr *UnresolvedName) Walk(_ Visitor) Expr {
	__antithesis_instrumentation__.Notify(616255)
	return expr
}

func (expr *AllColumnsSelector) Walk(_ Visitor) Expr {
	__antithesis_instrumentation__.Notify(616256)
	return expr
}

func (expr *ColumnItem) Walk(_ Visitor) Expr {
	__antithesis_instrumentation__.Notify(616257)

	return expr
}

func (expr DefaultVal) Walk(_ Visitor) Expr {
	__antithesis_instrumentation__.Notify(616258)
	return expr
}

func (expr PartitionMaxVal) Walk(_ Visitor) Expr {
	__antithesis_instrumentation__.Notify(616259)
	return expr
}

func (expr PartitionMinVal) Walk(_ Visitor) Expr {
	__antithesis_instrumentation__.Notify(616260)
	return expr
}

func (expr *NumVal) Walk(_ Visitor) Expr { __antithesis_instrumentation__.Notify(616261); return expr }

func (expr *StrVal) Walk(_ Visitor) Expr { __antithesis_instrumentation__.Notify(616262); return expr }

func (expr *Placeholder) Walk(_ Visitor) Expr {
	__antithesis_instrumentation__.Notify(616263)
	return expr
}

func (expr *DBitArray) Walk(_ Visitor) Expr {
	__antithesis_instrumentation__.Notify(616264)
	return expr
}

func (expr *DBool) Walk(_ Visitor) Expr { __antithesis_instrumentation__.Notify(616265); return expr }

func (expr *DBytes) Walk(_ Visitor) Expr { __antithesis_instrumentation__.Notify(616266); return expr }

func (expr *DDate) Walk(_ Visitor) Expr { __antithesis_instrumentation__.Notify(616267); return expr }

func (expr *DTime) Walk(_ Visitor) Expr { __antithesis_instrumentation__.Notify(616268); return expr }

func (expr *DTimeTZ) Walk(_ Visitor) Expr { __antithesis_instrumentation__.Notify(616269); return expr }

func (expr *DFloat) Walk(_ Visitor) Expr { __antithesis_instrumentation__.Notify(616270); return expr }

func (expr *DEnum) Walk(_ Visitor) Expr { __antithesis_instrumentation__.Notify(616271); return expr }

func (expr *DDecimal) Walk(_ Visitor) Expr {
	__antithesis_instrumentation__.Notify(616272)
	return expr
}

func (expr *DInt) Walk(_ Visitor) Expr { __antithesis_instrumentation__.Notify(616273); return expr }

func (expr *DInterval) Walk(_ Visitor) Expr {
	__antithesis_instrumentation__.Notify(616274)
	return expr
}

func (expr *DBox2D) Walk(_ Visitor) Expr { __antithesis_instrumentation__.Notify(616275); return expr }

func (expr *DGeography) Walk(_ Visitor) Expr {
	__antithesis_instrumentation__.Notify(616276)
	return expr
}

func (expr *DGeometry) Walk(_ Visitor) Expr {
	__antithesis_instrumentation__.Notify(616277)
	return expr
}

func (expr *DJSON) Walk(_ Visitor) Expr { __antithesis_instrumentation__.Notify(616278); return expr }

func (expr *DUuid) Walk(_ Visitor) Expr { __antithesis_instrumentation__.Notify(616279); return expr }

func (expr *DIPAddr) Walk(_ Visitor) Expr { __antithesis_instrumentation__.Notify(616280); return expr }

func (expr dNull) Walk(_ Visitor) Expr { __antithesis_instrumentation__.Notify(616281); return expr }

func (expr *DString) Walk(_ Visitor) Expr { __antithesis_instrumentation__.Notify(616282); return expr }

func (expr *DCollatedString) Walk(_ Visitor) Expr {
	__antithesis_instrumentation__.Notify(616283)
	return expr
}

func (expr *DTimestamp) Walk(_ Visitor) Expr {
	__antithesis_instrumentation__.Notify(616284)
	return expr
}

func (expr *DTimestampTZ) Walk(_ Visitor) Expr {
	__antithesis_instrumentation__.Notify(616285)
	return expr
}

func (expr *DTuple) Walk(v Visitor) Expr {
	__antithesis_instrumentation__.Notify(616286)
	for _, d := range expr.D {
		__antithesis_instrumentation__.Notify(616288)

		d.Walk(v)
	}
	__antithesis_instrumentation__.Notify(616287)
	return expr
}

func (expr *DArray) Walk(v Visitor) Expr {
	__antithesis_instrumentation__.Notify(616289)
	for _, d := range expr.Array {
		__antithesis_instrumentation__.Notify(616291)

		d.Walk(v)
	}
	__antithesis_instrumentation__.Notify(616290)
	return expr
}

func (expr *DOid) Walk(_ Visitor) Expr { __antithesis_instrumentation__.Notify(616292); return expr }

func (expr *DOidWrapper) Walk(_ Visitor) Expr {
	__antithesis_instrumentation__.Notify(616293)
	return expr
}

func WalkExpr(v Visitor, expr Expr) (newExpr Expr, changed bool) {
	__antithesis_instrumentation__.Notify(616294)
	recurse, newExpr := v.VisitPre(expr)

	if recurse {
		__antithesis_instrumentation__.Notify(616296)
		newExpr = newExpr.Walk(v)
		newExpr = v.VisitPost(newExpr)
	} else {
		__antithesis_instrumentation__.Notify(616297)
	}
	__antithesis_instrumentation__.Notify(616295)

	return newExpr, (reflect.ValueOf(expr) != reflect.ValueOf(newExpr))
}

func walkTableExpr(v Visitor, expr TableExpr) (newExpr TableExpr, changed bool) {
	__antithesis_instrumentation__.Notify(616298)
	newExpr = expr.WalkTableExpr(v)
	return newExpr, (reflect.ValueOf(expr) != reflect.ValueOf(newExpr))
}

func WalkExprConst(v Visitor, expr Expr) {
	__antithesis_instrumentation__.Notify(616299)
	WalkExpr(v, expr)

}

type walkableStmt interface {
	Statement
	walkStmt(Visitor) Statement
}

func walkReturningClause(v Visitor, clause ReturningClause) (ReturningClause, bool) {
	__antithesis_instrumentation__.Notify(616300)
	switch t := clause.(type) {
	case *ReturningExprs:
		__antithesis_instrumentation__.Notify(616301)
		ret := t
		for i, expr := range *t {
			__antithesis_instrumentation__.Notify(616305)
			e, changed := WalkExpr(v, expr.Expr)
			if changed {
				__antithesis_instrumentation__.Notify(616306)
				if ret == t {
					__antithesis_instrumentation__.Notify(616308)
					ret = t.copyNode()
				} else {
					__antithesis_instrumentation__.Notify(616309)
				}
				__antithesis_instrumentation__.Notify(616307)
				(*ret)[i].Expr = e
			} else {
				__antithesis_instrumentation__.Notify(616310)
			}
		}
		__antithesis_instrumentation__.Notify(616302)
		return ret, (ret != t)
	case *ReturningNothing, *NoReturningClause:
		__antithesis_instrumentation__.Notify(616303)
		return t, false
	default:
		__antithesis_instrumentation__.Notify(616304)
		panic(errors.AssertionFailedf("unexpected ReturningClause type: %T", t))
	}
}

func (n *ShowTenantClusterSetting) copyNode() *ShowTenantClusterSetting {
	__antithesis_instrumentation__.Notify(616311)
	stmtCopy := *n
	return &stmtCopy
}

func (n *ShowTenantClusterSetting) walkStmt(v Visitor) Statement {
	__antithesis_instrumentation__.Notify(616312)
	ret := n
	sc, changed := walkStmt(v, n.ShowClusterSetting)
	if changed {
		__antithesis_instrumentation__.Notify(616315)
		ret = n.copyNode()
		ret.ShowClusterSetting = sc.(*ShowClusterSetting)
	} else {
		__antithesis_instrumentation__.Notify(616316)
	}
	__antithesis_instrumentation__.Notify(616313)
	if n.TenantID != nil {
		__antithesis_instrumentation__.Notify(616317)
		e, changed := WalkExpr(v, n.TenantID)
		if changed {
			__antithesis_instrumentation__.Notify(616318)
			if ret == n {
				__antithesis_instrumentation__.Notify(616320)
				ret = n.copyNode()
			} else {
				__antithesis_instrumentation__.Notify(616321)
			}
			__antithesis_instrumentation__.Notify(616319)
			ret.TenantID = e
		} else {
			__antithesis_instrumentation__.Notify(616322)
		}
	} else {
		__antithesis_instrumentation__.Notify(616323)
	}
	__antithesis_instrumentation__.Notify(616314)
	return ret
}

func (n *ShowTenantClusterSettingList) copyNode() *ShowTenantClusterSettingList {
	__antithesis_instrumentation__.Notify(616324)
	stmtCopy := *n
	return &stmtCopy
}

func (n *ShowTenantClusterSettingList) walkStmt(v Visitor) Statement {
	__antithesis_instrumentation__.Notify(616325)
	ret := n
	sc, changed := walkStmt(v, n.ShowClusterSettingList)
	if changed {
		__antithesis_instrumentation__.Notify(616328)
		ret = n.copyNode()
		ret.ShowClusterSettingList = sc.(*ShowClusterSettingList)
	} else {
		__antithesis_instrumentation__.Notify(616329)
	}
	__antithesis_instrumentation__.Notify(616326)
	if n.TenantID != nil {
		__antithesis_instrumentation__.Notify(616330)
		e, changed := WalkExpr(v, n.TenantID)
		if changed {
			__antithesis_instrumentation__.Notify(616331)
			if ret == n {
				__antithesis_instrumentation__.Notify(616333)
				ret = n.copyNode()
			} else {
				__antithesis_instrumentation__.Notify(616334)
			}
			__antithesis_instrumentation__.Notify(616332)
			ret.TenantID = e
		} else {
			__antithesis_instrumentation__.Notify(616335)
		}
	} else {
		__antithesis_instrumentation__.Notify(616336)
	}
	__antithesis_instrumentation__.Notify(616327)
	return ret
}

func (n *AlterTenantSetClusterSetting) copyNode() *AlterTenantSetClusterSetting {
	__antithesis_instrumentation__.Notify(616337)
	stmtCopy := *n
	return &stmtCopy
}

func (n *AlterTenantSetClusterSetting) walkStmt(v Visitor) Statement {
	__antithesis_instrumentation__.Notify(616338)
	ret := n
	if n.Value != nil {
		__antithesis_instrumentation__.Notify(616341)
		e, changed := WalkExpr(v, n.Value)
		if changed {
			__antithesis_instrumentation__.Notify(616342)
			ret = n.copyNode()
			ret.Value = e
		} else {
			__antithesis_instrumentation__.Notify(616343)
		}
	} else {
		__antithesis_instrumentation__.Notify(616344)
	}
	__antithesis_instrumentation__.Notify(616339)
	if n.TenantID != nil {
		__antithesis_instrumentation__.Notify(616345)
		e, changed := WalkExpr(v, n.TenantID)
		if changed {
			__antithesis_instrumentation__.Notify(616346)
			if ret == n {
				__antithesis_instrumentation__.Notify(616348)
				ret = n.copyNode()
			} else {
				__antithesis_instrumentation__.Notify(616349)
			}
			__antithesis_instrumentation__.Notify(616347)
			ret.TenantID = e
		} else {
			__antithesis_instrumentation__.Notify(616350)
		}
	} else {
		__antithesis_instrumentation__.Notify(616351)
	}
	__antithesis_instrumentation__.Notify(616340)
	return ret
}

func (stmt *Backup) copyNode() *Backup {
	__antithesis_instrumentation__.Notify(616352)
	stmtCopy := *stmt
	stmtCopy.IncrementalFrom = append(Exprs(nil), stmt.IncrementalFrom...)
	return &stmtCopy
}

func (stmt *Backup) walkStmt(v Visitor) Statement {
	__antithesis_instrumentation__.Notify(616353)
	ret := stmt
	if stmt.AsOf.Expr != nil {
		__antithesis_instrumentation__.Notify(616358)
		e, changed := WalkExpr(v, stmt.AsOf.Expr)
		if changed {
			__antithesis_instrumentation__.Notify(616359)
			if ret == stmt {
				__antithesis_instrumentation__.Notify(616361)
				ret = stmt.copyNode()
			} else {
				__antithesis_instrumentation__.Notify(616362)
			}
			__antithesis_instrumentation__.Notify(616360)
			ret.AsOf.Expr = e
		} else {
			__antithesis_instrumentation__.Notify(616363)
		}
	} else {
		__antithesis_instrumentation__.Notify(616364)
	}
	__antithesis_instrumentation__.Notify(616354)
	for i, expr := range stmt.To {
		__antithesis_instrumentation__.Notify(616365)
		e, changed := WalkExpr(v, expr)
		if changed {
			__antithesis_instrumentation__.Notify(616366)
			if ret == stmt {
				__antithesis_instrumentation__.Notify(616368)
				ret = stmt.copyNode()
			} else {
				__antithesis_instrumentation__.Notify(616369)
			}
			__antithesis_instrumentation__.Notify(616367)
			ret.To[i] = e
		} else {
			__antithesis_instrumentation__.Notify(616370)
		}
	}
	__antithesis_instrumentation__.Notify(616355)
	for i, expr := range stmt.IncrementalFrom {
		__antithesis_instrumentation__.Notify(616371)
		e, changed := WalkExpr(v, expr)
		if changed {
			__antithesis_instrumentation__.Notify(616372)
			if ret == stmt {
				__antithesis_instrumentation__.Notify(616374)
				ret = stmt.copyNode()
			} else {
				__antithesis_instrumentation__.Notify(616375)
			}
			__antithesis_instrumentation__.Notify(616373)
			ret.IncrementalFrom[i] = e
		} else {
			__antithesis_instrumentation__.Notify(616376)
		}
	}
	__antithesis_instrumentation__.Notify(616356)
	if stmt.Options.EncryptionPassphrase != nil {
		__antithesis_instrumentation__.Notify(616377)
		pw, changed := WalkExpr(v, stmt.Options.EncryptionPassphrase)
		if changed {
			__antithesis_instrumentation__.Notify(616378)
			if ret == stmt {
				__antithesis_instrumentation__.Notify(616380)
				ret = stmt.copyNode()
			} else {
				__antithesis_instrumentation__.Notify(616381)
			}
			__antithesis_instrumentation__.Notify(616379)
			ret.Options.EncryptionPassphrase = pw
		} else {
			__antithesis_instrumentation__.Notify(616382)
		}
	} else {
		__antithesis_instrumentation__.Notify(616383)
	}
	__antithesis_instrumentation__.Notify(616357)
	return ret
}

func (stmt *Delete) copyNode() *Delete {
	__antithesis_instrumentation__.Notify(616384)
	stmtCopy := *stmt
	if stmt.Where != nil {
		__antithesis_instrumentation__.Notify(616386)
		wCopy := *stmt.Where
		stmtCopy.Where = &wCopy
	} else {
		__antithesis_instrumentation__.Notify(616387)
	}
	__antithesis_instrumentation__.Notify(616385)
	return &stmtCopy
}

func (stmt *Delete) walkStmt(v Visitor) Statement {
	__antithesis_instrumentation__.Notify(616388)
	ret := stmt
	if stmt.Where != nil {
		__antithesis_instrumentation__.Notify(616391)
		e, changed := WalkExpr(v, stmt.Where.Expr)
		if changed {
			__antithesis_instrumentation__.Notify(616392)
			ret = stmt.copyNode()
			ret.Where.Expr = e
		} else {
			__antithesis_instrumentation__.Notify(616393)
		}
	} else {
		__antithesis_instrumentation__.Notify(616394)
	}
	__antithesis_instrumentation__.Notify(616389)
	returning, changed := walkReturningClause(v, stmt.Returning)
	if changed {
		__antithesis_instrumentation__.Notify(616395)
		if ret == stmt {
			__antithesis_instrumentation__.Notify(616397)
			ret = stmt.copyNode()
		} else {
			__antithesis_instrumentation__.Notify(616398)
		}
		__antithesis_instrumentation__.Notify(616396)
		ret.Returning = returning
	} else {
		__antithesis_instrumentation__.Notify(616399)
	}
	__antithesis_instrumentation__.Notify(616390)
	return ret
}

func (stmt *Explain) copyNode() *Explain {
	__antithesis_instrumentation__.Notify(616400)
	stmtCopy := *stmt
	return &stmtCopy
}

func (stmt *Explain) walkStmt(v Visitor) Statement {
	__antithesis_instrumentation__.Notify(616401)
	s, changed := walkStmt(v, stmt.Statement)
	if changed {
		__antithesis_instrumentation__.Notify(616403)
		stmt = stmt.copyNode()
		stmt.Statement = s
	} else {
		__antithesis_instrumentation__.Notify(616404)
	}
	__antithesis_instrumentation__.Notify(616402)
	return stmt
}

func (stmt *ExplainAnalyze) copyNode() *ExplainAnalyze {
	__antithesis_instrumentation__.Notify(616405)
	stmtCopy := *stmt
	return &stmtCopy
}

func (stmt *ExplainAnalyze) walkStmt(v Visitor) Statement {
	__antithesis_instrumentation__.Notify(616406)
	s, changed := walkStmt(v, stmt.Statement)
	if changed {
		__antithesis_instrumentation__.Notify(616408)
		stmt = stmt.copyNode()
		stmt.Statement = s
	} else {
		__antithesis_instrumentation__.Notify(616409)
	}
	__antithesis_instrumentation__.Notify(616407)
	return stmt
}

func (stmt *Insert) copyNode() *Insert {
	__antithesis_instrumentation__.Notify(616410)
	stmtCopy := *stmt
	return &stmtCopy
}

func (stmt *Insert) walkStmt(v Visitor) Statement {
	__antithesis_instrumentation__.Notify(616411)
	ret := stmt
	if stmt.Rows != nil {
		__antithesis_instrumentation__.Notify(616414)
		rows, changed := walkStmt(v, stmt.Rows)
		if changed {
			__antithesis_instrumentation__.Notify(616415)
			ret = stmt.copyNode()
			ret.Rows = rows.(*Select)
		} else {
			__antithesis_instrumentation__.Notify(616416)
		}
	} else {
		__antithesis_instrumentation__.Notify(616417)
	}
	__antithesis_instrumentation__.Notify(616412)
	returning, changed := walkReturningClause(v, stmt.Returning)
	if changed {
		__antithesis_instrumentation__.Notify(616418)
		if ret == stmt {
			__antithesis_instrumentation__.Notify(616420)
			ret = stmt.copyNode()
		} else {
			__antithesis_instrumentation__.Notify(616421)
		}
		__antithesis_instrumentation__.Notify(616419)
		ret.Returning = returning
	} else {
		__antithesis_instrumentation__.Notify(616422)
	}
	__antithesis_instrumentation__.Notify(616413)

	return ret
}

func (stmt *CreateTable) copyNode() *CreateTable {
	__antithesis_instrumentation__.Notify(616423)
	stmtCopy := *stmt
	return &stmtCopy
}

func (stmt *CreateTable) walkStmt(v Visitor) Statement {
	__antithesis_instrumentation__.Notify(616424)
	ret := stmt
	if stmt.AsSource != nil {
		__antithesis_instrumentation__.Notify(616426)
		rows, changed := walkStmt(v, stmt.AsSource)
		if changed {
			__antithesis_instrumentation__.Notify(616427)
			ret = stmt.copyNode()
			ret.AsSource = rows.(*Select)
		} else {
			__antithesis_instrumentation__.Notify(616428)
		}
	} else {
		__antithesis_instrumentation__.Notify(616429)
	}
	__antithesis_instrumentation__.Notify(616425)
	return ret
}

func (stmt *CancelQueries) copyNode() *CancelQueries {
	__antithesis_instrumentation__.Notify(616430)
	stmtCopy := *stmt
	return &stmtCopy
}

func (stmt *CancelQueries) walkStmt(v Visitor) Statement {
	__antithesis_instrumentation__.Notify(616431)
	sel, changed := walkStmt(v, stmt.Queries)
	if changed {
		__antithesis_instrumentation__.Notify(616433)
		stmt = stmt.copyNode()
		stmt.Queries = sel.(*Select)
	} else {
		__antithesis_instrumentation__.Notify(616434)
	}
	__antithesis_instrumentation__.Notify(616432)
	return stmt
}

func (stmt *CancelSessions) copyNode() *CancelSessions {
	__antithesis_instrumentation__.Notify(616435)
	stmtCopy := *stmt
	return &stmtCopy
}

func (stmt *CancelSessions) walkStmt(v Visitor) Statement {
	__antithesis_instrumentation__.Notify(616436)
	sel, changed := walkStmt(v, stmt.Sessions)
	if changed {
		__antithesis_instrumentation__.Notify(616438)
		stmt = stmt.copyNode()
		stmt.Sessions = sel.(*Select)
	} else {
		__antithesis_instrumentation__.Notify(616439)
	}
	__antithesis_instrumentation__.Notify(616437)
	return stmt
}

func (stmt *ControlJobs) copyNode() *ControlJobs {
	__antithesis_instrumentation__.Notify(616440)
	stmtCopy := *stmt
	return &stmtCopy
}

func (stmt *ControlJobs) walkStmt(v Visitor) Statement {
	__antithesis_instrumentation__.Notify(616441)
	sel, changed := walkStmt(v, stmt.Jobs)
	if changed {
		__antithesis_instrumentation__.Notify(616443)
		stmt = stmt.copyNode()
		stmt.Jobs = sel.(*Select)
	} else {
		__antithesis_instrumentation__.Notify(616444)
	}
	__antithesis_instrumentation__.Notify(616442)
	return stmt
}

func (n *ControlSchedules) copyNode() *ControlSchedules {
	__antithesis_instrumentation__.Notify(616445)
	stmtCopy := *n
	return &stmtCopy
}

func (n *ControlSchedules) walkStmt(v Visitor) Statement {
	__antithesis_instrumentation__.Notify(616446)
	sel, changed := walkStmt(v, n.Schedules)
	if changed {
		__antithesis_instrumentation__.Notify(616448)
		n = n.copyNode()
		n.Schedules = sel.(*Select)
	} else {
		__antithesis_instrumentation__.Notify(616449)
	}
	__antithesis_instrumentation__.Notify(616447)
	return n
}

func (stmt *Import) copyNode() *Import {
	__antithesis_instrumentation__.Notify(616450)
	stmtCopy := *stmt
	stmtCopy.Files = append(Exprs(nil), stmt.Files...)
	stmtCopy.Options = append(KVOptions(nil), stmt.Options...)
	return &stmtCopy
}

func (stmt *Import) walkStmt(v Visitor) Statement {
	__antithesis_instrumentation__.Notify(616451)
	ret := stmt
	for i, expr := range stmt.Files {
		__antithesis_instrumentation__.Notify(616453)
		e, changed := WalkExpr(v, expr)
		if changed {
			__antithesis_instrumentation__.Notify(616454)
			if ret == stmt {
				__antithesis_instrumentation__.Notify(616456)
				ret = stmt.copyNode()
			} else {
				__antithesis_instrumentation__.Notify(616457)
			}
			__antithesis_instrumentation__.Notify(616455)
			ret.Files[i] = e
		} else {
			__antithesis_instrumentation__.Notify(616458)
		}
	}
	{
		__antithesis_instrumentation__.Notify(616459)
		opts, changed := walkKVOptions(v, stmt.Options)
		if changed {
			__antithesis_instrumentation__.Notify(616460)
			if ret == stmt {
				__antithesis_instrumentation__.Notify(616462)
				ret = stmt.copyNode()
			} else {
				__antithesis_instrumentation__.Notify(616463)
			}
			__antithesis_instrumentation__.Notify(616461)
			ret.Options = opts
		} else {
			__antithesis_instrumentation__.Notify(616464)
		}
	}
	__antithesis_instrumentation__.Notify(616452)
	return ret
}

func (stmt *ParenSelect) walkStmt(v Visitor) Statement {
	__antithesis_instrumentation__.Notify(616465)
	sel, changed := walkStmt(v, stmt.Select)
	if changed {
		__antithesis_instrumentation__.Notify(616467)
		return &ParenSelect{sel.(*Select)}
	} else {
		__antithesis_instrumentation__.Notify(616468)
	}
	__antithesis_instrumentation__.Notify(616466)
	return stmt
}

func (stmt *Restore) copyNode() *Restore {
	__antithesis_instrumentation__.Notify(616469)
	stmtCopy := *stmt
	stmtCopy.From = append([]StringOrPlaceholderOptList(nil), stmt.From...)
	return &stmtCopy
}

func (stmt *Restore) walkStmt(v Visitor) Statement {
	__antithesis_instrumentation__.Notify(616470)
	ret := stmt
	if stmt.AsOf.Expr != nil {
		__antithesis_instrumentation__.Notify(616475)
		e, changed := WalkExpr(v, stmt.AsOf.Expr)
		if changed {
			__antithesis_instrumentation__.Notify(616476)
			if ret == stmt {
				__antithesis_instrumentation__.Notify(616478)
				ret = stmt.copyNode()
			} else {
				__antithesis_instrumentation__.Notify(616479)
			}
			__antithesis_instrumentation__.Notify(616477)
			ret.AsOf.Expr = e
		} else {
			__antithesis_instrumentation__.Notify(616480)
		}
	} else {
		__antithesis_instrumentation__.Notify(616481)
	}
	__antithesis_instrumentation__.Notify(616471)
	for i, backup := range stmt.From {
		__antithesis_instrumentation__.Notify(616482)
		for j, expr := range backup {
			__antithesis_instrumentation__.Notify(616483)
			e, changed := WalkExpr(v, expr)
			if changed {
				__antithesis_instrumentation__.Notify(616484)
				if ret == stmt {
					__antithesis_instrumentation__.Notify(616486)
					ret = stmt.copyNode()
				} else {
					__antithesis_instrumentation__.Notify(616487)
				}
				__antithesis_instrumentation__.Notify(616485)
				ret.From[i][j] = e
			} else {
				__antithesis_instrumentation__.Notify(616488)
			}
		}
	}
	__antithesis_instrumentation__.Notify(616472)

	if stmt.Options.EncryptionPassphrase != nil {
		__antithesis_instrumentation__.Notify(616489)
		pw, changed := WalkExpr(v, stmt.Options.EncryptionPassphrase)
		if changed {
			__antithesis_instrumentation__.Notify(616490)
			if ret == stmt {
				__antithesis_instrumentation__.Notify(616492)
				ret = stmt.copyNode()
			} else {
				__antithesis_instrumentation__.Notify(616493)
			}
			__antithesis_instrumentation__.Notify(616491)
			ret.Options.EncryptionPassphrase = pw
		} else {
			__antithesis_instrumentation__.Notify(616494)
		}
	} else {
		__antithesis_instrumentation__.Notify(616495)
	}
	__antithesis_instrumentation__.Notify(616473)

	if stmt.Options.IntoDB != nil {
		__antithesis_instrumentation__.Notify(616496)
		intoDB, changed := WalkExpr(v, stmt.Options.IntoDB)
		if changed {
			__antithesis_instrumentation__.Notify(616497)
			if ret == stmt {
				__antithesis_instrumentation__.Notify(616499)
				ret = stmt.copyNode()
			} else {
				__antithesis_instrumentation__.Notify(616500)
			}
			__antithesis_instrumentation__.Notify(616498)
			ret.Options.IntoDB = intoDB
		} else {
			__antithesis_instrumentation__.Notify(616501)
		}
	} else {
		__antithesis_instrumentation__.Notify(616502)
	}
	__antithesis_instrumentation__.Notify(616474)
	return ret
}

func (stmt *ReturningExprs) copyNode() *ReturningExprs {
	__antithesis_instrumentation__.Notify(616503)
	stmtCopy := append(ReturningExprs(nil), *stmt...)
	return &stmtCopy
}

func walkOrderBy(v Visitor, order OrderBy) (OrderBy, bool) {
	__antithesis_instrumentation__.Notify(616504)
	copied := false
	for i := range order {
		__antithesis_instrumentation__.Notify(616506)
		if order[i].OrderType != OrderByColumn {
			__antithesis_instrumentation__.Notify(616508)
			continue
		} else {
			__antithesis_instrumentation__.Notify(616509)
		}
		__antithesis_instrumentation__.Notify(616507)
		e, changed := WalkExpr(v, order[i].Expr)
		if changed {
			__antithesis_instrumentation__.Notify(616510)
			if !copied {
				__antithesis_instrumentation__.Notify(616512)
				order = append(OrderBy(nil), order...)
				copied = true
			} else {
				__antithesis_instrumentation__.Notify(616513)
			}
			__antithesis_instrumentation__.Notify(616511)
			orderByCopy := *order[i]
			orderByCopy.Expr = e
			order[i] = &orderByCopy
		} else {
			__antithesis_instrumentation__.Notify(616514)
		}
	}
	__antithesis_instrumentation__.Notify(616505)
	return order, copied
}

func (stmt *Select) copyNode() *Select {
	__antithesis_instrumentation__.Notify(616515)
	stmtCopy := *stmt
	if stmt.Limit != nil {
		__antithesis_instrumentation__.Notify(616518)
		lCopy := *stmt.Limit
		stmtCopy.Limit = &lCopy
	} else {
		__antithesis_instrumentation__.Notify(616519)
	}
	__antithesis_instrumentation__.Notify(616516)
	if stmt.With != nil {
		__antithesis_instrumentation__.Notify(616520)
		withCopy := *stmt.With
		stmtCopy.With = &withCopy
		stmtCopy.With.CTEList = make([]*CTE, len(stmt.With.CTEList))
		for i, cte := range stmt.With.CTEList {
			__antithesis_instrumentation__.Notify(616521)
			cteCopy := *cte
			stmtCopy.With.CTEList[i] = &cteCopy
		}
	} else {
		__antithesis_instrumentation__.Notify(616522)
	}
	__antithesis_instrumentation__.Notify(616517)
	return &stmtCopy
}

func (stmt *Select) walkStmt(v Visitor) Statement {
	__antithesis_instrumentation__.Notify(616523)
	ret := stmt
	sel, changed := walkStmt(v, stmt.Select)
	if changed {
		__antithesis_instrumentation__.Notify(616528)
		ret = stmt.copyNode()
		ret.Select = sel.(SelectStatement)
	} else {
		__antithesis_instrumentation__.Notify(616529)
	}
	__antithesis_instrumentation__.Notify(616524)
	order, changed := walkOrderBy(v, stmt.OrderBy)
	if changed {
		__antithesis_instrumentation__.Notify(616530)
		if ret == stmt {
			__antithesis_instrumentation__.Notify(616532)
			ret = stmt.copyNode()
		} else {
			__antithesis_instrumentation__.Notify(616533)
		}
		__antithesis_instrumentation__.Notify(616531)
		ret.OrderBy = order
	} else {
		__antithesis_instrumentation__.Notify(616534)
	}
	__antithesis_instrumentation__.Notify(616525)
	if stmt.Limit != nil {
		__antithesis_instrumentation__.Notify(616535)
		if stmt.Limit.Offset != nil {
			__antithesis_instrumentation__.Notify(616537)
			e, changed := WalkExpr(v, stmt.Limit.Offset)
			if changed {
				__antithesis_instrumentation__.Notify(616538)
				if ret == stmt {
					__antithesis_instrumentation__.Notify(616540)
					ret = stmt.copyNode()
				} else {
					__antithesis_instrumentation__.Notify(616541)
				}
				__antithesis_instrumentation__.Notify(616539)
				ret.Limit.Offset = e
			} else {
				__antithesis_instrumentation__.Notify(616542)
			}
		} else {
			__antithesis_instrumentation__.Notify(616543)
		}
		__antithesis_instrumentation__.Notify(616536)
		if stmt.Limit.Count != nil {
			__antithesis_instrumentation__.Notify(616544)
			e, changed := WalkExpr(v, stmt.Limit.Count)
			if changed {
				__antithesis_instrumentation__.Notify(616545)
				if ret == stmt {
					__antithesis_instrumentation__.Notify(616547)
					ret = stmt.copyNode()
				} else {
					__antithesis_instrumentation__.Notify(616548)
				}
				__antithesis_instrumentation__.Notify(616546)
				ret.Limit.Count = e
			} else {
				__antithesis_instrumentation__.Notify(616549)
			}
		} else {
			__antithesis_instrumentation__.Notify(616550)
		}
	} else {
		__antithesis_instrumentation__.Notify(616551)
	}
	__antithesis_instrumentation__.Notify(616526)
	if stmt.With != nil {
		__antithesis_instrumentation__.Notify(616552)
		for i := range stmt.With.CTEList {
			__antithesis_instrumentation__.Notify(616553)
			if stmt.With.CTEList[i] != nil {
				__antithesis_instrumentation__.Notify(616554)
				withStmt, changed := walkStmt(v, stmt.With.CTEList[i].Stmt)
				if changed {
					__antithesis_instrumentation__.Notify(616555)
					if ret == stmt {
						__antithesis_instrumentation__.Notify(616557)
						ret = stmt.copyNode()
					} else {
						__antithesis_instrumentation__.Notify(616558)
					}
					__antithesis_instrumentation__.Notify(616556)
					ret.With.CTEList[i].Stmt = withStmt
				} else {
					__antithesis_instrumentation__.Notify(616559)
				}
			} else {
				__antithesis_instrumentation__.Notify(616560)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(616561)
	}
	__antithesis_instrumentation__.Notify(616527)

	return ret
}

func (stmt *SelectClause) copyNode() *SelectClause {
	__antithesis_instrumentation__.Notify(616562)
	stmtCopy := *stmt
	stmtCopy.Exprs = append(SelectExprs(nil), stmt.Exprs...)
	stmtCopy.From = From{
		Tables: append(TableExprs(nil), stmt.From.Tables...),
		AsOf:   stmt.From.AsOf,
	}
	if stmt.Where != nil {
		__antithesis_instrumentation__.Notify(616565)
		wCopy := *stmt.Where
		stmtCopy.Where = &wCopy
	} else {
		__antithesis_instrumentation__.Notify(616566)
	}
	__antithesis_instrumentation__.Notify(616563)
	stmtCopy.GroupBy = append(GroupBy(nil), stmt.GroupBy...)
	if stmt.Having != nil {
		__antithesis_instrumentation__.Notify(616567)
		hCopy := *stmt.Having
		stmtCopy.Having = &hCopy
	} else {
		__antithesis_instrumentation__.Notify(616568)
	}
	__antithesis_instrumentation__.Notify(616564)
	stmtCopy.Window = append(Window(nil), stmt.Window...)
	stmtCopy.DistinctOn = append(DistinctOn(nil), stmt.DistinctOn...)
	return &stmtCopy
}

func (stmt *SelectClause) walkStmt(v Visitor) Statement {
	__antithesis_instrumentation__.Notify(616569)
	ret := stmt

	for i, expr := range stmt.Exprs {
		__antithesis_instrumentation__.Notify(616578)
		e, changed := WalkExpr(v, expr.Expr)
		if changed {
			__antithesis_instrumentation__.Notify(616579)
			if ret == stmt {
				__antithesis_instrumentation__.Notify(616581)
				ret = stmt.copyNode()
			} else {
				__antithesis_instrumentation__.Notify(616582)
			}
			__antithesis_instrumentation__.Notify(616580)
			ret.Exprs[i].Expr = e
		} else {
			__antithesis_instrumentation__.Notify(616583)
		}
	}
	__antithesis_instrumentation__.Notify(616570)

	if stmt.From.AsOf.Expr != nil {
		__antithesis_instrumentation__.Notify(616584)
		e, changed := WalkExpr(v, stmt.From.AsOf.Expr)
		if changed {
			__antithesis_instrumentation__.Notify(616585)
			if ret == stmt {
				__antithesis_instrumentation__.Notify(616587)
				ret = stmt.copyNode()
			} else {
				__antithesis_instrumentation__.Notify(616588)
			}
			__antithesis_instrumentation__.Notify(616586)
			ret.From.AsOf.Expr = e
		} else {
			__antithesis_instrumentation__.Notify(616589)
		}
	} else {
		__antithesis_instrumentation__.Notify(616590)
	}
	__antithesis_instrumentation__.Notify(616571)

	if stmt.Where != nil {
		__antithesis_instrumentation__.Notify(616591)
		e, changed := WalkExpr(v, stmt.Where.Expr)
		if changed {
			__antithesis_instrumentation__.Notify(616592)
			if ret == stmt {
				__antithesis_instrumentation__.Notify(616594)
				ret = stmt.copyNode()
			} else {
				__antithesis_instrumentation__.Notify(616595)
			}
			__antithesis_instrumentation__.Notify(616593)
			ret.Where.Expr = e
		} else {
			__antithesis_instrumentation__.Notify(616596)
		}
	} else {
		__antithesis_instrumentation__.Notify(616597)
	}
	__antithesis_instrumentation__.Notify(616572)

	for i, expr := range stmt.GroupBy {
		__antithesis_instrumentation__.Notify(616598)
		e, changed := WalkExpr(v, expr)
		if changed {
			__antithesis_instrumentation__.Notify(616599)
			if ret == stmt {
				__antithesis_instrumentation__.Notify(616601)
				ret = stmt.copyNode()
			} else {
				__antithesis_instrumentation__.Notify(616602)
			}
			__antithesis_instrumentation__.Notify(616600)
			ret.GroupBy[i] = e
		} else {
			__antithesis_instrumentation__.Notify(616603)
		}
	}
	__antithesis_instrumentation__.Notify(616573)

	if stmt.Having != nil {
		__antithesis_instrumentation__.Notify(616604)
		e, changed := WalkExpr(v, stmt.Having.Expr)
		if changed {
			__antithesis_instrumentation__.Notify(616605)
			if ret == stmt {
				__antithesis_instrumentation__.Notify(616607)
				ret = stmt.copyNode()
			} else {
				__antithesis_instrumentation__.Notify(616608)
			}
			__antithesis_instrumentation__.Notify(616606)
			ret.Having.Expr = e
		} else {
			__antithesis_instrumentation__.Notify(616609)
		}
	} else {
		__antithesis_instrumentation__.Notify(616610)
	}
	__antithesis_instrumentation__.Notify(616574)

	for i := range stmt.Window {
		__antithesis_instrumentation__.Notify(616611)
		w, changed := walkWindowDef(v, stmt.Window[i])
		if changed {
			__antithesis_instrumentation__.Notify(616612)
			if ret == stmt {
				__antithesis_instrumentation__.Notify(616614)
				ret = stmt.copyNode()
			} else {
				__antithesis_instrumentation__.Notify(616615)
			}
			__antithesis_instrumentation__.Notify(616613)
			ret.Window[i] = w
		} else {
			__antithesis_instrumentation__.Notify(616616)
		}
	}
	__antithesis_instrumentation__.Notify(616575)

	for i := range stmt.From.Tables {
		__antithesis_instrumentation__.Notify(616617)
		t, changed := walkTableExpr(v, stmt.From.Tables[i])
		if changed {
			__antithesis_instrumentation__.Notify(616618)
			if ret == stmt {
				__antithesis_instrumentation__.Notify(616620)
				ret = stmt.copyNode()
			} else {
				__antithesis_instrumentation__.Notify(616621)
			}
			__antithesis_instrumentation__.Notify(616619)
			ret.From.Tables[i] = t
		} else {
			__antithesis_instrumentation__.Notify(616622)
		}
	}
	__antithesis_instrumentation__.Notify(616576)

	for i := range stmt.DistinctOn {
		__antithesis_instrumentation__.Notify(616623)
		e, changed := WalkExpr(v, stmt.DistinctOn[i])
		if changed {
			__antithesis_instrumentation__.Notify(616624)
			if ret == stmt {
				__antithesis_instrumentation__.Notify(616626)
				ret = stmt.copyNode()
			} else {
				__antithesis_instrumentation__.Notify(616627)
			}
			__antithesis_instrumentation__.Notify(616625)
			ret.DistinctOn[i] = e
		} else {
			__antithesis_instrumentation__.Notify(616628)
		}
	}
	__antithesis_instrumentation__.Notify(616577)

	return ret
}

func (stmt *UnionClause) walkStmt(v Visitor) Statement {
	__antithesis_instrumentation__.Notify(616629)
	left, changedL := walkStmt(v, stmt.Left)
	right, changedR := walkStmt(v, stmt.Right)
	if changedL || func() bool {
		__antithesis_instrumentation__.Notify(616631)
		return changedR == true
	}() == true {
		__antithesis_instrumentation__.Notify(616632)
		stmtCopy := *stmt
		stmtCopy.Left = left.(*Select)
		stmtCopy.Right = right.(*Select)
		return &stmtCopy
	} else {
		__antithesis_instrumentation__.Notify(616633)
	}
	__antithesis_instrumentation__.Notify(616630)
	return stmt
}

func (stmt *SetVar) copyNode() *SetVar {
	__antithesis_instrumentation__.Notify(616634)
	stmtCopy := *stmt
	stmtCopy.Values = append(Exprs(nil), stmt.Values...)
	return &stmtCopy
}

func (stmt *SetVar) walkStmt(v Visitor) Statement {
	__antithesis_instrumentation__.Notify(616635)
	ret := stmt
	for i, expr := range stmt.Values {
		__antithesis_instrumentation__.Notify(616637)
		e, changed := WalkExpr(v, expr)
		if changed {
			__antithesis_instrumentation__.Notify(616638)
			if ret == stmt {
				__antithesis_instrumentation__.Notify(616640)
				ret = stmt.copyNode()
			} else {
				__antithesis_instrumentation__.Notify(616641)
			}
			__antithesis_instrumentation__.Notify(616639)
			ret.Values[i] = e
		} else {
			__antithesis_instrumentation__.Notify(616642)
		}
	}
	__antithesis_instrumentation__.Notify(616636)
	return ret
}

func (stmt *SetZoneConfig) walkStmt(v Visitor) Statement {
	__antithesis_instrumentation__.Notify(616643)
	ret := stmt
	if stmt.YAMLConfig != nil {
		__antithesis_instrumentation__.Notify(616646)
		e, changed := WalkExpr(v, stmt.YAMLConfig)
		if changed {
			__antithesis_instrumentation__.Notify(616647)
			newStmt := *stmt
			ret = &newStmt
			ret.YAMLConfig = e
		} else {
			__antithesis_instrumentation__.Notify(616648)
		}
	} else {
		__antithesis_instrumentation__.Notify(616649)
	}
	__antithesis_instrumentation__.Notify(616644)
	if stmt.Options != nil {
		__antithesis_instrumentation__.Notify(616650)
		newOpts, changed := walkKVOptions(v, stmt.Options)
		if changed {
			__antithesis_instrumentation__.Notify(616651)
			if ret == stmt {
				__antithesis_instrumentation__.Notify(616653)
				newStmt := *stmt
				ret = &newStmt
			} else {
				__antithesis_instrumentation__.Notify(616654)
			}
			__antithesis_instrumentation__.Notify(616652)
			ret.Options = newOpts
		} else {
			__antithesis_instrumentation__.Notify(616655)
		}
	} else {
		__antithesis_instrumentation__.Notify(616656)
	}
	__antithesis_instrumentation__.Notify(616645)
	return ret
}

func (stmt *SetTracing) copyNode() *SetTracing {
	__antithesis_instrumentation__.Notify(616657)
	stmtCopy := *stmt
	stmtCopy.Values = append(Exprs(nil), stmt.Values...)
	return &stmtCopy
}

func (stmt *SetTracing) walkStmt(v Visitor) Statement {
	__antithesis_instrumentation__.Notify(616658)
	ret := stmt
	for i, expr := range stmt.Values {
		__antithesis_instrumentation__.Notify(616660)
		e, changed := WalkExpr(v, expr)
		if changed {
			__antithesis_instrumentation__.Notify(616661)
			if ret == stmt {
				__antithesis_instrumentation__.Notify(616663)
				ret = stmt.copyNode()
			} else {
				__antithesis_instrumentation__.Notify(616664)
			}
			__antithesis_instrumentation__.Notify(616662)
			ret.Values[i] = e
		} else {
			__antithesis_instrumentation__.Notify(616665)
		}
	}
	__antithesis_instrumentation__.Notify(616659)
	return ret
}

func (stmt *SetClusterSetting) copyNode() *SetClusterSetting {
	__antithesis_instrumentation__.Notify(616666)
	stmtCopy := *stmt
	return &stmtCopy
}

func (stmt *SetClusterSetting) walkStmt(v Visitor) Statement {
	__antithesis_instrumentation__.Notify(616667)
	ret := stmt
	if stmt.Value != nil {
		__antithesis_instrumentation__.Notify(616669)
		e, changed := WalkExpr(v, stmt.Value)
		if changed {
			__antithesis_instrumentation__.Notify(616670)
			ret = stmt.copyNode()
			ret.Value = e
		} else {
			__antithesis_instrumentation__.Notify(616671)
		}
	} else {
		__antithesis_instrumentation__.Notify(616672)
	}
	__antithesis_instrumentation__.Notify(616668)
	return ret
}

func (stmt *Update) copyNode() *Update {
	__antithesis_instrumentation__.Notify(616673)
	stmtCopy := *stmt
	stmtCopy.Exprs = make(UpdateExprs, len(stmt.Exprs))
	for i, e := range stmt.Exprs {
		__antithesis_instrumentation__.Notify(616676)
		eCopy := *e
		stmtCopy.Exprs[i] = &eCopy
	}
	__antithesis_instrumentation__.Notify(616674)
	if stmt.Where != nil {
		__antithesis_instrumentation__.Notify(616677)
		wCopy := *stmt.Where
		stmtCopy.Where = &wCopy
	} else {
		__antithesis_instrumentation__.Notify(616678)
	}
	__antithesis_instrumentation__.Notify(616675)
	return &stmtCopy
}

func (stmt *Update) walkStmt(v Visitor) Statement {
	__antithesis_instrumentation__.Notify(616679)
	ret := stmt
	for i, expr := range stmt.Exprs {
		__antithesis_instrumentation__.Notify(616683)
		e, changed := WalkExpr(v, expr.Expr)
		if changed {
			__antithesis_instrumentation__.Notify(616684)
			if ret == stmt {
				__antithesis_instrumentation__.Notify(616686)
				ret = stmt.copyNode()
			} else {
				__antithesis_instrumentation__.Notify(616687)
			}
			__antithesis_instrumentation__.Notify(616685)
			ret.Exprs[i].Expr = e
		} else {
			__antithesis_instrumentation__.Notify(616688)
		}
	}
	__antithesis_instrumentation__.Notify(616680)

	if stmt.Where != nil {
		__antithesis_instrumentation__.Notify(616689)
		e, changed := WalkExpr(v, stmt.Where.Expr)
		if changed {
			__antithesis_instrumentation__.Notify(616690)
			if ret == stmt {
				__antithesis_instrumentation__.Notify(616692)
				ret = stmt.copyNode()
			} else {
				__antithesis_instrumentation__.Notify(616693)
			}
			__antithesis_instrumentation__.Notify(616691)
			ret.Where.Expr = e
		} else {
			__antithesis_instrumentation__.Notify(616694)
		}
	} else {
		__antithesis_instrumentation__.Notify(616695)
	}
	__antithesis_instrumentation__.Notify(616681)

	returning, changed := walkReturningClause(v, stmt.Returning)
	if changed {
		__antithesis_instrumentation__.Notify(616696)
		if ret == stmt {
			__antithesis_instrumentation__.Notify(616698)
			ret = stmt.copyNode()
		} else {
			__antithesis_instrumentation__.Notify(616699)
		}
		__antithesis_instrumentation__.Notify(616697)
		ret.Returning = returning
	} else {
		__antithesis_instrumentation__.Notify(616700)
	}
	__antithesis_instrumentation__.Notify(616682)
	return ret
}

func (stmt *ValuesClause) walkStmt(v Visitor) Statement {
	__antithesis_instrumentation__.Notify(616701)
	ret := stmt
	for i, tuple := range stmt.Rows {
		__antithesis_instrumentation__.Notify(616703)
		exprs, changed := walkExprSlice(v, tuple)
		if changed {
			__antithesis_instrumentation__.Notify(616704)
			if ret == stmt {
				__antithesis_instrumentation__.Notify(616706)
				ret = &ValuesClause{append([]Exprs(nil), stmt.Rows...)}
			} else {
				__antithesis_instrumentation__.Notify(616707)
			}
			__antithesis_instrumentation__.Notify(616705)
			ret.Rows[i] = exprs
		} else {
			__antithesis_instrumentation__.Notify(616708)
		}
	}
	__antithesis_instrumentation__.Notify(616702)
	return ret
}

func (stmt *BeginTransaction) copyNode() *BeginTransaction {
	__antithesis_instrumentation__.Notify(616709)
	stmtCopy := *stmt
	return &stmtCopy
}

func (stmt *BeginTransaction) walkStmt(v Visitor) Statement {
	__antithesis_instrumentation__.Notify(616710)
	ret := stmt
	if stmt.Modes.AsOf.Expr != nil {
		__antithesis_instrumentation__.Notify(616712)
		e, changed := WalkExpr(v, stmt.Modes.AsOf.Expr)
		if changed {
			__antithesis_instrumentation__.Notify(616713)
			ret = stmt.copyNode()
			ret.Modes.AsOf.Expr = e
		} else {
			__antithesis_instrumentation__.Notify(616714)
		}
	} else {
		__antithesis_instrumentation__.Notify(616715)
	}
	__antithesis_instrumentation__.Notify(616711)
	return ret
}

var _ walkableStmt = &AlterTenantSetClusterSetting{}
var _ walkableStmt = &CreateTable{}
var _ walkableStmt = &Backup{}
var _ walkableStmt = &Delete{}
var _ walkableStmt = &Explain{}
var _ walkableStmt = &Insert{}
var _ walkableStmt = &Import{}
var _ walkableStmt = &ParenSelect{}
var _ walkableStmt = &Restore{}
var _ walkableStmt = &Select{}
var _ walkableStmt = &SelectClause{}
var _ walkableStmt = &SetClusterSetting{}
var _ walkableStmt = &SetVar{}
var _ walkableStmt = &Update{}
var _ walkableStmt = &ValuesClause{}
var _ walkableStmt = &CancelQueries{}
var _ walkableStmt = &CancelSessions{}
var _ walkableStmt = &ControlJobs{}
var _ walkableStmt = &ControlSchedules{}
var _ walkableStmt = &BeginTransaction{}
var _ walkableStmt = &UnionClause{}

func walkStmt(v Visitor, stmt Statement) (newStmt Statement, changed bool) {
	__antithesis_instrumentation__.Notify(616716)
	walkable, ok := stmt.(walkableStmt)
	if !ok {
		__antithesis_instrumentation__.Notify(616718)
		return stmt, false
	} else {
		__antithesis_instrumentation__.Notify(616719)
	}
	__antithesis_instrumentation__.Notify(616717)
	newStmt = walkable.walkStmt(v)
	return newStmt, (stmt != newStmt)
}

type simpleVisitor struct {
	fn  SimpleVisitFn
	err error
}

var _ Visitor = &simpleVisitor{}

func (v *simpleVisitor) VisitPre(expr Expr) (recurse bool, newExpr Expr) {
	__antithesis_instrumentation__.Notify(616720)
	if v.err != nil {
		__antithesis_instrumentation__.Notify(616723)
		return false, expr
	} else {
		__antithesis_instrumentation__.Notify(616724)
	}
	__antithesis_instrumentation__.Notify(616721)
	recurse, newExpr, v.err = v.fn(expr)
	if v.err != nil {
		__antithesis_instrumentation__.Notify(616725)
		return false, expr
	} else {
		__antithesis_instrumentation__.Notify(616726)
	}
	__antithesis_instrumentation__.Notify(616722)
	return recurse, newExpr
}

func (*simpleVisitor) VisitPost(expr Expr) Expr {
	__antithesis_instrumentation__.Notify(616727)
	return expr
}

type SimpleVisitFn func(expr Expr) (recurse bool, newExpr Expr, err error)

func SimpleVisit(expr Expr, preFn SimpleVisitFn) (Expr, error) {
	__antithesis_instrumentation__.Notify(616728)
	v := simpleVisitor{fn: preFn}
	newExpr, _ := WalkExpr(&v, expr)
	if v.err != nil {
		__antithesis_instrumentation__.Notify(616730)
		return nil, v.err
	} else {
		__antithesis_instrumentation__.Notify(616731)
	}
	__antithesis_instrumentation__.Notify(616729)
	return newExpr, nil
}

func SimpleStmtVisit(stmt Statement, preFn SimpleVisitFn) (Statement, error) {
	__antithesis_instrumentation__.Notify(616732)
	v := simpleVisitor{fn: preFn}
	newStmt, changed := walkStmt(&v, stmt)
	if changed {
		__antithesis_instrumentation__.Notify(616734)
		return newStmt, nil
	} else {
		__antithesis_instrumentation__.Notify(616735)
	}
	__antithesis_instrumentation__.Notify(616733)
	return stmt, nil
}

type debugVisitor struct {
	buf   bytes.Buffer
	level int
}

var _ Visitor = &debugVisitor{}

func (v *debugVisitor) VisitPre(expr Expr) (recurse bool, newExpr Expr) {
	__antithesis_instrumentation__.Notify(616736)
	v.level++
	fmt.Fprintf(&v.buf, "%*s", 2*v.level, " ")
	str := fmt.Sprintf("%#v\n", expr)

	str = strings.Replace(str, "parser.", "", -1)
	v.buf.WriteString(str)
	return true, expr
}

func (v *debugVisitor) VisitPost(expr Expr) Expr {
	__antithesis_instrumentation__.Notify(616737)
	v.level--
	return expr
}

func ExprDebugString(expr Expr) string {
	__antithesis_instrumentation__.Notify(616738)
	v := debugVisitor{}
	WalkExprConst(&v, expr)
	return v.buf.String()
}

func StmtDebugString(stmt Statement) string {
	__antithesis_instrumentation__.Notify(616739)
	v := debugVisitor{}
	walkStmt(&v, stmt)
	return v.buf.String()
}

var _ = ExprDebugString
var _ = StmtDebugString
