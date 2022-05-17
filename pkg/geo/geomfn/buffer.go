package geomfn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strconv"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geos"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/errors"
)

type BufferParams struct {
	p geos.BufferParams
}

func MakeDefaultBufferParams() BufferParams {
	__antithesis_instrumentation__.Notify(61399)
	return BufferParams{
		p: geos.BufferParams{
			EndCapStyle:      geos.BufferParamsEndCapStyleRound,
			JoinStyle:        geos.BufferParamsJoinStyleRound,
			SingleSided:      false,
			QuadrantSegments: 8,
			MitreLimit:       5.0,
		},
	}
}

func (b BufferParams) WithQuadrantSegments(quadrantSegments int) BufferParams {
	__antithesis_instrumentation__.Notify(61400)
	ret := b
	ret.p.QuadrantSegments = quadrantSegments
	return ret
}

func ParseBufferParams(s string, distance float64) (BufferParams, float64, error) {
	__antithesis_instrumentation__.Notify(61401)
	p := MakeDefaultBufferParams()
	fields := strings.Fields(s)
	for _, field := range fields {
		__antithesis_instrumentation__.Notify(61403)
		fParams := strings.Split(field, "=")
		if len(fParams) != 2 {
			__antithesis_instrumentation__.Notify(61405)
			return BufferParams{}, 0, pgerror.Newf(pgcode.InvalidParameterValue, "unknown buffer parameter: %s", fParams)
		} else {
			__antithesis_instrumentation__.Notify(61406)
		}
		__antithesis_instrumentation__.Notify(61404)
		f, val := fParams[0], fParams[1]
		switch strings.ToLower(f) {
		case "quad_segs":
			__antithesis_instrumentation__.Notify(61407)
			valInt, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				__antithesis_instrumentation__.Notify(61415)
				return BufferParams{}, 0, pgerror.WithCandidateCode(
					errors.Wrapf(err, "invalid int for %s: %s", f, val),
					pgcode.InvalidParameterValue,
				)
			} else {
				__antithesis_instrumentation__.Notify(61416)
			}
			__antithesis_instrumentation__.Notify(61408)
			p.p.QuadrantSegments = int(valInt)
		case "endcap":
			__antithesis_instrumentation__.Notify(61409)
			switch strings.ToLower(val) {
			case "round":
				__antithesis_instrumentation__.Notify(61417)
				p.p.EndCapStyle = geos.BufferParamsEndCapStyleRound
			case "flat", "butt":
				__antithesis_instrumentation__.Notify(61418)
				p.p.EndCapStyle = geos.BufferParamsEndCapStyleFlat
			case "square":
				__antithesis_instrumentation__.Notify(61419)
				p.p.EndCapStyle = geos.BufferParamsEndCapStyleSquare
			default:
				__antithesis_instrumentation__.Notify(61420)
				return BufferParams{}, 0, pgerror.Newf(
					pgcode.InvalidParameterValue,
					"unknown endcap: %s (accepted: round, flat, square)",
					val,
				)
			}
		case "join":
			__antithesis_instrumentation__.Notify(61410)
			switch strings.ToLower(val) {
			case "round":
				__antithesis_instrumentation__.Notify(61421)
				p.p.JoinStyle = geos.BufferParamsJoinStyleRound
			case "mitre", "miter":
				__antithesis_instrumentation__.Notify(61422)
				p.p.JoinStyle = geos.BufferParamsJoinStyleMitre
			case "bevel":
				__antithesis_instrumentation__.Notify(61423)
				p.p.JoinStyle = geos.BufferParamsJoinStyleBevel
			default:
				__antithesis_instrumentation__.Notify(61424)
				return BufferParams{}, 0, pgerror.Newf(
					pgcode.InvalidParameterValue,
					"unknown join: %s (accepted: round, mitre, bevel)",
					val,
				)
			}
		case "mitre_limit", "miter_limit":
			__antithesis_instrumentation__.Notify(61411)
			valFloat, err := strconv.ParseFloat(val, 64)
			if err != nil {
				__antithesis_instrumentation__.Notify(61425)
				return BufferParams{}, 0, pgerror.WithCandidateCode(
					errors.Wrapf(err, "invalid float for %s: %s", f, val),
					pgcode.InvalidParameterValue,
				)
			} else {
				__antithesis_instrumentation__.Notify(61426)
			}
			__antithesis_instrumentation__.Notify(61412)
			p.p.MitreLimit = valFloat
		case "side":
			__antithesis_instrumentation__.Notify(61413)
			switch strings.ToLower(val) {
			case "both":
				__antithesis_instrumentation__.Notify(61427)
				p.p.SingleSided = false
			case "left":
				__antithesis_instrumentation__.Notify(61428)
				p.p.SingleSided = true
			case "right":
				__antithesis_instrumentation__.Notify(61429)
				p.p.SingleSided = true
				distance *= -1
			default:
				__antithesis_instrumentation__.Notify(61430)
				return BufferParams{}, 0, pgerror.Newf(pgcode.InvalidParameterValue, "unknown side: %s (accepted: both, left, right)", val)
			}
		default:
			__antithesis_instrumentation__.Notify(61414)
			return BufferParams{}, 0, pgerror.Newf(pgcode.InvalidParameterValue, "unknown field: %s (accepted fields: quad_segs, endcap, join, mitre_limit, side)", f)
		}
	}
	__antithesis_instrumentation__.Notify(61402)
	return p, distance, nil
}

func Buffer(g geo.Geometry, params BufferParams, distance float64) (geo.Geometry, error) {
	__antithesis_instrumentation__.Notify(61431)
	if params.p.QuadrantSegments < 1 {
		__antithesis_instrumentation__.Notify(61435)
		return geo.Geometry{}, pgerror.Newf(
			pgcode.InvalidParameterValue,
			"must request at least 1 quadrant segment, requested %d quadrant segments",
			params.p.QuadrantSegments,
		)

	} else {
		__antithesis_instrumentation__.Notify(61436)
	}
	__antithesis_instrumentation__.Notify(61432)
	if params.p.QuadrantSegments > geo.MaxAllowedSplitPoints {
		__antithesis_instrumentation__.Notify(61437)
		return geo.Geometry{}, pgerror.Newf(
			pgcode.InvalidParameterValue,
			"attempting to split buffered geometry into too many quadrant segments; requested %d quadrant segments, max %d",
			params.p.QuadrantSegments,
			geo.MaxAllowedSplitPoints,
		)
	} else {
		__antithesis_instrumentation__.Notify(61438)
	}
	__antithesis_instrumentation__.Notify(61433)
	bufferedGeom, err := geos.Buffer(g.EWKB(), params.p, distance)
	if err != nil {
		__antithesis_instrumentation__.Notify(61439)
		return geo.Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(61440)
	}
	__antithesis_instrumentation__.Notify(61434)
	return geo.ParseGeometryFromEWKB(bufferedGeom)
}
