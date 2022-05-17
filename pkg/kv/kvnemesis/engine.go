package kvnemesis

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/bufalloc"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/vfs"
)

type Engine struct {
	kvs *pebble.DB
	b   bufalloc.ByteAllocator
}

func MakeEngine() (*Engine, error) {
	__antithesis_instrumentation__.Notify(90200)
	opts := storage.DefaultPebbleOptions()
	opts.FS = vfs.NewMem()
	kvs, err := pebble.Open(`kvnemesis`, opts)
	if err != nil {
		__antithesis_instrumentation__.Notify(90202)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(90203)
	}
	__antithesis_instrumentation__.Notify(90201)
	return &Engine{kvs: kvs}, nil
}

func (e *Engine) Close() {
	__antithesis_instrumentation__.Notify(90204)
	if err := e.kvs.Close(); err != nil {
		__antithesis_instrumentation__.Notify(90205)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(90206)
	}
}

func (e *Engine) Get(key roachpb.Key, ts hlc.Timestamp) roachpb.Value {
	__antithesis_instrumentation__.Notify(90207)
	iter := e.kvs.NewIter(nil)
	defer func() { __antithesis_instrumentation__.Notify(90213); _ = iter.Close() }()
	__antithesis_instrumentation__.Notify(90208)
	iter.SeekGE(storage.EncodeMVCCKey(storage.MVCCKey{Key: key, Timestamp: ts}))
	if !iter.Valid() {
		__antithesis_instrumentation__.Notify(90214)
		return roachpb.Value{}
	} else {
		__antithesis_instrumentation__.Notify(90215)
	}
	__antithesis_instrumentation__.Notify(90209)

	mvccKey, err := storage.DecodeMVCCKey(iter.Key())
	if err != nil {
		__antithesis_instrumentation__.Notify(90216)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(90217)
	}
	__antithesis_instrumentation__.Notify(90210)
	if !mvccKey.Key.Equal(key) {
		__antithesis_instrumentation__.Notify(90218)
		return roachpb.Value{}
	} else {
		__antithesis_instrumentation__.Notify(90219)
	}
	__antithesis_instrumentation__.Notify(90211)
	if len(iter.Value()) == 0 {
		__antithesis_instrumentation__.Notify(90220)
		return roachpb.Value{}
	} else {
		__antithesis_instrumentation__.Notify(90221)
	}
	__antithesis_instrumentation__.Notify(90212)
	var valCopy []byte
	e.b, valCopy = e.b.Copy(iter.Value(), 0)
	return roachpb.Value{RawBytes: valCopy, Timestamp: mvccKey.Timestamp}
}

func (e *Engine) Put(key storage.MVCCKey, value []byte) {
	__antithesis_instrumentation__.Notify(90222)
	if err := e.kvs.Set(storage.EncodeMVCCKey(key), value, nil); err != nil {
		__antithesis_instrumentation__.Notify(90223)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(90224)
	}
}

func (e *Engine) Delete(key storage.MVCCKey) {
	__antithesis_instrumentation__.Notify(90225)
	if err := e.kvs.Set(storage.EncodeMVCCKey(key), nil, nil); err != nil {
		__antithesis_instrumentation__.Notify(90226)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(90227)
	}
}

func (e *Engine) Iterate(fn func(key storage.MVCCKey, value []byte, err error)) {
	__antithesis_instrumentation__.Notify(90228)
	iter := e.kvs.NewIter(nil)
	defer func() { __antithesis_instrumentation__.Notify(90230); _ = iter.Close() }()
	__antithesis_instrumentation__.Notify(90229)
	for iter.First(); iter.Valid(); iter.Next() {
		__antithesis_instrumentation__.Notify(90231)
		if err := iter.Error(); err != nil {
			__antithesis_instrumentation__.Notify(90234)
			fn(storage.MVCCKey{}, nil, err)
			continue
		} else {
			__antithesis_instrumentation__.Notify(90235)
		}
		__antithesis_instrumentation__.Notify(90232)
		var keyCopy, valCopy []byte
		e.b, keyCopy = e.b.Copy(iter.Key(), 0)
		e.b, valCopy = e.b.Copy(iter.Value(), 0)
		key, err := storage.DecodeMVCCKey(keyCopy)
		if err != nil {
			__antithesis_instrumentation__.Notify(90236)
			fn(storage.MVCCKey{}, nil, err)
			continue
		} else {
			__antithesis_instrumentation__.Notify(90237)
		}
		__antithesis_instrumentation__.Notify(90233)
		fn(key, valCopy, nil)
	}
}

func (e *Engine) DebugPrint(indent string) string {
	__antithesis_instrumentation__.Notify(90238)
	var buf strings.Builder
	e.Iterate(func(key storage.MVCCKey, value []byte, err error) {
		__antithesis_instrumentation__.Notify(90240)
		if buf.Len() > 0 {
			__antithesis_instrumentation__.Notify(90242)
			buf.WriteString("\n")
		} else {
			__antithesis_instrumentation__.Notify(90243)
		}
		__antithesis_instrumentation__.Notify(90241)
		if err != nil {
			__antithesis_instrumentation__.Notify(90244)
			fmt.Fprintf(&buf, "(err:%s)", err)
		} else {
			__antithesis_instrumentation__.Notify(90245)
			fmt.Fprintf(&buf, "%s%s %s -> %s",
				indent, key.Key, key.Timestamp, roachpb.Value{RawBytes: value}.PrettyPrint())
		}
	})
	__antithesis_instrumentation__.Notify(90239)
	return buf.String()
}
