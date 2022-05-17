// Package sqltestutils provides helper methods for testing sql packages.
package sqltestutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/desctestutils"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/testutils/sqlutils"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func AddImmediateGCZoneConfig(sqlDB *gosql.DB, id descpb.ID) (zonepb.ZoneConfig, error) {
	__antithesis_instrumentation__.Notify(625894)
	cfg := zonepb.DefaultZoneConfig()
	cfg.GC.TTLSeconds = 0
	buf, err := protoutil.Marshal(&cfg)
	if err != nil {
		__antithesis_instrumentation__.Notify(625896)
		return cfg, err
	} else {
		__antithesis_instrumentation__.Notify(625897)
	}
	__antithesis_instrumentation__.Notify(625895)
	_, err = sqlDB.Exec(`UPSERT INTO system.zones VALUES ($1, $2)`, id, buf)
	return cfg, err
}

func DisableGCTTLStrictEnforcement(t *testing.T, db *gosql.DB) (cleanup func()) {
	__antithesis_instrumentation__.Notify(625898)
	_, err := db.Exec(`SET CLUSTER SETTING kv.gc_ttl.strict_enforcement.enabled = false`)
	require.NoError(t, err)
	return func() {
		__antithesis_instrumentation__.Notify(625899)
		_, err := db.Exec(`SET CLUSTER SETTING kv.gc_ttl.strict_enforcement.enabled = DEFAULT`)
		require.NoError(t, err)
	}
}

func SetShortRangeFeedIntervals(t *testing.T, db sqlutils.DBHandle) {
	__antithesis_instrumentation__.Notify(625900)
	tdb := sqlutils.MakeSQLRunner(db)
	short := "'20ms'"
	if util.RaceEnabled {
		__antithesis_instrumentation__.Notify(625902)
		short = "'200ms'"
	} else {
		__antithesis_instrumentation__.Notify(625903)
	}
	__antithesis_instrumentation__.Notify(625901)
	tdb.Exec(t, `SET CLUSTER SETTING kv.closed_timestamp.target_duration = `+short)
	tdb.Exec(t, `SET CLUSTER SETTING kv.closed_timestamp.side_transport_interval = `+short)
}

func AddDefaultZoneConfig(sqlDB *gosql.DB, id descpb.ID) (zonepb.ZoneConfig, error) {
	__antithesis_instrumentation__.Notify(625904)
	cfg := zonepb.DefaultZoneConfig()
	buf, err := protoutil.Marshal(&cfg)
	if err != nil {
		__antithesis_instrumentation__.Notify(625906)
		return cfg, err
	} else {
		__antithesis_instrumentation__.Notify(625907)
	}
	__antithesis_instrumentation__.Notify(625905)
	_, err = sqlDB.Exec(`UPSERT INTO system.zones VALUES ($1, $2)`, id, buf)
	return cfg, err
}

func BulkInsertIntoTable(sqlDB *gosql.DB, maxValue int) error {
	__antithesis_instrumentation__.Notify(625908)
	inserts := make([]string, maxValue+1)
	for i := 0; i < maxValue+1; i++ {
		__antithesis_instrumentation__.Notify(625910)
		inserts[i] = fmt.Sprintf(`(%d, %d)`, i, maxValue-i)
	}
	__antithesis_instrumentation__.Notify(625909)
	_, err := sqlDB.Exec(`INSERT INTO t.test (k, v) VALUES ` + strings.Join(inserts, ","))
	return err
}

func GetTableKeyCount(ctx context.Context, kvDB *kv.DB) (int, error) {
	__antithesis_instrumentation__.Notify(625911)
	tableDesc := desctestutils.TestingGetPublicTableDescriptor(kvDB, keys.SystemSQLCodec, "t", "test")
	tablePrefix := keys.SystemSQLCodec.TablePrefix(uint32(tableDesc.GetID()))
	tableEnd := tablePrefix.PrefixEnd()
	kvs, err := kvDB.Scan(ctx, tablePrefix, tableEnd, 0)
	return len(kvs), err
}

func CheckTableKeyCountExact(ctx context.Context, kvDB *kv.DB, e int) error {
	__antithesis_instrumentation__.Notify(625912)
	if count, err := GetTableKeyCount(ctx, kvDB); err != nil {
		__antithesis_instrumentation__.Notify(625914)
		return err
	} else {
		__antithesis_instrumentation__.Notify(625915)
		if count != e {
			__antithesis_instrumentation__.Notify(625916)
			return errors.Newf("expected %d key value pairs, but got %d", e, count)
		} else {
			__antithesis_instrumentation__.Notify(625917)
		}
	}
	__antithesis_instrumentation__.Notify(625913)
	return nil
}

func CheckTableKeyCount(ctx context.Context, kvDB *kv.DB, multiple int, maxValue int) error {
	__antithesis_instrumentation__.Notify(625918)
	return CheckTableKeyCountExact(ctx, kvDB, multiple*(maxValue+1))
}

func IsClientSideQueryCanceledErr(err error) bool {
	__antithesis_instrumentation__.Notify(625919)
	if pqErr := (*pq.Error)(nil); errors.As(err, &pqErr) {
		__antithesis_instrumentation__.Notify(625921)
		return pgcode.MakeCode(string(pqErr.Code)) == pgcode.QueryCanceled
	} else {
		__antithesis_instrumentation__.Notify(625922)
	}
	__antithesis_instrumentation__.Notify(625920)
	return pgerror.GetPGCode(err) == pgcode.QueryCanceled
}
