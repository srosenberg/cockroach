package geomfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geos"
)

func Snap(input, target geo.Geometry, tolerance float64) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(63101)
	snappedEWKB, err := geos.Snap(input.EWKB(), target.EWKB(), tolerance)
	if err != nil {
		__antithesis_instrumentation__.Notify(63103)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(63104)
	}
	__antithesis_instrumentation__.Notify(63102)
	return geo.ParseGeometryFromEWKB(snappedEWKB)
}
