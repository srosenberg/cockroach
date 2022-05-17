package pgwire

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"math"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/security/sessionrevival"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/hba"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/identmap"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgwirebase"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/errors"
	"github.com/xdg-go/scram"
)

func loadDefaultMethods() {
	__antithesis_instrumentation__.Notify(559012)

	RegisterAuthMethod("password", authPassword, hba.ConnAny, NoOptionsAllowed)

	RegisterAuthMethod("cert", authCert, hba.ConnHostSSL, nil)

	RegisterAuthMethod("cert-password", authCertPassword, hba.ConnAny, nil)

	RegisterAuthMethod("scram-sha-256", authScram, hba.ConnAny,
		chainOptions(
			requireClusterVersion(clusterversion.SCRAMAuthentication),
			NoOptionsAllowed))

	RegisterAuthMethod("cert-scram-sha-256", authCertScram, hba.ConnAny,
		chainOptions(
			requireClusterVersion(clusterversion.SCRAMAuthentication),
			NoOptionsAllowed))

	RegisterAuthMethod("reject", authReject, hba.ConnAny, NoOptionsAllowed)

	RegisterAuthMethod("trust", authTrust, hba.ConnAny, NoOptionsAllowed)
}

type AuthMethod = func(
	ctx context.Context,
	c AuthConn,
	tlsState tls.ConnectionState,
	execCfg *sql.ExecutorConfig,
	entry *hba.Entry,
	identMap *identmap.Conf,
) (*AuthBehaviors, error)

var _ AuthMethod = authPassword
var _ AuthMethod = authScram
var _ AuthMethod = authCert
var _ AuthMethod = authCertPassword
var _ AuthMethod = authCertScram
var _ AuthMethod = authTrust
var _ AuthMethod = authReject
var _ AuthMethod = authSessionRevivalToken([]byte{})

func authPassword(
	_ context.Context,
	c AuthConn,
	_ tls.ConnectionState,
	execCfg *sql.ExecutorConfig,
	_ *hba.Entry,
	_ *identmap.Conf,
) (*AuthBehaviors, error) {
	__antithesis_instrumentation__.Notify(559013)
	b := &AuthBehaviors{}
	b.SetRoleMapper(UseProvidedIdentity)
	b.SetAuthenticator(func(
		ctx context.Context,
		systemIdentity security.SQLUsername,
		clientConnection bool,
		pwRetrieveFn PasswordRetrievalFn,
	) error {
		__antithesis_instrumentation__.Notify(559015)
		return passwordAuthenticator(ctx, systemIdentity, clientConnection, pwRetrieveFn, c, execCfg)
	})
	__antithesis_instrumentation__.Notify(559014)
	return b, nil
}

var errExpiredPassword = errors.New("password is expired")

func passwordAuthenticator(
	ctx context.Context,
	systemIdentity security.SQLUsername,
	clientConnection bool,
	pwRetrieveFn PasswordRetrievalFn,
	c AuthConn,
	execCfg *sql.ExecutorConfig,
) error {
	__antithesis_instrumentation__.Notify(559016)

	if err := c.SendAuthRequest(authCleartextPassword, nil); err != nil {
		__antithesis_instrumentation__.Notify(559023)
		return err
	} else {
		__antithesis_instrumentation__.Notify(559024)
	}
	__antithesis_instrumentation__.Notify(559017)

	expired, hashedPassword, pwRetrievalErr := pwRetrieveFn(ctx)

	pwdData, err := c.GetPwdData()
	if err != nil {
		__antithesis_instrumentation__.Notify(559025)
		c.LogAuthFailed(ctx, eventpb.AuthFailReason_PRE_HOOK_ERROR, err)
		if pwRetrievalErr != nil {
			__antithesis_instrumentation__.Notify(559027)
			return errors.CombineErrors(err, pwRetrievalErr)
		} else {
			__antithesis_instrumentation__.Notify(559028)
		}
		__antithesis_instrumentation__.Notify(559026)
		return err
	} else {
		__antithesis_instrumentation__.Notify(559029)
	}
	__antithesis_instrumentation__.Notify(559018)

	if pwRetrievalErr != nil {
		__antithesis_instrumentation__.Notify(559030)
		c.LogAuthFailed(ctx, eventpb.AuthFailReason_USER_RETRIEVAL_ERROR, pwRetrievalErr)
		return pwRetrievalErr
	} else {
		__antithesis_instrumentation__.Notify(559031)
	}
	__antithesis_instrumentation__.Notify(559019)

	password, err := passwordString(pwdData)
	if err != nil {
		__antithesis_instrumentation__.Notify(559032)
		c.LogAuthFailed(ctx, eventpb.AuthFailReason_PRE_HOOK_ERROR, err)
		return err
	} else {
		__antithesis_instrumentation__.Notify(559033)
	}
	__antithesis_instrumentation__.Notify(559020)

	if expired {
		__antithesis_instrumentation__.Notify(559034)
		c.LogAuthFailed(ctx, eventpb.AuthFailReason_CREDENTIALS_EXPIRED, nil)
		return errExpiredPassword
	} else {
		__antithesis_instrumentation__.Notify(559035)
		if hashedPassword.Method() == security.HashMissingPassword {
			__antithesis_instrumentation__.Notify(559036)
			c.LogAuthInfof(ctx, "user has no password defined")

		} else {
			__antithesis_instrumentation__.Notify(559037)
		}
	}
	__antithesis_instrumentation__.Notify(559021)

	err = security.UserAuthPasswordHook(
		false, password, hashedPassword,
	)(ctx, systemIdentity, clientConnection)

	if err == nil {
		__antithesis_instrumentation__.Notify(559038)

		sql.MaybeUpgradeStoredPasswordHash(ctx,
			execCfg,
			systemIdentity,
			password, hashedPassword)
	} else {
		__antithesis_instrumentation__.Notify(559039)
	}
	__antithesis_instrumentation__.Notify(559022)

	return err
}

func passwordString(pwdData []byte) (string, error) {
	__antithesis_instrumentation__.Notify(559040)

	if bytes.IndexByte(pwdData, 0) != len(pwdData)-1 {
		__antithesis_instrumentation__.Notify(559042)
		return "", fmt.Errorf("expected 0-terminated byte array")
	} else {
		__antithesis_instrumentation__.Notify(559043)
	}
	__antithesis_instrumentation__.Notify(559041)
	return string(pwdData[:len(pwdData)-1]), nil
}

func authScram(
	ctx context.Context,
	c AuthConn,
	_ tls.ConnectionState,
	execCfg *sql.ExecutorConfig,
	_ *hba.Entry,
	_ *identmap.Conf,
) (*AuthBehaviors, error) {
	__antithesis_instrumentation__.Notify(559044)
	b := &AuthBehaviors{}
	b.SetRoleMapper(UseProvidedIdentity)
	b.SetAuthenticator(func(
		ctx context.Context,
		systemIdentity security.SQLUsername,
		clientConnection bool,
		pwRetrieveFn PasswordRetrievalFn,
	) error {
		__antithesis_instrumentation__.Notify(559046)
		return scramAuthenticator(ctx, systemIdentity, clientConnection, pwRetrieveFn, c, execCfg)
	})
	__antithesis_instrumentation__.Notify(559045)
	return b, nil
}

func scramAuthenticator(
	ctx context.Context,
	systemIdentity security.SQLUsername,
	clientConnection bool,
	pwRetrieveFn PasswordRetrievalFn,
	c AuthConn,
	execCfg *sql.ExecutorConfig,
) error {
	__antithesis_instrumentation__.Notify(559047)

	const supportedMethods = "SCRAM-SHA-256\x00\x00"
	if err := c.SendAuthRequest(authReqSASL, []byte(supportedMethods)); err != nil {
		__antithesis_instrumentation__.Notify(559052)
		return err
	} else {
		__antithesis_instrumentation__.Notify(559053)
	}
	__antithesis_instrumentation__.Notify(559048)

	expired, hashedPassword, pwRetrievalErr := pwRetrieveFn(ctx)

	scramServer, _ := scram.SHA256.NewServer(func(user string) (creds scram.StoredCredentials, err error) {
		__antithesis_instrumentation__.Notify(559054)

		if expired {
			__antithesis_instrumentation__.Notify(559057)
			c.LogAuthFailed(ctx, eventpb.AuthFailReason_CREDENTIALS_EXPIRED, nil)
			return creds, errExpiredPassword
		} else {
			__antithesis_instrumentation__.Notify(559058)
			if hashedPassword.Method() != security.HashSCRAMSHA256 {
				__antithesis_instrumentation__.Notify(559059)
				const credentialsNotSCRAM = "user password hash not in SCRAM format"
				c.LogAuthInfof(ctx, credentialsNotSCRAM)
				return creds, errors.New(credentialsNotSCRAM)
			} else {
				__antithesis_instrumentation__.Notify(559060)
			}
		}
		__antithesis_instrumentation__.Notify(559055)

		ok, creds := security.GetSCRAMStoredCredentials(hashedPassword)
		if !ok {
			__antithesis_instrumentation__.Notify(559061)
			return creds, errors.AssertionFailedf("programming error: hash method is SCRAM but no stored credentials")
		} else {
			__antithesis_instrumentation__.Notify(559062)
		}
		__antithesis_instrumentation__.Notify(559056)
		return creds, nil
	})
	__antithesis_instrumentation__.Notify(559049)

	handshake := scramServer.NewConversation()

	initial := true
	for {
		__antithesis_instrumentation__.Notify(559063)
		if handshake.Done() {
			__antithesis_instrumentation__.Notify(559070)
			break
		} else {
			__antithesis_instrumentation__.Notify(559071)
		}
		__antithesis_instrumentation__.Notify(559064)

		resp, err := c.GetPwdData()
		if err != nil {
			__antithesis_instrumentation__.Notify(559072)
			c.LogAuthFailed(ctx, eventpb.AuthFailReason_PRE_HOOK_ERROR, err)
			if pwRetrievalErr != nil {
				__antithesis_instrumentation__.Notify(559074)
				return errors.CombineErrors(err, pwRetrievalErr)
			} else {
				__antithesis_instrumentation__.Notify(559075)
			}
			__antithesis_instrumentation__.Notify(559073)
			return err
		} else {
			__antithesis_instrumentation__.Notify(559076)
		}
		__antithesis_instrumentation__.Notify(559065)

		if pwRetrievalErr != nil {
			__antithesis_instrumentation__.Notify(559077)
			c.LogAuthFailed(ctx, eventpb.AuthFailReason_USER_RETRIEVAL_ERROR, pwRetrievalErr)
			return pwRetrievalErr
		} else {
			__antithesis_instrumentation__.Notify(559078)
		}
		__antithesis_instrumentation__.Notify(559066)

		var input []byte
		if initial {
			__antithesis_instrumentation__.Notify(559079)

			rb := pgwirebase.ReadBuffer{Msg: resp}
			reqMethod, err := rb.GetString()
			if err != nil {
				__antithesis_instrumentation__.Notify(559084)
				c.LogAuthFailed(ctx, eventpb.AuthFailReason_PRE_HOOK_ERROR, err)
				return err
			} else {
				__antithesis_instrumentation__.Notify(559085)
			}
			__antithesis_instrumentation__.Notify(559080)
			if reqMethod != "SCRAM-SHA-256" {
				__antithesis_instrumentation__.Notify(559086)
				c.LogAuthInfof(ctx, "client requests unknown scram method %q", reqMethod)
				err := unimplemented.NewWithIssue(74300, "channel binding not supported")

				sqltelemetry.RecordError(ctx, err, &execCfg.Settings.SV)
				return err
			} else {
				__antithesis_instrumentation__.Notify(559087)
			}
			__antithesis_instrumentation__.Notify(559081)
			inputLen, err := rb.GetUint32()
			if err != nil {
				__antithesis_instrumentation__.Notify(559088)
				c.LogAuthFailed(ctx, eventpb.AuthFailReason_PRE_HOOK_ERROR, err)
				return err
			} else {
				__antithesis_instrumentation__.Notify(559089)
			}
			__antithesis_instrumentation__.Notify(559082)

			if inputLen < math.MaxUint32 {
				__antithesis_instrumentation__.Notify(559090)
				input, err = rb.GetBytes(int(inputLen))
				if err != nil {
					__antithesis_instrumentation__.Notify(559091)
					c.LogAuthFailed(ctx, eventpb.AuthFailReason_PRE_HOOK_ERROR, err)
					return err
				} else {
					__antithesis_instrumentation__.Notify(559092)
				}
			} else {
				__antithesis_instrumentation__.Notify(559093)
			}
			__antithesis_instrumentation__.Notify(559083)
			initial = false
		} else {
			__antithesis_instrumentation__.Notify(559094)
			input = resp
		}
		__antithesis_instrumentation__.Notify(559067)

		got, err := handshake.Step(string(input))
		if err != nil {
			__antithesis_instrumentation__.Notify(559095)
			c.LogAuthInfof(ctx, "scram handshake error: %v", err)
			break
		} else {
			__antithesis_instrumentation__.Notify(559096)
		}
		__antithesis_instrumentation__.Notify(559068)

		reqType := authReqSASLContinue
		if handshake.Done() {
			__antithesis_instrumentation__.Notify(559097)

			reqType = authReqSASLFin
		} else {
			__antithesis_instrumentation__.Notify(559098)
		}
		__antithesis_instrumentation__.Notify(559069)

		if err := c.SendAuthRequest(reqType, []byte(got)); err != nil {
			__antithesis_instrumentation__.Notify(559099)
			return err
		} else {
			__antithesis_instrumentation__.Notify(559100)
		}
	}
	__antithesis_instrumentation__.Notify(559050)

	if !handshake.Valid() {
		__antithesis_instrumentation__.Notify(559101)
		return security.NewErrPasswordUserAuthFailed(systemIdentity)
	} else {
		__antithesis_instrumentation__.Notify(559102)
	}
	__antithesis_instrumentation__.Notify(559051)

	return nil
}

func authCert(
	_ context.Context,
	_ AuthConn,
	tlsState tls.ConnectionState,
	_ *sql.ExecutorConfig,
	hbaEntry *hba.Entry,
	identMap *identmap.Conf,
) (*AuthBehaviors, error) {
	__antithesis_instrumentation__.Notify(559103)
	b := &AuthBehaviors{}
	b.SetRoleMapper(HbaMapper(hbaEntry, identMap))
	b.SetAuthenticator(func(
		ctx context.Context,
		systemIdentity security.SQLUsername,
		clientConnection bool,
		pwRetrieveFn PasswordRetrievalFn,
	) error {
		__antithesis_instrumentation__.Notify(559105)
		if len(tlsState.PeerCertificates) == 0 {
			__antithesis_instrumentation__.Notify(559108)
			return errors.New("no TLS peer certificates, but required for auth")
		} else {
			__antithesis_instrumentation__.Notify(559109)
		}
		__antithesis_instrumentation__.Notify(559106)

		tlsState.PeerCertificates[0].Subject.CommonName = tree.Name(
			tlsState.PeerCertificates[0].Subject.CommonName,
		).Normalize()
		hook, err := security.UserAuthCertHook(false, &tlsState)
		if err != nil {
			__antithesis_instrumentation__.Notify(559110)
			return err
		} else {
			__antithesis_instrumentation__.Notify(559111)
		}
		__antithesis_instrumentation__.Notify(559107)
		return hook(ctx, systemIdentity, clientConnection)
	})
	__antithesis_instrumentation__.Notify(559104)
	return b, nil
}

func authCertPassword(
	ctx context.Context,
	c AuthConn,
	tlsState tls.ConnectionState,
	execCfg *sql.ExecutorConfig,
	entry *hba.Entry,
	identMap *identmap.Conf,
) (*AuthBehaviors, error) {
	__antithesis_instrumentation__.Notify(559112)
	var fn AuthMethod
	if len(tlsState.PeerCertificates) == 0 {
		__antithesis_instrumentation__.Notify(559114)
		c.LogAuthInfof(ctx, "client did not present TLS certificate")
		if AutoSelectPasswordAuth.Get(&execCfg.Settings.SV) {
			__antithesis_instrumentation__.Notify(559115)

			fn = authAutoSelectPasswordProtocol
		} else {
			__antithesis_instrumentation__.Notify(559116)
			c.LogAuthInfof(ctx, "proceeding with password authentication")
			fn = authPassword
		}
	} else {
		__antithesis_instrumentation__.Notify(559117)
		c.LogAuthInfof(ctx, "client presented certificate, proceeding with certificate validation")
		fn = authCert
	}
	__antithesis_instrumentation__.Notify(559113)
	return fn(ctx, c, tlsState, execCfg, entry, identMap)
}

var AutoSelectPasswordAuth = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"server.user_login.cert_password_method.auto_scram_promotion.enabled",
	"whether to automatically promote cert-password authentication to use SCRAM",
	true,
).WithPublic()

func authAutoSelectPasswordProtocol(
	_ context.Context,
	c AuthConn,
	_ tls.ConnectionState,
	execCfg *sql.ExecutorConfig,
	_ *hba.Entry,
	_ *identmap.Conf,
) (*AuthBehaviors, error) {
	__antithesis_instrumentation__.Notify(559118)
	b := &AuthBehaviors{}
	b.SetRoleMapper(UseProvidedIdentity)
	b.SetAuthenticator(func(
		ctx context.Context,
		systemIdentity security.SQLUsername,
		clientConnection bool,
		pwRetrieveFn PasswordRetrievalFn,
	) error {
		__antithesis_instrumentation__.Notify(559120)

		expired, hashedPassword, err := pwRetrieveFn(ctx)

		newpwfn := func(ctx context.Context) (bool, security.PasswordHash, error) {
			__antithesis_instrumentation__.Notify(559123)
			return expired, hashedPassword, err
		}
		__antithesis_instrumentation__.Notify(559121)

		if err == nil && func() bool {
			__antithesis_instrumentation__.Notify(559124)
			return hashedPassword.Method() == security.HashBCrypt == true
		}() == true {
			__antithesis_instrumentation__.Notify(559125)

			c.LogAuthInfof(ctx, "found stored crdb-bcrypt credentials; requesting cleartext password")
			return passwordAuthenticator(ctx, systemIdentity, clientConnection, newpwfn, c, execCfg)
		} else {
			__antithesis_instrumentation__.Notify(559126)
		}
		__antithesis_instrumentation__.Notify(559122)

		c.LogAuthInfof(ctx, "no crdb-bcrypt credentials found; proceeding with SCRAM-SHA-256")
		return scramAuthenticator(ctx, systemIdentity, clientConnection, newpwfn, c, execCfg)
	})
	__antithesis_instrumentation__.Notify(559119)
	return b, nil
}

func authCertScram(
	ctx context.Context,
	c AuthConn,
	tlsState tls.ConnectionState,
	execCfg *sql.ExecutorConfig,
	entry *hba.Entry,
	identMap *identmap.Conf,
) (*AuthBehaviors, error) {
	__antithesis_instrumentation__.Notify(559127)
	var fn AuthMethod
	if len(tlsState.PeerCertificates) == 0 {
		__antithesis_instrumentation__.Notify(559129)
		c.LogAuthInfof(ctx, "no client certificate, proceeding with SCRAM authentication")
		fn = authScram
	} else {
		__antithesis_instrumentation__.Notify(559130)
		c.LogAuthInfof(ctx, "client presented certificate, proceeding with certificate validation")
		fn = authCert
	}
	__antithesis_instrumentation__.Notify(559128)
	return fn(ctx, c, tlsState, execCfg, entry, identMap)
}

func authTrust(
	_ context.Context,
	_ AuthConn,
	_ tls.ConnectionState,
	_ *sql.ExecutorConfig,
	_ *hba.Entry,
	_ *identmap.Conf,
) (*AuthBehaviors, error) {
	__antithesis_instrumentation__.Notify(559131)
	b := &AuthBehaviors{}
	b.SetRoleMapper(UseProvidedIdentity)
	b.SetAuthenticator(func(_ context.Context, _ security.SQLUsername, _ bool, _ PasswordRetrievalFn) error {
		__antithesis_instrumentation__.Notify(559133)
		return nil
	})
	__antithesis_instrumentation__.Notify(559132)
	return b, nil
}

func authReject(
	_ context.Context,
	_ AuthConn,
	_ tls.ConnectionState,
	_ *sql.ExecutorConfig,
	_ *hba.Entry,
	_ *identmap.Conf,
) (*AuthBehaviors, error) {
	__antithesis_instrumentation__.Notify(559134)
	b := &AuthBehaviors{}
	b.SetRoleMapper(UseProvidedIdentity)
	b.SetAuthenticator(func(_ context.Context, _ security.SQLUsername, _ bool, _ PasswordRetrievalFn) error {
		__antithesis_instrumentation__.Notify(559136)
		return errors.New("authentication rejected by configuration")
	})
	__antithesis_instrumentation__.Notify(559135)
	return b, nil
}

func authSessionRevivalToken(token []byte) AuthMethod {
	__antithesis_instrumentation__.Notify(559137)
	return func(
		_ context.Context,
		c AuthConn,
		_ tls.ConnectionState,
		execCfg *sql.ExecutorConfig,
		_ *hba.Entry,
		_ *identmap.Conf,
	) (*AuthBehaviors, error) {
		__antithesis_instrumentation__.Notify(559138)
		b := &AuthBehaviors{}
		b.SetRoleMapper(UseProvidedIdentity)
		b.SetAuthenticator(func(ctx context.Context, user security.SQLUsername, _ bool, _ PasswordRetrievalFn) error {
			__antithesis_instrumentation__.Notify(559140)
			c.LogAuthInfof(ctx, "session revival token detected; attempting to use it")
			if !sql.AllowSessionRevival.Get(&execCfg.Settings.SV) || func() bool {
				__antithesis_instrumentation__.Notify(559144)
				return execCfg.Codec.ForSystemTenant() == true
			}() == true {
				__antithesis_instrumentation__.Notify(559145)
				return errors.New("session revival tokens are not supported on this cluster")
			} else {
				__antithesis_instrumentation__.Notify(559146)
			}
			__antithesis_instrumentation__.Notify(559141)
			cm, err := execCfg.RPCContext.SecurityContext.GetCertificateManager()
			if err != nil {
				__antithesis_instrumentation__.Notify(559147)
				return err
			} else {
				__antithesis_instrumentation__.Notify(559148)
			}
			__antithesis_instrumentation__.Notify(559142)
			if err := sessionrevival.ValidateSessionRevivalToken(cm, user, token); err != nil {
				__antithesis_instrumentation__.Notify(559149)
				return errors.Wrap(err, "invalid session revival token")
			} else {
				__antithesis_instrumentation__.Notify(559150)
			}
			__antithesis_instrumentation__.Notify(559143)
			return nil
		})
		__antithesis_instrumentation__.Notify(559139)
		return b, nil
	}
}
