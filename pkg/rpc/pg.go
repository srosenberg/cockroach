package rpc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"net/url"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/pgurl"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/util/netutil/addr"
	"github.com/cockroachdb/errors"
)

func (ctx *SecurityContext) LoadSecurityOptions(u *pgurl.URL, username security.SQLUsername) error {
	__antithesis_instrumentation__.Notify(185493)
	u.WithUsername(username.Normalized())
	if ctx.config.Insecure {
		__antithesis_instrumentation__.Notify(185495)
		u.WithInsecure()
	} else {
		__antithesis_instrumentation__.Notify(185496)
		if net, _, _ := u.GetNetworking(); net == pgurl.ProtoTCP {
			__antithesis_instrumentation__.Notify(185497)
			tlsUsed, tlsMode, caCertPath := u.GetTLSOptions()
			if !tlsUsed {
				__antithesis_instrumentation__.Notify(185503)

				return nil
			} else {
				__antithesis_instrumentation__.Notify(185504)
			}
			__antithesis_instrumentation__.Notify(185498)

			if tlsMode == pgurl.TLSUnspecified {
				__antithesis_instrumentation__.Notify(185505)
				tlsMode = pgurl.TLSVerifyFull
			} else {
				__antithesis_instrumentation__.Notify(185506)
			}
			__antithesis_instrumentation__.Notify(185499)

			if tlsMode == pgurl.TLSVerifyFull || func() bool {
				__antithesis_instrumentation__.Notify(185507)
				return tlsMode == pgurl.TLSVerifyCA == true
			}() == true {
				__antithesis_instrumentation__.Notify(185508)
				if caCertPath == "" {
					__antithesis_instrumentation__.Notify(185509)

					cm, err := ctx.GetCertificateManager()
					if err != nil {
						__antithesis_instrumentation__.Notify(185511)

						if !errors.Is(err, errNoCertificatesFound) {
							__antithesis_instrumentation__.Notify(185512)

							return err
						} else {
							__antithesis_instrumentation__.Notify(185513)
						}

					} else {
						__antithesis_instrumentation__.Notify(185514)
					}
					__antithesis_instrumentation__.Notify(185510)
					if ourCACert := cm.CACert(); ourCACert != nil {
						__antithesis_instrumentation__.Notify(185515)

						caCertPath = cm.FullPath(ourCACert)
					} else {
						__antithesis_instrumentation__.Notify(185516)
					}
				} else {
					__antithesis_instrumentation__.Notify(185517)
				}

			} else {
				__antithesis_instrumentation__.Notify(185518)
			}
			__antithesis_instrumentation__.Notify(185500)

			u.WithTransport(pgurl.TransportTLS(tlsMode, caCertPath))

			var missing bool
			loader := security.GetAssetLoader()

			certPath := ctx.ClientCertPath(username)
			keyPath := ctx.ClientKeyPath(username)
			_, err1 := loader.Stat(certPath)
			_, err2 := loader.Stat(keyPath)
			if err1 != nil || func() bool {
				__antithesis_instrumentation__.Notify(185519)
				return err2 != nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(185520)
				missing = true
			} else {
				__antithesis_instrumentation__.Notify(185521)
			}
			__antithesis_instrumentation__.Notify(185501)

			if missing && func() bool {
				__antithesis_instrumentation__.Notify(185522)
				return username.IsNodeUser() == true
			}() == true {
				__antithesis_instrumentation__.Notify(185523)
				missing = false
				certPath = ctx.NodeCertPath()
				keyPath = ctx.NodeKeyPath()
				_, err1 = loader.Stat(certPath)
				_, err2 = loader.Stat(keyPath)
				if err1 != nil || func() bool {
					__antithesis_instrumentation__.Notify(185524)
					return err2 != nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(185525)
					missing = true
				} else {
					__antithesis_instrumentation__.Notify(185526)
				}
			} else {
				__antithesis_instrumentation__.Notify(185527)
			}
			__antithesis_instrumentation__.Notify(185502)

			if !missing {
				__antithesis_instrumentation__.Notify(185528)
				pwEnabled, hasPw, pwd := u.GetAuthnPassword()
				if !pwEnabled {
					__antithesis_instrumentation__.Notify(185529)
					u.WithAuthn(pgurl.AuthnClientCert(certPath, keyPath))
				} else {
					__antithesis_instrumentation__.Notify(185530)
					u.WithAuthn(pgurl.AuthnPasswordAndCert(certPath, keyPath, hasPw, pwd))
				}
			} else {
				__antithesis_instrumentation__.Notify(185531)
			}
		} else {
			__antithesis_instrumentation__.Notify(185532)
		}
	}
	__antithesis_instrumentation__.Notify(185494)
	return nil
}

func (ctx *SecurityContext) PGURL(user *url.Userinfo) (*pgurl.URL, error) {
	__antithesis_instrumentation__.Notify(185533)
	host, port, _ := addr.SplitHostPort(ctx.config.SQLAdvertiseAddr, base.DefaultPort)
	u := pgurl.New().
		WithNet(pgurl.NetTCP(host, port)).
		WithDatabase(catalogkeys.DefaultDatabaseName)

	username, _ := security.MakeSQLUsernameFromUserInput(user.Username(), security.UsernameValidation)
	if err := ctx.LoadSecurityOptions(u, username); err != nil {
		__antithesis_instrumentation__.Notify(185535)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(185536)
	}
	__antithesis_instrumentation__.Notify(185534)
	return u, nil
}
