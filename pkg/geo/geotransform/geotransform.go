package geotransform

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/geo/geoproj"
	"github.com/cockroachdb/cockroach/pkg/geo/geoprojbase"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/twpayne/go-geom"
)

func Transform(
	g geo.Geometry, from geoprojbase.Proj4Text, to geoprojbase.Proj4Text, newSRID geopb.SRID,
) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(64608)
	if from.Equal(to) {
		__antithesis_instrumentation__.Notify(64612)
		return g.CloneWithSRID(newSRID)
	} else {
		__antithesis_instrumentation__.Notify(64613)
	}
	__antithesis_instrumentation__.Notify(64609)

	t, err := g.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(64614)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(64615)
	}
	__antithesis_instrumentation__.Notify(64610)

	newT, err := transform(t, from, to, newSRID)
	if err != nil {
		__antithesis_instrumentation__.Notify(64616)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(64617)
	}
	__antithesis_instrumentation__.Notify(64611)
	return geo.MakeGeometryFromGeomT(newT)
}

func transform(
	t geom.T, from geoprojbase.Proj4Text, to geoprojbase.Proj4Text, newSRID geopb.SRID,
) (geom.T, error) {
	__antithesis_instrumentation__.Notify(64618)
	switch t := t.(type) {
	case *geom.Point:
		__antithesis_instrumentation__.Notify(64619)
		newCoords, err := projectFlatCoords(t.FlatCoords(), t.Layout(), from, to)
		if err != nil {
			__antithesis_instrumentation__.Notify(64634)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(64635)
		}
		__antithesis_instrumentation__.Notify(64620)
		return geom.NewPointFlat(t.Layout(), newCoords).SetSRID(int(newSRID)), nil
	case *geom.LineString:
		__antithesis_instrumentation__.Notify(64621)
		newCoords, err := projectFlatCoords(t.FlatCoords(), t.Layout(), from, to)
		if err != nil {
			__antithesis_instrumentation__.Notify(64636)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(64637)
		}
		__antithesis_instrumentation__.Notify(64622)
		return geom.NewLineStringFlat(t.Layout(), newCoords).SetSRID(int(newSRID)), nil
	case *geom.Polygon:
		__antithesis_instrumentation__.Notify(64623)
		newCoords, err := projectFlatCoords(t.FlatCoords(), t.Layout(), from, to)
		if err != nil {
			__antithesis_instrumentation__.Notify(64638)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(64639)
		}
		__antithesis_instrumentation__.Notify(64624)
		return geom.NewPolygonFlat(t.Layout(), newCoords, t.Ends()).SetSRID(int(newSRID)), nil
	case *geom.MultiPoint:
		__antithesis_instrumentation__.Notify(64625)
		newCoords, err := projectFlatCoords(t.FlatCoords(), t.Layout(), from, to)
		if err != nil {
			__antithesis_instrumentation__.Notify(64640)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(64641)
		}
		__antithesis_instrumentation__.Notify(64626)
		return geom.NewMultiPointFlat(t.Layout(), newCoords).SetSRID(int(newSRID)), nil
	case *geom.MultiLineString:
		__antithesis_instrumentation__.Notify(64627)
		newCoords, err := projectFlatCoords(t.FlatCoords(), t.Layout(), from, to)
		if err != nil {
			__antithesis_instrumentation__.Notify(64642)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(64643)
		}
		__antithesis_instrumentation__.Notify(64628)
		return geom.NewMultiLineStringFlat(t.Layout(), newCoords, t.Ends()).SetSRID(int(newSRID)), nil
	case *geom.MultiPolygon:
		__antithesis_instrumentation__.Notify(64629)
		newCoords, err := projectFlatCoords(t.FlatCoords(), t.Layout(), from, to)
		if err != nil {
			__antithesis_instrumentation__.Notify(64644)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(64645)
		}
		__antithesis_instrumentation__.Notify(64630)
		return geom.NewMultiPolygonFlat(t.Layout(), newCoords, t.Endss()).SetSRID(int(newSRID)), nil
	case *geom.GeometryCollection:
		__antithesis_instrumentation__.Notify(64631)
		g := geom.NewGeometryCollection().SetSRID(int(newSRID))
		for _, subG := range t.Geoms() {
			__antithesis_instrumentation__.Notify(64646)
			subGeom, err := transform(subG, from, to, 0)
			if err != nil {
				__antithesis_instrumentation__.Notify(64648)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(64649)
			}
			__antithesis_instrumentation__.Notify(64647)
			if err := g.Push(subGeom); err != nil {
				__antithesis_instrumentation__.Notify(64650)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(64651)
			}
		}
		__antithesis_instrumentation__.Notify(64632)
		return g, nil
	default:
		__antithesis_instrumentation__.Notify(64633)
		return nil, pgerror.Newf(pgcode.InvalidParameterValue, "unhandled type: %T", t)
	}
}

func projectFlatCoords(
	flatCoords []float64, layout geom.Layout, from geoprojbase.Proj4Text, to geoprojbase.Proj4Text,
) ([]float64, error) {
	__antithesis_instrumentation__.Notify(64652)
	numCoords := len(flatCoords) / layout.Stride()

	coords := make([]float64, numCoords*3)
	xCoords := coords[numCoords*0 : numCoords*1]
	yCoords := coords[numCoords*1 : numCoords*2]
	zCoords := coords[numCoords*2 : numCoords*3]
	for i := 0; i < numCoords; i++ {
		__antithesis_instrumentation__.Notify(64656)
		base := i * layout.Stride()
		xCoords[i] = flatCoords[base+0]
		yCoords[i] = flatCoords[base+1]

		if layout.ZIndex() != -1 {
			__antithesis_instrumentation__.Notify(64657)
			zCoords[i] = flatCoords[base+layout.ZIndex()]
		} else {
			__antithesis_instrumentation__.Notify(64658)
		}
	}
	__antithesis_instrumentation__.Notify(64653)

	if err := geoproj.Project(from, to, xCoords, yCoords, zCoords); err != nil {
		__antithesis_instrumentation__.Notify(64659)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64660)
	}
	__antithesis_instrumentation__.Notify(64654)
	newCoords := make([]float64, numCoords*layout.Stride())
	for i := 0; i < numCoords; i++ {
		__antithesis_instrumentation__.Notify(64661)
		base := i * layout.Stride()
		newCoords[base+0] = xCoords[i]
		newCoords[base+1] = yCoords[i]

		if layout.ZIndex() != -1 {
			__antithesis_instrumentation__.Notify(64663)
			newCoords[base+layout.ZIndex()] = zCoords[i]
		} else {
			__antithesis_instrumentation__.Notify(64664)
		}
		__antithesis_instrumentation__.Notify(64662)
		if layout.MIndex() != -1 {
			__antithesis_instrumentation__.Notify(64665)
			newCoords[base+layout.MIndex()] = flatCoords[i*layout.Stride()+layout.MIndex()]
		} else {
			__antithesis_instrumentation__.Notify(64666)
		}
	}
	__antithesis_instrumentation__.Notify(64655)

	return newCoords, nil
}
