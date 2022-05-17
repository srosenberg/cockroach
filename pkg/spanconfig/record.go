package spanconfig

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/errors"
)

type Record struct {
	target Target

	config roachpb.SpanConfig
}

func MakeRecord(target Target, cfg roachpb.SpanConfig) (Record, error) {
	__antithesis_instrumentation__.Notify(240279)
	if target.IsSystemTarget() {
		__antithesis_instrumentation__.Notify(240281)
		if err := cfg.ValidateSystemTargetSpanConfig(); err != nil {
			__antithesis_instrumentation__.Notify(240282)
			return Record{},
				errors.NewAssertionErrorWithWrappedErrf(err, "failed to validate SystemTarget SpanConfig")
		} else {
			__antithesis_instrumentation__.Notify(240283)
		}
	} else {
		__antithesis_instrumentation__.Notify(240284)
	}
	__antithesis_instrumentation__.Notify(240280)
	return Record{target: target, config: cfg}, nil
}

func (r *Record) IsEmpty() bool {
	__antithesis_instrumentation__.Notify(240285)
	return r.target.isEmpty() && func() bool {
		__antithesis_instrumentation__.Notify(240286)
		return r.config.IsEmpty() == true
	}() == true
}

func (r *Record) GetTarget() Target {
	__antithesis_instrumentation__.Notify(240287)
	return r.target
}

func (r *Record) GetConfig() roachpb.SpanConfig {
	__antithesis_instrumentation__.Notify(240288)
	return r.config
}
