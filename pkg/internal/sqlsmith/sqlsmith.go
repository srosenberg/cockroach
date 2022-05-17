package sqlsmith

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	gosql "database/sql"
	"fmt"
	"math/rand"
	"net/http/httptest"
	"regexp"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

const retryCount = 20

type Smither struct {
	rnd              *rand.Rand
	db               *gosql.DB
	lock             syncutil.RWMutex
	dbName           string
	schemas          []*schemaRef
	tables           []*tableRef
	columns          map[tree.TableName]map[tree.Name]*tree.ColumnTableDef
	indexes          map[tree.TableName]map[tree.Name]*tree.CreateIndex
	nameCounts       map[string]int
	activeSavepoints []string
	types            *typeInfo

	stmtWeights, alterWeights          []statementWeight
	stmtSampler, alterSampler          *statementSampler
	tableExprWeights                   []tableExprWeight
	tableExprSampler                   *tableExprSampler
	selectStmtWeights                  []selectStatementWeight
	selectStmtSampler                  *selectStatementSampler
	scalarExprWeights, boolExprWeights []scalarExprWeight
	scalarExprSampler, boolExprSampler *scalarExprSampler

	disableWith        bool
	disableImpureFns   bool
	disableLimits      bool
	disableWindowFuncs bool
	simpleDatums       bool
	avoidConsts        bool
	outputSort         bool
	postgres           bool
	ignoreFNs          []*regexp.Regexp
	complexity         float64

	bulkSrv     *httptest.Server
	bulkFiles   map[string][]byte
	bulkBackups map[string]tree.TargetList
	bulkExports []string
}

type (
	statement       func(*Smither) (tree.Statement, bool)
	tableExpr       func(s *Smither, refs colRefs, forJoin bool) (tree.TableExpr, colRefs, bool)
	selectStatement func(s *Smither, desiredTypes []*types.T, refs colRefs, withTables tableRefs) (tree.SelectStatement, colRefs, bool)
	scalarExpr      func(*Smither, Context, *types.T, colRefs) (expr tree.TypedExpr, ok bool)
)

func NewSmither(db *gosql.DB, rnd *rand.Rand, opts ...SmitherOption) (*Smither, error) {
	__antithesis_instrumentation__.Notify(69889)
	s := &Smither{
		rnd:        rnd,
		db:         db,
		nameCounts: map[string]int{},

		stmtWeights:       allStatements,
		alterWeights:      alters,
		tableExprWeights:  allTableExprs,
		selectStmtWeights: selectStmts,
		scalarExprWeights: scalars,
		boolExprWeights:   bools,

		complexity: 0.2,
	}
	for _, opt := range opts {
		__antithesis_instrumentation__.Notify(69892)
		opt.Apply(s)
	}
	__antithesis_instrumentation__.Notify(69890)
	s.stmtSampler = newWeightedStatementSampler(s.stmtWeights, rnd.Int63())
	s.alterSampler = newWeightedStatementSampler(s.alterWeights, rnd.Int63())
	s.tableExprSampler = newWeightedTableExprSampler(s.tableExprWeights, rnd.Int63())
	s.selectStmtSampler = newWeightedSelectStatementSampler(s.selectStmtWeights, rnd.Int63())
	s.scalarExprSampler = newWeightedScalarExprSampler(s.scalarExprWeights, rnd.Int63())
	s.boolExprSampler = newWeightedScalarExprSampler(s.boolExprWeights, rnd.Int63())
	s.enableBulkIO()
	row := s.db.QueryRow("SELECT current_database()")
	if err := row.Scan(&s.dbName); err != nil {
		__antithesis_instrumentation__.Notify(69893)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(69894)
	}
	__antithesis_instrumentation__.Notify(69891)
	return s, s.ReloadSchemas()
}

func (s *Smither) Close() {
	__antithesis_instrumentation__.Notify(69895)
	if s.bulkSrv != nil {
		__antithesis_instrumentation__.Notify(69896)
		s.bulkSrv.Close()
	} else {
		__antithesis_instrumentation__.Notify(69897)
	}
}

var prettyCfg = func() tree.PrettyCfg {
	__antithesis_instrumentation__.Notify(69898)
	cfg := tree.DefaultPrettyCfg()
	cfg.LineWidth = 120
	cfg.Simplify = false
	return cfg
}()

func (s *Smither) Generate() string {
	__antithesis_instrumentation__.Notify(69899)
	i := 0
	for {
		__antithesis_instrumentation__.Notify(69900)
		stmt, ok := s.makeStmt()
		if !ok {
			__antithesis_instrumentation__.Notify(69902)
			i++
			if i > 1000 {
				__antithesis_instrumentation__.Notify(69904)
				panic("exhausted generation attempts")
			} else {
				__antithesis_instrumentation__.Notify(69905)
			}
			__antithesis_instrumentation__.Notify(69903)
			continue
		} else {
			__antithesis_instrumentation__.Notify(69906)
		}
		__antithesis_instrumentation__.Notify(69901)
		i = 0
		return prettyCfg.Pretty(stmt)
	}
}

func (s *Smither) GenerateExpr() tree.TypedExpr {
	__antithesis_instrumentation__.Notify(69907)
	return makeScalar(s, s.randScalarType(), nil)
}

func (s *Smither) name(prefix string) tree.Name {
	__antithesis_instrumentation__.Notify(69908)
	s.lock.Lock()
	s.nameCounts[prefix]++
	count := s.nameCounts[prefix]
	s.lock.Unlock()
	return tree.Name(fmt.Sprintf("%s_%d", prefix, count))
}

type SmitherOption interface {
	Apply(*Smither)
	String() string
}

func simpleOption(name string, apply func(s *Smither)) func() SmitherOption {
	__antithesis_instrumentation__.Notify(69909)
	return func() SmitherOption {
		__antithesis_instrumentation__.Notify(69910)
		return option{
			name:  name,
			apply: apply,
		}
	}
}

func multiOption(name string, opts ...SmitherOption) func() SmitherOption {
	__antithesis_instrumentation__.Notify(69911)
	var sb strings.Builder
	sb.WriteString(name)
	sb.WriteString("(")
	delim := ""
	for _, opt := range opts {
		__antithesis_instrumentation__.Notify(69913)
		sb.WriteString(delim)
		delim = ", "
		sb.WriteString(opt.String())
	}
	__antithesis_instrumentation__.Notify(69912)
	sb.WriteString(")")
	return func() SmitherOption {
		__antithesis_instrumentation__.Notify(69914)
		return option{
			name: sb.String(),
			apply: func(s *Smither) {
				__antithesis_instrumentation__.Notify(69915)
				for _, opt := range opts {
					__antithesis_instrumentation__.Notify(69916)
					opt.Apply(s)
				}
			},
		}
	}
}

type option struct {
	name  string
	apply func(s *Smither)
}

func (o option) String() string {
	__antithesis_instrumentation__.Notify(69917)
	return o.name
}

func (o option) Apply(s *Smither) {
	__antithesis_instrumentation__.Notify(69918)
	o.apply(s)
}

var DisableMutations = simpleOption("disable mutations", func(s *Smither) {
	__antithesis_instrumentation__.Notify(69919)
	s.stmtWeights = nonMutatingStatements
	s.tableExprWeights = nonMutatingTableExprs
})

var DisableDDLs = simpleOption("disable DDLs", func(s *Smither) {
	__antithesis_instrumentation__.Notify(69920)
	s.stmtWeights = []statementWeight{
		{20, makeSelect},
		{5, makeInsert},
		{5, makeUpdate},
		{1, makeDelete},

		{2, makeBegin},
		{2, makeSavepoint},
		{2, makeReleaseSavepoint},
		{2, makeRollbackToSavepoint},
		{2, makeCommit},
		{2, makeRollback},
	}
})

var OnlyNoDropDDLs = simpleOption("only DDLs", func(s *Smither) {
	__antithesis_instrumentation__.Notify(69921)
	s.stmtWeights = append(append([]statementWeight{
		{1, makeBegin},
		{2, makeRollback},
		{6, makeCommit},
	},
		altersExistingTable...,
	),
		altersExistingTypes...,
	)
})

var MultiRegionDDLs = simpleOption("include multiregion DDLs", func(s *Smither) {
	__antithesis_instrumentation__.Notify(69922)
	s.alterWeights = append(s.alterWeights, alterMultiregion...)
})

var DisableWith = simpleOption("disable WITH", func(s *Smither) {
	__antithesis_instrumentation__.Notify(69923)
	s.disableWith = true
})

var DisableImpureFns = simpleOption("disable impure funcs", func(s *Smither) {
	__antithesis_instrumentation__.Notify(69924)
	s.disableImpureFns = true
})

func DisableCRDBFns() SmitherOption {
	__antithesis_instrumentation__.Notify(69925)
	return IgnoreFNs("^crdb_internal")
}

var SimpleDatums = simpleOption("simple datums", func(s *Smither) {
	__antithesis_instrumentation__.Notify(69926)
	s.simpleDatums = true
})

var MutationsOnly = simpleOption("mutations only", func(s *Smither) {
	__antithesis_instrumentation__.Notify(69927)
	s.stmtWeights = []statementWeight{
		{8, makeInsert},
		{1, makeUpdate},
		{1, makeDelete},
	}
})

func IgnoreFNs(regex string) SmitherOption {
	__antithesis_instrumentation__.Notify(69928)
	r := regexp.MustCompile(regex)
	return option{
		name: fmt.Sprintf("ignore fns: %q", r.String()),
		apply: func(s *Smither) {
			__antithesis_instrumentation__.Notify(69929)
			s.ignoreFNs = append(s.ignoreFNs, r)
		},
	}
}

var DisableLimits = simpleOption("disable LIMIT", func(s *Smither) {
	__antithesis_instrumentation__.Notify(69930)
	s.disableLimits = true
})

var AvoidConsts = simpleOption("avoid consts", func(s *Smither) {
	__antithesis_instrumentation__.Notify(69931)
	s.avoidConsts = true
})

var DisableWindowFuncs = simpleOption("disable window funcs", func(s *Smither) {
	__antithesis_instrumentation__.Notify(69932)
	s.disableWindowFuncs = true
})

var OutputSort = simpleOption("output sort", func(s *Smither) {
	__antithesis_instrumentation__.Notify(69933)
	s.outputSort = true
})

var CompareMode = multiOption(
	"compare mode",
	DisableMutations(),
	DisableImpureFns(),
	DisableCRDBFns(),
	IgnoreFNs("^version"),
	DisableLimits(),
	OutputSort(),
)

var PostgresMode = multiOption(
	"postgres mode",
	CompareMode(),
	DisableWith(),
	SimpleDatums(),
	IgnoreFNs("^current_"),
	simpleOption("postgres", func(s *Smither) {
		__antithesis_instrumentation__.Notify(69934)
		s.postgres = true
	})(),

	IgnoreFNs("^sha"),
	IgnoreFNs("^isnan"),
	IgnoreFNs("^crc32c"),
	IgnoreFNs("^fnv32a"),
	IgnoreFNs("^experimental_"),
	IgnoreFNs("^json_set"),
	IgnoreFNs("^concat_agg"),
	IgnoreFNs("^to_english"),
	IgnoreFNs("^substr$"),

	IgnoreFNs("^quote"),

	IgnoreFNs("^pg_collation_for"),

	IgnoreFNs("_escape$"),

	IgnoreFNs("st_.*withinexclusive$"),
	IgnoreFNs("^postgis_.*build_date"),
	IgnoreFNs("^postgis_.*version"),
	IgnoreFNs("^postgis_.*scripts"),
)
