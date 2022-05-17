package blobs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/blobs/blobspb"
	"github.com/cockroachdb/cockroach/pkg/util/fileutil"
	"github.com/cockroachdb/cockroach/pkg/util/ioctx"
	"github.com/cockroachdb/errors"
)

type LocalStorage struct {
	externalIODir string
}

func NewLocalStorage(externalIODir string) (*LocalStorage, error) {
	__antithesis_instrumentation__.Notify(3248)

	if externalIODir == "" {
		__antithesis_instrumentation__.Notify(3251)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(3252)
	}
	__antithesis_instrumentation__.Notify(3249)
	absPath, err := filepath.Abs(externalIODir)
	if err != nil {
		__antithesis_instrumentation__.Notify(3253)
		return nil, errors.Wrap(err, "creating LocalStorage object")
	} else {
		__antithesis_instrumentation__.Notify(3254)
	}
	__antithesis_instrumentation__.Notify(3250)
	return &LocalStorage{externalIODir: absPath}, nil
}

func (l *LocalStorage) prependExternalIODir(path string) (string, error) {
	__antithesis_instrumentation__.Notify(3255)
	if l == nil {
		__antithesis_instrumentation__.Notify(3257)
		return "", errors.Errorf("local file access is disabled")
	} else {
		__antithesis_instrumentation__.Notify(3258)
	}
	__antithesis_instrumentation__.Notify(3256)
	localBase := filepath.Join(l.externalIODir, path)
	return localBase, l.ensureContained(localBase, path)
}

func (l *LocalStorage) ensureContained(realPath, inputPath string) error {
	__antithesis_instrumentation__.Notify(3259)
	if !strings.HasPrefix(realPath, l.externalIODir) {
		__antithesis_instrumentation__.Notify(3261)
		return errors.Errorf("local file access to paths outside of external-io-dir is not allowed: %s", inputPath)
	} else {
		__antithesis_instrumentation__.Notify(3262)
	}
	__antithesis_instrumentation__.Notify(3260)
	return nil
}

type localWriter struct {
	f         *os.File
	ctx       context.Context
	tmp, dest string
}

func (l localWriter) Write(p []byte) (int, error) {
	__antithesis_instrumentation__.Notify(3263)
	return l.f.Write(p)
}

func (l localWriter) Close() error {
	__antithesis_instrumentation__.Notify(3264)
	if err := l.ctx.Err(); err != nil {
		__antithesis_instrumentation__.Notify(3267)
		closeErr := l.f.Close()
		rmErr := os.Remove(l.tmp)
		return errors.CombineErrors(err, errors.Wrap(errors.CombineErrors(rmErr, closeErr), "cleaning up"))
	} else {
		__antithesis_instrumentation__.Notify(3268)
	}
	__antithesis_instrumentation__.Notify(3265)

	syncErr := l.f.Sync()
	closeErr := l.f.Close()
	if err := errors.CombineErrors(closeErr, syncErr); err != nil {
		__antithesis_instrumentation__.Notify(3269)
		return err
	} else {
		__antithesis_instrumentation__.Notify(3270)
	}
	__antithesis_instrumentation__.Notify(3266)

	return errors.Wrapf(
		fileutil.Move(l.tmp, l.dest),
		"moving temporary file to final location %q",
		l.dest,
	)
}

func (l *LocalStorage) Writer(ctx context.Context, filename string) (io.WriteCloser, error) {
	__antithesis_instrumentation__.Notify(3271)
	fullPath, err := l.prependExternalIODir(filename)
	if err != nil {
		__antithesis_instrumentation__.Notify(3276)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(3277)
	}
	__antithesis_instrumentation__.Notify(3272)
	if err := ctx.Err(); err != nil {
		__antithesis_instrumentation__.Notify(3278)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(3279)
	}
	__antithesis_instrumentation__.Notify(3273)

	targetDir := filepath.Dir(fullPath)
	if err = os.MkdirAll(targetDir, 0755); err != nil {
		__antithesis_instrumentation__.Notify(3280)
		return nil, errors.Wrapf(err, "creating target local directory %q", targetDir)
	} else {
		__antithesis_instrumentation__.Notify(3281)
	}
	__antithesis_instrumentation__.Notify(3274)

	tmpFile, err := ioutil.TempFile(targetDir, filepath.Base(fullPath)+"*.tmp")
	if err != nil {
		__antithesis_instrumentation__.Notify(3282)
		return nil, errors.Wrap(err, "creating temporary file")
	} else {
		__antithesis_instrumentation__.Notify(3283)
	}
	__antithesis_instrumentation__.Notify(3275)
	return localWriter{tmp: tmpFile.Name(), dest: fullPath, f: tmpFile, ctx: ctx}, nil
}

func (l *LocalStorage) ReadFile(
	filename string, offset int64,
) (res ioctx.ReadCloserCtx, size int64, err error) {
	__antithesis_instrumentation__.Notify(3284)
	fullPath, err := l.prependExternalIODir(filename)
	if err != nil {
		__antithesis_instrumentation__.Notify(3291)
		return nil, 0, err
	} else {
		__antithesis_instrumentation__.Notify(3292)
	}
	__antithesis_instrumentation__.Notify(3285)
	f, err := os.Open(fullPath)
	if err != nil {
		__antithesis_instrumentation__.Notify(3293)
		return nil, 0, err
	} else {
		__antithesis_instrumentation__.Notify(3294)
	}
	__antithesis_instrumentation__.Notify(3286)
	defer func() {
		__antithesis_instrumentation__.Notify(3295)
		if err != nil {
			__antithesis_instrumentation__.Notify(3296)
			_ = f.Close()
		} else {
			__antithesis_instrumentation__.Notify(3297)
		}
	}()
	__antithesis_instrumentation__.Notify(3287)
	fi, err := f.Stat()
	if err != nil {
		__antithesis_instrumentation__.Notify(3298)
		return nil, 0, err
	} else {
		__antithesis_instrumentation__.Notify(3299)
	}
	__antithesis_instrumentation__.Notify(3288)
	if fi.IsDir() {
		__antithesis_instrumentation__.Notify(3300)
		return nil, 0, errors.Errorf("expected a file but %q is a directory", fi.Name())
	} else {
		__antithesis_instrumentation__.Notify(3301)
	}
	__antithesis_instrumentation__.Notify(3289)
	if offset != 0 {
		__antithesis_instrumentation__.Notify(3302)
		if ret, err := f.Seek(offset, 0); err != nil {
			__antithesis_instrumentation__.Notify(3303)
			return nil, 0, err
		} else {
			__antithesis_instrumentation__.Notify(3304)
			if ret != offset {
				__antithesis_instrumentation__.Notify(3305)
				return nil, 0, errors.Errorf("seek to offset %d returned %d", offset, ret)
			} else {
				__antithesis_instrumentation__.Notify(3306)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(3307)
	}
	__antithesis_instrumentation__.Notify(3290)
	return ioctx.ReadCloserAdapter(f), fi.Size(), nil
}

func (l *LocalStorage) List(pattern string) ([]string, error) {
	__antithesis_instrumentation__.Notify(3308)
	if pattern == "" {
		__antithesis_instrumentation__.Notify(3314)
		return nil, errors.New("pattern cannot be empty")
	} else {
		__antithesis_instrumentation__.Notify(3315)
	}
	__antithesis_instrumentation__.Notify(3309)
	fullPath, err := l.prependExternalIODir(pattern)
	if err != nil {
		__antithesis_instrumentation__.Notify(3316)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(3317)
	}
	__antithesis_instrumentation__.Notify(3310)

	if !strings.ContainsAny(pattern, "*?[") {
		__antithesis_instrumentation__.Notify(3318)
		var matches []string
		walkRoot := fullPath
		listingParent := false
		if f, err := os.Stat(fullPath); err != nil || func() bool {
			__antithesis_instrumentation__.Notify(3321)
			return !f.IsDir() == true
		}() == true {
			__antithesis_instrumentation__.Notify(3322)
			listingParent = true
			walkRoot = filepath.Dir(fullPath)
			if err := l.ensureContained(walkRoot, pattern); err != nil {
				__antithesis_instrumentation__.Notify(3323)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(3324)
			}
		} else {
			__antithesis_instrumentation__.Notify(3325)
		}
		__antithesis_instrumentation__.Notify(3319)

		if err := filepath.Walk(walkRoot, func(p string, f os.FileInfo, err error) error {
			__antithesis_instrumentation__.Notify(3326)
			if err != nil {
				__antithesis_instrumentation__.Notify(3330)
				return err
			} else {
				__antithesis_instrumentation__.Notify(3331)
			}
			__antithesis_instrumentation__.Notify(3327)
			if f.IsDir() {
				__antithesis_instrumentation__.Notify(3332)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(3333)
			}
			__antithesis_instrumentation__.Notify(3328)
			if listingParent && func() bool {
				__antithesis_instrumentation__.Notify(3334)
				return !strings.HasPrefix(p, fullPath) == true
			}() == true {
				__antithesis_instrumentation__.Notify(3335)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(3336)
			}
			__antithesis_instrumentation__.Notify(3329)
			matches = append(matches, strings.TrimPrefix(p, l.externalIODir))
			return nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(3337)
			if errors.Is(err, os.ErrNotExist) {
				__antithesis_instrumentation__.Notify(3339)
				return nil, nil
			} else {
				__antithesis_instrumentation__.Notify(3340)
			}
			__antithesis_instrumentation__.Notify(3338)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(3341)
		}
		__antithesis_instrumentation__.Notify(3320)
		return matches, nil
	} else {
		__antithesis_instrumentation__.Notify(3342)
	}
	__antithesis_instrumentation__.Notify(3311)

	matches, err := filepath.Glob(fullPath)
	if err != nil {
		__antithesis_instrumentation__.Notify(3343)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(3344)
	}
	__antithesis_instrumentation__.Notify(3312)

	var fileList []string
	for _, file := range matches {
		__antithesis_instrumentation__.Notify(3345)
		fileList = append(fileList, strings.TrimPrefix(file, l.externalIODir))
	}
	__antithesis_instrumentation__.Notify(3313)
	return fileList, nil
}

func (l *LocalStorage) Delete(filename string) error {
	__antithesis_instrumentation__.Notify(3346)
	fullPath, err := l.prependExternalIODir(filename)
	if err != nil {
		__antithesis_instrumentation__.Notify(3348)
		return errors.Wrap(err, "deleting file")
	} else {
		__antithesis_instrumentation__.Notify(3349)
	}
	__antithesis_instrumentation__.Notify(3347)
	return os.Remove(fullPath)
}

func (l *LocalStorage) Stat(filename string) (*blobspb.BlobStat, error) {
	__antithesis_instrumentation__.Notify(3350)
	fullPath, err := l.prependExternalIODir(filename)
	if err != nil {
		__antithesis_instrumentation__.Notify(3354)
		return nil, errors.Wrap(err, "getting stat of file")
	} else {
		__antithesis_instrumentation__.Notify(3355)
	}
	__antithesis_instrumentation__.Notify(3351)
	fi, err := os.Stat(fullPath)
	if err != nil {
		__antithesis_instrumentation__.Notify(3356)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(3357)
	}
	__antithesis_instrumentation__.Notify(3352)
	if fi.IsDir() {
		__antithesis_instrumentation__.Notify(3358)
		return nil, errors.Errorf("expected a file but %q is a directory", fi.Name())
	} else {
		__antithesis_instrumentation__.Notify(3359)
	}
	__antithesis_instrumentation__.Notify(3353)
	return &blobspb.BlobStat{Filesize: fi.Size()}, nil
}
