package storage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"io"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/pebble/sstable"
)

type SSTWriter struct {
	fw *sstable.Writer
	f  io.Writer

	DataSize int64
	scratch  []byte
}

var _ Writer = &SSTWriter{}

type writeCloseSyncer interface {
	io.WriteCloser
	Sync() error
}

type noopSyncCloser struct {
	io.Writer
}

func (noopSyncCloser) Sync() error {
	__antithesis_instrumentation__.Notify(643862)
	return nil
}

func (noopSyncCloser) Close() error {
	__antithesis_instrumentation__.Notify(643863)
	return nil
}

func MakeIngestionWriterOptions(ctx context.Context, cs *cluster.Settings) sstable.WriterOptions {
	__antithesis_instrumentation__.Notify(643864)

	format := sstable.TableFormatRocksDBv2

	switch {
	case cs.Version.IsActive(ctx, clusterversion.EnablePebbleFormatVersionBlockProperties):
		__antithesis_instrumentation__.Notify(643867)
		format = sstable.TableFormatPebblev1
	default:
		__antithesis_instrumentation__.Notify(643868)
	}
	__antithesis_instrumentation__.Notify(643865)
	opts := DefaultPebbleOptions().MakeWriterOptions(0, format)
	if format < sstable.TableFormatPebblev1 {
		__antithesis_instrumentation__.Notify(643869)

		opts.BlockPropertyCollectors = nil
	} else {
		__antithesis_instrumentation__.Notify(643870)
	}
	__antithesis_instrumentation__.Notify(643866)
	opts.MergerName = "nullptr"
	return opts
}

func MakeBackupSSTWriter(_ context.Context, _ *cluster.Settings, f io.Writer) SSTWriter {
	__antithesis_instrumentation__.Notify(643871)

	opts := DefaultPebbleOptions().MakeWriterOptions(0, sstable.TableFormatRocksDBv2)

	opts.BlockPropertyCollectors = nil

	opts.FilterPolicy = nil

	opts.BlockSize = 128 << 10

	opts.MergerName = "nullptr"
	sst := sstable.NewWriter(noopSyncCloser{f}, opts)
	return SSTWriter{fw: sst, f: f}
}

func MakeIngestionSSTWriter(
	ctx context.Context, cs *cluster.Settings, f writeCloseSyncer,
) SSTWriter {
	__antithesis_instrumentation__.Notify(643872)
	return SSTWriter{
		fw: sstable.NewWriter(f, MakeIngestionWriterOptions(ctx, cs)),
		f:  f,
	}
}

func (fw *SSTWriter) Finish() error {
	__antithesis_instrumentation__.Notify(643873)
	if fw.fw == nil {
		__antithesis_instrumentation__.Notify(643876)
		return errors.New("cannot call Finish on a closed writer")
	} else {
		__antithesis_instrumentation__.Notify(643877)
	}
	__antithesis_instrumentation__.Notify(643874)
	if err := fw.fw.Close(); err != nil {
		__antithesis_instrumentation__.Notify(643878)
		return err
	} else {
		__antithesis_instrumentation__.Notify(643879)
	}
	__antithesis_instrumentation__.Notify(643875)
	fw.fw = nil
	return nil
}

func (fw *SSTWriter) ClearRawRange(start, end roachpb.Key) error {
	__antithesis_instrumentation__.Notify(643880)
	return fw.clearRange(MVCCKey{Key: start}, MVCCKey{Key: end})
}

func (fw *SSTWriter) ClearMVCCRangeAndIntents(start, end roachpb.Key) error {
	__antithesis_instrumentation__.Notify(643881)
	panic("ClearMVCCRangeAndIntents is unsupported")
}

func (fw *SSTWriter) ClearMVCCRange(start, end MVCCKey) error {
	__antithesis_instrumentation__.Notify(643882)
	return fw.clearRange(start, end)
}

func (fw *SSTWriter) clearRange(start, end MVCCKey) error {
	__antithesis_instrumentation__.Notify(643883)
	if fw.fw == nil {
		__antithesis_instrumentation__.Notify(643885)
		return errors.New("cannot call ClearRange on a closed writer")
	} else {
		__antithesis_instrumentation__.Notify(643886)
	}
	__antithesis_instrumentation__.Notify(643884)
	fw.DataSize += int64(len(start.Key)) + int64(len(end.Key))
	fw.scratch = EncodeMVCCKeyToBuf(fw.scratch[:0], start)
	return fw.fw.DeleteRange(fw.scratch, EncodeMVCCKey(end))
}

func (fw *SSTWriter) Put(key MVCCKey, value []byte) error {
	__antithesis_instrumentation__.Notify(643887)
	if fw.fw == nil {
		__antithesis_instrumentation__.Notify(643889)
		return errors.New("cannot call Put on a closed writer")
	} else {
		__antithesis_instrumentation__.Notify(643890)
	}
	__antithesis_instrumentation__.Notify(643888)
	fw.DataSize += int64(len(key.Key)) + int64(len(value))
	fw.scratch = EncodeMVCCKeyToBuf(fw.scratch[:0], key)
	return fw.fw.Set(fw.scratch, value)
}

func (fw *SSTWriter) PutMVCC(key MVCCKey, value []byte) error {
	__antithesis_instrumentation__.Notify(643891)
	if key.Timestamp.IsEmpty() {
		__antithesis_instrumentation__.Notify(643893)
		panic("PutMVCC timestamp is empty")
	} else {
		__antithesis_instrumentation__.Notify(643894)
	}
	__antithesis_instrumentation__.Notify(643892)
	return fw.put(key, value)
}

func (fw *SSTWriter) PutUnversioned(key roachpb.Key, value []byte) error {
	__antithesis_instrumentation__.Notify(643895)
	return fw.put(MVCCKey{Key: key}, value)
}

func (fw *SSTWriter) PutIntent(
	ctx context.Context, key roachpb.Key, value []byte, txnUUID uuid.UUID,
) error {
	__antithesis_instrumentation__.Notify(643896)
	return fw.put(MVCCKey{Key: key}, value)
}

func (fw *SSTWriter) PutEngineKey(key EngineKey, value []byte) error {
	__antithesis_instrumentation__.Notify(643897)
	if fw.fw == nil {
		__antithesis_instrumentation__.Notify(643899)
		return errors.New("cannot call Put on a closed writer")
	} else {
		__antithesis_instrumentation__.Notify(643900)
	}
	__antithesis_instrumentation__.Notify(643898)
	fw.DataSize += int64(len(key.Key)) + int64(len(value))
	fw.scratch = key.EncodeToBuf(fw.scratch[:0])
	return fw.fw.Set(fw.scratch, value)
}

func (fw *SSTWriter) put(key MVCCKey, value []byte) error {
	__antithesis_instrumentation__.Notify(643901)
	if fw.fw == nil {
		__antithesis_instrumentation__.Notify(643903)
		return errors.New("cannot call Put on a closed writer")
	} else {
		__antithesis_instrumentation__.Notify(643904)
	}
	__antithesis_instrumentation__.Notify(643902)
	fw.DataSize += int64(len(key.Key)) + int64(len(value))
	fw.scratch = EncodeMVCCKeyToBuf(fw.scratch[:0], key)
	return fw.fw.Set(fw.scratch, value)
}

func (fw *SSTWriter) ApplyBatchRepr(repr []byte, sync bool) error {
	__antithesis_instrumentation__.Notify(643905)
	panic("unimplemented")
}

func (fw *SSTWriter) ClearMVCC(key MVCCKey) error {
	__antithesis_instrumentation__.Notify(643906)
	if key.Timestamp.IsEmpty() {
		__antithesis_instrumentation__.Notify(643908)
		panic("ClearMVCC timestamp is empty")
	} else {
		__antithesis_instrumentation__.Notify(643909)
	}
	__antithesis_instrumentation__.Notify(643907)
	return fw.clear(key)
}

func (fw *SSTWriter) ClearUnversioned(key roachpb.Key) error {
	__antithesis_instrumentation__.Notify(643910)
	return fw.clear(MVCCKey{Key: key})
}

func (fw *SSTWriter) ClearIntent(
	key roachpb.Key, txnDidNotUpdateMeta bool, txnUUID uuid.UUID,
) error {
	__antithesis_instrumentation__.Notify(643911)
	panic("ClearIntent is unsupported")
}

func (fw *SSTWriter) ClearEngineKey(key EngineKey) error {
	__antithesis_instrumentation__.Notify(643912)
	if fw.fw == nil {
		__antithesis_instrumentation__.Notify(643914)
		return errors.New("cannot call Clear on a closed writer")
	} else {
		__antithesis_instrumentation__.Notify(643915)
	}
	__antithesis_instrumentation__.Notify(643913)
	fw.scratch = key.EncodeToBuf(fw.scratch[:0])
	fw.DataSize += int64(len(key.Key))
	return fw.fw.Delete(fw.scratch)
}

func (fw *SSTWriter) clear(key MVCCKey) error {
	__antithesis_instrumentation__.Notify(643916)
	if fw.fw == nil {
		__antithesis_instrumentation__.Notify(643918)
		return errors.New("cannot call Clear on a closed writer")
	} else {
		__antithesis_instrumentation__.Notify(643919)
	}
	__antithesis_instrumentation__.Notify(643917)
	fw.scratch = EncodeMVCCKeyToBuf(fw.scratch[:0], key)
	fw.DataSize += int64(len(key.Key))
	return fw.fw.Delete(fw.scratch)
}

func (fw *SSTWriter) SingleClearEngineKey(key EngineKey) error {
	__antithesis_instrumentation__.Notify(643920)
	panic("unimplemented")
}

func (fw *SSTWriter) ClearIterRange(iter MVCCIterator, start, end roachpb.Key) error {
	__antithesis_instrumentation__.Notify(643921)
	panic("ClearIterRange is unsupported")
}

func (fw *SSTWriter) Merge(key MVCCKey, value []byte) error {
	__antithesis_instrumentation__.Notify(643922)
	if fw.fw == nil {
		__antithesis_instrumentation__.Notify(643924)
		return errors.New("cannot call Merge on a closed writer")
	} else {
		__antithesis_instrumentation__.Notify(643925)
	}
	__antithesis_instrumentation__.Notify(643923)
	fw.DataSize += int64(len(key.Key)) + int64(len(value))
	fw.scratch = EncodeMVCCKeyToBuf(fw.scratch[:0], key)
	return fw.fw.Merge(fw.scratch, value)
}

func (fw *SSTWriter) LogData(data []byte) error {
	__antithesis_instrumentation__.Notify(643926)

	return nil
}

func (fw *SSTWriter) LogLogicalOp(op MVCCLogicalOpType, details MVCCLogicalOpDetails) {
	__antithesis_instrumentation__.Notify(643927)

}

func (fw *SSTWriter) Close() {
	__antithesis_instrumentation__.Notify(643928)
	if fw.fw == nil {
		__antithesis_instrumentation__.Notify(643930)
		return
	} else {
		__antithesis_instrumentation__.Notify(643931)
	}
	__antithesis_instrumentation__.Notify(643929)

	_ = fw.fw.Close()
	fw.fw = nil
}

type MemFile struct {
	bytes.Buffer
}

func (*MemFile) Close() error {
	__antithesis_instrumentation__.Notify(643932)
	return nil
}

func (*MemFile) Flush() error {
	__antithesis_instrumentation__.Notify(643933)
	return nil
}

func (*MemFile) Sync() error {
	__antithesis_instrumentation__.Notify(643934)
	return nil
}

func (f *MemFile) Data() []byte {
	__antithesis_instrumentation__.Notify(643935)
	return f.Bytes()
}
