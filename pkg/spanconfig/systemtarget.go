package spanconfig

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/errors"
)

type SystemTarget struct {
	sourceTenantID roachpb.TenantID

	targetTenantID roachpb.TenantID

	systemTargetType systemTargetType
}

type systemTargetType int

const (
	_ systemTargetType = iota

	SystemTargetTypeSpecificTenantKeyspace

	SystemTargetTypeEntireKeyspace

	SystemTargetTypeAllTenantKeyspaceTargetsSet
)

func MakeTenantKeyspaceTarget(
	sourceTenantID roachpb.TenantID, targetTenantID roachpb.TenantID,
) (SystemTarget, error) {
	__antithesis_instrumentation__.Notify(241983)
	t := SystemTarget{
		sourceTenantID:   sourceTenantID,
		targetTenantID:   targetTenantID,
		systemTargetType: SystemTargetTypeSpecificTenantKeyspace,
	}
	return t, t.validate()
}

func makeSystemTargetFromProto(proto *roachpb.SystemSpanConfigTarget) (SystemTarget, error) {
	__antithesis_instrumentation__.Notify(241984)
	var t SystemTarget
	switch {
	case proto.IsSpecificTenantKeyspaceTarget():
		__antithesis_instrumentation__.Notify(241986)
		t = SystemTarget{
			sourceTenantID:   proto.SourceTenantID,
			targetTenantID:   proto.Type.GetSpecificTenantKeyspace().TenantID,
			systemTargetType: SystemTargetTypeSpecificTenantKeyspace,
		}
	case proto.IsEntireKeyspaceTarget():
		__antithesis_instrumentation__.Notify(241987)
		t = SystemTarget{
			sourceTenantID:   proto.SourceTenantID,
			targetTenantID:   roachpb.TenantID{},
			systemTargetType: SystemTargetTypeEntireKeyspace,
		}
	case proto.IsAllTenantKeyspaceTargetsSetTarget():
		__antithesis_instrumentation__.Notify(241988)
		t = SystemTarget{
			sourceTenantID:   proto.SourceTenantID,
			targetTenantID:   roachpb.TenantID{},
			systemTargetType: SystemTargetTypeAllTenantKeyspaceTargetsSet,
		}
	default:
		__antithesis_instrumentation__.Notify(241989)
		return SystemTarget{}, errors.AssertionFailedf("unknown system target type")
	}
	__antithesis_instrumentation__.Notify(241985)
	return t, t.validate()
}

func (st SystemTarget) toProto() *roachpb.SystemSpanConfigTarget {
	__antithesis_instrumentation__.Notify(241990)
	var systemTargetType *roachpb.SystemSpanConfigTarget_Type
	switch st.systemTargetType {
	case SystemTargetTypeEntireKeyspace:
		__antithesis_instrumentation__.Notify(241992)
		systemTargetType = roachpb.NewEntireKeyspaceTargetType()
	case SystemTargetTypeAllTenantKeyspaceTargetsSet:
		__antithesis_instrumentation__.Notify(241993)
		systemTargetType = roachpb.NewAllTenantKeyspaceTargetsSetTargetType()
	case SystemTargetTypeSpecificTenantKeyspace:
		__antithesis_instrumentation__.Notify(241994)
		systemTargetType = roachpb.NewSpecificTenantKeyspaceTargetType(st.targetTenantID)
	default:
		__antithesis_instrumentation__.Notify(241995)
		panic("unknown system target type")
	}
	__antithesis_instrumentation__.Notify(241991)
	return &roachpb.SystemSpanConfigTarget{
		SourceTenantID: st.sourceTenantID,
		Type:           systemTargetType,
	}
}

func MakeEntireKeyspaceTarget() SystemTarget {
	__antithesis_instrumentation__.Notify(241996)
	return SystemTarget{
		sourceTenantID:   roachpb.SystemTenantID,
		systemTargetType: SystemTargetTypeEntireKeyspace,
	}
}

func MakeAllTenantKeyspaceTargetsSet(sourceID roachpb.TenantID) SystemTarget {
	__antithesis_instrumentation__.Notify(241997)
	return SystemTarget{
		sourceTenantID:   sourceID,
		systemTargetType: SystemTargetTypeAllTenantKeyspaceTargetsSet,
	}
}

func (st SystemTarget) targetsEntireKeyspace() bool {
	__antithesis_instrumentation__.Notify(241998)
	return st.systemTargetType == SystemTargetTypeEntireKeyspace
}

func (st SystemTarget) keyspaceTargeted() roachpb.Span {
	__antithesis_instrumentation__.Notify(241999)
	switch st.systemTargetType {
	case SystemTargetTypeEntireKeyspace:
		__antithesis_instrumentation__.Notify(242000)
		return keys.EverythingSpan
	case SystemTargetTypeSpecificTenantKeyspace:
		__antithesis_instrumentation__.Notify(242001)

		if st.targetTenantID == roachpb.SystemTenantID {
			__antithesis_instrumentation__.Notify(242005)
			return roachpb.Span{
				Key:    keys.MinKey,
				EndKey: keys.TenantTableDataMin,
			}
		} else {
			__antithesis_instrumentation__.Notify(242006)
		}
		__antithesis_instrumentation__.Notify(242002)
		k := keys.MakeTenantPrefix(st.targetTenantID)
		return roachpb.Span{
			Key:    k,
			EndKey: k.PrefixEnd(),
		}
	case SystemTargetTypeAllTenantKeyspaceTargetsSet:
		__antithesis_instrumentation__.Notify(242003)

		panic("not applicable")
	default:
		__antithesis_instrumentation__.Notify(242004)
		panic("unknown target type")
	}
}

func (st SystemTarget) IsReadOnly() bool {
	__antithesis_instrumentation__.Notify(242007)
	return st.systemTargetType == SystemTargetTypeAllTenantKeyspaceTargetsSet
}

func (st SystemTarget) encode() roachpb.Span {
	__antithesis_instrumentation__.Notify(242008)
	var k roachpb.Key

	switch st.systemTargetType {
	case SystemTargetTypeEntireKeyspace:
		__antithesis_instrumentation__.Notify(242010)
		k = keys.SystemSpanConfigEntireKeyspace
	case SystemTargetTypeSpecificTenantKeyspace:
		__antithesis_instrumentation__.Notify(242011)
		if st.sourceTenantID == roachpb.SystemTenantID {
			__antithesis_instrumentation__.Notify(242014)
			k = encoding.EncodeUvarintAscending(
				keys.SystemSpanConfigHostOnTenantKeyspace, st.targetTenantID.ToUint64(),
			)
		} else {
			__antithesis_instrumentation__.Notify(242015)
			k = encoding.EncodeUvarintAscending(
				keys.SystemSpanConfigSecondaryTenantOnEntireKeyspace, st.sourceTenantID.ToUint64(),
			)
		}
	case SystemTargetTypeAllTenantKeyspaceTargetsSet:
		__antithesis_instrumentation__.Notify(242012)
		if st.sourceTenantID == roachpb.SystemTenantID {
			__antithesis_instrumentation__.Notify(242016)
			k = keys.SystemSpanConfigHostOnTenantKeyspace
		} else {
			__antithesis_instrumentation__.Notify(242017)
			k = encoding.EncodeUvarintAscending(
				keys.SystemSpanConfigSecondaryTenantOnEntireKeyspace, st.sourceTenantID.ToUint64(),
			)
		}
	default:
		__antithesis_instrumentation__.Notify(242013)
	}
	__antithesis_instrumentation__.Notify(242009)
	return roachpb.Span{Key: k, EndKey: k.PrefixEnd()}
}

func (st SystemTarget) validate() error {
	__antithesis_instrumentation__.Notify(242018)
	switch st.systemTargetType {
	case SystemTargetTypeAllTenantKeyspaceTargetsSet:
		__antithesis_instrumentation__.Notify(242020)
		if st.targetTenantID.IsSet() {
			__antithesis_instrumentation__.Notify(242026)
			return errors.AssertionFailedf(
				"targetTenantID must be unset when targeting everything installed on tenants",
			)
		} else {
			__antithesis_instrumentation__.Notify(242027)
		}
	case SystemTargetTypeEntireKeyspace:
		__antithesis_instrumentation__.Notify(242021)
		if st.sourceTenantID != roachpb.SystemTenantID {
			__antithesis_instrumentation__.Notify(242028)
			return errors.AssertionFailedf("only the host tenant is allowed to target the entire keyspace")
		} else {
			__antithesis_instrumentation__.Notify(242029)
		}
		__antithesis_instrumentation__.Notify(242022)
		if st.targetTenantID.IsSet() {
			__antithesis_instrumentation__.Notify(242030)
			return errors.AssertionFailedf("malformed system target for entire keyspace; targetTenantID set")
		} else {
			__antithesis_instrumentation__.Notify(242031)
		}
	case SystemTargetTypeSpecificTenantKeyspace:
		__antithesis_instrumentation__.Notify(242023)
		if !st.targetTenantID.IsSet() {
			__antithesis_instrumentation__.Notify(242032)
			return errors.AssertionFailedf(
				"malformed system target for specific tenant keyspace; targetTenantID unset",
			)
		} else {
			__antithesis_instrumentation__.Notify(242033)
		}
		__antithesis_instrumentation__.Notify(242024)
		if st.sourceTenantID != roachpb.SystemTenantID && func() bool {
			__antithesis_instrumentation__.Notify(242034)
			return st.sourceTenantID != st.targetTenantID == true
		}() == true {
			__antithesis_instrumentation__.Notify(242035)
			return errors.AssertionFailedf(
				"secondary tenant %s cannot target another tenant with ID %s",
				st.sourceTenantID,
				st.targetTenantID,
			)
		} else {
			__antithesis_instrumentation__.Notify(242036)
		}
	default:
		__antithesis_instrumentation__.Notify(242025)
		return errors.AssertionFailedf("invalid system target type")
	}
	__antithesis_instrumentation__.Notify(242019)
	return nil
}

func (st SystemTarget) IsEmpty() bool {
	__antithesis_instrumentation__.Notify(242037)
	return !st.sourceTenantID.IsSet() && func() bool {
		__antithesis_instrumentation__.Notify(242038)
		return !st.targetTenantID.IsSet() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(242039)
		return st.systemTargetType == 0 == true
	}() == true
}

func (st SystemTarget) less(ot SystemTarget) bool {
	__antithesis_instrumentation__.Notify(242040)
	if st.IsReadOnly() && func() bool {
		__antithesis_instrumentation__.Notify(242045)
		return ot.IsReadOnly() == true
	}() == true {
		__antithesis_instrumentation__.Notify(242046)
		return st.sourceTenantID.ToUint64() < ot.sourceTenantID.ToUint64()
	} else {
		__antithesis_instrumentation__.Notify(242047)
	}
	__antithesis_instrumentation__.Notify(242041)

	if st.IsReadOnly() {
		__antithesis_instrumentation__.Notify(242048)
		return true
	} else {
		__antithesis_instrumentation__.Notify(242049)
		if ot.IsReadOnly() {
			__antithesis_instrumentation__.Notify(242050)
			return false
		} else {
			__antithesis_instrumentation__.Notify(242051)
		}
	}
	__antithesis_instrumentation__.Notify(242042)

	if st.targetsEntireKeyspace() {
		__antithesis_instrumentation__.Notify(242052)
		return true
	} else {
		__antithesis_instrumentation__.Notify(242053)
		if ot.targetsEntireKeyspace() {
			__antithesis_instrumentation__.Notify(242054)
			return false
		} else {
			__antithesis_instrumentation__.Notify(242055)
		}
	}
	__antithesis_instrumentation__.Notify(242043)

	if st.sourceTenantID.ToUint64() == ot.sourceTenantID.ToUint64() {
		__antithesis_instrumentation__.Notify(242056)
		return st.targetTenantID.ToUint64() < ot.targetTenantID.ToUint64()
	} else {
		__antithesis_instrumentation__.Notify(242057)
	}
	__antithesis_instrumentation__.Notify(242044)

	return st.sourceTenantID.ToUint64() < ot.sourceTenantID.ToUint64()
}

func (st SystemTarget) equal(ot SystemTarget) bool {
	__antithesis_instrumentation__.Notify(242058)
	return st.sourceTenantID.Equal(ot.sourceTenantID) && func() bool {
		__antithesis_instrumentation__.Notify(242059)
		return st.targetTenantID.Equal(ot.targetTenantID) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(242060)
		return st.systemTargetType == ot.systemTargetType == true
	}() == true
}

func (st SystemTarget) String() string {
	__antithesis_instrumentation__.Notify(242061)
	switch st.systemTargetType {
	case SystemTargetTypeEntireKeyspace:
		__antithesis_instrumentation__.Notify(242062)
		return "{entire-keyspace}"
	case SystemTargetTypeAllTenantKeyspaceTargetsSet:
		__antithesis_instrumentation__.Notify(242063)
		return fmt.Sprintf("{source=%d, all-tenant-keyspace-targets-set}", st.sourceTenantID)
	case SystemTargetTypeSpecificTenantKeyspace:
		__antithesis_instrumentation__.Notify(242064)
		return fmt.Sprintf(
			"{source=%d,target=%d}",
			st.sourceTenantID.ToUint64(),
			st.targetTenantID.ToUint64(),
		)
	default:
		__antithesis_instrumentation__.Notify(242065)
		panic("unreachable")
	}
}

func decodeSystemTarget(span roachpb.Span) (SystemTarget, error) {
	__antithesis_instrumentation__.Notify(242066)

	if !span.EndKey.Equal(span.Key.PrefixEnd()) {
		__antithesis_instrumentation__.Notify(242068)
		return SystemTarget{}, errors.AssertionFailedf("invalid end key in span %s", span)
	} else {
		__antithesis_instrumentation__.Notify(242069)
	}
	__antithesis_instrumentation__.Notify(242067)
	switch {
	case bytes.Equal(span.Key, keys.SystemSpanConfigEntireKeyspace):
		__antithesis_instrumentation__.Notify(242070)
		return MakeEntireKeyspaceTarget(), nil
	case bytes.HasPrefix(span.Key, keys.SystemSpanConfigHostOnTenantKeyspace):
		__antithesis_instrumentation__.Notify(242071)

		tenIDBytes := span.Key[len(keys.SystemSpanConfigHostOnTenantKeyspace):]
		_, tenIDRaw, err := encoding.DecodeUvarintAscending(tenIDBytes)
		if err != nil {
			__antithesis_instrumentation__.Notify(242076)
			return SystemTarget{}, err
		} else {
			__antithesis_instrumentation__.Notify(242077)
		}
		__antithesis_instrumentation__.Notify(242072)
		tenID := roachpb.MakeTenantID(tenIDRaw)
		return MakeTenantKeyspaceTarget(roachpb.SystemTenantID, tenID)
	case bytes.HasPrefix(span.Key, keys.SystemSpanConfigSecondaryTenantOnEntireKeyspace):
		__antithesis_instrumentation__.Notify(242073)

		tenIDBytes := span.Key[len(keys.SystemSpanConfigSecondaryTenantOnEntireKeyspace):]
		_, tenIDRaw, err := encoding.DecodeUvarintAscending(tenIDBytes)
		if err != nil {
			__antithesis_instrumentation__.Notify(242078)
			return SystemTarget{}, err
		} else {
			__antithesis_instrumentation__.Notify(242079)
		}
		__antithesis_instrumentation__.Notify(242074)
		tenID := roachpb.MakeTenantID(tenIDRaw)
		return MakeTenantKeyspaceTarget(tenID, tenID)
	default:
		__antithesis_instrumentation__.Notify(242075)
		return SystemTarget{},
			errors.AssertionFailedf("span %s did not conform to SystemTarget encoding", span)
	}
}

func spanStartKeyConformsToSystemTargetEncoding(span roachpb.Span) bool {
	__antithesis_instrumentation__.Notify(242080)
	return bytes.Equal(span.Key, keys.SystemSpanConfigEntireKeyspace) || func() bool {
		__antithesis_instrumentation__.Notify(242081)
		return bytes.HasPrefix(span.Key, keys.SystemSpanConfigHostOnTenantKeyspace) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(242082)
		return bytes.HasPrefix(span.Key, keys.SystemSpanConfigSecondaryTenantOnEntireKeyspace) == true
	}() == true
}
