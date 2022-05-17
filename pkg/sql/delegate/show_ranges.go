package delegate

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"encoding/hex"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/cat"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/errors"
)

func checkPrivilegesForShowRanges(d *delegator, table cat.Table) error {
	__antithesis_instrumentation__.Notify(465703)

	if err := d.catalog.CheckPrivilege(d.ctx, table, privilege.SELECT); err != nil {
		__antithesis_instrumentation__.Notify(465708)
		return err
	} else {
		__antithesis_instrumentation__.Notify(465709)
	}
	__antithesis_instrumentation__.Notify(465704)
	hasAdmin, err := d.catalog.HasAdminRole(d.ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(465710)
		return err
	} else {
		__antithesis_instrumentation__.Notify(465711)
	}
	__antithesis_instrumentation__.Notify(465705)

	if hasAdmin {
		__antithesis_instrumentation__.Notify(465712)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(465713)
	}
	__antithesis_instrumentation__.Notify(465706)
	if err := d.catalog.CheckPrivilege(d.ctx, table, privilege.ZONECONFIG); err != nil {
		__antithesis_instrumentation__.Notify(465714)
		return pgerror.Wrapf(err, pgcode.InsufficientPrivilege, "only users with the ZONECONFIG privilege or the admin role can use SHOW RANGES on %s", table.Name())
	} else {
		__antithesis_instrumentation__.Notify(465715)
	}
	__antithesis_instrumentation__.Notify(465707)
	return nil
}

func (d *delegator) delegateShowRanges(n *tree.ShowRanges) (tree.Statement, error) {
	__antithesis_instrumentation__.Notify(465716)
	sqltelemetry.IncrementShowCounter(sqltelemetry.Ranges)
	if n.DatabaseName != "" {
		__antithesis_instrumentation__.Notify(465721)
		const dbQuery = `
		SELECT
			table_name,
			CASE
				WHEN crdb_internal.pretty_key(r.start_key, 2) = '' THEN NULL
				ELSE crdb_internal.pretty_key(r.start_key, 2)
			END AS start_key,
			CASE
				WHEN crdb_internal.pretty_key(r.end_key, 2) = '' THEN NULL
				ELSE crdb_internal.pretty_key(r.end_key, 2)
			END AS end_key,
			range_id,
			range_size / 1000000 as range_size_mb,
			lease_holder,
			replica_localities[array_position(replicas, lease_holder)] as lease_holder_locality,
			replicas,
			replica_localities
		FROM %[1]s.crdb_internal.ranges AS r
		WHERE database_name=%[2]s
		ORDER BY table_name, r.start_key
		`

		return parse(fmt.Sprintf(dbQuery, n.DatabaseName.String(), lexbase.EscapeSQLString(string(n.DatabaseName))))
	} else {
		__antithesis_instrumentation__.Notify(465722)
	}
	__antithesis_instrumentation__.Notify(465717)

	idx, resName, err := cat.ResolveTableIndex(
		d.ctx, d.catalog, cat.Flags{AvoidDescriptorCaches: true}, &n.TableOrIndex,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(465723)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(465724)
	}
	__antithesis_instrumentation__.Notify(465718)

	if err := checkPrivilegesForShowRanges(d, idx.Table()); err != nil {
		__antithesis_instrumentation__.Notify(465725)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(465726)
	}
	__antithesis_instrumentation__.Notify(465719)

	if idx.Table().IsVirtualTable() {
		__antithesis_instrumentation__.Notify(465727)
		return nil, errors.New("SHOW RANGES may not be called on a virtual table")
	} else {
		__antithesis_instrumentation__.Notify(465728)
	}
	__antithesis_instrumentation__.Notify(465720)

	span := idx.Span()
	startKey := hex.EncodeToString(span.Key)
	endKey := hex.EncodeToString(span.EndKey)
	return parse(fmt.Sprintf(`
SELECT 
  CASE WHEN r.start_key <= x'%[1]s' THEN NULL ELSE crdb_internal.pretty_key(r.start_key, 2) END AS start_key,
  CASE WHEN r.end_key >= x'%[2]s' THEN NULL ELSE crdb_internal.pretty_key(r.end_key, 2) END AS end_key,
  range_id,
  range_size / 1000000 as range_size_mb,
  lease_holder,
  replica_localities[array_position(replicas, lease_holder)] as lease_holder_locality,
  replicas,
  replica_localities
FROM %[3]s.crdb_internal.ranges AS r
WHERE (r.start_key < x'%[2]s')
  AND (r.end_key   > x'%[1]s') ORDER BY r.start_key
`,
		startKey, endKey, resName.CatalogName.String(),
	))
}
