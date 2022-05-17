package spanconfigsqlwatcher

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangefeed/rangefeedbuffer"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

type buffer struct {
	mu struct {
		syncutil.Mutex

		buffer *rangefeedbuffer.Buffer

		rangefeedFrontiers [numRangefeeds]hlc.Timestamp
	}
}

type event struct {
	timestamp hlc.Timestamp

	update spanconfig.SQLUpdate
}

func (e event) Timestamp() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(241242)
	return e.timestamp
}

type rangefeedKind int

const (
	zonesRangefeed rangefeedKind = iota
	descriptorsRangefeed
	protectedTimestampRangefeed

	numRangefeeds int = iota
)

func newBuffer(limit int, initialFrontierTS hlc.Timestamp) *buffer {
	__antithesis_instrumentation__.Notify(241243)
	rangefeedBuffer := rangefeedbuffer.New(limit)
	eventBuffer := &buffer{}
	eventBuffer.mu.buffer = rangefeedBuffer
	for i := range eventBuffer.mu.rangefeedFrontiers {
		__antithesis_instrumentation__.Notify(241245)
		eventBuffer.mu.rangefeedFrontiers[i].Forward(initialFrontierTS)
	}
	__antithesis_instrumentation__.Notify(241244)
	return eventBuffer
}

func (b *buffer) advance(rangefeed rangefeedKind, timestamp hlc.Timestamp) {
	__antithesis_instrumentation__.Notify(241246)
	b.mu.Lock()
	defer b.mu.Unlock()
	b.mu.rangefeedFrontiers[rangefeed].Forward(timestamp)
}

func (b *buffer) add(ev event) error {
	__antithesis_instrumentation__.Notify(241247)
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.mu.buffer.Add(ev)
}

type events []event

func (s events) Len() int {
	__antithesis_instrumentation__.Notify(241248)
	return len(s)
}

func (s events) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(241249)
	ei, ej := s[i], s[j]
	if ei.update.IsDescriptorUpdate() && func() bool {
		__antithesis_instrumentation__.Notify(241256)
		return ej.update.IsDescriptorUpdate() == true
	}() == true {
		__antithesis_instrumentation__.Notify(241257)
		descUpdatei := ei.update.GetDescriptorUpdate()
		descUpdatej := ej.update.GetDescriptorUpdate()
		if descUpdatei.ID == descUpdatej.ID {
			__antithesis_instrumentation__.Notify(241259)
			return ei.timestamp.Less(ej.timestamp)
		} else {
			__antithesis_instrumentation__.Notify(241260)
		}
		__antithesis_instrumentation__.Notify(241258)
		return descUpdatei.ID < descUpdatej.ID
	} else {
		__antithesis_instrumentation__.Notify(241261)
	}
	__antithesis_instrumentation__.Notify(241250)

	if ei.update.IsDescriptorUpdate() {
		__antithesis_instrumentation__.Notify(241262)
		return true
	} else {
		__antithesis_instrumentation__.Notify(241263)
	}
	__antithesis_instrumentation__.Notify(241251)

	if ej.update.IsDescriptorUpdate() {
		__antithesis_instrumentation__.Notify(241264)
		return false
	} else {
		__antithesis_instrumentation__.Notify(241265)
	}
	__antithesis_instrumentation__.Notify(241252)

	lhsUpdate := ei.update.GetProtectedTimestampUpdate()
	rhsUpdate := ej.update.GetProtectedTimestampUpdate()
	if lhsUpdate.IsTenantsUpdate() && func() bool {
		__antithesis_instrumentation__.Notify(241266)
		return rhsUpdate.IsTenantsUpdate() == true
	}() == true {
		__antithesis_instrumentation__.Notify(241267)
		if lhsUpdate.TenantTarget == rhsUpdate.TenantTarget {
			__antithesis_instrumentation__.Notify(241269)
			return ei.timestamp.Less(ej.timestamp)
		} else {
			__antithesis_instrumentation__.Notify(241270)
		}
		__antithesis_instrumentation__.Notify(241268)
		return lhsUpdate.TenantTarget.ToUint64() < rhsUpdate.TenantTarget.ToUint64()
	} else {
		__antithesis_instrumentation__.Notify(241271)
	}
	__antithesis_instrumentation__.Notify(241253)

	if lhsUpdate.IsTenantsUpdate() {
		__antithesis_instrumentation__.Notify(241272)
		return true
	} else {
		__antithesis_instrumentation__.Notify(241273)
	}
	__antithesis_instrumentation__.Notify(241254)

	if rhsUpdate.IsTenantsUpdate() {
		__antithesis_instrumentation__.Notify(241274)
		return false
	} else {
		__antithesis_instrumentation__.Notify(241275)
	}
	__antithesis_instrumentation__.Notify(241255)

	return ei.timestamp.Less(ej.timestamp)
}

func (s events) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(241276)
	s[i], s[j] = s[j], s[i]
}

var _ sort.Interface = &events{}

func (b *buffer) flushEvents(
	ctx context.Context,
) (updates []rangefeedbuffer.Event, combinedFrontierTS hlc.Timestamp) {
	__antithesis_instrumentation__.Notify(241277)
	b.mu.Lock()
	defer b.mu.Unlock()

	combinedFrontierTS = hlc.MaxTimestamp
	for _, ts := range b.mu.rangefeedFrontiers {
		__antithesis_instrumentation__.Notify(241279)
		combinedFrontierTS.Backward(ts)
	}
	__antithesis_instrumentation__.Notify(241278)

	return b.mu.buffer.Flush(ctx, combinedFrontierTS), combinedFrontierTS
}

func (b *buffer) flush(
	ctx context.Context,
) (sqlUpdates []spanconfig.SQLUpdate, _ hlc.Timestamp, _ error) {
	__antithesis_instrumentation__.Notify(241280)
	bufferedEvents, combinedFrontierTS := b.flushEvents(ctx)
	evs := make(events, 0, len(bufferedEvents))
	for i, ev := range bufferedEvents {
		__antithesis_instrumentation__.Notify(241285)
		if i != 0 {
			__antithesis_instrumentation__.Notify(241287)
			prevEv := bufferedEvents[i-1]
			if !prevEv.Timestamp().LessEq(ev.Timestamp()) {
				__antithesis_instrumentation__.Notify(241288)
				log.Fatalf(ctx, "expected events to be sorted by timestamp but found %+v after %+v",
					ev, prevEv)
			} else {
				__antithesis_instrumentation__.Notify(241289)
			}
		} else {
			__antithesis_instrumentation__.Notify(241290)
		}
		__antithesis_instrumentation__.Notify(241286)
		evs = append(evs, ev.(event))
	}
	__antithesis_instrumentation__.Notify(241281)

	bufferedEvents = nil

	sort.Sort(evs)

	descriptorUpdatesIdx := sort.Search(len(evs), func(i int) bool {
		__antithesis_instrumentation__.Notify(241291)
		update := evs[i].update
		return update.IsProtectedTimestampUpdate()
	})
	__antithesis_instrumentation__.Notify(241282)

	for i, ev := range evs[:descriptorUpdatesIdx] {
		__antithesis_instrumentation__.Notify(241292)
		update := ev.update
		descriptorUpdate := update.GetDescriptorUpdate()
		if i == 0 {
			__antithesis_instrumentation__.Notify(241296)
			sqlUpdates = append(sqlUpdates, update)
			continue
		} else {
			__antithesis_instrumentation__.Notify(241297)
		}
		__antithesis_instrumentation__.Notify(241293)

		prevUpdate := evs[i-1].update
		if prevUpdate.GetDescriptorUpdate().ID != descriptorUpdate.ID {
			__antithesis_instrumentation__.Notify(241298)
			sqlUpdates = append(sqlUpdates, update)
			continue
		} else {
			__antithesis_instrumentation__.Notify(241299)
		}
		__antithesis_instrumentation__.Notify(241294)

		prevDescriptorSQLUpdate := sqlUpdates[len(sqlUpdates)-1].GetDescriptorUpdate()
		descType, err := combine(prevDescriptorSQLUpdate.Type,
			descriptorUpdate.Type)
		if err != nil {
			__antithesis_instrumentation__.Notify(241300)
			return nil, hlc.Timestamp{}, err
		} else {
			__antithesis_instrumentation__.Notify(241301)
		}
		__antithesis_instrumentation__.Notify(241295)
		sqlUpdates[len(sqlUpdates)-1] = spanconfig.MakeDescriptorSQLUpdate(
			prevDescriptorSQLUpdate.ID, descType)
	}
	__antithesis_instrumentation__.Notify(241283)

	evs = evs[descriptorUpdatesIdx:]

	for i, ev := range evs {
		__antithesis_instrumentation__.Notify(241302)
		update := ev.update
		curUpdate := update.GetProtectedTimestampUpdate()
		if i == 0 {
			__antithesis_instrumentation__.Notify(241306)
			sqlUpdates = append(sqlUpdates, update)
			continue
		} else {
			__antithesis_instrumentation__.Notify(241307)
		}
		__antithesis_instrumentation__.Notify(241303)

		prevUpdate := evs[i-1].update.GetProtectedTimestampUpdate()

		if prevUpdate.IsTenantsUpdate() && func() bool {
			__antithesis_instrumentation__.Notify(241308)
			return curUpdate.IsTenantsUpdate() == true
		}() == true {
			__antithesis_instrumentation__.Notify(241309)
			if prevUpdate.TenantTarget == curUpdate.TenantTarget {
				__antithesis_instrumentation__.Notify(241310)
				continue
			} else {
				__antithesis_instrumentation__.Notify(241311)
				sqlUpdates = append(sqlUpdates, update)
			}
		} else {
			__antithesis_instrumentation__.Notify(241312)
		}
		__antithesis_instrumentation__.Notify(241304)

		if prevUpdate.IsClusterUpdate() && func() bool {
			__antithesis_instrumentation__.Notify(241313)
			return curUpdate.IsClusterUpdate() == true
		}() == true {
			__antithesis_instrumentation__.Notify(241314)
			continue
		} else {
			__antithesis_instrumentation__.Notify(241315)
		}
		__antithesis_instrumentation__.Notify(241305)

		sqlUpdates = append(sqlUpdates, update)
	}
	__antithesis_instrumentation__.Notify(241284)

	return sqlUpdates, combinedFrontierTS, nil
}

func combine(d1, d2 catalog.DescriptorType) (catalog.DescriptorType, error) {
	__antithesis_instrumentation__.Notify(241316)
	if d1 == d2 {
		__antithesis_instrumentation__.Notify(241320)
		return d1, nil
	} else {
		__antithesis_instrumentation__.Notify(241321)
	}
	__antithesis_instrumentation__.Notify(241317)
	if d1 == catalog.Any {
		__antithesis_instrumentation__.Notify(241322)
		return d2, nil
	} else {
		__antithesis_instrumentation__.Notify(241323)
	}
	__antithesis_instrumentation__.Notify(241318)
	if d2 == catalog.Any {
		__antithesis_instrumentation__.Notify(241324)
		return d1, nil
	} else {
		__antithesis_instrumentation__.Notify(241325)
	}
	__antithesis_instrumentation__.Notify(241319)
	return catalog.Any, spanconfig.NewMismatchedDescriptorTypesError(d1, d2)
}
