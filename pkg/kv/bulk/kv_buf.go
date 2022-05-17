package bulk

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/errors"
)

type kvBuf struct {
	entries []kvBufEntry
	slab    []byte
}

type kvBufEntry struct {
	keySpan uint64
	valSpan uint64
}

const entrySizeShift = 4
const minEntryGrow = 1 << 14
const maxEntryGrow = (4 << 20) >> entrySizeShift
const minSlabGrow = 512 << 10
const maxSlabGrow = 64 << 20

const (
	lenBits, lenMask  = 28, 1<<lenBits - 1
	maxLen, maxOffset = lenMask, 1<<(64-lenBits) - 1
)

func (b *kvBuf) fits(ctx context.Context, toAdd sz, maxUsed sz, acc *mon.BoundAccount) bool {
	__antithesis_instrumentation__.Notify(86327)
	if len(b.entries) < cap(b.entries) && func() bool {
		__antithesis_instrumentation__.Notify(86337)
		return sz(len(b.slab))+toAdd < sz(cap(b.slab)) == true
	}() == true {
		__antithesis_instrumentation__.Notify(86338)
		return true
	} else {
		__antithesis_instrumentation__.Notify(86339)
	}
	__antithesis_instrumentation__.Notify(86328)

	used := sz(acc.Used())
	remaining := maxUsed - used

	var entryGrow int
	var slabGrow sz

	if sz(len(b.slab))+toAdd > sz(cap(b.slab)) {
		__antithesis_instrumentation__.Notify(86340)
		slabGrow = sz(cap(b.slab))
		if slabGrow < minSlabGrow {
			__antithesis_instrumentation__.Notify(86345)
			slabGrow = minSlabGrow
		} else {
			__antithesis_instrumentation__.Notify(86346)
		}
		__antithesis_instrumentation__.Notify(86341)
		for slabGrow < toAdd {
			__antithesis_instrumentation__.Notify(86347)
			slabGrow += minSlabGrow
		}
		__antithesis_instrumentation__.Notify(86342)
		if slabGrow > maxSlabGrow {
			__antithesis_instrumentation__.Notify(86348)
			slabGrow = maxSlabGrow
		} else {
			__antithesis_instrumentation__.Notify(86349)
		}
		__antithesis_instrumentation__.Notify(86343)
		for slabGrow > remaining && func() bool {
			__antithesis_instrumentation__.Notify(86350)
			return slabGrow > minSlabGrow == true
		}() == true {
			__antithesis_instrumentation__.Notify(86351)
			slabGrow -= minSlabGrow
		}
		__antithesis_instrumentation__.Notify(86344)

		if slabGrow < minSlabGrow || func() bool {
			__antithesis_instrumentation__.Notify(86352)
			return slabGrow < toAdd == true
		}() == true {
			__antithesis_instrumentation__.Notify(86353)
			return false
		} else {
			__antithesis_instrumentation__.Notify(86354)
		}
	} else {
		__antithesis_instrumentation__.Notify(86355)
	}
	__antithesis_instrumentation__.Notify(86329)

	if len(b.entries) == cap(b.entries) {
		__antithesis_instrumentation__.Notify(86356)
		entryGrow = cap(b.entries)
		if entryGrow < minEntryGrow {
			__antithesis_instrumentation__.Notify(86357)
			entryGrow = minEntryGrow
		} else {
			__antithesis_instrumentation__.Notify(86358)
			if entryGrow > maxEntryGrow {
				__antithesis_instrumentation__.Notify(86359)
				entryGrow = maxEntryGrow
			} else {
				__antithesis_instrumentation__.Notify(86360)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(86361)
	}
	__antithesis_instrumentation__.Notify(86330)

	if slabGrow > 0 && func() bool {
		__antithesis_instrumentation__.Notify(86362)
		return len(b.slab) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(86363)
		for {
			__antithesis_instrumentation__.Notify(86364)
			fitsInNewSlab := int(float64(slabGrow) / (float64(len(b.slab)) / float64(len(b.entries))))
			if cap(b.entries)+entryGrow >= len(b.entries)+fitsInNewSlab {
				__antithesis_instrumentation__.Notify(86366)
				break
			} else {
				__antithesis_instrumentation__.Notify(86367)
			}
			__antithesis_instrumentation__.Notify(86365)
			entryGrow += minEntryGrow
			if sz(entryGrow<<entrySizeShift)+slabGrow > remaining {
				__antithesis_instrumentation__.Notify(86368)
				slabGrow -= minSlabGrow
			} else {
				__antithesis_instrumentation__.Notify(86369)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(86370)
	}
	__antithesis_instrumentation__.Notify(86331)

	if sz(len(b.slab))+toAdd > sz(cap(b.slab))+slabGrow {
		__antithesis_instrumentation__.Notify(86371)
		return false
	} else {
		__antithesis_instrumentation__.Notify(86372)
	}
	__antithesis_instrumentation__.Notify(86332)

	needed := sz(entryGrow<<entrySizeShift) + slabGrow
	if needed > remaining {
		__antithesis_instrumentation__.Notify(86373)
		return false
	} else {
		__antithesis_instrumentation__.Notify(86374)
	}
	__antithesis_instrumentation__.Notify(86333)
	if err := acc.Grow(ctx, int64(needed)); err != nil {
		__antithesis_instrumentation__.Notify(86375)
		return false
	} else {
		__antithesis_instrumentation__.Notify(86376)
	}
	__antithesis_instrumentation__.Notify(86334)

	if entryGrow > 0 {
		__antithesis_instrumentation__.Notify(86377)
		old := b.entries
		b.entries = make([]kvBufEntry, len(b.entries), cap(b.entries)+entryGrow)
		copy(b.entries, old)
	} else {
		__antithesis_instrumentation__.Notify(86378)
	}
	__antithesis_instrumentation__.Notify(86335)
	if slabGrow > 0 {
		__antithesis_instrumentation__.Notify(86379)
		old := b.slab
		b.slab = make([]byte, len(b.slab), sz(cap(b.slab))+slabGrow)
		copy(b.slab, old)
	} else {
		__antithesis_instrumentation__.Notify(86380)
	}
	__antithesis_instrumentation__.Notify(86336)
	return true
}

func (b *kvBuf) append(k, v []byte) error {
	__antithesis_instrumentation__.Notify(86381)
	if len(b.slab) > maxOffset {
		__antithesis_instrumentation__.Notify(86385)
		return errors.Errorf("buffer size %d exceeds limit %d", len(b.slab), maxOffset)
	} else {
		__antithesis_instrumentation__.Notify(86386)
	}
	__antithesis_instrumentation__.Notify(86382)
	if len(k) > maxLen {
		__antithesis_instrumentation__.Notify(86387)
		return errors.Errorf("length %d exceeds limit %d", len(k), maxLen)
	} else {
		__antithesis_instrumentation__.Notify(86388)
	}
	__antithesis_instrumentation__.Notify(86383)
	if len(v) > maxLen {
		__antithesis_instrumentation__.Notify(86389)
		return errors.Errorf("length %d exceeds limit %d", len(v), maxLen)
	} else {
		__antithesis_instrumentation__.Notify(86390)
	}
	__antithesis_instrumentation__.Notify(86384)

	var e kvBufEntry
	e.keySpan = uint64(len(b.slab)<<lenBits) | uint64(len(k)&lenMask)
	b.slab = append(b.slab, k...)
	e.valSpan = uint64(len(b.slab)<<lenBits) | uint64(len(v)&lenMask)
	b.slab = append(b.slab, v...)

	b.entries = append(b.entries, e)
	return nil
}

func (b *kvBuf) read(span uint64) []byte {
	__antithesis_instrumentation__.Notify(86391)
	length := span & lenMask
	if length == 0 {
		__antithesis_instrumentation__.Notify(86393)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(86394)
	}
	__antithesis_instrumentation__.Notify(86392)
	offset := span >> lenBits
	return b.slab[offset : offset+length]
}

func (b *kvBuf) Key(i int) roachpb.Key {
	__antithesis_instrumentation__.Notify(86395)
	return b.read(b.entries[i].keySpan)
}

func (b *kvBuf) Value(i int) []byte {
	__antithesis_instrumentation__.Notify(86396)
	return b.read(b.entries[i].valSpan)
}

func (b *kvBuf) Len() int {
	__antithesis_instrumentation__.Notify(86397)
	return len(b.entries)
}

func (b *kvBuf) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(86398)

	return bytes.Compare(b.read(b.entries[i].keySpan), b.read(b.entries[j].keySpan)) < 0
}

func (b *kvBuf) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(86399)
	b.entries[i], b.entries[j] = b.entries[j], b.entries[i]
}

func (b kvBuf) MemSize() sz {
	__antithesis_instrumentation__.Notify(86400)
	return sz(cap(b.entries)<<entrySizeShift) + sz(cap(b.slab))
}

func (b *kvBuf) KVSize() sz {
	__antithesis_instrumentation__.Notify(86401)
	return sz(len(b.slab))
}

func (b *kvBuf) unusedCap() (unusedEntryCap sz, unusedSlabCap sz) {
	__antithesis_instrumentation__.Notify(86402)
	unusedEntryCap = sz((cap(b.entries) - len(b.entries)) << entrySizeShift)
	unusedSlabCap = sz(cap(b.slab) - len(b.slab))
	return
}

func (b *kvBuf) Reset() {
	__antithesis_instrumentation__.Notify(86403)

	b.slab = b.slab[:0]
	b.entries = b.entries[:0]
}
