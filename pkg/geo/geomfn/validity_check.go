package geomfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geos"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/twpayne/go-geom"
)

type ValidDetail struct {
	IsValid bool

	Reason string

	InvalidLocation geo.Geometry
}

func IsValid(g geo.Geometry) (bool, error) {
	__antithesis_instrumentation__.Notify(63612)
	isValid, err := geos.IsValid(g.EWKB())
	if err != nil {
		__antithesis_instrumentation__.Notify(63614)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(63615)
	}
	__antithesis_instrumentation__.Notify(63613)
	return isValid, nil
}

func IsValidReason(g geo.Geometry) (string, error) {
	__antithesis_instrumentation__.Notify(63616)
	reason, err := geos.IsValidReason(g.EWKB())
	if err != nil {
		__antithesis_instrumentation__.Notify(63618)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(63619)
	}
	__antithesis_instrumentation__.Notify(63617)
	return reason, nil
}

func IsValidDetail(g geo.Geometry, flags int) (ValidDetail, error) {
	__antithesis_instrumentation__.Notify(63620)
	isValid, reason, locEWKB, err := geos.IsValidDetail(g.EWKB(), flags)
	if err != nil {
		__antithesis_instrumentation__.Notify(63623)
		return ValidDetail{}, err
	} else {
		__antithesis_instrumentation__.Notify(63624)
	}
	__antithesis_instrumentation__.Notify(63621)
	var loc geo.Geometry
	if len(locEWKB) > 0 {
		__antithesis_instrumentation__.Notify(63625)
		loc, err = geo.ParseGeometryFromEWKB(locEWKB)
		if err != nil {
			__antithesis_instrumentation__.Notify(63626)
			return ValidDetail{}, err
		} else {
			__antithesis_instrumentation__.Notify(63627)
		}
	} else {
		__antithesis_instrumentation__.Notify(63628)
	}
	__antithesis_instrumentation__.Notify(63622)
	return ValidDetail{
		IsValid:         isValid,
		Reason:          reason,
		InvalidLocation: loc,
	}, nil
}

func IsValidTrajectory(line geo.Geometry) (bool, error) {
	__antithesis_instrumentation__.Notify(63629)
	t, err := line.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(63634)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(63635)
	}
	__antithesis_instrumentation__.Notify(63630)
	lineString, ok := t.(*geom.LineString)
	if !ok {
		__antithesis_instrumentation__.Notify(63636)
		return false, pgerror.Newf(pgcode.InvalidParameterValue, "expected LineString, got %s", line.ShapeType().String())
	} else {
		__antithesis_instrumentation__.Notify(63637)
	}
	__antithesis_instrumentation__.Notify(63631)
	mIndex := t.Layout().MIndex()
	if mIndex < 0 {
		__antithesis_instrumentation__.Notify(63638)
		return false, pgerror.Newf(pgcode.InvalidParameterValue, "LineString does not have M coordinates")
	} else {
		__antithesis_instrumentation__.Notify(63639)
	}
	__antithesis_instrumentation__.Notify(63632)

	coords := lineString.Coords()
	for i := 1; i < len(coords); i++ {
		__antithesis_instrumentation__.Notify(63640)
		if coords[i][mIndex] <= coords[i-1][mIndex] {
			__antithesis_instrumentation__.Notify(63641)
			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(63642)
		}
	}
	__antithesis_instrumentation__.Notify(63633)
	return true, nil
}

func MakeValid(g geo.Geometry) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(63643)
	validEWKB, err := geos.MakeValid(g.EWKB())
	if err != nil {
		__antithesis_instrumentation__.Notify(63645)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(63646)
	}
	__antithesis_instrumentation__.Notify(63644)
	return geo.ParseGeometryFromEWKB(validEWKB)
}

func EqualsExact(lhs, rhs geo.Geometry, epsilon float64) bool {
	__antithesis_instrumentation__.Notify(63647)
	equalsExact, err := geos.EqualsExact(lhs.EWKB(), rhs.EWKB(), epsilon)
	if err != nil {
		__antithesis_instrumentation__.Notify(63649)
		return false
	} else {
		__antithesis_instrumentation__.Notify(63650)
	}
	__antithesis_instrumentation__.Notify(63648)
	return equalsExact
}
