package parser

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	unimp "github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/errors"
)

type lexer struct {
	in string

	tokens []sqlSymType

	nakedIntType *types.T

	lastPos int

	stmt tree.Statement

	numPlaceholders int
	numAnnotations  tree.AnnotationIdx

	lastError error
}

func (l *lexer) init(sql string, tokens []sqlSymType, nakedIntType *types.T) {
	l.in = sql
	l.tokens = tokens
	l.lastPos = -1
	l.stmt = nil
	l.numPlaceholders = 0
	l.numAnnotations = 0
	l.lastError = nil

	l.nakedIntType = nakedIntType
}

func (l *lexer) cleanup() {
	__antithesis_instrumentation__.Notify(552369)
	l.tokens = nil
	l.stmt = nil
	l.lastError = nil
}

func (l *lexer) Lex(lval *sqlSymType) int {
	__antithesis_instrumentation__.Notify(552370)
	l.lastPos++

	if l.lastPos >= len(l.tokens) {
		__antithesis_instrumentation__.Notify(552373)
		lval.id = 0
		lval.pos = int32(len(l.in))
		lval.str = "EOF"
		return 0
	} else {
		__antithesis_instrumentation__.Notify(552374)
	}
	__antithesis_instrumentation__.Notify(552371)
	*lval = l.tokens[l.lastPos]

	switch lval.id {
	case NOT, WITH, AS, GENERATED, NULLS, RESET, ROLE, USER, ON, TENANT:
		__antithesis_instrumentation__.Notify(552375)
		nextID := int32(0)
		if l.lastPos+1 < len(l.tokens) {
			__antithesis_instrumentation__.Notify(552379)
			nextID = l.tokens[l.lastPos+1].id
		} else {
			__antithesis_instrumentation__.Notify(552380)
		}
		__antithesis_instrumentation__.Notify(552376)
		secondID := int32(0)
		if l.lastPos+2 < len(l.tokens) {
			__antithesis_instrumentation__.Notify(552381)
			secondID = l.tokens[l.lastPos+2].id
		} else {
			__antithesis_instrumentation__.Notify(552382)
		}
		__antithesis_instrumentation__.Notify(552377)

		switch lval.id {
		case AS:
			__antithesis_instrumentation__.Notify(552383)
			switch nextID {
			case OF:
				__antithesis_instrumentation__.Notify(552394)
				lval.id = AS_LA
			default:
				__antithesis_instrumentation__.Notify(552395)
			}
		case NOT:
			__antithesis_instrumentation__.Notify(552384)
			switch nextID {
			case BETWEEN, IN, LIKE, ILIKE, SIMILAR:
				__antithesis_instrumentation__.Notify(552396)
				lval.id = NOT_LA
			default:
				__antithesis_instrumentation__.Notify(552397)
			}
		case GENERATED:
			__antithesis_instrumentation__.Notify(552385)
			switch nextID {
			case ALWAYS:
				__antithesis_instrumentation__.Notify(552398)
				lval.id = GENERATED_ALWAYS
			case BY:
				__antithesis_instrumentation__.Notify(552399)
				lval.id = GENERATED_BY_DEFAULT
			default:
				__antithesis_instrumentation__.Notify(552400)
			}

		case WITH:
			__antithesis_instrumentation__.Notify(552386)
			switch nextID {
			case TIME, ORDINALITY, BUCKET_COUNT:
				__antithesis_instrumentation__.Notify(552401)
				lval.id = WITH_LA
			default:
				__antithesis_instrumentation__.Notify(552402)
			}
		case NULLS:
			__antithesis_instrumentation__.Notify(552387)
			switch nextID {
			case FIRST, LAST:
				__antithesis_instrumentation__.Notify(552403)
				lval.id = NULLS_LA
			default:
				__antithesis_instrumentation__.Notify(552404)
			}
		case RESET:
			__antithesis_instrumentation__.Notify(552388)
			switch nextID {
			case ALL:
				__antithesis_instrumentation__.Notify(552405)
				lval.id = RESET_ALL
			default:
				__antithesis_instrumentation__.Notify(552406)
			}
		case ROLE:
			__antithesis_instrumentation__.Notify(552389)
			switch nextID {
			case ALL:
				__antithesis_instrumentation__.Notify(552407)
				lval.id = ROLE_ALL
			default:
				__antithesis_instrumentation__.Notify(552408)
			}
		case USER:
			__antithesis_instrumentation__.Notify(552390)
			switch nextID {
			case ALL:
				__antithesis_instrumentation__.Notify(552409)
				lval.id = USER_ALL
			default:
				__antithesis_instrumentation__.Notify(552410)
			}
		case ON:
			__antithesis_instrumentation__.Notify(552391)
			switch nextID {
			case DELETE:
				__antithesis_instrumentation__.Notify(552411)
				lval.id = ON_LA
			case UPDATE:
				__antithesis_instrumentation__.Notify(552412)
				switch secondID {
				case NO, RESTRICT, CASCADE, SET:
					__antithesis_instrumentation__.Notify(552414)
					lval.id = ON_LA
				default:
					__antithesis_instrumentation__.Notify(552415)
				}
			default:
				__antithesis_instrumentation__.Notify(552413)
			}
		case TENANT:
			__antithesis_instrumentation__.Notify(552392)
			switch nextID {
			case ALL:
				__antithesis_instrumentation__.Notify(552416)
				lval.id = TENANT_ALL
			default:
				__antithesis_instrumentation__.Notify(552417)
			}
		default:
			__antithesis_instrumentation__.Notify(552393)
		}
	default:
		__antithesis_instrumentation__.Notify(552378)
	}
	__antithesis_instrumentation__.Notify(552372)

	return int(lval.id)
}

func (l *lexer) lastToken() sqlSymType {
	__antithesis_instrumentation__.Notify(552418)
	if l.lastPos < 0 {
		__antithesis_instrumentation__.Notify(552421)
		return sqlSymType{}
	} else {
		__antithesis_instrumentation__.Notify(552422)
	}
	__antithesis_instrumentation__.Notify(552419)

	if l.lastPos >= len(l.tokens) {
		__antithesis_instrumentation__.Notify(552423)
		return sqlSymType{
			id:  0,
			pos: int32(len(l.in)),
			str: "EOF",
		}
	} else {
		__antithesis_instrumentation__.Notify(552424)
	}
	__antithesis_instrumentation__.Notify(552420)
	return l.tokens[l.lastPos]
}

func (l *lexer) NewAnnotation() tree.AnnotationIdx {
	__antithesis_instrumentation__.Notify(552425)
	l.numAnnotations++
	return l.numAnnotations
}

func (l *lexer) SetStmt(stmt tree.Statement) {
	__antithesis_instrumentation__.Notify(552426)
	l.stmt = stmt
}

func (l *lexer) UpdateNumPlaceholders(p *tree.Placeholder) {
	__antithesis_instrumentation__.Notify(552427)
	if n := int(p.Idx) + 1; l.numPlaceholders < n {
		__antithesis_instrumentation__.Notify(552428)
		l.numPlaceholders = n
	} else {
		__antithesis_instrumentation__.Notify(552429)
	}
}

func (l *lexer) PurposelyUnimplemented(feature string, reason string) {
	__antithesis_instrumentation__.Notify(552430)

	l.lastError = errors.WithHint(
		errors.WithTelemetry(
			pgerror.Newf(pgcode.Syntax, "unimplemented: this syntax"),
			fmt.Sprintf("sql.purposely_unimplemented.%s", feature),
		),
		reason,
	)
	l.populateErrorDetails()
	l.lastError = &tree.UnsupportedError{
		Err:         l.lastError,
		FeatureName: feature,
	}
}

func (l *lexer) UnimplementedWithIssue(issue int) {
	__antithesis_instrumentation__.Notify(552431)
	l.lastError = unimp.NewWithIssue(issue, "this syntax")
	l.populateErrorDetails()
	l.lastError = &tree.UnsupportedError{
		Err:         l.lastError,
		FeatureName: fmt.Sprintf("https://github.com/cockroachdb/cockroach/issues/%d", issue),
	}
}

func (l *lexer) UnimplementedWithIssueDetail(issue int, detail string) {
	__antithesis_instrumentation__.Notify(552432)
	l.lastError = unimp.NewWithIssueDetail(issue, detail, "this syntax")
	l.populateErrorDetails()
	l.lastError = &tree.UnsupportedError{
		Err:         l.lastError,
		FeatureName: detail,
	}
}

func (l *lexer) Unimplemented(feature string) {
	__antithesis_instrumentation__.Notify(552433)
	l.lastError = unimp.New(feature, "this syntax")
	l.populateErrorDetails()
	l.lastError = &tree.UnsupportedError{
		Err:         l.lastError,
		FeatureName: feature,
	}
}

func (l *lexer) setErr(err error) {
	__antithesis_instrumentation__.Notify(552434)
	err = pgerror.WithCandidateCode(err, pgcode.Syntax)
	l.lastError = err
	l.populateErrorDetails()
}

func (l *lexer) Error(e string) {
	__antithesis_instrumentation__.Notify(552435)
	e = strings.TrimPrefix(e, "syntax error: ")
	l.lastError = pgerror.WithCandidateCode(errors.Newf("%s", e), pgcode.Syntax)
	l.populateErrorDetails()
}

func (l *lexer) populateErrorDetails() {
	__antithesis_instrumentation__.Notify(552436)
	lastTok := l.lastToken()

	if lastTok.id == ERROR {
		__antithesis_instrumentation__.Notify(552439)

		err := pgerror.WithCandidateCode(errors.Newf("lexical error: %s", lastTok.str), pgcode.Syntax)
		l.lastError = errors.WithSecondaryError(err, l.lastError)
	} else {
		__antithesis_instrumentation__.Notify(552440)

		if !strings.Contains(l.lastError.Error(), "syntax error") {
			__antithesis_instrumentation__.Notify(552442)

			l.lastError = errors.Wrap(l.lastError, "syntax error")
		} else {
			__antithesis_instrumentation__.Notify(552443)
		}
		__antithesis_instrumentation__.Notify(552441)
		l.lastError = errors.Wrapf(l.lastError, "at or near \"%s\"", lastTok.str)
	}
	__antithesis_instrumentation__.Notify(552437)

	i := strings.IndexByte(l.in[lastTok.pos:], '\n')
	if i == -1 {
		__antithesis_instrumentation__.Notify(552444)
		i = len(l.in)
	} else {
		__antithesis_instrumentation__.Notify(552445)
		i += int(lastTok.pos)
	}
	__antithesis_instrumentation__.Notify(552438)

	j := strings.LastIndexByte(l.in[:lastTok.pos], '\n') + 1

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "source SQL:\n%s\n", l.in[:i])

	fmt.Fprintf(&buf, "%s^", strings.Repeat(" ", int(lastTok.pos)-j))
	l.lastError = errors.WithDetail(l.lastError, buf.String())
}

func (l *lexer) SetHelp(msg HelpMessage) {
	__antithesis_instrumentation__.Notify(552446)
	if l.lastError == nil {
		__antithesis_instrumentation__.Notify(552448)
		l.lastError = pgerror.WithCandidateCode(errors.New("help request"), pgcode.Syntax)
	} else {
		__antithesis_instrumentation__.Notify(552449)
	}
	__antithesis_instrumentation__.Notify(552447)

	if lastTok := l.lastToken(); lastTok.id == HELPTOKEN {
		__antithesis_instrumentation__.Notify(552450)
		l.populateHelpMsg(msg.String())
	} else {
		__antithesis_instrumentation__.Notify(552451)
		if msg.Command != "" {
			__antithesis_instrumentation__.Notify(552452)
			l.lastError = errors.WithHintf(l.lastError, `try \h %s`, msg.Command)
		} else {
			__antithesis_instrumentation__.Notify(552453)
			l.lastError = errors.WithHintf(l.lastError, `try \hf %s`, msg.Function)
		}
	}
}

const specialHelpErrorPrefix = "help token in input"

func (l *lexer) populateHelpMsg(msg string) {
	__antithesis_instrumentation__.Notify(552454)
	l.lastError = errors.WithHint(errors.Wrap(l.lastError, specialHelpErrorPrefix), msg)
}
