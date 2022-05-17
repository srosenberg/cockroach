package sessiondata

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

type SequenceState struct {
	mu struct {
		syncutil.Mutex

		latestValues map[uint32]int64

		lastSequenceIncremented uint32
	}
}

func NewSequenceState() *SequenceState {
	__antithesis_instrumentation__.Notify(617893)
	ss := SequenceState{}
	ss.mu.latestValues = make(map[uint32]int64)
	return &ss
}

func (ss *SequenceState) nextValEverCalledLocked() bool {
	__antithesis_instrumentation__.Notify(617894)
	return len(ss.mu.latestValues) > 0
}

func (ss *SequenceState) RecordValue(seqID uint32, val int64) {
	__antithesis_instrumentation__.Notify(617895)
	ss.mu.Lock()
	ss.mu.lastSequenceIncremented = seqID
	ss.mu.latestValues[seqID] = val
	ss.mu.Unlock()
}

func (ss *SequenceState) SetLastSequenceIncremented(seqID uint32) {
	__antithesis_instrumentation__.Notify(617896)
	ss.mu.Lock()
	ss.mu.lastSequenceIncremented = seqID
	ss.mu.Unlock()
}

func (ss *SequenceState) GetLastValue() (int64, error) {
	__antithesis_instrumentation__.Notify(617897)
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if !ss.nextValEverCalledLocked() {
		__antithesis_instrumentation__.Notify(617899)
		return 0, pgerror.New(
			pgcode.ObjectNotInPrerequisiteState, "lastval is not yet defined in this session")
	} else {
		__antithesis_instrumentation__.Notify(617900)
	}
	__antithesis_instrumentation__.Notify(617898)

	return ss.mu.latestValues[ss.mu.lastSequenceIncremented], nil
}

func (ss *SequenceState) GetLastValueByID(seqID uint32) (int64, bool) {
	__antithesis_instrumentation__.Notify(617901)
	ss.mu.Lock()
	defer ss.mu.Unlock()

	val, ok := ss.mu.latestValues[seqID]
	return val, ok
}

func (ss *SequenceState) Export() (map[uint32]int64, uint32) {
	__antithesis_instrumentation__.Notify(617902)
	ss.mu.Lock()
	defer ss.mu.Unlock()
	var res map[uint32]int64
	if len(ss.mu.latestValues) > 0 {
		__antithesis_instrumentation__.Notify(617904)
		res = make(map[uint32]int64, len(ss.mu.latestValues))
		for k, v := range ss.mu.latestValues {
			__antithesis_instrumentation__.Notify(617905)
			res[k] = v
		}
	} else {
		__antithesis_instrumentation__.Notify(617906)
	}
	__antithesis_instrumentation__.Notify(617903)
	return res, ss.mu.lastSequenceIncremented
}
