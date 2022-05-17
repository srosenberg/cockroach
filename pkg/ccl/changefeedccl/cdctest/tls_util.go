package cdctest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"net"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

const certLifetime = 30 * 24 * time.Hour

func EncodeBase64ToString(src []byte, dest *string) {
	__antithesis_instrumentation__.Notify(15008)
	if src != nil {
		__antithesis_instrumentation__.Notify(15009)
		encoded := base64.StdEncoding.EncodeToString(src)
		*dest = encoded
	} else {
		__antithesis_instrumentation__.Notify(15010)
	}
}

func NewCACertBase64Encoded() (*tls.Certificate, string, error) {
	__antithesis_instrumentation__.Notify(15011)
	keyLength := 2048

	caKey, err := rsa.GenerateKey(rand.Reader, keyLength)
	if err != nil {
		__antithesis_instrumentation__.Notify(15017)
		return nil, "", errors.Wrap(err, "CA private key")
	} else {
		__antithesis_instrumentation__.Notify(15018)
	}
	__antithesis_instrumentation__.Notify(15012)

	caCert, _, err := GenerateCACert(caKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(15019)
		return nil, "", errors.Wrap(err, "CA cert gen")
	} else {
		__antithesis_instrumentation__.Notify(15020)
	}
	__antithesis_instrumentation__.Notify(15013)

	caKeyPEM, err := PemEncodePrivateKey(caKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(15021)
		return nil, "", errors.Wrap(err, "pem encode CA key")
	} else {
		__antithesis_instrumentation__.Notify(15022)
	}
	__antithesis_instrumentation__.Notify(15014)

	caCertPEM, err := PemEncodeCert(caCert)
	if err != nil {
		__antithesis_instrumentation__.Notify(15023)
		return nil, "", errors.Wrap(err, "pem encode CA cert")
	} else {
		__antithesis_instrumentation__.Notify(15024)
	}
	__antithesis_instrumentation__.Notify(15015)

	cert, err := tls.X509KeyPair([]byte(caCertPEM), []byte(caKeyPEM))
	if err != nil {
		__antithesis_instrumentation__.Notify(15025)
		return nil, "", errors.Wrap(err, "CA cert parse from PEM")
	} else {
		__antithesis_instrumentation__.Notify(15026)
	}
	__antithesis_instrumentation__.Notify(15016)

	var caCertBase64 string
	EncodeBase64ToString([]byte(caCertPEM), &caCertBase64)

	return &cert, caCertBase64, nil
}

func GenerateCACert(priv *rsa.PrivateKey) ([]byte, *x509.Certificate, error) {
	__antithesis_instrumentation__.Notify(15027)
	serial, err := randomSerial()
	if err != nil {
		__antithesis_instrumentation__.Notify(15029)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(15030)
	}
	__antithesis_instrumentation__.Notify(15028)

	certSpec := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Country:            []string{"US"},
			Organization:       []string{"Cockroach Labs"},
			OrganizationalUnit: []string{"Engineering"},
			CommonName:         "Roachtest Temporary Insecure CA",
		},
		NotBefore:             timeutil.Now(),
		NotAfter:              timeutil.Now().Add(certLifetime),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		IsCA:                  true,
		BasicConstraintsValid: true,
		MaxPathLenZero:        true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}
	cert, err := x509.CreateCertificate(rand.Reader, certSpec, certSpec, &priv.PublicKey, priv)
	return cert, certSpec, err
}

func pemEncode(dataType string, data []byte) (string, error) {
	__antithesis_instrumentation__.Notify(15031)
	ret := new(strings.Builder)
	err := pem.Encode(ret, &pem.Block{Type: dataType, Bytes: data})
	if err != nil {
		__antithesis_instrumentation__.Notify(15033)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(15034)
	}
	__antithesis_instrumentation__.Notify(15032)

	return ret.String(), nil
}

func PemEncodePrivateKey(key *rsa.PrivateKey) (string, error) {
	__antithesis_instrumentation__.Notify(15035)
	return pemEncode("RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(key))
}

func PemEncodeCert(cert []byte) (string, error) {
	__antithesis_instrumentation__.Notify(15036)
	return pemEncode("CERTIFICATE", cert)
}

func GenerateClientCertAndKey(caCert *tls.Certificate) ([]byte, []byte, error) {
	__antithesis_instrumentation__.Notify(15037)
	clientCert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    timeutil.Now(),
		NotAfter:     timeutil.Now().Add(certLifetime),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	clientKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		__antithesis_instrumentation__.Notify(15043)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(15044)
	}
	__antithesis_instrumentation__.Notify(15038)

	cert, err := x509.ParseCertificate(caCert.Certificate[0])
	if err != nil {
		__antithesis_instrumentation__.Notify(15045)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(15046)
	}
	__antithesis_instrumentation__.Notify(15039)

	clientCertBytes, err := x509.CreateCertificate(rand.Reader, clientCert, cert, &clientKey.PublicKey, caCert.PrivateKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(15047)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(15048)
	}
	__antithesis_instrumentation__.Notify(15040)

	clientCertPEM := new(bytes.Buffer)
	err = pem.Encode(clientCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: clientCertBytes,
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(15049)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(15050)
	}
	__antithesis_instrumentation__.Notify(15041)

	clientKeyPEM := new(bytes.Buffer)
	err = pem.Encode(clientKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(clientKey),
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(15051)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(15052)
	}
	__antithesis_instrumentation__.Notify(15042)

	return clientCertPEM.Bytes(), clientKeyPEM.Bytes(), nil
}

func randomSerial() (*big.Int, error) {
	__antithesis_instrumentation__.Notify(15053)
	limit := new(big.Int).Lsh(big.NewInt(1), 128)
	ret, err := rand.Int(rand.Reader, limit)
	if err != nil {
		__antithesis_instrumentation__.Notify(15055)
		return nil, errors.Wrap(err, "generate random serial")
	} else {
		__antithesis_instrumentation__.Notify(15056)
	}
	__antithesis_instrumentation__.Notify(15054)
	return ret, nil
}
