package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/pem"
	"io/ioutil"
	"os"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/netutil/addr"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/oserror"
	"github.com/cockroachdb/logtags"
)

const defaultCALifetime = 10 * 366 * 24 * time.Hour
const defaultCertLifetime = 5 * 366 * 24 * time.Hour

const serviceNameInterNode = "cockroach-node"
const serviceNameUserAuth = "cockroach-client"
const serviceNameSQL = "cockroach-sql"
const serviceNameRPC = "cockroach-rpc"
const serviceNameUI = "cockroach-http"

type CertificateBundle struct {
	InterNode      ServiceCertificateBundle
	UserAuth       ServiceCertificateBundle
	SQLService     ServiceCertificateBundle
	RPCService     ServiceCertificateBundle
	AdminUIService ServiceCertificateBundle
}

type ServiceCertificateBundle struct {
	CACertificate   []byte
	CAKey           []byte
	HostCertificate []byte
	HostKey         []byte
}

func (sb *ServiceCertificateBundle) loadServiceCertAndKey(
	certPath string, keyPath string,
) (err error) {
	__antithesis_instrumentation__.Notify(189390)
	sb.HostCertificate, err = loadCertificateFile(certPath)
	if err != nil {
		__antithesis_instrumentation__.Notify(189393)
		return
	} else {
		__antithesis_instrumentation__.Notify(189394)
	}
	__antithesis_instrumentation__.Notify(189391)
	sb.HostKey, err = loadKeyFile(keyPath)
	if err != nil {
		__antithesis_instrumentation__.Notify(189395)
		return
	} else {
		__antithesis_instrumentation__.Notify(189396)
	}
	__antithesis_instrumentation__.Notify(189392)
	return
}

func (sb *ServiceCertificateBundle) loadCACertAndKey(certPath string, keyPath string) (err error) {
	__antithesis_instrumentation__.Notify(189397)
	sb.CACertificate, err = loadCertificateFile(certPath)
	if err != nil {
		__antithesis_instrumentation__.Notify(189400)
		return
	} else {
		__antithesis_instrumentation__.Notify(189401)
	}
	__antithesis_instrumentation__.Notify(189398)
	sb.CAKey, err = loadKeyFile(keyPath)
	if err != nil {
		__antithesis_instrumentation__.Notify(189402)
		return
	} else {
		__antithesis_instrumentation__.Notify(189403)
	}
	__antithesis_instrumentation__.Notify(189399)
	return
}

func (sb *ServiceCertificateBundle) loadOrCreateServiceCertificates(
	ctx context.Context,
	serviceCertPath string,
	serviceKeyPath string,
	caCertPath string,
	caKeyPath string,
	serviceCertLifespan time.Duration,
	caCertLifespan time.Duration,
	commonName string,
	serviceName string,
	hostnames []string,
	serviceCertIsAlsoValidAsClient bool,
) error {
	__antithesis_instrumentation__.Notify(189404)
	ctx = logtags.AddTag(ctx, "service", serviceName)

	var err error
	log.Ops.Infof(ctx, "attempting to load service cert: %s", serviceCertPath)

	sb.HostCertificate, err = loadCertificateFile(serviceCertPath)
	if err == nil {
		__antithesis_instrumentation__.Notify(189412)
		log.Ops.Infof(ctx, "found; loading service key: %s", serviceKeyPath)

		sb.HostKey, err = loadKeyFile(serviceKeyPath)
		if err != nil {
			__antithesis_instrumentation__.Notify(189414)

			if oserror.IsNotExist(err) {
				__antithesis_instrumentation__.Notify(189416)

				return errors.Wrapf(err,
					"failed to load service certificate key for %q expected key at %q",
					serviceCertPath, serviceKeyPath)
			} else {
				__antithesis_instrumentation__.Notify(189417)
			}
			__antithesis_instrumentation__.Notify(189415)
			return errors.Wrap(err, "something went wrong loading service key")
		} else {
			__antithesis_instrumentation__.Notify(189418)
		}
		__antithesis_instrumentation__.Notify(189413)

		log.Ops.Infof(ctx, "service cert is ready")
		return nil
	} else {
		__antithesis_instrumentation__.Notify(189419)
	}
	__antithesis_instrumentation__.Notify(189405)

	log.Ops.Infof(ctx, "not found; will attempt auto-creation")

	log.Ops.Infof(ctx, "attempting to load CA cert: %s", caCertPath)

	sb.CACertificate, err = loadCertificateFile(caCertPath)
	if err == nil {
		__antithesis_instrumentation__.Notify(189420)

		log.Ops.Infof(ctx, "found; loading CA key: %s", caKeyPath)
		sb.CAKey, err = loadKeyFile(caKeyPath)
		if err != nil {
			__antithesis_instrumentation__.Notify(189421)
			return errors.Wrapf(
				err, "loaded service CA cert but failed to load service CA key file: %q", caKeyPath,
			)
		} else {
			__antithesis_instrumentation__.Notify(189422)
		}
	} else {
		__antithesis_instrumentation__.Notify(189423)
		if oserror.IsNotExist(err) {
			__antithesis_instrumentation__.Notify(189424)
			log.Ops.Infof(ctx, "not found; CA cert does not exist, auto-creating")

			if err := sb.createServiceCA(ctx, caCertPath, caKeyPath, caCertLifespan); err != nil {
				__antithesis_instrumentation__.Notify(189425)
				return errors.Wrap(
					err, "failed to create Service CA",
				)
			} else {
				__antithesis_instrumentation__.Notify(189426)
			}
		} else {
			__antithesis_instrumentation__.Notify(189427)
		}
	}
	__antithesis_instrumentation__.Notify(189406)

	var hostCert, hostKey *pem.Block
	caCertPEM, err := security.PEMToCertificates(sb.CACertificate)
	if err != nil {
		__antithesis_instrumentation__.Notify(189428)
		return errors.Wrap(err, "error when decoding PEM CACertificate")
	} else {
		__antithesis_instrumentation__.Notify(189429)
	}
	__antithesis_instrumentation__.Notify(189407)
	caKeyPEM, rest := pem.Decode(sb.CAKey)
	if len(rest) > 0 || func() bool {
		__antithesis_instrumentation__.Notify(189430)
		return caKeyPEM == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(189431)
		return errors.New("error when decoding PEM CAKey")
	} else {
		__antithesis_instrumentation__.Notify(189432)
	}
	__antithesis_instrumentation__.Notify(189408)
	hostCert, hostKey, err = security.CreateServiceCertAndKey(
		ctx,
		log.Ops.Infof,
		serviceCertLifespan,
		commonName,
		hostnames,
		caCertPEM[0], caKeyPEM,
		serviceCertIsAlsoValidAsClient,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(189433)
		return errors.Wrap(
			err, "failed to create Service Cert and Key",
		)
	} else {
		__antithesis_instrumentation__.Notify(189434)
	}
	__antithesis_instrumentation__.Notify(189409)
	sb.HostCertificate = pem.EncodeToMemory(hostCert)
	sb.HostKey = pem.EncodeToMemory(hostKey)

	log.Ops.Infof(ctx, "writing service cert: %s", serviceCertPath)
	if err := writeCertificateFile(serviceCertPath, hostCert, false); err != nil {
		__antithesis_instrumentation__.Notify(189435)
		return err
	} else {
		__antithesis_instrumentation__.Notify(189436)
	}
	__antithesis_instrumentation__.Notify(189410)

	log.Ops.Infof(ctx, "writing service key: %s", serviceKeyPath)
	if err := writeKeyFile(serviceKeyPath, hostKey, false); err != nil {
		__antithesis_instrumentation__.Notify(189437)
		return err
	} else {
		__antithesis_instrumentation__.Notify(189438)
	}
	__antithesis_instrumentation__.Notify(189411)

	return nil
}

func (sb *ServiceCertificateBundle) createServiceCA(
	ctx context.Context, caCertPath string, caKeyPath string, initLifespan time.Duration,
) error {
	__antithesis_instrumentation__.Notify(189439)
	ctx = logtags.AddTag(ctx, "auto-create-ca", nil)

	var err error
	var caCert, caKey *pem.Block
	caCert, caKey, err = security.CreateCACertAndKey(ctx, log.Ops.Infof, initLifespan, caCommonName)
	if err != nil {
		__antithesis_instrumentation__.Notify(189443)
		return err
	} else {
		__antithesis_instrumentation__.Notify(189444)
	}
	__antithesis_instrumentation__.Notify(189440)
	sb.CACertificate = pem.EncodeToMemory(caCert)
	sb.CAKey = pem.EncodeToMemory(caKey)

	log.Ops.Infof(ctx, "writing CA cert: %s", caCertPath)
	if err := writeCertificateFile(caCertPath, caCert, false); err != nil {
		__antithesis_instrumentation__.Notify(189445)
		return err
	} else {
		__antithesis_instrumentation__.Notify(189446)
	}
	__antithesis_instrumentation__.Notify(189441)

	log.Ops.Infof(ctx, "writing CA key: %s", caKeyPath)
	if err := writeKeyFile(caKeyPath, caKey, false); err != nil {
		__antithesis_instrumentation__.Notify(189447)
		return err
	} else {
		__antithesis_instrumentation__.Notify(189448)
	}
	__antithesis_instrumentation__.Notify(189442)

	return nil
}

func loadCertificateFile(certPath string) (cert []byte, err error) {
	__antithesis_instrumentation__.Notify(189449)
	cert, err = ioutil.ReadFile(certPath)
	return
}

func loadKeyFile(keyPath string) (key []byte, err error) {
	__antithesis_instrumentation__.Notify(189450)
	key, err = ioutil.ReadFile(keyPath)
	return
}

func writeCertificateFile(certFilePath string, certificatePEM *pem.Block, overwrite bool) error {
	__antithesis_instrumentation__.Notify(189451)

	return security.WritePEMToFile(certFilePath, 0644, overwrite, certificatePEM)
}

func writeKeyFile(keyFilePath string, keyPEM *pem.Block, overwrite bool) error {
	__antithesis_instrumentation__.Notify(189452)

	return security.WritePEMToFile(keyFilePath, 0600, overwrite, keyPEM)
}

func (b *CertificateBundle) InitializeFromConfig(ctx context.Context, c base.Config) error {
	__antithesis_instrumentation__.Notify(189453)
	cl := security.MakeCertsLocator(c.SSLCertsDir)

	if exists, err := cl.HasNodeCert(); err != nil {
		__antithesis_instrumentation__.Notify(189462)
		return err
	} else {
		__antithesis_instrumentation__.Notify(189463)
		if exists {
			__antithesis_instrumentation__.Notify(189464)
			return errors.New("inter-node certificate already present")
		} else {
			__antithesis_instrumentation__.Notify(189465)
		}
	}
	__antithesis_instrumentation__.Notify(189454)

	rpcAddrs := extractHosts(c.Addr, c.AdvertiseAddr)
	sqlAddrs := rpcAddrs
	if c.SplitListenSQL {
		__antithesis_instrumentation__.Notify(189466)
		sqlAddrs = extractHosts(c.SQLAddr, c.SQLAdvertiseAddr)
	} else {
		__antithesis_instrumentation__.Notify(189467)
	}
	__antithesis_instrumentation__.Notify(189455)
	httpAddrs := extractHosts(c.HTTPAddr, c.HTTPAdvertiseAddr)

	if err := cl.EnsureCertsDirectory(); err != nil {
		__antithesis_instrumentation__.Notify(189468)
		return err
	} else {
		__antithesis_instrumentation__.Notify(189469)
	}
	__antithesis_instrumentation__.Notify(189456)

	if err := b.InterNode.loadOrCreateServiceCertificates(
		ctx,
		cl.NodeCertPath(),
		cl.NodeKeyPath(),
		cl.CACertPath(),
		cl.CAKeyPath(),
		defaultCertLifetime,
		defaultCALifetime,
		security.NodeUser,
		serviceNameInterNode,
		rpcAddrs,
		true,
	); err != nil {
		__antithesis_instrumentation__.Notify(189470)
		return errors.Wrap(err,
			"failed to load or create InterNode certificates")
	} else {
		__antithesis_instrumentation__.Notify(189471)
	}
	__antithesis_instrumentation__.Notify(189457)

	if err := b.UserAuth.loadOrCreateServiceCertificates(
		ctx,
		cl.ClientNodeCertPath(),
		cl.ClientNodeKeyPath(),
		cl.ClientCACertPath(),
		cl.ClientCAKeyPath(),
		defaultCertLifetime,
		defaultCALifetime,
		security.NodeUser,
		serviceNameUserAuth,
		nil,
		true,
	); err != nil {
		__antithesis_instrumentation__.Notify(189472)
		return errors.Wrap(err,
			"failed to load or create User auth certificate(s)")
	} else {
		__antithesis_instrumentation__.Notify(189473)
	}
	__antithesis_instrumentation__.Notify(189458)

	if err := b.SQLService.loadOrCreateServiceCertificates(
		ctx,
		cl.SQLServiceCertPath(),
		cl.SQLServiceKeyPath(),
		cl.SQLServiceCACertPath(),
		cl.SQLServiceCAKeyPath(),
		defaultCertLifetime,
		defaultCALifetime,
		security.NodeUser,
		serviceNameSQL,

		sqlAddrs,
		false,
	); err != nil {
		__antithesis_instrumentation__.Notify(189474)
		return errors.Wrap(err,
			"failed to load or create SQL service certificate(s)")
	} else {
		__antithesis_instrumentation__.Notify(189475)
	}
	__antithesis_instrumentation__.Notify(189459)

	if err := b.RPCService.loadOrCreateServiceCertificates(
		ctx,
		cl.RPCServiceCertPath(),
		cl.RPCServiceKeyPath(),
		cl.RPCServiceCACertPath(),
		cl.RPCServiceCAKeyPath(),
		defaultCertLifetime,
		defaultCALifetime,
		security.NodeUser,
		serviceNameRPC,

		rpcAddrs,
		false,
	); err != nil {
		__antithesis_instrumentation__.Notify(189476)
		return errors.Wrap(err,
			"failed to load or create RPC service certificate(s)")
	} else {
		__antithesis_instrumentation__.Notify(189477)
	}
	__antithesis_instrumentation__.Notify(189460)

	if err := b.AdminUIService.loadOrCreateServiceCertificates(
		ctx,
		cl.UICertPath(),
		cl.UIKeyPath(),
		cl.UICACertPath(),
		cl.UICAKeyPath(),
		defaultCertLifetime,
		defaultCALifetime,
		httpAddrs[0],
		serviceNameUI,
		httpAddrs,
		false,
	); err != nil {
		__antithesis_instrumentation__.Notify(189478)
		return errors.Wrap(err,
			"failed to load or create Admin UI service certificate(s)")
	} else {
		__antithesis_instrumentation__.Notify(189479)
	}
	__antithesis_instrumentation__.Notify(189461)

	return nil
}

func extractHosts(addrs ...string) []string {
	__antithesis_instrumentation__.Notify(189480)
	res := make([]string, 0, len(addrs))

	for _, a := range addrs {
		__antithesis_instrumentation__.Notify(189482)
		hostname, _, err := addr.SplitHostPort(a, "0")
		if err != nil {
			__antithesis_instrumentation__.Notify(189485)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(189486)
		}
		__antithesis_instrumentation__.Notify(189483)
		found := false
		for _, h := range res {
			__antithesis_instrumentation__.Notify(189487)
			if h == hostname {
				__antithesis_instrumentation__.Notify(189488)
				found = true
				break
			} else {
				__antithesis_instrumentation__.Notify(189489)
			}
		}
		__antithesis_instrumentation__.Notify(189484)
		if !found {
			__antithesis_instrumentation__.Notify(189490)
			res = append(res, hostname)
		} else {
			__antithesis_instrumentation__.Notify(189491)
		}
	}
	__antithesis_instrumentation__.Notify(189481)
	return res
}

func (b *CertificateBundle) InitializeNodeFromBundle(ctx context.Context, c base.Config) error {
	__antithesis_instrumentation__.Notify(189492)
	cl := security.MakeCertsLocator(c.SSLCertsDir)

	if exists, err := cl.HasNodeCert(); err != nil {
		__antithesis_instrumentation__.Notify(189501)
		return err
	} else {
		__antithesis_instrumentation__.Notify(189502)
		if exists {
			__antithesis_instrumentation__.Notify(189503)
			return errors.New("inter-node certificate already present")
		} else {
			__antithesis_instrumentation__.Notify(189504)
		}
	}
	__antithesis_instrumentation__.Notify(189493)

	if err := cl.EnsureCertsDirectory(); err != nil {
		__antithesis_instrumentation__.Notify(189505)
		return err
	} else {
		__antithesis_instrumentation__.Notify(189506)
	}
	__antithesis_instrumentation__.Notify(189494)

	if err := b.InterNode.writeCAOrFail(cl.CACertPath(), cl.CAKeyPath()); err != nil {
		__antithesis_instrumentation__.Notify(189507)
		return errors.Wrap(err, "failed to write InterNodeCA to disk")
	} else {
		__antithesis_instrumentation__.Notify(189508)
	}
	__antithesis_instrumentation__.Notify(189495)

	if err := b.UserAuth.writeCAOrFail(cl.ClientCACertPath(), cl.ClientCAKeyPath()); err != nil {
		__antithesis_instrumentation__.Notify(189509)
		return errors.Wrap(err, "failed to write ClientCA to disk")
	} else {
		__antithesis_instrumentation__.Notify(189510)
	}
	__antithesis_instrumentation__.Notify(189496)

	if err := b.SQLService.writeCAOrFail(cl.SQLServiceCACertPath(), cl.SQLServiceCAKeyPath()); err != nil {
		__antithesis_instrumentation__.Notify(189511)
		return errors.Wrap(err, "failed to write SQLServiceCA to disk")
	} else {
		__antithesis_instrumentation__.Notify(189512)
	}
	__antithesis_instrumentation__.Notify(189497)

	if err := b.RPCService.writeCAOrFail(cl.RPCServiceCACertPath(), cl.RPCServiceCAKeyPath()); err != nil {
		__antithesis_instrumentation__.Notify(189513)
		return errors.Wrap(err, "failed to write RPCServiceCA to disk")
	} else {
		__antithesis_instrumentation__.Notify(189514)
	}
	__antithesis_instrumentation__.Notify(189498)

	if err := b.AdminUIService.writeCAOrFail(cl.UICACertPath(), cl.UICAKeyPath()); err != nil {
		__antithesis_instrumentation__.Notify(189515)
		return errors.Wrap(err, "failed to write AdminUIServiceCA to disk")
	} else {
		__antithesis_instrumentation__.Notify(189516)
	}
	__antithesis_instrumentation__.Notify(189499)

	if err := b.InitializeFromConfig(ctx, c); err != nil {
		__antithesis_instrumentation__.Notify(189517)
		return errors.Wrap(
			err,
			"failed to initialize host certs after writing CAs to disk")
	} else {
		__antithesis_instrumentation__.Notify(189518)
	}
	__antithesis_instrumentation__.Notify(189500)

	return nil
}

func (sb *ServiceCertificateBundle) writeCAOrFail(certPath string, keyPath string) (err error) {
	__antithesis_instrumentation__.Notify(189519)
	caCertPEM, _ := pem.Decode(sb.CACertificate)
	if sb.CACertificate != nil {
		__antithesis_instrumentation__.Notify(189522)
		err = writeCertificateFile(certPath, caCertPEM, false)
		if err != nil {
			__antithesis_instrumentation__.Notify(189523)
			return
		} else {
			__antithesis_instrumentation__.Notify(189524)
		}
	} else {
		__antithesis_instrumentation__.Notify(189525)
	}
	__antithesis_instrumentation__.Notify(189520)

	caKeyPEM, _ := pem.Decode(sb.CAKey)
	if sb.CAKey != nil {
		__antithesis_instrumentation__.Notify(189526)
		err = writeKeyFile(keyPath, caKeyPEM, false)
		if err != nil {
			__antithesis_instrumentation__.Notify(189527)
			return
		} else {
			__antithesis_instrumentation__.Notify(189528)
		}
	} else {
		__antithesis_instrumentation__.Notify(189529)
	}
	__antithesis_instrumentation__.Notify(189521)

	return
}

func (sb *ServiceCertificateBundle) loadCACertAndKeyIfExists(
	certPath string, keyPath string,
) error {
	__antithesis_instrumentation__.Notify(189530)

	err := sb.loadCACertAndKey(certPath, keyPath)
	if oserror.IsNotExist(err) {
		__antithesis_instrumentation__.Notify(189532)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(189533)
	}
	__antithesis_instrumentation__.Notify(189531)
	return err
}

func collectLocalCABundle(SSLCertsDir string) (CertificateBundle, error) {
	__antithesis_instrumentation__.Notify(189534)
	cl := security.MakeCertsLocator(SSLCertsDir)
	var b CertificateBundle
	var err error

	err = b.InterNode.loadCACertAndKeyIfExists(cl.CACertPath(), cl.CAKeyPath())
	if err != nil {
		__antithesis_instrumentation__.Notify(189540)
		return b, errors.Wrap(
			err, "error loading InterNode CA cert and/or key")
	} else {
		__antithesis_instrumentation__.Notify(189541)
	}
	__antithesis_instrumentation__.Notify(189535)

	err = b.UserAuth.loadCACertAndKeyIfExists(
		cl.ClientCACertPath(), cl.ClientCAKeyPath())
	if err != nil {
		__antithesis_instrumentation__.Notify(189542)
		return b, errors.Wrap(
			err, "error loading UserAuth CA cert and/or key")
	} else {
		__antithesis_instrumentation__.Notify(189543)
	}
	__antithesis_instrumentation__.Notify(189536)

	err = b.SQLService.loadCACertAndKeyIfExists(
		cl.SQLServiceCACertPath(), cl.SQLServiceCAKeyPath())
	if err != nil {
		__antithesis_instrumentation__.Notify(189544)
		return b, errors.Wrap(
			err, "error loading SQL CA cert and/or key")
	} else {
		__antithesis_instrumentation__.Notify(189545)
	}
	__antithesis_instrumentation__.Notify(189537)
	err = b.RPCService.loadCACertAndKeyIfExists(
		cl.RPCServiceCACertPath(), cl.RPCServiceCAKeyPath())
	if err != nil {
		__antithesis_instrumentation__.Notify(189546)
		return b, errors.Wrap(
			err, "error loading RPC CA cert and/or key")
	} else {
		__antithesis_instrumentation__.Notify(189547)
	}
	__antithesis_instrumentation__.Notify(189538)

	err = b.AdminUIService.loadCACertAndKeyIfExists(
		cl.UICACertPath(), cl.UICAKeyPath())
	if err != nil {
		__antithesis_instrumentation__.Notify(189548)
		return b, errors.Wrap(
			err, "error loading AdminUI CA cert and/or key")
	} else {
		__antithesis_instrumentation__.Notify(189549)
	}
	__antithesis_instrumentation__.Notify(189539)

	return b, nil
}

func rotateGeneratedCerts(ctx context.Context, c base.Config) error {
	__antithesis_instrumentation__.Notify(189550)
	cl := security.MakeCertsLocator(c.SSLCertsDir)

	b, err := collectLocalCABundle(c.SSLCertsDir)
	if err != nil {
		__antithesis_instrumentation__.Notify(189558)
		return errors.Wrap(
			err, "failed to load local CAs for certificate rotation")
	} else {
		__antithesis_instrumentation__.Notify(189559)
	}
	__antithesis_instrumentation__.Notify(189551)

	rpcAddrs := extractHosts(c.Addr, c.AdvertiseAddr)
	sqlAddrs := rpcAddrs
	if c.SplitListenSQL {
		__antithesis_instrumentation__.Notify(189560)
		sqlAddrs = extractHosts(c.SQLAddr, c.SQLAdvertiseAddr)
	} else {
		__antithesis_instrumentation__.Notify(189561)
	}
	__antithesis_instrumentation__.Notify(189552)
	httpAddrs := extractHosts(c.HTTPAddr, c.HTTPAdvertiseAddr)

	if b.InterNode.CACertificate != nil {
		__antithesis_instrumentation__.Notify(189562)
		err = b.InterNode.rotateServiceCert(
			ctx,
			cl.NodeCertPath(),
			cl.NodeKeyPath(),
			defaultCertLifetime,
			security.NodeUser,
			serviceNameInterNode,
			rpcAddrs,
			true,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(189563)
			return errors.Wrap(err, "failed to rotate InterNode cert")
		} else {
			__antithesis_instrumentation__.Notify(189564)
		}
	} else {
		__antithesis_instrumentation__.Notify(189565)
	}
	__antithesis_instrumentation__.Notify(189553)

	if b.UserAuth.CACertificate != nil {
		__antithesis_instrumentation__.Notify(189566)
		err = b.UserAuth.rotateServiceCert(
			ctx,
			cl.ClientNodeCertPath(),
			cl.ClientNodeKeyPath(),
			defaultCertLifetime,
			security.NodeUser,
			serviceNameUserAuth,
			nil,
			true,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(189567)
			return errors.Wrap(err, "failed to rotate InterNode cert")
		} else {
			__antithesis_instrumentation__.Notify(189568)
		}
	} else {
		__antithesis_instrumentation__.Notify(189569)
	}
	__antithesis_instrumentation__.Notify(189554)

	if b.SQLService.CACertificate != nil {
		__antithesis_instrumentation__.Notify(189570)
		err = b.SQLService.rotateServiceCert(
			ctx,
			cl.SQLServiceCertPath(),
			cl.SQLServiceKeyPath(),
			defaultCertLifetime,
			security.NodeUser,
			serviceNameSQL,
			sqlAddrs,
			false,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(189571)
			return errors.Wrap(err, "failed to rotate SQLService cert")
		} else {
			__antithesis_instrumentation__.Notify(189572)
		}
	} else {
		__antithesis_instrumentation__.Notify(189573)
	}
	__antithesis_instrumentation__.Notify(189555)

	if b.RPCService.CACertificate != nil {
		__antithesis_instrumentation__.Notify(189574)
		err = b.RPCService.rotateServiceCert(
			ctx,
			cl.RPCServiceCertPath(),
			cl.RPCServiceKeyPath(),
			defaultCertLifetime,
			security.NodeUser,
			serviceNameRPC,
			rpcAddrs,
			false,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(189575)
			return errors.Wrap(err, "failed to rotate RPCService cert")
		} else {
			__antithesis_instrumentation__.Notify(189576)
		}
	} else {
		__antithesis_instrumentation__.Notify(189577)
	}
	__antithesis_instrumentation__.Notify(189556)

	if b.AdminUIService.CACertificate != nil {
		__antithesis_instrumentation__.Notify(189578)
		err = b.AdminUIService.rotateServiceCert(
			ctx,
			cl.UICertPath(),
			cl.UIKeyPath(),
			defaultCertLifetime,
			httpAddrs[0],
			serviceNameUI,
			httpAddrs,
			false,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(189579)
			return errors.Wrap(err, "failed to rotate AdminUIService cert")
		} else {
			__antithesis_instrumentation__.Notify(189580)
		}
	} else {
		__antithesis_instrumentation__.Notify(189581)
	}
	__antithesis_instrumentation__.Notify(189557)

	return nil
}

func (sb *ServiceCertificateBundle) rotateServiceCert(
	ctx context.Context,
	certPath string,
	keyPath string,
	serviceCertLifespan time.Duration,
	commonName, serviceString string,
	hostnames []string,
	serviceCertIsAlsoValidAsClient bool,
) error {
	__antithesis_instrumentation__.Notify(189582)

	caCertPEM, err := security.PEMToCertificates(sb.CACertificate)
	if err != nil {
		__antithesis_instrumentation__.Notify(189588)
		return errors.Wrap(err, "error when decoding PEM CACertificate")
	} else {
		__antithesis_instrumentation__.Notify(189589)
	}
	__antithesis_instrumentation__.Notify(189583)
	caKeyPEM, rest := pem.Decode(sb.CAKey)
	if len(rest) > 0 || func() bool {
		__antithesis_instrumentation__.Notify(189590)
		return caKeyPEM == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(189591)
		return errors.New("error when decoding PEM CAKey")
	} else {
		__antithesis_instrumentation__.Notify(189592)
	}
	__antithesis_instrumentation__.Notify(189584)
	certPEM, keyPEM, err := security.CreateServiceCertAndKey(
		ctx,
		log.Ops.Infof,
		serviceCertLifespan,
		commonName,
		hostnames,
		caCertPEM[0],
		caKeyPEM,
		serviceCertIsAlsoValidAsClient,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(189593)
		return errors.Wrapf(
			err, "failed to rotate certs for %q", serviceString)
	} else {
		__antithesis_instrumentation__.Notify(189594)
	}
	__antithesis_instrumentation__.Notify(189585)

	if _, err := os.Stat(certPath); err == nil {
		__antithesis_instrumentation__.Notify(189595)
		err = writeCertificateFile(certPath, certPEM, true)
		if err != nil {
			__antithesis_instrumentation__.Notify(189596)
			return errors.Wrapf(
				err, "failed to rotate certs for %q", serviceString)
		} else {
			__antithesis_instrumentation__.Notify(189597)
		}
	} else {
		__antithesis_instrumentation__.Notify(189598)
		return errors.Wrapf(
			err, "failed to rotate certs for %q", serviceString)
	}
	__antithesis_instrumentation__.Notify(189586)

	if _, err := os.Stat(certPath); err == nil {
		__antithesis_instrumentation__.Notify(189599)
		err = writeKeyFile(keyPath, keyPEM, true)
		if err != nil {
			__antithesis_instrumentation__.Notify(189600)
			return errors.Wrapf(
				err, "failed to rotate certs for %q", serviceString)
		} else {
			__antithesis_instrumentation__.Notify(189601)
		}
	} else {
		__antithesis_instrumentation__.Notify(189602)
		return errors.Wrapf(
			err, "failed to rotate certs for %q", serviceString)
	}
	__antithesis_instrumentation__.Notify(189587)

	return nil
}
