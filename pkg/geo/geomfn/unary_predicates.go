package geomfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/geo/geos"
	"github.com/cockroachdb/errors"
	"github.com/twpayne/go-geom"
)

func IsClosed(g geo.Geometry) (bool, error) {
	__antithesis_instrumentation__.Notify(63553)
	t, err := g.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(63555)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(63556)
	}
	__antithesis_instrumentation__.Notify(63554)
	return isClosedFromGeomT(t)
}

func isClosedFromGeomT(t geom.T) (bool, error) {
	__antithesis_instrumentation__.Notify(63557)
	if t.Empty() {
		__antithesis_instrumentation__.Notify(63559)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(63560)
	}
	__antithesis_instrumentation__.Notify(63558)
	switch t := t.(type) {
	case *geom.Point, *geom.MultiPoint:
		__antithesis_instrumentation__.Notify(63561)
		return true, nil
	case *geom.LinearRing:
		__antithesis_instrumentation__.Notify(63562)
		return t.Coord(0).Equal(t.Layout(), t.Coord(t.NumCoords()-1)), nil
	case *geom.LineString:
		__antithesis_instrumentation__.Notify(63563)
		return t.Coord(0).Equal(t.Layout(), t.Coord(t.NumCoords()-1)), nil
	case *geom.MultiLineString:
		__antithesis_instrumentation__.Notify(63564)
		for i := 0; i < t.NumLineStrings(); i++ {
			__antithesis_instrumentation__.Notify(63573)
			if closed, err := isClosedFromGeomT(t.LineString(i)); err != nil || func() bool {
				__antithesis_instrumentation__.Notify(63574)
				return !closed == true
			}() == true {
				__antithesis_instrumentation__.Notify(63575)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(63576)
			}
		}
		__antithesis_instrumentation__.Notify(63565)
		return true, nil
	case *geom.Polygon:
		__antithesis_instrumentation__.Notify(63566)
		for i := 0; i < t.NumLinearRings(); i++ {
			__antithesis_instrumentation__.Notify(63577)
			if closed, err := isClosedFromGeomT(t.LinearRing(i)); err != nil || func() bool {
				__antithesis_instrumentation__.Notify(63578)
				return !closed == true
			}() == true {
				__antithesis_instrumentation__.Notify(63579)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(63580)
			}
		}
		__antithesis_instrumentation__.Notify(63567)
		return true, nil
	case *geom.MultiPolygon:
		__antithesis_instrumentation__.Notify(63568)
		for i := 0; i < t.NumPolygons(); i++ {
			__antithesis_instrumentation__.Notify(63581)
			if closed, err := isClosedFromGeomT(t.Polygon(i)); err != nil || func() bool {
				__antithesis_instrumentation__.Notify(63582)
				return !closed == true
			}() == true {
				__antithesis_instrumentation__.Notify(63583)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(63584)
			}
		}
		__antithesis_instrumentation__.Notify(63569)
		return true, nil
	case *geom.GeometryCollection:
		__antithesis_instrumentation__.Notify(63570)
		for _, g := range t.Geoms() {
			__antithesis_instrumentation__.Notify(63585)
			if closed, err := isClosedFromGeomT(g); err != nil || func() bool {
				__antithesis_instrumentation__.Notify(63586)
				return !closed == true
			}() == true {
				__antithesis_instrumentation__.Notify(63587)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(63588)
			}
		}
		__antithesis_instrumentation__.Notify(63571)
		return true, nil
	default:
		__antithesis_instrumentation__.Notify(63572)
		return false, errors.AssertionFailedf("unknown geometry type: %T", t)
	}
}

func IsCollection(g geo.Geometry) (bool, error) {
	__antithesis_instrumentation__.Notify(63589)
	switch g.ShapeType() {
	case geopb.ShapeType_MultiPoint, geopb.ShapeType_MultiLineString, geopb.ShapeType_MultiPolygon,
		geopb.ShapeType_GeometryCollection:
		__antithesis_instrumentation__.Notify(63590)
		return true, nil
	default:
		__antithesis_instrumentation__.Notify(63591)
		return false, nil
	}
}

func IsEmpty(g geo.Geometry) (bool, error) {
	__antithesis_instrumentation__.Notify(63592)
	return g.Empty(), nil
}

func IsRing(g geo.Geometry) (bool, error) {
	__antithesis_instrumentation__.Notify(63593)

	if g.Empty() {
		__antithesis_instrumentation__.Notify(63598)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(63599)
	}
	__antithesis_instrumentation__.Notify(63594)
	if g.ShapeType() != geopb.ShapeType_LineString {
		__antithesis_instrumentation__.Notify(63600)
		t, err := g.AsGeomT()
		if err != nil {
			__antithesis_instrumentation__.Notify(63602)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(63603)
		}
		__antithesis_instrumentation__.Notify(63601)
		e := geom.ErrUnsupportedType{Value: t}
		return false, errors.Wrap(e, "should only be called on a linear feature")
	} else {
		__antithesis_instrumentation__.Notify(63604)
	}
	__antithesis_instrumentation__.Notify(63595)
	if closed, err := IsClosed(g); err != nil || func() bool {
		__antithesis_instrumentation__.Notify(63605)
		return !closed == true
	}() == true {
		__antithesis_instrumentation__.Notify(63606)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(63607)
	}
	__antithesis_instrumentation__.Notify(63596)
	if simple, err := IsSimple(g); err != nil || func() bool {
		__antithesis_instrumentation__.Notify(63608)
		return !simple == true
	}() == true {
		__antithesis_instrumentation__.Notify(63609)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(63610)
	}
	__antithesis_instrumentation__.Notify(63597)
	return true, nil
}

func IsSimple(g geo.Geometry) (bool, error) {
	__antithesis_instrumentation__.Notify(63611)
	return geos.IsSimple(g.EWKB())
}
