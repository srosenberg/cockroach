package geogfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"

	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/golang/geo/s2"
	"github.com/twpayne/go-geom"
)

func Azimuth(a geo.Geography, b geo.Geography) (*float64, error) {
	__antithesis_instrumentation__.Notify(59435)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(59444)
		return nil, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(59445)
	}
	__antithesis_instrumentation__.Notify(59436)

	aGeomT, err := a.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(59446)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(59447)
	}
	__antithesis_instrumentation__.Notify(59437)

	aPoint, ok := aGeomT.(*geom.Point)
	if !ok {
		__antithesis_instrumentation__.Notify(59448)
		return nil, pgerror.Newf(pgcode.InvalidParameterValue, "arguments must be POINT geometries")
	} else {
		__antithesis_instrumentation__.Notify(59449)
	}
	__antithesis_instrumentation__.Notify(59438)

	bGeomT, err := b.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(59450)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(59451)
	}
	__antithesis_instrumentation__.Notify(59439)

	bPoint, ok := bGeomT.(*geom.Point)
	if !ok {
		__antithesis_instrumentation__.Notify(59452)
		return nil, pgerror.Newf(pgcode.InvalidParameterValue, "arguments must be POINT geometries")
	} else {
		__antithesis_instrumentation__.Notify(59453)
	}
	__antithesis_instrumentation__.Notify(59440)

	if aPoint.Empty() || func() bool {
		__antithesis_instrumentation__.Notify(59454)
		return bPoint.Empty() == true
	}() == true {
		__antithesis_instrumentation__.Notify(59455)
		return nil, pgerror.Newf(pgcode.InvalidParameterValue, "cannot call ST_Azimuth with POINT EMPTY")
	} else {
		__antithesis_instrumentation__.Notify(59456)
	}
	__antithesis_instrumentation__.Notify(59441)

	if aPoint.X() == bPoint.X() && func() bool {
		__antithesis_instrumentation__.Notify(59457)
		return aPoint.Y() == bPoint.Y() == true
	}() == true {
		__antithesis_instrumentation__.Notify(59458)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(59459)
	}
	__antithesis_instrumentation__.Notify(59442)

	s, err := a.Spheroid()
	if err != nil {
		__antithesis_instrumentation__.Notify(59460)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(59461)
	}
	__antithesis_instrumentation__.Notify(59443)

	_, az1, _ := s.Inverse(
		s2.LatLngFromDegrees(aPoint.Y(), aPoint.X()),
		s2.LatLngFromDegrees(bPoint.Y(), bPoint.X()),
	)

	az1 = az1 * math.Pi / 180
	return &az1, nil
}
