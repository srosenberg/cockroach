package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cli/clierrorplus"
	"github.com/cockroachdb/cockroach/pkg/ts"
	"github.com/cockroachdb/cockroach/pkg/ts/tspb"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

var debugTimeSeriesDumpOpts = struct {
	format   tsDumpFormat
	from, to timestampValue
}{
	format: tsDumpText,
	from:   timestampValue{},
	to:     timestampValue(timeutil.Now().Add(24 * time.Hour)),
}

var debugTimeSeriesDumpCmd = &cobra.Command{
	Use:   "tsdump",
	Short: "dump all the raw timeseries values in a cluster",
	Long: `
Dumps all of the raw timeseries values in a cluster. Only the default resolution
is retrieved, i.e. typically datapoints older than the value of the
'timeseries.storage.resolution_10s.ttl' cluster setting will be absent from the
output.
`,
	RunE: clierrorplus.MaybeDecorateError(func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(34834)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		req := &tspb.DumpRequest{
			StartNanos: time.Time(debugTimeSeriesDumpOpts.from).UnixNano(),
			EndNanos:   time.Time(debugTimeSeriesDumpOpts.to).UnixNano(),
		}
		var w tsWriter
		switch debugTimeSeriesDumpOpts.format {
		case tsDumpRaw:
			__antithesis_instrumentation__.Notify(34838)

			conn, _, finish, err := getClientGRPCConn(ctx, serverCfg)
			if err != nil {
				__antithesis_instrumentation__.Notify(34846)
				return err
			} else {
				__antithesis_instrumentation__.Notify(34847)
			}
			__antithesis_instrumentation__.Notify(34839)
			defer finish()

			tsClient := tspb.NewTimeSeriesClient(conn)
			stream, err := tsClient.DumpRaw(context.Background(), req)
			if err != nil {
				__antithesis_instrumentation__.Notify(34848)
				return err
			} else {
				__antithesis_instrumentation__.Notify(34849)
			}
			__antithesis_instrumentation__.Notify(34840)

			w := bufio.NewWriter(os.Stdout)
			if err := ts.DumpRawTo(stream, w); err != nil {
				__antithesis_instrumentation__.Notify(34850)
				return err
			} else {
				__antithesis_instrumentation__.Notify(34851)
			}
			__antithesis_instrumentation__.Notify(34841)
			return w.Flush()
		case tsDumpCSV:
			__antithesis_instrumentation__.Notify(34842)
			w = csvTSWriter{w: csv.NewWriter(os.Stdout)}
		case tsDumpTSV:
			__antithesis_instrumentation__.Notify(34843)
			cw := csvTSWriter{w: csv.NewWriter(os.Stdout)}
			cw.w.Comma = '\t'
			w = cw
		case tsDumpText:
			__antithesis_instrumentation__.Notify(34844)
			w = defaultTSWriter{w: os.Stdout}
		default:
			__antithesis_instrumentation__.Notify(34845)
			return errors.Newf("unknown output format: %v", debugTimeSeriesDumpOpts.format)
		}
		__antithesis_instrumentation__.Notify(34835)

		conn, _, finish, err := getClientGRPCConn(ctx, serverCfg)
		if err != nil {
			__antithesis_instrumentation__.Notify(34852)
			return err
		} else {
			__antithesis_instrumentation__.Notify(34853)
		}
		__antithesis_instrumentation__.Notify(34836)
		defer finish()

		tsClient := tspb.NewTimeSeriesClient(conn)
		stream, err := tsClient.Dump(context.Background(), req)
		if err != nil {
			__antithesis_instrumentation__.Notify(34854)
			return err
		} else {
			__antithesis_instrumentation__.Notify(34855)
		}
		__antithesis_instrumentation__.Notify(34837)

		for {
			__antithesis_instrumentation__.Notify(34856)
			data, err := stream.Recv()
			if err == io.EOF {
				__antithesis_instrumentation__.Notify(34859)
				return w.Flush()
			} else {
				__antithesis_instrumentation__.Notify(34860)
			}
			__antithesis_instrumentation__.Notify(34857)
			if err != nil {
				__antithesis_instrumentation__.Notify(34861)
				return err
			} else {
				__antithesis_instrumentation__.Notify(34862)
			}
			__antithesis_instrumentation__.Notify(34858)
			if err := w.Emit(data); err != nil {
				__antithesis_instrumentation__.Notify(34863)
				return err
			} else {
				__antithesis_instrumentation__.Notify(34864)
			}
		}
	}),
}

type tsWriter interface {
	Emit(*tspb.TimeSeriesData) error
	Flush() error
}

type csvTSWriter struct {
	w *csv.Writer
}

func (w csvTSWriter) Emit(data *tspb.TimeSeriesData) error {
	__antithesis_instrumentation__.Notify(34865)
	for _, d := range data.Datapoints {
		__antithesis_instrumentation__.Notify(34867)
		if err := w.w.Write(
			[]string{data.Name, timeutil.Unix(0, d.TimestampNanos).In(time.UTC).Format(time.RFC3339), data.Source, fmt.Sprint(d.Value)},
		); err != nil {
			__antithesis_instrumentation__.Notify(34868)
			return err
		} else {
			__antithesis_instrumentation__.Notify(34869)
		}
	}
	__antithesis_instrumentation__.Notify(34866)
	return nil
}

func (w csvTSWriter) Flush() error {
	__antithesis_instrumentation__.Notify(34870)
	w.w.Flush()
	return w.w.Error()
}

type defaultTSWriter struct {
	last struct {
		name, source string
	}
	w io.Writer
}

func (w defaultTSWriter) Flush() error { __antithesis_instrumentation__.Notify(34871); return nil }

func (w defaultTSWriter) Emit(data *tspb.TimeSeriesData) error {
	__antithesis_instrumentation__.Notify(34872)
	if w.last.name != data.Name || func() bool {
		__antithesis_instrumentation__.Notify(34875)
		return w.last.source != data.Source == true
	}() == true {
		__antithesis_instrumentation__.Notify(34876)
		w.last.name, w.last.source = data.Name, data.Source
		fmt.Fprintf(w.w, "%s %s\n", data.Name, data.Source)
	} else {
		__antithesis_instrumentation__.Notify(34877)
	}
	__antithesis_instrumentation__.Notify(34873)
	for _, d := range data.Datapoints {
		__antithesis_instrumentation__.Notify(34878)
		fmt.Fprintf(w.w, "%v %v\n", d.TimestampNanos, d.Value)
	}
	__antithesis_instrumentation__.Notify(34874)
	return nil
}

type tsDumpFormat int

const (
	tsDumpText tsDumpFormat = iota
	tsDumpCSV
	tsDumpTSV
	tsDumpRaw
)

func (m *tsDumpFormat) Type() string { __antithesis_instrumentation__.Notify(34879); return "string" }

func (m *tsDumpFormat) String() string {
	__antithesis_instrumentation__.Notify(34880)
	switch *m {
	case tsDumpCSV:
		__antithesis_instrumentation__.Notify(34882)
		return "csv"
	case tsDumpTSV:
		__antithesis_instrumentation__.Notify(34883)
		return "tsv"
	case tsDumpText:
		__antithesis_instrumentation__.Notify(34884)
		return "text"
	case tsDumpRaw:
		__antithesis_instrumentation__.Notify(34885)
		return "raw"
	default:
		__antithesis_instrumentation__.Notify(34886)
	}
	__antithesis_instrumentation__.Notify(34881)
	return ""
}

func (m *tsDumpFormat) Set(s string) error {
	__antithesis_instrumentation__.Notify(34887)
	switch s {
	case "text":
		__antithesis_instrumentation__.Notify(34889)
		*m = tsDumpText
	case "csv":
		__antithesis_instrumentation__.Notify(34890)
		*m = tsDumpCSV
	case "tsv":
		__antithesis_instrumentation__.Notify(34891)
		*m = tsDumpTSV
	case "raw":
		__antithesis_instrumentation__.Notify(34892)
		*m = tsDumpRaw
	default:
		__antithesis_instrumentation__.Notify(34893)
		return fmt.Errorf("invalid value for --format: %s", s)
	}
	__antithesis_instrumentation__.Notify(34888)
	return nil
}
