package geomfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"

	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/errors"
	"github.com/twpayne/go-geom"
)

type AffineMatrix [][]float64

func Affine(g geo.Geometry, m AffineMatrix) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(61018)
	if g.Empty() {
		__antithesis_instrumentation__.Notify(61022)
		return g, nil
	} else {
		__antithesis_instrumentation__.Notify(61023)
	}
	__antithesis_instrumentation__.Notify(61019)

	t, err := g.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(61024)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(61025)
	}
	__antithesis_instrumentation__.Notify(61020)

	newT, err := affine(t, m)
	if err != nil {
		__antithesis_instrumentation__.Notify(61026)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(61027)
	}
	__antithesis_instrumentation__.Notify(61021)
	return geo.MakeGeometryFromGeomT(newT)
}

func affine(t geom.T, m AffineMatrix) (geom.T, error) {
	__antithesis_instrumentation__.Notify(61028)
	return applyOnCoordsForGeomT(t, func(l geom.Layout, dst, src []float64) error {
		__antithesis_instrumentation__.Notify(61029)
		var z float64
		if l.ZIndex() != -1 {
			__antithesis_instrumentation__.Notify(61033)
			z = src[l.ZIndex()]
		} else {
			__antithesis_instrumentation__.Notify(61034)
		}
		__antithesis_instrumentation__.Notify(61030)
		newX := m[0][0]*src[0] + m[0][1]*src[1] + m[0][2]*z + m[0][3]
		newY := m[1][0]*src[0] + m[1][1]*src[1] + m[1][2]*z + m[1][3]
		newZ := m[2][0]*src[0] + m[2][1]*src[1] + m[2][2]*z + m[2][3]

		dst[0] = newX
		dst[1] = newY
		if l.ZIndex() != -1 {
			__antithesis_instrumentation__.Notify(61035)
			dst[2] = newZ
		} else {
			__antithesis_instrumentation__.Notify(61036)
		}
		__antithesis_instrumentation__.Notify(61031)
		if l.MIndex() != -1 {
			__antithesis_instrumentation__.Notify(61037)
			dst[l.MIndex()] = src[l.MIndex()]
		} else {
			__antithesis_instrumentation__.Notify(61038)
		}
		__antithesis_instrumentation__.Notify(61032)
		return nil
	})
}

func Translate(g geo.Geometry, deltas []float64) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(61039)
	if g.Empty() {
		__antithesis_instrumentation__.Notify(61043)
		return g, nil
	} else {
		__antithesis_instrumentation__.Notify(61044)
	}
	__antithesis_instrumentation__.Notify(61040)

	t, err := g.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(61045)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(61046)
	}
	__antithesis_instrumentation__.Notify(61041)

	newT, err := translate(t, deltas)
	if err != nil {
		__antithesis_instrumentation__.Notify(61047)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(61048)
	}
	__antithesis_instrumentation__.Notify(61042)
	return geo.MakeGeometryFromGeomT(newT)
}

func translate(t geom.T, deltas []float64) (geom.T, error) {
	__antithesis_instrumentation__.Notify(61049)
	if t.Layout().Stride() != len(deltas) {
		__antithesis_instrumentation__.Notify(61052)
		err := geom.ErrStrideMismatch{
			Got:  len(deltas),
			Want: t.Layout().Stride(),
		}
		return nil, pgerror.WithCandidateCode(
			errors.Wrap(err, "translating coordinates"),
			pgcode.InvalidParameterValue,
		)
	} else {
		__antithesis_instrumentation__.Notify(61053)
	}
	__antithesis_instrumentation__.Notify(61050)
	var zOff float64
	if t.Layout().ZIndex() != -1 {
		__antithesis_instrumentation__.Notify(61054)
		zOff = deltas[t.Layout().ZIndex()]
	} else {
		__antithesis_instrumentation__.Notify(61055)
	}
	__antithesis_instrumentation__.Notify(61051)
	return affine(
		t,
		AffineMatrix([][]float64{
			{1, 0, 0, deltas[0]},
			{0, 1, 0, deltas[1]},
			{0, 0, 1, zOff},
			{0, 0, 0, 1},
		}),
	)
}

func Scale(g geo.Geometry, factors []float64) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(61056)
	var zFactor float64
	if len(factors) > 2 {
		__antithesis_instrumentation__.Notify(61058)
		zFactor = factors[2]
	} else {
		__antithesis_instrumentation__.Notify(61059)
	}
	__antithesis_instrumentation__.Notify(61057)
	return Affine(
		g,
		AffineMatrix([][]float64{
			{factors[0], 0, 0, 0},
			{0, factors[1], 0, 0},
			{0, 0, zFactor, 0},
			{0, 0, 0, 1},
		}),
	)
}

func ScaleRelativeToOrigin(
	g geo.Geometry, factor geo.Geometry, origin geo.Geometry,
) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(61060)
	if g.Empty() {
		__antithesis_instrumentation__.Notify(61075)
		return g, nil
	} else {
		__antithesis_instrumentation__.Notify(61076)
	}
	__antithesis_instrumentation__.Notify(61061)

	t, err := g.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(61077)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(61078)
	}
	__antithesis_instrumentation__.Notify(61062)

	factorG, err := factor.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(61079)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(61080)
	}
	__antithesis_instrumentation__.Notify(61063)

	factorPointG, ok := factorG.(*geom.Point)
	if !ok {
		__antithesis_instrumentation__.Notify(61081)
		return geo.Geometry{}, pgerror.Newf(pgcode.InvalidParameterValue, "the scaling factor must be a Point")
	} else {
		__antithesis_instrumentation__.Notify(61082)
	}
	__antithesis_instrumentation__.Notify(61064)

	originG, err := origin.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(61083)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(61084)
	}
	__antithesis_instrumentation__.Notify(61065)

	originPointG, ok := originG.(*geom.Point)
	if !ok {
		__antithesis_instrumentation__.Notify(61085)
		return geo.Geometry{}, pgerror.Newf(pgcode.InvalidParameterValue, "the false origin must be a Point")
	} else {
		__antithesis_instrumentation__.Notify(61086)
	}
	__antithesis_instrumentation__.Notify(61066)

	if factorG.Stride() != originG.Stride() {
		__antithesis_instrumentation__.Notify(61087)
		err := geom.ErrStrideMismatch{
			Got:  factorG.Stride(),
			Want: originG.Stride(),
		}
		return geo.Geometry{}, pgerror.WithCandidateCode(
			errors.Wrap(err, "number of dimensions for the scaling factor and origin must be equal"),
			pgcode.InvalidParameterValue,
		)
	} else {
		__antithesis_instrumentation__.Notify(61088)
	}
	__antithesis_instrumentation__.Notify(61067)

	if len(originPointG.FlatCoords()) < 2 {
		__antithesis_instrumentation__.Notify(61089)
		return geo.Geometry{}, pgerror.Newf(pgcode.InvalidParameterValue, "the origin must have at least 2 coordinates")
	} else {
		__antithesis_instrumentation__.Notify(61090)
	}
	__antithesis_instrumentation__.Notify(61068)

	offsetDeltas := make([]float64, 0, 3)
	offsetDeltas = append(offsetDeltas, -originPointG.X(), -originPointG.Y())
	if originG.Layout().ZIndex() != -1 {
		__antithesis_instrumentation__.Notify(61091)
		offsetDeltas = append(offsetDeltas, -originPointG.Z())
	} else {
		__antithesis_instrumentation__.Notify(61092)
	}
	__antithesis_instrumentation__.Notify(61069)
	retT, err := translate(t, offsetDeltas)
	if err != nil {
		__antithesis_instrumentation__.Notify(61093)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(61094)
	}
	__antithesis_instrumentation__.Notify(61070)

	var xFactor, yFactor, zFactor float64
	zFactor = 1
	if len(factorPointG.FlatCoords()) < 2 {
		__antithesis_instrumentation__.Notify(61095)

		xFactor, yFactor = 0, 0
	} else {
		__antithesis_instrumentation__.Notify(61096)
		xFactor, yFactor = factorPointG.X(), factorPointG.Y()
		if factorPointG.Layout().ZIndex() != -1 {
			__antithesis_instrumentation__.Notify(61097)
			zFactor = factorPointG.Z()
		} else {
			__antithesis_instrumentation__.Notify(61098)
		}
	}
	__antithesis_instrumentation__.Notify(61071)

	retT, err = affine(
		retT,
		AffineMatrix([][]float64{
			{xFactor, 0, 0, 0},
			{0, yFactor, 0, 0},
			{0, 0, zFactor, 0},
			{0, 0, 0, 1},
		}),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(61099)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(61100)
	}
	__antithesis_instrumentation__.Notify(61072)

	for i := range offsetDeltas {
		__antithesis_instrumentation__.Notify(61101)
		offsetDeltas[i] = -offsetDeltas[i]
	}
	__antithesis_instrumentation__.Notify(61073)
	retT, err = translate(retT, offsetDeltas)
	if err != nil {
		__antithesis_instrumentation__.Notify(61102)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(61103)
	}
	__antithesis_instrumentation__.Notify(61074)

	return geo.MakeGeometryFromGeomT(retT)
}

func Rotate(g geo.Geometry, rotRadians float64) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(61104)
	return Affine(
		g,
		AffineMatrix([][]float64{
			{math.Cos(rotRadians), -math.Sin(rotRadians), 0, 0},
			{math.Sin(rotRadians), math.Cos(rotRadians), 0, 0},
			{0, 0, 1, 0},
			{0, 0, 0, 1},
		}),
	)
}

var ErrPointOriginEmpty = pgerror.Newf(pgcode.InvalidParameterValue, "origin is an empty point")

func RotateWithPointOrigin(
	g geo.Geometry, rotRadians float64, pointOrigin geo.Geometry,
) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(61105)
	if pointOrigin.ShapeType() != geopb.ShapeType_Point {
		__antithesis_instrumentation__.Notify(61109)
		return g, pgerror.Newf(pgcode.InvalidParameterValue, "origin is not a POINT")
	} else {
		__antithesis_instrumentation__.Notify(61110)
	}
	__antithesis_instrumentation__.Notify(61106)
	t, err := pointOrigin.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(61111)
		return g, err
	} else {
		__antithesis_instrumentation__.Notify(61112)
	}
	__antithesis_instrumentation__.Notify(61107)
	if t.Empty() {
		__antithesis_instrumentation__.Notify(61113)
		return g, ErrPointOriginEmpty
	} else {
		__antithesis_instrumentation__.Notify(61114)
	}
	__antithesis_instrumentation__.Notify(61108)

	return RotateWithXY(g, rotRadians, t.FlatCoords()[0], t.FlatCoords()[1])
}

func RotateWithXY(g geo.Geometry, rotRadians, x, y float64) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(61115)
	cos, sin := math.Cos(rotRadians), math.Sin(rotRadians)
	return Affine(
		g,
		AffineMatrix{
			{cos, -sin, 0, x - x*cos + y*sin},
			{sin, cos, 0, y - x*sin - y*cos},
			{0, 0, 1, 0},
			{0, 0, 0, 1},
		},
	)
}

func RotateX(g geo.Geometry, rotRadians float64) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(61116)
	cos, sin := math.Cos(rotRadians), math.Sin(rotRadians)
	return Affine(
		g,
		AffineMatrix{
			{1, 0, 0, 0},
			{0, cos, -sin, 0},
			{0, sin, cos, 0},
			{0, 0, 0, 1},
		},
	)
}

func RotateY(g geo.Geometry, rotRadians float64) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(61117)
	cos, sin := math.Cos(rotRadians), math.Sin(rotRadians)
	return Affine(
		g,
		AffineMatrix{
			{cos, 0, sin, 0},
			{0, 1, 0, 0},
			{-sin, 0, cos, 0},
			{0, 0, 0, 1},
		},
	)
}

func RotateZ(g geo.Geometry, rotRadians float64) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(61118)
	cos, sin := math.Cos(rotRadians), math.Sin(rotRadians)
	return Affine(
		g,
		AffineMatrix{
			{cos, -sin, 0, 0},
			{sin, cos, 0, 0},
			{0, 0, 1, 0},
			{0, 0, 0, 1},
		},
	)
}

func TransScale(g geo.Geometry, deltaX, deltaY, xFactor, yFactor float64) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(61119)
	return Affine(g,
		AffineMatrix([][]float64{
			{xFactor, 0, 0, xFactor * deltaX},
			{0, yFactor, 0, yFactor * deltaY},
			{0, 0, 1, 0},
			{0, 0, 0, 1},
		}))
}
