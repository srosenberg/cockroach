package sqlsmith

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"math/rand"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/sql/randgen"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

type Setup func(*rand.Rand) []string

const RandTableSetupName = "rand-tables"

var Setups = map[string]Setup{
	"empty": wrapCommonSetup(stringSetup("")),

	"seed":              wrapCommonSetup(stringSetup(seedTable)),
	"seed-multi-region": wrapCommonSetup(stringSetup(multiregionSeed)),
	RandTableSetupName:  wrapCommonSetup(randTables),
}

func wrapCommonSetup(setupFn Setup) Setup {
	__antithesis_instrumentation__.Notify(69865)
	return func(r *rand.Rand) []string {
		__antithesis_instrumentation__.Notify(69866)
		return setupFn(r)
	}
}

var setupNames = func() []string {
	__antithesis_instrumentation__.Notify(69867)
	var ret []string
	for k := range Setups {
		__antithesis_instrumentation__.Notify(69869)
		ret = append(ret, k)
	}
	__antithesis_instrumentation__.Notify(69868)
	sort.Strings(ret)
	return ret
}()

func RandSetup(r *rand.Rand) string {
	__antithesis_instrumentation__.Notify(69870)
	n := r.Intn(len(setupNames))
	return setupNames[n]
}

func stringSetup(s string) Setup {
	__antithesis_instrumentation__.Notify(69871)
	return func(*rand.Rand) []string {
		__antithesis_instrumentation__.Notify(69872)
		return []string{s}
	}
}

func randTables(r *rand.Rand) []string {
	__antithesis_instrumentation__.Notify(69873)
	return randTablesN(r, r.Intn(5)+1)
}

func randTablesN(r *rand.Rand, n int) []string {
	__antithesis_instrumentation__.Notify(69874)
	var stmts []string

	stmts = append(stmts, `SET CLUSTER SETTING sql.stats.automatic_collection.enabled = false;`)
	stmts = append(stmts, `SET CLUSTER SETTING sql.stats.histogram_collection.enabled = false;`)

	createTableStatements := randgen.RandCreateTables(r, "table", n,
		randgen.StatisticsMutator,
		randgen.PartialIndexMutator,
		randgen.ForeignKeyMutator,
	)

	for _, stmt := range createTableStatements {
		__antithesis_instrumentation__.Notify(69877)
		stmts = append(stmts, tree.SerializeForDisplay(stmt))
	}
	__antithesis_instrumentation__.Notify(69875)

	numTypes := r.Intn(5) + 1
	for i := 0; i < numTypes; i++ {
		__antithesis_instrumentation__.Notify(69878)
		name := fmt.Sprintf("rand_typ_%d", i)
		stmt := randgen.RandCreateType(r, name, letters)
		stmts = append(stmts, stmt.String())
	}
	__antithesis_instrumentation__.Notify(69876)
	return stmts
}

const (
	seedTable = `
BEGIN; CREATE TYPE greeting AS ENUM ('hello', 'howdy', 'hi', 'good day', 'morning'); COMMIT;
BEGIN;
CREATE TABLE IF NOT EXISTS seed AS
	SELECT
		g::INT2 AS _int2,
		g::INT4 AS _int4,
		g::INT8 AS _int8,
		g::FLOAT4 AS _float4,
		g::FLOAT8 AS _float8,
		'2001-01-01'::DATE + g AS _date,
		'2001-01-01'::TIMESTAMP + g * '1 day'::INTERVAL AS _timestamp,
		'2001-01-01'::TIMESTAMPTZ + g * '1 day'::INTERVAL AS _timestamptz,
		g * '1 day'::INTERVAL AS _interval,
		g % 2 = 1 AS _bool,
		g::DECIMAL AS _decimal,
		g::STRING AS _string,
		g::STRING::BYTES AS _bytes,
		substring('00000000-0000-0000-0000-' || g::STRING || '00000000000', 1, 36)::UUID AS _uuid,
		'0.0.0.0'::INET + g AS _inet,
		g::STRING::JSONB AS _jsonb,
		enum_range('hello'::greeting)[g] as _enum
	FROM
		generate_series(1, 5) AS g;
COMMIT;

INSERT INTO seed DEFAULT VALUES;
CREATE INDEX on seed (_int8, _float8, _date);
CREATE INVERTED INDEX on seed (_jsonb);
`

	multiregionSeed = `
CREATE TABLE IF NOT EXISTS seed_mr_table AS
	SELECT
		g::INT2 AS _int2,
		g::INT4 AS _int4,
		g::INT8 AS _int8,
		g::FLOAT8 AS _float8,
		'2001-01-01'::DATE + g AS _date,
		'2001-01-01'::TIMESTAMP + g * '1 day'::INTERVAL AS _timestamp,
		'2001-01-01'::TIMESTAMPTZ + g * '1 day'::INTERVAL AS _timestamptz,
		g * '1 day'::INTERVAL AS _interval,
		g % 2 = 1 AS _bool,
		g::DECIMAL AS _decimal,
		g::STRING AS _string,
		g::STRING::BYTES AS _bytes,
		substring('00000000-0000-0000-0000-' || g::STRING || '00000000000', 1, 36)::UUID AS _uuid
	FROM
		generate_series(1, 5) AS g;
`
)

type SettingFunc func(*rand.Rand) Setting

type Setting struct {
	Options []SmitherOption
	Mode    ExecMode
}

type ExecMode int

const (
	NoParallel ExecMode = iota

	Parallel
)

var Settings = map[string]SettingFunc{
	"default":           staticSetting(Parallel),
	"no-mutations":      staticSetting(Parallel, DisableMutations()),
	"no-ddl":            staticSetting(NoParallel, DisableDDLs()),
	"default+rand":      randSetting(Parallel),
	"no-mutations+rand": randSetting(Parallel, DisableMutations()),
	"no-ddl+rand":       randSetting(NoParallel, DisableDDLs()),
	"ddl-nodrop":        randSetting(NoParallel, OnlyNoDropDDLs()),
	"multi-region":      randSetting(Parallel, MultiRegionDDLs()),
}

var settingNames = func() []string {
	__antithesis_instrumentation__.Notify(69879)
	var ret []string
	for k := range Settings {
		__antithesis_instrumentation__.Notify(69881)
		ret = append(ret, k)
	}
	__antithesis_instrumentation__.Notify(69880)
	sort.Strings(ret)
	return ret
}()

func RandSetting(r *rand.Rand) string {
	__antithesis_instrumentation__.Notify(69882)
	return settingNames[r.Intn(len(settingNames))]
}

func staticSetting(mode ExecMode, opts ...SmitherOption) SettingFunc {
	__antithesis_instrumentation__.Notify(69883)
	return func(*rand.Rand) Setting {
		__antithesis_instrumentation__.Notify(69884)
		return Setting{
			Options: opts,
			Mode:    mode,
		}
	}
}

func randSetting(mode ExecMode, staticOpts ...SmitherOption) SettingFunc {
	__antithesis_instrumentation__.Notify(69885)
	return func(r *rand.Rand) Setting {
		__antithesis_instrumentation__.Notify(69886)

		opts := append([]SmitherOption(nil), randOptions...)
		r.Shuffle(len(opts), func(i, j int) {
			__antithesis_instrumentation__.Notify(69888)
			opts[i], opts[j] = opts[j], opts[i]
		})
		__antithesis_instrumentation__.Notify(69887)

		opts = opts[:r.Intn(len(opts)+1)]
		opts = append(opts, staticOpts...)
		return Setting{
			Options: opts,
			Mode:    mode,
		}
	}
}

var randOptions = []SmitherOption{
	AvoidConsts(),
	CompareMode(),
	DisableLimits(),
	DisableWindowFuncs(),
	DisableWith(),
	PostgresMode(),
	SimpleDatums(),
}
