package cloud

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"net/url"
	"path"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
)

const (
	AuthParam = "AUTH"

	AuthParamImplicit = roachpb.ExternalStorageAuthImplicit

	AuthParamSpecified = roachpb.ExternalStorageAuthSpecified
)

func GetPrefixBeforeWildcard(p string) string {
	__antithesis_instrumentation__.Notify(36629)
	globIndex := strings.IndexAny(p, "*?[")
	if globIndex < 0 {
		__antithesis_instrumentation__.Notify(36631)
		return p
	} else {
		__antithesis_instrumentation__.Notify(36632)
	}
	__antithesis_instrumentation__.Notify(36630)
	return path.Dir(p[:globIndex])
}

func RedactKMSURI(kmsURI string) (string, error) {
	__antithesis_instrumentation__.Notify(36633)
	sanitizedKMSURI, err := SanitizeExternalStorageURI(kmsURI, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(36636)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(36637)
	}
	__antithesis_instrumentation__.Notify(36634)

	uri, err := url.ParseRequestURI(sanitizedKMSURI)
	if err != nil {
		__antithesis_instrumentation__.Notify(36638)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(36639)
	}
	__antithesis_instrumentation__.Notify(36635)
	uri.Path = "/redacted"
	return uri.String(), nil
}

func JoinPathPreservingTrailingSlash(prefix, suffix string) string {
	__antithesis_instrumentation__.Notify(36640)
	out := path.Join(prefix, suffix)

	if strings.HasSuffix(suffix, "/") {
		__antithesis_instrumentation__.Notify(36642)
		out += "/"
	} else {
		__antithesis_instrumentation__.Notify(36643)
	}
	__antithesis_instrumentation__.Notify(36641)
	return out
}
