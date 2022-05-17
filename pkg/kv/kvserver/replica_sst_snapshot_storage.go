package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/fs"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"golang.org/x/time/rate"
)

type SSTSnapshotStorage struct {
	engine  storage.Engine
	limiter *rate.Limiter
	dir     string
}

func NewSSTSnapshotStorage(engine storage.Engine, limiter *rate.Limiter) SSTSnapshotStorage {
	__antithesis_instrumentation__.Notify(120716)
	return SSTSnapshotStorage{
		engine:  engine,
		limiter: limiter,
		dir:     filepath.Join(engine.GetAuxiliaryDir(), "sstsnapshot"),
	}
}

func (s *SSTSnapshotStorage) NewScratchSpace(
	rangeID roachpb.RangeID, snapUUID uuid.UUID,
) *SSTSnapshotStorageScratch {
	__antithesis_instrumentation__.Notify(120717)
	snapDir := filepath.Join(s.dir, strconv.Itoa(int(rangeID)), snapUUID.String())
	return &SSTSnapshotStorageScratch{
		storage: s,
		snapDir: snapDir,
	}
}

func (s *SSTSnapshotStorage) Clear() error {
	__antithesis_instrumentation__.Notify(120718)
	return s.engine.RemoveAll(s.dir)
}

type SSTSnapshotStorageScratch struct {
	storage    *SSTSnapshotStorage
	ssts       []string
	snapDir    string
	dirCreated bool
}

func (s *SSTSnapshotStorageScratch) filename(id int) string {
	__antithesis_instrumentation__.Notify(120719)
	return filepath.Join(s.snapDir, fmt.Sprintf("%d.sst", id))
}

func (s *SSTSnapshotStorageScratch) createDir() error {
	__antithesis_instrumentation__.Notify(120720)
	err := s.storage.engine.MkdirAll(s.snapDir)
	s.dirCreated = s.dirCreated || func() bool {
		__antithesis_instrumentation__.Notify(120721)
		return err == nil == true
	}() == true
	return err
}

func (s *SSTSnapshotStorageScratch) NewFile(
	ctx context.Context, bytesPerSync int64,
) (*SSTSnapshotStorageFile, error) {
	__antithesis_instrumentation__.Notify(120722)
	id := len(s.ssts)
	filename := s.filename(id)
	s.ssts = append(s.ssts, filename)
	f := &SSTSnapshotStorageFile{
		scratch:      s,
		filename:     filename,
		ctx:          ctx,
		bytesPerSync: bytesPerSync,
	}
	return f, nil
}

func (s *SSTSnapshotStorageScratch) WriteSST(ctx context.Context, data []byte) error {
	__antithesis_instrumentation__.Notify(120723)
	if len(data) == 0 {
		__antithesis_instrumentation__.Notify(120729)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(120730)
	}
	__antithesis_instrumentation__.Notify(120724)
	f, err := s.NewFile(ctx, 512<<10)
	if err != nil {
		__antithesis_instrumentation__.Notify(120731)
		return err
	} else {
		__antithesis_instrumentation__.Notify(120732)
	}
	__antithesis_instrumentation__.Notify(120725)
	defer func() {
		__antithesis_instrumentation__.Notify(120733)

		_ = f.Close()
	}()
	__antithesis_instrumentation__.Notify(120726)
	if _, err := f.Write(data); err != nil {
		__antithesis_instrumentation__.Notify(120734)
		return err
	} else {
		__antithesis_instrumentation__.Notify(120735)
	}
	__antithesis_instrumentation__.Notify(120727)
	if err := f.Sync(); err != nil {
		__antithesis_instrumentation__.Notify(120736)
		return err
	} else {
		__antithesis_instrumentation__.Notify(120737)
	}
	__antithesis_instrumentation__.Notify(120728)
	return f.Close()
}

func (s *SSTSnapshotStorageScratch) SSTs() []string {
	__antithesis_instrumentation__.Notify(120738)
	return s.ssts
}

func (s *SSTSnapshotStorageScratch) Clear() error {
	__antithesis_instrumentation__.Notify(120739)
	return s.storage.engine.RemoveAll(s.snapDir)
}

type SSTSnapshotStorageFile struct {
	scratch      *SSTSnapshotStorageScratch
	created      bool
	file         fs.File
	filename     string
	ctx          context.Context
	bytesPerSync int64
}

func (f *SSTSnapshotStorageFile) ensureFile() error {
	__antithesis_instrumentation__.Notify(120740)
	if f.created {
		__antithesis_instrumentation__.Notify(120745)
		if f.file == nil {
			__antithesis_instrumentation__.Notify(120747)
			return errors.Errorf("file has already been closed")
		} else {
			__antithesis_instrumentation__.Notify(120748)
		}
		__antithesis_instrumentation__.Notify(120746)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(120749)
	}
	__antithesis_instrumentation__.Notify(120741)
	if !f.scratch.dirCreated {
		__antithesis_instrumentation__.Notify(120750)
		if err := f.scratch.createDir(); err != nil {
			__antithesis_instrumentation__.Notify(120751)
			return err
		} else {
			__antithesis_instrumentation__.Notify(120752)
		}
	} else {
		__antithesis_instrumentation__.Notify(120753)
	}
	__antithesis_instrumentation__.Notify(120742)
	var err error
	if f.bytesPerSync > 0 {
		__antithesis_instrumentation__.Notify(120754)
		f.file, err = f.scratch.storage.engine.CreateWithSync(f.filename, int(f.bytesPerSync))
	} else {
		__antithesis_instrumentation__.Notify(120755)
		f.file, err = f.scratch.storage.engine.Create(f.filename)
	}
	__antithesis_instrumentation__.Notify(120743)
	if err != nil {
		__antithesis_instrumentation__.Notify(120756)
		return err
	} else {
		__antithesis_instrumentation__.Notify(120757)
	}
	__antithesis_instrumentation__.Notify(120744)
	f.created = true
	return nil
}

func (f *SSTSnapshotStorageFile) Write(contents []byte) (int, error) {
	__antithesis_instrumentation__.Notify(120758)
	if len(contents) == 0 {
		__antithesis_instrumentation__.Notify(120762)
		return 0, nil
	} else {
		__antithesis_instrumentation__.Notify(120763)
	}
	__antithesis_instrumentation__.Notify(120759)
	if err := f.ensureFile(); err != nil {
		__antithesis_instrumentation__.Notify(120764)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(120765)
	}
	__antithesis_instrumentation__.Notify(120760)
	if err := limitBulkIOWrite(f.ctx, f.scratch.storage.limiter, len(contents)); err != nil {
		__antithesis_instrumentation__.Notify(120766)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(120767)
	}
	__antithesis_instrumentation__.Notify(120761)
	return f.file.Write(contents)
}

func (f *SSTSnapshotStorageFile) Close() error {
	__antithesis_instrumentation__.Notify(120768)

	if !f.created {
		__antithesis_instrumentation__.Notify(120772)
		return errors.New("file is empty")
	} else {
		__antithesis_instrumentation__.Notify(120773)
	}
	__antithesis_instrumentation__.Notify(120769)
	if f.file == nil {
		__antithesis_instrumentation__.Notify(120774)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(120775)
	}
	__antithesis_instrumentation__.Notify(120770)
	if err := f.file.Close(); err != nil {
		__antithesis_instrumentation__.Notify(120776)
		return err
	} else {
		__antithesis_instrumentation__.Notify(120777)
	}
	__antithesis_instrumentation__.Notify(120771)
	f.file = nil
	return nil
}

func (f *SSTSnapshotStorageFile) Sync() error {
	__antithesis_instrumentation__.Notify(120778)
	return f.file.Sync()
}
