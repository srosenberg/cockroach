package gossip

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/config"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
)

type SystemConfigDeltaFilter struct {
	keyPrefix roachpb.Key
	lastCfg   config.SystemConfigEntries
}

func MakeSystemConfigDeltaFilter(keyPrefix roachpb.Key) SystemConfigDeltaFilter {
	__antithesis_instrumentation__.Notify(68214)
	return SystemConfigDeltaFilter{
		keyPrefix: keyPrefix,
	}
}

func (df *SystemConfigDeltaFilter) ForModified(
	newCfg *config.SystemConfig, fn func(kv roachpb.KeyValue),
) {
	__antithesis_instrumentation__.Notify(68215)

	lastCfg := df.lastCfg
	df.lastCfg.Values = newCfg.Values

	lastIdx, newIdx := 0, 0
	if df.keyPrefix != nil {
		__antithesis_instrumentation__.Notify(68217)
		lastIdx = sort.Search(len(lastCfg.Values), func(i int) bool {
			__antithesis_instrumentation__.Notify(68219)
			return bytes.Compare(lastCfg.Values[i].Key, df.keyPrefix) >= 0
		})
		__antithesis_instrumentation__.Notify(68218)
		newIdx = sort.Search(len(newCfg.Values), func(i int) bool {
			__antithesis_instrumentation__.Notify(68220)
			return bytes.Compare(newCfg.Values[i].Key, df.keyPrefix) >= 0
		})
	} else {
		__antithesis_instrumentation__.Notify(68221)
	}
	__antithesis_instrumentation__.Notify(68216)

	for {
		__antithesis_instrumentation__.Notify(68222)
		if newIdx == len(newCfg.Values) {
			__antithesis_instrumentation__.Notify(68225)

			break
		} else {
			__antithesis_instrumentation__.Notify(68226)
		}
		__antithesis_instrumentation__.Notify(68223)

		newKV := newCfg.Values[newIdx]
		if df.keyPrefix != nil && func() bool {
			__antithesis_instrumentation__.Notify(68227)
			return !bytes.HasPrefix(newKV.Key, df.keyPrefix) == true
		}() == true {
			__antithesis_instrumentation__.Notify(68228)

			break
		} else {
			__antithesis_instrumentation__.Notify(68229)
		}
		__antithesis_instrumentation__.Notify(68224)

		if lastIdx < len(lastCfg.Values) {
			__antithesis_instrumentation__.Notify(68230)
			oldKV := lastCfg.Values[lastIdx]
			switch oldKV.Key.Compare(newKV.Key) {
			case -1:
				__antithesis_instrumentation__.Notify(68231)

				lastIdx++
			case 0:
				__antithesis_instrumentation__.Notify(68232)
				if !newKV.Value.EqualTagAndData(oldKV.Value) {
					__antithesis_instrumentation__.Notify(68236)

					fn(newKV)
				} else {
					__antithesis_instrumentation__.Notify(68237)
				}
				__antithesis_instrumentation__.Notify(68233)
				lastIdx++
				newIdx++
			case 1:
				__antithesis_instrumentation__.Notify(68234)

				fn(newKV)
				newIdx++
			default:
				__antithesis_instrumentation__.Notify(68235)
			}
		} else {
			__antithesis_instrumentation__.Notify(68238)

			fn(newKV)
			newIdx++
		}
	}
}
