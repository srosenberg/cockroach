package changefeedccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/changefeedbase"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/lease"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/cache"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
)

type rowFetcherCache struct {
	codec           keys.SQLCodec
	leaseMgr        *lease.Manager
	fetchers        *cache.UnorderedCache
	watchedFamilies map[watchedFamily]struct{}

	collection *descs.Collection
	db         *kv.DB

	a tree.DatumAlloc
}

type cachedFetcher struct {
	tableDesc  catalog.TableDescriptor
	fetcher    row.Fetcher
	familyDesc descpb.ColumnFamilyDescriptor
	skip       bool
}

type watchedFamily struct {
	tableID    descpb.ID
	familyName string
}

var rfCacheConfig = cache.Config{
	Policy: cache.CacheFIFO,

	ShouldEvict: func(size int, _ interface{}, _ interface{}) bool {
		__antithesis_instrumentation__.Notify(17589)
		return size > 1024
	},
}

type idVersion struct {
	id      descpb.ID
	version descpb.DescriptorVersion
	family  descpb.FamilyID
}

func newRowFetcherCache(
	ctx context.Context,
	codec keys.SQLCodec,
	leaseMgr *lease.Manager,
	cf *descs.CollectionFactory,
	db *kv.DB,
	details jobspb.ChangefeedDetails,
) *rowFetcherCache {
	__antithesis_instrumentation__.Notify(17590)
	specs := details.TargetSpecifications
	watchedFamilies := make(map[watchedFamily]struct{}, len(specs))
	for _, s := range specs {
		__antithesis_instrumentation__.Notify(17592)
		watchedFamilies[watchedFamily{tableID: s.TableID, familyName: s.FamilyName}] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(17591)
	return &rowFetcherCache{
		codec:           codec,
		leaseMgr:        leaseMgr,
		collection:      cf.NewCollection(ctx, nil),
		db:              db,
		fetchers:        cache.NewUnorderedCache(rfCacheConfig),
		watchedFamilies: watchedFamilies,
	}
}

func (c *rowFetcherCache) TableDescForKey(
	ctx context.Context, key roachpb.Key, ts hlc.Timestamp,
) (catalog.TableDescriptor, descpb.FamilyID, error) {
	__antithesis_instrumentation__.Notify(17593)
	var tableDesc catalog.TableDescriptor
	key, err := c.codec.StripTenantPrefix(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(17600)
		return nil, descpb.FamilyID(0), err
	} else {
		__antithesis_instrumentation__.Notify(17601)
	}
	__antithesis_instrumentation__.Notify(17594)
	remaining, tableID, _, err := rowenc.DecodePartialTableIDIndexID(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(17602)
		return nil, descpb.FamilyID(0), err
	} else {
		__antithesis_instrumentation__.Notify(17603)
	}
	__antithesis_instrumentation__.Notify(17595)

	familyID, err := keys.DecodeFamilyKey(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(17604)
		return nil, descpb.FamilyID(0), err
	} else {
		__antithesis_instrumentation__.Notify(17605)
	}
	__antithesis_instrumentation__.Notify(17596)

	family := descpb.FamilyID(familyID)

	desc, err := c.leaseMgr.Acquire(ctx, ts, tableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(17606)

		return nil, family, changefeedbase.MarkRetryableError(err)
	} else {
		__antithesis_instrumentation__.Notify(17607)
	}
	__antithesis_instrumentation__.Notify(17597)
	tableDesc = desc.Underlying().(catalog.TableDescriptor)

	desc.Release(ctx)
	if tableDesc.ContainsUserDefinedTypes() {
		__antithesis_instrumentation__.Notify(17608)

		if err := c.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(17610)
			err := txn.SetFixedTimestamp(ctx, ts)
			if err != nil {
				__antithesis_instrumentation__.Notify(17612)
				return err
			} else {
				__antithesis_instrumentation__.Notify(17613)
			}
			__antithesis_instrumentation__.Notify(17611)
			tableDesc, err = c.collection.GetImmutableTableByID(ctx, txn, tableID, tree.ObjectLookupFlags{})
			return err
		}); err != nil {
			__antithesis_instrumentation__.Notify(17614)

			return nil, family, changefeedbase.MarkRetryableError(err)
		} else {
			__antithesis_instrumentation__.Notify(17615)
		}
		__antithesis_instrumentation__.Notify(17609)

		c.collection.ReleaseAll(ctx)
	} else {
		__antithesis_instrumentation__.Notify(17616)
	}
	__antithesis_instrumentation__.Notify(17598)

	for skippedCols := 0; skippedCols < tableDesc.GetPrimaryIndex().NumKeyColumns(); skippedCols++ {
		__antithesis_instrumentation__.Notify(17617)
		l, err := encoding.PeekLength(remaining)
		if err != nil {
			__antithesis_instrumentation__.Notify(17619)
			return nil, family, err
		} else {
			__antithesis_instrumentation__.Notify(17620)
		}
		__antithesis_instrumentation__.Notify(17618)
		remaining = remaining[l:]
	}
	__antithesis_instrumentation__.Notify(17599)

	return tableDesc, family, nil
}

var ErrUnwatchedFamily = errors.New("watched table but unwatched family")

func (c *rowFetcherCache) RowFetcherForColumnFamily(
	tableDesc catalog.TableDescriptor, family descpb.FamilyID,
) (*row.Fetcher, error) {
	__antithesis_instrumentation__.Notify(17621)
	idVer := idVersion{id: tableDesc.GetID(), version: tableDesc.GetVersion(), family: family}
	if v, ok := c.fetchers.Get(idVer); ok {
		__antithesis_instrumentation__.Notify(17627)
		f := v.(*cachedFetcher)
		if f.skip {
			__antithesis_instrumentation__.Notify(17629)
			return &f.fetcher, ErrUnwatchedFamily
		} else {
			__antithesis_instrumentation__.Notify(17630)
		}
		__antithesis_instrumentation__.Notify(17628)

		if catalog.UserDefinedTypeColsHaveSameVersion(tableDesc, f.tableDesc) {
			__antithesis_instrumentation__.Notify(17631)
			return &f.fetcher, nil
		} else {
			__antithesis_instrumentation__.Notify(17632)
		}
	} else {
		__antithesis_instrumentation__.Notify(17633)
	}
	__antithesis_instrumentation__.Notify(17622)

	familyDesc, err := tableDesc.FindFamilyByID(family)
	if err != nil {
		__antithesis_instrumentation__.Notify(17634)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(17635)
	}
	__antithesis_instrumentation__.Notify(17623)

	f := &cachedFetcher{
		tableDesc:  tableDesc,
		familyDesc: *familyDesc,
	}
	rf := &f.fetcher

	_, wholeTableWatched := c.watchedFamilies[watchedFamily{tableID: tableDesc.GetID()}]
	if !wholeTableWatched {
		__antithesis_instrumentation__.Notify(17636)
		_, familyWatched := c.watchedFamilies[watchedFamily{tableID: tableDesc.GetID(), familyName: familyDesc.Name}]
		if !familyWatched {
			__antithesis_instrumentation__.Notify(17637)
			f.skip = true
			return rf, ErrUnwatchedFamily
		} else {
			__antithesis_instrumentation__.Notify(17638)
		}
	} else {
		__antithesis_instrumentation__.Notify(17639)
	}
	__antithesis_instrumentation__.Notify(17624)

	var spec descpb.IndexFetchSpec

	if err := rowenc.InitIndexFetchSpec(
		&spec, c.codec, tableDesc, tableDesc.GetPrimaryIndex(), tableDesc.PublicColumnIDs(),
	); err != nil {
		__antithesis_instrumentation__.Notify(17640)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(17641)
	}
	__antithesis_instrumentation__.Notify(17625)

	if err := rf.Init(
		context.TODO(),
		false,
		descpb.ScanLockingStrength_FOR_NONE,
		descpb.ScanLockingWaitPolicy_BLOCK,
		0,
		&c.a,
		nil,
		&spec,
	); err != nil {
		__antithesis_instrumentation__.Notify(17642)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(17643)
	}
	__antithesis_instrumentation__.Notify(17626)

	rf.IgnoreUnexpectedNulls = true

	c.fetchers.Add(idVer, f)
	return rf, nil
}
