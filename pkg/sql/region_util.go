package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/dbdesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/multiregion"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/errors"
	"github.com/gogo/protobuf/proto"
)

type LiveClusterRegions map[catpb.RegionName]struct{}

func (s *LiveClusterRegions) IsActive(region catpb.RegionName) bool {
	__antithesis_instrumentation__.Notify(565024)
	_, ok := (*s)[region]
	return ok
}

func (s *LiveClusterRegions) toStrings() []string {
	__antithesis_instrumentation__.Notify(565025)
	ret := make([]string, 0, len(*s))
	for region := range *s {
		__antithesis_instrumentation__.Notify(565028)
		ret = append(ret, string(region))
	}
	__antithesis_instrumentation__.Notify(565026)
	sort.Slice(ret, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(565029)
		return ret[i] < ret[j]
	})
	__antithesis_instrumentation__.Notify(565027)
	return ret
}

func (p *planner) getLiveClusterRegions(ctx context.Context) (LiveClusterRegions, error) {
	__antithesis_instrumentation__.Notify(565030)
	return GetLiveClusterRegions(ctx, p)
}

func GetLiveClusterRegions(ctx context.Context, p PlanHookState) (LiveClusterRegions, error) {
	__antithesis_instrumentation__.Notify(565031)

	override := sessiondata.InternalExecutorOverride{
		User: security.RootUserName(),
	}

	it, err := p.ExtendedEvalContext().ExecCfg.InternalExecutor.QueryIteratorEx(
		ctx,
		"get_live_cluster_regions",
		p.ExtendedEvalContext().Txn,
		override,
		"SELECT region FROM [SHOW REGIONS FROM CLUSTER]",
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(565035)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(565036)
	}
	__antithesis_instrumentation__.Notify(565032)

	var ret LiveClusterRegions = make(map[catpb.RegionName]struct{})
	var ok bool
	for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(565037)
		row := it.Cur()
		ret[catpb.RegionName(*row[0].(*tree.DString))] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(565033)
	if err != nil {
		__antithesis_instrumentation__.Notify(565038)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(565039)
	}
	__antithesis_instrumentation__.Notify(565034)
	return ret, nil
}

func CheckClusterRegionIsLive(
	liveClusterRegions LiveClusterRegions, region catpb.RegionName,
) error {
	__antithesis_instrumentation__.Notify(565040)
	if !liveClusterRegions.IsActive(region) {
		__antithesis_instrumentation__.Notify(565042)
		return errors.WithHintf(
			pgerror.Newf(
				pgcode.InvalidName,
				"region %q does not exist",
				region,
			),
			"valid regions: %s",
			strings.Join(liveClusterRegions.toStrings(), ", "),
		)
	} else {
		__antithesis_instrumentation__.Notify(565043)
	}
	__antithesis_instrumentation__.Notify(565041)
	return nil
}

func makeRequiredConstraintForRegion(r catpb.RegionName) zonepb.Constraint {
	__antithesis_instrumentation__.Notify(565044)
	return zonepb.Constraint{
		Type:  zonepb.Constraint_REQUIRED,
		Key:   "region",
		Value: string(r),
	}
}

func zoneConfigForMultiRegionDatabase(
	regionConfig multiregion.RegionConfig,
) (zonepb.ZoneConfig, error) {
	__antithesis_instrumentation__.Notify(565045)
	numVoters, numReplicas := getNumVotersAndNumReplicas(regionConfig)

	constraints, err := synthesizeReplicaConstraints(regionConfig.Regions(), regionConfig.Placement())
	if err != nil {
		__antithesis_instrumentation__.Notify(565048)
		return zonepb.ZoneConfig{}, err
	} else {
		__antithesis_instrumentation__.Notify(565049)
	}
	__antithesis_instrumentation__.Notify(565046)

	voterConstraints, err := synthesizeVoterConstraints(regionConfig.PrimaryRegion(), regionConfig)
	if err != nil {
		__antithesis_instrumentation__.Notify(565050)
		return zonepb.ZoneConfig{}, err
	} else {
		__antithesis_instrumentation__.Notify(565051)
	}
	__antithesis_instrumentation__.Notify(565047)

	leasePreferences := synthesizeLeasePreferences(regionConfig.PrimaryRegion())

	zc := zonepb.ZoneConfig{
		NumReplicas:                 &numReplicas,
		NumVoters:                   &numVoters,
		Constraints:                 constraints,
		VoterConstraints:            voterConstraints,
		LeasePreferences:            leasePreferences,
		InheritedConstraints:        false,
		NullVoterConstraintsIsEmpty: true,
		InheritedLeasePreferences:   false,
	}

	zc = regionConfig.ApplyZoneConfigExtensionForRegionalIn(zc, regionConfig.PrimaryRegion())
	return zc, nil
}

func addConstraintsForSuperRegion(
	zc *zonepb.ZoneConfig, regionConfig multiregion.RegionConfig, affinityRegion catpb.RegionName,
) error {
	__antithesis_instrumentation__.Notify(565052)
	regions, err := regionConfig.GetSuperRegionRegionsForRegion(affinityRegion)
	if err != nil {
		__antithesis_instrumentation__.Notify(565054)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565055)
	}
	__antithesis_instrumentation__.Notify(565053)
	_, numReplicas := getNumVotersAndNumReplicas(regionConfig.WithRegions(regions))

	zc.NumReplicas = &numReplicas
	zc.Constraints = nil
	zc.InheritedConstraints = false

	switch regionConfig.SurvivalGoal() {
	case descpb.SurvivalGoal_ZONE_FAILURE:
		__antithesis_instrumentation__.Notify(565056)
		for _, region := range regions {
			__antithesis_instrumentation__.Notify(565061)
			zc.Constraints = append(zc.Constraints, zonepb.ConstraintsConjunction{
				NumReplicas: 1,
				Constraints: []zonepb.Constraint{makeRequiredConstraintForRegion(region)},
			})
		}
		__antithesis_instrumentation__.Notify(565057)
		return nil
	case descpb.SurvivalGoal_REGION_FAILURE:
		__antithesis_instrumentation__.Notify(565058)

		extraReplicaToConstrain := len(regions) == 3
		for _, region := range regions {
			__antithesis_instrumentation__.Notify(565062)
			n := int32(1)
			if region != affinityRegion && func() bool {
				__antithesis_instrumentation__.Notify(565064)
				return extraReplicaToConstrain == true
			}() == true {
				__antithesis_instrumentation__.Notify(565065)
				n = 2
				extraReplicaToConstrain = false
			} else {
				__antithesis_instrumentation__.Notify(565066)
			}
			__antithesis_instrumentation__.Notify(565063)
			zc.Constraints = append(zc.Constraints, zonepb.ConstraintsConjunction{
				NumReplicas: n,
				Constraints: []zonepb.Constraint{makeRequiredConstraintForRegion(region)},
			})
		}
		__antithesis_instrumentation__.Notify(565059)
		return nil
	default:
		__antithesis_instrumentation__.Notify(565060)
		return errors.AssertionFailedf("unknown survival goal: %v", regionConfig.SurvivalGoal())
	}
}

func zoneConfigForMultiRegionPartition(
	partitionRegion catpb.RegionName, regionConfig multiregion.RegionConfig,
) (zonepb.ZoneConfig, error) {
	__antithesis_instrumentation__.Notify(565067)
	zc := *zonepb.NewZoneConfig()

	numVoters, numReplicas := getNumVotersAndNumReplicas(regionConfig)
	zc.NumVoters = &numVoters

	if regionConfig.IsMemberOfExplicitSuperRegion(partitionRegion) {
		__antithesis_instrumentation__.Notify(565070)
		err := addConstraintsForSuperRegion(&zc, regionConfig, partitionRegion)
		if err != nil {
			__antithesis_instrumentation__.Notify(565071)
			return zonepb.ZoneConfig{}, err
		} else {
			__antithesis_instrumentation__.Notify(565072)
		}
	} else {
		__antithesis_instrumentation__.Notify(565073)
		if !regionConfig.RegionalInTablesInheritDatabaseConstraints(partitionRegion) {
			__antithesis_instrumentation__.Notify(565074)

			zc.NumReplicas = &numReplicas

			constraints, err := synthesizeReplicaConstraints(regionConfig.Regions(), regionConfig.Placement())
			if err != nil {
				__antithesis_instrumentation__.Notify(565076)
				return zonepb.ZoneConfig{}, err
			} else {
				__antithesis_instrumentation__.Notify(565077)
			}
			__antithesis_instrumentation__.Notify(565075)
			zc.Constraints = constraints
			zc.InheritedConstraints = false
		} else {
			__antithesis_instrumentation__.Notify(565078)
		}
	}
	__antithesis_instrumentation__.Notify(565068)

	voterConstraints, err := synthesizeVoterConstraints(partitionRegion, regionConfig)
	if err != nil {
		__antithesis_instrumentation__.Notify(565079)
		return zonepb.ZoneConfig{}, err
	} else {
		__antithesis_instrumentation__.Notify(565080)
	}
	__antithesis_instrumentation__.Notify(565069)
	zc.VoterConstraints = voterConstraints
	zc.NullVoterConstraintsIsEmpty = true

	leasePreferences := synthesizeLeasePreferences(partitionRegion)
	zc.LeasePreferences = leasePreferences
	zc.InheritedLeasePreferences = false

	zc = regionConfig.ApplyZoneConfigExtensionForRegionalIn(zc, partitionRegion)
	return zc, err
}

func maxFailuresBeforeUnavailability(numVoters int32) int32 {
	__antithesis_instrumentation__.Notify(565081)
	return ((numVoters + 1) / 2) - 1
}

func getNumVotersAndNumReplicas(
	regionConfig multiregion.RegionConfig,
) (numVoters, numReplicas int32) {
	__antithesis_instrumentation__.Notify(565082)
	const numVotersForZoneSurvival = 3

	const numVotersForRegionSurvival = 5

	numRegions := int32(len(regionConfig.Regions()))
	switch regionConfig.SurvivalGoal() {

	case descpb.SurvivalGoal_ZONE_FAILURE:
		__antithesis_instrumentation__.Notify(565084)
		numVoters = numVotersForZoneSurvival
		switch regionConfig.Placement() {
		case descpb.DataPlacement_DEFAULT:
			__antithesis_instrumentation__.Notify(565087)

			numReplicas = (numVotersForZoneSurvival) + (numRegions - 1)
		case descpb.DataPlacement_RESTRICTED:
			__antithesis_instrumentation__.Notify(565088)
			numReplicas = numVoters
		default:
			__antithesis_instrumentation__.Notify(565089)
			panic(errors.AssertionFailedf("unknown data placement: %v", regionConfig.Placement()))
		}
	case descpb.SurvivalGoal_REGION_FAILURE:
		__antithesis_instrumentation__.Notify(565085)

		numVoters = numVotersForRegionSurvival

		numReplicas = maxFailuresBeforeUnavailability(numVotersForRegionSurvival) + (numRegions - 1)
		if numReplicas < numVoters {
			__antithesis_instrumentation__.Notify(565090)

			numReplicas = numVoters
		} else {
			__antithesis_instrumentation__.Notify(565091)
		}
	default:
		__antithesis_instrumentation__.Notify(565086)
	}
	__antithesis_instrumentation__.Notify(565083)
	return numVoters, numReplicas
}

func synthesizeVoterConstraints(
	region catpb.RegionName, regionConfig multiregion.RegionConfig,
) ([]zonepb.ConstraintsConjunction, error) {
	__antithesis_instrumentation__.Notify(565092)
	switch regionConfig.SurvivalGoal() {
	case descpb.SurvivalGoal_ZONE_FAILURE:
		__antithesis_instrumentation__.Notify(565093)
		return []zonepb.ConstraintsConjunction{
			{

				Constraints: []zonepb.Constraint{makeRequiredConstraintForRegion(region)},
			},
		}, nil
	case descpb.SurvivalGoal_REGION_FAILURE:
		__antithesis_instrumentation__.Notify(565094)
		numVoters, _ := getNumVotersAndNumReplicas(regionConfig)
		return []zonepb.ConstraintsConjunction{
			{

				NumReplicas: maxFailuresBeforeUnavailability(numVoters),
				Constraints: []zonepb.Constraint{makeRequiredConstraintForRegion(region)},
			},
		}, nil
	default:
		__antithesis_instrumentation__.Notify(565095)
		return nil, errors.AssertionFailedf("unknown survival goal: %v", regionConfig.SurvivalGoal())
	}
}

func synthesizeReplicaConstraints(
	regions catpb.RegionNames, placement descpb.DataPlacement,
) ([]zonepb.ConstraintsConjunction, error) {
	__antithesis_instrumentation__.Notify(565096)
	switch placement {
	case descpb.DataPlacement_DEFAULT:
		__antithesis_instrumentation__.Notify(565097)
		constraints := make([]zonepb.ConstraintsConjunction, len(regions))
		for i, region := range regions {
			__antithesis_instrumentation__.Notify(565101)

			constraints[i] = zonepb.ConstraintsConjunction{
				NumReplicas: 1,
				Constraints: []zonepb.Constraint{makeRequiredConstraintForRegion(region)},
			}
		}
		__antithesis_instrumentation__.Notify(565098)
		return constraints, nil
	case descpb.DataPlacement_RESTRICTED:
		__antithesis_instrumentation__.Notify(565099)

		return nil, nil
	default:
		__antithesis_instrumentation__.Notify(565100)
		return nil, errors.AssertionFailedf("unknown data placement: %v", placement)
	}
}

func synthesizeLeasePreferences(region catpb.RegionName) []zonepb.LeasePreference {
	__antithesis_instrumentation__.Notify(565102)
	return []zonepb.LeasePreference{
		{Constraints: []zonepb.Constraint{makeRequiredConstraintForRegion(region)}},
	}
}

func zoneConfigForMultiRegionTable(
	localityConfig catpb.LocalityConfig, regionConfig multiregion.RegionConfig,
) (zonepb.ZoneConfig, error) {
	__antithesis_instrumentation__.Notify(565103)
	zc := *zonepb.NewZoneConfig()

	switch l := localityConfig.Locality.(type) {
	case *catpb.LocalityConfig_Global_:
		__antithesis_instrumentation__.Notify(565104)

		zc.GlobalReads = proto.Bool(true)

		if !regionConfig.GlobalTablesInheritDatabaseConstraints() {
			__antithesis_instrumentation__.Notify(565113)

			regionConfig = regionConfig.WithPlacementDefault()

			numVoters, numReplicas := getNumVotersAndNumReplicas(regionConfig)
			zc.NumVoters = &numVoters
			zc.NumReplicas = &numReplicas

			constraints, err := synthesizeReplicaConstraints(regionConfig.Regions(), regionConfig.Placement())
			if err != nil {
				__antithesis_instrumentation__.Notify(565116)
				return zonepb.ZoneConfig{}, err
			} else {
				__antithesis_instrumentation__.Notify(565117)
			}
			__antithesis_instrumentation__.Notify(565114)
			zc.Constraints = constraints
			zc.InheritedConstraints = false

			voterConstraints, err := synthesizeVoterConstraints(regionConfig.PrimaryRegion(), regionConfig)
			if err != nil {
				__antithesis_instrumentation__.Notify(565118)
				return zonepb.ZoneConfig{}, err
			} else {
				__antithesis_instrumentation__.Notify(565119)
			}
			__antithesis_instrumentation__.Notify(565115)
			zc.VoterConstraints = voterConstraints
			zc.NullVoterConstraintsIsEmpty = true

			leasePreferences := synthesizeLeasePreferences(regionConfig.PrimaryRegion())
			zc.LeasePreferences = leasePreferences
			zc.InheritedLeasePreferences = false

			zc = regionConfig.ApplyZoneConfigExtensionForGlobal(zc)
		} else {
			__antithesis_instrumentation__.Notify(565120)
		}
		__antithesis_instrumentation__.Notify(565105)

		return zc, nil
	case *catpb.LocalityConfig_RegionalByTable_:
		__antithesis_instrumentation__.Notify(565106)
		affinityRegion := regionConfig.PrimaryRegion()
		if l.RegionalByTable.Region != nil {
			__antithesis_instrumentation__.Notify(565121)
			affinityRegion = *l.RegionalByTable.Region
		} else {
			__antithesis_instrumentation__.Notify(565122)
		}
		__antithesis_instrumentation__.Notify(565107)
		if l.RegionalByTable.Region == nil && func() bool {
			__antithesis_instrumentation__.Notify(565123)
			return !regionConfig.IsMemberOfExplicitSuperRegion(affinityRegion) == true
		}() == true {
			__antithesis_instrumentation__.Notify(565124)

			return zc, nil
		} else {
			__antithesis_instrumentation__.Notify(565125)
		}
		__antithesis_instrumentation__.Notify(565108)

		numVoters, numReplicas := getNumVotersAndNumReplicas(regionConfig)
		zc.NumVoters = &numVoters

		if regionConfig.IsMemberOfExplicitSuperRegion(affinityRegion) {
			__antithesis_instrumentation__.Notify(565126)
			err := addConstraintsForSuperRegion(&zc, regionConfig, affinityRegion)
			if err != nil {
				__antithesis_instrumentation__.Notify(565127)
				return zonepb.ZoneConfig{}, err
			} else {
				__antithesis_instrumentation__.Notify(565128)
			}
		} else {
			__antithesis_instrumentation__.Notify(565129)
			if !regionConfig.RegionalInTablesInheritDatabaseConstraints(affinityRegion) {
				__antithesis_instrumentation__.Notify(565130)

				zc.NumReplicas = &numReplicas

				constraints, err := synthesizeReplicaConstraints(regionConfig.Regions(), regionConfig.Placement())
				if err != nil {
					__antithesis_instrumentation__.Notify(565132)
					return zonepb.ZoneConfig{}, err
				} else {
					__antithesis_instrumentation__.Notify(565133)
				}
				__antithesis_instrumentation__.Notify(565131)
				zc.Constraints = constraints
				zc.InheritedConstraints = false
			} else {
				__antithesis_instrumentation__.Notify(565134)
			}
		}
		__antithesis_instrumentation__.Notify(565109)

		voterConstraints, err := synthesizeVoterConstraints(affinityRegion, regionConfig)
		if err != nil {
			__antithesis_instrumentation__.Notify(565135)
			return zonepb.ZoneConfig{}, err
		} else {
			__antithesis_instrumentation__.Notify(565136)
		}
		__antithesis_instrumentation__.Notify(565110)
		zc.VoterConstraints = voterConstraints
		zc.NullVoterConstraintsIsEmpty = true

		leasePreferences := synthesizeLeasePreferences(affinityRegion)
		zc.LeasePreferences = leasePreferences
		zc.InheritedLeasePreferences = false

		zc = regionConfig.ApplyZoneConfigExtensionForRegionalIn(zc, affinityRegion)
		return zc, nil
	case *catpb.LocalityConfig_RegionalByRow_:
		__antithesis_instrumentation__.Notify(565111)

		return zc, nil
	default:
		__antithesis_instrumentation__.Notify(565112)
		return zonepb.ZoneConfig{}, errors.AssertionFailedf(
			"unexpected unknown locality type %T", localityConfig.Locality)
	}
}

type applyZoneConfigForMultiRegionTableOption func(
	zoneConfig zonepb.ZoneConfig,
	regionConfig multiregion.RegionConfig,
	table catalog.TableDescriptor,
) (hasNewSubzones bool, newZoneConfig zonepb.ZoneConfig, err error)

func applyZoneConfigForMultiRegionTableOptionNewIndexes(
	indexIDs ...descpb.IndexID,
) applyZoneConfigForMultiRegionTableOption {
	__antithesis_instrumentation__.Notify(565137)
	return func(
		zoneConfig zonepb.ZoneConfig,
		regionConfig multiregion.RegionConfig,
		table catalog.TableDescriptor,
	) (hasNewSubzones bool, newZoneConfig zonepb.ZoneConfig, err error) {
		__antithesis_instrumentation__.Notify(565138)
		for _, indexID := range indexIDs {
			__antithesis_instrumentation__.Notify(565140)
			for _, region := range regionConfig.Regions() {
				__antithesis_instrumentation__.Notify(565141)
				zc, err := zoneConfigForMultiRegionPartition(region, regionConfig)
				if err != nil {
					__antithesis_instrumentation__.Notify(565143)
					return false, zoneConfig, err
				} else {
					__antithesis_instrumentation__.Notify(565144)
				}
				__antithesis_instrumentation__.Notify(565142)
				zoneConfig.SetSubzone(zonepb.Subzone{
					IndexID:       uint32(indexID),
					PartitionName: string(region),
					Config:        zc,
				})
			}
		}
		__antithesis_instrumentation__.Notify(565139)
		return true, zoneConfig, nil
	}
}

func dropZoneConfigsForMultiRegionIndexes(
	indexIDs ...descpb.IndexID,
) applyZoneConfigForMultiRegionTableOption {
	__antithesis_instrumentation__.Notify(565145)
	return func(
		zoneConfig zonepb.ZoneConfig,
		regionConfig multiregion.RegionConfig,
		table catalog.TableDescriptor,
	) (hasNewSubzones bool, newZoneConfig zonepb.ZoneConfig, err error) {
		__antithesis_instrumentation__.Notify(565146)

		zoneConfig.ClearFieldsOfAllSubzones(zonepb.MultiRegionZoneConfigFields)

		if len(zoneConfig.Subzones) == 0 && func() bool {
			__antithesis_instrumentation__.Notify(565148)
			return zoneConfig.IsSubzonePlaceholder() == true
		}() == true {
			__antithesis_instrumentation__.Notify(565149)
			zoneConfig.NumReplicas = nil
			zoneConfig.SubzoneSpans = nil
		} else {
			__antithesis_instrumentation__.Notify(565150)
		}
		__antithesis_instrumentation__.Notify(565147)
		return false, zoneConfig, nil
	}
}

func isPlaceholderZoneConfigForMultiRegion(zc zonepb.ZoneConfig) bool {
	__antithesis_instrumentation__.Notify(565151)

	if len(zc.Subzones) == 0 {
		__antithesis_instrumentation__.Notify(565153)
		return false
	} else {
		__antithesis_instrumentation__.Notify(565154)
	}
	__antithesis_instrumentation__.Notify(565152)

	strippedZC := zc
	strippedZC.Subzones, strippedZC.SubzoneSpans = nil, nil
	return strippedZC.Equal(zonepb.NewZoneConfig())
}

func applyZoneConfigForMultiRegionTableOptionTableNewConfig(
	newConfig catpb.LocalityConfig,
) applyZoneConfigForMultiRegionTableOption {
	__antithesis_instrumentation__.Notify(565155)
	return func(
		zc zonepb.ZoneConfig,
		regionConfig multiregion.RegionConfig,
		table catalog.TableDescriptor,
	) (bool, zonepb.ZoneConfig, error) {
		__antithesis_instrumentation__.Notify(565156)
		localityZoneConfig, err := zoneConfigForMultiRegionTable(
			newConfig,
			regionConfig,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(565158)
			return false, zonepb.ZoneConfig{}, err
		} else {
			__antithesis_instrumentation__.Notify(565159)
		}
		__antithesis_instrumentation__.Notify(565157)
		zc.CopyFromZone(localityZoneConfig, zonepb.MultiRegionZoneConfigFields)
		return false, zc, nil
	}
}

var ApplyZoneConfigForMultiRegionTableOptionTableAndIndexes = func(
	zc zonepb.ZoneConfig,
	regionConfig multiregion.RegionConfig,
	table catalog.TableDescriptor,
) (bool, zonepb.ZoneConfig, error) {
	__antithesis_instrumentation__.Notify(565160)
	localityConfig := *table.GetLocalityConfig()
	localityZoneConfig, err := zoneConfigForMultiRegionTable(
		localityConfig,
		regionConfig,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(565163)
		return false, zonepb.ZoneConfig{}, err
	} else {
		__antithesis_instrumentation__.Notify(565164)
	}
	__antithesis_instrumentation__.Notify(565161)

	zc.ClearFieldsOfAllSubzones(zonepb.MultiRegionZoneConfigFields)

	zc.CopyFromZone(localityZoneConfig, zonepb.MultiRegionZoneConfigFields)

	hasNewSubzones := table.IsLocalityRegionalByRow()
	if hasNewSubzones {
		__antithesis_instrumentation__.Notify(565165)
		for _, region := range regionConfig.Regions() {
			__antithesis_instrumentation__.Notify(565166)
			subzoneConfig, err := zoneConfigForMultiRegionPartition(region, regionConfig)
			if err != nil {
				__antithesis_instrumentation__.Notify(565168)
				return false, zc, err
			} else {
				__antithesis_instrumentation__.Notify(565169)
			}
			__antithesis_instrumentation__.Notify(565167)
			for _, idx := range table.NonDropIndexes() {
				__antithesis_instrumentation__.Notify(565170)
				zc.SetSubzone(zonepb.Subzone{
					IndexID:       uint32(idx.GetID()),
					PartitionName: string(region),
					Config:        subzoneConfig,
				})
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(565171)
	}
	__antithesis_instrumentation__.Notify(565162)
	return hasNewSubzones, zc, nil
}

var applyZoneConfigForMultiRegionTableOptionRemoveGlobalZoneConfig = func(
	zc zonepb.ZoneConfig,
	regionConfig multiregion.RegionConfig,
	table catalog.TableDescriptor,
) (bool, zonepb.ZoneConfig, error) {
	__antithesis_instrumentation__.Notify(565172)
	zc.CopyFromZone(*zonepb.NewZoneConfig(), zonepb.MultiRegionZoneConfigFields)
	return false, zc, nil
}

func prepareZoneConfigForMultiRegionTable(
	ctx context.Context,
	txn *kv.Txn,
	execCfg *ExecutorConfig,
	regionConfig multiregion.RegionConfig,
	table catalog.TableDescriptor,
	opts ...applyZoneConfigForMultiRegionTableOption,
) (*zoneConfigUpdate, error) {
	__antithesis_instrumentation__.Notify(565173)
	tableID := table.GetID()
	currentZoneConfig, err := getZoneConfigRaw(ctx, txn, execCfg.Codec, execCfg.Settings, tableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(565182)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(565183)
	}
	__antithesis_instrumentation__.Notify(565174)
	newZoneConfig := *zonepb.NewZoneConfig()
	if currentZoneConfig != nil {
		__antithesis_instrumentation__.Notify(565184)
		newZoneConfig = *currentZoneConfig
	} else {
		__antithesis_instrumentation__.Notify(565185)
	}
	__antithesis_instrumentation__.Notify(565175)

	var hasNewSubzones bool
	for _, opt := range opts {
		__antithesis_instrumentation__.Notify(565186)
		newHasNewSubzones, modifiedNewZoneConfig, err := opt(
			newZoneConfig,
			regionConfig,
			table,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(565188)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(565189)
		}
		__antithesis_instrumentation__.Notify(565187)
		hasNewSubzones = newHasNewSubzones || func() bool {
			__antithesis_instrumentation__.Notify(565190)
			return hasNewSubzones == true
		}() == true
		newZoneConfig = modifiedNewZoneConfig
	}
	__antithesis_instrumentation__.Notify(565176)

	if isPlaceholderZoneConfigForMultiRegion(newZoneConfig) {
		__antithesis_instrumentation__.Notify(565191)
		newZoneConfig.NumReplicas = proto.Int32(0)
	} else {
		__antithesis_instrumentation__.Notify(565192)
	}
	__antithesis_instrumentation__.Notify(565177)

	newZoneConfigIsEmpty := newZoneConfig.Equal(zonepb.NewZoneConfig())
	currentZoneConfigIsEmpty := currentZoneConfig.Equal(zonepb.NewZoneConfig())
	rewriteZoneConfig := !newZoneConfigIsEmpty
	deleteZoneConfig := newZoneConfigIsEmpty && func() bool {
		__antithesis_instrumentation__.Notify(565193)
		return !currentZoneConfigIsEmpty == true
	}() == true

	if deleteZoneConfig {
		__antithesis_instrumentation__.Notify(565194)
		return &zoneConfigUpdate{id: tableID}, nil
	} else {
		__antithesis_instrumentation__.Notify(565195)
	}
	__antithesis_instrumentation__.Notify(565178)
	if !rewriteZoneConfig {
		__antithesis_instrumentation__.Notify(565196)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(565197)
	}
	__antithesis_instrumentation__.Notify(565179)

	if err := newZoneConfig.Validate(); err != nil {
		__antithesis_instrumentation__.Notify(565198)
		return nil, pgerror.Wrap(
			err,
			pgcode.CheckViolation,
			"could not validate zone config",
		)
	} else {
		__antithesis_instrumentation__.Notify(565199)
	}
	__antithesis_instrumentation__.Notify(565180)
	if err := newZoneConfig.ValidateTandemFields(); err != nil {
		__antithesis_instrumentation__.Notify(565200)
		return nil, pgerror.Wrap(
			err,
			pgcode.CheckViolation,
			"could not validate zone config",
		)
	} else {
		__antithesis_instrumentation__.Notify(565201)
	}
	__antithesis_instrumentation__.Notify(565181)
	return prepareZoneConfigWrites(
		ctx, execCfg, tableID, table, &newZoneConfig, hasNewSubzones,
	)
}

func ApplyZoneConfigForMultiRegionTable(
	ctx context.Context,
	txn *kv.Txn,
	execCfg *ExecutorConfig,
	regionConfig multiregion.RegionConfig,
	table catalog.TableDescriptor,
	opts ...applyZoneConfigForMultiRegionTableOption,
) error {
	__antithesis_instrumentation__.Notify(565202)
	update, err := prepareZoneConfigForMultiRegionTable(ctx, txn, execCfg, regionConfig, table, opts...)
	if update == nil || func() bool {
		__antithesis_instrumentation__.Notify(565204)
		return err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(565205)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565206)
	}
	__antithesis_instrumentation__.Notify(565203)
	_, err = writeZoneConfigUpdate(ctx, txn, execCfg, update)
	return err
}

func ApplyZoneConfigFromDatabaseRegionConfig(
	ctx context.Context,
	dbID descpb.ID,
	regionConfig multiregion.RegionConfig,
	txn *kv.Txn,
	execConfig *ExecutorConfig,
) error {
	__antithesis_instrumentation__.Notify(565207)

	dbZoneConfig, err := zoneConfigForMultiRegionDatabase(regionConfig)
	if err != nil {
		__antithesis_instrumentation__.Notify(565209)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565210)
	}
	__antithesis_instrumentation__.Notify(565208)
	return applyZoneConfigForMultiRegionDatabase(
		ctx,
		dbID,
		dbZoneConfig,
		txn,
		execConfig,
	)
}

func discardMultiRegionFieldsForDatabaseZoneConfig(
	ctx context.Context, dbID descpb.ID, txn *kv.Txn, execConfig *ExecutorConfig,
) error {
	__antithesis_instrumentation__.Notify(565211)

	return applyZoneConfigForMultiRegionDatabase(
		ctx,
		dbID,
		*zonepb.NewZoneConfig(),
		txn,
		execConfig,
	)
}

func applyZoneConfigForMultiRegionDatabase(
	ctx context.Context,
	dbID descpb.ID,
	mergeZoneConfig zonepb.ZoneConfig,
	txn *kv.Txn,
	execConfig *ExecutorConfig,
) error {
	__antithesis_instrumentation__.Notify(565212)
	currentZoneConfig, err := getZoneConfigRaw(ctx, txn, execConfig.Codec, execConfig.Settings, dbID)
	if err != nil {
		__antithesis_instrumentation__.Notify(565217)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565218)
	}
	__antithesis_instrumentation__.Notify(565213)
	newZoneConfig := *zonepb.NewZoneConfig()
	if currentZoneConfig != nil {
		__antithesis_instrumentation__.Notify(565219)
		newZoneConfig = *currentZoneConfig
	} else {
		__antithesis_instrumentation__.Notify(565220)
	}
	__antithesis_instrumentation__.Notify(565214)
	newZoneConfig.CopyFromZone(
		mergeZoneConfig,
		zonepb.MultiRegionZoneConfigFields,
	)

	if newZoneConfig.Equal(zonepb.NewZoneConfig()) {
		__antithesis_instrumentation__.Notify(565221)
		_, err = execConfig.InternalExecutor.Exec(
			ctx,
			"delete-zone-multiregion-database",
			txn,
			"DELETE FROM system.zones WHERE id = $1",
			dbID,
		)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565222)
	}
	__antithesis_instrumentation__.Notify(565215)
	if _, err := writeZoneConfig(
		ctx,
		txn,
		dbID,
		nil,
		&newZoneConfig,
		execConfig,
		false,
	); err != nil {
		__antithesis_instrumentation__.Notify(565223)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565224)
	}
	__antithesis_instrumentation__.Notify(565216)
	return nil
}

type updateZoneConfigOptions struct {
	filterFunc func(tb *tabledesc.Mutable) bool
}

type updateZoneConfigOption func(options *updateZoneConfigOptions)

func (p *planner) updateZoneConfigsForTables(
	ctx context.Context, desc catalog.DatabaseDescriptor, updateOpts ...updateZoneConfigOption,
) error {
	__antithesis_instrumentation__.Notify(565225)
	opts := updateZoneConfigOptions{
		filterFunc: func(_ *tabledesc.Mutable) bool { __antithesis_instrumentation__.Notify(565229); return true },
	}
	__antithesis_instrumentation__.Notify(565226)
	for _, f := range updateOpts {
		__antithesis_instrumentation__.Notify(565230)
		f(&opts)
	}
	__antithesis_instrumentation__.Notify(565227)

	regionConfig, err := SynthesizeRegionConfig(ctx, p.txn, desc.GetID(), p.Descriptors())
	if err != nil {
		__antithesis_instrumentation__.Notify(565231)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565232)
	}
	__antithesis_instrumentation__.Notify(565228)

	return p.forEachMutableTableInDatabase(
		ctx,
		desc,
		func(ctx context.Context, scName string, tbDesc *tabledesc.Mutable) error {
			__antithesis_instrumentation__.Notify(565233)
			if opts.filterFunc(tbDesc) {
				__antithesis_instrumentation__.Notify(565235)
				return ApplyZoneConfigForMultiRegionTable(
					ctx,
					p.txn,
					p.ExecCfg(),
					regionConfig,
					tbDesc,
					ApplyZoneConfigForMultiRegionTableOptionTableAndIndexes,
				)
			} else {
				__antithesis_instrumentation__.Notify(565236)
			}
			__antithesis_instrumentation__.Notify(565234)
			return nil
		},
	)
}

func WithOnlyGlobalTables(opts *updateZoneConfigOptions) {
	__antithesis_instrumentation__.Notify(565237)
	opts.filterFunc = func(tb *tabledesc.Mutable) bool {
		__antithesis_instrumentation__.Notify(565238)
		return tb.IsLocalityGlobal()
	}
}

func WithOnlyRegionalTablesAndGlobalTables(opts *updateZoneConfigOptions) {
	__antithesis_instrumentation__.Notify(565239)
	opts.filterFunc = func(tb *tabledesc.Mutable) bool {
		__antithesis_instrumentation__.Notify(565240)
		return tb.IsLocalityGlobal() || func() bool {
			__antithesis_instrumentation__.Notify(565241)
			return tb.IsLocalityRegionalByTable() == true
		}() == true
	}
}

func (p *planner) maybeInitializeMultiRegionDatabase(
	ctx context.Context, desc *dbdesc.Mutable, regionConfig *multiregion.RegionConfig,
) error {
	__antithesis_instrumentation__.Notify(565242)

	if !desc.IsMultiRegion() {
		__antithesis_instrumentation__.Notify(565247)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(565248)
	}
	__antithesis_instrumentation__.Notify(565243)

	regionLabels := make(tree.EnumValueList, 0, len(regionConfig.Regions()))
	for _, regionName := range regionConfig.Regions() {
		__antithesis_instrumentation__.Notify(565249)
		regionLabels = append(regionLabels, tree.EnumValue(regionName))
	}
	__antithesis_instrumentation__.Notify(565244)

	if err := p.createEnumWithID(
		p.RunParams(ctx),
		regionConfig.RegionEnumID(),
		regionLabels,
		desc,
		tree.NewQualifiedTypeName(desc.Name, tree.PublicSchema, tree.RegionEnum),
		EnumTypeMultiRegion,
	); err != nil {
		__antithesis_instrumentation__.Notify(565250)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565251)
	}
	__antithesis_instrumentation__.Notify(565245)

	if err := ApplyZoneConfigFromDatabaseRegionConfig(
		ctx,
		desc.ID,
		*regionConfig,
		p.txn,
		p.execCfg); err != nil {
		__antithesis_instrumentation__.Notify(565252)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565253)
	}
	__antithesis_instrumentation__.Notify(565246)

	return nil
}

func partitionByForRegionalByRow(
	regionConfig multiregion.RegionConfig, col tree.Name,
) *tree.PartitionBy {
	__antithesis_instrumentation__.Notify(565254)
	listPartition := make([]tree.ListPartition, len(regionConfig.Regions()))
	for i, region := range regionConfig.Regions() {
		__antithesis_instrumentation__.Notify(565256)
		listPartition[i] = tree.ListPartition{
			Name:  tree.UnrestrictedName(region),
			Exprs: tree.Exprs{tree.NewStrVal(string(region))},
		}
	}
	__antithesis_instrumentation__.Notify(565255)

	return &tree.PartitionBy{
		Fields: tree.NameList{col},
		List:   listPartition,
	}
}

func (p *planner) ValidateAllMultiRegionZoneConfigsInCurrentDatabase(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(565257)
	dbDesc, err := p.Descriptors().GetImmutableDatabaseByName(
		p.EvalContext().Ctx(),
		p.txn,
		p.CurrentDatabase(),
		tree.DatabaseLookupFlags{
			Required: true,
		},
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(565261)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565262)
	}
	__antithesis_instrumentation__.Notify(565258)
	if !dbDesc.IsMultiRegion() {
		__antithesis_instrumentation__.Notify(565263)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(565264)
	}
	__antithesis_instrumentation__.Notify(565259)
	regionConfig, err := SynthesizeRegionConfig(
		ctx,
		p.txn,
		dbDesc.GetID(),
		p.Descriptors(),
		SynthesizeRegionConfigOptionForValidation,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(565265)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565266)
	}
	__antithesis_instrumentation__.Notify(565260)
	return p.validateAllMultiRegionZoneConfigsInDatabase(
		ctx,
		dbDesc,
		&zoneConfigForMultiRegionValidatorValidation{
			zoneConfigForMultiRegionValidatorExistingMultiRegionObject: zoneConfigForMultiRegionValidatorExistingMultiRegionObject{
				regionConfig: regionConfig,
			},
		},
	)
}

func (p *planner) ResetMultiRegionZoneConfigsForTable(ctx context.Context, id int64) error {
	__antithesis_instrumentation__.Notify(565267)
	desc, err := p.Descriptors().GetMutableTableVersionByID(ctx, descpb.ID(id), p.txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(565271)
		return errors.Wrapf(err, "error resolving referenced table ID %d", id)
	} else {
		__antithesis_instrumentation__.Notify(565272)
	}
	__antithesis_instrumentation__.Notify(565268)

	if desc.LocalityConfig == nil {
		__antithesis_instrumentation__.Notify(565273)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(565274)
	}
	__antithesis_instrumentation__.Notify(565269)
	regionConfig, err := SynthesizeRegionConfig(ctx, p.txn, desc.GetParentID(), p.Descriptors())
	if err != nil {
		__antithesis_instrumentation__.Notify(565275)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565276)
	}
	__antithesis_instrumentation__.Notify(565270)

	return ApplyZoneConfigForMultiRegionTable(
		ctx,
		p.txn,
		p.ExecCfg(),
		regionConfig,
		desc,
		ApplyZoneConfigForMultiRegionTableOptionTableAndIndexes,
	)
}

func (p *planner) ResetMultiRegionZoneConfigsForDatabase(ctx context.Context, id int64) error {
	__antithesis_instrumentation__.Notify(565277)
	_, dbDesc, err := p.Descriptors().GetImmutableDatabaseByID(
		p.EvalContext().Ctx(),
		p.txn,
		descpb.ID(id),
		tree.DatabaseLookupFlags{
			Required:    true,
			AvoidLeased: true,
		},
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(565282)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565283)
	}
	__antithesis_instrumentation__.Notify(565278)
	if !dbDesc.IsMultiRegion() {
		__antithesis_instrumentation__.Notify(565284)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(565285)
	}
	__antithesis_instrumentation__.Notify(565279)

	regionConfig, err := SynthesizeRegionConfig(ctx, p.txn, dbDesc.GetID(), p.Descriptors())
	if err != nil {
		__antithesis_instrumentation__.Notify(565286)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565287)
	}
	__antithesis_instrumentation__.Notify(565280)

	if err := ApplyZoneConfigFromDatabaseRegionConfig(
		ctx,
		dbDesc.GetID(),
		regionConfig,
		p.txn,
		p.execCfg,
	); err != nil {
		__antithesis_instrumentation__.Notify(565288)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565289)
	}
	__antithesis_instrumentation__.Notify(565281)
	return nil
}

func (p *planner) validateAllMultiRegionZoneConfigsInDatabase(
	ctx context.Context,
	dbDesc catalog.DatabaseDescriptor,
	zoneConfigForMultiRegionValidator zoneConfigForMultiRegionValidator,
) error {
	__antithesis_instrumentation__.Notify(565290)
	var ids []descpb.ID
	if err := p.forEachMutableTableInDatabase(
		ctx,
		dbDesc,
		func(ctx context.Context, scName string, tbDesc *tabledesc.Mutable) error {
			__antithesis_instrumentation__.Notify(565294)
			ids = append(ids, tbDesc.GetID())
			return nil
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(565295)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565296)
	}
	__antithesis_instrumentation__.Notify(565291)
	ids = append(ids, dbDesc.GetID())

	zoneConfigs, err := getZoneConfigRawBatch(
		ctx,
		p.txn,
		p.ExecCfg().Codec,
		p.ExecCfg().Settings,
		ids,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(565297)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565298)
	}
	__antithesis_instrumentation__.Notify(565292)

	if err := p.validateZoneConfigForMultiRegionDatabase(
		dbDesc,
		zoneConfigs[dbDesc.GetID()],
		zoneConfigForMultiRegionValidator,
	); err != nil {
		__antithesis_instrumentation__.Notify(565299)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565300)
	}
	__antithesis_instrumentation__.Notify(565293)

	return p.forEachMutableTableInDatabase(
		ctx,
		dbDesc,
		func(ctx context.Context, scName string, tbDesc *tabledesc.Mutable) error {
			__antithesis_instrumentation__.Notify(565301)
			return p.validateZoneConfigForMultiRegionTable(
				tbDesc,
				zoneConfigs[tbDesc.GetID()],
				zoneConfigForMultiRegionValidator,
			)
		},
	)
}

func (p *planner) CurrentDatabaseRegionConfig(
	ctx context.Context,
) (tree.DatabaseRegionConfig, error) {
	__antithesis_instrumentation__.Notify(565302)
	dbDesc, err := p.Descriptors().GetImmutableDatabaseByName(
		p.EvalContext().Ctx(),
		p.txn,
		p.CurrentDatabase(),
		tree.DatabaseLookupFlags{
			Required: true,
		},
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(565305)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(565306)
	}
	__antithesis_instrumentation__.Notify(565303)

	if !dbDesc.IsMultiRegion() {
		__antithesis_instrumentation__.Notify(565307)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(565308)
	}
	__antithesis_instrumentation__.Notify(565304)

	return SynthesizeRegionConfig(
		ctx,
		p.txn,
		dbDesc.GetID(),
		p.Descriptors(),
		SynthesizeRegionConfigOptionUseCache,
	)
}

type synthesizeRegionConfigOptions struct {
	includeOffline bool
	forValidation  bool
	useCache       bool
}

type SynthesizeRegionConfigOption func(o *synthesizeRegionConfigOptions)

var SynthesizeRegionConfigOptionIncludeOffline SynthesizeRegionConfigOption = func(o *synthesizeRegionConfigOptions) {
	__antithesis_instrumentation__.Notify(565309)
	o.includeOffline = true
}

var SynthesizeRegionConfigOptionForValidation SynthesizeRegionConfigOption = func(o *synthesizeRegionConfigOptions) {
	__antithesis_instrumentation__.Notify(565310)
	o.forValidation = true
}

var SynthesizeRegionConfigOptionUseCache SynthesizeRegionConfigOption = func(o *synthesizeRegionConfigOptions) {
	__antithesis_instrumentation__.Notify(565311)
	o.useCache = true
}

func SynthesizeRegionConfig(
	ctx context.Context,
	txn *kv.Txn,
	dbID descpb.ID,
	descsCol *descs.Collection,
	opts ...SynthesizeRegionConfigOption,
) (multiregion.RegionConfig, error) {
	__antithesis_instrumentation__.Notify(565312)
	var o synthesizeRegionConfigOptions
	for _, opt := range opts {
		__antithesis_instrumentation__.Notify(565323)
		opt(&o)
	}
	__antithesis_instrumentation__.Notify(565313)

	regionConfig := multiregion.RegionConfig{}
	_, dbDesc, err := descsCol.GetImmutableDatabaseByID(ctx, txn, dbID, tree.DatabaseLookupFlags{
		AvoidLeased:    !o.useCache,
		Required:       true,
		IncludeOffline: o.includeOffline,
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(565324)
		return regionConfig, err
	} else {
		__antithesis_instrumentation__.Notify(565325)
	}
	__antithesis_instrumentation__.Notify(565314)

	regionEnumID, err := dbDesc.MultiRegionEnumID()
	if err != nil {
		__antithesis_instrumentation__.Notify(565326)
		return regionConfig, err
	} else {
		__antithesis_instrumentation__.Notify(565327)
	}
	__antithesis_instrumentation__.Notify(565315)

	regionEnum, err := descsCol.GetImmutableTypeByID(
		ctx,
		txn,
		regionEnumID,
		tree.ObjectLookupFlags{
			CommonLookupFlags: tree.CommonLookupFlags{
				AvoidLeased:    !o.useCache,
				IncludeOffline: o.includeOffline,
			},
		},
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(565328)
		return multiregion.RegionConfig{}, err
	} else {
		__antithesis_instrumentation__.Notify(565329)
	}
	__antithesis_instrumentation__.Notify(565316)

	var regionNames catpb.RegionNames
	if o.forValidation {
		__antithesis_instrumentation__.Notify(565330)
		regionNames, err = regionEnum.RegionNamesForValidation()
	} else {
		__antithesis_instrumentation__.Notify(565331)
		regionNames, err = regionEnum.RegionNames()
	}
	__antithesis_instrumentation__.Notify(565317)
	if err != nil {
		__antithesis_instrumentation__.Notify(565332)
		return regionConfig, err
	} else {
		__antithesis_instrumentation__.Notify(565333)
	}
	__antithesis_instrumentation__.Notify(565318)

	zoneCfgExtensions, err := regionEnum.ZoneConfigExtensions()
	if err != nil {
		__antithesis_instrumentation__.Notify(565334)
		return regionConfig, err
	} else {
		__antithesis_instrumentation__.Notify(565335)
	}
	__antithesis_instrumentation__.Notify(565319)

	transitioningRegionNames, err := regionEnum.TransitioningRegionNames()
	if err != nil {
		__antithesis_instrumentation__.Notify(565336)
		return regionConfig, err
	} else {
		__antithesis_instrumentation__.Notify(565337)
	}
	__antithesis_instrumentation__.Notify(565320)

	superRegions, err := regionEnum.SuperRegions()
	if err != nil {
		__antithesis_instrumentation__.Notify(565338)
		return regionConfig, err
	} else {
		__antithesis_instrumentation__.Notify(565339)
	}
	__antithesis_instrumentation__.Notify(565321)

	regionConfig = multiregion.MakeRegionConfig(
		regionNames,
		dbDesc.GetRegionConfig().PrimaryRegion,
		dbDesc.GetRegionConfig().SurvivalGoal,
		regionEnumID,
		dbDesc.GetRegionConfig().Placement,
		superRegions,
		zoneCfgExtensions,
		multiregion.WithTransitioningRegions(transitioningRegionNames),
	)

	if err := multiregion.ValidateRegionConfig(regionConfig); err != nil {
		__antithesis_instrumentation__.Notify(565340)
		return multiregion.RegionConfig{}, err
	} else {
		__antithesis_instrumentation__.Notify(565341)
	}
	__antithesis_instrumentation__.Notify(565322)

	return regionConfig, nil
}

func blockDiscardOfZoneConfigForMultiRegionObject(
	zs tree.ZoneSpecifier, tblDesc catalog.TableDescriptor,
) (bool, error) {
	__antithesis_instrumentation__.Notify(565342)
	isIndex := zs.TableOrIndex.Index != ""
	isPartition := zs.Partition != ""

	if isPartition {
		__antithesis_instrumentation__.Notify(565344)

		if tblDesc.IsLocalityRegionalByRow() {
			__antithesis_instrumentation__.Notify(565345)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(565346)
		}
	} else {
		__antithesis_instrumentation__.Notify(565347)
		if isIndex {
			__antithesis_instrumentation__.Notify(565348)

			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(565349)

			if tblDesc.IsLocalityGlobal() {
				__antithesis_instrumentation__.Notify(565350)
				return true, nil
			} else {
				__antithesis_instrumentation__.Notify(565351)
				if tblDesc.IsLocalityRegionalByTable() {
					__antithesis_instrumentation__.Notify(565352)
					if tblDesc.GetLocalityConfig().GetRegionalByTable().Region != nil && func() bool {
						__antithesis_instrumentation__.Notify(565353)
						return tree.Name(*tblDesc.GetLocalityConfig().GetRegionalByTable().Region) !=
							tree.PrimaryRegionNotSpecifiedName == true
					}() == true {
						__antithesis_instrumentation__.Notify(565354)
						return true, nil
					} else {
						__antithesis_instrumentation__.Notify(565355)
					}
				} else {
					__antithesis_instrumentation__.Notify(565356)
					if tblDesc.IsLocalityRegionalByRow() {
						__antithesis_instrumentation__.Notify(565357)

						return false, nil
					} else {
						__antithesis_instrumentation__.Notify(565358)
						return false, errors.AssertionFailedf(
							"unknown table locality: %v",
							tblDesc.GetLocalityConfig(),
						)
					}
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(565343)
	return false, nil
}

func (p *planner) CheckZoneConfigChangePermittedForMultiRegion(
	ctx context.Context, zs tree.ZoneSpecifier, options tree.KVOptions,
) error {
	__antithesis_instrumentation__.Notify(565359)

	if p.SessionData().OverrideMultiRegionZoneConfigEnabled {
		__antithesis_instrumentation__.Notify(565364)

		telemetry.Inc(sqltelemetry.OverrideMultiRegionZoneConfigurationUser)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(565365)
	}
	__antithesis_instrumentation__.Notify(565360)

	var err error
	var tblDesc catalog.TableDescriptor
	isDB := false

	if zs.Database != "" {
		__antithesis_instrumentation__.Notify(565366)
		isDB = true
		dbDesc, err := p.Descriptors().GetImmutableDatabaseByName(
			ctx,
			p.txn,
			string(zs.Database),
			tree.DatabaseLookupFlags{Required: true},
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(565368)
			return err
		} else {
			__antithesis_instrumentation__.Notify(565369)
		}
		__antithesis_instrumentation__.Notify(565367)
		if dbDesc.GetRegionConfig() == nil {
			__antithesis_instrumentation__.Notify(565370)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(565371)
		}
	} else {
		__antithesis_instrumentation__.Notify(565372)

		tblDesc, err = p.resolveTableForZone(ctx, &zs)
		if err != nil {
			__antithesis_instrumentation__.Notify(565374)
			return err
		} else {
			__antithesis_instrumentation__.Notify(565375)
		}
		__antithesis_instrumentation__.Notify(565373)
		if tblDesc == nil || func() bool {
			__antithesis_instrumentation__.Notify(565376)
			return tblDesc.GetLocalityConfig() == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(565377)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(565378)
		}
	}
	__antithesis_instrumentation__.Notify(565361)

	hint := "to override this error, SET override_multi_region_zone_config = true and reissue the command"

	if options == nil {
		__antithesis_instrumentation__.Notify(565379)
		needToError := false

		if isDB {
			__antithesis_instrumentation__.Notify(565381)
			needToError = true
		} else {
			__antithesis_instrumentation__.Notify(565382)
			needToError, err = blockDiscardOfZoneConfigForMultiRegionObject(zs, tblDesc)
			if err != nil {
				__antithesis_instrumentation__.Notify(565383)
				return err
			} else {
				__antithesis_instrumentation__.Notify(565384)
			}
		}
		__antithesis_instrumentation__.Notify(565380)

		if needToError {
			__antithesis_instrumentation__.Notify(565385)

			err := errors.WithDetail(errors.Newf(
				"attempting to discard the zone configuration of a multi-region entity"),
				"discarding a multi-region zone configuration may result in sub-optimal performance or behavior",
			)
			return errors.WithHint(err, hint)
		} else {
			__antithesis_instrumentation__.Notify(565386)
		}
	} else {
		__antithesis_instrumentation__.Notify(565387)
	}
	__antithesis_instrumentation__.Notify(565362)

	for _, opt := range options {
		__antithesis_instrumentation__.Notify(565388)
		for _, cfg := range zonepb.MultiRegionZoneConfigFields {
			__antithesis_instrumentation__.Notify(565389)
			if opt.Key == cfg {
				__antithesis_instrumentation__.Notify(565390)

				err := errors.Newf("attempting to modify protected field %q of a multi-region zone configuration",
					string(opt.Key),
				)
				return errors.WithHint(err, hint)
			} else {
				__antithesis_instrumentation__.Notify(565391)
			}
		}
	}
	__antithesis_instrumentation__.Notify(565363)

	return nil
}

type zoneConfigForMultiRegionValidator interface {
	getExpectedDatabaseZoneConfig() (zonepb.ZoneConfig, error)
	getExpectedTableZoneConfig(desc catalog.TableDescriptor) (zonepb.ZoneConfig, error)
	transitioningRegions() catpb.RegionNames

	newMismatchFieldError(descType string, descName string, field string) error
	newMissingSubzoneError(descType string, descName string, field string) error
	newExtraSubzoneError(descType string, descName string, field string) error
}

type zoneConfigForMultiRegionValidatorSetInitialRegion struct{}

var _ zoneConfigForMultiRegionValidator = (*zoneConfigForMultiRegionValidatorSetInitialRegion)(nil)

func (v *zoneConfigForMultiRegionValidatorSetInitialRegion) getExpectedDatabaseZoneConfig() (
	zonepb.ZoneConfig,
	error,
) {
	__antithesis_instrumentation__.Notify(565392)

	return *zonepb.NewZoneConfig(), nil
}

func (v *zoneConfigForMultiRegionValidatorSetInitialRegion) transitioningRegions() catpb.RegionNames {
	__antithesis_instrumentation__.Notify(565393)

	return nil
}

func (v *zoneConfigForMultiRegionValidatorSetInitialRegion) getExpectedTableZoneConfig(
	desc catalog.TableDescriptor,
) (zonepb.ZoneConfig, error) {
	__antithesis_instrumentation__.Notify(565394)

	return *zonepb.NewZoneConfig(), nil
}

func (v *zoneConfigForMultiRegionValidatorSetInitialRegion) wrapErr(err error) error {
	__antithesis_instrumentation__.Notify(565395)

	return errors.WithHintf(
		err,
		"discard the zone config using CONFIGURE ZONE DISCARD before continuing",
	)
}

func (v *zoneConfigForMultiRegionValidatorSetInitialRegion) newMismatchFieldError(
	descType string, descName string, field string,
) error {
	__antithesis_instrumentation__.Notify(565396)
	return v.wrapErr(
		pgerror.Newf(
			pgcode.InvalidObjectDefinition,
			"zone configuration for %s %s has field %q set which will be overwritten when setting the the initial PRIMARY REGION",
			descType,
			descName,
			field,
		),
	)
}

func (v *zoneConfigForMultiRegionValidatorSetInitialRegion) newMissingSubzoneError(
	descType string, descName string, field string,
) error {
	__antithesis_instrumentation__.Notify(565397)

	return errors.AssertionFailedf(
		"unexpected missing subzone for %s %s",
		descType,
		descName,
	)
}

func (v *zoneConfigForMultiRegionValidatorSetInitialRegion) newExtraSubzoneError(
	descType string, descName string, field string,
) error {
	__antithesis_instrumentation__.Notify(565398)
	return v.wrapErr(
		pgerror.Newf(
			pgcode.InvalidObjectDefinition,
			"zone configuration for %s %s has field %q set which will be overwritten when setting the initial PRIMARY REGION",
			descType,
			descName,
			field,
		),
	)
}

type zoneConfigForMultiRegionValidatorExistingMultiRegionObject struct {
	regionConfig multiregion.RegionConfig
}

func (v *zoneConfigForMultiRegionValidatorExistingMultiRegionObject) getExpectedDatabaseZoneConfig() (
	zonepb.ZoneConfig,
	error,
) {
	__antithesis_instrumentation__.Notify(565399)
	return zoneConfigForMultiRegionDatabase(v.regionConfig)
}

func (v *zoneConfigForMultiRegionValidatorExistingMultiRegionObject) getExpectedTableZoneConfig(
	desc catalog.TableDescriptor,
) (zonepb.ZoneConfig, error) {
	__antithesis_instrumentation__.Notify(565400)
	_, expectedZoneConfig, err := ApplyZoneConfigForMultiRegionTableOptionTableAndIndexes(
		*zonepb.NewZoneConfig(),
		v.regionConfig,
		desc,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(565402)
		return zonepb.ZoneConfig{}, err
	} else {
		__antithesis_instrumentation__.Notify(565403)
	}
	__antithesis_instrumentation__.Notify(565401)
	return expectedZoneConfig, err
}

func (v *zoneConfigForMultiRegionValidatorExistingMultiRegionObject) transitioningRegions() catpb.RegionNames {
	__antithesis_instrumentation__.Notify(565404)
	return v.regionConfig.TransitioningRegions()
}

type zoneConfigForMultiRegionValidatorModifiedByUser struct {
	zoneConfigForMultiRegionValidatorExistingMultiRegionObject
}

var _ zoneConfigForMultiRegionValidator = (*zoneConfigForMultiRegionValidatorModifiedByUser)(nil)

func (v *zoneConfigForMultiRegionValidatorModifiedByUser) newMismatchFieldError(
	descType string, descName string, field string,
) error {
	__antithesis_instrumentation__.Notify(565405)
	return v.wrapErr(
		pgerror.Newf(
			pgcode.InvalidObjectDefinition,
			"attempting to update zone configuration for %s %s which contains modified field %q",
			descType,
			descName,
			field,
		),
	)
}

func (v *zoneConfigForMultiRegionValidatorModifiedByUser) wrapErr(err error) error {
	__antithesis_instrumentation__.Notify(565406)
	err = errors.WithDetail(
		err,
		"the attempted operation will overwrite a user modified field",
	)
	return errors.WithHint(
		err,
		"to proceed with the overwrite, SET override_multi_region_zone_config = true, "+
			"and reissue the statement",
	)
}

func (v *zoneConfigForMultiRegionValidatorModifiedByUser) newMissingSubzoneError(
	descType string, descName string, field string,
) error {
	__antithesis_instrumentation__.Notify(565407)
	return v.wrapErr(
		pgerror.Newf(
			pgcode.InvalidObjectDefinition,
			"attempting to update zone config which is missing an expected zone configuration for %s %s",
			descType,
			descName,
		),
	)
}

func (v *zoneConfigForMultiRegionValidatorModifiedByUser) newExtraSubzoneError(
	descType string, descName string, field string,
) error {
	__antithesis_instrumentation__.Notify(565408)
	return v.wrapErr(
		pgerror.Newf(
			pgcode.InvalidObjectDefinition,
			"attempting to update zone config which contains an extra zone configuration for %s %s with field %s populated",
			descType,
			descName,
			field,
		),
	)
}

type zoneConfigForMultiRegionValidatorValidation struct {
	zoneConfigForMultiRegionValidatorExistingMultiRegionObject
}

var _ zoneConfigForMultiRegionValidator = (*zoneConfigForMultiRegionValidatorValidation)(nil)

func (v *zoneConfigForMultiRegionValidatorValidation) newMismatchFieldError(
	descType string, descName string, field string,
) error {
	__antithesis_instrumentation__.Notify(565409)
	return pgerror.Newf(
		pgcode.InvalidObjectDefinition,
		"zone configuration for %s %s contains incorrectly configured field %q",
		descType,
		descName,
		field,
	)
}

func (v *zoneConfigForMultiRegionValidatorValidation) newMissingSubzoneError(
	descType string, descName string, field string,
) error {
	__antithesis_instrumentation__.Notify(565410)
	return pgerror.Newf(
		pgcode.InvalidObjectDefinition,
		"missing zone configuration for %s %s",
		descType,
		descName,
	)
}

func (v *zoneConfigForMultiRegionValidatorValidation) newExtraSubzoneError(
	descType string, descName string, field string,
) error {
	__antithesis_instrumentation__.Notify(565411)
	return pgerror.Newf(
		pgcode.InvalidObjectDefinition,
		"extraneous zone configuration for %s %s with field %s populated",
		descType,
		descName,
		field,
	)
}

func (p *planner) validateZoneConfigForMultiRegionDatabaseWasNotModifiedByUser(
	ctx context.Context, dbDesc catalog.DatabaseDescriptor,
) error {
	__antithesis_instrumentation__.Notify(565412)

	if p.SessionData().OverrideMultiRegionZoneConfigEnabled {
		__antithesis_instrumentation__.Notify(565416)
		telemetry.Inc(sqltelemetry.OverrideMultiRegionDatabaseZoneConfigurationSystem)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(565417)
	}
	__antithesis_instrumentation__.Notify(565413)
	currentZoneConfig, err := getZoneConfigRaw(ctx, p.txn, p.ExecCfg().Codec, p.ExecCfg().Settings, dbDesc.GetID())
	if err != nil {
		__antithesis_instrumentation__.Notify(565418)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565419)
	}
	__antithesis_instrumentation__.Notify(565414)
	regionConfig, err := SynthesizeRegionConfig(
		ctx,
		p.txn,
		dbDesc.GetID(),
		p.Descriptors(),
		SynthesizeRegionConfigOptionForValidation,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(565420)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565421)
	}
	__antithesis_instrumentation__.Notify(565415)
	return p.validateZoneConfigForMultiRegionDatabase(
		dbDesc,
		currentZoneConfig,
		&zoneConfigForMultiRegionValidatorModifiedByUser{
			zoneConfigForMultiRegionValidatorExistingMultiRegionObject: zoneConfigForMultiRegionValidatorExistingMultiRegionObject{
				regionConfig: regionConfig,
			},
		},
	)
}

func (p *planner) validateZoneConfigForMultiRegionDatabase(
	dbDesc catalog.DatabaseDescriptor,
	currentZoneConfig *zonepb.ZoneConfig,
	zoneConfigForMultiRegionValidator zoneConfigForMultiRegionValidator,
) error {
	__antithesis_instrumentation__.Notify(565422)
	if currentZoneConfig == nil {
		__antithesis_instrumentation__.Notify(565427)
		currentZoneConfig = zonepb.NewZoneConfig()
	} else {
		__antithesis_instrumentation__.Notify(565428)
	}
	__antithesis_instrumentation__.Notify(565423)
	expectedZoneConfig, err := zoneConfigForMultiRegionValidator.getExpectedDatabaseZoneConfig()
	if err != nil {
		__antithesis_instrumentation__.Notify(565429)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565430)
	}
	__antithesis_instrumentation__.Notify(565424)

	same, mismatch, err := currentZoneConfig.DiffWithZone(
		expectedZoneConfig,
		zonepb.MultiRegionZoneConfigFields,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(565431)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565432)
	}
	__antithesis_instrumentation__.Notify(565425)
	if !same {
		__antithesis_instrumentation__.Notify(565433)
		dbName := tree.Name(dbDesc.GetName())
		return zoneConfigForMultiRegionValidator.newMismatchFieldError(
			"database",
			dbName.String(),
			mismatch.Field,
		)
	} else {
		__antithesis_instrumentation__.Notify(565434)
	}
	__antithesis_instrumentation__.Notify(565426)

	return nil
}

func (p *planner) validateZoneConfigForMultiRegionTableWasNotModifiedByUser(
	ctx context.Context, dbDesc catalog.DatabaseDescriptor, desc *tabledesc.Mutable,
) error {
	__antithesis_instrumentation__.Notify(565435)

	if p.SessionData().OverrideMultiRegionZoneConfigEnabled || func() bool {
		__antithesis_instrumentation__.Notify(565439)
		return desc.GetLocalityConfig() == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(565440)
		telemetry.Inc(sqltelemetry.OverrideMultiRegionTableZoneConfigurationSystem)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(565441)
	}
	__antithesis_instrumentation__.Notify(565436)
	currentZoneConfig, err := getZoneConfigRaw(ctx, p.txn, p.ExecCfg().Codec, p.ExecCfg().Settings, desc.GetID())
	if err != nil {
		__antithesis_instrumentation__.Notify(565442)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565443)
	}
	__antithesis_instrumentation__.Notify(565437)
	regionConfig, err := SynthesizeRegionConfig(
		ctx,
		p.txn,
		dbDesc.GetID(),
		p.Descriptors(),
		SynthesizeRegionConfigOptionForValidation,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(565444)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565445)
	}
	__antithesis_instrumentation__.Notify(565438)

	return p.validateZoneConfigForMultiRegionTable(
		desc,
		currentZoneConfig,
		&zoneConfigForMultiRegionValidatorModifiedByUser{
			zoneConfigForMultiRegionValidatorExistingMultiRegionObject: zoneConfigForMultiRegionValidatorExistingMultiRegionObject{
				regionConfig: regionConfig,
			},
		},
	)
}

func (p *planner) validateZoneConfigForMultiRegionTable(
	desc catalog.TableDescriptor,
	currentZoneConfig *zonepb.ZoneConfig,
	zoneConfigForMultiRegionValidator zoneConfigForMultiRegionValidator,
) error {
	__antithesis_instrumentation__.Notify(565446)
	if currentZoneConfig == nil {
		__antithesis_instrumentation__.Notify(565458)
		currentZoneConfig = zonepb.NewZoneConfig()
	} else {
		__antithesis_instrumentation__.Notify(565459)
	}
	__antithesis_instrumentation__.Notify(565447)

	tableName := tree.Name(desc.GetName())

	expectedZoneConfig, err := zoneConfigForMultiRegionValidator.getExpectedTableZoneConfig(
		desc,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(565460)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565461)
	}
	__antithesis_instrumentation__.Notify(565448)

	regionalByRowNewIndexes := make(map[uint32]struct{})
	for _, mut := range desc.AllMutations() {
		__antithesis_instrumentation__.Notify(565462)
		if pkSwap := mut.AsPrimaryKeySwap(); pkSwap != nil {
			__antithesis_instrumentation__.Notify(565463)
			if pkSwap.HasLocalityConfig() {
				__antithesis_instrumentation__.Notify(565465)
				_ = pkSwap.ForEachNewIndexIDs(func(id descpb.IndexID) error {
					__antithesis_instrumentation__.Notify(565466)
					regionalByRowNewIndexes[uint32(id)] = struct{}{}
					if idx := catalog.FindCorrespondingTemporaryIndexByID(desc, id); idx != nil {
						__antithesis_instrumentation__.Notify(565468)
						regionalByRowNewIndexes[uint32(idx.GetID())] = struct{}{}
					} else {
						__antithesis_instrumentation__.Notify(565469)
					}
					__antithesis_instrumentation__.Notify(565467)
					return nil
				})

			} else {
				__antithesis_instrumentation__.Notify(565470)
			}
			__antithesis_instrumentation__.Notify(565464)

			break
		} else {
			__antithesis_instrumentation__.Notify(565471)
		}
	}
	__antithesis_instrumentation__.Notify(565449)

	subzoneIndexIDsToDiff := make(map[uint32]tree.Name, len(desc.NonDropIndexes()))
	for _, idx := range desc.NonDropIndexes() {
		__antithesis_instrumentation__.Notify(565472)
		if _, ok := regionalByRowNewIndexes[uint32(idx.GetID())]; !ok {
			__antithesis_instrumentation__.Notify(565473)
			subzoneIndexIDsToDiff[uint32(idx.GetID())] = tree.Name(idx.GetName())
		} else {
			__antithesis_instrumentation__.Notify(565474)
		}
	}
	__antithesis_instrumentation__.Notify(565450)

	transitioningRegions := make(map[string]struct{}, len(zoneConfigForMultiRegionValidator.transitioningRegions()))
	for _, transitioningRegion := range zoneConfigForMultiRegionValidator.transitioningRegions() {
		__antithesis_instrumentation__.Notify(565475)
		transitioningRegions[string(transitioningRegion)] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(565451)

	filteredCurrentZoneConfigSubzones := currentZoneConfig.Subzones[:0]
	for _, c := range currentZoneConfig.Subzones {
		__antithesis_instrumentation__.Notify(565476)
		if c.PartitionName != "" {
			__antithesis_instrumentation__.Notify(565479)
			if _, ok := transitioningRegions[c.PartitionName]; ok {
				__antithesis_instrumentation__.Notify(565480)
				continue
			} else {
				__antithesis_instrumentation__.Notify(565481)
			}
		} else {
			__antithesis_instrumentation__.Notify(565482)
		}
		__antithesis_instrumentation__.Notify(565477)
		if _, ok := subzoneIndexIDsToDiff[c.IndexID]; !ok {
			__antithesis_instrumentation__.Notify(565483)
			continue
		} else {
			__antithesis_instrumentation__.Notify(565484)
		}
		__antithesis_instrumentation__.Notify(565478)
		filteredCurrentZoneConfigSubzones = append(filteredCurrentZoneConfigSubzones, c)
	}
	__antithesis_instrumentation__.Notify(565452)
	currentZoneConfig.Subzones = filteredCurrentZoneConfigSubzones

	if len(filteredCurrentZoneConfigSubzones) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(565485)
		return currentZoneConfig.IsSubzonePlaceholder() == true
	}() == true {
		__antithesis_instrumentation__.Notify(565486)
		currentZoneConfig.NumReplicas = nil
	} else {
		__antithesis_instrumentation__.Notify(565487)
	}
	__antithesis_instrumentation__.Notify(565453)

	filteredExpectedZoneConfigSubzones := expectedZoneConfig.Subzones[:0]
	for _, c := range expectedZoneConfig.Subzones {
		__antithesis_instrumentation__.Notify(565488)
		if c.PartitionName != "" {
			__antithesis_instrumentation__.Notify(565491)
			if _, ok := transitioningRegions[c.PartitionName]; ok {
				__antithesis_instrumentation__.Notify(565492)
				continue
			} else {
				__antithesis_instrumentation__.Notify(565493)
			}
		} else {
			__antithesis_instrumentation__.Notify(565494)
		}
		__antithesis_instrumentation__.Notify(565489)
		if _, ok := regionalByRowNewIndexes[c.IndexID]; ok {
			__antithesis_instrumentation__.Notify(565495)
			continue
		} else {
			__antithesis_instrumentation__.Notify(565496)
		}
		__antithesis_instrumentation__.Notify(565490)
		filteredExpectedZoneConfigSubzones = append(
			filteredExpectedZoneConfigSubzones,
			c,
		)
	}
	__antithesis_instrumentation__.Notify(565454)
	expectedZoneConfig.Subzones = filteredExpectedZoneConfigSubzones

	if currentZoneConfig.IsSubzonePlaceholder() && func() bool {
		__antithesis_instrumentation__.Notify(565497)
		return isPlaceholderZoneConfigForMultiRegion(expectedZoneConfig) == true
	}() == true {
		__antithesis_instrumentation__.Notify(565498)
		expectedZoneConfig.NumReplicas = proto.Int32(0)
	} else {
		__antithesis_instrumentation__.Notify(565499)
	}
	__antithesis_instrumentation__.Notify(565455)

	same, mismatch, err := currentZoneConfig.DiffWithZone(
		expectedZoneConfig,
		zonepb.MultiRegionZoneConfigFields,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(565500)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565501)
	}
	__antithesis_instrumentation__.Notify(565456)
	if !same {
		__antithesis_instrumentation__.Notify(565502)
		descType := "table"
		name := tableName.String()
		if mismatch.IndexID != 0 {
			__antithesis_instrumentation__.Notify(565506)
			indexName, ok := subzoneIndexIDsToDiff[mismatch.IndexID]
			if !ok {
				__antithesis_instrumentation__.Notify(565508)
				return errors.AssertionFailedf(
					"unexpected unknown index id %d on table %s (mismatch %#v)",
					mismatch.IndexID,
					tableName,
					mismatch,
				)
			} else {
				__antithesis_instrumentation__.Notify(565509)
			}
			__antithesis_instrumentation__.Notify(565507)

			if mismatch.PartitionName != "" {
				__antithesis_instrumentation__.Notify(565510)
				descType = "partition"
				partitionName := tree.Name(mismatch.PartitionName)
				name = fmt.Sprintf(
					"%s of %s@%s",
					partitionName.String(),
					tableName.String(),
					indexName.String(),
				)
			} else {
				__antithesis_instrumentation__.Notify(565511)
				descType = "index"
				name = fmt.Sprintf("%s@%s", tableName.String(), indexName.String())
			}
		} else {
			__antithesis_instrumentation__.Notify(565512)
		}
		__antithesis_instrumentation__.Notify(565503)

		if mismatch.IsMissingSubzone {
			__antithesis_instrumentation__.Notify(565513)
			return zoneConfigForMultiRegionValidator.newMissingSubzoneError(
				descType,
				name,
				mismatch.Field,
			)
		} else {
			__antithesis_instrumentation__.Notify(565514)
		}
		__antithesis_instrumentation__.Notify(565504)
		if mismatch.IsExtraSubzone {
			__antithesis_instrumentation__.Notify(565515)
			return zoneConfigForMultiRegionValidator.newExtraSubzoneError(
				descType,
				name,
				mismatch.Field,
			)
		} else {
			__antithesis_instrumentation__.Notify(565516)
		}
		__antithesis_instrumentation__.Notify(565505)

		return zoneConfigForMultiRegionValidator.newMismatchFieldError(
			descType,
			name,
			mismatch.Field,
		)
	} else {
		__antithesis_instrumentation__.Notify(565517)
	}
	__antithesis_instrumentation__.Notify(565457)

	return nil
}

func (p *planner) checkNoRegionalByRowChangeUnderway(
	ctx context.Context, dbDesc catalog.DatabaseDescriptor,
) error {
	__antithesis_instrumentation__.Notify(565518)

	return p.forEachMutableTableInDatabase(
		ctx,
		dbDesc,
		func(ctx context.Context, scName string, table *tabledesc.Mutable) error {
			__antithesis_instrumentation__.Notify(565519)
			wrapErr := func(err error, detailSuffix string) error {
				__antithesis_instrumentation__.Notify(565523)
				return errors.WithHintf(
					errors.WithDetailf(
						err,
						"table %s.%s %s",
						tree.Name(scName),
						tree.Name(table.GetName()),
						detailSuffix,
					),
					"cancel the existing job or try again when the change is complete",
				)
			}
			__antithesis_instrumentation__.Notify(565520)
			for _, mut := range table.AllMutations() {
				__antithesis_instrumentation__.Notify(565524)

				if pkSwap := mut.AsPrimaryKeySwap(); pkSwap != nil {
					__antithesis_instrumentation__.Notify(565525)
					if lcSwap := pkSwap.PrimaryKeySwapDesc().LocalityConfigSwap; lcSwap != nil {
						__antithesis_instrumentation__.Notify(565527)
						return wrapErr(
							pgerror.Newf(
								pgcode.ObjectNotInPrerequisiteState,
								"cannot perform database region changes while a REGIONAL BY ROW transition is underway",
							),
							"is currently transitioning to or from REGIONAL BY ROW",
						)
					} else {
						__antithesis_instrumentation__.Notify(565528)
					}
					__antithesis_instrumentation__.Notify(565526)
					return wrapErr(
						pgerror.Newf(
							pgcode.ObjectNotInPrerequisiteState,
							"cannot perform database region changes while a ALTER PRIMARY KEY is underway",
						),
						"is currently undergoing an ALTER PRIMARY KEY change",
					)
				} else {
					__antithesis_instrumentation__.Notify(565529)
				}
			}
			__antithesis_instrumentation__.Notify(565521)

			for _, mut := range table.AllMutations() {
				__antithesis_instrumentation__.Notify(565530)
				if table.IsLocalityRegionalByRow() {
					__antithesis_instrumentation__.Notify(565531)
					if idx := mut.AsIndex(); idx != nil {
						__antithesis_instrumentation__.Notify(565532)
						return wrapErr(
							pgerror.Newf(
								pgcode.ObjectNotInPrerequisiteState,
								"cannot perform database region changes while an index is being created or dropped on a REGIONAL BY ROW table",
							),
							fmt.Sprintf("is currently modifying index %s", tree.Name(idx.GetName())),
						)
					} else {
						__antithesis_instrumentation__.Notify(565533)
					}
				} else {
					__antithesis_instrumentation__.Notify(565534)
				}
			}
			__antithesis_instrumentation__.Notify(565522)
			return nil
		},
	)
}

func (p *planner) checkNoRegionChangeUnderway(
	ctx context.Context, dbID descpb.ID, op string,
) error {
	__antithesis_instrumentation__.Notify(565535)

	r, err := SynthesizeRegionConfig(
		ctx,
		p.txn,
		dbID,
		p.Descriptors(),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(565538)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565539)
	}
	__antithesis_instrumentation__.Notify(565536)
	if len(r.TransitioningRegions()) > 0 {
		__antithesis_instrumentation__.Notify(565540)
		return errors.WithDetailf(
			errors.WithHintf(
				pgerror.Newf(
					pgcode.ObjectNotInPrerequisiteState,
					"cannot %s while a region is being added or dropped on the database",
					op,
				),
				"cancel the job which is adding or dropping the region or try again later",
			),
			"region %s is currently being added or dropped",
			r.TransitioningRegions()[0],
		)
	} else {
		__antithesis_instrumentation__.Notify(565541)
	}
	__antithesis_instrumentation__.Notify(565537)
	return nil
}
