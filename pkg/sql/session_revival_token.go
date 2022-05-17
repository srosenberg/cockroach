package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/security/sessionrevival"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
)

var AllowSessionRevival = settings.RegisterBoolSetting(
	settings.TenantReadOnly,
	"server.user_login.session_revival_token.enabled",
	"if set, the cluster is able to create session revival tokens and use them "+
		"to authenticate a new session",
	false,
)

func (p *planner) CreateSessionRevivalToken() (*tree.DBytes, error) {
	__antithesis_instrumentation__.Notify(617724)
	cm, err := p.ExecCfg().RPCContext.SecurityContext.GetCertificateManager()
	if err != nil {
		__antithesis_instrumentation__.Notify(617726)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(617727)
	}
	__antithesis_instrumentation__.Notify(617725)
	return createSessionRevivalToken(
		AllowSessionRevival.Get(&p.ExecCfg().Settings.SV) && func() bool {
			__antithesis_instrumentation__.Notify(617728)
			return !p.ExecCfg().Codec.ForSystemTenant() == true
		}() == true,
		p.SessionData(),
		cm,
	)
}

func createSessionRevivalToken(
	allowSessionRevival bool, sd *sessiondata.SessionData, cm *security.CertificateManager,
) (*tree.DBytes, error) {
	__antithesis_instrumentation__.Notify(617729)
	if !allowSessionRevival {
		__antithesis_instrumentation__.Notify(617733)
		return nil, pgerror.New(pgcode.FeatureNotSupported, "session revival tokens are not supported on this cluster")
	} else {
		__antithesis_instrumentation__.Notify(617734)
	}
	__antithesis_instrumentation__.Notify(617730)

	user := sd.SessionUser()
	if user.IsRootUser() {
		__antithesis_instrumentation__.Notify(617735)
		return nil, pgerror.New(pgcode.InsufficientPrivilege, "cannot create token for root user")
	} else {
		__antithesis_instrumentation__.Notify(617736)
	}
	__antithesis_instrumentation__.Notify(617731)

	tokenBytes, err := sessionrevival.CreateSessionRevivalToken(cm, user)
	if err != nil {
		__antithesis_instrumentation__.Notify(617737)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(617738)
	}
	__antithesis_instrumentation__.Notify(617732)
	return tree.NewDBytes(tree.DBytes(tokenBytes)), nil
}

func (p *planner) ValidateSessionRevivalToken(token *tree.DBytes) (*tree.DBool, error) {
	__antithesis_instrumentation__.Notify(617739)
	if !AllowSessionRevival.Get(&p.ExecCfg().Settings.SV) || func() bool {
		__antithesis_instrumentation__.Notify(617743)
		return p.ExecCfg().Codec.ForSystemTenant() == true
	}() == true {
		__antithesis_instrumentation__.Notify(617744)
		return nil, pgerror.New(pgcode.FeatureNotSupported, "session revival tokens are not supported on this cluster")
	} else {
		__antithesis_instrumentation__.Notify(617745)
	}
	__antithesis_instrumentation__.Notify(617740)
	cm, err := p.ExecCfg().RPCContext.SecurityContext.GetCertificateManager()
	if err != nil {
		__antithesis_instrumentation__.Notify(617746)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(617747)
	}
	__antithesis_instrumentation__.Notify(617741)
	if err := sessionrevival.ValidateSessionRevivalToken(cm, p.SessionData().SessionUser(), []byte(*token)); err != nil {
		__antithesis_instrumentation__.Notify(617748)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(617749)
	}
	__antithesis_instrumentation__.Notify(617742)
	return tree.DBoolTrue, nil
}
