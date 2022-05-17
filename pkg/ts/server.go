package ts

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/ts/catalog"
	"github.com/cockroachdb/cockroach/pkg/ts/tspb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/quotapool"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	URLPrefix = "/ts/"

	queryWorkerMax = 8

	DefaultQueryMemoryMax = int64(64 * 1024 * 1024)

	dumpBatchSize = 100
)

type ClusterNodeCountFn func() int64

type ServerConfig struct {
	QueryWorkerMax int

	QueryMemoryMax int64
}

type Server struct {
	log.AmbientContext
	db               *DB
	stopper          *stop.Stopper
	nodeCountFn      ClusterNodeCountFn
	queryMemoryMax   int64
	queryWorkerMax   int
	workerMemMonitor *mon.BytesMonitor
	resultMemMonitor *mon.BytesMonitor
	workerSem        *quotapool.IntPool
}

func MakeServer(
	ambient log.AmbientContext,
	db *DB,
	nodeCountFn ClusterNodeCountFn,
	cfg ServerConfig,
	memoryMonitor *mon.BytesMonitor,
	stopper *stop.Stopper,
) Server {
	__antithesis_instrumentation__.Notify(648584)
	ambient.AddLogTag("ts-srv", nil)
	ctx := ambient.AnnotateCtx(context.Background())

	queryWorkerMax := queryWorkerMax
	if cfg.QueryWorkerMax != 0 {
		__antithesis_instrumentation__.Notify(648589)
		queryWorkerMax = cfg.QueryWorkerMax
	} else {
		__antithesis_instrumentation__.Notify(648590)
	}
	__antithesis_instrumentation__.Notify(648585)
	queryMemoryMax := DefaultQueryMemoryMax
	if cfg.QueryMemoryMax > DefaultQueryMemoryMax {
		__antithesis_instrumentation__.Notify(648591)
		queryMemoryMax = cfg.QueryMemoryMax
	} else {
		__antithesis_instrumentation__.Notify(648592)
	}
	__antithesis_instrumentation__.Notify(648586)
	workerSem := quotapool.NewIntPool("ts.Server worker", uint64(queryWorkerMax))
	stopper.AddCloser(workerSem.Closer("stopper"))
	s := Server{
		AmbientContext: ambient,
		db:             db,
		stopper:        stopper,
		nodeCountFn:    nodeCountFn,
		workerMemMonitor: mon.NewMonitorInheritWithLimit(
			"timeseries-workers",
			queryMemoryMax*2,
			memoryMonitor,
		),
		resultMemMonitor: mon.NewMonitorInheritWithLimit(
			"timeseries-results",
			math.MaxInt64,
			memoryMonitor,
		),
		queryMemoryMax: queryMemoryMax,
		queryWorkerMax: queryWorkerMax,
		workerSem:      workerSem,
	}

	s.workerMemMonitor.Start(ctx, memoryMonitor, mon.BoundAccount{})
	stopper.AddCloser(stop.CloserFn(func() {
		__antithesis_instrumentation__.Notify(648593)
		s.workerMemMonitor.Stop(ctx)
	}))
	__antithesis_instrumentation__.Notify(648587)

	s.resultMemMonitor.Start(ambient.AnnotateCtx(context.Background()), memoryMonitor, mon.BoundAccount{})
	stopper.AddCloser(stop.CloserFn(func() {
		__antithesis_instrumentation__.Notify(648594)
		s.resultMemMonitor.Stop(ctx)
	}))
	__antithesis_instrumentation__.Notify(648588)

	return s
}

func (s *Server) RegisterService(g *grpc.Server) {
	__antithesis_instrumentation__.Notify(648595)
	tspb.RegisterTimeSeriesServer(g, s)
}

func (s *Server) RegisterGateway(
	ctx context.Context, mux *gwruntime.ServeMux, conn *grpc.ClientConn,
) error {
	__antithesis_instrumentation__.Notify(648596)
	return tspb.RegisterTimeSeriesHandler(ctx, mux, conn)
}

func (s *Server) Query(
	ctx context.Context, request *tspb.TimeSeriesQueryRequest,
) (*tspb.TimeSeriesQueryResponse, error) {
	__antithesis_instrumentation__.Notify(648597)
	ctx = s.AnnotateCtx(ctx)
	if len(request.Queries) == 0 {
		__antithesis_instrumentation__.Notify(648604)
		return nil, status.Errorf(codes.InvalidArgument, "Queries cannot be empty")
	} else {
		__antithesis_instrumentation__.Notify(648605)
	}
	__antithesis_instrumentation__.Notify(648598)

	sampleNanos := request.SampleNanos
	if sampleNanos == 0 {
		__antithesis_instrumentation__.Notify(648606)
		sampleNanos = Resolution10s.SampleDuration()
	} else {
		__antithesis_instrumentation__.Notify(648607)
	}
	__antithesis_instrumentation__.Notify(648599)

	interpolationLimit := kvserver.TimeUntilStoreDead.Get(&s.db.st.SV).Nanoseconds()

	estimatedClusterNodeCount := s.nodeCountFn()
	if estimatedClusterNodeCount == 0 {
		__antithesis_instrumentation__.Notify(648608)
		estimatedClusterNodeCount = 1
	} else {
		__antithesis_instrumentation__.Notify(648609)
	}
	__antithesis_instrumentation__.Notify(648600)

	response := tspb.TimeSeriesQueryResponse{
		Results: make([]tspb.TimeSeriesQueryResponse_Result, len(request.Queries)),
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	workerOutput := make(chan error)

	memContexts := make([]QueryMemoryContext, len(request.Queries))
	defer func() {
		__antithesis_instrumentation__.Notify(648610)
		for idx := range memContexts {
			__antithesis_instrumentation__.Notify(648611)
			memContexts[idx].Close(ctx)
		}
	}()
	__antithesis_instrumentation__.Notify(648601)

	timespan := QueryTimespan{
		StartNanos:          request.StartNanos,
		EndNanos:            request.EndNanos,
		SampleDurationNanos: sampleNanos,
		NowNanos:            timeutil.Now().UnixNano(),
	}

	if err := s.stopper.RunAsyncTask(ctx, "ts.Server: queries", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(648612)
		for queryIdx, query := range request.Queries {
			__antithesis_instrumentation__.Notify(648613)
			queryIdx := queryIdx
			query := query

			if err := s.stopper.RunAsyncTaskEx(
				ctx,
				stop.TaskOpts{
					TaskName:   "ts.Server: query",
					Sem:        s.workerSem,
					WaitForSem: true,
				},
				func(ctx context.Context) {
					__antithesis_instrumentation__.Notify(648614)

					var estimatedSourceCount int64
					if len(query.Sources) > 0 {
						__antithesis_instrumentation__.Notify(648617)
						estimatedSourceCount = int64(len(query.Sources))
					} else {
						__antithesis_instrumentation__.Notify(648618)
						estimatedSourceCount = estimatedClusterNodeCount
					}
					__antithesis_instrumentation__.Notify(648615)

					memContexts[queryIdx] = MakeQueryMemoryContext(
						s.workerMemMonitor,
						s.resultMemMonitor,
						QueryMemoryOptions{
							BudgetBytes:             s.queryMemoryMax / int64(s.queryWorkerMax),
							EstimatedSources:        estimatedSourceCount,
							InterpolationLimitNanos: interpolationLimit,
						},
					)

					datapoints, sources, err := s.db.Query(
						ctx,
						query,
						Resolution10s,
						timespan,
						memContexts[queryIdx],
					)
					if err == nil {
						__antithesis_instrumentation__.Notify(648619)
						response.Results[queryIdx] = tspb.TimeSeriesQueryResponse_Result{
							Query:      query,
							Datapoints: datapoints,
						}
						response.Results[queryIdx].Sources = sources
					} else {
						__antithesis_instrumentation__.Notify(648620)
					}
					__antithesis_instrumentation__.Notify(648616)
					select {
					case workerOutput <- err:
						__antithesis_instrumentation__.Notify(648621)
					case <-ctx.Done():
						__antithesis_instrumentation__.Notify(648622)
					}
				},
			); err != nil {
				__antithesis_instrumentation__.Notify(648623)

				select {
				case workerOutput <- err:
					__antithesis_instrumentation__.Notify(648625)
				case <-ctx.Done():
					__antithesis_instrumentation__.Notify(648626)
				}
				__antithesis_instrumentation__.Notify(648624)
				return
			} else {
				__antithesis_instrumentation__.Notify(648627)
			}
		}
	}); err != nil {
		__antithesis_instrumentation__.Notify(648628)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(648629)
	}
	__antithesis_instrumentation__.Notify(648602)

	for range request.Queries {
		__antithesis_instrumentation__.Notify(648630)
		select {
		case err := <-workerOutput:
			__antithesis_instrumentation__.Notify(648631)
			if err != nil {
				__antithesis_instrumentation__.Notify(648633)

				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(648634)
			}
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(648632)
			return nil, ctx.Err()
		}
	}
	__antithesis_instrumentation__.Notify(648603)

	return &response, nil
}

func (s *Server) Dump(req *tspb.DumpRequest, stream tspb.TimeSeries_DumpServer) error {
	__antithesis_instrumentation__.Notify(648635)
	d := defaultDumper{stream}.Dump
	return dumpImpl(stream.Context(), s.db.db, req, d)

}

func (s *Server) DumpRaw(req *tspb.DumpRequest, stream tspb.TimeSeries_DumpRawServer) error {
	__antithesis_instrumentation__.Notify(648636)
	d := rawDumper{stream}.Dump
	return dumpImpl(stream.Context(), s.db.db, req, d)
}

func dumpImpl(
	ctx context.Context, db *kv.DB, req *tspb.DumpRequest, d func(*roachpb.KeyValue) error,
) error {
	__antithesis_instrumentation__.Notify(648637)
	names := req.Names
	if len(names) == 0 {
		__antithesis_instrumentation__.Notify(648641)
		names = catalog.AllInternalTimeseriesMetricNames()
	} else {
		__antithesis_instrumentation__.Notify(648642)
	}
	__antithesis_instrumentation__.Notify(648638)
	resolutions := req.Resolutions
	if len(resolutions) == 0 {
		__antithesis_instrumentation__.Notify(648643)
		resolutions = []tspb.TimeSeriesResolution{tspb.TimeSeriesResolution_RESOLUTION_10S}
	} else {
		__antithesis_instrumentation__.Notify(648644)
	}
	__antithesis_instrumentation__.Notify(648639)
	for _, seriesName := range names {
		__antithesis_instrumentation__.Notify(648645)
		for _, res := range resolutions {
			__antithesis_instrumentation__.Notify(648646)
			if err := dumpTimeseriesAllSources(
				ctx,
				db,
				seriesName,
				ResolutionFromProto(res),
				req.StartNanos,
				req.EndNanos,
				d,
			); err != nil {
				__antithesis_instrumentation__.Notify(648647)
				return err
			} else {
				__antithesis_instrumentation__.Notify(648648)
			}
		}
	}
	__antithesis_instrumentation__.Notify(648640)
	return nil
}

type defaultDumper struct {
	stream tspb.TimeSeries_DumpServer
}

func (dd defaultDumper) Dump(kv *roachpb.KeyValue) error {
	__antithesis_instrumentation__.Notify(648649)
	name, source, _, _, err := DecodeDataKey(kv.Key)
	if err != nil {
		__antithesis_instrumentation__.Notify(648653)
		return err
	} else {
		__antithesis_instrumentation__.Notify(648654)
	}
	__antithesis_instrumentation__.Notify(648650)
	var idata roachpb.InternalTimeSeriesData
	if err := kv.Value.GetProto(&idata); err != nil {
		__antithesis_instrumentation__.Notify(648655)
		return err
	} else {
		__antithesis_instrumentation__.Notify(648656)
	}
	__antithesis_instrumentation__.Notify(648651)

	tsdata := &tspb.TimeSeriesData{
		Name:       name,
		Source:     source,
		Datapoints: make([]tspb.TimeSeriesDatapoint, idata.SampleCount()),
	}
	for i := 0; i < idata.SampleCount(); i++ {
		__antithesis_instrumentation__.Notify(648657)
		if idata.IsColumnar() {
			__antithesis_instrumentation__.Notify(648658)
			tsdata.Datapoints[i].TimestampNanos = idata.TimestampForOffset(idata.Offset[i])
			tsdata.Datapoints[i].Value = idata.Last[i]
		} else {
			__antithesis_instrumentation__.Notify(648659)
			tsdata.Datapoints[i].TimestampNanos = idata.TimestampForOffset(idata.Samples[i].Offset)
			tsdata.Datapoints[i].Value = idata.Samples[i].Sum
		}
	}
	__antithesis_instrumentation__.Notify(648652)
	return dd.stream.Send(tsdata)
}

type rawDumper struct {
	stream tspb.TimeSeries_DumpRawServer
}

func (rd rawDumper) Dump(kv *roachpb.KeyValue) error {
	__antithesis_instrumentation__.Notify(648660)
	return rd.stream.Send(kv)
}

func dumpTimeseriesAllSources(
	ctx context.Context,
	db *kv.DB,
	seriesName string,
	diskResolution Resolution,
	startNanos, endNanos int64,
	dump func(*roachpb.KeyValue) error,
) error {
	__antithesis_instrumentation__.Notify(648661)
	if endNanos == 0 {
		__antithesis_instrumentation__.Notify(648665)
		endNanos = math.MaxInt64
	} else {
		__antithesis_instrumentation__.Notify(648666)
	}
	__antithesis_instrumentation__.Notify(648662)

	if delta := diskResolution.SlabDuration() - 1; endNanos > math.MaxInt64-delta {
		__antithesis_instrumentation__.Notify(648667)
		endNanos = math.MaxInt64
	} else {
		__antithesis_instrumentation__.Notify(648668)
		endNanos += delta
	}
	__antithesis_instrumentation__.Notify(648663)

	span := &roachpb.Span{
		Key: MakeDataKey(
			seriesName, "", diskResolution, startNanos,
		),
		EndKey: MakeDataKey(
			seriesName, "", diskResolution, endNanos,
		),
	}

	for span != nil {
		__antithesis_instrumentation__.Notify(648669)
		b := &kv.Batch{}
		scan := roachpb.NewScan(span.Key, span.EndKey, false)
		b.AddRawRequest(scan)
		b.Header.MaxSpanRequestKeys = dumpBatchSize
		err := db.Run(ctx, b)
		if err != nil {
			__antithesis_instrumentation__.Notify(648671)
			return err
		} else {
			__antithesis_instrumentation__.Notify(648672)
		}
		__antithesis_instrumentation__.Notify(648670)
		resp := b.RawResponse().Responses[0].GetScan()
		span = resp.ResumeSpan
		for i := range resp.Rows {
			__antithesis_instrumentation__.Notify(648673)
			if err := dump(&resp.Rows[i]); err != nil {
				__antithesis_instrumentation__.Notify(648674)
				return err
			} else {
				__antithesis_instrumentation__.Notify(648675)
			}
		}
	}
	__antithesis_instrumentation__.Notify(648664)
	return nil
}
