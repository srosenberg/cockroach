package geoindex

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geogfn"
	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/golang/geo/s2"
	"github.com/twpayne/go-geom"
)

var RelationshipMap = map[string]RelationshipType{
	"st_covers":                Covers,
	"st_coveredby":             CoveredBy,
	"st_contains":              Covers,
	"st_containsproperly":      Covers,
	"st_crosses":               Intersects,
	"st_dwithin":               DWithin,
	"st_dfullywithin":          DFullyWithin,
	"st_equals":                Intersects,
	"st_intersects":            Intersects,
	"st_overlaps":              Intersects,
	"st_touches":               Intersects,
	"st_within":                CoveredBy,
	"st_dwithinexclusive":      DWithin,
	"st_dfullywithinexclusive": DFullyWithin,
}

var RelationshipReverseMap = map[RelationshipType]string{
	Covers:       "st_covers",
	CoveredBy:    "st_coveredby",
	DWithin:      "st_dwithin",
	DFullyWithin: "st_dfullywithin",
	Intersects:   "st_intersects",
}

var CommuteRelationshipMap = map[RelationshipType]RelationshipType{
	Covers:       CoveredBy,
	CoveredBy:    Covers,
	DWithin:      DWithin,
	DFullyWithin: DFullyWithin,
	Intersects:   Intersects,
}

type GeographyIndex interface {
	InvertedIndexKeys(c context.Context, g geo.Geography) ([]Key, geopb.BoundingBox, error)

	Covers(c context.Context, g geo.Geography) (UnionKeySpans, error)

	CoveredBy(c context.Context, g geo.Geography) (RPKeyExpr, error)

	Intersects(c context.Context, g geo.Geography) (UnionKeySpans, error)

	DWithin(
		c context.Context, g geo.Geography, distanceMeters float64,
		useSphereOrSpheroid geogfn.UseSphereOrSpheroid,
	) (UnionKeySpans, error)

	TestingInnerCovering(g geo.Geography) s2.CellUnion

	CoveringGeography(c context.Context, g geo.Geography) (geo.Geography, error)
}

type GeometryIndex interface {
	InvertedIndexKeys(c context.Context, g geo.Geometry) ([]Key, geopb.BoundingBox, error)

	Covers(c context.Context, g geo.Geometry) (UnionKeySpans, error)

	CoveredBy(c context.Context, g geo.Geometry) (RPKeyExpr, error)

	Intersects(c context.Context, g geo.Geometry) (UnionKeySpans, error)

	DWithin(c context.Context, g geo.Geometry, distance float64) (UnionKeySpans, error)

	DFullyWithin(c context.Context, g geo.Geometry, distance float64) (UnionKeySpans, error)

	TestingInnerCovering(g geo.Geometry) s2.CellUnion

	CoveringGeometry(c context.Context, g geo.Geometry) (geo.Geometry, error)
}

type RelationshipType uint8

const (
	Covers RelationshipType = (1 << iota)

	CoveredBy

	Intersects

	DWithin

	DFullyWithin
)

var geoRelationshipTypeStr = map[RelationshipType]string{
	Covers:       "covers",
	CoveredBy:    "covered by",
	Intersects:   "intersects",
	DWithin:      "dwithin",
	DFullyWithin: "dfullywithin",
}

func (gr RelationshipType) String() string {
	__antithesis_instrumentation__.Notify(60592)
	return geoRelationshipTypeStr[gr]
}

func IsEmptyConfig(cfg *Config) bool {
	__antithesis_instrumentation__.Notify(60593)
	if cfg == nil {
		__antithesis_instrumentation__.Notify(60595)
		return true
	} else {
		__antithesis_instrumentation__.Notify(60596)
	}
	__antithesis_instrumentation__.Notify(60594)
	return cfg.S2Geography == nil && func() bool {
		__antithesis_instrumentation__.Notify(60597)
		return cfg.S2Geometry == nil == true
	}() == true
}

func IsGeographyConfig(cfg *Config) bool {
	__antithesis_instrumentation__.Notify(60598)
	if cfg == nil {
		__antithesis_instrumentation__.Notify(60600)
		return false
	} else {
		__antithesis_instrumentation__.Notify(60601)
	}
	__antithesis_instrumentation__.Notify(60599)
	return cfg.S2Geography != nil
}

func IsGeometryConfig(cfg *Config) bool {
	__antithesis_instrumentation__.Notify(60602)
	if cfg == nil {
		__antithesis_instrumentation__.Notify(60604)
		return false
	} else {
		__antithesis_instrumentation__.Notify(60605)
	}
	__antithesis_instrumentation__.Notify(60603)
	return cfg.S2Geometry != nil
}

type Key uint64

func (k Key) rpExprElement() { __antithesis_instrumentation__.Notify(60606) }

func (k Key) String() string {
	__antithesis_instrumentation__.Notify(60607)
	c := s2.CellID(k)
	if !c.IsValid() {
		__antithesis_instrumentation__.Notify(60610)
		return "spilled"
	} else {
		__antithesis_instrumentation__.Notify(60611)
	}
	__antithesis_instrumentation__.Notify(60608)
	var b strings.Builder
	b.WriteByte('F')
	b.WriteByte("012345"[c.Face()])
	fmt.Fprintf(&b, "/L%d/", c.Level())
	for level := 1; level <= c.Level(); level++ {
		__antithesis_instrumentation__.Notify(60612)
		b.WriteByte("0123"[c.ChildPosition(level)])
	}
	__antithesis_instrumentation__.Notify(60609)
	return b.String()
}

func (k Key) S2CellID() s2.CellID {
	__antithesis_instrumentation__.Notify(60613)
	return s2.CellID(k)
}

type KeySpan struct {
	Start, End Key
}

type UnionKeySpans []KeySpan

func (s UnionKeySpans) String() string {
	__antithesis_instrumentation__.Notify(60614)
	return s.toString(math.MaxInt32)
}

func (s UnionKeySpans) toString(wrap int) string {
	__antithesis_instrumentation__.Notify(60615)
	b := newStringBuilderWithWrap(&strings.Builder{}, wrap)
	for i, span := range s {
		__antithesis_instrumentation__.Notify(60617)
		if span.Start == span.End {
			__antithesis_instrumentation__.Notify(60620)
			fmt.Fprintf(b, "%s", span.Start)
		} else {
			__antithesis_instrumentation__.Notify(60621)
			fmt.Fprintf(b, "[%s, %s]", span.Start, span.End)
		}
		__antithesis_instrumentation__.Notify(60618)
		if i != len(s)-1 {
			__antithesis_instrumentation__.Notify(60622)
			b.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(60623)
		}
		__antithesis_instrumentation__.Notify(60619)
		b.tryWrap()
	}
	__antithesis_instrumentation__.Notify(60616)
	return b.String()
}

type RPExprElement interface {
	rpExprElement()
}

type RPSetOperator int

const (
	RPSetUnion RPSetOperator = iota + 1

	RPSetIntersection
)

func (o RPSetOperator) rpExprElement() { __antithesis_instrumentation__.Notify(60624) }

type RPKeyExpr []RPExprElement

func (x RPKeyExpr) String() string {
	__antithesis_instrumentation__.Notify(60625)
	var elements []string
	for _, e := range x {
		__antithesis_instrumentation__.Notify(60627)
		switch elem := e.(type) {
		case Key:
			__antithesis_instrumentation__.Notify(60628)
			elements = append(elements, elem.String())
		case RPSetOperator:
			__antithesis_instrumentation__.Notify(60629)
			switch elem {
			case RPSetUnion:
				__antithesis_instrumentation__.Notify(60630)
				elements = append(elements, `\U`)
			case RPSetIntersection:
				__antithesis_instrumentation__.Notify(60631)
				elements = append(elements, `\I`)
			default:
				__antithesis_instrumentation__.Notify(60632)
			}
		}
	}
	__antithesis_instrumentation__.Notify(60626)
	return strings.Join(elements, " ")
}

type covererInterface interface {
	covering(regions []s2.Region) s2.CellUnion
}

type simpleCovererImpl struct {
	rc *s2.RegionCoverer
}

var _ covererInterface = simpleCovererImpl{}

func (rc simpleCovererImpl) covering(regions []s2.Region) s2.CellUnion {
	__antithesis_instrumentation__.Notify(60633)

	var u s2.CellUnion
	for _, r := range regions {
		__antithesis_instrumentation__.Notify(60635)
		u = append(u, rc.rc.Covering(r)...)
	}
	__antithesis_instrumentation__.Notify(60634)

	u.Normalize()
	return u
}

func innerCovering(rc *s2.RegionCoverer, regions []s2.Region) s2.CellUnion {
	__antithesis_instrumentation__.Notify(60636)
	var u s2.CellUnion
	for _, r := range regions {
		__antithesis_instrumentation__.Notify(60638)
		switch region := r.(type) {
		case s2.Point:
			__antithesis_instrumentation__.Notify(60639)
			cellID := cellIDCoveringPoint(region, rc.MaxLevel)
			u = append(u, cellID)
		case *s2.Polyline:
			__antithesis_instrumentation__.Notify(60640)

			for _, p := range *region {
				__antithesis_instrumentation__.Notify(60643)
				cellID := cellIDCoveringPoint(p, rc.MaxLevel)
				u = append(u, cellID)
			}
		case *s2.Polygon:
			__antithesis_instrumentation__.Notify(60641)

			if region.NumLoops() > 0 {
				__antithesis_instrumentation__.Notify(60644)
				loop := region.Loop(0)
				for _, p := range loop.Vertices() {
					__antithesis_instrumentation__.Notify(60646)
					cellID := cellIDCoveringPoint(p, rc.MaxLevel)
					u = append(u, cellID)
				}
				__antithesis_instrumentation__.Notify(60645)

				const smallPolygonLevelThreshold = 25
				if region.Area() > s2.AvgAreaMetric.Value(smallPolygonLevelThreshold) {
					__antithesis_instrumentation__.Notify(60647)
					u = append(u, rc.InteriorCovering(region)...)
				} else {
					__antithesis_instrumentation__.Notify(60648)
				}
			} else {
				__antithesis_instrumentation__.Notify(60649)
			}
		default:
			__antithesis_instrumentation__.Notify(60642)
			panic("bug: code should not be producing unhandled Region type")
		}
	}
	__antithesis_instrumentation__.Notify(60637)

	u.Normalize()

	return u
}

func cellIDCoveringPoint(point s2.Point, level int) s2.CellID {
	__antithesis_instrumentation__.Notify(60650)
	cellID := s2.CellFromPoint(point).ID()
	if !cellID.IsLeaf() {
		__antithesis_instrumentation__.Notify(60652)
		panic("bug in S2")
	} else {
		__antithesis_instrumentation__.Notify(60653)
	}
	__antithesis_instrumentation__.Notify(60651)
	return cellID.Parent(level)
}

func ancestorCells(cells []s2.CellID) []s2.CellID {
	__antithesis_instrumentation__.Notify(60654)
	var ancestors []s2.CellID
	var seen map[s2.CellID]struct{}
	if len(cells) > 1 {
		__antithesis_instrumentation__.Notify(60657)
		seen = make(map[s2.CellID]struct{})
	} else {
		__antithesis_instrumentation__.Notify(60658)
	}
	__antithesis_instrumentation__.Notify(60655)
	for _, c := range cells {
		__antithesis_instrumentation__.Notify(60659)
		for l := c.Level() - 1; l >= 0; l-- {
			__antithesis_instrumentation__.Notify(60660)
			p := c.Parent(l)
			if seen != nil {
				__antithesis_instrumentation__.Notify(60662)
				if _, ok := seen[p]; ok {
					__antithesis_instrumentation__.Notify(60664)
					break
				} else {
					__antithesis_instrumentation__.Notify(60665)
				}
				__antithesis_instrumentation__.Notify(60663)
				seen[p] = struct{}{}
			} else {
				__antithesis_instrumentation__.Notify(60666)
			}
			__antithesis_instrumentation__.Notify(60661)
			ancestors = append(ancestors, p)
		}
	}
	__antithesis_instrumentation__.Notify(60656)
	return ancestors
}

func invertedIndexKeys(_ context.Context, rc covererInterface, r []s2.Region) []Key {
	__antithesis_instrumentation__.Notify(60667)
	covering := rc.covering(r)
	keys := make([]Key, len(covering))
	for i, cid := range covering {
		__antithesis_instrumentation__.Notify(60669)
		keys[i] = Key(cid)
	}
	__antithesis_instrumentation__.Notify(60668)
	return keys
}

func covers(c context.Context, rc covererInterface, r []s2.Region) UnionKeySpans {
	__antithesis_instrumentation__.Notify(60670)

	return intersects(c, rc, r)
}

func intersects(_ context.Context, rc covererInterface, r []s2.Region) UnionKeySpans {
	__antithesis_instrumentation__.Notify(60671)
	covering := rc.covering(r)
	return intersectsUsingCovering(covering)
}

func intersectsUsingCovering(covering s2.CellUnion) UnionKeySpans {
	__antithesis_instrumentation__.Notify(60672)
	querySpans := make([]KeySpan, len(covering))
	for i, cid := range covering {
		__antithesis_instrumentation__.Notify(60676)
		querySpans[i] = KeySpan{Start: Key(cid.RangeMin()), End: Key(cid.RangeMax())}
	}
	__antithesis_instrumentation__.Notify(60673)
	for _, cid := range ancestorCells(covering) {
		__antithesis_instrumentation__.Notify(60677)
		querySpans = append(querySpans, KeySpan{Start: Key(cid), End: Key(cid)})
	}
	__antithesis_instrumentation__.Notify(60674)
	sort.Slice(querySpans, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(60678)
		return querySpans[i].Start < querySpans[j].Start
	})
	__antithesis_instrumentation__.Notify(60675)
	return querySpans
}

func coveredBy(_ context.Context, rc *s2.RegionCoverer, r []s2.Region) RPKeyExpr {
	__antithesis_instrumentation__.Notify(60679)
	covering := innerCovering(rc, r)
	ancestors := ancestorCells(covering)

	presentCells := make(map[s2.CellID]struct{}, len(covering)+len(ancestors))
	for _, c := range covering {
		__antithesis_instrumentation__.Notify(60683)
		presentCells[c] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(60680)
	for _, c := range ancestors {
		__antithesis_instrumentation__.Notify(60684)
		presentCells[c] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(60681)

	expr := make([]RPExprElement, 0, len(presentCells)*2)
	numFaces := 0
	for face := 0; face < 6; face++ {
		__antithesis_instrumentation__.Notify(60685)
		rootID := s2.CellIDFromFace(face)
		if _, ok := presentCells[rootID]; !ok {
			__antithesis_instrumentation__.Notify(60687)
			continue
		} else {
			__antithesis_instrumentation__.Notify(60688)
		}
		__antithesis_instrumentation__.Notify(60686)
		expr = generateRPExprForTree(rootID, presentCells, expr)
		numFaces++
		if numFaces > 1 {
			__antithesis_instrumentation__.Notify(60689)
			expr = append(expr, RPSetIntersection)
		} else {
			__antithesis_instrumentation__.Notify(60690)
		}
	}
	__antithesis_instrumentation__.Notify(60682)
	return expr
}

func generateRPExprForTree(
	rootID s2.CellID, presentCells map[s2.CellID]struct{}, expr []RPExprElement,
) []RPExprElement {
	__antithesis_instrumentation__.Notify(60691)
	expr = append(expr, Key(rootID))
	if rootID.IsLeaf() {
		__antithesis_instrumentation__.Notify(60695)
		return expr
	} else {
		__antithesis_instrumentation__.Notify(60696)
	}
	__antithesis_instrumentation__.Notify(60692)
	numChildren := 0
	for _, childCellID := range rootID.Children() {
		__antithesis_instrumentation__.Notify(60697)
		if _, ok := presentCells[childCellID]; !ok {
			__antithesis_instrumentation__.Notify(60699)
			continue
		} else {
			__antithesis_instrumentation__.Notify(60700)
		}
		__antithesis_instrumentation__.Notify(60698)
		expr = generateRPExprForTree(childCellID, presentCells, expr)
		numChildren++
		if numChildren > 1 {
			__antithesis_instrumentation__.Notify(60701)
			expr = append(expr, RPSetIntersection)
		} else {
			__antithesis_instrumentation__.Notify(60702)
		}
	}
	__antithesis_instrumentation__.Notify(60693)
	if numChildren > 0 {
		__antithesis_instrumentation__.Notify(60703)
		expr = append(expr, RPSetUnion)
	} else {
		__antithesis_instrumentation__.Notify(60704)
	}
	__antithesis_instrumentation__.Notify(60694)
	return expr
}

type stringBuilderWithWrap struct {
	*strings.Builder
	wrap     int
	lastWrap int
}

func newStringBuilderWithWrap(b *strings.Builder, wrap int) *stringBuilderWithWrap {
	__antithesis_instrumentation__.Notify(60705)
	return &stringBuilderWithWrap{
		Builder:  b,
		wrap:     wrap,
		lastWrap: b.Len(),
	}
}

func (b *stringBuilderWithWrap) tryWrap() {
	__antithesis_instrumentation__.Notify(60706)
	if b.Len()-b.lastWrap > b.wrap {
		__antithesis_instrumentation__.Notify(60707)
		b.doWrap()
	} else {
		__antithesis_instrumentation__.Notify(60708)
	}
}

func (b *stringBuilderWithWrap) doWrap() {
	__antithesis_instrumentation__.Notify(60709)
	fmt.Fprintln(b)
	b.lastWrap = b.Len()
}

func DefaultS2Config() *S2Config {
	__antithesis_instrumentation__.Notify(60710)
	return &S2Config{
		MinLevel: 0,
		MaxLevel: 30,
		LevelMod: 1,
		MaxCells: 4,
	}
}

func makeGeomTFromKeys(
	keys []Key, srid geopb.SRID, xyFromS2Point func(s2.Point) (float64, float64),
) (geom.T, error) {
	__antithesis_instrumentation__.Notify(60711)
	t := geom.NewMultiPolygon(geom.XY).SetSRID(int(srid))
	for _, key := range keys {
		__antithesis_instrumentation__.Notify(60713)
		cell := s2.CellFromCellID(key.S2CellID())
		flatCoords := make([]float64, 0, 10)
		for i := 0; i < 4; i++ {
			__antithesis_instrumentation__.Notify(60715)
			x, y := xyFromS2Point(cell.Vertex(i))
			flatCoords = append(flatCoords, x, y)
		}
		__antithesis_instrumentation__.Notify(60714)

		flatCoords = append(flatCoords, flatCoords[0:2]...)
		err := t.Push(geom.NewPolygonFlat(geom.XY, flatCoords, []int{10}))
		if err != nil {
			__antithesis_instrumentation__.Notify(60716)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(60717)
		}
	}
	__antithesis_instrumentation__.Notify(60712)
	return t, nil
}
