package geogfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"math"

	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/geo/geoprojbase"
	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

func BestGeomProjection(boundingRect s2.Rect) (geoprojbase.Proj4Text, error) {
	__antithesis_instrumentation__.Notify(59462)
	center := boundingRect.Center()

	latWidth := s1.Angle(boundingRect.Lat.Length())
	lngWidth := s1.Angle(boundingRect.Lng.Length())

	if center.Lat.Degrees() > 70 && func() bool {
		__antithesis_instrumentation__.Notify(59467)
		return boundingRect.Lo().Lat.Degrees() > 45 == true
	}() == true {
		__antithesis_instrumentation__.Notify(59468)

		return getGeomProjection(3574)
	} else {
		__antithesis_instrumentation__.Notify(59469)
	}
	__antithesis_instrumentation__.Notify(59463)

	if center.Lat.Degrees() < -70 && func() bool {
		__antithesis_instrumentation__.Notify(59470)
		return boundingRect.Hi().Lat.Degrees() < -45 == true
	}() == true {
		__antithesis_instrumentation__.Notify(59471)

		return getGeomProjection(3409)
	} else {
		__antithesis_instrumentation__.Notify(59472)
	}
	__antithesis_instrumentation__.Notify(59464)

	if lngWidth.Degrees() < 6 {
		__antithesis_instrumentation__.Notify(59473)

		sridOffset := geopb.SRID(math.Min(math.Floor((center.Lng.Degrees()+180)/6), 59))
		if center.Lat.Degrees() >= 0 {
			__antithesis_instrumentation__.Notify(59475)

			return getGeomProjection(32601 + sridOffset)
		} else {
			__antithesis_instrumentation__.Notify(59476)
		}
		__antithesis_instrumentation__.Notify(59474)

		return getGeomProjection(32701 + sridOffset)
	} else {
		__antithesis_instrumentation__.Notify(59477)
	}
	__antithesis_instrumentation__.Notify(59465)

	if latWidth.Degrees() < 25 {
		__antithesis_instrumentation__.Notify(59478)

		latZone := math.Min(math.Floor(center.Lat.Degrees()/30), 2)
		latZoneCenterDegrees := (latZone * 30) + 15

		if (latZone == 0 || func() bool {
			__antithesis_instrumentation__.Notify(59481)
			return latZone == -1 == true
		}() == true) && func() bool {
			__antithesis_instrumentation__.Notify(59482)
			return lngWidth.Degrees() <= 30 == true
		}() == true {
			__antithesis_instrumentation__.Notify(59483)
			lngZone := math.Floor(center.Lng.Degrees() / 30)
			return geoprojbase.MakeProj4Text(
				fmt.Sprintf(
					"+proj=laea +ellps=WGS84 +datum=WGS84 +lat_0=%g +lon_0=%g +units=m +no_defs",
					latZoneCenterDegrees,
					(lngZone*30)+15,
				),
			), nil
		} else {
			__antithesis_instrumentation__.Notify(59484)
		}
		__antithesis_instrumentation__.Notify(59479)

		if (latZone == -2 || func() bool {
			__antithesis_instrumentation__.Notify(59485)
			return latZone == 1 == true
		}() == true) && func() bool {
			__antithesis_instrumentation__.Notify(59486)
			return lngWidth.Degrees() <= 45 == true
		}() == true {
			__antithesis_instrumentation__.Notify(59487)
			lngZone := math.Floor(center.Lng.Degrees() / 45)
			return geoprojbase.MakeProj4Text(
				fmt.Sprintf(
					"+proj=laea +ellps=WGS84 +datum=WGS84 +lat_0=%g +lon_0=%g +units=m +no_defs",
					latZoneCenterDegrees,
					(lngZone*45)+22.5,
				),
			), nil
		} else {
			__antithesis_instrumentation__.Notify(59488)
		}
		__antithesis_instrumentation__.Notify(59480)

		if (latZone == -3 || func() bool {
			__antithesis_instrumentation__.Notify(59489)
			return latZone == 2 == true
		}() == true) && func() bool {
			__antithesis_instrumentation__.Notify(59490)
			return lngWidth.Degrees() <= 90 == true
		}() == true {
			__antithesis_instrumentation__.Notify(59491)
			lngZone := math.Floor(center.Lng.Degrees() / 90)
			return geoprojbase.MakeProj4Text(
				fmt.Sprintf(
					"+proj=laea +ellps=WGS84 +datum=WGS84 +lat_0=%g +lon_0=%g +units=m +no_defs",
					latZoneCenterDegrees,
					(lngZone*90)+45,
				),
			), nil
		} else {
			__antithesis_instrumentation__.Notify(59492)
		}
	} else {
		__antithesis_instrumentation__.Notify(59493)
	}
	__antithesis_instrumentation__.Notify(59466)

	return getGeomProjection(3857)
}

func getGeomProjection(srid geopb.SRID) (geoprojbase.Proj4Text, error) {
	__antithesis_instrumentation__.Notify(59494)
	proj, err := geoprojbase.Projection(srid)
	if err != nil {
		__antithesis_instrumentation__.Notify(59496)
		return geoprojbase.Proj4Text{}, err
	} else {
		__antithesis_instrumentation__.Notify(59497)
	}
	__antithesis_instrumentation__.Notify(59495)
	return proj.Proj4Text, nil
}
