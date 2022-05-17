// Package geogen provides utilities for generating various geospatial types.
package geogen

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"
	"math/rand"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/geo/geoprojbase"
	"github.com/cockroachdb/errors"
	"github.com/twpayne/go-geom"
)

var validShapeTypes = []geopb.ShapeType{
	geopb.ShapeType_Point,
	geopb.ShapeType_LineString,
	geopb.ShapeType_Polygon,
	geopb.ShapeType_MultiPoint,
	geopb.ShapeType_MultiLineString,
	geopb.ShapeType_MultiPolygon,
	geopb.ShapeType_GeometryCollection,
}

func RandomShapeType(rng *rand.Rand) geopb.ShapeType {
	__antithesis_instrumentation__.Notify(59331)
	return validShapeTypes[rng.Intn(len(validShapeTypes))]
}

var validLayouts = []geom.Layout{
	geom.XY,
	geom.XYM,
	geom.XYZ,
	geom.XYZM,
}

func RandomLayout(rng *rand.Rand, layout geom.Layout) geom.Layout {
	__antithesis_instrumentation__.Notify(59332)
	if layout != geom.NoLayout {
		__antithesis_instrumentation__.Notify(59334)
		return layout
	} else {
		__antithesis_instrumentation__.Notify(59335)
	}
	__antithesis_instrumentation__.Notify(59333)
	return validLayouts[rng.Intn(len(validLayouts))]
}

type RandomGeomBounds struct {
	minX, maxX float64
	minY, maxY float64
	minZ, maxZ float64
	minM, maxM float64
}

func MakeRandomGeomBoundsForGeography() RandomGeomBounds {
	__antithesis_instrumentation__.Notify(59336)
	randomBounds := MakeRandomGeomBounds()
	randomBounds.minX, randomBounds.maxX = -180.0, 180.0
	randomBounds.minY, randomBounds.maxY = -90.0, 90.0
	return randomBounds
}

func MakeRandomGeomBounds() RandomGeomBounds {
	__antithesis_instrumentation__.Notify(59337)
	return RandomGeomBounds{
		minX: -10e9, maxX: 10e9,
		minY: -10e9, maxY: 10e9,
		minZ: -10e9, maxZ: 10e9,
		minM: -10e9, maxM: 10e9,
	}
}

func RandomCoordSlice(rng *rand.Rand, randomBounds RandomGeomBounds, layout geom.Layout) []float64 {
	__antithesis_instrumentation__.Notify(59338)
	if layout == geom.NoLayout {
		__antithesis_instrumentation__.Notify(59342)
		panic(errors.Newf("must specify a layout for RandomCoordSlice"))
	} else {
		__antithesis_instrumentation__.Notify(59343)
	}
	__antithesis_instrumentation__.Notify(59339)
	var coords []float64
	coords = append(coords, RandomCoord(rng, randomBounds.minX, randomBounds.maxX))
	coords = append(coords, RandomCoord(rng, randomBounds.minY, randomBounds.maxY))

	if layout.ZIndex() != -1 {
		__antithesis_instrumentation__.Notify(59344)
		coords = append(coords, RandomCoord(rng, randomBounds.minZ, randomBounds.maxZ))
	} else {
		__antithesis_instrumentation__.Notify(59345)
	}
	__antithesis_instrumentation__.Notify(59340)
	if layout.MIndex() != -1 {
		__antithesis_instrumentation__.Notify(59346)
		coords = append(coords, RandomCoord(rng, randomBounds.minM, randomBounds.maxM))
	} else {
		__antithesis_instrumentation__.Notify(59347)
	}
	__antithesis_instrumentation__.Notify(59341)

	return coords
}

func RandomCoord(rng *rand.Rand, min float64, max float64) float64 {
	__antithesis_instrumentation__.Notify(59348)
	return rng.Float64()*(max-min) + min
}

func RandomValidLinearRingCoords(
	rng *rand.Rand, numPoints int, randomBounds RandomGeomBounds, layout geom.Layout,
) []geom.Coord {
	__antithesis_instrumentation__.Notify(59349)
	layout = RandomLayout(rng, layout)
	if numPoints < 3 {
		__antithesis_instrumentation__.Notify(59353)
		panic(errors.Newf("need at least 3 points, got %d", numPoints))
	} else {
		__antithesis_instrumentation__.Notify(59354)
	}
	__antithesis_instrumentation__.Notify(59350)

	coords := make([]geom.Coord, numPoints+1)
	var centerX, centerY float64
	for i := 0; i < numPoints; i++ {
		__antithesis_instrumentation__.Notify(59355)
		coords[i] = RandomCoordSlice(rng, randomBounds, layout)
		centerX += coords[i].X()
		centerY += coords[i].Y()
	}
	__antithesis_instrumentation__.Notify(59351)

	centerX /= float64(numPoints)
	centerY /= float64(numPoints)

	sort.Slice(coords[:numPoints], func(i, j int) bool {
		__antithesis_instrumentation__.Notify(59356)
		angleI := math.Atan2(coords[i].Y()-centerY, coords[i].X()-centerX)
		angleJ := math.Atan2(coords[j].Y()-centerY, coords[j].X()-centerX)
		return angleI < angleJ
	})
	__antithesis_instrumentation__.Notify(59352)

	coords[numPoints] = coords[0]
	return coords
}

func RandomPoint(
	rng *rand.Rand, randomBounds RandomGeomBounds, srid geopb.SRID, layout geom.Layout,
) *geom.Point {
	__antithesis_instrumentation__.Notify(59357)
	layout = RandomLayout(rng, layout)

	if rng.Intn(10) == 0 {
		__antithesis_instrumentation__.Notify(59359)
		return geom.NewPointEmpty(layout).SetSRID(int(srid))
	} else {
		__antithesis_instrumentation__.Notify(59360)
	}
	__antithesis_instrumentation__.Notify(59358)
	return geom.NewPointFlat(layout, RandomCoordSlice(rng, randomBounds, layout)).SetSRID(int(srid))
}

func RandomLineString(
	rng *rand.Rand, randomBounds RandomGeomBounds, srid geopb.SRID, layout geom.Layout,
) *geom.LineString {
	__antithesis_instrumentation__.Notify(59361)
	layout = RandomLayout(rng, layout)

	if rng.Intn(10) == 0 {
		__antithesis_instrumentation__.Notify(59364)
		return geom.NewLineString(layout).SetSRID(int(srid))
	} else {
		__antithesis_instrumentation__.Notify(59365)
	}
	__antithesis_instrumentation__.Notify(59362)
	numCoords := 3 + rand.Intn(10)
	randCoords := RandomValidLinearRingCoords(rng, numCoords, randomBounds, layout)

	var minTrunc, maxTrunc int

	for maxTrunc-minTrunc < 2 {
		__antithesis_instrumentation__.Notify(59366)
		minTrunc, maxTrunc = rand.Intn(numCoords+1), rand.Intn(numCoords+1)

		if minTrunc > maxTrunc {
			__antithesis_instrumentation__.Notify(59367)
			minTrunc, maxTrunc = maxTrunc, minTrunc
		} else {
			__antithesis_instrumentation__.Notify(59368)
		}
	}
	__antithesis_instrumentation__.Notify(59363)
	return geom.NewLineString(layout).MustSetCoords(randCoords[minTrunc:maxTrunc]).SetSRID(int(srid))
}

func RandomPolygon(
	rng *rand.Rand, randomBounds RandomGeomBounds, srid geopb.SRID, layout geom.Layout,
) *geom.Polygon {
	__antithesis_instrumentation__.Notify(59369)
	layout = RandomLayout(rng, layout)

	if rng.Intn(10) == 0 {
		__antithesis_instrumentation__.Notify(59371)
		return geom.NewPolygon(layout).SetSRID(int(srid))
	} else {
		__antithesis_instrumentation__.Notify(59372)
	}
	__antithesis_instrumentation__.Notify(59370)

	return geom.NewPolygon(layout).MustSetCoords([][]geom.Coord{
		RandomValidLinearRingCoords(rng, 3+rng.Intn(10), randomBounds, layout),
	}).SetSRID(int(srid))
}

func RandomMultiPoint(
	rng *rand.Rand, randomBounds RandomGeomBounds, srid geopb.SRID, layout geom.Layout,
) *geom.MultiPoint {
	__antithesis_instrumentation__.Notify(59373)
	layout = RandomLayout(rng, layout)
	ret := geom.NewMultiPoint(layout).SetSRID(int(srid))

	num := rng.Intn(10)
	for i := 0; i < num; i++ {
		__antithesis_instrumentation__.Notify(59375)

		var point *geom.Point
		for point == nil || func() bool {
			__antithesis_instrumentation__.Notify(59377)
			return point.Empty() == true
		}() == true {
			__antithesis_instrumentation__.Notify(59378)
			point = RandomPoint(rng, randomBounds, srid, layout)
		}
		__antithesis_instrumentation__.Notify(59376)
		if err := ret.Push(point); err != nil {
			__antithesis_instrumentation__.Notify(59379)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(59380)
		}
	}
	__antithesis_instrumentation__.Notify(59374)
	return ret
}

func RandomMultiLineString(
	rng *rand.Rand, randomBounds RandomGeomBounds, srid geopb.SRID, layout geom.Layout,
) *geom.MultiLineString {
	__antithesis_instrumentation__.Notify(59381)
	layout = RandomLayout(rng, layout)
	ret := geom.NewMultiLineString(layout).SetSRID(int(srid))

	num := rng.Intn(10)
	for i := 0; i < num; i++ {
		__antithesis_instrumentation__.Notify(59383)

		var lineString *geom.LineString
		for lineString == nil || func() bool {
			__antithesis_instrumentation__.Notify(59385)
			return lineString.Empty() == true
		}() == true {
			__antithesis_instrumentation__.Notify(59386)
			lineString = RandomLineString(rng, randomBounds, srid, layout)
		}
		__antithesis_instrumentation__.Notify(59384)
		if err := ret.Push(lineString); err != nil {
			__antithesis_instrumentation__.Notify(59387)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(59388)
		}
	}
	__antithesis_instrumentation__.Notify(59382)
	return ret
}

func RandomMultiPolygon(
	rng *rand.Rand, randomBounds RandomGeomBounds, srid geopb.SRID, layout geom.Layout,
) *geom.MultiPolygon {
	__antithesis_instrumentation__.Notify(59389)
	layout = RandomLayout(rng, layout)
	ret := geom.NewMultiPolygon(layout).SetSRID(int(srid))

	num := rng.Intn(10)
	for i := 0; i < num; i++ {
		__antithesis_instrumentation__.Notify(59391)

		var polygon *geom.Polygon
		for polygon == nil || func() bool {
			__antithesis_instrumentation__.Notify(59393)
			return polygon.Empty() == true
		}() == true {
			__antithesis_instrumentation__.Notify(59394)
			polygon = RandomPolygon(rng, randomBounds, srid, layout)
		}
		__antithesis_instrumentation__.Notify(59392)
		if err := ret.Push(polygon); err != nil {
			__antithesis_instrumentation__.Notify(59395)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(59396)
		}
	}
	__antithesis_instrumentation__.Notify(59390)
	return ret
}

func RandomGeometryCollection(
	rng *rand.Rand, randomBounds RandomGeomBounds, srid geopb.SRID, layout geom.Layout,
) *geom.GeometryCollection {
	__antithesis_instrumentation__.Notify(59397)
	layout = RandomLayout(rng, layout)
	ret := geom.NewGeometryCollection().SetSRID(int(srid))
	if err := ret.SetLayout(layout); err != nil {
		__antithesis_instrumentation__.Notify(59400)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(59401)
	}
	__antithesis_instrumentation__.Notify(59398)

	num := rng.Intn(10)
	for i := 0; i < num; i++ {
		__antithesis_instrumentation__.Notify(59402)
		var shape geom.T
		needShape := true

		for needShape {
			__antithesis_instrumentation__.Notify(59404)
			shape = RandomGeomT(rng, randomBounds, srid, layout)
			_, needShape = shape.(*geom.GeometryCollection)

			if shape.Empty() {
				__antithesis_instrumentation__.Notify(59405)
				needShape = true
			} else {
				__antithesis_instrumentation__.Notify(59406)
			}
		}
		__antithesis_instrumentation__.Notify(59403)
		if err := ret.Push(shape); err != nil {
			__antithesis_instrumentation__.Notify(59407)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(59408)
		}
	}
	__antithesis_instrumentation__.Notify(59399)
	return ret
}

func RandomGeomT(
	rng *rand.Rand, randomBounds RandomGeomBounds, srid geopb.SRID, layout geom.Layout,
) geom.T {
	__antithesis_instrumentation__.Notify(59409)
	layout = RandomLayout(rng, layout)
	shapeType := RandomShapeType(rng)
	switch shapeType {
	case geopb.ShapeType_Point:
		__antithesis_instrumentation__.Notify(59411)
		return RandomPoint(rng, randomBounds, srid, layout)
	case geopb.ShapeType_LineString:
		__antithesis_instrumentation__.Notify(59412)
		return RandomLineString(rng, randomBounds, srid, layout)
	case geopb.ShapeType_Polygon:
		__antithesis_instrumentation__.Notify(59413)
		return RandomPolygon(rng, randomBounds, srid, layout)
	case geopb.ShapeType_MultiPoint:
		__antithesis_instrumentation__.Notify(59414)
		return RandomMultiPoint(rng, randomBounds, srid, layout)
	case geopb.ShapeType_MultiLineString:
		__antithesis_instrumentation__.Notify(59415)
		return RandomMultiLineString(rng, randomBounds, srid, layout)
	case geopb.ShapeType_MultiPolygon:
		__antithesis_instrumentation__.Notify(59416)
		return RandomMultiPolygon(rng, randomBounds, srid, layout)
	case geopb.ShapeType_GeometryCollection:
		__antithesis_instrumentation__.Notify(59417)
		return RandomGeometryCollection(rng, randomBounds, srid, layout)
	default:
		__antithesis_instrumentation__.Notify(59418)
	}
	__antithesis_instrumentation__.Notify(59410)
	panic(errors.Newf("unknown shape type: %v", shapeType))
}

func RandomGeometry(rng *rand.Rand, srid geopb.SRID) geo.Geometry {
	__antithesis_instrumentation__.Notify(59419)
	return RandomGeometryWithLayout(rng, srid, geom.NoLayout)
}

func RandomGeometryWithLayout(rng *rand.Rand, srid geopb.SRID, layout geom.Layout) geo.Geometry {
	__antithesis_instrumentation__.Notify(59420)
	randomBounds := MakeRandomGeomBounds()
	if srid != 0 {
		__antithesis_instrumentation__.Notify(59423)
		proj, err := geoprojbase.Projection(srid)
		if err != nil {
			__antithesis_instrumentation__.Notify(59425)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(59426)
		}
		__antithesis_instrumentation__.Notify(59424)
		randomBounds.minX, randomBounds.maxX = proj.Bounds.MinX, proj.Bounds.MaxX
		randomBounds.minY, randomBounds.maxY = proj.Bounds.MinY, proj.Bounds.MaxY
	} else {
		__antithesis_instrumentation__.Notify(59427)
	}
	__antithesis_instrumentation__.Notify(59421)
	ret, err := geo.MakeGeometryFromGeomT(RandomGeomT(rng, randomBounds, srid, layout))
	if err != nil {
		__antithesis_instrumentation__.Notify(59428)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(59429)
	}
	__antithesis_instrumentation__.Notify(59422)
	return ret
}

func RandomGeography(rng *rand.Rand, srid geopb.SRID) geo.Geography {
	__antithesis_instrumentation__.Notify(59430)
	return RandomGeographyWithLayout(rng, srid, geom.NoLayout)
}

func RandomGeographyWithLayout(rng *rand.Rand, srid geopb.SRID, layout geom.Layout) geo.Geography {
	__antithesis_instrumentation__.Notify(59431)

	randomBoundsGeography := MakeRandomGeomBoundsForGeography()
	ret, err := geo.MakeGeographyFromGeomT(RandomGeomT(rng, randomBoundsGeography, srid, layout))
	if err != nil {
		__antithesis_instrumentation__.Notify(59433)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(59434)
	}
	__antithesis_instrumentation__.Notify(59432)
	return ret
}
