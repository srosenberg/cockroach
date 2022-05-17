//go:build !windows
// +build !windows

package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

type rlimit struct {
	Cur, Max uint64
}

func setOpenFileLimitInner(physicalStoreCount int) (uint64, error) {
	__antithesis_instrumentation__.Notify(190143)
	minimumOpenFileLimit := uint64(physicalStoreCount*storage.MinimumMaxOpenFiles + minimumNetworkFileDescriptors)
	networkConstrainedFileLimit := uint64(physicalStoreCount*storage.RecommendedMaxOpenFiles + minimumNetworkFileDescriptors)
	recommendedOpenFileLimit := uint64(physicalStoreCount*storage.RecommendedMaxOpenFiles + recommendedNetworkFileDescriptors)
	var rLimit rlimit
	if err := getRlimitNoFile(&rLimit); err != nil {
		__antithesis_instrumentation__.Notify(190153)
		if log.V(1) {
			__antithesis_instrumentation__.Notify(190155)
			log.Infof(context.TODO(), "could not get rlimit; setting maxOpenFiles to the recommended value %d - %s", storage.RecommendedMaxOpenFiles, err)
		} else {
			__antithesis_instrumentation__.Notify(190156)
		}
		__antithesis_instrumentation__.Notify(190154)
		return storage.RecommendedMaxOpenFiles, nil
	} else {
		__antithesis_instrumentation__.Notify(190157)
	}
	__antithesis_instrumentation__.Notify(190144)

	if rLimit.Max < minimumOpenFileLimit {
		__antithesis_instrumentation__.Notify(190158)
		return 0, fmt.Errorf("hard open file descriptor limit of %d is under the minimum required %d\n%s",
			rLimit.Max,
			minimumOpenFileLimit,
			productionSettingsWebpage)
	} else {
		__antithesis_instrumentation__.Notify(190159)
	}
	__antithesis_instrumentation__.Notify(190145)

	var newCurrent uint64
	if rLimit.Max > recommendedOpenFileLimit {
		__antithesis_instrumentation__.Notify(190160)
		newCurrent = recommendedOpenFileLimit
	} else {
		__antithesis_instrumentation__.Notify(190161)
		newCurrent = rLimit.Max
	}
	__antithesis_instrumentation__.Notify(190146)
	if rLimit.Cur < newCurrent {
		__antithesis_instrumentation__.Notify(190162)
		if log.V(1) {
			__antithesis_instrumentation__.Notify(190166)
			log.Infof(context.TODO(), "setting the soft limit for open file descriptors from %d to %d",
				rLimit.Cur, newCurrent)
		} else {
			__antithesis_instrumentation__.Notify(190167)
		}
		__antithesis_instrumentation__.Notify(190163)
		oldCurrent := rLimit.Cur
		rLimit.Cur = newCurrent
		if err := setRlimitNoFile(&rLimit); err != nil {
			__antithesis_instrumentation__.Notify(190168)

			log.Warningf(context.TODO(), "adjusting the limit for open file descriptors to %d failed: %s",
				rLimit.Cur, err)

			if oldCurrent < minimumOpenFileLimit {
				__antithesis_instrumentation__.Notify(190169)
				rLimit.Cur = minimumOpenFileLimit
				if err := setRlimitNoFile(&rLimit); err != nil {
					__antithesis_instrumentation__.Notify(190170)
					log.Warningf(context.TODO(), "adjusting the limit for open file descriptors to %d failed: %s",
						rLimit.Cur, err)
				} else {
					__antithesis_instrumentation__.Notify(190171)
				}
			} else {
				__antithesis_instrumentation__.Notify(190172)
			}
		} else {
			__antithesis_instrumentation__.Notify(190173)
		}
		__antithesis_instrumentation__.Notify(190164)

		if err := getRlimitNoFile(&rLimit); err != nil {
			__antithesis_instrumentation__.Notify(190174)
			return 0, errors.Wrap(err, "getting updated soft limit for open file descriptors")
		} else {
			__antithesis_instrumentation__.Notify(190175)
		}
		__antithesis_instrumentation__.Notify(190165)
		if log.V(1) {
			__antithesis_instrumentation__.Notify(190176)
			log.Infof(context.TODO(), "soft open file descriptor limit is now %d", rLimit.Cur)
		} else {
			__antithesis_instrumentation__.Notify(190177)
		}
	} else {
		__antithesis_instrumentation__.Notify(190178)
	}
	__antithesis_instrumentation__.Notify(190147)

	if rLimit.Cur < minimumOpenFileLimit {
		__antithesis_instrumentation__.Notify(190179)
		return 0, fmt.Errorf("soft open file descriptor limit of %d is under the minimum required %d and cannot be increased\n%s",
			rLimit.Cur,
			minimumOpenFileLimit,
			productionSettingsWebpage)
	} else {
		__antithesis_instrumentation__.Notify(190180)
	}
	__antithesis_instrumentation__.Notify(190148)

	if rLimit.Cur < recommendedOpenFileLimit {
		__antithesis_instrumentation__.Notify(190181)

		log.Warningf(context.TODO(), "soft open file descriptor limit %d is under the recommended limit %d; this may decrease performance\n%s",
			rLimit.Cur,
			recommendedOpenFileLimit,
			productionSettingsWebpage)
	} else {
		__antithesis_instrumentation__.Notify(190182)
	}
	__antithesis_instrumentation__.Notify(190149)

	if physicalStoreCount == 0 {
		__antithesis_instrumentation__.Notify(190183)
		return 0, nil
	} else {
		__antithesis_instrumentation__.Notify(190184)
	}
	__antithesis_instrumentation__.Notify(190150)

	if rLimit.Cur >= recommendedOpenFileLimit {
		__antithesis_instrumentation__.Notify(190185)
		return (rLimit.Cur - recommendedNetworkFileDescriptors) / uint64(physicalStoreCount), nil
	} else {
		__antithesis_instrumentation__.Notify(190186)
	}
	__antithesis_instrumentation__.Notify(190151)

	if rLimit.Cur >= networkConstrainedFileLimit {
		__antithesis_instrumentation__.Notify(190187)
		return storage.RecommendedMaxOpenFiles, nil
	} else {
		__antithesis_instrumentation__.Notify(190188)
	}
	__antithesis_instrumentation__.Notify(190152)

	return (rLimit.Cur - minimumNetworkFileDescriptors) / uint64(physicalStoreCount), nil
}
