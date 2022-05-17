package geomfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/twpayne/go-geom"
)

func Reverse(geometry geo.Geometry) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(62882)
	g, err := geometry.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(62885)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(62886)
	}
	__antithesis_instrumentation__.Notify(62883)

	g, err = reverse(g)
	if err != nil {
		__antithesis_instrumentation__.Notify(62887)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(62888)
	}
	__antithesis_instrumentation__.Notify(62884)

	return geo.MakeGeometryFromGeomT(g)
}

func reverse(g geom.T) (geom.T, error) {
	__antithesis_instrumentation__.Notify(62889)
	if geomCollection, ok := g.(*geom.GeometryCollection); ok {
		__antithesis_instrumentation__.Notify(62892)
		return reverseCollection(geomCollection)
	} else {
		__antithesis_instrumentation__.Notify(62893)
	}
	__antithesis_instrumentation__.Notify(62890)

	switch t := g.(type) {
	case *geom.Point, *geom.MultiPoint:
		__antithesis_instrumentation__.Notify(62894)
		return g, nil
	case *geom.LineString:
		__antithesis_instrumentation__.Notify(62895)
		g = geom.NewLineStringFlat(t.Layout(), reverseCoords(g.FlatCoords(), g.Stride())).SetSRID(g.SRID())
	case *geom.Polygon:
		__antithesis_instrumentation__.Notify(62896)
		g = geom.NewPolygonFlat(t.Layout(), reverseCoords(g.FlatCoords(), g.Stride()), t.Ends()).SetSRID(g.SRID())
	case *geom.MultiLineString:
		__antithesis_instrumentation__.Notify(62897)
		g = geom.NewMultiLineStringFlat(t.Layout(), reverseMulti(g, t.Ends()), t.Ends()).SetSRID(g.SRID())
	case *geom.MultiPolygon:
		__antithesis_instrumentation__.Notify(62898)
		var ends []int
		for _, e := range t.Endss() {
			__antithesis_instrumentation__.Notify(62901)
			ends = append(ends, e...)
		}
		__antithesis_instrumentation__.Notify(62899)
		g = geom.NewMultiPolygonFlat(t.Layout(), reverseMulti(g, ends), t.Endss()).SetSRID(g.SRID())

	default:
		__antithesis_instrumentation__.Notify(62900)
		return nil, geom.ErrUnsupportedType{Value: g}
	}
	__antithesis_instrumentation__.Notify(62891)

	return g, nil
}

func reverseCoords(coords []float64, stride int) []float64 {
	__antithesis_instrumentation__.Notify(62902)
	for i := 0; i < len(coords)/2; i += stride {
		__antithesis_instrumentation__.Notify(62904)
		for j := 0; j < stride; j++ {
			__antithesis_instrumentation__.Notify(62905)
			coords[i+j], coords[len(coords)-stride-i+j] = coords[len(coords)-stride-i+j], coords[i+j]
		}
	}
	__antithesis_instrumentation__.Notify(62903)

	return coords
}

func reverseMulti(g geom.T, ends []int) []float64 {
	__antithesis_instrumentation__.Notify(62906)
	coords := g.FlatCoords()
	prevEnd := 0

	for _, end := range ends {
		__antithesis_instrumentation__.Notify(62908)
		copy(
			coords[prevEnd:end],
			reverseCoords(coords[prevEnd:end], g.Stride()),
		)
		prevEnd = end
	}
	__antithesis_instrumentation__.Notify(62907)

	return coords
}

func reverseCollection(geomCollection *geom.GeometryCollection) (*geom.GeometryCollection, error) {
	__antithesis_instrumentation__.Notify(62909)
	res := geom.NewGeometryCollection()
	for _, subG := range geomCollection.Geoms() {
		__antithesis_instrumentation__.Notify(62911)
		subGeom, err := reverse(subG)
		if err != nil {
			__antithesis_instrumentation__.Notify(62913)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(62914)
		}
		__antithesis_instrumentation__.Notify(62912)

		if err := res.Push(subGeom); err != nil {
			__antithesis_instrumentation__.Notify(62915)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(62916)
		}
	}
	__antithesis_instrumentation__.Notify(62910)
	return res, nil
}
