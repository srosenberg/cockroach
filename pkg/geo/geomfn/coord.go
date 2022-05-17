package geomfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"

	"github.com/twpayne/go-geom"
)

func coordAdd(a geom.Coord, b geom.Coord) geom.Coord {
	__antithesis_instrumentation__.Notify(61672)
	return geom.Coord{a.X() + b.X(), a.Y() + b.Y()}
}

func coordSub(a geom.Coord, b geom.Coord) geom.Coord {
	__antithesis_instrumentation__.Notify(61673)
	return geom.Coord{a.X() - b.X(), a.Y() - b.Y()}
}

func coordMul(a geom.Coord, s float64) geom.Coord {
	__antithesis_instrumentation__.Notify(61674)
	return geom.Coord{a.X() * s, a.Y() * s}
}

func coordDet(a geom.Coord, b geom.Coord) float64 {
	__antithesis_instrumentation__.Notify(61675)
	return a.X()*b.Y() - b.X()*a.Y()
}

func coordDot(a geom.Coord, b geom.Coord) float64 {
	__antithesis_instrumentation__.Notify(61676)
	return a.X()*b.X() + a.Y()*b.Y()
}

func coordCross(a geom.Coord, b geom.Coord) float64 {
	__antithesis_instrumentation__.Notify(61677)
	return a.X()*b.Y() - a.Y()*b.X()
}

func coordNorm2(c geom.Coord) float64 {
	__antithesis_instrumentation__.Notify(61678)
	return coordDot(c, c)
}

func coordNorm(c geom.Coord) float64 {
	__antithesis_instrumentation__.Notify(61679)
	return math.Sqrt(coordNorm2(c))
}

func coordEqual(a geom.Coord, b geom.Coord) bool {
	__antithesis_instrumentation__.Notify(61680)
	return a.X() == b.X() && func() bool {
		__antithesis_instrumentation__.Notify(61681)
		return a.Y() == b.Y() == true
	}() == true
}

func coordMag2(c geom.Coord) float64 {
	__antithesis_instrumentation__.Notify(61682)
	return coordDot(c, c)
}
