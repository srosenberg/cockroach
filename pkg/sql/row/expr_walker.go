package row

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math/rand"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemaexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/seqexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/builtins"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

const reseedRandEveryN = 1000

const chunkSizeIncrementRate = 10
const initialChunkSize = 10
const maxChunkSize = 100000

type importRandPosition int64

func (pos importRandPosition) distance(o importRandPosition) int64 {
	__antithesis_instrumentation__.Notify(567477)
	diff := int64(pos) - int64(o)
	if diff < 0 {
		__antithesis_instrumentation__.Notify(567479)
		return -diff
	} else {
		__antithesis_instrumentation__.Notify(567480)
	}
	__antithesis_instrumentation__.Notify(567478)
	return diff
}

func getPosForRandImport(rowID int64, sourceID int32, numInstances int) importRandPosition {
	__antithesis_instrumentation__.Notify(567481)

	rowIDWithMultiplier := int64(numInstances) * rowID
	pos := (int64(sourceID) << rowIDBits) ^ rowIDWithMultiplier
	return importRandPosition(pos)
}

type randomSource interface {
	Float64(c *CellInfoAnnotation) float64

	Int63(c *CellInfoAnnotation) int64
}

var _ randomSource = (*importRand)(nil)

type importRand struct {
	*rand.Rand
	pos importRandPosition
}

func (r *importRand) reseed(pos importRandPosition) {
	__antithesis_instrumentation__.Notify(567482)
	adjPos := (pos / reseedRandEveryN) * reseedRandEveryN
	rnd := rand.New(rand.NewSource(int64(adjPos)))
	for i := int(pos % reseedRandEveryN); i > 0; i-- {
		__antithesis_instrumentation__.Notify(567484)
		_ = rnd.Float64()
	}
	__antithesis_instrumentation__.Notify(567483)

	r.Rand = rnd
	r.pos = pos
}

func (r *importRand) maybeReseed(c *CellInfoAnnotation) {
	__antithesis_instrumentation__.Notify(567485)

	newRowPos := getPosForRandImport(c.rowID, c.sourceID, c.randInstancePerRow)
	rowsSkipped := newRowPos.distance(r.pos) > int64(c.randInstancePerRow)
	if rowsSkipped {
		__antithesis_instrumentation__.Notify(567487)

		r.reseed(newRowPos)
	} else {
		__antithesis_instrumentation__.Notify(567488)
	}
	__antithesis_instrumentation__.Notify(567486)
	if r.pos%reseedRandEveryN == 0 {
		__antithesis_instrumentation__.Notify(567489)
		r.reseed(r.pos)
	} else {
		__antithesis_instrumentation__.Notify(567490)
	}
}

func (r *importRand) Float64(c *CellInfoAnnotation) float64 {
	__antithesis_instrumentation__.Notify(567491)
	r.maybeReseed(c)
	randNum := r.Rand.Float64()
	r.pos++
	return randNum
}

func (r *importRand) Int63(c *CellInfoAnnotation) int64 {
	__antithesis_instrumentation__.Notify(567492)
	r.maybeReseed(c)
	randNum := r.Rand.Int63()
	r.pos++
	return randNum
}

func makeBuiltinOverride(
	builtin *tree.FunctionDefinition, overloads ...tree.Overload,
) *tree.FunctionDefinition {
	__antithesis_instrumentation__.Notify(567493)
	props := builtin.FunctionProperties
	return tree.NewFunctionDefinition(
		"import."+builtin.Name, &props, overloads)
}

type SequenceMetadata struct {
	seqDesc         catalog.TableDescriptor
	instancesPerRow int64
	curChunk        *jobspb.SequenceValChunk
	curVal          int64
}

type overrideVolatility bool

const (
	overrideErrorTerm overrideVolatility = false
	overrideImmutable overrideVolatility = false
	overrideVolatile  overrideVolatility = true
)

const cellInfoAddr tree.AnnotationIdx = iota + 1

type CellInfoAnnotation struct {
	sourceID int32
	rowID    int64

	uniqueRowIDInstance int
	uniqueRowIDTotal    int

	randSource         randomSource
	randInstancePerRow int

	seqNameToMetadata map[string]*SequenceMetadata
	seqIDToMetadata   map[descpb.ID]*SequenceMetadata
	seqChunkProvider  *SeqChunkProvider
}

func getCellInfoAnnotation(t *tree.Annotations) *CellInfoAnnotation {
	__antithesis_instrumentation__.Notify(567494)
	return t.Get(cellInfoAddr).(*CellInfoAnnotation)
}

func (c *CellInfoAnnotation) reset(sourceID int32, rowID int64) {
	__antithesis_instrumentation__.Notify(567495)
	c.sourceID = sourceID
	c.rowID = rowID
	c.uniqueRowIDInstance = 0
}

func makeImportRand(c *CellInfoAnnotation) randomSource {
	__antithesis_instrumentation__.Notify(567496)
	pos := getPosForRandImport(c.rowID, c.sourceID, c.randInstancePerRow)
	randSource := &importRand{}
	randSource.reseed(pos)
	return randSource
}

func importUniqueRowID(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(567497)
	c := getCellInfoAnnotation(evalCtx.Annotations)
	avoidCollisionsWithSQLsIDs := uint64(1 << 63)
	shiftedIndex := int64(c.uniqueRowIDTotal)*c.rowID + int64(c.uniqueRowIDInstance)
	returnIndex := (uint64(c.sourceID) << rowIDBits) ^ uint64(shiftedIndex)
	c.uniqueRowIDInstance++
	evalCtx.Annotations.Set(cellInfoAddr, c)
	return tree.NewDInt(tree.DInt(avoidCollisionsWithSQLsIDs | returnIndex)), nil
}

func importRandom(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(567498)
	c := getCellInfoAnnotation(evalCtx.Annotations)
	if c.randSource == nil {
		__antithesis_instrumentation__.Notify(567500)
		c.randSource = makeImportRand(c)
	} else {
		__antithesis_instrumentation__.Notify(567501)
	}
	__antithesis_instrumentation__.Notify(567499)
	return tree.NewDFloat(tree.DFloat(c.randSource.Float64(c))), nil
}

func importGenUUID(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(567502)
	c := getCellInfoAnnotation(evalCtx.Annotations)
	if c.randSource == nil {
		__antithesis_instrumentation__.Notify(567504)
		c.randSource = makeImportRand(c)
	} else {
		__antithesis_instrumentation__.Notify(567505)
	}
	__antithesis_instrumentation__.Notify(567503)
	gen := c.randSource.Int63(c)
	id := uuid.MakeV4()
	id.DeterministicV4(uint64(gen), uint64(1<<63))
	return tree.NewDUuid(tree.DUuid{UUID: id}), nil
}

type SeqChunkProvider struct {
	JobID    jobspb.JobID
	Registry *jobs.Registry
}

func (j *SeqChunkProvider) RequestChunk(
	evalCtx *tree.EvalContext, c *CellInfoAnnotation, seqMetadata *SequenceMetadata,
) error {
	__antithesis_instrumentation__.Notify(567506)
	var hasAllocatedChunk bool
	return evalCtx.DB.Txn(evalCtx.Context, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(567507)
		var foundFromPreviouslyAllocatedChunk bool
		resolveChunkFunc := func(txn *kv.Txn, md jobs.JobMetadata, ju *jobs.JobUpdater) error {
			__antithesis_instrumentation__.Notify(567511)
			progress := md.Progress

			var err error
			if foundFromPreviouslyAllocatedChunk, err = j.checkForPreviouslyAllocatedChunks(
				seqMetadata, c, progress); err != nil {
				__antithesis_instrumentation__.Notify(567517)
				return err
			} else {
				__antithesis_instrumentation__.Notify(567518)
				if foundFromPreviouslyAllocatedChunk {
					__antithesis_instrumentation__.Notify(567519)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(567520)
				}
			}
			__antithesis_instrumentation__.Notify(567512)

			if !hasAllocatedChunk {
				__antithesis_instrumentation__.Notify(567521)
				if err := reserveChunkOfSeqVals(evalCtx, c, seqMetadata); err != nil {
					__antithesis_instrumentation__.Notify(567523)
					return err
				} else {
					__antithesis_instrumentation__.Notify(567524)
				}
				__antithesis_instrumentation__.Notify(567522)
				hasAllocatedChunk = true
			} else {
				__antithesis_instrumentation__.Notify(567525)
			}
			__antithesis_instrumentation__.Notify(567513)

			fileProgress := progress.GetImport().SequenceDetails[c.sourceID]
			if fileProgress.SeqIdToChunks == nil {
				__antithesis_instrumentation__.Notify(567526)
				fileProgress.SeqIdToChunks = make(map[int32]*jobspb.SequenceDetails_SequenceChunks)
			} else {
				__antithesis_instrumentation__.Notify(567527)
			}
			__antithesis_instrumentation__.Notify(567514)
			seqID := seqMetadata.seqDesc.GetID()
			if _, ok := fileProgress.SeqIdToChunks[int32(seqID)]; !ok {
				__antithesis_instrumentation__.Notify(567528)
				fileProgress.SeqIdToChunks[int32(seqID)] = &jobspb.SequenceDetails_SequenceChunks{
					Chunks: make([]*jobspb.SequenceValChunk, 0),
				}
			} else {
				__antithesis_instrumentation__.Notify(567529)
			}
			__antithesis_instrumentation__.Notify(567515)

			resumePos := progress.GetImport().ResumePos[c.sourceID]
			trim, chunks := 0, fileProgress.SeqIdToChunks[int32(seqID)].Chunks

			for ; trim < len(chunks) && func() bool {
				__antithesis_instrumentation__.Notify(567530)
				return chunks[trim].NextChunkStartRow <= resumePos == true
			}() == true; trim++ {
				__antithesis_instrumentation__.Notify(567531)
			}
			__antithesis_instrumentation__.Notify(567516)
			fileProgress.SeqIdToChunks[int32(seqID)].Chunks =
				fileProgress.SeqIdToChunks[int32(seqID)].Chunks[trim:]

			fileProgress.SeqIdToChunks[int32(seqID)].Chunks = append(
				fileProgress.SeqIdToChunks[int32(seqID)].Chunks, seqMetadata.curChunk)
			ju.UpdateProgress(progress)
			return nil
		}
		__antithesis_instrumentation__.Notify(567508)
		const useReadLock = true
		err := j.Registry.UpdateJobWithTxn(ctx, j.JobID, txn, useReadLock, resolveChunkFunc)
		if err != nil {
			__antithesis_instrumentation__.Notify(567532)
			return err
		} else {
			__antithesis_instrumentation__.Notify(567533)
		}
		__antithesis_instrumentation__.Notify(567509)

		if !foundFromPreviouslyAllocatedChunk {
			__antithesis_instrumentation__.Notify(567534)
			seqMetadata.curVal = seqMetadata.curChunk.ChunkStartVal
		} else {
			__antithesis_instrumentation__.Notify(567535)
		}
		__antithesis_instrumentation__.Notify(567510)
		return nil
	})
}

func incrementSequenceByVal(
	ctx context.Context,
	descriptor catalog.TableDescriptor,
	db *kv.DB,
	codec keys.SQLCodec,
	incrementBy int64,
) (int64, error) {
	__antithesis_instrumentation__.Notify(567536)
	seqOpts := descriptor.GetSequenceOpts()
	var val int64
	var err error

	if seqOpts.Virtual {
		__antithesis_instrumentation__.Notify(567540)
		return 0, errors.New("virtual sequences are not supported by IMPORT INTO")
	} else {
		__antithesis_instrumentation__.Notify(567541)
	}
	__antithesis_instrumentation__.Notify(567537)
	seqValueKey := codec.SequenceKey(uint32(descriptor.GetID()))
	val, err = kv.IncrementValRetryable(ctx, db, seqValueKey, incrementBy)
	if err != nil {
		__antithesis_instrumentation__.Notify(567542)
		if errors.HasType(err, (*roachpb.IntegerOverflowError)(nil)) {
			__antithesis_instrumentation__.Notify(567544)
			return 0, boundsExceededError(descriptor)
		} else {
			__antithesis_instrumentation__.Notify(567545)
		}
		__antithesis_instrumentation__.Notify(567543)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(567546)
	}
	__antithesis_instrumentation__.Notify(567538)
	if val > seqOpts.MaxValue || func() bool {
		__antithesis_instrumentation__.Notify(567547)
		return val < seqOpts.MinValue == true
	}() == true {
		__antithesis_instrumentation__.Notify(567548)
		return 0, boundsExceededError(descriptor)
	} else {
		__antithesis_instrumentation__.Notify(567549)
	}
	__antithesis_instrumentation__.Notify(567539)

	return val, nil
}

func boundsExceededError(descriptor catalog.TableDescriptor) error {
	__antithesis_instrumentation__.Notify(567550)
	seqOpts := descriptor.GetSequenceOpts()
	isAscending := seqOpts.Increment > 0

	var word string
	var value int64
	if isAscending {
		__antithesis_instrumentation__.Notify(567552)
		word = "maximum"
		value = seqOpts.MaxValue
	} else {
		__antithesis_instrumentation__.Notify(567553)
		word = "minimum"
		value = seqOpts.MinValue
	}
	__antithesis_instrumentation__.Notify(567551)
	name := descriptor.GetName()
	return pgerror.Newf(
		pgcode.SequenceGeneratorLimitExceeded,
		`reached %s value of sequence %q (%d)`, word,
		tree.ErrString((*tree.Name)(&name)), value)
}

func (j *SeqChunkProvider) checkForPreviouslyAllocatedChunks(
	seqMetadata *SequenceMetadata, c *CellInfoAnnotation, progress *jobspb.Progress,
) (bool, error) {
	__antithesis_instrumentation__.Notify(567554)
	seqOpts := seqMetadata.seqDesc.GetSequenceOpts()
	var found bool
	fileProgress := progress.GetImport().SequenceDetails[c.sourceID]
	if fileProgress.SeqIdToChunks == nil {
		__antithesis_instrumentation__.Notify(567558)
		return found, nil
	} else {
		__antithesis_instrumentation__.Notify(567559)
	}
	__antithesis_instrumentation__.Notify(567555)
	var allocatedSeqChunks *jobspb.SequenceDetails_SequenceChunks
	var ok bool
	if allocatedSeqChunks, ok = fileProgress.SeqIdToChunks[int32(seqMetadata.seqDesc.GetID())]; !ok {
		__antithesis_instrumentation__.Notify(567560)
		return found, nil
	} else {
		__antithesis_instrumentation__.Notify(567561)
	}
	__antithesis_instrumentation__.Notify(567556)

	for _, chunk := range allocatedSeqChunks.Chunks {
		__antithesis_instrumentation__.Notify(567562)

		if chunk.ChunkStartRow <= c.rowID && func() bool {
			__antithesis_instrumentation__.Notify(567563)
			return chunk.NextChunkStartRow > c.rowID == true
		}() == true {
			__antithesis_instrumentation__.Notify(567564)
			relativeRowIndex := c.rowID - chunk.ChunkStartRow
			seqMetadata.curVal = chunk.ChunkStartVal + seqOpts.Increment*(seqMetadata.instancesPerRow*relativeRowIndex)
			found = true
			return found, nil
		} else {
			__antithesis_instrumentation__.Notify(567565)
		}
	}
	__antithesis_instrumentation__.Notify(567557)
	return found, nil
}

func reserveChunkOfSeqVals(
	evalCtx *tree.EvalContext, c *CellInfoAnnotation, seqMetadata *SequenceMetadata,
) error {
	__antithesis_instrumentation__.Notify(567566)
	seqOpts := seqMetadata.seqDesc.GetSequenceOpts()
	newChunkSize := int64(initialChunkSize)

	if seqMetadata.curChunk != nil {
		__antithesis_instrumentation__.Notify(567570)
		newChunkSize = chunkSizeIncrementRate * seqMetadata.curChunk.ChunkSize
		if newChunkSize > maxChunkSize {
			__antithesis_instrumentation__.Notify(567571)
			newChunkSize = maxChunkSize
		} else {
			__antithesis_instrumentation__.Notify(567572)
		}
	} else {
		__antithesis_instrumentation__.Notify(567573)
	}
	__antithesis_instrumentation__.Notify(567567)

	if newChunkSize < seqMetadata.instancesPerRow {
		__antithesis_instrumentation__.Notify(567574)
		newChunkSize = seqMetadata.instancesPerRow
	} else {
		__antithesis_instrumentation__.Notify(567575)
	}
	__antithesis_instrumentation__.Notify(567568)

	incrementValBy := newChunkSize * seqOpts.Increment

	seqVal, err := incrementSequenceByVal(evalCtx.Context, seqMetadata.seqDesc, evalCtx.DB,
		evalCtx.Codec, incrementValBy)
	if err != nil {
		__antithesis_instrumentation__.Notify(567576)
		return err
	} else {
		__antithesis_instrumentation__.Notify(567577)
	}
	__antithesis_instrumentation__.Notify(567569)

	seqMetadata.curChunk = &jobspb.SequenceValChunk{
		ChunkStartVal:     seqVal - incrementValBy + seqOpts.Increment,
		ChunkSize:         newChunkSize,
		ChunkStartRow:     c.rowID,
		NextChunkStartRow: c.rowID + (newChunkSize / seqMetadata.instancesPerRow),
	}
	return nil
}

func importNextVal(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(567578)
	c := getCellInfoAnnotation(evalCtx.Annotations)
	seqName := tree.MustBeDString(args[0])
	seqMetadata, ok := c.seqNameToMetadata[string(seqName)]
	if !ok {
		__antithesis_instrumentation__.Notify(567580)
		return nil, errors.Newf("sequence %s not found in annotation", seqName)
	} else {
		__antithesis_instrumentation__.Notify(567581)
	}
	__antithesis_instrumentation__.Notify(567579)
	return importNextValHelper(evalCtx, c, seqMetadata)
}

func importNextValByID(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(567582)
	c := getCellInfoAnnotation(evalCtx.Annotations)
	oid := tree.MustBeDOid(args[0])
	seqMetadata, ok := c.seqIDToMetadata[descpb.ID(oid.DInt)]
	if !ok {
		__antithesis_instrumentation__.Notify(567584)
		return nil, errors.Newf("sequence with ID %v not found in annotation", oid)
	} else {
		__antithesis_instrumentation__.Notify(567585)
	}
	__antithesis_instrumentation__.Notify(567583)
	return importNextValHelper(evalCtx, c, seqMetadata)
}

func importDefaultToDatabasePrimaryRegion(
	evalCtx *tree.EvalContext, _ tree.Datums,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(567586)
	regionConfig, err := evalCtx.Regions.CurrentDatabaseRegionConfig(evalCtx.Context)
	if err != nil {
		__antithesis_instrumentation__.Notify(567589)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(567590)
	}
	__antithesis_instrumentation__.Notify(567587)
	primaryRegion := regionConfig.PrimaryRegionString()
	if primaryRegion == "" {
		__antithesis_instrumentation__.Notify(567591)
		return nil, errors.New("primary region on the database being imported into is empty; failed" +
			" to evaluate expression using `default_to_database_primary_region` builtin")
	} else {
		__antithesis_instrumentation__.Notify(567592)
	}
	__antithesis_instrumentation__.Notify(567588)
	return tree.NewDString(primaryRegion), nil
}

func importNextValHelper(
	evalCtx *tree.EvalContext, c *CellInfoAnnotation, seqMetadata *SequenceMetadata,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(567593)
	seqOpts := seqMetadata.seqDesc.GetSequenceOpts()
	if c.seqChunkProvider == nil {
		__antithesis_instrumentation__.Notify(567596)
		return nil, errors.New("no sequence chunk provider configured for the import job")
	} else {
		__antithesis_instrumentation__.Notify(567597)
	}
	__antithesis_instrumentation__.Notify(567594)

	if seqMetadata.curChunk == nil || func() bool {
		__antithesis_instrumentation__.Notify(567598)
		return c.rowID == seqMetadata.curChunk.NextChunkStartRow == true
	}() == true {
		__antithesis_instrumentation__.Notify(567599)
		if err := c.seqChunkProvider.RequestChunk(evalCtx, c, seqMetadata); err != nil {
			__antithesis_instrumentation__.Notify(567600)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(567601)
		}
	} else {
		__antithesis_instrumentation__.Notify(567602)

		seqMetadata.curVal += seqOpts.Increment
	}
	__antithesis_instrumentation__.Notify(567595)
	return tree.NewDInt(tree.DInt(seqMetadata.curVal)), nil
}

type customFunc struct {
	visitorSideEffect func(annotations *tree.Annotations, fn *tree.FuncExpr) error
	override          *tree.FunctionDefinition
}

var useDefaultBuiltin *customFunc

var supportedImportFuncOverrides = map[string]*customFunc{

	"current_date":          useDefaultBuiltin,
	"current_timestamp":     useDefaultBuiltin,
	"localtimestamp":        useDefaultBuiltin,
	"now":                   useDefaultBuiltin,
	"statement_timestamp":   useDefaultBuiltin,
	"timeofday":             useDefaultBuiltin,
	"transaction_timestamp": useDefaultBuiltin,
	"unique_rowid": {
		visitorSideEffect: func(annot *tree.Annotations, _ *tree.FuncExpr) error {
			__antithesis_instrumentation__.Notify(567603)
			getCellInfoAnnotation(annot).uniqueRowIDTotal++
			return nil
		},
		override: makeBuiltinOverride(
			tree.FunDefs["unique_rowid"],
			tree.Overload{
				Types:      tree.ArgTypes{},
				ReturnType: tree.FixedReturnType(types.Int),
				Fn:         importUniqueRowID,
				Info:       "Returns a unique rowid based on row position and time",
				Volatility: tree.VolatilityVolatile,
			},
		),
	},
	"random": {
		visitorSideEffect: func(annot *tree.Annotations, _ *tree.FuncExpr) error {
			__antithesis_instrumentation__.Notify(567604)
			getCellInfoAnnotation(annot).randInstancePerRow++
			return nil
		},
		override: makeBuiltinOverride(
			tree.FunDefs["random"],
			tree.Overload{
				Types:      tree.ArgTypes{},
				ReturnType: tree.FixedReturnType(types.Float),
				Fn:         importRandom,
				Info:       "Returns a random number between 0 and 1 based on row position and time.",
				Volatility: tree.VolatilityVolatile,
			},
		),
	},
	"gen_random_uuid": {
		visitorSideEffect: func(annot *tree.Annotations, _ *tree.FuncExpr) error {
			__antithesis_instrumentation__.Notify(567605)
			getCellInfoAnnotation(annot).randInstancePerRow++
			return nil
		},
		override: makeBuiltinOverride(
			tree.FunDefs["gen_random_uuid"],
			tree.Overload{
				Types:      tree.ArgTypes{},
				ReturnType: tree.FixedReturnType(types.Uuid),
				Fn:         importGenUUID,
				Info: "Generates a random UUID based on row position and time, " +
					"and returns it as a value of UUID type.",
				Volatility: tree.VolatilityVolatile,
			},
		),
	},
	"nextval": {
		visitorSideEffect: func(annot *tree.Annotations, fn *tree.FuncExpr) error {
			__antithesis_instrumentation__.Notify(567606)

			seqIdentifier, err := seqexpr.GetSequenceFromFunc(fn)
			if err != nil {
				__antithesis_instrumentation__.Notify(567609)
				return err
			} else {
				__antithesis_instrumentation__.Notify(567610)
			}
			__antithesis_instrumentation__.Notify(567607)

			var sequenceMetadata *SequenceMetadata
			var ok bool
			if seqIdentifier.IsByID() {
				__antithesis_instrumentation__.Notify(567611)
				if sequenceMetadata, ok = getCellInfoAnnotation(annot).seqIDToMetadata[descpb.ID(seqIdentifier.SeqID)]; !ok {
					__antithesis_instrumentation__.Notify(567612)
					return errors.Newf("sequence with ID %s not found in annotation", seqIdentifier.SeqID)
				} else {
					__antithesis_instrumentation__.Notify(567613)
				}
			} else {
				__antithesis_instrumentation__.Notify(567614)
				if sequenceMetadata, ok = getCellInfoAnnotation(annot).seqNameToMetadata[seqIdentifier.SeqName]; !ok {
					__antithesis_instrumentation__.Notify(567615)
					return errors.Newf("sequence %s not found in annotation", seqIdentifier.SeqName)
				} else {
					__antithesis_instrumentation__.Notify(567616)
				}
			}
			__antithesis_instrumentation__.Notify(567608)
			sequenceMetadata.instancesPerRow++
			return nil
		},
		override: makeBuiltinOverride(
			tree.FunDefs["nextval"],
			tree.Overload{
				Types:      tree.ArgTypes{{builtins.SequenceNameArg, types.String}},
				ReturnType: tree.FixedReturnType(types.Int),
				Info:       "Advances the value of the sequence and returns the final value.",
				Fn:         importNextVal,
			},
			tree.Overload{
				Types:      tree.ArgTypes{{builtins.SequenceNameArg, types.RegClass}},
				ReturnType: tree.FixedReturnType(types.Int),
				Info:       "Advances the value of the sequence and returns the final value.",
				Fn:         importNextValByID,
			},
		),
	},
	"default_to_database_primary_region": {
		override: makeBuiltinOverride(
			tree.FunDefs["default_to_database_primary_region"],
			tree.Overload{
				Types:      tree.ArgTypes{{"val", types.String}},
				ReturnType: tree.FixedReturnType(types.String),
				Info:       "Returns the primary region of the database.",
				Fn:         importDefaultToDatabasePrimaryRegion,
			},
		),
	},
	"gateway_region": {
		override: makeBuiltinOverride(
			tree.FunDefs["gateway_region"],
			tree.Overload{
				Types:      tree.ArgTypes{},
				ReturnType: tree.FixedReturnType(types.String),
				Info:       "Returns the primary region of the database.",

				Fn: importDefaultToDatabasePrimaryRegion,
			},
		),
	},
}

func unsafeExpressionError(err error, msg string, expr string) error {
	__antithesis_instrumentation__.Notify(567617)
	return errors.Wrapf(err, "default expression %q is unsafe for import: %s", expr, msg)
}

type unsafeErrExpr struct {
	tree.TypedExpr
	err error
}

var _ tree.TypedExpr = &unsafeErrExpr{}

func (e *unsafeErrExpr) Eval(_ *tree.EvalContext) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(567618)
	return nil, e.err
}

type importDefaultExprVisitor struct {
	err         error
	ctx         context.Context
	annotations *tree.Annotations
	semaCtx     *tree.SemaContext

	volatility overrideVolatility
}

func (v *importDefaultExprVisitor) VisitPre(expr tree.Expr) (recurse bool, newExpr tree.Expr) {
	__antithesis_instrumentation__.Notify(567619)
	return v.err == nil, expr
}

func (v *importDefaultExprVisitor) VisitPost(expr tree.Expr) (newExpr tree.Expr) {
	__antithesis_instrumentation__.Notify(567620)
	if v.err != nil {
		__antithesis_instrumentation__.Notify(567627)
		return expr
	} else {
		__antithesis_instrumentation__.Notify(567628)
	}
	__antithesis_instrumentation__.Notify(567621)
	fn, ok := expr.(*tree.FuncExpr)
	if !ok || func() bool {
		__antithesis_instrumentation__.Notify(567629)
		return fn.ResolvedOverload().Volatility <= tree.VolatilityImmutable == true
	}() == true {
		__antithesis_instrumentation__.Notify(567630)

		return expr
	} else {
		__antithesis_instrumentation__.Notify(567631)
	}
	__antithesis_instrumentation__.Notify(567622)
	resolvedFnName := fn.Func.FunctionReference.(*tree.FunctionDefinition).Name
	custom, isSafe := supportedImportFuncOverrides[resolvedFnName]
	if !isSafe {
		__antithesis_instrumentation__.Notify(567632)
		v.err = errors.Newf(`function %s unsupported by IMPORT INTO`, resolvedFnName)
		return expr
	} else {
		__antithesis_instrumentation__.Notify(567633)
	}
	__antithesis_instrumentation__.Notify(567623)
	if custom == useDefaultBuiltin {
		__antithesis_instrumentation__.Notify(567634)

		return expr
	} else {
		__antithesis_instrumentation__.Notify(567635)
	}
	__antithesis_instrumentation__.Notify(567624)

	v.volatility = overrideVolatile
	if custom.visitorSideEffect != nil {
		__antithesis_instrumentation__.Notify(567636)
		err := custom.visitorSideEffect(v.annotations, fn)
		if err != nil {
			__antithesis_instrumentation__.Notify(567637)
			v.err = errors.Wrapf(err, "function %s failed when invoking side effect", resolvedFnName)
			return expr
		} else {
			__antithesis_instrumentation__.Notify(567638)
		}
	} else {
		__antithesis_instrumentation__.Notify(567639)
	}
	__antithesis_instrumentation__.Notify(567625)
	funcExpr := &tree.FuncExpr{
		Func:  tree.ResolvableFunctionReference{FunctionReference: custom.override},
		Type:  fn.Type,
		Exprs: fn.Exprs,
	}

	overrideExpr, err := funcExpr.TypeCheck(v.ctx, v.semaCtx, fn.ResolvedType())
	if err != nil {
		__antithesis_instrumentation__.Notify(567640)
		v.err = errors.Wrapf(err, "error overloading function")
	} else {
		__antithesis_instrumentation__.Notify(567641)
	}
	__antithesis_instrumentation__.Notify(567626)
	return overrideExpr
}

func sanitizeExprsForImport(
	ctx context.Context, evalCtx *tree.EvalContext, expr tree.Expr, targetType *types.T,
) (tree.TypedExpr, overrideVolatility, error) {
	__antithesis_instrumentation__.Notify(567642)
	semaCtx := tree.MakeSemaContext()

	typedExpr, err := schemaexpr.SanitizeVarFreeExpr(
		ctx, expr, targetType, "import_default", &semaCtx, tree.VolatilityImmutable)
	if err == nil {
		__antithesis_instrumentation__.Notify(567646)
		return typedExpr, overrideImmutable, nil
	} else {
		__antithesis_instrumentation__.Notify(567647)
	}
	__antithesis_instrumentation__.Notify(567643)

	typedExpr, err = tree.TypeCheck(ctx, expr, &semaCtx, targetType)
	if err != nil {
		__antithesis_instrumentation__.Notify(567648)
		return nil, overrideErrorTerm,
			unsafeExpressionError(err, "type checking error", expr.String())
	} else {
		__antithesis_instrumentation__.Notify(567649)
	}
	__antithesis_instrumentation__.Notify(567644)
	v := &importDefaultExprVisitor{annotations: evalCtx.Annotations}
	newExpr, _ := tree.WalkExpr(v, typedExpr)
	if v.err != nil {
		__antithesis_instrumentation__.Notify(567650)
		return nil, overrideErrorTerm,
			unsafeExpressionError(v.err, "expr walking error", expr.String())
	} else {
		__antithesis_instrumentation__.Notify(567651)
	}
	__antithesis_instrumentation__.Notify(567645)
	return newExpr.(tree.TypedExpr), v.volatility, nil
}
