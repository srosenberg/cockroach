package changefeedccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"hash"
	"hash/fnv"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/kvevent"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/builtins"
	"github.com/cockroachdb/cockroach/pkg/util/bufalloc"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
)

const (
	sqlSinkCreateTableStmt = `CREATE TABLE IF NOT EXISTS "%s" (
		topic STRING,
		partition INT,
		message_id INT,
		key BYTES, value BYTES,
		resolved BYTES,
		PRIMARY KEY (topic, partition, message_id)
	)`
	sqlSinkEmitStmt = `INSERT INTO "%s" (topic, partition, message_id, key, value, resolved)`
	sqlSinkEmitCols = 6

	sqlSinkRowBatchSize = 3

	sqlSinkNumPartitions = 3
)

type sqlSink struct {
	db *gosql.DB

	uri        string
	tableName  string
	topicNamer *TopicNamer
	hasher     hash.Hash32

	rowBuf  []interface{}
	scratch bufalloc.ByteAllocator

	metrics *sliMetrics
}

const sqlSinkTableName = `sqlsink`

func makeSQLSink(
	u sinkURL, tableName string, targets []jobspb.ChangefeedTargetSpecification, m *sliMetrics,
) (Sink, error) {
	__antithesis_instrumentation__.Notify(18690)

	u.Scheme = `postgres`

	if u.Path == `` {
		__antithesis_instrumentation__.Notify(18694)
		return nil, errors.Errorf(`must specify database`)
	} else {
		__antithesis_instrumentation__.Notify(18695)
	}
	__antithesis_instrumentation__.Notify(18691)

	topicNamer, err := MakeTopicNamer(targets)
	if err != nil {
		__antithesis_instrumentation__.Notify(18696)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(18697)
	}
	__antithesis_instrumentation__.Notify(18692)

	uri := u.String()
	u.consumeParam(`sslcert`)
	u.consumeParam(`sslkey`)
	u.consumeParam(`sslmode`)
	u.consumeParam(`sslrootcert`)

	if unknownParams := u.remainingQueryParams(); len(unknownParams) > 0 {
		__antithesis_instrumentation__.Notify(18698)
		return nil, errors.Errorf(
			`unknown SQL sink query parameters: %s`, strings.Join(unknownParams, ", "))
	} else {
		__antithesis_instrumentation__.Notify(18699)
	}
	__antithesis_instrumentation__.Notify(18693)

	return &sqlSink{
		uri:        uri,
		tableName:  tableName,
		topicNamer: topicNamer,
		hasher:     fnv.New32a(),
		metrics:    m,
	}, nil
}

func (s *sqlSink) Dial() error {
	__antithesis_instrumentation__.Notify(18700)
	db, err := gosql.Open(`postgres`, s.uri)
	if err != nil {
		__antithesis_instrumentation__.Notify(18703)
		return err
	} else {
		__antithesis_instrumentation__.Notify(18704)
	}
	__antithesis_instrumentation__.Notify(18701)
	if _, err := db.Exec(fmt.Sprintf(sqlSinkCreateTableStmt, s.tableName)); err != nil {
		__antithesis_instrumentation__.Notify(18705)
		db.Close()
		return err
	} else {
		__antithesis_instrumentation__.Notify(18706)
	}
	__antithesis_instrumentation__.Notify(18702)
	s.db = db
	return nil
}

func (s *sqlSink) EmitRow(
	ctx context.Context,
	topicDescr TopicDescriptor,
	key, value []byte,
	updated, mvcc hlc.Timestamp,
	alloc kvevent.Alloc,
) error {
	__antithesis_instrumentation__.Notify(18707)
	defer alloc.Release(ctx)
	defer s.metrics.recordOneMessage()(mvcc, len(key)+len(value), sinkDoesNotCompress)

	topic, err := s.topicNamer.Name(topicDescr)
	if err != nil {
		__antithesis_instrumentation__.Notify(18711)
		return err
	} else {
		__antithesis_instrumentation__.Notify(18712)
	}
	__antithesis_instrumentation__.Notify(18708)

	s.hasher.Reset()
	if _, err := s.hasher.Write(key); err != nil {
		__antithesis_instrumentation__.Notify(18713)
		return err
	} else {
		__antithesis_instrumentation__.Notify(18714)
	}
	__antithesis_instrumentation__.Notify(18709)
	partition := int32(s.hasher.Sum32()) % sqlSinkNumPartitions
	if partition < 0 {
		__antithesis_instrumentation__.Notify(18715)
		partition = -partition
	} else {
		__antithesis_instrumentation__.Notify(18716)
	}
	__antithesis_instrumentation__.Notify(18710)

	var noResolved []byte
	return s.emit(ctx, topic, partition, key, value, noResolved)
}

func (s *sqlSink) EmitResolvedTimestamp(
	ctx context.Context, encoder Encoder, resolved hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(18717)
	defer s.metrics.recordResolvedCallback()()

	var noKey, noValue []byte
	return s.topicNamer.Each(func(topic string) error {
		__antithesis_instrumentation__.Notify(18718)
		payload, err := encoder.EncodeResolvedTimestamp(ctx, topic, resolved)
		if err != nil {
			__antithesis_instrumentation__.Notify(18721)
			return err
		} else {
			__antithesis_instrumentation__.Notify(18722)
		}
		__antithesis_instrumentation__.Notify(18719)
		s.scratch, payload = s.scratch.Copy(payload, 0)
		for partition := int32(0); partition < sqlSinkNumPartitions; partition++ {
			__antithesis_instrumentation__.Notify(18723)
			if err := s.emit(ctx, topic, partition, noKey, noValue, payload); err != nil {
				__antithesis_instrumentation__.Notify(18724)
				return err
			} else {
				__antithesis_instrumentation__.Notify(18725)
			}
		}
		__antithesis_instrumentation__.Notify(18720)
		return nil
	})
}

func (s *sqlSink) Topics() []string {
	__antithesis_instrumentation__.Notify(18726)
	return s.topicNamer.DisplayNamesSlice()
}

func (s *sqlSink) emit(
	ctx context.Context, topic string, partition int32, key, value, resolved []byte,
) error {
	__antithesis_instrumentation__.Notify(18727)

	messageID := builtins.GenerateUniqueInt(base.SQLInstanceID(partition))
	s.rowBuf = append(s.rowBuf, topic, partition, messageID, key, value, resolved)
	if len(s.rowBuf)/sqlSinkEmitCols >= sqlSinkRowBatchSize {
		__antithesis_instrumentation__.Notify(18729)
		return s.Flush(ctx)
	} else {
		__antithesis_instrumentation__.Notify(18730)
	}
	__antithesis_instrumentation__.Notify(18728)
	return nil
}

func (s *sqlSink) Flush(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(18731)
	defer s.metrics.recordFlushRequestCallback()()

	if len(s.rowBuf) == 0 {
		__antithesis_instrumentation__.Notify(18735)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(18736)
	}
	__antithesis_instrumentation__.Notify(18732)

	var stmt strings.Builder
	fmt.Fprintf(&stmt, sqlSinkEmitStmt, s.tableName)
	for i := 0; i < len(s.rowBuf); i++ {
		__antithesis_instrumentation__.Notify(18737)
		if i == 0 {
			__antithesis_instrumentation__.Notify(18739)
			stmt.WriteString(` VALUES (`)
		} else {
			__antithesis_instrumentation__.Notify(18740)
			if i%sqlSinkEmitCols == 0 {
				__antithesis_instrumentation__.Notify(18741)
				stmt.WriteString(`),(`)
			} else {
				__antithesis_instrumentation__.Notify(18742)
				stmt.WriteString(`,`)
			}
		}
		__antithesis_instrumentation__.Notify(18738)
		fmt.Fprintf(&stmt, `$%d`, i+1)
	}
	__antithesis_instrumentation__.Notify(18733)
	stmt.WriteString(`)`)
	_, err := s.db.Exec(stmt.String(), s.rowBuf...)
	if err != nil {
		__antithesis_instrumentation__.Notify(18743)
		return err
	} else {
		__antithesis_instrumentation__.Notify(18744)
	}
	__antithesis_instrumentation__.Notify(18734)
	s.rowBuf = s.rowBuf[:0]
	return nil
}

func (s *sqlSink) Close() error {
	__antithesis_instrumentation__.Notify(18745)
	return s.db.Close()
}
