// Package geodist finds distances between two geospatial shapes.
package geodist

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/golang/geo/s2"
	"github.com/twpayne/go-geom"
)

type Point struct {
	GeomPoint geom.Coord
	GeogPoint s2.Point
}

func (p *Point) IsShape() { __antithesis_instrumentation__.Notify(59220) }

type Edge struct {
	V0, V1 Point
}

type LineString interface {
	Edge(i int) Edge
	NumEdges() int
	Vertex(i int) Point
	NumVertexes() int
	IsShape()
	IsLineString()
}

type LinearRing interface {
	Edge(i int) Edge
	NumEdges() int
	Vertex(i int) Point
	NumVertexes() int
	IsShape()
	IsLinearRing()
}

type shapeWithEdges interface {
	Edge(i int) Edge
	NumEdges() int
}

type Polygon interface {
	LinearRing(i int) LinearRing
	NumLinearRings() int
	IsShape()
	IsPolygon()
}

type Shape interface {
	IsShape()
}

var _ Shape = (*Point)(nil)
var _ Shape = (LineString)(nil)
var _ Shape = (LinearRing)(nil)
var _ Shape = (Polygon)(nil)

type DistanceUpdater interface {
	Update(a Point, b Point) bool

	OnIntersects(p Point) bool

	Distance() float64

	IsMaxDistance() bool

	FlipGeometries()
}

type EdgeCrosser interface {
	ChainCrossing(p Point) (bool, Point)
}

type DistanceCalculator interface {
	DistanceUpdater() DistanceUpdater

	NewEdgeCrosser(edge Edge, startPoint Point) EdgeCrosser

	PointIntersectsLinearRing(point Point, linearRing LinearRing) bool

	ClosestPointToEdge(edge Edge, point Point) (Point, bool)

	BoundingBoxIntersects() bool
}

func ShapeDistance(c DistanceCalculator, a Shape, b Shape) (bool, error) {
	__antithesis_instrumentation__.Notify(59221)
	switch a := a.(type) {
	case *Point:
		__antithesis_instrumentation__.Notify(59223)
		switch b := b.(type) {
		case *Point:
			__antithesis_instrumentation__.Notify(59226)
			return c.DistanceUpdater().Update(*a, *b), nil
		case LineString:
			__antithesis_instrumentation__.Notify(59227)
			return onPointToLineString(c, *a, b), nil
		case Polygon:
			__antithesis_instrumentation__.Notify(59228)
			return onPointToPolygon(c, *a, b), nil
		default:
			__antithesis_instrumentation__.Notify(59229)
			return false, pgerror.Newf(pgcode.InvalidParameterValue, "unknown shape: %T", b)
		}
	case LineString:
		__antithesis_instrumentation__.Notify(59224)
		switch b := b.(type) {
		case *Point:
			__antithesis_instrumentation__.Notify(59230)
			c.DistanceUpdater().FlipGeometries()

			defer c.DistanceUpdater().FlipGeometries()
			return onPointToLineString(c, *b, a), nil
		case LineString:
			__antithesis_instrumentation__.Notify(59231)
			return onShapeEdgesToShapeEdges(c, a, b), nil
		case Polygon:
			__antithesis_instrumentation__.Notify(59232)
			return onLineStringToPolygon(c, a, b), nil
		default:
			__antithesis_instrumentation__.Notify(59233)
			return false, pgerror.Newf(pgcode.InvalidParameterValue, "unknown shape: %T", b)
		}
	case Polygon:
		__antithesis_instrumentation__.Notify(59225)
		switch b := b.(type) {
		case *Point:
			__antithesis_instrumentation__.Notify(59234)
			c.DistanceUpdater().FlipGeometries()

			defer c.DistanceUpdater().FlipGeometries()
			return onPointToPolygon(c, *b, a), nil
		case LineString:
			__antithesis_instrumentation__.Notify(59235)
			c.DistanceUpdater().FlipGeometries()

			defer c.DistanceUpdater().FlipGeometries()
			return onLineStringToPolygon(c, b, a), nil
		case Polygon:
			__antithesis_instrumentation__.Notify(59236)
			return onPolygonToPolygon(c, a, b), nil
		default:
			__antithesis_instrumentation__.Notify(59237)
			return false, pgerror.Newf(pgcode.InvalidParameterValue, "unknown shape: %T", b)
		}
	}
	__antithesis_instrumentation__.Notify(59222)
	return false, pgerror.Newf(pgcode.InvalidParameterValue, "unknown shape: %T", a)
}

func onPointToEdgesExceptFirstEdgeStart(c DistanceCalculator, a Point, b shapeWithEdges) bool {
	__antithesis_instrumentation__.Notify(59238)
	for edgeIdx, bNumEdges := 0, b.NumEdges(); edgeIdx < bNumEdges; edgeIdx++ {
		__antithesis_instrumentation__.Notify(59240)
		edge := b.Edge(edgeIdx)

		if c.DistanceUpdater().Update(a, edge.V1) {
			__antithesis_instrumentation__.Notify(59242)
			return true
		} else {
			__antithesis_instrumentation__.Notify(59243)
		}
		__antithesis_instrumentation__.Notify(59241)

		if !c.DistanceUpdater().IsMaxDistance() {
			__antithesis_instrumentation__.Notify(59244)

			if closestPoint, ok := c.ClosestPointToEdge(edge, a); ok {
				__antithesis_instrumentation__.Notify(59245)
				if c.DistanceUpdater().Update(a, closestPoint) {
					__antithesis_instrumentation__.Notify(59246)
					return true
				} else {
					__antithesis_instrumentation__.Notify(59247)
				}
			} else {
				__antithesis_instrumentation__.Notify(59248)
			}
		} else {
			__antithesis_instrumentation__.Notify(59249)
		}
	}
	__antithesis_instrumentation__.Notify(59239)
	return false
}

func onPointToLineString(c DistanceCalculator, a Point, b LineString) bool {
	__antithesis_instrumentation__.Notify(59250)

	if c.DistanceUpdater().Update(a, b.Vertex(0)) {
		__antithesis_instrumentation__.Notify(59252)
		return true
	} else {
		__antithesis_instrumentation__.Notify(59253)
	}
	__antithesis_instrumentation__.Notify(59251)
	return onPointToEdgesExceptFirstEdgeStart(c, a, b)
}

func onPointToPolygon(c DistanceCalculator, a Point, b Polygon) bool {
	__antithesis_instrumentation__.Notify(59254)

	if c.DistanceUpdater().IsMaxDistance() || func() bool {
		__antithesis_instrumentation__.Notify(59257)
		return !c.BoundingBoxIntersects() == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(59258)
		return !c.PointIntersectsLinearRing(a, b.LinearRing(0)) == true
	}() == true {
		__antithesis_instrumentation__.Notify(59259)
		return onPointToEdgesExceptFirstEdgeStart(c, a, b.LinearRing(0))
	} else {
		__antithesis_instrumentation__.Notify(59260)
	}
	__antithesis_instrumentation__.Notify(59255)

	for ringIdx := 1; ringIdx < b.NumLinearRings(); ringIdx++ {
		__antithesis_instrumentation__.Notify(59261)
		ring := b.LinearRing(ringIdx)
		if c.PointIntersectsLinearRing(a, ring) {
			__antithesis_instrumentation__.Notify(59262)
			return onPointToEdgesExceptFirstEdgeStart(c, a, ring)
		} else {
			__antithesis_instrumentation__.Notify(59263)
		}
	}
	__antithesis_instrumentation__.Notify(59256)

	return c.DistanceUpdater().OnIntersects(a)
}

func onShapeEdgesToShapeEdges(c DistanceCalculator, a shapeWithEdges, b shapeWithEdges) bool {
	__antithesis_instrumentation__.Notify(59264)
	for aEdgeIdx, aNumEdges := 0, a.NumEdges(); aEdgeIdx < aNumEdges; aEdgeIdx++ {
		__antithesis_instrumentation__.Notify(59266)
		aEdge := a.Edge(aEdgeIdx)
		var crosser EdgeCrosser

		if !c.DistanceUpdater().IsMaxDistance() && func() bool {
			__antithesis_instrumentation__.Notify(59268)
			return c.BoundingBoxIntersects() == true
		}() == true {
			__antithesis_instrumentation__.Notify(59269)
			crosser = c.NewEdgeCrosser(aEdge, b.Edge(0).V0)
		} else {
			__antithesis_instrumentation__.Notify(59270)
		}
		__antithesis_instrumentation__.Notify(59267)
		for bEdgeIdx, bNumEdges := 0, b.NumEdges(); bEdgeIdx < bNumEdges; bEdgeIdx++ {
			__antithesis_instrumentation__.Notify(59271)
			bEdge := b.Edge(bEdgeIdx)
			if crosser != nil {
				__antithesis_instrumentation__.Notify(59274)

				intersects, intersectionPoint := crosser.ChainCrossing(bEdge.V1)
				if intersects {
					__antithesis_instrumentation__.Notify(59275)
					return c.DistanceUpdater().OnIntersects(intersectionPoint)
				} else {
					__antithesis_instrumentation__.Notify(59276)
				}
			} else {
				__antithesis_instrumentation__.Notify(59277)
			}
			__antithesis_instrumentation__.Notify(59272)

			if c.DistanceUpdater().Update(aEdge.V0, bEdge.V0) || func() bool {
				__antithesis_instrumentation__.Notify(59278)
				return c.DistanceUpdater().Update(aEdge.V0, bEdge.V1) == true
			}() == true || func() bool {
				__antithesis_instrumentation__.Notify(59279)
				return c.DistanceUpdater().Update(aEdge.V1, bEdge.V0) == true
			}() == true || func() bool {
				__antithesis_instrumentation__.Notify(59280)
				return c.DistanceUpdater().Update(aEdge.V1, bEdge.V1) == true
			}() == true {
				__antithesis_instrumentation__.Notify(59281)
				return true
			} else {
				__antithesis_instrumentation__.Notify(59282)
			}
			__antithesis_instrumentation__.Notify(59273)

			if !c.DistanceUpdater().IsMaxDistance() {
				__antithesis_instrumentation__.Notify(59283)
				if projectVertexToEdge(c, aEdge.V0, bEdge) || func() bool {
					__antithesis_instrumentation__.Notify(59286)
					return projectVertexToEdge(c, aEdge.V1, bEdge) == true
				}() == true {
					__antithesis_instrumentation__.Notify(59287)
					return true
				} else {
					__antithesis_instrumentation__.Notify(59288)
				}
				__antithesis_instrumentation__.Notify(59284)
				c.DistanceUpdater().FlipGeometries()
				if projectVertexToEdge(c, bEdge.V0, aEdge) || func() bool {
					__antithesis_instrumentation__.Notify(59289)
					return projectVertexToEdge(c, bEdge.V1, aEdge) == true
				}() == true {
					__antithesis_instrumentation__.Notify(59290)

					c.DistanceUpdater().FlipGeometries()
					return true
				} else {
					__antithesis_instrumentation__.Notify(59291)
				}
				__antithesis_instrumentation__.Notify(59285)

				c.DistanceUpdater().FlipGeometries()
			} else {
				__antithesis_instrumentation__.Notify(59292)
			}
		}
	}
	__antithesis_instrumentation__.Notify(59265)
	return false
}

func projectVertexToEdge(c DistanceCalculator, vertex Point, edge Edge) bool {
	__antithesis_instrumentation__.Notify(59293)

	if closestPoint, ok := c.ClosestPointToEdge(edge, vertex); ok {
		__antithesis_instrumentation__.Notify(59295)
		if c.DistanceUpdater().Update(vertex, closestPoint) {
			__antithesis_instrumentation__.Notify(59296)
			return true
		} else {
			__antithesis_instrumentation__.Notify(59297)
		}
	} else {
		__antithesis_instrumentation__.Notify(59298)
	}
	__antithesis_instrumentation__.Notify(59294)
	return false
}

func onLineStringToPolygon(c DistanceCalculator, a LineString, b Polygon) bool {
	__antithesis_instrumentation__.Notify(59299)

	if c.DistanceUpdater().IsMaxDistance() || func() bool {
		__antithesis_instrumentation__.Notify(59302)
		return !c.BoundingBoxIntersects() == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(59303)
		return !c.PointIntersectsLinearRing(a.Vertex(0), b.LinearRing(0)) == true
	}() == true {
		__antithesis_instrumentation__.Notify(59304)
		return onShapeEdgesToShapeEdges(c, a, b.LinearRing(0))
	} else {
		__antithesis_instrumentation__.Notify(59305)
	}
	__antithesis_instrumentation__.Notify(59300)

	for ringIdx := 1; ringIdx < b.NumLinearRings(); ringIdx++ {
		__antithesis_instrumentation__.Notify(59306)
		hole := b.LinearRing(ringIdx)
		if onShapeEdgesToShapeEdges(c, a, hole) {
			__antithesis_instrumentation__.Notify(59308)
			return true
		} else {
			__antithesis_instrumentation__.Notify(59309)
		}
		__antithesis_instrumentation__.Notify(59307)
		for pointIdx := 0; pointIdx < a.NumVertexes(); pointIdx++ {
			__antithesis_instrumentation__.Notify(59310)
			if c.PointIntersectsLinearRing(a.Vertex(pointIdx), hole) {
				__antithesis_instrumentation__.Notify(59311)
				return false
			} else {
				__antithesis_instrumentation__.Notify(59312)
			}
		}
	}
	__antithesis_instrumentation__.Notify(59301)

	return c.DistanceUpdater().OnIntersects(a.Vertex(0))
}

func onPolygonToPolygon(c DistanceCalculator, a Polygon, b Polygon) bool {
	__antithesis_instrumentation__.Notify(59313)
	aFirstPoint := a.LinearRing(0).Vertex(0)
	bFirstPoint := b.LinearRing(0).Vertex(0)

	if c.DistanceUpdater().IsMaxDistance() || func() bool {
		__antithesis_instrumentation__.Notify(59318)
		return !c.BoundingBoxIntersects() == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(59319)
		return (!c.PointIntersectsLinearRing(bFirstPoint, a.LinearRing(0)) && func() bool {
			__antithesis_instrumentation__.Notify(59320)
			return !c.PointIntersectsLinearRing(aFirstPoint, b.LinearRing(0)) == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(59321)
		return onShapeEdgesToShapeEdges(c, a.LinearRing(0), b.LinearRing(0))
	} else {
		__antithesis_instrumentation__.Notify(59322)
	}
	__antithesis_instrumentation__.Notify(59314)

	for ringIdx := 1; ringIdx < b.NumLinearRings(); ringIdx++ {
		__antithesis_instrumentation__.Notify(59323)
		bHole := b.LinearRing(ringIdx)
		if c.PointIntersectsLinearRing(aFirstPoint, bHole) {
			__antithesis_instrumentation__.Notify(59324)
			return onShapeEdgesToShapeEdges(c, a.LinearRing(0), bHole)
		} else {
			__antithesis_instrumentation__.Notify(59325)
		}
	}
	__antithesis_instrumentation__.Notify(59315)

	c.DistanceUpdater().FlipGeometries()

	defer c.DistanceUpdater().FlipGeometries()
	for ringIdx := 1; ringIdx < a.NumLinearRings(); ringIdx++ {
		__antithesis_instrumentation__.Notify(59326)
		aHole := a.LinearRing(ringIdx)
		if c.PointIntersectsLinearRing(bFirstPoint, aHole) {
			__antithesis_instrumentation__.Notify(59327)
			return onShapeEdgesToShapeEdges(c, b.LinearRing(0), aHole)
		} else {
			__antithesis_instrumentation__.Notify(59328)
		}
	}
	__antithesis_instrumentation__.Notify(59316)

	if c.PointIntersectsLinearRing(aFirstPoint, b.LinearRing(0)) {
		__antithesis_instrumentation__.Notify(59329)
		return c.DistanceUpdater().OnIntersects(aFirstPoint)
	} else {
		__antithesis_instrumentation__.Notify(59330)
	}
	__antithesis_instrumentation__.Notify(59317)
	return c.DistanceUpdater().OnIntersects(bFirstPoint)
}
