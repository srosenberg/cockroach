package kvnemesis

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach-go/v2/crdb"
	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
)

type Env struct {
	sqlDBs []*gosql.DB
}

func (e *Env) anyNode() *gosql.DB {
	__antithesis_instrumentation__.Notify(90246)

	return e.sqlDBs[0]
}

func (e *Env) SetClosedTimestampInterval(ctx context.Context, d time.Duration) error {
	__antithesis_instrumentation__.Notify(90247)
	q := fmt.Sprintf(`SET CLUSTER SETTING kv.closed_timestamp.target_duration = '%s'`, d)
	_, err := e.anyNode().ExecContext(ctx, q)
	return err
}

func (e *Env) ResetClosedTimestampInterval(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(90248)
	const q = `SET CLUSTER SETTING kv.closed_timestamp.target_duration TO DEFAULT`
	_, err := e.anyNode().ExecContext(ctx, q)
	return err
}

func (e *Env) UpdateZoneConfig(
	ctx context.Context, zoneID int, updateZone func(*zonepb.ZoneConfig),
) error {
	__antithesis_instrumentation__.Notify(90249)
	return crdb.ExecuteTx(ctx, e.anyNode(), nil, func(tx *gosql.Tx) error {
		__antithesis_instrumentation__.Notify(90250)

		var zone zonepb.ZoneConfig
		var zoneRaw []byte
		const q1 = `SELECT config FROM system.zones WHERE id = $1`
		if err := tx.QueryRowContext(ctx, q1, zoneID).Scan(&zoneRaw); err != nil {
			__antithesis_instrumentation__.Notify(90253)
			if !errors.Is(err, gosql.ErrNoRows) {
				__antithesis_instrumentation__.Notify(90255)
				return err
			} else {
				__antithesis_instrumentation__.Notify(90256)
			}
			__antithesis_instrumentation__.Notify(90254)

			zone = zonepb.DefaultZoneConfig()
		} else {
			__antithesis_instrumentation__.Notify(90257)
			if err := protoutil.Unmarshal(zoneRaw, &zone); err != nil {
				__antithesis_instrumentation__.Notify(90258)
				return errors.Wrap(err, "unmarshaling existing zone")
			} else {
				__antithesis_instrumentation__.Notify(90259)
			}
		}
		__antithesis_instrumentation__.Notify(90251)

		updateZone(&zone)
		zoneRaw, err := protoutil.Marshal(&zone)
		if err != nil {
			__antithesis_instrumentation__.Notify(90260)
			return errors.Wrap(err, "marshaling new zone")
		} else {
			__antithesis_instrumentation__.Notify(90261)
		}
		__antithesis_instrumentation__.Notify(90252)

		const q2 = `UPSERT INTO system.zones (id, config) VALUES ($1, $2)`
		_, err = tx.ExecContext(ctx, q2, zoneID, zoneRaw)
		return err
	})
}
