package pgurl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"net/url"

	"github.com/cockroachdb/errors"
)

const Scheme = "postgresql"

type URL struct {
	username string
	database string

	net NetProtocol

	sec transportType

	authn authnType

	host, port string

	caCertPath string

	clientCertPath string
	clientKeyPath  string

	hasPassword bool
	password    string

	extraOptions url.Values
}

func New() *URL {
	__antithesis_instrumentation__.Notify(195258)
	return &URL{
		net:   ProtoTCP,
		sec:   tnUnspecified,
		authn: authnNone,
	}
}

func (u *URL) WithUsername(s string) *URL {
	__antithesis_instrumentation__.Notify(195259)
	u.username = s
	return u
}

func (u *URL) WithDefaultUsername(user string) *URL {
	__antithesis_instrumentation__.Notify(195260)
	if u.username == "" {
		__antithesis_instrumentation__.Notify(195262)
		u.username = user
	} else {
		__antithesis_instrumentation__.Notify(195263)
	}
	__antithesis_instrumentation__.Notify(195261)
	return u
}

func (u *URL) GetUsername() string {
	__antithesis_instrumentation__.Notify(195264)
	return u.username
}

func (u *URL) WithDatabase(s string) *URL {
	__antithesis_instrumentation__.Notify(195265)
	u.database = s
	return u
}

func (u *URL) WithDefaultDatabase(db string) *URL {
	__antithesis_instrumentation__.Notify(195266)
	if u.database == "" {
		__antithesis_instrumentation__.Notify(195268)
		u.database = db
	} else {
		__antithesis_instrumentation__.Notify(195269)
	}
	__antithesis_instrumentation__.Notify(195267)
	return u
}

func (u *URL) GetDatabase() string {
	__antithesis_instrumentation__.Notify(195270)
	return u.database
}

func (u *URL) AddOptions(opts url.Values) error {
	__antithesis_instrumentation__.Notify(195271)
	return u.parseOptions(opts)
}

func (u *URL) SetOption(key, value string) error {
	__antithesis_instrumentation__.Notify(195272)
	vals := []string{value}
	opts := url.Values{key: vals}
	if err := u.parseOptions(opts); err != nil {
		__antithesis_instrumentation__.Notify(195275)
		return err
	} else {
		__antithesis_instrumentation__.Notify(195276)
	}
	__antithesis_instrumentation__.Notify(195273)
	if _, ok := u.extraOptions[key]; ok {
		__antithesis_instrumentation__.Notify(195277)
		u.extraOptions[key] = vals
	} else {
		__antithesis_instrumentation__.Notify(195278)
	}
	__antithesis_instrumentation__.Notify(195274)
	return nil
}

func (u *URL) GetOption(opt string) string {
	__antithesis_instrumentation__.Notify(195279)
	return getVal(u.extraOptions, opt)
}

func (u *URL) GetExtraOptions() url.Values {
	__antithesis_instrumentation__.Notify(195280)
	return u.extraOptions
}

func (u *URL) WithInsecure() *URL {
	__antithesis_instrumentation__.Notify(195281)
	return u.
		WithAuthn(AuthnNone()).
		WithTransport(TransportNone())
}

type NetProtocol int

const (
	ProtoUndefined NetProtocol = iota

	ProtoUnix

	ProtoTCP
)

type NetOption func(u *URL) *URL

func (u *URL) WithNet(opt NetOption) *URL {
	__antithesis_instrumentation__.Notify(195282)
	return opt(u)
}

func NetTCP(host, port string) NetOption {
	__antithesis_instrumentation__.Notify(195283)
	return NetOption(func(u *URL) *URL {
		__antithesis_instrumentation__.Notify(195284)
		u.net = ProtoTCP
		u.host = host
		u.port = port
		return u
	})
}

func NetUnix(socketDir, port string) NetOption {
	__antithesis_instrumentation__.Notify(195285)
	return NetOption(func(u *URL) *URL {
		__antithesis_instrumentation__.Notify(195286)
		u.net = ProtoUnix
		u.host = socketDir
		u.port = port
		return u
	})
}

func (u *URL) WithDefaultHost(host string) *URL {
	__antithesis_instrumentation__.Notify(195287)
	if u.host == "" {
		__antithesis_instrumentation__.Notify(195289)
		u.host = host
	} else {
		__antithesis_instrumentation__.Notify(195290)
	}
	__antithesis_instrumentation__.Notify(195288)
	return u
}

func (u *URL) WithDefaultPort(port string) *URL {
	__antithesis_instrumentation__.Notify(195291)
	if u.port == "" {
		__antithesis_instrumentation__.Notify(195293)
		u.port = port
	} else {
		__antithesis_instrumentation__.Notify(195294)
	}
	__antithesis_instrumentation__.Notify(195292)
	return u
}

func (u *URL) GetNetworking() (NetProtocol, string, string) {
	__antithesis_instrumentation__.Notify(195295)
	return u.net, u.host, u.port
}

type authnType int

const (
	authnUndefined authnType = iota
	authnNone
	authnPassword
	authnClientCert
	authnPasswordWithClientCert
)

type AuthnOption func(u *URL) *URL

func (u *URL) WithAuthn(opt AuthnOption) *URL {
	__antithesis_instrumentation__.Notify(195296)
	return opt(u)
}

func (u *URL) GetAuthnOption() (AuthnOption, error) {
	__antithesis_instrumentation__.Notify(195297)
	switch u.authn {
	case authnNone:
		__antithesis_instrumentation__.Notify(195298)
		return AuthnNone(), nil
	case authnPassword:
		__antithesis_instrumentation__.Notify(195299)
		return AuthnPassword(u.hasPassword, u.password), nil
	case authnClientCert:
		__antithesis_instrumentation__.Notify(195300)
		return AuthnClientCert(u.clientCertPath, u.clientKeyPath), nil
	case authnPasswordWithClientCert:
		__antithesis_instrumentation__.Notify(195301)
		return AuthnPasswordAndCert(
			u.clientCertPath, u.clientKeyPath,
			u.hasPassword, u.password), nil
	default:
		__antithesis_instrumentation__.Notify(195302)
		return nil, errors.New("unable to retrieve authentication options")
	}
}

func AuthnNone() AuthnOption {
	__antithesis_instrumentation__.Notify(195303)
	return AuthnOption(func(u *URL) *URL {
		__antithesis_instrumentation__.Notify(195304)
		u.authn = authnNone
		return u
	})
}

func AuthnPassword(setPassword bool, password string) AuthnOption {
	__antithesis_instrumentation__.Notify(195305)
	return AuthnOption(func(u *URL) *URL {
		__antithesis_instrumentation__.Notify(195306)

		u.authn = authnPassword
		u.hasPassword = setPassword
		u.password = password
		return u
	})
}

func (u *URL) ClearPassword() {
	__antithesis_instrumentation__.Notify(195307)
	u.hasPassword = false
	u.password = ""
}

func (u *URL) GetAuthnPassword() (authnPwdEnabled bool, hasPassword bool, password string) {
	__antithesis_instrumentation__.Notify(195308)
	return u.authn == authnPassword || func() bool {
			__antithesis_instrumentation__.Notify(195309)
			return u.authn == authnPasswordWithClientCert == true
		}() == true,
		u.hasPassword,
		u.password
}

func AuthnClientCert(clientCertPath, clientKeyPath string) AuthnOption {
	__antithesis_instrumentation__.Notify(195310)
	return AuthnOption(func(u *URL) *URL {
		__antithesis_instrumentation__.Notify(195311)
		u.authn = authnClientCert
		u.clientCertPath = clientCertPath
		u.clientKeyPath = clientKeyPath
		return u
	})
}

func (u *URL) GetAuthnCert() (authnCertEnabled bool, clientCertPath, clientKeyPath string) {
	__antithesis_instrumentation__.Notify(195312)
	return u.authn == authnClientCert || func() bool {
			__antithesis_instrumentation__.Notify(195313)
			return u.authn == authnPasswordWithClientCert == true
		}() == true,
		u.clientCertPath,
		u.clientKeyPath
}

func AuthnPasswordAndCert(
	clientCertPath, clientKeyPath string, setPassword bool, password string,
) AuthnOption {
	__antithesis_instrumentation__.Notify(195314)
	p := AuthnPassword(setPassword, password)
	c := AuthnClientCert(clientCertPath, clientKeyPath)
	return AuthnOption(func(u *URL) *URL {
		__antithesis_instrumentation__.Notify(195315)
		u = c(p(u))
		u.authn = authnPasswordWithClientCert
		return u
	})
}

type transportType string

const (
	tnUnspecified   transportType = ""
	tnNone          transportType = "disable"
	tnTLSVerifyFull transportType = "verify-full"
	tnTLSVerifyCA   transportType = "verify-ca"
	tnTLSRequire    transportType = "require"
	tnTLSPrefer     transportType = "prefer"
	tnTLSAllow      transportType = "allow"
)

type TLSMode transportType

const (
	TLSVerifyFull TLSMode = TLSMode(tnTLSVerifyFull)

	TLSVerifyCA TLSMode = TLSMode(tnTLSVerifyCA)

	TLSRequire TLSMode = TLSMode(tnTLSRequire)

	TLSPrefer TLSMode = TLSMode(tnTLSPrefer)

	TLSAllow TLSMode = TLSMode(tnTLSAllow)

	TLSUnspecified TLSMode = TLSMode(tnUnspecified)
)

type TransportOption func(u *URL) *URL

func (u *URL) WithTransport(opt TransportOption) *URL {
	__antithesis_instrumentation__.Notify(195316)
	return opt(u)
}

func TransportTLS(mode TLSMode, caCertPath string) TransportOption {
	__antithesis_instrumentation__.Notify(195317)
	return TransportOption(func(u *URL) *URL {
		__antithesis_instrumentation__.Notify(195318)
		u.sec = transportType(mode)
		u.caCertPath = caCertPath
		return u
	})
}

func TransportNone() TransportOption {
	__antithesis_instrumentation__.Notify(195319)
	return TransportOption(func(u *URL) *URL {
		__antithesis_instrumentation__.Notify(195320)
		u.sec = tnNone
		return u
	})
}

func (u *URL) GetTLSOptions() (tlsUsed bool, mode TLSMode, caCertPath string) {
	__antithesis_instrumentation__.Notify(195321)
	if u.sec == tnNone {
		__antithesis_instrumentation__.Notify(195323)
		return false, "", ""
	} else {
		__antithesis_instrumentation__.Notify(195324)
	}
	__antithesis_instrumentation__.Notify(195322)
	return true, TLSMode(u.sec), u.caCertPath
}
