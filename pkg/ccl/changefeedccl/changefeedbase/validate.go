package changefeedbase

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/errors"
)

func ValidateTable(
	targets []jobspb.ChangefeedTargetSpecification,
	tableDesc catalog.TableDescriptor,
	opts map[string]string,
) error {
	__antithesis_instrumentation__.Notify(16634)
	var found bool
	for _, cts := range targets {
		__antithesis_instrumentation__.Notify(16637)
		var t jobspb.ChangefeedTargetSpecification
		if cts.TableID == tableDesc.GetID() {
			__antithesis_instrumentation__.Notify(16645)
			t = cts
			found = true
		} else {
			__antithesis_instrumentation__.Notify(16646)
			continue
		}
		__antithesis_instrumentation__.Notify(16638)

		if catalog.IsSystemDescriptor(tableDesc) {
			__antithesis_instrumentation__.Notify(16647)
			return errors.Errorf(`CHANGEFEEDs are not supported on system tables`)
		} else {
			__antithesis_instrumentation__.Notify(16648)
		}
		__antithesis_instrumentation__.Notify(16639)
		if tableDesc.IsView() {
			__antithesis_instrumentation__.Notify(16649)
			return errors.Errorf(`CHANGEFEED cannot target views: %s`, tableDesc.GetName())
		} else {
			__antithesis_instrumentation__.Notify(16650)
		}
		__antithesis_instrumentation__.Notify(16640)
		if tableDesc.IsVirtualTable() {
			__antithesis_instrumentation__.Notify(16651)
			return errors.Errorf(`CHANGEFEED cannot target virtual tables: %s`, tableDesc.GetName())
		} else {
			__antithesis_instrumentation__.Notify(16652)
		}
		__antithesis_instrumentation__.Notify(16641)
		if tableDesc.IsSequence() {
			__antithesis_instrumentation__.Notify(16653)
			return errors.Errorf(`CHANGEFEED cannot target sequences: %s`, tableDesc.GetName())
		} else {
			__antithesis_instrumentation__.Notify(16654)
		}
		__antithesis_instrumentation__.Notify(16642)
		switch t.Type {
		case jobspb.ChangefeedTargetSpecification_PRIMARY_FAMILY_ONLY:
			__antithesis_instrumentation__.Notify(16655)
			if len(tableDesc.GetFamilies()) != 1 {
				__antithesis_instrumentation__.Notify(16660)
				return errors.Errorf(
					`CHANGEFEED created on a table with a single column family (%s) cannot now target a table with %d families. targets: %+v, tableDesc: %+v`,
					tableDesc.GetName(), len(tableDesc.GetFamilies()), targets, tableDesc)
			} else {
				__antithesis_instrumentation__.Notify(16661)
			}
		case jobspb.ChangefeedTargetSpecification_EACH_FAMILY:
			__antithesis_instrumentation__.Notify(16656)
			_, columnFamiliesOpt := opts[OptSplitColumnFamilies]
			if !columnFamiliesOpt && func() bool {
				__antithesis_instrumentation__.Notify(16662)
				return len(tableDesc.GetFamilies()) != 1 == true
			}() == true {
				__antithesis_instrumentation__.Notify(16663)
				return errors.Errorf(
					`CHANGEFEED targeting a table (%s) with multiple column families requires WITH %s and will emit multiple events per row.`,
					tableDesc.GetName(), OptSplitColumnFamilies)
			} else {
				__antithesis_instrumentation__.Notify(16664)
			}
		case jobspb.ChangefeedTargetSpecification_COLUMN_FAMILY:
			__antithesis_instrumentation__.Notify(16657)
			cols := 0
			for _, family := range tableDesc.GetFamilies() {
				__antithesis_instrumentation__.Notify(16665)
				if family.Name == t.FamilyName {
					__antithesis_instrumentation__.Notify(16666)
					cols = len(family.ColumnIDs)
					break
				} else {
					__antithesis_instrumentation__.Notify(16667)
				}
			}
			__antithesis_instrumentation__.Notify(16658)
			if cols == 0 {
				__antithesis_instrumentation__.Notify(16668)
				return errors.Errorf("CHANGEFEED targeting nonexistent or removed column family %s of table %s", t.FamilyName, tableDesc.GetName())
			} else {
				__antithesis_instrumentation__.Notify(16669)
			}
		default:
			__antithesis_instrumentation__.Notify(16659)
		}
		__antithesis_instrumentation__.Notify(16643)

		if tableDesc.Dropped() {
			__antithesis_instrumentation__.Notify(16670)
			return errors.Errorf(`"%s" was dropped`, t.StatementTimeName)
		} else {
			__antithesis_instrumentation__.Notify(16671)
		}
		__antithesis_instrumentation__.Notify(16644)

		if tableDesc.Offline() {
			__antithesis_instrumentation__.Notify(16672)
			return errors.Errorf("CHANGEFEED cannot target offline table: %s (offline reason: %q)", tableDesc.GetName(), tableDesc.GetOfflineReason())
		} else {
			__antithesis_instrumentation__.Notify(16673)
		}
	}
	__antithesis_instrumentation__.Notify(16635)
	if !found {
		__antithesis_instrumentation__.Notify(16674)
		return errors.Errorf(`unwatched table: %s`, tableDesc.GetName())
	} else {
		__antithesis_instrumentation__.Notify(16675)
	}
	__antithesis_instrumentation__.Notify(16636)

	return nil
}

func WarningsForTable(
	targets jobspb.ChangefeedTargets, tableDesc catalog.TableDescriptor, opts map[string]string,
) []error {
	__antithesis_instrumentation__.Notify(16676)
	warnings := []error{}
	if _, ok := opts[OptVirtualColumns]; !ok {
		__antithesis_instrumentation__.Notify(16679)
		for _, col := range tableDesc.AccessibleColumns() {
			__antithesis_instrumentation__.Notify(16680)
			if col.IsVirtual() {
				__antithesis_instrumentation__.Notify(16681)
				warnings = append(warnings,
					errors.Errorf("Changefeeds will filter out values for virtual column %s in table %s", col.ColName(), tableDesc.GetName()),
				)
			} else {
				__antithesis_instrumentation__.Notify(16682)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(16683)
	}
	__antithesis_instrumentation__.Notify(16677)
	if tableDesc.NumFamilies() > 1 {
		__antithesis_instrumentation__.Notify(16684)
		warnings = append(warnings,
			errors.Errorf("Table %s has %d underlying column families. Messages will be emitted separately for each family specified, or each family if none specified.",
				tableDesc.GetName(), tableDesc.NumFamilies(),
			))
	} else {
		__antithesis_instrumentation__.Notify(16685)
	}
	__antithesis_instrumentation__.Notify(16678)
	return warnings
}
