package geo

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/twpayne/go-geom"
)

func Summary(
	t geom.T, hasBoundingBox bool, shape geopb.ShapeType, isGeography bool,
) (string, error) {
	__antithesis_instrumentation__.Notify(64866)
	return summary(t, hasBoundingBox, geopb.SRID(t.SRID()) != geopb.UnknownSRID, shape, isGeography, 0)
}

func summary(
	t geom.T, hasBoundingBox bool, hasSRID bool, shape geopb.ShapeType, isGeography bool, offset int,
) (summaryLine string, err error) {
	__antithesis_instrumentation__.Notify(64867)
	f, err := summaryFlag(t, hasBoundingBox, hasSRID, isGeography)
	if err != nil {
		__antithesis_instrumentation__.Notify(64869)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(64870)
	}
	__antithesis_instrumentation__.Notify(64868)

	summaryLine += strings.Repeat(" ", offset)
	switch t := t.(type) {
	case *geom.Point:
		__antithesis_instrumentation__.Notify(64871)
		return summaryLine + fmt.Sprintf("%s[%s]", shape.String(), f), nil
	case *geom.LineString:
		__antithesis_instrumentation__.Notify(64872)
		return summaryLine + fmt.Sprintf("%s[%s] with %d points", shape.String(), f, t.NumCoords()), nil
	case *geom.Polygon:
		__antithesis_instrumentation__.Notify(64873)
		numLinearRings := t.NumLinearRings()

		summaryLine += fmt.Sprintf("%s[%s] with %d ring", shape.String(), f, t.NumLinearRings())
		if numLinearRings > 1 {
			__antithesis_instrumentation__.Notify(64889)
			summaryLine += "s"
		} else {
			__antithesis_instrumentation__.Notify(64890)
		}
		__antithesis_instrumentation__.Notify(64874)

		for i := 0; i < numLinearRings; i++ {
			__antithesis_instrumentation__.Notify(64891)
			ring := t.LinearRing(i)
			summaryLine += fmt.Sprintf("\n   ring %d has %d points", i, ring.NumCoords())
		}
		__antithesis_instrumentation__.Notify(64875)

		return summaryLine, nil
	case *geom.MultiPoint:
		__antithesis_instrumentation__.Notify(64876)
		numPoints := t.NumPoints()

		summaryLine += fmt.Sprintf("%s[%s] with %d element", shape.String(), f, numPoints)
		if 1 < numPoints {
			__antithesis_instrumentation__.Notify(64892)
			summaryLine += "s"
		} else {
			__antithesis_instrumentation__.Notify(64893)
		}
		__antithesis_instrumentation__.Notify(64877)

		for i := 0; i < numPoints; i++ {
			__antithesis_instrumentation__.Notify(64894)
			point := t.Point(i)
			line, err := summary(point, false, hasSRID, geopb.ShapeType_Point, isGeography, offset+2)
			if err != nil {
				__antithesis_instrumentation__.Notify(64896)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(64897)
			}
			__antithesis_instrumentation__.Notify(64895)

			summaryLine += "\n" + line
		}
		__antithesis_instrumentation__.Notify(64878)

		return summaryLine, nil
	case *geom.MultiLineString:
		__antithesis_instrumentation__.Notify(64879)
		numLineStrings := t.NumLineStrings()

		summaryLine += fmt.Sprintf("%s[%s] with %d element", shape.String(), f, numLineStrings)
		if 1 < numLineStrings {
			__antithesis_instrumentation__.Notify(64898)
			summaryLine += "s"
		} else {
			__antithesis_instrumentation__.Notify(64899)
		}
		__antithesis_instrumentation__.Notify(64880)

		for i := 0; i < numLineStrings; i++ {
			__antithesis_instrumentation__.Notify(64900)
			lineString := t.LineString(i)
			line, err := summary(lineString, false, hasSRID, geopb.ShapeType_LineString, isGeography, offset+2)
			if err != nil {
				__antithesis_instrumentation__.Notify(64902)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(64903)
			}
			__antithesis_instrumentation__.Notify(64901)

			summaryLine += "\n" + line
		}
		__antithesis_instrumentation__.Notify(64881)

		return summaryLine, nil
	case *geom.MultiPolygon:
		__antithesis_instrumentation__.Notify(64882)
		numPolygons := t.NumPolygons()

		summaryLine += fmt.Sprintf("%s[%s] with %d element", shape.String(), f, numPolygons)
		if 1 < numPolygons {
			__antithesis_instrumentation__.Notify(64904)
			summaryLine += "s"
		} else {
			__antithesis_instrumentation__.Notify(64905)
		}
		__antithesis_instrumentation__.Notify(64883)

		for i := 0; i < numPolygons; i++ {
			__antithesis_instrumentation__.Notify(64906)
			polygon := t.Polygon(i)
			line, err := summary(polygon, false, hasSRID, geopb.ShapeType_Polygon, isGeography, offset+2)
			if err != nil {
				__antithesis_instrumentation__.Notify(64908)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(64909)
			}
			__antithesis_instrumentation__.Notify(64907)

			summaryLine += "\n" + line
		}
		__antithesis_instrumentation__.Notify(64884)

		return summaryLine, nil
	case *geom.GeometryCollection:
		__antithesis_instrumentation__.Notify(64885)
		numGeoms := t.NumGeoms()

		summaryLine += fmt.Sprintf("%s[%s] with %d element", shape.String(), f, numGeoms)
		if 1 < numGeoms {
			__antithesis_instrumentation__.Notify(64910)
			summaryLine += "s"
		} else {
			__antithesis_instrumentation__.Notify(64911)
		}
		__antithesis_instrumentation__.Notify(64886)

		for i := 0; i < numGeoms; i++ {
			__antithesis_instrumentation__.Notify(64912)
			g := t.Geom(i)
			gShape, err := shapeTypeFromGeomT(g)
			if err != nil {
				__antithesis_instrumentation__.Notify(64915)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(64916)
			}
			__antithesis_instrumentation__.Notify(64913)

			line, err := summary(g, false, hasSRID, gShape, isGeography, offset+2)
			if err != nil {
				__antithesis_instrumentation__.Notify(64917)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(64918)
			}
			__antithesis_instrumentation__.Notify(64914)

			summaryLine += "\n" + line
		}
		__antithesis_instrumentation__.Notify(64887)

		return summaryLine, nil
	default:
		__antithesis_instrumentation__.Notify(64888)
		return "", pgerror.Newf(pgcode.InvalidParameterValue, "unsupported geom type: %T", t)
	}
}

func summaryFlag(
	t geom.T, hasBoundingBox bool, hasSRID bool, isGeography bool,
) (f string, err error) {
	__antithesis_instrumentation__.Notify(64919)
	layout := t.Layout()
	if layout.MIndex() != -1 {
		__antithesis_instrumentation__.Notify(64925)
		f += "M"
	} else {
		__antithesis_instrumentation__.Notify(64926)
	}
	__antithesis_instrumentation__.Notify(64920)

	if layout.ZIndex() != -1 {
		__antithesis_instrumentation__.Notify(64927)
		f += "Z"
	} else {
		__antithesis_instrumentation__.Notify(64928)
	}
	__antithesis_instrumentation__.Notify(64921)

	if hasBoundingBox {
		__antithesis_instrumentation__.Notify(64929)
		f += "B"
	} else {
		__antithesis_instrumentation__.Notify(64930)
	}
	__antithesis_instrumentation__.Notify(64922)

	if isGeography {
		__antithesis_instrumentation__.Notify(64931)
		f += "G"
	} else {
		__antithesis_instrumentation__.Notify(64932)
	}
	__antithesis_instrumentation__.Notify(64923)

	if hasSRID {
		__antithesis_instrumentation__.Notify(64933)
		f += "S"
	} else {
		__antithesis_instrumentation__.Notify(64934)
	}
	__antithesis_instrumentation__.Notify(64924)

	return f, nil
}
