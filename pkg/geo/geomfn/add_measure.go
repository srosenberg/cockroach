package geomfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/twpayne/go-geom"
)

func AddMeasure(geometry geo.Geometry, start float64, end float64) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(60976)
	t, err := geometry.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(60978)
		return geometry, err
	} else {
		__antithesis_instrumentation__.Notify(60979)
	}
	__antithesis_instrumentation__.Notify(60977)

	switch t := t.(type) {
	case *geom.LineString:
		__antithesis_instrumentation__.Notify(60980)
		newLineString, err := addMeasureToLineString(t, start, end)
		if err != nil {
			__antithesis_instrumentation__.Notify(60985)
			return geometry, err
		} else {
			__antithesis_instrumentation__.Notify(60986)
		}
		__antithesis_instrumentation__.Notify(60981)
		return geo.MakeGeometryFromGeomT(newLineString)
	case *geom.MultiLineString:
		__antithesis_instrumentation__.Notify(60982)
		newMultiLineString, err := addMeasureToMultiLineString(t, start, end)
		if err != nil {
			__antithesis_instrumentation__.Notify(60987)
			return geometry, err
		} else {
			__antithesis_instrumentation__.Notify(60988)
		}
		__antithesis_instrumentation__.Notify(60983)
		return geo.MakeGeometryFromGeomT(newMultiLineString)
	default:
		__antithesis_instrumentation__.Notify(60984)

		return geometry, pgerror.Newf(pgcode.InvalidParameterValue, "input geometry must be LINESTRING or MULTILINESTRING")
	}
}

func addMeasureToMultiLineString(
	multiLineString *geom.MultiLineString, start float64, end float64,
) (*geom.MultiLineString, error) {
	__antithesis_instrumentation__.Notify(60989)
	newMultiLineString :=
		geom.NewMultiLineString(augmentLayoutWithM(multiLineString.Layout())).SetSRID(multiLineString.SRID())

	for i := 0; i < multiLineString.NumLineStrings(); i++ {
		__antithesis_instrumentation__.Notify(60991)
		newLineString, err := addMeasureToLineString(multiLineString.LineString(i), start, end)
		if err != nil {
			__antithesis_instrumentation__.Notify(60993)
			return multiLineString, err
		} else {
			__antithesis_instrumentation__.Notify(60994)
		}
		__antithesis_instrumentation__.Notify(60992)
		err = newMultiLineString.Push(newLineString)
		if err != nil {
			__antithesis_instrumentation__.Notify(60995)
			return multiLineString, err
		} else {
			__antithesis_instrumentation__.Notify(60996)
		}
	}
	__antithesis_instrumentation__.Notify(60990)

	return newMultiLineString, nil
}

func addMeasureToLineString(
	lineString *geom.LineString, start float64, end float64,
) (*geom.LineString, error) {
	__antithesis_instrumentation__.Notify(60997)
	newLineString := geom.NewLineString(augmentLayoutWithM(lineString.Layout())).SetSRID(lineString.SRID())

	if lineString.Empty() {
		__antithesis_instrumentation__.Notify(61003)
		return newLineString, nil
	} else {
		__antithesis_instrumentation__.Notify(61004)
	}
	__antithesis_instrumentation__.Notify(60998)

	lineCoords := lineString.Coords()

	prevPoint := lineCoords[0]
	lineLength := float64(0)
	pointMeasures := make([]float64, lineString.NumCoords())
	for i := 0; i < lineString.NumCoords(); i++ {
		__antithesis_instrumentation__.Notify(61005)
		curPoint := lineCoords[i]
		distBetweenPoints := coordNorm(coordSub(prevPoint, curPoint))
		lineLength += distBetweenPoints
		pointMeasures[i] = lineLength
		prevPoint = curPoint
	}
	__antithesis_instrumentation__.Notify(60999)

	for i := 0; i < lineString.NumCoords(); i++ {
		__antithesis_instrumentation__.Notify(61006)

		if lineLength == 0 {
			__antithesis_instrumentation__.Notify(61007)
			pointMeasures[i] = start + (end-start)*(float64(i)/float64(lineString.NumCoords()-1))
		} else {
			__antithesis_instrumentation__.Notify(61008)
			pointMeasures[i] = start + (end-start)*(pointMeasures[i]/lineLength)
		}
	}
	__antithesis_instrumentation__.Notify(61000)

	for i := 0; i < lineString.NumCoords(); i++ {
		__antithesis_instrumentation__.Notify(61009)
		if lineString.Layout().MIndex() == -1 {
			__antithesis_instrumentation__.Notify(61010)
			lineCoords[i] = append(lineCoords[i], pointMeasures[i])
		} else {
			__antithesis_instrumentation__.Notify(61011)
			lineCoords[i][lineString.Layout().MIndex()] = pointMeasures[i]
		}
	}
	__antithesis_instrumentation__.Notify(61001)

	_, err := newLineString.SetCoords(lineCoords)
	if err != nil {
		__antithesis_instrumentation__.Notify(61012)
		return lineString, err
	} else {
		__antithesis_instrumentation__.Notify(61013)
	}
	__antithesis_instrumentation__.Notify(61002)

	return newLineString, nil
}

func augmentLayoutWithM(layout geom.Layout) geom.Layout {
	__antithesis_instrumentation__.Notify(61014)
	switch layout {
	case geom.XY, geom.XYM:
		__antithesis_instrumentation__.Notify(61015)
		return geom.XYM
	case geom.XYZ, geom.XYZM:
		__antithesis_instrumentation__.Notify(61016)
		return geom.XYZM
	default:
		__antithesis_instrumentation__.Notify(61017)
		return layout
	}
}
