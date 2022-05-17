package security

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
)

var certPrincipalMap struct {
	syncutil.RWMutex
	m map[string]string
}

type UserAuthHook func(
	ctx context.Context,
	systemIdentity SQLUsername,
	clientConnection bool,
) error

func SetCertPrincipalMap(mappings []string) error {
	__antithesis_instrumentation__.Notify(185690)
	m := make(map[string]string, len(mappings))
	for _, v := range mappings {
		__antithesis_instrumentation__.Notify(185692)
		if v == "" {
			__antithesis_instrumentation__.Notify(185695)
			continue
		} else {
			__antithesis_instrumentation__.Notify(185696)
		}
		__antithesis_instrumentation__.Notify(185693)
		idx := strings.LastIndexByte(v, ':')
		if idx == -1 {
			__antithesis_instrumentation__.Notify(185697)
			return errors.Errorf("invalid <cert-principal>:<db-principal> mapping: %q", v)
		} else {
			__antithesis_instrumentation__.Notify(185698)
		}
		__antithesis_instrumentation__.Notify(185694)
		m[v[:idx]] = v[idx+1:]
	}
	__antithesis_instrumentation__.Notify(185691)
	certPrincipalMap.Lock()
	certPrincipalMap.m = m
	certPrincipalMap.Unlock()
	return nil
}

func transformPrincipal(commonName string) string {
	__antithesis_instrumentation__.Notify(185699)
	certPrincipalMap.RLock()
	mappedName, ok := certPrincipalMap.m[commonName]
	certPrincipalMap.RUnlock()
	if !ok {
		__antithesis_instrumentation__.Notify(185701)
		return commonName
	} else {
		__antithesis_instrumentation__.Notify(185702)
	}
	__antithesis_instrumentation__.Notify(185700)
	return mappedName
}

func getCertificatePrincipals(cert *x509.Certificate) []string {
	__antithesis_instrumentation__.Notify(185703)
	results := make([]string, 0, 1+len(cert.DNSNames))
	results = append(results, transformPrincipal(cert.Subject.CommonName))
	for _, name := range cert.DNSNames {
		__antithesis_instrumentation__.Notify(185705)
		results = append(results, transformPrincipal(name))
	}
	__antithesis_instrumentation__.Notify(185704)
	return results
}

func GetCertificateUsers(tlsState *tls.ConnectionState) ([]string, error) {
	__antithesis_instrumentation__.Notify(185706)
	if tlsState == nil {
		__antithesis_instrumentation__.Notify(185709)
		return nil, errors.Errorf("request is not using TLS")
	} else {
		__antithesis_instrumentation__.Notify(185710)
	}
	__antithesis_instrumentation__.Notify(185707)
	if len(tlsState.PeerCertificates) == 0 {
		__antithesis_instrumentation__.Notify(185711)
		return nil, errors.Errorf("no client certificates in request")
	} else {
		__antithesis_instrumentation__.Notify(185712)
	}
	__antithesis_instrumentation__.Notify(185708)

	peerCert := tlsState.PeerCertificates[0]
	return getCertificatePrincipals(peerCert), nil
}

func Contains(sl []string, s string) bool {
	__antithesis_instrumentation__.Notify(185713)
	for i := range sl {
		__antithesis_instrumentation__.Notify(185715)
		if sl[i] == s {
			__antithesis_instrumentation__.Notify(185716)
			return true
		} else {
			__antithesis_instrumentation__.Notify(185717)
		}
	}
	__antithesis_instrumentation__.Notify(185714)
	return false
}

func UserAuthCertHook(insecureMode bool, tlsState *tls.ConnectionState) (UserAuthHook, error) {
	__antithesis_instrumentation__.Notify(185718)
	var certUsers []string

	if !insecureMode {
		__antithesis_instrumentation__.Notify(185720)
		var err error
		certUsers, err = GetCertificateUsers(tlsState)
		if err != nil {
			__antithesis_instrumentation__.Notify(185721)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(185722)
		}
	} else {
		__antithesis_instrumentation__.Notify(185723)
	}
	__antithesis_instrumentation__.Notify(185719)

	return func(ctx context.Context, systemIdentity SQLUsername, clientConnection bool) error {
		__antithesis_instrumentation__.Notify(185724)

		if systemIdentity.Undefined() {
			__antithesis_instrumentation__.Notify(185730)
			return errors.New("user is missing")
		} else {
			__antithesis_instrumentation__.Notify(185731)
		}
		__antithesis_instrumentation__.Notify(185725)

		if !clientConnection && func() bool {
			__antithesis_instrumentation__.Notify(185732)
			return !systemIdentity.IsNodeUser() == true
		}() == true {
			__antithesis_instrumentation__.Notify(185733)
			return errors.Errorf("user %s is not allowed", systemIdentity)
		} else {
			__antithesis_instrumentation__.Notify(185734)
		}
		__antithesis_instrumentation__.Notify(185726)

		if insecureMode {
			__antithesis_instrumentation__.Notify(185735)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(185736)
		}
		__antithesis_instrumentation__.Notify(185727)

		if IsTenantCertificate(tlsState.PeerCertificates[0]) {
			__antithesis_instrumentation__.Notify(185737)
			return errors.Errorf("using tenant client certificate as user certificate is not allowed")
		} else {
			__antithesis_instrumentation__.Notify(185738)
		}
		__antithesis_instrumentation__.Notify(185728)

		if !Contains(certUsers, systemIdentity.Normalized()) {
			__antithesis_instrumentation__.Notify(185739)
			return errors.Errorf("requested user is %s, but certificate is for %s", systemIdentity, certUsers)
		} else {
			__antithesis_instrumentation__.Notify(185740)
		}
		__antithesis_instrumentation__.Notify(185729)

		return nil
	}, nil
}

func IsTenantCertificate(cert *x509.Certificate) bool {
	__antithesis_instrumentation__.Notify(185741)
	return Contains(cert.Subject.OrganizationalUnit, TenantsOU)
}

func UserAuthPasswordHook(
	insecureMode bool, password string, hashedPassword PasswordHash,
) UserAuthHook {
	__antithesis_instrumentation__.Notify(185742)
	return func(ctx context.Context, systemIdentity SQLUsername, clientConnection bool) error {
		__antithesis_instrumentation__.Notify(185743)
		if systemIdentity.Undefined() {
			__antithesis_instrumentation__.Notify(185750)
			return errors.New("user is missing")
		} else {
			__antithesis_instrumentation__.Notify(185751)
		}
		__antithesis_instrumentation__.Notify(185744)

		if !clientConnection {
			__antithesis_instrumentation__.Notify(185752)
			return errors.New("password authentication is only available for client connections")
		} else {
			__antithesis_instrumentation__.Notify(185753)
		}
		__antithesis_instrumentation__.Notify(185745)

		if insecureMode {
			__antithesis_instrumentation__.Notify(185754)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(185755)
		}
		__antithesis_instrumentation__.Notify(185746)

		if len(password) == 0 {
			__antithesis_instrumentation__.Notify(185756)
			return NewErrPasswordUserAuthFailed(systemIdentity)
		} else {
			__antithesis_instrumentation__.Notify(185757)
		}
		__antithesis_instrumentation__.Notify(185747)
		ok, err := CompareHashAndCleartextPassword(ctx, hashedPassword, password)
		if err != nil {
			__antithesis_instrumentation__.Notify(185758)
			return err
		} else {
			__antithesis_instrumentation__.Notify(185759)
		}
		__antithesis_instrumentation__.Notify(185748)
		if !ok {
			__antithesis_instrumentation__.Notify(185760)
			return NewErrPasswordUserAuthFailed(systemIdentity)
		} else {
			__antithesis_instrumentation__.Notify(185761)
		}
		__antithesis_instrumentation__.Notify(185749)

		return nil
	}
}

func NewErrPasswordUserAuthFailed(username SQLUsername) error {
	__antithesis_instrumentation__.Notify(185762)
	return &PasswordUserAuthError{errors.Newf("password authentication failed for user %s", username)}
}

type PasswordUserAuthError struct {
	err error
}

func (i *PasswordUserAuthError) Error() string {
	__antithesis_instrumentation__.Notify(185763)
	return i.err.Error()
}

func (i *PasswordUserAuthError) Cause() error {
	__antithesis_instrumentation__.Notify(185764)
	return i.err
}

func (i *PasswordUserAuthError) Unwrap() error {
	__antithesis_instrumentation__.Notify(185765)
	return i.err
}

func (i *PasswordUserAuthError) Format(s fmt.State, verb rune) {
	__antithesis_instrumentation__.Notify(185766)
	errors.FormatError(i, s, verb)
}

func (i *PasswordUserAuthError) FormatError(p errors.Printer) error {
	__antithesis_instrumentation__.Notify(185767)
	return i.err
}
