// Package seqexpr provides functionality to find usages of sequences in
// expressions.
//
// The logic here would fit nicely into schemaexpr if it weren't for the
// dependency on builtins, which itself depends on schemaexpr.
package seqexpr

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"go/constant"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/builtins"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

type SeqIdentifier struct {
	SeqName string
	SeqID   int64
}

func (si *SeqIdentifier) IsByID() bool {
	__antithesis_instrumentation__.Notify(268315)
	return len(si.SeqName) == 0
}

func GetSequenceFromFunc(funcExpr *tree.FuncExpr) (*SeqIdentifier, error) {
	__antithesis_instrumentation__.Notify(268316)
	searchPath := sessiondata.SearchPath{}

	def, err := funcExpr.Func.Resolve(searchPath)
	if err != nil {
		__antithesis_instrumentation__.Notify(268319)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(268320)
	}
	__antithesis_instrumentation__.Notify(268317)

	fnProps, overloads := builtins.GetBuiltinProperties(def.Name)
	if fnProps != nil && func() bool {
		__antithesis_instrumentation__.Notify(268321)
		return fnProps.HasSequenceArguments == true
	}() == true {
		__antithesis_instrumentation__.Notify(268322)
		found := false
		for _, overload := range overloads {
			__antithesis_instrumentation__.Notify(268324)

			if len(funcExpr.Exprs) == overload.Types.Length() {
				__antithesis_instrumentation__.Notify(268325)
				found = true
				argTypes, ok := overload.Types.(tree.ArgTypes)
				if !ok {
					__antithesis_instrumentation__.Notify(268327)
					panic(pgerror.Newf(
						pgcode.InvalidFunctionDefinition,
						"%s has invalid argument types", funcExpr.Func.String(),
					))
				} else {
					__antithesis_instrumentation__.Notify(268328)
				}
				__antithesis_instrumentation__.Notify(268326)
				for i := 0; i < overload.Types.Length(); i++ {
					__antithesis_instrumentation__.Notify(268329)

					argName := argTypes[i].Name
					if argName == builtins.SequenceNameArg {
						__antithesis_instrumentation__.Notify(268330)
						arg := funcExpr.Exprs[i]
						if seqIdentifier := getSequenceIdentifier(arg); seqIdentifier != nil {
							__antithesis_instrumentation__.Notify(268331)
							return seqIdentifier, nil
						} else {
							__antithesis_instrumentation__.Notify(268332)
						}
					} else {
						__antithesis_instrumentation__.Notify(268333)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(268334)
			}
		}
		__antithesis_instrumentation__.Notify(268323)
		if !found {
			__antithesis_instrumentation__.Notify(268335)
			panic(pgerror.New(
				pgcode.DatatypeMismatch,
				"could not find matching function overload for given arguments",
			))
		} else {
			__antithesis_instrumentation__.Notify(268336)
		}
	} else {
		__antithesis_instrumentation__.Notify(268337)
	}
	__antithesis_instrumentation__.Notify(268318)
	return nil, nil
}

func getSequenceIdentifier(expr tree.Expr) *SeqIdentifier {
	__antithesis_instrumentation__.Notify(268338)
	switch a := expr.(type) {
	case *tree.DString:
		__antithesis_instrumentation__.Notify(268340)
		seqName := string(*a)
		return &SeqIdentifier{
			SeqName: seqName,
		}
	case *tree.DOid:
		__antithesis_instrumentation__.Notify(268341)
		id := int64(a.DInt)
		return &SeqIdentifier{
			SeqID: id,
		}
	case *tree.StrVal:
		__antithesis_instrumentation__.Notify(268342)
		seqName := a.RawString()
		return &SeqIdentifier{
			SeqName: seqName,
		}
	case *tree.NumVal:
		__antithesis_instrumentation__.Notify(268343)
		id, err := a.AsInt64()
		if err == nil {
			__antithesis_instrumentation__.Notify(268346)
			return &SeqIdentifier{
				SeqID: id,
			}
		} else {
			__antithesis_instrumentation__.Notify(268347)
		}
	case *tree.CastExpr:
		__antithesis_instrumentation__.Notify(268344)
		return getSequenceIdentifier(a.Expr)
	case *tree.AnnotateTypeExpr:
		__antithesis_instrumentation__.Notify(268345)
		return getSequenceIdentifier(a.Expr)
	}
	__antithesis_instrumentation__.Notify(268339)
	return nil
}

func GetUsedSequences(defaultExpr tree.Expr) ([]SeqIdentifier, error) {
	__antithesis_instrumentation__.Notify(268348)
	var seqIdentifiers []SeqIdentifier
	_, err := tree.SimpleVisit(
		defaultExpr,
		func(expr tree.Expr) (recurse bool, newExpr tree.Expr, err error) {
			__antithesis_instrumentation__.Notify(268351)
			switch t := expr.(type) {
			case *tree.FuncExpr:
				__antithesis_instrumentation__.Notify(268353)
				identifier, err := GetSequenceFromFunc(t)
				if err != nil {
					__antithesis_instrumentation__.Notify(268355)
					return false, nil, err
				} else {
					__antithesis_instrumentation__.Notify(268356)
				}
				__antithesis_instrumentation__.Notify(268354)
				if identifier != nil {
					__antithesis_instrumentation__.Notify(268357)
					seqIdentifiers = append(seqIdentifiers, *identifier)
				} else {
					__antithesis_instrumentation__.Notify(268358)
				}
			}
			__antithesis_instrumentation__.Notify(268352)
			return true, expr, nil
		},
	)
	__antithesis_instrumentation__.Notify(268349)
	if err != nil {
		__antithesis_instrumentation__.Notify(268359)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(268360)
	}
	__antithesis_instrumentation__.Notify(268350)
	return seqIdentifiers, nil
}

func ReplaceSequenceNamesWithIDs(
	defaultExpr tree.Expr, nameToID map[string]int64,
) (tree.Expr, error) {
	__antithesis_instrumentation__.Notify(268361)
	replaceFn := func(expr tree.Expr) (recurse bool, newExpr tree.Expr, err error) {
		__antithesis_instrumentation__.Notify(268363)
		switch t := expr.(type) {
		case *tree.FuncExpr:
			__antithesis_instrumentation__.Notify(268365)
			identifier, err := GetSequenceFromFunc(t)
			if err != nil {
				__antithesis_instrumentation__.Notify(268369)
				return false, nil, err
			} else {
				__antithesis_instrumentation__.Notify(268370)
			}
			__antithesis_instrumentation__.Notify(268366)
			if identifier == nil || func() bool {
				__antithesis_instrumentation__.Notify(268371)
				return identifier.IsByID() == true
			}() == true {
				__antithesis_instrumentation__.Notify(268372)
				return true, expr, nil
			} else {
				__antithesis_instrumentation__.Notify(268373)
			}
			__antithesis_instrumentation__.Notify(268367)

			id, ok := nameToID[identifier.SeqName]
			if !ok {
				__antithesis_instrumentation__.Notify(268374)
				return true, expr, nil
			} else {
				__antithesis_instrumentation__.Notify(268375)
			}
			__antithesis_instrumentation__.Notify(268368)
			return false, &tree.FuncExpr{
				Func: t.Func,
				Exprs: tree.Exprs{
					&tree.AnnotateTypeExpr{
						Type:       types.RegClass,
						SyntaxMode: tree.AnnotateShort,
						Expr:       tree.NewNumVal(constant.MakeInt64(id), "", false),
					},
				},
			}, nil
		}
		__antithesis_instrumentation__.Notify(268364)
		return true, expr, nil
	}
	__antithesis_instrumentation__.Notify(268362)

	newExpr, err := tree.SimpleVisit(defaultExpr, replaceFn)
	return newExpr, err
}
