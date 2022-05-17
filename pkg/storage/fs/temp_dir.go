package fs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/oserror"
	"github.com/cockroachdb/pebble/vfs"
)

const lockFilename = `TEMP_DIR.LOCK`

type lockStruct struct {
	closer io.Closer
}

func lockFile(filename string) (lockStruct, error) {
	__antithesis_instrumentation__.Notify(639136)
	closer, err := vfs.Default.Lock(filename)
	if err != nil {
		__antithesis_instrumentation__.Notify(639138)
		return lockStruct{}, err
	} else {
		__antithesis_instrumentation__.Notify(639139)
	}
	__antithesis_instrumentation__.Notify(639137)
	return lockStruct{closer: closer}, nil
}

func unlockFile(lock lockStruct) error {
	__antithesis_instrumentation__.Notify(639140)
	if lock.closer != nil {
		__antithesis_instrumentation__.Notify(639142)
		return lock.closer.Close()
	} else {
		__antithesis_instrumentation__.Notify(639143)
	}
	__antithesis_instrumentation__.Notify(639141)
	return nil
}

func CreateTempDir(parentDir, prefix string, stopper *stop.Stopper) (string, error) {
	__antithesis_instrumentation__.Notify(639144)

	tempPath, err := ioutil.TempDir(parentDir, prefix)
	if err != nil {
		__antithesis_instrumentation__.Notify(639150)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(639151)
	}
	__antithesis_instrumentation__.Notify(639145)

	if err := os.Chmod(tempPath, 0755); err != nil {
		__antithesis_instrumentation__.Notify(639152)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(639153)
	}
	__antithesis_instrumentation__.Notify(639146)

	absPath, err := filepath.Abs(tempPath)
	if err != nil {
		__antithesis_instrumentation__.Notify(639154)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(639155)
	}
	__antithesis_instrumentation__.Notify(639147)

	flock, err := lockFile(filepath.Join(absPath, lockFilename))
	if err != nil {
		__antithesis_instrumentation__.Notify(639156)
		return "", errors.Wrapf(err, "could not create lock on new temporary directory")
	} else {
		__antithesis_instrumentation__.Notify(639157)
	}
	__antithesis_instrumentation__.Notify(639148)
	stopper.AddCloser(stop.CloserFn(func() {
		__antithesis_instrumentation__.Notify(639158)
		if err := unlockFile(flock); err != nil {
			__antithesis_instrumentation__.Notify(639159)
			log.Errorf(context.TODO(), "could not unlock file lock on temporary directory: %s", err.Error())
		} else {
			__antithesis_instrumentation__.Notify(639160)
		}
	}))
	__antithesis_instrumentation__.Notify(639149)

	return absPath, nil
}

func RecordTempDir(recordPath, tempPath string) error {
	__antithesis_instrumentation__.Notify(639161)

	f, err := os.OpenFile(recordPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		__antithesis_instrumentation__.Notify(639163)
		return err
	} else {
		__antithesis_instrumentation__.Notify(639164)
	}
	__antithesis_instrumentation__.Notify(639162)
	defer f.Close()

	_, err = f.Write(append([]byte(tempPath), '\n'))
	return err
}

func CleanupTempDirs(recordPath string) error {
	__antithesis_instrumentation__.Notify(639165)

	f, err := os.OpenFile(recordPath, os.O_RDWR, 0644)

	if oserror.IsNotExist(err) {
		__antithesis_instrumentation__.Notify(639169)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(639170)
	}
	__antithesis_instrumentation__.Notify(639166)
	if err != nil {
		__antithesis_instrumentation__.Notify(639171)
		return err
	} else {
		__antithesis_instrumentation__.Notify(639172)
	}
	__antithesis_instrumentation__.Notify(639167)
	defer f.Close()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		__antithesis_instrumentation__.Notify(639173)
		path := scanner.Text()
		if path == "" {
			__antithesis_instrumentation__.Notify(639178)
			continue
		} else {
			__antithesis_instrumentation__.Notify(639179)
		}
		__antithesis_instrumentation__.Notify(639174)

		if _, err := os.Stat(path); oserror.IsNotExist(err) {
			__antithesis_instrumentation__.Notify(639180)
			log.Warningf(context.Background(), "could not locate previous temporary directory %s, might require manual cleanup, or might have already been cleaned up.", path)
			continue
		} else {
			__antithesis_instrumentation__.Notify(639181)
		}
		__antithesis_instrumentation__.Notify(639175)

		flock, err := lockFile(filepath.Join(path, lockFilename))
		if err != nil {
			__antithesis_instrumentation__.Notify(639182)
			return errors.Wrapf(err, "could not lock temporary directory %s, may still be in use", path)
		} else {
			__antithesis_instrumentation__.Notify(639183)
		}
		__antithesis_instrumentation__.Notify(639176)

		if err := unlockFile(flock); err != nil {
			__antithesis_instrumentation__.Notify(639184)
			log.Errorf(context.TODO(), "could not unlock file lock when removing temporary directory: %s", err.Error())
		} else {
			__antithesis_instrumentation__.Notify(639185)
		}
		__antithesis_instrumentation__.Notify(639177)

		if err := os.RemoveAll(path); err != nil {
			__antithesis_instrumentation__.Notify(639186)
			return err
		} else {
			__antithesis_instrumentation__.Notify(639187)
		}
	}
	__antithesis_instrumentation__.Notify(639168)

	return f.Truncate(0)
}
