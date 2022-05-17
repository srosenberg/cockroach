package paramparse

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/geo/geoindex"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgnotice"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/errors"
)

func SetStorageParameters(
	ctx context.Context,
	semaCtx *tree.SemaContext,
	evalCtx *tree.EvalContext,
	params tree.StorageParams,
	paramObserver StorageParamObserver,
) error {
	__antithesis_instrumentation__.Notify(551933)
	for _, sp := range params {
		__antithesis_instrumentation__.Notify(551935)
		key := string(sp.Key)
		if sp.Value == nil {
			__antithesis_instrumentation__.Notify(551940)
			return pgerror.Newf(pgcode.InvalidParameterValue, "storage parameter %q requires a value", key)
		} else {
			__antithesis_instrumentation__.Notify(551941)
		}
		__antithesis_instrumentation__.Notify(551936)
		telemetry.Inc(sqltelemetry.SetTableStorageParameter(key))

		expr := UnresolvedNameToStrVal(sp.Value)

		typedExpr, err := tree.TypeCheck(ctx, expr, semaCtx, types.Any)
		if err != nil {
			__antithesis_instrumentation__.Notify(551942)
			return err
		} else {
			__antithesis_instrumentation__.Notify(551943)
		}
		__antithesis_instrumentation__.Notify(551937)
		if typedExpr, err = evalCtx.NormalizeExpr(typedExpr); err != nil {
			__antithesis_instrumentation__.Notify(551944)
			return err
		} else {
			__antithesis_instrumentation__.Notify(551945)
		}
		__antithesis_instrumentation__.Notify(551938)
		datum, err := typedExpr.Eval(evalCtx)
		if err != nil {
			__antithesis_instrumentation__.Notify(551946)
			return err
		} else {
			__antithesis_instrumentation__.Notify(551947)
		}
		__antithesis_instrumentation__.Notify(551939)

		if err := paramObserver.onSet(ctx, semaCtx, evalCtx, key, datum); err != nil {
			__antithesis_instrumentation__.Notify(551948)
			return err
		} else {
			__antithesis_instrumentation__.Notify(551949)
		}
	}
	__antithesis_instrumentation__.Notify(551934)
	return paramObserver.runPostChecks()
}

func ResetStorageParameters(
	ctx context.Context,
	evalCtx *tree.EvalContext,
	params tree.NameList,
	paramObserver StorageParamObserver,
) error {
	__antithesis_instrumentation__.Notify(551950)
	for _, p := range params {
		__antithesis_instrumentation__.Notify(551952)
		telemetry.Inc(sqltelemetry.ResetTableStorageParameter(string(p)))
		if err := paramObserver.onReset(evalCtx, string(p)); err != nil {
			__antithesis_instrumentation__.Notify(551953)
			return err
		} else {
			__antithesis_instrumentation__.Notify(551954)
		}
	}
	__antithesis_instrumentation__.Notify(551951)
	return paramObserver.runPostChecks()
}

type StorageParamObserver interface {
	onSet(ctx context.Context, semaCtx *tree.SemaContext, evalCtx *tree.EvalContext, key string, datum tree.Datum) error

	onReset(evalCtx *tree.EvalContext, key string) error

	runPostChecks() error
}

type TableStorageParamObserver struct {
	tableDesc          *tabledesc.Mutable
	setAutomaticColumn bool
}

var _ StorageParamObserver = (*TableStorageParamObserver)(nil)

func NewTableStorageParamObserver(tableDesc *tabledesc.Mutable) *TableStorageParamObserver {
	__antithesis_instrumentation__.Notify(551955)
	return &TableStorageParamObserver{tableDesc: tableDesc}
}

func (po *TableStorageParamObserver) runPostChecks() error {
	__antithesis_instrumentation__.Notify(551956)
	ttl := po.tableDesc.GetRowLevelTTL()
	if po.setAutomaticColumn && func() bool {
		__antithesis_instrumentation__.Notify(551959)
		return (ttl == nil || func() bool {
			__antithesis_instrumentation__.Notify(551960)
			return ttl.DurationExpr == "" == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(551961)
		return pgerror.Newf(
			pgcode.InvalidParameterValue,
			`"ttl_expire_after" must be set if "ttl_automatic_column" is set`,
		)
	} else {
		__antithesis_instrumentation__.Notify(551962)
	}
	__antithesis_instrumentation__.Notify(551957)
	if err := tabledesc.ValidateRowLevelTTL(ttl); err != nil {
		__antithesis_instrumentation__.Notify(551963)
		return err
	} else {
		__antithesis_instrumentation__.Notify(551964)
	}
	__antithesis_instrumentation__.Notify(551958)
	return nil
}

func boolFromDatum(evalCtx *tree.EvalContext, key string, datum tree.Datum) (bool, error) {
	__antithesis_instrumentation__.Notify(551965)
	if stringVal, err := DatumAsString(evalCtx, key, datum); err == nil {
		__antithesis_instrumentation__.Notify(551968)
		return ParseBoolVar(key, stringVal)
	} else {
		__antithesis_instrumentation__.Notify(551969)
	}
	__antithesis_instrumentation__.Notify(551966)
	s, err := GetSingleBool(key, datum)
	if err != nil {
		__antithesis_instrumentation__.Notify(551970)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(551971)
	}
	__antithesis_instrumentation__.Notify(551967)
	return bool(*s), nil
}

type tableParam struct {
	onSet   func(ctx context.Context, po *TableStorageParamObserver, semaCtx *tree.SemaContext, evalCtx *tree.EvalContext, key string, datum tree.Datum) error
	onReset func(po *TableStorageParamObserver, evalCtx *tree.EvalContext, key string) error
}

var tableParams = map[string]tableParam{
	`fillfactor`: {
		onSet: func(ctx context.Context, po *TableStorageParamObserver, semaCtx *tree.SemaContext, evalCtx *tree.EvalContext, key string, datum tree.Datum) error {
			__antithesis_instrumentation__.Notify(551972)
			return setFillFactorStorageParam(evalCtx, key, datum)
		},
		onReset: func(po *TableStorageParamObserver, evalCtx *tree.EvalContext, key string) error {
			__antithesis_instrumentation__.Notify(551973)

			return nil
		},
	},
	`autovacuum_enabled`: {
		onSet: func(ctx context.Context, po *TableStorageParamObserver, semaCtx *tree.SemaContext, evalCtx *tree.EvalContext, key string, datum tree.Datum) error {
			__antithesis_instrumentation__.Notify(551974)
			var boolVal bool
			if stringVal, err := DatumAsString(evalCtx, key, datum); err == nil {
				__antithesis_instrumentation__.Notify(551977)
				boolVal, err = ParseBoolVar(key, stringVal)
				if err != nil {
					__antithesis_instrumentation__.Notify(551978)
					return err
				} else {
					__antithesis_instrumentation__.Notify(551979)
				}
			} else {
				__antithesis_instrumentation__.Notify(551980)
				s, err := GetSingleBool(key, datum)
				if err != nil {
					__antithesis_instrumentation__.Notify(551982)
					return err
				} else {
					__antithesis_instrumentation__.Notify(551983)
				}
				__antithesis_instrumentation__.Notify(551981)
				boolVal = bool(*s)
			}
			__antithesis_instrumentation__.Notify(551975)
			if !boolVal && func() bool {
				__antithesis_instrumentation__.Notify(551984)
				return evalCtx != nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(551985)
				evalCtx.ClientNoticeSender.BufferClientNotice(
					evalCtx.Context,
					pgnotice.Newf(`storage parameter "%s = %s" is ignored`, key, datum.String()),
				)
			} else {
				__antithesis_instrumentation__.Notify(551986)
			}
			__antithesis_instrumentation__.Notify(551976)
			return nil
		},
		onReset: func(po *TableStorageParamObserver, evalCtx *tree.EvalContext, key string) error {
			__antithesis_instrumentation__.Notify(551987)

			return nil
		},
	},
	`ttl`: {
		onSet: func(ctx context.Context, po *TableStorageParamObserver, semaCtx *tree.SemaContext, evalCtx *tree.EvalContext, key string, datum tree.Datum) error {
			__antithesis_instrumentation__.Notify(551988)
			setTrue, err := boolFromDatum(evalCtx, key, datum)
			if err != nil {
				__antithesis_instrumentation__.Notify(551992)
				return err
			} else {
				__antithesis_instrumentation__.Notify(551993)
			}
			__antithesis_instrumentation__.Notify(551989)
			if setTrue && func() bool {
				__antithesis_instrumentation__.Notify(551994)
				return po.tableDesc.RowLevelTTL == nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(551995)

				po.tableDesc.RowLevelTTL = &catpb.RowLevelTTL{}
			} else {
				__antithesis_instrumentation__.Notify(551996)
			}
			__antithesis_instrumentation__.Notify(551990)
			if !setTrue && func() bool {
				__antithesis_instrumentation__.Notify(551997)
				return po.tableDesc.RowLevelTTL != nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(551998)
				po.tableDesc.RowLevelTTL = nil
			} else {
				__antithesis_instrumentation__.Notify(551999)
			}
			__antithesis_instrumentation__.Notify(551991)
			return nil
		},
		onReset: func(po *TableStorageParamObserver, evalCtx *tree.EvalContext, key string) error {
			__antithesis_instrumentation__.Notify(552000)
			po.tableDesc.RowLevelTTL = nil
			return nil
		},
	},
	`ttl_automatic_column`: {
		onSet: func(ctx context.Context, po *TableStorageParamObserver, semaCtx *tree.SemaContext, evalCtx *tree.EvalContext, key string, datum tree.Datum) error {
			__antithesis_instrumentation__.Notify(552001)
			setTrue, err := boolFromDatum(evalCtx, key, datum)
			if err != nil {
				__antithesis_instrumentation__.Notify(552005)
				return err
			} else {
				__antithesis_instrumentation__.Notify(552006)
			}
			__antithesis_instrumentation__.Notify(552002)
			if setTrue {
				__antithesis_instrumentation__.Notify(552007)
				po.setAutomaticColumn = true
			} else {
				__antithesis_instrumentation__.Notify(552008)
			}
			__antithesis_instrumentation__.Notify(552003)
			if !setTrue && func() bool {
				__antithesis_instrumentation__.Notify(552009)
				return po.tableDesc.RowLevelTTL != nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(552010)
				return unimplemented.NewWithIssue(76916, "unsetting TTL automatic column not yet implemented")
			} else {
				__antithesis_instrumentation__.Notify(552011)
			}
			__antithesis_instrumentation__.Notify(552004)
			return nil
		},
		onReset: func(po *TableStorageParamObserver, evalCtx *tree.EvalContext, key string) error {
			__antithesis_instrumentation__.Notify(552012)
			return unimplemented.NewWithIssue(76916, "unsetting TTL automatic column not yet implemented")
		},
	},
	`ttl_expire_after`: {
		onSet: func(ctx context.Context, po *TableStorageParamObserver, semaCtx *tree.SemaContext, evalCtx *tree.EvalContext, key string, datum tree.Datum) error {
			__antithesis_instrumentation__.Notify(552013)
			var d *tree.DInterval
			if stringVal, err := DatumAsString(evalCtx, key, datum); err == nil {
				__antithesis_instrumentation__.Notify(552017)
				d, err = tree.ParseDInterval(evalCtx.SessionData().GetIntervalStyle(), stringVal)
				if err != nil || func() bool {
					__antithesis_instrumentation__.Notify(552018)
					return d == nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(552019)
					return pgerror.Newf(
						pgcode.InvalidParameterValue,
						`value of "ttl_expire_after" must be an interval`,
					)
				} else {
					__antithesis_instrumentation__.Notify(552020)
				}
			} else {
				__antithesis_instrumentation__.Notify(552021)
				var ok bool
				d, ok = datum.(*tree.DInterval)
				if !ok || func() bool {
					__antithesis_instrumentation__.Notify(552022)
					return d == nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(552023)
					return pgerror.Newf(
						pgcode.InvalidParameterValue,
						`value of "%s" must be an interval`,
						key,
					)
				} else {
					__antithesis_instrumentation__.Notify(552024)
				}
			}
			__antithesis_instrumentation__.Notify(552014)

			if d.Duration.Compare(duration.MakeDuration(0, 0, 0)) < 0 {
				__antithesis_instrumentation__.Notify(552025)
				return pgerror.Newf(
					pgcode.InvalidParameterValue,
					`value of "%s" must be at least zero`,
					key,
				)
			} else {
				__antithesis_instrumentation__.Notify(552026)
			}
			__antithesis_instrumentation__.Notify(552015)
			if po.tableDesc.RowLevelTTL == nil {
				__antithesis_instrumentation__.Notify(552027)
				po.tableDesc.RowLevelTTL = &catpb.RowLevelTTL{}
			} else {
				__antithesis_instrumentation__.Notify(552028)
			}
			__antithesis_instrumentation__.Notify(552016)
			po.tableDesc.RowLevelTTL.DurationExpr = catpb.Expression(tree.Serialize(d))
			return nil
		},
		onReset: func(po *TableStorageParamObserver, evalCtx *tree.EvalContext, key string) error {
			__antithesis_instrumentation__.Notify(552029)
			return errors.WithHintf(
				pgerror.Newf(
					pgcode.InvalidParameterValue,
					`resetting "ttl_expire_after" is not permitted`,
				),
				"use `RESET (ttl)` to remove TTL from the table",
			)
		},
	},
	`ttl_select_batch_size`: {
		onSet: func(ctx context.Context, po *TableStorageParamObserver, semaCtx *tree.SemaContext, evalCtx *tree.EvalContext, key string, datum tree.Datum) error {
			__antithesis_instrumentation__.Notify(552030)
			if po.tableDesc.RowLevelTTL == nil {
				__antithesis_instrumentation__.Notify(552034)
				po.tableDesc.RowLevelTTL = &catpb.RowLevelTTL{}
			} else {
				__antithesis_instrumentation__.Notify(552035)
			}
			__antithesis_instrumentation__.Notify(552031)
			val, err := DatumAsInt(evalCtx, key, datum)
			if err != nil {
				__antithesis_instrumentation__.Notify(552036)
				return err
			} else {
				__antithesis_instrumentation__.Notify(552037)
			}
			__antithesis_instrumentation__.Notify(552032)
			if err := tabledesc.ValidateTTLBatchSize(key, val); err != nil {
				__antithesis_instrumentation__.Notify(552038)
				return err
			} else {
				__antithesis_instrumentation__.Notify(552039)
			}
			__antithesis_instrumentation__.Notify(552033)
			po.tableDesc.RowLevelTTL.SelectBatchSize = val
			return nil
		},
		onReset: func(po *TableStorageParamObserver, evalCtx *tree.EvalContext, key string) error {
			__antithesis_instrumentation__.Notify(552040)
			if po.tableDesc.RowLevelTTL != nil {
				__antithesis_instrumentation__.Notify(552042)
				po.tableDesc.RowLevelTTL.SelectBatchSize = 0
			} else {
				__antithesis_instrumentation__.Notify(552043)
			}
			__antithesis_instrumentation__.Notify(552041)
			return nil
		},
	},
	`ttl_delete_batch_size`: {
		onSet: func(ctx context.Context, po *TableStorageParamObserver, semaCtx *tree.SemaContext, evalCtx *tree.EvalContext, key string, datum tree.Datum) error {
			__antithesis_instrumentation__.Notify(552044)
			if po.tableDesc.RowLevelTTL == nil {
				__antithesis_instrumentation__.Notify(552048)
				po.tableDesc.RowLevelTTL = &catpb.RowLevelTTL{}
			} else {
				__antithesis_instrumentation__.Notify(552049)
			}
			__antithesis_instrumentation__.Notify(552045)
			val, err := DatumAsInt(evalCtx, key, datum)
			if err != nil {
				__antithesis_instrumentation__.Notify(552050)
				return err
			} else {
				__antithesis_instrumentation__.Notify(552051)
			}
			__antithesis_instrumentation__.Notify(552046)
			if err := tabledesc.ValidateTTLBatchSize(key, val); err != nil {
				__antithesis_instrumentation__.Notify(552052)
				return err
			} else {
				__antithesis_instrumentation__.Notify(552053)
			}
			__antithesis_instrumentation__.Notify(552047)
			po.tableDesc.RowLevelTTL.DeleteBatchSize = val
			return nil
		},
		onReset: func(po *TableStorageParamObserver, evalCtx *tree.EvalContext, key string) error {
			__antithesis_instrumentation__.Notify(552054)
			if po.tableDesc.RowLevelTTL != nil {
				__antithesis_instrumentation__.Notify(552056)
				po.tableDesc.RowLevelTTL.DeleteBatchSize = 0
			} else {
				__antithesis_instrumentation__.Notify(552057)
			}
			__antithesis_instrumentation__.Notify(552055)
			return nil
		},
	},
	`ttl_range_concurrency`: {
		onSet: func(ctx context.Context, po *TableStorageParamObserver, semaCtx *tree.SemaContext, evalCtx *tree.EvalContext, key string, datum tree.Datum) error {
			__antithesis_instrumentation__.Notify(552058)
			if po.tableDesc.RowLevelTTL == nil {
				__antithesis_instrumentation__.Notify(552062)
				po.tableDesc.RowLevelTTL = &catpb.RowLevelTTL{}
			} else {
				__antithesis_instrumentation__.Notify(552063)
			}
			__antithesis_instrumentation__.Notify(552059)
			val, err := DatumAsInt(evalCtx, key, datum)
			if err != nil {
				__antithesis_instrumentation__.Notify(552064)
				return err
			} else {
				__antithesis_instrumentation__.Notify(552065)
			}
			__antithesis_instrumentation__.Notify(552060)
			if err := tabledesc.ValidateTTLRangeConcurrency(key, val); err != nil {
				__antithesis_instrumentation__.Notify(552066)
				return err
			} else {
				__antithesis_instrumentation__.Notify(552067)
			}
			__antithesis_instrumentation__.Notify(552061)
			po.tableDesc.RowLevelTTL.RangeConcurrency = val
			return nil
		},
		onReset: func(po *TableStorageParamObserver, evalCtx *tree.EvalContext, key string) error {
			__antithesis_instrumentation__.Notify(552068)
			if po.tableDesc.RowLevelTTL != nil {
				__antithesis_instrumentation__.Notify(552070)
				po.tableDesc.RowLevelTTL.RangeConcurrency = 0
			} else {
				__antithesis_instrumentation__.Notify(552071)
			}
			__antithesis_instrumentation__.Notify(552069)
			return nil
		},
	},
	`ttl_delete_rate_limit`: {
		onSet: func(ctx context.Context, po *TableStorageParamObserver, semaCtx *tree.SemaContext, evalCtx *tree.EvalContext, key string, datum tree.Datum) error {
			__antithesis_instrumentation__.Notify(552072)
			if po.tableDesc.RowLevelTTL == nil {
				__antithesis_instrumentation__.Notify(552076)
				po.tableDesc.RowLevelTTL = &catpb.RowLevelTTL{}
			} else {
				__antithesis_instrumentation__.Notify(552077)
			}
			__antithesis_instrumentation__.Notify(552073)
			val, err := DatumAsInt(evalCtx, key, datum)
			if err != nil {
				__antithesis_instrumentation__.Notify(552078)
				return err
			} else {
				__antithesis_instrumentation__.Notify(552079)
			}
			__antithesis_instrumentation__.Notify(552074)
			if err := tabledesc.ValidateTTLRateLimit(key, val); err != nil {
				__antithesis_instrumentation__.Notify(552080)
				return err
			} else {
				__antithesis_instrumentation__.Notify(552081)
			}
			__antithesis_instrumentation__.Notify(552075)
			po.tableDesc.RowLevelTTL.DeleteRateLimit = val
			return nil
		},
		onReset: func(po *TableStorageParamObserver, evalCtx *tree.EvalContext, key string) error {
			__antithesis_instrumentation__.Notify(552082)
			if po.tableDesc.RowLevelTTL != nil {
				__antithesis_instrumentation__.Notify(552084)
				po.tableDesc.RowLevelTTL.DeleteRateLimit = 0
			} else {
				__antithesis_instrumentation__.Notify(552085)
			}
			__antithesis_instrumentation__.Notify(552083)
			return nil
		},
	},
	`ttl_label_metrics`: {
		onSet: func(ctx context.Context, po *TableStorageParamObserver, semaCtx *tree.SemaContext, evalCtx *tree.EvalContext, key string, datum tree.Datum) error {
			__antithesis_instrumentation__.Notify(552086)
			if po.tableDesc.RowLevelTTL == nil {
				__antithesis_instrumentation__.Notify(552089)
				po.tableDesc.RowLevelTTL = &catpb.RowLevelTTL{}
			} else {
				__antithesis_instrumentation__.Notify(552090)
			}
			__antithesis_instrumentation__.Notify(552087)
			val, err := boolFromDatum(evalCtx, key, datum)
			if err != nil {
				__antithesis_instrumentation__.Notify(552091)
				return err
			} else {
				__antithesis_instrumentation__.Notify(552092)
			}
			__antithesis_instrumentation__.Notify(552088)
			po.tableDesc.RowLevelTTL.LabelMetrics = val
			return nil
		},
		onReset: func(po *TableStorageParamObserver, evalCtx *tree.EvalContext, key string) error {
			__antithesis_instrumentation__.Notify(552093)
			po.tableDesc.RowLevelTTL.LabelMetrics = false
			return nil
		},
	},
	`ttl_job_cron`: {
		onSet: func(ctx context.Context, po *TableStorageParamObserver, semaCtx *tree.SemaContext, evalCtx *tree.EvalContext, key string, datum tree.Datum) error {
			__antithesis_instrumentation__.Notify(552094)
			if po.tableDesc.RowLevelTTL == nil {
				__antithesis_instrumentation__.Notify(552098)
				po.tableDesc.RowLevelTTL = &catpb.RowLevelTTL{}
			} else {
				__antithesis_instrumentation__.Notify(552099)
			}
			__antithesis_instrumentation__.Notify(552095)
			str, err := DatumAsString(evalCtx, key, datum)
			if err != nil {
				__antithesis_instrumentation__.Notify(552100)
				return err
			} else {
				__antithesis_instrumentation__.Notify(552101)
			}
			__antithesis_instrumentation__.Notify(552096)
			if err := tabledesc.ValidateTTLCronExpr(key, str); err != nil {
				__antithesis_instrumentation__.Notify(552102)
				return err
			} else {
				__antithesis_instrumentation__.Notify(552103)
			}
			__antithesis_instrumentation__.Notify(552097)
			po.tableDesc.RowLevelTTL.DeletionCron = str
			return nil
		},
		onReset: func(po *TableStorageParamObserver, evalCtx *tree.EvalContext, key string) error {
			__antithesis_instrumentation__.Notify(552104)
			if po.tableDesc.RowLevelTTL != nil {
				__antithesis_instrumentation__.Notify(552106)
				po.tableDesc.RowLevelTTL.DeletionCron = ""
			} else {
				__antithesis_instrumentation__.Notify(552107)
			}
			__antithesis_instrumentation__.Notify(552105)
			return nil
		},
	},
	`ttl_pause`: {
		onSet: func(ctx context.Context, po *TableStorageParamObserver, semaCtx *tree.SemaContext, evalCtx *tree.EvalContext, key string, datum tree.Datum) error {
			__antithesis_instrumentation__.Notify(552108)
			b, err := boolFromDatum(evalCtx, key, datum)
			if err != nil {
				__antithesis_instrumentation__.Notify(552111)
				return err
			} else {
				__antithesis_instrumentation__.Notify(552112)
			}
			__antithesis_instrumentation__.Notify(552109)
			if po.tableDesc.RowLevelTTL == nil {
				__antithesis_instrumentation__.Notify(552113)
				po.tableDesc.RowLevelTTL = &catpb.RowLevelTTL{}
			} else {
				__antithesis_instrumentation__.Notify(552114)
			}
			__antithesis_instrumentation__.Notify(552110)
			po.tableDesc.RowLevelTTL.Pause = b
			return nil
		},
		onReset: func(po *TableStorageParamObserver, evalCtx *tree.EvalContext, key string) error {
			__antithesis_instrumentation__.Notify(552115)
			po.tableDesc.RowLevelTTL.Pause = false
			return nil
		},
	},
	`ttl_row_stats_poll_interval`: {
		onSet: func(ctx context.Context, po *TableStorageParamObserver, semaCtx *tree.SemaContext, evalCtx *tree.EvalContext, key string, datum tree.Datum) error {
			__antithesis_instrumentation__.Notify(552116)
			d, err := DatumAsDuration(evalCtx, key, datum)
			if err != nil {
				__antithesis_instrumentation__.Notify(552120)
				return err
			} else {
				__antithesis_instrumentation__.Notify(552121)
			}
			__antithesis_instrumentation__.Notify(552117)
			if po.tableDesc.RowLevelTTL == nil {
				__antithesis_instrumentation__.Notify(552122)
				po.tableDesc.RowLevelTTL = &catpb.RowLevelTTL{}
			} else {
				__antithesis_instrumentation__.Notify(552123)
			}
			__antithesis_instrumentation__.Notify(552118)
			if err := tabledesc.ValidateTTLRowStatsPollInterval(key, d); err != nil {
				__antithesis_instrumentation__.Notify(552124)
				return err
			} else {
				__antithesis_instrumentation__.Notify(552125)
			}
			__antithesis_instrumentation__.Notify(552119)
			po.tableDesc.RowLevelTTL.RowStatsPollInterval = d
			return nil
		},
		onReset: func(po *TableStorageParamObserver, evalCtx *tree.EvalContext, key string) error {
			__antithesis_instrumentation__.Notify(552126)
			po.tableDesc.RowLevelTTL.RowStatsPollInterval = 0
			return nil
		},
	},
	`exclude_data_from_backup`: {
		onSet: func(ctx context.Context, po *TableStorageParamObserver, semaCtx *tree.SemaContext,
			evalCtx *tree.EvalContext, key string, datum tree.Datum) error {
			__antithesis_instrumentation__.Notify(552127)
			if po.tableDesc.Temporary {
				__antithesis_instrumentation__.Notify(552132)
				return pgerror.Newf(pgcode.FeatureNotSupported,
					"cannot set data in a temporary table to be excluded from backup")
			} else {
				__antithesis_instrumentation__.Notify(552133)
			}
			__antithesis_instrumentation__.Notify(552128)

			if len(po.tableDesc.InboundFKs) != 0 {
				__antithesis_instrumentation__.Notify(552134)
				return errors.New("cannot set data in a table with inbound foreign key constraints to be excluded from backup")
			} else {
				__antithesis_instrumentation__.Notify(552135)
			}
			__antithesis_instrumentation__.Notify(552129)

			excludeDataFromBackup, err := boolFromDatum(evalCtx, key, datum)
			if err != nil {
				__antithesis_instrumentation__.Notify(552136)
				return err
			} else {
				__antithesis_instrumentation__.Notify(552137)
			}
			__antithesis_instrumentation__.Notify(552130)

			if po.tableDesc.ExcludeDataFromBackup == excludeDataFromBackup {
				__antithesis_instrumentation__.Notify(552138)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(552139)
			}
			__antithesis_instrumentation__.Notify(552131)
			po.tableDesc.ExcludeDataFromBackup = excludeDataFromBackup
			return nil
		},
		onReset: func(po *TableStorageParamObserver, evalCtx *tree.EvalContext, key string) error {
			__antithesis_instrumentation__.Notify(552140)
			po.tableDesc.ExcludeDataFromBackup = false
			return nil
		},
	},
}

func init() {
	for _, param := range []string{
		`toast_tuple_target`,
		`parallel_workers`,
		`toast.autovacuum_enabled`,
		`autovacuum_vacuum_threshold`,
		`toast.autovacuum_vacuum_threshold`,
		`autovacuum_vacuum_scale_factor`,
		`toast.autovacuum_vacuum_scale_factor`,
		`autovacuum_analyze_threshold`,
		`autovacuum_analyze_scale_factor`,
		`autovacuum_vacuum_cost_delay`,
		`toast.autovacuum_vacuum_cost_delay`,
		`autovacuum_vacuum_cost_limit`,
		`autovacuum_freeze_min_age`,
		`toast.autovacuum_freeze_min_age`,
		`autovacuum_freeze_max_age`,
		`toast.autovacuum_freeze_max_age`,
		`autovacuum_freeze_table_age`,
		`toast.autovacuum_freeze_table_age`,
		`autovacuum_multixact_freeze_min_age`,
		`toast.autovacuum_multixact_freeze_min_age`,
		`autovacuum_multixact_freeze_max_age`,
		`toast.autovacuum_multixact_freeze_max_age`,
		`autovacuum_multixact_freeze_table_age`,
		`toast.autovacuum_multixact_freeze_table_age`,
		`log_autovacuum_min_duration`,
		`toast.log_autovacuum_min_duration`,
		`user_catalog_table`,
	} {
		tableParams[param] = tableParam{
			onSet: func(ctx context.Context, po *TableStorageParamObserver, semaCtx *tree.SemaContext, evalCtx *tree.EvalContext, key string, datum tree.Datum) error {
				return unimplemented.NewWithIssuef(43299, "storage parameter %q", key)
			},
			onReset: func(po *TableStorageParamObserver, evalCtx *tree.EvalContext, key string) error {
				return nil
			},
		}
	}
}

func (po *TableStorageParamObserver) onSet(
	ctx context.Context,
	semaCtx *tree.SemaContext,
	evalCtx *tree.EvalContext,
	key string,
	datum tree.Datum,
) error {
	__antithesis_instrumentation__.Notify(552141)
	if strings.HasPrefix(key, "ttl_") && func() bool {
		__antithesis_instrumentation__.Notify(552144)
		return len(po.tableDesc.AllMutations()) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(552145)
		return pgerror.Newf(
			pgcode.FeatureNotSupported,
			"cannot modify TTL settings while another schema change on the table is being processed",
		)
	} else {
		__antithesis_instrumentation__.Notify(552146)
	}
	__antithesis_instrumentation__.Notify(552142)
	if p, ok := tableParams[key]; ok {
		__antithesis_instrumentation__.Notify(552147)
		return p.onSet(ctx, po, semaCtx, evalCtx, key, datum)
	} else {
		__antithesis_instrumentation__.Notify(552148)
	}
	__antithesis_instrumentation__.Notify(552143)
	return pgerror.Newf(pgcode.InvalidParameterValue, "invalid storage parameter %q", key)
}

func (po *TableStorageParamObserver) onReset(evalCtx *tree.EvalContext, key string) error {
	__antithesis_instrumentation__.Notify(552149)
	if strings.HasPrefix(key, "ttl_") && func() bool {
		__antithesis_instrumentation__.Notify(552152)
		return len(po.tableDesc.AllMutations()) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(552153)
		return pgerror.Newf(
			pgcode.FeatureNotSupported,
			"cannot modify TTL settings while another schema change on the table is being processed",
		)
	} else {
		__antithesis_instrumentation__.Notify(552154)
	}
	__antithesis_instrumentation__.Notify(552150)
	if p, ok := tableParams[key]; ok {
		__antithesis_instrumentation__.Notify(552155)
		return p.onReset(po, evalCtx, key)
	} else {
		__antithesis_instrumentation__.Notify(552156)
	}
	__antithesis_instrumentation__.Notify(552151)
	return pgerror.Newf(pgcode.InvalidParameterValue, "invalid storage parameter %q", key)
}

func setFillFactorStorageParam(evalCtx *tree.EvalContext, key string, datum tree.Datum) error {
	__antithesis_instrumentation__.Notify(552157)
	val, err := DatumAsFloat(evalCtx, key, datum)
	if err != nil {
		__antithesis_instrumentation__.Notify(552161)
		return err
	} else {
		__antithesis_instrumentation__.Notify(552162)
	}
	__antithesis_instrumentation__.Notify(552158)
	if val < 0 || func() bool {
		__antithesis_instrumentation__.Notify(552163)
		return val > 100 == true
	}() == true {
		__antithesis_instrumentation__.Notify(552164)
		return pgerror.Newf(pgcode.InvalidParameterValue, "%q must be between 0 and 100", key)
	} else {
		__antithesis_instrumentation__.Notify(552165)
	}
	__antithesis_instrumentation__.Notify(552159)
	if evalCtx != nil {
		__antithesis_instrumentation__.Notify(552166)
		evalCtx.ClientNoticeSender.BufferClientNotice(
			evalCtx.Context,
			pgnotice.Newf("storage parameter %q is ignored", key),
		)
	} else {
		__antithesis_instrumentation__.Notify(552167)
	}
	__antithesis_instrumentation__.Notify(552160)
	return nil
}

type IndexStorageParamObserver struct {
	IndexDesc *descpb.IndexDescriptor
}

var _ StorageParamObserver = (*IndexStorageParamObserver)(nil)

func getS2ConfigFromIndex(indexDesc *descpb.IndexDescriptor) *geoindex.S2Config {
	__antithesis_instrumentation__.Notify(552168)
	var s2Config *geoindex.S2Config
	if indexDesc.GeoConfig.S2Geometry != nil {
		__antithesis_instrumentation__.Notify(552171)
		s2Config = indexDesc.GeoConfig.S2Geometry.S2Config
	} else {
		__antithesis_instrumentation__.Notify(552172)
	}
	__antithesis_instrumentation__.Notify(552169)
	if indexDesc.GeoConfig.S2Geography != nil {
		__antithesis_instrumentation__.Notify(552173)
		s2Config = indexDesc.GeoConfig.S2Geography.S2Config
	} else {
		__antithesis_instrumentation__.Notify(552174)
	}
	__antithesis_instrumentation__.Notify(552170)
	return s2Config
}

func (po *IndexStorageParamObserver) applyS2ConfigSetting(
	evalCtx *tree.EvalContext, key string, expr tree.Datum, min int64, max int64,
) error {
	__antithesis_instrumentation__.Notify(552175)
	s2Config := getS2ConfigFromIndex(po.IndexDesc)
	if s2Config == nil {
		__antithesis_instrumentation__.Notify(552180)
		return pgerror.Newf(
			pgcode.InvalidParameterValue,
			"index setting %q can only be set on GEOMETRY or GEOGRAPHY spatial indexes",
			key,
		)
	} else {
		__antithesis_instrumentation__.Notify(552181)
	}
	__antithesis_instrumentation__.Notify(552176)

	val, err := DatumAsInt(evalCtx, key, expr)
	if err != nil {
		__antithesis_instrumentation__.Notify(552182)
		return errors.Wrapf(err, "error decoding %q", key)
	} else {
		__antithesis_instrumentation__.Notify(552183)
	}
	__antithesis_instrumentation__.Notify(552177)
	if val < min || func() bool {
		__antithesis_instrumentation__.Notify(552184)
		return val > max == true
	}() == true {
		__antithesis_instrumentation__.Notify(552185)
		return pgerror.Newf(
			pgcode.InvalidParameterValue,
			"%q value must be between %d and %d inclusive",
			key,
			min,
			max,
		)
	} else {
		__antithesis_instrumentation__.Notify(552186)
	}
	__antithesis_instrumentation__.Notify(552178)
	switch key {
	case `s2_max_level`:
		__antithesis_instrumentation__.Notify(552187)
		s2Config.MaxLevel = int32(val)
	case `s2_level_mod`:
		__antithesis_instrumentation__.Notify(552188)
		s2Config.LevelMod = int32(val)
	case `s2_max_cells`:
		__antithesis_instrumentation__.Notify(552189)
		s2Config.MaxCells = int32(val)
	default:
		__antithesis_instrumentation__.Notify(552190)
	}
	__antithesis_instrumentation__.Notify(552179)

	return nil
}

func (po *IndexStorageParamObserver) applyGeometryIndexSetting(
	evalCtx *tree.EvalContext, key string, expr tree.Datum,
) error {
	__antithesis_instrumentation__.Notify(552191)
	if po.IndexDesc.GeoConfig.S2Geometry == nil {
		__antithesis_instrumentation__.Notify(552195)
		return pgerror.Newf(pgcode.InvalidParameterValue, "%q can only be applied to GEOMETRY spatial indexes", key)
	} else {
		__antithesis_instrumentation__.Notify(552196)
	}
	__antithesis_instrumentation__.Notify(552192)
	val, err := DatumAsFloat(evalCtx, key, expr)
	if err != nil {
		__antithesis_instrumentation__.Notify(552197)
		return errors.Wrapf(err, "error decoding %q", key)
	} else {
		__antithesis_instrumentation__.Notify(552198)
	}
	__antithesis_instrumentation__.Notify(552193)
	switch key {
	case `geometry_min_x`:
		__antithesis_instrumentation__.Notify(552199)
		po.IndexDesc.GeoConfig.S2Geometry.MinX = val
	case `geometry_max_x`:
		__antithesis_instrumentation__.Notify(552200)
		po.IndexDesc.GeoConfig.S2Geometry.MaxX = val
	case `geometry_min_y`:
		__antithesis_instrumentation__.Notify(552201)
		po.IndexDesc.GeoConfig.S2Geometry.MinY = val
	case `geometry_max_y`:
		__antithesis_instrumentation__.Notify(552202)
		po.IndexDesc.GeoConfig.S2Geometry.MaxY = val
	default:
		__antithesis_instrumentation__.Notify(552203)
		return pgerror.Newf(pgcode.InvalidParameterValue, "unknown key: %q", key)
	}
	__antithesis_instrumentation__.Notify(552194)
	return nil
}

func (po *IndexStorageParamObserver) onSet(
	ctx context.Context,
	semaCtx *tree.SemaContext,
	evalCtx *tree.EvalContext,
	key string,
	expr tree.Datum,
) error {
	__antithesis_instrumentation__.Notify(552204)
	switch key {
	case `fillfactor`:
		__antithesis_instrumentation__.Notify(552206)
		return setFillFactorStorageParam(evalCtx, key, expr)
	case `s2_max_level`:
		__antithesis_instrumentation__.Notify(552207)
		return po.applyS2ConfigSetting(evalCtx, key, expr, 0, 30)
	case `s2_level_mod`:
		__antithesis_instrumentation__.Notify(552208)
		return po.applyS2ConfigSetting(evalCtx, key, expr, 1, 3)
	case `s2_max_cells`:
		__antithesis_instrumentation__.Notify(552209)
		return po.applyS2ConfigSetting(evalCtx, key, expr, 1, 32)
	case `geometry_min_x`, `geometry_max_x`, `geometry_min_y`, `geometry_max_y`:
		__antithesis_instrumentation__.Notify(552210)
		return po.applyGeometryIndexSetting(evalCtx, key, expr)

	case `bucket_count`:
		__antithesis_instrumentation__.Notify(552211)
		return nil
	case `vacuum_cleanup_index_scale_factor`,
		`buffering`,
		`fastupdate`,
		`gin_pending_list_limit`,
		`pages_per_range`,
		`autosummarize`:
		__antithesis_instrumentation__.Notify(552212)
		return unimplemented.NewWithIssuef(43299, "storage parameter %q", key)
	default:
		__antithesis_instrumentation__.Notify(552213)
	}
	__antithesis_instrumentation__.Notify(552205)
	return pgerror.Newf(pgcode.InvalidParameterValue, "invalid storage parameter %q", key)
}

func (po *IndexStorageParamObserver) onReset(evalCtx *tree.EvalContext, key string) error {
	__antithesis_instrumentation__.Notify(552214)
	return errors.AssertionFailedf("non-implemented codepath")
}

func (po *IndexStorageParamObserver) runPostChecks() error {
	__antithesis_instrumentation__.Notify(552215)
	s2Config := getS2ConfigFromIndex(po.IndexDesc)
	if s2Config != nil {
		__antithesis_instrumentation__.Notify(552218)
		if (s2Config.MaxLevel)%s2Config.LevelMod != 0 {
			__antithesis_instrumentation__.Notify(552219)
			return pgerror.Newf(
				pgcode.InvalidParameterValue,
				"s2_max_level (%d) must be divisible by s2_level_mod (%d)",
				s2Config.MaxLevel,
				s2Config.LevelMod,
			)
		} else {
			__antithesis_instrumentation__.Notify(552220)
		}
	} else {
		__antithesis_instrumentation__.Notify(552221)
	}
	__antithesis_instrumentation__.Notify(552216)

	if cfg := po.IndexDesc.GeoConfig.S2Geometry; cfg != nil {
		__antithesis_instrumentation__.Notify(552222)
		if cfg.MaxX <= cfg.MinX {
			__antithesis_instrumentation__.Notify(552224)
			return pgerror.Newf(
				pgcode.InvalidParameterValue,
				"geometry_max_x (%f) must be greater than geometry_min_x (%f)",
				cfg.MaxX,
				cfg.MinX,
			)
		} else {
			__antithesis_instrumentation__.Notify(552225)
		}
		__antithesis_instrumentation__.Notify(552223)
		if cfg.MaxY <= cfg.MinY {
			__antithesis_instrumentation__.Notify(552226)
			return pgerror.Newf(
				pgcode.InvalidParameterValue,
				"geometry_max_y (%f) must be greater than geometry_min_y (%f)",
				cfg.MaxY,
				cfg.MinY,
			)
		} else {
			__antithesis_instrumentation__.Notify(552227)
		}
	} else {
		__antithesis_instrumentation__.Notify(552228)
	}
	__antithesis_instrumentation__.Notify(552217)
	return nil
}
