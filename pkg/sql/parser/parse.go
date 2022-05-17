package parser

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"go/constant"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/scanner"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
)

func init() {
	scanner.NewNumValFn = func(a constant.Value, s string, b bool) interface{} { return tree.NewNumVal(a, s, b) }
	scanner.NewPlaceholderFn = func(s string) (interface{}, error) { return tree.NewPlaceholder(s) }
}

type Statement struct {
	AST tree.Statement

	SQL string

	NumPlaceholders int

	NumAnnotations tree.AnnotationIdx
}

type Statements []Statement

func (stmts Statements) String() string {
	__antithesis_instrumentation__.Notify(552455)
	return stmts.StringWithFlags(tree.FmtSimple)
}

func (stmts Statements) StringWithFlags(flags tree.FmtFlags) string {
	__antithesis_instrumentation__.Notify(552456)
	ctx := tree.NewFmtCtx(flags)
	for i, s := range stmts {
		__antithesis_instrumentation__.Notify(552458)
		if i > 0 {
			__antithesis_instrumentation__.Notify(552460)
			ctx.WriteString("; ")
		} else {
			__antithesis_instrumentation__.Notify(552461)
		}
		__antithesis_instrumentation__.Notify(552459)
		ctx.FormatNode(s.AST)
	}
	__antithesis_instrumentation__.Notify(552457)
	return ctx.CloseAndGetString()
}

type Parser struct {
	scanner    scanner.Scanner
	lexer      lexer
	parserImpl sqlParserImpl
	tokBuf     [8]sqlSymType
	stmtBuf    [1]Statement
}

var defaultNakedIntType = types.Int

func NakedIntTypeFromDefaultIntSize(defaultIntSize int32) *types.T {
	__antithesis_instrumentation__.Notify(552462)
	switch defaultIntSize {
	case 4, 32:
		__antithesis_instrumentation__.Notify(552463)
		return types.Int4
	default:
		__antithesis_instrumentation__.Notify(552464)
		return types.Int
	}
}

func (p *Parser) Parse(sql string) (Statements, error) {
	__antithesis_instrumentation__.Notify(552465)
	return p.parseWithDepth(1, sql, defaultNakedIntType)
}

func (p *Parser) ParseWithInt(sql string, nakedIntType *types.T) (Statements, error) {
	__antithesis_instrumentation__.Notify(552466)
	return p.parseWithDepth(1, sql, nakedIntType)
}

func (p *Parser) parseOneWithInt(sql string, nakedIntType *types.T) (Statement, error) {
	__antithesis_instrumentation__.Notify(552467)
	stmts, err := p.parseWithDepth(1, sql, nakedIntType)
	if err != nil {
		__antithesis_instrumentation__.Notify(552470)
		return Statement{}, err
	} else {
		__antithesis_instrumentation__.Notify(552471)
	}
	__antithesis_instrumentation__.Notify(552468)
	if len(stmts) != 1 {
		__antithesis_instrumentation__.Notify(552472)
		return Statement{}, errors.AssertionFailedf("expected 1 statement, but found %d", len(stmts))
	} else {
		__antithesis_instrumentation__.Notify(552473)
	}
	__antithesis_instrumentation__.Notify(552469)
	return stmts[0], nil
}

func (p *Parser) scanOneStmt() (sql string, tokens []sqlSymType, done bool) {
	__antithesis_instrumentation__.Notify(552474)
	var lval sqlSymType
	tokens = p.tokBuf[:0]

	for {
		__antithesis_instrumentation__.Notify(552476)
		p.scanner.Scan(&lval)
		if lval.id == 0 {
			__antithesis_instrumentation__.Notify(552478)
			return "", nil, true
		} else {
			__antithesis_instrumentation__.Notify(552479)
		}
		__antithesis_instrumentation__.Notify(552477)
		if lval.id != ';' {
			__antithesis_instrumentation__.Notify(552480)
			break
		} else {
			__antithesis_instrumentation__.Notify(552481)
		}
	}
	__antithesis_instrumentation__.Notify(552475)

	startPos := lval.pos

	lval.pos = 0
	tokens = append(tokens, lval)
	for {
		__antithesis_instrumentation__.Notify(552482)
		if lval.id == ERROR {
			__antithesis_instrumentation__.Notify(552485)
			return p.scanner.In()[startPos:], tokens, true
		} else {
			__antithesis_instrumentation__.Notify(552486)
		}
		__antithesis_instrumentation__.Notify(552483)
		posBeforeScan := p.scanner.Pos()
		p.scanner.Scan(&lval)
		if lval.id == 0 || func() bool {
			__antithesis_instrumentation__.Notify(552487)
			return lval.id == ';' == true
		}() == true {
			__antithesis_instrumentation__.Notify(552488)
			return p.scanner.In()[startPos:posBeforeScan], tokens, (lval.id == 0)
		} else {
			__antithesis_instrumentation__.Notify(552489)
		}
		__antithesis_instrumentation__.Notify(552484)
		lval.pos -= startPos
		tokens = append(tokens, lval)
	}
}

func (p *Parser) parseWithDepth(depth int, sql string, nakedIntType *types.T) (Statements, error) {
	__antithesis_instrumentation__.Notify(552490)
	stmts := Statements(p.stmtBuf[:0])
	p.scanner.Init(sql)
	defer p.scanner.Cleanup()
	for {
		__antithesis_instrumentation__.Notify(552492)
		sql, tokens, done := p.scanOneStmt()
		stmt, err := p.parse(depth+1, sql, tokens, nakedIntType)
		if err != nil {
			__antithesis_instrumentation__.Notify(552495)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(552496)
		}
		__antithesis_instrumentation__.Notify(552493)
		if stmt.AST != nil {
			__antithesis_instrumentation__.Notify(552497)
			stmts = append(stmts, stmt)
		} else {
			__antithesis_instrumentation__.Notify(552498)
		}
		__antithesis_instrumentation__.Notify(552494)
		if done {
			__antithesis_instrumentation__.Notify(552499)
			break
		} else {
			__antithesis_instrumentation__.Notify(552500)
		}
	}
	__antithesis_instrumentation__.Notify(552491)
	return stmts, nil
}

func (p *Parser) parse(
	depth int, sql string, tokens []sqlSymType, nakedIntType *types.T,
) (Statement, error) {
	__antithesis_instrumentation__.Notify(552501)
	p.lexer.init(sql, tokens, nakedIntType)
	defer p.lexer.cleanup()
	if p.parserImpl.Parse(&p.lexer) != 0 {
		__antithesis_instrumentation__.Notify(552503)
		if p.lexer.lastError == nil {
			__antithesis_instrumentation__.Notify(552506)

			p.lexer.Error("syntax error")
		} else {
			__antithesis_instrumentation__.Notify(552507)
		}
		__antithesis_instrumentation__.Notify(552504)
		err := p.lexer.lastError

		tkeys := errors.GetTelemetryKeys(err)
		if len(tkeys) > 0 {
			__antithesis_instrumentation__.Notify(552508)
			for i := range tkeys {
				__antithesis_instrumentation__.Notify(552510)
				tkeys[i] = "syntax." + tkeys[i]
			}
			__antithesis_instrumentation__.Notify(552509)
			err = errors.WithTelemetry(err, tkeys...)
		} else {
			__antithesis_instrumentation__.Notify(552511)
		}
		__antithesis_instrumentation__.Notify(552505)

		return Statement{}, err
	} else {
		__antithesis_instrumentation__.Notify(552512)
	}
	__antithesis_instrumentation__.Notify(552502)
	return Statement{
		AST:             p.lexer.stmt,
		SQL:             sql,
		NumPlaceholders: p.lexer.numPlaceholders,
		NumAnnotations:  p.lexer.numAnnotations,
	}, nil
}

func unaryNegation(e tree.Expr) tree.Expr {
	__antithesis_instrumentation__.Notify(552513)
	if cst, ok := e.(*tree.NumVal); ok {
		__antithesis_instrumentation__.Notify(552515)
		cst.Negate()
		return cst
	} else {
		__antithesis_instrumentation__.Notify(552516)
	}
	__antithesis_instrumentation__.Notify(552514)

	return &tree.UnaryExpr{
		Operator: tree.MakeUnaryOperator(tree.UnaryMinus),
		Expr:     e,
	}
}

func Parse(sql string) (Statements, error) {
	__antithesis_instrumentation__.Notify(552517)
	var p Parser
	return p.parseWithDepth(1, sql, defaultNakedIntType)
}

func ParseOne(sql string) (Statement, error) {
	__antithesis_instrumentation__.Notify(552518)
	return ParseOneWithInt(sql, defaultNakedIntType)
}

func ParseOneWithInt(sql string, nakedIntType *types.T) (Statement, error) {
	__antithesis_instrumentation__.Notify(552519)
	var p Parser
	return p.parseOneWithInt(sql, nakedIntType)
}

func ParseQualifiedTableName(sql string) (*tree.TableName, error) {
	__antithesis_instrumentation__.Notify(552520)
	name, err := ParseTableName(sql)
	if err != nil {
		__antithesis_instrumentation__.Notify(552522)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(552523)
	}
	__antithesis_instrumentation__.Notify(552521)
	tn := name.ToTableName()
	return &tn, nil
}

func ParseTableName(sql string) (*tree.UnresolvedObjectName, error) {
	__antithesis_instrumentation__.Notify(552524)

	stmt, err := ParseOne(fmt.Sprintf("ALTER TABLE %s RENAME TO x", sql))
	if err != nil {
		__antithesis_instrumentation__.Notify(552527)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(552528)
	}
	__antithesis_instrumentation__.Notify(552525)
	rename, ok := stmt.AST.(*tree.RenameTable)
	if !ok {
		__antithesis_instrumentation__.Notify(552529)
		return nil, errors.AssertionFailedf("expected an ALTER TABLE statement, but found %T", stmt)
	} else {
		__antithesis_instrumentation__.Notify(552530)
	}
	__antithesis_instrumentation__.Notify(552526)
	return rename.Name, nil
}

func parseExprsWithInt(exprs []string, nakedIntType *types.T) (tree.Exprs, error) {
	__antithesis_instrumentation__.Notify(552531)
	stmt, err := ParseOneWithInt(fmt.Sprintf("SET ROW (%s)", strings.Join(exprs, ",")), nakedIntType)
	if err != nil {
		__antithesis_instrumentation__.Notify(552534)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(552535)
	}
	__antithesis_instrumentation__.Notify(552532)
	set, ok := stmt.AST.(*tree.SetVar)
	if !ok {
		__antithesis_instrumentation__.Notify(552536)
		return nil, errors.AssertionFailedf("expected a SET statement, but found %T", stmt)
	} else {
		__antithesis_instrumentation__.Notify(552537)
	}
	__antithesis_instrumentation__.Notify(552533)
	return set.Values, nil
}

func ParseExprs(sql []string) (tree.Exprs, error) {
	__antithesis_instrumentation__.Notify(552538)
	if len(sql) == 0 {
		__antithesis_instrumentation__.Notify(552540)
		return tree.Exprs{}, nil
	} else {
		__antithesis_instrumentation__.Notify(552541)
	}
	__antithesis_instrumentation__.Notify(552539)
	return parseExprsWithInt(sql, defaultNakedIntType)
}

func ParseExpr(sql string) (tree.Expr, error) {
	__antithesis_instrumentation__.Notify(552542)
	return ParseExprWithInt(sql, defaultNakedIntType)
}

func ParseExprWithInt(sql string, nakedIntType *types.T) (tree.Expr, error) {
	__antithesis_instrumentation__.Notify(552543)
	exprs, err := parseExprsWithInt([]string{sql}, nakedIntType)
	if err != nil {
		__antithesis_instrumentation__.Notify(552546)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(552547)
	}
	__antithesis_instrumentation__.Notify(552544)
	if len(exprs) != 1 {
		__antithesis_instrumentation__.Notify(552548)
		return nil, errors.AssertionFailedf("expected 1 expression, found %d", len(exprs))
	} else {
		__antithesis_instrumentation__.Notify(552549)
	}
	__antithesis_instrumentation__.Notify(552545)
	return exprs[0], nil
}

func GetTypeReferenceFromName(typeName tree.Name) (tree.ResolvableTypeReference, error) {
	__antithesis_instrumentation__.Notify(552550)
	expr, err := ParseExpr(fmt.Sprintf("1::%s", typeName.String()))
	if err != nil {
		__antithesis_instrumentation__.Notify(552553)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(552554)
	}
	__antithesis_instrumentation__.Notify(552551)

	cast, ok := expr.(*tree.CastExpr)
	if !ok {
		__antithesis_instrumentation__.Notify(552555)
		return nil, errors.AssertionFailedf("expected a tree.CastExpr, but found %T", expr)
	} else {
		__antithesis_instrumentation__.Notify(552556)
	}
	__antithesis_instrumentation__.Notify(552552)

	return cast.Type, nil
}

func GetTypeFromValidSQLSyntax(sql string) (tree.ResolvableTypeReference, error) {
	__antithesis_instrumentation__.Notify(552557)
	expr, err := ParseExpr(fmt.Sprintf("1::%s", sql))
	if err != nil {
		__antithesis_instrumentation__.Notify(552560)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(552561)
	}
	__antithesis_instrumentation__.Notify(552558)

	cast, ok := expr.(*tree.CastExpr)
	if !ok {
		__antithesis_instrumentation__.Notify(552562)
		return nil, errors.AssertionFailedf("expected a tree.CastExpr, but found %T", expr)
	} else {
		__antithesis_instrumentation__.Notify(552563)
	}
	__antithesis_instrumentation__.Notify(552559)

	return cast.Type, nil
}

var errBitLengthNotPositive = pgerror.WithCandidateCode(
	errors.New("length for type bit must be at least 1"), pgcode.InvalidParameterValue)

func newBitType(width int32, varying bool) (*types.T, error) {
	__antithesis_instrumentation__.Notify(552564)
	if width < 1 {
		__antithesis_instrumentation__.Notify(552567)
		return nil, errBitLengthNotPositive
	} else {
		__antithesis_instrumentation__.Notify(552568)
	}
	__antithesis_instrumentation__.Notify(552565)
	if varying {
		__antithesis_instrumentation__.Notify(552569)
		return types.MakeVarBit(width), nil
	} else {
		__antithesis_instrumentation__.Notify(552570)
	}
	__antithesis_instrumentation__.Notify(552566)
	return types.MakeBit(width), nil
}

var errFloatPrecAtLeast1 = pgerror.WithCandidateCode(
	errors.New("precision for type float must be at least 1 bit"), pgcode.InvalidParameterValue)
var errFloatPrecMax54 = pgerror.WithCandidateCode(
	errors.New("precision for type float must be less than 54 bits"), pgcode.InvalidParameterValue)

func newFloat(prec int64) (*types.T, error) {
	__antithesis_instrumentation__.Notify(552571)
	if prec < 1 {
		__antithesis_instrumentation__.Notify(552575)
		return nil, errFloatPrecAtLeast1
	} else {
		__antithesis_instrumentation__.Notify(552576)
	}
	__antithesis_instrumentation__.Notify(552572)
	if prec <= 24 {
		__antithesis_instrumentation__.Notify(552577)
		return types.Float4, nil
	} else {
		__antithesis_instrumentation__.Notify(552578)
	}
	__antithesis_instrumentation__.Notify(552573)
	if prec <= 54 {
		__antithesis_instrumentation__.Notify(552579)
		return types.Float, nil
	} else {
		__antithesis_instrumentation__.Notify(552580)
	}
	__antithesis_instrumentation__.Notify(552574)
	return nil, errFloatPrecMax54
}

func newDecimal(prec, scale int32) (*types.T, error) {
	__antithesis_instrumentation__.Notify(552581)
	if scale > prec {
		__antithesis_instrumentation__.Notify(552583)
		err := pgerror.WithCandidateCode(
			errors.Newf("scale (%d) must be between 0 and precision (%d)", scale, prec),
			pgcode.InvalidParameterValue)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(552584)
	}
	__antithesis_instrumentation__.Notify(552582)
	return types.MakeDecimal(prec, scale), nil
}

func arrayOf(
	ref tree.ResolvableTypeReference, bounds []int32,
) (tree.ResolvableTypeReference, error) {
	__antithesis_instrumentation__.Notify(552585)

	if typ, ok := tree.GetStaticallyKnownType(ref); ok {
		__antithesis_instrumentation__.Notify(552587)

		if typ.Family() == types.UnknownFamily {
			__antithesis_instrumentation__.Notify(552591)
			return nil, pgerror.Newf(pgcode.UndefinedObject, "type unknown[] does not exist")
		} else {
			__antithesis_instrumentation__.Notify(552592)
		}
		__antithesis_instrumentation__.Notify(552588)
		if typ.Family() == types.VoidFamily {
			__antithesis_instrumentation__.Notify(552593)
			return nil, pgerror.Newf(pgcode.UndefinedObject, "type void[] does not exist")
		} else {
			__antithesis_instrumentation__.Notify(552594)
		}
		__antithesis_instrumentation__.Notify(552589)
		if err := types.CheckArrayElementType(typ); err != nil {
			__antithesis_instrumentation__.Notify(552595)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(552596)
		}
		__antithesis_instrumentation__.Notify(552590)
		return types.MakeArray(typ), nil
	} else {
		__antithesis_instrumentation__.Notify(552597)
	}
	__antithesis_instrumentation__.Notify(552586)
	return &tree.ArrayTypeReference{ElementType: ref}, nil
}
