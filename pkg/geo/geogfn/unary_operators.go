package geogfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"

	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geographiclib"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
	"github.com/twpayne/go-geom"
)

func Area(g geo.Geography, useSphereOrSpheroid UseSphereOrSpheroid) (float64, error) {
	__antithesis_instrumentation__.Notify(59835)
	regions, err := g.AsS2(geo.EmptyBehaviorOmit)
	if err != nil {
		__antithesis_instrumentation__.Notify(59840)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(59841)
	}
	__antithesis_instrumentation__.Notify(59836)
	spheroid, err := g.Spheroid()
	if err != nil {
		__antithesis_instrumentation__.Notify(59842)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(59843)
	}
	__antithesis_instrumentation__.Notify(59837)

	var totalArea float64
	for _, region := range regions {
		__antithesis_instrumentation__.Notify(59844)
		switch region := region.(type) {
		case s2.Point, *s2.Polyline:
			__antithesis_instrumentation__.Notify(59845)
		case *s2.Polygon:
			__antithesis_instrumentation__.Notify(59846)
			if useSphereOrSpheroid == UseSpheroid {
				__antithesis_instrumentation__.Notify(59848)
				for _, loop := range region.Loops() {
					__antithesis_instrumentation__.Notify(59849)
					points := loop.Vertices()
					area, _ := spheroid.AreaAndPerimeter(points[:len(points)-1])
					totalArea += float64(loop.Sign()) * area
				}
			} else {
				__antithesis_instrumentation__.Notify(59850)
				totalArea += region.Area()
			}
		default:
			__antithesis_instrumentation__.Notify(59847)
			return 0, pgerror.Newf(pgcode.InvalidParameterValue, "unknown type: %T", region)
		}
	}
	__antithesis_instrumentation__.Notify(59838)
	if useSphereOrSpheroid == UseSphere {
		__antithesis_instrumentation__.Notify(59851)
		totalArea *= spheroid.SphereRadius * spheroid.SphereRadius
	} else {
		__antithesis_instrumentation__.Notify(59852)
	}
	__antithesis_instrumentation__.Notify(59839)
	return totalArea, nil
}

func Perimeter(g geo.Geography, useSphereOrSpheroid UseSphereOrSpheroid) (float64, error) {
	__antithesis_instrumentation__.Notify(59853)
	gt, err := g.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(59858)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(59859)
	}
	__antithesis_instrumentation__.Notify(59854)

	switch gt.(type) {
	case *geom.Polygon, *geom.MultiPolygon, *geom.GeometryCollection:
		__antithesis_instrumentation__.Notify(59860)
	default:
		__antithesis_instrumentation__.Notify(59861)
		return 0, nil
	}
	__antithesis_instrumentation__.Notify(59855)
	regions, err := geo.S2RegionsFromGeomT(gt, geo.EmptyBehaviorOmit)
	if err != nil {
		__antithesis_instrumentation__.Notify(59862)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(59863)
	}
	__antithesis_instrumentation__.Notify(59856)
	spheroid, err := g.Spheroid()
	if err != nil {
		__antithesis_instrumentation__.Notify(59864)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(59865)
	}
	__antithesis_instrumentation__.Notify(59857)
	return length(regions, spheroid, useSphereOrSpheroid)
}

func Length(g geo.Geography, useSphereOrSpheroid UseSphereOrSpheroid) (float64, error) {
	__antithesis_instrumentation__.Notify(59866)
	gt, err := g.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(59871)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(59872)
	}
	__antithesis_instrumentation__.Notify(59867)

	switch gt.(type) {
	case *geom.LineString, *geom.MultiLineString, *geom.GeometryCollection:
		__antithesis_instrumentation__.Notify(59873)
	default:
		__antithesis_instrumentation__.Notify(59874)
		return 0, nil
	}
	__antithesis_instrumentation__.Notify(59868)
	regions, err := geo.S2RegionsFromGeomT(gt, geo.EmptyBehaviorOmit)
	if err != nil {
		__antithesis_instrumentation__.Notify(59875)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(59876)
	}
	__antithesis_instrumentation__.Notify(59869)
	spheroid, err := g.Spheroid()
	if err != nil {
		__antithesis_instrumentation__.Notify(59877)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(59878)
	}
	__antithesis_instrumentation__.Notify(59870)
	return length(regions, spheroid, useSphereOrSpheroid)
}

func Project(g geo.Geography, distance float64, azimuth s1.Angle) (geo.Geography, error) {
	__antithesis_instrumentation__.Notify(59879)
	geomT, err := g.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(59886)
		return geo.Geography{}, err
	} else {
		__antithesis_instrumentation__.Notify(59887)
	}
	__antithesis_instrumentation__.Notify(59880)

	point, ok := geomT.(*geom.Point)
	if !ok {
		__antithesis_instrumentation__.Notify(59888)
		return geo.Geography{}, pgerror.Newf(pgcode.InvalidParameterValue, "ST_Project(geography) is only valid for point inputs")
	} else {
		__antithesis_instrumentation__.Notify(59889)
	}
	__antithesis_instrumentation__.Notify(59881)

	spheroid, err := g.Spheroid()
	if err != nil {
		__antithesis_instrumentation__.Notify(59890)
		return geo.Geography{}, err
	} else {
		__antithesis_instrumentation__.Notify(59891)
	}
	__antithesis_instrumentation__.Notify(59882)

	if distance < 0.0 {
		__antithesis_instrumentation__.Notify(59892)
		distance = -distance
		azimuth += math.Pi
	} else {
		__antithesis_instrumentation__.Notify(59893)
	}
	__antithesis_instrumentation__.Notify(59883)

	azimuth = azimuth.Normalized()

	if distance > (math.Pi * spheroid.Radius) {
		__antithesis_instrumentation__.Notify(59894)
		return geo.Geography{}, pgerror.Newf(pgcode.InvalidParameterValue, "distance must not be greater than %f", math.Pi*spheroid.Radius)
	} else {
		__antithesis_instrumentation__.Notify(59895)
	}
	__antithesis_instrumentation__.Notify(59884)

	if point.Empty() {
		__antithesis_instrumentation__.Notify(59896)
		return geo.Geography{}, pgerror.Newf(pgcode.InvalidParameterValue, "cannot project POINT EMPTY")
	} else {
		__antithesis_instrumentation__.Notify(59897)
	}
	__antithesis_instrumentation__.Notify(59885)

	x := point.X()
	y := point.Y()

	projected := spheroid.Project(
		s2.LatLngFromDegrees(x, y),
		distance,
		azimuth,
	)

	ret := geom.NewPointFlat(
		geom.XY,
		[]float64{
			geo.NormalizeLongitudeDegrees(projected.Lng.Degrees()),
			geo.NormalizeLatitudeDegrees(projected.Lat.Degrees()),
		},
	).SetSRID(point.SRID())
	return geo.MakeGeographyFromGeomT(ret)
}

func length(
	regions []s2.Region, spheroid *geographiclib.Spheroid, useSphereOrSpheroid UseSphereOrSpheroid,
) (float64, error) {
	__antithesis_instrumentation__.Notify(59898)
	var totalLength float64
	for _, region := range regions {
		__antithesis_instrumentation__.Notify(59901)
		switch region := region.(type) {
		case s2.Point:
			__antithesis_instrumentation__.Notify(59902)
		case *s2.Polyline:
			__antithesis_instrumentation__.Notify(59903)
			if useSphereOrSpheroid == UseSpheroid {
				__antithesis_instrumentation__.Notify(59906)
				totalLength += spheroid.InverseBatch((*region))
			} else {
				__antithesis_instrumentation__.Notify(59907)
				for edgeIdx, regionNumEdges := 0, region.NumEdges(); edgeIdx < regionNumEdges; edgeIdx++ {
					__antithesis_instrumentation__.Notify(59908)
					edge := region.Edge(edgeIdx)
					totalLength += s2.ChordAngleBetweenPoints(edge.V0, edge.V1).Angle().Radians()
				}
			}
		case *s2.Polygon:
			__antithesis_instrumentation__.Notify(59904)
			for _, loop := range region.Loops() {
				__antithesis_instrumentation__.Notify(59909)
				if useSphereOrSpheroid == UseSpheroid {
					__antithesis_instrumentation__.Notify(59910)
					totalLength += spheroid.InverseBatch(loop.Vertices())
				} else {
					__antithesis_instrumentation__.Notify(59911)
					for edgeIdx, loopNumEdges := 0, loop.NumEdges(); edgeIdx < loopNumEdges; edgeIdx++ {
						__antithesis_instrumentation__.Notify(59912)
						edge := loop.Edge(edgeIdx)
						totalLength += s2.ChordAngleBetweenPoints(edge.V0, edge.V1).Angle().Radians()
					}
				}
			}
		default:
			__antithesis_instrumentation__.Notify(59905)
			return 0, pgerror.Newf(pgcode.InvalidParameterValue, "unknown type: %T", region)
		}
	}
	__antithesis_instrumentation__.Notify(59899)
	if useSphereOrSpheroid == UseSphere {
		__antithesis_instrumentation__.Notify(59913)
		totalLength *= spheroid.SphereRadius
	} else {
		__antithesis_instrumentation__.Notify(59914)
	}
	__antithesis_instrumentation__.Notify(59900)
	return totalLength, nil
}
