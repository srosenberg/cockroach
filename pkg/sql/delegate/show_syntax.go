package delegate

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

func (d *delegator) delegateShowSyntax(n *tree.ShowSyntax) (tree.Statement, error) {
	__antithesis_instrumentation__.Notify(465803)

	var query bytes.Buffer
	fmt.Fprintf(
		&query, "SELECT @1 AS %s, @2 AS %s FROM (VALUES ",
		colinfo.ShowSyntaxColumns[0].Name, colinfo.ShowSyntaxColumns[1].Name,
	)

	comma := ""

	parser.RunShowSyntax(
		d.ctx, n.Statement,
		func(ctx context.Context, field, msg string) {
			__antithesis_instrumentation__.Notify(465805)
			fmt.Fprintf(&query, "%s('%s', ", comma, field)
			lexbase.EncodeSQLString(&query, msg)
			query.WriteByte(')')
			comma = ", "
		},
		nil,
	)
	__antithesis_instrumentation__.Notify(465804)
	query.WriteByte(')')

	return parse(query.String())
}
