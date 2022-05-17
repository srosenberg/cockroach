package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/oserror"
	"golang.org/x/time/rate"
)

var _ SideloadStorage = &diskSideloadStorage{}

type diskSideloadStorage struct {
	st         *cluster.Settings
	limiter    *rate.Limiter
	dir        string
	dirCreated bool
	eng        storage.Engine
}

func deprecatedSideloadedPath(
	baseDir string, rangeID roachpb.RangeID, replicaID roachpb.ReplicaID,
) string {
	__antithesis_instrumentation__.Notify(120593)
	return filepath.Join(
		baseDir,
		"sideloading",
		fmt.Sprintf("%d", rangeID%1000),
		fmt.Sprintf("%d.%d", rangeID, replicaID),
	)
}

func sideloadedPath(baseDir string, rangeID roachpb.RangeID) string {
	__antithesis_instrumentation__.Notify(120594)

	return filepath.Join(
		baseDir,
		"sideloading",
		fmt.Sprintf("r%dXXXX", rangeID/10000),
		fmt.Sprintf("r%d", rangeID),
	)
}

func exists(eng storage.Engine, path string) (bool, error) {
	__antithesis_instrumentation__.Notify(120595)
	_, err := eng.Stat(path)
	if err == nil {
		__antithesis_instrumentation__.Notify(120598)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(120599)
	}
	__antithesis_instrumentation__.Notify(120596)
	if oserror.IsNotExist(err) {
		__antithesis_instrumentation__.Notify(120600)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(120601)
	}
	__antithesis_instrumentation__.Notify(120597)
	return false, err
}

func newDiskSideloadStorage(
	st *cluster.Settings,
	rangeID roachpb.RangeID,
	replicaID roachpb.ReplicaID,
	baseDir string,
	limiter *rate.Limiter,
	eng storage.Engine,
) (*diskSideloadStorage, error) {
	__antithesis_instrumentation__.Notify(120602)
	path := deprecatedSideloadedPath(baseDir, rangeID, replicaID)
	newPath := sideloadedPath(baseDir, rangeID)

	exists, err := exists(eng, path)
	if err != nil {
		__antithesis_instrumentation__.Notify(120605)
		return nil, errors.Wrap(err, "checking pre-migration sideloaded directory")
	} else {
		__antithesis_instrumentation__.Notify(120606)
	}
	__antithesis_instrumentation__.Notify(120603)
	if exists {
		__antithesis_instrumentation__.Notify(120607)
		if err := eng.MkdirAll(filepath.Dir(newPath)); err != nil {
			__antithesis_instrumentation__.Notify(120609)
			return nil, errors.Wrap(err, "creating migrated sideloaded directory")
		} else {
			__antithesis_instrumentation__.Notify(120610)
		}
		__antithesis_instrumentation__.Notify(120608)
		if err := eng.Rename(path, newPath); err != nil {
			__antithesis_instrumentation__.Notify(120611)
			return nil, errors.Wrap(err, "while migrating sideloaded directory")
		} else {
			__antithesis_instrumentation__.Notify(120612)
		}
	} else {
		__antithesis_instrumentation__.Notify(120613)
	}
	__antithesis_instrumentation__.Notify(120604)
	path = newPath

	ss := &diskSideloadStorage{
		dir:     path,
		eng:     eng,
		st:      st,
		limiter: limiter,
	}
	return ss, nil
}

func (ss *diskSideloadStorage) createDir() error {
	__antithesis_instrumentation__.Notify(120614)
	err := ss.eng.MkdirAll(ss.dir)
	ss.dirCreated = ss.dirCreated || func() bool {
		__antithesis_instrumentation__.Notify(120615)
		return err == nil == true
	}() == true
	return err
}

func (ss *diskSideloadStorage) Dir() string {
	__antithesis_instrumentation__.Notify(120616)
	return ss.dir
}

func (ss *diskSideloadStorage) Put(ctx context.Context, index, term uint64, contents []byte) error {
	__antithesis_instrumentation__.Notify(120617)
	filename := ss.filename(ctx, index, term)

	for {
		__antithesis_instrumentation__.Notify(120618)

		if err := writeFileSyncing(ctx, filename, contents, ss.eng, 0644, ss.st, ss.limiter); err == nil {
			__antithesis_instrumentation__.Notify(120621)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(120622)
			if !oserror.IsNotExist(err) {
				__antithesis_instrumentation__.Notify(120623)
				return err
			} else {
				__antithesis_instrumentation__.Notify(120624)
			}
		}
		__antithesis_instrumentation__.Notify(120619)

		if err := ss.createDir(); err != nil {
			__antithesis_instrumentation__.Notify(120625)
			return err
		} else {
			__antithesis_instrumentation__.Notify(120626)
		}
		__antithesis_instrumentation__.Notify(120620)
		continue
	}
}

func (ss *diskSideloadStorage) Get(ctx context.Context, index, term uint64) ([]byte, error) {
	__antithesis_instrumentation__.Notify(120627)
	filename := ss.filename(ctx, index, term)
	b, err := ss.eng.ReadFile(filename)
	if oserror.IsNotExist(err) {
		__antithesis_instrumentation__.Notify(120629)
		return nil, errSideloadedFileNotFound
	} else {
		__antithesis_instrumentation__.Notify(120630)
	}
	__antithesis_instrumentation__.Notify(120628)
	return b, err
}

func (ss *diskSideloadStorage) Filename(ctx context.Context, index, term uint64) (string, error) {
	__antithesis_instrumentation__.Notify(120631)
	return ss.filename(ctx, index, term), nil
}

func (ss *diskSideloadStorage) filename(ctx context.Context, index, term uint64) string {
	__antithesis_instrumentation__.Notify(120632)
	return filepath.Join(ss.dir, fmt.Sprintf("i%d.t%d", index, term))
}

func (ss *diskSideloadStorage) Purge(ctx context.Context, index, term uint64) (int64, error) {
	__antithesis_instrumentation__.Notify(120633)
	return ss.purgeFile(ctx, ss.filename(ctx, index, term))
}

func (ss *diskSideloadStorage) fileSize(filename string) (int64, error) {
	__antithesis_instrumentation__.Notify(120634)
	info, err := ss.eng.Stat(filename)
	if err != nil {
		__antithesis_instrumentation__.Notify(120636)
		if oserror.IsNotExist(err) {
			__antithesis_instrumentation__.Notify(120638)
			return 0, errSideloadedFileNotFound
		} else {
			__antithesis_instrumentation__.Notify(120639)
		}
		__antithesis_instrumentation__.Notify(120637)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(120640)
	}
	__antithesis_instrumentation__.Notify(120635)
	return info.Size(), nil
}

func (ss *diskSideloadStorage) purgeFile(ctx context.Context, filename string) (int64, error) {
	__antithesis_instrumentation__.Notify(120641)
	size, err := ss.fileSize(filename)
	if err != nil {
		__antithesis_instrumentation__.Notify(120644)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(120645)
	}
	__antithesis_instrumentation__.Notify(120642)
	if err := ss.eng.Remove(filename); err != nil {
		__antithesis_instrumentation__.Notify(120646)
		if oserror.IsNotExist(err) {
			__antithesis_instrumentation__.Notify(120648)
			return 0, errSideloadedFileNotFound
		} else {
			__antithesis_instrumentation__.Notify(120649)
		}
		__antithesis_instrumentation__.Notify(120647)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(120650)
	}
	__antithesis_instrumentation__.Notify(120643)
	return size, nil
}

func (ss *diskSideloadStorage) Clear(_ context.Context) error {
	__antithesis_instrumentation__.Notify(120651)
	err := ss.eng.RemoveAll(ss.dir)
	ss.dirCreated = ss.dirCreated && func() bool {
		__antithesis_instrumentation__.Notify(120652)
		return err != nil == true
	}() == true
	return err
}

func (ss *diskSideloadStorage) TruncateTo(
	ctx context.Context, firstIndex uint64,
) (bytesFreed, bytesRetained int64, _ error) {
	__antithesis_instrumentation__.Notify(120653)
	return ss.possiblyTruncateTo(ctx, 0, firstIndex, true)
}

func (ss *diskSideloadStorage) possiblyTruncateTo(
	ctx context.Context, from uint64, to uint64, doTruncate bool,
) (bytesFreed, bytesRetained int64, _ error) {
	__antithesis_instrumentation__.Notify(120654)
	deletedAll := true
	if err := ss.forEach(ctx, func(index uint64, filename string) error {
		__antithesis_instrumentation__.Notify(120657)
		if index >= to {
			__antithesis_instrumentation__.Notify(120662)
			size, err := ss.fileSize(filename)
			if err != nil {
				__antithesis_instrumentation__.Notify(120664)
				return err
			} else {
				__antithesis_instrumentation__.Notify(120665)
			}
			__antithesis_instrumentation__.Notify(120663)
			bytesRetained += size
			deletedAll = false
			return nil
		} else {
			__antithesis_instrumentation__.Notify(120666)
		}
		__antithesis_instrumentation__.Notify(120658)
		if index < from {
			__antithesis_instrumentation__.Notify(120667)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(120668)
		}
		__antithesis_instrumentation__.Notify(120659)

		var fileSize int64
		var err error
		if doTruncate {
			__antithesis_instrumentation__.Notify(120669)
			fileSize, err = ss.purgeFile(ctx, filename)
		} else {
			__antithesis_instrumentation__.Notify(120670)
			fileSize, err = ss.fileSize(filename)
		}
		__antithesis_instrumentation__.Notify(120660)
		if err != nil {
			__antithesis_instrumentation__.Notify(120671)
			return err
		} else {
			__antithesis_instrumentation__.Notify(120672)
		}
		__antithesis_instrumentation__.Notify(120661)
		bytesFreed += fileSize
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(120673)
		return 0, 0, err
	} else {
		__antithesis_instrumentation__.Notify(120674)
	}
	__antithesis_instrumentation__.Notify(120655)

	if deletedAll && func() bool {
		__antithesis_instrumentation__.Notify(120675)
		return doTruncate == true
	}() == true {
		__antithesis_instrumentation__.Notify(120676)

		err := ss.eng.Remove(ss.dir)
		if err != nil && func() bool {
			__antithesis_instrumentation__.Notify(120677)
			return !oserror.IsNotExist(err) == true
		}() == true {
			__antithesis_instrumentation__.Notify(120678)
			log.Infof(ctx, "unable to remove sideloaded dir %s: %v", ss.dir, err)
			err = nil
		} else {
			__antithesis_instrumentation__.Notify(120679)
		}
	} else {
		__antithesis_instrumentation__.Notify(120680)
	}
	__antithesis_instrumentation__.Notify(120656)
	return bytesFreed, bytesRetained, nil
}

func (ss *diskSideloadStorage) BytesIfTruncatedFromTo(
	ctx context.Context, from uint64, to uint64,
) (freed, retained int64, _ error) {
	__antithesis_instrumentation__.Notify(120681)
	return ss.possiblyTruncateTo(ctx, from, to, false)
}

func (ss *diskSideloadStorage) forEach(
	ctx context.Context, visit func(index uint64, filename string) error,
) error {
	__antithesis_instrumentation__.Notify(120682)
	matches, err := ss.eng.List(ss.dir)
	if oserror.IsNotExist(err) {
		__antithesis_instrumentation__.Notify(120686)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(120687)
	}
	__antithesis_instrumentation__.Notify(120683)
	if err != nil {
		__antithesis_instrumentation__.Notify(120688)
		return err
	} else {
		__antithesis_instrumentation__.Notify(120689)
	}
	__antithesis_instrumentation__.Notify(120684)
	for _, match := range matches {
		__antithesis_instrumentation__.Notify(120690)

		match = filepath.Join(ss.dir, match)
		base := filepath.Base(match)

		if len(base) < 1 || func() bool {
			__antithesis_instrumentation__.Notify(120693)
			return base[0] != 'i' == true
		}() == true {
			__antithesis_instrumentation__.Notify(120694)
			continue
		} else {
			__antithesis_instrumentation__.Notify(120695)
		}
		__antithesis_instrumentation__.Notify(120691)
		base = base[1:]
		upToDot := strings.SplitN(base, ".", 2)
		logIdx, err := strconv.ParseUint(upToDot[0], 10, 64)
		if err != nil {
			__antithesis_instrumentation__.Notify(120696)
			log.Infof(ctx, "unexpected file %s in sideloaded directory %s", match, ss.dir)
			continue
		} else {
			__antithesis_instrumentation__.Notify(120697)
		}
		__antithesis_instrumentation__.Notify(120692)
		if err := visit(logIdx, match); err != nil {
			__antithesis_instrumentation__.Notify(120698)
			return errors.Wrapf(err, "matching pattern %q on dir %s", match, ss.dir)
		} else {
			__antithesis_instrumentation__.Notify(120699)
		}
	}
	__antithesis_instrumentation__.Notify(120685)
	return nil
}

func (ss *diskSideloadStorage) String() string {
	__antithesis_instrumentation__.Notify(120700)
	var buf strings.Builder
	var count int
	if err := ss.forEach(context.Background(), func(_ uint64, filename string) error {
		__antithesis_instrumentation__.Notify(120702)
		count++
		_, _ = fmt.Fprintln(&buf, filename)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(120703)
		return err.Error()
	} else {
		__antithesis_instrumentation__.Notify(120704)
	}
	__antithesis_instrumentation__.Notify(120701)
	fmt.Fprintf(&buf, "(%d files)\n", count)
	return buf.String()
}
