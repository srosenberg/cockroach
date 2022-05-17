package licenseccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"encoding/base64"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
)

const LicensePrefix = "crl-0-"

func (l *License) Encode() (string, error) {
	__antithesis_instrumentation__.Notify(27249)
	bytes, err := protoutil.Marshal(l)
	if err != nil {
		__antithesis_instrumentation__.Notify(27251)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(27252)
	}
	__antithesis_instrumentation__.Notify(27250)
	return LicensePrefix + base64.RawStdEncoding.EncodeToString(bytes), nil
}

func Decode(s string) (*License, error) {
	__antithesis_instrumentation__.Notify(27253)
	if s == "" {
		__antithesis_instrumentation__.Notify(27258)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(27259)
	}
	__antithesis_instrumentation__.Notify(27254)
	if !strings.HasPrefix(s, LicensePrefix) {
		__antithesis_instrumentation__.Notify(27260)
		return nil, errors.New("invalid license string")
	} else {
		__antithesis_instrumentation__.Notify(27261)
	}
	__antithesis_instrumentation__.Notify(27255)
	s = strings.TrimPrefix(s, LicensePrefix)
	data, err := base64.RawStdEncoding.DecodeString(s)
	if err != nil {
		__antithesis_instrumentation__.Notify(27262)
		return nil, errors.Wrapf(err, "invalid license string")
	} else {
		__antithesis_instrumentation__.Notify(27263)
	}
	__antithesis_instrumentation__.Notify(27256)
	var lic License
	if err := protoutil.Unmarshal(data, &lic); err != nil {
		__antithesis_instrumentation__.Notify(27264)
		return nil, errors.Wrap(err, "invalid license string")
	} else {
		__antithesis_instrumentation__.Notify(27265)
	}
	__antithesis_instrumentation__.Notify(27257)
	return &lic, nil
}
