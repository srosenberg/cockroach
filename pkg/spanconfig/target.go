package spanconfig

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"testing"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/require"
)

type Target struct {
	span roachpb.Span

	systemTarget SystemTarget
}

func MakeTarget(t roachpb.SpanConfigTarget) (Target, error) {
	__antithesis_instrumentation__.Notify(242083)
	switch t.Union.(type) {
	case *roachpb.SpanConfigTarget_Span:
		__antithesis_instrumentation__.Notify(242084)
		return MakeSpanTargetFromProto(t)
	case *roachpb.SpanConfigTarget_SystemSpanConfigTarget:
		__antithesis_instrumentation__.Notify(242085)
		systemTarget, err := makeSystemTargetFromProto(t.GetSystemSpanConfigTarget())
		if err != nil {
			__antithesis_instrumentation__.Notify(242088)
			return Target{}, err
		} else {
			__antithesis_instrumentation__.Notify(242089)
		}
		__antithesis_instrumentation__.Notify(242086)
		return MakeTargetFromSystemTarget(systemTarget), nil
	default:
		__antithesis_instrumentation__.Notify(242087)
		return Target{}, errors.AssertionFailedf("unknown type of system target %v", t)
	}
}

func MakeSpanTargetFromProto(spanTarget roachpb.SpanConfigTarget) (Target, error) {
	__antithesis_instrumentation__.Notify(242090)
	if spanTarget.GetSpan() == nil {
		__antithesis_instrumentation__.Notify(242093)
		return Target{}, errors.AssertionFailedf("span config target did not contain a span")
	} else {
		__antithesis_instrumentation__.Notify(242094)
	}
	__antithesis_instrumentation__.Notify(242091)
	if keys.SystemSpanConfigSpan.Overlaps(*spanTarget.GetSpan()) {
		__antithesis_instrumentation__.Notify(242095)
		return Target{}, errors.AssertionFailedf(
			"cannot target spans in reserved system span config keyspace",
		)
	} else {
		__antithesis_instrumentation__.Notify(242096)
	}
	__antithesis_instrumentation__.Notify(242092)
	return MakeTargetFromSpan(*spanTarget.GetSpan()), nil
}

func MakeTargetFromSpan(span roachpb.Span) Target {
	__antithesis_instrumentation__.Notify(242097)
	if keys.SystemSpanConfigSpan.Overlaps(span) {
		__antithesis_instrumentation__.Notify(242099)
		panic("cannot target spans in reserved system span config keyspace")
	} else {
		__antithesis_instrumentation__.Notify(242100)
	}
	__antithesis_instrumentation__.Notify(242098)
	return Target{span: span}
}

func MakeTargetFromSystemTarget(systemTarget SystemTarget) Target {
	__antithesis_instrumentation__.Notify(242101)
	return Target{systemTarget: systemTarget}
}

func (t Target) IsSpanTarget() bool {
	__antithesis_instrumentation__.Notify(242102)
	return !t.span.Equal(roachpb.Span{})
}

func (t Target) GetSpan() roachpb.Span {
	__antithesis_instrumentation__.Notify(242103)
	if !t.IsSpanTarget() {
		__antithesis_instrumentation__.Notify(242105)
		panic("target is not a span target")
	} else {
		__antithesis_instrumentation__.Notify(242106)
	}
	__antithesis_instrumentation__.Notify(242104)
	return t.span
}

func (t Target) IsSystemTarget() bool {
	__antithesis_instrumentation__.Notify(242107)
	return !t.systemTarget.IsEmpty()
}

func (t Target) GetSystemTarget() SystemTarget {
	__antithesis_instrumentation__.Notify(242108)
	if !t.IsSystemTarget() {
		__antithesis_instrumentation__.Notify(242110)
		panic("target is not a system target")
	} else {
		__antithesis_instrumentation__.Notify(242111)
	}
	__antithesis_instrumentation__.Notify(242109)
	return t.systemTarget
}

func (t Target) Encode() roachpb.Span {
	__antithesis_instrumentation__.Notify(242112)
	switch {
	case t.IsSpanTarget():
		__antithesis_instrumentation__.Notify(242113)
		return t.span
	case t.IsSystemTarget():
		__antithesis_instrumentation__.Notify(242114)
		return t.systemTarget.encode()
	default:
		__antithesis_instrumentation__.Notify(242115)
		panic("cannot handle any other type of target")
	}
}

func (t Target) KeyspaceTargeted() roachpb.Span {
	__antithesis_instrumentation__.Notify(242116)
	switch {
	case t.IsSpanTarget():
		__antithesis_instrumentation__.Notify(242117)
		return t.span
	case t.IsSystemTarget():
		__antithesis_instrumentation__.Notify(242118)
		return t.systemTarget.keyspaceTargeted()
	default:
		__antithesis_instrumentation__.Notify(242119)
		panic("cannot handle any other type of target")
	}
}

func (t Target) Less(o Target) bool {
	__antithesis_instrumentation__.Notify(242120)

	if t.IsSystemTarget() && func() bool {
		__antithesis_instrumentation__.Notify(242124)
		return o.IsSystemTarget() == true
	}() == true {
		__antithesis_instrumentation__.Notify(242125)
		return t.GetSystemTarget().less(o.GetSystemTarget())
	} else {
		__antithesis_instrumentation__.Notify(242126)
	}
	__antithesis_instrumentation__.Notify(242121)

	if t.IsSystemTarget() {
		__antithesis_instrumentation__.Notify(242127)
		return true
	} else {
		__antithesis_instrumentation__.Notify(242128)
		if o.IsSystemTarget() {
			__antithesis_instrumentation__.Notify(242129)
			return false
		} else {
			__antithesis_instrumentation__.Notify(242130)
		}
	}
	__antithesis_instrumentation__.Notify(242122)

	if !t.GetSpan().Key.Equal(o.GetSpan().Key) {
		__antithesis_instrumentation__.Notify(242131)
		return t.GetSpan().Key.Compare(o.GetSpan().Key) < 0
	} else {
		__antithesis_instrumentation__.Notify(242132)
	}
	__antithesis_instrumentation__.Notify(242123)

	return t.GetSpan().EndKey.Compare(o.GetSpan().EndKey) < 0
}

func (t Target) Equal(o Target) bool {
	__antithesis_instrumentation__.Notify(242133)
	if t.IsSpanTarget() && func() bool {
		__antithesis_instrumentation__.Notify(242136)
		return o.IsSpanTarget() == true
	}() == true {
		__antithesis_instrumentation__.Notify(242137)
		return t.GetSpan().Equal(o.GetSpan())
	} else {
		__antithesis_instrumentation__.Notify(242138)
	}
	__antithesis_instrumentation__.Notify(242134)

	if t.IsSystemTarget() && func() bool {
		__antithesis_instrumentation__.Notify(242139)
		return o.IsSystemTarget() == true
	}() == true {
		__antithesis_instrumentation__.Notify(242140)
		return t.GetSystemTarget().equal(o.GetSystemTarget())
	} else {
		__antithesis_instrumentation__.Notify(242141)
	}
	__antithesis_instrumentation__.Notify(242135)

	return false
}

func (t Target) String() string {
	__antithesis_instrumentation__.Notify(242142)
	if t.IsSpanTarget() {
		__antithesis_instrumentation__.Notify(242144)
		return t.GetSpan().String()
	} else {
		__antithesis_instrumentation__.Notify(242145)
	}
	__antithesis_instrumentation__.Notify(242143)
	return t.GetSystemTarget().String()
}

func (t Target) isEmpty() bool {
	__antithesis_instrumentation__.Notify(242146)
	return t.systemTarget.IsEmpty() && func() bool {
		__antithesis_instrumentation__.Notify(242147)
		return t.span.Equal(roachpb.Span{}) == true
	}() == true
}

func (t Target) ToProto() roachpb.SpanConfigTarget {
	__antithesis_instrumentation__.Notify(242148)
	switch {
	case t.IsSpanTarget():
		__antithesis_instrumentation__.Notify(242149)
		sp := t.GetSpan()
		return roachpb.SpanConfigTarget{
			Union: &roachpb.SpanConfigTarget_Span{
				Span: &sp,
			},
		}
	case t.IsSystemTarget():
		__antithesis_instrumentation__.Notify(242150)
		return roachpb.SpanConfigTarget{
			Union: &roachpb.SpanConfigTarget_SystemSpanConfigTarget{
				SystemSpanConfigTarget: t.GetSystemTarget().toProto(),
			},
		}
	default:
		__antithesis_instrumentation__.Notify(242151)
		panic("cannot handle any other type of target")
	}
}

func DecodeTarget(span roachpb.Span) Target {
	__antithesis_instrumentation__.Notify(242152)
	if spanStartKeyConformsToSystemTargetEncoding(span) {
		__antithesis_instrumentation__.Notify(242154)
		systemTarget, err := decodeSystemTarget(span)
		if err != nil {
			__antithesis_instrumentation__.Notify(242156)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(242157)
		}
		__antithesis_instrumentation__.Notify(242155)
		return Target{systemTarget: systemTarget}
	} else {
		__antithesis_instrumentation__.Notify(242158)
	}
	__antithesis_instrumentation__.Notify(242153)
	return Target{span: span}
}

type Targets []Target

func (t Targets) Len() int { __antithesis_instrumentation__.Notify(242159); return len(t) }

func (t Targets) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(242160)
	t[i], t[j] = t[j], t[i]
}

func (t Targets) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(242161)
	return t[i].Less(t[j])
}

func RecordsToEntries(records []Record) []roachpb.SpanConfigEntry {
	__antithesis_instrumentation__.Notify(242162)
	entries := make([]roachpb.SpanConfigEntry, 0, len(records))
	for _, rec := range records {
		__antithesis_instrumentation__.Notify(242164)
		entries = append(entries, roachpb.SpanConfigEntry{
			Target: rec.GetTarget().ToProto(),
			Config: rec.GetConfig(),
		})
	}
	__antithesis_instrumentation__.Notify(242163)
	return entries
}

func EntriesToRecords(entries []roachpb.SpanConfigEntry) ([]Record, error) {
	__antithesis_instrumentation__.Notify(242165)
	records := make([]Record, 0, len(entries))
	for _, entry := range entries {
		__antithesis_instrumentation__.Notify(242167)
		target, err := MakeTarget(entry.Target)
		if err != nil {
			__antithesis_instrumentation__.Notify(242170)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(242171)
		}
		__antithesis_instrumentation__.Notify(242168)
		record, err := MakeRecord(target, entry.Config)
		if err != nil {
			__antithesis_instrumentation__.Notify(242172)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(242173)
		}
		__antithesis_instrumentation__.Notify(242169)
		records = append(records, record)
	}
	__antithesis_instrumentation__.Notify(242166)
	return records, nil
}

func TargetsToProtos(targets []Target) []roachpb.SpanConfigTarget {
	__antithesis_instrumentation__.Notify(242174)
	targetProtos := make([]roachpb.SpanConfigTarget, 0, len(targets))
	for _, target := range targets {
		__antithesis_instrumentation__.Notify(242176)
		targetProtos = append(targetProtos, target.ToProto())
	}
	__antithesis_instrumentation__.Notify(242175)
	return targetProtos
}

func TargetsFromProtos(protoTargets []roachpb.SpanConfigTarget) ([]Target, error) {
	__antithesis_instrumentation__.Notify(242177)
	targets := make([]Target, 0, len(protoTargets))
	for _, t := range protoTargets {
		__antithesis_instrumentation__.Notify(242179)
		target, err := MakeTarget(t)
		if err != nil {
			__antithesis_instrumentation__.Notify(242181)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(242182)
		}
		__antithesis_instrumentation__.Notify(242180)
		targets = append(targets, target)
	}
	__antithesis_instrumentation__.Notify(242178)
	return targets, nil
}

func TargetsFromRecords(records []Record) []Target {
	__antithesis_instrumentation__.Notify(242183)
	targets := make([]Target, len(records))
	for i, rec := range records {
		__antithesis_instrumentation__.Notify(242185)
		targets[i] = rec.GetTarget()
	}
	__antithesis_instrumentation__.Notify(242184)
	return targets
}

func TestingEntireSpanConfigurationStateTargets() []Target {
	__antithesis_instrumentation__.Notify(242186)
	return Targets{
		Target{
			span: keys.EverythingSpan,
		},
	}
}

func TestingMakeTenantKeyspaceTargetOrFatal(
	t *testing.T, sourceID roachpb.TenantID, targetID roachpb.TenantID,
) SystemTarget {
	__antithesis_instrumentation__.Notify(242187)
	target, err := MakeTenantKeyspaceTarget(sourceID, targetID)
	require.NoError(t, err)
	return target
}
