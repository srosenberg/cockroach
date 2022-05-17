package geo

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"
	"strings"
)

func decodePolylinePoints(encoded string, precision int) []float64 {
	__antithesis_instrumentation__.Notify(64845)
	idx := 0
	latitude := float64(0)
	longitude := float64(0)
	bytes := []byte(encoded)
	results := []float64{}
	for idx < len(bytes) {
		__antithesis_instrumentation__.Notify(64847)
		var deltaLat float64
		idx, deltaLat = decodePointValue(idx, bytes)
		latitude += deltaLat

		var deltaLng float64
		idx, deltaLng = decodePointValue(idx, bytes)
		longitude += deltaLng
		results = append(results,
			longitude/math.Pow10(precision),
			latitude/math.Pow10(precision))
	}
	__antithesis_instrumentation__.Notify(64846)
	return results
}

func decodePointValue(idx int, bytes []byte) (int, float64) {
	__antithesis_instrumentation__.Notify(64848)
	res := int32(0)
	shift := 0
	for byte := byte(0x20); byte >= 0x20; {
		__antithesis_instrumentation__.Notify(64851)
		if idx > len(bytes)-1 {
			__antithesis_instrumentation__.Notify(64853)
			return idx, 0
		} else {
			__antithesis_instrumentation__.Notify(64854)
		}
		__antithesis_instrumentation__.Notify(64852)
		byte = bytes[idx] - 63
		idx++
		res |= int32(byte&0x1F) << shift
		shift += 5
	}
	__antithesis_instrumentation__.Notify(64849)
	var pointValue float64
	if (res & 1) == 1 {
		__antithesis_instrumentation__.Notify(64855)
		pointValue = float64(^(res >> 1))
	} else {
		__antithesis_instrumentation__.Notify(64856)
		pointValue = float64(res >> 1)
	}
	__antithesis_instrumentation__.Notify(64850)
	return idx, pointValue
}

func encodePolylinePoints(points []float64, precision int) string {
	__antithesis_instrumentation__.Notify(64857)
	lastLat := 0
	lastLng := 0
	var res strings.Builder
	for i := 1; i < len(points); i += 2 {
		__antithesis_instrumentation__.Notify(64859)
		lat := int(math.Round(points[i-1] * math.Pow10(precision)))
		lng := int(math.Round(points[i] * math.Pow10(precision)))
		res = encodePointValue(lng-lastLng, res)
		res = encodePointValue(lat-lastLat, res)
		lastLat = lat
		lastLng = lng
	}
	__antithesis_instrumentation__.Notify(64858)

	return res.String()
}

func encodePointValue(diff int, b strings.Builder) strings.Builder {
	__antithesis_instrumentation__.Notify(64860)
	var shifted int
	shifted = diff << 1
	if diff < 0 {
		__antithesis_instrumentation__.Notify(64863)
		shifted = ^shifted
	} else {
		__antithesis_instrumentation__.Notify(64864)
	}
	__antithesis_instrumentation__.Notify(64861)
	rem := shifted
	for rem >= 0x20 {
		__antithesis_instrumentation__.Notify(64865)
		b.WriteRune(rune(0x20 | (rem & 0x1f) + 63))

		rem = rem >> 5
	}
	__antithesis_instrumentation__.Notify(64862)

	b.WriteRune(rune(rem + 63))
	return b
}
