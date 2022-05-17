package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
)

type Statement struct {
	parser.Statement

	StmtNoConstants string
	StmtSummary     string
	QueryID         ClusterWideID

	ExpectedTypes colinfo.ResultColumns

	Prepared *PreparedStatement
}

func makeStatement(parserStmt parser.Statement, queryID ClusterWideID) Statement {
	__antithesis_instrumentation__.Notify(626042)
	return Statement{
		Statement:       parserStmt,
		StmtNoConstants: formatStatementHideConstants(parserStmt.AST),
		StmtSummary:     formatStatementSummary(parserStmt.AST),
		QueryID:         queryID,
	}
}

func makeStatementFromPrepared(prepared *PreparedStatement, queryID ClusterWideID) Statement {
	__antithesis_instrumentation__.Notify(626043)
	return Statement{
		Statement:       prepared.Statement,
		Prepared:        prepared,
		ExpectedTypes:   prepared.Columns,
		StmtNoConstants: prepared.StatementNoConstants,
		StmtSummary:     prepared.StatementSummary,
		QueryID:         queryID,
	}
}

func (s Statement) String() string {
	__antithesis_instrumentation__.Notify(626044)

	return s.AST.String()
}
