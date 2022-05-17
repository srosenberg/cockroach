package rowexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/binary"
	"math/rand"
	"time"

	"github.com/axiomhq/hyperloglog"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/stats"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/randutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

type sketchInfo struct {
	spec     execinfrapb.SketchSpec
	sketch   *hyperloglog.Sketch
	numNulls int64
	numRows  int64
	size     int64
}

type samplerProcessor struct {
	execinfra.ProcessorBase

	flowCtx         *execinfra.FlowCtx
	input           execinfra.RowSource
	memAcc          mon.BoundAccount
	sr              stats.SampleReservoir
	sketches        []sketchInfo
	outTypes        []*types.T
	maxFractionIdle float64

	invSr     map[uint32]*stats.SampleReservoir
	invSketch map[uint32]*sketchInfo

	rankCol      int
	sketchIdxCol int
	numRowsCol   int
	numNullsCol  int
	sizeCol      int
	sketchCol    int
	invColIdxCol int
	invIdxKeyCol int
}

var _ execinfra.Processor = &samplerProcessor{}

const samplerProcName = "sampler"

var SamplerProgressInterval = 10000

var supportedSketchTypes = map[execinfrapb.SketchType]struct{}{

	execinfrapb.SketchType_HLL_PLUS_PLUS_V1: {},
}

const maxIdleSleepTime = 10 * time.Second

const cpuUsageMinThrottle = 0.25

const cpuUsageMaxThrottle = 0.75

var bytesRowType = []*types.T{types.Bytes}

func newSamplerProcessor(
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec *execinfrapb.SamplerSpec,
	input execinfra.RowSource,
	post *execinfrapb.PostProcessSpec,
	output execinfra.RowReceiver,
) (*samplerProcessor, error) {
	__antithesis_instrumentation__.Notify(574669)
	for _, s := range spec.Sketches {
		__antithesis_instrumentation__.Notify(574674)
		if _, ok := supportedSketchTypes[s.SketchType]; !ok {
			__antithesis_instrumentation__.Notify(574675)
			return nil, errors.Errorf("unsupported sketch type %s", s.SketchType)
		} else {
			__antithesis_instrumentation__.Notify(574676)
		}
	}
	__antithesis_instrumentation__.Notify(574670)

	ctx := flowCtx.EvalCtx.Ctx()

	memMonitor := execinfra.NewLimitedMonitor(ctx, flowCtx.EvalCtx.Mon, flowCtx, "sampler-mem")
	s := &samplerProcessor{
		flowCtx:         flowCtx,
		input:           input,
		memAcc:          memMonitor.MakeBoundAccount(),
		sketches:        make([]sketchInfo, len(spec.Sketches)),
		maxFractionIdle: spec.MaxFractionIdle,
		invSr:           make(map[uint32]*stats.SampleReservoir, len(spec.InvertedSketches)),
		invSketch:       make(map[uint32]*sketchInfo, len(spec.InvertedSketches)),
	}

	inTypes := input.OutputTypes()
	var sampleCols util.FastIntSet
	for i := range spec.Sketches {
		__antithesis_instrumentation__.Notify(574677)
		s.sketches[i] = sketchInfo{
			spec:     spec.Sketches[i],
			sketch:   hyperloglog.New14(),
			numNulls: 0,
			numRows:  0,
		}
		if spec.Sketches[i].GenerateHistogram {
			__antithesis_instrumentation__.Notify(574678)
			sampleCols.Add(int(spec.Sketches[i].Columns[0]))
		} else {
			__antithesis_instrumentation__.Notify(574679)
		}
	}
	__antithesis_instrumentation__.Notify(574671)
	for i := range spec.InvertedSketches {
		__antithesis_instrumentation__.Notify(574680)
		var sr stats.SampleReservoir

		var srCols util.FastIntSet
		srCols.Add(0)
		sr.Init(int(spec.SampleSize), int(spec.MinSampleSize), bytesRowType, &s.memAcc, srCols)
		col := spec.InvertedSketches[i].Columns[0]
		s.invSr[col] = &sr
		sketchSpec := spec.InvertedSketches[i]

		sketchSpec.Columns = []uint32{0}
		s.invSketch[col] = &sketchInfo{
			spec:     sketchSpec,
			sketch:   hyperloglog.New14(),
			numNulls: 0,
			numRows:  0,
		}
	}
	__antithesis_instrumentation__.Notify(574672)

	s.sr.Init(int(spec.SampleSize), int(spec.MinSampleSize), inTypes, &s.memAcc, sampleCols)

	outTypes := make([]*types.T, 0, len(inTypes)+7)

	outTypes = append(outTypes, inTypes...)

	s.rankCol = len(outTypes)
	outTypes = append(outTypes, types.Int)

	s.sketchIdxCol = len(outTypes)
	outTypes = append(outTypes, types.Int)

	s.numRowsCol = len(outTypes)
	outTypes = append(outTypes, types.Int)

	s.numNullsCol = len(outTypes)
	outTypes = append(outTypes, types.Int)

	s.sizeCol = len(outTypes)
	outTypes = append(outTypes, types.Int)

	s.sketchCol = len(outTypes)
	outTypes = append(outTypes, types.Bytes)

	s.invColIdxCol = len(outTypes)
	outTypes = append(outTypes, types.Int)

	s.invIdxKeyCol = len(outTypes)
	outTypes = append(outTypes, types.Bytes)

	s.outTypes = outTypes

	if err := s.Init(
		nil, post, outTypes, flowCtx, processorID, output, memMonitor,
		execinfra.ProcStateOpts{
			TrailingMetaCallback: func() []execinfrapb.ProducerMetadata {
				__antithesis_instrumentation__.Notify(574681)
				s.close()
				return nil
			},
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(574682)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(574683)
	}
	__antithesis_instrumentation__.Notify(574673)
	return s, nil
}

func (s *samplerProcessor) pushTrailingMeta(ctx context.Context) {
	__antithesis_instrumentation__.Notify(574684)
	execinfra.SendTraceData(ctx, s.Output)
}

func (s *samplerProcessor) Run(ctx context.Context) {
	__antithesis_instrumentation__.Notify(574685)
	ctx = s.StartInternal(ctx, samplerProcName)
	s.input.Start(ctx)

	earlyExit, err := s.mainLoop(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(574687)
		execinfra.DrainAndClose(ctx, s.Output, err, s.pushTrailingMeta, s.input)
	} else {
		__antithesis_instrumentation__.Notify(574688)
		if !earlyExit {
			__antithesis_instrumentation__.Notify(574689)
			s.pushTrailingMeta(ctx)
			s.input.ConsumerClosed()
			s.Output.ProducerDone()
		} else {
			__antithesis_instrumentation__.Notify(574690)
		}
	}
	__antithesis_instrumentation__.Notify(574686)
	s.MoveToDraining(nil)
}

func (s *samplerProcessor) mainLoop(ctx context.Context) (earlyExit bool, err error) {
	__antithesis_instrumentation__.Notify(574691)
	rng, _ := randutil.NewPseudoRand()
	var da tree.DatumAlloc
	var buf []byte
	rowCount := 0
	lastWakeupTime := timeutil.Now()

	var invKeys [][]byte
	invRow := rowenc.EncDatumRow{rowenc.EncDatum{}}
	timer := timeutil.NewTimer()
	defer timer.Stop()

	for {
		__antithesis_instrumentation__.Notify(574702)
		row, meta := s.input.Next()
		if meta != nil {
			__antithesis_instrumentation__.Notify(574708)
			if !emitHelper(ctx, s.Output, &s.OutputHelper, nil, meta, s.pushTrailingMeta, s.input) {
				__antithesis_instrumentation__.Notify(574710)

				return true, nil
			} else {
				__antithesis_instrumentation__.Notify(574711)
			}
			__antithesis_instrumentation__.Notify(574709)
			continue
		} else {
			__antithesis_instrumentation__.Notify(574712)
		}
		__antithesis_instrumentation__.Notify(574703)
		if row == nil {
			__antithesis_instrumentation__.Notify(574713)
			break
		} else {
			__antithesis_instrumentation__.Notify(574714)
		}
		__antithesis_instrumentation__.Notify(574704)

		rowCount++
		if rowCount%SamplerProgressInterval == 0 {
			__antithesis_instrumentation__.Notify(574715)

			meta := &execinfrapb.ProducerMetadata{SamplerProgress: &execinfrapb.RemoteProducerMetadata_SamplerProgress{
				RowsProcessed: uint64(SamplerProgressInterval),
			}}
			if !emitHelper(ctx, s.Output, &s.OutputHelper, nil, meta, s.pushTrailingMeta, s.input) {
				__antithesis_instrumentation__.Notify(574717)
				return true, nil
			} else {
				__antithesis_instrumentation__.Notify(574718)
			}
			__antithesis_instrumentation__.Notify(574716)

			if s.maxFractionIdle > 0 {
				__antithesis_instrumentation__.Notify(574719)

				usage := s.flowCtx.Cfg.RuntimeStats.GetCPUCombinedPercentNorm()

				if usage > cpuUsageMinThrottle {
					__antithesis_instrumentation__.Notify(574721)
					fractionIdle := s.maxFractionIdle
					if usage < cpuUsageMaxThrottle {
						__antithesis_instrumentation__.Notify(574725)
						fractionIdle *= (usage - cpuUsageMinThrottle) /
							(cpuUsageMaxThrottle - cpuUsageMinThrottle)
					} else {
						__antithesis_instrumentation__.Notify(574726)
					}
					__antithesis_instrumentation__.Notify(574722)
					if log.V(1) {
						__antithesis_instrumentation__.Notify(574727)
						log.Infof(
							ctx, "throttling to fraction idle %.2f (based on usage %.2f)", fractionIdle, usage,
						)
					} else {
						__antithesis_instrumentation__.Notify(574728)
					}
					__antithesis_instrumentation__.Notify(574723)

					elapsed := timeutil.Since(lastWakeupTime)

					wait := time.Duration(float64(elapsed) * fractionIdle / (1 - fractionIdle))
					if wait > maxIdleSleepTime {
						__antithesis_instrumentation__.Notify(574729)
						wait = maxIdleSleepTime
					} else {
						__antithesis_instrumentation__.Notify(574730)
					}
					__antithesis_instrumentation__.Notify(574724)
					timer.Reset(wait)
					select {
					case <-timer.C:
						__antithesis_instrumentation__.Notify(574731)
						timer.Read = true
						break
					case <-s.flowCtx.Stopper().ShouldQuiesce():
						__antithesis_instrumentation__.Notify(574732)
						break
					}
				} else {
					__antithesis_instrumentation__.Notify(574733)
				}
				__antithesis_instrumentation__.Notify(574720)
				lastWakeupTime = timeutil.Now()
			} else {
				__antithesis_instrumentation__.Notify(574734)
			}
		} else {
			__antithesis_instrumentation__.Notify(574735)
		}
		__antithesis_instrumentation__.Notify(574705)

		for i := range s.sketches {
			__antithesis_instrumentation__.Notify(574736)
			if err := s.sketches[i].addRow(ctx, row, s.outTypes, &buf, &da); err != nil {
				__antithesis_instrumentation__.Notify(574737)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(574738)
			}
		}
		__antithesis_instrumentation__.Notify(574706)
		if earlyExit, err = s.sampleRow(ctx, &s.sr, row, rng); earlyExit || func() bool {
			__antithesis_instrumentation__.Notify(574739)
			return err != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(574740)
			return earlyExit, err
		} else {
			__antithesis_instrumentation__.Notify(574741)
		}
		__antithesis_instrumentation__.Notify(574707)

		for col, invSr := range s.invSr {
			__antithesis_instrumentation__.Notify(574742)
			if err := row[col].EnsureDecoded(s.outTypes[col], &da); err != nil {
				__antithesis_instrumentation__.Notify(574747)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(574748)
			}
			__antithesis_instrumentation__.Notify(574743)

			index := s.invSketch[col].spec.Index
			if index == nil {
				__antithesis_instrumentation__.Notify(574749)

				continue
			} else {
				__antithesis_instrumentation__.Notify(574750)
			}
			__antithesis_instrumentation__.Notify(574744)
			switch s.outTypes[col].Family() {
			case types.GeographyFamily, types.GeometryFamily:
				__antithesis_instrumentation__.Notify(574751)
				invKeys, err = rowenc.EncodeGeoInvertedIndexTableKeys(row[col].Datum, nil, index.GeoConfig)
			default:
				__antithesis_instrumentation__.Notify(574752)
				invKeys, err = rowenc.EncodeInvertedIndexTableKeys(row[col].Datum, nil, index.Version)
			}
			__antithesis_instrumentation__.Notify(574745)
			if err != nil {
				__antithesis_instrumentation__.Notify(574753)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(574754)
			}
			__antithesis_instrumentation__.Notify(574746)
			for _, key := range invKeys {
				__antithesis_instrumentation__.Notify(574755)
				invRow[0].Datum = da.NewDBytes(tree.DBytes(key))
				if err := s.invSketch[col].addRow(ctx, invRow, bytesRowType, &buf, &da); err != nil {
					__antithesis_instrumentation__.Notify(574757)
					return false, err
				} else {
					__antithesis_instrumentation__.Notify(574758)
				}
				__antithesis_instrumentation__.Notify(574756)
				if earlyExit, err = s.sampleRow(ctx, invSr, invRow, rng); earlyExit || func() bool {
					__antithesis_instrumentation__.Notify(574759)
					return err != nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(574760)
					return earlyExit, err
				} else {
					__antithesis_instrumentation__.Notify(574761)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(574692)

	outRow := make(rowenc.EncDatumRow, len(s.outTypes))

	for i := range outRow {
		__antithesis_instrumentation__.Notify(574762)
		outRow[i] = rowenc.DatumToEncDatum(s.outTypes[i], tree.DNull)
	}
	__antithesis_instrumentation__.Notify(574693)

	outRow[s.numRowsCol] = rowenc.EncDatum{Datum: tree.NewDInt(tree.DInt(s.sr.Cap()))}
	for _, sample := range s.sr.Get() {
		__antithesis_instrumentation__.Notify(574763)
		copy(outRow, sample.Row)
		outRow[s.rankCol] = rowenc.EncDatum{Datum: tree.NewDInt(tree.DInt(sample.Rank))}
		if !emitHelper(ctx, s.Output, &s.OutputHelper, outRow, nil, s.pushTrailingMeta, s.input) {
			__antithesis_instrumentation__.Notify(574764)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(574765)
		}
	}
	__antithesis_instrumentation__.Notify(574694)

	for i := range outRow {
		__antithesis_instrumentation__.Notify(574766)
		outRow[i] = rowenc.DatumToEncDatum(s.outTypes[i], tree.DNull)
	}
	__antithesis_instrumentation__.Notify(574695)
	for col, invSr := range s.invSr {
		__antithesis_instrumentation__.Notify(574767)

		outRow[s.numRowsCol] = rowenc.EncDatum{Datum: tree.NewDInt(tree.DInt(invSr.Cap()))}
		outRow[s.invColIdxCol] = rowenc.EncDatum{Datum: tree.NewDInt(tree.DInt(col))}
		for _, sample := range invSr.Get() {
			__antithesis_instrumentation__.Notify(574768)

			outRow[s.rankCol] = rowenc.EncDatum{Datum: tree.NewDInt(tree.DInt(sample.Rank))}
			outRow[s.invIdxKeyCol] = sample.Row[0]
			if !emitHelper(ctx, s.Output, &s.OutputHelper, outRow, nil, s.pushTrailingMeta, s.input) {
				__antithesis_instrumentation__.Notify(574769)
				return true, nil
			} else {
				__antithesis_instrumentation__.Notify(574770)
			}
		}
	}
	__antithesis_instrumentation__.Notify(574696)

	s.sr = stats.SampleReservoir{}
	s.invSr = nil

	for i := range outRow {
		__antithesis_instrumentation__.Notify(574771)
		outRow[i] = rowenc.DatumToEncDatum(s.outTypes[i], tree.DNull)
	}
	__antithesis_instrumentation__.Notify(574697)
	for i, si := range s.sketches {
		__antithesis_instrumentation__.Notify(574772)
		outRow[s.sketchIdxCol] = rowenc.EncDatum{Datum: tree.NewDInt(tree.DInt(i))}
		if earlyExit, err := s.emitSketchRow(ctx, &si, outRow); earlyExit || func() bool {
			__antithesis_instrumentation__.Notify(574773)
			return err != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(574774)
			return earlyExit, err
		} else {
			__antithesis_instrumentation__.Notify(574775)
		}
	}
	__antithesis_instrumentation__.Notify(574698)

	for i := range outRow {
		__antithesis_instrumentation__.Notify(574776)
		outRow[i] = rowenc.DatumToEncDatum(s.outTypes[i], tree.DNull)
	}
	__antithesis_instrumentation__.Notify(574699)
	for col, invSketch := range s.invSketch {
		__antithesis_instrumentation__.Notify(574777)
		outRow[s.invColIdxCol] = rowenc.EncDatum{Datum: tree.NewDInt(tree.DInt(col))}
		if earlyExit, err := s.emitSketchRow(ctx, invSketch, outRow); earlyExit || func() bool {
			__antithesis_instrumentation__.Notify(574778)
			return err != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(574779)
			return earlyExit, err
		} else {
			__antithesis_instrumentation__.Notify(574780)
		}
	}
	__antithesis_instrumentation__.Notify(574700)

	meta := &execinfrapb.ProducerMetadata{SamplerProgress: &execinfrapb.RemoteProducerMetadata_SamplerProgress{
		RowsProcessed: uint64(rowCount % SamplerProgressInterval),
	}}
	if !emitHelper(ctx, s.Output, &s.OutputHelper, nil, meta, s.pushTrailingMeta, s.input) {
		__antithesis_instrumentation__.Notify(574781)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(574782)
	}
	__antithesis_instrumentation__.Notify(574701)

	return false, nil
}

func (s *samplerProcessor) emitSketchRow(
	ctx context.Context, si *sketchInfo, outRow rowenc.EncDatumRow,
) (earlyExit bool, err error) {
	__antithesis_instrumentation__.Notify(574783)
	outRow[s.numRowsCol] = rowenc.EncDatum{Datum: tree.NewDInt(tree.DInt(si.numRows))}
	outRow[s.numNullsCol] = rowenc.EncDatum{Datum: tree.NewDInt(tree.DInt(si.numNulls))}
	outRow[s.sizeCol] = rowenc.EncDatum{Datum: tree.NewDInt(tree.DInt(si.size))}
	data, err := si.sketch.MarshalBinary()
	if err != nil {
		__antithesis_instrumentation__.Notify(574786)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(574787)
	}
	__antithesis_instrumentation__.Notify(574784)
	outRow[s.sketchCol] = rowenc.EncDatum{Datum: tree.NewDBytes(tree.DBytes(data))}
	if !emitHelper(ctx, s.Output, &s.OutputHelper, outRow, nil, s.pushTrailingMeta, s.input) {
		__antithesis_instrumentation__.Notify(574788)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(574789)
	}
	__antithesis_instrumentation__.Notify(574785)
	return false, nil
}

func (s *samplerProcessor) sampleRow(
	ctx context.Context, sr *stats.SampleReservoir, row rowenc.EncDatumRow, rng *rand.Rand,
) (earlyExit bool, err error) {
	__antithesis_instrumentation__.Notify(574790)

	rank := uint64(rng.Int63())
	prevCapacity := sr.Cap()
	if err := sr.SampleRow(ctx, s.EvalCtx, row, rank); err != nil {
		__antithesis_instrumentation__.Notify(574792)
		if !sqlerrors.IsOutOfMemoryError(err) {
			__antithesis_instrumentation__.Notify(574794)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(574795)
		}
		__antithesis_instrumentation__.Notify(574793)

		sr.Disable()
		log.Info(ctx, "disabling histogram collection due to excessive memory utilization")
		telemetry.Inc(sqltelemetry.StatsHistogramOOMCounter)

		meta := &execinfrapb.ProducerMetadata{SamplerProgress: &execinfrapb.RemoteProducerMetadata_SamplerProgress{
			HistogramDisabled: true,
		}}
		if !emitHelper(ctx, s.Output, &s.OutputHelper, nil, meta, s.pushTrailingMeta, s.input) {
			__antithesis_instrumentation__.Notify(574796)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(574797)
		}
	} else {
		__antithesis_instrumentation__.Notify(574798)
		if sr.Cap() != prevCapacity {
			__antithesis_instrumentation__.Notify(574799)
			log.Infof(
				ctx, "histogram samples reduced from %d to %d due to excessive memory utilization",
				prevCapacity, sr.Cap(),
			)
		} else {
			__antithesis_instrumentation__.Notify(574800)
		}
	}
	__antithesis_instrumentation__.Notify(574791)
	return false, nil
}

func (s *samplerProcessor) close() {
	__antithesis_instrumentation__.Notify(574801)
	if s.InternalClose() {
		__antithesis_instrumentation__.Notify(574802)
		s.memAcc.Close(s.Ctx)
		s.MemMonitor.Stop(s.Ctx)
	} else {
		__antithesis_instrumentation__.Notify(574803)
	}
}

var _ execinfra.DoesNotUseTxn = &samplerProcessor{}

func (s *samplerProcessor) DoesNotUseTxn() bool {
	__antithesis_instrumentation__.Notify(574804)
	txnUser, ok := s.input.(execinfra.DoesNotUseTxn)
	return ok && func() bool {
		__antithesis_instrumentation__.Notify(574805)
		return txnUser.DoesNotUseTxn() == true
	}() == true
}

func (s *sketchInfo) addRow(
	ctx context.Context, row rowenc.EncDatumRow, typs []*types.T, buf *[]byte, da *tree.DatumAlloc,
) error {
	__antithesis_instrumentation__.Notify(574806)
	var err error
	s.numRows++

	var col uint32
	var useFastPath bool
	if len(s.spec.Columns) == 1 {
		__antithesis_instrumentation__.Notify(574811)
		col = s.spec.Columns[0]
		isNull := row[col].IsNull()
		useFastPath = typs[col].Family() == types.IntFamily && func() bool {
			__antithesis_instrumentation__.Notify(574812)
			return !isNull == true
		}() == true
	} else {
		__antithesis_instrumentation__.Notify(574813)
	}
	__antithesis_instrumentation__.Notify(574807)

	if useFastPath {
		__antithesis_instrumentation__.Notify(574814)

		val, err := row[col].GetInt()
		if err != nil {
			__antithesis_instrumentation__.Notify(574817)
			return err
		} else {
			__antithesis_instrumentation__.Notify(574818)
		}
		__antithesis_instrumentation__.Notify(574815)

		if cap(*buf) < 8 {
			__antithesis_instrumentation__.Notify(574819)
			*buf = make([]byte, 8)
		} else {
			__antithesis_instrumentation__.Notify(574820)
			*buf = (*buf)[:8]
		}
		__antithesis_instrumentation__.Notify(574816)

		s.size += int64(row[col].DiskSize())

		binary.LittleEndian.PutUint64(*buf, uint64(val))
		s.sketch.Insert(*buf)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(574821)
	}
	__antithesis_instrumentation__.Notify(574808)
	isNull := true
	*buf = (*buf)[:0]
	for _, col := range s.spec.Columns {
		__antithesis_instrumentation__.Notify(574822)

		*buf, err = row[col].Fingerprint(ctx, typs[col], da, *buf, nil)
		if err != nil {
			__antithesis_instrumentation__.Notify(574824)
			return err
		} else {
			__antithesis_instrumentation__.Notify(574825)
		}
		__antithesis_instrumentation__.Notify(574823)
		isNull = isNull && func() bool {
			__antithesis_instrumentation__.Notify(574826)
			return row[col].IsNull() == true
		}() == true
		s.size += int64(row[col].DiskSize())
	}
	__antithesis_instrumentation__.Notify(574809)
	if isNull {
		__antithesis_instrumentation__.Notify(574827)
		s.numNulls++
	} else {
		__antithesis_instrumentation__.Notify(574828)
	}
	__antithesis_instrumentation__.Notify(574810)
	s.sketch.Insert(*buf)
	return nil
}
