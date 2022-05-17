package geogfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/geo/geographiclib"
	"github.com/golang/geo/s2"
)

func spheroidDistance(s *geographiclib.Spheroid, a s2.Point, b s2.Point) float64 {
	__antithesis_instrumentation__.Notify(59715)
	inv, _, _ := s.Inverse(s2.LatLngFromPoint(a), s2.LatLngFromPoint(b))
	return inv
}
