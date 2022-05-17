package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/bootstrap"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/systemschema"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

func rejectIfCantCoordinateMultiTenancy(codec keys.SQLCodec, op string) error {
	__antithesis_instrumentation__.Notify(628015)

	if !codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(628017)
		return pgerror.Newf(pgcode.InsufficientPrivilege,
			"only the system tenant can %s other tenants", op)
	} else {
		__antithesis_instrumentation__.Notify(628018)
	}
	__antithesis_instrumentation__.Notify(628016)
	return nil
}

func rejectIfSystemTenant(tenID uint64, op string) error {
	__antithesis_instrumentation__.Notify(628019)
	if roachpb.IsSystemTenantID(tenID) {
		__antithesis_instrumentation__.Notify(628021)
		return pgerror.Newf(pgcode.InvalidParameterValue,
			"cannot %s tenant \"%d\", ID assigned to system tenant", op, tenID)
	} else {
		__antithesis_instrumentation__.Notify(628022)
	}
	__antithesis_instrumentation__.Notify(628020)
	return nil
}

func CreateTenantRecord(
	ctx context.Context, execCfg *ExecutorConfig, txn *kv.Txn, info *descpb.TenantInfoWithUsage,
) error {
	__antithesis_instrumentation__.Notify(628023)
	const op = "create"
	if err := rejectIfCantCoordinateMultiTenancy(execCfg.Codec, op); err != nil {
		__antithesis_instrumentation__.Notify(628031)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628032)
	}
	__antithesis_instrumentation__.Notify(628024)
	if err := rejectIfSystemTenant(info.ID, op); err != nil {
		__antithesis_instrumentation__.Notify(628033)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628034)
	}
	__antithesis_instrumentation__.Notify(628025)

	tenID := info.ID
	active := info.State == descpb.TenantInfo_ACTIVE
	infoBytes, err := protoutil.Marshal(&info.TenantInfo)
	if err != nil {
		__antithesis_instrumentation__.Notify(628035)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628036)
	}
	__antithesis_instrumentation__.Notify(628026)

	if num, err := execCfg.InternalExecutor.ExecEx(
		ctx, "create-tenant", txn, sessiondata.NodeUserSessionDataOverride,
		`INSERT INTO system.tenants (id, active, info) VALUES ($1, $2, $3)`,
		tenID, active, infoBytes,
	); err != nil {
		__antithesis_instrumentation__.Notify(628037)
		if pgerror.GetPGCode(err) == pgcode.UniqueViolation {
			__antithesis_instrumentation__.Notify(628039)
			return pgerror.Newf(pgcode.DuplicateObject, "tenant \"%d\" already exists", tenID)
		} else {
			__antithesis_instrumentation__.Notify(628040)
		}
		__antithesis_instrumentation__.Notify(628038)
		return errors.Wrap(err, "inserting new tenant")
	} else {
		__antithesis_instrumentation__.Notify(628041)
		if num != 1 {
			__antithesis_instrumentation__.Notify(628042)
			log.Fatalf(ctx, "unexpected number of rows affected: %d", num)
		} else {
			__antithesis_instrumentation__.Notify(628043)
		}
	}
	__antithesis_instrumentation__.Notify(628027)

	if u := info.Usage; u != nil {
		__antithesis_instrumentation__.Notify(628044)
		consumption, err := protoutil.Marshal(&u.Consumption)
		if err != nil {
			__antithesis_instrumentation__.Notify(628046)
			return errors.Wrap(err, "marshaling tenant usage data")
		} else {
			__antithesis_instrumentation__.Notify(628047)
		}
		__antithesis_instrumentation__.Notify(628045)
		if num, err := execCfg.InternalExecutor.ExecEx(
			ctx, "create-tenant-usage", txn, sessiondata.NodeUserSessionDataOverride,
			`INSERT INTO system.tenant_usage (
			  tenant_id, instance_id, next_instance_id, last_update,
			  ru_burst_limit, ru_refill_rate, ru_current, current_share_sum,
			  total_consumption)
			VALUES (
				$1, 0, 0, now(),
				$2, $3, $4, 0, 
				$5)`,
			tenID,
			u.RUBurstLimit, u.RURefillRate, u.RUCurrent,
			tree.NewDBytes(tree.DBytes(consumption)),
		); err != nil {
			__antithesis_instrumentation__.Notify(628048)
			if pgerror.GetPGCode(err) == pgcode.UniqueViolation {
				__antithesis_instrumentation__.Notify(628050)
				return pgerror.Newf(pgcode.DuplicateObject, "tenant \"%d\" already has usage data", tenID)
			} else {
				__antithesis_instrumentation__.Notify(628051)
			}
			__antithesis_instrumentation__.Notify(628049)
			return errors.Wrap(err, "inserting tenant usage data")
		} else {
			__antithesis_instrumentation__.Notify(628052)
			if num != 1 {
				__antithesis_instrumentation__.Notify(628053)
				log.Fatalf(ctx, "unexpected number of rows affected: %d", num)
			} else {
				__antithesis_instrumentation__.Notify(628054)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(628055)
	}
	__antithesis_instrumentation__.Notify(628028)

	if !execCfg.Settings.Version.IsActive(ctx, clusterversion.PreSeedTenantSpanConfigs) {
		__antithesis_instrumentation__.Notify(628056)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(628057)
	}
	__antithesis_instrumentation__.Notify(628029)

	tenantSpanConfig := execCfg.DefaultZoneConfig.AsSpanConfig()

	tenantSpanConfig.RangefeedEnabled = true

	tenantSpanConfig.GCPolicy.IgnoreStrictEnforcement = true

	tenantPrefix := keys.MakeTenantPrefix(roachpb.MakeTenantID(tenID))
	record, err := spanconfig.MakeRecord(spanconfig.MakeTargetFromSpan(roachpb.Span{
		Key:    tenantPrefix,
		EndKey: tenantPrefix.Next(),
	}), tenantSpanConfig)
	if err != nil {
		__antithesis_instrumentation__.Notify(628058)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628059)
	}
	__antithesis_instrumentation__.Notify(628030)
	toUpsert := []spanconfig.Record{record}
	scKVAccessor := execCfg.SpanConfigKVAccessor.WithTxn(ctx, txn)
	return scKVAccessor.UpdateSpanConfigRecords(
		ctx, nil, toUpsert, hlc.MinTimestamp, hlc.MaxTimestamp,
	)
}

func GetTenantRecord(
	ctx context.Context, execCfg *ExecutorConfig, txn *kv.Txn, tenID uint64,
) (*descpb.TenantInfo, error) {
	__antithesis_instrumentation__.Notify(628060)
	row, err := execCfg.InternalExecutor.QueryRowEx(
		ctx, "activate-tenant", txn, sessiondata.NodeUserSessionDataOverride,
		`SELECT info FROM system.tenants WHERE id = $1`, tenID,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(628063)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(628064)
		if row == nil {
			__antithesis_instrumentation__.Notify(628065)
			return nil, pgerror.Newf(pgcode.UndefinedObject, "tenant \"%d\" does not exist", tenID)
		} else {
			__antithesis_instrumentation__.Notify(628066)
		}
	}
	__antithesis_instrumentation__.Notify(628061)

	info := &descpb.TenantInfo{}
	infoBytes := []byte(tree.MustBeDBytes(row[0]))
	if err := protoutil.Unmarshal(infoBytes, info); err != nil {
		__antithesis_instrumentation__.Notify(628067)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(628068)
	}
	__antithesis_instrumentation__.Notify(628062)
	return info, nil
}

func updateTenantRecord(
	ctx context.Context, execCfg *ExecutorConfig, txn *kv.Txn, info *descpb.TenantInfo,
) error {
	__antithesis_instrumentation__.Notify(628069)
	tenID := info.ID
	active := info.State == descpb.TenantInfo_ACTIVE
	infoBytes, err := protoutil.Marshal(info)
	if err != nil {
		__antithesis_instrumentation__.Notify(628072)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628073)
	}
	__antithesis_instrumentation__.Notify(628070)

	if num, err := execCfg.InternalExecutor.ExecEx(
		ctx, "activate-tenant", txn, sessiondata.NodeUserSessionDataOverride,
		`UPDATE system.tenants SET active = $2, info = $3 WHERE id = $1`,
		tenID, active, infoBytes,
	); err != nil {
		__antithesis_instrumentation__.Notify(628074)
		return errors.Wrap(err, "activating tenant")
	} else {
		__antithesis_instrumentation__.Notify(628075)
		if num != 1 {
			__antithesis_instrumentation__.Notify(628076)
			log.Fatalf(ctx, "unexpected number of rows affected: %d", num)
		} else {
			__antithesis_instrumentation__.Notify(628077)
		}
	}
	__antithesis_instrumentation__.Notify(628071)
	return nil
}

func (p *planner) CreateTenant(ctx context.Context, tenID uint64) error {
	__antithesis_instrumentation__.Notify(628078)
	info := &descpb.TenantInfoWithUsage{
		TenantInfo: descpb.TenantInfo{
			ID: tenID,

			State: descpb.TenantInfo_ACTIVE,
		},
	}
	if err := CreateTenantRecord(ctx, p.ExecCfg(), p.Txn(), info); err != nil {
		__antithesis_instrumentation__.Notify(628084)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628085)
	}
	__antithesis_instrumentation__.Notify(628079)

	codec := keys.MakeSQLCodec(roachpb.MakeTenantID(tenID))

	schema := bootstrap.MakeMetadataSchema(
		codec,
		p.ExtendedEvalContext().ExecCfg.DefaultZoneConfig,
		nil,
	)
	kvs, splits := schema.GetInitialValues()

	{
		__antithesis_instrumentation__.Notify(628086)

		v := p.EvalContext().Settings.Version.ActiveVersion(ctx)
		tenantSettingKV, err := generateTenantClusterSettingKV(codec, v)
		if err != nil {
			__antithesis_instrumentation__.Notify(628088)
			return err
		} else {
			__antithesis_instrumentation__.Notify(628089)
		}
		__antithesis_instrumentation__.Notify(628087)
		kvs = append(kvs, tenantSettingKV)
	}
	__antithesis_instrumentation__.Notify(628080)

	b := p.Txn().NewBatch()
	for _, kv := range kvs {
		__antithesis_instrumentation__.Notify(628090)
		b.CPut(kv.Key, &kv.Value, nil)
	}
	__antithesis_instrumentation__.Notify(628081)
	if err := p.Txn().Run(ctx, b); err != nil {
		__antithesis_instrumentation__.Notify(628091)
		if errors.HasType(err, (*roachpb.ConditionFailedError)(nil)) {
			__antithesis_instrumentation__.Notify(628093)
			return errors.Wrap(err, "programming error: "+
				"tenant already exists but was not in system.tenants table")
		} else {
			__antithesis_instrumentation__.Notify(628094)
		}
		__antithesis_instrumentation__.Notify(628092)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628095)
	}
	__antithesis_instrumentation__.Notify(628082)

	expTime := p.ExecCfg().Clock.Now().Add(time.Hour.Nanoseconds(), 0)
	for _, key := range splits {
		__antithesis_instrumentation__.Notify(628096)
		if err := p.ExecCfg().DB.AdminSplit(ctx, key, expTime); err != nil {
			__antithesis_instrumentation__.Notify(628097)
			return err
		} else {
			__antithesis_instrumentation__.Notify(628098)
		}
	}
	__antithesis_instrumentation__.Notify(628083)

	return nil
}

func generateTenantClusterSettingKV(
	codec keys.SQLCodec, v clusterversion.ClusterVersion,
) (roachpb.KeyValue, error) {
	__antithesis_instrumentation__.Notify(628099)
	encoded, err := protoutil.Marshal(&v)
	if err != nil {
		__antithesis_instrumentation__.Notify(628103)
		return roachpb.KeyValue{}, errors.NewAssertionErrorWithWrappedErrf(err,
			"failed to encode current cluster version %v", &v)
	} else {
		__antithesis_instrumentation__.Notify(628104)
	}
	__antithesis_instrumentation__.Notify(628100)
	kvs, err := rowenc.EncodePrimaryIndex(
		codec,
		systemschema.SettingsTable,
		systemschema.SettingsTable.GetPrimaryIndex(),
		catalog.ColumnIDToOrdinalMap(systemschema.SettingsTable.PublicColumns()),
		[]tree.Datum{
			tree.NewDString(clusterversion.KeyVersionSetting),
			tree.NewDString(string(encoded)),
			tree.NewDTimeTZFromTime(timeutil.Now()),
			tree.NewDString((*settings.VersionSetting)(nil).Typ()),
		},
		false,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(628105)
		return roachpb.KeyValue{}, errors.NewAssertionErrorWithWrappedErrf(err,
			"failed to encode cluster setting")
	} else {
		__antithesis_instrumentation__.Notify(628106)
	}
	__antithesis_instrumentation__.Notify(628101)
	if len(kvs) != 1 {
		__antithesis_instrumentation__.Notify(628107)
		return roachpb.KeyValue{}, errors.AssertionFailedf(
			"failed to encode cluster setting: expected 1 key-value, got %d", len(kvs))
	} else {
		__antithesis_instrumentation__.Notify(628108)
	}
	__antithesis_instrumentation__.Notify(628102)
	return roachpb.KeyValue{
		Key:   kvs[0].Key,
		Value: kvs[0].Value,
	}, nil
}

func ActivateTenant(ctx context.Context, execCfg *ExecutorConfig, txn *kv.Txn, tenID uint64) error {
	__antithesis_instrumentation__.Notify(628109)
	const op = "activate"
	if err := rejectIfCantCoordinateMultiTenancy(execCfg.Codec, op); err != nil {
		__antithesis_instrumentation__.Notify(628114)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628115)
	}
	__antithesis_instrumentation__.Notify(628110)
	if err := rejectIfSystemTenant(tenID, op); err != nil {
		__antithesis_instrumentation__.Notify(628116)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628117)
	}
	__antithesis_instrumentation__.Notify(628111)

	info, err := GetTenantRecord(ctx, execCfg, txn, tenID)
	if err != nil {
		__antithesis_instrumentation__.Notify(628118)
		return errors.Wrap(err, "activating tenant")
	} else {
		__antithesis_instrumentation__.Notify(628119)
	}
	__antithesis_instrumentation__.Notify(628112)

	info.State = descpb.TenantInfo_ACTIVE
	if err := updateTenantRecord(ctx, execCfg, txn, info); err != nil {
		__antithesis_instrumentation__.Notify(628120)
		return errors.Wrap(err, "activating tenant")
	} else {
		__antithesis_instrumentation__.Notify(628121)
	}
	__antithesis_instrumentation__.Notify(628113)

	return nil
}

func clearTenant(ctx context.Context, execCfg *ExecutorConfig, info *descpb.TenantInfo) error {
	__antithesis_instrumentation__.Notify(628122)

	if info.State != descpb.TenantInfo_DROP {
		__antithesis_instrumentation__.Notify(628124)
		return errors.Errorf("tenant %d is not in state DROP", info.ID)
	} else {
		__antithesis_instrumentation__.Notify(628125)
	}
	__antithesis_instrumentation__.Notify(628123)

	log.Infof(ctx, "clearing data for tenant %d", info.ID)

	prefix := keys.MakeTenantPrefix(roachpb.MakeTenantID(info.ID))
	prefixEnd := prefix.PrefixEnd()

	log.VEventf(ctx, 2, "ClearRange %s - %s", prefix, prefixEnd)

	b := &kv.Batch{}
	b.AddRawRequest(&roachpb.ClearRangeRequest{
		RequestHeader: roachpb.RequestHeader{Key: prefix, EndKey: prefixEnd},
	})

	return errors.Wrapf(execCfg.DB.Run(ctx, b), "clearing tenant %d data", info.ID)
}

func (p *planner) DestroyTenant(ctx context.Context, tenID uint64, synchronous bool) error {
	__antithesis_instrumentation__.Notify(628126)
	const op = "destroy"
	if err := rejectIfCantCoordinateMultiTenancy(p.execCfg.Codec, op); err != nil {
		__antithesis_instrumentation__.Notify(628134)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628135)
	}
	__antithesis_instrumentation__.Notify(628127)
	if err := rejectIfSystemTenant(tenID, op); err != nil {
		__antithesis_instrumentation__.Notify(628136)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628137)
	}
	__antithesis_instrumentation__.Notify(628128)

	info, err := GetTenantRecord(ctx, p.execCfg, p.txn, tenID)
	if err != nil {
		__antithesis_instrumentation__.Notify(628138)
		return errors.Wrap(err, "destroying tenant")
	} else {
		__antithesis_instrumentation__.Notify(628139)
	}
	__antithesis_instrumentation__.Notify(628129)

	if info.State == descpb.TenantInfo_DROP {
		__antithesis_instrumentation__.Notify(628140)
		return errors.Errorf("tenant %d is already in state DROP", tenID)
	} else {
		__antithesis_instrumentation__.Notify(628141)
	}
	__antithesis_instrumentation__.Notify(628130)

	info.State = descpb.TenantInfo_DROP
	if err := updateTenantRecord(ctx, p.execCfg, p.txn, info); err != nil {
		__antithesis_instrumentation__.Notify(628142)
		return errors.Wrap(err, "destroying tenant")
	} else {
		__antithesis_instrumentation__.Notify(628143)
	}
	__antithesis_instrumentation__.Notify(628131)

	jobID, err := gcTenantJob(ctx, p.execCfg, p.txn, p.User(), tenID, synchronous)
	if err != nil {
		__antithesis_instrumentation__.Notify(628144)
		return errors.Wrap(err, "scheduling gc job")
	} else {
		__antithesis_instrumentation__.Notify(628145)
	}
	__antithesis_instrumentation__.Notify(628132)
	if synchronous {
		__antithesis_instrumentation__.Notify(628146)
		p.extendedEvalCtx.Jobs.add(jobID)
	} else {
		__antithesis_instrumentation__.Notify(628147)
	}
	__antithesis_instrumentation__.Notify(628133)
	return nil
}

func GCTenantSync(ctx context.Context, execCfg *ExecutorConfig, info *descpb.TenantInfo) error {
	__antithesis_instrumentation__.Notify(628148)
	const op = "gc"
	if err := rejectIfCantCoordinateMultiTenancy(execCfg.Codec, op); err != nil {
		__antithesis_instrumentation__.Notify(628153)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628154)
	}
	__antithesis_instrumentation__.Notify(628149)
	if err := rejectIfSystemTenant(info.ID, op); err != nil {
		__antithesis_instrumentation__.Notify(628155)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628156)
	}
	__antithesis_instrumentation__.Notify(628150)

	if err := clearTenant(ctx, execCfg, info); err != nil {
		__antithesis_instrumentation__.Notify(628157)
		return errors.Wrap(err, "clear tenant")
	} else {
		__antithesis_instrumentation__.Notify(628158)
	}
	__antithesis_instrumentation__.Notify(628151)

	err := execCfg.DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(628159)
		if num, err := execCfg.InternalExecutor.ExecEx(
			ctx, "delete-tenant", txn, sessiondata.NodeUserSessionDataOverride,
			`DELETE FROM system.tenants WHERE id = $1`, info.ID,
		); err != nil {
			__antithesis_instrumentation__.Notify(628167)
			return errors.Wrapf(err, "deleting tenant %d", info.ID)
		} else {
			__antithesis_instrumentation__.Notify(628168)
			if num != 1 {
				__antithesis_instrumentation__.Notify(628169)
				log.Fatalf(ctx, "unexpected number of rows affected: %d", num)
			} else {
				__antithesis_instrumentation__.Notify(628170)
			}
		}
		__antithesis_instrumentation__.Notify(628160)

		if _, err := execCfg.InternalExecutor.ExecEx(
			ctx, "delete-tenant-usage", txn, sessiondata.NodeUserSessionDataOverride,
			`DELETE FROM system.tenant_usage WHERE tenant_id = $1`, info.ID,
		); err != nil {
			__antithesis_instrumentation__.Notify(628171)
			return errors.Wrapf(err, "deleting tenant %d usage", info.ID)
		} else {
			__antithesis_instrumentation__.Notify(628172)
		}
		__antithesis_instrumentation__.Notify(628161)

		if execCfg.Settings.Version.IsActive(ctx, clusterversion.TenantSettingsTable) {
			__antithesis_instrumentation__.Notify(628173)
			if _, err := execCfg.InternalExecutor.ExecEx(
				ctx, "delete-tenant-settings", txn, sessiondata.NodeUserSessionDataOverride,
				`DELETE FROM system.tenant_settings WHERE tenant_id = $1`, info.ID,
			); err != nil {
				__antithesis_instrumentation__.Notify(628174)
				return errors.Wrapf(err, "deleting tenant %d settings", info.ID)
			} else {
				__antithesis_instrumentation__.Notify(628175)
			}
		} else {
			__antithesis_instrumentation__.Notify(628176)
		}
		__antithesis_instrumentation__.Notify(628162)

		if !execCfg.Settings.Version.IsActive(ctx, clusterversion.PreSeedTenantSpanConfigs) {
			__antithesis_instrumentation__.Notify(628177)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(628178)
		}
		__antithesis_instrumentation__.Notify(628163)

		tenID := roachpb.MakeTenantID(info.ID)
		tenantPrefix := keys.MakeTenantPrefix(tenID)
		tenantSpan := roachpb.Span{
			Key:    tenantPrefix,
			EndKey: tenantPrefix.PrefixEnd(),
		}

		systemTarget, err := spanconfig.MakeTenantKeyspaceTarget(tenID, tenID)
		if err != nil {
			__antithesis_instrumentation__.Notify(628179)
			return err
		} else {
			__antithesis_instrumentation__.Notify(628180)
		}
		__antithesis_instrumentation__.Notify(628164)
		scKVAccessor := execCfg.SpanConfigKVAccessor.WithTxn(ctx, txn)
		records, err := scKVAccessor.GetSpanConfigRecords(
			ctx, []spanconfig.Target{
				spanconfig.MakeTargetFromSpan(tenantSpan),
				spanconfig.MakeTargetFromSystemTarget(systemTarget),
			},
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(628181)
			return err
		} else {
			__antithesis_instrumentation__.Notify(628182)
		}
		__antithesis_instrumentation__.Notify(628165)

		toDelete := make([]spanconfig.Target, len(records))
		for i, record := range records {
			__antithesis_instrumentation__.Notify(628183)
			toDelete[i] = record.GetTarget()
		}
		__antithesis_instrumentation__.Notify(628166)
		return scKVAccessor.UpdateSpanConfigRecords(
			ctx, toDelete, nil, hlc.MinTimestamp, hlc.MaxTimestamp,
		)
	})
	__antithesis_instrumentation__.Notify(628152)
	return errors.Wrapf(err, "deleting tenant %d record", info.ID)
}

func gcTenantJob(
	ctx context.Context,
	execCfg *ExecutorConfig,
	txn *kv.Txn,
	user security.SQLUsername,
	tenID uint64,
	synchronous bool,
) (jobspb.JobID, error) {
	__antithesis_instrumentation__.Notify(628184)

	gcDetails := jobspb.SchemaChangeGCDetails{}
	gcDetails.Tenant = &jobspb.SchemaChangeGCDetails_DroppedTenant{
		ID:       tenID,
		DropTime: timeutil.Now().UnixNano(),
	}
	progress := jobspb.SchemaChangeGCProgress{}
	if synchronous {
		__antithesis_instrumentation__.Notify(628187)
		progress.Tenant = &jobspb.SchemaChangeGCProgress_TenantProgress{
			Status: jobspb.SchemaChangeGCProgress_DELETING,
		}
	} else {
		__antithesis_instrumentation__.Notify(628188)
	}
	__antithesis_instrumentation__.Notify(628185)
	gcJobRecord := jobs.Record{
		Description:   fmt.Sprintf("GC for tenant %d", tenID),
		Username:      user,
		Details:       gcDetails,
		Progress:      progress,
		NonCancelable: true,
	}
	jobID := execCfg.JobRegistry.MakeJobID()
	if _, err := execCfg.JobRegistry.CreateJobWithTxn(
		ctx, gcJobRecord, jobID, txn,
	); err != nil {
		__antithesis_instrumentation__.Notify(628189)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(628190)
	}
	__antithesis_instrumentation__.Notify(628186)
	return jobID, nil
}

func (p *planner) GCTenant(ctx context.Context, tenID uint64) error {
	__antithesis_instrumentation__.Notify(628191)

	if !p.extendedEvalCtx.TxnIsSingleStmt {
		__antithesis_instrumentation__.Notify(628195)
		return errors.Errorf("gc_tenant cannot be used inside a multi-statement transaction")
	} else {
		__antithesis_instrumentation__.Notify(628196)
	}
	__antithesis_instrumentation__.Notify(628192)
	var info *descpb.TenantInfo
	if txnErr := p.ExecCfg().DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(628197)
		var err error
		info, err = GetTenantRecord(ctx, p.execCfg, p.txn, tenID)
		return err
	}); txnErr != nil {
		__antithesis_instrumentation__.Notify(628198)
		return errors.Wrapf(txnErr, "retrieving tenant %d", tenID)
	} else {
		__antithesis_instrumentation__.Notify(628199)
	}
	__antithesis_instrumentation__.Notify(628193)

	if info.State != descpb.TenantInfo_DROP {
		__antithesis_instrumentation__.Notify(628200)
		return errors.Errorf("tenant %d is not in state DROP", info.ID)
	} else {
		__antithesis_instrumentation__.Notify(628201)
	}
	__antithesis_instrumentation__.Notify(628194)

	_, err := gcTenantJob(
		ctx, p.ExecCfg(), p.Txn(), p.User(), tenID, false,
	)
	return err
}

func (p *planner) UpdateTenantResourceLimits(
	ctx context.Context,
	tenantID uint64,
	availableRU float64,
	refillRate float64,
	maxBurstRU float64,
	asOf time.Time,
	asOfConsumedRequestUnits float64,
) error {
	__antithesis_instrumentation__.Notify(628202)
	const op = "update-resource-limits"
	if err := rejectIfCantCoordinateMultiTenancy(p.execCfg.Codec, op); err != nil {
		__antithesis_instrumentation__.Notify(628205)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628206)
	}
	__antithesis_instrumentation__.Notify(628203)
	if err := rejectIfSystemTenant(tenantID, op); err != nil {
		__antithesis_instrumentation__.Notify(628207)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628208)
	}
	__antithesis_instrumentation__.Notify(628204)
	return p.ExecCfg().TenantUsageServer.ReconfigureTokenBucket(
		ctx, p.Txn(), roachpb.MakeTenantID(tenantID),
		availableRU, refillRate, maxBurstRU, asOf, asOfConsumedRequestUnits,
	)
}

func TestingUpdateTenantRecord(
	ctx context.Context, execCfg *ExecutorConfig, txn *kv.Txn, info *descpb.TenantInfo,
) error {
	__antithesis_instrumentation__.Notify(628209)
	return updateTenantRecord(ctx, execCfg, txn, info)
}
