package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"regexp"

	"github.com/cockroachdb/cockroach/pkg/util/cache"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

type RegexpCacheKey interface {
	Pattern() (string, error)
}

type RegexpCache struct {
	mu    syncutil.Mutex
	cache *cache.UnorderedCache
}

func NewRegexpCache(size int) *RegexpCache {
	__antithesis_instrumentation__.Notify(612832)
	return &RegexpCache{
		cache: cache.NewUnorderedCache(cache.Config{
			Policy: cache.CacheLRU,
			ShouldEvict: func(s int, key, value interface{}) bool {
				__antithesis_instrumentation__.Notify(612833)
				return s > size
			},
		}),
	}
}

func (rc *RegexpCache) GetRegexp(key RegexpCacheKey) (*regexp.Regexp, error) {
	__antithesis_instrumentation__.Notify(612834)
	if rc != nil {
		__antithesis_instrumentation__.Notify(612839)
		re := rc.lookup(key)
		if re != nil {
			__antithesis_instrumentation__.Notify(612840)
			return re, nil
		} else {
			__antithesis_instrumentation__.Notify(612841)
		}
	} else {
		__antithesis_instrumentation__.Notify(612842)
	}
	__antithesis_instrumentation__.Notify(612835)

	pattern, err := key.Pattern()
	if err != nil {
		__antithesis_instrumentation__.Notify(612843)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(612844)
	}
	__antithesis_instrumentation__.Notify(612836)

	re, err := regexp.Compile(pattern)
	if err != nil {
		__antithesis_instrumentation__.Notify(612845)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(612846)
	}
	__antithesis_instrumentation__.Notify(612837)

	if rc != nil {
		__antithesis_instrumentation__.Notify(612847)
		rc.update(key, re)
	} else {
		__antithesis_instrumentation__.Notify(612848)
	}
	__antithesis_instrumentation__.Notify(612838)
	return re, nil
}

func (rc *RegexpCache) lookup(key RegexpCacheKey) *regexp.Regexp {
	__antithesis_instrumentation__.Notify(612849)
	rc.mu.Lock()
	defer rc.mu.Unlock()
	v, ok := rc.cache.Get(key)
	if !ok {
		__antithesis_instrumentation__.Notify(612851)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(612852)
	}
	__antithesis_instrumentation__.Notify(612850)
	return v.(*regexp.Regexp)
}

func (rc *RegexpCache) update(key RegexpCacheKey, re *regexp.Regexp) {
	__antithesis_instrumentation__.Notify(612853)
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.cache.Del(key)
	if re != nil {
		__antithesis_instrumentation__.Notify(612854)
		rc.cache.Add(key, re)
	} else {
		__antithesis_instrumentation__.Notify(612855)
	}
}

func (rc *RegexpCache) Len() int {
	__antithesis_instrumentation__.Notify(612856)
	if rc == nil {
		__antithesis_instrumentation__.Notify(612858)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(612859)
	}
	__antithesis_instrumentation__.Notify(612857)
	rc.mu.Lock()
	defer rc.mu.Unlock()
	return rc.cache.Len()
}
