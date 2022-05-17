package dumpstore

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

type DumpStore struct {
	dir                        string
	maxCombinedFileSizeSetting *settings.ByteSizeSetting
	st                         *cluster.Settings
}

type Dumper interface {
	PreFilter(ctx context.Context, files []os.FileInfo, cleanupFn func(fileName string) error) (preserved map[int]bool, err error)

	CheckOwnsFile(ctx context.Context, fi os.FileInfo) bool
}

func NewStore(
	storeDir string, maxCombinedFileSizeSetting *settings.ByteSizeSetting, st *cluster.Settings,
) *DumpStore {
	__antithesis_instrumentation__.Notify(193328)
	return &DumpStore{
		dir:                        storeDir,
		maxCombinedFileSizeSetting: maxCombinedFileSizeSetting,
		st:                         st,
	}
}

func (s *DumpStore) GetFullPath(fileName string) string {
	__antithesis_instrumentation__.Notify(193329)
	return filepath.Join(s.dir, fileName)
}

func (s *DumpStore) GC(ctx context.Context, now time.Time, dumper Dumper) {
	__antithesis_instrumentation__.Notify(193330)

	files, err := ioutil.ReadDir(s.dir)
	if err != nil {
		__antithesis_instrumentation__.Notify(193334)
		log.Warningf(ctx, "%v", err)
		return
	} else {
		__antithesis_instrumentation__.Notify(193335)
	}
	__antithesis_instrumentation__.Notify(193331)

	maxS := s.maxCombinedFileSizeSetting.Get(&s.st.SV)

	cleanupFn := func(fileName string) error {
		__antithesis_instrumentation__.Notify(193336)
		path := filepath.Join(s.dir, fileName)
		return os.Remove(path)
	}
	__antithesis_instrumentation__.Notify(193332)

	preserved, err := dumper.PreFilter(ctx, files, cleanupFn)
	if err != nil {
		__antithesis_instrumentation__.Notify(193337)
		log.Warningf(ctx, "%v", err)
		return
	} else {
		__antithesis_instrumentation__.Notify(193338)
	}
	__antithesis_instrumentation__.Notify(193333)

	removeOldAndTooBigExcept(ctx, dumper, files, now, maxS, preserved, cleanupFn)
}

func removeOldAndTooBigExcept(
	ctx context.Context,
	dumper Dumper,
	files []os.FileInfo,
	now time.Time,
	maxS int64,
	preserved map[int]bool,
	fn func(string) error,
) {
	__antithesis_instrumentation__.Notify(193339)
	actualSize := int64(0)

	for i := len(files) - 1; i >= 0; i-- {
		__antithesis_instrumentation__.Notify(193340)
		fi := files[i]
		if !dumper.CheckOwnsFile(ctx, fi) {
			__antithesis_instrumentation__.Notify(193343)

			continue
		} else {
			__antithesis_instrumentation__.Notify(193344)
		}
		__antithesis_instrumentation__.Notify(193341)

		actualSize += fi.Size()

		if preserved[i] {
			__antithesis_instrumentation__.Notify(193345)
			continue
		} else {
			__antithesis_instrumentation__.Notify(193346)
		}
		__antithesis_instrumentation__.Notify(193342)

		if actualSize > maxS {
			__antithesis_instrumentation__.Notify(193347)

			if err := fn(fi.Name()); err != nil {
				__antithesis_instrumentation__.Notify(193348)
				log.Warningf(ctx, "cannot remove file %s: %v", fi.Name(), err)
			} else {
				__antithesis_instrumentation__.Notify(193349)
			}
		} else {
			__antithesis_instrumentation__.Notify(193350)
		}
	}
}
