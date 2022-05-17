package delegate

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

func (d *delegator) delegateShowCompletions(n *tree.ShowCompletions) (tree.Statement, error) {
	__antithesis_instrumentation__.Notify(465481)
	offset, err := n.Offset.AsInt64()
	if err != nil {
		__antithesis_instrumentation__.Notify(465486)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(465487)
	}
	__antithesis_instrumentation__.Notify(465482)

	completions, err := RunShowCompletions(n.Statement.RawString(), int(offset))
	if err != nil {
		__antithesis_instrumentation__.Notify(465488)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(465489)
	}
	__antithesis_instrumentation__.Notify(465483)

	if len(completions) == 0 {
		__antithesis_instrumentation__.Notify(465490)
		return parse(`SELECT '' as completions`)
	} else {
		__antithesis_instrumentation__.Notify(465491)
	}
	__antithesis_instrumentation__.Notify(465484)

	var query bytes.Buffer
	fmt.Fprint(&query, "SELECT @1 AS completions FROM (VALUES ")

	comma := ""
	for _, completion := range completions {
		__antithesis_instrumentation__.Notify(465492)
		fmt.Fprintf(&query, "%s(", comma)
		lexbase.EncodeSQLString(&query, completion)
		query.WriteByte(')')
		comma = ", "
	}
	__antithesis_instrumentation__.Notify(465485)

	fmt.Fprintf(&query, ")")

	return parse(query.String())
}

func RunShowCompletions(stmt string, offset int) ([]string, error) {
	__antithesis_instrumentation__.Notify(465493)
	if offset <= 0 || func() bool {
		__antithesis_instrumentation__.Notify(465499)
		return offset > len(stmt) == true
	}() == true {
		__antithesis_instrumentation__.Notify(465500)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(465501)
	}
	__antithesis_instrumentation__.Notify(465494)

	if unicode.IsSpace([]rune(stmt)[offset-1]) {
		__antithesis_instrumentation__.Notify(465502)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(465503)
	}
	__antithesis_instrumentation__.Notify(465495)

	sqlTokens := parser.TokensIgnoreErrors(string([]rune(stmt)[:offset]))
	if len(sqlTokens) == 0 {
		__antithesis_instrumentation__.Notify(465504)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(465505)
	}
	__antithesis_instrumentation__.Notify(465496)

	sqlTokenStrings := make([]string, len(sqlTokens))
	for i, sqlToken := range sqlTokens {
		__antithesis_instrumentation__.Notify(465506)
		sqlTokenStrings[i] = sqlToken.Str
	}
	__antithesis_instrumentation__.Notify(465497)

	lastWordTruncated := sqlTokenStrings[len(sqlTokenStrings)-1]

	allSQLTokens := parser.TokensIgnoreErrors(stmt)
	lastWordFull := allSQLTokens[len(sqlTokenStrings)-1]
	if lastWordFull.Str != lastWordTruncated {
		__antithesis_instrumentation__.Notify(465507)
		return []string{strings.ToUpper(lastWordFull.Str)}, nil
	} else {
		__antithesis_instrumentation__.Notify(465508)
	}
	__antithesis_instrumentation__.Notify(465498)

	return getCompletionsForWord(lastWordTruncated, lexbase.KeywordNames), nil
}

func binarySearch(w string, words []string) (int, int) {
	__antithesis_instrumentation__.Notify(465509)

	left := sort.Search(len(words), func(i int) bool { __antithesis_instrumentation__.Notify(465512); return words[i] >= w })
	__antithesis_instrumentation__.Notify(465510)

	right := sort.Search(len(words), func(i int) bool {
		__antithesis_instrumentation__.Notify(465513)
		return words[i][:min(len(words[i]), len(w))] > w
	})
	__antithesis_instrumentation__.Notify(465511)

	return left, right
}

func min(a int, b int) int {
	__antithesis_instrumentation__.Notify(465514)
	if a < b {
		__antithesis_instrumentation__.Notify(465516)
		return a
	} else {
		__antithesis_instrumentation__.Notify(465517)
	}
	__antithesis_instrumentation__.Notify(465515)
	return b
}

func getCompletionsForWord(w string, words []string) []string {
	__antithesis_instrumentation__.Notify(465518)
	left, right := binarySearch(strings.ToLower(w), words)
	completions := make([]string, right-left)

	for i, word := range words[left:right] {
		__antithesis_instrumentation__.Notify(465520)
		completions[i] = strings.ToUpper(word)
	}
	__antithesis_instrumentation__.Notify(465519)
	return completions
}
