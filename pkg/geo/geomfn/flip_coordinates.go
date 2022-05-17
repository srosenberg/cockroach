package geomfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/twpayne/go-geom"
)

func FlipCoordinates(g geo.Geometry) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(62043)
	if g.Empty() {
		__antithesis_instrumentation__.Notify(62048)
		return g, nil
	} else {
		__antithesis_instrumentation__.Notify(62049)
	}
	__antithesis_instrumentation__.Notify(62044)

	t, err := g.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(62050)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(62051)
	}
	__antithesis_instrumentation__.Notify(62045)

	newT, err := applyOnCoordsForGeomT(t, func(l geom.Layout, dst, src []float64) error {
		__antithesis_instrumentation__.Notify(62052)
		dst[0], dst[1] = src[1], src[0]
		if l.ZIndex() != -1 {
			__antithesis_instrumentation__.Notify(62055)
			dst[l.ZIndex()] = src[l.ZIndex()]
		} else {
			__antithesis_instrumentation__.Notify(62056)
		}
		__antithesis_instrumentation__.Notify(62053)
		if l.MIndex() != -1 {
			__antithesis_instrumentation__.Notify(62057)
			dst[l.MIndex()] = src[l.MIndex()]
		} else {
			__antithesis_instrumentation__.Notify(62058)
		}
		__antithesis_instrumentation__.Notify(62054)
		return nil
	})
	__antithesis_instrumentation__.Notify(62046)
	if err != nil {
		__antithesis_instrumentation__.Notify(62059)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(62060)
	}
	__antithesis_instrumentation__.Notify(62047)

	return geo.MakeGeometryFromGeomT(newT)
}
