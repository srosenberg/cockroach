package rpc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/severity"
	"github.com/cockroachdb/errors"
)

type lazyHTTPClient struct {
	sync.Once
	httpClient http.Client
	err        error
}

type lazyCertificateManager struct {
	sync.Once
	cm  *security.CertificateManager
	err error
}

func wrapError(err error) error {
	__antithesis_instrumentation__.Notify(185577)
	if !errors.HasType(err, (*security.Error)(nil)) {
		__antithesis_instrumentation__.Notify(185579)
		return &security.Error{
			Message: "problem using security settings",
			Err:     err,
		}
	} else {
		__antithesis_instrumentation__.Notify(185580)
	}
	__antithesis_instrumentation__.Notify(185578)
	return err
}

type SecurityContext struct {
	security.CertsLocator
	security.TLSSettings
	config *base.Config
	tenID  roachpb.TenantID
	lazy   struct {
		certificateManager lazyCertificateManager

		httpClient lazyHTTPClient
	}
}

func MakeSecurityContext(
	cfg *base.Config, tlsSettings security.TLSSettings, tenID roachpb.TenantID,
) SecurityContext {
	__antithesis_instrumentation__.Notify(185581)
	if tenID.ToUint64() == 0 {
		__antithesis_instrumentation__.Notify(185583)
		panic(errors.AssertionFailedf("programming error: tenant ID not defined"))
	} else {
		__antithesis_instrumentation__.Notify(185584)
	}
	__antithesis_instrumentation__.Notify(185582)
	return SecurityContext{
		CertsLocator: security.MakeCertsLocator(cfg.SSLCertsDir),
		TLSSettings:  tlsSettings,
		config:       cfg,
		tenID:        tenID,
	}
}

func (ctx *SecurityContext) GetCertificateManager() (*security.CertificateManager, error) {
	__antithesis_instrumentation__.Notify(185585)
	ctx.lazy.certificateManager.Do(func() {
		__antithesis_instrumentation__.Notify(185587)
		var opts []security.Option
		if ctx.tenID != roachpb.SystemTenantID {
			__antithesis_instrumentation__.Notify(185589)
			opts = append(opts, security.ForTenant(ctx.tenID.ToUint64()))
		} else {
			__antithesis_instrumentation__.Notify(185590)
		}
		__antithesis_instrumentation__.Notify(185588)
		ctx.lazy.certificateManager.cm, ctx.lazy.certificateManager.err =
			security.NewCertificateManager(ctx.config.SSLCertsDir, ctx, opts...)

		if ctx.lazy.certificateManager.err == nil && func() bool {
			__antithesis_instrumentation__.Notify(185591)
			return !ctx.config.Insecure == true
		}() == true {
			__antithesis_instrumentation__.Notify(185592)
			infos, err := ctx.lazy.certificateManager.cm.ListCertificates()
			if err != nil {
				__antithesis_instrumentation__.Notify(185593)
				ctx.lazy.certificateManager.err = err
			} else {
				__antithesis_instrumentation__.Notify(185594)
				if len(infos) == 0 {
					__antithesis_instrumentation__.Notify(185595)

					ctx.lazy.certificateManager.err = errNoCertificatesFound
				} else {
					__antithesis_instrumentation__.Notify(185596)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(185597)
		}
	})
	__antithesis_instrumentation__.Notify(185586)
	return ctx.lazy.certificateManager.cm, ctx.lazy.certificateManager.err
}

var errNoCertificatesFound = errors.New("no certificates found; does certs dir exist?")

func (ctx *SecurityContext) GetServerTLSConfig() (*tls.Config, error) {
	__antithesis_instrumentation__.Notify(185598)

	if ctx.config.Insecure {
		__antithesis_instrumentation__.Notify(185602)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(185603)
	}
	__antithesis_instrumentation__.Notify(185599)

	cm, err := ctx.GetCertificateManager()
	if err != nil {
		__antithesis_instrumentation__.Notify(185604)
		return nil, wrapError(err)
	} else {
		__antithesis_instrumentation__.Notify(185605)
	}
	__antithesis_instrumentation__.Notify(185600)

	tlsCfg, err := cm.GetServerTLSConfig()
	if err != nil {
		__antithesis_instrumentation__.Notify(185606)
		return nil, wrapError(err)
	} else {
		__antithesis_instrumentation__.Notify(185607)
	}
	__antithesis_instrumentation__.Notify(185601)
	return tlsCfg, nil
}

func (ctx *SecurityContext) GetClientTLSConfig() (*tls.Config, error) {
	__antithesis_instrumentation__.Notify(185608)

	if ctx.config.Insecure {
		__antithesis_instrumentation__.Notify(185612)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(185613)
	}
	__antithesis_instrumentation__.Notify(185609)

	cm, err := ctx.GetCertificateManager()
	if err != nil {
		__antithesis_instrumentation__.Notify(185614)
		return nil, wrapError(err)
	} else {
		__antithesis_instrumentation__.Notify(185615)
	}
	__antithesis_instrumentation__.Notify(185610)

	tlsCfg, err := cm.GetClientTLSConfig(ctx.config.User)
	if err != nil {
		__antithesis_instrumentation__.Notify(185616)
		return nil, wrapError(err)
	} else {
		__antithesis_instrumentation__.Notify(185617)
	}
	__antithesis_instrumentation__.Notify(185611)
	return tlsCfg, nil
}

func (ctx *SecurityContext) GetTenantTLSConfig() (*tls.Config, error) {
	__antithesis_instrumentation__.Notify(185618)

	if ctx.config.Insecure {
		__antithesis_instrumentation__.Notify(185622)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(185623)
	}
	__antithesis_instrumentation__.Notify(185619)

	cm, err := ctx.GetCertificateManager()
	if err != nil {
		__antithesis_instrumentation__.Notify(185624)
		return nil, wrapError(err)
	} else {
		__antithesis_instrumentation__.Notify(185625)
	}
	__antithesis_instrumentation__.Notify(185620)

	tlsCfg, err := cm.GetTenantTLSConfig()
	if err != nil {
		__antithesis_instrumentation__.Notify(185626)
		return nil, wrapError(err)
	} else {
		__antithesis_instrumentation__.Notify(185627)
	}
	__antithesis_instrumentation__.Notify(185621)
	return tlsCfg, nil
}

func (ctx *SecurityContext) getUIClientTLSConfig() (*tls.Config, error) {
	__antithesis_instrumentation__.Notify(185628)

	if ctx.config.Insecure {
		__antithesis_instrumentation__.Notify(185632)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(185633)
	}
	__antithesis_instrumentation__.Notify(185629)

	cm, err := ctx.GetCertificateManager()
	if err != nil {
		__antithesis_instrumentation__.Notify(185634)
		return nil, wrapError(err)
	} else {
		__antithesis_instrumentation__.Notify(185635)
	}
	__antithesis_instrumentation__.Notify(185630)

	tlsCfg, err := cm.GetUIClientTLSConfig()
	if err != nil {
		__antithesis_instrumentation__.Notify(185636)
		return nil, wrapError(err)
	} else {
		__antithesis_instrumentation__.Notify(185637)
	}
	__antithesis_instrumentation__.Notify(185631)
	return tlsCfg, nil
}

func (ctx *SecurityContext) GetUIServerTLSConfig() (*tls.Config, error) {
	__antithesis_instrumentation__.Notify(185638)

	if ctx.config.Insecure || func() bool {
		__antithesis_instrumentation__.Notify(185642)
		return ctx.config.DisableTLSForHTTP == true
	}() == true {
		__antithesis_instrumentation__.Notify(185643)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(185644)
	}
	__antithesis_instrumentation__.Notify(185639)

	cm, err := ctx.GetCertificateManager()
	if err != nil {
		__antithesis_instrumentation__.Notify(185645)
		return nil, wrapError(err)
	} else {
		__antithesis_instrumentation__.Notify(185646)
	}
	__antithesis_instrumentation__.Notify(185640)

	tlsCfg, err := cm.GetUIServerTLSConfig()
	if err != nil {
		__antithesis_instrumentation__.Notify(185647)
		return nil, wrapError(err)
	} else {
		__antithesis_instrumentation__.Notify(185648)
	}
	__antithesis_instrumentation__.Notify(185641)
	return tlsCfg, nil
}

func (ctx *SecurityContext) GetHTTPClient() (http.Client, error) {
	__antithesis_instrumentation__.Notify(185649)
	ctx.lazy.httpClient.Do(func() {
		__antithesis_instrumentation__.Notify(185651)
		ctx.lazy.httpClient.httpClient.Timeout = 10 * time.Second
		var transport http.Transport
		ctx.lazy.httpClient.httpClient.Transport = &transport
		transport.TLSClientConfig, ctx.lazy.httpClient.err = ctx.getUIClientTLSConfig()
	})
	__antithesis_instrumentation__.Notify(185650)

	return ctx.lazy.httpClient.httpClient, ctx.lazy.httpClient.err
}

func (ctx *SecurityContext) CheckCertificateAddrs(cctx context.Context) {
	__antithesis_instrumentation__.Notify(185652)
	if ctx.config.Insecure {
		__antithesis_instrumentation__.Notify(185656)
		return
	} else {
		__antithesis_instrumentation__.Notify(185657)
	}
	__antithesis_instrumentation__.Notify(185653)

	cm, _ := ctx.GetCertificateManager()

	certInfo := cm.NodeCert()
	if certInfo.Error != nil {
		__antithesis_instrumentation__.Notify(185658)
		log.Ops.Shoutf(cctx, severity.ERROR,
			"invalid node certificate: %v", certInfo.Error)
	} else {
		__antithesis_instrumentation__.Notify(185659)
		cert := certInfo.ParsedCertificates[0]
		addrInfo := certAddrs(cert)

		log.Ops.Infof(cctx, "server certificate addresses: %s", addrInfo)

		var msg bytes.Buffer

		host, _, err := net.SplitHostPort(ctx.config.AdvertiseAddr)
		if err != nil {
			__antithesis_instrumentation__.Notify(185664)
			panic(errors.AssertionFailedf("programming error: call ValidateAddrs() first"))
		} else {
			__antithesis_instrumentation__.Notify(185665)
		}
		__antithesis_instrumentation__.Notify(185660)
		if err := cert.VerifyHostname(host); err != nil {
			__antithesis_instrumentation__.Notify(185666)
			fmt.Fprintf(&msg, "advertise address %q not in node certificate (%s)\n", host, addrInfo)
		} else {
			__antithesis_instrumentation__.Notify(185667)
		}
		__antithesis_instrumentation__.Notify(185661)
		host, _, err = net.SplitHostPort(ctx.config.SQLAdvertiseAddr)
		if err != nil {
			__antithesis_instrumentation__.Notify(185668)
			panic(errors.AssertionFailedf("programming error: call ValidateAddrs() first"))
		} else {
			__antithesis_instrumentation__.Notify(185669)
		}
		__antithesis_instrumentation__.Notify(185662)
		if err := cert.VerifyHostname(host); err != nil {
			__antithesis_instrumentation__.Notify(185670)
			fmt.Fprintf(&msg, "advertise SQL address %q not in node certificate (%s)\n", host, addrInfo)
		} else {
			__antithesis_instrumentation__.Notify(185671)
		}
		__antithesis_instrumentation__.Notify(185663)
		if msg.Len() > 0 {
			__antithesis_instrumentation__.Notify(185672)
			log.Ops.Shoutf(cctx, severity.WARNING,
				"%s"+
					"Secure client connections are likely to fail.\n"+
					"Consider extending the node certificate or tweak --listen-addr/--advertise-addr/--sql-addr/--advertise-sql-addr.",
				msg.String())
		} else {
			__antithesis_instrumentation__.Notify(185673)
		}
	}
	__antithesis_instrumentation__.Notify(185654)

	certInfo = cm.UICert()
	if certInfo == nil {
		__antithesis_instrumentation__.Notify(185674)

		certInfo = cm.NodeCert()
	} else {
		__antithesis_instrumentation__.Notify(185675)
	}
	__antithesis_instrumentation__.Notify(185655)
	if certInfo.Error != nil {
		__antithesis_instrumentation__.Notify(185676)
		log.Ops.Shoutf(cctx, severity.ERROR,
			"invalid UI certificate: %v", certInfo.Error)
	} else {
		__antithesis_instrumentation__.Notify(185677)
		cert := certInfo.ParsedCertificates[0]
		addrInfo := certAddrs(cert)

		log.Ops.Infof(cctx, "web UI certificate addresses: %s", addrInfo)
	}
}

func (ctx *SecurityContext) HTTPRequestScheme() string {
	__antithesis_instrumentation__.Notify(185678)
	return ctx.config.HTTPRequestScheme()
}

func certAddrs(cert *x509.Certificate) string {
	__antithesis_instrumentation__.Notify(185679)

	addrs := make([]string, len(cert.IPAddresses))
	for i, ip := range cert.IPAddresses {
		__antithesis_instrumentation__.Notify(185681)
		addrs[i] = ip.String()
	}
	__antithesis_instrumentation__.Notify(185680)

	return fmt.Sprintf("IP=%s; DNS=%s; CN=%s",
		strings.Join(addrs, ","),
		strings.Join(cert.DNSNames, ","),
		cert.Subject.CommonName)
}
