package colcontainer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"strconv"

	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/col/colserde"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/storage/fs"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/golang/snappy"
)

const (
	compressionSizeReductionThreshold = 8

	bytesPerSync = 512 << 10
)

type file struct {
	name string

	offsets []int

	curOffsetIdx int
	totalSize    int

	finishedWriting bool
}

type diskQueueWriter struct {
	testingKnobAlwaysCompress bool
	buffer                    bytes.Buffer
	wrapped                   io.Writer
	scratch                   struct {
		blockType     [1]byte
		compressedBuf []byte
	}
}

const (
	snappyUncompressedBlock byte = 0
	snappyCompressedBlock   byte = 1
)

func (w *diskQueueWriter) Write(p []byte) (int, error) {
	__antithesis_instrumentation__.Notify(272616)
	return w.buffer.Write(p)
}

func (w *diskQueueWriter) reset(wrapped io.Writer) {
	__antithesis_instrumentation__.Notify(272617)
	w.wrapped = wrapped
	w.buffer.Reset()
}

func (w *diskQueueWriter) compressAndFlush() (int, error) {
	__antithesis_instrumentation__.Notify(272618)
	b := w.buffer.Bytes()
	compressed := snappy.Encode(w.scratch.compressedBuf, b)
	w.scratch.compressedBuf = compressed[:cap(compressed)]

	blockType := snappyUncompressedBlock

	if w.testingKnobAlwaysCompress || func() bool {
		__antithesis_instrumentation__.Notify(272622)
		return len(compressed) < len(b)-len(b)/compressionSizeReductionThreshold == true
	}() == true {
		__antithesis_instrumentation__.Notify(272623)
		blockType = snappyCompressedBlock
		b = compressed
	} else {
		__antithesis_instrumentation__.Notify(272624)
	}
	__antithesis_instrumentation__.Notify(272619)

	w.scratch.blockType[0] = blockType
	nType, err := w.wrapped.Write(w.scratch.blockType[:])
	if err != nil {
		__antithesis_instrumentation__.Notify(272625)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(272626)
	}
	__antithesis_instrumentation__.Notify(272620)

	nBody, err := w.wrapped.Write(b)
	if err != nil {
		__antithesis_instrumentation__.Notify(272627)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(272628)
	}
	__antithesis_instrumentation__.Notify(272621)
	w.buffer.Reset()
	return nType + nBody, err
}

func (w *diskQueueWriter) numBytesBuffered() int {
	__antithesis_instrumentation__.Notify(272629)
	return w.buffer.Len()
}

type diskQueueState int

const (
	diskQueueStateEnqueueing diskQueueState = iota
	diskQueueStateDequeueing
)

type diskQueue struct {
	dirName string

	typs  []*types.T
	cfg   DiskQueueCfg
	files []file
	seqNo int

	state      diskQueueState
	rewindable bool

	done bool

	serializer *colserde.FileSerializer

	numBufferedBatches int
	writer             *diskQueueWriter

	writeBufferLimit  int
	writeFileIdx      int
	writeFile         fs.File
	deserializerState struct {
		*colserde.FileDeserializer
		curBatch int
	}

	readFileIdx                  int
	readFile                     fs.File
	scratchDecompressedReadBytes []byte

	diskAcc *mon.BoundAccount
}

var _ RewindableQueue = &diskQueue{}

type Queue interface {
	Enqueue(context.Context, coldata.Batch) error

	Dequeue(context.Context, coldata.Batch) (bool, error)

	CloseRead() error

	Close(context.Context) error
}

type RewindableQueue interface {
	Queue

	Rewind() error
}

const (
	defaultBufferSizeBytesIntertwinedCallsCacheMode = 128 << 10

	defaultBufferSizeBytesReuseCacheMode = 64 << 10

	defaultMaxFileSizeBytes = 32 << 20
)

type DiskQueueCacheMode int

const (
	DiskQueueCacheModeReuseCache DiskQueueCacheMode = iota

	DiskQueueCacheModeClearAndReuseCache

	DiskQueueCacheModeIntertwinedCalls
)

type GetPather interface {
	GetPath(context.Context) string
}

type getPatherFunc struct {
	f func(ctx context.Context) string
}

func (f getPatherFunc) GetPath(ctx context.Context) string {
	__antithesis_instrumentation__.Notify(272630)
	return f.f(ctx)
}

func GetPatherFunc(f func(ctx context.Context) string) GetPather {
	__antithesis_instrumentation__.Notify(272631)
	return getPatherFunc{
		f: f,
	}
}

type DiskQueueCfg struct {
	FS fs.FS

	GetPather GetPather

	CacheMode DiskQueueCacheMode

	BufferSizeBytes int

	MaxFileSizeBytes int

	SpilledBytesWritten *metric.Counter
	SpilledBytesRead    *metric.Counter

	TestingKnobs struct {
		AlwaysCompress bool
	}
}

func (cfg *DiskQueueCfg) EnsureDefaults() error {
	__antithesis_instrumentation__.Notify(272632)
	if cfg.FS == nil {
		__antithesis_instrumentation__.Notify(272636)
		return errors.New("FS unset on DiskQueueCfg")
	} else {
		__antithesis_instrumentation__.Notify(272637)
	}
	__antithesis_instrumentation__.Notify(272633)
	if cfg.BufferSizeBytes == 0 {
		__antithesis_instrumentation__.Notify(272638)
		cfg.setDefaultBufferSizeBytesForCacheMode()
	} else {
		__antithesis_instrumentation__.Notify(272639)
	}
	__antithesis_instrumentation__.Notify(272634)
	if cfg.MaxFileSizeBytes == 0 {
		__antithesis_instrumentation__.Notify(272640)
		cfg.MaxFileSizeBytes = defaultMaxFileSizeBytes
	} else {
		__antithesis_instrumentation__.Notify(272641)
	}
	__antithesis_instrumentation__.Notify(272635)
	return nil
}

func (cfg *DiskQueueCfg) SetCacheMode(m DiskQueueCacheMode) {
	__antithesis_instrumentation__.Notify(272642)
	cfg.CacheMode = m
	cfg.setDefaultBufferSizeBytesForCacheMode()
}

func (cfg *DiskQueueCfg) setDefaultBufferSizeBytesForCacheMode() {
	__antithesis_instrumentation__.Notify(272643)
	if cfg.CacheMode == DiskQueueCacheModeIntertwinedCalls {
		__antithesis_instrumentation__.Notify(272644)
		cfg.BufferSizeBytes = defaultBufferSizeBytesIntertwinedCallsCacheMode
	} else {
		__antithesis_instrumentation__.Notify(272645)
		cfg.BufferSizeBytes = defaultBufferSizeBytesReuseCacheMode
	}
}

func NewDiskQueue(
	ctx context.Context, typs []*types.T, cfg DiskQueueCfg, diskAcc *mon.BoundAccount,
) (Queue, error) {
	__antithesis_instrumentation__.Notify(272646)
	return newDiskQueue(ctx, typs, cfg, diskAcc)
}

func NewRewindableDiskQueue(
	ctx context.Context, typs []*types.T, cfg DiskQueueCfg, diskAcc *mon.BoundAccount,
) (RewindableQueue, error) {
	__antithesis_instrumentation__.Notify(272647)
	d, err := newDiskQueue(ctx, typs, cfg, diskAcc)
	if err != nil {
		__antithesis_instrumentation__.Notify(272649)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(272650)
	}
	__antithesis_instrumentation__.Notify(272648)
	d.rewindable = true
	return d, nil
}

func newDiskQueue(
	ctx context.Context, typs []*types.T, cfg DiskQueueCfg, diskAcc *mon.BoundAccount,
) (*diskQueue, error) {
	__antithesis_instrumentation__.Notify(272651)
	if err := cfg.EnsureDefaults(); err != nil {
		__antithesis_instrumentation__.Notify(272655)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(272656)
	}
	__antithesis_instrumentation__.Notify(272652)
	d := &diskQueue{
		dirName:          uuid.FastMakeV4().String(),
		typs:             typs,
		cfg:              cfg,
		files:            make([]file, 0, 4),
		writeBufferLimit: cfg.BufferSizeBytes / 3,
		diskAcc:          diskAcc,
	}

	if d.cfg.CacheMode != DiskQueueCacheModeIntertwinedCalls {
		__antithesis_instrumentation__.Notify(272657)
		d.writeBufferLimit = d.cfg.BufferSizeBytes / 2
	} else {
		__antithesis_instrumentation__.Notify(272658)
	}
	__antithesis_instrumentation__.Notify(272653)
	if err := cfg.FS.MkdirAll(filepath.Join(cfg.GetPather.GetPath(ctx), d.dirName)); err != nil {
		__antithesis_instrumentation__.Notify(272659)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(272660)
	}
	__antithesis_instrumentation__.Notify(272654)

	return d, d.rotateFile(ctx)
}

func (d *diskQueue) CloseRead() error {
	__antithesis_instrumentation__.Notify(272661)
	if d.readFile != nil {
		__antithesis_instrumentation__.Notify(272663)
		if err := d.readFile.Close(); err != nil {
			__antithesis_instrumentation__.Notify(272665)
			return err
		} else {
			__antithesis_instrumentation__.Notify(272666)
		}
		__antithesis_instrumentation__.Notify(272664)
		d.readFile = nil
	} else {
		__antithesis_instrumentation__.Notify(272667)
	}
	__antithesis_instrumentation__.Notify(272662)
	return nil
}

func (d *diskQueue) closeFileDeserializer() error {
	__antithesis_instrumentation__.Notify(272668)
	if d.deserializerState.FileDeserializer != nil {
		__antithesis_instrumentation__.Notify(272670)
		if err := d.deserializerState.Close(); err != nil {
			__antithesis_instrumentation__.Notify(272671)
			return err
		} else {
			__antithesis_instrumentation__.Notify(272672)
		}
	} else {
		__antithesis_instrumentation__.Notify(272673)
	}
	__antithesis_instrumentation__.Notify(272669)
	d.deserializerState.FileDeserializer = nil
	return nil
}

func (d *diskQueue) Close(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(272674)
	defer func() {
		__antithesis_instrumentation__.Notify(272684)

		*d = diskQueue{}
	}()
	__antithesis_instrumentation__.Notify(272675)
	if d.serializer != nil {
		__antithesis_instrumentation__.Notify(272685)
		if err := d.writeFooterAndFlush(ctx); err != nil {
			__antithesis_instrumentation__.Notify(272687)
			return err
		} else {
			__antithesis_instrumentation__.Notify(272688)
		}
		__antithesis_instrumentation__.Notify(272686)
		d.serializer = nil
	} else {
		__antithesis_instrumentation__.Notify(272689)
	}
	__antithesis_instrumentation__.Notify(272676)
	if err := d.closeFileDeserializer(); err != nil {
		__antithesis_instrumentation__.Notify(272690)
		return err
	} else {
		__antithesis_instrumentation__.Notify(272691)
	}
	__antithesis_instrumentation__.Notify(272677)
	if d.writeFile != nil {
		__antithesis_instrumentation__.Notify(272692)
		if err := d.writeFile.Close(); err != nil {
			__antithesis_instrumentation__.Notify(272694)
			return err
		} else {
			__antithesis_instrumentation__.Notify(272695)
		}
		__antithesis_instrumentation__.Notify(272693)
		d.writeFile = nil
	} else {
		__antithesis_instrumentation__.Notify(272696)
	}
	__antithesis_instrumentation__.Notify(272678)

	if err := d.CloseRead(); err != nil {
		__antithesis_instrumentation__.Notify(272697)
		return err
	} else {
		__antithesis_instrumentation__.Notify(272698)
	}
	__antithesis_instrumentation__.Notify(272679)
	if err := d.cfg.FS.RemoveAll(filepath.Join(d.cfg.GetPather.GetPath(ctx), d.dirName)); err != nil {
		__antithesis_instrumentation__.Notify(272699)
		return err
	} else {
		__antithesis_instrumentation__.Notify(272700)
	}
	__antithesis_instrumentation__.Notify(272680)
	totalSize := int64(0)
	leftOverFileIdx := 0
	if !d.rewindable {
		__antithesis_instrumentation__.Notify(272701)
		leftOverFileIdx = d.readFileIdx
	} else {
		__antithesis_instrumentation__.Notify(272702)
	}
	__antithesis_instrumentation__.Notify(272681)
	for _, file := range d.files[leftOverFileIdx : d.writeFileIdx+1] {
		__antithesis_instrumentation__.Notify(272703)
		totalSize += int64(file.totalSize)
	}
	__antithesis_instrumentation__.Notify(272682)
	if totalSize > d.diskAcc.Used() {
		__antithesis_instrumentation__.Notify(272704)
		totalSize = d.diskAcc.Used()
	} else {
		__antithesis_instrumentation__.Notify(272705)
	}
	__antithesis_instrumentation__.Notify(272683)
	d.diskAcc.Shrink(ctx, totalSize)
	return nil
}

func (d *diskQueue) rotateFile(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(272706)
	fName := filepath.Join(d.cfg.GetPather.GetPath(ctx), d.dirName, strconv.Itoa(d.seqNo))
	f, err := d.cfg.FS.CreateWithSync(fName, bytesPerSync)
	if err != nil {
		__antithesis_instrumentation__.Notify(272710)
		return err
	} else {
		__antithesis_instrumentation__.Notify(272711)
	}
	__antithesis_instrumentation__.Notify(272707)
	d.seqNo++

	if d.serializer == nil {
		__antithesis_instrumentation__.Notify(272712)
		writer := &diskQueueWriter{testingKnobAlwaysCompress: d.cfg.TestingKnobs.AlwaysCompress, wrapped: f}
		d.serializer, err = colserde.NewFileSerializer(writer, d.typs)
		if err != nil {
			__antithesis_instrumentation__.Notify(272714)
			return err
		} else {
			__antithesis_instrumentation__.Notify(272715)
		}
		__antithesis_instrumentation__.Notify(272713)
		d.writer = writer
	} else {
		__antithesis_instrumentation__.Notify(272716)
		if err := d.writeFooterAndFlush(ctx); err != nil {
			__antithesis_instrumentation__.Notify(272718)
			return err
		} else {
			__antithesis_instrumentation__.Notify(272719)
		}
		__antithesis_instrumentation__.Notify(272717)
		if err := d.resetWriters(f); err != nil {
			__antithesis_instrumentation__.Notify(272720)
			return err
		} else {
			__antithesis_instrumentation__.Notify(272721)
		}
	}
	__antithesis_instrumentation__.Notify(272708)

	if d.writeFile != nil {
		__antithesis_instrumentation__.Notify(272722)
		d.files[d.writeFileIdx].finishedWriting = true
		if err := d.writeFile.Close(); err != nil {
			__antithesis_instrumentation__.Notify(272723)
			return err
		} else {
			__antithesis_instrumentation__.Notify(272724)
		}
	} else {
		__antithesis_instrumentation__.Notify(272725)
	}
	__antithesis_instrumentation__.Notify(272709)

	d.writeFileIdx = len(d.files)
	d.files = append(d.files, file{name: fName, offsets: make([]int, 1, 16)})
	d.writeFile = f
	return nil
}

func (d *diskQueue) resetWriters(f fs.File) error {
	__antithesis_instrumentation__.Notify(272726)
	d.writer.reset(f)
	return d.serializer.Reset(d.writer)
}

func (d *diskQueue) writeFooterAndFlush(ctx context.Context) (err error) {
	__antithesis_instrumentation__.Notify(272727)
	defer func() {
		__antithesis_instrumentation__.Notify(272733)
		if err != nil {
			__antithesis_instrumentation__.Notify(272734)

			d.serializer = nil
		} else {
			__antithesis_instrumentation__.Notify(272735)
		}
	}()
	__antithesis_instrumentation__.Notify(272728)
	if err := d.serializer.Finish(); err != nil {
		__antithesis_instrumentation__.Notify(272736)
		return err
	} else {
		__antithesis_instrumentation__.Notify(272737)
	}
	__antithesis_instrumentation__.Notify(272729)
	written, err := d.writer.compressAndFlush()
	if err != nil {
		__antithesis_instrumentation__.Notify(272738)
		return err
	} else {
		__antithesis_instrumentation__.Notify(272739)
	}
	__antithesis_instrumentation__.Notify(272730)
	d.numBufferedBatches = 0
	d.files[d.writeFileIdx].totalSize += written
	if d.cfg.SpilledBytesWritten != nil {
		__antithesis_instrumentation__.Notify(272740)
		d.cfg.SpilledBytesWritten.Inc(int64(written))
	} else {
		__antithesis_instrumentation__.Notify(272741)
	}
	__antithesis_instrumentation__.Notify(272731)
	if err := d.diskAcc.Grow(ctx, int64(written)); err != nil {
		__antithesis_instrumentation__.Notify(272742)
		return err
	} else {
		__antithesis_instrumentation__.Notify(272743)
	}
	__antithesis_instrumentation__.Notify(272732)

	d.files[d.writeFileIdx].offsets = append(d.files[d.writeFileIdx].offsets, d.files[d.writeFileIdx].totalSize)
	return nil
}

func (d *diskQueue) Enqueue(ctx context.Context, b coldata.Batch) error {
	__antithesis_instrumentation__.Notify(272744)
	if d.state == diskQueueStateDequeueing {
		__antithesis_instrumentation__.Notify(272749)
		if d.cfg.CacheMode != DiskQueueCacheModeIntertwinedCalls {
			__antithesis_instrumentation__.Notify(272751)
			return errors.Errorf(
				"attempted to Enqueue to DiskQueue after Dequeueing "+
					"in mode that disallows it: %d", d.cfg.CacheMode,
			)
		} else {
			__antithesis_instrumentation__.Notify(272752)
		}
		__antithesis_instrumentation__.Notify(272750)
		if d.rewindable {
			__antithesis_instrumentation__.Notify(272753)
			return errors.Errorf("attempted to Enqueue to RewindableDiskQueue after Dequeue has been called")
		} else {
			__antithesis_instrumentation__.Notify(272754)
		}
	} else {
		__antithesis_instrumentation__.Notify(272755)
	}
	__antithesis_instrumentation__.Notify(272745)
	d.state = diskQueueStateEnqueueing
	if b.Length() == 0 {
		__antithesis_instrumentation__.Notify(272756)
		if d.done {
			__antithesis_instrumentation__.Notify(272761)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(272762)
		}
		__antithesis_instrumentation__.Notify(272757)
		if err := d.writeFooterAndFlush(ctx); err != nil {
			__antithesis_instrumentation__.Notify(272763)
			return err
		} else {
			__antithesis_instrumentation__.Notify(272764)
		}
		__antithesis_instrumentation__.Notify(272758)
		if err := d.writeFile.Close(); err != nil {
			__antithesis_instrumentation__.Notify(272765)
			return err
		} else {
			__antithesis_instrumentation__.Notify(272766)
		}
		__antithesis_instrumentation__.Notify(272759)
		d.files[d.writeFileIdx].finishedWriting = true
		d.writeFile = nil

		d.serializer = nil

		d.done = true
		if d.cfg.CacheMode == DiskQueueCacheModeClearAndReuseCache {
			__antithesis_instrumentation__.Notify(272767)

			d.scratchDecompressedReadBytes = nil

			d.writer.buffer = bytes.Buffer{}
			d.writer.scratch.compressedBuf = nil
		} else {
			__antithesis_instrumentation__.Notify(272768)
		}
		__antithesis_instrumentation__.Notify(272760)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(272769)
	}
	__antithesis_instrumentation__.Notify(272746)
	if err := d.serializer.AppendBatch(b); err != nil {
		__antithesis_instrumentation__.Notify(272770)
		return err
	} else {
		__antithesis_instrumentation__.Notify(272771)
	}
	__antithesis_instrumentation__.Notify(272747)
	d.numBufferedBatches++

	bufferSizeLimitReached := d.writer.numBytesBuffered() > d.writeBufferLimit
	fileSizeLimitReached := d.files[d.writeFileIdx].totalSize+d.writer.numBytesBuffered() > d.cfg.MaxFileSizeBytes
	if bufferSizeLimitReached || func() bool {
		__antithesis_instrumentation__.Notify(272772)
		return fileSizeLimitReached == true
	}() == true {
		__antithesis_instrumentation__.Notify(272773)
		if fileSizeLimitReached {
			__antithesis_instrumentation__.Notify(272776)

			return d.rotateFile(ctx)
		} else {
			__antithesis_instrumentation__.Notify(272777)
		}
		__antithesis_instrumentation__.Notify(272774)
		if err := d.writeFooterAndFlush(ctx); err != nil {
			__antithesis_instrumentation__.Notify(272778)
			return err
		} else {
			__antithesis_instrumentation__.Notify(272779)
		}
		__antithesis_instrumentation__.Notify(272775)
		return d.resetWriters(d.writeFile)
	} else {
		__antithesis_instrumentation__.Notify(272780)
	}
	__antithesis_instrumentation__.Notify(272748)
	return nil
}

func (d *diskQueue) maybeInitDeserializer(ctx context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(272781)
	if d.deserializerState.FileDeserializer != nil {
		__antithesis_instrumentation__.Notify(272793)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(272794)
	}
	__antithesis_instrumentation__.Notify(272782)
	if d.readFileIdx >= len(d.files) {
		__antithesis_instrumentation__.Notify(272795)

		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(272796)
	}
	__antithesis_instrumentation__.Notify(272783)
	fileToRead := d.files[d.readFileIdx]
	if fileToRead.curOffsetIdx == len(fileToRead.offsets)-1 {
		__antithesis_instrumentation__.Notify(272797)

		if fileToRead.finishedWriting {
			__antithesis_instrumentation__.Notify(272799)

			if err := d.CloseRead(); err != nil {
				__antithesis_instrumentation__.Notify(272802)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(272803)
			}
			__antithesis_instrumentation__.Notify(272800)
			if !d.rewindable {
				__antithesis_instrumentation__.Notify(272804)

				if err := d.cfg.FS.Remove(d.files[d.readFileIdx].name); err != nil {
					__antithesis_instrumentation__.Notify(272807)
					return false, err
				} else {
					__antithesis_instrumentation__.Notify(272808)
				}
				__antithesis_instrumentation__.Notify(272805)
				fileSize := int64(d.files[d.readFileIdx].totalSize)
				if fileSize > d.diskAcc.Used() {
					__antithesis_instrumentation__.Notify(272809)
					fileSize = d.diskAcc.Used()
				} else {
					__antithesis_instrumentation__.Notify(272810)
				}
				__antithesis_instrumentation__.Notify(272806)
				d.diskAcc.Shrink(ctx, fileSize)
			} else {
				__antithesis_instrumentation__.Notify(272811)
			}
			__antithesis_instrumentation__.Notify(272801)
			d.readFile = nil

			d.readFileIdx++
			return d.maybeInitDeserializer(ctx)
		} else {
			__antithesis_instrumentation__.Notify(272812)
		}
		__antithesis_instrumentation__.Notify(272798)

		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(272813)
	}
	__antithesis_instrumentation__.Notify(272784)
	if d.readFile == nil {
		__antithesis_instrumentation__.Notify(272814)

		f, err := d.cfg.FS.Open(fileToRead.name)
		if err != nil {
			__antithesis_instrumentation__.Notify(272816)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(272817)
		}
		__antithesis_instrumentation__.Notify(272815)
		d.readFile = f
	} else {
		__antithesis_instrumentation__.Notify(272818)
	}
	__antithesis_instrumentation__.Notify(272785)
	readRegionStart := fileToRead.offsets[fileToRead.curOffsetIdx]
	readRegionLength := fileToRead.offsets[fileToRead.curOffsetIdx+1] - readRegionStart
	if cap(d.writer.scratch.compressedBuf) < readRegionLength {
		__antithesis_instrumentation__.Notify(272819)

		d.writer.scratch.compressedBuf = make([]byte, readRegionLength)
	} else {
		__antithesis_instrumentation__.Notify(272820)
	}
	__antithesis_instrumentation__.Notify(272786)

	d.writer.scratch.compressedBuf = d.writer.scratch.compressedBuf[0:readRegionLength]

	n, err := d.readFile.ReadAt(d.writer.scratch.compressedBuf, int64(readRegionStart))
	if err != nil && func() bool {
		__antithesis_instrumentation__.Notify(272821)
		return err != io.EOF == true
	}() == true {
		__antithesis_instrumentation__.Notify(272822)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(272823)
	}
	__antithesis_instrumentation__.Notify(272787)
	if d.cfg.SpilledBytesRead != nil {
		__antithesis_instrumentation__.Notify(272824)
		d.cfg.SpilledBytesRead.Inc(int64(n))
	} else {
		__antithesis_instrumentation__.Notify(272825)
	}
	__antithesis_instrumentation__.Notify(272788)
	if n != len(d.writer.scratch.compressedBuf) {
		__antithesis_instrumentation__.Notify(272826)
		return false, errors.Errorf("expected to read %d bytes but read %d", len(d.writer.scratch.compressedBuf), n)
	} else {
		__antithesis_instrumentation__.Notify(272827)
	}
	__antithesis_instrumentation__.Notify(272789)

	blockType := d.writer.scratch.compressedBuf[0]
	compressedBytes := d.writer.scratch.compressedBuf[1:]
	var decompressedBytes []byte
	if blockType == snappyCompressedBlock {
		__antithesis_instrumentation__.Notify(272828)
		decompressedBytes, err = snappy.Decode(d.scratchDecompressedReadBytes, compressedBytes)
		if err != nil {
			__antithesis_instrumentation__.Notify(272830)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(272831)
		}
		__antithesis_instrumentation__.Notify(272829)
		d.scratchDecompressedReadBytes = decompressedBytes[:cap(decompressedBytes)]
	} else {
		__antithesis_instrumentation__.Notify(272832)

		if cap(d.scratchDecompressedReadBytes) < len(compressedBytes) {
			__antithesis_instrumentation__.Notify(272834)
			d.scratchDecompressedReadBytes = make([]byte, len(compressedBytes))
		} else {
			__antithesis_instrumentation__.Notify(272835)
		}
		__antithesis_instrumentation__.Notify(272833)

		d.scratchDecompressedReadBytes = d.scratchDecompressedReadBytes[:len(compressedBytes)]
		copy(d.scratchDecompressedReadBytes, compressedBytes)
		decompressedBytes = d.scratchDecompressedReadBytes
	}
	__antithesis_instrumentation__.Notify(272790)

	deserializer, err := colserde.NewFileDeserializerFromBytes(d.typs, decompressedBytes)
	if err != nil {
		__antithesis_instrumentation__.Notify(272836)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(272837)
	}
	__antithesis_instrumentation__.Notify(272791)
	d.deserializerState.FileDeserializer = deserializer
	d.deserializerState.curBatch = 0
	if d.deserializerState.NumBatches() == 0 {
		__antithesis_instrumentation__.Notify(272838)

		if err := d.closeFileDeserializer(); err != nil {
			__antithesis_instrumentation__.Notify(272840)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(272841)
		}
		__antithesis_instrumentation__.Notify(272839)
		d.files[d.readFileIdx].curOffsetIdx++
		return d.maybeInitDeserializer(ctx)
	} else {
		__antithesis_instrumentation__.Notify(272842)
	}
	__antithesis_instrumentation__.Notify(272792)
	return true, nil
}

func (d *diskQueue) Dequeue(ctx context.Context, b coldata.Batch) (bool, error) {
	__antithesis_instrumentation__.Notify(272843)
	if d.serializer != nil && func() bool {
		__antithesis_instrumentation__.Notify(272848)
		return d.numBufferedBatches > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(272849)
		if err := d.writeFooterAndFlush(ctx); err != nil {
			__antithesis_instrumentation__.Notify(272851)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(272852)
		}
		__antithesis_instrumentation__.Notify(272850)
		if err := d.resetWriters(d.writeFile); err != nil {
			__antithesis_instrumentation__.Notify(272853)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(272854)
		}
	} else {
		__antithesis_instrumentation__.Notify(272855)
	}
	__antithesis_instrumentation__.Notify(272844)
	if d.state == diskQueueStateEnqueueing && func() bool {
		__antithesis_instrumentation__.Notify(272856)
		return d.cfg.CacheMode != DiskQueueCacheModeIntertwinedCalls == true
	}() == true {
		__antithesis_instrumentation__.Notify(272857)

		d.writer.buffer.Reset()
		d.scratchDecompressedReadBytes = d.writer.buffer.Bytes()
	} else {
		__antithesis_instrumentation__.Notify(272858)
	}
	__antithesis_instrumentation__.Notify(272845)
	d.state = diskQueueStateDequeueing

	if d.deserializerState.FileDeserializer != nil && func() bool {
		__antithesis_instrumentation__.Notify(272859)
		return d.deserializerState.curBatch >= d.deserializerState.NumBatches() == true
	}() == true {
		__antithesis_instrumentation__.Notify(272860)

		if err := d.closeFileDeserializer(); err != nil {
			__antithesis_instrumentation__.Notify(272862)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(272863)
		}
		__antithesis_instrumentation__.Notify(272861)
		d.files[d.readFileIdx].curOffsetIdx++
	} else {
		__antithesis_instrumentation__.Notify(272864)
	}
	__antithesis_instrumentation__.Notify(272846)

	if dataToRead, err := d.maybeInitDeserializer(ctx); err != nil {
		__antithesis_instrumentation__.Notify(272865)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(272866)
		if !dataToRead {
			__antithesis_instrumentation__.Notify(272867)

			if !d.done {
				__antithesis_instrumentation__.Notify(272869)

				return false, nil
			} else {
				__antithesis_instrumentation__.Notify(272870)
			}
			__antithesis_instrumentation__.Notify(272868)

			b.SetLength(0)
		} else {
			__antithesis_instrumentation__.Notify(272871)
			if err := d.deserializerState.GetBatch(d.deserializerState.curBatch, b); err != nil {
				__antithesis_instrumentation__.Notify(272873)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(272874)
			}
			__antithesis_instrumentation__.Notify(272872)
			d.deserializerState.curBatch++
		}
	}
	__antithesis_instrumentation__.Notify(272847)

	return true, nil
}

func (d *diskQueue) Rewind() error {
	__antithesis_instrumentation__.Notify(272875)
	if err := d.closeFileDeserializer(); err != nil {
		__antithesis_instrumentation__.Notify(272879)
		return err
	} else {
		__antithesis_instrumentation__.Notify(272880)
	}
	__antithesis_instrumentation__.Notify(272876)
	if err := d.CloseRead(); err != nil {
		__antithesis_instrumentation__.Notify(272881)
		return err
	} else {
		__antithesis_instrumentation__.Notify(272882)
	}
	__antithesis_instrumentation__.Notify(272877)
	d.deserializerState.curBatch = 0
	d.readFile = nil
	d.readFileIdx = 0
	for i := range d.files {
		__antithesis_instrumentation__.Notify(272883)
		d.files[i].curOffsetIdx = 0
	}
	__antithesis_instrumentation__.Notify(272878)
	return nil
}
