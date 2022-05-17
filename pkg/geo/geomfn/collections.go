package geomfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/errors"
	"github.com/twpayne/go-geom"
)

func Collect(g1 geo.Geometry, g2 geo.Geometry) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(61441)
	t1, err := g1.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(61447)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(61448)
	}
	__antithesis_instrumentation__.Notify(61442)
	t2, err := g2.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(61449)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(61450)
	}
	__antithesis_instrumentation__.Notify(61443)

	switch t1 := t1.(type) {
	case *geom.Point:
		__antithesis_instrumentation__.Notify(61451)
		if t2, ok := t2.(*geom.Point); ok {
			__antithesis_instrumentation__.Notify(61454)
			multi := geom.NewMultiPoint(t1.Layout()).SetSRID(t1.SRID())
			if err := multi.Push(t1); err != nil {
				__antithesis_instrumentation__.Notify(61457)
				return geo.Geometry{}, err
			} else {
				__antithesis_instrumentation__.Notify(61458)
			}
			__antithesis_instrumentation__.Notify(61455)
			if err := multi.Push(t2); err != nil {
				__antithesis_instrumentation__.Notify(61459)
				return geo.Geometry{}, err
			} else {
				__antithesis_instrumentation__.Notify(61460)
			}
			__antithesis_instrumentation__.Notify(61456)
			return geo.MakeGeometryFromGeomT(multi)
		} else {
			__antithesis_instrumentation__.Notify(61461)
		}
	case *geom.LineString:
		__antithesis_instrumentation__.Notify(61452)
		if t2, ok := t2.(*geom.LineString); ok {
			__antithesis_instrumentation__.Notify(61462)
			multi := geom.NewMultiLineString(t1.Layout()).SetSRID(t1.SRID())
			if err := multi.Push(t1); err != nil {
				__antithesis_instrumentation__.Notify(61465)
				return geo.Geometry{}, err
			} else {
				__antithesis_instrumentation__.Notify(61466)
			}
			__antithesis_instrumentation__.Notify(61463)
			if err := multi.Push(t2); err != nil {
				__antithesis_instrumentation__.Notify(61467)
				return geo.Geometry{}, err
			} else {
				__antithesis_instrumentation__.Notify(61468)
			}
			__antithesis_instrumentation__.Notify(61464)
			return geo.MakeGeometryFromGeomT(multi)
		} else {
			__antithesis_instrumentation__.Notify(61469)
		}
	case *geom.Polygon:
		__antithesis_instrumentation__.Notify(61453)
		if t2, ok := t2.(*geom.Polygon); ok {
			__antithesis_instrumentation__.Notify(61470)
			multi := geom.NewMultiPolygon(t1.Layout()).SetSRID(t1.SRID())
			if err := multi.Push(t1); err != nil {
				__antithesis_instrumentation__.Notify(61473)
				return geo.Geometry{}, err
			} else {
				__antithesis_instrumentation__.Notify(61474)
			}
			__antithesis_instrumentation__.Notify(61471)
			if err := multi.Push(t2); err != nil {
				__antithesis_instrumentation__.Notify(61475)
				return geo.Geometry{}, err
			} else {
				__antithesis_instrumentation__.Notify(61476)
			}
			__antithesis_instrumentation__.Notify(61472)
			return geo.MakeGeometryFromGeomT(multi)
		} else {
			__antithesis_instrumentation__.Notify(61477)
		}
	}
	__antithesis_instrumentation__.Notify(61444)

	gc := geom.NewGeometryCollection().SetSRID(t1.SRID())
	if err := gc.Push(t1); err != nil {
		__antithesis_instrumentation__.Notify(61478)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(61479)
	}
	__antithesis_instrumentation__.Notify(61445)
	if err := gc.Push(t2); err != nil {
		__antithesis_instrumentation__.Notify(61480)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(61481)
	}
	__antithesis_instrumentation__.Notify(61446)
	return geo.MakeGeometryFromGeomT(gc)
}

func CollectionExtract(g geo.Geometry, shapeType geopb.ShapeType) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(61482)
	switch shapeType {
	case geopb.ShapeType_Point, geopb.ShapeType_LineString, geopb.ShapeType_Polygon:
		__antithesis_instrumentation__.Notify(61486)
	default:
		__antithesis_instrumentation__.Notify(61487)
		return geo.Geometry{}, pgerror.Newf(
			pgcode.InvalidParameterValue,
			"only point, linestring and polygon may be extracted (got %s)",
			shapeType,
		)
	}
	__antithesis_instrumentation__.Notify(61483)

	if g.ShapeType() == shapeType || func() bool {
		__antithesis_instrumentation__.Notify(61488)
		return g.ShapeType() == shapeType.MultiType() == true
	}() == true {
		__antithesis_instrumentation__.Notify(61489)
		return g, nil
	} else {
		__antithesis_instrumentation__.Notify(61490)
	}
	__antithesis_instrumentation__.Notify(61484)

	t, err := g.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(61491)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(61492)
	}
	__antithesis_instrumentation__.Notify(61485)

	switch t := t.(type) {

	case *geom.Point, *geom.LineString, *geom.Polygon:
		__antithesis_instrumentation__.Notify(61493)
		switch shapeType {
		case geopb.ShapeType_Point:
			__antithesis_instrumentation__.Notify(61500)
			return geo.MakeGeometryFromGeomT(geom.NewPointEmpty(t.Layout()).SetSRID(t.SRID()))
		case geopb.ShapeType_LineString:
			__antithesis_instrumentation__.Notify(61501)
			return geo.MakeGeometryFromGeomT(geom.NewLineString(t.Layout()).SetSRID(t.SRID()))
		case geopb.ShapeType_Polygon:
			__antithesis_instrumentation__.Notify(61502)
			return geo.MakeGeometryFromGeomT(geom.NewPolygon(t.Layout()).SetSRID(t.SRID()))
		default:
			__antithesis_instrumentation__.Notify(61503)
			return geo.Geometry{}, pgerror.Newf(
				pgcode.InvalidParameterValue,
				"unexpected shape type %v",
				shapeType.String(),
			)
		}

	case *geom.MultiPoint, *geom.MultiLineString, *geom.MultiPolygon:
		__antithesis_instrumentation__.Notify(61494)
		switch shapeType.MultiType() {
		case geopb.ShapeType_MultiPoint:
			__antithesis_instrumentation__.Notify(61504)
			return geo.MakeGeometryFromGeomT(geom.NewMultiPoint(t.Layout()).SetSRID(t.SRID()))
		case geopb.ShapeType_MultiLineString:
			__antithesis_instrumentation__.Notify(61505)
			return geo.MakeGeometryFromGeomT(geom.NewMultiLineString(t.Layout()).SetSRID(t.SRID()))
		case geopb.ShapeType_MultiPolygon:
			__antithesis_instrumentation__.Notify(61506)
			return geo.MakeGeometryFromGeomT(geom.NewMultiPolygon(t.Layout()).SetSRID(t.SRID()))
		default:
			__antithesis_instrumentation__.Notify(61507)
			return geo.Geometry{}, pgerror.Newf(
				pgcode.InvalidParameterValue,
				"unexpected shape type %v",
				shapeType.MultiType().String(),
			)
		}

	case *geom.GeometryCollection:
		__antithesis_instrumentation__.Notify(61495)

		layout := t.Layout()
		if layout == geom.NoLayout && func() bool {
			__antithesis_instrumentation__.Notify(61508)
			return t.Empty() == true
		}() == true {
			__antithesis_instrumentation__.Notify(61509)
			layout = geom.XY
		} else {
			__antithesis_instrumentation__.Notify(61510)
		}
		__antithesis_instrumentation__.Notify(61496)
		iter := geo.NewGeomTIterator(t, geo.EmptyBehaviorOmit)
		srid := t.SRID()

		var (
			multi geom.T
			err   error
		)
		switch shapeType {
		case geopb.ShapeType_Point:
			__antithesis_instrumentation__.Notify(61511)
			multi, err = collectionExtractPoints(iter, layout, srid)
		case geopb.ShapeType_LineString:
			__antithesis_instrumentation__.Notify(61512)
			multi, err = collectionExtractLineStrings(iter, layout, srid)
		case geopb.ShapeType_Polygon:
			__antithesis_instrumentation__.Notify(61513)
			multi, err = collectionExtractPolygons(iter, layout, srid)
		default:
			__antithesis_instrumentation__.Notify(61514)
			return geo.Geometry{}, errors.AssertionFailedf("unexpected shape type %v", shapeType.String())
		}
		__antithesis_instrumentation__.Notify(61497)
		if err != nil {
			__antithesis_instrumentation__.Notify(61515)
			return geo.Geometry{}, err
		} else {
			__antithesis_instrumentation__.Notify(61516)
		}
		__antithesis_instrumentation__.Notify(61498)
		return geo.MakeGeometryFromGeomT(multi)

	default:
		__antithesis_instrumentation__.Notify(61499)
		return geo.Geometry{}, errors.AssertionFailedf("unexpected shape type: %T", t)
	}
}

func collectionExtractPoints(
	iter geo.GeomTIterator, layout geom.Layout, srid int,
) (*geom.MultiPoint, error) {
	__antithesis_instrumentation__.Notify(61517)
	points := geom.NewMultiPoint(layout).SetSRID(srid)
	for {
		__antithesis_instrumentation__.Notify(61519)
		if next, hasNext, err := iter.Next(); err != nil {
			__antithesis_instrumentation__.Notify(61520)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(61521)
			if !hasNext {
				__antithesis_instrumentation__.Notify(61522)
				break
			} else {
				__antithesis_instrumentation__.Notify(61523)
				if point, ok := next.(*geom.Point); ok {
					__antithesis_instrumentation__.Notify(61524)
					if err = points.Push(point); err != nil {
						__antithesis_instrumentation__.Notify(61525)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(61526)
					}
				} else {
					__antithesis_instrumentation__.Notify(61527)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(61518)
	return points, nil
}

func collectionExtractLineStrings(
	iter geo.GeomTIterator, layout geom.Layout, srid int,
) (*geom.MultiLineString, error) {
	__antithesis_instrumentation__.Notify(61528)
	lineStrings := geom.NewMultiLineString(layout).SetSRID(srid)
	for {
		__antithesis_instrumentation__.Notify(61530)
		if next, hasNext, err := iter.Next(); err != nil {
			__antithesis_instrumentation__.Notify(61531)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(61532)
			if !hasNext {
				__antithesis_instrumentation__.Notify(61533)
				break
			} else {
				__antithesis_instrumentation__.Notify(61534)
				if lineString, ok := next.(*geom.LineString); ok {
					__antithesis_instrumentation__.Notify(61535)
					if err = lineStrings.Push(lineString); err != nil {
						__antithesis_instrumentation__.Notify(61536)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(61537)
					}
				} else {
					__antithesis_instrumentation__.Notify(61538)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(61529)
	return lineStrings, nil
}

func collectionExtractPolygons(
	iter geo.GeomTIterator, layout geom.Layout, srid int,
) (*geom.MultiPolygon, error) {
	__antithesis_instrumentation__.Notify(61539)
	polygons := geom.NewMultiPolygon(layout).SetSRID(srid)
	for {
		__antithesis_instrumentation__.Notify(61541)
		if next, hasNext, err := iter.Next(); err != nil {
			__antithesis_instrumentation__.Notify(61542)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(61543)
			if !hasNext {
				__antithesis_instrumentation__.Notify(61544)
				break
			} else {
				__antithesis_instrumentation__.Notify(61545)
				if polygon, ok := next.(*geom.Polygon); ok {
					__antithesis_instrumentation__.Notify(61546)
					if err = polygons.Push(polygon); err != nil {
						__antithesis_instrumentation__.Notify(61547)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(61548)
					}
				} else {
					__antithesis_instrumentation__.Notify(61549)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(61540)
	return polygons, nil
}

func CollectionHomogenize(g geo.Geometry) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(61550)
	t, err := g.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(61554)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(61555)
	}
	__antithesis_instrumentation__.Notify(61551)
	srid := t.SRID()
	t, err = collectionHomogenizeGeomT(t)
	if err != nil {
		__antithesis_instrumentation__.Notify(61556)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(61557)
	}
	__antithesis_instrumentation__.Notify(61552)
	if srid != 0 {
		__antithesis_instrumentation__.Notify(61558)
		geo.AdjustGeomTSRID(t, geopb.SRID(srid))
	} else {
		__antithesis_instrumentation__.Notify(61559)
	}
	__antithesis_instrumentation__.Notify(61553)
	return geo.MakeGeometryFromGeomT(t)
}

func collectionHomogenizeGeomT(t geom.T) (geom.T, error) {
	__antithesis_instrumentation__.Notify(61560)
	switch t := t.(type) {
	case *geom.Point, *geom.LineString, *geom.Polygon:
		__antithesis_instrumentation__.Notify(61561)
		return t, nil

	case *geom.MultiPoint:
		__antithesis_instrumentation__.Notify(61562)
		if t.NumPoints() == 1 {
			__antithesis_instrumentation__.Notify(61576)
			return t.Point(0), nil
		} else {
			__antithesis_instrumentation__.Notify(61577)
		}
		__antithesis_instrumentation__.Notify(61563)
		return t, nil

	case *geom.MultiLineString:
		__antithesis_instrumentation__.Notify(61564)
		if t.NumLineStrings() == 1 {
			__antithesis_instrumentation__.Notify(61578)
			return t.LineString(0), nil
		} else {
			__antithesis_instrumentation__.Notify(61579)
		}
		__antithesis_instrumentation__.Notify(61565)
		return t, nil

	case *geom.MultiPolygon:
		__antithesis_instrumentation__.Notify(61566)
		if t.NumPolygons() == 1 {
			__antithesis_instrumentation__.Notify(61580)
			return t.Polygon(0), nil
		} else {
			__antithesis_instrumentation__.Notify(61581)
		}
		__antithesis_instrumentation__.Notify(61567)
		return t, nil

	case *geom.GeometryCollection:
		__antithesis_instrumentation__.Notify(61568)
		layout := t.Layout()
		if layout == geom.NoLayout && func() bool {
			__antithesis_instrumentation__.Notify(61582)
			return t.Empty() == true
		}() == true {
			__antithesis_instrumentation__.Notify(61583)
			layout = geom.XY
		} else {
			__antithesis_instrumentation__.Notify(61584)
		}
		__antithesis_instrumentation__.Notify(61569)
		points := geom.NewMultiPoint(layout)
		linestrings := geom.NewMultiLineString(layout)
		polygons := geom.NewMultiPolygon(layout)
		iter := geo.NewGeomTIterator(t, geo.EmptyBehaviorOmit)
		for {
			__antithesis_instrumentation__.Notify(61585)
			next, hasNext, err := iter.Next()
			if err != nil {
				__antithesis_instrumentation__.Notify(61589)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(61590)
			}
			__antithesis_instrumentation__.Notify(61586)
			if !hasNext {
				__antithesis_instrumentation__.Notify(61591)
				break
			} else {
				__antithesis_instrumentation__.Notify(61592)
			}
			__antithesis_instrumentation__.Notify(61587)
			switch next := next.(type) {
			case *geom.Point:
				__antithesis_instrumentation__.Notify(61593)
				err = points.Push(next)
			case *geom.LineString:
				__antithesis_instrumentation__.Notify(61594)
				err = linestrings.Push(next)
			case *geom.Polygon:
				__antithesis_instrumentation__.Notify(61595)
				err = polygons.Push(next)
			default:
				__antithesis_instrumentation__.Notify(61596)
				err = errors.AssertionFailedf("encountered unexpected geometry type: %T", next)
			}
			__antithesis_instrumentation__.Notify(61588)
			if err != nil {
				__antithesis_instrumentation__.Notify(61597)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(61598)
			}
		}
		__antithesis_instrumentation__.Notify(61570)
		homog := geom.NewGeometryCollection()
		switch points.NumPoints() {
		case 0:
			__antithesis_instrumentation__.Notify(61599)
		case 1:
			__antithesis_instrumentation__.Notify(61600)
			if err := homog.Push(points.Point(0)); err != nil {
				__antithesis_instrumentation__.Notify(61602)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(61603)
			}
		default:
			__antithesis_instrumentation__.Notify(61601)
			if err := homog.Push(points); err != nil {
				__antithesis_instrumentation__.Notify(61604)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(61605)
			}
		}
		__antithesis_instrumentation__.Notify(61571)
		switch linestrings.NumLineStrings() {
		case 0:
			__antithesis_instrumentation__.Notify(61606)
		case 1:
			__antithesis_instrumentation__.Notify(61607)
			if err := homog.Push(linestrings.LineString(0)); err != nil {
				__antithesis_instrumentation__.Notify(61609)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(61610)
			}
		default:
			__antithesis_instrumentation__.Notify(61608)
			if err := homog.Push(linestrings); err != nil {
				__antithesis_instrumentation__.Notify(61611)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(61612)
			}
		}
		__antithesis_instrumentation__.Notify(61572)
		switch polygons.NumPolygons() {
		case 0:
			__antithesis_instrumentation__.Notify(61613)
		case 1:
			__antithesis_instrumentation__.Notify(61614)
			if err := homog.Push(polygons.Polygon(0)); err != nil {
				__antithesis_instrumentation__.Notify(61616)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(61617)
			}
		default:
			__antithesis_instrumentation__.Notify(61615)
			if err := homog.Push(polygons); err != nil {
				__antithesis_instrumentation__.Notify(61618)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(61619)
			}
		}
		__antithesis_instrumentation__.Notify(61573)

		if homog.NumGeoms() == 1 {
			__antithesis_instrumentation__.Notify(61620)
			return homog.Geom(0), nil
		} else {
			__antithesis_instrumentation__.Notify(61621)
		}
		__antithesis_instrumentation__.Notify(61574)
		return homog, nil

	default:
		__antithesis_instrumentation__.Notify(61575)
		return nil, errors.AssertionFailedf("unknown geometry type: %T", t)
	}
}

func ForceCollection(g geo.Geometry) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(61622)
	t, err := g.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(61625)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(61626)
	}
	__antithesis_instrumentation__.Notify(61623)
	t, err = forceCollectionFromGeomT(t)
	if err != nil {
		__antithesis_instrumentation__.Notify(61627)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(61628)
	}
	__antithesis_instrumentation__.Notify(61624)
	return geo.MakeGeometryFromGeomT(t)
}

func forceCollectionFromGeomT(t geom.T) (geom.T, error) {
	__antithesis_instrumentation__.Notify(61629)
	gc := geom.NewGeometryCollection().SetSRID(t.SRID())
	switch t := t.(type) {
	case *geom.Point, *geom.LineString, *geom.Polygon:
		__antithesis_instrumentation__.Notify(61631)
		if err := gc.Push(t); err != nil {
			__antithesis_instrumentation__.Notify(61637)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(61638)
		}
	case *geom.MultiPoint:
		__antithesis_instrumentation__.Notify(61632)
		for i := 0; i < t.NumPoints(); i++ {
			__antithesis_instrumentation__.Notify(61639)
			if err := gc.Push(t.Point(i)); err != nil {
				__antithesis_instrumentation__.Notify(61640)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(61641)
			}
		}
	case *geom.MultiLineString:
		__antithesis_instrumentation__.Notify(61633)
		for i := 0; i < t.NumLineStrings(); i++ {
			__antithesis_instrumentation__.Notify(61642)
			if err := gc.Push(t.LineString(i)); err != nil {
				__antithesis_instrumentation__.Notify(61643)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(61644)
			}
		}
	case *geom.MultiPolygon:
		__antithesis_instrumentation__.Notify(61634)
		for i := 0; i < t.NumPolygons(); i++ {
			__antithesis_instrumentation__.Notify(61645)
			if err := gc.Push(t.Polygon(i)); err != nil {
				__antithesis_instrumentation__.Notify(61646)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(61647)
			}
		}
	case *geom.GeometryCollection:
		__antithesis_instrumentation__.Notify(61635)
		gc = t
	default:
		__antithesis_instrumentation__.Notify(61636)
		return nil, errors.AssertionFailedf("unknown geometry type: %T", t)
	}
	__antithesis_instrumentation__.Notify(61630)
	return gc, nil
}

func Multi(g geo.Geometry) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(61648)
	t, err := g.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(61650)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(61651)
	}
	__antithesis_instrumentation__.Notify(61649)
	switch t := t.(type) {
	case *geom.MultiPoint, *geom.MultiLineString, *geom.MultiPolygon, *geom.GeometryCollection:
		__antithesis_instrumentation__.Notify(61652)
		return geo.MakeGeometryFromGeomT(t)
	case *geom.Point:
		__antithesis_instrumentation__.Notify(61653)
		multi := geom.NewMultiPoint(t.Layout()).SetSRID(t.SRID())
		if !t.Empty() {
			__antithesis_instrumentation__.Notify(61660)
			if err = multi.Push(t); err != nil {
				__antithesis_instrumentation__.Notify(61661)
				return geo.Geometry{}, err
			} else {
				__antithesis_instrumentation__.Notify(61662)
			}
		} else {
			__antithesis_instrumentation__.Notify(61663)
		}
		__antithesis_instrumentation__.Notify(61654)
		return geo.MakeGeometryFromGeomT(multi)
	case *geom.LineString:
		__antithesis_instrumentation__.Notify(61655)
		multi := geom.NewMultiLineString(t.Layout()).SetSRID(t.SRID())
		if !t.Empty() {
			__antithesis_instrumentation__.Notify(61664)
			if err = multi.Push(t); err != nil {
				__antithesis_instrumentation__.Notify(61665)
				return geo.Geometry{}, err
			} else {
				__antithesis_instrumentation__.Notify(61666)
			}
		} else {
			__antithesis_instrumentation__.Notify(61667)
		}
		__antithesis_instrumentation__.Notify(61656)
		return geo.MakeGeometryFromGeomT(multi)
	case *geom.Polygon:
		__antithesis_instrumentation__.Notify(61657)
		multi := geom.NewMultiPolygon(t.Layout()).SetSRID(t.SRID())
		if !t.Empty() {
			__antithesis_instrumentation__.Notify(61668)
			if err = multi.Push(t); err != nil {
				__antithesis_instrumentation__.Notify(61669)
				return geo.Geometry{}, err
			} else {
				__antithesis_instrumentation__.Notify(61670)
			}
		} else {
			__antithesis_instrumentation__.Notify(61671)
		}
		__antithesis_instrumentation__.Notify(61658)
		return geo.MakeGeometryFromGeomT(multi)
	default:
		__antithesis_instrumentation__.Notify(61659)
		return geo.Geometry{}, errors.AssertionFailedf("unknown geometry type: %T", t)
	}
}
