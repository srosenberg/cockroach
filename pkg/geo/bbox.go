package geo

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/errors"
	"github.com/golang/geo/s2"
	"github.com/twpayne/go-geom"
)

type CartesianBoundingBox struct {
	geopb.BoundingBox
}

func NewCartesianBoundingBox() *CartesianBoundingBox {
	__antithesis_instrumentation__.Notify(58547)
	return nil
}

func (b *CartesianBoundingBox) Repr() string {
	__antithesis_instrumentation__.Notify(58548)

	return fmt.Sprintf(
		"BOX(%s %s,%s %s)",
		strconv.FormatFloat(b.LoX, 'f', -1, 64),
		strconv.FormatFloat(b.LoY, 'f', -1, 64),
		strconv.FormatFloat(b.HiX, 'f', -1, 64),
		strconv.FormatFloat(b.HiY, 'f', -1, 64),
	)
}

func ParseCartesianBoundingBox(s string) (CartesianBoundingBox, error) {
	__antithesis_instrumentation__.Notify(58549)
	b := CartesianBoundingBox{}
	var prefix string
	numScanned, err := fmt.Sscanf(s, "%3s(%f %f,%f %f)", &prefix, &b.LoX, &b.LoY, &b.HiX, &b.HiY)
	if err != nil {
		__antithesis_instrumentation__.Notify(58552)
		return b, errors.Wrapf(err, "error parsing box2d")
	} else {
		__antithesis_instrumentation__.Notify(58553)
	}
	__antithesis_instrumentation__.Notify(58550)
	if numScanned != 5 || func() bool {
		__antithesis_instrumentation__.Notify(58554)
		return strings.ToLower(prefix) != "box" == true
	}() == true {
		__antithesis_instrumentation__.Notify(58555)
		return b, pgerror.Newf(pgcode.InvalidParameterValue, "expected format 'box(min_x min_y,max_x max_y)'")
	} else {
		__antithesis_instrumentation__.Notify(58556)
	}
	__antithesis_instrumentation__.Notify(58551)
	return b, nil
}

func (b *CartesianBoundingBox) Compare(o *CartesianBoundingBox) int {
	__antithesis_instrumentation__.Notify(58557)
	if b.LoX < o.LoX {
		__antithesis_instrumentation__.Notify(58562)
		return -1
	} else {
		__antithesis_instrumentation__.Notify(58563)
		if b.LoX > o.LoX {
			__antithesis_instrumentation__.Notify(58564)
			return 1
		} else {
			__antithesis_instrumentation__.Notify(58565)
		}
	}
	__antithesis_instrumentation__.Notify(58558)

	if b.HiX < o.HiX {
		__antithesis_instrumentation__.Notify(58566)
		return -1
	} else {
		__antithesis_instrumentation__.Notify(58567)
		if b.HiX > o.HiX {
			__antithesis_instrumentation__.Notify(58568)
			return 1
		} else {
			__antithesis_instrumentation__.Notify(58569)
		}
	}
	__antithesis_instrumentation__.Notify(58559)

	if b.LoY < o.LoY {
		__antithesis_instrumentation__.Notify(58570)
		return -1
	} else {
		__antithesis_instrumentation__.Notify(58571)
		if b.LoY > o.LoY {
			__antithesis_instrumentation__.Notify(58572)
			return 1
		} else {
			__antithesis_instrumentation__.Notify(58573)
		}
	}
	__antithesis_instrumentation__.Notify(58560)

	if b.HiY < o.HiY {
		__antithesis_instrumentation__.Notify(58574)
		return -1
	} else {
		__antithesis_instrumentation__.Notify(58575)
		if b.HiY > o.HiY {
			__antithesis_instrumentation__.Notify(58576)
			return 1
		} else {
			__antithesis_instrumentation__.Notify(58577)
		}
	}
	__antithesis_instrumentation__.Notify(58561)

	return 0
}

func (b *CartesianBoundingBox) WithPoint(x, y float64) *CartesianBoundingBox {
	__antithesis_instrumentation__.Notify(58578)
	if b == nil {
		__antithesis_instrumentation__.Notify(58580)
		return &CartesianBoundingBox{
			BoundingBox: geopb.BoundingBox{
				LoX: x,
				HiX: x,
				LoY: y,
				HiY: y,
			},
		}
	} else {
		__antithesis_instrumentation__.Notify(58581)
	}
	__antithesis_instrumentation__.Notify(58579)
	b.BoundingBox = geopb.BoundingBox{
		LoX: math.Min(b.LoX, x),
		HiX: math.Max(b.HiX, x),
		LoY: math.Min(b.LoY, y),
		HiY: math.Max(b.HiY, y),
	}
	return b
}

func (b *CartesianBoundingBox) AddPoint(x, y float64) *CartesianBoundingBox {
	__antithesis_instrumentation__.Notify(58582)
	if b == nil {
		__antithesis_instrumentation__.Notify(58584)
		return &CartesianBoundingBox{
			BoundingBox: geopb.BoundingBox{
				LoX: x,
				HiX: x,
				LoY: y,
				HiY: y,
			},
		}
	} else {
		__antithesis_instrumentation__.Notify(58585)
	}
	__antithesis_instrumentation__.Notify(58583)
	return &CartesianBoundingBox{
		BoundingBox: geopb.BoundingBox{
			LoX: math.Min(b.LoX, x),
			HiX: math.Max(b.HiX, x),
			LoY: math.Min(b.LoY, y),
			HiY: math.Max(b.HiY, y),
		},
	}
}

func (b *CartesianBoundingBox) Combine(o *CartesianBoundingBox) *CartesianBoundingBox {
	__antithesis_instrumentation__.Notify(58586)
	if o == nil {
		__antithesis_instrumentation__.Notify(58588)
		return b
	} else {
		__antithesis_instrumentation__.Notify(58589)
	}
	__antithesis_instrumentation__.Notify(58587)
	return b.AddPoint(o.LoX, o.LoY).AddPoint(o.HiX, o.HiY)
}

func (b *CartesianBoundingBox) Buffer(deltaX, deltaY float64) *CartesianBoundingBox {
	__antithesis_instrumentation__.Notify(58590)
	if b == nil {
		__antithesis_instrumentation__.Notify(58592)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(58593)
	}
	__antithesis_instrumentation__.Notify(58591)
	return &CartesianBoundingBox{
		BoundingBox: geopb.BoundingBox{
			LoX: b.LoX - deltaX,
			HiX: b.HiX + deltaX,
			LoY: b.LoY - deltaY,
			HiY: b.HiY + deltaY,
		},
	}
}

func (b *CartesianBoundingBox) Intersects(o *CartesianBoundingBox) bool {
	__antithesis_instrumentation__.Notify(58594)

	if b == nil || func() bool {
		__antithesis_instrumentation__.Notify(58597)
		return o == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(58598)
		return false
	} else {
		__antithesis_instrumentation__.Notify(58599)
	}
	__antithesis_instrumentation__.Notify(58595)
	if b.LoY > o.HiY || func() bool {
		__antithesis_instrumentation__.Notify(58600)
		return o.LoY > b.HiY == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(58601)
		return b.LoX > o.HiX == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(58602)
		return o.LoX > b.HiX == true
	}() == true {
		__antithesis_instrumentation__.Notify(58603)
		return false
	} else {
		__antithesis_instrumentation__.Notify(58604)
	}
	__antithesis_instrumentation__.Notify(58596)
	return true
}

func (b *CartesianBoundingBox) Covers(o *CartesianBoundingBox) bool {
	__antithesis_instrumentation__.Notify(58605)
	if b == nil || func() bool {
		__antithesis_instrumentation__.Notify(58607)
		return o == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(58608)
		return false
	} else {
		__antithesis_instrumentation__.Notify(58609)
	}
	__antithesis_instrumentation__.Notify(58606)
	return b.LoX <= o.LoX && func() bool {
		__antithesis_instrumentation__.Notify(58610)
		return o.LoX <= b.HiX == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(58611)
		return b.LoX <= o.HiX == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(58612)
		return o.HiX <= b.HiX == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(58613)
		return b.LoY <= o.LoY == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(58614)
		return o.LoY <= b.HiY == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(58615)
		return b.LoY <= o.HiY == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(58616)
		return o.HiY <= b.HiY == true
	}() == true
}

func (b *CartesianBoundingBox) ToGeomT(srid geopb.SRID) geom.T {
	__antithesis_instrumentation__.Notify(58617)
	if b.LoX == b.HiX && func() bool {
		__antithesis_instrumentation__.Notify(58620)
		return b.LoY == b.HiY == true
	}() == true {
		__antithesis_instrumentation__.Notify(58621)
		return geom.NewPointFlat(geom.XY, []float64{b.LoX, b.LoY}).SetSRID(int(srid))
	} else {
		__antithesis_instrumentation__.Notify(58622)
	}
	__antithesis_instrumentation__.Notify(58618)
	if b.LoX == b.HiX || func() bool {
		__antithesis_instrumentation__.Notify(58623)
		return b.LoY == b.HiY == true
	}() == true {
		__antithesis_instrumentation__.Notify(58624)
		return geom.NewLineStringFlat(geom.XY, []float64{b.LoX, b.LoY, b.HiX, b.HiY}).SetSRID(int(srid))
	} else {
		__antithesis_instrumentation__.Notify(58625)
	}
	__antithesis_instrumentation__.Notify(58619)
	return geom.NewPolygonFlat(
		geom.XY,
		[]float64{
			b.LoX, b.LoY,
			b.LoX, b.HiY,
			b.HiX, b.HiY,
			b.HiX, b.LoY,
			b.LoX, b.LoY,
		},
		[]int{10},
	).SetSRID(int(srid))
}

func boundingBoxFromGeomT(g geom.T, soType geopb.SpatialObjectType) (*geopb.BoundingBox, error) {
	__antithesis_instrumentation__.Notify(58626)
	switch soType {
	case geopb.SpatialObjectType_GeometryType:
		__antithesis_instrumentation__.Notify(58628)
		ret := BoundingBoxFromGeomTGeometryType(g)
		if ret == nil {
			__antithesis_instrumentation__.Notify(58634)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(58635)
		}
		__antithesis_instrumentation__.Notify(58629)
		return &ret.BoundingBox, nil
	case geopb.SpatialObjectType_GeographyType:
		__antithesis_instrumentation__.Notify(58630)
		rect, err := boundingBoxFromGeomTGeographyType(g)
		if err != nil {
			__antithesis_instrumentation__.Notify(58636)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(58637)
		}
		__antithesis_instrumentation__.Notify(58631)
		if rect.IsEmpty() {
			__antithesis_instrumentation__.Notify(58638)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(58639)
		}
		__antithesis_instrumentation__.Notify(58632)
		return &geopb.BoundingBox{
			LoX: rect.Lng.Lo,
			HiX: rect.Lng.Hi,
			LoY: rect.Lat.Lo,
			HiY: rect.Lat.Hi,
		}, nil
	default:
		__antithesis_instrumentation__.Notify(58633)
	}
	__antithesis_instrumentation__.Notify(58627)
	return nil, pgerror.Newf(pgcode.InvalidParameterValue, "unknown spatial type: %s", soType)
}

func BoundingBoxFromGeomTGeometryType(g geom.T) *CartesianBoundingBox {
	__antithesis_instrumentation__.Notify(58640)
	if g.Empty() {
		__antithesis_instrumentation__.Notify(58643)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(58644)
	}
	__antithesis_instrumentation__.Notify(58641)
	bbox := NewCartesianBoundingBox()
	switch g := g.(type) {
	case *geom.GeometryCollection:
		__antithesis_instrumentation__.Notify(58645)
		for i := 0; i < g.NumGeoms(); i++ {
			__antithesis_instrumentation__.Notify(58647)
			shapeBBox := BoundingBoxFromGeomTGeometryType(g.Geom(i))
			if shapeBBox == nil {
				__antithesis_instrumentation__.Notify(58649)
				continue
			} else {
				__antithesis_instrumentation__.Notify(58650)
			}
			__antithesis_instrumentation__.Notify(58648)
			bbox = bbox.WithPoint(shapeBBox.LoX, shapeBBox.LoY).WithPoint(shapeBBox.HiX, shapeBBox.HiY)
		}
	default:
		__antithesis_instrumentation__.Notify(58646)
		flatCoords := g.FlatCoords()
		for i := 0; i < len(flatCoords); i += g.Stride() {
			__antithesis_instrumentation__.Notify(58651)
			bbox = bbox.WithPoint(flatCoords[i], flatCoords[i+1])
		}
	}
	__antithesis_instrumentation__.Notify(58642)
	return bbox
}

func boundingBoxFromGeomTGeographyType(g geom.T) (s2.Rect, error) {
	__antithesis_instrumentation__.Notify(58652)
	if g.Empty() {
		__antithesis_instrumentation__.Notify(58655)
		return s2.EmptyRect(), nil
	} else {
		__antithesis_instrumentation__.Notify(58656)
	}
	__antithesis_instrumentation__.Notify(58653)
	rect := s2.EmptyRect()
	switch g := g.(type) {
	case *geom.Point:
		__antithesis_instrumentation__.Notify(58657)
		return geogPointsBBox(g), nil
	case *geom.MultiPoint:
		__antithesis_instrumentation__.Notify(58658)
		return geogPointsBBox(g), nil
	case *geom.LineString:
		__antithesis_instrumentation__.Notify(58659)
		return geogLineBBox(g), nil
	case *geom.MultiLineString:
		__antithesis_instrumentation__.Notify(58660)
		for i := 0; i < g.NumLineStrings(); i++ {
			__antithesis_instrumentation__.Notify(58665)
			rect = rect.Union(geogLineBBox(g.LineString(i)))
		}
	case *geom.Polygon:
		__antithesis_instrumentation__.Notify(58661)
		for i := 0; i < g.NumLinearRings(); i++ {
			__antithesis_instrumentation__.Notify(58666)
			rect = rect.Union(geogLineBBox(g.LinearRing(i)))
		}
	case *geom.MultiPolygon:
		__antithesis_instrumentation__.Notify(58662)
		for i := 0; i < g.NumPolygons(); i++ {
			__antithesis_instrumentation__.Notify(58667)
			polyRect, err := boundingBoxFromGeomTGeographyType(g.Polygon(i))
			if err != nil {
				__antithesis_instrumentation__.Notify(58669)
				return s2.EmptyRect(), err
			} else {
				__antithesis_instrumentation__.Notify(58670)
			}
			__antithesis_instrumentation__.Notify(58668)
			rect = rect.Union(polyRect)
		}
	case *geom.GeometryCollection:
		__antithesis_instrumentation__.Notify(58663)
		for i := 0; i < g.NumGeoms(); i++ {
			__antithesis_instrumentation__.Notify(58671)
			collRect, err := boundingBoxFromGeomTGeographyType(g.Geom(i))
			if err != nil {
				__antithesis_instrumentation__.Notify(58673)
				return s2.EmptyRect(), err
			} else {
				__antithesis_instrumentation__.Notify(58674)
			}
			__antithesis_instrumentation__.Notify(58672)
			rect = rect.Union(collRect)
		}
	default:
		__antithesis_instrumentation__.Notify(58664)
		return s2.EmptyRect(), errors.Errorf("unknown type %T", g)
	}
	__antithesis_instrumentation__.Notify(58654)
	return rect, nil
}

func geogPointsBBox(g geom.T) s2.Rect {
	__antithesis_instrumentation__.Notify(58675)
	rect := s2.EmptyRect()
	flatCoords := g.FlatCoords()
	for i := 0; i < len(flatCoords); i += g.Stride() {
		__antithesis_instrumentation__.Notify(58677)
		rect = rect.AddPoint(s2.LatLngFromDegrees(flatCoords[i+1], flatCoords[i]))
	}
	__antithesis_instrumentation__.Notify(58676)
	return rect
}

func geogLineBBox(g geom.T) s2.Rect {
	__antithesis_instrumentation__.Notify(58678)
	bounder := s2.NewRectBounder()
	flatCoords := g.FlatCoords()
	for i := 0; i < len(flatCoords); i += g.Stride() {
		__antithesis_instrumentation__.Notify(58680)
		bounder.AddPoint(s2.PointFromLatLng(s2.LatLngFromDegrees(flatCoords[i+1], flatCoords[i])))
	}
	__antithesis_instrumentation__.Notify(58679)
	return bounder.RectBound()
}
