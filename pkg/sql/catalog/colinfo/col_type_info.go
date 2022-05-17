package colinfo

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"
	"golang.org/x/text/language"
)

type ColTypeInfo struct {
	resCols  ResultColumns
	colTypes []*types.T
}

func ColTypeInfoFromResCols(resCols ResultColumns) ColTypeInfo {
	__antithesis_instrumentation__.Notify(250891)
	return ColTypeInfo{resCols: resCols}
}

func ColTypeInfoFromColTypes(colTypes []*types.T) ColTypeInfo {
	__antithesis_instrumentation__.Notify(250892)
	return ColTypeInfo{colTypes: colTypes}
}

func ColTypeInfoFromColumns(columns []catalog.Column) ColTypeInfo {
	__antithesis_instrumentation__.Notify(250893)
	colTypes := make([]*types.T, len(columns))
	for i, col := range columns {
		__antithesis_instrumentation__.Notify(250895)
		colTypes[i] = col.GetType()
	}
	__antithesis_instrumentation__.Notify(250894)
	return ColTypeInfoFromColTypes(colTypes)
}

func (ti ColTypeInfo) NumColumns() int {
	__antithesis_instrumentation__.Notify(250896)
	if ti.resCols != nil {
		__antithesis_instrumentation__.Notify(250898)
		return len(ti.resCols)
	} else {
		__antithesis_instrumentation__.Notify(250899)
	}
	__antithesis_instrumentation__.Notify(250897)
	return len(ti.colTypes)
}

func (ti ColTypeInfo) Type(idx int) *types.T {
	__antithesis_instrumentation__.Notify(250900)
	if ti.resCols != nil {
		__antithesis_instrumentation__.Notify(250902)
		return ti.resCols[idx].Typ
	} else {
		__antithesis_instrumentation__.Notify(250903)
	}
	__antithesis_instrumentation__.Notify(250901)
	return ti.colTypes[idx]
}

func ValidateColumnDefType(t *types.T) error {
	__antithesis_instrumentation__.Notify(250904)
	switch t.Family() {
	case types.StringFamily, types.CollatedStringFamily:
		__antithesis_instrumentation__.Notify(250906)
		if t.Family() == types.CollatedStringFamily {
			__antithesis_instrumentation__.Notify(250914)
			if _, err := language.Parse(t.Locale()); err != nil {
				__antithesis_instrumentation__.Notify(250915)
				return pgerror.Newf(pgcode.Syntax, `invalid locale %s`, t.Locale())
			} else {
				__antithesis_instrumentation__.Notify(250916)
			}
		} else {
			__antithesis_instrumentation__.Notify(250917)
		}

	case types.DecimalFamily:
		__antithesis_instrumentation__.Notify(250907)
		switch {
		case t.Precision() == 0 && func() bool {
			__antithesis_instrumentation__.Notify(250921)
			return t.Scale() > 0 == true
		}() == true:
			__antithesis_instrumentation__.Notify(250918)

			return errors.New("invalid NUMERIC precision 0")
		case t.Precision() < t.Scale():
			__antithesis_instrumentation__.Notify(250919)
			return fmt.Errorf("NUMERIC scale %d must be between 0 and precision %d",
				t.Scale(), t.Precision())
		default:
			__antithesis_instrumentation__.Notify(250920)
		}

	case types.ArrayFamily:
		__antithesis_instrumentation__.Notify(250908)
		if t.ArrayContents().Family() == types.ArrayFamily {
			__antithesis_instrumentation__.Notify(250922)

			return errors.Errorf("nested array unsupported as column type: %s", t.String())
		} else {
			__antithesis_instrumentation__.Notify(250923)
		}
		__antithesis_instrumentation__.Notify(250909)
		if t.ArrayContents().Family() == types.JsonFamily {
			__antithesis_instrumentation__.Notify(250924)

			return unimplemented.NewWithIssueDetailf(23468, t.String(),
				"arrays of JSON unsupported as column type")
		} else {
			__antithesis_instrumentation__.Notify(250925)
		}
		__antithesis_instrumentation__.Notify(250910)
		if err := types.CheckArrayElementType(t.ArrayContents()); err != nil {
			__antithesis_instrumentation__.Notify(250926)
			return err
		} else {
			__antithesis_instrumentation__.Notify(250927)
		}
		__antithesis_instrumentation__.Notify(250911)
		return ValidateColumnDefType(t.ArrayContents())

	case types.BitFamily, types.IntFamily, types.FloatFamily, types.BoolFamily, types.BytesFamily, types.DateFamily,
		types.INetFamily, types.IntervalFamily, types.JsonFamily, types.OidFamily, types.TimeFamily,
		types.TimestampFamily, types.TimestampTZFamily, types.UuidFamily, types.TimeTZFamily,
		types.GeographyFamily, types.GeometryFamily, types.EnumFamily, types.Box2DFamily:
		__antithesis_instrumentation__.Notify(250912)

	default:
		__antithesis_instrumentation__.Notify(250913)
		return pgerror.Newf(pgcode.InvalidTableDefinition,
			"value type %s cannot be used for table columns", t.String())
	}
	__antithesis_instrumentation__.Notify(250905)

	return nil
}

func ColumnTypeIsIndexable(t *types.T) bool {
	__antithesis_instrumentation__.Notify(250928)
	if t.IsAmbiguous() || func() bool {
		__antithesis_instrumentation__.Notify(250930)
		return t.Family() == types.TupleFamily == true
	}() == true {
		__antithesis_instrumentation__.Notify(250931)
		return false
	} else {
		__antithesis_instrumentation__.Notify(250932)
	}
	__antithesis_instrumentation__.Notify(250929)

	return !MustBeValueEncoded(t) && func() bool {
		__antithesis_instrumentation__.Notify(250933)
		return !ColumnTypeIsInvertedIndexable(t) == true
	}() == true
}

func ColumnTypeIsInvertedIndexable(t *types.T) bool {
	__antithesis_instrumentation__.Notify(250934)
	if t.IsAmbiguous() || func() bool {
		__antithesis_instrumentation__.Notify(250936)
		return t.Family() == types.TupleFamily == true
	}() == true {
		__antithesis_instrumentation__.Notify(250937)
		return false
	} else {
		__antithesis_instrumentation__.Notify(250938)
	}
	__antithesis_instrumentation__.Notify(250935)
	family := t.Family()
	return family == types.JsonFamily || func() bool {
		__antithesis_instrumentation__.Notify(250939)
		return family == types.ArrayFamily == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(250940)
		return family == types.GeographyFamily == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(250941)
		return family == types.GeometryFamily == true
	}() == true
}

func MustBeValueEncoded(semanticType *types.T) bool {
	__antithesis_instrumentation__.Notify(250942)
	switch semanticType.Family() {
	case types.ArrayFamily:
		__antithesis_instrumentation__.Notify(250944)
		switch semanticType.Oid() {
		case oid.T_int2vector, oid.T_oidvector:
			__antithesis_instrumentation__.Notify(250947)
			return true
		default:
			__antithesis_instrumentation__.Notify(250948)
			return MustBeValueEncoded(semanticType.ArrayContents())
		}
	case types.JsonFamily, types.TupleFamily, types.GeographyFamily, types.GeometryFamily:
		__antithesis_instrumentation__.Notify(250945)
		return true
	default:
		__antithesis_instrumentation__.Notify(250946)
	}
	__antithesis_instrumentation__.Notify(250943)
	return false
}
