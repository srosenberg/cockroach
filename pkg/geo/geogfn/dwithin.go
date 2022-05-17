package geogfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/golang/geo/s1"
)

func DWithin(
	a geo.Geography,
	b geo.Geography,
	distance float64,
	useSphereOrSpheroid UseSphereOrSpheroid,
	exclusivity geo.FnExclusivity,
) (bool, error) {
	__antithesis_instrumentation__.Notify(59681)
	if a.SRID() != b.SRID() {
		__antithesis_instrumentation__.Notify(59691)
		return false, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(59692)
	}
	__antithesis_instrumentation__.Notify(59682)
	if distance < 0 {
		__antithesis_instrumentation__.Notify(59693)
		return false, pgerror.Newf(pgcode.InvalidParameterValue, "dwithin distance cannot be less than zero")
	} else {
		__antithesis_instrumentation__.Notify(59694)
	}
	__antithesis_instrumentation__.Notify(59683)
	spheroid, err := a.Spheroid()
	if err != nil {
		__antithesis_instrumentation__.Notify(59695)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(59696)
	}
	__antithesis_instrumentation__.Notify(59684)

	angleToExpand := s1.Angle(distance / spheroid.SphereRadius)
	if useSphereOrSpheroid == UseSpheroid {
		__antithesis_instrumentation__.Notify(59697)
		angleToExpand *= (1 + SpheroidErrorFraction)
	} else {
		__antithesis_instrumentation__.Notify(59698)
	}
	__antithesis_instrumentation__.Notify(59685)
	if !a.BoundingCap().Expanded(angleToExpand).Intersects(b.BoundingCap()) {
		__antithesis_instrumentation__.Notify(59699)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(59700)
	}
	__antithesis_instrumentation__.Notify(59686)

	aRegions, err := a.AsS2(geo.EmptyBehaviorError)
	if err != nil {
		__antithesis_instrumentation__.Notify(59701)
		if geo.IsEmptyGeometryError(err) {
			__antithesis_instrumentation__.Notify(59703)
			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(59704)
		}
		__antithesis_instrumentation__.Notify(59702)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(59705)
	}
	__antithesis_instrumentation__.Notify(59687)
	bRegions, err := b.AsS2(geo.EmptyBehaviorError)
	if err != nil {
		__antithesis_instrumentation__.Notify(59706)
		if geo.IsEmptyGeometryError(err) {
			__antithesis_instrumentation__.Notify(59708)
			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(59709)
		}
		__antithesis_instrumentation__.Notify(59707)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(59710)
	}
	__antithesis_instrumentation__.Notify(59688)
	maybeClosestDistance, err := distanceGeographyRegions(
		spheroid,
		useSphereOrSpheroid,
		aRegions,
		bRegions,
		a.BoundingRect().Intersects(b.BoundingRect()),
		distance,
		exclusivity,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(59711)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(59712)
	}
	__antithesis_instrumentation__.Notify(59689)
	if exclusivity == geo.FnExclusive {
		__antithesis_instrumentation__.Notify(59713)
		return maybeClosestDistance < distance, nil
	} else {
		__antithesis_instrumentation__.Notify(59714)
	}
	__antithesis_instrumentation__.Notify(59690)
	return maybeClosestDistance <= distance, nil
}
