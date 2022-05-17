package gossip

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

type stringMatcher interface {
	MatchString(string) bool
}

type allMatcher struct{}

func (allMatcher) MatchString(string) bool {
	__antithesis_instrumentation__.Notify(67788)
	return true
}

type callback struct {
	matcher   stringMatcher
	method    Callback
	redundant bool
}

type infoStore struct {
	log.AmbientContext

	nodeID  *base.NodeIDContainer
	stopper *stop.Stopper

	Infos           infoMap             `json:"infos,omitempty"`
	NodeAddr        util.UnresolvedAddr `json:"-"`
	highWaterStamps map[roachpb.NodeID]int64
	callbacks       []*callback

	callbackWorkMu syncutil.Mutex
	callbackWork   []func()
	callbackCh     chan struct{}
}

var monoTime struct {
	syncutil.Mutex
	last int64
}

var errNotFresh = errors.New("info not fresh")

func monotonicUnixNano() int64 {
	__antithesis_instrumentation__.Notify(67789)
	monoTime.Lock()
	defer monoTime.Unlock()

	now := timeutil.Now().UnixNano()
	if now <= monoTime.last {
		__antithesis_instrumentation__.Notify(67791)
		now = monoTime.last + 1
	} else {
		__antithesis_instrumentation__.Notify(67792)
	}
	__antithesis_instrumentation__.Notify(67790)
	monoTime.last = now
	return now
}

func ratchetMonotonic(v int64) {
	__antithesis_instrumentation__.Notify(67793)
	monoTime.Lock()
	if monoTime.last < v {
		__antithesis_instrumentation__.Notify(67795)
		monoTime.last = v
	} else {
		__antithesis_instrumentation__.Notify(67796)
	}
	__antithesis_instrumentation__.Notify(67794)
	monoTime.Unlock()
}

func ratchetHighWaterStamp(stamps map[roachpb.NodeID]int64, nodeID roachpb.NodeID, newStamp int64) {
	__antithesis_instrumentation__.Notify(67797)
	if nodeID != 0 && func() bool {
		__antithesis_instrumentation__.Notify(67798)
		return stamps[nodeID] < newStamp == true
	}() == true {
		__antithesis_instrumentation__.Notify(67799)
		stamps[nodeID] = newStamp
	} else {
		__antithesis_instrumentation__.Notify(67800)
	}
}

func mergeHighWaterStamps(dest *map[roachpb.NodeID]int64, src map[roachpb.NodeID]int64) {
	__antithesis_instrumentation__.Notify(67801)
	if *dest == nil {
		__antithesis_instrumentation__.Notify(67803)
		*dest = src
		return
	} else {
		__antithesis_instrumentation__.Notify(67804)
	}
	__antithesis_instrumentation__.Notify(67802)
	for nodeID, newStamp := range src {
		__antithesis_instrumentation__.Notify(67805)
		ratchetHighWaterStamp(*dest, nodeID, newStamp)
	}
}

func (is *infoStore) String() string {
	__antithesis_instrumentation__.Notify(67806)
	var buf strings.Builder
	if infoCount := len(is.Infos); infoCount > 0 {
		__antithesis_instrumentation__.Notify(67809)
		fmt.Fprintf(&buf, "infostore with %d info(s): ", infoCount)
	} else {
		__antithesis_instrumentation__.Notify(67810)
		return "infostore (empty)"
	}
	__antithesis_instrumentation__.Notify(67807)

	prepend := ""

	if err := is.visitInfos(func(key string, i *Info) error {
		__antithesis_instrumentation__.Notify(67811)
		fmt.Fprintf(&buf, "%sinfo %q: %+v", prepend, key, i.Value)
		prepend = ", "
		return nil
	}, false); err != nil {
		__antithesis_instrumentation__.Notify(67812)

		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(67813)
	}
	__antithesis_instrumentation__.Notify(67808)
	return buf.String()
}

func newInfoStore(
	ambient log.AmbientContext,
	nodeID *base.NodeIDContainer,
	nodeAddr util.UnresolvedAddr,
	stopper *stop.Stopper,
) *infoStore {
	__antithesis_instrumentation__.Notify(67814)
	is := &infoStore{
		AmbientContext:  ambient,
		nodeID:          nodeID,
		stopper:         stopper,
		Infos:           make(infoMap),
		NodeAddr:        nodeAddr,
		highWaterStamps: map[roachpb.NodeID]int64{},
		callbackCh:      make(chan struct{}, 1),
	}

	bgCtx := ambient.AnnotateCtx(context.Background())
	_ = is.stopper.RunAsyncTask(bgCtx, "infostore", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(67816)
		for {
			__antithesis_instrumentation__.Notify(67817)
			for {
				__antithesis_instrumentation__.Notify(67819)
				is.callbackWorkMu.Lock()
				work := is.callbackWork
				is.callbackWork = nil
				is.callbackWorkMu.Unlock()

				if len(work) == 0 {
					__antithesis_instrumentation__.Notify(67821)
					break
				} else {
					__antithesis_instrumentation__.Notify(67822)
				}
				__antithesis_instrumentation__.Notify(67820)
				for _, w := range work {
					__antithesis_instrumentation__.Notify(67823)
					w()
				}
			}
			__antithesis_instrumentation__.Notify(67818)

			select {
			case <-is.callbackCh:
				__antithesis_instrumentation__.Notify(67824)
			case <-is.stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(67825)
				return
			}
		}
	})
	__antithesis_instrumentation__.Notify(67815)
	return is
}

func (is *infoStore) newInfo(val []byte, ttl time.Duration) *Info {
	__antithesis_instrumentation__.Notify(67826)
	nodeID := is.nodeID.Get()
	if nodeID == 0 {
		__antithesis_instrumentation__.Notify(67829)
		panic("gossip infostore's NodeID is 0")
	} else {
		__antithesis_instrumentation__.Notify(67830)
	}
	__antithesis_instrumentation__.Notify(67827)
	now := monotonicUnixNano()
	ttlStamp := now + int64(ttl)
	if ttl == 0 {
		__antithesis_instrumentation__.Notify(67831)
		ttlStamp = math.MaxInt64
	} else {
		__antithesis_instrumentation__.Notify(67832)
	}
	__antithesis_instrumentation__.Notify(67828)
	v := roachpb.MakeValueFromBytesAndTimestamp(val, hlc.Timestamp{WallTime: now})
	return &Info{
		Value:    v,
		TTLStamp: ttlStamp,
		NodeID:   nodeID,
	}
}

func (is *infoStore) getInfo(key string) *Info {
	__antithesis_instrumentation__.Notify(67833)
	if info, ok := is.Infos[key]; ok {
		__antithesis_instrumentation__.Notify(67835)

		if !info.expired(monotonicUnixNano()) {
			__antithesis_instrumentation__.Notify(67836)
			return info
		} else {
			__antithesis_instrumentation__.Notify(67837)
		}
	} else {
		__antithesis_instrumentation__.Notify(67838)
	}
	__antithesis_instrumentation__.Notify(67834)
	return nil
}

func (is *infoStore) addInfo(key string, i *Info) error {
	__antithesis_instrumentation__.Notify(67839)
	if i.NodeID == 0 {
		__antithesis_instrumentation__.Notify(67843)
		panic("gossip info's NodeID is 0")
	} else {
		__antithesis_instrumentation__.Notify(67844)
	}
	__antithesis_instrumentation__.Notify(67840)

	existingInfo, ok := is.Infos[key]
	if ok {
		__antithesis_instrumentation__.Notify(67845)
		iNanos := i.Value.Timestamp.WallTime
		existingNanos := existingInfo.Value.Timestamp.WallTime
		if iNanos < existingNanos || func() bool {
			__antithesis_instrumentation__.Notify(67846)
			return (iNanos == existingNanos && func() bool {
				__antithesis_instrumentation__.Notify(67847)
				return i.Hops >= existingInfo.Hops == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(67848)
			return errNotFresh
		} else {
			__antithesis_instrumentation__.Notify(67849)
		}
	} else {
		__antithesis_instrumentation__.Notify(67850)
	}
	__antithesis_instrumentation__.Notify(67841)
	if i.OrigStamp == 0 {
		__antithesis_instrumentation__.Notify(67851)
		i.Value.InitChecksum([]byte(key))
		i.OrigStamp = monotonicUnixNano()
		if highWaterStamp, ok := is.highWaterStamps[i.NodeID]; ok && func() bool {
			__antithesis_instrumentation__.Notify(67852)
			return highWaterStamp >= i.OrigStamp == true
		}() == true {
			__antithesis_instrumentation__.Notify(67853)

			log.Fatalf(context.Background(),
				"high water stamp %d >= %d", redact.Safe(highWaterStamp), redact.Safe(i.OrigStamp))
		} else {
			__antithesis_instrumentation__.Notify(67854)
		}
	} else {
		__antithesis_instrumentation__.Notify(67855)
	}
	__antithesis_instrumentation__.Notify(67842)

	is.Infos[key] = i

	ratchetHighWaterStamp(is.highWaterStamps, i.NodeID, i.OrigStamp)
	changed := existingInfo == nil || func() bool {
		__antithesis_instrumentation__.Notify(67856)
		return !bytes.Equal(existingInfo.Value.RawBytes, i.Value.RawBytes) == true
	}() == true
	is.processCallbacks(key, i.Value, changed)
	return nil
}

func (is *infoStore) getHighWaterStamps() map[roachpb.NodeID]int64 {
	__antithesis_instrumentation__.Notify(67857)
	copy := make(map[roachpb.NodeID]int64, len(is.highWaterStamps))
	for k, hws := range is.highWaterStamps {
		__antithesis_instrumentation__.Notify(67859)
		copy[k] = hws
	}
	__antithesis_instrumentation__.Notify(67858)
	return copy
}

func (is *infoStore) registerCallback(
	pattern string, method Callback, opts ...CallbackOption,
) func() {
	__antithesis_instrumentation__.Notify(67860)
	var matcher stringMatcher
	if pattern == ".*" {
		__antithesis_instrumentation__.Notify(67864)
		matcher = allMatcher{}
	} else {
		__antithesis_instrumentation__.Notify(67865)
		matcher = regexp.MustCompile(pattern)
	}
	__antithesis_instrumentation__.Notify(67861)
	cb := &callback{matcher: matcher, method: method}
	for _, opt := range opts {
		__antithesis_instrumentation__.Notify(67866)
		opt.apply(cb)
	}
	__antithesis_instrumentation__.Notify(67862)

	is.callbacks = append(is.callbacks, cb)
	if err := is.visitInfos(func(key string, i *Info) error {
		__antithesis_instrumentation__.Notify(67867)
		if matcher.MatchString(key) {
			__antithesis_instrumentation__.Notify(67869)
			is.runCallbacks(key, i.Value, method)
		} else {
			__antithesis_instrumentation__.Notify(67870)
		}
		__antithesis_instrumentation__.Notify(67868)
		return nil
	}, true); err != nil {
		__antithesis_instrumentation__.Notify(67871)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(67872)
	}
	__antithesis_instrumentation__.Notify(67863)

	return func() {
		__antithesis_instrumentation__.Notify(67873)
		for i, targetCB := range is.callbacks {
			__antithesis_instrumentation__.Notify(67874)
			if targetCB == cb {
				__antithesis_instrumentation__.Notify(67875)
				numCBs := len(is.callbacks)
				is.callbacks[i] = is.callbacks[numCBs-1]
				is.callbacks[numCBs-1] = nil
				is.callbacks = is.callbacks[:numCBs-1]
				break
			} else {
				__antithesis_instrumentation__.Notify(67876)
			}
		}
	}
}

func (is *infoStore) processCallbacks(key string, content roachpb.Value, changed bool) {
	__antithesis_instrumentation__.Notify(67877)
	var matches []Callback
	for _, cb := range is.callbacks {
		__antithesis_instrumentation__.Notify(67879)
		if (changed || func() bool {
			__antithesis_instrumentation__.Notify(67880)
			return cb.redundant == true
		}() == true) && func() bool {
			__antithesis_instrumentation__.Notify(67881)
			return cb.matcher.MatchString(key) == true
		}() == true {
			__antithesis_instrumentation__.Notify(67882)
			matches = append(matches, cb.method)
		} else {
			__antithesis_instrumentation__.Notify(67883)
		}
	}
	__antithesis_instrumentation__.Notify(67878)
	is.runCallbacks(key, content, matches...)
}

func (is *infoStore) runCallbacks(key string, content roachpb.Value, callbacks ...Callback) {
	__antithesis_instrumentation__.Notify(67884)

	f := func() {
		__antithesis_instrumentation__.Notify(67886)
		for _, method := range callbacks {
			__antithesis_instrumentation__.Notify(67887)
			method(key, content)
		}
	}
	__antithesis_instrumentation__.Notify(67885)
	is.callbackWorkMu.Lock()
	is.callbackWork = append(is.callbackWork, f)
	is.callbackWorkMu.Unlock()

	select {
	case is.callbackCh <- struct{}{}:
		__antithesis_instrumentation__.Notify(67888)
	default:
		__antithesis_instrumentation__.Notify(67889)
	}
}

func (is *infoStore) visitInfos(visitInfo func(string, *Info) error, deleteExpired bool) error {
	__antithesis_instrumentation__.Notify(67890)
	now := monotonicUnixNano()

	if visitInfo != nil {
		__antithesis_instrumentation__.Notify(67892)
		for k, i := range is.Infos {
			__antithesis_instrumentation__.Notify(67893)
			if i.expired(now) {
				__antithesis_instrumentation__.Notify(67895)
				if deleteExpired {
					__antithesis_instrumentation__.Notify(67897)
					delete(is.Infos, k)
				} else {
					__antithesis_instrumentation__.Notify(67898)
				}
				__antithesis_instrumentation__.Notify(67896)
				continue
			} else {
				__antithesis_instrumentation__.Notify(67899)
			}
			__antithesis_instrumentation__.Notify(67894)
			if err := visitInfo(k, i); err != nil {
				__antithesis_instrumentation__.Notify(67900)
				return err
			} else {
				__antithesis_instrumentation__.Notify(67901)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(67902)
	}
	__antithesis_instrumentation__.Notify(67891)

	return nil
}

func (is *infoStore) combine(
	infos map[string]*Info, nodeID roachpb.NodeID,
) (freshCount int, err error) {
	__antithesis_instrumentation__.Notify(67903)
	localNodeID := is.nodeID.Get()
	for key, i := range infos {
		__antithesis_instrumentation__.Notify(67905)
		if i.NodeID == localNodeID {
			__antithesis_instrumentation__.Notify(67908)
			ratchetMonotonic(i.OrigStamp)
		} else {
			__antithesis_instrumentation__.Notify(67909)
		}
		__antithesis_instrumentation__.Notify(67906)

		infoCopy := *i
		infoCopy.Hops++
		infoCopy.PeerID = nodeID
		if infoCopy.OrigStamp == 0 {
			__antithesis_instrumentation__.Notify(67910)
			panic(errors.Errorf("combining info from n%d with 0 original timestamp", nodeID))
		} else {
			__antithesis_instrumentation__.Notify(67911)
		}
		__antithesis_instrumentation__.Notify(67907)

		if addErr := is.addInfo(key, &infoCopy); addErr == nil {
			__antithesis_instrumentation__.Notify(67912)
			freshCount++
		} else {
			__antithesis_instrumentation__.Notify(67913)
			if !errors.Is(addErr, errNotFresh) {
				__antithesis_instrumentation__.Notify(67914)
				err = addErr
			} else {
				__antithesis_instrumentation__.Notify(67915)
			}
		}
	}
	__antithesis_instrumentation__.Notify(67904)
	return
}

func (is *infoStore) delta(highWaterTimestamps map[roachpb.NodeID]int64) map[string]*Info {
	__antithesis_instrumentation__.Notify(67916)
	infos := make(map[string]*Info)

	if err := is.visitInfos(func(key string, i *Info) error {
		__antithesis_instrumentation__.Notify(67918)
		if i.isFresh(highWaterTimestamps[i.NodeID]) {
			__antithesis_instrumentation__.Notify(67920)
			infos[key] = i
		} else {
			__antithesis_instrumentation__.Notify(67921)
		}
		__antithesis_instrumentation__.Notify(67919)
		return nil
	}, true); err != nil {
		__antithesis_instrumentation__.Notify(67922)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(67923)
	}
	__antithesis_instrumentation__.Notify(67917)

	return infos
}

func (is *infoStore) populateMostDistantMarkers(infos map[string]*Info) {
	__antithesis_instrumentation__.Notify(67924)
	if err := is.visitInfos(func(key string, i *Info) error {
		__antithesis_instrumentation__.Notify(67925)
		if IsNodeIDKey(key) {
			__antithesis_instrumentation__.Notify(67927)
			infos[key] = i
		} else {
			__antithesis_instrumentation__.Notify(67928)
		}
		__antithesis_instrumentation__.Notify(67926)
		return nil
	}, true); err != nil {
		__antithesis_instrumentation__.Notify(67929)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(67930)
	}
}

func (is *infoStore) mostDistant(
	hasOutgoingConn func(roachpb.NodeID) bool,
) (roachpb.NodeID, uint32) {
	__antithesis_instrumentation__.Notify(67931)
	localNodeID := is.nodeID.Get()
	var nodeID roachpb.NodeID
	var maxHops uint32
	if err := is.visitInfos(func(key string, i *Info) error {
		__antithesis_instrumentation__.Notify(67933)

		if i.NodeID != localNodeID && func() bool {
			__antithesis_instrumentation__.Notify(67935)
			return i.Hops > maxHops == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(67936)
			return IsNodeIDKey(key) == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(67937)
			return !hasOutgoingConn(i.NodeID) == true
		}() == true {
			__antithesis_instrumentation__.Notify(67938)
			maxHops = i.Hops
			nodeID = i.NodeID
		} else {
			__antithesis_instrumentation__.Notify(67939)
		}
		__antithesis_instrumentation__.Notify(67934)
		return nil
	}, true); err != nil {
		__antithesis_instrumentation__.Notify(67940)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(67941)
	}
	__antithesis_instrumentation__.Notify(67932)
	return nodeID, maxHops
}

func (is *infoStore) leastUseful(nodes nodeSet) roachpb.NodeID {
	__antithesis_instrumentation__.Notify(67942)
	contrib := make(map[roachpb.NodeID]map[roachpb.NodeID]struct{}, nodes.len())
	for node := range nodes.nodes {
		__antithesis_instrumentation__.Notify(67946)
		contrib[node] = map[roachpb.NodeID]struct{}{}
	}
	__antithesis_instrumentation__.Notify(67943)
	if err := is.visitInfos(func(key string, i *Info) error {
		__antithesis_instrumentation__.Notify(67947)
		if _, ok := contrib[i.PeerID]; !ok {
			__antithesis_instrumentation__.Notify(67949)
			contrib[i.PeerID] = map[roachpb.NodeID]struct{}{}
		} else {
			__antithesis_instrumentation__.Notify(67950)
		}
		__antithesis_instrumentation__.Notify(67948)
		contrib[i.PeerID][i.NodeID] = struct{}{}
		return nil
	}, true); err != nil {
		__antithesis_instrumentation__.Notify(67951)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(67952)
	}
	__antithesis_instrumentation__.Notify(67944)

	least := math.MaxInt32
	var leastNode roachpb.NodeID
	for id, m := range contrib {
		__antithesis_instrumentation__.Notify(67953)
		count := len(m)
		if nodes.hasNode(id) {
			__antithesis_instrumentation__.Notify(67954)
			if count < least {
				__antithesis_instrumentation__.Notify(67955)
				least = count
				leastNode = id
			} else {
				__antithesis_instrumentation__.Notify(67956)
			}
		} else {
			__antithesis_instrumentation__.Notify(67957)
		}
	}
	__antithesis_instrumentation__.Notify(67945)
	return leastNode
}
