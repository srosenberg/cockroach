package sqlutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/security/securitytest"
	"github.com/cockroachdb/cockroach/pkg/util/fileutil"
)

func PGUrl(t testing.TB, servingAddr, prefix string, user *url.Userinfo) (url.URL, func()) {
	__antithesis_instrumentation__.Notify(646145)
	return PGUrlWithOptionalClientCerts(t, servingAddr, prefix, user, true)
}

func PGUrlE(servingAddr, prefix string, user *url.Userinfo) (url.URL, func(), error) {
	__antithesis_instrumentation__.Notify(646146)
	return PGUrlWithOptionalClientCertsE(servingAddr, prefix, user, true)
}

func PGUrlWithOptionalClientCerts(
	t testing.TB, servingAddr, prefix string, user *url.Userinfo, withClientCerts bool,
) (url.URL, func()) {
	__antithesis_instrumentation__.Notify(646147)
	u, f, err := PGUrlWithOptionalClientCertsE(servingAddr, prefix, user, withClientCerts)
	if err != nil {
		__antithesis_instrumentation__.Notify(646149)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(646150)
	}
	__antithesis_instrumentation__.Notify(646148)
	return u, f
}

func PGUrlWithOptionalClientCertsE(
	servingAddr, prefix string, user *url.Userinfo, withClientCerts bool,
) (url.URL, func(), error) {
	__antithesis_instrumentation__.Notify(646151)
	host, port, err := net.SplitHostPort(servingAddr)
	if err != nil {
		__antithesis_instrumentation__.Notify(646157)
		return url.URL{}, func() { __antithesis_instrumentation__.Notify(646158) }, err
	} else {
		__antithesis_instrumentation__.Notify(646159)
	}
	__antithesis_instrumentation__.Notify(646152)

	tempDir, err := ioutil.TempDir("", fileutil.EscapeFilename(prefix))
	if err != nil {
		__antithesis_instrumentation__.Notify(646160)
		return url.URL{}, func() { __antithesis_instrumentation__.Notify(646161) }, err
	} else {
		__antithesis_instrumentation__.Notify(646162)
	}
	__antithesis_instrumentation__.Notify(646153)

	caPath := filepath.Join(security.EmbeddedCertsDir, security.EmbeddedCACert)
	tempCAPath, err := securitytest.RestrictedCopy(caPath, tempDir, "ca")
	if err != nil {
		__antithesis_instrumentation__.Notify(646163)
		return url.URL{}, func() { __antithesis_instrumentation__.Notify(646164) }, err
	} else {
		__antithesis_instrumentation__.Notify(646165)
	}
	__antithesis_instrumentation__.Notify(646154)

	tenantCAPath := filepath.Join(security.EmbeddedCertsDir, security.EmbeddedTenantCACert)
	if err := securitytest.AppendFile(tenantCAPath, tempCAPath); err != nil {
		__antithesis_instrumentation__.Notify(646166)
		return url.URL{}, func() { __antithesis_instrumentation__.Notify(646167) }, err
	} else {
		__antithesis_instrumentation__.Notify(646168)
	}
	__antithesis_instrumentation__.Notify(646155)
	options := url.Values{}
	options.Add("sslrootcert", tempCAPath)

	if withClientCerts {
		__antithesis_instrumentation__.Notify(646169)
		certPath := filepath.Join(security.EmbeddedCertsDir, fmt.Sprintf("client.%s.crt", user.Username()))
		keyPath := filepath.Join(security.EmbeddedCertsDir, fmt.Sprintf("client.%s.key", user.Username()))

		tempCertPath, err := securitytest.RestrictedCopy(certPath, tempDir, "cert")
		if err != nil {
			__antithesis_instrumentation__.Notify(646172)
			return url.URL{}, func() { __antithesis_instrumentation__.Notify(646173) }, err
		} else {
			__antithesis_instrumentation__.Notify(646174)
		}
		__antithesis_instrumentation__.Notify(646170)
		tempKeyPath, err := securitytest.RestrictedCopy(keyPath, tempDir, "key")
		if err != nil {
			__antithesis_instrumentation__.Notify(646175)
			return url.URL{}, func() { __antithesis_instrumentation__.Notify(646176) }, err
		} else {
			__antithesis_instrumentation__.Notify(646177)
		}
		__antithesis_instrumentation__.Notify(646171)
		options.Add("sslcert", tempCertPath)
		options.Add("sslkey", tempKeyPath)
		options.Add("sslmode", "verify-full")
	} else {
		__antithesis_instrumentation__.Notify(646178)
		options.Add("sslmode", "verify-ca")
	}
	__antithesis_instrumentation__.Notify(646156)

	return url.URL{
		Scheme:   "postgres",
		User:     user,
		Host:     net.JoinHostPort(host, port),
		RawQuery: options.Encode(),
	}, func() { __antithesis_instrumentation__.Notify(646179); _ = os.RemoveAll(tempDir) }, nil
}
