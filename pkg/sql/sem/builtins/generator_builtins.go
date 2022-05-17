package builtins

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/protoreflect"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/arith"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil"
	"github.com/cockroachdb/cockroach/pkg/util/json"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/cockroach/pkg/util/tracing/tracingpb"
	"github.com/cockroachdb/errors"
)

var _ tree.ValueGenerator = &seriesValueGenerator{}
var _ tree.ValueGenerator = &arrayValueGenerator{}

func initGeneratorBuiltins() {
	__antithesis_instrumentation__.Notify(599362)

	for k, v := range generators {
		__antithesis_instrumentation__.Notify(599363)
		if _, exists := builtins[k]; exists {
			__antithesis_instrumentation__.Notify(599366)
			panic("duplicate builtin: " + k)
		} else {
			__antithesis_instrumentation__.Notify(599367)
		}
		__antithesis_instrumentation__.Notify(599364)

		if v.props.Class != tree.GeneratorClass {
			__antithesis_instrumentation__.Notify(599368)
			panic(errors.AssertionFailedf("generator functions should be marked with the tree.GeneratorClass "+
				"function class, found %v", v))
		} else {
			__antithesis_instrumentation__.Notify(599369)
		}
		__antithesis_instrumentation__.Notify(599365)

		builtins[k] = v
	}
}

func genProps() tree.FunctionProperties {
	__antithesis_instrumentation__.Notify(599370)
	return tree.FunctionProperties{
		Class:    tree.GeneratorClass,
		Category: categoryGenerator,
	}
}

func genPropsWithLabels(returnLabels []string) tree.FunctionProperties {
	__antithesis_instrumentation__.Notify(599371)
	return tree.FunctionProperties{
		Class:        tree.GeneratorClass,
		Category:     categoryGenerator,
		ReturnLabels: returnLabels,
	}
}

var aclexplodeGeneratorType = types.MakeLabeledTuple(
	[]*types.T{types.Oid, types.Oid, types.String, types.Bool},
	[]string{"grantor", "grantee", "privilege_type", "is_grantable"},
)

type aclexplodeGenerator struct{}

func (aclexplodeGenerator) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(599372)
	return aclexplodeGeneratorType
}
func (aclexplodeGenerator) Start(_ context.Context, _ *kv.Txn) error {
	__antithesis_instrumentation__.Notify(599373)
	return nil
}
func (aclexplodeGenerator) Close(_ context.Context) { __antithesis_instrumentation__.Notify(599374) }
func (aclexplodeGenerator) Next(_ context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(599375)
	return false, nil
}
func (aclexplodeGenerator) Values() (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(599376)
	return nil, nil
}

var generators = map[string]builtinDefinition{

	"aclexplode": makeBuiltin(genProps(),
		makeGeneratorOverload(
			tree.ArgTypes{{"aclitems", types.StringArray}},
			aclexplodeGeneratorType,
			func(ctx *tree.EvalContext, args tree.Datums) (tree.ValueGenerator, error) {
				__antithesis_instrumentation__.Notify(599377)
				return aclexplodeGenerator{}, nil
			},
			"Produces a virtual table containing aclitem stuff ("+
				"returns no rows as this feature is unsupported in CockroachDB)",
			tree.VolatilityStable,
		),
	),
	"generate_series": makeBuiltin(genProps(),

		makeGeneratorOverload(
			tree.ArgTypes{{"start", types.Int}, {"end", types.Int}},
			seriesValueGeneratorType,
			makeSeriesGenerator,
			"Produces a virtual table containing the integer values from `start` to `end`, inclusive.",
			tree.VolatilityImmutable,
		),
		makeGeneratorOverload(
			tree.ArgTypes{{"start", types.Int}, {"end", types.Int}, {"step", types.Int}},
			seriesValueGeneratorType,
			makeSeriesGenerator,
			"Produces a virtual table containing the integer values from `start` to `end`, inclusive, by increment of `step`.",
			tree.VolatilityImmutable,
		),
		makeGeneratorOverload(
			tree.ArgTypes{{"start", types.Timestamp}, {"end", types.Timestamp}, {"step", types.Interval}},
			seriesTSValueGeneratorType,
			makeTSSeriesGenerator,
			"Produces a virtual table containing the timestamp values from `start` to `end`, inclusive, by increment of `step`.",
			tree.VolatilityImmutable,
		),
		makeGeneratorOverload(
			tree.ArgTypes{{"start", types.TimestampTZ}, {"end", types.TimestampTZ}, {"step", types.Interval}},
			seriesTSTZValueGeneratorType,
			makeTSTZSeriesGenerator,
			"Produces a virtual table containing the timestampTZ values from `start` to `end`, inclusive, by increment of `step`.",
			tree.VolatilityImmutable,
		),
	),

	"crdb_internal.testing_callback": makeBuiltin(genProps(),
		makeGeneratorOverload(
			tree.ArgTypes{{"name", types.String}},
			types.Int,
			func(ctx *tree.EvalContext, args tree.Datums) (tree.ValueGenerator, error) {
				__antithesis_instrumentation__.Notify(599378)
				s, ok := tree.AsDString(args[0])
				if !ok {
					__antithesis_instrumentation__.Notify(599381)
					return nil, errors.Newf("expected string value, got %T", args[0])
				} else {
					__antithesis_instrumentation__.Notify(599382)
				}
				__antithesis_instrumentation__.Notify(599379)
				name := string(s)
				gen, ok := ctx.TestingKnobs.CallbackGenerators[name]
				if !ok {
					__antithesis_instrumentation__.Notify(599383)
					return nil, errors.Errorf("callback %q not registered", name)
				} else {
					__antithesis_instrumentation__.Notify(599384)
				}
				__antithesis_instrumentation__.Notify(599380)
				return gen, nil
			},
			"For internal CRDB testing only. "+
				"The function calls a callback identified by `name` registered with the server by "+
				"the test.",
			tree.VolatilityVolatile,
		),
	),

	"pg_get_keywords": makeBuiltin(genProps(),

		makeGeneratorOverload(
			tree.ArgTypes{},
			keywordsValueGeneratorType,
			makeKeywordsGenerator,
			"Produces a virtual table containing the keywords known to the SQL parser.",
			tree.VolatilityImmutable,
		),
	),

	"regexp_split_to_table": makeBuiltin(
		genProps(),
		makeGeneratorOverload(
			tree.ArgTypes{
				{"string", types.String},
				{"pattern", types.String},
			},
			types.String,
			makeRegexpSplitToTableGeneratorFactory(false),
			"Split string using a POSIX regular expression as the delimiter.",
			tree.VolatilityImmutable,
		),
		makeGeneratorOverload(
			tree.ArgTypes{
				{"string", types.String},
				{"pattern", types.String},
				{"flags", types.String},
			},
			types.String,
			makeRegexpSplitToTableGeneratorFactory(true),
			"Split string using a POSIX regular expression as the delimiter with flags."+regexpFlagInfo,
			tree.VolatilityImmutable,
		),
	),

	"unnest": makeBuiltin(genProps(),

		makeGeneratorOverloadWithReturnType(
			tree.ArgTypes{{"input", types.AnyArray}},
			func(args []tree.TypedExpr) *types.T {
				__antithesis_instrumentation__.Notify(599385)
				if len(args) == 0 || func() bool {
					__antithesis_instrumentation__.Notify(599387)
					return args[0].ResolvedType().Family() == types.UnknownFamily == true
				}() == true {
					__antithesis_instrumentation__.Notify(599388)
					return tree.UnknownReturnType
				} else {
					__antithesis_instrumentation__.Notify(599389)
				}
				__antithesis_instrumentation__.Notify(599386)
				return args[0].ResolvedType().ArrayContents()
			},
			makeArrayGenerator,
			"Returns the input array as a set of rows",
			tree.VolatilityImmutable,
		),
		makeGeneratorOverloadWithReturnType(
			tree.VariadicType{
				FixedTypes: []*types.T{types.AnyArray, types.AnyArray},
				VarType:    types.AnyArray,
			},

			func(args []tree.TypedExpr) *types.T {
				__antithesis_instrumentation__.Notify(599390)
				returnTypes := make([]*types.T, len(args))
				labels := make([]string, len(args))
				for i, arg := range args {
					__antithesis_instrumentation__.Notify(599392)
					if arg.ResolvedType().Family() == types.UnknownFamily {
						__antithesis_instrumentation__.Notify(599394)
						return tree.UnknownReturnType
					} else {
						__antithesis_instrumentation__.Notify(599395)
					}
					__antithesis_instrumentation__.Notify(599393)
					returnTypes[i] = arg.ResolvedType().ArrayContents()
					labels[i] = "unnest"
				}
				__antithesis_instrumentation__.Notify(599391)
				return types.MakeLabeledTuple(returnTypes, labels)
			},
			makeVariadicUnnestGenerator,
			"Returns the input arrays as a set of rows",
			tree.VolatilityImmutable,
		),
	),

	"information_schema._pg_expandarray": makeBuiltin(genProps(),
		makeGeneratorOverloadWithReturnType(
			tree.ArgTypes{{"input", types.AnyArray}},
			func(args []tree.TypedExpr) *types.T {
				__antithesis_instrumentation__.Notify(599396)
				if len(args) == 0 || func() bool {
					__antithesis_instrumentation__.Notify(599398)
					return args[0].ResolvedType().Family() == types.UnknownFamily == true
				}() == true {
					__antithesis_instrumentation__.Notify(599399)
					return tree.UnknownReturnType
				} else {
					__antithesis_instrumentation__.Notify(599400)
				}
				__antithesis_instrumentation__.Notify(599397)
				t := args[0].ResolvedType().ArrayContents()
				return types.MakeLabeledTuple([]*types.T{t, types.Int}, expandArrayValueGeneratorLabels)
			},
			makeExpandArrayGenerator,
			"Returns the input array as a set of rows with an index",
			tree.VolatilityImmutable,
		),
	),

	"crdb_internal.unary_table": makeBuiltin(genProps(),
		makeGeneratorOverload(
			tree.ArgTypes{},
			unaryValueGeneratorType,
			makeUnaryGenerator,
			"Produces a virtual table containing a single row with no values.\n\n"+
				"This function is used only by CockroachDB's developers for testing purposes.",
			tree.VolatilityVolatile,
		),
	),

	"generate_subscripts": makeBuiltin(genProps(),

		makeGeneratorOverload(
			tree.ArgTypes{{"array", types.AnyArray}},
			subscriptsValueGeneratorType,
			makeGenerateSubscriptsGenerator,
			"Returns a series comprising the given array's subscripts.",
			tree.VolatilityImmutable,
		),
		makeGeneratorOverload(
			tree.ArgTypes{{"array", types.AnyArray}, {"dim", types.Int}},
			subscriptsValueGeneratorType,
			makeGenerateSubscriptsGenerator,
			"Returns a series comprising the given array's subscripts.",
			tree.VolatilityImmutable,
		),
		makeGeneratorOverload(
			tree.ArgTypes{{"array", types.AnyArray}, {"dim", types.Int}, {"reverse", types.Bool}},
			subscriptsValueGeneratorType,
			makeGenerateSubscriptsGenerator,
			"Returns a series comprising the given array's subscripts.\n\n"+
				"When reverse is true, the series is returned in reverse order.",
			tree.VolatilityImmutable,
		),
	),

	"json_array_elements":       makeBuiltin(genPropsWithLabels(jsonArrayGeneratorLabels), jsonArrayElementsImpl),
	"jsonb_array_elements":      makeBuiltin(genPropsWithLabels(jsonArrayGeneratorLabels), jsonArrayElementsImpl),
	"json_array_elements_text":  makeBuiltin(genPropsWithLabels(jsonArrayGeneratorLabels), jsonArrayElementsTextImpl),
	"jsonb_array_elements_text": makeBuiltin(genPropsWithLabels(jsonArrayGeneratorLabels), jsonArrayElementsTextImpl),
	"json_object_keys":          makeBuiltin(genProps(), jsonObjectKeysImpl),
	"jsonb_object_keys":         makeBuiltin(genProps(), jsonObjectKeysImpl),
	"json_each":                 makeBuiltin(genPropsWithLabels(jsonEachGeneratorLabels), jsonEachImpl),
	"jsonb_each":                makeBuiltin(genPropsWithLabels(jsonEachGeneratorLabels), jsonEachImpl),
	"json_each_text":            makeBuiltin(genPropsWithLabels(jsonEachGeneratorLabels), jsonEachTextImpl),
	"jsonb_each_text":           makeBuiltin(genPropsWithLabels(jsonEachGeneratorLabels), jsonEachTextImpl),
	"json_populate_record": makeBuiltin(jsonPopulateProps, makeJSONPopulateImpl(makeJSONPopulateRecordGenerator,
		"Expands the object in from_json to a row whose columns match the record type defined by base.",
	)),
	"jsonb_populate_record": makeBuiltin(jsonPopulateProps, makeJSONPopulateImpl(makeJSONPopulateRecordGenerator,
		"Expands the object in from_json to a row whose columns match the record type defined by base.",
	)),
	"json_populate_recordset": makeBuiltin(jsonPopulateProps, makeJSONPopulateImpl(makeJSONPopulateRecordSetGenerator,
		"Expands the outermost array of objects in from_json to a set of rows whose columns match the record type defined by base")),
	"jsonb_populate_recordset": makeBuiltin(jsonPopulateProps, makeJSONPopulateImpl(makeJSONPopulateRecordSetGenerator,
		"Expands the outermost array of objects in from_json to a set of rows whose columns match the record type defined by base")),

	"crdb_internal.check_consistency": makeBuiltin(
		tree.FunctionProperties{
			Class:    tree.GeneratorClass,
			Category: categorySystemInfo,
		},
		makeGeneratorOverload(
			tree.ArgTypes{
				{Name: "stats_only", Typ: types.Bool},
				{Name: "start_key", Typ: types.Bytes},
				{Name: "end_key", Typ: types.Bytes},
			},
			checkConsistencyGeneratorType,
			makeCheckConsistencyGenerator,
			"Runs a consistency check on ranges touching the specified key range. "+
				"an empty start or end key is treated as the minimum and maximum possible, "+
				"respectively. stats_only should only be set to false when targeting a "+
				"small number of ranges to avoid overloading the cluster. Each returned row "+
				"contains the range ID, the status (a roachpb.CheckConsistencyResponse_Status), "+
				"and verbose detail.\n\n"+
				"Example usage:\n"+
				"SELECT * FROM crdb_internal.check_consistency(true, '\\x02', '\\x04')",
			tree.VolatilityVolatile,
		),
	),

	"crdb_internal.list_sql_keys_in_range": makeBuiltin(
		tree.FunctionProperties{
			Class:    tree.GeneratorClass,
			Category: categorySystemInfo,
		},
		makeGeneratorOverload(
			tree.ArgTypes{
				{Name: "range_id", Typ: types.Int},
			},
			rangeKeyIteratorType,
			makeRangeKeyIterator,
			"Returns all SQL K/V pairs within the requested range.",
			tree.VolatilityVolatile,
		),
	),

	"crdb_internal.payloads_for_span": makeBuiltin(
		tree.FunctionProperties{
			Class:    tree.GeneratorClass,
			Category: categorySystemInfo,
		},
		makeGeneratorOverload(
			tree.ArgTypes{
				{Name: "span_id", Typ: types.Int},
			},
			payloadsForSpanGeneratorType,
			makePayloadsForSpanGenerator,
			"Returns the payload(s) of the requested span and all its children.",
			tree.VolatilityVolatile,
		),
	),
	"crdb_internal.payloads_for_trace": makeBuiltin(
		tree.FunctionProperties{
			Class:    tree.GeneratorClass,
			Category: categorySystemInfo,
		},
		makeGeneratorOverload(
			tree.ArgTypes{
				{Name: "trace_id", Typ: types.Int},
			},
			payloadsForTraceGeneratorType,
			makePayloadsForTraceGenerator,
			"Returns the payload(s) of the requested trace.",
			tree.VolatilityVolatile,
		),
	),
	"crdb_internal.show_create_all_schemas": makeBuiltin(
		tree.FunctionProperties{
			Class: tree.GeneratorClass,
		},
		makeGeneratorOverload(
			tree.ArgTypes{
				{"database_name", types.String},
			},
			showCreateAllSchemasGeneratorType,
			makeShowCreateAllSchemasGenerator,
			`Returns rows of CREATE schema statements. 
The output can be used to recreate a database.'
`,
			tree.VolatilityVolatile,
		),
	),
	"crdb_internal.show_create_all_tables": makeBuiltin(
		tree.FunctionProperties{
			Class: tree.GeneratorClass,
		},
		makeGeneratorOverload(
			tree.ArgTypes{
				{"database_name", types.String},
			},
			showCreateAllTablesGeneratorType,
			makeShowCreateAllTablesGenerator,
			`Returns rows of CREATE table statements followed by 
ALTER table statements that add table constraints. The rows are ordered
by dependencies. All foreign keys are added after the creation of the table
in the alter statements.
It is not recommended to perform this operation on a database with many 
tables.
The output can be used to recreate a database.'
`,
			tree.VolatilityVolatile,
		),
	),
	"crdb_internal.show_create_all_types": makeBuiltin(
		tree.FunctionProperties{
			Class: tree.GeneratorClass,
		},
		makeGeneratorOverload(
			tree.ArgTypes{
				{"database_name", types.String},
			},
			showCreateAllTypesGeneratorType,
			makeShowCreateAllTypesGenerator,
			`Returns rows of CREATE type statements. 
The output can be used to recreate a database.'
`,
			tree.VolatilityVolatile,
		),
	),
	"crdb_internal.decode_plan_gist": makeBuiltin(
		tree.FunctionProperties{
			Class: tree.GeneratorClass,
		},
		makeGeneratorOverload(
			tree.ArgTypes{
				{"gist", types.String},
			},
			decodePlanGistGeneratorType,
			makeDecodePlanGistGenerator,
			`Returns rows of output similar to EXPLAIN from a gist such as those found in planGists element of the statistics column of the statement_statistics table.
			`,
			tree.VolatilityVolatile,
		),
	),
}

var decodePlanGistGeneratorType = types.String

type gistPlanGenerator struct {
	gist  string
	index int
	rows  []string
	p     tree.EvalPlanner
}

var _ tree.ValueGenerator = &gistPlanGenerator{}

func (g *gistPlanGenerator) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(599401)
	return types.String
}

func (g *gistPlanGenerator) Start(_ context.Context, _ *kv.Txn) error {
	__antithesis_instrumentation__.Notify(599402)
	rows, err := g.p.DecodeGist(g.gist)
	if err != nil {
		__antithesis_instrumentation__.Notify(599404)
		return err
	} else {
		__antithesis_instrumentation__.Notify(599405)
	}
	__antithesis_instrumentation__.Notify(599403)
	g.rows = rows
	g.index = -1
	return nil
}

func (g *gistPlanGenerator) Next(context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(599406)
	g.index++
	return g.index < len(g.rows), nil
}

func (g *gistPlanGenerator) Close(context.Context) { __antithesis_instrumentation__.Notify(599407) }

func (g *gistPlanGenerator) Values() (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(599408)
	return tree.Datums{tree.NewDString(g.rows[g.index])}, nil
}

func makeDecodePlanGistGenerator(
	ctx *tree.EvalContext, args tree.Datums,
) (tree.ValueGenerator, error) {
	__antithesis_instrumentation__.Notify(599409)
	gist := string(tree.MustBeDString(args[0]))
	return &gistPlanGenerator{gist: gist, p: ctx.Planner}, nil
}

func makeGeneratorOverload(
	in tree.TypeList, ret *types.T, g tree.GeneratorFactory, info string, volatility tree.Volatility,
) tree.Overload {
	__antithesis_instrumentation__.Notify(599410)
	return makeGeneratorOverloadWithReturnType(in, tree.FixedReturnType(ret), g, info, volatility)
}

var unsuitableUseOfGeneratorFn = func(_ *tree.EvalContext, _ tree.Datums) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(599411)
	return nil, errors.AssertionFailedf("generator functions cannot be evaluated as scalars")
}

var unsuitableUseOfGeneratorFnWithExprs = func(_ *tree.EvalContext, _ tree.Exprs) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(599412)
	return nil, errors.AssertionFailedf("generator functions cannot be evaluated as scalars")
}

func makeGeneratorOverloadWithReturnType(
	in tree.TypeList,
	retType tree.ReturnTyper,
	g tree.GeneratorFactory,
	info string,
	volatility tree.Volatility,
) tree.Overload {
	__antithesis_instrumentation__.Notify(599413)
	return tree.Overload{
		Types:      in,
		ReturnType: retType,
		Generator:  g,
		Info:       info,
		Volatility: volatility,
	}
}

type regexpSplitToTableGenerator struct {
	words []string
	curr  int
}

func makeRegexpSplitToTableGeneratorFactory(hasFlags bool) tree.GeneratorFactory {
	__antithesis_instrumentation__.Notify(599414)
	return func(
		ctx *tree.EvalContext, args tree.Datums,
	) (tree.ValueGenerator, error) {
		__antithesis_instrumentation__.Notify(599415)
		words, err := regexpSplit(ctx, args, hasFlags)
		if err != nil {
			__antithesis_instrumentation__.Notify(599417)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(599418)
		}
		__antithesis_instrumentation__.Notify(599416)
		return &regexpSplitToTableGenerator{
			words: words,
			curr:  -1,
		}, nil
	}
}

func (*regexpSplitToTableGenerator) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(599419)
	return types.String
}

func (*regexpSplitToTableGenerator) Close(_ context.Context) {
	__antithesis_instrumentation__.Notify(599420)
}

func (g *regexpSplitToTableGenerator) Start(_ context.Context, _ *kv.Txn) error {
	__antithesis_instrumentation__.Notify(599421)
	g.curr = -1
	return nil
}

func (g *regexpSplitToTableGenerator) Next(_ context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(599422)
	g.curr++
	return g.curr < len(g.words), nil
}

func (g *regexpSplitToTableGenerator) Values() (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(599423)
	return tree.Datums{tree.NewDString(g.words[g.curr])}, nil
}

type keywordsValueGenerator struct {
	curKeyword int
}

var keywordsValueGeneratorType = types.MakeLabeledTuple(
	[]*types.T{types.String, types.String, types.String},
	[]string{"word", "catcode", "catdesc"},
)

func makeKeywordsGenerator(_ *tree.EvalContext, _ tree.Datums) (tree.ValueGenerator, error) {
	__antithesis_instrumentation__.Notify(599424)
	return &keywordsValueGenerator{}, nil
}

func (*keywordsValueGenerator) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(599425)
	return keywordsValueGeneratorType
}

func (*keywordsValueGenerator) Close(_ context.Context) {
	__antithesis_instrumentation__.Notify(599426)
}

func (k *keywordsValueGenerator) Start(_ context.Context, _ *kv.Txn) error {
	__antithesis_instrumentation__.Notify(599427)
	k.curKeyword = -1
	return nil
}

func (k *keywordsValueGenerator) Next(_ context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(599428)
	k.curKeyword++
	return k.curKeyword < len(lexbase.KeywordNames), nil
}

func (k *keywordsValueGenerator) Values() (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(599429)
	kw := lexbase.KeywordNames[k.curKeyword]
	cat := lexbase.KeywordsCategories[kw]
	desc := keywordCategoryDescriptions[cat]
	return tree.Datums{tree.NewDString(kw), tree.NewDString(cat), tree.NewDString(desc)}, nil
}

var keywordCategoryDescriptions = map[string]string{
	"R": "reserved",
	"C": "unreserved (cannot be function or type name)",
	"T": "reserved (can be function or type name)",
	"U": "unreserved",
}

type seriesValueGenerator struct {
	origStart, value, start, stop, step interface{}
	nextOK                              bool
	genType                             *types.T
	next                                func(*seriesValueGenerator) (bool, error)
	genValue                            func(*seriesValueGenerator) (tree.Datums, error)
}

var seriesValueGeneratorType = types.Int

var seriesTSValueGeneratorType = types.Timestamp

var seriesTSTZValueGeneratorType = types.TimestampTZ

var errStepCannotBeZero = pgerror.New(pgcode.InvalidParameterValue, "step cannot be 0")

func seriesIntNext(s *seriesValueGenerator) (bool, error) {
	__antithesis_instrumentation__.Notify(599430)
	step := s.step.(int64)
	start := s.start.(int64)
	stop := s.stop.(int64)

	if !s.nextOK {
		__antithesis_instrumentation__.Notify(599434)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(599435)
	}
	__antithesis_instrumentation__.Notify(599431)
	if step < 0 && func() bool {
		__antithesis_instrumentation__.Notify(599436)
		return (start < stop) == true
	}() == true {
		__antithesis_instrumentation__.Notify(599437)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(599438)
	}
	__antithesis_instrumentation__.Notify(599432)
	if step > 0 && func() bool {
		__antithesis_instrumentation__.Notify(599439)
		return (stop < start) == true
	}() == true {
		__antithesis_instrumentation__.Notify(599440)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(599441)
	}
	__antithesis_instrumentation__.Notify(599433)
	s.value = start
	s.start, s.nextOK = arith.AddWithOverflow(start, step)
	return true, nil
}

func seriesGenIntValue(s *seriesValueGenerator) (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(599442)
	return tree.Datums{tree.NewDInt(tree.DInt(s.value.(int64)))}, nil
}

func seriesTSNext(s *seriesValueGenerator) (bool, error) {
	__antithesis_instrumentation__.Notify(599443)
	step := s.step.(duration.Duration)
	start := s.start.(time.Time)
	stop := s.stop.(time.Time)

	if !s.nextOK {
		__antithesis_instrumentation__.Notify(599447)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(599448)
	}
	__antithesis_instrumentation__.Notify(599444)

	stepForward := step.Compare(duration.Duration{}) > 0
	if !stepForward && func() bool {
		__antithesis_instrumentation__.Notify(599449)
		return (start.Before(stop)) == true
	}() == true {
		__antithesis_instrumentation__.Notify(599450)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(599451)
	}
	__antithesis_instrumentation__.Notify(599445)
	if stepForward && func() bool {
		__antithesis_instrumentation__.Notify(599452)
		return (stop.Before(start)) == true
	}() == true {
		__antithesis_instrumentation__.Notify(599453)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(599454)
	}
	__antithesis_instrumentation__.Notify(599446)

	s.value = start
	s.start = duration.Add(start, step)
	return true, nil
}

func seriesGenTSValue(s *seriesValueGenerator) (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(599455)
	ts, err := tree.MakeDTimestamp(s.value.(time.Time), time.Microsecond)
	if err != nil {
		__antithesis_instrumentation__.Notify(599457)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(599458)
	}
	__antithesis_instrumentation__.Notify(599456)
	return tree.Datums{ts}, nil
}

func seriesGenTSTZValue(s *seriesValueGenerator) (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(599459)
	ts, err := tree.MakeDTimestampTZ(s.value.(time.Time), time.Microsecond)
	if err != nil {
		__antithesis_instrumentation__.Notify(599461)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(599462)
	}
	__antithesis_instrumentation__.Notify(599460)
	return tree.Datums{ts}, nil
}

func makeSeriesGenerator(_ *tree.EvalContext, args tree.Datums) (tree.ValueGenerator, error) {
	__antithesis_instrumentation__.Notify(599463)
	start := int64(tree.MustBeDInt(args[0]))
	stop := int64(tree.MustBeDInt(args[1]))
	step := int64(1)
	if len(args) > 2 {
		__antithesis_instrumentation__.Notify(599466)
		step = int64(tree.MustBeDInt(args[2]))
	} else {
		__antithesis_instrumentation__.Notify(599467)
	}
	__antithesis_instrumentation__.Notify(599464)
	if step == 0 {
		__antithesis_instrumentation__.Notify(599468)
		return nil, errStepCannotBeZero
	} else {
		__antithesis_instrumentation__.Notify(599469)
	}
	__antithesis_instrumentation__.Notify(599465)
	return &seriesValueGenerator{
		origStart: start,
		stop:      stop,
		step:      step,
		genType:   seriesValueGeneratorType,
		genValue:  seriesGenIntValue,
		next:      seriesIntNext,
	}, nil
}

func makeTSSeriesGenerator(_ *tree.EvalContext, args tree.Datums) (tree.ValueGenerator, error) {
	__antithesis_instrumentation__.Notify(599470)
	start := args[0].(*tree.DTimestamp).Time
	stop := args[1].(*tree.DTimestamp).Time
	step := args[2].(*tree.DInterval).Duration

	if step.Compare(duration.Duration{}) == 0 {
		__antithesis_instrumentation__.Notify(599472)
		return nil, errStepCannotBeZero
	} else {
		__antithesis_instrumentation__.Notify(599473)
	}
	__antithesis_instrumentation__.Notify(599471)

	return &seriesValueGenerator{
		origStart: start,
		stop:      stop,
		step:      step,
		genType:   seriesTSValueGeneratorType,
		genValue:  seriesGenTSValue,
		next:      seriesTSNext,
	}, nil
}

func makeTSTZSeriesGenerator(_ *tree.EvalContext, args tree.Datums) (tree.ValueGenerator, error) {
	__antithesis_instrumentation__.Notify(599474)
	start := args[0].(*tree.DTimestampTZ).Time
	stop := args[1].(*tree.DTimestampTZ).Time
	step := args[2].(*tree.DInterval).Duration

	if step.Compare(duration.Duration{}) == 0 {
		__antithesis_instrumentation__.Notify(599476)
		return nil, errStepCannotBeZero
	} else {
		__antithesis_instrumentation__.Notify(599477)
	}
	__antithesis_instrumentation__.Notify(599475)

	return &seriesValueGenerator{
		origStart: start,
		stop:      stop,
		step:      step,
		genType:   seriesTSTZValueGeneratorType,
		genValue:  seriesGenTSTZValue,
		next:      seriesTSNext,
	}, nil
}

func (s *seriesValueGenerator) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(599478)
	return s.genType
}

func (s *seriesValueGenerator) Start(_ context.Context, _ *kv.Txn) error {
	__antithesis_instrumentation__.Notify(599479)
	s.nextOK = true
	s.start = s.origStart
	s.value = s.origStart
	return nil
}

func (s *seriesValueGenerator) Close(_ context.Context) {
	__antithesis_instrumentation__.Notify(599480)
}

func (s *seriesValueGenerator) Next(_ context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(599481)
	return s.next(s)
}

func (s *seriesValueGenerator) Values() (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(599482)
	return s.genValue(s)
}

func makeVariadicUnnestGenerator(
	_ *tree.EvalContext, args tree.Datums,
) (tree.ValueGenerator, error) {
	__antithesis_instrumentation__.Notify(599483)
	var arrays []*tree.DArray
	for _, a := range args {
		__antithesis_instrumentation__.Notify(599485)
		arrays = append(arrays, tree.MustBeDArray(a))
	}
	__antithesis_instrumentation__.Notify(599484)
	g := &multipleArrayValueGenerator{arrays: arrays}
	return g, nil
}

type multipleArrayValueGenerator struct {
	arrays    []*tree.DArray
	nextIndex int
	datums    tree.Datums
}

func (s *multipleArrayValueGenerator) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(599486)
	arraysN := len(s.arrays)
	returnTypes := make([]*types.T, arraysN)
	labels := make([]string, arraysN)
	for i, arr := range s.arrays {
		__antithesis_instrumentation__.Notify(599488)
		returnTypes[i] = arr.ParamTyp
		labels[i] = "unnest"
	}
	__antithesis_instrumentation__.Notify(599487)
	return types.MakeLabeledTuple(returnTypes, labels)
}

func (s *multipleArrayValueGenerator) Start(_ context.Context, _ *kv.Txn) error {
	__antithesis_instrumentation__.Notify(599489)
	s.datums = make(tree.Datums, len(s.arrays))
	s.nextIndex = -1
	return nil
}

func (s *multipleArrayValueGenerator) Close(_ context.Context) {
	__antithesis_instrumentation__.Notify(599490)
}

func (s *multipleArrayValueGenerator) Next(_ context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(599491)
	s.nextIndex++
	for _, arr := range s.arrays {
		__antithesis_instrumentation__.Notify(599493)
		if s.nextIndex < arr.Len() {
			__antithesis_instrumentation__.Notify(599494)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(599495)
		}
	}
	__antithesis_instrumentation__.Notify(599492)
	return false, nil
}

func (s *multipleArrayValueGenerator) Values() (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(599496)
	for i, arr := range s.arrays {
		__antithesis_instrumentation__.Notify(599498)
		if s.nextIndex < arr.Len() {
			__antithesis_instrumentation__.Notify(599499)
			s.datums[i] = arr.Array[s.nextIndex]
		} else {
			__antithesis_instrumentation__.Notify(599500)
			s.datums[i] = tree.DNull
		}
	}
	__antithesis_instrumentation__.Notify(599497)
	return s.datums, nil
}

func makeArrayGenerator(_ *tree.EvalContext, args tree.Datums) (tree.ValueGenerator, error) {
	__antithesis_instrumentation__.Notify(599501)
	arr := tree.MustBeDArray(args[0])
	return &arrayValueGenerator{array: arr}, nil
}

type arrayValueGenerator struct {
	array     *tree.DArray
	nextIndex int
}

func (s *arrayValueGenerator) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(599502)
	return s.array.ParamTyp
}

func (s *arrayValueGenerator) Start(_ context.Context, _ *kv.Txn) error {
	__antithesis_instrumentation__.Notify(599503)
	s.nextIndex = -1
	return nil
}

func (s *arrayValueGenerator) Close(_ context.Context) { __antithesis_instrumentation__.Notify(599504) }

func (s *arrayValueGenerator) Next(_ context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(599505)
	s.nextIndex++
	if s.nextIndex >= s.array.Len() {
		__antithesis_instrumentation__.Notify(599507)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(599508)
	}
	__antithesis_instrumentation__.Notify(599506)
	return true, nil
}

func (s *arrayValueGenerator) Values() (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(599509)
	return tree.Datums{s.array.Array[s.nextIndex]}, nil
}

func makeExpandArrayGenerator(
	evalCtx *tree.EvalContext, args tree.Datums,
) (tree.ValueGenerator, error) {
	__antithesis_instrumentation__.Notify(599510)
	arr := tree.MustBeDArray(args[0])
	g := &expandArrayValueGenerator{avg: arrayValueGenerator{array: arr}}
	g.buf[1] = tree.NewDInt(tree.DInt(-1))
	return g, nil
}

type expandArrayValueGenerator struct {
	avg arrayValueGenerator
	buf [2]tree.Datum
}

var expandArrayValueGeneratorLabels = []string{"x", "n"}

func (s *expandArrayValueGenerator) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(599511)
	return types.MakeLabeledTuple(
		[]*types.T{s.avg.array.ParamTyp, types.Int},
		expandArrayValueGeneratorLabels,
	)
}

func (s *expandArrayValueGenerator) Start(_ context.Context, _ *kv.Txn) error {
	__antithesis_instrumentation__.Notify(599512)
	s.avg.nextIndex = -1
	return nil
}

func (s *expandArrayValueGenerator) Close(_ context.Context) {
	__antithesis_instrumentation__.Notify(599513)
}

func (s *expandArrayValueGenerator) Next(_ context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(599514)
	s.avg.nextIndex++
	return s.avg.nextIndex < s.avg.array.Len(), nil
}

func (s *expandArrayValueGenerator) Values() (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(599515)

	s.buf[0] = s.avg.array.Array[s.avg.nextIndex]
	s.buf[1] = tree.NewDInt(tree.DInt(s.avg.nextIndex + 1))
	return s.buf[:], nil
}

func makeGenerateSubscriptsGenerator(
	evalCtx *tree.EvalContext, args tree.Datums,
) (tree.ValueGenerator, error) {
	__antithesis_instrumentation__.Notify(599516)
	var arr *tree.DArray
	dim := 1
	if len(args) > 1 {
		__antithesis_instrumentation__.Notify(599520)
		dim = int(tree.MustBeDInt(args[1]))
	} else {
		__antithesis_instrumentation__.Notify(599521)
	}
	__antithesis_instrumentation__.Notify(599517)

	if dim == 1 {
		__antithesis_instrumentation__.Notify(599522)
		arr = tree.MustBeDArray(args[0])
	} else {
		__antithesis_instrumentation__.Notify(599523)
		arr = &tree.DArray{}
	}
	__antithesis_instrumentation__.Notify(599518)
	var reverse bool
	if len(args) == 3 {
		__antithesis_instrumentation__.Notify(599524)
		reverse = bool(tree.MustBeDBool(args[2]))
	} else {
		__antithesis_instrumentation__.Notify(599525)
	}
	__antithesis_instrumentation__.Notify(599519)
	g := &subscriptsValueGenerator{
		avg:     arrayValueGenerator{array: arr},
		reverse: reverse,
	}
	g.buf[0] = tree.NewDInt(tree.DInt(-1))
	return g, nil
}

type subscriptsValueGenerator struct {
	avg arrayValueGenerator
	buf [1]tree.Datum

	firstIndex int
	reverse    bool
}

var subscriptsValueGeneratorType = types.Int

func (s *subscriptsValueGenerator) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(599526)
	return subscriptsValueGeneratorType
}

func (s *subscriptsValueGenerator) Start(_ context.Context, _ *kv.Txn) error {
	__antithesis_instrumentation__.Notify(599527)
	if s.reverse {
		__antithesis_instrumentation__.Notify(599529)
		s.avg.nextIndex = s.avg.array.Len()
	} else {
		__antithesis_instrumentation__.Notify(599530)
		s.avg.nextIndex = -1
	}
	__antithesis_instrumentation__.Notify(599528)

	s.firstIndex = s.avg.array.FirstIndex()
	return nil
}

func (s *subscriptsValueGenerator) Close(_ context.Context) {
	__antithesis_instrumentation__.Notify(599531)
}

func (s *subscriptsValueGenerator) Next(_ context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(599532)
	if s.reverse {
		__antithesis_instrumentation__.Notify(599534)
		s.avg.nextIndex--
		return s.avg.nextIndex >= 0, nil
	} else {
		__antithesis_instrumentation__.Notify(599535)
	}
	__antithesis_instrumentation__.Notify(599533)
	s.avg.nextIndex++
	return s.avg.nextIndex < s.avg.array.Len(), nil
}

func (s *subscriptsValueGenerator) Values() (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(599536)
	s.buf[0] = tree.NewDInt(tree.DInt(s.avg.nextIndex + s.firstIndex))
	return s.buf[:], nil
}

func EmptyGenerator() tree.ValueGenerator {
	__antithesis_instrumentation__.Notify(599537)
	return &arrayValueGenerator{array: tree.NewDArray(types.Any)}
}

type unaryValueGenerator struct {
	done bool
}

var unaryValueGeneratorType = types.EmptyTuple

func makeUnaryGenerator(_ *tree.EvalContext, args tree.Datums) (tree.ValueGenerator, error) {
	__antithesis_instrumentation__.Notify(599538)
	return &unaryValueGenerator{}, nil
}

func (*unaryValueGenerator) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(599539)
	return unaryValueGeneratorType
}

func (s *unaryValueGenerator) Start(_ context.Context, _ *kv.Txn) error {
	__antithesis_instrumentation__.Notify(599540)
	s.done = false
	return nil
}

func (s *unaryValueGenerator) Close(_ context.Context) { __antithesis_instrumentation__.Notify(599541) }

func (s *unaryValueGenerator) Next(_ context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(599542)
	if !s.done {
		__antithesis_instrumentation__.Notify(599544)
		s.done = true
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(599545)
	}
	__antithesis_instrumentation__.Notify(599543)
	return false, nil
}

var noDatums tree.Datums

func (s *unaryValueGenerator) Values() (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(599546)
	return noDatums, nil
}

func jsonAsText(j json.JSON) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(599547)
	text, err := j.AsText()
	if err != nil {
		__antithesis_instrumentation__.Notify(599550)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(599551)
	}
	__antithesis_instrumentation__.Notify(599548)
	if text == nil {
		__antithesis_instrumentation__.Notify(599552)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(599553)
	}
	__antithesis_instrumentation__.Notify(599549)
	return tree.NewDString(*text), nil
}

var (
	errJSONObjectKeysOnArray         = pgerror.New(pgcode.InvalidParameterValue, "cannot call json_object_keys on an array")
	errJSONObjectKeysOnScalar        = pgerror.Newf(pgcode.InvalidParameterValue, "cannot call json_object_keys on a scalar")
	errJSONDeconstructArrayAsObject  = pgerror.New(pgcode.InvalidParameterValue, "cannot deconstruct an array as an object")
	errJSONDeconstructScalarAsObject = pgerror.Newf(pgcode.InvalidParameterValue, "cannot deconstruct a scalar")
)

var jsonArrayElementsImpl = makeGeneratorOverload(
	tree.ArgTypes{{"input", types.Jsonb}},
	jsonArrayGeneratorType,
	makeJSONArrayAsJSONGenerator,
	"Expands a JSON array to a set of JSON values.",
	tree.VolatilityImmutable,
)

var jsonArrayElementsTextImpl = makeGeneratorOverload(
	tree.ArgTypes{{"input", types.Jsonb}},
	jsonArrayTextGeneratorType,
	makeJSONArrayAsTextGenerator,
	"Expands a JSON array to a set of text values.",
	tree.VolatilityImmutable,
)

var jsonArrayGeneratorLabels = []string{"value"}
var jsonArrayGeneratorType = types.Jsonb

var jsonArrayTextGeneratorType = types.String

type jsonArrayGenerator struct {
	json      tree.DJSON
	nextIndex int
	asText    bool
	buf       [1]tree.Datum
}

var errJSONCallOnNonArray = pgerror.New(pgcode.InvalidParameterValue,
	"cannot be called on a non-array")

func makeJSONArrayAsJSONGenerator(
	_ *tree.EvalContext, args tree.Datums,
) (tree.ValueGenerator, error) {
	__antithesis_instrumentation__.Notify(599554)
	return makeJSONArrayGenerator(args, false)
}

func makeJSONArrayAsTextGenerator(
	_ *tree.EvalContext, args tree.Datums,
) (tree.ValueGenerator, error) {
	__antithesis_instrumentation__.Notify(599555)
	return makeJSONArrayGenerator(args, true)
}

func makeJSONArrayGenerator(args tree.Datums, asText bool) (tree.ValueGenerator, error) {
	__antithesis_instrumentation__.Notify(599556)
	target := tree.MustBeDJSON(args[0])
	if target.Type() != json.ArrayJSONType {
		__antithesis_instrumentation__.Notify(599558)
		return nil, errJSONCallOnNonArray
	} else {
		__antithesis_instrumentation__.Notify(599559)
	}
	__antithesis_instrumentation__.Notify(599557)
	return &jsonArrayGenerator{
		json:   target,
		asText: asText,
	}, nil
}

func (g *jsonArrayGenerator) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(599560)
	if g.asText {
		__antithesis_instrumentation__.Notify(599562)
		return jsonArrayTextGeneratorType
	} else {
		__antithesis_instrumentation__.Notify(599563)
	}
	__antithesis_instrumentation__.Notify(599561)
	return jsonArrayGeneratorType
}

func (g *jsonArrayGenerator) Start(_ context.Context, _ *kv.Txn) error {
	__antithesis_instrumentation__.Notify(599564)
	g.nextIndex = -1
	g.json.JSON = g.json.JSON.MaybeDecode()
	g.buf[0] = nil
	return nil
}

func (g *jsonArrayGenerator) Close(_ context.Context) { __antithesis_instrumentation__.Notify(599565) }

func (g *jsonArrayGenerator) Next(_ context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(599566)
	g.nextIndex++
	next, err := g.json.FetchValIdx(g.nextIndex)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(599569)
		return next == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(599570)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(599571)
	}
	__antithesis_instrumentation__.Notify(599567)
	if g.asText {
		__antithesis_instrumentation__.Notify(599572)
		if g.buf[0], err = jsonAsText(next); err != nil {
			__antithesis_instrumentation__.Notify(599573)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(599574)
		}
	} else {
		__antithesis_instrumentation__.Notify(599575)
		g.buf[0] = tree.NewDJSON(next)
	}
	__antithesis_instrumentation__.Notify(599568)
	return true, nil
}

func (g *jsonArrayGenerator) Values() (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(599576)
	return g.buf[:], nil
}

var jsonObjectKeysImpl = makeGeneratorOverload(
	tree.ArgTypes{{"input", types.Jsonb}},
	jsonObjectKeysGeneratorType,
	makeJSONObjectKeysGenerator,
	"Returns sorted set of keys in the outermost JSON object.",
	tree.VolatilityImmutable,
)

var jsonObjectKeysGeneratorType = types.String

type jsonObjectKeysGenerator struct {
	iter *json.ObjectIterator
}

func makeJSONObjectKeysGenerator(
	_ *tree.EvalContext, args tree.Datums,
) (tree.ValueGenerator, error) {
	__antithesis_instrumentation__.Notify(599577)
	target := tree.MustBeDJSON(args[0])
	iter, err := target.ObjectIter()
	if err != nil {
		__antithesis_instrumentation__.Notify(599580)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(599581)
	}
	__antithesis_instrumentation__.Notify(599578)
	if iter == nil {
		__antithesis_instrumentation__.Notify(599582)
		switch target.Type() {
		case json.ArrayJSONType:
			__antithesis_instrumentation__.Notify(599583)
			return nil, errJSONObjectKeysOnArray
		default:
			__antithesis_instrumentation__.Notify(599584)
			return nil, errJSONObjectKeysOnScalar
		}
	} else {
		__antithesis_instrumentation__.Notify(599585)
	}
	__antithesis_instrumentation__.Notify(599579)
	return &jsonObjectKeysGenerator{
		iter: iter,
	}, nil
}

func (g *jsonObjectKeysGenerator) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(599586)
	return jsonObjectKeysGeneratorType
}

func (g *jsonObjectKeysGenerator) Start(_ context.Context, _ *kv.Txn) error {
	__antithesis_instrumentation__.Notify(599587)
	return nil
}

func (g *jsonObjectKeysGenerator) Close(_ context.Context) {
	__antithesis_instrumentation__.Notify(599588)
}

func (g *jsonObjectKeysGenerator) Next(_ context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(599589)
	return g.iter.Next(), nil
}

func (g *jsonObjectKeysGenerator) Values() (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(599590)
	return tree.Datums{tree.NewDString(g.iter.Key())}, nil
}

var jsonEachImpl = makeGeneratorOverload(
	tree.ArgTypes{{"input", types.Jsonb}},
	jsonEachGeneratorType,
	makeJSONEachImplGenerator,
	"Expands the outermost JSON or JSONB object into a set of key/value pairs.",
	tree.VolatilityImmutable,
)

var jsonEachTextImpl = makeGeneratorOverload(
	tree.ArgTypes{{"input", types.Jsonb}},
	jsonEachTextGeneratorType,
	makeJSONEachTextImplGenerator,
	"Expands the outermost JSON or JSONB object into a set of key/value pairs. "+
		"The returned values will be of type text.",
	tree.VolatilityImmutable,
)

var jsonEachGeneratorLabels = []string{"key", "value"}

var jsonEachGeneratorType = types.MakeLabeledTuple(
	[]*types.T{types.String, types.Jsonb},
	jsonEachGeneratorLabels,
)

var jsonEachTextGeneratorType = types.MakeLabeledTuple(
	[]*types.T{types.String, types.String},
	jsonEachGeneratorLabels,
)

type jsonEachGenerator struct {
	target tree.DJSON
	iter   *json.ObjectIterator
	key    tree.Datum
	value  tree.Datum
	asText bool
}

func makeJSONEachImplGenerator(_ *tree.EvalContext, args tree.Datums) (tree.ValueGenerator, error) {
	__antithesis_instrumentation__.Notify(599591)
	return makeJSONEachGenerator(args, false)
}

func makeJSONEachTextImplGenerator(
	_ *tree.EvalContext, args tree.Datums,
) (tree.ValueGenerator, error) {
	__antithesis_instrumentation__.Notify(599592)
	return makeJSONEachGenerator(args, true)
}

func makeJSONEachGenerator(args tree.Datums, asText bool) (tree.ValueGenerator, error) {
	__antithesis_instrumentation__.Notify(599593)
	target := tree.MustBeDJSON(args[0])
	return &jsonEachGenerator{
		target: target,
		key:    nil,
		value:  nil,
		asText: asText,
	}, nil
}

func (g *jsonEachGenerator) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(599594)
	if g.asText {
		__antithesis_instrumentation__.Notify(599596)
		return jsonEachTextGeneratorType
	} else {
		__antithesis_instrumentation__.Notify(599597)
	}
	__antithesis_instrumentation__.Notify(599595)
	return jsonEachGeneratorType
}

func (g *jsonEachGenerator) Start(_ context.Context, _ *kv.Txn) error {
	__antithesis_instrumentation__.Notify(599598)
	iter, err := g.target.ObjectIter()
	if err != nil {
		__antithesis_instrumentation__.Notify(599601)
		return err
	} else {
		__antithesis_instrumentation__.Notify(599602)
	}
	__antithesis_instrumentation__.Notify(599599)
	if iter == nil {
		__antithesis_instrumentation__.Notify(599603)
		switch g.target.Type() {
		case json.ArrayJSONType:
			__antithesis_instrumentation__.Notify(599604)
			return errJSONDeconstructArrayAsObject
		default:
			__antithesis_instrumentation__.Notify(599605)
			return errJSONDeconstructScalarAsObject
		}
	} else {
		__antithesis_instrumentation__.Notify(599606)
	}
	__antithesis_instrumentation__.Notify(599600)
	g.iter = iter
	return nil
}

func (g *jsonEachGenerator) Close(_ context.Context) { __antithesis_instrumentation__.Notify(599607) }

func (g *jsonEachGenerator) Next(_ context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(599608)
	if !g.iter.Next() {
		__antithesis_instrumentation__.Notify(599611)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(599612)
	}
	__antithesis_instrumentation__.Notify(599609)
	g.key = tree.NewDString(g.iter.Key())
	if g.asText {
		__antithesis_instrumentation__.Notify(599613)
		var err error
		if g.value, err = jsonAsText(g.iter.Value()); err != nil {
			__antithesis_instrumentation__.Notify(599614)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(599615)
		}
	} else {
		__antithesis_instrumentation__.Notify(599616)
		g.value = tree.NewDJSON(g.iter.Value())
	}
	__antithesis_instrumentation__.Notify(599610)
	return true, nil
}

func (g *jsonEachGenerator) Values() (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(599617)
	return tree.Datums{g.key, g.value}, nil
}

var jsonPopulateProps = tree.FunctionProperties{
	Class:    tree.GeneratorClass,
	Category: categoryGenerator,

	NullableArgs: true,
}

func makeJSONPopulateImpl(gen tree.GeneratorWithExprsFactory, info string) tree.Overload {
	__antithesis_instrumentation__.Notify(599618)
	return tree.Overload{

		Types:              tree.ArgTypes{{"base", types.Any}, {"from_json", types.Jsonb}},
		ReturnType:         tree.IdentityReturnType(0),
		GeneratorWithExprs: gen,
		Info:               info,
		Volatility:         tree.VolatilityStable,
	}
}

func makeJSONPopulateRecordGenerator(
	evalCtx *tree.EvalContext, args tree.Exprs,
) (tree.ValueGenerator, error) {
	__antithesis_instrumentation__.Notify(599619)
	tuple, j, err := jsonPopulateRecordEvalArgs(evalCtx, args)
	if err != nil {
		__antithesis_instrumentation__.Notify(599622)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(599623)
	}
	__antithesis_instrumentation__.Notify(599620)

	if j != nil {
		__antithesis_instrumentation__.Notify(599624)
		if j.Type() != json.ObjectJSONType {
			__antithesis_instrumentation__.Notify(599625)
			return nil, pgerror.Newf(pgcode.InvalidParameterValue, "argument of json_populate_record must be an object")
		} else {
			__antithesis_instrumentation__.Notify(599626)
		}
	} else {
		__antithesis_instrumentation__.Notify(599627)
		j = json.NewObjectBuilder(0).Build()
	}
	__antithesis_instrumentation__.Notify(599621)
	return &jsonPopulateRecordGenerator{
		evalCtx: evalCtx,
		input:   tuple,
		target:  j,
	}, nil
}

func jsonPopulateRecordEvalArgs(
	evalCtx *tree.EvalContext, args tree.Exprs,
) (tuple *tree.DTuple, jsonInputOrNil json.JSON, err error) {
	__antithesis_instrumentation__.Notify(599628)
	evalled := make(tree.Datums, len(args))
	for i := range args {
		__antithesis_instrumentation__.Notify(599633)
		var err error
		evalled[i], err = args[i].(tree.TypedExpr).Eval(evalCtx)
		if err != nil {
			__antithesis_instrumentation__.Notify(599634)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(599635)
		}
	}
	__antithesis_instrumentation__.Notify(599629)
	tupleType := args[0].(tree.TypedExpr).ResolvedType()
	if tupleType.Family() != types.TupleFamily && func() bool {
		__antithesis_instrumentation__.Notify(599636)
		return tupleType.Family() != types.UnknownFamily == true
	}() == true {
		__antithesis_instrumentation__.Notify(599637)
		return nil, nil, pgerror.New(
			pgcode.DatatypeMismatch,
			"first argument of json{b}_populate_record{set} must be a record type",
		)
	} else {
		__antithesis_instrumentation__.Notify(599638)
	}
	__antithesis_instrumentation__.Notify(599630)
	var defaultElems tree.Datums
	if evalled[0] == tree.DNull {
		__antithesis_instrumentation__.Notify(599639)
		defaultElems = make(tree.Datums, len(tupleType.TupleLabels()))
		for i := range defaultElems {
			__antithesis_instrumentation__.Notify(599640)
			defaultElems[i] = tree.DNull
		}
	} else {
		__antithesis_instrumentation__.Notify(599641)
		defaultElems = tree.MustBeDTuple(evalled[0]).D
	}
	__antithesis_instrumentation__.Notify(599631)
	var j json.JSON
	if evalled[1] != tree.DNull {
		__antithesis_instrumentation__.Notify(599642)
		j = tree.MustBeDJSON(evalled[1]).JSON
	} else {
		__antithesis_instrumentation__.Notify(599643)
	}
	__antithesis_instrumentation__.Notify(599632)
	return tree.NewDTuple(tupleType, defaultElems...), j, nil
}

type jsonPopulateRecordGenerator struct {
	input  *tree.DTuple
	target json.JSON

	wasCalled bool
	evalCtx   *tree.EvalContext
}

func (j jsonPopulateRecordGenerator) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(599644)
	return j.input.ResolvedType()
}

func (j *jsonPopulateRecordGenerator) Start(_ context.Context, _ *kv.Txn) error {
	__antithesis_instrumentation__.Notify(599645)
	return nil
}

func (j *jsonPopulateRecordGenerator) Close(_ context.Context) {
	__antithesis_instrumentation__.Notify(599646)
}

func (j *jsonPopulateRecordGenerator) Next(_ context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(599647)
	if !j.wasCalled {
		__antithesis_instrumentation__.Notify(599649)
		j.wasCalled = true
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(599650)
	}
	__antithesis_instrumentation__.Notify(599648)
	return false, nil
}

func (j jsonPopulateRecordGenerator) Values() (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(599651)
	if err := tree.PopulateRecordWithJSON(j.evalCtx, j.target, j.input.ResolvedType(), j.input); err != nil {
		__antithesis_instrumentation__.Notify(599653)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(599654)
	}
	__antithesis_instrumentation__.Notify(599652)
	return j.input.D, nil
}

func makeJSONPopulateRecordSetGenerator(
	evalCtx *tree.EvalContext, args tree.Exprs,
) (tree.ValueGenerator, error) {
	__antithesis_instrumentation__.Notify(599655)
	tuple, j, err := jsonPopulateRecordEvalArgs(evalCtx, args)
	if err != nil {
		__antithesis_instrumentation__.Notify(599658)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(599659)
	}
	__antithesis_instrumentation__.Notify(599656)

	if j != nil {
		__antithesis_instrumentation__.Notify(599660)
		if j.Type() != json.ArrayJSONType {
			__antithesis_instrumentation__.Notify(599661)
			return nil, pgerror.Newf(pgcode.InvalidParameterValue, "argument of json_populate_recordset must be an array")
		} else {
			__antithesis_instrumentation__.Notify(599662)
		}
	} else {
		__antithesis_instrumentation__.Notify(599663)
		j = json.NewArrayBuilder(0).Build()
	}
	__antithesis_instrumentation__.Notify(599657)

	return &jsonPopulateRecordSetGenerator{
		jsonPopulateRecordGenerator: jsonPopulateRecordGenerator{
			evalCtx: evalCtx,
			input:   tuple,
			target:  j,
		},
	}, nil
}

type jsonPopulateRecordSetGenerator struct {
	jsonPopulateRecordGenerator

	nextIdx int
}

func (j jsonPopulateRecordSetGenerator) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(599664)
	return j.input.ResolvedType()
}

func (j jsonPopulateRecordSetGenerator) Start(_ context.Context, _ *kv.Txn) error {
	__antithesis_instrumentation__.Notify(599665)
	return nil
}

func (j jsonPopulateRecordSetGenerator) Close(_ context.Context) {
	__antithesis_instrumentation__.Notify(599666)
}

func (j *jsonPopulateRecordSetGenerator) Next(_ context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(599667)
	if j.nextIdx >= j.target.Len() {
		__antithesis_instrumentation__.Notify(599669)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(599670)
	}
	__antithesis_instrumentation__.Notify(599668)
	j.nextIdx++
	return true, nil
}

func (j *jsonPopulateRecordSetGenerator) Values() (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(599671)
	obj, err := j.target.FetchValIdx(j.nextIdx - 1)
	if err != nil {
		__antithesis_instrumentation__.Notify(599676)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(599677)
	}
	__antithesis_instrumentation__.Notify(599672)
	if obj.Type() != json.ObjectJSONType {
		__antithesis_instrumentation__.Notify(599678)
		return nil, pgerror.Newf(pgcode.InvalidParameterValue, "argument of json_populate_recordset must be an array of objects")
	} else {
		__antithesis_instrumentation__.Notify(599679)
	}
	__antithesis_instrumentation__.Notify(599673)
	output := tree.NewDTupleWithLen(j.input.ResolvedType(), j.input.D.Len())
	for i := range j.input.D {
		__antithesis_instrumentation__.Notify(599680)
		output.D[i] = j.input.D[i]
	}
	__antithesis_instrumentation__.Notify(599674)
	if err := tree.PopulateRecordWithJSON(j.evalCtx, obj, j.input.ResolvedType(), output); err != nil {
		__antithesis_instrumentation__.Notify(599681)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(599682)
	}
	__antithesis_instrumentation__.Notify(599675)
	return output.D, nil
}

type checkConsistencyGenerator struct {
	db       *kv.DB
	from, to roachpb.Key
	mode     roachpb.ChecksumMode

	remainingRows []roachpb.CheckConsistencyResponse_Result
	curRow        roachpb.CheckConsistencyResponse_Result
}

var _ tree.ValueGenerator = &checkConsistencyGenerator{}

func makeCheckConsistencyGenerator(
	ctx *tree.EvalContext, args tree.Datums,
) (tree.ValueGenerator, error) {
	__antithesis_instrumentation__.Notify(599683)
	if !ctx.Codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(599691)
		return nil, errorutil.UnsupportedWithMultiTenancy(
			errorutil.FeatureNotAvailableToNonSystemTenantsIssue)
	} else {
		__antithesis_instrumentation__.Notify(599692)
	}
	__antithesis_instrumentation__.Notify(599684)

	keyFrom := roachpb.Key(*args[1].(*tree.DBytes))
	keyTo := roachpb.Key(*args[2].(*tree.DBytes))

	if len(keyFrom) == 0 {
		__antithesis_instrumentation__.Notify(599693)
		keyFrom = keys.LocalMax
	} else {
		__antithesis_instrumentation__.Notify(599694)
	}
	__antithesis_instrumentation__.Notify(599685)
	if len(keyTo) == 0 {
		__antithesis_instrumentation__.Notify(599695)
		keyTo = roachpb.KeyMax
	} else {
		__antithesis_instrumentation__.Notify(599696)
	}
	__antithesis_instrumentation__.Notify(599686)

	if bytes.Compare(keyFrom, keys.LocalMax) < 0 {
		__antithesis_instrumentation__.Notify(599697)
		return nil, errors.Errorf("start key must be >= %q", []byte(keys.LocalMax))
	} else {
		__antithesis_instrumentation__.Notify(599698)
	}
	__antithesis_instrumentation__.Notify(599687)
	if bytes.Compare(keyTo, roachpb.KeyMax) > 0 {
		__antithesis_instrumentation__.Notify(599699)
		return nil, errors.Errorf("end key must be < %q", []byte(roachpb.KeyMax))
	} else {
		__antithesis_instrumentation__.Notify(599700)
	}
	__antithesis_instrumentation__.Notify(599688)
	if bytes.Compare(keyFrom, keyTo) >= 0 {
		__antithesis_instrumentation__.Notify(599701)
		return nil, errors.New("start key must be less than end key")
	} else {
		__antithesis_instrumentation__.Notify(599702)
	}
	__antithesis_instrumentation__.Notify(599689)

	mode := roachpb.ChecksumMode_CHECK_FULL
	if statsOnly := bool(*args[0].(*tree.DBool)); statsOnly {
		__antithesis_instrumentation__.Notify(599703)
		mode = roachpb.ChecksumMode_CHECK_STATS
	} else {
		__antithesis_instrumentation__.Notify(599704)
	}
	__antithesis_instrumentation__.Notify(599690)

	return &checkConsistencyGenerator{
		db:   ctx.DB,
		from: keyFrom,
		to:   keyTo,
		mode: mode,
	}, nil
}

var checkConsistencyGeneratorType = types.MakeLabeledTuple(
	[]*types.T{types.Int, types.Bytes, types.String, types.String, types.String},
	[]string{"range_id", "start_key", "start_key_pretty", "status", "detail"},
)

func (*checkConsistencyGenerator) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(599705)
	return checkConsistencyGeneratorType
}

func (c *checkConsistencyGenerator) Start(ctx context.Context, _ *kv.Txn) error {
	__antithesis_instrumentation__.Notify(599706)
	var b kv.Batch
	b.AddRawRequest(&roachpb.CheckConsistencyRequest{
		RequestHeader: roachpb.RequestHeader{
			Key:    c.from,
			EndKey: c.to,
		},
		Mode: c.mode,

		WithDiff: c.mode == roachpb.ChecksumMode_CHECK_FULL,
	})

	if err := c.db.Run(ctx, &b); err != nil {
		__antithesis_instrumentation__.Notify(599708)
		return err
	} else {
		__antithesis_instrumentation__.Notify(599709)
	}
	__antithesis_instrumentation__.Notify(599707)
	resp := b.RawResponse().Responses[0].GetInner().(*roachpb.CheckConsistencyResponse)
	c.remainingRows = resp.Result
	return nil
}

func (c *checkConsistencyGenerator) Next(_ context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(599710)
	if len(c.remainingRows) == 0 {
		__antithesis_instrumentation__.Notify(599712)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(599713)
	}
	__antithesis_instrumentation__.Notify(599711)
	c.curRow = c.remainingRows[0]
	c.remainingRows = c.remainingRows[1:]
	return true, nil
}

func (c *checkConsistencyGenerator) Values() (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(599714)
	return tree.Datums{
		tree.NewDInt(tree.DInt(c.curRow.RangeID)),
		tree.NewDBytes(tree.DBytes(c.curRow.StartKey)),
		tree.NewDString(roachpb.Key(c.curRow.StartKey).String()),
		tree.NewDString(c.curRow.Status.String()),
		tree.NewDString(c.curRow.Detail),
	}, nil
}

func (c *checkConsistencyGenerator) Close(_ context.Context) {
	__antithesis_instrumentation__.Notify(599715)
}

const rangeKeyIteratorChunkSize = 256

var rangeKeyIteratorType = types.MakeLabeledTuple(

	[]*types.T{types.String, types.String},
	[]string{"key", "value"},
)

type rangeKeyIterator struct {
	rangeID roachpb.RangeID

	txn *kv.Txn

	kvs []kv.KeyValue

	index int

	buf [2]tree.Datum

	endKey roachpb.RKey
}

var _ tree.ValueGenerator = &rangeKeyIterator{}

func makeRangeKeyIterator(ctx *tree.EvalContext, args tree.Datums) (tree.ValueGenerator, error) {
	__antithesis_instrumentation__.Notify(599716)

	isAdmin, err := ctx.SessionAccessor.HasAdminRole(ctx.Context)
	if err != nil {
		__antithesis_instrumentation__.Notify(599719)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(599720)
	}
	__antithesis_instrumentation__.Notify(599717)
	if !isAdmin {
		__antithesis_instrumentation__.Notify(599721)
		return nil, pgerror.Newf(pgcode.InsufficientPrivilege, "user needs the admin role to view range data")
	} else {
		__antithesis_instrumentation__.Notify(599722)
	}
	__antithesis_instrumentation__.Notify(599718)
	rangeID := roachpb.RangeID(tree.MustBeDInt(args[0]))
	return &rangeKeyIterator{
		rangeID: rangeID,
	}, nil
}

func (rk *rangeKeyIterator) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(599723)
	return rangeKeyIteratorType
}

func (rk *rangeKeyIterator) Start(ctx context.Context, txn *kv.Txn) error {
	__antithesis_instrumentation__.Notify(599724)
	rk.txn = txn

	rangeDesc, err := kvclient.GetRangeWithID(ctx, txn, rk.rangeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(599728)
		return err
	} else {
		__antithesis_instrumentation__.Notify(599729)
	}
	__antithesis_instrumentation__.Notify(599725)
	if rangeDesc == nil {
		__antithesis_instrumentation__.Notify(599730)
		return errors.Newf("range with ID %d not found", rk.rangeID)
	} else {
		__antithesis_instrumentation__.Notify(599731)
	}
	__antithesis_instrumentation__.Notify(599726)

	rk.endKey = rangeDesc.EndKey

	kvs, err := txn.Scan(ctx, rangeDesc.StartKey, rk.endKey, rangeKeyIteratorChunkSize)
	if err != nil {
		__antithesis_instrumentation__.Notify(599732)
		return err
	} else {
		__antithesis_instrumentation__.Notify(599733)
	}
	__antithesis_instrumentation__.Notify(599727)
	rk.kvs = kvs

	rk.index = -1
	return nil
}

func (rk *rangeKeyIterator) Next(ctx context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(599734)
	rk.index++

	if rk.index < len(rk.kvs) {
		__antithesis_instrumentation__.Notify(599738)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(599739)
	}
	__antithesis_instrumentation__.Notify(599735)

	if len(rk.kvs) == 0 {
		__antithesis_instrumentation__.Notify(599740)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(599741)
	}
	__antithesis_instrumentation__.Notify(599736)

	startKey := rk.kvs[len(rk.kvs)-1].Key.Next()
	kvs, err := rk.txn.Scan(ctx, startKey, rk.endKey, rangeKeyIteratorChunkSize)
	if err != nil {
		__antithesis_instrumentation__.Notify(599742)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(599743)
	}
	__antithesis_instrumentation__.Notify(599737)
	rk.kvs = kvs
	rk.index = -1
	return rk.Next(ctx)
}

func (rk *rangeKeyIterator) Values() (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(599744)
	kv := rk.kvs[rk.index]
	rk.buf[0] = tree.NewDString(kv.Key.String())
	rk.buf[1] = tree.NewDString(kv.PrettyValue())
	return rk.buf[:], nil
}

func (rk *rangeKeyIterator) Close(_ context.Context) { __antithesis_instrumentation__.Notify(599745) }

var payloadsForSpanGeneratorLabels = []string{"payload_type", "payload_jsonb"}

var payloadsForSpanGeneratorType = types.MakeLabeledTuple(
	[]*types.T{types.String, types.Jsonb},
	payloadsForSpanGeneratorLabels,
)

type payloadsForSpanGenerator struct {
	span tracing.RegistrySpan

	payloads []json.JSON

	payloadIndex int
}

func makePayloadsForSpanGenerator(
	ctx *tree.EvalContext, args tree.Datums,
) (tree.ValueGenerator, error) {
	__antithesis_instrumentation__.Notify(599746)

	isAdmin, err := ctx.SessionAccessor.HasAdminRole(ctx.Context)
	if err != nil {
		__antithesis_instrumentation__.Notify(599750)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(599751)
	}
	__antithesis_instrumentation__.Notify(599747)
	if !isAdmin {
		__antithesis_instrumentation__.Notify(599752)
		return nil, pgerror.Newf(
			pgcode.InsufficientPrivilege,
			"only users with the admin role are allowed to use crdb_internal.payloads_for_span",
		)
	} else {
		__antithesis_instrumentation__.Notify(599753)
	}
	__antithesis_instrumentation__.Notify(599748)
	spanID := tracingpb.SpanID(*(args[0].(*tree.DInt)))
	span := ctx.Tracer.GetActiveSpanByID(spanID)
	if span == nil {
		__antithesis_instrumentation__.Notify(599754)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(599755)
	}
	__antithesis_instrumentation__.Notify(599749)

	return &payloadsForSpanGenerator{span: span}, nil
}

func (p *payloadsForSpanGenerator) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(599756)
	return payloadsForSpanGeneratorType
}

func (p *payloadsForSpanGenerator) Start(_ context.Context, _ *kv.Txn) error {
	__antithesis_instrumentation__.Notify(599757)

	p.payloadIndex = -1

	rec := p.span.GetFullRecording(tracing.RecordingStructured)
	if rec == nil {
		__antithesis_instrumentation__.Notify(599760)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(599761)
	}
	__antithesis_instrumentation__.Notify(599758)
	p.payloads = make([]json.JSON, len(rec[0].StructuredRecords))
	for i, sr := range rec[0].StructuredRecords {
		__antithesis_instrumentation__.Notify(599762)
		var err error
		p.payloads[i], err = protoreflect.MessageToJSON(sr.Payload, protoreflect.FmtFlags{EmitDefaults: true})
		if err != nil {
			__antithesis_instrumentation__.Notify(599763)
			return err
		} else {
			__antithesis_instrumentation__.Notify(599764)
		}
	}
	__antithesis_instrumentation__.Notify(599759)

	return nil
}

func (p *payloadsForSpanGenerator) Next(_ context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(599765)
	p.payloadIndex++
	return p.payloadIndex < len(p.payloads), nil
}

func (p *payloadsForSpanGenerator) Values() (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(599766)
	payload := p.payloads[p.payloadIndex]
	payloadTypeAsJSON, err := payload.FetchValKey("@type")
	if err != nil {
		__antithesis_instrumentation__.Notify(599768)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(599769)
	}
	__antithesis_instrumentation__.Notify(599767)

	payloadTypeAsString := strings.TrimSuffix(
		strings.TrimPrefix(
			strings.TrimPrefix(
				payloadTypeAsJSON.String(),
				"\"type.googleapis.com/",
			),
			"cockroach."),
		"\"",
	)

	return tree.Datums{
		tree.NewDString(payloadTypeAsString),
		tree.NewDJSON(payload),
	}, nil
}

func (p *payloadsForSpanGenerator) Close(_ context.Context) {
	__antithesis_instrumentation__.Notify(599770)
}

var payloadsForTraceGeneratorLabels = []string{"span_id", "payload_type", "payload_jsonb"}

var payloadsForTraceGeneratorType = types.MakeLabeledTuple(
	[]*types.T{types.Int, types.String, types.Jsonb},
	payloadsForTraceGeneratorLabels,
)

type payloadsForTraceGenerator struct {
	it tree.InternalRows
}

func makePayloadsForTraceGenerator(
	ctx *tree.EvalContext, args tree.Datums,
) (tree.ValueGenerator, error) {
	__antithesis_instrumentation__.Notify(599771)

	isAdmin, err := ctx.SessionAccessor.HasAdminRole(ctx.Context)
	if err != nil {
		__antithesis_instrumentation__.Notify(599775)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(599776)
	}
	__antithesis_instrumentation__.Notify(599772)
	if !isAdmin {
		__antithesis_instrumentation__.Notify(599777)
		return nil, pgerror.Newf(
			pgcode.InsufficientPrivilege,
			"only users with the admin role are allowed to use crdb_internal.payloads_for_trace",
		)
	} else {
		__antithesis_instrumentation__.Notify(599778)
	}
	__antithesis_instrumentation__.Notify(599773)
	traceID := uint64(*(args[0].(*tree.DInt)))

	const query = `WITH spans AS(
									SELECT span_id
  	 							FROM crdb_internal.node_inflight_trace_spans
 		 							WHERE trace_id = $1
									) SELECT * 
										FROM spans, LATERAL crdb_internal.payloads_for_span(spans.span_id)`

	it, err := ctx.Planner.QueryIteratorEx(
		ctx.Ctx(),
		"crdb_internal.payloads_for_trace",
		ctx.Txn,
		sessiondata.NoSessionDataOverride,
		query,
		traceID,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(599779)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(599780)
	}
	__antithesis_instrumentation__.Notify(599774)

	return &payloadsForTraceGenerator{it: it}, nil
}

func (p *payloadsForTraceGenerator) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(599781)
	return payloadsForSpanGeneratorType
}

func (p *payloadsForTraceGenerator) Start(_ context.Context, _ *kv.Txn) error {
	__antithesis_instrumentation__.Notify(599782)
	return nil
}

func (p *payloadsForTraceGenerator) Next(ctx context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(599783)
	return p.it.Next(ctx)
}

func (p *payloadsForTraceGenerator) Values() (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(599784)
	return p.it.Cur(), nil
}

func (p *payloadsForTraceGenerator) Close(_ context.Context) {
	__antithesis_instrumentation__.Notify(599785)
	err := p.it.Close()
	if err != nil {
		__antithesis_instrumentation__.Notify(599786)

		return
	} else {
		__antithesis_instrumentation__.Notify(599787)
	}
}

var showCreateAllSchemasGeneratorType = types.String
var showCreateAllTypesGeneratorType = types.String
var showCreateAllTablesGeneratorType = types.String

type Phase int

const (
	create Phase = iota
	alterAddFks
	alterValidateFks
)

type showCreateAllSchemasGenerator struct {
	evalPlanner tree.EvalPlanner
	txn         *kv.Txn
	ids         []int64
	dbName      string
	acc         mon.BoundAccount

	curr tree.Datum
	idx  int
}

func (s *showCreateAllSchemasGenerator) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(599788)
	return showCreateAllSchemasGeneratorType
}

func (s *showCreateAllSchemasGenerator) Start(ctx context.Context, txn *kv.Txn) error {
	__antithesis_instrumentation__.Notify(599789)
	ids, err := getSchemaIDs(
		ctx, s.evalPlanner, txn, s.dbName, &s.acc,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(599791)
		return err
	} else {
		__antithesis_instrumentation__.Notify(599792)
	}
	__antithesis_instrumentation__.Notify(599790)

	s.ids = ids

	s.txn = txn
	s.idx = -1
	return nil
}

func (s *showCreateAllSchemasGenerator) Next(ctx context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(599793)
	s.idx++
	if s.idx >= len(s.ids) {
		__antithesis_instrumentation__.Notify(599796)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(599797)
	}
	__antithesis_instrumentation__.Notify(599794)

	createStmt, err := getSchemaCreateStatement(
		ctx, s.evalPlanner, s.txn, s.ids[s.idx], s.dbName,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(599798)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(599799)
	}
	__antithesis_instrumentation__.Notify(599795)
	createStmtStr := string(tree.MustBeDString(createStmt))
	s.curr = tree.NewDString(createStmtStr + ";")

	return true, nil
}

func (s *showCreateAllSchemasGenerator) Values() (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(599800)
	return tree.Datums{s.curr}, nil
}

func (s *showCreateAllSchemasGenerator) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(599801)
	s.acc.Close(ctx)
}

func makeShowCreateAllSchemasGenerator(
	ctx *tree.EvalContext, args tree.Datums,
) (tree.ValueGenerator, error) {
	__antithesis_instrumentation__.Notify(599802)
	dbName := string(tree.MustBeDString(args[0]))
	return &showCreateAllSchemasGenerator{
		evalPlanner: ctx.Planner,
		dbName:      dbName,
		acc:         ctx.Mon.MakeBoundAccount(),
	}, nil
}

type showCreateAllTablesGenerator struct {
	evalPlanner tree.EvalPlanner
	txn         *kv.Txn
	ids         []int64
	dbName      string
	acc         mon.BoundAccount
	sessionData *sessiondata.SessionData

	curr           tree.Datum
	idx            int
	shouldValidate bool
	alterArr       tree.Datums
	alterArrIdx    int
	phase          Phase
}

func (s *showCreateAllTablesGenerator) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(599803)
	return showCreateAllTablesGeneratorType
}

func (s *showCreateAllTablesGenerator) Start(ctx context.Context, txn *kv.Txn) error {
	__antithesis_instrumentation__.Notify(599804)

	ids, err := getTopologicallySortedTableIDs(
		ctx, s.evalPlanner, txn, s.dbName, &s.acc,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(599806)
		return err
	} else {
		__antithesis_instrumentation__.Notify(599807)
	}
	__antithesis_instrumentation__.Notify(599805)

	s.ids = ids

	s.txn = txn
	s.idx = -1
	s.phase = create
	return nil
}

func (s *showCreateAllTablesGenerator) Next(ctx context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(599808)
	switch s.phase {
	case create:
		__antithesis_instrumentation__.Notify(599810)
		s.idx++
		if s.idx >= len(s.ids) {
			__antithesis_instrumentation__.Notify(599820)

			s.phase = alterAddFks
			s.idx = -1
			return s.Next(ctx)
		} else {
			__antithesis_instrumentation__.Notify(599821)
		}
		__antithesis_instrumentation__.Notify(599811)

		createStmt, err := getCreateStatement(
			ctx, s.evalPlanner, s.txn, s.ids[s.idx], s.dbName,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(599822)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(599823)
		}
		__antithesis_instrumentation__.Notify(599812)
		createStmtStr := string(tree.MustBeDString(createStmt))
		s.curr = tree.NewDString(createStmtStr + ";")
	case alterAddFks, alterValidateFks:
		__antithesis_instrumentation__.Notify(599813)

		s.alterArrIdx++
		if s.alterArrIdx < len(s.alterArr) {
			__antithesis_instrumentation__.Notify(599824)
			alterStmtStr := string(tree.MustBeDString(s.alterArr[s.alterArrIdx]))
			s.curr = tree.NewDString(alterStmtStr + ";")

			s.shouldValidate = true
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(599825)
		}
		__antithesis_instrumentation__.Notify(599814)

		s.idx++
		if s.idx >= len(s.ids) {
			__antithesis_instrumentation__.Notify(599826)
			if s.phase == alterAddFks {
				__antithesis_instrumentation__.Notify(599828)

				s.phase = alterValidateFks
				s.idx = -1

				if s.shouldValidate {
					__antithesis_instrumentation__.Notify(599830)

					s.curr = tree.NewDString(foreignKeyValidationWarning)
					return true, nil
				} else {
					__antithesis_instrumentation__.Notify(599831)
				}
				__antithesis_instrumentation__.Notify(599829)
				return s.Next(ctx)
			} else {
				__antithesis_instrumentation__.Notify(599832)
			}
			__antithesis_instrumentation__.Notify(599827)

			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(599833)
		}
		__antithesis_instrumentation__.Notify(599815)

		statementReturnType := alterAddFKStatements
		if s.phase == alterValidateFks {
			__antithesis_instrumentation__.Notify(599834)
			statementReturnType = alterValidateFKStatements
		} else {
			__antithesis_instrumentation__.Notify(599835)
		}
		__antithesis_instrumentation__.Notify(599816)
		alterStmt, err := getAlterStatements(
			ctx, s.evalPlanner, s.txn, s.ids[s.idx], s.dbName, statementReturnType,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(599836)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(599837)
		}
		__antithesis_instrumentation__.Notify(599817)
		if alterStmt == nil {
			__antithesis_instrumentation__.Notify(599838)

			return s.Next(ctx)
		} else {
			__antithesis_instrumentation__.Notify(599839)
		}
		__antithesis_instrumentation__.Notify(599818)
		s.alterArr = tree.MustBeDArray(alterStmt).Array
		s.alterArrIdx = -1
		return s.Next(ctx)
	default:
		__antithesis_instrumentation__.Notify(599819)
	}
	__antithesis_instrumentation__.Notify(599809)

	return true, nil
}

func (s *showCreateAllTablesGenerator) Values() (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(599840)
	return tree.Datums{s.curr}, nil
}

func (s *showCreateAllTablesGenerator) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(599841)
	s.acc.Close(ctx)
}

func makeShowCreateAllTablesGenerator(
	ctx *tree.EvalContext, args tree.Datums,
) (tree.ValueGenerator, error) {
	__antithesis_instrumentation__.Notify(599842)
	dbName := string(tree.MustBeDString(args[0]))
	return &showCreateAllTablesGenerator{
		evalPlanner: ctx.Planner,
		dbName:      dbName,
		acc:         ctx.Mon.MakeBoundAccount(),
		sessionData: ctx.SessionData(),
	}, nil
}

type showCreateAllTypesGenerator struct {
	evalPlanner tree.EvalPlanner
	txn         *kv.Txn
	ids         []int64
	dbName      string
	acc         mon.BoundAccount

	curr tree.Datum
	idx  int
}

func (s *showCreateAllTypesGenerator) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(599843)
	return showCreateAllTypesGeneratorType
}

func (s *showCreateAllTypesGenerator) Start(ctx context.Context, txn *kv.Txn) error {
	__antithesis_instrumentation__.Notify(599844)
	ids, err := getTypeIDs(
		ctx, s.evalPlanner, txn, s.dbName, &s.acc,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(599846)
		return err
	} else {
		__antithesis_instrumentation__.Notify(599847)
	}
	__antithesis_instrumentation__.Notify(599845)

	s.ids = ids

	s.txn = txn
	s.idx = -1
	return nil
}

func (s *showCreateAllTypesGenerator) Next(ctx context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(599848)
	s.idx++
	if s.idx >= len(s.ids) {
		__antithesis_instrumentation__.Notify(599851)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(599852)
	}
	__antithesis_instrumentation__.Notify(599849)

	createStmt, err := getTypeCreateStatement(
		ctx, s.evalPlanner, s.txn, s.ids[s.idx], s.dbName,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(599853)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(599854)
	}
	__antithesis_instrumentation__.Notify(599850)
	createStmtStr := string(tree.MustBeDString(createStmt))
	s.curr = tree.NewDString(createStmtStr + ";")

	return true, nil
}

func (s *showCreateAllTypesGenerator) Values() (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(599855)
	return tree.Datums{s.curr}, nil
}

func (s *showCreateAllTypesGenerator) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(599856)
	s.acc.Close(ctx)
}

func makeShowCreateAllTypesGenerator(
	ctx *tree.EvalContext, args tree.Datums,
) (tree.ValueGenerator, error) {
	__antithesis_instrumentation__.Notify(599857)
	dbName := string(tree.MustBeDString(args[0]))
	return &showCreateAllTypesGenerator{
		evalPlanner: ctx.Planner,
		dbName:      dbName,
		acc:         ctx.Mon.MakeBoundAccount(),
	}, nil
}
