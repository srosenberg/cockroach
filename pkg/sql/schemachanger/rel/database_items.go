package rel

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"
	"sync"

	"github.com/google/btree"
)

type item interface {
	btree.Item
	getIndexSpec() *indexSpec
	compareAttrs() ordinalSet
	getValues() *valuesMap
}

var _ item = (*containerItem)(nil)
var _ item = (*valuesItem)(nil)

func compareItems(a, b item) (less bool) {
	__antithesis_instrumentation__.Notify(578429)

	index := a.getIndexSpec()
	toCompare := a.compareAttrs().intersection(b.compareAttrs())
	for _, at := range index.attrs {
		__antithesis_instrumentation__.Notify(578433)
		if !toCompare.contains(at) {
			__antithesis_instrumentation__.Notify(578435)
			break
		} else {
			__antithesis_instrumentation__.Notify(578436)
		}
		__antithesis_instrumentation__.Notify(578434)

		var eq bool
		if less, eq = compareOn(
			at, a.getValues(), b.getValues(),
		); !eq {
			__antithesis_instrumentation__.Notify(578437)
			return less
		} else {
			__antithesis_instrumentation__.Notify(578438)
		}
	}
	__antithesis_instrumentation__.Notify(578430)

	if aValuesItem, ok := a.(*valuesItem); ok {
		__antithesis_instrumentation__.Notify(578439)
		return !aValuesItem.end
	} else {
		__antithesis_instrumentation__.Notify(578440)
	}
	__antithesis_instrumentation__.Notify(578431)
	if bValuesItem, ok := b.(*valuesItem); ok {
		__antithesis_instrumentation__.Notify(578441)
		return bValuesItem.end
	} else {
		__antithesis_instrumentation__.Notify(578442)
	}
	__antithesis_instrumentation__.Notify(578432)

	less, _ = compareEntities(
		a.(*containerItem).entity,
		b.(*containerItem).entity,
	)
	return less
}

type containerItem struct {
	*indexSpec
	*entity
}

func (c *containerItem) getValues() *valuesMap {
	__antithesis_instrumentation__.Notify(578443)
	return c.asMap()
}

func (c *containerItem) compareAttrs() ordinalSet {
	__antithesis_instrumentation__.Notify(578444)
	return math.MaxUint64
}
func (c *containerItem) getIndexSpec() *indexSpec {
	__antithesis_instrumentation__.Notify(578445)
	return c.indexSpec
}

func (c *containerItem) Less(than btree.Item) bool {
	__antithesis_instrumentation__.Notify(578446)
	return compareItems(c, than.(item))
}

type valuesItem struct {
	*indexSpec
	*valuesMap
	m   ordinalSet
	end bool
}

func (v *valuesItem) getIndexSpec() *indexSpec {
	__antithesis_instrumentation__.Notify(578447)
	return v.indexSpec
}
func (v *valuesItem) compareAttrs() ordinalSet {
	__antithesis_instrumentation__.Notify(578448)
	return v.m
}
func (v *valuesItem) getValues() *valuesMap {
	__antithesis_instrumentation__.Notify(578449)
	return v.valuesMap
}

var valuesItemPool = sync.Pool{
	New: func() interface{} { __antithesis_instrumentation__.Notify(578450); return new(valuesItem) },
}

func getValuesItems(idx *indexSpec, values *valuesMap, m ordinalSet) (from, to *valuesItem) {
	__antithesis_instrumentation__.Notify(578451)
	from = valuesItemPool.Get().(*valuesItem)
	to = valuesItemPool.Get().(*valuesItem)
	*from = valuesItem{indexSpec: idx, valuesMap: values, m: m, end: false}
	*to = valuesItem{indexSpec: idx, valuesMap: values, m: m, end: true}
	return from, to
}

func putValuesItems(from, to *valuesItem) {
	__antithesis_instrumentation__.Notify(578452)
	*from = valuesItem{}
	*to = valuesItem{}
	valuesItemPool.Put(from)
	valuesItemPool.Put(to)
}

func (v *valuesItem) Less(than btree.Item) bool {
	__antithesis_instrumentation__.Notify(578453)
	return compareItems(v, than.(item))
}
