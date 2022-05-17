package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/dbdesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/multiregion"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgnotice"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/roleoption"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/errors"
)

type alterDatabaseOwnerNode struct {
	n    *tree.AlterDatabaseOwner
	desc *dbdesc.Mutable
}

func (p *planner) AlterDatabaseOwner(
	ctx context.Context, n *tree.AlterDatabaseOwner,
) (planNode, error) {
	__antithesis_instrumentation__.Notify(242442)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"ALTER DATABASE",
	); err != nil {
		__antithesis_instrumentation__.Notify(242445)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242446)
	}
	__antithesis_instrumentation__.Notify(242443)

	dbDesc, err := p.Descriptors().GetMutableDatabaseByName(ctx, p.txn, string(n.Name),
		tree.DatabaseLookupFlags{Required: true})
	if err != nil {
		__antithesis_instrumentation__.Notify(242447)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242448)
	}
	__antithesis_instrumentation__.Notify(242444)

	return &alterDatabaseOwnerNode{n: n, desc: dbDesc}, nil
}

func (n *alterDatabaseOwnerNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(242449)
	newOwner, err := n.n.Owner.ToSQLUsername(params.p.SessionData(), security.UsernameValidation)
	if err != nil {
		__antithesis_instrumentation__.Notify(242456)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242457)
	}
	__antithesis_instrumentation__.Notify(242450)
	oldOwner := n.desc.GetPrivileges().Owner()

	if err := params.p.checkCanAlterToNewOwner(params.ctx, n.desc, newOwner); err != nil {
		__antithesis_instrumentation__.Notify(242458)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242459)
	}
	__antithesis_instrumentation__.Notify(242451)

	if err := params.p.CheckRoleOption(params.ctx, roleoption.CREATEDB); err != nil {
		__antithesis_instrumentation__.Notify(242460)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242461)
	}
	__antithesis_instrumentation__.Notify(242452)

	if err := params.p.setNewDatabaseOwner(params.ctx, n.desc, newOwner); err != nil {
		__antithesis_instrumentation__.Notify(242462)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242463)
	}
	__antithesis_instrumentation__.Notify(242453)

	if newOwner == oldOwner {
		__antithesis_instrumentation__.Notify(242464)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(242465)
	}
	__antithesis_instrumentation__.Notify(242454)

	if err := params.p.writeNonDropDatabaseChange(
		params.ctx,
		n.desc,
		tree.AsStringWithFQNames(n.n, params.Ann()),
	); err != nil {
		__antithesis_instrumentation__.Notify(242466)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242467)
	}
	__antithesis_instrumentation__.Notify(242455)

	return nil
}

func (p *planner) setNewDatabaseOwner(
	ctx context.Context, desc catalog.MutableDescriptor, newOwner security.SQLUsername,
) error {
	__antithesis_instrumentation__.Notify(242468)
	privs := desc.GetPrivileges()
	privs.SetOwner(newOwner)

	return p.logEvent(ctx,
		desc.GetID(),
		&eventpb.AlterDatabaseOwner{
			DatabaseName: desc.GetName(),
			Owner:        newOwner.Normalized(),
		})
}

func (n *alterDatabaseOwnerNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(242469)
	return false, nil
}
func (n *alterDatabaseOwnerNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(242470)
	return tree.Datums{}
}
func (n *alterDatabaseOwnerNode) Close(context.Context) {
	__antithesis_instrumentation__.Notify(242471)
}

type alterDatabaseAddRegionNode struct {
	n    *tree.AlterDatabaseAddRegion
	desc *dbdesc.Mutable
}

func (p *planner) AlterDatabaseAddRegion(
	ctx context.Context, n *tree.AlterDatabaseAddRegion,
) (planNode, error) {
	__antithesis_instrumentation__.Notify(242472)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"ALTER DATABASE",
	); err != nil {
		__antithesis_instrumentation__.Notify(242479)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242480)
	}
	__antithesis_instrumentation__.Notify(242473)

	dbDesc, err := p.Descriptors().GetMutableDatabaseByName(ctx, p.txn, string(n.Name),
		tree.DatabaseLookupFlags{Required: true},
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(242481)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242482)
	}
	__antithesis_instrumentation__.Notify(242474)

	if !dbDesc.IsMultiRegion() {
		__antithesis_instrumentation__.Notify(242483)
		return nil, errors.WithHintf(
			pgerror.Newf(pgcode.InvalidDatabaseDefinition, "cannot add region %s to database %s",
				n.Region.String(),
				n.Name.String(),
			),
			"you must add a PRIMARY REGION first using ALTER DATABASE %s PRIMARY REGION %s",
			n.Name.String(),
			n.Region.String(),
		)
	} else {
		__antithesis_instrumentation__.Notify(242484)
	}
	__antithesis_instrumentation__.Notify(242475)

	if err := p.checkPrivilegesForMultiRegionOp(ctx, dbDesc); err != nil {
		__antithesis_instrumentation__.Notify(242485)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242486)
	}
	__antithesis_instrumentation__.Notify(242476)

	if err := p.checkNoRegionalByRowChangeUnderway(
		ctx,
		dbDesc,
	); err != nil {
		__antithesis_instrumentation__.Notify(242487)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242488)
	}
	__antithesis_instrumentation__.Notify(242477)

	if err := p.checkPrivilegesForRepartitioningRegionalByRowTables(
		ctx,
		dbDesc,
	); err != nil {
		__antithesis_instrumentation__.Notify(242489)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242490)
	}
	__antithesis_instrumentation__.Notify(242478)

	return &alterDatabaseAddRegionNode{n: n, desc: dbDesc}, nil
}

var GetMultiRegionEnumAddValuePlacementCCL = func(
	execCfg *ExecutorConfig, typeDesc *typedesc.Mutable, region tree.Name,
) (tree.AlterTypeAddValue, error) {
	__antithesis_instrumentation__.Notify(242491)
	return tree.AlterTypeAddValue{}, sqlerrors.NewCCLRequiredError(
		errors.New("adding regions to a multi-region database requires a CCL binary"),
	)
}

func (n *alterDatabaseAddRegionNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(242492)
	if err := params.p.validateZoneConfigForMultiRegionDatabaseWasNotModifiedByUser(
		params.ctx,
		n.desc,
	); err != nil {
		__antithesis_instrumentation__.Notify(242499)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242500)
	}
	__antithesis_instrumentation__.Notify(242493)

	telemetry.Inc(sqltelemetry.AlterDatabaseAddRegionCounter)

	if err := params.p.checkRegionIsCurrentlyActive(
		params.ctx,
		catpb.RegionName(n.n.Region),
	); err != nil {
		__antithesis_instrumentation__.Notify(242501)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242502)
	}
	__antithesis_instrumentation__.Notify(242494)

	typeDesc, err := params.p.Descriptors().GetMutableTypeVersionByID(
		params.ctx,
		params.p.txn,
		n.desc.RegionConfig.RegionEnumID,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(242503)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242504)
	}
	__antithesis_instrumentation__.Notify(242495)

	placement, err := GetMultiRegionEnumAddValuePlacementCCL(
		params.p.ExecCfg(),
		typeDesc,
		n.n.Region,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(242505)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242506)
	}
	__antithesis_instrumentation__.Notify(242496)

	jobDesc := fmt.Sprintf("Adding new region value %q to %q", tree.EnumValue(n.n.Region), tree.RegionEnum)
	if err := params.p.addEnumValue(
		params.ctx,
		typeDesc,
		&placement,
		jobDesc,
	); err != nil {
		__antithesis_instrumentation__.Notify(242507)
		if pgerror.GetPGCode(err) == pgcode.DuplicateObject {
			__antithesis_instrumentation__.Notify(242509)
			if n.n.IfNotExists {
				__antithesis_instrumentation__.Notify(242511)
				params.p.BufferClientNotice(
					params.ctx,
					pgnotice.Newf("region %q already exists; skipping", n.n.Region),
				)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(242512)
			}
			__antithesis_instrumentation__.Notify(242510)
			return pgerror.Newf(
				pgcode.DuplicateObject,
				"region %q already added to database",
				n.n.Region,
			)
		} else {
			__antithesis_instrumentation__.Notify(242513)
		}
		__antithesis_instrumentation__.Notify(242508)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242514)
	}
	__antithesis_instrumentation__.Notify(242497)

	if err := validateDescriptor(params.ctx, params.p, typeDesc); err != nil {
		__antithesis_instrumentation__.Notify(242515)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242516)
	}
	__antithesis_instrumentation__.Notify(242498)

	return params.p.logEvent(params.ctx,
		n.desc.GetID(),
		&eventpb.AlterDatabaseAddRegion{
			DatabaseName: n.desc.GetName(),
			RegionName:   n.n.Region.String(),
		})
}

func (n *alterDatabaseAddRegionNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(242517)
	return false, nil
}
func (n *alterDatabaseAddRegionNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(242518)
	return tree.Datums{}
}
func (n *alterDatabaseAddRegionNode) Close(context.Context) {
	__antithesis_instrumentation__.Notify(242519)
}

type alterDatabaseDropRegionNode struct {
	n                     *tree.AlterDatabaseDropRegion
	desc                  *dbdesc.Mutable
	removingPrimaryRegion bool
	toDrop                []*typedesc.Mutable
}

var allowDropFinalRegion = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.multiregion.drop_primary_region.enabled",
	"allows dropping the PRIMARY REGION of a database if it is the last region",
	true,
).WithPublic()

func (p *planner) AlterDatabaseDropRegion(
	ctx context.Context, n *tree.AlterDatabaseDropRegion,
) (planNode, error) {
	__antithesis_instrumentation__.Notify(242520)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"ALTER DATABASE",
	); err != nil {
		__antithesis_instrumentation__.Notify(242531)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242532)
	}
	__antithesis_instrumentation__.Notify(242521)

	dbDesc, err := p.Descriptors().GetMutableDatabaseByName(ctx, p.txn, string(n.Name),
		tree.DatabaseLookupFlags{Required: true})
	if err != nil {
		__antithesis_instrumentation__.Notify(242533)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242534)
	}
	__antithesis_instrumentation__.Notify(242522)

	if !dbDesc.IsMultiRegion() {
		__antithesis_instrumentation__.Notify(242535)
		if n.IfExists {
			__antithesis_instrumentation__.Notify(242537)
			p.BufferClientNotice(
				ctx,
				pgnotice.Newf("region %q is not defined on the database; skipping", n.Region),
			)
			return &alterDatabaseDropRegionNode{}, nil
		} else {
			__antithesis_instrumentation__.Notify(242538)
		}
		__antithesis_instrumentation__.Notify(242536)
		return nil, pgerror.New(pgcode.InvalidDatabaseDefinition, "database has no regions to drop")
	} else {
		__antithesis_instrumentation__.Notify(242539)
	}
	__antithesis_instrumentation__.Notify(242523)

	if err := p.checkPrivilegesForMultiRegionOp(ctx, dbDesc); err != nil {
		__antithesis_instrumentation__.Notify(242540)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242541)
	}
	__antithesis_instrumentation__.Notify(242524)

	if err := p.checkNoRegionalByRowChangeUnderway(
		ctx,
		dbDesc,
	); err != nil {
		__antithesis_instrumentation__.Notify(242542)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242543)
	}
	__antithesis_instrumentation__.Notify(242525)

	if err := p.checkPrivilegesForRepartitioningRegionalByRowTables(
		ctx,
		dbDesc,
	); err != nil {
		__antithesis_instrumentation__.Notify(242544)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242545)
	}
	__antithesis_instrumentation__.Notify(242526)

	if err := p.validateZoneConfigForMultiRegionDatabaseWasNotModifiedByUser(
		ctx,
		dbDesc,
	); err != nil {
		__antithesis_instrumentation__.Notify(242546)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242547)
	}
	__antithesis_instrumentation__.Notify(242527)

	removingPrimaryRegion := false
	var toDrop []*typedesc.Mutable

	if dbDesc.RegionConfig.PrimaryRegion == catpb.RegionName(n.Region) {
		__antithesis_instrumentation__.Notify(242548)
		removingPrimaryRegion = true

		typeID, err := dbDesc.MultiRegionEnumID()
		if err != nil {
			__antithesis_instrumentation__.Notify(242555)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(242556)
		}
		__antithesis_instrumentation__.Notify(242549)
		typeDesc, err := p.Descriptors().GetMutableTypeVersionByID(ctx, p.txn, typeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(242557)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(242558)
		}
		__antithesis_instrumentation__.Notify(242550)

		regions, err := typeDesc.RegionNamesIncludingTransitioning()
		if err != nil {
			__antithesis_instrumentation__.Notify(242559)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(242560)
		}
		__antithesis_instrumentation__.Notify(242551)
		if len(regions) != 1 {
			__antithesis_instrumentation__.Notify(242561)
			return nil, errors.WithHintf(
				pgerror.Newf(
					pgcode.InvalidDatabaseDefinition,
					"cannot drop region %q",
					dbDesc.RegionConfig.PrimaryRegion,
				),
				"You must designate another region as the primary region using "+
					"ALTER DATABASE %s PRIMARY REGION <region name> or remove all other regions before "+
					"attempting to drop region %q", dbDesc.GetName(), n.Region,
			)
		} else {
			__antithesis_instrumentation__.Notify(242562)
		}
		__antithesis_instrumentation__.Notify(242552)

		if allowDrop := allowDropFinalRegion.Get(&p.execCfg.Settings.SV); !allowDrop {
			__antithesis_instrumentation__.Notify(242563)
			return nil, pgerror.Newf(
				pgcode.InvalidDatabaseDefinition,
				"databases in this cluster must have at least 1 region",
				n.Region,
				DefaultPrimaryRegionClusterSettingName,
			)
		} else {
			__antithesis_instrumentation__.Notify(242564)
		}
		__antithesis_instrumentation__.Notify(242553)

		toDrop = append(toDrop, typeDesc)
		arrayTypeDesc, err := p.Descriptors().GetMutableTypeVersionByID(ctx, p.txn, typeDesc.ArrayTypeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(242565)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(242566)
		}
		__antithesis_instrumentation__.Notify(242554)
		toDrop = append(toDrop, arrayTypeDesc)
		for _, desc := range toDrop {
			__antithesis_instrumentation__.Notify(242567)

			if err := p.canDropTypeDesc(ctx, desc, tree.DropRestrict); err != nil {
				__antithesis_instrumentation__.Notify(242568)
				return nil, errors.Wrapf(
					err, "error removing primary region from database %s", dbDesc.Name)
			} else {
				__antithesis_instrumentation__.Notify(242569)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(242570)
	}
	__antithesis_instrumentation__.Notify(242528)

	regionConfig, err := SynthesizeRegionConfig(ctx, p.txn, dbDesc.ID, p.Descriptors())
	if err != nil {
		__antithesis_instrumentation__.Notify(242571)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242572)
	}
	__antithesis_instrumentation__.Notify(242529)
	if err := multiregion.CanDropRegion(catpb.RegionName(n.Region), regionConfig); err != nil {
		__antithesis_instrumentation__.Notify(242573)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242574)
	}
	__antithesis_instrumentation__.Notify(242530)

	return &alterDatabaseDropRegionNode{
		n,
		dbDesc,
		removingPrimaryRegion,
		toDrop,
	}, nil
}

func (p *planner) checkPrivilegesForMultiRegionOp(
	ctx context.Context, desc catalog.Descriptor,
) error {
	__antithesis_instrumentation__.Notify(242575)
	hasAdminRole, err := p.HasAdminRole(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(242578)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242579)
	}
	__antithesis_instrumentation__.Notify(242576)
	if !hasAdminRole {
		__antithesis_instrumentation__.Notify(242580)

		err := p.CheckPrivilege(ctx, desc, privilege.CREATE)

		if pgerror.GetPGCode(err) == pgcode.InsufficientPrivilege {
			__antithesis_instrumentation__.Notify(242582)
			return pgerror.Newf(pgcode.InsufficientPrivilege,
				"user %s must be owner of %s or have %s privilege on %s %s",
				p.SessionData().User(),
				desc.GetName(),
				privilege.CREATE,
				desc.DescriptorType(),
				desc.GetName(),
			)
		} else {
			__antithesis_instrumentation__.Notify(242583)
		}
		__antithesis_instrumentation__.Notify(242581)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242584)
	}
	__antithesis_instrumentation__.Notify(242577)
	return nil
}

func (p *planner) checkPrivilegesForRepartitioningRegionalByRowTables(
	ctx context.Context, dbDesc catalog.DatabaseDescriptor,
) error {
	__antithesis_instrumentation__.Notify(242585)
	return p.forEachMutableTableInDatabase(ctx, dbDesc,
		func(ctx context.Context, scName string, tbDesc *tabledesc.Mutable) error {
			__antithesis_instrumentation__.Notify(242586)
			if tbDesc.IsLocalityRegionalByRow() {
				__antithesis_instrumentation__.Notify(242588)
				err := p.checkPrivilegesForMultiRegionOp(ctx, tbDesc)

				if pgerror.GetPGCode(err) == pgcode.InsufficientPrivilege {
					__antithesis_instrumentation__.Notify(242590)
					return errors.Wrapf(err,
						"cannot repartition regional by row table",
					)
				} else {
					__antithesis_instrumentation__.Notify(242591)
				}
				__antithesis_instrumentation__.Notify(242589)
				if err != nil {
					__antithesis_instrumentation__.Notify(242592)
					return err
				} else {
					__antithesis_instrumentation__.Notify(242593)
				}
			} else {
				__antithesis_instrumentation__.Notify(242594)
			}
			__antithesis_instrumentation__.Notify(242587)
			return nil
		})
}

func removeLocalityConfigFromAllTablesInDB(
	ctx context.Context, p *planner, desc catalog.DatabaseDescriptor,
) error {
	__antithesis_instrumentation__.Notify(242595)
	if !desc.IsMultiRegion() {
		__antithesis_instrumentation__.Notify(242598)
		return errors.AssertionFailedf(
			"cannot remove locality configs from tables in non multi-region database with ID %d",
			desc.GetID(),
		)
	} else {
		__antithesis_instrumentation__.Notify(242599)
	}
	__antithesis_instrumentation__.Notify(242596)
	b := p.Txn().NewBatch()
	if err := p.forEachMutableTableInDatabase(
		ctx,
		desc,
		func(ctx context.Context, scName string, tbDesc *tabledesc.Mutable) error {
			__antithesis_instrumentation__.Notify(242600)

			if err := p.checkPrivilegesForMultiRegionOp(ctx, tbDesc); err != nil {
				__antithesis_instrumentation__.Notify(242603)
				return err
			} else {
				__antithesis_instrumentation__.Notify(242604)
			}
			__antithesis_instrumentation__.Notify(242601)

			switch t := tbDesc.LocalityConfig.Locality.(type) {
			case *catpb.LocalityConfig_Global_:
				__antithesis_instrumentation__.Notify(242605)
				if err := ApplyZoneConfigForMultiRegionTable(
					ctx,
					p.txn,
					p.ExecCfg(),
					multiregion.RegionConfig{},
					tbDesc,
					applyZoneConfigForMultiRegionTableOptionRemoveGlobalZoneConfig,
				); err != nil {
					__antithesis_instrumentation__.Notify(242609)
					return err
				} else {
					__antithesis_instrumentation__.Notify(242610)
				}
			case *catpb.LocalityConfig_RegionalByTable_:
				__antithesis_instrumentation__.Notify(242606)
				if t.RegionalByTable.Region != nil {
					__antithesis_instrumentation__.Notify(242611)

					return errors.AssertionFailedf(
						"unexpected REGIONAL BY TABLE IN <region> on table %s during DROP REGION",
						tbDesc.Name,
					)
				} else {
					__antithesis_instrumentation__.Notify(242612)
				}
			case *catpb.LocalityConfig_RegionalByRow_:
				__antithesis_instrumentation__.Notify(242607)

				return errors.AssertionFailedf(
					"unexpected REGIONAL BY ROW on table %s during DROP REGION",
					tbDesc.Name,
				)
			default:
				__antithesis_instrumentation__.Notify(242608)
				return errors.AssertionFailedf(
					"unexpected locality %T on table %s during DROP REGION",
					t,
					tbDesc.Name,
				)
			}
			__antithesis_instrumentation__.Notify(242602)
			tbDesc.LocalityConfig = nil
			return p.writeSchemaChangeToBatch(ctx, tbDesc, b)
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(242613)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242614)
	}
	__antithesis_instrumentation__.Notify(242597)
	return p.Txn().Run(ctx, b)
}

func (n *alterDatabaseDropRegionNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(242615)
	if n.n == nil {
		__antithesis_instrumentation__.Notify(242620)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(242621)
	}
	__antithesis_instrumentation__.Notify(242616)
	typeDesc, err := params.p.Descriptors().GetMutableTypeVersionByID(
		params.ctx,
		params.p.txn,
		n.desc.RegionConfig.RegionEnumID,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(242622)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242623)
	}
	__antithesis_instrumentation__.Notify(242617)

	if n.removingPrimaryRegion {
		__antithesis_instrumentation__.Notify(242624)
		telemetry.Inc(sqltelemetry.AlterDatabaseDropPrimaryRegionCounter)
		for _, desc := range n.toDrop {
			__antithesis_instrumentation__.Notify(242627)
			jobDesc := fmt.Sprintf("drop multi-region enum with ID %d", desc.ID)
			err := params.p.dropTypeImpl(params.ctx, desc, jobDesc, true)
			if err != nil {
				__antithesis_instrumentation__.Notify(242628)
				return err
			} else {
				__antithesis_instrumentation__.Notify(242629)
			}
		}
		__antithesis_instrumentation__.Notify(242625)

		err = removeLocalityConfigFromAllTablesInDB(params.ctx, params.p, n.desc)
		if err != nil {
			__antithesis_instrumentation__.Notify(242630)
			return errors.Wrap(err, "error removing locality configs from tables")
		} else {
			__antithesis_instrumentation__.Notify(242631)
		}
		__antithesis_instrumentation__.Notify(242626)

		n.desc.UnsetMultiRegionConfig()
		if err := discardMultiRegionFieldsForDatabaseZoneConfig(
			params.ctx,
			n.desc.ID,
			params.p.txn,
			params.p.execCfg,
		); err != nil {
			__antithesis_instrumentation__.Notify(242632)
			return err
		} else {
			__antithesis_instrumentation__.Notify(242633)
		}
	} else {
		__antithesis_instrumentation__.Notify(242634)
		telemetry.Inc(sqltelemetry.AlterDatabaseDropRegionCounter)

		if err := params.p.dropEnumValue(params.ctx, typeDesc, tree.EnumValue(n.n.Region)); err != nil {
			__antithesis_instrumentation__.Notify(242635)
			if pgerror.GetPGCode(err) == pgcode.UndefinedObject {
				__antithesis_instrumentation__.Notify(242637)
				if n.n.IfExists {
					__antithesis_instrumentation__.Notify(242639)
					params.p.BufferClientNotice(
						params.ctx,
						pgnotice.Newf("region %q is not defined on the database; skipping", n.n.Region),
					)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(242640)
				}
				__antithesis_instrumentation__.Notify(242638)
				return pgerror.Newf(
					pgcode.UndefinedObject,
					"region %q has not been added to the database",
					n.n.Region,
				)
			} else {
				__antithesis_instrumentation__.Notify(242641)
			}
			__antithesis_instrumentation__.Notify(242636)
			return err
		} else {
			__antithesis_instrumentation__.Notify(242642)
		}
	}
	__antithesis_instrumentation__.Notify(242618)

	if err := params.p.writeNonDropDatabaseChange(
		params.ctx,
		n.desc,
		tree.AsStringWithFQNames(n.n, params.Ann()),
	); err != nil {
		__antithesis_instrumentation__.Notify(242643)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242644)
	}
	__antithesis_instrumentation__.Notify(242619)

	return params.p.logEvent(params.ctx,
		n.desc.GetID(),
		&eventpb.AlterDatabaseDropRegion{
			DatabaseName: n.desc.GetName(),
			RegionName:   n.n.Region.String(),
		})
}

func (n *alterDatabaseDropRegionNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(242645)
	return false, nil
}
func (n *alterDatabaseDropRegionNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(242646)
	return tree.Datums{}
}
func (n *alterDatabaseDropRegionNode) Close(context.Context) {
	__antithesis_instrumentation__.Notify(242647)
}

type alterDatabasePrimaryRegionNode struct {
	n    *tree.AlterDatabasePrimaryRegion
	desc *dbdesc.Mutable
}

func (p *planner) AlterDatabasePrimaryRegion(
	ctx context.Context, n *tree.AlterDatabasePrimaryRegion,
) (planNode, error) {
	__antithesis_instrumentation__.Notify(242648)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"ALTER DATABASE",
	); err != nil {
		__antithesis_instrumentation__.Notify(242652)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242653)
	}
	__antithesis_instrumentation__.Notify(242649)

	dbDesc, err := p.Descriptors().GetMutableDatabaseByName(ctx, p.txn, string(n.Name),
		tree.DatabaseLookupFlags{Required: true},
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(242654)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242655)
	}
	__antithesis_instrumentation__.Notify(242650)

	if err := p.checkPrivilegesForMultiRegionOp(ctx, dbDesc); err != nil {
		__antithesis_instrumentation__.Notify(242656)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242657)
	}
	__antithesis_instrumentation__.Notify(242651)

	return &alterDatabasePrimaryRegionNode{n: n, desc: dbDesc}, nil
}

func (n *alterDatabasePrimaryRegionNode) switchPrimaryRegion(params runParams) error {
	__antithesis_instrumentation__.Notify(242658)
	telemetry.Inc(sqltelemetry.SwitchPrimaryRegionCounter)

	prevRegionConfig, err := SynthesizeRegionConfig(params.ctx, params.p.txn, n.desc.ID, params.p.Descriptors())
	if err != nil {
		__antithesis_instrumentation__.Notify(242672)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242673)
	}
	__antithesis_instrumentation__.Notify(242659)
	found := false
	for _, r := range prevRegionConfig.Regions() {
		__antithesis_instrumentation__.Notify(242674)
		if r == catpb.RegionName(n.n.PrimaryRegion) {
			__antithesis_instrumentation__.Notify(242675)
			found = true
			break
		} else {
			__antithesis_instrumentation__.Notify(242676)
		}
	}
	__antithesis_instrumentation__.Notify(242660)

	if !found {
		__antithesis_instrumentation__.Notify(242677)
		return errors.WithHintf(
			pgerror.Newf(pgcode.InvalidName,
				"region %s has not been added to the database",
				n.n.PrimaryRegion.String(),
			),
			"you must add the region to the database before setting it as primary region, using "+
				"ALTER DATABASE %s ADD REGION %s",
			n.n.Name.String(),
			n.n.PrimaryRegion.String(),
		)
	} else {
		__antithesis_instrumentation__.Notify(242678)
	}
	__antithesis_instrumentation__.Notify(242661)

	typeDesc, err := params.p.Descriptors().GetMutableTypeVersionByID(
		params.ctx,
		params.p.txn,
		n.desc.RegionConfig.RegionEnumID)
	if err != nil {
		__antithesis_instrumentation__.Notify(242679)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242680)
	}
	__antithesis_instrumentation__.Notify(242662)

	oldPrimaryRegion := n.desc.RegionConfig.PrimaryRegion

	n.desc.RegionConfig.PrimaryRegion = catpb.RegionName(n.n.PrimaryRegion)
	if err := params.p.writeNonDropDatabaseChange(
		params.ctx,
		n.desc,
		tree.AsStringWithFQNames(n.n, params.Ann()),
	); err != nil {
		__antithesis_instrumentation__.Notify(242681)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242682)
	}
	__antithesis_instrumentation__.Notify(242663)

	typeDesc.RegionConfig.PrimaryRegion = catpb.RegionName(n.n.PrimaryRegion)
	if err := params.p.writeTypeDesc(params.ctx, typeDesc); err != nil {
		__antithesis_instrumentation__.Notify(242683)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242684)
	}
	__antithesis_instrumentation__.Notify(242664)

	updatedRegionConfig, err := SynthesizeRegionConfig(
		params.ctx, params.p.txn, n.desc.ID, params.p.Descriptors(),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(242685)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242686)
	}
	__antithesis_instrumentation__.Notify(242665)

	if err := ApplyZoneConfigFromDatabaseRegionConfig(
		params.ctx,
		n.desc.ID,
		updatedRegionConfig,
		params.p.txn,
		params.p.execCfg,
	); err != nil {
		__antithesis_instrumentation__.Notify(242687)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242688)
	}
	__antithesis_instrumentation__.Notify(242666)

	isNewPrimaryRegionMemberOfASuperRegion, superRegionOfNewPrimaryRegion := multiregion.IsMemberOfSuperRegion(catpb.RegionName(n.n.PrimaryRegion), updatedRegionConfig)
	isOldPrimaryRegionMemberOfASuperRegion, superRegionOfOldPrimaryRegion := multiregion.IsMemberOfSuperRegion(oldPrimaryRegion, updatedRegionConfig)

	if (isNewPrimaryRegionMemberOfASuperRegion || func() bool {
		__antithesis_instrumentation__.Notify(242689)
		return isOldPrimaryRegionMemberOfASuperRegion == true
	}() == true) && func() bool {
		__antithesis_instrumentation__.Notify(242690)
		return superRegionOfNewPrimaryRegion != superRegionOfOldPrimaryRegion == true
	}() == true {
		__antithesis_instrumentation__.Notify(242691)
		newSuperRegionMsg := ""
		if isNewPrimaryRegionMemberOfASuperRegion {
			__antithesis_instrumentation__.Notify(242694)
			newSuperRegionMsg = fmt.Sprintf("\nthe new primary region %s is a member of super region %s", n.n.PrimaryRegion, superRegionOfNewPrimaryRegion)
		} else {
			__antithesis_instrumentation__.Notify(242695)
		}
		__antithesis_instrumentation__.Notify(242692)

		oldSuperRegionMsg := ""
		if isOldPrimaryRegionMemberOfASuperRegion {
			__antithesis_instrumentation__.Notify(242696)
			oldSuperRegionMsg = fmt.Sprintf("\nthe old primary region %s is a member of super region %s", oldPrimaryRegion, superRegionOfOldPrimaryRegion)
		} else {
			__antithesis_instrumentation__.Notify(242697)
		}
		__antithesis_instrumentation__.Notify(242693)

		if !params.p.SessionData().OverrideAlterPrimaryRegionInSuperRegion {
			__antithesis_instrumentation__.Notify(242698)
			return errors.WithTelemetry(
				errors.WithHint(errors.Newf("moving the primary region into or "+
					"out of a super region could have undesirable data placement "+
					"implications as all regional tables in the primary region will now moved.\n%s%s", newSuperRegionMsg, oldSuperRegionMsg),
					"You can enable this operation by running `SET alter_primary_region_super_region_override = 'on'`."),
				"sql.schema.alter_primary_region_in_super_region",
			)
		} else {
			__antithesis_instrumentation__.Notify(242699)
		}
	} else {
		__antithesis_instrumentation__.Notify(242700)
	}
	__antithesis_instrumentation__.Notify(242667)

	if isNewPrimaryRegionMemberOfASuperRegion {
		__antithesis_instrumentation__.Notify(242701)
		params.p.BufferClientNotice(params.ctx, pgnotice.Newf("all REGIONAL BY TABLE tables without a region explicitly set are now part of the super region %s", superRegionOfNewPrimaryRegion))
	} else {
		__antithesis_instrumentation__.Notify(242702)
	}
	__antithesis_instrumentation__.Notify(242668)

	if isOldPrimaryRegionMemberOfASuperRegion {
		__antithesis_instrumentation__.Notify(242703)
		params.p.BufferClientNotice(params.ctx, pgnotice.Newf("all REGIONAL BY TABLE tables without a region explicitly set are no longer part of the super region %s", superRegionOfOldPrimaryRegion))
	} else {
		__antithesis_instrumentation__.Notify(242704)
	}
	__antithesis_instrumentation__.Notify(242669)

	opts := WithOnlyGlobalTables
	if isNewPrimaryRegionMemberOfASuperRegion || func() bool {
		__antithesis_instrumentation__.Notify(242705)
		return isOldPrimaryRegionMemberOfASuperRegion == true
	}() == true {
		__antithesis_instrumentation__.Notify(242706)
		opts = WithOnlyRegionalTablesAndGlobalTables
	} else {
		__antithesis_instrumentation__.Notify(242707)
	}
	__antithesis_instrumentation__.Notify(242670)

	if updatedRegionConfig.IsPlacementRestricted() || func() bool {
		__antithesis_instrumentation__.Notify(242708)
		return isNewPrimaryRegionMemberOfASuperRegion == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(242709)
		return isOldPrimaryRegionMemberOfASuperRegion == true
	}() == true {
		__antithesis_instrumentation__.Notify(242710)
		if err := params.p.updateZoneConfigsForTables(
			params.ctx,
			n.desc,
			opts,
		); err != nil {
			__antithesis_instrumentation__.Notify(242711)
			return err
		} else {
			__antithesis_instrumentation__.Notify(242712)
		}
	} else {
		__antithesis_instrumentation__.Notify(242713)
	}
	__antithesis_instrumentation__.Notify(242671)

	return nil
}

func addDefaultLocalityConfigToAllTables(
	ctx context.Context, p *planner, dbDesc catalog.DatabaseDescriptor, regionEnumID descpb.ID,
) error {
	__antithesis_instrumentation__.Notify(242714)
	if !dbDesc.IsMultiRegion() {
		__antithesis_instrumentation__.Notify(242717)
		return errors.AssertionFailedf(
			"cannot add locality config to tables in non multi-region database with ID %d",
			dbDesc.GetID(),
		)
	} else {
		__antithesis_instrumentation__.Notify(242718)
	}
	__antithesis_instrumentation__.Notify(242715)
	b := p.Txn().NewBatch()
	if err := p.forEachMutableTableInDatabase(
		ctx,
		dbDesc,
		func(ctx context.Context, scName string, tbDesc *tabledesc.Mutable) error {
			__antithesis_instrumentation__.Notify(242719)
			if err := p.checkPrivilegesForMultiRegionOp(ctx, tbDesc); err != nil {
				__antithesis_instrumentation__.Notify(242724)
				return err
			} else {
				__antithesis_instrumentation__.Notify(242725)
			}
			__antithesis_instrumentation__.Notify(242720)

			if err := checkCanConvertTableToMultiRegion(dbDesc, tbDesc); err != nil {
				__antithesis_instrumentation__.Notify(242726)
				return err
			} else {
				__antithesis_instrumentation__.Notify(242727)
			}
			__antithesis_instrumentation__.Notify(242721)

			if tbDesc.MaterializedView() {
				__antithesis_instrumentation__.Notify(242728)
				if err := p.alterTableDescLocalityToGlobal(
					ctx, tbDesc, regionEnumID,
				); err != nil {
					__antithesis_instrumentation__.Notify(242729)
					return err
				} else {
					__antithesis_instrumentation__.Notify(242730)
				}
			} else {
				__antithesis_instrumentation__.Notify(242731)
				if err := p.alterTableDescLocalityToRegionalByTable(
					ctx, tree.PrimaryRegionNotSpecifiedName, tbDesc, regionEnumID,
				); err != nil {
					__antithesis_instrumentation__.Notify(242732)
					return err
				} else {
					__antithesis_instrumentation__.Notify(242733)
				}
			}
			__antithesis_instrumentation__.Notify(242722)
			if err := p.writeSchemaChangeToBatch(ctx, tbDesc, b); err != nil {
				__antithesis_instrumentation__.Notify(242734)
				return err
			} else {
				__antithesis_instrumentation__.Notify(242735)
			}
			__antithesis_instrumentation__.Notify(242723)
			return nil
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(242736)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242737)
	}
	__antithesis_instrumentation__.Notify(242716)
	return p.Txn().Run(ctx, b)
}

func checkCanConvertTableToMultiRegion(
	dbDesc catalog.DatabaseDescriptor, tableDesc catalog.TableDescriptor,
) error {
	__antithesis_instrumentation__.Notify(242738)
	if tableDesc.GetPrimaryIndex().GetPartitioning().NumColumns() > 0 {
		__antithesis_instrumentation__.Notify(242741)
		return errors.WithDetailf(
			pgerror.Newf(
				pgcode.ObjectNotInPrerequisiteState,
				"cannot convert database %s to a multi-region database",
				dbDesc.GetName(),
			),
			"cannot convert table %s to a multi-region table as it is partitioned",
			tableDesc.GetName(),
		)
	} else {
		__antithesis_instrumentation__.Notify(242742)
	}
	__antithesis_instrumentation__.Notify(242739)
	for _, idx := range tableDesc.AllIndexes() {
		__antithesis_instrumentation__.Notify(242743)
		if idx.GetPartitioning().NumColumns() > 0 {
			__antithesis_instrumentation__.Notify(242744)
			return errors.WithDetailf(
				pgerror.Newf(
					pgcode.ObjectNotInPrerequisiteState,
					"cannot convert database %s to a multi-region database",
					dbDesc.GetName(),
				),
				"cannot convert table %s to a multi-region table as it has index/constraint %s with partitioning",
				tableDesc.GetName(),
				idx.GetName(),
			)
		} else {
			__antithesis_instrumentation__.Notify(242745)
		}
	}
	__antithesis_instrumentation__.Notify(242740)
	return nil
}

func (n *alterDatabasePrimaryRegionNode) setInitialPrimaryRegion(params runParams) error {
	__antithesis_instrumentation__.Notify(242746)
	telemetry.Inc(sqltelemetry.SetInitialPrimaryRegionCounter)

	regionConfig, err := params.p.maybeInitializeMultiRegionMetadata(
		params.ctx,
		tree.SurvivalGoalDefault,
		n.n.PrimaryRegion,
		[]tree.Name{n.n.PrimaryRegion},
		tree.DataPlacementUnspecified,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(242753)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242754)
	}
	__antithesis_instrumentation__.Notify(242747)

	if err := params.p.validateAllMultiRegionZoneConfigsInDatabase(
		params.ctx,
		n.desc,
		&zoneConfigForMultiRegionValidatorSetInitialRegion{},
	); err != nil {
		__antithesis_instrumentation__.Notify(242755)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242756)
	}
	__antithesis_instrumentation__.Notify(242748)

	if err := n.desc.SetInitialMultiRegionConfig(regionConfig); err != nil {
		__antithesis_instrumentation__.Notify(242757)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242758)
	}
	__antithesis_instrumentation__.Notify(242749)

	if err := addDefaultLocalityConfigToAllTables(
		params.ctx,
		params.p,
		n.desc,
		regionConfig.RegionEnumID(),
	); err != nil {
		__antithesis_instrumentation__.Notify(242759)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242760)
	}
	__antithesis_instrumentation__.Notify(242750)

	if err := params.p.writeNonDropDatabaseChange(
		params.ctx,
		n.desc,
		tree.AsStringWithFQNames(n.n, params.Ann()),
	); err != nil {
		__antithesis_instrumentation__.Notify(242761)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242762)
	}
	__antithesis_instrumentation__.Notify(242751)

	err = params.p.maybeInitializeMultiRegionDatabase(params.ctx, n.desc, regionConfig)
	if err != nil {
		__antithesis_instrumentation__.Notify(242763)

		if pgerror.GetPGCode(err) == pgcode.DuplicateObject {
			__antithesis_instrumentation__.Notify(242765)
			return errors.WithHint(
				errors.WithDetail(err,
					"multi-region databases employ an internal enum called crdb_internal_region to "+
						"manage regions which conflicts with the existing object"),
				`object "crdb_internal_regions" must be renamed or dropped before adding the primary region`)
		} else {
			__antithesis_instrumentation__.Notify(242766)
		}
		__antithesis_instrumentation__.Notify(242764)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242767)
	}
	__antithesis_instrumentation__.Notify(242752)
	return nil
}

func (n *alterDatabasePrimaryRegionNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(242768)

	if n.desc.GetID() == keys.SystemDatabaseID {
		__antithesis_instrumentation__.Notify(242771)
		return pgerror.Newf(
			pgcode.FeatureNotSupported,
			"adding a primary region to the system database is not supported",
		)
	} else {
		__antithesis_instrumentation__.Notify(242772)
	}
	__antithesis_instrumentation__.Notify(242769)

	if !n.desc.IsMultiRegion() {
		__antithesis_instrumentation__.Notify(242773)

		err := n.setInitialPrimaryRegion(params)
		if err != nil {
			__antithesis_instrumentation__.Notify(242774)
			return err
		} else {
			__antithesis_instrumentation__.Notify(242775)
		}
	} else {
		__antithesis_instrumentation__.Notify(242776)
		if err := params.p.validateZoneConfigForMultiRegionDatabaseWasNotModifiedByUser(
			params.ctx,
			n.desc,
		); err != nil {
			__antithesis_instrumentation__.Notify(242778)
			return err
		} else {
			__antithesis_instrumentation__.Notify(242779)
		}
		__antithesis_instrumentation__.Notify(242777)

		err := n.switchPrimaryRegion(params)
		if err != nil {
			__antithesis_instrumentation__.Notify(242780)
			return err
		} else {
			__antithesis_instrumentation__.Notify(242781)
		}
	}
	__antithesis_instrumentation__.Notify(242770)

	return params.p.logEvent(params.ctx,
		n.desc.GetID(),
		&eventpb.AlterDatabasePrimaryRegion{
			DatabaseName:      n.desc.GetName(),
			PrimaryRegionName: n.n.PrimaryRegion.String(),
		})
}

func (n *alterDatabasePrimaryRegionNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(242782)
	return false, nil
}
func (n *alterDatabasePrimaryRegionNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(242783)
	return tree.Datums{}
}
func (n *alterDatabasePrimaryRegionNode) Close(context.Context) {
	__antithesis_instrumentation__.Notify(242784)
}
func (n *alterDatabasePrimaryRegionNode) ReadingOwnWrites() {
	__antithesis_instrumentation__.Notify(242785)
}

type alterDatabaseSurvivalGoalNode struct {
	n    *tree.AlterDatabaseSurvivalGoal
	desc *dbdesc.Mutable
}

func (p *planner) AlterDatabaseSurvivalGoal(
	ctx context.Context, n *tree.AlterDatabaseSurvivalGoal,
) (planNode, error) {
	__antithesis_instrumentation__.Notify(242786)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"ALTER DATABASE",
	); err != nil {
		__antithesis_instrumentation__.Notify(242790)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242791)
	}
	__antithesis_instrumentation__.Notify(242787)

	dbDesc, err := p.Descriptors().GetMutableDatabaseByName(ctx, p.txn, string(n.Name),
		tree.DatabaseLookupFlags{Required: true},
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(242792)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242793)
	}
	__antithesis_instrumentation__.Notify(242788)
	if err := p.checkPrivilegesForMultiRegionOp(ctx, dbDesc); err != nil {
		__antithesis_instrumentation__.Notify(242794)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242795)
	}
	__antithesis_instrumentation__.Notify(242789)

	return &alterDatabaseSurvivalGoalNode{n: n, desc: dbDesc}, nil
}

func (n *alterDatabaseSurvivalGoalNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(242796)

	if !n.desc.IsMultiRegion() {
		__antithesis_instrumentation__.Notify(242806)
		return errors.WithHintf(
			pgerror.New(pgcode.InvalidName,
				"database must have associated regions before a survival goal can be set",
			),
			"you must first add a primary region to the database using "+
				"ALTER DATABASE %s PRIMARY REGION <region_name>",
			n.n.Name.String(),
		)
	} else {
		__antithesis_instrumentation__.Notify(242807)
	}
	__antithesis_instrumentation__.Notify(242797)

	if n.n.SurvivalGoal == tree.SurvivalGoalRegionFailure && func() bool {
		__antithesis_instrumentation__.Notify(242808)
		return n.desc.RegionConfig.Placement == descpb.DataPlacement_RESTRICTED == true
	}() == true {
		__antithesis_instrumentation__.Notify(242809)
		return errors.WithDetailf(
			pgerror.New(pgcode.InvalidParameterValue,
				"a region-survivable database cannot also have a restricted placement policy"),
			"PLACEMENT RESTRICTED may only be used with SURVIVE ZONE FAILURE",
		)
	} else {
		__antithesis_instrumentation__.Notify(242810)
	}
	__antithesis_instrumentation__.Notify(242798)

	if n.n.SurvivalGoal == tree.SurvivalGoalRegionFailure {
		__antithesis_instrumentation__.Notify(242811)
		superRegions, err := params.p.getSuperRegionsForDatabase(params.ctx, n.desc)
		if err != nil {
			__antithesis_instrumentation__.Notify(242813)
			return err
		} else {
			__antithesis_instrumentation__.Notify(242814)
		}
		__antithesis_instrumentation__.Notify(242812)
		for _, sr := range superRegions {
			__antithesis_instrumentation__.Notify(242815)
			if err := multiregion.CanSatisfySurvivalGoal(descpb.SurvivalGoal_REGION_FAILURE, len(sr.Regions)); err != nil {
				__antithesis_instrumentation__.Notify(242816)
				return errors.Wrapf(err, "super region %s only has %d region(s)", sr.SuperRegionName, len(sr.Regions))
			} else {
				__antithesis_instrumentation__.Notify(242817)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(242818)
	}
	__antithesis_instrumentation__.Notify(242799)

	if err := params.p.validateZoneConfigForMultiRegionDatabaseWasNotModifiedByUser(
		params.ctx,
		n.desc,
	); err != nil {
		__antithesis_instrumentation__.Notify(242819)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242820)
	}
	__antithesis_instrumentation__.Notify(242800)

	telemetry.Inc(
		sqltelemetry.AlterDatabaseSurvivalGoalCounter(
			n.n.SurvivalGoal.TelemetryName(),
		),
	)

	survivalGoal, err := TranslateSurvivalGoal(n.n.SurvivalGoal)
	if err != nil {
		__antithesis_instrumentation__.Notify(242821)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242822)
	}
	__antithesis_instrumentation__.Notify(242801)
	n.desc.RegionConfig.SurvivalGoal = survivalGoal

	if err := params.p.writeNonDropDatabaseChange(
		params.ctx,
		n.desc,
		tree.AsStringWithFQNames(n.n, params.Ann()),
	); err != nil {
		__antithesis_instrumentation__.Notify(242823)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242824)
	}
	__antithesis_instrumentation__.Notify(242802)

	regionConfig, err := SynthesizeRegionConfig(params.ctx, params.p.txn, n.desc.ID, params.p.Descriptors())
	if err != nil {
		__antithesis_instrumentation__.Notify(242825)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242826)
	}
	__antithesis_instrumentation__.Notify(242803)

	if err := ApplyZoneConfigFromDatabaseRegionConfig(
		params.ctx,
		n.desc.ID,
		regionConfig,
		params.p.txn,
		params.p.execCfg,
	); err != nil {
		__antithesis_instrumentation__.Notify(242827)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242828)
	}
	__antithesis_instrumentation__.Notify(242804)

	if err := params.p.updateZoneConfigsForTables(params.ctx, n.desc); err != nil {
		__antithesis_instrumentation__.Notify(242829)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242830)
	}
	__antithesis_instrumentation__.Notify(242805)

	return params.p.logEvent(params.ctx,
		n.desc.GetID(),
		&eventpb.AlterDatabaseSurvivalGoal{
			DatabaseName: n.desc.GetName(),
			SurvivalGoal: survivalGoal.String(),
		},
	)
}

func (n *alterDatabaseSurvivalGoalNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(242831)
	return false, nil
}
func (n *alterDatabaseSurvivalGoalNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(242832)
	return tree.Datums{}
}
func (n *alterDatabaseSurvivalGoalNode) Close(context.Context) {
	__antithesis_instrumentation__.Notify(242833)
}

type alterDatabasePlacementNode struct {
	n    *tree.AlterDatabasePlacement
	desc *dbdesc.Mutable
}

func (p *planner) AlterDatabasePlacement(
	ctx context.Context, n *tree.AlterDatabasePlacement,
) (planNode, error) {
	__antithesis_instrumentation__.Notify(242834)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"ALTER DATABASE",
	); err != nil {
		__antithesis_instrumentation__.Notify(242839)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242840)
	}
	__antithesis_instrumentation__.Notify(242835)

	if !p.EvalContext().SessionData().PlacementEnabled {
		__antithesis_instrumentation__.Notify(242841)
		return nil, errors.WithHint(pgerror.New(
			pgcode.FeatureNotSupported,
			"ALTER DATABASE PLACEMENT requires that the session setting "+
				"enable_multiregion_placement_policy is enabled",
		),
			"to enable, enable the session setting or the cluster "+
				"setting sql.defaults.multiregion_placement_policy.enabled",
		)
	} else {
		__antithesis_instrumentation__.Notify(242842)
	}
	__antithesis_instrumentation__.Notify(242836)

	dbDesc, err := p.Descriptors().GetMutableDatabaseByName(ctx, p.txn, string(n.Name),
		tree.DatabaseLookupFlags{Required: true},
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(242843)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242844)
	}
	__antithesis_instrumentation__.Notify(242837)
	if err := p.checkPrivilegesForMultiRegionOp(ctx, dbDesc); err != nil {
		__antithesis_instrumentation__.Notify(242845)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242846)
	}
	__antithesis_instrumentation__.Notify(242838)

	return &alterDatabasePlacementNode{n: n, desc: dbDesc}, nil
}

func (n *alterDatabasePlacementNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(242847)

	if !n.desc.IsMultiRegion() {
		__antithesis_instrumentation__.Notify(242856)
		return errors.WithHintf(
			pgerror.New(pgcode.InvalidName,
				"database must have associated regions before a placement policy can be set",
			),
			"you must first add a primary region to the database using "+
				"ALTER DATABASE %s PRIMARY REGION <region_name>",
			n.n.Name.String(),
		)
	} else {
		__antithesis_instrumentation__.Notify(242857)
	}
	__antithesis_instrumentation__.Notify(242848)

	if n.n.Placement == tree.DataPlacementRestricted && func() bool {
		__antithesis_instrumentation__.Notify(242858)
		return n.desc.RegionConfig.SurvivalGoal == descpb.SurvivalGoal_REGION_FAILURE == true
	}() == true {
		__antithesis_instrumentation__.Notify(242859)
		return errors.WithDetailf(
			pgerror.New(pgcode.InvalidParameterValue,
				"a region-survivable database cannot also have a restricted placement policy"),
			"PLACEMENT RESTRICTED may only be used with SURVIVE ZONE FAILURE",
		)
	} else {
		__antithesis_instrumentation__.Notify(242860)
	}
	__antithesis_instrumentation__.Notify(242849)

	if err := params.p.validateZoneConfigForMultiRegionDatabaseWasNotModifiedByUser(
		params.ctx,
		n.desc,
	); err != nil {
		__antithesis_instrumentation__.Notify(242861)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242862)
	}
	__antithesis_instrumentation__.Notify(242850)

	telemetry.Inc(
		sqltelemetry.AlterDatabasePlacementCounter(
			n.n.Placement.TelemetryName(),
		),
	)

	newPlacement, err := TranslateDataPlacement(n.n.Placement)
	if err != nil {
		__antithesis_instrumentation__.Notify(242863)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242864)
	}
	__antithesis_instrumentation__.Notify(242851)
	n.desc.SetPlacement(newPlacement)

	if err := params.p.writeNonDropDatabaseChange(
		params.ctx,
		n.desc,
		tree.AsStringWithFQNames(n.n, params.Ann()),
	); err != nil {
		__antithesis_instrumentation__.Notify(242865)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242866)
	}
	__antithesis_instrumentation__.Notify(242852)

	regionConfig, err := SynthesizeRegionConfig(params.ctx, params.p.txn, n.desc.ID, params.p.Descriptors())
	if err != nil {
		__antithesis_instrumentation__.Notify(242867)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242868)
	}
	__antithesis_instrumentation__.Notify(242853)

	if err := ApplyZoneConfigFromDatabaseRegionConfig(
		params.ctx,
		n.desc.ID,
		regionConfig,
		params.p.txn,
		params.p.execCfg,
	); err != nil {
		__antithesis_instrumentation__.Notify(242869)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242870)
	}
	__antithesis_instrumentation__.Notify(242854)

	if err := params.p.updateZoneConfigsForTables(
		params.ctx,
		n.desc,
		WithOnlyGlobalTables,
	); err != nil {
		__antithesis_instrumentation__.Notify(242871)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242872)
	}
	__antithesis_instrumentation__.Notify(242855)

	return params.p.logEvent(params.ctx,
		n.desc.GetID(),
		&eventpb.AlterDatabasePlacement{
			DatabaseName: n.desc.GetName(),
			Placement:    newPlacement.String(),
		},
	)
}

func (n *alterDatabasePlacementNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(242873)
	return false, nil
}
func (n *alterDatabasePlacementNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(242874)
	return tree.Datums{}
}
func (n *alterDatabasePlacementNode) Close(context.Context) {
	__antithesis_instrumentation__.Notify(242875)
}

type alterDatabaseAddSuperRegion struct {
	n    *tree.AlterDatabaseAddSuperRegion
	desc *dbdesc.Mutable
}

func (p *planner) AlterDatabaseAddSuperRegion(
	ctx context.Context, n *tree.AlterDatabaseAddSuperRegion,
) (planNode, error) {
	__antithesis_instrumentation__.Notify(242876)
	if !p.ExecCfg().Settings.Version.IsActive(ctx, clusterversion.SuperRegions) {
		__antithesis_instrumentation__.Notify(242882)
		return nil, errors.Newf("super regions are not supported until upgrade to version %s is finalized", clusterversion.SuperRegions.String())
	} else {
		__antithesis_instrumentation__.Notify(242883)
	}
	__antithesis_instrumentation__.Notify(242877)
	if err := p.isSuperRegionEnabled(); err != nil {
		__antithesis_instrumentation__.Notify(242884)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242885)
	}
	__antithesis_instrumentation__.Notify(242878)

	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"ALTER DATABASE",
	); err != nil {
		__antithesis_instrumentation__.Notify(242886)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242887)
	}
	__antithesis_instrumentation__.Notify(242879)

	dbDesc, err := p.Descriptors().GetMutableDatabaseByName(ctx, p.txn, string(n.DatabaseName),
		tree.DatabaseLookupFlags{Required: true},
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(242888)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242889)
	}
	__antithesis_instrumentation__.Notify(242880)
	if err := p.checkPrivilegesForMultiRegionOp(ctx, dbDesc); err != nil {
		__antithesis_instrumentation__.Notify(242890)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242891)
	}
	__antithesis_instrumentation__.Notify(242881)

	return &alterDatabaseAddSuperRegion{n: n, desc: dbDesc}, nil
}

func (n *alterDatabaseAddSuperRegion) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(242892)

	if !n.desc.IsMultiRegion() {
		__antithesis_instrumentation__.Notify(242896)
		return errors.WithHintf(
			pgerror.New(pgcode.InvalidName,
				"database must be multi-region to support super regions",
			),
			"you must first add a primary region to the database using "+
				"ALTER DATABASE %s PRIMARY REGION <region_name>",
			n.n.DatabaseName.String(),
		)
	} else {
		__antithesis_instrumentation__.Notify(242897)
	}
	__antithesis_instrumentation__.Notify(242893)

	typeID, err := n.desc.MultiRegionEnumID()
	if err != nil {
		__antithesis_instrumentation__.Notify(242898)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242899)
	}
	__antithesis_instrumentation__.Notify(242894)
	typeDesc, err := params.p.Descriptors().GetMutableTypeVersionByID(params.ctx, params.p.txn, typeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(242900)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242901)
	}
	__antithesis_instrumentation__.Notify(242895)

	return params.p.addSuperRegion(params.ctx, n.desc, typeDesc, n.n.Regions, n.n.SuperRegionName, tree.AsStringWithFQNames(n.n, params.Ann()))
}

func addSuperRegion(r *descpb.TypeDescriptor_RegionConfig, superRegion descpb.SuperRegion) {
	__antithesis_instrumentation__.Notify(242902)
	idx := sort.Search(len(r.SuperRegions), func(i int) bool {
		__antithesis_instrumentation__.Notify(242904)
		return !(r.SuperRegions[i].SuperRegionName < superRegion.SuperRegionName)
	})
	__antithesis_instrumentation__.Notify(242903)
	if idx == len(r.SuperRegions) {
		__antithesis_instrumentation__.Notify(242905)

		r.SuperRegions = append(r.SuperRegions, superRegion)
	} else {
		__antithesis_instrumentation__.Notify(242906)

		r.SuperRegions = append(r.SuperRegions, descpb.SuperRegion{})
		copy(r.SuperRegions[idx+1:], r.SuperRegions[idx:])
		r.SuperRegions[idx] = superRegion
	}
}

func (n *alterDatabaseAddSuperRegion) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(242907)
	return false, nil
}
func (n *alterDatabaseAddSuperRegion) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(242908)
	return tree.Datums{}
}
func (n *alterDatabaseAddSuperRegion) Close(context.Context) {
	__antithesis_instrumentation__.Notify(242909)
}

type alterDatabaseDropSuperRegion struct {
	n    *tree.AlterDatabaseDropSuperRegion
	desc *dbdesc.Mutable
}

func (p *planner) AlterDatabaseDropSuperRegion(
	ctx context.Context, n *tree.AlterDatabaseDropSuperRegion,
) (planNode, error) {
	__antithesis_instrumentation__.Notify(242910)
	if !p.ExecCfg().Settings.Version.IsActive(ctx, clusterversion.SuperRegions) {
		__antithesis_instrumentation__.Notify(242916)
		return nil, errors.Newf("super regions are not supported until upgrade to version %s is finalized", clusterversion.SuperRegions.String())
	} else {
		__antithesis_instrumentation__.Notify(242917)
	}
	__antithesis_instrumentation__.Notify(242911)

	if err := p.isSuperRegionEnabled(); err != nil {
		__antithesis_instrumentation__.Notify(242918)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242919)
	}
	__antithesis_instrumentation__.Notify(242912)

	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"ALTER DATABASE",
	); err != nil {
		__antithesis_instrumentation__.Notify(242920)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242921)
	}
	__antithesis_instrumentation__.Notify(242913)

	dbDesc, err := p.Descriptors().GetMutableDatabaseByName(ctx, p.txn, string(n.DatabaseName),
		tree.DatabaseLookupFlags{Required: true},
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(242922)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242923)
	}
	__antithesis_instrumentation__.Notify(242914)
	if err := p.checkPrivilegesForMultiRegionOp(ctx, dbDesc); err != nil {
		__antithesis_instrumentation__.Notify(242924)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242925)
	}
	__antithesis_instrumentation__.Notify(242915)

	return &alterDatabaseDropSuperRegion{n: n, desc: dbDesc}, nil
}

func (n *alterDatabaseDropSuperRegion) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(242926)

	if !n.desc.IsMultiRegion() {
		__antithesis_instrumentation__.Notify(242933)
		return errors.WithHintf(
			pgerror.New(pgcode.InvalidName,
				"database must be multi-region to support super regions",
			),
			"you must first add a primary region to the database using "+
				"ALTER DATABASE %s PRIMARY REGION <region_name>",
			n.n.DatabaseName.String(),
		)
	} else {
		__antithesis_instrumentation__.Notify(242934)
	}
	__antithesis_instrumentation__.Notify(242927)

	typeID, err := n.desc.MultiRegionEnumID()
	if err != nil {
		__antithesis_instrumentation__.Notify(242935)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242936)
	}
	__antithesis_instrumentation__.Notify(242928)
	typeDesc, err := params.p.Descriptors().GetMutableTypeVersionByID(params.ctx, params.p.txn, typeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(242937)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242938)
	}
	__antithesis_instrumentation__.Notify(242929)

	dropped := removeSuperRegion(typeDesc.RegionConfig, n.n.SuperRegionName)

	if !dropped {
		__antithesis_instrumentation__.Notify(242939)
		return errors.Newf("super region %s not found", n.n.SuperRegionName)
	} else {
		__antithesis_instrumentation__.Notify(242940)
	}
	__antithesis_instrumentation__.Notify(242930)

	if err := params.p.writeTypeSchemaChange(params.ctx, typeDesc, tree.AsStringWithFQNames(n.n, params.Ann())); err != nil {
		__antithesis_instrumentation__.Notify(242941)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242942)
	}
	__antithesis_instrumentation__.Notify(242931)

	if err := params.p.updateZoneConfigsForTables(
		params.ctx,
		n.desc,
	); err != nil {
		__antithesis_instrumentation__.Notify(242943)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242944)
	}
	__antithesis_instrumentation__.Notify(242932)

	return nil
}

func removeSuperRegion(
	r *descpb.TypeDescriptor_RegionConfig, superRegionName tree.Name,
) (dropped bool) {
	__antithesis_instrumentation__.Notify(242945)
	for i, superRegion := range r.SuperRegions {
		__antithesis_instrumentation__.Notify(242947)
		if superRegion.SuperRegionName == string(superRegionName) {
			__antithesis_instrumentation__.Notify(242948)
			r.SuperRegions = append(r.SuperRegions[:i], r.SuperRegions[i+1:]...)
			dropped = true
			break
		} else {
			__antithesis_instrumentation__.Notify(242949)
		}
	}
	__antithesis_instrumentation__.Notify(242946)

	return dropped
}

func (n *alterDatabaseDropSuperRegion) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(242950)
	return false, nil
}
func (n *alterDatabaseDropSuperRegion) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(242951)
	return tree.Datums{}
}
func (n *alterDatabaseDropSuperRegion) Close(context.Context) {
	__antithesis_instrumentation__.Notify(242952)
}

func (p *planner) isSuperRegionEnabled() error {
	__antithesis_instrumentation__.Notify(242953)
	if !p.SessionData().EnableSuperRegions {
		__antithesis_instrumentation__.Notify(242955)
		return errors.WithTelemetry(
			pgerror.WithCandidateCode(
				errors.WithHint(errors.New("super regions are only supported experimentally"),
					"You can enable super regions by running `SET enable_super_regions = 'on'`."),
				pgcode.FeatureNotSupported),
			"sql.schema.super_regions_disabled",
		)
	} else {
		__antithesis_instrumentation__.Notify(242956)
	}
	__antithesis_instrumentation__.Notify(242954)

	return nil
}

func (p *planner) getSuperRegionsForDatabase(
	ctx context.Context, desc catalog.DatabaseDescriptor,
) ([]descpb.SuperRegion, error) {
	__antithesis_instrumentation__.Notify(242957)
	typeID, err := desc.MultiRegionEnumID()
	if err != nil {
		__antithesis_instrumentation__.Notify(242960)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242961)
	}
	__antithesis_instrumentation__.Notify(242958)
	typeDesc, err := p.Descriptors().GetMutableTypeVersionByID(ctx, p.txn, typeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(242962)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242963)
	}
	__antithesis_instrumentation__.Notify(242959)

	return typeDesc.RegionConfig.SuperRegions, nil
}

type alterDatabaseAlterSuperRegion struct {
	n    *tree.AlterDatabaseAlterSuperRegion
	desc *dbdesc.Mutable
}

func (n *alterDatabaseAlterSuperRegion) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(242964)
	return false, nil
}
func (n *alterDatabaseAlterSuperRegion) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(242965)
	return tree.Datums{}
}
func (n *alterDatabaseAlterSuperRegion) Close(context.Context) {
	__antithesis_instrumentation__.Notify(242966)
}

func (p *planner) AlterDatabaseAlterSuperRegion(
	ctx context.Context, n *tree.AlterDatabaseAlterSuperRegion,
) (planNode, error) {
	__antithesis_instrumentation__.Notify(242967)
	if !p.ExecCfg().Settings.Version.IsActive(ctx, clusterversion.SuperRegions) {
		__antithesis_instrumentation__.Notify(242973)
		return nil, errors.Newf("super regions are not supported until upgrade to version %s is finalized", clusterversion.SuperRegions.String())
	} else {
		__antithesis_instrumentation__.Notify(242974)
	}
	__antithesis_instrumentation__.Notify(242968)
	if err := p.isSuperRegionEnabled(); err != nil {
		__antithesis_instrumentation__.Notify(242975)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242976)
	}
	__antithesis_instrumentation__.Notify(242969)

	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"ALTER DATABASE",
	); err != nil {
		__antithesis_instrumentation__.Notify(242977)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242978)
	}
	__antithesis_instrumentation__.Notify(242970)

	dbDesc, err := p.Descriptors().GetMutableDatabaseByName(ctx, p.txn, string(n.DatabaseName),
		tree.DatabaseLookupFlags{Required: true},
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(242979)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242980)
	}
	__antithesis_instrumentation__.Notify(242971)
	if err := p.checkPrivilegesForMultiRegionOp(ctx, dbDesc); err != nil {
		__antithesis_instrumentation__.Notify(242981)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(242982)
	}
	__antithesis_instrumentation__.Notify(242972)

	return &alterDatabaseAlterSuperRegion{n: n, desc: dbDesc}, nil
}

func (n *alterDatabaseAlterSuperRegion) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(242983)

	if !n.desc.IsMultiRegion() {
		__antithesis_instrumentation__.Notify(242988)
		return errors.WithHintf(
			pgerror.New(pgcode.InvalidName,
				"database must be multi-region to support super regions",
			),
			"you must first add a primary region to the database using "+
				"ALTER DATABASE %s PRIMARY REGION <region_name>",
			n.n.DatabaseName.String(),
		)
	} else {
		__antithesis_instrumentation__.Notify(242989)
	}
	__antithesis_instrumentation__.Notify(242984)

	typeID, err := n.desc.MultiRegionEnumID()
	if err != nil {
		__antithesis_instrumentation__.Notify(242990)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242991)
	}
	__antithesis_instrumentation__.Notify(242985)
	typeDesc, err := params.p.Descriptors().GetMutableTypeVersionByID(params.ctx, params.p.txn, typeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(242992)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242993)
	}
	__antithesis_instrumentation__.Notify(242986)

	dropped := removeSuperRegion(typeDesc.RegionConfig, n.n.SuperRegionName)
	if !dropped {
		__antithesis_instrumentation__.Notify(242994)
		return errors.WithHint(pgerror.Newf(pgcode.UndefinedObject,
			"super region %q of database %q does not exist", n.n.SuperRegionName, n.n.DatabaseName),
			"super region must be added before it can be altered")
	} else {
		__antithesis_instrumentation__.Notify(242995)
	}
	__antithesis_instrumentation__.Notify(242987)

	return params.p.addSuperRegion(params.ctx, n.desc, typeDesc, n.n.Regions, n.n.SuperRegionName, tree.AsStringWithFQNames(n.n, params.Ann()))

}

func (p *planner) addSuperRegion(
	ctx context.Context,
	desc *dbdesc.Mutable,
	typeDesc *typedesc.Mutable,
	regionList []tree.Name,
	superRegionName tree.Name,
	op string,
) error {
	__antithesis_instrumentation__.Notify(242996)

	regionNames, err := typeDesc.RegionNames()
	if err != nil {
		__antithesis_instrumentation__.Notify(243004)
		return err
	} else {
		__antithesis_instrumentation__.Notify(243005)
	}
	__antithesis_instrumentation__.Notify(242997)

	regionsInDatabase := make(map[catpb.RegionName]struct{})
	for _, regionName := range regionNames {
		__antithesis_instrumentation__.Notify(243006)
		regionsInDatabase[regionName] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(242998)

	regionSet := make(map[catpb.RegionName]struct{})
	regions := make([]catpb.RegionName, len(regionList))

	for i, region := range regionList {
		__antithesis_instrumentation__.Notify(243007)
		_, found := regionsInDatabase[catpb.RegionName(region)]
		if !found {
			__antithesis_instrumentation__.Notify(243009)
			return errors.Newf("region %s not part of database", region)
		} else {
			__antithesis_instrumentation__.Notify(243010)
		}
		__antithesis_instrumentation__.Notify(243008)

		regionSet[catpb.RegionName(region)] = struct{}{}
		regions[i] = catpb.RegionName(region)
	}
	__antithesis_instrumentation__.Notify(242999)

	if err := multiregion.CanSatisfySurvivalGoal(desc.RegionConfig.SurvivalGoal, len(regionList)); err != nil {
		__antithesis_instrumentation__.Notify(243011)
		return errors.Wrapf(err, "super region %s only has %d region(s)", superRegionName, len(regionList))
	} else {
		__antithesis_instrumentation__.Notify(243012)
	}
	__antithesis_instrumentation__.Notify(243000)

	sort.Slice(regions, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(243013)
		return regions[i] < regions[j]
	})
	__antithesis_instrumentation__.Notify(243001)

	for _, superRegion := range typeDesc.RegionConfig.SuperRegions {
		__antithesis_instrumentation__.Notify(243014)
		if superRegion.SuperRegionName == string(superRegionName) {
			__antithesis_instrumentation__.Notify(243016)
			return errors.Newf("super region %s already exists", superRegion.SuperRegionName)
		} else {
			__antithesis_instrumentation__.Notify(243017)
		}
		__antithesis_instrumentation__.Notify(243015)

		for _, region := range superRegion.Regions {
			__antithesis_instrumentation__.Notify(243018)
			if _, found := regionSet[region]; found {
				__antithesis_instrumentation__.Notify(243019)
				return errors.Newf("region %s is already part of super region %s", region, superRegion.SuperRegionName)
			} else {
				__antithesis_instrumentation__.Notify(243020)
			}
		}
	}
	__antithesis_instrumentation__.Notify(243002)

	addSuperRegion(typeDesc.RegionConfig, descpb.SuperRegion{
		SuperRegionName: string(superRegionName),
		Regions:         regions,
	})

	if err := p.writeTypeSchemaChange(ctx, typeDesc, op); err != nil {
		__antithesis_instrumentation__.Notify(243021)
		return err
	} else {
		__antithesis_instrumentation__.Notify(243022)
	}
	__antithesis_instrumentation__.Notify(243003)

	return p.updateZoneConfigsForTables(
		ctx,
		desc,
	)
}
