package geogfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"

	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geosegmentize"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/golang/geo/s2"
	"github.com/twpayne/go-geom"
)

func Segmentize(geography geo.Geography, segmentMaxLength float64) (geo.Geography, error) {
	__antithesis_instrumentation__.Notify(59771)
	if math.IsNaN(segmentMaxLength) || func() bool {
		__antithesis_instrumentation__.Notify(59774)
		return math.IsInf(segmentMaxLength, 1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(59775)
		return geography, nil
	} else {
		__antithesis_instrumentation__.Notify(59776)
	}
	__antithesis_instrumentation__.Notify(59772)
	geometry, err := geography.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(59777)
		return geo.Geography{}, err
	} else {
		__antithesis_instrumentation__.Notify(59778)
	}
	__antithesis_instrumentation__.Notify(59773)
	switch geometry := geometry.(type) {
	case *geom.Point, *geom.MultiPoint:
		__antithesis_instrumentation__.Notify(59779)
		return geography, nil
	default:
		__antithesis_instrumentation__.Notify(59780)
		if segmentMaxLength <= 0 {
			__antithesis_instrumentation__.Notify(59784)
			return geo.Geography{}, pgerror.Newf(pgcode.InvalidParameterValue, "maximum segment length must be positive")
		} else {
			__antithesis_instrumentation__.Notify(59785)
		}
		__antithesis_instrumentation__.Notify(59781)
		spheroid, err := geography.Spheroid()
		if err != nil {
			__antithesis_instrumentation__.Notify(59786)
			return geo.Geography{}, err
		} else {
			__antithesis_instrumentation__.Notify(59787)
		}
		__antithesis_instrumentation__.Notify(59782)

		segmentMaxAngle := segmentMaxLength / spheroid.SphereRadius
		ret, err := geosegmentize.Segmentize(geometry, segmentMaxAngle, segmentizeCoords)
		if err != nil {
			__antithesis_instrumentation__.Notify(59788)
			return geo.Geography{}, err
		} else {
			__antithesis_instrumentation__.Notify(59789)
		}
		__antithesis_instrumentation__.Notify(59783)
		return geo.MakeGeographyFromGeomT(ret)
	}
}

func segmentizeCoords(a geom.Coord, b geom.Coord, segmentMaxAngle float64) ([]float64, error) {
	__antithesis_instrumentation__.Notify(59790)
	if len(a) != len(b) {
		__antithesis_instrumentation__.Notify(59795)
		return nil, pgerror.Newf(pgcode.InvalidParameterValue, "cannot segmentize two coordinates of different dimensions")
	} else {
		__antithesis_instrumentation__.Notify(59796)
	}
	__antithesis_instrumentation__.Notify(59791)
	if segmentMaxAngle <= 0 {
		__antithesis_instrumentation__.Notify(59797)
		return nil, pgerror.Newf(pgcode.InvalidParameterValue, "maximum segment angle must be positive")
	} else {
		__antithesis_instrumentation__.Notify(59798)
	}
	__antithesis_instrumentation__.Notify(59792)

	pointA := s2.PointFromLatLng(s2.LatLngFromDegrees(a.Y(), a.X()))
	pointB := s2.PointFromLatLng(s2.LatLngFromDegrees(b.Y(), b.X()))

	chordAngleBetweenPoints := s2.ChordAngleBetweenPoints(pointA, pointB).Angle().Radians()

	doubleNumberOfSegmentsToCreate := math.Pow(2, math.Ceil(math.Log2(chordAngleBetweenPoints/segmentMaxAngle)))
	doubleNumPoints := float64(len(a)) * (1 + doubleNumberOfSegmentsToCreate)
	if err := geosegmentize.CheckSegmentizeValidNumPoints(doubleNumPoints, a, b); err != nil {
		__antithesis_instrumentation__.Notify(59799)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(59800)
	}
	__antithesis_instrumentation__.Notify(59793)
	numberOfSegmentsToCreate := int(doubleNumberOfSegmentsToCreate)
	numPoints := int(doubleNumPoints)

	allSegmentizedCoordinates := make([]float64, 0, numPoints)
	allSegmentizedCoordinates = append(allSegmentizedCoordinates, a.Clone()...)
	segmentFraction := 1.0 / float64(numberOfSegmentsToCreate)
	for pointInserted := 1; pointInserted < numberOfSegmentsToCreate; pointInserted++ {
		__antithesis_instrumentation__.Notify(59801)
		newPoint := s2.Interpolate(float64(pointInserted)/float64(numberOfSegmentsToCreate), pointA, pointB)
		latLng := s2.LatLngFromPoint(newPoint)
		allSegmentizedCoordinates = append(allSegmentizedCoordinates, latLng.Lng.Degrees(), latLng.Lat.Degrees())

		for i := 2; i < len(a); i++ {
			__antithesis_instrumentation__.Notify(59802)
			allSegmentizedCoordinates = append(
				allSegmentizedCoordinates,
				a[i]*(1-float64(pointInserted)*segmentFraction)+b[i]*(float64(pointInserted)*segmentFraction),
			)
		}
	}
	__antithesis_instrumentation__.Notify(59794)
	return allSegmentizedCoordinates, nil
}
