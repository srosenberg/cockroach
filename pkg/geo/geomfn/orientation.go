package geomfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/twpayne/go-geom"
)

type Orientation int

const (
	OrientationCW Orientation = iota

	OrientationCCW
)

func HasPolygonOrientation(g geo.Geometry, o Orientation) (bool, error) {
	__antithesis_instrumentation__.Notify(62698)
	t, err := g.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(62700)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(62701)
	}
	__antithesis_instrumentation__.Notify(62699)
	return hasPolygonOrientation(t, o)
}

func hasPolygonOrientation(g geom.T, o Orientation) (bool, error) {
	__antithesis_instrumentation__.Notify(62702)
	switch g := g.(type) {
	case *geom.Polygon:
		__antithesis_instrumentation__.Notify(62703)
		for i := 0; i < g.NumLinearRings(); i++ {
			__antithesis_instrumentation__.Notify(62711)
			isCCW := geo.IsLinearRingCCW(g.LinearRing(i))

			if i > 0 {
				__antithesis_instrumentation__.Notify(62713)
				isCCW = !isCCW
			} else {
				__antithesis_instrumentation__.Notify(62714)
			}
			__antithesis_instrumentation__.Notify(62712)
			switch o {
			case OrientationCW:
				__antithesis_instrumentation__.Notify(62715)
				if isCCW {
					__antithesis_instrumentation__.Notify(62718)
					return false, nil
				} else {
					__antithesis_instrumentation__.Notify(62719)
				}
			case OrientationCCW:
				__antithesis_instrumentation__.Notify(62716)
				if !isCCW {
					__antithesis_instrumentation__.Notify(62720)
					return false, nil
				} else {
					__antithesis_instrumentation__.Notify(62721)
				}
			default:
				__antithesis_instrumentation__.Notify(62717)
				return false, pgerror.Newf(pgcode.InvalidParameterValue, "unexpected orientation: %v", o)
			}
		}
		__antithesis_instrumentation__.Notify(62704)
		return true, nil
	case *geom.MultiPolygon:
		__antithesis_instrumentation__.Notify(62705)
		for i := 0; i < g.NumPolygons(); i++ {
			__antithesis_instrumentation__.Notify(62722)
			if ret, err := hasPolygonOrientation(g.Polygon(i), o); !ret || func() bool {
				__antithesis_instrumentation__.Notify(62723)
				return err != nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(62724)
				return ret, err
			} else {
				__antithesis_instrumentation__.Notify(62725)
			}
		}
		__antithesis_instrumentation__.Notify(62706)
		return true, nil
	case *geom.GeometryCollection:
		__antithesis_instrumentation__.Notify(62707)
		for i := 0; i < g.NumGeoms(); i++ {
			__antithesis_instrumentation__.Notify(62726)
			if ret, err := hasPolygonOrientation(g.Geom(i), o); !ret || func() bool {
				__antithesis_instrumentation__.Notify(62727)
				return err != nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(62728)
				return ret, err
			} else {
				__antithesis_instrumentation__.Notify(62729)
			}
		}
		__antithesis_instrumentation__.Notify(62708)
		return true, nil
	case *geom.Point, *geom.MultiPoint, *geom.LineString, *geom.MultiLineString:
		__antithesis_instrumentation__.Notify(62709)
		return true, nil
	default:
		__antithesis_instrumentation__.Notify(62710)
		return false, pgerror.Newf(pgcode.InvalidParameterValue, "unhandled geometry type: %T", g)
	}
}

func ForcePolygonOrientation(g geo.Geometry, o Orientation) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(62730)
	t, err := g.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(62733)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(62734)
	}
	__antithesis_instrumentation__.Notify(62731)

	if err := forcePolygonOrientation(t, o); err != nil {
		__antithesis_instrumentation__.Notify(62735)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(62736)
	}
	__antithesis_instrumentation__.Notify(62732)
	return geo.MakeGeometryFromGeomT(t)
}

func forcePolygonOrientation(g geom.T, o Orientation) error {
	__antithesis_instrumentation__.Notify(62737)
	switch g := g.(type) {
	case *geom.Polygon:
		__antithesis_instrumentation__.Notify(62738)
		for i := 0; i < g.NumLinearRings(); i++ {
			__antithesis_instrumentation__.Notify(62746)
			isCCW := geo.IsLinearRingCCW(g.LinearRing(i))

			if i > 0 {
				__antithesis_instrumentation__.Notify(62749)
				isCCW = !isCCW
			} else {
				__antithesis_instrumentation__.Notify(62750)
			}
			__antithesis_instrumentation__.Notify(62747)
			reverse := false
			switch o {
			case OrientationCW:
				__antithesis_instrumentation__.Notify(62751)
				if isCCW {
					__antithesis_instrumentation__.Notify(62754)
					reverse = true
				} else {
					__antithesis_instrumentation__.Notify(62755)
				}
			case OrientationCCW:
				__antithesis_instrumentation__.Notify(62752)
				if !isCCW {
					__antithesis_instrumentation__.Notify(62756)
					reverse = true
				} else {
					__antithesis_instrumentation__.Notify(62757)
				}
			default:
				__antithesis_instrumentation__.Notify(62753)
				return pgerror.Newf(pgcode.InvalidParameterValue, "unexpected orientation: %v", o)
			}
			__antithesis_instrumentation__.Notify(62748)

			if reverse {
				__antithesis_instrumentation__.Notify(62758)

				coords := g.LinearRing(i).FlatCoords()
				for cIdx := 0; cIdx < len(coords)/2; cIdx += g.Stride() {
					__antithesis_instrumentation__.Notify(62759)
					for sIdx := 0; sIdx < g.Stride(); sIdx++ {
						__antithesis_instrumentation__.Notify(62760)
						coords[cIdx+sIdx], coords[len(coords)-cIdx-g.Stride()+sIdx] = coords[len(coords)-cIdx-g.Stride()+sIdx], coords[cIdx+sIdx]
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(62761)
			}
		}
		__antithesis_instrumentation__.Notify(62739)
		return nil
	case *geom.MultiPolygon:
		__antithesis_instrumentation__.Notify(62740)
		for i := 0; i < g.NumPolygons(); i++ {
			__antithesis_instrumentation__.Notify(62762)
			if err := forcePolygonOrientation(g.Polygon(i), o); err != nil {
				__antithesis_instrumentation__.Notify(62763)
				return err
			} else {
				__antithesis_instrumentation__.Notify(62764)
			}
		}
		__antithesis_instrumentation__.Notify(62741)
		return nil
	case *geom.GeometryCollection:
		__antithesis_instrumentation__.Notify(62742)
		for i := 0; i < g.NumGeoms(); i++ {
			__antithesis_instrumentation__.Notify(62765)
			if err := forcePolygonOrientation(g.Geom(i), o); err != nil {
				__antithesis_instrumentation__.Notify(62766)
				return err
			} else {
				__antithesis_instrumentation__.Notify(62767)
			}
		}
		__antithesis_instrumentation__.Notify(62743)
		return nil
	case *geom.Point, *geom.MultiPoint, *geom.LineString, *geom.MultiLineString:
		__antithesis_instrumentation__.Notify(62744)
		return nil
	default:
		__antithesis_instrumentation__.Notify(62745)
		return pgerror.Newf(pgcode.InvalidParameterValue, "unhandled geometry type: %T", g)
	}
}
