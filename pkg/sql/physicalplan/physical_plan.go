package physicalplan

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

type Processor struct {
	SQLInstanceID base.SQLInstanceID

	Spec execinfrapb.ProcessorSpec
}

type ProcessorIdx int

type Stream struct {
	SourceProcessor ProcessorIdx

	SourceRouterSlot int

	DestProcessor ProcessorIdx

	DestInput int
}

type PhysicalInfrastructure struct {
	FlowID               uuid.UUID
	GatewaySQLInstanceID base.SQLInstanceID

	Processors []Processor

	LocalProcessors []execinfra.LocalProcessor

	Streams []Stream

	stageCounter int32
}

func (p *PhysicalInfrastructure) AddProcessor(proc Processor) ProcessorIdx {
	__antithesis_instrumentation__.Notify(562170)
	idx := ProcessorIdx(len(p.Processors))
	p.Processors = append(p.Processors, proc)
	return idx
}

func (p *PhysicalInfrastructure) AddLocalProcessor(proc execinfra.LocalProcessor) int {
	__antithesis_instrumentation__.Notify(562171)
	idx := len(p.LocalProcessors)
	p.LocalProcessors = append(p.LocalProcessors, proc)
	return idx
}

type PhysicalPlan struct {
	*PhysicalInfrastructure

	ResultRouters []ProcessorIdx

	ResultColumns colinfo.ResultColumns

	MergeOrdering execinfrapb.Ordering

	TotalEstimatedScannedRows uint64

	Distribution PlanDistribution
}

func MakePhysicalInfrastructure(
	flowID uuid.UUID, gatewaySQLInstanceID base.SQLInstanceID,
) PhysicalInfrastructure {
	__antithesis_instrumentation__.Notify(562172)
	return PhysicalInfrastructure{
		FlowID:               flowID,
		GatewaySQLInstanceID: gatewaySQLInstanceID,
	}
}

func MakePhysicalPlan(infra *PhysicalInfrastructure) PhysicalPlan {
	__antithesis_instrumentation__.Notify(562173)
	return PhysicalPlan{
		PhysicalInfrastructure: infra,
	}
}

func (p *PhysicalPlan) GetResultTypes() []*types.T {
	__antithesis_instrumentation__.Notify(562174)
	if len(p.ResultRouters) == 0 {
		__antithesis_instrumentation__.Notify(562176)
		panic(errors.AssertionFailedf("unexpectedly no result routers in %v", *p))
	} else {
		__antithesis_instrumentation__.Notify(562177)
	}
	__antithesis_instrumentation__.Notify(562175)
	return p.Processors[p.ResultRouters[0]].Spec.ResultTypes
}

func (p *PhysicalPlan) NewStage(containsRemoteProcessor bool, allowPartialDistribution bool) int32 {
	__antithesis_instrumentation__.Notify(562178)
	newStageDistribution := LocalPlan
	if containsRemoteProcessor {
		__antithesis_instrumentation__.Notify(562181)
		newStageDistribution = FullyDistributedPlan
	} else {
		__antithesis_instrumentation__.Notify(562182)
	}
	__antithesis_instrumentation__.Notify(562179)
	if p.stageCounter == 0 {
		__antithesis_instrumentation__.Notify(562183)

		p.Distribution = newStageDistribution
	} else {
		__antithesis_instrumentation__.Notify(562184)
		p.Distribution = p.Distribution.compose(newStageDistribution, allowPartialDistribution)
	}
	__antithesis_instrumentation__.Notify(562180)
	p.stageCounter++
	return p.stageCounter
}

func (p *PhysicalPlan) NewStageOnNodes(sqlInstanceIDs []base.SQLInstanceID) int32 {
	__antithesis_instrumentation__.Notify(562185)

	return p.NewStage(
		len(sqlInstanceIDs) > 1 || func() bool {
			__antithesis_instrumentation__.Notify(562186)
			return sqlInstanceIDs[0] != p.GatewaySQLInstanceID == true
		}() == true,
		false,
	)
}

func (p *PhysicalPlan) SetMergeOrdering(o execinfrapb.Ordering) {
	__antithesis_instrumentation__.Notify(562187)
	if len(p.ResultRouters) > 1 {
		__antithesis_instrumentation__.Notify(562188)
		p.MergeOrdering = o
	} else {
		__antithesis_instrumentation__.Notify(562189)
		p.MergeOrdering = execinfrapb.Ordering{}
	}
}

type ProcessorCorePlacement struct {
	SQLInstanceID base.SQLInstanceID
	Core          execinfrapb.ProcessorCoreUnion

	EstimatedRowCount uint64
}

func (p *PhysicalPlan) AddNoInputStage(
	corePlacements []ProcessorCorePlacement,
	post execinfrapb.PostProcessSpec,
	outputTypes []*types.T,
	newOrdering execinfrapb.Ordering,
) {
	__antithesis_instrumentation__.Notify(562190)

	containsRemoteProcessor := false
	for i := range corePlacements {
		__antithesis_instrumentation__.Notify(562193)
		if corePlacements[i].SQLInstanceID != p.GatewaySQLInstanceID {
			__antithesis_instrumentation__.Notify(562194)
			containsRemoteProcessor = true
			break
		} else {
			__antithesis_instrumentation__.Notify(562195)
		}
	}
	__antithesis_instrumentation__.Notify(562191)
	stageID := p.NewStage(containsRemoteProcessor, false)
	p.ResultRouters = make([]ProcessorIdx, len(corePlacements))
	for i := range p.ResultRouters {
		__antithesis_instrumentation__.Notify(562196)
		proc := Processor{
			SQLInstanceID: corePlacements[i].SQLInstanceID,
			Spec: execinfrapb.ProcessorSpec{
				Core: corePlacements[i].Core,
				Post: post,
				Output: []execinfrapb.OutputRouterSpec{{
					Type: execinfrapb.OutputRouterSpec_PASS_THROUGH,
				}},
				StageID:           stageID,
				ResultTypes:       outputTypes,
				EstimatedRowCount: corePlacements[i].EstimatedRowCount,
			},
		}

		pIdx := p.AddProcessor(proc)
		p.ResultRouters[i] = pIdx
	}
	__antithesis_instrumentation__.Notify(562192)
	p.SetMergeOrdering(newOrdering)
}

func (p *PhysicalPlan) AddNoGroupingStage(
	core execinfrapb.ProcessorCoreUnion,
	post execinfrapb.PostProcessSpec,
	outputTypes []*types.T,
	newOrdering execinfrapb.Ordering,
) {
	__antithesis_instrumentation__.Notify(562197)
	p.AddNoGroupingStageWithCoreFunc(
		func(_ int, _ *Processor) execinfrapb.ProcessorCoreUnion {
			__antithesis_instrumentation__.Notify(562198)
			return core
		},
		post,
		outputTypes,
		newOrdering,
	)
}

func (p *PhysicalPlan) AddNoGroupingStageWithCoreFunc(
	coreFunc func(int, *Processor) execinfrapb.ProcessorCoreUnion,
	post execinfrapb.PostProcessSpec,
	outputTypes []*types.T,
	newOrdering execinfrapb.Ordering,
) {
	__antithesis_instrumentation__.Notify(562199)

	stageID := p.NewStage(p.IsLastStageDistributed(), false)
	prevStageResultTypes := p.GetResultTypes()
	for i, resultProc := range p.ResultRouters {
		__antithesis_instrumentation__.Notify(562201)
		prevProc := &p.Processors[resultProc]

		proc := Processor{
			SQLInstanceID: prevProc.SQLInstanceID,
			Spec: execinfrapb.ProcessorSpec{
				Input: []execinfrapb.InputSyncSpec{{
					Type:        execinfrapb.InputSyncSpec_PARALLEL_UNORDERED,
					ColumnTypes: prevStageResultTypes,
				}},
				Core: coreFunc(int(resultProc), prevProc),
				Post: post,
				Output: []execinfrapb.OutputRouterSpec{{
					Type: execinfrapb.OutputRouterSpec_PASS_THROUGH,
				}},
				StageID:     stageID,
				ResultTypes: outputTypes,
			},
		}

		pIdx := p.AddProcessor(proc)

		p.Streams = append(p.Streams, Stream{
			SourceProcessor:  resultProc,
			DestProcessor:    pIdx,
			SourceRouterSlot: 0,
			DestInput:        0,
		})

		p.ResultRouters[i] = pIdx
	}
	__antithesis_instrumentation__.Notify(562200)
	p.SetMergeOrdering(newOrdering)
}

func (p *PhysicalPlan) MergeResultStreams(
	resultRouters []ProcessorIdx,
	sourceRouterSlot int,
	ordering execinfrapb.Ordering,
	destProcessor ProcessorIdx,
	destInput int,
	forceSerialization bool,
) {
	__antithesis_instrumentation__.Notify(562202)
	proc := &p.Processors[destProcessor]
	if len(ordering.Columns) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(562204)
		return len(resultRouters) > 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(562205)
		proc.Spec.Input[destInput].Type = execinfrapb.InputSyncSpec_ORDERED
		proc.Spec.Input[destInput].Ordering = ordering
	} else {
		__antithesis_instrumentation__.Notify(562206)
		if forceSerialization && func() bool {
			__antithesis_instrumentation__.Notify(562207)
			return len(resultRouters) > 1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(562208)

			proc.Spec.Input[destInput].Type = execinfrapb.InputSyncSpec_SERIAL_UNORDERED
		} else {
			__antithesis_instrumentation__.Notify(562209)
			proc.Spec.Input[destInput].Type = execinfrapb.InputSyncSpec_PARALLEL_UNORDERED
		}
	}
	__antithesis_instrumentation__.Notify(562203)

	for _, resultProc := range resultRouters {
		__antithesis_instrumentation__.Notify(562210)
		p.Streams = append(p.Streams, Stream{
			SourceProcessor:  resultProc,
			SourceRouterSlot: sourceRouterSlot,
			DestProcessor:    destProcessor,
			DestInput:        destInput,
		})
	}
}

func (p *PhysicalPlan) AddSingleGroupStage(
	sqlInstanceID base.SQLInstanceID,
	core execinfrapb.ProcessorCoreUnion,
	post execinfrapb.PostProcessSpec,
	outputTypes []*types.T,
) {
	__antithesis_instrumentation__.Notify(562211)
	proc := Processor{
		SQLInstanceID: sqlInstanceID,
		Spec: execinfrapb.ProcessorSpec{
			Input: []execinfrapb.InputSyncSpec{{

				ColumnTypes: p.GetResultTypes(),
			}},
			Core: core,
			Post: post,
			Output: []execinfrapb.OutputRouterSpec{{
				Type: execinfrapb.OutputRouterSpec_PASS_THROUGH,
			}},

			StageID:     p.NewStage(sqlInstanceID != p.GatewaySQLInstanceID, false),
			ResultTypes: outputTypes,
		},
	}

	pIdx := p.AddProcessor(proc)

	p.MergeResultStreams(p.ResultRouters, 0, p.MergeOrdering, pIdx, 0, false)

	p.ResultRouters = p.ResultRouters[:1]
	p.ResultRouters[0] = pIdx

	p.MergeOrdering = execinfrapb.Ordering{}
}

func (p *PhysicalPlan) EnsureSingleStreamOnGateway() {
	__antithesis_instrumentation__.Notify(562212)

	if len(p.ResultRouters) != 1 || func() bool {
		__antithesis_instrumentation__.Notify(562214)
		return p.Processors[p.ResultRouters[0]].SQLInstanceID != p.GatewaySQLInstanceID == true
	}() == true {
		__antithesis_instrumentation__.Notify(562215)
		p.AddSingleGroupStage(
			p.GatewaySQLInstanceID,
			execinfrapb.ProcessorCoreUnion{Noop: &execinfrapb.NoopCoreSpec{}},
			execinfrapb.PostProcessSpec{},
			p.GetResultTypes(),
		)
		if len(p.ResultRouters) != 1 || func() bool {
			__antithesis_instrumentation__.Notify(562216)
			return p.Processors[p.ResultRouters[0]].SQLInstanceID != p.GatewaySQLInstanceID == true
		}() == true {
			__antithesis_instrumentation__.Notify(562217)
			panic("ensuring a single stream on the gateway failed")
		} else {
			__antithesis_instrumentation__.Notify(562218)
		}
	} else {
		__antithesis_instrumentation__.Notify(562219)
	}
	__antithesis_instrumentation__.Notify(562213)

	p.MergeOrdering = execinfrapb.Ordering{}
}

func (p *PhysicalPlan) CheckLastStagePost() error {
	__antithesis_instrumentation__.Notify(562220)
	post := p.Processors[p.ResultRouters[0]].Spec.Post

	for i := 1; i < len(p.ResultRouters); i++ {
		__antithesis_instrumentation__.Notify(562222)
		pi := &p.Processors[p.ResultRouters[i]].Spec.Post
		if pi.Projection != post.Projection || func() bool {
			__antithesis_instrumentation__.Notify(562225)
			return len(pi.OutputColumns) != len(post.OutputColumns) == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(562226)
			return len(pi.RenderExprs) != len(post.RenderExprs) == true
		}() == true {
			__antithesis_instrumentation__.Notify(562227)
			return errors.Errorf("inconsistent post-processing: %v vs %v", post, pi)
		} else {
			__antithesis_instrumentation__.Notify(562228)
		}
		__antithesis_instrumentation__.Notify(562223)
		for j, col := range pi.OutputColumns {
			__antithesis_instrumentation__.Notify(562229)
			if col != post.OutputColumns[j] {
				__antithesis_instrumentation__.Notify(562230)
				return errors.Errorf("inconsistent post-processing: %v vs %v", post, pi)
			} else {
				__antithesis_instrumentation__.Notify(562231)
			}
		}
		__antithesis_instrumentation__.Notify(562224)
		for j, col := range pi.RenderExprs {
			__antithesis_instrumentation__.Notify(562232)
			if col != post.RenderExprs[j] {
				__antithesis_instrumentation__.Notify(562233)
				return errors.Errorf("inconsistent post-processing: %v vs %v", post, pi)
			} else {
				__antithesis_instrumentation__.Notify(562234)
			}
		}
	}
	__antithesis_instrumentation__.Notify(562221)

	return nil
}

func (p *PhysicalPlan) GetLastStagePost() execinfrapb.PostProcessSpec {
	__antithesis_instrumentation__.Notify(562235)
	if err := p.CheckLastStagePost(); err != nil {
		__antithesis_instrumentation__.Notify(562237)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(562238)
	}
	__antithesis_instrumentation__.Notify(562236)
	return p.Processors[p.ResultRouters[0]].Spec.Post
}

func (p *PhysicalPlan) SetLastStagePost(post execinfrapb.PostProcessSpec, outputTypes []*types.T) {
	__antithesis_instrumentation__.Notify(562239)
	for _, pIdx := range p.ResultRouters {
		__antithesis_instrumentation__.Notify(562240)
		p.Processors[pIdx].Spec.Post = post
		p.Processors[pIdx].Spec.ResultTypes = outputTypes
	}
}

func isIdentityProjection(columns []uint32, numExistingCols int) bool {
	__antithesis_instrumentation__.Notify(562241)
	if len(columns) != numExistingCols {
		__antithesis_instrumentation__.Notify(562244)
		return false
	} else {
		__antithesis_instrumentation__.Notify(562245)
	}
	__antithesis_instrumentation__.Notify(562242)
	for i, c := range columns {
		__antithesis_instrumentation__.Notify(562246)
		if c != uint32(i) {
			__antithesis_instrumentation__.Notify(562247)
			return false
		} else {
			__antithesis_instrumentation__.Notify(562248)
		}
	}
	__antithesis_instrumentation__.Notify(562243)
	return true
}

func (p *PhysicalPlan) AddProjection(columns []uint32, newMergeOrdering execinfrapb.Ordering) {
	__antithesis_instrumentation__.Notify(562249)

	p.SetMergeOrdering(newMergeOrdering)

	if isIdentityProjection(columns, len(p.GetResultTypes())) {
		__antithesis_instrumentation__.Notify(562253)
		return
	} else {
		__antithesis_instrumentation__.Notify(562254)
	}
	__antithesis_instrumentation__.Notify(562250)

	newResultTypes := make([]*types.T, len(columns))
	for i, c := range columns {
		__antithesis_instrumentation__.Notify(562255)
		newResultTypes[i] = p.GetResultTypes()[c]
	}
	__antithesis_instrumentation__.Notify(562251)

	post := p.GetLastStagePost()

	if post.RenderExprs != nil {
		__antithesis_instrumentation__.Notify(562256)

		oldRenders := post.RenderExprs
		post.RenderExprs = make([]execinfrapb.Expression, len(columns))
		for i, c := range columns {
			__antithesis_instrumentation__.Notify(562257)
			post.RenderExprs[i] = oldRenders[c]
		}
	} else {
		__antithesis_instrumentation__.Notify(562258)

		if post.Projection {
			__antithesis_instrumentation__.Notify(562260)

			for i, c := range columns {
				__antithesis_instrumentation__.Notify(562261)
				columns[i] = post.OutputColumns[c]
			}
		} else {
			__antithesis_instrumentation__.Notify(562262)
		}
		__antithesis_instrumentation__.Notify(562259)
		post.OutputColumns = columns
		post.Projection = true
	}
	__antithesis_instrumentation__.Notify(562252)

	p.SetLastStagePost(post, newResultTypes)
}

func exprColumn(expr tree.TypedExpr, indexVarMap []int) (int, bool) {
	__antithesis_instrumentation__.Notify(562263)
	v, ok := expr.(*tree.IndexedVar)
	if !ok {
		__antithesis_instrumentation__.Notify(562265)
		return -1, false
	} else {
		__antithesis_instrumentation__.Notify(562266)
	}
	__antithesis_instrumentation__.Notify(562264)
	return indexVarMap[v.Idx], true
}

func (p *PhysicalPlan) AddRendering(
	exprs []tree.TypedExpr,
	exprCtx ExprContext,
	indexVarMap []int,
	outTypes []*types.T,
	newMergeOrdering execinfrapb.Ordering,
) error {
	__antithesis_instrumentation__.Notify(562267)

	needRendering := false
	identity := len(exprs) == len(p.GetResultTypes())

	for exprIdx, e := range exprs {
		__antithesis_instrumentation__.Notify(562273)
		varIdx, ok := exprColumn(e, indexVarMap)
		if !ok {
			__antithesis_instrumentation__.Notify(562275)
			needRendering = true
			break
		} else {
			__antithesis_instrumentation__.Notify(562276)
		}
		__antithesis_instrumentation__.Notify(562274)
		identity = identity && func() bool {
			__antithesis_instrumentation__.Notify(562277)
			return (varIdx == exprIdx) == true
		}() == true
	}
	__antithesis_instrumentation__.Notify(562268)

	if !needRendering {
		__antithesis_instrumentation__.Notify(562278)
		if identity {
			__antithesis_instrumentation__.Notify(562281)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(562282)
		}
		__antithesis_instrumentation__.Notify(562279)

		cols := make([]uint32, len(exprs))
		for i, e := range exprs {
			__antithesis_instrumentation__.Notify(562283)
			streamCol, _ := exprColumn(e, indexVarMap)
			if streamCol == -1 {
				__antithesis_instrumentation__.Notify(562285)
				panic(errors.AssertionFailedf("render %d refers to column not in source: %s", i, e))
			} else {
				__antithesis_instrumentation__.Notify(562286)
			}
			__antithesis_instrumentation__.Notify(562284)
			cols[i] = uint32(streamCol)
		}
		__antithesis_instrumentation__.Notify(562280)
		p.AddProjection(cols, newMergeOrdering)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(562287)
	}
	__antithesis_instrumentation__.Notify(562269)

	post := p.GetLastStagePost()
	if len(post.RenderExprs) > 0 {
		__antithesis_instrumentation__.Notify(562288)
		post = execinfrapb.PostProcessSpec{}

		p.AddNoGroupingStage(
			execinfrapb.ProcessorCoreUnion{Noop: &execinfrapb.NoopCoreSpec{}},
			post,
			p.GetResultTypes(),
			p.MergeOrdering,
		)
	} else {
		__antithesis_instrumentation__.Notify(562289)
	}
	__antithesis_instrumentation__.Notify(562270)

	compositeMap := indexVarMap
	if post.Projection {
		__antithesis_instrumentation__.Notify(562290)
		compositeMap = reverseProjection(post.OutputColumns, indexVarMap)
	} else {
		__antithesis_instrumentation__.Notify(562291)
	}
	__antithesis_instrumentation__.Notify(562271)
	post.RenderExprs = make([]execinfrapb.Expression, len(exprs))
	for i, e := range exprs {
		__antithesis_instrumentation__.Notify(562292)
		var err error
		post.RenderExprs[i], err = MakeExpression(e, exprCtx, compositeMap)
		if err != nil {
			__antithesis_instrumentation__.Notify(562293)
			return err
		} else {
			__antithesis_instrumentation__.Notify(562294)
		}
	}
	__antithesis_instrumentation__.Notify(562272)
	p.SetMergeOrdering(newMergeOrdering)
	post.Projection = false
	post.OutputColumns = nil
	p.SetLastStagePost(post, outTypes)
	return nil
}

func reverseProjection(outputColumns []uint32, indexVarMap []int) []int {
	__antithesis_instrumentation__.Notify(562295)
	if indexVarMap == nil {
		__antithesis_instrumentation__.Notify(562298)
		panic("no indexVarMap")
	} else {
		__antithesis_instrumentation__.Notify(562299)
	}
	__antithesis_instrumentation__.Notify(562296)
	compositeMap := make([]int, len(indexVarMap))
	for i, col := range indexVarMap {
		__antithesis_instrumentation__.Notify(562300)
		if col == -1 {
			__antithesis_instrumentation__.Notify(562301)
			compositeMap[i] = -1
		} else {
			__antithesis_instrumentation__.Notify(562302)
			compositeMap[i] = int(outputColumns[col])
		}
	}
	__antithesis_instrumentation__.Notify(562297)
	return compositeMap
}

func (p *PhysicalPlan) AddFilter(
	expr tree.TypedExpr, exprCtx ExprContext, indexVarMap []int,
) error {
	__antithesis_instrumentation__.Notify(562303)
	if expr == nil {
		__antithesis_instrumentation__.Notify(562306)
		return errors.Errorf("nil filter")
	} else {
		__antithesis_instrumentation__.Notify(562307)
	}
	__antithesis_instrumentation__.Notify(562304)
	filter, err := MakeExpression(expr, exprCtx, indexVarMap)
	if err != nil {
		__antithesis_instrumentation__.Notify(562308)
		return err
	} else {
		__antithesis_instrumentation__.Notify(562309)
	}
	__antithesis_instrumentation__.Notify(562305)
	p.AddNoGroupingStage(
		execinfrapb.ProcessorCoreUnion{Filterer: &execinfrapb.FiltererSpec{
			Filter: filter,
		}},
		execinfrapb.PostProcessSpec{},
		p.GetResultTypes(),
		p.MergeOrdering,
	)
	return nil
}

func (p *PhysicalPlan) AddLimit(count int64, offset int64, exprCtx ExprContext) error {
	__antithesis_instrumentation__.Notify(562310)
	if count < 0 {
		__antithesis_instrumentation__.Notify(562318)
		return errors.Errorf("negative limit")
	} else {
		__antithesis_instrumentation__.Notify(562319)
	}
	__antithesis_instrumentation__.Notify(562311)
	if offset < 0 {
		__antithesis_instrumentation__.Notify(562320)
		return errors.Errorf("negative offset")
	} else {
		__antithesis_instrumentation__.Notify(562321)
	}
	__antithesis_instrumentation__.Notify(562312)

	limitZero := false
	if count == 0 {
		__antithesis_instrumentation__.Notify(562322)
		count = 1
		limitZero = true
	} else {
		__antithesis_instrumentation__.Notify(562323)
	}
	__antithesis_instrumentation__.Notify(562313)

	if len(p.ResultRouters) == 1 {
		__antithesis_instrumentation__.Notify(562324)

		post := p.GetLastStagePost()
		if offset != 0 {
			__antithesis_instrumentation__.Notify(562328)
			switch {
			case post.Limit > 0 && func() bool {
				__antithesis_instrumentation__.Notify(562332)
				return post.Limit <= uint64(offset) == true
			}() == true:
				__antithesis_instrumentation__.Notify(562329)

				count = 1
				limitZero = true

			case post.Offset > math.MaxUint64-uint64(offset):
				__antithesis_instrumentation__.Notify(562330)

				count = 1
				limitZero = true

			default:
				__antithesis_instrumentation__.Notify(562331)

				post.Offset += uint64(offset)
				if post.Limit > 0 {
					__antithesis_instrumentation__.Notify(562333)

					post.Limit -= uint64(offset)
				} else {
					__antithesis_instrumentation__.Notify(562334)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(562335)
		}
		__antithesis_instrumentation__.Notify(562325)
		if count != math.MaxInt64 && func() bool {
			__antithesis_instrumentation__.Notify(562336)
			return (post.Limit == 0 || func() bool {
				__antithesis_instrumentation__.Notify(562337)
				return post.Limit > uint64(count) == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(562338)
			post.Limit = uint64(count)
		} else {
			__antithesis_instrumentation__.Notify(562339)
		}
		__antithesis_instrumentation__.Notify(562326)
		p.SetLastStagePost(post, p.GetResultTypes())
		if limitZero {
			__antithesis_instrumentation__.Notify(562340)
			if err := p.AddFilter(tree.DBoolFalse, exprCtx, nil); err != nil {
				__antithesis_instrumentation__.Notify(562341)
				return err
			} else {
				__antithesis_instrumentation__.Notify(562342)
			}
		} else {
			__antithesis_instrumentation__.Notify(562343)
		}
		__antithesis_instrumentation__.Notify(562327)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(562344)
	}
	__antithesis_instrumentation__.Notify(562314)

	if count != math.MaxInt64 {
		__antithesis_instrumentation__.Notify(562345)
		post := p.GetLastStagePost()

		localLimit := uint64(count + offset)
		if post.Limit == 0 || func() bool {
			__antithesis_instrumentation__.Notify(562346)
			return post.Limit > localLimit == true
		}() == true {
			__antithesis_instrumentation__.Notify(562347)
			post.Limit = localLimit
			p.SetLastStagePost(post, p.GetResultTypes())
		} else {
			__antithesis_instrumentation__.Notify(562348)
		}
	} else {
		__antithesis_instrumentation__.Notify(562349)
	}
	__antithesis_instrumentation__.Notify(562315)

	post := execinfrapb.PostProcessSpec{
		Offset: uint64(offset),
	}
	if count != math.MaxInt64 {
		__antithesis_instrumentation__.Notify(562350)
		post.Limit = uint64(count)
	} else {
		__antithesis_instrumentation__.Notify(562351)
	}
	__antithesis_instrumentation__.Notify(562316)
	p.AddSingleGroupStage(
		p.GatewaySQLInstanceID,
		execinfrapb.ProcessorCoreUnion{Noop: &execinfrapb.NoopCoreSpec{}},
		post,
		p.GetResultTypes(),
	)
	if limitZero {
		__antithesis_instrumentation__.Notify(562352)
		if err := p.AddFilter(tree.DBoolFalse, exprCtx, nil); err != nil {
			__antithesis_instrumentation__.Notify(562353)
			return err
		} else {
			__antithesis_instrumentation__.Notify(562354)
		}
	} else {
		__antithesis_instrumentation__.Notify(562355)
	}
	__antithesis_instrumentation__.Notify(562317)
	return nil
}

func (p *PhysicalPlan) PopulateEndpoints() {
	__antithesis_instrumentation__.Notify(562356)

	for sIdx, s := range p.Streams {
		__antithesis_instrumentation__.Notify(562357)
		p1 := &p.Processors[s.SourceProcessor]
		p2 := &p.Processors[s.DestProcessor]
		endpoint := execinfrapb.StreamEndpointSpec{StreamID: execinfrapb.StreamID(sIdx)}
		if p1.SQLInstanceID == p2.SQLInstanceID {
			__antithesis_instrumentation__.Notify(562361)
			endpoint.Type = execinfrapb.StreamEndpointSpec_LOCAL
		} else {
			__antithesis_instrumentation__.Notify(562362)
			endpoint.Type = execinfrapb.StreamEndpointSpec_REMOTE
		}
		__antithesis_instrumentation__.Notify(562358)
		if endpoint.Type == execinfrapb.StreamEndpointSpec_REMOTE {
			__antithesis_instrumentation__.Notify(562363)
			endpoint.OriginNodeID = p1.SQLInstanceID
			endpoint.TargetNodeID = p2.SQLInstanceID
		} else {
			__antithesis_instrumentation__.Notify(562364)
		}
		__antithesis_instrumentation__.Notify(562359)
		p2.Spec.Input[s.DestInput].Streams = append(p2.Spec.Input[s.DestInput].Streams, endpoint)

		router := &p1.Spec.Output[0]

		if len(router.Streams) != s.SourceRouterSlot {
			__antithesis_instrumentation__.Notify(562365)
			panic(errors.AssertionFailedf(
				"sourceRouterSlot mismatch: %d, expected %d", len(router.Streams), s.SourceRouterSlot,
			))
		} else {
			__antithesis_instrumentation__.Notify(562366)
		}
		__antithesis_instrumentation__.Notify(562360)
		router.Streams = append(router.Streams, endpoint)
	}
}

func (p *PhysicalPlan) GenerateFlowSpecs() map[base.SQLInstanceID]*execinfrapb.FlowSpec {
	__antithesis_instrumentation__.Notify(562367)
	flowID := execinfrapb.FlowID{
		UUID: p.FlowID,
	}
	flows := make(map[base.SQLInstanceID]*execinfrapb.FlowSpec, 1)

	for _, proc := range p.Processors {
		__antithesis_instrumentation__.Notify(562369)
		flowSpec, ok := flows[proc.SQLInstanceID]
		if !ok {
			__antithesis_instrumentation__.Notify(562371)
			flowSpec = NewFlowSpec(flowID, p.GatewaySQLInstanceID)
			flows[proc.SQLInstanceID] = flowSpec
		} else {
			__antithesis_instrumentation__.Notify(562372)
		}
		__antithesis_instrumentation__.Notify(562370)
		flowSpec.Processors = append(flowSpec.Processors, proc.Spec)
	}
	__antithesis_instrumentation__.Notify(562368)
	return flows
}

func (p *PhysicalPlan) SetRowEstimates(left, right *PhysicalPlan) {
	__antithesis_instrumentation__.Notify(562373)
	p.TotalEstimatedScannedRows = left.TotalEstimatedScannedRows + right.TotalEstimatedScannedRows
}

func MergePlans(
	mergedPlan *PhysicalPlan,
	left, right *PhysicalPlan,
	leftPlanDistribution, rightPlanDistribution PlanDistribution,
	allowPartialDistribution bool,
) {
	__antithesis_instrumentation__.Notify(562374)
	if mergedPlan.PhysicalInfrastructure != left.PhysicalInfrastructure || func() bool {
		__antithesis_instrumentation__.Notify(562376)
		return mergedPlan.PhysicalInfrastructure != right.PhysicalInfrastructure == true
	}() == true {
		__antithesis_instrumentation__.Notify(562377)
		panic(errors.AssertionFailedf("can only merge plans that share infrastructure"))
	} else {
		__antithesis_instrumentation__.Notify(562378)
	}
	__antithesis_instrumentation__.Notify(562375)
	mergedPlan.SetRowEstimates(left, right)
	mergedPlan.Distribution = leftPlanDistribution.compose(rightPlanDistribution, allowPartialDistribution)
}

func (p *PhysicalPlan) AddJoinStage(
	sqlInstanceIDs []base.SQLInstanceID,
	core execinfrapb.ProcessorCoreUnion,
	post execinfrapb.PostProcessSpec,
	leftEqCols, rightEqCols []uint32,
	leftTypes, rightTypes []*types.T,
	leftMergeOrd, rightMergeOrd execinfrapb.Ordering,
	leftRouters, rightRouters []ProcessorIdx,
	resultTypes []*types.T,
) {
	__antithesis_instrumentation__.Notify(562379)
	pIdxStart := ProcessorIdx(len(p.Processors))
	stageID := p.NewStageOnNodes(sqlInstanceIDs)

	for _, sqlInstanceID := range sqlInstanceIDs {
		__antithesis_instrumentation__.Notify(562382)
		inputs := make([]execinfrapb.InputSyncSpec, 0, 2)
		inputs = append(inputs, execinfrapb.InputSyncSpec{ColumnTypes: leftTypes})
		inputs = append(inputs, execinfrapb.InputSyncSpec{ColumnTypes: rightTypes})

		proc := Processor{
			SQLInstanceID: sqlInstanceID,
			Spec: execinfrapb.ProcessorSpec{
				Input:       inputs,
				Core:        core,
				Post:        post,
				Output:      []execinfrapb.OutputRouterSpec{{Type: execinfrapb.OutputRouterSpec_PASS_THROUGH}},
				StageID:     stageID,
				ResultTypes: resultTypes,
			},
		}
		p.Processors = append(p.Processors, proc)
	}
	__antithesis_instrumentation__.Notify(562380)

	if len(sqlInstanceIDs) > 1 {
		__antithesis_instrumentation__.Notify(562383)

		for _, resultProc := range leftRouters {
			__antithesis_instrumentation__.Notify(562385)
			p.Processors[resultProc].Spec.Output[0] = execinfrapb.OutputRouterSpec{
				Type:        execinfrapb.OutputRouterSpec_BY_HASH,
				HashColumns: leftEqCols,
			}
		}
		__antithesis_instrumentation__.Notify(562384)

		for _, resultProc := range rightRouters {
			__antithesis_instrumentation__.Notify(562386)
			p.Processors[resultProc].Spec.Output[0] = execinfrapb.OutputRouterSpec{
				Type:        execinfrapb.OutputRouterSpec_BY_HASH,
				HashColumns: rightEqCols,
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(562387)
	}
	__antithesis_instrumentation__.Notify(562381)
	p.ResultRouters = p.ResultRouters[:0]

	for bucket := 0; bucket < len(sqlInstanceIDs); bucket++ {
		__antithesis_instrumentation__.Notify(562388)
		pIdx := pIdxStart + ProcessorIdx(bucket)

		p.MergeResultStreams(leftRouters, bucket, leftMergeOrd, pIdx, 0, false)

		p.MergeResultStreams(rightRouters, bucket, rightMergeOrd, pIdx, 1, false)

		p.ResultRouters = append(p.ResultRouters, pIdx)
	}
}

func (p *PhysicalPlan) AddStageOnNodes(
	sqlInstanceIDs []base.SQLInstanceID,
	core execinfrapb.ProcessorCoreUnion,
	post execinfrapb.PostProcessSpec,
	hashCols []uint32,
	inputTypes, resultTypes []*types.T,
	mergeOrd execinfrapb.Ordering,
	routers []ProcessorIdx,
) {
	__antithesis_instrumentation__.Notify(562389)
	pIdxStart := len(p.Processors)
	newStageID := p.NewStageOnNodes(sqlInstanceIDs)

	for _, sqlInstanceID := range sqlInstanceIDs {
		__antithesis_instrumentation__.Notify(562393)
		proc := Processor{
			SQLInstanceID: sqlInstanceID,
			Spec: execinfrapb.ProcessorSpec{
				Input: []execinfrapb.InputSyncSpec{
					{ColumnTypes: inputTypes},
				},
				Core:        core,
				Post:        post,
				Output:      []execinfrapb.OutputRouterSpec{{Type: execinfrapb.OutputRouterSpec_PASS_THROUGH}},
				StageID:     newStageID,
				ResultTypes: resultTypes,
			},
		}
		p.AddProcessor(proc)
	}
	__antithesis_instrumentation__.Notify(562390)

	if len(sqlInstanceIDs) > 1 {
		__antithesis_instrumentation__.Notify(562394)

		for _, resultProc := range routers {
			__antithesis_instrumentation__.Notify(562395)
			p.Processors[resultProc].Spec.Output[0] = execinfrapb.OutputRouterSpec{
				Type:        execinfrapb.OutputRouterSpec_BY_HASH,
				HashColumns: hashCols,
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(562396)
	}
	__antithesis_instrumentation__.Notify(562391)

	for bucket := 0; bucket < len(sqlInstanceIDs); bucket++ {
		__antithesis_instrumentation__.Notify(562397)
		pIdx := ProcessorIdx(pIdxStart + bucket)
		p.MergeResultStreams(routers, bucket, mergeOrd, pIdx, 0, false)
	}
	__antithesis_instrumentation__.Notify(562392)

	p.ResultRouters = p.ResultRouters[:0]
	for i := 0; i < len(sqlInstanceIDs); i++ {
		__antithesis_instrumentation__.Notify(562398)
		p.ResultRouters = append(p.ResultRouters, ProcessorIdx(pIdxStart+i))
	}
}

func (p *PhysicalPlan) AddDistinctSetOpStage(
	sqlInstanceIDs []base.SQLInstanceID,
	joinCore execinfrapb.ProcessorCoreUnion,
	distinctCores []execinfrapb.ProcessorCoreUnion,
	post execinfrapb.PostProcessSpec,
	eqCols []uint32,
	leftTypes, rightTypes []*types.T,
	leftMergeOrd, rightMergeOrd execinfrapb.Ordering,
	leftRouters, rightRouters []ProcessorIdx,
	resultTypes []*types.T,
) {
	__antithesis_instrumentation__.Notify(562399)

	distinctProcs := make(map[base.SQLInstanceID][]ProcessorIdx)
	p.AddStageOnNodes(
		sqlInstanceIDs, distinctCores[0], execinfrapb.PostProcessSpec{}, eqCols,
		leftTypes, leftTypes, leftMergeOrd, leftRouters,
	)
	for _, leftDistinctProcIdx := range p.ResultRouters {
		__antithesis_instrumentation__.Notify(562402)
		node := p.Processors[leftDistinctProcIdx].SQLInstanceID
		distinctProcs[node] = append(distinctProcs[node], leftDistinctProcIdx)
	}
	__antithesis_instrumentation__.Notify(562400)
	p.AddStageOnNodes(
		sqlInstanceIDs, distinctCores[1], execinfrapb.PostProcessSpec{}, eqCols,
		rightTypes, rightTypes, rightMergeOrd, rightRouters,
	)
	for _, rightDistinctProcIdx := range p.ResultRouters {
		__antithesis_instrumentation__.Notify(562403)
		node := p.Processors[rightDistinctProcIdx].SQLInstanceID
		distinctProcs[node] = append(distinctProcs[node], rightDistinctProcIdx)
	}
	__antithesis_instrumentation__.Notify(562401)

	joinStageID := p.NewStageOnNodes(sqlInstanceIDs)
	p.ResultRouters = p.ResultRouters[:0]

	for _, n := range sqlInstanceIDs {
		__antithesis_instrumentation__.Notify(562404)
		proc := Processor{
			SQLInstanceID: n,
			Spec: execinfrapb.ProcessorSpec{
				Input: []execinfrapb.InputSyncSpec{
					{ColumnTypes: leftTypes},
					{ColumnTypes: rightTypes},
				},
				Core:        joinCore,
				Post:        post,
				Output:      []execinfrapb.OutputRouterSpec{{Type: execinfrapb.OutputRouterSpec_PASS_THROUGH}},
				StageID:     joinStageID,
				ResultTypes: resultTypes,
			},
		}
		pIdx := p.AddProcessor(proc)

		for side, distinctProc := range distinctProcs[n] {
			__antithesis_instrumentation__.Notify(562406)
			p.Streams = append(p.Streams, Stream{
				SourceProcessor:  distinctProc,
				SourceRouterSlot: 0,
				DestProcessor:    pIdx,
				DestInput:        side,
			})
		}
		__antithesis_instrumentation__.Notify(562405)

		p.ResultRouters = append(p.ResultRouters, pIdx)
	}
}

func (p *PhysicalPlan) EnsureSingleStreamPerNode(
	forceSerialization bool, post execinfrapb.PostProcessSpec,
) {
	__antithesis_instrumentation__.Notify(562407)

	var nodes util.FastIntSet
	var foundDuplicates bool
	for _, pIdx := range p.ResultRouters {
		__antithesis_instrumentation__.Notify(562410)
		proc := &p.Processors[pIdx]
		if nodes.Contains(int(proc.SQLInstanceID)) {
			__antithesis_instrumentation__.Notify(562412)
			foundDuplicates = true
			break
		} else {
			__antithesis_instrumentation__.Notify(562413)
		}
		__antithesis_instrumentation__.Notify(562411)
		nodes.Add(int(proc.SQLInstanceID))
	}
	__antithesis_instrumentation__.Notify(562408)
	if !foundDuplicates {
		__antithesis_instrumentation__.Notify(562414)
		return
	} else {
		__antithesis_instrumentation__.Notify(562415)
	}
	__antithesis_instrumentation__.Notify(562409)
	streams := make([]ProcessorIdx, 0, 2)

	for i := 0; i < len(p.ResultRouters); i++ {
		__antithesis_instrumentation__.Notify(562416)
		pIdx := p.ResultRouters[i]
		node := p.Processors[p.ResultRouters[i]].SQLInstanceID
		streams = append(streams[:0], pIdx)

		for j := i + 1; j < len(p.ResultRouters); {
			__antithesis_instrumentation__.Notify(562419)
			if p.Processors[p.ResultRouters[j]].SQLInstanceID == node {
				__antithesis_instrumentation__.Notify(562420)
				streams = append(streams, p.ResultRouters[j])

				copy(p.ResultRouters[j:], p.ResultRouters[j+1:])
				p.ResultRouters = p.ResultRouters[:len(p.ResultRouters)-1]
			} else {
				__antithesis_instrumentation__.Notify(562421)
				j++
			}
		}
		__antithesis_instrumentation__.Notify(562417)
		if len(streams) == 1 {
			__antithesis_instrumentation__.Notify(562422)

			continue
		} else {
			__antithesis_instrumentation__.Notify(562423)
		}
		__antithesis_instrumentation__.Notify(562418)

		proc := Processor{
			SQLInstanceID: node,
			Spec: execinfrapb.ProcessorSpec{
				Input: []execinfrapb.InputSyncSpec{{

					ColumnTypes: p.GetResultTypes(),
				}},
				Post:        post,
				Core:        execinfrapb.ProcessorCoreUnion{Noop: &execinfrapb.NoopCoreSpec{}},
				Output:      []execinfrapb.OutputRouterSpec{{Type: execinfrapb.OutputRouterSpec_PASS_THROUGH}},
				ResultTypes: p.GetResultTypes(),
			},
		}
		mergedProcIdx := p.AddProcessor(proc)
		p.MergeResultStreams(streams, 0, p.MergeOrdering, mergedProcIdx, 0, forceSerialization)
		p.ResultRouters[i] = mergedProcIdx
	}
}

func (p *PhysicalPlan) GetLastStageDistribution() PlanDistribution {
	__antithesis_instrumentation__.Notify(562424)
	for i := range p.ResultRouters {
		__antithesis_instrumentation__.Notify(562426)
		if p.Processors[p.ResultRouters[i]].SQLInstanceID != p.GatewaySQLInstanceID {
			__antithesis_instrumentation__.Notify(562427)
			return FullyDistributedPlan
		} else {
			__antithesis_instrumentation__.Notify(562428)
		}
	}
	__antithesis_instrumentation__.Notify(562425)
	return LocalPlan
}

func (p *PhysicalPlan) IsLastStageDistributed() bool {
	__antithesis_instrumentation__.Notify(562429)
	return p.GetLastStageDistribution() != LocalPlan
}

type PlanDistribution int

const (
	LocalPlan PlanDistribution = iota

	PartiallyDistributedPlan

	FullyDistributedPlan
)

func (a PlanDistribution) WillDistribute() bool {
	__antithesis_instrumentation__.Notify(562430)
	return a != LocalPlan
}

func (a PlanDistribution) String() string {
	__antithesis_instrumentation__.Notify(562431)
	switch a {
	case LocalPlan:
		__antithesis_instrumentation__.Notify(562432)
		return "local"
	case PartiallyDistributedPlan:
		__antithesis_instrumentation__.Notify(562433)
		return "partial"
	case FullyDistributedPlan:
		__antithesis_instrumentation__.Notify(562434)
		return "full"
	default:
		__antithesis_instrumentation__.Notify(562435)
		panic(errors.AssertionFailedf("unsupported PlanDistribution %d", a))
	}
}

func (a PlanDistribution) compose(
	b PlanDistribution, allowPartialDistribution bool,
) PlanDistribution {
	__antithesis_instrumentation__.Notify(562436)
	if allowPartialDistribution && func() bool {
		__antithesis_instrumentation__.Notify(562439)
		return a != b == true
	}() == true {
		__antithesis_instrumentation__.Notify(562440)
		return PartiallyDistributedPlan
	} else {
		__antithesis_instrumentation__.Notify(562441)
	}
	__antithesis_instrumentation__.Notify(562437)

	if a != LocalPlan || func() bool {
		__antithesis_instrumentation__.Notify(562442)
		return b != LocalPlan == true
	}() == true {
		__antithesis_instrumentation__.Notify(562443)
		return FullyDistributedPlan
	} else {
		__antithesis_instrumentation__.Notify(562444)
	}
	__antithesis_instrumentation__.Notify(562438)
	return LocalPlan
}
