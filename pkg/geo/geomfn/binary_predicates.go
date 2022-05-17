package geomfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/geo/geos"
	"github.com/cockroachdb/errors"
	"github.com/twpayne/go-geom"
)

func Covers(a geo.Geometry, b geo.Geometry) (bool, error) {
	__antithesis_instrumentation__.Notify(61209)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(61213)
		return false, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(61214)
	}
	__antithesis_instrumentation__.Notify(61210)
	if !a.CartesianBoundingBox().Covers(b.CartesianBoundingBox()) {
		__antithesis_instrumentation__.Notify(61215)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(61216)
	}
	__antithesis_instrumentation__.Notify(61211)

	pointPolygonPair, pointKind, polygonKind := PointKindAndPolygonKind(a, b)
	switch pointPolygonPair {
	case PointAndPolygon:
		__antithesis_instrumentation__.Notify(61217)

		return false, nil
	case PolygonAndPoint:
		__antithesis_instrumentation__.Notify(61218)

		return PointKindCoveredByPolygonKind(pointKind, polygonKind)
	default:
		__antithesis_instrumentation__.Notify(61219)
	}
	__antithesis_instrumentation__.Notify(61212)

	return geos.Covers(a.EWKB(), b.EWKB())
}

func CoveredBy(a geo.Geometry, b geo.Geometry) (bool, error) {
	__antithesis_instrumentation__.Notify(61220)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(61224)
		return false, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(61225)
	}
	__antithesis_instrumentation__.Notify(61221)
	if !b.CartesianBoundingBox().Covers(a.CartesianBoundingBox()) {
		__antithesis_instrumentation__.Notify(61226)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(61227)
	}
	__antithesis_instrumentation__.Notify(61222)

	pointPolygonPair, pointKind, polygonKind := PointKindAndPolygonKind(a, b)
	switch pointPolygonPair {
	case PolygonAndPoint:
		__antithesis_instrumentation__.Notify(61228)

		return false, nil
	case PointAndPolygon:
		__antithesis_instrumentation__.Notify(61229)
		return PointKindCoveredByPolygonKind(pointKind, polygonKind)
	default:
		__antithesis_instrumentation__.Notify(61230)
	}
	__antithesis_instrumentation__.Notify(61223)

	return geos.CoveredBy(a.EWKB(), b.EWKB())
}

func Contains(a geo.Geometry, b geo.Geometry) (bool, error) {
	__antithesis_instrumentation__.Notify(61231)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(61235)
		return false, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(61236)
	}
	__antithesis_instrumentation__.Notify(61232)
	if !a.CartesianBoundingBox().Covers(b.CartesianBoundingBox()) {
		__antithesis_instrumentation__.Notify(61237)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(61238)
	}
	__antithesis_instrumentation__.Notify(61233)

	pointPolygonPair, pointKind, polygonKind := PointKindAndPolygonKind(a, b)
	switch pointPolygonPair {
	case PointAndPolygon:
		__antithesis_instrumentation__.Notify(61239)

		return false, nil
	case PolygonAndPoint:
		__antithesis_instrumentation__.Notify(61240)

		return PointKindWithinPolygonKind(pointKind, polygonKind)
	default:
		__antithesis_instrumentation__.Notify(61241)
	}
	__antithesis_instrumentation__.Notify(61234)

	return geos.Contains(a.EWKB(), b.EWKB())
}

func ContainsProperly(a geo.Geometry, b geo.Geometry) (bool, error) {
	__antithesis_instrumentation__.Notify(61242)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(61245)
		return false, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(61246)
	}
	__antithesis_instrumentation__.Notify(61243)
	if !a.CartesianBoundingBox().Covers(b.CartesianBoundingBox()) {
		__antithesis_instrumentation__.Notify(61247)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(61248)
	}
	__antithesis_instrumentation__.Notify(61244)
	return geos.RelatePattern(a.EWKB(), b.EWKB(), "T**FF*FF*")
}

func Crosses(a geo.Geometry, b geo.Geometry) (bool, error) {
	__antithesis_instrumentation__.Notify(61249)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(61252)
		return false, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(61253)
	}
	__antithesis_instrumentation__.Notify(61250)
	if !a.CartesianBoundingBox().Intersects(b.CartesianBoundingBox()) {
		__antithesis_instrumentation__.Notify(61254)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(61255)
	}
	__antithesis_instrumentation__.Notify(61251)
	return geos.Crosses(a.EWKB(), b.EWKB())
}

func Disjoint(a geo.Geometry, b geo.Geometry) (bool, error) {
	__antithesis_instrumentation__.Notify(61256)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(61258)
		return false, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(61259)
	}
	__antithesis_instrumentation__.Notify(61257)
	return geos.Disjoint(a.EWKB(), b.EWKB())
}

func Equals(a geo.Geometry, b geo.Geometry) (bool, error) {
	__antithesis_instrumentation__.Notify(61260)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(61264)
		return false, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(61265)
	}
	__antithesis_instrumentation__.Notify(61261)

	if a.Empty() && func() bool {
		__antithesis_instrumentation__.Notify(61266)
		return b.Empty() == true
	}() == true {
		__antithesis_instrumentation__.Notify(61267)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(61268)
	}
	__antithesis_instrumentation__.Notify(61262)
	if !a.CartesianBoundingBox().Covers(b.CartesianBoundingBox()) {
		__antithesis_instrumentation__.Notify(61269)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(61270)
	}
	__antithesis_instrumentation__.Notify(61263)
	return geos.Equals(a.EWKB(), b.EWKB())
}

func Intersects(a geo.Geometry, b geo.Geometry) (bool, error) {
	__antithesis_instrumentation__.Notify(61271)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(61275)
		return false, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(61276)
	}
	__antithesis_instrumentation__.Notify(61272)
	if !a.CartesianBoundingBox().Intersects(b.CartesianBoundingBox()) {
		__antithesis_instrumentation__.Notify(61277)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(61278)
	}
	__antithesis_instrumentation__.Notify(61273)

	pointPolygonPair, pointKind, polygonKind := PointKindAndPolygonKind(a, b)
	switch pointPolygonPair {
	case PointAndPolygon, PolygonAndPoint:
		__antithesis_instrumentation__.Notify(61279)
		return PointKindIntersectsPolygonKind(pointKind, polygonKind)
	default:
		__antithesis_instrumentation__.Notify(61280)
	}
	__antithesis_instrumentation__.Notify(61274)

	return geos.Intersects(a.EWKB(), b.EWKB())
}

func OrderingEquals(a, b geo.Geometry) (bool, error) {
	__antithesis_instrumentation__.Notify(61281)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(61286)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(61287)
	}
	__antithesis_instrumentation__.Notify(61282)
	aBox, bBox := a.CartesianBoundingBox(), b.CartesianBoundingBox()
	switch {
	case aBox == nil && func() bool {
		__antithesis_instrumentation__.Notify(61292)
		return bBox == nil == true
	}() == true:
		__antithesis_instrumentation__.Notify(61288)
	case aBox == nil || func() bool {
		__antithesis_instrumentation__.Notify(61293)
		return bBox == nil == true
	}() == true:
		__antithesis_instrumentation__.Notify(61289)
		return false, nil
	case aBox.Compare(bBox) != 0:
		__antithesis_instrumentation__.Notify(61290)
		return false, nil
	default:
		__antithesis_instrumentation__.Notify(61291)
	}
	__antithesis_instrumentation__.Notify(61283)

	geomA, err := a.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(61294)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(61295)
	}
	__antithesis_instrumentation__.Notify(61284)
	geomB, err := b.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(61296)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(61297)
	}
	__antithesis_instrumentation__.Notify(61285)
	return orderingEqualsFromGeomT(geomA, geomB)
}

func orderingEqualsFromGeomT(a, b geom.T) (bool, error) {
	__antithesis_instrumentation__.Notify(61298)
	if a.Layout() != b.Layout() {
		__antithesis_instrumentation__.Notify(61301)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(61302)
	}
	__antithesis_instrumentation__.Notify(61299)
	switch a := a.(type) {
	case *geom.Point:
		__antithesis_instrumentation__.Notify(61303)
		if b, ok := b.(*geom.Point); ok {
			__antithesis_instrumentation__.Notify(61311)

			switch {
			case a.Empty() && func() bool {
				__antithesis_instrumentation__.Notify(61315)
				return b.Empty() == true
			}() == true:
				__antithesis_instrumentation__.Notify(61312)
				return true, nil
			case a.Empty() || func() bool {
				__antithesis_instrumentation__.Notify(61316)
				return b.Empty() == true
			}() == true:
				__antithesis_instrumentation__.Notify(61313)
				return false, nil
			default:
				__antithesis_instrumentation__.Notify(61314)
				return a.Coords().Equal(b.Layout(), b.Coords()), nil
			}
		} else {
			__antithesis_instrumentation__.Notify(61317)
		}
	case *geom.LineString:
		__antithesis_instrumentation__.Notify(61304)
		if b, ok := b.(*geom.LineString); ok && func() bool {
			__antithesis_instrumentation__.Notify(61318)
			return a.NumCoords() == b.NumCoords() == true
		}() == true {
			__antithesis_instrumentation__.Notify(61319)
			for i := 0; i < a.NumCoords(); i++ {
				__antithesis_instrumentation__.Notify(61321)
				if !a.Coord(i).Equal(b.Layout(), b.Coord(i)) {
					__antithesis_instrumentation__.Notify(61322)
					return false, nil
				} else {
					__antithesis_instrumentation__.Notify(61323)
				}
			}
			__antithesis_instrumentation__.Notify(61320)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(61324)
		}
	case *geom.Polygon:
		__antithesis_instrumentation__.Notify(61305)
		if b, ok := b.(*geom.Polygon); ok && func() bool {
			__antithesis_instrumentation__.Notify(61325)
			return a.NumLinearRings() == b.NumLinearRings() == true
		}() == true {
			__antithesis_instrumentation__.Notify(61326)
			for i := 0; i < a.NumLinearRings(); i++ {
				__antithesis_instrumentation__.Notify(61328)
				for j := 0; j < a.LinearRing(i).NumCoords(); j++ {
					__antithesis_instrumentation__.Notify(61329)
					if !a.LinearRing(i).Coord(j).Equal(b.Layout(), b.LinearRing(i).Coord(j)) {
						__antithesis_instrumentation__.Notify(61330)
						return false, nil
					} else {
						__antithesis_instrumentation__.Notify(61331)
					}
				}
			}
			__antithesis_instrumentation__.Notify(61327)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(61332)
		}
	case *geom.MultiPoint:
		__antithesis_instrumentation__.Notify(61306)
		if b, ok := b.(*geom.MultiPoint); ok && func() bool {
			__antithesis_instrumentation__.Notify(61333)
			return a.NumPoints() == b.NumPoints() == true
		}() == true {
			__antithesis_instrumentation__.Notify(61334)
			for i := 0; i < a.NumPoints(); i++ {
				__antithesis_instrumentation__.Notify(61336)
				if eq, err := orderingEqualsFromGeomT(a.Point(i), b.Point(i)); err != nil || func() bool {
					__antithesis_instrumentation__.Notify(61337)
					return !eq == true
				}() == true {
					__antithesis_instrumentation__.Notify(61338)
					return false, err
				} else {
					__antithesis_instrumentation__.Notify(61339)
				}
			}
			__antithesis_instrumentation__.Notify(61335)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(61340)
		}
	case *geom.MultiLineString:
		__antithesis_instrumentation__.Notify(61307)
		if b, ok := b.(*geom.MultiLineString); ok && func() bool {
			__antithesis_instrumentation__.Notify(61341)
			return a.NumLineStrings() == b.NumLineStrings() == true
		}() == true {
			__antithesis_instrumentation__.Notify(61342)
			for i := 0; i < a.NumLineStrings(); i++ {
				__antithesis_instrumentation__.Notify(61344)
				if eq, err := orderingEqualsFromGeomT(a.LineString(i), b.LineString(i)); err != nil || func() bool {
					__antithesis_instrumentation__.Notify(61345)
					return !eq == true
				}() == true {
					__antithesis_instrumentation__.Notify(61346)
					return false, err
				} else {
					__antithesis_instrumentation__.Notify(61347)
				}
			}
			__antithesis_instrumentation__.Notify(61343)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(61348)
		}
	case *geom.MultiPolygon:
		__antithesis_instrumentation__.Notify(61308)
		if b, ok := b.(*geom.MultiPolygon); ok && func() bool {
			__antithesis_instrumentation__.Notify(61349)
			return a.NumPolygons() == b.NumPolygons() == true
		}() == true {
			__antithesis_instrumentation__.Notify(61350)
			for i := 0; i < a.NumPolygons(); i++ {
				__antithesis_instrumentation__.Notify(61352)
				if eq, err := orderingEqualsFromGeomT(a.Polygon(i), b.Polygon(i)); err != nil || func() bool {
					__antithesis_instrumentation__.Notify(61353)
					return !eq == true
				}() == true {
					__antithesis_instrumentation__.Notify(61354)
					return false, err
				} else {
					__antithesis_instrumentation__.Notify(61355)
				}
			}
			__antithesis_instrumentation__.Notify(61351)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(61356)
		}
	case *geom.GeometryCollection:
		__antithesis_instrumentation__.Notify(61309)
		if b, ok := b.(*geom.GeometryCollection); ok && func() bool {
			__antithesis_instrumentation__.Notify(61357)
			return a.NumGeoms() == b.NumGeoms() == true
		}() == true {
			__antithesis_instrumentation__.Notify(61358)
			for i := 0; i < a.NumGeoms(); i++ {
				__antithesis_instrumentation__.Notify(61360)
				if eq, err := orderingEqualsFromGeomT(a.Geom(i), b.Geom(i)); err != nil || func() bool {
					__antithesis_instrumentation__.Notify(61361)
					return !eq == true
				}() == true {
					__antithesis_instrumentation__.Notify(61362)
					return false, err
				} else {
					__antithesis_instrumentation__.Notify(61363)
				}
			}
			__antithesis_instrumentation__.Notify(61359)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(61364)
		}
	default:
		__antithesis_instrumentation__.Notify(61310)
		return false, errors.AssertionFailedf("unknown geometry type: %T", a)
	}
	__antithesis_instrumentation__.Notify(61300)
	return false, nil
}

func Overlaps(a geo.Geometry, b geo.Geometry) (bool, error) {
	__antithesis_instrumentation__.Notify(61365)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(61368)
		return false, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(61369)
	}
	__antithesis_instrumentation__.Notify(61366)
	if !a.CartesianBoundingBox().Intersects(b.CartesianBoundingBox()) {
		__antithesis_instrumentation__.Notify(61370)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(61371)
	}
	__antithesis_instrumentation__.Notify(61367)
	return geos.Overlaps(a.EWKB(), b.EWKB())
}

type PointPolygonOrder int

const (
	NotPointAndPolygon PointPolygonOrder = iota

	PointAndPolygon

	PolygonAndPoint
)

func PointKindAndPolygonKind(
	a geo.Geometry, b geo.Geometry,
) (PointPolygonOrder, geo.Geometry, geo.Geometry) {
	__antithesis_instrumentation__.Notify(61372)
	switch a.ShapeType2D() {
	case geopb.ShapeType_Point, geopb.ShapeType_MultiPoint:
		__antithesis_instrumentation__.Notify(61374)
		switch b.ShapeType2D() {
		case geopb.ShapeType_Polygon, geopb.ShapeType_MultiPolygon:
			__antithesis_instrumentation__.Notify(61377)
			return PointAndPolygon, a, b
		default:
			__antithesis_instrumentation__.Notify(61378)
		}
	case geopb.ShapeType_Polygon, geopb.ShapeType_MultiPolygon:
		__antithesis_instrumentation__.Notify(61375)
		switch b.ShapeType2D() {
		case geopb.ShapeType_Point, geopb.ShapeType_MultiPoint:
			__antithesis_instrumentation__.Notify(61379)
			return PolygonAndPoint, b, a
		default:
			__antithesis_instrumentation__.Notify(61380)
		}
	default:
		__antithesis_instrumentation__.Notify(61376)
	}
	__antithesis_instrumentation__.Notify(61373)
	return NotPointAndPolygon, a, b
}

func Touches(a geo.Geometry, b geo.Geometry) (bool, error) {
	__antithesis_instrumentation__.Notify(61381)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(61384)
		return false, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(61385)
	}
	__antithesis_instrumentation__.Notify(61382)
	if !a.CartesianBoundingBox().Intersects(b.CartesianBoundingBox()) {
		__antithesis_instrumentation__.Notify(61386)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(61387)
	}
	__antithesis_instrumentation__.Notify(61383)
	return geos.Touches(a.EWKB(), b.EWKB())
}

func Within(a geo.Geometry, b geo.Geometry) (bool, error) {
	__antithesis_instrumentation__.Notify(61388)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(61392)
		return false, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(61393)
	}
	__antithesis_instrumentation__.Notify(61389)
	if !b.CartesianBoundingBox().Covers(a.CartesianBoundingBox()) {
		__antithesis_instrumentation__.Notify(61394)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(61395)
	}
	__antithesis_instrumentation__.Notify(61390)

	pointPolygonPair, pointKind, polygonKind := PointKindAndPolygonKind(a, b)
	switch pointPolygonPair {
	case PolygonAndPoint:
		__antithesis_instrumentation__.Notify(61396)

		return false, nil
	case PointAndPolygon:
		__antithesis_instrumentation__.Notify(61397)
		return PointKindWithinPolygonKind(pointKind, polygonKind)
	default:
		__antithesis_instrumentation__.Notify(61398)
	}
	__antithesis_instrumentation__.Notify(61391)

	return geos.Within(a.EWKB(), b.EWKB())
}
