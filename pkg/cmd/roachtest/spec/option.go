package spec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "time"

type Option interface {
	apply(spec *ClusterSpec)
}

type nodeCPUOption int

func (o nodeCPUOption) apply(spec *ClusterSpec) {
	__antithesis_instrumentation__.Notify(44560)
	spec.CPUs = int(o)
}

func CPU(n int) Option {
	__antithesis_instrumentation__.Notify(44561)
	return nodeCPUOption(n)
}

type volumeSizeOption int

func (o volumeSizeOption) apply(spec *ClusterSpec) {
	__antithesis_instrumentation__.Notify(44562)
	spec.VolumeSize = int(o)
}

func VolumeSize(n int) Option {
	__antithesis_instrumentation__.Notify(44563)
	return volumeSizeOption(n)
}

type nodeSSDOption int

func (o nodeSSDOption) apply(spec *ClusterSpec) {
	__antithesis_instrumentation__.Notify(44564)
	spec.SSDs = int(o)
}

func SSD(n int) Option {
	__antithesis_instrumentation__.Notify(44565)
	return nodeSSDOption(n)
}

type raid0Option bool

func (o raid0Option) apply(spec *ClusterSpec) {
	__antithesis_instrumentation__.Notify(44566)
	spec.RAID0 = bool(o)
}

func RAID0(enabled bool) Option {
	__antithesis_instrumentation__.Notify(44567)
	return raid0Option(enabled)
}

type nodeGeoOption struct{}

func (o nodeGeoOption) apply(spec *ClusterSpec) {
	__antithesis_instrumentation__.Notify(44568)
	spec.Geo = true
}

func Geo() Option {
	__antithesis_instrumentation__.Notify(44569)
	return nodeGeoOption{}
}

type nodeZonesOption string

func (o nodeZonesOption) apply(spec *ClusterSpec) {
	__antithesis_instrumentation__.Notify(44570)
	spec.Zones = string(o)
}

func Zones(s string) Option {
	__antithesis_instrumentation__.Notify(44571)
	return nodeZonesOption(s)
}

type nodeLifetimeOption time.Duration

func (o nodeLifetimeOption) apply(spec *ClusterSpec) {
	__antithesis_instrumentation__.Notify(44572)
	spec.Lifetime = time.Duration(o)
}

type clusterReusePolicy interface {
	clusterReusePolicy()
}

type ReusePolicyAny struct{}

type ReusePolicyNone struct{}

type ReusePolicyTagged struct{ Tag string }

func (ReusePolicyAny) clusterReusePolicy()    { __antithesis_instrumentation__.Notify(44573) }
func (ReusePolicyNone) clusterReusePolicy()   { __antithesis_instrumentation__.Notify(44574) }
func (ReusePolicyTagged) clusterReusePolicy() { __antithesis_instrumentation__.Notify(44575) }

type clusterReusePolicyOption struct {
	p clusterReusePolicy
}

func ReuseAny() Option {
	__antithesis_instrumentation__.Notify(44576)
	return clusterReusePolicyOption{p: ReusePolicyAny{}}
}

func ReuseNone() Option {
	__antithesis_instrumentation__.Notify(44577)
	return clusterReusePolicyOption{p: ReusePolicyNone{}}
}

func ReuseTagged(tag string) Option {
	__antithesis_instrumentation__.Notify(44578)
	return clusterReusePolicyOption{p: ReusePolicyTagged{Tag: tag}}
}

func (p clusterReusePolicyOption) apply(spec *ClusterSpec) {
	__antithesis_instrumentation__.Notify(44579)
	spec.ReusePolicy = p.p
}

type preferSSDOption struct{}

func (*preferSSDOption) apply(spec *ClusterSpec) {
	__antithesis_instrumentation__.Notify(44580)
	spec.PreferLocalSSD = true
}

func PreferSSD() Option {
	__antithesis_instrumentation__.Notify(44581)
	return &preferSSDOption{}
}

type setFileSystem struct {
	fs fileSystemType
}

func (s *setFileSystem) apply(spec *ClusterSpec) {
	__antithesis_instrumentation__.Notify(44582)
	spec.FileSystem = s.fs
}

func SetFileSystem(fs fileSystemType) Option {
	__antithesis_instrumentation__.Notify(44583)
	return &setFileSystem{fs}
}

type randomlyUseZfs struct{}

func (r *randomlyUseZfs) apply(spec *ClusterSpec) {
	__antithesis_instrumentation__.Notify(44584)
	spec.RandomlyUseZfs = true
}

func RandomlyUseZfs() Option {
	__antithesis_instrumentation__.Notify(44585)
	return &randomlyUseZfs{}
}
