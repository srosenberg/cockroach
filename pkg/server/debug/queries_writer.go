package debug

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"io"
	"strconv"

	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/util/encoding/csv"
	"github.com/cockroachdb/cockroach/pkg/util/uint128"
)

var csvHeader = []string{

	"ID",

	"Session ID",

	"TxnID",

	"Start Time",

	"Session Alloc. Bytes",

	"SQL Statement Fingerprint",

	"IsDistributed",

	"Phase",

	"Progress",
}

type ActiveQueriesWriter struct {
	sessions []serverpb.Session
	writer   io.Writer
}

func NewActiveQueriesWriter(s []serverpb.Session, w io.Writer) *ActiveQueriesWriter {
	__antithesis_instrumentation__.Notify(190447)
	return &ActiveQueriesWriter{
		sessions: s,
		writer:   w,
	}
}

func (qw *ActiveQueriesWriter) Write() error {
	__antithesis_instrumentation__.Notify(190448)
	csvWriter := csv.NewWriter(qw.writer)
	defer csvWriter.Flush()
	err := csvWriter.Write(csvHeader)
	if err != nil {
		__antithesis_instrumentation__.Notify(190451)
		return err
	} else {
		__antithesis_instrumentation__.Notify(190452)
	}
	__antithesis_instrumentation__.Notify(190449)
	for _, session := range qw.sessions {
		__antithesis_instrumentation__.Notify(190453)
		for _, q := range session.ActiveQueries {
			__antithesis_instrumentation__.Notify(190454)
			row := []string{
				q.ID,
				uint128.FromBytes(session.ID).String(),
				q.TxnID.String(),
				q.Start.String(),
				strconv.FormatInt(session.AllocBytes, 10),
				q.SqlNoConstants,
				strconv.FormatBool(q.IsDistributed),
				q.Phase.String(),
				strconv.FormatFloat(float64(q.Progress), 'f', 3, 64),
			}
			err := csvWriter.Write(row)
			if err != nil {
				__antithesis_instrumentation__.Notify(190455)
				return err
			} else {
				__antithesis_instrumentation__.Notify(190456)
			}
		}
	}
	__antithesis_instrumentation__.Notify(190450)
	return nil
}
