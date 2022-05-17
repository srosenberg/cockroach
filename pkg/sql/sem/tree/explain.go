package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/pretty"
	"github.com/cockroachdb/errors"
)

type Explain struct {
	ExplainOptions

	Statement Statement
}

type ExplainAnalyze struct {
	ExplainOptions

	Statement Statement
}

type ExplainOptions struct {
	Mode  ExplainMode
	Flags [numExplainFlags + 1]bool
}

type ExplainMode uint8

const (
	ExplainPlan ExplainMode = 1 + iota

	ExplainDistSQL

	ExplainOpt

	ExplainVec

	ExplainDebug

	ExplainDDL

	ExplainGist

	numExplainModes = iota
)

var explainModeStrings = [...]string{
	ExplainPlan:    "PLAN",
	ExplainDistSQL: "DISTSQL",
	ExplainOpt:     "OPT",
	ExplainVec:     "VEC",
	ExplainDebug:   "DEBUG",
	ExplainDDL:     "DDL",
	ExplainGist:    "GIST",
}

var explainModeStringMap = func() map[string]ExplainMode {
	__antithesis_instrumentation__.Notify(609236)
	m := make(map[string]ExplainMode, numExplainModes)
	for i := ExplainMode(1); i <= numExplainModes; i++ {
		__antithesis_instrumentation__.Notify(609238)
		m[explainModeStrings[i]] = i
	}
	__antithesis_instrumentation__.Notify(609237)
	return m
}()

func (m ExplainMode) String() string {
	__antithesis_instrumentation__.Notify(609239)
	if m == 0 || func() bool {
		__antithesis_instrumentation__.Notify(609241)
		return m > numExplainModes == true
	}() == true {
		__antithesis_instrumentation__.Notify(609242)
		panic(errors.AssertionFailedf("invalid ExplainMode %d", m))
	} else {
		__antithesis_instrumentation__.Notify(609243)
	}
	__antithesis_instrumentation__.Notify(609240)
	return explainModeStrings[m]
}

type ExplainFlag uint8

const (
	ExplainFlagVerbose ExplainFlag = 1 + iota
	ExplainFlagTypes
	ExplainFlagEnv
	ExplainFlagCatalog
	ExplainFlagJSON
	ExplainFlagMemo
	ExplainFlagShape
	ExplainFlagViz
	numExplainFlags = iota
)

var explainFlagStrings = [...]string{
	ExplainFlagVerbose: "VERBOSE",
	ExplainFlagTypes:   "TYPES",
	ExplainFlagEnv:     "ENV",
	ExplainFlagCatalog: "CATALOG",
	ExplainFlagJSON:    "JSON",
	ExplainFlagMemo:    "MEMO",
	ExplainFlagShape:   "SHAPE",
	ExplainFlagViz:     "VIZ",
}

var explainFlagStringMap = func() map[string]ExplainFlag {
	__antithesis_instrumentation__.Notify(609244)
	m := make(map[string]ExplainFlag, numExplainFlags)
	for i := ExplainFlag(1); i <= numExplainFlags; i++ {
		__antithesis_instrumentation__.Notify(609246)
		m[explainFlagStrings[i]] = i
	}
	__antithesis_instrumentation__.Notify(609245)
	return m
}()

func (f ExplainFlag) String() string {
	__antithesis_instrumentation__.Notify(609247)
	if f == 0 || func() bool {
		__antithesis_instrumentation__.Notify(609249)
		return f > numExplainFlags == true
	}() == true {
		__antithesis_instrumentation__.Notify(609250)
		panic(errors.AssertionFailedf("invalid ExplainFlag %d", f))
	} else {
		__antithesis_instrumentation__.Notify(609251)
	}
	__antithesis_instrumentation__.Notify(609248)
	return explainFlagStrings[f]
}

func (node *Explain) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609252)
	ctx.WriteString("EXPLAIN ")
	b := util.MakeStringListBuilder("(", ", ", ") ")
	if node.Mode != ExplainPlan {
		__antithesis_instrumentation__.Notify(609255)
		b.Add(ctx, node.Mode.String())
	} else {
		__antithesis_instrumentation__.Notify(609256)
	}
	__antithesis_instrumentation__.Notify(609253)

	for f := ExplainFlag(1); f <= numExplainFlags; f++ {
		__antithesis_instrumentation__.Notify(609257)
		if node.Flags[f] {
			__antithesis_instrumentation__.Notify(609258)
			b.Add(ctx, f.String())
		} else {
			__antithesis_instrumentation__.Notify(609259)
		}
	}
	__antithesis_instrumentation__.Notify(609254)
	b.Finish(ctx)
	ctx.FormatNode(node.Statement)
}

func (node *Explain) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(609260)
	d := pretty.Keyword("EXPLAIN")
	var opts []pretty.Doc
	if node.Mode != ExplainPlan {
		__antithesis_instrumentation__.Notify(609264)
		opts = append(opts, pretty.Keyword(node.Mode.String()))
	} else {
		__antithesis_instrumentation__.Notify(609265)
	}
	__antithesis_instrumentation__.Notify(609261)
	for f := ExplainFlag(1); f <= numExplainFlags; f++ {
		__antithesis_instrumentation__.Notify(609266)
		if node.Flags[f] {
			__antithesis_instrumentation__.Notify(609267)
			opts = append(opts, pretty.Keyword(f.String()))
		} else {
			__antithesis_instrumentation__.Notify(609268)
		}
	}
	__antithesis_instrumentation__.Notify(609262)
	if len(opts) > 0 {
		__antithesis_instrumentation__.Notify(609269)
		d = pretty.ConcatSpace(
			d,
			p.bracket("(", p.commaSeparated(opts...), ")"),
		)
	} else {
		__antithesis_instrumentation__.Notify(609270)
	}
	__antithesis_instrumentation__.Notify(609263)
	return p.nestUnder(d, p.Doc(node.Statement))
}

func (node *ExplainAnalyze) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609271)
	ctx.WriteString("EXPLAIN ANALYZE ")
	b := util.MakeStringListBuilder("(", ", ", ") ")
	if node.Mode != ExplainPlan {
		__antithesis_instrumentation__.Notify(609274)
		b.Add(ctx, node.Mode.String())
	} else {
		__antithesis_instrumentation__.Notify(609275)
	}
	__antithesis_instrumentation__.Notify(609272)

	for f := ExplainFlag(1); f <= numExplainFlags; f++ {
		__antithesis_instrumentation__.Notify(609276)
		if node.Flags[f] {
			__antithesis_instrumentation__.Notify(609277)
			b.Add(ctx, f.String())
		} else {
			__antithesis_instrumentation__.Notify(609278)
		}
	}
	__antithesis_instrumentation__.Notify(609273)
	b.Finish(ctx)
	ctx.FormatNode(node.Statement)
}

func (node *ExplainAnalyze) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(609279)
	d := pretty.Keyword("EXPLAIN ANALYZE")
	var opts []pretty.Doc
	if node.Mode != ExplainPlan {
		__antithesis_instrumentation__.Notify(609283)
		opts = append(opts, pretty.Keyword(node.Mode.String()))
	} else {
		__antithesis_instrumentation__.Notify(609284)
	}
	__antithesis_instrumentation__.Notify(609280)
	for f := ExplainFlag(1); f <= numExplainFlags; f++ {
		__antithesis_instrumentation__.Notify(609285)
		if node.Flags[f] {
			__antithesis_instrumentation__.Notify(609286)
			opts = append(opts, pretty.Keyword(f.String()))
		} else {
			__antithesis_instrumentation__.Notify(609287)
		}
	}
	__antithesis_instrumentation__.Notify(609281)
	if len(opts) > 0 {
		__antithesis_instrumentation__.Notify(609288)
		d = pretty.ConcatSpace(
			d,
			p.bracket("(", p.commaSeparated(opts...), ")"),
		)
	} else {
		__antithesis_instrumentation__.Notify(609289)
	}
	__antithesis_instrumentation__.Notify(609282)
	return p.nestUnder(d, p.Doc(node.Statement))
}

func MakeExplain(options []string, stmt Statement) (Statement, error) {
	__antithesis_instrumentation__.Notify(609290)
	for i := range options {
		__antithesis_instrumentation__.Notify(609297)
		options[i] = strings.ToUpper(options[i])
	}
	__antithesis_instrumentation__.Notify(609291)
	var opts ExplainOptions
	var analyze bool
	for _, opt := range options {
		__antithesis_instrumentation__.Notify(609298)
		opt = strings.ToUpper(opt)
		if m, ok := explainModeStringMap[opt]; ok {
			__antithesis_instrumentation__.Notify(609302)
			if opts.Mode != 0 {
				__antithesis_instrumentation__.Notify(609304)
				return nil, pgerror.Newf(pgcode.Syntax, "cannot set EXPLAIN mode more than once: %s", opt)
			} else {
				__antithesis_instrumentation__.Notify(609305)
			}
			__antithesis_instrumentation__.Notify(609303)
			opts.Mode = m
			continue
		} else {
			__antithesis_instrumentation__.Notify(609306)
		}
		__antithesis_instrumentation__.Notify(609299)
		if opt == "ANALYZE" {
			__antithesis_instrumentation__.Notify(609307)
			analyze = true
			continue
		} else {
			__antithesis_instrumentation__.Notify(609308)
		}
		__antithesis_instrumentation__.Notify(609300)
		flag, ok := explainFlagStringMap[opt]
		if !ok {
			__antithesis_instrumentation__.Notify(609309)
			return nil, pgerror.Newf(pgcode.Syntax, "unsupported EXPLAIN option: %s", opt)
		} else {
			__antithesis_instrumentation__.Notify(609310)
		}
		__antithesis_instrumentation__.Notify(609301)
		opts.Flags[flag] = true
	}
	__antithesis_instrumentation__.Notify(609292)
	if opts.Mode == 0 {
		__antithesis_instrumentation__.Notify(609311)

		opts.Mode = ExplainPlan
	} else {
		__antithesis_instrumentation__.Notify(609312)
	}
	__antithesis_instrumentation__.Notify(609293)
	if opts.Flags[ExplainFlagJSON] {
		__antithesis_instrumentation__.Notify(609313)
		if opts.Mode != ExplainDistSQL {
			__antithesis_instrumentation__.Notify(609315)
			return nil, pgerror.Newf(pgcode.Syntax, "the JSON flag can only be used with DISTSQL")
		} else {
			__antithesis_instrumentation__.Notify(609316)
		}
		__antithesis_instrumentation__.Notify(609314)
		if analyze {
			__antithesis_instrumentation__.Notify(609317)
			return nil, pgerror.Newf(pgcode.Syntax, "the JSON flag cannot be used with ANALYZE")
		} else {
			__antithesis_instrumentation__.Notify(609318)
		}
	} else {
		__antithesis_instrumentation__.Notify(609319)
	}
	__antithesis_instrumentation__.Notify(609294)

	if analyze {
		__antithesis_instrumentation__.Notify(609320)
		if opts.Mode != ExplainDistSQL && func() bool {
			__antithesis_instrumentation__.Notify(609322)
			return opts.Mode != ExplainDebug == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(609323)
			return opts.Mode != ExplainPlan == true
		}() == true {
			__antithesis_instrumentation__.Notify(609324)
			return nil, pgerror.Newf(pgcode.Syntax, "EXPLAIN ANALYZE cannot be used with %s", opts.Mode)
		} else {
			__antithesis_instrumentation__.Notify(609325)
		}
		__antithesis_instrumentation__.Notify(609321)
		return &ExplainAnalyze{
			ExplainOptions: opts,
			Statement:      stmt,
		}, nil
	} else {
		__antithesis_instrumentation__.Notify(609326)
	}
	__antithesis_instrumentation__.Notify(609295)

	if opts.Mode == ExplainDebug {
		__antithesis_instrumentation__.Notify(609327)
		return nil, pgerror.Newf(pgcode.Syntax, "DEBUG flag can only be used with EXPLAIN ANALYZE")
	} else {
		__antithesis_instrumentation__.Notify(609328)
	}
	__antithesis_instrumentation__.Notify(609296)
	return &Explain{
		ExplainOptions: opts,
		Statement:      stmt,
	}, nil
}
