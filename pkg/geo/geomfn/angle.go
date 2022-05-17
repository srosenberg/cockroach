package geomfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"

	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/twpayne/go-geom"
)

func Angle(g1, g2, g3, g4 geo.Geometry) (*float64, error) {
	__antithesis_instrumentation__.Notify(61120)
	if g4.Empty() {
		__antithesis_instrumentation__.Notify(61131)
		g1, g2, g3, g4 = g2, g1, g2, g3
	} else {
		__antithesis_instrumentation__.Notify(61132)
	}
	__antithesis_instrumentation__.Notify(61121)

	if g1.SRID() != g2.SRID() {
		__antithesis_instrumentation__.Notify(61133)
		return nil, geo.NewMismatchingSRIDsError(g1.SpatialObject(), g2.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(61134)
	}
	__antithesis_instrumentation__.Notify(61122)
	if g1.SRID() != g3.SRID() {
		__antithesis_instrumentation__.Notify(61135)
		return nil, geo.NewMismatchingSRIDsError(g1.SpatialObject(), g3.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(61136)
	}
	__antithesis_instrumentation__.Notify(61123)
	if g1.SRID() != g4.SRID() {
		__antithesis_instrumentation__.Notify(61137)
		return nil, geo.NewMismatchingSRIDsError(g1.SpatialObject(), g4.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(61138)
	}
	__antithesis_instrumentation__.Notify(61124)

	t1, err := g1.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(61139)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(61140)
	}
	__antithesis_instrumentation__.Notify(61125)
	t2, err := g2.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(61141)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(61142)
	}
	__antithesis_instrumentation__.Notify(61126)
	t3, err := g3.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(61143)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(61144)
	}
	__antithesis_instrumentation__.Notify(61127)
	t4, err := g4.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(61145)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(61146)
	}
	__antithesis_instrumentation__.Notify(61128)

	p1, p1ok := t1.(*geom.Point)
	p2, p2ok := t2.(*geom.Point)
	p3, p3ok := t3.(*geom.Point)
	p4, p4ok := t4.(*geom.Point)

	if !p1ok || func() bool {
		__antithesis_instrumentation__.Notify(61147)
		return !p2ok == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(61148)
		return !p3ok == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(61149)
		return !p4ok == true
	}() == true {
		__antithesis_instrumentation__.Notify(61150)
		return nil, pgerror.Newf(pgcode.InvalidParameterValue, "arguments must be POINT geometries")
	} else {
		__antithesis_instrumentation__.Notify(61151)
	}
	__antithesis_instrumentation__.Notify(61129)
	if p1.Empty() || func() bool {
		__antithesis_instrumentation__.Notify(61152)
		return p2.Empty() == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(61153)
		return p3.Empty() == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(61154)
		return p4.Empty() == true
	}() == true {
		__antithesis_instrumentation__.Notify(61155)
		return nil, pgerror.Newf(pgcode.InvalidParameterValue, "received EMPTY geometry")
	} else {
		__antithesis_instrumentation__.Notify(61156)
	}
	__antithesis_instrumentation__.Notify(61130)

	return angleFromCoords(p1.Coords(), p2.Coords(), p3.Coords(), p4.Coords()), nil
}

func AngleLineString(g1, g2 geo.Geometry) (*float64, error) {
	__antithesis_instrumentation__.Notify(61157)
	if g1.SRID() != g2.SRID() {
		__antithesis_instrumentation__.Notify(61162)
		return nil, geo.NewMismatchingSRIDsError(g1.SpatialObject(), g2.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(61163)
	}
	__antithesis_instrumentation__.Notify(61158)
	t1, err := g1.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(61164)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(61165)
	}
	__antithesis_instrumentation__.Notify(61159)
	t2, err := g2.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(61166)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(61167)
	}
	__antithesis_instrumentation__.Notify(61160)
	l1, l1ok := t1.(*geom.LineString)
	l2, l2ok := t2.(*geom.LineString)
	if !l1ok || func() bool {
		__antithesis_instrumentation__.Notify(61168)
		return !l2ok == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(61169)
		return l1.Empty() == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(61170)
		return l2.Empty() == true
	}() == true {
		__antithesis_instrumentation__.Notify(61171)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(61172)
	}
	__antithesis_instrumentation__.Notify(61161)
	return angleFromCoords(
		l1.Coord(0), l1.Coord(l1.NumCoords()-1), l2.Coord(0), l2.Coord(l2.NumCoords()-1)), nil
}

func angleFromCoords(c1, c2, c3, c4 geom.Coord) *float64 {
	__antithesis_instrumentation__.Notify(61173)
	a := coordSub(c2, c1)
	b := coordSub(c4, c3)
	if (a.X() == 0 && func() bool {
		__antithesis_instrumentation__.Notify(61176)
		return a.Y() == 0 == true
	}() == true) || func() bool {
		__antithesis_instrumentation__.Notify(61177)
		return (b.X() == 0 && func() bool {
			__antithesis_instrumentation__.Notify(61178)
			return b.Y() == 0 == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(61179)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(61180)
	}
	__antithesis_instrumentation__.Notify(61174)

	angle := math.Atan2(-coordDet(a, b), coordDot(a, b))

	if angle == -0.0 {
		__antithesis_instrumentation__.Notify(61181)
		angle = 0.0
	} else {
		__antithesis_instrumentation__.Notify(61182)
		if angle < 0 {
			__antithesis_instrumentation__.Notify(61183)
			angle += 2 * math.Pi
		} else {
			__antithesis_instrumentation__.Notify(61184)
		}
	}
	__antithesis_instrumentation__.Notify(61175)
	return &angle
}
