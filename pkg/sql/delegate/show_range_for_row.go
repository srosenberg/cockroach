package delegate

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"encoding/hex"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/opt/cat"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/errors"
)

func (d *delegator) delegateShowRangeForRow(n *tree.ShowRangeForRow) (tree.Statement, error) {
	__antithesis_instrumentation__.Notify(465690)
	flags := cat.Flags{AvoidDescriptorCaches: true}
	idx, resName, err := cat.ResolveTableIndex(d.ctx, d.catalog, flags, &n.TableOrIndex)
	if err != nil {
		__antithesis_instrumentation__.Notify(465695)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(465696)
	}
	__antithesis_instrumentation__.Notify(465691)
	if err := checkPrivilegesForShowRanges(d, idx.Table()); err != nil {
		__antithesis_instrumentation__.Notify(465697)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(465698)
	}
	__antithesis_instrumentation__.Notify(465692)
	if idx.Table().IsVirtualTable() {
		__antithesis_instrumentation__.Notify(465699)
		return nil, errors.New("SHOW RANGE FOR ROW may not be called on a virtual table")
	} else {
		__antithesis_instrumentation__.Notify(465700)
	}
	__antithesis_instrumentation__.Notify(465693)
	span := idx.Span()
	table := idx.Table()
	idxSpanStart := hex.EncodeToString(span.Key)
	idxSpanEnd := hex.EncodeToString(span.EndKey)

	sqltelemetry.IncrementShowCounter(sqltelemetry.RangeForRow)

	var fmtCtx tree.FmtCtx
	fmtCtx.WriteString("(")
	if len(n.Row) == 1 {
		__antithesis_instrumentation__.Notify(465701)
		fmtCtx.FormatNode(n.Row[0])
		fmtCtx.WriteString(",")
	} else {
		__antithesis_instrumentation__.Notify(465702)
		fmtCtx.FormatNode(&n.Row)
	}
	__antithesis_instrumentation__.Notify(465694)
	fmtCtx.WriteString(")")
	rowString := fmtCtx.String()

	const query = `
SELECT
	CASE WHEN r.start_key < x'%[5]s' THEN NULL ELSE crdb_internal.pretty_key(r.start_key, 2) END AS start_key,
	CASE WHEN r.end_key >= x'%[6]s' THEN NULL ELSE crdb_internal.pretty_key(r.end_key, 2) END AS end_key,
	range_id,
	lease_holder,
	replica_localities[array_position(replicas, lease_holder)] as lease_holder_locality,
	replicas,
	replica_localities
FROM %[4]s.crdb_internal.ranges AS r
WHERE (r.start_key <= crdb_internal.encode_key(%[1]d, %[2]d, %[3]s))
  AND (r.end_key   >  crdb_internal.encode_key(%[1]d, %[2]d, %[3]s)) ORDER BY r.start_key
	`

	return parse(
		fmt.Sprintf(
			query,
			table.ID(),
			idx.ID(),
			rowString,
			resName.CatalogName.String(),
			idxSpanStart,
			idxSpanEnd,
		),
	)
}
