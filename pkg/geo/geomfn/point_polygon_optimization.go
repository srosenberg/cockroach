package geomfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/errors"
)

type PointPolygonControlFlowType int

const (
	PPCFCheckNextPolygon PointPolygonControlFlowType = iota

	PPCFSkipToNextPoint

	PPCFReturnTrue
)

type PointInPolygonEventListener interface {
	OnPointIntersectsPolygon(strictlyInside bool) PointPolygonControlFlowType

	ExitIfPointDoesNotIntersect() bool

	AfterPointPolygonLoops() bool
}

type intersectsPIPEventListener struct{}

func (el *intersectsPIPEventListener) OnPointIntersectsPolygon(
	strictlyInside bool,
) PointPolygonControlFlowType {
	__antithesis_instrumentation__.Notify(62768)

	return PPCFReturnTrue
}

func (el *intersectsPIPEventListener) ExitIfPointDoesNotIntersect() bool {
	__antithesis_instrumentation__.Notify(62769)
	return false
}

func (el *intersectsPIPEventListener) AfterPointPolygonLoops() bool {
	__antithesis_instrumentation__.Notify(62770)
	return false
}

var _ PointInPolygonEventListener = (*intersectsPIPEventListener)(nil)

func newIntersectsPIPEventListener() *intersectsPIPEventListener {
	__antithesis_instrumentation__.Notify(62771)
	return &intersectsPIPEventListener{}
}

type coveredByPIPEventListener struct {
	intersectsOnce bool
}

func (el *coveredByPIPEventListener) OnPointIntersectsPolygon(
	strictlyInside bool,
) PointPolygonControlFlowType {
	__antithesis_instrumentation__.Notify(62772)

	el.intersectsOnce = true
	return PPCFSkipToNextPoint
}

func (el *coveredByPIPEventListener) ExitIfPointDoesNotIntersect() bool {
	__antithesis_instrumentation__.Notify(62773)

	return true
}

func (el *coveredByPIPEventListener) AfterPointPolygonLoops() bool {
	__antithesis_instrumentation__.Notify(62774)
	return el.intersectsOnce
}

var _ PointInPolygonEventListener = (*coveredByPIPEventListener)(nil)

func newCoveredByPIPEventListener() *coveredByPIPEventListener {
	__antithesis_instrumentation__.Notify(62775)
	return &coveredByPIPEventListener{intersectsOnce: false}
}

type withinPIPEventListener struct {
	insideOnce bool
}

func (el *withinPIPEventListener) OnPointIntersectsPolygon(
	strictlyInside bool,
) PointPolygonControlFlowType {
	__antithesis_instrumentation__.Notify(62776)

	if el.insideOnce {
		__antithesis_instrumentation__.Notify(62779)
		return PPCFSkipToNextPoint
	} else {
		__antithesis_instrumentation__.Notify(62780)
	}
	__antithesis_instrumentation__.Notify(62777)
	if strictlyInside {
		__antithesis_instrumentation__.Notify(62781)
		el.insideOnce = true
		return PPCFSkipToNextPoint
	} else {
		__antithesis_instrumentation__.Notify(62782)
	}
	__antithesis_instrumentation__.Notify(62778)
	return PPCFCheckNextPolygon
}

func (el *withinPIPEventListener) ExitIfPointDoesNotIntersect() bool {
	__antithesis_instrumentation__.Notify(62783)

	return true
}

func (el *withinPIPEventListener) AfterPointPolygonLoops() bool {
	__antithesis_instrumentation__.Notify(62784)
	return el.insideOnce
}

var _ PointInPolygonEventListener = (*withinPIPEventListener)(nil)

func newWithinPIPEventListener() *withinPIPEventListener {
	__antithesis_instrumentation__.Notify(62785)
	return &withinPIPEventListener{insideOnce: false}
}

func PointKindIntersectsPolygonKind(
	pointKind geo.Geometry, polygonKind geo.Geometry,
) (bool, error) {
	__antithesis_instrumentation__.Notify(62786)
	return pointKindRelatesToPolygonKind(pointKind, polygonKind, newIntersectsPIPEventListener())
}

func PointKindCoveredByPolygonKind(pointKind geo.Geometry, polygonKind geo.Geometry) (bool, error) {
	__antithesis_instrumentation__.Notify(62787)
	return pointKindRelatesToPolygonKind(pointKind, polygonKind, newCoveredByPIPEventListener())
}

func PointKindWithinPolygonKind(pointKind geo.Geometry, polygonKind geo.Geometry) (bool, error) {
	__antithesis_instrumentation__.Notify(62788)
	return pointKindRelatesToPolygonKind(pointKind, polygonKind, newWithinPIPEventListener())
}

func pointKindRelatesToPolygonKind(
	pointKind geo.Geometry, polygonKind geo.Geometry, eventListener PointInPolygonEventListener,
) (bool, error) {
	__antithesis_instrumentation__.Notify(62789)

	if BoundingBoxHasNaNCoordinates(pointKind) || func() bool {
		__antithesis_instrumentation__.Notify(62794)
		return BoundingBoxHasNaNCoordinates(polygonKind) == true
	}() == true {
		__antithesis_instrumentation__.Notify(62795)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(62796)
	}
	__antithesis_instrumentation__.Notify(62790)
	pointKindBaseT, err := pointKind.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(62797)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(62798)
	}
	__antithesis_instrumentation__.Notify(62791)
	polygonKindBaseT, err := polygonKind.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(62799)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(62800)
	}
	__antithesis_instrumentation__.Notify(62792)
	pointKindIterator := geo.NewGeomTIterator(pointKindBaseT, geo.EmptyBehaviorOmit)
	polygonKindIterator := geo.NewGeomTIterator(polygonKindBaseT, geo.EmptyBehaviorOmit)

pointOuterLoop:
	for {
		__antithesis_instrumentation__.Notify(62801)
		point, hasPoint, err := pointKindIterator.Next()
		if err != nil {
			__antithesis_instrumentation__.Notify(62805)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(62806)
		}
		__antithesis_instrumentation__.Notify(62802)
		if !hasPoint {
			__antithesis_instrumentation__.Notify(62807)
			break
		} else {
			__antithesis_instrumentation__.Notify(62808)
		}
		__antithesis_instrumentation__.Notify(62803)

		polygonKindIterator.Reset()
		curIntersects := false
		for {
			__antithesis_instrumentation__.Notify(62809)
			polygon, hasPolygon, err := polygonKindIterator.Next()
			if err != nil {
				__antithesis_instrumentation__.Notify(62813)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(62814)
			}
			__antithesis_instrumentation__.Notify(62810)
			if !hasPolygon {
				__antithesis_instrumentation__.Notify(62815)
				break
			} else {
				__antithesis_instrumentation__.Notify(62816)
			}
			__antithesis_instrumentation__.Notify(62811)
			pointSide, err := findPointSideOfPolygon(point, polygon)
			if err != nil {
				__antithesis_instrumentation__.Notify(62817)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(62818)
			}
			__antithesis_instrumentation__.Notify(62812)
			switch pointSide {
			case insideLinearRing, onLinearRing:
				__antithesis_instrumentation__.Notify(62819)
				curIntersects = true
				strictlyInside := pointSide == insideLinearRing
				switch eventListener.OnPointIntersectsPolygon(strictlyInside) {
				case PPCFCheckNextPolygon:
					__antithesis_instrumentation__.Notify(62822)
				case PPCFSkipToNextPoint:
					__antithesis_instrumentation__.Notify(62823)
					continue pointOuterLoop
				case PPCFReturnTrue:
					__antithesis_instrumentation__.Notify(62824)
					return true, nil
				default:
					__antithesis_instrumentation__.Notify(62825)
				}
			case outsideLinearRing:
				__antithesis_instrumentation__.Notify(62820)
			default:
				__antithesis_instrumentation__.Notify(62821)
				return false, errors.AssertionFailedf("findPointSideOfPolygon returned unknown linearRingSide %d", pointSide)
			}
		}
		__antithesis_instrumentation__.Notify(62804)
		if !curIntersects && func() bool {
			__antithesis_instrumentation__.Notify(62826)
			return eventListener.ExitIfPointDoesNotIntersect() == true
		}() == true {
			__antithesis_instrumentation__.Notify(62827)
			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(62828)
		}
	}
	__antithesis_instrumentation__.Notify(62793)
	return eventListener.AfterPointPolygonLoops(), nil
}
