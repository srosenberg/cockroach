// Package geo contains the base types for spatial data type operations.
package geo

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"encoding/binary"
	"math"

	"github.com/cockroachdb/cockroach/pkg/geo/geographiclib"
	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/geo/geoprojbase"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
	"github.com/golang/geo/r1"
	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/ewkb"
)

var DefaultEWKBEncodingFormat = binary.LittleEndian

type EmptyBehavior uint8

const (
	EmptyBehaviorError EmptyBehavior = 0

	EmptyBehaviorOmit EmptyBehavior = 1
)

type FnExclusivity bool

const MaxAllowedSplitPoints = 65336

const (
	FnExclusive FnExclusivity = true

	FnInclusive FnExclusivity = false
)

func SpatialObjectFitsColumnMetadata(
	so geopb.SpatialObject, srid geopb.SRID, shapeType geopb.ShapeType,
) error {
	__antithesis_instrumentation__.Notify(58806)

	if srid != 0 && func() bool {
		__antithesis_instrumentation__.Notify(58809)
		return so.SRID != srid == true
	}() == true {
		__antithesis_instrumentation__.Notify(58810)
		return pgerror.Newf(
			pgcode.InvalidParameterValue,
			"object SRID %d does not match column SRID %d",
			so.SRID,
			srid,
		)
	} else {
		__antithesis_instrumentation__.Notify(58811)
	}
	__antithesis_instrumentation__.Notify(58807)

	switch shapeType {
	case geopb.ShapeType_Unset:
		__antithesis_instrumentation__.Notify(58812)
		break
	case geopb.ShapeType_Geometry, geopb.ShapeType_GeometryM, geopb.ShapeType_GeometryZ, geopb.ShapeType_GeometryZM:
		__antithesis_instrumentation__.Notify(58813)
		if ShapeTypeToLayout(shapeType) != ShapeTypeToLayout(so.ShapeType) {
			__antithesis_instrumentation__.Notify(58815)
			return pgerror.Newf(
				pgcode.InvalidParameterValue,
				"object type %s does not match column dimensionality %s",
				so.ShapeType,
				shapeType,
			)
		} else {
			__antithesis_instrumentation__.Notify(58816)
		}
	default:
		__antithesis_instrumentation__.Notify(58814)
		if shapeType != so.ShapeType {
			__antithesis_instrumentation__.Notify(58817)
			return pgerror.Newf(
				pgcode.InvalidParameterValue,
				"object type %s does not match column type %s",
				so.ShapeType,
				shapeType,
			)
		} else {
			__antithesis_instrumentation__.Notify(58818)
		}
	}
	__antithesis_instrumentation__.Notify(58808)
	return nil
}

func ShapeTypeToLayout(s geopb.ShapeType) geom.Layout {
	__antithesis_instrumentation__.Notify(58819)
	switch {
	case (s&geopb.MShapeTypeFlag > 0) && func() bool {
		__antithesis_instrumentation__.Notify(58824)
		return (s&geopb.ZShapeTypeFlag > 0) == true
	}() == true:
		__antithesis_instrumentation__.Notify(58820)
		return geom.XYZM
	case s&geopb.ZShapeTypeFlag > 0:
		__antithesis_instrumentation__.Notify(58821)
		return geom.XYZ
	case s&geopb.MShapeTypeFlag > 0:
		__antithesis_instrumentation__.Notify(58822)
		return geom.XYM
	default:
		__antithesis_instrumentation__.Notify(58823)
		return geom.XY
	}
}

type Geometry struct {
	spatialObject geopb.SpatialObject
}

func MakeGeometry(spatialObject geopb.SpatialObject) (Geometry, error) {
	__antithesis_instrumentation__.Notify(58825)
	if spatialObject.SRID != 0 {
		__antithesis_instrumentation__.Notify(58828)
		if _, err := geoprojbase.Projection(spatialObject.SRID); err != nil {
			__antithesis_instrumentation__.Notify(58829)
			return Geometry{}, err
		} else {
			__antithesis_instrumentation__.Notify(58830)
		}
	} else {
		__antithesis_instrumentation__.Notify(58831)
	}
	__antithesis_instrumentation__.Notify(58826)
	if spatialObject.Type != geopb.SpatialObjectType_GeometryType {
		__antithesis_instrumentation__.Notify(58832)
		return Geometry{}, pgerror.Newf(
			pgcode.InvalidObjectDefinition,
			"expected geometry type, found %s",
			spatialObject.Type,
		)
	} else {
		__antithesis_instrumentation__.Notify(58833)
	}
	__antithesis_instrumentation__.Notify(58827)
	return Geometry{spatialObject: spatialObject}, nil
}

func MakeGeometryUnsafe(spatialObject geopb.SpatialObject) Geometry {
	__antithesis_instrumentation__.Notify(58834)
	return Geometry{spatialObject: spatialObject}
}

func MakeGeometryFromPointCoords(x, y float64) (Geometry, error) {
	__antithesis_instrumentation__.Notify(58835)
	return MakeGeometryFromLayoutAndPointCoords(geom.XY, []float64{x, y})
}

func MakeGeometryFromLayoutAndPointCoords(
	layout geom.Layout, flatCoords []float64,
) (Geometry, error) {
	__antithesis_instrumentation__.Notify(58836)

	switch {
	case layout == geom.XY && func() bool {
		__antithesis_instrumentation__.Notify(58844)
		return len(flatCoords) == 2 == true
	}() == true:
		__antithesis_instrumentation__.Notify(58839)
	case layout == geom.XYM && func() bool {
		__antithesis_instrumentation__.Notify(58845)
		return len(flatCoords) == 3 == true
	}() == true:
		__antithesis_instrumentation__.Notify(58840)
	case layout == geom.XYZ && func() bool {
		__antithesis_instrumentation__.Notify(58846)
		return len(flatCoords) == 3 == true
	}() == true:
		__antithesis_instrumentation__.Notify(58841)
	case layout == geom.XYZM && func() bool {
		__antithesis_instrumentation__.Notify(58847)
		return len(flatCoords) == 4 == true
	}() == true:
		__antithesis_instrumentation__.Notify(58842)
	default:
		__antithesis_instrumentation__.Notify(58843)
		return Geometry{}, pgerror.Newf(
			pgcode.InvalidParameterValue,
			"mismatch between layout %d and stride %d",
			layout,
			len(flatCoords),
		)
	}
	__antithesis_instrumentation__.Notify(58837)
	s, err := spatialObjectFromGeomT(geom.NewPointFlat(layout, flatCoords), geopb.SpatialObjectType_GeometryType)
	if err != nil {
		__antithesis_instrumentation__.Notify(58848)
		return Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(58849)
	}
	__antithesis_instrumentation__.Notify(58838)
	return MakeGeometry(s)
}

func MakeGeometryFromGeomT(g geom.T) (Geometry, error) {
	__antithesis_instrumentation__.Notify(58850)
	spatialObject, err := spatialObjectFromGeomT(g, geopb.SpatialObjectType_GeometryType)
	if err != nil {
		__antithesis_instrumentation__.Notify(58852)
		return Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(58853)
	}
	__antithesis_instrumentation__.Notify(58851)
	return MakeGeometry(spatialObject)
}

func ParseGeometry(str string) (Geometry, error) {
	__antithesis_instrumentation__.Notify(58854)
	spatialObject, err := parseAmbiguousText(geopb.SpatialObjectType_GeometryType, str, geopb.DefaultGeometrySRID)
	if err != nil {
		__antithesis_instrumentation__.Notify(58856)
		return Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(58857)
	}
	__antithesis_instrumentation__.Notify(58855)
	return MakeGeometry(spatialObject)
}

func MustParseGeometry(str string) Geometry {
	__antithesis_instrumentation__.Notify(58858)
	g, err := ParseGeometry(str)
	if err != nil {
		__antithesis_instrumentation__.Notify(58860)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(58861)
	}
	__antithesis_instrumentation__.Notify(58859)
	return g
}

func ParseGeometryFromEWKT(
	ewkt geopb.EWKT, srid geopb.SRID, defaultSRIDOverwriteSetting defaultSRIDOverwriteSetting,
) (Geometry, error) {
	__antithesis_instrumentation__.Notify(58862)
	g, err := parseEWKT(geopb.SpatialObjectType_GeometryType, ewkt, srid, defaultSRIDOverwriteSetting)
	if err != nil {
		__antithesis_instrumentation__.Notify(58864)
		return Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(58865)
	}
	__antithesis_instrumentation__.Notify(58863)
	return MakeGeometry(g)
}

func ParseGeometryFromEWKB(ewkb geopb.EWKB) (Geometry, error) {
	__antithesis_instrumentation__.Notify(58866)
	g, err := parseEWKB(geopb.SpatialObjectType_GeometryType, ewkb, geopb.DefaultGeometrySRID, DefaultSRIDIsHint)
	if err != nil {
		__antithesis_instrumentation__.Notify(58868)
		return Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(58869)
	}
	__antithesis_instrumentation__.Notify(58867)
	return MakeGeometry(g)
}

func ParseGeometryFromEWKBAndSRID(ewkb geopb.EWKB, srid geopb.SRID) (Geometry, error) {
	__antithesis_instrumentation__.Notify(58870)
	g, err := parseEWKB(geopb.SpatialObjectType_GeometryType, ewkb, srid, DefaultSRIDShouldOverwrite)
	if err != nil {
		__antithesis_instrumentation__.Notify(58872)
		return Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(58873)
	}
	__antithesis_instrumentation__.Notify(58871)
	return MakeGeometry(g)
}

func MustParseGeometryFromEWKB(ewkb geopb.EWKB) Geometry {
	__antithesis_instrumentation__.Notify(58874)
	ret, err := ParseGeometryFromEWKB(ewkb)
	if err != nil {
		__antithesis_instrumentation__.Notify(58876)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(58877)
	}
	__antithesis_instrumentation__.Notify(58875)
	return ret
}

func ParseGeometryFromGeoJSON(json []byte) (Geometry, error) {
	__antithesis_instrumentation__.Notify(58878)

	g, err := parseGeoJSON(geopb.SpatialObjectType_GeometryType, json, geopb.DefaultGeographySRID)
	if err != nil {
		__antithesis_instrumentation__.Notify(58880)
		return Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(58881)
	}
	__antithesis_instrumentation__.Notify(58879)
	return MakeGeometry(g)
}

func ParseGeometryFromEWKBUnsafe(ewkb geopb.EWKB) (Geometry, error) {
	__antithesis_instrumentation__.Notify(58882)
	base, err := parseEWKBRaw(geopb.SpatialObjectType_GeometryType, ewkb)
	if err != nil {
		__antithesis_instrumentation__.Notify(58884)
		return Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(58885)
	}
	__antithesis_instrumentation__.Notify(58883)
	return MakeGeometryUnsafe(base), nil
}

func (g *Geometry) AsGeography() (Geography, error) {
	__antithesis_instrumentation__.Notify(58886)
	srid := g.SRID()
	if srid == 0 {
		__antithesis_instrumentation__.Notify(58889)

		srid = geopb.DefaultGeographySRID
	} else {
		__antithesis_instrumentation__.Notify(58890)
	}
	__antithesis_instrumentation__.Notify(58887)
	spatialObject, err := adjustSpatialObject(g.spatialObject, srid, geopb.SpatialObjectType_GeographyType)
	if err != nil {
		__antithesis_instrumentation__.Notify(58891)
		return Geography{}, err
	} else {
		__antithesis_instrumentation__.Notify(58892)
	}
	__antithesis_instrumentation__.Notify(58888)
	return MakeGeography(spatialObject)
}

func (g *Geometry) CloneWithSRID(srid geopb.SRID) (Geometry, error) {
	__antithesis_instrumentation__.Notify(58893)
	spatialObject, err := adjustSpatialObject(g.spatialObject, srid, geopb.SpatialObjectType_GeometryType)
	if err != nil {
		__antithesis_instrumentation__.Notify(58895)
		return Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(58896)
	}
	__antithesis_instrumentation__.Notify(58894)
	return MakeGeometry(spatialObject)
}

func adjustSpatialObject(
	so geopb.SpatialObject, srid geopb.SRID, soType geopb.SpatialObjectType,
) (geopb.SpatialObject, error) {
	__antithesis_instrumentation__.Notify(58897)
	t, err := ewkb.Unmarshal(so.EWKB)
	if err != nil {
		__antithesis_instrumentation__.Notify(58899)
		return geopb.SpatialObject{}, err
	} else {
		__antithesis_instrumentation__.Notify(58900)
	}
	__antithesis_instrumentation__.Notify(58898)
	AdjustGeomTSRID(t, srid)
	return spatialObjectFromGeomT(t, soType)
}

func (g *Geometry) AsGeomT() (geom.T, error) {
	__antithesis_instrumentation__.Notify(58901)
	return ewkb.Unmarshal(g.spatialObject.EWKB)
}

func (g *Geometry) Empty() bool {
	__antithesis_instrumentation__.Notify(58902)
	return g.spatialObject.BoundingBox == nil
}

func (g *Geometry) EWKB() geopb.EWKB {
	__antithesis_instrumentation__.Notify(58903)
	return g.spatialObject.EWKB
}

func (g *Geometry) SpatialObject() geopb.SpatialObject {
	__antithesis_instrumentation__.Notify(58904)
	return g.spatialObject
}

func (g *Geometry) SpatialObjectRef() *geopb.SpatialObject {
	__antithesis_instrumentation__.Notify(58905)
	return &g.spatialObject
}

func (g *Geometry) EWKBHex() string {
	__antithesis_instrumentation__.Notify(58906)
	return g.spatialObject.EWKBHex()
}

func (g *Geometry) SRID() geopb.SRID {
	__antithesis_instrumentation__.Notify(58907)
	return g.spatialObject.SRID
}

func (g *Geometry) ShapeType() geopb.ShapeType {
	__antithesis_instrumentation__.Notify(58908)
	return g.spatialObject.ShapeType
}

func (g *Geometry) ShapeType2D() geopb.ShapeType {
	__antithesis_instrumentation__.Notify(58909)
	return g.ShapeType().To2D()
}

func (g *Geometry) CartesianBoundingBox() *CartesianBoundingBox {
	__antithesis_instrumentation__.Notify(58910)
	if g.spatialObject.BoundingBox == nil {
		__antithesis_instrumentation__.Notify(58912)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(58913)
	}
	__antithesis_instrumentation__.Notify(58911)
	return &CartesianBoundingBox{BoundingBox: *g.spatialObject.BoundingBox}
}

func (g *Geometry) BoundingBoxRef() *geopb.BoundingBox {
	__antithesis_instrumentation__.Notify(58914)
	return g.spatialObject.BoundingBox
}

func (g *Geometry) SpaceCurveIndex() (uint64, error) {
	__antithesis_instrumentation__.Notify(58915)
	bbox := g.CartesianBoundingBox()
	if bbox == nil {
		__antithesis_instrumentation__.Notify(58919)
		return 0, nil
	} else {
		__antithesis_instrumentation__.Notify(58920)
	}
	__antithesis_instrumentation__.Notify(58916)
	centerX := (bbox.BoundingBox.LoX + bbox.BoundingBox.HiX) / 2
	centerY := (bbox.BoundingBox.LoY + bbox.BoundingBox.HiY) / 2

	bounds := geoprojbase.Bounds{
		MinX: math.MinInt32,
		MaxX: math.MaxInt32,
		MinY: math.MinInt32,
		MaxY: math.MaxInt32,
	}
	if g.SRID() != 0 {
		__antithesis_instrumentation__.Notify(58921)
		proj, err := geoprojbase.Projection(g.SRID())
		if err != nil {
			__antithesis_instrumentation__.Notify(58923)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(58924)
		}
		__antithesis_instrumentation__.Notify(58922)
		bounds = proj.Bounds
	} else {
		__antithesis_instrumentation__.Notify(58925)
	}
	__antithesis_instrumentation__.Notify(58917)

	if centerX > bounds.MaxX || func() bool {
		__antithesis_instrumentation__.Notify(58926)
		return centerY > bounds.MaxY == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(58927)
		return centerX < bounds.MinX == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(58928)
		return centerY < bounds.MinY == true
	}() == true {
		__antithesis_instrumentation__.Notify(58929)
		return math.MaxUint64, nil
	} else {
		__antithesis_instrumentation__.Notify(58930)
	}
	__antithesis_instrumentation__.Notify(58918)

	const boxLength = 1 << 32

	xBounds := (bounds.MaxX - bounds.MinX) + 1
	yBounds := (bounds.MaxY - bounds.MinY) + 1

	xPos := uint64(((centerX - bounds.MinX) / xBounds) * boxLength)
	yPos := uint64(((centerY - bounds.MinY) / yBounds) * boxLength)
	return hilbertInverse(boxLength, xPos, yPos), nil
}

func (g *Geometry) Compare(o Geometry) int {
	__antithesis_instrumentation__.Notify(58931)
	lhs, err := g.SpaceCurveIndex()
	if err != nil {
		__antithesis_instrumentation__.Notify(58936)

		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(58937)
	}
	__antithesis_instrumentation__.Notify(58932)
	rhs, err := o.SpaceCurveIndex()
	if err != nil {
		__antithesis_instrumentation__.Notify(58938)

		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(58939)
	}
	__antithesis_instrumentation__.Notify(58933)
	if lhs > rhs {
		__antithesis_instrumentation__.Notify(58940)
		return 1
	} else {
		__antithesis_instrumentation__.Notify(58941)
	}
	__antithesis_instrumentation__.Notify(58934)
	if lhs < rhs {
		__antithesis_instrumentation__.Notify(58942)
		return -1
	} else {
		__antithesis_instrumentation__.Notify(58943)
	}
	__antithesis_instrumentation__.Notify(58935)
	return compareSpatialObjectBytes(g.SpatialObjectRef(), o.SpatialObjectRef())
}

type Geography struct {
	spatialObject geopb.SpatialObject
}

func MakeGeography(spatialObject geopb.SpatialObject) (Geography, error) {
	__antithesis_instrumentation__.Notify(58944)
	projection, err := geoprojbase.Projection(spatialObject.SRID)
	if err != nil {
		__antithesis_instrumentation__.Notify(58948)
		return Geography{}, err
	} else {
		__antithesis_instrumentation__.Notify(58949)
	}
	__antithesis_instrumentation__.Notify(58945)
	if !projection.IsLatLng {
		__antithesis_instrumentation__.Notify(58950)
		return Geography{}, pgerror.Newf(
			pgcode.InvalidParameterValue,
			"SRID %d cannot be used for geography as it is not in a lon/lat coordinate system",
			spatialObject.SRID,
		)
	} else {
		__antithesis_instrumentation__.Notify(58951)
	}
	__antithesis_instrumentation__.Notify(58946)
	if spatialObject.Type != geopb.SpatialObjectType_GeographyType {
		__antithesis_instrumentation__.Notify(58952)
		return Geography{}, pgerror.Newf(
			pgcode.InvalidObjectDefinition,
			"expected geography type, found %s",
			spatialObject.Type,
		)
	} else {
		__antithesis_instrumentation__.Notify(58953)
	}
	__antithesis_instrumentation__.Notify(58947)
	return Geography{spatialObject: spatialObject}, nil
}

func MakeGeographyUnsafe(spatialObject geopb.SpatialObject) Geography {
	__antithesis_instrumentation__.Notify(58954)
	return Geography{spatialObject: spatialObject}
}

func MakeGeographyFromGeomT(g geom.T) (Geography, error) {
	__antithesis_instrumentation__.Notify(58955)
	spatialObject, err := spatialObjectFromGeomT(g, geopb.SpatialObjectType_GeographyType)
	if err != nil {
		__antithesis_instrumentation__.Notify(58957)
		return Geography{}, err
	} else {
		__antithesis_instrumentation__.Notify(58958)
	}
	__antithesis_instrumentation__.Notify(58956)
	return MakeGeography(spatialObject)
}

func MustMakeGeographyFromGeomT(g geom.T) Geography {
	__antithesis_instrumentation__.Notify(58959)
	ret, err := MakeGeographyFromGeomT(g)
	if err != nil {
		__antithesis_instrumentation__.Notify(58961)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(58962)
	}
	__antithesis_instrumentation__.Notify(58960)
	return ret
}

func ParseGeography(str string) (Geography, error) {
	__antithesis_instrumentation__.Notify(58963)
	spatialObject, err := parseAmbiguousText(geopb.SpatialObjectType_GeographyType, str, geopb.DefaultGeographySRID)
	if err != nil {
		__antithesis_instrumentation__.Notify(58965)
		return Geography{}, err
	} else {
		__antithesis_instrumentation__.Notify(58966)
	}
	__antithesis_instrumentation__.Notify(58964)
	return MakeGeography(spatialObject)
}

func MustParseGeography(str string) Geography {
	__antithesis_instrumentation__.Notify(58967)
	g, err := ParseGeography(str)
	if err != nil {
		__antithesis_instrumentation__.Notify(58969)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(58970)
	}
	__antithesis_instrumentation__.Notify(58968)
	return g
}

func ParseGeographyFromEWKT(
	ewkt geopb.EWKT, srid geopb.SRID, defaultSRIDOverwriteSetting defaultSRIDOverwriteSetting,
) (Geography, error) {
	__antithesis_instrumentation__.Notify(58971)
	g, err := parseEWKT(geopb.SpatialObjectType_GeographyType, ewkt, srid, defaultSRIDOverwriteSetting)
	if err != nil {
		__antithesis_instrumentation__.Notify(58973)
		return Geography{}, err
	} else {
		__antithesis_instrumentation__.Notify(58974)
	}
	__antithesis_instrumentation__.Notify(58972)
	return MakeGeography(g)
}

func ParseGeographyFromEWKB(ewkb geopb.EWKB) (Geography, error) {
	__antithesis_instrumentation__.Notify(58975)
	g, err := parseEWKB(geopb.SpatialObjectType_GeographyType, ewkb, geopb.DefaultGeographySRID, DefaultSRIDIsHint)
	if err != nil {
		__antithesis_instrumentation__.Notify(58977)
		return Geography{}, err
	} else {
		__antithesis_instrumentation__.Notify(58978)
	}
	__antithesis_instrumentation__.Notify(58976)
	return MakeGeography(g)
}

func ParseGeographyFromEWKBAndSRID(ewkb geopb.EWKB, srid geopb.SRID) (Geography, error) {
	__antithesis_instrumentation__.Notify(58979)
	g, err := parseEWKB(geopb.SpatialObjectType_GeographyType, ewkb, srid, DefaultSRIDShouldOverwrite)
	if err != nil {
		__antithesis_instrumentation__.Notify(58981)
		return Geography{}, err
	} else {
		__antithesis_instrumentation__.Notify(58982)
	}
	__antithesis_instrumentation__.Notify(58980)
	return MakeGeography(g)
}

func MustParseGeographyFromEWKB(ewkb geopb.EWKB) Geography {
	__antithesis_instrumentation__.Notify(58983)
	ret, err := ParseGeographyFromEWKB(ewkb)
	if err != nil {
		__antithesis_instrumentation__.Notify(58985)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(58986)
	}
	__antithesis_instrumentation__.Notify(58984)
	return ret
}

func ParseGeographyFromGeoJSON(json []byte) (Geography, error) {
	__antithesis_instrumentation__.Notify(58987)
	g, err := parseGeoJSON(geopb.SpatialObjectType_GeographyType, json, geopb.DefaultGeographySRID)
	if err != nil {
		__antithesis_instrumentation__.Notify(58989)
		return Geography{}, err
	} else {
		__antithesis_instrumentation__.Notify(58990)
	}
	__antithesis_instrumentation__.Notify(58988)
	return MakeGeography(g)
}

func ParseGeographyFromEWKBUnsafe(ewkb geopb.EWKB) (Geography, error) {
	__antithesis_instrumentation__.Notify(58991)
	base, err := parseEWKBRaw(geopb.SpatialObjectType_GeographyType, ewkb)
	if err != nil {
		__antithesis_instrumentation__.Notify(58993)
		return Geography{}, err
	} else {
		__antithesis_instrumentation__.Notify(58994)
	}
	__antithesis_instrumentation__.Notify(58992)
	return MakeGeographyUnsafe(base), nil
}

func (g *Geography) CloneWithSRID(srid geopb.SRID) (Geography, error) {
	__antithesis_instrumentation__.Notify(58995)
	spatialObject, err := adjustSpatialObject(g.spatialObject, srid, geopb.SpatialObjectType_GeographyType)
	if err != nil {
		__antithesis_instrumentation__.Notify(58997)
		return Geography{}, err
	} else {
		__antithesis_instrumentation__.Notify(58998)
	}
	__antithesis_instrumentation__.Notify(58996)
	return MakeGeography(spatialObject)
}

func (g *Geography) AsGeometry() (Geometry, error) {
	__antithesis_instrumentation__.Notify(58999)
	spatialObject, err := adjustSpatialObject(g.spatialObject, g.SRID(), geopb.SpatialObjectType_GeometryType)
	if err != nil {
		__antithesis_instrumentation__.Notify(59001)
		return Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(59002)
	}
	__antithesis_instrumentation__.Notify(59000)
	return MakeGeometry(spatialObject)
}

func (g *Geography) AsGeomT() (geom.T, error) {
	__antithesis_instrumentation__.Notify(59003)
	return ewkb.Unmarshal(g.spatialObject.EWKB)
}

func (g *Geography) EWKB() geopb.EWKB {
	__antithesis_instrumentation__.Notify(59004)
	return g.spatialObject.EWKB
}

func (g *Geography) SpatialObject() geopb.SpatialObject {
	__antithesis_instrumentation__.Notify(59005)
	return g.spatialObject
}

func (g *Geography) SpatialObjectRef() *geopb.SpatialObject {
	__antithesis_instrumentation__.Notify(59006)
	return &g.spatialObject
}

func (g *Geography) EWKBHex() string {
	__antithesis_instrumentation__.Notify(59007)
	return g.spatialObject.EWKBHex()
}

func (g *Geography) SRID() geopb.SRID {
	__antithesis_instrumentation__.Notify(59008)
	return g.spatialObject.SRID
}

func (g *Geography) ShapeType() geopb.ShapeType {
	__antithesis_instrumentation__.Notify(59009)
	return g.spatialObject.ShapeType
}

func (g *Geography) ShapeType2D() geopb.ShapeType {
	__antithesis_instrumentation__.Notify(59010)
	return g.ShapeType().To2D()
}

func (g *Geography) Spheroid() (*geographiclib.Spheroid, error) {
	__antithesis_instrumentation__.Notify(59011)
	proj, err := geoprojbase.Projection(g.SRID())
	if err != nil {
		__antithesis_instrumentation__.Notify(59013)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(59014)
	}
	__antithesis_instrumentation__.Notify(59012)
	return proj.Spheroid, nil
}

func (g *Geography) AsS2(emptyBehavior EmptyBehavior) ([]s2.Region, error) {
	__antithesis_instrumentation__.Notify(59015)
	geomRepr, err := g.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(59017)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(59018)
	}
	__antithesis_instrumentation__.Notify(59016)

	return S2RegionsFromGeomT(geomRepr, emptyBehavior)
}

func (g *Geography) BoundingRect() s2.Rect {
	__antithesis_instrumentation__.Notify(59019)
	bbox := g.spatialObject.BoundingBox
	if bbox == nil {
		__antithesis_instrumentation__.Notify(59021)
		return s2.EmptyRect()
	} else {
		__antithesis_instrumentation__.Notify(59022)
	}
	__antithesis_instrumentation__.Notify(59020)
	return s2.Rect{
		Lat: r1.Interval{Lo: bbox.LoY, Hi: bbox.HiY},
		Lng: s1.Interval{Lo: bbox.LoX, Hi: bbox.HiX},
	}
}

func (g *Geography) BoundingCap() s2.Cap {
	__antithesis_instrumentation__.Notify(59023)
	return g.BoundingRect().CapBound()
}

func (g *Geography) SpaceCurveIndex() uint64 {
	__antithesis_instrumentation__.Notify(59024)
	rect := g.BoundingRect()
	if rect.IsEmpty() {
		__antithesis_instrumentation__.Notify(59026)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(59027)
	}
	__antithesis_instrumentation__.Notify(59025)
	return uint64(s2.CellIDFromLatLng(rect.Center()))
}

func (g *Geography) Compare(o Geography) int {
	__antithesis_instrumentation__.Notify(59028)
	lhs := g.SpaceCurveIndex()
	rhs := o.SpaceCurveIndex()
	if lhs > rhs {
		__antithesis_instrumentation__.Notify(59031)
		return 1
	} else {
		__antithesis_instrumentation__.Notify(59032)
	}
	__antithesis_instrumentation__.Notify(59029)
	if lhs < rhs {
		__antithesis_instrumentation__.Notify(59033)
		return -1
	} else {
		__antithesis_instrumentation__.Notify(59034)
	}
	__antithesis_instrumentation__.Notify(59030)
	return compareSpatialObjectBytes(g.SpatialObjectRef(), o.SpatialObjectRef())
}

func AdjustGeomTSRID(t geom.T, srid geopb.SRID) {
	__antithesis_instrumentation__.Notify(59035)
	switch t := t.(type) {
	case *geom.Point:
		__antithesis_instrumentation__.Notify(59036)
		t.SetSRID(int(srid))
	case *geom.LineString:
		__antithesis_instrumentation__.Notify(59037)
		t.SetSRID(int(srid))
	case *geom.Polygon:
		__antithesis_instrumentation__.Notify(59038)
		t.SetSRID(int(srid))
	case *geom.GeometryCollection:
		__antithesis_instrumentation__.Notify(59039)
		t.SetSRID(int(srid))
	case *geom.MultiPoint:
		__antithesis_instrumentation__.Notify(59040)
		t.SetSRID(int(srid))
	case *geom.MultiLineString:
		__antithesis_instrumentation__.Notify(59041)
		t.SetSRID(int(srid))
	case *geom.MultiPolygon:
		__antithesis_instrumentation__.Notify(59042)
		t.SetSRID(int(srid))
	default:
		__antithesis_instrumentation__.Notify(59043)
		panic(errors.AssertionFailedf("geo: unknown geom type: %v", t))
	}
}

func IsLinearRingCCW(linearRing *geom.LinearRing) bool {
	__antithesis_instrumentation__.Notify(59044)
	smallestIdx := 0
	smallest := linearRing.Coord(0)

	for pointIdx := 1; pointIdx < linearRing.NumCoords()-1; pointIdx++ {
		__antithesis_instrumentation__.Notify(59050)
		curr := linearRing.Coord(pointIdx)
		if curr.Y() < smallest.Y() || func() bool {
			__antithesis_instrumentation__.Notify(59051)
			return (curr.Y() == smallest.Y() && func() bool {
				__antithesis_instrumentation__.Notify(59052)
				return curr.X() > smallest.X() == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(59053)
			smallestIdx = pointIdx
			smallest = curr
		} else {
			__antithesis_instrumentation__.Notify(59054)
		}
	}
	__antithesis_instrumentation__.Notify(59045)

	prevIdx := smallestIdx - 1
	if prevIdx < 0 {
		__antithesis_instrumentation__.Notify(59055)
		prevIdx = linearRing.NumCoords() - 1
	} else {
		__antithesis_instrumentation__.Notify(59056)
	}
	__antithesis_instrumentation__.Notify(59046)
	for prevIdx != smallestIdx {
		__antithesis_instrumentation__.Notify(59057)
		a := linearRing.Coord(prevIdx)
		if a.X() != smallest.X() || func() bool {
			__antithesis_instrumentation__.Notify(59059)
			return a.Y() != smallest.Y() == true
		}() == true {
			__antithesis_instrumentation__.Notify(59060)
			break
		} else {
			__antithesis_instrumentation__.Notify(59061)
		}
		__antithesis_instrumentation__.Notify(59058)
		prevIdx--
		if prevIdx < 0 {
			__antithesis_instrumentation__.Notify(59062)
			prevIdx = linearRing.NumCoords() - 1
		} else {
			__antithesis_instrumentation__.Notify(59063)
		}
	}
	__antithesis_instrumentation__.Notify(59047)

	nextIdx := smallestIdx + 1
	if nextIdx >= linearRing.NumCoords() {
		__antithesis_instrumentation__.Notify(59064)
		nextIdx = 0
	} else {
		__antithesis_instrumentation__.Notify(59065)
	}
	__antithesis_instrumentation__.Notify(59048)
	for nextIdx != smallestIdx {
		__antithesis_instrumentation__.Notify(59066)
		c := linearRing.Coord(nextIdx)
		if c.X() != smallest.X() || func() bool {
			__antithesis_instrumentation__.Notify(59068)
			return c.Y() != smallest.Y() == true
		}() == true {
			__antithesis_instrumentation__.Notify(59069)
			break
		} else {
			__antithesis_instrumentation__.Notify(59070)
		}
		__antithesis_instrumentation__.Notify(59067)
		nextIdx++
		if nextIdx >= linearRing.NumCoords() {
			__antithesis_instrumentation__.Notify(59071)
			nextIdx = 0
		} else {
			__antithesis_instrumentation__.Notify(59072)
		}
	}
	__antithesis_instrumentation__.Notify(59049)

	a := linearRing.Coord(prevIdx)
	b := smallest
	c := linearRing.Coord(nextIdx)

	areaSign := a.X()*b.Y() - a.Y()*b.X() +
		a.Y()*c.X() - a.X()*c.Y() +
		b.X()*c.Y() - c.X()*b.Y()

	return areaSign > 0
}

func S2RegionsFromGeomT(geomRepr geom.T, emptyBehavior EmptyBehavior) ([]s2.Region, error) {
	__antithesis_instrumentation__.Notify(59073)
	var regions []s2.Region
	if geomRepr.Empty() {
		__antithesis_instrumentation__.Notify(59076)
		switch emptyBehavior {
		case EmptyBehaviorOmit:
			__antithesis_instrumentation__.Notify(59077)
			return nil, nil
		case EmptyBehaviorError:
			__antithesis_instrumentation__.Notify(59078)
			return nil, NewEmptyGeometryError()
		default:
			__antithesis_instrumentation__.Notify(59079)
			return nil, errors.AssertionFailedf("programmer error: unknown behavior")
		}
	} else {
		__antithesis_instrumentation__.Notify(59080)
	}
	__antithesis_instrumentation__.Notify(59074)
	switch repr := geomRepr.(type) {
	case *geom.Point:
		__antithesis_instrumentation__.Notify(59081)
		regions = []s2.Region{
			s2.PointFromLatLng(s2.LatLngFromDegrees(repr.Y(), repr.X())),
		}
	case *geom.LineString:
		__antithesis_instrumentation__.Notify(59082)
		latLngs := make([]s2.LatLng, repr.NumCoords())
		for i := 0; i < repr.NumCoords(); i++ {
			__antithesis_instrumentation__.Notify(59090)
			p := repr.Coord(i)
			latLngs[i] = s2.LatLngFromDegrees(p.Y(), p.X())
		}
		__antithesis_instrumentation__.Notify(59083)
		regions = []s2.Region{
			s2.PolylineFromLatLngs(latLngs),
		}
	case *geom.Polygon:
		__antithesis_instrumentation__.Notify(59084)
		loops := make([]*s2.Loop, repr.NumLinearRings())

		for ringIdx := 0; ringIdx < repr.NumLinearRings(); ringIdx++ {
			__antithesis_instrumentation__.Notify(59091)
			linearRing := repr.LinearRing(ringIdx)
			points := make([]s2.Point, linearRing.NumCoords())
			isCCW := IsLinearRingCCW(linearRing)
			for pointIdx := 0; pointIdx < linearRing.NumCoords(); pointIdx++ {
				__antithesis_instrumentation__.Notify(59093)
				p := linearRing.Coord(pointIdx)
				pt := s2.PointFromLatLng(s2.LatLngFromDegrees(p.Y(), p.X()))
				if isCCW {
					__antithesis_instrumentation__.Notify(59094)
					points[pointIdx] = pt
				} else {
					__antithesis_instrumentation__.Notify(59095)
					points[len(points)-pointIdx-1] = pt
				}
			}
			__antithesis_instrumentation__.Notify(59092)
			loops[ringIdx] = s2.LoopFromPoints(points)
		}
		__antithesis_instrumentation__.Notify(59085)
		regions = []s2.Region{
			s2.PolygonFromLoops(loops),
		}
	case *geom.GeometryCollection:
		__antithesis_instrumentation__.Notify(59086)
		for _, geom := range repr.Geoms() {
			__antithesis_instrumentation__.Notify(59096)
			subRegions, err := S2RegionsFromGeomT(geom, emptyBehavior)
			if err != nil {
				__antithesis_instrumentation__.Notify(59098)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(59099)
			}
			__antithesis_instrumentation__.Notify(59097)
			regions = append(regions, subRegions...)
		}
	case *geom.MultiPoint:
		__antithesis_instrumentation__.Notify(59087)
		for i := 0; i < repr.NumPoints(); i++ {
			__antithesis_instrumentation__.Notify(59100)
			subRegions, err := S2RegionsFromGeomT(repr.Point(i), emptyBehavior)
			if err != nil {
				__antithesis_instrumentation__.Notify(59102)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(59103)
			}
			__antithesis_instrumentation__.Notify(59101)
			regions = append(regions, subRegions...)
		}
	case *geom.MultiLineString:
		__antithesis_instrumentation__.Notify(59088)
		for i := 0; i < repr.NumLineStrings(); i++ {
			__antithesis_instrumentation__.Notify(59104)
			subRegions, err := S2RegionsFromGeomT(repr.LineString(i), emptyBehavior)
			if err != nil {
				__antithesis_instrumentation__.Notify(59106)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(59107)
			}
			__antithesis_instrumentation__.Notify(59105)
			regions = append(regions, subRegions...)
		}
	case *geom.MultiPolygon:
		__antithesis_instrumentation__.Notify(59089)
		for i := 0; i < repr.NumPolygons(); i++ {
			__antithesis_instrumentation__.Notify(59108)
			subRegions, err := S2RegionsFromGeomT(repr.Polygon(i), emptyBehavior)
			if err != nil {
				__antithesis_instrumentation__.Notify(59110)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(59111)
			}
			__antithesis_instrumentation__.Notify(59109)
			regions = append(regions, subRegions...)
		}
	}
	__antithesis_instrumentation__.Notify(59075)
	return regions, nil
}

func normalizeLngLat(lng float64, lat float64) (float64, float64) {
	__antithesis_instrumentation__.Notify(59112)
	if lat > 90 || func() bool {
		__antithesis_instrumentation__.Notify(59115)
		return lat < -90 == true
	}() == true {
		__antithesis_instrumentation__.Notify(59116)
		lat = NormalizeLatitudeDegrees(lat)
	} else {
		__antithesis_instrumentation__.Notify(59117)
	}
	__antithesis_instrumentation__.Notify(59113)
	if lng > 180 || func() bool {
		__antithesis_instrumentation__.Notify(59118)
		return lng < -180 == true
	}() == true {
		__antithesis_instrumentation__.Notify(59119)
		lng = NormalizeLongitudeDegrees(lng)
	} else {
		__antithesis_instrumentation__.Notify(59120)
	}
	__antithesis_instrumentation__.Notify(59114)
	return lng, lat
}

func normalizeGeographyGeomT(t geom.T) {
	__antithesis_instrumentation__.Notify(59121)
	switch repr := t.(type) {
	case *geom.GeometryCollection:
		__antithesis_instrumentation__.Notify(59122)
		for _, geom := range repr.Geoms() {
			__antithesis_instrumentation__.Notify(59124)
			normalizeGeographyGeomT(geom)
		}
	default:
		__antithesis_instrumentation__.Notify(59123)
		coords := repr.FlatCoords()
		for i := 0; i < len(coords); i += repr.Stride() {
			__antithesis_instrumentation__.Notify(59125)
			coords[i], coords[i+1] = normalizeLngLat(coords[i], coords[i+1])
		}
	}
}

func validateGeomT(t geom.T) error {
	__antithesis_instrumentation__.Notify(59126)
	if t.Empty() {
		__antithesis_instrumentation__.Notify(59129)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(59130)
	}
	__antithesis_instrumentation__.Notify(59127)
	switch t := t.(type) {
	case *geom.Point:
		__antithesis_instrumentation__.Notify(59131)
	case *geom.LineString:
		__antithesis_instrumentation__.Notify(59132)
		if t.NumCoords() < 2 {
			__antithesis_instrumentation__.Notify(59139)
			return pgerror.Newf(
				pgcode.InvalidParameterValue,
				"LineString must have at least 2 coordinates",
			)
		} else {
			__antithesis_instrumentation__.Notify(59140)
		}
	case *geom.Polygon:
		__antithesis_instrumentation__.Notify(59133)
		for i := 0; i < t.NumLinearRings(); i++ {
			__antithesis_instrumentation__.Notify(59141)
			linearRing := t.LinearRing(i)
			if linearRing.NumCoords() < 4 {
				__antithesis_instrumentation__.Notify(59143)
				return pgerror.Newf(
					pgcode.InvalidParameterValue,
					"Polygon LinearRing must have at least 4 points, found %d at position %d",
					linearRing.NumCoords(),
					i+1,
				)
			} else {
				__antithesis_instrumentation__.Notify(59144)
			}
			__antithesis_instrumentation__.Notify(59142)
			if !linearRing.Coord(0).Equal(linearRing.Layout(), linearRing.Coord(linearRing.NumCoords()-1)) {
				__antithesis_instrumentation__.Notify(59145)
				return pgerror.Newf(
					pgcode.InvalidParameterValue,
					"Polygon LinearRing at position %d is not closed",
					i+1,
				)
			} else {
				__antithesis_instrumentation__.Notify(59146)
			}
		}
	case *geom.MultiPoint:
		__antithesis_instrumentation__.Notify(59134)
	case *geom.MultiLineString:
		__antithesis_instrumentation__.Notify(59135)
		for i := 0; i < t.NumLineStrings(); i++ {
			__antithesis_instrumentation__.Notify(59147)
			if err := validateGeomT(t.LineString(i)); err != nil {
				__antithesis_instrumentation__.Notify(59148)
				return errors.Wrapf(err, "invalid MultiLineString component at position %d", i+1)
			} else {
				__antithesis_instrumentation__.Notify(59149)
			}
		}
	case *geom.MultiPolygon:
		__antithesis_instrumentation__.Notify(59136)
		for i := 0; i < t.NumPolygons(); i++ {
			__antithesis_instrumentation__.Notify(59150)
			if err := validateGeomT(t.Polygon(i)); err != nil {
				__antithesis_instrumentation__.Notify(59151)
				return errors.Wrapf(err, "invalid MultiPolygon component at position %d", i+1)
			} else {
				__antithesis_instrumentation__.Notify(59152)
			}
		}
	case *geom.GeometryCollection:
		__antithesis_instrumentation__.Notify(59137)

		for i := 0; i < t.NumGeoms(); i++ {
			__antithesis_instrumentation__.Notify(59153)
			if err := validateGeomT(t.Geom(i)); err != nil {
				__antithesis_instrumentation__.Notify(59154)
				return errors.Wrapf(err, "invalid GeometryCollection component at position %d", i+1)
			} else {
				__antithesis_instrumentation__.Notify(59155)
			}
		}
	default:
		__antithesis_instrumentation__.Notify(59138)
		return pgerror.Newf(
			pgcode.InvalidParameterValue,
			"unknown geom.T type: %T",
			t,
		)
	}
	__antithesis_instrumentation__.Notify(59128)
	return nil
}

func spatialObjectFromGeomT(t geom.T, soType geopb.SpatialObjectType) (geopb.SpatialObject, error) {
	__antithesis_instrumentation__.Notify(59156)
	if err := validateGeomT(t); err != nil {
		__antithesis_instrumentation__.Notify(59162)
		return geopb.SpatialObject{}, err
	} else {
		__antithesis_instrumentation__.Notify(59163)
	}
	__antithesis_instrumentation__.Notify(59157)
	if soType == geopb.SpatialObjectType_GeographyType {
		__antithesis_instrumentation__.Notify(59164)
		normalizeGeographyGeomT(t)
	} else {
		__antithesis_instrumentation__.Notify(59165)
	}
	__antithesis_instrumentation__.Notify(59158)
	ret, err := ewkb.Marshal(t, DefaultEWKBEncodingFormat)
	if err != nil {
		__antithesis_instrumentation__.Notify(59166)
		return geopb.SpatialObject{}, err
	} else {
		__antithesis_instrumentation__.Notify(59167)
	}
	__antithesis_instrumentation__.Notify(59159)
	shapeType, err := shapeTypeFromGeomT(t)
	if err != nil {
		__antithesis_instrumentation__.Notify(59168)
		return geopb.SpatialObject{}, err
	} else {
		__antithesis_instrumentation__.Notify(59169)
	}
	__antithesis_instrumentation__.Notify(59160)
	bbox, err := boundingBoxFromGeomT(t, soType)
	if err != nil {
		__antithesis_instrumentation__.Notify(59170)
		return geopb.SpatialObject{}, err
	} else {
		__antithesis_instrumentation__.Notify(59171)
	}
	__antithesis_instrumentation__.Notify(59161)
	return geopb.SpatialObject{
		Type:        soType,
		EWKB:        geopb.EWKB(ret),
		SRID:        geopb.SRID(t.SRID()),
		ShapeType:   shapeType,
		BoundingBox: bbox,
	}, nil
}

func shapeTypeFromGeomT(t geom.T) (geopb.ShapeType, error) {
	__antithesis_instrumentation__.Notify(59172)
	var shapeType geopb.ShapeType
	switch t := t.(type) {
	case *geom.Point:
		__antithesis_instrumentation__.Notify(59175)
		shapeType = geopb.ShapeType_Point
	case *geom.LineString:
		__antithesis_instrumentation__.Notify(59176)
		shapeType = geopb.ShapeType_LineString
	case *geom.Polygon:
		__antithesis_instrumentation__.Notify(59177)
		shapeType = geopb.ShapeType_Polygon
	case *geom.MultiPoint:
		__antithesis_instrumentation__.Notify(59178)
		shapeType = geopb.ShapeType_MultiPoint
	case *geom.MultiLineString:
		__antithesis_instrumentation__.Notify(59179)
		shapeType = geopb.ShapeType_MultiLineString
	case *geom.MultiPolygon:
		__antithesis_instrumentation__.Notify(59180)
		shapeType = geopb.ShapeType_MultiPolygon
	case *geom.GeometryCollection:
		__antithesis_instrumentation__.Notify(59181)
		shapeType = geopb.ShapeType_GeometryCollection
	default:
		__antithesis_instrumentation__.Notify(59182)
		return geopb.ShapeType_Unset, pgerror.Newf(pgcode.InvalidParameterValue, "unknown shape: %T", t)
	}
	__antithesis_instrumentation__.Notify(59173)
	switch t.Layout() {
	case geom.NoLayout:
		__antithesis_instrumentation__.Notify(59183)
		if gc, ok := t.(*geom.GeometryCollection); !ok || func() bool {
			__antithesis_instrumentation__.Notify(59189)
			return !gc.Empty() == true
		}() == true {
			__antithesis_instrumentation__.Notify(59190)
			return geopb.ShapeType_Unset, pgerror.Newf(pgcode.InvalidParameterValue, "no layout found on object")
		} else {
			__antithesis_instrumentation__.Notify(59191)
		}
	case geom.XY:
		__antithesis_instrumentation__.Notify(59184)
		break
	case geom.XYM:
		__antithesis_instrumentation__.Notify(59185)
		shapeType = shapeType | geopb.MShapeTypeFlag
	case geom.XYZ:
		__antithesis_instrumentation__.Notify(59186)
		shapeType = shapeType | geopb.ZShapeTypeFlag
	case geom.XYZM:
		__antithesis_instrumentation__.Notify(59187)
		shapeType = shapeType | geopb.ZShapeTypeFlag | geopb.MShapeTypeFlag
	default:
		__antithesis_instrumentation__.Notify(59188)
		return geopb.ShapeType_Unset, pgerror.Newf(pgcode.InvalidParameterValue, "unknown layout: %s", t.Layout())
	}
	__antithesis_instrumentation__.Notify(59174)
	return shapeType, nil
}

func GeomTContainsEmpty(g geom.T) bool {
	__antithesis_instrumentation__.Notify(59192)
	if g.Empty() {
		__antithesis_instrumentation__.Notify(59195)
		return true
	} else {
		__antithesis_instrumentation__.Notify(59196)
	}
	__antithesis_instrumentation__.Notify(59193)
	switch g := g.(type) {
	case *geom.MultiPoint:
		__antithesis_instrumentation__.Notify(59197)
		for i := 0; i < g.NumPoints(); i++ {
			__antithesis_instrumentation__.Notify(59201)
			if g.Point(i).Empty() {
				__antithesis_instrumentation__.Notify(59202)
				return true
			} else {
				__antithesis_instrumentation__.Notify(59203)
			}
		}
	case *geom.MultiLineString:
		__antithesis_instrumentation__.Notify(59198)
		for i := 0; i < g.NumLineStrings(); i++ {
			__antithesis_instrumentation__.Notify(59204)
			if g.LineString(i).Empty() {
				__antithesis_instrumentation__.Notify(59205)
				return true
			} else {
				__antithesis_instrumentation__.Notify(59206)
			}
		}
	case *geom.MultiPolygon:
		__antithesis_instrumentation__.Notify(59199)
		for i := 0; i < g.NumPolygons(); i++ {
			__antithesis_instrumentation__.Notify(59207)
			if g.Polygon(i).Empty() {
				__antithesis_instrumentation__.Notify(59208)
				return true
			} else {
				__antithesis_instrumentation__.Notify(59209)
			}
		}
	case *geom.GeometryCollection:
		__antithesis_instrumentation__.Notify(59200)
		for i := 0; i < g.NumGeoms(); i++ {
			__antithesis_instrumentation__.Notify(59210)
			if GeomTContainsEmpty(g.Geom(i)) {
				__antithesis_instrumentation__.Notify(59211)
				return true
			} else {
				__antithesis_instrumentation__.Notify(59212)
			}
		}
	}
	__antithesis_instrumentation__.Notify(59194)
	return false
}

func compareSpatialObjectBytes(lhs *geopb.SpatialObject, rhs *geopb.SpatialObject) int {
	__antithesis_instrumentation__.Notify(59213)
	marshalledLHS, err := protoutil.Marshal(lhs)
	if err != nil {
		__antithesis_instrumentation__.Notify(59216)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(59217)
	}
	__antithesis_instrumentation__.Notify(59214)
	marshalledRHS, err := protoutil.Marshal(rhs)
	if err != nil {
		__antithesis_instrumentation__.Notify(59218)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(59219)
	}
	__antithesis_instrumentation__.Notify(59215)
	return bytes.Compare(marshalledLHS, marshalledRHS)
}
