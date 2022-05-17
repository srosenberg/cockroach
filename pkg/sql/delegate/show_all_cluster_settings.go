package delegate

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/roleoption"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

func (d *delegator) delegateShowClusterSettingList(
	stmt *tree.ShowClusterSettingList,
) (tree.Statement, error) {
	__antithesis_instrumentation__.Notify(465449)
	isAdmin, err := d.catalog.HasAdminRole(d.ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(465455)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(465456)
	}
	__antithesis_instrumentation__.Notify(465450)
	hasModify, err := d.catalog.HasRoleOption(d.ctx, roleoption.MODIFYCLUSTERSETTING)
	if err != nil {
		__antithesis_instrumentation__.Notify(465457)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(465458)
	}
	__antithesis_instrumentation__.Notify(465451)
	hasView, err := d.catalog.HasRoleOption(d.ctx, roleoption.VIEWCLUSTERSETTING)
	if err != nil {
		__antithesis_instrumentation__.Notify(465459)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(465460)
	}
	__antithesis_instrumentation__.Notify(465452)
	if !hasModify && func() bool {
		__antithesis_instrumentation__.Notify(465461)
		return !hasView == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(465462)
		return !isAdmin == true
	}() == true {
		__antithesis_instrumentation__.Notify(465463)
		return nil, pgerror.Newf(pgcode.InsufficientPrivilege,
			"only users with either %s or %s privileges are allowed to SHOW CLUSTER SETTINGS",
			roleoption.MODIFYCLUSTERSETTING, roleoption.VIEWCLUSTERSETTING)
	} else {
		__antithesis_instrumentation__.Notify(465464)
	}
	__antithesis_instrumentation__.Notify(465453)
	if stmt.All {
		__antithesis_instrumentation__.Notify(465465)
		return parse(
			`SELECT variable, value, type AS setting_type, public, description
       FROM   crdb_internal.cluster_settings`,
		)
	} else {
		__antithesis_instrumentation__.Notify(465466)
	}
	__antithesis_instrumentation__.Notify(465454)
	return parse(
		`SELECT variable, value, type AS setting_type, description
     FROM   crdb_internal.cluster_settings
     WHERE  public IS TRUE`,
	)
}

func (d *delegator) delegateShowTenantClusterSettingList(
	stmt *tree.ShowTenantClusterSettingList,
) (tree.Statement, error) {
	__antithesis_instrumentation__.Notify(465467)

	if err := d.catalog.RequireAdminRole(d.ctx, "show a tenant cluster setting"); err != nil {
		__antithesis_instrumentation__.Notify(465471)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(465472)
	}
	__antithesis_instrumentation__.Notify(465468)

	if !d.evalCtx.Codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(465473)
		return nil, pgerror.Newf(pgcode.InsufficientPrivilege,
			"SHOW CLUSTER SETTINGS FOR TENANT can only be called by system operators")
	} else {
		__antithesis_instrumentation__.Notify(465474)
	}
	__antithesis_instrumentation__.Notify(465469)

	publicCol := `allsettings.public,`
	var publicFilter string
	if !stmt.All {
		__antithesis_instrumentation__.Notify(465475)
		publicCol = ``
		publicFilter = `WHERE public IS TRUE`
	} else {
		__antithesis_instrumentation__.Notify(465476)
	}
	__antithesis_instrumentation__.Notify(465470)

	return parse(`
WITH
  tenant_id AS (SELECT (` + stmt.TenantID.String() + `):::INT AS tenant_id),
  isvalid AS (
    SELECT
      CASE
       WHEN tenant_id=0 THEN
         crdb_internal.force_error('22023', 'tenant ID must be non-zero')
       WHEN tenant_id=1 THEN
         crdb_internal.force_error('22023', 'use SHOW CLUSTER SETTINGS to display settings for the system tenant')
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
  allsettings AS (
    SELECT variable, value, public, type, description
    FROM system.crdb_internal.cluster_settings ` + publicFilter + `
  )
SELECT
  allsettings.variable || substr('', (SELECT ok FROM isvalid)) AS variable,
  crdb_internal.decode_cluster_setting(allsettings.variable,
     -- NB: careful not to coalesce with allsettings.value directly!
     -- This is the value for the system tenant and is not relevant to other tenants.
     COALESCE(tenantspecific.value,
              overrideall.value,
              -- NB: we can't compute the actual value here, which is the entry in the tenant's settings table.
              -- See discussion on issue #77935.
              NULL)
  ) AS value,
  allsettings.type,
  ` + publicCol + `
  CASE
    WHEN tenantspecific.value IS NOT NULL THEN 'per-tenant-override'
    WHEN overrideall.value IS NOT NULL THEN 'all-tenants-override'
    ELSE 'no-override'
  END AS origin,
  allsettings.description
FROM
  allsettings
  LEFT JOIN tenantspecific ON
                  allsettings.variable = tenantspecific.name
  LEFT JOIN system.tenant_settings AS overrideall ON
                  allsettings.variable = overrideall.name AND overrideall.tenant_id = 0
`)
}
