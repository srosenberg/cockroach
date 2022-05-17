package kvnemesis

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math/rand"
	"strconv"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/bootstrap"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

type GeneratorConfig struct {
	Ops                   OperationConfig
	NumNodes, NumReplicas int
}

type OperationConfig struct {
	DB             ClientOperationConfig
	Batch          BatchOperationConfig
	ClosureTxn     ClosureTxnConfig
	Split          SplitConfig
	Merge          MergeConfig
	ChangeReplicas ChangeReplicasConfig
	ChangeLease    ChangeLeaseConfig
	ChangeZone     ChangeZoneConfig
}

type ClosureTxnConfig struct {
	TxnClientOps ClientOperationConfig
	TxnBatchOps  BatchOperationConfig

	Commit int

	Rollback int

	CommitInBatch int

	CommitBatchOps ClientOperationConfig
}

type ClientOperationConfig struct {
	GetMissing int

	GetMissingForUpdate int

	GetExisting int

	GetExistingForUpdate int

	PutMissing int

	PutExisting int

	Scan int

	ScanForUpdate int

	ReverseScan int

	ReverseScanForUpdate int

	DeleteMissing int

	DeleteExisting int

	DeleteRange int
}

type BatchOperationConfig struct {
	Batch int
	Ops   ClientOperationConfig
}

type SplitConfig struct {
	SplitNew int

	SplitAgain int
}

type MergeConfig struct {
	MergeNotSplit int

	MergeIsSplit int
}

type ChangeReplicasConfig struct {
	AddReplica int

	RemoveReplica int

	AtomicSwapReplica int
}

type ChangeLeaseConfig struct {
	TransferLease int
}

type ChangeZoneConfig struct {
	ToggleGlobalReads int
}

func newAllOperationsConfig() GeneratorConfig {
	__antithesis_instrumentation__.Notify(90262)
	clientOpConfig := ClientOperationConfig{
		GetMissing:           1,
		GetMissingForUpdate:  1,
		GetExisting:          1,
		GetExistingForUpdate: 1,
		PutMissing:           1,
		PutExisting:          1,
		Scan:                 1,
		ScanForUpdate:        1,
		ReverseScan:          1,
		ReverseScanForUpdate: 1,
		DeleteMissing:        1,
		DeleteExisting:       1,
		DeleteRange:          1,
	}
	batchOpConfig := BatchOperationConfig{
		Batch: 4,
		Ops:   clientOpConfig,
	}
	return GeneratorConfig{Ops: OperationConfig{
		DB:    clientOpConfig,
		Batch: batchOpConfig,
		ClosureTxn: ClosureTxnConfig{
			Commit:         5,
			Rollback:       5,
			CommitInBatch:  5,
			TxnClientOps:   clientOpConfig,
			TxnBatchOps:    batchOpConfig,
			CommitBatchOps: clientOpConfig,
		},
		Split: SplitConfig{
			SplitNew:   1,
			SplitAgain: 1,
		},
		Merge: MergeConfig{
			MergeNotSplit: 1,
			MergeIsSplit:  1,
		},
		ChangeReplicas: ChangeReplicasConfig{
			AddReplica:        1,
			RemoveReplica:     1,
			AtomicSwapReplica: 1,
		},
		ChangeLease: ChangeLeaseConfig{
			TransferLease: 1,
		},
		ChangeZone: ChangeZoneConfig{
			ToggleGlobalReads: 1,
		},
	}}
}

func NewDefaultConfig() GeneratorConfig {
	__antithesis_instrumentation__.Notify(90263)
	config := newAllOperationsConfig()

	config.Ops.DB.DeleteRange = 0
	config.Ops.Batch.Ops.DeleteRange = 0

	config.Ops.ClosureTxn.CommitBatchOps.DeleteRange = 0
	config.Ops.ClosureTxn.TxnBatchOps.Ops.DeleteRange = 0

	config.Ops.Batch = BatchOperationConfig{}

	config.Ops.ClosureTxn.CommitBatchOps.GetExisting = 0
	config.Ops.ClosureTxn.CommitBatchOps.GetMissing = 0
	return config
}

var GeneratorDataTableID = bootstrap.TestingMinUserDescID()

func GeneratorDataSpan() roachpb.Span {
	__antithesis_instrumentation__.Notify(90264)
	return roachpb.Span{
		Key:    keys.SystemSQLCodec.TablePrefix(GeneratorDataTableID),
		EndKey: keys.SystemSQLCodec.TablePrefix(GeneratorDataTableID + 1),
	}
}

type GetReplicasFn func(roachpb.Key) []roachpb.ReplicationTarget

type Generator struct {
	mu struct {
		syncutil.Mutex
		generator
	}
}

func MakeGenerator(config GeneratorConfig, replicasFn GetReplicasFn) (*Generator, error) {
	__antithesis_instrumentation__.Notify(90265)
	if config.NumNodes <= 0 {
		__antithesis_instrumentation__.Notify(90269)
		return nil, errors.Errorf(`NumNodes must be positive got: %d`, config.NumNodes)
	} else {
		__antithesis_instrumentation__.Notify(90270)
	}
	__antithesis_instrumentation__.Notify(90266)
	if config.NumReplicas <= 0 {
		__antithesis_instrumentation__.Notify(90271)
		return nil, errors.Errorf(`NumReplicas must be positive got: %d`, config.NumReplicas)
	} else {
		__antithesis_instrumentation__.Notify(90272)
	}
	__antithesis_instrumentation__.Notify(90267)
	if config.NumReplicas > config.NumNodes {
		__antithesis_instrumentation__.Notify(90273)
		return nil, errors.Errorf(`NumReplicas (%d) must <= NumNodes (%d)`,
			config.NumReplicas, config.NumNodes)
	} else {
		__antithesis_instrumentation__.Notify(90274)
	}
	__antithesis_instrumentation__.Notify(90268)
	g := &Generator{}
	g.mu.generator = generator{
		Config:           config,
		replicasFn:       replicasFn,
		keys:             make(map[string]struct{}),
		currentSplits:    make(map[string]struct{}),
		historicalSplits: make(map[string]struct{}),
	}
	return g, nil
}

func (g *Generator) RandStep(rng *rand.Rand) Step {
	__antithesis_instrumentation__.Notify(90275)
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.mu.RandStep(rng)
}

type generator struct {
	Config     GeneratorConfig
	replicasFn GetReplicasFn

	nextValue int

	keys map[string]struct{}

	currentSplits map[string]struct{}

	historicalSplits map[string]struct{}
}

func (g *generator) RandStep(rng *rand.Rand) Step {
	__antithesis_instrumentation__.Notify(90276)
	var allowed []opGen
	g.registerClientOps(&allowed, &g.Config.Ops.DB)
	g.registerBatchOps(&allowed, &g.Config.Ops.Batch)
	g.registerClosureTxnOps(&allowed, &g.Config.Ops.ClosureTxn)

	addOpGen(&allowed, randSplitNew, g.Config.Ops.Split.SplitNew)
	if len(g.historicalSplits) > 0 {
		__antithesis_instrumentation__.Notify(90282)
		addOpGen(&allowed, randSplitAgain, g.Config.Ops.Split.SplitAgain)
	} else {
		__antithesis_instrumentation__.Notify(90283)
	}
	__antithesis_instrumentation__.Notify(90277)

	addOpGen(&allowed, randMergeNotSplit, g.Config.Ops.Merge.MergeNotSplit)
	if len(g.currentSplits) > 0 {
		__antithesis_instrumentation__.Notify(90284)
		addOpGen(&allowed, randMergeIsSplit, g.Config.Ops.Merge.MergeIsSplit)
	} else {
		__antithesis_instrumentation__.Notify(90285)
	}
	__antithesis_instrumentation__.Notify(90278)

	key := randKey(rng)
	current := g.replicasFn(roachpb.Key(key))
	if len(current) < g.Config.NumNodes {
		__antithesis_instrumentation__.Notify(90286)
		addReplicaFn := makeAddReplicaFn(key, current, false)
		addOpGen(&allowed, addReplicaFn, g.Config.Ops.ChangeReplicas.AddReplica)
	} else {
		__antithesis_instrumentation__.Notify(90287)
	}
	__antithesis_instrumentation__.Notify(90279)
	if len(current) == g.Config.NumReplicas && func() bool {
		__antithesis_instrumentation__.Notify(90288)
		return len(current) < g.Config.NumNodes == true
	}() == true {
		__antithesis_instrumentation__.Notify(90289)
		atomicSwapReplicaFn := makeAddReplicaFn(key, current, true)
		addOpGen(&allowed, atomicSwapReplicaFn, g.Config.Ops.ChangeReplicas.AtomicSwapReplica)
	} else {
		__antithesis_instrumentation__.Notify(90290)
	}
	__antithesis_instrumentation__.Notify(90280)
	if len(current) > g.Config.NumReplicas {
		__antithesis_instrumentation__.Notify(90291)
		removeReplicaFn := makeRemoveReplicaFn(key, current)
		addOpGen(&allowed, removeReplicaFn, g.Config.Ops.ChangeReplicas.RemoveReplica)
	} else {
		__antithesis_instrumentation__.Notify(90292)
	}
	__antithesis_instrumentation__.Notify(90281)
	transferLeaseFn := makeTransferLeaseFn(key, current)
	addOpGen(&allowed, transferLeaseFn, g.Config.Ops.ChangeLease.TransferLease)

	addOpGen(&allowed, toggleGlobalReads, g.Config.Ops.ChangeZone.ToggleGlobalReads)

	return step(g.selectOp(rng, allowed))
}

type opGenFunc func(*generator, *rand.Rand) Operation

type opGen struct {
	fn     opGenFunc
	weight int
}

func addOpGen(valid *[]opGen, fn opGenFunc, weight int) {
	__antithesis_instrumentation__.Notify(90293)
	*valid = append(*valid, opGen{fn: fn, weight: weight})
}

func (g *generator) selectOp(rng *rand.Rand, contextuallyValid []opGen) Operation {
	__antithesis_instrumentation__.Notify(90294)
	var total int
	for _, x := range contextuallyValid {
		__antithesis_instrumentation__.Notify(90297)
		total += x.weight
	}
	__antithesis_instrumentation__.Notify(90295)
	target := rng.Intn(total)
	var sum int
	for _, x := range contextuallyValid {
		__antithesis_instrumentation__.Notify(90298)
		sum += x.weight
		if sum > target {
			__antithesis_instrumentation__.Notify(90299)
			return x.fn(g, rng)
		} else {
			__antithesis_instrumentation__.Notify(90300)
		}
	}
	__antithesis_instrumentation__.Notify(90296)
	panic(`unreachable`)
}

func (g *generator) registerClientOps(allowed *[]opGen, c *ClientOperationConfig) {
	__antithesis_instrumentation__.Notify(90301)
	addOpGen(allowed, randGetMissing, c.GetMissing)
	addOpGen(allowed, randGetMissingForUpdate, c.GetMissingForUpdate)
	addOpGen(allowed, randPutMissing, c.PutMissing)
	addOpGen(allowed, randDelMissing, c.DeleteMissing)
	if len(g.keys) > 0 {
		__antithesis_instrumentation__.Notify(90303)
		addOpGen(allowed, randGetExisting, c.GetExisting)
		addOpGen(allowed, randGetExistingForUpdate, c.GetExistingForUpdate)
		addOpGen(allowed, randPutExisting, c.PutExisting)
		addOpGen(allowed, randDelExisting, c.DeleteExisting)
	} else {
		__antithesis_instrumentation__.Notify(90304)
	}
	__antithesis_instrumentation__.Notify(90302)
	addOpGen(allowed, randScan, c.Scan)
	addOpGen(allowed, randScanForUpdate, c.ScanForUpdate)
	addOpGen(allowed, randReverseScan, c.ReverseScan)
	addOpGen(allowed, randReverseScanForUpdate, c.ReverseScanForUpdate)
	addOpGen(allowed, randDelRange, c.DeleteRange)
}

func (g *generator) registerBatchOps(allowed *[]opGen, c *BatchOperationConfig) {
	__antithesis_instrumentation__.Notify(90305)
	addOpGen(allowed, makeRandBatch(&c.Ops), c.Batch)
}

func randGetMissing(_ *generator, rng *rand.Rand) Operation {
	__antithesis_instrumentation__.Notify(90306)
	return get(randKey(rng))
}

func randGetMissingForUpdate(_ *generator, rng *rand.Rand) Operation {
	__antithesis_instrumentation__.Notify(90307)
	op := get(randKey(rng))
	op.Get.ForUpdate = true
	return op
}

func randGetExisting(g *generator, rng *rand.Rand) Operation {
	__antithesis_instrumentation__.Notify(90308)
	key := randMapKey(rng, g.keys)
	return get(key)
}

func randGetExistingForUpdate(g *generator, rng *rand.Rand) Operation {
	__antithesis_instrumentation__.Notify(90309)
	key := randMapKey(rng, g.keys)
	op := get(key)
	op.Get.ForUpdate = true
	return op
}

func randPutMissing(g *generator, rng *rand.Rand) Operation {
	__antithesis_instrumentation__.Notify(90310)
	value := g.getNextValue()
	key := randKey(rng)
	g.keys[key] = struct{}{}
	return put(key, value)
}

func randPutExisting(g *generator, rng *rand.Rand) Operation {
	__antithesis_instrumentation__.Notify(90311)
	value := g.getNextValue()
	key := randMapKey(rng, g.keys)
	return put(key, value)
}

func randScan(g *generator, rng *rand.Rand) Operation {
	__antithesis_instrumentation__.Notify(90312)
	key, endKey := randSpan(rng)
	return scan(key, endKey)
}

func randScanForUpdate(g *generator, rng *rand.Rand) Operation {
	__antithesis_instrumentation__.Notify(90313)
	op := randScan(g, rng)
	op.Scan.ForUpdate = true
	return op
}

func randReverseScan(g *generator, rng *rand.Rand) Operation {
	__antithesis_instrumentation__.Notify(90314)
	op := randScan(g, rng)
	op.Scan.Reverse = true
	return op
}

func randReverseScanForUpdate(g *generator, rng *rand.Rand) Operation {
	__antithesis_instrumentation__.Notify(90315)
	op := randReverseScan(g, rng)
	op.Scan.ForUpdate = true
	return op
}

func randDelMissing(g *generator, rng *rand.Rand) Operation {
	__antithesis_instrumentation__.Notify(90316)
	key := randKey(rng)
	g.keys[key] = struct{}{}
	return del(key)
}

func randDelExisting(g *generator, rng *rand.Rand) Operation {
	__antithesis_instrumentation__.Notify(90317)
	key := randMapKey(rng, g.keys)
	return del(key)
}

func randDelRange(g *generator, rng *rand.Rand) Operation {
	__antithesis_instrumentation__.Notify(90318)

	key, endKey := randSpan(rng)
	return delRange(key, endKey)
}

func randSplitNew(g *generator, rng *rand.Rand) Operation {
	__antithesis_instrumentation__.Notify(90319)
	key := randKey(rng)
	g.currentSplits[key] = struct{}{}
	g.historicalSplits[key] = struct{}{}
	return split(key)
}

func randSplitAgain(g *generator, rng *rand.Rand) Operation {
	__antithesis_instrumentation__.Notify(90320)
	key := randMapKey(rng, g.historicalSplits)
	g.currentSplits[key] = struct{}{}
	return split(key)
}

func randMergeNotSplit(g *generator, rng *rand.Rand) Operation {
	__antithesis_instrumentation__.Notify(90321)
	key := randKey(rng)
	return merge(key)
}

func randMergeIsSplit(g *generator, rng *rand.Rand) Operation {
	__antithesis_instrumentation__.Notify(90322)
	key := randMapKey(rng, g.currentSplits)

	delete(g.currentSplits, key)
	return merge(key)
}

func makeRemoveReplicaFn(key string, current []roachpb.ReplicationTarget) opGenFunc {
	__antithesis_instrumentation__.Notify(90323)
	return func(g *generator, rng *rand.Rand) Operation {
		__antithesis_instrumentation__.Notify(90324)
		change := roachpb.ReplicationChange{
			ChangeType: roachpb.REMOVE_VOTER,
			Target:     current[rng.Intn(len(current))],
		}
		return changeReplicas(key, change)
	}
}

func makeAddReplicaFn(key string, current []roachpb.ReplicationTarget, atomicSwap bool) opGenFunc {
	__antithesis_instrumentation__.Notify(90325)
	return func(g *generator, rng *rand.Rand) Operation {
		__antithesis_instrumentation__.Notify(90326)
		candidatesMap := make(map[roachpb.ReplicationTarget]struct{})
		for i := 0; i < g.Config.NumNodes; i++ {
			__antithesis_instrumentation__.Notify(90331)
			t := roachpb.ReplicationTarget{NodeID: roachpb.NodeID(i + 1), StoreID: roachpb.StoreID(i + 1)}
			candidatesMap[t] = struct{}{}
		}
		__antithesis_instrumentation__.Notify(90327)
		for _, replica := range current {
			__antithesis_instrumentation__.Notify(90332)
			delete(candidatesMap, replica)
		}
		__antithesis_instrumentation__.Notify(90328)
		var candidates []roachpb.ReplicationTarget
		for candidate := range candidatesMap {
			__antithesis_instrumentation__.Notify(90333)
			candidates = append(candidates, candidate)
		}
		__antithesis_instrumentation__.Notify(90329)
		candidate := candidates[rng.Intn(len(candidates))]
		changes := []roachpb.ReplicationChange{{
			ChangeType: roachpb.ADD_VOTER,
			Target:     candidate,
		}}
		if atomicSwap {
			__antithesis_instrumentation__.Notify(90334)
			changes = append(changes, roachpb.ReplicationChange{
				ChangeType: roachpb.REMOVE_VOTER,
				Target:     current[rng.Intn(len(current))],
			})
		} else {
			__antithesis_instrumentation__.Notify(90335)
		}
		__antithesis_instrumentation__.Notify(90330)
		return changeReplicas(key, changes...)
	}
}

func makeTransferLeaseFn(key string, current []roachpb.ReplicationTarget) opGenFunc {
	__antithesis_instrumentation__.Notify(90336)
	return func(g *generator, rng *rand.Rand) Operation {
		__antithesis_instrumentation__.Notify(90337)
		target := current[rng.Intn(len(current))]
		return transferLease(key, target.StoreID)
	}
}

func toggleGlobalReads(_ *generator, _ *rand.Rand) Operation {
	__antithesis_instrumentation__.Notify(90338)
	return changeZone(ChangeZoneType_ToggleGlobalReads)
}

func makeRandBatch(c *ClientOperationConfig) opGenFunc {
	__antithesis_instrumentation__.Notify(90339)
	return func(g *generator, rng *rand.Rand) Operation {
		__antithesis_instrumentation__.Notify(90340)
		var allowed []opGen
		g.registerClientOps(&allowed, c)

		numOps := rng.Intn(4)
		ops := make([]Operation, numOps)
		for i := range ops {
			__antithesis_instrumentation__.Notify(90342)
			ops[i] = g.selectOp(rng, allowed)
		}
		__antithesis_instrumentation__.Notify(90341)
		return batch(ops...)
	}
}

func (g *generator) registerClosureTxnOps(allowed *[]opGen, c *ClosureTxnConfig) {
	__antithesis_instrumentation__.Notify(90343)
	addOpGen(allowed,
		makeClosureTxn(ClosureTxnType_Commit, &c.TxnClientOps, &c.TxnBatchOps, nil), c.Commit)
	addOpGen(allowed,
		makeClosureTxn(ClosureTxnType_Rollback, &c.TxnClientOps, &c.TxnBatchOps, nil), c.Rollback)
	addOpGen(allowed,
		makeClosureTxn(ClosureTxnType_Commit, &c.TxnClientOps, &c.TxnBatchOps, &c.CommitBatchOps), c.CommitInBatch)
}

func makeClosureTxn(
	txnType ClosureTxnType,
	txnClientOps *ClientOperationConfig,
	txnBatchOps *BatchOperationConfig,
	commitInBatch *ClientOperationConfig,
) opGenFunc {
	__antithesis_instrumentation__.Notify(90344)
	return func(g *generator, rng *rand.Rand) Operation {
		__antithesis_instrumentation__.Notify(90345)
		var allowed []opGen
		g.registerClientOps(&allowed, txnClientOps)
		g.registerBatchOps(&allowed, txnBatchOps)
		numOps := rng.Intn(4)
		ops := make([]Operation, numOps)
		for i := range ops {
			__antithesis_instrumentation__.Notify(90348)
			ops[i] = g.selectOp(rng, allowed)
		}
		__antithesis_instrumentation__.Notify(90346)
		op := closureTxn(txnType, ops...)
		if commitInBatch != nil {
			__antithesis_instrumentation__.Notify(90349)
			if txnType != ClosureTxnType_Commit {
				__antithesis_instrumentation__.Notify(90351)
				panic(errors.AssertionFailedf(`CommitInBatch must commit got: %s`, txnType))
			} else {
				__antithesis_instrumentation__.Notify(90352)
			}
			__antithesis_instrumentation__.Notify(90350)
			op.ClosureTxn.CommitInBatch = makeRandBatch(commitInBatch)(g, rng).Batch
		} else {
			__antithesis_instrumentation__.Notify(90353)
		}
		__antithesis_instrumentation__.Notify(90347)
		return op
	}
}

func (g *generator) getNextValue() string {
	__antithesis_instrumentation__.Notify(90354)
	value := `v-` + strconv.Itoa(g.nextValue)
	g.nextValue++
	return value
}

func randKey(rng *rand.Rand) string {
	__antithesis_instrumentation__.Notify(90355)
	u, err := uuid.NewGenWithReader(rng).NewV4()
	if err != nil {
		__antithesis_instrumentation__.Notify(90357)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(90358)
	}
	__antithesis_instrumentation__.Notify(90356)
	key := GeneratorDataSpan().Key
	key = encoding.EncodeStringAscending(key, u.Short())
	return string(key)
}

func randMapKey(rng *rand.Rand, m map[string]struct{}) string {
	__antithesis_instrumentation__.Notify(90359)
	keys := make([]string, 0, len(m))
	for key := range m {
		__antithesis_instrumentation__.Notify(90362)
		keys = append(keys, key)
	}
	__antithesis_instrumentation__.Notify(90360)
	if len(keys) == 0 {
		__antithesis_instrumentation__.Notify(90363)
		return randKey(rng)
	} else {
		__antithesis_instrumentation__.Notify(90364)
	}
	__antithesis_instrumentation__.Notify(90361)
	return keys[rng.Intn(len(keys))]
}

func randSpan(rng *rand.Rand) (string, string) {
	__antithesis_instrumentation__.Notify(90365)
	key, endKey := randKey(rng), randKey(rng)
	if endKey < key {
		__antithesis_instrumentation__.Notify(90367)
		key, endKey = endKey, key
	} else {
		__antithesis_instrumentation__.Notify(90368)
		if endKey == key {
			__antithesis_instrumentation__.Notify(90369)
			endKey = string(roachpb.Key(key).Next())
		} else {
			__antithesis_instrumentation__.Notify(90370)
		}
	}
	__antithesis_instrumentation__.Notify(90366)
	return key, endKey
}

func step(op Operation) Step {
	__antithesis_instrumentation__.Notify(90371)
	return Step{Op: op}
}

func batch(ops ...Operation) Operation {
	__antithesis_instrumentation__.Notify(90372)
	return Operation{Batch: &BatchOperation{Ops: ops}}
}

func opSlice(ops ...Operation) []Operation {
	__antithesis_instrumentation__.Notify(90373)
	return ops
}

func closureTxn(typ ClosureTxnType, ops ...Operation) Operation {
	__antithesis_instrumentation__.Notify(90374)
	return Operation{ClosureTxn: &ClosureTxnOperation{Ops: ops, Type: typ}}
}

func closureTxnCommitInBatch(commitInBatch []Operation, ops ...Operation) Operation {
	__antithesis_instrumentation__.Notify(90375)
	o := closureTxn(ClosureTxnType_Commit, ops...)
	if len(commitInBatch) > 0 {
		__antithesis_instrumentation__.Notify(90377)
		o.ClosureTxn.CommitInBatch = &BatchOperation{Ops: commitInBatch}
	} else {
		__antithesis_instrumentation__.Notify(90378)
	}
	__antithesis_instrumentation__.Notify(90376)
	return o
}

func get(key string) Operation {
	__antithesis_instrumentation__.Notify(90379)
	return Operation{Get: &GetOperation{Key: []byte(key)}}
}

func getForUpdate(key string) Operation {
	__antithesis_instrumentation__.Notify(90380)
	return Operation{Get: &GetOperation{Key: []byte(key), ForUpdate: true}}
}

func put(key, value string) Operation {
	__antithesis_instrumentation__.Notify(90381)
	return Operation{Put: &PutOperation{Key: []byte(key), Value: []byte(value)}}
}

func scan(key, endKey string) Operation {
	__antithesis_instrumentation__.Notify(90382)
	return Operation{Scan: &ScanOperation{Key: []byte(key), EndKey: []byte(endKey)}}
}

func scanForUpdate(key, endKey string) Operation {
	__antithesis_instrumentation__.Notify(90383)
	return Operation{Scan: &ScanOperation{Key: []byte(key), EndKey: []byte(endKey), ForUpdate: true}}
}

func reverseScan(key, endKey string) Operation {
	__antithesis_instrumentation__.Notify(90384)
	return Operation{Scan: &ScanOperation{Key: []byte(key), EndKey: []byte(endKey), Reverse: true}}
}

func reverseScanForUpdate(key, endKey string) Operation {
	__antithesis_instrumentation__.Notify(90385)
	return Operation{Scan: &ScanOperation{Key: []byte(key), EndKey: []byte(endKey), Reverse: true, ForUpdate: true}}
}

func del(key string) Operation {
	__antithesis_instrumentation__.Notify(90386)
	return Operation{Delete: &DeleteOperation{Key: []byte(key)}}
}

func delRange(key, endKey string) Operation {
	__antithesis_instrumentation__.Notify(90387)
	return Operation{DeleteRange: &DeleteRangeOperation{Key: []byte(key), EndKey: []byte(endKey)}}
}

func split(key string) Operation {
	__antithesis_instrumentation__.Notify(90388)
	return Operation{Split: &SplitOperation{Key: []byte(key)}}
}

func merge(key string) Operation {
	__antithesis_instrumentation__.Notify(90389)
	return Operation{Merge: &MergeOperation{Key: []byte(key)}}
}

func changeReplicas(key string, changes ...roachpb.ReplicationChange) Operation {
	__antithesis_instrumentation__.Notify(90390)
	return Operation{ChangeReplicas: &ChangeReplicasOperation{Key: []byte(key), Changes: changes}}
}

func transferLease(key string, target roachpb.StoreID) Operation {
	__antithesis_instrumentation__.Notify(90391)
	return Operation{TransferLease: &TransferLeaseOperation{Key: []byte(key), Target: target}}
}

func changeZone(changeType ChangeZoneType) Operation {
	__antithesis_instrumentation__.Notify(90392)
	return Operation{ChangeZone: &ChangeZoneOperation{Type: changeType}}
}
