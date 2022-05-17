package geogfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/golang/geo/s2"
)

func Intersects(a geo.Geography, b geo.Geography) (bool, error) {
	__antithesis_instrumentation__.Notify(59716)
	if !a.BoundingRect().Intersects(b.BoundingRect()) {
		__antithesis_instrumentation__.Notify(59722)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(59723)
	}
	__antithesis_instrumentation__.Notify(59717)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(59724)
		return false, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(59725)
	}
	__antithesis_instrumentation__.Notify(59718)

	aRegions, err := a.AsS2(geo.EmptyBehaviorOmit)
	if err != nil {
		__antithesis_instrumentation__.Notify(59726)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(59727)
	}
	__antithesis_instrumentation__.Notify(59719)
	bRegions, err := b.AsS2(geo.EmptyBehaviorOmit)
	if err != nil {
		__antithesis_instrumentation__.Notify(59728)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(59729)
	}
	__antithesis_instrumentation__.Notify(59720)

	for _, aRegion := range aRegions {
		__antithesis_instrumentation__.Notify(59730)
		for _, bRegion := range bRegions {
			__antithesis_instrumentation__.Notify(59731)
			intersects, err := singleRegionIntersects(aRegion, bRegion)
			if err != nil {
				__antithesis_instrumentation__.Notify(59733)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(59734)
			}
			__antithesis_instrumentation__.Notify(59732)
			if intersects {
				__antithesis_instrumentation__.Notify(59735)
				return true, nil
			} else {
				__antithesis_instrumentation__.Notify(59736)
			}
		}
	}
	__antithesis_instrumentation__.Notify(59721)
	return false, nil
}

func singleRegionIntersects(aRegion s2.Region, bRegion s2.Region) (bool, error) {
	__antithesis_instrumentation__.Notify(59737)
	switch aRegion := aRegion.(type) {
	case s2.Point:
		__antithesis_instrumentation__.Notify(59739)
		switch bRegion := bRegion.(type) {
		case s2.Point:
			__antithesis_instrumentation__.Notify(59742)
			return aRegion.IntersectsCell(s2.CellFromPoint(bRegion)), nil
		case *s2.Polyline:
			__antithesis_instrumentation__.Notify(59743)
			return bRegion.IntersectsCell(s2.CellFromPoint(aRegion)), nil
		case *s2.Polygon:
			__antithesis_instrumentation__.Notify(59744)
			return bRegion.IntersectsCell(s2.CellFromPoint(aRegion)), nil
		default:
			__antithesis_instrumentation__.Notify(59745)
			return false, pgerror.Newf(pgcode.InvalidParameterValue, "unknown s2 type of b: %#v", bRegion)
		}
	case *s2.Polyline:
		__antithesis_instrumentation__.Notify(59740)
		switch bRegion := bRegion.(type) {
		case s2.Point:
			__antithesis_instrumentation__.Notify(59746)
			return aRegion.IntersectsCell(s2.CellFromPoint(bRegion)), nil
		case *s2.Polyline:
			__antithesis_instrumentation__.Notify(59747)
			return polylineIntersectsPolyline(aRegion, bRegion), nil
		case *s2.Polygon:
			__antithesis_instrumentation__.Notify(59748)
			return polygonIntersectsPolyline(bRegion, aRegion), nil
		default:
			__antithesis_instrumentation__.Notify(59749)
			return false, pgerror.Newf(pgcode.InvalidParameterValue, "unknown s2 type of b: %#v", bRegion)
		}
	case *s2.Polygon:
		__antithesis_instrumentation__.Notify(59741)
		switch bRegion := bRegion.(type) {
		case s2.Point:
			__antithesis_instrumentation__.Notify(59750)
			return aRegion.IntersectsCell(s2.CellFromPoint(bRegion)), nil
		case *s2.Polyline:
			__antithesis_instrumentation__.Notify(59751)
			return polygonIntersectsPolyline(aRegion, bRegion), nil
		case *s2.Polygon:
			__antithesis_instrumentation__.Notify(59752)
			return aRegion.Intersects(bRegion), nil
		default:
			__antithesis_instrumentation__.Notify(59753)
			return false, pgerror.Newf(pgcode.InvalidParameterValue, "unknown s2 type of b: %#v", bRegion)
		}
	}
	__antithesis_instrumentation__.Notify(59738)
	return false, pgerror.Newf(pgcode.InvalidParameterValue, "unknown s2 type of a: %#v", aRegion)
}

func polylineIntersectsPolyline(a *s2.Polyline, b *s2.Polyline) bool {
	__antithesis_instrumentation__.Notify(59754)
	for aEdgeIdx, aNumEdges := 0, a.NumEdges(); aEdgeIdx < aNumEdges; aEdgeIdx++ {
		__antithesis_instrumentation__.Notify(59756)
		edge := a.Edge(aEdgeIdx)
		crosser := s2.NewChainEdgeCrosser(edge.V0, edge.V1, (*b)[0])
		for _, nextVertex := range (*b)[1:] {
			__antithesis_instrumentation__.Notify(59757)
			crossing := crosser.ChainCrossingSign(nextVertex)
			if crossing != s2.DoNotCross {
				__antithesis_instrumentation__.Notify(59758)
				return true
			} else {
				__antithesis_instrumentation__.Notify(59759)
			}
		}
	}
	__antithesis_instrumentation__.Notify(59755)
	return false
}

func polygonIntersectsPolyline(a *s2.Polygon, b *s2.Polyline) bool {
	__antithesis_instrumentation__.Notify(59760)

	for _, vertex := range *b {
		__antithesis_instrumentation__.Notify(59763)
		if a.IntersectsCell(s2.CellFromPoint(vertex)) {
			__antithesis_instrumentation__.Notify(59764)
			return true
		} else {
			__antithesis_instrumentation__.Notify(59765)
		}
	}
	__antithesis_instrumentation__.Notify(59761)

	for _, loop := range a.Loops() {
		__antithesis_instrumentation__.Notify(59766)
		for loopEdgeIdx, loopNumEdges := 0, loop.NumEdges(); loopEdgeIdx < loopNumEdges; loopEdgeIdx++ {
			__antithesis_instrumentation__.Notify(59767)
			loopEdge := loop.Edge(loopEdgeIdx)
			crosser := s2.NewChainEdgeCrosser(loopEdge.V0, loopEdge.V1, (*b)[0])
			for _, nextVertex := range (*b)[1:] {
				__antithesis_instrumentation__.Notify(59768)
				crossing := crosser.ChainCrossingSign(nextVertex)
				if crossing != s2.DoNotCross {
					__antithesis_instrumentation__.Notify(59769)
					return true
				} else {
					__antithesis_instrumentation__.Notify(59770)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(59762)
	return false
}
