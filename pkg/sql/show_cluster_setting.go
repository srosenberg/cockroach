package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

func (p *planner) getCurrentEncodedVersionSettingValue(
	ctx context.Context, s *settings.VersionSetting, name string,
) (string, error) {
	__antithesis_instrumentation__.Notify(622700)
	st := p.ExecCfg().Settings
	var res string

	if err := contextutil.RunWithTimeout(ctx, fmt.Sprintf("show cluster setting %s", name), 2*time.Minute,
		func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(622702)
			tBegin := timeutil.Now()

			return retry.WithMaxAttempts(ctx, retry.Options{}, math.MaxInt32, func() error {
				__antithesis_instrumentation__.Notify(622703)
				return p.execCfg.DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
					__antithesis_instrumentation__.Notify(622704)
					datums, err := p.ExtendedEvalContext().ExecCfg.InternalExecutor.QueryRowEx(
						ctx, "read-setting",
						txn,
						sessiondata.InternalExecutorOverride{User: security.RootUserName()},
						"SELECT value FROM system.settings WHERE name = $1", name,
					)
					if err != nil {
						__antithesis_instrumentation__.Notify(622708)
						return err
					} else {
						__antithesis_instrumentation__.Notify(622709)
					}
					__antithesis_instrumentation__.Notify(622705)
					var kvRawVal []byte
					if len(datums) != 0 {
						__antithesis_instrumentation__.Notify(622710)
						dStr, ok := datums[0].(*tree.DString)
						if !ok {
							__antithesis_instrumentation__.Notify(622712)
							return errors.New("the existing value is not a string")
						} else {
							__antithesis_instrumentation__.Notify(622713)
						}
						__antithesis_instrumentation__.Notify(622711)
						kvRawVal = []byte(string(*dStr))
					} else {
						__antithesis_instrumentation__.Notify(622714)
						if !p.execCfg.Codec.ForSystemTenant() {
							__antithesis_instrumentation__.Notify(622715)

							tenantDefaultVersion := clusterversion.ClusterVersion{
								Version: roachpb.Version{Major: 20, Minor: 2},
							}
							encoded, err := protoutil.Marshal(&tenantDefaultVersion)
							if err != nil {
								__antithesis_instrumentation__.Notify(622717)
								return errors.WithAssertionFailure(err)
							} else {
								__antithesis_instrumentation__.Notify(622718)
							}
							__antithesis_instrumentation__.Notify(622716)
							kvRawVal = encoded
						} else {
							__antithesis_instrumentation__.Notify(622719)

							return errors.AssertionFailedf("no value found for version setting")
						}
					}
					__antithesis_instrumentation__.Notify(622706)

					localRawVal := []byte(s.Get(&st.SV))
					if !bytes.Equal(localRawVal, kvRawVal) {
						__antithesis_instrumentation__.Notify(622720)

						return errors.Errorf(
							"value differs between local setting (%v) and KV (%v); try again later (%v after %s)",
							localRawVal, kvRawVal, ctx.Err(), timeutil.Since(tBegin))
					} else {
						__antithesis_instrumentation__.Notify(622721)
					}
					__antithesis_instrumentation__.Notify(622707)

					res = string(kvRawVal)
					return nil
				})
			})
		}); err != nil {
		__antithesis_instrumentation__.Notify(622722)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(622723)
	}
	__antithesis_instrumentation__.Notify(622701)

	return res, nil
}

func (p *planner) ShowClusterSetting(
	ctx context.Context, n *tree.ShowClusterSetting,
) (planNode, error) {
	__antithesis_instrumentation__.Notify(622724)
	name := strings.ToLower(n.Name)
	val, ok := settings.Lookup(
		name, settings.LookupForLocalAccess, p.ExecCfg().Codec.ForSystemTenant(),
	)
	if !ok {
		__antithesis_instrumentation__.Notify(622729)
		return nil, errors.Errorf("unknown setting: %q", name)
	} else {
		__antithesis_instrumentation__.Notify(622730)
	}
	__antithesis_instrumentation__.Notify(622725)

	if err := checkPrivilegesForSetting(ctx, p, name, "show"); err != nil {
		__antithesis_instrumentation__.Notify(622731)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(622732)
	}
	__antithesis_instrumentation__.Notify(622726)

	setting, ok := val.(settings.NonMaskedSetting)
	if !ok {
		__antithesis_instrumentation__.Notify(622733)
		return nil, errors.AssertionFailedf("setting is masked: %v", name)
	} else {
		__antithesis_instrumentation__.Notify(622734)
	}
	__antithesis_instrumentation__.Notify(622727)

	columns, err := getShowClusterSettingPlanColumns(setting, name)
	if err != nil {
		__antithesis_instrumentation__.Notify(622735)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(622736)
	}
	__antithesis_instrumentation__.Notify(622728)

	return planShowClusterSetting(setting, name, columns,
		func(ctx context.Context, p *planner) (bool, string, error) {
			__antithesis_instrumentation__.Notify(622737)
			if verSetting, ok := setting.(*settings.VersionSetting); ok {
				__antithesis_instrumentation__.Notify(622739)
				encoded, err := p.getCurrentEncodedVersionSettingValue(ctx, verSetting, name)
				return true, encoded, err
			} else {
				__antithesis_instrumentation__.Notify(622740)
			}
			__antithesis_instrumentation__.Notify(622738)
			return true, setting.Encoded(&p.ExecCfg().Settings.SV), nil
		},
	)
}

func getShowClusterSettingPlanColumns(
	val settings.NonMaskedSetting, name string,
) (colinfo.ResultColumns, error) {
	__antithesis_instrumentation__.Notify(622741)
	var dType *types.T
	switch val.(type) {
	case *settings.IntSetting:
		__antithesis_instrumentation__.Notify(622743)
		dType = types.Int
	case *settings.StringSetting, *settings.ByteSizeSetting, *settings.VersionSetting, *settings.EnumSetting:
		__antithesis_instrumentation__.Notify(622744)
		dType = types.String
	case *settings.BoolSetting:
		__antithesis_instrumentation__.Notify(622745)
		dType = types.Bool
	case *settings.FloatSetting:
		__antithesis_instrumentation__.Notify(622746)
		dType = types.Float
	case *settings.DurationSetting:
		__antithesis_instrumentation__.Notify(622747)
		dType = types.Interval
	case *settings.DurationSettingWithExplicitUnit:
		__antithesis_instrumentation__.Notify(622748)
		dType = types.Interval
	default:
		__antithesis_instrumentation__.Notify(622749)
		return nil, errors.Errorf("unknown setting type for %s: %s", name, val.Typ())
	}
	__antithesis_instrumentation__.Notify(622742)
	return colinfo.ResultColumns{{Name: name, Typ: dType}}, nil
}

func planShowClusterSetting(
	val settings.NonMaskedSetting,
	name string,
	columns colinfo.ResultColumns,
	getEncodedValue func(ctx context.Context, p *planner) (bool, string, error),
) (planNode, error) {
	__antithesis_instrumentation__.Notify(622750)
	return &delayedNode{
		name:    "SHOW CLUSTER SETTING " + name,
		columns: columns,
		constructor: func(ctx context.Context, p *planner) (planNode, error) {
			__antithesis_instrumentation__.Notify(622751)
			isNotNull, encoded, err := getEncodedValue(ctx, p)
			if err != nil {
				__antithesis_instrumentation__.Notify(622755)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(622756)
			}
			__antithesis_instrumentation__.Notify(622752)

			var d tree.Datum
			d = tree.DNull
			if isNotNull {
				__antithesis_instrumentation__.Notify(622757)
				switch s := val.(type) {
				case *settings.IntSetting:
					__antithesis_instrumentation__.Notify(622758)
					v, err := s.DecodeValue(encoded)
					if err != nil {
						__antithesis_instrumentation__.Notify(622771)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(622772)
					}
					__antithesis_instrumentation__.Notify(622759)
					d = tree.NewDInt(tree.DInt(v))
				case *settings.StringSetting, *settings.EnumSetting,
					*settings.ByteSizeSetting, *settings.VersionSetting:
					__antithesis_instrumentation__.Notify(622760)
					v, err := val.DecodeToString(encoded)
					if err != nil {
						__antithesis_instrumentation__.Notify(622773)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(622774)
					}
					__antithesis_instrumentation__.Notify(622761)
					d = tree.NewDString(v)
				case *settings.BoolSetting:
					__antithesis_instrumentation__.Notify(622762)
					v, err := s.DecodeValue(encoded)
					if err != nil {
						__antithesis_instrumentation__.Notify(622775)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(622776)
					}
					__antithesis_instrumentation__.Notify(622763)
					d = tree.MakeDBool(tree.DBool(v))
				case *settings.FloatSetting:
					__antithesis_instrumentation__.Notify(622764)
					v, err := s.DecodeValue(encoded)
					if err != nil {
						__antithesis_instrumentation__.Notify(622777)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(622778)
					}
					__antithesis_instrumentation__.Notify(622765)
					d = tree.NewDFloat(tree.DFloat(v))
				case *settings.DurationSetting:
					__antithesis_instrumentation__.Notify(622766)
					v, err := s.DecodeValue(encoded)
					if err != nil {
						__antithesis_instrumentation__.Notify(622779)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(622780)
					}
					__antithesis_instrumentation__.Notify(622767)
					d = &tree.DInterval{Duration: duration.MakeDuration(v.Nanoseconds(), 0, 0)}
				case *settings.DurationSettingWithExplicitUnit:
					__antithesis_instrumentation__.Notify(622768)
					v, err := s.DecodeValue(encoded)
					if err != nil {
						__antithesis_instrumentation__.Notify(622781)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(622782)
					}
					__antithesis_instrumentation__.Notify(622769)
					d = &tree.DInterval{Duration: duration.MakeDuration(v.Nanoseconds(), 0, 0)}
				default:
					__antithesis_instrumentation__.Notify(622770)
					return nil, errors.AssertionFailedf("unknown setting type for %s: %s (%T)", name, val.Typ(), val)
				}
			} else {
				__antithesis_instrumentation__.Notify(622783)
			}
			__antithesis_instrumentation__.Notify(622753)

			v := p.newContainerValuesNode(columns, 0)
			if _, err := v.rows.AddRow(ctx, tree.Datums{d}); err != nil {
				__antithesis_instrumentation__.Notify(622784)
				v.rows.Close(ctx)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(622785)
			}
			__antithesis_instrumentation__.Notify(622754)
			return v, nil
		},
	}, nil
}
