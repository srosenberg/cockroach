package storage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"encoding/binary"
	"math"

	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/pebble"
)

type BatchType byte

const (
	BatchTypeDeletion BatchType = 0x0
	BatchTypeValue    BatchType = 0x1
	BatchTypeMerge    BatchType = 0x2
	BatchTypeLogData  BatchType = 0x3

	BatchTypeSingleDeletion BatchType = 0x7

	BatchTypeRangeDeletion BatchType = 0xF
)

const (
	headerSize int = 12
	countPos   int = 8
)

func rocksDBBatchDecodeHeader(repr []byte) (count int, orepr pebble.BatchReader, err error) {
	__antithesis_instrumentation__.Notify(633558)
	if len(repr) < headerSize {
		__antithesis_instrumentation__.Notify(633562)
		return 0, nil, errors.Errorf("batch repr too small: %d < %d", len(repr), headerSize)
	} else {
		__antithesis_instrumentation__.Notify(633563)
	}
	__antithesis_instrumentation__.Notify(633559)
	seq := binary.LittleEndian.Uint64(repr[:countPos])
	if seq != 0 {
		__antithesis_instrumentation__.Notify(633564)
		return 0, nil, errors.Errorf("bad sequence: expected 0, but found %d", seq)
	} else {
		__antithesis_instrumentation__.Notify(633565)
	}
	__antithesis_instrumentation__.Notify(633560)
	r, c := pebble.ReadBatch(repr)
	if c > math.MaxInt32 {
		__antithesis_instrumentation__.Notify(633566)
		return 0, nil, errors.Errorf("count %d would overflow max int", c)
	} else {
		__antithesis_instrumentation__.Notify(633567)
	}
	__antithesis_instrumentation__.Notify(633561)
	return int(c), r, nil
}

type RocksDBBatchReader struct {
	batchReader pebble.BatchReader

	err error

	count int

	typ   BatchType
	key   []byte
	value []byte
}

func NewRocksDBBatchReader(repr []byte) (*RocksDBBatchReader, error) {
	__antithesis_instrumentation__.Notify(633568)
	count, batchReader, err := rocksDBBatchDecodeHeader(repr)
	if err != nil {
		__antithesis_instrumentation__.Notify(633570)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(633571)
	}
	__antithesis_instrumentation__.Notify(633569)
	return &RocksDBBatchReader{batchReader: batchReader, count: count}, nil
}

func (r *RocksDBBatchReader) Count() int {
	__antithesis_instrumentation__.Notify(633572)
	return r.count
}

func (r *RocksDBBatchReader) Error() error {
	__antithesis_instrumentation__.Notify(633573)
	return r.err
}

func (r *RocksDBBatchReader) BatchType() BatchType {
	__antithesis_instrumentation__.Notify(633574)
	return r.typ
}

func (r *RocksDBBatchReader) Key() []byte {
	__antithesis_instrumentation__.Notify(633575)
	return r.key
}

func (r *RocksDBBatchReader) MVCCKey() (MVCCKey, error) {
	__antithesis_instrumentation__.Notify(633576)
	return DecodeMVCCKey(r.Key())
}

func (r *RocksDBBatchReader) EngineKey() (EngineKey, error) {
	__antithesis_instrumentation__.Notify(633577)
	key, ok := DecodeEngineKey(r.Key())
	if !ok {
		__antithesis_instrumentation__.Notify(633579)
		return key, errors.Errorf("invalid encoded engine key: %x", r.Key())
	} else {
		__antithesis_instrumentation__.Notify(633580)
	}
	__antithesis_instrumentation__.Notify(633578)
	return key, nil
}

func (r *RocksDBBatchReader) Value() []byte {
	__antithesis_instrumentation__.Notify(633581)
	if r.typ == BatchTypeDeletion || func() bool {
		__antithesis_instrumentation__.Notify(633583)
		return r.typ == BatchTypeSingleDeletion == true
	}() == true {
		__antithesis_instrumentation__.Notify(633584)
		panic("cannot call Value on a deletion entry")
	} else {
		__antithesis_instrumentation__.Notify(633585)
	}
	__antithesis_instrumentation__.Notify(633582)
	return r.value
}

func (r *RocksDBBatchReader) MVCCEndKey() (MVCCKey, error) {
	__antithesis_instrumentation__.Notify(633586)
	if r.typ != BatchTypeRangeDeletion {
		__antithesis_instrumentation__.Notify(633588)
		panic("can only ask for EndKey on a range deletion entry")
	} else {
		__antithesis_instrumentation__.Notify(633589)
	}
	__antithesis_instrumentation__.Notify(633587)
	return DecodeMVCCKey(r.Value())
}

func (r *RocksDBBatchReader) EngineEndKey() (EngineKey, error) {
	__antithesis_instrumentation__.Notify(633590)
	if r.typ != BatchTypeRangeDeletion {
		__antithesis_instrumentation__.Notify(633593)
		panic("can only ask for EndKey on a range deletion entry")
	} else {
		__antithesis_instrumentation__.Notify(633594)
	}
	__antithesis_instrumentation__.Notify(633591)
	key, ok := DecodeEngineKey(r.Value())
	if !ok {
		__antithesis_instrumentation__.Notify(633595)
		return key, errors.Errorf("invalid encoded engine key: %x", r.Value())
	} else {
		__antithesis_instrumentation__.Notify(633596)
	}
	__antithesis_instrumentation__.Notify(633592)
	return key, nil
}

func (r *RocksDBBatchReader) Next() bool {
	__antithesis_instrumentation__.Notify(633597)
	kind, ukey, value, ok := r.batchReader.Next()

	r.typ = BatchType(kind)
	r.key = ukey
	r.value = value

	return ok
}

func RocksDBBatchCount(repr []byte) (int, error) {
	__antithesis_instrumentation__.Notify(633598)
	if len(repr) < headerSize {
		__antithesis_instrumentation__.Notify(633600)
		return 0, errors.Errorf("batch repr too small: %d < %d", len(repr), headerSize)
	} else {
		__antithesis_instrumentation__.Notify(633601)
	}
	__antithesis_instrumentation__.Notify(633599)
	return int(binary.LittleEndian.Uint32(repr[countPos:headerSize])), nil
}
