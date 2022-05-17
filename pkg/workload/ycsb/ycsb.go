// Package ycsb is the workload specified by the Yahoo! Cloud Serving Benchmark.
package ycsb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"encoding/binary"
	"fmt"
	"hash"
	"hash/fnv"
	"math"
	"strings"
	"sync/atomic"

	"github.com/cockroachdb/cockroach-go/v2/crdb"
	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/bufalloc"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"github.com/cockroachdb/cockroach/pkg/workload/histogram"
	"github.com/cockroachdb/errors"
	"github.com/spf13/pflag"
	"golang.org/x/exp/rand"
)

const (
	numTableFields = 10
	fieldLength    = 100
	zipfIMin       = 0

	usertableSchemaRelational = `(
		ycsb_key VARCHAR(255) PRIMARY KEY NOT NULL,
		FIELD0 TEXT NOT NULL,
		FIELD1 TEXT NOT NULL,
		FIELD2 TEXT NOT NULL,
		FIELD3 TEXT NOT NULL,
		FIELD4 TEXT NOT NULL,
		FIELD5 TEXT NOT NULL,
		FIELD6 TEXT NOT NULL,
		FIELD7 TEXT NOT NULL,
		FIELD8 TEXT NOT NULL,
		FIELD9 TEXT NOT NULL
	)`
	usertableSchemaRelationalWithFamilies = `(
		ycsb_key VARCHAR(255) PRIMARY KEY NOT NULL,
		FIELD0 TEXT NOT NULL,
		FIELD1 TEXT NOT NULL,
		FIELD2 TEXT NOT NULL,
		FIELD3 TEXT NOT NULL,
		FIELD4 TEXT NOT NULL,
		FIELD5 TEXT NOT NULL,
		FIELD6 TEXT NOT NULL,
		FIELD7 TEXT NOT NULL,
		FIELD8 TEXT NOT NULL,
		FIELD9 TEXT NOT NULL,
		FAMILY (ycsb_key),
		FAMILY (FIELD0),
		FAMILY (FIELD1),
		FAMILY (FIELD2),
		FAMILY (FIELD3),
		FAMILY (FIELD4),
		FAMILY (FIELD5),
		FAMILY (FIELD6),
		FAMILY (FIELD7),
		FAMILY (FIELD8),
		FAMILY (FIELD9)
	)`
	usertableSchemaJSON = `(
		ycsb_key VARCHAR(255) PRIMARY KEY NOT NULL,
		FIELD JSONB
	)`

	timeFormatTemplate = `2006-01-02 15:04:05.000000-07:00`
)

type ycsb struct {
	flags     workload.Flags
	connFlags *workload.ConnFlags

	seed        uint64
	timeString  bool
	insertHash  bool
	zeroPadding int
	insertStart int
	insertCount int
	recordCount int
	json        bool
	families    bool
	sfu         bool
	splits      int

	workload                                                        string
	requestDistribution                                             string
	scanLengthDistribution                                          string
	minScanLength, maxScanLength                                    uint64
	readFreq, insertFreq, updateFreq, scanFreq, readModifyWriteFreq float32
}

func init() {
	workload.Register(ycsbMeta)
}

var ycsbMeta = workload.Meta{
	Name:         `ycsb`,
	Description:  `YCSB is the Yahoo! Cloud Serving Benchmark`,
	Version:      `1.0.0`,
	PublicFacing: true,
	New: func() workload.Generator {
		__antithesis_instrumentation__.Notify(699307)
		g := &ycsb{}
		g.flags.FlagSet = pflag.NewFlagSet(`ycsb`, pflag.ContinueOnError)
		g.flags.Meta = map[string]workload.FlagMeta{
			`workload`: {RuntimeOnly: true},
		}
		g.flags.Uint64Var(&g.seed, `seed`, 1, `Key hash seed.`)
		g.flags.BoolVar(&g.timeString, `time-string`, false, `Prepend field[0-9] data with current time in microsecond precision.`)
		g.flags.BoolVar(&g.insertHash, `insert-hash`, true, `Key to be hashed or ordered.`)
		g.flags.IntVar(&g.zeroPadding, `zero-padding`, 1, `Key using "insert-hash=false" has zeros padded to left to make this length of digits.`)
		g.flags.IntVar(&g.insertStart, `insert-start`, 0, `Key to start initial sequential insertions from. (default 0)`)
		g.flags.IntVar(&g.insertCount, `insert-count`, 10000, `Number of rows to sequentially insert before beginning workload.`)
		g.flags.IntVar(&g.recordCount, `record-count`, 0, `Key to start workload insertions from. Must be >= insert-start + insert-count. (Default: insert-start + insert-count)`)
		g.flags.BoolVar(&g.json, `json`, false, `Use JSONB rather than relational data`)
		g.flags.BoolVar(&g.families, `families`, true, `Place each column in its own column family`)
		g.flags.BoolVar(&g.sfu, `select-for-update`, true, `Use SELECT FOR UPDATE syntax in read-modify-write transactions`)
		g.flags.IntVar(&g.splits, `splits`, 0, `Number of splits to perform before starting normal operations`)
		g.flags.StringVar(&g.workload, `workload`, `B`, `Workload type. Choose from A-F.`)
		g.flags.StringVar(&g.requestDistribution, `request-distribution`, ``, `Distribution for request key generation [zipfian, uniform, latest]. The default for workloads A, B, C, E, and F is zipfian, and the default for workload D is latest.`)
		g.flags.StringVar(&g.scanLengthDistribution, `scan-length-distribution`, `uniform`, `Distribution for scan length generation [zipfian, uniform]. Primarily used for workload E.`)
		g.flags.Uint64Var(&g.minScanLength, `min-scan-length`, 1, `The minimum length for scan operations. Primarily used for workload E.`)
		g.flags.Uint64Var(&g.maxScanLength, `max-scan-length`, 1000, `The maximum length for scan operations. Primarily used for workload E.`)

		g.connFlags = workload.NewConnFlags(&g.flags)
		return g
	},
}

func (*ycsb) Meta() workload.Meta { __antithesis_instrumentation__.Notify(699308); return ycsbMeta }

func (g *ycsb) Flags() workload.Flags { __antithesis_instrumentation__.Notify(699309); return g.flags }

func (g *ycsb) Hooks() workload.Hooks {
	__antithesis_instrumentation__.Notify(699310)
	return workload.Hooks{
		Validate: func() error {
			__antithesis_instrumentation__.Notify(699311)
			g.workload = strings.ToUpper(g.workload)
			switch g.workload {
			case "A":
				__antithesis_instrumentation__.Notify(699316)
				g.readFreq = 0.5
				g.updateFreq = 0.5
				g.requestDistribution = "zipfian"
			case "B":
				__antithesis_instrumentation__.Notify(699317)
				g.readFreq = 0.95
				g.updateFreq = 0.05
				g.requestDistribution = "zipfian"
			case "C":
				__antithesis_instrumentation__.Notify(699318)
				g.readFreq = 1.0
				g.requestDistribution = "zipfian"
			case "D":
				__antithesis_instrumentation__.Notify(699319)
				g.readFreq = 0.95
				g.insertFreq = 0.05
				g.requestDistribution = "latest"
			case "E":
				__antithesis_instrumentation__.Notify(699320)
				g.scanFreq = 0.95
				g.insertFreq = 0.05
				g.requestDistribution = "zipfian"
			case "F":
				__antithesis_instrumentation__.Notify(699321)
				g.readFreq = 0.5
				g.readModifyWriteFreq = 0.5
				g.requestDistribution = "zipfian"
			default:
				__antithesis_instrumentation__.Notify(699322)
				return errors.Errorf("Unknown workload: %q", g.workload)
			}
			__antithesis_instrumentation__.Notify(699312)

			if !g.flags.Lookup(`families`).Changed {
				__antithesis_instrumentation__.Notify(699323)

				g.families = preferColumnFamilies(g.workload)
			} else {
				__antithesis_instrumentation__.Notify(699324)
			}
			__antithesis_instrumentation__.Notify(699313)

			if g.recordCount == 0 {
				__antithesis_instrumentation__.Notify(699325)
				g.recordCount = g.insertStart + g.insertCount
			} else {
				__antithesis_instrumentation__.Notify(699326)
			}
			__antithesis_instrumentation__.Notify(699314)
			if g.insertStart+g.insertCount > g.recordCount {
				__antithesis_instrumentation__.Notify(699327)
				return errors.Errorf("insertStart + insertCount (%d) must be <= recordCount (%d)", g.insertStart+g.insertCount, g.recordCount)
			} else {
				__antithesis_instrumentation__.Notify(699328)
			}
			__antithesis_instrumentation__.Notify(699315)
			return nil
		},
	}
}

func preferColumnFamilies(workload string) bool {
	__antithesis_instrumentation__.Notify(699329)

	switch workload {
	case "A":
		__antithesis_instrumentation__.Notify(699330)

		return true
	case "B":
		__antithesis_instrumentation__.Notify(699331)

		return true
	case "C":
		__antithesis_instrumentation__.Notify(699332)

		return false
	case "D":
		__antithesis_instrumentation__.Notify(699333)

		return false
	case "E":
		__antithesis_instrumentation__.Notify(699334)

		return false
	case "F":
		__antithesis_instrumentation__.Notify(699335)

		return true
	default:
		__antithesis_instrumentation__.Notify(699336)
		panic(fmt.Sprintf("unexpected workload: %s", workload))
	}
}

var usertableTypes = []*types.T{
	types.Bytes, types.Bytes, types.Bytes, types.Bytes, types.Bytes, types.Bytes,
	types.Bytes, types.Bytes, types.Bytes, types.Bytes, types.Bytes,
}

func (g *ycsb) Tables() []workload.Table {
	__antithesis_instrumentation__.Notify(699337)
	usertable := workload.Table{
		Name: `usertable`,
		Splits: workload.Tuples(
			g.splits,
			func(splitIdx int) []interface{} {
				__antithesis_instrumentation__.Notify(699340)
				step := math.MaxUint64 / uint64(g.splits+1)
				return []interface{}{
					keyNameFromHash(step * uint64(splitIdx+1)),
				}
			},
		),
	}
	__antithesis_instrumentation__.Notify(699338)
	if g.json {
		__antithesis_instrumentation__.Notify(699341)
		usertable.Schema = usertableSchemaJSON
		usertable.InitialRows = workload.Tuples(
			g.insertCount,
			func(rowIdx int) []interface{} {
				__antithesis_instrumentation__.Notify(699342)
				w := ycsbWorker{
					config:   g,
					hashFunc: fnv.New64(),
				}
				key := w.buildKeyName(uint64(g.insertStart + rowIdx))

				return []interface{}{key, "{}"}
			})
	} else {
		__antithesis_instrumentation__.Notify(699343)
		if g.families {
			__antithesis_instrumentation__.Notify(699345)
			usertable.Schema = usertableSchemaRelationalWithFamilies
		} else {
			__antithesis_instrumentation__.Notify(699346)
			usertable.Schema = usertableSchemaRelational
		}
		__antithesis_instrumentation__.Notify(699344)

		const batchSize = 1000
		usertable.InitialRows = workload.BatchedTuples{
			NumBatches: (g.insertCount + batchSize - 1) / batchSize,
			FillBatch: func(batchIdx int, cb coldata.Batch, _ *bufalloc.ByteAllocator) {
				__antithesis_instrumentation__.Notify(699347)
				rowBegin, rowEnd := batchIdx*batchSize, (batchIdx+1)*batchSize
				if rowEnd > g.insertCount {
					__antithesis_instrumentation__.Notify(699350)
					rowEnd = g.insertCount
				} else {
					__antithesis_instrumentation__.Notify(699351)
				}
				__antithesis_instrumentation__.Notify(699348)
				cb.Reset(usertableTypes, rowEnd-rowBegin, coldata.StandardColumnFactory)

				key := cb.ColVec(0).Bytes()

				key.Reset()

				var fields [numTableFields]*coldata.Bytes
				for i := range fields {
					__antithesis_instrumentation__.Notify(699352)
					fields[i] = cb.ColVec(i + 1).Bytes()

					fields[i].Reset()
				}
				__antithesis_instrumentation__.Notify(699349)

				w := ycsbWorker{
					config:   g,
					hashFunc: fnv.New64(),
				}
				rng := rand.NewSource(g.seed + uint64(batchIdx))

				var tmpbuf [fieldLength]byte
				for rowIdx := rowBegin; rowIdx < rowEnd; rowIdx++ {
					__antithesis_instrumentation__.Notify(699353)
					rowOffset := rowIdx - rowBegin

					key.Set(rowOffset, []byte(w.buildKeyName(uint64(rowIdx))))

					for i := range fields {
						__antithesis_instrumentation__.Notify(699354)
						randStringLetters(rng, tmpbuf[:])
						fields[i].Set(rowOffset, tmpbuf[:])
					}
				}
			},
		}
	}
	__antithesis_instrumentation__.Notify(699339)
	return []workload.Table{usertable}
}

func (g *ycsb) Ops(
	ctx context.Context, urls []string, reg *histogram.Registry,
) (workload.QueryLoad, error) {
	__antithesis_instrumentation__.Notify(699355)
	sqlDatabase, err := workload.SanitizeUrls(g, g.connFlags.DBOverride, urls)
	if err != nil {
		__antithesis_instrumentation__.Notify(699369)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(699370)
	}
	__antithesis_instrumentation__.Notify(699356)
	db, err := gosql.Open(`cockroach`, strings.Join(urls, ` `))
	if err != nil {
		__antithesis_instrumentation__.Notify(699371)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(699372)
	}
	__antithesis_instrumentation__.Notify(699357)

	db.SetMaxOpenConns(g.connFlags.Concurrency + 1)
	db.SetMaxIdleConns(g.connFlags.Concurrency + 1)

	readStmt, err := db.Prepare(`SELECT * FROM usertable WHERE ycsb_key = $1`)
	if err != nil {
		__antithesis_instrumentation__.Notify(699373)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(699374)
	}
	__antithesis_instrumentation__.Notify(699358)

	readFieldForUpdateStmts := make([]*gosql.Stmt, numTableFields)
	for i := 0; i < numTableFields; i++ {
		__antithesis_instrumentation__.Notify(699375)
		var q string
		if g.json {
			__antithesis_instrumentation__.Notify(699379)
			q = fmt.Sprintf(`SELECT field->>'field%d' FROM usertable WHERE ycsb_key = $1`, i)
		} else {
			__antithesis_instrumentation__.Notify(699380)
			q = fmt.Sprintf(`SELECT field%d FROM usertable WHERE ycsb_key = $1`, i)
		}
		__antithesis_instrumentation__.Notify(699376)
		if g.sfu {
			__antithesis_instrumentation__.Notify(699381)
			q = fmt.Sprintf(`%s FOR UPDATE`, q)
		} else {
			__antithesis_instrumentation__.Notify(699382)
		}
		__antithesis_instrumentation__.Notify(699377)

		stmt, err := db.Prepare(q)
		if err != nil {
			__antithesis_instrumentation__.Notify(699383)
			return workload.QueryLoad{}, err
		} else {
			__antithesis_instrumentation__.Notify(699384)
		}
		__antithesis_instrumentation__.Notify(699378)
		readFieldForUpdateStmts[i] = stmt
	}
	__antithesis_instrumentation__.Notify(699359)

	scanStmt, err := db.Prepare(`SELECT * FROM usertable WHERE ycsb_key >= $1 LIMIT $2`)
	if err != nil {
		__antithesis_instrumentation__.Notify(699385)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(699386)
	}
	__antithesis_instrumentation__.Notify(699360)

	var insertStmt *gosql.Stmt
	if g.json {
		__antithesis_instrumentation__.Notify(699387)
		insertStmt, err = db.Prepare(`INSERT INTO usertable VALUES ($1, json_build_object(
			'field0',  $2:::text,
			'field1',  $3:::text,
			'field2',  $4:::text,
			'field3',  $5:::text,
			'field4',  $6:::text,
			'field5',  $7:::text,
			'field6',  $8:::text,
			'field7',  $9:::text,
			'field8',  $10:::text,
			'field9',  $11:::text
		))`)
	} else {
		__antithesis_instrumentation__.Notify(699388)
		insertStmt, err = db.Prepare(`INSERT INTO usertable VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)`)
	}
	__antithesis_instrumentation__.Notify(699361)
	if err != nil {
		__antithesis_instrumentation__.Notify(699389)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(699390)
	}
	__antithesis_instrumentation__.Notify(699362)

	updateStmts := make([]*gosql.Stmt, numTableFields)
	if g.json {
		__antithesis_instrumentation__.Notify(699391)
		stmt, err := db.Prepare(`UPDATE usertable SET field = field || $2 WHERE ycsb_key = $1`)
		if err != nil {
			__antithesis_instrumentation__.Notify(699393)
			return workload.QueryLoad{}, err
		} else {
			__antithesis_instrumentation__.Notify(699394)
		}
		__antithesis_instrumentation__.Notify(699392)
		updateStmts[0] = stmt
	} else {
		__antithesis_instrumentation__.Notify(699395)
		for i := 0; i < numTableFields; i++ {
			__antithesis_instrumentation__.Notify(699396)
			q := fmt.Sprintf(`UPDATE usertable SET field%d = $2 WHERE ycsb_key = $1`, i)
			stmt, err := db.Prepare(q)
			if err != nil {
				__antithesis_instrumentation__.Notify(699398)
				return workload.QueryLoad{}, err
			} else {
				__antithesis_instrumentation__.Notify(699399)
			}
			__antithesis_instrumentation__.Notify(699397)
			updateStmts[i] = stmt
		}
	}
	__antithesis_instrumentation__.Notify(699363)

	rowIndexVal := uint64(g.recordCount)
	rowIndex := &rowIndexVal
	rowCounter := NewAcknowledgedCounter((uint64)(g.recordCount))

	var requestGen randGenerator
	requestGenRng := rand.New(rand.NewSource(g.seed))
	switch strings.ToLower(g.requestDistribution) {
	case "zipfian":
		__antithesis_instrumentation__.Notify(699400)
		requestGen, err = NewZipfGenerator(
			requestGenRng, zipfIMin, defaultIMax-1, defaultTheta, false)
	case "uniform":
		__antithesis_instrumentation__.Notify(699401)
		requestGen, err = NewUniformGenerator(requestGenRng, 0, uint64(g.recordCount)-1)
	case "latest":
		__antithesis_instrumentation__.Notify(699402)
		requestGen, err = NewSkewedLatestGenerator(
			requestGenRng, zipfIMin, uint64(g.recordCount)-1, defaultTheta, false)
	default:
		__antithesis_instrumentation__.Notify(699403)
		return workload.QueryLoad{}, errors.Errorf("Unknown request distribution: %s", g.requestDistribution)
	}
	__antithesis_instrumentation__.Notify(699364)
	if err != nil {
		__antithesis_instrumentation__.Notify(699404)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(699405)
	}
	__antithesis_instrumentation__.Notify(699365)

	var scanLengthGen randGenerator
	scanLengthGenRng := rand.New(rand.NewSource(g.seed + 1))
	switch strings.ToLower(g.scanLengthDistribution) {
	case "zipfian":
		__antithesis_instrumentation__.Notify(699406)
		scanLengthGen, err = NewZipfGenerator(scanLengthGenRng, g.minScanLength, g.maxScanLength, defaultTheta, false)
	case "uniform":
		__antithesis_instrumentation__.Notify(699407)
		scanLengthGen, err = NewUniformGenerator(scanLengthGenRng, g.minScanLength, g.maxScanLength)
	default:
		__antithesis_instrumentation__.Notify(699408)
		return workload.QueryLoad{}, errors.Errorf("Unknown scan length distribution: %s", g.scanLengthDistribution)
	}
	__antithesis_instrumentation__.Notify(699366)
	if err != nil {
		__antithesis_instrumentation__.Notify(699409)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(699410)
	}
	__antithesis_instrumentation__.Notify(699367)

	ql := workload.QueryLoad{SQLDatabase: sqlDatabase}
	for i := 0; i < g.connFlags.Concurrency; i++ {
		__antithesis_instrumentation__.Notify(699411)
		rng := rand.New(rand.NewSource(g.seed + uint64(i)))
		w := &ycsbWorker{
			config:                  g,
			hists:                   reg.GetHandle(),
			db:                      db,
			readStmt:                readStmt,
			readFieldForUpdateStmts: readFieldForUpdateStmts,
			scanStmt:                scanStmt,
			insertStmt:              insertStmt,
			updateStmts:             updateStmts,
			rowIndex:                rowIndex,
			rowCounter:              rowCounter,
			nextInsertIndex:         nil,
			requestGen:              requestGen,
			scanLengthGen:           scanLengthGen,
			rng:                     rng,
			hashFunc:                fnv.New64(),
		}
		ql.WorkerFns = append(ql.WorkerFns, w.run)
	}
	__antithesis_instrumentation__.Notify(699368)
	return ql, nil
}

type randGenerator interface {
	Uint64() uint64
	IncrementIMax(count uint64) error
}

type ycsbWorker struct {
	config *ycsb
	hists  *histogram.Histograms
	db     *gosql.DB

	readStmt *gosql.Stmt

	readFieldForUpdateStmts []*gosql.Stmt
	scanStmt, insertStmt    *gosql.Stmt

	updateStmts []*gosql.Stmt

	rowIndex *uint64

	rowCounter *AcknowledgedCounter

	nextInsertIndex *uint64

	requestGen    randGenerator
	scanLengthGen randGenerator
	rng           *rand.Rand
	hashFunc      hash.Hash64
	hashBuf       [8]byte
}

func (yw *ycsbWorker) run(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(699412)
	op := yw.chooseOp()
	var err error

	start := timeutil.Now()
	switch op {
	case updateOp:
		__antithesis_instrumentation__.Notify(699415)
		err = yw.updateRow(ctx)
	case readOp:
		__antithesis_instrumentation__.Notify(699416)
		err = yw.readRow(ctx)
	case insertOp:
		__antithesis_instrumentation__.Notify(699417)
		err = yw.insertRow(ctx)
	case scanOp:
		__antithesis_instrumentation__.Notify(699418)
		err = yw.scanRows(ctx)
	case readModifyWriteOp:
		__antithesis_instrumentation__.Notify(699419)
		err = yw.readModifyWriteRow(ctx)
	default:
		__antithesis_instrumentation__.Notify(699420)
		return errors.Errorf(`unknown operation: %s`, op)
	}
	__antithesis_instrumentation__.Notify(699413)
	if err != nil {
		__antithesis_instrumentation__.Notify(699421)
		return err
	} else {
		__antithesis_instrumentation__.Notify(699422)
	}
	__antithesis_instrumentation__.Notify(699414)

	elapsed := timeutil.Since(start)
	yw.hists.Get(string(op)).Record(elapsed)
	return nil
}

var readOnly int32

type operation string

const (
	updateOp          operation = `update`
	insertOp          operation = `insert`
	readOp            operation = `read`
	scanOp            operation = `scan`
	readModifyWriteOp operation = `readModifyWrite`
)

func (yw *ycsbWorker) hashKey(key uint64) uint64 {
	__antithesis_instrumentation__.Notify(699423)
	yw.hashBuf = [8]byte{}
	binary.PutUvarint(yw.hashBuf[:], key)
	yw.hashFunc.Reset()
	if _, err := yw.hashFunc.Write(yw.hashBuf[:]); err != nil {
		__antithesis_instrumentation__.Notify(699425)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(699426)
	}
	__antithesis_instrumentation__.Notify(699424)
	return yw.hashFunc.Sum64()
}

func (yw *ycsbWorker) buildKeyName(keynum uint64) string {
	__antithesis_instrumentation__.Notify(699427)
	if yw.config.insertHash {
		__antithesis_instrumentation__.Notify(699429)
		return keyNameFromHash(yw.hashKey(keynum))
	} else {
		__antithesis_instrumentation__.Notify(699430)
	}
	__antithesis_instrumentation__.Notify(699428)
	return keyNameFromOrder(keynum, yw.config.zeroPadding)
}

func keyNameFromHash(hashedKey uint64) string {
	__antithesis_instrumentation__.Notify(699431)
	return fmt.Sprintf("user%d", hashedKey)
}

func keyNameFromOrder(keynum uint64, zeroPadding int) string {
	__antithesis_instrumentation__.Notify(699432)
	return fmt.Sprintf("user%0*d", zeroPadding, keynum)
}

func (yw *ycsbWorker) nextReadKey() string {
	__antithesis_instrumentation__.Notify(699433)
	rowCount := yw.rowCounter.Last()

	rowIndex := yw.requestGen.Uint64() % rowCount
	return yw.buildKeyName(rowIndex)
}

func (yw *ycsbWorker) nextInsertKeyIndex() uint64 {
	__antithesis_instrumentation__.Notify(699434)
	if yw.nextInsertIndex != nil {
		__antithesis_instrumentation__.Notify(699436)
		result := *yw.nextInsertIndex
		yw.nextInsertIndex = nil
		return result
	} else {
		__antithesis_instrumentation__.Notify(699437)
	}
	__antithesis_instrumentation__.Notify(699435)
	return atomic.AddUint64(yw.rowIndex, 1) - 1
}

var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func (yw *ycsbWorker) randString(length int) string {
	__antithesis_instrumentation__.Notify(699438)
	str := make([]byte, length)

	strStart := 0
	if yw.config.timeString {
		__antithesis_instrumentation__.Notify(699441)
		currentTime := timeutil.Now().UTC()
		str = currentTime.AppendFormat(str[:0], timeFormatTemplate)
		strStart = len(str)
		str = str[:length]
	} else {
		__antithesis_instrumentation__.Notify(699442)
	}
	__antithesis_instrumentation__.Notify(699439)

	for i := strStart; i < length; i++ {
		__antithesis_instrumentation__.Notify(699443)
		str[i] = letters[yw.rng.Intn(len(letters))]
	}
	__antithesis_instrumentation__.Notify(699440)
	return string(str)
}

func randStringLetters(rng rand.Source, buf []byte) {
	__antithesis_instrumentation__.Notify(699444)
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const lettersLen = uint64(len(letters))
	const lettersCharsPerRand = uint64(11)

	var r, charsLeft uint64
	for i := 0; i < len(buf); i++ {
		__antithesis_instrumentation__.Notify(699445)
		if charsLeft == 0 {
			__antithesis_instrumentation__.Notify(699447)
			r = rng.Uint64()
			charsLeft = lettersCharsPerRand
		} else {
			__antithesis_instrumentation__.Notify(699448)
		}
		__antithesis_instrumentation__.Notify(699446)
		buf[i] = letters[r%lettersLen]
		r = r / lettersLen
		charsLeft--
	}
}

func (yw *ycsbWorker) insertRow(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(699449)
	var args [numTableFields + 1]interface{}
	keyIndex := yw.nextInsertKeyIndex()
	args[0] = yw.buildKeyName(keyIndex)
	for i := 1; i <= numTableFields; i++ {
		__antithesis_instrumentation__.Notify(699453)
		args[i] = yw.randString(fieldLength)
	}
	__antithesis_instrumentation__.Notify(699450)
	if _, err := yw.insertStmt.ExecContext(ctx, args[:]...); err != nil {
		__antithesis_instrumentation__.Notify(699454)
		yw.nextInsertIndex = new(uint64)
		*yw.nextInsertIndex = keyIndex
		return err
	} else {
		__antithesis_instrumentation__.Notify(699455)
	}
	__antithesis_instrumentation__.Notify(699451)

	count, err := yw.rowCounter.Acknowledge(keyIndex)
	if err != nil {
		__antithesis_instrumentation__.Notify(699456)
		return err
	} else {
		__antithesis_instrumentation__.Notify(699457)
	}
	__antithesis_instrumentation__.Notify(699452)
	return yw.requestGen.IncrementIMax(count)
}

func (yw *ycsbWorker) updateRow(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(699458)
	var stmt *gosql.Stmt
	var args [2]interface{}
	args[0] = yw.nextReadKey()
	fieldIdx := yw.rng.Intn(numTableFields)
	value := yw.randString(fieldLength)
	if yw.config.json {
		__antithesis_instrumentation__.Notify(699461)
		stmt = yw.updateStmts[0]
		args[1] = fmt.Sprintf(`{"field%d": "%s"}`, fieldIdx, value)
	} else {
		__antithesis_instrumentation__.Notify(699462)
		stmt = yw.updateStmts[fieldIdx]
		args[1] = value
	}
	__antithesis_instrumentation__.Notify(699459)
	if _, err := stmt.ExecContext(ctx, args[:]...); err != nil {
		__antithesis_instrumentation__.Notify(699463)
		return err
	} else {
		__antithesis_instrumentation__.Notify(699464)
	}
	__antithesis_instrumentation__.Notify(699460)
	return nil
}

func (yw *ycsbWorker) readRow(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(699465)
	key := yw.nextReadKey()
	res, err := yw.readStmt.QueryContext(ctx, key)
	if err != nil {
		__antithesis_instrumentation__.Notify(699468)
		return err
	} else {
		__antithesis_instrumentation__.Notify(699469)
	}
	__antithesis_instrumentation__.Notify(699466)
	defer res.Close()
	for res.Next() {
		__antithesis_instrumentation__.Notify(699470)
	}
	__antithesis_instrumentation__.Notify(699467)
	return res.Err()
}

func (yw *ycsbWorker) scanRows(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(699471)
	key := yw.nextReadKey()
	scanLength := yw.scanLengthGen.Uint64()

	return crdb.Execute(func() error {
		__antithesis_instrumentation__.Notify(699472)
		res, err := yw.scanStmt.QueryContext(ctx, key, scanLength)
		if err != nil {
			__antithesis_instrumentation__.Notify(699475)
			return err
		} else {
			__antithesis_instrumentation__.Notify(699476)
		}
		__antithesis_instrumentation__.Notify(699473)
		defer res.Close()
		for res.Next() {
			__antithesis_instrumentation__.Notify(699477)
		}
		__antithesis_instrumentation__.Notify(699474)
		return res.Err()
	})
}

func (yw *ycsbWorker) readModifyWriteRow(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(699478)
	key := yw.nextReadKey()
	newValue := yw.randString(fieldLength)
	fieldIdx := yw.rng.Intn(numTableFields)
	var args [2]interface{}
	args[0] = key
	err := crdb.ExecuteTx(ctx, yw.db, nil, func(tx *gosql.Tx) error {
		__antithesis_instrumentation__.Notify(699481)
		var oldValue []byte
		readStmt := yw.readFieldForUpdateStmts[fieldIdx]
		if err := tx.StmtContext(ctx, readStmt).QueryRowContext(ctx, key).Scan(&oldValue); err != nil {
			__antithesis_instrumentation__.Notify(699484)
			return err
		} else {
			__antithesis_instrumentation__.Notify(699485)
		}
		__antithesis_instrumentation__.Notify(699482)
		var updateStmt *gosql.Stmt
		if yw.config.json {
			__antithesis_instrumentation__.Notify(699486)
			updateStmt = yw.updateStmts[0]
			args[1] = fmt.Sprintf(`{"field%d": "%s"}`, fieldIdx, newValue)
		} else {
			__antithesis_instrumentation__.Notify(699487)
			updateStmt = yw.updateStmts[fieldIdx]
			args[1] = newValue
		}
		__antithesis_instrumentation__.Notify(699483)
		_, err := tx.StmtContext(ctx, updateStmt).ExecContext(ctx, args[:]...)
		return err
	})
	__antithesis_instrumentation__.Notify(699479)
	if errors.Is(err, gosql.ErrNoRows) && func() bool {
		__antithesis_instrumentation__.Notify(699488)
		return ctx.Err() != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(699489)

		return ctx.Err()
	} else {
		__antithesis_instrumentation__.Notify(699490)
	}
	__antithesis_instrumentation__.Notify(699480)
	return err
}

func (yw *ycsbWorker) chooseOp() operation {
	__antithesis_instrumentation__.Notify(699491)
	p := yw.rng.Float32()
	if atomic.LoadInt32(&readOnly) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(699496)
		return p <= yw.config.updateFreq == true
	}() == true {
		__antithesis_instrumentation__.Notify(699497)
		return updateOp
	} else {
		__antithesis_instrumentation__.Notify(699498)
	}
	__antithesis_instrumentation__.Notify(699492)
	p -= yw.config.updateFreq
	if atomic.LoadInt32(&readOnly) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(699499)
		return p <= yw.config.insertFreq == true
	}() == true {
		__antithesis_instrumentation__.Notify(699500)
		return insertOp
	} else {
		__antithesis_instrumentation__.Notify(699501)
	}
	__antithesis_instrumentation__.Notify(699493)
	p -= yw.config.insertFreq
	if atomic.LoadInt32(&readOnly) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(699502)
		return p <= yw.config.readModifyWriteFreq == true
	}() == true {
		__antithesis_instrumentation__.Notify(699503)
		return readModifyWriteOp
	} else {
		__antithesis_instrumentation__.Notify(699504)
	}
	__antithesis_instrumentation__.Notify(699494)
	p -= yw.config.readModifyWriteFreq

	if p <= yw.config.scanFreq {
		__antithesis_instrumentation__.Notify(699505)
		return scanOp
	} else {
		__antithesis_instrumentation__.Notify(699506)
	}
	__antithesis_instrumentation__.Notify(699495)
	return readOp
}
