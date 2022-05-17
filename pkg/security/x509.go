package security

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

const (
	validFrom     = -time.Hour * 24
	maxPathLength = 1
	caCommonName  = "Cockroach CA"

	TenantsOU = "Tenants"
)

func newTemplate(
	commonName string, lifetime time.Duration, orgUnits ...string,
) (*x509.Certificate, error) {
	__antithesis_instrumentation__.Notify(187421)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		__antithesis_instrumentation__.Notify(187423)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187424)
	}
	__antithesis_instrumentation__.Notify(187422)

	now := timeutil.Now()
	notBefore := now.Add(validFrom)
	notAfter := now.Add(lifetime)

	cert := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization:       []string{"Cockroach"},
			OrganizationalUnit: orgUnits,
			CommonName:         commonName,
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage: x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
	}

	return cert, nil
}

func GenerateCA(signer crypto.Signer, lifetime time.Duration) ([]byte, error) {
	__antithesis_instrumentation__.Notify(187425)
	template, err := newTemplate(caCommonName, lifetime)
	if err != nil {
		__antithesis_instrumentation__.Notify(187428)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187429)
	}
	__antithesis_instrumentation__.Notify(187426)

	template.BasicConstraintsValid = true
	template.IsCA = true
	template.MaxPathLen = maxPathLength
	template.KeyUsage |= x509.KeyUsageCertSign
	template.KeyUsage |= x509.KeyUsageContentCommitment

	certBytes, err := x509.CreateCertificate(
		rand.Reader,
		template,
		template,
		signer.Public(),
		signer)
	if err != nil {
		__antithesis_instrumentation__.Notify(187430)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187431)
	}
	__antithesis_instrumentation__.Notify(187427)

	return certBytes, nil
}

func checkLifetimeAgainstCA(cert, ca *x509.Certificate) error {
	__antithesis_instrumentation__.Notify(187432)
	if ca.NotAfter.After(cert.NotAfter) || func() bool {
		__antithesis_instrumentation__.Notify(187434)
		return ca.NotAfter.Equal(cert.NotAfter) == true
	}() == true {
		__antithesis_instrumentation__.Notify(187435)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(187436)
	}
	__antithesis_instrumentation__.Notify(187433)

	now := timeutil.Now()

	niceCALifetime := ca.NotAfter.Sub(now).Hours()
	niceCertLifetime := cert.NotAfter.Sub(now).Hours()
	return errors.Errorf("CA lifetime is %fh, shorter than the requested %fh. "+
		"Renew CA certificate, or rerun with --lifetime=%dh for a shorter duration.",
		niceCALifetime, niceCertLifetime, int64(niceCALifetime))
}

func GenerateServerCert(
	caCert *x509.Certificate,
	caPrivateKey crypto.PrivateKey,
	nodePublicKey crypto.PublicKey,
	lifetime time.Duration,
	user SQLUsername,
	hosts []string,
) ([]byte, error) {
	__antithesis_instrumentation__.Notify(187437)

	template, err := newTemplate(user.Normalized(), lifetime)
	if err != nil {
		__antithesis_instrumentation__.Notify(187441)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187442)
	}
	__antithesis_instrumentation__.Notify(187438)

	if err := checkLifetimeAgainstCA(template, caCert); err != nil {
		__antithesis_instrumentation__.Notify(187443)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187444)
	}
	__antithesis_instrumentation__.Notify(187439)

	template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}
	addHostsToTemplate(template, hosts)

	certBytes, err := x509.CreateCertificate(rand.Reader, template, caCert, nodePublicKey, caPrivateKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(187445)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187446)
	}
	__antithesis_instrumentation__.Notify(187440)

	return certBytes, nil
}

func addHostsToTemplate(template *x509.Certificate, hosts []string) {
	__antithesis_instrumentation__.Notify(187447)
	for _, h := range hosts {
		__antithesis_instrumentation__.Notify(187448)
		if ip := net.ParseIP(h); ip != nil {
			__antithesis_instrumentation__.Notify(187449)
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			__antithesis_instrumentation__.Notify(187450)
			template.DNSNames = append(template.DNSNames, h)
		}
	}
}

func GenerateUIServerCert(
	caCert *x509.Certificate,
	caPrivateKey crypto.PrivateKey,
	certPublicKey crypto.PublicKey,
	lifetime time.Duration,
	hosts []string,
) ([]byte, error) {
	__antithesis_instrumentation__.Notify(187451)

	template, err := newTemplate(hosts[0], lifetime)
	if err != nil {
		__antithesis_instrumentation__.Notify(187455)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187456)
	}
	__antithesis_instrumentation__.Notify(187452)

	if err := checkLifetimeAgainstCA(template, caCert); err != nil {
		__antithesis_instrumentation__.Notify(187457)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187458)
	}
	__antithesis_instrumentation__.Notify(187453)

	template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	addHostsToTemplate(template, hosts)

	certBytes, err := x509.CreateCertificate(rand.Reader, template, caCert, certPublicKey, caPrivateKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(187459)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187460)
	}
	__antithesis_instrumentation__.Notify(187454)

	return certBytes, nil
}

func GenerateTenantCert(
	caCert *x509.Certificate,
	caPrivateKey crypto.PrivateKey,
	clientPublicKey crypto.PublicKey,
	lifetime time.Duration,
	tenantID uint64,
	hosts []string,
) ([]byte, error) {
	__antithesis_instrumentation__.Notify(187461)

	if tenantID == 0 {
		__antithesis_instrumentation__.Notify(187466)
		return nil, errors.Errorf("tenantId %d is invalid (requires != 0)", tenantID)
	} else {
		__antithesis_instrumentation__.Notify(187467)
	}
	__antithesis_instrumentation__.Notify(187462)

	template, err := newTemplate(fmt.Sprintf("%d", tenantID), lifetime, TenantsOU)
	if err != nil {
		__antithesis_instrumentation__.Notify(187468)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187469)
	}
	__antithesis_instrumentation__.Notify(187463)

	if err := checkLifetimeAgainstCA(template, caCert); err != nil {
		__antithesis_instrumentation__.Notify(187470)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187471)
	}
	__antithesis_instrumentation__.Notify(187464)

	template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}
	addHostsToTemplate(template, hosts)

	certBytes, err := x509.CreateCertificate(rand.Reader, template, caCert, clientPublicKey, caPrivateKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(187472)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187473)
	}
	__antithesis_instrumentation__.Notify(187465)

	return certBytes, nil
}

func GenerateClientCert(
	caCert *x509.Certificate,
	caPrivateKey crypto.PrivateKey,
	clientPublicKey crypto.PublicKey,
	lifetime time.Duration,
	user SQLUsername,
) ([]byte, error) {
	__antithesis_instrumentation__.Notify(187474)

	if user.Undefined() {
		__antithesis_instrumentation__.Notify(187479)
		return nil, errors.Errorf("user cannot be empty")
	} else {
		__antithesis_instrumentation__.Notify(187480)
	}
	__antithesis_instrumentation__.Notify(187475)

	template, err := newTemplate(user.Normalized(), lifetime)
	if err != nil {
		__antithesis_instrumentation__.Notify(187481)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187482)
	}
	__antithesis_instrumentation__.Notify(187476)

	if err := checkLifetimeAgainstCA(template, caCert); err != nil {
		__antithesis_instrumentation__.Notify(187483)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187484)
	}
	__antithesis_instrumentation__.Notify(187477)

	template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}

	certBytes, err := x509.CreateCertificate(rand.Reader, template, caCert, clientPublicKey, caPrivateKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(187485)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187486)
	}
	__antithesis_instrumentation__.Notify(187478)

	return certBytes, nil
}

func GenerateTenantSigningCert(
	publicKey crypto.PublicKey, privateKey crypto.PrivateKey, lifetime time.Duration, tenantID uint64,
) ([]byte, error) {
	__antithesis_instrumentation__.Notify(187487)
	now := timeutil.Now()
	template := &x509.Certificate{
		Subject: pkix.Name{
			CommonName: fmt.Sprintf("Tenant %d Token Signing Certificate", tenantID),
		},
		SerialNumber:          big.NewInt(1),
		BasicConstraintsValid: true,
		IsCA:                  false,
		PublicKey:             publicKey,
		NotBefore:             now.Add(validFrom),
		NotAfter:              now.Add(lifetime),
		KeyUsage:              x509.KeyUsageDigitalSignature,
	}

	certBytes, err := x509.CreateCertificate(
		rand.Reader,
		template,
		template,
		publicKey,
		privateKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(187489)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187490)
	}
	__antithesis_instrumentation__.Notify(187488)

	return certBytes, nil
}
