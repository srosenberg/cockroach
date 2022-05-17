package zonepb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/gogo/protobuf/proto"
	"gopkg.in/yaml.v2"
)

var _ yaml.Marshaler = LeasePreference{}
var _ yaml.Unmarshaler = &LeasePreference{}

func (l LeasePreference) MarshalYAML() (interface{}, error) {
	__antithesis_instrumentation__.Notify(58350)
	short := make([]string, len(l.Constraints))
	for i, c := range l.Constraints {
		__antithesis_instrumentation__.Notify(58352)
		short[i] = c.String()
	}
	__antithesis_instrumentation__.Notify(58351)
	return short, nil
}

func (l *LeasePreference) UnmarshalYAML(unmarshal func(interface{}) error) error {
	__antithesis_instrumentation__.Notify(58353)
	var shortConstraints []string
	if err := unmarshal(&shortConstraints); err != nil {
		__antithesis_instrumentation__.Notify(58356)
		return err
	} else {
		__antithesis_instrumentation__.Notify(58357)
	}
	__antithesis_instrumentation__.Notify(58354)
	constraints := make([]Constraint, len(shortConstraints))
	for i, short := range shortConstraints {
		__antithesis_instrumentation__.Notify(58358)
		if err := constraints[i].FromString(short); err != nil {
			__antithesis_instrumentation__.Notify(58359)
			return err
		} else {
			__antithesis_instrumentation__.Notify(58360)
		}
	}
	__antithesis_instrumentation__.Notify(58355)
	l.Constraints = constraints
	return nil
}

var _ yaml.Marshaler = ConstraintsConjunction{}
var _ yaml.Unmarshaler = &ConstraintsConjunction{}

func (c ConstraintsConjunction) MarshalYAML() (interface{}, error) {
	__antithesis_instrumentation__.Notify(58361)
	return nil, fmt.Errorf(
		"MarshalYAML should never be called directly on Constraints (%v): %v", c, debug.Stack())
}

func (c *ConstraintsConjunction) UnmarshalYAML(unmarshal func(interface{}) error) error {
	__antithesis_instrumentation__.Notify(58362)
	return fmt.Errorf(
		"UnmarshalYAML should never be called directly on Constraints: %v", debug.Stack())
}

type ConstraintsList struct {
	Constraints []ConstraintsConjunction
	Inherited   bool
}

var _ yaml.Marshaler = ConstraintsList{}
var _ yaml.Unmarshaler = &ConstraintsList{}

func (c ConstraintsList) MarshalYAML() (interface{}, error) {
	__antithesis_instrumentation__.Notify(58363)

	if c.Inherited || func() bool {
		__antithesis_instrumentation__.Notify(58367)
		return len(c.Constraints) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(58368)
		return []string{}, nil
	} else {
		__antithesis_instrumentation__.Notify(58369)
	}
	__antithesis_instrumentation__.Notify(58364)
	if len(c.Constraints) == 1 && func() bool {
		__antithesis_instrumentation__.Notify(58370)
		return c.Constraints[0].NumReplicas == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(58371)
		short := make([]string, len(c.Constraints[0].Constraints))
		for i, constraint := range c.Constraints[0].Constraints {
			__antithesis_instrumentation__.Notify(58373)
			short[i] = constraint.String()
		}
		__antithesis_instrumentation__.Notify(58372)
		return short, nil
	} else {
		__antithesis_instrumentation__.Notify(58374)
	}
	__antithesis_instrumentation__.Notify(58365)

	constraintsMap := make(map[string]int32)
	for _, constraints := range c.Constraints {
		__antithesis_instrumentation__.Notify(58375)
		short := make([]string, len(constraints.Constraints))
		for i, constraint := range constraints.Constraints {
			__antithesis_instrumentation__.Notify(58377)
			short[i] = constraint.String()
		}
		__antithesis_instrumentation__.Notify(58376)
		constraintsMap[strings.Join(short, ",")] = constraints.NumReplicas
	}
	__antithesis_instrumentation__.Notify(58366)
	return constraintsMap, nil
}

func (c *ConstraintsList) UnmarshalYAML(unmarshal func(interface{}) error) error {
	__antithesis_instrumentation__.Notify(58378)

	var strs []string
	c.Inherited = true
	if err := unmarshal(&strs); err == nil {
		__antithesis_instrumentation__.Notify(58383)
		constraints := make([]Constraint, len(strs))
		for i, short := range strs {
			__antithesis_instrumentation__.Notify(58386)
			if err := constraints[i].FromString(short); err != nil {
				__antithesis_instrumentation__.Notify(58387)
				return err
			} else {
				__antithesis_instrumentation__.Notify(58388)
			}
		}
		__antithesis_instrumentation__.Notify(58384)
		if len(constraints) == 0 {
			__antithesis_instrumentation__.Notify(58389)
			c.Constraints = []ConstraintsConjunction{}
			c.Inherited = false
		} else {
			__antithesis_instrumentation__.Notify(58390)
			c.Constraints = []ConstraintsConjunction{
				{
					Constraints: constraints,
					NumReplicas: 0,
				},
			}
			c.Inherited = false
		}
		__antithesis_instrumentation__.Notify(58385)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(58391)
	}
	__antithesis_instrumentation__.Notify(58379)

	constraintsMap := make(map[string]int32)
	if err := unmarshal(&constraintsMap); err != nil {
		__antithesis_instrumentation__.Notify(58392)
		return errors.New(
			"invalid constraints format. expected an array of strings or a map of strings to ints")
	} else {
		__antithesis_instrumentation__.Notify(58393)
	}
	__antithesis_instrumentation__.Notify(58380)

	constraintsList := make([]ConstraintsConjunction, 0, len(constraintsMap))
	for constraintsStr, numReplicas := range constraintsMap {
		__antithesis_instrumentation__.Notify(58394)
		shortConstraints := strings.Split(constraintsStr, ",")
		constraints := make([]Constraint, len(shortConstraints))
		for i, short := range shortConstraints {
			__antithesis_instrumentation__.Notify(58396)
			if err := constraints[i].FromString(short); err != nil {
				__antithesis_instrumentation__.Notify(58397)
				return err
			} else {
				__antithesis_instrumentation__.Notify(58398)
			}
		}
		__antithesis_instrumentation__.Notify(58395)
		constraintsList = append(constraintsList, ConstraintsConjunction{
			Constraints: constraints,
			NumReplicas: numReplicas,
		})
	}
	__antithesis_instrumentation__.Notify(58381)

	sort.Slice(constraintsList, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(58399)

		for k := range constraintsList[i].Constraints {
			__antithesis_instrumentation__.Notify(58402)
			if k >= len(constraintsList[j].Constraints) {
				__antithesis_instrumentation__.Notify(58405)
				return false
			} else {
				__antithesis_instrumentation__.Notify(58406)
			}
			__antithesis_instrumentation__.Notify(58403)
			lStr := constraintsList[i].Constraints[k].String()
			rStr := constraintsList[j].Constraints[k].String()
			if lStr < rStr {
				__antithesis_instrumentation__.Notify(58407)
				return true
			} else {
				__antithesis_instrumentation__.Notify(58408)
			}
			__antithesis_instrumentation__.Notify(58404)
			if lStr > rStr {
				__antithesis_instrumentation__.Notify(58409)
				return false
			} else {
				__antithesis_instrumentation__.Notify(58410)
			}
		}
		__antithesis_instrumentation__.Notify(58400)
		if len(constraintsList[i].Constraints) < len(constraintsList[j].Constraints) {
			__antithesis_instrumentation__.Notify(58411)
			return true
		} else {
			__antithesis_instrumentation__.Notify(58412)
		}
		__antithesis_instrumentation__.Notify(58401)

		return constraintsList[i].NumReplicas < constraintsList[j].NumReplicas
	})
	__antithesis_instrumentation__.Notify(58382)

	c.Constraints = constraintsList
	c.Inherited = false
	return nil
}

type marshalableZoneConfig struct {
	RangeMinBytes                *int64            `json:"range_min_bytes" yaml:"range_min_bytes"`
	RangeMaxBytes                *int64            `json:"range_max_bytes" yaml:"range_max_bytes"`
	GC                           *GCPolicy         `json:"gc"`
	GlobalReads                  *bool             `json:"global_reads" yaml:"global_reads"`
	NumReplicas                  *int32            `json:"num_replicas" yaml:"num_replicas"`
	NumVoters                    *int32            `json:"num_voters" yaml:"num_voters"`
	Constraints                  ConstraintsList   `json:"constraints" yaml:"constraints,flow"`
	VoterConstraints             ConstraintsList   `json:"voter_constraints" yaml:"voter_constraints,flow"`
	LeasePreferences             []LeasePreference `json:"lease_preferences" yaml:"lease_preferences,flow"`
	ExperimentalLeasePreferences []LeasePreference `json:"experimental_lease_preferences" yaml:"experimental_lease_preferences,flow,omitempty"`
	Subzones                     []Subzone         `json:"subzones" yaml:"-"`
	SubzoneSpans                 []SubzoneSpan     `json:"subzone_spans" yaml:"-"`
}

func zoneConfigToMarshalable(c ZoneConfig) marshalableZoneConfig {
	__antithesis_instrumentation__.Notify(58413)
	var m marshalableZoneConfig
	if c.RangeMinBytes != nil {
		__antithesis_instrumentation__.Notify(58421)
		m.RangeMinBytes = proto.Int64(*c.RangeMinBytes)
	} else {
		__antithesis_instrumentation__.Notify(58422)
	}
	__antithesis_instrumentation__.Notify(58414)
	if c.RangeMaxBytes != nil {
		__antithesis_instrumentation__.Notify(58423)
		m.RangeMaxBytes = proto.Int64(*c.RangeMaxBytes)
	} else {
		__antithesis_instrumentation__.Notify(58424)
	}
	__antithesis_instrumentation__.Notify(58415)
	if c.GC != nil {
		__antithesis_instrumentation__.Notify(58425)
		tempGC := *c.GC
		m.GC = &tempGC
	} else {
		__antithesis_instrumentation__.Notify(58426)
	}
	__antithesis_instrumentation__.Notify(58416)
	if c.GlobalReads != nil {
		__antithesis_instrumentation__.Notify(58427)
		m.GlobalReads = proto.Bool(*c.GlobalReads)
	} else {
		__antithesis_instrumentation__.Notify(58428)
	}
	__antithesis_instrumentation__.Notify(58417)
	if c.NumReplicas != nil && func() bool {
		__antithesis_instrumentation__.Notify(58429)
		return *c.NumReplicas != 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(58430)
		m.NumReplicas = proto.Int32(*c.NumReplicas)
	} else {
		__antithesis_instrumentation__.Notify(58431)
	}
	__antithesis_instrumentation__.Notify(58418)
	m.Constraints = ConstraintsList{c.Constraints, c.InheritedConstraints}
	if c.NumVoters != nil && func() bool {
		__antithesis_instrumentation__.Notify(58432)
		return *c.NumVoters != 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(58433)
		m.NumVoters = proto.Int32(*c.NumVoters)
	} else {
		__antithesis_instrumentation__.Notify(58434)
	}
	__antithesis_instrumentation__.Notify(58419)

	m.VoterConstraints = ConstraintsList{c.VoterConstraints, !c.NullVoterConstraintsIsEmpty}
	if !c.InheritedLeasePreferences {
		__antithesis_instrumentation__.Notify(58435)
		m.LeasePreferences = c.LeasePreferences
	} else {
		__antithesis_instrumentation__.Notify(58436)
	}
	__antithesis_instrumentation__.Notify(58420)

	m.Subzones = c.Subzones
	m.SubzoneSpans = c.SubzoneSpans
	return m
}

func zoneConfigFromMarshalable(m marshalableZoneConfig, c ZoneConfig) ZoneConfig {
	__antithesis_instrumentation__.Notify(58437)
	if m.RangeMinBytes != nil {
		__antithesis_instrumentation__.Notify(58447)
		c.RangeMinBytes = proto.Int64(*m.RangeMinBytes)
	} else {
		__antithesis_instrumentation__.Notify(58448)
	}
	__antithesis_instrumentation__.Notify(58438)
	if m.RangeMaxBytes != nil {
		__antithesis_instrumentation__.Notify(58449)
		c.RangeMaxBytes = proto.Int64(*m.RangeMaxBytes)
	} else {
		__antithesis_instrumentation__.Notify(58450)
	}
	__antithesis_instrumentation__.Notify(58439)
	if m.GC != nil {
		__antithesis_instrumentation__.Notify(58451)
		tempGC := *m.GC
		c.GC = &tempGC
	} else {
		__antithesis_instrumentation__.Notify(58452)
	}
	__antithesis_instrumentation__.Notify(58440)
	if m.GlobalReads != nil {
		__antithesis_instrumentation__.Notify(58453)
		c.GlobalReads = proto.Bool(*m.GlobalReads)
	} else {
		__antithesis_instrumentation__.Notify(58454)
	}
	__antithesis_instrumentation__.Notify(58441)
	if m.NumReplicas != nil {
		__antithesis_instrumentation__.Notify(58455)
		c.NumReplicas = proto.Int32(*m.NumReplicas)
	} else {
		__antithesis_instrumentation__.Notify(58456)
	}
	__antithesis_instrumentation__.Notify(58442)
	c.Constraints = m.Constraints.Constraints
	c.InheritedConstraints = m.Constraints.Inherited
	if m.NumVoters != nil {
		__antithesis_instrumentation__.Notify(58457)
		c.NumVoters = proto.Int32(*m.NumVoters)
	} else {
		__antithesis_instrumentation__.Notify(58458)
	}
	__antithesis_instrumentation__.Notify(58443)
	c.VoterConstraints = m.VoterConstraints.Constraints
	c.NullVoterConstraintsIsEmpty = !m.VoterConstraints.Inherited
	if m.LeasePreferences != nil {
		__antithesis_instrumentation__.Notify(58459)
		c.LeasePreferences = m.LeasePreferences
	} else {
		__antithesis_instrumentation__.Notify(58460)
	}
	__antithesis_instrumentation__.Notify(58444)

	if m.ExperimentalLeasePreferences != nil {
		__antithesis_instrumentation__.Notify(58461)
		c.LeasePreferences = m.ExperimentalLeasePreferences
	} else {
		__antithesis_instrumentation__.Notify(58462)
	}
	__antithesis_instrumentation__.Notify(58445)

	if m.LeasePreferences != nil || func() bool {
		__antithesis_instrumentation__.Notify(58463)
		return m.ExperimentalLeasePreferences != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(58464)
		c.InheritedLeasePreferences = false
	} else {
		__antithesis_instrumentation__.Notify(58465)
	}
	__antithesis_instrumentation__.Notify(58446)
	c.Subzones = m.Subzones
	c.SubzoneSpans = m.SubzoneSpans
	return c
}

var _ yaml.Marshaler = ZoneConfig{}
var _ yaml.Unmarshaler = &ZoneConfig{}

func (c ZoneConfig) MarshalYAML() (interface{}, error) {
	__antithesis_instrumentation__.Notify(58466)
	return zoneConfigToMarshalable(c), nil
}

func (c *ZoneConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	__antithesis_instrumentation__.Notify(58467)

	aux := zoneConfigToMarshalable(*c)
	if err := unmarshal(&aux); err != nil {
		__antithesis_instrumentation__.Notify(58469)
		return err
	} else {
		__antithesis_instrumentation__.Notify(58470)
	}
	__antithesis_instrumentation__.Notify(58468)
	*c = zoneConfigFromMarshalable(aux, *c)
	return nil
}
