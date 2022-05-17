package spanconfig

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts/ptpb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
)

type ProtectedTimestampStateReader struct {
	schemaObjectProtections map[descpb.ID][]roachpb.ProtectionPolicy
	tenantProtections       []TenantProtectedTimestamps
	clusterProtections      []roachpb.ProtectionPolicy
}

func NewProtectedTimestampStateReader(
	_ context.Context, ptsState ptpb.State,
) *ProtectedTimestampStateReader {
	__antithesis_instrumentation__.Notify(240256)
	reader := &ProtectedTimestampStateReader{
		schemaObjectProtections: make(map[descpb.ID][]roachpb.ProtectionPolicy),
		tenantProtections:       make([]TenantProtectedTimestamps, 0),
		clusterProtections:      make([]roachpb.ProtectionPolicy, 0),
	}
	reader.loadProtectedTimestampRecords(ptsState)
	return reader
}

func (p *ProtectedTimestampStateReader) GetProtectionPoliciesForCluster() []roachpb.ProtectionPolicy {
	__antithesis_instrumentation__.Notify(240257)
	return p.clusterProtections
}

type TenantProtectedTimestamps struct {
	Protections []roachpb.ProtectionPolicy
	TenantID    roachpb.TenantID
}

func (t *TenantProtectedTimestamps) GetTenantProtections() []roachpb.ProtectionPolicy {
	__antithesis_instrumentation__.Notify(240258)
	return t.Protections
}

func (t *TenantProtectedTimestamps) GetTenantID() roachpb.TenantID {
	__antithesis_instrumentation__.Notify(240259)
	return t.TenantID
}

func (p *ProtectedTimestampStateReader) GetProtectionPoliciesForTenants() []TenantProtectedTimestamps {
	__antithesis_instrumentation__.Notify(240260)
	return p.tenantProtections
}

func (p *ProtectedTimestampStateReader) GetProtectionsForTenant(
	tenantID roachpb.TenantID,
) []roachpb.ProtectionPolicy {
	__antithesis_instrumentation__.Notify(240261)
	protectionsOnTenant := make([]roachpb.ProtectionPolicy, 0)
	for _, tp := range p.tenantProtections {
		__antithesis_instrumentation__.Notify(240263)
		if tp.TenantID.Equal(tenantID) {
			__antithesis_instrumentation__.Notify(240264)
			protectionsOnTenant = append(protectionsOnTenant, tp.Protections...)
		} else {
			__antithesis_instrumentation__.Notify(240265)
		}
	}
	__antithesis_instrumentation__.Notify(240262)
	return protectionsOnTenant
}

func (p *ProtectedTimestampStateReader) GetProtectionPoliciesForSchemaObject(
	descID descpb.ID,
) []roachpb.ProtectionPolicy {
	__antithesis_instrumentation__.Notify(240266)
	return p.schemaObjectProtections[descID]
}

func (p *ProtectedTimestampStateReader) loadProtectedTimestampRecords(ptsState ptpb.State) {
	__antithesis_instrumentation__.Notify(240267)
	tenantProtections := make(map[roachpb.TenantID][]roachpb.ProtectionPolicy)
	for _, record := range ptsState.Records {
		__antithesis_instrumentation__.Notify(240269)

		if record.Target == nil {
			__antithesis_instrumentation__.Notify(240271)
			continue
		} else {
			__antithesis_instrumentation__.Notify(240272)
		}
		__antithesis_instrumentation__.Notify(240270)
		protectionPolicy := roachpb.ProtectionPolicy{
			ProtectedTimestamp:         record.Timestamp,
			IgnoreIfExcludedFromBackup: record.Target.IgnoreIfExcludedFromBackup,
		}
		switch t := record.Target.GetUnion().(type) {
		case *ptpb.Target_Cluster:
			__antithesis_instrumentation__.Notify(240273)
			p.clusterProtections = append(p.clusterProtections, protectionPolicy)
		case *ptpb.Target_Tenants:
			__antithesis_instrumentation__.Notify(240274)
			for _, tenID := range t.Tenants.IDs {
				__antithesis_instrumentation__.Notify(240276)
				tenantProtections[tenID] = append(tenantProtections[tenID], protectionPolicy)
			}
		case *ptpb.Target_SchemaObjects:
			__antithesis_instrumentation__.Notify(240275)
			for _, descID := range t.SchemaObjects.IDs {
				__antithesis_instrumentation__.Notify(240277)
				p.schemaObjectProtections[descID] = append(p.schemaObjectProtections[descID], protectionPolicy)
			}
		}
	}
	__antithesis_instrumentation__.Notify(240268)

	for tenID, tenantProtections := range tenantProtections {
		__antithesis_instrumentation__.Notify(240278)
		p.tenantProtections = append(p.tenantProtections,
			TenantProtectedTimestamps{TenantID: tenID, Protections: tenantProtections})
	}
}
