// Package multiregion provides functions and structs for interacting with the
// static multi-region state configured by users on their databases.
package multiregion

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"sort"

	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/errors"
)

const minNumRegionsForSurviveRegionGoal = 3

type RegionConfig struct {
	survivalGoal         descpb.SurvivalGoal
	regions              catpb.RegionNames
	transitioningRegions catpb.RegionNames
	primaryRegion        catpb.RegionName
	regionEnumID         descpb.ID
	placement            descpb.DataPlacement
	superRegions         []descpb.SuperRegion
	zoneCfgExtensions    descpb.ZoneConfigExtensions
}

func (r *RegionConfig) SurvivalGoal() descpb.SurvivalGoal {
	__antithesis_instrumentation__.Notify(266691)
	return r.survivalGoal
}

func (r *RegionConfig) PrimaryRegion() catpb.RegionName {
	__antithesis_instrumentation__.Notify(266692)
	return r.primaryRegion
}

func (r *RegionConfig) Regions() catpb.RegionNames {
	__antithesis_instrumentation__.Notify(266693)
	return r.regions
}

func (r *RegionConfig) WithRegions(regions catpb.RegionNames) RegionConfig {
	__antithesis_instrumentation__.Notify(266694)
	cpy := *r
	cpy.regions = regions
	return cpy
}

func (r *RegionConfig) IsMemberOfExplicitSuperRegion(region catpb.RegionName) bool {
	__antithesis_instrumentation__.Notify(266695)
	for _, superRegion := range r.SuperRegions() {
		__antithesis_instrumentation__.Notify(266697)
		for _, regionOfSuperRegion := range superRegion.Regions {
			__antithesis_instrumentation__.Notify(266698)
			if region == regionOfSuperRegion {
				__antithesis_instrumentation__.Notify(266699)
				return true
			} else {
				__antithesis_instrumentation__.Notify(266700)
			}
		}
	}
	__antithesis_instrumentation__.Notify(266696)
	return false
}

func (r *RegionConfig) GetSuperRegionRegionsForRegion(
	region catpb.RegionName,
) (catpb.RegionNames, error) {
	__antithesis_instrumentation__.Notify(266701)
	for _, superRegion := range r.SuperRegions() {
		__antithesis_instrumentation__.Notify(266703)
		for _, regionOfSuperRegion := range superRegion.Regions {
			__antithesis_instrumentation__.Notify(266704)
			if region == regionOfSuperRegion {
				__antithesis_instrumentation__.Notify(266705)
				return superRegion.Regions, nil
			} else {
				__antithesis_instrumentation__.Notify(266706)
			}
		}
	}
	__antithesis_instrumentation__.Notify(266702)
	return nil, errors.AssertionFailedf("region %s is not part of a super region", region)
}

func (r RegionConfig) IsValidRegionNameString(regionStr string) bool {
	__antithesis_instrumentation__.Notify(266707)
	for _, region := range r.Regions() {
		__antithesis_instrumentation__.Notify(266709)
		if string(region) == regionStr {
			__antithesis_instrumentation__.Notify(266710)
			return true
		} else {
			__antithesis_instrumentation__.Notify(266711)
		}
	}
	__antithesis_instrumentation__.Notify(266708)
	return false
}

func (r RegionConfig) PrimaryRegionString() string {
	__antithesis_instrumentation__.Notify(266712)
	return string(r.PrimaryRegion())
}

func (r RegionConfig) TransitioningRegions() catpb.RegionNames {
	__antithesis_instrumentation__.Notify(266713)
	return r.transitioningRegions
}

func (r *RegionConfig) RegionEnumID() descpb.ID {
	__antithesis_instrumentation__.Notify(266714)
	return r.regionEnumID
}

func (r *RegionConfig) Placement() descpb.DataPlacement {
	__antithesis_instrumentation__.Notify(266715)
	return r.placement
}

func (r *RegionConfig) IsPlacementRestricted() bool {
	__antithesis_instrumentation__.Notify(266716)
	return r.placement == descpb.DataPlacement_RESTRICTED
}

func (r *RegionConfig) WithPlacementDefault() RegionConfig {
	__antithesis_instrumentation__.Notify(266717)
	cpy := *r
	cpy.placement = descpb.DataPlacement_DEFAULT
	return cpy
}

func (r *RegionConfig) SuperRegions() []descpb.SuperRegion {
	__antithesis_instrumentation__.Notify(266718)
	return r.superRegions
}

func (r *RegionConfig) ZoneConfigExtensions() descpb.ZoneConfigExtensions {
	__antithesis_instrumentation__.Notify(266719)
	return r.zoneCfgExtensions
}

func (r *RegionConfig) ApplyZoneConfigExtensionForGlobal(zc zonepb.ZoneConfig) zonepb.ZoneConfig {
	__antithesis_instrumentation__.Notify(266720)
	if ext := r.zoneCfgExtensions.Global; ext != nil {
		__antithesis_instrumentation__.Notify(266722)
		zc = extendZoneCfg(zc, *ext)
	} else {
		__antithesis_instrumentation__.Notify(266723)
	}
	__antithesis_instrumentation__.Notify(266721)
	return zc
}

func (r *RegionConfig) ApplyZoneConfigExtensionForRegionalIn(
	zc zonepb.ZoneConfig, region catpb.RegionName,
) zonepb.ZoneConfig {
	__antithesis_instrumentation__.Notify(266724)
	if ext := r.zoneCfgExtensions.Regional; ext != nil {
		__antithesis_instrumentation__.Notify(266727)
		zc = extendZoneCfg(zc, *ext)
	} else {
		__antithesis_instrumentation__.Notify(266728)
	}
	__antithesis_instrumentation__.Notify(266725)
	if ext, ok := r.zoneCfgExtensions.RegionalIn[region]; ok {
		__antithesis_instrumentation__.Notify(266729)
		zc = extendZoneCfg(zc, ext)
	} else {
		__antithesis_instrumentation__.Notify(266730)
	}
	__antithesis_instrumentation__.Notify(266726)
	return zc
}

func extendZoneCfg(zc, ext zonepb.ZoneConfig) zonepb.ZoneConfig {
	__antithesis_instrumentation__.Notify(266731)
	ext.InheritFromParent(&zc)
	return ext
}

func (r *RegionConfig) GlobalTablesInheritDatabaseConstraints() bool {
	__antithesis_instrumentation__.Notify(266732)
	if r.placement == descpb.DataPlacement_RESTRICTED {
		__antithesis_instrumentation__.Notify(266737)

		return false
	} else {
		__antithesis_instrumentation__.Notify(266738)
	}
	__antithesis_instrumentation__.Notify(266733)
	if r.zoneCfgExtensions.Global != nil {
		__antithesis_instrumentation__.Notify(266739)

		return false
	} else {
		__antithesis_instrumentation__.Notify(266740)
	}
	__antithesis_instrumentation__.Notify(266734)
	if r.zoneCfgExtensions.Regional != nil {
		__antithesis_instrumentation__.Notify(266741)

		return false
	} else {
		__antithesis_instrumentation__.Notify(266742)
	}
	__antithesis_instrumentation__.Notify(266735)
	if _, ok := r.zoneCfgExtensions.RegionalIn[r.primaryRegion]; ok {
		__antithesis_instrumentation__.Notify(266743)

		return false
	} else {
		__antithesis_instrumentation__.Notify(266744)
	}
	__antithesis_instrumentation__.Notify(266736)
	return true
}

func (r *RegionConfig) RegionalInTablesInheritDatabaseConstraints(region catpb.RegionName) bool {
	__antithesis_instrumentation__.Notify(266745)
	if _, ok := r.zoneCfgExtensions.RegionalIn[r.primaryRegion]; ok {
		__antithesis_instrumentation__.Notify(266747)

		return r.primaryRegion == region
	} else {
		__antithesis_instrumentation__.Notify(266748)
	}
	__antithesis_instrumentation__.Notify(266746)
	return true
}

type MakeRegionConfigOption func(r *RegionConfig)

func WithTransitioningRegions(transitioningRegions catpb.RegionNames) MakeRegionConfigOption {
	__antithesis_instrumentation__.Notify(266749)
	return func(r *RegionConfig) {
		__antithesis_instrumentation__.Notify(266750)
		r.transitioningRegions = transitioningRegions
	}
}

func MakeRegionConfig(
	regions catpb.RegionNames,
	primaryRegion catpb.RegionName,
	survivalGoal descpb.SurvivalGoal,
	regionEnumID descpb.ID,
	placement descpb.DataPlacement,
	superRegions []descpb.SuperRegion,
	zoneCfgExtensions descpb.ZoneConfigExtensions,
	opts ...MakeRegionConfigOption,
) RegionConfig {
	__antithesis_instrumentation__.Notify(266751)
	ret := RegionConfig{
		regions:           regions,
		primaryRegion:     primaryRegion,
		survivalGoal:      survivalGoal,
		regionEnumID:      regionEnumID,
		placement:         placement,
		superRegions:      superRegions,
		zoneCfgExtensions: zoneCfgExtensions,
	}
	for _, opt := range opts {
		__antithesis_instrumentation__.Notify(266753)
		opt(&ret)
	}
	__antithesis_instrumentation__.Notify(266752)
	return ret
}

func CanSatisfySurvivalGoal(survivalGoal descpb.SurvivalGoal, numRegions int) error {
	__antithesis_instrumentation__.Notify(266754)
	if survivalGoal == descpb.SurvivalGoal_REGION_FAILURE {
		__antithesis_instrumentation__.Notify(266756)
		if numRegions < minNumRegionsForSurviveRegionGoal {
			__antithesis_instrumentation__.Notify(266757)
			return errors.WithHintf(
				pgerror.Newf(
					pgcode.InvalidParameterValue,
					"at least %d regions are required for surviving a region failure",
					minNumRegionsForSurviveRegionGoal,
				),
				"you must add additional regions to the database or "+
					"change the survivability goal",
			)
		} else {
			__antithesis_instrumentation__.Notify(266758)
		}
	} else {
		__antithesis_instrumentation__.Notify(266759)
	}
	__antithesis_instrumentation__.Notify(266755)
	return nil
}

func ValidateRegionConfig(config RegionConfig) error {
	__antithesis_instrumentation__.Notify(266760)
	if config.regionEnumID == descpb.InvalidID {
		__antithesis_instrumentation__.Notify(266768)
		return errors.AssertionFailedf("expected a valid multi-region enum ID to be initialized")
	} else {
		__antithesis_instrumentation__.Notify(266769)
	}
	__antithesis_instrumentation__.Notify(266761)
	if len(config.regions) == 0 {
		__antithesis_instrumentation__.Notify(266770)
		return errors.AssertionFailedf("expected > 0 number of regions in the region config")
	} else {
		__antithesis_instrumentation__.Notify(266771)
	}
	__antithesis_instrumentation__.Notify(266762)
	if config.placement == descpb.DataPlacement_RESTRICTED && func() bool {
		__antithesis_instrumentation__.Notify(266772)
		return config.survivalGoal == descpb.SurvivalGoal_REGION_FAILURE == true
	}() == true {
		__antithesis_instrumentation__.Notify(266773)
		return errors.AssertionFailedf(
			"cannot have a database with restricted placement that is also region survivable")
	} else {
		__antithesis_instrumentation__.Notify(266774)
	}
	__antithesis_instrumentation__.Notify(266763)

	var err error
	ValidateSuperRegions(config.SuperRegions(), config.SurvivalGoal(), config.Regions(), func(validateErr error) {
		__antithesis_instrumentation__.Notify(266775)
		if err == nil {
			__antithesis_instrumentation__.Notify(266776)
			err = validateErr
		} else {
			__antithesis_instrumentation__.Notify(266777)
		}
	})
	__antithesis_instrumentation__.Notify(266764)
	if err != nil {
		__antithesis_instrumentation__.Notify(266778)
		return err
	} else {
		__antithesis_instrumentation__.Notify(266779)
	}
	__antithesis_instrumentation__.Notify(266765)

	ValidateZoneConfigExtensions(config.Regions(), config.ZoneConfigExtensions(), func(validateErr error) {
		__antithesis_instrumentation__.Notify(266780)
		if err == nil {
			__antithesis_instrumentation__.Notify(266781)
			err = validateErr
		} else {
			__antithesis_instrumentation__.Notify(266782)
		}
	})
	__antithesis_instrumentation__.Notify(266766)
	if err != nil {
		__antithesis_instrumentation__.Notify(266783)
		return err
	} else {
		__antithesis_instrumentation__.Notify(266784)
	}
	__antithesis_instrumentation__.Notify(266767)

	return CanSatisfySurvivalGoal(config.survivalGoal, len(config.regions))
}

func ValidateSuperRegions(
	superRegions []descpb.SuperRegion,
	survivalGoal descpb.SurvivalGoal,
	regionNames catpb.RegionNames,
	errorHandler func(error),
) {
	__antithesis_instrumentation__.Notify(266785)
	seenRegions := make(map[catpb.RegionName]struct{})
	superRegionNames := make(map[string]struct{})

	if !sort.SliceIsSorted(superRegions, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(266787)
		return superRegions[i].SuperRegionName < superRegions[j].SuperRegionName
	}) {
		__antithesis_instrumentation__.Notify(266788)
		err := errors.AssertionFailedf("super regions are not in sorted order based on the super region name %v", superRegions)
		errorHandler(err)
	} else {
		__antithesis_instrumentation__.Notify(266789)
	}
	__antithesis_instrumentation__.Notify(266786)

	for _, superRegion := range superRegions {
		__antithesis_instrumentation__.Notify(266790)
		if len(superRegion.Regions) == 0 {
			__antithesis_instrumentation__.Notify(266795)
			err := errors.AssertionFailedf("no regions found within super region %s", superRegion.SuperRegionName)
			errorHandler(err)
		} else {
			__antithesis_instrumentation__.Notify(266796)
		}
		__antithesis_instrumentation__.Notify(266791)

		if err := CanSatisfySurvivalGoal(survivalGoal, len(superRegion.Regions)); err != nil {
			__antithesis_instrumentation__.Notify(266797)
			err := errors.HandleAsAssertionFailure(errors.Wrapf(err, "super region %s only has %d regions", superRegion.SuperRegionName, len(superRegion.Regions)))
			errorHandler(err)
		} else {
			__antithesis_instrumentation__.Notify(266798)
		}
		__antithesis_instrumentation__.Notify(266792)

		_, found := superRegionNames[superRegion.SuperRegionName]
		if found {
			__antithesis_instrumentation__.Notify(266799)
			err := errors.AssertionFailedf("duplicate super regions with name %s found", superRegion.SuperRegionName)
			errorHandler(err)
		} else {
			__antithesis_instrumentation__.Notify(266800)
		}
		__antithesis_instrumentation__.Notify(266793)
		superRegionNames[superRegion.SuperRegionName] = struct{}{}

		if !sort.SliceIsSorted(superRegion.Regions, func(i, j int) bool {
			__antithesis_instrumentation__.Notify(266801)
			return superRegion.Regions[i] < superRegion.Regions[j]
		}) {
			__antithesis_instrumentation__.Notify(266802)
			err := errors.AssertionFailedf("the regions within super region %s were not in a sorted order", superRegion.SuperRegionName)
			errorHandler(err)
		} else {
			__antithesis_instrumentation__.Notify(266803)
		}
		__antithesis_instrumentation__.Notify(266794)

		seenRegionsInCurrentSuperRegion := make(map[catpb.RegionName]struct{})
		for _, region := range superRegion.Regions {
			__antithesis_instrumentation__.Notify(266804)
			_, found := seenRegionsInCurrentSuperRegion[region]
			if found {
				__antithesis_instrumentation__.Notify(266808)
				err := errors.AssertionFailedf("duplicate region %s found in super region %s", region, superRegion.SuperRegionName)
				errorHandler(err)
				continue
			} else {
				__antithesis_instrumentation__.Notify(266809)
			}
			__antithesis_instrumentation__.Notify(266805)
			seenRegionsInCurrentSuperRegion[region] = struct{}{}
			_, found = seenRegions[region]
			if found {
				__antithesis_instrumentation__.Notify(266810)
				err := errors.AssertionFailedf("region %s found in multiple super regions", region)
				errorHandler(err)
				continue
			} else {
				__antithesis_instrumentation__.Notify(266811)
			}
			__antithesis_instrumentation__.Notify(266806)
			seenRegions[region] = struct{}{}

			found = false
			for _, regionName := range regionNames {
				__antithesis_instrumentation__.Notify(266812)
				if region == regionName {
					__antithesis_instrumentation__.Notify(266813)
					found = true
				} else {
					__antithesis_instrumentation__.Notify(266814)
				}
			}
			__antithesis_instrumentation__.Notify(266807)
			if !found {
				__antithesis_instrumentation__.Notify(266815)
				err := errors.Newf("region %s not part of database", region)
				errorHandler(err)
			} else {
				__antithesis_instrumentation__.Notify(266816)
			}
		}
	}
}

func ValidateZoneConfigExtensions(
	regionNames catpb.RegionNames,
	zoneCfgExtensions descpb.ZoneConfigExtensions,
	errorHandler func(error),
) {
	__antithesis_instrumentation__.Notify(266817)

	for regionExt := range zoneCfgExtensions.RegionalIn {
		__antithesis_instrumentation__.Notify(266818)
		found := false
		for _, regionInDB := range regionNames {
			__antithesis_instrumentation__.Notify(266820)
			if regionExt == regionInDB {
				__antithesis_instrumentation__.Notify(266821)
				found = true
				break
			} else {
				__antithesis_instrumentation__.Notify(266822)
			}
		}
		__antithesis_instrumentation__.Notify(266819)
		if !found {
			__antithesis_instrumentation__.Notify(266823)
			errorHandler(errors.AssertionFailedf("region %s has REGIONAL IN "+
				"zone config extension, but is not a region in the database", regionExt))
		} else {
			__antithesis_instrumentation__.Notify(266824)
		}
	}
}

func IsMemberOfSuperRegion(name catpb.RegionName, config RegionConfig) (bool, string) {
	__antithesis_instrumentation__.Notify(266825)
	for _, superRegion := range config.SuperRegions() {
		__antithesis_instrumentation__.Notify(266827)
		for _, region := range superRegion.Regions {
			__antithesis_instrumentation__.Notify(266828)
			if region == name {
				__antithesis_instrumentation__.Notify(266829)
				return true, superRegion.SuperRegionName
			} else {
				__antithesis_instrumentation__.Notify(266830)
			}
		}
	}
	__antithesis_instrumentation__.Notify(266826)

	return false, ""
}

func CanDropRegion(name catpb.RegionName, config RegionConfig) error {
	__antithesis_instrumentation__.Notify(266831)
	isMember, superRegion := IsMemberOfSuperRegion(name, config)
	if isMember {
		__antithesis_instrumentation__.Notify(266833)
		return errors.WithHintf(
			pgerror.Newf(pgcode.DependentObjectsStillExist, "region %s is part of super region %s", name, superRegion),
			"you must first drop super region %s before you can drop the region %s", superRegion, name,
		)
	} else {
		__antithesis_instrumentation__.Notify(266834)
	}
	__antithesis_instrumentation__.Notify(266832)
	return CanSatisfySurvivalGoal(config.survivalGoal, len(config.regions)-1)
}
