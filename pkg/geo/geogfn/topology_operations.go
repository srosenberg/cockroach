package geogfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/golang/geo/r3"
	"github.com/golang/geo/s2"
	"github.com/twpayne/go-geom"
)

func Centroid(g geo.Geography, useSphereOrSpheroid UseSphereOrSpheroid) (geo.Geography, error) {
	__antithesis_instrumentation__.Notify(59803)
	geomRepr, err := g.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(59811)
		return geo.Geography{}, err
	} else {
		__antithesis_instrumentation__.Notify(59812)
	}
	__antithesis_instrumentation__.Notify(59804)
	if geomRepr.Empty() {
		__antithesis_instrumentation__.Notify(59813)
		return geo.MakeGeographyFromGeomT(geom.NewGeometryCollection().SetSRID(geomRepr.SRID()))
	} else {
		__antithesis_instrumentation__.Notify(59814)
	}
	__antithesis_instrumentation__.Notify(59805)
	switch geomRepr.(type) {
	case *geom.Point, *geom.LineString, *geom.Polygon, *geom.MultiPoint, *geom.MultiLineString, *geom.MultiPolygon:
		__antithesis_instrumentation__.Notify(59815)
	default:
		__antithesis_instrumentation__.Notify(59816)
		return geo.Geography{}, pgerror.Newf(pgcode.InvalidParameterValue, "unhandled geography type %s", g.ShapeType().String())
	}
	__antithesis_instrumentation__.Notify(59806)

	regions, err := geo.S2RegionsFromGeomT(geomRepr, geo.EmptyBehaviorOmit)
	if err != nil {
		__antithesis_instrumentation__.Notify(59817)
		return geo.Geography{}, err
	} else {
		__antithesis_instrumentation__.Notify(59818)
	}
	__antithesis_instrumentation__.Notify(59807)
	spheroid, err := g.Spheroid()
	if err != nil {
		__antithesis_instrumentation__.Notify(59819)
		return geo.Geography{}, err
	} else {
		__antithesis_instrumentation__.Notify(59820)
	}
	__antithesis_instrumentation__.Notify(59808)

	var localWeightedCentroids []s2.Point
	for _, region := range regions {
		__antithesis_instrumentation__.Notify(59821)
		switch region := region.(type) {
		case s2.Point:
			__antithesis_instrumentation__.Notify(59822)
			localWeightedCentroids = append(localWeightedCentroids, region)
		case *s2.Polyline:
			__antithesis_instrumentation__.Notify(59823)

			for edgeIdx, regionNumEdges := 0, region.NumEdges(); edgeIdx < regionNumEdges; edgeIdx++ {
				__antithesis_instrumentation__.Notify(59825)
				var edgeWeight float64
				eV0 := region.Edge(edgeIdx).V0
				eV1 := region.Edge(edgeIdx).V1
				if useSphereOrSpheroid == UseSpheroid {
					__antithesis_instrumentation__.Notify(59827)
					edgeWeight = spheroidDistance(spheroid, eV0, eV1)
				} else {
					__antithesis_instrumentation__.Notify(59828)
					edgeWeight = float64(s2.ChordAngleBetweenPoints(eV0, eV1).Angle())
				}
				__antithesis_instrumentation__.Notify(59826)
				localWeightedCentroids = append(localWeightedCentroids, s2.Point{Vector: eV0.Add(eV1.Vector).Mul(edgeWeight)})
			}
		case *s2.Polygon:
			__antithesis_instrumentation__.Notify(59824)

			for _, loop := range region.Loops() {
				__antithesis_instrumentation__.Notify(59829)
				triangleVertices := make([]s2.Point, 4)
				triangleVertices[0] = loop.Vertex(0)
				triangleVertices[3] = loop.Vertex(0)

				for pointIdx := 1; pointIdx+2 < loop.NumVertices(); pointIdx++ {
					__antithesis_instrumentation__.Notify(59830)
					triangleVertices[1] = loop.Vertex(pointIdx)
					triangleVertices[2] = loop.Vertex(pointIdx + 1)
					triangleCentroid := s2.PlanarCentroid(triangleVertices[0], triangleVertices[1], triangleVertices[2])
					var area float64
					if useSphereOrSpheroid == UseSpheroid {
						__antithesis_instrumentation__.Notify(59832)
						area, _ = spheroid.AreaAndPerimeter(triangleVertices[:3])
					} else {
						__antithesis_instrumentation__.Notify(59833)
						area = s2.LoopFromPoints(triangleVertices).Area()
					}
					__antithesis_instrumentation__.Notify(59831)
					area = area * float64(loop.Sign())
					localWeightedCentroids = append(localWeightedCentroids, s2.Point{Vector: triangleCentroid.Mul(area)})
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(59809)
	var centroidVector r3.Vector
	for _, point := range localWeightedCentroids {
		__antithesis_instrumentation__.Notify(59834)
		centroidVector = centroidVector.Add(point.Vector)
	}
	__antithesis_instrumentation__.Notify(59810)
	latLng := s2.LatLngFromPoint(s2.Point{Vector: centroidVector.Normalize()})
	centroid := geom.NewPointFlat(geom.XY, []float64{latLng.Lng.Degrees(), latLng.Lat.Degrees()}).SetSRID(int(g.SRID()))
	return geo.MakeGeographyFromGeomT(centroid)
}
