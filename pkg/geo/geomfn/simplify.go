package geomfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/errors"
	"github.com/twpayne/go-geom"
)

func Simplify(
	g geo.Geometry, tolerance float64, preserveCollapsed bool,
) (geo.Geometry, bool, error) {
	__antithesis_instrumentation__.Notify(62961)
	if g.ShapeType2D() == geopb.ShapeType_Point || func() bool {
		__antithesis_instrumentation__.Notify(62967)
		return g.ShapeType2D() == geopb.ShapeType_MultiPoint == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(62968)
		return g.Empty() == true
	}() == true {
		__antithesis_instrumentation__.Notify(62969)
		return g, false, nil
	} else {
		__antithesis_instrumentation__.Notify(62970)
	}
	__antithesis_instrumentation__.Notify(62962)

	if tolerance < 0 {
		__antithesis_instrumentation__.Notify(62971)
		tolerance = 0
	} else {
		__antithesis_instrumentation__.Notify(62972)
	}
	__antithesis_instrumentation__.Notify(62963)
	t, err := g.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(62973)
		return geo.Geometry{}, false, err
	} else {
		__antithesis_instrumentation__.Notify(62974)
	}
	__antithesis_instrumentation__.Notify(62964)
	simplifiedT, err := simplify(t, tolerance, preserveCollapsed)
	if err != nil {
		__antithesis_instrumentation__.Notify(62975)
		return geo.Geometry{}, false, err
	} else {
		__antithesis_instrumentation__.Notify(62976)
	}
	__antithesis_instrumentation__.Notify(62965)
	if simplifiedT == nil {
		__antithesis_instrumentation__.Notify(62977)
		return geo.Geometry{}, true, nil
	} else {
		__antithesis_instrumentation__.Notify(62978)
	}
	__antithesis_instrumentation__.Notify(62966)

	ret, err := geo.MakeGeometryFromGeomT(simplifiedT)
	return ret, false, err
}

func simplify(t geom.T, tolerance float64, preserveCollapsed bool) (geom.T, error) {
	__antithesis_instrumentation__.Notify(62979)
	if t.Empty() {
		__antithesis_instrumentation__.Notify(62981)
		return t, nil
	} else {
		__antithesis_instrumentation__.Notify(62982)
	}
	__antithesis_instrumentation__.Notify(62980)
	switch t := t.(type) {
	case *geom.Point, *geom.MultiPoint:
		__antithesis_instrumentation__.Notify(62983)

		return t, nil
	case *geom.LineString:
		__antithesis_instrumentation__.Notify(62984)
		return simplifySimplifiableShape(t, tolerance, preserveCollapsed)
	case *geom.MultiLineString:
		__antithesis_instrumentation__.Notify(62985)
		ret := geom.NewMultiLineString(t.Layout()).SetSRID(t.SRID())
		for i := 0; i < t.NumLineStrings(); i++ {
			__antithesis_instrumentation__.Notify(62994)
			appendT, err := simplify(t.LineString(i), tolerance, preserveCollapsed)
			if err != nil {
				__antithesis_instrumentation__.Notify(62996)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(62997)
			}
			__antithesis_instrumentation__.Notify(62995)
			if appendT != nil {
				__antithesis_instrumentation__.Notify(62998)
				if err := ret.Push(appendT.(*geom.LineString)); err != nil {
					__antithesis_instrumentation__.Notify(62999)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(63000)
				}
			} else {
				__antithesis_instrumentation__.Notify(63001)
			}
		}
		__antithesis_instrumentation__.Notify(62986)
		return ret, nil
	case *geom.Polygon:
		__antithesis_instrumentation__.Notify(62987)
		p := geom.NewPolygon(t.Layout()).SetSRID(t.SRID())
		for i := 0; i < t.NumLinearRings(); i++ {
			__antithesis_instrumentation__.Notify(63002)
			lrT, err := simplifySimplifiableShape(
				t.LinearRing(i),
				tolerance,

				preserveCollapsed && func() bool {
					__antithesis_instrumentation__.Notify(63005)
					return i == 0 == true
				}() == true,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(63006)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(63007)
			}
			__antithesis_instrumentation__.Notify(63003)
			lr := lrT.(*geom.LinearRing)
			if lr.NumCoords() < 4 {
				__antithesis_instrumentation__.Notify(63008)

				if i == 0 {
					__antithesis_instrumentation__.Notify(63010)
					return nil, nil
				} else {
					__antithesis_instrumentation__.Notify(63011)
				}
				__antithesis_instrumentation__.Notify(63009)

				continue
			} else {
				__antithesis_instrumentation__.Notify(63012)
			}
			__antithesis_instrumentation__.Notify(63004)

			if err := p.Push(lr); err != nil {
				__antithesis_instrumentation__.Notify(63013)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(63014)
			}
		}
		__antithesis_instrumentation__.Notify(62988)
		return p, nil
	case *geom.MultiPolygon:
		__antithesis_instrumentation__.Notify(62989)
		ret := geom.NewMultiPolygon(t.Layout()).SetSRID(t.SRID())
		for i := 0; i < t.NumPolygons(); i++ {
			__antithesis_instrumentation__.Notify(63015)
			appendT, err := simplify(t.Polygon(i), tolerance, preserveCollapsed)
			if err != nil {
				__antithesis_instrumentation__.Notify(63017)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(63018)
			}
			__antithesis_instrumentation__.Notify(63016)
			if appendT != nil {
				__antithesis_instrumentation__.Notify(63019)
				if err := ret.Push(appendT.(*geom.Polygon)); err != nil {
					__antithesis_instrumentation__.Notify(63020)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(63021)
				}
			} else {
				__antithesis_instrumentation__.Notify(63022)
			}
		}
		__antithesis_instrumentation__.Notify(62990)
		return ret, nil
	case *geom.GeometryCollection:
		__antithesis_instrumentation__.Notify(62991)
		ret := geom.NewGeometryCollection().SetSRID(t.SRID())
		for _, gcT := range t.Geoms() {
			__antithesis_instrumentation__.Notify(63023)
			appendT, err := simplify(gcT, tolerance, preserveCollapsed)
			if err != nil {
				__antithesis_instrumentation__.Notify(63025)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(63026)
			}
			__antithesis_instrumentation__.Notify(63024)
			if appendT != nil {
				__antithesis_instrumentation__.Notify(63027)
				if err := ret.Push(appendT); err != nil {
					__antithesis_instrumentation__.Notify(63028)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(63029)
				}
			} else {
				__antithesis_instrumentation__.Notify(63030)
			}
		}
		__antithesis_instrumentation__.Notify(62992)
		return ret, nil
	default:
		__antithesis_instrumentation__.Notify(62993)
		return nil, pgerror.Newf(pgcode.InvalidParameterValue, "unknown shape type: %T", t)
	}
}

type simplifiableShape interface {
	geom.T

	Coord(i int) geom.Coord
	NumCoords() int
}

func simplifySimplifiableShape(
	t simplifiableShape, tolerance float64, preserveCollapsed bool,
) (geom.T, error) {
	__antithesis_instrumentation__.Notify(63031)
	minPoints := 0
	if preserveCollapsed {
		__antithesis_instrumentation__.Notify(63035)
		switch t.(type) {
		case *geom.LinearRing:
			__antithesis_instrumentation__.Notify(63036)
			minPoints = 4
		case *geom.LineString:
			__antithesis_instrumentation__.Notify(63037)
			minPoints = 2
		}
	} else {
		__antithesis_instrumentation__.Notify(63038)
	}
	__antithesis_instrumentation__.Notify(63032)

	ret, err := simplifySimplifiableShapeInternal(t, tolerance, minPoints)
	if err != nil {
		__antithesis_instrumentation__.Notify(63039)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(63040)
	}
	__antithesis_instrumentation__.Notify(63033)
	switch ret := ret.(type) {
	case *geom.LineString:
		__antithesis_instrumentation__.Notify(63041)

		if !preserveCollapsed && func() bool {
			__antithesis_instrumentation__.Notify(63042)
			return ret.NumCoords() == 2 == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(63043)
			return coordEqual(ret.Coord(0), ret.Coord(1)) == true
		}() == true {
			__antithesis_instrumentation__.Notify(63044)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(63045)
		}
	}
	__antithesis_instrumentation__.Notify(63034)
	return ret, nil
}

func simplifySimplifiableShapeInternal(
	t simplifiableShape, tolerance float64, minPoints int,
) (geom.T, error) {
	__antithesis_instrumentation__.Notify(63046)
	numCoords := t.NumCoords()

	if numCoords <= 2 || func() bool {
		__antithesis_instrumentation__.Notify(63052)
		return numCoords <= minPoints == true
	}() == true {
		__antithesis_instrumentation__.Notify(63053)
		return t, nil
	} else {
		__antithesis_instrumentation__.Notify(63054)
	}
	__antithesis_instrumentation__.Notify(63047)

	keepPoints := make([]bool, numCoords)
	keepPoints[0] = true
	keepPoints[numCoords-1] = true
	numKeepPoints := 2

	if tolerance == 0 {
		__antithesis_instrumentation__.Notify(63055)

		prevCoord := t.Coord(0)
		for i := 1; i < t.NumCoords()-1; i++ {
			__antithesis_instrumentation__.Notify(63056)
			currCoord := t.Coord(i)
			nextCoord := t.Coord(i + 1)

			ab := coordSub(currCoord, prevCoord)
			ac := coordSub(nextCoord, prevCoord)

			if coordCross(ab, ac) == 0 {
				__antithesis_instrumentation__.Notify(63058)

				r := coordDot(ab, ac) / coordNorm2(ac)
				if r >= 0 && func() bool {
					__antithesis_instrumentation__.Notify(63059)
					return r <= 1 == true
				}() == true {
					__antithesis_instrumentation__.Notify(63060)
					continue
				} else {
					__antithesis_instrumentation__.Notify(63061)
				}
			} else {
				__antithesis_instrumentation__.Notify(63062)
			}
			__antithesis_instrumentation__.Notify(63057)

			prevCoord = currCoord
			keepPoints[i] = true
			numKeepPoints++
		}
	} else {
		__antithesis_instrumentation__.Notify(63063)

		var recurse func(startIdx, endIdx int)
		recurse = func(startIdx, endIdx int) {
			__antithesis_instrumentation__.Notify(63065)

			splitIdx := findSplitIndexForSimplifiableShape(t, startIdx, endIdx, tolerance, minPoints > numKeepPoints)

			if splitIdx == startIdx {
				__antithesis_instrumentation__.Notify(63067)
				return
			} else {
				__antithesis_instrumentation__.Notify(63068)
			}
			__antithesis_instrumentation__.Notify(63066)
			keepPoints[splitIdx] = true
			numKeepPoints++
			recurse(startIdx, splitIdx)
			recurse(splitIdx, endIdx)
		}
		__antithesis_instrumentation__.Notify(63064)
		recurse(0, numCoords-1)
	}
	__antithesis_instrumentation__.Notify(63048)

	if numKeepPoints == numCoords {
		__antithesis_instrumentation__.Notify(63069)
		return t, nil
	} else {
		__antithesis_instrumentation__.Notify(63070)
	}
	__antithesis_instrumentation__.Notify(63049)

	newFlatCoords := make([]float64, 0, numKeepPoints*t.Layout().Stride())
	for i, keep := range keepPoints {
		__antithesis_instrumentation__.Notify(63071)
		shouldAdd := keep

		if !shouldAdd && func() bool {
			__antithesis_instrumentation__.Notify(63073)
			return numKeepPoints < minPoints == true
		}() == true {
			__antithesis_instrumentation__.Notify(63074)
			shouldAdd = true
			numKeepPoints++
		} else {
			__antithesis_instrumentation__.Notify(63075)
		}
		__antithesis_instrumentation__.Notify(63072)
		if shouldAdd {
			__antithesis_instrumentation__.Notify(63076)
			newFlatCoords = append(newFlatCoords, []float64(t.Coord(i))...)
		} else {
			__antithesis_instrumentation__.Notify(63077)
		}
	}
	__antithesis_instrumentation__.Notify(63050)

	switch t.(type) {
	case *geom.LineString:
		__antithesis_instrumentation__.Notify(63078)
		return geom.NewLineStringFlat(t.Layout(), newFlatCoords).SetSRID(t.SRID()), nil
	case *geom.LinearRing:
		__antithesis_instrumentation__.Notify(63079)
		return geom.NewLinearRingFlat(t.Layout(), newFlatCoords), nil
	}
	__antithesis_instrumentation__.Notify(63051)
	return nil, errors.AssertionFailedf("unknown shape: %T", t)
}

func findSplitIndexForSimplifiableShape(
	t simplifiableShape, startIdx int, endIdx int, tolerance float64, forceSplit bool,
) int {
	__antithesis_instrumentation__.Notify(63080)

	if endIdx-startIdx < 2 {
		__antithesis_instrumentation__.Notify(63083)
		return startIdx
	} else {
		__antithesis_instrumentation__.Notify(63084)
	}
	__antithesis_instrumentation__.Notify(63081)

	edgeV0 := t.Coord(startIdx)
	edgeV1 := t.Coord(endIdx)

	edgeIsSamePoint := coordEqual(edgeV0, edgeV1)
	splitIdx := startIdx
	var maxDistance float64

	for i := startIdx + 1; i < endIdx; i++ {
		__antithesis_instrumentation__.Notify(63085)
		curr := t.Coord(i)
		var dist float64

		if !coordEqual(curr, edgeV0) && func() bool {
			__antithesis_instrumentation__.Notify(63087)
			return !coordEqual(curr, edgeV1) == true
		}() == true {
			__antithesis_instrumentation__.Notify(63088)

			closestPointToEdge := edgeV0
			if !edgeIsSamePoint {
				__antithesis_instrumentation__.Notify(63090)
				ac := coordSub(curr, edgeV0)
				ab := coordSub(edgeV1, edgeV0)

				r := coordDot(ac, ab) / coordNorm2(ab)
				if r >= 1 {
					__antithesis_instrumentation__.Notify(63091)

					closestPointToEdge = edgeV1
				} else {
					__antithesis_instrumentation__.Notify(63092)
					if r > 0 {
						__antithesis_instrumentation__.Notify(63093)

						closestPointToEdge = coordAdd(edgeV0, coordMul(ab, r))
					} else {
						__antithesis_instrumentation__.Notify(63094)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(63095)
			}
			__antithesis_instrumentation__.Notify(63089)

			dist = coordNorm(coordSub(closestPointToEdge, curr))
		} else {
			__antithesis_instrumentation__.Notify(63096)
		}
		__antithesis_instrumentation__.Notify(63086)
		if (dist > tolerance && func() bool {
			__antithesis_instrumentation__.Notify(63097)
			return dist > maxDistance == true
		}() == true) || func() bool {
			__antithesis_instrumentation__.Notify(63098)
			return forceSplit == true
		}() == true {
			__antithesis_instrumentation__.Notify(63099)
			forceSplit = false
			splitIdx = i
			maxDistance = dist
		} else {
			__antithesis_instrumentation__.Notify(63100)
		}
	}
	__antithesis_instrumentation__.Notify(63082)
	return splitIdx
}
