package geomfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/twpayne/go-geom"
)

func MakePolygon(outer geo.Geometry, interior ...geo.Geometry) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(62559)
	layout := geom.XY
	outerGeomT, err := outer.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(62565)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(62566)
	}
	__antithesis_instrumentation__.Notify(62560)
	outerRing, ok := outerGeomT.(*geom.LineString)
	if !ok {
		__antithesis_instrumentation__.Notify(62567)
		return geo.Geometry{}, pgerror.Newf(pgcode.InvalidParameterValue, "argument must be LINESTRING geometries")
	} else {
		__antithesis_instrumentation__.Notify(62568)
	}
	__antithesis_instrumentation__.Notify(62561)
	if outerRing.Empty() {
		__antithesis_instrumentation__.Notify(62569)
		return geo.Geometry{}, pgerror.Newf(pgcode.InvalidParameterValue, "polygon shell must not be empty")
	} else {
		__antithesis_instrumentation__.Notify(62570)
	}
	__antithesis_instrumentation__.Notify(62562)
	srid := outerRing.SRID()
	coords := make([][]geom.Coord, len(interior)+1)
	coords[0] = outerRing.Coords()
	for i, g := range interior {
		__antithesis_instrumentation__.Notify(62571)
		interiorRingGeomT, err := g.AsGeomT()
		if err != nil {
			__antithesis_instrumentation__.Notify(62576)
			return geo.Geometry{}, err
		} else {
			__antithesis_instrumentation__.Notify(62577)
		}
		__antithesis_instrumentation__.Notify(62572)
		interiorRing, ok := interiorRingGeomT.(*geom.LineString)
		if !ok {
			__antithesis_instrumentation__.Notify(62578)
			return geo.Geometry{}, pgerror.Newf(pgcode.InvalidParameterValue, "argument must be LINESTRING geometries")
		} else {
			__antithesis_instrumentation__.Notify(62579)
		}
		__antithesis_instrumentation__.Notify(62573)
		if interiorRing.SRID() != srid {
			__antithesis_instrumentation__.Notify(62580)
			return geo.Geometry{}, pgerror.Newf(pgcode.InvalidParameterValue, "mixed SRIDs are not allowed")
		} else {
			__antithesis_instrumentation__.Notify(62581)
		}
		__antithesis_instrumentation__.Notify(62574)
		if outerRing.Layout() != interiorRing.Layout() {
			__antithesis_instrumentation__.Notify(62582)
			return geo.Geometry{}, pgerror.Newf(pgcode.InvalidParameterValue, "mixed dimension rings")
		} else {
			__antithesis_instrumentation__.Notify(62583)
		}
		__antithesis_instrumentation__.Notify(62575)
		coords[i+1] = interiorRing.Coords()
	}
	__antithesis_instrumentation__.Notify(62563)

	polygon, err := geom.NewPolygon(layout).SetSRID(srid).SetCoords(coords)
	if err != nil {
		__antithesis_instrumentation__.Notify(62584)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(62585)
	}
	__antithesis_instrumentation__.Notify(62564)
	return geo.MakeGeometryFromGeomT(polygon)
}

func MakePolygonWithSRID(g geo.Geometry, srid int) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(62586)
	polygon, err := MakePolygon(g)
	if err != nil {
		__antithesis_instrumentation__.Notify(62589)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(62590)
	}
	__antithesis_instrumentation__.Notify(62587)
	t, err := polygon.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(62591)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(62592)
	}
	__antithesis_instrumentation__.Notify(62588)
	geo.AdjustGeomTSRID(t, geopb.SRID(srid))
	return geo.MakeGeometryFromGeomT(t)
}
