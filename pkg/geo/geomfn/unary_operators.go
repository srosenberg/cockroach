package geomfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geos"
	"github.com/cockroachdb/errors"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/ewkb"
)

func Length(g geo.Geometry) (float64, error) {
	__antithesis_instrumentation__.Notify(63449)
	geomRepr, err := g.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(63452)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(63453)
	}
	__antithesis_instrumentation__.Notify(63450)

	switch geomRepr.(type) {
	case *geom.LineString, *geom.MultiLineString:
		__antithesis_instrumentation__.Notify(63454)
		return geos.Length(g.EWKB())
	}
	__antithesis_instrumentation__.Notify(63451)
	return lengthFromGeomT(geomRepr)
}

func lengthFromGeomT(geomRepr geom.T) (float64, error) {
	__antithesis_instrumentation__.Notify(63455)

	switch geomRepr := geomRepr.(type) {
	case *geom.Point, *geom.MultiPoint, *geom.Polygon, *geom.MultiPolygon:
		__antithesis_instrumentation__.Notify(63456)
		return 0, nil
	case *geom.LineString, *geom.MultiLineString:
		__antithesis_instrumentation__.Notify(63457)
		ewkb, err := ewkb.Marshal(geomRepr, geo.DefaultEWKBEncodingFormat)
		if err != nil {
			__antithesis_instrumentation__.Notify(63462)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(63463)
		}
		__antithesis_instrumentation__.Notify(63458)
		return geos.Length(ewkb)
	case *geom.GeometryCollection:
		__antithesis_instrumentation__.Notify(63459)
		total := float64(0)
		for _, subG := range geomRepr.Geoms() {
			__antithesis_instrumentation__.Notify(63464)
			subLength, err := lengthFromGeomT(subG)
			if err != nil {
				__antithesis_instrumentation__.Notify(63466)
				return 0, err
			} else {
				__antithesis_instrumentation__.Notify(63467)
			}
			__antithesis_instrumentation__.Notify(63465)
			total += subLength
		}
		__antithesis_instrumentation__.Notify(63460)
		return total, nil
	default:
		__antithesis_instrumentation__.Notify(63461)
		return 0, errors.AssertionFailedf("unknown geometry type: %T", geomRepr)
	}
}

func Perimeter(g geo.Geometry) (float64, error) {
	__antithesis_instrumentation__.Notify(63468)
	geomRepr, err := g.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(63471)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(63472)
	}
	__antithesis_instrumentation__.Notify(63469)

	switch geomRepr.(type) {
	case *geom.Polygon, *geom.MultiPolygon:
		__antithesis_instrumentation__.Notify(63473)
		return geos.Length(g.EWKB())
	}
	__antithesis_instrumentation__.Notify(63470)
	return perimeterFromGeomT(geomRepr)
}

func perimeterFromGeomT(geomRepr geom.T) (float64, error) {
	__antithesis_instrumentation__.Notify(63474)

	switch geomRepr := geomRepr.(type) {
	case *geom.Point, *geom.MultiPoint, *geom.LineString, *geom.MultiLineString:
		__antithesis_instrumentation__.Notify(63475)
		return 0, nil
	case *geom.Polygon, *geom.MultiPolygon:
		__antithesis_instrumentation__.Notify(63476)
		ewkb, err := ewkb.Marshal(geomRepr, geo.DefaultEWKBEncodingFormat)
		if err != nil {
			__antithesis_instrumentation__.Notify(63481)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(63482)
		}
		__antithesis_instrumentation__.Notify(63477)
		return geos.Length(ewkb)
	case *geom.GeometryCollection:
		__antithesis_instrumentation__.Notify(63478)
		total := float64(0)
		for _, subG := range geomRepr.Geoms() {
			__antithesis_instrumentation__.Notify(63483)
			subLength, err := perimeterFromGeomT(subG)
			if err != nil {
				__antithesis_instrumentation__.Notify(63485)
				return 0, err
			} else {
				__antithesis_instrumentation__.Notify(63486)
			}
			__antithesis_instrumentation__.Notify(63484)
			total += subLength
		}
		__antithesis_instrumentation__.Notify(63479)
		return total, nil
	default:
		__antithesis_instrumentation__.Notify(63480)
		return 0, errors.AssertionFailedf("unknown geometry type: %T", geomRepr)
	}
}

func Area(g geo.Geometry) (float64, error) {
	__antithesis_instrumentation__.Notify(63487)
	return geos.Area(g.EWKB())
}

func Dimension(g geo.Geometry) (int, error) {
	__antithesis_instrumentation__.Notify(63488)
	t, err := g.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(63490)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(63491)
	}
	__antithesis_instrumentation__.Notify(63489)
	return dimensionFromGeomT(t)
}

func dimensionFromGeomT(geomRepr geom.T) (int, error) {
	__antithesis_instrumentation__.Notify(63492)
	switch geomRepr := geomRepr.(type) {
	case *geom.Point, *geom.MultiPoint:
		__antithesis_instrumentation__.Notify(63493)
		return 0, nil
	case *geom.LineString, *geom.MultiLineString:
		__antithesis_instrumentation__.Notify(63494)
		return 1, nil
	case *geom.Polygon, *geom.MultiPolygon:
		__antithesis_instrumentation__.Notify(63495)
		return 2, nil
	case *geom.GeometryCollection:
		__antithesis_instrumentation__.Notify(63496)
		maxDim := 0
		for _, g := range geomRepr.Geoms() {
			__antithesis_instrumentation__.Notify(63499)
			dim, err := dimensionFromGeomT(g)
			if err != nil {
				__antithesis_instrumentation__.Notify(63501)
				return 0, err
			} else {
				__antithesis_instrumentation__.Notify(63502)
			}
			__antithesis_instrumentation__.Notify(63500)
			if dim > maxDim {
				__antithesis_instrumentation__.Notify(63503)
				maxDim = dim
			} else {
				__antithesis_instrumentation__.Notify(63504)
			}
		}
		__antithesis_instrumentation__.Notify(63497)
		return maxDim, nil
	default:
		__antithesis_instrumentation__.Notify(63498)
		return 0, errors.AssertionFailedf("unknown geometry type: %T", geomRepr)
	}
}

func Points(g geo.Geometry) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(63505)
	t, err := g.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(63509)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(63510)
	}
	__antithesis_instrumentation__.Notify(63506)
	layout := t.Layout()
	if gc, ok := t.(*geom.GeometryCollection); ok && func() bool {
		__antithesis_instrumentation__.Notify(63511)
		return gc.Empty() == true
	}() == true {
		__antithesis_instrumentation__.Notify(63512)
		layout = geom.XY
	} else {
		__antithesis_instrumentation__.Notify(63513)
	}
	__antithesis_instrumentation__.Notify(63507)
	points := geom.NewMultiPoint(layout).SetSRID(t.SRID())
	iter := geo.NewGeomTIterator(t, geo.EmptyBehaviorOmit)
	for {
		__antithesis_instrumentation__.Notify(63514)
		geomRepr, hasNext, err := iter.Next()
		if err != nil {
			__antithesis_instrumentation__.Notify(63516)
			return geo.Geometry{}, err
		} else {
			__antithesis_instrumentation__.Notify(63517)
			if !hasNext {
				__antithesis_instrumentation__.Notify(63518)
				break
			} else {
				__antithesis_instrumentation__.Notify(63519)
				if geomRepr.Empty() {
					__antithesis_instrumentation__.Notify(63520)
					continue
				} else {
					__antithesis_instrumentation__.Notify(63521)
				}
			}
		}
		__antithesis_instrumentation__.Notify(63515)
		switch geomRepr := geomRepr.(type) {
		case *geom.Point:
			__antithesis_instrumentation__.Notify(63522)
			if err = pushCoord(points, geomRepr.Coords()); err != nil {
				__antithesis_instrumentation__.Notify(63526)
				return geo.Geometry{}, err
			} else {
				__antithesis_instrumentation__.Notify(63527)
			}
		case *geom.LineString:
			__antithesis_instrumentation__.Notify(63523)
			for i := 0; i < geomRepr.NumCoords(); i++ {
				__antithesis_instrumentation__.Notify(63528)
				if err = pushCoord(points, geomRepr.Coord(i)); err != nil {
					__antithesis_instrumentation__.Notify(63529)
					return geo.Geometry{}, err
				} else {
					__antithesis_instrumentation__.Notify(63530)
				}
			}
		case *geom.Polygon:
			__antithesis_instrumentation__.Notify(63524)
			for i := 0; i < geomRepr.NumLinearRings(); i++ {
				__antithesis_instrumentation__.Notify(63531)
				linearRing := geomRepr.LinearRing(i)
				for j := 0; j < linearRing.NumCoords(); j++ {
					__antithesis_instrumentation__.Notify(63532)
					if err = pushCoord(points, linearRing.Coord(j)); err != nil {
						__antithesis_instrumentation__.Notify(63533)
						return geo.Geometry{}, err
					} else {
						__antithesis_instrumentation__.Notify(63534)
					}
				}
			}
		default:
			__antithesis_instrumentation__.Notify(63525)
			return geo.Geometry{}, errors.AssertionFailedf("unexpected type: %T", geomRepr)
		}
	}
	__antithesis_instrumentation__.Notify(63508)
	return geo.MakeGeometryFromGeomT(points)
}

func pushCoord(points *geom.MultiPoint, coord geom.Coord) error {
	__antithesis_instrumentation__.Notify(63535)
	point, err := geom.NewPoint(points.Layout()).SetCoords(coord)
	if err != nil {
		__antithesis_instrumentation__.Notify(63537)
		return err
	} else {
		__antithesis_instrumentation__.Notify(63538)
	}
	__antithesis_instrumentation__.Notify(63536)
	return points.Push(point)
}

func Normalize(g geo.Geometry) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(63539)
	retEWKB, err := geos.Normalize(g.EWKB())
	if err != nil {
		__antithesis_instrumentation__.Notify(63541)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(63542)
	}
	__antithesis_instrumentation__.Notify(63540)
	return geo.ParseGeometryFromEWKB(retEWKB)
}

func MinimumClearance(g geo.Geometry) (float64, error) {
	__antithesis_instrumentation__.Notify(63543)
	return geos.MinimumClearance(g.EWKB())
}

func MinimumClearanceLine(g geo.Geometry) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(63544)
	retEWKB, err := geos.MinimumClearanceLine(g.EWKB())
	if err != nil {
		__antithesis_instrumentation__.Notify(63546)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(63547)
	}
	__antithesis_instrumentation__.Notify(63545)
	return geo.ParseGeometryFromEWKB(retEWKB)
}

func CountVertices(t geom.T) int {
	__antithesis_instrumentation__.Notify(63548)
	switch t := t.(type) {
	case *geom.GeometryCollection:
		__antithesis_instrumentation__.Notify(63549)

		numPoints := 0
		for _, g := range t.Geoms() {
			__antithesis_instrumentation__.Notify(63552)
			numPoints += CountVertices(g)
		}
		__antithesis_instrumentation__.Notify(63550)
		return numPoints
	default:
		__antithesis_instrumentation__.Notify(63551)
		return len(t.FlatCoords()) / t.Stride()
	}
}
