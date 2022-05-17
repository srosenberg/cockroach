package security

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

const defaultKeySize = 2048

const notBeforeMargin = time.Hour * 24

func createCertificateSerialNumber() (serialNumber *big.Int, err error) {
	__antithesis_instrumentation__.Notify(185768)
	max := new(big.Int)
	max.Exp(big.NewInt(2), big.NewInt(130), nil).Sub(max, big.NewInt(1))

	serialNumber, err = rand.Int(rand.Reader, max)
	if err != nil {
		__antithesis_instrumentation__.Notify(185770)
		err = errors.Wrap(err, "failed to create new serial number")
	} else {
		__antithesis_instrumentation__.Notify(185771)
	}
	__antithesis_instrumentation__.Notify(185769)

	serialNumber.Add(serialNumber, big.NewInt(1))

	return
}

type LoggerFn = func(ctx context.Context, format string, args ...interface{})

func describeCert(cert *x509.Certificate) redact.RedactableString {
	__antithesis_instrumentation__.Notify(185772)
	var buf redact.StringBuilder
	buf.SafeString("{\n")
	buf.Printf("  SN: %s,\n", cert.SerialNumber)
	buf.Printf("  CA: %v,\n", cert.IsCA)
	buf.Printf("  Issuer: %q,\n", cert.Issuer)
	buf.Printf("  Subject: %q,\n", cert.Subject)
	buf.Printf("  NotBefore: %s,\n", cert.NotBefore)
	buf.Printf("  NotAfter: %s", cert.NotAfter)
	buf.Printf(" (Validity: %s),\n", cert.NotAfter.Sub(timeutil.Now()))
	if !cert.IsCA {
		__antithesis_instrumentation__.Notify(185774)
		buf.Printf("  DNS: %v,\n", cert.DNSNames)
		buf.Printf("  IP: %v\n", cert.IPAddresses)
	} else {
		__antithesis_instrumentation__.Notify(185775)
	}
	__antithesis_instrumentation__.Notify(185773)
	buf.SafeString("}")
	return buf.RedactableString()
}

const (
	crlOrg      = "Cockroach"
	crlIssuerOU = "automatic cert generator"
)

func CreateCACertAndKey(
	ctx context.Context, loggerFn LoggerFn, lifespan time.Duration, service string,
) (certPEM, keyPEM *pem.Block, err error) {
	__antithesis_instrumentation__.Notify(185776)
	notBefore := timeutil.Now().Add(-notBeforeMargin)
	notAfter := timeutil.Now().Add(lifespan)

	serialNumber, err := createCertificateSerialNumber()
	if err != nil {
		__antithesis_instrumentation__.Notify(185783)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(185784)
	}
	__antithesis_instrumentation__.Notify(185777)

	ca := &x509.Certificate{
		SerialNumber: serialNumber,
		Issuer: pkix.Name{
			Organization: []string{crlOrg},
			CommonName:   service,
		},
		Subject: pkix.Name{
			Organization: []string{crlOrg},
			CommonName:   service,
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageContentCommitment,
		BasicConstraintsValid: true,
		MaxPathLen:            1,
	}
	if loggerFn != nil {
		__antithesis_instrumentation__.Notify(185785)
		loggerFn(ctx, "creating CA cert from template: %s", describeCert(ca))
	} else {
		__antithesis_instrumentation__.Notify(185786)
	}
	__antithesis_instrumentation__.Notify(185778)

	caPrivKey, err := rsa.GenerateKey(rand.Reader, defaultKeySize)
	if err != nil {
		__antithesis_instrumentation__.Notify(185787)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(185788)
	}
	__antithesis_instrumentation__.Notify(185779)

	caPrivKeyPEM, err := PrivateKeyToPEM(caPrivKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(185789)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(185790)
	}
	__antithesis_instrumentation__.Notify(185780)

	if loggerFn != nil {
		__antithesis_instrumentation__.Notify(185791)
		loggerFn(ctx, "signing CA cert")
	} else {
		__antithesis_instrumentation__.Notify(185792)
	}
	__antithesis_instrumentation__.Notify(185781)

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(185793)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(185794)
	}
	__antithesis_instrumentation__.Notify(185782)

	caPEM := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	}

	return caPEM, caPrivKeyPEM, nil
}

func CreateServiceCertAndKey(
	ctx context.Context,
	loggerFn LoggerFn,
	lifespan time.Duration,
	commonName string,
	hostnames []string,
	caCertBlock, caKeyBlock *pem.Block,
	serviceCertIsAlsoValidAsClient bool,
) (certPEM *pem.Block, keyPEM *pem.Block, err error) {
	__antithesis_instrumentation__.Notify(185795)
	notBefore := timeutil.Now().Add(-notBeforeMargin)
	notAfter := timeutil.Now().Add(lifespan)

	serialNumber, err := createCertificateSerialNumber()
	if err != nil {
		__antithesis_instrumentation__.Notify(185806)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(185807)
	}
	__antithesis_instrumentation__.Notify(185796)

	caCert, err := x509.ParseCertificate(caCertBlock.Bytes)
	if err != nil {
		__antithesis_instrumentation__.Notify(185808)
		err = errors.Wrap(err, "failed to parse valid Certificate from PEM blob")
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(185809)
	}
	__antithesis_instrumentation__.Notify(185797)

	caKey, err := x509.ParsePKCS1PrivateKey(caKeyBlock.Bytes)
	if err != nil {
		__antithesis_instrumentation__.Notify(185810)
		err = errors.Wrap(err, "failed to parse valid Private Key from PEM blob")
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(185811)
	}
	__antithesis_instrumentation__.Notify(185798)

	serviceCert := &x509.Certificate{
		SerialNumber: serialNumber,
		Issuer: pkix.Name{
			Organization:       []string{crlOrg},
			OrganizationalUnit: []string{crlIssuerOU},
			CommonName:         caCommonName,
		},
		Subject: pkix.Name{
			Organization: []string{crlOrg},
			CommonName:   commonName,
		},
		NotBefore:   notBefore,
		NotAfter:    notAfter,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
	}

	if serviceCertIsAlsoValidAsClient {
		__antithesis_instrumentation__.Notify(185812)
		serviceCert.ExtKeyUsage = append(serviceCert.ExtKeyUsage, x509.ExtKeyUsageClientAuth)
	} else {
		__antithesis_instrumentation__.Notify(185813)
	}
	__antithesis_instrumentation__.Notify(185799)

	for _, hostname := range hostnames {
		__antithesis_instrumentation__.Notify(185814)
		ip := net.ParseIP(hostname)
		if ip != nil {
			__antithesis_instrumentation__.Notify(185815)
			serviceCert.IPAddresses = []net.IP{ip}
		} else {
			__antithesis_instrumentation__.Notify(185816)
			serviceCert.DNSNames = []string{hostname}
		}
	}
	__antithesis_instrumentation__.Notify(185800)

	if loggerFn != nil {
		__antithesis_instrumentation__.Notify(185817)
		loggerFn(ctx, "creating service cert from template: %s", describeCert(serviceCert))
	} else {
		__antithesis_instrumentation__.Notify(185818)
	}
	__antithesis_instrumentation__.Notify(185801)

	servicePrivKey, err := rsa.GenerateKey(rand.Reader, defaultKeySize)
	if err != nil {
		__antithesis_instrumentation__.Notify(185819)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(185820)
	}
	__antithesis_instrumentation__.Notify(185802)

	if loggerFn != nil {
		__antithesis_instrumentation__.Notify(185821)
		loggerFn(ctx, "signing service cert")
	} else {
		__antithesis_instrumentation__.Notify(185822)
	}
	__antithesis_instrumentation__.Notify(185803)
	serviceCertBytes, err := x509.CreateCertificate(rand.Reader, serviceCert, caCert, &servicePrivKey.PublicKey, caKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(185823)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(185824)
	}
	__antithesis_instrumentation__.Notify(185804)

	serviceCertBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serviceCertBytes,
	}

	servicePrivKeyPEM, err := PrivateKeyToPEM(servicePrivKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(185825)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(185826)
	}
	__antithesis_instrumentation__.Notify(185805)

	return serviceCertBlock, servicePrivKeyPEM, nil
}
