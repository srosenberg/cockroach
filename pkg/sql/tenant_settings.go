package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/errors"
)

type alterTenantSetClusterSettingNode struct {
	name     string
	tenantID tree.TypedExpr
	st       *cluster.Settings
	setting  settings.NonMaskedSetting

	value tree.TypedExpr
}

func (p *planner) AlterTenantSetClusterSetting(
	ctx context.Context, n *tree.AlterTenantSetClusterSetting,
) (planNode, error) {
	__antithesis_instrumentation__.Notify(628210)

	if err := p.RequireAdminRole(ctx, "change a tenant cluster setting"); err != nil {
		__antithesis_instrumentation__.Notify(628219)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(628220)
	}
	__antithesis_instrumentation__.Notify(628211)

	if !p.execCfg.Codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(628221)
		return nil, pgerror.Newf(pgcode.InsufficientPrivilege,
			"ALTER TENANT can only be called by system operators")
	} else {
		__antithesis_instrumentation__.Notify(628222)
	}
	__antithesis_instrumentation__.Notify(628212)

	name := strings.ToLower(n.Name)
	st := p.EvalContext().Settings
	v, ok := settings.Lookup(name, settings.LookupForLocalAccess, true)
	if !ok {
		__antithesis_instrumentation__.Notify(628223)
		return nil, errors.Errorf("unknown cluster setting '%s'", name)
	} else {
		__antithesis_instrumentation__.Notify(628224)
	}
	__antithesis_instrumentation__.Notify(628213)

	if v.Class() == settings.SystemOnly {
		__antithesis_instrumentation__.Notify(628225)
		return nil, pgerror.Newf(pgcode.InsufficientPrivilege,
			"%s is a system-only setting and must be set in the admin tenant using SET CLUSTER SETTING", name)
	} else {
		__antithesis_instrumentation__.Notify(628226)
	}
	__antithesis_instrumentation__.Notify(628214)

	var typedTenantID tree.TypedExpr
	if !n.TenantAll {
		__antithesis_instrumentation__.Notify(628227)
		var dummyHelper tree.IndexedVarHelper
		var err error
		typedTenantID, err = p.analyzeExpr(
			ctx, n.TenantID, nil, dummyHelper, types.Int, true, "ALTER TENANT SET CLUSTER SETTING "+name)
		if err != nil {
			__antithesis_instrumentation__.Notify(628228)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(628229)
		}
	} else {
		__antithesis_instrumentation__.Notify(628230)
	}
	__antithesis_instrumentation__.Notify(628215)

	setting, ok := v.(settings.NonMaskedSetting)
	if !ok {
		__antithesis_instrumentation__.Notify(628231)
		return nil, errors.AssertionFailedf("expected writable setting, got %T", v)
	} else {
		__antithesis_instrumentation__.Notify(628232)
	}
	__antithesis_instrumentation__.Notify(628216)

	if _, isVersion := setting.(*settings.VersionSetting); isVersion {
		__antithesis_instrumentation__.Notify(628233)
		return nil, unimplemented.NewWithIssue(77733, "cannot change the version of another tenant")
	} else {
		__antithesis_instrumentation__.Notify(628234)
	}
	__antithesis_instrumentation__.Notify(628217)

	value, err := p.getAndValidateTypedClusterSetting(ctx, name, n.Value, setting)
	if err != nil {
		__antithesis_instrumentation__.Notify(628235)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(628236)
	}
	__antithesis_instrumentation__.Notify(628218)

	node := alterTenantSetClusterSettingNode{
		name: name, tenantID: typedTenantID, st: st,
		setting: setting, value: value,
	}
	return &node, nil
}

func (n *alterTenantSetClusterSettingNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(628237)
	var tenantIDi uint64
	var tenantID tree.Datum
	if n.tenantID == nil {
		__antithesis_instrumentation__.Notify(628240)

		tenantID = tree.NewDInt(0)
	} else {
		__antithesis_instrumentation__.Notify(628241)

		var err error
		tenantIDi, tenantID, err = resolveTenantID(params.p, n.tenantID)
		if err != nil {
			__antithesis_instrumentation__.Notify(628243)
			return err
		} else {
			__antithesis_instrumentation__.Notify(628244)
		}
		__antithesis_instrumentation__.Notify(628242)
		if err := assertTenantExists(params.ctx, params.p, tenantID); err != nil {
			__antithesis_instrumentation__.Notify(628245)
			return err
		} else {
			__antithesis_instrumentation__.Notify(628246)
		}
	}
	__antithesis_instrumentation__.Notify(628238)

	var reportedValue string
	if n.value == nil {
		__antithesis_instrumentation__.Notify(628247)

		reportedValue = "DEFAULT"
		if _, err := params.p.execCfg.InternalExecutor.ExecEx(
			params.ctx, "reset-tenant-setting", params.p.Txn(),
			sessiondata.InternalExecutorOverride{User: security.RootUserName()},
			"DELETE FROM system.tenant_settings WHERE tenant_id = $1 AND name = $2", tenantID, n.name,
		); err != nil {
			__antithesis_instrumentation__.Notify(628248)
			return err
		} else {
			__antithesis_instrumentation__.Notify(628249)
		}
	} else {
		__antithesis_instrumentation__.Notify(628250)
		reportedValue = tree.AsStringWithFlags(n.value, tree.FmtBareStrings)
		value, err := n.value.Eval(params.p.EvalContext())
		if err != nil {
			__antithesis_instrumentation__.Notify(628253)
			return err
		} else {
			__antithesis_instrumentation__.Notify(628254)
		}
		__antithesis_instrumentation__.Notify(628251)
		encoded, err := toSettingString(params.ctx, n.st, n.name, n.setting, value)
		if err != nil {
			__antithesis_instrumentation__.Notify(628255)
			return err
		} else {
			__antithesis_instrumentation__.Notify(628256)
		}
		__antithesis_instrumentation__.Notify(628252)
		if _, err := params.p.execCfg.InternalExecutor.ExecEx(
			params.ctx, "update-tenant-setting", params.p.Txn(),
			sessiondata.InternalExecutorOverride{User: security.RootUserName()},
			`UPSERT INTO system.tenant_settings (tenant_id, name, value, last_updated, value_type) VALUES ($1, $2, $3, now(), $4)`,
			tenantID, n.name, encoded, n.setting.Typ(),
		); err != nil {
			__antithesis_instrumentation__.Notify(628257)
			return err
		} else {
			__antithesis_instrumentation__.Notify(628258)
		}
	}
	__antithesis_instrumentation__.Notify(628239)

	return params.p.logEvent(
		params.ctx,
		0,
		&eventpb.SetTenantClusterSetting{
			SettingName: n.name,
			Value:       reportedValue,
			TenantId:    tenantIDi,
			AllTenants:  tenantIDi == 0,
		})
}

func (n *alterTenantSetClusterSettingNode) Next(_ runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(628259)
	return false, nil
}
func (n *alterTenantSetClusterSettingNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(628260)
	return nil
}
func (n *alterTenantSetClusterSettingNode) Close(_ context.Context) {
	__antithesis_instrumentation__.Notify(628261)
}

func resolveTenantID(p *planner, expr tree.TypedExpr) (uint64, tree.Datum, error) {
	__antithesis_instrumentation__.Notify(628262)
	tenantIDd, err := expr.Eval(p.EvalContext())
	if err != nil {
		__antithesis_instrumentation__.Notify(628267)
		return 0, nil, err
	} else {
		__antithesis_instrumentation__.Notify(628268)
	}
	__antithesis_instrumentation__.Notify(628263)
	tenantID, ok := tenantIDd.(*tree.DInt)
	if !ok {
		__antithesis_instrumentation__.Notify(628269)
		return 0, nil, errors.AssertionFailedf("expected int, got %T", tenantIDd)
	} else {
		__antithesis_instrumentation__.Notify(628270)
	}
	__antithesis_instrumentation__.Notify(628264)
	if *tenantID == 0 {
		__antithesis_instrumentation__.Notify(628271)
		return 0, nil, pgerror.Newf(pgcode.InvalidParameterValue, "tenant ID must be non-zero")
	} else {
		__antithesis_instrumentation__.Notify(628272)
	}
	__antithesis_instrumentation__.Notify(628265)
	if roachpb.MakeTenantID(uint64(*tenantID)) == roachpb.SystemTenantID {
		__antithesis_instrumentation__.Notify(628273)
		return 0, nil, errors.WithHint(pgerror.Newf(pgcode.InvalidParameterValue,
			"cannot use this statement to access cluster settings in system tenant"),
			"Use a regular SHOW/SET CLUSTER SETTING statement.")
	} else {
		__antithesis_instrumentation__.Notify(628274)
	}
	__antithesis_instrumentation__.Notify(628266)
	return uint64(*tenantID), tenantIDd, nil
}

func assertTenantExists(ctx context.Context, p *planner, tenantID tree.Datum) error {
	__antithesis_instrumentation__.Notify(628275)
	exists, err := p.ExecCfg().InternalExecutor.QueryRowEx(
		ctx, "get-tenant", p.txn,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		`SELECT EXISTS(SELECT id FROM system.tenants WHERE id = $1)`, tenantID)
	if err != nil {
		__antithesis_instrumentation__.Notify(628278)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628279)
	}
	__antithesis_instrumentation__.Notify(628276)
	if exists[0] != tree.DBoolTrue {
		__antithesis_instrumentation__.Notify(628280)
		return pgerror.Newf(pgcode.InvalidParameterValue, "no tenant found with ID %v", tenantID)
	} else {
		__antithesis_instrumentation__.Notify(628281)
	}
	__antithesis_instrumentation__.Notify(628277)
	return nil
}

func (p *planner) ShowTenantClusterSetting(
	ctx context.Context, n *tree.ShowTenantClusterSetting,
) (planNode, error) {
	__antithesis_instrumentation__.Notify(628282)

	if err := p.RequireAdminRole(ctx, "view a tenant cluster setting"); err != nil {
		__antithesis_instrumentation__.Notify(628288)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(628289)
	}
	__antithesis_instrumentation__.Notify(628283)

	name := strings.ToLower(n.Name)
	val, ok := settings.Lookup(
		name, settings.LookupForLocalAccess, p.ExecCfg().Codec.ForSystemTenant(),
	)
	if !ok {
		__antithesis_instrumentation__.Notify(628290)
		return nil, errors.Errorf("unknown setting: %q", name)
	} else {
		__antithesis_instrumentation__.Notify(628291)
	}
	__antithesis_instrumentation__.Notify(628284)
	setting, ok := val.(settings.NonMaskedSetting)
	if !ok {
		__antithesis_instrumentation__.Notify(628292)

		return nil, errors.AssertionFailedf("setting is masked: %v", name)
	} else {
		__antithesis_instrumentation__.Notify(628293)
	}
	__antithesis_instrumentation__.Notify(628285)

	if !p.execCfg.Codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(628294)
		return nil, pgerror.Newf(pgcode.InsufficientPrivilege,
			"SHOW CLUSTER SETTING FOR TENANT can only be called by system operators")
	} else {
		__antithesis_instrumentation__.Notify(628295)
	}
	__antithesis_instrumentation__.Notify(628286)

	columns, err := getShowClusterSettingPlanColumns(setting, name)
	if err != nil {
		__antithesis_instrumentation__.Notify(628296)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(628297)
	}
	__antithesis_instrumentation__.Notify(628287)

	return planShowClusterSetting(
		setting, name, columns,
		func(ctx context.Context, p *planner) (bool, string, error) {
			__antithesis_instrumentation__.Notify(628298)

			lookupEncodedTenantSetting := `
WITH
  tenant_id AS (SELECT (` + n.TenantID.String() + `):::INT AS tenant_id),
  isvalid AS (
    SELECT
      CASE
       WHEN tenant_id=0 THEN
         crdb_internal.force_error('22023', 'tenant ID must be non-zero')
       WHEN tenant_id=1 THEN
         crdb_internal.force_error('22023', 'use SHOW CLUSTER SETTING to display a setting for the system tenant')
       WHEN st.id IS NULL THEN
         crdb_internal.force_error('22023', 'no tenant found with ID '||tenant_id)
       ELSE 0
      END AS ok
    FROM      tenant_id
    LEFT JOIN system.tenants st ON id = tenant_id.tenant_id
  ),
  tenantspecific AS (
     SELECT t.name, t.value
     FROM system.tenant_settings t, tenant_id
     WHERE t.tenant_id = tenant_id.tenant_id
  ),
  setting AS (
   SELECT $1 AS variable
  )
SELECT COALESCE(
   tenantspecific.value,
   overrideall.value,
   -- NB: we can't compute the actual value here, see discussion on issue #77935.
   NULL
   )
FROM
  setting
  LEFT JOIN tenantspecific
         ON setting.variable = tenantspecific.name
  LEFT JOIN system.tenant_settings AS overrideall
         ON setting.variable = overrideall.name AND overrideall.tenant_id = 0`

			datums, err := p.ExecCfg().InternalExecutor.QueryRowEx(
				ctx, "get-tenant-setting-value", p.txn,
				sessiondata.InternalExecutorOverride{User: security.RootUserName()},
				lookupEncodedTenantSetting,
				name)
			if err != nil {
				__antithesis_instrumentation__.Notify(628303)
				return false, "", err
			} else {
				__antithesis_instrumentation__.Notify(628304)
			}
			__antithesis_instrumentation__.Notify(628299)
			if len(datums) != 1 {
				__antithesis_instrumentation__.Notify(628305)
				return false, "", errors.AssertionFailedf("expected 1 column, got %+v", datums)
			} else {
				__antithesis_instrumentation__.Notify(628306)
			}
			__antithesis_instrumentation__.Notify(628300)
			if datums[0] == tree.DNull {
				__antithesis_instrumentation__.Notify(628307)
				return false, "", nil
			} else {
				__antithesis_instrumentation__.Notify(628308)
			}
			__antithesis_instrumentation__.Notify(628301)
			encoded, ok := tree.AsDString(datums[0])
			if !ok {
				__antithesis_instrumentation__.Notify(628309)
				return false, "", errors.AssertionFailedf("expected string value, got %T", datums[0])
			} else {
				__antithesis_instrumentation__.Notify(628310)
			}
			__antithesis_instrumentation__.Notify(628302)
			return true, string(encoded), nil
		})
}
