package backupccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
)

type restorationData interface {
	getSpans() []roachpb.Span

	getSystemTables() []catalog.TableDescriptor

	getRekeys() []execinfrapb.TableRekey
	getTenantRekeys() []execinfrapb.TenantRekey
	getPKIDs() map[uint64]bool

	addTenant(fromID, toID roachpb.TenantID)

	isEmpty() bool

	isMainBundle() bool
}

type mainRestorationData struct {
	restorationDataBase
}

var _ restorationData = &mainRestorationData{}

func (*mainRestorationData) isMainBundle() bool {
	__antithesis_instrumentation__.Notify(10345)
	return true
}

type restorationDataBase struct {
	spans []roachpb.Span

	tableRekeys []execinfrapb.TableRekey

	tenantRekeys []execinfrapb.TenantRekey

	pkIDs map[uint64]bool

	systemTables []catalog.TableDescriptor
}

var _ restorationData = &restorationDataBase{}

func (b *restorationDataBase) getRekeys() []execinfrapb.TableRekey {
	__antithesis_instrumentation__.Notify(10346)
	return b.tableRekeys
}

func (b *restorationDataBase) getTenantRekeys() []execinfrapb.TenantRekey {
	__antithesis_instrumentation__.Notify(10347)
	return b.tenantRekeys
}

func (b *restorationDataBase) getPKIDs() map[uint64]bool {
	__antithesis_instrumentation__.Notify(10348)
	return b.pkIDs
}

func (b *restorationDataBase) getSpans() []roachpb.Span {
	__antithesis_instrumentation__.Notify(10349)
	return b.spans
}

func (b *restorationDataBase) getSystemTables() []catalog.TableDescriptor {
	__antithesis_instrumentation__.Notify(10350)
	return b.systemTables
}

func (b *restorationDataBase) addTenant(fromTenantID, toTenantID roachpb.TenantID) {
	__antithesis_instrumentation__.Notify(10351)
	prefix := keys.MakeTenantPrefix(fromTenantID)
	b.spans = append(b.spans, roachpb.Span{Key: prefix, EndKey: prefix.PrefixEnd()})
	b.tenantRekeys = append(b.tenantRekeys, execinfrapb.TenantRekey{
		OldID: fromTenantID,
		NewID: toTenantID,
	})
}

func (b *restorationDataBase) isEmpty() bool {
	__antithesis_instrumentation__.Notify(10352)
	return len(b.spans) == 0
}

func (restorationDataBase) isMainBundle() bool {
	__antithesis_instrumentation__.Notify(10353)
	return false
}

func checkForMigratedData(details jobspb.RestoreDetails, dataToRestore restorationData) bool {
	__antithesis_instrumentation__.Notify(10354)
	for _, systemTable := range dataToRestore.getSystemTables() {
		__antithesis_instrumentation__.Notify(10356)

		if _, ok := details.SystemTablesMigrated[systemTable.GetName()]; ok {
			__antithesis_instrumentation__.Notify(10357)
			return true
		} else {
			__antithesis_instrumentation__.Notify(10358)
		}
	}
	__antithesis_instrumentation__.Notify(10355)

	return false
}
