package geosegmentize

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"
	"strconv"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/twpayne/go-geom"
)

func Segmentize(
	geometry geom.T,
	segmentMaxAngleOrLength float64,
	segmentizeCoords func(geom.Coord, geom.Coord, float64) ([]float64, error),
) (geom.T, error) {
	__antithesis_instrumentation__.Notify(64548)
	if geometry.Empty() {
		__antithesis_instrumentation__.Notify(64551)
		return geometry, nil
	} else {
		__antithesis_instrumentation__.Notify(64552)
	}
	__antithesis_instrumentation__.Notify(64549)
	layout := geometry.Layout()
	switch geometry := geometry.(type) {
	case *geom.Point, *geom.MultiPoint:
		__antithesis_instrumentation__.Notify(64553)
		return geometry, nil
	case *geom.LineString:
		__antithesis_instrumentation__.Notify(64554)
		var allFlatCoordinates []float64
		for pointIdx := 1; pointIdx < geometry.NumCoords(); pointIdx++ {
			__antithesis_instrumentation__.Notify(64567)
			coords, err := segmentizeCoords(geometry.Coord(pointIdx-1), geometry.Coord(pointIdx), segmentMaxAngleOrLength)
			if err != nil {
				__antithesis_instrumentation__.Notify(64569)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(64570)
			}
			__antithesis_instrumentation__.Notify(64568)
			allFlatCoordinates = append(
				allFlatCoordinates,
				coords...,
			)
		}
		__antithesis_instrumentation__.Notify(64555)

		allFlatCoordinates = append(allFlatCoordinates, geometry.Coord(geometry.NumCoords()-1)...)
		return geom.NewLineStringFlat(layout, allFlatCoordinates).SetSRID(geometry.SRID()), nil
	case *geom.MultiLineString:
		__antithesis_instrumentation__.Notify(64556)
		segMultiLine := geom.NewMultiLineString(layout).SetSRID(geometry.SRID())
		for lineIdx := 0; lineIdx < geometry.NumLineStrings(); lineIdx++ {
			__antithesis_instrumentation__.Notify(64571)
			l, err := Segmentize(geometry.LineString(lineIdx), segmentMaxAngleOrLength, segmentizeCoords)
			if err != nil {
				__antithesis_instrumentation__.Notify(64573)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(64574)
			}
			__antithesis_instrumentation__.Notify(64572)
			err = segMultiLine.Push(l.(*geom.LineString))
			if err != nil {
				__antithesis_instrumentation__.Notify(64575)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(64576)
			}
		}
		__antithesis_instrumentation__.Notify(64557)
		return segMultiLine, nil
	case *geom.LinearRing:
		__antithesis_instrumentation__.Notify(64558)
		var allFlatCoordinates []float64
		for pointIdx := 1; pointIdx < geometry.NumCoords(); pointIdx++ {
			__antithesis_instrumentation__.Notify(64577)
			coords, err := segmentizeCoords(geometry.Coord(pointIdx-1), geometry.Coord(pointIdx), segmentMaxAngleOrLength)
			if err != nil {
				__antithesis_instrumentation__.Notify(64579)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(64580)
			}
			__antithesis_instrumentation__.Notify(64578)
			allFlatCoordinates = append(
				allFlatCoordinates,
				coords...,
			)
		}
		__antithesis_instrumentation__.Notify(64559)

		allFlatCoordinates = append(allFlatCoordinates, geometry.Coord(geometry.NumCoords()-1)...)
		return geom.NewLinearRingFlat(layout, allFlatCoordinates).SetSRID(geometry.SRID()), nil
	case *geom.Polygon:
		__antithesis_instrumentation__.Notify(64560)
		segPolygon := geom.NewPolygon(layout).SetSRID(geometry.SRID())
		for loopIdx := 0; loopIdx < geometry.NumLinearRings(); loopIdx++ {
			__antithesis_instrumentation__.Notify(64581)
			l, err := Segmentize(geometry.LinearRing(loopIdx), segmentMaxAngleOrLength, segmentizeCoords)
			if err != nil {
				__antithesis_instrumentation__.Notify(64583)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(64584)
			}
			__antithesis_instrumentation__.Notify(64582)
			err = segPolygon.Push(l.(*geom.LinearRing))
			if err != nil {
				__antithesis_instrumentation__.Notify(64585)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(64586)
			}
		}
		__antithesis_instrumentation__.Notify(64561)
		return segPolygon, nil
	case *geom.MultiPolygon:
		__antithesis_instrumentation__.Notify(64562)
		segMultiPolygon := geom.NewMultiPolygon(layout).SetSRID(geometry.SRID())
		for polygonIdx := 0; polygonIdx < geometry.NumPolygons(); polygonIdx++ {
			__antithesis_instrumentation__.Notify(64587)
			p, err := Segmentize(geometry.Polygon(polygonIdx), segmentMaxAngleOrLength, segmentizeCoords)
			if err != nil {
				__antithesis_instrumentation__.Notify(64589)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(64590)
			}
			__antithesis_instrumentation__.Notify(64588)
			err = segMultiPolygon.Push(p.(*geom.Polygon))
			if err != nil {
				__antithesis_instrumentation__.Notify(64591)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(64592)
			}
		}
		__antithesis_instrumentation__.Notify(64563)
		return segMultiPolygon, nil
	case *geom.GeometryCollection:
		__antithesis_instrumentation__.Notify(64564)
		segGeomCollection := geom.NewGeometryCollection().SetSRID(geometry.SRID())
		err := segGeomCollection.SetLayout(layout)
		if err != nil {
			__antithesis_instrumentation__.Notify(64593)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(64594)
		}
		__antithesis_instrumentation__.Notify(64565)
		for geoIdx := 0; geoIdx < geometry.NumGeoms(); geoIdx++ {
			__antithesis_instrumentation__.Notify(64595)
			g, err := Segmentize(geometry.Geom(geoIdx), segmentMaxAngleOrLength, segmentizeCoords)
			if err != nil {
				__antithesis_instrumentation__.Notify(64597)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(64598)
			}
			__antithesis_instrumentation__.Notify(64596)
			err = segGeomCollection.Push(g)
			if err != nil {
				__antithesis_instrumentation__.Notify(64599)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(64600)
			}
		}
		__antithesis_instrumentation__.Notify(64566)
		return segGeomCollection, nil
	}
	__antithesis_instrumentation__.Notify(64550)
	return nil, pgerror.Newf(pgcode.InvalidParameterValue, "unknown type: %T", geometry)
}

func CheckSegmentizeValidNumPoints(numPoints float64, a geom.Coord, b geom.Coord) error {
	__antithesis_instrumentation__.Notify(64601)
	if math.IsNaN(numPoints) {
		__antithesis_instrumentation__.Notify(64604)
		return pgerror.Newf(pgcode.InvalidParameterValue, "cannot segmentize into %f points", numPoints)
	} else {
		__antithesis_instrumentation__.Notify(64605)
	}
	__antithesis_instrumentation__.Notify(64602)
	if numPoints > float64(geo.MaxAllowedSplitPoints) {
		__antithesis_instrumentation__.Notify(64606)
		return pgerror.Newf(
			pgcode.InvalidParameterValue,
			"attempting to segmentize into too many coordinates; need %s points between %v and %v, max %d",
			strings.TrimRight(strconv.FormatFloat(numPoints, 'f', -1, 64), "."),
			a,
			b,
			geo.MaxAllowedSplitPoints,
		)
	} else {
		__antithesis_instrumentation__.Notify(64607)
	}
	__antithesis_instrumentation__.Notify(64603)
	return nil
}
