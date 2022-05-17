package kvcoord

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/base"

type ClientTestingKnobs struct {
	TransportFactory TransportFactory

	DontConsiderConnHealth bool

	MaxTxnRefreshAttempts int

	CondenseRefreshSpansFilter func() bool

	LatencyFunc LatencyFunc

	DontReorderReplicas bool

	DisableCommitSanityCheck bool
}

var _ base.ModuleTestingKnobs = &ClientTestingKnobs{}

func (*ClientTestingKnobs) ModuleTestingKnobs() { __antithesis_instrumentation__.Notify(87975) }
