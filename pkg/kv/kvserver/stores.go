package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/gossip"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/closedts/ctpb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
)

type Stores struct {
	log.AmbientContext
	clock    *hlc.Clock
	storeMap syncutil.IntMap

	mu struct {
		syncutil.Mutex
		biLatestTS hlc.Timestamp
		latestBI   *gossip.BootstrapInfo
	}
}

var _ kv.Sender = &Stores{}
var _ gossip.Storage = &Stores{}

func NewStores(ambient log.AmbientContext, clock *hlc.Clock) *Stores {
	__antithesis_instrumentation__.Notify(126419)
	return &Stores{
		AmbientContext: ambient,
		clock:          clock,
	}
}

func (ls *Stores) IsMeta1Leaseholder(ctx context.Context, now hlc.ClockTimestamp) (bool, error) {
	__antithesis_instrumentation__.Notify(126420)
	repl, _, err := ls.GetReplicaForRangeID(ctx, 1)
	if roachpb.IsRangeNotFoundError(err) {
		__antithesis_instrumentation__.Notify(126423)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(126424)
	}
	__antithesis_instrumentation__.Notify(126421)
	if err != nil {
		__antithesis_instrumentation__.Notify(126425)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(126426)
	}
	__antithesis_instrumentation__.Notify(126422)
	return repl.OwnsValidLease(ctx, now), nil
}

func (ls *Stores) GetStoreCount() int {
	__antithesis_instrumentation__.Notify(126427)
	var count int
	ls.storeMap.Range(func(_ int64, _ unsafe.Pointer) bool {
		__antithesis_instrumentation__.Notify(126429)
		count++
		return true
	})
	__antithesis_instrumentation__.Notify(126428)
	return count
}

func (ls *Stores) HasStore(storeID roachpb.StoreID) bool {
	__antithesis_instrumentation__.Notify(126430)
	_, ok := ls.storeMap.Load(int64(storeID))
	return ok
}

func (ls *Stores) GetStore(storeID roachpb.StoreID) (*Store, error) {
	__antithesis_instrumentation__.Notify(126431)
	if value, ok := ls.storeMap.Load(int64(storeID)); ok {
		__antithesis_instrumentation__.Notify(126433)
		return (*Store)(value), nil
	} else {
		__antithesis_instrumentation__.Notify(126434)
	}
	__antithesis_instrumentation__.Notify(126432)
	return nil, roachpb.NewStoreNotFoundError(storeID)
}

func (ls *Stores) AddStore(s *Store) {
	__antithesis_instrumentation__.Notify(126435)
	if _, loaded := ls.storeMap.LoadOrStore(int64(s.Ident.StoreID), unsafe.Pointer(s)); loaded {
		__antithesis_instrumentation__.Notify(126437)
		panic(fmt.Sprintf("cannot add store twice: %+v", s.Ident))
	} else {
		__antithesis_instrumentation__.Notify(126438)
	}
	__antithesis_instrumentation__.Notify(126436)

	ls.mu.Lock()
	defer ls.mu.Unlock()
	if !ls.mu.biLatestTS.IsEmpty() {
		__antithesis_instrumentation__.Notify(126439)
		if err := ls.updateBootstrapInfoLocked(ls.mu.latestBI); err != nil {
			__antithesis_instrumentation__.Notify(126440)
			ctx := ls.AnnotateCtx(context.TODO())
			log.Errorf(ctx, "failed to update bootstrap info on newly added store: %+v", err)
		} else {
			__antithesis_instrumentation__.Notify(126441)
		}
	} else {
		__antithesis_instrumentation__.Notify(126442)
	}
}

func (ls *Stores) RemoveStore(s *Store) {
	__antithesis_instrumentation__.Notify(126443)
	ls.storeMap.Delete(int64(s.Ident.StoreID))
}

func (ls *Stores) ForwardSideTransportClosedTimestampForRange(
	ctx context.Context, rangeID roachpb.RangeID, closedTS hlc.Timestamp, lai ctpb.LAI,
) {
	__antithesis_instrumentation__.Notify(126444)
	if err := ls.VisitStores(func(s *Store) error {
		__antithesis_instrumentation__.Notify(126445)
		r := s.GetReplicaIfExists(rangeID)
		if r != nil {
			__antithesis_instrumentation__.Notify(126447)
			r.ForwardSideTransportClosedTimestamp(ctx, closedTS, lai)
		} else {
			__antithesis_instrumentation__.Notify(126448)
		}
		__antithesis_instrumentation__.Notify(126446)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(126449)
		log.Fatalf(ctx, "unexpected error: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(126450)
	}
}

func (ls *Stores) VisitStores(visitor func(s *Store) error) error {
	__antithesis_instrumentation__.Notify(126451)
	var err error
	ls.storeMap.Range(func(k int64, v unsafe.Pointer) bool {
		__antithesis_instrumentation__.Notify(126453)
		err = visitor((*Store)(v))
		return err == nil
	})
	__antithesis_instrumentation__.Notify(126452)
	return err
}

func (ls *Stores) GetReplicaForRangeID(
	ctx context.Context, rangeID roachpb.RangeID,
) (*Replica, *Store, error) {
	__antithesis_instrumentation__.Notify(126454)
	var replica *Replica
	var store *Store
	if err := ls.VisitStores(func(s *Store) error {
		__antithesis_instrumentation__.Notify(126457)
		r := s.GetReplicaIfExists(rangeID)
		if r != nil {
			__antithesis_instrumentation__.Notify(126459)
			replica, store = r, s
		} else {
			__antithesis_instrumentation__.Notify(126460)
		}
		__antithesis_instrumentation__.Notify(126458)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(126461)
		log.Fatalf(ctx, "unexpected error: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(126462)
	}
	__antithesis_instrumentation__.Notify(126455)
	if replica == nil {
		__antithesis_instrumentation__.Notify(126463)
		return nil, nil, roachpb.NewRangeNotFoundError(rangeID, 0)
	} else {
		__antithesis_instrumentation__.Notify(126464)
	}
	__antithesis_instrumentation__.Notify(126456)
	return replica, store, nil
}

func (ls *Stores) Send(
	ctx context.Context, ba roachpb.BatchRequest,
) (*roachpb.BatchResponse, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(126465)
	if err := ba.ValidateForEvaluation(); err != nil {
		__antithesis_instrumentation__.Notify(126469)
		log.Fatalf(ctx, "invalid batch (%s): %s", ba, err)
	} else {
		__antithesis_instrumentation__.Notify(126470)
	}
	__antithesis_instrumentation__.Notify(126466)

	store, err := ls.GetStore(ba.Replica.StoreID)
	if err != nil {
		__antithesis_instrumentation__.Notify(126471)
		return nil, roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(126472)
	}
	__antithesis_instrumentation__.Notify(126467)

	br, pErr := store.Send(ctx, ba)
	if br != nil && func() bool {
		__antithesis_instrumentation__.Notify(126473)
		return br.Error != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(126474)
		panic(roachpb.ErrorUnexpectedlySet(store, br))
	} else {
		__antithesis_instrumentation__.Notify(126475)
	}
	__antithesis_instrumentation__.Notify(126468)
	return br, pErr
}

func (ls *Stores) RangeFeed(
	args *roachpb.RangeFeedRequest, stream roachpb.Internal_RangeFeedServer,
) *roachpb.Error {
	__antithesis_instrumentation__.Notify(126476)
	ctx := stream.Context()
	if args.RangeID == 0 {
		__antithesis_instrumentation__.Notify(126479)
		log.Fatal(ctx, "rangefeed request missing range ID")
	} else {
		__antithesis_instrumentation__.Notify(126480)
		if args.Replica.StoreID == 0 {
			__antithesis_instrumentation__.Notify(126481)
			log.Fatal(ctx, "rangefeed request missing store ID")
		} else {
			__antithesis_instrumentation__.Notify(126482)
		}
	}
	__antithesis_instrumentation__.Notify(126477)

	store, err := ls.GetStore(args.Replica.StoreID)
	if err != nil {
		__antithesis_instrumentation__.Notify(126483)
		return roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(126484)
	}
	__antithesis_instrumentation__.Notify(126478)

	return store.RangeFeed(args, stream)
}

func (ls *Stores) ReadBootstrapInfo(bi *gossip.BootstrapInfo) error {
	__antithesis_instrumentation__.Notify(126485)
	var latestTS hlc.Timestamp

	ctx := ls.AnnotateCtx(context.TODO())
	var err error

	ls.storeMap.Range(func(k int64, v unsafe.Pointer) bool {
		__antithesis_instrumentation__.Notify(126488)
		s := (*Store)(v)
		var storeBI gossip.BootstrapInfo
		var ok bool
		ok, err = storage.MVCCGetProto(ctx, s.engine, keys.StoreGossipKey(), hlc.Timestamp{}, &storeBI,
			storage.MVCCGetOptions{})
		if err != nil {
			__antithesis_instrumentation__.Notify(126491)
			return false
		} else {
			__antithesis_instrumentation__.Notify(126492)
		}
		__antithesis_instrumentation__.Notify(126489)
		if ok && func() bool {
			__antithesis_instrumentation__.Notify(126493)
			return latestTS.Less(storeBI.Timestamp) == true
		}() == true {
			__antithesis_instrumentation__.Notify(126494)
			latestTS = storeBI.Timestamp
			*bi = storeBI
		} else {
			__antithesis_instrumentation__.Notify(126495)
		}
		__antithesis_instrumentation__.Notify(126490)
		return true
	})
	__antithesis_instrumentation__.Notify(126486)
	if err != nil {
		__antithesis_instrumentation__.Notify(126496)
		return err
	} else {
		__antithesis_instrumentation__.Notify(126497)
	}
	__antithesis_instrumentation__.Notify(126487)
	log.Infof(ctx, "read %d node addresses from persistent storage", len(bi.Addresses))

	ls.mu.Lock()
	defer ls.mu.Unlock()
	return ls.updateBootstrapInfoLocked(bi)
}

func (ls *Stores) WriteBootstrapInfo(bi *gossip.BootstrapInfo) error {
	__antithesis_instrumentation__.Notify(126498)
	ls.mu.Lock()
	defer ls.mu.Unlock()
	bi.Timestamp = ls.clock.Now()
	if err := ls.updateBootstrapInfoLocked(bi); err != nil {
		__antithesis_instrumentation__.Notify(126500)
		return err
	} else {
		__antithesis_instrumentation__.Notify(126501)
	}
	__antithesis_instrumentation__.Notify(126499)
	ctx := ls.AnnotateCtx(context.TODO())
	log.Infof(ctx, "wrote %d node addresses to persistent storage", len(bi.Addresses))
	return nil
}

func (ls *Stores) updateBootstrapInfoLocked(bi *gossip.BootstrapInfo) error {
	__antithesis_instrumentation__.Notify(126502)
	if bi.Timestamp.Less(ls.mu.biLatestTS) {
		__antithesis_instrumentation__.Notify(126505)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(126506)
	}
	__antithesis_instrumentation__.Notify(126503)
	ctx := ls.AnnotateCtx(context.TODO())

	ls.mu.biLatestTS = bi.Timestamp
	ls.mu.latestBI = protoutil.Clone(bi).(*gossip.BootstrapInfo)

	var err error
	ls.storeMap.Range(func(k int64, v unsafe.Pointer) bool {
		__antithesis_instrumentation__.Notify(126507)
		s := (*Store)(v)
		err = storage.MVCCPutProto(ctx, s.engine, nil, keys.StoreGossipKey(), hlc.Timestamp{}, nil, bi)
		return err == nil
	})
	__antithesis_instrumentation__.Notify(126504)
	return err
}

func WriteClusterVersionToEngines(
	ctx context.Context, engines []storage.Engine, cv clusterversion.ClusterVersion,
) error {
	__antithesis_instrumentation__.Notify(126508)
	for _, eng := range engines {
		__antithesis_instrumentation__.Notify(126510)
		if err := WriteClusterVersion(ctx, eng, cv); err != nil {
			__antithesis_instrumentation__.Notify(126511)
			return errors.Wrapf(err, "error writing version to engine %s", eng)
		} else {
			__antithesis_instrumentation__.Notify(126512)
		}
	}
	__antithesis_instrumentation__.Notify(126509)
	return nil
}

func SynthesizeClusterVersionFromEngines(
	ctx context.Context,
	engines []storage.Engine,
	binaryVersion, binaryMinSupportedVersion roachpb.Version,
) (clusterversion.ClusterVersion, error) {
	__antithesis_instrumentation__.Notify(126513)

	type originVersion struct {
		roachpb.Version
		origin string
	}

	maxPossibleVersion := roachpb.Version{Major: 999999}
	minStoreVersion := originVersion{
		Version: maxPossibleVersion,
		origin:  "(no store)",
	}

	for _, eng := range engines {
		__antithesis_instrumentation__.Notify(126517)
		eng := eng.(storage.Reader)
		var cv clusterversion.ClusterVersion
		cv, err := ReadClusterVersion(ctx, eng)
		if err != nil {
			__antithesis_instrumentation__.Notify(126521)
			return clusterversion.ClusterVersion{}, err
		} else {
			__antithesis_instrumentation__.Notify(126522)
		}
		__antithesis_instrumentation__.Notify(126518)
		if cv.Version == (roachpb.Version{}) {
			__antithesis_instrumentation__.Notify(126523)

			cv.Version = binaryMinSupportedVersion
		} else {
			__antithesis_instrumentation__.Notify(126524)
		}
		__antithesis_instrumentation__.Notify(126519)

		if binaryVersion.Less(cv.Version) {
			__antithesis_instrumentation__.Notify(126525)
			return clusterversion.ClusterVersion{}, errors.Errorf(
				"cockroach version v%s is incompatible with data in store %s; use version v%s or later",
				binaryVersion, eng, cv.Version)
		} else {
			__antithesis_instrumentation__.Notify(126526)
		}
		__antithesis_instrumentation__.Notify(126520)

		if cv.Version.Less(minStoreVersion.Version) {
			__antithesis_instrumentation__.Notify(126527)
			minStoreVersion.Version = cv.Version
			minStoreVersion.origin = fmt.Sprint(eng)
		} else {
			__antithesis_instrumentation__.Notify(126528)
		}
	}
	__antithesis_instrumentation__.Notify(126514)

	if minStoreVersion.Version == maxPossibleVersion {
		__antithesis_instrumentation__.Notify(126529)
		minStoreVersion.Version = binaryMinSupportedVersion
	} else {
		__antithesis_instrumentation__.Notify(126530)
	}
	__antithesis_instrumentation__.Notify(126515)

	cv := clusterversion.ClusterVersion{
		Version: minStoreVersion.Version,
	}
	log.Eventf(ctx, "read ClusterVersion %+v", cv)

	if minStoreVersion.Version.Less(binaryMinSupportedVersion) {
		__antithesis_instrumentation__.Notify(126531)
		return clusterversion.ClusterVersion{}, errors.Errorf("store %s, last used with cockroach version v%s, "+
			"is too old for running version v%s (which requires data from v%s or later)",
			minStoreVersion.origin, minStoreVersion.Version, binaryVersion, binaryMinSupportedVersion)
	} else {
		__antithesis_instrumentation__.Notify(126532)
	}
	__antithesis_instrumentation__.Notify(126516)
	return cv, nil
}

func (ls *Stores) engines() []storage.Engine {
	__antithesis_instrumentation__.Notify(126533)
	var engines []storage.Engine
	ls.storeMap.Range(func(_ int64, v unsafe.Pointer) bool {
		__antithesis_instrumentation__.Notify(126535)
		engines = append(engines, (*Store)(v).Engine())
		return true
	})
	__antithesis_instrumentation__.Notify(126534)
	return engines
}
