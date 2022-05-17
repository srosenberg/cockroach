//go:build gss
// +build gss

package gssapiccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"crypto/tls"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/ccl/utilccl"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/hba"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/identmap"
	"github.com/cockroachdb/errors"
)

const (
	authTypeGSS         int32 = 7
	authTypeGSSContinue int32 = 8
)

func authGSS(
	_ context.Context,
	c pgwire.AuthConn,
	_ tls.ConnectionState,
	execCfg *sql.ExecutorConfig,
	entry *hba.Entry,
	identMap *identmap.Conf,
) (*pgwire.AuthBehaviors, error) {
	__antithesis_instrumentation__.Notify(19544)
	behaviors := &pgwire.AuthBehaviors{}

	connClose, gssUser, err := getGssUser(c)
	behaviors.SetConnClose(connClose)
	if err != nil {
		__antithesis_instrumentation__.Notify(19549)
		return behaviors, err
	} else {
		__antithesis_instrumentation__.Notify(19550)
	}
	__antithesis_instrumentation__.Notify(19545)

	if u, err := security.MakeSQLUsernameFromUserInput(gssUser, security.UsernameValidation); err != nil {
		__antithesis_instrumentation__.Notify(19551)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(19552)
		behaviors.SetReplacementIdentity(u)
	}
	__antithesis_instrumentation__.Notify(19546)

	include0 := entry.GetOption("include_realm") == "0"
	if entry.GetOption("map") != "" {
		__antithesis_instrumentation__.Notify(19553)
		mapper := pgwire.HbaMapper(entry, identMap)

		if include0 {
			__antithesis_instrumentation__.Notify(19555)
			mapper = stripAndDelegateMapper(mapper)
		} else {
			__antithesis_instrumentation__.Notify(19556)
		}
		__antithesis_instrumentation__.Notify(19554)
		behaviors.SetRoleMapper(mapper)
	} else {
		__antithesis_instrumentation__.Notify(19557)
		if include0 {
			__antithesis_instrumentation__.Notify(19558)

			behaviors.SetRoleMapper(stripRealmMapper)
		} else {
			__antithesis_instrumentation__.Notify(19559)
			return nil, errors.New("unsupported HBA entry configuration")
		}
	}
	__antithesis_instrumentation__.Notify(19547)

	behaviors.SetAuthenticator(func(
		_ context.Context, _ security.SQLUsername, _ bool, _ pgwire.PasswordRetrievalFn,
	) error {
		__antithesis_instrumentation__.Notify(19560)

		if realms := entry.GetOptions("krb_realm"); len(realms) > 0 {
			__antithesis_instrumentation__.Notify(19562)
			if idx := strings.IndexByte(gssUser, '@'); idx >= 0 {
				__antithesis_instrumentation__.Notify(19563)
				realm := gssUser[idx+1:]
				matched := false
				for _, krbRealm := range realms {
					__antithesis_instrumentation__.Notify(19565)
					if realm == krbRealm {
						__antithesis_instrumentation__.Notify(19566)
						matched = true
						break
					} else {
						__antithesis_instrumentation__.Notify(19567)
					}
				}
				__antithesis_instrumentation__.Notify(19564)
				if !matched {
					__antithesis_instrumentation__.Notify(19568)
					return errors.Errorf("GSSAPI realm (%s) didn't match any configured realm", realm)
				} else {
					__antithesis_instrumentation__.Notify(19569)
				}
			} else {
				__antithesis_instrumentation__.Notify(19570)
				return errors.New("GSSAPI did not return realm but realm matching was requested")
			}
		} else {
			__antithesis_instrumentation__.Notify(19571)
		}
		__antithesis_instrumentation__.Notify(19561)

		return utilccl.CheckEnterpriseEnabled(execCfg.Settings, execCfg.LogicalClusterID(), execCfg.Organization(), "GSS authentication")
	})
	__antithesis_instrumentation__.Notify(19548)
	return behaviors, nil
}

func checkEntry(_ *settings.Values, entry hba.Entry) error {
	__antithesis_instrumentation__.Notify(19572)
	hasInclude0 := false
	hasMap := false
	for _, op := range entry.Options {
		__antithesis_instrumentation__.Notify(19575)
		switch op[0] {
		case "include_realm":
			__antithesis_instrumentation__.Notify(19576)
			if op[1] == "0" {
				__antithesis_instrumentation__.Notify(19580)
				hasInclude0 = true
			} else {
				__antithesis_instrumentation__.Notify(19581)
				return errors.Errorf("include_realm must be set to 0: %s", op[1])
			}
		case "krb_realm":
			__antithesis_instrumentation__.Notify(19577)

		case "map":
			__antithesis_instrumentation__.Notify(19578)
			hasMap = true
		default:
			__antithesis_instrumentation__.Notify(19579)
			return errors.Errorf("unsupported option %s", op[0])
		}
	}
	__antithesis_instrumentation__.Notify(19573)
	if !hasMap && func() bool {
		__antithesis_instrumentation__.Notify(19582)
		return !hasInclude0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(19583)
		return errors.New(`at least one of "include_realm=0" or "map" options required`)
	} else {
		__antithesis_instrumentation__.Notify(19584)
	}
	__antithesis_instrumentation__.Notify(19574)
	return nil
}

func stripRealm(u security.SQLUsername) (security.SQLUsername, error) {
	__antithesis_instrumentation__.Notify(19585)
	norm := u.Normalized()
	if idx := strings.Index(norm, "@"); idx != -1 {
		__antithesis_instrumentation__.Notify(19587)
		norm = norm[:idx]
	} else {
		__antithesis_instrumentation__.Notify(19588)
	}
	__antithesis_instrumentation__.Notify(19586)
	return security.MakeSQLUsernameFromUserInput(norm, security.UsernameValidation)
}

func stripRealmMapper(
	_ context.Context, systemIdentity security.SQLUsername,
) ([]security.SQLUsername, error) {
	__antithesis_instrumentation__.Notify(19589)
	ret, err := stripRealm(systemIdentity)
	return []security.SQLUsername{ret}, err
}

func stripAndDelegateMapper(delegate pgwire.RoleMapper) pgwire.RoleMapper {
	__antithesis_instrumentation__.Notify(19590)
	return func(ctx context.Context, systemIdentity security.SQLUsername) ([]security.SQLUsername, error) {
		__antithesis_instrumentation__.Notify(19591)
		next, err := stripRealm(systemIdentity)
		if err != nil {
			__antithesis_instrumentation__.Notify(19593)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(19594)
		}
		__antithesis_instrumentation__.Notify(19592)
		return delegate(ctx, next)
	}
}

func init() {
	pgwire.RegisterAuthMethod("gss", authGSS, hba.ConnHostSSL, checkEntry)
}
