package multiregion

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
)

func ValidateTableLocalityConfig(
	desc catalog.TableDescriptor, db catalog.DatabaseDescriptor, vdg catalog.ValidationDescGetter,
) error {
	__antithesis_instrumentation__.Notify(266835)

	lc := desc.GetLocalityConfig()
	if lc == nil {
		__antithesis_instrumentation__.Notify(266846)
		if db.IsMultiRegion() {
			__antithesis_instrumentation__.Notify(266848)
			return pgerror.Newf(
				pgcode.InvalidTableDefinition,
				"database %s is multi-region enabled, but table %s has no locality set",
				db.GetName(),
				desc.GetName(),
			)
		} else {
			__antithesis_instrumentation__.Notify(266849)
		}
		__antithesis_instrumentation__.Notify(266847)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(266850)
	}
	__antithesis_instrumentation__.Notify(266836)

	if !db.IsMultiRegion() {
		__antithesis_instrumentation__.Notify(266851)
		s := tree.NewFmtCtx(tree.FmtSimple)
		var locality string

		if err := FormatTableLocalityConfig(lc, s); err != nil {
			__antithesis_instrumentation__.Notify(266853)
			locality = "INVALID LOCALITY"
		} else {
			__antithesis_instrumentation__.Notify(266854)
		}
		__antithesis_instrumentation__.Notify(266852)
		locality = s.String()
		return pgerror.Newf(
			pgcode.InvalidTableDefinition,
			"database %s is not multi-region enabled, but table %s has locality %s set",
			db.GetName(),
			desc.GetName(),
			locality,
		)
	} else {
		__antithesis_instrumentation__.Notify(266855)
	}
	__antithesis_instrumentation__.Notify(266837)

	regionsEnumID, err := db.MultiRegionEnumID()
	if err != nil {
		__antithesis_instrumentation__.Notify(266856)
		return err
	} else {
		__antithesis_instrumentation__.Notify(266857)
	}
	__antithesis_instrumentation__.Notify(266838)
	regionsEnumDesc, err := vdg.GetTypeDescriptor(regionsEnumID)
	if err != nil {
		__antithesis_instrumentation__.Notify(266858)
		return errors.Wrapf(err, "multi-region enum with ID %d does not exist", regionsEnumID)
	} else {
		__antithesis_instrumentation__.Notify(266859)
	}
	__antithesis_instrumentation__.Notify(266839)
	if regionsEnumDesc.Dropped() {
		__antithesis_instrumentation__.Notify(266860)
		return errors.AssertionFailedf("multi-region enum type %q (%d) is dropped",
			regionsEnumDesc.GetName(), regionsEnumDesc.GetID())
	} else {
		__antithesis_instrumentation__.Notify(266861)
	}
	__antithesis_instrumentation__.Notify(266840)

	if desc.IsSequence() {
		__antithesis_instrumentation__.Notify(266862)
		if !desc.IsLocalityRegionalByTable() {
			__antithesis_instrumentation__.Notify(266863)
			return errors.AssertionFailedf(
				"expected sequence %s to have locality REGIONAL BY TABLE",
				desc.GetName(),
			)
		} else {
			__antithesis_instrumentation__.Notify(266864)
		}
	} else {
		__antithesis_instrumentation__.Notify(266865)
	}
	__antithesis_instrumentation__.Notify(266841)
	if desc.IsView() {
		__antithesis_instrumentation__.Notify(266866)
		if desc.MaterializedView() {
			__antithesis_instrumentation__.Notify(266867)
			if !desc.IsLocalityGlobal() {
				__antithesis_instrumentation__.Notify(266868)
				return errors.AssertionFailedf(
					"expected materialized view %s to have locality GLOBAL",
					desc.GetName(),
				)
			} else {
				__antithesis_instrumentation__.Notify(266869)
			}
		} else {
			__antithesis_instrumentation__.Notify(266870)
			if !desc.IsLocalityRegionalByTable() {
				__antithesis_instrumentation__.Notify(266871)
				return errors.AssertionFailedf(
					"expected view %s to have locality REGIONAL BY TABLE",
					desc.GetName(),
				)
			} else {
				__antithesis_instrumentation__.Notify(266872)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(266873)
	}
	__antithesis_instrumentation__.Notify(266842)

	typeIDs, typeIDsReferencedByColumns, err := desc.GetAllReferencedTypeIDs(db, vdg.GetTypeDescriptor)
	if err != nil {
		__antithesis_instrumentation__.Notify(266874)
		return err
	} else {
		__antithesis_instrumentation__.Notify(266875)
	}
	__antithesis_instrumentation__.Notify(266843)
	regionEnumIDReferenced := false
	for _, typeID := range typeIDs {
		__antithesis_instrumentation__.Notify(266876)
		if typeID == regionsEnumID {
			__antithesis_instrumentation__.Notify(266877)
			regionEnumIDReferenced = true
			break
		} else {
			__antithesis_instrumentation__.Notify(266878)
		}
	}
	__antithesis_instrumentation__.Notify(266844)
	columnTypesTypeIDs := catalog.MakeDescriptorIDSet(typeIDsReferencedByColumns...)
	switch lc := lc.Locality.(type) {
	case *catpb.LocalityConfig_Global_:
		__antithesis_instrumentation__.Notify(266879)
		if regionEnumIDReferenced {
			__antithesis_instrumentation__.Notify(266890)
			if !columnTypesTypeIDs.Contains(regionsEnumID) {
				__antithesis_instrumentation__.Notify(266891)
				return errors.AssertionFailedf(
					"expected no region Enum ID to be referenced by a GLOBAL TABLE: %q"+
						" but found: %d",
					desc.GetName(),
					regionsEnumDesc.GetID(),
				)
			} else {
				__antithesis_instrumentation__.Notify(266892)
			}
		} else {
			__antithesis_instrumentation__.Notify(266893)
		}
	case *catpb.LocalityConfig_RegionalByRow_:
		__antithesis_instrumentation__.Notify(266880)
		if !desc.IsPartitionAllBy() {
			__antithesis_instrumentation__.Notify(266894)
			return errors.AssertionFailedf("expected REGIONAL BY ROW table to have PartitionAllBy set")
		} else {
			__antithesis_instrumentation__.Notify(266895)
		}
		__antithesis_instrumentation__.Notify(266881)

		regions, err := regionsEnumDesc.RegionNames()
		if err != nil {
			__antithesis_instrumentation__.Notify(266896)
			return err
		} else {
			__antithesis_instrumentation__.Notify(266897)
		}
		__antithesis_instrumentation__.Notify(266882)
		regionNames := make(map[catpb.RegionName]struct{}, len(regions))
		for _, region := range regions {
			__antithesis_instrumentation__.Notify(266898)
			regionNames[region] = struct{}{}
		}
		__antithesis_instrumentation__.Notify(266883)
		transitioningRegions, err := regionsEnumDesc.TransitioningRegionNames()
		if err != nil {
			__antithesis_instrumentation__.Notify(266899)
			return err
		} else {
			__antithesis_instrumentation__.Notify(266900)
		}
		__antithesis_instrumentation__.Notify(266884)
		transitioningRegionNames := make(map[catpb.RegionName]struct{}, len(regions))
		for _, region := range transitioningRegions {
			__antithesis_instrumentation__.Notify(266901)
			transitioningRegionNames[region] = struct{}{}
		}
		__antithesis_instrumentation__.Notify(266885)

		part := desc.GetPrimaryIndex().GetPartitioning()
		err = part.ForEachList(func(name string, _ [][]byte, _ catalog.Partitioning) error {
			__antithesis_instrumentation__.Notify(266902)
			regionName := catpb.RegionName(name)

			if _, ok := transitioningRegionNames[regionName]; ok {
				__antithesis_instrumentation__.Notify(266905)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(266906)
			}
			__antithesis_instrumentation__.Notify(266903)

			if _, ok := regionNames[regionName]; !ok {
				__antithesis_instrumentation__.Notify(266907)
				return errors.AssertionFailedf(
					"unknown partition %s on PRIMARY INDEX of table %s",
					name,
					desc.GetName(),
				)
			} else {
				__antithesis_instrumentation__.Notify(266908)
			}
			__antithesis_instrumentation__.Notify(266904)
			delete(regionNames, regionName)
			return nil
		})
		__antithesis_instrumentation__.Notify(266886)
		if err != nil {
			__antithesis_instrumentation__.Notify(266909)
			return err
		} else {
			__antithesis_instrumentation__.Notify(266910)
		}
		__antithesis_instrumentation__.Notify(266887)

		for regionName := range regionNames {
			__antithesis_instrumentation__.Notify(266911)
			return errors.AssertionFailedf(
				"missing partition %s on PRIMARY INDEX of table %s",
				regionName,
				desc.GetName(),
			)
		}

	case *catpb.LocalityConfig_RegionalByTable_:
		__antithesis_instrumentation__.Notify(266888)

		if lc.RegionalByTable.Region != nil {
			__antithesis_instrumentation__.Notify(266912)
			foundRegion := false
			regions, err := regionsEnumDesc.RegionNamesForValidation()
			if err != nil {
				__antithesis_instrumentation__.Notify(266916)
				return err
			} else {
				__antithesis_instrumentation__.Notify(266917)
			}
			__antithesis_instrumentation__.Notify(266913)
			for _, r := range regions {
				__antithesis_instrumentation__.Notify(266918)
				if *lc.RegionalByTable.Region == r {
					__antithesis_instrumentation__.Notify(266919)
					foundRegion = true
					break
				} else {
					__antithesis_instrumentation__.Notify(266920)
				}
			}
			__antithesis_instrumentation__.Notify(266914)
			if !foundRegion {
				__antithesis_instrumentation__.Notify(266921)
				return errors.WithHintf(
					pgerror.Newf(
						pgcode.InvalidTableDefinition,
						`region "%s" has not been added to database "%s"`,
						*lc.RegionalByTable.Region,
						db.DatabaseDesc().Name,
					),
					"available regions: %s",
					strings.Join(regions.ToStrings(), ", "),
				)
			} else {
				__antithesis_instrumentation__.Notify(266922)
			}
			__antithesis_instrumentation__.Notify(266915)
			if !regionEnumIDReferenced {
				__antithesis_instrumentation__.Notify(266923)
				return errors.AssertionFailedf(
					"expected multi-region enum ID %d to be referenced on REGIONAL BY TABLE: %q locality "+
						"config, but did not find it",
					regionsEnumID,
					desc.GetName(),
				)
			} else {
				__antithesis_instrumentation__.Notify(266924)
			}
		} else {
			__antithesis_instrumentation__.Notify(266925)
			if regionEnumIDReferenced {
				__antithesis_instrumentation__.Notify(266926)

				if !columnTypesTypeIDs.Contains(regionsEnumID) {
					__antithesis_instrumentation__.Notify(266927)
					return errors.AssertionFailedf(
						"expected no region Enum ID to be referenced by a REGIONAL BY TABLE: %q homed in the "+
							"primary region, but found: %d",
						desc.GetName(),
						regionsEnumDesc.GetID(),
					)
				} else {
					__antithesis_instrumentation__.Notify(266928)
				}
			} else {
				__antithesis_instrumentation__.Notify(266929)
			}
		}
	default:
		__antithesis_instrumentation__.Notify(266889)
		return pgerror.Newf(
			pgcode.InvalidTableDefinition,
			"unknown locality level: %T",
			lc,
		)
	}
	__antithesis_instrumentation__.Notify(266845)
	return nil
}

func FormatTableLocalityConfig(c *catpb.LocalityConfig, f *tree.FmtCtx) error {
	__antithesis_instrumentation__.Notify(266930)
	switch v := c.Locality.(type) {
	case *catpb.LocalityConfig_Global_:
		__antithesis_instrumentation__.Notify(266932)
		f.WriteString("GLOBAL")
	case *catpb.LocalityConfig_RegionalByTable_:
		__antithesis_instrumentation__.Notify(266933)
		f.WriteString("REGIONAL BY TABLE IN ")
		if v.RegionalByTable.Region != nil {
			__antithesis_instrumentation__.Notify(266936)
			region := tree.Name(*v.RegionalByTable.Region)
			f.FormatNode(&region)
		} else {
			__antithesis_instrumentation__.Notify(266937)
			f.WriteString("PRIMARY REGION")
		}
	case *catpb.LocalityConfig_RegionalByRow_:
		__antithesis_instrumentation__.Notify(266934)
		f.WriteString("REGIONAL BY ROW")
		if v.RegionalByRow.As != nil {
			__antithesis_instrumentation__.Notify(266938)
			f.WriteString(" AS ")
			col := tree.Name(*v.RegionalByRow.As)
			f.FormatNode(&col)
		} else {
			__antithesis_instrumentation__.Notify(266939)
		}
	default:
		__antithesis_instrumentation__.Notify(266935)
		return errors.Newf("unknown locality: %T", v)
	}
	__antithesis_instrumentation__.Notify(266931)
	return nil
}
