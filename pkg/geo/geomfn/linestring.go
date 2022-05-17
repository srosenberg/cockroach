package geomfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geos"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/errors"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/ewkb"
)

func LineStringFromMultiPoint(g geo.Geometry) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(62380)
	t, err := g.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(62385)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(62386)
	}
	__antithesis_instrumentation__.Notify(62381)
	mp, ok := t.(*geom.MultiPoint)
	if !ok {
		__antithesis_instrumentation__.Notify(62387)
		return geo.Geometry{}, errors.Wrap(geom.ErrUnsupportedType{Value: t},
			"geometry must be a MultiPoint")
	} else {
		__antithesis_instrumentation__.Notify(62388)
	}
	__antithesis_instrumentation__.Notify(62382)
	if mp.NumPoints() == 1 {
		__antithesis_instrumentation__.Notify(62389)
		return geo.Geometry{}, pgerror.Newf(pgcode.InvalidParameterValue, "a LineString must have at least 2 points")
	} else {
		__antithesis_instrumentation__.Notify(62390)
	}
	__antithesis_instrumentation__.Notify(62383)
	flatCoords := make([]float64, 0, mp.NumCoords()*mp.Stride())
	var prevPoint *geom.Point
	for i := 0; i < mp.NumPoints(); i++ {
		__antithesis_instrumentation__.Notify(62391)
		p := mp.Point(i)

		if p.Empty() {
			__antithesis_instrumentation__.Notify(62393)
			if prevPoint == nil {
				__antithesis_instrumentation__.Notify(62395)
				prevPoint = geom.NewPointFlat(mp.Layout(), make([]float64, mp.Stride()))
			} else {
				__antithesis_instrumentation__.Notify(62396)
			}
			__antithesis_instrumentation__.Notify(62394)
			flatCoords = append(flatCoords, prevPoint.FlatCoords()...)
			continue
		} else {
			__antithesis_instrumentation__.Notify(62397)
		}
		__antithesis_instrumentation__.Notify(62392)
		flatCoords = append(flatCoords, p.FlatCoords()...)
		prevPoint = p
	}
	__antithesis_instrumentation__.Notify(62384)
	lineString := geom.NewLineStringFlat(mp.Layout(), flatCoords).SetSRID(mp.SRID())
	return geo.MakeGeometryFromGeomT(lineString)
}

func LineMerge(g geo.Geometry) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(62398)

	if g.Empty() {
		__antithesis_instrumentation__.Notify(62402)
		return g, nil
	} else {
		__antithesis_instrumentation__.Notify(62403)
	}
	__antithesis_instrumentation__.Notify(62399)
	if BoundingBoxHasInfiniteCoordinates(g) {
		__antithesis_instrumentation__.Notify(62404)
		return geo.Geometry{}, pgerror.Newf(pgcode.InvalidParameterValue, "value out of range: overflow")
	} else {
		__antithesis_instrumentation__.Notify(62405)
	}
	__antithesis_instrumentation__.Notify(62400)
	ret, err := geos.LineMerge(g.EWKB())
	if err != nil {
		__antithesis_instrumentation__.Notify(62406)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(62407)
	}
	__antithesis_instrumentation__.Notify(62401)
	return geo.ParseGeometryFromEWKB(ret)
}

func LineLocatePoint(line geo.Geometry, point geo.Geometry) (float64, error) {
	__antithesis_instrumentation__.Notify(62408)
	lineT, err := line.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(62413)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(62414)
	}
	__antithesis_instrumentation__.Notify(62409)
	lineString, ok := lineT.(*geom.LineString)
	if !ok {
		__antithesis_instrumentation__.Notify(62415)
		return 0, pgerror.Newf(pgcode.InvalidParameterValue,
			"first parameter has to be of type LineString")
	} else {
		__antithesis_instrumentation__.Notify(62416)
	}
	__antithesis_instrumentation__.Notify(62410)

	closestPoint, err := ClosestPoint(line, point)
	if err != nil {
		__antithesis_instrumentation__.Notify(62417)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(62418)
	}
	__antithesis_instrumentation__.Notify(62411)
	closestT, err := closestPoint.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(62419)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(62420)
	}
	__antithesis_instrumentation__.Notify(62412)

	p := closestT.(*geom.Point)
	lineStart := geom.Coord{lineString.Coord(0).X(), lineString.Coord(0).Y()}

	lineSegment := geom.NewLineString(geom.XY).MustSetCoords([]geom.Coord{lineStart, p.Coords()})

	return lineSegment.Length() / lineString.Length(), nil
}

func AddPoint(lineString geo.Geometry, index int, point geo.Geometry) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(62421)
	g, err := lineString.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(62427)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(62428)
	}
	__antithesis_instrumentation__.Notify(62422)
	lineStringG, ok := g.(*geom.LineString)
	if !ok {
		__antithesis_instrumentation__.Notify(62429)
		e := geom.ErrUnsupportedType{Value: g}
		return geo.Geometry{}, errors.Wrap(e, "geometry to be modified must be a LineString")
	} else {
		__antithesis_instrumentation__.Notify(62430)
	}
	__antithesis_instrumentation__.Notify(62423)

	g, err = point.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(62431)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(62432)
	}
	__antithesis_instrumentation__.Notify(62424)

	pointG, ok := g.(*geom.Point)
	if !ok {
		__antithesis_instrumentation__.Notify(62433)
		e := geom.ErrUnsupportedType{Value: g}
		return geo.Geometry{}, errors.Wrapf(e, "invalid geometry used to add a Point to a LineString")
	} else {
		__antithesis_instrumentation__.Notify(62434)
	}
	__antithesis_instrumentation__.Notify(62425)

	g, err = addPoint(lineStringG, index, pointG)
	if err != nil {
		__antithesis_instrumentation__.Notify(62435)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(62436)
	}
	__antithesis_instrumentation__.Notify(62426)

	return geo.MakeGeometryFromGeomT(g)
}

func addPoint(lineString *geom.LineString, index int, point *geom.Point) (*geom.LineString, error) {
	__antithesis_instrumentation__.Notify(62437)
	if lineString.Layout() != point.Layout() {
		__antithesis_instrumentation__.Notify(62441)
		return nil, pgerror.WithCandidateCode(
			geom.ErrLayoutMismatch{Got: point.Layout(), Want: lineString.Layout()},
			pgcode.InvalidParameterValue,
		)
	} else {
		__antithesis_instrumentation__.Notify(62442)
	}
	__antithesis_instrumentation__.Notify(62438)
	if point.Empty() {
		__antithesis_instrumentation__.Notify(62443)
		point = geom.NewPointFlat(point.Layout(), make([]float64, point.Stride()))
	} else {
		__antithesis_instrumentation__.Notify(62444)
	}
	__antithesis_instrumentation__.Notify(62439)

	coords := lineString.Coords()

	if index > len(coords) {
		__antithesis_instrumentation__.Notify(62445)
		return nil, pgerror.Newf(pgcode.InvalidParameterValue, "index %d out of range of LineString with %d coordinates",
			index, len(coords))
	} else {
		__antithesis_instrumentation__.Notify(62446)
		if index == -1 {
			__antithesis_instrumentation__.Notify(62447)
			index = len(coords)
		} else {
			__antithesis_instrumentation__.Notify(62448)
			if index < 0 {
				__antithesis_instrumentation__.Notify(62449)
				return nil, pgerror.Newf(pgcode.InvalidParameterValue, "invalid index %v", index)
			} else {
				__antithesis_instrumentation__.Notify(62450)
			}
		}
	}
	__antithesis_instrumentation__.Notify(62440)

	coords = append(coords, geom.Coord{})
	copy(coords[index+1:], coords[index:])
	coords[index] = point.Coords()

	return lineString.SetCoords(coords)
}

func SetPoint(lineString geo.Geometry, index int, point geo.Geometry) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(62451)
	g, err := lineString.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(62457)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(62458)
	}
	__antithesis_instrumentation__.Notify(62452)

	lineStringG, ok := g.(*geom.LineString)
	if !ok {
		__antithesis_instrumentation__.Notify(62459)
		e := geom.ErrUnsupportedType{Value: g}
		return geo.Geometry{}, errors.Wrap(e, "geometry to be modified must be a LineString")
	} else {
		__antithesis_instrumentation__.Notify(62460)
	}
	__antithesis_instrumentation__.Notify(62453)

	g, err = point.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(62461)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(62462)
	}
	__antithesis_instrumentation__.Notify(62454)

	pointG, ok := g.(*geom.Point)
	if !ok {
		__antithesis_instrumentation__.Notify(62463)
		e := geom.ErrUnsupportedType{Value: g}
		return geo.Geometry{}, errors.Wrapf(e, "invalid geometry used to replace a Point on a LineString")
	} else {
		__antithesis_instrumentation__.Notify(62464)
	}
	__antithesis_instrumentation__.Notify(62455)

	g, err = setPoint(lineStringG, index, pointG)
	if err != nil {
		__antithesis_instrumentation__.Notify(62465)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(62466)
	}
	__antithesis_instrumentation__.Notify(62456)

	return geo.MakeGeometryFromGeomT(g)
}

func setPoint(lineString *geom.LineString, index int, point *geom.Point) (*geom.LineString, error) {
	__antithesis_instrumentation__.Notify(62467)
	if lineString.Layout() != point.Layout() {
		__antithesis_instrumentation__.Notify(62472)
		return nil, geom.ErrLayoutMismatch{Got: point.Layout(), Want: lineString.Layout()}
	} else {
		__antithesis_instrumentation__.Notify(62473)
	}
	__antithesis_instrumentation__.Notify(62468)
	if point.Empty() {
		__antithesis_instrumentation__.Notify(62474)
		point = geom.NewPointFlat(point.Layout(), make([]float64, point.Stride()))
	} else {
		__antithesis_instrumentation__.Notify(62475)
	}
	__antithesis_instrumentation__.Notify(62469)

	coords := lineString.Coords()
	hasNegIndex := index < 0

	if index >= len(coords) || func() bool {
		__antithesis_instrumentation__.Notify(62476)
		return (hasNegIndex && func() bool {
			__antithesis_instrumentation__.Notify(62477)
			return index*-1 > len(coords) == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(62478)
		return nil, pgerror.Newf(pgcode.InvalidParameterValue, "index %d out of range of LineString with %d coordinates", index, len(coords))
	} else {
		__antithesis_instrumentation__.Notify(62479)
	}
	__antithesis_instrumentation__.Notify(62470)

	if hasNegIndex {
		__antithesis_instrumentation__.Notify(62480)
		index = len(coords) + index
	} else {
		__antithesis_instrumentation__.Notify(62481)
	}
	__antithesis_instrumentation__.Notify(62471)

	coords[index].Set(point.Coords())

	return lineString.SetCoords(coords)
}

func RemovePoint(lineString geo.Geometry, index int) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(62482)
	g, err := lineString.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(62487)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(62488)
	}
	__antithesis_instrumentation__.Notify(62483)

	lineStringG, ok := g.(*geom.LineString)
	if !ok {
		__antithesis_instrumentation__.Notify(62489)
		e := geom.ErrUnsupportedType{Value: g}
		return geo.Geometry{}, errors.Wrap(e, "geometry to be modified must be a LineString")
	} else {
		__antithesis_instrumentation__.Notify(62490)
	}
	__antithesis_instrumentation__.Notify(62484)

	if lineStringG.NumCoords() == 2 {
		__antithesis_instrumentation__.Notify(62491)
		return geo.Geometry{}, pgerror.Newf(pgcode.InvalidParameterValue, "cannot remove a point from a LineString with only two Points")
	} else {
		__antithesis_instrumentation__.Notify(62492)
	}
	__antithesis_instrumentation__.Notify(62485)

	g, err = removePoint(lineStringG, index)
	if err != nil {
		__antithesis_instrumentation__.Notify(62493)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(62494)
	}
	__antithesis_instrumentation__.Notify(62486)

	return geo.MakeGeometryFromGeomT(g)
}

func removePoint(lineString *geom.LineString, index int) (*geom.LineString, error) {
	__antithesis_instrumentation__.Notify(62495)
	coords := lineString.Coords()

	if index >= len(coords) || func() bool {
		__antithesis_instrumentation__.Notify(62497)
		return index < 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(62498)
		return nil, pgerror.Newf(pgcode.InvalidParameterValue, "index %d out of range of LineString with %d coordinates", index, len(coords))
	} else {
		__antithesis_instrumentation__.Notify(62499)
	}
	__antithesis_instrumentation__.Notify(62496)

	coords = append(coords[:index], coords[index+1:]...)

	return lineString.SetCoords(coords)
}

func LineSubstring(g geo.Geometry, start, end float64) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(62500)
	if start < 0 || func() bool {
		__antithesis_instrumentation__.Notify(62509)
		return start > 1 == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(62510)
		return end < 0 == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(62511)
		return end > 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(62512)
		return g, pgerror.Newf(pgcode.InvalidParameterValue,
			"start and and must be within 0 and 1")
	} else {
		__antithesis_instrumentation__.Notify(62513)
	}
	__antithesis_instrumentation__.Notify(62501)
	if start > end {
		__antithesis_instrumentation__.Notify(62514)
		return g, pgerror.Newf(pgcode.InvalidParameterValue,
			"end must be greater or equal to the start")
	} else {
		__antithesis_instrumentation__.Notify(62515)
	}
	__antithesis_instrumentation__.Notify(62502)

	lineT, err := g.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(62516)
		return g, err
	} else {
		__antithesis_instrumentation__.Notify(62517)
	}
	__antithesis_instrumentation__.Notify(62503)
	lineString, ok := lineT.(*geom.LineString)
	if !ok {
		__antithesis_instrumentation__.Notify(62518)
		return g, pgerror.Newf(pgcode.InvalidParameterValue,
			"geometry has to be of type LineString")
	} else {
		__antithesis_instrumentation__.Notify(62519)
	}
	__antithesis_instrumentation__.Notify(62504)
	if lineString.Empty() {
		__antithesis_instrumentation__.Notify(62520)
		return geo.MakeGeometryFromGeomT(
			geom.NewLineString(geom.XY).SetSRID(lineString.SRID()),
		)
	} else {
		__antithesis_instrumentation__.Notify(62521)
	}
	__antithesis_instrumentation__.Notify(62505)
	if start == end {
		__antithesis_instrumentation__.Notify(62522)
		return LineInterpolatePoints(g, start, false)
	} else {
		__antithesis_instrumentation__.Notify(62523)
	}
	__antithesis_instrumentation__.Notify(62506)

	lsLength := lineString.Length()

	if lsLength == 0 {
		__antithesis_instrumentation__.Notify(62524)
		return geo.MakeGeometryFromGeomT(
			geom.NewPointFlat(geom.XY, lineString.FlatCoords()[0:2]).SetSRID(lineString.SRID()),
		)
	} else {
		__antithesis_instrumentation__.Notify(62525)
	}
	__antithesis_instrumentation__.Notify(62507)

	var newFlatCoords []float64

	startDistance, endDistance := start*lsLength, end*lsLength
	for i := range lineString.Coords() {
		__antithesis_instrumentation__.Notify(62526)
		currentLineString, err := geom.NewLineString(geom.XY).SetCoords(lineString.Coords()[0 : i+1])
		if err != nil {
			__antithesis_instrumentation__.Notify(62529)
			return geo.Geometry{}, err
		} else {
			__antithesis_instrumentation__.Notify(62530)
		}
		__antithesis_instrumentation__.Notify(62527)

		if currentLineString.Length() >= endDistance {
			__antithesis_instrumentation__.Notify(62531)

			if len(newFlatCoords) == 0 {
				__antithesis_instrumentation__.Notify(62534)
				coords, err := interpolateFlatCoordsFromDistance(g, startDistance)
				if err != nil {
					__antithesis_instrumentation__.Notify(62536)
					return geo.Geometry{}, err
				} else {
					__antithesis_instrumentation__.Notify(62537)
				}
				__antithesis_instrumentation__.Notify(62535)
				newFlatCoords = append(newFlatCoords, coords...)
			} else {
				__antithesis_instrumentation__.Notify(62538)
			}
			__antithesis_instrumentation__.Notify(62532)

			coords, err := interpolateFlatCoordsFromDistance(g, endDistance)
			if err != nil {
				__antithesis_instrumentation__.Notify(62539)
				return geo.Geometry{}, err
			} else {
				__antithesis_instrumentation__.Notify(62540)
			}
			__antithesis_instrumentation__.Notify(62533)
			newFlatCoords = append(newFlatCoords, coords...)
			break
		} else {
			__antithesis_instrumentation__.Notify(62541)
		}
		__antithesis_instrumentation__.Notify(62528)

		if currentLineString.Length() >= startDistance {
			__antithesis_instrumentation__.Notify(62542)
			if len(newFlatCoords) == 0 {
				__antithesis_instrumentation__.Notify(62544)

				coords, err := interpolateFlatCoordsFromDistance(g, startDistance)
				if err != nil {
					__antithesis_instrumentation__.Notify(62546)
					return geo.Geometry{}, err
				} else {
					__antithesis_instrumentation__.Notify(62547)
				}
				__antithesis_instrumentation__.Notify(62545)
				newFlatCoords = append(newFlatCoords, coords...)
			} else {
				__antithesis_instrumentation__.Notify(62548)
			}
			__antithesis_instrumentation__.Notify(62543)

			prevCoords := geom.Coord(newFlatCoords[len(newFlatCoords)-2:])
			if !currentLineString.Coord(i).Equal(geom.XY, prevCoords) {
				__antithesis_instrumentation__.Notify(62549)
				newFlatCoords = append(newFlatCoords, currentLineString.Coord(i)...)
			} else {
				__antithesis_instrumentation__.Notify(62550)
			}
		} else {
			__antithesis_instrumentation__.Notify(62551)
		}
	}
	__antithesis_instrumentation__.Notify(62508)
	return geo.MakeGeometryFromGeomT(geom.NewLineStringFlat(geom.XY, newFlatCoords).SetSRID(lineString.SRID()))
}

func interpolateFlatCoordsFromDistance(g geo.Geometry, distance float64) ([]float64, error) {
	__antithesis_instrumentation__.Notify(62552)
	pointEWKB, err := geos.InterpolateLine(g.EWKB(), distance)
	if err != nil {
		__antithesis_instrumentation__.Notify(62555)
		return []float64{}, err
	} else {
		__antithesis_instrumentation__.Notify(62556)
	}
	__antithesis_instrumentation__.Notify(62553)
	point, err := ewkb.Unmarshal(pointEWKB)
	if err != nil {
		__antithesis_instrumentation__.Notify(62557)
		return []float64{}, err
	} else {
		__antithesis_instrumentation__.Notify(62558)
	}
	__antithesis_instrumentation__.Notify(62554)
	return point.FlatCoords(), nil
}
