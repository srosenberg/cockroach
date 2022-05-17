package kvserverbase

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil"
)

type StoresIterator interface {
	ForEachStore(func(Store) error) error
}

type Store interface {
	StoreID() roachpb.StoreID

	Enqueue(
		ctx context.Context,
		queue string,
		rangeID roachpb.RangeID,
		skipShouldQueue bool,
	) error

	SetQueueActive(active bool, queue string) error
}

type UnsupportedStoresIterator struct{}

var _ StoresIterator = UnsupportedStoresIterator{}

func (i UnsupportedStoresIterator) ForEachStore(f func(Store) error) error {
	__antithesis_instrumentation__.Notify(101857)
	return errorutil.UnsupportedWithMultiTenancy(errorutil.FeatureNotAvailableToNonSystemTenantsIssue)
}
