package clusterversion

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
	"github.com/kr/pretty"
)

type keyedVersion struct {
	Key Key
	roachpb.Version
}

type keyedVersions []keyedVersion

func (kv keyedVersions) MustByKey(k Key) roachpb.Version {
	__antithesis_instrumentation__.Notify(37228)
	key := int(k)
	if key >= len(kv) || func() bool {
		__antithesis_instrumentation__.Notify(37230)
		return key < 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(37231)
		log.Fatalf(context.Background(), "version with key %d does not exist, have:\n%s",
			key, pretty.Sprint(kv))
	} else {
		__antithesis_instrumentation__.Notify(37232)
	}
	__antithesis_instrumentation__.Notify(37229)
	return kv[key].Version
}

func (kv keyedVersions) Validate() error {
	__antithesis_instrumentation__.Notify(37233)
	type majorMinor struct {
		major, minor int32
		vs           []keyedVersion
	}

	var byRelease []majorMinor
	for i, namedVersion := range kv {
		__antithesis_instrumentation__.Notify(37237)
		if int(namedVersion.Key) != i {
			__antithesis_instrumentation__.Notify(37241)
			return errors.Errorf("version %s should have key %d but has %d",
				namedVersion, i, namedVersion.Key)
		} else {
			__antithesis_instrumentation__.Notify(37242)
		}
		__antithesis_instrumentation__.Notify(37238)
		if i > 0 {
			__antithesis_instrumentation__.Notify(37243)
			prev := kv[i-1]
			if !prev.Version.Less(namedVersion.Version) {
				__antithesis_instrumentation__.Notify(37244)
				return errors.Errorf("version %s must be larger than %s", namedVersion, prev)
			} else {
				__antithesis_instrumentation__.Notify(37245)
			}
		} else {
			__antithesis_instrumentation__.Notify(37246)
		}
		__antithesis_instrumentation__.Notify(37239)
		mami := majorMinor{major: namedVersion.Major, minor: namedVersion.Minor}
		n := len(byRelease)
		if n == 0 || func() bool {
			__antithesis_instrumentation__.Notify(37247)
			return byRelease[n-1].major != mami.major == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(37248)
			return byRelease[n-1].minor != mami.minor == true
		}() == true {
			__antithesis_instrumentation__.Notify(37249)

			byRelease = append(byRelease, mami)
			n++
		} else {
			__antithesis_instrumentation__.Notify(37250)
		}
		__antithesis_instrumentation__.Notify(37240)

		byRelease[n-1].vs = append(byRelease[n-1].vs, namedVersion)
	}
	__antithesis_instrumentation__.Notify(37234)

	if n := len(byRelease) - 5; n >= 0 {
		__antithesis_instrumentation__.Notify(37251)
		var buf strings.Builder
		for i, mami := range byRelease[:n+1] {
			__antithesis_instrumentation__.Notify(37253)
			s := "next release"
			if i+1 < len(byRelease)-1 {
				__antithesis_instrumentation__.Notify(37255)
				nextMM := byRelease[i+1]
				s = fmt.Sprintf("%d.%d", nextMM.major, nextMM.minor)
			} else {
				__antithesis_instrumentation__.Notify(37256)
			}
			__antithesis_instrumentation__.Notify(37254)
			for _, nv := range mami.vs {
				__antithesis_instrumentation__.Notify(37257)
				fmt.Fprintf(&buf, "introduced in %s: %s\n", s, nv.Key)
			}
		}
		__antithesis_instrumentation__.Notify(37252)
		mostRecentRelease := byRelease[len(byRelease)-1]
		return errors.Errorf(
			"found versions that are always active because %d.%d is already "+
				"released; these should be removed:\n%s",
			mostRecentRelease.minor, mostRecentRelease.major,
			buf.String(),
		)
	} else {
		__antithesis_instrumentation__.Notify(37258)
	}
	__antithesis_instrumentation__.Notify(37235)

	for _, release := range byRelease {
		__antithesis_instrumentation__.Notify(37259)

		if release.major < 20 {
			__antithesis_instrumentation__.Notify(37262)
			continue
		} else {
			__antithesis_instrumentation__.Notify(37263)
		}
		__antithesis_instrumentation__.Notify(37260)
		if release.major == 20 && func() bool {
			__antithesis_instrumentation__.Notify(37264)
			return release.minor == 1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(37265)
			continue
		} else {
			__antithesis_instrumentation__.Notify(37266)
		}
		__antithesis_instrumentation__.Notify(37261)

		for _, v := range release.vs {
			__antithesis_instrumentation__.Notify(37267)
			if (v.Internal % 2) != 0 {
				__antithesis_instrumentation__.Notify(37268)
				return errors.Errorf("found version %s with odd-numbered internal version (%s);"+
					" versions introduced in 21.1+ must have even-numbered internal versions", v.Key, v.Version)
			} else {
				__antithesis_instrumentation__.Notify(37269)
			}
		}
	}
	__antithesis_instrumentation__.Notify(37236)

	return nil
}
