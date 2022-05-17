package tenantsettingswatcher

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"sort"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

type overridesStore struct {
	mu struct {
		syncutil.RWMutex

		tenants map[roachpb.TenantID]*tenantOverrides
	}
}

type tenantOverrides struct {
	overrides []roachpb.TenantSetting

	changeCh chan struct{}
}

func newTenantOverrides(overrides []roachpb.TenantSetting) *tenantOverrides {
	__antithesis_instrumentation__.Notify(238882)
	return &tenantOverrides{
		overrides: overrides,
		changeCh:  make(chan struct{}),
	}
}

func (s *overridesStore) Init() {
	__antithesis_instrumentation__.Notify(238883)
	s.mu.tenants = make(map[roachpb.TenantID]*tenantOverrides)
}

func (s *overridesStore) SetAll(allOverrides map[roachpb.TenantID][]roachpb.TenantSetting) {
	__antithesis_instrumentation__.Notify(238884)
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, existing := range s.mu.tenants {
		__antithesis_instrumentation__.Notify(238886)
		close(existing.changeCh)
	}
	__antithesis_instrumentation__.Notify(238885)
	s.mu.tenants = make(map[roachpb.TenantID]*tenantOverrides, len(allOverrides))

	for tenantID, overrides := range allOverrides {
		__antithesis_instrumentation__.Notify(238887)

		sort.Slice(overrides, func(i, j int) bool {
			__antithesis_instrumentation__.Notify(238890)
			return overrides[i].Name < overrides[j].Name
		})
		__antithesis_instrumentation__.Notify(238888)

		for i := 1; i < len(overrides); i++ {
			__antithesis_instrumentation__.Notify(238891)
			if overrides[i].Name == overrides[i-1].Name {
				__antithesis_instrumentation__.Notify(238892)
				panic("duplicate setting")
			} else {
				__antithesis_instrumentation__.Notify(238893)
			}
		}
		__antithesis_instrumentation__.Notify(238889)
		s.mu.tenants[tenantID] = newTenantOverrides(overrides)
	}
}

func (s *overridesStore) GetTenantOverrides(tenantID roachpb.TenantID) *tenantOverrides {
	__antithesis_instrumentation__.Notify(238894)
	s.mu.RLock()
	res, ok := s.mu.tenants[tenantID]
	s.mu.RUnlock()
	if ok {
		__antithesis_instrumentation__.Notify(238897)
		return res
	} else {
		__antithesis_instrumentation__.Notify(238898)
	}
	__antithesis_instrumentation__.Notify(238895)

	s.mu.Lock()
	defer s.mu.Unlock()

	if res, ok = s.mu.tenants[tenantID]; ok {
		__antithesis_instrumentation__.Notify(238899)
		return res
	} else {
		__antithesis_instrumentation__.Notify(238900)
	}
	__antithesis_instrumentation__.Notify(238896)
	res = newTenantOverrides(nil)
	s.mu.tenants[tenantID] = res
	return res
}

func (s *overridesStore) SetTenantOverride(
	tenantID roachpb.TenantID, setting roachpb.TenantSetting,
) {
	__antithesis_instrumentation__.Notify(238901)
	s.mu.Lock()
	defer s.mu.Unlock()
	var before []roachpb.TenantSetting
	if existing, ok := s.mu.tenants[tenantID]; ok {
		__antithesis_instrumentation__.Notify(238906)
		before = existing.overrides
		close(existing.changeCh)
	} else {
		__antithesis_instrumentation__.Notify(238907)
	}
	__antithesis_instrumentation__.Notify(238902)
	after := make([]roachpb.TenantSetting, 0, len(before)+1)

	for len(before) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(238908)
		return before[0].Name < setting.Name == true
	}() == true {
		__antithesis_instrumentation__.Notify(238909)
		after = append(after, before[0])
		before = before[1:]
	}
	__antithesis_instrumentation__.Notify(238903)

	if setting.Value != (settings.EncodedValue{}) {
		__antithesis_instrumentation__.Notify(238910)
		after = append(after, setting)
	} else {
		__antithesis_instrumentation__.Notify(238911)
	}
	__antithesis_instrumentation__.Notify(238904)

	if len(before) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(238912)
		return before[0].Name == setting.Name == true
	}() == true {
		__antithesis_instrumentation__.Notify(238913)
		before = before[1:]
	} else {
		__antithesis_instrumentation__.Notify(238914)
	}
	__antithesis_instrumentation__.Notify(238905)

	after = append(after, before...)
	s.mu.tenants[tenantID] = newTenantOverrides(after)
}
