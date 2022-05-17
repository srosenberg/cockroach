// Package nilness inspects the control-flow graph of an SSA function
// and reports errors such as nil pointer dereferences and degenerate
// nil pointer comparisons.
package nilness

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/ssa"
)

const doc = `check for redundant or impossible nil comparisons

The nilness checker inspects the control-flow graph of each function in
a package and reports nil pointer dereferences, degenerate nil
pointers, and panics with nil values. A degenerate comparison is of the form
x==nil or x!=nil where x is statically known to be nil or non-nil. These are
often a mistake, especially in control flow related to errors. Panics with nil
values are checked because they are not detectable by

	if r := recover(); r != nil {

This check reports conditions such as:

	if f == nil { // impossible condition (f is a function)
	}

and:

	p := &v
	...
	if p != nil { // tautological condition
	}

and:

	if p == nil {
		print(*p) // nil dereference
	}

and:

	if p == nil {
		panic(p)
	}
`

type argExtractor func(c *ssa.CallCommon) ssa.Value

type analyzerConfig struct {
	reportDegenerateIfConditions bool

	argExtractors map[string]argExtractor
}

var crdbConfig = analyzerConfig{
	reportDegenerateIfConditions: false,
	argExtractors: map[string]argExtractor{
		"github.com/cockroachdb/errors.Handled":                          firstArg,
		"github.com/cockroachdb/errors.HandledWithMessage":               firstArg,
		"github.com/cockroachdb/errors.NewAssertionErrorWithWrappedErrf": firstArg,
		"github.com/cockroachdb/errors.Mark":                             firstArg,
		"github.com/cockroachdb/errors.WithContextTags":                  firstArg,
		"github.com/cockroachdb/errors.WithDetail":                       firstArg,
		"github.com/cockroachdb/errors.WithDetailf":                      firstArg,
		"github.com/cockroachdb/errors.WithHint":                         firstArg,
		"github.com/cockroachdb/errors.WithHintf":                        firstArg,
		"github.com/cockroachdb/errors.WithMessage":                      firstArg,
		"github.com/cockroachdb/errors.WithMessagef":                     firstArg,
		"github.com/cockroachdb/errors.WithStack":                        firstArg,
		"github.com/cockroachdb/errors.WithStackDepth":                   firstArg,
		"github.com/cockroachdb/errors.Wrap":                             firstArg,
		"github.com/cockroachdb/errors.Wrapf":                            firstArg,
		"github.com/cockroachdb/errors.WrapWithDepth":                    firstArg,
		"github.com/cockroachdb/errors.WrapWithDepthf":                   firstArg,
	},
}

var defaultConfig = analyzerConfig{
	reportDegenerateIfConditions: true,
}

var Analyzer = &analysis.Analyzer{
	Name:     "nilness",
	Doc:      doc,
	Run:      crdbConfig.run,
	Requires: []*analysis.Analyzer{buildssa.Analyzer},
}

var TestAnalyzer = &analysis.Analyzer{
	Name:     "nilness",
	Doc:      doc,
	Run:      defaultConfig.run,
	Requires: []*analysis.Analyzer{buildssa.Analyzer},
}

func (a *analyzerConfig) run(pass *analysis.Pass) (interface{}, error) {
	__antithesis_instrumentation__.Notify(644870)
	ssainput := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)
	for _, fn := range ssainput.SrcFuncs {
		__antithesis_instrumentation__.Notify(644872)
		a.runFunc(pass, fn)
	}
	__antithesis_instrumentation__.Notify(644871)
	return nil, nil
}

func (a *analyzerConfig) runFunc(pass *analysis.Pass, fn *ssa.Function) {
	__antithesis_instrumentation__.Notify(644873)
	reportf := func(category string, pos token.Pos, format string, args ...interface{}) {
		__antithesis_instrumentation__.Notify(644878)
		pass.Report(analysis.Diagnostic{
			Pos:      pos,
			Category: category,
			Message:  fmt.Sprintf(format, args...),
		})
	}
	__antithesis_instrumentation__.Notify(644874)

	notNil := func(stack []fact, instr ssa.Instruction, v ssa.Value, descr string) {
		__antithesis_instrumentation__.Notify(644879)
		if nilnessOf(stack, v) == isnil {
			__antithesis_instrumentation__.Notify(644880)
			reportf("nilderef", instr.Pos(), "nil dereference in "+descr)
		} else {
			__antithesis_instrumentation__.Notify(644881)
		}
	}
	__antithesis_instrumentation__.Notify(644875)

	notNilArg := func(stack []fact, instr ssa.Instruction, v ssa.Value, descr string) {
		__antithesis_instrumentation__.Notify(644882)
		if nilnessOf(stack, v) == isnil {
			__antithesis_instrumentation__.Notify(644883)
			reportf("nilarg", instr.Pos(), "nil argument to "+descr)
		} else {
			__antithesis_instrumentation__.Notify(644884)
		}
	}
	__antithesis_instrumentation__.Notify(644876)

	seen := make([]bool, len(fn.Blocks))
	var visit func(b *ssa.BasicBlock, stack []fact)
	visit = func(b *ssa.BasicBlock, stack []fact) {
		__antithesis_instrumentation__.Notify(644885)
		if seen[b.Index] {
			__antithesis_instrumentation__.Notify(644890)
			return
		} else {
			__antithesis_instrumentation__.Notify(644891)
		}
		__antithesis_instrumentation__.Notify(644886)
		seen[b.Index] = true

		for _, instr := range b.Instrs {
			__antithesis_instrumentation__.Notify(644892)
			switch instr := instr.(type) {
			case ssa.CallInstruction:
				__antithesis_instrumentation__.Notify(644893)
				call := instr.Common()
				notNil(stack, instr, call.Value, call.Description())

				switch f := call.Value.(type) {
				case *ssa.Function:
					__antithesis_instrumentation__.Notify(644901)
					if tf, ok := f.Object().(*types.Func); ok {
						__antithesis_instrumentation__.Notify(644902)
						if extract, ok := a.argExtractors[tf.FullName()]; ok {
							__antithesis_instrumentation__.Notify(644903)
							notNilArg(stack, instr, extract(call), tf.FullName())
						} else {
							__antithesis_instrumentation__.Notify(644904)
						}
					} else {
						__antithesis_instrumentation__.Notify(644905)
					}
				}
			case *ssa.FieldAddr:
				__antithesis_instrumentation__.Notify(644894)
				notNil(stack, instr, instr.X, "field selection")
			case *ssa.IndexAddr:
				__antithesis_instrumentation__.Notify(644895)
				notNil(stack, instr, instr.X, "index operation")
			case *ssa.MapUpdate:
				__antithesis_instrumentation__.Notify(644896)
				notNil(stack, instr, instr.Map, "map update")
			case *ssa.Slice:
				__antithesis_instrumentation__.Notify(644897)

				if _, ok := instr.X.Type().Underlying().(*types.Pointer); ok {
					__antithesis_instrumentation__.Notify(644906)
					notNil(stack, instr, instr.X, "slice operation")
				} else {
					__antithesis_instrumentation__.Notify(644907)
				}
			case *ssa.Store:
				__antithesis_instrumentation__.Notify(644898)
				notNil(stack, instr, instr.Addr, "store")
			case *ssa.TypeAssert:
				__antithesis_instrumentation__.Notify(644899)
				if !instr.CommaOk {
					__antithesis_instrumentation__.Notify(644908)
					notNil(stack, instr, instr.X, "type assertion")
				} else {
					__antithesis_instrumentation__.Notify(644909)
				}
			case *ssa.UnOp:
				__antithesis_instrumentation__.Notify(644900)
				if instr.Op == token.MUL {
					__antithesis_instrumentation__.Notify(644910)
					notNil(stack, instr, instr.X, "load")
				} else {
					__antithesis_instrumentation__.Notify(644911)
				}
			}
		}
		__antithesis_instrumentation__.Notify(644887)

		for _, instr := range b.Instrs {
			__antithesis_instrumentation__.Notify(644912)
			switch instr := instr.(type) {
			case *ssa.Panic:
				__antithesis_instrumentation__.Notify(644913)
				if nilnessOf(stack, instr.X) == isnil {
					__antithesis_instrumentation__.Notify(644914)
					reportf("nilpanic", instr.Pos(), "panic with nil value")
				} else {
					__antithesis_instrumentation__.Notify(644915)
				}
			}
		}
		__antithesis_instrumentation__.Notify(644888)

		if binop, tsucc, fsucc := eq(b); binop != nil {
			__antithesis_instrumentation__.Notify(644916)
			xnil := nilnessOf(stack, binop.X)
			ynil := nilnessOf(stack, binop.Y)

			if ynil != unknown && func() bool {
				__antithesis_instrumentation__.Notify(644918)
				return xnil != unknown == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(644919)
				return (xnil == isnil || func() bool {
					__antithesis_instrumentation__.Notify(644920)
					return ynil == isnil == true
				}() == true) == true
			}() == true {
				__antithesis_instrumentation__.Notify(644921)

				if a.reportDegenerateIfConditions {
					__antithesis_instrumentation__.Notify(644925)
					var adj string
					if (xnil == ynil) == (binop.Op == token.EQL) {
						__antithesis_instrumentation__.Notify(644927)
						adj = "tautological"
					} else {
						__antithesis_instrumentation__.Notify(644928)
						adj = "impossible"
					}
					__antithesis_instrumentation__.Notify(644926)
					reportf("cond", binop.Pos(), "%s condition: %s %s %s", adj, xnil, binop.Op, ynil)
				} else {
					__antithesis_instrumentation__.Notify(644929)
				}
				__antithesis_instrumentation__.Notify(644922)

				var skip *ssa.BasicBlock
				if xnil == ynil {
					__antithesis_instrumentation__.Notify(644930)
					skip = fsucc
				} else {
					__antithesis_instrumentation__.Notify(644931)
					skip = tsucc
				}
				__antithesis_instrumentation__.Notify(644923)
				for _, d := range b.Dominees() {
					__antithesis_instrumentation__.Notify(644932)
					if d == skip && func() bool {
						__antithesis_instrumentation__.Notify(644934)
						return len(d.Preds) == 1 == true
					}() == true {
						__antithesis_instrumentation__.Notify(644935)
						continue
					} else {
						__antithesis_instrumentation__.Notify(644936)
					}
					__antithesis_instrumentation__.Notify(644933)
					visit(d, stack)
				}
				__antithesis_instrumentation__.Notify(644924)
				return
			} else {
				__antithesis_instrumentation__.Notify(644937)
			}
			__antithesis_instrumentation__.Notify(644917)

			if xnil == isnil || func() bool {
				__antithesis_instrumentation__.Notify(644938)
				return ynil == isnil == true
			}() == true {
				__antithesis_instrumentation__.Notify(644939)
				var newFacts facts
				if xnil == isnil {
					__antithesis_instrumentation__.Notify(644942)

					newFacts = expandFacts(fact{binop.Y, isnil})
				} else {
					__antithesis_instrumentation__.Notify(644943)

					newFacts = expandFacts(fact{binop.X, isnil})
				}
				__antithesis_instrumentation__.Notify(644940)

				for _, d := range b.Dominees() {
					__antithesis_instrumentation__.Notify(644944)

					s := stack
					if len(d.Preds) == 1 {
						__antithesis_instrumentation__.Notify(644946)
						if d == tsucc {
							__antithesis_instrumentation__.Notify(644947)
							s = append(s, newFacts...)
						} else {
							__antithesis_instrumentation__.Notify(644948)
							if d == fsucc {
								__antithesis_instrumentation__.Notify(644949)
								s = append(s, newFacts.negate()...)
							} else {
								__antithesis_instrumentation__.Notify(644950)
							}
						}
					} else {
						__antithesis_instrumentation__.Notify(644951)
					}
					__antithesis_instrumentation__.Notify(644945)
					visit(d, s)
				}
				__antithesis_instrumentation__.Notify(644941)
				return
			} else {
				__antithesis_instrumentation__.Notify(644952)
			}
		} else {
			__antithesis_instrumentation__.Notify(644953)
		}
		__antithesis_instrumentation__.Notify(644889)

		for _, d := range b.Dominees() {
			__antithesis_instrumentation__.Notify(644954)
			visit(d, stack)
		}
	}
	__antithesis_instrumentation__.Notify(644877)

	if fn.Blocks != nil {
		__antithesis_instrumentation__.Notify(644955)
		visit(fn.Blocks[0], make([]fact, 0, 20))
	} else {
		__antithesis_instrumentation__.Notify(644956)
	}
}

type fact struct {
	value   ssa.Value
	nilness nilness
}

func (f fact) negate() fact {
	__antithesis_instrumentation__.Notify(644957)
	return fact{f.value, -f.nilness}
}

type nilness int

const (
	isnonnil         = -1
	unknown  nilness = 0
	isnil            = 1
)

var nilnessStrings = []string{"non-nil", "unknown", "nil"}

func (n nilness) String() string {
	__antithesis_instrumentation__.Notify(644958)
	return nilnessStrings[n+1]
}

func nilnessOf(stack []fact, v ssa.Value) nilness {
	__antithesis_instrumentation__.Notify(644959)
	switch v := v.(type) {

	case *ssa.ChangeInterface:
		__antithesis_instrumentation__.Notify(644963)
		if underlying := nilnessOf(stack, v.X); underlying != unknown {
			__antithesis_instrumentation__.Notify(644964)
			return underlying
		} else {
			__antithesis_instrumentation__.Notify(644965)
		}
	}
	__antithesis_instrumentation__.Notify(644960)

	switch v := v.(type) {
	case *ssa.Alloc,
		*ssa.FieldAddr,
		*ssa.FreeVar,
		*ssa.Function,
		*ssa.Global,
		*ssa.IndexAddr,
		*ssa.MakeChan,
		*ssa.MakeClosure,
		*ssa.MakeInterface,
		*ssa.MakeMap,
		*ssa.MakeSlice:
		__antithesis_instrumentation__.Notify(644966)
		return isnonnil
	case *ssa.Const:
		__antithesis_instrumentation__.Notify(644967)
		if v.IsNil() {
			__antithesis_instrumentation__.Notify(644969)
			return isnil
		} else {
			__antithesis_instrumentation__.Notify(644970)
		}
		__antithesis_instrumentation__.Notify(644968)
		return isnonnil
	}
	__antithesis_instrumentation__.Notify(644961)

	for _, f := range stack {
		__antithesis_instrumentation__.Notify(644971)
		if f.value == v {
			__antithesis_instrumentation__.Notify(644972)
			return f.nilness
		} else {
			__antithesis_instrumentation__.Notify(644973)
		}
	}
	__antithesis_instrumentation__.Notify(644962)
	return unknown
}

func eq(b *ssa.BasicBlock) (op *ssa.BinOp, tsucc, fsucc *ssa.BasicBlock) {
	__antithesis_instrumentation__.Notify(644974)
	if If, ok := b.Instrs[len(b.Instrs)-1].(*ssa.If); ok {
		__antithesis_instrumentation__.Notify(644976)
		if binop, ok := If.Cond.(*ssa.BinOp); ok {
			__antithesis_instrumentation__.Notify(644977)
			switch binop.Op {
			case token.EQL:
				__antithesis_instrumentation__.Notify(644978)
				return binop, b.Succs[0], b.Succs[1]
			case token.NEQ:
				__antithesis_instrumentation__.Notify(644979)
				return binop, b.Succs[1], b.Succs[0]
			default:
				__antithesis_instrumentation__.Notify(644980)
			}
		} else {
			__antithesis_instrumentation__.Notify(644981)
		}
	} else {
		__antithesis_instrumentation__.Notify(644982)
	}
	__antithesis_instrumentation__.Notify(644975)
	return nil, nil, nil
}

func expandFacts(f fact) []fact {
	__antithesis_instrumentation__.Notify(644983)
	ff := []fact{f}

Loop:
	for {
		__antithesis_instrumentation__.Notify(644985)
		switch v := f.value.(type) {
		case *ssa.ChangeInterface:
			__antithesis_instrumentation__.Notify(644986)
			f = fact{v.X, f.nilness}
			ff = append(ff, f)
		default:
			__antithesis_instrumentation__.Notify(644987)
			break Loop
		}
	}
	__antithesis_instrumentation__.Notify(644984)

	return ff
}

type facts []fact

func (ff facts) negate() facts {
	__antithesis_instrumentation__.Notify(644988)
	nn := make([]fact, len(ff))
	for i, f := range ff {
		__antithesis_instrumentation__.Notify(644990)
		nn[i] = f.negate()
	}
	__antithesis_instrumentation__.Notify(644989)
	return nn
}

var firstArg = func(c *ssa.CallCommon) ssa.Value { __antithesis_instrumentation__.Notify(644991); return c.Args[0] }
