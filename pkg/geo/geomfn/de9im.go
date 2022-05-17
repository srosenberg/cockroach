package geomfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geos"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/util"
)

func Relate(a geo.Geometry, b geo.Geometry) (string, error) {
	__antithesis_instrumentation__.Notify(61683)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(61685)
		return "", geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(61686)
	}
	__antithesis_instrumentation__.Notify(61684)
	return geos.Relate(a.EWKB(), b.EWKB())
}

func RelateBoundaryNodeRule(a, b geo.Geometry, bnr int) (string, error) {
	__antithesis_instrumentation__.Notify(61687)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(61689)
		return "", geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(61690)
	}
	__antithesis_instrumentation__.Notify(61688)
	return geos.RelateBoundaryNodeRule(a.EWKB(), b.EWKB(), bnr)
}

func RelatePattern(a geo.Geometry, b geo.Geometry, pattern string) (bool, error) {
	__antithesis_instrumentation__.Notify(61691)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(61693)
		return false, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(61694)
	}
	__antithesis_instrumentation__.Notify(61692)
	return geos.RelatePattern(a.EWKB(), b.EWKB(), pattern)
}

func MatchesDE9IM(relation string, pattern string) (bool, error) {
	__antithesis_instrumentation__.Notify(61695)
	if len(relation) != 9 {
		__antithesis_instrumentation__.Notify(61699)
		return false, pgerror.Newf(pgcode.InvalidParameterValue, "relation %q should be of length 9", relation)
	} else {
		__antithesis_instrumentation__.Notify(61700)
	}
	__antithesis_instrumentation__.Notify(61696)
	if len(pattern) != 9 {
		__antithesis_instrumentation__.Notify(61701)
		return false, pgerror.Newf(pgcode.InvalidParameterValue, "pattern %q should be of length 9", pattern)
	} else {
		__antithesis_instrumentation__.Notify(61702)
	}
	__antithesis_instrumentation__.Notify(61697)
	for i := 0; i < len(relation); i++ {
		__antithesis_instrumentation__.Notify(61703)
		matches, err := relationByteMatchesPatternByte(relation[i], pattern[i])
		if err != nil {
			__antithesis_instrumentation__.Notify(61705)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(61706)
		}
		__antithesis_instrumentation__.Notify(61704)
		if !matches {
			__antithesis_instrumentation__.Notify(61707)
			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(61708)
		}
	}
	__antithesis_instrumentation__.Notify(61698)
	return true, nil
}

func relationByteMatchesPatternByte(r byte, p byte) (bool, error) {
	__antithesis_instrumentation__.Notify(61709)
	switch util.ToLowerSingleByte(p) {
	case '*':
		__antithesis_instrumentation__.Notify(61711)
		return true, nil
	case 't':
		__antithesis_instrumentation__.Notify(61712)
		if r < '0' || func() bool {
			__antithesis_instrumentation__.Notify(61716)
			return r > '2' == true
		}() == true {
			__antithesis_instrumentation__.Notify(61717)
			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(61718)
		}
	case 'f':
		__antithesis_instrumentation__.Notify(61713)
		if util.ToLowerSingleByte(r) != 'f' {
			__antithesis_instrumentation__.Notify(61719)
			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(61720)
		}
	case '0', '1', '2':
		__antithesis_instrumentation__.Notify(61714)
		return r == p, nil
	default:
		__antithesis_instrumentation__.Notify(61715)
		return false, pgerror.Newf(pgcode.InvalidParameterValue, "unrecognized pattern character: %s", string(p))
	}
	__antithesis_instrumentation__.Notify(61710)
	return true, nil
}
