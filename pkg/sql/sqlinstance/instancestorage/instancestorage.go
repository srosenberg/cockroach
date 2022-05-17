// Package instancestorage package provides API to read from and write to the
// sql_instance system table.
package instancestorage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/multitenant"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlinstance"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlliveness"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

type Storage struct {
	codec    keys.SQLCodec
	db       *kv.DB
	tableID  descpb.ID
	slReader sqlliveness.Reader
	rowcodec rowCodec
}

type instancerow struct {
	instanceID base.SQLInstanceID
	addr       string
	sessionID  sqlliveness.SessionID
	timestamp  hlc.Timestamp
}

func NewTestingStorage(
	db *kv.DB, codec keys.SQLCodec, sqlInstancesTableID descpb.ID, slReader sqlliveness.Reader,
) *Storage {
	__antithesis_instrumentation__.Notify(624055)
	s := &Storage{
		db:       db,
		codec:    codec,
		tableID:  sqlInstancesTableID,
		rowcodec: makeRowCodec(codec),
		slReader: slReader,
	}
	return s
}

func NewStorage(db *kv.DB, codec keys.SQLCodec, slReader sqlliveness.Reader) *Storage {
	__antithesis_instrumentation__.Notify(624056)
	return NewTestingStorage(db, codec, keys.SQLInstancesTableID, slReader)
}

func (s *Storage) CreateInstance(
	ctx context.Context,
	sessionID sqlliveness.SessionID,
	sessionExpiration hlc.Timestamp,
	addr string,
) (instanceID base.SQLInstanceID, _ error) {
	__antithesis_instrumentation__.Notify(624057)
	if len(addr) == 0 {
		__antithesis_instrumentation__.Notify(624062)
		return base.SQLInstanceID(0), errors.New("no address information for instance")
	} else {
		__antithesis_instrumentation__.Notify(624063)
	}
	__antithesis_instrumentation__.Notify(624058)
	if len(sessionID) == 0 {
		__antithesis_instrumentation__.Notify(624064)
		return base.SQLInstanceID(0), errors.New("no session information for instance")
	} else {
		__antithesis_instrumentation__.Notify(624065)
	}
	__antithesis_instrumentation__.Notify(624059)
	ctx = multitenant.WithTenantCostControlExemption(ctx)
	err := s.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(624066)

		err := txn.UpdateDeadline(ctx, sessionExpiration)
		if err != nil {
			__antithesis_instrumentation__.Notify(624070)
			return err
		} else {
			__antithesis_instrumentation__.Notify(624071)
		}
		__antithesis_instrumentation__.Notify(624067)
		rows, err := s.getAllInstanceRows(ctx, txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(624072)
			return err
		} else {
			__antithesis_instrumentation__.Notify(624073)
		}
		__antithesis_instrumentation__.Notify(624068)
		instanceID = s.getAvailableInstanceID(ctx, rows)
		row, err := s.rowcodec.encodeRow(instanceID, addr, sessionID, s.codec, s.tableID)
		if err != nil {
			__antithesis_instrumentation__.Notify(624074)
			log.Warningf(ctx, "failed to encode row for instance id %d: %v", instanceID, err)
			return err
		} else {
			__antithesis_instrumentation__.Notify(624075)
		}
		__antithesis_instrumentation__.Notify(624069)
		return txn.Put(ctx, row.Key, row.Value)
	})
	__antithesis_instrumentation__.Notify(624060)

	if err != nil {
		__antithesis_instrumentation__.Notify(624076)
		return base.SQLInstanceID(0), err
	} else {
		__antithesis_instrumentation__.Notify(624077)
	}
	__antithesis_instrumentation__.Notify(624061)
	return instanceID, nil
}

func (s *Storage) getAvailableInstanceID(
	ctx context.Context, rows []instancerow,
) base.SQLInstanceID {
	__antithesis_instrumentation__.Notify(624078)
	if len(rows) == 0 {
		__antithesis_instrumentation__.Notify(624082)

		return base.SQLInstanceID(1)
	} else {
		__antithesis_instrumentation__.Notify(624083)
	}
	__antithesis_instrumentation__.Notify(624079)

	sort.SliceStable(rows, func(idx1, idx2 int) bool {
		__antithesis_instrumentation__.Notify(624084)
		return rows[idx1].instanceID < rows[idx2].instanceID
	})
	__antithesis_instrumentation__.Notify(624080)
	instanceCnt := len(rows)

	instanceID := rows[instanceCnt-1].instanceID + 1

	prevInstanceID := base.SQLInstanceID(0)
	for i := 0; i < instanceCnt; i++ {
		__antithesis_instrumentation__.Notify(624085)

		if rows[i].instanceID-prevInstanceID > 1 {
			__antithesis_instrumentation__.Notify(624088)
			instanceID = prevInstanceID + 1
			break
		} else {
			__antithesis_instrumentation__.Notify(624089)
		}
		__antithesis_instrumentation__.Notify(624086)

		sessionAlive, _ := s.slReader.IsAlive(ctx, rows[i].sessionID)
		if !sessionAlive {
			__antithesis_instrumentation__.Notify(624090)
			instanceID = rows[i].instanceID
			break
		} else {
			__antithesis_instrumentation__.Notify(624091)
		}
		__antithesis_instrumentation__.Notify(624087)
		prevInstanceID = rows[i].instanceID
	}
	__antithesis_instrumentation__.Notify(624081)
	return instanceID
}

func (s *Storage) getInstanceData(
	ctx context.Context, instanceID base.SQLInstanceID,
) (instanceData instancerow, _ error) {
	__antithesis_instrumentation__.Notify(624092)
	k := makeInstanceKey(s.codec, s.tableID, instanceID)
	ctx = multitenant.WithTenantCostControlExemption(ctx)
	row, err := s.db.Get(ctx, k)
	if err != nil {
		__antithesis_instrumentation__.Notify(624096)
		return instancerow{}, errors.Wrapf(err, "could not fetch instance %d", instanceID)
	} else {
		__antithesis_instrumentation__.Notify(624097)
	}
	__antithesis_instrumentation__.Notify(624093)
	if row.Value == nil {
		__antithesis_instrumentation__.Notify(624098)
		return instancerow{}, sqlinstance.NonExistentInstanceError
	} else {
		__antithesis_instrumentation__.Notify(624099)
	}
	__antithesis_instrumentation__.Notify(624094)
	_, addr, sessionID, timestamp, _, err := s.rowcodec.decodeRow(row)
	if err != nil {
		__antithesis_instrumentation__.Notify(624100)
		return instancerow{}, errors.Wrapf(err, "could not decode data for instance %d", instanceID)
	} else {
		__antithesis_instrumentation__.Notify(624101)
	}
	__antithesis_instrumentation__.Notify(624095)
	instanceData = instancerow{
		instanceID: instanceID,
		addr:       addr,
		sessionID:  sessionID,
		timestamp:  timestamp,
	}
	return instanceData, nil
}

func (s *Storage) getAllInstancesData(ctx context.Context) (instances []instancerow, err error) {
	__antithesis_instrumentation__.Notify(624102)
	ctx = multitenant.WithTenantCostControlExemption(ctx)
	err = s.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(624105)
		instances, err = s.getAllInstanceRows(ctx, txn)
		return err
	})
	__antithesis_instrumentation__.Notify(624103)
	if err != nil {
		__antithesis_instrumentation__.Notify(624106)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(624107)
	}
	__antithesis_instrumentation__.Notify(624104)
	return instances, nil
}

func (s *Storage) getAllInstanceRows(
	ctx context.Context, txn *kv.Txn,
) (instances []instancerow, _ error) {
	__antithesis_instrumentation__.Notify(624108)
	start := makeTablePrefix(s.codec, s.tableID)
	end := start.PrefixEnd()

	const maxRows = 0
	rows, err := txn.Scan(ctx, start, end, maxRows)
	if err != nil {
		__antithesis_instrumentation__.Notify(624111)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(624112)
	}
	__antithesis_instrumentation__.Notify(624109)
	for i := range rows {
		__antithesis_instrumentation__.Notify(624113)
		instanceID, addr, sessionID, timestamp, _, err := s.rowcodec.decodeRow(rows[i])
		if err != nil {
			__antithesis_instrumentation__.Notify(624115)
			log.Warningf(ctx, "failed to decode row %v: %v", rows[i].Key, err)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(624116)
		}
		__antithesis_instrumentation__.Notify(624114)
		curInstance := instancerow{
			instanceID: instanceID,
			addr:       addr,
			sessionID:  sessionID,
			timestamp:  timestamp,
		}
		instances = append(instances, curInstance)
	}
	__antithesis_instrumentation__.Notify(624110)
	return instances, nil
}

func (s *Storage) ReleaseInstanceID(ctx context.Context, id base.SQLInstanceID) error {
	__antithesis_instrumentation__.Notify(624117)
	key := makeInstanceKey(s.codec, s.tableID, id)
	ctx = multitenant.WithTenantCostControlExemption(ctx)
	if err := s.db.Del(ctx, key); err != nil {
		__antithesis_instrumentation__.Notify(624119)
		return errors.Wrapf(err, "could not delete instance %d", id)
	} else {
		__antithesis_instrumentation__.Notify(624120)
	}
	__antithesis_instrumentation__.Notify(624118)
	return nil
}
