package builtins

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc/valueside"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/ipaddr"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"
)

const notUsableInfo = "Not usable; exposed only for compatibility with PostgreSQL."

func makeNotUsableFalseBuiltin() builtinDefinition {
	__antithesis_instrumentation__.Notify(601660)
	return builtinDefinition{
		props: defProps(),
		overloads: []tree.Overload{
			{
				Types:      tree.ArgTypes{},
				ReturnType: tree.FixedReturnType(types.Bool),
				Fn: func(*tree.EvalContext, tree.Datums) (tree.Datum, error) {
					__antithesis_instrumentation__.Notify(601661)
					return tree.DBoolFalse, nil
				},
				Info:       notUsableInfo,
				Volatility: tree.VolatilityVolatile,
			},
		},
	}
}

var typeBuiltinsHaveUnderscore = map[oid.Oid]struct{}{
	types.Any.Oid():         {},
	types.AnyArray.Oid():    {},
	types.Date.Oid():        {},
	types.Time.Oid():        {},
	types.TimeTZ.Oid():      {},
	types.Decimal.Oid():     {},
	types.Interval.Oid():    {},
	types.Jsonb.Oid():       {},
	types.Uuid.Oid():        {},
	types.VarBit.Oid():      {},
	types.Geometry.Oid():    {},
	types.Geography.Oid():   {},
	types.Box2D.Oid():       {},
	oid.T_bit:               {},
	types.Timestamp.Oid():   {},
	types.TimestampTZ.Oid(): {},
	types.AnyTuple.Oid():    {},
}

type UpdatableCommand tree.DInt

const (
	UpdateCommand UpdatableCommand = 2 + iota
	InsertCommand
	DeleteCommand
)

var (
	nonUpdatableEvents = tree.NewDInt(0)
	allUpdatableEvents = tree.NewDInt((1 << UpdateCommand) | (1 << InsertCommand) | (1 << DeleteCommand))
)

func PGIOBuiltinPrefix(typ *types.T) string {
	__antithesis_instrumentation__.Notify(601662)
	builtinPrefix := typ.PGName()
	if _, ok := typeBuiltinsHaveUnderscore[typ.Oid()]; ok {
		__antithesis_instrumentation__.Notify(601664)
		return builtinPrefix + "_"
	} else {
		__antithesis_instrumentation__.Notify(601665)
	}
	__antithesis_instrumentation__.Notify(601663)
	return builtinPrefix
}

func initPGBuiltins() {
	__antithesis_instrumentation__.Notify(601666)
	for k, v := range pgBuiltins {
		__antithesis_instrumentation__.Notify(601672)
		if _, exists := builtins[k]; exists {
			__antithesis_instrumentation__.Notify(601674)
			panic("duplicate builtin: " + k)
		} else {
			__antithesis_instrumentation__.Notify(601675)
		}
		__antithesis_instrumentation__.Notify(601673)
		v.props.Category = categoryCompatibility
		builtins[k] = v
	}
	__antithesis_instrumentation__.Notify(601667)

	for _, typ := range types.OidToType {
		__antithesis_instrumentation__.Notify(601676)

		switch typ.Oid() {
		case oid.T_int2vector, oid.T_oidvector:
			__antithesis_instrumentation__.Notify(601678)
		default:
			__antithesis_instrumentation__.Notify(601679)
			if typ.Family() == types.ArrayFamily {
				__antithesis_instrumentation__.Notify(601680)
				continue
			} else {
				__antithesis_instrumentation__.Notify(601681)
			}
		}
		__antithesis_instrumentation__.Notify(601677)
		builtinPrefix := PGIOBuiltinPrefix(typ)
		for name, builtin := range makeTypeIOBuiltins(builtinPrefix, typ) {
			__antithesis_instrumentation__.Notify(601682)
			builtins[name] = builtin
		}
	}
	__antithesis_instrumentation__.Notify(601668)

	for name, builtin := range makeTypeIOBuiltins("array_", types.AnyArray) {
		__antithesis_instrumentation__.Notify(601683)
		builtins[name] = builtin
	}
	__antithesis_instrumentation__.Notify(601669)
	for name, builtin := range makeTypeIOBuiltins("anyarray_", types.AnyArray) {
		__antithesis_instrumentation__.Notify(601684)
		builtins[name] = builtin
	}
	__antithesis_instrumentation__.Notify(601670)

	for name, builtin := range makeTypeIOBuiltins("enum_", types.AnyEnum) {
		__antithesis_instrumentation__.Notify(601685)
		builtins[name] = builtin
	}
	__antithesis_instrumentation__.Notify(601671)

	for _, typ := range []*types.T{
		types.RegClass,
		types.RegNamespace,
		types.RegProc,
		types.RegProcedure,
		types.RegRole,
		types.RegType,
	} {
		__antithesis_instrumentation__.Notify(601686)
		typName := typ.SQLStandardName()
		builtins["crdb_internal.create_"+typName] = makeCreateRegDef(typ)
	}
}

var errUnimplemented = pgerror.New(pgcode.FeatureNotSupported, "unimplemented")

func makeTypeIOBuiltin(argTypes tree.TypeList, returnType *types.T) builtinDefinition {
	__antithesis_instrumentation__.Notify(601687)
	return builtinDefinition{
		props: tree.FunctionProperties{
			Category: categoryCompatibility,
		},
		overloads: []tree.Overload{
			{
				Types:      argTypes,
				ReturnType: tree.FixedReturnType(returnType),
				Fn: func(_ *tree.EvalContext, _ tree.Datums) (tree.Datum, error) {
					__antithesis_instrumentation__.Notify(601688)
					return nil, errUnimplemented
				},
				Info:       notUsableInfo,
				Volatility: tree.VolatilityVolatile,

				IgnoreVolatilityCheck: true,
			},
		},
	}
}

func makeTypeIOBuiltins(builtinPrefix string, typ *types.T) map[string]builtinDefinition {
	__antithesis_instrumentation__.Notify(601689)
	typname := typ.String()
	return map[string]builtinDefinition{
		builtinPrefix + "send": makeTypeIOBuiltin(tree.ArgTypes{{typname, typ}}, types.Bytes),

		builtinPrefix + "recv": makeTypeIOBuiltin(tree.ArgTypes{{"input", types.Any}}, typ),

		builtinPrefix + "out": makeTypeIOBuiltin(tree.ArgTypes{{typname, typ}}, types.Bytes),

		builtinPrefix + "in": makeTypeIOBuiltin(tree.ArgTypes{{"input", types.Any}}, typ),
	}
}

var (
	DatEncodingUTFId = tree.NewDInt(6)

	DatEncodingEnUTF8        = tree.NewDString("en_US.utf8")
	datEncodingUTF8ShortName = tree.NewDString("UTF8")
)

func makePGGetIndexDef(argTypes tree.ArgTypes) tree.Overload {
	__antithesis_instrumentation__.Notify(601690)
	return tree.Overload{
		Types:      argTypes,
		ReturnType: tree.FixedReturnType(types.String),
		Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601691)
			colNumber := *tree.NewDInt(0)
			if len(args) == 3 {
				__antithesis_instrumentation__.Notify(601699)
				colNumber = *args[1].(*tree.DInt)
			} else {
				__antithesis_instrumentation__.Notify(601700)
			}
			__antithesis_instrumentation__.Notify(601692)
			r, err := ctx.Planner.QueryRowEx(
				ctx.Ctx(), "pg_get_indexdef",
				ctx.Txn,
				sessiondata.NoSessionDataOverride,
				"SELECT indexdef FROM pg_catalog.pg_indexes WHERE crdb_oid = $1", args[0])
			if err != nil {
				__antithesis_instrumentation__.Notify(601701)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(601702)
			}
			__antithesis_instrumentation__.Notify(601693)

			if len(r) == 0 {
				__antithesis_instrumentation__.Notify(601703)
				return tree.DNull, nil
			} else {
				__antithesis_instrumentation__.Notify(601704)
			}
			__antithesis_instrumentation__.Notify(601694)

			if colNumber == 0 {
				__antithesis_instrumentation__.Notify(601705)
				return r[0], nil
			} else {
				__antithesis_instrumentation__.Notify(601706)
			}
			__antithesis_instrumentation__.Notify(601695)

			r, err = ctx.Planner.QueryRowEx(
				ctx.Ctx(), "pg_get_indexdef",
				ctx.Txn,
				sessiondata.NoSessionDataOverride,
				`SELECT ischema.column_name as pg_get_indexdef 
		               FROM information_schema.statistics AS ischema 
											INNER JOIN pg_catalog.pg_indexes AS pgindex 
													ON ischema.table_schema = pgindex.schemaname 
													AND ischema.table_name = pgindex.tablename 
													AND ischema.index_name = pgindex.indexname 
													AND pgindex.crdb_oid = $1 
													AND ischema.seq_in_index = $2`, args[0], args[1])
			if err != nil {
				__antithesis_instrumentation__.Notify(601707)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(601708)
			}
			__antithesis_instrumentation__.Notify(601696)

			if len(r) == 0 {
				__antithesis_instrumentation__.Notify(601709)
				return tree.NewDString(""), nil
			} else {
				__antithesis_instrumentation__.Notify(601710)
			}
			__antithesis_instrumentation__.Notify(601697)
			if len(r) > 1 {
				__antithesis_instrumentation__.Notify(601711)
				return nil, errors.AssertionFailedf("pg_get_indexdef query has more than 1 result row: %+v", r)
			} else {
				__antithesis_instrumentation__.Notify(601712)
			}
			__antithesis_instrumentation__.Notify(601698)
			return r[0], nil
		},
		Info:       notUsableInfo,
		Volatility: tree.VolatilityStable,
	}
}

func makePGGetViewDef(argTypes tree.ArgTypes) tree.Overload {
	__antithesis_instrumentation__.Notify(601713)
	return tree.Overload{
		Types:      argTypes,
		ReturnType: tree.FixedReturnType(types.String),
		Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601714)
			r, err := ctx.Planner.QueryRowEx(
				ctx.Ctx(), "pg_get_viewdef",
				ctx.Txn,
				sessiondata.NoSessionDataOverride,
				"SELECT definition FROM pg_catalog.pg_views v JOIN pg_catalog.pg_class c ON "+
					"c.relname=v.viewname WHERE oid=$1", args[0])
			if err != nil {
				__antithesis_instrumentation__.Notify(601717)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(601718)
			}
			__antithesis_instrumentation__.Notify(601715)
			if len(r) == 0 {
				__antithesis_instrumentation__.Notify(601719)
				return tree.DNull, nil
			} else {
				__antithesis_instrumentation__.Notify(601720)
			}
			__antithesis_instrumentation__.Notify(601716)
			return r[0], nil
		},
		Info:       notUsableInfo,
		Volatility: tree.VolatilityStable,
	}
}

func makePGGetConstraintDef(argTypes tree.ArgTypes) tree.Overload {
	__antithesis_instrumentation__.Notify(601721)
	return tree.Overload{
		Types:      argTypes,
		ReturnType: tree.FixedReturnType(types.String),
		Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601722)
			r, err := ctx.Planner.QueryRowEx(
				ctx.Ctx(), "pg_get_constraintdef",
				ctx.Txn,
				sessiondata.NoSessionDataOverride,
				"SELECT condef FROM pg_catalog.pg_constraint WHERE oid=$1", args[0])
			if err != nil {
				__antithesis_instrumentation__.Notify(601725)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(601726)
			}
			__antithesis_instrumentation__.Notify(601723)
			if len(r) == 0 {
				__antithesis_instrumentation__.Notify(601727)
				return nil, pgerror.Newf(pgcode.InvalidParameterValue, "unknown constraint (OID=%s)", args[0])
			} else {
				__antithesis_instrumentation__.Notify(601728)
			}
			__antithesis_instrumentation__.Notify(601724)
			return r[0], nil
		},
		Info:       notUsableInfo,
		Volatility: tree.VolatilityStable,
	}
}

type argTypeOpts []struct {
	Name string
	Typ  []*types.T
}

var strOrOidTypes = []*types.T{types.String, types.Oid}

func makePGPrivilegeInquiryDef(
	infoDetail string,
	objSpecArgs argTypeOpts,
	fn func(ctx *tree.EvalContext, args tree.Datums, user security.SQLUsername) (tree.HasAnyPrivilegeResult, error),
) builtinDefinition {
	__antithesis_instrumentation__.Notify(601729)

	argTypes := []tree.ArgTypes{
		{},
	}
	for _, typ := range strOrOidTypes {
		__antithesis_instrumentation__.Notify(601734)
		argTypes = append(argTypes, tree.ArgTypes{{"user", typ}})
	}
	__antithesis_instrumentation__.Notify(601730)

	for _, objSpecArg := range objSpecArgs {
		__antithesis_instrumentation__.Notify(601735)
		prevArgTypes := argTypes
		argTypes = make([]tree.ArgTypes, 0, len(argTypes)*len(objSpecArg.Typ))
		for _, argType := range prevArgTypes {
			__antithesis_instrumentation__.Notify(601736)
			for _, typ := range objSpecArg.Typ {
				__antithesis_instrumentation__.Notify(601737)
				argTypeVariant := append(argType, tree.ArgTypes{{objSpecArg.Name, typ}}...)
				argTypes = append(argTypes, argTypeVariant)
			}
		}
	}
	__antithesis_instrumentation__.Notify(601731)

	for i, argType := range argTypes {
		__antithesis_instrumentation__.Notify(601738)
		argTypes[i] = append(argType, tree.ArgTypes{{"privilege", types.String}}...)
	}
	__antithesis_instrumentation__.Notify(601732)

	var variants []tree.Overload
	for _, argType := range argTypes {
		__antithesis_instrumentation__.Notify(601739)
		withUser := argType[0].Name == "user"

		infoFmt := "Returns whether or not the current user has privileges for %s."
		if withUser {
			__antithesis_instrumentation__.Notify(601741)
			infoFmt = "Returns whether or not the user has privileges for %s."
		} else {
			__antithesis_instrumentation__.Notify(601742)
		}
		__antithesis_instrumentation__.Notify(601740)

		variants = append(variants, tree.Overload{
			Types:      argType,
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601743)
				var user security.SQLUsername
				if withUser {
					__antithesis_instrumentation__.Notify(601746)
					arg := tree.UnwrapDatum(ctx, args[0])
					userS, err := getNameForArg(ctx, arg, "pg_roles", "rolname")
					if err != nil {
						__antithesis_instrumentation__.Notify(601749)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(601750)
					}
					__antithesis_instrumentation__.Notify(601747)

					user = security.MakeSQLUsernameFromPreNormalizedString(userS)
					if user.Undefined() {
						__antithesis_instrumentation__.Notify(601751)
						if _, ok := arg.(*tree.DOid); ok {
							__antithesis_instrumentation__.Notify(601753)

							return tree.DBoolFalse, nil
						} else {
							__antithesis_instrumentation__.Notify(601754)
						}
						__antithesis_instrumentation__.Notify(601752)
						return nil, pgerror.Newf(pgcode.UndefinedObject,
							"role %s does not exist", arg)
					} else {
						__antithesis_instrumentation__.Notify(601755)
					}
					__antithesis_instrumentation__.Notify(601748)

					args = args[1:]
				} else {
					__antithesis_instrumentation__.Notify(601756)
					if ctx.SessionData().User().Undefined() {
						__antithesis_instrumentation__.Notify(601758)

						return tree.DNull, nil
					} else {
						__antithesis_instrumentation__.Notify(601759)
					}
					__antithesis_instrumentation__.Notify(601757)
					user = ctx.SessionData().User()
				}
				__antithesis_instrumentation__.Notify(601744)
				ret, err := fn(ctx, args, user)
				if err != nil {
					__antithesis_instrumentation__.Notify(601760)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601761)
				}
				__antithesis_instrumentation__.Notify(601745)
				switch ret {
				case tree.HasPrivilege:
					__antithesis_instrumentation__.Notify(601762)
					return tree.DBoolTrue, nil
				case tree.HasNoPrivilege:
					__antithesis_instrumentation__.Notify(601763)
					return tree.DBoolFalse, nil
				case tree.ObjectNotFound:
					__antithesis_instrumentation__.Notify(601764)
					return tree.DNull, nil
				default:
					__antithesis_instrumentation__.Notify(601765)
					panic(fmt.Sprintf("unrecognized HasAnyPrivilegeResult %d", ret))
				}
			},
			Info:       fmt.Sprintf(infoFmt, infoDetail),
			Volatility: tree.VolatilityStable,
		})
	}
	__antithesis_instrumentation__.Notify(601733)
	return builtinDefinition{
		props: tree.FunctionProperties{
			DistsqlBlocklist: true,
		},
		overloads: variants,
	}
}

func getNameForArg(ctx *tree.EvalContext, arg tree.Datum, pgTable, pgCol string) (string, error) {
	__antithesis_instrumentation__.Notify(601766)
	var query string
	switch t := arg.(type) {
	case *tree.DString:
		__antithesis_instrumentation__.Notify(601769)
		query = fmt.Sprintf("SELECT %s FROM pg_catalog.%s WHERE %s = $1 LIMIT 1", pgCol, pgTable, pgCol)
	case *tree.DOid:
		__antithesis_instrumentation__.Notify(601770)
		query = fmt.Sprintf("SELECT %s FROM pg_catalog.%s WHERE oid = $1 LIMIT 1", pgCol, pgTable)
	default:
		__antithesis_instrumentation__.Notify(601771)
		return "", errors.AssertionFailedf("unexpected arg type %T", t)
	}
	__antithesis_instrumentation__.Notify(601767)
	r, err := ctx.Planner.QueryRowEx(ctx.Ctx(), "get-name-for-arg",
		ctx.Txn, sessiondata.NoSessionDataOverride, query, arg)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(601772)
		return r == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(601773)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(601774)
	}
	__antithesis_instrumentation__.Notify(601768)
	return string(tree.MustBeDString(r[0])), nil
}

type privMap map[string]privilege.Privilege

func normalizePrivilegeStr(arg tree.Datum) []string {
	__antithesis_instrumentation__.Notify(601775)
	argStr := string(tree.MustBeDString(arg))
	privStrs := strings.Split(argStr, ",")
	res := make([]string, len(privStrs))
	for i, privStr := range privStrs {
		__antithesis_instrumentation__.Notify(601777)

		privStr = strings.ToUpper(privStr)

		privStr = strings.TrimSpace(privStr)
		res[i] = privStr
	}
	__antithesis_instrumentation__.Notify(601776)
	return res
}

func parsePrivilegeStr(arg tree.Datum, m privMap) ([]privilege.Privilege, error) {
	__antithesis_instrumentation__.Notify(601778)
	privStrs := normalizePrivilegeStr(arg)
	res := make([]privilege.Privilege, len(privStrs))
	for i, privStr := range privStrs {
		__antithesis_instrumentation__.Notify(601780)

		p, ok := m[privStr]
		if !ok {
			__antithesis_instrumentation__.Notify(601782)
			return nil, pgerror.Newf(pgcode.InvalidParameterValue,
				"unrecognized privilege type: %q", privStr)
		} else {
			__antithesis_instrumentation__.Notify(601783)
		}
		__antithesis_instrumentation__.Notify(601781)
		res[i] = p
	}
	__antithesis_instrumentation__.Notify(601779)
	return res, nil
}

func makeCreateRegDef(typ *types.T) builtinDefinition {
	__antithesis_instrumentation__.Notify(601784)
	return makeBuiltin(defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"oid", types.Int},
				{"name", types.String},
			},
			ReturnType: tree.FixedReturnType(typ),
			Fn: func(_ *tree.EvalContext, d tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601785)
				return tree.NewDOidWithName(tree.MustBeDInt(d[0]), typ, string(tree.MustBeDString(d[1]))), nil
			},
			Info:       notUsableInfo,
			Volatility: tree.VolatilityImmutable,
		},
	)
}

var pgBuiltins = map[string]builtinDefinition{

	"pg_backend_pid": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(_ *tree.EvalContext, _ tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601786)
				return tree.NewDInt(-1), nil
			},
			Info:       notUsableInfo,
			Volatility: tree.VolatilityStable,
		},
	),

	"pg_encoding_to_char": makeBuiltin(defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"encoding_id", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601787)
				if args[0].Compare(ctx, DatEncodingUTFId) == 0 {
					__antithesis_instrumentation__.Notify(601789)
					return datEncodingUTF8ShortName, nil
				} else {
					__antithesis_instrumentation__.Notify(601790)
				}
				__antithesis_instrumentation__.Notify(601788)
				return tree.DNull, nil
			},
			Info:       notUsableInfo,
			Volatility: tree.VolatilityStable,
		},
	),

	"getdatabaseencoding": makeBuiltin(
		tree.FunctionProperties{Category: categorySystemInfo},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601791)

				return datEncodingUTF8ShortName, nil
			},
			Info:       "Returns the current encoding name used by the database.",
			Volatility: tree.VolatilityStable,
		},
	),

	"pg_get_expr": makeBuiltin(defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"pg_node_tree", types.String},
				{"relation_oid", types.Oid},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601792)
				return args[0], nil
			},
			Info:       notUsableInfo,
			Volatility: tree.VolatilityStable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"pg_node_tree", types.String},
				{"relation_oid", types.Oid},
				{"pretty_bool", types.Bool},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601793)
				return args[0], nil
			},
			Info:       notUsableInfo,
			Volatility: tree.VolatilityStable,
		},
	),

	"pg_get_constraintdef": makeBuiltin(tree.FunctionProperties{DistsqlBlocklist: true},
		makePGGetConstraintDef(tree.ArgTypes{
			{"constraint_oid", types.Oid}, {"pretty_bool", types.Bool}}),
		makePGGetConstraintDef(tree.ArgTypes{{"constraint_oid", types.Oid}}),
	),

	"pg_get_partkeydef": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"oid", types.Oid}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601794)
				return tree.DNull, nil
			},
			Info:       notUsableInfo,
			Volatility: tree.VolatilityStable,
		},
	),

	"pg_get_function_result": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"func_oid", types.Oid}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601795)
				funcOid := tree.MustBeDOid(args[0])
				t, err := ctx.Planner.QueryRowEx(
					ctx.Ctx(), "pg_get_function_result",
					ctx.Txn,
					sessiondata.NoSessionDataOverride,
					`SELECT prorettype::REGTYPE::TEXT FROM pg_proc WHERE oid=$1`, int(funcOid.DInt))
				if err != nil {
					__antithesis_instrumentation__.Notify(601798)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601799)
				}
				__antithesis_instrumentation__.Notify(601796)
				if len(t) == 0 {
					__antithesis_instrumentation__.Notify(601800)
					return tree.NewDString(""), nil
				} else {
					__antithesis_instrumentation__.Notify(601801)
				}
				__antithesis_instrumentation__.Notify(601797)
				return t[0], nil
			},
			Info:       notUsableInfo,
			Volatility: tree.VolatilityStable,
		},
	),

	"pg_get_function_identity_arguments": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"func_oid", types.Oid}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601802)
				funcOid := tree.MustBeDOid(args[0])
				t, err := ctx.Planner.QueryRowEx(
					ctx.Ctx(), "pg_get_function_identity_arguments",
					ctx.Txn,
					sessiondata.NoSessionDataOverride,
					`SELECT array_agg(unnest(proargtypes)::REGTYPE::TEXT) FROM pg_proc WHERE oid=$1`, int(funcOid.DInt))
				if err != nil {
					__antithesis_instrumentation__.Notify(601806)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601807)
				}
				__antithesis_instrumentation__.Notify(601803)
				if len(t) == 0 || func() bool {
					__antithesis_instrumentation__.Notify(601808)
					return t[0] == tree.DNull == true
				}() == true {
					__antithesis_instrumentation__.Notify(601809)
					return tree.NewDString(""), nil
				} else {
					__antithesis_instrumentation__.Notify(601810)
				}
				__antithesis_instrumentation__.Notify(601804)
				arr := tree.MustBeDArray(t[0])
				var sb strings.Builder
				for i, elem := range arr.Array {
					__antithesis_instrumentation__.Notify(601811)
					if i > 0 {
						__antithesis_instrumentation__.Notify(601815)
						sb.WriteString(", ")
					} else {
						__antithesis_instrumentation__.Notify(601816)
					}
					__antithesis_instrumentation__.Notify(601812)
					if elem == tree.DNull {
						__antithesis_instrumentation__.Notify(601817)

						sb.WriteString("NULL")
						continue
					} else {
						__antithesis_instrumentation__.Notify(601818)
					}
					__antithesis_instrumentation__.Notify(601813)
					str, ok := tree.AsDString(elem)
					if !ok {
						__antithesis_instrumentation__.Notify(601819)

						sb.WriteString(elem.String())
						continue
					} else {
						__antithesis_instrumentation__.Notify(601820)
					}
					__antithesis_instrumentation__.Notify(601814)
					sb.WriteString(string(str))
				}
				__antithesis_instrumentation__.Notify(601805)
				return tree.NewDString(sb.String()), nil
			},
			Info:       notUsableInfo,
			Volatility: tree.VolatilityStable,
		},
	),

	"pg_get_indexdef": makeBuiltin(tree.FunctionProperties{DistsqlBlocklist: true},
		makePGGetIndexDef(tree.ArgTypes{{"index_oid", types.Oid}}),
		makePGGetIndexDef(tree.ArgTypes{{"index_oid", types.Oid}, {"column_no", types.Int}, {"pretty_bool", types.Bool}}),
	),

	"pg_get_viewdef": makeBuiltin(tree.FunctionProperties{DistsqlBlocklist: true},
		makePGGetViewDef(tree.ArgTypes{{"view_oid", types.Oid}}),
		makePGGetViewDef(tree.ArgTypes{{"view_oid", types.Oid}, {"pretty_bool", types.Bool}}),
	),

	"pg_get_serial_sequence": makeBuiltin(
		tree.FunctionProperties{
			Category: categorySequences,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"table_name", types.String}, {"column_name", types.String}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601821)
				tableName := tree.MustBeDString(args[0])
				columnName := tree.MustBeDString(args[1])
				qualifiedName, err := parser.ParseQualifiedTableName(string(tableName))
				if err != nil {
					__antithesis_instrumentation__.Notify(601825)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601826)
				}
				__antithesis_instrumentation__.Notify(601822)
				res, err := ctx.Sequence.GetSerialSequenceNameFromColumn(ctx.Ctx(), qualifiedName, tree.Name(columnName))
				if err != nil {
					__antithesis_instrumentation__.Notify(601827)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601828)
				}
				__antithesis_instrumentation__.Notify(601823)
				if res == nil {
					__antithesis_instrumentation__.Notify(601829)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(601830)
				}
				__antithesis_instrumentation__.Notify(601824)
				res.ExplicitCatalog = false
				return tree.NewDString(fmt.Sprintf(`%s.%s`, res.Schema(), res.Object())), nil
			},
			Info:       "Returns the name of the sequence used by the given column_name in the table table_name.",
			Volatility: tree.VolatilityStable,
		},
	),

	"pg_my_temp_schema": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.Oid),
			Fn: func(ctx *tree.EvalContext, _ tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601831)
				schema := ctx.SessionData().SearchPath.GetTemporarySchemaName()
				if schema == "" {
					__antithesis_instrumentation__.Notify(601834)

					return tree.NewDOid(0), nil
				} else {
					__antithesis_instrumentation__.Notify(601835)
				}
				__antithesis_instrumentation__.Notify(601832)
				oid, err := ctx.Planner.ResolveOIDFromString(
					ctx.Ctx(), types.RegNamespace, tree.NewDString(schema))
				if err != nil {
					__antithesis_instrumentation__.Notify(601836)

					if pgerror.GetPGCode(err) == pgcode.UndefinedObject {
						__antithesis_instrumentation__.Notify(601838)
						return tree.NewDOid(0), nil
					} else {
						__antithesis_instrumentation__.Notify(601839)
					}
					__antithesis_instrumentation__.Notify(601837)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601840)
				}
				__antithesis_instrumentation__.Notify(601833)
				return oid, nil
			},
			Info: "Returns the OID of the current session's temporary schema, " +
				"or zero if it has none (because it has not created any temporary tables).",
			Volatility: tree.VolatilityStable,
		},
	),

	"pg_is_other_temp_schema": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"oid", types.Oid}},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601841)
				schemaArg := tree.UnwrapDatum(ctx, args[0])
				schema, err := getNameForArg(ctx, schemaArg, "pg_namespace", "nspname")
				if err != nil {
					__antithesis_instrumentation__.Notify(601846)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601847)
				}
				__antithesis_instrumentation__.Notify(601842)
				if schema == "" {
					__antithesis_instrumentation__.Notify(601848)

					return tree.DBoolFalse, nil
				} else {
					__antithesis_instrumentation__.Notify(601849)
				}
				__antithesis_instrumentation__.Notify(601843)
				if !strings.HasPrefix(schema, catconstants.PgTempSchemaName) {
					__antithesis_instrumentation__.Notify(601850)

					return tree.DBoolFalse, nil
				} else {
					__antithesis_instrumentation__.Notify(601851)
				}
				__antithesis_instrumentation__.Notify(601844)
				if schema == ctx.SessionData().SearchPath.GetTemporarySchemaName() {
					__antithesis_instrumentation__.Notify(601852)

					return tree.DBoolFalse, nil
				} else {
					__antithesis_instrumentation__.Notify(601853)
				}
				__antithesis_instrumentation__.Notify(601845)
				return tree.DBoolTrue, nil
			},
			Info:       "Returns true if the given OID is the OID of another session's temporary schema. (This can be useful, for example, to exclude other sessions' temporary tables from a catalog display.)",
			Volatility: tree.VolatilityStable,
		},
	),

	"pg_typeof": makeBuiltin(tree.FunctionProperties{NullableArgs: true},
		tree.Overload{
			Types:      tree.ArgTypes{{"val", types.Any}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601854)
				return tree.NewDString(args[0].ResolvedType().SQLStandardName()), nil
			},
			Info:       notUsableInfo,
			Volatility: tree.VolatilityStable,
		},
	),

	"pg_collation_for": makeBuiltin(
		tree.FunctionProperties{Category: categoryString},
		tree.Overload{
			Types:      tree.ArgTypes{{"str", types.Any}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601855)
				var collation string
				switch t := args[0].(type) {
				case *tree.DString:
					__antithesis_instrumentation__.Notify(601857)
					collation = "default"
				case *tree.DCollatedString:
					__antithesis_instrumentation__.Notify(601858)
					collation = t.Locale
				default:
					__antithesis_instrumentation__.Notify(601859)
					return tree.DNull, pgerror.Newf(pgcode.DatatypeMismatch,
						"collations are not supported by type: %s", t.ResolvedType())
				}
				__antithesis_instrumentation__.Notify(601856)
				return tree.NewDString(fmt.Sprintf(`"%s"`, collation)), nil
			},
			Info:       "Returns the collation of the argument",
			Volatility: tree.VolatilityStable,
		},
	),

	"pg_get_userbyid": makeBuiltin(tree.FunctionProperties{DistsqlBlocklist: true},
		tree.Overload{
			Types: tree.ArgTypes{
				{"role_oid", types.Oid},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601860)
				oid := args[0]
				t, err := ctx.Planner.QueryRowEx(
					ctx.Ctx(), "pg_get_userbyid",
					ctx.Txn,
					sessiondata.NoSessionDataOverride,
					"SELECT rolname FROM pg_catalog.pg_roles WHERE oid=$1", oid)
				if err != nil {
					__antithesis_instrumentation__.Notify(601863)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601864)
				}
				__antithesis_instrumentation__.Notify(601861)
				if len(t) == 0 {
					__antithesis_instrumentation__.Notify(601865)
					return tree.NewDString(fmt.Sprintf("unknown (OID=%s)", args[0])), nil
				} else {
					__antithesis_instrumentation__.Notify(601866)
				}
				__antithesis_instrumentation__.Notify(601862)
				return t[0], nil
			},
			Info:       notUsableInfo,
			Volatility: tree.VolatilityStable,
		},
	),

	"pg_sequence_parameters": makeBuiltin(tree.FunctionProperties{DistsqlBlocklist: true},

		tree.Overload{
			Types:      tree.ArgTypes{{"sequence_oid", types.Oid}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601867)
				r, err := ctx.Planner.QueryRowEx(
					ctx.Ctx(), "pg_sequence_parameters",
					ctx.Txn,
					sessiondata.NoSessionDataOverride,
					`SELECT seqstart, seqmin, seqmax, seqincrement, seqcycle, seqcache, seqtypid `+
						`FROM pg_catalog.pg_sequence WHERE seqrelid=$1`, args[0])
				if err != nil {
					__antithesis_instrumentation__.Notify(601871)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601872)
				}
				__antithesis_instrumentation__.Notify(601868)
				if len(r) == 0 {
					__antithesis_instrumentation__.Notify(601873)
					return nil, pgerror.Newf(pgcode.UndefinedTable, "unknown sequence (OID=%s)", args[0])
				} else {
					__antithesis_instrumentation__.Notify(601874)
				}
				__antithesis_instrumentation__.Notify(601869)
				seqstart, seqmin, seqmax, seqincrement, seqcycle, seqcache, seqtypid := r[0], r[1], r[2], r[3], r[4], r[5], r[6]
				seqcycleStr := "t"
				if seqcycle.(*tree.DBool) == tree.DBoolFalse {
					__antithesis_instrumentation__.Notify(601875)
					seqcycleStr = "f"
				} else {
					__antithesis_instrumentation__.Notify(601876)
				}
				__antithesis_instrumentation__.Notify(601870)
				return tree.NewDString(fmt.Sprintf("(%s,%s,%s,%s,%s,%s,%s)", seqstart, seqmin, seqmax, seqincrement, seqcycleStr, seqcache, seqtypid)), nil
			},
			Info:       notUsableInfo,
			Volatility: tree.VolatilityStable,
		},
	),

	"format_type": makeBuiltin(tree.FunctionProperties{NullableArgs: true, DistsqlBlocklist: true},
		tree.Overload{
			Types:      tree.ArgTypes{{"type_oid", types.Oid}, {"typemod", types.Int}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601877)

				oidArg := args[0]
				if oidArg == tree.DNull {
					__antithesis_instrumentation__.Notify(601881)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(601882)
				}
				__antithesis_instrumentation__.Notify(601878)
				maybeTypmod := args[1]
				oid := oid.Oid(oidArg.(*tree.DOid).DInt)
				typ, ok := types.OidToType[oid]
				if !ok {
					__antithesis_instrumentation__.Notify(601883)

					var err error
					typ, err = ctx.Planner.ResolveTypeByOID(ctx.Context, oid)
					if err != nil {
						__antithesis_instrumentation__.Notify(601884)

						unknown := tree.NewDString(fmt.Sprintf("unknown (OID=%s)", oidArg))
						switch {
						case errors.Is(err, catalog.ErrDescriptorNotFound):
							__antithesis_instrumentation__.Notify(601885)
							return unknown, nil
						case pgerror.GetPGCode(err) == pgcode.UndefinedObject:
							__antithesis_instrumentation__.Notify(601886)
							return unknown, nil
						default:
							__antithesis_instrumentation__.Notify(601887)
							return nil, err
						}
					} else {
						__antithesis_instrumentation__.Notify(601888)
					}
				} else {
					__antithesis_instrumentation__.Notify(601889)
				}
				__antithesis_instrumentation__.Notify(601879)
				var hasTypmod bool
				var typmod int
				if maybeTypmod != tree.DNull {
					__antithesis_instrumentation__.Notify(601890)
					hasTypmod = true
					typmod = int(tree.MustBeDInt(maybeTypmod))
				} else {
					__antithesis_instrumentation__.Notify(601891)
				}
				__antithesis_instrumentation__.Notify(601880)
				return tree.NewDString(typ.SQLStandardNameWithTypmod(hasTypmod, typmod)), nil
			},
			Info: "Returns the SQL name of a data type that is " +
				"identified by its type OID and possibly a type modifier. " +
				"Currently, the type modifier is ignored.",
			Volatility: tree.VolatilityStable,
		},
	),

	"col_description": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"table_oid", types.Oid}, {"column_number", types.Int}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601892)
				if *args[1].(*tree.DInt) == 0 {
					__antithesis_instrumentation__.Notify(601896)

					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(601897)
				}
				__antithesis_instrumentation__.Notify(601893)

				r, err := ctx.Planner.QueryRowEx(
					ctx.Ctx(), "pg_get_coldesc",
					ctx.Txn,
					sessiondata.NoSessionDataOverride,
					`
SELECT COALESCE(c.comment, pc.comment) FROM system.comments c
FULL OUTER JOIN crdb_internal.predefined_comments pc
ON pc.object_id=c.object_id AND pc.sub_id=c.sub_id AND pc.type = c.type
WHERE c.type=$1::int AND c.object_id=$2::int AND c.sub_id=$3::int LIMIT 1
`, keys.ColumnCommentType, args[0], args[1])
				if err != nil {
					__antithesis_instrumentation__.Notify(601898)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601899)
				}
				__antithesis_instrumentation__.Notify(601894)
				if len(r) == 0 {
					__antithesis_instrumentation__.Notify(601900)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(601901)
				}
				__antithesis_instrumentation__.Notify(601895)
				return r[0], nil
			},
			Info:       notUsableInfo,
			Volatility: tree.VolatilityStable,
		},
	),

	"obj_description": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"object_oid", types.Oid}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601902)
				return getPgObjDesc(ctx, "", int(args[0].(*tree.DOid).DInt))
			},
			Info:       notUsableInfo,
			Volatility: tree.VolatilityStable,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"object_oid", types.Oid}, {"catalog_name", types.String}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601903)
				return getPgObjDesc(ctx,
					string(tree.MustBeDString(args[1])),
					int(args[0].(*tree.DOid).DInt),
				)
			},
			Info:       notUsableInfo,
			Volatility: tree.VolatilityStable,
		},
	),

	"oid": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"int", types.Int}},
			ReturnType: tree.FixedReturnType(types.Oid),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601904)
				return tree.NewDOid(*args[0].(*tree.DInt)), nil
			},
			Info:       "Converts an integer to an OID.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"shobj_description": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"object_oid", types.Oid}, {"catalog_name", types.String}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601905)
				catalogName := string(tree.MustBeDString(args[1]))
				objOid := int(args[0].(*tree.DOid).DInt)

				classOid, ok := getCatalogOidForComments(catalogName)
				if !ok {
					__antithesis_instrumentation__.Notify(601909)

					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(601910)
				}
				__antithesis_instrumentation__.Notify(601906)

				r, err := ctx.Planner.QueryRowEx(
					ctx.Ctx(), "pg_get_shobjdesc", ctx.Txn,
					sessiondata.NoSessionDataOverride,
					fmt.Sprintf(`
SELECT description
  FROM pg_catalog.pg_shdescription
 WHERE objoid = %[1]d
   AND classoid = %[2]d
 LIMIT 1`,
						objOid,
						classOid,
					))
				if err != nil {
					__antithesis_instrumentation__.Notify(601911)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601912)
				}
				__antithesis_instrumentation__.Notify(601907)
				if len(r) == 0 {
					__antithesis_instrumentation__.Notify(601913)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(601914)
				}
				__antithesis_instrumentation__.Notify(601908)
				return r[0], nil
			},
			Info:       notUsableInfo,
			Volatility: tree.VolatilityStable,
		},
	),

	"pg_try_advisory_lock": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"int", types.Int}},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(_ *tree.EvalContext, _ tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601915)
				return tree.DBoolTrue, nil
			},
			Info:       notUsableInfo,
			Volatility: tree.VolatilityVolatile,
		},
	),

	"pg_advisory_unlock": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"int", types.Int}},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(_ *tree.EvalContext, _ tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601916)
				return tree.DBoolTrue, nil
			},
			Info:       notUsableInfo,
			Volatility: tree.VolatilityVolatile,
		},
	),

	"pg_client_encoding": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, _ tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601917)
				return tree.NewDString("UTF8"), nil
			},
			Info:       notUsableInfo,
			Volatility: tree.VolatilityStable,
		},
	),

	"pg_function_is_visible": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"oid", types.Oid}},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601918)
				oid := tree.MustBeDOid(args[0])
				t, err := ctx.Planner.QueryRowEx(
					ctx.Ctx(), "pg_function_is_visible",
					ctx.Txn,
					sessiondata.NoSessionDataOverride,
					"SELECT * from pg_proc WHERE oid=$1 LIMIT 1", int(oid.DInt))
				if err != nil {
					__antithesis_instrumentation__.Notify(601921)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601922)
				}
				__antithesis_instrumentation__.Notify(601919)
				if t != nil {
					__antithesis_instrumentation__.Notify(601923)
					return tree.DBoolTrue, nil
				} else {
					__antithesis_instrumentation__.Notify(601924)
				}
				__antithesis_instrumentation__.Notify(601920)
				return tree.DNull, nil
			},
			Info:       notUsableInfo,
			Volatility: tree.VolatilityStable,
		},
	),

	"pg_table_is_visible": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"oid", types.Oid}},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601925)
				oidArg := tree.MustBeDOid(args[0])
				isVisible, exists, err := ctx.Planner.IsTableVisible(
					ctx.Context, ctx.SessionData().Database, ctx.SessionData().SearchPath, oid.Oid(oidArg.DInt),
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(601928)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601929)
				}
				__antithesis_instrumentation__.Notify(601926)
				if !exists {
					__antithesis_instrumentation__.Notify(601930)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(601931)
				}
				__antithesis_instrumentation__.Notify(601927)
				return tree.MakeDBool(tree.DBool(isVisible)), nil
			},
			Info:       "Returns whether the table with the given OID belongs to one of the schemas on the search path.",
			Volatility: tree.VolatilityStable,
		},
	),

	"pg_type_is_visible": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"oid", types.Oid}},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601932)
				oidArg := tree.MustBeDOid(args[0])
				isVisible, exists, err := ctx.Planner.IsTypeVisible(
					ctx.Context, ctx.SessionData().Database, ctx.SessionData().SearchPath, oid.Oid(oidArg.DInt),
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(601935)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601936)
				}
				__antithesis_instrumentation__.Notify(601933)
				if !exists {
					__antithesis_instrumentation__.Notify(601937)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(601938)
				}
				__antithesis_instrumentation__.Notify(601934)
				return tree.MakeDBool(tree.DBool(isVisible)), nil
			},
			Info:       "Returns whether the type with the given OID belongs to one of the schemas on the search path.",
			Volatility: tree.VolatilityStable,
		},
	),

	"pg_relation_is_updatable": makeBuiltin(
		defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"reloid", types.Oid}, {"include_triggers", types.Bool}},
			ReturnType: tree.FixedReturnType(types.Int4),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601939)
				oidArg := tree.MustBeDOid(args[0])
				oid := int(oidArg.DInt)
				tableDescI, err := ctx.Planner.GetImmutableTableInterfaceByID(ctx.Ctx(), oid)
				if err != nil {
					__antithesis_instrumentation__.Notify(601942)

					if sqlerrors.IsUndefinedRelationError(err) {
						__antithesis_instrumentation__.Notify(601944)
						return nonUpdatableEvents, nil
					} else {
						__antithesis_instrumentation__.Notify(601945)
					}
					__antithesis_instrumentation__.Notify(601943)
					return nonUpdatableEvents, err
				} else {
					__antithesis_instrumentation__.Notify(601946)
				}
				__antithesis_instrumentation__.Notify(601940)
				tableDesc := tableDescI.(catalog.TableDescriptor)
				if !tableDesc.IsTable() || func() bool {
					__antithesis_instrumentation__.Notify(601947)
					return tableDesc.IsVirtualTable() == true
				}() == true {
					__antithesis_instrumentation__.Notify(601948)
					return nonUpdatableEvents, nil
				} else {
					__antithesis_instrumentation__.Notify(601949)
				}
				__antithesis_instrumentation__.Notify(601941)

				return allUpdatableEvents, nil
			},
			Info:       `Returns the update events the relation supports.`,
			Volatility: tree.VolatilityStable,
		},
	),

	"pg_column_is_updatable": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"reloid", types.Oid},
				{"attnum", types.Int2},
				{"include_triggers", types.Bool},
			},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601950)
				oidArg := tree.MustBeDOid(args[0])
				attNumArg := tree.MustBeDInt(args[1])
				oid := int(oidArg.DInt)
				attNum := uint32(attNumArg)
				if attNumArg < 0 {
					__antithesis_instrumentation__.Notify(601955)

					return tree.DBoolFalse, nil
				} else {
					__antithesis_instrumentation__.Notify(601956)
				}
				__antithesis_instrumentation__.Notify(601951)
				tableDescI, err := ctx.Planner.GetImmutableTableInterfaceByID(ctx.Ctx(), oid)
				if err != nil {
					__antithesis_instrumentation__.Notify(601957)
					if sqlerrors.IsUndefinedRelationError(err) {
						__antithesis_instrumentation__.Notify(601959)

						return tree.DBoolFalse, nil
					} else {
						__antithesis_instrumentation__.Notify(601960)
					}
					__antithesis_instrumentation__.Notify(601958)
					return tree.DBoolFalse, err
				} else {
					__antithesis_instrumentation__.Notify(601961)
				}
				__antithesis_instrumentation__.Notify(601952)
				tableDesc := tableDescI.(catalog.TableDescriptor)
				if !tableDesc.IsTable() || func() bool {
					__antithesis_instrumentation__.Notify(601962)
					return tableDesc.IsVirtualTable() == true
				}() == true {
					__antithesis_instrumentation__.Notify(601963)
					return tree.DBoolFalse, nil
				} else {
					__antithesis_instrumentation__.Notify(601964)
				}
				__antithesis_instrumentation__.Notify(601953)

				column, err := tableDesc.FindColumnWithID(descpb.ColumnID(attNum))
				if err != nil {
					__antithesis_instrumentation__.Notify(601965)
					if sqlerrors.IsUndefinedColumnError(err) {
						__antithesis_instrumentation__.Notify(601967)

						return tree.DBoolTrue, nil
					} else {
						__antithesis_instrumentation__.Notify(601968)
					}
					__antithesis_instrumentation__.Notify(601966)
					return tree.DBoolFalse, err
				} else {
					__antithesis_instrumentation__.Notify(601969)
				}
				__antithesis_instrumentation__.Notify(601954)

				return tree.MakeDBool(tree.DBool(!column.IsComputed())), nil
			},
			Info:       `Returns whether the given column can be updated.`,
			Volatility: tree.VolatilityStable,
		},
	),

	"pg_sleep": makeBuiltin(
		tree.FunctionProperties{},
		tree.Overload{
			Types:      tree.ArgTypes{{"seconds", types.Float}},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601970)
				durationNanos := int64(float64(*args[0].(*tree.DFloat)) * float64(1000000000))
				dur := time.Duration(durationNanos)
				select {
				case <-ctx.Ctx().Done():
					__antithesis_instrumentation__.Notify(601971)
					return nil, ctx.Ctx().Err()
				case <-time.After(dur):
					__antithesis_instrumentation__.Notify(601972)
					return tree.DBoolTrue, nil
				}
			},
			Info: "pg_sleep makes the current session's process sleep until " +
				"seconds seconds have elapsed. seconds is a value of type " +
				"double precision, so fractional-second delays can be specified.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"pg_is_in_recovery": makeNotUsableFalseBuiltin(),

	"pg_is_xlog_replay_paused": makeNotUsableFalseBuiltin(),

	"has_any_column_privilege": makePGPrivilegeInquiryDef(
		"any column of table",
		argTypeOpts{{"table", strOrOidTypes}},
		func(ctx *tree.EvalContext, args tree.Datums, user security.SQLUsername) (tree.HasAnyPrivilegeResult, error) {
			__antithesis_instrumentation__.Notify(601973)
			tableArg := tree.UnwrapDatum(ctx, args[0])
			specifier, err := tableHasPrivilegeSpecifier(tableArg, false)
			if err != nil {
				__antithesis_instrumentation__.Notify(601976)
				return tree.HasNoPrivilege, err
			} else {
				__antithesis_instrumentation__.Notify(601977)
			}
			__antithesis_instrumentation__.Notify(601974)

			privs, err := parsePrivilegeStr(args[1], privMap{
				"SELECT":                       {Kind: privilege.SELECT},
				"SELECT WITH GRANT OPTION":     {Kind: privilege.SELECT, GrantOption: true},
				"INSERT":                       {Kind: privilege.INSERT},
				"INSERT WITH GRANT OPTION":     {Kind: privilege.INSERT, GrantOption: true},
				"UPDATE":                       {Kind: privilege.UPDATE},
				"UPDATE WITH GRANT OPTION":     {Kind: privilege.UPDATE, GrantOption: true},
				"REFERENCES":                   {Kind: privilege.SELECT},
				"REFERENCES WITH GRANT OPTION": {Kind: privilege.SELECT, GrantOption: true},
			})
			if err != nil {
				__antithesis_instrumentation__.Notify(601978)
				return tree.HasNoPrivilege, err
			} else {
				__antithesis_instrumentation__.Notify(601979)
			}
			__antithesis_instrumentation__.Notify(601975)
			return ctx.Planner.HasAnyPrivilege(ctx.Context, specifier, user, privs)
		},
	),

	"has_column_privilege": makePGPrivilegeInquiryDef(
		"column",
		argTypeOpts{{"table", strOrOidTypes}, {"column", []*types.T{types.String, types.Int}}},
		func(ctx *tree.EvalContext, args tree.Datums, user security.SQLUsername) (tree.HasAnyPrivilegeResult, error) {
			__antithesis_instrumentation__.Notify(601980)
			tableArg := tree.UnwrapDatum(ctx, args[0])
			colArg := tree.UnwrapDatum(ctx, args[1])
			specifier, err := columnHasPrivilegeSpecifier(tableArg, colArg)
			if err != nil {
				__antithesis_instrumentation__.Notify(601983)
				return tree.HasNoPrivilege, err
			} else {
				__antithesis_instrumentation__.Notify(601984)
			}
			__antithesis_instrumentation__.Notify(601981)

			privs, err := parsePrivilegeStr(args[2], privMap{
				"SELECT":                       {Kind: privilege.SELECT},
				"SELECT WITH GRANT OPTION":     {Kind: privilege.SELECT, GrantOption: true},
				"INSERT":                       {Kind: privilege.INSERT},
				"INSERT WITH GRANT OPTION":     {Kind: privilege.INSERT, GrantOption: true},
				"UPDATE":                       {Kind: privilege.UPDATE},
				"UPDATE WITH GRANT OPTION":     {Kind: privilege.UPDATE, GrantOption: true},
				"REFERENCES":                   {Kind: privilege.SELECT},
				"REFERENCES WITH GRANT OPTION": {Kind: privilege.SELECT, GrantOption: true},
			})
			if err != nil {
				__antithesis_instrumentation__.Notify(601985)
				return tree.HasNoPrivilege, err
			} else {
				__antithesis_instrumentation__.Notify(601986)
			}
			__antithesis_instrumentation__.Notify(601982)
			return ctx.Planner.HasAnyPrivilege(ctx.Context, specifier, user, privs)
		},
	),

	"has_database_privilege": makePGPrivilegeInquiryDef(
		"database",
		argTypeOpts{{"database", strOrOidTypes}},
		func(ctx *tree.EvalContext, args tree.Datums, user security.SQLUsername) (tree.HasAnyPrivilegeResult, error) {
			__antithesis_instrumentation__.Notify(601987)

			databaseArg := tree.UnwrapDatum(ctx, args[0])
			specifier, err := databaseHasPrivilegeSpecifier(databaseArg)
			if err != nil {
				__antithesis_instrumentation__.Notify(601990)
				return tree.HasNoPrivilege, err
			} else {
				__antithesis_instrumentation__.Notify(601991)
			}
			__antithesis_instrumentation__.Notify(601988)

			privs, err := parsePrivilegeStr(args[1], privMap{
				"CREATE":                      {Kind: privilege.CREATE},
				"CREATE WITH GRANT OPTION":    {Kind: privilege.CREATE, GrantOption: true},
				"CONNECT":                     {Kind: privilege.CONNECT},
				"CONNECT WITH GRANT OPTION":   {Kind: privilege.CONNECT, GrantOption: true},
				"TEMPORARY":                   {Kind: privilege.CREATE},
				"TEMPORARY WITH GRANT OPTION": {Kind: privilege.CREATE, GrantOption: true},
				"TEMP":                        {Kind: privilege.CREATE},
				"TEMP WITH GRANT OPTION":      {Kind: privilege.CREATE, GrantOption: true},
			})
			if err != nil {
				__antithesis_instrumentation__.Notify(601992)
				return tree.HasNoPrivilege, err
			} else {
				__antithesis_instrumentation__.Notify(601993)
			}
			__antithesis_instrumentation__.Notify(601989)

			return ctx.Planner.HasAnyPrivilege(ctx.Context, specifier, user, privs)
		},
	),

	"has_foreign_data_wrapper_privilege": makePGPrivilegeInquiryDef(
		"foreign-data wrapper",
		argTypeOpts{{"fdw", strOrOidTypes}},
		func(ctx *tree.EvalContext, args tree.Datums, user security.SQLUsername) (tree.HasAnyPrivilegeResult, error) {
			__antithesis_instrumentation__.Notify(601994)
			fdwArg := tree.UnwrapDatum(ctx, args[0])
			fdw, err := getNameForArg(ctx, fdwArg, "pg_foreign_data_wrapper", "fdwname")
			if err != nil {
				__antithesis_instrumentation__.Notify(601999)
				return tree.HasNoPrivilege, err
			} else {
				__antithesis_instrumentation__.Notify(602000)
			}
			__antithesis_instrumentation__.Notify(601995)
			retNull := false
			if fdw == "" {
				__antithesis_instrumentation__.Notify(602001)
				switch fdwArg.(type) {
				case *tree.DString:
					__antithesis_instrumentation__.Notify(602002)
					return tree.HasNoPrivilege, pgerror.Newf(pgcode.UndefinedObject,
						"foreign-data wrapper %s does not exist", fdwArg)
				case *tree.DOid:
					__antithesis_instrumentation__.Notify(602003)

					retNull = true
				}
			} else {
				__antithesis_instrumentation__.Notify(602004)
			}
			__antithesis_instrumentation__.Notify(601996)

			privs, err := parsePrivilegeStr(args[1], privMap{
				"USAGE":                   {Kind: privilege.USAGE},
				"USAGE WITH GRANT OPTION": {Kind: privilege.USAGE, GrantOption: true},
			})
			if err != nil {
				__antithesis_instrumentation__.Notify(602005)
				return tree.HasNoPrivilege, err
			} else {
				__antithesis_instrumentation__.Notify(602006)
			}
			__antithesis_instrumentation__.Notify(601997)
			if retNull {
				__antithesis_instrumentation__.Notify(602007)
				return tree.ObjectNotFound, nil
			} else {
				__antithesis_instrumentation__.Notify(602008)
			}
			__antithesis_instrumentation__.Notify(601998)

			_ = privs
			return tree.HasPrivilege, nil
		},
	),

	"has_function_privilege": makePGPrivilegeInquiryDef(
		"function",
		argTypeOpts{{"function", strOrOidTypes}},
		func(ctx *tree.EvalContext, args tree.Datums, user security.SQLUsername) (tree.HasAnyPrivilegeResult, error) {
			__antithesis_instrumentation__.Notify(602009)
			oidArg := tree.UnwrapDatum(ctx, args[0])

			var oid tree.Datum
			switch t := oidArg.(type) {
			case *tree.DString:
				__antithesis_instrumentation__.Notify(602015)
				var err error
				oid, err = tree.ParseDOid(ctx, string(*t), types.RegProcedure)
				if err != nil {
					__antithesis_instrumentation__.Notify(602017)
					return tree.HasNoPrivilege, err
				} else {
					__antithesis_instrumentation__.Notify(602018)
				}
			case *tree.DOid:
				__antithesis_instrumentation__.Notify(602016)
				oid = t
			}
			__antithesis_instrumentation__.Notify(602010)

			fn, err := getNameForArg(ctx, oid, "pg_proc", "proname")
			if err != nil {
				__antithesis_instrumentation__.Notify(602019)
				return tree.HasNoPrivilege, err
			} else {
				__antithesis_instrumentation__.Notify(602020)
			}
			__antithesis_instrumentation__.Notify(602011)
			retNull := false
			if fn == "" {
				__antithesis_instrumentation__.Notify(602021)

				retNull = true
			} else {
				__antithesis_instrumentation__.Notify(602022)
			}
			__antithesis_instrumentation__.Notify(602012)

			privs, err := parsePrivilegeStr(args[1], privMap{

				"EXECUTE":                   {Kind: privilege.USAGE},
				"EXECUTE WITH GRANT OPTION": {Kind: privilege.USAGE, GrantOption: true},
			})
			if err != nil {
				__antithesis_instrumentation__.Notify(602023)
				return tree.HasNoPrivilege, err
			} else {
				__antithesis_instrumentation__.Notify(602024)
			}
			__antithesis_instrumentation__.Notify(602013)
			if retNull {
				__antithesis_instrumentation__.Notify(602025)
				return tree.ObjectNotFound, nil
			} else {
				__antithesis_instrumentation__.Notify(602026)
			}
			__antithesis_instrumentation__.Notify(602014)

			_ = privs
			return tree.HasPrivilege, nil
		},
	),

	"has_language_privilege": makePGPrivilegeInquiryDef(
		"language",
		argTypeOpts{{"language", strOrOidTypes}},
		func(ctx *tree.EvalContext, args tree.Datums, user security.SQLUsername) (tree.HasAnyPrivilegeResult, error) {
			__antithesis_instrumentation__.Notify(602027)
			langArg := tree.UnwrapDatum(ctx, args[0])
			lang, err := getNameForArg(ctx, langArg, "pg_language", "lanname")
			if err != nil {
				__antithesis_instrumentation__.Notify(602032)
				return tree.HasNoPrivilege, err
			} else {
				__antithesis_instrumentation__.Notify(602033)
			}
			__antithesis_instrumentation__.Notify(602028)
			retNull := false
			if lang == "" {
				__antithesis_instrumentation__.Notify(602034)
				switch langArg.(type) {
				case *tree.DString:
					__antithesis_instrumentation__.Notify(602035)
					return tree.HasNoPrivilege, pgerror.Newf(pgcode.UndefinedObject,
						"language %s does not exist", langArg)
				case *tree.DOid:
					__antithesis_instrumentation__.Notify(602036)

					retNull = true
				}
			} else {
				__antithesis_instrumentation__.Notify(602037)
			}
			__antithesis_instrumentation__.Notify(602029)

			privs, err := parsePrivilegeStr(args[1], privMap{
				"USAGE":                   {Kind: privilege.USAGE},
				"USAGE WITH GRANT OPTION": {Kind: privilege.USAGE, GrantOption: true},
			})
			if err != nil {
				__antithesis_instrumentation__.Notify(602038)
				return tree.HasNoPrivilege, err
			} else {
				__antithesis_instrumentation__.Notify(602039)
			}
			__antithesis_instrumentation__.Notify(602030)
			if retNull {
				__antithesis_instrumentation__.Notify(602040)
				return tree.ObjectNotFound, nil
			} else {
				__antithesis_instrumentation__.Notify(602041)
			}
			__antithesis_instrumentation__.Notify(602031)

			_ = privs
			return tree.HasPrivilege, nil
		},
	),

	"has_schema_privilege": makePGPrivilegeInquiryDef(
		"schema",
		argTypeOpts{{"schema", strOrOidTypes}},
		func(ctx *tree.EvalContext, args tree.Datums, user security.SQLUsername) (tree.HasAnyPrivilegeResult, error) {
			__antithesis_instrumentation__.Notify(602042)
			schemaArg := tree.UnwrapDatum(ctx, args[0])
			databaseName := ctx.SessionData().Database
			specifier, err := schemaHasPrivilegeSpecifier(ctx, schemaArg, databaseName)
			if err != nil {
				__antithesis_instrumentation__.Notify(602046)
				return tree.HasNoPrivilege, err
			} else {
				__antithesis_instrumentation__.Notify(602047)
			}
			__antithesis_instrumentation__.Notify(602043)

			privs, err := parsePrivilegeStr(args[1], privMap{
				"CREATE":                   {Kind: privilege.CREATE},
				"CREATE WITH GRANT OPTION": {Kind: privilege.CREATE, GrantOption: true},
				"USAGE":                    {Kind: privilege.USAGE},
				"USAGE WITH GRANT OPTION":  {Kind: privilege.USAGE, GrantOption: true},
			})
			if err != nil {
				__antithesis_instrumentation__.Notify(602048)
				return tree.HasNoPrivilege, err
			} else {
				__antithesis_instrumentation__.Notify(602049)
			}
			__antithesis_instrumentation__.Notify(602044)
			if len(databaseName) == 0 {
				__antithesis_instrumentation__.Notify(602050)

				return tree.ObjectNotFound, nil
			} else {
				__antithesis_instrumentation__.Notify(602051)
			}
			__antithesis_instrumentation__.Notify(602045)

			return ctx.Planner.HasAnyPrivilege(ctx.Context, specifier, user, privs)
		},
	),

	"has_sequence_privilege": makePGPrivilegeInquiryDef(
		"sequence",
		argTypeOpts{{"sequence", strOrOidTypes}},
		func(ctx *tree.EvalContext, args tree.Datums, user security.SQLUsername) (tree.HasAnyPrivilegeResult, error) {
			__antithesis_instrumentation__.Notify(602052)
			seqArg := tree.UnwrapDatum(ctx, args[0])
			specifier, err := tableHasPrivilegeSpecifier(seqArg, true)
			if err != nil {
				__antithesis_instrumentation__.Notify(602055)
				return tree.HasNoPrivilege, err
			} else {
				__antithesis_instrumentation__.Notify(602056)
			}
			__antithesis_instrumentation__.Notify(602053)
			privs, err := parsePrivilegeStr(args[1], privMap{

				"USAGE":                    {Kind: privilege.SELECT},
				"USAGE WITH GRANT OPTION":  {Kind: privilege.SELECT, GrantOption: true},
				"SELECT":                   {Kind: privilege.SELECT},
				"SELECT WITH GRANT OPTION": {Kind: privilege.SELECT, GrantOption: true},
				"UPDATE":                   {Kind: privilege.UPDATE},
				"UPDATE WITH GRANT OPTION": {Kind: privilege.UPDATE, GrantOption: true},
			})
			if err != nil {
				__antithesis_instrumentation__.Notify(602057)
				return tree.HasPrivilege, err
			} else {
				__antithesis_instrumentation__.Notify(602058)
			}
			__antithesis_instrumentation__.Notify(602054)
			return ctx.Planner.HasAnyPrivilege(ctx.Context, specifier, user, privs)
		},
	),

	"has_server_privilege": makePGPrivilegeInquiryDef(
		"foreign server",
		argTypeOpts{{"server", strOrOidTypes}},
		func(ctx *tree.EvalContext, args tree.Datums, user security.SQLUsername) (tree.HasAnyPrivilegeResult, error) {
			__antithesis_instrumentation__.Notify(602059)
			serverArg := tree.UnwrapDatum(ctx, args[0])
			server, err := getNameForArg(ctx, serverArg, "pg_foreign_server", "srvname")
			if err != nil {
				__antithesis_instrumentation__.Notify(602064)
				return tree.HasNoPrivilege, err
			} else {
				__antithesis_instrumentation__.Notify(602065)
			}
			__antithesis_instrumentation__.Notify(602060)
			retNull := false
			if server == "" {
				__antithesis_instrumentation__.Notify(602066)
				switch serverArg.(type) {
				case *tree.DString:
					__antithesis_instrumentation__.Notify(602067)
					return tree.HasNoPrivilege, pgerror.Newf(pgcode.UndefinedObject,
						"server %s does not exist", serverArg)
				case *tree.DOid:
					__antithesis_instrumentation__.Notify(602068)

					retNull = true
				}
			} else {
				__antithesis_instrumentation__.Notify(602069)
			}
			__antithesis_instrumentation__.Notify(602061)

			privs, err := parsePrivilegeStr(args[1], privMap{
				"USAGE":                   {Kind: privilege.USAGE},
				"USAGE WITH GRANT OPTION": {Kind: privilege.USAGE, GrantOption: true},
			})
			if err != nil {
				__antithesis_instrumentation__.Notify(602070)
				return tree.HasNoPrivilege, err
			} else {
				__antithesis_instrumentation__.Notify(602071)
			}
			__antithesis_instrumentation__.Notify(602062)
			if retNull {
				__antithesis_instrumentation__.Notify(602072)
				return tree.ObjectNotFound, nil
			} else {
				__antithesis_instrumentation__.Notify(602073)
			}
			__antithesis_instrumentation__.Notify(602063)

			_ = privs
			return tree.HasPrivilege, nil
		},
	),

	"has_table_privilege": makePGPrivilegeInquiryDef(
		"table",
		argTypeOpts{{"table", strOrOidTypes}},
		func(ctx *tree.EvalContext, args tree.Datums, user security.SQLUsername) (tree.HasAnyPrivilegeResult, error) {
			__antithesis_instrumentation__.Notify(602074)
			tableArg := tree.UnwrapDatum(ctx, args[0])
			specifier, err := tableHasPrivilegeSpecifier(tableArg, false)
			if err != nil {
				__antithesis_instrumentation__.Notify(602077)
				return tree.HasNoPrivilege, err
			} else {
				__antithesis_instrumentation__.Notify(602078)
			}
			__antithesis_instrumentation__.Notify(602075)

			privs, err := parsePrivilegeStr(args[1], privMap{
				"SELECT":                       {Kind: privilege.SELECT},
				"SELECT WITH GRANT OPTION":     {Kind: privilege.SELECT, GrantOption: true},
				"INSERT":                       {Kind: privilege.INSERT},
				"INSERT WITH GRANT OPTION":     {Kind: privilege.INSERT, GrantOption: true},
				"UPDATE":                       {Kind: privilege.UPDATE},
				"UPDATE WITH GRANT OPTION":     {Kind: privilege.UPDATE, GrantOption: true},
				"DELETE":                       {Kind: privilege.DELETE},
				"DELETE WITH GRANT OPTION":     {Kind: privilege.DELETE, GrantOption: true},
				"TRUNCATE":                     {Kind: privilege.DELETE},
				"TRUNCATE WITH GRANT OPTION":   {Kind: privilege.DELETE, GrantOption: true},
				"REFERENCES":                   {Kind: privilege.SELECT},
				"REFERENCES WITH GRANT OPTION": {Kind: privilege.SELECT, GrantOption: true},
				"TRIGGER":                      {Kind: privilege.CREATE},
				"TRIGGER WITH GRANT OPTION":    {Kind: privilege.CREATE, GrantOption: true},
				"RULE":                         {Kind: privilege.RULE},
				"RULE WITH GRANT OPTION":       {Kind: privilege.RULE, GrantOption: true},
			})
			if err != nil {
				__antithesis_instrumentation__.Notify(602079)
				return tree.HasNoPrivilege, err
			} else {
				__antithesis_instrumentation__.Notify(602080)
			}
			__antithesis_instrumentation__.Notify(602076)
			return ctx.Planner.HasAnyPrivilege(ctx.Context, specifier, user, privs)
		},
	),

	"has_tablespace_privilege": makePGPrivilegeInquiryDef(
		"tablespace",
		argTypeOpts{{"tablespace", strOrOidTypes}},
		func(ctx *tree.EvalContext, args tree.Datums, user security.SQLUsername) (tree.HasAnyPrivilegeResult, error) {
			__antithesis_instrumentation__.Notify(602081)
			tablespaceArg := tree.UnwrapDatum(ctx, args[0])
			tablespace, err := getNameForArg(ctx, tablespaceArg, "pg_tablespace", "spcname")
			if err != nil {
				__antithesis_instrumentation__.Notify(602086)
				return tree.HasNoPrivilege, err
			} else {
				__antithesis_instrumentation__.Notify(602087)
			}
			__antithesis_instrumentation__.Notify(602082)
			retNull := false
			if tablespace == "" {
				__antithesis_instrumentation__.Notify(602088)
				switch tablespaceArg.(type) {
				case *tree.DString:
					__antithesis_instrumentation__.Notify(602089)
					return tree.HasNoPrivilege, pgerror.Newf(pgcode.UndefinedObject,
						"tablespace %s does not exist", tablespaceArg)
				case *tree.DOid:
					__antithesis_instrumentation__.Notify(602090)

					retNull = true
				}
			} else {
				__antithesis_instrumentation__.Notify(602091)
			}
			__antithesis_instrumentation__.Notify(602083)

			privs, err := parsePrivilegeStr(args[1], privMap{
				"CREATE":                   {Kind: privilege.CREATE},
				"CREATE WITH GRANT OPTION": {Kind: privilege.CREATE, GrantOption: true},
			})
			if err != nil {
				__antithesis_instrumentation__.Notify(602092)
				return tree.HasNoPrivilege, err
			} else {
				__antithesis_instrumentation__.Notify(602093)
			}
			__antithesis_instrumentation__.Notify(602084)
			if retNull {
				__antithesis_instrumentation__.Notify(602094)
				return tree.ObjectNotFound, nil
			} else {
				__antithesis_instrumentation__.Notify(602095)
			}
			__antithesis_instrumentation__.Notify(602085)

			_ = privs
			return tree.HasPrivilege, nil
		},
	),

	"has_type_privilege": makePGPrivilegeInquiryDef(
		"type",
		argTypeOpts{{"type", strOrOidTypes}},
		func(ctx *tree.EvalContext, args tree.Datums, user security.SQLUsername) (tree.HasAnyPrivilegeResult, error) {
			__antithesis_instrumentation__.Notify(602096)
			oidArg := tree.UnwrapDatum(ctx, args[0])

			var oid tree.Datum
			switch t := oidArg.(type) {
			case *tree.DString:
				__antithesis_instrumentation__.Notify(602102)
				var err error
				oid, err = tree.ParseDOid(ctx, string(*t), types.RegType)
				if err != nil {
					__antithesis_instrumentation__.Notify(602104)
					return tree.HasNoPrivilege, err
				} else {
					__antithesis_instrumentation__.Notify(602105)
				}
			case *tree.DOid:
				__antithesis_instrumentation__.Notify(602103)
				oid = t
			}
			__antithesis_instrumentation__.Notify(602097)

			typ, err := getNameForArg(ctx, oid, "pg_type", "typname")
			if err != nil {
				__antithesis_instrumentation__.Notify(602106)
				return tree.HasNoPrivilege, err
			} else {
				__antithesis_instrumentation__.Notify(602107)
			}
			__antithesis_instrumentation__.Notify(602098)
			retNull := false
			if typ == "" {
				__antithesis_instrumentation__.Notify(602108)

				retNull = true
			} else {
				__antithesis_instrumentation__.Notify(602109)
			}
			__antithesis_instrumentation__.Notify(602099)

			privs, err := parsePrivilegeStr(args[1], privMap{
				"USAGE":                   {Kind: privilege.USAGE},
				"USAGE WITH GRANT OPTION": {Kind: privilege.USAGE, GrantOption: true},
			})
			if err != nil {
				__antithesis_instrumentation__.Notify(602110)
				return tree.HasNoPrivilege, err
			} else {
				__antithesis_instrumentation__.Notify(602111)
			}
			__antithesis_instrumentation__.Notify(602100)
			if retNull {
				__antithesis_instrumentation__.Notify(602112)
				return tree.ObjectNotFound, nil
			} else {
				__antithesis_instrumentation__.Notify(602113)
			}
			__antithesis_instrumentation__.Notify(602101)

			_ = privs
			return tree.HasPrivilege, nil
		},
	),

	"pg_has_role": makePGPrivilegeInquiryDef(
		"role",
		argTypeOpts{{"role", strOrOidTypes}},
		func(ctx *tree.EvalContext, args tree.Datums, user security.SQLUsername) (tree.HasAnyPrivilegeResult, error) {
			__antithesis_instrumentation__.Notify(602114)
			roleArg := tree.UnwrapDatum(ctx, args[0])
			roleS, err := getNameForArg(ctx, roleArg, "pg_roles", "rolname")
			if err != nil {
				__antithesis_instrumentation__.Notify(602118)
				return tree.HasNoPrivilege, err
			} else {
				__antithesis_instrumentation__.Notify(602119)
			}
			__antithesis_instrumentation__.Notify(602115)

			role := security.MakeSQLUsernameFromPreNormalizedString(roleS)
			if role.Undefined() {
				__antithesis_instrumentation__.Notify(602120)
				switch roleArg.(type) {
				case *tree.DString:
					__antithesis_instrumentation__.Notify(602121)
					return tree.HasNoPrivilege, pgerror.Newf(pgcode.UndefinedObject,
						"role %s does not exist", roleArg)
				case *tree.DOid:
					__antithesis_instrumentation__.Notify(602122)

					return tree.ObjectNotFound, nil
				}
			} else {
				__antithesis_instrumentation__.Notify(602123)
			}
			__antithesis_instrumentation__.Notify(602116)

			privStrs := normalizePrivilegeStr(args[1])
			for _, privStr := range privStrs {
				__antithesis_instrumentation__.Notify(602124)
				var hasAnyPrivilegeResult tree.HasAnyPrivilegeResult
				var err error
				switch privStr {
				case "USAGE":
					__antithesis_instrumentation__.Notify(602127)
					hasAnyPrivilegeResult, err = hasPrivsOfRole(ctx, user, role)
				case "MEMBER":
					__antithesis_instrumentation__.Notify(602128)
					hasAnyPrivilegeResult, err = isMemberOfRole(ctx, user, role)
				case
					"USAGE WITH GRANT OPTION",
					"USAGE WITH ADMIN OPTION",
					"MEMBER WITH GRANT OPTION",
					"MEMBER WITH ADMIN OPTION":
					__antithesis_instrumentation__.Notify(602129)
					hasAnyPrivilegeResult, err = isAdminOfRole(ctx, user, role)
				default:
					__antithesis_instrumentation__.Notify(602130)
					return tree.HasNoPrivilege, pgerror.Newf(pgcode.InvalidParameterValue,
						"unrecognized privilege type: %q", privStr)
				}
				__antithesis_instrumentation__.Notify(602125)
				if err != nil {
					__antithesis_instrumentation__.Notify(602131)
					return tree.HasNoPrivilege, err
				} else {
					__antithesis_instrumentation__.Notify(602132)
				}
				__antithesis_instrumentation__.Notify(602126)
				if hasAnyPrivilegeResult == tree.HasPrivilege {
					__antithesis_instrumentation__.Notify(602133)
					return hasAnyPrivilegeResult, nil
				} else {
					__antithesis_instrumentation__.Notify(602134)
				}
			}
			__antithesis_instrumentation__.Notify(602117)
			return tree.HasNoPrivilege, nil
		},
	),

	"current_setting": makeBuiltin(
		tree.FunctionProperties{
			Category:         categorySystemInfo,
			DistsqlBlocklist: true,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"setting_name", types.String}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(602135)
				return getSessionVar(ctx, string(tree.MustBeDString(args[0])), false)
			},
			Info:       categorySystemInfo,
			Volatility: tree.VolatilityStable,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"setting_name", types.String}, {"missing_ok", types.Bool}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(602136)
				return getSessionVar(ctx, string(tree.MustBeDString(args[0])), bool(tree.MustBeDBool(args[1])))
			},
			Info:       categorySystemInfo,
			Volatility: tree.VolatilityStable,
		},
	),

	"set_config": makeBuiltin(
		tree.FunctionProperties{
			Category:         categorySystemInfo,
			DistsqlBlocklist: true,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"setting_name", types.String}, {"new_value", types.String}, {"is_local", types.Bool}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(602137)
				varName := string(tree.MustBeDString(args[0]))
				newValue := string(tree.MustBeDString(args[1]))
				err := setSessionVar(ctx, varName, newValue, bool(tree.MustBeDBool(args[2])))
				if err != nil {
					__antithesis_instrumentation__.Notify(602139)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(602140)
				}
				__antithesis_instrumentation__.Notify(602138)
				return getSessionVar(ctx, varName, false)
			},
			Info:       categorySystemInfo,
			Volatility: tree.VolatilityVolatile,
		},
	),

	"inet_client_addr": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.INet),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(602141)
				return tree.NewDIPAddr(tree.DIPAddr{IPAddr: ipaddr.IPAddr{}}), nil
			},
			Info:       notUsableInfo,
			Volatility: tree.VolatilityStable,
		},
	),

	"inet_client_port": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(602142)
				return tree.DZero, nil
			},
			Info:       notUsableInfo,
			Volatility: tree.VolatilityStable,
		},
	),

	"inet_server_addr": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.INet),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(602143)
				return tree.NewDIPAddr(tree.DIPAddr{IPAddr: ipaddr.IPAddr{}}), nil
			},
			Info:       notUsableInfo,
			Volatility: tree.VolatilityStable,
		},
	),

	"inet_server_port": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(602144)
				return tree.DZero, nil
			},
			Info:       notUsableInfo,
			Volatility: tree.VolatilityStable,
		},
	),

	"pg_column_size": makeBuiltin(defProps(),
		tree.Overload{
			Types: tree.VariadicType{
				VarType: types.Any,
			},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(602145)
				var totalSize int
				for _, arg := range args {
					__antithesis_instrumentation__.Notify(602147)
					encodeTableValue, err := valueside.Encode(nil, valueside.NoColumnID, arg, nil)
					if err != nil {
						__antithesis_instrumentation__.Notify(602149)
						return tree.DNull, err
					} else {
						__antithesis_instrumentation__.Notify(602150)
					}
					__antithesis_instrumentation__.Notify(602148)
					totalSize += len(encodeTableValue)
				}
				__antithesis_instrumentation__.Notify(602146)
				return tree.NewDInt(tree.DInt(totalSize)), nil
			},
			Info:       "Return size in bytes of the column provided as an argument",
			Volatility: tree.VolatilityImmutable,
		}),

	"information_schema._pg_truetypid": pgTrueTypImpl("atttypid", "typbasetype", types.Oid),

	"information_schema._pg_truetypmod": pgTrueTypImpl("atttypmod", "typtypmod", types.Int4),

	"information_schema._pg_char_max_length": makeBuiltin(defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"typid", types.Oid},
				{"typmod", types.Int4},
			},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(602151)
				typid := oid.Oid(args[0].(*tree.DOid).DInt)
				typmod := *args[1].(*tree.DInt)
				if typmod == -1 {
					__antithesis_instrumentation__.Notify(602153)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(602154)
					if typid == oid.T_bpchar || func() bool {
						__antithesis_instrumentation__.Notify(602155)
						return typid == oid.T_varchar == true
					}() == true {
						__antithesis_instrumentation__.Notify(602156)
						return tree.NewDInt(typmod - 4), nil
					} else {
						__antithesis_instrumentation__.Notify(602157)
						if typid == oid.T_bit || func() bool {
							__antithesis_instrumentation__.Notify(602158)
							return typid == oid.T_varbit == true
						}() == true {
							__antithesis_instrumentation__.Notify(602159)
							return tree.NewDInt(typmod), nil
						} else {
							__antithesis_instrumentation__.Notify(602160)
						}
					}
				}
				__antithesis_instrumentation__.Notify(602152)
				return tree.DNull, nil
			},
			Info:       notUsableInfo,
			Volatility: tree.VolatilityImmutable,
		},
	),

	"information_schema._pg_index_position": makeBuiltin(defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"oid", types.Oid},
				{"col", types.Int2},
			},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(602161)
				r, err := ctx.Planner.QueryRowEx(
					ctx.Ctx(), "information_schema._pg_index_position",
					ctx.Txn,
					sessiondata.NoSessionDataOverride,
					`SELECT (ss.a).n FROM
					  (SELECT information_schema._pg_expandarray(indkey) AS a
					   FROM pg_catalog.pg_index WHERE indexrelid = $1) ss
            WHERE (ss.a).x = $2`,
					args[0], args[1])
				if err != nil {
					__antithesis_instrumentation__.Notify(602164)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(602165)
				}
				__antithesis_instrumentation__.Notify(602162)
				if len(r) == 0 {
					__antithesis_instrumentation__.Notify(602166)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(602167)
				}
				__antithesis_instrumentation__.Notify(602163)
				return r[0], nil
			},
			Info:       notUsableInfo,
			Volatility: tree.VolatilityStable,
		},
	),

	"information_schema._pg_numeric_precision": makeBuiltin(tree.FunctionProperties{Category: categorySystemInfo},
		tree.Overload{
			Types: tree.ArgTypes{
				{"typid", types.Oid},
				{"typmod", types.Int4},
			},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(602168)
				typid := oid.Oid(tree.MustBeDOid(args[0]).DInt)
				typmod := tree.MustBeDInt(args[1])
				switch typid {
				case oid.T_int2:
					__antithesis_instrumentation__.Notify(602170)
					return tree.NewDInt(16), nil
				case oid.T_int4:
					__antithesis_instrumentation__.Notify(602171)
					return tree.NewDInt(32), nil
				case oid.T_int8:
					__antithesis_instrumentation__.Notify(602172)
					return tree.NewDInt(64), nil
				case oid.T_numeric:
					__antithesis_instrumentation__.Notify(602173)
					if typmod != -1 {
						__antithesis_instrumentation__.Notify(602178)

						return tree.NewDInt(((typmod - 4) >> 16) & 65535), nil
					} else {
						__antithesis_instrumentation__.Notify(602179)
					}
					__antithesis_instrumentation__.Notify(602174)
					return tree.DNull, nil
				case oid.T_float4:
					__antithesis_instrumentation__.Notify(602175)
					return tree.NewDInt(24), nil
				case oid.T_float8:
					__antithesis_instrumentation__.Notify(602176)
					return tree.NewDInt(53), nil
				default:
					__antithesis_instrumentation__.Notify(602177)
				}
				__antithesis_instrumentation__.Notify(602169)
				return tree.DNull, nil
			},
			Info:       "Returns the precision of the given type with type modifier",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"information_schema._pg_numeric_precision_radix": makeBuiltin(tree.FunctionProperties{Category: categorySystemInfo},
		tree.Overload{
			Types: tree.ArgTypes{
				{"typid", types.Oid},
				{"typmod", types.Int4},
			},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(602180)
				typid := oid.Oid(tree.MustBeDOid(args[0]).DInt)
				if typid == oid.T_int2 || func() bool {
					__antithesis_instrumentation__.Notify(602181)
					return typid == oid.T_int4 == true
				}() == true || func() bool {
					__antithesis_instrumentation__.Notify(602182)
					return typid == oid.T_int8 == true
				}() == true || func() bool {
					__antithesis_instrumentation__.Notify(602183)
					return typid == oid.T_float4 == true
				}() == true || func() bool {
					__antithesis_instrumentation__.Notify(602184)
					return typid == oid.T_float8 == true
				}() == true {
					__antithesis_instrumentation__.Notify(602185)
					return tree.NewDInt(2), nil
				} else {
					__antithesis_instrumentation__.Notify(602186)
					if typid == oid.T_numeric {
						__antithesis_instrumentation__.Notify(602187)
						return tree.NewDInt(10), nil
					} else {
						__antithesis_instrumentation__.Notify(602188)
						return tree.DNull, nil
					}
				}
			},
			Info:       "Returns the radix of the given type with type modifier",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"information_schema._pg_numeric_scale": makeBuiltin(tree.FunctionProperties{Category: categorySystemInfo},
		tree.Overload{
			Types: tree.ArgTypes{
				{"typid", types.Oid},
				{"typmod", types.Int4},
			},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(602189)
				typid := oid.Oid(tree.MustBeDOid(args[0]).DInt)
				typmod := tree.MustBeDInt(args[1])
				if typid == oid.T_int2 || func() bool {
					__antithesis_instrumentation__.Notify(602191)
					return typid == oid.T_int4 == true
				}() == true || func() bool {
					__antithesis_instrumentation__.Notify(602192)
					return typid == oid.T_int8 == true
				}() == true {
					__antithesis_instrumentation__.Notify(602193)
					return tree.NewDInt(0), nil
				} else {
					__antithesis_instrumentation__.Notify(602194)
					if typid == oid.T_numeric {
						__antithesis_instrumentation__.Notify(602195)
						if typmod == -1 {
							__antithesis_instrumentation__.Notify(602197)
							return tree.DNull, nil
						} else {
							__antithesis_instrumentation__.Notify(602198)
						}
						__antithesis_instrumentation__.Notify(602196)

						return tree.NewDInt((typmod - 4) & 65535), nil
					} else {
						__antithesis_instrumentation__.Notify(602199)
					}
				}
				__antithesis_instrumentation__.Notify(602190)
				return tree.DNull, nil
			},
			Info:       "Returns the scale of the given type with type modifier",
			Volatility: tree.VolatilityImmutable,
		},
	),
}

func getSessionVar(ctx *tree.EvalContext, settingName string, missingOk bool) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(602200)
	if ctx.SessionAccessor == nil {
		__antithesis_instrumentation__.Notify(602204)
		return nil, errors.AssertionFailedf("session accessor not set")
	} else {
		__antithesis_instrumentation__.Notify(602205)
	}
	__antithesis_instrumentation__.Notify(602201)
	ok, s, err := ctx.SessionAccessor.GetSessionVar(ctx.Context, settingName, missingOk)
	if err != nil {
		__antithesis_instrumentation__.Notify(602206)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(602207)
	}
	__antithesis_instrumentation__.Notify(602202)
	if !ok {
		__antithesis_instrumentation__.Notify(602208)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(602209)
	}
	__antithesis_instrumentation__.Notify(602203)
	return tree.NewDString(s), nil
}

func setSessionVar(ctx *tree.EvalContext, settingName, newVal string, isLocal bool) error {
	__antithesis_instrumentation__.Notify(602210)
	if ctx.SessionAccessor == nil {
		__antithesis_instrumentation__.Notify(602212)
		return errors.AssertionFailedf("session accessor not set")
	} else {
		__antithesis_instrumentation__.Notify(602213)
	}
	__antithesis_instrumentation__.Notify(602211)
	return ctx.SessionAccessor.SetSessionVar(ctx.Context, settingName, newVal, isLocal)
}

func getCatalogOidForComments(catalogName string) (id int, ok bool) {
	__antithesis_instrumentation__.Notify(602214)
	switch catalogName {
	case "pg_class":
		__antithesis_instrumentation__.Notify(602215)
		return catconstants.PgCatalogClassTableID, true
	case "pg_database":
		__antithesis_instrumentation__.Notify(602216)
		return catconstants.PgCatalogDatabaseTableID, true
	case "pg_description":
		__antithesis_instrumentation__.Notify(602217)
		return catconstants.PgCatalogDescriptionTableID, true
	case "pg_constraint":
		__antithesis_instrumentation__.Notify(602218)
		return catconstants.PgCatalogConstraintTableID, true
	default:
		__antithesis_instrumentation__.Notify(602219)

		return 0, false
	}
}

func getPgObjDesc(ctx *tree.EvalContext, catalogName string, oid int) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(602220)
	classOidFilter := ""
	if catalogName != "" {
		__antithesis_instrumentation__.Notify(602224)
		classOid, ok := getCatalogOidForComments(catalogName)
		if !ok {
			__antithesis_instrumentation__.Notify(602226)

			return tree.DNull, nil
		} else {
			__antithesis_instrumentation__.Notify(602227)
		}
		__antithesis_instrumentation__.Notify(602225)
		classOidFilter = fmt.Sprintf("AND classoid = %d", classOid)
	} else {
		__antithesis_instrumentation__.Notify(602228)
	}
	__antithesis_instrumentation__.Notify(602221)
	r, err := ctx.Planner.QueryRowEx(
		ctx.Ctx(), "pg_get_objdesc", ctx.Txn,
		sessiondata.NoSessionDataOverride,
		fmt.Sprintf(`
SELECT description
  FROM pg_catalog.pg_description
 WHERE objoid = %[1]d
   AND objsubid = 0
   %[2]s
 LIMIT 1`,
			oid,
			classOidFilter,
		))
	if err != nil {
		__antithesis_instrumentation__.Notify(602229)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(602230)
	}
	__antithesis_instrumentation__.Notify(602222)
	if len(r) == 0 {
		__antithesis_instrumentation__.Notify(602231)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(602232)
	}
	__antithesis_instrumentation__.Notify(602223)
	return r[0], nil
}

func databaseHasPrivilegeSpecifier(databaseArg tree.Datum) (tree.HasPrivilegeSpecifier, error) {
	__antithesis_instrumentation__.Notify(602233)
	var specifier tree.HasPrivilegeSpecifier
	switch t := databaseArg.(type) {
	case *tree.DString:
		__antithesis_instrumentation__.Notify(602235)
		s := string(*t)
		specifier.DatabaseName = &s
	case *tree.DOid:
		__antithesis_instrumentation__.Notify(602236)
		oid := oid.Oid(t.DInt)
		specifier.DatabaseOID = &oid
	default:
		__antithesis_instrumentation__.Notify(602237)
		return specifier, errors.AssertionFailedf("unknown privilege specifier: %#v", databaseArg)
	}
	__antithesis_instrumentation__.Notify(602234)
	return specifier, nil
}

func tableHasPrivilegeSpecifier(
	tableArg tree.Datum, isSequence bool,
) (tree.HasPrivilegeSpecifier, error) {
	__antithesis_instrumentation__.Notify(602238)
	specifier := tree.HasPrivilegeSpecifier{
		IsSequence: &isSequence,
	}
	switch t := tableArg.(type) {
	case *tree.DString:
		__antithesis_instrumentation__.Notify(602240)
		s := string(*t)
		specifier.TableName = &s
	case *tree.DOid:
		__antithesis_instrumentation__.Notify(602241)
		oid := oid.Oid(t.DInt)
		specifier.TableOID = &oid
	default:
		__antithesis_instrumentation__.Notify(602242)
		return specifier, errors.AssertionFailedf("unknown privilege specifier: %#v", tableArg)
	}
	__antithesis_instrumentation__.Notify(602239)
	return specifier, nil
}

func columnHasPrivilegeSpecifier(
	tableArg tree.Datum, colArg tree.Datum,
) (tree.HasPrivilegeSpecifier, error) {
	__antithesis_instrumentation__.Notify(602243)
	specifier, err := tableHasPrivilegeSpecifier(tableArg, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(602246)
		return specifier, err
	} else {
		__antithesis_instrumentation__.Notify(602247)
	}
	__antithesis_instrumentation__.Notify(602244)
	switch t := colArg.(type) {
	case *tree.DString:
		__antithesis_instrumentation__.Notify(602248)
		n := tree.Name(*t)
		specifier.ColumnName = &n
	case *tree.DInt:
		__antithesis_instrumentation__.Notify(602249)
		attNum := uint32(*t)
		specifier.ColumnAttNum = &attNum
	default:
		__antithesis_instrumentation__.Notify(602250)
		return specifier, errors.AssertionFailedf("unexpected arg type %T", t)
	}
	__antithesis_instrumentation__.Notify(602245)
	return specifier, nil
}

func schemaHasPrivilegeSpecifier(
	ctx *tree.EvalContext, schemaArg tree.Datum, databaseName string,
) (tree.HasPrivilegeSpecifier, error) {
	__antithesis_instrumentation__.Notify(602251)
	specifier := tree.HasPrivilegeSpecifier{
		SchemaDatabaseName: &databaseName,
	}
	var schemaIsRequired bool
	switch t := schemaArg.(type) {
	case *tree.DString:
		__antithesis_instrumentation__.Notify(602253)
		s := string(*t)
		specifier.SchemaName = &s
		schemaIsRequired = true
	case *tree.DOid:
		__antithesis_instrumentation__.Notify(602254)
		schemaName, err := getNameForArg(ctx, schemaArg, "pg_namespace", "nspname")
		if err != nil {
			__antithesis_instrumentation__.Notify(602257)
			return specifier, err
		} else {
			__antithesis_instrumentation__.Notify(602258)
		}
		__antithesis_instrumentation__.Notify(602255)
		specifier.SchemaName = &schemaName
	default:
		__antithesis_instrumentation__.Notify(602256)
		return specifier, errors.AssertionFailedf("unknown privilege specifier: %#v", schemaArg)
	}
	__antithesis_instrumentation__.Notify(602252)
	specifier.SchemaIsRequired = &schemaIsRequired
	return specifier, nil
}

func pgTrueTypImpl(attrField, typField string, retType *types.T) builtinDefinition {
	__antithesis_instrumentation__.Notify(602259)
	return makeBuiltin(defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"pg_attribute", types.AnyTuple},
				{"pg_type", types.AnyTuple},
			},
			ReturnType: tree.FixedReturnType(retType),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(602260)

				fieldIdx := func(t *tree.DTuple, field string) int {
					__antithesis_instrumentation__.Notify(602264)
					for i, label := range t.ResolvedType().TupleLabels() {
						__antithesis_instrumentation__.Notify(602266)
						if label == field {
							__antithesis_instrumentation__.Notify(602267)
							return i
						} else {
							__antithesis_instrumentation__.Notify(602268)
						}
					}
					__antithesis_instrumentation__.Notify(602265)
					return -1
				}
				__antithesis_instrumentation__.Notify(602261)

				pgAttr, pgType := args[0].(*tree.DTuple), args[1].(*tree.DTuple)
				pgAttrFieldIdx := fieldIdx(pgAttr, attrField)
				pgTypeTypeIdx := fieldIdx(pgType, "typtype")
				pgTypeFieldIdx := fieldIdx(pgType, typField)
				if pgAttrFieldIdx == -1 || func() bool {
					__antithesis_instrumentation__.Notify(602269)
					return pgTypeTypeIdx == -1 == true
				}() == true || func() bool {
					__antithesis_instrumentation__.Notify(602270)
					return pgTypeFieldIdx == -1 == true
				}() == true {
					__antithesis_instrumentation__.Notify(602271)
					return nil, pgerror.Newf(pgcode.UndefinedFunction,
						"No function matches the given name and argument types.")
				} else {
					__antithesis_instrumentation__.Notify(602272)
				}
				__antithesis_instrumentation__.Notify(602262)

				pgAttrField := pgAttr.D[pgAttrFieldIdx]
				pgTypeType := pgType.D[pgTypeTypeIdx].(*tree.DString)
				pgTypeField := pgType.D[pgTypeFieldIdx]

				if *pgTypeType == "d" {
					__antithesis_instrumentation__.Notify(602273)
					return pgTypeField, nil
				} else {
					__antithesis_instrumentation__.Notify(602274)
				}
				__antithesis_instrumentation__.Notify(602263)
				return pgAttrField, nil
			},
			Info:       notUsableInfo,
			Volatility: tree.VolatilityImmutable,
		},
	)
}

func hasPrivsOfRole(
	ctx *tree.EvalContext, user, role security.SQLUsername,
) (tree.HasAnyPrivilegeResult, error) {
	__antithesis_instrumentation__.Notify(602275)
	return isMemberOfRole(ctx, user, role)
}

func isMemberOfRole(
	ctx *tree.EvalContext, user, role security.SQLUsername,
) (tree.HasAnyPrivilegeResult, error) {
	__antithesis_instrumentation__.Notify(602276)

	if user == role {
		__antithesis_instrumentation__.Notify(602281)
		return tree.HasPrivilege, nil
	} else {
		__antithesis_instrumentation__.Notify(602282)
	}
	__antithesis_instrumentation__.Notify(602277)

	if isSuper, err := ctx.Planner.UserHasAdminRole(ctx.Context, user); err != nil {
		__antithesis_instrumentation__.Notify(602283)
		return tree.HasNoPrivilege, err
	} else {
		__antithesis_instrumentation__.Notify(602284)
		if isSuper {
			__antithesis_instrumentation__.Notify(602285)
			return tree.HasPrivilege, nil
		} else {
			__antithesis_instrumentation__.Notify(602286)
		}
	}
	__antithesis_instrumentation__.Notify(602278)

	allRoleMemberships, err := ctx.Planner.MemberOfWithAdminOption(ctx.Context, user)
	if err != nil {
		__antithesis_instrumentation__.Notify(602287)
		return tree.HasNoPrivilege, err
	} else {
		__antithesis_instrumentation__.Notify(602288)
	}
	__antithesis_instrumentation__.Notify(602279)
	_, member := allRoleMemberships[role]
	if member {
		__antithesis_instrumentation__.Notify(602289)
		return tree.HasPrivilege, nil
	} else {
		__antithesis_instrumentation__.Notify(602290)
	}
	__antithesis_instrumentation__.Notify(602280)
	return tree.HasNoPrivilege, nil
}

func isAdminOfRole(
	ctx *tree.EvalContext, user, role security.SQLUsername,
) (tree.HasAnyPrivilegeResult, error) {
	__antithesis_instrumentation__.Notify(602291)

	if isSuper, err := ctx.Planner.UserHasAdminRole(ctx.Context, user); err != nil {
		__antithesis_instrumentation__.Notify(602296)
		return tree.HasNoPrivilege, err
	} else {
		__antithesis_instrumentation__.Notify(602297)
		if isSuper {
			__antithesis_instrumentation__.Notify(602298)
			return tree.HasPrivilege, nil
		} else {
			__antithesis_instrumentation__.Notify(602299)
		}
	}
	__antithesis_instrumentation__.Notify(602292)

	if user == role {
		__antithesis_instrumentation__.Notify(602300)

		if isSessionUser := user == ctx.SessionData().SessionUser(); isSessionUser {
			__antithesis_instrumentation__.Notify(602302)
			return tree.HasPrivilege, nil
		} else {
			__antithesis_instrumentation__.Notify(602303)
		}
		__antithesis_instrumentation__.Notify(602301)
		return tree.HasNoPrivilege, nil
	} else {
		__antithesis_instrumentation__.Notify(602304)
	}
	__antithesis_instrumentation__.Notify(602293)

	allRoleMemberships, err := ctx.Planner.MemberOfWithAdminOption(ctx.Context, user)
	if err != nil {
		__antithesis_instrumentation__.Notify(602305)
		return tree.HasNoPrivilege, err
	} else {
		__antithesis_instrumentation__.Notify(602306)
	}
	__antithesis_instrumentation__.Notify(602294)
	if isAdmin := allRoleMemberships[role]; isAdmin {
		__antithesis_instrumentation__.Notify(602307)
		return tree.HasPrivilege, nil
	} else {
		__antithesis_instrumentation__.Notify(602308)
	}
	__antithesis_instrumentation__.Notify(602295)
	return tree.HasNoPrivilege, nil
}
