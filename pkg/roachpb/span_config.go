package roachpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
)

func StoreMatchesConstraint(store StoreDescriptor, c Constraint) bool {
	__antithesis_instrumentation__.Notify(176944)
	if c.Key == "" {
		__antithesis_instrumentation__.Notify(176947)
		for _, attrs := range []Attributes{store.Attrs, store.Node.Attrs} {
			__antithesis_instrumentation__.Notify(176949)
			for _, attr := range attrs.Attrs {
				__antithesis_instrumentation__.Notify(176950)
				if attr == c.Value {
					__antithesis_instrumentation__.Notify(176951)
					return true
				} else {
					__antithesis_instrumentation__.Notify(176952)
				}
			}
		}
		__antithesis_instrumentation__.Notify(176948)
		return false
	} else {
		__antithesis_instrumentation__.Notify(176953)
	}
	__antithesis_instrumentation__.Notify(176945)
	for _, tier := range store.Node.Locality.Tiers {
		__antithesis_instrumentation__.Notify(176954)
		if c.Key == tier.Key && func() bool {
			__antithesis_instrumentation__.Notify(176955)
			return c.Value == tier.Value == true
		}() == true {
			__antithesis_instrumentation__.Notify(176956)
			return true
		} else {
			__antithesis_instrumentation__.Notify(176957)
		}
	}
	__antithesis_instrumentation__.Notify(176946)
	return false
}

var emptySpanConfig = &SpanConfig{}

func (s *SpanConfig) IsEmpty() bool {
	__antithesis_instrumentation__.Notify(176958)
	return s.Equal(emptySpanConfig)
}

func (s *SpanConfig) TTL() time.Duration {
	__antithesis_instrumentation__.Notify(176959)
	return time.Duration(s.GCPolicy.TTLSeconds) * time.Second
}

func (s *SpanConfig) ValidateSystemTargetSpanConfig() error {
	__antithesis_instrumentation__.Notify(176960)
	if s.RangeMinBytes != 0 {
		__antithesis_instrumentation__.Notify(176973)
		return errors.AssertionFailedf("RangeMinBytes set on system span config")
	} else {
		__antithesis_instrumentation__.Notify(176974)
	}
	__antithesis_instrumentation__.Notify(176961)
	if s.RangeMaxBytes != 0 {
		__antithesis_instrumentation__.Notify(176975)
		return errors.AssertionFailedf("RangeMaxBytes set on system span config")
	} else {
		__antithesis_instrumentation__.Notify(176976)
	}
	__antithesis_instrumentation__.Notify(176962)
	if s.GCPolicy.TTLSeconds != 0 {
		__antithesis_instrumentation__.Notify(176977)
		return errors.AssertionFailedf("TTLSeconds set on system span config")
	} else {
		__antithesis_instrumentation__.Notify(176978)
	}
	__antithesis_instrumentation__.Notify(176963)
	if s.GCPolicy.IgnoreStrictEnforcement {
		__antithesis_instrumentation__.Notify(176979)
		return errors.AssertionFailedf("IgnoreStrictEnforcement set on system span config")
	} else {
		__antithesis_instrumentation__.Notify(176980)
	}
	__antithesis_instrumentation__.Notify(176964)
	if s.GlobalReads {
		__antithesis_instrumentation__.Notify(176981)
		return errors.AssertionFailedf("GlobalReads set on system span config")
	} else {
		__antithesis_instrumentation__.Notify(176982)
	}
	__antithesis_instrumentation__.Notify(176965)
	if s.NumReplicas != 0 {
		__antithesis_instrumentation__.Notify(176983)
		return errors.AssertionFailedf("NumReplicas set on system span config")
	} else {
		__antithesis_instrumentation__.Notify(176984)
	}
	__antithesis_instrumentation__.Notify(176966)
	if s.NumVoters != 0 {
		__antithesis_instrumentation__.Notify(176985)
		return errors.AssertionFailedf("NumVoters set on system span config")
	} else {
		__antithesis_instrumentation__.Notify(176986)
	}
	__antithesis_instrumentation__.Notify(176967)
	if len(s.Constraints) != 0 {
		__antithesis_instrumentation__.Notify(176987)
		return errors.AssertionFailedf("Constraints set on system span config")
	} else {
		__antithesis_instrumentation__.Notify(176988)
	}
	__antithesis_instrumentation__.Notify(176968)
	if len(s.VoterConstraints) != 0 {
		__antithesis_instrumentation__.Notify(176989)
		return errors.AssertionFailedf("VoterConstraints set on system span config")
	} else {
		__antithesis_instrumentation__.Notify(176990)
	}
	__antithesis_instrumentation__.Notify(176969)
	if len(s.LeasePreferences) != 0 {
		__antithesis_instrumentation__.Notify(176991)
		return errors.AssertionFailedf("LeasePreferences set on system span config")
	} else {
		__antithesis_instrumentation__.Notify(176992)
	}
	__antithesis_instrumentation__.Notify(176970)
	if s.RangefeedEnabled {
		__antithesis_instrumentation__.Notify(176993)
		return errors.AssertionFailedf("RangefeedEnabled set on system span config")
	} else {
		__antithesis_instrumentation__.Notify(176994)
	}
	__antithesis_instrumentation__.Notify(176971)
	if s.ExcludeDataFromBackup {
		__antithesis_instrumentation__.Notify(176995)
		return errors.AssertionFailedf("ExcludeDataFromBackup set on system span config")
	} else {
		__antithesis_instrumentation__.Notify(176996)
	}
	__antithesis_instrumentation__.Notify(176972)
	return nil
}

func (s *SpanConfig) GetNumVoters() int32 {
	__antithesis_instrumentation__.Notify(176997)
	if s.NumVoters != 0 {
		__antithesis_instrumentation__.Notify(176999)
		return s.NumVoters
	} else {
		__antithesis_instrumentation__.Notify(177000)
	}
	__antithesis_instrumentation__.Notify(176998)
	return s.NumReplicas
}

func (s *SpanConfig) GetNumNonVoters() int32 {
	__antithesis_instrumentation__.Notify(177001)
	return s.NumReplicas - s.GetNumVoters()
}

func (c Constraint) String() string {
	__antithesis_instrumentation__.Notify(177002)
	var str string
	switch c.Type {
	case Constraint_REQUIRED:
		__antithesis_instrumentation__.Notify(177005)
		str += "+"
	case Constraint_PROHIBITED:
		__antithesis_instrumentation__.Notify(177006)
		str += "-"
	default:
		__antithesis_instrumentation__.Notify(177007)
	}
	__antithesis_instrumentation__.Notify(177003)
	if len(c.Key) > 0 {
		__antithesis_instrumentation__.Notify(177008)
		str += c.Key + "="
	} else {
		__antithesis_instrumentation__.Notify(177009)
	}
	__antithesis_instrumentation__.Notify(177004)
	str += c.Value
	return str
}

func (c ConstraintsConjunction) String() string {
	__antithesis_instrumentation__.Notify(177010)
	var sb strings.Builder
	for i, cons := range c.Constraints {
		__antithesis_instrumentation__.Notify(177013)
		if i > 0 {
			__antithesis_instrumentation__.Notify(177015)
			sb.WriteRune(',')
		} else {
			__antithesis_instrumentation__.Notify(177016)
		}
		__antithesis_instrumentation__.Notify(177014)
		sb.WriteString(cons.String())
	}
	__antithesis_instrumentation__.Notify(177011)
	if c.NumReplicas != 0 {
		__antithesis_instrumentation__.Notify(177017)
		fmt.Fprintf(&sb, ":%d", c.NumReplicas)
	} else {
		__antithesis_instrumentation__.Notify(177018)
	}
	__antithesis_instrumentation__.Notify(177012)
	return sb.String()
}

func (p ProtectionPolicy) String() string {
	__antithesis_instrumentation__.Notify(177019)
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("{ts: %d", int(p.ProtectedTimestamp.WallTime)))
	if p.IgnoreIfExcludedFromBackup {
		__antithesis_instrumentation__.Notify(177021)
		sb.WriteString(fmt.Sprintf(",ignore_if_excluded_from_backup: %t",
			p.IgnoreIfExcludedFromBackup))
	} else {
		__antithesis_instrumentation__.Notify(177022)
	}
	__antithesis_instrumentation__.Notify(177020)
	sb.WriteString("}")
	return sb.String()
}

func TestingDefaultSpanConfig() SpanConfig {
	__antithesis_instrumentation__.Notify(177023)
	return SpanConfig{
		RangeMinBytes: 128 << 20,
		RangeMaxBytes: 512 << 20,
		GCPolicy: GCPolicy{
			TTLSeconds: 25 * 60 * 60,
		},
		NumReplicas: 3,
	}
}

func TestingDefaultSystemSpanConfiguration() SpanConfig {
	__antithesis_instrumentation__.Notify(177024)
	return SpanConfig{}
}

func TestingSystemSpanConfig() SpanConfig {
	__antithesis_instrumentation__.Notify(177025)
	config := TestingDefaultSpanConfig()
	config.NumReplicas = 5
	return config
}

func TestingDatabaseSystemSpanConfig(host bool) SpanConfig {
	__antithesis_instrumentation__.Notify(177026)
	config := TestingSystemSpanConfig()
	if !host {
		__antithesis_instrumentation__.Notify(177028)
		config = TestingDefaultSpanConfig()
	} else {
		__antithesis_instrumentation__.Notify(177029)
	}
	__antithesis_instrumentation__.Notify(177027)
	config.RangefeedEnabled = true
	config.GCPolicy.IgnoreStrictEnforcement = true
	return config
}

func (st SystemSpanConfigTarget) IsEntireKeyspaceTarget() bool {
	__antithesis_instrumentation__.Notify(177030)
	return st.Type.GetEntireKeyspace() != nil
}

func (st SystemSpanConfigTarget) IsSpecificTenantKeyspaceTarget() bool {
	__antithesis_instrumentation__.Notify(177031)
	return st.Type.GetSpecificTenantKeyspace() != nil
}

func (st SystemSpanConfigTarget) IsAllTenantKeyspaceTargetsSetTarget() bool {
	__antithesis_instrumentation__.Notify(177032)
	return st.Type.GetAllTenantKeyspaceTargetsSet() != nil
}

func NewEntireKeyspaceTargetType() *SystemSpanConfigTarget_Type {
	__antithesis_instrumentation__.Notify(177033)
	return &SystemSpanConfigTarget_Type{
		Type: &SystemSpanConfigTarget_Type_EntireKeyspace{
			EntireKeyspace: &SystemSpanConfigTarget_EntireKeyspace{},
		},
	}
}

func NewSpecificTenantKeyspaceTargetType(tenantID TenantID) *SystemSpanConfigTarget_Type {
	__antithesis_instrumentation__.Notify(177034)
	return &SystemSpanConfigTarget_Type{
		Type: &SystemSpanConfigTarget_Type_SpecificTenantKeyspace{
			SpecificTenantKeyspace: &SystemSpanConfigTarget_TenantKeyspace{
				TenantID: tenantID,
			},
		},
	}
}

func NewAllTenantKeyspaceTargetsSetTargetType() *SystemSpanConfigTarget_Type {
	__antithesis_instrumentation__.Notify(177035)
	return &SystemSpanConfigTarget_Type{
		Type: &SystemSpanConfigTarget_Type_AllTenantKeyspaceTargetsSet{
			AllTenantKeyspaceTargetsSet: &SystemSpanConfigTarget_AllTenantKeyspaceTargetsSet{},
		},
	}
}
