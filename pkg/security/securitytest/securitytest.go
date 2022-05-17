// Package securitytest embeds the TLS test certificates.
package securitytest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/errors"
)

func RestrictedCopy(path, tempdir, name string) (string, error) {
	__antithesis_instrumentation__.Notify(187205)
	contents, err := Asset(path)
	if err != nil {
		__antithesis_instrumentation__.Notify(187208)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(187209)
	}
	__antithesis_instrumentation__.Notify(187206)
	tempPath := filepath.Join(tempdir, name)
	if err := ioutil.WriteFile(tempPath, contents, 0600); err != nil {
		__antithesis_instrumentation__.Notify(187210)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(187211)
	}
	__antithesis_instrumentation__.Notify(187207)
	return tempPath, nil
}

func AppendFile(assetPath, dstPath string) error {
	__antithesis_instrumentation__.Notify(187212)
	contents, err := Asset(assetPath)
	if err != nil {
		__antithesis_instrumentation__.Notify(187215)
		return err
	} else {
		__antithesis_instrumentation__.Notify(187216)
	}
	__antithesis_instrumentation__.Notify(187213)
	f, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		__antithesis_instrumentation__.Notify(187217)
		return err
	} else {
		__antithesis_instrumentation__.Notify(187218)
	}
	__antithesis_instrumentation__.Notify(187214)
	_, err = f.Write(contents)
	return errors.CombineErrors(err, f.Close())
}

func AssetReadDir(name string) ([]os.FileInfo, error) {
	__antithesis_instrumentation__.Notify(187219)
	names, err := AssetDir(name)
	if err != nil {
		__antithesis_instrumentation__.Notify(187222)
		if strings.HasSuffix(err.Error(), "not found") {
			__antithesis_instrumentation__.Notify(187224)
			return nil, os.ErrNotExist
		} else {
			__antithesis_instrumentation__.Notify(187225)
		}
		__antithesis_instrumentation__.Notify(187223)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187226)
	}
	__antithesis_instrumentation__.Notify(187220)
	infos := make([]os.FileInfo, 0, len(names))
	for _, n := range names {
		__antithesis_instrumentation__.Notify(187227)
		joined := filepath.Join(name, n)
		info, err := AssetInfo(joined)
		if err != nil {
			__antithesis_instrumentation__.Notify(187229)
			if _, dirErr := AssetDir(joined); dirErr != nil {
				__antithesis_instrumentation__.Notify(187231)

				return nil, errors.Wrapf(err, "missing directory (%v)", dirErr)
			} else {
				__antithesis_instrumentation__.Notify(187232)
			}
			__antithesis_instrumentation__.Notify(187230)
			continue
		} else {
			__antithesis_instrumentation__.Notify(187233)
		}
		__antithesis_instrumentation__.Notify(187228)

		binInfo := info.(bindataFileInfo)
		binInfo.name = filepath.Base(binInfo.name)
		infos = append(infos, binInfo)
	}
	__antithesis_instrumentation__.Notify(187221)
	return infos, nil
}

func AssetStat(name string) (os.FileInfo, error) {
	__antithesis_instrumentation__.Notify(187234)
	info, err := AssetInfo(name)
	if err != nil && func() bool {
		__antithesis_instrumentation__.Notify(187236)
		return strings.HasSuffix(err.Error(), "not found") == true
	}() == true {
		__antithesis_instrumentation__.Notify(187237)
		return info, os.ErrNotExist
	} else {
		__antithesis_instrumentation__.Notify(187238)
	}
	__antithesis_instrumentation__.Notify(187235)
	return info, err
}

var EmbeddedAssets = security.AssetLoader{
	ReadDir:  AssetReadDir,
	ReadFile: Asset,
	Stat:     AssetStat,
}
