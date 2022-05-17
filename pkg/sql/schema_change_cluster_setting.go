package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/featureflag"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

var featureSchemaChangeEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"feature.schema_change.enabled",
	"set to true to enable schema changes, false to disable; default is true",
	featureflag.FeatureFlagEnabledDefault,
).WithPublic()

func checkSchemaChangeEnabled(
	ctx context.Context, execCfg *ExecutorConfig, schemaFeatureName string,
) error {
	__antithesis_instrumentation__.Notify(577012)
	if err := featureflag.CheckEnabled(
		ctx,
		execCfg,
		featureSchemaChangeEnabled,
		fmt.Sprintf("%s is part of the schema change category, which", schemaFeatureName),
	); err != nil {
		__antithesis_instrumentation__.Notify(577014)
		return err
	} else {
		__antithesis_instrumentation__.Notify(577015)
	}
	__antithesis_instrumentation__.Notify(577013)
	return nil
}

func (p *planner) CheckFeature(ctx context.Context, featureName tree.SchemaFeatureName) error {
	__antithesis_instrumentation__.Notify(577016)
	return checkSchemaChangeEnabled(ctx, p.ExecCfg(), string(featureName))
}
