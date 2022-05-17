package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

type ResolvableFunctionReference struct {
	FunctionReference
}

func (fn *ResolvableFunctionReference) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609893)
	ctx.FormatNode(fn.FunctionReference)
}
func (fn *ResolvableFunctionReference) String() string {
	__antithesis_instrumentation__.Notify(609894)
	return AsString(fn)
}

func (fn *ResolvableFunctionReference) Resolve(
	searchPath sessiondata.SearchPath,
) (*FunctionDefinition, error) {
	__antithesis_instrumentation__.Notify(609895)
	switch t := fn.FunctionReference.(type) {
	case *FunctionDefinition:
		__antithesis_instrumentation__.Notify(609896)
		return t, nil
	case *UnresolvedName:
		__antithesis_instrumentation__.Notify(609897)
		fd, err := t.ResolveFunction(searchPath)
		if err != nil {
			__antithesis_instrumentation__.Notify(609900)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(609901)
		}
		__antithesis_instrumentation__.Notify(609898)
		fn.FunctionReference = fd
		return fd, nil
	default:
		__antithesis_instrumentation__.Notify(609899)
		return nil, errors.AssertionFailedf("unknown function name type: %+v (%T)",
			fn.FunctionReference, fn.FunctionReference,
		)
	}
}

func WrapFunction(n string) ResolvableFunctionReference {
	__antithesis_instrumentation__.Notify(609902)
	fd, ok := FunDefs[n]
	if !ok {
		__antithesis_instrumentation__.Notify(609904)
		panic(errors.AssertionFailedf("function %s() not defined", redact.Safe(n)))
	} else {
		__antithesis_instrumentation__.Notify(609905)
	}
	__antithesis_instrumentation__.Notify(609903)
	return ResolvableFunctionReference{fd}
}

type FunctionReference interface {
	fmt.Stringer
	NodeFormatter
	functionReference()
}

func (*UnresolvedName) functionReference()     { __antithesis_instrumentation__.Notify(609906) }
func (*FunctionDefinition) functionReference() { __antithesis_instrumentation__.Notify(609907) }
