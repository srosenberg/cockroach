package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/resolver"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/seqexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/builtins"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

func (p *planner) GetSerialSequenceNameFromColumn(
	ctx context.Context, tn *tree.TableName, columnName tree.Name,
) (*tree.TableName, error) {
	__antithesis_instrumentation__.Notify(617203)
	flags := tree.ObjectLookupFlagsWithRequiredTableKind(tree.ResolveRequireTableDesc)
	_, tableDesc, err := resolver.ResolveExistingTableObject(ctx, p, tn, flags)
	if err != nil {
		__antithesis_instrumentation__.Notify(617206)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(617207)
	}
	__antithesis_instrumentation__.Notify(617204)
	for _, col := range tableDesc.PublicColumns() {
		__antithesis_instrumentation__.Notify(617208)
		if col.ColName() == columnName {
			__antithesis_instrumentation__.Notify(617209)

			if col.NumUsesSequences() == 1 {
				__antithesis_instrumentation__.Notify(617211)
				seq, err := p.Descriptors().GetImmutableTableByID(
					ctx,
					p.txn,
					col.GetUsesSequenceID(0),
					tree.ObjectLookupFlagsWithRequiredTableKind(tree.ResolveRequireSequenceDesc),
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(617213)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(617214)
				}
				__antithesis_instrumentation__.Notify(617212)
				return p.getQualifiedTableName(ctx, seq)
			} else {
				__antithesis_instrumentation__.Notify(617215)
			}
			__antithesis_instrumentation__.Notify(617210)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(617216)
		}
	}
	__antithesis_instrumentation__.Notify(617205)
	return nil, colinfo.NewUndefinedColumnError(string(columnName))
}

func (p *planner) IncrementSequenceByID(ctx context.Context, seqID int64) (int64, error) {
	__antithesis_instrumentation__.Notify(617217)
	if p.EvalContext().TxnReadOnly {
		__antithesis_instrumentation__.Notify(617221)
		return 0, readOnlyError("nextval()")
	} else {
		__antithesis_instrumentation__.Notify(617222)
	}
	__antithesis_instrumentation__.Notify(617218)
	flags := tree.ObjectLookupFlagsWithRequiredTableKind(tree.ResolveRequireSequenceDesc)
	descriptor, err := p.Descriptors().GetImmutableTableByID(ctx, p.txn, descpb.ID(seqID), flags)
	if err != nil {
		__antithesis_instrumentation__.Notify(617223)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(617224)
	}
	__antithesis_instrumentation__.Notify(617219)
	if !descriptor.IsSequence() {
		__antithesis_instrumentation__.Notify(617225)
		seqName, err := p.getQualifiedTableName(ctx, descriptor)
		if err != nil {
			__antithesis_instrumentation__.Notify(617227)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(617228)
		}
		__antithesis_instrumentation__.Notify(617226)
		return 0, sqlerrors.NewWrongObjectTypeError(seqName, "sequence")
	} else {
		__antithesis_instrumentation__.Notify(617229)
	}
	__antithesis_instrumentation__.Notify(617220)
	return incrementSequenceHelper(ctx, p, descriptor)
}

func incrementSequenceHelper(
	ctx context.Context, p *planner, descriptor catalog.TableDescriptor,
) (int64, error) {
	__antithesis_instrumentation__.Notify(617230)
	if err := p.CheckPrivilege(ctx, descriptor, privilege.UPDATE); err != nil {
		__antithesis_instrumentation__.Notify(617235)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(617236)
	}
	__antithesis_instrumentation__.Notify(617231)

	seqOpts := descriptor.GetSequenceOpts()

	var val int64
	var err error
	if seqOpts.Virtual {
		__antithesis_instrumentation__.Notify(617237)
		rowid := builtins.GenerateUniqueInt(p.EvalContext().NodeID.SQLInstanceID())
		val = int64(rowid)
	} else {
		__antithesis_instrumentation__.Notify(617238)
		val, err = p.incrementSequenceUsingCache(ctx, descriptor)
	}
	__antithesis_instrumentation__.Notify(617232)
	if err != nil {
		__antithesis_instrumentation__.Notify(617239)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(617240)
	}
	__antithesis_instrumentation__.Notify(617233)

	p.sessionDataMutatorIterator.applyOnEachMutator(
		func(m sessionDataMutator) {
			__antithesis_instrumentation__.Notify(617241)
			m.RecordLatestSequenceVal(uint32(descriptor.GetID()), val)
		},
	)
	__antithesis_instrumentation__.Notify(617234)

	return val, nil
}

func (p *planner) incrementSequenceUsingCache(
	ctx context.Context, descriptor catalog.TableDescriptor,
) (int64, error) {
	__antithesis_instrumentation__.Notify(617242)
	seqOpts := descriptor.GetSequenceOpts()

	sequenceID := descriptor.GetID()
	createdInCurrentTxn := p.createdSequences.isCreatedSequence(sequenceID)
	var cacheSize int64
	if createdInCurrentTxn {
		__antithesis_instrumentation__.Notify(617246)
		cacheSize = 1
	} else {
		__antithesis_instrumentation__.Notify(617247)
		cacheSize = seqOpts.EffectiveCacheSize()
	}
	__antithesis_instrumentation__.Notify(617243)

	fetchNextValues := func() (currentValue, incrementAmount, sizeOfCache int64, err error) {
		__antithesis_instrumentation__.Notify(617248)
		seqValueKey := p.ExecCfg().Codec.SequenceKey(uint32(sequenceID))

		var endValue int64
		if createdInCurrentTxn {
			__antithesis_instrumentation__.Notify(617252)
			var res kv.KeyValue
			res, err = p.txn.Inc(ctx, seqValueKey, seqOpts.Increment*cacheSize)
			endValue = res.ValueInt()
		} else {
			__antithesis_instrumentation__.Notify(617253)
			endValue, err = kv.IncrementValRetryable(
				ctx, p.ExecCfg().DB, seqValueKey, seqOpts.Increment*cacheSize)
		}
		__antithesis_instrumentation__.Notify(617249)

		if err != nil {
			__antithesis_instrumentation__.Notify(617254)
			if errors.HasType(err, (*roachpb.IntegerOverflowError)(nil)) {
				__antithesis_instrumentation__.Notify(617256)
				return 0, 0, 0, boundsExceededError(descriptor)
			} else {
				__antithesis_instrumentation__.Notify(617257)
			}
			__antithesis_instrumentation__.Notify(617255)
			return 0, 0, 0, err
		} else {
			__antithesis_instrumentation__.Notify(617258)
		}
		__antithesis_instrumentation__.Notify(617250)

		if endValue > seqOpts.MaxValue || func() bool {
			__antithesis_instrumentation__.Notify(617259)
			return endValue < seqOpts.MinValue == true
		}() == true {
			__antithesis_instrumentation__.Notify(617260)

			if (seqOpts.Increment > 0 && func() bool {
				__antithesis_instrumentation__.Notify(617264)
				return endValue-seqOpts.Increment*cacheSize >= seqOpts.MaxValue == true
			}() == true) || func() bool {
				__antithesis_instrumentation__.Notify(617265)
				return (seqOpts.Increment < 0 && func() bool {
					__antithesis_instrumentation__.Notify(617266)
					return endValue-seqOpts.Increment*cacheSize <= seqOpts.MinValue == true
				}() == true) == true
			}() == true {
				__antithesis_instrumentation__.Notify(617267)
				return 0, 0, 0, boundsExceededError(descriptor)
			} else {
				__antithesis_instrumentation__.Notify(617268)
			}
			__antithesis_instrumentation__.Notify(617261)

			limit := seqOpts.MaxValue
			if seqOpts.Increment < 0 {
				__antithesis_instrumentation__.Notify(617269)
				limit = seqOpts.MinValue
			} else {
				__antithesis_instrumentation__.Notify(617270)
			}
			__antithesis_instrumentation__.Notify(617262)
			abs := func(i int64) int64 {
				__antithesis_instrumentation__.Notify(617271)
				if i < 0 {
					__antithesis_instrumentation__.Notify(617273)
					return -i
				} else {
					__antithesis_instrumentation__.Notify(617274)
				}
				__antithesis_instrumentation__.Notify(617272)
				return i
			}
			__antithesis_instrumentation__.Notify(617263)
			currentValue = endValue - seqOpts.Increment*(cacheSize-1)
			incrementAmount = seqOpts.Increment
			sizeOfCache = abs(limit-(endValue-seqOpts.Increment*cacheSize)) / abs(seqOpts.Increment)
			return currentValue, incrementAmount, sizeOfCache, nil
		} else {
			__antithesis_instrumentation__.Notify(617275)
		}
		__antithesis_instrumentation__.Notify(617251)

		return endValue - seqOpts.Increment*(cacheSize-1), seqOpts.Increment, cacheSize, nil
	}
	__antithesis_instrumentation__.Notify(617244)

	var val int64
	var err error
	if cacheSize == 1 {
		__antithesis_instrumentation__.Notify(617276)
		val, _, _, err = fetchNextValues()
		if err != nil {
			__antithesis_instrumentation__.Notify(617277)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(617278)
		}
	} else {
		__antithesis_instrumentation__.Notify(617279)
		val, err = p.GetOrInitSequenceCache().NextValue(uint32(sequenceID), uint32(descriptor.GetVersion()), fetchNextValues)
		if err != nil {
			__antithesis_instrumentation__.Notify(617280)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(617281)
		}
	}
	__antithesis_instrumentation__.Notify(617245)
	return val, nil
}

func boundsExceededError(descriptor catalog.TableDescriptor) error {
	__antithesis_instrumentation__.Notify(617282)
	seqOpts := descriptor.GetSequenceOpts()
	isAscending := seqOpts.Increment > 0

	var word string
	var value int64
	if isAscending {
		__antithesis_instrumentation__.Notify(617284)
		word = "maximum"
		value = seqOpts.MaxValue
	} else {
		__antithesis_instrumentation__.Notify(617285)
		word = "minimum"
		value = seqOpts.MinValue
	}
	__antithesis_instrumentation__.Notify(617283)
	name := descriptor.GetName()
	return pgerror.Newf(
		pgcode.SequenceGeneratorLimitExceeded,
		`reached %s value of sequence %q (%d)`, word,
		tree.ErrString((*tree.Name)(&name)), value)
}

func (p *planner) GetLatestValueInSessionForSequenceByID(
	ctx context.Context, seqID int64,
) (int64, error) {
	__antithesis_instrumentation__.Notify(617286)
	flags := tree.ObjectLookupFlagsWithRequiredTableKind(tree.ResolveRequireSequenceDesc)
	descriptor, err := p.Descriptors().GetImmutableTableByID(ctx, p.txn, descpb.ID(seqID), flags)
	if err != nil {
		__antithesis_instrumentation__.Notify(617290)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(617291)
	}
	__antithesis_instrumentation__.Notify(617287)
	seqName, err := p.getQualifiedTableName(ctx, descriptor)
	if err != nil {
		__antithesis_instrumentation__.Notify(617292)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(617293)
	}
	__antithesis_instrumentation__.Notify(617288)
	if !descriptor.IsSequence() {
		__antithesis_instrumentation__.Notify(617294)
		return 0, sqlerrors.NewWrongObjectTypeError(seqName, "sequence")
	} else {
		__antithesis_instrumentation__.Notify(617295)
	}
	__antithesis_instrumentation__.Notify(617289)
	return getLatestValueInSessionForSequenceHelper(p, descriptor, seqName)
}

func getLatestValueInSessionForSequenceHelper(
	p *planner, descriptor catalog.TableDescriptor, seqName *tree.TableName,
) (int64, error) {
	__antithesis_instrumentation__.Notify(617296)
	val, ok := p.SessionData().SequenceState.GetLastValueByID(uint32(descriptor.GetID()))
	if !ok {
		__antithesis_instrumentation__.Notify(617298)
		return 0, pgerror.Newf(
			pgcode.ObjectNotInPrerequisiteState,
			`currval of sequence %q is not yet defined in this session`, tree.ErrString(seqName))
	} else {
		__antithesis_instrumentation__.Notify(617299)
	}
	__antithesis_instrumentation__.Notify(617297)

	return val, nil
}

func (p *planner) SetSequenceValueByID(
	ctx context.Context, seqID uint32, newVal int64, isCalled bool,
) error {
	__antithesis_instrumentation__.Notify(617300)
	if p.EvalContext().TxnReadOnly {
		__antithesis_instrumentation__.Notify(617310)
		return readOnlyError("setval()")
	} else {
		__antithesis_instrumentation__.Notify(617311)
	}
	__antithesis_instrumentation__.Notify(617301)

	flags := tree.ObjectLookupFlagsWithRequiredTableKind(tree.ResolveRequireSequenceDesc)
	descriptor, err := p.Descriptors().GetImmutableTableByID(ctx, p.txn, descpb.ID(seqID), flags)
	if err != nil {
		__antithesis_instrumentation__.Notify(617312)
		return err
	} else {
		__antithesis_instrumentation__.Notify(617313)
	}
	__antithesis_instrumentation__.Notify(617302)
	seqName, err := p.getQualifiedTableName(ctx, descriptor)
	if err != nil {
		__antithesis_instrumentation__.Notify(617314)
		return err
	} else {
		__antithesis_instrumentation__.Notify(617315)
	}
	__antithesis_instrumentation__.Notify(617303)
	if !descriptor.IsSequence() {
		__antithesis_instrumentation__.Notify(617316)
		return sqlerrors.NewWrongObjectTypeError(seqName, "sequence")
	} else {
		__antithesis_instrumentation__.Notify(617317)
	}
	__antithesis_instrumentation__.Notify(617304)

	if err := p.CheckPrivilege(ctx, descriptor, privilege.UPDATE); err != nil {
		__antithesis_instrumentation__.Notify(617318)
		return err
	} else {
		__antithesis_instrumentation__.Notify(617319)
	}
	__antithesis_instrumentation__.Notify(617305)

	if descriptor.GetSequenceOpts().Virtual {
		__antithesis_instrumentation__.Notify(617320)

		return pgerror.Newf(
			pgcode.ObjectNotInPrerequisiteState,
			`cannot set the value of virtual sequence %q`, tree.ErrString(seqName))
	} else {
		__antithesis_instrumentation__.Notify(617321)
	}
	__antithesis_instrumentation__.Notify(617306)

	seqValueKey, newVal, err := MakeSequenceKeyVal(p.ExecCfg().Codec, descriptor, newVal, isCalled)
	if err != nil {
		__antithesis_instrumentation__.Notify(617322)
		return err
	} else {
		__antithesis_instrumentation__.Notify(617323)
	}
	__antithesis_instrumentation__.Notify(617307)

	createdInCurrentTxn := p.createdSequences.isCreatedSequence(descriptor.GetID())
	if createdInCurrentTxn {
		__antithesis_instrumentation__.Notify(617324)
		if err := p.txn.Put(ctx, seqValueKey, newVal); err != nil {
			__antithesis_instrumentation__.Notify(617325)
			return err
		} else {
			__antithesis_instrumentation__.Notify(617326)
		}
	} else {
		__antithesis_instrumentation__.Notify(617327)

		if err := p.ExecCfg().DB.Put(ctx, seqValueKey, newVal); err != nil {
			__antithesis_instrumentation__.Notify(617328)
			return err
		} else {
			__antithesis_instrumentation__.Notify(617329)
		}
	}
	__antithesis_instrumentation__.Notify(617308)

	p.sessionDataMutatorIterator.applyOnEachMutator(func(m sessionDataMutator) {
		__antithesis_instrumentation__.Notify(617330)
		m.initSequenceCache()
		if isCalled {
			__antithesis_instrumentation__.Notify(617331)
			m.RecordLatestSequenceVal(seqID, newVal)
		} else {
			__antithesis_instrumentation__.Notify(617332)
		}
	})
	__antithesis_instrumentation__.Notify(617309)
	return nil
}

func MakeSequenceKeyVal(
	codec keys.SQLCodec, sequence catalog.TableDescriptor, newVal int64, isCalled bool,
) ([]byte, int64, error) {
	__antithesis_instrumentation__.Notify(617333)
	opts := sequence.GetSequenceOpts()
	if newVal > opts.MaxValue || func() bool {
		__antithesis_instrumentation__.Notify(617336)
		return newVal < opts.MinValue == true
	}() == true {
		__antithesis_instrumentation__.Notify(617337)
		return nil, 0, pgerror.Newf(
			pgcode.NumericValueOutOfRange,
			`value %d is out of bounds for sequence "%s" (%d..%d)`,
			newVal, sequence.GetName(), opts.MinValue, opts.MaxValue,
		)
	} else {
		__antithesis_instrumentation__.Notify(617338)
	}
	__antithesis_instrumentation__.Notify(617334)
	if !isCalled {
		__antithesis_instrumentation__.Notify(617339)
		newVal = newVal - opts.Increment
	} else {
		__antithesis_instrumentation__.Notify(617340)
	}
	__antithesis_instrumentation__.Notify(617335)

	seqValueKey := codec.SequenceKey(uint32(sequence.GetID()))
	return seqValueKey, newVal, nil
}

func (p *planner) GetSequenceValue(
	ctx context.Context, codec keys.SQLCodec, desc catalog.TableDescriptor,
) (int64, error) {
	__antithesis_instrumentation__.Notify(617341)
	if desc.GetSequenceOpts() == nil {
		__antithesis_instrumentation__.Notify(617344)
		return 0, errors.New("descriptor is not a sequence")
	} else {
		__antithesis_instrumentation__.Notify(617345)
	}
	__antithesis_instrumentation__.Notify(617342)
	keyValue, err := p.txn.Get(ctx, codec.SequenceKey(uint32(desc.GetID())))
	if err != nil {
		__antithesis_instrumentation__.Notify(617346)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(617347)
	}
	__antithesis_instrumentation__.Notify(617343)
	return keyValue.ValueInt(), nil
}

func readOnlyError(s string) error {
	__antithesis_instrumentation__.Notify(617348)
	return pgerror.Newf(pgcode.ReadOnlySQLTransaction,
		"cannot execute %s in a read-only transaction", s)
}

func getSequenceIntegerBounds(
	integerType *types.T,
) (lowerIntBound int64, upperIntBound int64, err error) {
	__antithesis_instrumentation__.Notify(617349)
	switch integerType {
	case types.Int2:
		__antithesis_instrumentation__.Notify(617351)
		return math.MinInt16, math.MaxInt16, nil
	case types.Int4:
		__antithesis_instrumentation__.Notify(617352)
		return math.MinInt32, math.MaxInt32, nil
	case types.Int:
		__antithesis_instrumentation__.Notify(617353)
		return math.MinInt64, math.MaxInt64, nil
	default:
		__antithesis_instrumentation__.Notify(617354)
	}
	__antithesis_instrumentation__.Notify(617350)

	return 0, 0, pgerror.Newf(
		pgcode.InvalidParameterValue,
		"CREATE SEQUENCE option AS received type %s, must be integer",
		integerType,
	)
}

func setSequenceIntegerBounds(
	opts *descpb.TableDescriptor_SequenceOpts,
	integerType *types.T,
	isAscending bool,
	setMinValue bool,
	setMaxValue bool,
) error {
	__antithesis_instrumentation__.Notify(617355)
	var minValue int64 = math.MinInt64
	var maxValue int64 = math.MaxInt64

	if isAscending {
		__antithesis_instrumentation__.Notify(617359)
		minValue = 1

		switch integerType {
		case types.Int2:
			__antithesis_instrumentation__.Notify(617360)
			maxValue = math.MaxInt16
		case types.Int4:
			__antithesis_instrumentation__.Notify(617361)
			maxValue = math.MaxInt32
		case types.Int:
			__antithesis_instrumentation__.Notify(617362)

		default:
			__antithesis_instrumentation__.Notify(617363)
			return pgerror.Newf(
				pgcode.InvalidParameterValue,
				"CREATE SEQUENCE option AS received type %s, must be integer",
				integerType,
			)
		}
	} else {
		__antithesis_instrumentation__.Notify(617364)
		maxValue = -1
		switch integerType {
		case types.Int2:
			__antithesis_instrumentation__.Notify(617365)
			minValue = math.MinInt16
		case types.Int4:
			__antithesis_instrumentation__.Notify(617366)
			minValue = math.MinInt32
		case types.Int:
			__antithesis_instrumentation__.Notify(617367)

		default:
			__antithesis_instrumentation__.Notify(617368)
			return pgerror.Newf(
				pgcode.InvalidParameterValue,
				"CREATE SEQUENCE option AS received type %s, must be integer",
				integerType,
			)
		}
	}
	__antithesis_instrumentation__.Notify(617356)
	if setMinValue {
		__antithesis_instrumentation__.Notify(617369)
		opts.MinValue = minValue
	} else {
		__antithesis_instrumentation__.Notify(617370)
	}
	__antithesis_instrumentation__.Notify(617357)
	if setMaxValue {
		__antithesis_instrumentation__.Notify(617371)
		opts.MaxValue = maxValue
	} else {
		__antithesis_instrumentation__.Notify(617372)
	}
	__antithesis_instrumentation__.Notify(617358)
	return nil
}

func assignSequenceOptions(
	ctx context.Context,
	p *planner,
	opts *descpb.TableDescriptor_SequenceOpts,
	optsNode tree.SequenceOptions,
	setDefaults bool,
	sequenceID descpb.ID,
	sequenceParentID descpb.ID,
	existingType *types.T,
) error {
	__antithesis_instrumentation__.Notify(617373)

	wasAscending := opts.Increment > 0

	var integerType = types.Int

	for _, option := range optsNode {
		__antithesis_instrumentation__.Notify(617387)
		if option.Name == tree.SeqOptIncrement {
			__antithesis_instrumentation__.Notify(617388)
			opts.Increment = *option.IntVal
		} else {
			__antithesis_instrumentation__.Notify(617389)
			if option.Name == tree.SeqOptAs {
				__antithesis_instrumentation__.Notify(617390)
				integerType = option.AsIntegerType
				opts.AsIntegerType = integerType.SQLString()
			} else {
				__antithesis_instrumentation__.Notify(617391)
			}
		}
	}
	__antithesis_instrumentation__.Notify(617374)
	if opts.Increment == 0 {
		__antithesis_instrumentation__.Notify(617392)
		return pgerror.New(
			pgcode.InvalidParameterValue, "INCREMENT must not be zero")
	} else {
		__antithesis_instrumentation__.Notify(617393)
	}
	__antithesis_instrumentation__.Notify(617375)
	isAscending := opts.Increment > 0

	if setDefaults {
		__antithesis_instrumentation__.Notify(617394)
		if isAscending {
			__antithesis_instrumentation__.Notify(617396)
			opts.MinValue = 1
			opts.MaxValue = math.MaxInt64
			opts.Start = opts.MinValue
		} else {
			__antithesis_instrumentation__.Notify(617397)
			opts.MinValue = math.MinInt64
			opts.MaxValue = -1
			opts.Start = opts.MaxValue
		}
		__antithesis_instrumentation__.Notify(617395)

		opts.CacheSize = 1
	} else {
		__antithesis_instrumentation__.Notify(617398)
	}
	__antithesis_instrumentation__.Notify(617376)

	lowerIntBound, upperIntBound, err := getSequenceIntegerBounds(integerType)
	if err != nil {
		__antithesis_instrumentation__.Notify(617399)
		return err
	} else {
		__antithesis_instrumentation__.Notify(617400)
	}
	__antithesis_instrumentation__.Notify(617377)

	if opts.AsIntegerType != "" {
		__antithesis_instrumentation__.Notify(617401)

		setMinValue := setDefaults
		setMaxValue := setDefaults
		if !setDefaults && func() bool {
			__antithesis_instrumentation__.Notify(617403)
			return existingType != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(617404)
			existingLowerIntBound, existingUpperIntBound, err := getSequenceIntegerBounds(existingType)
			if err != nil {
				__antithesis_instrumentation__.Notify(617407)
				return err
			} else {
				__antithesis_instrumentation__.Notify(617408)
			}
			__antithesis_instrumentation__.Notify(617405)
			if (wasAscending && func() bool {
				__antithesis_instrumentation__.Notify(617409)
				return opts.MinValue == 1 == true
			}() == true) || func() bool {
				__antithesis_instrumentation__.Notify(617410)
				return (!wasAscending && func() bool {
					__antithesis_instrumentation__.Notify(617411)
					return opts.MinValue == existingLowerIntBound == true
				}() == true) == true
			}() == true {
				__antithesis_instrumentation__.Notify(617412)
				setMinValue = true
			} else {
				__antithesis_instrumentation__.Notify(617413)
			}
			__antithesis_instrumentation__.Notify(617406)
			if (wasAscending && func() bool {
				__antithesis_instrumentation__.Notify(617414)
				return opts.MaxValue == existingUpperIntBound == true
			}() == true) || func() bool {
				__antithesis_instrumentation__.Notify(617415)
				return (!wasAscending && func() bool {
					__antithesis_instrumentation__.Notify(617416)
					return opts.MaxValue == -1 == true
				}() == true) == true
			}() == true {
				__antithesis_instrumentation__.Notify(617417)
				setMaxValue = true
			} else {
				__antithesis_instrumentation__.Notify(617418)
			}
		} else {
			__antithesis_instrumentation__.Notify(617419)
		}
		__antithesis_instrumentation__.Notify(617402)

		if err := setSequenceIntegerBounds(
			opts,
			integerType,
			isAscending,
			setMinValue,
			setMaxValue,
		); err != nil {
			__antithesis_instrumentation__.Notify(617420)
			return err
		} else {
			__antithesis_instrumentation__.Notify(617421)
		}
	} else {
		__antithesis_instrumentation__.Notify(617422)
	}
	__antithesis_instrumentation__.Notify(617378)

	optionsSeen := map[string]bool{}
	for _, option := range optsNode {
		__antithesis_instrumentation__.Notify(617423)

		_, seenBefore := optionsSeen[option.Name]
		if seenBefore {
			__antithesis_instrumentation__.Notify(617425)
			return pgerror.New(pgcode.Syntax, "conflicting or redundant options")
		} else {
			__antithesis_instrumentation__.Notify(617426)
		}
		__antithesis_instrumentation__.Notify(617424)
		optionsSeen[option.Name] = true

		switch option.Name {
		case tree.SeqOptCycle:
			__antithesis_instrumentation__.Notify(617427)
			return unimplemented.NewWithIssue(20961,
				"CYCLE option is not supported")
		case tree.SeqOptNoCycle:
			__antithesis_instrumentation__.Notify(617428)

		case tree.SeqOptCache:
			__antithesis_instrumentation__.Notify(617429)
			if v := *option.IntVal; v >= 1 {
				__antithesis_instrumentation__.Notify(617438)
				opts.CacheSize = v
			} else {
				__antithesis_instrumentation__.Notify(617439)
				return pgerror.Newf(pgcode.InvalidParameterValue,
					"CACHE (%d) must be greater than zero", v)
			}
		case tree.SeqOptIncrement:
			__antithesis_instrumentation__.Notify(617430)

		case tree.SeqOptMinValue:
			__antithesis_instrumentation__.Notify(617431)

			if option.IntVal != nil {
				__antithesis_instrumentation__.Notify(617440)
				opts.MinValue = *option.IntVal
			} else {
				__antithesis_instrumentation__.Notify(617441)
			}
		case tree.SeqOptMaxValue:
			__antithesis_instrumentation__.Notify(617432)

			if option.IntVal != nil {
				__antithesis_instrumentation__.Notify(617442)
				opts.MaxValue = *option.IntVal
			} else {
				__antithesis_instrumentation__.Notify(617443)
			}
		case tree.SeqOptStart:
			__antithesis_instrumentation__.Notify(617433)
			opts.Start = *option.IntVal
		case tree.SeqOptVirtual:
			__antithesis_instrumentation__.Notify(617434)
			opts.Virtual = true
		case tree.SeqOptOwnedBy:
			__antithesis_instrumentation__.Notify(617435)
			if p == nil {
				__antithesis_instrumentation__.Notify(617444)
				return pgerror.Newf(pgcode.Internal,
					"Trying to add/remove Sequence Owner outside of context of a planner")
			} else {
				__antithesis_instrumentation__.Notify(617445)
			}
			__antithesis_instrumentation__.Notify(617436)

			if option.ColumnItemVal == nil {
				__antithesis_instrumentation__.Notify(617446)
				if err := removeSequenceOwnerIfExists(ctx, p, sequenceID, opts); err != nil {
					__antithesis_instrumentation__.Notify(617447)
					return err
				} else {
					__antithesis_instrumentation__.Notify(617448)
				}
			} else {
				__antithesis_instrumentation__.Notify(617449)

				tableDesc, col, err := resolveColumnItemToDescriptors(
					ctx, p, option.ColumnItemVal,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(617452)
					return err
				} else {
					__antithesis_instrumentation__.Notify(617453)
				}
				__antithesis_instrumentation__.Notify(617450)
				if tableDesc.ParentID != sequenceParentID && func() bool {
					__antithesis_instrumentation__.Notify(617454)
					return !allowCrossDatabaseSeqOwner.Get(&p.execCfg.Settings.SV) == true
				}() == true {
					__antithesis_instrumentation__.Notify(617455)
					return errors.WithHintf(
						pgerror.Newf(pgcode.FeatureNotSupported,
							"OWNED BY cannot refer to other databases; (see the '%s' cluster setting)",
							allowCrossDatabaseSeqOwnerSetting),
						crossDBReferenceDeprecationHint(),
					)
				} else {
					__antithesis_instrumentation__.Notify(617456)
				}
				__antithesis_instrumentation__.Notify(617451)

				if opts.SequenceOwner.OwnerTableID != tableDesc.ID || func() bool {
					__antithesis_instrumentation__.Notify(617457)
					return opts.SequenceOwner.OwnerColumnID != col.GetID() == true
				}() == true {
					__antithesis_instrumentation__.Notify(617458)
					if err := removeSequenceOwnerIfExists(ctx, p, sequenceID, opts); err != nil {
						__antithesis_instrumentation__.Notify(617460)
						return err
					} else {
						__antithesis_instrumentation__.Notify(617461)
					}
					__antithesis_instrumentation__.Notify(617459)
					err := addSequenceOwner(ctx, p, option.ColumnItemVal, sequenceID, opts)
					if err != nil {
						__antithesis_instrumentation__.Notify(617462)
						return err
					} else {
						__antithesis_instrumentation__.Notify(617463)
					}
				} else {
					__antithesis_instrumentation__.Notify(617464)
				}
			}
		default:
			__antithesis_instrumentation__.Notify(617437)
		}
	}
	__antithesis_instrumentation__.Notify(617379)

	if setDefaults || func() bool {
		__antithesis_instrumentation__.Notify(617465)
		return (wasAscending && func() bool {
			__antithesis_instrumentation__.Notify(617466)
			return opts.Start == 1 == true
		}() == true) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(617467)
		return (!wasAscending && func() bool {
			__antithesis_instrumentation__.Notify(617468)
			return opts.Start == -1 == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(617469)

		if _, startSeen := optionsSeen[tree.SeqOptStart]; !startSeen {
			__antithesis_instrumentation__.Notify(617470)
			if opts.Increment > 0 {
				__antithesis_instrumentation__.Notify(617471)
				opts.Start = opts.MinValue
			} else {
				__antithesis_instrumentation__.Notify(617472)
				opts.Start = opts.MaxValue
			}
		} else {
			__antithesis_instrumentation__.Notify(617473)
		}
	} else {
		__antithesis_instrumentation__.Notify(617474)
	}
	__antithesis_instrumentation__.Notify(617380)

	if opts.MinValue < lowerIntBound {
		__antithesis_instrumentation__.Notify(617475)
		return pgerror.Newf(
			pgcode.InvalidParameterValue,
			"MINVALUE (%d) must be greater than (%d) for type %s",
			opts.MinValue,
			lowerIntBound,
			integerType.SQLString(),
		)
	} else {
		__antithesis_instrumentation__.Notify(617476)
	}
	__antithesis_instrumentation__.Notify(617381)
	if opts.MaxValue < lowerIntBound {
		__antithesis_instrumentation__.Notify(617477)
		return pgerror.Newf(
			pgcode.InvalidParameterValue,
			"MAXVALUE (%d) must be greater than (%d) for type %s",
			opts.MaxValue,
			lowerIntBound,
			integerType.SQLString(),
		)
	} else {
		__antithesis_instrumentation__.Notify(617478)
	}
	__antithesis_instrumentation__.Notify(617382)
	if opts.MinValue > upperIntBound {
		__antithesis_instrumentation__.Notify(617479)
		return pgerror.Newf(
			pgcode.InvalidParameterValue,
			"MINVALUE (%d) must be less than (%d) for type %s",
			opts.MinValue,
			upperIntBound,
			integerType.SQLString(),
		)
	} else {
		__antithesis_instrumentation__.Notify(617480)
	}
	__antithesis_instrumentation__.Notify(617383)
	if opts.MaxValue > upperIntBound {
		__antithesis_instrumentation__.Notify(617481)
		return pgerror.Newf(
			pgcode.InvalidParameterValue,
			"MAXVALUE (%d) must be less than (%d) for type %s",
			opts.MaxValue,
			upperIntBound,
			integerType.SQLString(),
		)
	} else {
		__antithesis_instrumentation__.Notify(617482)
	}
	__antithesis_instrumentation__.Notify(617384)
	if opts.Start > opts.MaxValue {
		__antithesis_instrumentation__.Notify(617483)
		return pgerror.Newf(
			pgcode.InvalidParameterValue,
			"START value (%d) cannot be greater than MAXVALUE (%d)",
			opts.Start,
			opts.MaxValue,
		)
	} else {
		__antithesis_instrumentation__.Notify(617484)
	}
	__antithesis_instrumentation__.Notify(617385)
	if opts.Start < opts.MinValue {
		__antithesis_instrumentation__.Notify(617485)
		return pgerror.Newf(
			pgcode.InvalidParameterValue,
			"START value (%d) cannot be less than MINVALUE (%d)",
			opts.Start,
			opts.MinValue,
		)
	} else {
		__antithesis_instrumentation__.Notify(617486)
	}
	__antithesis_instrumentation__.Notify(617386)

	return nil
}

func removeSequenceOwnerIfExists(
	ctx context.Context, p *planner, sequenceID descpb.ID, opts *descpb.TableDescriptor_SequenceOpts,
) error {
	__antithesis_instrumentation__.Notify(617487)
	if !opts.HasOwner() {
		__antithesis_instrumentation__.Notify(617495)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(617496)
	}
	__antithesis_instrumentation__.Notify(617488)
	tableDesc, err := p.Descriptors().GetMutableTableVersionByID(ctx, opts.SequenceOwner.OwnerTableID, p.txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(617497)

		if errors.Is(err, catalog.ErrDescriptorDropped) || func() bool {
			__antithesis_instrumentation__.Notify(617499)
			return pgerror.GetPGCode(err) == pgcode.UndefinedTable == true
		}() == true {
			__antithesis_instrumentation__.Notify(617500)
			log.Eventf(ctx, "swallowing error during sequence ownership unlinking: %s", err.Error())
			return nil
		} else {
			__antithesis_instrumentation__.Notify(617501)
		}
		__antithesis_instrumentation__.Notify(617498)
		return err
	} else {
		__antithesis_instrumentation__.Notify(617502)
	}
	__antithesis_instrumentation__.Notify(617489)

	if tableDesc.Dropped() {
		__antithesis_instrumentation__.Notify(617503)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(617504)
	}
	__antithesis_instrumentation__.Notify(617490)
	col, err := tableDesc.FindColumnWithID(opts.SequenceOwner.OwnerColumnID)
	if err != nil {
		__antithesis_instrumentation__.Notify(617505)
		return err
	} else {
		__antithesis_instrumentation__.Notify(617506)
	}
	__antithesis_instrumentation__.Notify(617491)

	newOwnsSequenceIDs := make([]descpb.ID, 0, col.NumOwnsSequences())
	for i := 0; i < col.NumOwnsSequences(); i++ {
		__antithesis_instrumentation__.Notify(617507)
		id := col.GetOwnsSequenceID(i)
		if id != sequenceID {
			__antithesis_instrumentation__.Notify(617508)
			newOwnsSequenceIDs = append(newOwnsSequenceIDs, id)
		} else {
			__antithesis_instrumentation__.Notify(617509)
		}
	}
	__antithesis_instrumentation__.Notify(617492)
	if len(newOwnsSequenceIDs) == col.NumOwnsSequences() {
		__antithesis_instrumentation__.Notify(617510)
		return errors.AssertionFailedf("couldn't find reference from column to this sequence")
	} else {
		__antithesis_instrumentation__.Notify(617511)
	}
	__antithesis_instrumentation__.Notify(617493)
	col.ColumnDesc().OwnsSequenceIds = newOwnsSequenceIDs
	if err := p.writeSchemaChange(
		ctx, tableDesc, descpb.InvalidMutationID,
		fmt.Sprintf("removing sequence owner %s(%d) for sequence %d",
			tableDesc.Name, tableDesc.ID, sequenceID,
		),
	); err != nil {
		__antithesis_instrumentation__.Notify(617512)
		return err
	} else {
		__antithesis_instrumentation__.Notify(617513)
	}
	__antithesis_instrumentation__.Notify(617494)

	opts.SequenceOwner.Reset()
	return nil
}

func resolveColumnItemToDescriptors(
	ctx context.Context, p *planner, columnItem *tree.ColumnItem,
) (*tabledesc.Mutable, catalog.Column, error) {
	__antithesis_instrumentation__.Notify(617514)
	if columnItem.TableName == nil {
		__antithesis_instrumentation__.Notify(617518)
		err := pgerror.New(pgcode.Syntax, "invalid OWNED BY option")
		return nil, nil, errors.WithHint(err, "Specify OWNED BY table.column or OWNED BY NONE.")
	} else {
		__antithesis_instrumentation__.Notify(617519)
	}
	__antithesis_instrumentation__.Notify(617515)
	tableName := columnItem.TableName.ToTableName()
	_, tableDesc, err := p.ResolveMutableTableDescriptor(ctx, &tableName, true, tree.ResolveRequireTableDesc)
	if err != nil {
		__antithesis_instrumentation__.Notify(617520)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(617521)
	}
	__antithesis_instrumentation__.Notify(617516)
	col, err := tableDesc.FindColumnWithName(columnItem.ColumnName)
	if err != nil {
		__antithesis_instrumentation__.Notify(617522)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(617523)
	}
	__antithesis_instrumentation__.Notify(617517)
	return tableDesc, col, nil
}

func addSequenceOwner(
	ctx context.Context,
	p *planner,
	columnItemVal *tree.ColumnItem,
	sequenceID descpb.ID,
	opts *descpb.TableDescriptor_SequenceOpts,
) error {
	__antithesis_instrumentation__.Notify(617524)
	tableDesc, col, err := resolveColumnItemToDescriptors(ctx, p, columnItemVal)
	if err != nil {
		__antithesis_instrumentation__.Notify(617526)
		return err
	} else {
		__antithesis_instrumentation__.Notify(617527)
	}
	__antithesis_instrumentation__.Notify(617525)

	col.ColumnDesc().OwnsSequenceIds = append(col.ColumnDesc().OwnsSequenceIds, sequenceID)

	opts.SequenceOwner.OwnerColumnID = col.GetID()
	opts.SequenceOwner.OwnerTableID = tableDesc.GetID()
	return p.writeSchemaChange(
		ctx, tableDesc, descpb.InvalidMutationID, fmt.Sprintf(
			"adding sequence owner %s(%d) for sequence %d",
			tableDesc.Name, tableDesc.ID, sequenceID),
	)
}

func maybeAddSequenceDependencies(
	ctx context.Context,
	st *cluster.Settings,
	sc resolver.SchemaResolver,
	tableDesc catalog.TableDescriptor,
	col *descpb.ColumnDescriptor,
	expr tree.TypedExpr,
	backrefs map[descpb.ID]*tabledesc.Mutable,
) ([]*tabledesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(617528)
	seqIdentifiers, err := seqexpr.GetUsedSequences(expr)
	if err != nil {
		__antithesis_instrumentation__.Notify(617532)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(617533)
	}
	__antithesis_instrumentation__.Notify(617529)

	var seqDescs []*tabledesc.Mutable
	seqNameToID := make(map[string]int64)
	for _, seqIdentifier := range seqIdentifiers {
		__antithesis_instrumentation__.Notify(617534)
		seqDesc, err := GetSequenceDescFromIdentifier(ctx, sc, seqIdentifier)
		if err != nil {
			__antithesis_instrumentation__.Notify(617540)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(617541)
		}
		__antithesis_instrumentation__.Notify(617535)

		if seqDesc.GetParentID() != tableDesc.GetParentID() && func() bool {
			__antithesis_instrumentation__.Notify(617542)
			return !allowCrossDatabaseSeqReferences.Get(&st.SV) == true
		}() == true {
			__antithesis_instrumentation__.Notify(617543)
			return nil, errors.WithHintf(
				pgerror.Newf(pgcode.FeatureNotSupported,
					"sequence references cannot come from other databases; (see the '%s' cluster setting)",
					allowCrossDatabaseSeqReferencesSetting),
				crossDBReferenceDeprecationHint(),
			)

		} else {
			__antithesis_instrumentation__.Notify(617544)
		}
		__antithesis_instrumentation__.Notify(617536)
		seqNameToID[seqIdentifier.SeqName] = int64(seqDesc.ID)

		if prev, ok := backrefs[seqDesc.ID]; ok {
			__antithesis_instrumentation__.Notify(617545)
			seqDesc = prev
		} else {
			__antithesis_instrumentation__.Notify(617546)
		}

		{
			__antithesis_instrumentation__.Notify(617547)
			var found bool
			for _, seqID := range col.UsesSequenceIds {
				__antithesis_instrumentation__.Notify(617549)
				if seqID == seqDesc.ID {
					__antithesis_instrumentation__.Notify(617550)
					found = true
					break
				} else {
					__antithesis_instrumentation__.Notify(617551)
				}
			}
			__antithesis_instrumentation__.Notify(617548)
			if !found {
				__antithesis_instrumentation__.Notify(617552)
				col.UsesSequenceIds = append(col.UsesSequenceIds, seqDesc.ID)
			} else {
				__antithesis_instrumentation__.Notify(617553)
			}
		}
		__antithesis_instrumentation__.Notify(617537)
		refIdx := -1
		for i, reference := range seqDesc.DependedOnBy {
			__antithesis_instrumentation__.Notify(617554)
			if reference.ID == tableDesc.GetID() {
				__antithesis_instrumentation__.Notify(617555)
				refIdx = i
			} else {
				__antithesis_instrumentation__.Notify(617556)
			}
		}
		__antithesis_instrumentation__.Notify(617538)
		if refIdx == -1 {
			__antithesis_instrumentation__.Notify(617557)
			seqDesc.DependedOnBy = append(seqDesc.DependedOnBy, descpb.TableDescriptor_Reference{
				ID:        tableDesc.GetID(),
				ColumnIDs: []descpb.ColumnID{col.ID},
				ByID:      true,
			})
		} else {
			__antithesis_instrumentation__.Notify(617558)
			ref := &seqDesc.DependedOnBy[refIdx]
			var found bool
			for _, colID := range ref.ColumnIDs {
				__antithesis_instrumentation__.Notify(617560)
				if colID == col.ID {
					__antithesis_instrumentation__.Notify(617561)
					found = true
					break
				} else {
					__antithesis_instrumentation__.Notify(617562)
				}
			}
			__antithesis_instrumentation__.Notify(617559)
			if !found {
				__antithesis_instrumentation__.Notify(617563)
				ref.ColumnIDs = append(ref.ColumnIDs, col.ID)
			} else {
				__antithesis_instrumentation__.Notify(617564)
			}
		}
		__antithesis_instrumentation__.Notify(617539)
		seqDescs = append(seqDescs, seqDesc)
	}
	__antithesis_instrumentation__.Notify(617530)

	if len(seqIdentifiers) > 0 {
		__antithesis_instrumentation__.Notify(617565)
		newExpr, err := seqexpr.ReplaceSequenceNamesWithIDs(expr, seqNameToID)
		if err != nil {
			__antithesis_instrumentation__.Notify(617567)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(617568)
		}
		__antithesis_instrumentation__.Notify(617566)
		s := tree.Serialize(newExpr)
		col.DefaultExpr = &s
	} else {
		__antithesis_instrumentation__.Notify(617569)
	}
	__antithesis_instrumentation__.Notify(617531)

	return seqDescs, nil
}

func GetSequenceDescFromIdentifier(
	ctx context.Context, sc resolver.SchemaResolver, seqIdentifier seqexpr.SeqIdentifier,
) (*tabledesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(617570)
	var tn tree.TableName
	if seqIdentifier.IsByID() {
		__antithesis_instrumentation__.Notify(617573)
		name, err := sc.GetQualifiedTableNameByID(ctx, seqIdentifier.SeqID, tree.ResolveRequireSequenceDesc)
		if err != nil {
			__antithesis_instrumentation__.Notify(617575)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(617576)
		}
		__antithesis_instrumentation__.Notify(617574)
		tn = *name
	} else {
		__antithesis_instrumentation__.Notify(617577)
		parsedSeqName, err := parser.ParseTableName(seqIdentifier.SeqName)
		if err != nil {
			__antithesis_instrumentation__.Notify(617579)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(617580)
		}
		__antithesis_instrumentation__.Notify(617578)
		tn = parsedSeqName.ToTableName()
	}
	__antithesis_instrumentation__.Notify(617571)

	var seqDesc *tabledesc.Mutable
	var err error
	p, ok := sc.(*planner)
	if ok {
		__antithesis_instrumentation__.Notify(617581)
		_, seqDesc, err = p.ResolveMutableTableDescriptor(ctx, &tn, true, tree.ResolveRequireSequenceDesc)
		if err != nil {
			__antithesis_instrumentation__.Notify(617582)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(617583)
		}
	} else {
		__antithesis_instrumentation__.Notify(617584)

		_, seqDesc, err = resolver.ResolveMutableExistingTableObject(ctx, sc, &tn, true, tree.ResolveRequireSequenceDesc)
		if err != nil {
			__antithesis_instrumentation__.Notify(617585)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(617586)
		}
	}
	__antithesis_instrumentation__.Notify(617572)
	return seqDesc, nil
}

func (p *planner) dropSequencesOwnedByCol(
	ctx context.Context, col catalog.Column, queueJob bool, behavior tree.DropBehavior,
) error {
	__antithesis_instrumentation__.Notify(617587)

	colOwnsSequenceIDs := make([]descpb.ID, col.NumOwnsSequences())
	for i := 0; i < col.NumOwnsSequences(); i++ {
		__antithesis_instrumentation__.Notify(617590)
		colOwnsSequenceIDs[i] = col.GetOwnsSequenceID(i)
	}
	__antithesis_instrumentation__.Notify(617588)

	for _, sequenceID := range colOwnsSequenceIDs {
		__antithesis_instrumentation__.Notify(617591)
		seqDesc, err := p.Descriptors().GetMutableTableVersionByID(ctx, sequenceID, p.txn)

		if err != nil {
			__antithesis_instrumentation__.Notify(617594)
			if errors.Is(err, catalog.ErrDescriptorDropped) || func() bool {
				__antithesis_instrumentation__.Notify(617596)
				return pgerror.GetPGCode(err) == pgcode.UndefinedTable == true
			}() == true {
				__antithesis_instrumentation__.Notify(617597)
				log.Eventf(ctx, "swallowing error dropping owned sequences: %s", err.Error())
				continue
			} else {
				__antithesis_instrumentation__.Notify(617598)
			}
			__antithesis_instrumentation__.Notify(617595)
			return err
		} else {
			__antithesis_instrumentation__.Notify(617599)
		}
		__antithesis_instrumentation__.Notify(617592)

		if seqDesc.Dropped() {
			__antithesis_instrumentation__.Notify(617600)
			continue
		} else {
			__antithesis_instrumentation__.Notify(617601)
		}
		__antithesis_instrumentation__.Notify(617593)
		jobDesc := fmt.Sprintf("removing sequence %q dependent on column %q which is being dropped",
			seqDesc.Name, col.ColName())

		if err := p.dropSequenceImpl(
			ctx, seqDesc, queueJob, jobDesc, behavior,
		); err != nil {
			__antithesis_instrumentation__.Notify(617602)
			return err
		} else {
			__antithesis_instrumentation__.Notify(617603)
		}
	}
	__antithesis_instrumentation__.Notify(617589)
	return nil
}

func (p *planner) removeSequenceDependencies(
	ctx context.Context, tableDesc *tabledesc.Mutable, col catalog.Column,
) error {
	__antithesis_instrumentation__.Notify(617604)
	for i := 0; i < col.NumUsesSequences(); i++ {
		__antithesis_instrumentation__.Notify(617606)
		sequenceID := col.GetUsesSequenceID(i)

		seqDesc, err := p.Descriptors().GetMutableTableVersionByID(ctx, sequenceID, p.txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(617612)
			return err
		} else {
			__antithesis_instrumentation__.Notify(617613)
		}
		__antithesis_instrumentation__.Notify(617607)

		if seqDesc.Dropped() {
			__antithesis_instrumentation__.Notify(617614)
			continue
		} else {
			__antithesis_instrumentation__.Notify(617615)
		}
		__antithesis_instrumentation__.Notify(617608)

		refTableIdx := -1
		refColIdx := -1
	found:
		for i, reference := range seqDesc.DependedOnBy {
			__antithesis_instrumentation__.Notify(617616)
			if reference.ID == tableDesc.ID {
				__antithesis_instrumentation__.Notify(617617)
				refTableIdx = i
				for j, colRefID := range seqDesc.DependedOnBy[i].ColumnIDs {
					__antithesis_instrumentation__.Notify(617618)
					if colRefID == col.GetID() {
						__antithesis_instrumentation__.Notify(617620)
						refColIdx = j
						break found
					} else {
						__antithesis_instrumentation__.Notify(617621)
					}
					__antithesis_instrumentation__.Notify(617619)

					if colRefID == 0 {
						__antithesis_instrumentation__.Notify(617622)
						refColIdx = j
					} else {
						__antithesis_instrumentation__.Notify(617623)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(617624)
			}
		}
		__antithesis_instrumentation__.Notify(617609)
		if refColIdx == -1 {
			__antithesis_instrumentation__.Notify(617625)
			return errors.AssertionFailedf("couldn't find reference from sequence to this column")
		} else {
			__antithesis_instrumentation__.Notify(617626)
		}
		__antithesis_instrumentation__.Notify(617610)

		seqDesc.DependedOnBy[refTableIdx].ColumnIDs = append(
			seqDesc.DependedOnBy[refTableIdx].ColumnIDs[:refColIdx],
			seqDesc.DependedOnBy[refTableIdx].ColumnIDs[refColIdx+1:]...)

		if len(seqDesc.DependedOnBy[refTableIdx].ColumnIDs) == 0 {
			__antithesis_instrumentation__.Notify(617627)
			seqDesc.DependedOnBy = append(
				seqDesc.DependedOnBy[:refTableIdx],
				seqDesc.DependedOnBy[refTableIdx+1:]...)
		} else {
			__antithesis_instrumentation__.Notify(617628)
		}
		__antithesis_instrumentation__.Notify(617611)

		jobDesc := fmt.Sprintf("removing sequence %q dependent on column %q which is being dropped",
			seqDesc.Name, col.ColName())
		if err := p.writeSchemaChange(
			ctx, seqDesc, descpb.InvalidMutationID, jobDesc,
		); err != nil {
			__antithesis_instrumentation__.Notify(617629)
			return err
		} else {
			__antithesis_instrumentation__.Notify(617630)
		}
	}
	__antithesis_instrumentation__.Notify(617605)

	col.ColumnDesc().UsesSequenceIds = []descpb.ID{}
	return nil
}
