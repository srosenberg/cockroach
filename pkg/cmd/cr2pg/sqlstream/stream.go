// Package sqlstream streams an io.Reader into SQL statements.
package sqlstream

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"io"

	"github.com/cockroachdb/cockroach/pkg/sql/parser"

	_ "github.com/cockroachdb/cockroach/pkg/sql/sem/builtins"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
)

type Stream struct {
	scan *bufio.Scanner
}

func NewStream(r io.Reader) *Stream {
	__antithesis_instrumentation__.Notify(38212)
	const defaultMax = 1024 * 1024 * 32
	s := bufio.NewScanner(r)
	s.Buffer(make([]byte, 0, defaultMax), defaultMax)
	p := &Stream{scan: s}
	s.Split(splitSQLSemicolon)
	return p
}

func splitSQLSemicolon(data []byte, atEOF bool) (advance int, token []byte, err error) {
	__antithesis_instrumentation__.Notify(38213)
	if atEOF && func() bool {
		__antithesis_instrumentation__.Notify(38217)
		return len(data) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(38218)
		return 0, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(38219)
	}
	__antithesis_instrumentation__.Notify(38214)

	if pos, ok := parser.SplitFirstStatement(string(data)); ok {
		__antithesis_instrumentation__.Notify(38220)
		return pos, data[:pos], nil
	} else {
		__antithesis_instrumentation__.Notify(38221)
	}
	__antithesis_instrumentation__.Notify(38215)

	if atEOF {
		__antithesis_instrumentation__.Notify(38222)
		return len(data), data, nil
	} else {
		__antithesis_instrumentation__.Notify(38223)
	}
	__antithesis_instrumentation__.Notify(38216)

	return 0, nil, nil
}

func (s *Stream) Next() (tree.Statement, error) {
	__antithesis_instrumentation__.Notify(38224)
	for s.scan.Scan() {
		__antithesis_instrumentation__.Notify(38227)
		t := s.scan.Text()
		stmts, err := parser.Parse(t)
		if err != nil {
			__antithesis_instrumentation__.Notify(38229)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(38230)
		}
		__antithesis_instrumentation__.Notify(38228)
		switch len(stmts) {
		case 0:
			__antithesis_instrumentation__.Notify(38231)

		case 1:
			__antithesis_instrumentation__.Notify(38232)
			return stmts[0].AST, nil
		default:
			__antithesis_instrumentation__.Notify(38233)
			return nil, errors.Errorf("unexpected: got %d statements", len(stmts))
		}
	}
	__antithesis_instrumentation__.Notify(38225)
	if err := s.scan.Err(); err != nil {
		__antithesis_instrumentation__.Notify(38234)
		if errors.Is(err, bufio.ErrTooLong) {
			__antithesis_instrumentation__.Notify(38236)
			err = errors.HandledWithMessage(err, "line too long")
		} else {
			__antithesis_instrumentation__.Notify(38237)
		}
		__antithesis_instrumentation__.Notify(38235)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(38238)
	}
	__antithesis_instrumentation__.Notify(38226)
	return nil, io.EOF
}
