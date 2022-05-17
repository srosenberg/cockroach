package security

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"crypto/tls"
	"crypto/x509"

	"github.com/cockroachdb/errors"
)

const (
	EmbeddedCertsDir     = "test_certs"
	EmbeddedCACert       = "ca.crt"
	EmbeddedCAKey        = "ca.key"
	EmbeddedClientCACert = "ca-client.crt"
	EmbeddedClientCAKey  = "ca-client.key"
	EmbeddedUICACert     = "ca-ui.crt"
	EmbeddedUICAKey      = "ca-ui.key"
	EmbeddedNodeCert     = "node.crt"
	EmbeddedNodeKey      = "node.key"
	EmbeddedRootCert     = "client.root.crt"
	EmbeddedRootKey      = "client.root.key"
	EmbeddedTestUserCert = "client.testuser.crt"
	EmbeddedTestUserKey  = "client.testuser.key"
)

var EmbeddedTenantIDs = func() []uint64 { __antithesis_instrumentation__.Notify(187300); return []uint64{10, 11, 20} }

const (
	EmbeddedTenantCACert = "ca-client-tenant.crt"
	EmbeddedTenantCAKey  = "ca-client-tenant.key"
)

func newServerTLSConfig(
	settings TLSSettings, certPEM, keyPEM, caPEM []byte, caClientPEMs ...[]byte,
) (*tls.Config, error) {
	__antithesis_instrumentation__.Notify(187301)
	cfg, err := newBaseTLSConfigWithCertificate(settings, certPEM, keyPEM, caPEM)
	if err != nil {
		__antithesis_instrumentation__.Notify(187304)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187305)
	}
	__antithesis_instrumentation__.Notify(187302)
	cfg.ClientAuth = tls.VerifyClientCertIfGiven

	if len(caClientPEMs) != 0 {
		__antithesis_instrumentation__.Notify(187306)
		certPool := x509.NewCertPool()
		for _, pem := range caClientPEMs {
			__antithesis_instrumentation__.Notify(187308)
			if !certPool.AppendCertsFromPEM(pem) {
				__antithesis_instrumentation__.Notify(187309)
				return nil, errors.Errorf("failed to parse client CA PEM data to pool")
			} else {
				__antithesis_instrumentation__.Notify(187310)
			}
		}
		__antithesis_instrumentation__.Notify(187307)
		cfg.ClientCAs = certPool
	} else {
		__antithesis_instrumentation__.Notify(187311)
	}
	__antithesis_instrumentation__.Notify(187303)

	cfg.PreferServerCipherSuites = true

	return cfg, nil
}

func newUIServerTLSConfig(settings TLSSettings, certPEM, keyPEM []byte) (*tls.Config, error) {
	__antithesis_instrumentation__.Notify(187312)
	cfg, err := newBaseTLSConfigWithCertificate(settings, certPEM, keyPEM, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(187314)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187315)
	}
	__antithesis_instrumentation__.Notify(187313)

	cfg.PreferServerCipherSuites = true

	return cfg, nil
}

func newClientTLSConfig(settings TLSSettings, certPEM, keyPEM, caPEM []byte) (*tls.Config, error) {
	__antithesis_instrumentation__.Notify(187316)
	return newBaseTLSConfigWithCertificate(settings, certPEM, keyPEM, caPEM)
}

func newUIClientTLSConfig(settings TLSSettings, caPEM []byte) (*tls.Config, error) {
	__antithesis_instrumentation__.Notify(187317)
	return newBaseTLSConfig(settings, caPEM)
}

func newBaseTLSConfigWithCertificate(
	settings TLSSettings, certPEM, keyPEM, caPEM []byte,
) (*tls.Config, error) {
	__antithesis_instrumentation__.Notify(187318)
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		__antithesis_instrumentation__.Notify(187321)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187322)
	}
	__antithesis_instrumentation__.Notify(187319)

	cfg, err := newBaseTLSConfig(settings, caPEM)
	if err != nil {
		__antithesis_instrumentation__.Notify(187323)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187324)
	}
	__antithesis_instrumentation__.Notify(187320)

	cfg.Certificates = []tls.Certificate{cert}
	return cfg, nil
}

func newBaseTLSConfig(settings TLSSettings, caPEM []byte) (*tls.Config, error) {
	__antithesis_instrumentation__.Notify(187325)
	var certPool *x509.CertPool
	if caPEM != nil {
		__antithesis_instrumentation__.Notify(187327)
		certPool = x509.NewCertPool()

		if !certPool.AppendCertsFromPEM(caPEM) {
			__antithesis_instrumentation__.Notify(187328)
			return nil, errors.Errorf("failed to parse PEM data to pool")
		} else {
			__antithesis_instrumentation__.Notify(187329)
		}
	} else {
		__antithesis_instrumentation__.Notify(187330)
	}
	__antithesis_instrumentation__.Notify(187326)

	return &tls.Config{
		RootCAs: certPool,

		VerifyPeerCertificate: makeOCSPVerifier(settings),

		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},

		MinVersion: tls.VersionTLS12,
	}, nil
}
