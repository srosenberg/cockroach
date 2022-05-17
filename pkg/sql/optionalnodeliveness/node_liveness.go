package optionalnodeliveness

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness/livenesspb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil"
)

type Interface interface {
	Self() (livenesspb.Liveness, bool)
	GetLivenesses() []livenesspb.Liveness
	GetLivenessesFromKV(ctx context.Context) ([]livenesspb.Liveness, error)
	IsAvailable(roachpb.NodeID) bool
	IsAvailableNotDraining(roachpb.NodeID) bool
	IsLive(roachpb.NodeID) (bool, error)
}

type Container struct {
	w errorutil.TenantSQLDeprecatedWrapper
}

func MakeContainer(nl Interface) Container {
	__antithesis_instrumentation__.Notify(551920)
	return Container{
		w: errorutil.MakeTenantSQLDeprecatedWrapper(nl, nl != nil),
	}
}

func (nl *Container) OptionalErr(issue int) (Interface, error) {
	__antithesis_instrumentation__.Notify(551921)
	v, err := nl.w.OptionalErr(issue)
	if err != nil {
		__antithesis_instrumentation__.Notify(551923)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551924)
	}
	__antithesis_instrumentation__.Notify(551922)
	return v.(Interface), nil
}

var _ = (*Container)(nil).OptionalErr

func (nl *Container) Optional(issue int) (Interface, bool) {
	__antithesis_instrumentation__.Notify(551925)
	v, ok := nl.w.Optional()
	if !ok {
		__antithesis_instrumentation__.Notify(551927)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(551928)
	}
	__antithesis_instrumentation__.Notify(551926)
	return v.(Interface), true
}
