package geoindex

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math"

	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geomfn"
	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/geo/geoprojbase"
	"github.com/cockroachdb/cockroach/pkg/geo/geos"
	"github.com/cockroachdb/errors"
	"github.com/golang/geo/r3"
	"github.com/golang/geo/s2"
	"github.com/twpayne/go-geom"
)

type s2GeometryIndex struct {
	rc                     *s2.RegionCoverer
	minX, maxX, minY, maxY float64
	deltaX, deltaY         float64
}

var _ GeometryIndex = (*s2GeometryIndex)(nil)

const clippingBoundsDelta = 0.01

func NewS2GeometryIndex(cfg S2GeometryConfig) GeometryIndex {
	__antithesis_instrumentation__.Notify(60788)

	return &s2GeometryIndex{
		rc: &s2.RegionCoverer{
			MinLevel: int(cfg.S2Config.MinLevel),
			MaxLevel: int(cfg.S2Config.MaxLevel),
			LevelMod: int(cfg.S2Config.LevelMod),
			MaxCells: int(cfg.S2Config.MaxCells),
		},
		minX:   cfg.MinX,
		maxX:   cfg.MaxX,
		minY:   cfg.MinY,
		maxY:   cfg.MaxY,
		deltaX: clippingBoundsDelta * (cfg.MaxX - cfg.MinX),
		deltaY: clippingBoundsDelta * (cfg.MaxY - cfg.MinY),
	}
}

func DefaultGeometryIndexConfig() *Config {
	__antithesis_instrumentation__.Notify(60789)
	return &Config{
		S2Geometry: &S2GeometryConfig{

			MinX:     -(1 << 31),
			MaxX:     (1 << 31) - 1,
			MinY:     -(1 << 31),
			MaxY:     (1 << 31) - 1,
			S2Config: DefaultS2Config(),
		},
	}
}

func GeometryIndexConfigForSRID(srid geopb.SRID) (*Config, error) {
	__antithesis_instrumentation__.Notify(60790)
	if srid == 0 {
		__antithesis_instrumentation__.Notify(60796)
		return DefaultGeometryIndexConfig(), nil
	} else {
		__antithesis_instrumentation__.Notify(60797)
	}
	__antithesis_instrumentation__.Notify(60791)
	p, err := geoprojbase.Projection(srid)
	if err != nil {
		__antithesis_instrumentation__.Notify(60798)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(60799)
	}
	__antithesis_instrumentation__.Notify(60792)
	b := p.Bounds
	minX, maxX, minY, maxY := b.MinX, b.MaxX, b.MinY, b.MaxY

	if maxX-minX < 1 {
		__antithesis_instrumentation__.Notify(60800)
		maxX++
	} else {
		__antithesis_instrumentation__.Notify(60801)
	}
	__antithesis_instrumentation__.Notify(60793)
	if maxY-minY < 1 {
		__antithesis_instrumentation__.Notify(60802)
		maxY++
	} else {
		__antithesis_instrumentation__.Notify(60803)
	}
	__antithesis_instrumentation__.Notify(60794)

	diffX := maxX - minX
	diffY := maxY - minY
	if diffX > diffY {
		__antithesis_instrumentation__.Notify(60804)
		adjustment := (diffX - diffY) / 2
		minY -= adjustment
		maxY += adjustment
	} else {
		__antithesis_instrumentation__.Notify(60805)
		adjustment := (diffY - diffX) / 2
		minX -= adjustment
		maxX += adjustment
	}
	__antithesis_instrumentation__.Notify(60795)

	boundsExpansion := 2 * clippingBoundsDelta
	deltaX := (maxX - minX) * boundsExpansion
	deltaY := (maxY - minY) * boundsExpansion
	return &Config{
		S2Geometry: &S2GeometryConfig{
			MinX:     minX - deltaX,
			MaxX:     maxX + deltaX,
			MinY:     minY - deltaY,
			MaxY:     maxY + deltaY,
			S2Config: DefaultS2Config()},
	}, nil
}

const exceedsBoundsCellID = s2.CellID(^uint64(0))

type geomCovererWithBBoxFallback struct {
	s    *s2GeometryIndex
	geom geom.T
}

var _ covererInterface = geomCovererWithBBoxFallback{}

func (rc geomCovererWithBBoxFallback) covering(regions []s2.Region) s2.CellUnion {
	__antithesis_instrumentation__.Notify(60806)
	cu := simpleCovererImpl{rc: rc.s.rc}.covering(regions)
	if isBadGeomCovering(cu) {
		__antithesis_instrumentation__.Notify(60808)
		bbox := geo.BoundingBoxFromGeomTGeometryType(rc.geom)
		flatCoords := []float64{
			bbox.LoX, bbox.LoY, bbox.HiX, bbox.LoY, bbox.HiX, bbox.HiY, bbox.LoX, bbox.HiY,
			bbox.LoX, bbox.LoY}
		bboxT := geom.NewPolygonFlat(geom.XY, flatCoords, []int{len(flatCoords)})
		bboxRegions := rc.s.s2RegionsFromPlanarGeomT(bboxT)
		bboxCU := simpleCovererImpl{rc: rc.s.rc}.covering(bboxRegions)
		if !isBadGeomCovering(bboxCU) {
			__antithesis_instrumentation__.Notify(60809)
			cu = bboxCU
		} else {
			__antithesis_instrumentation__.Notify(60810)
		}
	} else {
		__antithesis_instrumentation__.Notify(60811)
	}
	__antithesis_instrumentation__.Notify(60807)
	return cu
}

func isBadGeomCovering(cu s2.CellUnion) bool {
	__antithesis_instrumentation__.Notify(60812)
	for _, c := range cu {
		__antithesis_instrumentation__.Notify(60814)
		if c.Face() != 0 {
			__antithesis_instrumentation__.Notify(60815)

			return true
		} else {
			__antithesis_instrumentation__.Notify(60816)
		}
	}
	__antithesis_instrumentation__.Notify(60813)
	return false
}

func (s *s2GeometryIndex) InvertedIndexKeys(
	c context.Context, g geo.Geometry,
) ([]Key, geopb.BoundingBox, error) {
	__antithesis_instrumentation__.Notify(60817)

	gt, clipped, err := s.convertToGeomTAndTryClip(g)
	if err != nil {
		__antithesis_instrumentation__.Notify(60823)
		return nil, geopb.BoundingBox{}, err
	} else {
		__antithesis_instrumentation__.Notify(60824)
	}
	__antithesis_instrumentation__.Notify(60818)
	var keys []Key
	if gt != nil {
		__antithesis_instrumentation__.Notify(60825)
		r := s.s2RegionsFromPlanarGeomT(gt)
		keys = invertedIndexKeys(c, geomCovererWithBBoxFallback{s: s, geom: gt}, r)
	} else {
		__antithesis_instrumentation__.Notify(60826)
	}
	__antithesis_instrumentation__.Notify(60819)
	if clipped {
		__antithesis_instrumentation__.Notify(60827)
		keys = append(keys, Key(exceedsBoundsCellID))
	} else {
		__antithesis_instrumentation__.Notify(60828)
	}
	__antithesis_instrumentation__.Notify(60820)
	bbox := geopb.BoundingBox{}
	bboxRef := g.BoundingBoxRef()
	if bboxRef == nil && func() bool {
		__antithesis_instrumentation__.Notify(60829)
		return len(keys) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(60830)
		return keys, bbox, errors.AssertionFailedf("non-empty geometry should have bounding box")
	} else {
		__antithesis_instrumentation__.Notify(60831)
	}
	__antithesis_instrumentation__.Notify(60821)
	if bboxRef != nil {
		__antithesis_instrumentation__.Notify(60832)
		bbox = *bboxRef
	} else {
		__antithesis_instrumentation__.Notify(60833)
	}
	__antithesis_instrumentation__.Notify(60822)
	return keys, bbox, nil
}

func (s *s2GeometryIndex) Covers(c context.Context, g geo.Geometry) (UnionKeySpans, error) {
	__antithesis_instrumentation__.Notify(60834)
	return s.Intersects(c, g)
}

func (s *s2GeometryIndex) CoveredBy(c context.Context, g geo.Geometry) (RPKeyExpr, error) {
	__antithesis_instrumentation__.Notify(60835)

	gt, clipped, err := s.convertToGeomTAndTryClip(g)
	if err != nil {
		__antithesis_instrumentation__.Notify(60839)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(60840)
	}
	__antithesis_instrumentation__.Notify(60836)
	var expr RPKeyExpr
	if gt != nil {
		__antithesis_instrumentation__.Notify(60841)
		r := s.s2RegionsFromPlanarGeomT(gt)
		expr = coveredBy(c, s.rc, r)
	} else {
		__antithesis_instrumentation__.Notify(60842)
	}
	__antithesis_instrumentation__.Notify(60837)
	if clipped {
		__antithesis_instrumentation__.Notify(60843)

		expr = append(expr, Key(exceedsBoundsCellID))
		if len(expr) > 1 {
			__antithesis_instrumentation__.Notify(60844)
			expr = append(expr, RPSetIntersection)
		} else {
			__antithesis_instrumentation__.Notify(60845)
		}
	} else {
		__antithesis_instrumentation__.Notify(60846)
	}
	__antithesis_instrumentation__.Notify(60838)
	return expr, nil
}

func (s *s2GeometryIndex) Intersects(c context.Context, g geo.Geometry) (UnionKeySpans, error) {
	__antithesis_instrumentation__.Notify(60847)

	gt, clipped, err := s.convertToGeomTAndTryClip(g)
	if err != nil {
		__antithesis_instrumentation__.Notify(60851)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(60852)
	}
	__antithesis_instrumentation__.Notify(60848)
	var spans UnionKeySpans
	if gt != nil {
		__antithesis_instrumentation__.Notify(60853)
		r := s.s2RegionsFromPlanarGeomT(gt)
		spans = intersects(c, geomCovererWithBBoxFallback{s: s, geom: gt}, r)
	} else {
		__antithesis_instrumentation__.Notify(60854)
	}
	__antithesis_instrumentation__.Notify(60849)
	if clipped {
		__antithesis_instrumentation__.Notify(60855)

		spans = append(spans, KeySpan{Start: Key(exceedsBoundsCellID), End: Key(exceedsBoundsCellID)})
	} else {
		__antithesis_instrumentation__.Notify(60856)
	}
	__antithesis_instrumentation__.Notify(60850)
	return spans, nil
}

func (s *s2GeometryIndex) DWithin(
	c context.Context, g geo.Geometry, distance float64,
) (UnionKeySpans, error) {
	__antithesis_instrumentation__.Notify(60857)

	g, err := geomfn.Buffer(g, geomfn.MakeDefaultBufferParams(), distance)
	if err != nil {
		__antithesis_instrumentation__.Notify(60859)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(60860)
	}
	__antithesis_instrumentation__.Notify(60858)
	return s.Intersects(c, g)
}

func (s *s2GeometryIndex) DFullyWithin(
	c context.Context, g geo.Geometry, distance float64,
) (UnionKeySpans, error) {
	__antithesis_instrumentation__.Notify(60861)

	g, err := geomfn.Buffer(g, geomfn.MakeDefaultBufferParams(), distance)
	if err != nil {
		__antithesis_instrumentation__.Notify(60863)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(60864)
	}
	__antithesis_instrumentation__.Notify(60862)
	return s.Covers(c, g)
}

func (s *s2GeometryIndex) convertToGeomTAndTryClip(g geo.Geometry) (geom.T, bool, error) {
	__antithesis_instrumentation__.Notify(60865)
	gt, err := g.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(60869)
		return nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(60870)
	}
	__antithesis_instrumentation__.Notify(60866)
	if gt.Empty() {
		__antithesis_instrumentation__.Notify(60871)
		return gt, false, nil
	} else {
		__antithesis_instrumentation__.Notify(60872)
	}
	__antithesis_instrumentation__.Notify(60867)
	clipped := false
	if s.geomExceedsBounds(gt) {
		__antithesis_instrumentation__.Notify(60873)

		g, err = geomfn.ForceLayout(g, geom.XY)
		if err != nil {
			__antithesis_instrumentation__.Notify(60876)
			return nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(60877)
		}
		__antithesis_instrumentation__.Notify(60874)
		clipped = true
		clippedEWKB, err :=
			geos.ClipByRect(g.EWKB(), s.minX+s.deltaX, s.minY+s.deltaY, s.maxX-s.deltaX, s.maxY-s.deltaY)
		if err != nil {
			__antithesis_instrumentation__.Notify(60878)
			return nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(60879)
		}
		__antithesis_instrumentation__.Notify(60875)
		gt = nil
		if clippedEWKB != nil {
			__antithesis_instrumentation__.Notify(60880)
			g, err = geo.ParseGeometryFromEWKBUnsafe(clippedEWKB)
			if err != nil {
				__antithesis_instrumentation__.Notify(60882)
				return nil, false, err
			} else {
				__antithesis_instrumentation__.Notify(60883)
			}
			__antithesis_instrumentation__.Notify(60881)
			gt, err = g.AsGeomT()
			if err != nil {
				__antithesis_instrumentation__.Notify(60884)
				return nil, false, err
			} else {
				__antithesis_instrumentation__.Notify(60885)
			}
		} else {
			__antithesis_instrumentation__.Notify(60886)
		}
	} else {
		__antithesis_instrumentation__.Notify(60887)
	}
	__antithesis_instrumentation__.Notify(60868)
	return gt, clipped, nil
}

func (s *s2GeometryIndex) xyExceedsBounds(x float64, y float64) bool {
	__antithesis_instrumentation__.Notify(60888)
	if x < (s.minX+s.deltaX) || func() bool {
		__antithesis_instrumentation__.Notify(60891)
		return x > (s.maxX - s.deltaX) == true
	}() == true {
		__antithesis_instrumentation__.Notify(60892)
		return true
	} else {
		__antithesis_instrumentation__.Notify(60893)
	}
	__antithesis_instrumentation__.Notify(60889)
	if y < (s.minY+s.deltaY) || func() bool {
		__antithesis_instrumentation__.Notify(60894)
		return y > (s.maxY - s.deltaY) == true
	}() == true {
		__antithesis_instrumentation__.Notify(60895)
		return true
	} else {
		__antithesis_instrumentation__.Notify(60896)
	}
	__antithesis_instrumentation__.Notify(60890)
	return false
}

func (s *s2GeometryIndex) geomExceedsBounds(g geom.T) bool {
	__antithesis_instrumentation__.Notify(60897)
	switch repr := g.(type) {
	case *geom.Point:
		__antithesis_instrumentation__.Notify(60899)
		return s.xyExceedsBounds(repr.X(), repr.Y())
	case *geom.LineString:
		__antithesis_instrumentation__.Notify(60900)
		for i := 0; i < repr.NumCoords(); i++ {
			__antithesis_instrumentation__.Notify(60906)
			p := repr.Coord(i)
			if s.xyExceedsBounds(p.X(), p.Y()) {
				__antithesis_instrumentation__.Notify(60907)
				return true
			} else {
				__antithesis_instrumentation__.Notify(60908)
			}
		}
	case *geom.Polygon:
		__antithesis_instrumentation__.Notify(60901)
		if repr.NumLinearRings() > 0 {
			__antithesis_instrumentation__.Notify(60909)
			lr := repr.LinearRing(0)
			for i := 0; i < lr.NumCoords(); i++ {
				__antithesis_instrumentation__.Notify(60910)
				if s.xyExceedsBounds(lr.Coord(i).X(), lr.Coord(i).Y()) {
					__antithesis_instrumentation__.Notify(60911)
					return true
				} else {
					__antithesis_instrumentation__.Notify(60912)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(60913)
		}
	case *geom.GeometryCollection:
		__antithesis_instrumentation__.Notify(60902)
		for _, geom := range repr.Geoms() {
			__antithesis_instrumentation__.Notify(60914)
			if s.geomExceedsBounds(geom) {
				__antithesis_instrumentation__.Notify(60915)
				return true
			} else {
				__antithesis_instrumentation__.Notify(60916)
			}
		}
	case *geom.MultiPoint:
		__antithesis_instrumentation__.Notify(60903)
		for i := 0; i < repr.NumPoints(); i++ {
			__antithesis_instrumentation__.Notify(60917)
			if s.geomExceedsBounds(repr.Point(i)) {
				__antithesis_instrumentation__.Notify(60918)
				return true
			} else {
				__antithesis_instrumentation__.Notify(60919)
			}
		}
	case *geom.MultiLineString:
		__antithesis_instrumentation__.Notify(60904)
		for i := 0; i < repr.NumLineStrings(); i++ {
			__antithesis_instrumentation__.Notify(60920)
			if s.geomExceedsBounds(repr.LineString(i)) {
				__antithesis_instrumentation__.Notify(60921)
				return true
			} else {
				__antithesis_instrumentation__.Notify(60922)
			}
		}
	case *geom.MultiPolygon:
		__antithesis_instrumentation__.Notify(60905)
		for i := 0; i < repr.NumPolygons(); i++ {
			__antithesis_instrumentation__.Notify(60923)
			if s.geomExceedsBounds(repr.Polygon(i)) {
				__antithesis_instrumentation__.Notify(60924)
				return true
			} else {
				__antithesis_instrumentation__.Notify(60925)
			}
		}
	}
	__antithesis_instrumentation__.Notify(60898)
	return false
}

func stToUV(s float64) float64 {
	__antithesis_instrumentation__.Notify(60926)
	if s >= 0.5 {
		__antithesis_instrumentation__.Notify(60928)
		return (1 / 3.) * (4*s*s - 1)
	} else {
		__antithesis_instrumentation__.Notify(60929)
	}
	__antithesis_instrumentation__.Notify(60927)
	return (1 / 3.) * (1 - 4*(1-s)*(1-s))
}

func uvToST(u float64) float64 {
	__antithesis_instrumentation__.Notify(60930)
	if u >= 0 {
		__antithesis_instrumentation__.Notify(60932)
		return 0.5 * math.Sqrt(1+3*u)
	} else {
		__antithesis_instrumentation__.Notify(60933)
	}
	__antithesis_instrumentation__.Notify(60931)
	return 1 - 0.5*math.Sqrt(1-3*u)
}

func face0UVToXYZPoint(u, v float64) s2.Point {
	__antithesis_instrumentation__.Notify(60934)
	return s2.Point{Vector: r3.Vector{X: 1, Y: u, Z: v}}
}

func xyzToFace0UV(r s2.Point) (u, v float64) {
	__antithesis_instrumentation__.Notify(60935)
	return r.Y / r.X, r.Z / r.X
}

func (s *s2GeometryIndex) planarPointToS2Point(x float64, y float64) s2.Point {
	__antithesis_instrumentation__.Notify(60936)
	ss := (x - s.minX) / (s.maxX - s.minX)
	tt := (y - s.minY) / (s.maxY - s.minY)
	u := stToUV(ss)
	v := stToUV(tt)
	return face0UVToXYZPoint(u, v)
}

func (s *s2GeometryIndex) s2PointToPlanarPoint(p s2.Point) (x, y float64) {
	__antithesis_instrumentation__.Notify(60937)
	u, v := xyzToFace0UV(p)
	ss := uvToST(u)
	tt := uvToST(v)
	return ss*(s.maxX-s.minX) + s.minX, tt*(s.maxY-s.minY) + s.minY
}

func (s *s2GeometryIndex) s2RegionsFromPlanarGeomT(geomRepr geom.T) []s2.Region {
	__antithesis_instrumentation__.Notify(60938)
	if geomRepr.Empty() {
		__antithesis_instrumentation__.Notify(60941)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(60942)
	}
	__antithesis_instrumentation__.Notify(60939)
	var regions []s2.Region
	switch repr := geomRepr.(type) {
	case *geom.Point:
		__antithesis_instrumentation__.Notify(60943)
		regions = []s2.Region{
			s.planarPointToS2Point(repr.X(), repr.Y()),
		}
	case *geom.LineString:
		__antithesis_instrumentation__.Notify(60944)
		points := make([]s2.Point, repr.NumCoords())
		for i := 0; i < repr.NumCoords(); i++ {
			__antithesis_instrumentation__.Notify(60952)
			p := repr.Coord(i)
			points[i] = s.planarPointToS2Point(p.X(), p.Y())
		}
		__antithesis_instrumentation__.Notify(60945)
		pl := s2.Polyline(points)
		regions = []s2.Region{&pl}
	case *geom.Polygon:
		__antithesis_instrumentation__.Notify(60946)
		loops := make([]*s2.Loop, repr.NumLinearRings())

		for ringIdx := 0; ringIdx < repr.NumLinearRings(); ringIdx++ {
			__antithesis_instrumentation__.Notify(60953)
			linearRing := repr.LinearRing(ringIdx)
			points := make([]s2.Point, linearRing.NumCoords())
			isCCW := geo.IsLinearRingCCW(linearRing)
			for pointIdx := 0; pointIdx < linearRing.NumCoords(); pointIdx++ {
				__antithesis_instrumentation__.Notify(60955)
				p := linearRing.Coord(pointIdx)
				pt := s.planarPointToS2Point(p.X(), p.Y())
				if isCCW {
					__antithesis_instrumentation__.Notify(60956)
					points[pointIdx] = pt
				} else {
					__antithesis_instrumentation__.Notify(60957)
					points[len(points)-pointIdx-1] = pt
				}
			}
			__antithesis_instrumentation__.Notify(60954)
			loops[ringIdx] = s2.LoopFromPoints(points)
		}
		__antithesis_instrumentation__.Notify(60947)
		regions = []s2.Region{
			s2.PolygonFromLoops(loops),
		}
	case *geom.GeometryCollection:
		__antithesis_instrumentation__.Notify(60948)
		for _, geom := range repr.Geoms() {
			__antithesis_instrumentation__.Notify(60958)
			regions = append(regions, s.s2RegionsFromPlanarGeomT(geom)...)
		}
	case *geom.MultiPoint:
		__antithesis_instrumentation__.Notify(60949)
		for i := 0; i < repr.NumPoints(); i++ {
			__antithesis_instrumentation__.Notify(60959)
			regions = append(regions, s.s2RegionsFromPlanarGeomT(repr.Point(i))...)
		}
	case *geom.MultiLineString:
		__antithesis_instrumentation__.Notify(60950)
		for i := 0; i < repr.NumLineStrings(); i++ {
			__antithesis_instrumentation__.Notify(60960)
			regions = append(regions, s.s2RegionsFromPlanarGeomT(repr.LineString(i))...)
		}
	case *geom.MultiPolygon:
		__antithesis_instrumentation__.Notify(60951)
		for i := 0; i < repr.NumPolygons(); i++ {
			__antithesis_instrumentation__.Notify(60961)
			regions = append(regions, s.s2RegionsFromPlanarGeomT(repr.Polygon(i))...)
		}
	}
	__antithesis_instrumentation__.Notify(60940)
	return regions
}

func (s *s2GeometryIndex) TestingInnerCovering(g geo.Geometry) s2.CellUnion {
	__antithesis_instrumentation__.Notify(60962)
	gt, _, err := s.convertToGeomTAndTryClip(g)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(60964)
		return gt == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(60965)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(60966)
	}
	__antithesis_instrumentation__.Notify(60963)
	r := s.s2RegionsFromPlanarGeomT(gt)
	return innerCovering(s.rc, r)
}

func (s *s2GeometryIndex) CoveringGeometry(
	c context.Context, g geo.Geometry,
) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(60967)
	keys, _, err := s.InvertedIndexKeys(c, g)
	if err != nil {
		__antithesis_instrumentation__.Notify(60971)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(60972)
	}
	__antithesis_instrumentation__.Notify(60968)
	t, err := makeGeomTFromKeys(keys, g.SRID(), func(p s2.Point) (float64, float64) {
		__antithesis_instrumentation__.Notify(60973)
		return s.s2PointToPlanarPoint(p)
	})
	__antithesis_instrumentation__.Notify(60969)
	if err != nil {
		__antithesis_instrumentation__.Notify(60974)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(60975)
	}
	__antithesis_instrumentation__.Notify(60970)
	return geo.MakeGeometryFromGeomT(t)
}
