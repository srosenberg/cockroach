package geomfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/geo/geos"
)

func VoronoiDiagram(
	g geo.Geometry, env *geo.Geometry, tolerance float64, onlyEdges bool,
) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(63651)
	var envEWKB geopb.EWKB
	if !(env == nil) {
		__antithesis_instrumentation__.Notify(63655)
		if g.SRID() != env.SRID() {
			__antithesis_instrumentation__.Notify(63657)
			return geo.Geometry{}, geo.NewMismatchingSRIDsError(g.SpatialObject(), env.SpatialObject())
		} else {
			__antithesis_instrumentation__.Notify(63658)
		}
		__antithesis_instrumentation__.Notify(63656)
		envEWKB = env.EWKB()
	} else {
		__antithesis_instrumentation__.Notify(63659)
	}
	__antithesis_instrumentation__.Notify(63652)
	paths, err := geos.VoronoiDiagram(g.EWKB(), envEWKB, tolerance, onlyEdges)
	if err != nil {
		__antithesis_instrumentation__.Notify(63660)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(63661)
	}
	__antithesis_instrumentation__.Notify(63653)
	gm, err := geo.ParseGeometryFromEWKB(paths)
	if err != nil {
		__antithesis_instrumentation__.Notify(63662)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(63663)
	}
	__antithesis_instrumentation__.Notify(63654)
	return gm, nil
}
