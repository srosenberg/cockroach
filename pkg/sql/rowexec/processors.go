package rowexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/backfill"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

func emitHelper(
	ctx context.Context,
	output execinfra.RowReceiver,
	procOutputHelper *execinfra.ProcOutputHelper,
	row rowenc.EncDatumRow,
	meta *execinfrapb.ProducerMetadata,
	pushTrailingMeta func(context.Context),
	inputs ...execinfra.RowSource,
) bool {
	__antithesis_instrumentation__.Notify(574075)
	if output == nil {
		__antithesis_instrumentation__.Notify(574078)
		panic("output RowReceiver is not set for emitting")
	} else {
		__antithesis_instrumentation__.Notify(574079)
	}
	__antithesis_instrumentation__.Notify(574076)
	var consumerStatus execinfra.ConsumerStatus
	if meta != nil {
		__antithesis_instrumentation__.Notify(574080)
		if row != nil {
			__antithesis_instrumentation__.Notify(574082)
			panic("both row data and metadata in the same emitHelper call")
		} else {
			__antithesis_instrumentation__.Notify(574083)
		}
		__antithesis_instrumentation__.Notify(574081)

		foundErr := meta.Err != nil
		consumerStatus = output.Push(nil, meta)
		if foundErr {
			__antithesis_instrumentation__.Notify(574084)
			consumerStatus = execinfra.ConsumerClosed
		} else {
			__antithesis_instrumentation__.Notify(574085)
		}
	} else {
		__antithesis_instrumentation__.Notify(574086)
		var err error
		consumerStatus, err = procOutputHelper.EmitRow(ctx, row, output)
		if err != nil {
			__antithesis_instrumentation__.Notify(574087)
			output.Push(nil, &execinfrapb.ProducerMetadata{Err: err})
			consumerStatus = execinfra.ConsumerClosed
		} else {
			__antithesis_instrumentation__.Notify(574088)
		}
	}
	__antithesis_instrumentation__.Notify(574077)
	switch consumerStatus {
	case execinfra.NeedMoreRows:
		__antithesis_instrumentation__.Notify(574089)
		return true
	case execinfra.DrainRequested:
		__antithesis_instrumentation__.Notify(574090)
		log.VEventf(ctx, 1, "no more rows required. drain requested.")
		execinfra.DrainAndClose(ctx, output, nil, pushTrailingMeta, inputs...)
		return false
	case execinfra.ConsumerClosed:
		__antithesis_instrumentation__.Notify(574091)
		log.VEventf(ctx, 1, "no more rows required. Consumer shut down.")
		for _, input := range inputs {
			__antithesis_instrumentation__.Notify(574094)
			input.ConsumerClosed()
		}
		__antithesis_instrumentation__.Notify(574092)
		output.ProducerDone()
		return false
	default:
		__antithesis_instrumentation__.Notify(574093)
		log.Fatalf(ctx, "unexpected consumerStatus: %d", consumerStatus)
		return false
	}
}

func checkNumInOut(
	inputs []execinfra.RowSource, outputs []execinfra.RowReceiver, numIn, numOut int,
) error {
	__antithesis_instrumentation__.Notify(574095)
	if len(inputs) != numIn {
		__antithesis_instrumentation__.Notify(574098)
		return errors.Errorf("expected %d input(s), got %d", numIn, len(inputs))
	} else {
		__antithesis_instrumentation__.Notify(574099)
	}
	__antithesis_instrumentation__.Notify(574096)
	if len(outputs) != numOut {
		__antithesis_instrumentation__.Notify(574100)
		return errors.Errorf("expected %d output(s), got %d", numOut, len(outputs))
	} else {
		__antithesis_instrumentation__.Notify(574101)
	}
	__antithesis_instrumentation__.Notify(574097)
	return nil
}

func NewProcessor(
	ctx context.Context,
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	core *execinfrapb.ProcessorCoreUnion,
	post *execinfrapb.PostProcessSpec,
	inputs []execinfra.RowSource,
	outputs []execinfra.RowReceiver,
	localProcessors []execinfra.LocalProcessor,
) (execinfra.Processor, error) {
	__antithesis_instrumentation__.Notify(574102)
	if core.Noop != nil {
		__antithesis_instrumentation__.Notify(574136)
		if err := checkNumInOut(inputs, outputs, 1, 1); err != nil {
			__antithesis_instrumentation__.Notify(574138)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(574139)
		}
		__antithesis_instrumentation__.Notify(574137)
		return newNoopProcessor(flowCtx, processorID, inputs[0], post, outputs[0])
	} else {
		__antithesis_instrumentation__.Notify(574140)
	}
	__antithesis_instrumentation__.Notify(574103)
	if core.Values != nil {
		__antithesis_instrumentation__.Notify(574141)
		if err := checkNumInOut(inputs, outputs, 0, 1); err != nil {
			__antithesis_instrumentation__.Notify(574143)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(574144)
		}
		__antithesis_instrumentation__.Notify(574142)
		return newValuesProcessor(flowCtx, processorID, core.Values, post, outputs[0])
	} else {
		__antithesis_instrumentation__.Notify(574145)
	}
	__antithesis_instrumentation__.Notify(574104)
	if core.TableReader != nil {
		__antithesis_instrumentation__.Notify(574146)
		if err := checkNumInOut(inputs, outputs, 0, 1); err != nil {
			__antithesis_instrumentation__.Notify(574148)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(574149)
		}
		__antithesis_instrumentation__.Notify(574147)
		return newTableReader(flowCtx, processorID, core.TableReader, post, outputs[0])
	} else {
		__antithesis_instrumentation__.Notify(574150)
	}
	__antithesis_instrumentation__.Notify(574105)
	if core.Filterer != nil {
		__antithesis_instrumentation__.Notify(574151)
		if err := checkNumInOut(inputs, outputs, 1, 1); err != nil {
			__antithesis_instrumentation__.Notify(574153)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(574154)
		}
		__antithesis_instrumentation__.Notify(574152)
		return newFiltererProcessor(flowCtx, processorID, core.Filterer, inputs[0], post, outputs[0])
	} else {
		__antithesis_instrumentation__.Notify(574155)
	}
	__antithesis_instrumentation__.Notify(574106)
	if core.JoinReader != nil {
		__antithesis_instrumentation__.Notify(574156)
		if err := checkNumInOut(inputs, outputs, 1, 1); err != nil {
			__antithesis_instrumentation__.Notify(574159)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(574160)
		}
		__antithesis_instrumentation__.Notify(574157)
		if core.JoinReader.IsIndexJoin() {
			__antithesis_instrumentation__.Notify(574161)
			return newJoinReader(
				flowCtx, processorID, core.JoinReader, inputs[0], post, outputs[0], indexJoinReaderType)
		} else {
			__antithesis_instrumentation__.Notify(574162)
		}
		__antithesis_instrumentation__.Notify(574158)
		return newJoinReader(flowCtx, processorID, core.JoinReader, inputs[0], post, outputs[0], lookupJoinReaderType)
	} else {
		__antithesis_instrumentation__.Notify(574163)
	}
	__antithesis_instrumentation__.Notify(574107)
	if core.Sorter != nil {
		__antithesis_instrumentation__.Notify(574164)
		if err := checkNumInOut(inputs, outputs, 1, 1); err != nil {
			__antithesis_instrumentation__.Notify(574166)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(574167)
		}
		__antithesis_instrumentation__.Notify(574165)
		return newSorter(ctx, flowCtx, processorID, core.Sorter, inputs[0], post, outputs[0])
	} else {
		__antithesis_instrumentation__.Notify(574168)
	}
	__antithesis_instrumentation__.Notify(574108)
	if core.Distinct != nil {
		__antithesis_instrumentation__.Notify(574169)
		if err := checkNumInOut(inputs, outputs, 1, 1); err != nil {
			__antithesis_instrumentation__.Notify(574171)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(574172)
		}
		__antithesis_instrumentation__.Notify(574170)
		return newDistinct(flowCtx, processorID, core.Distinct, inputs[0], post, outputs[0])
	} else {
		__antithesis_instrumentation__.Notify(574173)
	}
	__antithesis_instrumentation__.Notify(574109)
	if core.Ordinality != nil {
		__antithesis_instrumentation__.Notify(574174)
		if err := checkNumInOut(inputs, outputs, 1, 1); err != nil {
			__antithesis_instrumentation__.Notify(574176)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(574177)
		}
		__antithesis_instrumentation__.Notify(574175)
		return newOrdinalityProcessor(flowCtx, processorID, core.Ordinality, inputs[0], post, outputs[0])
	} else {
		__antithesis_instrumentation__.Notify(574178)
	}
	__antithesis_instrumentation__.Notify(574110)
	if core.Aggregator != nil {
		__antithesis_instrumentation__.Notify(574179)
		if err := checkNumInOut(inputs, outputs, 1, 1); err != nil {
			__antithesis_instrumentation__.Notify(574181)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(574182)
		}
		__antithesis_instrumentation__.Notify(574180)
		return newAggregator(flowCtx, processorID, core.Aggregator, inputs[0], post, outputs[0])
	} else {
		__antithesis_instrumentation__.Notify(574183)
	}
	__antithesis_instrumentation__.Notify(574111)
	if core.MergeJoiner != nil {
		__antithesis_instrumentation__.Notify(574184)
		if err := checkNumInOut(inputs, outputs, 2, 1); err != nil {
			__antithesis_instrumentation__.Notify(574186)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(574187)
		}
		__antithesis_instrumentation__.Notify(574185)
		return newMergeJoiner(
			flowCtx, processorID, core.MergeJoiner, inputs[0], inputs[1], post, outputs[0],
		)
	} else {
		__antithesis_instrumentation__.Notify(574188)
	}
	__antithesis_instrumentation__.Notify(574112)
	if core.ZigzagJoiner != nil {
		__antithesis_instrumentation__.Notify(574189)
		if err := checkNumInOut(inputs, outputs, 0, 1); err != nil {
			__antithesis_instrumentation__.Notify(574191)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(574192)
		}
		__antithesis_instrumentation__.Notify(574190)
		return newZigzagJoiner(
			flowCtx, processorID, core.ZigzagJoiner, nil, post, outputs[0],
		)
	} else {
		__antithesis_instrumentation__.Notify(574193)
	}
	__antithesis_instrumentation__.Notify(574113)
	if core.HashJoiner != nil {
		__antithesis_instrumentation__.Notify(574194)
		if err := checkNumInOut(inputs, outputs, 2, 1); err != nil {
			__antithesis_instrumentation__.Notify(574196)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(574197)
		}
		__antithesis_instrumentation__.Notify(574195)
		return newHashJoiner(
			flowCtx, processorID, core.HashJoiner, inputs[0], inputs[1], post, outputs[0],
		)
	} else {
		__antithesis_instrumentation__.Notify(574198)
	}
	__antithesis_instrumentation__.Notify(574114)
	if core.InvertedJoiner != nil {
		__antithesis_instrumentation__.Notify(574199)
		if err := checkNumInOut(inputs, outputs, 1, 1); err != nil {
			__antithesis_instrumentation__.Notify(574201)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(574202)
		}
		__antithesis_instrumentation__.Notify(574200)
		return newInvertedJoiner(
			flowCtx, processorID, core.InvertedJoiner, nil, inputs[0], post, outputs[0])
	} else {
		__antithesis_instrumentation__.Notify(574203)
	}
	__antithesis_instrumentation__.Notify(574115)
	if core.Backfiller != nil {
		__antithesis_instrumentation__.Notify(574204)
		if err := checkNumInOut(inputs, outputs, 0, 1); err != nil {
			__antithesis_instrumentation__.Notify(574206)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(574207)
		}
		__antithesis_instrumentation__.Notify(574205)
		switch core.Backfiller.Type {
		case execinfrapb.BackfillerSpec_Index:
			__antithesis_instrumentation__.Notify(574208)
			return newIndexBackfiller(ctx, flowCtx, processorID, *core.Backfiller, post, outputs[0])
		case execinfrapb.BackfillerSpec_Column:
			__antithesis_instrumentation__.Notify(574209)
			return newColumnBackfiller(ctx, flowCtx, processorID, *core.Backfiller, post, outputs[0])
		default:
			__antithesis_instrumentation__.Notify(574210)
		}
	} else {
		__antithesis_instrumentation__.Notify(574211)
	}
	__antithesis_instrumentation__.Notify(574116)
	if core.Sampler != nil {
		__antithesis_instrumentation__.Notify(574212)
		if err := checkNumInOut(inputs, outputs, 1, 1); err != nil {
			__antithesis_instrumentation__.Notify(574214)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(574215)
		}
		__antithesis_instrumentation__.Notify(574213)
		return newSamplerProcessor(flowCtx, processorID, core.Sampler, inputs[0], post, outputs[0])
	} else {
		__antithesis_instrumentation__.Notify(574216)
	}
	__antithesis_instrumentation__.Notify(574117)
	if core.SampleAggregator != nil {
		__antithesis_instrumentation__.Notify(574217)
		if err := checkNumInOut(inputs, outputs, 1, 1); err != nil {
			__antithesis_instrumentation__.Notify(574219)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(574220)
		}
		__antithesis_instrumentation__.Notify(574218)
		return newSampleAggregator(flowCtx, processorID, core.SampleAggregator, inputs[0], post, outputs[0])
	} else {
		__antithesis_instrumentation__.Notify(574221)
	}
	__antithesis_instrumentation__.Notify(574118)
	if core.ReadImport != nil {
		__antithesis_instrumentation__.Notify(574222)
		if err := checkNumInOut(inputs, outputs, 0, 1); err != nil {
			__antithesis_instrumentation__.Notify(574225)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(574226)
		}
		__antithesis_instrumentation__.Notify(574223)
		if NewReadImportDataProcessor == nil {
			__antithesis_instrumentation__.Notify(574227)
			return nil, errors.New("ReadImportData processor unimplemented")
		} else {
			__antithesis_instrumentation__.Notify(574228)
		}
		__antithesis_instrumentation__.Notify(574224)
		return NewReadImportDataProcessor(flowCtx, processorID, *core.ReadImport, post, outputs[0])
	} else {
		__antithesis_instrumentation__.Notify(574229)
	}
	__antithesis_instrumentation__.Notify(574119)
	if core.BackupData != nil {
		__antithesis_instrumentation__.Notify(574230)
		if err := checkNumInOut(inputs, outputs, 0, 1); err != nil {
			__antithesis_instrumentation__.Notify(574233)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(574234)
		}
		__antithesis_instrumentation__.Notify(574231)
		if NewBackupDataProcessor == nil {
			__antithesis_instrumentation__.Notify(574235)
			return nil, errors.New("BackupData processor unimplemented")
		} else {
			__antithesis_instrumentation__.Notify(574236)
		}
		__antithesis_instrumentation__.Notify(574232)
		return NewBackupDataProcessor(flowCtx, processorID, *core.BackupData, post, outputs[0])
	} else {
		__antithesis_instrumentation__.Notify(574237)
	}
	__antithesis_instrumentation__.Notify(574120)
	if core.SplitAndScatter != nil {
		__antithesis_instrumentation__.Notify(574238)
		if err := checkNumInOut(inputs, outputs, 0, 1); err != nil {
			__antithesis_instrumentation__.Notify(574241)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(574242)
		}
		__antithesis_instrumentation__.Notify(574239)
		if NewSplitAndScatterProcessor == nil {
			__antithesis_instrumentation__.Notify(574243)
			return nil, errors.New("SplitAndScatter processor unimplemented")
		} else {
			__antithesis_instrumentation__.Notify(574244)
		}
		__antithesis_instrumentation__.Notify(574240)
		return NewSplitAndScatterProcessor(flowCtx, processorID, *core.SplitAndScatter, post, outputs[0])
	} else {
		__antithesis_instrumentation__.Notify(574245)
	}
	__antithesis_instrumentation__.Notify(574121)
	if core.RestoreData != nil {
		__antithesis_instrumentation__.Notify(574246)
		if err := checkNumInOut(inputs, outputs, 1, 1); err != nil {
			__antithesis_instrumentation__.Notify(574249)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(574250)
		}
		__antithesis_instrumentation__.Notify(574247)
		if NewRestoreDataProcessor == nil {
			__antithesis_instrumentation__.Notify(574251)
			return nil, errors.New("RestoreData processor unimplemented")
		} else {
			__antithesis_instrumentation__.Notify(574252)
		}
		__antithesis_instrumentation__.Notify(574248)
		return NewRestoreDataProcessor(flowCtx, processorID, *core.RestoreData, post, inputs[0], outputs[0])
	} else {
		__antithesis_instrumentation__.Notify(574253)
	}
	__antithesis_instrumentation__.Notify(574122)
	if core.StreamIngestionData != nil {
		__antithesis_instrumentation__.Notify(574254)
		if err := checkNumInOut(inputs, outputs, 0, 1); err != nil {
			__antithesis_instrumentation__.Notify(574257)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(574258)
		}
		__antithesis_instrumentation__.Notify(574255)
		if NewStreamIngestionDataProcessor == nil {
			__antithesis_instrumentation__.Notify(574259)
			return nil, errors.New("StreamIngestionData processor unimplemented")
		} else {
			__antithesis_instrumentation__.Notify(574260)
		}
		__antithesis_instrumentation__.Notify(574256)
		return NewStreamIngestionDataProcessor(flowCtx, processorID, *core.StreamIngestionData, post, outputs[0])
	} else {
		__antithesis_instrumentation__.Notify(574261)
	}
	__antithesis_instrumentation__.Notify(574123)
	if core.Exporter != nil {
		__antithesis_instrumentation__.Notify(574262)
		if err := checkNumInOut(inputs, outputs, 1, 1); err != nil {
			__antithesis_instrumentation__.Notify(574265)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(574266)
		}
		__antithesis_instrumentation__.Notify(574263)

		if core.Exporter.Format.Format == roachpb.IOFileFormat_Parquet {
			__antithesis_instrumentation__.Notify(574267)
			return NewParquetWriterProcessor(flowCtx, processorID, *core.Exporter, inputs[0], outputs[0])
		} else {
			__antithesis_instrumentation__.Notify(574268)
		}
		__antithesis_instrumentation__.Notify(574264)
		return NewCSVWriterProcessor(flowCtx, processorID, *core.Exporter, inputs[0], outputs[0])
	} else {
		__antithesis_instrumentation__.Notify(574269)
	}
	__antithesis_instrumentation__.Notify(574124)

	if core.BulkRowWriter != nil {
		__antithesis_instrumentation__.Notify(574270)
		if err := checkNumInOut(inputs, outputs, 1, 1); err != nil {
			__antithesis_instrumentation__.Notify(574272)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(574273)
		}
		__antithesis_instrumentation__.Notify(574271)
		return newBulkRowWriterProcessor(flowCtx, processorID, *core.BulkRowWriter, inputs[0], outputs[0])
	} else {
		__antithesis_instrumentation__.Notify(574274)
	}
	__antithesis_instrumentation__.Notify(574125)
	if core.MetadataTestSender != nil {
		__antithesis_instrumentation__.Notify(574275)
		if err := checkNumInOut(inputs, outputs, 1, 1); err != nil {
			__antithesis_instrumentation__.Notify(574277)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(574278)
		}
		__antithesis_instrumentation__.Notify(574276)
		return execinfra.NewMetadataTestSender(flowCtx, processorID, inputs[0], post, outputs[0], core.MetadataTestSender.ID)
	} else {
		__antithesis_instrumentation__.Notify(574279)
	}
	__antithesis_instrumentation__.Notify(574126)
	if core.MetadataTestReceiver != nil {
		__antithesis_instrumentation__.Notify(574280)
		if err := checkNumInOut(inputs, outputs, 1, 1); err != nil {
			__antithesis_instrumentation__.Notify(574282)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(574283)
		}
		__antithesis_instrumentation__.Notify(574281)
		return execinfra.NewMetadataTestReceiver(
			flowCtx, processorID, inputs[0], post, outputs[0], core.MetadataTestReceiver.SenderIDs,
		)
	} else {
		__antithesis_instrumentation__.Notify(574284)
	}
	__antithesis_instrumentation__.Notify(574127)
	if core.ProjectSet != nil {
		__antithesis_instrumentation__.Notify(574285)
		if err := checkNumInOut(inputs, outputs, 1, 1); err != nil {
			__antithesis_instrumentation__.Notify(574287)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(574288)
		}
		__antithesis_instrumentation__.Notify(574286)
		return newProjectSetProcessor(flowCtx, processorID, core.ProjectSet, inputs[0], post, outputs[0])
	} else {
		__antithesis_instrumentation__.Notify(574289)
	}
	__antithesis_instrumentation__.Notify(574128)
	if core.Windower != nil {
		__antithesis_instrumentation__.Notify(574290)
		if err := checkNumInOut(inputs, outputs, 1, 1); err != nil {
			__antithesis_instrumentation__.Notify(574292)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(574293)
		}
		__antithesis_instrumentation__.Notify(574291)
		return newWindower(flowCtx, processorID, core.Windower, inputs[0], post, outputs[0])
	} else {
		__antithesis_instrumentation__.Notify(574294)
	}
	__antithesis_instrumentation__.Notify(574129)
	if core.LocalPlanNode != nil {
		__antithesis_instrumentation__.Notify(574295)
		numInputs := int(core.LocalPlanNode.NumInputs)
		if err := checkNumInOut(inputs, outputs, numInputs, 1); err != nil {
			__antithesis_instrumentation__.Notify(574299)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(574300)
		}
		__antithesis_instrumentation__.Notify(574296)
		processor := localProcessors[core.LocalPlanNode.RowSourceIdx]
		if err := processor.InitWithOutput(flowCtx, post, outputs[0]); err != nil {
			__antithesis_instrumentation__.Notify(574301)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(574302)
		}
		__antithesis_instrumentation__.Notify(574297)
		if numInputs == 1 {
			__antithesis_instrumentation__.Notify(574303)
			if err := processor.SetInput(ctx, inputs[0]); err != nil {
				__antithesis_instrumentation__.Notify(574304)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(574305)
			}
		} else {
			__antithesis_instrumentation__.Notify(574306)
			if numInputs > 1 {
				__antithesis_instrumentation__.Notify(574307)
				return nil, errors.Errorf("invalid localPlanNode core with multiple inputs %+v", core.LocalPlanNode)
			} else {
				__antithesis_instrumentation__.Notify(574308)
			}
		}
		__antithesis_instrumentation__.Notify(574298)
		return processor, nil
	} else {
		__antithesis_instrumentation__.Notify(574309)
	}
	__antithesis_instrumentation__.Notify(574130)
	if core.ChangeAggregator != nil {
		__antithesis_instrumentation__.Notify(574310)
		if err := checkNumInOut(inputs, outputs, 0, 1); err != nil {
			__antithesis_instrumentation__.Notify(574313)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(574314)
		}
		__antithesis_instrumentation__.Notify(574311)
		if NewChangeAggregatorProcessor == nil {
			__antithesis_instrumentation__.Notify(574315)
			return nil, errors.New("ChangeAggregator processor unimplemented")
		} else {
			__antithesis_instrumentation__.Notify(574316)
		}
		__antithesis_instrumentation__.Notify(574312)
		return NewChangeAggregatorProcessor(flowCtx, processorID, *core.ChangeAggregator, post, outputs[0])
	} else {
		__antithesis_instrumentation__.Notify(574317)
	}
	__antithesis_instrumentation__.Notify(574131)
	if core.ChangeFrontier != nil {
		__antithesis_instrumentation__.Notify(574318)
		if err := checkNumInOut(inputs, outputs, 1, 1); err != nil {
			__antithesis_instrumentation__.Notify(574321)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(574322)
		}
		__antithesis_instrumentation__.Notify(574319)
		if NewChangeFrontierProcessor == nil {
			__antithesis_instrumentation__.Notify(574323)
			return nil, errors.New("ChangeFrontier processor unimplemented")
		} else {
			__antithesis_instrumentation__.Notify(574324)
		}
		__antithesis_instrumentation__.Notify(574320)
		return NewChangeFrontierProcessor(flowCtx, processorID, *core.ChangeFrontier, inputs[0], post, outputs[0])
	} else {
		__antithesis_instrumentation__.Notify(574325)
	}
	__antithesis_instrumentation__.Notify(574132)
	if core.InvertedFilterer != nil {
		__antithesis_instrumentation__.Notify(574326)
		if err := checkNumInOut(inputs, outputs, 1, 1); err != nil {
			__antithesis_instrumentation__.Notify(574328)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(574329)
		}
		__antithesis_instrumentation__.Notify(574327)
		return newInvertedFilterer(flowCtx, processorID, core.InvertedFilterer, inputs[0], post, outputs[0])
	} else {
		__antithesis_instrumentation__.Notify(574330)
	}
	__antithesis_instrumentation__.Notify(574133)
	if core.StreamIngestionFrontier != nil {
		__antithesis_instrumentation__.Notify(574331)
		if err := checkNumInOut(inputs, outputs, 1, 1); err != nil {
			__antithesis_instrumentation__.Notify(574333)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(574334)
		}
		__antithesis_instrumentation__.Notify(574332)
		return NewStreamIngestionFrontierProcessor(flowCtx, processorID, *core.StreamIngestionFrontier, inputs[0], post, outputs[0])
	} else {
		__antithesis_instrumentation__.Notify(574335)
	}
	__antithesis_instrumentation__.Notify(574134)
	if core.IndexBackfillMerger != nil {
		__antithesis_instrumentation__.Notify(574336)
		if err := checkNumInOut(inputs, outputs, 0, 1); err != nil {
			__antithesis_instrumentation__.Notify(574338)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(574339)
		}
		__antithesis_instrumentation__.Notify(574337)
		return backfill.NewIndexBackfillMerger(ctx, flowCtx, *core.IndexBackfillMerger, outputs[0])
	} else {
		__antithesis_instrumentation__.Notify(574340)
	}
	__antithesis_instrumentation__.Notify(574135)
	return nil, errors.Errorf("unsupported processor core %q", core)
}

var NewReadImportDataProcessor func(*execinfra.FlowCtx, int32, execinfrapb.ReadImportDataSpec, *execinfrapb.PostProcessSpec, execinfra.RowReceiver) (execinfra.Processor, error)

var NewBackupDataProcessor func(*execinfra.FlowCtx, int32, execinfrapb.BackupDataSpec, *execinfrapb.PostProcessSpec, execinfra.RowReceiver) (execinfra.Processor, error)

var NewSplitAndScatterProcessor func(*execinfra.FlowCtx, int32, execinfrapb.SplitAndScatterSpec, *execinfrapb.PostProcessSpec, execinfra.RowReceiver) (execinfra.Processor, error)

var NewRestoreDataProcessor func(*execinfra.FlowCtx, int32, execinfrapb.RestoreDataSpec, *execinfrapb.PostProcessSpec, execinfra.RowSource, execinfra.RowReceiver) (execinfra.Processor, error)

var NewStreamIngestionDataProcessor func(*execinfra.FlowCtx, int32, execinfrapb.StreamIngestionDataSpec, *execinfrapb.PostProcessSpec, execinfra.RowReceiver) (execinfra.Processor, error)

var NewCSVWriterProcessor func(*execinfra.FlowCtx, int32, execinfrapb.ExportSpec, execinfra.RowSource, execinfra.RowReceiver) (execinfra.Processor, error)

var NewParquetWriterProcessor func(*execinfra.FlowCtx, int32, execinfrapb.ExportSpec, execinfra.RowSource, execinfra.RowReceiver) (execinfra.Processor, error)

var NewChangeAggregatorProcessor func(*execinfra.FlowCtx, int32, execinfrapb.ChangeAggregatorSpec, *execinfrapb.PostProcessSpec, execinfra.RowReceiver) (execinfra.Processor, error)

var NewChangeFrontierProcessor func(*execinfra.FlowCtx, int32, execinfrapb.ChangeFrontierSpec, execinfra.RowSource, *execinfrapb.PostProcessSpec, execinfra.RowReceiver) (execinfra.Processor, error)

var NewStreamIngestionFrontierProcessor func(*execinfra.FlowCtx, int32, execinfrapb.StreamIngestionFrontierSpec, execinfra.RowSource, *execinfrapb.PostProcessSpec, execinfra.RowReceiver) (execinfra.Processor, error)
