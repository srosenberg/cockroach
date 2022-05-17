package tpch

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"github.com/cockroachdb/cockroach/pkg/workload/histogram"
	"github.com/cockroachdb/errors"
	"github.com/spf13/pflag"
	"golang.org/x/exp/rand"
)

const (
	numNation           = 25
	numRegion           = 5
	numPartPerSF        = 200000
	numPartSuppPerPart  = 4
	numSupplierPerSF    = 10000
	numCustomerPerSF    = 150000
	numOrderPerCustomer = 10
	numLineItemPerSF    = 6001215
)

type wrongOutputError struct {
	error
}

const TPCHWrongOutputErrorPrefix = "TPCH wrong output "

func (e wrongOutputError) Error() string {
	__antithesis_instrumentation__.Notify(698820)
	return TPCHWrongOutputErrorPrefix + e.error.Error()
}

type tpch struct {
	flags     workload.Flags
	connFlags *workload.ConnFlags

	seed        uint64
	scaleFactor int
	fks         bool

	enableChecks               bool
	vectorize                  string
	useClusterVectorizeSetting bool
	verbose                    bool

	queriesRaw      string
	selectedQueries []int

	textPool   textPool
	localsPool *sync.Pool
}

func init() {
	workload.Register(tpchMeta)
}

func FromScaleFactor(scaleFactor int) workload.Generator {
	__antithesis_instrumentation__.Notify(698821)
	return workload.FromFlags(tpchMeta, fmt.Sprintf(`--scale-factor=%d`, scaleFactor))
}

var tpchMeta = workload.Meta{
	Name:        `tpch`,
	Description: `TPC-H is a read-only workload of "analytics" queries on large datasets.`,
	Version:     `1.0.0`,
	New: func() workload.Generator {
		__antithesis_instrumentation__.Notify(698822)
		g := &tpch{}
		g.flags.FlagSet = pflag.NewFlagSet(`tpch`, pflag.ContinueOnError)
		g.flags.Meta = map[string]workload.FlagMeta{
			`queries`:       {RuntimeOnly: true},
			`dist-sql`:      {RuntimeOnly: true},
			`enable-checks`: {RuntimeOnly: true},
			`vectorize`:     {RuntimeOnly: true},
		}
		g.flags.Uint64Var(&g.seed, `seed`, 1, `Random number generator seed`)
		g.flags.IntVar(&g.scaleFactor, `scale-factor`, 1,
			`Linear scale of how much data to use (each SF is ~1GB)`)
		g.flags.BoolVar(&g.fks, `fks`, true, `Add the foreign keys`)
		g.flags.StringVar(&g.queriesRaw, `queries`,
			`1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22`,
			`Queries to run. Use a comma separated list of query numbers`)
		g.flags.BoolVar(&g.enableChecks, `enable-checks`, false,
			"Enable checking the output against the expected rows (default false). "+
				"Note that the checks are only supported for scale factor 1 of the backup "+
				"stored at 'gs://cockroach-fixtures/workload/tpch/scalefactor=1/backup'")
		g.flags.StringVar(&g.vectorize, `vectorize`, `on`,
			`Set vectorize session variable`)
		g.flags.BoolVar(&g.useClusterVectorizeSetting, `default-vectorize`, false,
			`Ignore vectorize option and use the current cluster setting sql.defaults.vectorize`)
		g.flags.BoolVar(&g.verbose, `verbose`, false,
			`Prints out the queries being run as well as histograms`)
		g.connFlags = workload.NewConnFlags(&g.flags)
		return g
	},
}

func (*tpch) Meta() workload.Meta { __antithesis_instrumentation__.Notify(698823); return tpchMeta }

func (w *tpch) Flags() workload.Flags { __antithesis_instrumentation__.Notify(698824); return w.flags }

func (w *tpch) Hooks() workload.Hooks {
	__antithesis_instrumentation__.Notify(698825)
	return workload.Hooks{
		Validate: func() error {
			__antithesis_instrumentation__.Notify(698826)
			if w.scaleFactor != 1 {
				__antithesis_instrumentation__.Notify(698829)
				fmt.Printf("check for expected rows is only supported with " +
					"scale factor 1, so it was disabled\n")
				w.enableChecks = false
			} else {
				__antithesis_instrumentation__.Notify(698830)
			}
			__antithesis_instrumentation__.Notify(698827)
			for _, queryName := range strings.Split(w.queriesRaw, `,`) {
				__antithesis_instrumentation__.Notify(698831)
				queryNum, err := strconv.Atoi(queryName)
				if err != nil {
					__antithesis_instrumentation__.Notify(698834)
					return err
				} else {
					__antithesis_instrumentation__.Notify(698835)
				}
				__antithesis_instrumentation__.Notify(698832)
				if _, ok := QueriesByNumber[queryNum]; !ok {
					__antithesis_instrumentation__.Notify(698836)
					return errors.Errorf(`unknown query: %s`, queryName)
				} else {
					__antithesis_instrumentation__.Notify(698837)
				}
				__antithesis_instrumentation__.Notify(698833)
				w.selectedQueries = append(w.selectedQueries, queryNum)
			}
			__antithesis_instrumentation__.Notify(698828)
			return nil
		},
		PostLoad: func(db *gosql.DB) error {
			__antithesis_instrumentation__.Notify(698838)
			if w.fks {
				__antithesis_instrumentation__.Notify(698840)

				fkStmts := []string{
					`ALTER TABLE nation ADD CONSTRAINT nation_fkey_region FOREIGN KEY (n_regionkey) REFERENCES region (r_regionkey) NOT VALID`,
					`ALTER TABLE supplier ADD CONSTRAINT supplier_fkey_nation FOREIGN KEY (s_nationkey) REFERENCES nation (n_nationkey) NOT VALID`,
					`ALTER TABLE partsupp ADD CONSTRAINT partsupp_fkey_part FOREIGN KEY (ps_partkey) REFERENCES part (p_partkey) NOT VALID`,
					`ALTER TABLE partsupp ADD CONSTRAINT partsupp_fkey_supplier FOREIGN KEY (ps_suppkey) REFERENCES supplier (s_suppkey) NOT VALID`,
					`ALTER TABLE customer ADD CONSTRAINT customer_fkey_nation FOREIGN KEY (c_nationkey) REFERENCES nation (n_nationkey) NOT VALID`,
					`ALTER TABLE orders ADD CONSTRAINT orders_fkey_customer FOREIGN KEY (o_custkey) REFERENCES customer (c_custkey) NOT VALID`,
					`ALTER TABLE lineitem ADD CONSTRAINT lineitem_fkey_orders FOREIGN KEY (l_orderkey) REFERENCES orders (o_orderkey) NOT VALID`,
					`ALTER TABLE lineitem ADD CONSTRAINT lineitem_fkey_part FOREIGN KEY (l_partkey) REFERENCES part (p_partkey) NOT VALID`,
					`ALTER TABLE lineitem ADD CONSTRAINT lineitem_fkey_supplier FOREIGN KEY (l_suppkey) REFERENCES supplier (s_suppkey) NOT VALID`,
				}

				for _, fkStmt := range fkStmts {
					__antithesis_instrumentation__.Notify(698841)
					if _, err := db.Exec(fkStmt); err != nil {
						__antithesis_instrumentation__.Notify(698842)

						const duplFKErr = "columns cannot be used by multiple foreign key constraints"
						if !strings.Contains(err.Error(), duplFKErr) {
							__antithesis_instrumentation__.Notify(698843)
							return errors.Wrapf(err, "while executing %s", fkStmt)
						} else {
							__antithesis_instrumentation__.Notify(698844)
						}
					} else {
						__antithesis_instrumentation__.Notify(698845)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(698846)
			}
			__antithesis_instrumentation__.Notify(698839)
			return nil
		},
	}
}

type generateLocals struct {
	rng *rand.Rand

	namePerm []int

	orderData *orderSharedRandomData
}

func (w *tpch) Tables() []workload.Table {
	__antithesis_instrumentation__.Notify(698847)
	if w.localsPool == nil {
		__antithesis_instrumentation__.Notify(698849)
		w.localsPool = &sync.Pool{
			New: func() interface{} {
				__antithesis_instrumentation__.Notify(698850)
				namePerm := make([]int, len(randPartNames))
				for i := range namePerm {
					__antithesis_instrumentation__.Notify(698852)
					namePerm[i] = i
				}
				__antithesis_instrumentation__.Notify(698851)
				return &generateLocals{
					rng:      rand.New(rand.NewSource(uint64(timeutil.Now().UnixNano()))),
					namePerm: namePerm,
					orderData: &orderSharedRandomData{
						partKeys:   make([]int, 0, 7),
						shipDates:  make([]int64, 0, 7),
						quantities: make([]float32, 0, 7),
						discount:   make([]float32, 0, 7),
						tax:        make([]float32, 0, 7),
					},
				}
			},
		}
	} else {
		__antithesis_instrumentation__.Notify(698853)
	}
	__antithesis_instrumentation__.Notify(698848)

	w.textPool = &fakeTextPool{seed: w.seed}

	nation := workload.Table{
		Name:   `nation`,
		Schema: tpchNationSchema,
		InitialRows: workload.BatchedTuples{
			NumBatches: numNation,
			FillBatch:  w.tpchNationInitialRowBatch,
		},
	}
	region := workload.Table{
		Name:   `region`,
		Schema: tpchRegionSchema,
		InitialRows: workload.BatchedTuples{
			NumBatches: numRegion,
			FillBatch:  w.tpchRegionInitialRowBatch,
		},
	}
	numSupplier := numSupplierPerSF * w.scaleFactor
	supplier := workload.Table{
		Name:   `supplier`,
		Schema: tpchSupplierSchema,
		InitialRows: workload.BatchedTuples{
			NumBatches: numSupplier,
			FillBatch:  w.tpchSupplierInitialRowBatch,
		},
	}
	numPart := numPartPerSF * w.scaleFactor
	part := workload.Table{
		Name:   `part`,
		Schema: tpchPartSchema,
		InitialRows: workload.BatchedTuples{
			NumBatches: numPart,
			FillBatch:  w.tpchPartInitialRowBatch,
		},
	}
	partsupp := workload.Table{
		Name:   `partsupp`,
		Schema: tpchPartSuppSchema,
		InitialRows: workload.BatchedTuples{

			NumBatches: numPart,
			FillBatch:  w.tpchPartSuppInitialRowBatch,
		},
	}
	numCustomer := numCustomerPerSF * w.scaleFactor
	customer := workload.Table{
		Name:   `customer`,
		Schema: tpchCustomerSchema,
		InitialRows: workload.BatchedTuples{
			NumBatches: numCustomer,
			FillBatch:  w.tpchCustomerInitialRowBatch,
		},
	}
	orders := workload.Table{
		Name:   `orders`,
		Schema: tpchOrdersSchema,
		InitialRows: workload.BatchedTuples{

			NumBatches: numCustomer,
			FillBatch:  w.tpchOrdersInitialRowBatch,
		},
	}
	lineitem := workload.Table{
		Name:   `lineitem`,
		Schema: tpchLineItemSchema,
		InitialRows: workload.BatchedTuples{

			NumBatches: numCustomer,
			FillBatch:  w.tpchLineItemInitialRowBatch,
		},
	}

	return []workload.Table{
		nation, region, part, supplier, partsupp, customer, orders, lineitem,
	}
}

func (w *tpch) Ops(
	ctx context.Context, urls []string, reg *histogram.Registry,
) (workload.QueryLoad, error) {
	__antithesis_instrumentation__.Notify(698854)
	sqlDatabase, err := workload.SanitizeUrls(w, w.connFlags.DBOverride, urls)
	if err != nil {
		__antithesis_instrumentation__.Notify(698858)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(698859)
	}
	__antithesis_instrumentation__.Notify(698855)
	db, err := gosql.Open(`cockroach`, strings.Join(urls, ` `))
	if err != nil {
		__antithesis_instrumentation__.Notify(698860)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(698861)
	}
	__antithesis_instrumentation__.Notify(698856)

	db.SetMaxOpenConns(w.connFlags.Concurrency + 1)
	db.SetMaxIdleConns(w.connFlags.Concurrency + 1)

	ql := workload.QueryLoad{SQLDatabase: sqlDatabase}
	for i := 0; i < w.connFlags.Concurrency; i++ {
		__antithesis_instrumentation__.Notify(698862)
		worker := &worker{
			config: w,
			hists:  reg.GetHandle(),
			db:     db,
		}
		ql.WorkerFns = append(ql.WorkerFns, worker.run)
	}
	__antithesis_instrumentation__.Notify(698857)
	return ql, nil
}

type worker struct {
	config *tpch
	hists  *histogram.Histograms
	db     *gosql.DB
	ops    int
}

func (w *worker) run(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(698863)
	queryNum := w.config.selectedQueries[w.ops%len(w.config.selectedQueries)]
	w.ops++

	var prefix string
	if !w.config.useClusterVectorizeSetting {
		__antithesis_instrumentation__.Notify(698874)
		prefix = fmt.Sprintf("SET vectorize = '%s';", w.config.vectorize)
	} else {
		__antithesis_instrumentation__.Notify(698875)
	}
	__antithesis_instrumentation__.Notify(698864)
	query := fmt.Sprintf("%s %s", prefix, QueriesByNumber[queryNum])

	vals := make([]interface{}, maxCols)
	for i := range vals {
		__antithesis_instrumentation__.Notify(698876)
		vals[i] = new(interface{})
	}
	__antithesis_instrumentation__.Notify(698865)

	start := timeutil.Now()
	rows, err := w.db.Query(query)
	if rows != nil {
		__antithesis_instrumentation__.Notify(698877)
		defer rows.Close()
	} else {
		__antithesis_instrumentation__.Notify(698878)
	}
	__antithesis_instrumentation__.Notify(698866)
	if err != nil {
		__antithesis_instrumentation__.Notify(698879)
		return errors.Wrapf(err, "[q%d]", queryNum)
	} else {
		__antithesis_instrumentation__.Notify(698880)
	}
	__antithesis_instrumentation__.Notify(698867)
	var numRows int

	checkExpectedOutput := func() error {
		__antithesis_instrumentation__.Notify(698881)
		for rows.Next() {
			__antithesis_instrumentation__.Notify(698883)
			if w.config.enableChecks {
				__antithesis_instrumentation__.Notify(698885)
				if _, checkOnlyRowCount := numExpectedRowsByQueryNumber[queryNum]; !checkOnlyRowCount {
					__antithesis_instrumentation__.Notify(698886)
					if err = rows.Scan(vals[:numColsByQueryNumber[queryNum]]...); err != nil {
						__antithesis_instrumentation__.Notify(698888)
						return errors.Wrapf(err, "[q%d]", queryNum)
					} else {
						__antithesis_instrumentation__.Notify(698889)
					}
					__antithesis_instrumentation__.Notify(698887)

					expectedRow := expectedRowsByQueryNumber[queryNum][numRows]
					for i, expectedValue := range expectedRow {
						__antithesis_instrumentation__.Notify(698890)
						if val := *vals[i].(*interface{}); val != nil {
							__antithesis_instrumentation__.Notify(698891)
							var actualValue string

							if byteArray, ok := val.([]byte); ok {
								__antithesis_instrumentation__.Notify(698893)
								actualValue = string(byteArray)
							} else {
								__antithesis_instrumentation__.Notify(698894)
								actualValue = fmt.Sprint(val)
							}
							__antithesis_instrumentation__.Notify(698892)
							if strings.Compare(expectedValue, actualValue) != 0 {
								__antithesis_instrumentation__.Notify(698895)
								var expectedFloat, actualFloat float64
								var expectedFloatRounded, actualFloatRounded float64
								expectedFloat, err = strconv.ParseFloat(expectedValue, 64)
								if err != nil {
									__antithesis_instrumentation__.Notify(698900)
									return errors.Errorf("[q%d] failed parsing expected value as float64 with %s\n"+
										"wrong result in row %d in column %d: got %q, expected %q",
										queryNum, err, numRows, i, actualValue, expectedValue)
								} else {
									__antithesis_instrumentation__.Notify(698901)
								}
								__antithesis_instrumentation__.Notify(698896)
								actualFloat, err = strconv.ParseFloat(actualValue, 64)
								if err != nil {
									__antithesis_instrumentation__.Notify(698902)
									return errors.Errorf("[q%d] failed parsing actual value as float64 with %s\n"+
										"wrong result in row %d in column %d: got %q, expected %q",
										queryNum, err, numRows, i, actualValue, expectedValue)
								} else {
									__antithesis_instrumentation__.Notify(698903)
								}
								__antithesis_instrumentation__.Notify(698897)

								expectedFloatRounded, err = strconv.ParseFloat(fmt.Sprintf("%.3f", expectedFloat), 64)
								if err != nil {
									__antithesis_instrumentation__.Notify(698904)
									return errors.Errorf("[q%d] failed parsing rounded expected value as float64 with %s\n"+
										"wrong result in row %d in column %d: got %q, expected %q",
										queryNum, err, numRows, i, actualValue, expectedValue)
								} else {
									__antithesis_instrumentation__.Notify(698905)
								}
								__antithesis_instrumentation__.Notify(698898)
								actualFloatRounded, err = strconv.ParseFloat(fmt.Sprintf("%.3f", actualFloat), 64)
								if err != nil {
									__antithesis_instrumentation__.Notify(698906)
									return errors.Errorf("[q%d] failed parsing rounded actual value as float64 with %s\n"+
										"wrong result in row %d in column %d: got %q, expected %q",
										queryNum, err, numRows, i, actualValue, expectedValue)
								} else {
									__antithesis_instrumentation__.Notify(698907)
								}
								__antithesis_instrumentation__.Notify(698899)
								if math.Abs(expectedFloatRounded-actualFloatRounded) > 0.02 {
									__antithesis_instrumentation__.Notify(698908)

									return errors.Errorf("[q%d] %f and %f differ by more than 0.02\n"+
										"wrong result in row %d in column %d: got %q, expected %q",
										queryNum, actualFloatRounded, expectedFloatRounded,
										numRows, i, actualValue, expectedValue)
								} else {
									__antithesis_instrumentation__.Notify(698909)
								}
							} else {
								__antithesis_instrumentation__.Notify(698910)
							}
						} else {
							__antithesis_instrumentation__.Notify(698911)
						}
					}
				} else {
					__antithesis_instrumentation__.Notify(698912)
				}
			} else {
				__antithesis_instrumentation__.Notify(698913)
			}
			__antithesis_instrumentation__.Notify(698884)
			numRows++
		}
		__antithesis_instrumentation__.Notify(698882)
		return nil
	}
	__antithesis_instrumentation__.Notify(698868)

	expectedOutputError := checkExpectedOutput()

	for rows.Next() {
		__antithesis_instrumentation__.Notify(698914)
	}
	__antithesis_instrumentation__.Notify(698869)

	if err := rows.Err(); err != nil {
		__antithesis_instrumentation__.Notify(698915)
		return errors.Wrapf(err, "[q%d]", queryNum)
	} else {
		__antithesis_instrumentation__.Notify(698916)
	}
	__antithesis_instrumentation__.Notify(698870)

	if expectedOutputError != nil {
		__antithesis_instrumentation__.Notify(698917)
		return wrongOutputError{error: expectedOutputError}
	} else {
		__antithesis_instrumentation__.Notify(698918)
	}
	__antithesis_instrumentation__.Notify(698871)
	if w.config.enableChecks {
		__antithesis_instrumentation__.Notify(698919)
		numRowsExpected, checkOnlyRowCount := numExpectedRowsByQueryNumber[queryNum]
		if checkOnlyRowCount && func() bool {
			__antithesis_instrumentation__.Notify(698920)
			return numRows != numRowsExpected == true
		}() == true {
			__antithesis_instrumentation__.Notify(698921)
			return wrongOutputError{
				error: errors.Errorf(
					"[q%d] returned wrong number of rows: got %d, expected %d",
					queryNum, numRows, numRowsExpected,
				)}
		} else {
			__antithesis_instrumentation__.Notify(698922)
		}
	} else {
		__antithesis_instrumentation__.Notify(698923)
	}
	__antithesis_instrumentation__.Notify(698872)
	elapsed := timeutil.Since(start)
	if w.config.verbose {
		__antithesis_instrumentation__.Notify(698924)
		w.hists.Get(fmt.Sprintf("%d", queryNum)).Record(elapsed)

		log.Infof(ctx, "[q%d] returned %d rows after %4.2f seconds:\n%s",
			queryNum, numRows, elapsed.Seconds(), query)
	} else {
		__antithesis_instrumentation__.Notify(698925)

		log.Infof(ctx, "[q%d] returned %d rows after %4.2f seconds",
			queryNum, numRows, elapsed.Seconds())
	}
	__antithesis_instrumentation__.Notify(698873)
	return nil
}

const (
	tpchNationSchema = `(
		n_nationkey       INTEGER NOT NULL PRIMARY KEY,
		n_name            CHAR(25) NOT NULL,
		n_regionkey       INTEGER NOT NULL,
		n_comment         VARCHAR(152),
		INDEX n_rk (n_regionkey ASC)
	)`
	tpchRegionSchema = `(
		r_regionkey       INTEGER NOT NULL PRIMARY KEY,
		r_name            CHAR(25) NOT NULL,
		r_comment         VARCHAR(152)
	)`
	tpchPartSchema = `(
		p_partkey         INTEGER NOT NULL PRIMARY KEY,
		p_name            VARCHAR(55) NOT NULL,
		p_mfgr            CHAR(25) NOT NULL,
		p_brand           CHAR(10) NOT NULL,
		p_type            VARCHAR(25) NOT NULL,
		p_size            INTEGER NOT NULL,
		p_container       CHAR(10) NOT NULL,
		p_retailprice     FLOAT NOT NULL,
		p_comment         VARCHAR(23) NOT NULL
	)`
	tpchSupplierSchema = `(
		s_suppkey         INTEGER NOT NULL PRIMARY KEY,
		s_name            CHAR(25) NOT NULL,
		s_address         VARCHAR(40) NOT NULL,
		s_nationkey       INTEGER NOT NULL,
		s_phone           CHAR(15) NOT NULL,
		s_acctbal         FLOAT NOT NULL,
		s_comment         VARCHAR(101) NOT NULL,
		INDEX s_nk (s_nationkey ASC)
	)`
	tpchPartSuppSchema = `(
		ps_partkey            INTEGER NOT NULL,
		ps_suppkey            INTEGER NOT NULL,
		ps_availqty           INTEGER NOT NULL,
		ps_supplycost         FLOAT NOT NULL,
		ps_comment            VARCHAR(199) NOT NULL,
		PRIMARY KEY (ps_partkey ASC, ps_suppkey ASC),
		INDEX ps_sk (ps_suppkey ASC)
	)`
	tpchCustomerSchema = `(
		c_custkey         INTEGER NOT NULL PRIMARY KEY,
		c_name            VARCHAR(25) NOT NULL,
		c_address         VARCHAR(40) NOT NULL,
		c_nationkey       INTEGER NOT NULL,
		c_phone           CHAR(15) NOT NULL,
		c_acctbal         FLOAT NOT NULL,
		c_mktsegment      CHAR(10) NOT NULL,
		c_comment         VARCHAR(117) NOT NULL,
		INDEX c_nk (c_nationkey ASC)
	)`
	tpchOrdersSchema = `(
		o_orderkey           INTEGER NOT NULL PRIMARY KEY,
		o_custkey            INTEGER NOT NULL,
		o_orderstatus        CHAR(1) NOT NULL,
		o_totalprice         FLOAT NOT NULL,
		o_orderdate          DATE NOT NULL,
		o_orderpriority      CHAR(15) NOT NULL,
		o_clerk              CHAR(15) NOT NULL,
		o_shippriority       INTEGER NOT NULL,
		o_comment            VARCHAR(79) NOT NULL,
		INDEX o_ck (o_custkey ASC),
		INDEX o_od (o_orderdate ASC)
	)`
	tpchLineItemSchema = `(
		l_orderkey      INTEGER NOT NULL,
		l_partkey       INTEGER NOT NULL,
		l_suppkey       INTEGER NOT NULL,
		l_linenumber    INTEGER NOT NULL,
		l_quantity      FLOAT NOT NULL,
		l_extendedprice FLOAT NOT NULL,
		l_discount      FLOAT NOT NULL,
		l_tax           FLOAT NOT NULL,
		l_returnflag    CHAR(1) NOT NULL,
		l_linestatus    CHAR(1) NOT NULL,
		l_shipdate      DATE NOT NULL,
		l_commitdate    DATE NOT NULL,
		l_receiptdate   DATE NOT NULL,
		l_shipinstruct  CHAR(25) NOT NULL,
		l_shipmode      CHAR(10) NOT NULL,
		l_comment       VARCHAR(44) NOT NULL,
		PRIMARY KEY (l_orderkey, l_linenumber),
		INDEX l_ok (l_orderkey ASC),
		INDEX l_pk (l_partkey ASC),
		INDEX l_sk (l_suppkey ASC),
		INDEX l_sd (l_shipdate ASC),
		INDEX l_cd (l_commitdate ASC),
		INDEX l_rd (l_receiptdate ASC),
		INDEX l_pk_sk (l_partkey ASC, l_suppkey ASC),
		INDEX l_sk_pk (l_suppkey ASC, l_partkey ASC)
	)`
)
