package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/config"
	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
	"github.com/gogo/protobuf/proto"
	yaml "gopkg.in/yaml.v2"
)

type optionValue struct {
	inheritValue  bool
	explicitValue tree.TypedExpr
}

type setZoneConfigNode struct {
	zoneSpecifier tree.ZoneSpecifier
	allIndexes    bool
	yamlConfig    tree.TypedExpr
	options       map[tree.Name]optionValue
	setDefault    bool

	run setZoneConfigRun
}

var supportedZoneConfigOptions = map[tree.Name]struct {
	requiredType *types.T
	setter       func(*zonepb.ZoneConfig, tree.Datum)
	checkAllowed func(context.Context, *ExecutorConfig, tree.Datum) error
}{
	"range_min_bytes": {
		requiredType: types.Int,
		setter: func(c *zonepb.ZoneConfig, d tree.Datum) {
			__antithesis_instrumentation__.Notify(622246)
			c.RangeMinBytes = proto.Int64(int64(tree.MustBeDInt(d)))
		},
	},
	"range_max_bytes": {
		requiredType: types.Int,
		setter: func(c *zonepb.ZoneConfig, d tree.Datum) {
			__antithesis_instrumentation__.Notify(622247)
			c.RangeMaxBytes = proto.Int64(int64(tree.MustBeDInt(d)))
		},
	},
	"global_reads": {
		requiredType: types.Bool,
		setter: func(c *zonepb.ZoneConfig, d tree.Datum) {
			__antithesis_instrumentation__.Notify(622248)
			c.GlobalReads = proto.Bool(bool(tree.MustBeDBool(d)))
		},
		checkAllowed: func(ctx context.Context, execCfg *ExecutorConfig, d tree.Datum) error {
			__antithesis_instrumentation__.Notify(622249)
			if !tree.MustBeDBool(d) {
				__antithesis_instrumentation__.Notify(622251)

				return nil
			} else {
				__antithesis_instrumentation__.Notify(622252)
			}
			__antithesis_instrumentation__.Notify(622250)
			return base.CheckEnterpriseEnabled(
				execCfg.Settings,
				execCfg.LogicalClusterID(),
				execCfg.Organization(),
				"global_reads",
			)
		},
	},
	"num_replicas": {
		requiredType: types.Int,
		setter: func(c *zonepb.ZoneConfig, d tree.Datum) {
			__antithesis_instrumentation__.Notify(622253)
			c.NumReplicas = proto.Int32(int32(tree.MustBeDInt(d)))
		},
	},
	"num_voters": {
		requiredType: types.Int,
		setter: func(c *zonepb.ZoneConfig, d tree.Datum) {
			__antithesis_instrumentation__.Notify(622254)
			c.NumVoters = proto.Int32(int32(tree.MustBeDInt(d)))
		},
	},
	"gc.ttlseconds": {
		requiredType: types.Int,
		setter: func(c *zonepb.ZoneConfig, d tree.Datum) {
			__antithesis_instrumentation__.Notify(622255)
			c.GC = &zonepb.GCPolicy{TTLSeconds: int32(tree.MustBeDInt(d))}
		},
	},
	"constraints": {
		requiredType: types.String,
		setter: func(c *zonepb.ZoneConfig, d tree.Datum) {
			__antithesis_instrumentation__.Notify(622256)
			constraintsList := zonepb.ConstraintsList{
				Constraints: c.Constraints,
				Inherited:   c.InheritedConstraints,
			}
			loadYAML(&constraintsList, string(tree.MustBeDString(d)))
			c.Constraints = constraintsList.Constraints
			c.InheritedConstraints = false
		},
	},
	"voter_constraints": {
		requiredType: types.String,
		setter: func(c *zonepb.ZoneConfig, d tree.Datum) {
			__antithesis_instrumentation__.Notify(622257)
			voterConstraintsList := zonepb.ConstraintsList{
				Constraints: c.VoterConstraints,
				Inherited:   c.InheritedVoterConstraints(),
			}
			loadYAML(&voterConstraintsList, string(tree.MustBeDString(d)))
			c.VoterConstraints = voterConstraintsList.Constraints
			c.NullVoterConstraintsIsEmpty = true
		},
	},
	"lease_preferences": {
		requiredType: types.String,
		setter: func(c *zonepb.ZoneConfig, d tree.Datum) {
			__antithesis_instrumentation__.Notify(622258)
			loadYAML(&c.LeasePreferences, string(tree.MustBeDString(d)))
			c.InheritedLeasePreferences = false
		},
	},
}

var zoneOptionKeys = func() []string {
	__antithesis_instrumentation__.Notify(622259)
	l := make([]string, 0, len(supportedZoneConfigOptions))
	for k := range supportedZoneConfigOptions {
		__antithesis_instrumentation__.Notify(622261)
		l = append(l, string(k))
	}
	__antithesis_instrumentation__.Notify(622260)
	sort.Strings(l)
	return l
}()

func loadYAML(dst interface{}, yamlString string) {
	__antithesis_instrumentation__.Notify(622262)
	if err := yaml.UnmarshalStrict([]byte(yamlString), dst); err != nil {
		__antithesis_instrumentation__.Notify(622263)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(622264)
	}
}

func (p *planner) SetZoneConfig(ctx context.Context, n *tree.SetZoneConfig) (planNode, error) {
	__antithesis_instrumentation__.Notify(622265)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"CONFIGURE ZONE",
	); err != nil {
		__antithesis_instrumentation__.Notify(622272)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(622273)
	}
	__antithesis_instrumentation__.Notify(622266)

	if !p.ExecCfg().Codec.ForSystemTenant() && func() bool {
		__antithesis_instrumentation__.Notify(622274)
		return !secondaryTenantZoneConfigsEnabled.Get(&p.ExecCfg().Settings.SV) == true
	}() == true {
		__antithesis_instrumentation__.Notify(622275)

		return nil, errorutil.UnsupportedWithMultiTenancy(MultitenancyZoneCfgIssueNo)
	} else {
		__antithesis_instrumentation__.Notify(622276)
	}
	__antithesis_instrumentation__.Notify(622267)

	if err := checkPrivilegeForSetZoneConfig(ctx, p, n.ZoneSpecifier); err != nil {
		__antithesis_instrumentation__.Notify(622277)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(622278)
	}
	__antithesis_instrumentation__.Notify(622268)

	if err := p.CheckZoneConfigChangePermittedForMultiRegion(
		ctx,
		n.ZoneSpecifier,
		n.Options,
	); err != nil {
		__antithesis_instrumentation__.Notify(622279)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(622280)
	}
	__antithesis_instrumentation__.Notify(622269)

	var yamlConfig tree.TypedExpr

	if n.YAMLConfig != nil {
		__antithesis_instrumentation__.Notify(622281)

		var err error
		yamlConfig, err = p.analyzeExpr(
			ctx, n.YAMLConfig, nil, tree.IndexedVarHelper{}, types.String, false, "configure zone")
		if err != nil {
			__antithesis_instrumentation__.Notify(622283)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(622284)
		}
		__antithesis_instrumentation__.Notify(622282)

		switch typ := yamlConfig.ResolvedType(); typ.Family() {
		case types.UnknownFamily:
			__antithesis_instrumentation__.Notify(622285)

		case types.StringFamily:
			__antithesis_instrumentation__.Notify(622286)
		case types.BytesFamily:
			__antithesis_instrumentation__.Notify(622287)
		default:
			__antithesis_instrumentation__.Notify(622288)
			return nil, pgerror.Newf(pgcode.InvalidParameterValue,
				"zone config must be of type string or bytes, not %s", typ)
		}
	} else {
		__antithesis_instrumentation__.Notify(622289)
	}
	__antithesis_instrumentation__.Notify(622270)

	var options map[tree.Name]optionValue
	if n.Options != nil {
		__antithesis_instrumentation__.Notify(622290)

		options = make(map[tree.Name]optionValue)
		for _, opt := range n.Options {
			__antithesis_instrumentation__.Notify(622291)
			if _, alreadyExists := options[opt.Key]; alreadyExists {
				__antithesis_instrumentation__.Notify(622296)
				return nil, pgerror.Newf(pgcode.InvalidParameterValue,
					"duplicate zone config parameter: %q", tree.ErrString(&opt.Key))
			} else {
				__antithesis_instrumentation__.Notify(622297)
			}
			__antithesis_instrumentation__.Notify(622292)
			req, ok := supportedZoneConfigOptions[opt.Key]
			if !ok {
				__antithesis_instrumentation__.Notify(622298)
				return nil, pgerror.Newf(pgcode.InvalidParameterValue,
					"unsupported zone config parameter: %q", tree.ErrString(&opt.Key))
			} else {
				__antithesis_instrumentation__.Notify(622299)
			}
			__antithesis_instrumentation__.Notify(622293)
			telemetry.Inc(
				sqltelemetry.SchemaSetZoneConfigCounter(
					n.ZoneSpecifier.TelemetryName(),
					string(opt.Key),
				),
			)
			if opt.Value == nil {
				__antithesis_instrumentation__.Notify(622300)
				options[opt.Key] = optionValue{inheritValue: true, explicitValue: nil}
				continue
			} else {
				__antithesis_instrumentation__.Notify(622301)
			}
			__antithesis_instrumentation__.Notify(622294)
			valExpr, err := p.analyzeExpr(
				ctx, opt.Value, nil, tree.IndexedVarHelper{}, req.requiredType, true, string(opt.Key))
			if err != nil {
				__antithesis_instrumentation__.Notify(622302)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(622303)
			}
			__antithesis_instrumentation__.Notify(622295)
			options[opt.Key] = optionValue{inheritValue: false, explicitValue: valExpr}
		}
	} else {
		__antithesis_instrumentation__.Notify(622304)
	}
	__antithesis_instrumentation__.Notify(622271)

	return &setZoneConfigNode{
		zoneSpecifier: n.ZoneSpecifier,
		allIndexes:    n.AllIndexes,
		yamlConfig:    yamlConfig,
		options:       options,
		setDefault:    n.SetDefault,
	}, nil
}

func checkPrivilegeForSetZoneConfig(ctx context.Context, p *planner, zs tree.ZoneSpecifier) error {
	__antithesis_instrumentation__.Notify(622305)

	if zs.NamedZone != "" {
		__antithesis_instrumentation__.Notify(622311)
		return p.RequireAdminRole(ctx, "alter system ranges")
	} else {
		__antithesis_instrumentation__.Notify(622312)
	}
	__antithesis_instrumentation__.Notify(622306)
	if zs.Database != "" {
		__antithesis_instrumentation__.Notify(622313)
		if zs.Database == "system" {
			__antithesis_instrumentation__.Notify(622317)
			return p.RequireAdminRole(ctx, "alter the system database")
		} else {
			__antithesis_instrumentation__.Notify(622318)
		}
		__antithesis_instrumentation__.Notify(622314)
		dbDesc, err := p.Descriptors().GetImmutableDatabaseByName(ctx, p.txn,
			string(zs.Database), tree.DatabaseLookupFlags{Required: true})
		if err != nil {
			__antithesis_instrumentation__.Notify(622319)
			return err
		} else {
			__antithesis_instrumentation__.Notify(622320)
		}
		__antithesis_instrumentation__.Notify(622315)
		dbCreatePrivilegeErr := p.CheckPrivilege(ctx, dbDesc, privilege.CREATE)
		dbZoneConfigPrivilegeErr := p.CheckPrivilege(ctx, dbDesc, privilege.ZONECONFIG)

		if dbZoneConfigPrivilegeErr == nil || func() bool {
			__antithesis_instrumentation__.Notify(622321)
			return dbCreatePrivilegeErr == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(622322)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(622323)
		}
		__antithesis_instrumentation__.Notify(622316)

		return pgerror.Newf(pgcode.InsufficientPrivilege,
			"user %s does not have %s or %s privilege on %s %s",
			p.SessionData().User(), privilege.ZONECONFIG, privilege.CREATE, dbDesc.DescriptorType(), dbDesc.GetName())
	} else {
		__antithesis_instrumentation__.Notify(622324)
	}
	__antithesis_instrumentation__.Notify(622307)
	tableDesc, err := p.resolveTableForZone(ctx, &zs)
	if err != nil {
		__antithesis_instrumentation__.Notify(622325)
		if zs.TargetsIndex() && func() bool {
			__antithesis_instrumentation__.Notify(622327)
			return zs.TableOrIndex.Table.ObjectName == "" == true
		}() == true {
			__antithesis_instrumentation__.Notify(622328)
			err = errors.WithHint(err, "try specifying the index as <tablename>@<indexname>")
		} else {
			__antithesis_instrumentation__.Notify(622329)
		}
		__antithesis_instrumentation__.Notify(622326)
		return err
	} else {
		__antithesis_instrumentation__.Notify(622330)
	}
	__antithesis_instrumentation__.Notify(622308)
	if tableDesc.GetParentID() == keys.SystemDatabaseID {
		__antithesis_instrumentation__.Notify(622331)
		return p.RequireAdminRole(ctx, "alter system tables")
	} else {
		__antithesis_instrumentation__.Notify(622332)
	}
	__antithesis_instrumentation__.Notify(622309)

	tableCreatePrivilegeErr := p.CheckPrivilege(ctx, tableDesc, privilege.CREATE)
	tableZoneConfigPrivilegeErr := p.CheckPrivilege(ctx, tableDesc, privilege.ZONECONFIG)

	if tableCreatePrivilegeErr == nil || func() bool {
		__antithesis_instrumentation__.Notify(622333)
		return tableZoneConfigPrivilegeErr == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(622334)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(622335)
	}
	__antithesis_instrumentation__.Notify(622310)

	return pgerror.Newf(pgcode.InsufficientPrivilege,
		"user %s does not have %s or %s privilege on %s %s",
		p.SessionData().User(), privilege.ZONECONFIG, privilege.CREATE, tableDesc.DescriptorType(), tableDesc.GetName())
}

type setZoneConfigRun struct {
	numAffected int
}

func (n *setZoneConfigNode) ReadingOwnWrites() { __antithesis_instrumentation__.Notify(622336) }

func (n *setZoneConfigNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(622337)
	var yamlConfig string
	var setters []func(c *zonepb.ZoneConfig)
	deleteZone := false

	if n.yamlConfig != nil {
		__antithesis_instrumentation__.Notify(622346)

		datum, err := n.yamlConfig.Eval(params.EvalContext())
		if err != nil {
			__antithesis_instrumentation__.Notify(622349)
			return err
		} else {
			__antithesis_instrumentation__.Notify(622350)
		}
		__antithesis_instrumentation__.Notify(622347)
		switch val := datum.(type) {
		case *tree.DString:
			__antithesis_instrumentation__.Notify(622351)
			yamlConfig = string(*val)
		case *tree.DBytes:
			__antithesis_instrumentation__.Notify(622352)
			yamlConfig = string(*val)
		default:
			__antithesis_instrumentation__.Notify(622353)
			deleteZone = true
		}
		__antithesis_instrumentation__.Notify(622348)

		yamlConfig = strings.TrimSpace(yamlConfig)
	} else {
		__antithesis_instrumentation__.Notify(622354)
	}
	__antithesis_instrumentation__.Notify(622338)
	var optionsStr []string
	var copyFromParentList []tree.Name
	if n.options != nil {
		__antithesis_instrumentation__.Notify(622355)

		for i := range zoneOptionKeys {
			__antithesis_instrumentation__.Notify(622356)
			name := (*tree.Name)(&zoneOptionKeys[i])
			val, ok := n.options[*name]
			if !ok {
				__antithesis_instrumentation__.Notify(622363)
				continue
			} else {
				__antithesis_instrumentation__.Notify(622364)
			}
			__antithesis_instrumentation__.Notify(622357)

			inheritVal, expr := val.inheritValue, val.explicitValue
			if inheritVal {
				__antithesis_instrumentation__.Notify(622365)
				copyFromParentList = append(copyFromParentList, *name)
				optionsStr = append(optionsStr, fmt.Sprintf("%s = COPY FROM PARENT", name))
				continue
			} else {
				__antithesis_instrumentation__.Notify(622366)
			}
			__antithesis_instrumentation__.Notify(622358)
			datum, err := expr.Eval(params.EvalContext())
			if err != nil {
				__antithesis_instrumentation__.Notify(622367)
				return err
			} else {
				__antithesis_instrumentation__.Notify(622368)
			}
			__antithesis_instrumentation__.Notify(622359)
			if datum == tree.DNull {
				__antithesis_instrumentation__.Notify(622369)
				return pgerror.Newf(pgcode.InvalidParameterValue,
					"unsupported NULL value for %q", tree.ErrString(name))
			} else {
				__antithesis_instrumentation__.Notify(622370)
			}
			__antithesis_instrumentation__.Notify(622360)
			opt := supportedZoneConfigOptions[*name]
			if opt.checkAllowed != nil {
				__antithesis_instrumentation__.Notify(622371)
				if err := opt.checkAllowed(params.ctx, params.ExecCfg(), datum); err != nil {
					__antithesis_instrumentation__.Notify(622372)
					return err
				} else {
					__antithesis_instrumentation__.Notify(622373)
				}
			} else {
				__antithesis_instrumentation__.Notify(622374)
			}
			__antithesis_instrumentation__.Notify(622361)
			setter := opt.setter
			setters = append(setters, func(c *zonepb.ZoneConfig) { __antithesis_instrumentation__.Notify(622375); setter(c, datum) })
			__antithesis_instrumentation__.Notify(622362)
			optionsStr = append(optionsStr, fmt.Sprintf("%s = %s", name, datum))
		}
	} else {
		__antithesis_instrumentation__.Notify(622376)
	}
	__antithesis_instrumentation__.Notify(622339)

	telemetry.Inc(
		sqltelemetry.SchemaChangeAlterCounterWithExtra(n.zoneSpecifier.TelemetryName(), "configure_zone"),
	)

	table, err := params.p.resolveTableForZone(params.ctx, &n.zoneSpecifier)
	if err != nil {
		__antithesis_instrumentation__.Notify(622377)
		return err
	} else {
		__antithesis_instrumentation__.Notify(622378)
	}
	__antithesis_instrumentation__.Notify(622340)

	if table != nil && func() bool {
		__antithesis_instrumentation__.Notify(622379)
		return !table.IsPhysicalTable() == true
	}() == true {
		__antithesis_instrumentation__.Notify(622380)
		return pgerror.Newf(pgcode.WrongObjectType, "cannot set a zone configuration on non-physical object %s", table.GetName())
	} else {
		__antithesis_instrumentation__.Notify(622381)
	}
	__antithesis_instrumentation__.Notify(622341)

	if n.zoneSpecifier.TargetsPartition() && func() bool {
		__antithesis_instrumentation__.Notify(622382)
		return len(n.zoneSpecifier.TableOrIndex.Index) == 0 == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(622383)
		return !n.allIndexes == true
	}() == true {
		__antithesis_instrumentation__.Notify(622384)

		partitionName := string(n.zoneSpecifier.Partition)

		var indexes []catalog.Index
		for _, idx := range table.NonDropIndexes() {
			__antithesis_instrumentation__.Notify(622386)
			if idx.GetPartitioning().FindPartitionByName(partitionName) != nil {
				__antithesis_instrumentation__.Notify(622387)
				indexes = append(indexes, idx)
			} else {
				__antithesis_instrumentation__.Notify(622388)
			}
		}
		__antithesis_instrumentation__.Notify(622385)

		switch len(indexes) {
		case 0:
			__antithesis_instrumentation__.Notify(622389)
			return fmt.Errorf("partition %q does not exist on table %q", partitionName, table.GetName())
		case 1:
			__antithesis_instrumentation__.Notify(622390)
			n.zoneSpecifier.TableOrIndex.Index = tree.UnrestrictedName(indexes[0].GetName())
		case 2:
			__antithesis_instrumentation__.Notify(622391)

			if catalog.IsCorrespondingTemporaryIndex(indexes[1], indexes[0]) {
				__antithesis_instrumentation__.Notify(622394)
				n.zoneSpecifier.TableOrIndex.Index = tree.UnrestrictedName(indexes[0].GetName())
				break
			} else {
				__antithesis_instrumentation__.Notify(622395)
			}
			__antithesis_instrumentation__.Notify(622392)
			fallthrough
		default:
			__antithesis_instrumentation__.Notify(622393)
			err := fmt.Errorf(
				"partition %q exists on multiple indexes of table %q", partitionName, table.GetName())
			err = pgerror.WithCandidateCode(err, pgcode.InvalidParameterValue)
			err = errors.WithHint(err, "try ALTER PARTITION ... OF INDEX ...")
			return err
		}
	} else {
		__antithesis_instrumentation__.Notify(622396)
	}
	__antithesis_instrumentation__.Notify(622342)

	var specifiers []tree.ZoneSpecifier
	if n.zoneSpecifier.TargetsPartition() && func() bool {
		__antithesis_instrumentation__.Notify(622397)
		return n.allIndexes == true
	}() == true {
		__antithesis_instrumentation__.Notify(622398)
		sqltelemetry.IncrementPartitioningCounter(sqltelemetry.AlterAllPartitions)
		for _, idx := range table.NonDropIndexes() {
			__antithesis_instrumentation__.Notify(622399)
			if idx.GetPartitioning().FindPartitionByName(string(n.zoneSpecifier.Partition)) != nil {
				__antithesis_instrumentation__.Notify(622400)
				zs := n.zoneSpecifier
				zs.TableOrIndex.Index = tree.UnrestrictedName(idx.GetName())
				specifiers = append(specifiers, zs)
			} else {
				__antithesis_instrumentation__.Notify(622401)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(622402)
		specifiers = append(specifiers, n.zoneSpecifier)
	}
	__antithesis_instrumentation__.Notify(622343)

	applyZoneConfig := func(zs tree.ZoneSpecifier) error {
		__antithesis_instrumentation__.Notify(622403)
		subzonePlaceholder := false

		targetID, err := resolveZone(params.ctx, params.p.txn, params.p.Descriptors(), &zs, params.ExecCfg().Settings.Version)
		if err != nil {
			__antithesis_instrumentation__.Notify(622416)
			return err
		} else {
			__antithesis_instrumentation__.Notify(622417)
		}
		__antithesis_instrumentation__.Notify(622404)

		if targetID != keys.SystemDatabaseID && func() bool {
			__antithesis_instrumentation__.Notify(622418)
			return descpb.IsSystemConfigID(targetID) == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(622419)
			return targetID == keys.NamespaceTableID == true
		}() == true {
			__antithesis_instrumentation__.Notify(622420)
			return pgerror.Newf(pgcode.CheckViolation,
				`cannot set zone configs for system config tables; `+
					`try setting your config on the entire "system" database instead`)
		} else {
			__antithesis_instrumentation__.Notify(622421)
			if targetID == keys.RootNamespaceID && func() bool {
				__antithesis_instrumentation__.Notify(622422)
				return deleteZone == true
			}() == true {
				__antithesis_instrumentation__.Notify(622423)
				return pgerror.Newf(pgcode.CheckViolation,
					"cannot remove default zone")
			} else {
				__antithesis_instrumentation__.Notify(622424)
			}
		}
		__antithesis_instrumentation__.Notify(622405)

		if !params.p.execCfg.Codec.ForSystemTenant() {
			__antithesis_instrumentation__.Notify(622425)
			zoneName, found := zonepb.NamedZonesByID[uint32(targetID)]
			if found && func() bool {
				__antithesis_instrumentation__.Notify(622426)
				return zoneName != zonepb.DefaultZoneName == true
			}() == true {
				__antithesis_instrumentation__.Notify(622427)
				return pgerror.Newf(
					pgcode.CheckViolation,
					"non-system tenants cannot configure zone for %s range",
					zoneName,
				)
			} else {
				__antithesis_instrumentation__.Notify(622428)
			}
		} else {
			__antithesis_instrumentation__.Notify(622429)
		}
		__antithesis_instrumentation__.Notify(622406)

		index, partition, err := resolveSubzone(&zs, table)
		if err != nil {
			__antithesis_instrumentation__.Notify(622430)
			return err
		} else {
			__antithesis_instrumentation__.Notify(622431)
		}
		__antithesis_instrumentation__.Notify(622407)

		var tempIndex catalog.Index
		if index != nil {
			__antithesis_instrumentation__.Notify(622432)
			tempIndex = catalog.FindCorrespondingTemporaryIndexByID(table, index.GetID())
		} else {
			__antithesis_instrumentation__.Notify(622433)
		}
		__antithesis_instrumentation__.Notify(622408)

		partialZone, err := getZoneConfigRaw(
			params.ctx, params.p.txn, params.ExecCfg().Codec, params.ExecCfg().Settings, targetID,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(622434)
			return err
		} else {
			__antithesis_instrumentation__.Notify(622435)
		}
		__antithesis_instrumentation__.Notify(622409)

		if partialZone == nil {
			__antithesis_instrumentation__.Notify(622436)
			partialZone = zonepb.NewZoneConfig()
			if index != nil {
				__antithesis_instrumentation__.Notify(622437)
				subzonePlaceholder = true
			} else {
				__antithesis_instrumentation__.Notify(622438)
			}
		} else {
			__antithesis_instrumentation__.Notify(622439)
		}
		__antithesis_instrumentation__.Notify(622410)

		var partialSubzone *zonepb.Subzone
		if index != nil {
			__antithesis_instrumentation__.Notify(622440)
			partialSubzone = partialZone.GetSubzoneExact(uint32(index.GetID()), partition)
			if partialSubzone == nil {
				__antithesis_instrumentation__.Notify(622441)
				partialSubzone = &zonepb.Subzone{Config: *zonepb.NewZoneConfig()}
			} else {
				__antithesis_instrumentation__.Notify(622442)
			}
		} else {
			__antithesis_instrumentation__.Notify(622443)
		}
		__antithesis_instrumentation__.Notify(622411)

		_, completeZone, completeSubzone, err := GetZoneConfigInTxn(
			params.ctx, params.p.txn, params.ExecCfg().Codec, targetID, index, partition, n.setDefault,
		)

		if errors.Is(err, errNoZoneConfigApplies) {
			__antithesis_instrumentation__.Notify(622444)

			completeZone = protoutil.Clone(
				params.extendedEvalCtx.ExecCfg.DefaultZoneConfig).(*zonepb.ZoneConfig)
		} else {
			__antithesis_instrumentation__.Notify(622445)
			if err != nil {
				__antithesis_instrumentation__.Notify(622446)
				return err
			} else {
				__antithesis_instrumentation__.Notify(622447)
			}
		}

		{
			__antithesis_instrumentation__.Notify(622448)

			getKey := func(key roachpb.Key) (*roachpb.Value, error) {
				__antithesis_instrumentation__.Notify(622450)
				kv, err := params.p.txn.Get(params.ctx, key)
				if err != nil {
					__antithesis_instrumentation__.Notify(622452)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(622453)
				}
				__antithesis_instrumentation__.Notify(622451)
				return kv.Value, nil
			}
			__antithesis_instrumentation__.Notify(622449)
			if index == nil {
				__antithesis_instrumentation__.Notify(622454)

				zoneInheritedFields := zonepb.ZoneConfig{}
				if err := completeZoneConfig(
					&zoneInheritedFields, params.ExecCfg().Codec, targetID, getKey,
				); err != nil {
					__antithesis_instrumentation__.Notify(622456)
					return err
				} else {
					__antithesis_instrumentation__.Notify(622457)
				}
				__antithesis_instrumentation__.Notify(622455)
				partialZone.CopyFromZone(zoneInheritedFields, copyFromParentList)
			} else {
				__antithesis_instrumentation__.Notify(622458)

				zoneInheritedFields := *partialZone
				if err := completeZoneConfig(
					&zoneInheritedFields, params.ExecCfg().Codec, targetID, getKey,
				); err != nil {
					__antithesis_instrumentation__.Notify(622460)
					return err
				} else {
					__antithesis_instrumentation__.Notify(622461)
				}
				__antithesis_instrumentation__.Notify(622459)

				if partition == "" {
					__antithesis_instrumentation__.Notify(622462)
					partialSubzone.Config.CopyFromZone(zoneInheritedFields, copyFromParentList)
				} else {
					__antithesis_instrumentation__.Notify(622463)

					subzoneInheritedFields := zonepb.ZoneConfig{}
					if indexSubzone := completeZone.GetSubzone(uint32(index.GetID()), ""); indexSubzone != nil {
						__antithesis_instrumentation__.Notify(622465)
						subzoneInheritedFields.InheritFromParent(&indexSubzone.Config)
					} else {
						__antithesis_instrumentation__.Notify(622466)
					}
					__antithesis_instrumentation__.Notify(622464)
					subzoneInheritedFields.InheritFromParent(&zoneInheritedFields)

					partialSubzone.Config.CopyFromZone(subzoneInheritedFields, copyFromParentList)
				}
			}
		}
		__antithesis_instrumentation__.Notify(622412)

		if deleteZone {
			__antithesis_instrumentation__.Notify(622467)
			if index != nil {
				__antithesis_instrumentation__.Notify(622468)
				didDelete := completeZone.DeleteSubzone(uint32(index.GetID()), partition)
				_ = partialZone.DeleteSubzone(uint32(index.GetID()), partition)
				if !didDelete {
					__antithesis_instrumentation__.Notify(622469)

					return nil
				} else {
					__antithesis_instrumentation__.Notify(622470)
				}
			} else {
				__antithesis_instrumentation__.Notify(622471)
				completeZone.DeleteTableConfig()
				partialZone.DeleteTableConfig()
			}
		} else {
			__antithesis_instrumentation__.Notify(622472)

			if len(yamlConfig) == 0 || func() bool {
				__antithesis_instrumentation__.Notify(622484)
				return yamlConfig[len(yamlConfig)-1] != '\n' == true
			}() == true {
				__antithesis_instrumentation__.Notify(622485)

				yamlConfig += "\n"
			} else {
				__antithesis_instrumentation__.Notify(622486)
			}
			__antithesis_instrumentation__.Notify(622473)

			newZone := *completeZone
			if completeSubzone != nil {
				__antithesis_instrumentation__.Notify(622487)
				newZone = completeSubzone.Config
			} else {
				__antithesis_instrumentation__.Notify(622488)
			}
			__antithesis_instrumentation__.Notify(622474)

			finalZone := *partialZone
			if partialSubzone != nil {
				__antithesis_instrumentation__.Notify(622489)
				finalZone = partialSubzone.Config
			} else {
				__antithesis_instrumentation__.Notify(622490)
			}
			__antithesis_instrumentation__.Notify(622475)

			if n.setDefault && func() bool {
				__antithesis_instrumentation__.Notify(622491)
				return keys.RootNamespaceID == uint32(targetID) == true
			}() == true {
				__antithesis_instrumentation__.Notify(622492)
				finalZone = *protoutil.Clone(
					params.extendedEvalCtx.ExecCfg.DefaultZoneConfig).(*zonepb.ZoneConfig)
			} else {
				__antithesis_instrumentation__.Notify(622493)
				if n.setDefault {
					__antithesis_instrumentation__.Notify(622494)
					finalZone = *zonepb.NewZoneConfig()
				} else {
					__antithesis_instrumentation__.Notify(622495)
				}
			}
			__antithesis_instrumentation__.Notify(622476)

			if err := yaml.UnmarshalStrict([]byte(yamlConfig), &newZone); err != nil {
				__antithesis_instrumentation__.Notify(622496)
				return pgerror.Wrap(err, pgcode.CheckViolation, "could not parse zone config")
			} else {
				__antithesis_instrumentation__.Notify(622497)
			}
			__antithesis_instrumentation__.Notify(622477)

			if err := yaml.UnmarshalStrict([]byte(yamlConfig), &finalZone); err != nil {
				__antithesis_instrumentation__.Notify(622498)
				return pgerror.Wrap(err, pgcode.CheckViolation, "could not parse zone config")
			} else {
				__antithesis_instrumentation__.Notify(622499)
			}
			__antithesis_instrumentation__.Notify(622478)

			for _, setter := range setters {
				__antithesis_instrumentation__.Notify(622500)

				if err := func() (err error) {
					__antithesis_instrumentation__.Notify(622501)
					defer func() {
						__antithesis_instrumentation__.Notify(622503)
						if p := recover(); p != nil {
							__antithesis_instrumentation__.Notify(622504)
							if errP, ok := p.(error); ok {
								__antithesis_instrumentation__.Notify(622505)

								err = errP
							} else {
								__antithesis_instrumentation__.Notify(622506)

								panic(p)
							}
						} else {
							__antithesis_instrumentation__.Notify(622507)
						}
					}()
					__antithesis_instrumentation__.Notify(622502)

					setter(&newZone)
					setter(&finalZone)
					return nil
				}(); err != nil {
					__antithesis_instrumentation__.Notify(622508)
					return err
				} else {
					__antithesis_instrumentation__.Notify(622509)
				}
			}
			__antithesis_instrumentation__.Notify(622479)

			if err := validateNoRepeatKeysInZone(&newZone); err != nil {
				__antithesis_instrumentation__.Notify(622510)
				return err
			} else {
				__antithesis_instrumentation__.Notify(622511)
			}
			__antithesis_instrumentation__.Notify(622480)

			if err := validateZoneAttrsAndLocalities(params.ctx, params.p.ExecCfg(), &newZone); err != nil {
				__antithesis_instrumentation__.Notify(622512)
				return err
			} else {
				__antithesis_instrumentation__.Notify(622513)
			}
			__antithesis_instrumentation__.Notify(622481)

			if index == nil {
				__antithesis_instrumentation__.Notify(622514)

				completeZone = &newZone
				partialZone = &finalZone

				if partialZone.IsSubzonePlaceholder() {
					__antithesis_instrumentation__.Notify(622515)
					partialZone.NumReplicas = nil
				} else {
					__antithesis_instrumentation__.Notify(622516)
				}
			} else {
				__antithesis_instrumentation__.Notify(622517)

				completeZone, err = getZoneConfigRaw(
					params.ctx, params.p.txn, params.ExecCfg().Codec, params.ExecCfg().Settings, targetID,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(622520)
					return err
				} else {
					__antithesis_instrumentation__.Notify(622521)
					if completeZone == nil {
						__antithesis_instrumentation__.Notify(622522)
						completeZone = zonepb.NewZoneConfig()
					} else {
						__antithesis_instrumentation__.Notify(622523)
					}
				}
				__antithesis_instrumentation__.Notify(622518)
				completeZone.SetSubzone(zonepb.Subzone{
					IndexID:       uint32(index.GetID()),
					PartitionName: partition,
					Config:        newZone,
				})

				if subzonePlaceholder {
					__antithesis_instrumentation__.Notify(622524)
					partialZone.DeleteTableConfig()
				} else {
					__antithesis_instrumentation__.Notify(622525)
				}
				__antithesis_instrumentation__.Notify(622519)

				partialZone.SetSubzone(zonepb.Subzone{
					IndexID:       uint32(index.GetID()),
					PartitionName: partition,
					Config:        finalZone,
				})

				if tempIndex != nil {
					__antithesis_instrumentation__.Notify(622526)
					completeZone.SetSubzone(zonepb.Subzone{
						IndexID:       uint32(tempIndex.GetID()),
						PartitionName: partition,
						Config:        newZone,
					})

					partialZone.SetSubzone(zonepb.Subzone{
						IndexID:       uint32(tempIndex.GetID()),
						PartitionName: partition,
						Config:        finalZone,
					})
				} else {
					__antithesis_instrumentation__.Notify(622527)
				}
			}
			__antithesis_instrumentation__.Notify(622482)

			if err := completeZone.Validate(); err != nil {
				__antithesis_instrumentation__.Notify(622528)
				return pgerror.Wrap(err, pgcode.CheckViolation, "could not validate zone config")
			} else {
				__antithesis_instrumentation__.Notify(622529)
			}
			__antithesis_instrumentation__.Notify(622483)

			if err := finalZone.ValidateTandemFields(); err != nil {
				__antithesis_instrumentation__.Notify(622530)
				err = errors.Wrap(err, "could not validate zone config")
				err = pgerror.WithCandidateCode(err, pgcode.InvalidParameterValue)
				err = errors.WithHint(err,
					"try ALTER ... CONFIGURE ZONE USING <field_name> = COPY FROM PARENT [, ...] to populate the field")
				return err
			} else {
				__antithesis_instrumentation__.Notify(622531)
			}
		}
		__antithesis_instrumentation__.Notify(622413)

		hasNewSubzones := !deleteZone && func() bool {
			__antithesis_instrumentation__.Notify(622532)
			return index != nil == true
		}() == true
		execConfig := params.extendedEvalCtx.ExecCfg
		zoneToWrite := partialZone

		n.run.numAffected, err = writeZoneConfig(params.ctx, params.p.txn,
			targetID, table, zoneToWrite, execConfig, hasNewSubzones)
		if err != nil {
			__antithesis_instrumentation__.Notify(622533)
			return err
		} else {
			__antithesis_instrumentation__.Notify(622534)
		}
		__antithesis_instrumentation__.Notify(622414)

		eventDetails := eventpb.CommonZoneConfigDetails{
			Target:  tree.AsStringWithFQNames(&zs, params.Ann()),
			Config:  strings.TrimSpace(yamlConfig),
			Options: optionsStr,
		}
		var info eventpb.EventPayload
		if deleteZone {
			__antithesis_instrumentation__.Notify(622535)
			info = &eventpb.RemoveZoneConfig{CommonZoneConfigDetails: eventDetails}
		} else {
			__antithesis_instrumentation__.Notify(622536)
			info = &eventpb.SetZoneConfig{CommonZoneConfigDetails: eventDetails}
		}
		__antithesis_instrumentation__.Notify(622415)
		return params.p.logEvent(params.ctx, targetID, info)
	}
	__antithesis_instrumentation__.Notify(622344)
	for _, zs := range specifiers {
		__antithesis_instrumentation__.Notify(622537)

		if err := applyZoneConfig(zs); err != nil {
			__antithesis_instrumentation__.Notify(622538)
			return err
		} else {
			__antithesis_instrumentation__.Notify(622539)
		}
	}
	__antithesis_instrumentation__.Notify(622345)
	return nil
}

func (n *setZoneConfigNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(622540)
	return false, nil
}
func (n *setZoneConfigNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(622541)
	return nil
}
func (*setZoneConfigNode) Close(context.Context) { __antithesis_instrumentation__.Notify(622542) }

func (n *setZoneConfigNode) FastPathResults() (int, bool) {
	__antithesis_instrumentation__.Notify(622543)
	return n.run.numAffected, true
}

type nodeGetter func(context.Context, *serverpb.NodesRequest) (*serverpb.NodesResponse, error)
type regionsGetter func(context.Context, *serverpb.RegionsRequest) (*serverpb.RegionsResponse, error)

func validateNoRepeatKeysInZone(zone *zonepb.ZoneConfig) error {
	__antithesis_instrumentation__.Notify(622544)
	if err := validateNoRepeatKeysInConjunction(zone.Constraints); err != nil {
		__antithesis_instrumentation__.Notify(622546)
		return err
	} else {
		__antithesis_instrumentation__.Notify(622547)
	}
	__antithesis_instrumentation__.Notify(622545)
	return validateNoRepeatKeysInConjunction(zone.VoterConstraints)
}

func validateNoRepeatKeysInConjunction(conjunctions []zonepb.ConstraintsConjunction) error {
	__antithesis_instrumentation__.Notify(622548)
	for _, constraints := range conjunctions {
		__antithesis_instrumentation__.Notify(622550)

		for i, curr := range constraints.Constraints {
			__antithesis_instrumentation__.Notify(622551)
			for _, other := range constraints.Constraints[i+1:] {
				__antithesis_instrumentation__.Notify(622552)

				if curr.Key == "" && func() bool {
					__antithesis_instrumentation__.Notify(622553)
					return other.Key == "" == true
				}() == true {
					__antithesis_instrumentation__.Notify(622554)
					if curr.Value == other.Value {
						__antithesis_instrumentation__.Notify(622555)
						return pgerror.Newf(pgcode.CheckViolation,
							"incompatible zone constraints: %q and %q", curr, other)
					} else {
						__antithesis_instrumentation__.Notify(622556)
					}
				} else {
					__antithesis_instrumentation__.Notify(622557)
					if curr.Type == zonepb.Constraint_REQUIRED {
						__antithesis_instrumentation__.Notify(622558)
						if other.Type == zonepb.Constraint_REQUIRED && func() bool {
							__antithesis_instrumentation__.Notify(622559)
							return other.Key == curr.Key == true
						}() == true || func() bool {
							__antithesis_instrumentation__.Notify(622560)
							return (other.Type == zonepb.Constraint_PROHIBITED && func() bool {
								__antithesis_instrumentation__.Notify(622561)
								return other.Key == curr.Key == true
							}() == true && func() bool {
								__antithesis_instrumentation__.Notify(622562)
								return other.Value == curr.Value == true
							}() == true) == true
						}() == true {
							__antithesis_instrumentation__.Notify(622563)
							return pgerror.Newf(pgcode.CheckViolation,
								"incompatible zone constraints: %q and %q", curr, other)
						} else {
							__antithesis_instrumentation__.Notify(622564)
						}
					} else {
						__antithesis_instrumentation__.Notify(622565)
						if curr.Type == zonepb.Constraint_PROHIBITED {
							__antithesis_instrumentation__.Notify(622566)

							if other.Type == zonepb.Constraint_REQUIRED && func() bool {
								__antithesis_instrumentation__.Notify(622567)
								return other.Key == curr.Key == true
							}() == true && func() bool {
								__antithesis_instrumentation__.Notify(622568)
								return other.Value == curr.Value == true
							}() == true {
								__antithesis_instrumentation__.Notify(622569)
								return pgerror.Newf(pgcode.CheckViolation,
									"incompatible zone constraints: %q and %q", curr, other)
							} else {
								__antithesis_instrumentation__.Notify(622570)
							}
						} else {
							__antithesis_instrumentation__.Notify(622571)
						}
					}
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(622549)
	return nil
}

func accumulateUniqueConstraints(zone *zonepb.ZoneConfig) []zonepb.Constraint {
	__antithesis_instrumentation__.Notify(622572)
	constraints := make([]zonepb.Constraint, 0)
	addToValidate := func(c zonepb.Constraint) {
		__antithesis_instrumentation__.Notify(622577)
		var alreadyInList bool
		for _, val := range constraints {
			__antithesis_instrumentation__.Notify(622579)
			if c == val {
				__antithesis_instrumentation__.Notify(622580)
				alreadyInList = true
				break
			} else {
				__antithesis_instrumentation__.Notify(622581)
			}
		}
		__antithesis_instrumentation__.Notify(622578)
		if !alreadyInList {
			__antithesis_instrumentation__.Notify(622582)
			constraints = append(constraints, c)
		} else {
			__antithesis_instrumentation__.Notify(622583)
		}
	}
	__antithesis_instrumentation__.Notify(622573)
	for _, constraints := range zone.Constraints {
		__antithesis_instrumentation__.Notify(622584)
		for _, constraint := range constraints.Constraints {
			__antithesis_instrumentation__.Notify(622585)
			addToValidate(constraint)
		}
	}
	__antithesis_instrumentation__.Notify(622574)
	for _, constraints := range zone.VoterConstraints {
		__antithesis_instrumentation__.Notify(622586)
		for _, constraint := range constraints.Constraints {
			__antithesis_instrumentation__.Notify(622587)
			addToValidate(constraint)
		}
	}
	__antithesis_instrumentation__.Notify(622575)
	for _, leasePreferences := range zone.LeasePreferences {
		__antithesis_instrumentation__.Notify(622588)
		for _, constraint := range leasePreferences.Constraints {
			__antithesis_instrumentation__.Notify(622589)
			addToValidate(constraint)
		}
	}
	__antithesis_instrumentation__.Notify(622576)
	return constraints
}

func validateZoneAttrsAndLocalities(
	ctx context.Context, execCfg *ExecutorConfig, zone *zonepb.ZoneConfig,
) error {
	__antithesis_instrumentation__.Notify(622590)

	if len(zone.Constraints) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(622593)
		return len(zone.VoterConstraints) == 0 == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(622594)
		return len(zone.LeasePreferences) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(622595)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(622596)
	}
	__antithesis_instrumentation__.Notify(622591)
	if execCfg.Codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(622597)
		ss, err := execCfg.NodesStatusServer.OptionalNodesStatusServer(MultitenancyZoneCfgIssueNo)
		if err != nil {
			__antithesis_instrumentation__.Notify(622599)
			return err
		} else {
			__antithesis_instrumentation__.Notify(622600)
		}
		__antithesis_instrumentation__.Notify(622598)
		return validateZoneAttrsAndLocalitiesForSystemTenant(ctx, ss.ListNodesInternal, zone)
	} else {
		__antithesis_instrumentation__.Notify(622601)
	}
	__antithesis_instrumentation__.Notify(622592)
	return validateZoneLocalitiesForSecondaryTenants(ctx, execCfg.RegionsServer.Regions, zone)
}

func validateZoneAttrsAndLocalitiesForSystemTenant(
	ctx context.Context, getNodes nodeGetter, zone *zonepb.ZoneConfig,
) error {
	__antithesis_instrumentation__.Notify(622602)
	nodes, err := getNodes(ctx, &serverpb.NodesRequest{})
	if err != nil {
		__antithesis_instrumentation__.Notify(622605)
		return err
	} else {
		__antithesis_instrumentation__.Notify(622606)
	}
	__antithesis_instrumentation__.Notify(622603)

	toValidate := accumulateUniqueConstraints(zone)

	for _, constraint := range toValidate {
		__antithesis_instrumentation__.Notify(622607)

		if constraint.Type == zonepb.Constraint_PROHIBITED {
			__antithesis_instrumentation__.Notify(622610)
			continue
		} else {
			__antithesis_instrumentation__.Notify(622611)
		}
		__antithesis_instrumentation__.Notify(622608)
		var found bool
	node:
		for _, node := range nodes.Nodes {
			__antithesis_instrumentation__.Notify(622612)
			for _, store := range node.StoreStatuses {
				__antithesis_instrumentation__.Notify(622613)

				if zonepb.StoreSatisfiesConstraint(store.Desc, constraint) {
					__antithesis_instrumentation__.Notify(622614)
					found = true
					break node
				} else {
					__antithesis_instrumentation__.Notify(622615)
				}
			}
		}
		__antithesis_instrumentation__.Notify(622609)
		if !found {
			__antithesis_instrumentation__.Notify(622616)
			return pgerror.Newf(pgcode.CheckViolation,
				"constraint %q matches no existing nodes within the cluster - did you enter it correctly?",
				constraint)
		} else {
			__antithesis_instrumentation__.Notify(622617)
		}
	}
	__antithesis_instrumentation__.Notify(622604)

	return nil
}

func validateZoneLocalitiesForSecondaryTenants(
	ctx context.Context, getRegions regionsGetter, zone *zonepb.ZoneConfig,
) error {
	__antithesis_instrumentation__.Notify(622618)
	toValidate := accumulateUniqueConstraints(zone)
	resp, err := getRegions(ctx, &serverpb.RegionsRequest{})
	if err != nil {
		__antithesis_instrumentation__.Notify(622622)
		return err
	} else {
		__antithesis_instrumentation__.Notify(622623)
	}
	__antithesis_instrumentation__.Notify(622619)
	regions := make(map[string]struct{})
	zones := make(map[string]struct{})
	for regionName, regionMeta := range resp.Regions {
		__antithesis_instrumentation__.Notify(622624)
		regions[regionName] = struct{}{}
		for _, zone := range regionMeta.Zones {
			__antithesis_instrumentation__.Notify(622625)
			zones[zone] = struct{}{}
		}
	}
	__antithesis_instrumentation__.Notify(622620)

	for _, constraint := range toValidate {
		__antithesis_instrumentation__.Notify(622626)
		switch constraint.Key {
		case "zone":
			__antithesis_instrumentation__.Notify(622627)
			_, found := zones[constraint.Value]
			if !found {
				__antithesis_instrumentation__.Notify(622630)
				return pgerror.Newf(
					pgcode.CheckViolation,
					"zone %q not found",
					constraint.Value,
				)
			} else {
				__antithesis_instrumentation__.Notify(622631)
			}
		case "region":
			__antithesis_instrumentation__.Notify(622628)
			_, found := regions[constraint.Value]
			if !found {
				__antithesis_instrumentation__.Notify(622632)
				return pgerror.Newf(
					pgcode.CheckViolation,
					"region %q not found",
					constraint.Value,
				)
			} else {
				__antithesis_instrumentation__.Notify(622633)
			}
		default:
			__antithesis_instrumentation__.Notify(622629)
			return errors.WithHint(pgerror.Newf(
				pgcode.CheckViolation,
				"invalid constraint attribute: %q",
				constraint.Key,
			),
				`only "zone" and "region" are allowed`,
			)
		}
	}
	__antithesis_instrumentation__.Notify(622621)
	return nil
}

const MultitenancyZoneCfgIssueNo = 49854

type zoneConfigUpdate struct {
	id    descpb.ID
	value []byte
}

func prepareZoneConfigWrites(
	ctx context.Context,
	execCfg *ExecutorConfig,
	targetID descpb.ID,
	table catalog.TableDescriptor,
	zone *zonepb.ZoneConfig,
	hasNewSubzones bool,
) (_ *zoneConfigUpdate, err error) {
	__antithesis_instrumentation__.Notify(622634)
	if len(zone.Subzones) > 0 {
		__antithesis_instrumentation__.Notify(622638)
		st := execCfg.Settings
		zone.SubzoneSpans, err = GenerateSubzoneSpans(
			st, execCfg.LogicalClusterID(), execCfg.Codec, table, zone.Subzones, hasNewSubzones)
		if err != nil {
			__antithesis_instrumentation__.Notify(622639)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(622640)
		}
	} else {
		__antithesis_instrumentation__.Notify(622641)

		zone.SubzoneSpans = nil
	}
	__antithesis_instrumentation__.Notify(622635)
	if zone.IsSubzonePlaceholder() && func() bool {
		__antithesis_instrumentation__.Notify(622642)
		return len(zone.Subzones) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(622643)
		return &zoneConfigUpdate{id: targetID}, nil
	} else {
		__antithesis_instrumentation__.Notify(622644)
	}
	__antithesis_instrumentation__.Notify(622636)
	buf, err := protoutil.Marshal(zone)
	if err != nil {
		__antithesis_instrumentation__.Notify(622645)
		return nil, pgerror.Wrap(err, pgcode.CheckViolation, "could not marshal zone config")
	} else {
		__antithesis_instrumentation__.Notify(622646)
	}
	__antithesis_instrumentation__.Notify(622637)
	return &zoneConfigUpdate{id: targetID, value: buf}, nil
}

func writeZoneConfig(
	ctx context.Context,
	txn *kv.Txn,
	targetID descpb.ID,
	table catalog.TableDescriptor,
	zone *zonepb.ZoneConfig,
	execCfg *ExecutorConfig,
	hasNewSubzones bool,
) (numAffected int, err error) {
	__antithesis_instrumentation__.Notify(622647)
	update, err := prepareZoneConfigWrites(ctx, execCfg, targetID, table, zone, hasNewSubzones)
	if err != nil {
		__antithesis_instrumentation__.Notify(622649)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(622650)
	}
	__antithesis_instrumentation__.Notify(622648)
	return writeZoneConfigUpdate(ctx, txn, execCfg, update)
}

func writeZoneConfigUpdate(
	ctx context.Context, txn *kv.Txn, execCfg *ExecutorConfig, update *zoneConfigUpdate,
) (numAffected int, _ error) {
	__antithesis_instrumentation__.Notify(622651)
	if update.value == nil {
		__antithesis_instrumentation__.Notify(622653)
		return execCfg.InternalExecutor.Exec(ctx, "delete-zone", txn,
			"DELETE FROM system.zones WHERE id = $1", update.id)
	} else {
		__antithesis_instrumentation__.Notify(622654)
	}
	__antithesis_instrumentation__.Notify(622652)
	return execCfg.InternalExecutor.Exec(ctx, "update-zone", txn,
		"UPSERT INTO system.zones (id, config) VALUES ($1, $2)", update.id, update.value)
}

func getZoneConfigRaw(
	ctx context.Context, txn *kv.Txn, codec keys.SQLCodec, settings *cluster.Settings, id descpb.ID,
) (*zonepb.ZoneConfig, error) {
	__antithesis_instrumentation__.Notify(622655)
	kv, err := txn.Get(ctx, config.MakeZoneKey(codec, id))
	if err != nil {
		__antithesis_instrumentation__.Notify(622659)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(622660)
	}
	__antithesis_instrumentation__.Notify(622656)
	if kv.Value == nil {
		__antithesis_instrumentation__.Notify(622661)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(622662)
	}
	__antithesis_instrumentation__.Notify(622657)
	var zone zonepb.ZoneConfig
	if err := kv.ValueProto(&zone); err != nil {
		__antithesis_instrumentation__.Notify(622663)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(622664)
	}
	__antithesis_instrumentation__.Notify(622658)
	return &zone, nil
}

func getZoneConfigRawBatch(
	ctx context.Context,
	txn *kv.Txn,
	codec keys.SQLCodec,
	settings *cluster.Settings,
	ids []descpb.ID,
) (map[descpb.ID]*zonepb.ZoneConfig, error) {
	__antithesis_instrumentation__.Notify(622665)
	b := txn.NewBatch()
	for _, id := range ids {
		__antithesis_instrumentation__.Notify(622669)
		b.Get(config.MakeZoneKey(codec, id))
	}
	__antithesis_instrumentation__.Notify(622666)
	if err := txn.Run(ctx, b); err != nil {
		__antithesis_instrumentation__.Notify(622670)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(622671)
	}
	__antithesis_instrumentation__.Notify(622667)
	ret := make(map[descpb.ID]*zonepb.ZoneConfig, len(b.Results))
	for idx, r := range b.Results {
		__antithesis_instrumentation__.Notify(622672)
		if r.Err != nil {
			__antithesis_instrumentation__.Notify(622676)
			return nil, r.Err
		} else {
			__antithesis_instrumentation__.Notify(622677)
		}
		__antithesis_instrumentation__.Notify(622673)
		var zone zonepb.ZoneConfig
		row := r.Rows[0]
		if row.Value == nil {
			__antithesis_instrumentation__.Notify(622678)
			continue
		} else {
			__antithesis_instrumentation__.Notify(622679)
		}
		__antithesis_instrumentation__.Notify(622674)
		if err := row.ValueProto(&zone); err != nil {
			__antithesis_instrumentation__.Notify(622680)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(622681)
		}
		__antithesis_instrumentation__.Notify(622675)
		ret[ids[idx]] = &zone
	}
	__antithesis_instrumentation__.Notify(622668)
	return ret, nil
}

func RemoveIndexZoneConfigs(
	ctx context.Context,
	txn *kv.Txn,
	execCfg *ExecutorConfig,
	tableDesc catalog.TableDescriptor,
	indexIDs []uint32,
) error {
	__antithesis_instrumentation__.Notify(622682)
	zone, err := getZoneConfigRaw(ctx, txn, execCfg.Codec, execCfg.Settings, tableDesc.GetID())
	if err != nil {
		__antithesis_instrumentation__.Notify(622687)
		return err
	} else {
		__antithesis_instrumentation__.Notify(622688)
	}
	__antithesis_instrumentation__.Notify(622683)

	if zone == nil {
		__antithesis_instrumentation__.Notify(622689)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(622690)
	}
	__antithesis_instrumentation__.Notify(622684)

	zcRewriteNecessary := false
	for _, indexID := range indexIDs {
		__antithesis_instrumentation__.Notify(622691)
		for _, s := range zone.Subzones {
			__antithesis_instrumentation__.Notify(622692)
			if s.IndexID == indexID {
				__antithesis_instrumentation__.Notify(622693)

				zone.DeleteIndexSubzones(indexID)
				zcRewriteNecessary = true
				break
			} else {
				__antithesis_instrumentation__.Notify(622694)
			}
		}
	}
	__antithesis_instrumentation__.Notify(622685)

	if zcRewriteNecessary {
		__antithesis_instrumentation__.Notify(622695)

		_, err = writeZoneConfig(ctx, txn, tableDesc.GetID(), tableDesc, zone, execCfg, false)
		if err != nil && func() bool {
			__antithesis_instrumentation__.Notify(622696)
			return !sqlerrors.IsCCLRequiredError(err) == true
		}() == true {
			__antithesis_instrumentation__.Notify(622697)
			return err
		} else {
			__antithesis_instrumentation__.Notify(622698)
		}
	} else {
		__antithesis_instrumentation__.Notify(622699)
	}
	__antithesis_instrumentation__.Notify(622686)

	return nil
}
