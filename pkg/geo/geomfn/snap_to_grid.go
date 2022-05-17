package geomfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"

	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/twpayne/go-geom"
)

func SnapToGrid(g geo.Geometry, origin geom.Coord, gridSize geom.Coord) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(63105)
	if len(origin) != 4 {
		__antithesis_instrumentation__.Notify(63111)
		return geo.Geometry{}, pgerror.Newf(pgcode.InvalidParameterValue, "origin must be 4D")
	} else {
		__antithesis_instrumentation__.Notify(63112)
	}
	__antithesis_instrumentation__.Notify(63106)
	if len(gridSize) != 4 {
		__antithesis_instrumentation__.Notify(63113)
		return geo.Geometry{}, pgerror.Newf(pgcode.InvalidParameterValue, "gridSize must be 4D")
	} else {
		__antithesis_instrumentation__.Notify(63114)
	}
	__antithesis_instrumentation__.Notify(63107)
	if g.Empty() {
		__antithesis_instrumentation__.Notify(63115)
		return g, nil
	} else {
		__antithesis_instrumentation__.Notify(63116)
	}
	__antithesis_instrumentation__.Notify(63108)
	geomT, err := g.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(63117)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(63118)
	}
	__antithesis_instrumentation__.Notify(63109)
	retGeomT, err := snapToGrid(geomT, origin, gridSize)
	if err != nil {
		__antithesis_instrumentation__.Notify(63119)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(63120)
	}
	__antithesis_instrumentation__.Notify(63110)
	return geo.MakeGeometryFromGeomT(retGeomT)
}

func snapCoordinateToGrid(
	l geom.Layout, dst []float64, src []float64, origin geom.Coord, gridSize geom.Coord,
) {
	__antithesis_instrumentation__.Notify(63121)
	dst[0] = snapOrdinateToGrid(src[0], origin[0], gridSize[0])
	dst[1] = snapOrdinateToGrid(src[1], origin[1], gridSize[1])
	if l.ZIndex() != -1 {
		__antithesis_instrumentation__.Notify(63123)
		dst[l.ZIndex()] = snapOrdinateToGrid(src[l.ZIndex()], origin[l.ZIndex()], gridSize[l.ZIndex()])
	} else {
		__antithesis_instrumentation__.Notify(63124)
	}
	__antithesis_instrumentation__.Notify(63122)
	if l.MIndex() != -1 {
		__antithesis_instrumentation__.Notify(63125)
		dst[l.MIndex()] = snapOrdinateToGrid(src[l.MIndex()], origin[l.MIndex()], gridSize[l.MIndex()])
	} else {
		__antithesis_instrumentation__.Notify(63126)
	}
}

func snapOrdinateToGrid(
	ordinate float64, originOrdinate float64, gridSizeOrdinate float64,
) float64 {
	__antithesis_instrumentation__.Notify(63127)

	if gridSizeOrdinate <= 0 {
		__antithesis_instrumentation__.Notify(63129)
		return ordinate
	} else {
		__antithesis_instrumentation__.Notify(63130)
	}
	__antithesis_instrumentation__.Notify(63128)
	return math.RoundToEven((ordinate-originOrdinate)/gridSizeOrdinate)*gridSizeOrdinate + originOrdinate
}

func snapToGrid(t geom.T, origin geom.Coord, gridSize geom.Coord) (geom.T, error) {
	__antithesis_instrumentation__.Notify(63131)
	if t.Empty() {
		__antithesis_instrumentation__.Notify(63135)
		return t, nil
	} else {
		__antithesis_instrumentation__.Notify(63136)
	}
	__antithesis_instrumentation__.Notify(63132)
	t, err := applyOnCoordsForGeomT(t, func(l geom.Layout, dst []float64, src []float64) error {
		__antithesis_instrumentation__.Notify(63137)
		snapCoordinateToGrid(l, dst, src, origin, gridSize)
		return nil
	})
	__antithesis_instrumentation__.Notify(63133)
	if err != nil {
		__antithesis_instrumentation__.Notify(63138)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(63139)
	}
	__antithesis_instrumentation__.Notify(63134)
	return removeConsecutivePointsFromGeomT(t)
}
