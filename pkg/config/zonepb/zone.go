package zonepb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
	"github.com/gogo/protobuf/proto"
)

type NamedZone string

const (
	DefaultZoneName    NamedZone = "default"
	LivenessZoneName   NamedZone = "liveness"
	MetaZoneName       NamedZone = "meta"
	SystemZoneName     NamedZone = "system"
	TimeseriesZoneName NamedZone = "timeseries"
	TenantsZoneName    NamedZone = "tenants"
)

var NamedZonesList = [...]NamedZone{
	DefaultZoneName,
	LivenessZoneName,
	MetaZoneName,
	SystemZoneName,
	TimeseriesZoneName,
	TenantsZoneName,
}

var NamedZones = map[NamedZone]uint32{
	DefaultZoneName:    keys.RootNamespaceID,
	LivenessZoneName:   keys.LivenessRangesID,
	MetaZoneName:       keys.MetaRangesID,
	SystemZoneName:     keys.SystemRangesID,
	TimeseriesZoneName: keys.TimeseriesRangesID,
	TenantsZoneName:    keys.TenantsRangesID,
}

var NamedZonesByID = func() map[uint32]NamedZone {
	__antithesis_instrumentation__.Notify(56244)
	out := map[uint32]NamedZone{}
	for name, id := range NamedZones {
		__antithesis_instrumentation__.Notify(56246)
		out[id] = name
	}
	__antithesis_instrumentation__.Notify(56245)
	return out
}()

func IsNamedZoneID(id uint32) bool {
	__antithesis_instrumentation__.Notify(56247)
	_, ok := NamedZonesByID[id]
	return ok
}

var MultiRegionZoneConfigFields = []tree.Name{
	"global_reads",
	"num_replicas",
	"num_voters",
	"constraints",
	"voter_constraints",
	"lease_preferences",
}

var MultiRegionZoneConfigFieldsSet = func() map[tree.Name]struct{} {
	__antithesis_instrumentation__.Notify(56248)
	ret := make(map[tree.Name]struct{}, len(MultiRegionZoneConfigFields))
	for _, f := range MultiRegionZoneConfigFields {
		__antithesis_instrumentation__.Notify(56250)
		ret[f] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(56249)
	return ret
}()

func ZoneSpecifierFromID(
	id uint32, resolveID func(id uint32) (parentID, parentSchemaID uint32, name string, err error),
) (tree.ZoneSpecifier, error) {
	__antithesis_instrumentation__.Notify(56251)
	if name, ok := NamedZonesByID[id]; ok {
		__antithesis_instrumentation__.Notify(56257)
		return tree.ZoneSpecifier{NamedZone: tree.UnrestrictedName(name)}, nil
	} else {
		__antithesis_instrumentation__.Notify(56258)
	}
	__antithesis_instrumentation__.Notify(56252)
	parentID, parentSchemaID, name, err := resolveID(id)
	if err != nil {
		__antithesis_instrumentation__.Notify(56259)
		return tree.ZoneSpecifier{}, err
	} else {
		__antithesis_instrumentation__.Notify(56260)
	}
	__antithesis_instrumentation__.Notify(56253)
	if parentID == keys.RootNamespaceID {
		__antithesis_instrumentation__.Notify(56261)
		return tree.ZoneSpecifier{Database: tree.Name(name)}, nil
	} else {
		__antithesis_instrumentation__.Notify(56262)
	}
	__antithesis_instrumentation__.Notify(56254)
	_, _, schemaName, err := resolveID(parentSchemaID)
	if err != nil {
		__antithesis_instrumentation__.Notify(56263)
		return tree.ZoneSpecifier{}, err
	} else {
		__antithesis_instrumentation__.Notify(56264)
	}
	__antithesis_instrumentation__.Notify(56255)
	_, _, databaseName, err := resolveID(parentID)
	if err != nil {
		__antithesis_instrumentation__.Notify(56265)
		return tree.ZoneSpecifier{}, err
	} else {
		__antithesis_instrumentation__.Notify(56266)
	}
	__antithesis_instrumentation__.Notify(56256)
	return tree.ZoneSpecifier{
		TableOrIndex: tree.TableIndexName{
			Table: tree.MakeTableNameWithSchema(tree.Name(databaseName), tree.Name(schemaName), tree.Name(name)),
		},
	}, nil
}

func ResolveZoneSpecifier(
	ctx context.Context,
	zs *tree.ZoneSpecifier,
	resolveName func(parentID uint32, schemaID uint32, name string) (id uint32, err error),
	version clusterversion.Handle,
) (uint32, error) {
	__antithesis_instrumentation__.Notify(56267)

	if zs.NamedZone != "" {
		__antithesis_instrumentation__.Notify(56273)
		if NamedZone(zs.NamedZone) == DefaultZoneName {
			__antithesis_instrumentation__.Notify(56276)
			return keys.RootNamespaceID, nil
		} else {
			__antithesis_instrumentation__.Notify(56277)
		}
		__antithesis_instrumentation__.Notify(56274)
		if id, ok := NamedZones[NamedZone(zs.NamedZone)]; ok {
			__antithesis_instrumentation__.Notify(56278)
			return id, nil
		} else {
			__antithesis_instrumentation__.Notify(56279)
		}
		__antithesis_instrumentation__.Notify(56275)
		return 0, fmt.Errorf("%q is not a built-in zone", string(zs.NamedZone))
	} else {
		__antithesis_instrumentation__.Notify(56280)
	}
	__antithesis_instrumentation__.Notify(56268)

	if zs.Database != "" {
		__antithesis_instrumentation__.Notify(56281)
		return resolveName(keys.RootNamespaceID, keys.RootNamespaceID, string(zs.Database))
	} else {
		__antithesis_instrumentation__.Notify(56282)
	}
	__antithesis_instrumentation__.Notify(56269)

	tn := &zs.TableOrIndex.Table
	databaseID, err := resolveName(keys.RootNamespaceID, keys.RootNamespaceID, tn.Catalog())
	if err != nil {
		__antithesis_instrumentation__.Notify(56283)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(56284)
	}
	__antithesis_instrumentation__.Notify(56270)

	var schemaID uint32
	if !version.IsActive(ctx, clusterversion.PublicSchemasWithDescriptors) && func() bool {
		__antithesis_instrumentation__.Notify(56285)
		return tn.SchemaName == tree.PublicSchemaName == true
	}() == true {
		__antithesis_instrumentation__.Notify(56286)

		schemaID = keys.PublicSchemaID
	} else {
		__antithesis_instrumentation__.Notify(56287)
		schemaID, err = resolveName(databaseID, keys.RootNamespaceID, tn.Schema())
		if err != nil {
			__antithesis_instrumentation__.Notify(56288)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(56289)
		}
	}
	__antithesis_instrumentation__.Notify(56271)
	tableID, err := resolveName(databaseID, schemaID, tn.Table())
	if err != nil {
		__antithesis_instrumentation__.Notify(56290)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(56291)
	}
	__antithesis_instrumentation__.Notify(56272)
	return tableID, err
}

func (c Constraint) String() string {
	__antithesis_instrumentation__.Notify(56292)
	var str string
	switch c.Type {
	case Constraint_REQUIRED:
		__antithesis_instrumentation__.Notify(56295)
		str += "+"
	case Constraint_PROHIBITED:
		__antithesis_instrumentation__.Notify(56296)
		str += "-"
	default:
		__antithesis_instrumentation__.Notify(56297)
	}
	__antithesis_instrumentation__.Notify(56293)
	if len(c.Key) > 0 {
		__antithesis_instrumentation__.Notify(56298)
		str += c.Key + "="
	} else {
		__antithesis_instrumentation__.Notify(56299)
	}
	__antithesis_instrumentation__.Notify(56294)
	str += c.Value
	return str
}

func (c *Constraint) FromString(short string) error {
	__antithesis_instrumentation__.Notify(56300)
	if len(short) == 0 {
		__antithesis_instrumentation__.Notify(56304)
		return fmt.Errorf("the empty string is not a valid constraint")
	} else {
		__antithesis_instrumentation__.Notify(56305)
	}
	__antithesis_instrumentation__.Notify(56301)
	switch short[0] {
	case '+':
		__antithesis_instrumentation__.Notify(56306)
		c.Type = Constraint_REQUIRED
		short = short[1:]
	case '-':
		__antithesis_instrumentation__.Notify(56307)
		c.Type = Constraint_PROHIBITED
		short = short[1:]
	default:
		__antithesis_instrumentation__.Notify(56308)
		c.Type = Constraint_DEPRECATED_POSITIVE
	}
	__antithesis_instrumentation__.Notify(56302)
	parts := strings.Split(short, "=")
	if len(parts) == 1 {
		__antithesis_instrumentation__.Notify(56309)
		c.Value = parts[0]
	} else {
		__antithesis_instrumentation__.Notify(56310)
		if len(parts) == 2 {
			__antithesis_instrumentation__.Notify(56311)
			c.Key = parts[0]
			c.Value = parts[1]
		} else {
			__antithesis_instrumentation__.Notify(56312)
			return errors.Errorf("constraint needs to be in the form \"(key=)value\", not %q", short)
		}
	}
	__antithesis_instrumentation__.Notify(56303)
	return nil
}

func NewZoneConfig() *ZoneConfig {
	__antithesis_instrumentation__.Notify(56313)
	return &ZoneConfig{
		InheritedConstraints:      true,
		InheritedLeasePreferences: true,
	}
}

func DefaultZoneConfig() ZoneConfig {
	__antithesis_instrumentation__.Notify(56314)
	return ZoneConfig{
		NumReplicas:   proto.Int32(3),
		RangeMinBytes: proto.Int64(128 << 20),
		RangeMaxBytes: proto.Int64(512 << 20),
		GC: &GCPolicy{

			TTLSeconds: 25 * 60 * 60,
		},

		NullVoterConstraintsIsEmpty: true,
	}
}

func DefaultZoneConfigRef() *ZoneConfig {
	__antithesis_instrumentation__.Notify(56315)
	zoneConfig := DefaultZoneConfig()
	return &zoneConfig
}

func DefaultSystemZoneConfig() ZoneConfig {
	__antithesis_instrumentation__.Notify(56316)
	defaultSystemZoneConfig := DefaultZoneConfig()
	defaultSystemZoneConfig.NumReplicas = proto.Int32(5)
	return defaultSystemZoneConfig
}

func DefaultSystemZoneConfigRef() *ZoneConfig {
	__antithesis_instrumentation__.Notify(56317)
	systemZoneConfig := DefaultSystemZoneConfig()
	return &systemZoneConfig
}

func (z *ZoneConfig) IsComplete() bool {
	__antithesis_instrumentation__.Notify(56318)
	return ((z.NumReplicas != nil) && func() bool {
		__antithesis_instrumentation__.Notify(56319)
		return (z.RangeMinBytes != nil) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(56320)
		return (z.RangeMaxBytes != nil) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(56321)
		return (z.GC != nil) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(56322)
		return (!z.InheritedVoterConstraints()) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(56323)
		return (!z.InheritedConstraints) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(56324)
		return (!z.InheritedLeasePreferences) == true
	}() == true)
}

func (z *ZoneConfig) InheritedVoterConstraints() bool {
	__antithesis_instrumentation__.Notify(56325)
	return len(z.VoterConstraints) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(56326)
		return !z.NullVoterConstraintsIsEmpty == true
	}() == true
}

func (z *ZoneConfig) ValidateTandemFields() error {
	__antithesis_instrumentation__.Notify(56327)
	var numConstrainedRepls int32
	numVotersExplicit := z.NumVoters != nil && func() bool {
		__antithesis_instrumentation__.Notify(56334)
		return *z.NumVoters > 0 == true
	}() == true
	for _, constraint := range z.Constraints {
		__antithesis_instrumentation__.Notify(56335)
		numConstrainedRepls += constraint.NumReplicas
	}
	__antithesis_instrumentation__.Notify(56328)

	if numConstrainedRepls > 0 && func() bool {
		__antithesis_instrumentation__.Notify(56336)
		return z.NumReplicas == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(56337)
		return fmt.Errorf("when per-replica constraints are set, num_replicas must be set as well")
	} else {
		__antithesis_instrumentation__.Notify(56338)
	}
	__antithesis_instrumentation__.Notify(56329)

	var numConstrainedVoters int32
	for _, constraint := range z.VoterConstraints {
		__antithesis_instrumentation__.Notify(56339)
		numConstrainedVoters += constraint.NumReplicas
	}
	__antithesis_instrumentation__.Notify(56330)

	if (numConstrainedVoters > 0 && func() bool {
		__antithesis_instrumentation__.Notify(56340)
		return z.NumVoters == nil == true
	}() == true) || func() bool {
		__antithesis_instrumentation__.Notify(56341)
		return (!numVotersExplicit && func() bool {
			__antithesis_instrumentation__.Notify(56342)
			return len(z.VoterConstraints) > 0 == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(56343)
		return fmt.Errorf("when voter_constraints are set, num_voters must be set as well")
	} else {
		__antithesis_instrumentation__.Notify(56344)
	}
	__antithesis_instrumentation__.Notify(56331)

	if (z.RangeMinBytes != nil || func() bool {
		__antithesis_instrumentation__.Notify(56345)
		return z.RangeMaxBytes != nil == true
	}() == true) && func() bool {
		__antithesis_instrumentation__.Notify(56346)
		return (z.RangeMinBytes == nil || func() bool {
			__antithesis_instrumentation__.Notify(56347)
			return z.RangeMaxBytes == nil == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(56348)
		return fmt.Errorf("range_min_bytes and range_max_bytes must be set together")
	} else {
		__antithesis_instrumentation__.Notify(56349)
	}
	__antithesis_instrumentation__.Notify(56332)
	if numVotersExplicit {
		__antithesis_instrumentation__.Notify(56350)
		if !z.InheritedLeasePreferences && func() bool {
			__antithesis_instrumentation__.Notify(56351)
			return z.InheritedVoterConstraints() == true
		}() == true {
			__antithesis_instrumentation__.Notify(56352)
			return fmt.Errorf("lease preferences can not be set unless the voter_constraints are explicitly set as well")
		} else {
			__antithesis_instrumentation__.Notify(56353)
		}
	} else {
		__antithesis_instrumentation__.Notify(56354)
		if !z.InheritedLeasePreferences && func() bool {
			__antithesis_instrumentation__.Notify(56355)
			return z.InheritedConstraints == true
		}() == true {
			__antithesis_instrumentation__.Notify(56356)
			return fmt.Errorf("lease preferences can not be set unless the constraints are explicitly set as well")
		} else {
			__antithesis_instrumentation__.Notify(56357)
		}
	}
	__antithesis_instrumentation__.Notify(56333)
	return nil
}

func (z *ZoneConfig) Validate() error {
	__antithesis_instrumentation__.Notify(56358)
	for _, s := range z.Subzones {
		__antithesis_instrumentation__.Notify(56372)
		if err := s.Config.Validate(); err != nil {
			__antithesis_instrumentation__.Notify(56373)
			return err
		} else {
			__antithesis_instrumentation__.Notify(56374)
		}
	}
	__antithesis_instrumentation__.Notify(56359)

	if z.NumReplicas != nil {
		__antithesis_instrumentation__.Notify(56375)
		switch {
		case *z.NumReplicas < 0:
			__antithesis_instrumentation__.Notify(56376)
			return fmt.Errorf("at least one replica is required")
		case *z.NumReplicas == 0:
			__antithesis_instrumentation__.Notify(56377)
			if len(z.Subzones) > 0 {
				__antithesis_instrumentation__.Notify(56381)

				return nil
			} else {
				__antithesis_instrumentation__.Notify(56382)
			}
			__antithesis_instrumentation__.Notify(56378)
			return fmt.Errorf("at least one replica is required")
		case *z.NumReplicas == 2:
			__antithesis_instrumentation__.Notify(56379)
			if !(z.NumVoters != nil && func() bool {
				__antithesis_instrumentation__.Notify(56383)
				return *z.NumVoters > 0 == true
			}() == true) {
				__antithesis_instrumentation__.Notify(56384)
				return fmt.Errorf("at least 3 replicas are required for multi-replica configurations")
			} else {
				__antithesis_instrumentation__.Notify(56385)
			}
		default:
			__antithesis_instrumentation__.Notify(56380)
		}
	} else {
		__antithesis_instrumentation__.Notify(56386)
	}
	__antithesis_instrumentation__.Notify(56360)

	var numVotersExplicit bool
	if z.NumVoters != nil {
		__antithesis_instrumentation__.Notify(56387)
		numVotersExplicit = true
		switch {
		case *z.NumVoters <= 0:
			__antithesis_instrumentation__.Notify(56389)
			return fmt.Errorf("at least one voting replica is required")
		case *z.NumVoters == 2:
			__antithesis_instrumentation__.Notify(56390)
			return fmt.Errorf("at least 3 voting replicas are required for multi-replica configurations")
		default:
			__antithesis_instrumentation__.Notify(56391)
		}
		__antithesis_instrumentation__.Notify(56388)
		if z.NumReplicas != nil && func() bool {
			__antithesis_instrumentation__.Notify(56392)
			return *z.NumVoters > *z.NumReplicas == true
		}() == true {
			__antithesis_instrumentation__.Notify(56393)
			return fmt.Errorf("num_voters cannot be greater than num_replicas")
		} else {
			__antithesis_instrumentation__.Notify(56394)
		}
	} else {
		__antithesis_instrumentation__.Notify(56395)
	}
	__antithesis_instrumentation__.Notify(56361)

	if z.RangeMaxBytes != nil && func() bool {
		__antithesis_instrumentation__.Notify(56396)
		return *z.RangeMaxBytes < base.MinRangeMaxBytes == true
	}() == true {
		__antithesis_instrumentation__.Notify(56397)
		return fmt.Errorf("RangeMaxBytes %d less than minimum allowed %d",
			*z.RangeMaxBytes, base.MinRangeMaxBytes)
	} else {
		__antithesis_instrumentation__.Notify(56398)
	}
	__antithesis_instrumentation__.Notify(56362)

	if z.RangeMinBytes != nil && func() bool {
		__antithesis_instrumentation__.Notify(56399)
		return *z.RangeMinBytes < 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(56400)
		return fmt.Errorf("RangeMinBytes %d less than minimum allowed 0", *z.RangeMinBytes)
	} else {
		__antithesis_instrumentation__.Notify(56401)
	}
	__antithesis_instrumentation__.Notify(56363)
	if z.RangeMinBytes != nil && func() bool {
		__antithesis_instrumentation__.Notify(56402)
		return z.RangeMaxBytes != nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(56403)
		return *z.RangeMinBytes >= *z.RangeMaxBytes == true
	}() == true {
		__antithesis_instrumentation__.Notify(56404)
		return fmt.Errorf("RangeMinBytes %d is greater than or equal to RangeMaxBytes %d",
			*z.RangeMinBytes, *z.RangeMaxBytes)
	} else {
		__antithesis_instrumentation__.Notify(56405)
	}
	__antithesis_instrumentation__.Notify(56364)

	if z.GC != nil && func() bool {
		__antithesis_instrumentation__.Notify(56406)
		return z.GC.TTLSeconds < 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(56407)
		return fmt.Errorf("GC.TTLSeconds %d less than minimum allowed 1", z.GC.TTLSeconds)
	} else {
		__antithesis_instrumentation__.Notify(56408)
	}
	__antithesis_instrumentation__.Notify(56365)

	for _, constraints := range z.Constraints {
		__antithesis_instrumentation__.Notify(56409)
		for _, constraint := range constraints.Constraints {
			__antithesis_instrumentation__.Notify(56410)
			if constraint.Type == Constraint_DEPRECATED_POSITIVE {
				__antithesis_instrumentation__.Notify(56411)
				return fmt.Errorf("constraints must either be required (prefixed with a '+') or " +
					"prohibited (prefixed with a '-')")
			} else {
				__antithesis_instrumentation__.Notify(56412)
			}
		}
	}
	__antithesis_instrumentation__.Notify(56366)

	for _, constraints := range z.VoterConstraints {
		__antithesis_instrumentation__.Notify(56413)
		for _, constraint := range constraints.Constraints {
			__antithesis_instrumentation__.Notify(56414)
			if constraint.Type == Constraint_DEPRECATED_POSITIVE {
				__antithesis_instrumentation__.Notify(56416)
				return fmt.Errorf("voter_constraints must be of type 'required' (prefixed with a '+')")
			} else {
				__antithesis_instrumentation__.Notify(56417)
			}
			__antithesis_instrumentation__.Notify(56415)

			if constraint.Type == Constraint_PROHIBITED {
				__antithesis_instrumentation__.Notify(56418)
				return fmt.Errorf("voter_constraints cannot contain prohibitive constraints")
			} else {
				__antithesis_instrumentation__.Notify(56419)
			}
		}
	}
	__antithesis_instrumentation__.Notify(56367)

	if len(z.Constraints) > 1 || func() bool {
		__antithesis_instrumentation__.Notify(56420)
		return (len(z.Constraints) == 1 && func() bool {
			__antithesis_instrumentation__.Notify(56421)
			return z.Constraints[0].NumReplicas != 0 == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(56422)
		var numConstrainedRepls int64
		for _, constraints := range z.Constraints {
			__antithesis_instrumentation__.Notify(56424)
			if constraints.NumReplicas <= 0 {
				__antithesis_instrumentation__.Notify(56426)
				return fmt.Errorf("constraints must apply to at least one replica")
			} else {
				__antithesis_instrumentation__.Notify(56427)
			}
			__antithesis_instrumentation__.Notify(56425)
			numConstrainedRepls += int64(constraints.NumReplicas)
			for _, constraint := range constraints.Constraints {
				__antithesis_instrumentation__.Notify(56428)

				if constraint.Type != Constraint_REQUIRED && func() bool {
					__antithesis_instrumentation__.Notify(56429)
					return z.NumReplicas != nil == true
				}() == true && func() bool {
					__antithesis_instrumentation__.Notify(56430)
					return constraints.NumReplicas != *z.NumReplicas == true
				}() == true {
					__antithesis_instrumentation__.Notify(56431)
					return fmt.Errorf(
						"only required constraints (prefixed with a '+') can be applied to a subset of replicas")
				} else {
					__antithesis_instrumentation__.Notify(56432)
				}
			}
		}
		__antithesis_instrumentation__.Notify(56423)
		if z.NumReplicas != nil && func() bool {
			__antithesis_instrumentation__.Notify(56433)
			return numConstrainedRepls > int64(*z.NumReplicas) == true
		}() == true {
			__antithesis_instrumentation__.Notify(56434)
			return fmt.Errorf("the number of replicas specified in constraints (%d) cannot be greater "+
				"than the number of replicas configured for the zone (%d)",
				numConstrainedRepls, *z.NumReplicas)
		} else {
			__antithesis_instrumentation__.Notify(56435)
		}
	} else {
		__antithesis_instrumentation__.Notify(56436)
	}
	__antithesis_instrumentation__.Notify(56368)

	if numVotersExplicit {
		__antithesis_instrumentation__.Notify(56437)
		if len(z.VoterConstraints) > 1 || func() bool {
			__antithesis_instrumentation__.Notify(56438)
			return (len(z.VoterConstraints) == 1 && func() bool {
				__antithesis_instrumentation__.Notify(56439)
				return z.VoterConstraints[0].NumReplicas != 0 == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(56440)
			var numConstrainedRepls int64
			for _, constraints := range z.VoterConstraints {
				__antithesis_instrumentation__.Notify(56442)
				if constraints.NumReplicas <= 0 {
					__antithesis_instrumentation__.Notify(56444)
					return fmt.Errorf("constraints must apply to at least one replica")
				} else {
					__antithesis_instrumentation__.Notify(56445)
				}
				__antithesis_instrumentation__.Notify(56443)
				numConstrainedRepls += int64(constraints.NumReplicas)
			}
			__antithesis_instrumentation__.Notify(56441)

			if z.NumVoters != nil && func() bool {
				__antithesis_instrumentation__.Notify(56446)
				return numConstrainedRepls > int64(*z.NumVoters) == true
			}() == true {
				__antithesis_instrumentation__.Notify(56447)
				return fmt.Errorf("the number of replicas specified in voter_constraints (%d) cannot be greater "+
					"than the number of voters configured for the zone (%d)",
					numConstrainedRepls, *z.NumVoters)
			} else {
				__antithesis_instrumentation__.Notify(56448)
			}
		} else {
			__antithesis_instrumentation__.Notify(56449)
		}
	} else {
		__antithesis_instrumentation__.Notify(56450)
	}
	__antithesis_instrumentation__.Notify(56369)

	if err := validateVoterConstraintsCompatibility(z.VoterConstraints, z.Constraints); err != nil {
		__antithesis_instrumentation__.Notify(56451)
		return err
	} else {
		__antithesis_instrumentation__.Notify(56452)
	}
	__antithesis_instrumentation__.Notify(56370)

	for _, leasePref := range z.LeasePreferences {
		__antithesis_instrumentation__.Notify(56453)
		if len(leasePref.Constraints) == 0 {
			__antithesis_instrumentation__.Notify(56455)
			return fmt.Errorf("every lease preference must include at least one constraint")
		} else {
			__antithesis_instrumentation__.Notify(56456)
		}
		__antithesis_instrumentation__.Notify(56454)
		for _, constraint := range leasePref.Constraints {
			__antithesis_instrumentation__.Notify(56457)
			if constraint.Type == Constraint_DEPRECATED_POSITIVE {
				__antithesis_instrumentation__.Notify(56458)
				return fmt.Errorf("lease preference constraints must either be required " +
					"(prefixed with a '+') or prohibited (prefixed with a '-')")
			} else {
				__antithesis_instrumentation__.Notify(56459)
			}
		}
	}
	__antithesis_instrumentation__.Notify(56371)

	return nil
}

func validateVoterConstraintsCompatibility(
	voterConstraints, overallConstraints []ConstraintsConjunction,
) error {
	__antithesis_instrumentation__.Notify(56460)

	for _, constraints := range overallConstraints {
		__antithesis_instrumentation__.Notify(56462)
		for _, constraint := range constraints.Constraints {
			__antithesis_instrumentation__.Notify(56463)
			if constraint.Type == Constraint_PROHIBITED {
				__antithesis_instrumentation__.Notify(56464)
				for _, otherConstraints := range voterConstraints {
					__antithesis_instrumentation__.Notify(56465)
					for _, otherConstraint := range otherConstraints.Constraints {
						__antithesis_instrumentation__.Notify(56466)
						conflicting := otherConstraint.Value == constraint.Value && func() bool {
							__antithesis_instrumentation__.Notify(56467)
							return otherConstraint.Key == constraint.Key == true
						}() == true
						if conflicting {
							__antithesis_instrumentation__.Notify(56468)
							return fmt.Errorf("prohibitive constraint %s conflicts with voter_constraint %s", constraint, otherConstraint)
						} else {
							__antithesis_instrumentation__.Notify(56469)
						}
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(56470)
			}
		}
	}
	__antithesis_instrumentation__.Notify(56461)
	return nil
}

func (z *ZoneConfig) InheritFromParent(parent *ZoneConfig) {
	__antithesis_instrumentation__.Notify(56471)

	if z.NumReplicas == nil || func() bool {
		__antithesis_instrumentation__.Notify(56480)
		return (z.NumReplicas != nil && func() bool {
			__antithesis_instrumentation__.Notify(56481)
			return *z.NumReplicas == 0 == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(56482)
		if parent.NumReplicas != nil {
			__antithesis_instrumentation__.Notify(56483)
			z.NumReplicas = proto.Int32(*parent.NumReplicas)
		} else {
			__antithesis_instrumentation__.Notify(56484)
		}
	} else {
		__antithesis_instrumentation__.Notify(56485)
	}
	__antithesis_instrumentation__.Notify(56472)
	if z.NumVoters == nil || func() bool {
		__antithesis_instrumentation__.Notify(56486)
		return (z.NumVoters != nil && func() bool {
			__antithesis_instrumentation__.Notify(56487)
			return *z.NumVoters == 0 == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(56488)
		if parent.NumVoters != nil {
			__antithesis_instrumentation__.Notify(56489)
			z.NumVoters = proto.Int32(*parent.NumVoters)
		} else {
			__antithesis_instrumentation__.Notify(56490)
		}
	} else {
		__antithesis_instrumentation__.Notify(56491)
	}
	__antithesis_instrumentation__.Notify(56473)
	if z.GlobalReads == nil {
		__antithesis_instrumentation__.Notify(56492)
		if parent.GlobalReads != nil {
			__antithesis_instrumentation__.Notify(56493)
			z.GlobalReads = proto.Bool(*parent.GlobalReads)
		} else {
			__antithesis_instrumentation__.Notify(56494)
		}
	} else {
		__antithesis_instrumentation__.Notify(56495)
	}
	__antithesis_instrumentation__.Notify(56474)
	if z.RangeMinBytes == nil {
		__antithesis_instrumentation__.Notify(56496)
		if parent.RangeMinBytes != nil {
			__antithesis_instrumentation__.Notify(56497)
			z.RangeMinBytes = proto.Int64(*parent.RangeMinBytes)
		} else {
			__antithesis_instrumentation__.Notify(56498)
		}
	} else {
		__antithesis_instrumentation__.Notify(56499)
	}
	__antithesis_instrumentation__.Notify(56475)
	if z.RangeMaxBytes == nil {
		__antithesis_instrumentation__.Notify(56500)
		if parent.RangeMaxBytes != nil {
			__antithesis_instrumentation__.Notify(56501)
			z.RangeMaxBytes = proto.Int64(*parent.RangeMaxBytes)
		} else {
			__antithesis_instrumentation__.Notify(56502)
		}
	} else {
		__antithesis_instrumentation__.Notify(56503)
	}
	__antithesis_instrumentation__.Notify(56476)
	if z.GC == nil {
		__antithesis_instrumentation__.Notify(56504)
		if parent.GC != nil {
			__antithesis_instrumentation__.Notify(56505)
			tempGC := *parent.GC
			z.GC = &tempGC
		} else {
			__antithesis_instrumentation__.Notify(56506)
		}
	} else {
		__antithesis_instrumentation__.Notify(56507)
	}
	__antithesis_instrumentation__.Notify(56477)
	if z.InheritedConstraints {
		__antithesis_instrumentation__.Notify(56508)
		if !parent.InheritedConstraints {
			__antithesis_instrumentation__.Notify(56509)
			z.Constraints = parent.Constraints
			z.InheritedConstraints = false
		} else {
			__antithesis_instrumentation__.Notify(56510)
		}
	} else {
		__antithesis_instrumentation__.Notify(56511)
	}
	__antithesis_instrumentation__.Notify(56478)
	if z.InheritedVoterConstraints() {
		__antithesis_instrumentation__.Notify(56512)
		if !parent.InheritedVoterConstraints() {
			__antithesis_instrumentation__.Notify(56513)
			z.VoterConstraints = parent.VoterConstraints
			z.NullVoterConstraintsIsEmpty = parent.NullVoterConstraintsIsEmpty
		} else {
			__antithesis_instrumentation__.Notify(56514)
		}
	} else {
		__antithesis_instrumentation__.Notify(56515)
	}
	__antithesis_instrumentation__.Notify(56479)
	if z.InheritedLeasePreferences {
		__antithesis_instrumentation__.Notify(56516)
		if !parent.InheritedLeasePreferences {
			__antithesis_instrumentation__.Notify(56517)
			z.LeasePreferences = parent.LeasePreferences
			z.InheritedLeasePreferences = false
		} else {
			__antithesis_instrumentation__.Notify(56518)
		}
	} else {
		__antithesis_instrumentation__.Notify(56519)
	}
}

func (z *ZoneConfig) CopyFromZone(other ZoneConfig, fieldList []tree.Name) {
	__antithesis_instrumentation__.Notify(56520)
	for _, fieldName := range fieldList {
		__antithesis_instrumentation__.Notify(56521)
		switch fieldName {
		case "num_replicas":
			__antithesis_instrumentation__.Notify(56522)
			z.NumReplicas = nil
			if other.NumReplicas != nil {
				__antithesis_instrumentation__.Notify(56532)
				z.NumReplicas = proto.Int32(*other.NumReplicas)
			} else {
				__antithesis_instrumentation__.Notify(56533)
			}
		case "num_voters":
			__antithesis_instrumentation__.Notify(56523)
			z.NumVoters = nil
			if other.NumVoters != nil {
				__antithesis_instrumentation__.Notify(56534)
				z.NumVoters = proto.Int32(*other.NumVoters)
			} else {
				__antithesis_instrumentation__.Notify(56535)
			}
		case "range_min_bytes":
			__antithesis_instrumentation__.Notify(56524)
			z.RangeMinBytes = nil
			if other.RangeMinBytes != nil {
				__antithesis_instrumentation__.Notify(56536)
				z.RangeMinBytes = proto.Int64(*other.RangeMinBytes)
			} else {
				__antithesis_instrumentation__.Notify(56537)
			}
		case "range_max_bytes":
			__antithesis_instrumentation__.Notify(56525)
			z.RangeMaxBytes = nil
			if other.RangeMaxBytes != nil {
				__antithesis_instrumentation__.Notify(56538)
				z.RangeMaxBytes = proto.Int64(*other.RangeMaxBytes)
			} else {
				__antithesis_instrumentation__.Notify(56539)
			}
		case "global_reads":
			__antithesis_instrumentation__.Notify(56526)
			z.GlobalReads = nil
			if other.GlobalReads != nil {
				__antithesis_instrumentation__.Notify(56540)
				z.GlobalReads = proto.Bool(*other.GlobalReads)
			} else {
				__antithesis_instrumentation__.Notify(56541)
			}
		case "gc.ttlseconds":
			__antithesis_instrumentation__.Notify(56527)
			z.GC = nil
			if other.GC != nil {
				__antithesis_instrumentation__.Notify(56542)
				tempGC := *other.GC
				z.GC = &tempGC
			} else {
				__antithesis_instrumentation__.Notify(56543)
			}
		case "constraints":
			__antithesis_instrumentation__.Notify(56528)
			z.Constraints = other.Constraints
			z.InheritedConstraints = other.InheritedConstraints
		case "voter_constraints":
			__antithesis_instrumentation__.Notify(56529)
			z.VoterConstraints = other.VoterConstraints
			z.NullVoterConstraintsIsEmpty = other.NullVoterConstraintsIsEmpty
		case "lease_preferences":
			__antithesis_instrumentation__.Notify(56530)
			z.LeasePreferences = other.LeasePreferences
			z.InheritedLeasePreferences = other.InheritedLeasePreferences
		default:
			__antithesis_instrumentation__.Notify(56531)
		}
	}
}

type DiffWithZoneMismatch struct {
	IndexID uint32

	PartitionName string

	IsMissingSubzone bool

	IsExtraSubzone bool

	Field string
}

func (z *ZoneConfig) DiffWithZone(
	other ZoneConfig, fieldList []tree.Name,
) (bool, DiffWithZoneMismatch, error) {
	__antithesis_instrumentation__.Notify(56544)
	mismatchingNumReplicas := false
	for _, fieldName := range fieldList {
		__antithesis_instrumentation__.Notify(56550)
		switch fieldName {
		case "num_replicas":
			__antithesis_instrumentation__.Notify(56551)
			if other.NumReplicas == nil && func() bool {
				__antithesis_instrumentation__.Notify(56573)
				return z.NumReplicas == nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(56574)
				continue
			} else {
				__antithesis_instrumentation__.Notify(56575)
			}
			__antithesis_instrumentation__.Notify(56552)
			if z.NumReplicas == nil || func() bool {
				__antithesis_instrumentation__.Notify(56576)
				return other.NumReplicas == nil == true
			}() == true || func() bool {
				__antithesis_instrumentation__.Notify(56577)
				return *z.NumReplicas != *other.NumReplicas == true
			}() == true {
				__antithesis_instrumentation__.Notify(56578)

				if z.IsSubzonePlaceholder() || func() bool {
					__antithesis_instrumentation__.Notify(56580)
					return other.IsSubzonePlaceholder() == true
				}() == true {
					__antithesis_instrumentation__.Notify(56581)
					mismatchingNumReplicas = true
					continue
				} else {
					__antithesis_instrumentation__.Notify(56582)
				}
				__antithesis_instrumentation__.Notify(56579)
				return false, DiffWithZoneMismatch{
					Field: "num_replicas",
				}, nil
			} else {
				__antithesis_instrumentation__.Notify(56583)
			}
		case "num_voters":
			__antithesis_instrumentation__.Notify(56553)
			if other.NumVoters == nil && func() bool {
				__antithesis_instrumentation__.Notify(56584)
				return z.NumVoters == nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(56585)
				continue
			} else {
				__antithesis_instrumentation__.Notify(56586)
			}
			__antithesis_instrumentation__.Notify(56554)
			if z.NumVoters == nil || func() bool {
				__antithesis_instrumentation__.Notify(56587)
				return other.NumVoters == nil == true
			}() == true || func() bool {
				__antithesis_instrumentation__.Notify(56588)
				return *z.NumVoters != *other.NumVoters == true
			}() == true {
				__antithesis_instrumentation__.Notify(56589)
				return false, DiffWithZoneMismatch{
					Field: "num_voters",
				}, nil
			} else {
				__antithesis_instrumentation__.Notify(56590)
			}
		case "range_min_bytes":
			__antithesis_instrumentation__.Notify(56555)
			if other.RangeMinBytes == nil && func() bool {
				__antithesis_instrumentation__.Notify(56591)
				return z.RangeMinBytes == nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(56592)
				continue
			} else {
				__antithesis_instrumentation__.Notify(56593)
			}
			__antithesis_instrumentation__.Notify(56556)
			if z.RangeMinBytes == nil || func() bool {
				__antithesis_instrumentation__.Notify(56594)
				return other.RangeMinBytes == nil == true
			}() == true || func() bool {
				__antithesis_instrumentation__.Notify(56595)
				return *z.RangeMinBytes != *other.RangeMinBytes == true
			}() == true {
				__antithesis_instrumentation__.Notify(56596)
				return false, DiffWithZoneMismatch{
					Field: "range_min_bytes",
				}, nil
			} else {
				__antithesis_instrumentation__.Notify(56597)
			}
		case "range_max_bytes":
			__antithesis_instrumentation__.Notify(56557)
			if other.RangeMaxBytes == nil && func() bool {
				__antithesis_instrumentation__.Notify(56598)
				return z.RangeMaxBytes == nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(56599)
				continue
			} else {
				__antithesis_instrumentation__.Notify(56600)
			}
			__antithesis_instrumentation__.Notify(56558)
			if z.RangeMaxBytes == nil || func() bool {
				__antithesis_instrumentation__.Notify(56601)
				return other.RangeMaxBytes == nil == true
			}() == true || func() bool {
				__antithesis_instrumentation__.Notify(56602)
				return *z.RangeMaxBytes != *other.RangeMaxBytes == true
			}() == true {
				__antithesis_instrumentation__.Notify(56603)
				return false, DiffWithZoneMismatch{
					Field: "range_max_bytes",
				}, nil
			} else {
				__antithesis_instrumentation__.Notify(56604)
			}
		case "global_reads":
			__antithesis_instrumentation__.Notify(56559)
			if other.GlobalReads == nil && func() bool {
				__antithesis_instrumentation__.Notify(56605)
				return z.GlobalReads == nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(56606)
				continue
			} else {
				__antithesis_instrumentation__.Notify(56607)
			}
			__antithesis_instrumentation__.Notify(56560)
			if z.GlobalReads == nil || func() bool {
				__antithesis_instrumentation__.Notify(56608)
				return other.GlobalReads == nil == true
			}() == true || func() bool {
				__antithesis_instrumentation__.Notify(56609)
				return *z.GlobalReads != *other.GlobalReads == true
			}() == true {
				__antithesis_instrumentation__.Notify(56610)
				return false, DiffWithZoneMismatch{
					Field: "global_reads",
				}, nil
			} else {
				__antithesis_instrumentation__.Notify(56611)
			}
		case "gc.ttlseconds":
			__antithesis_instrumentation__.Notify(56561)
			if other.GC == nil && func() bool {
				__antithesis_instrumentation__.Notify(56612)
				return z.GC == nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(56613)
				continue
			} else {
				__antithesis_instrumentation__.Notify(56614)
			}
			__antithesis_instrumentation__.Notify(56562)
			if z.GC == nil || func() bool {
				__antithesis_instrumentation__.Notify(56615)
				return other.GC == nil == true
			}() == true || func() bool {
				__antithesis_instrumentation__.Notify(56616)
				return *z.GC != *other.GC == true
			}() == true {
				__antithesis_instrumentation__.Notify(56617)
				return false, DiffWithZoneMismatch{
					Field: "gc.ttlseconds",
				}, nil
			} else {
				__antithesis_instrumentation__.Notify(56618)
			}
		case "constraints":
			__antithesis_instrumentation__.Notify(56563)
			if other.Constraints == nil && func() bool {
				__antithesis_instrumentation__.Notify(56619)
				return z.Constraints == nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(56620)
				continue
			} else {
				__antithesis_instrumentation__.Notify(56621)
			}
			__antithesis_instrumentation__.Notify(56564)
			if z.Constraints == nil || func() bool {
				__antithesis_instrumentation__.Notify(56622)
				return other.Constraints == nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(56623)
				return false, DiffWithZoneMismatch{
					Field: "constraints",
				}, nil
			} else {
				__antithesis_instrumentation__.Notify(56624)
			}
			__antithesis_instrumentation__.Notify(56565)
			for i, c := range z.Constraints {
				__antithesis_instrumentation__.Notify(56625)
				for j, constraint := range c.Constraints {
					__antithesis_instrumentation__.Notify(56626)
					if len(other.Constraints) <= i || func() bool {
						__antithesis_instrumentation__.Notify(56627)
						return len(other.Constraints[i].Constraints) <= j == true
					}() == true || func() bool {
						__antithesis_instrumentation__.Notify(56628)
						return constraint != other.Constraints[i].Constraints[j] == true
					}() == true {
						__antithesis_instrumentation__.Notify(56629)
						return false, DiffWithZoneMismatch{
							Field: "constraints",
						}, nil
					} else {
						__antithesis_instrumentation__.Notify(56630)
					}
				}
			}
		case "voter_constraints":
			__antithesis_instrumentation__.Notify(56566)
			if other.VoterConstraints == nil && func() bool {
				__antithesis_instrumentation__.Notify(56631)
				return z.VoterConstraints == nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(56632)
				continue
			} else {
				__antithesis_instrumentation__.Notify(56633)
			}
			__antithesis_instrumentation__.Notify(56567)
			if z.VoterConstraints == nil || func() bool {
				__antithesis_instrumentation__.Notify(56634)
				return other.VoterConstraints == nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(56635)
				return false, DiffWithZoneMismatch{
					Field: "voter_constraints",
				}, nil
			} else {
				__antithesis_instrumentation__.Notify(56636)
			}
			__antithesis_instrumentation__.Notify(56568)
			for i, c := range z.VoterConstraints {
				__antithesis_instrumentation__.Notify(56637)
				for j, constraint := range c.Constraints {
					__antithesis_instrumentation__.Notify(56638)
					if len(other.VoterConstraints) <= i || func() bool {
						__antithesis_instrumentation__.Notify(56639)
						return len(other.VoterConstraints[i].Constraints) <= j == true
					}() == true || func() bool {
						__antithesis_instrumentation__.Notify(56640)
						return constraint != other.VoterConstraints[i].Constraints[j] == true
					}() == true {
						__antithesis_instrumentation__.Notify(56641)
						return false, DiffWithZoneMismatch{
							Field: "voter_constraints",
						}, nil
					} else {
						__antithesis_instrumentation__.Notify(56642)
					}
				}
			}
		case "lease_preferences":
			__antithesis_instrumentation__.Notify(56569)
			if other.LeasePreferences == nil && func() bool {
				__antithesis_instrumentation__.Notify(56643)
				return z.LeasePreferences == nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(56644)
				continue
			} else {
				__antithesis_instrumentation__.Notify(56645)
			}
			__antithesis_instrumentation__.Notify(56570)
			if z.LeasePreferences == nil || func() bool {
				__antithesis_instrumentation__.Notify(56646)
				return other.LeasePreferences == nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(56647)
				return false, DiffWithZoneMismatch{
					Field: "voter_constraints",
				}, nil
			} else {
				__antithesis_instrumentation__.Notify(56648)
			}
			__antithesis_instrumentation__.Notify(56571)
			for i, c := range z.LeasePreferences {
				__antithesis_instrumentation__.Notify(56649)
				for j, constraint := range c.Constraints {
					__antithesis_instrumentation__.Notify(56650)
					if len(other.LeasePreferences) <= i || func() bool {
						__antithesis_instrumentation__.Notify(56651)
						return len(other.LeasePreferences[i].Constraints) <= j == true
					}() == true || func() bool {
						__antithesis_instrumentation__.Notify(56652)
						return constraint != other.LeasePreferences[i].Constraints[j] == true
					}() == true {
						__antithesis_instrumentation__.Notify(56653)
						return false, DiffWithZoneMismatch{
							Field: "lease_preferences",
						}, nil
					} else {
						__antithesis_instrumentation__.Notify(56654)
					}
				}
			}
		default:
			__antithesis_instrumentation__.Notify(56572)
			return false, DiffWithZoneMismatch{}, errors.AssertionFailedf("unknown zone configuration field %q", fieldName)
		}
	}
	__antithesis_instrumentation__.Notify(56545)

	type subzoneKey struct {
		indexID       uint32
		partitionName string
	}
	otherSubzonesBySubzoneKey := make(map[subzoneKey]Subzone, len(other.Subzones))
	for _, o := range other.Subzones {
		__antithesis_instrumentation__.Notify(56655)
		k := subzoneKey{indexID: o.IndexID, partitionName: o.PartitionName}
		otherSubzonesBySubzoneKey[k] = o
	}
	__antithesis_instrumentation__.Notify(56546)
	for _, s := range z.Subzones {
		__antithesis_instrumentation__.Notify(56656)
		k := subzoneKey{indexID: s.IndexID, partitionName: s.PartitionName}
		o, found := otherSubzonesBySubzoneKey[k]
		if !found {
			__antithesis_instrumentation__.Notify(56659)

			if b, subzoneMismatch, err := s.Config.DiffWithZone(
				*NewZoneConfig(),
				fieldList,
			); err != nil {
				__antithesis_instrumentation__.Notify(56661)
				return b, subzoneMismatch, err
			} else {
				__antithesis_instrumentation__.Notify(56662)
				if !b {
					__antithesis_instrumentation__.Notify(56663)
					return false, DiffWithZoneMismatch{
						IndexID:        s.IndexID,
						PartitionName:  s.PartitionName,
						IsExtraSubzone: true,
						Field:          subzoneMismatch.Field,
					}, nil
				} else {
					__antithesis_instrumentation__.Notify(56664)
				}
			}
			__antithesis_instrumentation__.Notify(56660)
			continue
		} else {
			__antithesis_instrumentation__.Notify(56665)
		}
		__antithesis_instrumentation__.Notify(56657)
		if b, subzoneMismatch, err := s.Config.DiffWithZone(
			o.Config,
			fieldList,
		); err != nil {
			__antithesis_instrumentation__.Notify(56666)
			return b, subzoneMismatch, err
		} else {
			__antithesis_instrumentation__.Notify(56667)
			if !b {
				__antithesis_instrumentation__.Notify(56668)

				if subzoneMismatch.IndexID > 0 {
					__antithesis_instrumentation__.Notify(56670)
					return false, DiffWithZoneMismatch{}, errors.AssertionFailedf(
						"unexpected subzone index id %d",
						subzoneMismatch.IndexID,
					)
				} else {
					__antithesis_instrumentation__.Notify(56671)
				}
				__antithesis_instrumentation__.Notify(56669)
				return b, DiffWithZoneMismatch{
					IndexID:       o.IndexID,
					PartitionName: o.PartitionName,
					Field:         subzoneMismatch.Field,
				}, nil
			} else {
				__antithesis_instrumentation__.Notify(56672)
			}
		}
		__antithesis_instrumentation__.Notify(56658)
		delete(otherSubzonesBySubzoneKey, k)
	}
	__antithesis_instrumentation__.Notify(56547)

	for _, o := range otherSubzonesBySubzoneKey {
		__antithesis_instrumentation__.Notify(56673)
		if b, subzoneMismatch, err := NewZoneConfig().DiffWithZone(
			o.Config,
			fieldList,
		); err != nil {
			__antithesis_instrumentation__.Notify(56674)
			return b, subzoneMismatch, err
		} else {
			__antithesis_instrumentation__.Notify(56675)
			if !b {
				__antithesis_instrumentation__.Notify(56676)
				return false, DiffWithZoneMismatch{
					IndexID:          o.IndexID,
					PartitionName:    o.PartitionName,
					IsMissingSubzone: true,
					Field:            subzoneMismatch.Field,
				}, nil
			} else {
				__antithesis_instrumentation__.Notify(56677)
			}
		}
	}
	__antithesis_instrumentation__.Notify(56548)

	if mismatchingNumReplicas {
		__antithesis_instrumentation__.Notify(56678)
		return false, DiffWithZoneMismatch{
			Field: "num_replicas",
		}, nil
	} else {
		__antithesis_instrumentation__.Notify(56679)
	}
	__antithesis_instrumentation__.Notify(56549)
	return true, DiffWithZoneMismatch{}, nil
}

func (z *ZoneConfig) ClearFieldsOfAllSubzones(fieldList []tree.Name) {
	__antithesis_instrumentation__.Notify(56680)
	newSubzones := z.Subzones[:0]
	emptyZone := NewZoneConfig()
	for _, sz := range z.Subzones {
		__antithesis_instrumentation__.Notify(56682)

		sz.Config.CopyFromZone(*emptyZone, fieldList)

		if !sz.Config.Equal(emptyZone) {
			__antithesis_instrumentation__.Notify(56683)
			newSubzones = append(newSubzones, sz)
		} else {
			__antithesis_instrumentation__.Notify(56684)
		}
	}
	__antithesis_instrumentation__.Notify(56681)
	z.Subzones = newSubzones
}

func StoreSatisfiesConstraint(store roachpb.StoreDescriptor, constraint Constraint) bool {
	__antithesis_instrumentation__.Notify(56685)
	hasConstraint := StoreMatchesConstraint(store, constraint)
	if (constraint.Type == Constraint_REQUIRED && func() bool {
		__antithesis_instrumentation__.Notify(56687)
		return !hasConstraint == true
	}() == true) || func() bool {
		__antithesis_instrumentation__.Notify(56688)
		return (constraint.Type == Constraint_PROHIBITED && func() bool {
			__antithesis_instrumentation__.Notify(56689)
			return hasConstraint == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(56690)
		return false
	} else {
		__antithesis_instrumentation__.Notify(56691)
	}
	__antithesis_instrumentation__.Notify(56686)
	return true
}

func StoreMatchesConstraint(store roachpb.StoreDescriptor, c Constraint) bool {
	__antithesis_instrumentation__.Notify(56692)
	if c.Key == "" {
		__antithesis_instrumentation__.Notify(56695)
		for _, attrs := range []roachpb.Attributes{store.Attrs, store.Node.Attrs} {
			__antithesis_instrumentation__.Notify(56697)
			for _, attr := range attrs.Attrs {
				__antithesis_instrumentation__.Notify(56698)
				if attr == c.Value {
					__antithesis_instrumentation__.Notify(56699)
					return true
				} else {
					__antithesis_instrumentation__.Notify(56700)
				}
			}
		}
		__antithesis_instrumentation__.Notify(56696)
		return false
	} else {
		__antithesis_instrumentation__.Notify(56701)
	}
	__antithesis_instrumentation__.Notify(56693)
	for _, tier := range store.Node.Locality.Tiers {
		__antithesis_instrumentation__.Notify(56702)
		if c.Key == tier.Key && func() bool {
			__antithesis_instrumentation__.Notify(56703)
			return c.Value == tier.Value == true
		}() == true {
			__antithesis_instrumentation__.Notify(56704)
			return true
		} else {
			__antithesis_instrumentation__.Notify(56705)
		}
	}
	__antithesis_instrumentation__.Notify(56694)
	return false
}

func (z *ZoneConfig) DeleteTableConfig() {
	__antithesis_instrumentation__.Notify(56706)
	*z = ZoneConfig{

		NumReplicas:  proto.Int32(0),
		Subzones:     z.Subzones,
		SubzoneSpans: z.SubzoneSpans,
	}
}

func (z *ZoneConfig) IsSubzonePlaceholder() bool {
	__antithesis_instrumentation__.Notify(56707)

	return z.NumReplicas != nil && func() bool {
		__antithesis_instrumentation__.Notify(56708)
		return *z.NumReplicas == 0 == true
	}() == true
}

func (z *ZoneConfig) GetSubzone(indexID uint32, partition string) *Subzone {
	__antithesis_instrumentation__.Notify(56709)
	for _, s := range z.Subzones {
		__antithesis_instrumentation__.Notify(56712)
		if s.IndexID == indexID && func() bool {
			__antithesis_instrumentation__.Notify(56713)
			return s.PartitionName == partition == true
		}() == true {
			__antithesis_instrumentation__.Notify(56714)
			copySubzone := s
			return &copySubzone
		} else {
			__antithesis_instrumentation__.Notify(56715)
		}
	}
	__antithesis_instrumentation__.Notify(56710)
	if partition != "" {
		__antithesis_instrumentation__.Notify(56716)
		return z.GetSubzone(indexID, "")
	} else {
		__antithesis_instrumentation__.Notify(56717)
	}
	__antithesis_instrumentation__.Notify(56711)
	return nil
}

func (z *ZoneConfig) GetSubzoneExact(indexID uint32, partition string) *Subzone {
	__antithesis_instrumentation__.Notify(56718)
	for _, s := range z.Subzones {
		__antithesis_instrumentation__.Notify(56720)
		if s.IndexID == indexID && func() bool {
			__antithesis_instrumentation__.Notify(56721)
			return s.PartitionName == partition == true
		}() == true {
			__antithesis_instrumentation__.Notify(56722)
			copySubzone := s
			return &copySubzone
		} else {
			__antithesis_instrumentation__.Notify(56723)
		}
	}
	__antithesis_instrumentation__.Notify(56719)
	return nil
}

func (z ZoneConfig) GetSubzoneForKeySuffix(keySuffix []byte) (*Subzone, int32) {
	__antithesis_instrumentation__.Notify(56724)

	for _, s := range z.SubzoneSpans {
		__antithesis_instrumentation__.Notify(56726)

		if (s.Key.Compare(keySuffix) <= 0) && func() bool {
			__antithesis_instrumentation__.Notify(56727)
			return ((s.EndKey == nil && func() bool {
				__antithesis_instrumentation__.Notify(56728)
				return bytes.HasPrefix(keySuffix, s.Key) == true
			}() == true) || func() bool {
				__antithesis_instrumentation__.Notify(56729)
				return s.EndKey.Compare(keySuffix) > 0 == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(56730)
			copySubzone := z.Subzones[s.SubzoneIndex]
			return &copySubzone, s.SubzoneIndex
		} else {
			__antithesis_instrumentation__.Notify(56731)
		}
	}
	__antithesis_instrumentation__.Notify(56725)
	return nil, -1
}

func (z *ZoneConfig) SetSubzone(subzone Subzone) {
	__antithesis_instrumentation__.Notify(56732)
	for i, s := range z.Subzones {
		__antithesis_instrumentation__.Notify(56734)
		if s.IndexID == subzone.IndexID && func() bool {
			__antithesis_instrumentation__.Notify(56735)
			return s.PartitionName == subzone.PartitionName == true
		}() == true {
			__antithesis_instrumentation__.Notify(56736)
			z.Subzones[i] = subzone
			return
		} else {
			__antithesis_instrumentation__.Notify(56737)
		}
	}
	__antithesis_instrumentation__.Notify(56733)
	z.Subzones = append(z.Subzones, subzone)
}

func (z *ZoneConfig) DeleteSubzone(indexID uint32, partition string) bool {
	__antithesis_instrumentation__.Notify(56738)
	for i, s := range z.Subzones {
		__antithesis_instrumentation__.Notify(56740)
		if s.IndexID == indexID && func() bool {
			__antithesis_instrumentation__.Notify(56741)
			return s.PartitionName == partition == true
		}() == true {
			__antithesis_instrumentation__.Notify(56742)
			z.Subzones = append(z.Subzones[:i], z.Subzones[i+1:]...)
			return true
		} else {
			__antithesis_instrumentation__.Notify(56743)
		}
	}
	__antithesis_instrumentation__.Notify(56739)
	return false
}

func (z *ZoneConfig) DeleteIndexSubzones(indexID uint32) {
	__antithesis_instrumentation__.Notify(56744)
	subzones := z.Subzones[:0]
	for _, s := range z.Subzones {
		__antithesis_instrumentation__.Notify(56746)
		if s.IndexID != indexID {
			__antithesis_instrumentation__.Notify(56747)
			subzones = append(subzones, s)
		} else {
			__antithesis_instrumentation__.Notify(56748)
		}
	}
	__antithesis_instrumentation__.Notify(56745)
	z.Subzones = subzones
}

func (z ZoneConfig) SubzoneSplits() []roachpb.RKey {
	__antithesis_instrumentation__.Notify(56749)
	var out []roachpb.RKey
	for _, span := range z.SubzoneSpans {
		__antithesis_instrumentation__.Notify(56751)

		if len(out) == 0 || func() bool {
			__antithesis_instrumentation__.Notify(56754)
			return !out[len(out)-1].Equal(span.Key) == true
		}() == true {
			__antithesis_instrumentation__.Notify(56755)

			out = append(out, roachpb.RKey(span.Key))
		} else {
			__antithesis_instrumentation__.Notify(56756)
		}
		__antithesis_instrumentation__.Notify(56752)
		endKey := span.EndKey
		if len(endKey) == 0 {
			__antithesis_instrumentation__.Notify(56757)
			endKey = span.Key.PrefixEnd()
		} else {
			__antithesis_instrumentation__.Notify(56758)
		}
		__antithesis_instrumentation__.Notify(56753)
		out = append(out, roachpb.RKey(endKey))

	}
	__antithesis_instrumentation__.Notify(56750)
	return out
}

func (c ConstraintsConjunction) String() string {
	__antithesis_instrumentation__.Notify(56759)
	var sb strings.Builder
	for i, cons := range c.Constraints {
		__antithesis_instrumentation__.Notify(56762)
		if i > 0 {
			__antithesis_instrumentation__.Notify(56764)
			sb.WriteRune(',')
		} else {
			__antithesis_instrumentation__.Notify(56765)
		}
		__antithesis_instrumentation__.Notify(56763)
		sb.WriteString(cons.String())
	}
	__antithesis_instrumentation__.Notify(56760)
	if c.NumReplicas != 0 {
		__antithesis_instrumentation__.Notify(56766)
		fmt.Fprintf(&sb, ":%d", c.NumReplicas)
	} else {
		__antithesis_instrumentation__.Notify(56767)
	}
	__antithesis_instrumentation__.Notify(56761)
	return sb.String()
}

func (z *ZoneConfig) EnsureFullyHydrated() error {
	__antithesis_instrumentation__.Notify(56768)
	var unsetFields []string
	if z.RangeMaxBytes == nil {
		__antithesis_instrumentation__.Notify(56774)
		unsetFields = append(unsetFields, "RangeMaxBytes")
	} else {
		__antithesis_instrumentation__.Notify(56775)
	}
	__antithesis_instrumentation__.Notify(56769)
	if z.RangeMinBytes == nil {
		__antithesis_instrumentation__.Notify(56776)
		unsetFields = append(unsetFields, "RangeMinBytes")
	} else {
		__antithesis_instrumentation__.Notify(56777)
	}
	__antithesis_instrumentation__.Notify(56770)
	if z.GC == nil {
		__antithesis_instrumentation__.Notify(56778)
		unsetFields = append(unsetFields, "GCPolicy")
	} else {
		__antithesis_instrumentation__.Notify(56779)
	}
	__antithesis_instrumentation__.Notify(56771)
	if z.NumReplicas == nil {
		__antithesis_instrumentation__.Notify(56780)
		unsetFields = append(unsetFields, "NumReplicas")
	} else {
		__antithesis_instrumentation__.Notify(56781)
	}
	__antithesis_instrumentation__.Notify(56772)

	if len(unsetFields) > 0 {
		__antithesis_instrumentation__.Notify(56782)
		return errors.AssertionFailedf("expected hydrated zone config: %s unset", strings.Join(unsetFields, ", "))
	} else {
		__antithesis_instrumentation__.Notify(56783)
	}
	__antithesis_instrumentation__.Notify(56773)
	return nil
}

func (z ZoneConfig) AsSpanConfig() roachpb.SpanConfig {
	__antithesis_instrumentation__.Notify(56784)
	spanConfig, err := z.toSpanConfig()
	if err != nil {
		__antithesis_instrumentation__.Notify(56786)
		log.Fatalf(context.Background(), "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(56787)
	}
	__antithesis_instrumentation__.Notify(56785)
	return spanConfig
}

func (z *ZoneConfig) toSpanConfig() (roachpb.SpanConfig, error) {
	__antithesis_instrumentation__.Notify(56788)
	var sc roachpb.SpanConfig
	var err error

	if err = z.EnsureFullyHydrated(); err != nil {
		__antithesis_instrumentation__.Notify(56797)
		return sc, err
	} else {
		__antithesis_instrumentation__.Notify(56798)
	}
	__antithesis_instrumentation__.Notify(56789)

	sc.RangeMinBytes = *z.RangeMinBytes
	sc.RangeMaxBytes = *z.RangeMaxBytes
	sc.GCPolicy.TTLSeconds = z.GC.TTLSeconds

	if z.GlobalReads != nil {
		__antithesis_instrumentation__.Notify(56799)
		sc.GlobalReads = *z.GlobalReads
	} else {
		__antithesis_instrumentation__.Notify(56800)
	}
	__antithesis_instrumentation__.Notify(56790)
	sc.NumReplicas = *z.NumReplicas
	if z.NumVoters != nil {
		__antithesis_instrumentation__.Notify(56801)
		sc.NumVoters = *z.NumVoters
	} else {
		__antithesis_instrumentation__.Notify(56802)
	}
	__antithesis_instrumentation__.Notify(56791)

	toSpanConfigConstraints := func(src []Constraint) ([]roachpb.Constraint, error) {
		__antithesis_instrumentation__.Notify(56803)
		spanConfigConstraints := make([]roachpb.Constraint, len(src))
		for i, c := range src {
			__antithesis_instrumentation__.Notify(56805)
			switch c.Type {
			case Constraint_REQUIRED:
				__antithesis_instrumentation__.Notify(56807)
				spanConfigConstraints[i].Type = roachpb.Constraint_REQUIRED
			case Constraint_PROHIBITED:
				__antithesis_instrumentation__.Notify(56808)
				spanConfigConstraints[i].Type = roachpb.Constraint_PROHIBITED
			default:
				__antithesis_instrumentation__.Notify(56809)
				return nil, errors.AssertionFailedf("unknown constraint type: %v", c.Type)
			}
			__antithesis_instrumentation__.Notify(56806)
			spanConfigConstraints[i].Key = c.Key
			spanConfigConstraints[i].Value = c.Value
		}
		__antithesis_instrumentation__.Notify(56804)
		return spanConfigConstraints, nil
	}
	__antithesis_instrumentation__.Notify(56792)

	toSpanConfigConstraintsConjunction := func(src []ConstraintsConjunction) ([]roachpb.ConstraintsConjunction, error) {
		__antithesis_instrumentation__.Notify(56810)
		constraintsConjunction := make([]roachpb.ConstraintsConjunction, len(src))
		for i, constraint := range src {
			__antithesis_instrumentation__.Notify(56812)
			constraintsConjunction[i].NumReplicas = constraint.NumReplicas
			constraintsConjunction[i].Constraints, err = toSpanConfigConstraints(constraint.Constraints)
			if err != nil {
				__antithesis_instrumentation__.Notify(56813)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(56814)
			}
		}
		__antithesis_instrumentation__.Notify(56811)
		return constraintsConjunction, nil
	}
	__antithesis_instrumentation__.Notify(56793)

	if len(z.Constraints) != 0 {
		__antithesis_instrumentation__.Notify(56815)
		sc.Constraints = make([]roachpb.ConstraintsConjunction, len(z.Constraints))
		sc.Constraints, err = toSpanConfigConstraintsConjunction(z.Constraints)
		if err != nil {
			__antithesis_instrumentation__.Notify(56816)
			return roachpb.SpanConfig{}, err
		} else {
			__antithesis_instrumentation__.Notify(56817)
		}
	} else {
		__antithesis_instrumentation__.Notify(56818)
	}
	__antithesis_instrumentation__.Notify(56794)
	if len(z.VoterConstraints) != 0 {
		__antithesis_instrumentation__.Notify(56819)
		sc.VoterConstraints, err = toSpanConfigConstraintsConjunction(z.VoterConstraints)
		if err != nil {
			__antithesis_instrumentation__.Notify(56820)
			return roachpb.SpanConfig{}, err
		} else {
			__antithesis_instrumentation__.Notify(56821)
		}
	} else {
		__antithesis_instrumentation__.Notify(56822)
	}
	__antithesis_instrumentation__.Notify(56795)

	if len(z.LeasePreferences) != 0 {
		__antithesis_instrumentation__.Notify(56823)
		sc.LeasePreferences = make([]roachpb.LeasePreference, len(z.LeasePreferences))
		for i, leasePreference := range z.LeasePreferences {
			__antithesis_instrumentation__.Notify(56824)
			sc.LeasePreferences[i].Constraints, err = toSpanConfigConstraints(leasePreference.Constraints)
			if err != nil {
				__antithesis_instrumentation__.Notify(56825)
				return roachpb.SpanConfig{}, err
			} else {
				__antithesis_instrumentation__.Notify(56826)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(56827)
	}
	__antithesis_instrumentation__.Notify(56796)
	return sc, nil
}

func init() {
	if len(NamedZonesList) != len(NamedZones) {
		panic(fmt.Errorf(
			"NamedZonesList (%d) and NamedZones (%d) should have the same number of entries",
			len(NamedZones),
			len(NamedZonesList),
		))
	}
}
