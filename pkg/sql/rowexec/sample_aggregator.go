package rowexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math"
	"time"

	"github.com/axiomhq/hyperloglog"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/stats"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
)

type sampleAggregator struct {
	execinfra.ProcessorBase

	spec    *execinfrapb.SampleAggregatorSpec
	input   execinfra.RowSource
	inTypes []*types.T
	sr      stats.SampleReservoir

	memAcc mon.BoundAccount

	tempMemAcc mon.BoundAccount

	tableID     descpb.ID
	sampledCols []descpb.ColumnID
	sketches    []sketchInfo

	rankCol      int
	sketchIdxCol int
	numRowsCol   int
	numNullsCol  int
	sumSizeCol   int
	sketchCol    int
	invColIdxCol int
	invIdxKeyCol int

	invSr     map[uint32]*stats.SampleReservoir
	invSketch map[uint32]*sketchInfo
}

var _ execinfra.Processor = &sampleAggregator{}

const sampleAggregatorProcName = "sample aggregator"

var SampleAggregatorProgressInterval = 5 * time.Second

func newSampleAggregator(
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec *execinfrapb.SampleAggregatorSpec,
	input execinfra.RowSource,
	post *execinfrapb.PostProcessSpec,
	output execinfra.RowReceiver,
) (*sampleAggregator, error) {
	__antithesis_instrumentation__.Notify(574463)
	for _, s := range spec.Sketches {
		__antithesis_instrumentation__.Notify(574468)
		if len(s.Columns) == 0 {
			__antithesis_instrumentation__.Notify(574472)
			return nil, errors.Errorf("no columns")
		} else {
			__antithesis_instrumentation__.Notify(574473)
		}
		__antithesis_instrumentation__.Notify(574469)
		if _, ok := supportedSketchTypes[s.SketchType]; !ok {
			__antithesis_instrumentation__.Notify(574474)
			return nil, errors.Errorf("unsupported sketch type %s", s.SketchType)
		} else {
			__antithesis_instrumentation__.Notify(574475)
		}
		__antithesis_instrumentation__.Notify(574470)
		if s.GenerateHistogram && func() bool {
			__antithesis_instrumentation__.Notify(574476)
			return s.HistogramMaxBuckets == 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(574477)
			return nil, errors.Errorf("histogram max buckets not specified")
		} else {
			__antithesis_instrumentation__.Notify(574478)
		}
		__antithesis_instrumentation__.Notify(574471)
		if s.GenerateHistogram && func() bool {
			__antithesis_instrumentation__.Notify(574479)
			return len(s.Columns) != 1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(574480)
			return nil, errors.Errorf("histograms require one column")
		} else {
			__antithesis_instrumentation__.Notify(574481)
		}
	}
	__antithesis_instrumentation__.Notify(574464)

	ctx := flowCtx.EvalCtx.Ctx()

	memMonitor := execinfra.NewLimitedMonitor(ctx, flowCtx.EvalCtx.Mon, flowCtx, "sample-aggregator-mem")
	rankCol := len(input.OutputTypes()) - 8
	s := &sampleAggregator{
		spec:         spec,
		input:        input,
		inTypes:      input.OutputTypes(),
		memAcc:       memMonitor.MakeBoundAccount(),
		tempMemAcc:   memMonitor.MakeBoundAccount(),
		tableID:      spec.TableID,
		sampledCols:  spec.SampledColumnIDs,
		sketches:     make([]sketchInfo, len(spec.Sketches)),
		rankCol:      rankCol,
		sketchIdxCol: rankCol + 1,
		numRowsCol:   rankCol + 2,
		numNullsCol:  rankCol + 3,
		sumSizeCol:   rankCol + 4,
		sketchCol:    rankCol + 5,
		invColIdxCol: rankCol + 6,
		invIdxKeyCol: rankCol + 7,
		invSr:        make(map[uint32]*stats.SampleReservoir, len(spec.InvertedSketches)),
		invSketch:    make(map[uint32]*sketchInfo, len(spec.InvertedSketches)),
	}

	var sampleCols util.FastIntSet
	for i := range spec.Sketches {
		__antithesis_instrumentation__.Notify(574482)
		s.sketches[i] = sketchInfo{
			spec:     spec.Sketches[i],
			sketch:   hyperloglog.New14(),
			numNulls: 0,
			numRows:  0,
		}
		if spec.Sketches[i].GenerateHistogram {
			__antithesis_instrumentation__.Notify(574483)
			sampleCols.Add(int(spec.Sketches[i].Columns[0]))
		} else {
			__antithesis_instrumentation__.Notify(574484)
		}
	}
	__antithesis_instrumentation__.Notify(574465)

	s.sr.Init(
		int(spec.SampleSize), int(spec.MinSampleSize), input.OutputTypes()[:rankCol], &s.memAcc,
		sampleCols,
	)
	for i := range spec.InvertedSketches {
		__antithesis_instrumentation__.Notify(574485)
		var sr stats.SampleReservoir

		var srCols util.FastIntSet
		srCols.Add(0)
		sr.Init(int(spec.SampleSize), int(spec.MinSampleSize), bytesRowType, &s.memAcc, srCols)
		col := spec.InvertedSketches[i].Columns[0]
		s.invSr[col] = &sr
		s.invSketch[col] = &sketchInfo{
			spec:     spec.InvertedSketches[i],
			sketch:   hyperloglog.New14(),
			numNulls: 0,
			numRows:  0,
		}
	}
	__antithesis_instrumentation__.Notify(574466)

	if err := s.Init(
		nil, post, input.OutputTypes(), flowCtx, processorID, output, memMonitor,
		execinfra.ProcStateOpts{
			TrailingMetaCallback: func() []execinfrapb.ProducerMetadata {
				__antithesis_instrumentation__.Notify(574486)
				s.close()
				return nil
			},
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(574487)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(574488)
	}
	__antithesis_instrumentation__.Notify(574467)
	return s, nil
}

func (s *sampleAggregator) pushTrailingMeta(ctx context.Context) {
	__antithesis_instrumentation__.Notify(574489)
	execinfra.SendTraceData(ctx, s.Output)
}

func (s *sampleAggregator) Run(ctx context.Context) {
	__antithesis_instrumentation__.Notify(574490)
	ctx = s.StartInternal(ctx, sampleAggregatorProcName)
	s.input.Start(ctx)

	earlyExit, err := s.mainLoop(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(574492)
		execinfra.DrainAndClose(ctx, s.Output, err, s.pushTrailingMeta, s.input)
	} else {
		__antithesis_instrumentation__.Notify(574493)
		if !earlyExit {
			__antithesis_instrumentation__.Notify(574494)
			s.pushTrailingMeta(ctx)
			s.input.ConsumerClosed()
			s.Output.ProducerDone()
		} else {
			__antithesis_instrumentation__.Notify(574495)
		}
	}
	__antithesis_instrumentation__.Notify(574491)
	s.MoveToDraining(nil)
}

func (s *sampleAggregator) close() {
	__antithesis_instrumentation__.Notify(574496)
	if s.InternalClose() {
		__antithesis_instrumentation__.Notify(574497)
		s.memAcc.Close(s.Ctx)
		s.tempMemAcc.Close(s.Ctx)
		s.MemMonitor.Stop(s.Ctx)
	} else {
		__antithesis_instrumentation__.Notify(574498)
	}
}

func (s *sampleAggregator) mainLoop(ctx context.Context) (earlyExit bool, err error) {
	__antithesis_instrumentation__.Notify(574499)
	var job *jobs.Job
	jobID := s.spec.JobID

	if jobID != 0 {
		__antithesis_instrumentation__.Notify(574504)
		job, err = s.FlowCtx.Cfg.JobRegistry.LoadJob(ctx, s.spec.JobID)
		if err != nil {
			__antithesis_instrumentation__.Notify(574505)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(574506)
		}
	} else {
		__antithesis_instrumentation__.Notify(574507)
	}
	__antithesis_instrumentation__.Notify(574500)

	lastReportedFractionCompleted := float32(-1)

	progFn := func(fractionCompleted float32) error {
		__antithesis_instrumentation__.Notify(574508)
		if jobID == 0 {
			__antithesis_instrumentation__.Notify(574511)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(574512)
		}
		__antithesis_instrumentation__.Notify(574509)

		if fractionCompleted < 1.0 && func() bool {
			__antithesis_instrumentation__.Notify(574513)
			return fractionCompleted < lastReportedFractionCompleted+0.01 == true
		}() == true {
			__antithesis_instrumentation__.Notify(574514)
			return job.CheckStatus(ctx, nil)
		} else {
			__antithesis_instrumentation__.Notify(574515)
		}
		__antithesis_instrumentation__.Notify(574510)
		lastReportedFractionCompleted = fractionCompleted
		return job.FractionProgressed(ctx, nil, jobs.FractionUpdater(fractionCompleted))
	}
	__antithesis_instrumentation__.Notify(574501)

	var rowsProcessed uint64
	progressUpdates := util.Every(SampleAggregatorProgressInterval)
	var da tree.DatumAlloc
	for {
		__antithesis_instrumentation__.Notify(574516)
		row, meta := s.input.Next()
		if meta != nil {
			__antithesis_instrumentation__.Notify(574523)
			if meta.SamplerProgress != nil {
				__antithesis_instrumentation__.Notify(574525)
				rowsProcessed += meta.SamplerProgress.RowsProcessed
				if progressUpdates.ShouldProcess(timeutil.Now()) {
					__antithesis_instrumentation__.Notify(574527)

					var fractionCompleted float32
					if s.spec.RowsExpected > 0 {
						__antithesis_instrumentation__.Notify(574529)
						fractionCompleted = float32(float64(rowsProcessed) / float64(s.spec.RowsExpected))
						const maxProgress = 0.99
						if fractionCompleted > maxProgress {
							__antithesis_instrumentation__.Notify(574530)

							fractionCompleted = maxProgress
						} else {
							__antithesis_instrumentation__.Notify(574531)
						}
					} else {
						__antithesis_instrumentation__.Notify(574532)
					}
					__antithesis_instrumentation__.Notify(574528)

					if err := progFn(fractionCompleted); err != nil {
						__antithesis_instrumentation__.Notify(574533)
						return false, err
					} else {
						__antithesis_instrumentation__.Notify(574534)
					}
				} else {
					__antithesis_instrumentation__.Notify(574535)
				}
				__antithesis_instrumentation__.Notify(574526)
				if meta.SamplerProgress.HistogramDisabled {
					__antithesis_instrumentation__.Notify(574536)

					s.sr.Disable()
					for _, sr := range s.invSr {
						__antithesis_instrumentation__.Notify(574537)
						sr.Disable()
					}
				} else {
					__antithesis_instrumentation__.Notify(574538)
				}
			} else {
				__antithesis_instrumentation__.Notify(574539)
				if !emitHelper(ctx, s.Output, &s.OutputHelper, nil, meta, s.pushTrailingMeta, s.input) {
					__antithesis_instrumentation__.Notify(574540)

					return true, nil
				} else {
					__antithesis_instrumentation__.Notify(574541)
				}
			}
			__antithesis_instrumentation__.Notify(574524)
			continue
		} else {
			__antithesis_instrumentation__.Notify(574542)
		}
		__antithesis_instrumentation__.Notify(574517)
		if row == nil {
			__antithesis_instrumentation__.Notify(574543)
			break
		} else {
			__antithesis_instrumentation__.Notify(574544)
		}
		__antithesis_instrumentation__.Notify(574518)

		if invColIdx, err := row[s.invColIdxCol].GetInt(); err == nil {
			__antithesis_instrumentation__.Notify(574545)
			colIdx := uint32(invColIdx)
			if rank, err := row[s.rankCol].GetInt(); err == nil {
				__antithesis_instrumentation__.Notify(574549)

				s.maybeDecreaseSamples(ctx, s.invSr[colIdx], row)
				sampleRow := row[s.invIdxKeyCol : s.invIdxKeyCol+1]
				if err := s.sampleRow(ctx, s.invSr[colIdx], sampleRow, uint64(rank)); err != nil {
					__antithesis_instrumentation__.Notify(574551)
					return false, err
				} else {
					__antithesis_instrumentation__.Notify(574552)
				}
				__antithesis_instrumentation__.Notify(574550)
				continue
			} else {
				__antithesis_instrumentation__.Notify(574553)
			}
			__antithesis_instrumentation__.Notify(574546)

			invSketch, ok := s.invSketch[colIdx]
			if !ok {
				__antithesis_instrumentation__.Notify(574554)
				return false, errors.AssertionFailedf("unknown inverted sketch")
			} else {
				__antithesis_instrumentation__.Notify(574555)
			}
			__antithesis_instrumentation__.Notify(574547)
			if err := s.processSketchRow(invSketch, row, &da); err != nil {
				__antithesis_instrumentation__.Notify(574556)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(574557)
			}
			__antithesis_instrumentation__.Notify(574548)
			continue
		} else {
			__antithesis_instrumentation__.Notify(574558)
		}
		__antithesis_instrumentation__.Notify(574519)
		if rank, err := row[s.rankCol].GetInt(); err == nil {
			__antithesis_instrumentation__.Notify(574559)

			s.maybeDecreaseSamples(ctx, &s.sr, row)
			if err := s.sampleRow(ctx, &s.sr, row[:s.rankCol], uint64(rank)); err != nil {
				__antithesis_instrumentation__.Notify(574561)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(574562)
			}
			__antithesis_instrumentation__.Notify(574560)
			continue
		} else {
			__antithesis_instrumentation__.Notify(574563)
		}
		__antithesis_instrumentation__.Notify(574520)

		sketchIdx, err := row[s.sketchIdxCol].GetInt()
		if err != nil {
			__antithesis_instrumentation__.Notify(574564)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(574565)
		}
		__antithesis_instrumentation__.Notify(574521)
		if sketchIdx < 0 || func() bool {
			__antithesis_instrumentation__.Notify(574566)
			return sketchIdx > int64(len(s.sketches)) == true
		}() == true {
			__antithesis_instrumentation__.Notify(574567)
			return false, errors.Errorf("invalid sketch index %d", sketchIdx)
		} else {
			__antithesis_instrumentation__.Notify(574568)
		}
		__antithesis_instrumentation__.Notify(574522)
		if err := s.processSketchRow(&s.sketches[sketchIdx], row, &da); err != nil {
			__antithesis_instrumentation__.Notify(574569)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(574570)
		}
	}
	__antithesis_instrumentation__.Notify(574502)

	if err = progFn(1.0); err != nil {
		__antithesis_instrumentation__.Notify(574571)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(574572)
	}
	__antithesis_instrumentation__.Notify(574503)
	return false, s.writeResults(ctx)
}

func (s *sampleAggregator) processSketchRow(
	sketch *sketchInfo, row rowenc.EncDatumRow, da *tree.DatumAlloc,
) error {
	__antithesis_instrumentation__.Notify(574573)
	var tmpSketch hyperloglog.Sketch

	numRows, err := row[s.numRowsCol].GetInt()
	if err != nil {
		__antithesis_instrumentation__.Notify(574581)
		return err
	} else {
		__antithesis_instrumentation__.Notify(574582)
	}
	__antithesis_instrumentation__.Notify(574574)
	sketch.numRows += numRows

	numNulls, err := row[s.numNullsCol].GetInt()
	if err != nil {
		__antithesis_instrumentation__.Notify(574583)
		return err
	} else {
		__antithesis_instrumentation__.Notify(574584)
	}
	__antithesis_instrumentation__.Notify(574575)
	sketch.numNulls += numNulls

	size, err := row[s.sumSizeCol].GetInt()
	if err != nil {
		__antithesis_instrumentation__.Notify(574585)
		return err
	} else {
		__antithesis_instrumentation__.Notify(574586)
	}
	__antithesis_instrumentation__.Notify(574576)
	sketch.size += size

	if err := row[s.sketchCol].EnsureDecoded(s.inTypes[s.sketchCol], da); err != nil {
		__antithesis_instrumentation__.Notify(574587)
		return err
	} else {
		__antithesis_instrumentation__.Notify(574588)
	}
	__antithesis_instrumentation__.Notify(574577)
	d := row[s.sketchCol].Datum
	if d == tree.DNull {
		__antithesis_instrumentation__.Notify(574589)
		return errors.AssertionFailedf("NULL sketch data")
	} else {
		__antithesis_instrumentation__.Notify(574590)
	}
	__antithesis_instrumentation__.Notify(574578)
	if err := tmpSketch.UnmarshalBinary([]byte(*d.(*tree.DBytes))); err != nil {
		__antithesis_instrumentation__.Notify(574591)
		return err
	} else {
		__antithesis_instrumentation__.Notify(574592)
	}
	__antithesis_instrumentation__.Notify(574579)
	if err := sketch.sketch.Merge(&tmpSketch); err != nil {
		__antithesis_instrumentation__.Notify(574593)
		return errors.NewAssertionErrorWithWrappedErrf(err, "merging sketch data")
	} else {
		__antithesis_instrumentation__.Notify(574594)
	}
	__antithesis_instrumentation__.Notify(574580)
	return nil
}

func (s *sampleAggregator) maybeDecreaseSamples(
	ctx context.Context, sr *stats.SampleReservoir, row rowenc.EncDatumRow,
) {
	__antithesis_instrumentation__.Notify(574595)
	if capacity, err := row[s.numRowsCol].GetInt(); err == nil {
		__antithesis_instrumentation__.Notify(574596)
		prevCapacity := sr.Cap()
		if sr.MaybeResize(ctx, int(capacity)) {
			__antithesis_instrumentation__.Notify(574597)
			log.Infof(
				ctx, "histogram samples reduced from %d to %d to match sampler processor",
				prevCapacity, sr.Cap(),
			)
		} else {
			__antithesis_instrumentation__.Notify(574598)
		}
	} else {
		__antithesis_instrumentation__.Notify(574599)
	}
}

func (s *sampleAggregator) sampleRow(
	ctx context.Context, sr *stats.SampleReservoir, sampleRow rowenc.EncDatumRow, rank uint64,
) error {
	__antithesis_instrumentation__.Notify(574600)
	prevCapacity := sr.Cap()
	if err := sr.SampleRow(ctx, s.EvalCtx, sampleRow, rank); err != nil {
		__antithesis_instrumentation__.Notify(574602)
		if code := pgerror.GetPGCode(err); code != pgcode.OutOfMemory {
			__antithesis_instrumentation__.Notify(574604)
			return err
		} else {
			__antithesis_instrumentation__.Notify(574605)
		}
		__antithesis_instrumentation__.Notify(574603)

		sr.Disable()
		log.Info(ctx, "disabling histogram collection due to excessive memory utilization")
		telemetry.Inc(sqltelemetry.StatsHistogramOOMCounter)
	} else {
		__antithesis_instrumentation__.Notify(574606)
		if sr.Cap() != prevCapacity {
			__antithesis_instrumentation__.Notify(574607)
			log.Infof(
				ctx, "histogram samples reduced from %d to %d due to excessive memory utilization",
				prevCapacity, sr.Cap(),
			)
		} else {
			__antithesis_instrumentation__.Notify(574608)
		}
	}
	__antithesis_instrumentation__.Notify(574601)
	return nil
}

func (s *sampleAggregator) writeResults(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(574609)

	if span := tracing.SpanFromContext(ctx); span != nil && func() bool {
		__antithesis_instrumentation__.Notify(574612)
		return span.IsVerbose() == true
	}() == true {
		__antithesis_instrumentation__.Notify(574613)

		ctx = tracing.ContextWithSpan(ctx, nil)
	} else {
		__antithesis_instrumentation__.Notify(574614)
	}
	__antithesis_instrumentation__.Notify(574610)

	if err := s.FlowCtx.Cfg.DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(574615)
		for _, si := range s.sketches {
			__antithesis_instrumentation__.Notify(574617)
			var histogram *stats.HistogramData
			if si.spec.GenerateHistogram && func() bool {
				__antithesis_instrumentation__.Notify(574622)
				return len(s.sr.Get()) != 0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(574623)
				colIdx := int(si.spec.Columns[0])
				typ := s.inTypes[colIdx]

				h, err := s.generateHistogram(
					ctx,
					s.EvalCtx,
					&s.sr,
					colIdx,
					typ,
					si.numRows-si.numNulls,
					s.getDistinctCount(&si, false),
					int(si.spec.HistogramMaxBuckets),
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(574625)
					return err
				} else {
					__antithesis_instrumentation__.Notify(574626)
				}
				__antithesis_instrumentation__.Notify(574624)
				histogram = &h
			} else {
				__antithesis_instrumentation__.Notify(574627)
				if invSr, ok := s.invSr[si.spec.Columns[0]]; ok && func() bool {
					__antithesis_instrumentation__.Notify(574628)
					return len(invSr.Get()) != 0 == true
				}() == true {
					__antithesis_instrumentation__.Notify(574629)
					invSketch, ok := s.invSketch[si.spec.Columns[0]]
					if !ok {
						__antithesis_instrumentation__.Notify(574632)
						return errors.Errorf("no associated inverted sketch")
					} else {
						__antithesis_instrumentation__.Notify(574633)
					}
					__antithesis_instrumentation__.Notify(574630)

					invDistinctCount := s.getDistinctCount(invSketch, false)

					h, err := s.generateHistogram(
						ctx,
						s.EvalCtx,
						invSr,
						0,
						types.Bytes,
						invSketch.numRows-invSketch.numNulls,
						invDistinctCount,
						int(invSketch.spec.HistogramMaxBuckets),
					)
					if err != nil {
						__antithesis_instrumentation__.Notify(574634)
						return err
					} else {
						__antithesis_instrumentation__.Notify(574635)
					}
					__antithesis_instrumentation__.Notify(574631)
					histogram = &h
				} else {
					__antithesis_instrumentation__.Notify(574636)
				}
			}
			__antithesis_instrumentation__.Notify(574618)

			columnIDs := make([]descpb.ColumnID, len(si.spec.Columns))
			for i, c := range si.spec.Columns {
				__antithesis_instrumentation__.Notify(574637)
				columnIDs[i] = s.sampledCols[c]
			}
			__antithesis_instrumentation__.Notify(574619)

			if err := stats.DeleteOldStatsForColumns(
				ctx,
				s.FlowCtx.Cfg.Executor,
				txn,
				s.tableID,
				columnIDs,
			); err != nil {
				__antithesis_instrumentation__.Notify(574638)
				return err
			} else {
				__antithesis_instrumentation__.Notify(574639)
			}
			__antithesis_instrumentation__.Notify(574620)

			if err := stats.InsertNewStat(
				ctx,
				s.FlowCtx.Cfg.Settings,
				s.FlowCtx.Cfg.Executor,
				txn,
				s.tableID,
				si.spec.StatName,
				columnIDs,
				si.numRows,
				s.getDistinctCount(&si, true),
				si.numNulls,
				s.getAvgSize(&si),
				histogram); err != nil {
				__antithesis_instrumentation__.Notify(574640)
				return err
			} else {
				__antithesis_instrumentation__.Notify(574641)
			}
			__antithesis_instrumentation__.Notify(574621)

			s.tempMemAcc.Clear(ctx)
		}
		__antithesis_instrumentation__.Notify(574616)

		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(574642)
		return err
	} else {
		__antithesis_instrumentation__.Notify(574643)
	}
	__antithesis_instrumentation__.Notify(574611)

	return nil
}

func (s *sampleAggregator) getAvgSize(si *sketchInfo) int64 {
	__antithesis_instrumentation__.Notify(574644)
	if si.numRows == 0 {
		__antithesis_instrumentation__.Notify(574646)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(574647)
	}
	__antithesis_instrumentation__.Notify(574645)
	return int64(math.Ceil(float64(si.size) / float64(si.numRows)))
}

func (s *sampleAggregator) getDistinctCount(si *sketchInfo, includeNulls bool) int64 {
	__antithesis_instrumentation__.Notify(574648)
	distinctCount := int64(si.sketch.Estimate())
	if si.numNulls > 0 && func() bool {
		__antithesis_instrumentation__.Notify(574652)
		return !includeNulls == true
	}() == true {
		__antithesis_instrumentation__.Notify(574653)

		distinctCount--
	} else {
		__antithesis_instrumentation__.Notify(574654)
	}
	__antithesis_instrumentation__.Notify(574649)

	maxDistinctCount := si.numRows - si.numNulls
	if si.numNulls > 0 && func() bool {
		__antithesis_instrumentation__.Notify(574655)
		return includeNulls == true
	}() == true {
		__antithesis_instrumentation__.Notify(574656)
		maxDistinctCount++
	} else {
		__antithesis_instrumentation__.Notify(574657)
	}
	__antithesis_instrumentation__.Notify(574650)
	if distinctCount > maxDistinctCount {
		__antithesis_instrumentation__.Notify(574658)
		distinctCount = maxDistinctCount
	} else {
		__antithesis_instrumentation__.Notify(574659)
	}
	__antithesis_instrumentation__.Notify(574651)
	return distinctCount
}

func (s *sampleAggregator) generateHistogram(
	ctx context.Context,
	evalCtx *tree.EvalContext,
	sr *stats.SampleReservoir,
	colIdx int,
	colType *types.T,
	numRows int64,
	distinctCount int64,
	maxBuckets int,
) (stats.HistogramData, error) {
	__antithesis_instrumentation__.Notify(574660)
	prevCapacity := sr.Cap()
	values, err := sr.GetNonNullDatums(ctx, &s.tempMemAcc, colIdx)
	if err != nil {
		__antithesis_instrumentation__.Notify(574663)
		return stats.HistogramData{}, err
	} else {
		__antithesis_instrumentation__.Notify(574664)
	}
	__antithesis_instrumentation__.Notify(574661)
	if sr.Cap() != prevCapacity {
		__antithesis_instrumentation__.Notify(574665)
		log.Infof(
			ctx, "histogram samples reduced from %d to %d due to excessive memory utilization",
			prevCapacity, sr.Cap(),
		)
	} else {
		__antithesis_instrumentation__.Notify(574666)
	}
	__antithesis_instrumentation__.Notify(574662)
	h, _, err := stats.EquiDepthHistogram(evalCtx, colType, values, numRows, distinctCount, maxBuckets)
	return h, err
}

var _ execinfra.DoesNotUseTxn = &sampleAggregator{}

func (s *sampleAggregator) DoesNotUseTxn() bool {
	__antithesis_instrumentation__.Notify(574667)
	txnUser, ok := s.input.(execinfra.DoesNotUseTxn)
	return ok && func() bool {
		__antithesis_instrumentation__.Notify(574668)
		return txnUser.DoesNotUseTxn() == true
	}() == true
}
