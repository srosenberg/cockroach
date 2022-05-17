package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/lib/pq/oid"
)

type FunctionDefinition struct {
	Name string

	Definition []overloadImpl

	FunctionProperties
}

type FunctionProperties struct {
	UnsupportedWithIssue int

	Undocumented bool

	Private bool

	NullableArgs bool

	DistsqlBlocklist bool

	Class FunctionClass

	Category string

	AvailableOnPublicSchema bool

	ReturnLabels []string

	AmbiguousReturnType bool

	HasSequenceArguments bool

	CompositeInsensitive bool
}

func (fp *FunctionProperties) ShouldDocument() bool {
	__antithesis_instrumentation__.Notify(609883)
	return !(fp.Undocumented || func() bool {
		__antithesis_instrumentation__.Notify(609884)
		return fp.Private == true
	}() == true)
}

type FunctionClass int

const (
	NormalClass FunctionClass = iota

	AggregateClass

	WindowClass

	GeneratorClass

	SQLClass
)

var _ = NormalClass

func NewFunctionDefinition(
	name string, props *FunctionProperties, def []Overload,
) *FunctionDefinition {
	__antithesis_instrumentation__.Notify(609885)
	overloads := make([]overloadImpl, len(def))

	for i := range def {
		__antithesis_instrumentation__.Notify(609887)
		if def[i].PreferredOverload {
			__antithesis_instrumentation__.Notify(609889)

			props.AmbiguousReturnType = true
		} else {
			__antithesis_instrumentation__.Notify(609890)
		}
		__antithesis_instrumentation__.Notify(609888)

		def[i].counter = sqltelemetry.BuiltinCounter(name, def[i].Signature(false))

		overloads[i] = &def[i]
	}
	__antithesis_instrumentation__.Notify(609886)
	return &FunctionDefinition{
		Name:               name,
		Definition:         overloads,
		FunctionProperties: *props,
	}
}

var FunDefs map[string]*FunctionDefinition

var OidToBuiltinName map[oid.Oid]string

func (fd *FunctionDefinition) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609891)
	ctx.WriteString(fd.Name)
}
func (fd *FunctionDefinition) String() string {
	__antithesis_instrumentation__.Notify(609892)
	return AsString(fd)
}
