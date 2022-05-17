package colinfo

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

func CheckDatumTypeFitsColumnType(col catalog.Column, typ *types.T) error {
	__antithesis_instrumentation__.Notify(251107)
	if typ.Family() == types.UnknownFamily {
		__antithesis_instrumentation__.Notify(251110)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(251111)
	}
	__antithesis_instrumentation__.Notify(251108)
	if !typ.Equivalent(col.GetType()) {
		__antithesis_instrumentation__.Notify(251112)
		return pgerror.Newf(pgcode.DatatypeMismatch,
			"value type %s doesn't match type %s of column %q",
			typ.String(), col.GetType().String(), tree.ErrNameString(col.GetName()))
	} else {
		__antithesis_instrumentation__.Notify(251113)
	}
	__antithesis_instrumentation__.Notify(251109)
	return nil
}

func CanHaveCompositeKeyEncoding(typ *types.T) bool {
	__antithesis_instrumentation__.Notify(251114)
	switch typ.Family() {
	case types.FloatFamily,
		types.DecimalFamily,
		types.CollatedStringFamily:
		__antithesis_instrumentation__.Notify(251115)
		return true
	case types.ArrayFamily:
		__antithesis_instrumentation__.Notify(251116)
		return CanHaveCompositeKeyEncoding(typ.ArrayContents())
	case types.TupleFamily:
		__antithesis_instrumentation__.Notify(251117)
		for _, t := range typ.TupleContents() {
			__antithesis_instrumentation__.Notify(251122)
			if CanHaveCompositeKeyEncoding(t) {
				__antithesis_instrumentation__.Notify(251123)
				return true
			} else {
				__antithesis_instrumentation__.Notify(251124)
			}
		}
		__antithesis_instrumentation__.Notify(251118)
		return false
	case types.BoolFamily,
		types.IntFamily,
		types.DateFamily,
		types.TimestampFamily,
		types.IntervalFamily,
		types.StringFamily,
		types.BytesFamily,
		types.TimestampTZFamily,
		types.OidFamily,
		types.UuidFamily,
		types.INetFamily,
		types.TimeFamily,
		types.JsonFamily,
		types.TimeTZFamily,
		types.BitFamily,
		types.GeometryFamily,
		types.GeographyFamily,
		types.EnumFamily,
		types.Box2DFamily:
		__antithesis_instrumentation__.Notify(251119)
		return false
	case types.UnknownFamily,
		types.AnyFamily:
		__antithesis_instrumentation__.Notify(251120)
		fallthrough
	default:
		__antithesis_instrumentation__.Notify(251121)
		return true
	}
}
