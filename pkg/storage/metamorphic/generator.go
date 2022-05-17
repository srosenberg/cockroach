package metamorphic

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/vfs"
)

const zipfMax uint64 = 100000

func makeStorageConfig(path string) base.StorageConfig {
	__antithesis_instrumentation__.Notify(639640)
	return base.StorageConfig{
		Dir:      path,
		Settings: cluster.MakeTestingClusterSettings(),
	}
}

func rngIntRange(rng *rand.Rand, min int64, max int64) int64 {
	__antithesis_instrumentation__.Notify(639641)
	return min + rng.Int63n(max-min)
}

type engineConfig struct {
	name string
	opts *pebble.Options
}

func (e *engineConfig) create(path string, fs vfs.FS) (storage.Engine, error) {
	__antithesis_instrumentation__.Notify(639642)
	pebbleConfig := storage.PebbleConfig{
		StorageConfig: makeStorageConfig(path),
		Opts:          e.opts,
	}
	if pebbleConfig.Opts == nil {
		__antithesis_instrumentation__.Notify(639645)
		pebbleConfig.Opts = storage.DefaultPebbleOptions()
	} else {
		__antithesis_instrumentation__.Notify(639646)
	}
	__antithesis_instrumentation__.Notify(639643)
	if fs != nil {
		__antithesis_instrumentation__.Notify(639647)
		pebbleConfig.Opts.FS = fs
	} else {
		__antithesis_instrumentation__.Notify(639648)
	}
	__antithesis_instrumentation__.Notify(639644)
	pebbleConfig.Opts.Cache = pebble.NewCache(1 << 20)
	defer pebbleConfig.Opts.Cache.Unref()

	return storage.NewPebble(context.Background(), pebbleConfig)
}

var _ fmt.Stringer = &engineConfig{}

func (e *engineConfig) String() string {
	__antithesis_instrumentation__.Notify(639649)
	return e.name
}

var engineConfigStandard = engineConfig{"standard=0", storage.DefaultPebbleOptions()}

type engineSequence struct {
	name    string
	configs []engineConfig
}

type metaTestRunner struct {
	ctx             context.Context
	w               io.Writer
	t               *testing.T
	rng             *rand.Rand
	seed            int64
	path            string
	engineFS        vfs.FS
	engineSeq       engineSequence
	curEngine       int
	restarts        bool
	engine          storage.Engine
	tsGenerator     tsGenerator
	opGenerators    map[operandType]operandGenerator
	txnGenerator    *txnGenerator
	spGenerator     *savepointGenerator
	rwGenerator     *readWriterGenerator
	iterGenerator   *iteratorGenerator
	keyGenerator    *keyGenerator
	txnKeyGenerator *txnKeyGenerator
	valueGenerator  *valueGenerator
	pastTSGenerator *pastTSGenerator
	nextTSGenerator *nextTSGenerator
	floatGenerator  *floatGenerator
	boolGenerator   *boolGenerator
	openIters       map[iteratorID]iteratorInfo
	openBatches     map[readWriterID]storage.ReadWriter
	openTxns        map[txnID]*roachpb.Transaction
	openSavepoints  map[txnID][]enginepb.TxnSeq
	nameToGenerator map[string]*opGenerator
	ops             []opRun
	weights         []int
	st              *cluster.Settings
}

func (m *metaTestRunner) init() {

	m.rng = rand.New(rand.NewSource(m.seed))
	m.tsGenerator.init(m.rng)
	m.curEngine = 0
	m.printComment(fmt.Sprintf("seed: %d", m.seed))

	var err error
	m.engine, err = m.engineSeq.configs[0].create(m.path, m.engineFS)
	m.printComment(fmt.Sprintf("engine options: %s", m.engineSeq.configs[0].opts.String()))
	if err != nil {
		m.engine = nil
		m.t.Fatal(err)
	}

	m.txnGenerator = &txnGenerator{
		rng:            m.rng,
		tsGenerator:    &m.tsGenerator,
		txnIDMap:       make(map[txnID]*roachpb.Transaction),
		openBatches:    make(map[txnID]map[readWriterID]struct{}),
		inUseKeys:      make([]writtenKeySpan, 0),
		openSavepoints: make(map[txnID]int),
		testRunner:     m,
	}
	m.spGenerator = &savepointGenerator{
		rng:          m.rng,
		txnGenerator: m.txnGenerator,
	}
	m.rwGenerator = &readWriterGenerator{
		rng:        m.rng,
		m:          m,
		batchIDMap: make(map[readWriterID]storage.Batch),
	}
	m.iterGenerator = &iteratorGenerator{
		rng:          m.rng,
		readerToIter: make(map[readWriterID][]iteratorID),
		iterInfo:     make(map[iteratorID]iteratorInfo),
	}
	m.keyGenerator = &keyGenerator{
		rng:         m.rng,
		tsGenerator: &m.tsGenerator,
	}
	m.txnKeyGenerator = &txnKeyGenerator{
		txns: m.txnGenerator,
		keys: m.keyGenerator,
	}
	m.valueGenerator = &valueGenerator{m.rng}
	m.pastTSGenerator = &pastTSGenerator{
		rng:         m.rng,
		tsGenerator: &m.tsGenerator,
	}
	m.nextTSGenerator = &nextTSGenerator{
		pastTSGenerator: pastTSGenerator{
			rng:         m.rng,
			tsGenerator: &m.tsGenerator,
		},
	}
	m.floatGenerator = &floatGenerator{rng: m.rng}
	m.boolGenerator = &boolGenerator{rng: m.rng}

	m.opGenerators = map[operandType]operandGenerator{
		operandTransaction:   m.txnGenerator,
		operandReadWriter:    m.rwGenerator,
		operandMVCCKey:       m.keyGenerator,
		operandUnusedMVCCKey: m.txnKeyGenerator,
		operandPastTS:        m.pastTSGenerator,
		operandNextTS:        m.nextTSGenerator,
		operandValue:         m.valueGenerator,
		operandIterator:      m.iterGenerator,
		operandFloat:         m.floatGenerator,
		operandBool:          m.boolGenerator,
		operandSavepoint:     m.spGenerator,
	}

	m.nameToGenerator = make(map[string]*opGenerator)
	m.weights = make([]int, len(opGenerators))
	for i := range opGenerators {
		m.weights[i] = opGenerators[i].weight
		if !m.restarts && opGenerators[i].name == "restart" {

			m.weights[i] = 0
		}
		m.nameToGenerator[opGenerators[i].name] = &opGenerators[i]
	}
	m.ops = nil
	m.openIters = make(map[iteratorID]iteratorInfo)
	m.openBatches = make(map[readWriterID]storage.ReadWriter)
	m.openTxns = make(map[txnID]*roachpb.Transaction)
	m.openSavepoints = make(map[txnID][]enginepb.TxnSeq)
}

func (m *metaTestRunner) closeGenerators() {
	__antithesis_instrumentation__.Notify(639650)
	closingOrder := []operandGenerator{
		m.iterGenerator,
		m.rwGenerator,
		m.txnGenerator,
	}
	for _, generator := range closingOrder {
		__antithesis_instrumentation__.Notify(639651)
		generator.closeAll()
	}
}

func (m *metaTestRunner) closeAll() {
	__antithesis_instrumentation__.Notify(639652)
	if m.engine == nil {
		__antithesis_instrumentation__.Notify(639656)

		return
	} else {
		__antithesis_instrumentation__.Notify(639657)
	}
	__antithesis_instrumentation__.Notify(639653)

	for _, iter := range m.openIters {
		__antithesis_instrumentation__.Notify(639658)
		iter.iter.Close()
	}
	__antithesis_instrumentation__.Notify(639654)
	for _, batch := range m.openBatches {
		__antithesis_instrumentation__.Notify(639659)
		batch.Close()
	}
	__antithesis_instrumentation__.Notify(639655)

	m.openIters = make(map[iteratorID]iteratorInfo)
	m.openBatches = make(map[readWriterID]storage.ReadWriter)
	m.openTxns = make(map[txnID]*roachpb.Transaction)
	m.openSavepoints = make(map[txnID][]enginepb.TxnSeq)
	if m.engine != nil {
		__antithesis_instrumentation__.Notify(639660)
		m.engine.Close()
		m.engine = nil
	} else {
		__antithesis_instrumentation__.Notify(639661)
	}
}

func (m *metaTestRunner) getTxn(id txnID) *roachpb.Transaction {
	__antithesis_instrumentation__.Notify(639662)
	txn, ok := m.openTxns[id]
	if !ok {
		__antithesis_instrumentation__.Notify(639664)
		panic(fmt.Sprintf("txn with id %s not found", string(id)))
	} else {
		__antithesis_instrumentation__.Notify(639665)
	}
	__antithesis_instrumentation__.Notify(639663)
	return txn
}

func (m *metaTestRunner) setTxn(id txnID, txn *roachpb.Transaction) {
	__antithesis_instrumentation__.Notify(639666)
	m.openTxns[id] = txn
}

func (m *metaTestRunner) getReadWriter(id readWriterID) storage.ReadWriter {
	__antithesis_instrumentation__.Notify(639667)
	if id == "engine" {
		__antithesis_instrumentation__.Notify(639670)
		return m.engine
	} else {
		__antithesis_instrumentation__.Notify(639671)
	}
	__antithesis_instrumentation__.Notify(639668)

	batch, ok := m.openBatches[id]
	if !ok {
		__antithesis_instrumentation__.Notify(639672)
		panic(fmt.Sprintf("batch with id %s not found", string(id)))
	} else {
		__antithesis_instrumentation__.Notify(639673)
	}
	__antithesis_instrumentation__.Notify(639669)
	return batch
}

func (m *metaTestRunner) setReadWriter(id readWriterID, rw storage.ReadWriter) {
	__antithesis_instrumentation__.Notify(639674)
	if id == "engine" {
		__antithesis_instrumentation__.Notify(639676)

		return
	} else {
		__antithesis_instrumentation__.Notify(639677)
	}
	__antithesis_instrumentation__.Notify(639675)
	m.openBatches[id] = rw
}

func (m *metaTestRunner) getIterInfo(id iteratorID) iteratorInfo {
	__antithesis_instrumentation__.Notify(639678)
	iter, ok := m.openIters[id]
	if !ok {
		__antithesis_instrumentation__.Notify(639680)
		panic(fmt.Sprintf("iter with id %s not found", string(id)))
	} else {
		__antithesis_instrumentation__.Notify(639681)
	}
	__antithesis_instrumentation__.Notify(639679)
	return iter
}

func (m *metaTestRunner) setIterInfo(id iteratorID, iterInfo iteratorInfo) {
	__antithesis_instrumentation__.Notify(639682)
	m.openIters[id] = iterInfo
}

func (m *metaTestRunner) generateAndRun(n int) {
	__antithesis_instrumentation__.Notify(639683)
	deck := newDeck(m.rng, m.weights...)
	for i := 0; i < n; i++ {
		__antithesis_instrumentation__.Notify(639685)
		op := &opGenerators[deck.Int()]

		m.resolveAndAddOp(op)
	}
	__antithesis_instrumentation__.Notify(639684)

	for i := range m.ops {
		__antithesis_instrumentation__.Notify(639686)
		opRun := &m.ops[i]
		output := opRun.op.run(m.ctx)
		m.printOp(opRun.name, opRun.args, output)
	}
}

func (m *metaTestRunner) restart() (engineConfig, engineConfig) {
	__antithesis_instrumentation__.Notify(639687)
	m.closeAll()
	oldEngine := m.engineSeq.configs[m.curEngine]

	m.curEngine++
	if m.curEngine >= len(m.engineSeq.configs) {
		__antithesis_instrumentation__.Notify(639690)

		m.curEngine = 0
	} else {
		__antithesis_instrumentation__.Notify(639691)
	}
	__antithesis_instrumentation__.Notify(639688)

	var err error
	m.engine, err = m.engineSeq.configs[m.curEngine].create(m.path, m.engineFS)
	if err != nil {
		__antithesis_instrumentation__.Notify(639692)
		m.engine = nil
		m.t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(639693)
	}
	__antithesis_instrumentation__.Notify(639689)
	return oldEngine, m.engineSeq.configs[m.curEngine]
}

func (m *metaTestRunner) parseFileAndRun(f io.Reader) {
	__antithesis_instrumentation__.Notify(639694)
	reader := bufio.NewReader(f)
	lineCount := uint64(0)
	for {
		__antithesis_instrumentation__.Notify(639696)
		var opName, argListString, expectedOutput string
		var firstByte byte
		var err error

		lineCount++

		firstByte, err = reader.ReadByte()
		if err != nil {
			__antithesis_instrumentation__.Notify(639705)
			if err == io.EOF {
				__antithesis_instrumentation__.Notify(639707)
				break
			} else {
				__antithesis_instrumentation__.Notify(639708)
			}
			__antithesis_instrumentation__.Notify(639706)
			m.t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(639709)
		}
		__antithesis_instrumentation__.Notify(639697)
		if firstByte == '#' {
			__antithesis_instrumentation__.Notify(639710)

			if _, err := reader.ReadString('\n'); err != nil {
				__antithesis_instrumentation__.Notify(639712)
				if err == io.EOF {
					__antithesis_instrumentation__.Notify(639714)
					break
				} else {
					__antithesis_instrumentation__.Notify(639715)
				}
				__antithesis_instrumentation__.Notify(639713)
				m.t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(639716)
			}
			__antithesis_instrumentation__.Notify(639711)
			continue
		} else {
			__antithesis_instrumentation__.Notify(639717)
		}
		__antithesis_instrumentation__.Notify(639698)

		if opName, err = reader.ReadString('('); err != nil {
			__antithesis_instrumentation__.Notify(639718)
			if err == io.EOF {
				__antithesis_instrumentation__.Notify(639720)
				break
			} else {
				__antithesis_instrumentation__.Notify(639721)
			}
			__antithesis_instrumentation__.Notify(639719)
			m.t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(639722)
		}
		__antithesis_instrumentation__.Notify(639699)
		opName = string(firstByte) + opName[:len(opName)-1]

		if argListString, err = reader.ReadString(')'); err != nil {
			__antithesis_instrumentation__.Notify(639723)
			m.t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(639724)
		}
		__antithesis_instrumentation__.Notify(639700)

		argStrings := strings.Split(argListString, ", ")

		lastElem := argStrings[len(argStrings)-1]
		if strings.HasSuffix(lastElem, ")") {
			__antithesis_instrumentation__.Notify(639725)
			lastElem = lastElem[:len(lastElem)-1]
			if len(lastElem) > 0 {
				__antithesis_instrumentation__.Notify(639726)
				argStrings[len(argStrings)-1] = lastElem
			} else {
				__antithesis_instrumentation__.Notify(639727)
				argStrings = argStrings[:len(argStrings)-1]
			}
		} else {
			__antithesis_instrumentation__.Notify(639728)
			m.t.Fatalf("while parsing: last element %s did not have ) suffix", lastElem)
		}
		__antithesis_instrumentation__.Notify(639701)

		if _, err = reader.ReadString('>'); err != nil {
			__antithesis_instrumentation__.Notify(639729)
			m.t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(639730)
		}
		__antithesis_instrumentation__.Notify(639702)

		if _, err = reader.Discard(1); err != nil {
			__antithesis_instrumentation__.Notify(639731)
			m.t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(639732)
		}
		__antithesis_instrumentation__.Notify(639703)
		if expectedOutput, err = reader.ReadString('\n'); err != nil {
			__antithesis_instrumentation__.Notify(639733)
			m.t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(639734)
		}
		__antithesis_instrumentation__.Notify(639704)
		opGenerator := m.nameToGenerator[opName]
		m.ops = append(m.ops, opRun{
			name:           opGenerator.name,
			op:             opGenerator.generate(m.ctx, m, argStrings...),
			args:           argStrings,
			lineNum:        lineCount,
			expectedOutput: expectedOutput,
		})
	}
	__antithesis_instrumentation__.Notify(639695)

	for i := range m.ops {
		__antithesis_instrumentation__.Notify(639735)
		op := &m.ops[i]
		actualOutput := op.op.run(m.ctx)
		m.printOp(op.name, op.args, actualOutput)
		if strings.Compare(strings.TrimSpace(op.expectedOutput), strings.TrimSpace(actualOutput)) != 0 {
			__antithesis_instrumentation__.Notify(639736)

			if strings.Contains(op.expectedOutput, "error") && func() bool {
				__antithesis_instrumentation__.Notify(639738)
				return strings.Contains(actualOutput, "error") == true
			}() == true {
				__antithesis_instrumentation__.Notify(639739)
				continue
			} else {
				__antithesis_instrumentation__.Notify(639740)
			}
			__antithesis_instrumentation__.Notify(639737)
			m.t.Fatalf("mismatching output at line %d, operation index %d: expected %s, got %s", op.lineNum, i, op.expectedOutput, actualOutput)
		} else {
			__antithesis_instrumentation__.Notify(639741)
		}
	}
}

func (m *metaTestRunner) generateAndAddOp(run opReference) mvccOp {
	__antithesis_instrumentation__.Notify(639742)
	opGenerator := run.generator

	if opGenerator.dependentOps != nil {
		__antithesis_instrumentation__.Notify(639744)
		for _, opReference := range opGenerator.dependentOps(m, run.args...) {
			__antithesis_instrumentation__.Notify(639745)
			m.generateAndAddOp(opReference)
		}
	} else {
		__antithesis_instrumentation__.Notify(639746)
	}
	__antithesis_instrumentation__.Notify(639743)

	op := opGenerator.generate(m.ctx, m, run.args...)
	m.ops = append(m.ops, opRun{
		name: opGenerator.name,
		op:   op,
		args: run.args,
	})
	return op
}

func (m *metaTestRunner) resolveAndAddOp(op *opGenerator, fixedArgs ...string) {
	__antithesis_instrumentation__.Notify(639747)
	argStrings := make([]string, len(op.operands))
	copy(argStrings, fixedArgs)

	for i, operand := range op.operands {
		__antithesis_instrumentation__.Notify(639749)
		if i < len(fixedArgs) {
			__antithesis_instrumentation__.Notify(639753)
			continue
		} else {
			__antithesis_instrumentation__.Notify(639754)
		}
		__antithesis_instrumentation__.Notify(639750)
		opGenerator := m.opGenerators[operand]

		if i == len(op.operands)-1 && func() bool {
			__antithesis_instrumentation__.Notify(639755)
			return op.isOpener == true
		}() == true {
			__antithesis_instrumentation__.Notify(639756)
			argStrings[i] = opGenerator.getNew(argStrings[:i])
			continue
		} else {
			__antithesis_instrumentation__.Notify(639757)
		}
		__antithesis_instrumentation__.Notify(639751)
		if opGenerator.count(argStrings[:i]) == 0 {
			__antithesis_instrumentation__.Notify(639758)
			var args []string

			switch opGenerator.opener() {
			case "txn_create_savepoint":
				__antithesis_instrumentation__.Notify(639760)
				args = append(args, argStrings[0])
			default:
				__antithesis_instrumentation__.Notify(639761)
			}
			__antithesis_instrumentation__.Notify(639759)

			m.resolveAndAddOp(m.nameToGenerator[opGenerator.opener()], args...)
		} else {
			__antithesis_instrumentation__.Notify(639762)
		}
		__antithesis_instrumentation__.Notify(639752)
		argStrings[i] = opGenerator.get(argStrings[:i])
	}
	__antithesis_instrumentation__.Notify(639748)

	m.generateAndAddOp(opReference{
		generator: op,
		args:      argStrings,
	})
}

func (m *metaTestRunner) printOp(opName string, argStrings []string, output string) {
	__antithesis_instrumentation__.Notify(639763)
	fmt.Fprintf(m.w, "%s(", opName)
	for i, arg := range argStrings {
		__antithesis_instrumentation__.Notify(639765)
		if i > 0 {
			__antithesis_instrumentation__.Notify(639767)
			fmt.Fprintf(m.w, ", ")
		} else {
			__antithesis_instrumentation__.Notify(639768)
		}
		__antithesis_instrumentation__.Notify(639766)
		fmt.Fprintf(m.w, "%s", arg)
	}
	__antithesis_instrumentation__.Notify(639764)
	fmt.Fprintf(m.w, ") -> %s\n", output)
}

func (m *metaTestRunner) printComment(comment string) {
	__antithesis_instrumentation__.Notify(639769)
	comment = strings.ReplaceAll(comment, "\n", "\n# ")
	fmt.Fprintf(m.w, "# %s\n", comment)
}

type tsGenerator struct {
	lastTS hlc.Timestamp
	zipf   *rand.Zipf
}

func (t *tsGenerator) init(rng *rand.Rand) {
	t.zipf = rand.NewZipf(rng, 2, 5, zipfMax)
}

func (t *tsGenerator) generate() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(639770)
	t.lastTS.WallTime++
	return t.lastTS
}

func (t *tsGenerator) randomPastTimestamp(rng *rand.Rand) hlc.Timestamp {
	__antithesis_instrumentation__.Notify(639771)
	var result hlc.Timestamp

	result.WallTime = int64(float64(t.lastTS.WallTime) * float64((zipfMax - t.zipf.Uint64())) / float64(zipfMax))
	return result
}
