package changefeedccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"net/http"
	"strconv"

	"github.com/cockroachdb/cockroach/pkg/util/httputil"
	"github.com/cockroachdb/errors"
)

func strToBool(src string, dest *bool) (wasSet bool, err error) {
	__antithesis_instrumentation__.Notify(18995)
	b, err := strconv.ParseBool(src)
	if err != nil {
		__antithesis_instrumentation__.Notify(18997)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(18998)
	}
	__antithesis_instrumentation__.Notify(18996)
	*dest = b
	return true, nil
}

func decodeBase64FromString(src string, dest *[]byte) error {
	__antithesis_instrumentation__.Notify(18999)
	if src == `` {
		__antithesis_instrumentation__.Notify(19002)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(19003)
	}
	__antithesis_instrumentation__.Notify(19000)
	decoded, err := base64.StdEncoding.DecodeString(src)
	if err != nil {
		__antithesis_instrumentation__.Notify(19004)
		return err
	} else {
		__antithesis_instrumentation__.Notify(19005)
	}
	__antithesis_instrumentation__.Notify(19001)
	*dest = decoded
	return nil
}

func newClientFromTLSKeyPair(caCert []byte) (*httputil.Client, error) {
	__antithesis_instrumentation__.Notify(19006)
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		__antithesis_instrumentation__.Notify(19010)
		return nil, errors.Wrap(err, "could not load system root CA pool")
	} else {
		__antithesis_instrumentation__.Notify(19011)
	}
	__antithesis_instrumentation__.Notify(19007)
	if rootCAs == nil {
		__antithesis_instrumentation__.Notify(19012)
		rootCAs = x509.NewCertPool()
	} else {
		__antithesis_instrumentation__.Notify(19013)
	}
	__antithesis_instrumentation__.Notify(19008)

	if !rootCAs.AppendCertsFromPEM(caCert) {
		__antithesis_instrumentation__.Notify(19014)
		return nil, errors.Errorf("failed to parse certificate data:%s", string(caCert))
	} else {
		__antithesis_instrumentation__.Notify(19015)
	}
	__antithesis_instrumentation__.Notify(19009)

	tlsConfig := &tls.Config{
		RootCAs: rootCAs,
	}

	client := httputil.NewClientWithTimeout(httputil.StandardHTTPTimeout)
	transport := client.Client.Transport.(*http.Transport)
	transport.TLSClientConfig = tlsConfig
	client.Client.Transport = transport

	return client, nil
}
