package protectedts

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/base"

type TestingKnobs struct {
	DisableProtectedTimestampForMultiTenant bool
}

func (t *TestingKnobs) ModuleTestingKnobs() { __antithesis_instrumentation__.Notify(112210) }

var _ base.ModuleTestingKnobs = (*TestingKnobs)(nil)
