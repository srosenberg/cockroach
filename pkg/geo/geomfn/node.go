package geomfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/geo/geos"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/errors"
	geom "github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/xy"
)

func Node(g geo.Geometry) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(62593)
	if g.ShapeType() != geopb.ShapeType_LineString && func() bool {
		__antithesis_instrumentation__.Notify(62607)
		return g.ShapeType() != geopb.ShapeType_MultiLineString == true
	}() == true {
		__antithesis_instrumentation__.Notify(62608)
		return geo.Geometry{}, pgerror.Newf(
			pgcode.InvalidParameterValue,
			"geometry type is unsupported. Please pass a LineString or a MultiLineString",
		)
	} else {
		__antithesis_instrumentation__.Notify(62609)
	}
	__antithesis_instrumentation__.Notify(62594)

	if g.Empty() {
		__antithesis_instrumentation__.Notify(62610)
		return geo.MakeGeometryFromGeomT(geom.NewGeometryCollection().SetSRID(int(g.SRID())))
	} else {
		__antithesis_instrumentation__.Notify(62611)
	}
	__antithesis_instrumentation__.Notify(62595)

	res, err := geos.Node(g.EWKB())
	if err != nil {
		__antithesis_instrumentation__.Notify(62612)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(62613)
	}
	__antithesis_instrumentation__.Notify(62596)
	node, err := geo.ParseGeometryFromEWKB(res)
	if err != nil {
		__antithesis_instrumentation__.Notify(62614)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(62615)
	}
	__antithesis_instrumentation__.Notify(62597)

	res, err = geos.LineMerge(node.EWKB())
	if err != nil {
		__antithesis_instrumentation__.Notify(62616)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(62617)
	}
	__antithesis_instrumentation__.Notify(62598)
	lines, err := geo.ParseGeometryFromEWKB(res)
	if err != nil {
		__antithesis_instrumentation__.Notify(62618)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(62619)
	}
	__antithesis_instrumentation__.Notify(62599)
	if lines.ShapeType() == geopb.ShapeType_LineString {
		__antithesis_instrumentation__.Notify(62620)

		return node, nil
	} else {
		__antithesis_instrumentation__.Notify(62621)
	}
	__antithesis_instrumentation__.Notify(62600)

	glines, err := lines.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(62622)
		return geo.Geometry{}, errors.Wrap(err, "error transforming lines")
	} else {
		__antithesis_instrumentation__.Notify(62623)
	}
	__antithesis_instrumentation__.Notify(62601)
	ep, err := extractEndpoints(g)
	if err != nil {
		__antithesis_instrumentation__.Notify(62624)
		return geo.Geometry{}, errors.Wrap(err, "error extracting endpoints")
	} else {
		__antithesis_instrumentation__.Notify(62625)
	}
	__antithesis_instrumentation__.Notify(62602)
	var mllines *geom.MultiLineString
	switch t := glines.(type) {
	case *geom.MultiLineString:
		__antithesis_instrumentation__.Notify(62626)
		mllines = t
	case *geom.GeometryCollection:
		__antithesis_instrumentation__.Notify(62627)
		if t.Empty() {
			__antithesis_instrumentation__.Notify(62630)
			t.SetSRID(int(g.SRID()))
			return geo.MakeGeometryFromGeomT(t)
		} else {
			__antithesis_instrumentation__.Notify(62631)
		}
		__antithesis_instrumentation__.Notify(62628)
		return geo.Geometry{}, errors.AssertionFailedf("unknown GEOMETRYCOLLECTION: %T", t)
	default:
		__antithesis_instrumentation__.Notify(62629)
		return geo.Geometry{}, errors.AssertionFailedf("unknown LineMerge result type: %T", t)
	}
	__antithesis_instrumentation__.Notify(62603)

	gep, err := ep.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(62632)
		return geo.Geometry{}, errors.Wrap(err, "error transforming endpoints")
	} else {
		__antithesis_instrumentation__.Notify(62633)
	}
	__antithesis_instrumentation__.Notify(62604)
	mpep := gep.(*geom.MultiPoint)
	mlout, err := splitLinesByPoints(mllines, mpep)
	if err != nil {
		__antithesis_instrumentation__.Notify(62634)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(62635)
	}
	__antithesis_instrumentation__.Notify(62605)
	mlout.SetSRID(int(g.SRID()))
	out, err := geo.MakeGeometryFromGeomT(mlout)
	if err != nil {
		__antithesis_instrumentation__.Notify(62636)
		return geo.Geometry{}, errors.Wrap(err, "could not transform output into geometry")
	} else {
		__antithesis_instrumentation__.Notify(62637)
	}
	__antithesis_instrumentation__.Notify(62606)
	return out, nil
}

func splitLinesByPoints(
	mllines *geom.MultiLineString, mpep *geom.MultiPoint,
) (*geom.MultiLineString, error) {
	__antithesis_instrumentation__.Notify(62638)
	mlout := geom.NewMultiLineString(geom.XY)
	splitted := false
	var err error
	var splitLines []*geom.LineString
	for i := 0; i < mllines.NumLineStrings(); i++ {
		__antithesis_instrumentation__.Notify(62640)
		l := mllines.LineString(i)
		for j := 0; j < mpep.NumPoints(); j++ {
			__antithesis_instrumentation__.Notify(62642)
			p := mpep.Point(j)
			splitted, splitLines, err = splitLineByPoint(l, p.Coords())
			if err != nil {
				__antithesis_instrumentation__.Notify(62644)
				return nil, errors.Wrap(err, "could not split line")
			} else {
				__antithesis_instrumentation__.Notify(62645)
			}
			__antithesis_instrumentation__.Notify(62643)
			if splitted {
				__antithesis_instrumentation__.Notify(62646)
				err = mlout.Push(splitLines[0])
				if err != nil {
					__antithesis_instrumentation__.Notify(62649)
					return nil, errors.Wrap(err, "could not construct output geometry")
				} else {
					__antithesis_instrumentation__.Notify(62650)
				}
				__antithesis_instrumentation__.Notify(62647)
				err = mlout.Push(splitLines[1])
				if err != nil {
					__antithesis_instrumentation__.Notify(62651)
					return nil, errors.Wrap(err, "could not construct output geometry")
				} else {
					__antithesis_instrumentation__.Notify(62652)
				}
				__antithesis_instrumentation__.Notify(62648)
				break
			} else {
				__antithesis_instrumentation__.Notify(62653)
			}
		}
		__antithesis_instrumentation__.Notify(62641)
		if !splitted {
			__antithesis_instrumentation__.Notify(62654)
			err = mlout.Push(l)
			if err != nil {
				__antithesis_instrumentation__.Notify(62655)
				return nil, errors.Wrap(err, "could not construct output geometry")
			} else {
				__antithesis_instrumentation__.Notify(62656)
			}
		} else {
			__antithesis_instrumentation__.Notify(62657)
		}
	}
	__antithesis_instrumentation__.Notify(62639)
	return mlout, nil
}

func extractEndpoints(g geo.Geometry) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(62658)
	mp := geom.NewMultiPoint(geom.XY)

	gt, err := g.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(62662)
		return geo.Geometry{}, errors.Wrap(err, "error transforming geometry")
	} else {
		__antithesis_instrumentation__.Notify(62663)
	}
	__antithesis_instrumentation__.Notify(62659)

	switch gt := gt.(type) {
	case *geom.LineString:
		__antithesis_instrumentation__.Notify(62664)
		endpoints := collectEndpoints(gt)
		for _, endpoint := range endpoints {
			__antithesis_instrumentation__.Notify(62667)
			err := mp.Push(endpoint)
			if err != nil {
				__antithesis_instrumentation__.Notify(62668)
				return geo.Geometry{}, errors.Wrap(err, "error creating output geometry")
			} else {
				__antithesis_instrumentation__.Notify(62669)
			}
		}
	case *geom.MultiLineString:
		__antithesis_instrumentation__.Notify(62665)
		for i := 0; i < gt.NumLineStrings(); i++ {
			__antithesis_instrumentation__.Notify(62670)
			ls := gt.LineString(i)
			endpoints := collectEndpoints(ls)
			for _, endpoint := range endpoints {
				__antithesis_instrumentation__.Notify(62671)
				err := mp.Push(endpoint)
				if err != nil {
					__antithesis_instrumentation__.Notify(62672)
					return geo.Geometry{}, errors.Wrap(err, "error creating output geometry")
				} else {
					__antithesis_instrumentation__.Notify(62673)
				}
			}
		}
	default:
		__antithesis_instrumentation__.Notify(62666)
		return geo.Geometry{}, pgerror.Newf(pgcode.InvalidParameterValue, "unsupported type: %T", gt)
	}
	__antithesis_instrumentation__.Notify(62660)

	result, err := geo.MakeGeometryFromGeomT(mp)
	if err != nil {
		__antithesis_instrumentation__.Notify(62674)
		return geo.Geometry{}, errors.Wrap(err, "error creating output geometry")
	} else {
		__antithesis_instrumentation__.Notify(62675)
	}
	__antithesis_instrumentation__.Notify(62661)
	return result, nil
}

func collectEndpoints(ls *geom.LineString) []*geom.Point {
	__antithesis_instrumentation__.Notify(62676)
	coord := ls.Coord(0)
	startPoint := geom.NewPointFlat(geom.XY, []float64{coord.X(), coord.Y()})
	coord = ls.Coord(ls.NumCoords() - 1)
	endPoint := geom.NewPointFlat(geom.XY, []float64{coord.X(), coord.Y()})
	return []*geom.Point{startPoint, endPoint}
}

func splitLineByPoint(l *geom.LineString, p geom.Coord) (bool, []*geom.LineString, error) {
	__antithesis_instrumentation__.Notify(62677)

	if !xy.IsOnLine(l.Layout(), p, l.FlatCoords()) {
		__antithesis_instrumentation__.Notify(62683)
		return false, []*geom.LineString{l}, nil
	} else {
		__antithesis_instrumentation__.Notify(62684)
	}
	__antithesis_instrumentation__.Notify(62678)

	startCoord := l.Coord(0)
	endCoord := l.Coord(l.NumCoords() - 1)
	if p.Equal(l.Layout(), startCoord) || func() bool {
		__antithesis_instrumentation__.Notify(62685)
		return p.Equal(l.Layout(), endCoord) == true
	}() == true {
		__antithesis_instrumentation__.Notify(62686)
		return false, []*geom.LineString{l}, nil
	} else {
		__antithesis_instrumentation__.Notify(62687)
	}
	__antithesis_instrumentation__.Notify(62679)

	coordsA := []geom.Coord{}
	coordsB := []geom.Coord{}
	for i := 1; i < l.NumCoords(); i++ {
		__antithesis_instrumentation__.Notify(62688)
		if xy.IsPointWithinLineBounds(p, l.Coord(i-1), l.Coord(i)) {
			__antithesis_instrumentation__.Notify(62689)
			coordsA = append(l.Coords()[0:i], p)
			if p.Equal(l.Layout(), l.Coord(i)) {
				__antithesis_instrumentation__.Notify(62691)
				coordsB = l.Coords()[i:]
			} else {
				__antithesis_instrumentation__.Notify(62692)
				coordsB = append([]geom.Coord{p}, l.Coords()[i:]...)
			}
			__antithesis_instrumentation__.Notify(62690)
			break
		} else {
			__antithesis_instrumentation__.Notify(62693)
		}
	}
	__antithesis_instrumentation__.Notify(62680)
	l1 := geom.NewLineString(l.Layout())
	_, err := l1.SetCoords(coordsA)
	if err != nil {
		__antithesis_instrumentation__.Notify(62694)
		return false, []*geom.LineString{}, errors.Wrap(err, "could not set coords")
	} else {
		__antithesis_instrumentation__.Notify(62695)
	}
	__antithesis_instrumentation__.Notify(62681)
	l2 := geom.NewLineString(l.Layout())
	_, err = l2.SetCoords(coordsB)
	if err != nil {
		__antithesis_instrumentation__.Notify(62696)
		return false, []*geom.LineString{}, errors.Wrap(err, "could not set coords")
	} else {
		__antithesis_instrumentation__.Notify(62697)
	}
	__antithesis_instrumentation__.Notify(62682)
	return true, []*geom.LineString{l1, l2}, nil
}
