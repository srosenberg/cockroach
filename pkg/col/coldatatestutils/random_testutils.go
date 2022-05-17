package coldatatestutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math/rand"
	"time"

	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecerror"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecop"
	"github.com/cockroachdb/cockroach/pkg/sql/colmem"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/randgen"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/cockroach/pkg/util/json"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

const maxVarLen = 64

var locations []*time.Location

func init() {

	for _, locationName := range []string{
		"Africa/Addis_Ababa",
		"America/Anchorage",
		"Antarctica/Davis",
		"Asia/Ashkhabad",
		"Australia/Sydney",
		"Europe/Minsk",
		"Pacific/Palau",
	} {
		loc, err := timeutil.LoadLocation(locationName)
		if err == nil {
			locations = append(locations, loc)
		}
	}
}

type RandomVecArgs struct {
	Rand *rand.Rand

	Vec coldata.Vec

	N int

	NullProbability float64

	BytesFixedLength int

	IntRange int

	ZeroProhibited bool
}

func RandomVec(args RandomVecArgs) {
	__antithesis_instrumentation__.Notify(54710)
	switch args.Vec.CanonicalTypeFamily() {
	case types.BoolFamily:
		__antithesis_instrumentation__.Notify(54713)
		bools := args.Vec.Bool()
		for i := 0; i < args.N; i++ {
			__antithesis_instrumentation__.Notify(54722)
			if args.Rand.Float64() < 0.5 {
				__antithesis_instrumentation__.Notify(54723)
				bools[i] = true
			} else {
				__antithesis_instrumentation__.Notify(54724)
				bools[i] = false
			}
		}
	case types.BytesFamily:
		__antithesis_instrumentation__.Notify(54714)
		bytes := args.Vec.Bytes()
		isUUID := args.Vec.Type().Family() == types.UuidFamily
		for i := 0; i < args.N; i++ {
			__antithesis_instrumentation__.Notify(54725)
			bytesLen := args.BytesFixedLength
			if bytesLen <= 0 {
				__antithesis_instrumentation__.Notify(54728)
				bytesLen = args.Rand.Intn(maxVarLen)
			} else {
				__antithesis_instrumentation__.Notify(54729)
			}
			__antithesis_instrumentation__.Notify(54726)
			if isUUID {
				__antithesis_instrumentation__.Notify(54730)
				bytesLen = uuid.Size
			} else {
				__antithesis_instrumentation__.Notify(54731)
			}
			__antithesis_instrumentation__.Notify(54727)
			randBytes := make([]byte, bytesLen)

			_, _ = rand.Read(randBytes)
			bytes.Set(i, randBytes)
		}
	case types.DecimalFamily:
		__antithesis_instrumentation__.Notify(54715)
		decs := args.Vec.Decimal()
		for i := 0; i < args.N; i++ {
			__antithesis_instrumentation__.Notify(54732)

			decs[i].SetFinite(int64(args.Rand.Uint64()), int32(args.Rand.Intn(40)-20))
			if args.ZeroProhibited {
				__antithesis_instrumentation__.Notify(54733)
				if decs[i].IsZero() {
					__antithesis_instrumentation__.Notify(54734)
					i--
				} else {
					__antithesis_instrumentation__.Notify(54735)
				}
			} else {
				__antithesis_instrumentation__.Notify(54736)
			}
		}
	case types.IntFamily:
		__antithesis_instrumentation__.Notify(54716)
		switch args.Vec.Type().Width() {
		case 16:
			__antithesis_instrumentation__.Notify(54737)
			ints := args.Vec.Int16()
			for i := 0; i < args.N; i++ {
				__antithesis_instrumentation__.Notify(54741)
				ints[i] = int16(args.Rand.Uint64())
				if args.IntRange != 0 {
					__antithesis_instrumentation__.Notify(54743)
					ints[i] = ints[i] % int16(args.IntRange)
				} else {
					__antithesis_instrumentation__.Notify(54744)
				}
				__antithesis_instrumentation__.Notify(54742)
				if args.ZeroProhibited {
					__antithesis_instrumentation__.Notify(54745)
					if ints[i] == 0 {
						__antithesis_instrumentation__.Notify(54746)
						i--
					} else {
						__antithesis_instrumentation__.Notify(54747)
					}
				} else {
					__antithesis_instrumentation__.Notify(54748)
				}
			}
		case 32:
			__antithesis_instrumentation__.Notify(54738)
			ints := args.Vec.Int32()
			for i := 0; i < args.N; i++ {
				__antithesis_instrumentation__.Notify(54749)
				ints[i] = int32(args.Rand.Uint64())
				if args.IntRange != 0 {
					__antithesis_instrumentation__.Notify(54751)
					ints[i] = ints[i] % int32(args.IntRange)
				} else {
					__antithesis_instrumentation__.Notify(54752)
				}
				__antithesis_instrumentation__.Notify(54750)
				if args.ZeroProhibited {
					__antithesis_instrumentation__.Notify(54753)
					if ints[i] == 0 {
						__antithesis_instrumentation__.Notify(54754)
						i--
					} else {
						__antithesis_instrumentation__.Notify(54755)
					}
				} else {
					__antithesis_instrumentation__.Notify(54756)
				}
			}
		case 0, 64:
			__antithesis_instrumentation__.Notify(54739)
			ints := args.Vec.Int64()
			for i := 0; i < args.N; i++ {
				__antithesis_instrumentation__.Notify(54757)
				ints[i] = int64(args.Rand.Uint64())
				if args.IntRange != 0 {
					__antithesis_instrumentation__.Notify(54759)
					ints[i] = ints[i] % int64(args.IntRange)
				} else {
					__antithesis_instrumentation__.Notify(54760)
				}
				__antithesis_instrumentation__.Notify(54758)
				if args.ZeroProhibited {
					__antithesis_instrumentation__.Notify(54761)
					if ints[i] == 0 {
						__antithesis_instrumentation__.Notify(54762)
						i--
					} else {
						__antithesis_instrumentation__.Notify(54763)
					}
				} else {
					__antithesis_instrumentation__.Notify(54764)
				}
			}
		default:
			__antithesis_instrumentation__.Notify(54740)
		}
	case types.FloatFamily:
		__antithesis_instrumentation__.Notify(54717)
		floats := args.Vec.Float64()
		for i := 0; i < args.N; i++ {
			__antithesis_instrumentation__.Notify(54765)
			floats[i] = args.Rand.Float64()
			if args.ZeroProhibited {
				__antithesis_instrumentation__.Notify(54766)
				if floats[i] == 0 {
					__antithesis_instrumentation__.Notify(54767)
					i--
				} else {
					__antithesis_instrumentation__.Notify(54768)
				}
			} else {
				__antithesis_instrumentation__.Notify(54769)
			}
		}
	case types.TimestampTZFamily:
		__antithesis_instrumentation__.Notify(54718)
		timestamps := args.Vec.Timestamp()
		for i := 0; i < args.N; i++ {
			__antithesis_instrumentation__.Notify(54770)
			timestamps[i] = timeutil.Unix(args.Rand.Int63n(1000000), args.Rand.Int63n(1000000))
			loc := locations[args.Rand.Intn(len(locations))]
			timestamps[i] = timestamps[i].In(loc)
		}
	case types.IntervalFamily:
		__antithesis_instrumentation__.Notify(54719)
		intervals := args.Vec.Interval()
		for i := 0; i < args.N; i++ {
			__antithesis_instrumentation__.Notify(54771)
			intervals[i] = duration.FromFloat64(args.Rand.Float64())
		}
	case types.JsonFamily:
		__antithesis_instrumentation__.Notify(54720)
		j := args.Vec.JSON()
		for i := 0; i < args.N; i++ {
			__antithesis_instrumentation__.Notify(54772)
			random, err := json.Random(20, args.Rand)
			if err != nil {
				__antithesis_instrumentation__.Notify(54774)
				panic(err)
			} else {
				__antithesis_instrumentation__.Notify(54775)
			}
			__antithesis_instrumentation__.Notify(54773)
			j.Set(i, random)
		}
	default:
		__antithesis_instrumentation__.Notify(54721)
		datums := args.Vec.Datum()
		for i := 0; i < args.N; i++ {
			__antithesis_instrumentation__.Notify(54776)
			datums.Set(i, randgen.RandDatum(args.Rand, args.Vec.Type(), false))
		}
	}
	__antithesis_instrumentation__.Notify(54711)
	args.Vec.Nulls().UnsetNulls()
	if args.NullProbability == 0 {
		__antithesis_instrumentation__.Notify(54777)
		return
	} else {
		__antithesis_instrumentation__.Notify(54778)
	}
	__antithesis_instrumentation__.Notify(54712)

	for i := 0; i < args.N; i++ {
		__antithesis_instrumentation__.Notify(54779)
		if args.Rand.Float64() < args.NullProbability {
			__antithesis_instrumentation__.Notify(54780)
			setNull(args.Rand, args.Vec, i)
		} else {
			__antithesis_instrumentation__.Notify(54781)
		}
	}
}

func setNull(rng *rand.Rand, vec coldata.Vec, i int) {
	__antithesis_instrumentation__.Notify(54782)
	vec.Nulls().SetNull(i)
	switch vec.CanonicalTypeFamily() {
	case types.DecimalFamily:
		__antithesis_instrumentation__.Notify(54783)
		_, err := vec.Decimal()[i].SetFloat64(rng.Float64())
		if err != nil {
			__antithesis_instrumentation__.Notify(54786)
			colexecerror.InternalError(errors.NewAssertionErrorWithWrappedErrf(err, "could not set decimal"))
		} else {
			__antithesis_instrumentation__.Notify(54787)
		}
	case types.IntervalFamily:
		__antithesis_instrumentation__.Notify(54784)
		vec.Interval()[i] = duration.MakeDuration(rng.Int63(), rng.Int63(), rng.Int63())
	default:
		__antithesis_instrumentation__.Notify(54785)
	}
}

func RandomBatch(
	allocator *colmem.Allocator,
	rng *rand.Rand,
	typs []*types.T,
	capacity int,
	length int,
	nullProbability float64,
) coldata.Batch {
	__antithesis_instrumentation__.Notify(54788)
	batch := allocator.NewMemBatchWithFixedCapacity(typs, capacity)
	if length == 0 {
		__antithesis_instrumentation__.Notify(54791)
		length = capacity
	} else {
		__antithesis_instrumentation__.Notify(54792)
	}
	__antithesis_instrumentation__.Notify(54789)
	for _, colVec := range batch.ColVecs() {
		__antithesis_instrumentation__.Notify(54793)
		RandomVec(RandomVecArgs{
			Rand:            rng,
			Vec:             colVec,
			N:               length,
			NullProbability: nullProbability,
		})
	}
	__antithesis_instrumentation__.Notify(54790)
	batch.SetLength(length)
	return batch
}

func RandomSel(rng *rand.Rand, batchSize int, probOfOmitting float64) []int {
	__antithesis_instrumentation__.Notify(54794)
	if probOfOmitting < 0 || func() bool {
		__antithesis_instrumentation__.Notify(54797)
		return probOfOmitting > 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(54798)
		colexecerror.InternalError(errors.AssertionFailedf("probability of omitting a row is %f - outside of [0, 1] range", probOfOmitting))
	} else {
		__antithesis_instrumentation__.Notify(54799)
	}
	__antithesis_instrumentation__.Notify(54795)
	sel := make([]int, 0, batchSize)
	for i := 0; i < batchSize; i++ {
		__antithesis_instrumentation__.Notify(54800)
		if rng.Float64() < probOfOmitting {
			__antithesis_instrumentation__.Notify(54802)
			continue
		} else {
			__antithesis_instrumentation__.Notify(54803)
		}
		__antithesis_instrumentation__.Notify(54801)
		sel = append(sel, i)
	}
	__antithesis_instrumentation__.Notify(54796)
	return sel
}

func RandomBatchWithSel(
	allocator *colmem.Allocator,
	rng *rand.Rand,
	typs []*types.T,
	n int,
	nullProbability float64,
	selProbability float64,
) coldata.Batch {
	__antithesis_instrumentation__.Notify(54804)
	batch := RandomBatch(allocator, rng, typs, n, 0, nullProbability)
	if selProbability != 0 {
		__antithesis_instrumentation__.Notify(54806)
		sel := RandomSel(rng, n, 1-selProbability)
		batch.SetSelection(true)
		copy(batch.Selection(), sel)
		batch.SetLength(len(sel))
	} else {
		__antithesis_instrumentation__.Notify(54807)
	}
	__antithesis_instrumentation__.Notify(54805)
	return batch
}

const (
	defaultMaxSchemaLength = 8
	defaultNumBatches      = 4
)

type RandomDataOpArgs struct {
	DeterministicTyps []*types.T

	MaxSchemaLength int

	BatchSize int

	NumBatches int

	Selection bool

	Nulls bool

	BatchAccumulator func(ctx context.Context, b coldata.Batch, typs []*types.T)
}

type RandomDataOp struct {
	ctx              context.Context
	allocator        *colmem.Allocator
	batchAccumulator func(ctx context.Context, b coldata.Batch, typs []*types.T)
	typs             []*types.T
	rng              *rand.Rand
	batchSize        int
	numBatches       int
	numReturned      int
	selection        bool
	nulls            bool
}

var _ colexecop.Operator = &RandomDataOp{}

func NewRandomDataOp(
	allocator *colmem.Allocator, rng *rand.Rand, args RandomDataOpArgs,
) *RandomDataOp {
	__antithesis_instrumentation__.Notify(54808)
	var (
		maxSchemaLength = defaultMaxSchemaLength
		batchSize       = coldata.BatchSize()
		numBatches      = defaultNumBatches
	)
	if args.MaxSchemaLength > 0 {
		__antithesis_instrumentation__.Notify(54813)
		maxSchemaLength = args.MaxSchemaLength
	} else {
		__antithesis_instrumentation__.Notify(54814)
	}
	__antithesis_instrumentation__.Notify(54809)
	if args.BatchSize > 0 {
		__antithesis_instrumentation__.Notify(54815)
		batchSize = args.BatchSize
	} else {
		__antithesis_instrumentation__.Notify(54816)
	}
	__antithesis_instrumentation__.Notify(54810)
	if args.NumBatches > 0 {
		__antithesis_instrumentation__.Notify(54817)
		numBatches = args.NumBatches
	} else {
		__antithesis_instrumentation__.Notify(54818)
	}
	__antithesis_instrumentation__.Notify(54811)

	typs := args.DeterministicTyps
	if typs == nil {
		__antithesis_instrumentation__.Notify(54819)

		typs = make([]*types.T, 1+rng.Intn(maxSchemaLength))
		for i := range typs {
			__antithesis_instrumentation__.Notify(54820)
			typs[i] = randgen.RandType(rng)
		}
	} else {
		__antithesis_instrumentation__.Notify(54821)
	}
	__antithesis_instrumentation__.Notify(54812)
	return &RandomDataOp{
		allocator:        allocator,
		batchAccumulator: args.BatchAccumulator,
		typs:             typs,
		rng:              rng,
		batchSize:        batchSize,
		numBatches:       numBatches,
		selection:        args.Selection,
		nulls:            args.Nulls,
	}
}

func (o *RandomDataOp) Init(ctx context.Context) {
	__antithesis_instrumentation__.Notify(54822)
	o.ctx = ctx
}

func (o *RandomDataOp) Next() coldata.Batch {
	__antithesis_instrumentation__.Notify(54823)
	if o.numReturned == o.numBatches {
		__antithesis_instrumentation__.Notify(54827)

		b := coldata.ZeroBatch
		if o.batchAccumulator != nil {
			__antithesis_instrumentation__.Notify(54829)
			o.batchAccumulator(o.ctx, b, o.typs)
		} else {
			__antithesis_instrumentation__.Notify(54830)
		}
		__antithesis_instrumentation__.Notify(54828)
		return b
	} else {
		__antithesis_instrumentation__.Notify(54831)
	}
	__antithesis_instrumentation__.Notify(54824)

	var (
		selProbability  float64
		nullProbability float64
	)
	if o.selection {
		__antithesis_instrumentation__.Notify(54832)
		selProbability = o.rng.Float64()
	} else {
		__antithesis_instrumentation__.Notify(54833)
	}
	__antithesis_instrumentation__.Notify(54825)
	if o.nulls && func() bool {
		__antithesis_instrumentation__.Notify(54834)
		return o.rng.Float64() > 0.1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(54835)

		nullProbability = o.rng.Float64()
	} else {
		__antithesis_instrumentation__.Notify(54836)
	}
	__antithesis_instrumentation__.Notify(54826)
	for {
		__antithesis_instrumentation__.Notify(54837)
		b := RandomBatchWithSel(o.allocator, o.rng, o.typs, o.batchSize, nullProbability, selProbability)
		if b.Length() == 0 {
			__antithesis_instrumentation__.Notify(54840)

			continue
		} else {
			__antithesis_instrumentation__.Notify(54841)
		}
		__antithesis_instrumentation__.Notify(54838)
		o.numReturned++
		if o.batchAccumulator != nil {
			__antithesis_instrumentation__.Notify(54842)
			o.batchAccumulator(o.ctx, b, o.typs)
		} else {
			__antithesis_instrumentation__.Notify(54843)
		}
		__antithesis_instrumentation__.Notify(54839)
		return b
	}
}

func (o *RandomDataOp) ChildCount(verbose bool) int {
	__antithesis_instrumentation__.Notify(54844)
	return 0
}

func (o *RandomDataOp) Child(nth int, verbose bool) execinfra.OpNode {
	__antithesis_instrumentation__.Notify(54845)
	colexecerror.InternalError(errors.AssertionFailedf("invalid index %d", nth))

	return nil
}

func (o *RandomDataOp) Typs() []*types.T {
	__antithesis_instrumentation__.Notify(54846)
	return o.typs
}
