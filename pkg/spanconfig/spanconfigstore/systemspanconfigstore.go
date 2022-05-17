package spanconfigstore

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"sort"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/errors"
)

type systemSpanConfigStore struct {
	store map[spanconfig.SystemTarget]roachpb.SpanConfig
}

func newSystemSpanConfigStore() *systemSpanConfigStore {
	__antithesis_instrumentation__.Notify(241637)
	store := systemSpanConfigStore{}
	store.store = make(map[spanconfig.SystemTarget]roachpb.SpanConfig)
	return &store
}

func (s *systemSpanConfigStore) apply(
	updates ...spanconfig.Update,
) (deleted []spanconfig.SystemTarget, added []spanconfig.Record, _ error) {
	__antithesis_instrumentation__.Notify(241638)
	if err := s.validateApplyArgs(updates); err != nil {
		__antithesis_instrumentation__.Notify(241641)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(241642)
	}
	__antithesis_instrumentation__.Notify(241639)

	for _, update := range updates {
		__antithesis_instrumentation__.Notify(241643)
		if update.Addition() {
			__antithesis_instrumentation__.Notify(241645)
			conf, found := s.store[update.GetTarget().GetSystemTarget()]
			if found {
				__antithesis_instrumentation__.Notify(241647)
				if conf.Equal(update.GetConfig()) {
					__antithesis_instrumentation__.Notify(241649)

					continue
				} else {
					__antithesis_instrumentation__.Notify(241650)
				}
				__antithesis_instrumentation__.Notify(241648)
				deleted = append(deleted, update.GetTarget().GetSystemTarget())
			} else {
				__antithesis_instrumentation__.Notify(241651)
			}
			__antithesis_instrumentation__.Notify(241646)

			s.store[update.GetTarget().GetSystemTarget()] = update.GetConfig()
			added = append(added, spanconfig.Record(update))
		} else {
			__antithesis_instrumentation__.Notify(241652)
		}
		__antithesis_instrumentation__.Notify(241644)

		if update.Deletion() {
			__antithesis_instrumentation__.Notify(241653)
			_, found := s.store[update.GetTarget().GetSystemTarget()]
			if !found {
				__antithesis_instrumentation__.Notify(241655)
				continue
			} else {
				__antithesis_instrumentation__.Notify(241656)
			}
			__antithesis_instrumentation__.Notify(241654)
			delete(s.store, update.GetTarget().GetSystemTarget())
			deleted = append(deleted, update.GetTarget().GetSystemTarget())
		} else {
			__antithesis_instrumentation__.Notify(241657)
		}
	}
	__antithesis_instrumentation__.Notify(241640)

	return deleted, added, nil
}

func (s *systemSpanConfigStore) combine(
	key roachpb.RKey, config roachpb.SpanConfig,
) (roachpb.SpanConfig, error) {
	__antithesis_instrumentation__.Notify(241658)
	_, tenID, err := keys.DecodeTenantPrefix(roachpb.Key(key))
	if err != nil {
		__antithesis_instrumentation__.Notify(241663)
		return roachpb.SpanConfig{}, err
	} else {
		__antithesis_instrumentation__.Notify(241664)
	}
	__antithesis_instrumentation__.Notify(241659)

	hostSetOnTenant, err := spanconfig.MakeTenantKeyspaceTarget(
		roachpb.SystemTenantID, tenID,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(241665)
		return roachpb.SpanConfig{}, err
	} else {
		__antithesis_instrumentation__.Notify(241666)
	}
	__antithesis_instrumentation__.Notify(241660)

	targets := []spanconfig.SystemTarget{
		spanconfig.MakeEntireKeyspaceTarget(),
		hostSetOnTenant,
	}

	if tenID != roachpb.SystemTenantID {
		__antithesis_instrumentation__.Notify(241667)
		target, err := spanconfig.MakeTenantKeyspaceTarget(tenID, tenID)
		if err != nil {
			__antithesis_instrumentation__.Notify(241669)
			return roachpb.SpanConfig{}, err
		} else {
			__antithesis_instrumentation__.Notify(241670)
		}
		__antithesis_instrumentation__.Notify(241668)
		targets = append(targets, target)
	} else {
		__antithesis_instrumentation__.Notify(241671)
	}
	__antithesis_instrumentation__.Notify(241661)

	for _, target := range targets {
		__antithesis_instrumentation__.Notify(241672)
		systemSpanConfig, found := s.store[target]
		if found {
			__antithesis_instrumentation__.Notify(241673)
			config.GCPolicy.ProtectionPolicies = append(
				config.GCPolicy.ProtectionPolicies, systemSpanConfig.GCPolicy.ProtectionPolicies...,
			)
		} else {
			__antithesis_instrumentation__.Notify(241674)
		}
	}
	__antithesis_instrumentation__.Notify(241662)
	return config, nil
}

func (s *systemSpanConfigStore) copy() *systemSpanConfigStore {
	__antithesis_instrumentation__.Notify(241675)
	clone := newSystemSpanConfigStore()
	for k := range s.store {
		__antithesis_instrumentation__.Notify(241677)
		clone.store[k] = s.store[k]
	}
	__antithesis_instrumentation__.Notify(241676)
	return clone
}

func (s *systemSpanConfigStore) iterate(f func(record spanconfig.Record) error) error {
	__antithesis_instrumentation__.Notify(241678)

	targets := make([]spanconfig.Target, 0, len(s.store))
	for k := range s.store {
		__antithesis_instrumentation__.Notify(241681)
		targets = append(targets, spanconfig.MakeTargetFromSystemTarget(k))
	}
	__antithesis_instrumentation__.Notify(241679)
	sort.Sort(spanconfig.Targets(targets))

	for _, target := range targets {
		__antithesis_instrumentation__.Notify(241682)
		record, err := spanconfig.MakeRecord(target, s.store[target.GetSystemTarget()])
		if err != nil {
			__antithesis_instrumentation__.Notify(241684)
			return err
		} else {
			__antithesis_instrumentation__.Notify(241685)
		}
		__antithesis_instrumentation__.Notify(241683)
		if err := f(record); err != nil {
			__antithesis_instrumentation__.Notify(241686)
			return err
		} else {
			__antithesis_instrumentation__.Notify(241687)
		}
	}
	__antithesis_instrumentation__.Notify(241680)
	return nil
}

func (s *systemSpanConfigStore) validateApplyArgs(updates []spanconfig.Update) error {
	__antithesis_instrumentation__.Notify(241688)
	for _, update := range updates {
		__antithesis_instrumentation__.Notify(241692)
		if !update.GetTarget().IsSystemTarget() {
			__antithesis_instrumentation__.Notify(241694)
			return errors.AssertionFailedf("expected update to system target update")
		} else {
			__antithesis_instrumentation__.Notify(241695)
		}
		__antithesis_instrumentation__.Notify(241693)

		if update.GetTarget().GetSystemTarget().IsReadOnly() {
			__antithesis_instrumentation__.Notify(241696)
			return errors.AssertionFailedf("invalid system target update")
		} else {
			__antithesis_instrumentation__.Notify(241697)
		}
	}
	__antithesis_instrumentation__.Notify(241689)

	sorted := make([]spanconfig.Update, len(updates))
	copy(sorted, updates)
	sort.Slice(sorted, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(241698)
		return sorted[i].GetTarget().Less(sorted[j].GetTarget())
	})
	__antithesis_instrumentation__.Notify(241690)
	updates = sorted

	for i := range updates {
		__antithesis_instrumentation__.Notify(241699)
		if i == 0 {
			__antithesis_instrumentation__.Notify(241701)
			continue
		} else {
			__antithesis_instrumentation__.Notify(241702)
		}
		__antithesis_instrumentation__.Notify(241700)

		if updates[i].GetTarget().Equal(updates[i-1].GetTarget()) {
			__antithesis_instrumentation__.Notify(241703)
			return errors.Newf(
				"found duplicate updates %s and %s",
				updates[i-1].GetTarget(),
				updates[i].GetTarget(),
			)
		} else {
			__antithesis_instrumentation__.Notify(241704)
		}
	}
	__antithesis_instrumentation__.Notify(241691)
	return nil
}
