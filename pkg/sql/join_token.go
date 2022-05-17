package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/featureflag"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

var FeatureTLSAutoJoinEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"feature.tls_auto_join.enabled",
	"set to true to enable tls auto join through join tokens, false to disable; default is false",
	false,
)

func (p *planner) CreateJoinToken(ctx context.Context) (string, error) {
	__antithesis_instrumentation__.Notify(498835)
	hasAdmin, err := p.HasAdminRole(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(498844)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(498845)
	}
	__antithesis_instrumentation__.Notify(498836)
	if !hasAdmin {
		__antithesis_instrumentation__.Notify(498846)
		return "", pgerror.New(pgcode.InsufficientPrivilege, "must be admin to create join token")
	} else {
		__antithesis_instrumentation__.Notify(498847)
	}
	__antithesis_instrumentation__.Notify(498837)
	if err = featureflag.CheckEnabled(
		ctx, p.ExecCfg(), FeatureTLSAutoJoinEnabled, "create join tokens"); err != nil {
		__antithesis_instrumentation__.Notify(498848)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(498849)
	}
	__antithesis_instrumentation__.Notify(498838)

	cm, err := p.ExecCfg().RPCContext.SecurityContext.GetCertificateManager()
	if err != nil {
		__antithesis_instrumentation__.Notify(498850)
		return "", errors.Wrap(err, "error when getting certificate manager")
	} else {
		__antithesis_instrumentation__.Notify(498851)
	}
	__antithesis_instrumentation__.Notify(498839)

	jt, err := security.GenerateJoinToken(cm)
	if err != nil {
		__antithesis_instrumentation__.Notify(498852)
		return "", errors.Wrap(err, "error when generating join token")
	} else {
		__antithesis_instrumentation__.Notify(498853)
	}
	__antithesis_instrumentation__.Notify(498840)
	token, err := jt.MarshalText()
	if err != nil {
		__antithesis_instrumentation__.Notify(498854)
		return "", errors.Wrap(err, "error when marshaling join token")
	} else {
		__antithesis_instrumentation__.Notify(498855)
	}
	__antithesis_instrumentation__.Notify(498841)
	expiration := timeutil.Now().Add(security.JoinTokenExpiration)
	err = p.ExecCfg().DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(498856)
		_, err = p.ExecCfg().InternalExecutor.Exec(
			ctx, "insert-join-token", txn,
			"insert into system.join_tokens(id, secret, expiration) "+
				"values($1, $2, $3)",
			jt.TokenID.String(), jt.SharedSecret, expiration.Format(time.RFC3339),
		)
		return err
	})
	__antithesis_instrumentation__.Notify(498842)
	if err != nil {
		__antithesis_instrumentation__.Notify(498857)
		return "", errors.Wrap(err, "could not persist join token in system table")
	} else {
		__antithesis_instrumentation__.Notify(498858)
	}
	__antithesis_instrumentation__.Notify(498843)
	return string(token), nil
}
