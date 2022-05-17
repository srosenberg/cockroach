package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/gossip"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/apply"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

func funcName(f interface{}) string {
	__antithesis_instrumentation__.Notify(126736)
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

func TrackRaftProtos() func() []reflect.Type {
	__antithesis_instrumentation__.Notify(126737)

	applyRaftEntryFunc := funcName((*apply.Task).ApplyCommittedEntries)

	allowlist := []string{

		funcName((*gossip.Gossip).AddInfoProto),

		funcName((*Replica).setTombstoneKey),

		funcName((*Replica).tryReproposeWithNewLeaseIndex),
	}

	belowRaftProtos := struct {
		syncutil.Mutex
		inner map[reflect.Type]struct{}
	}{
		inner: make(map[reflect.Type]struct{}),
	}

	protoutil.Interceptor = func(pb protoutil.Message) {
		__antithesis_instrumentation__.Notify(126739)
		t := reflect.TypeOf(pb)

		if meta, ok := pb.(*enginepb.MVCCMetadata); ok && func() bool {
			__antithesis_instrumentation__.Notify(126743)
			return meta.Txn != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(126744)
			protoutil.Interceptor(meta.Txn)
		} else {
			__antithesis_instrumentation__.Notify(126745)
		}
		__antithesis_instrumentation__.Notify(126740)

		belowRaftProtos.Lock()
		_, ok := belowRaftProtos.inner[t]
		belowRaftProtos.Unlock()
		if ok {
			__antithesis_instrumentation__.Notify(126746)
			return
		} else {
			__antithesis_instrumentation__.Notify(126747)
		}
		__antithesis_instrumentation__.Notify(126741)

		var pcs [256]uintptr
		if numCallers := runtime.Callers(0, pcs[:]); numCallers == len(pcs) {
			__antithesis_instrumentation__.Notify(126748)
			panic(fmt.Sprintf("number of callers %d might have exceeded slice size %d", numCallers, len(pcs)))
		} else {
			__antithesis_instrumentation__.Notify(126749)
		}
		__antithesis_instrumentation__.Notify(126742)
		frames := runtime.CallersFrames(pcs[:])
		for {
			__antithesis_instrumentation__.Notify(126750)
			f, more := frames.Next()

			allowlisted := false
			for _, s := range allowlist {
				__antithesis_instrumentation__.Notify(126754)
				if strings.Contains(f.Function, s) {
					__antithesis_instrumentation__.Notify(126755)
					allowlisted = true
					break
				} else {
					__antithesis_instrumentation__.Notify(126756)
				}
			}
			__antithesis_instrumentation__.Notify(126751)
			if allowlisted {
				__antithesis_instrumentation__.Notify(126757)
				break
			} else {
				__antithesis_instrumentation__.Notify(126758)
			}
			__antithesis_instrumentation__.Notify(126752)

			if strings.Contains(f.Function, applyRaftEntryFunc) {
				__antithesis_instrumentation__.Notify(126759)
				belowRaftProtos.Lock()
				belowRaftProtos.inner[t] = struct{}{}
				belowRaftProtos.Unlock()
				break
			} else {
				__antithesis_instrumentation__.Notify(126760)
			}
			__antithesis_instrumentation__.Notify(126753)
			if !more {
				__antithesis_instrumentation__.Notify(126761)
				break
			} else {
				__antithesis_instrumentation__.Notify(126762)
			}
		}
	}
	__antithesis_instrumentation__.Notify(126738)

	return func() []reflect.Type {
		__antithesis_instrumentation__.Notify(126763)
		protoutil.Interceptor = func(_ protoutil.Message) { __antithesis_instrumentation__.Notify(126766) }
		__antithesis_instrumentation__.Notify(126764)

		belowRaftProtos.Lock()
		types := make([]reflect.Type, 0, len(belowRaftProtos.inner))
		for t := range belowRaftProtos.inner {
			__antithesis_instrumentation__.Notify(126767)
			types = append(types, t)
		}
		__antithesis_instrumentation__.Notify(126765)
		belowRaftProtos.Unlock()

		return types
	}
}
