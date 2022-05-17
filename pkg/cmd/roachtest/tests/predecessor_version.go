package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/util/version"
	"github.com/cockroachdb/errors"
)

func PredecessorVersion(buildVersion version.Version) (string, error) {
	__antithesis_instrumentation__.Notify(49822)
	if buildVersion == (version.Version{}) {
		__antithesis_instrumentation__.Notify(49825)
		return "", errors.Errorf("buildVersion not set")
	} else {
		__antithesis_instrumentation__.Notify(49826)
	}
	__antithesis_instrumentation__.Notify(49823)

	buildVersionMajorMinor := fmt.Sprintf("%d.%d", buildVersion.Major(), buildVersion.Minor())

	verMap := map[string]string{
		"22.1": "21.2.9",
		"21.2": "21.1.12",
		"21.1": "20.2.12",
		"20.2": "20.1.16",
		"20.1": "19.2.11",
		"19.2": "19.1.11",
		"19.1": "2.1.9",
		"2.2":  "2.1.9",
		"2.1":  "2.0.7",
	}
	v, ok := verMap[buildVersionMajorMinor]
	if !ok {
		__antithesis_instrumentation__.Notify(49827)
		return "", errors.Errorf("prev version not set for version: %s", buildVersionMajorMinor)
	} else {
		__antithesis_instrumentation__.Notify(49828)
	}
	__antithesis_instrumentation__.Notify(49824)
	return v, nil
}
