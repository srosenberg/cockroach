package geographiclib

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

// #cgo CXXFLAGS: -std=c++14
// #cgo LDFLAGS: -lm
//
// #include "geodesic.h"
// #include "geographiclib.h"
import "C"

import (
	"math"

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

var (
	WGS84Spheroid = NewSpheroid(6378137, 1/298.257223563)
)

type Spheroid struct {
	cRepr        C.struct_geod_geodesic
	Radius       float64
	Flattening   float64
	SphereRadius float64
}

func NewSpheroid(radius float64, flattening float64) *Spheroid {
	__antithesis_instrumentation__.Notify(59915)
	minorAxis := radius - radius*flattening
	s := &Spheroid{
		Radius:       radius,
		Flattening:   flattening,
		SphereRadius: (radius*2 + minorAxis) / 3,
	}
	C.geod_init(&s.cRepr, C.double(radius), C.double(flattening))
	return s
}

func (s *Spheroid) Inverse(a, b s2.LatLng) (s12, az1, az2 float64) {
	__antithesis_instrumentation__.Notify(59916)
	var retS12, retAZ1, retAZ2 C.double
	C.geod_inverse(
		&s.cRepr,
		C.double(a.Lat.Degrees()),
		C.double(a.Lng.Degrees()),
		C.double(b.Lat.Degrees()),
		C.double(b.Lng.Degrees()),
		&retS12,
		&retAZ1,
		&retAZ2,
	)
	return float64(retS12), float64(retAZ1), float64(retAZ2)
}

func (s *Spheroid) InverseBatch(points []s2.Point) float64 {
	__antithesis_instrumentation__.Notify(59917)
	lats := make([]C.double, len(points))
	lngs := make([]C.double, len(points))
	for i, p := range points {
		__antithesis_instrumentation__.Notify(59919)
		latlng := s2.LatLngFromPoint(p)
		lats[i] = C.double(latlng.Lat.Degrees())
		lngs[i] = C.double(latlng.Lng.Degrees())
	}
	__antithesis_instrumentation__.Notify(59918)
	var result C.double
	C.CR_GEOGRAPHICLIB_InverseBatch(
		&s.cRepr,
		&lats[0],
		&lngs[0],
		C.int(len(points)),
		&result,
	)
	return float64(result)
}

func (s *Spheroid) AreaAndPerimeter(points []s2.Point) (area float64, perimeter float64) {
	__antithesis_instrumentation__.Notify(59920)
	lats := make([]C.double, len(points))
	lngs := make([]C.double, len(points))
	for i, p := range points {
		__antithesis_instrumentation__.Notify(59922)
		latlng := s2.LatLngFromPoint(p)
		lats[i] = C.double(latlng.Lat.Degrees())
		lngs[i] = C.double(latlng.Lng.Degrees())
	}
	__antithesis_instrumentation__.Notify(59921)
	var areaDouble, perimeterDouble C.double
	C.geod_polygonarea(
		&s.cRepr,
		&lats[0],
		&lngs[0],
		C.int(len(points)),
		&areaDouble,
		&perimeterDouble,
	)
	return float64(areaDouble), float64(perimeterDouble)
}

func (s *Spheroid) Project(point s2.LatLng, distance float64, azimuth s1.Angle) s2.LatLng {
	__antithesis_instrumentation__.Notify(59923)
	var lat, lng C.double

	C.geod_direct(
		&s.cRepr,
		C.double(point.Lat.Degrees()),
		C.double(point.Lng.Degrees()),
		C.double(azimuth*180.0/math.Pi),
		C.double(distance),
		&lat,
		&lng,
		nil,
	)

	return s2.LatLngFromDegrees(float64(lat), float64(lng))
}
