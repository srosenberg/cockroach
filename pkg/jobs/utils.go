package jobs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/errors"
)

func RunningJobExists(
	ctx context.Context,
	jobID jobspb.JobID,
	ie sqlutil.InternalExecutor,
	txn *kv.Txn,
	payloadPredicate func(payload *jobspb.Payload) bool,
) (exists bool, retErr error) {
	__antithesis_instrumentation__.Notify(85064)
	const stmt = `
SELECT
  id, payload
FROM
  system.jobs
WHERE
  status IN ` + NonTerminalStatusTupleString + `
ORDER BY created`

	it, err := ie.QueryIterator(
		ctx,
		"get-jobs",
		txn,
		stmt,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(85068)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(85069)
	}
	__antithesis_instrumentation__.Notify(85065)

	defer func() {
		__antithesis_instrumentation__.Notify(85070)
		retErr = errors.CombineErrors(retErr, it.Close())
	}()
	__antithesis_instrumentation__.Notify(85066)

	var ok bool
	for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(85071)
		row := it.Cur()
		payload, err := UnmarshalPayload(row[1])
		if err != nil {
			__antithesis_instrumentation__.Notify(85073)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(85074)
		}
		__antithesis_instrumentation__.Notify(85072)

		if payloadPredicate(payload) {
			__antithesis_instrumentation__.Notify(85075)
			id := jobspb.JobID(*row[0].(*tree.DInt))
			if id == jobID {
				__antithesis_instrumentation__.Notify(85077)
				break
			} else {
				__antithesis_instrumentation__.Notify(85078)
			}
			__antithesis_instrumentation__.Notify(85076)

			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(85079)
		}
	}
	__antithesis_instrumentation__.Notify(85067)
	return false, err
}
