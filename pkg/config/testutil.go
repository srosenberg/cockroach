package config

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

type zoneConfigMap map[ObjectID]zonepb.ZoneConfig

var (
	testingZoneConfig   zoneConfigMap
	testingHasHook      bool
	testingPreviousHook zoneConfigHook
	testingLock         syncutil.Mutex
)

func TestingSetupZoneConfigHook(stopper *stop.Stopper) {
	__antithesis_instrumentation__.Notify(56222)
	stopper.AddCloser(stop.CloserFn(testingResetZoneConfigHook))

	testingLock.Lock()
	defer testingLock.Unlock()
	if testingHasHook {
		__antithesis_instrumentation__.Notify(56224)
		panic("TestingSetupZoneConfigHook called without restoring state")
	} else {
		__antithesis_instrumentation__.Notify(56225)
	}
	__antithesis_instrumentation__.Notify(56223)
	testingHasHook = true
	testingZoneConfig = make(zoneConfigMap)
	testingPreviousHook = ZoneConfigHook
	ZoneConfigHook = testingZoneConfigHook
	testingLargestIDHook = func(maxID ObjectID) (max ObjectID) {
		__antithesis_instrumentation__.Notify(56226)
		testingLock.Lock()
		defer testingLock.Unlock()
		for id := range testingZoneConfig {
			__antithesis_instrumentation__.Notify(56228)
			if maxID != 0 && func() bool {
				__antithesis_instrumentation__.Notify(56230)
				return maxID < id == true
			}() == true {
				__antithesis_instrumentation__.Notify(56231)
				continue
			} else {
				__antithesis_instrumentation__.Notify(56232)
			}
			__antithesis_instrumentation__.Notify(56229)
			if id > max {
				__antithesis_instrumentation__.Notify(56233)
				max = id
			} else {
				__antithesis_instrumentation__.Notify(56234)
			}
		}
		__antithesis_instrumentation__.Notify(56227)
		return
	}
}

func testingResetZoneConfigHook() {
	__antithesis_instrumentation__.Notify(56235)
	testingLock.Lock()
	defer testingLock.Unlock()
	if !testingHasHook {
		__antithesis_instrumentation__.Notify(56237)
		panic("TestingResetZoneConfigHook called on uninitialized testing hook")
	} else {
		__antithesis_instrumentation__.Notify(56238)
	}
	__antithesis_instrumentation__.Notify(56236)
	testingHasHook = false
	ZoneConfigHook = testingPreviousHook
	testingLargestIDHook = nil
}

func TestingSetZoneConfig(id ObjectID, zone zonepb.ZoneConfig) {
	__antithesis_instrumentation__.Notify(56239)
	testingLock.Lock()
	defer testingLock.Unlock()
	testingZoneConfig[id] = zone
}

func testingZoneConfigHook(
	_ *SystemConfig, codec keys.SQLCodec, id ObjectID,
) (*zonepb.ZoneConfig, *zonepb.ZoneConfig, bool, error) {
	__antithesis_instrumentation__.Notify(56240)
	testingLock.Lock()
	defer testingLock.Unlock()
	if zone, ok := testingZoneConfig[id]; ok {
		__antithesis_instrumentation__.Notify(56242)
		return &zone, nil, false, nil
	} else {
		__antithesis_instrumentation__.Notify(56243)
	}
	__antithesis_instrumentation__.Notify(56241)
	return nil, nil, false, nil
}
