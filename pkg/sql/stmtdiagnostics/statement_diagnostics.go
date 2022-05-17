package stmtdiagnostics

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/gossip"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

var pollingInterval = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"sql.stmt_diagnostics.poll_interval",
	"rate at which the stmtdiagnostics.Registry polls for requests, set to zero to disable",
	10*time.Second)

var bundleChunkSize = settings.RegisterByteSizeSetting(
	settings.TenantWritable,
	"sql.stmt_diagnostics.bundle_chunk_size",
	"chunk size for statement diagnostic bundles",
	1024*1024,
	func(val int64) error {
		__antithesis_instrumentation__.Notify(627375)
		if val < 16 {
			__antithesis_instrumentation__.Notify(627377)
			return errors.Errorf("chunk size must be at least 16 bytes")
		} else {
			__antithesis_instrumentation__.Notify(627378)
		}
		__antithesis_instrumentation__.Notify(627376)
		return nil
	},
)

type Registry struct {
	mu struct {
		syncutil.Mutex

		requestFingerprints map[RequestID]Request

		ongoing map[RequestID]Request

		epoch int
	}
	st     *cluster.Settings
	ie     sqlutil.InternalExecutor
	db     *kv.DB
	gossip gossip.OptionalGossip

	gossipUpdateChan chan RequestID

	gossipCancelChan chan RequestID
}

type Request struct {
	fingerprint         string
	minExecutionLatency time.Duration
	expiresAt           time.Time
}

func (r *Request) isExpired(now time.Time) bool {
	__antithesis_instrumentation__.Notify(627379)
	return !r.expiresAt.IsZero() && func() bool {
		__antithesis_instrumentation__.Notify(627380)
		return r.expiresAt.Before(now) == true
	}() == true
}

func (r *Request) isConditional() bool {
	__antithesis_instrumentation__.Notify(627381)
	return r.minExecutionLatency != 0
}

func NewRegistry(
	ie sqlutil.InternalExecutor, db *kv.DB, gw gossip.OptionalGossip, st *cluster.Settings,
) *Registry {
	__antithesis_instrumentation__.Notify(627382)
	r := &Registry{
		ie:               ie,
		db:               db,
		gossip:           gw,
		gossipUpdateChan: make(chan RequestID, 1),
		gossipCancelChan: make(chan RequestID, 1),
		st:               st,
	}

	g, ok := gw.Optional(47893)
	if ok && func() bool {
		__antithesis_instrumentation__.Notify(627384)
		return g != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(627385)
		g.RegisterCallback(gossip.KeyGossipStatementDiagnosticsRequest, r.gossipNotification)
	} else {
		__antithesis_instrumentation__.Notify(627386)
	}
	__antithesis_instrumentation__.Notify(627383)
	return r
}

func (r *Registry) Start(ctx context.Context, stopper *stop.Stopper) {
	__antithesis_instrumentation__.Notify(627387)
	ctx, _ = stopper.WithCancelOnQuiesce(ctx)

	_ = stopper.RunAsyncTask(ctx, "stmt-diag-poll", r.poll)
}

func (r *Registry) poll(ctx context.Context) {
	__antithesis_instrumentation__.Notify(627388)
	var (
		timer               timeutil.Timer
		lastPoll            time.Time
		deadline            time.Time
		pollIntervalChanged = make(chan struct{}, 1)
		maybeResetTimer     = func() {
			__antithesis_instrumentation__.Notify(627391)
			if interval := pollingInterval.Get(&r.st.SV); interval <= 0 {
				__antithesis_instrumentation__.Notify(627392)

				timer.Stop()
			} else {
				__antithesis_instrumentation__.Notify(627393)
				newDeadline := lastPoll.Add(interval)
				if deadline.IsZero() || func() bool {
					__antithesis_instrumentation__.Notify(627394)
					return !deadline.Equal(newDeadline) == true
				}() == true {
					__antithesis_instrumentation__.Notify(627395)
					deadline = newDeadline
					timer.Reset(timeutil.Until(deadline))
				} else {
					__antithesis_instrumentation__.Notify(627396)
				}
			}
		}
		poll = func() {
			__antithesis_instrumentation__.Notify(627397)
			if err := r.pollRequests(ctx); err != nil {
				__antithesis_instrumentation__.Notify(627399)
				if ctx.Err() != nil {
					__antithesis_instrumentation__.Notify(627401)
					return
				} else {
					__antithesis_instrumentation__.Notify(627402)
				}
				__antithesis_instrumentation__.Notify(627400)
				log.Warningf(ctx, "error polling for statement diagnostics requests: %s", err)
			} else {
				__antithesis_instrumentation__.Notify(627403)
			}
			__antithesis_instrumentation__.Notify(627398)
			lastPoll = timeutil.Now()
		}
	)
	__antithesis_instrumentation__.Notify(627389)
	pollingInterval.SetOnChange(&r.st.SV, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(627404)
		select {
		case pollIntervalChanged <- struct{}{}:
			__antithesis_instrumentation__.Notify(627405)
		default:
			__antithesis_instrumentation__.Notify(627406)
		}
	})
	__antithesis_instrumentation__.Notify(627390)
	for {
		__antithesis_instrumentation__.Notify(627407)
		maybeResetTimer()
		select {
		case <-pollIntervalChanged:
			__antithesis_instrumentation__.Notify(627409)
			continue
		case reqID := <-r.gossipUpdateChan:
			__antithesis_instrumentation__.Notify(627410)
			if r.findRequest(reqID) {
				__antithesis_instrumentation__.Notify(627414)
				continue
			} else {
				__antithesis_instrumentation__.Notify(627415)
			}

		case reqID := <-r.gossipCancelChan:
			__antithesis_instrumentation__.Notify(627411)
			r.cancelRequest(reqID)

			continue
		case <-timer.C:
			__antithesis_instrumentation__.Notify(627412)
			timer.Read = true
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(627413)
			return
		}
		__antithesis_instrumentation__.Notify(627408)
		poll()
	}
}

func (r *Registry) isMinExecutionLatencySupported(ctx context.Context) bool {
	__antithesis_instrumentation__.Notify(627416)
	return r.st.Version.IsActive(ctx, clusterversion.AlterSystemStmtDiagReqs)
}

type RequestID int

type CollectedInstanceID int

func (r *Registry) addRequestInternalLocked(
	ctx context.Context,
	id RequestID,
	queryFingerprint string,
	minExecutionLatency time.Duration,
	expiresAt time.Time,
) {
	__antithesis_instrumentation__.Notify(627417)
	if r.findRequestLocked(id) {
		__antithesis_instrumentation__.Notify(627420)

		return
	} else {
		__antithesis_instrumentation__.Notify(627421)
	}
	__antithesis_instrumentation__.Notify(627418)
	if r.mu.requestFingerprints == nil {
		__antithesis_instrumentation__.Notify(627422)
		r.mu.requestFingerprints = make(map[RequestID]Request)
	} else {
		__antithesis_instrumentation__.Notify(627423)
	}
	__antithesis_instrumentation__.Notify(627419)
	r.mu.requestFingerprints[id] = Request{
		fingerprint:         queryFingerprint,
		minExecutionLatency: minExecutionLatency,
		expiresAt:           expiresAt,
	}
}

func (r *Registry) findRequest(requestID RequestID) bool {
	__antithesis_instrumentation__.Notify(627424)
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.findRequestLocked(requestID)
}

func (r *Registry) findRequestLocked(requestID RequestID) bool {
	__antithesis_instrumentation__.Notify(627425)
	f, ok := r.mu.requestFingerprints[requestID]
	if ok {
		__antithesis_instrumentation__.Notify(627427)
		if f.isExpired(timeutil.Now()) {
			__antithesis_instrumentation__.Notify(627429)

			delete(r.mu.requestFingerprints, requestID)
		} else {
			__antithesis_instrumentation__.Notify(627430)
		}
		__antithesis_instrumentation__.Notify(627428)
		return true
	} else {
		__antithesis_instrumentation__.Notify(627431)
	}
	__antithesis_instrumentation__.Notify(627426)
	_, ok = r.mu.ongoing[requestID]
	return ok
}

func (r *Registry) cancelRequest(requestID RequestID) {
	__antithesis_instrumentation__.Notify(627432)
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.mu.requestFingerprints, requestID)
	delete(r.mu.ongoing, requestID)
}

func (r *Registry) InsertRequest(
	ctx context.Context,
	stmtFingerprint string,
	minExecutionLatency time.Duration,
	expiresAfter time.Duration,
) error {
	__antithesis_instrumentation__.Notify(627433)
	_, err := r.insertRequestInternal(ctx, stmtFingerprint, minExecutionLatency, expiresAfter)
	return err
}

func (r *Registry) insertRequestInternal(
	ctx context.Context,
	stmtFingerprint string,
	minExecutionLatency time.Duration,
	expiresAfter time.Duration,
) (RequestID, error) {
	__antithesis_instrumentation__.Notify(627434)
	g, err := r.gossip.OptionalErr(48274)
	if err != nil {
		__antithesis_instrumentation__.Notify(627440)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(627441)
	}
	__antithesis_instrumentation__.Notify(627435)

	if !r.isMinExecutionLatencySupported(ctx) {
		__antithesis_instrumentation__.Notify(627442)
		if minExecutionLatency != 0 || func() bool {
			__antithesis_instrumentation__.Notify(627443)
			return expiresAfter != 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(627444)
			return 0, errors.New(
				"conditional statement diagnostics are only supported " +
					"after 22.1 version migrations have completed",
			)
		} else {
			__antithesis_instrumentation__.Notify(627445)
		}
	} else {
		__antithesis_instrumentation__.Notify(627446)
	}
	__antithesis_instrumentation__.Notify(627436)

	var reqID RequestID
	var expiresAt time.Time
	err = r.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(627447)

		var extraConditions string
		if r.isMinExecutionLatencySupported(ctx) {
			__antithesis_instrumentation__.Notify(627457)
			extraConditions = " AND (expires_at IS NULL OR expires_at > now())"
		} else {
			__antithesis_instrumentation__.Notify(627458)
		}
		__antithesis_instrumentation__.Notify(627448)
		row, err := r.ie.QueryRowEx(ctx, "stmt-diag-check-pending", txn,
			sessiondata.InternalExecutorOverride{
				User: security.RootUserName(),
			},
			fmt.Sprintf("SELECT count(1) FROM system.statement_diagnostics_requests "+
				"WHERE completed = false AND statement_fingerprint = $1%s", extraConditions),
			stmtFingerprint)
		if err != nil {
			__antithesis_instrumentation__.Notify(627459)
			return err
		} else {
			__antithesis_instrumentation__.Notify(627460)
		}
		__antithesis_instrumentation__.Notify(627449)
		if row == nil {
			__antithesis_instrumentation__.Notify(627461)
			return errors.New("failed to check pending statement diagnostics")
		} else {
			__antithesis_instrumentation__.Notify(627462)
		}
		__antithesis_instrumentation__.Notify(627450)
		count := int(*row[0].(*tree.DInt))
		if count != 0 {
			__antithesis_instrumentation__.Notify(627463)
			return errors.New(
				"A pending request for the requested fingerprint already exists. " +
					"Cancel the existing request first and try again.",
			)
		} else {
			__antithesis_instrumentation__.Notify(627464)
		}
		__antithesis_instrumentation__.Notify(627451)

		now := timeutil.Now()
		insertColumns := "statement_fingerprint, requested_at"
		qargs := make([]interface{}, 2, 4)
		qargs[0] = stmtFingerprint
		qargs[1] = now
		if minExecutionLatency != 0 {
			__antithesis_instrumentation__.Notify(627465)
			insertColumns += ", min_execution_latency"
			qargs = append(qargs, minExecutionLatency)
		} else {
			__antithesis_instrumentation__.Notify(627466)
		}
		__antithesis_instrumentation__.Notify(627452)
		if expiresAfter != 0 {
			__antithesis_instrumentation__.Notify(627467)
			insertColumns += ", expires_at"
			expiresAt = now.Add(expiresAfter)
			qargs = append(qargs, expiresAt)
		} else {
			__antithesis_instrumentation__.Notify(627468)
		}
		__antithesis_instrumentation__.Notify(627453)
		valuesClause := "$1, $2"
		for i := range qargs[2:] {
			__antithesis_instrumentation__.Notify(627469)
			valuesClause += fmt.Sprintf(", $%d", i+3)
		}
		__antithesis_instrumentation__.Notify(627454)
		stmt := "INSERT INTO system.statement_diagnostics_requests (" +
			insertColumns + ") VALUES (" + valuesClause + ") RETURNING id;"
		row, err = r.ie.QueryRowEx(
			ctx, "stmt-diag-insert-request", txn,
			sessiondata.InternalExecutorOverride{User: security.RootUserName()},
			stmt, qargs...,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(627470)
			return err
		} else {
			__antithesis_instrumentation__.Notify(627471)
		}
		__antithesis_instrumentation__.Notify(627455)
		if row == nil {
			__antithesis_instrumentation__.Notify(627472)
			return errors.New("failed to insert statement diagnostics request")
		} else {
			__antithesis_instrumentation__.Notify(627473)
		}
		__antithesis_instrumentation__.Notify(627456)
		reqID = RequestID(*row[0].(*tree.DInt))
		return nil
	})
	__antithesis_instrumentation__.Notify(627437)
	if err != nil {
		__antithesis_instrumentation__.Notify(627474)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(627475)
	}
	__antithesis_instrumentation__.Notify(627438)

	r.mu.Lock()
	r.mu.epoch++
	r.addRequestInternalLocked(ctx, reqID, stmtFingerprint, minExecutionLatency, expiresAt)
	r.mu.Unlock()

	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(reqID))
	if err := g.AddInfo(gossip.KeyGossipStatementDiagnosticsRequest, buf, 0); err != nil {
		__antithesis_instrumentation__.Notify(627476)
		log.Warningf(ctx, "error notifying of diagnostics request: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(627477)
	}
	__antithesis_instrumentation__.Notify(627439)

	return reqID, nil
}

func (r *Registry) CancelRequest(ctx context.Context, requestID int64) error {
	__antithesis_instrumentation__.Notify(627478)
	g, err := r.gossip.OptionalErr(48274)
	if err != nil {
		__antithesis_instrumentation__.Notify(627484)
		return err
	} else {
		__antithesis_instrumentation__.Notify(627485)
	}
	__antithesis_instrumentation__.Notify(627479)

	if !r.isMinExecutionLatencySupported(ctx) {
		__antithesis_instrumentation__.Notify(627486)

		return errors.New(
			"statement diagnostics can only be canceled after 22.1 version migrations have completed",
		)
	} else {
		__antithesis_instrumentation__.Notify(627487)
	}
	__antithesis_instrumentation__.Notify(627480)

	row, err := r.ie.QueryRowEx(ctx, "stmt-diag-cancel-request", nil,
		sessiondata.InternalExecutorOverride{
			User: security.RootUserName(),
		},

		"UPDATE system.statement_diagnostics_requests SET expires_at = '1970-01-01' "+
			"WHERE completed = false AND id = $1 "+
			"AND (expires_at IS NULL OR expires_at > now()) RETURNING id;",
		requestID,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(627488)
		return err
	} else {
		__antithesis_instrumentation__.Notify(627489)
	}
	__antithesis_instrumentation__.Notify(627481)

	if row == nil {
		__antithesis_instrumentation__.Notify(627490)

		return errors.Newf("no pending request found for the fingerprint: %s", requestID)
	} else {
		__antithesis_instrumentation__.Notify(627491)
	}
	__antithesis_instrumentation__.Notify(627482)

	reqID := RequestID(requestID)
	r.cancelRequest(reqID)

	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(reqID))
	if err := g.AddInfo(gossip.KeyGossipStatementDiagnosticsRequestCancellation, buf, 0); err != nil {
		__antithesis_instrumentation__.Notify(627492)
		log.Warningf(ctx, "error notifying of diagnostics request cancellation: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(627493)
	}
	__antithesis_instrumentation__.Notify(627483)

	return nil
}

func (r *Registry) IsExecLatencyConditionMet(
	requestID RequestID, req Request, execLatency time.Duration,
) bool {
	__antithesis_instrumentation__.Notify(627494)
	if req.minExecutionLatency <= execLatency {
		__antithesis_instrumentation__.Notify(627497)
		return true
	} else {
		__antithesis_instrumentation__.Notify(627498)
	}
	__antithesis_instrumentation__.Notify(627495)

	if req.isExpired(timeutil.Now()) {
		__antithesis_instrumentation__.Notify(627499)
		r.mu.Lock()
		defer r.mu.Unlock()
		delete(r.mu.requestFingerprints, requestID)
	} else {
		__antithesis_instrumentation__.Notify(627500)
	}
	__antithesis_instrumentation__.Notify(627496)
	return false
}

func (r *Registry) RemoveOngoing(requestID RequestID, req Request) {
	__antithesis_instrumentation__.Notify(627501)
	r.mu.Lock()
	defer r.mu.Unlock()
	if req.isConditional() {
		__antithesis_instrumentation__.Notify(627502)
		if req.isExpired(timeutil.Now()) {
			__antithesis_instrumentation__.Notify(627503)
			delete(r.mu.requestFingerprints, requestID)
		} else {
			__antithesis_instrumentation__.Notify(627504)
		}
	} else {
		__antithesis_instrumentation__.Notify(627505)
		delete(r.mu.ongoing, requestID)
	}
}

func (r *Registry) ShouldCollectDiagnostics(
	ctx context.Context, fingerprint string,
) (shouldCollect bool, reqID RequestID, req Request) {
	__antithesis_instrumentation__.Notify(627506)
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.mu.requestFingerprints) == 0 {
		__antithesis_instrumentation__.Notify(627511)
		return false, 0, req
	} else {
		__antithesis_instrumentation__.Notify(627512)
	}
	__antithesis_instrumentation__.Notify(627507)

	for id, f := range r.mu.requestFingerprints {
		__antithesis_instrumentation__.Notify(627513)
		if f.fingerprint == fingerprint {
			__antithesis_instrumentation__.Notify(627514)
			if f.isExpired(timeutil.Now()) {
				__antithesis_instrumentation__.Notify(627516)
				delete(r.mu.requestFingerprints, id)
				return false, 0, req
			} else {
				__antithesis_instrumentation__.Notify(627517)
			}
			__antithesis_instrumentation__.Notify(627515)
			reqID = id
			req = f
			break
		} else {
			__antithesis_instrumentation__.Notify(627518)
		}
	}
	__antithesis_instrumentation__.Notify(627508)

	if reqID == 0 {
		__antithesis_instrumentation__.Notify(627519)
		return false, 0, req
	} else {
		__antithesis_instrumentation__.Notify(627520)
	}
	__antithesis_instrumentation__.Notify(627509)

	if !req.isConditional() {
		__antithesis_instrumentation__.Notify(627521)
		if r.mu.ongoing == nil {
			__antithesis_instrumentation__.Notify(627523)
			r.mu.ongoing = make(map[RequestID]Request)
		} else {
			__antithesis_instrumentation__.Notify(627524)
		}
		__antithesis_instrumentation__.Notify(627522)
		r.mu.ongoing[reqID] = req
		delete(r.mu.requestFingerprints, reqID)
	} else {
		__antithesis_instrumentation__.Notify(627525)
	}
	__antithesis_instrumentation__.Notify(627510)
	return true, reqID, req
}

func (r *Registry) InsertStatementDiagnostics(
	ctx context.Context,
	requestID RequestID,
	stmtFingerprint string,
	stmt string,
	bundle []byte,
	collectionErr error,
) (CollectedInstanceID, error) {
	__antithesis_instrumentation__.Notify(627526)
	var diagID CollectedInstanceID
	err := r.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(627529)
		if requestID != 0 {
			__antithesis_instrumentation__.Notify(627536)
			row, err := r.ie.QueryRowEx(ctx, "stmt-diag-check-completed", txn,
				sessiondata.InternalExecutorOverride{User: security.RootUserName()},
				"SELECT count(1) FROM system.statement_diagnostics_requests WHERE id = $1 AND completed = false",
				requestID)
			if err != nil {
				__antithesis_instrumentation__.Notify(627539)
				return err
			} else {
				__antithesis_instrumentation__.Notify(627540)
			}
			__antithesis_instrumentation__.Notify(627537)
			if row == nil {
				__antithesis_instrumentation__.Notify(627541)
				return errors.New("failed to check completed statement diagnostics")
			} else {
				__antithesis_instrumentation__.Notify(627542)
			}
			__antithesis_instrumentation__.Notify(627538)
			cnt := int(*row[0].(*tree.DInt))
			if cnt == 0 {
				__antithesis_instrumentation__.Notify(627543)

				return nil
			} else {
				__antithesis_instrumentation__.Notify(627544)
			}
		} else {
			__antithesis_instrumentation__.Notify(627545)
		}
		__antithesis_instrumentation__.Notify(627530)

		errorVal := tree.DNull
		if collectionErr != nil {
			__antithesis_instrumentation__.Notify(627546)
			errorVal = tree.NewDString(collectionErr.Error())
		} else {
			__antithesis_instrumentation__.Notify(627547)
		}
		__antithesis_instrumentation__.Notify(627531)

		bundleChunksVal := tree.NewDArray(types.Int)
		for len(bundle) > 0 {
			__antithesis_instrumentation__.Notify(627548)
			chunkSize := int(bundleChunkSize.Get(&r.st.SV))
			chunk := bundle
			if len(chunk) > chunkSize {
				__antithesis_instrumentation__.Notify(627552)
				chunk = chunk[:chunkSize]
			} else {
				__antithesis_instrumentation__.Notify(627553)
			}
			__antithesis_instrumentation__.Notify(627549)
			bundle = bundle[len(chunk):]

			row, err := r.ie.QueryRowEx(
				ctx, "stmt-bundle-chunks-insert", txn,
				sessiondata.InternalExecutorOverride{User: security.RootUserName()},
				"INSERT INTO system.statement_bundle_chunks(description, data) VALUES ($1, $2) RETURNING id",
				"statement diagnostics bundle",
				tree.NewDBytes(tree.DBytes(chunk)),
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(627554)
				return err
			} else {
				__antithesis_instrumentation__.Notify(627555)
			}
			__antithesis_instrumentation__.Notify(627550)
			if row == nil {
				__antithesis_instrumentation__.Notify(627556)
				return errors.New("failed to check statement bundle chunk")
			} else {
				__antithesis_instrumentation__.Notify(627557)
			}
			__antithesis_instrumentation__.Notify(627551)
			chunkID := row[0].(*tree.DInt)
			if err := bundleChunksVal.Append(chunkID); err != nil {
				__antithesis_instrumentation__.Notify(627558)
				return err
			} else {
				__antithesis_instrumentation__.Notify(627559)
			}
		}
		__antithesis_instrumentation__.Notify(627532)

		collectionTime := timeutil.Now()

		row, err := r.ie.QueryRowEx(
			ctx, "stmt-diag-insert", txn,
			sessiondata.InternalExecutorOverride{User: security.RootUserName()},
			"INSERT INTO system.statement_diagnostics "+
				"(statement_fingerprint, statement, collected_at, bundle_chunks, error) "+
				"VALUES ($1, $2, $3, $4, $5) RETURNING id",
			stmtFingerprint, stmt, collectionTime, bundleChunksVal, errorVal,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(627560)
			return err
		} else {
			__antithesis_instrumentation__.Notify(627561)
		}
		__antithesis_instrumentation__.Notify(627533)
		if row == nil {
			__antithesis_instrumentation__.Notify(627562)
			return errors.New("failed to insert statement diagnostics")
		} else {
			__antithesis_instrumentation__.Notify(627563)
		}
		__antithesis_instrumentation__.Notify(627534)
		diagID = CollectedInstanceID(*row[0].(*tree.DInt))

		if requestID != 0 {
			__antithesis_instrumentation__.Notify(627564)

			_, err := r.ie.ExecEx(ctx, "stmt-diag-mark-completed", txn,
				sessiondata.InternalExecutorOverride{User: security.RootUserName()},
				"UPDATE system.statement_diagnostics_requests "+
					"SET completed = true, statement_diagnostics_id = $1 WHERE id = $2",
				diagID, requestID)
			if err != nil {
				__antithesis_instrumentation__.Notify(627565)
				return err
			} else {
				__antithesis_instrumentation__.Notify(627566)
			}
		} else {
			__antithesis_instrumentation__.Notify(627567)

			_, err := r.ie.ExecEx(ctx, "stmt-diag-add-completed", txn,
				sessiondata.InternalExecutorOverride{User: security.RootUserName()},
				"INSERT INTO system.statement_diagnostics_requests"+
					" (completed, statement_fingerprint, statement_diagnostics_id, requested_at)"+
					" VALUES (true, $1, $2, $3)",
				stmtFingerprint, diagID, collectionTime)
			if err != nil {
				__antithesis_instrumentation__.Notify(627568)
				return err
			} else {
				__antithesis_instrumentation__.Notify(627569)
			}
		}
		__antithesis_instrumentation__.Notify(627535)
		return nil
	})
	__antithesis_instrumentation__.Notify(627527)
	if err != nil {
		__antithesis_instrumentation__.Notify(627570)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(627571)
	}
	__antithesis_instrumentation__.Notify(627528)
	return diagID, nil
}

func (r *Registry) pollRequests(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(627572)
	var rows []tree.Datums
	isMinExecutionLatencySupported := r.isMinExecutionLatencySupported(ctx)

	for {
		__antithesis_instrumentation__.Notify(627576)
		r.mu.Lock()
		epoch := r.mu.epoch
		r.mu.Unlock()

		var extraColumns string
		var extraConditions string
		if isMinExecutionLatencySupported {
			__antithesis_instrumentation__.Notify(627582)
			extraColumns = ", min_execution_latency, expires_at"
			extraConditions = " AND (expires_at IS NULL OR expires_at > now())"
		} else {
			__antithesis_instrumentation__.Notify(627583)
		}
		__antithesis_instrumentation__.Notify(627577)
		it, err := r.ie.QueryIteratorEx(ctx, "stmt-diag-poll", nil,
			sessiondata.InternalExecutorOverride{
				User: security.RootUserName(),
			},
			fmt.Sprintf("SELECT id, statement_fingerprint%s FROM system.statement_diagnostics_requests "+
				"WHERE completed = false%s", extraColumns, extraConditions))
		if err != nil {
			__antithesis_instrumentation__.Notify(627584)
			return err
		} else {
			__antithesis_instrumentation__.Notify(627585)
		}
		__antithesis_instrumentation__.Notify(627578)
		rows = rows[:0]
		var ok bool
		for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
			__antithesis_instrumentation__.Notify(627586)
			rows = append(rows, it.Cur())
		}
		__antithesis_instrumentation__.Notify(627579)
		if err != nil {
			__antithesis_instrumentation__.Notify(627587)
			return err
		} else {
			__antithesis_instrumentation__.Notify(627588)
		}
		__antithesis_instrumentation__.Notify(627580)

		r.mu.Lock()

		if r.mu.epoch != epoch {
			__antithesis_instrumentation__.Notify(627589)
			r.mu.Unlock()
			continue
		} else {
			__antithesis_instrumentation__.Notify(627590)
		}
		__antithesis_instrumentation__.Notify(627581)
		break
	}
	__antithesis_instrumentation__.Notify(627573)
	defer r.mu.Unlock()

	now := timeutil.Now()
	var ids util.FastIntSet
	for _, row := range rows {
		__antithesis_instrumentation__.Notify(627591)
		id := RequestID(*row[0].(*tree.DInt))
		stmtFingerprint := string(*row[1].(*tree.DString))
		var minExecutionLatency time.Duration
		var expiresAt time.Time
		if isMinExecutionLatencySupported {
			__antithesis_instrumentation__.Notify(627593)
			if minExecLatency, ok := row[2].(*tree.DInterval); ok {
				__antithesis_instrumentation__.Notify(627595)
				minExecutionLatency = time.Duration(minExecLatency.Nanos())
			} else {
				__antithesis_instrumentation__.Notify(627596)
			}
			__antithesis_instrumentation__.Notify(627594)
			if e, ok := row[3].(*tree.DTimestampTZ); ok {
				__antithesis_instrumentation__.Notify(627597)
				expiresAt = e.Time
			} else {
				__antithesis_instrumentation__.Notify(627598)
			}
		} else {
			__antithesis_instrumentation__.Notify(627599)
		}
		__antithesis_instrumentation__.Notify(627592)
		ids.Add(int(id))
		r.addRequestInternalLocked(ctx, id, stmtFingerprint, minExecutionLatency, expiresAt)
	}
	__antithesis_instrumentation__.Notify(627574)

	for id, req := range r.mu.requestFingerprints {
		__antithesis_instrumentation__.Notify(627600)
		if !ids.Contains(int(id)) || func() bool {
			__antithesis_instrumentation__.Notify(627601)
			return req.isExpired(now) == true
		}() == true {
			__antithesis_instrumentation__.Notify(627602)
			delete(r.mu.requestFingerprints, id)
		} else {
			__antithesis_instrumentation__.Notify(627603)
		}
	}
	__antithesis_instrumentation__.Notify(627575)
	return nil
}

func (r *Registry) gossipNotification(s string, value roachpb.Value) {
	__antithesis_instrumentation__.Notify(627604)
	switch s {
	case gossip.KeyGossipStatementDiagnosticsRequest:
		__antithesis_instrumentation__.Notify(627605)
		select {
		case r.gossipUpdateChan <- RequestID(binary.LittleEndian.Uint64(value.RawBytes)):
			__antithesis_instrumentation__.Notify(627608)
		default:
			__antithesis_instrumentation__.Notify(627609)

		}
	case gossip.KeyGossipStatementDiagnosticsRequestCancellation:
		__antithesis_instrumentation__.Notify(627606)
		select {
		case r.gossipCancelChan <- RequestID(binary.LittleEndian.Uint64(value.RawBytes)):
			__antithesis_instrumentation__.Notify(627610)
		default:
			__antithesis_instrumentation__.Notify(627611)

		}
	default:
		__antithesis_instrumentation__.Notify(627607)

		return
	}
}
