package clierror

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq"
)

func IsSQLSyntaxError(err error) bool {
	__antithesis_instrumentation__.Notify(28348)
	if pqErr := (*pq.Error)(nil); errors.As(err, &pqErr) {
		__antithesis_instrumentation__.Notify(28350)
		return string(pqErr.Code) == pgcode.Syntax.String()
	} else {
		__antithesis_instrumentation__.Notify(28351)
	}
	__antithesis_instrumentation__.Notify(28349)
	return false
}
