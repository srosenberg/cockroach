package geomfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"

	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/twpayne/go-geom"
)

func Azimuth(a geo.Geometry, b geo.Geometry) (*float64, error) {
	__antithesis_instrumentation__.Notify(61185)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(61193)
		return nil, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(61194)
	}
	__antithesis_instrumentation__.Notify(61186)

	aGeomT, err := a.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(61195)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(61196)
	}
	__antithesis_instrumentation__.Notify(61187)

	aPoint, ok := aGeomT.(*geom.Point)
	if !ok {
		__antithesis_instrumentation__.Notify(61197)
		return nil, pgerror.Newf(pgcode.InvalidParameterValue, "arguments must be POINT geometries")
	} else {
		__antithesis_instrumentation__.Notify(61198)
	}
	__antithesis_instrumentation__.Notify(61188)

	bGeomT, err := b.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(61199)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(61200)
	}
	__antithesis_instrumentation__.Notify(61189)

	bPoint, ok := bGeomT.(*geom.Point)
	if !ok {
		__antithesis_instrumentation__.Notify(61201)
		return nil, pgerror.Newf(pgcode.InvalidParameterValue, "arguments must be POINT geometries")
	} else {
		__antithesis_instrumentation__.Notify(61202)
	}
	__antithesis_instrumentation__.Notify(61190)

	if aPoint.Empty() || func() bool {
		__antithesis_instrumentation__.Notify(61203)
		return bPoint.Empty() == true
	}() == true {
		__antithesis_instrumentation__.Notify(61204)
		return nil, pgerror.Newf(pgcode.InvalidParameterValue, "cannot call ST_Azimuth with POINT EMPTY")
	} else {
		__antithesis_instrumentation__.Notify(61205)
	}
	__antithesis_instrumentation__.Notify(61191)

	if aPoint.X() == bPoint.X() && func() bool {
		__antithesis_instrumentation__.Notify(61206)
		return aPoint.Y() == bPoint.Y() == true
	}() == true {
		__antithesis_instrumentation__.Notify(61207)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(61208)
	}
	__antithesis_instrumentation__.Notify(61192)

	atan := math.Atan2(bPoint.Y()-aPoint.Y(), bPoint.X()-aPoint.X())

	azimuth := math.Mod(math.Pi/2-atan+2*math.Pi, 2*math.Pi)
	return &azimuth, nil
}
