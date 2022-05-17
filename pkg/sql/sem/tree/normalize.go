package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree/treebin"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree/treecmp"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
)

type normalizableExpr interface {
	Expr
	normalize(*NormalizeVisitor) TypedExpr
}

func (expr *CastExpr) normalize(v *NormalizeVisitor) TypedExpr {
	__antithesis_instrumentation__.Notify(610538)
	return expr
}

func (expr *CoalesceExpr) normalize(v *NormalizeVisitor) TypedExpr {
	__antithesis_instrumentation__.Notify(610539)

	last := len(expr.Exprs) - 1
	for i := range expr.Exprs {
		__antithesis_instrumentation__.Notify(610541)
		subExpr := expr.TypedExprAt(i)

		if i == last {
			__antithesis_instrumentation__.Notify(610545)
			return subExpr
		} else {
			__antithesis_instrumentation__.Notify(610546)
		}
		__antithesis_instrumentation__.Notify(610542)

		if !v.isConst(subExpr) {
			__antithesis_instrumentation__.Notify(610547)
			exprCopy := *expr
			exprCopy.Exprs = expr.Exprs[i:]
			return &exprCopy
		} else {
			__antithesis_instrumentation__.Notify(610548)
		}
		__antithesis_instrumentation__.Notify(610543)

		val, err := subExpr.Eval(v.ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(610549)
			v.err = err
			return expr
		} else {
			__antithesis_instrumentation__.Notify(610550)
		}
		__antithesis_instrumentation__.Notify(610544)

		if val != DNull {
			__antithesis_instrumentation__.Notify(610551)
			return subExpr
		} else {
			__antithesis_instrumentation__.Notify(610552)
		}
	}
	__antithesis_instrumentation__.Notify(610540)
	return expr
}

func (expr *IfExpr) normalize(v *NormalizeVisitor) TypedExpr {
	__antithesis_instrumentation__.Notify(610553)
	if v.isConst(expr.Cond) {
		__antithesis_instrumentation__.Notify(610555)
		cond, err := expr.TypedCondExpr().Eval(v.ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(610558)
			v.err = err
			return expr
		} else {
			__antithesis_instrumentation__.Notify(610559)
		}
		__antithesis_instrumentation__.Notify(610556)
		if d, err := GetBool(cond); err == nil {
			__antithesis_instrumentation__.Notify(610560)
			if d {
				__antithesis_instrumentation__.Notify(610562)
				return expr.TypedTrueExpr()
			} else {
				__antithesis_instrumentation__.Notify(610563)
			}
			__antithesis_instrumentation__.Notify(610561)
			return expr.TypedElseExpr()
		} else {
			__antithesis_instrumentation__.Notify(610564)
		}
		__antithesis_instrumentation__.Notify(610557)
		return DNull
	} else {
		__antithesis_instrumentation__.Notify(610565)
	}
	__antithesis_instrumentation__.Notify(610554)
	return expr
}

func (expr *UnaryExpr) normalize(v *NormalizeVisitor) TypedExpr {
	__antithesis_instrumentation__.Notify(610566)
	val := expr.TypedInnerExpr()

	if val == DNull {
		__antithesis_instrumentation__.Notify(610569)
		return val
	} else {
		__antithesis_instrumentation__.Notify(610570)
	}
	__antithesis_instrumentation__.Notify(610567)

	switch expr.Operator.Symbol {
	case UnaryMinus:
		__antithesis_instrumentation__.Notify(610571)
		if expr.Operator.IsExplicitOperator {
			__antithesis_instrumentation__.Notify(610575)
			return expr
		} else {
			__antithesis_instrumentation__.Notify(610576)
		}
		__antithesis_instrumentation__.Notify(610572)

		if val.ResolvedType().Family() != types.FloatFamily && func() bool {
			__antithesis_instrumentation__.Notify(610577)
			return v.isNumericZero(val) == true
		}() == true {
			__antithesis_instrumentation__.Notify(610578)
			return val
		} else {
			__antithesis_instrumentation__.Notify(610579)
		}
		__antithesis_instrumentation__.Notify(610573)
		switch b := val.(type) {

		case *BinaryExpr:
			__antithesis_instrumentation__.Notify(610580)
			if b.Operator.Symbol == treebin.Minus {
				__antithesis_instrumentation__.Notify(610582)
				newBinExpr := newBinExprIfValidOverload(
					treebin.MakeBinaryOperator(treebin.Minus),
					b.TypedRight(),
					b.TypedLeft(),
				)
				if newBinExpr != nil {
					__antithesis_instrumentation__.Notify(610584)
					newBinExpr.memoizeFn()
					b = newBinExpr
				} else {
					__antithesis_instrumentation__.Notify(610585)
				}
				__antithesis_instrumentation__.Notify(610583)
				return b
			} else {
				__antithesis_instrumentation__.Notify(610586)
			}

		case *UnaryExpr:
			__antithesis_instrumentation__.Notify(610581)
			if b.Operator.Symbol == UnaryMinus {
				__antithesis_instrumentation__.Notify(610587)
				return b.TypedInnerExpr()
			} else {
				__antithesis_instrumentation__.Notify(610588)
			}
		}
	default:
		__antithesis_instrumentation__.Notify(610574)
	}
	__antithesis_instrumentation__.Notify(610568)

	return expr
}

func (expr *BinaryExpr) normalize(v *NormalizeVisitor) TypedExpr {
	__antithesis_instrumentation__.Notify(610589)
	left := expr.TypedLeft()
	right := expr.TypedRight()
	expectedType := expr.ResolvedType()

	if !expr.Fn.NullableArgs && func() bool {
		__antithesis_instrumentation__.Notify(610593)
		return (left == DNull || func() bool {
			__antithesis_instrumentation__.Notify(610594)
			return right == DNull == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(610595)
		return DNull
	} else {
		__antithesis_instrumentation__.Notify(610596)
	}
	__antithesis_instrumentation__.Notify(610590)

	var final TypedExpr

	switch expr.Operator.Symbol {
	case treebin.Plus:
		__antithesis_instrumentation__.Notify(610597)
		if v.isNumericZero(right) {
			__antithesis_instrumentation__.Notify(610604)
			final, _ = ReType(left, expectedType)
			break
		} else {
			__antithesis_instrumentation__.Notify(610605)
		}
		__antithesis_instrumentation__.Notify(610598)
		if v.isNumericZero(left) {
			__antithesis_instrumentation__.Notify(610606)
			final, _ = ReType(right, expectedType)
			break
		} else {
			__antithesis_instrumentation__.Notify(610607)
		}
	case treebin.Minus:
		__antithesis_instrumentation__.Notify(610599)
		if types.IsAdditiveType(left.ResolvedType()) && func() bool {
			__antithesis_instrumentation__.Notify(610608)
			return v.isNumericZero(right) == true
		}() == true {
			__antithesis_instrumentation__.Notify(610609)
			final, _ = ReType(left, expectedType)
			break
		} else {
			__antithesis_instrumentation__.Notify(610610)
		}
	case treebin.Mult:
		__antithesis_instrumentation__.Notify(610600)
		if v.isNumericOne(right) {
			__antithesis_instrumentation__.Notify(610611)
			final, _ = ReType(left, expectedType)
			break
		} else {
			__antithesis_instrumentation__.Notify(610612)
		}
		__antithesis_instrumentation__.Notify(610601)
		if v.isNumericOne(left) {
			__antithesis_instrumentation__.Notify(610613)
			final, _ = ReType(right, expectedType)
			break
		} else {
			__antithesis_instrumentation__.Notify(610614)
		}

	case treebin.Div, treebin.FloorDiv:
		__antithesis_instrumentation__.Notify(610602)
		if v.isNumericOne(right) {
			__antithesis_instrumentation__.Notify(610615)
			final, _ = ReType(left, expectedType)
			break
		} else {
			__antithesis_instrumentation__.Notify(610616)
		}
	default:
		__antithesis_instrumentation__.Notify(610603)
	}
	__antithesis_instrumentation__.Notify(610591)

	if final == nil {
		__antithesis_instrumentation__.Notify(610617)
		return expr
	} else {
		__antithesis_instrumentation__.Notify(610618)
	}
	__antithesis_instrumentation__.Notify(610592)
	return final
}

func (expr *AndExpr) normalize(v *NormalizeVisitor) TypedExpr {
	__antithesis_instrumentation__.Notify(610619)
	left := expr.TypedLeft()
	right := expr.TypedRight()
	var dleft, dright Datum

	if left == DNull && func() bool {
		__antithesis_instrumentation__.Notify(610623)
		return right == DNull == true
	}() == true {
		__antithesis_instrumentation__.Notify(610624)
		return DNull
	} else {
		__antithesis_instrumentation__.Notify(610625)
	}
	__antithesis_instrumentation__.Notify(610620)

	if v.isConst(left) {
		__antithesis_instrumentation__.Notify(610626)
		dleft, v.err = left.Eval(v.ctx)
		if v.err != nil {
			__antithesis_instrumentation__.Notify(610629)
			return expr
		} else {
			__antithesis_instrumentation__.Notify(610630)
		}
		__antithesis_instrumentation__.Notify(610627)
		if dleft != DNull {
			__antithesis_instrumentation__.Notify(610631)
			if d, err := GetBool(dleft); err == nil {
				__antithesis_instrumentation__.Notify(610633)
				if !d {
					__antithesis_instrumentation__.Notify(610635)
					return dleft
				} else {
					__antithesis_instrumentation__.Notify(610636)
				}
				__antithesis_instrumentation__.Notify(610634)
				return right
			} else {
				__antithesis_instrumentation__.Notify(610637)
			}
			__antithesis_instrumentation__.Notify(610632)
			return DNull
		} else {
			__antithesis_instrumentation__.Notify(610638)
		}
		__antithesis_instrumentation__.Notify(610628)
		return NewTypedAndExpr(
			dleft,
			right,
		)
	} else {
		__antithesis_instrumentation__.Notify(610639)
	}
	__antithesis_instrumentation__.Notify(610621)
	if v.isConst(right) {
		__antithesis_instrumentation__.Notify(610640)
		dright, v.err = right.Eval(v.ctx)
		if v.err != nil {
			__antithesis_instrumentation__.Notify(610643)
			return expr
		} else {
			__antithesis_instrumentation__.Notify(610644)
		}
		__antithesis_instrumentation__.Notify(610641)
		if dright != DNull {
			__antithesis_instrumentation__.Notify(610645)
			if d, err := GetBool(dright); err == nil {
				__antithesis_instrumentation__.Notify(610647)
				if !d {
					__antithesis_instrumentation__.Notify(610649)
					return right
				} else {
					__antithesis_instrumentation__.Notify(610650)
				}
				__antithesis_instrumentation__.Notify(610648)
				return left
			} else {
				__antithesis_instrumentation__.Notify(610651)
			}
			__antithesis_instrumentation__.Notify(610646)
			return DNull
		} else {
			__antithesis_instrumentation__.Notify(610652)
		}
		__antithesis_instrumentation__.Notify(610642)
		return NewTypedAndExpr(
			left,
			dright,
		)
	} else {
		__antithesis_instrumentation__.Notify(610653)
	}
	__antithesis_instrumentation__.Notify(610622)
	return expr
}

func (expr *ComparisonExpr) normalize(v *NormalizeVisitor) TypedExpr {
	__antithesis_instrumentation__.Notify(610654)
	switch expr.Operator.Symbol {
	case treecmp.EQ, treecmp.GE, treecmp.GT, treecmp.LE, treecmp.LT:
		__antithesis_instrumentation__.Notify(610656)

		exprCopied := false
		for {
			__antithesis_instrumentation__.Notify(610661)
			if expr.TypedLeft() == DNull || func() bool {
				__antithesis_instrumentation__.Notify(610666)
				return expr.TypedRight() == DNull == true
			}() == true {
				__antithesis_instrumentation__.Notify(610667)
				return DNull
			} else {
				__antithesis_instrumentation__.Notify(610668)
			}
			__antithesis_instrumentation__.Notify(610662)

			if v.isConst(expr.Left) {
				__antithesis_instrumentation__.Notify(610669)
				switch expr.Right.(type) {
				case *BinaryExpr, VariableExpr:
					__antithesis_instrumentation__.Notify(610673)
					break
				default:
					__antithesis_instrumentation__.Notify(610674)
					return expr
				}
				__antithesis_instrumentation__.Notify(610670)

				invertedOp, err := invertComparisonOp(expr.Operator)
				if err != nil {
					__antithesis_instrumentation__.Notify(610675)
					v.err = err
					return expr
				} else {
					__antithesis_instrumentation__.Notify(610676)
				}
				__antithesis_instrumentation__.Notify(610671)

				if !exprCopied {
					__antithesis_instrumentation__.Notify(610677)
					exprCopy := *expr
					expr = &exprCopy
					exprCopied = true
				} else {
					__antithesis_instrumentation__.Notify(610678)
				}
				__antithesis_instrumentation__.Notify(610672)

				expr = NewTypedComparisonExpr(invertedOp, expr.TypedRight(), expr.TypedLeft())
			} else {
				__antithesis_instrumentation__.Notify(610679)
				if !v.isConst(expr.Right) {
					__antithesis_instrumentation__.Notify(610680)
					return expr
				} else {
					__antithesis_instrumentation__.Notify(610681)
				}
			}
			__antithesis_instrumentation__.Notify(610663)

			left, ok := expr.Left.(*BinaryExpr)
			if !ok {
				__antithesis_instrumentation__.Notify(610682)
				return expr
			} else {
				__antithesis_instrumentation__.Notify(610683)
			}
			__antithesis_instrumentation__.Notify(610664)

			switch {
			case v.isConst(left.Right) && func() bool {
				__antithesis_instrumentation__.Notify(610695)
				return (left.Operator.Symbol == treebin.Plus || func() bool {
					__antithesis_instrumentation__.Notify(610696)
					return left.Operator.Symbol == treebin.Minus == true
				}() == true || func() bool {
					__antithesis_instrumentation__.Notify(610697)
					return left.Operator.Symbol == treebin.Div == true
				}() == true) == true
			}() == true:
				__antithesis_instrumentation__.Notify(610684)

				var op treebin.BinaryOperator
				switch left.Operator.Symbol {
				case treebin.Plus:
					__antithesis_instrumentation__.Notify(610698)
					op = treebin.MakeBinaryOperator(treebin.Minus)
				case treebin.Minus:
					__antithesis_instrumentation__.Notify(610699)
					op = treebin.MakeBinaryOperator(treebin.Plus)
				case treebin.Div:
					__antithesis_instrumentation__.Notify(610700)
					op = treebin.MakeBinaryOperator(treebin.Mult)
					if expr.Operator.Symbol != treecmp.EQ {
						__antithesis_instrumentation__.Notify(610702)

						divisor, err := left.TypedRight().Eval(v.ctx)
						if err != nil {
							__antithesis_instrumentation__.Notify(610704)
							v.err = err
							return expr
						} else {
							__antithesis_instrumentation__.Notify(610705)
						}
						__antithesis_instrumentation__.Notify(610703)
						if divisor.Compare(v.ctx, DZero) < 0 {
							__antithesis_instrumentation__.Notify(610706)
							if !exprCopied {
								__antithesis_instrumentation__.Notify(610709)
								exprCopy := *expr
								expr = &exprCopy
								exprCopied = true
							} else {
								__antithesis_instrumentation__.Notify(610710)
							}
							__antithesis_instrumentation__.Notify(610707)

							invertedOp, err := invertComparisonOp(expr.Operator)
							if err != nil {
								__antithesis_instrumentation__.Notify(610711)
								v.err = err
								return expr
							} else {
								__antithesis_instrumentation__.Notify(610712)
							}
							__antithesis_instrumentation__.Notify(610708)
							expr = NewTypedComparisonExpr(invertedOp, expr.TypedLeft(), expr.TypedRight())
						} else {
							__antithesis_instrumentation__.Notify(610713)
						}
					} else {
						__antithesis_instrumentation__.Notify(610714)
					}
				default:
					__antithesis_instrumentation__.Notify(610701)
				}
				__antithesis_instrumentation__.Notify(610685)

				newBinExpr := newBinExprIfValidOverload(op,
					expr.TypedRight(), left.TypedRight())
				if newBinExpr == nil {
					__antithesis_instrumentation__.Notify(610715)

					break
				} else {
					__antithesis_instrumentation__.Notify(610716)
				}
				__antithesis_instrumentation__.Notify(610686)

				newRightExpr, err := newBinExpr.Eval(v.ctx)
				if err != nil {
					__antithesis_instrumentation__.Notify(610717)

					break
				} else {
					__antithesis_instrumentation__.Notify(610718)
				}
				__antithesis_instrumentation__.Notify(610687)

				if !exprCopied {
					__antithesis_instrumentation__.Notify(610719)
					exprCopy := *expr
					expr = &exprCopy
					exprCopied = true
				} else {
					__antithesis_instrumentation__.Notify(610720)
				}
				__antithesis_instrumentation__.Notify(610688)

				expr.Left = left.Left
				expr.Right = newRightExpr
				expr.memoizeFn()
				if !isVar(v.ctx, expr.Left, true) {
					__antithesis_instrumentation__.Notify(610721)

					continue
				} else {
					__antithesis_instrumentation__.Notify(610722)
				}

			case v.isConst(left.Left) && func() bool {
				__antithesis_instrumentation__.Notify(610723)
				return (left.Operator.Symbol == treebin.Plus || func() bool {
					__antithesis_instrumentation__.Notify(610724)
					return left.Operator.Symbol == treebin.Minus == true
				}() == true) == true
			}() == true:
				__antithesis_instrumentation__.Notify(610689)

				op := expr.Operator
				var newBinExpr *BinaryExpr

				switch left.Operator.Symbol {
				case treebin.Plus:
					__antithesis_instrumentation__.Notify(610725)

					newBinExpr = newBinExprIfValidOverload(
						treebin.MakeBinaryOperator(treebin.Minus),
						expr.TypedRight(),
						left.TypedLeft(),
					)
				case treebin.Minus:
					__antithesis_instrumentation__.Notify(610726)

					newBinExpr = newBinExprIfValidOverload(
						treebin.MakeBinaryOperator(treebin.Minus),
						left.TypedLeft(),
						expr.TypedRight(),
					)
					op, v.err = invertComparisonOp(op)
					if v.err != nil {
						__antithesis_instrumentation__.Notify(610728)
						return expr
					} else {
						__antithesis_instrumentation__.Notify(610729)
					}
				default:
					__antithesis_instrumentation__.Notify(610727)
				}
				__antithesis_instrumentation__.Notify(610690)

				if newBinExpr == nil {
					__antithesis_instrumentation__.Notify(610730)
					break
				} else {
					__antithesis_instrumentation__.Notify(610731)
				}
				__antithesis_instrumentation__.Notify(610691)

				newRightExpr, err := newBinExpr.Eval(v.ctx)
				if err != nil {
					__antithesis_instrumentation__.Notify(610732)
					break
				} else {
					__antithesis_instrumentation__.Notify(610733)
				}
				__antithesis_instrumentation__.Notify(610692)

				if !exprCopied {
					__antithesis_instrumentation__.Notify(610734)
					exprCopy := *expr
					expr = &exprCopy
					exprCopied = true
				} else {
					__antithesis_instrumentation__.Notify(610735)
				}
				__antithesis_instrumentation__.Notify(610693)

				expr.Operator = op
				expr.Left = left.Right
				expr.Right = newRightExpr
				expr.memoizeFn()
				if !isVar(v.ctx, expr.Left, true) {
					__antithesis_instrumentation__.Notify(610736)

					continue
				} else {
					__antithesis_instrumentation__.Notify(610737)
				}
			default:
				__antithesis_instrumentation__.Notify(610694)
			}
			__antithesis_instrumentation__.Notify(610665)

			break
		}
	case treecmp.In, treecmp.NotIn:
		__antithesis_instrumentation__.Notify(610657)

		tuple, ok := expr.Right.(*DTuple)
		if ok {
			__antithesis_instrumentation__.Notify(610738)
			tupleCopy := *tuple
			tupleCopy.Normalize(v.ctx)

			if len(tupleCopy.D) == 1 && func() bool {
				__antithesis_instrumentation__.Notify(610742)
				return tupleCopy.D[0] == DNull == true
			}() == true {
				__antithesis_instrumentation__.Notify(610743)
				return DNull
			} else {
				__antithesis_instrumentation__.Notify(610744)
			}
			__antithesis_instrumentation__.Notify(610739)
			if len(tupleCopy.D) == 0 {
				__antithesis_instrumentation__.Notify(610745)

				if expr.Operator.Symbol == treecmp.In {
					__antithesis_instrumentation__.Notify(610747)
					return DBoolFalse
				} else {
					__antithesis_instrumentation__.Notify(610748)
				}
				__antithesis_instrumentation__.Notify(610746)
				return DBoolTrue
			} else {
				__antithesis_instrumentation__.Notify(610749)
			}
			__antithesis_instrumentation__.Notify(610740)
			if expr.TypedLeft() == DNull {
				__antithesis_instrumentation__.Notify(610750)

				return DNull
			} else {
				__antithesis_instrumentation__.Notify(610751)
			}
			__antithesis_instrumentation__.Notify(610741)

			exprCopy := *expr
			expr = &exprCopy
			expr.Right = &tupleCopy
		} else {
			__antithesis_instrumentation__.Notify(610752)
		}
	case treecmp.IsDistinctFrom, treecmp.IsNotDistinctFrom:
		__antithesis_instrumentation__.Notify(610658)
		left := expr.TypedLeft()
		right := expr.TypedRight()

		if v.isConst(left) && func() bool {
			__antithesis_instrumentation__.Notify(610753)
			return !v.isConst(right) == true
		}() == true {
			__antithesis_instrumentation__.Notify(610754)

			return NewTypedComparisonExpr(expr.Operator, right, left)
		} else {
			__antithesis_instrumentation__.Notify(610755)
		}
	case treecmp.NE,
		treecmp.Like, treecmp.NotLike,
		treecmp.ILike, treecmp.NotILike,
		treecmp.SimilarTo, treecmp.NotSimilarTo,
		treecmp.RegMatch, treecmp.NotRegMatch,
		treecmp.RegIMatch, treecmp.NotRegIMatch,
		treecmp.Any, treecmp.Some, treecmp.All:
		__antithesis_instrumentation__.Notify(610659)
		if expr.TypedLeft() == DNull || func() bool {
			__antithesis_instrumentation__.Notify(610756)
			return expr.TypedRight() == DNull == true
		}() == true {
			__antithesis_instrumentation__.Notify(610757)
			return DNull
		} else {
			__antithesis_instrumentation__.Notify(610758)
		}
	default:
		__antithesis_instrumentation__.Notify(610660)
	}
	__antithesis_instrumentation__.Notify(610655)

	return expr
}

func (expr *OrExpr) normalize(v *NormalizeVisitor) TypedExpr {
	__antithesis_instrumentation__.Notify(610759)
	left := expr.TypedLeft()
	right := expr.TypedRight()
	var dleft, dright Datum

	if left == DNull && func() bool {
		__antithesis_instrumentation__.Notify(610763)
		return right == DNull == true
	}() == true {
		__antithesis_instrumentation__.Notify(610764)
		return DNull
	} else {
		__antithesis_instrumentation__.Notify(610765)
	}
	__antithesis_instrumentation__.Notify(610760)

	if v.isConst(left) {
		__antithesis_instrumentation__.Notify(610766)
		dleft, v.err = left.Eval(v.ctx)
		if v.err != nil {
			__antithesis_instrumentation__.Notify(610769)
			return expr
		} else {
			__antithesis_instrumentation__.Notify(610770)
		}
		__antithesis_instrumentation__.Notify(610767)
		if dleft != DNull {
			__antithesis_instrumentation__.Notify(610771)
			if d, err := GetBool(dleft); err == nil {
				__antithesis_instrumentation__.Notify(610773)
				if d {
					__antithesis_instrumentation__.Notify(610775)
					return dleft
				} else {
					__antithesis_instrumentation__.Notify(610776)
				}
				__antithesis_instrumentation__.Notify(610774)
				return right
			} else {
				__antithesis_instrumentation__.Notify(610777)
			}
			__antithesis_instrumentation__.Notify(610772)
			return DNull
		} else {
			__antithesis_instrumentation__.Notify(610778)
		}
		__antithesis_instrumentation__.Notify(610768)
		return NewTypedOrExpr(
			dleft,
			right,
		)
	} else {
		__antithesis_instrumentation__.Notify(610779)
	}
	__antithesis_instrumentation__.Notify(610761)
	if v.isConst(right) {
		__antithesis_instrumentation__.Notify(610780)
		dright, v.err = right.Eval(v.ctx)
		if v.err != nil {
			__antithesis_instrumentation__.Notify(610783)
			return expr
		} else {
			__antithesis_instrumentation__.Notify(610784)
		}
		__antithesis_instrumentation__.Notify(610781)
		if dright != DNull {
			__antithesis_instrumentation__.Notify(610785)
			if d, err := GetBool(dright); err == nil {
				__antithesis_instrumentation__.Notify(610787)
				if d {
					__antithesis_instrumentation__.Notify(610789)
					return right
				} else {
					__antithesis_instrumentation__.Notify(610790)
				}
				__antithesis_instrumentation__.Notify(610788)
				return left
			} else {
				__antithesis_instrumentation__.Notify(610791)
			}
			__antithesis_instrumentation__.Notify(610786)
			return DNull
		} else {
			__antithesis_instrumentation__.Notify(610792)
		}
		__antithesis_instrumentation__.Notify(610782)
		return NewTypedOrExpr(
			left,
			dright,
		)
	} else {
		__antithesis_instrumentation__.Notify(610793)
	}
	__antithesis_instrumentation__.Notify(610762)
	return expr
}

func (expr *NotExpr) normalize(v *NormalizeVisitor) TypedExpr {
	__antithesis_instrumentation__.Notify(610794)
	inner := expr.TypedInnerExpr()
	switch t := inner.(type) {
	case *NotExpr:
		__antithesis_instrumentation__.Notify(610796)
		return t.TypedInnerExpr()
	}
	__antithesis_instrumentation__.Notify(610795)
	return expr
}

func (expr *ParenExpr) normalize(v *NormalizeVisitor) TypedExpr {
	__antithesis_instrumentation__.Notify(610797)
	return expr.TypedInnerExpr()
}

func (expr *AnnotateTypeExpr) normalize(v *NormalizeVisitor) TypedExpr {
	__antithesis_instrumentation__.Notify(610798)

	return expr.TypedInnerExpr()
}

func (expr *RangeCond) normalize(v *NormalizeVisitor) TypedExpr {
	__antithesis_instrumentation__.Notify(610799)
	leftFrom, from := expr.TypedLeftFrom(), expr.TypedFrom()
	leftTo, to := expr.TypedLeftTo(), expr.TypedTo()

	if leftTo, v.err = v.ctx.NormalizeExpr(leftTo); v.err != nil {
		__antithesis_instrumentation__.Notify(610805)
		return expr
	} else {
		__antithesis_instrumentation__.Notify(610806)
	}
	__antithesis_instrumentation__.Notify(610800)

	if (leftFrom == DNull || func() bool {
		__antithesis_instrumentation__.Notify(610807)
		return from == DNull == true
	}() == true) && func() bool {
		__antithesis_instrumentation__.Notify(610808)
		return (leftTo == DNull || func() bool {
			__antithesis_instrumentation__.Notify(610809)
			return to == DNull == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(610810)
		return DNull
	} else {
		__antithesis_instrumentation__.Notify(610811)
	}
	__antithesis_instrumentation__.Notify(610801)

	leftCmp := treecmp.GE
	rightCmp := treecmp.LE
	if expr.Not {
		__antithesis_instrumentation__.Notify(610812)
		leftCmp = treecmp.LT
		rightCmp = treecmp.GT
	} else {
		__antithesis_instrumentation__.Notify(610813)
	}
	__antithesis_instrumentation__.Notify(610802)

	transform := func(from, to TypedExpr) TypedExpr {
		__antithesis_instrumentation__.Notify(610814)
		var newLeft, newRight TypedExpr
		if from == DNull {
			__antithesis_instrumentation__.Notify(610818)
			newLeft = DNull
		} else {
			__antithesis_instrumentation__.Notify(610819)
			newLeft = NewTypedComparisonExpr(treecmp.MakeComparisonOperator(leftCmp), leftFrom, from).normalize(v)
			if v.err != nil {
				__antithesis_instrumentation__.Notify(610820)
				return expr
			} else {
				__antithesis_instrumentation__.Notify(610821)
			}
		}
		__antithesis_instrumentation__.Notify(610815)
		if to == DNull {
			__antithesis_instrumentation__.Notify(610822)
			newRight = DNull
		} else {
			__antithesis_instrumentation__.Notify(610823)
			newRight = NewTypedComparisonExpr(treecmp.MakeComparisonOperator(rightCmp), leftTo, to).normalize(v)
			if v.err != nil {
				__antithesis_instrumentation__.Notify(610824)
				return expr
			} else {
				__antithesis_instrumentation__.Notify(610825)
			}
		}
		__antithesis_instrumentation__.Notify(610816)
		if expr.Not {
			__antithesis_instrumentation__.Notify(610826)
			return NewTypedOrExpr(newLeft, newRight).normalize(v)
		} else {
			__antithesis_instrumentation__.Notify(610827)
		}
		__antithesis_instrumentation__.Notify(610817)
		return NewTypedAndExpr(newLeft, newRight).normalize(v)
	}
	__antithesis_instrumentation__.Notify(610803)

	out := transform(from, to)
	if expr.Symmetric {
		__antithesis_instrumentation__.Notify(610828)
		if expr.Not {
			__antithesis_instrumentation__.Notify(610829)

			out = NewTypedAndExpr(out, transform(to, from)).normalize(v)
		} else {
			__antithesis_instrumentation__.Notify(610830)

			out = NewTypedOrExpr(out, transform(to, from)).normalize(v)
		}
	} else {
		__antithesis_instrumentation__.Notify(610831)
	}
	__antithesis_instrumentation__.Notify(610804)
	return out
}

func (expr *Tuple) normalize(v *NormalizeVisitor) TypedExpr {
	__antithesis_instrumentation__.Notify(610832)

	isConst := true
	for _, subExpr := range expr.Exprs {
		__antithesis_instrumentation__.Notify(610836)
		if !v.isConst(subExpr) {
			__antithesis_instrumentation__.Notify(610837)
			isConst = false
			break
		} else {
			__antithesis_instrumentation__.Notify(610838)
		}
	}
	__antithesis_instrumentation__.Notify(610833)
	if !isConst {
		__antithesis_instrumentation__.Notify(610839)
		return expr
	} else {
		__antithesis_instrumentation__.Notify(610840)
	}
	__antithesis_instrumentation__.Notify(610834)
	e, err := expr.Eval(v.ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(610841)
		v.err = err
	} else {
		__antithesis_instrumentation__.Notify(610842)
	}
	__antithesis_instrumentation__.Notify(610835)
	return e
}

func (ctx *EvalContext) NormalizeExpr(typedExpr TypedExpr) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(610843)
	v := MakeNormalizeVisitor(ctx)
	expr, _ := WalkExpr(&v, typedExpr)
	if v.err != nil {
		__antithesis_instrumentation__.Notify(610845)
		return nil, v.err
	} else {
		__antithesis_instrumentation__.Notify(610846)
	}
	__antithesis_instrumentation__.Notify(610844)
	return expr.(TypedExpr), nil
}

type NormalizeVisitor struct {
	ctx *EvalContext
	err error

	fastIsConstVisitor fastIsConstVisitor
}

var _ Visitor = &NormalizeVisitor{}

func MakeNormalizeVisitor(ctx *EvalContext) NormalizeVisitor {
	__antithesis_instrumentation__.Notify(610847)
	return NormalizeVisitor{ctx: ctx, fastIsConstVisitor: fastIsConstVisitor{ctx: ctx}}
}

func (v *NormalizeVisitor) Err() error { __antithesis_instrumentation__.Notify(610848); return v.err }

func (v *NormalizeVisitor) VisitPre(expr Expr) (recurse bool, newExpr Expr) {
	__antithesis_instrumentation__.Notify(610849)
	if v.err != nil {
		__antithesis_instrumentation__.Notify(610852)
		return false, expr
	} else {
		__antithesis_instrumentation__.Notify(610853)
	}
	__antithesis_instrumentation__.Notify(610850)

	switch expr.(type) {
	case *Subquery:
		__antithesis_instrumentation__.Notify(610854)

		return false, expr
	}
	__antithesis_instrumentation__.Notify(610851)

	return true, expr
}

func (v *NormalizeVisitor) VisitPost(expr Expr) Expr {
	__antithesis_instrumentation__.Notify(610855)
	if v.err != nil {
		__antithesis_instrumentation__.Notify(610859)
		return expr
	} else {
		__antithesis_instrumentation__.Notify(610860)
	}
	__antithesis_instrumentation__.Notify(610856)

	if normalizable, ok := expr.(normalizableExpr); ok {
		__antithesis_instrumentation__.Notify(610861)
		expr = normalizable.normalize(v)
		if v.err != nil {
			__antithesis_instrumentation__.Notify(610862)
			return expr
		} else {
			__antithesis_instrumentation__.Notify(610863)
		}
	} else {
		__antithesis_instrumentation__.Notify(610864)
	}
	__antithesis_instrumentation__.Notify(610857)

	if v.isConst(expr) {
		__antithesis_instrumentation__.Notify(610865)
		value, err := expr.(TypedExpr).Eval(v.ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(610868)

			return expr
		} else {
			__antithesis_instrumentation__.Notify(610869)
		}
		__antithesis_instrumentation__.Notify(610866)
		if value == DNull {
			__antithesis_instrumentation__.Notify(610870)

			retypedNull, ok := ReType(DNull, expr.(TypedExpr).ResolvedType())
			if !ok {
				__antithesis_instrumentation__.Notify(610872)
				v.err = errors.AssertionFailedf("failed to retype NULL to %s", expr.(TypedExpr).ResolvedType())
				return expr
			} else {
				__antithesis_instrumentation__.Notify(610873)
			}
			__antithesis_instrumentation__.Notify(610871)
			return retypedNull
		} else {
			__antithesis_instrumentation__.Notify(610874)
		}
		__antithesis_instrumentation__.Notify(610867)
		return value
	} else {
		__antithesis_instrumentation__.Notify(610875)
	}
	__antithesis_instrumentation__.Notify(610858)

	return expr
}

func (v *NormalizeVisitor) isConst(expr Expr) bool {
	__antithesis_instrumentation__.Notify(610876)
	return v.fastIsConstVisitor.run(expr)
}

func (v *NormalizeVisitor) isNumericZero(expr TypedExpr) bool {
	__antithesis_instrumentation__.Notify(610877)
	if d, ok := expr.(Datum); ok {
		__antithesis_instrumentation__.Notify(610879)
		switch t := UnwrapDatum(v.ctx, d).(type) {
		case *DDecimal:
			__antithesis_instrumentation__.Notify(610880)
			return t.Decimal.Sign() == 0
		case *DFloat:
			__antithesis_instrumentation__.Notify(610881)
			return *t == 0
		case *DInt:
			__antithesis_instrumentation__.Notify(610882)
			return *t == 0
		}
	} else {
		__antithesis_instrumentation__.Notify(610883)
	}
	__antithesis_instrumentation__.Notify(610878)
	return false
}

func (v *NormalizeVisitor) isNumericOne(expr TypedExpr) bool {
	__antithesis_instrumentation__.Notify(610884)
	if d, ok := expr.(Datum); ok {
		__antithesis_instrumentation__.Notify(610886)
		switch t := UnwrapDatum(v.ctx, d).(type) {
		case *DDecimal:
			__antithesis_instrumentation__.Notify(610887)
			return t.Decimal.Cmp(&DecimalOne.Decimal) == 0
		case *DFloat:
			__antithesis_instrumentation__.Notify(610888)
			return *t == 1.0
		case *DInt:
			__antithesis_instrumentation__.Notify(610889)
			return *t == 1
		}
	} else {
		__antithesis_instrumentation__.Notify(610890)
	}
	__antithesis_instrumentation__.Notify(610885)
	return false
}

func invertComparisonOp(op treecmp.ComparisonOperator) (treecmp.ComparisonOperator, error) {
	__antithesis_instrumentation__.Notify(610891)
	switch op.Symbol {
	case treecmp.EQ:
		__antithesis_instrumentation__.Notify(610892)
		return treecmp.MakeComparisonOperator(treecmp.EQ), nil
	case treecmp.GE:
		__antithesis_instrumentation__.Notify(610893)
		return treecmp.MakeComparisonOperator(treecmp.LE), nil
	case treecmp.GT:
		__antithesis_instrumentation__.Notify(610894)
		return treecmp.MakeComparisonOperator(treecmp.LT), nil
	case treecmp.LE:
		__antithesis_instrumentation__.Notify(610895)
		return treecmp.MakeComparisonOperator(treecmp.GE), nil
	case treecmp.LT:
		__antithesis_instrumentation__.Notify(610896)
		return treecmp.MakeComparisonOperator(treecmp.GT), nil
	default:
		__antithesis_instrumentation__.Notify(610897)
		return op, errors.AssertionFailedf("unable to invert: %s", op)
	}
}

type isConstVisitor struct {
	ctx     *EvalContext
	isConst bool
}

var _ Visitor = &isConstVisitor{}

func (v *isConstVisitor) VisitPre(expr Expr) (recurse bool, newExpr Expr) {
	__antithesis_instrumentation__.Notify(610898)
	if v.isConst {
		__antithesis_instrumentation__.Notify(610900)
		if !operatorIsImmutable(expr, v.ctx.SessionData()) || func() bool {
			__antithesis_instrumentation__.Notify(610901)
			return isVar(v.ctx, expr, true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(610902)
			v.isConst = false
			return false, expr
		} else {
			__antithesis_instrumentation__.Notify(610903)
		}
	} else {
		__antithesis_instrumentation__.Notify(610904)
	}
	__antithesis_instrumentation__.Notify(610899)
	return true, expr
}

func operatorIsImmutable(expr Expr, sd *sessiondata.SessionData) bool {
	__antithesis_instrumentation__.Notify(610905)
	switch t := expr.(type) {
	case *FuncExpr:
		__antithesis_instrumentation__.Notify(610906)
		return t.fnProps.Class == NormalClass && func() bool {
			__antithesis_instrumentation__.Notify(610912)
			return t.fn.Volatility <= VolatilityImmutable == true
		}() == true

	case *CastExpr:
		__antithesis_instrumentation__.Notify(610907)
		volatility, ok := LookupCastVolatility(t.Expr.(TypedExpr).ResolvedType(), t.typ, sd)
		return ok && func() bool {
			__antithesis_instrumentation__.Notify(610913)
			return volatility <= VolatilityImmutable == true
		}() == true

	case *UnaryExpr:
		__antithesis_instrumentation__.Notify(610908)
		return t.fn.Volatility <= VolatilityImmutable

	case *BinaryExpr:
		__antithesis_instrumentation__.Notify(610909)
		return t.Fn.Volatility <= VolatilityImmutable

	case *ComparisonExpr:
		__antithesis_instrumentation__.Notify(610910)
		return t.Fn.Volatility <= VolatilityImmutable

	default:
		__antithesis_instrumentation__.Notify(610911)
		return true
	}
}

func (*isConstVisitor) VisitPost(expr Expr) Expr {
	__antithesis_instrumentation__.Notify(610914)
	return expr
}

func (v *isConstVisitor) run(expr Expr) bool {
	__antithesis_instrumentation__.Notify(610915)
	v.isConst = true
	WalkExprConst(v, expr)
	return v.isConst
}

func IsConst(evalCtx *EvalContext, expr TypedExpr) bool {
	__antithesis_instrumentation__.Notify(610916)
	v := isConstVisitor{ctx: evalCtx}
	return v.run(expr)
}

type fastIsConstVisitor struct {
	ctx     *EvalContext
	isConst bool

	visited bool
}

var _ Visitor = &fastIsConstVisitor{}

func (v *fastIsConstVisitor) VisitPre(expr Expr) (recurse bool, newExpr Expr) {
	__antithesis_instrumentation__.Notify(610917)
	if v.visited {
		__antithesis_instrumentation__.Notify(610920)
		if _, ok := expr.(*CastExpr); ok {
			__antithesis_instrumentation__.Notify(610923)

			return true, expr
		} else {
			__antithesis_instrumentation__.Notify(610924)
		}
		__antithesis_instrumentation__.Notify(610921)
		if _, ok := expr.(Datum); !ok || func() bool {
			__antithesis_instrumentation__.Notify(610925)
			return isVar(v.ctx, expr, true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(610926)

			v.isConst = false
		} else {
			__antithesis_instrumentation__.Notify(610927)
		}
		__antithesis_instrumentation__.Notify(610922)
		return false, expr
	} else {
		__antithesis_instrumentation__.Notify(610928)
	}
	__antithesis_instrumentation__.Notify(610918)
	v.visited = true

	if !operatorIsImmutable(expr, v.ctx.SessionData()) || func() bool {
		__antithesis_instrumentation__.Notify(610929)
		return isVar(v.ctx, expr, true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(610930)
		v.isConst = false
		return false, expr
	} else {
		__antithesis_instrumentation__.Notify(610931)
	}
	__antithesis_instrumentation__.Notify(610919)

	return true, expr
}

func (*fastIsConstVisitor) VisitPost(expr Expr) Expr {
	__antithesis_instrumentation__.Notify(610932)
	return expr
}

func (v *fastIsConstVisitor) run(expr Expr) bool {
	__antithesis_instrumentation__.Notify(610933)
	v.isConst = true
	v.visited = false
	WalkExprConst(v, expr)
	return v.isConst
}

func isVar(evalCtx *EvalContext, expr Expr, allowConstPlaceholders bool) bool {
	__antithesis_instrumentation__.Notify(610934)
	switch expr.(type) {
	case VariableExpr:
		__antithesis_instrumentation__.Notify(610936)
		return true
	case *Placeholder:
		__antithesis_instrumentation__.Notify(610937)
		if allowConstPlaceholders {
			__antithesis_instrumentation__.Notify(610939)
			if evalCtx == nil || func() bool {
				__antithesis_instrumentation__.Notify(610941)
				return !evalCtx.HasPlaceholders() == true
			}() == true {
				__antithesis_instrumentation__.Notify(610942)

				return true
			} else {
				__antithesis_instrumentation__.Notify(610943)
			}
			__antithesis_instrumentation__.Notify(610940)
			return evalCtx.Placeholders.IsUnresolvedPlaceholder(expr)
		} else {
			__antithesis_instrumentation__.Notify(610944)
		}
		__antithesis_instrumentation__.Notify(610938)

		return true
	}
	__antithesis_instrumentation__.Notify(610935)
	return false
}

type containsVarsVisitor struct {
	containsVars bool
}

var _ Visitor = &containsVarsVisitor{}

func (v *containsVarsVisitor) VisitPre(expr Expr) (recurse bool, newExpr Expr) {
	__antithesis_instrumentation__.Notify(610945)
	if !v.containsVars && func() bool {
		__antithesis_instrumentation__.Notify(610948)
		return isVar(nil, expr, false) == true
	}() == true {
		__antithesis_instrumentation__.Notify(610949)
		v.containsVars = true
	} else {
		__antithesis_instrumentation__.Notify(610950)
	}
	__antithesis_instrumentation__.Notify(610946)
	if v.containsVars {
		__antithesis_instrumentation__.Notify(610951)
		return false, expr
	} else {
		__antithesis_instrumentation__.Notify(610952)
	}
	__antithesis_instrumentation__.Notify(610947)
	return true, expr
}

func (*containsVarsVisitor) VisitPost(expr Expr) Expr {
	__antithesis_instrumentation__.Notify(610953)
	return expr
}

func ContainsVars(expr Expr) bool {
	__antithesis_instrumentation__.Notify(610954)
	v := containsVarsVisitor{containsVars: false}
	WalkExprConst(&v, expr)
	return v.containsVars
}

var DecimalOne DDecimal

func init() {
	DecimalOne.SetInt64(1)
}

func ReType(expr TypedExpr, wantedType *types.T) (_ TypedExpr, ok bool) {
	__antithesis_instrumentation__.Notify(610955)
	resolvedType := expr.ResolvedType()
	if wantedType.Family() == types.AnyFamily || func() bool {
		__antithesis_instrumentation__.Notify(610958)
		return resolvedType.Identical(wantedType) == true
	}() == true {
		__antithesis_instrumentation__.Notify(610959)
		return expr, true
	} else {
		__antithesis_instrumentation__.Notify(610960)
	}
	__antithesis_instrumentation__.Notify(610956)

	if !ValidCast(resolvedType, wantedType, CastContextExplicit) {
		__antithesis_instrumentation__.Notify(610961)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(610962)
	}
	__antithesis_instrumentation__.Notify(610957)
	res := &CastExpr{Expr: expr, Type: wantedType}
	res.typ = wantedType
	return res, true
}
