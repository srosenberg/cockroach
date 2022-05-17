package paramparse

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

type UniqueConstraintParamContext struct {
	IsPrimaryKey bool
	IsSharded    bool
}

func ValidateUniqueConstraintParams(
	params tree.StorageParams, ctx UniqueConstraintParamContext,
) error {
	__antithesis_instrumentation__.Notify(552284)

	for _, param := range params {
		__antithesis_instrumentation__.Notify(552286)
		switch param.Key {
		case `bucket_count`:
			__antithesis_instrumentation__.Notify(552287)
			if ctx.IsSharded {
				__antithesis_instrumentation__.Notify(552291)
				continue
			} else {
				__antithesis_instrumentation__.Notify(552292)
			}
			__antithesis_instrumentation__.Notify(552288)
			return pgerror.New(
				pgcode.InvalidParameterValue,
				`"bucket_count" storage param should only be set with "USING HASH" for hash sharded index`,
			)
		default:
			__antithesis_instrumentation__.Notify(552289)
			if ctx.IsPrimaryKey {
				__antithesis_instrumentation__.Notify(552293)
				return pgerror.Newf(pgcode.InvalidParameterValue, "invalid storage param %q on primary key", params[0].Key)
			} else {
				__antithesis_instrumentation__.Notify(552294)
			}
			__antithesis_instrumentation__.Notify(552290)
			return pgerror.Newf(pgcode.InvalidParameterValue, "invalid storage param %q on unique index", params[0].Key)
		}
	}
	__antithesis_instrumentation__.Notify(552285)
	return nil
}
