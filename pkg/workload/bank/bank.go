package bank

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"strings"

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
	bankSchema = `(
		id INT PRIMARY KEY,
		balance INT,
		payload STRING,
		FAMILY (id, balance, payload)
	)`

	defaultRows         = 1000
	defaultBatchSize    = 1000
	defaultPayloadBytes = 100
	defaultRanges       = 10
	maxTransfer         = 999
)

type bank struct {
	flags     workload.Flags
	connFlags *workload.ConnFlags

	seed                 uint64
	rows, batchSize      int
	payloadBytes, ranges int
}

func init() {
	workload.Register(bankMeta)
}

var bankMeta = workload.Meta{
	Name:         `bank`,
	Description:  `Bank models a set of accounts with currency balances`,
	Version:      `1.0.0`,
	PublicFacing: true,
	New: func() workload.Generator {
		__antithesis_instrumentation__.Notify(693547)
		g := &bank{}
		g.flags.FlagSet = pflag.NewFlagSet(`bank`, pflag.ContinueOnError)
		g.flags.Meta = map[string]workload.FlagMeta{
			`batch-size`: {RuntimeOnly: true},
		}
		g.flags.Uint64Var(&g.seed, `seed`, 1, `Key hash seed.`)
		g.flags.IntVar(&g.rows, `rows`, defaultRows, `Initial number of accounts in bank table.`)
		g.flags.IntVar(&g.batchSize, `batch-size`, defaultBatchSize, `Number of rows in each batch of initial data.`)
		g.flags.IntVar(&g.payloadBytes, `payload-bytes`, defaultPayloadBytes, `Size of the payload field in each initial row.`)
		g.flags.IntVar(&g.ranges, `ranges`, defaultRanges, `Initial number of ranges in bank table.`)
		g.connFlags = workload.NewConnFlags(&g.flags)
		return g
	},
}

func FromRows(rows int) workload.Generator {
	__antithesis_instrumentation__.Notify(693548)
	return FromConfig(rows, 1, defaultPayloadBytes, defaultRanges)
}

func FromConfig(rows int, batchSize int, payloadBytes int, ranges int) workload.Generator {
	__antithesis_instrumentation__.Notify(693549)
	if ranges > rows {
		__antithesis_instrumentation__.Notify(693552)
		ranges = rows
	} else {
		__antithesis_instrumentation__.Notify(693553)
	}
	__antithesis_instrumentation__.Notify(693550)
	if batchSize <= 0 {
		__antithesis_instrumentation__.Notify(693554)
		batchSize = defaultBatchSize
	} else {
		__antithesis_instrumentation__.Notify(693555)
	}
	__antithesis_instrumentation__.Notify(693551)
	return workload.FromFlags(bankMeta,
		fmt.Sprintf(`--rows=%d`, rows),
		fmt.Sprintf(`--batch-size=%d`, batchSize),
		fmt.Sprintf(`--payload-bytes=%d`, payloadBytes),
		fmt.Sprintf(`--ranges=%d`, ranges),
	)
}

func (*bank) Meta() workload.Meta { __antithesis_instrumentation__.Notify(693556); return bankMeta }

func (b *bank) Flags() workload.Flags { __antithesis_instrumentation__.Notify(693557); return b.flags }

func (b *bank) Hooks() workload.Hooks {
	__antithesis_instrumentation__.Notify(693558)
	return workload.Hooks{
		Validate: func() error {
			__antithesis_instrumentation__.Notify(693559)
			if b.rows < b.ranges {
				__antithesis_instrumentation__.Notify(693562)
				return errors.Errorf(
					"Value of 'rows' (%d) must be greater than or equal to value of 'ranges' (%d)",
					b.rows, b.ranges)
			} else {
				__antithesis_instrumentation__.Notify(693563)
			}
			__antithesis_instrumentation__.Notify(693560)
			if b.batchSize <= 0 {
				__antithesis_instrumentation__.Notify(693564)
				return errors.Errorf(`Value of batch-size must be greater than zero; was %d`, b.batchSize)
			} else {
				__antithesis_instrumentation__.Notify(693565)
			}
			__antithesis_instrumentation__.Notify(693561)
			return nil
		},
	}
}

var bankTypes = []*types.T{
	types.Int,
	types.Int,
	types.Bytes,
}

func (b *bank) Tables() []workload.Table {
	__antithesis_instrumentation__.Notify(693566)
	numBatches := (b.rows + b.batchSize - 1) / b.batchSize
	table := workload.Table{
		Name:   `bank`,
		Schema: bankSchema,
		InitialRows: workload.BatchedTuples{
			NumBatches: numBatches,
			FillBatch: func(batchIdx int, cb coldata.Batch, a *bufalloc.ByteAllocator) {
				__antithesis_instrumentation__.Notify(693568)
				rng := rand.NewSource(b.seed + uint64(batchIdx))

				rowBegin, rowEnd := batchIdx*b.batchSize, (batchIdx+1)*b.batchSize
				if rowEnd > b.rows {
					__antithesis_instrumentation__.Notify(693570)
					rowEnd = b.rows
				} else {
					__antithesis_instrumentation__.Notify(693571)
				}
				__antithesis_instrumentation__.Notify(693569)
				cb.Reset(bankTypes, rowEnd-rowBegin, coldata.StandardColumnFactory)
				idCol := cb.ColVec(0).Int64()
				balanceCol := cb.ColVec(1).Int64()
				payloadCol := cb.ColVec(2).Bytes()

				payloadCol.Reset()
				for rowIdx := rowBegin; rowIdx < rowEnd; rowIdx++ {
					__antithesis_instrumentation__.Notify(693572)
					var payload []byte
					*a, payload = a.Alloc(b.payloadBytes, 0)
					const initialPrefix = `initial-`
					copy(payload[:len(initialPrefix)], []byte(initialPrefix))
					randStringLetters(rng, payload[len(initialPrefix):])

					rowOffset := rowIdx - rowBegin
					idCol[rowOffset] = int64(rowIdx)
					balanceCol[rowOffset] = 0
					payloadCol.Set(rowOffset, payload)
				}
			},
		},
		Splits: workload.Tuples(
			b.ranges-1,
			func(splitIdx int) []interface{} {
				__antithesis_instrumentation__.Notify(693573)
				return []interface{}{
					(splitIdx + 1) * (b.rows / b.ranges),
				}
			},
		),
	}
	__antithesis_instrumentation__.Notify(693567)
	return []workload.Table{table}
}

func (b *bank) Ops(
	ctx context.Context, urls []string, reg *histogram.Registry,
) (workload.QueryLoad, error) {
	__antithesis_instrumentation__.Notify(693574)
	sqlDatabase, err := workload.SanitizeUrls(b, b.connFlags.DBOverride, urls)
	if err != nil {
		__antithesis_instrumentation__.Notify(693579)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(693580)
	}
	__antithesis_instrumentation__.Notify(693575)
	db, err := gosql.Open(`cockroach`, strings.Join(urls, ` `))
	if err != nil {
		__antithesis_instrumentation__.Notify(693581)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(693582)
	}
	__antithesis_instrumentation__.Notify(693576)

	db.SetMaxOpenConns(b.connFlags.Concurrency + 1)
	db.SetMaxIdleConns(b.connFlags.Concurrency + 1)

	updateStmt, err := db.Prepare(`
		UPDATE bank
		SET balance = CASE id WHEN $1 THEN balance-$3 WHEN $2 THEN balance+$3 END
		WHERE id IN ($1, $2)
	`)
	if err != nil {
		__antithesis_instrumentation__.Notify(693583)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(693584)
	}
	__antithesis_instrumentation__.Notify(693577)

	ql := workload.QueryLoad{SQLDatabase: sqlDatabase}
	for i := 0; i < b.connFlags.Concurrency; i++ {
		__antithesis_instrumentation__.Notify(693585)
		rng := rand.New(rand.NewSource(b.seed))
		hists := reg.GetHandle()
		workerFn := func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(693587)
			from := rng.Intn(b.rows)
			to := rng.Intn(b.rows - 1)
			for from == to && func() bool {
				__antithesis_instrumentation__.Notify(693589)
				return b.rows != 1 == true
			}() == true {
				__antithesis_instrumentation__.Notify(693590)
				to = rng.Intn(b.rows - 1)
			}
			__antithesis_instrumentation__.Notify(693588)
			amount := rand.Intn(maxTransfer)
			start := timeutil.Now()
			_, err := updateStmt.Exec(from, to, amount)
			elapsed := timeutil.Since(start)
			hists.Get(`transfer`).Record(elapsed)
			return err
		}
		__antithesis_instrumentation__.Notify(693586)
		ql.WorkerFns = append(ql.WorkerFns, workerFn)
	}
	__antithesis_instrumentation__.Notify(693578)
	return ql, nil
}

func randStringLetters(rng rand.Source, buf []byte) {
	__antithesis_instrumentation__.Notify(693591)
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const lettersLen = uint64(len(letters))
	const lettersCharsPerRand = uint64(11)

	var r, charsLeft uint64
	for i := 0; i < len(buf); i++ {
		__antithesis_instrumentation__.Notify(693592)
		if charsLeft == 0 {
			__antithesis_instrumentation__.Notify(693594)
			r = rng.Uint64()
			charsLeft = lettersCharsPerRand
		} else {
			__antithesis_instrumentation__.Notify(693595)
		}
		__antithesis_instrumentation__.Notify(693593)
		buf[i] = letters[r%lettersLen]
		r = r / lettersLen
		charsLeft--
	}
}
