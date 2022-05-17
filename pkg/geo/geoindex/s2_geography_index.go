package geoindex

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geogfn"
	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/geo/geoprojbase"
	"github.com/cockroachdb/errors"
	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
	"github.com/twpayne/go-geom"
)

type s2GeographyIndex struct {
	rc *s2.RegionCoverer
}

var _ GeographyIndex = (*s2GeographyIndex)(nil)

func NewS2GeographyIndex(cfg S2GeographyConfig) GeographyIndex {
	__antithesis_instrumentation__.Notify(60718)

	return &s2GeographyIndex{
		rc: &s2.RegionCoverer{
			MinLevel: int(cfg.S2Config.MinLevel),
			MaxLevel: int(cfg.S2Config.MaxLevel),
			LevelMod: int(cfg.S2Config.LevelMod),
			MaxCells: int(cfg.S2Config.MaxCells),
		},
	}
}

func DefaultGeographyIndexConfig() *Config {
	__antithesis_instrumentation__.Notify(60719)
	return &Config{
		S2Geography: &S2GeographyConfig{S2Config: DefaultS2Config()},
	}
}

type geogCovererWithBBoxFallback struct {
	rc *s2.RegionCoverer
	g  geo.Geography
}

var _ covererInterface = geogCovererWithBBoxFallback{}

func toDeg(radians float64) float64 {
	__antithesis_instrumentation__.Notify(60720)
	return s1.Angle(radians).Degrees()
}

func (rc geogCovererWithBBoxFallback) covering(regions []s2.Region) s2.CellUnion {
	__antithesis_instrumentation__.Notify(60721)
	cu := simpleCovererImpl{rc: rc.rc}.covering(regions)
	if isBadGeogCovering(cu) {
		__antithesis_instrumentation__.Notify(60723)
		bbox := rc.g.SpatialObject().BoundingBox
		if bbox == nil {
			__antithesis_instrumentation__.Notify(60726)
			return cu
		} else {
			__antithesis_instrumentation__.Notify(60727)
		}
		__antithesis_instrumentation__.Notify(60724)
		flatCoords := []float64{
			toDeg(bbox.LoX), toDeg(bbox.LoY), toDeg(bbox.HiX), toDeg(bbox.LoY),
			toDeg(bbox.HiX), toDeg(bbox.HiY), toDeg(bbox.LoX), toDeg(bbox.HiY),
			toDeg(bbox.LoX), toDeg(bbox.LoY)}
		bboxT := geom.NewPolygonFlat(geom.XY, flatCoords, []int{len(flatCoords)})
		bboxRegions, err := geo.S2RegionsFromGeomT(bboxT, geo.EmptyBehaviorOmit)
		if err != nil {
			__antithesis_instrumentation__.Notify(60728)
			return cu
		} else {
			__antithesis_instrumentation__.Notify(60729)
		}
		__antithesis_instrumentation__.Notify(60725)
		bboxCU := simpleCovererImpl{rc: rc.rc}.covering(bboxRegions)
		if !isBadGeogCovering(bboxCU) {
			__antithesis_instrumentation__.Notify(60730)
			cu = bboxCU
		} else {
			__antithesis_instrumentation__.Notify(60731)
		}
	} else {
		__antithesis_instrumentation__.Notify(60732)
	}
	__antithesis_instrumentation__.Notify(60722)
	return cu
}

func isBadGeogCovering(cu s2.CellUnion) bool {
	__antithesis_instrumentation__.Notify(60733)
	const numFaces = 6
	if len(cu) != numFaces {
		__antithesis_instrumentation__.Notify(60736)
		return false
	} else {
		__antithesis_instrumentation__.Notify(60737)
	}
	__antithesis_instrumentation__.Notify(60734)
	numFaceCells := 0
	for _, c := range cu {
		__antithesis_instrumentation__.Notify(60738)
		if c.Level() == 0 {
			__antithesis_instrumentation__.Notify(60739)
			numFaceCells++
		} else {
			__antithesis_instrumentation__.Notify(60740)
		}
	}
	__antithesis_instrumentation__.Notify(60735)
	return numFaces == numFaceCells
}

func (i *s2GeographyIndex) InvertedIndexKeys(
	c context.Context, g geo.Geography,
) ([]Key, geopb.BoundingBox, error) {
	__antithesis_instrumentation__.Notify(60741)
	r, err := g.AsS2(geo.EmptyBehaviorOmit)
	if err != nil {
		__antithesis_instrumentation__.Notify(60743)
		return nil, geopb.BoundingBox{}, err
	} else {
		__antithesis_instrumentation__.Notify(60744)
	}
	__antithesis_instrumentation__.Notify(60742)
	rect := g.BoundingRect()
	bbox := geopb.BoundingBox{
		LoX: rect.Lng.Lo,
		HiX: rect.Lng.Hi,
		LoY: rect.Lat.Lo,
		HiY: rect.Lat.Hi,
	}
	return invertedIndexKeys(c, geogCovererWithBBoxFallback{rc: i.rc, g: g}, r), bbox, nil
}

func (i *s2GeographyIndex) Covers(c context.Context, g geo.Geography) (UnionKeySpans, error) {
	__antithesis_instrumentation__.Notify(60745)
	r, err := g.AsS2(geo.EmptyBehaviorOmit)
	if err != nil {
		__antithesis_instrumentation__.Notify(60747)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(60748)
	}
	__antithesis_instrumentation__.Notify(60746)
	return covers(c, geogCovererWithBBoxFallback{rc: i.rc, g: g}, r), nil
}

func (i *s2GeographyIndex) CoveredBy(c context.Context, g geo.Geography) (RPKeyExpr, error) {
	__antithesis_instrumentation__.Notify(60749)
	r, err := g.AsS2(geo.EmptyBehaviorOmit)
	if err != nil {
		__antithesis_instrumentation__.Notify(60751)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(60752)
	}
	__antithesis_instrumentation__.Notify(60750)
	return coveredBy(c, i.rc, r), nil
}

func (i *s2GeographyIndex) Intersects(c context.Context, g geo.Geography) (UnionKeySpans, error) {
	__antithesis_instrumentation__.Notify(60753)
	r, err := g.AsS2(geo.EmptyBehaviorOmit)
	if err != nil {
		__antithesis_instrumentation__.Notify(60755)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(60756)
	}
	__antithesis_instrumentation__.Notify(60754)
	return intersects(c, geogCovererWithBBoxFallback{rc: i.rc, g: g}, r), nil
}

func (i *s2GeographyIndex) DWithin(
	_ context.Context,
	g geo.Geography,
	distanceMeters float64,
	useSphereOrSpheroid geogfn.UseSphereOrSpheroid,
) (UnionKeySpans, error) {
	__antithesis_instrumentation__.Notify(60757)
	projInfo, err := geoprojbase.Projection(g.SRID())
	if err != nil {
		__antithesis_instrumentation__.Notify(60763)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(60764)
	}
	__antithesis_instrumentation__.Notify(60758)
	if projInfo.Spheroid == nil {
		__antithesis_instrumentation__.Notify(60765)
		return nil, errors.Errorf("projection %d does not have spheroid", g.SRID())
	} else {
		__antithesis_instrumentation__.Notify(60766)
	}
	__antithesis_instrumentation__.Notify(60759)
	r, err := g.AsS2(geo.EmptyBehaviorOmit)
	if err != nil {
		__antithesis_instrumentation__.Notify(60767)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(60768)
	}
	__antithesis_instrumentation__.Notify(60760)

	gCovering := geogCovererWithBBoxFallback{rc: i.rc, g: g}.covering(r)

	multiplier := 1.0
	if useSphereOrSpheroid == geogfn.UseSpheroid {
		__antithesis_instrumentation__.Notify(60769)

		multiplier += geogfn.SpheroidErrorFraction
	} else {
		__antithesis_instrumentation__.Notify(60770)
	}
	__antithesis_instrumentation__.Notify(60761)
	angle := s1.Angle(multiplier * distanceMeters / projInfo.Spheroid.SphereRadius)

	const maxLevelDiff = 2
	gCovering.ExpandByRadius(angle, maxLevelDiff)

	var covering s2.CellUnion
	for _, c := range gCovering {
		__antithesis_instrumentation__.Notify(60771)
		if c.Level() > i.rc.MaxLevel {
			__antithesis_instrumentation__.Notify(60773)
			c = c.Parent(i.rc.MaxLevel)
		} else {
			__antithesis_instrumentation__.Notify(60774)
		}
		__antithesis_instrumentation__.Notify(60772)
		covering = append(covering, c)
	}
	__antithesis_instrumentation__.Notify(60762)
	covering.Normalize()
	return intersectsUsingCovering(covering), nil
}

func (i *s2GeographyIndex) TestingInnerCovering(g geo.Geography) s2.CellUnion {
	__antithesis_instrumentation__.Notify(60775)
	r, _ := g.AsS2(geo.EmptyBehaviorOmit)
	if r == nil {
		__antithesis_instrumentation__.Notify(60777)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(60778)
	}
	__antithesis_instrumentation__.Notify(60776)
	return innerCovering(i.rc, r)
}

func (i *s2GeographyIndex) CoveringGeography(
	c context.Context, g geo.Geography,
) (geo.Geography, error) {
	__antithesis_instrumentation__.Notify(60779)
	keys, _, err := i.InvertedIndexKeys(c, g)
	if err != nil {
		__antithesis_instrumentation__.Notify(60783)
		return geo.Geography{}, err
	} else {
		__antithesis_instrumentation__.Notify(60784)
	}
	__antithesis_instrumentation__.Notify(60780)
	t, err := makeGeomTFromKeys(keys, g.SRID(), func(p s2.Point) (float64, float64) {
		__antithesis_instrumentation__.Notify(60785)
		latlng := s2.LatLngFromPoint(p)
		return latlng.Lng.Degrees(), latlng.Lat.Degrees()
	})
	__antithesis_instrumentation__.Notify(60781)
	if err != nil {
		__antithesis_instrumentation__.Notify(60786)
		return geo.Geography{}, err
	} else {
		__antithesis_instrumentation__.Notify(60787)
	}
	__antithesis_instrumentation__.Notify(60782)
	return geo.MakeGeographyFromGeomT(t)
}
