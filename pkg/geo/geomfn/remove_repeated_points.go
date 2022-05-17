package geomfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/errors"
	"github.com/twpayne/go-geom"
)

func RemoveRepeatedPoints(g geo.Geometry, tolerance float64) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(62829)
	t, err := g.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(62832)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(62833)
	}
	__antithesis_instrumentation__.Notify(62830)

	t, err = removeRepeatedPointsFromGeomT(t, tolerance*tolerance)
	if err != nil {
		__antithesis_instrumentation__.Notify(62834)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(62835)
	}
	__antithesis_instrumentation__.Notify(62831)
	return geo.MakeGeometryFromGeomT(t)
}

func removeRepeatedPointsFromGeomT(t geom.T, tolerance2 float64) (geom.T, error) {
	__antithesis_instrumentation__.Notify(62836)
	switch t := t.(type) {
	case *geom.Point:
		__antithesis_instrumentation__.Notify(62838)
	case *geom.LineString:
		__antithesis_instrumentation__.Notify(62839)
		if coords, modified := removeRepeatedCoords(t.Layout(), t.Coords(), tolerance2, 2); modified {
			__antithesis_instrumentation__.Notify(62846)
			return t.SetCoords(coords)
		} else {
			__antithesis_instrumentation__.Notify(62847)
		}
	case *geom.Polygon:
		__antithesis_instrumentation__.Notify(62840)
		if coords, modified := removeRepeatedCoords2(t.Layout(), t.Coords(), tolerance2, 4); modified {
			__antithesis_instrumentation__.Notify(62848)
			return t.SetCoords(coords)
		} else {
			__antithesis_instrumentation__.Notify(62849)
		}
	case *geom.MultiPoint:
		__antithesis_instrumentation__.Notify(62841)
		if coords, modified := removeRepeatedCoords(t.Layout(), t.Coords(), tolerance2, 0); modified {
			__antithesis_instrumentation__.Notify(62850)
			return t.SetCoords(coords)
		} else {
			__antithesis_instrumentation__.Notify(62851)
		}
	case *geom.MultiLineString:
		__antithesis_instrumentation__.Notify(62842)
		if coords, modified := removeRepeatedCoords2(t.Layout(), t.Coords(), tolerance2, 2); modified {
			__antithesis_instrumentation__.Notify(62852)
			return t.SetCoords(coords)
		} else {
			__antithesis_instrumentation__.Notify(62853)
		}
	case *geom.MultiPolygon:
		__antithesis_instrumentation__.Notify(62843)
		if coords, modified := removeRepeatedCoords3(t.Layout(), t.Coords(), tolerance2, 4); modified {
			__antithesis_instrumentation__.Notify(62854)
			return t.SetCoords(coords)
		} else {
			__antithesis_instrumentation__.Notify(62855)
		}
	case *geom.GeometryCollection:
		__antithesis_instrumentation__.Notify(62844)
		for _, g := range t.Geoms() {
			__antithesis_instrumentation__.Notify(62856)
			if _, err := removeRepeatedPointsFromGeomT(g, tolerance2); err != nil {
				__antithesis_instrumentation__.Notify(62857)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(62858)
			}
		}
	default:
		__antithesis_instrumentation__.Notify(62845)
		return nil, errors.AssertionFailedf("unknown geometry type: %T", t)
	}
	__antithesis_instrumentation__.Notify(62837)
	return t, nil
}

func removeRepeatedCoords(
	layout geom.Layout, coords []geom.Coord, tolerance2 float64, minCoords int,
) ([]geom.Coord, bool) {
	__antithesis_instrumentation__.Notify(62859)
	modified := false
	switch tolerance2 {
	case 0:
		__antithesis_instrumentation__.Notify(62861)
		for i := 1; i < len(coords) && func() bool {
			__antithesis_instrumentation__.Notify(62863)
			return len(coords) > minCoords == true
		}() == true; i++ {
			__antithesis_instrumentation__.Notify(62864)
			if coords[i].Equal(layout, coords[i-1]) {
				__antithesis_instrumentation__.Notify(62865)
				coords = append(coords[:i], coords[i+1:]...)
				modified = true
				i--
			} else {
				__antithesis_instrumentation__.Notify(62866)
			}
		}
	default:
		__antithesis_instrumentation__.Notify(62862)
		for i := 1; i < len(coords) && func() bool {
			__antithesis_instrumentation__.Notify(62867)
			return len(coords) > minCoords == true
		}() == true; i++ {
			__antithesis_instrumentation__.Notify(62868)
			if coordMag2(coordSub(coords[i], coords[i-1])) <= tolerance2 {
				__antithesis_instrumentation__.Notify(62869)
				coords = append(coords[:i], coords[i+1:]...)
				modified = true
				i--
			} else {
				__antithesis_instrumentation__.Notify(62870)
			}
		}
	}
	__antithesis_instrumentation__.Notify(62860)
	return coords, modified
}

func removeRepeatedCoords2(
	layout geom.Layout, coords2 [][]geom.Coord, tolerance2 float64, minCoords int,
) ([][]geom.Coord, bool) {
	__antithesis_instrumentation__.Notify(62871)
	modified := false
	for i, coords := range coords2 {
		__antithesis_instrumentation__.Notify(62873)
		if c, m := removeRepeatedCoords(layout, coords, tolerance2, minCoords); m {
			__antithesis_instrumentation__.Notify(62874)
			coords2[i] = c
			modified = true
		} else {
			__antithesis_instrumentation__.Notify(62875)
		}
	}
	__antithesis_instrumentation__.Notify(62872)
	return coords2, modified
}

func removeRepeatedCoords3(
	layout geom.Layout, coords3 [][][]geom.Coord, tolerance2 float64, minCoords int,
) ([][][]geom.Coord, bool) {
	__antithesis_instrumentation__.Notify(62876)
	modified := false
	for i, coords2 := range coords3 {
		__antithesis_instrumentation__.Notify(62878)
		for j, coords := range coords2 {
			__antithesis_instrumentation__.Notify(62879)
			if c, m := removeRepeatedCoords(layout, coords, tolerance2, minCoords); m {
				__antithesis_instrumentation__.Notify(62880)
				coords3[i][j] = c
				modified = true
			} else {
				__antithesis_instrumentation__.Notify(62881)
			}
		}
	}
	__antithesis_instrumentation__.Notify(62877)
	return coords3, modified
}
