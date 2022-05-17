package security

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
)

const (
	ocspOff    = 0
	ocspLax    = 1
	ocspStrict = 2
)

type TLSSettings interface {
	ocspEnabled() bool
	ocspStrict() bool
	ocspTimeout() time.Duration
}

var ocspMode = settings.RegisterEnumSetting(
	settings.TenantWritable, "security.ocsp.mode",
	"use OCSP to check whether TLS certificates are revoked. If the OCSP "+
		"server is unreachable, in strict mode all certificates will be rejected "+
		"and in lax mode all certificates will be accepted.",
	"off", map[int64]string{ocspOff: "off", ocspLax: "lax", ocspStrict: "strict"}).WithPublic()

var ocspTimeout = settings.RegisterDurationSetting(
	settings.TenantWritable, "security.ocsp.timeout",
	"timeout before considering the OCSP server unreachable",
	3*time.Second,
	settings.NonNegativeDuration,
).WithPublic()

type clusterTLSSettings struct {
	settings *cluster.Settings
}

var _ TLSSettings = clusterTLSSettings{}

func (c clusterTLSSettings) ocspEnabled() bool {
	__antithesis_instrumentation__.Notify(187331)
	return ocspMode.Get(&c.settings.SV) != ocspOff
}

func (c clusterTLSSettings) ocspStrict() bool {
	__antithesis_instrumentation__.Notify(187332)
	return ocspMode.Get(&c.settings.SV) == ocspStrict
}

func (c clusterTLSSettings) ocspTimeout() time.Duration {
	__antithesis_instrumentation__.Notify(187333)
	return ocspTimeout.Get(&c.settings.SV)
}

func ClusterTLSSettings(settings *cluster.Settings) TLSSettings {
	__antithesis_instrumentation__.Notify(187334)
	return clusterTLSSettings{settings}
}

type CommandTLSSettings struct{}

var _ TLSSettings = CommandTLSSettings{}

func (CommandTLSSettings) ocspEnabled() bool {
	__antithesis_instrumentation__.Notify(187335)
	return false
}

func (CommandTLSSettings) ocspStrict() bool {
	__antithesis_instrumentation__.Notify(187336)
	return false
}

func (CommandTLSSettings) ocspTimeout() time.Duration {
	__antithesis_instrumentation__.Notify(187337)
	return 0
}
