package delegate

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

func (d *delegator) delegateShowSequences(n *tree.ShowSequences) (tree.Statement, error) {
	__antithesis_instrumentation__.Notify(465788)
	name, err := d.getSpecifiedOrCurrentDatabase(n.Database)
	if err != nil {
		__antithesis_instrumentation__.Notify(465790)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(465791)
	}
	__antithesis_instrumentation__.Notify(465789)

	getSequencesQuery := fmt.Sprintf(`
	  SELECT sequence_schema, sequence_name
	    FROM %[1]s.information_schema.sequences
	   WHERE sequence_catalog = %[2]s
	ORDER BY sequence_name`,
		name.String(),
		lexbase.EscapeSQLString(string(name)),
	)
	return parse(getSequencesQuery)
}
