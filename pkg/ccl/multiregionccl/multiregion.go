package multiregionccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/ccl/utilccl"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descidgen"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/multiregion"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

func init() {
	sql.InitializeMultiRegionMetadataCCL = initializeMultiRegionMetadata
	sql.GetMultiRegionEnumAddValuePlacementCCL = getMultiRegionEnumAddValuePlacement
}

func initializeMultiRegionMetadata(
	ctx context.Context,
	execCfg *sql.ExecutorConfig,
	liveRegions sql.LiveClusterRegions,
	goal tree.SurvivalGoal,
	primaryRegion catpb.RegionName,
	regions []tree.Name,
	dataPlacement tree.DataPlacement,
) (*multiregion.RegionConfig, error) {
	__antithesis_instrumentation__.Notify(19878)
	if err := CheckClusterSupportsMultiRegion(execCfg); err != nil {
		__antithesis_instrumentation__.Notify(19888)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(19889)
	}
	__antithesis_instrumentation__.Notify(19879)

	survivalGoal, err := sql.TranslateSurvivalGoal(goal)
	if err != nil {
		__antithesis_instrumentation__.Notify(19890)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(19891)
	}
	__antithesis_instrumentation__.Notify(19880)
	placement, err := sql.TranslateDataPlacement(dataPlacement)
	if err != nil {
		__antithesis_instrumentation__.Notify(19892)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(19893)
	}
	__antithesis_instrumentation__.Notify(19881)

	if primaryRegion != catpb.RegionName(tree.PrimaryRegionNotSpecifiedName) {
		__antithesis_instrumentation__.Notify(19894)
		if err := sql.CheckClusterRegionIsLive(liveRegions, primaryRegion); err != nil {
			__antithesis_instrumentation__.Notify(19895)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(19896)
		}
	} else {
		__antithesis_instrumentation__.Notify(19897)
	}
	__antithesis_instrumentation__.Notify(19882)
	regionNames := make(catpb.RegionNames, 0, len(regions)+1)
	seenRegions := make(map[catpb.RegionName]struct{}, len(regions)+1)
	if len(regions) > 0 {
		__antithesis_instrumentation__.Notify(19898)
		if primaryRegion == catpb.RegionName(tree.PrimaryRegionNotSpecifiedName) {
			__antithesis_instrumentation__.Notify(19900)
			return nil, pgerror.Newf(
				pgcode.InvalidDatabaseDefinition,
				"PRIMARY REGION must be specified if REGIONS are specified",
			)
		} else {
			__antithesis_instrumentation__.Notify(19901)
		}
		__antithesis_instrumentation__.Notify(19899)
		for _, r := range regions {
			__antithesis_instrumentation__.Notify(19902)
			region := catpb.RegionName(r)
			if err := sql.CheckClusterRegionIsLive(liveRegions, region); err != nil {
				__antithesis_instrumentation__.Notify(19905)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(19906)
			}
			__antithesis_instrumentation__.Notify(19903)

			if _, ok := seenRegions[region]; ok {
				__antithesis_instrumentation__.Notify(19907)
				return nil, pgerror.Newf(
					pgcode.InvalidName,
					"region %q defined multiple times",
					region,
				)
			} else {
				__antithesis_instrumentation__.Notify(19908)
			}
			__antithesis_instrumentation__.Notify(19904)
			seenRegions[region] = struct{}{}
			regionNames = append(regionNames, region)
		}
	} else {
		__antithesis_instrumentation__.Notify(19909)
	}
	__antithesis_instrumentation__.Notify(19883)

	if _, ok := seenRegions[primaryRegion]; !ok {
		__antithesis_instrumentation__.Notify(19910)
		regionNames = append(regionNames, primaryRegion)
	} else {
		__antithesis_instrumentation__.Notify(19911)
	}
	__antithesis_instrumentation__.Notify(19884)

	sort.SliceStable(regionNames, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(19912)
		return regionNames[i] < regionNames[j]
	})
	__antithesis_instrumentation__.Notify(19885)

	regionEnumID, err := descidgen.GenerateUniqueDescID(ctx, execCfg.DB, execCfg.Codec)
	if err != nil {
		__antithesis_instrumentation__.Notify(19913)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(19914)
	}
	__antithesis_instrumentation__.Notify(19886)
	regionConfig := multiregion.MakeRegionConfig(
		regionNames,
		primaryRegion,
		survivalGoal,
		regionEnumID,
		placement,
		nil,
		descpb.ZoneConfigExtensions{},
	)
	if err := multiregion.ValidateRegionConfig(regionConfig); err != nil {
		__antithesis_instrumentation__.Notify(19915)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(19916)
	}
	__antithesis_instrumentation__.Notify(19887)

	return &regionConfig, nil
}

func CheckClusterSupportsMultiRegion(execCfg *sql.ExecutorConfig) error {
	__antithesis_instrumentation__.Notify(19917)
	return utilccl.CheckEnterpriseEnabled(
		execCfg.Settings,
		execCfg.LogicalClusterID(),
		execCfg.Organization(),
		"multi-region features",
	)
}

func getMultiRegionEnumAddValuePlacement(
	execCfg *sql.ExecutorConfig, typeDesc *typedesc.Mutable, region tree.Name,
) (tree.AlterTypeAddValue, error) {
	__antithesis_instrumentation__.Notify(19918)
	if err := utilccl.CheckEnterpriseEnabled(
		execCfg.Settings,
		execCfg.LogicalClusterID(),
		execCfg.Organization(),
		"ADD REGION",
	); err != nil {
		__antithesis_instrumentation__.Notify(19922)
		return tree.AlterTypeAddValue{}, err
	} else {
		__antithesis_instrumentation__.Notify(19923)
	}
	__antithesis_instrumentation__.Notify(19919)

	loc := sort.Search(
		len(typeDesc.EnumMembers),
		func(i int) bool {
			__antithesis_instrumentation__.Notify(19924)
			return string(region) < typeDesc.EnumMembers[i].LogicalRepresentation
		},
	)
	__antithesis_instrumentation__.Notify(19920)

	before := true
	if loc == len(typeDesc.EnumMembers) {
		__antithesis_instrumentation__.Notify(19925)
		before = false
		loc = len(typeDesc.EnumMembers) - 1
	} else {
		__antithesis_instrumentation__.Notify(19926)
	}
	__antithesis_instrumentation__.Notify(19921)

	return tree.AlterTypeAddValue{
		IfNotExists: false,
		NewVal:      tree.EnumValue(region),
		Placement: &tree.AlterTypeAddValuePlacement{
			Before:      before,
			ExistingVal: tree.EnumValue(typeDesc.EnumMembers[loc].LogicalRepresentation),
		},
	}, nil
}
