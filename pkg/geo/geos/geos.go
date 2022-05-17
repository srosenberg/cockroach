package geos

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sync"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/build/bazel"
	"github.com/cockroachdb/cockroach/pkg/docs"
	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/errors"
)

// #cgo CXXFLAGS: -std=c++14
// #cgo !windows LDFLAGS: -ldl -lm
//
// #include "geos.h"
import "C"

type EnsureInitErrorDisplay int

const (
	EnsureInitErrorDisplayPrivate EnsureInitErrorDisplay = iota

	EnsureInitErrorDisplayPublic
)

const maxArrayLen = 1<<31 - 1

var geosOnce struct {
	geos *C.CR_GEOS
	loc  string
	err  error
	once sync.Once
}

type PreparedGeometry *C.CR_GEOS_PreparedGeometry

func EnsureInit(
	errDisplay EnsureInitErrorDisplay, flagLibraryDirectoryValue string,
) (string, error) {
	__antithesis_instrumentation__.Notify(64093)
	crdbBinaryLoc := ""
	if len(os.Args) > 0 {
		__antithesis_instrumentation__.Notify(64095)
		crdbBinaryLoc = os.Args[0]
	} else {
		__antithesis_instrumentation__.Notify(64096)
	}
	__antithesis_instrumentation__.Notify(64094)
	_, err := ensureInit(errDisplay, flagLibraryDirectoryValue, crdbBinaryLoc)
	return geosOnce.loc, err
}

func ensureInitInternal() (*C.CR_GEOS, error) {
	__antithesis_instrumentation__.Notify(64097)
	return ensureInit(EnsureInitErrorDisplayPrivate, "", "")
}

func ensureInit(
	errDisplay EnsureInitErrorDisplay, flagLibraryDirectoryValue string, crdbBinaryLoc string,
) (*C.CR_GEOS, error) {
	__antithesis_instrumentation__.Notify(64098)
	geosOnce.once.Do(func() {
		__antithesis_instrumentation__.Notify(64101)
		geosOnce.geos, geosOnce.loc, geosOnce.err = initGEOS(
			findLibraryDirectories(flagLibraryDirectoryValue, crdbBinaryLoc),
		)
	})
	__antithesis_instrumentation__.Notify(64099)
	if geosOnce.err != nil && func() bool {
		__antithesis_instrumentation__.Notify(64102)
		return errDisplay == EnsureInitErrorDisplayPublic == true
	}() == true {
		__antithesis_instrumentation__.Notify(64103)
		return nil, pgerror.Newf(pgcode.System, "geos: this operation is not available")
	} else {
		__antithesis_instrumentation__.Notify(64104)
	}
	__antithesis_instrumentation__.Notify(64100)
	return geosOnce.geos, geosOnce.err
}

func getLibraryExt(base string) string {
	__antithesis_instrumentation__.Notify(64105)
	switch runtime.GOOS {
	case "darwin":
		__antithesis_instrumentation__.Notify(64106)
		return base + ".dylib"
	case "windows":
		__antithesis_instrumentation__.Notify(64107)
		return base + ".dll"
	default:
		__antithesis_instrumentation__.Notify(64108)
		return base + ".so"
	}
}

const (
	libgeosFileName  = "libgeos"
	libgeoscFileName = "libgeos_c"
)

func findLibraryDirectories(flagLibraryDirectoryValue string, crdbBinaryLoc string) []string {
	__antithesis_instrumentation__.Notify(64109)

	cwd, err := os.Getwd()
	if err != nil {
		__antithesis_instrumentation__.Notify(64113)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(64114)
	}
	__antithesis_instrumentation__.Notify(64110)
	locs := []string{}
	if flagLibraryDirectoryValue != "" {
		__antithesis_instrumentation__.Notify(64115)
		locs = append(locs, flagLibraryDirectoryValue)
	} else {
		__antithesis_instrumentation__.Notify(64116)
	}
	__antithesis_instrumentation__.Notify(64111)

	if bazel.BuiltWithBazel() {
		__antithesis_instrumentation__.Notify(64117)
		if p, err := bazel.Runfile(path.Join("c-deps", "libgeos", "lib")); err == nil {
			__antithesis_instrumentation__.Notify(64118)
			locs = append(locs, p)
		} else {
			__antithesis_instrumentation__.Notify(64119)
		}
	} else {
		__antithesis_instrumentation__.Notify(64120)
	}
	__antithesis_instrumentation__.Notify(64112)
	locs = append(
		append(
			locs,
			findLibraryDirectoriesInParentingDirectories(crdbBinaryLoc)...,
		),
		findLibraryDirectoriesInParentingDirectories(cwd)...,
	)
	return locs
}

func findLibraryDirectoriesInParentingDirectories(dir string) []string {
	__antithesis_instrumentation__.Notify(64121)
	locs := []string{}

	for {
		__antithesis_instrumentation__.Notify(64123)
		checkDir := filepath.Join(dir, "lib")
		found := true
		for _, file := range []string{
			filepath.Join(checkDir, getLibraryExt(libgeoscFileName)),
			filepath.Join(checkDir, getLibraryExt(libgeosFileName)),
		} {
			__antithesis_instrumentation__.Notify(64127)
			if _, err := os.Stat(file); err != nil {
				__antithesis_instrumentation__.Notify(64128)
				found = false
				break
			} else {
				__antithesis_instrumentation__.Notify(64129)
			}
		}
		__antithesis_instrumentation__.Notify(64124)
		if found {
			__antithesis_instrumentation__.Notify(64130)
			locs = append(locs, checkDir)
		} else {
			__antithesis_instrumentation__.Notify(64131)
		}
		__antithesis_instrumentation__.Notify(64125)
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			__antithesis_instrumentation__.Notify(64132)
			break
		} else {
			__antithesis_instrumentation__.Notify(64133)
		}
		__antithesis_instrumentation__.Notify(64126)
		dir = parentDir
	}
	__antithesis_instrumentation__.Notify(64122)
	return locs
}

func initGEOS(dirs []string) (*C.CR_GEOS, string, error) {
	__antithesis_instrumentation__.Notify(64134)
	var err error
	for _, dir := range dirs {
		__antithesis_instrumentation__.Notify(64137)
		var ret *C.CR_GEOS
		newErr := statusToError(
			C.CR_GEOS_Init(
				goToCSlice([]byte(filepath.Join(dir, getLibraryExt(libgeoscFileName)))),
				goToCSlice([]byte(filepath.Join(dir, getLibraryExt(libgeosFileName)))),
				&ret,
			),
		)
		if newErr == nil {
			__antithesis_instrumentation__.Notify(64139)
			return ret, dir, nil
		} else {
			__antithesis_instrumentation__.Notify(64140)
		}
		__antithesis_instrumentation__.Notify(64138)
		err = errors.CombineErrors(
			err,
			errors.Wrapf(
				newErr,
				"geos: cannot load GEOS from dir %q",
				dir,
			),
		)
	}
	__antithesis_instrumentation__.Notify(64135)
	if err != nil {
		__antithesis_instrumentation__.Notify(64141)
		return nil, "", wrapGEOSInitError(errors.Wrap(err, "geos: error during GEOS init"))
	} else {
		__antithesis_instrumentation__.Notify(64142)
	}
	__antithesis_instrumentation__.Notify(64136)
	return nil, "", wrapGEOSInitError(errors.Newf("geos: no locations to init GEOS"))
}

func wrapGEOSInitError(err error) error {
	__antithesis_instrumentation__.Notify(64143)
	page := "linux"
	switch runtime.GOOS {
	case "darwin":
		__antithesis_instrumentation__.Notify(64145)
		page = "mac"
	case "windows":
		__antithesis_instrumentation__.Notify(64146)
		page = "windows"
	default:
		__antithesis_instrumentation__.Notify(64147)
	}
	__antithesis_instrumentation__.Notify(64144)
	return errors.WithHintf(
		err,
		"Ensure you have the spatial libraries installed as per the instructions in %s",
		docs.URL("install-cockroachdb-"+page),
	)
}

func goToCSlice(b []byte) C.CR_GEOS_Slice {
	__antithesis_instrumentation__.Notify(64148)
	if len(b) == 0 {
		__antithesis_instrumentation__.Notify(64150)
		return C.CR_GEOS_Slice{data: nil, len: 0}
	} else {
		__antithesis_instrumentation__.Notify(64151)
	}
	__antithesis_instrumentation__.Notify(64149)
	return C.CR_GEOS_Slice{
		data: (*C.char)(unsafe.Pointer(&b[0])),
		len:  C.size_t(len(b)),
	}
}

func cStringToUnsafeGoBytes(s C.CR_GEOS_String) []byte {
	__antithesis_instrumentation__.Notify(64152)
	return cToUnsafeGoBytes(s.data, s.len)
}

func cToUnsafeGoBytes(data *C.char, len C.size_t) []byte {
	__antithesis_instrumentation__.Notify(64153)
	if data == nil {
		__antithesis_instrumentation__.Notify(64155)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(64156)
	}
	__antithesis_instrumentation__.Notify(64154)

	return (*[maxArrayLen]byte)(unsafe.Pointer(data))[:len:len]
}

func cStringToSafeGoBytes(s C.CR_GEOS_String) []byte {
	__antithesis_instrumentation__.Notify(64157)
	unsafeBytes := cStringToUnsafeGoBytes(s)
	b := make([]byte, len(unsafeBytes))
	copy(b, unsafeBytes)
	C.free(unsafe.Pointer(s.data))
	return b
}

type Error struct {
	msg string
}

func (err *Error) Error() string {
	__antithesis_instrumentation__.Notify(64158)
	return err.msg
}

func statusToError(s C.CR_GEOS_Status) error {
	__antithesis_instrumentation__.Notify(64159)
	if s.data == nil {
		__antithesis_instrumentation__.Notify(64161)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(64162)
	}
	__antithesis_instrumentation__.Notify(64160)
	return &Error{msg: string(cStringToSafeGoBytes(s))}
}

type BufferParamsJoinStyle int

const (
	BufferParamsJoinStyleRound = 1
	BufferParamsJoinStyleMitre = 2
	BufferParamsJoinStyleBevel = 3
)

type BufferParamsEndCapStyle int

const (
	BufferParamsEndCapStyleRound  = 1
	BufferParamsEndCapStyleFlat   = 2
	BufferParamsEndCapStyleSquare = 3
)

type BufferParams struct {
	JoinStyle        BufferParamsJoinStyle
	EndCapStyle      BufferParamsEndCapStyle
	SingleSided      bool
	QuadrantSegments int
	MitreLimit       float64
}

func Buffer(ewkb geopb.EWKB, params BufferParams, distance float64) (geopb.EWKB, error) {
	__antithesis_instrumentation__.Notify(64163)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64167)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64168)
	}
	__antithesis_instrumentation__.Notify(64164)
	singleSided := 0
	if params.SingleSided {
		__antithesis_instrumentation__.Notify(64169)
		singleSided = 1
	} else {
		__antithesis_instrumentation__.Notify(64170)
	}
	__antithesis_instrumentation__.Notify(64165)
	cParams := C.CR_GEOS_BufferParamsInput{
		endCapStyle:      C.int(params.EndCapStyle),
		joinStyle:        C.int(params.JoinStyle),
		singleSided:      C.int(singleSided),
		quadrantSegments: C.int(params.QuadrantSegments),
		mitreLimit:       C.double(params.MitreLimit),
	}
	var cEWKB C.CR_GEOS_String
	if err := statusToError(C.CR_GEOS_Buffer(g, goToCSlice(ewkb), cParams, C.double(distance), &cEWKB)); err != nil {
		__antithesis_instrumentation__.Notify(64171)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64172)
	}
	__antithesis_instrumentation__.Notify(64166)
	return cStringToSafeGoBytes(cEWKB), nil
}

func Area(ewkb geopb.EWKB) (float64, error) {
	__antithesis_instrumentation__.Notify(64173)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64176)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(64177)
	}
	__antithesis_instrumentation__.Notify(64174)
	var area C.double
	if err := statusToError(C.CR_GEOS_Area(g, goToCSlice(ewkb), &area)); err != nil {
		__antithesis_instrumentation__.Notify(64178)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(64179)
	}
	__antithesis_instrumentation__.Notify(64175)
	return float64(area), nil
}

func Boundary(ewkb geopb.EWKB) (geopb.EWKB, error) {
	__antithesis_instrumentation__.Notify(64180)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64183)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64184)
	}
	__antithesis_instrumentation__.Notify(64181)
	var cEWKB C.CR_GEOS_String
	if err := statusToError(C.CR_GEOS_Boundary(g, goToCSlice(ewkb), &cEWKB)); err != nil {
		__antithesis_instrumentation__.Notify(64185)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64186)
	}
	__antithesis_instrumentation__.Notify(64182)
	return cStringToSafeGoBytes(cEWKB), nil
}

func Difference(ewkb1 geopb.EWKB, ewkb2 geopb.EWKB) (geopb.EWKB, error) {
	__antithesis_instrumentation__.Notify(64187)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64190)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64191)
	}
	__antithesis_instrumentation__.Notify(64188)
	var diffEWKB C.CR_GEOS_String
	if err := statusToError(C.CR_GEOS_Difference(g, goToCSlice(ewkb1), goToCSlice(ewkb2), &diffEWKB)); err != nil {
		__antithesis_instrumentation__.Notify(64192)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64193)
	}
	__antithesis_instrumentation__.Notify(64189)
	return cStringToSafeGoBytes(diffEWKB), nil
}

func Length(ewkb geopb.EWKB) (float64, error) {
	__antithesis_instrumentation__.Notify(64194)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64197)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(64198)
	}
	__antithesis_instrumentation__.Notify(64195)
	var length C.double
	if err := statusToError(C.CR_GEOS_Length(g, goToCSlice(ewkb), &length)); err != nil {
		__antithesis_instrumentation__.Notify(64199)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(64200)
	}
	__antithesis_instrumentation__.Notify(64196)
	return float64(length), nil
}

func Normalize(a geopb.EWKB) (geopb.EWKB, error) {
	__antithesis_instrumentation__.Notify(64201)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64204)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64205)
	}
	__antithesis_instrumentation__.Notify(64202)
	var cEWKB C.CR_GEOS_String
	if err := statusToError(C.CR_GEOS_Normalize(g, goToCSlice(a), &cEWKB)); err != nil {
		__antithesis_instrumentation__.Notify(64206)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64207)
	}
	__antithesis_instrumentation__.Notify(64203)
	return cStringToSafeGoBytes(cEWKB), nil
}

func LineMerge(a geopb.EWKB) (geopb.EWKB, error) {
	__antithesis_instrumentation__.Notify(64208)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64211)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64212)
	}
	__antithesis_instrumentation__.Notify(64209)
	var cEWKB C.CR_GEOS_String
	if err := statusToError(C.CR_GEOS_LineMerge(g, goToCSlice(a), &cEWKB)); err != nil {
		__antithesis_instrumentation__.Notify(64213)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64214)
	}
	__antithesis_instrumentation__.Notify(64210)
	return cStringToSafeGoBytes(cEWKB), nil
}

func IsSimple(ewkb geopb.EWKB) (bool, error) {
	__antithesis_instrumentation__.Notify(64215)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64218)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(64219)
	}
	__antithesis_instrumentation__.Notify(64216)
	var ret C.char
	if err := statusToError(C.CR_GEOS_IsSimple(g, goToCSlice(ewkb), &ret)); err != nil {
		__antithesis_instrumentation__.Notify(64220)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(64221)
	}
	__antithesis_instrumentation__.Notify(64217)
	return ret == 1, nil
}

func Centroid(ewkb geopb.EWKB) (geopb.EWKB, error) {
	__antithesis_instrumentation__.Notify(64222)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64225)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64226)
	}
	__antithesis_instrumentation__.Notify(64223)
	var cEWKB C.CR_GEOS_String
	if err := statusToError(C.CR_GEOS_Centroid(g, goToCSlice(ewkb), &cEWKB)); err != nil {
		__antithesis_instrumentation__.Notify(64227)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64228)
	}
	__antithesis_instrumentation__.Notify(64224)
	return cStringToSafeGoBytes(cEWKB), nil
}

func MinimumBoundingCircle(ewkb geopb.EWKB) (geopb.EWKB, geopb.EWKB, float64, error) {
	__antithesis_instrumentation__.Notify(64229)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64232)
		return nil, nil, 0, err
	} else {
		__antithesis_instrumentation__.Notify(64233)
	}
	__antithesis_instrumentation__.Notify(64230)
	var centerEWKB C.CR_GEOS_String
	var polygonEWKB C.CR_GEOS_String
	var radius C.double

	if err := statusToError(C.CR_GEOS_MinimumBoundingCircle(g, goToCSlice(ewkb), &radius, &centerEWKB, &polygonEWKB)); err != nil {
		__antithesis_instrumentation__.Notify(64234)
		return nil, nil, 0, err
	} else {
		__antithesis_instrumentation__.Notify(64235)
	}
	__antithesis_instrumentation__.Notify(64231)
	return cStringToSafeGoBytes(polygonEWKB), cStringToSafeGoBytes(centerEWKB), float64(radius), nil

}

func ConvexHull(ewkb geopb.EWKB) (geopb.EWKB, error) {
	__antithesis_instrumentation__.Notify(64236)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64239)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64240)
	}
	__antithesis_instrumentation__.Notify(64237)
	var cEWKB C.CR_GEOS_String
	if err := statusToError(C.CR_GEOS_ConvexHull(g, goToCSlice(ewkb), &cEWKB)); err != nil {
		__antithesis_instrumentation__.Notify(64241)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64242)
	}
	__antithesis_instrumentation__.Notify(64238)
	return cStringToSafeGoBytes(cEWKB), nil
}

func Simplify(ewkb geopb.EWKB, tolerance float64) (geopb.EWKB, error) {
	__antithesis_instrumentation__.Notify(64243)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64246)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64247)
	}
	__antithesis_instrumentation__.Notify(64244)
	var cEWKB C.CR_GEOS_String
	if err := statusToError(
		C.CR_GEOS_Simplify(g, goToCSlice(ewkb), &cEWKB, C.double(tolerance)),
	); err != nil {
		__antithesis_instrumentation__.Notify(64248)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64249)
	}
	__antithesis_instrumentation__.Notify(64245)
	return cStringToSafeGoBytes(cEWKB), nil
}

func TopologyPreserveSimplify(ewkb geopb.EWKB, tolerance float64) (geopb.EWKB, error) {
	__antithesis_instrumentation__.Notify(64250)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64253)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64254)
	}
	__antithesis_instrumentation__.Notify(64251)
	var cEWKB C.CR_GEOS_String
	if err := statusToError(
		C.CR_GEOS_TopologyPreserveSimplify(g, goToCSlice(ewkb), &cEWKB, C.double(tolerance)),
	); err != nil {
		__antithesis_instrumentation__.Notify(64255)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64256)
	}
	__antithesis_instrumentation__.Notify(64252)
	return cStringToSafeGoBytes(cEWKB), nil
}

func PointOnSurface(ewkb geopb.EWKB) (geopb.EWKB, error) {
	__antithesis_instrumentation__.Notify(64257)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64260)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64261)
	}
	__antithesis_instrumentation__.Notify(64258)
	var cEWKB C.CR_GEOS_String
	if err := statusToError(C.CR_GEOS_PointOnSurface(g, goToCSlice(ewkb), &cEWKB)); err != nil {
		__antithesis_instrumentation__.Notify(64262)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64263)
	}
	__antithesis_instrumentation__.Notify(64259)
	return cStringToSafeGoBytes(cEWKB), nil
}

func Intersection(a geopb.EWKB, b geopb.EWKB) (geopb.EWKB, error) {
	__antithesis_instrumentation__.Notify(64264)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64267)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64268)
	}
	__antithesis_instrumentation__.Notify(64265)
	var cEWKB C.CR_GEOS_String
	if err := statusToError(C.CR_GEOS_Intersection(g, goToCSlice(a), goToCSlice(b), &cEWKB)); err != nil {
		__antithesis_instrumentation__.Notify(64269)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64270)
	}
	__antithesis_instrumentation__.Notify(64266)
	return cStringToSafeGoBytes(cEWKB), nil
}

func UnaryUnion(a geopb.EWKB) (geopb.EWKB, error) {
	__antithesis_instrumentation__.Notify(64271)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64274)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64275)
	}
	__antithesis_instrumentation__.Notify(64272)
	var unionEWKB C.CR_GEOS_String
	if err := statusToError(C.CR_GEOS_UnaryUnion(g, goToCSlice(a), &unionEWKB)); err != nil {
		__antithesis_instrumentation__.Notify(64276)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64277)
	}
	__antithesis_instrumentation__.Notify(64273)
	return cStringToSafeGoBytes(unionEWKB), nil
}

func Union(a geopb.EWKB, b geopb.EWKB) (geopb.EWKB, error) {
	__antithesis_instrumentation__.Notify(64278)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64281)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64282)
	}
	__antithesis_instrumentation__.Notify(64279)
	var cEWKB C.CR_GEOS_String
	if err := statusToError(C.CR_GEOS_Union(g, goToCSlice(a), goToCSlice(b), &cEWKB)); err != nil {
		__antithesis_instrumentation__.Notify(64283)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64284)
	}
	__antithesis_instrumentation__.Notify(64280)
	return cStringToSafeGoBytes(cEWKB), nil
}

func SymDifference(a geopb.EWKB, b geopb.EWKB) (geopb.EWKB, error) {
	__antithesis_instrumentation__.Notify(64285)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64288)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64289)
	}
	__antithesis_instrumentation__.Notify(64286)
	var cEWKB C.CR_GEOS_String
	if err := statusToError(C.CR_GEOS_SymDifference(g, goToCSlice(a), goToCSlice(b), &cEWKB)); err != nil {
		__antithesis_instrumentation__.Notify(64290)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64291)
	}
	__antithesis_instrumentation__.Notify(64287)
	return cStringToSafeGoBytes(cEWKB), nil
}

func InterpolateLine(ewkb geopb.EWKB, distance float64) (geopb.EWKB, error) {
	__antithesis_instrumentation__.Notify(64292)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64295)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64296)
	}
	__antithesis_instrumentation__.Notify(64293)
	var cEWKB C.CR_GEOS_String
	if err := statusToError(C.CR_GEOS_Interpolate(g, goToCSlice(ewkb), C.double(distance), &cEWKB)); err != nil {
		__antithesis_instrumentation__.Notify(64297)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64298)
	}
	__antithesis_instrumentation__.Notify(64294)
	return cStringToSafeGoBytes(cEWKB), nil
}

func MinDistance(a geopb.EWKB, b geopb.EWKB) (float64, error) {
	__antithesis_instrumentation__.Notify(64299)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64302)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(64303)
	}
	__antithesis_instrumentation__.Notify(64300)
	var distance C.double
	if err := statusToError(C.CR_GEOS_Distance(g, goToCSlice(a), goToCSlice(b), &distance)); err != nil {
		__antithesis_instrumentation__.Notify(64304)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(64305)
	}
	__antithesis_instrumentation__.Notify(64301)
	return float64(distance), nil
}

func MinimumClearance(ewkb geopb.EWKB) (float64, error) {
	__antithesis_instrumentation__.Notify(64306)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64309)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(64310)
	}
	__antithesis_instrumentation__.Notify(64307)
	var distance C.double
	if err := statusToError(C.CR_GEOS_MinimumClearance(g, goToCSlice(ewkb), &distance)); err != nil {
		__antithesis_instrumentation__.Notify(64311)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(64312)
	}
	__antithesis_instrumentation__.Notify(64308)
	return float64(distance), nil
}

func MinimumClearanceLine(ewkb geopb.EWKB) (geopb.EWKB, error) {
	__antithesis_instrumentation__.Notify(64313)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64316)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64317)
	}
	__antithesis_instrumentation__.Notify(64314)
	var clearanceEWKB C.CR_GEOS_String
	if err := statusToError(C.CR_GEOS_MinimumClearanceLine(g, goToCSlice(ewkb), &clearanceEWKB)); err != nil {
		__antithesis_instrumentation__.Notify(64318)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64319)
	}
	__antithesis_instrumentation__.Notify(64315)
	return cStringToSafeGoBytes(clearanceEWKB), nil
}

func ClipByRect(
	ewkb geopb.EWKB, xMin float64, yMin float64, xMax float64, yMax float64,
) (geopb.EWKB, error) {
	__antithesis_instrumentation__.Notify(64320)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64323)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64324)
	}
	__antithesis_instrumentation__.Notify(64321)
	var cEWKB C.CR_GEOS_String
	if err := statusToError(C.CR_GEOS_ClipByRect(g, goToCSlice(ewkb), C.double(xMin),
		C.double(yMin), C.double(xMax), C.double(yMax), &cEWKB)); err != nil {
		__antithesis_instrumentation__.Notify(64325)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64326)
	}
	__antithesis_instrumentation__.Notify(64322)
	return cStringToSafeGoBytes(cEWKB), nil
}

func PrepareGeometry(a geopb.EWKB) (PreparedGeometry, error) {
	__antithesis_instrumentation__.Notify(64327)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64330)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64331)
	}
	__antithesis_instrumentation__.Notify(64328)
	var ret *C.CR_GEOS_PreparedGeometry
	if err := statusToError(C.CR_GEOS_Prepare(g, goToCSlice(a), &ret)); err != nil {
		__antithesis_instrumentation__.Notify(64332)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64333)
	}
	__antithesis_instrumentation__.Notify(64329)
	return PreparedGeometry(ret), nil
}

func PreparedGeomDestroy(a PreparedGeometry) {
	__antithesis_instrumentation__.Notify(64334)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64336)
		panic(errors.NewAssertionErrorWithWrappedErrf(err, "trying to destroy PreparedGeometry with no GEOS"))
	} else {
		__antithesis_instrumentation__.Notify(64337)
	}
	__antithesis_instrumentation__.Notify(64335)
	ap := (*C.CR_GEOS_PreparedGeometry)(unsafe.Pointer(a))
	if err := statusToError(C.CR_GEOS_PreparedGeometryDestroy(g, ap)); err != nil {
		__antithesis_instrumentation__.Notify(64338)
		panic(errors.NewAssertionErrorWithWrappedErrf(err, "PreparedGeometryDestroy returned an error"))
	} else {
		__antithesis_instrumentation__.Notify(64339)
	}
}

func Covers(a geopb.EWKB, b geopb.EWKB) (bool, error) {
	__antithesis_instrumentation__.Notify(64340)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64343)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(64344)
	}
	__antithesis_instrumentation__.Notify(64341)
	var ret C.char
	if err := statusToError(C.CR_GEOS_Covers(g, goToCSlice(a), goToCSlice(b), &ret)); err != nil {
		__antithesis_instrumentation__.Notify(64345)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(64346)
	}
	__antithesis_instrumentation__.Notify(64342)
	return ret == 1, nil
}

func CoveredBy(a geopb.EWKB, b geopb.EWKB) (bool, error) {
	__antithesis_instrumentation__.Notify(64347)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64350)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(64351)
	}
	__antithesis_instrumentation__.Notify(64348)
	var ret C.char
	if err := statusToError(C.CR_GEOS_CoveredBy(g, goToCSlice(a), goToCSlice(b), &ret)); err != nil {
		__antithesis_instrumentation__.Notify(64352)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(64353)
	}
	__antithesis_instrumentation__.Notify(64349)
	return ret == 1, nil
}

func Contains(a geopb.EWKB, b geopb.EWKB) (bool, error) {
	__antithesis_instrumentation__.Notify(64354)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64357)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(64358)
	}
	__antithesis_instrumentation__.Notify(64355)
	var ret C.char
	if err := statusToError(C.CR_GEOS_Contains(g, goToCSlice(a), goToCSlice(b), &ret)); err != nil {
		__antithesis_instrumentation__.Notify(64359)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(64360)
	}
	__antithesis_instrumentation__.Notify(64356)
	return ret == 1, nil
}

func Crosses(a geopb.EWKB, b geopb.EWKB) (bool, error) {
	__antithesis_instrumentation__.Notify(64361)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64364)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(64365)
	}
	__antithesis_instrumentation__.Notify(64362)
	var ret C.char
	if err := statusToError(C.CR_GEOS_Crosses(g, goToCSlice(a), goToCSlice(b), &ret)); err != nil {
		__antithesis_instrumentation__.Notify(64366)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(64367)
	}
	__antithesis_instrumentation__.Notify(64363)
	return ret == 1, nil
}

func Disjoint(a geopb.EWKB, b geopb.EWKB) (bool, error) {
	__antithesis_instrumentation__.Notify(64368)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64371)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(64372)
	}
	__antithesis_instrumentation__.Notify(64369)
	var ret C.char
	if err := statusToError(C.CR_GEOS_Disjoint(g, goToCSlice(a), goToCSlice(b), &ret)); err != nil {
		__antithesis_instrumentation__.Notify(64373)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(64374)
	}
	__antithesis_instrumentation__.Notify(64370)
	return ret == 1, nil
}

func Equals(a geopb.EWKB, b geopb.EWKB) (bool, error) {
	__antithesis_instrumentation__.Notify(64375)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64378)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(64379)
	}
	__antithesis_instrumentation__.Notify(64376)
	var ret C.char
	if err := statusToError(C.CR_GEOS_Equals(g, goToCSlice(a), goToCSlice(b), &ret)); err != nil {
		__antithesis_instrumentation__.Notify(64380)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(64381)
	}
	__antithesis_instrumentation__.Notify(64377)
	return ret == 1, nil
}

func PreparedIntersects(a PreparedGeometry, b geopb.EWKB) (bool, error) {
	__antithesis_instrumentation__.Notify(64382)

	if a == nil {
		__antithesis_instrumentation__.Notify(64386)
		return false, errors.New("provided PreparedGeometry is nil")
	} else {
		__antithesis_instrumentation__.Notify(64387)
	}
	__antithesis_instrumentation__.Notify(64383)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64388)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(64389)
	}
	__antithesis_instrumentation__.Notify(64384)
	var ret C.char
	ap := (*C.CR_GEOS_PreparedGeometry)(unsafe.Pointer(a))
	if err := statusToError(C.CR_GEOS_PreparedIntersects(g, ap, goToCSlice(b), &ret)); err != nil {
		__antithesis_instrumentation__.Notify(64390)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(64391)
	}
	__antithesis_instrumentation__.Notify(64385)
	return ret == 1, nil
}

func Intersects(a geopb.EWKB, b geopb.EWKB) (bool, error) {
	__antithesis_instrumentation__.Notify(64392)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64395)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(64396)
	}
	__antithesis_instrumentation__.Notify(64393)
	var ret C.char
	if err := statusToError(C.CR_GEOS_Intersects(g, goToCSlice(a), goToCSlice(b), &ret)); err != nil {
		__antithesis_instrumentation__.Notify(64397)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(64398)
	}
	__antithesis_instrumentation__.Notify(64394)
	return ret == 1, nil
}

func Overlaps(a geopb.EWKB, b geopb.EWKB) (bool, error) {
	__antithesis_instrumentation__.Notify(64399)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64402)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(64403)
	}
	__antithesis_instrumentation__.Notify(64400)
	var ret C.char
	if err := statusToError(C.CR_GEOS_Overlaps(g, goToCSlice(a), goToCSlice(b), &ret)); err != nil {
		__antithesis_instrumentation__.Notify(64404)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(64405)
	}
	__antithesis_instrumentation__.Notify(64401)
	return ret == 1, nil
}

func Touches(a geopb.EWKB, b geopb.EWKB) (bool, error) {
	__antithesis_instrumentation__.Notify(64406)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64409)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(64410)
	}
	__antithesis_instrumentation__.Notify(64407)
	var ret C.char
	if err := statusToError(C.CR_GEOS_Touches(g, goToCSlice(a), goToCSlice(b), &ret)); err != nil {
		__antithesis_instrumentation__.Notify(64411)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(64412)
	}
	__antithesis_instrumentation__.Notify(64408)
	return ret == 1, nil
}

func Within(a geopb.EWKB, b geopb.EWKB) (bool, error) {
	__antithesis_instrumentation__.Notify(64413)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64416)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(64417)
	}
	__antithesis_instrumentation__.Notify(64414)
	var ret C.char
	if err := statusToError(C.CR_GEOS_Within(g, goToCSlice(a), goToCSlice(b), &ret)); err != nil {
		__antithesis_instrumentation__.Notify(64418)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(64419)
	}
	__antithesis_instrumentation__.Notify(64415)
	return ret == 1, nil
}

func FrechetDistance(a, b geopb.EWKB) (float64, error) {
	__antithesis_instrumentation__.Notify(64420)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64423)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(64424)
	}
	__antithesis_instrumentation__.Notify(64421)
	var distance C.double
	if err := statusToError(
		C.CR_GEOS_FrechetDistance(g, goToCSlice(a), goToCSlice(b), &distance),
	); err != nil {
		__antithesis_instrumentation__.Notify(64425)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(64426)
	}
	__antithesis_instrumentation__.Notify(64422)
	return float64(distance), nil
}

func FrechetDistanceDensify(a, b geopb.EWKB, densifyFrac float64) (float64, error) {
	__antithesis_instrumentation__.Notify(64427)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64430)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(64431)
	}
	__antithesis_instrumentation__.Notify(64428)
	var distance C.double
	if err := statusToError(
		C.CR_GEOS_FrechetDistanceDensify(g, goToCSlice(a), goToCSlice(b), C.double(densifyFrac), &distance),
	); err != nil {
		__antithesis_instrumentation__.Notify(64432)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(64433)
	}
	__antithesis_instrumentation__.Notify(64429)
	return float64(distance), nil
}

func HausdorffDistance(a, b geopb.EWKB) (float64, error) {
	__antithesis_instrumentation__.Notify(64434)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64437)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(64438)
	}
	__antithesis_instrumentation__.Notify(64435)
	var distance C.double
	if err := statusToError(
		C.CR_GEOS_HausdorffDistance(g, goToCSlice(a), goToCSlice(b), &distance),
	); err != nil {
		__antithesis_instrumentation__.Notify(64439)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(64440)
	}
	__antithesis_instrumentation__.Notify(64436)
	return float64(distance), nil
}

func HausdorffDistanceDensify(a, b geopb.EWKB, densifyFrac float64) (float64, error) {
	__antithesis_instrumentation__.Notify(64441)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64444)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(64445)
	}
	__antithesis_instrumentation__.Notify(64442)
	var distance C.double
	if err := statusToError(
		C.CR_GEOS_HausdorffDistanceDensify(g, goToCSlice(a), goToCSlice(b), C.double(densifyFrac), &distance),
	); err != nil {
		__antithesis_instrumentation__.Notify(64446)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(64447)
	}
	__antithesis_instrumentation__.Notify(64443)
	return float64(distance), nil
}

func EqualsExact(lhs, rhs geopb.EWKB, epsilon float64) (bool, error) {
	__antithesis_instrumentation__.Notify(64448)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64451)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(64452)
	}
	__antithesis_instrumentation__.Notify(64449)
	var ret C.char
	if err := statusToError(
		C.CR_GEOS_EqualsExact(g, goToCSlice(lhs), goToCSlice(rhs), C.double(epsilon), &ret),
	); err != nil {
		__antithesis_instrumentation__.Notify(64453)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(64454)
	}
	__antithesis_instrumentation__.Notify(64450)
	return ret == 1, nil
}

func Relate(a geopb.EWKB, b geopb.EWKB) (string, error) {
	__antithesis_instrumentation__.Notify(64455)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64459)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(64460)
	}
	__antithesis_instrumentation__.Notify(64456)
	var ret C.CR_GEOS_String
	if err := statusToError(C.CR_GEOS_Relate(g, goToCSlice(a), goToCSlice(b), &ret)); err != nil {
		__antithesis_instrumentation__.Notify(64461)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(64462)
	}
	__antithesis_instrumentation__.Notify(64457)
	if ret.data == nil {
		__antithesis_instrumentation__.Notify(64463)
		return "", errors.Newf("expected DE-9IM string but found nothing")
	} else {
		__antithesis_instrumentation__.Notify(64464)
	}
	__antithesis_instrumentation__.Notify(64458)
	return string(cStringToSafeGoBytes(ret)), nil
}

func RelateBoundaryNodeRule(a geopb.EWKB, b geopb.EWKB, bnr int) (string, error) {
	__antithesis_instrumentation__.Notify(64465)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64469)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(64470)
	}
	__antithesis_instrumentation__.Notify(64466)
	var ret C.CR_GEOS_String
	if err := statusToError(C.CR_GEOS_RelateBoundaryNodeRule(g, goToCSlice(a), goToCSlice(b), C.int(bnr), &ret)); err != nil {
		__antithesis_instrumentation__.Notify(64471)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(64472)
	}
	__antithesis_instrumentation__.Notify(64467)
	if ret.data == nil {
		__antithesis_instrumentation__.Notify(64473)
		return "", errors.Newf("expected DE-9IM string but found nothing")
	} else {
		__antithesis_instrumentation__.Notify(64474)
	}
	__antithesis_instrumentation__.Notify(64468)
	return string(cStringToSafeGoBytes(ret)), nil
}

func RelatePattern(a geopb.EWKB, b geopb.EWKB, pattern string) (bool, error) {
	__antithesis_instrumentation__.Notify(64475)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64478)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(64479)
	}
	__antithesis_instrumentation__.Notify(64476)
	var ret C.char
	if err := statusToError(
		C.CR_GEOS_RelatePattern(g, goToCSlice(a), goToCSlice(b), goToCSlice([]byte(pattern)), &ret),
	); err != nil {
		__antithesis_instrumentation__.Notify(64480)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(64481)
	}
	__antithesis_instrumentation__.Notify(64477)
	return ret == 1, nil
}

func IsValid(ewkb geopb.EWKB) (bool, error) {
	__antithesis_instrumentation__.Notify(64482)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64485)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(64486)
	}
	__antithesis_instrumentation__.Notify(64483)
	var ret C.char
	if err := statusToError(
		C.CR_GEOS_IsValid(g, goToCSlice(ewkb), &ret),
	); err != nil {
		__antithesis_instrumentation__.Notify(64487)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(64488)
	}
	__antithesis_instrumentation__.Notify(64484)
	return ret == 1, nil
}

func IsValidReason(ewkb geopb.EWKB) (string, error) {
	__antithesis_instrumentation__.Notify(64489)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64492)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(64493)
	}
	__antithesis_instrumentation__.Notify(64490)
	var ret C.CR_GEOS_String
	if err := statusToError(
		C.CR_GEOS_IsValidReason(g, goToCSlice(ewkb), &ret),
	); err != nil {
		__antithesis_instrumentation__.Notify(64494)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(64495)
	}
	__antithesis_instrumentation__.Notify(64491)

	return string(cStringToSafeGoBytes(ret)), nil
}

func IsValidDetail(ewkb geopb.EWKB, flags int) (bool, string, geopb.EWKB, error) {
	__antithesis_instrumentation__.Notify(64496)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64499)
		return false, "", nil, err
	} else {
		__antithesis_instrumentation__.Notify(64500)
	}
	__antithesis_instrumentation__.Notify(64497)
	var retIsValid C.char
	var retReason C.CR_GEOS_String
	var retLocationEWKB C.CR_GEOS_String
	if err := statusToError(
		C.CR_GEOS_IsValidDetail(
			g,
			goToCSlice(ewkb),
			C.int(flags),
			&retIsValid,
			&retReason,
			&retLocationEWKB,
		),
	); err != nil {
		__antithesis_instrumentation__.Notify(64501)
		return false, "", nil, err
	} else {
		__antithesis_instrumentation__.Notify(64502)
	}
	__antithesis_instrumentation__.Notify(64498)
	return retIsValid == 1,
		string(cStringToSafeGoBytes(retReason)),
		cStringToSafeGoBytes(retLocationEWKB),
		nil
}

func MakeValid(ewkb geopb.EWKB) (geopb.EWKB, error) {
	__antithesis_instrumentation__.Notify(64503)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64506)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64507)
	}
	__antithesis_instrumentation__.Notify(64504)
	var cEWKB C.CR_GEOS_String
	if err := statusToError(C.CR_GEOS_MakeValid(g, goToCSlice(ewkb), &cEWKB)); err != nil {
		__antithesis_instrumentation__.Notify(64508)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64509)
	}
	__antithesis_instrumentation__.Notify(64505)
	return cStringToSafeGoBytes(cEWKB), nil
}

func SharedPaths(a geopb.EWKB, b geopb.EWKB) (geopb.EWKB, error) {
	__antithesis_instrumentation__.Notify(64510)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64513)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64514)
	}
	__antithesis_instrumentation__.Notify(64511)
	var cEWKB C.CR_GEOS_String
	if err := statusToError(
		C.CR_GEOS_SharedPaths(g, goToCSlice(a), goToCSlice(b), &cEWKB),
	); err != nil {
		__antithesis_instrumentation__.Notify(64515)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64516)
	}
	__antithesis_instrumentation__.Notify(64512)
	return cStringToSafeGoBytes(cEWKB), nil
}

func Node(a geopb.EWKB) (geopb.EWKB, error) {
	__antithesis_instrumentation__.Notify(64517)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64520)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64521)
	}
	__antithesis_instrumentation__.Notify(64518)
	var cEWKB C.CR_GEOS_String
	err = statusToError(C.CR_GEOS_Node(g, goToCSlice(a), &cEWKB))
	if err != nil {
		__antithesis_instrumentation__.Notify(64522)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64523)
	}
	__antithesis_instrumentation__.Notify(64519)
	return cStringToSafeGoBytes(cEWKB), nil
}

func VoronoiDiagram(a, env geopb.EWKB, tolerance float64, onlyEdges bool) (geopb.EWKB, error) {
	__antithesis_instrumentation__.Notify(64524)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64528)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64529)
	}
	__antithesis_instrumentation__.Notify(64525)
	var cEWKB C.CR_GEOS_String
	flag := 0
	if onlyEdges {
		__antithesis_instrumentation__.Notify(64530)
		flag = 1
	} else {
		__antithesis_instrumentation__.Notify(64531)
	}
	__antithesis_instrumentation__.Notify(64526)
	if err := statusToError(
		C.CR_GEOS_VoronoiDiagram(g, goToCSlice(a), goToCSlice(env), C.double(tolerance), C.int(flag), &cEWKB),
	); err != nil {
		__antithesis_instrumentation__.Notify(64532)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64533)
	}
	__antithesis_instrumentation__.Notify(64527)
	return cStringToSafeGoBytes(cEWKB), nil
}

func MinimumRotatedRectangle(ewkb geopb.EWKB) (geopb.EWKB, error) {
	__antithesis_instrumentation__.Notify(64534)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64537)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64538)
	}
	__antithesis_instrumentation__.Notify(64535)
	var cEWKB C.CR_GEOS_String
	if err := statusToError(
		C.CR_GEOS_MinimumRotatedRectangle(g, goToCSlice(ewkb), &cEWKB),
	); err != nil {
		__antithesis_instrumentation__.Notify(64539)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64540)
	}
	__antithesis_instrumentation__.Notify(64536)
	return cStringToSafeGoBytes(cEWKB), nil
}

func Snap(input, target geopb.EWKB, tolerance float64) (geopb.EWKB, error) {
	__antithesis_instrumentation__.Notify(64541)
	g, err := ensureInitInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(64544)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64545)
	}
	__antithesis_instrumentation__.Notify(64542)
	var cEWKB C.CR_GEOS_String
	if err := statusToError(
		C.CR_GEOS_Snap(g, goToCSlice(input), goToCSlice(target), C.double(tolerance), &cEWKB),
	); err != nil {
		__antithesis_instrumentation__.Notify(64546)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64547)
	}
	__antithesis_instrumentation__.Notify(64543)
	return cStringToSafeGoBytes(cEWKB), nil
}
