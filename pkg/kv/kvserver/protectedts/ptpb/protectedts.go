package ptpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
)

func MakeClusterTarget() *Target {
	__antithesis_instrumentation__.Notify(110481)
	return &Target{Union: &Target_Cluster{Cluster: &Target_ClusterTarget{}}}
}

func MakeTenantsTarget(ids []roachpb.TenantID) *Target {
	__antithesis_instrumentation__.Notify(110482)
	return &Target{Union: &Target_Tenants{Tenants: &Target_TenantsTarget{IDs: ids}}}
}

func MakeSchemaObjectsTarget(ids descpb.IDs) *Target {
	__antithesis_instrumentation__.Notify(110483)
	return &Target{Union: &Target_SchemaObjects{SchemaObjects: &Target_SchemaObjectsTarget{IDs: ids}}}
}
