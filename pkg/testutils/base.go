package testutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/security"
)

func NewNodeTestBaseContext() *base.Config {
	__antithesis_instrumentation__.Notify(643975)
	return NewTestBaseContext(security.NodeUserName())
}

func NewTestBaseContext(user security.SQLUsername) *base.Config {
	__antithesis_instrumentation__.Notify(643976)
	cfg := &base.Config{
		Insecure: false,
		User:     user,
	}
	FillCerts(cfg)
	return cfg
}

func FillCerts(cfg *base.Config) {
	__antithesis_instrumentation__.Notify(643977)
	cfg.SSLCertsDir = security.EmbeddedCertsDir
}
