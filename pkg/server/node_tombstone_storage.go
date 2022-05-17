package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

type nodeTombstoneStorage struct {
	engs []storage.Engine
	mu   struct {
		syncutil.RWMutex

		cache map[roachpb.NodeID]time.Time
	}
}

func (s *nodeTombstoneStorage) key(nodeID roachpb.NodeID) roachpb.Key {
	__antithesis_instrumentation__.Notify(194872)
	return keys.StoreNodeTombstoneKey(nodeID)
}

func (s *nodeTombstoneStorage) IsDecommissioned(
	ctx context.Context, nodeID roachpb.NodeID,
) (time.Time, error) {
	__antithesis_instrumentation__.Notify(194873)
	s.mu.RLock()
	ts, ok := s.mu.cache[nodeID]
	s.mu.RUnlock()
	if ok {
		__antithesis_instrumentation__.Notify(194876)

		return ts, nil
	} else {
		__antithesis_instrumentation__.Notify(194877)
	}
	__antithesis_instrumentation__.Notify(194874)

	k := s.key(nodeID)
	for _, eng := range s.engs {
		__antithesis_instrumentation__.Notify(194878)
		v, _, err := storage.MVCCGet(ctx, eng, k, hlc.Timestamp{}, storage.MVCCGetOptions{})
		if err != nil {
			__antithesis_instrumentation__.Notify(194882)
			return time.Time{}, err
		} else {
			__antithesis_instrumentation__.Notify(194883)
		}
		__antithesis_instrumentation__.Notify(194879)
		if v == nil {
			__antithesis_instrumentation__.Notify(194884)

			continue
		} else {
			__antithesis_instrumentation__.Notify(194885)
		}
		__antithesis_instrumentation__.Notify(194880)
		var tsp hlc.Timestamp
		if err := v.GetProto(&tsp); err != nil {
			__antithesis_instrumentation__.Notify(194886)
			return time.Time{}, err
		} else {
			__antithesis_instrumentation__.Notify(194887)
		}
		__antithesis_instrumentation__.Notify(194881)

		ts := timeutil.Unix(0, tsp.WallTime).UTC()
		s.maybeAddCached(nodeID, ts)
		return ts, nil
	}
	__antithesis_instrumentation__.Notify(194875)

	s.maybeAddCached(nodeID, time.Time{})
	return time.Time{}, nil
}

func (s *nodeTombstoneStorage) maybeAddCached(nodeID roachpb.NodeID, ts time.Time) (updated bool) {
	__antithesis_instrumentation__.Notify(194888)
	s.mu.Lock()
	defer s.mu.Unlock()
	if oldTS, ok := s.mu.cache[nodeID]; !ok || func() bool {
		__antithesis_instrumentation__.Notify(194890)
		return oldTS.IsZero() == true
	}() == true {
		__antithesis_instrumentation__.Notify(194891)
		if s.mu.cache == nil {
			__antithesis_instrumentation__.Notify(194893)
			s.mu.cache = map[roachpb.NodeID]time.Time{}
		} else {
			__antithesis_instrumentation__.Notify(194894)
		}
		__antithesis_instrumentation__.Notify(194892)
		s.mu.cache[nodeID] = ts
		return true
	} else {
		__antithesis_instrumentation__.Notify(194895)
	}
	__antithesis_instrumentation__.Notify(194889)
	return false
}

func (s *nodeTombstoneStorage) SetDecommissioned(
	ctx context.Context, nodeID roachpb.NodeID, ts time.Time,
) error {
	__antithesis_instrumentation__.Notify(194896)
	if len(s.engs) == 0 {
		__antithesis_instrumentation__.Notify(194901)
		return errors.New("no engines configured for nodeTombstoneStorage")
	} else {
		__antithesis_instrumentation__.Notify(194902)
	}
	__antithesis_instrumentation__.Notify(194897)
	if ts.IsZero() {
		__antithesis_instrumentation__.Notify(194903)
		return errors.New("can't mark as decommissioned at timestamp zero")
	} else {
		__antithesis_instrumentation__.Notify(194904)
	}
	__antithesis_instrumentation__.Notify(194898)
	if !s.maybeAddCached(nodeID, ts.UTC()) {
		__antithesis_instrumentation__.Notify(194905)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(194906)
	}
	__antithesis_instrumentation__.Notify(194899)

	k := s.key(nodeID)
	for _, eng := range s.engs {
		__antithesis_instrumentation__.Notify(194907)

		if _, err := kvserver.ReadStoreIdent(ctx, eng); err != nil {
			__antithesis_instrumentation__.Notify(194910)
			if errors.Is(err, &kvserver.NotBootstrappedError{}) {
				__antithesis_instrumentation__.Notify(194912)
				continue
			} else {
				__antithesis_instrumentation__.Notify(194913)
			}
			__antithesis_instrumentation__.Notify(194911)
			return err
		} else {
			__antithesis_instrumentation__.Notify(194914)
		}
		__antithesis_instrumentation__.Notify(194908)
		var v roachpb.Value
		if err := v.SetProto(&hlc.Timestamp{WallTime: ts.UnixNano()}); err != nil {
			__antithesis_instrumentation__.Notify(194915)
			return err
		} else {
			__antithesis_instrumentation__.Notify(194916)
		}
		__antithesis_instrumentation__.Notify(194909)

		if err := storage.MVCCPut(
			ctx, eng, nil, k, hlc.Timestamp{}, v, nil,
		); err != nil {
			__antithesis_instrumentation__.Notify(194917)
			return err
		} else {
			__antithesis_instrumentation__.Notify(194918)
		}
	}
	__antithesis_instrumentation__.Notify(194900)
	return nil
}
