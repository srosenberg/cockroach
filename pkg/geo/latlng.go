package geo

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "math"

func NormalizeLatitudeDegrees(lat float64) float64 {
	__antithesis_instrumentation__.Notify(64732)

	lat = math.Remainder(lat, 360)

	if lat > 90 {
		__antithesis_instrumentation__.Notify(64735)
		return 180 - lat
	} else {
		__antithesis_instrumentation__.Notify(64736)
	}
	__antithesis_instrumentation__.Notify(64733)

	if lat < -90 {
		__antithesis_instrumentation__.Notify(64737)
		return -180 - lat
	} else {
		__antithesis_instrumentation__.Notify(64738)
	}
	__antithesis_instrumentation__.Notify(64734)
	return lat
}

func NormalizeLongitudeDegrees(lng float64) float64 {
	__antithesis_instrumentation__.Notify(64739)

	return math.Remainder(lng, 360)
}
