package pgurl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"net"
	"net/url"
	"sort"
	"strings"
)

func (u *URL) ToDSN() string {
	__antithesis_instrumentation__.Notify(195100)
	escaper := strings.NewReplacer(` `, `\ `, `'`, `\'`, `\`, `\\`)
	var s strings.Builder

	accrue := func(key, val string) {
		__antithesis_instrumentation__.Notify(195112)
		s.WriteByte(' ')
		s.WriteString(key)
		s.WriteByte('=')
		s.WriteString(escaper.Replace(val))
	}
	__antithesis_instrumentation__.Notify(195101)
	if u.database != "" {
		__antithesis_instrumentation__.Notify(195113)
		accrue("database", u.database)
	} else {
		__antithesis_instrumentation__.Notify(195114)
	}
	__antithesis_instrumentation__.Notify(195102)
	if u.username != "" {
		__antithesis_instrumentation__.Notify(195115)
		accrue("user", u.username)
	} else {
		__antithesis_instrumentation__.Notify(195116)
	}
	__antithesis_instrumentation__.Notify(195103)
	switch u.net {
	case ProtoUnix, ProtoTCP:
		__antithesis_instrumentation__.Notify(195117)
		if u.host != "" {
			__antithesis_instrumentation__.Notify(195120)
			accrue("host", u.host)
		} else {
			__antithesis_instrumentation__.Notify(195121)
		}
		__antithesis_instrumentation__.Notify(195118)
		if u.port != "" {
			__antithesis_instrumentation__.Notify(195122)
			accrue("port", u.port)
		} else {
			__antithesis_instrumentation__.Notify(195123)
		}
	default:
		__antithesis_instrumentation__.Notify(195119)
	}
	__antithesis_instrumentation__.Notify(195104)
	switch u.authn {
	case authnPassword, authnPasswordWithClientCert:
		__antithesis_instrumentation__.Notify(195124)
		if u.hasPassword {
			__antithesis_instrumentation__.Notify(195126)
			accrue("password", u.password)
		} else {
			__antithesis_instrumentation__.Notify(195127)
		}
	default:
		__antithesis_instrumentation__.Notify(195125)
	}
	__antithesis_instrumentation__.Notify(195105)
	switch u.authn {
	case authnClientCert, authnPasswordWithClientCert:
		__antithesis_instrumentation__.Notify(195128)
		accrue("sslcert", u.clientCertPath)
		accrue("sslkey", u.clientKeyPath)
	default:
		__antithesis_instrumentation__.Notify(195129)
	}
	__antithesis_instrumentation__.Notify(195106)

	if u.caCertPath != "" {
		__antithesis_instrumentation__.Notify(195130)
		accrue("sslrootcert", u.caCertPath)
	} else {
		__antithesis_instrumentation__.Notify(195131)
	}
	__antithesis_instrumentation__.Notify(195107)

	if u.sec != tnUnspecified {
		__antithesis_instrumentation__.Notify(195132)
		accrue("sslmode", string(u.sec))
	} else {
		__antithesis_instrumentation__.Notify(195133)
	}
	__antithesis_instrumentation__.Notify(195108)

	keys := make([]string, len(u.extraOptions))
	for k := range u.extraOptions {
		__antithesis_instrumentation__.Notify(195134)
		keys = append(keys, k)
	}
	__antithesis_instrumentation__.Notify(195109)
	sort.Strings(keys)
	for _, k := range keys {
		__antithesis_instrumentation__.Notify(195135)
		v := u.extraOptions[k]
		for _, val := range v {
			__antithesis_instrumentation__.Notify(195136)
			accrue(k, val)
		}
	}
	__antithesis_instrumentation__.Notify(195110)

	out := s.String()
	if len(out) > 0 {
		__antithesis_instrumentation__.Notify(195137)

		out = out[1:]
	} else {
		__antithesis_instrumentation__.Notify(195138)
	}
	__antithesis_instrumentation__.Notify(195111)
	return out
}

func (u *URL) ToPQ() *url.URL {
	__antithesis_instrumentation__.Notify(195139)
	nu, opts := u.baseURL()

	if u.username != "" {
		__antithesis_instrumentation__.Notify(195142)
		nu.User = url.User(u.username)
	} else {
		__antithesis_instrumentation__.Notify(195143)
	}
	__antithesis_instrumentation__.Notify(195140)
	switch u.authn {
	case authnPassword, authnPasswordWithClientCert:
		__antithesis_instrumentation__.Notify(195144)
		if u.hasPassword {
			__antithesis_instrumentation__.Notify(195146)
			nu.User = url.UserPassword(u.username, u.password)
		} else {
			__antithesis_instrumentation__.Notify(195147)
		}
	default:
		__antithesis_instrumentation__.Notify(195145)
	}
	__antithesis_instrumentation__.Notify(195141)

	nu.RawQuery = opts.Encode()
	return nu
}

func (u *URL) String() string {
	__antithesis_instrumentation__.Notify(195148)
	return u.ToPQ().String()
}

func (u *URL) ToJDBC() *url.URL {
	__antithesis_instrumentation__.Notify(195149)
	nu, opts := u.baseURL()

	nu.Scheme = "jdbc:" + nu.Scheme

	if u.username != "" {
		__antithesis_instrumentation__.Notify(195152)
		opts.Set("user", u.username)
	} else {
		__antithesis_instrumentation__.Notify(195153)
	}
	__antithesis_instrumentation__.Notify(195150)
	switch u.authn {
	case authnPassword, authnPasswordWithClientCert:
		__antithesis_instrumentation__.Notify(195154)
		if u.hasPassword {
			__antithesis_instrumentation__.Notify(195156)
			opts.Set("password", u.password)
		} else {
			__antithesis_instrumentation__.Notify(195157)
		}
	default:
		__antithesis_instrumentation__.Notify(195155)
	}
	__antithesis_instrumentation__.Notify(195151)

	nu.RawQuery = opts.Encode()
	return nu
}

func (u *URL) baseURL() (*url.URL, url.Values) {
	__antithesis_instrumentation__.Notify(195158)
	nu := &url.URL{
		Scheme: Scheme,
		Path:   "/" + u.database,
	}
	opts := url.Values{}
	for k, v := range u.extraOptions {
		__antithesis_instrumentation__.Notify(195164)
		opts[k] = v
	}
	__antithesis_instrumentation__.Notify(195159)

	switch u.net {
	case ProtoTCP:
		__antithesis_instrumentation__.Notify(195165)
		nu.Host = u.host
		if u.port != "" {
			__antithesis_instrumentation__.Notify(195168)
			nu.Host = net.JoinHostPort(u.host, u.port)
		} else {
			__antithesis_instrumentation__.Notify(195169)
		}
	case ProtoUnix:
		__antithesis_instrumentation__.Notify(195166)

		opts.Set("host", u.host)
		if u.port != "" {
			__antithesis_instrumentation__.Notify(195170)
			opts.Set("port", u.port)
		} else {
			__antithesis_instrumentation__.Notify(195171)
		}
	default:
		__antithesis_instrumentation__.Notify(195167)
	}
	__antithesis_instrumentation__.Notify(195160)

	switch u.authn {
	case authnClientCert, authnPasswordWithClientCert:
		__antithesis_instrumentation__.Notify(195172)
		opts.Set("sslcert", u.clientCertPath)
		opts.Set("sslkey", u.clientKeyPath)
	default:
		__antithesis_instrumentation__.Notify(195173)
	}
	__antithesis_instrumentation__.Notify(195161)

	if u.caCertPath != "" {
		__antithesis_instrumentation__.Notify(195174)
		opts.Set("sslrootcert", u.caCertPath)
	} else {
		__antithesis_instrumentation__.Notify(195175)
	}
	__antithesis_instrumentation__.Notify(195162)

	if u.sec != tnUnspecified {
		__antithesis_instrumentation__.Notify(195176)
		opts.Set("sslmode", string(u.sec))
	} else {
		__antithesis_instrumentation__.Notify(195177)
	}
	__antithesis_instrumentation__.Notify(195163)

	return nu, opts
}
