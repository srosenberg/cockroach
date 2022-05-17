package storageccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/ioctx"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/pebble/sstable"
)

var remoteSSTs = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"kv.bulk_ingest.stream_external_ssts.enabled",
	"if enabled, external SSTables are iterated directly in some cases, rather than being downloaded entirely first",
	true,
)

var remoteSSTSuffixCacheSize = settings.RegisterByteSizeSetting(
	settings.TenantWritable,
	"kv.bulk_ingest.stream_external_ssts.suffix_cache_size",
	"size of suffix of remote SSTs to download and cache before reading from remote stream",
	64<<10,
)

func ExternalSSTReader(
	ctx context.Context,
	e cloud.ExternalStorage,
	basename string,
	encryption *roachpb.FileEncryptionOptions,
) (storage.SimpleMVCCIterator, error) {
	__antithesis_instrumentation__.Notify(24855)

	var f ioctx.ReadCloserCtx
	var sz int64

	const maxAttempts = 3
	if err := retry.WithMaxAttempts(ctx, base.DefaultRetryOptions(), maxAttempts, func() error {
		__antithesis_instrumentation__.Notify(24861)
		var err error
		f, sz, err = e.ReadFileAt(ctx, basename, 0)
		return err
	}); err != nil {
		__antithesis_instrumentation__.Notify(24862)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(24863)
	}
	__antithesis_instrumentation__.Notify(24856)

	if !remoteSSTs.Get(&e.Settings().SV) {
		__antithesis_instrumentation__.Notify(24864)
		content, err := ioctx.ReadAll(ctx, f)
		f.Close(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(24867)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(24868)
		}
		__antithesis_instrumentation__.Notify(24865)
		if encryption != nil {
			__antithesis_instrumentation__.Notify(24869)
			content, err = DecryptFile(ctx, content, encryption.Key, nil)
			if err != nil {
				__antithesis_instrumentation__.Notify(24870)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(24871)
			}
		} else {
			__antithesis_instrumentation__.Notify(24872)
		}
		__antithesis_instrumentation__.Notify(24866)
		return storage.NewMemSSTIterator(content, false)
	} else {
		__antithesis_instrumentation__.Notify(24873)
	}
	__antithesis_instrumentation__.Notify(24857)

	raw := &sstReader{
		ctx:  ctx,
		sz:   sizeStat(sz),
		body: f,
		openAt: func(offset int64) (ioctx.ReadCloserCtx, error) {
			__antithesis_instrumentation__.Notify(24874)
			reader, _, err := e.ReadFileAt(ctx, basename, offset)
			return reader, err
		},
	}
	__antithesis_instrumentation__.Notify(24858)

	var reader sstable.ReadableFile = raw

	if encryption != nil {
		__antithesis_instrumentation__.Notify(24875)
		r, err := decryptingReader(raw, encryption.Key)
		if err != nil {
			__antithesis_instrumentation__.Notify(24877)
			f.Close(ctx)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(24878)
		}
		__antithesis_instrumentation__.Notify(24876)
		reader = r
	} else {
		__antithesis_instrumentation__.Notify(24879)

		if err := raw.readAndCacheSuffix(remoteSSTSuffixCacheSize.Get(&e.Settings().SV)); err != nil {
			__antithesis_instrumentation__.Notify(24880)
			f.Close(ctx)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(24881)
		}
	}
	__antithesis_instrumentation__.Notify(24859)

	iter, err := storage.NewSSTIterator(reader)
	if err != nil {
		__antithesis_instrumentation__.Notify(24882)
		reader.Close()
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(24883)
	}
	__antithesis_instrumentation__.Notify(24860)

	return iter, nil
}

type sstReader struct {
	ctx    context.Context
	sz     sizeStat
	openAt func(int64) (ioctx.ReadCloserCtx, error)

	body ioctx.ReadCloserCtx
	pos  int64

	readPos int64

	cache struct {
		pos int64
		buf []byte
	}
}

func (r *sstReader) reset() error {
	__antithesis_instrumentation__.Notify(24884)
	r.pos = 0
	var err error
	if r.body != nil {
		__antithesis_instrumentation__.Notify(24886)
		err = r.body.Close(r.ctx)
		r.body = nil
	} else {
		__antithesis_instrumentation__.Notify(24887)
	}
	__antithesis_instrumentation__.Notify(24885)
	return err
}

func (r *sstReader) Close() error {
	__antithesis_instrumentation__.Notify(24888)
	err := r.reset()
	r.ctx = nil
	return err
}

func (r *sstReader) Stat() (os.FileInfo, error) {
	__antithesis_instrumentation__.Notify(24889)
	return r.sz, nil
}

func (r *sstReader) Read(p []byte) (int, error) {
	__antithesis_instrumentation__.Notify(24890)
	n, err := r.ReadAt(p, r.readPos)
	r.readPos += int64(n)
	return n, err
}

func (r *sstReader) readAndCacheSuffix(size int64) error {
	__antithesis_instrumentation__.Notify(24891)
	if size == 0 {
		__antithesis_instrumentation__.Notify(24896)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(24897)
	}
	__antithesis_instrumentation__.Notify(24892)
	r.cache.buf = nil
	r.cache.pos = int64(r.sz) - size
	if r.cache.pos <= 0 {
		__antithesis_instrumentation__.Notify(24898)
		r.cache.pos = 0
	} else {
		__antithesis_instrumentation__.Notify(24899)
	}
	__antithesis_instrumentation__.Notify(24893)
	reader, err := r.openAt(r.cache.pos)
	if err != nil {
		__antithesis_instrumentation__.Notify(24900)
		return err
	} else {
		__antithesis_instrumentation__.Notify(24901)
	}
	__antithesis_instrumentation__.Notify(24894)
	defer reader.Close(r.ctx)
	read, err := ioctx.ReadAll(r.ctx, reader)
	if err != nil {
		__antithesis_instrumentation__.Notify(24902)
		return err
	} else {
		__antithesis_instrumentation__.Notify(24903)
	}
	__antithesis_instrumentation__.Notify(24895)
	r.cache.buf = read
	return nil
}

func (r *sstReader) ReadAt(p []byte, offset int64) (int, error) {
	__antithesis_instrumentation__.Notify(24904)
	var read int
	if offset >= r.cache.pos && func() bool {
		__antithesis_instrumentation__.Notify(24910)
		return offset < r.cache.pos+int64(len(r.cache.buf)) == true
	}() == true {
		__antithesis_instrumentation__.Notify(24911)
		read += copy(p, r.cache.buf[offset-r.cache.pos:])
		if read == len(p) {
			__antithesis_instrumentation__.Notify(24913)
			return read, nil
		} else {
			__antithesis_instrumentation__.Notify(24914)
		}
		__antithesis_instrumentation__.Notify(24912)

		offset += int64(read)
	} else {
		__antithesis_instrumentation__.Notify(24915)
	}
	__antithesis_instrumentation__.Notify(24905)

	if offset == int64(r.sz) {
		__antithesis_instrumentation__.Notify(24916)
		return read, io.EOF
	} else {
		__antithesis_instrumentation__.Notify(24917)
	}
	__antithesis_instrumentation__.Notify(24906)

	if r.pos != offset {
		__antithesis_instrumentation__.Notify(24918)
		if err := r.reset(); err != nil {
			__antithesis_instrumentation__.Notify(24921)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(24922)
		}
		__antithesis_instrumentation__.Notify(24919)
		b, err := r.openAt(offset)
		if err != nil {
			__antithesis_instrumentation__.Notify(24923)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(24924)
		}
		__antithesis_instrumentation__.Notify(24920)
		r.pos = offset
		r.body = b
	} else {
		__antithesis_instrumentation__.Notify(24925)
	}
	__antithesis_instrumentation__.Notify(24907)

	var err error
	for n := 0; read < len(p); n, err = r.body.Read(r.ctx, p[read:]) {
		__antithesis_instrumentation__.Notify(24926)
		read += n
		if err != nil {
			__antithesis_instrumentation__.Notify(24927)
			break
		} else {
			__antithesis_instrumentation__.Notify(24928)
		}
	}
	__antithesis_instrumentation__.Notify(24908)
	r.pos += int64(read)

	if read == len(p) && func() bool {
		__antithesis_instrumentation__.Notify(24929)
		return err == io.EOF == true
	}() == true {
		__antithesis_instrumentation__.Notify(24930)
		return read, nil
	} else {
		__antithesis_instrumentation__.Notify(24931)
	}
	__antithesis_instrumentation__.Notify(24909)

	return read, err
}

type sizeStat int64

func (s sizeStat) Size() int64 { __antithesis_instrumentation__.Notify(24932); return int64(s) }
func (sizeStat) IsDir() bool {
	__antithesis_instrumentation__.Notify(24933)
	panic(errors.AssertionFailedf("unimplemented"))
}
func (sizeStat) ModTime() time.Time {
	__antithesis_instrumentation__.Notify(24934)
	panic(errors.AssertionFailedf("unimplemented"))
}
func (sizeStat) Mode() os.FileMode {
	__antithesis_instrumentation__.Notify(24935)
	panic(errors.AssertionFailedf("unimplemented"))
}
func (sizeStat) Name() string {
	__antithesis_instrumentation__.Notify(24936)
	panic(errors.AssertionFailedf("unimplemented"))
}
func (sizeStat) Sys() interface{} {
	__antithesis_instrumentation__.Notify(24937)
	panic(errors.AssertionFailedf("unimplemented"))
}
