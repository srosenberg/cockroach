package sessiondata

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"net"
	"time"

	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil/pgdate"
	"github.com/cockroachdb/errors"
)

type SessionData struct {
	sessiondatapb.SessionData

	sessiondatapb.LocalOnlySessionData

	LocalUnmigratableSessionData

	Location *time.Location

	SearchPath SearchPath

	SequenceState *SequenceState
}

func (s *SessionData) Clone() *SessionData {
	__antithesis_instrumentation__.Notify(617907)
	var newCustomOptions map[string]string
	if len(s.CustomOptions) > 0 {
		__antithesis_instrumentation__.Notify(617909)
		newCustomOptions = make(map[string]string, len(s.CustomOptions))
		for k, v := range s.CustomOptions {
			__antithesis_instrumentation__.Notify(617910)
			newCustomOptions[k] = v
		}
	} else {
		__antithesis_instrumentation__.Notify(617911)
	}
	__antithesis_instrumentation__.Notify(617908)

	ret := *s
	ret.CustomOptions = newCustomOptions
	return &ret
}

func MarshalNonLocal(sd *SessionData, proto *sessiondatapb.SessionData) {
	__antithesis_instrumentation__.Notify(617912)
	proto.Location = sd.GetLocation().String()

	proto.SearchPath = sd.SearchPath.GetPathArray()
	proto.TemporarySchemaName = sd.SearchPath.GetTemporarySchemaName()

	latestValues, lastIncremented := sd.SequenceState.Export()
	if len(latestValues) > 0 {
		__antithesis_instrumentation__.Notify(617913)
		proto.SeqState.LastSeqIncremented = lastIncremented
		for seqID, latestVal := range latestValues {
			__antithesis_instrumentation__.Notify(617914)
			proto.SeqState.Seqs = append(proto.SeqState.Seqs,
				&sessiondatapb.SequenceState_Seq{SeqID: seqID, LatestVal: latestVal},
			)
		}
	} else {
		__antithesis_instrumentation__.Notify(617915)
	}
}

func UnmarshalNonLocal(proto sessiondatapb.SessionData) (*SessionData, error) {
	__antithesis_instrumentation__.Notify(617916)
	location, err := timeutil.TimeZoneStringToLocation(
		proto.Location,
		timeutil.TimeZoneStringToLocationISO8601Standard,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(617920)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(617921)
	}
	__antithesis_instrumentation__.Notify(617917)
	seqState := NewSequenceState()
	var haveSequences bool
	for _, seq := range proto.SeqState.Seqs {
		__antithesis_instrumentation__.Notify(617922)
		seqState.RecordValue(seq.SeqID, seq.LatestVal)
		haveSequences = true
	}
	__antithesis_instrumentation__.Notify(617918)
	if haveSequences {
		__antithesis_instrumentation__.Notify(617923)
		seqState.SetLastSequenceIncremented(proto.SeqState.LastSeqIncremented)
	} else {
		__antithesis_instrumentation__.Notify(617924)
	}
	__antithesis_instrumentation__.Notify(617919)
	return &SessionData{
		SessionData: proto,
		SearchPath: MakeSearchPath(
			proto.SearchPath,
		).WithTemporarySchemaName(
			proto.TemporarySchemaName,
		).WithUserSchemaName(proto.UserProto.Decode().Normalized()),
		SequenceState: seqState,
		Location:      location,
	}, nil
}

func (s *SessionData) GetLocation() *time.Location {
	__antithesis_instrumentation__.Notify(617925)
	if s == nil || func() bool {
		__antithesis_instrumentation__.Notify(617927)
		return s.Location == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(617928)
		return time.UTC
	} else {
		__antithesis_instrumentation__.Notify(617929)
	}
	__antithesis_instrumentation__.Notify(617926)
	return s.Location
}

func (s *SessionData) GetIntervalStyle() duration.IntervalStyle {
	__antithesis_instrumentation__.Notify(617930)
	if s == nil {
		__antithesis_instrumentation__.Notify(617932)
		return duration.IntervalStyle_POSTGRES
	} else {
		__antithesis_instrumentation__.Notify(617933)
	}
	__antithesis_instrumentation__.Notify(617931)
	return s.DataConversionConfig.IntervalStyle
}

func (s *SessionData) GetDateStyle() pgdate.DateStyle {
	__antithesis_instrumentation__.Notify(617934)
	if s == nil {
		__antithesis_instrumentation__.Notify(617936)
		return pgdate.DefaultDateStyle()
	} else {
		__antithesis_instrumentation__.Notify(617937)
	}
	__antithesis_instrumentation__.Notify(617935)
	return s.DataConversionConfig.DateStyle
}

func (s *SessionData) SessionUser() security.SQLUsername {
	__antithesis_instrumentation__.Notify(617938)
	if s.SessionUserProto == "" {
		__antithesis_instrumentation__.Notify(617940)
		return s.User()
	} else {
		__antithesis_instrumentation__.Notify(617941)
	}
	__antithesis_instrumentation__.Notify(617939)
	return s.SessionUserProto.Decode()
}

type LocalUnmigratableSessionData struct {
	RemoteAddr net.Addr

	DatabaseIDToTempSchemaID map[uint32]uint32
}

func (s *SessionData) IsTemporarySchemaID(schemaID uint32) bool {
	__antithesis_instrumentation__.Notify(617942)
	_, exists := s.MaybeGetDatabaseForTemporarySchemaID(schemaID)
	return exists
}

func (s *SessionData) MaybeGetDatabaseForTemporarySchemaID(schemaID uint32) (uint32, bool) {
	__antithesis_instrumentation__.Notify(617943)
	for dbID, tempSchemaID := range s.DatabaseIDToTempSchemaID {
		__antithesis_instrumentation__.Notify(617945)
		if tempSchemaID == schemaID {
			__antithesis_instrumentation__.Notify(617946)
			return dbID, true
		} else {
			__antithesis_instrumentation__.Notify(617947)
		}
	}
	__antithesis_instrumentation__.Notify(617944)
	return 0, false
}

func (s *SessionData) GetTemporarySchemaIDForDB(dbID uint32) (uint32, bool) {
	__antithesis_instrumentation__.Notify(617948)
	schemaID, found := s.DatabaseIDToTempSchemaID[dbID]
	return schemaID, found
}

type Stack struct {
	stack []*SessionData

	base *SessionData
}

func NewStack(firstElem *SessionData) *Stack {
	__antithesis_instrumentation__.Notify(617949)
	return &Stack{stack: []*SessionData{firstElem}, base: firstElem}
}

func (s *Stack) Clone() *Stack {
	__antithesis_instrumentation__.Notify(617950)
	ret := &Stack{
		stack: make([]*SessionData, len(s.stack)),
	}
	for i, st := range s.stack {
		__antithesis_instrumentation__.Notify(617952)
		ret.stack[i] = st.Clone()
	}
	__antithesis_instrumentation__.Notify(617951)
	ret.base = ret.stack[0]
	return ret
}

func (s *Stack) Replace(repl *Stack) {
	__antithesis_instrumentation__.Notify(617953)

	*s = *repl.Clone()
}

func (s *Stack) Top() *SessionData {
	__antithesis_instrumentation__.Notify(617954)
	if len(s.stack) == 0 {
		__antithesis_instrumentation__.Notify(617956)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(617957)
	}
	__antithesis_instrumentation__.Notify(617955)
	return s.stack[len(s.stack)-1]
}

func (s *Stack) Base() *SessionData {
	__antithesis_instrumentation__.Notify(617958)
	return s.base
}

func (s *Stack) Push(elem *SessionData) {
	__antithesis_instrumentation__.Notify(617959)
	s.stack = append(s.stack, elem)
}

func (s *Stack) PushTopClone() {
	__antithesis_instrumentation__.Notify(617960)
	if len(s.stack) == 0 {
		__antithesis_instrumentation__.Notify(617962)
		return
	} else {
		__antithesis_instrumentation__.Notify(617963)
	}
	__antithesis_instrumentation__.Notify(617961)
	sd := s.stack[len(s.stack)-1]
	s.stack = append(s.stack, sd.Clone())
}

func (s *Stack) Pop() error {
	__antithesis_instrumentation__.Notify(617964)
	if len(s.stack) <= 1 {
		__antithesis_instrumentation__.Notify(617966)
		return errors.AssertionFailedf("there must always be at least one element in the SessionData stack")
	} else {
		__antithesis_instrumentation__.Notify(617967)
	}
	__antithesis_instrumentation__.Notify(617965)
	idx := len(s.stack) - 1
	s.stack[idx] = nil
	s.stack = s.stack[:idx]
	return nil
}

func (s *Stack) PopN(n int) error {
	__antithesis_instrumentation__.Notify(617968)
	if len(s.stack)-n <= 0 {
		__antithesis_instrumentation__.Notify(617971)
		return errors.AssertionFailedf("there must always be at least one element in the SessionData stack")
	} else {
		__antithesis_instrumentation__.Notify(617972)
	}
	__antithesis_instrumentation__.Notify(617969)

	for i := 0; i < n; i++ {
		__antithesis_instrumentation__.Notify(617973)
		s.stack[len(s.stack)-1-i] = nil
	}
	__antithesis_instrumentation__.Notify(617970)
	s.stack = s.stack[:len(s.stack)-n]
	return nil
}

func (s *Stack) PopAll() {
	__antithesis_instrumentation__.Notify(617974)

	for i := 1; i < len(s.stack); i++ {
		__antithesis_instrumentation__.Notify(617976)
		s.stack[i] = nil
	}
	__antithesis_instrumentation__.Notify(617975)
	s.stack = s.stack[:1]
}

func (s *Stack) Elems() []*SessionData {
	__antithesis_instrumentation__.Notify(617977)
	return s.stack
}
