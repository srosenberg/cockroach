package geomfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"

	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geosegmentize"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/twpayne/go-geom"
)

func Segmentize(g geo.Geometry, segmentMaxLength float64) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(62917)
	if math.IsNaN(segmentMaxLength) || func() bool {
		__antithesis_instrumentation__.Notify(62920)
		return math.IsInf(segmentMaxLength, 1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(62921)
		return g, nil
	} else {
		__antithesis_instrumentation__.Notify(62922)
	}
	__antithesis_instrumentation__.Notify(62918)
	geometry, err := g.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(62923)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(62924)
	}
	__antithesis_instrumentation__.Notify(62919)
	switch geometry := geometry.(type) {
	case *geom.Point, *geom.MultiPoint:
		__antithesis_instrumentation__.Notify(62925)
		return g, nil
	default:
		__antithesis_instrumentation__.Notify(62926)
		if segmentMaxLength <= 0 {
			__antithesis_instrumentation__.Notify(62929)
			return geo.Geometry{}, pgerror.Newf(pgcode.InvalidParameterValue, "maximum segment length must be positive")
		} else {
			__antithesis_instrumentation__.Notify(62930)
		}
		__antithesis_instrumentation__.Notify(62927)
		segGeometry, err := geosegmentize.Segmentize(geometry, segmentMaxLength, segmentizeCoords)
		if err != nil {
			__antithesis_instrumentation__.Notify(62931)
			return geo.Geometry{}, err
		} else {
			__antithesis_instrumentation__.Notify(62932)
		}
		__antithesis_instrumentation__.Notify(62928)
		return geo.MakeGeometryFromGeomT(segGeometry)
	}
}

func segmentizeCoords(a geom.Coord, b geom.Coord, maxSegmentLength float64) ([]float64, error) {
	__antithesis_instrumentation__.Notify(62933)
	if len(a) != len(b) {
		__antithesis_instrumentation__.Notify(62938)
		return nil, pgerror.Newf(pgcode.InvalidParameterValue, "cannot segmentize two coordinates of different dimensions")
	} else {
		__antithesis_instrumentation__.Notify(62939)
	}
	__antithesis_instrumentation__.Notify(62934)
	if maxSegmentLength <= 0 {
		__antithesis_instrumentation__.Notify(62940)
		return nil, pgerror.Newf(pgcode.InvalidParameterValue, "maximum segment length must be positive")
	} else {
		__antithesis_instrumentation__.Notify(62941)
	}
	__antithesis_instrumentation__.Notify(62935)

	distanceBetweenPoints := math.Sqrt(math.Pow(a.X()-b.X(), 2) + math.Pow(b.Y()-a.Y(), 2))

	doubleNumberOfSegmentsToCreate := math.Ceil(distanceBetweenPoints / maxSegmentLength)
	doubleNumPoints := float64(len(a)) * (1 + doubleNumberOfSegmentsToCreate)
	if err := geosegmentize.CheckSegmentizeValidNumPoints(doubleNumPoints, a, b); err != nil {
		__antithesis_instrumentation__.Notify(62942)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(62943)
	}
	__antithesis_instrumentation__.Notify(62936)

	numberOfSegmentsToCreate := int(doubleNumberOfSegmentsToCreate)
	numPoints := int(doubleNumPoints)

	allSegmentizedCoordinates := make([]float64, 0, numPoints)
	allSegmentizedCoordinates = append(allSegmentizedCoordinates, a.Clone()...)
	segmentFraction := 1.0 / float64(numberOfSegmentsToCreate)
	for pointInserted := 1; pointInserted < numberOfSegmentsToCreate; pointInserted++ {
		__antithesis_instrumentation__.Notify(62944)
		segmentPoint := make([]float64, 0, len(a))
		for i := 0; i < len(a); i++ {
			__antithesis_instrumentation__.Notify(62946)
			segmentPoint = append(
				segmentPoint,
				a[i]*(1-float64(pointInserted)*segmentFraction)+b[i]*(float64(pointInserted)*segmentFraction),
			)
		}
		__antithesis_instrumentation__.Notify(62945)
		allSegmentizedCoordinates = append(
			allSegmentizedCoordinates,
			segmentPoint...,
		)
	}
	__antithesis_instrumentation__.Notify(62937)

	return allSegmentizedCoordinates, nil
}
