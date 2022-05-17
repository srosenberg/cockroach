package geomfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/twpayne/go-geom"
)

func ForceLayout(g geo.Geometry, layout geom.Layout) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(62061)
	return ForceLayoutWithDefaultZM(g, layout, 0, 0)
}

func ForceLayoutWithDefaultZ(
	g geo.Geometry, layout geom.Layout, defaultZ float64,
) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(62062)
	return ForceLayoutWithDefaultZM(g, layout, defaultZ, 0)
}

func ForceLayoutWithDefaultM(
	g geo.Geometry, layout geom.Layout, defaultM float64,
) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(62063)
	return ForceLayoutWithDefaultZM(g, layout, 0, defaultM)
}

func ForceLayoutWithDefaultZM(
	g geo.Geometry, layout geom.Layout, defaultZ float64, defaultM float64,
) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(62064)
	geomT, err := g.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(62067)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(62068)
	}
	__antithesis_instrumentation__.Notify(62065)
	retGeomT, err := forceLayout(geomT, layout, defaultZ, defaultM)
	if err != nil {
		__antithesis_instrumentation__.Notify(62069)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(62070)
	}
	__antithesis_instrumentation__.Notify(62066)
	return geo.MakeGeometryFromGeomT(retGeomT)
}

func forceLayout(t geom.T, layout geom.Layout, defaultZ float64, defaultM float64) (geom.T, error) {
	__antithesis_instrumentation__.Notify(62071)
	if t.Layout() == layout {
		__antithesis_instrumentation__.Notify(62073)
		return t, nil
	} else {
		__antithesis_instrumentation__.Notify(62074)
	}
	__antithesis_instrumentation__.Notify(62072)
	switch t := t.(type) {
	case *geom.GeometryCollection:
		__antithesis_instrumentation__.Notify(62075)
		ret := geom.NewGeometryCollection().SetSRID(t.SRID())
		if err := ret.SetLayout(layout); err != nil {
			__antithesis_instrumentation__.Notify(62086)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(62087)
		}
		__antithesis_instrumentation__.Notify(62076)
		for i := 0; i < t.NumGeoms(); i++ {
			__antithesis_instrumentation__.Notify(62088)
			toPush, err := forceLayout(t.Geom(i), layout, defaultZ, defaultM)
			if err != nil {
				__antithesis_instrumentation__.Notify(62090)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(62091)
			}
			__antithesis_instrumentation__.Notify(62089)
			if err := ret.Push(toPush); err != nil {
				__antithesis_instrumentation__.Notify(62092)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(62093)
			}
		}
		__antithesis_instrumentation__.Notify(62077)
		return ret, nil
	case *geom.Point:
		__antithesis_instrumentation__.Notify(62078)
		return geom.NewPointFlat(
			layout, forceFlatCoordsLayout(t, layout, defaultZ, defaultM)).SetSRID(t.SRID()), nil
	case *geom.LineString:
		__antithesis_instrumentation__.Notify(62079)
		return geom.NewLineStringFlat(
			layout, forceFlatCoordsLayout(t, layout, defaultZ, defaultM)).SetSRID(t.SRID()), nil
	case *geom.Polygon:
		__antithesis_instrumentation__.Notify(62080)
		return geom.NewPolygonFlat(
			layout,
			forceFlatCoordsLayout(t, layout, defaultZ, defaultM),
			forceEnds(t.Ends(), t.Layout(), layout),
		).SetSRID(t.SRID()), nil
	case *geom.MultiPoint:
		__antithesis_instrumentation__.Notify(62081)
		return geom.NewMultiPointFlat(
			layout,
			forceFlatCoordsLayout(t, layout, defaultZ, defaultM),
			geom.NewMultiPointFlatOptionWithEnds(forceEnds(t.Ends(), t.Layout(), layout)),
		).SetSRID(t.SRID()), nil
	case *geom.MultiLineString:
		__antithesis_instrumentation__.Notify(62082)
		return geom.NewMultiLineStringFlat(
			layout,
			forceFlatCoordsLayout(t, layout, defaultZ, defaultM),
			forceEnds(t.Ends(), t.Layout(), layout),
		).SetSRID(t.SRID()), nil
	case *geom.MultiPolygon:
		__antithesis_instrumentation__.Notify(62083)
		endss := make([][]int, len(t.Endss()))
		for i := range t.Endss() {
			__antithesis_instrumentation__.Notify(62094)
			endss[i] = forceEnds(t.Endss()[i], t.Layout(), layout)
		}
		__antithesis_instrumentation__.Notify(62084)
		return geom.NewMultiPolygonFlat(
			layout,
			forceFlatCoordsLayout(t, layout, defaultZ, defaultM),
			endss,
		).SetSRID(t.SRID()), nil
	default:
		__antithesis_instrumentation__.Notify(62085)
		return nil, pgerror.Newf(pgcode.InvalidParameterValue, "unknown geom.T type: %T", t)
	}
}

func forceEnds(ends []int, oldLayout geom.Layout, newLayout geom.Layout) []int {
	__antithesis_instrumentation__.Notify(62095)
	if oldLayout.Stride() == newLayout.Stride() {
		__antithesis_instrumentation__.Notify(62098)
		return ends
	} else {
		__antithesis_instrumentation__.Notify(62099)
	}
	__antithesis_instrumentation__.Notify(62096)
	newEnds := make([]int, len(ends))
	for i := range ends {
		__antithesis_instrumentation__.Notify(62100)
		newEnds[i] = (ends[i] / oldLayout.Stride()) * newLayout.Stride()
	}
	__antithesis_instrumentation__.Notify(62097)
	return newEnds
}

func forceFlatCoordsLayout(
	t geom.T, layout geom.Layout, defaultZ float64, defaultM float64,
) []float64 {
	__antithesis_instrumentation__.Notify(62101)
	oldFlatCoords := t.FlatCoords()
	newFlatCoords := make([]float64, (len(oldFlatCoords)/t.Stride())*layout.Stride())
	for coordIdx := 0; coordIdx < len(oldFlatCoords)/t.Stride(); coordIdx++ {
		__antithesis_instrumentation__.Notify(62103)
		newFlatCoords[coordIdx*layout.Stride()] = oldFlatCoords[coordIdx*t.Stride()]
		newFlatCoords[coordIdx*layout.Stride()+1] = oldFlatCoords[coordIdx*t.Stride()+1]
		if layout.ZIndex() != -1 {
			__antithesis_instrumentation__.Notify(62105)
			z := defaultZ
			if t.Layout().ZIndex() != -1 {
				__antithesis_instrumentation__.Notify(62107)
				z = oldFlatCoords[coordIdx*t.Stride()+t.Layout().ZIndex()]
			} else {
				__antithesis_instrumentation__.Notify(62108)
			}
			__antithesis_instrumentation__.Notify(62106)
			newFlatCoords[coordIdx*layout.Stride()+layout.ZIndex()] = z
		} else {
			__antithesis_instrumentation__.Notify(62109)
		}
		__antithesis_instrumentation__.Notify(62104)
		if layout.MIndex() != -1 {
			__antithesis_instrumentation__.Notify(62110)
			m := defaultM
			if t.Layout().MIndex() != -1 {
				__antithesis_instrumentation__.Notify(62112)
				m = oldFlatCoords[coordIdx*t.Stride()+t.Layout().MIndex()]
			} else {
				__antithesis_instrumentation__.Notify(62113)
			}
			__antithesis_instrumentation__.Notify(62111)
			newFlatCoords[coordIdx*layout.Stride()+layout.MIndex()] = m
		} else {
			__antithesis_instrumentation__.Notify(62114)
		}
	}
	__antithesis_instrumentation__.Notify(62102)
	return newFlatCoords
}
