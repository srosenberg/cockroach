package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"strings"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/errors"
)

type StoresIterator Stores

var _ kvserverbase.StoresIterator = &StoresIterator{}

func MakeStoresIterator(stores *Stores) *StoresIterator {
	__antithesis_instrumentation__.Notify(126536)
	return (*StoresIterator)(stores)
}

func (s *StoresIterator) ForEachStore(f func(kvserverbase.Store) error) error {
	__antithesis_instrumentation__.Notify(126537)
	var err error
	s.storeMap.Range(func(k int64, v unsafe.Pointer) bool {
		__antithesis_instrumentation__.Notify(126539)
		store := (*Store)(v)

		err = f((*baseStore)(store))
		return err == nil
	})
	__antithesis_instrumentation__.Notify(126538)
	return err
}

type baseStore Store

var _ kvserverbase.Store = &baseStore{}

func (s *baseStore) StoreID() roachpb.StoreID {
	__antithesis_instrumentation__.Notify(126540)
	store := (*Store)(s)
	return store.StoreID()
}

func (s *baseStore) Enqueue(
	ctx context.Context, queue string, rangeID roachpb.RangeID, skipShouldQueue bool,
) error {
	__antithesis_instrumentation__.Notify(126541)
	store := (*Store)(s)
	repl, err := store.GetReplica(rangeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(126545)
		return err
	} else {
		__antithesis_instrumentation__.Notify(126546)
	}
	__antithesis_instrumentation__.Notify(126542)

	_, processErr, enqueueErr := store.ManuallyEnqueue(ctx, queue, repl, skipShouldQueue)
	if processErr != nil {
		__antithesis_instrumentation__.Notify(126547)
		return processErr
	} else {
		__antithesis_instrumentation__.Notify(126548)
	}
	__antithesis_instrumentation__.Notify(126543)
	if enqueueErr != nil {
		__antithesis_instrumentation__.Notify(126549)
		return enqueueErr
	} else {
		__antithesis_instrumentation__.Notify(126550)
	}
	__antithesis_instrumentation__.Notify(126544)
	return nil
}

func (s *baseStore) SetQueueActive(active bool, queue string) error {
	__antithesis_instrumentation__.Notify(126551)
	store := (*Store)(s)
	var kvQueue replicaQueue
	for _, rq := range store.scanner.queues {
		__antithesis_instrumentation__.Notify(126554)
		if strings.EqualFold(rq.Name(), queue) {
			__antithesis_instrumentation__.Notify(126555)
			kvQueue = rq
			break
		} else {
			__antithesis_instrumentation__.Notify(126556)
		}
	}
	__antithesis_instrumentation__.Notify(126552)

	if kvQueue == nil {
		__antithesis_instrumentation__.Notify(126557)
		return errors.Errorf("unknown queue %q", queue)
	} else {
		__antithesis_instrumentation__.Notify(126558)
	}
	__antithesis_instrumentation__.Notify(126553)

	kvQueue.SetDisabled(!active)
	return nil
}
