package geomfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/geo"

func Envelope(g geo.Geometry) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(62039)
	if g.Empty() {
		__antithesis_instrumentation__.Notify(62041)
		return g, nil
	} else {
		__antithesis_instrumentation__.Notify(62042)
	}
	__antithesis_instrumentation__.Notify(62040)
	return geo.MakeGeometryFromGeomT(g.CartesianBoundingBox().ToGeomT(g.SRID()))
}
