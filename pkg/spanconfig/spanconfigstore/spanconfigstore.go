package spanconfigstore

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/util/interval"
	"github.com/cockroachdb/cockroach/pkg/util/iterutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

type spanConfigStore struct {
	tree    interval.Tree
	idAlloc int64
}

func newSpanConfigStore() *spanConfigStore {
	__antithesis_instrumentation__.Notify(241469)
	s := &spanConfigStore{}
	s.tree = interval.NewTree(interval.ExclusiveOverlapper)
	return s
}

func (s *spanConfigStore) copy(ctx context.Context) *spanConfigStore {
	__antithesis_instrumentation__.Notify(241470)
	clone := newSpanConfigStore()
	_ = s.forEachOverlapping(keys.EverythingSpan, func(entry spanConfigEntry) error {
		__antithesis_instrumentation__.Notify(241472)
		record, err := spanconfig.MakeRecord(spanconfig.MakeTargetFromSpan(entry.span), entry.config)
		if err != nil {
			__antithesis_instrumentation__.Notify(241475)
			log.Fatalf(ctx, "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(241476)
		}
		__antithesis_instrumentation__.Notify(241473)
		_, _, err = clone.apply(false, spanconfig.Update(record))
		if err != nil {
			__antithesis_instrumentation__.Notify(241477)
			log.Fatalf(ctx, "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(241478)
		}
		__antithesis_instrumentation__.Notify(241474)
		return nil
	})
	__antithesis_instrumentation__.Notify(241471)
	return clone
}

func (s *spanConfigStore) forEachOverlapping(sp roachpb.Span, f func(spanConfigEntry) error) error {
	__antithesis_instrumentation__.Notify(241479)

	for _, overlapping := range s.tree.Get(sp.AsRange()) {
		__antithesis_instrumentation__.Notify(241481)
		entry := overlapping.(*spanConfigStoreEntry).spanConfigEntry
		if err := f(entry); err != nil {
			__antithesis_instrumentation__.Notify(241482)
			if iterutil.Done(err) {
				__antithesis_instrumentation__.Notify(241484)
				err = nil
			} else {
				__antithesis_instrumentation__.Notify(241485)
			}
			__antithesis_instrumentation__.Notify(241483)
			return err
		} else {
			__antithesis_instrumentation__.Notify(241486)
		}
	}
	__antithesis_instrumentation__.Notify(241480)
	return nil
}

func (s *spanConfigStore) computeSplitKey(start, end roachpb.RKey) roachpb.RKey {
	__antithesis_instrumentation__.Notify(241487)
	sp := roachpb.Span{Key: start.AsRawKey(), EndKey: end.AsRawKey()}

	if keys.SystemConfigSpan.Contains(sp) {
		__antithesis_instrumentation__.Notify(241491)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(241492)
	}
	__antithesis_instrumentation__.Notify(241488)
	if keys.SystemConfigSpan.ContainsKey(sp.Key) {
		__antithesis_instrumentation__.Notify(241493)
		return roachpb.RKey(keys.SystemConfigSpan.EndKey)
	} else {
		__antithesis_instrumentation__.Notify(241494)
	}
	__antithesis_instrumentation__.Notify(241489)

	idx := 0
	var splitKey roachpb.RKey = nil
	s.tree.DoMatching(func(i interval.Interface) (done bool) {
		__antithesis_instrumentation__.Notify(241495)
		if idx > 0 {
			__antithesis_instrumentation__.Notify(241497)
			splitKey = roachpb.RKey(i.(*spanConfigStoreEntry).span.Key)
			return true
		} else {
			__antithesis_instrumentation__.Notify(241498)
		}
		__antithesis_instrumentation__.Notify(241496)

		idx++
		return false
	}, sp.AsRange())
	__antithesis_instrumentation__.Notify(241490)

	return splitKey
}

func (s *spanConfigStore) getSpanConfigForKey(
	ctx context.Context, key roachpb.RKey,
) (roachpb.SpanConfig, bool, error) {
	__antithesis_instrumentation__.Notify(241499)
	sp := roachpb.Span{Key: key.AsRawKey(), EndKey: key.Next().AsRawKey()}

	var conf roachpb.SpanConfig
	found := false
	s.tree.DoMatching(func(i interval.Interface) (done bool) {
		__antithesis_instrumentation__.Notify(241502)
		conf = i.(*spanConfigStoreEntry).config
		found = true
		return true
	}, sp.AsRange())
	__antithesis_instrumentation__.Notify(241500)

	if !found {
		__antithesis_instrumentation__.Notify(241503)
		if log.ExpensiveLogEnabled(ctx, 1) {
			__antithesis_instrumentation__.Notify(241504)
			log.Warningf(ctx, "span config not found for %s", key.String())
		} else {
			__antithesis_instrumentation__.Notify(241505)
		}
	} else {
		__antithesis_instrumentation__.Notify(241506)
	}
	__antithesis_instrumentation__.Notify(241501)
	return conf, found, nil
}

func (s *spanConfigStore) apply(
	dryrun bool, updates ...spanconfig.Update,
) (deleted []roachpb.Span, added []spanConfigStoreEntry, err error) {
	__antithesis_instrumentation__.Notify(241507)
	if err := validateApplyArgs(updates...); err != nil {
		__antithesis_instrumentation__.Notify(241512)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(241513)
	}
	__antithesis_instrumentation__.Notify(241508)

	sorted := make([]spanconfig.Update, len(updates))
	copy(sorted, updates)
	sort.Slice(sorted, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(241514)
		return sorted[i].GetTarget().Less(sorted[j].GetTarget())
	})
	__antithesis_instrumentation__.Notify(241509)
	updates = sorted

	entriesToDelete, entriesToAdd := s.accumulateOpsFor(updates)

	deleted = make([]roachpb.Span, len(entriesToDelete))
	for i := range entriesToDelete {
		__antithesis_instrumentation__.Notify(241515)
		entry := &entriesToDelete[i]
		if !dryrun {
			__antithesis_instrumentation__.Notify(241517)
			if err := s.tree.Delete(entry, false); err != nil {
				__antithesis_instrumentation__.Notify(241518)
				return nil, nil, err
			} else {
				__antithesis_instrumentation__.Notify(241519)
			}
		} else {
			__antithesis_instrumentation__.Notify(241520)
		}
		__antithesis_instrumentation__.Notify(241516)
		deleted[i] = entry.span
	}
	__antithesis_instrumentation__.Notify(241510)

	added = make([]spanConfigStoreEntry, len(entriesToAdd))
	for i := range entriesToAdd {
		__antithesis_instrumentation__.Notify(241521)
		entry := &entriesToAdd[i]
		if !dryrun {
			__antithesis_instrumentation__.Notify(241523)
			if err := s.tree.Insert(entry, false); err != nil {
				__antithesis_instrumentation__.Notify(241524)
				return nil, nil, err
			} else {
				__antithesis_instrumentation__.Notify(241525)
			}
		} else {
			__antithesis_instrumentation__.Notify(241526)
		}
		__antithesis_instrumentation__.Notify(241522)
		added[i] = *entry
	}
	__antithesis_instrumentation__.Notify(241511)

	return deleted, added, nil
}

func (s *spanConfigStore) accumulateOpsFor(
	updates []spanconfig.Update,
) (toDelete, toAdd []spanConfigStoreEntry) {
	__antithesis_instrumentation__.Notify(241527)
	var carryOver spanConfigEntry
	for _, update := range updates {
		__antithesis_instrumentation__.Notify(241530)
		var carriedOver spanConfigEntry
		carriedOver, carryOver = carryOver, spanConfigEntry{}
		if update.GetTarget().GetSpan().Overlaps(carriedOver.span) {
			__antithesis_instrumentation__.Notify(241533)
			gapBetweenUpdates := roachpb.Span{
				Key:    carriedOver.span.Key,
				EndKey: update.GetTarget().GetSpan().Key}
			if gapBetweenUpdates.Valid() {
				__antithesis_instrumentation__.Notify(241535)
				toAdd = append(toAdd, s.makeEntry(gapBetweenUpdates, carriedOver.config))
			} else {
				__antithesis_instrumentation__.Notify(241536)
			}
			__antithesis_instrumentation__.Notify(241534)

			carryOverSpanAfterUpdate := roachpb.Span{
				Key:    update.GetTarget().GetSpan().EndKey,
				EndKey: carriedOver.span.EndKey}
			if carryOverSpanAfterUpdate.Valid() {
				__antithesis_instrumentation__.Notify(241537)
				carryOver = spanConfigEntry{
					span:   carryOverSpanAfterUpdate,
					config: carriedOver.config,
				}
			} else {
				__antithesis_instrumentation__.Notify(241538)
			}
		} else {
			__antithesis_instrumentation__.Notify(241539)
			if !carriedOver.isEmpty() {
				__antithesis_instrumentation__.Notify(241540)
				toAdd = append(toAdd, s.makeEntry(carriedOver.span, carriedOver.config))
			} else {
				__antithesis_instrumentation__.Notify(241541)
			}
		}
		__antithesis_instrumentation__.Notify(241531)

		skipAddingSelf := false
		for _, overlapping := range s.tree.Get(update.GetTarget().GetSpan().AsRange()) {
			__antithesis_instrumentation__.Notify(241542)
			existing := overlapping.(*spanConfigStoreEntry)
			if existing.span.Overlaps(carriedOver.span) {
				__antithesis_instrumentation__.Notify(241546)
				continue
			} else {
				__antithesis_instrumentation__.Notify(241547)
			}
			__antithesis_instrumentation__.Notify(241543)

			var (
				union = existing.span.Combine(update.GetTarget().GetSpan())
				inter = existing.span.Intersect(update.GetTarget().GetSpan())

				pre  = roachpb.Span{Key: union.Key, EndKey: inter.Key}
				post = roachpb.Span{Key: inter.EndKey, EndKey: union.EndKey}
			)

			if update.Addition() {
				__antithesis_instrumentation__.Notify(241548)
				if existing.span.Equal(update.GetTarget().GetSpan()) && func() bool {
					__antithesis_instrumentation__.Notify(241549)
					return existing.config.Equal(update.GetConfig()) == true
				}() == true {
					__antithesis_instrumentation__.Notify(241550)
					skipAddingSelf = true
					break
				} else {
					__antithesis_instrumentation__.Notify(241551)
				}
			} else {
				__antithesis_instrumentation__.Notify(241552)
			}
			__antithesis_instrumentation__.Notify(241544)

			toDelete = append(toDelete, *existing)

			if existing.span.ContainsKey(update.GetTarget().GetSpan().Key) {
				__antithesis_instrumentation__.Notify(241553)

				if pre.Valid() {
					__antithesis_instrumentation__.Notify(241554)
					toAdd = append(toAdd, s.makeEntry(pre, existing.config))
				} else {
					__antithesis_instrumentation__.Notify(241555)
				}
			} else {
				__antithesis_instrumentation__.Notify(241556)
			}
			__antithesis_instrumentation__.Notify(241545)

			if existing.span.ContainsKey(update.GetTarget().GetSpan().EndKey) {
				__antithesis_instrumentation__.Notify(241557)

				carryOver = spanConfigEntry{
					span:   post,
					config: existing.config,
				}
			} else {
				__antithesis_instrumentation__.Notify(241558)
			}
		}
		__antithesis_instrumentation__.Notify(241532)

		if update.Addition() && func() bool {
			__antithesis_instrumentation__.Notify(241559)
			return !skipAddingSelf == true
		}() == true {
			__antithesis_instrumentation__.Notify(241560)

			toAdd = append(toAdd, s.makeEntry(update.GetTarget().GetSpan(), update.GetConfig()))

		} else {
			__antithesis_instrumentation__.Notify(241561)
		}
	}
	__antithesis_instrumentation__.Notify(241528)

	if !carryOver.isEmpty() {
		__antithesis_instrumentation__.Notify(241562)
		toAdd = append(toAdd, s.makeEntry(carryOver.span, carryOver.config))
	} else {
		__antithesis_instrumentation__.Notify(241563)
	}
	__antithesis_instrumentation__.Notify(241529)
	return toDelete, toAdd
}

func (s *spanConfigStore) makeEntry(sp roachpb.Span, conf roachpb.SpanConfig) spanConfigStoreEntry {
	__antithesis_instrumentation__.Notify(241564)
	s.idAlloc++
	return spanConfigStoreEntry{
		spanConfigEntry: spanConfigEntry{span: sp, config: conf},
		id:              s.idAlloc,
	}
}

func validateApplyArgs(updates ...spanconfig.Update) error {
	__antithesis_instrumentation__.Notify(241565)
	for i := range updates {
		__antithesis_instrumentation__.Notify(241569)
		if !updates[i].GetTarget().IsSpanTarget() {
			__antithesis_instrumentation__.Notify(241571)
			return errors.New("expected update to target a span")
		} else {
			__antithesis_instrumentation__.Notify(241572)
		}
		__antithesis_instrumentation__.Notify(241570)

		sp := updates[i].GetTarget().GetSpan()
		if !sp.Valid() || func() bool {
			__antithesis_instrumentation__.Notify(241573)
			return len(sp.EndKey) == 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(241574)
			return errors.New("invalid span")
		} else {
			__antithesis_instrumentation__.Notify(241575)
		}
	}
	__antithesis_instrumentation__.Notify(241566)

	sorted := make([]spanconfig.Update, len(updates))
	copy(sorted, updates)
	sort.Slice(sorted, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(241576)
		return sorted[i].GetTarget().GetSpan().Key.Compare(sorted[j].GetTarget().GetSpan().Key) < 0
	})
	__antithesis_instrumentation__.Notify(241567)
	updates = sorted

	for i := range updates {
		__antithesis_instrumentation__.Notify(241577)
		if i == 0 {
			__antithesis_instrumentation__.Notify(241579)
			continue
		} else {
			__antithesis_instrumentation__.Notify(241580)
		}
		__antithesis_instrumentation__.Notify(241578)
		if updates[i].GetTarget().GetSpan().Overlaps(updates[i-1].GetTarget().GetSpan()) {
			__antithesis_instrumentation__.Notify(241581)
			return errors.Newf(
				"found overlapping updates %s and %s",
				updates[i-1].GetTarget().GetSpan(),
				updates[i].GetTarget().GetSpan(),
			)
		} else {
			__antithesis_instrumentation__.Notify(241582)
		}
	}
	__antithesis_instrumentation__.Notify(241568)
	return nil
}

type spanConfigEntry struct {
	span   roachpb.Span
	config roachpb.SpanConfig
}

func (s *spanConfigEntry) isEmpty() bool {
	__antithesis_instrumentation__.Notify(241583)
	return s.span.Equal(roachpb.Span{}) && func() bool {
		__antithesis_instrumentation__.Notify(241584)
		return s.config.IsEmpty() == true
	}() == true
}

type spanConfigStoreEntry struct {
	spanConfigEntry
	id int64
}

var _ interval.Interface = &spanConfigStoreEntry{}

func (s *spanConfigStoreEntry) Range() interval.Range {
	__antithesis_instrumentation__.Notify(241585)
	return s.span.AsRange()
}

func (s *spanConfigStoreEntry) ID() uintptr {
	__antithesis_instrumentation__.Notify(241586)
	return uintptr(s.id)
}
