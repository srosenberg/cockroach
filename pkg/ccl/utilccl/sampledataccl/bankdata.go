package sampledataccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/testutils/serverutils"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"github.com/cockroachdb/cockroach/pkg/workload/workloadsql"
)

func ToBackup(t testing.TB, data workload.Table, externalIODir, path string) (*Backup, error) {
	__antithesis_instrumentation__.Notify(27485)
	return toBackup(t, data, externalIODir, path, 0)
}

func toBackup(
	t testing.TB, data workload.Table, externalIODir, path string, chunkBytes int64,
) (*Backup, error) {
	__antithesis_instrumentation__.Notify(27486)

	ctx := context.Background()
	s, db, _ := serverutils.StartServer(t, base.TestServerArgs{ExternalIODir: externalIODir})
	defer s.Stopper().Stop(ctx)
	if _, err := db.Exec(`CREATE DATABASE data`); err != nil {
		__antithesis_instrumentation__.Notify(27491)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(27492)
	}
	__antithesis_instrumentation__.Notify(27487)

	if _, err := db.Exec(fmt.Sprintf("CREATE TABLE %s %s;\n", data.Name, data.Schema)); err != nil {
		__antithesis_instrumentation__.Notify(27493)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(27494)
	}
	__antithesis_instrumentation__.Notify(27488)

	for rowIdx := 0; rowIdx < data.InitialRows.NumBatches; rowIdx++ {
		__antithesis_instrumentation__.Notify(27495)
		var stmts bytes.Buffer
		for _, row := range data.InitialRows.BatchRows(rowIdx) {
			__antithesis_instrumentation__.Notify(27497)
			rowBatch := strings.Join(workloadsql.StringTuple(row), `,`)
			fmt.Fprintf(&stmts, "INSERT INTO %s VALUES (%s);\n", data.Name, rowBatch)
		}
		__antithesis_instrumentation__.Notify(27496)
		if _, err := db.Exec(stmts.String()); err != nil {
			__antithesis_instrumentation__.Notify(27498)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(27499)
		}
	}
	__antithesis_instrumentation__.Notify(27489)

	if _, err := db.Exec("BACKUP DATABASE data TO $1", "nodelocal://0/"+path); err != nil {
		__antithesis_instrumentation__.Notify(27500)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(27501)
	}
	__antithesis_instrumentation__.Notify(27490)
	return &Backup{BaseDir: filepath.Join(externalIODir, path)}, nil
}

type Backup struct {
	BaseDir string
}
