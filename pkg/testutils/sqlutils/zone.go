package sqlutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	gosql "database/sql"
	"fmt"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
)

type ZoneRow struct {
	ID     uint32
	Config zonepb.ZoneConfig
}

func (row ZoneRow) sqlRowString() ([]string, error) {
	__antithesis_instrumentation__.Notify(646375)
	configProto, err := protoutil.Marshal(&row.Config)
	if err != nil {
		__antithesis_instrumentation__.Notify(646377)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(646378)
	}
	__antithesis_instrumentation__.Notify(646376)
	return []string{
		fmt.Sprintf("%d", row.ID),
		string(configProto),
	}, nil
}

func RemoveAllZoneConfigs(t testing.TB, sqlDB *SQLRunner) {
	__antithesis_instrumentation__.Notify(646379)
	t.Helper()
	for _, row := range sqlDB.QueryStr(t, "SHOW ALL ZONE CONFIGURATIONS") {
		__antithesis_instrumentation__.Notify(646380)
		target := row[0]
		if target == fmt.Sprintf("RANGE %s", zonepb.DefaultZoneName) {
			__antithesis_instrumentation__.Notify(646382)

			continue
		} else {
			__antithesis_instrumentation__.Notify(646383)
		}
		__antithesis_instrumentation__.Notify(646381)
		DeleteZoneConfig(t, sqlDB, target)
	}
}

func DeleteZoneConfig(t testing.TB, sqlDB *SQLRunner, target string) {
	__antithesis_instrumentation__.Notify(646384)
	t.Helper()
	sqlDB.Exec(t, fmt.Sprintf("ALTER %s CONFIGURE ZONE DISCARD", target))
}

func SetZoneConfig(t testing.TB, sqlDB *SQLRunner, target string, config string) {
	__antithesis_instrumentation__.Notify(646385)
	t.Helper()
	sqlDB.Exec(t, fmt.Sprintf("ALTER %s CONFIGURE ZONE = %s",
		target, lexbase.EscapeSQLString(config)))
}

func TxnSetZoneConfig(t testing.TB, sqlDB *SQLRunner, txn *gosql.Tx, target string, config string) {
	__antithesis_instrumentation__.Notify(646386)
	t.Helper()
	_, err := txn.Exec(fmt.Sprintf("ALTER %s CONFIGURE ZONE = %s",
		target, lexbase.EscapeSQLString(config)))
	if err != nil {
		__antithesis_instrumentation__.Notify(646387)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(646388)
	}
}

func VerifyZoneConfigForTarget(t testing.TB, sqlDB *SQLRunner, target string, row ZoneRow) {
	__antithesis_instrumentation__.Notify(646389)
	t.Helper()
	sqlRow, err := row.sqlRowString()
	if err != nil {
		__antithesis_instrumentation__.Notify(646391)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(646392)
	}
	__antithesis_instrumentation__.Notify(646390)
	sqlDB.CheckQueryResults(t, fmt.Sprintf(`
SELECT zone_id, raw_config_protobuf
  FROM [SHOW ZONE CONFIGURATION FOR %s]`, target),
		[][]string{sqlRow})
}

func VerifyAllZoneConfigs(t testing.TB, sqlDB *SQLRunner, rows ...ZoneRow) {
	__antithesis_instrumentation__.Notify(646393)
	t.Helper()
	expected := make([][]string, len(rows))
	for i, row := range rows {
		__antithesis_instrumentation__.Notify(646395)
		var err error
		expected[i], err = row.sqlRowString()
		if err != nil {
			__antithesis_instrumentation__.Notify(646396)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(646397)
		}
	}
	__antithesis_instrumentation__.Notify(646394)
	sqlDB.CheckQueryResults(t, `SELECT zone_id, raw_config_protobuf FROM crdb_internal.zones`, expected)
}

func ZoneConfigExists(t testing.TB, sqlDB *SQLRunner, name string) bool {
	__antithesis_instrumentation__.Notify(646398)
	t.Helper()
	var exists bool
	sqlDB.QueryRow(
		t, "SELECT EXISTS (SELECT 1 FROM crdb_internal.zones WHERE target = $1)", name,
	).Scan(&exists)
	return exists
}
