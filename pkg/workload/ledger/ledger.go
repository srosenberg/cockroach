package ledger

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"hash/fnv"
	"math/rand"
	"strings"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"github.com/cockroachdb/cockroach/pkg/workload/histogram"
	"github.com/spf13/pflag"
)

type ledger struct {
	flags     workload.Flags
	connFlags *workload.ConnFlags

	seed              int64
	customers         int
	inlineArgs        bool
	splits            int
	fks               bool
	historicalBalance bool
	mix               string

	txs  []tx
	deck []int

	reg      *histogram.Registry
	rngPool  *sync.Pool
	hashPool *sync.Pool
}

func init() {
	workload.Register(ledgerMeta)
}

var ledgerMeta = workload.Meta{
	Name:        `ledger`,
	Description: `Ledger simulates an accounting system using double-entry bookkeeping`,
	Version:     `1.0.0`,
	New: func() workload.Generator {
		__antithesis_instrumentation__.Notify(694617)
		g := &ledger{}
		g.flags.FlagSet = pflag.NewFlagSet(`ledger`, pflag.ContinueOnError)
		g.connFlags = workload.NewConnFlags(&g.flags)
		g.flags.Int64Var(&g.seed, `seed`, 1, `Random number generator seed`)
		g.flags.IntVar(&g.customers, `customers`, 1000, `Number of customers`)
		g.flags.BoolVar(&g.inlineArgs, `inline-args`, false, `Use inline query arguments`)
		g.flags.IntVar(&g.splits, `splits`, 0, `Number of splits to perform before starting normal operations`)
		g.flags.BoolVar(&g.fks, `fks`, true, `Add the foreign keys`)
		g.flags.BoolVar(&g.historicalBalance, `historical-balance`, false, `Perform balance txns using historical reads`)
		g.flags.StringVar(&g.mix, `mix`,
			`balance=50,withdrawal=37,deposit=12,reversal=0`,
			`Weights for the transaction mix.`)
		return g
	},
}

func FromFlags(flags ...string) workload.Generator {
	__antithesis_instrumentation__.Notify(694618)
	return workload.FromFlags(ledgerMeta, flags...)
}

func (*ledger) Meta() workload.Meta { __antithesis_instrumentation__.Notify(694619); return ledgerMeta }

func (w *ledger) Flags() workload.Flags {
	__antithesis_instrumentation__.Notify(694620)
	return w.flags
}

func (w *ledger) Hooks() workload.Hooks {
	__antithesis_instrumentation__.Notify(694621)
	return workload.Hooks{
		Validate: func() error {
			__antithesis_instrumentation__.Notify(694622)
			return initializeMix(w)
		},
		PostLoad: func(sqlDB *gosql.DB) error {
			__antithesis_instrumentation__.Notify(694623)
			if w.fks {
				__antithesis_instrumentation__.Notify(694625)
				fkStmts := []string{
					`create index entry_auto_index_fk_customer on entry (customer_id ASC)`,
					`create index entry_auto_index_fk_transaction on entry (transaction_id ASC)`,
					`alter table entry add foreign key (customer_id) references customer (id)`,
					`alter table entry add foreign key (transaction_id) references transaction (external_id)`,
				}
				for _, fkStmt := range fkStmts {
					__antithesis_instrumentation__.Notify(694626)
					if _, err := sqlDB.Exec(fkStmt); err != nil {
						__antithesis_instrumentation__.Notify(694627)
						return err
					} else {
						__antithesis_instrumentation__.Notify(694628)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(694629)
			}
			__antithesis_instrumentation__.Notify(694624)
			return nil
		},
	}
}

func (w *ledger) Tables() []workload.Table {
	__antithesis_instrumentation__.Notify(694630)
	if w.rngPool == nil {
		__antithesis_instrumentation__.Notify(694633)
		w.rngPool = &sync.Pool{
			New: func() interface{} {
				__antithesis_instrumentation__.Notify(694634)
				return rand.New(rand.NewSource(timeutil.Now().UnixNano()))
			},
		}
	} else {
		__antithesis_instrumentation__.Notify(694635)
	}
	__antithesis_instrumentation__.Notify(694631)
	if w.hashPool == nil {
		__antithesis_instrumentation__.Notify(694636)
		w.hashPool = &sync.Pool{
			New: func() interface{} { __antithesis_instrumentation__.Notify(694637); return fnv.New64() },
		}
	} else {
		__antithesis_instrumentation__.Notify(694638)
	}
	__antithesis_instrumentation__.Notify(694632)

	customer := workload.Table{
		Name:   `customer`,
		Schema: ledgerCustomerSchema,
		InitialRows: workload.TypedTuples(
			w.customers,
			ledgerCustomerTypes,
			w.ledgerCustomerInitialRow,
		),
		Splits: workload.Tuples(
			numTxnsPerCustomer*w.splits,
			w.ledgerCustomerSplitRow,
		),
	}
	transaction := workload.Table{
		Name:   `transaction`,
		Schema: ledgerTransactionSchema,
		InitialRows: workload.TypedTuples(
			numTxnsPerCustomer*w.customers,
			ledgerTransactionColTypes,
			w.ledgerTransactionInitialRow,
		),
		Splits: workload.Tuples(
			w.splits,
			w.ledgerTransactionSplitRow,
		),
	}
	entry := workload.Table{
		Name:   `entry`,
		Schema: ledgerEntrySchema,
		InitialRows: workload.Tuples(
			numEntriesPerCustomer*w.customers,
			w.ledgerEntryInitialRow,
		),
		Splits: workload.Tuples(
			numEntriesPerCustomer*w.splits,
			w.ledgerEntrySplitRow,
		),
	}
	session := workload.Table{
		Name:   `session`,
		Schema: ledgerSessionSchema,
		InitialRows: workload.Tuples(
			w.customers,
			w.ledgerSessionInitialRow,
		),
		Splits: workload.Tuples(
			w.splits,
			w.ledgerSessionSplitRow,
		),
	}
	return []workload.Table{
		customer, transaction, entry, session,
	}
}

func (w *ledger) Ops(
	ctx context.Context, urls []string, reg *histogram.Registry,
) (workload.QueryLoad, error) {
	__antithesis_instrumentation__.Notify(694639)
	sqlDatabase, err := workload.SanitizeUrls(w, w.connFlags.DBOverride, urls)
	if err != nil {
		__antithesis_instrumentation__.Notify(694643)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(694644)
	}
	__antithesis_instrumentation__.Notify(694640)
	db, err := gosql.Open(`cockroach`, strings.Join(urls, ` `))
	if err != nil {
		__antithesis_instrumentation__.Notify(694645)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(694646)
	}
	__antithesis_instrumentation__.Notify(694641)

	db.SetMaxOpenConns(w.connFlags.Concurrency + 1)
	db.SetMaxIdleConns(w.connFlags.Concurrency + 1)

	w.reg = reg
	ql := workload.QueryLoad{SQLDatabase: sqlDatabase}
	now := timeutil.Now().UnixNano()
	for i := 0; i < w.connFlags.Concurrency; i++ {
		__antithesis_instrumentation__.Notify(694647)
		worker := &worker{
			config:   w,
			hists:    reg.GetHandle(),
			db:       db,
			rng:      rand.New(rand.NewSource(now + int64(i))),
			deckPerm: append([]int(nil), w.deck...),
			permIdx:  len(w.deck),
		}
		ql.WorkerFns = append(ql.WorkerFns, worker.run)
	}
	__antithesis_instrumentation__.Notify(694642)
	return ql, nil
}
