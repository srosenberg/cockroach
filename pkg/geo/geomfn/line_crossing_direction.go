package geomfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/twpayne/go-geom"
)

type lineCrossingDirection int

type LineCrossingDirectionValue int

const (
	LineNoCross                    LineCrossingDirectionValue = 0
	LineCrossLeft                  LineCrossingDirectionValue = -1
	LineCrossRight                 LineCrossingDirectionValue = 1
	LineMultiCrossToLeft           LineCrossingDirectionValue = -2
	LineMultiCrossToRight          LineCrossingDirectionValue = 2
	LineMultiCrossToSameFirstLeft  LineCrossingDirectionValue = -3
	LineMultiCrossToSameFirstRight LineCrossingDirectionValue = 3

	leftDir            lineCrossingDirection = -1
	noCrossOrCollinear lineCrossingDirection = 0
	rightDir           lineCrossingDirection = 1

	inLeft      lineCrossingDirection = -1
	isCollinear lineCrossingDirection = 0
	inRight     lineCrossingDirection = 1
)

func getPosition(fromSeg, toSeg, givenPoint geom.Coord) lineCrossingDirection {
	__antithesis_instrumentation__.Notify(62295)

	crossProductValueInZ := (toSeg.X()-givenPoint.X())*(fromSeg.Y()-givenPoint.Y()) - (toSeg.Y()-givenPoint.Y())*(fromSeg.X()-givenPoint.X())

	if crossProductValueInZ < 0 {
		__antithesis_instrumentation__.Notify(62298)
		return inLeft
	} else {
		__antithesis_instrumentation__.Notify(62299)
	}
	__antithesis_instrumentation__.Notify(62296)
	if crossProductValueInZ > 0 {
		__antithesis_instrumentation__.Notify(62300)
		return inRight
	} else {
		__antithesis_instrumentation__.Notify(62301)
	}
	__antithesis_instrumentation__.Notify(62297)
	return isCollinear
}

func getSegCrossDirection(fromSeg1, toSeg1, fromSeg2, toSeg2 geom.Coord) lineCrossingDirection {
	__antithesis_instrumentation__.Notify(62302)
	posF1FromSeg2 := getPosition(fromSeg2, toSeg2, fromSeg1)
	posT1FromSeg2 := getPosition(fromSeg2, toSeg2, toSeg1)

	posF2FromSeg1 := getPosition(fromSeg1, toSeg1, fromSeg2)
	posT2FromSeg1 := getPosition(fromSeg1, toSeg1, toSeg2)

	if posF1FromSeg2 == posT1FromSeg2 || func() bool {
		__antithesis_instrumentation__.Notify(62304)
		return posF2FromSeg1 == posT2FromSeg1 == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(62305)
		return posT1FromSeg2 == isCollinear == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(62306)
		return posT2FromSeg1 == isCollinear == true
	}() == true {
		__antithesis_instrumentation__.Notify(62307)
		return noCrossOrCollinear
	} else {
		__antithesis_instrumentation__.Notify(62308)
	}
	__antithesis_instrumentation__.Notify(62303)
	return posT2FromSeg1
}

func LineCrossingDirection(geometry1, geometry2 geo.Geometry) (LineCrossingDirectionValue, error) {
	__antithesis_instrumentation__.Notify(62309)

	t1, err1 := geometry1.AsGeomT()
	if err1 != nil {
		__antithesis_instrumentation__.Notify(62320)
		return 0, err1
	} else {
		__antithesis_instrumentation__.Notify(62321)
	}
	__antithesis_instrumentation__.Notify(62310)

	t2, err2 := geometry2.AsGeomT()
	if err2 != nil {
		__antithesis_instrumentation__.Notify(62322)
		return 0, err2
	} else {
		__antithesis_instrumentation__.Notify(62323)
	}
	__antithesis_instrumentation__.Notify(62311)

	g1, ok1 := t1.(*geom.LineString)
	g2, ok2 := t2.(*geom.LineString)

	if !ok1 || func() bool {
		__antithesis_instrumentation__.Notify(62324)
		return !ok2 == true
	}() == true {
		__antithesis_instrumentation__.Notify(62325)
		return 0, pgerror.Newf(pgcode.InvalidParameterValue, "arguments must be LINESTRING")
	} else {
		__antithesis_instrumentation__.Notify(62326)
	}
	__antithesis_instrumentation__.Notify(62312)

	line1, line2 := g1.Coords(), g2.Coords()

	firstSegCrossDirection := noCrossOrCollinear
	countDirection := make(map[lineCrossingDirection]int)

	for idx2 := 1; idx2 < len(line2); idx2++ {
		__antithesis_instrumentation__.Notify(62327)
		fromSeg2, toSeg2 := line2[idx2-1], line2[idx2]

		for idx1 := 1; idx1 < len(line1); idx1++ {
			__antithesis_instrumentation__.Notify(62328)
			fromSeg1, toSeg1 := line1[idx1-1], line1[idx1]

			segCrossDirection := getSegCrossDirection(fromSeg1, toSeg1, fromSeg2, toSeg2)

			countDirection[segCrossDirection]++

			if firstSegCrossDirection == noCrossOrCollinear {
				__antithesis_instrumentation__.Notify(62329)
				firstSegCrossDirection = segCrossDirection
			} else {
				__antithesis_instrumentation__.Notify(62330)
			}
		}
	}
	__antithesis_instrumentation__.Notify(62313)

	if countDirection[leftDir] == 0 && func() bool {
		__antithesis_instrumentation__.Notify(62331)
		return countDirection[rightDir] == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(62332)
		return LineNoCross, nil
	} else {
		__antithesis_instrumentation__.Notify(62333)
	}
	__antithesis_instrumentation__.Notify(62314)
	if countDirection[leftDir] == 1 && func() bool {
		__antithesis_instrumentation__.Notify(62334)
		return countDirection[rightDir] == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(62335)
		return LineCrossLeft, nil
	} else {
		__antithesis_instrumentation__.Notify(62336)
	}
	__antithesis_instrumentation__.Notify(62315)
	if countDirection[leftDir] == 0 && func() bool {
		__antithesis_instrumentation__.Notify(62337)
		return countDirection[rightDir] == 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(62338)
		return LineCrossRight, nil
	} else {
		__antithesis_instrumentation__.Notify(62339)
	}
	__antithesis_instrumentation__.Notify(62316)
	if countDirection[leftDir]-countDirection[rightDir] == 1 {
		__antithesis_instrumentation__.Notify(62340)
		return LineMultiCrossToLeft, nil
	} else {
		__antithesis_instrumentation__.Notify(62341)
	}
	__antithesis_instrumentation__.Notify(62317)
	if countDirection[rightDir]-countDirection[leftDir] == 1 {
		__antithesis_instrumentation__.Notify(62342)
		return LineMultiCrossToRight, nil
	} else {
		__antithesis_instrumentation__.Notify(62343)
	}
	__antithesis_instrumentation__.Notify(62318)
	if countDirection[leftDir] == countDirection[rightDir] {
		__antithesis_instrumentation__.Notify(62344)

		if firstSegCrossDirection == leftDir {
			__antithesis_instrumentation__.Notify(62346)
			return LineMultiCrossToSameFirstLeft, nil
		} else {
			__antithesis_instrumentation__.Notify(62347)
		}
		__antithesis_instrumentation__.Notify(62345)
		return LineMultiCrossToSameFirstRight, nil
	} else {
		__antithesis_instrumentation__.Notify(62348)
	}
	__antithesis_instrumentation__.Notify(62319)

	return LineNoCross, nil
}
