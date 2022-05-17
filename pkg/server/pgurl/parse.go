package pgurl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"net/url"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/util/netutil/addr"
	"github.com/cockroachdb/errors"
)

func Parse(s string) (*URL, error) {
	__antithesis_instrumentation__.Notify(195178)
	u, err := url.Parse(s)
	if err != nil {
		__antithesis_instrumentation__.Notify(195185)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(195186)
	}
	__antithesis_instrumentation__.Notify(195179)

	if u.Opaque != "" {
		__antithesis_instrumentation__.Notify(195187)
		return nil, errors.Newf("unknown URL format: %s", s)
	} else {
		__antithesis_instrumentation__.Notify(195188)
	}
	__antithesis_instrumentation__.Notify(195180)

	dst := New()

	sc := strings.TrimPrefix(u.Scheme, "jdbc:")
	if sc != "postgres" && func() bool {
		__antithesis_instrumentation__.Notify(195189)
		return sc != "postgresql" == true
	}() == true {
		__antithesis_instrumentation__.Notify(195190)
		return nil, errors.Newf("unrecognized URL scheme: %s", u.Scheme)
	} else {
		__antithesis_instrumentation__.Notify(195191)
	}
	__antithesis_instrumentation__.Notify(195181)

	if u.User != nil {
		__antithesis_instrumentation__.Notify(195192)
		dst.username = u.User.Username()
		dst.password, dst.hasPassword = u.User.Password()
	} else {
		__antithesis_instrumentation__.Notify(195193)
	}
	__antithesis_instrumentation__.Notify(195182)

	if u.Host != "" {
		__antithesis_instrumentation__.Notify(195194)
		host, port, err := addr.SplitHostPort(u.Host, "0")
		if err != nil {
			__antithesis_instrumentation__.Notify(195196)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(195197)
		}
		__antithesis_instrumentation__.Notify(195195)
		dst.host = host
		if port != "0" {
			__antithesis_instrumentation__.Notify(195198)
			dst.port = port
		} else {
			__antithesis_instrumentation__.Notify(195199)
		}
	} else {
		__antithesis_instrumentation__.Notify(195200)
	}
	__antithesis_instrumentation__.Notify(195183)

	if u.Path != "" {
		__antithesis_instrumentation__.Notify(195201)

		dst.database = u.Path[1:]
	} else {
		__antithesis_instrumentation__.Notify(195202)
	}
	__antithesis_instrumentation__.Notify(195184)

	q := u.Query()

	err = dst.parseOptions(q)
	return dst, err
}

func (u *URL) parseOptions(extra url.Values) error {
	__antithesis_instrumentation__.Notify(195203)
	q := u.extraOptions
	if q == nil {
		__antithesis_instrumentation__.Notify(195216)

		q = make(url.Values)
		for k, v := range extra {
			__antithesis_instrumentation__.Notify(195217)
			q[k] = v
		}
	} else {
		__antithesis_instrumentation__.Notify(195218)

		for u, vs := range extra {
			__antithesis_instrumentation__.Notify(195219)
			q[u] = append(q[u], vs...)
		}
	}
	__antithesis_instrumentation__.Notify(195204)

	if _, hasDbOpt := q["database"]; hasDbOpt {
		__antithesis_instrumentation__.Notify(195220)
		u.database = getVal(q, "database")
		delete(q, "database")
	} else {
		__antithesis_instrumentation__.Notify(195221)
	}
	__antithesis_instrumentation__.Notify(195205)
	if _, hasUserOpt := q["user"]; hasUserOpt {
		__antithesis_instrumentation__.Notify(195222)
		u.username = getVal(q, "user")
		delete(q, "user")
	} else {
		__antithesis_instrumentation__.Notify(195223)
	}
	__antithesis_instrumentation__.Notify(195206)
	if _, hasHostOpt := q["host"]; hasHostOpt {
		__antithesis_instrumentation__.Notify(195224)
		u.host = getVal(q, "host")
		delete(q, "host")
	} else {
		__antithesis_instrumentation__.Notify(195225)
	}
	__antithesis_instrumentation__.Notify(195207)
	if _, hasPortOpt := q["port"]; hasPortOpt {
		__antithesis_instrumentation__.Notify(195226)
		u.port = getVal(q, "port")
		delete(q, "port")
	} else {
		__antithesis_instrumentation__.Notify(195227)
	}
	__antithesis_instrumentation__.Notify(195208)
	_, hasPasswordOpt := q["password"]
	if hasPasswordOpt {
		__antithesis_instrumentation__.Notify(195228)
		u.hasPassword = true
		u.password = getVal(q, "password")
		delete(q, "password")
	} else {
		__antithesis_instrumentation__.Notify(195229)
	}
	__antithesis_instrumentation__.Notify(195209)
	_, hasRootCert := q["sslrootcert"]
	if hasRootCert {
		__antithesis_instrumentation__.Notify(195230)
		u.caCertPath = getVal(q, "sslrootcert")
		delete(q, "sslrootcert")
	} else {
		__antithesis_instrumentation__.Notify(195231)
	}
	__antithesis_instrumentation__.Notify(195210)
	_, hasClientCert := q["sslcert"]
	if hasClientCert {
		__antithesis_instrumentation__.Notify(195232)
		u.clientCertPath = getVal(q, "sslcert")
		delete(q, "sslcert")
	} else {
		__antithesis_instrumentation__.Notify(195233)
	}
	__antithesis_instrumentation__.Notify(195211)
	_, hasClientKey := q["sslkey"]
	if hasClientKey {
		__antithesis_instrumentation__.Notify(195234)
		u.clientKeyPath = getVal(q, "sslkey")
		delete(q, "sslkey")
	} else {
		__antithesis_instrumentation__.Notify(195235)
	}
	__antithesis_instrumentation__.Notify(195212)

	if strings.HasPrefix(u.host, "/") {
		__antithesis_instrumentation__.Notify(195236)
		u.net = ProtoUnix
	} else {
		__antithesis_instrumentation__.Notify(195237)
		u.net = ProtoTCP
	}
	__antithesis_instrumentation__.Notify(195213)

	useCerts := hasClientCert || func() bool {
		__antithesis_instrumentation__.Notify(195238)
		return hasClientKey == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(195239)
		return u.clientCertPath != "" == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(195240)
		return u.clientKeyPath != "" == true
	}() == true
	if u.hasPassword && func() bool {
		__antithesis_instrumentation__.Notify(195241)
		return !useCerts == true
	}() == true {
		__antithesis_instrumentation__.Notify(195242)
		u.authn = authnPassword
	} else {
		__antithesis_instrumentation__.Notify(195243)
		if u.hasPassword && func() bool {
			__antithesis_instrumentation__.Notify(195244)
			return useCerts == true
		}() == true {
			__antithesis_instrumentation__.Notify(195245)
			u.authn = authnPasswordWithClientCert
		} else {
			__antithesis_instrumentation__.Notify(195246)
			if !u.hasPassword && func() bool {
				__antithesis_instrumentation__.Notify(195247)
				return useCerts == true
			}() == true {
				__antithesis_instrumentation__.Notify(195248)
				u.authn = authnClientCert
			} else {
				__antithesis_instrumentation__.Notify(195249)
				u.authn = authnNone
			}
		}
	}
	__antithesis_instrumentation__.Notify(195214)

	if _, hasSSLMode := q["sslmode"]; hasSSLMode {
		__antithesis_instrumentation__.Notify(195250)
		sslMode := getVal(q, "sslmode")
		delete(q, "sslmode")
		switch sslMode {
		case string(tnNone),
			string(tnTLSVerifyFull),
			string(tnTLSVerifyCA),
			string(tnTLSRequire),
			string(tnTLSPrefer),
			string(tnTLSAllow):
			__antithesis_instrumentation__.Notify(195251)
			u.sec = transportType(sslMode)
		default:
			__antithesis_instrumentation__.Notify(195252)
			return errors.Newf("unrecognized sslmode parameter in URL: %s", sslMode)
		}
	} else {
		__antithesis_instrumentation__.Notify(195253)
	}
	__antithesis_instrumentation__.Notify(195215)

	u.extraOptions = q

	return nil
}

func getVal(q url.Values, key string) string {
	__antithesis_instrumentation__.Notify(195254)
	v := q[key]
	if len(v) == 0 {
		__antithesis_instrumentation__.Notify(195256)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(195257)
	}
	__antithesis_instrumentation__.Notify(195255)
	return v[len(v)-1]
}
