package schemafeed

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/changefeedbase"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/errors"
)

type tableEventType uint64

const (
	tableEventTypeUnknown             tableEventType = 0
	tableEventTypeAddColumnNoBackfill tableEventType = 1 << (iota - 1)
	tableEventTypeAddColumnWithBackfill
	tableEventTypeDropColumn
	tableEventTruncate
	tableEventPrimaryKeyChange
	tableEventLocalityRegionalByRowChange
	tableEventAddHiddenColumn
)

var (
	defaultTableEventFilter = tableEventFilter{
		tableEventTypeDropColumn:              false,
		tableEventTypeAddColumnWithBackfill:   false,
		tableEventTypeAddColumnNoBackfill:     true,
		tableEventTypeUnknown:                 true,
		tableEventPrimaryKeyChange:            false,
		tableEventLocalityRegionalByRowChange: false,
		tableEventAddHiddenColumn:             true,
	}

	columnChangeTableEventFilter = tableEventFilter{
		tableEventTypeDropColumn:              false,
		tableEventTypeAddColumnWithBackfill:   false,
		tableEventTypeAddColumnNoBackfill:     false,
		tableEventTypeUnknown:                 true,
		tableEventPrimaryKeyChange:            false,
		tableEventLocalityRegionalByRowChange: false,
		tableEventAddHiddenColumn:             true,
	}

	schemaChangeEventFilters = map[changefeedbase.SchemaChangeEventClass]tableEventFilter{
		changefeedbase.OptSchemaChangeEventClassDefault:      defaultTableEventFilter,
		changefeedbase.OptSchemaChangeEventClassColumnChange: columnChangeTableEventFilter,
	}
)

func (e tableEventType) Contains(event tableEventType) bool {
	__antithesis_instrumentation__.Notify(17966)
	return e&event == event
}

func (e tableEventType) Clear(event tableEventType) tableEventType {
	__antithesis_instrumentation__.Notify(17967)
	return e & (^event)
}

func classifyTableEvent(e TableEvent) tableEventType {
	__antithesis_instrumentation__.Notify(17968)
	et := tableEventTypeUnknown
	if primaryKeyChanged(e) {
		__antithesis_instrumentation__.Notify(17976)
		et = et | tableEventPrimaryKeyChange
	} else {
		__antithesis_instrumentation__.Notify(17977)
	}
	__antithesis_instrumentation__.Notify(17969)

	if newVisibleColumnBackfillComplete(e) {
		__antithesis_instrumentation__.Notify(17978)
		et = et | tableEventTypeAddColumnWithBackfill
	} else {
		__antithesis_instrumentation__.Notify(17979)
	}
	__antithesis_instrumentation__.Notify(17970)

	if newHiddenColumnBackfillComplete(e) {
		__antithesis_instrumentation__.Notify(17980)
		et = et | tableEventAddHiddenColumn
	} else {
		__antithesis_instrumentation__.Notify(17981)
	}
	__antithesis_instrumentation__.Notify(17971)

	if newColumnNoBackfill(e) {
		__antithesis_instrumentation__.Notify(17982)
		et = et | tableEventTypeAddColumnNoBackfill
	} else {
		__antithesis_instrumentation__.Notify(17983)
	}
	__antithesis_instrumentation__.Notify(17972)

	if hasNewColumnDropBackfillMutation(e) {
		__antithesis_instrumentation__.Notify(17984)
		et = et | tableEventTypeDropColumn
	} else {
		__antithesis_instrumentation__.Notify(17985)
	}
	__antithesis_instrumentation__.Notify(17973)

	if tableTruncated(e) {
		__antithesis_instrumentation__.Notify(17986)
		et = et | tableEventTruncate
	} else {
		__antithesis_instrumentation__.Notify(17987)
	}
	__antithesis_instrumentation__.Notify(17974)

	if regionalByRowChanged(e) {
		__antithesis_instrumentation__.Notify(17988)
		et = et | tableEventLocalityRegionalByRowChange
	} else {
		__antithesis_instrumentation__.Notify(17989)
	}
	__antithesis_instrumentation__.Notify(17975)

	return et
}

type tableEventFilter map[tableEventType]bool

func (filter tableEventFilter) shouldFilter(ctx context.Context, e TableEvent) (bool, error) {
	__antithesis_instrumentation__.Notify(17990)
	et := classifyTableEvent(e)

	if et.Contains(tableEventTruncate) {
		__antithesis_instrumentation__.Notify(17995)
		return false, errors.Errorf(`"%s" was truncated`, e.Before.GetName())
	} else {
		__antithesis_instrumentation__.Notify(17996)
	}
	__antithesis_instrumentation__.Notify(17991)

	if et == tableEventTypeUnknown {
		__antithesis_instrumentation__.Notify(17997)
		shouldFilter, ok := filter[tableEventTypeUnknown]
		if !ok {
			__antithesis_instrumentation__.Notify(17999)
			return false, errors.AssertionFailedf("policy does not specify how to handle event type %v", et)
		} else {
			__antithesis_instrumentation__.Notify(18000)
		}
		__antithesis_instrumentation__.Notify(17998)
		return shouldFilter, nil
	} else {
		__antithesis_instrumentation__.Notify(18001)
	}
	__antithesis_instrumentation__.Notify(17992)

	shouldFilter := true
	for filterEvent, filterPolicy := range filter {
		__antithesis_instrumentation__.Notify(18002)
		if et.Contains(filterEvent) && func() bool {
			__antithesis_instrumentation__.Notify(18004)
			return !filterPolicy == true
		}() == true {
			__antithesis_instrumentation__.Notify(18005)
			shouldFilter = false
		} else {
			__antithesis_instrumentation__.Notify(18006)
		}
		__antithesis_instrumentation__.Notify(18003)
		et = et.Clear(filterEvent)
	}
	__antithesis_instrumentation__.Notify(17993)
	if et > 0 {
		__antithesis_instrumentation__.Notify(18007)
		return false, errors.AssertionFailedf("policy does not specify how to handle event (unhandled event types: %v)", et)
	} else {
		__antithesis_instrumentation__.Notify(18008)
	}
	__antithesis_instrumentation__.Notify(17994)
	return shouldFilter, nil
}

func hasNewColumnDropBackfillMutation(e TableEvent) (res bool) {
	__antithesis_instrumentation__.Notify(18009)

	return !dropColumnMutationExists(e.Before) && func() bool {
		__antithesis_instrumentation__.Notify(18010)
		return dropColumnMutationExists(e.After) == true
	}() == true
}

func dropColumnMutationExists(desc catalog.TableDescriptor) bool {
	__antithesis_instrumentation__.Notify(18011)
	for _, m := range desc.AllMutations() {
		__antithesis_instrumentation__.Notify(18013)
		if m.AsColumn() == nil {
			__antithesis_instrumentation__.Notify(18015)
			continue
		} else {
			__antithesis_instrumentation__.Notify(18016)
		}
		__antithesis_instrumentation__.Notify(18014)
		if m.Dropped() && func() bool {
			__antithesis_instrumentation__.Notify(18017)
			return m.WriteAndDeleteOnly() == true
		}() == true {
			__antithesis_instrumentation__.Notify(18018)
			return true
		} else {
			__antithesis_instrumentation__.Notify(18019)
		}
	}
	__antithesis_instrumentation__.Notify(18012)
	return false
}

func newVisibleColumnBackfillComplete(e TableEvent) (res bool) {
	__antithesis_instrumentation__.Notify(18020)

	return len(e.Before.VisibleColumns()) < len(e.After.VisibleColumns()) && func() bool {
		__antithesis_instrumentation__.Notify(18021)
		return e.Before.HasColumnBackfillMutation() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(18022)
		return !e.After.HasColumnBackfillMutation() == true
	}() == true
}

func newHiddenColumnBackfillComplete(e TableEvent) (res bool) {
	__antithesis_instrumentation__.Notify(18023)
	return len(e.Before.VisibleColumns()) == len(e.After.VisibleColumns()) && func() bool {
		__antithesis_instrumentation__.Notify(18024)
		return e.Before.HasColumnBackfillMutation() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(18025)
		return !e.After.HasColumnBackfillMutation() == true
	}() == true
}

func newColumnNoBackfill(e TableEvent) (res bool) {
	__antithesis_instrumentation__.Notify(18026)
	return len(e.Before.PublicColumns()) < len(e.After.PublicColumns()) && func() bool {
		__antithesis_instrumentation__.Notify(18027)
		return !e.Before.HasColumnBackfillMutation() == true
	}() == true
}

func pkChangeMutationExists(desc catalog.TableDescriptor) bool {
	__antithesis_instrumentation__.Notify(18028)
	for _, m := range desc.AllMutations() {
		__antithesis_instrumentation__.Notify(18030)
		if m.Adding() && func() bool {
			__antithesis_instrumentation__.Notify(18031)
			return m.AsPrimaryKeySwap() != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(18032)
			return true
		} else {
			__antithesis_instrumentation__.Notify(18033)
		}
	}
	__antithesis_instrumentation__.Notify(18029)
	return false
}

func tableTruncated(e TableEvent) bool {
	__antithesis_instrumentation__.Notify(18034)

	return e.Before.GetPrimaryIndexID() != e.After.GetPrimaryIndexID() && func() bool {
		__antithesis_instrumentation__.Notify(18035)
		return !pkChangeMutationExists(e.Before) == true
	}() == true
}

func primaryKeyChanged(e TableEvent) bool {
	__antithesis_instrumentation__.Notify(18036)
	return e.Before.GetPrimaryIndexID() != e.After.GetPrimaryIndexID() && func() bool {
		__antithesis_instrumentation__.Notify(18037)
		return pkChangeMutationExists(e.Before) == true
	}() == true
}

func regionalByRowChanged(e TableEvent) bool {
	__antithesis_instrumentation__.Notify(18038)
	return e.Before.IsLocalityRegionalByRow() != e.After.IsLocalityRegionalByRow()
}

func IsPrimaryIndexChange(e TableEvent) bool {
	__antithesis_instrumentation__.Notify(18039)
	et := classifyTableEvent(e)
	return et.Contains(tableEventPrimaryKeyChange)
}

func IsOnlyPrimaryIndexChange(e TableEvent) bool {
	__antithesis_instrumentation__.Notify(18040)
	et := classifyTableEvent(e)
	return et == tableEventPrimaryKeyChange
}

func IsRegionalByRowChange(e TableEvent) bool {
	__antithesis_instrumentation__.Notify(18041)
	et := classifyTableEvent(e)
	return et.Contains(tableEventLocalityRegionalByRowChange)
}
