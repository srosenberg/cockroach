package pgurl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

func (u *URL) Validate() error {
	__antithesis_instrumentation__.Notify(195325)
	var details bytes.Buffer
	var incorrect []redact.RedactableString

	switch u.net {
	case ProtoUnix:
		__antithesis_instrumentation__.Notify(195330)
		if !strings.HasPrefix(u.host, "/") {
			__antithesis_instrumentation__.Notify(195334)
			incorrect = append(incorrect, "host")
			fmt.Fprintln(&details, "Host parameter must start with '/' when using unix sockets.")
		} else {
			__antithesis_instrumentation__.Notify(195335)
		}
		__antithesis_instrumentation__.Notify(195331)
		if u.sec != tnUnspecified && func() bool {
			__antithesis_instrumentation__.Notify(195336)
			return u.sec != tnNone == true
		}() == true {
			__antithesis_instrumentation__.Notify(195337)
			incorrect = append(incorrect, "sslmode")
			fmt.Fprintln(&details, "Cannot specify TLS settings when using unix sockets.")
		} else {
			__antithesis_instrumentation__.Notify(195338)
		}
	case ProtoTCP:
		__antithesis_instrumentation__.Notify(195332)
		if strings.Contains(u.host, "/") {
			__antithesis_instrumentation__.Notify(195339)
			incorrect = append(incorrect, "host")
			fmt.Fprintln(&details, "Host parameter cannot contain '/' when using TCP.")
		} else {
			__antithesis_instrumentation__.Notify(195340)
		}
	default:
		__antithesis_instrumentation__.Notify(195333)
		incorrect = append(incorrect, "net")
		fmt.Fprintln(&details, "Network protocol unspecified.")
	}
	__antithesis_instrumentation__.Notify(195326)

	if u.username == "" && func() bool {
		__antithesis_instrumentation__.Notify(195341)
		return u.hasPassword == true
	}() == true {
		__antithesis_instrumentation__.Notify(195342)
		incorrect = append(incorrect, "user")
		fmt.Fprintln(&details, "Username cannot be empty when a password is provided.")
	} else {
		__antithesis_instrumentation__.Notify(195343)
	}
	__antithesis_instrumentation__.Notify(195327)

	switch u.authn {
	case authnClientCert, authnPasswordWithClientCert:
		__antithesis_instrumentation__.Notify(195344)
		if u.sec == tnUnspecified || func() bool {
			__antithesis_instrumentation__.Notify(195349)
			return u.sec == tnNone == true
		}() == true {
			__antithesis_instrumentation__.Notify(195350)
			incorrect = append(incorrect, "sslmode")
			fmt.Fprintln(&details, "Cannot use TLS client certificate authentication without a TLS transport.")
		} else {
			__antithesis_instrumentation__.Notify(195351)
		}
		__antithesis_instrumentation__.Notify(195345)
		if u.clientCertPath == "" {
			__antithesis_instrumentation__.Notify(195352)
			incorrect = append(incorrect, "sslcert")
			fmt.Fprintln(&details, "Client certificate missing.")
		} else {
			__antithesis_instrumentation__.Notify(195353)
		}
		__antithesis_instrumentation__.Notify(195346)
		if u.clientKeyPath == "" {
			__antithesis_instrumentation__.Notify(195354)
			incorrect = append(incorrect, "sslkey")
			fmt.Fprintln(&details, "Client key missing.")
		} else {
			__antithesis_instrumentation__.Notify(195355)
		}
	case authnUndefined:
		__antithesis_instrumentation__.Notify(195347)
		incorrect = append(incorrect, "authn")
		fmt.Fprintln(&details, "Authentication method unspecified.")
	default:
		__antithesis_instrumentation__.Notify(195348)
	}
	__antithesis_instrumentation__.Notify(195328)

	if len(incorrect) > 0 {
		__antithesis_instrumentation__.Notify(195356)
		return errors.WithDetail(errors.Newf("URL validation error: %s", redact.Join(", ", incorrect)), details.String())
	} else {
		__antithesis_instrumentation__.Notify(195357)
	}
	__antithesis_instrumentation__.Notify(195329)
	return nil
}
