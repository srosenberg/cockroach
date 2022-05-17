package geomfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geos"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/ewkb"
)

func LineInterpolatePoints(g geo.Geometry, fraction float64, repeat bool) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(62349)
	if fraction < 0 || func() bool {
		__antithesis_instrumentation__.Notify(62352)
		return fraction > 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(62353)
		return geo.Geometry{}, pgerror.Newf(pgcode.InvalidParameterValue, "fraction %f should be within [0 1] range", fraction)
	} else {
		__antithesis_instrumentation__.Notify(62354)
	}
	__antithesis_instrumentation__.Notify(62350)
	geomRepr, err := g.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(62355)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(62356)
	}
	__antithesis_instrumentation__.Notify(62351)
	switch geomRepr := geomRepr.(type) {
	case *geom.LineString:
		__antithesis_instrumentation__.Notify(62357)

		lengthOfLineString := geomRepr.Length()
		if repeat && func() bool {
			__antithesis_instrumentation__.Notify(62361)
			return fraction <= 0.5 == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(62362)
			return fraction != 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(62363)
			numberOfInterpolatedPoints := int(1 / fraction)
			if numberOfInterpolatedPoints > geo.MaxAllowedSplitPoints {
				__antithesis_instrumentation__.Notify(62366)
				return geo.Geometry{}, pgerror.Newf(
					pgcode.InvalidParameterValue,
					"attempting to interpolate into too many points; requires %d points, max %d",
					numberOfInterpolatedPoints,
					geo.MaxAllowedSplitPoints,
				)
			} else {
				__antithesis_instrumentation__.Notify(62367)
			}
			__antithesis_instrumentation__.Notify(62364)
			interpolatedPoints := geom.NewMultiPoint(geom.XY).SetSRID(geomRepr.SRID())
			for pointInserted := 1; pointInserted <= numberOfInterpolatedPoints; pointInserted++ {
				__antithesis_instrumentation__.Notify(62368)
				pointEWKB, err := geos.InterpolateLine(g.EWKB(), float64(pointInserted)*fraction*lengthOfLineString)
				if err != nil {
					__antithesis_instrumentation__.Notify(62371)
					return geo.Geometry{}, err
				} else {
					__antithesis_instrumentation__.Notify(62372)
				}
				__antithesis_instrumentation__.Notify(62369)
				point, err := ewkb.Unmarshal(pointEWKB)
				if err != nil {
					__antithesis_instrumentation__.Notify(62373)
					return geo.Geometry{}, err
				} else {
					__antithesis_instrumentation__.Notify(62374)
				}
				__antithesis_instrumentation__.Notify(62370)
				err = interpolatedPoints.Push(point.(*geom.Point))
				if err != nil {
					__antithesis_instrumentation__.Notify(62375)
					return geo.Geometry{}, err
				} else {
					__antithesis_instrumentation__.Notify(62376)
				}
			}
			__antithesis_instrumentation__.Notify(62365)
			return geo.MakeGeometryFromGeomT(interpolatedPoints)
		} else {
			__antithesis_instrumentation__.Notify(62377)
		}
		__antithesis_instrumentation__.Notify(62358)
		interpolatedPointEWKB, err := geos.InterpolateLine(g.EWKB(), fraction*lengthOfLineString)
		if err != nil {
			__antithesis_instrumentation__.Notify(62378)
			return geo.Geometry{}, err
		} else {
			__antithesis_instrumentation__.Notify(62379)
		}
		__antithesis_instrumentation__.Notify(62359)
		return geo.ParseGeometryFromEWKB(interpolatedPointEWKB)
	default:
		__antithesis_instrumentation__.Notify(62360)
		return geo.Geometry{}, pgerror.Newf(pgcode.InvalidParameterValue, "geometry %s should be LineString", g.ShapeType())
	}
}
