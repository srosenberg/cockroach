package gossiputil

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"sync"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/gossip"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

type StoreGossiper struct {
	g           *gossip.Gossip
	mu          syncutil.Mutex
	cond        *sync.Cond
	storeKeyMap map[string]struct{}
}

func NewStoreGossiper(g *gossip.Gossip) *StoreGossiper {
	__antithesis_instrumentation__.Notify(644234)
	sg := &StoreGossiper{
		g:           g,
		storeKeyMap: make(map[string]struct{}),
	}
	sg.cond = sync.NewCond(&sg.mu)

	g.RegisterCallback(gossip.MakePrefixPattern(gossip.KeyStorePrefix), func(key string, _ roachpb.Value) {
		__antithesis_instrumentation__.Notify(644236)
		sg.mu.Lock()
		defer sg.mu.Unlock()
		delete(sg.storeKeyMap, key)
		sg.cond.Broadcast()
	}, gossip.Redundant)
	__antithesis_instrumentation__.Notify(644235)
	return sg
}

func (sg *StoreGossiper) GossipStores(storeDescs []*roachpb.StoreDescriptor, t *testing.T) {
	__antithesis_instrumentation__.Notify(644237)
	storeIDs := make([]roachpb.StoreID, len(storeDescs))
	for i, store := range storeDescs {
		__antithesis_instrumentation__.Notify(644239)
		storeIDs[i] = store.StoreID
	}
	__antithesis_instrumentation__.Notify(644238)
	sg.GossipWithFunction(storeIDs, func() {
		__antithesis_instrumentation__.Notify(644240)
		for i, storeDesc := range storeDescs {
			__antithesis_instrumentation__.Notify(644241)
			if err := sg.g.AddInfoProto(gossip.MakeStoreKey(storeIDs[i]), storeDesc, 0); err != nil {
				__antithesis_instrumentation__.Notify(644242)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(644243)
			}
		}
	})
}

func (sg *StoreGossiper) GossipWithFunction(storeIDs []roachpb.StoreID, gossipFn func()) {
	__antithesis_instrumentation__.Notify(644244)
	sg.mu.Lock()
	defer sg.mu.Unlock()
	sg.storeKeyMap = make(map[string]struct{})
	for _, storeID := range storeIDs {
		__antithesis_instrumentation__.Notify(644246)
		storeKey := gossip.MakeStoreKey(storeID)
		sg.storeKeyMap[storeKey] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(644245)

	gossipFn()

	for len(sg.storeKeyMap) > 0 {
		__antithesis_instrumentation__.Notify(644247)
		sg.cond.Wait()
	}
}
