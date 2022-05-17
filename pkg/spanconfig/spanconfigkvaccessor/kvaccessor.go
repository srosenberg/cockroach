package spanconfigkvaccessor

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
)

var batchSizeSetting = settings.RegisterIntSetting(
	settings.SystemOnly,
	"spanconfig.kvaccessor.batch_size",
	`number of span config records to access in a single batch`,
	10000,
)

type KVAccessor struct {
	db *kv.DB
	ie sqlutil.InternalExecutor

	optionalTxn *kv.Txn
	settings    *cluster.Settings
	clock       *hlc.Clock

	configurationsTableFQN string

	knobs *spanconfig.TestingKnobs
}

var _ spanconfig.KVAccessor = &KVAccessor{}

func New(
	db *kv.DB,
	ie sqlutil.InternalExecutor,
	settings *cluster.Settings,
	clock *hlc.Clock,
	configurationsTableFQN string,
	knobs *spanconfig.TestingKnobs,
) *KVAccessor {
	__antithesis_instrumentation__.Notify(240386)
	if _, err := parser.ParseQualifiedTableName(configurationsTableFQN); err != nil {
		__antithesis_instrumentation__.Notify(240388)
		panic(fmt.Sprintf("unabled to parse configurations table FQN: %s", configurationsTableFQN))
	} else {
		__antithesis_instrumentation__.Notify(240389)
	}
	__antithesis_instrumentation__.Notify(240387)

	return newKVAccessor(db, ie, settings, clock, configurationsTableFQN, knobs, nil)
}

func (k *KVAccessor) WithTxn(ctx context.Context, txn *kv.Txn) spanconfig.KVAccessor {
	__antithesis_instrumentation__.Notify(240390)
	if k.optionalTxn != nil {
		__antithesis_instrumentation__.Notify(240392)
		log.Fatalf(ctx, "KVAccessor already scoped to txn (was .WithTxn(...) chained multiple times?)")
	} else {
		__antithesis_instrumentation__.Notify(240393)
	}
	__antithesis_instrumentation__.Notify(240391)
	return newKVAccessor(k.db, k.ie, k.settings, k.clock, k.configurationsTableFQN, k.knobs, txn)
}

func (k *KVAccessor) GetAllSystemSpanConfigsThatApply(
	ctx context.Context, id roachpb.TenantID,
) (spanConfigs []roachpb.SpanConfig, _ error) {
	__antithesis_instrumentation__.Notify(240394)
	hostSetOnTenant, err := spanconfig.MakeTenantKeyspaceTarget(
		roachpb.SystemTenantID, id,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(240399)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(240400)
	}
	__antithesis_instrumentation__.Notify(240395)

	targets := []spanconfig.Target{
		spanconfig.MakeTargetFromSystemTarget(spanconfig.MakeEntireKeyspaceTarget()),
		spanconfig.MakeTargetFromSystemTarget(hostSetOnTenant),
	}

	if id != roachpb.SystemTenantID {
		__antithesis_instrumentation__.Notify(240401)
		target, err := spanconfig.MakeTenantKeyspaceTarget(id, id)
		if err != nil {
			__antithesis_instrumentation__.Notify(240403)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(240404)
		}
		__antithesis_instrumentation__.Notify(240402)
		targets = append(targets, spanconfig.MakeTargetFromSystemTarget(target))
	} else {
		__antithesis_instrumentation__.Notify(240405)
	}
	__antithesis_instrumentation__.Notify(240396)

	records, err := k.GetSpanConfigRecords(ctx, targets)
	if err != nil {
		__antithesis_instrumentation__.Notify(240406)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(240407)
	}
	__antithesis_instrumentation__.Notify(240397)

	for _, record := range records {
		__antithesis_instrumentation__.Notify(240408)
		spanConfigs = append(spanConfigs, record.GetConfig())
	}
	__antithesis_instrumentation__.Notify(240398)

	return spanConfigs, nil
}

func (k *KVAccessor) GetSpanConfigRecords(
	ctx context.Context, targets []spanconfig.Target,
) ([]spanconfig.Record, error) {
	__antithesis_instrumentation__.Notify(240409)
	if k.optionalTxn != nil {
		__antithesis_instrumentation__.Notify(240412)
		return k.getSpanConfigRecordsWithTxn(ctx, targets, k.optionalTxn)
	} else {
		__antithesis_instrumentation__.Notify(240413)
	}
	__antithesis_instrumentation__.Notify(240410)

	var records []spanconfig.Record
	if err := k.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(240414)
		var err error
		records, err = k.getSpanConfigRecordsWithTxn(ctx, targets, txn)
		return err
	}); err != nil {
		__antithesis_instrumentation__.Notify(240415)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(240416)
	}
	__antithesis_instrumentation__.Notify(240411)

	return records, nil
}

func (k *KVAccessor) UpdateSpanConfigRecords(
	ctx context.Context,
	toDelete []spanconfig.Target,
	toUpsert []spanconfig.Record,
	minCommitTS, maxCommitTS hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(240417)
	if k.optionalTxn != nil {
		__antithesis_instrumentation__.Notify(240421)
		return k.updateSpanConfigRecordsWithTxn(ctx, toDelete, toUpsert, k.optionalTxn, minCommitTS, maxCommitTS)
	} else {
		__antithesis_instrumentation__.Notify(240422)
	}
	__antithesis_instrumentation__.Notify(240418)

	if fn := k.knobs.KVAccessorPreCommitMinTSWaitInterceptor; fn != nil {
		__antithesis_instrumentation__.Notify(240423)
		fn()
	} else {
		__antithesis_instrumentation__.Notify(240424)
	}
	__antithesis_instrumentation__.Notify(240419)
	if err := k.clock.SleepUntil(ctx, minCommitTS); err != nil {
		__antithesis_instrumentation__.Notify(240425)
		return errors.Wrapf(err, "waiting for clock to be in advance of minimum commit timestamp (%s)",
			minCommitTS)
	} else {
		__antithesis_instrumentation__.Notify(240426)
	}
	__antithesis_instrumentation__.Notify(240420)

	return k.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(240427)
		return k.updateSpanConfigRecordsWithTxn(ctx, toDelete, toUpsert, txn, minCommitTS, maxCommitTS)
	})
}

func newKVAccessor(
	db *kv.DB,
	ie sqlutil.InternalExecutor,
	settings *cluster.Settings,
	clock *hlc.Clock,
	configurationsTableFQN string,
	knobs *spanconfig.TestingKnobs,
	optionalTxn *kv.Txn,
) *KVAccessor {
	__antithesis_instrumentation__.Notify(240428)
	if knobs == nil {
		__antithesis_instrumentation__.Notify(240430)
		knobs = &spanconfig.TestingKnobs{}
	} else {
		__antithesis_instrumentation__.Notify(240431)
	}
	__antithesis_instrumentation__.Notify(240429)

	return &KVAccessor{
		db:                     db,
		ie:                     ie,
		clock:                  clock,
		optionalTxn:            optionalTxn,
		settings:               settings,
		configurationsTableFQN: configurationsTableFQN,
		knobs:                  knobs,
	}
}

func (k *KVAccessor) getSpanConfigRecordsWithTxn(
	ctx context.Context, targets []spanconfig.Target, txn *kv.Txn,
) ([]spanconfig.Record, error) {
	__antithesis_instrumentation__.Notify(240432)
	if txn == nil {
		__antithesis_instrumentation__.Notify(240437)
		log.Fatalf(ctx, "expected non-nil txn")
	} else {
		__antithesis_instrumentation__.Notify(240438)
	}
	__antithesis_instrumentation__.Notify(240433)

	if len(targets) == 0 {
		__antithesis_instrumentation__.Notify(240439)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(240440)
	}
	__antithesis_instrumentation__.Notify(240434)
	if err := validateSpanTargets(targets); err != nil {
		__antithesis_instrumentation__.Notify(240441)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(240442)
	}
	__antithesis_instrumentation__.Notify(240435)

	var records []spanconfig.Record
	if err := k.paginate(len(targets), func(startIdx, endIdx int) (retErr error) {
		__antithesis_instrumentation__.Notify(240443)
		targetsBatch := targets[startIdx:endIdx]
		getStmt, getQueryArgs := k.constructGetStmtAndArgs(targetsBatch)
		it, err := k.ie.QueryIteratorEx(ctx, "get-span-cfgs", txn,
			sessiondata.InternalExecutorOverride{User: security.RootUserName()},
			getStmt, getQueryArgs...,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(240448)
			return err
		} else {
			__antithesis_instrumentation__.Notify(240449)
		}
		__antithesis_instrumentation__.Notify(240444)
		defer func() {
			__antithesis_instrumentation__.Notify(240450)
			if closeErr := it.Close(); closeErr != nil {
				__antithesis_instrumentation__.Notify(240451)
				records, retErr = nil, errors.CombineErrors(retErr, closeErr)
			} else {
				__antithesis_instrumentation__.Notify(240452)
			}
		}()
		__antithesis_instrumentation__.Notify(240445)

		var ok bool
		for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
			__antithesis_instrumentation__.Notify(240453)
			row := it.Cur()
			span := roachpb.Span{
				Key:    []byte(*row[0].(*tree.DBytes)),
				EndKey: []byte(*row[1].(*tree.DBytes)),
			}
			var conf roachpb.SpanConfig
			if err := protoutil.Unmarshal(([]byte)(*row[2].(*tree.DBytes)), &conf); err != nil {
				__antithesis_instrumentation__.Notify(240456)
				return err
			} else {
				__antithesis_instrumentation__.Notify(240457)
			}
			__antithesis_instrumentation__.Notify(240454)

			record, err := spanconfig.MakeRecord(spanconfig.DecodeTarget(span), conf)
			if err != nil {
				__antithesis_instrumentation__.Notify(240458)
				return err
			} else {
				__antithesis_instrumentation__.Notify(240459)
			}
			__antithesis_instrumentation__.Notify(240455)
			records = append(records, record)
		}
		__antithesis_instrumentation__.Notify(240446)
		if err != nil {
			__antithesis_instrumentation__.Notify(240460)
			return err
		} else {
			__antithesis_instrumentation__.Notify(240461)
		}
		__antithesis_instrumentation__.Notify(240447)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(240462)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(240463)
	}
	__antithesis_instrumentation__.Notify(240436)

	return records, nil
}

func (k *KVAccessor) updateSpanConfigRecordsWithTxn(
	ctx context.Context,
	toDelete []spanconfig.Target,
	toUpsert []spanconfig.Record,
	txn *kv.Txn,
	minCommitTS, maxCommitTS hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(240464)
	if txn == nil {
		__antithesis_instrumentation__.Notify(240474)
		log.Fatalf(ctx, "expected non-nil txn")
	} else {
		__antithesis_instrumentation__.Notify(240475)
	}
	__antithesis_instrumentation__.Notify(240465)

	if !minCommitTS.Less(maxCommitTS) {
		__antithesis_instrumentation__.Notify(240476)
		return errors.AssertionFailedf("invalid commit interval [%s, %s)", minCommitTS, maxCommitTS)
	} else {
		__antithesis_instrumentation__.Notify(240477)
	}
	__antithesis_instrumentation__.Notify(240466)

	if txn.ReadTimestamp().Less(minCommitTS) {
		__antithesis_instrumentation__.Notify(240478)
		return errors.AssertionFailedf(
			"transaction's read timestamp (%s) below the minimum commit timestamp (%s)",
			txn.ReadTimestamp(), minCommitTS,
		)
	} else {
		__antithesis_instrumentation__.Notify(240479)
	}
	__antithesis_instrumentation__.Notify(240467)

	if maxCommitTS.Less(txn.ReadTimestamp()) {
		__antithesis_instrumentation__.Notify(240480)
		return spanconfig.NewCommitTimestampOutOfBoundsError()
	} else {
		__antithesis_instrumentation__.Notify(240481)
	}
	__antithesis_instrumentation__.Notify(240468)

	if err := txn.UpdateDeadline(ctx, maxCommitTS); err != nil {
		__antithesis_instrumentation__.Notify(240482)
		return errors.Wrapf(err, "transaction deadline could not be updated")
	} else {
		__antithesis_instrumentation__.Notify(240483)
	}
	__antithesis_instrumentation__.Notify(240469)

	if fn := k.knobs.KVAccessorPostCommitDeadlineSetInterceptor; fn != nil {
		__antithesis_instrumentation__.Notify(240484)
		fn(txn)
	} else {
		__antithesis_instrumentation__.Notify(240485)
	}
	__antithesis_instrumentation__.Notify(240470)

	if err := validateUpdateArgs(toDelete, toUpsert); err != nil {
		__antithesis_instrumentation__.Notify(240486)
		return err
	} else {
		__antithesis_instrumentation__.Notify(240487)
	}
	__antithesis_instrumentation__.Notify(240471)

	if len(toDelete) > 0 {
		__antithesis_instrumentation__.Notify(240488)
		if err := k.paginate(len(toDelete), func(startIdx, endIdx int) error {
			__antithesis_instrumentation__.Notify(240489)
			toDeleteBatch := toDelete[startIdx:endIdx]
			deleteStmt, deleteQueryArgs := k.constructDeleteStmtAndArgs(toDeleteBatch)
			n, err := k.ie.ExecEx(ctx, "delete-span-cfgs", txn,
				sessiondata.InternalExecutorOverride{User: security.RootUserName()},
				deleteStmt, deleteQueryArgs...,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(240492)
				return err
			} else {
				__antithesis_instrumentation__.Notify(240493)
			}
			__antithesis_instrumentation__.Notify(240490)
			if n != len(toDeleteBatch) {
				__antithesis_instrumentation__.Notify(240494)
				return errors.AssertionFailedf("expected to delete %d row(s), deleted %d", len(toDeleteBatch), n)
			} else {
				__antithesis_instrumentation__.Notify(240495)
			}
			__antithesis_instrumentation__.Notify(240491)
			return nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(240496)
			return err
		} else {
			__antithesis_instrumentation__.Notify(240497)
		}
	} else {
		__antithesis_instrumentation__.Notify(240498)
	}
	__antithesis_instrumentation__.Notify(240472)

	if len(toUpsert) == 0 {
		__antithesis_instrumentation__.Notify(240499)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(240500)
	}
	__antithesis_instrumentation__.Notify(240473)

	return k.paginate(len(toUpsert), func(startIdx, endIdx int) error {
		__antithesis_instrumentation__.Notify(240501)
		toUpsertBatch := toUpsert[startIdx:endIdx]
		upsertStmt, upsertQueryArgs, err := k.constructUpsertStmtAndArgs(toUpsertBatch)
		if err != nil {
			__antithesis_instrumentation__.Notify(240505)
			return err
		} else {
			__antithesis_instrumentation__.Notify(240506)
		}
		__antithesis_instrumentation__.Notify(240502)
		if n, err := k.ie.ExecEx(ctx, "upsert-span-cfgs", txn,
			sessiondata.InternalExecutorOverride{User: security.RootUserName()},
			upsertStmt, upsertQueryArgs...,
		); err != nil {
			__antithesis_instrumentation__.Notify(240507)
			return err
		} else {
			__antithesis_instrumentation__.Notify(240508)
			if n != len(toUpsertBatch) {
				__antithesis_instrumentation__.Notify(240509)
				return errors.AssertionFailedf("expected to upsert %d row(s), upserted %d", len(toUpsertBatch), n)
			} else {
				__antithesis_instrumentation__.Notify(240510)
			}
		}
		__antithesis_instrumentation__.Notify(240503)

		validationStmt, validationQueryArgs := k.constructValidationStmtAndArgs(toUpsertBatch)
		if datums, err := k.ie.QueryRowEx(ctx, "validate-span-cfgs", txn,
			sessiondata.InternalExecutorOverride{User: security.RootUserName()},
			validationStmt, validationQueryArgs...,
		); err != nil {
			__antithesis_instrumentation__.Notify(240511)
			return err
		} else {
			__antithesis_instrumentation__.Notify(240512)
			if valid := bool(tree.MustBeDBool(datums[0])); !valid {
				__antithesis_instrumentation__.Notify(240513)
				return errors.AssertionFailedf("expected to find single row containing upserted spans")
			} else {
				__antithesis_instrumentation__.Notify(240514)
			}
		}
		__antithesis_instrumentation__.Notify(240504)
		return nil
	})
}

func (k *KVAccessor) constructGetStmtAndArgs(targets []spanconfig.Target) (string, []interface{}) {
	__antithesis_instrumentation__.Notify(240515)

	var getStmtBuilder strings.Builder
	queryArgs := make([]interface{}, len(targets)*2)
	for i, target := range targets {
		__antithesis_instrumentation__.Notify(240517)
		if i > 0 {
			__antithesis_instrumentation__.Notify(240519)
			getStmtBuilder.WriteString(`UNION ALL`)
		} else {
			__antithesis_instrumentation__.Notify(240520)
		}
		__antithesis_instrumentation__.Notify(240518)

		startKeyIdx, endKeyIdx := i*2, (i*2)+1
		encodedSp := target.Encode()
		queryArgs[startKeyIdx] = encodedSp.Key
		queryArgs[endKeyIdx] = encodedSp.EndKey

		fmt.Fprintf(&getStmtBuilder, `
SELECT start_key, end_key, config FROM %[1]s
 WHERE start_key >= $%[2]d AND start_key < $%[3]d
UNION ALL
SELECT start_key, end_key, config FROM (
  SELECT start_key, end_key, config FROM %[1]s
  WHERE start_key < $%[2]d ORDER BY start_key DESC LIMIT 1
) WHERE end_key > $%[2]d
`,
			k.configurationsTableFQN,
			startKeyIdx+1,
			endKeyIdx+1,
		)
	}
	__antithesis_instrumentation__.Notify(240516)
	return getStmtBuilder.String(), queryArgs
}

func (k *KVAccessor) constructDeleteStmtAndArgs(
	toDelete []spanconfig.Target,
) (string, []interface{}) {
	__antithesis_instrumentation__.Notify(240521)

	values := make([]string, len(toDelete))
	deleteQueryArgs := make([]interface{}, len(toDelete)*2)
	for i, toDel := range toDelete {
		__antithesis_instrumentation__.Notify(240523)
		startKeyIdx, endKeyIdx := i*2, (i*2)+1
		encodedSp := toDel.Encode()
		deleteQueryArgs[startKeyIdx] = encodedSp.Key
		deleteQueryArgs[endKeyIdx] = encodedSp.EndKey
		values[i] = fmt.Sprintf("($%d::BYTES, $%d::BYTES)",
			startKeyIdx+1, endKeyIdx+1)
	}
	__antithesis_instrumentation__.Notify(240522)
	deleteStmt := fmt.Sprintf(`DELETE FROM %[1]s WHERE (start_key, end_key) IN (VALUES %[2]s)`,
		k.configurationsTableFQN, strings.Join(values, ", "))
	return deleteStmt, deleteQueryArgs
}

func (k *KVAccessor) constructUpsertStmtAndArgs(
	toUpsert []spanconfig.Record,
) (string, []interface{}, error) {
	__antithesis_instrumentation__.Notify(240524)

	upsertValues := make([]string, len(toUpsert))
	upsertQueryArgs := make([]interface{}, len(toUpsert)*3)
	for i, record := range toUpsert {
		__antithesis_instrumentation__.Notify(240526)
		cfg := record.GetConfig()
		marshaled, err := protoutil.Marshal(&cfg)
		if err != nil {
			__antithesis_instrumentation__.Notify(240528)
			return "", nil, err
		} else {
			__antithesis_instrumentation__.Notify(240529)
		}
		__antithesis_instrumentation__.Notify(240527)

		startKeyIdx, endKeyIdx, configIdx := i*3, (i*3)+1, (i*3)+2
		upsertQueryArgs[startKeyIdx] = record.GetTarget().Encode().Key
		upsertQueryArgs[endKeyIdx] = record.GetTarget().Encode().EndKey
		upsertQueryArgs[configIdx] = marshaled
		upsertValues[i] = fmt.Sprintf("($%d::BYTES, $%d::BYTES, $%d::BYTES)",
			startKeyIdx+1, endKeyIdx+1, configIdx+1)
	}
	__antithesis_instrumentation__.Notify(240525)
	upsertStmt := fmt.Sprintf(`UPSERT INTO %[1]s (start_key, end_key, config) VALUES %[2]s`,
		k.configurationsTableFQN, strings.Join(upsertValues, ", "))
	return upsertStmt, upsertQueryArgs, nil
}

func (k *KVAccessor) constructValidationStmtAndArgs(
	toUpsert []spanconfig.Record,
) (string, []interface{}) {
	__antithesis_instrumentation__.Notify(240530)

	targetsToUpsert := spanconfig.TargetsFromRecords(toUpsert)
	sort.Sort(spanconfig.Targets(targetsToUpsert))

	validationQueryArgs := make([]interface{}, 2)
	validationQueryArgs[0] = targetsToUpsert[0].Encode().Key

	validationQueryArgs[1] = targetsToUpsert[len(targetsToUpsert)-1].Encode().EndKey

	validationStmt := fmt.Sprintf(`
SELECT bool_and(valid) FROM (
  SELECT bool_and(prev_end_key IS NULL OR start_key >= prev_end_key) AS valid FROM (
    SELECT start_key, lag(end_key, 1) OVER (ORDER BY start_key) AS prev_end_key FROM %[1]s
    WHERE start_key >= $1 AND start_key < $2
  )
  UNION ALL
  SELECT * FROM (
    SELECT $1 >= end_key FROM %[1]s
    WHERE start_key < $1 ORDER BY start_key DESC LIMIT 1
  )
)`, k.configurationsTableFQN)

	return validationStmt, validationQueryArgs
}

func validateUpdateArgs(toDelete []spanconfig.Target, toUpsert []spanconfig.Record) error {
	__antithesis_instrumentation__.Notify(240531)
	targetsToUpdate := spanconfig.TargetsFromRecords(toUpsert)
	for _, list := range [][]spanconfig.Target{toDelete, targetsToUpdate} {
		__antithesis_instrumentation__.Notify(240533)
		if err := validateSpanTargets(list); err != nil {
			__antithesis_instrumentation__.Notify(240535)
			return err
		} else {
			__antithesis_instrumentation__.Notify(240536)
		}
		__antithesis_instrumentation__.Notify(240534)

		targets := make([]spanconfig.Target, len(list))
		copy(targets, list)
		sort.Sort(spanconfig.Targets(targets))
		for i := range targets {
			__antithesis_instrumentation__.Notify(240537)
			if targets[i].IsSystemTarget() && func() bool {
				__antithesis_instrumentation__.Notify(240541)
				return targets[i].GetSystemTarget().IsReadOnly() == true
			}() == true {
				__antithesis_instrumentation__.Notify(240542)
				return errors.AssertionFailedf(
					"cannot use read only system target %s as an update argument", targets[i],
				)
			} else {
				__antithesis_instrumentation__.Notify(240543)
			}
			__antithesis_instrumentation__.Notify(240538)

			if i == 0 {
				__antithesis_instrumentation__.Notify(240544)
				continue
			} else {
				__antithesis_instrumentation__.Notify(240545)
			}
			__antithesis_instrumentation__.Notify(240539)

			curTarget := targets[i]
			prevTarget := targets[i-1]
			if curTarget.IsSpanTarget() && func() bool {
				__antithesis_instrumentation__.Notify(240546)
				return prevTarget.IsSpanTarget() == true
			}() == true {
				__antithesis_instrumentation__.Notify(240547)
				if curTarget.GetSpan().Overlaps(prevTarget.GetSpan()) {
					__antithesis_instrumentation__.Notify(240548)
					return errors.AssertionFailedf("overlapping spans %s and %s in same list",
						prevTarget.GetSpan(), curTarget.GetSpan())
				} else {
					__antithesis_instrumentation__.Notify(240549)
				}
			} else {
				__antithesis_instrumentation__.Notify(240550)
			}
			__antithesis_instrumentation__.Notify(240540)

			if curTarget.IsSystemTarget() && func() bool {
				__antithesis_instrumentation__.Notify(240551)
				return prevTarget.IsSystemTarget() == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(240552)
				return curTarget.Equal(prevTarget) == true
			}() == true {
				__antithesis_instrumentation__.Notify(240553)
				return errors.AssertionFailedf("duplicate system targets %s in the same list",
					prevTarget.GetSystemTarget())
			} else {
				__antithesis_instrumentation__.Notify(240554)
			}

		}
	}
	__antithesis_instrumentation__.Notify(240532)

	return nil
}

func validateSpanTargets(targets []spanconfig.Target) error {
	__antithesis_instrumentation__.Notify(240555)
	for _, target := range targets {
		__antithesis_instrumentation__.Notify(240557)
		if !target.IsSpanTarget() {
			__antithesis_instrumentation__.Notify(240559)

			continue
		} else {
			__antithesis_instrumentation__.Notify(240560)
		}
		__antithesis_instrumentation__.Notify(240558)
		if err := validateSpans(target.GetSpan()); err != nil {
			__antithesis_instrumentation__.Notify(240561)
			return err
		} else {
			__antithesis_instrumentation__.Notify(240562)
		}
	}
	__antithesis_instrumentation__.Notify(240556)
	return nil
}

func validateSpans(spans ...roachpb.Span) error {
	__antithesis_instrumentation__.Notify(240563)
	for _, span := range spans {
		__antithesis_instrumentation__.Notify(240565)
		if !span.Valid() || func() bool {
			__antithesis_instrumentation__.Notify(240566)
			return len(span.EndKey) == 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(240567)
			return errors.AssertionFailedf("invalid span: %s", span)
		} else {
			__antithesis_instrumentation__.Notify(240568)
		}
	}
	__antithesis_instrumentation__.Notify(240564)
	return nil
}

func (k *KVAccessor) paginate(totalLen int, f func(startIdx, endIdx int) error) error {
	__antithesis_instrumentation__.Notify(240569)
	batchSize := math.MaxInt32
	if b := batchSizeSetting.Get(&k.settings.SV); int(b) > 0 {
		__antithesis_instrumentation__.Notify(240573)
		batchSize = int(b)
	} else {
		__antithesis_instrumentation__.Notify(240574)
	}
	__antithesis_instrumentation__.Notify(240570)

	if fn := k.knobs.KVAccessorBatchSizeOverrideFn; fn != nil {
		__antithesis_instrumentation__.Notify(240575)
		batchSize = fn()
	} else {
		__antithesis_instrumentation__.Notify(240576)
	}
	__antithesis_instrumentation__.Notify(240571)

	for i := 0; i < totalLen; i += batchSize {
		__antithesis_instrumentation__.Notify(240577)
		j := i + batchSize
		if j > totalLen {
			__antithesis_instrumentation__.Notify(240580)
			j = totalLen
		} else {
			__antithesis_instrumentation__.Notify(240581)
		}
		__antithesis_instrumentation__.Notify(240578)

		if fn := k.knobs.KVAccessorPaginationInterceptor; fn != nil {
			__antithesis_instrumentation__.Notify(240582)
			fn()
		} else {
			__antithesis_instrumentation__.Notify(240583)
		}
		__antithesis_instrumentation__.Notify(240579)

		if err := f(i, j); err != nil {
			__antithesis_instrumentation__.Notify(240584)
			return err
		} else {
			__antithesis_instrumentation__.Notify(240585)
		}
	}
	__antithesis_instrumentation__.Notify(240572)
	return nil
}
