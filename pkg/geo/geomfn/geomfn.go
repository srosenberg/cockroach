// Package geomfn contains functions that are used for geometry-based builtins.
package geomfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/twpayne/go-geom"

type applyCoordFunc func(l geom.Layout, dst []float64, src []float64) error

func applyOnCoords(flatCoords []float64, l geom.Layout, f applyCoordFunc) ([]float64, error) {
	__antithesis_instrumentation__.Notify(62205)
	newCoords := make([]float64, len(flatCoords))
	for i := 0; i < len(flatCoords); i += l.Stride() {
		__antithesis_instrumentation__.Notify(62207)
		if err := f(l, newCoords[i:i+l.Stride()], flatCoords[i:i+l.Stride()]); err != nil {
			__antithesis_instrumentation__.Notify(62208)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(62209)
		}
	}
	__antithesis_instrumentation__.Notify(62206)
	return newCoords, nil
}

func applyOnCoordsForGeomT(g geom.T, f applyCoordFunc) (geom.T, error) {
	__antithesis_instrumentation__.Notify(62210)
	if geomCollection, ok := g.(*geom.GeometryCollection); ok {
		__antithesis_instrumentation__.Notify(62214)
		return applyOnCoordsForGeometryCollection(geomCollection, f)
	} else {
		__antithesis_instrumentation__.Notify(62215)
	}
	__antithesis_instrumentation__.Notify(62211)

	newCoords, err := applyOnCoords(g.FlatCoords(), g.Layout(), f)
	if err != nil {
		__antithesis_instrumentation__.Notify(62216)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(62217)
	}
	__antithesis_instrumentation__.Notify(62212)

	switch t := g.(type) {
	case *geom.Point:
		__antithesis_instrumentation__.Notify(62218)
		g = geom.NewPointFlat(t.Layout(), newCoords).SetSRID(g.SRID())
	case *geom.LineString:
		__antithesis_instrumentation__.Notify(62219)
		g = geom.NewLineStringFlat(t.Layout(), newCoords).SetSRID(g.SRID())
	case *geom.Polygon:
		__antithesis_instrumentation__.Notify(62220)
		g = geom.NewPolygonFlat(t.Layout(), newCoords, t.Ends()).SetSRID(g.SRID())
	case *geom.MultiPoint:
		__antithesis_instrumentation__.Notify(62221)
		g = geom.NewMultiPointFlat(t.Layout(), newCoords, geom.NewMultiPointFlatOptionWithEnds(t.Ends())).SetSRID(g.SRID())
	case *geom.MultiLineString:
		__antithesis_instrumentation__.Notify(62222)
		g = geom.NewMultiLineStringFlat(t.Layout(), newCoords, t.Ends()).SetSRID(g.SRID())
	case *geom.MultiPolygon:
		__antithesis_instrumentation__.Notify(62223)
		g = geom.NewMultiPolygonFlat(t.Layout(), newCoords, t.Endss()).SetSRID(g.SRID())
	default:
		__antithesis_instrumentation__.Notify(62224)
		return nil, geom.ErrUnsupportedType{Value: g}
	}
	__antithesis_instrumentation__.Notify(62213)

	return g, nil
}

func applyOnCoordsForGeometryCollection(
	geomCollection *geom.GeometryCollection, f applyCoordFunc,
) (*geom.GeometryCollection, error) {
	__antithesis_instrumentation__.Notify(62225)
	res := geom.NewGeometryCollection()
	for _, subG := range geomCollection.Geoms() {
		__antithesis_instrumentation__.Notify(62227)
		subGeom, err := applyOnCoordsForGeomT(subG, f)
		if err != nil {
			__antithesis_instrumentation__.Notify(62229)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(62230)
		}
		__antithesis_instrumentation__.Notify(62228)

		if err := res.Push(subGeom); err != nil {
			__antithesis_instrumentation__.Notify(62231)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(62232)
		}
	}
	__antithesis_instrumentation__.Notify(62226)
	return res, nil
}

func removeConsecutivePointsFromGeomT(t geom.T) (geom.T, error) {
	__antithesis_instrumentation__.Notify(62233)
	if t.Empty() {
		__antithesis_instrumentation__.Notify(62236)
		return t, nil
	} else {
		__antithesis_instrumentation__.Notify(62237)
	}
	__antithesis_instrumentation__.Notify(62234)
	switch t := t.(type) {
	case *geom.Point, *geom.MultiPoint:
		__antithesis_instrumentation__.Notify(62238)
		return t, nil
	case *geom.LineString:
		__antithesis_instrumentation__.Notify(62239)
		newCoords := make([]float64, 0, len(t.FlatCoords()))
		newCoords = append(newCoords, t.Coord(0)...)
		for i := 1; i < t.NumCoords(); i++ {
			__antithesis_instrumentation__.Notify(62250)
			if !t.Coord(i).Equal(t.Layout(), t.Coord(i-1)) {
				__antithesis_instrumentation__.Notify(62251)
				newCoords = append(newCoords, t.Coord(i)...)
			} else {
				__antithesis_instrumentation__.Notify(62252)
			}
		}
		__antithesis_instrumentation__.Notify(62240)
		if len(newCoords) < t.Stride()*2 {
			__antithesis_instrumentation__.Notify(62253)
			newCoords = newCoords[:0]
		} else {
			__antithesis_instrumentation__.Notify(62254)
		}
		__antithesis_instrumentation__.Notify(62241)
		return geom.NewLineStringFlat(t.Layout(), newCoords).SetSRID(t.SRID()), nil
	case *geom.Polygon:
		__antithesis_instrumentation__.Notify(62242)
		ret := geom.NewPolygon(t.Layout()).SetSRID(t.SRID())
		for ringIdx := 0; ringIdx < t.NumLinearRings(); ringIdx++ {
			__antithesis_instrumentation__.Notify(62255)
			ring := t.LinearRing(ringIdx)
			newCoords := make([]float64, 0, len(ring.FlatCoords()))
			newCoords = append(newCoords, ring.Coord(0)...)
			for i := 1; i < ring.NumCoords(); i++ {
				__antithesis_instrumentation__.Notify(62258)
				if !ring.Coord(i).Equal(ring.Layout(), ring.Coord(i-1)) {
					__antithesis_instrumentation__.Notify(62259)
					newCoords = append(newCoords, ring.Coord(i)...)
				} else {
					__antithesis_instrumentation__.Notify(62260)
				}
			}
			__antithesis_instrumentation__.Notify(62256)
			if len(newCoords) < t.Stride()*4 {
				__antithesis_instrumentation__.Notify(62261)

				if ringIdx == 0 {
					__antithesis_instrumentation__.Notify(62263)
					return ret, nil
				} else {
					__antithesis_instrumentation__.Notify(62264)
				}
				__antithesis_instrumentation__.Notify(62262)

				continue
			} else {
				__antithesis_instrumentation__.Notify(62265)
			}
			__antithesis_instrumentation__.Notify(62257)
			if err := ret.Push(geom.NewLinearRingFlat(t.Layout(), newCoords)); err != nil {
				__antithesis_instrumentation__.Notify(62266)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(62267)
			}
		}
		__antithesis_instrumentation__.Notify(62243)
		return ret, nil
	case *geom.MultiLineString:
		__antithesis_instrumentation__.Notify(62244)
		ret := geom.NewMultiLineString(t.Layout()).SetSRID(t.SRID())
		for i := 0; i < t.NumLineStrings(); i++ {
			__antithesis_instrumentation__.Notify(62268)
			ls, err := removeConsecutivePointsFromGeomT(t.LineString(i))
			if err != nil {
				__antithesis_instrumentation__.Notify(62271)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(62272)
			}
			__antithesis_instrumentation__.Notify(62269)
			if ls.Empty() {
				__antithesis_instrumentation__.Notify(62273)
				continue
			} else {
				__antithesis_instrumentation__.Notify(62274)
			}
			__antithesis_instrumentation__.Notify(62270)
			if err := ret.Push(ls.(*geom.LineString)); err != nil {
				__antithesis_instrumentation__.Notify(62275)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(62276)
			}
		}
		__antithesis_instrumentation__.Notify(62245)
		return ret, nil
	case *geom.MultiPolygon:
		__antithesis_instrumentation__.Notify(62246)
		ret := geom.NewMultiPolygon(t.Layout()).SetSRID(t.SRID())
		for i := 0; i < t.NumPolygons(); i++ {
			__antithesis_instrumentation__.Notify(62277)
			p, err := removeConsecutivePointsFromGeomT(t.Polygon(i))
			if err != nil {
				__antithesis_instrumentation__.Notify(62280)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(62281)
			}
			__antithesis_instrumentation__.Notify(62278)
			if p.Empty() {
				__antithesis_instrumentation__.Notify(62282)
				continue
			} else {
				__antithesis_instrumentation__.Notify(62283)
			}
			__antithesis_instrumentation__.Notify(62279)
			if err := ret.Push(p.(*geom.Polygon)); err != nil {
				__antithesis_instrumentation__.Notify(62284)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(62285)
			}
		}
		__antithesis_instrumentation__.Notify(62247)
		return ret, nil
	case *geom.GeometryCollection:
		__antithesis_instrumentation__.Notify(62248)
		ret := geom.NewGeometryCollection().SetSRID(t.SRID())
		for i := 0; i < t.NumGeoms(); i++ {
			__antithesis_instrumentation__.Notify(62286)
			g, err := removeConsecutivePointsFromGeomT(t.Geom(i))
			if err != nil {
				__antithesis_instrumentation__.Notify(62289)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(62290)
			}
			__antithesis_instrumentation__.Notify(62287)
			if g.Empty() {
				__antithesis_instrumentation__.Notify(62291)
				continue
			} else {
				__antithesis_instrumentation__.Notify(62292)
			}
			__antithesis_instrumentation__.Notify(62288)
			if err := ret.Push(g); err != nil {
				__antithesis_instrumentation__.Notify(62293)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(62294)
			}
		}
		__antithesis_instrumentation__.Notify(62249)
		return ret, nil
	}
	__antithesis_instrumentation__.Notify(62235)
	return nil, geom.ErrUnsupportedType{Value: t}
}
