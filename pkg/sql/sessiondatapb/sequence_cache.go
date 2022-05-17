package sessiondatapb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type SequenceCache map[uint32]*SequenceCacheEntry

func (sc SequenceCache) NextValue(
	seqID uint32, clientVersion uint32, fetchNextValues func() (int64, int64, int64, error),
) (int64, error) {
	__antithesis_instrumentation__.Notify(619856)

	if _, found := sc[seqID]; !found {
		__antithesis_instrumentation__.Notify(619860)
		sc[seqID] = &SequenceCacheEntry{}
	} else {
		__antithesis_instrumentation__.Notify(619861)
	}
	__antithesis_instrumentation__.Notify(619857)
	cacheEntry := sc[seqID]

	if cacheEntry.NumValues > 0 && func() bool {
		__antithesis_instrumentation__.Notify(619862)
		return cacheEntry.CachedVersion == clientVersion == true
	}() == true {
		__antithesis_instrumentation__.Notify(619863)
		cacheEntry.CurrentValue += cacheEntry.Increment
		cacheEntry.NumValues--
		return cacheEntry.CurrentValue - cacheEntry.Increment, nil
	} else {
		__antithesis_instrumentation__.Notify(619864)
	}
	__antithesis_instrumentation__.Notify(619858)

	currentValue, increment, numValues, err := fetchNextValues()
	if err != nil {
		__antithesis_instrumentation__.Notify(619865)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(619866)
	}
	__antithesis_instrumentation__.Notify(619859)

	val := currentValue
	cacheEntry.CurrentValue = currentValue + increment
	cacheEntry.Increment = increment
	cacheEntry.NumValues = numValues - 1
	cacheEntry.CachedVersion = clientVersion
	return val, nil
}
