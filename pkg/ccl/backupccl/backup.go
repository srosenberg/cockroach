package backupccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	descpb "github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/protoreflect"
)

func (m *BackupManifest) GetTenants() []descpb.TenantInfoWithUsage {
	__antithesis_instrumentation__.Notify(4132)
	if len(m.Tenants) > 0 {
		__antithesis_instrumentation__.Notify(4135)
		return m.Tenants
	} else {
		__antithesis_instrumentation__.Notify(4136)
	}
	__antithesis_instrumentation__.Notify(4133)
	if len(m.TenantsDeprecated) > 0 {
		__antithesis_instrumentation__.Notify(4137)
		res := make([]descpb.TenantInfoWithUsage, len(m.TenantsDeprecated))
		for i := range res {
			__antithesis_instrumentation__.Notify(4139)
			res[i].TenantInfo = m.TenantsDeprecated[i]
		}
		__antithesis_instrumentation__.Notify(4138)
		return res
	} else {
		__antithesis_instrumentation__.Notify(4140)
	}
	__antithesis_instrumentation__.Notify(4134)
	return nil
}

func (m *BackupManifest) HasTenants() bool {
	__antithesis_instrumentation__.Notify(4141)
	return len(m.Tenants) > 0 || func() bool {
		__antithesis_instrumentation__.Notify(4142)
		return len(m.TenantsDeprecated) > 0 == true
	}() == true
}

func init() {
	protoreflect.RegisterShorthands((*BackupManifest)(nil), "backup", "backup_manifest")
}
