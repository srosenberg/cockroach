package builtins

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

var AllBuiltinNames []string

var AllAggregateBuiltinNames []string

var AllWindowBuiltinNames []string

func init() {
	initAggregateBuiltins()
	initWindowBuiltins()
	initGeneratorBuiltins()
	initGeoBuiltins()
	initPGBuiltins()
	initMathBuiltins()
	initReplicationBuiltins()

	AllBuiltinNames = make([]string, 0, len(builtins))
	AllAggregateBuiltinNames = make([]string, 0, len(aggregates))
	tree.FunDefs = make(map[string]*tree.FunctionDefinition)
	for name, def := range builtins {
		fDef := tree.NewFunctionDefinition(name, &def.props, def.overloads)
		tree.FunDefs[name] = fDef
		if !fDef.ShouldDocument() {

			continue
		}
		AllBuiltinNames = append(AllBuiltinNames, name)
		if def.props.Class == tree.AggregateClass {
			AllAggregateBuiltinNames = append(AllAggregateBuiltinNames, name)
		} else if def.props.Class == tree.WindowClass {
			AllWindowBuiltinNames = append(AllWindowBuiltinNames, name)
		}
		for _, overload := range def.overloads {
			fnCount := 0
			if overload.Fn != nil {
				fnCount++
			}
			if overload.FnWithExprs != nil {
				fnCount++
			}
			if overload.Generator != nil {
				overload.Fn = unsuitableUseOfGeneratorFn
				overload.FnWithExprs = unsuitableUseOfGeneratorFnWithExprs
				fnCount++
			}
			if overload.GeneratorWithExprs != nil {
				overload.Fn = unsuitableUseOfGeneratorFn
				overload.FnWithExprs = unsuitableUseOfGeneratorFnWithExprs
				fnCount++
			}
			if fnCount > 1 {
				panic(fmt.Sprintf(
					"builtin %s: at most 1 of Fn, FnWithExprs, Generator, and GeneratorWithExprs"+
						"must be set on overloads; (found %d)",
					name, fnCount,
				))
			}
		}
	}

	for _, name := range AllBuiltinNames {
		def := builtins[name]
		if def.props.Category == "" {
			def.props.Category = getCategory(def.overloads)
			builtins[name] = def
		}
	}

	sort.Strings(AllBuiltinNames)
	sort.Strings(AllAggregateBuiltinNames)
	sort.Strings(AllWindowBuiltinNames)
}

func getCategory(b []tree.Overload) string {
	__antithesis_instrumentation__.Notify(596937)

	for _, ovl := range b {
		__antithesis_instrumentation__.Notify(596939)
		switch typ := ovl.Types.(type) {
		case tree.ArgTypes:
			__antithesis_instrumentation__.Notify(596941)
			if len(typ) == 1 {
				__antithesis_instrumentation__.Notify(596942)
				return categorizeType(typ[0].Typ)
			} else {
				__antithesis_instrumentation__.Notify(596943)
			}
		}
		__antithesis_instrumentation__.Notify(596940)

		if retType := ovl.FixedReturnType(); retType != nil {
			__antithesis_instrumentation__.Notify(596944)
			return categorizeType(retType)
		} else {
			__antithesis_instrumentation__.Notify(596945)
		}
	}
	__antithesis_instrumentation__.Notify(596938)
	return ""
}

func collectOverloads(
	props tree.FunctionProperties, types []*types.T, gens ...func(*types.T) tree.Overload,
) builtinDefinition {
	__antithesis_instrumentation__.Notify(596946)
	r := make([]tree.Overload, 0, len(types)*len(gens))
	for _, f := range gens {
		__antithesis_instrumentation__.Notify(596948)
		for _, t := range types {
			__antithesis_instrumentation__.Notify(596949)
			r = append(r, f(t))
		}
	}
	__antithesis_instrumentation__.Notify(596947)
	return builtinDefinition{
		props:     props,
		overloads: r,
	}
}
