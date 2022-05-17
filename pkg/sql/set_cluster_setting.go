package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/colexec"
	"github.com/cockroachdb/cockroach/pkg/sql/paramparse"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/roleoption"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/sql/sessioninit"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/stats"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/hintdetail"
	"github.com/cockroachdb/redact"
)

type setClusterSettingNode struct {
	name    string
	st      *cluster.Settings
	setting settings.NonMaskedSetting

	value tree.TypedExpr
}

func checkPrivilegesForSetting(ctx context.Context, p *planner, name string, action string) error {
	__antithesis_instrumentation__.Notify(621711)
	if settings.AdminOnly(name) {
		__antithesis_instrumentation__.Notify(621717)
		return p.RequireAdminRole(ctx, fmt.Sprintf("%s cluster setting '%s'", action, name))
	} else {
		__antithesis_instrumentation__.Notify(621718)
	}
	__antithesis_instrumentation__.Notify(621712)
	hasModify, err := p.HasRoleOption(ctx, roleoption.MODIFYCLUSTERSETTING)
	if err != nil {
		__antithesis_instrumentation__.Notify(621719)
		return err
	} else {
		__antithesis_instrumentation__.Notify(621720)
	}
	__antithesis_instrumentation__.Notify(621713)
	if action == "set" && func() bool {
		__antithesis_instrumentation__.Notify(621721)
		return !hasModify == true
	}() == true {
		__antithesis_instrumentation__.Notify(621722)
		return pgerror.Newf(pgcode.InsufficientPrivilege,
			"only users with the %s privilege are allowed to %s cluster setting '%s'",
			roleoption.MODIFYCLUSTERSETTING, action, name)
	} else {
		__antithesis_instrumentation__.Notify(621723)
	}
	__antithesis_instrumentation__.Notify(621714)
	hasView, err := p.HasRoleOption(ctx, roleoption.VIEWCLUSTERSETTING)
	if err != nil {
		__antithesis_instrumentation__.Notify(621724)
		return err
	} else {
		__antithesis_instrumentation__.Notify(621725)
	}
	__antithesis_instrumentation__.Notify(621715)

	if action == "show" && func() bool {
		__antithesis_instrumentation__.Notify(621726)
		return !(hasModify || func() bool {
			__antithesis_instrumentation__.Notify(621727)
			return hasView == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(621728)
		return pgerror.Newf(pgcode.InsufficientPrivilege,
			"only users with either %s or %s privileges are allowed to %s cluster setting '%s'",
			roleoption.MODIFYCLUSTERSETTING, roleoption.VIEWCLUSTERSETTING, action, name)
	} else {
		__antithesis_instrumentation__.Notify(621729)
	}
	__antithesis_instrumentation__.Notify(621716)
	return nil
}

func (p *planner) SetClusterSetting(
	ctx context.Context, n *tree.SetClusterSetting,
) (planNode, error) {
	__antithesis_instrumentation__.Notify(621730)
	name := strings.ToLower(n.Name)
	st := p.EvalContext().Settings
	v, ok := settings.Lookup(name, settings.LookupForLocalAccess, p.ExecCfg().Codec.ForSystemTenant())
	if !ok {
		__antithesis_instrumentation__.Notify(621737)
		return nil, errors.Errorf("unknown cluster setting '%s'", name)
	} else {
		__antithesis_instrumentation__.Notify(621738)
	}
	__antithesis_instrumentation__.Notify(621731)

	if err := checkPrivilegesForSetting(ctx, p, name, "set"); err != nil {
		__antithesis_instrumentation__.Notify(621739)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(621740)
	}
	__antithesis_instrumentation__.Notify(621732)

	setting, ok := v.(settings.NonMaskedSetting)
	if !ok {
		__antithesis_instrumentation__.Notify(621741)
		return nil, errors.AssertionFailedf("expected writable setting, got %T", v)
	} else {
		__antithesis_instrumentation__.Notify(621742)
	}
	__antithesis_instrumentation__.Notify(621733)

	if !p.execCfg.Codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(621743)
		switch setting.Class() {
		case settings.SystemOnly:
			__antithesis_instrumentation__.Notify(621744)

			return nil, errors.AssertionFailedf("looked up system-only setting")
		case settings.TenantReadOnly:
			__antithesis_instrumentation__.Notify(621745)
			return nil, pgerror.Newf(pgcode.InsufficientPrivilege, "setting %s is only settable by the operator", name)
		default:
			__antithesis_instrumentation__.Notify(621746)
		}
	} else {
		__antithesis_instrumentation__.Notify(621747)
	}
	__antithesis_instrumentation__.Notify(621734)

	if st.OverridesInformer != nil && func() bool {
		__antithesis_instrumentation__.Notify(621748)
		return st.OverridesInformer.IsOverridden(name) == true
	}() == true {
		__antithesis_instrumentation__.Notify(621749)
		return nil, errors.Errorf("cluster setting '%s' is currently overridden by the operator", name)
	} else {
		__antithesis_instrumentation__.Notify(621750)
	}
	__antithesis_instrumentation__.Notify(621735)

	value, err := p.getAndValidateTypedClusterSetting(ctx, name, n.Value, setting)
	if err != nil {
		__antithesis_instrumentation__.Notify(621751)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(621752)
	}
	__antithesis_instrumentation__.Notify(621736)

	csNode := setClusterSettingNode{
		name: name, st: st, setting: setting, value: value,
	}
	return &csNode, nil
}

func (p *planner) getAndValidateTypedClusterSetting(
	ctx context.Context, name string, expr tree.Expr, setting settings.NonMaskedSetting,
) (tree.TypedExpr, error) {
	__antithesis_instrumentation__.Notify(621753)
	var value tree.TypedExpr
	if expr != nil {
		__antithesis_instrumentation__.Notify(621755)

		if _, ok := expr.(tree.DefaultVal); !ok {
			__antithesis_instrumentation__.Notify(621756)
			expr = paramparse.UnresolvedNameToStrVal(expr)

			var requiredType *types.T
			var dummyHelper tree.IndexedVarHelper

			switch setting.(type) {
			case *settings.StringSetting, *settings.VersionSetting, *settings.ByteSizeSetting:
				__antithesis_instrumentation__.Notify(621759)
				requiredType = types.String
			case *settings.BoolSetting:
				__antithesis_instrumentation__.Notify(621760)
				requiredType = types.Bool
			case *settings.IntSetting:
				__antithesis_instrumentation__.Notify(621761)
				requiredType = types.Int
			case *settings.FloatSetting:
				__antithesis_instrumentation__.Notify(621762)
				requiredType = types.Float
			case *settings.EnumSetting:
				__antithesis_instrumentation__.Notify(621763)
				requiredType = types.Any
			case *settings.DurationSetting:
				__antithesis_instrumentation__.Notify(621764)
				requiredType = types.Interval
			case *settings.DurationSettingWithExplicitUnit:
				__antithesis_instrumentation__.Notify(621765)
				requiredType = types.Interval

				_, err := p.analyzeExpr(
					ctx, expr, nil, dummyHelper, types.Float, false, "SET CLUSTER SETTING "+name,
				)

				if err == nil {
					__antithesis_instrumentation__.Notify(621767)
					_, hint := setting.ErrorHint()
					return nil, hintdetail.WithHint(errors.New("invalid cluster setting argument type"), hint)
				} else {
					__antithesis_instrumentation__.Notify(621768)
				}
			default:
				__antithesis_instrumentation__.Notify(621766)
				return nil, errors.Errorf("unsupported setting type %T", setting)
			}
			__antithesis_instrumentation__.Notify(621757)

			typed, err := p.analyzeExpr(
				ctx, expr, nil, dummyHelper, requiredType, true, "SET CLUSTER SETTING "+name)
			if err != nil {
				__antithesis_instrumentation__.Notify(621769)
				hasHint, hint := setting.ErrorHint()
				if hasHint {
					__antithesis_instrumentation__.Notify(621771)
					return nil, hintdetail.WithHint(err, hint)
				} else {
					__antithesis_instrumentation__.Notify(621772)
				}
				__antithesis_instrumentation__.Notify(621770)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(621773)
			}
			__antithesis_instrumentation__.Notify(621758)
			value = typed
		} else {
			__antithesis_instrumentation__.Notify(621774)
			if _, isVersionSetting := setting.(*settings.VersionSetting); isVersionSetting {
				__antithesis_instrumentation__.Notify(621775)
				return nil, errors.New("cannot RESET cluster version setting")
			} else {
				__antithesis_instrumentation__.Notify(621776)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(621777)
	}
	__antithesis_instrumentation__.Notify(621754)
	return value, nil
}

func (n *setClusterSettingNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(621778)
	if !params.extendedEvalCtx.TxnIsSingleStmt {
		__antithesis_instrumentation__.Notify(621783)
		return errors.Errorf("SET CLUSTER SETTING cannot be used inside a multi-statement transaction")
	} else {
		__antithesis_instrumentation__.Notify(621784)
	}
	__antithesis_instrumentation__.Notify(621779)

	expectedEncodedValue, err := writeSettingInternal(
		params.ctx,
		params.extendedEvalCtx.ExecCfg,
		n.setting, n.name,
		params.p.User(),
		n.st,
		n.value,
		params.p.EvalContext(),
		params.extendedEvalCtx.Codec.ForSystemTenant(),
		params.p.logEvent,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(621785)
		return err
	} else {
		__antithesis_instrumentation__.Notify(621786)
	}
	__antithesis_instrumentation__.Notify(621780)

	if n.name == sessioninit.CacheEnabledSettingName {
		__antithesis_instrumentation__.Notify(621787)
		if expectedEncodedValue == "false" {
			__antithesis_instrumentation__.Notify(621788)

			if err := params.p.bumpUsersTableVersion(params.ctx); err != nil {
				__antithesis_instrumentation__.Notify(621790)
				return err
			} else {
				__antithesis_instrumentation__.Notify(621791)
			}
			__antithesis_instrumentation__.Notify(621789)
			if err := params.p.bumpRoleOptionsTableVersion(params.ctx); err != nil {
				__antithesis_instrumentation__.Notify(621792)
				return err
			} else {
				__antithesis_instrumentation__.Notify(621793)
			}
		} else {
			__antithesis_instrumentation__.Notify(621794)
		}
	} else {
		__antithesis_instrumentation__.Notify(621795)
	}
	__antithesis_instrumentation__.Notify(621781)

	switch n.name {
	case stats.AutoStatsClusterSettingName:
		__antithesis_instrumentation__.Notify(621796)
		switch expectedEncodedValue {
		case "true":
			__antithesis_instrumentation__.Notify(621806)
			telemetry.Inc(sqltelemetry.TurnAutoStatsOnUseCounter)
		case "false":
			__antithesis_instrumentation__.Notify(621807)
			telemetry.Inc(sqltelemetry.TurnAutoStatsOffUseCounter)
		default:
			__antithesis_instrumentation__.Notify(621808)
		}
	case ConnAuditingClusterSettingName:
		__antithesis_instrumentation__.Notify(621797)
		switch expectedEncodedValue {
		case "true":
			__antithesis_instrumentation__.Notify(621809)
			telemetry.Inc(sqltelemetry.TurnConnAuditingOnUseCounter)
		case "false":
			__antithesis_instrumentation__.Notify(621810)
			telemetry.Inc(sqltelemetry.TurnConnAuditingOffUseCounter)
		default:
			__antithesis_instrumentation__.Notify(621811)
		}
	case AuthAuditingClusterSettingName:
		__antithesis_instrumentation__.Notify(621798)
		switch expectedEncodedValue {
		case "true":
			__antithesis_instrumentation__.Notify(621812)
			telemetry.Inc(sqltelemetry.TurnAuthAuditingOnUseCounter)
		case "false":
			__antithesis_instrumentation__.Notify(621813)
			telemetry.Inc(sqltelemetry.TurnAuthAuditingOffUseCounter)
		default:
			__antithesis_instrumentation__.Notify(621814)
		}
	case ReorderJoinsLimitClusterSettingName:
		__antithesis_instrumentation__.Notify(621799)
		val, err := strconv.ParseInt(expectedEncodedValue, 10, 64)
		if err != nil {
			__antithesis_instrumentation__.Notify(621815)
			break
		} else {
			__antithesis_instrumentation__.Notify(621816)
		}
		__antithesis_instrumentation__.Notify(621800)
		sqltelemetry.ReportJoinReorderLimit(int(val))
	case VectorizeClusterSettingName:
		__antithesis_instrumentation__.Notify(621801)
		val, err := strconv.Atoi(expectedEncodedValue)
		if err != nil {
			__antithesis_instrumentation__.Notify(621817)
			break
		} else {
			__antithesis_instrumentation__.Notify(621818)
		}
		__antithesis_instrumentation__.Notify(621802)
		validatedExecMode, isValid := sessiondatapb.VectorizeExecModeFromString(sessiondatapb.VectorizeExecMode(val).String())
		if !isValid {
			__antithesis_instrumentation__.Notify(621819)
			break
		} else {
			__antithesis_instrumentation__.Notify(621820)
		}
		__antithesis_instrumentation__.Notify(621803)
		telemetry.Inc(sqltelemetry.VecModeCounter(validatedExecMode.String()))
	case colexec.HashAggregationDiskSpillingEnabledSettingName:
		__antithesis_instrumentation__.Notify(621804)
		if expectedEncodedValue == "false" {
			__antithesis_instrumentation__.Notify(621821)
			telemetry.Inc(sqltelemetry.HashAggregationDiskSpillingDisabled)
		} else {
			__antithesis_instrumentation__.Notify(621822)
		}
	default:
		__antithesis_instrumentation__.Notify(621805)
	}
	__antithesis_instrumentation__.Notify(621782)

	return waitForSettingUpdate(params.ctx, params.extendedEvalCtx.ExecCfg,
		n.setting, n.value == nil, n.name, expectedEncodedValue)
}

func writeSettingInternal(
	ctx context.Context,
	execCfg *ExecutorConfig,
	setting settings.NonMaskedSetting,
	name string,
	user security.SQLUsername,
	st *cluster.Settings,
	value tree.TypedExpr,
	evalCtx *tree.EvalContext,
	forSystemTenant bool,
	logFn func(context.Context, descpb.ID, eventpb.EventPayload) error,
) (expectedEncodedValue string, err error) {
	__antithesis_instrumentation__.Notify(621823)
	err = execCfg.DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(621825)
		var reportedValue string
		if value == nil {
			__antithesis_instrumentation__.Notify(621827)

			var err error
			reportedValue, expectedEncodedValue, err = writeDefaultSettingValue(ctx, execCfg, setting, name, txn)
			if err != nil {
				__antithesis_instrumentation__.Notify(621828)
				return err
			} else {
				__antithesis_instrumentation__.Notify(621829)
			}
		} else {
			__antithesis_instrumentation__.Notify(621830)

			value, err := value.Eval(evalCtx)
			if err != nil {
				__antithesis_instrumentation__.Notify(621832)
				return err
			} else {
				__antithesis_instrumentation__.Notify(621833)
			}
			__antithesis_instrumentation__.Notify(621831)
			reportedValue, expectedEncodedValue, err = writeNonDefaultSettingValue(
				ctx, execCfg, setting, name, txn,
				user, st, value, forSystemTenant,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(621834)
				return err
			} else {
				__antithesis_instrumentation__.Notify(621835)
			}
		}
		__antithesis_instrumentation__.Notify(621826)

		return logFn(ctx,
			0,
			&eventpb.SetClusterSetting{
				SettingName: name,
				Value:       reportedValue,
			})
	})
	__antithesis_instrumentation__.Notify(621824)
	return expectedEncodedValue, err
}

func writeDefaultSettingValue(
	ctx context.Context,
	execCfg *ExecutorConfig,
	setting settings.NonMaskedSetting,
	name string,
	txn *kv.Txn,
) (reportedValue string, expectedEncodedValue string, err error) {
	__antithesis_instrumentation__.Notify(621836)
	reportedValue = "DEFAULT"
	expectedEncodedValue = setting.EncodedDefault()
	_, err = execCfg.InternalExecutor.ExecEx(
		ctx, "reset-setting", txn,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		"DELETE FROM system.settings WHERE name = $1", name,
	)
	return reportedValue, expectedEncodedValue, err
}

func writeNonDefaultSettingValue(
	ctx context.Context,
	execCfg *ExecutorConfig,
	setting settings.NonMaskedSetting,
	name string,
	txn *kv.Txn,
	user security.SQLUsername,
	st *cluster.Settings,
	value tree.Datum,
	forSystemTenant bool,
) (reportedValue string, expectedEncodedValue string, err error) {
	__antithesis_instrumentation__.Notify(621837)

	reportedValue = tree.AsStringWithFlags(value, tree.FmtBareStrings)

	encoded, err := toSettingString(ctx, st, name, setting, value)
	expectedEncodedValue = encoded
	if err != nil {
		__antithesis_instrumentation__.Notify(621841)
		return reportedValue, expectedEncodedValue, err
	} else {
		__antithesis_instrumentation__.Notify(621842)
	}
	__antithesis_instrumentation__.Notify(621838)

	verSetting, isSetVersion := setting.(*settings.VersionSetting)
	if isSetVersion {
		__antithesis_instrumentation__.Notify(621843)
		if err := setVersionSetting(
			ctx, execCfg, verSetting, name, txn, user, st, value, encoded, forSystemTenant); err != nil {
			__antithesis_instrumentation__.Notify(621844)
			return reportedValue, expectedEncodedValue, err
		} else {
			__antithesis_instrumentation__.Notify(621845)
		}
	} else {
		__antithesis_instrumentation__.Notify(621846)

		if _, err = execCfg.InternalExecutor.ExecEx(
			ctx, "update-setting", txn,
			sessiondata.InternalExecutorOverride{User: security.RootUserName()},
			`UPSERT INTO system.settings (name, value, "lastUpdated", "valueType") VALUES ($1, $2, now(), $3)`,
			name, encoded, setting.Typ(),
		); err != nil {
			__antithesis_instrumentation__.Notify(621847)
			return reportedValue, expectedEncodedValue, err
		} else {
			__antithesis_instrumentation__.Notify(621848)
		}
	}
	__antithesis_instrumentation__.Notify(621839)

	if knobs := execCfg.TenantTestingKnobs; knobs != nil && func() bool {
		__antithesis_instrumentation__.Notify(621849)
		return knobs.ClusterSettingsUpdater != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(621850)
		encVal := settings.EncodedValue{
			Value: encoded,
			Type:  setting.Typ(),
		}
		if err := execCfg.TenantTestingKnobs.ClusterSettingsUpdater.Set(ctx, name, encVal); err != nil {
			__antithesis_instrumentation__.Notify(621851)
			return reportedValue, expectedEncodedValue, err
		} else {
			__antithesis_instrumentation__.Notify(621852)
		}
	} else {
		__antithesis_instrumentation__.Notify(621853)
	}
	__antithesis_instrumentation__.Notify(621840)
	return reportedValue, expectedEncodedValue, nil
}

func setVersionSetting(
	ctx context.Context,
	execCfg *ExecutorConfig,
	setting *settings.VersionSetting,
	name string,
	txn *kv.Txn,
	user security.SQLUsername,
	st *cluster.Settings,
	value tree.Datum,
	encoded string,
	forSystemTenant bool,
) error {
	__antithesis_instrumentation__.Notify(621854)

	datums, err := execCfg.InternalExecutor.QueryRowEx(
		ctx, "retrieve-prev-setting", txn,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		"SELECT value FROM system.settings WHERE name = $1", name,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(621860)
		return err
	} else {
		__antithesis_instrumentation__.Notify(621861)
	}
	__antithesis_instrumentation__.Notify(621855)
	var prev tree.Datum
	if len(datums) == 0 {
		__antithesis_instrumentation__.Notify(621862)

		if forSystemTenant {
			__antithesis_instrumentation__.Notify(621865)
			return errors.New("no persisted cluster version found, please retry later")
		} else {
			__antithesis_instrumentation__.Notify(621866)
		}
		__antithesis_instrumentation__.Notify(621863)

		tenantDefaultVersion := clusterversion.ClusterVersion{
			Version: roachpb.Version{Major: 20, Minor: 2},
		}

		prevEncoded, err := protoutil.Marshal(&tenantDefaultVersion)
		if err != nil {
			__antithesis_instrumentation__.Notify(621867)
			return errors.WithAssertionFailure(err)
		} else {
			__antithesis_instrumentation__.Notify(621868)
		}
		__antithesis_instrumentation__.Notify(621864)
		prev = tree.NewDString(string(prevEncoded))
	} else {
		__antithesis_instrumentation__.Notify(621869)
		prev = datums[0]
	}
	__antithesis_instrumentation__.Notify(621856)

	dStr, ok := prev.(*tree.DString)
	if !ok {
		__antithesis_instrumentation__.Notify(621870)
		return errors.Errorf("the existing value is not a string, got %T", prev)
	} else {
		__antithesis_instrumentation__.Notify(621871)
	}
	__antithesis_instrumentation__.Notify(621857)
	if err := setting.Validate(ctx, &st.SV, []byte(string(*dStr)), []byte(encoded)); err != nil {
		__antithesis_instrumentation__.Notify(621872)
		return err
	} else {
		__antithesis_instrumentation__.Notify(621873)
	}
	__antithesis_instrumentation__.Notify(621858)

	updateVersionSystemSetting := func(ctx context.Context, version clusterversion.ClusterVersion) error {
		__antithesis_instrumentation__.Notify(621874)
		rawValue, err := protoutil.Marshal(&version)
		if err != nil {
			__antithesis_instrumentation__.Notify(621876)
			return err
		} else {
			__antithesis_instrumentation__.Notify(621877)
		}
		__antithesis_instrumentation__.Notify(621875)
		return execCfg.DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(621878)

			datums, err := execCfg.InternalExecutor.QueryRowEx(
				ctx, "retrieve-prev-setting", txn,
				sessiondata.InternalExecutorOverride{User: security.RootUserName()},
				"SELECT value FROM system.settings WHERE name = $1", name,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(621881)
				return err
			} else {
				__antithesis_instrumentation__.Notify(621882)
			}
			__antithesis_instrumentation__.Notify(621879)
			if len(datums) > 0 {
				__antithesis_instrumentation__.Notify(621883)
				dStr, ok := datums[0].(*tree.DString)
				if !ok {
					__antithesis_instrumentation__.Notify(621887)
					return errors.AssertionFailedf("existing version value is not a string, got %T", datums[0])
				} else {
					__antithesis_instrumentation__.Notify(621888)
				}
				__antithesis_instrumentation__.Notify(621884)
				oldRawValue := []byte(string(*dStr))
				if bytes.Equal(oldRawValue, rawValue) {
					__antithesis_instrumentation__.Notify(621889)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(621890)
				}
				__antithesis_instrumentation__.Notify(621885)
				var oldValue clusterversion.ClusterVersion

				if err := protoutil.Unmarshal(oldRawValue, &oldValue); err != nil {
					__antithesis_instrumentation__.Notify(621891)
					return errors.NewAssertionErrorWithWrappedErrf(err, "decoding previous version %s",
						redact.SafeString(base64.StdEncoding.EncodeToString(oldRawValue)))
				} else {
					__antithesis_instrumentation__.Notify(621892)
				}
				__antithesis_instrumentation__.Notify(621886)
				if !oldValue.Less(version.Version) {
					__antithesis_instrumentation__.Notify(621893)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(621894)
				}
			} else {
				__antithesis_instrumentation__.Notify(621895)
			}
			__antithesis_instrumentation__.Notify(621880)

			_, err = execCfg.InternalExecutor.ExecEx(
				ctx, "update-setting", txn,
				sessiondata.InternalExecutorOverride{User: security.RootUserName()},
				`UPSERT INTO system.settings (name, value, "lastUpdated", "valueType") VALUES ($1, $2, now(), $3)`,
				name, string(rawValue), setting.Typ(),
			)
			return err
		})
	}
	__antithesis_instrumentation__.Notify(621859)

	return runMigrationsAndUpgradeVersion(
		ctx, execCfg, user, prev, value, updateVersionSystemSetting,
	)
}

func waitForSettingUpdate(
	ctx context.Context,
	execCfg *ExecutorConfig,
	setting settings.NonMaskedSetting,
	reset bool,
	name string,
	expectedEncodedValue string,
) error {
	__antithesis_instrumentation__.Notify(621896)
	if _, ok := setting.(*settings.VersionSetting); ok && func() bool {
		__antithesis_instrumentation__.Notify(621900)
		return reset == true
	}() == true {
		__antithesis_instrumentation__.Notify(621901)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(621902)
	}
	__antithesis_instrumentation__.Notify(621897)
	errNotReady := errors.New("setting updated but timed out waiting to read new value")
	var observed string
	err := retry.ForDuration(10*time.Second, func() error {
		__antithesis_instrumentation__.Notify(621903)
		observed = setting.Encoded(&execCfg.Settings.SV)
		if observed != expectedEncodedValue {
			__antithesis_instrumentation__.Notify(621905)
			return errNotReady
		} else {
			__antithesis_instrumentation__.Notify(621906)
		}
		__antithesis_instrumentation__.Notify(621904)
		return nil
	})
	__antithesis_instrumentation__.Notify(621898)
	if err != nil {
		__antithesis_instrumentation__.Notify(621907)
		log.Warningf(
			ctx, "SET CLUSTER SETTING %q timed out waiting for value %q, observed %q",
			name, expectedEncodedValue, observed,
		)
	} else {
		__antithesis_instrumentation__.Notify(621908)
	}
	__antithesis_instrumentation__.Notify(621899)
	return err
}

func runMigrationsAndUpgradeVersion(
	ctx context.Context,
	execCfg *ExecutorConfig,
	user security.SQLUsername,
	prev tree.Datum,
	value tree.Datum,
	updateVersionSystemSetting UpdateVersionSystemSettingHook,
) error {
	__antithesis_instrumentation__.Notify(621909)
	var from, to clusterversion.ClusterVersion

	fromVersionVal := []byte(string(*prev.(*tree.DString)))
	if err := protoutil.Unmarshal(fromVersionVal, &from); err != nil {
		__antithesis_instrumentation__.Notify(621912)
		return err
	} else {
		__antithesis_instrumentation__.Notify(621913)
	}
	__antithesis_instrumentation__.Notify(621910)

	targetVersionStr := string(*value.(*tree.DString))
	to.Version = roachpb.MustParseVersion(targetVersionStr)

	if err := execCfg.VersionUpgradeHook(ctx, user, from, to, updateVersionSystemSetting); err != nil {
		__antithesis_instrumentation__.Notify(621914)
		return err
	} else {
		__antithesis_instrumentation__.Notify(621915)
	}
	__antithesis_instrumentation__.Notify(621911)
	return nil
}

func (n *setClusterSettingNode) Next(_ runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(621916)
	return false, nil
}
func (n *setClusterSettingNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(621917)
	return nil
}
func (n *setClusterSettingNode) Close(_ context.Context) {
	__antithesis_instrumentation__.Notify(621918)
}

func toSettingString(
	ctx context.Context, st *cluster.Settings, name string, s settings.Setting, d tree.Datum,
) (string, error) {
	__antithesis_instrumentation__.Notify(621919)
	switch setting := s.(type) {
	case *settings.StringSetting:
		__antithesis_instrumentation__.Notify(621920)
		if s, ok := d.(*tree.DString); ok {
			__antithesis_instrumentation__.Notify(621939)
			if err := setting.Validate(&st.SV, string(*s)); err != nil {
				__antithesis_instrumentation__.Notify(621941)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(621942)
			}
			__antithesis_instrumentation__.Notify(621940)
			return string(*s), nil
		} else {
			__antithesis_instrumentation__.Notify(621943)
		}
		__antithesis_instrumentation__.Notify(621921)
		return "", errors.Errorf("cannot use %s %T value for string setting", d.ResolvedType(), d)
	case *settings.VersionSetting:
		__antithesis_instrumentation__.Notify(621922)
		if s, ok := d.(*tree.DString); ok {
			__antithesis_instrumentation__.Notify(621944)
			newRawVal, err := clusterversion.EncodingFromVersionStr(string(*s))
			if err != nil {
				__antithesis_instrumentation__.Notify(621946)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(621947)
			}
			__antithesis_instrumentation__.Notify(621945)
			return string(newRawVal), nil
		} else {
			__antithesis_instrumentation__.Notify(621948)
		}
		__antithesis_instrumentation__.Notify(621923)
		return "", errors.Errorf("cannot use %s %T value for string setting", d.ResolvedType(), d)
	case *settings.BoolSetting:
		__antithesis_instrumentation__.Notify(621924)
		if b, ok := d.(*tree.DBool); ok {
			__antithesis_instrumentation__.Notify(621949)
			return settings.EncodeBool(bool(*b)), nil
		} else {
			__antithesis_instrumentation__.Notify(621950)
		}
		__antithesis_instrumentation__.Notify(621925)
		return "", errors.Errorf("cannot use %s %T value for bool setting", d.ResolvedType(), d)
	case *settings.IntSetting:
		__antithesis_instrumentation__.Notify(621926)
		if i, ok := d.(*tree.DInt); ok {
			__antithesis_instrumentation__.Notify(621951)
			if err := setting.Validate(int64(*i)); err != nil {
				__antithesis_instrumentation__.Notify(621953)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(621954)
			}
			__antithesis_instrumentation__.Notify(621952)
			return settings.EncodeInt(int64(*i)), nil
		} else {
			__antithesis_instrumentation__.Notify(621955)
		}
		__antithesis_instrumentation__.Notify(621927)
		return "", errors.Errorf("cannot use %s %T value for int setting", d.ResolvedType(), d)
	case *settings.FloatSetting:
		__antithesis_instrumentation__.Notify(621928)
		if f, ok := d.(*tree.DFloat); ok {
			__antithesis_instrumentation__.Notify(621956)
			if err := setting.Validate(float64(*f)); err != nil {
				__antithesis_instrumentation__.Notify(621958)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(621959)
			}
			__antithesis_instrumentation__.Notify(621957)
			return settings.EncodeFloat(float64(*f)), nil
		} else {
			__antithesis_instrumentation__.Notify(621960)
		}
		__antithesis_instrumentation__.Notify(621929)
		return "", errors.Errorf("cannot use %s %T value for float setting", d.ResolvedType(), d)
	case *settings.EnumSetting:
		__antithesis_instrumentation__.Notify(621930)
		if i, intOK := d.(*tree.DInt); intOK {
			__antithesis_instrumentation__.Notify(621961)
			v, ok := setting.ParseEnum(settings.EncodeInt(int64(*i)))
			if ok {
				__antithesis_instrumentation__.Notify(621963)
				return settings.EncodeInt(v), nil
			} else {
				__antithesis_instrumentation__.Notify(621964)
			}
			__antithesis_instrumentation__.Notify(621962)
			return "", errors.WithHintf(errors.Errorf("invalid integer value '%d' for enum setting", *i), setting.GetAvailableValuesAsHint())
		} else {
			__antithesis_instrumentation__.Notify(621965)
			if s, ok := d.(*tree.DString); ok {
				__antithesis_instrumentation__.Notify(621966)
				str := string(*s)
				v, ok := setting.ParseEnum(str)
				if ok {
					__antithesis_instrumentation__.Notify(621968)
					return settings.EncodeInt(v), nil
				} else {
					__antithesis_instrumentation__.Notify(621969)
				}
				__antithesis_instrumentation__.Notify(621967)
				return "", errors.WithHintf(errors.Errorf("invalid string value '%s' for enum setting", str), setting.GetAvailableValuesAsHint())
			} else {
				__antithesis_instrumentation__.Notify(621970)
			}
		}
		__antithesis_instrumentation__.Notify(621931)
		return "", errors.Errorf("cannot use %s %T value for enum setting, must be int or string", d.ResolvedType(), d)
	case *settings.ByteSizeSetting:
		__antithesis_instrumentation__.Notify(621932)
		if s, ok := d.(*tree.DString); ok {
			__antithesis_instrumentation__.Notify(621971)
			bytes, err := humanizeutil.ParseBytes(string(*s))
			if err != nil {
				__antithesis_instrumentation__.Notify(621974)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(621975)
			}
			__antithesis_instrumentation__.Notify(621972)
			if err := setting.Validate(bytes); err != nil {
				__antithesis_instrumentation__.Notify(621976)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(621977)
			}
			__antithesis_instrumentation__.Notify(621973)
			return settings.EncodeInt(bytes), nil
		} else {
			__antithesis_instrumentation__.Notify(621978)
		}
		__antithesis_instrumentation__.Notify(621933)
		return "", errors.Errorf("cannot use %s %T value for byte size setting", d.ResolvedType(), d)
	case *settings.DurationSetting:
		__antithesis_instrumentation__.Notify(621934)
		if f, ok := d.(*tree.DInterval); ok {
			__antithesis_instrumentation__.Notify(621979)
			if f.Duration.Months > 0 || func() bool {
				__antithesis_instrumentation__.Notify(621982)
				return f.Duration.Days > 0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(621983)
				return "", errors.Errorf("cannot use day or month specifiers: %s", d.String())
			} else {
				__antithesis_instrumentation__.Notify(621984)
			}
			__antithesis_instrumentation__.Notify(621980)
			d := time.Duration(f.Duration.Nanos()) * time.Nanosecond
			if err := setting.Validate(d); err != nil {
				__antithesis_instrumentation__.Notify(621985)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(621986)
			}
			__antithesis_instrumentation__.Notify(621981)
			return settings.EncodeDuration(d), nil
		} else {
			__antithesis_instrumentation__.Notify(621987)
		}
		__antithesis_instrumentation__.Notify(621935)
		return "", errors.Errorf("cannot use %s %T value for duration setting", d.ResolvedType(), d)
	case *settings.DurationSettingWithExplicitUnit:
		__antithesis_instrumentation__.Notify(621936)
		if f, ok := d.(*tree.DInterval); ok {
			__antithesis_instrumentation__.Notify(621988)
			if f.Duration.Months > 0 || func() bool {
				__antithesis_instrumentation__.Notify(621991)
				return f.Duration.Days > 0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(621992)
				return "", errors.Errorf("cannot use day or month specifiers: %s", d.String())
			} else {
				__antithesis_instrumentation__.Notify(621993)
			}
			__antithesis_instrumentation__.Notify(621989)
			d := time.Duration(f.Duration.Nanos()) * time.Nanosecond
			if err := setting.Validate(d); err != nil {
				__antithesis_instrumentation__.Notify(621994)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(621995)
			}
			__antithesis_instrumentation__.Notify(621990)
			return settings.EncodeDuration(d), nil
		} else {
			__antithesis_instrumentation__.Notify(621996)
		}
		__antithesis_instrumentation__.Notify(621937)
		return "", errors.Errorf("cannot use %s %T value for duration setting", d.ResolvedType(), d)
	default:
		__antithesis_instrumentation__.Notify(621938)
		return "", errors.Errorf("unsupported setting type %T", setting)
	}
}
