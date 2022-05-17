// Package hydratedtables contains logic to cache table descriptors with user
// defined types hydrated.
package hydratedtables

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"sync"

	"github.com/biogo/store/llrb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/lease"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/cache"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil/singleflight"
	io_prometheus_client "github.com/prometheus/client_model/go"
)

type Cache struct {
	settings *cluster.Settings
	g        singleflight.Group
	metrics  Metrics
	mu       struct {
		syncutil.Mutex
		cache *cache.OrderedCache
	}
}

func (c *Cache) Metrics() *Metrics {
	__antithesis_instrumentation__.Notify(265464)
	return &c.metrics
}

var _ metric.Struct = (*Metrics)(nil)

type Metrics struct {
	Hits   *metric.Counter
	Misses *metric.Counter
}

func makeMetrics() Metrics {
	__antithesis_instrumentation__.Notify(265465)
	return Metrics{
		Hits:   metric.NewCounter(metaHits),
		Misses: metric.NewCounter(metaMisses),
	}
}

func (m *Metrics) MetricStruct() { __antithesis_instrumentation__.Notify(265466) }

var (
	metaHits = metric.Metadata{
		Name:        "sql.hydrated_table_cache.hits",
		Help:        "counter on the number of cache hits",
		Measurement: "reads",
		Unit:        metric.Unit_COUNT,
		MetricType:  io_prometheus_client.MetricType_COUNTER,
	}
	metaMisses = metric.Metadata{
		Name:        "sql.hydrated_table_cache.misses",
		Help:        "counter on the number of cache misses",
		Measurement: "reads",
		Unit:        metric.Unit_COUNT,
		MetricType:  io_prometheus_client.MetricType_COUNTER,
	}
)

var CacheSize = settings.RegisterIntSetting(
	settings.TenantWritable,
	"sql.catalog.hydrated_tables.cache_size",
	"number of table descriptor versions retained in the hydratedtables LRU cache",
	128,
	settings.NonNegativeInt,
)

func NewCache(settings *cluster.Settings) *Cache {
	__antithesis_instrumentation__.Notify(265467)
	c := &Cache{
		settings: settings,
		metrics:  makeMetrics(),
	}
	c.mu.cache = cache.NewOrderedCache(cache.Config{
		Policy: cache.CacheLRU,
		ShouldEvict: func(size int, key, value interface{}) bool {
			__antithesis_instrumentation__.Notify(265469)
			return size > int(CacheSize.Get(&settings.SV))
		},
		OnEvicted: func(key, value interface{}) {
			__antithesis_instrumentation__.Notify(265470)
			putCacheKey(key.(*cacheKey))
		},
	})
	__antithesis_instrumentation__.Notify(265468)
	return c
}

type hydratedTableDescriptor struct {
	tableDesc catalog.TableDescriptor
	typeDescs []*cachedType
}

type cachedType struct {
	name    tree.TypeName
	typDesc catalog.TypeDescriptor
}

func (c *Cache) GetHydratedTableDescriptor(
	ctx context.Context, table catalog.TableDescriptor, res catalog.TypeDescriptorResolver,
) (hydrated catalog.TableDescriptor, err error) {
	__antithesis_instrumentation__.Notify(265471)

	if table.IsUncommittedVersion() {
		__antithesis_instrumentation__.Notify(265476)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(265477)
	}
	__antithesis_instrumentation__.Notify(265472)

	if !table.ContainsUserDefinedTypes() {
		__antithesis_instrumentation__.Notify(265478)
		return table, nil
	} else {
		__antithesis_instrumentation__.Notify(265479)
	}
	__antithesis_instrumentation__.Notify(265473)

	var groupKey string
	defer func() {
		__antithesis_instrumentation__.Notify(265480)
		if hydrated != nil {
			__antithesis_instrumentation__.Notify(265481)
			c.recordMetrics(groupKey == "")
		} else {
			__antithesis_instrumentation__.Notify(265482)
		}
	}()
	__antithesis_instrumentation__.Notify(265474)
	key := newCacheKey(table)
	defer func() {
		__antithesis_instrumentation__.Notify(265483)
		if key != nil {
			__antithesis_instrumentation__.Notify(265484)
			putCacheKey(key)
		} else {
			__antithesis_instrumentation__.Notify(265485)
		}
	}()
	__antithesis_instrumentation__.Notify(265475)

	for {
		__antithesis_instrumentation__.Notify(265486)
		if cached, ok := c.getFromCache(key); ok {
			__antithesis_instrumentation__.Notify(265492)
			canUse, skipCache, err := canUseCachedDescriptor(ctx, cached, res)
			if err != nil || func() bool {
				__antithesis_instrumentation__.Notify(265494)
				return skipCache == true
			}() == true {
				__antithesis_instrumentation__.Notify(265495)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(265496)
			}
			__antithesis_instrumentation__.Notify(265493)
			if canUse {
				__antithesis_instrumentation__.Notify(265497)
				return cached.tableDesc, nil
			} else {
				__antithesis_instrumentation__.Notify(265498)
			}
		} else {
			__antithesis_instrumentation__.Notify(265499)
		}
		__antithesis_instrumentation__.Notify(265487)

		if groupKey == "" {
			__antithesis_instrumentation__.Notify(265500)
			groupKey = fmt.Sprintf("%d@%d", key.ID, key.Version)
		} else {
			__antithesis_instrumentation__.Notify(265501)
		}
		__antithesis_instrumentation__.Notify(265488)

		var called bool
		res, _, err := c.g.Do(groupKey, func() (interface{}, error) {
			__antithesis_instrumentation__.Notify(265502)
			called = true
			cachedRes := cachedTypeDescriptorResolver{
				underlying: res,
				cache:      map[descpb.ID]*cachedType{},
			}
			descBase := protoutil.Clone(table.TableDesc()).(*descpb.TableDescriptor)
			if err := typedesc.HydrateTypesInTableDescriptor(ctx, descBase, &cachedRes); err != nil {
				__antithesis_instrumentation__.Notify(265505)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(265506)
			}
			__antithesis_instrumentation__.Notify(265503)
			hydrated := tabledesc.NewBuilder(descBase).BuildImmutableTable()

			if !cachedRes.haveUncommitted {
				__antithesis_instrumentation__.Notify(265507)
				c.addToCache(key, &hydratedTableDescriptor{
					tableDesc: hydrated,
					typeDescs: cachedRes.types,
				})

				key = nil
			} else {
				__antithesis_instrumentation__.Notify(265508)
			}
			__antithesis_instrumentation__.Notify(265504)

			return hydrated, nil
		})
		__antithesis_instrumentation__.Notify(265489)

		if !called {
			__antithesis_instrumentation__.Notify(265509)
			continue
		} else {
			__antithesis_instrumentation__.Notify(265510)
		}
		__antithesis_instrumentation__.Notify(265490)
		if err != nil {
			__antithesis_instrumentation__.Notify(265511)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(265512)
		}
		__antithesis_instrumentation__.Notify(265491)
		return res.(catalog.TableDescriptor), nil
	}
}

func canUseCachedDescriptor(
	ctx context.Context, cached *hydratedTableDescriptor, res catalog.TypeDescriptorResolver,
) (canUse, skipCache bool, _ error) {
	__antithesis_instrumentation__.Notify(265513)
	for _, typ := range cached.typeDescs {
		__antithesis_instrumentation__.Notify(265515)
		name, typDesc, err := res.GetTypeDescriptor(ctx, typ.typDesc.GetID())
		if err != nil {
			__antithesis_instrumentation__.Notify(265517)
			return false, false, err
		} else {
			__antithesis_instrumentation__.Notify(265518)
		}
		__antithesis_instrumentation__.Notify(265516)

		if isUncommittedVersion := typDesc.IsUncommittedVersion(); isUncommittedVersion || func() bool {
			__antithesis_instrumentation__.Notify(265519)
			return typ.typDesc.GetVersion() != typDesc.GetVersion() == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(265520)
			return (name.ObjectNamePrefix != (tree.ObjectNamePrefix{}) && func() bool {
				__antithesis_instrumentation__.Notify(265521)
				return typ.name != name == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(265522)

			skipCache = isUncommittedVersion || func() bool {
				__antithesis_instrumentation__.Notify(265523)
				return typ.typDesc.GetVersion() > typDesc.GetVersion() == true
			}() == true
			return false, skipCache, nil
		} else {
			__antithesis_instrumentation__.Notify(265524)
		}
	}
	__antithesis_instrumentation__.Notify(265514)
	return true, false, nil
}

func (c *Cache) getFromCache(key *cacheKey) (*hydratedTableDescriptor, bool) {
	__antithesis_instrumentation__.Notify(265525)
	c.mu.Lock()
	defer c.mu.Unlock()
	got, ok := c.mu.cache.Get(key)
	if !ok {
		__antithesis_instrumentation__.Notify(265527)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(265528)
	}
	__antithesis_instrumentation__.Notify(265526)
	return got.(*hydratedTableDescriptor), true
}

func (c *Cache) addToCache(key *cacheKey, toCache *hydratedTableDescriptor) {
	__antithesis_instrumentation__.Notify(265529)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.mu.cache.Add(key, toCache)
}

func (c *Cache) recordMetrics(hitCache bool) {
	__antithesis_instrumentation__.Notify(265530)
	if hitCache {
		__antithesis_instrumentation__.Notify(265531)
		c.metrics.Hits.Inc(1)
	} else {
		__antithesis_instrumentation__.Notify(265532)
		c.metrics.Misses.Inc(1)
	}
}

type cachedTypeDescriptorResolver struct {
	underlying      catalog.TypeDescriptorResolver
	cache           map[descpb.ID]*cachedType
	types           []*cachedType
	haveUncommitted bool
}

func (c *cachedTypeDescriptorResolver) GetTypeDescriptor(
	ctx context.Context, id descpb.ID,
) (tree.TypeName, catalog.TypeDescriptor, error) {
	__antithesis_instrumentation__.Notify(265533)
	if typ, exists := c.cache[id]; exists {
		__antithesis_instrumentation__.Notify(265536)
		return typ.name, typ.typDesc, nil
	} else {
		__antithesis_instrumentation__.Notify(265537)
	}
	__antithesis_instrumentation__.Notify(265534)
	name, typDesc, err := c.underlying.GetTypeDescriptor(ctx, id)
	if err != nil {
		__antithesis_instrumentation__.Notify(265538)
		return tree.TypeName{}, nil, err
	} else {
		__antithesis_instrumentation__.Notify(265539)
	}
	__antithesis_instrumentation__.Notify(265535)
	typ := &cachedType{
		name:    name,
		typDesc: typDesc,
	}
	c.cache[id] = typ
	c.types = append(c.types, typ)
	c.haveUncommitted = c.haveUncommitted || func() bool {
		__antithesis_instrumentation__.Notify(265540)
		return typDesc.IsUncommittedVersion() == true
	}() == true
	return name, typDesc, nil
}

type cacheKey lease.IDVersion

func (c cacheKey) Compare(comparable llrb.Comparable) int {
	__antithesis_instrumentation__.Notify(265541)
	o := comparable.(*cacheKey)
	switch {
	case c.ID < o.ID:
		__antithesis_instrumentation__.Notify(265542)
		return -1
	case c.ID > o.ID:
		__antithesis_instrumentation__.Notify(265543)
		return 1
	default:
		__antithesis_instrumentation__.Notify(265544)
		switch {
		case c.Version < o.Version:
			__antithesis_instrumentation__.Notify(265545)
			return -1
		case c.Version > o.Version:
			__antithesis_instrumentation__.Notify(265546)
			return 1
		default:
			__antithesis_instrumentation__.Notify(265547)
			return 0
		}
	}
}

var _ llrb.Comparable = (*cacheKey)(nil)

var cacheKeySyncPool = sync.Pool{
	New: func() interface{} { __antithesis_instrumentation__.Notify(265548); return new(cacheKey) },
}

func newCacheKey(table catalog.TableDescriptor) *cacheKey {
	__antithesis_instrumentation__.Notify(265549)
	k := cacheKeySyncPool.Get().(*cacheKey)
	*k = cacheKey{
		ID:      table.GetID(),
		Version: table.GetVersion(),
	}
	return k
}

func putCacheKey(k *cacheKey) {
	__antithesis_instrumentation__.Notify(265550)
	*k = cacheKey{}
	cacheKeySyncPool.Put(k)
}
