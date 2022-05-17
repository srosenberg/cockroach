package geomfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/twpayne/go-geom"
)

func ShiftLongitude(geometry geo.Geometry) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(62947)
	t, err := geometry.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(62951)
		return geometry, err
	} else {
		__antithesis_instrumentation__.Notify(62952)
	}
	__antithesis_instrumentation__.Notify(62948)

	newT, err := applyOnCoordsForGeomT(t, func(l geom.Layout, dst []float64, src []float64) error {
		__antithesis_instrumentation__.Notify(62953)
		copy(dst, src)
		if src[0] < 0 {
			__antithesis_instrumentation__.Notify(62955)
			dst[0] += 360
		} else {
			__antithesis_instrumentation__.Notify(62956)
			if src[0] > 180 {
				__antithesis_instrumentation__.Notify(62957)
				dst[0] -= 360
			} else {
				__antithesis_instrumentation__.Notify(62958)
			}
		}
		__antithesis_instrumentation__.Notify(62954)
		return nil
	})
	__antithesis_instrumentation__.Notify(62949)

	if err != nil {
		__antithesis_instrumentation__.Notify(62959)
		return geometry, err
	} else {
		__antithesis_instrumentation__.Notify(62960)
	}
	__antithesis_instrumentation__.Notify(62950)

	return geo.MakeGeometryFromGeomT(newT)
}
