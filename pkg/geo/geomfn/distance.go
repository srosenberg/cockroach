package geomfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"

	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geodist"
	"github.com/cockroachdb/cockroach/pkg/geo/geos"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/errors"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/xy/lineintersector"
)

type geometricalObjectsOrder int

const (
	geometricalObjectsFlipped geometricalObjectsOrder = -1

	geometricalObjectsNotFlipped geometricalObjectsOrder = 1
)

func MinDistance(a geo.Geometry, b geo.Geometry) (float64, error) {
	__antithesis_instrumentation__.Notify(61721)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(61723)
		return 0, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(61724)
	}
	__antithesis_instrumentation__.Notify(61722)
	return minDistanceInternal(a, b, 0, geo.EmptyBehaviorOmit, geo.FnInclusive)
}

func MaxDistance(a geo.Geometry, b geo.Geometry) (float64, error) {
	__antithesis_instrumentation__.Notify(61725)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(61727)
		return 0, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(61728)
	}
	__antithesis_instrumentation__.Notify(61726)
	return maxDistanceInternal(a, b, math.MaxFloat64, geo.EmptyBehaviorOmit, geo.FnInclusive)
}

func DWithin(
	a geo.Geometry, b geo.Geometry, d float64, exclusivity geo.FnExclusivity,
) (bool, error) {
	__antithesis_instrumentation__.Notify(61729)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(61735)
		return false, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(61736)
	}
	__antithesis_instrumentation__.Notify(61730)
	if d < 0 {
		__antithesis_instrumentation__.Notify(61737)
		return false, pgerror.Newf(pgcode.InvalidParameterValue, "dwithin distance cannot be less than zero")
	} else {
		__antithesis_instrumentation__.Notify(61738)
	}
	__antithesis_instrumentation__.Notify(61731)
	if !a.CartesianBoundingBox().Buffer(d, d).Intersects(b.CartesianBoundingBox()) {
		__antithesis_instrumentation__.Notify(61739)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(61740)
	}
	__antithesis_instrumentation__.Notify(61732)
	dist, err := minDistanceInternal(a, b, d, geo.EmptyBehaviorError, exclusivity)
	if err != nil {
		__antithesis_instrumentation__.Notify(61741)

		if geo.IsEmptyGeometryError(err) {
			__antithesis_instrumentation__.Notify(61743)
			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(61744)
		}
		__antithesis_instrumentation__.Notify(61742)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(61745)
	}
	__antithesis_instrumentation__.Notify(61733)
	if exclusivity == geo.FnExclusive {
		__antithesis_instrumentation__.Notify(61746)
		return dist < d, nil
	} else {
		__antithesis_instrumentation__.Notify(61747)
	}
	__antithesis_instrumentation__.Notify(61734)
	return dist <= d, nil
}

func DFullyWithin(
	a geo.Geometry, b geo.Geometry, d float64, exclusivity geo.FnExclusivity,
) (bool, error) {
	__antithesis_instrumentation__.Notify(61748)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(61754)
		return false, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(61755)
	}
	__antithesis_instrumentation__.Notify(61749)
	if d < 0 {
		__antithesis_instrumentation__.Notify(61756)
		return false, pgerror.Newf(pgcode.InvalidParameterValue, "dwithin distance cannot be less than zero")
	} else {
		__antithesis_instrumentation__.Notify(61757)
	}
	__antithesis_instrumentation__.Notify(61750)
	if !a.CartesianBoundingBox().Buffer(d, d).Covers(b.CartesianBoundingBox()) {
		__antithesis_instrumentation__.Notify(61758)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(61759)
	}
	__antithesis_instrumentation__.Notify(61751)
	dist, err := maxDistanceInternal(a, b, d, geo.EmptyBehaviorError, exclusivity)
	if err != nil {
		__antithesis_instrumentation__.Notify(61760)

		if geo.IsEmptyGeometryError(err) {
			__antithesis_instrumentation__.Notify(61762)
			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(61763)
		}
		__antithesis_instrumentation__.Notify(61761)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(61764)
	}
	__antithesis_instrumentation__.Notify(61752)
	if exclusivity == geo.FnExclusive {
		__antithesis_instrumentation__.Notify(61765)
		return dist < d, nil
	} else {
		__antithesis_instrumentation__.Notify(61766)
	}
	__antithesis_instrumentation__.Notify(61753)
	return dist <= d, nil
}

func LongestLineString(a geo.Geometry, b geo.Geometry) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(61767)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(61769)
		return geo.Geometry{}, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(61770)
	}
	__antithesis_instrumentation__.Notify(61768)
	u := newGeomMaxDistanceUpdater(math.MaxFloat64, geo.FnInclusive)
	return distanceLineStringInternal(a, b, u, geo.EmptyBehaviorOmit)
}

func ShortestLineString(a geo.Geometry, b geo.Geometry) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(61771)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(61773)
		return geo.Geometry{}, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(61774)
	}
	__antithesis_instrumentation__.Notify(61772)
	u := newGeomMinDistanceUpdater(0, geo.FnInclusive)
	return distanceLineStringInternal(a, b, u, geo.EmptyBehaviorOmit)
}

func distanceLineStringInternal(
	a geo.Geometry, b geo.Geometry, u geodist.DistanceUpdater, emptyBehavior geo.EmptyBehavior,
) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(61775)
	c := &geomDistanceCalculator{updater: u, boundingBoxIntersects: a.CartesianBoundingBox().Intersects(b.CartesianBoundingBox())}
	_, err := distanceInternal(a, b, c, emptyBehavior)
	if err != nil {
		__antithesis_instrumentation__.Notify(61778)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(61779)
	}
	__antithesis_instrumentation__.Notify(61776)
	var coordA, coordB geom.Coord
	switch u := u.(type) {
	case *geomMaxDistanceUpdater:
		__antithesis_instrumentation__.Notify(61780)
		coordA = u.coordA
		coordB = u.coordB
	case *geomMinDistanceUpdater:
		__antithesis_instrumentation__.Notify(61781)
		coordA = u.coordA
		coordB = u.coordB
	default:
		__antithesis_instrumentation__.Notify(61782)
		return geo.Geometry{}, errors.AssertionFailedf("programmer error: unknown behavior")
	}
	__antithesis_instrumentation__.Notify(61777)
	lineCoords := []float64{coordA.X(), coordA.Y(), coordB.X(), coordB.Y()}
	lineString := geom.NewLineStringFlat(geom.XY, lineCoords).SetSRID(int(a.SRID()))
	return geo.MakeGeometryFromGeomT(lineString)
}

func maxDistanceInternal(
	a geo.Geometry,
	b geo.Geometry,
	stopAfter float64,
	emptyBehavior geo.EmptyBehavior,
	exclusivity geo.FnExclusivity,
) (float64, error) {
	__antithesis_instrumentation__.Notify(61783)
	u := newGeomMaxDistanceUpdater(stopAfter, exclusivity)
	c := &geomDistanceCalculator{updater: u, boundingBoxIntersects: a.CartesianBoundingBox().Intersects(b.CartesianBoundingBox())}
	return distanceInternal(a, b, c, emptyBehavior)
}

func minDistanceInternal(
	a geo.Geometry,
	b geo.Geometry,
	stopAfter float64,
	emptyBehavior geo.EmptyBehavior,
	exclusivity geo.FnExclusivity,
) (float64, error) {
	__antithesis_instrumentation__.Notify(61784)
	u := newGeomMinDistanceUpdater(stopAfter, exclusivity)
	c := &geomDistanceCalculator{updater: u, boundingBoxIntersects: a.CartesianBoundingBox().Intersects(b.CartesianBoundingBox())}
	return distanceInternal(a, b, c, emptyBehavior)
}

func distanceInternal(
	a geo.Geometry, b geo.Geometry, c geodist.DistanceCalculator, emptyBehavior geo.EmptyBehavior,
) (float64, error) {
	__antithesis_instrumentation__.Notify(61785)

	if a.Empty() || func() bool {
		__antithesis_instrumentation__.Notify(61792)
		return b.Empty() == true
	}() == true {
		__antithesis_instrumentation__.Notify(61793)
		return 0, geo.NewEmptyGeometryError()
	} else {
		__antithesis_instrumentation__.Notify(61794)
	}
	__antithesis_instrumentation__.Notify(61786)

	aGeomT, err := a.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(61795)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(61796)
	}
	__antithesis_instrumentation__.Notify(61787)
	bGeomT, err := b.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(61797)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(61798)
	}
	__antithesis_instrumentation__.Notify(61788)

	if emptyBehavior == geo.EmptyBehaviorError && func() bool {
		__antithesis_instrumentation__.Notify(61799)
		return (geo.GeomTContainsEmpty(aGeomT) || func() bool {
			__antithesis_instrumentation__.Notify(61800)
			return geo.GeomTContainsEmpty(bGeomT) == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(61801)
		return 0, geo.NewEmptyGeometryError()
	} else {
		__antithesis_instrumentation__.Notify(61802)
	}
	__antithesis_instrumentation__.Notify(61789)

	aIt := geo.NewGeomTIterator(aGeomT, emptyBehavior)
	aGeom, aNext, aErr := aIt.Next()
	if aErr != nil {
		__antithesis_instrumentation__.Notify(61803)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(61804)
	}
	__antithesis_instrumentation__.Notify(61790)
	for aNext {
		__antithesis_instrumentation__.Notify(61805)
		aGeodist, err := geomToGeodist(aGeom)
		if err != nil {
			__antithesis_instrumentation__.Notify(61809)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(61810)
		}
		__antithesis_instrumentation__.Notify(61806)

		bIt := geo.NewGeomTIterator(bGeomT, emptyBehavior)
		bGeom, bNext, bErr := bIt.Next()
		if bErr != nil {
			__antithesis_instrumentation__.Notify(61811)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(61812)
		}
		__antithesis_instrumentation__.Notify(61807)
		for bNext {
			__antithesis_instrumentation__.Notify(61813)
			bGeodist, err := geomToGeodist(bGeom)
			if err != nil {
				__antithesis_instrumentation__.Notify(61817)
				return 0, err
			} else {
				__antithesis_instrumentation__.Notify(61818)
			}
			__antithesis_instrumentation__.Notify(61814)
			earlyExit, err := geodist.ShapeDistance(c, aGeodist, bGeodist)
			if err != nil {
				__antithesis_instrumentation__.Notify(61819)
				return 0, err
			} else {
				__antithesis_instrumentation__.Notify(61820)
			}
			__antithesis_instrumentation__.Notify(61815)
			if earlyExit {
				__antithesis_instrumentation__.Notify(61821)
				return c.DistanceUpdater().Distance(), nil
			} else {
				__antithesis_instrumentation__.Notify(61822)
			}
			__antithesis_instrumentation__.Notify(61816)

			bGeom, bNext, bErr = bIt.Next()
			if bErr != nil {
				__antithesis_instrumentation__.Notify(61823)
				return 0, err
			} else {
				__antithesis_instrumentation__.Notify(61824)
			}
		}
		__antithesis_instrumentation__.Notify(61808)

		aGeom, aNext, aErr = aIt.Next()
		if aErr != nil {
			__antithesis_instrumentation__.Notify(61825)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(61826)
		}
	}
	__antithesis_instrumentation__.Notify(61791)
	return c.DistanceUpdater().Distance(), nil
}

func geomToGeodist(g geom.T) (geodist.Shape, error) {
	__antithesis_instrumentation__.Notify(61827)
	switch g := g.(type) {
	case *geom.Point:
		__antithesis_instrumentation__.Notify(61829)
		return &geodist.Point{GeomPoint: g.Coords()}, nil
	case *geom.LineString:
		__antithesis_instrumentation__.Notify(61830)
		return &geomGeodistLineString{LineString: g}, nil
	case *geom.Polygon:
		__antithesis_instrumentation__.Notify(61831)
		return &geomGeodistPolygon{Polygon: g}, nil
	}
	__antithesis_instrumentation__.Notify(61828)
	return nil, pgerror.Newf(pgcode.InvalidParameterValue, "could not find shape: %T", g)
}

type geomGeodistLineString struct {
	*geom.LineString
}

var _ geodist.LineString = (*geomGeodistLineString)(nil)

func (*geomGeodistLineString) IsShape() { __antithesis_instrumentation__.Notify(61832) }

func (*geomGeodistLineString) IsLineString() { __antithesis_instrumentation__.Notify(61833) }

func (g *geomGeodistLineString) Edge(i int) geodist.Edge {
	__antithesis_instrumentation__.Notify(61834)
	return geodist.Edge{
		V0: geodist.Point{GeomPoint: g.LineString.Coord(i)},
		V1: geodist.Point{GeomPoint: g.LineString.Coord(i + 1)},
	}
}

func (g *geomGeodistLineString) NumEdges() int {
	__antithesis_instrumentation__.Notify(61835)
	return g.LineString.NumCoords() - 1
}

func (g *geomGeodistLineString) Vertex(i int) geodist.Point {
	__antithesis_instrumentation__.Notify(61836)
	return geodist.Point{GeomPoint: g.LineString.Coord(i)}
}

func (g *geomGeodistLineString) NumVertexes() int {
	__antithesis_instrumentation__.Notify(61837)
	return g.LineString.NumCoords()
}

type geomGeodistLinearRing struct {
	*geom.LinearRing
}

var _ geodist.LinearRing = (*geomGeodistLinearRing)(nil)

func (*geomGeodistLinearRing) IsShape() { __antithesis_instrumentation__.Notify(61838) }

func (*geomGeodistLinearRing) IsLinearRing() { __antithesis_instrumentation__.Notify(61839) }

func (g *geomGeodistLinearRing) Edge(i int) geodist.Edge {
	__antithesis_instrumentation__.Notify(61840)
	return geodist.Edge{
		V0: geodist.Point{GeomPoint: g.LinearRing.Coord(i)},
		V1: geodist.Point{GeomPoint: g.LinearRing.Coord(i + 1)},
	}
}

func (g *geomGeodistLinearRing) NumEdges() int {
	__antithesis_instrumentation__.Notify(61841)
	return g.LinearRing.NumCoords() - 1
}

func (g *geomGeodistLinearRing) Vertex(i int) geodist.Point {
	__antithesis_instrumentation__.Notify(61842)
	return geodist.Point{GeomPoint: g.LinearRing.Coord(i)}
}

func (g *geomGeodistLinearRing) NumVertexes() int {
	__antithesis_instrumentation__.Notify(61843)
	return g.LinearRing.NumCoords()
}

type geomGeodistPolygon struct {
	*geom.Polygon
}

var _ geodist.Polygon = (*geomGeodistPolygon)(nil)

func (*geomGeodistPolygon) IsShape() { __antithesis_instrumentation__.Notify(61844) }

func (*geomGeodistPolygon) IsPolygon() { __antithesis_instrumentation__.Notify(61845) }

func (g *geomGeodistPolygon) LinearRing(i int) geodist.LinearRing {
	__antithesis_instrumentation__.Notify(61846)
	return &geomGeodistLinearRing{LinearRing: g.Polygon.LinearRing(i)}
}

func (g *geomGeodistPolygon) NumLinearRings() int {
	__antithesis_instrumentation__.Notify(61847)
	return g.Polygon.NumLinearRings()
}

type geomGeodistEdgeCrosser struct {
	strategy   lineintersector.Strategy
	edgeV0     geom.Coord
	edgeV1     geom.Coord
	nextEdgeV0 geom.Coord
}

var _ geodist.EdgeCrosser = (*geomGeodistEdgeCrosser)(nil)

func (c *geomGeodistEdgeCrosser) ChainCrossing(p geodist.Point) (bool, geodist.Point) {
	__antithesis_instrumentation__.Notify(61848)
	nextEdgeV1 := p.GeomPoint
	result := lineintersector.LineIntersectsLine(
		c.strategy,
		c.edgeV0,
		c.edgeV1,
		c.nextEdgeV0,
		nextEdgeV1,
	)
	c.nextEdgeV0 = nextEdgeV1
	if result.HasIntersection() {
		__antithesis_instrumentation__.Notify(61850)
		return true, geodist.Point{GeomPoint: result.Intersection()[0]}
	} else {
		__antithesis_instrumentation__.Notify(61851)
	}
	__antithesis_instrumentation__.Notify(61849)
	return false, geodist.Point{}
}

type geomMinDistanceUpdater struct {
	currentValue float64
	stopAfter    float64
	exclusivity  geo.FnExclusivity

	coordA geom.Coord

	coordB geom.Coord

	geometricalObjOrder geometricalObjectsOrder
}

var _ geodist.DistanceUpdater = (*geomMinDistanceUpdater)(nil)

func newGeomMinDistanceUpdater(
	stopAfter float64, exclusivity geo.FnExclusivity,
) *geomMinDistanceUpdater {
	__antithesis_instrumentation__.Notify(61852)
	return &geomMinDistanceUpdater{
		currentValue:        math.MaxFloat64,
		stopAfter:           stopAfter,
		exclusivity:         exclusivity,
		coordA:              nil,
		coordB:              nil,
		geometricalObjOrder: geometricalObjectsNotFlipped,
	}
}

func (u *geomMinDistanceUpdater) Distance() float64 {
	__antithesis_instrumentation__.Notify(61853)
	return u.currentValue
}

func (u *geomMinDistanceUpdater) Update(aPoint geodist.Point, bPoint geodist.Point) bool {
	__antithesis_instrumentation__.Notify(61854)
	a := aPoint.GeomPoint
	b := bPoint.GeomPoint

	dist := coordNorm(coordSub(a, b))
	if dist < u.currentValue || func() bool {
		__antithesis_instrumentation__.Notify(61856)
		return u.coordA == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(61857)
		u.currentValue = dist
		if u.geometricalObjOrder == geometricalObjectsFlipped {
			__antithesis_instrumentation__.Notify(61860)
			u.coordA = b
			u.coordB = a
		} else {
			__antithesis_instrumentation__.Notify(61861)
			u.coordA = a
			u.coordB = b
		}
		__antithesis_instrumentation__.Notify(61858)
		if u.exclusivity == geo.FnExclusive {
			__antithesis_instrumentation__.Notify(61862)
			return dist < u.stopAfter
		} else {
			__antithesis_instrumentation__.Notify(61863)
		}
		__antithesis_instrumentation__.Notify(61859)
		return dist <= u.stopAfter
	} else {
		__antithesis_instrumentation__.Notify(61864)
	}
	__antithesis_instrumentation__.Notify(61855)
	return false
}

func (u *geomMinDistanceUpdater) OnIntersects(p geodist.Point) bool {
	__antithesis_instrumentation__.Notify(61865)
	u.coordA = p.GeomPoint
	u.coordB = p.GeomPoint
	u.currentValue = 0
	return true
}

func (u *geomMinDistanceUpdater) IsMaxDistance() bool {
	__antithesis_instrumentation__.Notify(61866)
	return false
}

func (u *geomMinDistanceUpdater) FlipGeometries() {
	__antithesis_instrumentation__.Notify(61867)
	u.geometricalObjOrder = -u.geometricalObjOrder
}

type geomMaxDistanceUpdater struct {
	currentValue float64
	stopAfter    float64
	exclusivity  geo.FnExclusivity

	coordA geom.Coord

	coordB geom.Coord

	geometricalObjOrder geometricalObjectsOrder
}

var _ geodist.DistanceUpdater = (*geomMaxDistanceUpdater)(nil)

func newGeomMaxDistanceUpdater(
	stopAfter float64, exclusivity geo.FnExclusivity,
) *geomMaxDistanceUpdater {
	__antithesis_instrumentation__.Notify(61868)
	return &geomMaxDistanceUpdater{
		currentValue:        -math.MaxFloat64,
		stopAfter:           stopAfter,
		exclusivity:         exclusivity,
		coordA:              nil,
		coordB:              nil,
		geometricalObjOrder: geometricalObjectsNotFlipped,
	}
}

func (u *geomMaxDistanceUpdater) Distance() float64 {
	__antithesis_instrumentation__.Notify(61869)
	return u.currentValue
}

func (u *geomMaxDistanceUpdater) Update(aPoint geodist.Point, bPoint geodist.Point) bool {
	__antithesis_instrumentation__.Notify(61870)
	a := aPoint.GeomPoint
	b := bPoint.GeomPoint

	dist := coordNorm(coordSub(a, b))
	if dist > u.currentValue || func() bool {
		__antithesis_instrumentation__.Notify(61872)
		return u.coordA == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(61873)
		u.currentValue = dist
		if u.geometricalObjOrder == geometricalObjectsFlipped {
			__antithesis_instrumentation__.Notify(61876)
			u.coordA = b
			u.coordB = a
		} else {
			__antithesis_instrumentation__.Notify(61877)
			u.coordA = a
			u.coordB = b
		}
		__antithesis_instrumentation__.Notify(61874)
		if u.exclusivity == geo.FnExclusive {
			__antithesis_instrumentation__.Notify(61878)
			return dist >= u.stopAfter
		} else {
			__antithesis_instrumentation__.Notify(61879)
		}
		__antithesis_instrumentation__.Notify(61875)
		return dist > u.stopAfter
	} else {
		__antithesis_instrumentation__.Notify(61880)
	}
	__antithesis_instrumentation__.Notify(61871)
	return false
}

func (u *geomMaxDistanceUpdater) OnIntersects(p geodist.Point) bool {
	__antithesis_instrumentation__.Notify(61881)
	return false
}

func (u *geomMaxDistanceUpdater) IsMaxDistance() bool {
	__antithesis_instrumentation__.Notify(61882)
	return true
}

func (u *geomMaxDistanceUpdater) FlipGeometries() {
	__antithesis_instrumentation__.Notify(61883)
	u.geometricalObjOrder = -u.geometricalObjOrder
}

type geomDistanceCalculator struct {
	updater               geodist.DistanceUpdater
	boundingBoxIntersects bool
}

var _ geodist.DistanceCalculator = (*geomDistanceCalculator)(nil)

func (c *geomDistanceCalculator) DistanceUpdater() geodist.DistanceUpdater {
	__antithesis_instrumentation__.Notify(61884)
	return c.updater
}

func (c *geomDistanceCalculator) BoundingBoxIntersects() bool {
	__antithesis_instrumentation__.Notify(61885)
	return c.boundingBoxIntersects
}

func (c *geomDistanceCalculator) NewEdgeCrosser(
	edge geodist.Edge, startPoint geodist.Point,
) geodist.EdgeCrosser {
	__antithesis_instrumentation__.Notify(61886)
	return &geomGeodistEdgeCrosser{
		strategy:   &lineintersector.NonRobustLineIntersector{},
		edgeV0:     edge.V0.GeomPoint,
		edgeV1:     edge.V1.GeomPoint,
		nextEdgeV0: startPoint.GeomPoint,
	}
}

type pointSide int

const (
	pointSideLeft  pointSide = -1
	pointSideOn    pointSide = 0
	pointSideRight pointSide = 1
)

func findPointSide(p geom.Coord, eV0 geom.Coord, eV1 geom.Coord) pointSide {
	__antithesis_instrumentation__.Notify(61887)

	sign := (p.X()-eV0.X())*(eV1.Y()-eV0.Y()) - (eV1.X()-eV0.X())*(p.Y()-eV0.Y())
	switch {
	case sign == 0:
		__antithesis_instrumentation__.Notify(61888)
		return pointSideOn
	case sign > 0:
		__antithesis_instrumentation__.Notify(61889)
		return pointSideRight
	default:
		__antithesis_instrumentation__.Notify(61890)
		return pointSideLeft
	}
}

type linearRingSide int

const (
	outsideLinearRing linearRingSide = -1
	onLinearRing      linearRingSide = 0
	insideLinearRing  linearRingSide = 1
)

func findPointSideOfLinearRing(point geodist.Point, linearRing geodist.LinearRing) linearRingSide {
	__antithesis_instrumentation__.Notify(61891)

	windingNumber := 0
	p := point.GeomPoint
	for edgeIdx, numEdges := 0, linearRing.NumEdges(); edgeIdx < numEdges; edgeIdx++ {
		__antithesis_instrumentation__.Notify(61894)
		e := linearRing.Edge(edgeIdx)
		eV0 := e.V0.GeomPoint
		eV1 := e.V1.GeomPoint

		if coordEqual(eV0, eV1) {
			__antithesis_instrumentation__.Notify(61899)
			continue
		} else {
			__antithesis_instrumentation__.Notify(61900)
		}
		__antithesis_instrumentation__.Notify(61895)
		yMin := math.Min(eV0.Y(), eV1.Y())
		yMax := math.Max(eV0.Y(), eV1.Y())

		if p.Y() > yMax || func() bool {
			__antithesis_instrumentation__.Notify(61901)
			return p.Y() < yMin == true
		}() == true {
			__antithesis_instrumentation__.Notify(61902)
			continue
		} else {
			__antithesis_instrumentation__.Notify(61903)
		}
		__antithesis_instrumentation__.Notify(61896)
		side := findPointSide(p, eV0, eV1)

		if side == pointSideOn && func() bool {
			__antithesis_instrumentation__.Notify(61904)
			return (eV0.X() <= p.X() && func() bool {
				__antithesis_instrumentation__.Notify(61905)
				return p.X() <= eV1.X() == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(61906)
			return onLinearRing
		} else {
			__antithesis_instrumentation__.Notify(61907)
		}
		__antithesis_instrumentation__.Notify(61897)

		if side == pointSideLeft && func() bool {
			__antithesis_instrumentation__.Notify(61908)
			return eV0.Y() <= p.Y() == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(61909)
			return p.Y() < eV1.Y() == true
		}() == true {
			__antithesis_instrumentation__.Notify(61910)
			windingNumber++
		} else {
			__antithesis_instrumentation__.Notify(61911)
		}
		__antithesis_instrumentation__.Notify(61898)

		if side == pointSideRight && func() bool {
			__antithesis_instrumentation__.Notify(61912)
			return eV1.Y() <= p.Y() == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(61913)
			return p.Y() < eV0.Y() == true
		}() == true {
			__antithesis_instrumentation__.Notify(61914)
			windingNumber--
		} else {
			__antithesis_instrumentation__.Notify(61915)
		}
	}
	__antithesis_instrumentation__.Notify(61892)
	if windingNumber != 0 {
		__antithesis_instrumentation__.Notify(61916)
		return insideLinearRing
	} else {
		__antithesis_instrumentation__.Notify(61917)
	}
	__antithesis_instrumentation__.Notify(61893)
	return outsideLinearRing
}

func (c *geomDistanceCalculator) PointIntersectsLinearRing(
	point geodist.Point, linearRing geodist.LinearRing,
) bool {
	__antithesis_instrumentation__.Notify(61918)
	switch findPointSideOfLinearRing(point, linearRing) {
	case insideLinearRing, onLinearRing:
		__antithesis_instrumentation__.Notify(61919)
		return true
	default:
		__antithesis_instrumentation__.Notify(61920)
		return false
	}
}

func (c *geomDistanceCalculator) ClosestPointToEdge(
	e geodist.Edge, p geodist.Point,
) (geodist.Point, bool) {
	__antithesis_instrumentation__.Notify(61921)

	if coordEqual(e.V0.GeomPoint, e.V1.GeomPoint) {
		__antithesis_instrumentation__.Notify(61926)
		return e.V0, coordEqual(e.V0.GeomPoint, p.GeomPoint)
	} else {
		__antithesis_instrumentation__.Notify(61927)
	}
	__antithesis_instrumentation__.Notify(61922)

	if coordEqual(p.GeomPoint, e.V0.GeomPoint) {
		__antithesis_instrumentation__.Notify(61928)
		return p, true
	} else {
		__antithesis_instrumentation__.Notify(61929)
	}
	__antithesis_instrumentation__.Notify(61923)
	if coordEqual(p.GeomPoint, e.V1.GeomPoint) {
		__antithesis_instrumentation__.Notify(61930)
		return p, true
	} else {
		__antithesis_instrumentation__.Notify(61931)
	}
	__antithesis_instrumentation__.Notify(61924)

	ac := coordSub(p.GeomPoint, e.V0.GeomPoint)
	ab := coordSub(e.V1.GeomPoint, e.V0.GeomPoint)

	r := coordDot(ac, ab) / coordNorm2(ab)
	if r < 0 || func() bool {
		__antithesis_instrumentation__.Notify(61932)
		return r > 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(61933)
		return p, false
	} else {
		__antithesis_instrumentation__.Notify(61934)
	}
	__antithesis_instrumentation__.Notify(61925)
	return geodist.Point{GeomPoint: coordAdd(e.V0.GeomPoint, coordMul(ab, r))}, true
}

func FrechetDistance(a, b geo.Geometry) (*float64, error) {
	__antithesis_instrumentation__.Notify(61935)
	if a.Empty() || func() bool {
		__antithesis_instrumentation__.Notify(61939)
		return b.Empty() == true
	}() == true {
		__antithesis_instrumentation__.Notify(61940)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(61941)
	}
	__antithesis_instrumentation__.Notify(61936)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(61942)
		return nil, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(61943)
	}
	__antithesis_instrumentation__.Notify(61937)
	distance, err := geos.FrechetDistance(a.EWKB(), b.EWKB())
	if err != nil {
		__antithesis_instrumentation__.Notify(61944)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(61945)
	}
	__antithesis_instrumentation__.Notify(61938)
	return &distance, nil
}

func FrechetDistanceDensify(a, b geo.Geometry, densifyFrac float64) (*float64, error) {
	__antithesis_instrumentation__.Notify(61946)

	if densifyFrac <= 0 {
		__antithesis_instrumentation__.Notify(61952)
		return FrechetDistance(a, b)
	} else {
		__antithesis_instrumentation__.Notify(61953)
	}
	__antithesis_instrumentation__.Notify(61947)
	if a.Empty() || func() bool {
		__antithesis_instrumentation__.Notify(61954)
		return b.Empty() == true
	}() == true {
		__antithesis_instrumentation__.Notify(61955)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(61956)
	}
	__antithesis_instrumentation__.Notify(61948)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(61957)
		return nil, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(61958)
	}
	__antithesis_instrumentation__.Notify(61949)
	if err := verifyDensifyFrac(densifyFrac); err != nil {
		__antithesis_instrumentation__.Notify(61959)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(61960)
	}
	__antithesis_instrumentation__.Notify(61950)
	distance, err := geos.FrechetDistanceDensify(a.EWKB(), b.EWKB(), densifyFrac)
	if err != nil {
		__antithesis_instrumentation__.Notify(61961)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(61962)
	}
	__antithesis_instrumentation__.Notify(61951)
	return &distance, nil
}

func HausdorffDistance(a, b geo.Geometry) (*float64, error) {
	__antithesis_instrumentation__.Notify(61963)
	if a.Empty() || func() bool {
		__antithesis_instrumentation__.Notify(61967)
		return b.Empty() == true
	}() == true {
		__antithesis_instrumentation__.Notify(61968)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(61969)
	}
	__antithesis_instrumentation__.Notify(61964)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(61970)
		return nil, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(61971)
	}
	__antithesis_instrumentation__.Notify(61965)
	distance, err := geos.HausdorffDistance(a.EWKB(), b.EWKB())
	if err != nil {
		__antithesis_instrumentation__.Notify(61972)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(61973)
	}
	__antithesis_instrumentation__.Notify(61966)
	return &distance, nil
}

func HausdorffDistanceDensify(a, b geo.Geometry, densifyFrac float64) (*float64, error) {
	__antithesis_instrumentation__.Notify(61974)
	if a.Empty() || func() bool {
		__antithesis_instrumentation__.Notify(61979)
		return b.Empty() == true
	}() == true {
		__antithesis_instrumentation__.Notify(61980)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(61981)
	}
	__antithesis_instrumentation__.Notify(61975)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(61982)
		return nil, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(61983)
	}
	__antithesis_instrumentation__.Notify(61976)
	if err := verifyDensifyFrac(densifyFrac); err != nil {
		__antithesis_instrumentation__.Notify(61984)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(61985)
	}
	__antithesis_instrumentation__.Notify(61977)

	distance, err := geos.HausdorffDistanceDensify(a.EWKB(), b.EWKB(), densifyFrac)
	if err != nil {
		__antithesis_instrumentation__.Notify(61986)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(61987)
	}
	__antithesis_instrumentation__.Notify(61978)
	return &distance, nil
}

func ClosestPoint(a, b geo.Geometry) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(61988)
	shortestLine, err := ShortestLineString(a, b)
	if err != nil {
		__antithesis_instrumentation__.Notify(61992)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(61993)
	}
	__antithesis_instrumentation__.Notify(61989)
	shortestLineT, err := shortestLine.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(61994)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(61995)
	}
	__antithesis_instrumentation__.Notify(61990)
	closestPoint, err := geo.MakeGeometryFromPointCoords(
		shortestLineT.(*geom.LineString).Coord(0).X(),
		shortestLineT.(*geom.LineString).Coord(0).Y())
	if err != nil {
		__antithesis_instrumentation__.Notify(61996)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(61997)
	}
	__antithesis_instrumentation__.Notify(61991)
	return closestPoint, nil
}

func verifyDensifyFrac(f float64) error {
	__antithesis_instrumentation__.Notify(61998)
	if f < 0 || func() bool {
		__antithesis_instrumentation__.Notify(62001)
		return f > 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(62002)
		return pgerror.Newf(pgcode.InvalidParameterValue, "fraction must be in range [0, 1], got %f", f)
	} else {
		__antithesis_instrumentation__.Notify(62003)
	}
	__antithesis_instrumentation__.Notify(61999)

	const fracTooSmall = 1e-6
	if f > 0 && func() bool {
		__antithesis_instrumentation__.Notify(62004)
		return f < fracTooSmall == true
	}() == true {
		__antithesis_instrumentation__.Notify(62005)
		return pgerror.Newf(pgcode.InvalidParameterValue, "fraction %f is too small, must be at least %f", f, fracTooSmall)
	} else {
		__antithesis_instrumentation__.Notify(62006)
	}
	__antithesis_instrumentation__.Notify(62000)
	return nil
}

func findPointSideOfPolygon(point geom.T, polygon geom.T) (linearRingSide, error) {
	__antithesis_instrumentation__.Notify(62007)

	_, ok := point.(*geom.Point)
	if !ok {
		__antithesis_instrumentation__.Notify(62017)
		return outsideLinearRing, pgerror.Newf(pgcode.InvalidParameterValue, "first geometry passed to findPointSideOfPolygon must be a point")
	} else {
		__antithesis_instrumentation__.Notify(62018)
	}
	__antithesis_instrumentation__.Notify(62008)
	pointGeodistShape, err := geomToGeodist(point)
	if err != nil {
		__antithesis_instrumentation__.Notify(62019)
		return outsideLinearRing, err
	} else {
		__antithesis_instrumentation__.Notify(62020)
	}
	__antithesis_instrumentation__.Notify(62009)
	pointGeodistPoint, ok := pointGeodistShape.(*geodist.Point)
	if !ok {
		__antithesis_instrumentation__.Notify(62021)
		return outsideLinearRing, pgerror.Newf(pgcode.InvalidParameterValue, "geomToGeodist failed to convert a *geom.Point to a *geodist.Point")
	} else {
		__antithesis_instrumentation__.Notify(62022)
	}
	__antithesis_instrumentation__.Notify(62010)

	_, ok = polygon.(*geom.Polygon)
	if !ok {
		__antithesis_instrumentation__.Notify(62023)
		return outsideLinearRing, pgerror.Newf(pgcode.InvalidParameterValue, "second geometry passed to findPointSideOfPolygon must be a polygon")
	} else {
		__antithesis_instrumentation__.Notify(62024)
	}
	__antithesis_instrumentation__.Notify(62011)
	polygonGeodistShape, err := geomToGeodist(polygon)
	if err != nil {
		__antithesis_instrumentation__.Notify(62025)
		return outsideLinearRing, err
	} else {
		__antithesis_instrumentation__.Notify(62026)
	}
	__antithesis_instrumentation__.Notify(62012)
	polygonGeodistPolygon, ok := polygonGeodistShape.(geodist.Polygon)
	if !ok {
		__antithesis_instrumentation__.Notify(62027)
		return outsideLinearRing, pgerror.Newf(pgcode.InvalidParameterValue, "geomToGeodist failed to convert a *geom.Polygon to a geodist.Polygon")
	} else {
		__antithesis_instrumentation__.Notify(62028)
	}
	__antithesis_instrumentation__.Notify(62013)

	if polygonGeodistPolygon.NumLinearRings() == 0 {
		__antithesis_instrumentation__.Notify(62029)
		return outsideLinearRing, nil
	} else {
		__antithesis_instrumentation__.Notify(62030)
	}
	__antithesis_instrumentation__.Notify(62014)

	mainRing := polygonGeodistPolygon.LinearRing(0)
	switch pointSide := findPointSideOfLinearRing(*pointGeodistPoint, mainRing); pointSide {
	case insideLinearRing:
		__antithesis_instrumentation__.Notify(62031)
	case outsideLinearRing, onLinearRing:
		__antithesis_instrumentation__.Notify(62032)
		return pointSide, nil
	default:
		__antithesis_instrumentation__.Notify(62033)
		return outsideLinearRing, pgerror.Newf(pgcode.InvalidParameterValue, "unknown linearRingSide %d", pointSide)
	}
	__antithesis_instrumentation__.Notify(62015)

	for ringNum := 1; ringNum < polygonGeodistPolygon.NumLinearRings(); ringNum++ {
		__antithesis_instrumentation__.Notify(62034)
		polygonHole := polygonGeodistPolygon.LinearRing(ringNum)
		switch pointSide := findPointSideOfLinearRing(*pointGeodistPoint, polygonHole); pointSide {
		case insideLinearRing:
			__antithesis_instrumentation__.Notify(62035)
			return outsideLinearRing, nil
		case onLinearRing:
			__antithesis_instrumentation__.Notify(62036)
			return onLinearRing, nil
		case outsideLinearRing:
			__antithesis_instrumentation__.Notify(62037)
			continue
		default:
			__antithesis_instrumentation__.Notify(62038)
			return outsideLinearRing, pgerror.Newf(pgcode.InvalidParameterValue, "unknown linearRingSide %d", pointSide)
		}
	}
	__antithesis_instrumentation__.Notify(62016)

	return insideLinearRing, nil
}
