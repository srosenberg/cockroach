package storage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"path/filepath"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/elastic/gosigar"
)

func computeStoreProperties(
	ctx context.Context, dir string, readonly bool, encryptionEnabled bool,
) roachpb.StoreProperties {
	__antithesis_instrumentation__.Notify(643936)
	props := roachpb.StoreProperties{
		ReadOnly:  readonly,
		Encrypted: encryptionEnabled,
	}

	if dir == "" {
		__antithesis_instrumentation__.Notify(643938)
		return props
	} else {
		__antithesis_instrumentation__.Notify(643939)
	}
	__antithesis_instrumentation__.Notify(643937)

	fsprops := getFileSystemProperties(ctx, dir)
	props.FileStoreProperties = &fsprops
	return props
}

func getFileSystemProperties(ctx context.Context, dir string) roachpb.FileStoreProperties {
	__antithesis_instrumentation__.Notify(643940)
	fsprops := roachpb.FileStoreProperties{
		Path: dir,
	}

	absPath, err := filepath.Abs(dir)
	if err != nil {
		__antithesis_instrumentation__.Notify(643945)
		log.Warningf(ctx, "cannot compute absolute file path for %q: %v", dir, err)
		return fsprops
	} else {
		__antithesis_instrumentation__.Notify(643946)
	}
	__antithesis_instrumentation__.Notify(643941)

	var fslist gosigar.FileSystemList
	if err := fslist.Get(); err != nil {
		__antithesis_instrumentation__.Notify(643947)
		log.Warningf(ctx, "cannot retrieve filesystem list: %v", err)
		return fsprops
	} else {
		__antithesis_instrumentation__.Notify(643948)
	}
	__antithesis_instrumentation__.Notify(643942)

	var fsInfo *gosigar.FileSystem

	for i := len(fslist.List) - 1; i >= 0; i-- {
		__antithesis_instrumentation__.Notify(643949)

		_, err := filepath.Rel(fslist.List[i].DirName, absPath)
		if err == nil {
			__antithesis_instrumentation__.Notify(643950)
			fsInfo = &fslist.List[i]
			break
		} else {
			__antithesis_instrumentation__.Notify(643951)
		}
	}
	__antithesis_instrumentation__.Notify(643943)
	if fsInfo == nil {
		__antithesis_instrumentation__.Notify(643952)

		return fsprops
	} else {
		__antithesis_instrumentation__.Notify(643953)
	}
	__antithesis_instrumentation__.Notify(643944)

	fsprops.FsType = fsInfo.SysTypeName
	fsprops.BlockDevice = fsInfo.DevName
	fsprops.MountPoint = fsInfo.DirName
	fsprops.MountOptions = fsInfo.Options
	return fsprops
}
