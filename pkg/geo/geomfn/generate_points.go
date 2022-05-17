package geomfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"
	"math/rand"

	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/geo/geos"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/errors"
	"github.com/twpayne/go-geom"
)

const maxAllowedGridSize = 100 * geo.MaxAllowedSplitPoints

var ErrGenerateRandomPointsInvalidPoints = pgerror.Newf(
	pgcode.InvalidParameterValue,
	"points must be positive and geometry must not be empty",
)

func GenerateRandomPoints(g geo.Geometry, nPoints int, rng *rand.Rand) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(62115)
	var generateRandomPointsFunction func(g geo.Geometry, nPoints int, rng *rand.Rand) (*geom.MultiPoint, error)
	switch g.ShapeType() {
	case geopb.ShapeType_Polygon:
		__antithesis_instrumentation__.Notify(62123)
		generateRandomPointsFunction = generateRandomPointsFromPolygon
	case geopb.ShapeType_MultiPolygon:
		__antithesis_instrumentation__.Notify(62124)
		generateRandomPointsFunction = generateRandomPointsFromMultiPolygon
	default:
		__antithesis_instrumentation__.Notify(62125)
		return geo.Geometry{}, pgerror.Newf(pgcode.InvalidParameterValue, "unsupported type: %v", g.ShapeType().String())
	}
	__antithesis_instrumentation__.Notify(62116)
	if nPoints <= 0 {
		__antithesis_instrumentation__.Notify(62126)
		return geo.Geometry{}, ErrGenerateRandomPointsInvalidPoints
	} else {
		__antithesis_instrumentation__.Notify(62127)
	}
	__antithesis_instrumentation__.Notify(62117)
	if nPoints > geo.MaxAllowedSplitPoints {
		__antithesis_instrumentation__.Notify(62128)
		return geo.Geometry{}, pgerror.Newf(
			pgcode.InvalidParameterValue,
			"failed to generate random points, too many points to generate: requires %d points, max %d",
			nPoints,
			geo.MaxAllowedSplitPoints,
		)
	} else {
		__antithesis_instrumentation__.Notify(62129)
	}
	__antithesis_instrumentation__.Notify(62118)
	empty, err := IsEmpty(g)
	if err != nil {
		__antithesis_instrumentation__.Notify(62130)
		return geo.Geometry{}, errors.Wrap(err, "could not check if geometry is empty")
	} else {
		__antithesis_instrumentation__.Notify(62131)
	}
	__antithesis_instrumentation__.Notify(62119)
	if empty {
		__antithesis_instrumentation__.Notify(62132)
		return geo.Geometry{}, ErrGenerateRandomPointsInvalidPoints
	} else {
		__antithesis_instrumentation__.Notify(62133)
	}
	__antithesis_instrumentation__.Notify(62120)
	mpt, err := generateRandomPointsFunction(g, nPoints, rng)
	if err != nil {
		__antithesis_instrumentation__.Notify(62134)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(62135)
	}
	__antithesis_instrumentation__.Notify(62121)
	srid := g.SRID()
	mpt.SetSRID(int(srid))
	out, err := geo.MakeGeometryFromGeomT(mpt)
	if err != nil {
		__antithesis_instrumentation__.Notify(62136)
		return geo.Geometry{}, errors.Wrap(err, "could not transform geom.T into geometry")
	} else {
		__antithesis_instrumentation__.Notify(62137)
	}
	__antithesis_instrumentation__.Notify(62122)
	return out, nil
}

func generateRandomPointsFromPolygon(
	g geo.Geometry, nPoints int, rng *rand.Rand,
) (*geom.MultiPoint, error) {
	__antithesis_instrumentation__.Notify(62138)
	area, err := Area(g)
	if err != nil {
		__antithesis_instrumentation__.Notify(62148)
		return nil, errors.Wrap(err, "could not calculate Polygon area")
	} else {
		__antithesis_instrumentation__.Notify(62149)
	}
	__antithesis_instrumentation__.Notify(62139)
	if area == 0.0 {
		__antithesis_instrumentation__.Notify(62150)
		return nil, pgerror.Newf(pgcode.InvalidParameterValue, "zero area input Polygon")
	} else {
		__antithesis_instrumentation__.Notify(62151)
	}
	__antithesis_instrumentation__.Notify(62140)
	bbox := g.CartesianBoundingBox()
	bboxWidth := bbox.HiX - bbox.LoX
	bboxHeight := bbox.HiY - bbox.LoY
	bboxArea := bboxHeight * bboxWidth

	sampleNPoints := float64(nPoints) * bboxArea / area

	sampleSqrt := math.Round(math.Sqrt(sampleNPoints))
	if sampleSqrt == 0 {
		__antithesis_instrumentation__.Notify(62152)
		sampleSqrt = 1
	} else {
		__antithesis_instrumentation__.Notify(62153)
	}
	__antithesis_instrumentation__.Notify(62141)
	var sampleHeight, sampleWidth int
	var sampleCellSize float64

	if bboxWidth > bboxHeight {
		__antithesis_instrumentation__.Notify(62154)
		sampleWidth = int(sampleSqrt)
		sampleHeight = int(math.Ceil(sampleNPoints / float64(sampleWidth)))
		sampleCellSize = bboxWidth / float64(sampleWidth)
	} else {
		__antithesis_instrumentation__.Notify(62155)
		sampleHeight = int(sampleSqrt)
		sampleWidth = int(math.Ceil(sampleNPoints / float64(sampleHeight)))
		sampleCellSize = bboxHeight / float64(sampleHeight)
	}
	__antithesis_instrumentation__.Notify(62142)
	n := sampleHeight * sampleWidth
	if n > maxAllowedGridSize {
		__antithesis_instrumentation__.Notify(62156)
		return nil, pgerror.Newf(
			pgcode.InvalidParameterValue,
			"generated area is too large: %d, max %d",
			n,
			maxAllowedGridSize,
		)
	} else {
		__antithesis_instrumentation__.Notify(62157)
	}
	__antithesis_instrumentation__.Notify(62143)

	gPrep, err := geos.PrepareGeometry(g.EWKB())
	if err != nil {
		__antithesis_instrumentation__.Notify(62158)
		return nil, errors.Wrap(err, "could not prepare geometry")
	} else {
		__antithesis_instrumentation__.Notify(62159)
	}
	__antithesis_instrumentation__.Notify(62144)
	defer geos.PreparedGeomDestroy(gPrep)

	cells := make([]geom.Coord, n)
	for i := 0; i < sampleWidth; i++ {
		__antithesis_instrumentation__.Notify(62160)
		for j := 0; j < sampleHeight; j++ {
			__antithesis_instrumentation__.Notify(62161)
			cells[i*sampleHeight+j] = geom.Coord{float64(i), float64(j)}
		}
	}
	__antithesis_instrumentation__.Notify(62145)

	if n > 1 {
		__antithesis_instrumentation__.Notify(62162)
		rng.Shuffle(n, func(i int, j int) {
			__antithesis_instrumentation__.Notify(62163)
			temp := cells[j]
			cells[j] = cells[i]
			cells[i] = temp
		})
	} else {
		__antithesis_instrumentation__.Notify(62164)
	}
	__antithesis_instrumentation__.Notify(62146)
	results := geom.NewMultiPoint(geom.XY)

	for nPointsGenerated, iterations := 0, 0; nPointsGenerated < nPoints && func() bool {
		__antithesis_instrumentation__.Notify(62165)
		return iterations <= 100 == true
	}() == true; iterations++ {
		__antithesis_instrumentation__.Notify(62166)
		for _, cell := range cells {
			__antithesis_instrumentation__.Notify(62167)
			y := bbox.LoY + cell.X()*sampleCellSize
			x := bbox.LoX + cell.Y()*sampleCellSize
			x += rng.Float64() * sampleCellSize
			y += rng.Float64() * sampleCellSize
			if x > bbox.HiX || func() bool {
				__antithesis_instrumentation__.Notify(62171)
				return y > bbox.HiY == true
			}() == true {
				__antithesis_instrumentation__.Notify(62172)
				continue
			} else {
				__antithesis_instrumentation__.Notify(62173)
			}
			__antithesis_instrumentation__.Notify(62168)
			gpt, err := geo.MakeGeometryFromPointCoords(x, y)
			if err != nil {
				__antithesis_instrumentation__.Notify(62174)
				return nil, errors.Wrap(err, "could not create geometry Point")
			} else {
				__antithesis_instrumentation__.Notify(62175)
			}
			__antithesis_instrumentation__.Notify(62169)
			intersects, err := geos.PreparedIntersects(gPrep, gpt.EWKB())
			if err != nil {
				__antithesis_instrumentation__.Notify(62176)
				return nil, errors.Wrap(err, "could not check prepared intersection")
			} else {
				__antithesis_instrumentation__.Notify(62177)
			}
			__antithesis_instrumentation__.Notify(62170)
			if intersects {
				__antithesis_instrumentation__.Notify(62178)
				nPointsGenerated++
				p := geom.NewPointFlat(geom.XY, []float64{x, y})
				srid := g.SRID()
				p.SetSRID(int(srid))
				err = results.Push(p)
				if err != nil {
					__antithesis_instrumentation__.Notify(62180)
					return nil, errors.Wrap(err, "could not add point to the results")
				} else {
					__antithesis_instrumentation__.Notify(62181)
				}
				__antithesis_instrumentation__.Notify(62179)
				if nPointsGenerated == nPoints {
					__antithesis_instrumentation__.Notify(62182)
					return results, nil
				} else {
					__antithesis_instrumentation__.Notify(62183)
				}
			} else {
				__antithesis_instrumentation__.Notify(62184)
			}
		}
	}
	__antithesis_instrumentation__.Notify(62147)
	return results, nil
}

func generateRandomPointsFromMultiPolygon(
	g geo.Geometry, nPoints int, rng *rand.Rand,
) (*geom.MultiPoint, error) {
	__antithesis_instrumentation__.Notify(62185)
	results := geom.NewMultiPoint(geom.XY)

	area, err := Area(g)
	if err != nil {
		__antithesis_instrumentation__.Notify(62189)
		return nil, errors.Wrap(err, "could not calculate MultiPolygon area")
	} else {
		__antithesis_instrumentation__.Notify(62190)
	}
	__antithesis_instrumentation__.Notify(62186)

	gt, err := g.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(62191)
		return nil, errors.Wrap(err, "could not transform MultiPolygon into geom.T")
	} else {
		__antithesis_instrumentation__.Notify(62192)
	}
	__antithesis_instrumentation__.Notify(62187)

	gmp := gt.(*geom.MultiPolygon)
	for i := 0; i < gmp.NumPolygons(); i++ {
		__antithesis_instrumentation__.Notify(62193)
		poly := gmp.Polygon(i)
		subarea := poly.Area()
		subNPoints := int(math.Round(float64(nPoints) * subarea / area))
		if subNPoints > 0 {
			__antithesis_instrumentation__.Notify(62194)
			g, err := geo.MakeGeometryFromGeomT(poly)
			if err != nil {
				__antithesis_instrumentation__.Notify(62197)
				return nil, errors.Wrap(err, "could not transform geom.T into Geometry")
			} else {
				__antithesis_instrumentation__.Notify(62198)
			}
			__antithesis_instrumentation__.Notify(62195)
			subMPT, err := generateRandomPointsFromPolygon(g, subNPoints, rng)
			if err != nil {
				__antithesis_instrumentation__.Notify(62199)
				return nil, errors.Wrap(err, "error generating points for Polygon")
			} else {
				__antithesis_instrumentation__.Notify(62200)
			}
			__antithesis_instrumentation__.Notify(62196)
			for j := 0; j < subMPT.NumPoints(); j++ {
				__antithesis_instrumentation__.Notify(62201)
				if err := results.Push(subMPT.Point(j)); err != nil {
					__antithesis_instrumentation__.Notify(62202)
					return nil, errors.Wrap(err, "could not push point to the results")
				} else {
					__antithesis_instrumentation__.Notify(62203)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(62204)
		}
	}
	__antithesis_instrumentation__.Notify(62188)
	return results, nil
}
