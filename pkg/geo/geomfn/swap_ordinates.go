package geomfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strings"

	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/twpayne/go-geom"
)

func SwapOrdinates(geometry geo.Geometry, ords string) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(63281)
	if geometry.Empty() {
		__antithesis_instrumentation__.Notify(63286)
		return geometry, nil
	} else {
		__antithesis_instrumentation__.Notify(63287)
	}
	__antithesis_instrumentation__.Notify(63282)

	t, err := geometry.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(63288)
		return geometry, err
	} else {
		__antithesis_instrumentation__.Notify(63289)
	}
	__antithesis_instrumentation__.Notify(63283)

	newT, err := applyOnCoordsForGeomT(t, func(l geom.Layout, dst, src []float64) error {
		__antithesis_instrumentation__.Notify(63290)
		if len(ords) != 2 {
			__antithesis_instrumentation__.Notify(63293)
			return pgerror.Newf(pgcode.InvalidParameterValue, "invalid ordinate specification. need two letters from the set (x, y, z and m)")
		} else {
			__antithesis_instrumentation__.Notify(63294)
		}
		__antithesis_instrumentation__.Notify(63291)
		ordsIndices, err := getOrdsIndices(l, ords)
		if err != nil {
			__antithesis_instrumentation__.Notify(63295)
			return err
		} else {
			__antithesis_instrumentation__.Notify(63296)
		}
		__antithesis_instrumentation__.Notify(63292)

		dst[ordsIndices[0]], dst[ordsIndices[1]] = src[ordsIndices[1]], src[ordsIndices[0]]
		return nil
	})
	__antithesis_instrumentation__.Notify(63284)
	if err != nil {
		__antithesis_instrumentation__.Notify(63297)
		return geometry, err
	} else {
		__antithesis_instrumentation__.Notify(63298)
	}
	__antithesis_instrumentation__.Notify(63285)

	return geo.MakeGeometryFromGeomT(newT)
}

func getOrdsIndices(l geom.Layout, ords string) ([2]int, error) {
	__antithesis_instrumentation__.Notify(63299)
	ords = strings.ToUpper(ords)
	var ordIndices [2]int
	for i := 0; i < len(ords); i++ {
		__antithesis_instrumentation__.Notify(63301)
		oi := findOrdIndex(ords[i], l)
		if oi == -2 {
			__antithesis_instrumentation__.Notify(63304)
			return ordIndices, pgerror.Newf(pgcode.InvalidParameterValue, "invalid ordinate specification. need two letters from the set (x, y, z and m)")
		} else {
			__antithesis_instrumentation__.Notify(63305)
		}
		__antithesis_instrumentation__.Notify(63302)
		if oi == -1 {
			__antithesis_instrumentation__.Notify(63306)
			return ordIndices, pgerror.Newf(pgcode.InvalidParameterValue, "geometry does not have a %s ordinate", string(ords[i]))
		} else {
			__antithesis_instrumentation__.Notify(63307)
		}
		__antithesis_instrumentation__.Notify(63303)
		ordIndices[i] = oi
	}
	__antithesis_instrumentation__.Notify(63300)

	return ordIndices, nil
}

func findOrdIndex(ordString uint8, l geom.Layout) int {
	__antithesis_instrumentation__.Notify(63308)
	switch ordString {
	case 'X':
		__antithesis_instrumentation__.Notify(63309)
		return 0
	case 'Y':
		__antithesis_instrumentation__.Notify(63310)
		return 1
	case 'Z':
		__antithesis_instrumentation__.Notify(63311)
		return l.ZIndex()
	case 'M':
		__antithesis_instrumentation__.Notify(63312)
		return l.MIndex()
	default:
		__antithesis_instrumentation__.Notify(63313)
		return -2
	}
}
