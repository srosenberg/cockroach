package logictest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	gosql "database/sql"
	"flag"
	"fmt"
	gobuild "go/build"
	"io"
	"math"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"runtime/debug"
	"runtime/trace"
	"sort"
	"strconv"
	"strings"
	"testing"
	"text/tabwriter"
	"time"
	"unicode/utf8"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/build"
	"github.com/cockroachdb/cockroach/pkg/build/bazel"
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver"
	"github.com/cockroachdb/cockroach/pkg/migration"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/randgen"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats"
	"github.com/cockroachdb/cockroach/pkg/sql/stats"
	"github.com/cockroachdb/cockroach/pkg/testutils"
	"github.com/cockroachdb/cockroach/pkg/testutils/floatcmp"
	"github.com/cockroachdb/cockroach/pkg/testutils/physicalplanutils"
	"github.com/cockroachdb/cockroach/pkg/testutils/serverutils"
	"github.com/cockroachdb/cockroach/pkg/testutils/skip"
	"github.com/cockroachdb/cockroach/pkg/testutils/sqlutils"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/randutil"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/oserror"
	"github.com/kr/pretty"
	"github.com/lib/pq"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/stretchr/testify/require"
)

var (
	resultsRE = regexp.MustCompile(`^(\d+)\s+values?\s+hashing\s+to\s+([0-9A-Fa-f]+)$`)
	noticeRE  = regexp.MustCompile(`^statement\s+notice\s+(.*)$`)
	errorRE   = regexp.MustCompile(`^(?:statement|query)\s+error\s+(?:pgcode\s+([[:alnum:]]+)\s+)?(.*)$`)
	varRE     = regexp.MustCompile(`\$[a-zA-Z][a-zA-Z_0-9]*`)

	logictestdata  = flag.String("d", "", "glob that selects subset of files to run")
	bigtest        = flag.Bool("bigtest", false, "enable the long-running SqlLiteLogic test")
	overrideConfig = flag.String(
		"config", "",
		"sets the test cluster configuration; comma-separated values",
	)

	maxErrs = flag.Int(
		"max-errors", 1,
		"stop processing input files after this number of errors (set to 0 for no limit)",
	)
	allowPrepareFail = flag.Bool(
		"allow-prepare-fail", false, "tolerate unexpected errors when preparing a query",
	)
	flexTypes = flag.Bool(
		"flex-types", false,
		"do not fail when a test expects a column of a numeric type but the query provides another type",
	)

	showSQL = flag.Bool("show-sql", false,
		"print the individual SQL statement/queries before processing",
	)

	showDiff          = flag.Bool("show-diff", false, "generate a diff for expectation mismatches when possible")
	printErrorSummary = flag.Bool("error-summary", false,
		"print a per-error summary of failing queries at the end of testing, "+
			"when -allow-prepare-fail is set",
	)
	fullMessages = flag.Bool("full-messages", false,
		"do not shorten the error or SQL strings when printing the summary for -allow-prepare-fail "+
			"or -flex-types.",
	)
	rewriteResultsInTestfiles = flag.Bool(
		"rewrite", false,
		"ignore the expected results and rewrite the test files with the actual results from this "+
			"run. Used to update tests when a change affects many cases; please verify the testfile "+
			"diffs carefully!",
	)
	rewriteSQL = flag.Bool(
		"rewrite-sql", false,
		"pretty-reformat the SQL queries. Only use this incidentally when importing new tests. "+
			"beware! some tests INTEND to use non-formatted SQL queries (e.g. case sensitivity). "+
			"do not bluntly apply!",
	)
	sqlfmtLen = flag.Int("line-length", tree.DefaultPrettyCfg().LineWidth,
		"target line length when using -rewrite-sql")
	disableOptRuleProbability = flag.Float64(
		"disable-opt-rule-probability", 0,
		"disable transformation rules in the cost-based optimizer with the given probability.")
	optimizerCostPerturbation = flag.Float64(
		"optimizer-cost-perturbation", 0,
		"randomly perturb the estimated cost of each expression in the query tree by at most the "+
			"given fraction for the purpose of creating alternate query plans in the optimizer.")
	printBlocklistIssues = flag.Bool(
		"print-blocklist-issues", false,
		"for any test files that contain a blocklist directive, print a link to the associated issue",
	)
)

type testClusterConfig struct {
	name     string
	numNodes int

	useFakeSpanResolver bool

	overrideDistSQLMode string

	overrideVectorize string

	overrideAutoStats string

	overrideExperimentalDistSQLPlanning string

	sqlExecUseDisk bool

	distSQLMetadataTestEnabled bool

	skipShort bool

	bootstrapVersion roachpb.Version

	binaryVersion  roachpb.Version
	disableUpgrade bool

	useTenant bool

	isCCLConfig bool

	localities map[int]roachpb.Locality

	backupRestoreProbability float64
}

const queryRewritePlaceholderPrefix = "__async_query_rewrite_placeholder"

const threeNodeTenantConfigName = "3node-tenant"

var multiregion9node3region3azsLocalities = map[int]roachpb.Locality{
	1: {
		Tiers: []roachpb.Tier{
			{Key: "region", Value: "ap-southeast-2"},
			{Key: "availability-zone", Value: "ap-az1"},
		},
	},
	2: {
		Tiers: []roachpb.Tier{
			{Key: "region", Value: "ap-southeast-2"},
			{Key: "availability-zone", Value: "ap-az2"},
		},
	},
	3: {
		Tiers: []roachpb.Tier{
			{Key: "region", Value: "ap-southeast-2"},
			{Key: "availability-zone", Value: "ap-az3"},
		},
	},
	4: {
		Tiers: []roachpb.Tier{
			{Key: "region", Value: "ca-central-1"},
			{Key: "availability-zone", Value: "ca-az1"},
		},
	},
	5: {
		Tiers: []roachpb.Tier{
			{Key: "region", Value: "ca-central-1"},
			{Key: "availability-zone", Value: "ca-az2"},
		},
	},
	6: {
		Tiers: []roachpb.Tier{
			{Key: "region", Value: "ca-central-1"},
			{Key: "availability-zone", Value: "ca-az3"},
		},
	},
	7: {
		Tiers: []roachpb.Tier{
			{Key: "region", Value: "us-east-1"},
			{Key: "availability-zone", Value: "us-az1"},
		},
	},
	8: {
		Tiers: []roachpb.Tier{
			{Key: "region", Value: "us-east-1"},
			{Key: "availability-zone", Value: "us-az2"},
		},
	},
	9: {
		Tiers: []roachpb.Tier{
			{Key: "region", Value: "us-east-1"},
			{Key: "availability-zone", Value: "us-az3"},
		},
	},
}

var multiregion15node5region3azsLocalities = map[int]roachpb.Locality{
	1: {
		Tiers: []roachpb.Tier{
			{Key: "region", Value: "ap-southeast-2"},
			{Key: "availability-zone", Value: "ap-az1"},
		},
	},
	2: {
		Tiers: []roachpb.Tier{
			{Key: "region", Value: "ap-southeast-2"},
			{Key: "availability-zone", Value: "ap-az2"},
		},
	},
	3: {
		Tiers: []roachpb.Tier{
			{Key: "region", Value: "ap-southeast-2"},
			{Key: "availability-zone", Value: "ap-az3"},
		},
	},
	4: {
		Tiers: []roachpb.Tier{
			{Key: "region", Value: "ca-central-1"},
			{Key: "availability-zone", Value: "ca-az1"},
		},
	},
	5: {
		Tiers: []roachpb.Tier{
			{Key: "region", Value: "ca-central-1"},
			{Key: "availability-zone", Value: "ca-az2"},
		},
	},
	6: {
		Tiers: []roachpb.Tier{
			{Key: "region", Value: "ca-central-1"},
			{Key: "availability-zone", Value: "ca-az3"},
		},
	},
	7: {
		Tiers: []roachpb.Tier{
			{Key: "region", Value: "us-east-1"},
			{Key: "availability-zone", Value: "us-az1"},
		},
	},
	8: {
		Tiers: []roachpb.Tier{
			{Key: "region", Value: "us-east-1"},
			{Key: "availability-zone", Value: "us-az2"},
		},
	},
	9: {
		Tiers: []roachpb.Tier{
			{Key: "region", Value: "us-east-1"},
			{Key: "availability-zone", Value: "us-az3"},
		},
	},
	10: {
		Tiers: []roachpb.Tier{
			{Key: "region", Value: "us-west-1"},
			{Key: "availability-zone", Value: "usw-az1"},
		},
	},
	11: {
		Tiers: []roachpb.Tier{
			{Key: "region", Value: "us-west-1"},
			{Key: "availability-zone", Value: "usw-az2"},
		},
	},
	12: {
		Tiers: []roachpb.Tier{
			{Key: "region", Value: "us-west-1"},
			{Key: "availability-zone", Value: "usw-az3"},
		},
	},
	13: {
		Tiers: []roachpb.Tier{
			{Key: "region", Value: "us-central-1"},
			{Key: "availability-zone", Value: "usc-az1"},
		},
	},
	14: {
		Tiers: []roachpb.Tier{
			{Key: "region", Value: "us-central-1"},
			{Key: "availability-zone", Value: "usc-az2"},
		},
	},
	15: {
		Tiers: []roachpb.Tier{
			{Key: "region", Value: "us-central-1"},
			{Key: "availability-zone", Value: "usc-az3"},
		},
	},
}

var logicTestConfigs = []testClusterConfig{
	{
		name:                "local",
		numNodes:            1,
		overrideDistSQLMode: "off",
		overrideAutoStats:   "false",
	},
	{
		name:                "local-vec-off",
		numNodes:            1,
		overrideDistSQLMode: "off",
		overrideAutoStats:   "false",
		overrideVectorize:   "off",
	},
	{
		name:                "local-v1.1@v1.0-noupgrade",
		numNodes:            1,
		overrideDistSQLMode: "off",
		overrideAutoStats:   "false",
		bootstrapVersion:    roachpb.Version{Major: 1},
		binaryVersion:       roachpb.Version{Major: 1, Minor: 1},
		disableUpgrade:      true,
	},
	{
		name:                                "local-spec-planning",
		numNodes:                            1,
		overrideDistSQLMode:                 "off",
		overrideAutoStats:                   "false",
		overrideExperimentalDistSQLPlanning: "on",
	},
	{
		name:                "fakedist",
		numNodes:            3,
		useFakeSpanResolver: true,
		overrideDistSQLMode: "on",
		overrideAutoStats:   "false",
	},
	{
		name:                "fakedist-vec-off",
		numNodes:            3,
		useFakeSpanResolver: true,
		overrideDistSQLMode: "on",
		overrideAutoStats:   "false",
		overrideVectorize:   "off",
	},
	{
		name:                       "fakedist-metadata",
		numNodes:                   3,
		useFakeSpanResolver:        true,
		overrideDistSQLMode:        "on",
		overrideAutoStats:          "false",
		distSQLMetadataTestEnabled: true,
		skipShort:                  true,
	},
	{
		name:                "fakedist-disk",
		numNodes:            3,
		useFakeSpanResolver: true,
		overrideDistSQLMode: "on",
		overrideAutoStats:   "false",
		sqlExecUseDisk:      true,
		skipShort:           true,
	},
	{
		name:                                "fakedist-spec-planning",
		numNodes:                            3,
		useFakeSpanResolver:                 true,
		overrideDistSQLMode:                 "on",
		overrideAutoStats:                   "false",
		overrideExperimentalDistSQLPlanning: "on",
	},
	{
		name:                "5node",
		numNodes:            5,
		overrideDistSQLMode: "on",
		overrideAutoStats:   "false",
	},
	{
		name:                       "5node-metadata",
		numNodes:                   5,
		overrideDistSQLMode:        "on",
		overrideAutoStats:          "false",
		distSQLMetadataTestEnabled: true,
		skipShort:                  true,
	},
	{
		name:                "5node-disk",
		numNodes:            5,
		overrideDistSQLMode: "on",
		overrideAutoStats:   "false",
		sqlExecUseDisk:      true,
		skipShort:           true,
	},
	{
		name:                                "5node-spec-planning",
		numNodes:                            5,
		overrideDistSQLMode:                 "on",
		overrideAutoStats:                   "false",
		overrideExperimentalDistSQLPlanning: "on",
	},
	{

		name:     threeNodeTenantConfigName,
		numNodes: 3,

		overrideAutoStats: "false",
		useTenant:         true,
		isCCLConfig:       true,
	},

	{
		name:              "multiregion-invalid-locality",
		numNodes:          3,
		overrideAutoStats: "false",
		localities: map[int]roachpb.Locality{
			1: {
				Tiers: []roachpb.Tier{
					{Key: "invalid-region-setup", Value: "test1"},
					{Key: "availability-zone", Value: "test1-az1"},
				},
			},
			2: {
				Tiers: []roachpb.Tier{},
			},
			3: {
				Tiers: []roachpb.Tier{
					{Key: "region", Value: "test1"},
					{Key: "availability-zone", Value: "test1-az3"},
				},
			},
		},
	},
	{
		name:              "multiregion-3node-3superlongregions",
		numNodes:          3,
		overrideAutoStats: "false",
		localities: map[int]roachpb.Locality{
			1: {
				Tiers: []roachpb.Tier{
					{Key: "region", Value: "veryveryveryveryveryveryverylongregion1"},
				},
			},
			2: {
				Tiers: []roachpb.Tier{
					{Key: "region", Value: "veryveryveryveryveryveryverylongregion2"},
				},
			},
			3: {
				Tiers: []roachpb.Tier{
					{Key: "region", Value: "veryveryveryveryveryveryverylongregion3"},
				},
			},
		},
	},
	{
		name:              "multiregion-9node-3region-3azs",
		numNodes:          9,
		overrideAutoStats: "false",
		localities:        multiregion9node3region3azsLocalities,
	},
	{
		name:              "multiregion-9node-3region-3azs-tenant",
		numNodes:          9,
		overrideAutoStats: "false",
		localities:        multiregion9node3region3azsLocalities,
		useTenant:         true,
	},
	{
		name:              "multiregion-9node-3region-3azs-vec-off",
		numNodes:          9,
		overrideAutoStats: "false",
		localities:        multiregion9node3region3azsLocalities,
		overrideVectorize: "off",
	},
	{
		name:              "multiregion-15node-5region-3azs",
		numNodes:          15,
		overrideAutoStats: "false",
		localities:        multiregion15node5region3azsLocalities,
	},
	{
		name:                "local-mixed-21.2-22.1",
		numNodes:            1,
		overrideDistSQLMode: "off",
		overrideAutoStats:   "false",
		bootstrapVersion:    roachpb.Version{Major: 21, Minor: 2},
		binaryVersion:       roachpb.Version{Major: 22, Minor: 1},
		disableUpgrade:      true,
	},
	{

		name:                     "3node-backup",
		numNodes:                 3,
		backupRestoreProbability: envutil.EnvOrDefaultFloat64("COCKROACH_LOGIC_TEST_BACKUP_RESTORE_PROBABILITY", 0.0),
		isCCLConfig:              true,
	},
}

type logicTestConfigIdx int

type configSet []logicTestConfigIdx

var logicTestConfigIdxToName = make(map[logicTestConfigIdx]string)

func init() {
	for i, cfg := range logicTestConfigs {
		logicTestConfigIdxToName[logicTestConfigIdx(i)] = cfg.name
	}
}

func parseTestConfig(names []string) configSet {
	__antithesis_instrumentation__.Notify(500237)
	ret := make(configSet, len(names))
	for i, name := range names {
		__antithesis_instrumentation__.Notify(500239)
		idx, ok := findLogicTestConfig(name)
		if !ok {
			__antithesis_instrumentation__.Notify(500241)
			panic(fmt.Errorf("unknown config %s", name))
		} else {
			__antithesis_instrumentation__.Notify(500242)
		}
		__antithesis_instrumentation__.Notify(500240)
		ret[i] = idx
	}
	__antithesis_instrumentation__.Notify(500238)
	return ret
}

var (
	defaultConfigName  = "default-configs"
	defaultConfigNames = []string{
		"local",
		"local-vec-off",
		"local-spec-planning",
		"fakedist",
		"fakedist-vec-off",
		"fakedist-metadata",
		"fakedist-disk",
		"fakedist-spec-planning",
	}

	fiveNodeDefaultConfigName  = "5node-default-configs"
	fiveNodeDefaultConfigNames = []string{
		"5node",
		"5node-metadata",
		"5node-disk",
		"5node-spec-planning",
	}
	defaultConfig         = parseTestConfig(defaultConfigNames)
	fiveNodeDefaultConfig = parseTestConfig(fiveNodeDefaultConfigNames)
)

func findLogicTestConfig(name string) (logicTestConfigIdx, bool) {
	__antithesis_instrumentation__.Notify(500243)
	for i, cfg := range logicTestConfigs {
		__antithesis_instrumentation__.Notify(500245)
		if cfg.name == name {
			__antithesis_instrumentation__.Notify(500246)
			return logicTestConfigIdx(i), true
		} else {
			__antithesis_instrumentation__.Notify(500247)
		}
	}
	__antithesis_instrumentation__.Notify(500244)
	return -1, false
}

type lineScanner struct {
	*bufio.Scanner
	line       int
	skip       bool
	skipReason string
}

func (l *lineScanner) SetSkip(reason string) {
	__antithesis_instrumentation__.Notify(500248)
	l.skip = true
	l.skipReason = reason
}

func (l *lineScanner) LogAndResetSkip(t *logicTest) {
	__antithesis_instrumentation__.Notify(500249)
	if l.skipReason != "" {
		__antithesis_instrumentation__.Notify(500251)
		t.t().Logf("statement/query skipped with reason: %s", l.skipReason)
	} else {
		__antithesis_instrumentation__.Notify(500252)
	}
	__antithesis_instrumentation__.Notify(500250)
	l.skipReason = ""
	l.skip = false
}

func newLineScanner(r io.Reader) *lineScanner {
	__antithesis_instrumentation__.Notify(500253)
	return &lineScanner{
		Scanner: bufio.NewScanner(r),
		line:    0,
	}
}

func (l *lineScanner) Scan() bool {
	__antithesis_instrumentation__.Notify(500254)
	ok := l.Scanner.Scan()
	if ok {
		__antithesis_instrumentation__.Notify(500256)
		l.line++
	} else {
		__antithesis_instrumentation__.Notify(500257)
	}
	__antithesis_instrumentation__.Notify(500255)
	return ok
}

func (l *lineScanner) Text() string {
	__antithesis_instrumentation__.Notify(500258)
	return l.Scanner.Text()
}

type logicStatement struct {
	pos string

	sql string

	expectNotice string

	expectErr string

	expectErrCode string

	expectCount int64

	expectAsync bool

	statementName string
}

type pendingExecResult struct {
	execSQL string
	res     gosql.Result
	err     error
}

type pendingStatement struct {
	logicStatement

	resultChan chan pendingExecResult
}

type pendingQueryResult struct {
	rows *gosql.Rows
	err  error
}

type pendingQuery struct {
	logicQuery

	resultChan chan pendingQueryResult
}

func (ls *logicStatement) readSQL(
	t *logicTest, s *lineScanner, allowSeparator bool,
) (separator bool, _ error) {
	__antithesis_instrumentation__.Notify(500259)
	var buf bytes.Buffer
	hasVars := false
	for s.Scan() {
		__antithesis_instrumentation__.Notify(500262)
		line := s.Text()
		if !*rewriteSQL {
			__antithesis_instrumentation__.Notify(500267)
			t.emit(line)
		} else {
			__antithesis_instrumentation__.Notify(500268)
		}
		__antithesis_instrumentation__.Notify(500263)
		substLine := t.substituteVars(line)
		if line != substLine {
			__antithesis_instrumentation__.Notify(500269)
			hasVars = true
			line = substLine
		} else {
			__antithesis_instrumentation__.Notify(500270)
		}
		__antithesis_instrumentation__.Notify(500264)
		if line == "" {
			__antithesis_instrumentation__.Notify(500271)
			break
		} else {
			__antithesis_instrumentation__.Notify(500272)
		}
		__antithesis_instrumentation__.Notify(500265)
		if line == "----" {
			__antithesis_instrumentation__.Notify(500273)
			separator = true
			if ls.expectNotice != "" {
				__antithesis_instrumentation__.Notify(500277)
				return false, errors.Errorf(
					"%s: invalid ---- separator after a statement expecting a notice: %s",
					ls.pos, ls.expectNotice,
				)
			} else {
				__antithesis_instrumentation__.Notify(500278)
			}
			__antithesis_instrumentation__.Notify(500274)
			if ls.expectErr != "" {
				__antithesis_instrumentation__.Notify(500279)
				return false, errors.Errorf(
					"%s: invalid ---- separator after a statement or query expecting an error: %s",
					ls.pos, ls.expectErr,
				)
			} else {
				__antithesis_instrumentation__.Notify(500280)
			}
			__antithesis_instrumentation__.Notify(500275)
			if !allowSeparator {
				__antithesis_instrumentation__.Notify(500281)
				return false, errors.Errorf("%s: unexpected ---- separator", ls.pos)
			} else {
				__antithesis_instrumentation__.Notify(500282)
			}
			__antithesis_instrumentation__.Notify(500276)
			break
		} else {
			__antithesis_instrumentation__.Notify(500283)
		}
		__antithesis_instrumentation__.Notify(500266)
		fmt.Fprintln(&buf, line)
	}
	__antithesis_instrumentation__.Notify(500260)
	ls.sql = strings.TrimSpace(buf.String())
	if *rewriteSQL {
		__antithesis_instrumentation__.Notify(500284)
		if !hasVars {
			__antithesis_instrumentation__.Notify(500286)
			newSyntax, err := func(inSql string) (string, error) {
				__antithesis_instrumentation__.Notify(500289)

				stmtList, err := parser.Parse(inSql)
				if err != nil {
					__antithesis_instrumentation__.Notify(500292)
					if ls.expectErr != "" {
						__antithesis_instrumentation__.Notify(500294)

						return inSql, nil
					} else {
						__antithesis_instrumentation__.Notify(500295)
					}
					__antithesis_instrumentation__.Notify(500293)
					return "", errors.Wrapf(err, "%s: error while parsing SQL for reformat:\n%s", ls.pos, ls.sql)
				} else {
					__antithesis_instrumentation__.Notify(500296)
				}
				__antithesis_instrumentation__.Notify(500290)
				var newSyntax bytes.Buffer
				pcfg := tree.DefaultPrettyCfg()
				pcfg.LineWidth = *sqlfmtLen
				pcfg.Simplify = false
				pcfg.UseTabs = false
				for i := range stmtList {
					__antithesis_instrumentation__.Notify(500297)
					if i > 0 {
						__antithesis_instrumentation__.Notify(500299)
						fmt.Fprintln(&newSyntax, ";")
					} else {
						__antithesis_instrumentation__.Notify(500300)
					}
					__antithesis_instrumentation__.Notify(500298)
					fmt.Fprint(&newSyntax, pcfg.Pretty(stmtList[i].AST))
				}
				__antithesis_instrumentation__.Notify(500291)
				return newSyntax.String(), nil
			}(ls.sql)
			__antithesis_instrumentation__.Notify(500287)
			if err != nil {
				__antithesis_instrumentation__.Notify(500301)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(500302)
			}
			__antithesis_instrumentation__.Notify(500288)
			ls.sql = newSyntax
		} else {
			__antithesis_instrumentation__.Notify(500303)
		}
		__antithesis_instrumentation__.Notify(500285)

		t.emit(ls.sql)
		if separator {
			__antithesis_instrumentation__.Notify(500304)
			t.emit("----")
		} else {
			__antithesis_instrumentation__.Notify(500305)
			t.emit("")
		}
	} else {
		__antithesis_instrumentation__.Notify(500306)
	}
	__antithesis_instrumentation__.Notify(500261)
	return separator, nil
}

type logicSorter func(numCols int, values []string)

type rowSorter struct {
	numCols int
	numRows int
	values  []string
}

func (r rowSorter) row(i int) []string {
	__antithesis_instrumentation__.Notify(500307)
	return r.values[i*r.numCols : (i+1)*r.numCols]
}

func (r rowSorter) Len() int {
	__antithesis_instrumentation__.Notify(500308)
	return r.numRows
}

func (r rowSorter) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(500309)
	a := r.row(i)
	b := r.row(j)
	for k := range a {
		__antithesis_instrumentation__.Notify(500311)
		if a[k] < b[k] {
			__antithesis_instrumentation__.Notify(500313)
			return true
		} else {
			__antithesis_instrumentation__.Notify(500314)
		}
		__antithesis_instrumentation__.Notify(500312)
		if a[k] > b[k] {
			__antithesis_instrumentation__.Notify(500315)
			return false
		} else {
			__antithesis_instrumentation__.Notify(500316)
		}
	}
	__antithesis_instrumentation__.Notify(500310)
	return false
}

func (r rowSorter) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(500317)
	a := r.row(i)
	b := r.row(j)
	for i := range a {
		__antithesis_instrumentation__.Notify(500318)
		a[i], b[i] = b[i], a[i]
	}
}

func rowSort(numCols int, values []string) {
	__antithesis_instrumentation__.Notify(500319)
	sort.Sort(rowSorter{
		numCols: numCols,
		numRows: len(values) / numCols,
		values:  values,
	})
}

func valueSort(numCols int, values []string) {
	__antithesis_instrumentation__.Notify(500320)
	sort.Strings(values)
}

func partialSort(numCols int, orderedCols []int, values []string) {
	__antithesis_instrumentation__.Notify(500321)

	c := rowSorter{
		numCols: numCols,
		numRows: len(values) / numCols,
		values:  values,
	}

	sortGroup := func(rowStart, rowEnd int) {
		__antithesis_instrumentation__.Notify(500324)
		sort.Sort(rowSorter{
			numCols: numCols,
			numRows: rowEnd - rowStart,
			values:  values[rowStart*numCols : rowEnd*numCols],
		})
	}
	__antithesis_instrumentation__.Notify(500322)

	groupStart := 0
	for rIdx := 1; rIdx < c.numRows; rIdx++ {
		__antithesis_instrumentation__.Notify(500325)

		row := c.row(rIdx)
		start := c.row(groupStart)
		differs := false
		for _, i := range orderedCols {
			__antithesis_instrumentation__.Notify(500327)
			if start[i] != row[i] {
				__antithesis_instrumentation__.Notify(500328)
				differs = true
				break
			} else {
				__antithesis_instrumentation__.Notify(500329)
			}
		}
		__antithesis_instrumentation__.Notify(500326)
		if differs {
			__antithesis_instrumentation__.Notify(500330)

			sortGroup(groupStart, rIdx)
			groupStart = rIdx
		} else {
			__antithesis_instrumentation__.Notify(500331)
		}
	}
	__antithesis_instrumentation__.Notify(500323)
	sortGroup(groupStart, c.numRows)
}

type logicQuery struct {
	logicStatement

	colTypes string

	colNames bool

	retry bool

	sorter logicSorter

	label string

	checkResults bool

	valsPerLine int

	expectedResults []string

	expectedResultsRaw []string

	expectedHash string

	expectedValues int

	kvtrace bool

	kvOpTypes        []string
	keyPrefixFilters []string

	nodeIdx int

	noticetrace bool

	rawOpts string

	roundFloatsInStrings bool
}

var allowedKVOpTypes = []string{
	"CPut",
	"Put",
	"InitPut",
	"Del",
	"DelRange",
	"ClearRange",
	"Get",
	"Scan",
}

func isAllowedKVOp(op string) bool {
	__antithesis_instrumentation__.Notify(500332)
	for _, s := range allowedKVOpTypes {
		__antithesis_instrumentation__.Notify(500334)
		if op == s {
			__antithesis_instrumentation__.Notify(500335)
			return true
		} else {
			__antithesis_instrumentation__.Notify(500336)
		}
	}
	__antithesis_instrumentation__.Notify(500333)
	return false
}

type logicTest struct {
	rootT    *testing.T
	subtestT *testing.T
	rng      *rand.Rand
	cfg      testClusterConfig

	serverArgs *TestServerArgs

	clusterOpts []clusterOpt

	tenantClusterSettingOverrideOpts []tenantClusterSettingOverrideOpt

	cluster serverutils.TestClusterInterface

	sharedIODir string

	nodeIdx int

	tenantAddrs []string

	clients map[string]*gosql.DB

	user string
	db   *gosql.DB

	clusterCleanupFuncs []func()

	testCleanupFuncs []func()

	progress int

	failures int

	unsupported int

	lastProgress time.Time

	traceFile *os.File

	verbose bool

	perErrorSummary map[string][]string

	labelMap map[string]string

	varMap map[string]string

	pendingStatements map[string]pendingStatement

	pendingQueries map[string]pendingQuery

	noticeBuffer []string

	rewriteResTestBuf bytes.Buffer

	curPath   string
	curLineNo int

	skipOnRetry    bool
	skippedOnRetry bool
}

func (t *logicTest) t() *testing.T {
	__antithesis_instrumentation__.Notify(500337)
	if t.subtestT != nil {
		__antithesis_instrumentation__.Notify(500339)
		return t.subtestT
	} else {
		__antithesis_instrumentation__.Notify(500340)
	}
	__antithesis_instrumentation__.Notify(500338)
	return t.rootT
}

func (t *logicTest) traceStart(filename string) {
	__antithesis_instrumentation__.Notify(500341)
	if t.traceFile != nil {
		__antithesis_instrumentation__.Notify(500344)
		t.Fatalf("tracing already active")
	} else {
		__antithesis_instrumentation__.Notify(500345)
	}
	__antithesis_instrumentation__.Notify(500342)
	var err error
	t.traceFile, err = os.Create(filename)
	if err != nil {
		__antithesis_instrumentation__.Notify(500346)
		t.Fatalf("unable to open trace output file: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(500347)
	}
	__antithesis_instrumentation__.Notify(500343)
	if err := trace.Start(t.traceFile); err != nil {
		__antithesis_instrumentation__.Notify(500348)
		t.Fatalf("unable to start tracing: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(500349)
	}
}

func (t *logicTest) traceStop() {
	__antithesis_instrumentation__.Notify(500350)
	if t.traceFile != nil {
		__antithesis_instrumentation__.Notify(500351)
		trace.Stop()
		t.traceFile.Close()
		t.traceFile = nil
	} else {
		__antithesis_instrumentation__.Notify(500352)
	}
}

func (t *logicTest) substituteVars(line string) string {
	__antithesis_instrumentation__.Notify(500353)
	if len(t.varMap) == 0 {
		__antithesis_instrumentation__.Notify(500355)
		return line
	} else {
		__antithesis_instrumentation__.Notify(500356)
	}
	__antithesis_instrumentation__.Notify(500354)

	return varRE.ReplaceAllStringFunc(line, func(varName string) string {
		__antithesis_instrumentation__.Notify(500357)
		if replace, ok := t.varMap[varName]; ok {
			__antithesis_instrumentation__.Notify(500359)
			return replace
		} else {
			__antithesis_instrumentation__.Notify(500360)
		}
		__antithesis_instrumentation__.Notify(500358)
		return line
	})
}

func (t *logicTest) rewriteUpToRegex(matchRE *regexp.Regexp) *bufio.Scanner {
	__antithesis_instrumentation__.Notify(500361)
	remainder := bytes.NewReader(t.rewriteResTestBuf.Bytes())
	t.rewriteResTestBuf = bytes.Buffer{}
	scanner := bufio.NewScanner(remainder)
	for scanner.Scan() {
		__antithesis_instrumentation__.Notify(500363)
		if matchRE.Match(scanner.Bytes()) {
			__antithesis_instrumentation__.Notify(500365)
			break
		} else {
			__antithesis_instrumentation__.Notify(500366)
		}
		__antithesis_instrumentation__.Notify(500364)
		t.rewriteResTestBuf.Write(scanner.Bytes())
		t.rewriteResTestBuf.WriteString("\n")
	}
	__antithesis_instrumentation__.Notify(500362)
	return scanner
}

func (t *logicTest) emit(line string) {
	__antithesis_instrumentation__.Notify(500367)
	if *rewriteResultsInTestfiles || func() bool {
		__antithesis_instrumentation__.Notify(500368)
		return *rewriteSQL == true
	}() == true {
		__antithesis_instrumentation__.Notify(500369)
		t.rewriteResTestBuf.WriteString(line)
		t.rewriteResTestBuf.WriteString("\n")
	} else {
		__antithesis_instrumentation__.Notify(500370)
	}
}

func (t *logicTest) close() {
	__antithesis_instrumentation__.Notify(500371)
	t.traceStop()

	t.shutdownCluster()

	for _, cleanup := range t.testCleanupFuncs {
		__antithesis_instrumentation__.Notify(500373)
		cleanup()
	}
	__antithesis_instrumentation__.Notify(500372)
	t.testCleanupFuncs = nil
}

func (t *logicTest) outf(format string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(500374)
	if !t.verbose {
		__antithesis_instrumentation__.Notify(500376)
		return
	} else {
		__antithesis_instrumentation__.Notify(500377)
	}
	__antithesis_instrumentation__.Notify(500375)
	log.Infof(context.Background(), format, args...)
	msg := fmt.Sprintf(format, args...)
	now := timeutil.Now().Format("15:04:05")
	fmt.Printf("[%s] %s\n", now, msg)
}

func (t *logicTest) setUser(user string, nodeIdxOverride int) func() {
	__antithesis_instrumentation__.Notify(500378)
	if t.clients == nil {
		__antithesis_instrumentation__.Notify(500385)
		t.clients = map[string]*gosql.DB{}
	} else {
		__antithesis_instrumentation__.Notify(500386)
	}
	__antithesis_instrumentation__.Notify(500379)
	if db, ok := t.clients[user]; ok {
		__antithesis_instrumentation__.Notify(500387)
		t.db = db
		t.user = user

		return func() { __antithesis_instrumentation__.Notify(500388) }
	} else {
		__antithesis_instrumentation__.Notify(500389)
	}
	__antithesis_instrumentation__.Notify(500380)

	nodeIdx := t.nodeIdx
	if nodeIdxOverride > 0 {
		__antithesis_instrumentation__.Notify(500390)
		nodeIdx = nodeIdxOverride
	} else {
		__antithesis_instrumentation__.Notify(500391)
	}
	__antithesis_instrumentation__.Notify(500381)

	addr := t.cluster.Server(nodeIdx).ServingSQLAddr()
	if len(t.tenantAddrs) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(500392)
		return !strings.HasPrefix(user, "host-cluster-") == true
	}() == true {
		__antithesis_instrumentation__.Notify(500393)
		addr = t.tenantAddrs[nodeIdx]
	} else {
		__antithesis_instrumentation__.Notify(500394)
	}
	__antithesis_instrumentation__.Notify(500382)
	pgUser := strings.TrimPrefix(user, "host-cluster-")
	pgURL, cleanupFunc := sqlutils.PGUrl(t.rootT, addr, "TestLogic", url.User(pgUser))
	pgURL.Path = "test"
	db := t.openDB(pgURL)

	if _, err := db.Exec("SET extra_float_digits = 0"); err != nil {
		__antithesis_instrumentation__.Notify(500395)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(500396)
	}
	__antithesis_instrumentation__.Notify(500383)

	if _, err := db.Exec("SET index_recommendations_enabled = false"); err != nil {
		__antithesis_instrumentation__.Notify(500397)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(500398)
	}
	__antithesis_instrumentation__.Notify(500384)
	t.clients[user] = db
	t.db = db
	t.user = pgUser

	return cleanupFunc
}

func (t *logicTest) openDB(pgURL url.URL) *gosql.DB {
	__antithesis_instrumentation__.Notify(500399)
	base, err := pq.NewConnector(pgURL.String())
	if err != nil {
		__antithesis_instrumentation__.Notify(500402)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(500403)
	}
	__antithesis_instrumentation__.Notify(500400)

	connector := pq.ConnectorWithNoticeHandler(base, func(notice *pq.Error) {
		__antithesis_instrumentation__.Notify(500404)
		t.noticeBuffer = append(t.noticeBuffer, notice.Severity+": "+notice.Message)
		if notice.Detail != "" {
			__antithesis_instrumentation__.Notify(500406)
			t.noticeBuffer = append(t.noticeBuffer, "DETAIL: "+notice.Detail)
		} else {
			__antithesis_instrumentation__.Notify(500407)
		}
		__antithesis_instrumentation__.Notify(500405)
		if notice.Hint != "" {
			__antithesis_instrumentation__.Notify(500408)
			t.noticeBuffer = append(t.noticeBuffer, "HINT: "+notice.Hint)
		} else {
			__antithesis_instrumentation__.Notify(500409)
		}
	})
	__antithesis_instrumentation__.Notify(500401)

	return gosql.OpenDB(connector)
}

func (t *logicTest) newCluster(
	serverArgs TestServerArgs,
	clusterOpts []clusterOpt,
	tenantClusterSettingOverrideOpts []tenantClusterSettingOverrideOpt,
) {
	__antithesis_instrumentation__.Notify(500410)

	if serverArgs.maxSQLMemoryLimit == 0 {
		__antithesis_instrumentation__.Notify(500425)

		serverArgs.maxSQLMemoryLimit = 192 * 1024 * 1024
	} else {
		__antithesis_instrumentation__.Notify(500426)
	}
	__antithesis_instrumentation__.Notify(500411)
	var tempStorageConfig base.TempStorageConfig
	if serverArgs.tempStorageDiskLimit == 0 {
		__antithesis_instrumentation__.Notify(500427)
		tempStorageConfig = base.DefaultTestTempStorageConfig(cluster.MakeTestingClusterSettings())
	} else {
		__antithesis_instrumentation__.Notify(500428)
		tempStorageConfig = base.DefaultTestTempStorageConfigWithSize(cluster.MakeTestingClusterSettings(), serverArgs.tempStorageDiskLimit)
	}
	__antithesis_instrumentation__.Notify(500412)

	params := base.TestClusterArgs{
		ServerArgs: base.TestServerArgs{
			SQLMemoryPoolSize: serverArgs.maxSQLMemoryLimit,
			TempStorageConfig: tempStorageConfig,
			Knobs: base.TestingKnobs{
				Store: &kvserver.StoreTestingKnobs{

					DisableConsistencyQueue: true,
				},
				SQLEvalContext: &tree.EvalContextTestingKnobs{
					AssertBinaryExprReturnTypes:     true,
					AssertUnaryExprReturnTypes:      true,
					AssertFuncExprReturnTypes:       true,
					DisableOptimizerRuleProbability: *disableOptRuleProbability,
					OptimizerCostPerturbation:       *optimizerCostPerturbation,
					ForceProductionBatchSizes:       serverArgs.forceProductionBatchSizes,
				},
				SQLExecutor: &sql.ExecutorTestingKnobs{
					DeterministicExplain: true,
				},
				SQLStatsKnobs: &sqlstats.TestingKnobs{
					AOSTClause: "AS OF SYSTEM TIME '-1us'",
				},
			},
			ClusterName:   "testclustername",
			ExternalIODir: t.sharedIODir,
		},

		ReplicationMode: base.ReplicationManual,
	}

	cfg := t.cfg
	distSQLKnobs := &execinfra.TestingKnobs{
		MetadataTestLevel: execinfra.Off,
	}
	if cfg.sqlExecUseDisk {
		__antithesis_instrumentation__.Notify(500429)
		distSQLKnobs.ForceDiskSpill = true
	} else {
		__antithesis_instrumentation__.Notify(500430)
	}
	__antithesis_instrumentation__.Notify(500413)
	if cfg.distSQLMetadataTestEnabled {
		__antithesis_instrumentation__.Notify(500431)
		distSQLKnobs.MetadataTestLevel = execinfra.On
	} else {
		__antithesis_instrumentation__.Notify(500432)
	}
	__antithesis_instrumentation__.Notify(500414)
	params.ServerArgs.Knobs.DistSQL = distSQLKnobs
	if cfg.bootstrapVersion != (roachpb.Version{}) {
		__antithesis_instrumentation__.Notify(500433)
		if params.ServerArgs.Knobs.Server == nil {
			__antithesis_instrumentation__.Notify(500435)
			params.ServerArgs.Knobs.Server = &server.TestingKnobs{}
		} else {
			__antithesis_instrumentation__.Notify(500436)
		}
		__antithesis_instrumentation__.Notify(500434)
		params.ServerArgs.Knobs.Server.(*server.TestingKnobs).BinaryVersionOverride = cfg.bootstrapVersion
	} else {
		__antithesis_instrumentation__.Notify(500437)
	}
	__antithesis_instrumentation__.Notify(500415)
	if cfg.disableUpgrade {
		__antithesis_instrumentation__.Notify(500438)
		if params.ServerArgs.Knobs.Server == nil {
			__antithesis_instrumentation__.Notify(500440)
			params.ServerArgs.Knobs.Server = &server.TestingKnobs{}
		} else {
			__antithesis_instrumentation__.Notify(500441)
		}
		__antithesis_instrumentation__.Notify(500439)
		params.ServerArgs.Knobs.Server.(*server.TestingKnobs).DisableAutomaticVersionUpgrade = make(chan struct{})
	} else {
		__antithesis_instrumentation__.Notify(500442)
	}
	__antithesis_instrumentation__.Notify(500416)
	for _, opt := range clusterOpts {
		__antithesis_instrumentation__.Notify(500443)
		opt.apply(&params.ServerArgs)
	}
	__antithesis_instrumentation__.Notify(500417)

	paramsPerNode := map[int]base.TestServerArgs{}
	require.Truef(
		t.rootT,
		len(cfg.localities) == 0 || func() bool {
			__antithesis_instrumentation__.Notify(500444)
			return len(cfg.localities) == cfg.numNodes == true
		}() == true,
		"localities must be set for each node -- got %#v for %d nodes",
		cfg.localities,
		cfg.numNodes,
	)
	for i := 0; i < cfg.numNodes; i++ {
		__antithesis_instrumentation__.Notify(500445)
		nodeParams := params.ServerArgs
		if locality, ok := cfg.localities[i+1]; ok {
			__antithesis_instrumentation__.Notify(500448)
			nodeParams.Locality = locality
		} else {
			__antithesis_instrumentation__.Notify(500449)
			require.Lenf(t.rootT, cfg.localities, 0, "node %d does not have a locality set", i+1)
		}
		__antithesis_instrumentation__.Notify(500446)

		if cfg.binaryVersion != (roachpb.Version{}) {
			__antithesis_instrumentation__.Notify(500450)
			binaryMinSupportedVersion := cfg.binaryVersion
			if cfg.bootstrapVersion != (roachpb.Version{}) {
				__antithesis_instrumentation__.Notify(500452)

				binaryMinSupportedVersion = cfg.bootstrapVersion
			} else {
				__antithesis_instrumentation__.Notify(500453)
			}
			__antithesis_instrumentation__.Notify(500451)
			nodeParams.Settings = cluster.MakeTestingClusterSettingsWithVersions(
				cfg.binaryVersion,
				binaryMinSupportedVersion,
				false,
			)

			from := clusterversion.ClusterVersion{Version: cfg.bootstrapVersion}
			to := clusterversion.ClusterVersion{Version: cfg.binaryVersion}
			if len(clusterversion.ListBetween(from, to)) == 0 {
				__antithesis_instrumentation__.Notify(500454)
				mm, ok := nodeParams.Knobs.MigrationManager.(*migration.TestingKnobs)
				if !ok {
					__antithesis_instrumentation__.Notify(500456)
					mm = &migration.TestingKnobs{}
					nodeParams.Knobs.MigrationManager = mm
				} else {
					__antithesis_instrumentation__.Notify(500457)
				}
				__antithesis_instrumentation__.Notify(500455)
				mm.ListBetweenOverride = func(
					from, to clusterversion.ClusterVersion,
				) []clusterversion.ClusterVersion {
					__antithesis_instrumentation__.Notify(500458)
					return []clusterversion.ClusterVersion{to}
				}
			} else {
				__antithesis_instrumentation__.Notify(500459)
			}
		} else {
			__antithesis_instrumentation__.Notify(500460)
		}
		__antithesis_instrumentation__.Notify(500447)
		paramsPerNode[i] = nodeParams
	}
	__antithesis_instrumentation__.Notify(500418)
	params.ServerArgsPerNode = paramsPerNode

	stats.DefaultAsOfTime = 10 * time.Millisecond
	stats.DefaultRefreshInterval = time.Millisecond

	t.cluster = serverutils.StartNewTestCluster(t.rootT, cfg.numNodes, params)
	if cfg.useFakeSpanResolver {
		__antithesis_instrumentation__.Notify(500461)
		fakeResolver := physicalplanutils.FakeResolverForTestCluster(t.cluster)
		t.cluster.Server(t.nodeIdx).SetDistSQLSpanResolver(fakeResolver)
	} else {
		__antithesis_instrumentation__.Notify(500462)
	}
	__antithesis_instrumentation__.Notify(500419)

	connsForClusterSettingChanges := []*gosql.DB{t.cluster.ServerConn(0)}
	clusterSettingOverrideArgs := &tenantClusterSettingOverrideArgs{}
	if cfg.useTenant {
		__antithesis_instrumentation__.Notify(500463)
		t.tenantAddrs = make([]string, cfg.numNodes)
		for i := 0; i < cfg.numNodes; i++ {
			__antithesis_instrumentation__.Notify(500470)
			tenantArgs := base.TestTenantArgs{
				TenantID:                    serverutils.TestTenantID(),
				AllowSettingClusterSettings: true,
				TestingKnobs: base.TestingKnobs{
					SQLExecutor: &sql.ExecutorTestingKnobs{
						DeterministicExplain: true,
					},
					SQLStatsKnobs: &sqlstats.TestingKnobs{
						AOSTClause: "AS OF SYSTEM TIME '-1us'",
					},
				},
				MemoryPoolSize:    params.ServerArgs.SQLMemoryPoolSize,
				TempStorageConfig: &params.ServerArgs.TempStorageConfig,
				Locality:          paramsPerNode[i].Locality,
				Existing:          i > 0,
				TracingDefault:    params.ServerArgs.TracingDefault,
			}

			tenant, err := t.cluster.Server(i).StartTenant(context.Background(), tenantArgs)
			if err != nil {
				__antithesis_instrumentation__.Notify(500472)
				t.rootT.Fatalf("%+v", err)
			} else {
				__antithesis_instrumentation__.Notify(500473)
			}
			__antithesis_instrumentation__.Notify(500471)
			t.tenantAddrs[i] = tenant.SQLAddr()
		}
		__antithesis_instrumentation__.Notify(500464)

		pgURL, cleanup := sqlutils.PGUrl(t.rootT, t.tenantAddrs[0], "Tenant", url.User(security.RootUser))
		defer cleanup()
		if params.ServerArgs.Insecure {
			__antithesis_instrumentation__.Notify(500474)
			pgURL.RawQuery = "sslmode=disable"
		} else {
			__antithesis_instrumentation__.Notify(500475)
		}
		__antithesis_instrumentation__.Notify(500465)
		db, err := gosql.Open("postgres", pgURL.String())
		if err != nil {
			__antithesis_instrumentation__.Notify(500476)
			t.rootT.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(500477)
		}
		__antithesis_instrumentation__.Notify(500466)
		defer db.Close()
		connsForClusterSettingChanges = append(connsForClusterSettingChanges, db)

		conn := t.cluster.ServerConn(0)
		if _, err := conn.Exec("SET CLUSTER SETTING kv.tenant_rate_limiter.rate_limit = 100000"); err != nil {
			__antithesis_instrumentation__.Notify(500478)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(500479)
		}
		__antithesis_instrumentation__.Notify(500467)

		for _, opt := range tenantClusterSettingOverrideOpts {
			__antithesis_instrumentation__.Notify(500480)
			opt.apply(clusterSettingOverrideArgs)
		}
		__antithesis_instrumentation__.Notify(500468)

		if clusterSettingOverrideArgs.overrideMultiTenantZoneConfigsAllowed {
			__antithesis_instrumentation__.Notify(500481)
			conn := t.cluster.ServerConn(0)

			if _, err := conn.Exec(
				"SET CLUSTER SETTING kv.closed_timestamp.target_duration = '50ms'",
			); err != nil {
				__antithesis_instrumentation__.Notify(500484)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(500485)
			}
			__antithesis_instrumentation__.Notify(500482)
			if _, err := conn.Exec(
				"SET CLUSTER SETTING kv.closed_timestamp.side_transport_interval = '50ms'",
			); err != nil {
				__antithesis_instrumentation__.Notify(500486)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(500487)
			}
			__antithesis_instrumentation__.Notify(500483)

			if _, err := conn.Exec(
				"ALTER TENANT $1 SET CLUSTER SETTING sql.zone_configs.allow_for_secondary_tenant.enabled = true",
				serverutils.TestTenantID().ToUint64(),
			); err != nil {
				__antithesis_instrumentation__.Notify(500488)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(500489)
			}
		} else {
			__antithesis_instrumentation__.Notify(500490)
		}
		__antithesis_instrumentation__.Notify(500469)

		if clusterSettingOverrideArgs.overrideMultiTenantMultiRegionAbstractionsAllowed {
			__antithesis_instrumentation__.Notify(500491)
			conn := t.cluster.ServerConn(0)

			if _, err := conn.Exec(
				fmt.Sprintf(
					"ALTER TENANT $1 SET CLUSTER SETTING %s = true",
					sql.SecondaryTenantsMultiRegionAbstractionsEnabledSettingName,
				),
				serverutils.TestTenantID().ToUint64(),
			); err != nil {
				__antithesis_instrumentation__.Notify(500492)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(500493)
			}
		} else {
			__antithesis_instrumentation__.Notify(500494)
		}
	} else {
		__antithesis_instrumentation__.Notify(500495)
	}
	__antithesis_instrumentation__.Notify(500420)

	for _, conn := range connsForClusterSettingChanges {
		__antithesis_instrumentation__.Notify(500496)
		if _, err := conn.Exec(
			"SET CLUSTER SETTING sql.stats.automatic_collection.min_stale_rows = $1::int", 5,
		); err != nil {
			__antithesis_instrumentation__.Notify(500502)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(500503)
		}
		__antithesis_instrumentation__.Notify(500497)

		if cfg.overrideDistSQLMode != "" {
			__antithesis_instrumentation__.Notify(500504)
			if _, err := conn.Exec(
				"SET CLUSTER SETTING sql.defaults.distsql = $1::string", cfg.overrideDistSQLMode,
			); err != nil {
				__antithesis_instrumentation__.Notify(500505)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(500506)
			}
		} else {
			__antithesis_instrumentation__.Notify(500507)
		}
		__antithesis_instrumentation__.Notify(500498)

		if cfg.overrideVectorize != "" {
			__antithesis_instrumentation__.Notify(500508)
			if _, err := conn.Exec(
				"SET CLUSTER SETTING sql.defaults.vectorize = $1::string", cfg.overrideVectorize,
			); err != nil {
				__antithesis_instrumentation__.Notify(500509)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(500510)
			}
		} else {
			__antithesis_instrumentation__.Notify(500511)
		}
		__antithesis_instrumentation__.Notify(500499)

		if cfg.overrideAutoStats != "" {
			__antithesis_instrumentation__.Notify(500512)
			if _, err := conn.Exec(
				"SET CLUSTER SETTING sql.stats.automatic_collection.enabled = $1::bool", cfg.overrideAutoStats,
			); err != nil {
				__antithesis_instrumentation__.Notify(500513)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(500514)
			}
		} else {
			__antithesis_instrumentation__.Notify(500515)

			if _, err := conn.Exec(
				"SET CLUSTER SETTING sql.stats.automatic_collection.enabled = false",
			); err != nil {
				__antithesis_instrumentation__.Notify(500516)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(500517)
			}
		}
		__antithesis_instrumentation__.Notify(500500)

		if cfg.overrideExperimentalDistSQLPlanning != "" {
			__antithesis_instrumentation__.Notify(500518)
			if _, err := conn.Exec(
				"SET CLUSTER SETTING sql.defaults.experimental_distsql_planning = $1::string", cfg.overrideExperimentalDistSQLPlanning,
			); err != nil {
				__antithesis_instrumentation__.Notify(500519)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(500520)
			}
		} else {
			__antithesis_instrumentation__.Notify(500521)
		}
		__antithesis_instrumentation__.Notify(500501)

		if _, err := conn.Exec(
			"SET CLUSTER SETTING sql.crdb_internal.table_row_statistics.as_of_time = '-1µs'",
		); err != nil {
			__antithesis_instrumentation__.Notify(500522)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(500523)
		}
	}
	__antithesis_instrumentation__.Notify(500421)

	if cfg.overrideDistSQLMode != "" {
		__antithesis_instrumentation__.Notify(500524)
		_, ok := sessiondatapb.DistSQLExecModeFromString(cfg.overrideDistSQLMode)
		if !ok {
			__antithesis_instrumentation__.Notify(500526)
			t.Fatalf("invalid distsql mode override: %s", cfg.overrideDistSQLMode)
		} else {
			__antithesis_instrumentation__.Notify(500527)
		}
		__antithesis_instrumentation__.Notify(500525)

		testutils.SucceedsSoon(t.rootT, func() error {
			__antithesis_instrumentation__.Notify(500528)
			for i := 0; i < t.cluster.NumServers(); i++ {
				__antithesis_instrumentation__.Notify(500530)
				var m string
				err := t.cluster.ServerConn(i % t.cluster.NumServers()).QueryRow(
					"SHOW CLUSTER SETTING sql.defaults.distsql",
				).Scan(&m)
				if err != nil {
					__antithesis_instrumentation__.Notify(500532)
					t.Fatal(errors.Wrapf(err, "%d", i))
				} else {
					__antithesis_instrumentation__.Notify(500533)
				}
				__antithesis_instrumentation__.Notify(500531)
				if m != cfg.overrideDistSQLMode {
					__antithesis_instrumentation__.Notify(500534)
					return errors.Errorf("node %d is still waiting for update of DistSQLMode to %s (have %s)",
						i, cfg.overrideDistSQLMode, m,
					)
				} else {
					__antithesis_instrumentation__.Notify(500535)
				}
			}
			__antithesis_instrumentation__.Notify(500529)
			return nil
		})
	} else {
		__antithesis_instrumentation__.Notify(500536)
	}
	__antithesis_instrumentation__.Notify(500422)

	if clusterSettingOverrideArgs.overrideMultiTenantZoneConfigsAllowed {
		__antithesis_instrumentation__.Notify(500537)
		t.waitForTenantReadOnlyClusterSettingToTakeEffectOrFatal(
			"sql.zone_configs.allow_for_secondary_tenant.enabled", "true", params.ServerArgs.Insecure,
		)
	} else {
		__antithesis_instrumentation__.Notify(500538)
	}
	__antithesis_instrumentation__.Notify(500423)
	if clusterSettingOverrideArgs.overrideMultiTenantMultiRegionAbstractionsAllowed {
		__antithesis_instrumentation__.Notify(500539)
		t.waitForTenantReadOnlyClusterSettingToTakeEffectOrFatal(
			sql.SecondaryTenantsMultiRegionAbstractionsEnabledSettingName,
			"true",
			params.ServerArgs.Insecure,
		)
	} else {
		__antithesis_instrumentation__.Notify(500540)
	}
	__antithesis_instrumentation__.Notify(500424)

	t.clusterCleanupFuncs = append(t.clusterCleanupFuncs, t.setUser(security.RootUser, 0))
}

func (t *logicTest) waitForTenantReadOnlyClusterSettingToTakeEffectOrFatal(
	settingName string, expValue string, insecure bool,
) {
	__antithesis_instrumentation__.Notify(500541)

	testutils.SucceedsSoon(t.rootT, func() error {
		__antithesis_instrumentation__.Notify(500542)
		for i := 0; i < len(t.tenantAddrs); i++ {
			__antithesis_instrumentation__.Notify(500544)
			pgURL, cleanup := sqlutils.PGUrl(t.rootT, t.tenantAddrs[0], "Tenant", url.User(security.RootUser))
			defer cleanup()
			if insecure {
				__antithesis_instrumentation__.Notify(500548)
				pgURL.RawQuery = "sslmode=disable"
			} else {
				__antithesis_instrumentation__.Notify(500549)
			}
			__antithesis_instrumentation__.Notify(500545)
			db, err := gosql.Open("postgres", pgURL.String())
			if err != nil {
				__antithesis_instrumentation__.Notify(500550)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(500551)
			}
			__antithesis_instrumentation__.Notify(500546)
			defer db.Close()

			var val string
			err = db.QueryRow(
				fmt.Sprintf("SHOW CLUSTER SETTING %s", settingName),
			).Scan(&val)
			if err != nil {
				__antithesis_instrumentation__.Notify(500552)
				t.Fatal(errors.Wrapf(err, "%d", i))
			} else {
				__antithesis_instrumentation__.Notify(500553)
			}
			__antithesis_instrumentation__.Notify(500547)
			if val != expValue {
				__antithesis_instrumentation__.Notify(500554)
				return errors.Errorf("tenant server %d is still waiting zone config cluster setting update",
					i,
				)
			} else {
				__antithesis_instrumentation__.Notify(500555)
			}
		}
		__antithesis_instrumentation__.Notify(500543)
		return nil
	})
}

func (t *logicTest) shutdownCluster() {
	__antithesis_instrumentation__.Notify(500556)
	for _, cleanup := range t.clusterCleanupFuncs {
		__antithesis_instrumentation__.Notify(500560)
		cleanup()
	}
	__antithesis_instrumentation__.Notify(500557)
	t.clusterCleanupFuncs = nil

	if t.cluster != nil {
		__antithesis_instrumentation__.Notify(500561)
		t.cluster.Stopper().Stop(context.TODO())
		t.cluster = nil
	} else {
		__antithesis_instrumentation__.Notify(500562)
	}
	__antithesis_instrumentation__.Notify(500558)
	if t.clients != nil {
		__antithesis_instrumentation__.Notify(500563)
		for _, c := range t.clients {
			__antithesis_instrumentation__.Notify(500565)
			c.Close()
		}
		__antithesis_instrumentation__.Notify(500564)
		t.clients = nil
	} else {
		__antithesis_instrumentation__.Notify(500566)
	}
	__antithesis_instrumentation__.Notify(500559)
	t.db = nil
}

func (t *logicTest) resetCluster() {
	__antithesis_instrumentation__.Notify(500567)
	t.shutdownCluster()
	if t.serverArgs == nil {
		__antithesis_instrumentation__.Notify(500569)

		t.Fatal("resetting the cluster before server args were set")
	} else {
		__antithesis_instrumentation__.Notify(500570)
	}
	__antithesis_instrumentation__.Notify(500568)
	serverArgs := *t.serverArgs
	t.newCluster(serverArgs, t.clusterOpts, t.tenantClusterSettingOverrideOpts)
}

func (t *logicTest) setup(
	cfg testClusterConfig,
	serverArgs TestServerArgs,
	clusterOpts []clusterOpt,
	tenantClusterSettingOverrideOpts []tenantClusterSettingOverrideOpt,
) {
	__antithesis_instrumentation__.Notify(500571)
	t.cfg = cfg
	t.serverArgs = &serverArgs
	t.clusterOpts = clusterOpts[:]
	t.tenantClusterSettingOverrideOpts = tenantClusterSettingOverrideOpts[:]

	tempExternalIODir, tempExternalIODirCleanup := testutils.TempDir(t.rootT)
	t.sharedIODir = tempExternalIODir
	t.testCleanupFuncs = append(t.testCleanupFuncs, tempExternalIODirCleanup)

	t.newCluster(serverArgs, t.clusterOpts, t.tenantClusterSettingOverrideOpts)

	if _, err := t.db.Exec(`
CREATE DATABASE test; USE test;
`); err != nil {
		__antithesis_instrumentation__.Notify(500574)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(500575)
	}
	__antithesis_instrumentation__.Notify(500572)

	if _, err := t.db.Exec(fmt.Sprintf("CREATE USER %s;", security.TestUser)); err != nil {
		__antithesis_instrumentation__.Notify(500576)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(500577)
	}
	__antithesis_instrumentation__.Notify(500573)

	t.labelMap = make(map[string]string)
	t.varMap = make(map[string]string)
	t.pendingStatements = make(map[string]pendingStatement)
	t.pendingQueries = make(map[string]pendingQuery)

	t.progress = 0
	t.failures = 0
	t.unsupported = 0
}

func applyBlocklistToConfigs(configs configSet, blocklist map[string]int) configSet {
	__antithesis_instrumentation__.Notify(500578)
	if len(blocklist) == 0 {
		__antithesis_instrumentation__.Notify(500581)
		return configs
	} else {
		__antithesis_instrumentation__.Notify(500582)
	}
	__antithesis_instrumentation__.Notify(500579)
	var newConfigs configSet
	for _, idx := range configs {
		__antithesis_instrumentation__.Notify(500583)
		if _, ok := blocklist[logicTestConfigIdxToName[idx]]; ok {
			__antithesis_instrumentation__.Notify(500585)
			continue
		} else {
			__antithesis_instrumentation__.Notify(500586)
		}
		__antithesis_instrumentation__.Notify(500584)
		newConfigs = append(newConfigs, idx)
	}
	__antithesis_instrumentation__.Notify(500580)
	return newConfigs
}

func getBlocklistIssueNo(blocklistDirective string) (string, int) {
	__antithesis_instrumentation__.Notify(500587)
	parts := strings.Split(blocklistDirective, "(")
	if len(parts) != 2 {
		__antithesis_instrumentation__.Notify(500590)
		return blocklistDirective, 0
	} else {
		__antithesis_instrumentation__.Notify(500591)
	}
	__antithesis_instrumentation__.Notify(500588)

	issueNo, err := strconv.Atoi(strings.TrimRight(parts[1], ")"))
	if err != nil {
		__antithesis_instrumentation__.Notify(500592)
		panic(fmt.Sprintf("possibly malformed blocklist directive: %s: %v", blocklistDirective, err))
	} else {
		__antithesis_instrumentation__.Notify(500593)
	}
	__antithesis_instrumentation__.Notify(500589)
	return parts[0], issueNo
}

func processConfigs(
	t *testing.T, path string, defaults configSet, configNames []string,
) (_ configSet, onlyNonMetamorphic bool) {
	__antithesis_instrumentation__.Notify(500594)
	const blocklistChar = '!'

	blocklist := make(map[string]int)
	allConfigNamesAreBlocklistDirectives := true
	for _, configName := range configNames {
		__antithesis_instrumentation__.Notify(500599)
		if configName[0] != blocklistChar {
			__antithesis_instrumentation__.Notify(500602)
			allConfigNamesAreBlocklistDirectives = false
			continue
		} else {
			__antithesis_instrumentation__.Notify(500603)
		}
		__antithesis_instrumentation__.Notify(500600)

		blockedConfig, issueNo := getBlocklistIssueNo(configName[1:])
		if *printBlocklistIssues && func() bool {
			__antithesis_instrumentation__.Notify(500604)
			return issueNo != 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(500605)
			t.Logf("will skip %s config in test %s due to issue: %s", blockedConfig, path, build.MakeIssueURL(issueNo))
		} else {
			__antithesis_instrumentation__.Notify(500606)
		}
		__antithesis_instrumentation__.Notify(500601)
		blocklist[blockedConfig] = issueNo
	}
	__antithesis_instrumentation__.Notify(500595)

	if _, ok := blocklist["metamorphic"]; ok && func() bool {
		__antithesis_instrumentation__.Notify(500607)
		return util.IsMetamorphicBuild() == true
	}() == true {
		__antithesis_instrumentation__.Notify(500608)
		onlyNonMetamorphic = true
	} else {
		__antithesis_instrumentation__.Notify(500609)
	}
	__antithesis_instrumentation__.Notify(500596)
	if len(blocklist) != 0 && func() bool {
		__antithesis_instrumentation__.Notify(500610)
		return allConfigNamesAreBlocklistDirectives == true
	}() == true {
		__antithesis_instrumentation__.Notify(500611)

		return applyBlocklistToConfigs(defaults, blocklist), onlyNonMetamorphic
	} else {
		__antithesis_instrumentation__.Notify(500612)
	}
	__antithesis_instrumentation__.Notify(500597)

	var configs configSet
	for _, configName := range configNames {
		__antithesis_instrumentation__.Notify(500613)
		if configName[0] == blocklistChar {
			__antithesis_instrumentation__.Notify(500616)
			continue
		} else {
			__antithesis_instrumentation__.Notify(500617)
		}
		__antithesis_instrumentation__.Notify(500614)
		if _, ok := blocklist[configName]; ok {
			__antithesis_instrumentation__.Notify(500618)
			continue
		} else {
			__antithesis_instrumentation__.Notify(500619)
		}
		__antithesis_instrumentation__.Notify(500615)

		idx, ok := findLogicTestConfig(configName)
		if !ok {
			__antithesis_instrumentation__.Notify(500620)
			switch configName {
			case defaultConfigName:
				__antithesis_instrumentation__.Notify(500621)
				configs = append(configs, applyBlocklistToConfigs(defaults, blocklist)...)
			case fiveNodeDefaultConfigName:
				__antithesis_instrumentation__.Notify(500622)
				configs = append(configs, applyBlocklistToConfigs(fiveNodeDefaultConfig, blocklist)...)
			default:
				__antithesis_instrumentation__.Notify(500623)
				t.Fatalf("%s: unknown config name %s", path, configName)
			}
		} else {
			__antithesis_instrumentation__.Notify(500624)
			configs = append(configs, idx)
		}
	}
	__antithesis_instrumentation__.Notify(500598)

	return configs, onlyNonMetamorphic
}

func readTestFileConfigs(
	t *testing.T, path string, defaults configSet,
) (_ configSet, onlyNonMetamorphic bool) {
	__antithesis_instrumentation__.Notify(500625)
	file, err := os.Open(path)
	if err != nil {
		__antithesis_instrumentation__.Notify(500628)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(500629)
	}
	__antithesis_instrumentation__.Notify(500626)
	defer file.Close()

	s := newLineScanner(file)
	for s.Scan() {
		__antithesis_instrumentation__.Notify(500630)
		fields := strings.Fields(s.Text())
		if len(fields) == 0 {
			__antithesis_instrumentation__.Notify(500633)
			continue
		} else {
			__antithesis_instrumentation__.Notify(500634)
		}
		__antithesis_instrumentation__.Notify(500631)
		cmd := fields[0]
		if !strings.HasPrefix(cmd, "#") {
			__antithesis_instrumentation__.Notify(500635)

			break
		} else {
			__antithesis_instrumentation__.Notify(500636)
		}
		__antithesis_instrumentation__.Notify(500632)

		if len(fields) > 1 && func() bool {
			__antithesis_instrumentation__.Notify(500637)
			return cmd == "#" == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(500638)
			return fields[1] == "LogicTest:" == true
		}() == true {
			__antithesis_instrumentation__.Notify(500639)
			if len(fields) == 2 {
				__antithesis_instrumentation__.Notify(500641)
				t.Fatalf("%s: empty LogicTest directive", path)
			} else {
				__antithesis_instrumentation__.Notify(500642)
			}
			__antithesis_instrumentation__.Notify(500640)
			return processConfigs(t, path, defaults, fields[2:])
		} else {
			__antithesis_instrumentation__.Notify(500643)
		}
	}
	__antithesis_instrumentation__.Notify(500627)

	return defaults, false
}

type tenantClusterSettingOverrideArgs struct {
	overrideMultiTenantZoneConfigsAllowed bool

	overrideMultiTenantMultiRegionAbstractionsAllowed bool
}

type tenantClusterSettingOverrideOpt interface {
	apply(*tenantClusterSettingOverrideArgs)
}

type tenantClusterSettingOverrideMultiTenantMultiRegionAbstractionsAllowed struct{}

var _ tenantClusterSettingOverrideOpt = &tenantClusterSettingOverrideMultiTenantMultiRegionAbstractionsAllowed{}

func (t tenantClusterSettingOverrideMultiTenantMultiRegionAbstractionsAllowed) apply(
	args *tenantClusterSettingOverrideArgs,
) {
	__antithesis_instrumentation__.Notify(500644)
	args.overrideMultiTenantMultiRegionAbstractionsAllowed = true
}

type tenantClusterSettingOverrideMultiTenantZoneConfigsAllowed struct{}

var _ tenantClusterSettingOverrideOpt = tenantClusterSettingOverrideMultiTenantZoneConfigsAllowed{}

func (t tenantClusterSettingOverrideMultiTenantZoneConfigsAllowed) apply(
	args *tenantClusterSettingOverrideArgs,
) {
	__antithesis_instrumentation__.Notify(500645)
	args.overrideMultiTenantZoneConfigsAllowed = true
}

type clusterOpt interface {
	apply(args *base.TestServerArgs)
}

type clusterOptDisableSpanConfigs struct{}

var _ clusterOpt = clusterOptDisableSpanConfigs{}

func (c clusterOptDisableSpanConfigs) apply(args *base.TestServerArgs) {
	__antithesis_instrumentation__.Notify(500646)
	args.DisableSpanConfigs = true
}

type clusterOptTracingOff struct{}

var _ clusterOpt = clusterOptTracingOff{}

func (c clusterOptTracingOff) apply(args *base.TestServerArgs) {
	__antithesis_instrumentation__.Notify(500647)
	args.TracingDefault = tracing.TracingModeOnDemand
}

type clusterOptIgnoreStrictGCForTenants struct{}

var _ clusterOpt = clusterOptIgnoreStrictGCForTenants{}

func (c clusterOptIgnoreStrictGCForTenants) apply(args *base.TestServerArgs) {
	__antithesis_instrumentation__.Notify(500648)
	_, ok := args.Knobs.Store.(*kvserver.StoreTestingKnobs)
	if !ok {
		__antithesis_instrumentation__.Notify(500650)
		args.Knobs.Store = &kvserver.StoreTestingKnobs{}
	} else {
		__antithesis_instrumentation__.Notify(500651)
	}
	__antithesis_instrumentation__.Notify(500649)
	args.Knobs.Store.(*kvserver.StoreTestingKnobs).IgnoreStrictGCEnforcement = true
}

func parseDirectiveOptions(t *testing.T, path string, directiveName string, f func(opt string)) {
	__antithesis_instrumentation__.Notify(500652)
	switch directiveName {
	case "cluster-opt", "tenant-cluster-setting-override-opt":
		__antithesis_instrumentation__.Notify(500654)

	default:
		__antithesis_instrumentation__.Notify(500655)
		t.Fatalf("cannot parse unknown directive %s", directiveName)
	}
	__antithesis_instrumentation__.Notify(500653)
	file, err := os.Open(path)
	require.NoError(t, err)
	defer file.Close()

	beginningOfFile := true
	directiveFound := false
	s := newLineScanner(file)

	for s.Scan() {
		__antithesis_instrumentation__.Notify(500656)
		fields := strings.Fields(s.Text())
		if len(fields) == 0 {
			__antithesis_instrumentation__.Notify(500659)
			continue
		} else {
			__antithesis_instrumentation__.Notify(500660)
		}
		__antithesis_instrumentation__.Notify(500657)
		cmd := fields[0]
		if !strings.HasPrefix(cmd, "#") {
			__antithesis_instrumentation__.Notify(500661)

			beginningOfFile = false
		} else {
			__antithesis_instrumentation__.Notify(500662)
		}
		__antithesis_instrumentation__.Notify(500658)

		if len(fields) > 1 && func() bool {
			__antithesis_instrumentation__.Notify(500663)
			return cmd == "#" == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(500664)
			return fields[1] == fmt.Sprintf("%s:", directiveName) == true
		}() == true {
			__antithesis_instrumentation__.Notify(500665)
			require.True(
				t,
				beginningOfFile,
				"%s directive needs to be at the beginning of file",
				directiveName,
			)
			require.False(
				t,
				directiveFound,
				"only one %s directive allowed per file; second one found: %s",
				directiveName,
				s.Text(),
			)
			directiveFound = true
			if len(fields) == 2 {
				__antithesis_instrumentation__.Notify(500667)
				t.Fatalf("%s: empty LogicTest directive", path)
			} else {
				__antithesis_instrumentation__.Notify(500668)
			}
			__antithesis_instrumentation__.Notify(500666)
			for _, opt := range fields[2:] {
				__antithesis_instrumentation__.Notify(500669)
				f(opt)
			}
		} else {
			__antithesis_instrumentation__.Notify(500670)
		}
	}
}

func readTenantClusterSettingOverrideArgs(
	t *testing.T, path string,
) []tenantClusterSettingOverrideOpt {
	__antithesis_instrumentation__.Notify(500671)
	file, err := os.Open(path)
	require.NoError(t, err)
	defer file.Close()

	var res []tenantClusterSettingOverrideOpt
	parseDirectiveOptions(t, path, "tenant-cluster-setting-override-opt", func(opt string) {
		__antithesis_instrumentation__.Notify(500673)
		switch opt {
		case "allow-zone-configs-for-secondary-tenants":
			__antithesis_instrumentation__.Notify(500674)
			res = append(res, tenantClusterSettingOverrideMultiTenantZoneConfigsAllowed{})
		case "allow-multi-region-abstractions-for-secondary-tenants":
			__antithesis_instrumentation__.Notify(500675)
			res = append(res, tenantClusterSettingOverrideMultiTenantMultiRegionAbstractionsAllowed{})
		default:
			__antithesis_instrumentation__.Notify(500676)
			t.Fatalf("unrecognized cluster option: %s", opt)
		}
	})
	__antithesis_instrumentation__.Notify(500672)
	return res
}

func readClusterOptions(t *testing.T, path string) []clusterOpt {
	__antithesis_instrumentation__.Notify(500677)
	var res []clusterOpt
	parseDirectiveOptions(t, path, "cluster-opt", func(opt string) {
		__antithesis_instrumentation__.Notify(500679)
		switch opt {
		case "disable-span-configs":
			__antithesis_instrumentation__.Notify(500680)
			res = append(res, clusterOptDisableSpanConfigs{})
		case "tracing-off":
			__antithesis_instrumentation__.Notify(500681)
			res = append(res, clusterOptTracingOff{})
		case "ignore-tenant-strict-gc-enforcement":
			__antithesis_instrumentation__.Notify(500682)
			res = append(res, clusterOptIgnoreStrictGCForTenants{})
		default:
			__antithesis_instrumentation__.Notify(500683)
			t.Fatalf("unrecognized cluster option: %s", opt)
		}
	})
	__antithesis_instrumentation__.Notify(500678)
	return res
}

type subtestDetails struct {
	name                  string
	buffer                *bytes.Buffer
	lineLineIndexIntoFile int
}

func (t *logicTest) processTestFile(path string, config testClusterConfig) error {
	__antithesis_instrumentation__.Notify(500684)
	rng, seed := randutil.NewPseudoRand()
	t.outf("rng seed: %d\n", seed)

	subtests, err := fetchSubtests(path)
	if err != nil {
		__antithesis_instrumentation__.Notify(500689)
		return err
	} else {
		__antithesis_instrumentation__.Notify(500690)
	}
	__antithesis_instrumentation__.Notify(500685)

	if *showSQL {
		__antithesis_instrumentation__.Notify(500691)
		t.outf("--- queries start here (file: %s)", path)
	} else {
		__antithesis_instrumentation__.Notify(500692)
	}
	__antithesis_instrumentation__.Notify(500686)
	defer t.printCompletion(path, config)

	for _, subtest := range subtests {
		__antithesis_instrumentation__.Notify(500693)
		if *maxErrs > 0 && func() bool {
			__antithesis_instrumentation__.Notify(500695)
			return t.failures >= *maxErrs == true
		}() == true {
			__antithesis_instrumentation__.Notify(500696)
			break
		} else {
			__antithesis_instrumentation__.Notify(500697)
		}
		__antithesis_instrumentation__.Notify(500694)

		if len(subtest.name) == 0 {
			__antithesis_instrumentation__.Notify(500698)
			if err := t.processSubtest(subtest, path, config, rng); err != nil {
				__antithesis_instrumentation__.Notify(500699)
				return err
			} else {
				__antithesis_instrumentation__.Notify(500700)
			}
		} else {
			__antithesis_instrumentation__.Notify(500701)
			t.emit(fmt.Sprintf("subtest %s", subtest.name))
			t.rootT.Run(subtest.name, func(subtestT *testing.T) {
				__antithesis_instrumentation__.Notify(500703)
				t.subtestT = subtestT
				defer func() {
					__antithesis_instrumentation__.Notify(500705)
					t.subtestT = nil
				}()
				__antithesis_instrumentation__.Notify(500704)
				if err := t.processSubtest(subtest, path, config, rng); err != nil {
					__antithesis_instrumentation__.Notify(500706)
					t.Error(err)
				} else {
					__antithesis_instrumentation__.Notify(500707)
				}
			})
			__antithesis_instrumentation__.Notify(500702)
			t.maybeSkipOnRetry(nil)
		}
	}
	__antithesis_instrumentation__.Notify(500687)

	if (*rewriteResultsInTestfiles || func() bool {
		__antithesis_instrumentation__.Notify(500708)
		return *rewriteSQL == true
	}() == true) && func() bool {
		__antithesis_instrumentation__.Notify(500709)
		return !t.rootT.Failed() == true
	}() == true {
		__antithesis_instrumentation__.Notify(500710)

		file, err := os.Create(path)
		if err != nil {
			__antithesis_instrumentation__.Notify(500713)
			return err
		} else {
			__antithesis_instrumentation__.Notify(500714)
		}
		__antithesis_instrumentation__.Notify(500711)
		defer file.Close()

		data := t.rewriteResTestBuf.String()
		if l := len(data); l > 2 && func() bool {
			__antithesis_instrumentation__.Notify(500715)
			return data[l-1] == '\n' == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(500716)
			return data[l-2] == '\n' == true
		}() == true {
			__antithesis_instrumentation__.Notify(500717)
			data = data[:l-1]
		} else {
			__antithesis_instrumentation__.Notify(500718)
		}
		__antithesis_instrumentation__.Notify(500712)
		fmt.Fprint(file, data)
	} else {
		__antithesis_instrumentation__.Notify(500719)
	}
	__antithesis_instrumentation__.Notify(500688)

	return nil
}

func (t *logicTest) hasOpenTxns(ctx context.Context) bool {
	__antithesis_instrumentation__.Notify(500720)
	for _, user := range t.clients {
		__antithesis_instrumentation__.Notify(500722)
		existingTxnPriority := "NORMAL"
		err := user.QueryRow("SHOW TRANSACTION PRIORITY").Scan(&existingTxnPriority)
		if err != nil {
			__antithesis_instrumentation__.Notify(500724)

			log.Warningf(ctx, "failed to check txn priority with %v", err)
			continue
		} else {
			__antithesis_instrumentation__.Notify(500725)
		}
		__antithesis_instrumentation__.Notify(500723)
		if _, err := user.Exec("SET TRANSACTION PRIORITY NORMAL;"); !testutils.IsError(err, "there is no transaction in progress") {
			__antithesis_instrumentation__.Notify(500726)

			_, err := user.Exec(fmt.Sprintf(`SET TRANSACTION PRIORITY %s`, existingTxnPriority))
			if err != nil {
				__antithesis_instrumentation__.Notify(500728)
				log.Warningf(ctx, "failed to reset txn priority with %v", err)
			} else {
				__antithesis_instrumentation__.Notify(500729)
			}
			__antithesis_instrumentation__.Notify(500727)
			return true
		} else {
			__antithesis_instrumentation__.Notify(500730)
		}
	}
	__antithesis_instrumentation__.Notify(500721)
	return false
}

func (t *logicTest) maybeBackupRestore(rng *rand.Rand, config testClusterConfig) error {
	__antithesis_instrumentation__.Notify(500731)
	if config.backupRestoreProbability != 0 && func() bool {
		__antithesis_instrumentation__.Notify(500740)
		return !config.isCCLConfig == true
	}() == true {
		__antithesis_instrumentation__.Notify(500741)
		return errors.Newf("logic test config %s specifies a backup restore probability but is not CCL",
			config.name)
	} else {
		__antithesis_instrumentation__.Notify(500742)
	}
	__antithesis_instrumentation__.Notify(500732)

	if rng.Float64() > config.backupRestoreProbability {
		__antithesis_instrumentation__.Notify(500743)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(500744)
	}
	__antithesis_instrumentation__.Notify(500733)

	if t.hasOpenTxns(context.Background()) {
		__antithesis_instrumentation__.Notify(500745)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(500746)
	}
	__antithesis_instrumentation__.Notify(500734)

	oldUser := t.user
	defer func() {
		__antithesis_instrumentation__.Notify(500747)
		t.setUser(oldUser, 0)
	}()
	__antithesis_instrumentation__.Notify(500735)

	users := make([]string, 0, len(t.clients))
	userToHexSession := make(map[string]string, len(t.clients))
	userToSessionVars := make(map[string]map[string]string, len(t.clients))
	for user := range t.clients {
		__antithesis_instrumentation__.Notify(500748)
		t.setUser(user, 0)
		users = append(users, user)

		var userSession string
		var err error
		if err = t.db.QueryRow(`SELECT encode(crdb_internal.serialize_session(), 'hex')`).Scan(&userSession); err == nil {
			__antithesis_instrumentation__.Notify(500752)
			userToHexSession[user] = userSession
			continue
		} else {
			__antithesis_instrumentation__.Notify(500753)
		}
		__antithesis_instrumentation__.Notify(500749)
		log.Warningf(context.Background(), "failed to serialize session: %+v", err)

		userSessionVars := make(map[string]string)
		existingSessionVars, err := t.db.Query("SHOW ALL")
		if err != nil {
			__antithesis_instrumentation__.Notify(500754)
			return err
		} else {
			__antithesis_instrumentation__.Notify(500755)
		}
		__antithesis_instrumentation__.Notify(500750)
		for existingSessionVars.Next() {
			__antithesis_instrumentation__.Notify(500756)
			var key, value string
			if err := existingSessionVars.Scan(&key, &value); err != nil {
				__antithesis_instrumentation__.Notify(500758)
				return errors.Wrap(err, "scanning session variables")
			} else {
				__antithesis_instrumentation__.Notify(500759)
			}
			__antithesis_instrumentation__.Notify(500757)
			userSessionVars[key] = value
		}
		__antithesis_instrumentation__.Notify(500751)
		userToSessionVars[user] = userSessionVars
	}
	__antithesis_instrumentation__.Notify(500736)

	backupLocation := fmt.Sprintf("nodelocal://1/logic-test-backup-%s",
		strconv.FormatInt(timeutil.Now().UnixNano(), 10))

	t.setUser(security.RootUser, 0)

	if _, err := t.db.Exec(fmt.Sprintf("BACKUP TO '%s'", backupLocation)); err != nil {
		__antithesis_instrumentation__.Notify(500760)
		return errors.Wrap(err, "backing up cluster")
	} else {
		__antithesis_instrumentation__.Notify(500761)
	}
	__antithesis_instrumentation__.Notify(500737)

	t.resetCluster()

	t.setUser(security.RootUser, 0)
	if _, err := t.db.Exec(fmt.Sprintf("RESTORE FROM '%s'", backupLocation)); err != nil {
		__antithesis_instrumentation__.Notify(500762)
		return errors.Wrap(err, "restoring cluster")
	} else {
		__antithesis_instrumentation__.Notify(500763)
	}
	__antithesis_instrumentation__.Notify(500738)

	for _, user := range users {
		__antithesis_instrumentation__.Notify(500764)

		t.setUser(user, 0)

		if userSession, ok := userToHexSession[user]; ok {
			__antithesis_instrumentation__.Notify(500765)
			if _, err := t.db.Exec(fmt.Sprintf(`SELECT crdb_internal.deserialize_session(decode('%s', 'hex'))`, userSession)); err != nil {
				__antithesis_instrumentation__.Notify(500766)
				return errors.Wrapf(err, "deserializing session")
			} else {
				__antithesis_instrumentation__.Notify(500767)
			}
		} else {
			__antithesis_instrumentation__.Notify(500768)
			if vars, ok := userToSessionVars[user]; ok {
				__antithesis_instrumentation__.Notify(500769)

				for key, value := range vars {
					__antithesis_instrumentation__.Notify(500770)

					if _, err := t.db.Exec(fmt.Sprintf("SET %s='%s'", key, value)); err != nil {
						__antithesis_instrumentation__.Notify(500771)

						log.Infof(context.Background(), "setting session variable as string failed (err: %v), trying as int", pretty.Formatter(err))
						if _, err := t.db.Exec(fmt.Sprintf("SET %s=%s", key, value)); err != nil {
							__antithesis_instrumentation__.Notify(500772)

							log.Infof(context.Background(), "setting session variable as int failed: %v (continuing anyway)", pretty.Formatter(err))
							continue
						} else {
							__antithesis_instrumentation__.Notify(500773)
						}
					} else {
						__antithesis_instrumentation__.Notify(500774)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(500775)
			}
		}
	}
	__antithesis_instrumentation__.Notify(500739)

	return nil
}

func fetchSubtests(path string) ([]subtestDetails, error) {
	__antithesis_instrumentation__.Notify(500776)
	file, err := os.Open(path)
	if err != nil {
		__antithesis_instrumentation__.Notify(500779)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(500780)
	}
	__antithesis_instrumentation__.Notify(500777)
	defer file.Close()

	s := newLineScanner(file)
	var subtests []subtestDetails
	var curName string
	var curLineIndexIntoFile int
	buffer := &bytes.Buffer{}
	for s.Scan() {
		__antithesis_instrumentation__.Notify(500781)
		line := s.Text()
		fields := strings.Fields(line)
		if len(fields) > 0 && func() bool {
			__antithesis_instrumentation__.Notify(500782)
			return fields[0] == "subtest" == true
		}() == true {
			__antithesis_instrumentation__.Notify(500783)
			if len(fields) != 2 {
				__antithesis_instrumentation__.Notify(500785)
				return nil, errors.Errorf(
					"%s:%d expected only one field following the subtest command\n"+
						"Note that this check does not respect the other commands so if a query result has a "+
						"line that starts with \"subtest\" it will either fail or be split into a subtest.",
					path, s.line,
				)
			} else {
				__antithesis_instrumentation__.Notify(500786)
			}
			__antithesis_instrumentation__.Notify(500784)
			subtests = append(subtests, subtestDetails{
				name:                  curName,
				buffer:                buffer,
				lineLineIndexIntoFile: curLineIndexIntoFile,
			})
			buffer = &bytes.Buffer{}
			curName = fields[1]
			curLineIndexIntoFile = s.line + 1
		} else {
			__antithesis_instrumentation__.Notify(500787)
			buffer.WriteString(line)
			buffer.WriteRune('\n')
		}
	}
	__antithesis_instrumentation__.Notify(500778)
	subtests = append(subtests, subtestDetails{
		name:                  curName,
		buffer:                buffer,
		lineLineIndexIntoFile: curLineIndexIntoFile,
	})

	return subtests, nil
}

func (t *logicTest) processSubtest(
	subtest subtestDetails, path string, config testClusterConfig, rng *rand.Rand,
) error {
	__antithesis_instrumentation__.Notify(500788)
	defer t.traceStop()

	s := newLineScanner(subtest.buffer)
	t.lastProgress = timeutil.Now()

	repeat := 1
	for s.Scan() {
		__antithesis_instrumentation__.Notify(500790)
		t.curPath, t.curLineNo = path, s.line+subtest.lineLineIndexIntoFile
		if *maxErrs > 0 && func() bool {
			__antithesis_instrumentation__.Notify(500796)
			return t.failures >= *maxErrs == true
		}() == true {
			__antithesis_instrumentation__.Notify(500797)
			return errors.Errorf("%s:%d: too many errors encountered, skipping the rest of the input",
				path, s.line+subtest.lineLineIndexIntoFile,
			)
		} else {
			__antithesis_instrumentation__.Notify(500798)
		}
		__antithesis_instrumentation__.Notify(500791)
		line := s.Text()
		t.emit(line)
		fields := strings.Fields(line)
		if len(fields) == 0 {
			__antithesis_instrumentation__.Notify(500799)
			continue
		} else {
			__antithesis_instrumentation__.Notify(500800)
		}
		__antithesis_instrumentation__.Notify(500792)
		cmd := fields[0]
		if strings.HasPrefix(cmd, "#") {
			__antithesis_instrumentation__.Notify(500801)

			continue
		} else {
			__antithesis_instrumentation__.Notify(500802)
		}
		__antithesis_instrumentation__.Notify(500793)
		if len(fields) == 2 && func() bool {
			__antithesis_instrumentation__.Notify(500803)
			return fields[1] == "error" == true
		}() == true {
			__antithesis_instrumentation__.Notify(500804)
			return errors.Errorf("%s:%d: no expected error provided",
				path, s.line+subtest.lineLineIndexIntoFile,
			)
		} else {
			__antithesis_instrumentation__.Notify(500805)
		}
		__antithesis_instrumentation__.Notify(500794)
		if err := t.maybeBackupRestore(rng, config); err != nil {
			__antithesis_instrumentation__.Notify(500806)
			return err
		} else {
			__antithesis_instrumentation__.Notify(500807)
		}
		__antithesis_instrumentation__.Notify(500795)
		switch cmd {
		case "repeat":
			__antithesis_instrumentation__.Notify(500808)

			var err error
			count := 0
			if len(fields) != 2 {
				__antithesis_instrumentation__.Notify(500860)
				err = errors.New("invalid line format")
			} else {
				__antithesis_instrumentation__.Notify(500861)
				if count, err = strconv.Atoi(fields[1]); err == nil && func() bool {
					__antithesis_instrumentation__.Notify(500862)
					return count < 2 == true
				}() == true {
					__antithesis_instrumentation__.Notify(500863)
					err = errors.New("invalid count")
				} else {
					__antithesis_instrumentation__.Notify(500864)
				}
			}
			__antithesis_instrumentation__.Notify(500809)
			if err != nil {
				__antithesis_instrumentation__.Notify(500865)
				return errors.Wrapf(err, "%s:%d invalid repeat line",
					path, s.line+subtest.lineLineIndexIntoFile,
				)
			} else {
				__antithesis_instrumentation__.Notify(500866)
			}
			__antithesis_instrumentation__.Notify(500810)
			repeat = count
		case "skip_on_retry":
			__antithesis_instrumentation__.Notify(500811)
			t.skipOnRetry = true

		case "sleep":
			__antithesis_instrumentation__.Notify(500812)
			var err error
			var duration time.Duration

			if len(fields) != 2 {
				__antithesis_instrumentation__.Notify(500867)
				err = errors.New("invalid line format")
			} else {
				__antithesis_instrumentation__.Notify(500868)
				if duration, err = time.ParseDuration(fields[1]); err != nil {
					__antithesis_instrumentation__.Notify(500869)
					err = errors.New("invalid duration")
				} else {
					__antithesis_instrumentation__.Notify(500870)
				}
			}
			__antithesis_instrumentation__.Notify(500813)
			if err != nil {
				__antithesis_instrumentation__.Notify(500871)
				return errors.Wrapf(err, "%s:%d invalid sleep line",
					path, s.line+subtest.lineLineIndexIntoFile,
				)
			} else {
				__antithesis_instrumentation__.Notify(500872)
			}
			__antithesis_instrumentation__.Notify(500814)
			time.Sleep(duration)

		case "awaitstatement":
			__antithesis_instrumentation__.Notify(500815)
			if len(fields) != 2 {
				__antithesis_instrumentation__.Notify(500873)
				return errors.New("invalid line format")
			} else {
				__antithesis_instrumentation__.Notify(500874)
			}
			__antithesis_instrumentation__.Notify(500816)

			name := fields[1]

			var pending pendingStatement
			var ok bool
			if pending, ok = t.pendingStatements[name]; !ok {
				__antithesis_instrumentation__.Notify(500875)
				return errors.Newf("pending statement with name %q unknown", name)
			} else {
				__antithesis_instrumentation__.Notify(500876)
			}
			__antithesis_instrumentation__.Notify(500817)

			execRes := <-pending.resultChan
			cont, err := t.finishExecStatement(pending.logicStatement, execRes.execSQL, execRes.res, execRes.err)

			if err != nil {
				__antithesis_instrumentation__.Notify(500877)
				if !cont {
					__antithesis_instrumentation__.Notify(500879)
					return err
				} else {
					__antithesis_instrumentation__.Notify(500880)
				}
				__antithesis_instrumentation__.Notify(500878)
				t.Error(err)
			} else {
				__antithesis_instrumentation__.Notify(500881)
			}
			__antithesis_instrumentation__.Notify(500818)

			delete(t.pendingStatements, name)

			t.success(path)

		case "statement":
			__antithesis_instrumentation__.Notify(500819)
			stmt := logicStatement{
				pos:         fmt.Sprintf("\n%s:%d", path, s.line+subtest.lineLineIndexIntoFile),
				expectCount: -1,
			}

			if m := noticeRE.FindStringSubmatch(s.Text()); m != nil {
				__antithesis_instrumentation__.Notify(500882)
				stmt.expectNotice = m[1]
			} else {
				__antithesis_instrumentation__.Notify(500883)
				if m := errorRE.FindStringSubmatch(s.Text()); m != nil {
					__antithesis_instrumentation__.Notify(500884)
					stmt.expectErrCode = m[1]
					stmt.expectErr = m[2]
				} else {
					__antithesis_instrumentation__.Notify(500885)
				}
			}
			__antithesis_instrumentation__.Notify(500820)
			if len(fields) >= 3 && func() bool {
				__antithesis_instrumentation__.Notify(500886)
				return fields[1] == "async" == true
			}() == true {
				__antithesis_instrumentation__.Notify(500887)
				stmt.expectAsync = true
				stmt.statementName = fields[2]
				copy(fields[1:], fields[3:])
				fields = fields[:len(fields)-2]
			} else {
				__antithesis_instrumentation__.Notify(500888)
			}
			__antithesis_instrumentation__.Notify(500821)
			if len(fields) >= 3 && func() bool {
				__antithesis_instrumentation__.Notify(500889)
				return fields[1] == "count" == true
			}() == true {
				__antithesis_instrumentation__.Notify(500890)
				n, err := strconv.ParseInt(fields[2], 10, 64)
				if err != nil {
					__antithesis_instrumentation__.Notify(500892)
					return err
				} else {
					__antithesis_instrumentation__.Notify(500893)
				}
				__antithesis_instrumentation__.Notify(500891)
				stmt.expectCount = n
			} else {
				__antithesis_instrumentation__.Notify(500894)
			}
			__antithesis_instrumentation__.Notify(500822)
			if _, err := stmt.readSQL(t, s, false); err != nil {
				__antithesis_instrumentation__.Notify(500895)
				return err
			} else {
				__antithesis_instrumentation__.Notify(500896)
			}
			__antithesis_instrumentation__.Notify(500823)
			if !s.skip {
				__antithesis_instrumentation__.Notify(500897)
				for i := 0; i < repeat; i++ {
					__antithesis_instrumentation__.Notify(500898)
					if cont, err := t.execStatement(stmt); err != nil {
						__antithesis_instrumentation__.Notify(500899)
						if !cont {
							__antithesis_instrumentation__.Notify(500901)
							return err
						} else {
							__antithesis_instrumentation__.Notify(500902)
						}
						__antithesis_instrumentation__.Notify(500900)
						t.Error(err)
					} else {
						__antithesis_instrumentation__.Notify(500903)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(500904)
				s.LogAndResetSkip(t)
			}
			__antithesis_instrumentation__.Notify(500824)
			repeat = 1
			t.success(path)

		case "awaitquery":
			__antithesis_instrumentation__.Notify(500825)
			if len(fields) != 2 {
				__antithesis_instrumentation__.Notify(500905)
				return errors.New("invalid line format")
			} else {
				__antithesis_instrumentation__.Notify(500906)
			}
			__antithesis_instrumentation__.Notify(500826)

			name := fields[1]

			var pending pendingQuery
			var ok bool
			if pending, ok = t.pendingQueries[name]; !ok {
				__antithesis_instrumentation__.Notify(500907)
				return errors.Newf("pending query with name %q unknown", name)
			} else {
				__antithesis_instrumentation__.Notify(500908)
			}
			__antithesis_instrumentation__.Notify(500827)

			execRes := <-pending.resultChan
			err := t.finishExecQuery(pending.logicQuery, execRes.rows, execRes.err)
			if err != nil {
				__antithesis_instrumentation__.Notify(500909)
				t.Error(err)
			} else {
				__antithesis_instrumentation__.Notify(500910)
			}
			__antithesis_instrumentation__.Notify(500828)

			delete(t.pendingQueries, name)
			t.success(path)

		case "query":
			__antithesis_instrumentation__.Notify(500829)
			var query logicQuery
			query.pos = fmt.Sprintf("\n%s:%d", path, s.line+subtest.lineLineIndexIntoFile)

			if m := errorRE.FindStringSubmatch(s.Text()); m != nil {
				__antithesis_instrumentation__.Notify(500911)
				query.expectErrCode = m[1]
				query.expectErr = m[2]
			} else {
				__antithesis_instrumentation__.Notify(500912)
				if len(fields) < 2 {
					__antithesis_instrumentation__.Notify(500913)
					return errors.Errorf("%s: invalid test statement: %s", query.pos, s.Text())
				} else {
					__antithesis_instrumentation__.Notify(500914)

					query.colTypes = fields[1]
					if *bigtest {
						__antithesis_instrumentation__.Notify(500917)

						query.valsPerLine = 1
					} else {
						__antithesis_instrumentation__.Notify(500918)

						query.valsPerLine = len(query.colTypes)
					}
					__antithesis_instrumentation__.Notify(500915)

					if len(fields) >= 3 {
						__antithesis_instrumentation__.Notify(500919)
						query.rawOpts = fields[2]

						tokens := strings.Split(query.rawOpts, ",")

						buildArgumentTokens := func(argToken string) {
							__antithesis_instrumentation__.Notify(500921)
							for i := 0; i < len(tokens)-1; i++ {
								__antithesis_instrumentation__.Notify(500922)
								if strings.HasPrefix(tokens[i], argToken+"(") && func() bool {
									__antithesis_instrumentation__.Notify(500923)
									return !strings.HasSuffix(tokens[i], ")") == true
								}() == true {
									__antithesis_instrumentation__.Notify(500924)

									tokens[i] = tokens[i] + "," + tokens[i+1]

									copy(tokens[i+1:], tokens[i+2:])
									tokens = tokens[:len(tokens)-1]

									i--
								} else {
									__antithesis_instrumentation__.Notify(500925)
								}
							}
						}
						__antithesis_instrumentation__.Notify(500920)

						buildArgumentTokens("partialsort")
						buildArgumentTokens("kvtrace")

						for _, opt := range tokens {
							__antithesis_instrumentation__.Notify(500926)
							if strings.HasPrefix(opt, "partialsort(") && func() bool {
								__antithesis_instrumentation__.Notify(500929)
								return strings.HasSuffix(opt, ")") == true
							}() == true {
								__antithesis_instrumentation__.Notify(500930)
								s := opt
								s = strings.TrimPrefix(s, "partialsort(")
								s = strings.TrimSuffix(s, ")")

								var orderedCols []int
								for _, c := range strings.Split(s, ",") {
									__antithesis_instrumentation__.Notify(500934)
									colIdx, err := strconv.Atoi(c)
									if err != nil || func() bool {
										__antithesis_instrumentation__.Notify(500936)
										return colIdx < 1 == true
									}() == true {
										__antithesis_instrumentation__.Notify(500937)
										return errors.Errorf("%s: invalid sort mode: %s", query.pos, opt)
									} else {
										__antithesis_instrumentation__.Notify(500938)
									}
									__antithesis_instrumentation__.Notify(500935)
									orderedCols = append(orderedCols, colIdx-1)
								}
								__antithesis_instrumentation__.Notify(500931)
								if len(orderedCols) == 0 {
									__antithesis_instrumentation__.Notify(500939)
									return errors.Errorf("%s: invalid sort mode: %s", query.pos, opt)
								} else {
									__antithesis_instrumentation__.Notify(500940)
								}
								__antithesis_instrumentation__.Notify(500932)
								query.sorter = func(numCols int, values []string) {
									__antithesis_instrumentation__.Notify(500941)
									partialSort(numCols, orderedCols, values)
								}
								__antithesis_instrumentation__.Notify(500933)
								continue
							} else {
								__antithesis_instrumentation__.Notify(500942)
							}
							__antithesis_instrumentation__.Notify(500927)

							if strings.HasPrefix(opt, "kvtrace(") && func() bool {
								__antithesis_instrumentation__.Notify(500943)
								return strings.HasSuffix(opt, ")") == true
							}() == true {
								__antithesis_instrumentation__.Notify(500944)
								s := opt
								s = strings.TrimPrefix(s, "kvtrace(")
								s = strings.TrimSuffix(s, ")")

								query.kvtrace = true
								query.kvOpTypes = nil
								query.keyPrefixFilters = nil
								for _, c := range strings.Split(s, ",") {
									__antithesis_instrumentation__.Notify(500946)
									if strings.HasPrefix(c, "prefix=") {
										__antithesis_instrumentation__.Notify(500947)
										matched := strings.TrimPrefix(c, "prefix=")
										query.keyPrefixFilters = append(query.keyPrefixFilters, matched)
									} else {
										__antithesis_instrumentation__.Notify(500948)
										if isAllowedKVOp(c) {
											__antithesis_instrumentation__.Notify(500949)
											query.kvOpTypes = append(query.kvOpTypes, c)
										} else {
											__antithesis_instrumentation__.Notify(500950)
											return errors.Errorf(
												"invalid filter '%s' provided. Expected one of %v or a prefix of the form 'prefix=x'",
												c,
												allowedKVOpTypes,
											)
										}
									}
								}
								__antithesis_instrumentation__.Notify(500945)
								continue
							} else {
								__antithesis_instrumentation__.Notify(500951)
							}
							__antithesis_instrumentation__.Notify(500928)

							switch opt {
							case "nosort":
								__antithesis_instrumentation__.Notify(500952)
								query.sorter = nil

							case "rowsort":
								__antithesis_instrumentation__.Notify(500953)
								query.sorter = rowSort

							case "valuesort":
								__antithesis_instrumentation__.Notify(500954)
								query.sorter = valueSort

							case "colnames":
								__antithesis_instrumentation__.Notify(500955)
								query.colNames = true

							case "retry":
								__antithesis_instrumentation__.Notify(500956)
								query.retry = true

							case "kvtrace":
								__antithesis_instrumentation__.Notify(500957)

								query.kvtrace = true
								query.kvOpTypes = nil
								query.keyPrefixFilters = nil

							case "noticetrace":
								__antithesis_instrumentation__.Notify(500958)
								query.noticetrace = true

							case "round-in-strings":
								__antithesis_instrumentation__.Notify(500959)
								query.roundFloatsInStrings = true

							case "async":
								__antithesis_instrumentation__.Notify(500960)
								query.expectAsync = true

							default:
								__antithesis_instrumentation__.Notify(500961)
								if strings.HasPrefix(opt, "nodeidx=") {
									__antithesis_instrumentation__.Notify(500963)
									idx, err := strconv.ParseInt(strings.SplitN(opt, "=", 2)[1], 10, 64)
									if err != nil {
										__antithesis_instrumentation__.Notify(500965)
										return errors.Wrapf(err, "error parsing nodeidx")
									} else {
										__antithesis_instrumentation__.Notify(500966)
									}
									__antithesis_instrumentation__.Notify(500964)
									query.nodeIdx = int(idx)
									break
								} else {
									__antithesis_instrumentation__.Notify(500967)
								}
								__antithesis_instrumentation__.Notify(500962)

								return errors.Errorf("%s: unknown sort mode: %s", query.pos, opt)
							}
						}
					} else {
						__antithesis_instrumentation__.Notify(500968)
					}
					__antithesis_instrumentation__.Notify(500916)
					if len(fields) >= 4 {
						__antithesis_instrumentation__.Notify(500969)
						query.label = fields[3]
						if query.expectAsync {
							__antithesis_instrumentation__.Notify(500970)
							query.statementName = fields[3]
						} else {
							__antithesis_instrumentation__.Notify(500971)
						}
					} else {
						__antithesis_instrumentation__.Notify(500972)
					}
				}
			}
			__antithesis_instrumentation__.Notify(500830)

			if query.noticetrace && func() bool {
				__antithesis_instrumentation__.Notify(500973)
				return query.kvtrace == true
			}() == true {
				__antithesis_instrumentation__.Notify(500974)
				return errors.Errorf(
					"%s: cannot have both noticetrace and kvtrace on at the same time",
					query.pos,
				)
			} else {
				__antithesis_instrumentation__.Notify(500975)
			}
			__antithesis_instrumentation__.Notify(500831)

			if query.expectAsync && func() bool {
				__antithesis_instrumentation__.Notify(500976)
				return query.statementName == "" == true
			}() == true {
				__antithesis_instrumentation__.Notify(500977)
				return errors.Errorf(
					"%s: cannot have async enabled without a label",
					query.pos,
				)
			} else {
				__antithesis_instrumentation__.Notify(500978)
			}
			__antithesis_instrumentation__.Notify(500832)

			separator, err := query.readSQL(t, s, true)
			if err != nil {
				__antithesis_instrumentation__.Notify(500979)
				return err
			} else {
				__antithesis_instrumentation__.Notify(500980)
			}
			__antithesis_instrumentation__.Notify(500833)

			query.checkResults = true
			if separator {
				__antithesis_instrumentation__.Notify(500981)

				if s.Scan() {
					__antithesis_instrumentation__.Notify(500982)
					if m := resultsRE.FindStringSubmatch(s.Text()); m != nil {
						__antithesis_instrumentation__.Notify(500983)
						var err error
						query.expectedValues, err = strconv.Atoi(m[1])
						if err != nil {
							__antithesis_instrumentation__.Notify(500985)
							return err
						} else {
							__antithesis_instrumentation__.Notify(500986)
						}
						__antithesis_instrumentation__.Notify(500984)
						query.expectedHash = m[2]
						query.checkResults = false
					} else {
						__antithesis_instrumentation__.Notify(500987)
						for {
							__antithesis_instrumentation__.Notify(500988)

							query.expectedResultsRaw = append(query.expectedResultsRaw, s.Text())
							results := strings.Fields(s.Text())
							if len(results) == 0 {
								__antithesis_instrumentation__.Notify(500991)
								break
							} else {
								__antithesis_instrumentation__.Notify(500992)
							}
							__antithesis_instrumentation__.Notify(500989)

							if query.sorter == nil {
								__antithesis_instrumentation__.Notify(500993)

								query.expectedResults = append(query.expectedResults, results...)
							} else {
								__antithesis_instrumentation__.Notify(500994)

								if query.valsPerLine == 1 {
									__antithesis_instrumentation__.Notify(500995)

									query.expectedResults = append(query.expectedResults, strings.Join(results, " "))
								} else {
									__antithesis_instrumentation__.Notify(500996)

									if !*rewriteResultsInTestfiles && func() bool {
										__antithesis_instrumentation__.Notify(500998)
										return len(results) != len(query.colTypes) == true
									}() == true {
										__antithesis_instrumentation__.Notify(500999)
										return errors.Errorf("expected results are invalid: unexpected column count")
									} else {
										__antithesis_instrumentation__.Notify(501000)
									}
									__antithesis_instrumentation__.Notify(500997)
									query.expectedResults = append(query.expectedResults, results...)
								}
							}
							__antithesis_instrumentation__.Notify(500990)

							if !s.Scan() {
								__antithesis_instrumentation__.Notify(501001)
								break
							} else {
								__antithesis_instrumentation__.Notify(501002)
							}
						}
					}
				} else {
					__antithesis_instrumentation__.Notify(501003)
				}
			} else {
				__antithesis_instrumentation__.Notify(501004)
				if query.label != "" {
					__antithesis_instrumentation__.Notify(501005)

					query.checkResults = false
				} else {
					__antithesis_instrumentation__.Notify(501006)
				}
			}
			__antithesis_instrumentation__.Notify(500834)

			if !s.skip {
				__antithesis_instrumentation__.Notify(501007)
				if query.kvtrace {
					__antithesis_instrumentation__.Notify(501010)
					_, err := t.db.Exec("SET TRACING=on,kv")
					if err != nil {
						__antithesis_instrumentation__.Notify(501015)
						return err
					} else {
						__antithesis_instrumentation__.Notify(501016)
					}
					__antithesis_instrumentation__.Notify(501011)
					_, err = t.db.Exec(query.sql)
					if err != nil {
						__antithesis_instrumentation__.Notify(501017)
						t.Error(err)
					} else {
						__antithesis_instrumentation__.Notify(501018)
					}
					__antithesis_instrumentation__.Notify(501012)
					_, err = t.db.Exec("SET TRACING=off")
					if err != nil {
						__antithesis_instrumentation__.Notify(501019)
						return err
					} else {
						__antithesis_instrumentation__.Notify(501020)
					}
					__antithesis_instrumentation__.Notify(501013)

					queryPrefix := `SELECT message FROM [SHOW KV TRACE FOR SESSION] `
					buildQuery := func(ops []string, keyFilters []string) string {
						__antithesis_instrumentation__.Notify(501021)
						var sb strings.Builder
						sb.WriteString(queryPrefix)
						if len(keyFilters) == 0 {
							__antithesis_instrumentation__.Notify(501024)
							keyFilters = []string{""}
						} else {
							__antithesis_instrumentation__.Notify(501025)
						}
						__antithesis_instrumentation__.Notify(501022)
						for i, c := range ops {
							__antithesis_instrumentation__.Notify(501026)
							for j, f := range keyFilters {
								__antithesis_instrumentation__.Notify(501027)
								if i+j == 0 {
									__antithesis_instrumentation__.Notify(501029)
									sb.WriteString("WHERE ")
								} else {
									__antithesis_instrumentation__.Notify(501030)
									sb.WriteString("OR ")
								}
								__antithesis_instrumentation__.Notify(501028)
								sb.WriteString(fmt.Sprintf("message like '%s %s%%'", c, f))
							}
						}
						__antithesis_instrumentation__.Notify(501023)
						return sb.String()
					}
					__antithesis_instrumentation__.Notify(501014)

					query.colTypes = "T"
					if len(query.kvOpTypes) == 0 {
						__antithesis_instrumentation__.Notify(501031)
						query.sql = buildQuery(allowedKVOpTypes, query.keyPrefixFilters)
					} else {
						__antithesis_instrumentation__.Notify(501032)
						query.sql = buildQuery(query.kvOpTypes, query.keyPrefixFilters)
					}
				} else {
					__antithesis_instrumentation__.Notify(501033)
				}
				__antithesis_instrumentation__.Notify(501008)

				if query.noticetrace {
					__antithesis_instrumentation__.Notify(501034)
					query.colTypes = "T"
				} else {
					__antithesis_instrumentation__.Notify(501035)
				}
				__antithesis_instrumentation__.Notify(501009)

				for i := 0; i < repeat; i++ {
					__antithesis_instrumentation__.Notify(501036)
					if query.retry && func() bool {
						__antithesis_instrumentation__.Notify(501037)
						return !*rewriteResultsInTestfiles == true
					}() == true {
						__antithesis_instrumentation__.Notify(501038)
						testutils.SucceedsSoon(t.rootT, func() error {
							__antithesis_instrumentation__.Notify(501039)
							return t.execQuery(query)
						})
					} else {
						__antithesis_instrumentation__.Notify(501040)
						if query.retry && func() bool {
							__antithesis_instrumentation__.Notify(501042)
							return *rewriteResultsInTestfiles == true
						}() == true {
							__antithesis_instrumentation__.Notify(501043)

							time.Sleep(time.Millisecond * 500)
						} else {
							__antithesis_instrumentation__.Notify(501044)
						}
						__antithesis_instrumentation__.Notify(501041)
						if err := t.execQuery(query); err != nil {
							__antithesis_instrumentation__.Notify(501045)
							t.Error(err)
						} else {
							__antithesis_instrumentation__.Notify(501046)
						}
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(501047)
				if *rewriteResultsInTestfiles {
					__antithesis_instrumentation__.Notify(501049)
					for _, l := range query.expectedResultsRaw {
						__antithesis_instrumentation__.Notify(501050)
						t.emit(l)
					}
				} else {
					__antithesis_instrumentation__.Notify(501051)
				}
				__antithesis_instrumentation__.Notify(501048)
				s.LogAndResetSkip(t)
			}
			__antithesis_instrumentation__.Notify(500835)
			repeat = 1
			t.success(path)

		case "let":
			__antithesis_instrumentation__.Notify(500836)

			if len(fields) != 2 {
				__antithesis_instrumentation__.Notify(501052)
				return errors.Errorf("let command requires one argument, found: %v", fields)
			} else {
				__antithesis_instrumentation__.Notify(501053)
			}
			__antithesis_instrumentation__.Notify(500837)
			varName := fields[1]
			if !varRE.MatchString(varName) {
				__antithesis_instrumentation__.Notify(501054)
				return errors.Errorf("invalid target name for let: %s", varName)
			} else {
				__antithesis_instrumentation__.Notify(501055)
			}
			__antithesis_instrumentation__.Notify(500838)

			stmt := logicStatement{
				pos: fmt.Sprintf("\n%s:%d", path, s.line+subtest.lineLineIndexIntoFile),
			}
			if _, err := stmt.readSQL(t, s, false); err != nil {
				__antithesis_instrumentation__.Notify(501056)
				return err
			} else {
				__antithesis_instrumentation__.Notify(501057)
			}
			__antithesis_instrumentation__.Notify(500839)
			rows, err := t.db.Query(stmt.sql)
			if err != nil {
				__antithesis_instrumentation__.Notify(501058)
				return errors.Wrapf(err, "%s: error running query %s", stmt.pos, stmt.sql)
			} else {
				__antithesis_instrumentation__.Notify(501059)
			}
			__antithesis_instrumentation__.Notify(500840)
			if !rows.Next() {
				__antithesis_instrumentation__.Notify(501060)
				return errors.Errorf("%s: no rows returned by query %s", stmt.pos, stmt.sql)
			} else {
				__antithesis_instrumentation__.Notify(501061)
			}
			__antithesis_instrumentation__.Notify(500841)
			var val string
			if err := rows.Scan(&val); err != nil {
				__antithesis_instrumentation__.Notify(501062)
				return errors.Wrapf(err, "%s: error getting result from query %s", stmt.pos, stmt.sql)
			} else {
				__antithesis_instrumentation__.Notify(501063)
			}
			__antithesis_instrumentation__.Notify(500842)
			if rows.Next() {
				__antithesis_instrumentation__.Notify(501064)
				return errors.Errorf("%s: more than one row returned by query  %s", stmt.pos, stmt.sql)
			} else {
				__antithesis_instrumentation__.Notify(501065)
			}
			__antithesis_instrumentation__.Notify(500843)
			t.t().Logf("let %s = %s\n", varName, val)
			t.varMap[varName] = val

		case "halt", "hash-threshold":
			__antithesis_instrumentation__.Notify(500844)

		case "user":
			__antithesis_instrumentation__.Notify(500845)
			var nodeIdx int
			if len(fields) < 2 {
				__antithesis_instrumentation__.Notify(501066)
				return errors.Errorf("user command requires one argument, found: %v", fields)
			} else {
				__antithesis_instrumentation__.Notify(501067)
			}
			__antithesis_instrumentation__.Notify(500846)
			if len(fields[1]) == 0 {
				__antithesis_instrumentation__.Notify(501068)
				return errors.Errorf("user command requires a non-blank argument")
			} else {
				__antithesis_instrumentation__.Notify(501069)
			}
			__antithesis_instrumentation__.Notify(500847)
			if len(fields) >= 3 {
				__antithesis_instrumentation__.Notify(501070)
				if strings.HasPrefix(fields[2], "nodeidx=") {
					__antithesis_instrumentation__.Notify(501071)
					idx, err := strconv.ParseInt(strings.SplitN(fields[2], "=", 2)[1], 10, 64)
					if err != nil {
						__antithesis_instrumentation__.Notify(501073)
						return errors.Wrapf(err, "error parsing nodeidx")
					} else {
						__antithesis_instrumentation__.Notify(501074)
					}
					__antithesis_instrumentation__.Notify(501072)
					nodeIdx = int(idx)
				} else {
					__antithesis_instrumentation__.Notify(501075)
				}
			} else {
				__antithesis_instrumentation__.Notify(501076)
			}
			__antithesis_instrumentation__.Notify(500848)
			cleanupUserFunc := t.setUser(fields[1], nodeIdx)
			defer cleanupUserFunc()

		case "skip":
			__antithesis_instrumentation__.Notify(500849)
			reason := "skipped"
			if len(fields) > 1 {
				__antithesis_instrumentation__.Notify(501077)
				reason = fields[1]
			} else {
				__antithesis_instrumentation__.Notify(501078)
			}
			__antithesis_instrumentation__.Notify(500850)
			skip.IgnoreLint(t.t(), reason)

		case "skipif":
			__antithesis_instrumentation__.Notify(500851)
			if len(fields) < 2 {
				__antithesis_instrumentation__.Notify(501079)
				return errors.Errorf("skipif command requires one argument, found: %v", fields)
			} else {
				__antithesis_instrumentation__.Notify(501080)
			}
			__antithesis_instrumentation__.Notify(500852)
			switch fields[1] {
			case "":
				__antithesis_instrumentation__.Notify(501081)
				return errors.Errorf("skipif command requires a non-blank argument")
			case "mysql", "mssql":
				__antithesis_instrumentation__.Notify(501082)
			case "postgresql", "cockroachdb":
				__antithesis_instrumentation__.Notify(501083)
				s.SetSkip("")
				continue
			case "config":
				__antithesis_instrumentation__.Notify(501084)
				if len(fields) < 3 {
					__antithesis_instrumentation__.Notify(501088)
					return errors.New("skipif config CONFIG [ISSUE] command requires configuration parameter")
				} else {
					__antithesis_instrumentation__.Notify(501089)
				}
				__antithesis_instrumentation__.Notify(501085)
				configName := fields[2]
				if t.cfg.name == configName {
					__antithesis_instrumentation__.Notify(501090)
					issue := "no issue given"
					if len(fields) > 3 {
						__antithesis_instrumentation__.Notify(501092)
						issue = fields[3]
					} else {
						__antithesis_instrumentation__.Notify(501093)
					}
					__antithesis_instrumentation__.Notify(501091)
					s.SetSkip(fmt.Sprintf("unsupported configuration %s (%s)", configName, issue))
				} else {
					__antithesis_instrumentation__.Notify(501094)
				}
				__antithesis_instrumentation__.Notify(501086)
				continue
			default:
				__antithesis_instrumentation__.Notify(501087)
				return errors.Errorf("unimplemented test statement: %s", s.Text())
			}

		case "onlyif":
			__antithesis_instrumentation__.Notify(500853)
			if len(fields) < 2 {
				__antithesis_instrumentation__.Notify(501095)
				return errors.Errorf("onlyif command requires one argument, found: %v", fields)
			} else {
				__antithesis_instrumentation__.Notify(501096)
			}
			__antithesis_instrumentation__.Notify(500854)
			switch fields[1] {
			case "":
				__antithesis_instrumentation__.Notify(501097)
				return errors.New("onlyif command requires a non-blank argument")
			case "cockroachdb":
				__antithesis_instrumentation__.Notify(501098)
			case "mysql", "mssql":
				__antithesis_instrumentation__.Notify(501099)
				s.SetSkip("")
				continue
			case "config":
				__antithesis_instrumentation__.Notify(501100)
				if len(fields) < 3 {
					__antithesis_instrumentation__.Notify(501104)
					return errors.New("onlyif config CONFIG [ISSUE] command requires configuration parameter")
				} else {
					__antithesis_instrumentation__.Notify(501105)
				}
				__antithesis_instrumentation__.Notify(501101)
				configName := fields[2]
				if t.cfg.name != configName {
					__antithesis_instrumentation__.Notify(501106)
					issue := "no issue given"
					if len(fields) > 3 {
						__antithesis_instrumentation__.Notify(501108)
						issue = fields[3]
					} else {
						__antithesis_instrumentation__.Notify(501109)
					}
					__antithesis_instrumentation__.Notify(501107)
					s.SetSkip(fmt.Sprintf("unsupported configuration %s, statement/query only supports %s (%s)", t.cfg.name, configName, issue))
				} else {
					__antithesis_instrumentation__.Notify(501110)
				}
				__antithesis_instrumentation__.Notify(501102)
				continue
			default:
				__antithesis_instrumentation__.Notify(501103)
				return errors.Errorf("unimplemented test statement: %s", s.Text())
			}

		case "traceon":
			__antithesis_instrumentation__.Notify(500855)
			if len(fields) != 2 {
				__antithesis_instrumentation__.Notify(501111)
				return errors.Errorf("traceon requires a filename argument, found: %v", fields)
			} else {
				__antithesis_instrumentation__.Notify(501112)
			}
			__antithesis_instrumentation__.Notify(500856)
			t.traceStart(fields[1])

		case "traceoff":
			__antithesis_instrumentation__.Notify(500857)
			if t.traceFile == nil {
				__antithesis_instrumentation__.Notify(501113)
				return errors.Errorf("no trace active")
			} else {
				__antithesis_instrumentation__.Notify(501114)
			}
			__antithesis_instrumentation__.Notify(500858)
			t.traceStop()

		default:
			__antithesis_instrumentation__.Notify(500859)
			return errors.Errorf("%s:%d: unknown command: %s",
				path, s.line+subtest.lineLineIndexIntoFile, cmd,
			)
		}
	}
	__antithesis_instrumentation__.Notify(500789)
	return s.Err()
}

func (t *logicTest) maybeSkipOnRetry(err error) {
	__antithesis_instrumentation__.Notify(501115)
	if pqErr := (*pq.Error)(nil); t.skippedOnRetry || func() bool {
		__antithesis_instrumentation__.Notify(501116)
		return (t.skipOnRetry && func() bool {
			__antithesis_instrumentation__.Notify(501117)
			return errors.As(err, &pqErr) == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(501118)
			return pgcode.MakeCode(string(pqErr.Code)) == pgcode.SerializationFailure == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(501119)
		t.skippedOnRetry = true
		skip.WithIssue(t.t(), 53724)
	} else {
		__antithesis_instrumentation__.Notify(501120)
	}
}

func (t *logicTest) verifyError(
	sql, pos, expectNotice, expectErr, expectErrCode string, err error,
) (bool, error) {
	__antithesis_instrumentation__.Notify(501121)
	if expectErr == "" && func() bool {
		__antithesis_instrumentation__.Notify(501127)
		return expectErrCode == "" == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(501128)
		return err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(501129)
		t.maybeSkipOnRetry(err)
		cont := t.unexpectedError(sql, pos, err)
		if cont {
			__antithesis_instrumentation__.Notify(501131)

			err = nil
		} else {
			__antithesis_instrumentation__.Notify(501132)
		}
		__antithesis_instrumentation__.Notify(501130)
		return cont, err
	} else {
		__antithesis_instrumentation__.Notify(501133)
	}
	__antithesis_instrumentation__.Notify(501122)
	if expectNotice != "" {
		__antithesis_instrumentation__.Notify(501134)
		foundNotice := strings.Join(t.noticeBuffer, "\n")
		match, _ := regexp.MatchString(expectNotice, foundNotice)
		if !match {
			__antithesis_instrumentation__.Notify(501136)
			return false, errors.Errorf("%s: %s\nexpected notice pattern:\n%s\n\ngot:\n%s", pos, sql, expectNotice, foundNotice)
		} else {
			__antithesis_instrumentation__.Notify(501137)
		}
		__antithesis_instrumentation__.Notify(501135)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(501138)
	}
	__antithesis_instrumentation__.Notify(501123)
	if !testutils.IsError(err, expectErr) {
		__antithesis_instrumentation__.Notify(501139)
		if err == nil {
			__antithesis_instrumentation__.Notify(501143)
			newErr := errors.Errorf("%s: %s\nexpected %q, but no error occurred", pos, sql, expectErr)
			return false, newErr
		} else {
			__antithesis_instrumentation__.Notify(501144)
		}
		__antithesis_instrumentation__.Notify(501140)

		errString := pgerror.FullError(err)
		newErr := errors.Errorf("%s: %s\nexpected:\n%s\n\ngot:\n%s", pos, sql, expectErr, errString)
		if strings.Contains(errString, expectErr) {
			__antithesis_instrumentation__.Notify(501145)
			t.t().Logf("The output string contained the input regexp. Perhaps you meant to write:\n"+
				"query error %s", regexp.QuoteMeta(errString))
		} else {
			__antithesis_instrumentation__.Notify(501146)
		}
		__antithesis_instrumentation__.Notify(501141)

		if *rewriteResultsInTestfiles {
			__antithesis_instrumentation__.Notify(501147)
			r := regexp.QuoteMeta(errString)
			r = strings.Trim(r, "\n ")
			r = strings.ReplaceAll(r, "\n", "\\n")
			t.t().Logf("Error regexp: %s\n", r)
		} else {
			__antithesis_instrumentation__.Notify(501148)
		}
		__antithesis_instrumentation__.Notify(501142)
		return expectErr != "", newErr
	} else {
		__antithesis_instrumentation__.Notify(501149)
	}
	__antithesis_instrumentation__.Notify(501124)
	if err != nil {
		__antithesis_instrumentation__.Notify(501150)
		if pqErr := (*pq.Error)(nil); errors.As(err, &pqErr) && func() bool {
			__antithesis_instrumentation__.Notify(501151)
			return strings.HasPrefix(string(pqErr.Code), "XX") == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(501152)
			return pgcode.MakeCode(string(pqErr.Code)) != pgcode.Uncategorized == true
		}() == true {
			__antithesis_instrumentation__.Notify(501153)
			if expectErrCode != string(pqErr.Code) {
				__antithesis_instrumentation__.Notify(501154)
				return false, errors.Errorf(
					"%s: %s: serious error with code %q occurred; if expected, must use 'error pgcode %s ...' in test:\n%s",
					pos, sql, pqErr.Code, pqErr.Code, pgerror.FullError(err))
			} else {
				__antithesis_instrumentation__.Notify(501155)
			}
		} else {
			__antithesis_instrumentation__.Notify(501156)
		}
	} else {
		__antithesis_instrumentation__.Notify(501157)
	}
	__antithesis_instrumentation__.Notify(501125)
	if expectErrCode != "" {
		__antithesis_instrumentation__.Notify(501158)
		if err != nil {
			__antithesis_instrumentation__.Notify(501159)
			var pqErr *pq.Error
			if !errors.As(err, &pqErr) {
				__antithesis_instrumentation__.Notify(501161)
				newErr := errors.Errorf("%s %s\n: expected error code %q, but the error we found is not "+
					"a libpq error: %s", pos, sql, expectErrCode, err)
				return true, newErr
			} else {
				__antithesis_instrumentation__.Notify(501162)
			}
			__antithesis_instrumentation__.Notify(501160)
			if pqErr.Code != pq.ErrorCode(expectErrCode) {
				__antithesis_instrumentation__.Notify(501163)
				newErr := errors.Errorf("%s: %s\nexpected error code %q, but found code %q (%s)",
					pos, sql, expectErrCode, pqErr.Code, pqErr.Code.Name())
				return true, newErr
			} else {
				__antithesis_instrumentation__.Notify(501164)
			}
		} else {
			__antithesis_instrumentation__.Notify(501165)
			newErr := errors.Errorf("%s: %s\nexpected error code %q, but found success",
				pos, sql, expectErrCode)
			return (err != nil), newErr
		}
	} else {
		__antithesis_instrumentation__.Notify(501166)
	}
	__antithesis_instrumentation__.Notify(501126)
	return true, nil
}

func formatErr(err error) string {
	__antithesis_instrumentation__.Notify(501167)
	if pqErr := (*pq.Error)(nil); errors.As(err, &pqErr) {
		__antithesis_instrumentation__.Notify(501169)
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "(%s) %s", pqErr.Code, pqErr.Message)
		if pqErr.File != "" || func() bool {
			__antithesis_instrumentation__.Notify(501173)
			return pqErr.Line != "" == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(501174)
			return pqErr.Routine != "" == true
		}() == true {
			__antithesis_instrumentation__.Notify(501175)
			fmt.Fprintf(&buf, "\n%s:%s: in %s()", pqErr.File, pqErr.Line, pqErr.Routine)
		} else {
			__antithesis_instrumentation__.Notify(501176)
		}
		__antithesis_instrumentation__.Notify(501170)
		if pqErr.Detail != "" {
			__antithesis_instrumentation__.Notify(501177)
			fmt.Fprintf(&buf, "\nDETAIL: %s", pqErr.Detail)
		} else {
			__antithesis_instrumentation__.Notify(501178)
		}
		__antithesis_instrumentation__.Notify(501171)
		if pgcode.MakeCode(string(pqErr.Code)) == pgcode.Internal {
			__antithesis_instrumentation__.Notify(501179)
			fmt.Fprintln(&buf, "\nNOTE: internal errors may have more details in logs. Use -show-logs.")
		} else {
			__antithesis_instrumentation__.Notify(501180)
		}
		__antithesis_instrumentation__.Notify(501172)
		return buf.String()
	} else {
		__antithesis_instrumentation__.Notify(501181)
	}
	__antithesis_instrumentation__.Notify(501168)
	return err.Error()
}

func (t *logicTest) unexpectedError(sql string, pos string, err error) bool {
	__antithesis_instrumentation__.Notify(501182)
	if *allowPrepareFail && func() bool {
		__antithesis_instrumentation__.Notify(501184)
		return sql != "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(501185)

		stmt, err := t.db.Prepare(sql)
		if err != nil {
			__antithesis_instrumentation__.Notify(501187)
			if *showSQL {
				__antithesis_instrumentation__.Notify(501189)
				t.outf("\t-- fails prepare: %s", formatErr(err))
			} else {
				__antithesis_instrumentation__.Notify(501190)
			}
			__antithesis_instrumentation__.Notify(501188)
			t.signalIgnoredError(err, pos, sql)
			return true
		} else {
			__antithesis_instrumentation__.Notify(501191)
		}
		__antithesis_instrumentation__.Notify(501186)
		if err := stmt.Close(); err != nil {
			__antithesis_instrumentation__.Notify(501192)
			t.Errorf("%s: %s\nerror when closing prepared statement: %s", sql, pos, formatErr(err))
		} else {
			__antithesis_instrumentation__.Notify(501193)
		}
	} else {
		__antithesis_instrumentation__.Notify(501194)
	}
	__antithesis_instrumentation__.Notify(501183)
	t.Errorf("%s: %s\nexpected success, but found\n%s", pos, sql, formatErr(err))
	return false
}

func (t *logicTest) execStatement(stmt logicStatement) (bool, error) {
	__antithesis_instrumentation__.Notify(501195)
	db := t.db
	t.noticeBuffer = nil
	if *showSQL {
		__antithesis_instrumentation__.Notify(501199)
		t.outf("%s;", stmt.sql)
	} else {
		__antithesis_instrumentation__.Notify(501200)
	}
	__antithesis_instrumentation__.Notify(501196)
	execSQL, changed := randgen.ApplyString(t.rng, stmt.sql, randgen.ColumnFamilyMutator)
	if changed {
		__antithesis_instrumentation__.Notify(501201)
		log.Infof(context.Background(), "Rewrote test statement:\n%s", execSQL)
		if *showSQL {
			__antithesis_instrumentation__.Notify(501202)
			t.outf("rewrote:\n%s\n", execSQL)
		} else {
			__antithesis_instrumentation__.Notify(501203)
		}
	} else {
		__antithesis_instrumentation__.Notify(501204)
	}
	__antithesis_instrumentation__.Notify(501197)

	if stmt.expectAsync {
		__antithesis_instrumentation__.Notify(501205)
		if _, ok := t.pendingStatements[stmt.statementName]; ok {
			__antithesis_instrumentation__.Notify(501208)
			return false, errors.Newf("pending statement with name %q already exists", stmt.statementName)
		} else {
			__antithesis_instrumentation__.Notify(501209)
		}
		__antithesis_instrumentation__.Notify(501206)

		pending := pendingStatement{
			logicStatement: stmt,
			resultChan:     make(chan pendingExecResult),
		}
		t.pendingStatements[stmt.statementName] = pending

		startedChan := make(chan struct{})
		go func() {
			__antithesis_instrumentation__.Notify(501210)
			startedChan <- struct{}{}
			res, err := db.Exec(execSQL)
			pending.resultChan <- pendingExecResult{execSQL, res, err}
		}()
		__antithesis_instrumentation__.Notify(501207)

		<-startedChan
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(501211)
	}
	__antithesis_instrumentation__.Notify(501198)

	res, err := db.Exec(execSQL)
	return t.finishExecStatement(stmt, execSQL, res, err)
}

func (t *logicTest) finishExecStatement(
	stmt logicStatement, execSQL string, res gosql.Result, err error,
) (bool, error) {
	__antithesis_instrumentation__.Notify(501212)
	if err == nil {
		__antithesis_instrumentation__.Notify(501216)
		sqlutils.VerifyStatementPrettyRoundtrip(t.t(), stmt.sql)
	} else {
		__antithesis_instrumentation__.Notify(501217)
	}
	__antithesis_instrumentation__.Notify(501213)
	if err == nil && func() bool {
		__antithesis_instrumentation__.Notify(501218)
		return stmt.expectCount >= 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(501219)
		var count int64
		count, err = res.RowsAffected()

		if err == nil && func() bool {
			__antithesis_instrumentation__.Notify(501220)
			return count != stmt.expectCount == true
		}() == true {
			__antithesis_instrumentation__.Notify(501221)
			t.Errorf("%s: %s\nexpected %d rows affected, got %d", stmt.pos, execSQL, stmt.expectCount, count)
		} else {
			__antithesis_instrumentation__.Notify(501222)
		}
	} else {
		__antithesis_instrumentation__.Notify(501223)
	}
	__antithesis_instrumentation__.Notify(501214)

	cont, err := t.verifyError("", stmt.pos, stmt.expectNotice, stmt.expectErr, stmt.expectErrCode, err)
	if err != nil {
		__antithesis_instrumentation__.Notify(501224)
		t.finishOne("OK")
	} else {
		__antithesis_instrumentation__.Notify(501225)
	}
	__antithesis_instrumentation__.Notify(501215)
	return cont, err
}

func (t *logicTest) hashResults(results []string) (string, error) {
	__antithesis_instrumentation__.Notify(501226)

	h := md5.New()
	for _, r := range results {
		__antithesis_instrumentation__.Notify(501228)
		if _, err := h.Write(append([]byte(r), byte('\n'))); err != nil {
			__antithesis_instrumentation__.Notify(501229)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(501230)
		}
	}
	__antithesis_instrumentation__.Notify(501227)
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func (t *logicTest) execQuery(query logicQuery) error {
	__antithesis_instrumentation__.Notify(501231)
	if *showSQL {
		__antithesis_instrumentation__.Notify(501235)
		t.outf("%s;", query.sql)
	} else {
		__antithesis_instrumentation__.Notify(501236)
	}
	__antithesis_instrumentation__.Notify(501232)

	t.noticeBuffer = nil

	db := t.db
	var closeDB func()
	if query.nodeIdx != 0 {
		__antithesis_instrumentation__.Notify(501237)
		addr := t.cluster.Server(query.nodeIdx).ServingSQLAddr()
		if len(t.tenantAddrs) > 0 {
			__antithesis_instrumentation__.Notify(501239)
			addr = t.tenantAddrs[query.nodeIdx]
		} else {
			__antithesis_instrumentation__.Notify(501240)
		}
		__antithesis_instrumentation__.Notify(501238)
		pgURL, cleanupFunc := sqlutils.PGUrl(t.rootT, addr, "TestLogic", url.User(t.user))
		defer cleanupFunc()
		pgURL.Path = "test"

		db = t.openDB(pgURL)
		closeDB = func() {
			__antithesis_instrumentation__.Notify(501241)
			if err := db.Close(); err != nil {
				__antithesis_instrumentation__.Notify(501242)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(501243)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(501244)
	}
	__antithesis_instrumentation__.Notify(501233)

	if query.expectAsync {
		__antithesis_instrumentation__.Notify(501245)
		if _, ok := t.pendingQueries[query.statementName]; ok {
			__antithesis_instrumentation__.Notify(501249)
			return errors.Newf("pending query with name %q already exists", query.statementName)
		} else {
			__antithesis_instrumentation__.Notify(501250)
		}
		__antithesis_instrumentation__.Notify(501246)

		pending := pendingQuery{
			logicQuery: query,
			resultChan: make(chan pendingQueryResult),
		}
		t.pendingQueries[query.statementName] = pending

		if *rewriteResultsInTestfiles || func() bool {
			__antithesis_instrumentation__.Notify(501251)
			return *rewriteSQL == true
		}() == true {
			__antithesis_instrumentation__.Notify(501252)
			t.emit(fmt.Sprintf("%s_%s", queryRewritePlaceholderPrefix, query.statementName))
		} else {
			__antithesis_instrumentation__.Notify(501253)
		}
		__antithesis_instrumentation__.Notify(501247)

		startedChan := make(chan struct{})
		go func() {
			__antithesis_instrumentation__.Notify(501254)
			if closeDB != nil {
				__antithesis_instrumentation__.Notify(501256)
				defer closeDB()
			} else {
				__antithesis_instrumentation__.Notify(501257)
			}
			__antithesis_instrumentation__.Notify(501255)
			startedChan <- struct{}{}
			rows, err := db.Query(query.sql)
			pending.resultChan <- pendingQueryResult{rows, err}
		}()
		__antithesis_instrumentation__.Notify(501248)

		<-startedChan
		return nil
	} else {
		__antithesis_instrumentation__.Notify(501258)
		if closeDB != nil {
			__antithesis_instrumentation__.Notify(501259)
			defer closeDB()
		} else {
			__antithesis_instrumentation__.Notify(501260)
		}
	}
	__antithesis_instrumentation__.Notify(501234)

	rows, err := db.Query(query.sql)
	return t.finishExecQuery(query, rows, err)
}

func (t *logicTest) finishExecQuery(query logicQuery, rows *gosql.Rows, err error) error {
	__antithesis_instrumentation__.Notify(501261)
	if err == nil {
		__antithesis_instrumentation__.Notify(501274)
		sqlutils.VerifyStatementPrettyRoundtrip(t.t(), query.sql)

		if query.expectErr != "" {
			__antithesis_instrumentation__.Notify(501275)

			for rows.Next() {
				__antithesis_instrumentation__.Notify(501277)
				if rows.Err() != nil {
					__antithesis_instrumentation__.Notify(501278)
					break
				} else {
					__antithesis_instrumentation__.Notify(501279)
				}
			}
			__antithesis_instrumentation__.Notify(501276)
			err = rows.Err()
		} else {
			__antithesis_instrumentation__.Notify(501280)
		}
	} else {
		__antithesis_instrumentation__.Notify(501281)
	}
	__antithesis_instrumentation__.Notify(501262)
	if _, err := t.verifyError(query.sql, query.pos, "", query.expectErr, query.expectErrCode, err); err != nil {
		__antithesis_instrumentation__.Notify(501282)
		return err
	} else {
		__antithesis_instrumentation__.Notify(501283)
	}
	__antithesis_instrumentation__.Notify(501263)
	if err != nil {
		__antithesis_instrumentation__.Notify(501284)

		t.finishOne("XFAIL")

		return nil
	} else {
		__antithesis_instrumentation__.Notify(501285)
	}
	__antithesis_instrumentation__.Notify(501264)
	defer rows.Close()

	var actualResultsRaw []string
	if query.noticetrace {
		__antithesis_instrumentation__.Notify(501286)

		if err := rows.Err(); err != nil {
			__antithesis_instrumentation__.Notify(501288)
			return err
		} else {
			__antithesis_instrumentation__.Notify(501289)
		}
		__antithesis_instrumentation__.Notify(501287)
		rows.Close()
		actualResultsRaw = t.noticeBuffer
	} else {
		__antithesis_instrumentation__.Notify(501290)
		cols, err := rows.Columns()
		if err != nil {
			__antithesis_instrumentation__.Notify(501296)
			return err
		} else {
			__antithesis_instrumentation__.Notify(501297)
		}
		__antithesis_instrumentation__.Notify(501291)
		if len(query.colTypes) != len(cols) {
			__antithesis_instrumentation__.Notify(501298)
			return fmt.Errorf("%s: expected %d columns, but found %d",
				query.pos, len(query.colTypes), len(cols))
		} else {
			__antithesis_instrumentation__.Notify(501299)
		}
		__antithesis_instrumentation__.Notify(501292)
		vals := make([]interface{}, len(cols))
		for i := range vals {
			__antithesis_instrumentation__.Notify(501300)
			vals[i] = new(interface{})
		}
		__antithesis_instrumentation__.Notify(501293)

		if query.colNames {
			__antithesis_instrumentation__.Notify(501301)
			actualResultsRaw = append(actualResultsRaw, cols...)
		} else {
			__antithesis_instrumentation__.Notify(501302)
		}
		__antithesis_instrumentation__.Notify(501294)
		for nextResultSet := true; nextResultSet; nextResultSet = rows.NextResultSet() {
			__antithesis_instrumentation__.Notify(501303)
			for rows.Next() {
				__antithesis_instrumentation__.Notify(501305)
				if err := rows.Scan(vals...); err != nil {
					__antithesis_instrumentation__.Notify(501307)
					return err
				} else {
					__antithesis_instrumentation__.Notify(501308)
				}
				__antithesis_instrumentation__.Notify(501306)
				for i, v := range vals {
					__antithesis_instrumentation__.Notify(501309)
					if val := *v.(*interface{}); val != nil {
						__antithesis_instrumentation__.Notify(501310)
						valT := reflect.TypeOf(val).Kind()
						colT := query.colTypes[i]
						switch colT {
						case 'T':
							__antithesis_instrumentation__.Notify(501315)
							if valT != reflect.String && func() bool {
								__antithesis_instrumentation__.Notify(501321)
								return valT != reflect.Slice == true
							}() == true && func() bool {
								__antithesis_instrumentation__.Notify(501322)
								return valT != reflect.Struct == true
							}() == true {
								__antithesis_instrumentation__.Notify(501323)
								return fmt.Errorf("%s: expected text value for column %d, but found %T: %#v",
									query.pos, i, val, val,
								)
							} else {
								__antithesis_instrumentation__.Notify(501324)
							}
						case 'I':
							__antithesis_instrumentation__.Notify(501316)
							if valT != reflect.Int64 {
								__antithesis_instrumentation__.Notify(501325)
								if *flexTypes && func() bool {
									__antithesis_instrumentation__.Notify(501327)
									return (valT == reflect.Float64 || func() bool {
										__antithesis_instrumentation__.Notify(501328)
										return valT == reflect.Slice == true
									}() == true) == true
								}() == true {
									__antithesis_instrumentation__.Notify(501329)
									t.signalIgnoredError(
										fmt.Errorf("result type mismatch: expected I, got %T", val), query.pos, query.sql,
									)
									return nil
								} else {
									__antithesis_instrumentation__.Notify(501330)
								}
								__antithesis_instrumentation__.Notify(501326)
								return fmt.Errorf("%s: expected int value for column %d, but found %T: %#v",
									query.pos, i, val, val,
								)
							} else {
								__antithesis_instrumentation__.Notify(501331)
							}
						case 'F', 'R':
							__antithesis_instrumentation__.Notify(501317)
							if valT != reflect.Float64 && func() bool {
								__antithesis_instrumentation__.Notify(501332)
								return valT != reflect.Slice == true
							}() == true {
								__antithesis_instrumentation__.Notify(501333)
								if *flexTypes && func() bool {
									__antithesis_instrumentation__.Notify(501335)
									return (valT == reflect.Int64) == true
								}() == true {
									__antithesis_instrumentation__.Notify(501336)
									t.signalIgnoredError(
										fmt.Errorf("result type mismatch: expected F or R, got %T", val), query.pos, query.sql,
									)
									return nil
								} else {
									__antithesis_instrumentation__.Notify(501337)
								}
								__antithesis_instrumentation__.Notify(501334)
								return fmt.Errorf("%s: expected float/decimal value for column %d, but found %T: %#v",
									query.pos, i, val, val,
								)
							} else {
								__antithesis_instrumentation__.Notify(501338)
							}
						case 'B':
							__antithesis_instrumentation__.Notify(501318)
							if valT != reflect.Bool {
								__antithesis_instrumentation__.Notify(501339)
								return fmt.Errorf("%s: expected boolean value for column %d, but found %T: %#v",
									query.pos, i, val, val,
								)
							} else {
								__antithesis_instrumentation__.Notify(501340)
							}
						case 'O':
							__antithesis_instrumentation__.Notify(501319)
							if valT != reflect.Slice {
								__antithesis_instrumentation__.Notify(501341)
								return fmt.Errorf("%s: expected oid value for column %d, but found %T: %#v",
									query.pos, i, val, val,
								)
							} else {
								__antithesis_instrumentation__.Notify(501342)
							}
						default:
							__antithesis_instrumentation__.Notify(501320)
							return fmt.Errorf("%s: unknown type in type string: %c in %s",
								query.pos, colT, query.colTypes,
							)
						}
						__antithesis_instrumentation__.Notify(501311)

						if byteArray, ok := val.([]byte); ok {
							__antithesis_instrumentation__.Notify(501343)

							if str := string(byteArray); utf8.ValidString(str) {
								__antithesis_instrumentation__.Notify(501344)
								val = str
							} else {
								__antithesis_instrumentation__.Notify(501345)
							}
						} else {
							__antithesis_instrumentation__.Notify(501346)
						}
						__antithesis_instrumentation__.Notify(501312)

						if val == "" {
							__antithesis_instrumentation__.Notify(501347)
							val = "·"
						} else {
							__antithesis_instrumentation__.Notify(501348)
						}
						__antithesis_instrumentation__.Notify(501313)
						s := fmt.Sprint(val)
						if query.roundFloatsInStrings {
							__antithesis_instrumentation__.Notify(501349)
							s = roundFloatsInString(s)
						} else {
							__antithesis_instrumentation__.Notify(501350)
						}
						__antithesis_instrumentation__.Notify(501314)
						actualResultsRaw = append(actualResultsRaw, s)
					} else {
						__antithesis_instrumentation__.Notify(501351)
						actualResultsRaw = append(actualResultsRaw, "NULL")
					}
				}
			}
			__antithesis_instrumentation__.Notify(501304)
			if err := rows.Err(); err != nil {
				__antithesis_instrumentation__.Notify(501352)
				return err
			} else {
				__antithesis_instrumentation__.Notify(501353)
			}
		}
		__antithesis_instrumentation__.Notify(501295)
		if err := rows.Err(); err != nil {
			__antithesis_instrumentation__.Notify(501354)
			return err
		} else {
			__antithesis_instrumentation__.Notify(501355)
		}
	}
	__antithesis_instrumentation__.Notify(501265)

	var actualResults []string
	if actualResultsRaw != nil {
		__antithesis_instrumentation__.Notify(501356)
		actualResults = make([]string, 0, len(actualResultsRaw))
		for _, result := range actualResultsRaw {
			__antithesis_instrumentation__.Notify(501357)
			if query.sorter == nil || func() bool {
				__antithesis_instrumentation__.Notify(501358)
				return query.valsPerLine != 1 == true
			}() == true {
				__antithesis_instrumentation__.Notify(501359)
				actualResults = append(actualResults, strings.Fields(result)...)
			} else {
				__antithesis_instrumentation__.Notify(501360)
				actualResults = append(actualResults, strings.Join(strings.Fields(result), " "))
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(501361)
	}
	__antithesis_instrumentation__.Notify(501266)

	if query.sorter != nil {
		__antithesis_instrumentation__.Notify(501362)
		query.sorter(len(query.colTypes), actualResults)
		query.sorter(len(query.colTypes), query.expectedResults)
	} else {
		__antithesis_instrumentation__.Notify(501363)
	}
	__antithesis_instrumentation__.Notify(501267)

	hash, err := t.hashResults(actualResults)
	if err != nil {
		__antithesis_instrumentation__.Notify(501364)
		return err
	} else {
		__antithesis_instrumentation__.Notify(501365)
	}
	__antithesis_instrumentation__.Notify(501268)

	if query.expectedHash != "" {
		__antithesis_instrumentation__.Notify(501366)
		n := len(actualResults)
		if query.expectedValues != n {
			__antithesis_instrumentation__.Notify(501368)
			return fmt.Errorf("%s: expected %d results, but found %d", query.pos, query.expectedValues, n)
		} else {
			__antithesis_instrumentation__.Notify(501369)
		}
		__antithesis_instrumentation__.Notify(501367)
		if query.expectedHash != hash {
			__antithesis_instrumentation__.Notify(501370)
			var suffix string
			for _, colT := range query.colTypes {
				__antithesis_instrumentation__.Notify(501372)
				if colT == 'F' {
					__antithesis_instrumentation__.Notify(501373)
					suffix = "\tthis might be due to floating numbers precision deviation"
					break
				} else {
					__antithesis_instrumentation__.Notify(501374)
				}
			}
			__antithesis_instrumentation__.Notify(501371)
			return fmt.Errorf("%s: expected %s, but found %s%s", query.pos, query.expectedHash, hash, suffix)
		} else {
			__antithesis_instrumentation__.Notify(501375)
		}
	} else {
		__antithesis_instrumentation__.Notify(501376)
	}
	__antithesis_instrumentation__.Notify(501269)

	resultsMatch := func() error {
		__antithesis_instrumentation__.Notify(501377)
		makeError := func() error {
			__antithesis_instrumentation__.Notify(501381)
			var expFormatted strings.Builder
			var actFormatted strings.Builder
			for _, line := range query.expectedResultsRaw {
				__antithesis_instrumentation__.Notify(501386)
				fmt.Fprintf(&expFormatted, "    %s\n", line)
			}
			__antithesis_instrumentation__.Notify(501382)
			for _, line := range t.formatValues(actualResultsRaw, query.valsPerLine) {
				__antithesis_instrumentation__.Notify(501387)
				fmt.Fprintf(&actFormatted, "    %s\n", line)
			}
			__antithesis_instrumentation__.Notify(501383)

			sortMsg := ""
			if query.sorter != nil {
				__antithesis_instrumentation__.Notify(501388)

				sortMsg = " -> ignore the following ordering of rows"
			} else {
				__antithesis_instrumentation__.Notify(501389)
			}
			__antithesis_instrumentation__.Notify(501384)
			var buf bytes.Buffer
			fmt.Fprintf(&buf, "%s: %s\nexpected:\n%s", query.pos, query.sql, expFormatted.String())
			fmt.Fprintf(&buf, "but found (query options: %q%s) :\n%s", query.rawOpts, sortMsg, actFormatted.String())
			if *showDiff {
				__antithesis_instrumentation__.Notify(501390)
				if diff, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
					A:        difflib.SplitLines(expFormatted.String()),
					B:        difflib.SplitLines(actFormatted.String()),
					FromFile: "Expected",
					FromDate: "",
					ToFile:   "Actual",
					ToDate:   "",
					Context:  1,
				}); err == nil {
					__antithesis_instrumentation__.Notify(501391)
					fmt.Fprintf(&buf, "\nDiff:\n%s", diff)
				} else {
					__antithesis_instrumentation__.Notify(501392)
				}
			} else {
				__antithesis_instrumentation__.Notify(501393)
			}
			__antithesis_instrumentation__.Notify(501385)
			return errors.Newf("%s\n", buf.String())
		}
		__antithesis_instrumentation__.Notify(501378)
		if len(query.expectedResults) != len(actualResults) {
			__antithesis_instrumentation__.Notify(501394)
			return makeError()
		} else {
			__antithesis_instrumentation__.Notify(501395)
		}
		__antithesis_instrumentation__.Notify(501379)
		for i := range query.expectedResults {
			__antithesis_instrumentation__.Notify(501396)
			expected, actual := query.expectedResults[i], actualResults[i]
			resultMatches := expected == actual

			colT := query.colTypes[i%len(query.colTypes)]
			if !resultMatches {
				__antithesis_instrumentation__.Notify(501398)
				var err error

				if runtime.GOARCH == "s390x" && func() bool {
					__antithesis_instrumentation__.Notify(501400)
					return (colT == 'F' || func() bool {
						__antithesis_instrumentation__.Notify(501401)
						return colT == 'R' == true
					}() == true) == true
				}() == true {
					__antithesis_instrumentation__.Notify(501402)
					resultMatches, err = floatsMatchApprox(expected, actual)
				} else {
					__antithesis_instrumentation__.Notify(501403)
					if colT == 'F' {
						__antithesis_instrumentation__.Notify(501404)
						resultMatches, err = floatsMatch(expected, actual)
					} else {
						__antithesis_instrumentation__.Notify(501405)
					}
				}
				__antithesis_instrumentation__.Notify(501399)
				if err != nil {
					__antithesis_instrumentation__.Notify(501406)
					return errors.CombineErrors(makeError(), err)
				} else {
					__antithesis_instrumentation__.Notify(501407)
				}
			} else {
				__antithesis_instrumentation__.Notify(501408)
			}
			__antithesis_instrumentation__.Notify(501397)
			if !resultMatches {
				__antithesis_instrumentation__.Notify(501409)
				return makeError()
			} else {
				__antithesis_instrumentation__.Notify(501410)
			}
		}
		__antithesis_instrumentation__.Notify(501380)
		return nil
	}
	__antithesis_instrumentation__.Notify(501270)

	if *rewriteResultsInTestfiles || func() bool {
		__antithesis_instrumentation__.Notify(501411)
		return *rewriteSQL == true
	}() == true {
		__antithesis_instrumentation__.Notify(501412)
		var remainder *bufio.Scanner
		if query.expectAsync {
			__antithesis_instrumentation__.Notify(501417)
			remainder = t.rewriteUpToRegex(regexp.MustCompile(fmt.Sprintf("^%s_%s$", queryRewritePlaceholderPrefix, query.statementName)))
		} else {
			__antithesis_instrumentation__.Notify(501418)
		}
		__antithesis_instrumentation__.Notify(501413)
		if query.expectedHash != "" {
			__antithesis_instrumentation__.Notify(501419)
			if query.expectedValues == 1 {
				__antithesis_instrumentation__.Notify(501420)
				t.emit(fmt.Sprintf("1 value hashing to %s", query.expectedHash))
			} else {
				__antithesis_instrumentation__.Notify(501421)
				t.emit(fmt.Sprintf("%d values hashing to %s", query.expectedValues, query.expectedHash))
			}
		} else {
			__antithesis_instrumentation__.Notify(501422)
		}
		__antithesis_instrumentation__.Notify(501414)

		if query.checkResults {
			__antithesis_instrumentation__.Notify(501423)

			if !*rewriteResultsInTestfiles || func() bool {
				__antithesis_instrumentation__.Notify(501424)
				return resultsMatch() == nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(501425)
				for _, l := range query.expectedResultsRaw {
					__antithesis_instrumentation__.Notify(501426)
					t.emit(l)
				}
			} else {
				__antithesis_instrumentation__.Notify(501427)

				for _, line := range t.formatValues(actualResultsRaw, query.valsPerLine) {
					__antithesis_instrumentation__.Notify(501428)
					t.emit(line)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(501429)
		}
		__antithesis_instrumentation__.Notify(501415)
		if remainder != nil {
			__antithesis_instrumentation__.Notify(501430)
			for remainder.Scan() {
				__antithesis_instrumentation__.Notify(501431)
				t.emit(remainder.Text())
			}
		} else {
			__antithesis_instrumentation__.Notify(501432)
		}
		__antithesis_instrumentation__.Notify(501416)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(501433)
	}
	__antithesis_instrumentation__.Notify(501271)

	if query.checkResults {
		__antithesis_instrumentation__.Notify(501434)
		if err := resultsMatch(); err != nil {
			__antithesis_instrumentation__.Notify(501435)
			return err
		} else {
			__antithesis_instrumentation__.Notify(501436)
		}
	} else {
		__antithesis_instrumentation__.Notify(501437)
	}
	__antithesis_instrumentation__.Notify(501272)

	if query.label != "" {
		__antithesis_instrumentation__.Notify(501438)
		if prevHash, ok := t.labelMap[query.label]; ok && func() bool {
			__antithesis_instrumentation__.Notify(501440)
			return prevHash != hash == true
		}() == true {
			__antithesis_instrumentation__.Notify(501441)
			t.Errorf(
				"%s: error in input: previous values for label %s (hash %s) do not match (hash %s)",
				query.pos, query.label, prevHash, hash,
			)
		} else {
			__antithesis_instrumentation__.Notify(501442)
		}
		__antithesis_instrumentation__.Notify(501439)
		t.labelMap[query.label] = hash
	} else {
		__antithesis_instrumentation__.Notify(501443)
	}
	__antithesis_instrumentation__.Notify(501273)

	t.finishOne("OK")
	return nil
}

func parseExpectedAndActualFloats(expectedString, actualString string) (float64, float64, error) {
	__antithesis_instrumentation__.Notify(501444)
	expected, err := strconv.ParseFloat(expectedString, 64)
	if err != nil {
		__antithesis_instrumentation__.Notify(501447)
		return 0, 0, errors.Wrap(err, "when parsing expected")
	} else {
		__antithesis_instrumentation__.Notify(501448)
	}
	__antithesis_instrumentation__.Notify(501445)
	actual, err := strconv.ParseFloat(actualString, 64)
	if err != nil {
		__antithesis_instrumentation__.Notify(501449)
		return 0, 0, errors.Wrap(err, "when parsing actual")
	} else {
		__antithesis_instrumentation__.Notify(501450)
	}
	__antithesis_instrumentation__.Notify(501446)
	return expected, actual, nil
}

func floatsMatchApprox(expectedString, actualString string) (bool, error) {
	__antithesis_instrumentation__.Notify(501451)
	expected, actual, err := parseExpectedAndActualFloats(expectedString, actualString)
	if err != nil {
		__antithesis_instrumentation__.Notify(501453)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(501454)
	}
	__antithesis_instrumentation__.Notify(501452)
	return floatcmp.EqualApprox(expected, actual, floatcmp.CloseFraction, floatcmp.CloseMargin), nil
}

func floatsMatch(expectedString, actualString string) (bool, error) {
	__antithesis_instrumentation__.Notify(501455)
	expected, actual, err := parseExpectedAndActualFloats(expectedString, actualString)
	if err != nil {
		__antithesis_instrumentation__.Notify(501464)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(501465)
	}
	__antithesis_instrumentation__.Notify(501456)

	if math.IsNaN(expected) || func() bool {
		__antithesis_instrumentation__.Notify(501466)
		return math.IsNaN(actual) == true
	}() == true {
		__antithesis_instrumentation__.Notify(501467)
		return math.IsNaN(expected) == math.IsNaN(actual), nil
	} else {
		__antithesis_instrumentation__.Notify(501468)
	}
	__antithesis_instrumentation__.Notify(501457)
	if math.IsInf(expected, 0) || func() bool {
		__antithesis_instrumentation__.Notify(501469)
		return math.IsInf(actual, 0) == true
	}() == true {
		__antithesis_instrumentation__.Notify(501470)
		bothNegativeInf := math.IsInf(expected, -1) == math.IsInf(actual, -1)
		bothPositiveInf := math.IsInf(expected, 1) == math.IsInf(actual, 1)
		return bothNegativeInf || func() bool {
			__antithesis_instrumentation__.Notify(501471)
			return bothPositiveInf == true
		}() == true, nil
	} else {
		__antithesis_instrumentation__.Notify(501472)
	}
	__antithesis_instrumentation__.Notify(501458)
	if expected == 0 || func() bool {
		__antithesis_instrumentation__.Notify(501473)
		return actual == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(501474)
		return expected == actual, nil
	} else {
		__antithesis_instrumentation__.Notify(501475)
	}
	__antithesis_instrumentation__.Notify(501459)

	if expected*actual < 0 {
		__antithesis_instrumentation__.Notify(501476)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(501477)
	}
	__antithesis_instrumentation__.Notify(501460)
	expected = math.Abs(expected)
	actual = math.Abs(actual)

	normalize := func(f float64) (base float64, power int) {
		__antithesis_instrumentation__.Notify(501478)
		for f >= 10 {
			__antithesis_instrumentation__.Notify(501481)
			f = f / 10
			power++
		}
		__antithesis_instrumentation__.Notify(501479)
		for f < 1 {
			__antithesis_instrumentation__.Notify(501482)
			f *= 10
			power--
		}
		__antithesis_instrumentation__.Notify(501480)
		return f, power
	}
	__antithesis_instrumentation__.Notify(501461)
	var expPower, actPower int
	expected, expPower = normalize(expected)
	actual, actPower = normalize(actual)
	if expPower != actPower {
		__antithesis_instrumentation__.Notify(501483)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(501484)
	}
	__antithesis_instrumentation__.Notify(501462)

	for i := 0; i < 14; i++ {
		__antithesis_instrumentation__.Notify(501485)
		expDigit := int(expected)
		actDigit := int(actual)
		if expDigit != actDigit {
			__antithesis_instrumentation__.Notify(501487)
			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(501488)
		}
		__antithesis_instrumentation__.Notify(501486)
		expected -= (expected - float64(expDigit)) * 10
		actual -= (actual - float64(actDigit)) * 10
	}
	__antithesis_instrumentation__.Notify(501463)
	return true, nil
}

func (t *logicTest) formatValues(vals []string, valsPerLine int) []string {
	__antithesis_instrumentation__.Notify(501489)
	var buf bytes.Buffer
	tw := tabwriter.NewWriter(&buf, 2, 1, 2, ' ', 0)

	for line := 0; line < len(vals)/valsPerLine; line++ {
		__antithesis_instrumentation__.Notify(501492)
		for i := 0; i < valsPerLine; i++ {
			__antithesis_instrumentation__.Notify(501494)
			fmt.Fprintf(tw, "%s\t", vals[line*valsPerLine+i])
		}
		__antithesis_instrumentation__.Notify(501493)
		fmt.Fprint(tw, "\n")
	}
	__antithesis_instrumentation__.Notify(501490)
	_ = tw.Flush()

	results := make([]string, 0, len(vals)/valsPerLine)
	for _, s := range strings.Split(buf.String(), "\n") {
		__antithesis_instrumentation__.Notify(501495)
		results = append(results, strings.TrimRight(s, " "))
	}
	__antithesis_instrumentation__.Notify(501491)
	return results
}

func (t *logicTest) success(file string) {
	__antithesis_instrumentation__.Notify(501496)
	t.progress++
	now := timeutil.Now()
	if now.Sub(t.lastProgress) >= 2*time.Second {
		__antithesis_instrumentation__.Notify(501497)
		t.lastProgress = now
		t.outf("--- progress: %s: %d statements", file, t.progress)
	} else {
		__antithesis_instrumentation__.Notify(501498)
	}
}

func (t *logicTest) validateAfterTestCompletion() error {
	__antithesis_instrumentation__.Notify(501499)

	if len(t.pendingStatements) > 0 || func() bool {
		__antithesis_instrumentation__.Notify(501510)
		return len(t.pendingQueries) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(501511)
		t.Fatalf("%d remaining async statements, %d remaining async queries", len(t.pendingStatements), len(t.pendingQueries))
	} else {
		__antithesis_instrumentation__.Notify(501512)
	}
	__antithesis_instrumentation__.Notify(501500)

	for username, c := range t.clients {
		__antithesis_instrumentation__.Notify(501513)
		if username == "root" {
			__antithesis_instrumentation__.Notify(501515)
			continue
		} else {
			__antithesis_instrumentation__.Notify(501516)
		}
		__antithesis_instrumentation__.Notify(501514)
		delete(t.clients, username)
		if err := c.Close(); err != nil {
			__antithesis_instrumentation__.Notify(501517)
			t.Fatalf("failed to close connection for user %s: %v", username, err)
		} else {
			__antithesis_instrumentation__.Notify(501518)
		}
	}
	__antithesis_instrumentation__.Notify(501501)
	t.setUser("root", 0)

	_, _ = t.db.Exec("ROLLBACK")
	_, err := t.db.Exec("RESET vectorize")
	if err != nil {
		__antithesis_instrumentation__.Notify(501519)
		t.Fatal(errors.Wrap(err, "could not reset vectorize mode"))
	} else {
		__antithesis_instrumentation__.Notify(501520)
	}
	__antithesis_instrumentation__.Notify(501502)

	validate := func() (string, error) {
		__antithesis_instrumentation__.Notify(501521)
		rows, err := t.db.Query(`SELECT * FROM "".crdb_internal.invalid_objects ORDER BY id`)
		if err != nil {
			__antithesis_instrumentation__.Notify(501525)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(501526)
		}
		__antithesis_instrumentation__.Notify(501522)
		defer rows.Close()

		var id int64
		var db, schema, objName, errStr string
		invalidObjects := make([]string, 0)
		for rows.Next() {
			__antithesis_instrumentation__.Notify(501527)
			if err := rows.Scan(&id, &db, &schema, &objName, &errStr); err != nil {
				__antithesis_instrumentation__.Notify(501529)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(501530)
			}
			__antithesis_instrumentation__.Notify(501528)
			invalidObjects = append(
				invalidObjects,
				fmt.Sprintf("id %d, db %s, schema %s, name %s: %s", id, db, schema, objName, errStr),
			)
		}
		__antithesis_instrumentation__.Notify(501523)
		if err := rows.Err(); err != nil {
			__antithesis_instrumentation__.Notify(501531)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(501532)
		}
		__antithesis_instrumentation__.Notify(501524)
		return strings.Join(invalidObjects, "\n"), nil
	}
	__antithesis_instrumentation__.Notify(501503)

	invalidObjects, err := validate()
	if err != nil {
		__antithesis_instrumentation__.Notify(501533)
		return errors.Wrap(err, "running object validation failed")
	} else {
		__antithesis_instrumentation__.Notify(501534)
	}
	__antithesis_instrumentation__.Notify(501504)
	if invalidObjects != "" {
		__antithesis_instrumentation__.Notify(501535)
		return errors.Errorf("descriptor validation failed:\n%s", invalidObjects)
	} else {
		__antithesis_instrumentation__.Notify(501536)
	}

	{
		__antithesis_instrumentation__.Notify(501537)
		rows, err := t.db.Query(
			`
SELECT encode(descriptor, 'hex') AS descriptor
  FROM system.descriptor
 WHERE descriptor
       != crdb_internal.json_to_pb(
            'cockroach.sql.sqlbase.Descriptor',
            crdb_internal.pb_to_json(
                'cockroach.sql.sqlbase.Descriptor',
                descriptor,
                false -- emit_defaults
            )
        );
`,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(501540)
			return errors.Wrap(err, "failed to test for descriptor JSON round-trip")
		} else {
			__antithesis_instrumentation__.Notify(501541)
		}
		__antithesis_instrumentation__.Notify(501538)
		rowsMat, err := sqlutils.RowsToStrMatrix(rows)
		if err != nil {
			__antithesis_instrumentation__.Notify(501542)
			return errors.Wrap(err, "failed read rows from descriptor JSON round-trip")
		} else {
			__antithesis_instrumentation__.Notify(501543)
		}
		__antithesis_instrumentation__.Notify(501539)
		if len(rowsMat) > 0 {
			__antithesis_instrumentation__.Notify(501544)
			return errors.Errorf("some descriptors did not round-trip:\n%s",
				sqlutils.MatrixToStr(rowsMat))
		} else {
			__antithesis_instrumentation__.Notify(501545)
		}
	}
	__antithesis_instrumentation__.Notify(501505)

	var dbNames pq.StringArray
	if err := t.db.QueryRow(
		`SELECT array_agg(database_name) FROM [SHOW DATABASES] WHERE database_name NOT IN ('system', 'postgres')`,
	).Scan(&dbNames); err != nil {
		__antithesis_instrumentation__.Notify(501546)
		return errors.Wrap(err, "error getting database names")
	} else {
		__antithesis_instrumentation__.Notify(501547)
	}
	__antithesis_instrumentation__.Notify(501506)

	for _, dbName := range dbNames {
		__antithesis_instrumentation__.Notify(501548)
		if err := func() (retErr error) {
			__antithesis_instrumentation__.Notify(501549)
			ctx := context.Background()

			conn, err := t.db.Conn(ctx)
			if err != nil {
				__antithesis_instrumentation__.Notify(501555)
				return errors.Wrap(err, "error grabbing new connection")
			} else {
				__antithesis_instrumentation__.Notify(501556)
			}
			__antithesis_instrumentation__.Notify(501550)
			defer func() {
				__antithesis_instrumentation__.Notify(501557)
				retErr = errors.CombineErrors(retErr, conn.Close())
			}()
			__antithesis_instrumentation__.Notify(501551)
			if _, err := conn.ExecContext(ctx, "SET database = $1", dbName); err != nil {
				__antithesis_instrumentation__.Notify(501558)
				return errors.Wrapf(err, "error validating zone config for database %s", dbName)
			} else {
				__antithesis_instrumentation__.Notify(501559)
			}
			__antithesis_instrumentation__.Notify(501552)

			if _, err := conn.ExecContext(ctx, "SELECT crdb_internal.validate_multi_region_zone_configs()"); err != nil {
				__antithesis_instrumentation__.Notify(501560)
				return errors.Wrapf(err, "error validating zone config for database %s", dbName)
			} else {
				__antithesis_instrumentation__.Notify(501561)
			}
			__antithesis_instrumentation__.Notify(501553)

			dbTreeName := tree.Name(dbName)
			dropDatabaseStmt := fmt.Sprintf(
				"DROP DATABASE %s CASCADE",
				dbTreeName.String(),
			)
			if _, err := t.db.Exec(dropDatabaseStmt); err != nil {
				__antithesis_instrumentation__.Notify(501562)
				return errors.Wrapf(err, "dropping database %s failed", dbName)
			} else {
				__antithesis_instrumentation__.Notify(501563)
			}
			__antithesis_instrumentation__.Notify(501554)
			return nil
		}(); err != nil {
			__antithesis_instrumentation__.Notify(501564)
			return err
		} else {
			__antithesis_instrumentation__.Notify(501565)
		}
	}
	__antithesis_instrumentation__.Notify(501507)

	invalidObjects, err = validate()
	if err != nil {
		__antithesis_instrumentation__.Notify(501566)
		return errors.Wrap(err, "running object validation after database drops failed")
	} else {
		__antithesis_instrumentation__.Notify(501567)
	}
	__antithesis_instrumentation__.Notify(501508)
	if invalidObjects != "" {
		__antithesis_instrumentation__.Notify(501568)
		return errors.Errorf(
			"descriptor validation failed after dropping databases:\n%s", invalidObjects,
		)
	} else {
		__antithesis_instrumentation__.Notify(501569)
	}
	__antithesis_instrumentation__.Notify(501509)

	return nil
}

func (t *logicTest) runFile(path string, config testClusterConfig) {
	__antithesis_instrumentation__.Notify(501570)
	defer t.close()

	defer func() {
		__antithesis_instrumentation__.Notify(501573)
		if r := recover(); r != nil {
			__antithesis_instrumentation__.Notify(501574)

			t.Fatalf("panic: %v\n%s", r, string(debug.Stack()))
		} else {
			__antithesis_instrumentation__.Notify(501575)
		}
	}()
	__antithesis_instrumentation__.Notify(501571)

	if err := t.processTestFile(path, config); err != nil {
		__antithesis_instrumentation__.Notify(501576)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(501577)
	}
	__antithesis_instrumentation__.Notify(501572)

	if err := t.validateAfterTestCompletion(); err != nil {
		__antithesis_instrumentation__.Notify(501578)
		t.Fatal(errors.Wrap(err, "test was successful but validation upon completion failed"))
	} else {
		__antithesis_instrumentation__.Notify(501579)
	}
}

var skipLogicTests = envutil.EnvOrDefaultBool("COCKROACH_LOGIC_TESTS_SKIP", false)
var logicTestsConfigExclude = envutil.EnvOrDefaultString("COCKROACH_LOGIC_TESTS_SKIP_CONFIG", "")
var logicTestsConfigFilter = envutil.EnvOrDefaultString("COCKROACH_LOGIC_TESTS_CONFIG", "")

type TestServerArgs struct {
	maxSQLMemoryLimit int64

	tempStorageDiskLimit int64

	forceProductionBatchSizes bool
}

func RunLogicTest(t *testing.T, serverArgs TestServerArgs, globs ...string) {
	__antithesis_instrumentation__.Notify(501580)
	RunLogicTestWithDefaultConfig(t, serverArgs, *overrideConfig, false, globs...)
}

func RunLogicTestWithDefaultConfig(
	t *testing.T,
	serverArgs TestServerArgs,
	configOverride string,
	runCCLConfigs bool,
	globs ...string,
) {
	__antithesis_instrumentation__.Notify(501581)

	skip.UnderStressRace(t, "logic tests and race detector don't mix: #37993")

	if skipLogicTests {
		__antithesis_instrumentation__.Notify(501590)
		skip.IgnoreLint(t, "COCKROACH_LOGIC_TESTS_SKIP")
	} else {
		__antithesis_instrumentation__.Notify(501591)
	}
	__antithesis_instrumentation__.Notify(501582)

	if *logictestdata != "" {
		__antithesis_instrumentation__.Notify(501592)
		globs = []string{*logictestdata}
	} else {
		__antithesis_instrumentation__.Notify(501593)
	}
	__antithesis_instrumentation__.Notify(501583)

	var paths []string
	for _, g := range globs {
		__antithesis_instrumentation__.Notify(501594)
		match, err := filepath.Glob(g)
		if err != nil {
			__antithesis_instrumentation__.Notify(501596)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(501597)
		}
		__antithesis_instrumentation__.Notify(501595)
		paths = append(paths, match...)
	}
	__antithesis_instrumentation__.Notify(501584)

	if len(paths) == 0 {
		__antithesis_instrumentation__.Notify(501598)
		t.Fatalf("No testfiles found (globs: %v)", globs)
	} else {
		__antithesis_instrumentation__.Notify(501599)
	}
	__antithesis_instrumentation__.Notify(501585)

	var progress = struct {
		syncutil.Mutex
		total, totalFail, totalUnsupported int
		lastProgress                       time.Time
	}{
		lastProgress: timeutil.Now(),
	}

	configPaths := make([][]string, len(logicTestConfigs))

	nonMetamorphic := make([][]bool, len(logicTestConfigs))
	configDefaults := defaultConfig
	var configFilter map[string]struct{}
	if configOverride != "" {
		__antithesis_instrumentation__.Notify(501600)

		names := strings.Split(configOverride, ",")
		configDefaults = parseTestConfig(names)
		configFilter = make(map[string]struct{})
		for _, name := range names {
			__antithesis_instrumentation__.Notify(501601)
			configFilter[name] = struct{}{}
		}
	} else {
		__antithesis_instrumentation__.Notify(501602)
	}
	__antithesis_instrumentation__.Notify(501586)
	for _, path := range paths {
		__antithesis_instrumentation__.Notify(501603)
		configs, onlyNonMetamorphic := readTestFileConfigs(t, path, configDefaults)
		for _, idx := range configs {
			__antithesis_instrumentation__.Notify(501604)
			config := logicTestConfigs[idx]
			configName := config.name
			if _, ok := configFilter[configName]; configFilter != nil && func() bool {
				__antithesis_instrumentation__.Notify(501607)
				return !ok == true
			}() == true {
				__antithesis_instrumentation__.Notify(501608)

				continue
			} else {
				__antithesis_instrumentation__.Notify(501609)
			}
			__antithesis_instrumentation__.Notify(501605)
			if config.isCCLConfig && func() bool {
				__antithesis_instrumentation__.Notify(501610)
				return !runCCLConfigs == true
			}() == true {
				__antithesis_instrumentation__.Notify(501611)

				continue
			} else {
				__antithesis_instrumentation__.Notify(501612)
			}
			__antithesis_instrumentation__.Notify(501606)
			configPaths[idx] = append(configPaths[idx], path)
			nonMetamorphic[idx] = append(nonMetamorphic[idx], onlyNonMetamorphic)
		}
	}
	__antithesis_instrumentation__.Notify(501587)

	logScope := log.Scope(t)
	defer logScope.Close(t)

	verbose := testing.Verbose() || func() bool {
		__antithesis_instrumentation__.Notify(501613)
		return log.V(1) == true
	}() == true

	seenPaths := make(map[string]struct{})
	for idx, cfg := range logicTestConfigs {
		__antithesis_instrumentation__.Notify(501614)
		paths := configPaths[idx]
		nonMetamorphic := nonMetamorphic[idx]
		if len(paths) == 0 {
			__antithesis_instrumentation__.Notify(501616)
			continue
		} else {
			__antithesis_instrumentation__.Notify(501617)
		}
		__antithesis_instrumentation__.Notify(501615)

		t.Run(cfg.name, func(t *testing.T) {
			__antithesis_instrumentation__.Notify(501618)
			if testing.Short() && func() bool {
				__antithesis_instrumentation__.Notify(501622)
				return cfg.skipShort == true
			}() == true {
				__antithesis_instrumentation__.Notify(501623)
				skip.IgnoreLint(t, "config skipped by -test.short")
			} else {
				__antithesis_instrumentation__.Notify(501624)
			}
			__antithesis_instrumentation__.Notify(501619)
			if logicTestsConfigExclude != "" && func() bool {
				__antithesis_instrumentation__.Notify(501625)
				return cfg.name == logicTestsConfigExclude == true
			}() == true {
				__antithesis_instrumentation__.Notify(501626)
				skip.IgnoreLint(t, "config excluded via env var")
			} else {
				__antithesis_instrumentation__.Notify(501627)
			}
			__antithesis_instrumentation__.Notify(501620)
			if logicTestsConfigFilter != "" && func() bool {
				__antithesis_instrumentation__.Notify(501628)
				return cfg.name != logicTestsConfigFilter == true
			}() == true {
				__antithesis_instrumentation__.Notify(501629)
				skip.IgnoreLint(t, "config does not match env var")
			} else {
				__antithesis_instrumentation__.Notify(501630)
			}
			__antithesis_instrumentation__.Notify(501621)
			for i, path := range paths {
				__antithesis_instrumentation__.Notify(501631)
				path := path
				onlyNonMetamorphic := nonMetamorphic[i]

				t.Run(filepath.Base(path), func(t *testing.T) {
					__antithesis_instrumentation__.Notify(501632)
					if *rewriteResultsInTestfiles {
						__antithesis_instrumentation__.Notify(501636)
						if _, seen := seenPaths[path]; seen {
							__antithesis_instrumentation__.Notify(501638)
							skip.IgnoreLint(t, "test file already rewritten")
						} else {
							__antithesis_instrumentation__.Notify(501639)
						}
						__antithesis_instrumentation__.Notify(501637)
						seenPaths[path] = struct{}{}
					} else {
						__antithesis_instrumentation__.Notify(501640)
					}
					__antithesis_instrumentation__.Notify(501633)

					if !*showSQL && func() bool {
						__antithesis_instrumentation__.Notify(501641)
						return !*rewriteResultsInTestfiles == true
					}() == true && func() bool {
						__antithesis_instrumentation__.Notify(501642)
						return !*rewriteSQL == true
					}() == true && func() bool {
						__antithesis_instrumentation__.Notify(501643)
						return !util.RaceEnabled == true
					}() == true && func() bool {
						__antithesis_instrumentation__.Notify(501644)
						return !cfg.useTenant == true
					}() == true && func() bool {
						__antithesis_instrumentation__.Notify(501645)
						return cfg.numNodes <= 5 == true
					}() == true {
						__antithesis_instrumentation__.Notify(501646)

						if filepath.Base(path) != "select_index_span_ranges" {
							__antithesis_instrumentation__.Notify(501647)
							t.Parallel()
						} else {
							__antithesis_instrumentation__.Notify(501648)
						}
					} else {
						__antithesis_instrumentation__.Notify(501649)
					}
					__antithesis_instrumentation__.Notify(501634)
					rng, _ := randutil.NewTestRand()
					lt := logicTest{
						rootT:           t,
						verbose:         verbose,
						perErrorSummary: make(map[string][]string),
						rng:             rng,
					}
					if *printErrorSummary {
						__antithesis_instrumentation__.Notify(501650)
						defer lt.printErrorSummary()
					} else {
						__antithesis_instrumentation__.Notify(501651)
					}
					__antithesis_instrumentation__.Notify(501635)

					serverArgsCopy := serverArgs
					serverArgsCopy.forceProductionBatchSizes = onlyNonMetamorphic
					lt.setup(
						cfg, serverArgsCopy, readClusterOptions(t, path), readTenantClusterSettingOverrideArgs(t, path),
					)
					lt.runFile(path, cfg)

					progress.Lock()
					defer progress.Unlock()
					progress.total += lt.progress
					progress.totalFail += lt.failures
					progress.totalUnsupported += lt.unsupported
					now := timeutil.Now()
					if now.Sub(progress.lastProgress) >= 2*time.Second {
						__antithesis_instrumentation__.Notify(501652)
						progress.lastProgress = now
						lt.outf("--- total progress: %d statements", progress.total)
					} else {
						__antithesis_instrumentation__.Notify(501653)
					}
				})
			}
		})
	}
	__antithesis_instrumentation__.Notify(501588)

	unsupportedMsg := ""
	if progress.totalUnsupported > 0 {
		__antithesis_instrumentation__.Notify(501654)
		unsupportedMsg = fmt.Sprintf(", ignored %d unsupported queries", progress.totalUnsupported)
	} else {
		__antithesis_instrumentation__.Notify(501655)
	}
	__antithesis_instrumentation__.Notify(501589)

	if verbose {
		__antithesis_instrumentation__.Notify(501656)
		fmt.Printf("--- total: %d tests, %d failures%s\n",
			progress.total, progress.totalFail, unsupportedMsg,
		)
	} else {
		__antithesis_instrumentation__.Notify(501657)
	}
}

func RunSQLLiteLogicTest(t *testing.T, configOverride string) {
	__antithesis_instrumentation__.Notify(501658)
	runSQLLiteLogicTest(t,
		configOverride,
		"/test/index/between/*/*.test",
		"/test/index/commute/*/*.test",
		"/test/index/delete/*/*.test",
		"/test/index/in/*/*.test",
		"/test/index/orderby/*/*.test",
		"/test/index/orderby_nosort/*/*.test",
		"/test/index/view/*/*.test",

		"/test/select1.test",
		"/test/select2.test",
		"/test/select3.test",
		"/test/select4.test",
	)
}

func runSQLLiteLogicTest(t *testing.T, configOverride string, globs ...string) {
	__antithesis_instrumentation__.Notify(501659)
	if !*bigtest {
		__antithesis_instrumentation__.Notify(501663)
		skip.IgnoreLint(t, "-bigtest flag must be specified to run this test")
	} else {
		__antithesis_instrumentation__.Notify(501664)
	}
	__antithesis_instrumentation__.Notify(501660)

	var logicTestPath string
	if bazel.BuiltWithBazel() {
		__antithesis_instrumentation__.Notify(501665)
		runfilesPath, err := bazel.RunfilesPath()
		if err != nil {
			__antithesis_instrumentation__.Notify(501667)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(501668)
		}
		__antithesis_instrumentation__.Notify(501666)
		logicTestPath = filepath.Join(runfilesPath, "external", "com_github_cockroachdb_sqllogictest")
	} else {
		__antithesis_instrumentation__.Notify(501669)
		logicTestPath = gobuild.Default.GOPATH + "/src/github.com/cockroachdb/sqllogictest"
		if _, err := os.Stat(logicTestPath); oserror.IsNotExist(err) {
			__antithesis_instrumentation__.Notify(501670)
			fullPath, err := filepath.Abs(logicTestPath)
			if err != nil {
				__antithesis_instrumentation__.Notify(501672)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(501673)
			}
			__antithesis_instrumentation__.Notify(501671)
			t.Fatalf("unable to find sqllogictest repo: %s\n"+
				"git clone https://github.com/cockroachdb/sqllogictest %s",
				logicTestPath, fullPath)
			return
		} else {
			__antithesis_instrumentation__.Notify(501674)
		}
	}
	__antithesis_instrumentation__.Notify(501661)

	prefixedGlobs := make([]string, len(globs))
	for i, glob := range globs {
		__antithesis_instrumentation__.Notify(501675)
		prefixedGlobs[i] = logicTestPath + glob
	}
	__antithesis_instrumentation__.Notify(501662)

	serverArgs := TestServerArgs{
		maxSQLMemoryLimit:    512 << 20,
		tempStorageDiskLimit: 512 << 20,
	}
	RunLogicTestWithDefaultConfig(t, serverArgs, configOverride, true, prefixedGlobs...)
}

type errorSummaryEntry struct {
	errmsg string
	sql    []string
}
type errorSummary []errorSummaryEntry

func (e errorSummary) Len() int { __antithesis_instrumentation__.Notify(501676); return len(e) }
func (e errorSummary) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(501677)
	if len(e[i].sql) == len(e[j].sql) {
		__antithesis_instrumentation__.Notify(501679)
		return e[i].errmsg < e[j].errmsg
	} else {
		__antithesis_instrumentation__.Notify(501680)
	}
	__antithesis_instrumentation__.Notify(501678)
	return len(e[i].sql) < len(e[j].sql)
}
func (e errorSummary) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(501681)
	t := e[i]
	e[i] = e[j]
	e[j] = t
}

func (t *logicTest) printErrorSummary() {
	__antithesis_instrumentation__.Notify(501682)
	if t.unsupported == 0 {
		__antithesis_instrumentation__.Notify(501685)
		return
	} else {
		__antithesis_instrumentation__.Notify(501686)
	}
	__antithesis_instrumentation__.Notify(501683)

	t.outf("--- summary of ignored errors:")
	summary := make(errorSummary, len(t.perErrorSummary))
	i := 0
	for errmsg, sql := range t.perErrorSummary {
		__antithesis_instrumentation__.Notify(501687)
		summary[i] = errorSummaryEntry{errmsg: errmsg, sql: sql}
	}
	__antithesis_instrumentation__.Notify(501684)
	sort.Sort(summary)
	for _, s := range summary {
		__antithesis_instrumentation__.Notify(501688)
		t.outf("%s (%d entries)", s.errmsg, len(s.sql))
		var buf bytes.Buffer
		for _, q := range s.sql {
			__antithesis_instrumentation__.Notify(501690)
			buf.WriteByte('\t')
			buf.WriteString(strings.Replace(q, "\n", "\n\t", -1))
		}
		__antithesis_instrumentation__.Notify(501689)
		t.outf("%s", buf.String())
	}
}

func shortenString(msg string) string {
	__antithesis_instrumentation__.Notify(501691)
	if *fullMessages {
		__antithesis_instrumentation__.Notify(501696)
		return msg
	} else {
		__antithesis_instrumentation__.Notify(501697)
	}
	__antithesis_instrumentation__.Notify(501692)

	shortened := false

	nlPos := strings.IndexRune(msg, '\n')
	if nlPos >= 0 {
		__antithesis_instrumentation__.Notify(501698)
		shortened = true
		msg = msg[:nlPos]
	} else {
		__antithesis_instrumentation__.Notify(501699)
	}
	__antithesis_instrumentation__.Notify(501693)

	if len(msg) > 80 {
		__antithesis_instrumentation__.Notify(501700)
		shortened = true
		msg = msg[:80]
	} else {
		__antithesis_instrumentation__.Notify(501701)
	}
	__antithesis_instrumentation__.Notify(501694)

	if shortened {
		__antithesis_instrumentation__.Notify(501702)
		msg = msg + "..."
	} else {
		__antithesis_instrumentation__.Notify(501703)
	}
	__antithesis_instrumentation__.Notify(501695)

	return msg
}

func simplifyError(msg string) (string, string) {
	__antithesis_instrumentation__.Notify(501704)
	prefix := strings.Split(msg, ": ")

	expected := ""
	if len(prefix) > 1 {
		__antithesis_instrumentation__.Notify(501708)
		expected = prefix[len(prefix)-1]
		prefix = prefix[:len(prefix)-1]
	} else {
		__antithesis_instrumentation__.Notify(501709)
	}
	__antithesis_instrumentation__.Notify(501705)

	if !*fullMessages && func() bool {
		__antithesis_instrumentation__.Notify(501710)
		return len(prefix) > 2 == true
	}() == true {
		__antithesis_instrumentation__.Notify(501711)
		prefix[1] = prefix[len(prefix)-1]
		prefix = prefix[:2]
	} else {
		__antithesis_instrumentation__.Notify(501712)
	}
	__antithesis_instrumentation__.Notify(501706)

	if expected != "" {
		__antithesis_instrumentation__.Notify(501713)
		prefix = append(prefix, "...")
	} else {
		__antithesis_instrumentation__.Notify(501714)
	}
	__antithesis_instrumentation__.Notify(501707)

	return strings.Join(prefix, ": "), expected
}

func (t *logicTest) signalIgnoredError(err error, pos string, sql string) {
	__antithesis_instrumentation__.Notify(501715)
	t.unsupported++

	if !*printErrorSummary {
		__antithesis_instrumentation__.Notify(501717)
		return
	} else {
		__antithesis_instrumentation__.Notify(501718)
	}
	__antithesis_instrumentation__.Notify(501716)

	errmsg, expected := simplifyError(fmt.Sprintf("%s", err))
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "-- %s (%s)\n%s", pos, shortenString(expected), shortenString(sql+";"))
	errmsg = shortenString(errmsg)
	t.perErrorSummary[errmsg] = append(t.perErrorSummary[errmsg], buf.String())
}

func (t *logicTest) Error(args ...interface{}) {
	__antithesis_instrumentation__.Notify(501719)
	t.t().Helper()
	if *showSQL {
		__antithesis_instrumentation__.Notify(501721)
		t.outf("\t-- FAIL")
	} else {
		__antithesis_instrumentation__.Notify(501722)
	}
	__antithesis_instrumentation__.Notify(501720)
	log.Errorf(context.Background(), "\n%s", fmt.Sprint(args...))
	t.t().Error("\n", fmt.Sprint(args...))
	t.failures++
}

func (t *logicTest) Errorf(format string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(501723)
	t.t().Helper()
	if *showSQL {
		__antithesis_instrumentation__.Notify(501725)
		t.outf("\t-- FAIL")
	} else {
		__antithesis_instrumentation__.Notify(501726)
	}
	__antithesis_instrumentation__.Notify(501724)
	log.Errorf(context.Background(), format, args...)
	t.t().Errorf("\n"+format, args...)
	t.failures++
}

func (t *logicTest) Fatal(args ...interface{}) {
	__antithesis_instrumentation__.Notify(501727)
	t.t().Helper()
	if *showSQL {
		__antithesis_instrumentation__.Notify(501729)
		fmt.Println()
	} else {
		__antithesis_instrumentation__.Notify(501730)
	}
	__antithesis_instrumentation__.Notify(501728)
	log.Errorf(context.Background(), "%s", fmt.Sprint(args...))
	t.t().Logf("\n%s:%d: error while processing", t.curPath, t.curLineNo)
	t.t().Fatal(args...)
}

func (t *logicTest) Fatalf(format string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(501731)
	if *showSQL {
		__antithesis_instrumentation__.Notify(501733)
		fmt.Println()
	} else {
		__antithesis_instrumentation__.Notify(501734)
	}
	__antithesis_instrumentation__.Notify(501732)
	log.Errorf(context.Background(), format, args...)
	t.t().Logf("\n%s:%d: error while processing", t.curPath, t.curLineNo)
	t.t().Fatalf(format, args...)
}

func (t *logicTest) finishOne(msg string) {
	__antithesis_instrumentation__.Notify(501735)
	if *showSQL {
		__antithesis_instrumentation__.Notify(501736)
		t.outf("\t-- %s;", msg)
	} else {
		__antithesis_instrumentation__.Notify(501737)
	}
}

func (t *logicTest) printCompletion(path string, config testClusterConfig) {
	__antithesis_instrumentation__.Notify(501738)
	unsupportedMsg := ""
	if t.unsupported > 0 {
		__antithesis_instrumentation__.Notify(501740)
		unsupportedMsg = fmt.Sprintf(", ignored %d unsupported queries", t.unsupported)
	} else {
		__antithesis_instrumentation__.Notify(501741)
	}
	__antithesis_instrumentation__.Notify(501739)
	t.outf("--- done: %s with config %s: %d tests, %d failures%s", path, config.name,
		t.progress, t.failures, unsupportedMsg)
}

func roundFloatsInString(s string) string {
	__antithesis_instrumentation__.Notify(501742)
	return string(regexp.MustCompile(`(\d+\.\d+)`).ReplaceAllFunc([]byte(s), func(x []byte) []byte {
		__antithesis_instrumentation__.Notify(501743)
		f, err := strconv.ParseFloat(string(x), 64)
		if err != nil {
			__antithesis_instrumentation__.Notify(501745)
			return []byte(err.Error())
		} else {
			__antithesis_instrumentation__.Notify(501746)
		}
		__antithesis_instrumentation__.Notify(501744)
		return []byte(fmt.Sprintf("%.6g", f))
	}))
}
