package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/redact"
)

type ReplicaSnapshotDiff struct {
	LeaseHolder bool
	Key         roachpb.Key
	Timestamp   hlc.Timestamp
	Value       []byte
}

type ReplicaSnapshotDiffSlice []ReplicaSnapshotDiff

func (rsds ReplicaSnapshotDiffSlice) SafeFormat(buf redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(117087)
	buf.Printf("--- leaseholder\n+++ follower\n")
	for _, d := range rsds {
		__antithesis_instrumentation__.Notify(117088)
		prefix := redact.SafeString("+")
		if d.LeaseHolder {
			__antithesis_instrumentation__.Notify(117090)

			prefix = redact.SafeString("-")
		} else {
			__antithesis_instrumentation__.Notify(117091)
		}
		__antithesis_instrumentation__.Notify(117089)
		const format = `%s%s %s
%s    ts:%s
%s    value:%s
%s    raw mvcc_key/value: %x %x
`
		mvccKey := storage.MVCCKey{Key: d.Key, Timestamp: d.Timestamp}
		buf.Printf(format,
			prefix, d.Timestamp, d.Key,
			prefix, d.Timestamp.GoTime(),
			prefix, SprintMVCCKeyValue(storage.MVCCKeyValue{Key: mvccKey, Value: d.Value}, false),
			prefix, storage.EncodeMVCCKey(mvccKey), d.Value)
	}
}

func (rsds ReplicaSnapshotDiffSlice) String() string {
	__antithesis_instrumentation__.Notify(117092)
	return redact.StringWithoutMarkers(rsds)
}

func diffRange(l, r *roachpb.RaftSnapshotData) ReplicaSnapshotDiffSlice {
	__antithesis_instrumentation__.Notify(117093)
	if l == nil || func() bool {
		__antithesis_instrumentation__.Notify(117096)
		return r == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(117097)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(117098)
	}
	__antithesis_instrumentation__.Notify(117094)
	var diff []ReplicaSnapshotDiff
	i, j := 0, 0
	for {
		__antithesis_instrumentation__.Notify(117099)
		var e, v roachpb.RaftSnapshotData_KeyValue
		if i < len(l.KV) {
			__antithesis_instrumentation__.Notify(117105)
			e = l.KV[i]
		} else {
			__antithesis_instrumentation__.Notify(117106)
		}
		__antithesis_instrumentation__.Notify(117100)
		if j < len(r.KV) {
			__antithesis_instrumentation__.Notify(117107)
			v = r.KV[j]
		} else {
			__antithesis_instrumentation__.Notify(117108)
		}
		__antithesis_instrumentation__.Notify(117101)

		addLeaseHolder := func() {
			__antithesis_instrumentation__.Notify(117109)
			diff = append(diff, ReplicaSnapshotDiff{LeaseHolder: true, Key: e.Key, Timestamp: e.Timestamp, Value: e.Value})
			i++
		}
		__antithesis_instrumentation__.Notify(117102)
		addReplica := func() {
			__antithesis_instrumentation__.Notify(117110)
			diff = append(diff, ReplicaSnapshotDiff{LeaseHolder: false, Key: v.Key, Timestamp: v.Timestamp, Value: v.Value})
			j++
		}
		__antithesis_instrumentation__.Notify(117103)

		var comp int

		if e.Key == nil {
			__antithesis_instrumentation__.Notify(117111)
			if v.Key == nil {
				__antithesis_instrumentation__.Notify(117112)

				break
			} else {
				__antithesis_instrumentation__.Notify(117113)
				comp = 1
			}
		} else {
			__antithesis_instrumentation__.Notify(117114)

			if v.Key == nil {
				__antithesis_instrumentation__.Notify(117115)
				comp = -1
			} else {
				__antithesis_instrumentation__.Notify(117116)

				comp = bytes.Compare(e.Key, v.Key)
			}
		}
		__antithesis_instrumentation__.Notify(117104)
		switch comp {
		case -1:
			__antithesis_instrumentation__.Notify(117117)
			addLeaseHolder()

		case 0:
			__antithesis_instrumentation__.Notify(117118)

			if !e.Timestamp.EqOrdering(v.Timestamp) {
				__antithesis_instrumentation__.Notify(117121)
				if e.Timestamp.IsEmpty() {
					__antithesis_instrumentation__.Notify(117122)
					addLeaseHolder()
				} else {
					__antithesis_instrumentation__.Notify(117123)
					if v.Timestamp.IsEmpty() {
						__antithesis_instrumentation__.Notify(117124)
						addReplica()
					} else {
						__antithesis_instrumentation__.Notify(117125)
						if v.Timestamp.Less(e.Timestamp) {
							__antithesis_instrumentation__.Notify(117126)
							addLeaseHolder()
						} else {
							__antithesis_instrumentation__.Notify(117127)
							addReplica()
						}
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(117128)
				if !bytes.Equal(e.Value, v.Value) {
					__antithesis_instrumentation__.Notify(117129)
					addLeaseHolder()
					addReplica()
				} else {
					__antithesis_instrumentation__.Notify(117130)

					i++
					j++
				}
			}

		case 1:
			__antithesis_instrumentation__.Notify(117119)
			addReplica()
		default:
			__antithesis_instrumentation__.Notify(117120)

		}
	}
	__antithesis_instrumentation__.Notify(117095)
	return diff
}
