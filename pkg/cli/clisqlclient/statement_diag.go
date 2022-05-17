package clisqlclient

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"database/sql/driver"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/cockroachdb/errors"
)

type StmtDiagBundleInfo struct {
	ID int64

	Statement   string
	CollectedAt time.Time
}

func StmtDiagListBundles(ctx context.Context, conn Conn) ([]StmtDiagBundleInfo, error) {
	__antithesis_instrumentation__.Notify(28853)
	result, err := stmtDiagListBundlesInternal(ctx, conn)
	if err != nil {
		__antithesis_instrumentation__.Notify(28855)
		return nil, errors.Wrap(
			err, "failed to retrieve statement diagnostics bundles",
		)
	} else {
		__antithesis_instrumentation__.Notify(28856)
	}
	__antithesis_instrumentation__.Notify(28854)
	return result, nil
}

func stmtDiagListBundlesInternal(ctx context.Context, conn Conn) ([]StmtDiagBundleInfo, error) {
	__antithesis_instrumentation__.Notify(28857)
	rows, err := conn.Query(ctx,
		`SELECT id, statement_fingerprint, collected_at
		 FROM system.statement_diagnostics
		 WHERE error IS NULL
		 ORDER BY collected_at DESC`,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(28861)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(28862)
	}
	__antithesis_instrumentation__.Notify(28858)
	var result []StmtDiagBundleInfo
	vals := make([]driver.Value, 3)
	for {
		__antithesis_instrumentation__.Notify(28863)
		if err := rows.Next(vals); err == io.EOF {
			__antithesis_instrumentation__.Notify(28865)
			break
		} else {
			__antithesis_instrumentation__.Notify(28866)
			if err != nil {
				__antithesis_instrumentation__.Notify(28867)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(28868)
			}
		}
		__antithesis_instrumentation__.Notify(28864)
		info := StmtDiagBundleInfo{
			ID:          vals[0].(int64),
			Statement:   vals[1].(string),
			CollectedAt: vals[2].(time.Time),
		}
		result = append(result, info)
	}
	__antithesis_instrumentation__.Notify(28859)
	if err := rows.Close(); err != nil {
		__antithesis_instrumentation__.Notify(28869)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(28870)
	}
	__antithesis_instrumentation__.Notify(28860)
	return result, nil
}

type StmtDiagActivationRequest struct {
	ID int64

	Statement   string
	RequestedAt time.Time

	MinExecutionLatency time.Duration

	ExpiresAt time.Time
}

func StmtDiagListOutstandingRequests(
	ctx context.Context, conn Conn,
) ([]StmtDiagActivationRequest, error) {
	__antithesis_instrumentation__.Notify(28871)
	result, err := stmtDiagListOutstandingRequestsInternal(ctx, conn)
	if err != nil {
		__antithesis_instrumentation__.Notify(28873)
		return nil, errors.Wrap(
			err, "failed to retrieve outstanding statement diagnostics activation requests",
		)
	} else {
		__antithesis_instrumentation__.Notify(28874)
	}
	__antithesis_instrumentation__.Notify(28872)
	return result, nil
}

func isAtLeast22dot1ClusterVersion(ctx context.Context, conn Conn) (bool, error) {
	__antithesis_instrumentation__.Notify(28875)

	row, err := conn.QueryRow(ctx, `
SELECT
  count(*)
FROM
  [SHOW COLUMNS FROM system.statement_diagnostics_requests]
WHERE
  column_name = 'min_execution_latency';`)
	if err != nil {
		__antithesis_instrumentation__.Notify(28878)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(28879)
	}
	__antithesis_instrumentation__.Notify(28876)
	c, ok := row[0].(int64)
	if !ok {
		__antithesis_instrumentation__.Notify(28880)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(28881)
	}
	__antithesis_instrumentation__.Notify(28877)
	return c == 1, nil
}

func stmtDiagListOutstandingRequestsInternal(
	ctx context.Context, conn Conn,
) ([]StmtDiagActivationRequest, error) {
	__antithesis_instrumentation__.Notify(28882)
	var extraColumns string
	atLeast22dot1, err := isAtLeast22dot1ClusterVersion(ctx, conn)
	if err != nil {
		__antithesis_instrumentation__.Notify(28888)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(28889)
	}
	__antithesis_instrumentation__.Notify(28883)
	if atLeast22dot1 {
		__antithesis_instrumentation__.Notify(28890)

		getMilliseconds := `EXTRACT(epoch FROM min_execution_latency)::INT8 * 1000 +
                        EXTRACT(millisecond FROM min_execution_latency)::INT8 -
                        EXTRACT(second FROM min_execution_latency)::INT8 * 1000`
		extraColumns = ", " + getMilliseconds + ", expires_at"
	} else {
		__antithesis_instrumentation__.Notify(28891)
	}
	__antithesis_instrumentation__.Notify(28884)
	rows, err := conn.Query(ctx,
		fmt.Sprintf(`SELECT id, statement_fingerprint, requested_at%s
		 FROM system.statement_diagnostics_requests
		 WHERE NOT completed
		 ORDER BY requested_at DESC`, extraColumns),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(28892)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(28893)
	}
	__antithesis_instrumentation__.Notify(28885)
	var result []StmtDiagActivationRequest
	vals := make([]driver.Value, 5)
	for {
		__antithesis_instrumentation__.Notify(28894)
		if err := rows.Next(vals); err == io.EOF {
			__antithesis_instrumentation__.Notify(28897)
			break
		} else {
			__antithesis_instrumentation__.Notify(28898)
			if err != nil {
				__antithesis_instrumentation__.Notify(28899)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(28900)
			}
		}
		__antithesis_instrumentation__.Notify(28895)
		var minExecutionLatency time.Duration
		var expiresAt time.Time
		if atLeast22dot1 {
			__antithesis_instrumentation__.Notify(28901)
			if ms, ok := vals[3].(int64); ok {
				__antithesis_instrumentation__.Notify(28903)
				minExecutionLatency = time.Millisecond * time.Duration(ms)
			} else {
				__antithesis_instrumentation__.Notify(28904)
			}
			__antithesis_instrumentation__.Notify(28902)
			if e, ok := vals[4].(time.Time); ok {
				__antithesis_instrumentation__.Notify(28905)
				expiresAt = e
			} else {
				__antithesis_instrumentation__.Notify(28906)
			}
		} else {
			__antithesis_instrumentation__.Notify(28907)
		}
		__antithesis_instrumentation__.Notify(28896)
		info := StmtDiagActivationRequest{
			ID:                  vals[0].(int64),
			Statement:           vals[1].(string),
			RequestedAt:         vals[2].(time.Time),
			MinExecutionLatency: minExecutionLatency,
			ExpiresAt:           expiresAt,
		}
		result = append(result, info)
	}
	__antithesis_instrumentation__.Notify(28886)
	if err := rows.Close(); err != nil {
		__antithesis_instrumentation__.Notify(28908)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(28909)
	}
	__antithesis_instrumentation__.Notify(28887)
	return result, nil
}

func StmtDiagDownloadBundle(ctx context.Context, conn Conn, id int64, filename string) error {
	__antithesis_instrumentation__.Notify(28910)
	if err := stmtDiagDownloadBundleInternal(ctx, conn, id, filename); err != nil {
		__antithesis_instrumentation__.Notify(28912)
		return errors.Wrapf(
			err, "failed to download statement diagnostics bundle %d to '%s'", id, filename,
		)
	} else {
		__antithesis_instrumentation__.Notify(28913)
	}
	__antithesis_instrumentation__.Notify(28911)
	return nil
}

func stmtDiagDownloadBundleInternal(
	ctx context.Context, conn Conn, id int64, filename string,
) error {
	__antithesis_instrumentation__.Notify(28914)

	rows, err := conn.Query(ctx,
		"SELECT unnest(bundle_chunks) FROM system.statement_diagnostics WHERE id = $1",
		id,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(28921)
		return err
	} else {
		__antithesis_instrumentation__.Notify(28922)
	}
	__antithesis_instrumentation__.Notify(28915)
	var chunkIDs []int64
	vals := make([]driver.Value, 1)
	for {
		__antithesis_instrumentation__.Notify(28923)
		if err := rows.Next(vals); err == io.EOF {
			__antithesis_instrumentation__.Notify(28925)
			break
		} else {
			__antithesis_instrumentation__.Notify(28926)
			if err != nil {
				__antithesis_instrumentation__.Notify(28927)
				return err
			} else {
				__antithesis_instrumentation__.Notify(28928)
			}
		}
		__antithesis_instrumentation__.Notify(28924)
		chunkIDs = append(chunkIDs, vals[0].(int64))
	}
	__antithesis_instrumentation__.Notify(28916)
	if err := rows.Close(); err != nil {
		__antithesis_instrumentation__.Notify(28929)
		return err
	} else {
		__antithesis_instrumentation__.Notify(28930)
	}
	__antithesis_instrumentation__.Notify(28917)

	if len(chunkIDs) == 0 {
		__antithesis_instrumentation__.Notify(28931)
		return errors.Newf("no statement diagnostics bundle with ID %d", id)
	} else {
		__antithesis_instrumentation__.Notify(28932)
	}
	__antithesis_instrumentation__.Notify(28918)

	out, err := os.Create(filename)
	if err != nil {
		__antithesis_instrumentation__.Notify(28933)
		return err
	} else {
		__antithesis_instrumentation__.Notify(28934)
	}
	__antithesis_instrumentation__.Notify(28919)

	for _, chunkID := range chunkIDs {
		__antithesis_instrumentation__.Notify(28935)
		data, err := conn.QueryRow(ctx,
			"SELECT data FROM system.statement_bundle_chunks WHERE id = $1",
			chunkID,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(28937)
			_ = out.Close()
			return err
		} else {
			__antithesis_instrumentation__.Notify(28938)
		}
		__antithesis_instrumentation__.Notify(28936)
		if _, err := out.Write(data[0].([]byte)); err != nil {
			__antithesis_instrumentation__.Notify(28939)
			_ = out.Close()
			return err
		} else {
			__antithesis_instrumentation__.Notify(28940)
		}
	}
	__antithesis_instrumentation__.Notify(28920)

	return out.Close()
}

func StmtDiagDeleteBundle(ctx context.Context, conn Conn, id int64) error {
	__antithesis_instrumentation__.Notify(28941)
	_, err := conn.QueryRow(ctx,
		"SELECT 1 FROM system.statement_diagnostics WHERE id = $1",
		id,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(28943)
		if err == io.EOF {
			__antithesis_instrumentation__.Notify(28945)
			return errors.Newf("no statement diagnostics bundle with ID %d", id)
		} else {
			__antithesis_instrumentation__.Notify(28946)
		}
		__antithesis_instrumentation__.Notify(28944)
		return err
	} else {
		__antithesis_instrumentation__.Notify(28947)
	}
	__antithesis_instrumentation__.Notify(28942)
	return conn.ExecTxn(ctx, func(ctx context.Context, conn TxBoundConn) error {
		__antithesis_instrumentation__.Notify(28948)

		if err := conn.Exec(ctx,
			"DELETE FROM system.statement_diagnostics_requests WHERE statement_diagnostics_id = $1",
			id,
		); err != nil {
			__antithesis_instrumentation__.Notify(28951)
			return err
		} else {
			__antithesis_instrumentation__.Notify(28952)
		}
		__antithesis_instrumentation__.Notify(28949)

		if err := conn.Exec(ctx,
			`DELETE FROM system.statement_bundle_chunks
			  WHERE id IN (
				  SELECT unnest(bundle_chunks) FROM system.statement_diagnostics WHERE id = $1
				)`,
			id,
		); err != nil {
			__antithesis_instrumentation__.Notify(28953)
			return err
		} else {
			__antithesis_instrumentation__.Notify(28954)
		}
		__antithesis_instrumentation__.Notify(28950)

		return conn.Exec(ctx,
			"DELETE FROM system.statement_diagnostics WHERE id = $1",
			id,
		)
	})
}

func StmtDiagDeleteAllBundles(ctx context.Context, conn Conn) error {
	__antithesis_instrumentation__.Notify(28955)
	return conn.ExecTxn(ctx, func(ctx context.Context, conn TxBoundConn) error {
		__antithesis_instrumentation__.Notify(28956)

		if err := conn.Exec(ctx,
			"DELETE FROM system.statement_diagnostics_requests WHERE completed",
		); err != nil {
			__antithesis_instrumentation__.Notify(28959)
			return err
		} else {
			__antithesis_instrumentation__.Notify(28960)
		}
		__antithesis_instrumentation__.Notify(28957)

		if err := conn.Exec(ctx,
			`DELETE FROM system.statement_bundle_chunks WHERE true`,
		); err != nil {
			__antithesis_instrumentation__.Notify(28961)
			return err
		} else {
			__antithesis_instrumentation__.Notify(28962)
		}
		__antithesis_instrumentation__.Notify(28958)

		return conn.Exec(ctx,
			"DELETE FROM system.statement_diagnostics WHERE true",
		)
	})
}

func StmtDiagCancelOutstandingRequest(ctx context.Context, conn Conn, id int64) error {
	__antithesis_instrumentation__.Notify(28963)
	_, err := conn.QueryRow(ctx,
		"DELETE FROM system.statement_diagnostics_requests WHERE id = $1 RETURNING id",
		id,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(28965)
		if err == io.EOF {
			__antithesis_instrumentation__.Notify(28967)
			return errors.Newf("no outstanding activation request with ID %d", id)
		} else {
			__antithesis_instrumentation__.Notify(28968)
		}
		__antithesis_instrumentation__.Notify(28966)
		return err
	} else {
		__antithesis_instrumentation__.Notify(28969)
	}
	__antithesis_instrumentation__.Notify(28964)
	return nil
}

func StmtDiagCancelAllOutstandingRequests(ctx context.Context, conn Conn) error {
	__antithesis_instrumentation__.Notify(28970)
	return conn.Exec(ctx,
		"DELETE FROM system.statement_diagnostics_requests WHERE NOT completed",
	)
}
