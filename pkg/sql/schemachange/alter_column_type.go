package schemachange

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

type ColumnConversionKind int

const (
	ColumnConversionDangerous ColumnConversionKind = iota - 1

	ColumnConversionImpossible

	ColumnConversionTrivial

	ColumnConversionValidate

	ColumnConversionGeneral
)

func (i ColumnConversionKind) classifier() classifier {
	__antithesis_instrumentation__.Notify(578240)
	return func(_ *types.T, _ *types.T) ColumnConversionKind {
		__antithesis_instrumentation__.Notify(578241)
		return i
	}
}

type classifier func(oldType *types.T, newType *types.T) ColumnConversionKind

var classifiers = map[types.Family]map[types.Family]classifier{
	types.BytesFamily: {
		types.BytesFamily:  classifierWidth,
		types.StringFamily: ColumnConversionValidate.classifier(),
		types.UuidFamily:   ColumnConversionValidate.classifier(),
	},
	types.DecimalFamily: {

		types.DecimalFamily: classifierHardestOf(classifierDecimalPrecision, classifierWidth),
	},
	types.FloatFamily: {

		types.FloatFamily: ColumnConversionTrivial.classifier(),
	},
	types.IntFamily: {
		types.IntFamily: func(from *types.T, to *types.T) ColumnConversionKind {
			__antithesis_instrumentation__.Notify(578242)
			return classifierWidth(from, to)
		},
	},
	types.BitFamily: {
		types.BitFamily: func(from *types.T, to *types.T) ColumnConversionKind {
			__antithesis_instrumentation__.Notify(578243)
			return classifierWidth(from, to)
		},
	},
	types.StringFamily: {
		types.BytesFamily: func(s *types.T, b *types.T) ColumnConversionKind {
			__antithesis_instrumentation__.Notify(578244)
			return ColumnConversionValidate
		},
		types.StringFamily: classifierWidth,
	},
	types.TimestampFamily: {
		types.TimestampTZFamily: classifierTimePrecision,
		types.TimestampFamily:   classifierTimePrecision,
	},
	types.TimestampTZFamily: {
		types.TimestampFamily:   classifierTimePrecision,
		types.TimestampTZFamily: classifierTimePrecision,
	},
	types.TimeFamily: {
		types.TimeFamily: classifierTimePrecision,
	},
	types.TimeTZFamily: {
		types.TimeTZFamily: classifierTimePrecision,
	},
}

func classifierHardestOf(classifiers ...classifier) classifier {
	__antithesis_instrumentation__.Notify(578245)
	return func(oldType *types.T, newType *types.T) ColumnConversionKind {
		__antithesis_instrumentation__.Notify(578246)
		ret := ColumnConversionTrivial

		for _, c := range classifiers {
			__antithesis_instrumentation__.Notify(578248)
			next := c(oldType, newType)
			switch {
			case next == ColumnConversionImpossible:
				__antithesis_instrumentation__.Notify(578249)
				return ColumnConversionImpossible
			case next > ret:
				__antithesis_instrumentation__.Notify(578250)
				ret = next
			default:
				__antithesis_instrumentation__.Notify(578251)
			}
		}
		__antithesis_instrumentation__.Notify(578247)

		return ret
	}
}

func classifierTimePrecision(oldType *types.T, newType *types.T) ColumnConversionKind {
	__antithesis_instrumentation__.Notify(578252)
	oldPrecision := oldType.Precision()
	newPrecision := newType.Precision()

	switch {
	case newPrecision >= oldPrecision:
		__antithesis_instrumentation__.Notify(578253)
		return ColumnConversionTrivial
	default:
		__antithesis_instrumentation__.Notify(578254)
		return ColumnConversionGeneral
	}
}

func classifierDecimalPrecision(oldType *types.T, newType *types.T) ColumnConversionKind {
	__antithesis_instrumentation__.Notify(578255)
	oldPrecision := oldType.Precision()
	newPrecision := newType.Precision()

	switch {
	case oldPrecision == newPrecision:
		__antithesis_instrumentation__.Notify(578256)
		return ColumnConversionTrivial
	case oldPrecision == 0:
		__antithesis_instrumentation__.Notify(578257)
		return ColumnConversionValidate
	case newPrecision == 0 || func() bool {
		__antithesis_instrumentation__.Notify(578260)
		return newPrecision > oldPrecision == true
	}() == true:
		__antithesis_instrumentation__.Notify(578258)
		return ColumnConversionTrivial
	default:
		__antithesis_instrumentation__.Notify(578259)
		return ColumnConversionValidate
	}
}

func classifierWidth(oldType *types.T, newType *types.T) ColumnConversionKind {
	__antithesis_instrumentation__.Notify(578261)
	switch {
	case oldType.Width() == newType.Width():
		__antithesis_instrumentation__.Notify(578262)
		return ColumnConversionTrivial
	case oldType.Width() == 0 && func() bool {
		__antithesis_instrumentation__.Notify(578266)
		return newType.Width() < 64 == true
	}() == true:
		__antithesis_instrumentation__.Notify(578263)
		return ColumnConversionValidate
	case newType.Width() == 0 || func() bool {
		__antithesis_instrumentation__.Notify(578267)
		return newType.Width() > oldType.Width() == true
	}() == true:
		__antithesis_instrumentation__.Notify(578264)
		return ColumnConversionTrivial
	default:
		__antithesis_instrumentation__.Notify(578265)
		return ColumnConversionValidate
	}
}

func ClassifyConversion(
	ctx context.Context, oldType *types.T, newType *types.T,
) (ColumnConversionKind, error) {
	__antithesis_instrumentation__.Notify(578268)
	if oldType.Identical(newType) {
		__antithesis_instrumentation__.Notify(578274)
		return ColumnConversionTrivial, nil
	} else {
		__antithesis_instrumentation__.Notify(578275)
	}
	__antithesis_instrumentation__.Notify(578269)

	if mid, ok := classifiers[oldType.Family()]; ok {
		__antithesis_instrumentation__.Notify(578276)
		if fn, ok := mid[newType.Family()]; ok {
			__antithesis_instrumentation__.Notify(578277)
			ret := fn(oldType, newType)
			if ret != ColumnConversionImpossible {
				__antithesis_instrumentation__.Notify(578278)
				return ret, nil
			} else {
				__antithesis_instrumentation__.Notify(578279)
			}
		} else {
			__antithesis_instrumentation__.Notify(578280)
		}
	} else {
		__antithesis_instrumentation__.Notify(578281)
	}
	__antithesis_instrumentation__.Notify(578270)

	semaCtx := tree.MakeSemaContext()
	if err := semaCtx.Placeholders.Init(1, nil); err != nil {
		__antithesis_instrumentation__.Notify(578282)
		return ColumnConversionImpossible, err
	} else {
		__antithesis_instrumentation__.Notify(578283)
	}
	__antithesis_instrumentation__.Notify(578271)

	fromPlaceholder, err := (&tree.Placeholder{Idx: 0}).TypeCheck(ctx, &semaCtx, oldType)
	if err != nil {
		__antithesis_instrumentation__.Notify(578284)
		return ColumnConversionImpossible, err
	} else {
		__antithesis_instrumentation__.Notify(578285)
	}
	__antithesis_instrumentation__.Notify(578272)

	cast := tree.NewTypedCastExpr(fromPlaceholder, newType)
	if _, err := cast.TypeCheck(ctx, &semaCtx, nil); err == nil {
		__antithesis_instrumentation__.Notify(578286)
		return ColumnConversionGeneral, nil
	} else {
		__antithesis_instrumentation__.Notify(578287)
	}
	__antithesis_instrumentation__.Notify(578273)

	return ColumnConversionImpossible,
		pgerror.Newf(pgcode.CannotCoerce, "cannot convert %s to %s", oldType.SQLString(), newType.SQLString())
}
