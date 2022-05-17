package geoproj

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

// #cgo CXXFLAGS: -std=c++14
// #cgo CPPFLAGS: -I../../../c-deps/proj/src
// #cgo !windows LDFLAGS: -lproj
// #cgo linux LDFLAGS: -lrt -lm -lpthread
// #cgo windows LDFLAGS: -lproj_4_9 -lshlwapi -lrpcrt4
//
// #include "proj.h"
import "C"
import (
	"math"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/geo/geographiclib"
	"github.com/cockroachdb/cockroach/pkg/geo/geoprojbase"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
)

const maxArrayLen = 1<<31 - 1

func cStatusToUnsafeGoBytes(s C.CR_PROJ_Status) []byte {
	__antithesis_instrumentation__.Notify(64028)
	if s.data == nil {
		__antithesis_instrumentation__.Notify(64030)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(64031)
	}
	__antithesis_instrumentation__.Notify(64029)

	return (*[maxArrayLen]byte)(unsafe.Pointer(s.data))[:s.len:s.len]
}

func GetProjMetadata(b geoprojbase.Proj4Text) (bool, *geographiclib.Spheroid, error) {
	__antithesis_instrumentation__.Notify(64032)
	var majorAxis, eccentricitySquared C.double
	var isLatLng C.int
	if err := cStatusToUnsafeGoBytes(
		C.CR_PROJ_GetProjMetadata(
			(*C.char)(unsafe.Pointer(&b.Bytes()[0])),
			(*C.int)(unsafe.Pointer(&isLatLng)),
			(*C.double)(unsafe.Pointer(&majorAxis)),
			(*C.double)(unsafe.Pointer(&eccentricitySquared)),
		),
	); err != nil {
		__antithesis_instrumentation__.Notify(64034)
		return false, nil, pgerror.Newf(pgcode.InvalidParameterValue, "error from PROJ: %s", string(err))
	} else {
		__antithesis_instrumentation__.Notify(64035)
	}
	__antithesis_instrumentation__.Notify(64033)

	flattening := float64(eccentricitySquared) / (1 + math.Sqrt(1-float64(eccentricitySquared)))
	return isLatLng != 0, geographiclib.NewSpheroid(float64(majorAxis), flattening), nil
}

func Project(
	from geoprojbase.Proj4Text,
	to geoprojbase.Proj4Text,
	xCoords []float64,
	yCoords []float64,
	zCoords []float64,
) error {
	__antithesis_instrumentation__.Notify(64036)
	if len(xCoords) != len(yCoords) || func() bool {
		__antithesis_instrumentation__.Notify(64040)
		return len(xCoords) != len(zCoords) == true
	}() == true {
		__antithesis_instrumentation__.Notify(64041)
		return pgerror.Newf(
			pgcode.InvalidParameterValue,
			"len(xCoords) != len(yCoords) != len(zCoords): %d != %d != %d",
			len(xCoords),
			len(yCoords),
			len(zCoords),
		)
	} else {
		__antithesis_instrumentation__.Notify(64042)
	}
	__antithesis_instrumentation__.Notify(64037)
	if len(xCoords) == 0 {
		__antithesis_instrumentation__.Notify(64043)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(64044)
	}
	__antithesis_instrumentation__.Notify(64038)
	if err := cStatusToUnsafeGoBytes(C.CR_PROJ_Transform(
		(*C.char)(unsafe.Pointer(&from.Bytes()[0])),
		(*C.char)(unsafe.Pointer(&to.Bytes()[0])),
		C.long(len(xCoords)),
		(*C.double)(unsafe.Pointer(&xCoords[0])),
		(*C.double)(unsafe.Pointer(&yCoords[0])),
		(*C.double)(unsafe.Pointer(&zCoords[0])),
	)); err != nil {
		__antithesis_instrumentation__.Notify(64045)
		return pgerror.Newf(pgcode.InvalidParameterValue, "error from PROJ: %s", string(err))
	} else {
		__antithesis_instrumentation__.Notify(64046)
	}
	__antithesis_instrumentation__.Notify(64039)
	return nil
}
