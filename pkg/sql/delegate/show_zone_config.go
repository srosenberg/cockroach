package delegate

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/sql/sem/tree"

func (d *delegator) delegateShowZoneConfig(n *tree.ShowZoneConfig) (tree.Statement, error) {
	__antithesis_instrumentation__.Notify(465887)

	if n.ZoneSpecifier != (tree.ZoneSpecifier{}) {
		__antithesis_instrumentation__.Notify(465889)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(465890)
	}
	__antithesis_instrumentation__.Notify(465888)
	return parse(`SELECT target, raw_config_sql FROM crdb_internal.zones`)
}
