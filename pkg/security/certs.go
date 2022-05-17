package security

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/oserror"
)

const (
	keyFileMode  = 0600
	certFileMode = 0644
)

func loadCACertAndKey(sslCA, sslCAKey string) (*x509.Certificate, crypto.PrivateKey, error) {
	__antithesis_instrumentation__.Notify(186400)

	caCert, err := tls.LoadX509KeyPair(sslCA, sslCAKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(186403)
		return nil, nil, errors.Wrapf(err, "error loading CA certificate %s and key %s",
			sslCA, sslCAKey)
	} else {
		__antithesis_instrumentation__.Notify(186404)
	}
	__antithesis_instrumentation__.Notify(186401)

	x509Cert, err := x509.ParseCertificate(caCert.Certificate[0])
	if err != nil {
		__antithesis_instrumentation__.Notify(186405)
		return nil, nil, errors.Wrapf(err, "error parsing CA certificate %s", sslCA)
	} else {
		__antithesis_instrumentation__.Notify(186406)
	}
	__antithesis_instrumentation__.Notify(186402)
	return x509Cert, caCert.PrivateKey, nil
}

func writeCertificateToFile(certFilePath string, certificate []byte, overwrite bool) error {
	__antithesis_instrumentation__.Notify(186407)
	certBlock := &pem.Block{Type: "CERTIFICATE", Bytes: certificate}

	return WritePEMToFile(certFilePath, certFileMode, overwrite, certBlock)
}

func writeKeyToFile(keyFilePath string, key crypto.PrivateKey, overwrite bool) error {
	__antithesis_instrumentation__.Notify(186408)
	keyBlock, err := PrivateKeyToPEM(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(186410)
		return err
	} else {
		__antithesis_instrumentation__.Notify(186411)
	}
	__antithesis_instrumentation__.Notify(186409)

	return WritePEMToFile(keyFilePath, keyFileMode, overwrite, keyBlock)
}

func writePKCS8KeyToFile(keyFilePath string, key crypto.PrivateKey, overwrite bool) error {
	__antithesis_instrumentation__.Notify(186412)
	keyBytes, err := PrivateKeyToPKCS8(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(186414)
		return err
	} else {
		__antithesis_instrumentation__.Notify(186415)
	}
	__antithesis_instrumentation__.Notify(186413)

	return SafeWriteToFile(keyFilePath, keyFileMode, overwrite, keyBytes)
}

func CreateCAPair(
	certsDir, caKeyPath string,
	keySize int,
	lifetime time.Duration,
	allowKeyReuse bool,
	overwrite bool,
) error {
	__antithesis_instrumentation__.Notify(186416)
	return createCACertAndKey(certsDir, caKeyPath, CAPem, keySize, lifetime, allowKeyReuse, overwrite)
}

func CreateTenantCAPair(
	certsDir, caKeyPath string,
	keySize int,
	lifetime time.Duration,
	allowKeyReuse bool,
	overwrite bool,
) error {
	__antithesis_instrumentation__.Notify(186417)
	return createCACertAndKey(certsDir, caKeyPath, TenantCAPem, keySize, lifetime, allowKeyReuse, overwrite)
}

func CreateClientCAPair(
	certsDir, caKeyPath string,
	keySize int,
	lifetime time.Duration,
	allowKeyReuse bool,
	overwrite bool,
) error {
	__antithesis_instrumentation__.Notify(186418)
	return createCACertAndKey(certsDir, caKeyPath, ClientCAPem, keySize, lifetime, allowKeyReuse, overwrite)
}

func CreateUICAPair(
	certsDir, caKeyPath string,
	keySize int,
	lifetime time.Duration,
	allowKeyReuse bool,
	overwrite bool,
) error {
	__antithesis_instrumentation__.Notify(186419)
	return createCACertAndKey(certsDir, caKeyPath, UICAPem, keySize, lifetime, allowKeyReuse, overwrite)
}

func createCACertAndKey(
	certsDir, caKeyPath string,
	caType PemUsage,
	keySize int,
	lifetime time.Duration,
	allowKeyReuse bool,
	overwrite bool,
) error {
	__antithesis_instrumentation__.Notify(186420)
	if len(caKeyPath) == 0 {
		__antithesis_instrumentation__.Notify(186430)
		return errors.New("the path to the CA key is required")
	} else {
		__antithesis_instrumentation__.Notify(186431)
	}
	__antithesis_instrumentation__.Notify(186421)
	if len(certsDir) == 0 {
		__antithesis_instrumentation__.Notify(186432)
		return errors.New("the path to the certs directory is required")
	} else {
		__antithesis_instrumentation__.Notify(186433)
	}
	__antithesis_instrumentation__.Notify(186422)
	if caType != CAPem && func() bool {
		__antithesis_instrumentation__.Notify(186434)
		return caType != TenantCAPem == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(186435)
		return caType != ClientCAPem == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(186436)
		return caType != UICAPem == true
	}() == true {
		__antithesis_instrumentation__.Notify(186437)

		return fmt.Errorf("caType argument to createCACertAndKey must be one of CAPem (%d), ClientCAPem (%d), or UICAPem (%d), got: %d",
			CAPem, ClientCAPem, UICAPem, caType)
	} else {
		__antithesis_instrumentation__.Notify(186438)
	}
	__antithesis_instrumentation__.Notify(186423)

	caKeyPath = os.ExpandEnv(caKeyPath)

	cm, err := NewCertificateManagerFirstRun(certsDir, CommandTLSSettings{})
	if err != nil {
		__antithesis_instrumentation__.Notify(186439)
		return err
	} else {
		__antithesis_instrumentation__.Notify(186440)
	}
	__antithesis_instrumentation__.Notify(186424)

	var key crypto.PrivateKey
	if _, err := os.Stat(caKeyPath); err != nil {
		__antithesis_instrumentation__.Notify(186441)
		if !oserror.IsNotExist(err) {
			__antithesis_instrumentation__.Notify(186445)
			return errors.Wrapf(err, "could not stat CA key file %s", caKeyPath)
		} else {
			__antithesis_instrumentation__.Notify(186446)
		}
		__antithesis_instrumentation__.Notify(186442)

		key, err = rsa.GenerateKey(rand.Reader, keySize)
		if err != nil {
			__antithesis_instrumentation__.Notify(186447)
			return errors.Wrap(err, "could not generate new CA key")
		} else {
			__antithesis_instrumentation__.Notify(186448)
		}
		__antithesis_instrumentation__.Notify(186443)

		if err := writeKeyToFile(caKeyPath, key, overwrite); err != nil {
			__antithesis_instrumentation__.Notify(186449)
			return errors.Wrapf(err, "could not write CA key to file %s", caKeyPath)
		} else {
			__antithesis_instrumentation__.Notify(186450)
		}
		__antithesis_instrumentation__.Notify(186444)

		log.Infof(context.Background(), "generated CA key %s", caKeyPath)
	} else {
		__antithesis_instrumentation__.Notify(186451)
		if !allowKeyReuse {
			__antithesis_instrumentation__.Notify(186455)
			return errors.Errorf("CA key %s exists, but key reuse is disabled", caKeyPath)
		} else {
			__antithesis_instrumentation__.Notify(186456)
		}
		__antithesis_instrumentation__.Notify(186452)

		contents, err := ioutil.ReadFile(caKeyPath)
		if err != nil {
			__antithesis_instrumentation__.Notify(186457)
			return errors.Wrapf(err, "could not read CA key file %s", caKeyPath)
		} else {
			__antithesis_instrumentation__.Notify(186458)
		}
		__antithesis_instrumentation__.Notify(186453)

		key, err = PEMToPrivateKey(contents)
		if err != nil {
			__antithesis_instrumentation__.Notify(186459)
			return errors.Wrapf(err, "could not parse CA key file %s", caKeyPath)
		} else {
			__antithesis_instrumentation__.Notify(186460)
		}
		__antithesis_instrumentation__.Notify(186454)

		log.Infof(context.Background(), "using CA key from file %s", caKeyPath)
	}
	__antithesis_instrumentation__.Notify(186425)

	certContents, err := GenerateCA(key.(crypto.Signer), lifetime)
	if err != nil {
		__antithesis_instrumentation__.Notify(186461)
		return errors.Wrap(err, "could not generate CA certificate")
	} else {
		__antithesis_instrumentation__.Notify(186462)
	}
	__antithesis_instrumentation__.Notify(186426)

	var certPath string

	switch caType {
	case CAPem:
		__antithesis_instrumentation__.Notify(186463)
		certPath = cm.CACertPath()
	case TenantCAPem:
		__antithesis_instrumentation__.Notify(186464)
		certPath = cm.TenantCACertPath()
	case ClientCAPem:
		__antithesis_instrumentation__.Notify(186465)
		certPath = cm.ClientCACertPath()
	case UICAPem:
		__antithesis_instrumentation__.Notify(186466)
		certPath = cm.UICACertPath()
	default:
		__antithesis_instrumentation__.Notify(186467)
		return errors.Newf("unknown CA type %v", caType)
	}
	__antithesis_instrumentation__.Notify(186427)

	var existingCertificates []*pem.Block
	if _, err := os.Stat(certPath); err == nil {
		__antithesis_instrumentation__.Notify(186468)

		contents, err := ioutil.ReadFile(certPath)
		if err != nil {
			__antithesis_instrumentation__.Notify(186471)
			return errors.Wrapf(err, "could not read existing CA cert file %s", certPath)
		} else {
			__antithesis_instrumentation__.Notify(186472)
		}
		__antithesis_instrumentation__.Notify(186469)

		existingCertificates, err = PEMToCertificates(contents)
		if err != nil {
			__antithesis_instrumentation__.Notify(186473)
			return errors.Wrapf(err, "could not parse existing CA cert file %s", certPath)
		} else {
			__antithesis_instrumentation__.Notify(186474)
		}
		__antithesis_instrumentation__.Notify(186470)
		log.Infof(context.Background(), "found %d certificates in %s",
			len(existingCertificates), certPath)
	} else {
		__antithesis_instrumentation__.Notify(186475)
		if !oserror.IsNotExist(err) {
			__antithesis_instrumentation__.Notify(186476)
			return errors.Wrapf(err, "could not stat CA cert file %s", certPath)
		} else {
			__antithesis_instrumentation__.Notify(186477)
		}
	}
	__antithesis_instrumentation__.Notify(186428)

	certificates := []*pem.Block{{Type: "CERTIFICATE", Bytes: certContents}}
	certificates = append(certificates, existingCertificates...)

	if err := WritePEMToFile(certPath, certFileMode, overwrite, certificates...); err != nil {
		__antithesis_instrumentation__.Notify(186478)
		return errors.Wrapf(err, "could not write CA certificate file %s", certPath)
	} else {
		__antithesis_instrumentation__.Notify(186479)
	}
	__antithesis_instrumentation__.Notify(186429)

	log.Infof(context.Background(), "wrote %d certificates to %s", len(certificates), certPath)

	return nil
}

func CreateNodePair(
	certsDir, caKeyPath string, keySize int, lifetime time.Duration, overwrite bool, hosts []string,
) error {
	__antithesis_instrumentation__.Notify(186480)
	if len(caKeyPath) == 0 {
		__antithesis_instrumentation__.Notify(186489)
		return errors.New("the path to the CA key is required")
	} else {
		__antithesis_instrumentation__.Notify(186490)
	}
	__antithesis_instrumentation__.Notify(186481)
	if len(certsDir) == 0 {
		__antithesis_instrumentation__.Notify(186491)
		return errors.New("the path to the certs directory is required")
	} else {
		__antithesis_instrumentation__.Notify(186492)
	}
	__antithesis_instrumentation__.Notify(186482)

	caKeyPath = os.ExpandEnv(caKeyPath)

	cm, err := NewCertificateManagerFirstRun(certsDir, CommandTLSSettings{})
	if err != nil {
		__antithesis_instrumentation__.Notify(186493)
		return err
	} else {
		__antithesis_instrumentation__.Notify(186494)
	}
	__antithesis_instrumentation__.Notify(186483)

	caCert, caPrivateKey, err := loadCACertAndKey(cm.CACertPath(), caKeyPath)
	if err != nil {
		__antithesis_instrumentation__.Notify(186495)
		return err
	} else {
		__antithesis_instrumentation__.Notify(186496)
	}
	__antithesis_instrumentation__.Notify(186484)

	nodeKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		__antithesis_instrumentation__.Notify(186497)
		return errors.Wrap(err, "could not generate new node key")
	} else {
		__antithesis_instrumentation__.Notify(186498)
	}
	__antithesis_instrumentation__.Notify(186485)

	nodeUser, _ := MakeSQLUsernameFromUserInput(
		envutil.EnvOrDefaultString("COCKROACH_CERT_NODE_USER", NodeUser),
		UsernameValidation)

	nodeCert, err := GenerateServerCert(caCert, caPrivateKey,
		nodeKey.Public(), lifetime, nodeUser, hosts)
	if err != nil {
		__antithesis_instrumentation__.Notify(186499)
		return errors.Wrap(err, "error creating node server certificate and key")
	} else {
		__antithesis_instrumentation__.Notify(186500)
	}
	__antithesis_instrumentation__.Notify(186486)

	certPath := cm.NodeCertPath()
	if err := writeCertificateToFile(certPath, nodeCert, overwrite); err != nil {
		__antithesis_instrumentation__.Notify(186501)
		return errors.Wrapf(err, "error writing node server certificate to %s", certPath)
	} else {
		__antithesis_instrumentation__.Notify(186502)
	}
	__antithesis_instrumentation__.Notify(186487)
	log.Infof(context.Background(), "generated node certificate: %s", certPath)

	keyPath := cm.NodeKeyPath()
	if err := writeKeyToFile(keyPath, nodeKey, overwrite); err != nil {
		__antithesis_instrumentation__.Notify(186503)
		return errors.Wrapf(err, "error writing node server key to %s", keyPath)
	} else {
		__antithesis_instrumentation__.Notify(186504)
	}
	__antithesis_instrumentation__.Notify(186488)
	log.Infof(context.Background(), "generated node key: %s", keyPath)

	return nil
}

func CreateUIPair(
	certsDir, caKeyPath string, keySize int, lifetime time.Duration, overwrite bool, hosts []string,
) error {
	__antithesis_instrumentation__.Notify(186505)
	if len(caKeyPath) == 0 {
		__antithesis_instrumentation__.Notify(186514)
		return errors.New("the path to the CA key is required")
	} else {
		__antithesis_instrumentation__.Notify(186515)
	}
	__antithesis_instrumentation__.Notify(186506)
	if len(certsDir) == 0 {
		__antithesis_instrumentation__.Notify(186516)
		return errors.New("the path to the certs directory is required")
	} else {
		__antithesis_instrumentation__.Notify(186517)
	}
	__antithesis_instrumentation__.Notify(186507)

	caKeyPath = os.ExpandEnv(caKeyPath)

	cm, err := NewCertificateManagerFirstRun(certsDir, CommandTLSSettings{})
	if err != nil {
		__antithesis_instrumentation__.Notify(186518)
		return err
	} else {
		__antithesis_instrumentation__.Notify(186519)
	}
	__antithesis_instrumentation__.Notify(186508)

	caCert, caPrivateKey, err := loadCACertAndKey(cm.UICACertPath(), caKeyPath)
	if err != nil {
		__antithesis_instrumentation__.Notify(186520)
		return err
	} else {
		__antithesis_instrumentation__.Notify(186521)
	}
	__antithesis_instrumentation__.Notify(186509)

	uiKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		__antithesis_instrumentation__.Notify(186522)
		return errors.Wrap(err, "could not generate new UI key")
	} else {
		__antithesis_instrumentation__.Notify(186523)
	}
	__antithesis_instrumentation__.Notify(186510)

	uiCert, err := GenerateUIServerCert(caCert, caPrivateKey, uiKey.Public(), lifetime, hosts)
	if err != nil {
		__antithesis_instrumentation__.Notify(186524)
		return errors.Wrap(err, "error creating UI server certificate and key")
	} else {
		__antithesis_instrumentation__.Notify(186525)
	}
	__antithesis_instrumentation__.Notify(186511)

	certPath := cm.UICertPath()
	if err := writeCertificateToFile(certPath, uiCert, overwrite); err != nil {
		__antithesis_instrumentation__.Notify(186526)
		return errors.Wrapf(err, "error writing UI server certificate to %s", certPath)
	} else {
		__antithesis_instrumentation__.Notify(186527)
	}
	__antithesis_instrumentation__.Notify(186512)
	log.Infof(context.Background(), "generated UI certificate: %s", certPath)

	keyPath := cm.UIKeyPath()
	if err := writeKeyToFile(keyPath, uiKey, overwrite); err != nil {
		__antithesis_instrumentation__.Notify(186528)
		return errors.Wrapf(err, "error writing UI server key to %s", keyPath)
	} else {
		__antithesis_instrumentation__.Notify(186529)
	}
	__antithesis_instrumentation__.Notify(186513)
	log.Infof(context.Background(), "generated UI key: %s", keyPath)

	return nil
}

func CreateClientPair(
	certsDir, caKeyPath string,
	keySize int,
	lifetime time.Duration,
	overwrite bool,
	user SQLUsername,
	wantPKCS8Key bool,
) error {
	__antithesis_instrumentation__.Notify(186530)
	if len(caKeyPath) == 0 {
		__antithesis_instrumentation__.Notify(186541)
		return errors.New("the path to the CA key is required")
	} else {
		__antithesis_instrumentation__.Notify(186542)
	}
	__antithesis_instrumentation__.Notify(186531)
	if len(certsDir) == 0 {
		__antithesis_instrumentation__.Notify(186543)
		return errors.New("the path to the certs directory is required")
	} else {
		__antithesis_instrumentation__.Notify(186544)
	}
	__antithesis_instrumentation__.Notify(186532)

	caKeyPath = os.ExpandEnv(caKeyPath)

	cm, err := NewCertificateManagerFirstRun(certsDir, CommandTLSSettings{})
	if err != nil {
		__antithesis_instrumentation__.Notify(186545)
		return err
	} else {
		__antithesis_instrumentation__.Notify(186546)
	}
	__antithesis_instrumentation__.Notify(186533)

	var caCertPath string

	if cm.ClientCACert() != nil {
		__antithesis_instrumentation__.Notify(186547)
		caCertPath = cm.ClientCACertPath()
	} else {
		__antithesis_instrumentation__.Notify(186548)
		caCertPath = cm.CACertPath()
	}
	__antithesis_instrumentation__.Notify(186534)

	caCert, caPrivateKey, err := loadCACertAndKey(caCertPath, caKeyPath)
	if err != nil {
		__antithesis_instrumentation__.Notify(186549)
		return err
	} else {
		__antithesis_instrumentation__.Notify(186550)
	}
	__antithesis_instrumentation__.Notify(186535)

	clientKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		__antithesis_instrumentation__.Notify(186551)
		return errors.Wrap(err, "could not generate new client key")
	} else {
		__antithesis_instrumentation__.Notify(186552)
	}
	__antithesis_instrumentation__.Notify(186536)

	clientCert, err := GenerateClientCert(caCert, caPrivateKey, clientKey.Public(), lifetime, user)
	if err != nil {
		__antithesis_instrumentation__.Notify(186553)
		return errors.Wrap(err, "error creating client certificate and key")
	} else {
		__antithesis_instrumentation__.Notify(186554)
	}
	__antithesis_instrumentation__.Notify(186537)

	certPath := cm.ClientCertPath(user)
	if err := writeCertificateToFile(certPath, clientCert, overwrite); err != nil {
		__antithesis_instrumentation__.Notify(186555)
		return errors.Wrapf(err, "error writing client certificate to %s", certPath)
	} else {
		__antithesis_instrumentation__.Notify(186556)
	}
	__antithesis_instrumentation__.Notify(186538)
	log.Infof(context.Background(), "generated client certificate: %s", certPath)

	keyPath := cm.ClientKeyPath(user)
	if err := writeKeyToFile(keyPath, clientKey, overwrite); err != nil {
		__antithesis_instrumentation__.Notify(186557)
		return errors.Wrapf(err, "error writing client key to %s", keyPath)
	} else {
		__antithesis_instrumentation__.Notify(186558)
	}
	__antithesis_instrumentation__.Notify(186539)
	log.Infof(context.Background(), "generated client key: %s", keyPath)

	if wantPKCS8Key {
		__antithesis_instrumentation__.Notify(186559)
		pkcs8KeyPath := keyPath + ".pk8"
		if err := writePKCS8KeyToFile(pkcs8KeyPath, clientKey, overwrite); err != nil {
			__antithesis_instrumentation__.Notify(186561)
			return errors.Wrapf(err, "error writing client PKCS8 key to %s", pkcs8KeyPath)
		} else {
			__antithesis_instrumentation__.Notify(186562)
		}
		__antithesis_instrumentation__.Notify(186560)
		log.Infof(context.Background(), "generated PKCS8 client key: %s", pkcs8KeyPath)
	} else {
		__antithesis_instrumentation__.Notify(186563)
	}
	__antithesis_instrumentation__.Notify(186540)

	return nil
}

type TenantPair struct {
	PrivateKey *rsa.PrivateKey
	Cert       []byte
}

func CreateTenantPair(
	certsDir, caKeyPath string,
	keySize int,
	lifetime time.Duration,
	tenantIdentifier uint64,
	hosts []string,
) (*TenantPair, error) {
	__antithesis_instrumentation__.Notify(186564)
	if len(caKeyPath) == 0 {
		__antithesis_instrumentation__.Notify(186572)
		return nil, errors.New("the path to the CA key is required")
	} else {
		__antithesis_instrumentation__.Notify(186573)
	}
	__antithesis_instrumentation__.Notify(186565)
	if len(certsDir) == 0 {
		__antithesis_instrumentation__.Notify(186574)
		return nil, errors.New("the path to the certs directory is required")
	} else {
		__antithesis_instrumentation__.Notify(186575)
	}
	__antithesis_instrumentation__.Notify(186566)

	caKeyPath = os.ExpandEnv(caKeyPath)

	cm, err := NewCertificateManagerFirstRun(certsDir, CommandTLSSettings{})
	if err != nil {
		__antithesis_instrumentation__.Notify(186576)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(186577)
	}
	__antithesis_instrumentation__.Notify(186567)

	clientCA, err := cm.getTenantCACertLocked()
	if err != nil {
		__antithesis_instrumentation__.Notify(186578)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(186579)
	}
	__antithesis_instrumentation__.Notify(186568)

	caCert, caPrivateKey, err := loadCACertAndKey(filepath.Join(certsDir, clientCA.Filename), caKeyPath)
	if err != nil {
		__antithesis_instrumentation__.Notify(186580)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(186581)
	}
	__antithesis_instrumentation__.Notify(186569)

	clientKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		__antithesis_instrumentation__.Notify(186582)
		return nil, errors.Wrap(err, "could not generate new tenant key")
	} else {
		__antithesis_instrumentation__.Notify(186583)
	}
	__antithesis_instrumentation__.Notify(186570)

	clientCert, err := GenerateTenantCert(
		caCert, caPrivateKey, clientKey.Public(), lifetime, tenantIdentifier, hosts,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(186584)
		return nil, errors.Wrap(err, "error creating tenant certificate and key")
	} else {
		__antithesis_instrumentation__.Notify(186585)
	}
	__antithesis_instrumentation__.Notify(186571)
	return &TenantPair{
		PrivateKey: clientKey,
		Cert:       clientCert,
	}, nil
}

func WriteTenantPair(certsDir string, cp *TenantPair, overwrite bool) error {
	__antithesis_instrumentation__.Notify(186586)
	cm, err := NewCertificateManagerFirstRun(certsDir, CommandTLSSettings{})
	if err != nil {
		__antithesis_instrumentation__.Notify(186591)
		return err
	} else {
		__antithesis_instrumentation__.Notify(186592)
	}
	__antithesis_instrumentation__.Notify(186587)
	cert, err := x509.ParseCertificate(cp.Cert)
	if err != nil {
		__antithesis_instrumentation__.Notify(186593)
		return err
	} else {
		__antithesis_instrumentation__.Notify(186594)
	}
	__antithesis_instrumentation__.Notify(186588)
	tenantIdentifier := cert.Subject.CommonName
	certPath := cm.TenantCertPath(tenantIdentifier)
	if err := writeCertificateToFile(certPath, cp.Cert, overwrite); err != nil {
		__antithesis_instrumentation__.Notify(186595)
		return errors.Wrapf(err, "error writing tenant certificate to %s", certPath)
	} else {
		__antithesis_instrumentation__.Notify(186596)
	}
	__antithesis_instrumentation__.Notify(186589)
	log.Infof(context.Background(), "wrote SQL tenant client certificate: %s", certPath)

	keyPath := cm.TenantKeyPath(tenantIdentifier)
	if err := writeKeyToFile(keyPath, cp.PrivateKey, overwrite); err != nil {
		__antithesis_instrumentation__.Notify(186597)
		return errors.Wrapf(err, "error writing tenant key to %s", keyPath)
	} else {
		__antithesis_instrumentation__.Notify(186598)
	}
	__antithesis_instrumentation__.Notify(186590)
	log.Infof(context.Background(), "generated tenant key: %s", keyPath)
	return nil
}

func CreateTenantSigningPair(
	certsDir string, lifetime time.Duration, overwrite bool, tenantID uint64,
) error {
	__antithesis_instrumentation__.Notify(186599)
	if len(certsDir) == 0 {
		__antithesis_instrumentation__.Notify(186607)
		return errors.New("the path to the certs directory is required")
	} else {
		__antithesis_instrumentation__.Notify(186608)
	}
	__antithesis_instrumentation__.Notify(186600)
	if tenantID == 0 {
		__antithesis_instrumentation__.Notify(186609)
		return errors.Errorf("tenantId %d is invalid (requires != 0)", tenantID)
	} else {
		__antithesis_instrumentation__.Notify(186610)
	}
	__antithesis_instrumentation__.Notify(186601)

	tenantIdentifier := fmt.Sprintf("%d", tenantID)

	cm, err := NewCertificateManagerFirstRun(certsDir, CommandTLSSettings{})
	if err != nil {
		__antithesis_instrumentation__.Notify(186611)
		return err
	} else {
		__antithesis_instrumentation__.Notify(186612)
	}
	__antithesis_instrumentation__.Notify(186602)

	signingKeyPath := cm.TenantSigningKeyPath(tenantIdentifier)
	signingCertPath := cm.TenantSigningCertPath(tenantIdentifier)
	var pubKey crypto.PublicKey
	var privKey crypto.PrivateKey
	pubKey, privKey, err = ed25519.GenerateKey(rand.Reader)
	if err != nil {
		__antithesis_instrumentation__.Notify(186613)
		return errors.Wrap(err, "could not generate new tenant signing key")
	} else {
		__antithesis_instrumentation__.Notify(186614)
	}
	__antithesis_instrumentation__.Notify(186603)

	if err := writeKeyToFile(signingKeyPath, privKey, overwrite); err != nil {
		__antithesis_instrumentation__.Notify(186615)
		return errors.Wrapf(err, "could not write tenant signing key to file %s", signingKeyPath)
	} else {
		__antithesis_instrumentation__.Notify(186616)
	}
	__antithesis_instrumentation__.Notify(186604)

	log.Infof(context.Background(), "generated tenant signing key %s", signingKeyPath)

	certContents, err := GenerateTenantSigningCert(pubKey, privKey, lifetime, tenantID)
	if err != nil {
		__antithesis_instrumentation__.Notify(186617)
		return errors.Wrap(err, "could not generate tenant signing certificate")
	} else {
		__antithesis_instrumentation__.Notify(186618)
	}
	__antithesis_instrumentation__.Notify(186605)

	certificates := []*pem.Block{{Type: "CERTIFICATE", Bytes: certContents}}

	if err := WritePEMToFile(signingCertPath, certFileMode, overwrite, certificates...); err != nil {
		__antithesis_instrumentation__.Notify(186619)
		return errors.Wrapf(err, "could not write tenant signing certificate file %s", signingCertPath)
	} else {
		__antithesis_instrumentation__.Notify(186620)
	}
	__antithesis_instrumentation__.Notify(186606)

	log.Infof(context.Background(), "wrote certificate to %s", signingCertPath)

	return nil
}

func PEMContentsToX509(contents []byte) ([]*x509.Certificate, error) {
	__antithesis_instrumentation__.Notify(186621)
	derCerts, err := PEMToCertificates(contents)
	if err != nil {
		__antithesis_instrumentation__.Notify(186624)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(186625)
	}
	__antithesis_instrumentation__.Notify(186622)

	certs := make([]*x509.Certificate, len(derCerts))
	for i, c := range derCerts {
		__antithesis_instrumentation__.Notify(186626)
		x509Cert, err := x509.ParseCertificate(c.Bytes)
		if err != nil {
			__antithesis_instrumentation__.Notify(186628)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(186629)
		}
		__antithesis_instrumentation__.Notify(186627)

		certs[i] = x509Cert
	}
	__antithesis_instrumentation__.Notify(186623)

	return certs, nil
}

func AppendCertificatesToBlob(certBlob []byte, newCerts ...[]byte) []byte {
	__antithesis_instrumentation__.Notify(186630)
	return bytes.Join(
		append(
			[][]byte{certBlob},
			newCerts...,
		),
		[]byte{'\n'},
	)
}
