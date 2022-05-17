package scmutationexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scop"
)

func NewMutationVisitor(
	s MutationVisitorStateUpdater, nr NameResolver, sd SyntheticDescriptors, clock Clock,
) scop.MutationVisitor {
	__antithesis_instrumentation__.Notify(582365)
	return &visitor{
		nr:    nr,
		sd:    sd,
		s:     s,
		clock: clock,
	}
}

var _ scop.MutationVisitor = (*visitor)(nil)

type visitor struct {
	clock Clock
	nr    NameResolver
	sd    SyntheticDescriptors
	s     MutationVisitorStateUpdater
}

func (m *visitor) NotImplemented(_ context.Context, _ scop.NotImplemented) error {
	__antithesis_instrumentation__.Notify(582366)
	return nil
}
