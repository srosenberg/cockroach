package tabledesc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/geo/geoindex"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/iterutil"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

var _ catalog.Index = (*index)(nil)

type index struct {
	maybeMutation
	desc    *descpb.IndexDescriptor
	ordinal int
}

func (w index) IndexDesc() *descpb.IndexDescriptor {
	__antithesis_instrumentation__.Notify(268751)
	return w.desc
}

func (w index) IndexDescDeepCopy() descpb.IndexDescriptor {
	__antithesis_instrumentation__.Notify(268752)
	return *protoutil.Clone(w.desc).(*descpb.IndexDescriptor)
}

func (w index) Ordinal() int {
	__antithesis_instrumentation__.Notify(268753)
	return w.ordinal
}

func (w index) Primary() bool {
	__antithesis_instrumentation__.Notify(268754)
	return w.ordinal == 0
}

func (w index) Public() bool {
	__antithesis_instrumentation__.Notify(268755)
	return !w.IsMutation()
}

func (w index) GetID() descpb.IndexID {
	__antithesis_instrumentation__.Notify(268756)
	return w.desc.ID
}

func (w index) GetConstraintID() descpb.ConstraintID {
	__antithesis_instrumentation__.Notify(268757)
	return w.desc.ConstraintID
}

func (w index) GetName() string {
	__antithesis_instrumentation__.Notify(268758)
	return w.desc.Name
}

func (w index) IsPartial() bool {
	__antithesis_instrumentation__.Notify(268759)
	return w.desc.IsPartial()
}

func (w index) IsUnique() bool {
	__antithesis_instrumentation__.Notify(268760)
	return w.desc.Unique
}

func (w index) IsDisabled() bool {
	__antithesis_instrumentation__.Notify(268761)
	return w.desc.Disabled
}

func (w index) IsSharded() bool {
	__antithesis_instrumentation__.Notify(268762)
	return w.desc.IsSharded()
}

func (w index) IsCreatedExplicitly() bool {
	__antithesis_instrumentation__.Notify(268763)
	return w.desc.CreatedExplicitly
}

func (w index) GetPredicate() string {
	__antithesis_instrumentation__.Notify(268764)
	return w.desc.Predicate
}

func (w index) GetType() descpb.IndexDescriptor_Type {
	__antithesis_instrumentation__.Notify(268765)
	return w.desc.Type
}

func (w index) GetPartitioning() catalog.Partitioning {
	__antithesis_instrumentation__.Notify(268766)
	return &partitioning{desc: &w.desc.Partitioning}
}

func (w index) ExplicitColumnStartIdx() int {
	__antithesis_instrumentation__.Notify(268767)
	return w.desc.ExplicitColumnStartIdx()
}

func (w index) IsValidOriginIndex(originColIDs descpb.ColumnIDs) bool {
	__antithesis_instrumentation__.Notify(268768)
	return w.desc.IsValidOriginIndex(originColIDs)
}

func (w index) IsValidReferencedUniqueConstraint(referencedColIDs descpb.ColumnIDs) bool {
	__antithesis_instrumentation__.Notify(268769)
	return w.desc.IsValidReferencedUniqueConstraint(referencedColIDs)
}

func (w index) HasOldStoredColumns() bool {
	__antithesis_instrumentation__.Notify(268770)
	return w.NumKeySuffixColumns() > 0 && func() bool {
		__antithesis_instrumentation__.Notify(268771)
		return !w.Primary() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(268772)
		return len(w.desc.StoreColumnIDs) < len(w.desc.StoreColumnNames) == true
	}() == true
}

func (w index) InvertedColumnID() descpb.ColumnID {
	__antithesis_instrumentation__.Notify(268773)
	return w.desc.InvertedColumnID()
}

func (w index) InvertedColumnName() string {
	__antithesis_instrumentation__.Notify(268774)
	return w.desc.InvertedColumnName()
}

func (w index) InvertedColumnKeyType() *types.T {
	__antithesis_instrumentation__.Notify(268775)
	return w.desc.InvertedColumnKeyType()
}

func (w index) CollectKeyColumnIDs() catalog.TableColSet {
	__antithesis_instrumentation__.Notify(268776)
	return catalog.MakeTableColSet(w.desc.KeyColumnIDs...)
}

func (w index) CollectPrimaryStoredColumnIDs() catalog.TableColSet {
	__antithesis_instrumentation__.Notify(268777)
	if !w.Primary() {
		__antithesis_instrumentation__.Notify(268779)
		return catalog.TableColSet{}
	} else {
		__antithesis_instrumentation__.Notify(268780)
	}
	__antithesis_instrumentation__.Notify(268778)
	return catalog.MakeTableColSet(w.desc.StoreColumnIDs...)
}

func (w index) CollectSecondaryStoredColumnIDs() catalog.TableColSet {
	__antithesis_instrumentation__.Notify(268781)
	if w.Primary() {
		__antithesis_instrumentation__.Notify(268783)
		return catalog.TableColSet{}
	} else {
		__antithesis_instrumentation__.Notify(268784)
	}
	__antithesis_instrumentation__.Notify(268782)
	return catalog.MakeTableColSet(w.desc.StoreColumnIDs...)
}

func (w index) CollectKeySuffixColumnIDs() catalog.TableColSet {
	__antithesis_instrumentation__.Notify(268785)
	return catalog.MakeTableColSet(w.desc.KeySuffixColumnIDs...)
}

func (w index) CollectCompositeColumnIDs() catalog.TableColSet {
	__antithesis_instrumentation__.Notify(268786)
	return catalog.MakeTableColSet(w.desc.CompositeColumnIDs...)
}

func (w index) GetGeoConfig() geoindex.Config {
	__antithesis_instrumentation__.Notify(268787)
	return w.desc.GeoConfig
}

func (w index) GetSharded() catpb.ShardedDescriptor {
	__antithesis_instrumentation__.Notify(268788)
	return w.desc.Sharded
}

func (w index) GetShardColumnName() string {
	__antithesis_instrumentation__.Notify(268789)
	return w.desc.Sharded.Name
}

func (w index) GetVersion() descpb.IndexDescriptorVersion {
	__antithesis_instrumentation__.Notify(268790)
	return w.desc.Version
}

func (w index) GetEncodingType() descpb.IndexDescriptorEncodingType {
	__antithesis_instrumentation__.Notify(268791)
	if w.Primary() {
		__antithesis_instrumentation__.Notify(268793)

		return descpb.PrimaryIndexEncoding
	} else {
		__antithesis_instrumentation__.Notify(268794)
	}
	__antithesis_instrumentation__.Notify(268792)
	return w.desc.EncodingType
}

func (w index) NumKeyColumns() int {
	__antithesis_instrumentation__.Notify(268795)
	return len(w.desc.KeyColumnIDs)
}

func (w index) GetKeyColumnID(columnOrdinal int) descpb.ColumnID {
	__antithesis_instrumentation__.Notify(268796)
	return w.desc.KeyColumnIDs[columnOrdinal]
}

func (w index) GetKeyColumnName(columnOrdinal int) string {
	__antithesis_instrumentation__.Notify(268797)
	return w.desc.KeyColumnNames[columnOrdinal]
}

func (w index) GetKeyColumnDirection(columnOrdinal int) descpb.IndexDescriptor_Direction {
	__antithesis_instrumentation__.Notify(268798)
	return w.desc.KeyColumnDirections[columnOrdinal]
}

func (w index) NumPrimaryStoredColumns() int {
	__antithesis_instrumentation__.Notify(268799)
	if !w.Primary() {
		__antithesis_instrumentation__.Notify(268801)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(268802)
	}
	__antithesis_instrumentation__.Notify(268800)
	return len(w.desc.StoreColumnIDs)
}

func (w index) NumSecondaryStoredColumns() int {
	__antithesis_instrumentation__.Notify(268803)
	if w.Primary() {
		__antithesis_instrumentation__.Notify(268805)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(268806)
	}
	__antithesis_instrumentation__.Notify(268804)
	return len(w.desc.StoreColumnIDs)
}

func (w index) GetStoredColumnID(storedColumnOrdinal int) descpb.ColumnID {
	__antithesis_instrumentation__.Notify(268807)
	return w.desc.StoreColumnIDs[storedColumnOrdinal]
}

func (w index) GetStoredColumnName(storedColumnOrdinal int) string {
	__antithesis_instrumentation__.Notify(268808)
	return w.desc.StoreColumnNames[storedColumnOrdinal]
}

func (w index) NumKeySuffixColumns() int {
	__antithesis_instrumentation__.Notify(268809)
	return len(w.desc.KeySuffixColumnIDs)
}

func (w index) GetKeySuffixColumnID(keySuffixColumnOrdinal int) descpb.ColumnID {
	__antithesis_instrumentation__.Notify(268810)
	return w.desc.KeySuffixColumnIDs[keySuffixColumnOrdinal]
}

func (w index) NumCompositeColumns() int {
	__antithesis_instrumentation__.Notify(268811)
	return len(w.desc.CompositeColumnIDs)
}

func (w index) GetCompositeColumnID(compositeColumnOrdinal int) descpb.ColumnID {
	__antithesis_instrumentation__.Notify(268812)
	return w.desc.CompositeColumnIDs[compositeColumnOrdinal]
}

func (w index) UseDeletePreservingEncoding() bool {
	__antithesis_instrumentation__.Notify(268813)
	return w.desc.UseDeletePreservingEncoding && func() bool {
		__antithesis_instrumentation__.Notify(268814)
		return !w.maybeMutation.DeleteOnly() == true
	}() == true
}

func (w index) ForcePut() bool {
	__antithesis_instrumentation__.Notify(268815)
	return w.Merging() || func() bool {
		__antithesis_instrumentation__.Notify(268816)
		return w.desc.UseDeletePreservingEncoding == true
	}() == true
}

func (w index) CreatedAt() time.Time {
	__antithesis_instrumentation__.Notify(268817)
	if w.desc.CreatedAtNanos == 0 {
		__antithesis_instrumentation__.Notify(268819)
		return time.Time{}
	} else {
		__antithesis_instrumentation__.Notify(268820)
	}
	__antithesis_instrumentation__.Notify(268818)
	return timeutil.Unix(0, w.desc.CreatedAtNanos)
}

func (w index) IsTemporaryIndexForBackfill() bool {
	__antithesis_instrumentation__.Notify(268821)
	return w.desc.UseDeletePreservingEncoding
}

type partitioning struct {
	desc *catpb.PartitioningDescriptor
}

func (p partitioning) PartitioningDesc() *catpb.PartitioningDescriptor {
	__antithesis_instrumentation__.Notify(268822)
	return p.desc
}

func (p partitioning) DeepCopy() catalog.Partitioning {
	__antithesis_instrumentation__.Notify(268823)
	return &partitioning{desc: protoutil.Clone(p.desc).(*catpb.PartitioningDescriptor)}
}

func (p partitioning) FindPartitionByName(name string) (found catalog.Partitioning) {
	__antithesis_instrumentation__.Notify(268824)
	_ = p.forEachPartitionName(func(partitioning catalog.Partitioning, currentName string) error {
		__antithesis_instrumentation__.Notify(268826)
		if name == currentName {
			__antithesis_instrumentation__.Notify(268828)
			found = partitioning
			return iterutil.StopIteration()
		} else {
			__antithesis_instrumentation__.Notify(268829)
		}
		__antithesis_instrumentation__.Notify(268827)
		return nil
	})
	__antithesis_instrumentation__.Notify(268825)
	return found
}

func (p partitioning) ForEachPartitionName(fn func(name string) error) error {
	__antithesis_instrumentation__.Notify(268830)
	err := p.forEachPartitionName(func(_ catalog.Partitioning, name string) error {
		__antithesis_instrumentation__.Notify(268833)
		return fn(name)
	})
	__antithesis_instrumentation__.Notify(268831)
	if iterutil.Done(err) {
		__antithesis_instrumentation__.Notify(268834)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(268835)
	}
	__antithesis_instrumentation__.Notify(268832)
	return err
}

func (p partitioning) forEachPartitionName(
	fn func(partitioning catalog.Partitioning, name string) error,
) error {
	__antithesis_instrumentation__.Notify(268836)
	for _, l := range p.desc.List {
		__antithesis_instrumentation__.Notify(268839)
		err := fn(p, l.Name)
		if err != nil {
			__antithesis_instrumentation__.Notify(268841)
			return err
		} else {
			__antithesis_instrumentation__.Notify(268842)
		}
		__antithesis_instrumentation__.Notify(268840)
		err = partitioning{desc: &l.Subpartitioning}.forEachPartitionName(fn)
		if err != nil {
			__antithesis_instrumentation__.Notify(268843)
			return err
		} else {
			__antithesis_instrumentation__.Notify(268844)
		}
	}
	__antithesis_instrumentation__.Notify(268837)
	for _, r := range p.desc.Range {
		__antithesis_instrumentation__.Notify(268845)
		err := fn(p, r.Name)
		if err != nil {
			__antithesis_instrumentation__.Notify(268846)
			return err
		} else {
			__antithesis_instrumentation__.Notify(268847)
		}
	}
	__antithesis_instrumentation__.Notify(268838)
	return nil
}

func (p partitioning) NumLists() int {
	__antithesis_instrumentation__.Notify(268848)
	return len(p.desc.List)
}

func (p partitioning) NumRanges() int {
	__antithesis_instrumentation__.Notify(268849)
	return len(p.desc.Range)
}

func (p partitioning) ForEachList(
	fn func(name string, values [][]byte, subPartitioning catalog.Partitioning) error,
) error {
	__antithesis_instrumentation__.Notify(268850)
	for _, l := range p.desc.List {
		__antithesis_instrumentation__.Notify(268852)
		subp := partitioning{desc: &l.Subpartitioning}
		err := fn(l.Name, l.Values, subp)
		if err != nil {
			__antithesis_instrumentation__.Notify(268853)
			if iterutil.Done(err) {
				__antithesis_instrumentation__.Notify(268855)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(268856)
			}
			__antithesis_instrumentation__.Notify(268854)
			return err
		} else {
			__antithesis_instrumentation__.Notify(268857)
		}
	}
	__antithesis_instrumentation__.Notify(268851)
	return nil
}

func (p partitioning) ForEachRange(fn func(name string, from, to []byte) error) error {
	__antithesis_instrumentation__.Notify(268858)
	for _, r := range p.desc.Range {
		__antithesis_instrumentation__.Notify(268860)
		err := fn(r.Name, r.FromInclusive, r.ToExclusive)
		if err != nil {
			__antithesis_instrumentation__.Notify(268861)
			if iterutil.Done(err) {
				__antithesis_instrumentation__.Notify(268863)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(268864)
			}
			__antithesis_instrumentation__.Notify(268862)
			return err
		} else {
			__antithesis_instrumentation__.Notify(268865)
		}
	}
	__antithesis_instrumentation__.Notify(268859)
	return nil
}

func (p partitioning) NumColumns() int {
	__antithesis_instrumentation__.Notify(268866)
	return int(p.desc.NumColumns)
}

func (p partitioning) NumImplicitColumns() int {
	__antithesis_instrumentation__.Notify(268867)
	return int(p.desc.NumImplicitColumns)
}

type indexCache struct {
	primary              catalog.Index
	all                  []catalog.Index
	active               []catalog.Index
	nonDrop              []catalog.Index
	nonPrimary           []catalog.Index
	publicNonPrimary     []catalog.Index
	writableNonPrimary   []catalog.Index
	deletableNonPrimary  []catalog.Index
	deleteOnlyNonPrimary []catalog.Index
	partial              []catalog.Index
}

func newIndexCache(desc *descpb.TableDescriptor, mutations *mutationCache) *indexCache {
	__antithesis_instrumentation__.Notify(268868)
	c := indexCache{}

	numPublic := 1 + len(desc.Indexes)
	backingStructs := make([]index, numPublic)
	backingStructs[0] = index{desc: &desc.PrimaryIndex}
	for i := range desc.Indexes {
		__antithesis_instrumentation__.Notify(268875)
		backingStructs[i+1] = index{desc: &desc.Indexes[i], ordinal: i + 1}
	}
	__antithesis_instrumentation__.Notify(268869)

	numMutations := len(mutations.indexes)
	c.all = make([]catalog.Index, numPublic, numPublic+numMutations)
	for i := range backingStructs {
		__antithesis_instrumentation__.Notify(268876)
		c.all[i] = &backingStructs[i]
	}
	__antithesis_instrumentation__.Notify(268870)
	for _, m := range mutations.indexes {
		__antithesis_instrumentation__.Notify(268877)
		c.all = append(c.all, m.AsIndex())
	}
	__antithesis_instrumentation__.Notify(268871)

	c.primary = c.all[0]
	c.active = c.all[:numPublic]
	c.publicNonPrimary = c.active[1:]
	for _, idx := range c.all[1:] {
		__antithesis_instrumentation__.Notify(268878)
		if !idx.Backfilling() {
			__antithesis_instrumentation__.Notify(268880)
			lazyAllocAppendIndex(&c.deletableNonPrimary, idx, len(c.all[1:]))
		} else {
			__antithesis_instrumentation__.Notify(268881)
		}
		__antithesis_instrumentation__.Notify(268879)
		lazyAllocAppendIndex(&c.nonPrimary, idx, len(c.all[1:]))
	}
	__antithesis_instrumentation__.Notify(268872)

	if numMutations == 0 {
		__antithesis_instrumentation__.Notify(268882)
		c.writableNonPrimary = c.publicNonPrimary
	} else {
		__antithesis_instrumentation__.Notify(268883)
		for _, idx := range c.deletableNonPrimary {
			__antithesis_instrumentation__.Notify(268884)
			if idx.DeleteOnly() {
				__antithesis_instrumentation__.Notify(268885)
				lazyAllocAppendIndex(&c.deleteOnlyNonPrimary, idx, numMutations)
			} else {
				__antithesis_instrumentation__.Notify(268886)
				lazyAllocAppendIndex(&c.writableNonPrimary, idx, len(c.deletableNonPrimary))
			}
		}
	}
	__antithesis_instrumentation__.Notify(268873)
	for _, idx := range c.all {
		__antithesis_instrumentation__.Notify(268887)
		if !idx.Dropped() && func() bool {
			__antithesis_instrumentation__.Notify(268889)
			return (!idx.Primary() || func() bool {
				__antithesis_instrumentation__.Notify(268890)
				return desc.IsPhysicalTable() == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(268891)
			lazyAllocAppendIndex(&c.nonDrop, idx, len(c.all))
		} else {
			__antithesis_instrumentation__.Notify(268892)
		}
		__antithesis_instrumentation__.Notify(268888)

		if idx.IsPartial() && func() bool {
			__antithesis_instrumentation__.Notify(268893)
			return !idx.Backfilling() == true
		}() == true {
			__antithesis_instrumentation__.Notify(268894)
			lazyAllocAppendIndex(&c.partial, idx, len(c.all))
		} else {
			__antithesis_instrumentation__.Notify(268895)
		}
	}
	__antithesis_instrumentation__.Notify(268874)
	return &c
}

func lazyAllocAppendIndex(slice *[]catalog.Index, idx catalog.Index, cap int) {
	__antithesis_instrumentation__.Notify(268896)
	if *slice == nil {
		__antithesis_instrumentation__.Notify(268898)
		*slice = make([]catalog.Index, 0, cap)
	} else {
		__antithesis_instrumentation__.Notify(268899)
	}
	__antithesis_instrumentation__.Notify(268897)
	*slice = append(*slice, idx)
}

func ForeignKeyConstraintName(fromTable string, columnNames []string) string {
	__antithesis_instrumentation__.Notify(268900)
	return fmt.Sprintf("%s_%s_fkey", fromTable, strings.Join(columnNames, "_"))
}
