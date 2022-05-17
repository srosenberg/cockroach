package storage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"io/ioutil"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage/fs"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/oserror"
	"github.com/cockroachdb/pebble/vfs"
)

const MinVersionFilename = "STORAGE_MIN_VERSION"

func writeMinVersionFile(atomicRenameFS vfs.FS, dir string, version roachpb.Version) error {
	__antithesis_instrumentation__.Notify(640214)

	if version == (roachpb.Version{}) {
		__antithesis_instrumentation__.Notify(640220)
		return errors.New("min version should not be empty")
	} else {
		__antithesis_instrumentation__.Notify(640221)
	}
	__antithesis_instrumentation__.Notify(640215)
	ok, err := MinVersionIsAtLeastTargetVersion(atomicRenameFS, dir, version)
	if err != nil {
		__antithesis_instrumentation__.Notify(640222)
		return err
	} else {
		__antithesis_instrumentation__.Notify(640223)
	}
	__antithesis_instrumentation__.Notify(640216)
	if ok {
		__antithesis_instrumentation__.Notify(640224)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(640225)
	}
	__antithesis_instrumentation__.Notify(640217)
	b, err := protoutil.Marshal(&version)
	if err != nil {
		__antithesis_instrumentation__.Notify(640226)
		return err
	} else {
		__antithesis_instrumentation__.Notify(640227)
	}
	__antithesis_instrumentation__.Notify(640218)
	filename := atomicRenameFS.PathJoin(dir, MinVersionFilename)
	if err := fs.SafeWriteToFile(atomicRenameFS, dir, filename, b); err != nil {
		__antithesis_instrumentation__.Notify(640228)
		return err
	} else {
		__antithesis_instrumentation__.Notify(640229)
	}
	__antithesis_instrumentation__.Notify(640219)
	return nil
}

func MinVersionIsAtLeastTargetVersion(
	atomicRenameFS vfs.FS, dir string, target roachpb.Version,
) (bool, error) {
	__antithesis_instrumentation__.Notify(640230)

	if target == (roachpb.Version{}) {
		__antithesis_instrumentation__.Notify(640234)
		return false, errors.New("target version should not be empty")
	} else {
		__antithesis_instrumentation__.Notify(640235)
	}
	__antithesis_instrumentation__.Notify(640231)
	minVersion, err := getMinVersion(atomicRenameFS, dir)
	if err != nil {
		__antithesis_instrumentation__.Notify(640236)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(640237)
	}
	__antithesis_instrumentation__.Notify(640232)
	if minVersion == (roachpb.Version{}) {
		__antithesis_instrumentation__.Notify(640238)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(640239)
	}
	__antithesis_instrumentation__.Notify(640233)
	return !minVersion.Less(target), nil
}

func getMinVersion(atomicRenameFS vfs.FS, dir string) (roachpb.Version, error) {
	__antithesis_instrumentation__.Notify(640240)

	filename := atomicRenameFS.PathJoin(dir, MinVersionFilename)
	f, err := atomicRenameFS.Open(filename)
	if oserror.IsNotExist(err) {
		__antithesis_instrumentation__.Notify(640245)
		return roachpb.Version{}, nil
	} else {
		__antithesis_instrumentation__.Notify(640246)
	}
	__antithesis_instrumentation__.Notify(640241)
	if err != nil {
		__antithesis_instrumentation__.Notify(640247)
		return roachpb.Version{}, err
	} else {
		__antithesis_instrumentation__.Notify(640248)
	}
	__antithesis_instrumentation__.Notify(640242)
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		__antithesis_instrumentation__.Notify(640249)
		return roachpb.Version{}, err
	} else {
		__antithesis_instrumentation__.Notify(640250)
	}
	__antithesis_instrumentation__.Notify(640243)
	version := roachpb.Version{}
	if err := protoutil.Unmarshal(b, &version); err != nil {
		__antithesis_instrumentation__.Notify(640251)
		return version, err
	} else {
		__antithesis_instrumentation__.Notify(640252)
	}
	__antithesis_instrumentation__.Notify(640244)
	return version, nil
}
