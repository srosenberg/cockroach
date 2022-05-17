package storage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/sysutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/oserror"
	"github.com/cockroachdb/pebble/vfs"
)

var ballastsEnabled bool = envutil.EnvOrDefaultBool("COCKROACH_AUTO_BALLAST", true)

func IsDiskFull(fs vfs.FS, spec base.StoreSpec) (bool, error) {
	__antithesis_instrumentation__.Notify(633509)
	if spec.InMemory {
		__antithesis_instrumentation__.Notify(633514)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(633515)
	}
	__antithesis_instrumentation__.Notify(633510)

	path := spec.Path
	diskUsage, err := fs.GetDiskUsage(path)
	for oserror.IsNotExist(err) {
		__antithesis_instrumentation__.Notify(633516)
		if parentPath := fs.PathDir(path); parentPath == path {
			__antithesis_instrumentation__.Notify(633518)
			break
		} else {
			__antithesis_instrumentation__.Notify(633519)
			path = parentPath
		}
		__antithesis_instrumentation__.Notify(633517)
		diskUsage, err = fs.GetDiskUsage(path)
	}
	__antithesis_instrumentation__.Notify(633511)
	if err != nil {
		__antithesis_instrumentation__.Notify(633520)
		return false, errors.Wrapf(err, "retrieving disk usage: %s", spec.Path)
	} else {
		__antithesis_instrumentation__.Notify(633521)
	}
	__antithesis_instrumentation__.Notify(633512)

	desiredSizeBytes := BallastSizeBytes(spec, diskUsage)
	ballastPath := base.EmergencyBallastFile(fs.PathJoin, spec.Path)
	if resized, err := maybeEstablishBallast(fs, ballastPath, desiredSizeBytes, diskUsage); err != nil {
		__antithesis_instrumentation__.Notify(633522)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(633523)
		if resized {
			__antithesis_instrumentation__.Notify(633524)
			diskUsage, err = fs.GetDiskUsage(path)
			if err != nil {
				__antithesis_instrumentation__.Notify(633525)
				return false, errors.Wrapf(err, "retrieving disk usage: %s", spec.Path)
			} else {
				__antithesis_instrumentation__.Notify(633526)
			}
		} else {
			__antithesis_instrumentation__.Notify(633527)
		}
	}
	__antithesis_instrumentation__.Notify(633513)

	return diskUsage.AvailBytes < uint64(desiredSizeBytes/2), nil
}

func BallastSizeBytes(spec base.StoreSpec, diskUsage vfs.DiskUsage) int64 {
	__antithesis_instrumentation__.Notify(633528)
	if spec.BallastSize != nil {
		__antithesis_instrumentation__.Notify(633531)
		v := spec.BallastSize.InBytes
		if spec.BallastSize.Percent != 0 {
			__antithesis_instrumentation__.Notify(633533)
			v = int64(float64(diskUsage.TotalBytes) * spec.BallastSize.Percent / 100)
		} else {
			__antithesis_instrumentation__.Notify(633534)
		}
		__antithesis_instrumentation__.Notify(633532)
		return v
	} else {
		__antithesis_instrumentation__.Notify(633535)
	}
	__antithesis_instrumentation__.Notify(633529)

	var v int64 = 1 << 30
	if p := int64(float64(diskUsage.TotalBytes) * 0.01); v > p {
		__antithesis_instrumentation__.Notify(633536)
		v = p
	} else {
		__antithesis_instrumentation__.Notify(633537)
	}
	__antithesis_instrumentation__.Notify(633530)
	return v
}

func maybeEstablishBallast(
	fs vfs.FS, ballastPath string, ballastSizeBytes int64, diskUsage vfs.DiskUsage,
) (resized bool, err error) {
	__antithesis_instrumentation__.Notify(633538)
	var currentSizeBytes int64
	fi, err := fs.Stat(ballastPath)
	if err != nil && func() bool {
		__antithesis_instrumentation__.Notify(633540)
		return !oserror.IsNotExist(err) == true
	}() == true {
		__antithesis_instrumentation__.Notify(633541)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(633542)
		if err == nil {
			__antithesis_instrumentation__.Notify(633543)
			currentSizeBytes = fi.Size()
		} else {
			__antithesis_instrumentation__.Notify(633544)
		}
	}
	__antithesis_instrumentation__.Notify(633539)

	switch {
	case currentSizeBytes > ballastSizeBytes:
		__antithesis_instrumentation__.Notify(633545)

		return true, sysutil.ResizeLargeFile(ballastPath, ballastSizeBytes)
	case currentSizeBytes < ballastSizeBytes && func() bool {
		__antithesis_instrumentation__.Notify(633551)
		return ballastsEnabled == true
	}() == true:
		__antithesis_instrumentation__.Notify(633546)
		if err := fs.MkdirAll(fs.PathDir(ballastPath), 0755); err != nil {
			__antithesis_instrumentation__.Notify(633552)
			return false, errors.Wrap(err, "creating data directory")
		} else {
			__antithesis_instrumentation__.Notify(633553)
		}
		__antithesis_instrumentation__.Notify(633547)

		extendBytes := ballastSizeBytes - currentSizeBytes

		if extendBytes <= int64(diskUsage.AvailBytes)/4 {
			__antithesis_instrumentation__.Notify(633554)
			return true, sysutil.ResizeLargeFile(ballastPath, ballastSizeBytes)
		} else {
			__antithesis_instrumentation__.Notify(633555)
		}
		__antithesis_instrumentation__.Notify(633548)

		if int64(diskUsage.AvailBytes)-extendBytes > (10 << 30) {
			__antithesis_instrumentation__.Notify(633556)
			return true, sysutil.ResizeLargeFile(ballastPath, ballastSizeBytes)
		} else {
			__antithesis_instrumentation__.Notify(633557)
		}
		__antithesis_instrumentation__.Notify(633549)

		return false, nil
	default:
		__antithesis_instrumentation__.Notify(633550)
		return false, nil
	}
}
