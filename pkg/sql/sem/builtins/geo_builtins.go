package builtins

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gojson "encoding/json"
	"fmt"
	"math/rand"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geogfn"
	"github.com/cockroachdb/cockroach/pkg/geo/geoindex"
	"github.com/cockroachdb/cockroach/pkg/geo/geomfn"
	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/geo/geoprojbase"
	"github.com/cockroachdb/cockroach/pkg/geo/geotransform"
	"github.com/cockroachdb/cockroach/pkg/geo/twkb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/paramparse"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/cockroach/pkg/util/json"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/golang/geo/s1"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/ewkb"
)

type libraryUsage uint64

const (
	usesGEOS libraryUsage = (1 << (iota + 1))
	usesS2
	usesGeographicLib
	usesPROJ
)

const usesSpheroidMessage = " Uses a spheroid to perform the operation."
const spheroidDistanceMessage = `"\n\nWhen operating on a spheroid, this function will use the sphere to calculate ` +
	`the closest two points. The spheroid distance between these two points is calculated using GeographicLib. ` +
	`This follows observed PostGIS behavior.`

const (
	defaultWKTDecimalDigits = 15
)

type infoBuilder struct {
	info         string
	libraryUsage libraryUsage
	precision    string
}

func (ib infoBuilder) String() string {
	__antithesis_instrumentation__.Notify(599858)
	var sb strings.Builder
	sb.WriteString(ib.info)
	if ib.precision != "" {
		__antithesis_instrumentation__.Notify(599864)
		sb.WriteString(fmt.Sprintf("\n\nThe calculations performed are have a precision of %s.", ib.precision))
	} else {
		__antithesis_instrumentation__.Notify(599865)
	}
	__antithesis_instrumentation__.Notify(599859)
	if ib.libraryUsage&usesGEOS != 0 {
		__antithesis_instrumentation__.Notify(599866)
		sb.WriteString("\n\nThis function utilizes the GEOS module.")
	} else {
		__antithesis_instrumentation__.Notify(599867)
	}
	__antithesis_instrumentation__.Notify(599860)
	if ib.libraryUsage&usesS2 != 0 {
		__antithesis_instrumentation__.Notify(599868)
		sb.WriteString("\n\nThis function utilizes the S2 library for spherical calculations.")
	} else {
		__antithesis_instrumentation__.Notify(599869)
	}
	__antithesis_instrumentation__.Notify(599861)
	if ib.libraryUsage&usesGeographicLib != 0 {
		__antithesis_instrumentation__.Notify(599870)
		sb.WriteString("\n\nThis function utilizes the GeographicLib library for spheroid calculations.")
	} else {
		__antithesis_instrumentation__.Notify(599871)
	}
	__antithesis_instrumentation__.Notify(599862)
	if ib.libraryUsage&usesPROJ != 0 {
		__antithesis_instrumentation__.Notify(599872)
		sb.WriteString("\n\nThis function utilizes the PROJ library for coordinate projections.")
	} else {
		__antithesis_instrumentation__.Notify(599873)
	}
	__antithesis_instrumentation__.Notify(599863)
	return sb.String()
}

var geomFromWKTOverload = stringOverload1(
	func(_ *tree.EvalContext, s string) (tree.Datum, error) {
		__antithesis_instrumentation__.Notify(599874)
		g, err := geo.ParseGeometryFromEWKT(geopb.EWKT(s), geopb.DefaultGeometrySRID, geo.DefaultSRIDIsHint)
		if err != nil {
			__antithesis_instrumentation__.Notify(599876)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(599877)
		}
		__antithesis_instrumentation__.Notify(599875)

		return tree.NewDGeometry(g), nil
	},
	types.Geometry,
	infoBuilder{info: "Returns the Geometry from a WKT or EWKT representation."}.String(),
	tree.VolatilityImmutable,
)

var geomFromWKBOverload = bytesOverload1(
	func(_ *tree.EvalContext, s string) (tree.Datum, error) {
		__antithesis_instrumentation__.Notify(599878)
		g, err := geo.ParseGeometryFromEWKB([]byte(s))
		if err != nil {
			__antithesis_instrumentation__.Notify(599880)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(599881)
		}
		__antithesis_instrumentation__.Notify(599879)
		return tree.NewDGeometry(g), nil
	},
	types.Geometry,
	infoBuilder{info: "Returns the Geometry from a WKB (or EWKB) representation."}.String(),
	tree.VolatilityImmutable,
)

func geometryFromTextCheckShapeBuiltin(shapeType geopb.ShapeType) builtinDefinition {
	__antithesis_instrumentation__.Notify(599882)
	return makeBuiltin(
		defProps(),
		stringOverload1(
			func(_ *tree.EvalContext, s string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(599883)
				g, err := geo.ParseGeometryFromEWKT(geopb.EWKT(s), geopb.DefaultGeometrySRID, geo.DefaultSRIDIsHint)
				if err != nil {
					__antithesis_instrumentation__.Notify(599886)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(599887)
				}
				__antithesis_instrumentation__.Notify(599884)
				if g.ShapeType2D() != shapeType {
					__antithesis_instrumentation__.Notify(599888)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(599889)
				}
				__antithesis_instrumentation__.Notify(599885)
				return tree.NewDGeometry(g), nil
			},
			types.Geometry,
			infoBuilder{
				info: fmt.Sprintf(
					"Returns the Geometry from a WKT or EWKT representation. If the shape underneath is not %s, NULL is returned.",
					shapeType.String(),
				),
			}.String(),
			tree.VolatilityImmutable,
		),
		tree.Overload{
			Types:      tree.ArgTypes{{"str", types.String}, {"srid", types.Int}},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(599890)
				s := string(tree.MustBeDString(args[0]))
				srid := geopb.SRID(tree.MustBeDInt(args[1]))
				g, err := geo.ParseGeometryFromEWKT(geopb.EWKT(s), srid, geo.DefaultSRIDShouldOverwrite)
				if err != nil {
					__antithesis_instrumentation__.Notify(599893)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(599894)
				}
				__antithesis_instrumentation__.Notify(599891)
				if g.ShapeType2D() != shapeType {
					__antithesis_instrumentation__.Notify(599895)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(599896)
				}
				__antithesis_instrumentation__.Notify(599892)
				return tree.NewDGeometry(g), nil
			},
			Info: infoBuilder{
				info: fmt.Sprintf(
					`Returns the Geometry from a WKT or EWKT representation with an SRID. If the shape underneath is not %s, NULL is returned. If the SRID is present in both the EWKT and the argument, the argument value is used.`,
					shapeType.String(),
				),
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	)
}

func geometryFromWKBCheckShapeBuiltin(shapeType geopb.ShapeType) builtinDefinition {
	__antithesis_instrumentation__.Notify(599897)
	return makeBuiltin(
		defProps(),
		bytesOverload1(
			func(_ *tree.EvalContext, s string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(599898)
				g, err := geo.ParseGeometryFromEWKB(geopb.EWKB(s))
				if err != nil {
					__antithesis_instrumentation__.Notify(599901)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(599902)
				}
				__antithesis_instrumentation__.Notify(599899)
				if g.ShapeType2D() != shapeType {
					__antithesis_instrumentation__.Notify(599903)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(599904)
				}
				__antithesis_instrumentation__.Notify(599900)
				return tree.NewDGeometry(g), nil
			},
			types.Geometry,
			infoBuilder{
				info: fmt.Sprintf(
					"Returns the Geometry from a WKB (or EWKB) representation. If the shape underneath is not %s, NULL is returned.",
					shapeType.String(),
				),
			}.String(),
			tree.VolatilityImmutable,
		),
		tree.Overload{
			Types:      tree.ArgTypes{{"wkb", types.Bytes}, {"srid", types.Int}},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(599905)
				s := string(tree.MustBeDBytes(args[0]))
				srid := geopb.SRID(tree.MustBeDInt(args[1]))
				g, err := geo.ParseGeometryFromEWKBAndSRID(geopb.EWKB(s), srid)
				if err != nil {
					__antithesis_instrumentation__.Notify(599908)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(599909)
				}
				__antithesis_instrumentation__.Notify(599906)
				if g.ShapeType2D() != shapeType {
					__antithesis_instrumentation__.Notify(599910)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(599911)
				}
				__antithesis_instrumentation__.Notify(599907)
				return tree.NewDGeometry(g), nil
			},
			Info: infoBuilder{
				info: fmt.Sprintf(
					`Returns the Geometry from a WKB (or EWKB) representation with an SRID. If the shape underneath is not %s, NULL is returned.`,
					shapeType.String(),
				),
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	)
}

var areaOverloadGeometry1 = geometryOverload1(
	func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
		__antithesis_instrumentation__.Notify(599912)
		ret, err := geomfn.Area(g.Geometry)
		if err != nil {
			__antithesis_instrumentation__.Notify(599914)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(599915)
		}
		__antithesis_instrumentation__.Notify(599913)
		return tree.NewDFloat(tree.DFloat(ret)), nil
	},
	types.Float,
	infoBuilder{
		info:         "Returns the area of the given geometry.",
		libraryUsage: usesGEOS,
	},
	tree.VolatilityImmutable,
)

var lengthOverloadGeometry1 = geometryOverload1(
	func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
		__antithesis_instrumentation__.Notify(599916)
		ret, err := geomfn.Length(g.Geometry)
		if err != nil {
			__antithesis_instrumentation__.Notify(599918)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(599919)
		}
		__antithesis_instrumentation__.Notify(599917)
		return tree.NewDFloat(tree.DFloat(ret)), nil
	},
	types.Float,
	infoBuilder{
		info: `Returns the length of the given geometry.

Note ST_Length is only valid for LineString - use ST_Perimeter for Polygon.`,
		libraryUsage: usesGEOS,
	},
	tree.VolatilityImmutable,
)

var perimeterOverloadGeometry1 = geometryOverload1(
	func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
		__antithesis_instrumentation__.Notify(599920)
		ret, err := geomfn.Perimeter(g.Geometry)
		if err != nil {
			__antithesis_instrumentation__.Notify(599922)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(599923)
		}
		__antithesis_instrumentation__.Notify(599921)
		return tree.NewDFloat(tree.DFloat(ret)), nil
	},
	types.Float,
	infoBuilder{
		info: `Returns the perimeter of the given geometry.

Note ST_Perimeter is only valid for Polygon - use ST_Length for LineString.`,
		libraryUsage: usesGEOS,
	},
	tree.VolatilityImmutable,
)

var stBufferInfoBuilder = infoBuilder{
	info: `Returns a Geometry that represents all points whose distance is less than or equal to the given distance
from the given Geometry.`,
	libraryUsage: usesGEOS,
}

var stBufferWithParamsInfoBuilder = infoBuilder{
	info: `Returns a Geometry that represents all points whose distance is less than or equal to the given distance from the
given Geometry.

This variant takes in a space separate parameter string, which will augment the buffer styles. Valid parameters are:
* quad_segs=<int>, default 8
* endcap=<round|flat|butt|square>, default round
* join=<round|mitre|miter|bevel>, default round
* side=<both|left|right>, default both
* mitre_limit=<float>, default 5.0`,
	libraryUsage: usesGEOS,
}

var stBufferWithQuadSegInfoBuilder = infoBuilder{
	info: `Returns a Geometry that represents all points whose distance is less than or equal to the given distance from the
given Geometry.

This variant approximates the circle into quad_seg segments per line (the default is 8).`,
	libraryUsage: usesGEOS,
}

var usingBestGeomProjectionWarning = `

This operation is done by transforming the object into a Geometry. This occurs by translating
the Geography objects into Geometry objects before applying an LAEA, UTM or Web Mercator
based projection based on the bounding boxes of the given Geography objects. When the result is
calculated, the result is transformed back into a Geography with SRID 4326.`

func performGeographyOperationUsingBestGeomProjection(
	g geo.Geography, f func(geo.Geometry) (geo.Geometry, error),
) (geo.Geography, error) {
	__antithesis_instrumentation__.Notify(599924)
	bestProj, err := geogfn.BestGeomProjection(g.BoundingRect())
	if err != nil {
		__antithesis_instrumentation__.Notify(599932)
		return geo.Geography{}, err
	} else {
		__antithesis_instrumentation__.Notify(599933)
	}
	__antithesis_instrumentation__.Notify(599925)
	geogDefaultProj, err := geoprojbase.Projection(geopb.DefaultGeographySRID)
	if err != nil {
		__antithesis_instrumentation__.Notify(599934)
		return geo.Geography{}, err
	} else {
		__antithesis_instrumentation__.Notify(599935)
	}
	__antithesis_instrumentation__.Notify(599926)
	gProj, err := geoprojbase.Projection(g.SRID())
	if err != nil {
		__antithesis_instrumentation__.Notify(599936)
		return geo.Geography{}, err
	} else {
		__antithesis_instrumentation__.Notify(599937)
	}
	__antithesis_instrumentation__.Notify(599927)

	inLatLonGeom, err := g.AsGeometry()
	if err != nil {
		__antithesis_instrumentation__.Notify(599938)
		return geo.Geography{}, err
	} else {
		__antithesis_instrumentation__.Notify(599939)
	}
	__antithesis_instrumentation__.Notify(599928)

	inProjectedGeom, err := geotransform.Transform(
		inLatLonGeom,
		gProj.Proj4Text,
		bestProj,
		g.SRID(),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(599940)
		return geo.Geography{}, err
	} else {
		__antithesis_instrumentation__.Notify(599941)
	}
	__antithesis_instrumentation__.Notify(599929)

	outProjectedGeom, err := f(inProjectedGeom)
	if err != nil {
		__antithesis_instrumentation__.Notify(599942)
		return geo.Geography{}, err
	} else {
		__antithesis_instrumentation__.Notify(599943)
	}
	__antithesis_instrumentation__.Notify(599930)

	outGeom, err := geotransform.Transform(
		outProjectedGeom,
		bestProj,
		geogDefaultProj.Proj4Text,
		geopb.DefaultGeographySRID,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(599944)
		return geo.Geography{}, err
	} else {
		__antithesis_instrumentation__.Notify(599945)
	}
	__antithesis_instrumentation__.Notify(599931)
	return outGeom.AsGeography()
}

func fitMaxDecimalDigitsToBounds(maxDecimalDigits int) int {
	__antithesis_instrumentation__.Notify(599946)
	if maxDecimalDigits < -1 {
		__antithesis_instrumentation__.Notify(599949)
		return -1
	} else {
		__antithesis_instrumentation__.Notify(599950)
	}
	__antithesis_instrumentation__.Notify(599947)
	if maxDecimalDigits > 64 {
		__antithesis_instrumentation__.Notify(599951)
		return 64
	} else {
		__antithesis_instrumentation__.Notify(599952)
	}
	__antithesis_instrumentation__.Notify(599948)
	return maxDecimalDigits
}

func makeMinimumBoundGenerator(
	ctx *tree.EvalContext, args tree.Datums,
) (tree.ValueGenerator, error) {
	__antithesis_instrumentation__.Notify(599953)
	geometry := tree.MustBeDGeometry(args[0])

	_, center, radius, err := geomfn.MinimumBoundingCircle(geometry.Geometry)
	if err != nil {
		__antithesis_instrumentation__.Notify(599955)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(599956)
	}
	__antithesis_instrumentation__.Notify(599954)
	return &minimumBoundRadiusGen{
		center: center,
		radius: radius,
		next:   true,
	}, nil
}

var minimumBoundingRadiusReturnType = types.MakeLabeledTuple(
	[]*types.T{types.Geometry, types.Float},
	[]string{"center", "radius"},
)

type minimumBoundRadiusGen struct {
	center geo.Geometry
	radius float64
	next   bool
}

func (m *minimumBoundRadiusGen) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(599957)
	return minimumBoundingRadiusReturnType
}

func (m *minimumBoundRadiusGen) Start(ctx context.Context, txn *kv.Txn) error {
	__antithesis_instrumentation__.Notify(599958)
	return nil
}

func (m *minimumBoundRadiusGen) Next(ctx context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(599959)
	if m.next {
		__antithesis_instrumentation__.Notify(599961)
		m.next = false
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(599962)
	}
	__antithesis_instrumentation__.Notify(599960)
	return false, nil
}

func (m *minimumBoundRadiusGen) Values() (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(599963)
	return []tree.Datum{tree.NewDGeometry(m.center),
		tree.NewDFloat(tree.DFloat(m.radius))}, nil
}

func (m *minimumBoundRadiusGen) Close(_ context.Context) {
	__antithesis_instrumentation__.Notify(599964)
}

func makeSubdividedGeometriesGeneratorFactory(expectMaxVerticesArg bool) tree.GeneratorFactory {
	__antithesis_instrumentation__.Notify(599965)
	return func(
		ctx *tree.EvalContext, args tree.Datums,
	) (tree.ValueGenerator, error) {
		__antithesis_instrumentation__.Notify(599966)
		geometry := tree.MustBeDGeometry(args[0])
		var maxVertices int
		if expectMaxVerticesArg {
			__antithesis_instrumentation__.Notify(599969)
			maxVertices = int(tree.MustBeDInt(args[1]))
		} else {
			__antithesis_instrumentation__.Notify(599970)
			maxVertices = 256
		}
		__antithesis_instrumentation__.Notify(599967)
		results, err := geomfn.Subdivide(geometry.Geometry, maxVertices)
		if err != nil {
			__antithesis_instrumentation__.Notify(599971)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(599972)
		}
		__antithesis_instrumentation__.Notify(599968)
		return &subdividedGeometriesGen{
			geometries: results,
			curr:       -1,
		}, nil
	}
}

type subdividedGeometriesGen struct {
	geometries []geo.Geometry
	curr       int
}

func (s *subdividedGeometriesGen) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(599973)
	return types.Geometry
}

func (s *subdividedGeometriesGen) Close(_ context.Context) {
	__antithesis_instrumentation__.Notify(599974)
}

func (s *subdividedGeometriesGen) Start(_ context.Context, _ *kv.Txn) error {
	__antithesis_instrumentation__.Notify(599975)
	s.curr = -1
	return nil
}

func (s *subdividedGeometriesGen) Values() (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(599976)
	return tree.Datums{tree.NewDGeometry(s.geometries[s.curr])}, nil
}

func (s *subdividedGeometriesGen) Next(_ context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(599977)
	s.curr++
	return s.curr < len(s.geometries), nil
}

var geoBuiltins = map[string]builtinDefinition{

	"postgis_addbbox": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(_ *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(599978)
				return g, nil
			},
			types.Geometry,
			infoBuilder{
				info: "Compatibility placeholder function with PostGIS. This does not perform any operation on the Geometry.",
			},
			tree.VolatilityImmutable,
		),
	),
	"postgis_dropbbox": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(_ *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(599979)
				return g, nil
			},
			types.Geometry,
			infoBuilder{
				info: "Compatibility placeholder function with PostGIS. This does not perform any operation on the Geometry.",
			},
			tree.VolatilityImmutable,
		),
	),
	"postgis_hasbbox": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(_ *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(599980)
				if g.Geometry.Empty() {
					__antithesis_instrumentation__.Notify(599983)
					return tree.DBoolFalse, nil
				} else {
					__antithesis_instrumentation__.Notify(599984)
				}
				__antithesis_instrumentation__.Notify(599981)
				if g.Geometry.ShapeType2D() == geopb.ShapeType_Point {
					__antithesis_instrumentation__.Notify(599985)
					return tree.DBoolFalse, nil
				} else {
					__antithesis_instrumentation__.Notify(599986)
				}
				__antithesis_instrumentation__.Notify(599982)
				return tree.DBoolTrue, nil
			},
			types.Bool,
			infoBuilder{
				info: "Returns whether a given Geometry has a bounding box. False for points and empty geometries; always true otherwise.",
			},
			tree.VolatilityImmutable,
		),
	),
	"postgis_getbbox": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(_ *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(599987)
				bbox := g.CartesianBoundingBox()
				if bbox == nil {
					__antithesis_instrumentation__.Notify(599989)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(599990)
				}
				__antithesis_instrumentation__.Notify(599988)
				return tree.NewDBox2D(*bbox), nil
			},
			types.Box2D,
			infoBuilder{
				info: "Returns a box2d encapsulating the given Geometry.",
			},
			tree.VolatilityImmutable,
		),
	),
	"postgis_extensions_upgrade": returnCompatibilityFixedStringBuiltin(
		"Upgrade completed, run SELECT postgis_full_version(); for details",
	),
	"postgis_full_version": returnCompatibilityFixedStringBuiltin(
		`POSTGIS="3.0.1 ec2a9aa" [EXTENSION] PGSQL="120" GEOS="3.8.1-CAPI-1.13.3" PROJ="4.9.3" LIBXML="2.9.10" LIBJSON="0.13.1" LIBPROTOBUF="1.4.2" WAGYU="0.4.3 (Internal)"`,
	),
	"postgis_geos_version":       returnCompatibilityFixedStringBuiltin("3.8.1-CAPI-1.13.3"),
	"postgis_libxml_version":     returnCompatibilityFixedStringBuiltin("2.9.10"),
	"postgis_lib_build_date":     returnCompatibilityFixedStringBuiltin("2020-03-06 18:23:24"),
	"postgis_lib_version":        returnCompatibilityFixedStringBuiltin("3.0.1"),
	"postgis_liblwgeom_version":  returnCompatibilityFixedStringBuiltin("3.0.1 ec2a9aa"),
	"postgis_proj_version":       returnCompatibilityFixedStringBuiltin("4.9.3"),
	"postgis_scripts_build_date": returnCompatibilityFixedStringBuiltin("2020-02-24 13:54:19"),
	"postgis_scripts_installed":  returnCompatibilityFixedStringBuiltin("3.0.1 ec2a9aa"),
	"postgis_scripts_released":   returnCompatibilityFixedStringBuiltin("3.0.1 ec2a9aa"),
	"postgis_version":            returnCompatibilityFixedStringBuiltin("3.0 USE_GEOS=1 USE_PROJ=1 USE_STATS=1"),
	"postgis_wagyu_version":      returnCompatibilityFixedStringBuiltin("0.4.3 (Internal)"),

	"st_s2covering": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(evalCtx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(599991)
				cfg, err := geoindex.GeometryIndexConfigForSRID(g.SRID())
				if err != nil {
					__antithesis_instrumentation__.Notify(599994)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(599995)
				}
				__antithesis_instrumentation__.Notify(599992)
				ret, err := geoindex.NewS2GeometryIndex(*cfg.S2Geometry).CoveringGeometry(evalCtx.Context, g.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(599996)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(599997)
				}
				__antithesis_instrumentation__.Notify(599993)
				return tree.NewDGeometry(ret), nil
			},
			types.Geometry,
			infoBuilder{
				info: "Returns a geometry which represents the S2 covering used by the index using the default index configuration.",
			},
			tree.VolatilityImmutable,
		),
		tree.Overload{
			Types:      tree.ArgTypes{{"geometry", types.Geometry}, {"settings", types.String}},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(599998)
				g := tree.MustBeDGeometry(args[0])
				params := tree.MustBeDString(args[1])

				startCfg, err := geoindex.GeometryIndexConfigForSRID(g.SRID())
				if err != nil {
					__antithesis_instrumentation__.Notify(600002)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600003)
				}
				__antithesis_instrumentation__.Notify(599999)
				cfg, err := applyGeoindexConfigStorageParams(evalCtx, *startCfg, string(params))
				if err != nil {
					__antithesis_instrumentation__.Notify(600004)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600005)
				}
				__antithesis_instrumentation__.Notify(600000)
				ret, err := geoindex.NewS2GeometryIndex(*cfg.S2Geometry).CoveringGeometry(evalCtx.Context, g.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600006)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600007)
				}
				__antithesis_instrumentation__.Notify(600001)
				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: `
Returns a geometry which represents the S2 covering used by the index using the index configuration specified
by the settings parameter.

The settings parameter uses the same format as the parameters inside the WITH in CREATE INDEX ... WITH (...),
e.g. CREATE INDEX t_idx ON t USING GIST(geom) WITH (s2_max_level=15, s2_level_mod=3) can be tried using
SELECT ST_S2Covering(geometry, 's2_max_level=15,s2_level_mod=3').
`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		geographyOverload1(
			func(evalCtx *tree.EvalContext, g *tree.DGeography) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600008)
				cfg := geoindex.DefaultGeographyIndexConfig().S2Geography
				ret, err := geoindex.NewS2GeographyIndex(*cfg).CoveringGeography(evalCtx.Context, g.Geography)
				if err != nil {
					__antithesis_instrumentation__.Notify(600010)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600011)
				}
				__antithesis_instrumentation__.Notify(600009)
				return tree.NewDGeography(ret), nil
			},
			types.Geography,
			infoBuilder{
				info: "Returns a geography which represents the S2 covering used by the index using the default index configuration.",
			},
			tree.VolatilityImmutable,
		),
		tree.Overload{
			Types:      tree.ArgTypes{{"geography", types.Geography}, {"settings", types.String}},
			ReturnType: tree.FixedReturnType(types.Geography),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600012)
				g := tree.MustBeDGeography(args[0])
				params := tree.MustBeDString(args[1])

				startCfg := geoindex.DefaultGeographyIndexConfig()
				cfg, err := applyGeoindexConfigStorageParams(evalCtx, *startCfg, string(params))
				if err != nil {
					__antithesis_instrumentation__.Notify(600015)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600016)
				}
				__antithesis_instrumentation__.Notify(600013)
				ret, err := geoindex.NewS2GeographyIndex(*cfg.S2Geography).CoveringGeography(evalCtx.Context, g.Geography)
				if err != nil {
					__antithesis_instrumentation__.Notify(600017)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600018)
				}
				__antithesis_instrumentation__.Notify(600014)
				return tree.NewDGeography(ret), nil
			},
			Info: infoBuilder{
				info: `
Returns a geography which represents the S2 covering used by the index using the index configuration specified
by the settings parameter.

The settings parameter uses the same format as the parameters inside the WITH in CREATE INDEX ... WITH (...),
e.g. CREATE INDEX t_idx ON t USING GIST(geom) WITH (s2_max_level=15, s2_level_mod=3) can be tried using
SELECT ST_S2Covering(geography, 's2_max_level=15,s2_level_mod=3').
`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),

	"st_geometryfromtext": makeBuiltin(
		defProps(),
		geomFromWKTOverload,
		tree.Overload{
			Types:      tree.ArgTypes{{"str", types.String}, {"srid", types.Int}},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600019)
				s := string(tree.MustBeDString(args[0]))
				srid := geopb.SRID(tree.MustBeDInt(args[1]))
				g, err := geo.ParseGeometryFromEWKT(geopb.EWKT(s), srid, geo.DefaultSRIDShouldOverwrite)
				if err != nil {
					__antithesis_instrumentation__.Notify(600021)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600022)
				}
				__antithesis_instrumentation__.Notify(600020)
				return tree.NewDGeometry(g), nil
			},
			Info: infoBuilder{
				info: `Returns the Geometry from a WKT or EWKT representation with an SRID. If the SRID is present in both the EWKT and the argument, the argument value is used.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_geomfromewkt": makeBuiltin(
		defProps(),
		stringOverload1(
			func(_ *tree.EvalContext, s string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600023)
				g, err := geo.ParseGeometryFromEWKT(geopb.EWKT(s), geopb.DefaultGeometrySRID, geo.DefaultSRIDIsHint)
				if err != nil {
					__antithesis_instrumentation__.Notify(600025)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600026)
				}
				__antithesis_instrumentation__.Notify(600024)
				return tree.NewDGeometry(g), nil
			},
			types.Geometry,
			infoBuilder{info: "Returns the Geometry from an EWKT representation."}.String(),
			tree.VolatilityImmutable,
		),
	),
	"st_wkbtosql": makeBuiltin(defProps(), geomFromWKBOverload),
	"st_wkttosql": makeBuiltin(defProps(), geomFromWKTOverload),
	"st_geomfromwkb": makeBuiltin(
		defProps(),
		geomFromWKBOverload,
		tree.Overload{
			Types:      tree.ArgTypes{{"bytes", types.Bytes}, {"srid", types.Int}},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600027)
				b := string(tree.MustBeDBytes(args[0]))
				srid := geopb.SRID(tree.MustBeDInt(args[1]))
				g, err := geo.ParseGeometryFromEWKBAndSRID(geopb.EWKB(b), srid)
				if err != nil {
					__antithesis_instrumentation__.Notify(600029)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600030)
				}
				__antithesis_instrumentation__.Notify(600028)
				return tree.NewDGeometry(g), nil
			},
			Info: infoBuilder{
				info: `Returns the Geometry from a WKB (or EWKB) representation with the given SRID set.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_geomfromewkb": makeBuiltin(
		defProps(),
		bytesOverload1(
			func(_ *tree.EvalContext, s string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600031)
				g, err := geo.ParseGeometryFromEWKB([]byte(s))
				if err != nil {
					__antithesis_instrumentation__.Notify(600033)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600034)
				}
				__antithesis_instrumentation__.Notify(600032)
				return tree.NewDGeometry(g), nil
			},
			types.Geometry,
			infoBuilder{info: "Returns the Geometry from an EWKB representation."}.String(),
			tree.VolatilityImmutable,
		),
	),
	"st_geomfromgeojson": makeBuiltin(
		defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"val", types.String}},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600035)
				g, err := geo.ParseGeometryFromGeoJSON([]byte(tree.MustBeDString(args[0])))
				if err != nil {
					__antithesis_instrumentation__.Notify(600037)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600038)
				}
				__antithesis_instrumentation__.Notify(600036)
				return tree.NewDGeometry(g), nil
			},

			PreferredOverload: true,
			Info:              infoBuilder{info: "Returns the Geometry from an GeoJSON representation."}.String(),
			Volatility:        tree.VolatilityImmutable,
		},
		jsonOverload1(
			func(_ *tree.EvalContext, s json.JSON) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600039)

				asString, err := s.AsText()
				if err != nil {
					__antithesis_instrumentation__.Notify(600043)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600044)
				}
				__antithesis_instrumentation__.Notify(600040)
				if asString == nil {
					__antithesis_instrumentation__.Notify(600045)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(600046)
				}
				__antithesis_instrumentation__.Notify(600041)
				g, err := geo.ParseGeometryFromGeoJSON([]byte(*asString))
				if err != nil {
					__antithesis_instrumentation__.Notify(600047)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600048)
				}
				__antithesis_instrumentation__.Notify(600042)
				return tree.NewDGeometry(g), nil
			},
			types.Geometry,
			infoBuilder{info: "Returns the Geometry from an GeoJSON representation."}.String(),
			tree.VolatilityImmutable,
		),
	),
	"st_makepoint": makeBuiltin(
		defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"x", types.Float}, {"y", types.Float}},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600049)
				x := float64(tree.MustBeDFloat(args[0]))
				y := float64(tree.MustBeDFloat(args[1]))
				g, err := geo.MakeGeometryFromLayoutAndPointCoords(geom.XY, []float64{x, y})
				if err != nil {
					__antithesis_instrumentation__.Notify(600051)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600052)
				}
				__antithesis_instrumentation__.Notify(600050)
				return tree.NewDGeometry(g), nil
			},
			Info:       infoBuilder{info: `Returns a new Point with the given X and Y coordinates.`}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"x", types.Float}, {"y", types.Float}, {"z", types.Float}},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600053)
				x := float64(tree.MustBeDFloat(args[0]))
				y := float64(tree.MustBeDFloat(args[1]))
				z := float64(tree.MustBeDFloat(args[2]))
				g, err := geo.MakeGeometryFromLayoutAndPointCoords(geom.XYZ, []float64{x, y, z})
				if err != nil {
					__antithesis_instrumentation__.Notify(600055)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600056)
				}
				__antithesis_instrumentation__.Notify(600054)
				return tree.NewDGeometry(g), nil
			},
			Info:       infoBuilder{info: `Returns a new Point with the given X, Y, and Z coordinates.`}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"x", types.Float}, {"y", types.Float}, {"z", types.Float}, {"m", types.Float}},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600057)
				x := float64(tree.MustBeDFloat(args[0]))
				y := float64(tree.MustBeDFloat(args[1]))
				z := float64(tree.MustBeDFloat(args[2]))
				m := float64(tree.MustBeDFloat(args[3]))
				g, err := geo.MakeGeometryFromLayoutAndPointCoords(geom.XYZM, []float64{x, y, z, m})
				if err != nil {
					__antithesis_instrumentation__.Notify(600059)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600060)
				}
				__antithesis_instrumentation__.Notify(600058)
				return tree.NewDGeometry(g), nil
			},
			Info:       infoBuilder{info: `Returns a new Point with the given X, Y, Z, and M coordinates.`}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_makepointm": makeBuiltin(
		defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"x", types.Float}, {"y", types.Float}, {"m", types.Float}},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600061)
				x := float64(tree.MustBeDFloat(args[0]))
				y := float64(tree.MustBeDFloat(args[1]))
				m := float64(tree.MustBeDFloat(args[2]))
				g, err := geo.MakeGeometryFromLayoutAndPointCoords(geom.XYM, []float64{x, y, m})
				if err != nil {
					__antithesis_instrumentation__.Notify(600063)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600064)
				}
				__antithesis_instrumentation__.Notify(600062)
				return tree.NewDGeometry(g), nil
			},
			Info:       infoBuilder{info: `Returns a new Point with the given X, Y, and M coordinates.`}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_makepolygon": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, outer *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600065)
				g, err := geomfn.MakePolygon(outer.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600067)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600068)
				}
				__antithesis_instrumentation__.Notify(600066)
				return tree.NewDGeometry(g), nil
			},
			types.Geometry,
			infoBuilder{
				info: `Returns a new Polygon with the given outer LineString.`,
			},
			tree.VolatilityImmutable,
		),
		tree.Overload{
			Types: tree.ArgTypes{
				{"outer", types.Geometry},
				{"interior", types.AnyArray},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600069)
				outer := tree.MustBeDGeometry(args[0])
				interiorArr := tree.MustBeDArray(args[1])
				interior := make([]geo.Geometry, len(interiorArr.Array))
				for i, v := range interiorArr.Array {
					__antithesis_instrumentation__.Notify(600072)
					g, ok := v.(*tree.DGeometry)
					if !ok {
						__antithesis_instrumentation__.Notify(600074)
						return nil, errors.Newf("argument must be LINESTRING geometries")
					} else {
						__antithesis_instrumentation__.Notify(600075)
					}
					__antithesis_instrumentation__.Notify(600073)
					interior[i] = g.Geometry
				}
				__antithesis_instrumentation__.Notify(600070)
				g, err := geomfn.MakePolygon(outer.Geometry, interior...)
				if err != nil {
					__antithesis_instrumentation__.Notify(600076)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600077)
				}
				__antithesis_instrumentation__.Notify(600071)
				return tree.NewDGeometry(g), nil
			},
			Info: infoBuilder{
				info: `Returns a new Polygon with the given outer LineString and interior (hole) LineString(s).`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_polygon": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"srid", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600078)
				g := tree.MustBeDGeometry(args[0])
				srid := tree.MustBeDInt(args[1])
				polygon, err := geomfn.MakePolygonWithSRID(g.Geometry, int(srid))
				if err != nil {
					__antithesis_instrumentation__.Notify(600080)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600081)
				}
				__antithesis_instrumentation__.Notify(600079)
				return tree.NewDGeometry(polygon), nil
			},
			Info: infoBuilder{
				info: `Returns a new Polygon from the given LineString and sets its SRID. It is equivalent ` +
					`to ST_MakePolygon with a single argument followed by ST_SetSRID.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),

	"st_geomcollfromtext":        geometryFromTextCheckShapeBuiltin(geopb.ShapeType_GeometryCollection),
	"st_geomcollfromwkb":         geometryFromWKBCheckShapeBuiltin(geopb.ShapeType_GeometryCollection),
	"st_linefromtext":            geometryFromTextCheckShapeBuiltin(geopb.ShapeType_LineString),
	"st_linefromwkb":             geometryFromWKBCheckShapeBuiltin(geopb.ShapeType_LineString),
	"st_linestringfromtext":      geometryFromTextCheckShapeBuiltin(geopb.ShapeType_LineString),
	"st_linestringfromwkb":       geometryFromWKBCheckShapeBuiltin(geopb.ShapeType_LineString),
	"st_mlinefromtext":           geometryFromTextCheckShapeBuiltin(geopb.ShapeType_MultiLineString),
	"st_mlinefromwkb":            geometryFromWKBCheckShapeBuiltin(geopb.ShapeType_MultiLineString),
	"st_mpointfromtext":          geometryFromTextCheckShapeBuiltin(geopb.ShapeType_MultiPoint),
	"st_mpointfromwkb":           geometryFromWKBCheckShapeBuiltin(geopb.ShapeType_MultiPoint),
	"st_mpolyfromtext":           geometryFromTextCheckShapeBuiltin(geopb.ShapeType_MultiPolygon),
	"st_mpolyfromwkb":            geometryFromWKBCheckShapeBuiltin(geopb.ShapeType_MultiPolygon),
	"st_multilinefromtext":       geometryFromTextCheckShapeBuiltin(geopb.ShapeType_MultiLineString),
	"st_multilinefromwkb":        geometryFromWKBCheckShapeBuiltin(geopb.ShapeType_MultiLineString),
	"st_multilinestringfromtext": geometryFromTextCheckShapeBuiltin(geopb.ShapeType_MultiLineString),
	"st_multilinestringfromwkb":  geometryFromWKBCheckShapeBuiltin(geopb.ShapeType_MultiLineString),
	"st_multipointfromtext":      geometryFromTextCheckShapeBuiltin(geopb.ShapeType_MultiPoint),
	"st_multipointfromwkb":       geometryFromWKBCheckShapeBuiltin(geopb.ShapeType_MultiPoint),
	"st_multipolyfromtext":       geometryFromTextCheckShapeBuiltin(geopb.ShapeType_MultiPolygon),
	"st_multipolyfromwkb":        geometryFromWKBCheckShapeBuiltin(geopb.ShapeType_MultiPolygon),
	"st_multipolygonfromtext":    geometryFromTextCheckShapeBuiltin(geopb.ShapeType_MultiPolygon),
	"st_multipolygonfromwkb":     geometryFromWKBCheckShapeBuiltin(geopb.ShapeType_MultiPolygon),
	"st_pointfromtext":           geometryFromTextCheckShapeBuiltin(geopb.ShapeType_Point),
	"st_pointfromwkb":            geometryFromWKBCheckShapeBuiltin(geopb.ShapeType_Point),
	"st_polyfromtext":            geometryFromTextCheckShapeBuiltin(geopb.ShapeType_Polygon),
	"st_polyfromwkb":             geometryFromWKBCheckShapeBuiltin(geopb.ShapeType_Polygon),
	"st_polygonfromtext":         geometryFromTextCheckShapeBuiltin(geopb.ShapeType_Polygon),
	"st_polygonfromwkb":          geometryFromWKBCheckShapeBuiltin(geopb.ShapeType_Polygon),

	"st_geographyfromtext": makeBuiltin(
		defProps(),
		stringOverload1(
			func(_ *tree.EvalContext, s string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600082)
				g, err := geo.ParseGeographyFromEWKT(geopb.EWKT(s), geopb.DefaultGeographySRID, geo.DefaultSRIDIsHint)
				if err != nil {
					__antithesis_instrumentation__.Notify(600084)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600085)
				}
				__antithesis_instrumentation__.Notify(600083)
				return tree.NewDGeography(g), nil
			},
			types.Geography,
			infoBuilder{info: "Returns the Geography from a WKT or EWKT representation."}.String(),
			tree.VolatilityImmutable,
		),
		tree.Overload{
			Types:      tree.ArgTypes{{"str", types.String}, {"srid", types.Int}},
			ReturnType: tree.FixedReturnType(types.Geography),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600086)
				s := string(tree.MustBeDString(args[0]))
				srid := geopb.SRID(tree.MustBeDInt(args[1]))
				g, err := geo.ParseGeographyFromEWKT(geopb.EWKT(s), srid, geo.DefaultSRIDShouldOverwrite)
				if err != nil {
					__antithesis_instrumentation__.Notify(600088)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600089)
				}
				__antithesis_instrumentation__.Notify(600087)
				return tree.NewDGeography(g), nil
			},
			Info: infoBuilder{
				info: `Returns the Geography from a WKT or EWKT representation with an SRID. If the SRID is present in both the EWKT and the argument, the argument value is used.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),

	"st_geogfromewkt": makeBuiltin(
		defProps(),
		stringOverload1(
			func(_ *tree.EvalContext, s string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600090)
				g, err := geo.ParseGeographyFromEWKT(geopb.EWKT(s), geopb.DefaultGeographySRID, geo.DefaultSRIDIsHint)
				if err != nil {
					__antithesis_instrumentation__.Notify(600092)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600093)
				}
				__antithesis_instrumentation__.Notify(600091)
				return tree.NewDGeography(g), nil
			},
			types.Geography,
			infoBuilder{info: "Returns the Geography from an EWKT representation."}.String(),
			tree.VolatilityImmutable,
		),
	),
	"st_geogfromwkb": makeBuiltin(
		defProps(),
		bytesOverload1(
			func(_ *tree.EvalContext, s string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600094)
				g, err := geo.ParseGeographyFromEWKB([]byte(s))
				if err != nil {
					__antithesis_instrumentation__.Notify(600096)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600097)
				}
				__antithesis_instrumentation__.Notify(600095)
				return tree.NewDGeography(g), nil
			},
			types.Geography,
			infoBuilder{info: "Returns the Geography from a WKB (or EWKB) representation."}.String(),
			tree.VolatilityImmutable,
		),
		tree.Overload{
			Types:      tree.ArgTypes{{"bytes", types.Bytes}, {"srid", types.Int}},
			ReturnType: tree.FixedReturnType(types.Geography),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600098)
				b := string(tree.MustBeDBytes(args[0]))
				srid := geopb.SRID(tree.MustBeDInt(args[1]))
				g, err := geo.ParseGeographyFromEWKBAndSRID(geopb.EWKB(b), srid)
				if err != nil {
					__antithesis_instrumentation__.Notify(600100)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600101)
				}
				__antithesis_instrumentation__.Notify(600099)
				return tree.NewDGeography(g), nil
			},
			Info: infoBuilder{
				info: `Returns the Geography from a WKB (or EWKB) representation with the given SRID set.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_geogfromewkb": makeBuiltin(
		defProps(),
		bytesOverload1(
			func(_ *tree.EvalContext, s string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600102)
				g, err := geo.ParseGeographyFromEWKB([]byte(s))
				if err != nil {
					__antithesis_instrumentation__.Notify(600104)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600105)
				}
				__antithesis_instrumentation__.Notify(600103)
				return tree.NewDGeography(g), nil
			},
			types.Geography,
			infoBuilder{info: "Returns the Geography from an EWKB representation."}.String(),
			tree.VolatilityImmutable,
		),
	),
	"st_geogfromgeojson": makeBuiltin(
		defProps(),
		stringOverload1(
			func(_ *tree.EvalContext, s string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600106)
				g, err := geo.ParseGeographyFromGeoJSON([]byte(s))
				if err != nil {
					__antithesis_instrumentation__.Notify(600108)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600109)
				}
				__antithesis_instrumentation__.Notify(600107)
				return tree.NewDGeography(g), nil
			},
			types.Geography,
			infoBuilder{info: "Returns the Geography from an GeoJSON representation."}.String(),
			tree.VolatilityImmutable,
		),
		jsonOverload1(
			func(_ *tree.EvalContext, s json.JSON) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600110)

				asString, err := s.AsText()
				if err != nil {
					__antithesis_instrumentation__.Notify(600114)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600115)
				}
				__antithesis_instrumentation__.Notify(600111)
				if asString == nil {
					__antithesis_instrumentation__.Notify(600116)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(600117)
				}
				__antithesis_instrumentation__.Notify(600112)
				g, err := geo.ParseGeographyFromGeoJSON([]byte(*asString))
				if err != nil {
					__antithesis_instrumentation__.Notify(600118)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600119)
				}
				__antithesis_instrumentation__.Notify(600113)
				return tree.NewDGeography(g), nil
			},
			types.Geography,
			infoBuilder{info: "Returns the Geography from an GeoJSON representation."}.String(),
			tree.VolatilityImmutable,
		),
	),
	"st_point": makeBuiltin(
		defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"x", types.Float}, {"y", types.Float}},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600120)
				x := float64(tree.MustBeDFloat(args[0]))
				y := float64(tree.MustBeDFloat(args[1]))
				g, err := geo.MakeGeometryFromPointCoords(x, y)
				if err != nil {
					__antithesis_instrumentation__.Notify(600122)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600123)
				}
				__antithesis_instrumentation__.Notify(600121)
				return tree.NewDGeometry(g), nil
			},
			Info:       infoBuilder{info: `Returns a new Point with the given X and Y coordinates.`}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_pointfromgeohash": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geohash", types.String},
				{"precision", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600124)
				g := tree.MustBeDString(args[0])
				p := tree.MustBeDInt(args[1])
				ret, err := geo.ParseGeometryPointFromGeoHash(string(g), int(p))
				if err != nil {
					__antithesis_instrumentation__.Notify(600126)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600127)
				}
				__antithesis_instrumentation__.Notify(600125)
				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: "Return a POINT Geometry from a GeoHash string with supplied precision.",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geohash", types.String},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600128)
				g := tree.MustBeDString(args[0])
				p := len(string(g))
				ret, err := geo.ParseGeometryPointFromGeoHash(string(g), p)
				if err != nil {
					__antithesis_instrumentation__.Notify(600130)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600131)
				}
				__antithesis_instrumentation__.Notify(600129)
				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: "Return a POINT Geometry from a GeoHash string with max precision.",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_geomfromgeohash": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geohash", types.String},
				{"precision", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600132)
				g := tree.MustBeDString(args[0])
				p := tree.MustBeDInt(args[1])
				bbox, err := geo.ParseCartesianBoundingBoxFromGeoHash(string(g), int(p))
				if err != nil {
					__antithesis_instrumentation__.Notify(600135)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600136)
				}
				__antithesis_instrumentation__.Notify(600133)
				ret, err := geo.MakeGeometryFromGeomT(bbox.ToGeomT(geopb.DefaultGeometrySRID))
				if err != nil {
					__antithesis_instrumentation__.Notify(600137)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600138)
				}
				__antithesis_instrumentation__.Notify(600134)
				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: "Return a POLYGON Geometry from a GeoHash string with supplied precision.",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geohash", types.String},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600139)
				g := tree.MustBeDString(args[0])
				p := len(string(g))
				bbox, err := geo.ParseCartesianBoundingBoxFromGeoHash(string(g), p)
				if err != nil {
					__antithesis_instrumentation__.Notify(600142)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600143)
				}
				__antithesis_instrumentation__.Notify(600140)
				ret, err := geo.MakeGeometryFromGeomT(bbox.ToGeomT(geopb.DefaultGeometrySRID))
				if err != nil {
					__antithesis_instrumentation__.Notify(600144)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600145)
				}
				__antithesis_instrumentation__.Notify(600141)
				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: "Return a POLYGON Geometry from a GeoHash string with max precision.",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_box2dfromgeohash": makeBuiltin(
		tree.FunctionProperties{NullableArgs: true},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geohash", types.String},
				{"precision", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.Box2D),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600146)
				if args[0] == tree.DNull {
					__antithesis_instrumentation__.Notify(600150)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(600151)
				}
				__antithesis_instrumentation__.Notify(600147)

				g := tree.MustBeDString(args[0])

				p := len(string(g))
				if args[1] != tree.DNull {
					__antithesis_instrumentation__.Notify(600152)
					p = int(tree.MustBeDInt(args[1]))
				} else {
					__antithesis_instrumentation__.Notify(600153)
				}
				__antithesis_instrumentation__.Notify(600148)

				bbox, err := geo.ParseCartesianBoundingBoxFromGeoHash(string(g), p)
				if err != nil {
					__antithesis_instrumentation__.Notify(600154)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600155)
				}
				__antithesis_instrumentation__.Notify(600149)
				return tree.NewDBox2D(bbox), nil
			},
			Info: infoBuilder{
				info: "Return a Box2D from a GeoHash string with supplied precision.",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geohash", types.String},
			},
			ReturnType: tree.FixedReturnType(types.Box2D),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600156)
				if args[0] == tree.DNull {
					__antithesis_instrumentation__.Notify(600159)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(600160)
				}
				__antithesis_instrumentation__.Notify(600157)
				g := tree.MustBeDString(args[0])
				p := len(string(g))
				bbox, err := geo.ParseCartesianBoundingBoxFromGeoHash(string(g), p)
				if err != nil {
					__antithesis_instrumentation__.Notify(600161)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600162)
				}
				__antithesis_instrumentation__.Notify(600158)
				return tree.NewDBox2D(bbox), nil
			},
			Info: infoBuilder{
				info: "Return a Box2D from a GeoHash string with max precision.",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),

	"st_astext": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(_ *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600163)
				wkt, err := geo.SpatialObjectToWKT(g.Geometry.SpatialObject(), defaultWKTDecimalDigits)
				return tree.NewDString(string(wkt)), err
			},
			types.String,
			infoBuilder{
				info: fmt.Sprintf("Returns the WKT representation of a given Geometry. A maximum of %d decimal digits is used.", defaultWKTDecimalDigits),
			},
			tree.VolatilityImmutable,
		),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"max_decimal_digits", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600164)
				g := tree.MustBeDGeometry(args[0])
				maxDecimalDigits := int(tree.MustBeDInt(args[1]))
				wkt, err := geo.SpatialObjectToWKT(g.Geometry.SpatialObject(), fitMaxDecimalDigitsToBounds(maxDecimalDigits))
				return tree.NewDString(string(wkt)), err
			},
			Info: infoBuilder{
				info: "Returns the WKT representation of a given Geometry. The max_decimal_digits parameter controls the maximum decimal digits to print after the `.`. Use -1 to print as many digits as required to rebuild the same number.",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		geographyOverload1(
			func(_ *tree.EvalContext, g *tree.DGeography) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600165)
				wkt, err := geo.SpatialObjectToWKT(g.Geography.SpatialObject(), defaultWKTDecimalDigits)
				return tree.NewDString(string(wkt)), err
			},
			types.String,
			infoBuilder{
				info: fmt.Sprintf("Returns the WKT representation of a given Geography. A default of %d decimal digits is used.", defaultWKTDecimalDigits),
			},
			tree.VolatilityImmutable,
		),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geography", types.Geography},
				{"max_decimal_digits", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600166)
				g := tree.MustBeDGeography(args[0])
				maxDecimalDigits := int(tree.MustBeDInt(args[1]))
				wkt, err := geo.SpatialObjectToWKT(g.Geography.SpatialObject(), fitMaxDecimalDigitsToBounds(maxDecimalDigits))
				return tree.NewDString(string(wkt)), err
			},
			Info: infoBuilder{
				info: "Returns the WKT representation of a given Geography. The max_decimal_digits parameter controls the maximum decimal digits to print after the `.`. Use -1 to print as many digits as required to rebuild the same number.",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_asewkt": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(_ *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600167)
				ewkt, err := geo.SpatialObjectToEWKT(g.Geometry.SpatialObject(), defaultWKTDecimalDigits)
				return tree.NewDString(string(ewkt)), err
			},
			types.String,
			infoBuilder{
				info: fmt.Sprintf("Returns the EWKT representation of a given Geometry. A maximum of %d decimal digits is used.", defaultWKTDecimalDigits),
			},
			tree.VolatilityImmutable,
		),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"max_decimal_digits", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600168)
				g := tree.MustBeDGeometry(args[0])
				maxDecimalDigits := int(tree.MustBeDInt(args[1]))
				ewkt, err := geo.SpatialObjectToEWKT(g.Geometry.SpatialObject(), fitMaxDecimalDigitsToBounds(maxDecimalDigits))
				return tree.NewDString(string(ewkt)), err
			},
			Info: infoBuilder{
				info: "Returns the WKT representation of a given Geometry. The max_decimal_digits parameter controls the maximum decimal digits to print after the `.`. Use -1 to print as many digits as required to rebuild the same number.",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		geographyOverload1(
			func(_ *tree.EvalContext, g *tree.DGeography) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600169)
				ewkt, err := geo.SpatialObjectToEWKT(g.Geography.SpatialObject(), defaultWKTDecimalDigits)
				return tree.NewDString(string(ewkt)), err
			},
			types.String,
			infoBuilder{
				info: fmt.Sprintf("Returns the EWKT representation of a given Geography. A default of %d decimal digits is used.", defaultWKTDecimalDigits),
			},
			tree.VolatilityImmutable,
		),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geography", types.Geography},
				{"max_decimal_digits", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600170)
				g := tree.MustBeDGeography(args[0])
				maxDecimalDigits := int(tree.MustBeDInt(args[1]))
				ewkt, err := geo.SpatialObjectToEWKT(g.Geography.SpatialObject(), fitMaxDecimalDigitsToBounds(maxDecimalDigits))
				return tree.NewDString(string(ewkt)), err
			},
			Info: infoBuilder{
				info: "Returns the EWKT representation of a given Geography. The max_decimal_digits parameter controls the maximum decimal digits to print after the `.`. Use -1 to print as many digits as required to rebuild the same number.",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_asbinary": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(_ *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600171)
				wkb, err := geo.SpatialObjectToWKB(g.Geometry.SpatialObject(), geo.DefaultEWKBEncodingFormat)
				return tree.NewDBytes(tree.DBytes(wkb)), err
			},
			types.Bytes,
			infoBuilder{info: "Returns the WKB representation of a given Geometry."},
			tree.VolatilityImmutable,
		),
		geographyOverload1(
			func(_ *tree.EvalContext, g *tree.DGeography) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600172)
				wkb, err := geo.SpatialObjectToWKB(g.Geography.SpatialObject(), geo.DefaultEWKBEncodingFormat)
				return tree.NewDBytes(tree.DBytes(wkb)), err
			},
			types.Bytes,
			infoBuilder{info: "Returns the WKB representation of a given Geography."},
			tree.VolatilityImmutable,
		),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"xdr_or_ndr", types.String},
			},
			ReturnType: tree.FixedReturnType(types.Bytes),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600173)
				g := tree.MustBeDGeometry(args[0])
				text := string(tree.MustBeDString(args[1]))

				wkb, err := geo.SpatialObjectToWKB(g.Geometry.SpatialObject(), geo.StringToByteOrder(text))
				return tree.NewDBytes(tree.DBytes(wkb)), err
			},
			Info: infoBuilder{
				info: "Returns the WKB representation of a given Geometry. " +
					"This variant has a second argument denoting the encoding - `xdr` for big endian and `ndr` for little endian.",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geography", types.Geography},
				{"xdr_or_ndr", types.String},
			},
			ReturnType: tree.FixedReturnType(types.Bytes),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600174)
				g := tree.MustBeDGeography(args[0])
				text := string(tree.MustBeDString(args[1]))

				wkb, err := geo.SpatialObjectToWKB(g.Geography.SpatialObject(), geo.StringToByteOrder(text))
				return tree.NewDBytes(tree.DBytes(wkb)), err
			},
			Info: infoBuilder{
				info: "Returns the WKB representation of a given Geography. " +
					"This variant has a second argument denoting the encoding - `xdr` for big endian and `ndr` for little endian.",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_asewkb": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(_ *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600175)
				return tree.NewDBytes(tree.DBytes(g.EWKB())), nil
			},
			types.Bytes,
			infoBuilder{info: "Returns the EWKB representation of a given Geometry."},
			tree.VolatilityImmutable,
		),
		geographyOverload1(
			func(_ *tree.EvalContext, g *tree.DGeography) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600176)
				return tree.NewDBytes(tree.DBytes(g.EWKB())), nil
			},
			types.Bytes,
			infoBuilder{info: "Returns the EWKB representation of a given Geography."},
			tree.VolatilityImmutable,
		),
	),
	"st_ashexwkb": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(_ *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600177)
				hexwkb, err := geo.SpatialObjectToWKBHex(g.Geometry.SpatialObject())
				return tree.NewDString(hexwkb), err
			},
			types.String,
			infoBuilder{info: "Returns the WKB representation in hex of a given Geometry."},
			tree.VolatilityImmutable,
		),
		geographyOverload1(
			func(_ *tree.EvalContext, g *tree.DGeography) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600178)
				hexwkb, err := geo.SpatialObjectToWKBHex(g.Geography.SpatialObject())
				return tree.NewDString(hexwkb), err
			},
			types.String,
			infoBuilder{info: "Returns the WKB representation in hex of a given Geography."},
			tree.VolatilityImmutable,
		),
	),
	"st_ashexewkb": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(_ *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600179)
				return tree.NewDString(g.Geometry.EWKBHex()), nil
			},
			types.String,
			infoBuilder{info: "Returns the EWKB representation in hex of a given Geometry."},
			tree.VolatilityImmutable,
		),
		geographyOverload1(
			func(_ *tree.EvalContext, g *tree.DGeography) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600180)
				return tree.NewDString(g.Geography.EWKBHex()), nil
			},
			types.String,
			infoBuilder{info: "Returns the EWKB representation in hex of a given Geography."},
			tree.VolatilityImmutable,
		),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"xdr_or_ndr", types.String},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600181)
				g := tree.MustBeDGeometry(args[0])
				text := string(tree.MustBeDString(args[1]))

				byteOrder := geo.StringToByteOrder(text)
				if byteOrder == geo.DefaultEWKBEncodingFormat {
					__antithesis_instrumentation__.Notify(600185)
					return tree.NewDString(fmt.Sprintf("%X", g.EWKB())), nil
				} else {
					__antithesis_instrumentation__.Notify(600186)
				}
				__antithesis_instrumentation__.Notify(600182)

				geomT, err := g.AsGeomT()
				if err != nil {
					__antithesis_instrumentation__.Notify(600187)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600188)
				}
				__antithesis_instrumentation__.Notify(600183)

				b, err := ewkb.Marshal(geomT, byteOrder)
				if err != nil {
					__antithesis_instrumentation__.Notify(600189)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600190)
				}
				__antithesis_instrumentation__.Notify(600184)

				return tree.NewDString(fmt.Sprintf("%X", b)), nil
			},
			Info: infoBuilder{
				info: "Returns the EWKB representation in hex of a given Geometry. " +
					"This variant has a second argument denoting the encoding - `xdr` for big endian and `ndr` for little endian.",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geography", types.Geography},
				{"xdr_or_ndr", types.String},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600191)
				g := tree.MustBeDGeography(args[0])
				text := string(tree.MustBeDString(args[1]))

				byteOrder := geo.StringToByteOrder(text)
				if byteOrder == geo.DefaultEWKBEncodingFormat {
					__antithesis_instrumentation__.Notify(600195)
					return tree.NewDString(fmt.Sprintf("%X", g.EWKB())), nil
				} else {
					__antithesis_instrumentation__.Notify(600196)
				}
				__antithesis_instrumentation__.Notify(600192)

				geomT, err := g.AsGeomT()
				if err != nil {
					__antithesis_instrumentation__.Notify(600197)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600198)
				}
				__antithesis_instrumentation__.Notify(600193)

				b, err := ewkb.Marshal(geomT, byteOrder)
				if err != nil {
					__antithesis_instrumentation__.Notify(600199)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600200)
				}
				__antithesis_instrumentation__.Notify(600194)

				return tree.NewDString(fmt.Sprintf("%X", b)), nil
			},
			Info: infoBuilder{
				info: "Returns the EWKB representation in hex of a given Geography. " +
					"This variant has a second argument denoting the encoding - `xdr` for big endian and `ndr` for little endian.",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_astwkb": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"precision_xy", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.Bytes),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600201)
				t, err := tree.MustBeDGeometry(args[0]).AsGeomT()
				if err != nil {
					__antithesis_instrumentation__.Notify(600204)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600205)
				}
				__antithesis_instrumentation__.Notify(600202)
				ret, err := twkb.Marshal(
					t,
					twkb.MarshalOptionPrecisionXY(int64(tree.MustBeDInt(args[1]))),
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(600206)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600207)
				}
				__antithesis_instrumentation__.Notify(600203)
				return tree.NewDBytes(tree.DBytes(ret)), nil
			},
			Info: infoBuilder{
				info: "Returns the TWKB representation of a given geometry.",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"precision_xy", types.Int},
				{"precision_z", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.Bytes),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600208)
				t, err := tree.MustBeDGeometry(args[0]).AsGeomT()
				if err != nil {
					__antithesis_instrumentation__.Notify(600211)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600212)
				}
				__antithesis_instrumentation__.Notify(600209)
				ret, err := twkb.Marshal(
					t,
					twkb.MarshalOptionPrecisionXY(int64(tree.MustBeDInt(args[1]))),
					twkb.MarshalOptionPrecisionZ(int64(tree.MustBeDInt(args[2]))),
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(600213)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600214)
				}
				__antithesis_instrumentation__.Notify(600210)
				return tree.NewDBytes(tree.DBytes(ret)), nil
			},
			Info: infoBuilder{
				info: "Returns the TWKB representation of a given geometry.",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"precision_xy", types.Int},
				{"precision_z", types.Int},
				{"precision_m", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.Bytes),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600215)
				t, err := tree.MustBeDGeometry(args[0]).AsGeomT()
				if err != nil {
					__antithesis_instrumentation__.Notify(600218)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600219)
				}
				__antithesis_instrumentation__.Notify(600216)
				ret, err := twkb.Marshal(
					t,
					twkb.MarshalOptionPrecisionXY(int64(tree.MustBeDInt(args[1]))),
					twkb.MarshalOptionPrecisionZ(int64(tree.MustBeDInt(args[2]))),
					twkb.MarshalOptionPrecisionM(int64(tree.MustBeDInt(args[3]))),
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(600220)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600221)
				}
				__antithesis_instrumentation__.Notify(600217)
				return tree.NewDBytes(tree.DBytes(ret)), nil
			},
			Info: infoBuilder{
				info: "Returns the TWKB representation of a given geometry.",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_askml": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(_ *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600222)
				kml, err := geo.SpatialObjectToKML(g.Geometry.SpatialObject())
				return tree.NewDString(kml), err
			},
			types.String,
			infoBuilder{info: "Returns the KML representation of a given Geometry."},
			tree.VolatilityImmutable,
		),
		geographyOverload1(
			func(_ *tree.EvalContext, g *tree.DGeography) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600223)
				kml, err := geo.SpatialObjectToKML(g.Geography.SpatialObject())
				return tree.NewDString(kml), err
			},
			types.String,
			infoBuilder{info: "Returns the KML representation of a given Geography."},
			tree.VolatilityImmutable,
		),
	),
	"st_geohash": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(_ *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600224)
				ret, err := geo.SpatialObjectToGeoHash(g.Geometry.SpatialObject(), geo.GeoHashAutoPrecision)
				if err != nil {
					__antithesis_instrumentation__.Notify(600226)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600227)
				}
				__antithesis_instrumentation__.Notify(600225)
				return tree.NewDString(ret), nil
			},
			types.String,
			infoBuilder{
				info: "Returns a GeoHash representation of the geometry with full precision if a point is provided, or with variable precision based on the size of the feature. This will error any coordinates are outside the bounds of longitude/latitude.",
			},
			tree.VolatilityImmutable,
		),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"precision", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600228)
				g := tree.MustBeDGeometry(args[0])
				p := tree.MustBeDInt(args[1])
				ret, err := geo.SpatialObjectToGeoHash(g.Geometry.SpatialObject(), int(p))
				if err != nil {
					__antithesis_instrumentation__.Notify(600230)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600231)
				}
				__antithesis_instrumentation__.Notify(600229)
				return tree.NewDString(ret), nil
			},
			Info: infoBuilder{
				info: "Returns a GeoHash representation of the geometry with the supplied precision. This will error any coordinates are outside the bounds of longitude/latitude.",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		geographyOverload1(
			func(_ *tree.EvalContext, g *tree.DGeography) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600232)
				ret, err := geo.SpatialObjectToGeoHash(g.Geography.SpatialObject(), geo.GeoHashAutoPrecision)
				if err != nil {
					__antithesis_instrumentation__.Notify(600234)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600235)
				}
				__antithesis_instrumentation__.Notify(600233)
				return tree.NewDString(ret), nil
			},
			types.String,
			infoBuilder{
				info: "Returns a GeoHash representation of the geeographywith full precision if a point is provided, or with variable precision based on the size of the feature.",
			},
			tree.VolatilityImmutable,
		),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geography", types.Geography},
				{"precision", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600236)
				g := tree.MustBeDGeography(args[0])
				p := tree.MustBeDInt(args[1])
				ret, err := geo.SpatialObjectToGeoHash(g.Geography.SpatialObject(), int(p))
				if err != nil {
					__antithesis_instrumentation__.Notify(600238)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600239)
				}
				__antithesis_instrumentation__.Notify(600237)
				return tree.NewDString(ret), nil
			},
			Info: infoBuilder{
				info: "Returns a GeoHash representation of the geography with the supplied precision.",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_asgeojson": makeBuiltin(
		defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"row", types.AnyTuple}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600240)
				tuple := tree.MustBeDTuple(args[0])
				return stAsGeoJSONFromTuple(
					ctx,
					tuple,
					"",
					geo.DefaultGeoJSONDecimalDigits,
					false,
				)
			},
			Info: infoBuilder{
				info: fmt.Sprintf(
					"Returns the GeoJSON representation of a given Geometry. Coordinates have a maximum of %d decimal digits.",
					geo.DefaultGeoJSONDecimalDigits,
				),
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"row", types.AnyTuple}, {"geo_column", types.String}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600241)
				tuple := tree.MustBeDTuple(args[0])
				return stAsGeoJSONFromTuple(
					ctx,
					tuple,
					string(tree.MustBeDString(args[1])),
					geo.DefaultGeoJSONDecimalDigits,
					false,
				)
			},
			Info: infoBuilder{
				info: fmt.Sprintf(
					"Returns the GeoJSON representation of a given Geometry, using geo_column as the geometry for the given Feature. Coordinates have a maximum of %d decimal digits.",
					geo.DefaultGeoJSONDecimalDigits,
				),
			}.String(),
			Volatility: tree.VolatilityStable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"row", types.AnyTuple},
				{"geo_column", types.String},
				{"max_decimal_digits", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600242)
				tuple := tree.MustBeDTuple(args[0])
				return stAsGeoJSONFromTuple(
					ctx,
					tuple,
					string(tree.MustBeDString(args[1])),
					fitMaxDecimalDigitsToBounds(int(tree.MustBeDInt(args[2]))),
					false,
				)
			},
			Info: infoBuilder{
				info: fmt.Sprintf(
					"Returns the GeoJSON representation of a given Geometry, using geo_column as the geometry for the given Feature. " +
						"max_decimal_digits will be output for each coordinate value.",
				),
			}.String(),
			Volatility: tree.VolatilityStable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"row", types.AnyTuple},
				{"geo_column", types.String},
				{"max_decimal_digits", types.Int},
				{"pretty", types.Bool},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600243)
				tuple := tree.MustBeDTuple(args[0])
				return stAsGeoJSONFromTuple(
					ctx,
					tuple,
					string(tree.MustBeDString(args[1])),
					fitMaxDecimalDigitsToBounds(int(tree.MustBeDInt(args[2]))),
					bool(tree.MustBeDBool(args[3])),
				)
			},
			Info: infoBuilder{
				info: fmt.Sprintf(
					"Returns the GeoJSON representation of a given Geometry, using geo_column as the geometry for the given Feature. " +
						"max_decimal_digits will be output for each coordinate value. Output will be pretty printed in JSON if pretty is true.",
				),
			}.String(),
			Volatility: tree.VolatilityStable,
		},
		geometryOverload1(
			func(_ *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600244)
				geojson, err := geo.SpatialObjectToGeoJSON(g.Geometry.SpatialObject(), geo.DefaultGeoJSONDecimalDigits, geo.SpatialObjectToGeoJSONFlagShortCRSIfNot4326)
				return tree.NewDString(string(geojson)), err
			},
			types.String,
			infoBuilder{
				info: fmt.Sprintf(
					"Returns the GeoJSON representation of a given Geometry. Coordinates have a maximum of %d decimal digits.",
					geo.DefaultGeoJSONDecimalDigits,
				),
			},
			tree.VolatilityImmutable,
		),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"max_decimal_digits", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600245)
				g := tree.MustBeDGeometry(args[0])
				maxDecimalDigits := int(tree.MustBeDInt(args[1]))
				geojson, err := geo.SpatialObjectToGeoJSON(g.Geometry.SpatialObject(), fitMaxDecimalDigitsToBounds(maxDecimalDigits), geo.SpatialObjectToGeoJSONFlagShortCRSIfNot4326)
				return tree.NewDString(string(geojson)), err
			},
			Info: infoBuilder{
				info: `Returns the GeoJSON representation of a given Geometry with max_decimal_digits output for each coordinate value.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"max_decimal_digits", types.Int},
				{"options", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600246)
				g := tree.MustBeDGeometry(args[0])
				maxDecimalDigits := int(tree.MustBeDInt(args[1]))
				options := geo.SpatialObjectToGeoJSONFlag(tree.MustBeDInt(args[2]))
				geojson, err := geo.SpatialObjectToGeoJSON(g.Geometry.SpatialObject(), fitMaxDecimalDigitsToBounds(maxDecimalDigits), options)
				return tree.NewDString(string(geojson)), err
			},
			Info: infoBuilder{
				info: `Returns the GeoJSON representation of a given Geometry with max_decimal_digits output for each coordinate value.

Options is a flag that can be bitmasked. The options are:
* 0: no option
* 1: GeoJSON BBOX
* 2: GeoJSON Short CRS (e.g EPSG:4326)
* 4: GeoJSON Long CRS (e.g urn:ogc:def:crs:EPSG::4326)
* 8: GeoJSON Short CRS if not EPSG:4326 (default for Geometry)
`}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		geographyOverload1(
			func(_ *tree.EvalContext, g *tree.DGeography) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600247)
				geojson, err := geo.SpatialObjectToGeoJSON(g.Geography.SpatialObject(), geo.DefaultGeoJSONDecimalDigits, geo.SpatialObjectToGeoJSONFlagZero)
				return tree.NewDString(string(geojson)), err
			},
			types.String,
			infoBuilder{
				info: fmt.Sprintf(
					"Returns the GeoJSON representation of a given Geography. Coordinates have a maximum of %d decimal digits.",
					geo.DefaultGeoJSONDecimalDigits,
				),
			},
			tree.VolatilityImmutable,
		),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geography", types.Geography},
				{"max_decimal_digits", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600248)
				g := tree.MustBeDGeography(args[0])
				maxDecimalDigits := int(tree.MustBeDInt(args[1]))
				geojson, err := geo.SpatialObjectToGeoJSON(g.Geography.SpatialObject(), fitMaxDecimalDigitsToBounds(maxDecimalDigits), geo.SpatialObjectToGeoJSONFlagZero)
				return tree.NewDString(string(geojson)), err
			},
			Info: infoBuilder{
				info: `Returns the GeoJSON representation of a given Geography with max_decimal_digits output for each coordinate value.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geography", types.Geography},
				{"max_decimal_digits", types.Int},
				{"options", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600249)
				g := tree.MustBeDGeography(args[0])
				maxDecimalDigits := int(tree.MustBeDInt(args[1]))
				options := geo.SpatialObjectToGeoJSONFlag(tree.MustBeDInt(args[2]))
				geojson, err := geo.SpatialObjectToGeoJSON(g.Geography.SpatialObject(), fitMaxDecimalDigitsToBounds(maxDecimalDigits), options)
				return tree.NewDString(string(geojson)), err
			},
			Info: infoBuilder{
				info: `Returns the GeoJSON representation of a given Geography with max_decimal_digits output for each coordinate value.

Options is a flag that can be bitmasked. The options are:
* 0: no option (default for Geography)
* 1: GeoJSON BBOX
* 2: GeoJSON Short CRS (e.g EPSG:4326)
* 4: GeoJSON Long CRS (e.g urn:ogc:def:crs:EPSG::4326)
* 8: GeoJSON Short CRS if not EPSG:4326
`}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_project": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geography", types.Geography},
				{"distance", types.Float},
				{"azimuth", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Geography),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600250)
				g := tree.MustBeDGeography(args[0])
				distance := float64(tree.MustBeDFloat(args[1]))
				azimuth := float64(tree.MustBeDFloat(args[2]))

				geog, err := geogfn.Project(g.Geography, distance, s1.Angle(azimuth))
				if err != nil {
					__antithesis_instrumentation__.Notify(600252)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600253)
				}
				__antithesis_instrumentation__.Notify(600251)

				return &tree.DGeography{Geography: geog}, nil
			},
			Info: infoBuilder{
				info: `Returns a point projected from a start point along a geodesic using a given distance and azimuth (bearing).
This is known as the direct geodesic problem.

The distance is given in meters. Negative values are supported.

The azimuth (also known as heading or bearing) is given in radians. It is measured clockwise from true north (azimuth zero).
East is azimuth π/2 (90 degrees); south is azimuth π (180 degrees); west is azimuth 3π/2 (270 degrees).
Negative azimuth values and values greater than 2π (360 degrees) are supported.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),

	"st_ndims": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600254)
				t, err := g.Geometry.AsGeomT()
				if err != nil {
					__antithesis_instrumentation__.Notify(600256)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600257)
				}
				__antithesis_instrumentation__.Notify(600255)
				switch t.Layout() {
				case geom.NoLayout:
					__antithesis_instrumentation__.Notify(600258)
					if gc, ok := t.(*geom.GeometryCollection); ok && func() bool {
						__antithesis_instrumentation__.Notify(600261)
						return gc.Empty() == true
					}() == true {
						__antithesis_instrumentation__.Notify(600262)
						return tree.NewDInt(tree.DInt(geom.XY.Stride())), nil
					} else {
						__antithesis_instrumentation__.Notify(600263)
					}
					__antithesis_instrumentation__.Notify(600259)
					return nil, errors.AssertionFailedf("no layout found on object")
				default:
					__antithesis_instrumentation__.Notify(600260)
					return tree.NewDInt(tree.DInt(t.Stride())), nil
				}
			},
			types.Int,
			infoBuilder{
				info: "Returns the number of coordinate dimensions of a given Geometry.",
			},
			tree.VolatilityImmutable,
		),
	),
	"st_dimension": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600264)
				dim, err := geomfn.Dimension(g.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600266)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600267)
				}
				__antithesis_instrumentation__.Notify(600265)
				return tree.NewDInt(tree.DInt(dim)), nil
			},
			types.Int,
			infoBuilder{
				info: "Returns the number of topological dimensions of a given Geometry.",
			},
			tree.VolatilityImmutable,
		),
	),
	"st_startpoint": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600268)
				t, err := g.Geometry.AsGeomT()
				if err != nil {
					__antithesis_instrumentation__.Notify(600271)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600272)
				}
				__antithesis_instrumentation__.Notify(600269)
				switch t := t.(type) {
				case *geom.LineString:
					__antithesis_instrumentation__.Notify(600273)
					if t.Empty() {
						__antithesis_instrumentation__.Notify(600276)
						return tree.DNull, nil
					} else {
						__antithesis_instrumentation__.Notify(600277)
					}
					__antithesis_instrumentation__.Notify(600274)
					coord := t.Coord(0)
					retG, err := geo.MakeGeometryFromGeomT(
						geom.NewPointFlat(geom.XY, []float64{coord.X(), coord.Y()}).SetSRID(t.SRID()),
					)
					if err != nil {
						__antithesis_instrumentation__.Notify(600278)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(600279)
					}
					__antithesis_instrumentation__.Notify(600275)
					return &tree.DGeometry{Geometry: retG}, nil
				}
				__antithesis_instrumentation__.Notify(600270)
				return tree.DNull, nil
			},
			types.Geometry,
			infoBuilder{
				info: "Returns the first point of a geometry which has shape LineString. Returns NULL if the geometry is not a LineString.",
			},
			tree.VolatilityImmutable,
		),
	),
	"st_summary": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600280)
				t, err := g.AsGeomT()
				if err != nil {
					__antithesis_instrumentation__.Notify(600283)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600284)
				}
				__antithesis_instrumentation__.Notify(600281)

				summary, err := geo.Summary(t, g.SpatialObject().BoundingBox != nil, g.ShapeType2D(), false)
				if err != nil {
					__antithesis_instrumentation__.Notify(600285)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600286)
				}
				__antithesis_instrumentation__.Notify(600282)

				return tree.NewDString(summary), nil
			},
			types.String,
			infoBuilder{
				info: `Returns a text summary of the contents of the geometry.

Flags shown square brackets after the geometry type have the following meaning:
* M: has M coordinate
* Z: has Z coordinate
* B: has a cached bounding box
* G: is geography
* S: has spatial reference system
`,
			},
			tree.VolatilityImmutable,
		),
		geographyOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeography) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600287)
				t, err := g.AsGeomT()
				if err != nil {
					__antithesis_instrumentation__.Notify(600290)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600291)
				}
				__antithesis_instrumentation__.Notify(600288)

				summary, err := geo.Summary(t, g.SpatialObject().BoundingBox != nil, g.ShapeType2D(), true)
				if err != nil {
					__antithesis_instrumentation__.Notify(600292)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600293)
				}
				__antithesis_instrumentation__.Notify(600289)

				return tree.NewDString(summary), nil
			},
			types.String,
			infoBuilder{
				info: `Returns a text summary of the contents of the geography.

Flags shown square brackets after the geometry type have the following meaning:
* M: has M coordinate
* Z: has Z coordinate
* B: has a cached bounding box
* G: is geography
* S: has spatial reference system
`,
			},
			tree.VolatilityImmutable,
		),
	),
	"st_endpoint": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600294)
				t, err := g.Geometry.AsGeomT()
				if err != nil {
					__antithesis_instrumentation__.Notify(600297)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600298)
				}
				__antithesis_instrumentation__.Notify(600295)
				switch t := t.(type) {
				case *geom.LineString:
					__antithesis_instrumentation__.Notify(600299)
					if t.Empty() {
						__antithesis_instrumentation__.Notify(600302)
						return tree.DNull, nil
					} else {
						__antithesis_instrumentation__.Notify(600303)
					}
					__antithesis_instrumentation__.Notify(600300)
					coord := t.Coord(t.NumCoords() - 1)
					retG, err := geo.MakeGeometryFromGeomT(
						geom.NewPointFlat(geom.XY, []float64{coord.X(), coord.Y()}).SetSRID(t.SRID()),
					)
					if err != nil {
						__antithesis_instrumentation__.Notify(600304)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(600305)
					}
					__antithesis_instrumentation__.Notify(600301)
					return &tree.DGeometry{Geometry: retG}, nil
				}
				__antithesis_instrumentation__.Notify(600296)
				return tree.DNull, nil
			},
			types.Geometry,
			infoBuilder{
				info: "Returns the last point of a geometry which has shape LineString. Returns NULL if the geometry is not a LineString.",
			},
			tree.VolatilityImmutable,
		),
	),
	"st_generatepoints": makeBuiltin(
		defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"geometry", types.Geometry}, {"npoints", types.Int4}},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600306)
				geometry := tree.MustBeDGeometry(args[0]).Geometry
				npoints := int(tree.MustBeDInt(args[1]))
				seed := timeutil.Now().Unix()
				generatedPoints, err := geomfn.GenerateRandomPoints(geometry, npoints, rand.New(rand.NewSource(seed)))
				if err != nil {
					__antithesis_instrumentation__.Notify(600308)
					if errors.Is(err, geomfn.ErrGenerateRandomPointsInvalidPoints) {
						__antithesis_instrumentation__.Notify(600310)
						return tree.DNull, nil
					} else {
						__antithesis_instrumentation__.Notify(600311)
					}
					__antithesis_instrumentation__.Notify(600309)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600312)
				}
				__antithesis_instrumentation__.Notify(600307)
				return tree.NewDGeometry(generatedPoints), nil
			},
			Info: infoBuilder{
				info: `Generates pseudo-random points until the requested number are found within the input area. Uses system time as a seed.
The requested number of points must be not larger than 65336.`,
			}.String(),
			Volatility: tree.VolatilityVolatile,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"geometry", types.Geometry}, {"npoints", types.Int4}, {"seed", types.Int4}},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600313)
				geometry := tree.MustBeDGeometry(args[0]).Geometry
				npoints := int(tree.MustBeDInt(args[1]))
				seed := int64(tree.MustBeDInt(args[2]))
				if seed < 1 {
					__antithesis_instrumentation__.Notify(600316)
					return nil, errors.New("seed must be greater than zero")
				} else {
					__antithesis_instrumentation__.Notify(600317)
				}
				__antithesis_instrumentation__.Notify(600314)
				generatedPoints, err := geomfn.GenerateRandomPoints(geometry, npoints, rand.New(rand.NewSource(seed)))
				if err != nil {
					__antithesis_instrumentation__.Notify(600318)
					if errors.Is(err, geomfn.ErrGenerateRandomPointsInvalidPoints) {
						__antithesis_instrumentation__.Notify(600320)
						return tree.DNull, nil
					} else {
						__antithesis_instrumentation__.Notify(600321)
					}
					__antithesis_instrumentation__.Notify(600319)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600322)
				}
				__antithesis_instrumentation__.Notify(600315)
				return tree.NewDGeometry(generatedPoints), nil
			},
			Info: infoBuilder{
				info: `Generates pseudo-random points until the requested number are found within the input area.
The requested number of points must be not larger than 65336.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_numpoints": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600323)
				t, err := g.Geometry.AsGeomT()
				if err != nil {
					__antithesis_instrumentation__.Notify(600326)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600327)
				}
				__antithesis_instrumentation__.Notify(600324)
				switch t := t.(type) {
				case *geom.LineString:
					__antithesis_instrumentation__.Notify(600328)
					return tree.NewDInt(tree.DInt(t.NumCoords())), nil
				}
				__antithesis_instrumentation__.Notify(600325)
				return tree.DNull, nil
			},
			types.Int,
			infoBuilder{
				info: "Returns the number of points in a LineString. Returns NULL if the Geometry is not a LineString.",
			},
			tree.VolatilityImmutable,
		),
	),
	"st_hasarc": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600329)

				return tree.DBoolFalse, nil
			},
			types.Bool,
			infoBuilder{
				info: "Returns whether there is a CIRCULARSTRING in the geometry.",
			},
			tree.VolatilityImmutable,
		),
	),
	"st_npoints": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600330)
				t, err := g.Geometry.AsGeomT()
				if err != nil {
					__antithesis_instrumentation__.Notify(600332)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600333)
				}
				__antithesis_instrumentation__.Notify(600331)
				return tree.NewDInt(tree.DInt(geomfn.CountVertices(t))), nil
			},
			types.Int,
			infoBuilder{
				info: "Returns the number of points in a given Geometry. Works for any shape type.",
			},
			tree.VolatilityImmutable,
		),
	),
	"st_points": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600334)
				points, err := geomfn.Points(g.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600336)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600337)
				}
				__antithesis_instrumentation__.Notify(600335)
				return tree.NewDGeometry(points), nil
			},
			types.Geometry,
			infoBuilder{
				info: "Returns all coordinates in the given Geometry as a MultiPoint, including duplicates.",
			},
			tree.VolatilityImmutable,
		),
	),
	"st_exteriorring": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600338)
				t, err := g.Geometry.AsGeomT()
				if err != nil {
					__antithesis_instrumentation__.Notify(600341)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600342)
				}
				__antithesis_instrumentation__.Notify(600339)
				switch t := t.(type) {
				case *geom.Polygon:
					__antithesis_instrumentation__.Notify(600343)
					var lineString *geom.LineString
					if t.Empty() {
						__antithesis_instrumentation__.Notify(600346)
						lineString = geom.NewLineString(t.Layout())
					} else {
						__antithesis_instrumentation__.Notify(600347)
						ring := t.LinearRing(0)
						lineString = geom.NewLineStringFlat(t.Layout(), ring.FlatCoords())
					}
					__antithesis_instrumentation__.Notify(600344)
					ret, err := geo.MakeGeometryFromGeomT(lineString.SetSRID(t.SRID()))
					if err != nil {
						__antithesis_instrumentation__.Notify(600348)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(600349)
					}
					__antithesis_instrumentation__.Notify(600345)
					return tree.NewDGeometry(ret), nil
				}
				__antithesis_instrumentation__.Notify(600340)
				return tree.DNull, nil
			},
			types.Geometry,
			infoBuilder{
				info: "Returns the exterior ring of a Polygon as a LineString. Returns NULL if the shape is not a Polygon.",
			},
			tree.VolatilityImmutable,
		),
	),
	"st_interiorringn": makeBuiltin(
		defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"geometry", types.Geometry}, {"n", types.Int}},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600350)
				g := tree.MustBeDGeometry(args[0])
				n := int(tree.MustBeDInt(args[1]))
				t, err := g.Geometry.AsGeomT()
				if err != nil {
					__antithesis_instrumentation__.Notify(600353)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600354)
				}
				__antithesis_instrumentation__.Notify(600351)
				switch t := t.(type) {
				case *geom.Polygon:
					__antithesis_instrumentation__.Notify(600355)

					if n >= t.NumLinearRings() || func() bool {
						__antithesis_instrumentation__.Notify(600358)
						return n <= 0 == true
					}() == true {
						__antithesis_instrumentation__.Notify(600359)
						return tree.DNull, nil
					} else {
						__antithesis_instrumentation__.Notify(600360)
					}
					__antithesis_instrumentation__.Notify(600356)
					ring := t.LinearRing(n)
					lineString := geom.NewLineStringFlat(t.Layout(), ring.FlatCoords()).SetSRID(t.SRID())
					ret, err := geo.MakeGeometryFromGeomT(lineString)
					if err != nil {
						__antithesis_instrumentation__.Notify(600361)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(600362)
					}
					__antithesis_instrumentation__.Notify(600357)
					return tree.NewDGeometry(ret), nil
				}
				__antithesis_instrumentation__.Notify(600352)
				return tree.DNull, nil
			},
			Info: infoBuilder{
				info: `Returns the n-th (1-indexed) interior ring of a Polygon as a LineString. Returns NULL if the shape is not a Polygon, or the ring does not exist.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_pointn": makeBuiltin(
		defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"geometry", types.Geometry}, {"n", types.Int}},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600363)
				g := tree.MustBeDGeometry(args[0])
				n := int(tree.MustBeDInt(args[1])) - 1
				if n < 0 {
					__antithesis_instrumentation__.Notify(600367)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(600368)
				}
				__antithesis_instrumentation__.Notify(600364)
				t, err := g.Geometry.AsGeomT()
				if err != nil {
					__antithesis_instrumentation__.Notify(600369)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600370)
				}
				__antithesis_instrumentation__.Notify(600365)
				switch t := t.(type) {
				case *geom.LineString:
					__antithesis_instrumentation__.Notify(600371)
					if n >= t.NumCoords() {
						__antithesis_instrumentation__.Notify(600374)
						return tree.DNull, nil
					} else {
						__antithesis_instrumentation__.Notify(600375)
					}
					__antithesis_instrumentation__.Notify(600372)
					g, err := geo.MakeGeometryFromGeomT(geom.NewPointFlat(t.Layout(), t.Coord(n)).SetSRID(t.SRID()))
					if err != nil {
						__antithesis_instrumentation__.Notify(600376)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(600377)
					}
					__antithesis_instrumentation__.Notify(600373)
					return tree.NewDGeometry(g), nil
				}
				__antithesis_instrumentation__.Notify(600366)
				return tree.DNull, nil
			},
			Info: infoBuilder{
				info: `Returns the n-th Point of a LineString (1-indexed). Returns NULL if out of bounds or not a LineString.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_geometryn": makeBuiltin(
		defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"geometry", types.Geometry}, {"n", types.Int}},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600378)
				g := tree.MustBeDGeometry(args[0])
				n := int(tree.MustBeDInt(args[1])) - 1
				if n < 0 {
					__antithesis_instrumentation__.Notify(600382)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(600383)
				}
				__antithesis_instrumentation__.Notify(600379)
				t, err := g.Geometry.AsGeomT()
				if err != nil {
					__antithesis_instrumentation__.Notify(600384)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600385)
				}
				__antithesis_instrumentation__.Notify(600380)
				switch t := t.(type) {
				case *geom.Point, *geom.Polygon, *geom.LineString:
					__antithesis_instrumentation__.Notify(600386)
					if n > 0 {
						__antithesis_instrumentation__.Notify(600401)
						return tree.DNull, nil
					} else {
						__antithesis_instrumentation__.Notify(600402)
					}
					__antithesis_instrumentation__.Notify(600387)
					return args[0], nil
				case *geom.MultiPoint:
					__antithesis_instrumentation__.Notify(600388)
					if n >= t.NumPoints() {
						__antithesis_instrumentation__.Notify(600403)
						return tree.DNull, nil
					} else {
						__antithesis_instrumentation__.Notify(600404)
					}
					__antithesis_instrumentation__.Notify(600389)
					g, err := geo.MakeGeometryFromGeomT(t.Point(n).SetSRID(t.SRID()))
					if err != nil {
						__antithesis_instrumentation__.Notify(600405)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(600406)
					}
					__antithesis_instrumentation__.Notify(600390)
					return tree.NewDGeometry(g), nil
				case *geom.MultiLineString:
					__antithesis_instrumentation__.Notify(600391)
					if n >= t.NumLineStrings() {
						__antithesis_instrumentation__.Notify(600407)
						return tree.DNull, nil
					} else {
						__antithesis_instrumentation__.Notify(600408)
					}
					__antithesis_instrumentation__.Notify(600392)
					g, err := geo.MakeGeometryFromGeomT(t.LineString(n).SetSRID(t.SRID()))
					if err != nil {
						__antithesis_instrumentation__.Notify(600409)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(600410)
					}
					__antithesis_instrumentation__.Notify(600393)
					return tree.NewDGeometry(g), nil
				case *geom.MultiPolygon:
					__antithesis_instrumentation__.Notify(600394)
					if n >= t.NumPolygons() {
						__antithesis_instrumentation__.Notify(600411)
						return tree.DNull, nil
					} else {
						__antithesis_instrumentation__.Notify(600412)
					}
					__antithesis_instrumentation__.Notify(600395)
					g, err := geo.MakeGeometryFromGeomT(t.Polygon(n).SetSRID(t.SRID()))
					if err != nil {
						__antithesis_instrumentation__.Notify(600413)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(600414)
					}
					__antithesis_instrumentation__.Notify(600396)
					return tree.NewDGeometry(g), nil
				case *geom.GeometryCollection:
					__antithesis_instrumentation__.Notify(600397)
					if n >= t.NumGeoms() {
						__antithesis_instrumentation__.Notify(600415)
						return tree.DNull, nil
					} else {
						__antithesis_instrumentation__.Notify(600416)
					}
					__antithesis_instrumentation__.Notify(600398)
					g, err := geo.MakeGeometryFromGeomT(t.Geom(n))
					if err != nil {
						__antithesis_instrumentation__.Notify(600417)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(600418)
					}
					__antithesis_instrumentation__.Notify(600399)
					gWithSRID, err := g.CloneWithSRID(geopb.SRID(t.SRID()))
					if err != nil {
						__antithesis_instrumentation__.Notify(600419)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(600420)
					}
					__antithesis_instrumentation__.Notify(600400)
					return tree.NewDGeometry(gWithSRID), nil
				}
				__antithesis_instrumentation__.Notify(600381)
				return tree.DNull, nil
			},
			Info: infoBuilder{
				info: `Returns the n-th Geometry (1-indexed). Returns NULL if out of bounds.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_minimumclearance": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600421)
				ret, err := geomfn.MinimumClearance(g.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600423)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600424)
				}
				__antithesis_instrumentation__.Notify(600422)
				return tree.NewDFloat(tree.DFloat(ret)), nil
			},
			types.Float,
			infoBuilder{
				info: `Returns the minimum distance a vertex can move before producing an invalid geometry. ` +
					`Returns Infinity if no minimum clearance can be found (e.g. for a single point).`,
			},
			tree.VolatilityImmutable,
		),
	),
	"st_minimumclearanceline": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600425)
				ret, err := geomfn.MinimumClearanceLine(g.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600427)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600428)
				}
				__antithesis_instrumentation__.Notify(600426)
				return tree.NewDGeometry(ret), nil
			},
			types.Geometry,
			infoBuilder{
				info: `Returns a LINESTRING spanning the minimum distance a vertex can move before producing ` +
					`an invalid geometry. If no minimum clearance can be found (e.g. for a single point), an ` +
					`empty LINESTRING is returned.`,
			},
			tree.VolatilityImmutable,
		),
	),
	"st_numinteriorrings": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600429)
				t, err := g.Geometry.AsGeomT()
				if err != nil {
					__antithesis_instrumentation__.Notify(600432)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600433)
				}
				__antithesis_instrumentation__.Notify(600430)
				switch t := t.(type) {
				case *geom.Polygon:
					__antithesis_instrumentation__.Notify(600434)
					numRings := t.NumLinearRings()
					if numRings <= 1 {
						__antithesis_instrumentation__.Notify(600436)
						return tree.NewDInt(0), nil
					} else {
						__antithesis_instrumentation__.Notify(600437)
					}
					__antithesis_instrumentation__.Notify(600435)
					return tree.NewDInt(tree.DInt(numRings - 1)), nil
				}
				__antithesis_instrumentation__.Notify(600431)
				return tree.DNull, nil
			},
			types.Int,
			infoBuilder{
				info: "Returns the number of interior rings in a Polygon Geometry. Returns NULL if the shape is not a Polygon.",
			},
			tree.VolatilityImmutable,
		),
	),
	"st_nrings": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600438)
				t, err := g.Geometry.AsGeomT()
				if err != nil {
					__antithesis_instrumentation__.Notify(600441)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600442)
				}
				__antithesis_instrumentation__.Notify(600439)
				switch t := t.(type) {
				case *geom.Polygon:
					__antithesis_instrumentation__.Notify(600443)
					return tree.NewDInt(tree.DInt(t.NumLinearRings())), nil
				}
				__antithesis_instrumentation__.Notify(600440)
				return tree.NewDInt(0), nil
			},
			types.Int,
			infoBuilder{
				info: "Returns the number of rings in a Polygon Geometry. Returns 0 if the shape is not a Polygon.",
			},
			tree.VolatilityImmutable,
		),
	),
	"st_force2d": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600444)
				ret, err := geomfn.ForceLayout(g.Geometry, geom.XY)
				if err != nil {
					__antithesis_instrumentation__.Notify(600446)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600447)
				}
				__antithesis_instrumentation__.Notify(600445)
				return tree.NewDGeometry(ret), nil
			},
			types.Geometry,
			infoBuilder{
				info: "Returns a Geometry that is forced into XY layout with any Z or M dimensions discarded.",
			},
			tree.VolatilityImmutable,
		),
	),

	"st_force3dz": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600448)
				ret, err := geomfn.ForceLayout(g.Geometry, geom.XYZ)
				if err != nil {
					__antithesis_instrumentation__.Notify(600450)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600451)
				}
				__antithesis_instrumentation__.Notify(600449)
				return tree.NewDGeometry(ret), nil
			},
			types.Geometry,
			infoBuilder{
				info: "Returns a Geometry that is forced into XYZ layout. " +
					"If a Z coordinate doesn't exist, it will be set to 0. " +
					"If a M coordinate is present, it will be discarded.",
			},
			tree.VolatilityImmutable,
		),
		tree.Overload{
			Types:      tree.ArgTypes{{"geometry", types.Geometry}, {"defaultZ", types.Float}},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600452)
				g := tree.MustBeDGeometry(args[0])
				defaultZ := tree.MustBeDFloat(args[1])

				ret, err := geomfn.ForceLayoutWithDefaultZ(g.Geometry, geom.XYZ, float64(defaultZ))
				if err != nil {
					__antithesis_instrumentation__.Notify(600454)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600455)
				}
				__antithesis_instrumentation__.Notify(600453)
				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{info: "Returns a Geometry that is forced into XYZ layout. " +
				"If a Z coordinate doesn't exist, it will be set to the specified default Z value. " +
				"If a M coordinate is present, it will be discarded."}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_force3dm": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600456)
				ret, err := geomfn.ForceLayout(g.Geometry, geom.XYM)
				if err != nil {
					__antithesis_instrumentation__.Notify(600458)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600459)
				}
				__antithesis_instrumentation__.Notify(600457)
				return tree.NewDGeometry(ret), nil
			},
			types.Geometry,
			infoBuilder{
				info: "Returns a Geometry that is forced into XYM layout. " +
					"If a M coordinate doesn't exist, it will be set to 0. " +
					"If a Z coordinate is present, it will be discarded.",
			},
			tree.VolatilityImmutable,
		),
		tree.Overload{
			Types:      tree.ArgTypes{{"geometry", types.Geometry}, {"defaultM", types.Float}},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600460)
				g := tree.MustBeDGeometry(args[0])
				defaultM := tree.MustBeDFloat(args[1])

				ret, err := geomfn.ForceLayoutWithDefaultM(g.Geometry, geom.XYM, float64(defaultM))
				if err != nil {
					__antithesis_instrumentation__.Notify(600462)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600463)
				}
				__antithesis_instrumentation__.Notify(600461)
				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{info: "Returns a Geometry that is forced into XYM layout. " +
				"If a M coordinate doesn't exist, it will be set to the specified default M value. " +
				"If a Z coordinate is present, it will be discarded."}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_force4d": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600464)
				ret, err := geomfn.ForceLayout(g.Geometry, geom.XYZM)
				if err != nil {
					__antithesis_instrumentation__.Notify(600466)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600467)
				}
				__antithesis_instrumentation__.Notify(600465)
				return tree.NewDGeometry(ret), nil
			},
			types.Geometry,
			infoBuilder{
				info: "Returns a Geometry that is forced into XYZM layout. " +
					"If a Z coordinate doesn't exist, it will be set to 0. " +
					"If a M coordinate doesn't exist, it will be set to 0.",
			},
			tree.VolatilityImmutable,
		),
		tree.Overload{
			Types:      tree.ArgTypes{{"geometry", types.Geometry}, {"defaultZ", types.Float}},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600468)
				g := tree.MustBeDGeometry(args[0])
				defaultZ := tree.MustBeDFloat(args[1])

				ret, err := geomfn.ForceLayoutWithDefaultZ(g.Geometry, geom.XYZM, float64(defaultZ))
				if err != nil {
					__antithesis_instrumentation__.Notify(600470)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600471)
				}
				__antithesis_instrumentation__.Notify(600469)
				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{info: "Returns a Geometry that is forced into XYZ layout. " +
				"If a Z coordinate doesn't exist, it will be set to the specified default Z value. " +
				"If a M coordinate doesn't exist, it will be set to 0."}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"geometry", types.Geometry}, {"defaultZ", types.Float}, {"defaultM", types.Float}},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600472)
				g := tree.MustBeDGeometry(args[0])
				defaultZ := tree.MustBeDFloat(args[1])
				defaultM := tree.MustBeDFloat(args[2])

				ret, err := geomfn.ForceLayoutWithDefaultZM(g.Geometry, geom.XYZM, float64(defaultZ), float64(defaultM))
				if err != nil {
					__antithesis_instrumentation__.Notify(600474)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600475)
				}
				__antithesis_instrumentation__.Notify(600473)
				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{info: "Returns a Geometry that is forced into XYZ layout. " +
				"If a Z coordinate doesn't exist, it will be set to the specified Z value. " +
				"If a M coordinate doesn't exist, it will be set to the specified M value."}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_forcepolygoncw": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600476)
				ret, err := geomfn.ForcePolygonOrientation(g.Geometry, geomfn.OrientationCW)
				if err != nil {
					__antithesis_instrumentation__.Notify(600478)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600479)
				}
				__antithesis_instrumentation__.Notify(600477)
				return tree.NewDGeometry(ret), nil
			},
			types.Geometry,
			infoBuilder{
				info: "Returns a Geometry where all Polygon objects have exterior rings in the clockwise orientation and interior rings in the counter-clockwise orientation. Non-Polygon objects are unchanged.",
			},
			tree.VolatilityImmutable,
		),
	),
	"st_forcepolygonccw": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600480)
				ret, err := geomfn.ForcePolygonOrientation(g.Geometry, geomfn.OrientationCCW)
				if err != nil {
					__antithesis_instrumentation__.Notify(600482)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600483)
				}
				__antithesis_instrumentation__.Notify(600481)
				return tree.NewDGeometry(ret), nil
			},
			types.Geometry,
			infoBuilder{
				info: "Returns a Geometry where all Polygon objects have exterior rings in the counter-clockwise orientation and interior rings in the clockwise orientation. Non-Polygon objects are unchanged.",
			},
			tree.VolatilityImmutable,
		),
	),
	"st_ispolygoncw": makeBuiltin(
		defProps(),
		geometryOverload1UnaryPredicate(
			func(g geo.Geometry) (bool, error) {
				__antithesis_instrumentation__.Notify(600484)
				return geomfn.HasPolygonOrientation(g, geomfn.OrientationCW)
			},
			infoBuilder{
				info: "Returns whether the Polygon objects inside the Geometry have exterior rings in the clockwise orientation and interior rings in the counter-clockwise orientation. Non-Polygon objects are considered clockwise.",
			},
		),
	),
	"st_ispolygonccw": makeBuiltin(
		defProps(),
		geometryOverload1UnaryPredicate(
			func(g geo.Geometry) (bool, error) {
				__antithesis_instrumentation__.Notify(600485)
				return geomfn.HasPolygonOrientation(g, geomfn.OrientationCCW)
			},
			infoBuilder{
				info: "Returns whether the Polygon objects inside the Geometry have exterior rings in the counter-clockwise orientation and interior rings in the clockwise orientation. Non-Polygon objects are considered counter-clockwise.",
			},
		),
	),
	"st_numgeometries": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600486)
				t, err := g.Geometry.AsGeomT()
				if err != nil {
					__antithesis_instrumentation__.Notify(600489)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600490)
				}
				__antithesis_instrumentation__.Notify(600487)
				switch t := t.(type) {
				case *geom.Point:
					__antithesis_instrumentation__.Notify(600491)
					if t.Empty() {
						__antithesis_instrumentation__.Notify(600505)
						return tree.NewDInt(0), nil
					} else {
						__antithesis_instrumentation__.Notify(600506)
					}
					__antithesis_instrumentation__.Notify(600492)
					return tree.NewDInt(1), nil
				case *geom.LineString:
					__antithesis_instrumentation__.Notify(600493)
					if t.Empty() {
						__antithesis_instrumentation__.Notify(600507)
						return tree.NewDInt(0), nil
					} else {
						__antithesis_instrumentation__.Notify(600508)
					}
					__antithesis_instrumentation__.Notify(600494)
					return tree.NewDInt(1), nil
				case *geom.Polygon:
					__antithesis_instrumentation__.Notify(600495)
					if t.Empty() {
						__antithesis_instrumentation__.Notify(600509)
						return tree.NewDInt(0), nil
					} else {
						__antithesis_instrumentation__.Notify(600510)
					}
					__antithesis_instrumentation__.Notify(600496)
					return tree.NewDInt(1), nil
				case *geom.MultiPoint:
					__antithesis_instrumentation__.Notify(600497)
					if t.Empty() {
						__antithesis_instrumentation__.Notify(600511)
						return tree.NewDInt(0), nil
					} else {
						__antithesis_instrumentation__.Notify(600512)
					}
					__antithesis_instrumentation__.Notify(600498)
					return tree.NewDInt(tree.DInt(t.NumPoints())), nil
				case *geom.MultiLineString:
					__antithesis_instrumentation__.Notify(600499)
					if t.Empty() {
						__antithesis_instrumentation__.Notify(600513)
						return tree.NewDInt(0), nil
					} else {
						__antithesis_instrumentation__.Notify(600514)
					}
					__antithesis_instrumentation__.Notify(600500)
					return tree.NewDInt(tree.DInt(t.NumLineStrings())), nil
				case *geom.MultiPolygon:
					__antithesis_instrumentation__.Notify(600501)
					if t.Empty() {
						__antithesis_instrumentation__.Notify(600515)
						return tree.NewDInt(0), nil
					} else {
						__antithesis_instrumentation__.Notify(600516)
					}
					__antithesis_instrumentation__.Notify(600502)
					return tree.NewDInt(tree.DInt(t.NumPolygons())), nil
				case *geom.GeometryCollection:
					__antithesis_instrumentation__.Notify(600503)
					if t.Empty() {
						__antithesis_instrumentation__.Notify(600517)
						return tree.NewDInt(0), nil
					} else {
						__antithesis_instrumentation__.Notify(600518)
					}
					__antithesis_instrumentation__.Notify(600504)
					return tree.NewDInt(tree.DInt(t.NumGeoms())), nil
				}
				__antithesis_instrumentation__.Notify(600488)
				return nil, errors.Newf("unknown type: %T", t)
			},
			types.Int,
			infoBuilder{
				info: "Returns the number of shapes inside a given Geometry.",
			},
			tree.VolatilityImmutable,
		),
	),
	"st_x": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600519)
				t, err := g.Geometry.AsGeomT()
				if err != nil {
					__antithesis_instrumentation__.Notify(600522)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600523)
				}
				__antithesis_instrumentation__.Notify(600520)
				switch t := t.(type) {
				case *geom.Point:
					__antithesis_instrumentation__.Notify(600524)
					if t.Empty() {
						__antithesis_instrumentation__.Notify(600526)
						return tree.DNull, nil
					} else {
						__antithesis_instrumentation__.Notify(600527)
					}
					__antithesis_instrumentation__.Notify(600525)
					return tree.NewDFloat(tree.DFloat(t.X())), nil
				}
				__antithesis_instrumentation__.Notify(600521)

				return nil, errors.Newf("argument to st_x() must have shape POINT")
			},
			types.Float,
			infoBuilder{
				info: "Returns the X coordinate of a geometry if it is a Point.",
			},
			tree.VolatilityImmutable,
		),
	),
	"st_y": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600528)
				t, err := g.Geometry.AsGeomT()
				if err != nil {
					__antithesis_instrumentation__.Notify(600531)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600532)
				}
				__antithesis_instrumentation__.Notify(600529)
				switch t := t.(type) {
				case *geom.Point:
					__antithesis_instrumentation__.Notify(600533)
					if t.Empty() {
						__antithesis_instrumentation__.Notify(600535)
						return tree.DNull, nil
					} else {
						__antithesis_instrumentation__.Notify(600536)
					}
					__antithesis_instrumentation__.Notify(600534)
					return tree.NewDFloat(tree.DFloat(t.Y())), nil
				}
				__antithesis_instrumentation__.Notify(600530)

				return nil, errors.Newf("argument to st_y() must have shape POINT")
			},
			types.Float,
			infoBuilder{
				info: "Returns the Y coordinate of a geometry if it is a Point.",
			},
			tree.VolatilityImmutable,
		),
	),
	"st_z": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(evalContext *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600537)
				t, err := g.Geometry.AsGeomT()
				if err != nil {
					__antithesis_instrumentation__.Notify(600540)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600541)
				}
				__antithesis_instrumentation__.Notify(600538)
				switch t := t.(type) {
				case *geom.Point:
					__antithesis_instrumentation__.Notify(600542)
					if t.Empty() || func() bool {
						__antithesis_instrumentation__.Notify(600544)
						return t.Layout().ZIndex() == -1 == true
					}() == true {
						__antithesis_instrumentation__.Notify(600545)
						return tree.DNull, nil
					} else {
						__antithesis_instrumentation__.Notify(600546)
					}
					__antithesis_instrumentation__.Notify(600543)
					return tree.NewDFloat(tree.DFloat(t.Z())), nil
				}
				__antithesis_instrumentation__.Notify(600539)

				return nil, errors.Newf("argument to st_z() must have shape POINT")
			},
			types.Float,
			infoBuilder{
				info: "Returns the Z coordinate of a geometry if it is a Point.",
			},
			tree.VolatilityImmutable,
		),
	),
	"st_m": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(evalContext *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600547)
				t, err := g.Geometry.AsGeomT()
				if err != nil {
					__antithesis_instrumentation__.Notify(600550)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600551)
				}
				__antithesis_instrumentation__.Notify(600548)
				switch t := t.(type) {
				case *geom.Point:
					__antithesis_instrumentation__.Notify(600552)
					if t.Empty() || func() bool {
						__antithesis_instrumentation__.Notify(600554)
						return t.Layout().MIndex() == -1 == true
					}() == true {
						__antithesis_instrumentation__.Notify(600555)
						return tree.DNull, nil
					} else {
						__antithesis_instrumentation__.Notify(600556)
					}
					__antithesis_instrumentation__.Notify(600553)
					return tree.NewDFloat(tree.DFloat(t.M())), nil
				}
				__antithesis_instrumentation__.Notify(600549)

				return nil, errors.Newf("argument to st_m() must have shape POINT")
			},
			types.Float,
			infoBuilder{
				info: "Returns the M coordinate of a geometry if it is a Point.",
			},
			tree.VolatilityImmutable,
		),
	),
	"st_zmflag": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(evalContext *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600557)
				t, err := g.Geometry.AsGeomT()
				if err != nil {
					__antithesis_instrumentation__.Notify(600559)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600560)
				}
				__antithesis_instrumentation__.Notify(600558)
				switch layout := t.Layout(); layout {
				case geom.XY:
					__antithesis_instrumentation__.Notify(600561)
					return tree.NewDInt(tree.DInt(0)), nil
				case geom.XYM:
					__antithesis_instrumentation__.Notify(600562)
					return tree.NewDInt(tree.DInt(1)), nil
				case geom.XYZ:
					__antithesis_instrumentation__.Notify(600563)
					return tree.NewDInt(tree.DInt(2)), nil
				case geom.XYZM:
					__antithesis_instrumentation__.Notify(600564)
					return tree.NewDInt(tree.DInt(3)), nil
				default:
					__antithesis_instrumentation__.Notify(600565)
					return nil, errors.Newf("unknown geom.Layout %d", layout)
				}
			},
			types.Int2,
			infoBuilder{
				info: "Returns a code based on the ZM coordinate dimension of a geometry (XY = 0, XYM = 1, XYZ = 2, XYZM = 3).",
			},
			tree.VolatilityImmutable,
		),
	),
	"st_area": makeBuiltin(
		defProps(),
		append(
			geographyOverload1WithUseSpheroid(
				func(ctx *tree.EvalContext, g *tree.DGeography, useSphereOrSpheroid geogfn.UseSphereOrSpheroid) (tree.Datum, error) {
					__antithesis_instrumentation__.Notify(600566)
					ret, err := geogfn.Area(g.Geography, useSphereOrSpheroid)
					if err != nil {
						__antithesis_instrumentation__.Notify(600568)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(600569)
					}
					__antithesis_instrumentation__.Notify(600567)
					return tree.NewDFloat(tree.DFloat(ret)), nil
				},
				types.Float,
				infoBuilder{
					info: "Returns the area of the given geography in meters^2.",
				},
				tree.VolatilityImmutable,
			),
			areaOverloadGeometry1,
		)...,
	),
	"st_area2d": makeBuiltin(
		defProps(),
		areaOverloadGeometry1,
	),
	"st_length": makeBuiltin(
		defProps(),
		append(
			geographyOverload1WithUseSpheroid(
				func(ctx *tree.EvalContext, g *tree.DGeography, useSphereOrSpheroid geogfn.UseSphereOrSpheroid) (tree.Datum, error) {
					__antithesis_instrumentation__.Notify(600570)
					ret, err := geogfn.Length(g.Geography, useSphereOrSpheroid)
					if err != nil {
						__antithesis_instrumentation__.Notify(600572)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(600573)
					}
					__antithesis_instrumentation__.Notify(600571)
					return tree.NewDFloat(tree.DFloat(ret)), nil
				},
				types.Float,
				infoBuilder{
					info: "Returns the length of the given geography in meters.",
				},
				tree.VolatilityImmutable,
			),
			lengthOverloadGeometry1,
		)...,
	),
	"st_length2d": makeBuiltin(
		defProps(),
		lengthOverloadGeometry1,
	),
	"st_perimeter": makeBuiltin(
		defProps(),
		append(
			geographyOverload1WithUseSpheroid(
				func(ctx *tree.EvalContext, g *tree.DGeography, useSphereOrSpheroid geogfn.UseSphereOrSpheroid) (tree.Datum, error) {
					__antithesis_instrumentation__.Notify(600574)
					ret, err := geogfn.Perimeter(g.Geography, useSphereOrSpheroid)
					if err != nil {
						__antithesis_instrumentation__.Notify(600576)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(600577)
					}
					__antithesis_instrumentation__.Notify(600575)
					return tree.NewDFloat(tree.DFloat(ret)), nil
				},
				types.Float,
				infoBuilder{
					info: "Returns the perimeter of the given geography in meters.",
				},
				tree.VolatilityImmutable,
			),
			perimeterOverloadGeometry1,
		)...,
	),
	"st_perimeter2d": makeBuiltin(
		defProps(),
		perimeterOverloadGeometry1,
	),
	"st_srid": makeBuiltin(
		defProps(),
		geographyOverload1(
			func(_ *tree.EvalContext, g *tree.DGeography) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600578)
				return tree.NewDInt(tree.DInt(g.SRID())), nil
			},
			types.Int,
			infoBuilder{
				info: "Returns the Spatial Reference Identifier (SRID) for the ST_Geography as defined in spatial_ref_sys table.",
			},
			tree.VolatilityImmutable,
		),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600579)
				return tree.NewDInt(tree.DInt(g.SRID())), nil
			},
			types.Int,
			infoBuilder{
				info: "Returns the Spatial Reference Identifier (SRID) for the ST_Geometry as defined in spatial_ref_sys table.",
			},
			tree.VolatilityImmutable,
		),
	),
	"geometrytype": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600580)
				return tree.NewDString(g.ShapeType2D().String()), nil
			},
			types.String,
			infoBuilder{
				info:         "Returns the type of geometry as a string.",
				libraryUsage: usesGEOS,
			},
			tree.VolatilityImmutable,
		),
	),
	"st_geometrytype": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600581)
				return tree.NewDString(fmt.Sprintf("ST_%s", g.ShapeType2D().String())), nil
			},
			types.String,
			infoBuilder{
				info:         "Returns the type of geometry as a string prefixed with `ST_`.",
				libraryUsage: usesGEOS,
			},
			tree.VolatilityImmutable,
		),
	),
	"st_addmeasure": makeBuiltin(
		defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"geometry", types.Geometry}, {"start", types.Float}, {"end", types.Float}},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600582)
				g := tree.MustBeDGeometry(args[0])
				start := tree.MustBeDFloat(args[1])
				end := tree.MustBeDFloat(args[2])

				ret, err := geomfn.AddMeasure(g.Geometry, float64(start), float64(end))
				if err != nil {
					__antithesis_instrumentation__.Notify(600584)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600585)
				}
				__antithesis_instrumentation__.Notify(600583)
				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{info: "Returns a copy of a LineString or MultiLineString with measure coordinates " +
				"linearly interpolated between the specified start and end values. " +
				"Any existing M coordinates will be overwritten."}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_lineinterpolatepoint": makeBuiltin(
		defProps(),
		lineInterpolatePointForRepeatOverload(
			false,
			`Returns a point along the given LineString which is at given fraction of LineString's total length.`,
		),
	),
	"st_lineinterpolatepoints": makeBuiltin(
		defProps(),
		lineInterpolatePointForRepeatOverload(
			true,
			`Returns one or more points along the LineString which is at an integral multiples of `+
				`given fraction of LineString's total length.

Note If the result has zero or one points, it will be returned as a POINT. If it has two or more points, it will be returned as a MULTIPOINT.`,
		),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"fraction", types.Float},
				{"repeat", types.Bool},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600586)
				g := tree.MustBeDGeometry(args[0])
				fraction := float64(tree.MustBeDFloat(args[1]))
				repeat := bool(tree.MustBeDBool(args[2]))
				interpolatedPoints, err := geomfn.LineInterpolatePoints(g.Geometry, fraction, repeat)
				if err != nil {
					__antithesis_instrumentation__.Notify(600588)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600589)
				}
				__antithesis_instrumentation__.Notify(600587)
				return tree.NewDGeometry(interpolatedPoints), nil
			},
			Info: infoBuilder{
				info: `Returns one or more points along the LineString which is at an integral multiples of given fraction ` +
					`of LineString's total length. If repeat is false (default true) then it returns first point.

Note If the result has zero or one points, it will be returned as a POINT. If it has two or more points, it will be returned as a MULTIPOINT.`,
				libraryUsage: usesGEOS,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_multi": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600590)
				multi, err := geomfn.Multi(g.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600592)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600593)
				}
				__antithesis_instrumentation__.Notify(600591)
				return &tree.DGeometry{Geometry: multi}, nil
			},
			types.Geometry,
			infoBuilder{
				info: `Returns the geometry as a new multi-geometry, e.g converts a POINT to a MULTIPOINT. If the input ` +
					`is already a multitype or collection, it is returned as is.`,
			},
			tree.VolatilityImmutable,
		),
	),
	"st_collectionextract": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"type", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600594)
				g := tree.MustBeDGeometry(args[0])
				shapeType := tree.MustBeDInt(args[1])
				res, err := geomfn.CollectionExtract(g.Geometry, geopb.ShapeType(shapeType))
				if err != nil {
					__antithesis_instrumentation__.Notify(600596)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600597)
				}
				__antithesis_instrumentation__.Notify(600595)
				return &tree.DGeometry{Geometry: res}, nil
			},
			Info: infoBuilder{
				info: `Given a collection, returns a multitype consisting only of elements of the specified type. ` +
					`If there are no elements of the given type, an EMPTY geometry is returned. Types are specified as ` +
					`1=POINT, 2=LINESTRING, 3=POLYGON - other types are not supported.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_collectionhomogenize": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600598)
				ret, err := geomfn.CollectionHomogenize(g.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600600)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600601)
				}
				__antithesis_instrumentation__.Notify(600599)
				return &tree.DGeometry{Geometry: ret}, nil
			},
			types.Geometry,
			infoBuilder{
				info: `Returns the "simplest" representation of a collection's contents. Collections of a single ` +
					`type will be returned as an appopriate multitype, or a singleton if it only contains a ` +
					`single geometry.`,
			},
			tree.VolatilityImmutable,
		),
	),
	"st_forcecollection": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600602)
				ret, err := geomfn.ForceCollection(g.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600604)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600605)
				}
				__antithesis_instrumentation__.Notify(600603)
				return &tree.DGeometry{Geometry: ret}, nil
			},
			types.Geometry,
			infoBuilder{
				info: `Converts the geometry into a GeometryCollection.`,
			},
			tree.VolatilityImmutable,
		),
	),
	"st_linefrommultipoint": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600606)
				line, err := geomfn.LineStringFromMultiPoint(g.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600608)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600609)
				}
				__antithesis_instrumentation__.Notify(600607)
				return &tree.DGeometry{Geometry: line}, nil
			},
			types.Geometry,
			infoBuilder{
				info: `Creates a LineString from a MultiPoint geometry.`,
			},
			tree.VolatilityImmutable,
		),
	),
	"st_linemerge": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600610)
				line, err := geomfn.LineMerge(g.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600612)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600613)
				}
				__antithesis_instrumentation__.Notify(600611)
				return &tree.DGeometry{Geometry: line}, nil
			},
			types.Geometry,
			infoBuilder{
				info: `Returns a LineString or MultiLineString by joining together constituents of a ` +
					`MultiLineString with matching endpoints. If the input is not a MultiLineString or LineString, ` +
					`an empty GeometryCollection is returned.`,
				libraryUsage: usesGEOS,
			},
			tree.VolatilityImmutable,
		),
	),
	"st_shiftlongitude": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600614)
				ret, err := geomfn.ShiftLongitude(g.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600616)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600617)
				}
				__antithesis_instrumentation__.Notify(600615)
				return tree.NewDGeometry(ret), nil
			},
			types.Geometry,
			infoBuilder{
				info: `Returns a modified version of a geometry in which the longitude (X coordinate) of each point is ` +
					`incremented by 360 if it is <0 and decremented by 360 if it is >180. The result is only meaningful ` +
					`if the coordinates are in longitude/latitude.`,
			},
			tree.VolatilityImmutable,
		),
	),

	"st_isclosed": makeBuiltin(
		defProps(),
		geometryOverload1UnaryPredicate(
			geomfn.IsClosed,
			infoBuilder{
				info: `Returns whether the geometry is closed as defined by whether the start and end points are coincident. ` +
					`Points are considered closed, empty geometries are not. For collections and multi-types, all members must be closed, ` +
					`as must all polygon rings.`,
			},
		),
	),
	"st_iscollection": makeBuiltin(
		defProps(),
		geometryOverload1UnaryPredicate(
			geomfn.IsCollection,
			infoBuilder{
				info: "Returns whether the geometry is of a collection type (including multi-types).",
			},
		),
	),
	"st_isempty": makeBuiltin(
		defProps(),
		geometryOverload1UnaryPredicate(
			geomfn.IsEmpty,
			infoBuilder{
				info: "Returns whether the geometry is empty.",
			},
		),
	),
	"st_isring": makeBuiltin(
		defProps(),
		geometryOverload1UnaryPredicate(
			geomfn.IsRing,
			infoBuilder{
				info: `Returns whether the geometry is a single linestring that is closed and simple, as defined by ` +
					`ST_IsClosed and ST_IsSimple.`,
				libraryUsage: usesGEOS,
			},
		),
	),
	"st_issimple": makeBuiltin(
		defProps(),
		geometryOverload1UnaryPredicate(
			geomfn.IsSimple,
			infoBuilder{
				info: `Returns true if the geometry has no anomalous geometric points, e.g. that it intersects with ` +
					`or lies tangent to itself.`,
				libraryUsage: usesGEOS,
			},
		),
	),

	"st_azimuth": makeBuiltin(
		defProps(),
		geometryOverload2(
			func(ctx *tree.EvalContext, a, b *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600618)
				azimuth, err := geomfn.Azimuth(a.Geometry, b.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600621)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600622)
				}
				__antithesis_instrumentation__.Notify(600619)

				if azimuth == nil {
					__antithesis_instrumentation__.Notify(600623)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(600624)
				}
				__antithesis_instrumentation__.Notify(600620)

				return tree.NewDFloat(tree.DFloat(*azimuth)), nil
			},
			types.Float,
			infoBuilder{
				info: `Returns the azimuth in radians of the segment defined by the given point geometries, or NULL if the two points are coincident.

The azimuth is angle is referenced from north, and is positive clockwise: North = 0; East = π/2; South = π; West = 3π/2.`,
			},
			tree.VolatilityImmutable,
		),
		geographyOverload2(
			func(ctx *tree.EvalContext, a, b *tree.DGeography) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600625)
				azimuth, err := geogfn.Azimuth(a.Geography, b.Geography)
				if err != nil {
					__antithesis_instrumentation__.Notify(600628)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600629)
				}
				__antithesis_instrumentation__.Notify(600626)

				if azimuth == nil {
					__antithesis_instrumentation__.Notify(600630)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(600631)
				}
				__antithesis_instrumentation__.Notify(600627)

				return tree.NewDFloat(tree.DFloat(*azimuth)), nil
			},
			types.Float,
			infoBuilder{
				info: `Returns the azimuth in radians of the segment defined by the given point geographies, or NULL if the two points are coincident. It is solved using the Inverse geodesic problem.

The azimuth is angle is referenced from north, and is positive clockwise: North = 0; East = π/2; South = π; West = 3π/2.`,
				libraryUsage: usesGeographicLib,
			},
			tree.VolatilityImmutable,
		),
	),
	"st_distance": makeBuiltin(
		defProps(),
		geometryOverload2(
			func(ctx *tree.EvalContext, a, b *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600632)
				ret, err := geomfn.MinDistance(a.Geometry, b.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600634)
					if geo.IsEmptyGeometryError(err) {
						__antithesis_instrumentation__.Notify(600636)
						return tree.DNull, nil
					} else {
						__antithesis_instrumentation__.Notify(600637)
					}
					__antithesis_instrumentation__.Notify(600635)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600638)
				}
				__antithesis_instrumentation__.Notify(600633)
				return tree.NewDFloat(tree.DFloat(ret)), nil
			},
			types.Float,
			infoBuilder{
				info: `Returns the distance between the given geometries.`,
			},
			tree.VolatilityImmutable,
		),
		geographyOverload2(
			func(ctx *tree.EvalContext, a *tree.DGeography, b *tree.DGeography) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600639)
				ret, err := geogfn.Distance(a.Geography, b.Geography, geogfn.UseSpheroid)
				if err != nil {
					__antithesis_instrumentation__.Notify(600641)
					if geo.IsEmptyGeometryError(err) {
						__antithesis_instrumentation__.Notify(600643)
						return tree.DNull, nil
					} else {
						__antithesis_instrumentation__.Notify(600644)
					}
					__antithesis_instrumentation__.Notify(600642)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600645)
				}
				__antithesis_instrumentation__.Notify(600640)
				return tree.NewDFloat(tree.DFloat(ret)), nil
			},
			types.Float,
			infoBuilder{
				info:         "Returns the distance in meters between geography_a and geography_b. " + usesSpheroidMessage + spheroidDistanceMessage,
				libraryUsage: usesGeographicLib,
			},
			tree.VolatilityImmutable,
		),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geography_a", types.Geography},
				{"geography_b", types.Geography},
				{"use_spheroid", types.Bool},
			},
			ReturnType: tree.FixedReturnType(types.Float),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600646)
				a := tree.MustBeDGeography(args[0])
				b := tree.MustBeDGeography(args[1])
				useSpheroid := tree.MustBeDBool(args[2])

				ret, err := geogfn.Distance(a.Geography, b.Geography, toUseSphereOrSpheroid(useSpheroid))
				if err != nil {
					__antithesis_instrumentation__.Notify(600648)
					if geo.IsEmptyGeometryError(err) {
						__antithesis_instrumentation__.Notify(600650)
						return tree.DNull, nil
					} else {
						__antithesis_instrumentation__.Notify(600651)
					}
					__antithesis_instrumentation__.Notify(600649)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600652)
				}
				__antithesis_instrumentation__.Notify(600647)
				return tree.NewDFloat(tree.DFloat(ret)), nil
			},
			Info: infoBuilder{
				info:         "Returns the distance in meters between geography_a and geography_b." + spheroidDistanceMessage,
				libraryUsage: usesGeographicLib | usesS2,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_distancesphere": makeBuiltin(
		defProps(),
		geometryOverload2(
			func(ctx *tree.EvalContext, a, b *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600653)
				aGeog, err := a.Geometry.AsGeography()
				if err != nil {
					__antithesis_instrumentation__.Notify(600657)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600658)
				}
				__antithesis_instrumentation__.Notify(600654)
				bGeog, err := b.Geometry.AsGeography()
				if err != nil {
					__antithesis_instrumentation__.Notify(600659)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600660)
				}
				__antithesis_instrumentation__.Notify(600655)
				ret, err := geogfn.Distance(aGeog, bGeog, geogfn.UseSphere)
				if err != nil {
					__antithesis_instrumentation__.Notify(600661)
					if geo.IsEmptyGeometryError(err) {
						__antithesis_instrumentation__.Notify(600663)
						return tree.DNull, nil
					} else {
						__antithesis_instrumentation__.Notify(600664)
					}
					__antithesis_instrumentation__.Notify(600662)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600665)
				}
				__antithesis_instrumentation__.Notify(600656)
				return tree.NewDFloat(tree.DFloat(ret)), nil
			},
			types.Float,
			infoBuilder{
				info: "Returns the distance in meters between geometry_a and geometry_b assuming the coordinates " +
					"represent lng/lat points on a sphere.",
				libraryUsage: usesS2,
			},
			tree.VolatilityImmutable,
		),
	),
	"st_distancespheroid": makeBuiltin(
		defProps(),
		geometryOverload2(
			func(ctx *tree.EvalContext, a, b *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600666)
				aGeog, err := a.Geometry.AsGeography()
				if err != nil {
					__antithesis_instrumentation__.Notify(600670)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600671)
				}
				__antithesis_instrumentation__.Notify(600667)
				bGeog, err := b.Geometry.AsGeography()
				if err != nil {
					__antithesis_instrumentation__.Notify(600672)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600673)
				}
				__antithesis_instrumentation__.Notify(600668)
				ret, err := geogfn.Distance(aGeog, bGeog, geogfn.UseSpheroid)
				if err != nil {
					__antithesis_instrumentation__.Notify(600674)
					if geo.IsEmptyGeometryError(err) {
						__antithesis_instrumentation__.Notify(600676)
						return tree.DNull, nil
					} else {
						__antithesis_instrumentation__.Notify(600677)
					}
					__antithesis_instrumentation__.Notify(600675)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600678)
				}
				__antithesis_instrumentation__.Notify(600669)
				return tree.NewDFloat(tree.DFloat(ret)), nil
			},
			types.Float,
			infoBuilder{
				info: "Returns the distance in meters between geometry_a and geometry_b assuming the coordinates " +
					"represent lng/lat points on a spheroid." + spheroidDistanceMessage,
				libraryUsage: usesGeographicLib | usesS2,
			},
			tree.VolatilityImmutable,
		),
	),
	"st_frechetdistance": makeBuiltin(
		defProps(),
		geometryOverload2(
			func(ctx *tree.EvalContext, a, b *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600679)
				ret, err := geomfn.FrechetDistance(a.Geometry, b.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600682)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600683)
				}
				__antithesis_instrumentation__.Notify(600680)
				if ret == nil {
					__antithesis_instrumentation__.Notify(600684)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(600685)
				}
				__antithesis_instrumentation__.Notify(600681)
				return tree.NewDFloat(tree.DFloat(*ret)), nil
			},
			types.Float,
			infoBuilder{
				info:         `Returns the Frechet distance between the given geometries.`,
				libraryUsage: usesGEOS,
			},
			tree.VolatilityImmutable,
		),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry_a", types.Geometry},
				{"geometry_b", types.Geometry},
				{"densify_frac", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Float),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600686)
				a := tree.MustBeDGeometry(args[0])
				b := tree.MustBeDGeometry(args[1])
				densifyFrac := tree.MustBeDFloat(args[2])

				ret, err := geomfn.FrechetDistanceDensify(a.Geometry, b.Geometry, float64(densifyFrac))
				if err != nil {
					__antithesis_instrumentation__.Notify(600689)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600690)
				}
				__antithesis_instrumentation__.Notify(600687)
				if ret == nil {
					__antithesis_instrumentation__.Notify(600691)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(600692)
				}
				__antithesis_instrumentation__.Notify(600688)
				return tree.NewDFloat(tree.DFloat(*ret)), nil
			},
			Info: infoBuilder{
				info: "Returns the Frechet distance between the given geometries, with the given " +
					"segment densification (range 0.0-1.0, -1 to disable).\n\n" +
					"Smaller densify_frac gives a more accurate Fréchet distance. However, the computation " +
					"time and memory usage increases with the square of the number of subsegments.",
				libraryUsage: usesGEOS,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_hausdorffdistance": makeBuiltin(
		defProps(),
		geometryOverload2(
			func(ctx *tree.EvalContext, a, b *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600693)
				ret, err := geomfn.HausdorffDistance(a.Geometry, b.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600696)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600697)
				}
				__antithesis_instrumentation__.Notify(600694)
				if ret == nil {
					__antithesis_instrumentation__.Notify(600698)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(600699)
				}
				__antithesis_instrumentation__.Notify(600695)
				return tree.NewDFloat(tree.DFloat(*ret)), nil
			},
			types.Float,
			infoBuilder{
				info:         `Returns the Hausdorff distance between the given geometries.`,
				libraryUsage: usesGEOS,
			},
			tree.VolatilityImmutable,
		),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry_a", types.Geometry},
				{"geometry_b", types.Geometry},
				{"densify_frac", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Float),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600700)
				a := tree.MustBeDGeometry(args[0])
				b := tree.MustBeDGeometry(args[1])
				densifyFrac := tree.MustBeDFloat(args[2])

				ret, err := geomfn.HausdorffDistanceDensify(a.Geometry, b.Geometry, float64(densifyFrac))
				if err != nil {
					__antithesis_instrumentation__.Notify(600703)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600704)
				}
				__antithesis_instrumentation__.Notify(600701)
				if ret == nil {
					__antithesis_instrumentation__.Notify(600705)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(600706)
				}
				__antithesis_instrumentation__.Notify(600702)
				return tree.NewDFloat(tree.DFloat(*ret)), nil
			},
			Info: infoBuilder{
				info: `Returns the Hausdorff distance between the given geometries, with the given ` +
					`segment densification (range 0.0-1.0).`,
				libraryUsage: usesGEOS,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_maxdistance": makeBuiltin(
		defProps(),
		geometryOverload2(
			func(ctx *tree.EvalContext, a, b *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600707)
				ret, err := geomfn.MaxDistance(a.Geometry, b.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600709)
					if geo.IsEmptyGeometryError(err) {
						__antithesis_instrumentation__.Notify(600711)
						return tree.DNull, nil
					} else {
						__antithesis_instrumentation__.Notify(600712)
					}
					__antithesis_instrumentation__.Notify(600710)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600713)
				}
				__antithesis_instrumentation__.Notify(600708)
				return tree.NewDFloat(tree.DFloat(ret)), nil
			},
			types.Float,
			infoBuilder{
				info: `Returns the maximum distance across every pair of points comprising the given geometries. ` +
					`Note if the geometries are the same, it will return the maximum distance between the geometry's vertexes.`,
			},
			tree.VolatilityImmutable,
		),
	),
	"st_longestline": makeBuiltin(
		defProps(),
		geometryOverload2(
			func(ctx *tree.EvalContext, a, b *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600714)
				longestLineString, err := geomfn.LongestLineString(a.Geometry, b.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600716)
					if geo.IsEmptyGeometryError(err) {
						__antithesis_instrumentation__.Notify(600718)
						return tree.DNull, nil
					} else {
						__antithesis_instrumentation__.Notify(600719)
					}
					__antithesis_instrumentation__.Notify(600717)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600720)
				}
				__antithesis_instrumentation__.Notify(600715)
				return tree.NewDGeometry(longestLineString), nil
			},
			types.Geometry,
			infoBuilder{
				info: `Returns the LineString corresponds to the max distance across every pair of points comprising the ` +
					`given geometries.

Note if geometries are the same, it will return the LineString with the maximum distance between the geometry's ` +
					`vertexes. The function will return the longest line that was discovered first when comparing maximum ` +
					`distances if more than one is found.`,
			},
			tree.VolatilityImmutable,
		),
	),
	"st_shortestline": makeBuiltin(
		defProps(),
		geometryOverload2(
			func(ctx *tree.EvalContext, a, b *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600721)
				shortestLineString, err := geomfn.ShortestLineString(a.Geometry, b.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600723)
					if geo.IsEmptyGeometryError(err) {
						__antithesis_instrumentation__.Notify(600725)
						return tree.DNull, nil
					} else {
						__antithesis_instrumentation__.Notify(600726)
					}
					__antithesis_instrumentation__.Notify(600724)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600727)
				}
				__antithesis_instrumentation__.Notify(600722)
				return tree.NewDGeometry(shortestLineString), nil
			},
			types.Geometry,
			infoBuilder{
				info: `Returns the LineString corresponds to the minimum distance across every pair of points comprising ` +
					`the given geometries.

Note if geometries are the same, it will return the LineString with the minimum distance between the geometry's ` +
					`vertexes. The function will return the shortest line that was discovered first when comparing minimum ` +
					`distances if more than one is found.`,
			},
			tree.VolatilityImmutable,
		),
	),

	"st_covers": makeBuiltin(
		defProps(),
		geometryOverload2BinaryPredicate(
			geomfn.Covers,
			infoBuilder{
				info:         "Returns true if no point in geometry_b is outside geometry_a.",
				libraryUsage: usesGEOS,
			},
		),
		geographyOverload2BinaryPredicate(
			geogfn.Covers,
			infoBuilder{
				info:         `Returns true if no point in geography_b is outside geography_a.`,
				libraryUsage: usesS2,
			},
		),
	),
	"st_coveredby": makeBuiltin(
		defProps(),
		geometryOverload2BinaryPredicate(
			geomfn.CoveredBy,
			infoBuilder{
				info:         `Returns true if no point in geometry_a is outside geometry_b.`,
				libraryUsage: usesGEOS,
			},
		),
		geographyOverload2BinaryPredicate(
			geogfn.CoveredBy,
			infoBuilder{
				info:         `Returns true if no point in geography_a is outside geography_b.`,
				libraryUsage: usesS2,
				precision:    "1cm",
			},
		),
	),
	"st_contains": makeBuiltin(
		defProps(),
		geometryOverload2BinaryPredicate(
			geomfn.Contains,
			infoBuilder{
				info: "Returns true if no points of geometry_b lie in the exterior of geometry_a, " +
					"and there is at least one point in the interior of geometry_b that lies in the interior of geometry_a.",
				libraryUsage: usesGEOS,
			},
		),
	),
	"st_containsproperly": makeBuiltin(
		defProps(),
		geometryOverload2BinaryPredicate(
			geomfn.ContainsProperly,
			infoBuilder{
				info:         "Returns true if geometry_b intersects the interior of geometry_a but not the boundary or exterior of geometry_a.",
				libraryUsage: usesGEOS,
			},
		),
	),
	"st_crosses": makeBuiltin(
		defProps(),
		geometryOverload2BinaryPredicate(
			geomfn.Crosses,
			infoBuilder{
				info:         "Returns true if geometry_a has some - but not all - interior points in common with geometry_b.",
				libraryUsage: usesGEOS,
			},
		),
	),
	"st_disjoint": makeBuiltin(
		defProps(),
		geometryOverload2BinaryPredicate(
			geomfn.Disjoint,
			infoBuilder{
				info:         "Returns true if geometry_a does not overlap, touch or is within geometry_b.",
				libraryUsage: usesGEOS,
			},
		),
	),
	"st_dfullywithin": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry_a", types.Geometry},
				{"geometry_b", types.Geometry},
				{"distance", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600728)
				a := tree.MustBeDGeometry(args[0])
				b := tree.MustBeDGeometry(args[1])
				dist := tree.MustBeDFloat(args[2])
				ret, err := geomfn.DFullyWithin(a.Geometry, b.Geometry, float64(dist), geo.FnInclusive)
				if err != nil {
					__antithesis_instrumentation__.Notify(600730)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600731)
				}
				__antithesis_instrumentation__.Notify(600729)
				return tree.MakeDBool(tree.DBool(ret)), nil
			},
			Info: infoBuilder{
				info: "Returns true if every pair of points comprising geometry_a and geometry_b are within distance units, inclusive. " +
					"In other words, the ST_MaxDistance between geometry_a and geometry_b is less than or equal to distance units.",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_dwithin": makeSTDWithinBuiltin(geo.FnInclusive),
	"st_equals": makeBuiltin(
		defProps(),
		geometryOverload2BinaryPredicate(
			geomfn.Equals,
			infoBuilder{
				info: "Returns true if geometry_a is spatially equal to geometry_b, " +
					"i.e. ST_Within(geometry_a, geometry_b) = ST_Within(geometry_b, geometry_a) = true.",
				libraryUsage: usesGEOS,
			},
		),
	),
	"st_orderingequals": makeBuiltin(
		defProps(),
		geometryOverload2BinaryPredicate(
			geomfn.OrderingEquals,
			infoBuilder{
				info: "Returns true if geometry_a is exactly equal to geometry_b, having all coordinates " +
					"in the same order, as well as the same type, SRID, bounding box, and so on.",
			},
		),
	),
	"st_normalize": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600732)
				ret, err := geomfn.Normalize(g.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600734)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600735)
				}
				__antithesis_instrumentation__.Notify(600733)
				return tree.NewDGeometry(ret), nil
			},
			types.Geometry,
			infoBuilder{
				info:         "Returns the geometry in its normalized form.",
				libraryUsage: usesGEOS,
			},
			tree.VolatilityImmutable,
		),
	),
	"st_intersects": makeBuiltin(
		defProps(),
		geometryOverload2BinaryPredicate(
			geomfn.Intersects,
			infoBuilder{
				info:         "Returns true if geometry_a shares any portion of space with geometry_b.",
				libraryUsage: usesGEOS,
				precision:    "1cm",
			},
		),
		geographyOverload2BinaryPredicate(
			geogfn.Intersects,
			infoBuilder{
				info:         `Returns true if geography_a shares any portion of space with geography_b.`,
				libraryUsage: usesS2,
				precision:    "1cm",
			},
		),
	),
	"st_overlaps": makeBuiltin(
		defProps(),
		geometryOverload2BinaryPredicate(
			geomfn.Overlaps,
			infoBuilder{
				info: "Returns true if geometry_a intersects but does not completely contain geometry_b, or vice versa. " +
					`"Does not completely" implies ST_Within(geometry_a, geometry_b) = ST_Within(geometry_b, geometry_a) = false.`,
				libraryUsage: usesGEOS,
			},
		),
	),
	"st_touches": makeBuiltin(
		defProps(),
		geometryOverload2BinaryPredicate(
			geomfn.Touches,
			infoBuilder{
				info: "Returns true if the only points in common between geometry_a and geometry_b are on the boundary. " +
					"Note points do not touch other points.",
				libraryUsage: usesGEOS,
			},
		),
	),
	"st_within": makeBuiltin(
		defProps(),
		geometryOverload2BinaryPredicate(
			geomfn.Within,
			infoBuilder{
				info:         "Returns true if geometry_a is completely inside geometry_b.",
				libraryUsage: usesGEOS,
			},
		),
	),

	"st_relate": makeBuiltin(
		defProps(),
		geometryOverload2(
			func(ctx *tree.EvalContext, a *tree.DGeometry, b *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600736)
				ret, err := geomfn.Relate(a.Geometry, b.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600738)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600739)
				}
				__antithesis_instrumentation__.Notify(600737)
				return tree.NewDString(ret), nil
			},
			types.String,
			infoBuilder{
				info:         `Returns the DE-9IM spatial relation between geometry_a and geometry_b.`,
				libraryUsage: usesGEOS,
			},
			tree.VolatilityImmutable,
		),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry_a", types.Geometry},
				{"geometry_b", types.Geometry},
				{"pattern", types.String},
			},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600740)
				a := tree.MustBeDGeometry(args[0])
				b := tree.MustBeDGeometry(args[1])
				pattern := tree.MustBeDString(args[2])
				ret, err := geomfn.RelatePattern(a.Geometry, b.Geometry, string(pattern))
				if err != nil {
					__antithesis_instrumentation__.Notify(600742)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600743)
				}
				__antithesis_instrumentation__.Notify(600741)
				return tree.MakeDBool(tree.DBool(ret)), nil
			},
			Info: infoBuilder{
				info:         `Returns whether the DE-9IM spatial relation between geometry_a and geometry_b matches the DE-9IM pattern.`,
				libraryUsage: usesGEOS,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry_a", types.Geometry},
				{"geometry_b", types.Geometry},
				{"bnr", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600744)
				a := tree.MustBeDGeometry(args[0])
				b := tree.MustBeDGeometry(args[1])
				bnr := tree.MustBeDInt(args[2])
				ret, err := geomfn.RelateBoundaryNodeRule(a.Geometry, b.Geometry, int(bnr))
				if err != nil {
					__antithesis_instrumentation__.Notify(600746)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600747)
				}
				__antithesis_instrumentation__.Notify(600745)
				return tree.NewDString(ret), nil
			},
			Info: infoBuilder{
				info: `Returns the DE-9IM spatial relation between geometry_a and geometry_b using the given ` +
					`boundary node rule (1:OGC/MOD2, 2:Endpoint, 3:MultivalentEndpoint, 4:MonovalentEndpoint).`,
				libraryUsage: usesGEOS,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_relatematch": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"intersection_matrix", types.String},
				{"pattern", types.String},
			},
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600748)
				matrix := string(tree.MustBeDString(args[0]))
				pattern := string(tree.MustBeDString(args[1]))

				matches, err := geomfn.MatchesDE9IM(matrix, pattern)
				if err != nil {
					__antithesis_instrumentation__.Notify(600750)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600751)
				}
				__antithesis_instrumentation__.Notify(600749)
				return tree.MakeDBool(tree.DBool(matches)), nil
			},
			ReturnType: tree.FixedReturnType(types.Bool),
			Info: infoBuilder{
				info: "Returns whether the given DE-9IM intersection matrix satisfies the given pattern.",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),

	"st_isvalid": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600752)
				ret, err := geomfn.IsValid(g.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600754)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600755)
				}
				__antithesis_instrumentation__.Notify(600753)
				return tree.MakeDBool(tree.DBool(ret)), nil
			},
			types.Bool,
			infoBuilder{
				info:         `Returns whether the geometry is valid as defined by the OGC spec.`,
				libraryUsage: usesGEOS,
			},
			tree.VolatilityImmutable,
		),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"flags", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600756)
				g := tree.MustBeDGeometry(args[0])
				flags := int(tree.MustBeDInt(args[1]))
				validDetail, err := geomfn.IsValidDetail(g.Geometry, flags)
				if err != nil {
					__antithesis_instrumentation__.Notify(600758)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600759)
				}
				__antithesis_instrumentation__.Notify(600757)
				return tree.MakeDBool(tree.DBool(validDetail.IsValid)), nil
			},
			Info: infoBuilder{
				info: `Returns whether the geometry is valid.

For flags=0, validity is defined by the OGC spec.

For flags=1, validity considers self-intersecting rings forming holes as valid as per ESRI. This is not valid under OGC and CRDB spatial operations may not operate correctly.`,
				libraryUsage: usesGEOS,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_isvalidreason": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600760)
				ret, err := geomfn.IsValidReason(g.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600762)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600763)
				}
				__antithesis_instrumentation__.Notify(600761)
				return tree.NewDString(ret), nil
			},
			types.String,
			infoBuilder{
				info:         `Returns a string containing the reason the geometry is invalid along with the point of interest, or "Valid Geometry" if it is valid. Validity is defined by the OGC spec.`,
				libraryUsage: usesGEOS,
			},
			tree.VolatilityImmutable,
		),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"flags", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600764)
				g := tree.MustBeDGeometry(args[0])
				flags := int(tree.MustBeDInt(args[1]))
				validDetail, err := geomfn.IsValidDetail(g.Geometry, flags)
				if err != nil {
					__antithesis_instrumentation__.Notify(600767)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600768)
				}
				__antithesis_instrumentation__.Notify(600765)
				if validDetail.IsValid {
					__antithesis_instrumentation__.Notify(600769)
					return tree.NewDString("Valid Geometry"), nil
				} else {
					__antithesis_instrumentation__.Notify(600770)
				}
				__antithesis_instrumentation__.Notify(600766)
				return tree.NewDString(validDetail.Reason), nil
			},
			Info: infoBuilder{
				info: `Returns the reason the geometry is invalid or "Valid Geometry" if it is valid.

For flags=0, validity is defined by the OGC spec.

For flags=1, validity considers self-intersecting rings forming holes as valid as per ESRI. This is not valid under OGC and CRDB spatial operations may not operate correctly.`,
				libraryUsage: usesGEOS,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_isvalidtrajectory": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600771)
				ret, err := geomfn.IsValidTrajectory(g.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600773)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600774)
				}
				__antithesis_instrumentation__.Notify(600772)
				return tree.MakeDBool(tree.DBool(ret)), nil
			},
			types.Bool,
			infoBuilder{
				info: `Returns whether the geometry encodes a valid trajectory.

Note the geometry must be a LineString with M coordinates.`,
			},
			tree.VolatilityImmutable,
		),
	),
	"st_makevalid": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600775)
				validGeom, err := geomfn.MakeValid(g.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600777)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600778)
				}
				__antithesis_instrumentation__.Notify(600776)
				return tree.NewDGeometry(validGeom), err
			},
			types.Geometry,
			infoBuilder{
				info:         "Returns a valid form of the given geometry according to the OGC spec.",
				libraryUsage: usesGEOS,
			},
			tree.VolatilityImmutable,
		),
	),

	"st_boundary": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600779)
				centroid, err := geomfn.Boundary(g.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600781)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600782)
				}
				__antithesis_instrumentation__.Notify(600780)
				return tree.NewDGeometry(centroid), err
			},
			types.Geometry,
			infoBuilder{
				info:         "Returns the closure of the combinatorial boundary of this Geometry.",
				libraryUsage: usesGEOS,
			},
			tree.VolatilityImmutable,
		),
	),
	"st_centroid": makeBuiltin(
		defProps(),
		append(
			geographyOverload1WithUseSpheroid(
				func(ctx *tree.EvalContext, g *tree.DGeography, useSphereOrSpheroid geogfn.UseSphereOrSpheroid) (tree.Datum, error) {
					__antithesis_instrumentation__.Notify(600783)
					ret, err := geogfn.Centroid(g.Geography, useSphereOrSpheroid)
					if err != nil {
						__antithesis_instrumentation__.Notify(600785)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(600786)
					}
					__antithesis_instrumentation__.Notify(600784)
					return tree.NewDGeography(ret), nil
				},
				types.Geography,
				infoBuilder{
					info: "Returns the centroid of given geography.",
				},
				tree.VolatilityImmutable,
			),
			geometryOverload1(
				func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
					__antithesis_instrumentation__.Notify(600787)
					centroid, err := geomfn.Centroid(g.Geometry)
					if err != nil {
						__antithesis_instrumentation__.Notify(600789)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(600790)
					}
					__antithesis_instrumentation__.Notify(600788)
					return tree.NewDGeometry(centroid), err
				},
				types.Geometry,
				infoBuilder{
					info:         "Returns the centroid of the given geometry.",
					libraryUsage: usesGEOS,
				},
				tree.VolatilityImmutable,
			),
		)...,
	),
	"st_clipbybox2d": makeBuiltin(
		defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"geometry", types.Geometry}, {"box2d", types.Box2D}},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600791)
				g := tree.MustBeDGeometry(args[0])
				bbox := tree.MustBeDBox2D(args[1])
				ret, err := geomfn.ClipByRect(g.Geometry, bbox.CartesianBoundingBox)
				if err != nil {
					__antithesis_instrumentation__.Notify(600793)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600794)
				}
				__antithesis_instrumentation__.Notify(600792)
				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: "Clips the geometry to conform to the bounding box specified by box2d.",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_convexhull": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600795)
				convexHull, err := geomfn.ConvexHull(g.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600797)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600798)
				}
				__antithesis_instrumentation__.Notify(600796)
				return tree.NewDGeometry(convexHull), err
			},
			types.Geometry,
			infoBuilder{
				info:         "Returns a geometry that represents the Convex Hull of the given geometry.",
				libraryUsage: usesGEOS,
			},
			tree.VolatilityImmutable,
		),
	),
	"st_difference": makeBuiltin(
		defProps(),
		geometryOverload2(
			func(ctx *tree.EvalContext, a, b *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600799)
				diff, err := geomfn.Difference(a.Geometry, b.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600801)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600802)
				}
				__antithesis_instrumentation__.Notify(600800)
				return tree.NewDGeometry(diff), err
			},
			types.Geometry,
			infoBuilder{
				info:         "Returns the difference of two Geometries.",
				libraryUsage: usesGEOS,
			},
			tree.VolatilityImmutable,
		),
	),
	"st_pointonsurface": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600803)
				pointOnSurface, err := geomfn.PointOnSurface(g.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600805)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600806)
				}
				__antithesis_instrumentation__.Notify(600804)
				return tree.NewDGeometry(pointOnSurface), err
			},
			types.Geometry,
			infoBuilder{
				info:         "Returns a point that intersects with the given Geometry.",
				libraryUsage: usesGEOS,
			},
			tree.VolatilityImmutable,
		),
	),
	"st_intersection": makeBuiltin(
		defProps(),
		geometryOverload2(
			func(ctx *tree.EvalContext, a *tree.DGeometry, b *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600807)
				intersection, err := geomfn.Intersection(a.Geometry, b.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600809)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600810)
				}
				__antithesis_instrumentation__.Notify(600808)
				return tree.NewDGeometry(intersection), err
			},
			types.Geometry,
			infoBuilder{
				info:         "Returns the point intersections of the given geometries.",
				libraryUsage: usesGEOS,
			},
			tree.VolatilityImmutable,
		),
		geographyOverload2(
			func(ctx *tree.EvalContext, a *tree.DGeography, b *tree.DGeography) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600811)
				proj, err := geogfn.BestGeomProjection(a.Geography.BoundingRect().Union(b.Geography.BoundingRect()))
				if err != nil {
					__antithesis_instrumentation__.Notify(600823)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600824)
				}
				__antithesis_instrumentation__.Notify(600812)
				aProj, err := geoprojbase.Projection(a.Geography.SRID())
				if err != nil {
					__antithesis_instrumentation__.Notify(600825)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600826)
				}
				__antithesis_instrumentation__.Notify(600813)
				bProj, err := geoprojbase.Projection(b.Geography.SRID())
				if err != nil {
					__antithesis_instrumentation__.Notify(600827)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600828)
				}
				__antithesis_instrumentation__.Notify(600814)
				geogDefaultProj, err := geoprojbase.Projection(geopb.DefaultGeographySRID)
				if err != nil {
					__antithesis_instrumentation__.Notify(600829)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600830)
				}
				__antithesis_instrumentation__.Notify(600815)

				aInGeom, err := a.Geography.AsGeometry()
				if err != nil {
					__antithesis_instrumentation__.Notify(600831)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600832)
				}
				__antithesis_instrumentation__.Notify(600816)
				bInGeom, err := b.Geography.AsGeometry()
				if err != nil {
					__antithesis_instrumentation__.Notify(600833)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600834)
				}
				__antithesis_instrumentation__.Notify(600817)

				aInProjected, err := geotransform.Transform(
					aInGeom,
					aProj.Proj4Text,
					proj,
					a.Geography.SRID(),
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(600835)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600836)
				}
				__antithesis_instrumentation__.Notify(600818)
				bInProjected, err := geotransform.Transform(
					bInGeom,
					bProj.Proj4Text,
					proj,
					b.Geography.SRID(),
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(600837)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600838)
				}
				__antithesis_instrumentation__.Notify(600819)

				projectedIntersection, err := geomfn.Intersection(aInProjected, bInProjected)
				if err != nil {
					__antithesis_instrumentation__.Notify(600839)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600840)
				}
				__antithesis_instrumentation__.Notify(600820)

				outGeom, err := geotransform.Transform(
					projectedIntersection,
					proj,
					geogDefaultProj.Proj4Text,
					geopb.DefaultGeographySRID,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(600841)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600842)
				}
				__antithesis_instrumentation__.Notify(600821)
				ret, err := outGeom.AsGeography()
				if err != nil {
					__antithesis_instrumentation__.Notify(600843)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600844)
				}
				__antithesis_instrumentation__.Notify(600822)
				return tree.NewDGeography(ret), nil
			},
			types.Geography,
			infoBuilder{
				info:         "Returns the point intersections of the given geographies." + usingBestGeomProjectionWarning,
				libraryUsage: usesGEOS,
			},
			tree.VolatilityImmutable,
		),
	),
	"st_sharedpaths": makeBuiltin(
		defProps(),
		geometryOverload2(
			func(ctx *tree.EvalContext, a, b *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600845)
				ret, err := geomfn.SharedPaths(a.Geometry, b.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600847)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600848)
				}
				__antithesis_instrumentation__.Notify(600846)
				geom := tree.NewDGeometry(ret)
				return geom, nil
			},
			types.Geometry,
			infoBuilder{
				info: `Returns a collection containing paths shared by the two input geometries.

Those going in the same direction are in the first element of the collection,
those going in the opposite direction are in the second element.
The paths themselves are given in the direction of the first geometry.`,
			},
			tree.VolatilityImmutable,
		),
	),
	"st_closestpoint": makeBuiltin(
		defProps(),
		geometryOverload2(
			func(ctx *tree.EvalContext, a, b *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600849)
				ret, err := geomfn.ClosestPoint(a.Geometry, b.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600851)
					if geo.IsEmptyGeometryError(err) {
						__antithesis_instrumentation__.Notify(600853)
						return tree.DNull, nil
					} else {
						__antithesis_instrumentation__.Notify(600854)
					}
					__antithesis_instrumentation__.Notify(600852)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600855)
				}
				__antithesis_instrumentation__.Notify(600850)
				geom := tree.NewDGeometry(ret)
				return geom, nil
			},
			types.Geometry,
			infoBuilder{
				info: `Returns the 2-dimensional point on geometry_a that is closest to geometry_b. This is the first point of the shortest line.`,
			},
			tree.VolatilityImmutable,
		),
	),
	"st_symdifference": makeBuiltin(
		defProps(),
		geometryOverload2(
			func(ctx *tree.EvalContext, a, b *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600856)
				ret, err := geomfn.SymDifference(a.Geometry, b.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600858)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600859)
				}
				__antithesis_instrumentation__.Notify(600857)
				return tree.NewDGeometry(ret), nil
			},
			types.Geometry,
			infoBuilder{
				info:         "Returns the symmetric difference of both geometries.",
				libraryUsage: usesGEOS,
			},
			tree.VolatilityImmutable,
		),
	),
	"st_simplify": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"tolerance", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600860)
				g := tree.MustBeDGeometry(args[0])
				tolerance := float64(tree.MustBeDFloat(args[1]))

				ret, err := geomfn.SimplifyGEOS(g.Geometry, tolerance)
				if err != nil {
					__antithesis_instrumentation__.Notify(600862)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600863)
				}
				__antithesis_instrumentation__.Notify(600861)
				return &tree.DGeometry{Geometry: ret}, nil
			},
			Info: infoBuilder{
				info:         `Simplifies the given geometry using the Douglas-Peucker algorithm.`,
				libraryUsage: usesGEOS,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"tolerance", types.Float},
				{"preserve_collapsed", types.Bool},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600864)
				g := tree.MustBeDGeometry(args[0])
				tolerance := float64(tree.MustBeDFloat(args[1]))
				preserveCollapsed := bool(tree.MustBeDBool(args[2]))
				ret, collapsed, err := geomfn.Simplify(g.Geometry, tolerance, preserveCollapsed)
				if err != nil {
					__antithesis_instrumentation__.Notify(600867)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600868)
				}
				__antithesis_instrumentation__.Notify(600865)
				if collapsed {
					__antithesis_instrumentation__.Notify(600869)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(600870)
				}
				__antithesis_instrumentation__.Notify(600866)
				return &tree.DGeometry{Geometry: ret}, nil
			},
			Info: infoBuilder{
				info: `Simplifies the given geometry using the Douglas-Peucker algorithm, retaining objects that would be too small given the tolerance if preserve_collapsed is set to true.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_simplifypreservetopology": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"tolerance", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600871)
				g := tree.MustBeDGeometry(args[0])
				tolerance := float64(tree.MustBeDFloat(args[1]))
				ret, err := geomfn.SimplifyPreserveTopology(g.Geometry, tolerance)
				if err != nil {
					__antithesis_instrumentation__.Notify(600873)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600874)
				}
				__antithesis_instrumentation__.Notify(600872)
				return &tree.DGeometry{Geometry: ret}, nil
			},
			Info: infoBuilder{
				info:         `Simplifies the given geometry using the Douglas-Peucker algorithm, avoiding the creation of invalid geometries.`,
				libraryUsage: usesGEOS,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),

	"st_setsrid": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"srid", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600875)
				g := tree.MustBeDGeometry(args[0])
				srid := tree.MustBeDInt(args[1])
				newGeom, err := g.Geometry.CloneWithSRID(geopb.SRID(srid))
				if err != nil {
					__antithesis_instrumentation__.Notify(600877)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600878)
				}
				__antithesis_instrumentation__.Notify(600876)
				return &tree.DGeometry{Geometry: newGeom}, nil
			},
			Info: infoBuilder{
				info: `Sets a Geometry to a new SRID without transforming the coordinates.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geography", types.Geography},
				{"srid", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.Geography),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600879)
				g := tree.MustBeDGeography(args[0])
				srid := tree.MustBeDInt(args[1])
				newGeom, err := g.Geography.CloneWithSRID(geopb.SRID(srid))
				if err != nil {
					__antithesis_instrumentation__.Notify(600881)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600882)
				}
				__antithesis_instrumentation__.Notify(600880)
				return &tree.DGeography{Geography: newGeom}, nil
			},
			Info: infoBuilder{
				info: `Sets a Geography to a new SRID without transforming the coordinates.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_transform": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"srid", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600883)
				g := tree.MustBeDGeometry(args[0])
				srid := geopb.SRID(tree.MustBeDInt(args[1]))

				fromProj, err := geoprojbase.Projection(g.SRID())
				if err != nil {
					__antithesis_instrumentation__.Notify(600887)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600888)
				}
				__antithesis_instrumentation__.Notify(600884)
				toProj, err := geoprojbase.Projection(srid)
				if err != nil {
					__antithesis_instrumentation__.Notify(600889)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600890)
				}
				__antithesis_instrumentation__.Notify(600885)
				ret, err := geotransform.Transform(g.Geometry, fromProj.Proj4Text, toProj.Proj4Text, srid)
				if err != nil {
					__antithesis_instrumentation__.Notify(600891)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600892)
				}
				__antithesis_instrumentation__.Notify(600886)
				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info:         `Transforms a geometry into the given SRID coordinate reference system by projecting its coordinates.`,
				libraryUsage: usesPROJ,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"to_proj_text", types.String},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600893)
				g := tree.MustBeDGeometry(args[0])
				toProj := string(tree.MustBeDString(args[1]))

				fromProj, err := geoprojbase.Projection(g.SRID())
				if err != nil {
					__antithesis_instrumentation__.Notify(600896)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600897)
				}
				__antithesis_instrumentation__.Notify(600894)
				ret, err := geotransform.Transform(
					g.Geometry,
					fromProj.Proj4Text,
					geoprojbase.MakeProj4Text(toProj),
					geopb.SRID(0),
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(600898)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600899)
				}
				__antithesis_instrumentation__.Notify(600895)
				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info:         `Transforms a geometry into the coordinate reference system referenced by the projection text by projecting its coordinates.`,
				libraryUsage: usesPROJ,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"from_proj_text", types.String},
				{"to_proj_text", types.String},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600900)
				g := tree.MustBeDGeometry(args[0])
				fromProj := string(tree.MustBeDString(args[1]))
				toProj := string(tree.MustBeDString(args[2]))

				ret, err := geotransform.Transform(
					g.Geometry,
					geoprojbase.MakeProj4Text(fromProj),
					geoprojbase.MakeProj4Text(toProj),
					geopb.SRID(0),
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(600902)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600903)
				}
				__antithesis_instrumentation__.Notify(600901)
				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info:         `Transforms a geometry into the coordinate reference system assuming the from_proj_text to the new to_proj_text by projecting its coordinates.`,
				libraryUsage: usesPROJ,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"from_proj_text", types.String},
				{"srid", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600904)
				g := tree.MustBeDGeometry(args[0])
				fromProj := string(tree.MustBeDString(args[1]))
				srid := geopb.SRID(tree.MustBeDInt(args[2]))

				toProj, err := geoprojbase.Projection(srid)
				if err != nil {
					__antithesis_instrumentation__.Notify(600907)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600908)
				}
				__antithesis_instrumentation__.Notify(600905)
				ret, err := geotransform.Transform(
					g.Geometry,
					geoprojbase.MakeProj4Text(fromProj),
					toProj.Proj4Text,
					srid,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(600909)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600910)
				}
				__antithesis_instrumentation__.Notify(600906)
				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info:         `Transforms a geometry into the coordinate reference system assuming the from_proj_text to the new to_proj_text by projecting its coordinates. The supplied SRID is set on the new geometry.`,
				libraryUsage: usesPROJ,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_translate": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"g", types.Geometry},
				{"delta_x", types.Float},
				{"delta_y", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600911)
				g := tree.MustBeDGeometry(args[0])
				deltaX := float64(tree.MustBeDFloat(args[1]))
				deltaY := float64(tree.MustBeDFloat(args[2]))

				ret, err := geomfn.Translate(g.Geometry, []float64{deltaX, deltaY})
				if err != nil {
					__antithesis_instrumentation__.Notify(600913)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600914)
				}
				__antithesis_instrumentation__.Notify(600912)

				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: `Returns a modified Geometry translated by the given deltas.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"g", types.Geometry},
				{"delta_x", types.Float},
				{"delta_y", types.Float},
				{"delta_z", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600915)
				g := tree.MustBeDGeometry(args[0])
				deltaX := float64(tree.MustBeDFloat(args[1]))
				deltaY := float64(tree.MustBeDFloat(args[2]))
				deltaZ := float64(tree.MustBeDFloat(args[3]))

				ret, err := geomfn.Translate(g.Geometry, []float64{deltaX, deltaY, deltaZ})
				if err != nil {
					__antithesis_instrumentation__.Notify(600917)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600918)
				}
				__antithesis_instrumentation__.Notify(600916)

				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: `Returns a modified Geometry translated by the given deltas.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_affine": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"a", types.Float},
				{"b", types.Float},
				{"d", types.Float},
				{"e", types.Float},
				{"x_off", types.Float},
				{"y_off", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600919)
				g := tree.MustBeDGeometry(args[0])
				ret, err := geomfn.Affine(
					g.Geometry,
					geomfn.AffineMatrix([][]float64{
						{float64(tree.MustBeDFloat(args[1])), float64(tree.MustBeDFloat(args[2])), 0, float64(tree.MustBeDFloat(args[5]))},
						{float64(tree.MustBeDFloat(args[3])), float64(tree.MustBeDFloat(args[4])), 0, float64(tree.MustBeDFloat(args[6]))},
						{0, 0, 1, 0},
						{0, 0, 0, 1},
					}),
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(600921)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600922)
				}
				__antithesis_instrumentation__.Notify(600920)
				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: `Applies a 2D affine transformation to the given geometry.

The matrix transformation will be applied as follows for each coordinate:
/ a  b  x_off \  / x \
| d  e  y_off |  | y |
\ 0  0      1 /  \ 0 /
				`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"a", types.Float},
				{"b", types.Float},
				{"c", types.Float},
				{"d", types.Float},
				{"e", types.Float},
				{"f", types.Float},
				{"g", types.Float},
				{"h", types.Float},
				{"i", types.Float},
				{"x_off", types.Float},
				{"y_off", types.Float},
				{"z_off", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600923)
				g := tree.MustBeDGeometry(args[0])
				ret, err := geomfn.Affine(
					g.Geometry,
					geomfn.AffineMatrix([][]float64{
						{float64(tree.MustBeDFloat(args[1])), float64(tree.MustBeDFloat(args[2])), float64(tree.MustBeDFloat(args[3])), float64(tree.MustBeDFloat(args[10]))},
						{float64(tree.MustBeDFloat(args[4])), float64(tree.MustBeDFloat(args[5])), float64(tree.MustBeDFloat(args[6])), float64(tree.MustBeDFloat(args[11]))},
						{float64(tree.MustBeDFloat(args[7])), float64(tree.MustBeDFloat(args[8])), float64(tree.MustBeDFloat(args[9])), float64(tree.MustBeDFloat(args[12]))},
						{0, 0, 0, 1},
					}),
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(600925)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600926)
				}
				__antithesis_instrumentation__.Notify(600924)
				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: `Applies a 3D affine transformation to the given geometry.

The matrix transformation will be applied as follows for each coordinate:
/ a  b  c x_off \  / x \
| d  e  f y_off |  | y |
| g  h  i z_off |  | z |
\ 0  0  0     1 /  \ 0 /
				`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_scale": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"x_factor", types.Float},
				{"y_factor", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600927)
				g := tree.MustBeDGeometry(args[0])
				xFactor := float64(tree.MustBeDFloat(args[1]))
				yFactor := float64(tree.MustBeDFloat(args[2]))

				ret, err := geomfn.Scale(g.Geometry, []float64{xFactor, yFactor})
				if err != nil {
					__antithesis_instrumentation__.Notify(600929)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600930)
				}
				__antithesis_instrumentation__.Notify(600928)

				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: `Returns a modified Geometry scaled by the given factors.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"g", types.Geometry},
				{"factor", types.Geometry},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600931)
				g := tree.MustBeDGeometry(args[0])
				factor, err := tree.MustBeDGeometry(args[1]).AsGeomT()
				if err != nil {
					__antithesis_instrumentation__.Notify(600936)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600937)
				}
				__antithesis_instrumentation__.Notify(600932)

				pointFactor, ok := factor.(*geom.Point)
				if !ok {
					__antithesis_instrumentation__.Notify(600938)
					return nil, errors.Newf("a Point must be used as the scaling factor")
				} else {
					__antithesis_instrumentation__.Notify(600939)
				}
				__antithesis_instrumentation__.Notify(600933)

				factors := pointFactor.FlatCoords()
				if len(factors) < 2 {
					__antithesis_instrumentation__.Notify(600940)

					factors = []float64{0, 0, 1}
				} else {
					__antithesis_instrumentation__.Notify(600941)
				}
				__antithesis_instrumentation__.Notify(600934)
				ret, err := geomfn.Scale(g.Geometry, factors)
				if err != nil {
					__antithesis_instrumentation__.Notify(600942)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600943)
				}
				__antithesis_instrumentation__.Notify(600935)

				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: `Returns a modified Geometry scaled by taking in a Geometry as the factor.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"g", types.Geometry},
				{"factor", types.Geometry},
				{"origin", types.Geometry},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600944)
				g := tree.MustBeDGeometry(args[0])
				factor := tree.MustBeDGeometry(args[1])
				origin := tree.MustBeDGeometry(args[2])

				ret, err := geomfn.ScaleRelativeToOrigin(g.Geometry, factor.Geometry, origin.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600946)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600947)
				}
				__antithesis_instrumentation__.Notify(600945)

				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: `Returns a modified Geometry scaled by the Geometry factor relative to a false origin.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_rotate": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"g", types.Geometry},
				{"angle_radians", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600948)
				g := tree.MustBeDGeometry(args[0])
				rotRadians := float64(tree.MustBeDFloat(args[1]))

				ret, err := geomfn.Rotate(g.Geometry, rotRadians)
				if err != nil {
					__antithesis_instrumentation__.Notify(600950)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600951)
				}
				__antithesis_instrumentation__.Notify(600949)

				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: `Returns a modified Geometry whose coordinates are rotated around the origin by a rotation angle.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"g", types.Geometry},
				{"angle_radians", types.Float},
				{"origin_x", types.Float},
				{"origin_y", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600952)
				g := tree.MustBeDGeometry(args[0])
				rotRadians := float64(tree.MustBeDFloat(args[1]))
				x := float64(tree.MustBeDFloat(args[2]))
				y := float64(tree.MustBeDFloat(args[3]))
				geometry, err := geomfn.RotateWithXY(g.Geometry, rotRadians, x, y)
				if err != nil {
					__antithesis_instrumentation__.Notify(600954)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600955)
				}
				__antithesis_instrumentation__.Notify(600953)
				return tree.NewDGeometry(geometry), nil
			},
			Info: infoBuilder{
				info: `Returns a modified Geometry whose coordinates are rotated around the provided origin by a rotation angle.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"g", types.Geometry},
				{"angle_radians", types.Float},
				{"origin_point", types.Geometry},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600956)
				g := tree.MustBeDGeometry(args[0])
				rotRadians := float64(tree.MustBeDFloat(args[1]))
				originPoint := tree.MustBeDGeometry(args[2])

				ret, err := geomfn.RotateWithPointOrigin(g.Geometry, rotRadians, originPoint.Geometry)
				if errors.Is(err, geomfn.ErrPointOriginEmpty) {
					__antithesis_instrumentation__.Notify(600959)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(600960)
				}
				__antithesis_instrumentation__.Notify(600957)

				if err != nil {
					__antithesis_instrumentation__.Notify(600961)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600962)
				}
				__antithesis_instrumentation__.Notify(600958)
				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: `Returns a modified Geometry whose coordinates are rotated around the provided origin by a rotation angle.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_rotatex": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"g", types.Geometry},
				{"angle_radians", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600963)
				g := tree.MustBeDGeometry(args[0])
				rotRadians := float64(tree.MustBeDFloat(args[1]))

				ret, err := geomfn.RotateX(g.Geometry, rotRadians)
				if err != nil {
					__antithesis_instrumentation__.Notify(600965)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600966)
				}
				__antithesis_instrumentation__.Notify(600964)

				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: `Returns a modified Geometry whose coordinates are rotated about the x axis by a rotation angle.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_rotatey": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"g", types.Geometry},
				{"angle_radians", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600967)
				g := tree.MustBeDGeometry(args[0])
				rotRadians := float64(tree.MustBeDFloat(args[1]))

				ret, err := geomfn.RotateY(g.Geometry, rotRadians)
				if err != nil {
					__antithesis_instrumentation__.Notify(600969)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600970)
				}
				__antithesis_instrumentation__.Notify(600968)

				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: `Returns a modified Geometry whose coordinates are rotated about the y axis by a rotation angle.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_rotatez": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"g", types.Geometry},
				{"angle_radians", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600971)
				g := tree.MustBeDGeometry(args[0])
				rotRadians := float64(tree.MustBeDFloat(args[1]))

				ret, err := geomfn.RotateZ(g.Geometry, rotRadians)
				if err != nil {
					__antithesis_instrumentation__.Notify(600973)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600974)
				}
				__antithesis_instrumentation__.Notify(600972)

				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: `Returns a modified Geometry whose coordinates are rotated about the z axis by a rotation angle.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_addpoint": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"line_string", types.Geometry},
				{"point", types.Geometry},
				{"index", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600975)
				lineString := tree.MustBeDGeometry(args[0])
				point := tree.MustBeDGeometry(args[1])
				index := int(tree.MustBeDInt(args[2]))

				ret, err := geomfn.AddPoint(lineString.Geometry, index, point.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600977)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600978)
				}
				__antithesis_instrumentation__.Notify(600976)

				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: `Adds a Point to a LineString at the given 0-based index (-1 to append).`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"line_string", types.Geometry},
				{"point", types.Geometry},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600979)
				lineString := tree.MustBeDGeometry(args[0])
				point := tree.MustBeDGeometry(args[1])

				ret, err := geomfn.AddPoint(lineString.Geometry, -1, point.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600981)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600982)
				}
				__antithesis_instrumentation__.Notify(600980)

				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: `Adds a Point to the end of a LineString.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_setpoint": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"line_string", types.Geometry},
				{"index", types.Int},
				{"point", types.Geometry},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600983)
				lineString := tree.MustBeDGeometry(args[0])
				index := int(tree.MustBeDInt(args[1]))
				point := tree.MustBeDGeometry(args[2])

				ret, err := geomfn.SetPoint(lineString.Geometry, index, point.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(600985)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600986)
				}
				__antithesis_instrumentation__.Notify(600984)

				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: `Sets the Point at the given 0-based index and returns the modified LineString geometry.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_removepoint": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"line_string", types.Geometry},
				{"index", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600987)
				lineString := tree.MustBeDGeometry(args[0])
				index := int(tree.MustBeDInt(args[1]))

				ret, err := geomfn.RemovePoint(lineString.Geometry, index)
				if err != nil {
					__antithesis_instrumentation__.Notify(600989)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600990)
				}
				__antithesis_instrumentation__.Notify(600988)

				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: `Removes the Point at the given 0-based index and returns the modified LineString geometry.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_removerepeatedpoints": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"tolerance", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600991)
				g := tree.MustBeDGeometry(args[0])
				tolerance := float64(tree.MustBeDFloat(args[1]))

				ret, err := geomfn.RemoveRepeatedPoints(g.Geometry, tolerance)
				if err != nil {
					__antithesis_instrumentation__.Notify(600993)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600994)
				}
				__antithesis_instrumentation__.Notify(600992)

				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: `Returns a geometry with repeated points removed, within the given distance tolerance.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600995)
				g := tree.MustBeDGeometry(args[0])
				ret, err := geomfn.RemoveRepeatedPoints(g.Geometry, 0)
				if err != nil {
					__antithesis_instrumentation__.Notify(600997)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(600998)
				}
				__antithesis_instrumentation__.Notify(600996)

				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: `Returns a geometry with repeated points removed.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_reverse": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(600999)
				geometry := tree.MustBeDGeometry(args[0])

				ret, err := geomfn.Reverse(geometry.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(601001)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601002)
				}
				__antithesis_instrumentation__.Notify(601000)

				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: `Returns a modified geometry by reversing the order of its vertices.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_segmentize": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geography", types.Geography},
				{"max_segment_length_meters", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Geography),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601003)
				g := tree.MustBeDGeography(args[0])
				segmentMaxLength := float64(tree.MustBeDFloat(args[1]))
				segGeography, err := geogfn.Segmentize(g.Geography, segmentMaxLength)
				if err != nil {
					__antithesis_instrumentation__.Notify(601005)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601006)
				}
				__antithesis_instrumentation__.Notify(601004)
				return tree.NewDGeography(segGeography), nil
			},
			Info: infoBuilder{
				info: `Returns a modified Geography having no segment longer than the given max_segment_length meters.

The calculations are done on a sphere.`,
				libraryUsage: usesS2,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"max_segment_length", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601007)
				g := tree.MustBeDGeometry(args[0])
				segmentMaxLength := float64(tree.MustBeDFloat(args[1]))
				segGeometry, err := geomfn.Segmentize(g.Geometry, segmentMaxLength)
				if err != nil {
					__antithesis_instrumentation__.Notify(601009)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601010)
				}
				__antithesis_instrumentation__.Notify(601008)
				return tree.NewDGeometry(segGeometry), nil
			},
			Info: infoBuilder{
				info: `Returns a modified Geometry having no segment longer than the given max_segment_length. ` +
					`Length units are in units of spatial reference.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_snaptogrid": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"size", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601011)
				g := tree.MustBeDGeometry(args[0])
				size := float64(tree.MustBeDFloat(args[1]))
				ret, err := geomfn.SnapToGrid(g.Geometry, geom.Coord{0, 0, 0, 0}, geom.Coord{size, size, 0, 0})
				if err != nil {
					__antithesis_instrumentation__.Notify(601013)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601014)
				}
				__antithesis_instrumentation__.Notify(601012)
				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: "Snap a geometry to a grid of the given size. " +
					"The specified size is only used to snap X and Y coordinates.",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"size_x", types.Float},
				{"size_y", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601015)
				g := tree.MustBeDGeometry(args[0])
				sizeX := float64(tree.MustBeDFloat(args[1]))
				sizeY := float64(tree.MustBeDFloat(args[2]))
				ret, err := geomfn.SnapToGrid(g.Geometry, geom.Coord{0, 0, 0, 0}, geom.Coord{sizeX, sizeY, 0, 0})
				if err != nil {
					__antithesis_instrumentation__.Notify(601017)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601018)
				}
				__antithesis_instrumentation__.Notify(601016)
				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: "Snap a geometry to a grid of with X coordinates snapped to size_x and Y coordinates snapped to size_y.",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"origin_x", types.Float},
				{"origin_y", types.Float},
				{"size_x", types.Float},
				{"size_y", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601019)
				g := tree.MustBeDGeometry(args[0])
				originX := float64(tree.MustBeDFloat(args[1]))
				originY := float64(tree.MustBeDFloat(args[2]))
				sizeX := float64(tree.MustBeDFloat(args[3]))
				sizeY := float64(tree.MustBeDFloat(args[4]))
				ret, err := geomfn.SnapToGrid(g.Geometry, geom.Coord{originX, originY, 0, 0}, geom.Coord{sizeX, sizeY, 0, 0})
				if err != nil {
					__antithesis_instrumentation__.Notify(601021)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601022)
				}
				__antithesis_instrumentation__.Notify(601020)
				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: "Snap a geometry to a grid of with X coordinates snapped to size_x and Y coordinates snapped to size_y based on an origin of (origin_x, origin_y).",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"origin", types.Geometry},
				{"size_x", types.Float},
				{"size_y", types.Float},
				{"size_z", types.Float},
				{"size_m", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601023)
				g := tree.MustBeDGeometry(args[0])
				origin := tree.MustBeDGeometry(args[1])
				sizeX := float64(tree.MustBeDFloat(args[2]))
				sizeY := float64(tree.MustBeDFloat(args[3]))
				sizeZ := float64(tree.MustBeDFloat(args[4]))
				sizeM := float64(tree.MustBeDFloat(args[5]))
				originT, err := origin.Geometry.AsGeomT()
				if err != nil {
					__antithesis_instrumentation__.Notify(601025)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601026)
				}
				__antithesis_instrumentation__.Notify(601024)
				switch originT := originT.(type) {
				case *geom.Point:
					__antithesis_instrumentation__.Notify(601027)

					if originT.Empty() {
						__antithesis_instrumentation__.Notify(601031)
						originT = geom.NewPoint(originT.Layout())
					} else {
						__antithesis_instrumentation__.Notify(601032)
					}
					__antithesis_instrumentation__.Notify(601028)
					ret, err := geomfn.SnapToGrid(
						g.Geometry,
						geom.Coord{originT.X(), originT.Y(), originT.Z(), originT.M()},
						geom.Coord{sizeX, sizeY, sizeZ, sizeM})
					if err != nil {
						__antithesis_instrumentation__.Notify(601033)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(601034)
					}
					__antithesis_instrumentation__.Notify(601029)
					return tree.NewDGeometry(ret), nil
				default:
					__antithesis_instrumentation__.Notify(601030)
					return nil, errors.Newf("origin must be a POINT")
				}
			},
			Info: infoBuilder{
				info: "Snap a geometry to a grid defined by the given origin and X, Y, Z, and M cell sizes. " +
					"Any dimension with a 0 cell size will not be snapped.",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_snap": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"input", types.Geometry},
				{"target", types.Geometry},
				{"tolerance", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601035)
				g1 := tree.MustBeDGeometry(args[0])
				g2 := tree.MustBeDGeometry(args[1])
				tolerance := tree.MustBeDFloat(args[2])
				ret, err := geomfn.Snap(g1.Geometry, g2.Geometry, float64(tolerance))
				if err != nil {
					__antithesis_instrumentation__.Notify(601037)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601038)
				}
				__antithesis_instrumentation__.Notify(601036)
				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: `Snaps the vertices and segments of input geometry the target geometry's vertices.
Tolerance is used to control where snapping is performed. The result geometry is the input geometry with the vertices snapped. 
If no snapping occurs then the input geometry is returned unchanged.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_buffer": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"distance", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601039)
				g := tree.MustBeDGeometry(args[0])
				distance := tree.MustBeDInt(args[1])

				ret, err := geomfn.Buffer(g.Geometry, geomfn.MakeDefaultBufferParams(), float64(distance))
				if err != nil {
					__antithesis_instrumentation__.Notify(601041)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601042)
				}
				__antithesis_instrumentation__.Notify(601040)
				return tree.NewDGeometry(ret), nil
			},
			Info:       stBufferInfoBuilder.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"distance", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601043)
				g := tree.MustBeDGeometry(args[0])
				distance := tree.MustBeDFloat(args[1])

				ret, err := geomfn.Buffer(g.Geometry, geomfn.MakeDefaultBufferParams(), float64(distance))
				if err != nil {
					__antithesis_instrumentation__.Notify(601045)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601046)
				}
				__antithesis_instrumentation__.Notify(601044)
				return tree.NewDGeometry(ret), nil
			},
			Info:       stBufferInfoBuilder.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"distance", types.Decimal},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601047)
				g := tree.MustBeDGeometry(args[0])
				distanceDec := tree.MustBeDDecimal(args[1])

				distance, err := distanceDec.Float64()
				if err != nil {
					__antithesis_instrumentation__.Notify(601050)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601051)
				}
				__antithesis_instrumentation__.Notify(601048)

				ret, err := geomfn.Buffer(g.Geometry, geomfn.MakeDefaultBufferParams(), distance)
				if err != nil {
					__antithesis_instrumentation__.Notify(601052)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601053)
				}
				__antithesis_instrumentation__.Notify(601049)
				return tree.NewDGeometry(ret), nil
			},
			Info:       stBufferInfoBuilder.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"distance", types.Float},
				{"quad_segs", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601054)
				g := tree.MustBeDGeometry(args[0])
				distance := tree.MustBeDFloat(args[1])
				quadSegs := tree.MustBeDInt(args[2])

				ret, err := geomfn.Buffer(
					g.Geometry,
					geomfn.MakeDefaultBufferParams().WithQuadrantSegments(int(quadSegs)),
					float64(distance),
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(601056)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601057)
				}
				__antithesis_instrumentation__.Notify(601055)
				return tree.NewDGeometry(ret), nil
			},
			Info:       stBufferWithQuadSegInfoBuilder.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"distance", types.Float},
				{"buffer_style_params", types.String},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601058)
				g := tree.MustBeDGeometry(args[0])
				distance := tree.MustBeDFloat(args[1])
				paramsString := tree.MustBeDString(args[2])

				params, modifiedDistance, err := geomfn.ParseBufferParams(string(paramsString), float64(distance))
				if err != nil {
					__antithesis_instrumentation__.Notify(601061)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601062)
				}
				__antithesis_instrumentation__.Notify(601059)

				ret, err := geomfn.Buffer(
					g.Geometry,
					params,
					modifiedDistance,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(601063)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601064)
				}
				__antithesis_instrumentation__.Notify(601060)
				return tree.NewDGeometry(ret), nil
			},
			Info:       stBufferWithParamsInfoBuilder.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geography", types.Geography},
				{"distance", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Geography),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601065)
				g := tree.MustBeDGeography(args[0])
				distance := tree.MustBeDFloat(args[1])

				ret, err := performGeographyOperationUsingBestGeomProjection(
					g.Geography,
					func(g geo.Geometry) (geo.Geometry, error) {
						__antithesis_instrumentation__.Notify(601068)
						return geomfn.Buffer(g, geomfn.MakeDefaultBufferParams(), float64(distance))
					},
				)
				__antithesis_instrumentation__.Notify(601066)
				if err != nil {
					__antithesis_instrumentation__.Notify(601069)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601070)
				}
				__antithesis_instrumentation__.Notify(601067)

				return tree.NewDGeography(ret), nil
			},
			Info:       stBufferInfoBuilder.String() + usingBestGeomProjectionWarning,
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geography", types.Geography},
				{"distance", types.Float},
				{"quad_segs", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.Geography),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601071)
				g := tree.MustBeDGeography(args[0])
				distance := tree.MustBeDFloat(args[1])
				quadSegs := tree.MustBeDInt(args[2])

				ret, err := performGeographyOperationUsingBestGeomProjection(
					g.Geography,
					func(g geo.Geometry) (geo.Geometry, error) {
						__antithesis_instrumentation__.Notify(601074)
						return geomfn.Buffer(
							g,
							geomfn.MakeDefaultBufferParams().WithQuadrantSegments(int(quadSegs)),
							float64(distance),
						)
					},
				)
				__antithesis_instrumentation__.Notify(601072)
				if err != nil {
					__antithesis_instrumentation__.Notify(601075)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601076)
				}
				__antithesis_instrumentation__.Notify(601073)
				return tree.NewDGeography(ret), nil
			},
			Info:       stBufferWithQuadSegInfoBuilder.String() + usingBestGeomProjectionWarning,
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geography", types.Geography},
				{"distance", types.Float},
				{"buffer_style_params", types.String},
			},
			ReturnType: tree.FixedReturnType(types.Geography),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601077)
				g := tree.MustBeDGeography(args[0])
				distance := tree.MustBeDFloat(args[1])
				paramsString := tree.MustBeDString(args[2])

				params, modifiedDistance, err := geomfn.ParseBufferParams(string(paramsString), float64(distance))
				if err != nil {
					__antithesis_instrumentation__.Notify(601081)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601082)
				}
				__antithesis_instrumentation__.Notify(601078)

				ret, err := performGeographyOperationUsingBestGeomProjection(
					g.Geography,
					func(g geo.Geometry) (geo.Geometry, error) {
						__antithesis_instrumentation__.Notify(601083)
						return geomfn.Buffer(
							g,
							params,
							modifiedDistance,
						)
					},
				)
				__antithesis_instrumentation__.Notify(601079)
				if err != nil {
					__antithesis_instrumentation__.Notify(601084)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601085)
				}
				__antithesis_instrumentation__.Notify(601080)
				return tree.NewDGeography(ret), nil
			},
			Info:       stBufferWithParamsInfoBuilder.String() + usingBestGeomProjectionWarning,
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_envelope": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(ctx *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601086)
				envelope, err := geomfn.Envelope(g.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(601088)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601089)
				}
				__antithesis_instrumentation__.Notify(601087)
				return tree.NewDGeometry(envelope), nil
			},
			types.Geometry,
			infoBuilder{
				info: `Returns a bounding envelope for the given geometry.

For geometries which have a POINT or LINESTRING bounding box (i.e. is a single point
or a horizontal or vertical line), a POINT or LINESTRING is returned. Otherwise, the
returned POLYGON will be ordered Bottom Left, Top Left, Top Right, Bottom Right,
Bottom Left.`,
			},
			tree.VolatilityImmutable,
		),
		tree.Overload{
			Types: tree.ArgTypes{
				{"box2d", types.Box2D},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601090)
				bbox := tree.MustBeDBox2D(args[0]).CartesianBoundingBox
				ret, err := geo.MakeGeometryFromGeomT(bbox.ToGeomT(0))
				if err != nil {
					__antithesis_instrumentation__.Notify(601092)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601093)
				}
				__antithesis_instrumentation__.Notify(601091)
				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: "Returns a bounding geometry for the given box.",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_flipcoordinates": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601094)
				g := tree.MustBeDGeometry(args[0])

				ret, err := geomfn.FlipCoordinates(g.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(601096)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601097)
				}
				__antithesis_instrumentation__.Notify(601095)

				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: "Returns a new geometry with the X and Y axes flipped.",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_swapordinates": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{
					Name: "geometry",
					Typ:  types.Geometry,
				},
				{
					Name: "swap_ordinate_string",
					Typ:  types.String,
				},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601098)
				g := tree.MustBeDGeometry(args[0])
				cString := tree.MustBeDString(args[1])

				ret, err := geomfn.SwapOrdinates(g.Geometry, string(cString))
				if err != nil {
					__antithesis_instrumentation__.Notify(601100)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601101)
				}
				__antithesis_instrumentation__.Notify(601099)

				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: `Returns a version of the given geometry with given ordinates swapped.
The swap_ordinate_string parameter is a 2-character string naming the ordinates to swap. Valid names are: x, y, z and m.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),

	"st_angle": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"point1", types.Geometry},
				{"point2", types.Geometry},
				{"point3", types.Geometry},
				{"point4", types.Geometry},
			},
			ReturnType: tree.FixedReturnType(types.Float),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601102)
				g1 := tree.MustBeDGeometry(args[0]).Geometry
				g2 := tree.MustBeDGeometry(args[1]).Geometry
				g3 := tree.MustBeDGeometry(args[2]).Geometry
				g4 := tree.MustBeDGeometry(args[3]).Geometry
				angle, err := geomfn.Angle(g1, g2, g3, g4)
				if err != nil {
					__antithesis_instrumentation__.Notify(601105)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601106)
				}
				__antithesis_instrumentation__.Notify(601103)
				if angle == nil {
					__antithesis_instrumentation__.Notify(601107)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(601108)
				}
				__antithesis_instrumentation__.Notify(601104)
				return tree.NewDFloat(tree.DFloat(*angle)), nil
			},
			Info: infoBuilder{
				info: `Returns the clockwise angle between the vectors formed by point1,point2 and point3,point4. ` +
					`The arguments must be POINT geometries. Returns NULL if any vectors have 0 length.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"point1", types.Geometry},
				{"point2", types.Geometry},
				{"point3", types.Geometry},
			},
			ReturnType: tree.FixedReturnType(types.Float),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601109)
				g1 := tree.MustBeDGeometry(args[0]).Geometry
				g2 := tree.MustBeDGeometry(args[1]).Geometry
				g3 := tree.MustBeDGeometry(args[2]).Geometry
				g4, err := geo.MakeGeometryFromGeomT(geom.NewPointEmpty(geom.XY))
				if err != nil {
					__antithesis_instrumentation__.Notify(601113)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601114)
				}
				__antithesis_instrumentation__.Notify(601110)
				angle, err := geomfn.Angle(g1, g2, g3, g4)
				if err != nil {
					__antithesis_instrumentation__.Notify(601115)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601116)
				}
				__antithesis_instrumentation__.Notify(601111)
				if angle == nil {
					__antithesis_instrumentation__.Notify(601117)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(601118)
				}
				__antithesis_instrumentation__.Notify(601112)
				return tree.NewDFloat(tree.DFloat(*angle)), nil
			},
			Info: infoBuilder{
				info: `Returns the clockwise angle between the vectors formed by point2,point1 and point2,point3. ` +
					`The arguments must be POINT geometries. Returns NULL if any vectors have 0 length.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"line1", types.Geometry},
				{"line2", types.Geometry},
			},
			ReturnType: tree.FixedReturnType(types.Float),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601119)
				g1 := tree.MustBeDGeometry(args[0]).Geometry
				g2 := tree.MustBeDGeometry(args[1]).Geometry
				angle, err := geomfn.AngleLineString(g1, g2)
				if err != nil {
					__antithesis_instrumentation__.Notify(601122)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601123)
				}
				__antithesis_instrumentation__.Notify(601120)
				if angle == nil {
					__antithesis_instrumentation__.Notify(601124)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(601125)
				}
				__antithesis_instrumentation__.Notify(601121)
				return tree.NewDFloat(tree.DFloat(*angle)), nil
			},
			Info: infoBuilder{
				info: `Returns the clockwise angle between two LINESTRING geometries, treating them as vectors ` +
					`between their start- and endpoints. Returns NULL if any vectors have 0 length.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),

	"st_asencodedpolyline": makeBuiltin(defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601126)
				g := tree.MustBeDGeometry(args[0]).Geometry
				s, err := geo.GeometryToEncodedPolyline(g, 5)
				if err != nil {
					__antithesis_instrumentation__.Notify(601128)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601129)
				}
				__antithesis_instrumentation__.Notify(601127)
				return tree.NewDString(s), nil
			},
			Info: infoBuilder{
				info: `Returns the geometry as an Encoded Polyline.
This format is used by Google Maps with precision=5 and by Open Source Routing Machine with precision=5 and 6.
Preserves 5 decimal places.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"precision", types.Int4},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601130)
				g := tree.MustBeDGeometry(args[0]).Geometry
				p := int(tree.MustBeDInt(args[1]))
				s, err := geo.GeometryToEncodedPolyline(g, p)
				if err != nil {
					__antithesis_instrumentation__.Notify(601132)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601133)
				}
				__antithesis_instrumentation__.Notify(601131)
				return tree.NewDString(s), nil
			},
			Info: infoBuilder{
				info: `Returns the geometry as an Encoded Polyline.
This format is used by Google Maps with precision=5 and by Open Source Routing Machine with precision=5 and 6.
Precision specifies how many decimal places will be preserved in Encoded Polyline. Value should be the same on encoding and decoding, or coordinates will be incorrect.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_linefromencodedpolyline": makeBuiltin(defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"encoded_polyline", types.String},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601134)
				s := string(tree.MustBeDString(args[0]))
				p := 5
				g, err := geo.ParseEncodedPolyline(s, p)
				if err != nil {
					__antithesis_instrumentation__.Notify(601136)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601137)
				}
				__antithesis_instrumentation__.Notify(601135)
				return tree.NewDGeometry(g), nil
			},
			Info: infoBuilder{
				info: `Creates a LineString from an Encoded Polyline string.

Returns valid results only if the polyline was encoded with 5 decimal places.

See http://developers.google.com/maps/documentation/utilities/polylinealgorithm`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"encoded_polyline", types.String},
				{"precision", types.Int4},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601138)
				s := string(tree.MustBeDString(args[0]))
				p := int(tree.MustBeDInt(args[1]))
				g, err := geo.ParseEncodedPolyline(s, p)
				if err != nil {
					__antithesis_instrumentation__.Notify(601140)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601141)
				}
				__antithesis_instrumentation__.Notify(601139)
				return tree.NewDGeometry(g), nil
			},
			Info: infoBuilder{
				info: `Creates a LineString from an Encoded Polyline string.

Precision specifies how many decimal places will be preserved in Encoded Polyline. Value should be the same on encoding and decoding, or coordinates will be incorrect.

See http://developers.google.com/maps/documentation/utilities/polylinealgorithm`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_unaryunion": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(_ *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601142)
				res, err := geomfn.UnaryUnion(g.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(601144)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601145)
				}
				__antithesis_instrumentation__.Notify(601143)
				return tree.NewDGeometry(res), nil
			},
			types.Geometry,
			infoBuilder{
				info: "Returns a union of the components for any geometry or geometry collection provided. Dissolves boundaries of a multipolygon.",
			},
			tree.VolatilityImmutable,
		),
	),
	"st_node": makeBuiltin(
		defProps(),
		geometryOverload1(
			func(_ *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601146)
				res, err := geomfn.Node(g.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(601148)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601149)
				}
				__antithesis_instrumentation__.Notify(601147)
				return tree.NewDGeometry(res), nil
			},
			types.Geometry,
			infoBuilder{
				info: "Adds a node on a geometry for each intersection. Resulting geometry is always a MultiLineString.",
			},
			tree.VolatilityImmutable,
		),
	),
	"st_subdivide": makeBuiltin(
		genProps(),
		makeGeneratorOverload(
			tree.ArgTypes{
				{"geometry", types.Geometry},
			},
			types.Geometry,
			makeSubdividedGeometriesGeneratorFactory(false),
			"Returns a geometry divided into parts, where each part contains no more than 256 vertices.",
			tree.VolatilityImmutable,
		),
		makeGeneratorOverload(
			tree.ArgTypes{
				{"geometry", types.Geometry},
				{"max_vertices", types.Int4},
			},
			types.Geometry,
			makeSubdividedGeometriesGeneratorFactory(true),
			"Returns a geometry divided into parts, where each part contains no more than the number of vertices provided.",
			tree.VolatilityImmutable,
		),
	),

	"st_makebox2d": makeBuiltin(
		defProps(),
		geometryOverload2(
			func(ctx *tree.EvalContext, a *tree.DGeometry, b *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601150)
				if a.Geometry.SRID() != b.Geometry.SRID() {
					__antithesis_instrumentation__.Notify(601154)
					return nil, geo.NewMismatchingSRIDsError(a.Geometry.SpatialObject(), b.Geometry.SpatialObject())
				} else {
					__antithesis_instrumentation__.Notify(601155)
				}
				__antithesis_instrumentation__.Notify(601151)
				aGeomT, err := a.AsGeomT()
				if err != nil {
					__antithesis_instrumentation__.Notify(601156)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601157)
				}
				__antithesis_instrumentation__.Notify(601152)
				bGeomT, err := b.AsGeomT()
				if err != nil {
					__antithesis_instrumentation__.Notify(601158)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601159)
				}
				__antithesis_instrumentation__.Notify(601153)

				switch aGeomT := aGeomT.(type) {
				case *geom.Point:
					__antithesis_instrumentation__.Notify(601160)
					switch bGeomT := bGeomT.(type) {
					case *geom.Point:
						__antithesis_instrumentation__.Notify(601162)
						if aGeomT.Empty() || func() bool {
							__antithesis_instrumentation__.Notify(601165)
							return bGeomT.Empty() == true
						}() == true {
							__antithesis_instrumentation__.Notify(601166)
							return nil, pgerror.Newf(pgcode.InvalidParameterValue, "cannot use POINT EMPTY")
						} else {
							__antithesis_instrumentation__.Notify(601167)
						}
						__antithesis_instrumentation__.Notify(601163)
						bbox := a.CartesianBoundingBox().Combine(b.CartesianBoundingBox())
						return tree.NewDBox2D(*bbox), nil
					default:
						__antithesis_instrumentation__.Notify(601164)
						return nil, pgerror.Newf(pgcode.InvalidParameterValue, "second argument is not a POINT")
					}
				default:
					__antithesis_instrumentation__.Notify(601161)
					return nil, pgerror.Newf(pgcode.InvalidParameterValue, "first argument is not a POINT")
				}
			},
			types.Box2D,
			infoBuilder{
				info: "Creates a box2d from two points. Errors if arguments are not two non-empty points.",
			},
			tree.VolatilityImmutable,
		),
	),
	"st_combinebbox": makeBuiltin(
		tree.FunctionProperties{NullableArgs: true},
		tree.Overload{
			Types:      tree.ArgTypes{{"box2d", types.Box2D}, {"geometry", types.Geometry}},
			ReturnType: tree.FixedReturnType(types.Box2D),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601168)
				if args[1] == tree.DNull {
					__antithesis_instrumentation__.Notify(601172)
					return args[0], nil
				} else {
					__antithesis_instrumentation__.Notify(601173)
				}
				__antithesis_instrumentation__.Notify(601169)
				if args[0] == tree.DNull {
					__antithesis_instrumentation__.Notify(601174)
					bbox := tree.MustBeDGeometry(args[1]).CartesianBoundingBox()
					if bbox == nil {
						__antithesis_instrumentation__.Notify(601176)
						return tree.DNull, nil
					} else {
						__antithesis_instrumentation__.Notify(601177)
					}
					__antithesis_instrumentation__.Notify(601175)
					return tree.NewDBox2D(*bbox), nil
				} else {
					__antithesis_instrumentation__.Notify(601178)
				}
				__antithesis_instrumentation__.Notify(601170)
				bbox := &tree.MustBeDBox2D(args[0]).CartesianBoundingBox
				g := tree.MustBeDGeometry(args[1])
				bbox = bbox.Combine(g.CartesianBoundingBox())
				if bbox == nil {
					__antithesis_instrumentation__.Notify(601179)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(601180)
				}
				__antithesis_instrumentation__.Notify(601171)
				return tree.NewDBox2D(*bbox), nil
			},
			Info: infoBuilder{
				info: "Combines the current bounding box with the bounding box of the Geometry.",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_expand": makeBuiltin(
		defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"box2d", types.Box2D}, {"delta", types.Float}},
			ReturnType: tree.FixedReturnType(types.Box2D),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601181)
				bbox := tree.MustBeDBox2D(args[0])
				delta := float64(tree.MustBeDFloat(args[1]))
				bboxBuffered := bbox.Buffer(delta, delta)
				if bboxBuffered == nil {
					__antithesis_instrumentation__.Notify(601183)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(601184)
				}
				__antithesis_instrumentation__.Notify(601182)
				return tree.NewDBox2D(*bboxBuffered), nil
			},
			Info: infoBuilder{
				info: "Extends the box2d by delta units across all dimensions.",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"box2d", types.Box2D},
				{"delta_x", types.Float},
				{"delta_y", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Box2D),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601185)
				bbox := tree.MustBeDBox2D(args[0])
				deltaX := float64(tree.MustBeDFloat(args[1]))
				deltaY := float64(tree.MustBeDFloat(args[2]))
				bboxBuffered := bbox.Buffer(deltaX, deltaY)
				if bboxBuffered == nil {
					__antithesis_instrumentation__.Notify(601187)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(601188)
				}
				__antithesis_instrumentation__.Notify(601186)
				return tree.NewDBox2D(*bboxBuffered), nil
			},
			Info: infoBuilder{
				info: "Extends the box2d by delta_x units in the x dimension and delta_y units in the y dimension.",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"geometry", types.Geometry}, {"delta", types.Float}},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601189)
				g := tree.MustBeDGeometry(args[0])
				delta := float64(tree.MustBeDFloat(args[1]))
				if g.Empty() {
					__antithesis_instrumentation__.Notify(601192)
					return g, nil
				} else {
					__antithesis_instrumentation__.Notify(601193)
				}
				__antithesis_instrumentation__.Notify(601190)
				bbox := g.CartesianBoundingBox().Buffer(delta, delta)
				ret, err := geo.MakeGeometryFromGeomT(bbox.ToGeomT(g.SRID()))
				if err != nil {
					__antithesis_instrumentation__.Notify(601194)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601195)
				}
				__antithesis_instrumentation__.Notify(601191)
				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: "Extends the bounding box represented by the geometry by delta units across all dimensions, returning a Polygon representing the new bounding box.",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"delta_x", types.Float},
				{"delta_y", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601196)
				g := tree.MustBeDGeometry(args[0])
				deltaX := float64(tree.MustBeDFloat(args[1]))
				deltaY := float64(tree.MustBeDFloat(args[2]))
				if g.Empty() {
					__antithesis_instrumentation__.Notify(601199)
					return g, nil
				} else {
					__antithesis_instrumentation__.Notify(601200)
				}
				__antithesis_instrumentation__.Notify(601197)
				bbox := g.CartesianBoundingBox().Buffer(deltaX, deltaY)
				ret, err := geo.MakeGeometryFromGeomT(bbox.ToGeomT(g.SRID()))
				if err != nil {
					__antithesis_instrumentation__.Notify(601201)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601202)
				}
				__antithesis_instrumentation__.Notify(601198)
				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: "Extends the bounding box represented by the geometry by delta_x units in the x dimension and delta_y units in the y dimension, returning a Polygon representing the new bounding box.",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),

	"st_estimatedextent": makeBuiltin(
		tree.FunctionProperties{
			Category: categorySpatial,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"schema_name", types.String},
				{"table_name", types.String},
				{"geocolumn_name", types.String},
				{"parent_only", types.Bool},
			},
			ReturnType: tree.FixedReturnType(types.Box2D),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601203)

				return tree.DNull, nil
			},
			Info: infoBuilder{
				info: `Returns the estimated extent of the geometries in the column of the given table. This currently always returns NULL.

The parent_only boolean is always ignored.`,
			}.String(),
			Volatility: tree.VolatilityStable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"schema_name", types.String},
				{"table_name", types.String},
				{"geocolumn_name", types.String},
			},
			ReturnType: tree.FixedReturnType(types.Box2D),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601204)

				return tree.DNull, nil
			},
			Info: infoBuilder{
				info: `Returns the estimated extent of the geometries in the column of the given table. This currently always returns NULL.`,
			}.String(),
			Volatility: tree.VolatilityStable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"table_name", types.String},
				{"geocolumn_name", types.String},
			},
			ReturnType: tree.FixedReturnType(types.Box2D),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601205)

				return tree.DNull, nil
			},
			Info: infoBuilder{
				info: `Returns the estimated extent of the geometries in the column of the given table. This currently always returns NULL.`,
			}.String(),
			Volatility: tree.VolatilityStable,
		},
	),

	"addgeometrycolumn": makeBuiltin(
		tree.FunctionProperties{
			Class:    tree.SQLClass,
			Category: categorySpatial,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"table_name", types.String},
				{"column_name", types.String},
				{"srid", types.Int},
				{"type", types.String},
				{"dimension", types.Int},
				{"use_typmod", types.Bool},
			},
			ReturnType: tree.FixedReturnType(types.String),
			SQLFn: func(ctx *tree.EvalContext, args tree.Datums) (string, error) {
				__antithesis_instrumentation__.Notify(601206)
				return addGeometryColumnSQL(
					ctx,
					"",
					"",
					string(tree.MustBeDString(args[0])),
					string(tree.MustBeDString(args[1])),
					int(tree.MustBeDInt(args[2])),
					string(tree.MustBeDString(args[3])),
					int(tree.MustBeDInt(args[4])),
					bool(tree.MustBeDBool(args[5])),
				)
			},
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601207)
				return addGeometryColumnSummary(
					ctx,
					"",
					"",
					string(tree.MustBeDString(args[0])),
					string(tree.MustBeDString(args[1])),
					int(tree.MustBeDInt(args[2])),
					string(tree.MustBeDString(args[3])),
					int(tree.MustBeDInt(args[4])),
					bool(tree.MustBeDBool(args[5])),
				)
			},
			Info: infoBuilder{
				info: `Adds a new geometry column to an existing table and returns metadata about the column created.`,
			}.String(),
			Volatility: tree.VolatilityVolatile,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"schema_name", types.String},
				{"table_name", types.String},
				{"column_name", types.String},
				{"srid", types.Int},
				{"type", types.String},
				{"dimension", types.Int},
				{"use_typmod", types.Bool},
			},
			ReturnType: tree.FixedReturnType(types.String),
			SQLFn: func(ctx *tree.EvalContext, args tree.Datums) (string, error) {
				__antithesis_instrumentation__.Notify(601208)
				return addGeometryColumnSQL(
					ctx,
					"",
					string(tree.MustBeDString(args[0])),
					string(tree.MustBeDString(args[1])),
					string(tree.MustBeDString(args[2])),
					int(tree.MustBeDInt(args[3])),
					string(tree.MustBeDString(args[4])),
					int(tree.MustBeDInt(args[5])),
					bool(tree.MustBeDBool(args[6])),
				)
			},
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601209)
				return addGeometryColumnSummary(
					ctx,
					"",
					string(tree.MustBeDString(args[0])),
					string(tree.MustBeDString(args[1])),
					string(tree.MustBeDString(args[2])),
					int(tree.MustBeDInt(args[3])),
					string(tree.MustBeDString(args[4])),
					int(tree.MustBeDInt(args[5])),
					bool(tree.MustBeDBool(args[6])),
				)
			},
			Info: infoBuilder{
				info: `Adds a new geometry column to an existing table and returns metadata about the column created.`,
			}.String(),
			Volatility: tree.VolatilityVolatile,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"catalog_name", types.String},
				{"schema_name", types.String},
				{"table_name", types.String},
				{"column_name", types.String},
				{"srid", types.Int},
				{"type", types.String},
				{"dimension", types.Int},
				{"use_typmod", types.Bool},
			},
			ReturnType: tree.FixedReturnType(types.String),
			SQLFn: func(ctx *tree.EvalContext, args tree.Datums) (string, error) {
				__antithesis_instrumentation__.Notify(601210)
				return addGeometryColumnSQL(
					ctx,
					string(tree.MustBeDString(args[0])),
					string(tree.MustBeDString(args[1])),
					string(tree.MustBeDString(args[2])),
					string(tree.MustBeDString(args[3])),
					int(tree.MustBeDInt(args[4])),
					string(tree.MustBeDString(args[5])),
					int(tree.MustBeDInt(args[6])),
					bool(tree.MustBeDBool(args[7])),
				)
			},
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601211)
				return addGeometryColumnSummary(
					ctx,
					string(tree.MustBeDString(args[0])),
					string(tree.MustBeDString(args[1])),
					string(tree.MustBeDString(args[2])),
					string(tree.MustBeDString(args[3])),
					int(tree.MustBeDInt(args[4])),
					string(tree.MustBeDString(args[5])),
					int(tree.MustBeDInt(args[6])),
					bool(tree.MustBeDBool(args[7])),
				)
			},
			Info: infoBuilder{
				info: `Adds a new geometry column to an existing table and returns metadata about the column created.`,
			}.String(),
			Volatility: tree.VolatilityVolatile,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"table_name", types.String},
				{"column_name", types.String},
				{"srid", types.Int},
				{"type", types.String},
				{"dimension", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.String),
			SQLFn: func(ctx *tree.EvalContext, args tree.Datums) (string, error) {
				__antithesis_instrumentation__.Notify(601212)
				return addGeometryColumnSQL(
					ctx,
					"",
					"",
					string(tree.MustBeDString(args[0])),
					string(tree.MustBeDString(args[1])),
					int(tree.MustBeDInt(args[2])),
					string(tree.MustBeDString(args[3])),
					int(tree.MustBeDInt(args[4])),
					true,
				)
			},
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601213)
				return addGeometryColumnSummary(
					ctx,
					"",
					"",
					string(tree.MustBeDString(args[0])),
					string(tree.MustBeDString(args[1])),
					int(tree.MustBeDInt(args[2])),
					string(tree.MustBeDString(args[3])),
					int(tree.MustBeDInt(args[4])),
					true,
				)
			},
			Info: infoBuilder{
				info: `Adds a new geometry column to an existing table and returns metadata about the column created.`,
			}.String(),
			Volatility: tree.VolatilityVolatile,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"schema_name", types.String},
				{"table_name", types.String},
				{"column_name", types.String},
				{"srid", types.Int},
				{"type", types.String},
				{"dimension", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.String),
			SQLFn: func(ctx *tree.EvalContext, args tree.Datums) (string, error) {
				__antithesis_instrumentation__.Notify(601214)
				return addGeometryColumnSQL(
					ctx,
					"",
					string(tree.MustBeDString(args[0])),
					string(tree.MustBeDString(args[1])),
					string(tree.MustBeDString(args[2])),
					int(tree.MustBeDInt(args[3])),
					string(tree.MustBeDString(args[4])),
					int(tree.MustBeDInt(args[5])),
					true,
				)
			},
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601215)
				return addGeometryColumnSummary(
					ctx,
					"",
					string(tree.MustBeDString(args[0])),
					string(tree.MustBeDString(args[1])),
					string(tree.MustBeDString(args[2])),
					int(tree.MustBeDInt(args[3])),
					string(tree.MustBeDString(args[4])),
					int(tree.MustBeDInt(args[5])),
					true,
				)
			},
			Info: infoBuilder{
				info: `Adds a new geometry column to an existing table and returns metadata about the column created.`,
			}.String(),
			Volatility: tree.VolatilityVolatile,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"catalog_name", types.String},
				{"schema_name", types.String},
				{"table_name", types.String},
				{"column_name", types.String},
				{"srid", types.Int},
				{"type", types.String},
				{"dimension", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.String),
			SQLFn: func(ctx *tree.EvalContext, args tree.Datums) (string, error) {
				__antithesis_instrumentation__.Notify(601216)
				return addGeometryColumnSQL(
					ctx,
					string(tree.MustBeDString(args[0])),
					string(tree.MustBeDString(args[1])),
					string(tree.MustBeDString(args[2])),
					string(tree.MustBeDString(args[3])),
					int(tree.MustBeDInt(args[4])),
					string(tree.MustBeDString(args[5])),
					int(tree.MustBeDInt(args[6])),
					true,
				)
			},
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601217)
				return addGeometryColumnSummary(
					ctx,
					string(tree.MustBeDString(args[0])),
					string(tree.MustBeDString(args[1])),
					string(tree.MustBeDString(args[2])),
					string(tree.MustBeDString(args[3])),
					int(tree.MustBeDInt(args[4])),
					string(tree.MustBeDString(args[5])),
					int(tree.MustBeDInt(args[6])),
					true,
				)
			},
			Info: infoBuilder{
				info: `Adds a new geometry column to an existing table and returns metadata about the column created.`,
			}.String(),
			Volatility: tree.VolatilityVolatile,
		},
	),
	"st_dwithinexclusive": makeSTDWithinBuiltin(geo.FnExclusive),
	"st_dfullywithinexclusive": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry_a", types.Geometry},
				{"geometry_b", types.Geometry},
				{"distance", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601218)
				a := tree.MustBeDGeometry(args[0])
				b := tree.MustBeDGeometry(args[1])
				dist := tree.MustBeDFloat(args[2])
				ret, err := geomfn.DFullyWithin(a.Geometry, b.Geometry, float64(dist), geo.FnExclusive)
				if err != nil {
					__antithesis_instrumentation__.Notify(601220)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601221)
				}
				__antithesis_instrumentation__.Notify(601219)
				return tree.MakeDBool(tree.DBool(ret)), nil
			},
			Info: infoBuilder{
				info: "Returns true if every pair of points comprising geometry_a and geometry_b are within distance units, exclusive. " +
					"In other words, the ST_MaxDistance between geometry_a and geometry_b is less than distance units.",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_pointinsidecircle": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"x_coord", types.Float},
				{"y_coord", types.Float},
				{"radius", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601222)
				point := tree.MustBeDGeometry(args[0])

				geomT, err := point.AsGeomT()
				if err != nil {
					__antithesis_instrumentation__.Notify(601228)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601229)
				}
				__antithesis_instrumentation__.Notify(601223)

				if _, ok := geomT.(*geom.Point); !ok {
					__antithesis_instrumentation__.Notify(601230)
					return nil, pgerror.Newf(
						pgcode.InvalidParameterValue,
						"first parameter has to be of type Point",
					)
				} else {
					__antithesis_instrumentation__.Notify(601231)
				}
				__antithesis_instrumentation__.Notify(601224)

				x := float64(tree.MustBeDFloat(args[1]))
				y := float64(tree.MustBeDFloat(args[2]))
				radius := float64(tree.MustBeDFloat(args[3]))
				if radius <= 0 {
					__antithesis_instrumentation__.Notify(601232)
					return nil, pgerror.Newf(
						pgcode.InvalidParameterValue,
						"radius of the circle has to be positive",
					)
				} else {
					__antithesis_instrumentation__.Notify(601233)
				}
				__antithesis_instrumentation__.Notify(601225)
				center, err := geo.MakeGeometryFromPointCoords(x, y)
				if err != nil {
					__antithesis_instrumentation__.Notify(601234)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601235)
				}
				__antithesis_instrumentation__.Notify(601226)
				dist, err := geomfn.MinDistance(point.Geometry, center)
				if err != nil {
					__antithesis_instrumentation__.Notify(601236)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601237)
				}
				__antithesis_instrumentation__.Notify(601227)
				return tree.MakeDBool(dist <= radius), nil
			},
			Info: infoBuilder{
				info: "Returns the true if the geometry is a point and is inside the circle. Returns false otherwise.",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),

	"st_memsize": makeBuiltin(defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{
					Name: "geometry",
					Typ:  types.Geometry,
				},
			},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601238)
				geo := tree.MustBeDGeometry(args[0])
				return tree.NewDInt(tree.DInt(geo.Size())), nil
			},
			Info:       "Returns the amount of memory space (in bytes) the geometry takes.",
			Volatility: tree.VolatilityImmutable,
		}),

	"st_linelocatepoint": makeBuiltin(defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{
					Name: "line",
					Typ:  types.Geometry,
				},
				{
					Name: "point",
					Typ:  types.Geometry,
				},
			},
			ReturnType: tree.FixedReturnType(types.Float),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601239)
				line := tree.MustBeDGeometry(args[0])
				p := tree.MustBeDGeometry(args[1])

				fraction, err := geomfn.LineLocatePoint(line.Geometry, p.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(601241)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601242)
				}
				__antithesis_instrumentation__.Notify(601240)
				return tree.NewDFloat(tree.DFloat(fraction)), nil
			},
			Info: "Returns a float between 0 and 1 representing the location of the closest point " +
				"on LineString to the given Point, as a fraction of total 2d line length.",
			Volatility: tree.VolatilityImmutable,
		}),

	"st_minimumboundingradius": makeBuiltin(genProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"geometry", types.Geometry}},
			ReturnType: tree.FixedReturnType(minimumBoundingRadiusReturnType),
			Generator:  makeMinimumBoundGenerator,
			Info:       "Returns a record containing the center point and radius of the smallest circle that can fully contains the given geometry.",
			Volatility: tree.VolatilityImmutable,
		}),

	"st_minimumboundingcircle": makeBuiltin(defProps(),
		geometryOverload1(
			func(evalContext *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601243)
				polygon, _, _, err := geomfn.MinimumBoundingCircle(g.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(601245)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601246)
				}
				__antithesis_instrumentation__.Notify(601244)
				return tree.NewDGeometry(polygon), nil
			},
			types.Geometry,
			infoBuilder{
				info: "Returns the smallest circle polygon that can fully contain a geometry.",
			},
			tree.VolatilityImmutable,
		),
		tree.Overload{
			Types:      tree.ArgTypes{{"geometry", types.Geometry}, {" num_segs", types.Int}},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Info: infoBuilder{
				info: "Returns the smallest circle polygon that can fully contain a geometry.",
			}.String(),
			Volatility: tree.VolatilityImmutable,
			Fn: func(evalContext *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601247)
				g := tree.MustBeDGeometry(args[0])
				numOfSeg := tree.MustBeDInt(args[1])
				_, centroid, radius, err := geomfn.MinimumBoundingCircle(g.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(601250)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601251)
				}
				__antithesis_instrumentation__.Notify(601248)

				polygon, err := geomfn.Buffer(
					centroid,
					geomfn.MakeDefaultBufferParams().WithQuadrantSegments(int(numOfSeg)),
					radius,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(601252)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601253)
				}
				__antithesis_instrumentation__.Notify(601249)

				return tree.NewDGeometry(polygon), nil
			},
		},
	),
	"st_transscale": makeBuiltin(defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"delta_x", types.Float},
				{"delta_y", types.Float},
				{"x_factor", types.Float},
				{"y_factor", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Volatility: tree.VolatilityImmutable,
			Info: infoBuilder{
				info: "Translates the geometry using the deltaX and deltaY args, then scales it using the XFactor, YFactor args, working in 2D only.",
			}.String(),
			Fn: func(_ *tree.EvalContext, datums tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601254)
				g := tree.MustBeDGeometry(datums[0])
				deltaX := float64(tree.MustBeDFloat(datums[1]))
				deltaY := float64(tree.MustBeDFloat(datums[2]))
				xFactor := float64(tree.MustBeDFloat(datums[3]))
				yFactor := float64(tree.MustBeDFloat(datums[4]))
				geometry, err := geomfn.TransScale(g.Geometry, deltaX, deltaY, xFactor, yFactor)
				if err != nil {
					__antithesis_instrumentation__.Notify(601256)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601257)
				}
				__antithesis_instrumentation__.Notify(601255)
				return tree.NewDGeometry(geometry), nil
			},
		}),
	"st_voronoipolygons": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601258)
				g := tree.MustBeDGeometry(args[0])
				var env *geo.Geometry
				ret, err := geomfn.VoronoiDiagram(g.Geometry, env, 0.0, false)
				if err != nil {
					__antithesis_instrumentation__.Notify(601260)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601261)
				}
				__antithesis_instrumentation__.Notify(601259)
				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: `Returns a two-dimensional Voronoi diagram from the vertices of the supplied geometry.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"tolerance", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601262)
				g := tree.MustBeDGeometry(args[0])
				tolerance := tree.MustBeDFloat(args[1])
				var env *geo.Geometry
				ret, err := geomfn.VoronoiDiagram(g.Geometry, env, float64(tolerance), false)
				if err != nil {
					__antithesis_instrumentation__.Notify(601264)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601265)
				}
				__antithesis_instrumentation__.Notify(601263)
				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: `Returns a two-dimensional Voronoi diagram from the vertices of the supplied geometry.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"tolerance", types.Float},
				{"extend_to", types.Geometry},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601266)
				g := tree.MustBeDGeometry(args[0])
				tolerance := tree.MustBeDFloat(args[1])
				env := tree.MustBeDGeometry(args[2])
				ret, err := geomfn.VoronoiDiagram(g.Geometry, &env.Geometry, float64(tolerance), false)
				if err != nil {
					__antithesis_instrumentation__.Notify(601268)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601269)
				}
				__antithesis_instrumentation__.Notify(601267)
				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: `Returns a two-dimensional Voronoi diagram from the vertices of the supplied geometry.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_voronoilines": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601270)
				g := tree.MustBeDGeometry(args[0])
				var env *geo.Geometry
				ret, err := geomfn.VoronoiDiagram(g.Geometry, env, 0.0, true)
				if err != nil {
					__antithesis_instrumentation__.Notify(601272)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601273)
				}
				__antithesis_instrumentation__.Notify(601271)
				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: `Returns a two-dimensional Voronoi diagram from the vertices of the supplied geometry as` +
					`the boundaries between cells in that diagram as a MultiLineString.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"tolerance", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601274)
				g := tree.MustBeDGeometry(args[0])
				tolerance := tree.MustBeDFloat(args[1])
				var env *geo.Geometry
				ret, err := geomfn.VoronoiDiagram(g.Geometry, env, float64(tolerance), true)
				if err != nil {
					__antithesis_instrumentation__.Notify(601276)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601277)
				}
				__antithesis_instrumentation__.Notify(601275)
				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: `Returns a two-dimensional Voronoi diagram from the vertices of the supplied geometry as` +
					`the boundaries between cells in that diagram as a MultiLineString.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
				{"tolerance", types.Float},
				{"extend_to", types.Geometry},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601278)
				g := tree.MustBeDGeometry(args[0])
				tolerance := tree.MustBeDFloat(args[1])
				env := tree.MustBeDGeometry(args[2])
				ret, err := geomfn.VoronoiDiagram(g.Geometry, &env.Geometry, float64(tolerance), true)
				if err != nil {
					__antithesis_instrumentation__.Notify(601280)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601281)
				}
				__antithesis_instrumentation__.Notify(601279)
				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: `Returns a two-dimensional Voronoi diagram from the vertices of the supplied geometry as` +
					`the boundaries between cells in that diagram as a MultiLineString.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_orientedenvelope": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry", types.Geometry},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601282)
				g := tree.MustBeDGeometry(args[0])
				ret, err := geomfn.MinimumRotatedRectangle(g.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(601284)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601285)
				}
				__antithesis_instrumentation__.Notify(601283)
				return tree.NewDGeometry(ret), nil
			},
			Info: infoBuilder{
				info: `Returns a minimum rotated rectangle enclosing a geometry.
Note that more than one minimum rotated rectangle may exist.
May return a Point or LineString in the case of degenerate inputs.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),
	"st_linesubstring": makeBuiltin(defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"linestring", types.Geometry},
				{"start_fraction", types.Float},
				{"end_fraction", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Volatility: tree.VolatilityImmutable,
			Info: infoBuilder{
				info: "Return a linestring being a substring of the input one starting and ending at the given fractions of total 2D length. Second and third arguments are float8 values between 0 and 1.",
			}.String(),
			Fn: func(_ *tree.EvalContext, datums tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601286)
				g := tree.MustBeDGeometry(datums[0])
				startFraction := float64(tree.MustBeDFloat(datums[1]))
				endFraction := float64(tree.MustBeDFloat(datums[2]))
				geometry, err := geomfn.LineSubstring(g.Geometry, startFraction, endFraction)
				if err != nil {
					__antithesis_instrumentation__.Notify(601288)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601289)
				}
				__antithesis_instrumentation__.Notify(601287)
				return tree.NewDGeometry(geometry), nil
			},
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"linestring", types.Geometry},
				{"start_fraction", types.Decimal},
				{"end_fraction", types.Decimal},
			},
			ReturnType: tree.FixedReturnType(types.Geometry),
			Volatility: tree.VolatilityImmutable,
			Info: infoBuilder{
				info: "Return a linestring being a substring of the input one starting and ending at the given fractions of total 2D length. Second and third arguments are float8 values between 0 and 1.",
			}.String(),
			Fn: func(_ *tree.EvalContext, datums tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601290)
				g := tree.MustBeDGeometry(datums[0])
				startFraction := tree.MustBeDDecimal(datums[1])
				startFractionFloat, err := startFraction.Float64()
				if err != nil {
					__antithesis_instrumentation__.Notify(601295)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601296)
				}
				__antithesis_instrumentation__.Notify(601291)
				endFraction := tree.MustBeDDecimal(datums[2])
				endFractionFloat, err := endFraction.Float64()
				if err != nil {
					__antithesis_instrumentation__.Notify(601297)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601298)
				}
				__antithesis_instrumentation__.Notify(601292)

				if g.Empty() && func() bool {
					__antithesis_instrumentation__.Notify(601299)
					return g.ShapeType2D() == geopb.ShapeType_LineString == true
				}() == true {
					__antithesis_instrumentation__.Notify(601300)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(601301)
				}
				__antithesis_instrumentation__.Notify(601293)
				geometry, err := geomfn.LineSubstring(g.Geometry, startFractionFloat, endFractionFloat)
				if err != nil {
					__antithesis_instrumentation__.Notify(601302)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601303)
				}
				__antithesis_instrumentation__.Notify(601294)
				return tree.NewDGeometry(geometry), nil
			},
		}),

	"st_linecrossingdirection": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"linestring_a", types.Geometry},
				{"linestring_b", types.Geometry},
			},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601304)
				linestringA := tree.MustBeDGeometry(args[0])
				linestringB := tree.MustBeDGeometry(args[1])

				ret, err := geomfn.LineCrossingDirection(linestringA.Geometry, linestringB.Geometry)
				if err != nil {
					__antithesis_instrumentation__.Notify(601306)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601307)
				}
				__antithesis_instrumentation__.Notify(601305)
				return tree.NewDInt(tree.DInt(ret)), nil
			},
			Info: infoBuilder{
				info: `Returns an interger value defining behavior of crossing of lines: 
0: lines do not cross,
-1: linestring_b crosses linestring_a from right to left,
1: linestring_b crosses linestring_a from left to right,
-2: linestring_b crosses linestring_a multiple times from right to left,
2: linestring_b crosses linestring_a multiple times from left to right,
-3: linestring_b crosses linestring_a multiple times from left to left,
3: linestring_b crosses linestring_a multiple times from right to right.

Note that the top vertex of the segment touching another line does not count as a crossing, but the bottom vertex of segment touching another line is considered a crossing.`,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	),

	"st_asgml":               makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 48877}),
	"st_aslatlontext":        makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 48882}),
	"st_assvg":               makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 48883}),
	"st_boundingdiagonal":    makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 48889}),
	"st_buildarea":           makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 48892}),
	"st_chaikinsmoothing":    makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 48894}),
	"st_cleangeometry":       makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 48895}),
	"st_clusterdbscan":       makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 48898}),
	"st_clusterintersecting": makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 48899}),
	"st_clusterkmeans":       makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 48900}),
	"st_clusterwithin":       makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 48901}),
	"st_concavehull":         makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 48906}),
	"st_delaunaytriangles":   makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 48915}),
	"st_dump":                makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 49785}),
	"st_dumppoints":          makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 49786}),
	"st_dumprings":           makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 49787}),
	"st_geometricmedian":     makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 48944}),
	"st_interpolatepoint":    makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 48950}),
	"st_isvaliddetail":       makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 48962}),
	"st_length2dspheroid":    makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 48967}),
	"st_lengthspheroid":      makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 48968}),
	"st_polygonize":          makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 49011}),
	"st_quantizecoordinates": makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 49012}),
	"st_seteffectivearea":    makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 49030}),
	"st_simplifyvw":          makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 49039}),
	"st_split":               makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 49045}),
	"st_tileenvelope":        makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 49053}),
	"st_wrapx":               makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 49068}),
	"st_bdpolyfromtext":      makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 48801}),
	"st_geomfromgml":         makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 48807}),
	"st_geomfromtwkb":        makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 48809}),
	"st_gmltosql":            makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 48810}),
}

func returnCompatibilityFixedStringBuiltin(ret string) builtinDefinition {
	__antithesis_instrumentation__.Notify(601308)
	return makeBuiltin(
		defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, _ tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601309)
				return tree.NewDString(ret), nil
			},
			Info: infoBuilder{
				info: "Compatibility placeholder function with PostGIS. Returns a fixed string based on PostGIS 3.0.1, with minor edits.",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	)
}

func geometryOverload1(
	f func(*tree.EvalContext, *tree.DGeometry) (tree.Datum, error),
	returnType *types.T,
	ib infoBuilder,
	volatility tree.Volatility,
) tree.Overload {
	__antithesis_instrumentation__.Notify(601310)
	return tree.Overload{
		Types: tree.ArgTypes{
			{"geometry", types.Geometry},
		},
		ReturnType: tree.FixedReturnType(returnType),
		Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601311)
			a := tree.MustBeDGeometry(args[0])
			return f(ctx, a)
		},
		Info:       ib.String(),
		Volatility: volatility,
	}
}

func geometryOverload1UnaryPredicate(
	f func(geo.Geometry) (bool, error), ib infoBuilder,
) tree.Overload {
	__antithesis_instrumentation__.Notify(601312)
	return geometryOverload1(
		func(_ *tree.EvalContext, g *tree.DGeometry) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601313)
			ret, err := f(g.Geometry)
			if err != nil {
				__antithesis_instrumentation__.Notify(601315)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(601316)
			}
			__antithesis_instrumentation__.Notify(601314)
			return tree.MakeDBool(tree.DBool(ret)), nil
		},
		types.Bool,
		ib,
		tree.VolatilityImmutable,
	)
}

func geometryOverload2(
	f func(*tree.EvalContext, *tree.DGeometry, *tree.DGeometry) (tree.Datum, error),
	returnType *types.T,
	ib infoBuilder,
	volatility tree.Volatility,
) tree.Overload {
	__antithesis_instrumentation__.Notify(601317)
	return tree.Overload{
		Types: tree.ArgTypes{
			{"geometry_a", types.Geometry},
			{"geometry_b", types.Geometry},
		},
		ReturnType: tree.FixedReturnType(returnType),
		Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601318)
			a := tree.MustBeDGeometry(args[0])
			b := tree.MustBeDGeometry(args[1])
			return f(ctx, a, b)
		},
		Info:       ib.String(),
		Volatility: volatility,
	}
}

func geometryOverload2BinaryPredicate(
	f func(geo.Geometry, geo.Geometry) (bool, error), ib infoBuilder,
) tree.Overload {
	__antithesis_instrumentation__.Notify(601319)
	return geometryOverload2(
		func(_ *tree.EvalContext, a *tree.DGeometry, b *tree.DGeometry) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601320)
			ret, err := f(a.Geometry, b.Geometry)
			if err != nil {
				__antithesis_instrumentation__.Notify(601322)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(601323)
			}
			__antithesis_instrumentation__.Notify(601321)
			return tree.MakeDBool(tree.DBool(ret)), nil
		},
		types.Bool,
		ib,
		tree.VolatilityImmutable,
	)
}

func geographyOverload1(
	f func(*tree.EvalContext, *tree.DGeography) (tree.Datum, error),
	returnType *types.T,
	ib infoBuilder,
	volatility tree.Volatility,
) tree.Overload {
	__antithesis_instrumentation__.Notify(601324)
	return tree.Overload{
		Types: tree.ArgTypes{
			{"geography", types.Geography},
		},
		ReturnType: tree.FixedReturnType(returnType),
		Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601325)
			a := tree.MustBeDGeography(args[0])
			return f(ctx, a)
		},
		Info:       ib.String(),
		Volatility: volatility,
	}
}

func geographyOverload1WithUseSpheroid(
	f func(*tree.EvalContext, *tree.DGeography, geogfn.UseSphereOrSpheroid) (tree.Datum, error),
	returnType *types.T,
	ib infoBuilder,
	volatility tree.Volatility,
) []tree.Overload {
	__antithesis_instrumentation__.Notify(601326)
	infoWithSphereAndSpheroid := ib
	infoWithSphereAndSpheroid.libraryUsage = usesS2 | usesGeographicLib
	infoWithSpheroid := ib
	infoWithSpheroid.info += usesSpheroidMessage
	infoWithSpheroid.libraryUsage = usesGeographicLib

	return []tree.Overload{
		{
			Types: tree.ArgTypes{
				{"geography", types.Geography},
			},
			ReturnType: tree.FixedReturnType(returnType),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601327)
				a := tree.MustBeDGeography(args[0])
				return f(ctx, a, geogfn.UseSpheroid)
			},
			Info:       infoWithSpheroid.String(),
			Volatility: volatility,
		},
		{
			Types: tree.ArgTypes{
				{"geography", types.Geography},
				{"use_spheroid", types.Bool},
			},
			ReturnType: tree.FixedReturnType(returnType),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601328)
				a := tree.MustBeDGeography(args[0])
				b := tree.MustBeDBool(args[1])
				return f(ctx, a, toUseSphereOrSpheroid(b))
			},
			Info:       infoWithSphereAndSpheroid.String(),
			Volatility: volatility,
		},
	}
}

func geographyOverload2(
	f func(*tree.EvalContext, *tree.DGeography, *tree.DGeography) (tree.Datum, error),
	returnType *types.T,
	ib infoBuilder,
	volatility tree.Volatility,
) tree.Overload {
	__antithesis_instrumentation__.Notify(601329)
	return tree.Overload{
		Types: tree.ArgTypes{
			{"geography_a", types.Geography},
			{"geography_b", types.Geography},
		},
		ReturnType: tree.FixedReturnType(returnType),
		Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601330)
			a := tree.MustBeDGeography(args[0])
			b := tree.MustBeDGeography(args[1])
			return f(ctx, a, b)
		},
		Info:       ib.String(),
		Volatility: volatility,
	}
}

func geographyOverload2BinaryPredicate(
	f func(geo.Geography, geo.Geography) (bool, error), ib infoBuilder,
) tree.Overload {
	__antithesis_instrumentation__.Notify(601331)
	return geographyOverload2(
		func(_ *tree.EvalContext, a *tree.DGeography, b *tree.DGeography) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601332)
			ret, err := f(a.Geography, b.Geography)
			if err != nil {
				__antithesis_instrumentation__.Notify(601334)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(601335)
			}
			__antithesis_instrumentation__.Notify(601333)
			return tree.MakeDBool(tree.DBool(ret)), nil
		},
		types.Bool,
		ib,
		tree.VolatilityImmutable,
	)
}

func toUseSphereOrSpheroid(useSpheroid tree.DBool) geogfn.UseSphereOrSpheroid {
	__antithesis_instrumentation__.Notify(601336)
	if useSpheroid {
		__antithesis_instrumentation__.Notify(601338)
		return geogfn.UseSpheroid
	} else {
		__antithesis_instrumentation__.Notify(601339)
	}
	__antithesis_instrumentation__.Notify(601337)
	return geogfn.UseSphere
}

func initGeoBuiltins() {
	__antithesis_instrumentation__.Notify(601340)

	for _, alias := range []struct {
		alias       string
		builtinName string
	}{
		{"geomfromewkt", "st_geomfromewkt"},
		{"geomfromewkb", "st_geomfromewkb"},
		{"st_coorddim", "st_ndims"},
		{"st_geogfromtext", "st_geographyfromtext"},
		{"st_geomfromtext", "st_geometryfromtext"},
		{"st_numinteriorring", "st_numinteriorrings"},
		{"st_symmetricdifference", "st_symdifference"},
		{"st_force3d", "st_force3dz"},
	} {
		__antithesis_instrumentation__.Notify(601344)
		if _, ok := geoBuiltins[alias.builtinName]; !ok {
			__antithesis_instrumentation__.Notify(601346)
			panic("expected builtin definition for alias: " + alias.builtinName)
		} else {
			__antithesis_instrumentation__.Notify(601347)
		}
		__antithesis_instrumentation__.Notify(601345)
		geoBuiltins[alias.alias] = geoBuiltins[alias.builtinName]
	}
	__antithesis_instrumentation__.Notify(601341)

	for indexBuiltinName := range geoindex.RelationshipMap {
		__antithesis_instrumentation__.Notify(601348)
		builtin, exists := geoBuiltins[indexBuiltinName]
		if !exists {
			__antithesis_instrumentation__.Notify(601351)
			panic("expected builtin: " + indexBuiltinName)
		} else {
			__antithesis_instrumentation__.Notify(601352)
		}
		__antithesis_instrumentation__.Notify(601349)

		overloads := make([]tree.Overload, len(builtin.overloads))
		for i, ovCopy := range builtin.overloads {
			__antithesis_instrumentation__.Notify(601353)
			builtin.overloads[i].Info += "\n\nThis function variant will attempt to utilize any available spatial index."

			ovCopy.Info += "\n\nThis function variant does not utilize any spatial index."
			overloads[i] = ovCopy
		}
		__antithesis_instrumentation__.Notify(601350)
		underscoreBuiltin := makeBuiltin(
			builtin.props,
			overloads...,
		)
		geoBuiltins["_"+indexBuiltinName] = underscoreBuiltin
	}
	__antithesis_instrumentation__.Notify(601342)

	for _, builtinName := range []string{
		"st_area",
		"st_asewkt",
		"st_asgeojson",

		"st_askml",

		"st_astext",
		"st_buffer",
		"st_centroid",

		"st_coveredby",
		"st_covers",
		"st_distance",
		"st_dwithin",
		"st_intersects",
		"st_intersection",
		"st_length",
		"st_dwithinexclusive",
	} {
		__antithesis_instrumentation__.Notify(601354)
		builtin, exists := geoBuiltins[builtinName]
		if !exists {
			__antithesis_instrumentation__.Notify(601356)
			panic("expected builtin: " + builtinName)
		} else {
			__antithesis_instrumentation__.Notify(601357)
		}
		__antithesis_instrumentation__.Notify(601355)
		geoBuiltins[builtinName] = appendStrArgOverloadForGeometryArgOverloads(builtin)
	}
	__antithesis_instrumentation__.Notify(601343)

	for k, v := range geoBuiltins {
		__antithesis_instrumentation__.Notify(601358)
		if _, exists := builtins[k]; exists {
			__antithesis_instrumentation__.Notify(601360)
			panic("duplicate builtin: " + k)
		} else {
			__antithesis_instrumentation__.Notify(601361)
		}
		__antithesis_instrumentation__.Notify(601359)
		v.props.Category = categorySpatial
		v.props.AvailableOnPublicSchema = true
		builtins[k] = v
	}
}

func addGeometryColumnSQL(
	ctx *tree.EvalContext,
	catalogName string,
	schemaName string,
	tableName string,
	columnName string,
	srid int,
	shape string,
	dimension int,
	useTypmod bool,
) (string, error) {
	__antithesis_instrumentation__.Notify(601362)
	if dimension != 2 {
		__antithesis_instrumentation__.Notify(601365)
		return "", pgerror.Newf(
			pgcode.FeatureNotSupported,
			"only dimension=2 is currently supported",
		)
	} else {
		__antithesis_instrumentation__.Notify(601366)
	}
	__antithesis_instrumentation__.Notify(601363)
	if !useTypmod {
		__antithesis_instrumentation__.Notify(601367)
		return "", unimplemented.NewWithIssue(
			49402,
			"useTypmod=false is currently not supported with AddGeometryColumn",
		)
	} else {
		__antithesis_instrumentation__.Notify(601368)
	}
	__antithesis_instrumentation__.Notify(601364)

	tn := makeTableName(catalogName, schemaName, tableName)
	stmt := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s GEOMETRY(%s,%d)",
		tn.String(),
		columnName,
		shape,
		srid,
	)
	return stmt, nil
}

func addGeometryColumnSummary(
	ctx *tree.EvalContext,
	catalogName string,
	schemaName string,
	tableName string,
	columnName string,
	srid int,
	shape string,
	dimension int,
	useTypmod bool,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(601369)
	tn := makeTableName(catalogName, schemaName, tableName)
	summary := fmt.Sprintf("%s.%s SRID:%d TYPE:%s DIMS:%d",
		tn.String(),
		columnName,
		srid,
		strings.ToUpper(shape),
		dimension,
	)
	return tree.NewDString(summary), nil
}

func makeTableName(catalogName string, schemaName string, tableName string) tree.UnresolvedName {
	__antithesis_instrumentation__.Notify(601370)
	if catalogName != "" {
		__antithesis_instrumentation__.Notify(601372)
		return tree.MakeUnresolvedName(catalogName, schemaName, tableName)
	} else {
		__antithesis_instrumentation__.Notify(601373)
		if schemaName != "" {
			__antithesis_instrumentation__.Notify(601374)
			return tree.MakeUnresolvedName(schemaName, tableName)
		} else {
			__antithesis_instrumentation__.Notify(601375)
		}
	}
	__antithesis_instrumentation__.Notify(601371)
	return tree.MakeUnresolvedName(tableName)
}

func lineInterpolatePointForRepeatOverload(repeat bool, builtinInfo string) tree.Overload {
	__antithesis_instrumentation__.Notify(601376)
	return tree.Overload{
		Types: tree.ArgTypes{
			{"geometry", types.Geometry},
			{"fraction", types.Float},
		},
		ReturnType: tree.FixedReturnType(types.Geometry),
		Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601377)
			g := tree.MustBeDGeometry(args[0])
			fraction := float64(tree.MustBeDFloat(args[1]))
			interpolatedPoints, err := geomfn.LineInterpolatePoints(g.Geometry, fraction, repeat)
			if err != nil {
				__antithesis_instrumentation__.Notify(601379)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(601380)
			}
			__antithesis_instrumentation__.Notify(601378)
			return tree.NewDGeometry(interpolatedPoints), nil
		},
		Info: infoBuilder{
			info:         builtinInfo,
			libraryUsage: usesGEOS,
		}.String(),
		Volatility: tree.VolatilityImmutable,
	}
}

func appendStrArgOverloadForGeometryArgOverloads(def builtinDefinition) builtinDefinition {
	__antithesis_instrumentation__.Notify(601381)
	newOverloads := make([]tree.Overload, len(def.overloads))
	copy(newOverloads, def.overloads)

	for i := range def.overloads {
		__antithesis_instrumentation__.Notify(601383)

		ov := def.overloads[i]

		argTypes, ok := ov.Types.(tree.ArgTypes)
		if !ok {
			__antithesis_instrumentation__.Notify(601389)
			continue
		} else {
			__antithesis_instrumentation__.Notify(601390)
		}
		__antithesis_instrumentation__.Notify(601384)

		var argsToCast util.FastIntSet
		for i, argType := range argTypes {
			__antithesis_instrumentation__.Notify(601391)
			if argType.Typ.Equal(types.Geometry) {
				__antithesis_instrumentation__.Notify(601392)
				argsToCast.Add(i)
			} else {
				__antithesis_instrumentation__.Notify(601393)
			}
		}
		__antithesis_instrumentation__.Notify(601385)
		if argsToCast.Len() == 0 {
			__antithesis_instrumentation__.Notify(601394)
			continue
		} else {
			__antithesis_instrumentation__.Notify(601395)
		}
		__antithesis_instrumentation__.Notify(601386)

		newOverload := ov

		newArgTypes := make(tree.ArgTypes, len(argTypes))
		for i := range argTypes {
			__antithesis_instrumentation__.Notify(601396)
			newArgTypes[i] = argTypes[i]
			if argsToCast.Contains(i) {
				__antithesis_instrumentation__.Notify(601397)
				newArgTypes[i].Name += "_str"
				newArgTypes[i].Typ = types.String
			} else {
				__antithesis_instrumentation__.Notify(601398)
			}
		}
		__antithesis_instrumentation__.Notify(601387)

		newOverload.Types = newArgTypes
		newOverload.Fn = func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601399)
			for i, ok := argsToCast.Next(0); ok; i, ok = argsToCast.Next(i + 1) {
				__antithesis_instrumentation__.Notify(601401)
				arg := string(tree.MustBeDString(args[i]))
				g, err := geo.ParseGeometry(arg)
				if err != nil {
					__antithesis_instrumentation__.Notify(601403)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601404)
				}
				__antithesis_instrumentation__.Notify(601402)
				args[i] = tree.NewDGeometry(g)
			}
			__antithesis_instrumentation__.Notify(601400)
			return ov.Fn(ctx, args)
		}
		__antithesis_instrumentation__.Notify(601388)

		newOverload.Info += `

This variant will cast all geometry_str arguments into Geometry types.
`

		newOverloads = append(newOverloads, newOverload)
	}
	__antithesis_instrumentation__.Notify(601382)

	def.overloads = newOverloads
	return def
}

func stAsGeoJSONFromTuple(
	ctx *tree.EvalContext, tuple *tree.DTuple, geoColumn string, numDecimalDigits int, pretty bool,
) (*tree.DString, error) {
	__antithesis_instrumentation__.Notify(601405)
	typ := tuple.ResolvedType()
	labels := typ.TupleLabels()

	var geometry json.JSON
	properties := json.NewObjectBuilder(len(tuple.D))

	foundGeoColumn := false
	for i, d := range tuple.D {
		__antithesis_instrumentation__.Notify(601412)
		var label string
		if labels != nil {
			__antithesis_instrumentation__.Notify(601417)
			label = labels[i]
		} else {
			__antithesis_instrumentation__.Notify(601418)
		}
		__antithesis_instrumentation__.Notify(601413)

		if label == "" {
			__antithesis_instrumentation__.Notify(601419)
			label = fmt.Sprintf("f%d", i+1)
		} else {
			__antithesis_instrumentation__.Notify(601420)
		}
		__antithesis_instrumentation__.Notify(601414)

		if (geoColumn == "" && func() bool {
			__antithesis_instrumentation__.Notify(601421)
			return geometry == nil == true
		}() == true) || func() bool {
			__antithesis_instrumentation__.Notify(601422)
			return geoColumn == label == true
		}() == true {
			__antithesis_instrumentation__.Notify(601423)

			if g, ok := d.(*tree.DGeometry); ok {
				__antithesis_instrumentation__.Notify(601426)
				foundGeoColumn = true
				var err error
				geometry, err = json.FromSpatialObject(g.SpatialObject(), numDecimalDigits)
				if err != nil {
					__antithesis_instrumentation__.Notify(601428)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601429)
				}
				__antithesis_instrumentation__.Notify(601427)
				continue
			} else {
				__antithesis_instrumentation__.Notify(601430)
			}
			__antithesis_instrumentation__.Notify(601424)
			if g, ok := d.(*tree.DGeography); ok {
				__antithesis_instrumentation__.Notify(601431)
				foundGeoColumn = true
				var err error
				geometry, err = json.FromSpatialObject(g.SpatialObject(), numDecimalDigits)
				if err != nil {
					__antithesis_instrumentation__.Notify(601433)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601434)
				}
				__antithesis_instrumentation__.Notify(601432)
				continue
			} else {
				__antithesis_instrumentation__.Notify(601435)
			}
			__antithesis_instrumentation__.Notify(601425)

			if geoColumn == label {
				__antithesis_instrumentation__.Notify(601436)
				foundGeoColumn = true

				if d == tree.DNull {
					__antithesis_instrumentation__.Notify(601438)
					continue
				} else {
					__antithesis_instrumentation__.Notify(601439)
				}
				__antithesis_instrumentation__.Notify(601437)
				return nil, errors.Newf(
					"expected column %s to be a geo type, but it is of type %s",
					geoColumn,
					d.ResolvedType().SQLString(),
				)
			} else {
				__antithesis_instrumentation__.Notify(601440)
			}
		} else {
			__antithesis_instrumentation__.Notify(601441)
		}
		__antithesis_instrumentation__.Notify(601415)
		tupleJSON, err := tree.AsJSON(
			d,
			ctx.SessionData().DataConversionConfig,
			ctx.GetLocation(),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(601442)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(601443)
		}
		__antithesis_instrumentation__.Notify(601416)
		properties.Add(label, tupleJSON)
	}
	__antithesis_instrumentation__.Notify(601406)

	if !foundGeoColumn && func() bool {
		__antithesis_instrumentation__.Notify(601444)
		return geoColumn != "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(601445)
		return nil, errors.Newf("%q column not found", geoColumn)
	} else {
		__antithesis_instrumentation__.Notify(601446)
	}
	__antithesis_instrumentation__.Notify(601407)

	if geometry == nil {
		__antithesis_instrumentation__.Notify(601447)
		geometryBuilder := json.NewObjectBuilder(1)
		geometryBuilder.Add("type", json.NullJSONValue)
		geometry = geometryBuilder.Build()
	} else {
		__antithesis_instrumentation__.Notify(601448)
	}
	__antithesis_instrumentation__.Notify(601408)
	retJSON := json.NewObjectBuilder(3)
	retJSON.Add("type", json.FromString("Feature"))
	retJSON.Add("geometry", geometry)
	retJSON.Add("properties", properties.Build())
	retString := retJSON.Build().String()
	if !pretty {
		__antithesis_instrumentation__.Notify(601449)
		return tree.NewDString(retString), nil
	} else {
		__antithesis_instrumentation__.Notify(601450)
	}
	__antithesis_instrumentation__.Notify(601409)
	var reserializedJSON map[string]interface{}
	err := gojson.Unmarshal([]byte(retString), &reserializedJSON)
	if err != nil {
		__antithesis_instrumentation__.Notify(601451)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(601452)
	}
	__antithesis_instrumentation__.Notify(601410)
	marshalledIndent, err := gojson.MarshalIndent(reserializedJSON, "", "\t")
	if err != nil {
		__antithesis_instrumentation__.Notify(601453)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(601454)
	}
	__antithesis_instrumentation__.Notify(601411)
	return tree.NewDString(string(marshalledIndent)), nil
}

func makeSTDWithinBuiltin(exclusivity geo.FnExclusivity) builtinDefinition {
	__antithesis_instrumentation__.Notify(601455)
	exclusivityStr := ", inclusive."
	if exclusivity == geo.FnExclusive {
		__antithesis_instrumentation__.Notify(601457)
		exclusivityStr = ", exclusive."
	} else {
		__antithesis_instrumentation__.Notify(601458)
	}
	__antithesis_instrumentation__.Notify(601456)
	return makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry_a", types.Geometry},
				{"geometry_b", types.Geometry},
				{"distance", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601459)
				a := tree.MustBeDGeometry(args[0])
				b := tree.MustBeDGeometry(args[1])
				dist := tree.MustBeDFloat(args[2])
				ret, err := geomfn.DWithin(a.Geometry, b.Geometry, float64(dist), exclusivity)
				if err != nil {
					__antithesis_instrumentation__.Notify(601461)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601462)
				}
				__antithesis_instrumentation__.Notify(601460)
				return tree.MakeDBool(tree.DBool(ret)), nil
			},
			Info: infoBuilder{
				info: "Returns true if any of geometry_a is within distance units of geometry_b" +
					exclusivityStr,
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geography_a", types.Geography},
				{"geography_b", types.Geography},
				{"distance", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601463)
				a := tree.MustBeDGeography(args[0])
				b := tree.MustBeDGeography(args[1])
				dist := tree.MustBeDFloat(args[2])
				ret, err := geogfn.DWithin(a.Geography, b.Geography, float64(dist), geogfn.UseSpheroid, exclusivity)
				if err != nil {
					__antithesis_instrumentation__.Notify(601465)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601466)
				}
				__antithesis_instrumentation__.Notify(601464)
				return tree.MakeDBool(tree.DBool(ret)), nil
			},
			Info: infoBuilder{
				info: "Returns true if any of geography_a is within distance meters of geography_b" +
					exclusivityStr + usesSpheroidMessage + spheroidDistanceMessage,
				libraryUsage: usesGeographicLib,
				precision:    "1cm",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"geography_a", types.Geography},
				{"geography_b", types.Geography},
				{"distance", types.Float},
				{"use_spheroid", types.Bool},
			},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601467)
				a := tree.MustBeDGeography(args[0])
				b := tree.MustBeDGeography(args[1])
				dist := tree.MustBeDFloat(args[2])
				useSpheroid := tree.MustBeDBool(args[3])

				ret, err := geogfn.DWithin(a.Geography, b.Geography, float64(dist), toUseSphereOrSpheroid(useSpheroid), exclusivity)
				if err != nil {
					__antithesis_instrumentation__.Notify(601469)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601470)
				}
				__antithesis_instrumentation__.Notify(601468)
				return tree.MakeDBool(tree.DBool(ret)), nil
			},
			Info: infoBuilder{
				info: "Returns true if any of geography_a is within distance meters of geography_b" +
					exclusivityStr + spheroidDistanceMessage,
				libraryUsage: usesGeographicLib | usesS2,
				precision:    "1cm",
			}.String(),
			Volatility: tree.VolatilityImmutable,
		},
	)
}

func applyGeoindexConfigStorageParams(
	evalCtx *tree.EvalContext, cfg geoindex.Config, params string,
) (geoindex.Config, error) {
	__antithesis_instrumentation__.Notify(601471)
	indexDesc := &descpb.IndexDescriptor{GeoConfig: cfg}
	stmt, err := parser.ParseOne(
		fmt.Sprintf("CREATE INDEX t_idx ON t USING GIST(geom) WITH (%s)", params),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(601474)
		return geoindex.Config{}, errors.Newf("invalid storage parameters specified: %s", params)
	} else {
		__antithesis_instrumentation__.Notify(601475)
	}
	__antithesis_instrumentation__.Notify(601472)
	semaCtx := tree.MakeSemaContext()
	if err := paramparse.SetStorageParameters(
		evalCtx.Context,
		&semaCtx,
		evalCtx,
		stmt.AST.(*tree.CreateIndex).StorageParams,
		&paramparse.IndexStorageParamObserver{IndexDesc: indexDesc},
	); err != nil {
		__antithesis_instrumentation__.Notify(601476)
		return geoindex.Config{}, err
	} else {
		__antithesis_instrumentation__.Notify(601477)
	}
	__antithesis_instrumentation__.Notify(601473)
	return indexDesc.GeoConfig, nil
}
