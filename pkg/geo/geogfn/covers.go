package geogfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/golang/geo/s2"
)

func Covers(a geo.Geography, b geo.Geography) (bool, error) {
	__antithesis_instrumentation__.Notify(59498)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(59500)
		return false, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(59501)
	}
	__antithesis_instrumentation__.Notify(59499)
	return covers(a, b)
}

func covers(a geo.Geography, b geo.Geography) (bool, error) {
	__antithesis_instrumentation__.Notify(59502)

	if !a.BoundingRect().Contains(b.BoundingRect()) {
		__antithesis_instrumentation__.Notify(59508)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(59509)
	}
	__antithesis_instrumentation__.Notify(59503)

	aRegions, err := a.AsS2(geo.EmptyBehaviorOmit)
	if err != nil {
		__antithesis_instrumentation__.Notify(59510)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(59511)
	}
	__antithesis_instrumentation__.Notify(59504)

	bRegions, err := b.AsS2(geo.EmptyBehaviorError)
	if err != nil {
		__antithesis_instrumentation__.Notify(59512)
		if geo.IsEmptyGeometryError(err) {
			__antithesis_instrumentation__.Notify(59514)
			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(59515)
		}
		__antithesis_instrumentation__.Notify(59513)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(59516)
	}
	__antithesis_instrumentation__.Notify(59505)

	bRegionsRemaining := make(map[int]struct{}, len(bRegions))
	for i := range bRegions {
		__antithesis_instrumentation__.Notify(59517)
		bRegionsRemaining[i] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(59506)
	for _, aRegion := range aRegions {
		__antithesis_instrumentation__.Notify(59518)
		for bRegionIdx := range bRegionsRemaining {
			__antithesis_instrumentation__.Notify(59520)
			regionCovers, err := regionCovers(aRegion, bRegions[bRegionIdx])
			if err != nil {
				__antithesis_instrumentation__.Notify(59522)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(59523)
			}
			__antithesis_instrumentation__.Notify(59521)
			if regionCovers {
				__antithesis_instrumentation__.Notify(59524)
				delete(bRegionsRemaining, bRegionIdx)
			} else {
				__antithesis_instrumentation__.Notify(59525)
			}
		}
		__antithesis_instrumentation__.Notify(59519)
		if len(bRegionsRemaining) == 0 {
			__antithesis_instrumentation__.Notify(59526)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(59527)
		}
	}
	__antithesis_instrumentation__.Notify(59507)
	return false, nil
}

func CoveredBy(a geo.Geography, b geo.Geography) (bool, error) {
	__antithesis_instrumentation__.Notify(59528)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(59530)
		return false, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(59531)
	}
	__antithesis_instrumentation__.Notify(59529)
	return covers(b, a)
}

func regionCovers(aRegion s2.Region, bRegion s2.Region) (bool, error) {
	__antithesis_instrumentation__.Notify(59532)
	switch aRegion := aRegion.(type) {
	case s2.Point:
		__antithesis_instrumentation__.Notify(59534)
		switch bRegion := bRegion.(type) {
		case s2.Point:
			__antithesis_instrumentation__.Notify(59537)
			return aRegion.ContainsPoint(bRegion), nil
		case *s2.Polyline:
			__antithesis_instrumentation__.Notify(59538)
			return false, nil
		case *s2.Polygon:
			__antithesis_instrumentation__.Notify(59539)
			return false, nil
		default:
			__antithesis_instrumentation__.Notify(59540)
			return false, pgerror.Newf(pgcode.InvalidParameterValue, "unknown s2 type of b: %#v", bRegion)
		}
	case *s2.Polyline:
		__antithesis_instrumentation__.Notify(59535)
		switch bRegion := bRegion.(type) {
		case s2.Point:
			__antithesis_instrumentation__.Notify(59541)
			return polylineCoversPoint(aRegion, bRegion), nil
		case *s2.Polyline:
			__antithesis_instrumentation__.Notify(59542)
			return polylineCoversPolyline(aRegion, bRegion), nil
		case *s2.Polygon:
			__antithesis_instrumentation__.Notify(59543)
			return false, nil
		default:
			__antithesis_instrumentation__.Notify(59544)
			return false, pgerror.Newf(pgcode.InvalidParameterValue, "unknown s2 type of b: %#v", bRegion)
		}
	case *s2.Polygon:
		__antithesis_instrumentation__.Notify(59536)
		switch bRegion := bRegion.(type) {
		case s2.Point:
			__antithesis_instrumentation__.Notify(59545)
			return polygonCoversPoint(aRegion, bRegion), nil
		case *s2.Polyline:
			__antithesis_instrumentation__.Notify(59546)
			return polygonCoversPolyline(aRegion, bRegion), nil
		case *s2.Polygon:
			__antithesis_instrumentation__.Notify(59547)
			return polygonCoversPolygon(aRegion, bRegion), nil
		default:
			__antithesis_instrumentation__.Notify(59548)
			return false, pgerror.Newf(pgcode.InvalidParameterValue, "unknown s2 type of b: %#v", bRegion)
		}
	}
	__antithesis_instrumentation__.Notify(59533)
	return false, pgerror.Newf(pgcode.InvalidParameterValue, "unknown s2 type of a: %#v", aRegion)
}

func polylineCoversPoint(a *s2.Polyline, b s2.Point) bool {
	__antithesis_instrumentation__.Notify(59549)
	return a.IntersectsCell(s2.CellFromPoint(b))
}

func polylineCoversPointWithIdx(a *s2.Polyline, b s2.Point) (bool, int) {
	__antithesis_instrumentation__.Notify(59550)
	for edgeIdx, aNumEdges := 0, a.NumEdges(); edgeIdx < aNumEdges; edgeIdx++ {
		__antithesis_instrumentation__.Notify(59552)
		if edgeCoversPoint(a.Edge(edgeIdx), b) {
			__antithesis_instrumentation__.Notify(59553)
			return true, edgeIdx
		} else {
			__antithesis_instrumentation__.Notify(59554)
		}
	}
	__antithesis_instrumentation__.Notify(59551)
	return false, -1
}

func polygonCoversPoint(a *s2.Polygon, b s2.Point) bool {
	__antithesis_instrumentation__.Notify(59555)
	return a.IntersectsCell(s2.CellFromPoint(b))
}

func edgeCoversPoint(e s2.Edge, p s2.Point) bool {
	__antithesis_instrumentation__.Notify(59556)
	return (&s2.Polyline{e.V0, e.V1}).IntersectsCell(s2.CellFromPoint(p))
}

func polylineCoversPolyline(a *s2.Polyline, b *s2.Polyline) bool {
	__antithesis_instrumentation__.Notify(59557)
	if polylineCoversPolylineOrdered(a, b) {
		__antithesis_instrumentation__.Notify(59560)
		return true
	} else {
		__antithesis_instrumentation__.Notify(59561)
	}
	__antithesis_instrumentation__.Notify(59558)

	reversedB := make([]s2.Point, len(*b))
	for i, point := range *b {
		__antithesis_instrumentation__.Notify(59562)
		reversedB[len(reversedB)-1-i] = point
	}
	__antithesis_instrumentation__.Notify(59559)
	newBAsPolyline := s2.Polyline(reversedB)
	return polylineCoversPolylineOrdered(a, &newBAsPolyline)
}

func polylineCoversPolylineOrdered(a *s2.Polyline, b *s2.Polyline) bool {
	__antithesis_instrumentation__.Notify(59563)
	aCoversStartOfB, aCoverBStart := polylineCoversPointWithIdx(a, (*b)[0])

	if !aCoversStartOfB {
		__antithesis_instrumentation__.Notify(59565)
		return false
	} else {
		__antithesis_instrumentation__.Notify(59566)
	}
	__antithesis_instrumentation__.Notify(59564)

	aPoints := *a
	bPoints := *b

	aIdx := aCoverBStart
	bIdx := 0

	aEdge := s2.Edge{V0: aPoints[aIdx], V1: aPoints[aIdx+1]}
	bEdge := s2.Edge{V0: bPoints[bIdx], V1: bPoints[bIdx+1]}
	for {
		__antithesis_instrumentation__.Notify(59567)
		aEdgeCoversBStart := edgeCoversPoint(aEdge, bEdge.V0)
		aEdgeCoversBEnd := edgeCoversPoint(aEdge, bEdge.V1)
		bEdgeCoversAEnd := edgeCoversPoint(bEdge, aEdge.V1)
		if aEdgeCoversBStart && func() bool {
			__antithesis_instrumentation__.Notify(59570)
			return aEdgeCoversBEnd == true
		}() == true {
			__antithesis_instrumentation__.Notify(59571)

			bIdx++

			if bIdx == len(bPoints)-1 {
				__antithesis_instrumentation__.Notify(59574)
				return true
			} else {
				__antithesis_instrumentation__.Notify(59575)
			}
			__antithesis_instrumentation__.Notify(59572)
			bEdge = s2.Edge{V0: bPoints[bIdx], V1: bPoints[bIdx+1]}

			if bEdgeCoversAEnd {
				__antithesis_instrumentation__.Notify(59576)
				aIdx++
				if aIdx == len(aPoints)-1 {
					__antithesis_instrumentation__.Notify(59578)

					return false
				} else {
					__antithesis_instrumentation__.Notify(59579)
				}
				__antithesis_instrumentation__.Notify(59577)
				aEdge = s2.Edge{V0: aPoints[aIdx], V1: aPoints[aIdx+1]}
			} else {
				__antithesis_instrumentation__.Notify(59580)
			}
			__antithesis_instrumentation__.Notify(59573)
			continue
		} else {
			__antithesis_instrumentation__.Notify(59581)
		}
		__antithesis_instrumentation__.Notify(59568)

		if aEdgeCoversBStart {
			__antithesis_instrumentation__.Notify(59582)

			if !bEdgeCoversAEnd {
				__antithesis_instrumentation__.Notify(59585)
				return false
			} else {
				__antithesis_instrumentation__.Notify(59586)
			}
			__antithesis_instrumentation__.Notify(59583)

			bEdge.V0 = aEdge.V1
			aIdx++
			if aIdx == len(aPoints)-1 {
				__antithesis_instrumentation__.Notify(59587)

				return false
			} else {
				__antithesis_instrumentation__.Notify(59588)
			}
			__antithesis_instrumentation__.Notify(59584)
			aEdge = s2.Edge{V0: aPoints[aIdx], V1: aPoints[aIdx+1]}
			continue
		} else {
			__antithesis_instrumentation__.Notify(59589)
		}
		__antithesis_instrumentation__.Notify(59569)

		return false
	}
}

func polygonCoversPolyline(a *s2.Polygon, b *s2.Polyline) bool {
	__antithesis_instrumentation__.Notify(59590)

	for _, vertex := range *b {
		__antithesis_instrumentation__.Notify(59592)
		if !polygonCoversPoint(a, vertex) {
			__antithesis_instrumentation__.Notify(59593)
			return false
		} else {
			__antithesis_instrumentation__.Notify(59594)
		}
	}
	__antithesis_instrumentation__.Notify(59591)

	return !polygonIntersectsPolylineEdge(a, b)
}

func polygonIntersectsPolylineEdge(a *s2.Polygon, b *s2.Polyline) bool {
	__antithesis_instrumentation__.Notify(59595)

	for _, loop := range a.Loops() {
		__antithesis_instrumentation__.Notify(59597)
		for loopEdgeIdx, loopNumEdges := 0, loop.NumEdges(); loopEdgeIdx < loopNumEdges; loopEdgeIdx++ {
			__antithesis_instrumentation__.Notify(59598)
			loopEdge := loop.Edge(loopEdgeIdx)
			crosser := s2.NewChainEdgeCrosser(loopEdge.V0, loopEdge.V1, (*b)[0])
			for _, nextVertex := range (*b)[1:] {
				__antithesis_instrumentation__.Notify(59599)
				if crosser.ChainCrossingSign(nextVertex) != s2.DoNotCross {
					__antithesis_instrumentation__.Notify(59600)
					return true
				} else {
					__antithesis_instrumentation__.Notify(59601)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(59596)
	return false
}

func polygonCoversPolygon(a *s2.Polygon, b *s2.Polygon) bool {
	__antithesis_instrumentation__.Notify(59602)

	return a.Contains(b)
}
