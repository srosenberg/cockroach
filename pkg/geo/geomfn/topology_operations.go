package geomfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"

	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/geo/geos"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
)

func Boundary(g geo.Geometry) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(63314)

	if g.Empty() {
		__antithesis_instrumentation__.Notify(63317)
		return g, nil
	} else {
		__antithesis_instrumentation__.Notify(63318)
	}
	__antithesis_instrumentation__.Notify(63315)
	boundaryEWKB, err := geos.Boundary(g.EWKB())
	if err != nil {
		__antithesis_instrumentation__.Notify(63319)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(63320)
	}
	__antithesis_instrumentation__.Notify(63316)
	return geo.ParseGeometryFromEWKB(boundaryEWKB)
}

func Centroid(g geo.Geometry) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(63321)
	centroidEWKB, err := geos.Centroid(g.EWKB())
	if err != nil {
		__antithesis_instrumentation__.Notify(63323)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(63324)
	}
	__antithesis_instrumentation__.Notify(63322)
	return geo.ParseGeometryFromEWKB(centroidEWKB)
}

func MinimumBoundingCircle(g geo.Geometry) (geo.Geometry, geo.Geometry, float64, error) {
	__antithesis_instrumentation__.Notify(63325)
	if BoundingBoxHasInfiniteCoordinates(g) {
		__antithesis_instrumentation__.Notify(63330)
		return geo.Geometry{}, geo.Geometry{}, 0, pgerror.Newf(pgcode.InvalidParameterValue, "value out of range: overflow")
	} else {
		__antithesis_instrumentation__.Notify(63331)
	}
	__antithesis_instrumentation__.Notify(63326)

	polygonEWKB, centroidEWKB, radius, err := geos.MinimumBoundingCircle(g.EWKB())
	if err != nil {
		__antithesis_instrumentation__.Notify(63332)
		return geo.Geometry{}, geo.Geometry{}, 0, err
	} else {
		__antithesis_instrumentation__.Notify(63333)
	}
	__antithesis_instrumentation__.Notify(63327)

	polygon, err := geo.ParseGeometryFromEWKB(polygonEWKB)
	if err != nil {
		__antithesis_instrumentation__.Notify(63334)
		return geo.Geometry{}, geo.Geometry{}, 0, err
	} else {
		__antithesis_instrumentation__.Notify(63335)
	}
	__antithesis_instrumentation__.Notify(63328)

	centroid, err := geo.ParseGeometryFromEWKB(centroidEWKB)
	if err != nil {
		__antithesis_instrumentation__.Notify(63336)
		return geo.Geometry{}, geo.Geometry{}, 0, err
	} else {
		__antithesis_instrumentation__.Notify(63337)
	}
	__antithesis_instrumentation__.Notify(63329)

	return polygon, centroid, radius, nil
}

func ClipByRect(g geo.Geometry, b geo.CartesianBoundingBox) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(63338)
	if g.Empty() {
		__antithesis_instrumentation__.Notify(63341)
		return g, nil
	} else {
		__antithesis_instrumentation__.Notify(63342)
	}
	__antithesis_instrumentation__.Notify(63339)
	clipByRectEWKB, err := geos.ClipByRect(g.EWKB(), b.LoX, b.LoY, b.HiX, b.HiY)
	if err != nil {
		__antithesis_instrumentation__.Notify(63343)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(63344)
	}
	__antithesis_instrumentation__.Notify(63340)
	return geo.ParseGeometryFromEWKB(clipByRectEWKB)
}

func ConvexHull(g geo.Geometry) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(63345)
	convexHullEWKB, err := geos.ConvexHull(g.EWKB())
	if err != nil {
		__antithesis_instrumentation__.Notify(63347)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(63348)
	}
	__antithesis_instrumentation__.Notify(63346)
	return geo.ParseGeometryFromEWKB(convexHullEWKB)
}

func Difference(a, b geo.Geometry) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(63349)

	if a.Empty() || func() bool {
		__antithesis_instrumentation__.Notify(63353)
		return b.Empty() == true
	}() == true {
		__antithesis_instrumentation__.Notify(63354)
		return a, nil
	} else {
		__antithesis_instrumentation__.Notify(63355)
	}
	__antithesis_instrumentation__.Notify(63350)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(63356)
		return geo.Geometry{}, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(63357)
	}
	__antithesis_instrumentation__.Notify(63351)
	diffEWKB, err := geos.Difference(a.EWKB(), b.EWKB())
	if err != nil {
		__antithesis_instrumentation__.Notify(63358)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(63359)
	}
	__antithesis_instrumentation__.Notify(63352)
	return geo.ParseGeometryFromEWKB(diffEWKB)
}

func SimplifyGEOS(g geo.Geometry, tolerance float64) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(63360)
	if math.IsNaN(tolerance) || func() bool {
		__antithesis_instrumentation__.Notify(63363)
		return g.ShapeType2D() == geopb.ShapeType_Point == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(63364)
		return g.ShapeType2D() == geopb.ShapeType_MultiPoint == true
	}() == true {
		__antithesis_instrumentation__.Notify(63365)
		return g, nil
	} else {
		__antithesis_instrumentation__.Notify(63366)
	}
	__antithesis_instrumentation__.Notify(63361)
	simplifiedEWKB, err := geos.Simplify(g.EWKB(), tolerance)
	if err != nil {
		__antithesis_instrumentation__.Notify(63367)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(63368)
	}
	__antithesis_instrumentation__.Notify(63362)
	return geo.ParseGeometryFromEWKB(simplifiedEWKB)
}

func SimplifyPreserveTopology(g geo.Geometry, tolerance float64) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(63369)
	simplifiedEWKB, err := geos.TopologyPreserveSimplify(g.EWKB(), tolerance)
	if err != nil {
		__antithesis_instrumentation__.Notify(63371)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(63372)
	}
	__antithesis_instrumentation__.Notify(63370)
	return geo.ParseGeometryFromEWKB(simplifiedEWKB)
}

func PointOnSurface(g geo.Geometry) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(63373)
	pointOnSurfaceEWKB, err := geos.PointOnSurface(g.EWKB())
	if err != nil {
		__antithesis_instrumentation__.Notify(63375)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(63376)
	}
	__antithesis_instrumentation__.Notify(63374)
	return geo.ParseGeometryFromEWKB(pointOnSurfaceEWKB)
}

func Intersection(a geo.Geometry, b geo.Geometry) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(63377)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(63382)
		return geo.Geometry{}, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(63383)
	}
	__antithesis_instrumentation__.Notify(63378)

	if a.Empty() {
		__antithesis_instrumentation__.Notify(63384)
		return a, nil
	} else {
		__antithesis_instrumentation__.Notify(63385)
	}
	__antithesis_instrumentation__.Notify(63379)
	if b.Empty() {
		__antithesis_instrumentation__.Notify(63386)
		return b, nil
	} else {
		__antithesis_instrumentation__.Notify(63387)
	}
	__antithesis_instrumentation__.Notify(63380)
	retEWKB, err := geos.Intersection(a.EWKB(), b.EWKB())
	if err != nil {
		__antithesis_instrumentation__.Notify(63388)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(63389)
	}
	__antithesis_instrumentation__.Notify(63381)
	return geo.ParseGeometryFromEWKB(retEWKB)
}

func UnaryUnion(g geo.Geometry) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(63390)
	retEWKB, err := geos.UnaryUnion(g.EWKB())
	if err != nil {
		__antithesis_instrumentation__.Notify(63392)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(63393)
	}
	__antithesis_instrumentation__.Notify(63391)
	return geo.ParseGeometryFromEWKB(retEWKB)
}

func Union(a geo.Geometry, b geo.Geometry) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(63394)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(63397)
		return geo.Geometry{}, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(63398)
	}
	__antithesis_instrumentation__.Notify(63395)
	retEWKB, err := geos.Union(a.EWKB(), b.EWKB())
	if err != nil {
		__antithesis_instrumentation__.Notify(63399)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(63400)
	}
	__antithesis_instrumentation__.Notify(63396)
	return geo.ParseGeometryFromEWKB(retEWKB)
}

func SymDifference(a geo.Geometry, b geo.Geometry) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(63401)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(63404)
		return geo.Geometry{}, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(63405)
	}
	__antithesis_instrumentation__.Notify(63402)
	retEWKB, err := geos.SymDifference(a.EWKB(), b.EWKB())
	if err != nil {
		__antithesis_instrumentation__.Notify(63406)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(63407)
	}
	__antithesis_instrumentation__.Notify(63403)
	return geo.ParseGeometryFromEWKB(retEWKB)
}

func SharedPaths(a geo.Geometry, b geo.Geometry) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(63408)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(63412)
		return geo.Geometry{}, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(63413)
	}
	__antithesis_instrumentation__.Notify(63409)
	paths, err := geos.SharedPaths(a.EWKB(), b.EWKB())
	if err != nil {
		__antithesis_instrumentation__.Notify(63414)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(63415)
	}
	__antithesis_instrumentation__.Notify(63410)
	gm, err := geo.ParseGeometryFromEWKB(paths)
	if err != nil {
		__antithesis_instrumentation__.Notify(63416)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(63417)
	}
	__antithesis_instrumentation__.Notify(63411)
	return gm, nil
}

func MinimumRotatedRectangle(g geo.Geometry) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(63418)
	paths, err := geos.MinimumRotatedRectangle(g.EWKB())
	if err != nil {
		__antithesis_instrumentation__.Notify(63421)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(63422)
	}
	__antithesis_instrumentation__.Notify(63419)
	gm, err := geo.ParseGeometryFromEWKB(paths)
	if err != nil {
		__antithesis_instrumentation__.Notify(63423)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(63424)
	}
	__antithesis_instrumentation__.Notify(63420)
	return gm, nil
}

func BoundingBoxHasInfiniteCoordinates(g geo.Geometry) bool {
	__antithesis_instrumentation__.Notify(63425)
	boundingBox := g.BoundingBoxRef()
	if boundingBox == nil {
		__antithesis_instrumentation__.Notify(63429)
		return false
	} else {
		__antithesis_instrumentation__.Notify(63430)
	}
	__antithesis_instrumentation__.Notify(63426)

	isInf := func(ord float64) bool {
		__antithesis_instrumentation__.Notify(63431)
		return math.IsInf(ord, 0)
	}
	__antithesis_instrumentation__.Notify(63427)
	if isInf(boundingBox.LoX) || func() bool {
		__antithesis_instrumentation__.Notify(63432)
		return isInf(boundingBox.LoY) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(63433)
		return isInf(boundingBox.HiX) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(63434)
		return isInf(boundingBox.HiY) == true
	}() == true {
		__antithesis_instrumentation__.Notify(63435)
		return true
	} else {
		__antithesis_instrumentation__.Notify(63436)
	}
	__antithesis_instrumentation__.Notify(63428)
	return false
}

func BoundingBoxHasNaNCoordinates(g geo.Geometry) bool {
	__antithesis_instrumentation__.Notify(63437)
	boundingBox := g.BoundingBoxRef()
	if boundingBox == nil {
		__antithesis_instrumentation__.Notify(63441)
		return false
	} else {
		__antithesis_instrumentation__.Notify(63442)
	}
	__antithesis_instrumentation__.Notify(63438)

	isNaN := func(ord float64) bool {
		__antithesis_instrumentation__.Notify(63443)
		return math.IsNaN(ord)
	}
	__antithesis_instrumentation__.Notify(63439)
	if isNaN(boundingBox.LoX) || func() bool {
		__antithesis_instrumentation__.Notify(63444)
		return isNaN(boundingBox.LoY) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(63445)
		return isNaN(boundingBox.HiX) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(63446)
		return isNaN(boundingBox.HiY) == true
	}() == true {
		__antithesis_instrumentation__.Notify(63447)
		return true
	} else {
		__antithesis_instrumentation__.Notify(63448)
	}
	__antithesis_instrumentation__.Notify(63440)
	return false
}
