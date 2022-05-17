package cmpconn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math/big"
	"strings"

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/errors"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jackc/pgtype"
)

func CompareVals(a, b []interface{}) error {
	__antithesis_instrumentation__.Notify(37973)
	if len(a) != len(b) {
		__antithesis_instrumentation__.Notify(37977)
		return errors.Errorf("size difference: %d != %d", len(a), len(b))
	} else {
		__antithesis_instrumentation__.Notify(37978)
	}
	__antithesis_instrumentation__.Notify(37974)
	if len(a) == 0 {
		__antithesis_instrumentation__.Notify(37979)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(37980)
	}
	__antithesis_instrumentation__.Notify(37975)
	if diff := cmp.Diff(a, b, cmpOptions...); diff != "" {
		__antithesis_instrumentation__.Notify(37981)
		return errors.Newf("unexpected diff:\n%s", diff)
	} else {
		__antithesis_instrumentation__.Notify(37982)
	}
	__antithesis_instrumentation__.Notify(37976)
	return nil
}

var (
	cmpOptions = []cmp.Option{
		cmp.Transformer("", func(x []interface{}) []interface{} {
			__antithesis_instrumentation__.Notify(37983)
			out := make([]interface{}, len(x))
			for i, v := range x {
				__antithesis_instrumentation__.Notify(37985)
				switch t := v.(type) {
				case pgtype.TextArray:
					__antithesis_instrumentation__.Notify(37987)
					if t.Status == pgtype.Present && func() bool {
						__antithesis_instrumentation__.Notify(38004)
						return len(t.Elements) == 0 == true
					}() == true {
						__antithesis_instrumentation__.Notify(38005)
						v = ""
					} else {
						__antithesis_instrumentation__.Notify(38006)
					}
				case pgtype.BPCharArray:
					__antithesis_instrumentation__.Notify(37988)
					if t.Status == pgtype.Present && func() bool {
						__antithesis_instrumentation__.Notify(38007)
						return len(t.Elements) == 0 == true
					}() == true {
						__antithesis_instrumentation__.Notify(38008)
						v = ""
					} else {
						__antithesis_instrumentation__.Notify(38009)
					}
				case pgtype.VarcharArray:
					__antithesis_instrumentation__.Notify(37989)
					if t.Status == pgtype.Present && func() bool {
						__antithesis_instrumentation__.Notify(38010)
						return len(t.Elements) == 0 == true
					}() == true {
						__antithesis_instrumentation__.Notify(38011)
						v = ""
					} else {
						__antithesis_instrumentation__.Notify(38012)
					}
				case pgtype.Int8Array:
					__antithesis_instrumentation__.Notify(37990)
					if t.Status == pgtype.Present && func() bool {
						__antithesis_instrumentation__.Notify(38013)
						return len(t.Elements) == 0 == true
					}() == true {
						__antithesis_instrumentation__.Notify(38014)
						v = &pgtype.Int8Array{}
					} else {
						__antithesis_instrumentation__.Notify(38015)
					}
				case pgtype.Float8Array:
					__antithesis_instrumentation__.Notify(37991)
					if t.Status == pgtype.Present && func() bool {
						__antithesis_instrumentation__.Notify(38016)
						return len(t.Elements) == 0 == true
					}() == true {
						__antithesis_instrumentation__.Notify(38017)
						v = &pgtype.Float8Array{}
					} else {
						__antithesis_instrumentation__.Notify(38018)
					}
				case pgtype.UUIDArray:
					__antithesis_instrumentation__.Notify(37992)
					if t.Status == pgtype.Present && func() bool {
						__antithesis_instrumentation__.Notify(38019)
						return len(t.Elements) == 0 == true
					}() == true {
						__antithesis_instrumentation__.Notify(38020)
						v = &pgtype.UUIDArray{}
					} else {
						__antithesis_instrumentation__.Notify(38021)
					}
				case pgtype.ByteaArray:
					__antithesis_instrumentation__.Notify(37993)
					if t.Status == pgtype.Present && func() bool {
						__antithesis_instrumentation__.Notify(38022)
						return len(t.Elements) == 0 == true
					}() == true {
						__antithesis_instrumentation__.Notify(38023)
						v = &pgtype.ByteaArray{}
					} else {
						__antithesis_instrumentation__.Notify(38024)
					}
				case pgtype.InetArray:
					__antithesis_instrumentation__.Notify(37994)
					if t.Status == pgtype.Present && func() bool {
						__antithesis_instrumentation__.Notify(38025)
						return len(t.Elements) == 0 == true
					}() == true {
						__antithesis_instrumentation__.Notify(38026)
						v = &pgtype.InetArray{}
					} else {
						__antithesis_instrumentation__.Notify(38027)
					}
				case pgtype.TimestampArray:
					__antithesis_instrumentation__.Notify(37995)
					if t.Status == pgtype.Present && func() bool {
						__antithesis_instrumentation__.Notify(38028)
						return len(t.Elements) == 0 == true
					}() == true {
						__antithesis_instrumentation__.Notify(38029)
						v = &pgtype.TimestampArray{}
					} else {
						__antithesis_instrumentation__.Notify(38030)
					}
				case pgtype.BoolArray:
					__antithesis_instrumentation__.Notify(37996)
					if t.Status == pgtype.Present && func() bool {
						__antithesis_instrumentation__.Notify(38031)
						return len(t.Elements) == 0 == true
					}() == true {
						__antithesis_instrumentation__.Notify(38032)
						v = &pgtype.BoolArray{}
					} else {
						__antithesis_instrumentation__.Notify(38033)
					}
				case pgtype.DateArray:
					__antithesis_instrumentation__.Notify(37997)
					if t.Status == pgtype.Present && func() bool {
						__antithesis_instrumentation__.Notify(38034)
						return len(t.Elements) == 0 == true
					}() == true {
						__antithesis_instrumentation__.Notify(38035)
						v = &pgtype.BoolArray{}
					} else {
						__antithesis_instrumentation__.Notify(38036)
					}
				case pgtype.Varbit:
					__antithesis_instrumentation__.Notify(37998)
					if t.Status == pgtype.Present {
						__antithesis_instrumentation__.Notify(38037)
						s, _ := t.EncodeText(nil, nil)
						v = string(s)
					} else {
						__antithesis_instrumentation__.Notify(38038)
					}
				case pgtype.Bit:
					__antithesis_instrumentation__.Notify(37999)
					vb := pgtype.Varbit(t)
					v = &vb
				case pgtype.Interval:
					__antithesis_instrumentation__.Notify(38000)
					if t.Status == pgtype.Present {
						__antithesis_instrumentation__.Notify(38039)
						v = duration.DecodeDuration(int64(t.Months), int64(t.Days), t.Microseconds*1000)
					} else {
						__antithesis_instrumentation__.Notify(38040)
					}
				case string:
					__antithesis_instrumentation__.Notify(38001)

					t = strings.TrimSpace(t)
					v = strings.Replace(t, "T00:00:00+00:00", "T00:00:00Z", 1)

					v = strings.Replace(t, ":00+00:00", ":00+00", 1)
				case pgtype.Numeric:
					__antithesis_instrumentation__.Notify(38002)
					if t.Status == pgtype.Present {
						__antithesis_instrumentation__.Notify(38041)
						if t.NaN {
							__antithesis_instrumentation__.Notify(38042)
							v = &apd.Decimal{Form: apd.NaN}
						} else {
							__antithesis_instrumentation__.Notify(38043)
							var coeff apd.BigInt
							coeff.SetMathBigInt(t.Int)
							v = apd.NewWithBigInt(&coeff, t.Exp)
						}
					} else {
						__antithesis_instrumentation__.Notify(38044)
					}
				case int64:
					__antithesis_instrumentation__.Notify(38003)
					v = apd.New(t, 0)
				}
				__antithesis_instrumentation__.Notify(37986)
				out[i] = v
			}
			__antithesis_instrumentation__.Notify(37984)
			return out
		}),

		cmpopts.EquateEmpty(),
		cmpopts.EquateNaNs(),
		cmpopts.EquateApprox(0.00001, 0),
		cmp.Comparer(func(x, y *big.Int) bool {
			__antithesis_instrumentation__.Notify(38045)
			return x.Cmp(y) == 0
		}),
		cmp.Comparer(func(x, y *apd.Decimal) bool {
			__antithesis_instrumentation__.Notify(38046)
			var a, b, min, sub apd.Decimal
			a.Abs(x)
			b.Abs(y)
			if a.Cmp(&b) > 0 {
				__antithesis_instrumentation__.Notify(38048)
				min.Set(&b)
			} else {
				__antithesis_instrumentation__.Notify(38049)
				min.Set(&a)
			}
			__antithesis_instrumentation__.Notify(38047)
			ctx := tree.DecimalCtx
			_, _ = ctx.Mul(&min, &min, decimalCloseness)
			_, _ = ctx.Sub(&sub, x, y)
			sub.Abs(&sub)
			return sub.Cmp(&min) <= 0
		}),
		cmp.Comparer(func(x, y duration.Duration) bool {
			__antithesis_instrumentation__.Notify(38050)
			return x.Compare(y) == 0
		}),
	}
	decimalCloseness = apd.New(1, -6)
)
