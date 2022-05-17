package config

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"sort"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
)

type SystemConfigMask struct {
	allowed []roachpb.Key
}

func MakeSystemConfigMask(allowed ...roachpb.Key) SystemConfigMask {
	__antithesis_instrumentation__.Notify(56211)
	sort.Slice(allowed, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(56213)
		return allowed[i].Compare(allowed[j]) < 0
	})
	__antithesis_instrumentation__.Notify(56212)
	return SystemConfigMask{allowed: allowed}
}

func (m SystemConfigMask) Apply(entries SystemConfigEntries) SystemConfigEntries {
	__antithesis_instrumentation__.Notify(56214)
	var res SystemConfigEntries
	for _, key := range m.allowed {
		__antithesis_instrumentation__.Notify(56216)
		i := sort.Search(len(entries.Values), func(i int) bool {
			__antithesis_instrumentation__.Notify(56218)
			return entries.Values[i].Key.Compare(key) >= 0
		})
		__antithesis_instrumentation__.Notify(56217)
		if i < len(entries.Values) && func() bool {
			__antithesis_instrumentation__.Notify(56219)
			return entries.Values[i].Key.Equal(key) == true
		}() == true {
			__antithesis_instrumentation__.Notify(56220)
			res.Values = append(res.Values, entries.Values[i])
		} else {
			__antithesis_instrumentation__.Notify(56221)
		}
	}
	__antithesis_instrumentation__.Notify(56215)
	return res
}
