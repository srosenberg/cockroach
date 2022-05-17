package install

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/roachprod/config"

type ClusterSettings struct {
	Binary        string
	Secure        bool
	PGUrlCertsDir string
	Env           []string
	Tag           string
	UseTreeDist   bool
	NumRacks      int

	DebugDir string
}

type ClusterSettingOption interface {
	apply(settings *ClusterSettings)
}

type TagOption string

func (o TagOption) apply(settings *ClusterSettings) {
	__antithesis_instrumentation__.Notify(180454)
	settings.Tag = string(o)
}

type BinaryOption string

func (o BinaryOption) apply(settings *ClusterSettings) {
	__antithesis_instrumentation__.Notify(180455)
	settings.Binary = string(o)
}

type PGUrlCertsDirOption string

func (o PGUrlCertsDirOption) apply(settings *ClusterSettings) {
	__antithesis_instrumentation__.Notify(180456)
	settings.PGUrlCertsDir = string(o)
}

type SecureOption bool

func (o SecureOption) apply(settings *ClusterSettings) {
	__antithesis_instrumentation__.Notify(180457)
	settings.Secure = bool(o)
}

type UseTreeDistOption bool

func (o UseTreeDistOption) apply(settings *ClusterSettings) {
	__antithesis_instrumentation__.Notify(180458)
	settings.UseTreeDist = bool(o)
}

type EnvOption []string

var _ EnvOption

func (o EnvOption) apply(settings *ClusterSettings) {
	__antithesis_instrumentation__.Notify(180459)
	settings.Env = append(settings.Env, []string(o)...)
}

type NumRacksOption int

var _ NumRacksOption

func (o NumRacksOption) apply(settings *ClusterSettings) {
	__antithesis_instrumentation__.Notify(180460)
	settings.NumRacks = int(o)
}

type DebugDirOption string

var _ DebugDirOption

func (o DebugDirOption) apply(settings *ClusterSettings) {
	__antithesis_instrumentation__.Notify(180461)
	settings.DebugDir = string(o)
}

func MakeClusterSettings(opts ...ClusterSettingOption) ClusterSettings {
	__antithesis_instrumentation__.Notify(180462)
	clusterSettings := ClusterSettings{
		Binary:        config.Binary,
		Tag:           "",
		PGUrlCertsDir: "./certs",
		Secure:        false,
		UseTreeDist:   true,
		Env: []string{
			"COCKROACH_ENABLE_RPC_COMPRESSION=false",
			"COCKROACH_UI_RELEASE_NOTES_SIGNUP_DISMISSED=true",
		},
		NumRacks: 0,
	}

	for _, opt := range opts {
		__antithesis_instrumentation__.Notify(180464)
		opt.apply(&clusterSettings)
	}
	__antithesis_instrumentation__.Notify(180463)
	return clusterSettings
}
