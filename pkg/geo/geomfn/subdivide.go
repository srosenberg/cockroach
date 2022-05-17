package geomfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"

	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/errors"
	geom "github.com/twpayne/go-geom"
)

func Subdivide(g geo.Geometry, maxVertices int) ([]geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(63140)
	if g.Empty() {
		__antithesis_instrumentation__.Notify(63147)
		return []geo.Geometry{g}, nil
	} else {
		__antithesis_instrumentation__.Notify(63148)
	}
	__antithesis_instrumentation__.Notify(63141)
	const minMaxVertices = 5
	if maxVertices < minMaxVertices {
		__antithesis_instrumentation__.Notify(63149)
		return nil, pgerror.Newf(pgcode.InvalidParameterValue, "max_vertices number cannot be less than %v", minMaxVertices)
	} else {
		__antithesis_instrumentation__.Notify(63150)
	}
	__antithesis_instrumentation__.Notify(63142)

	gt, err := g.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(63151)
		return nil, errors.Wrap(err, "could not transform input geometry into geom.T")
	} else {
		__antithesis_instrumentation__.Notify(63152)
	}
	__antithesis_instrumentation__.Notify(63143)
	dim, err := dimensionFromGeomT(gt)
	if err != nil {
		__antithesis_instrumentation__.Notify(63153)
		return nil, errors.Wrap(err, "could not calculate geometry dimension")
	} else {
		__antithesis_instrumentation__.Notify(63154)
	}
	__antithesis_instrumentation__.Notify(63144)

	const maxDepth = 50
	const startDepth = 0
	geomTs, err := subdivideRecursive(gt, maxVertices, startDepth, dim, maxDepth)
	if err != nil {
		__antithesis_instrumentation__.Notify(63155)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(63156)
	}
	__antithesis_instrumentation__.Notify(63145)

	output := []geo.Geometry{}
	for _, cg := range geomTs {
		__antithesis_instrumentation__.Notify(63157)
		geo.AdjustGeomTSRID(cg, g.SRID())
		g, err := geo.MakeGeometryFromGeomT(cg)
		if err != nil {
			__antithesis_instrumentation__.Notify(63159)
			return []geo.Geometry{}, errors.Wrap(err, "could not transform output geom.T into geometry")
		} else {
			__antithesis_instrumentation__.Notify(63160)
		}
		__antithesis_instrumentation__.Notify(63158)
		output = append(output, g)
	}
	__antithesis_instrumentation__.Notify(63146)
	return output, nil
}

func subdivideRecursive(
	gt geom.T, maxVertices int, depth int, dim int, maxDepth int,
) ([]geom.T, error) {
	__antithesis_instrumentation__.Notify(63161)
	var results []geom.T
	clip := geo.BoundingBoxFromGeomTGeometryType(gt)
	width := clip.HiX - clip.LoX
	height := clip.HiY - clip.LoY
	if width == 0 && func() bool {
		__antithesis_instrumentation__.Notify(63177)
		return height == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(63178)

		if gt, ok := gt.(*geom.Point); ok && func() bool {
			__antithesis_instrumentation__.Notify(63180)
			return dim == 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(63181)
			results = append(results, gt)
		} else {
			__antithesis_instrumentation__.Notify(63182)
		}
		__antithesis_instrumentation__.Notify(63179)
		return results, nil
	} else {
		__antithesis_instrumentation__.Notify(63183)
	}
	__antithesis_instrumentation__.Notify(63162)

	const tolerance = 1e-12
	if width == 0 {
		__antithesis_instrumentation__.Notify(63184)
		clip.HiX += tolerance
		clip.LoX -= tolerance
		width = 2 * tolerance
	} else {
		__antithesis_instrumentation__.Notify(63185)
	}
	__antithesis_instrumentation__.Notify(63163)
	if height == 0 {
		__antithesis_instrumentation__.Notify(63186)
		clip.HiY += tolerance
		clip.LoY -= tolerance
		height = 2 * tolerance
	} else {
		__antithesis_instrumentation__.Notify(63187)
	}
	__antithesis_instrumentation__.Notify(63164)

	switch gt.(type) {
	case *geom.MultiLineString, *geom.MultiPolygon, *geom.GeometryCollection:
		__antithesis_instrumentation__.Notify(63188)
		err := forEachGeomTInCollectionForSubdivide(gt, func(gi geom.T) error {
			__antithesis_instrumentation__.Notify(63191)

			subdivisions, err := subdivideRecursive(gi, maxVertices, depth, dim, maxDepth)
			if err != nil {
				__antithesis_instrumentation__.Notify(63193)
				return err
			} else {
				__antithesis_instrumentation__.Notify(63194)
			}
			__antithesis_instrumentation__.Notify(63192)
			results = append(results, subdivisions...)
			return nil
		})
		__antithesis_instrumentation__.Notify(63189)
		if err != nil {
			__antithesis_instrumentation__.Notify(63195)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(63196)
		}
		__antithesis_instrumentation__.Notify(63190)
		return results, nil
	}
	__antithesis_instrumentation__.Notify(63165)
	currDim, err := dimensionFromGeomT(gt)
	if err != nil {
		__antithesis_instrumentation__.Notify(63197)
		return nil, errors.Wrap(err, "error checking geom.T dimension")
	} else {
		__antithesis_instrumentation__.Notify(63198)
	}
	__antithesis_instrumentation__.Notify(63166)
	if currDim < dim {
		__antithesis_instrumentation__.Notify(63199)

		return results, nil
	} else {
		__antithesis_instrumentation__.Notify(63200)
	}
	__antithesis_instrumentation__.Notify(63167)
	if depth > maxDepth {
		__antithesis_instrumentation__.Notify(63201)
		results = append(results, gt)
		return results, nil
	} else {
		__antithesis_instrumentation__.Notify(63202)
	}
	__antithesis_instrumentation__.Notify(63168)
	nVertices := CountVertices(gt)
	if nVertices == 0 {
		__antithesis_instrumentation__.Notify(63203)
		return results, nil
	} else {
		__antithesis_instrumentation__.Notify(63204)
	}
	__antithesis_instrumentation__.Notify(63169)
	if nVertices <= maxVertices {
		__antithesis_instrumentation__.Notify(63205)
		results = append(results, gt)
		return results, nil
	} else {
		__antithesis_instrumentation__.Notify(63206)
	}
	__antithesis_instrumentation__.Notify(63170)

	splitHorizontally := width > height
	splitPoint, err := calculateSplitPointCoordForSubdivide(gt, nVertices, splitHorizontally, *clip)
	if err != nil {
		__antithesis_instrumentation__.Notify(63207)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(63208)
	}
	__antithesis_instrumentation__.Notify(63171)
	subBox1 := *clip
	subBox2 := *clip
	if splitHorizontally {
		__antithesis_instrumentation__.Notify(63209)
		subBox1.HiX = splitPoint
		subBox2.LoX = splitPoint
	} else {
		__antithesis_instrumentation__.Notify(63210)
		subBox1.HiY = splitPoint
		subBox2.LoY = splitPoint
	}
	__antithesis_instrumentation__.Notify(63172)

	clipped1, err := clipGeomTByBoundingBoxForSubdivide(gt, &subBox1)
	if err != nil {
		__antithesis_instrumentation__.Notify(63211)
		return nil, errors.Wrap(err, "error clipping geom.T")
	} else {
		__antithesis_instrumentation__.Notify(63212)
	}
	__antithesis_instrumentation__.Notify(63173)
	clipped2, err := clipGeomTByBoundingBoxForSubdivide(gt, &subBox2)
	if err != nil {
		__antithesis_instrumentation__.Notify(63213)
		return nil, errors.Wrap(err, "error clipping geom.T")
	} else {
		__antithesis_instrumentation__.Notify(63214)
	}
	__antithesis_instrumentation__.Notify(63174)
	depth++
	subdivisions1, err := subdivideRecursive(clipped1, maxVertices, depth, dim, maxDepth)
	if err != nil {
		__antithesis_instrumentation__.Notify(63215)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(63216)
	}
	__antithesis_instrumentation__.Notify(63175)
	subdivisions2, err := subdivideRecursive(clipped2, maxVertices, depth, dim, maxDepth)
	if err != nil {
		__antithesis_instrumentation__.Notify(63217)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(63218)
	}
	__antithesis_instrumentation__.Notify(63176)
	results = append(results, subdivisions1...)
	results = append(results, subdivisions2...)
	return results, nil
}

func forEachGeomTInCollectionForSubdivide(gt geom.T, fn func(geom.T) error) error {
	__antithesis_instrumentation__.Notify(63219)
	switch gt := gt.(type) {
	case *geom.GeometryCollection:
		__antithesis_instrumentation__.Notify(63221)
		for i := 0; i < gt.NumGeoms(); i++ {
			__antithesis_instrumentation__.Notify(63225)
			err := fn(gt.Geom(i))
			if err != nil {
				__antithesis_instrumentation__.Notify(63226)
				return err
			} else {
				__antithesis_instrumentation__.Notify(63227)
			}
		}
		__antithesis_instrumentation__.Notify(63222)
		return nil
	case *geom.MultiPolygon:
		__antithesis_instrumentation__.Notify(63223)
		for i := 0; i < gt.NumPolygons(); i++ {
			__antithesis_instrumentation__.Notify(63228)
			err := fn(gt.Polygon(i))
			if err != nil {
				__antithesis_instrumentation__.Notify(63229)
				return err
			} else {
				__antithesis_instrumentation__.Notify(63230)
			}
		}
	case *geom.MultiLineString:
		__antithesis_instrumentation__.Notify(63224)
		for i := 0; i < gt.NumLineStrings(); i++ {
			__antithesis_instrumentation__.Notify(63231)
			err := fn(gt.LineString(i))
			if err != nil {
				__antithesis_instrumentation__.Notify(63232)
				return err
			} else {
				__antithesis_instrumentation__.Notify(63233)
			}
		}
	}
	__antithesis_instrumentation__.Notify(63220)
	return nil
}

func calculateSplitPointCoordForSubdivide(
	gt geom.T, nVertices int, splitHorizontally bool, clip geo.CartesianBoundingBox,
) (float64, error) {
	__antithesis_instrumentation__.Notify(63234)
	var pivot float64
	switch gt := gt.(type) {
	case *geom.Polygon:
		__antithesis_instrumentation__.Notify(63236)
		var err error
		pivot, err = findMostCentralPointValueForPolygon(gt, nVertices, splitHorizontally, &clip)
		if err != nil {
			__antithesis_instrumentation__.Notify(63239)
			return 0, errors.Wrap(err, "error finding most central point for polygon")
		} else {
			__antithesis_instrumentation__.Notify(63240)
		}
		__antithesis_instrumentation__.Notify(63237)
		if splitHorizontally {
			__antithesis_instrumentation__.Notify(63241)

			if clip.LoX == pivot || func() bool {
				__antithesis_instrumentation__.Notify(63242)
				return clip.HiX == pivot == true
			}() == true {
				__antithesis_instrumentation__.Notify(63243)
				pivot = (clip.LoX + clip.HiX) / 2
			} else {
				__antithesis_instrumentation__.Notify(63244)
			}
		} else {
			__antithesis_instrumentation__.Notify(63245)

			if clip.LoY == pivot || func() bool {
				__antithesis_instrumentation__.Notify(63246)
				return clip.HiY == pivot == true
			}() == true {
				__antithesis_instrumentation__.Notify(63247)
				pivot = (clip.LoY + clip.HiY) / 2
			} else {
				__antithesis_instrumentation__.Notify(63248)
			}
		}
	default:
		__antithesis_instrumentation__.Notify(63238)
		if splitHorizontally {
			__antithesis_instrumentation__.Notify(63249)
			pivot = (clip.LoX + clip.HiX) / 2
		} else {
			__antithesis_instrumentation__.Notify(63250)
			pivot = (clip.LoY + clip.HiY) / 2
		}
	}
	__antithesis_instrumentation__.Notify(63235)
	return pivot, nil
}

func findMostCentralPointValueForPolygon(
	poly *geom.Polygon, nVertices int, splitHorizontally bool, clip *geo.CartesianBoundingBox,
) (float64, error) {
	__antithesis_instrumentation__.Notify(63251)
	pivot := math.MaxFloat64
	ringToTrimIndex := 0
	ringArea := float64(0)
	pivotEps := math.MaxFloat64
	var ptEps float64
	nVerticesOuterRing := CountVertices(poly.LinearRing(0))
	if nVertices >= 2*nVerticesOuterRing {
		__antithesis_instrumentation__.Notify(63254)

		for i := 1; i < poly.NumLinearRings(); i++ {
			__antithesis_instrumentation__.Notify(63255)
			currentRingArea := poly.LinearRing(i).Area()
			if currentRingArea >= ringArea {
				__antithesis_instrumentation__.Notify(63256)
				ringArea = currentRingArea
				ringToTrimIndex = i
			} else {
				__antithesis_instrumentation__.Notify(63257)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(63258)
	}
	__antithesis_instrumentation__.Notify(63252)
	pa := poly.LinearRing(ringToTrimIndex)

	for j := 0; j < pa.NumCoords(); j++ {
		__antithesis_instrumentation__.Notify(63259)
		var pt float64
		if splitHorizontally {
			__antithesis_instrumentation__.Notify(63261)
			pt = pa.Coord(j).X()
			xHalf := (clip.LoX + clip.HiX) / 2
			ptEps = math.Abs(pt - xHalf)
		} else {
			__antithesis_instrumentation__.Notify(63262)
			pt = pa.Coord(j).Y()
			yHalf := (clip.LoY + clip.HiY) / 2
			ptEps = math.Abs(pt - yHalf)
		}
		__antithesis_instrumentation__.Notify(63260)
		if ptEps < pivotEps {
			__antithesis_instrumentation__.Notify(63263)
			pivot = pt
			pivotEps = ptEps
		} else {
			__antithesis_instrumentation__.Notify(63264)
		}
	}
	__antithesis_instrumentation__.Notify(63253)
	return pivot, nil
}

func clipGeomTByBoundingBoxForSubdivide(gt geom.T, clip *geo.CartesianBoundingBox) (geom.T, error) {
	__antithesis_instrumentation__.Notify(63265)
	g, err := geo.MakeGeometryFromGeomT(gt)
	if err != nil {
		__antithesis_instrumentation__.Notify(63271)
		return nil, errors.Wrap(err, "error transforming geom.T to geometry")
	} else {
		__antithesis_instrumentation__.Notify(63272)
	}
	__antithesis_instrumentation__.Notify(63266)
	clipgt := clip.ToGeomT(g.SRID())
	clipg, err := geo.MakeGeometryFromGeomT(clipgt)
	if err != nil {
		__antithesis_instrumentation__.Notify(63273)
		return nil, errors.Wrap(err, "error transforming geom.T to geometry")
	} else {
		__antithesis_instrumentation__.Notify(63274)
	}
	__antithesis_instrumentation__.Notify(63267)
	out, err := Intersection(g, clipg)
	if err != nil {
		__antithesis_instrumentation__.Notify(63275)
		return nil, errors.Wrap(err, "error applying intersection")
	} else {
		__antithesis_instrumentation__.Notify(63276)
	}
	__antithesis_instrumentation__.Notify(63268)

	out, err = SimplifyGEOS(out, 0)
	if err != nil {
		__antithesis_instrumentation__.Notify(63277)
		return nil, errors.Wrap(err, "simplifying error")
	} else {
		__antithesis_instrumentation__.Notify(63278)
	}
	__antithesis_instrumentation__.Notify(63269)
	gt, err = out.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(63279)
		return nil, errors.Wrap(err, "error transforming geometry to geom.T")
	} else {
		__antithesis_instrumentation__.Notify(63280)
	}
	__antithesis_instrumentation__.Notify(63270)
	return gt, nil
}
