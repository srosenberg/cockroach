package geogfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"

	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geodist"
	"github.com/cockroachdb/cockroach/pkg/geo/geographiclib"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

const SpheroidErrorFraction = 0.05

func Distance(
	a geo.Geography, b geo.Geography, useSphereOrSpheroid UseSphereOrSpheroid,
) (float64, error) {
	__antithesis_instrumentation__.Notify(59603)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(59608)
		return 0, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(59609)
	}
	__antithesis_instrumentation__.Notify(59604)

	aRegions, err := a.AsS2(geo.EmptyBehaviorError)
	if err != nil {
		__antithesis_instrumentation__.Notify(59610)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(59611)
	}
	__antithesis_instrumentation__.Notify(59605)
	bRegions, err := b.AsS2(geo.EmptyBehaviorError)
	if err != nil {
		__antithesis_instrumentation__.Notify(59612)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(59613)
	}
	__antithesis_instrumentation__.Notify(59606)
	spheroid, err := a.Spheroid()
	if err != nil {
		__antithesis_instrumentation__.Notify(59614)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(59615)
	}
	__antithesis_instrumentation__.Notify(59607)
	return distanceGeographyRegions(
		spheroid,
		useSphereOrSpheroid,
		aRegions,
		bRegions,
		a.BoundingRect().Intersects(b.BoundingRect()),
		0,
		geo.FnInclusive,
	)
}

type s2GeodistLineString struct {
	*s2.Polyline
}

var _ geodist.LineString = (*s2GeodistLineString)(nil)

func (*s2GeodistLineString) IsShape() { __antithesis_instrumentation__.Notify(59616) }

func (*s2GeodistLineString) IsLineString() { __antithesis_instrumentation__.Notify(59617) }

func (g *s2GeodistLineString) Edge(i int) geodist.Edge {
	__antithesis_instrumentation__.Notify(59618)
	return geodist.Edge{
		V0: geodist.Point{GeogPoint: (*g.Polyline)[i]},
		V1: geodist.Point{GeogPoint: (*g.Polyline)[i+1]},
	}
}

func (g *s2GeodistLineString) NumEdges() int {
	__antithesis_instrumentation__.Notify(59619)
	return len(*g.Polyline) - 1
}

func (g *s2GeodistLineString) Vertex(i int) geodist.Point {
	__antithesis_instrumentation__.Notify(59620)
	return geodist.Point{
		GeogPoint: (*g.Polyline)[i],
	}
}

func (g *s2GeodistLineString) NumVertexes() int {
	__antithesis_instrumentation__.Notify(59621)
	return len(*g.Polyline)
}

type s2GeodistLinearRing struct {
	*s2.Loop
}

var _ geodist.LinearRing = (*s2GeodistLinearRing)(nil)

func (*s2GeodistLinearRing) IsShape() { __antithesis_instrumentation__.Notify(59622) }

func (*s2GeodistLinearRing) IsLinearRing() { __antithesis_instrumentation__.Notify(59623) }

func (g *s2GeodistLinearRing) Edge(i int) geodist.Edge {
	__antithesis_instrumentation__.Notify(59624)
	return geodist.Edge{
		V0: geodist.Point{GeogPoint: g.Loop.Vertex(i)},
		V1: geodist.Point{GeogPoint: g.Loop.Vertex(i + 1)},
	}
}

func (g *s2GeodistLinearRing) NumEdges() int {
	__antithesis_instrumentation__.Notify(59625)
	return g.Loop.NumEdges()
}

func (g *s2GeodistLinearRing) Vertex(i int) geodist.Point {
	__antithesis_instrumentation__.Notify(59626)
	return geodist.Point{
		GeogPoint: g.Loop.Vertex(i),
	}
}

func (g *s2GeodistLinearRing) NumVertexes() int {
	__antithesis_instrumentation__.Notify(59627)
	return g.Loop.NumVertices()
}

type s2GeodistPolygon struct {
	*s2.Polygon
}

var _ geodist.Polygon = (*s2GeodistPolygon)(nil)

func (*s2GeodistPolygon) IsShape() { __antithesis_instrumentation__.Notify(59628) }

func (*s2GeodistPolygon) IsPolygon() { __antithesis_instrumentation__.Notify(59629) }

func (g *s2GeodistPolygon) LinearRing(i int) geodist.LinearRing {
	__antithesis_instrumentation__.Notify(59630)
	return &s2GeodistLinearRing{Loop: g.Polygon.Loop(i)}
}

func (g *s2GeodistPolygon) NumLinearRings() int {
	__antithesis_instrumentation__.Notify(59631)
	return g.Polygon.NumLoops()
}

type s2GeodistEdgeCrosser struct {
	*s2.EdgeCrosser
}

var _ geodist.EdgeCrosser = (*s2GeodistEdgeCrosser)(nil)

func (c *s2GeodistEdgeCrosser) ChainCrossing(p geodist.Point) (bool, geodist.Point) {
	__antithesis_instrumentation__.Notify(59632)

	return c.EdgeCrosser.ChainCrossingSign(p.GeogPoint) != s2.DoNotCross, geodist.Point{}
}

func distanceGeographyRegions(
	spheroid *geographiclib.Spheroid,
	useSphereOrSpheroid UseSphereOrSpheroid,
	aRegions []s2.Region,
	bRegions []s2.Region,
	boundingBoxIntersects bool,
	stopAfter float64,
	exclusivity geo.FnExclusivity,
) (float64, error) {
	__antithesis_instrumentation__.Notify(59633)
	minDistance := math.MaxFloat64
	for _, aRegion := range aRegions {
		__antithesis_instrumentation__.Notify(59635)
		aGeodist, err := regionToGeodistShape(aRegion)
		if err != nil {
			__antithesis_instrumentation__.Notify(59637)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(59638)
		}
		__antithesis_instrumentation__.Notify(59636)
		for _, bRegion := range bRegions {
			__antithesis_instrumentation__.Notify(59639)
			minDistanceUpdater := newGeographyMinDistanceUpdater(
				spheroid,
				useSphereOrSpheroid,
				stopAfter,
				exclusivity,
			)
			bGeodist, err := regionToGeodistShape(bRegion)
			if err != nil {
				__antithesis_instrumentation__.Notify(59642)
				return 0, err
			} else {
				__antithesis_instrumentation__.Notify(59643)
			}
			__antithesis_instrumentation__.Notify(59640)
			earlyExit, err := geodist.ShapeDistance(
				&geographyDistanceCalculator{
					updater:               minDistanceUpdater,
					boundingBoxIntersects: boundingBoxIntersects,
				},
				aGeodist,
				bGeodist,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(59644)
				return 0, err
			} else {
				__antithesis_instrumentation__.Notify(59645)
			}
			__antithesis_instrumentation__.Notify(59641)
			minDistance = math.Min(minDistance, minDistanceUpdater.Distance())
			if earlyExit {
				__antithesis_instrumentation__.Notify(59646)
				return minDistance, nil
			} else {
				__antithesis_instrumentation__.Notify(59647)
			}
		}
	}
	__antithesis_instrumentation__.Notify(59634)
	return minDistance, nil
}

type geographyMinDistanceUpdater struct {
	spheroid            *geographiclib.Spheroid
	useSphereOrSpheroid UseSphereOrSpheroid
	minEdge             s2.Edge
	minD                s1.ChordAngle
	stopAfter           s1.ChordAngle
	exclusivity         geo.FnExclusivity
}

var _ geodist.DistanceUpdater = (*geographyMinDistanceUpdater)(nil)

func newGeographyMinDistanceUpdater(
	spheroid *geographiclib.Spheroid,
	useSphereOrSpheroid UseSphereOrSpheroid,
	stopAfter float64,
	exclusivity geo.FnExclusivity,
) *geographyMinDistanceUpdater {
	__antithesis_instrumentation__.Notify(59648)
	multiplier := 1.0
	if useSphereOrSpheroid == UseSpheroid {
		__antithesis_instrumentation__.Notify(59650)

		multiplier -= SpheroidErrorFraction
	} else {
		__antithesis_instrumentation__.Notify(59651)
	}
	__antithesis_instrumentation__.Notify(59649)
	stopAfterChordAngle := s1.ChordAngleFromAngle(s1.Angle(stopAfter * multiplier / spheroid.SphereRadius))
	return &geographyMinDistanceUpdater{
		spheroid:            spheroid,
		minD:                math.MaxFloat64,
		useSphereOrSpheroid: useSphereOrSpheroid,
		stopAfter:           stopAfterChordAngle,
		exclusivity:         exclusivity,
	}
}

func (u *geographyMinDistanceUpdater) Distance() float64 {
	__antithesis_instrumentation__.Notify(59652)

	if u.minD == 0 {
		__antithesis_instrumentation__.Notify(59655)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(59656)
	}
	__antithesis_instrumentation__.Notify(59653)
	if u.useSphereOrSpheroid == UseSpheroid {
		__antithesis_instrumentation__.Notify(59657)
		return spheroidDistance(u.spheroid, u.minEdge.V0, u.minEdge.V1)
	} else {
		__antithesis_instrumentation__.Notify(59658)
	}
	__antithesis_instrumentation__.Notify(59654)
	return u.minD.Angle().Radians() * u.spheroid.SphereRadius
}

func (u *geographyMinDistanceUpdater) Update(aPoint geodist.Point, bPoint geodist.Point) bool {
	__antithesis_instrumentation__.Notify(59659)
	a := aPoint.GeogPoint
	b := bPoint.GeogPoint

	sphereDistance := s2.ChordAngleBetweenPoints(a, b)
	if sphereDistance < u.minD {
		__antithesis_instrumentation__.Notify(59661)
		u.minD = sphereDistance
		u.minEdge = s2.Edge{V0: a, V1: b}

		if (u.exclusivity == geo.FnInclusive && func() bool {
			__antithesis_instrumentation__.Notify(59662)
			return u.minD <= u.stopAfter == true
		}() == true) || func() bool {
			__antithesis_instrumentation__.Notify(59663)
			return (u.exclusivity == geo.FnExclusive && func() bool {
				__antithesis_instrumentation__.Notify(59664)
				return u.minD < u.stopAfter == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(59665)
			return true
		} else {
			__antithesis_instrumentation__.Notify(59666)
		}
	} else {
		__antithesis_instrumentation__.Notify(59667)
	}
	__antithesis_instrumentation__.Notify(59660)
	return false
}

func (u *geographyMinDistanceUpdater) OnIntersects(p geodist.Point) bool {
	__antithesis_instrumentation__.Notify(59668)
	u.minD = 0
	return true
}

func (u *geographyMinDistanceUpdater) IsMaxDistance() bool {
	__antithesis_instrumentation__.Notify(59669)
	return false
}

func (u *geographyMinDistanceUpdater) FlipGeometries() {
	__antithesis_instrumentation__.Notify(59670)

}

type geographyDistanceCalculator struct {
	updater               *geographyMinDistanceUpdater
	boundingBoxIntersects bool
}

var _ geodist.DistanceCalculator = (*geographyDistanceCalculator)(nil)

func (c *geographyDistanceCalculator) DistanceUpdater() geodist.DistanceUpdater {
	__antithesis_instrumentation__.Notify(59671)
	return c.updater
}

func (c *geographyDistanceCalculator) BoundingBoxIntersects() bool {
	__antithesis_instrumentation__.Notify(59672)
	return c.boundingBoxIntersects
}

func (c *geographyDistanceCalculator) NewEdgeCrosser(
	edge geodist.Edge, startPoint geodist.Point,
) geodist.EdgeCrosser {
	__antithesis_instrumentation__.Notify(59673)
	return &s2GeodistEdgeCrosser{
		EdgeCrosser: s2.NewChainEdgeCrosser(
			edge.V0.GeogPoint,
			edge.V1.GeogPoint,
			startPoint.GeogPoint,
		),
	}
}

func (c *geographyDistanceCalculator) PointIntersectsLinearRing(
	point geodist.Point, polygon geodist.LinearRing,
) bool {
	__antithesis_instrumentation__.Notify(59674)
	return polygon.(*s2GeodistLinearRing).ContainsPoint(point.GeogPoint)
}

func (c *geographyDistanceCalculator) ClosestPointToEdge(
	edge geodist.Edge, point geodist.Point,
) (geodist.Point, bool) {
	__antithesis_instrumentation__.Notify(59675)
	eV0 := edge.V0.GeogPoint
	eV1 := edge.V1.GeogPoint

	normal := eV0.Vector.Cross(eV1.Vector).Normalize()

	normalScaledToPoint := normal.Mul(normal.Dot(point.GeogPoint.Vector))

	closestPoint := s2.Point{Vector: point.GeogPoint.Vector.Sub(normalScaledToPoint).Normalize()}

	return geodist.Point{GeogPoint: closestPoint}, (&s2.Polyline{eV0, eV1}).IntersectsCell(s2.CellFromPoint(closestPoint))
}

func regionToGeodistShape(r s2.Region) (geodist.Shape, error) {
	__antithesis_instrumentation__.Notify(59676)
	switch r := r.(type) {
	case s2.Point:
		__antithesis_instrumentation__.Notify(59678)
		return &geodist.Point{GeogPoint: r}, nil
	case *s2.Polyline:
		__antithesis_instrumentation__.Notify(59679)
		return &s2GeodistLineString{Polyline: r}, nil
	case *s2.Polygon:
		__antithesis_instrumentation__.Notify(59680)
		return &s2GeodistPolygon{Polygon: r}, nil
	}
	__antithesis_instrumentation__.Notify(59677)
	return nil, pgerror.Newf(pgcode.InvalidParameterValue, "unknown region: %T", r)
}
