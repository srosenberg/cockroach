package tpcc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"github.com/cockroachdb/cockroach/pkg/workload/histogram"
	"github.com/cockroachdb/cockroach/pkg/workload/workloadimpl"
	"github.com/cockroachdb/errors"
	"github.com/jackc/pgx/v4"
	"github.com/spf13/pflag"
	"golang.org/x/exp/rand"
	"golang.org/x/sync/errgroup"
)

type tpcc struct {
	flags     workload.Flags
	connFlags *workload.ConnFlags

	seed             uint64
	warehouses       int
	activeWarehouses int
	nowString        []byte
	numConns         int

	idleConns int

	cLoad, cCustomerID, cItemID int

	mix                    string
	waitFraction           float64
	workers                int
	fks                    bool
	separateColumnFamilies bool

	deprecatedFkIndexes bool
	dbOverride          string

	txInfos []txInfo

	deck []int

	auditor *auditor

	reg        *histogram.Registry
	txCounters txCounters

	split   bool
	scatter bool

	partitions         int
	clientPartitions   int
	affinityPartitions []int
	wPart              *partitioner
	wMRPart            *partitioner
	zoneCfg            zoneConfig
	multiRegionCfg     multiRegionConfig

	localWarehouses bool

	usePostgres  bool
	serializable bool
	txOpts       pgx.TxOptions

	expensiveChecks bool

	replicateStaticColumns bool

	randomCIDsCache struct {
		syncutil.Mutex
		values [][]int
	}
	localsPool *sync.Pool
}

type waitSetter struct {
	val *float64
}

func (w *waitSetter) Set(val string) error {
	__antithesis_instrumentation__.Notify(698318)
	switch strings.ToLower(val) {
	case "true", "on":
		__antithesis_instrumentation__.Notify(698320)
		*w.val = 1.0
	case "false", "off":
		__antithesis_instrumentation__.Notify(698321)
		*w.val = 0.0
	default:
		__antithesis_instrumentation__.Notify(698322)
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			__antithesis_instrumentation__.Notify(698325)
			return err
		} else {
			__antithesis_instrumentation__.Notify(698326)
		}
		__antithesis_instrumentation__.Notify(698323)
		if f < 0 {
			__antithesis_instrumentation__.Notify(698327)
			return errors.New("cannot set --wait to a negative value")
		} else {
			__antithesis_instrumentation__.Notify(698328)
		}
		__antithesis_instrumentation__.Notify(698324)
		*w.val = f
	}
	__antithesis_instrumentation__.Notify(698319)
	return nil
}

func (*waitSetter) Type() string {
	__antithesis_instrumentation__.Notify(698329)
	return "0.0/false - 1.0/true"
}

func (w *waitSetter) String() string {
	__antithesis_instrumentation__.Notify(698330)
	switch *w.val {
	case 0:
		__antithesis_instrumentation__.Notify(698331)
		return "false"
	case 1:
		__antithesis_instrumentation__.Notify(698332)
		return "true"
	default:
		__antithesis_instrumentation__.Notify(698333)
		return fmt.Sprintf("%f", *w.val)
	}
}

func init() {
	workload.Register(tpccMeta)
}

func FromWarehouses(warehouses int) workload.Generator {
	__antithesis_instrumentation__.Notify(698334)
	return workload.FromFlags(tpccMeta, fmt.Sprintf(`--warehouses=%d`, warehouses))
}

var tpccMeta = workload.Meta{
	Name: `tpcc`,
	Description: `TPC-C simulates a transaction processing workload` +
		` using a rich schema of multiple tables`,
	Version:      `2.2.0`,
	PublicFacing: true,
	New: func() workload.Generator {
		__antithesis_instrumentation__.Notify(698335)
		g := &tpcc{}
		g.flags.FlagSet = pflag.NewFlagSet(`tpcc`, pflag.ContinueOnError)
		g.flags.Meta = map[string]workload.FlagMeta{
			`db`:                       {RuntimeOnly: true},
			`mix`:                      {RuntimeOnly: true},
			`partitions`:               {RuntimeOnly: true},
			`client-partitions`:        {RuntimeOnly: true},
			`partition-affinity`:       {RuntimeOnly: true},
			`partition-strategy`:       {RuntimeOnly: true},
			`zones`:                    {RuntimeOnly: true},
			`active-warehouses`:        {RuntimeOnly: true},
			`scatter`:                  {RuntimeOnly: true},
			`serializable`:             {RuntimeOnly: true},
			`split`:                    {RuntimeOnly: true},
			`wait`:                     {RuntimeOnly: true},
			`workers`:                  {RuntimeOnly: true},
			`conns`:                    {RuntimeOnly: true},
			`idle-conns`:               {RuntimeOnly: true},
			`expensive-checks`:         {RuntimeOnly: true, CheckConsistencyOnly: true},
			`local-warehouses`:         {RuntimeOnly: true},
			`regions`:                  {RuntimeOnly: true},
			`survival-goal`:            {RuntimeOnly: true},
			`replicate-static-columns`: {RuntimeOnly: true},
			`deprecated-fk-indexes`:    {RuntimeOnly: true},
		}

		g.flags.Uint64Var(&g.seed, `seed`, 1, `Random number generator seed`)
		g.flags.IntVar(&g.warehouses, `warehouses`, 1, `Number of warehouses for loading`)
		g.flags.BoolVar(&g.fks, `fks`, true, `Add the foreign keys`)
		g.flags.BoolVar(&g.deprecatedFkIndexes, `deprecated-fk-indexes`, false, `Add deprecated foreign keys (needed when running against v20.1 or below clusters)`)

		g.flags.StringVar(&g.mix, `mix`,
			`newOrder=10,payment=10,orderStatus=1,delivery=1,stockLevel=1`,
			`Weights for the transaction mix. The default matches the TPCC spec.`)
		g.waitFraction = 1.0
		g.flags.Var(&waitSetter{&g.waitFraction}, `wait`, `Wait mode (include think/keying sleeps): 1/true for tpcc-standard wait, 0/false for no waits, other factors also allowed`)
		g.flags.StringVar(&g.dbOverride, `db`, ``,
			`Override for the SQL database to use. If empty, defaults to the generator name`)
		g.flags.IntVar(&g.workers, `workers`, 0, fmt.Sprintf(
			`Number of concurrent workers. Defaults to --warehouses * %d`, NumWorkersPerWarehouse,
		))
		g.flags.IntVar(&g.numConns, `conns`, 0, fmt.Sprintf(
			`Number of connections. Defaults to --warehouses * %d (except in nowait mode, where it defaults to --workers`,
			numConnsPerWarehouse,
		))
		g.flags.IntVar(&g.idleConns, `idle-conns`, 0, `Number of idle connections. Defaults to 0`)
		g.flags.IntVar(&g.partitions, `partitions`, 1, `Partition tables`)
		g.flags.IntVar(&g.clientPartitions, `client-partitions`, 0, `Make client behave as if the tables are partitioned, but does not actually partition underlying data. Requires --partition-affinity.`)
		g.flags.IntSliceVar(&g.affinityPartitions, `partition-affinity`, nil, `Run load generator against specific partition (requires partitions). `+
			`Note that if one value is provided, the assumption is that all urls are associated with that partition. In all other cases the assumption `+
			`is that the URLs are distributed evenly over the partitions`)
		g.flags.Var(&g.zoneCfg.strategy, `partition-strategy`, `Partition tables according to which strategy [replication, leases]`)
		g.flags.StringSliceVar(&g.zoneCfg.zones, "zones", []string{}, "Zones for legacy partitioning, the number of zones should match the number of partitions and the zones used to start cockroach. Does not work with --regions.")
		g.flags.StringSliceVar(&g.multiRegionCfg.regions, "regions", []string{}, "Regions to use for multi-region partitioning. The first region is the PRIMARY REGION. Does not work with --zones.")
		g.flags.Var(&g.multiRegionCfg.survivalGoal, "survival-goal", "Survival goal to use for multi-region setups. Allowed values: [zone, region].")
		g.flags.IntVar(&g.activeWarehouses, `active-warehouses`, 0, `Run the load generator against a specific number of warehouses. Defaults to --warehouses'`)
		g.flags.BoolVar(&g.scatter, `scatter`, false, `Scatter ranges`)
		g.flags.BoolVar(&g.serializable, `serializable`, false, `Force serializable mode`)
		g.flags.BoolVar(&g.split, `split`, false, `Split tables`)
		g.flags.BoolVar(&g.expensiveChecks, `expensive-checks`, false, `Run expensive checks`)
		g.flags.BoolVar(&g.separateColumnFamilies, `families`, false, `Use separate column families for dynamic and static columns`)
		g.flags.BoolVar(&g.replicateStaticColumns, `replicate-static-columns`, false, "Create duplicate indexes for all static columns in district, items and warehouse tables, such that each zone or rack has them locally.")
		g.flags.BoolVar(&g.localWarehouses, `local-warehouses`, false, `Force transactions to use a local warehouse in all cases (in violation of the TPC-C specification)`)
		g.connFlags = workload.NewConnFlags(&g.flags)

		g.nowString = []byte(`2006-01-02 15:04:05`)
		return g
	},
}

func queryDatabaseRegions(db *gosql.DB) (map[string]struct{}, error) {
	__antithesis_instrumentation__.Notify(698336)
	regions := make(map[string]struct{})
	rows, err := db.Query(`SELECT region FROM [SHOW REGIONS FROM DATABASE]`)
	if err != nil {
		__antithesis_instrumentation__.Notify(698341)
		return regions, err
	} else {
		__antithesis_instrumentation__.Notify(698342)
	}
	__antithesis_instrumentation__.Notify(698337)
	defer func() {
		__antithesis_instrumentation__.Notify(698343)
		_ = rows.Close()
	}()
	__antithesis_instrumentation__.Notify(698338)
	for rows.Next() {
		__antithesis_instrumentation__.Notify(698344)
		if rows.Err() != nil {
			__antithesis_instrumentation__.Notify(698347)
			return regions, err
		} else {
			__antithesis_instrumentation__.Notify(698348)
		}
		__antithesis_instrumentation__.Notify(698345)
		var region string
		if err := rows.Scan(&region); err != nil {
			__antithesis_instrumentation__.Notify(698349)
			return regions, err
		} else {
			__antithesis_instrumentation__.Notify(698350)
		}
		__antithesis_instrumentation__.Notify(698346)
		regions[region] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(698339)
	if rows.Err() != nil {
		__antithesis_instrumentation__.Notify(698351)
		return regions, err
	} else {
		__antithesis_instrumentation__.Notify(698352)
	}
	__antithesis_instrumentation__.Notify(698340)
	return regions, nil
}

func (*tpcc) Meta() workload.Meta { __antithesis_instrumentation__.Notify(698353); return tpccMeta }

func (w *tpcc) Flags() workload.Flags { __antithesis_instrumentation__.Notify(698354); return w.flags }

func (w *tpcc) Hooks() workload.Hooks {
	__antithesis_instrumentation__.Notify(698355)
	return workload.Hooks{
		Validate: func() error {
			__antithesis_instrumentation__.Notify(698356)
			if w.warehouses < 1 {
				__antithesis_instrumentation__.Notify(698369)
				return errors.Errorf(`--warehouses must be positive`)
			} else {
				__antithesis_instrumentation__.Notify(698370)
			}
			__antithesis_instrumentation__.Notify(698357)

			if w.activeWarehouses > w.warehouses {
				__antithesis_instrumentation__.Notify(698371)
				return errors.Errorf(`--active-warehouses needs to be less than or equal to warehouses`)
			} else {
				__antithesis_instrumentation__.Notify(698372)
				if w.activeWarehouses == 0 {
					__antithesis_instrumentation__.Notify(698373)
					w.activeWarehouses = w.warehouses
				} else {
					__antithesis_instrumentation__.Notify(698374)
				}
			}
			__antithesis_instrumentation__.Notify(698358)

			if w.partitions < 1 {
				__antithesis_instrumentation__.Notify(698375)
				return errors.Errorf(`--partitions must be positive`)
			} else {
				__antithesis_instrumentation__.Notify(698376)
			}
			__antithesis_instrumentation__.Notify(698359)

			if len(w.zoneCfg.zones) > 0 && func() bool {
				__antithesis_instrumentation__.Notify(698377)
				return len(w.multiRegionCfg.regions) > 0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(698378)
				return errors.Errorf("cannot specify both --regions and --zones")
			} else {
				__antithesis_instrumentation__.Notify(698379)
			}
			__antithesis_instrumentation__.Notify(698360)

			if w.clientPartitions > 0 {
				__antithesis_instrumentation__.Notify(698380)
				if w.partitions > 1 {
					__antithesis_instrumentation__.Notify(698383)
					return errors.Errorf(`cannot specify both --partitions and --client-partitions;
					--partitions actually partitions underlying data.
					--client-partitions only modifies client behavior to access a subset of warehouses. Must be used with --partition-affinity`)
				} else {
					__antithesis_instrumentation__.Notify(698384)
				}
				__antithesis_instrumentation__.Notify(698381)

				if len(w.affinityPartitions) == 0 {
					__antithesis_instrumentation__.Notify(698385)
					return errors.Errorf(`--client-partitions must be used with --partition-affinity.`)
				} else {
					__antithesis_instrumentation__.Notify(698386)
				}
				__antithesis_instrumentation__.Notify(698382)

				for _, p := range w.affinityPartitions {
					__antithesis_instrumentation__.Notify(698387)
					if p >= w.clientPartitions {
						__antithesis_instrumentation__.Notify(698388)
						return errors.Errorf(`--partition-affinity %d in %v out of bounds of --client-partitions`,
							p, w.affinityPartitions)
					} else {
						__antithesis_instrumentation__.Notify(698389)
					}
				}

			} else {
				__antithesis_instrumentation__.Notify(698390)
				for _, p := range w.affinityPartitions {
					__antithesis_instrumentation__.Notify(698394)
					if p < 0 || func() bool {
						__antithesis_instrumentation__.Notify(698395)
						return p >= w.partitions == true
					}() == true {
						__antithesis_instrumentation__.Notify(698396)
						return errors.Errorf(`--partition-affinity out of bounds of --partitions`)
					} else {
						__antithesis_instrumentation__.Notify(698397)
					}
				}
				__antithesis_instrumentation__.Notify(698391)

				if len(w.multiRegionCfg.regions) > 0 && func() bool {
					__antithesis_instrumentation__.Notify(698398)
					return (len(w.multiRegionCfg.regions) != w.partitions) == true
				}() == true {
					__antithesis_instrumentation__.Notify(698399)
					return errors.Errorf(`--regions should have the same length as --partitions.`)
				} else {
					__antithesis_instrumentation__.Notify(698400)
				}
				__antithesis_instrumentation__.Notify(698392)

				if len(w.multiRegionCfg.regions) < 3 && func() bool {
					__antithesis_instrumentation__.Notify(698401)
					return w.multiRegionCfg.survivalGoal == survivalGoalRegion == true
				}() == true {
					__antithesis_instrumentation__.Notify(698402)
					return errors.Errorf(`REGION survivability needs at least 3 regions.`)
				} else {
					__antithesis_instrumentation__.Notify(698403)
				}
				__antithesis_instrumentation__.Notify(698393)

				if len(w.zoneCfg.zones) > 0 && func() bool {
					__antithesis_instrumentation__.Notify(698404)
					return (len(w.zoneCfg.zones) != w.partitions) == true
				}() == true {
					__antithesis_instrumentation__.Notify(698405)
					return errors.Errorf(`--zones should have the same length as --partitions.`)
				} else {
					__antithesis_instrumentation__.Notify(698406)
				}
			}
			__antithesis_instrumentation__.Notify(698361)

			w.initNonUniformRandomConstants()

			if w.workers == 0 {
				__antithesis_instrumentation__.Notify(698407)
				w.workers = w.activeWarehouses * NumWorkersPerWarehouse
			} else {
				__antithesis_instrumentation__.Notify(698408)
			}
			__antithesis_instrumentation__.Notify(698362)

			if w.numConns == 0 {
				__antithesis_instrumentation__.Notify(698409)

				if w.waitFraction == 0 {
					__antithesis_instrumentation__.Notify(698410)
					w.numConns = w.workers
				} else {
					__antithesis_instrumentation__.Notify(698411)
					w.numConns = w.activeWarehouses * numConnsPerWarehouse
				}
			} else {
				__antithesis_instrumentation__.Notify(698412)
			}
			__antithesis_instrumentation__.Notify(698363)

			if w.waitFraction > 0 && func() bool {
				__antithesis_instrumentation__.Notify(698413)
				return w.workers != w.activeWarehouses*NumWorkersPerWarehouse == true
			}() == true {
				__antithesis_instrumentation__.Notify(698414)
				return errors.Errorf(`--wait > 0 and --warehouses=%d requires --workers=%d`,
					w.activeWarehouses, w.warehouses*NumWorkersPerWarehouse)
			} else {
				__antithesis_instrumentation__.Notify(698415)
			}
			__antithesis_instrumentation__.Notify(698364)

			if w.serializable {
				__antithesis_instrumentation__.Notify(698416)
				w.txOpts = pgx.TxOptions{IsoLevel: pgx.Serializable}
			} else {
				__antithesis_instrumentation__.Notify(698417)
			}
			__antithesis_instrumentation__.Notify(698365)

			w.auditor = newAuditor(w.activeWarehouses)

			partitions := w.partitions
			if w.clientPartitions > 0 {
				__antithesis_instrumentation__.Notify(698418)
				partitions = w.clientPartitions
			} else {
				__antithesis_instrumentation__.Notify(698419)
			}
			__antithesis_instrumentation__.Notify(698366)
			var err error

			w.wPart, err = makePartitioner(w.warehouses, w.activeWarehouses, partitions)
			if err != nil {
				__antithesis_instrumentation__.Notify(698420)
				return errors.Wrap(err, "error creating partitioner")
			} else {
				__antithesis_instrumentation__.Notify(698421)
			}
			__antithesis_instrumentation__.Notify(698367)
			if len(w.multiRegionCfg.regions) != 0 {
				__antithesis_instrumentation__.Notify(698422)

				w.wMRPart, err = makeMRPartitioner(w.warehouses, w.activeWarehouses, partitions)
				if err != nil {
					__antithesis_instrumentation__.Notify(698423)
					return errors.Wrap(err, "error creating multi-region partitioner")
				} else {
					__antithesis_instrumentation__.Notify(698424)
				}
			} else {
				__antithesis_instrumentation__.Notify(698425)
			}
			__antithesis_instrumentation__.Notify(698368)
			return initializeMix(w)
		},
		PreCreate: func(db *gosql.DB) error {
			__antithesis_instrumentation__.Notify(698426)
			if len(w.multiRegionCfg.regions) == 0 {
				__antithesis_instrumentation__.Notify(698433)

				return nil
			} else {
				__antithesis_instrumentation__.Notify(698434)
			}
			__antithesis_instrumentation__.Notify(698427)

			regions, err := queryDatabaseRegions(db)
			if err != nil {
				__antithesis_instrumentation__.Notify(698435)
				return err
			} else {
				__antithesis_instrumentation__.Notify(698436)
			}
			__antithesis_instrumentation__.Notify(698428)

			var dbName string
			if err := db.QueryRow(`SHOW DATABASE`).Scan(&dbName); err != nil {
				__antithesis_instrumentation__.Notify(698437)
				return err
			} else {
				__antithesis_instrumentation__.Notify(698438)
			}
			__antithesis_instrumentation__.Notify(698429)

			var stmts []string
			for i, region := range w.multiRegionCfg.regions {
				__antithesis_instrumentation__.Notify(698439)
				var stmt string

				if i == 0 {
					__antithesis_instrumentation__.Notify(698441)
					stmt = fmt.Sprintf(`alter database %s set primary region %q`, dbName, region)
				} else {
					__antithesis_instrumentation__.Notify(698442)

					if _, ok := regions[region]; ok {
						__antithesis_instrumentation__.Notify(698444)
						continue
					} else {
						__antithesis_instrumentation__.Notify(698445)
					}
					__antithesis_instrumentation__.Notify(698443)
					stmt = fmt.Sprintf(`alter database %s add region %q`, dbName, region)
				}
				__antithesis_instrumentation__.Notify(698440)
				stmts = append(stmts, stmt)
			}
			__antithesis_instrumentation__.Notify(698430)

			var survivalGoal string
			switch w.multiRegionCfg.survivalGoal {
			case survivalGoalZone:
				__antithesis_instrumentation__.Notify(698446)
				survivalGoal = `zone failure`
			case survivalGoalRegion:
				__antithesis_instrumentation__.Notify(698447)
				survivalGoal = `region failure`
			default:
				__antithesis_instrumentation__.Notify(698448)
				panic("unexpected")
			}
			__antithesis_instrumentation__.Notify(698431)
			stmts = append(stmts, fmt.Sprintf(`alter database %s survive %s`, dbName, survivalGoal))

			for _, stmt := range stmts {
				__antithesis_instrumentation__.Notify(698449)
				if _, err := db.Exec(stmt); err != nil {
					__antithesis_instrumentation__.Notify(698450)
					return err
				} else {
					__antithesis_instrumentation__.Notify(698451)
				}
			}
			__antithesis_instrumentation__.Notify(698432)

			return nil
		},
		PostLoad: func(db *gosql.DB) error {
			__antithesis_instrumentation__.Notify(698452)
			if w.fks {
				__antithesis_instrumentation__.Notify(698455)

				fkStmts := []string{
					`alter table district add foreign key (d_w_id) references warehouse (w_id) not valid`,
					`alter table customer add foreign key (c_w_id, c_d_id) references district (d_w_id, d_id) not valid`,
					`alter table history add foreign key (h_c_w_id, h_c_d_id, h_c_id) references customer (c_w_id, c_d_id, c_id) not valid`,
					`alter table history add foreign key (h_w_id, h_d_id) references district (d_w_id, d_id) not valid`,
					`alter table "order" add foreign key (o_w_id, o_d_id, o_c_id) references customer (c_w_id, c_d_id, c_id) not valid`,
					`alter table new_order add foreign key (no_w_id, no_d_id, no_o_id) references "order" (o_w_id, o_d_id, o_id) not valid`,
					`alter table stock add foreign key (s_w_id) references warehouse (w_id) not valid`,
					`alter table stock add foreign key (s_i_id) references item (i_id) not valid`,
					`alter table order_line add foreign key (ol_w_id, ol_d_id, ol_o_id) references "order" (o_w_id, o_d_id, o_id) not valid`,
					`alter table order_line add foreign key (ol_supply_w_id, ol_i_id) references stock (s_w_id, s_i_id) not valid`,
				}

				for _, fkStmt := range fkStmts {
					__antithesis_instrumentation__.Notify(698457)
					if _, err := db.Exec(fkStmt); err != nil {
						__antithesis_instrumentation__.Notify(698458)
						const duplFKErr = "columns cannot be used by multiple foreign key constraints"
						const idxErr = "foreign key requires an existing index on columns"
						switch {
						case strings.Contains(err.Error(), idxErr):
							__antithesis_instrumentation__.Notify(698459)
							fmt.Println(errors.WithHint(err, "try using the --deprecated-fk-indexes flag"))

							return errors.WithHint(err, "try using the --deprecated-fk-indexes flag")
						case strings.Contains(err.Error(), duplFKErr):
							__antithesis_instrumentation__.Notify(698460)

						default:
							__antithesis_instrumentation__.Notify(698461)
							return err
						}
					} else {
						__antithesis_instrumentation__.Notify(698462)
					}
				}
				__antithesis_instrumentation__.Notify(698456)

				if len(w.multiRegionCfg.regions) > 0 {
					__antithesis_instrumentation__.Notify(698463)
					if _, err := db.Exec(fmt.Sprintf(`ALTER TABLE item SET %s`, localityGlobalSuffix)); err != nil {
						__antithesis_instrumentation__.Notify(698464)
						return err
					} else {
						__antithesis_instrumentation__.Notify(698465)
					}
				} else {
					__antithesis_instrumentation__.Notify(698466)
				}
			} else {
				__antithesis_instrumentation__.Notify(698467)
			}
			__antithesis_instrumentation__.Notify(698453)

			if len(w.multiRegionCfg.regions) != 0 {
				__antithesis_instrumentation__.Notify(698468)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(698469)
			}
			__antithesis_instrumentation__.Notify(698454)
			return w.partitionAndScatterWithDB(db)
		},
		PostRun: func(startElapsed time.Duration) error {
			__antithesis_instrumentation__.Notify(698470)
			w.auditor.runChecks(w.localWarehouses)
			const totalHeader = "\n_elapsed_______tpmC____efc__avg(ms)__p50(ms)__p90(ms)__p95(ms)__p99(ms)_pMax(ms)"
			fmt.Println(totalHeader)

			const newOrderName = `newOrder`
			w.reg.Tick(func(t histogram.Tick) {
				__antithesis_instrumentation__.Notify(698472)
				if newOrderName == t.Name {
					__antithesis_instrumentation__.Notify(698473)
					tpmC := float64(t.Cumulative.TotalCount()) / startElapsed.Seconds() * 60
					fmt.Printf("%7.1fs %10.1f %5.1f%% %8.1f %8.1f %8.1f %8.1f %8.1f %8.1f\n",
						startElapsed.Seconds(),
						tpmC,
						100*tpmC/(SpecWarehouseFactor*float64(w.activeWarehouses)),
						time.Duration(t.Cumulative.Mean()).Seconds()*1000,
						time.Duration(t.Cumulative.ValueAtQuantile(50)).Seconds()*1000,
						time.Duration(t.Cumulative.ValueAtQuantile(90)).Seconds()*1000,
						time.Duration(t.Cumulative.ValueAtQuantile(95)).Seconds()*1000,
						time.Duration(t.Cumulative.ValueAtQuantile(99)).Seconds()*1000,
						time.Duration(t.Cumulative.ValueAtQuantile(100)).Seconds()*1000,
					)
				} else {
					__antithesis_instrumentation__.Notify(698474)
				}
			})
			__antithesis_instrumentation__.Notify(698471)
			return nil
		},
		CheckConsistency: func(ctx context.Context, db *gosql.DB) error {
			__antithesis_instrumentation__.Notify(698475)
			for _, check := range AllChecks() {
				__antithesis_instrumentation__.Notify(698477)
				if !w.expensiveChecks && func() bool {
					__antithesis_instrumentation__.Notify(698479)
					return check.Expensive == true
				}() == true {
					__antithesis_instrumentation__.Notify(698480)
					continue
				} else {
					__antithesis_instrumentation__.Notify(698481)
				}
				__antithesis_instrumentation__.Notify(698478)
				start := timeutil.Now()
				err := check.Fn(db, "")
				log.Infof(ctx, `check %s took %s`, check.Name, timeutil.Since(start))
				if err != nil {
					__antithesis_instrumentation__.Notify(698482)
					return errors.Wrapf(err, `check failed: %s`, check.Name)
				} else {
					__antithesis_instrumentation__.Notify(698483)
				}
			}
			__antithesis_instrumentation__.Notify(698476)
			return nil
		},
	}
}

func (w *tpcc) Tables() []workload.Table {
	__antithesis_instrumentation__.Notify(698484)
	aCharsInit := workloadimpl.PrecomputedRandInit(rand.New(rand.NewSource(w.seed)), precomputedLength, aCharsAlphabet)
	lettersInit := workloadimpl.PrecomputedRandInit(rand.New(rand.NewSource(w.seed)), precomputedLength, lettersAlphabet)
	numbersInit := workloadimpl.PrecomputedRandInit(rand.New(rand.NewSource(w.seed)), precomputedLength, numbersAlphabet)
	if w.localsPool == nil {
		__antithesis_instrumentation__.Notify(698492)
		w.localsPool = &sync.Pool{
			New: func() interface{} {
				__antithesis_instrumentation__.Notify(698493)
				return &generateLocals{
					rng: tpccRand{
						Rand: rand.New(rand.NewSource(uint64(timeutil.Now().UnixNano()))),

						aChars:  aCharsInit(),
						letters: lettersInit(),
						numbers: numbersInit(),
					},
				}
			},
		}
	} else {
		__antithesis_instrumentation__.Notify(698494)
	}
	__antithesis_instrumentation__.Notify(698485)

	splits := func(t workload.BatchedTuples) workload.BatchedTuples {
		__antithesis_instrumentation__.Notify(698495)
		if w.split {
			__antithesis_instrumentation__.Notify(698497)
			return t
		} else {
			__antithesis_instrumentation__.Notify(698498)
		}
		__antithesis_instrumentation__.Notify(698496)
		return workload.BatchedTuples{}
	}
	__antithesis_instrumentation__.Notify(698486)

	numBatches := func(total, per int) int {
		__antithesis_instrumentation__.Notify(698499)
		batches := total / per
		if total%per == 0 {
			__antithesis_instrumentation__.Notify(698501)
			batches--
		} else {
			__antithesis_instrumentation__.Notify(698502)
		}
		__antithesis_instrumentation__.Notify(698500)
		return batches
	}
	__antithesis_instrumentation__.Notify(698487)
	warehouse := workload.Table{
		Name: `warehouse`,
		Schema: makeSchema(
			tpccWarehouseSchema,
			maybeAddLocalityRegionalByRow(w.multiRegionCfg, `w_id`),
			maybeAddColumnFamiliesSuffix(
				w.separateColumnFamilies,
				tpccWarehouseColumnFamiliesSuffix,
			),
		),
		InitialRows: workload.BatchedTuples{
			NumBatches: w.warehouses,
			FillBatch:  w.tpccWarehouseInitialRowBatch,
		},
		Splits: splits(workload.Tuples(
			numBatches(w.warehouses, numWarehousesPerRange),
			func(i int) []interface{} {
				__antithesis_instrumentation__.Notify(698503)
				return []interface{}{(i + 1) * numWarehousesPerRange}
			},
		)),
		Stats: w.tpccWarehouseStats(),
	}
	__antithesis_instrumentation__.Notify(698488)
	district := workload.Table{
		Name: `district`,
		Schema: makeSchema(
			tpccDistrictSchemaBase,
			maybeAddColumnFamiliesSuffix(
				w.separateColumnFamilies,
				tpccDistrictColumnFamiliesSuffix,
			),
			maybeAddLocalityRegionalByRow(w.multiRegionCfg, `d_w_id`),
		),
		InitialRows: workload.BatchedTuples{
			NumBatches: numDistrictsPerWarehouse * w.warehouses,
			FillBatch:  w.tpccDistrictInitialRowBatch,
		},
		Splits: splits(workload.Tuples(
			numBatches(w.warehouses, numWarehousesPerRange),
			func(i int) []interface{} {
				__antithesis_instrumentation__.Notify(698504)
				return []interface{}{(i + 1) * numWarehousesPerRange, 0}
			},
		)),
		Stats: w.tpccDistrictStats(),
	}
	__antithesis_instrumentation__.Notify(698489)
	customer := workload.Table{
		Name: `customer`,
		Schema: makeSchema(
			tpccCustomerSchemaBase,
			maybeAddColumnFamiliesSuffix(
				w.separateColumnFamilies,
				tpccCustomerColumnFamiliesSuffix,
			),
			maybeAddLocalityRegionalByRow(w.multiRegionCfg, `c_w_id`),
		),
		InitialRows: workload.BatchedTuples{
			NumBatches: numCustomersPerWarehouse * w.warehouses,
			FillBatch:  w.tpccCustomerInitialRowBatch,
		},
		Stats: w.tpccCustomerStats(),
	}
	history := workload.Table{
		Name: `history`,
		Schema: makeSchema(
			tpccHistorySchemaBase,
			maybeAddFkSuffix(
				w.deprecatedFkIndexes,
				deprecatedTpccHistorySchemaFkSuffix,
			),
			maybeAddLocalityRegionalByRow(w.multiRegionCfg, `h_w_id`),
		),
		InitialRows: workload.BatchedTuples{
			NumBatches: numHistoryPerWarehouse * w.warehouses,
			FillBatch:  w.tpccHistoryInitialRowBatch,
		},
		Splits: splits(workload.Tuples(
			numBatches(w.warehouses, numWarehousesPerRange),
			func(i int) []interface{} {
				__antithesis_instrumentation__.Notify(698505)
				return []interface{}{(i + 1) * numWarehousesPerRange}
			},
		)),
		Stats: w.tpccHistoryStats(),
	}
	__antithesis_instrumentation__.Notify(698490)
	order := workload.Table{
		Name: `order`,
		Schema: makeSchema(
			tpccOrderSchemaBase,
			maybeAddLocalityRegionalByRow(w.multiRegionCfg, `o_w_id`),
		),
		InitialRows: workload.BatchedTuples{
			NumBatches: numOrdersPerWarehouse * w.warehouses,
			FillBatch:  w.tpccOrderInitialRowBatch,
		},
		Stats: w.tpccOrderStats(),
	}
	newOrder := workload.Table{
		Name: `new_order`,
		Schema: makeSchema(
			tpccNewOrderSchema,
			maybeAddLocalityRegionalByRow(w.multiRegionCfg, `no_w_id`),
		),
		InitialRows: workload.BatchedTuples{
			NumBatches: numNewOrdersPerWarehouse * w.warehouses,
			FillBatch:  w.tpccNewOrderInitialRowBatch,
		},
		Stats: w.tpccNewOrderStats(),
	}
	item := workload.Table{
		Name: `item`,
		Schema: makeSchema(
			tpccItemSchema,
		),
		InitialRows: workload.BatchedTuples{
			NumBatches: numItems,
			FillBatch:  w.tpccItemInitialRowBatch,
		},
		Splits: splits(workload.Tuples(
			numBatches(numItems, numItemsPerRange),
			func(i int) []interface{} {
				__antithesis_instrumentation__.Notify(698506)
				return []interface{}{numItemsPerRange * (i + 1)}
			},
		)),
		Stats: w.tpccItemStats(),
	}
	__antithesis_instrumentation__.Notify(698491)
	stock := workload.Table{
		Name: `stock`,
		Schema: makeSchema(
			tpccStockSchemaBase,
			maybeAddFkSuffix(
				w.deprecatedFkIndexes,
				deprecatedTpccStockSchemaFkSuffix,
			),
			maybeAddLocalityRegionalByRow(w.multiRegionCfg, `s_w_id`),
		),
		InitialRows: workload.BatchedTuples{
			NumBatches: numStockPerWarehouse * w.warehouses,
			FillBatch:  w.tpccStockInitialRowBatch,
		},
		Stats: w.tpccStockStats(),
	}
	orderLine := workload.Table{
		Name: `order_line`,
		Schema: makeSchema(
			tpccOrderLineSchemaBase,
			maybeAddFkSuffix(
				w.deprecatedFkIndexes,
				deprecatedTpccOrderLineSchemaFkSuffix,
			),
			maybeAddLocalityRegionalByRow(w.multiRegionCfg, `ol_w_id`),
		),
		InitialRows: workload.BatchedTuples{
			NumBatches: numOrdersPerWarehouse * w.warehouses,
			FillBatch:  w.tpccOrderLineInitialRowBatch,
		},
		Stats: w.tpccOrderLineStats(),
	}
	return []workload.Table{
		warehouse, district, customer, history, order, newOrder, item, stock, orderLine,
	}
}

func (w *tpcc) Ops(
	ctx context.Context, urls []string, reg *histogram.Registry,
) (workload.QueryLoad, error) {
	__antithesis_instrumentation__.Notify(698507)
	if len(w.multiRegionCfg.regions) == 0 {
		__antithesis_instrumentation__.Notify(698522)

		if err := w.partitionAndScatter(urls); err != nil {
			__antithesis_instrumentation__.Notify(698523)
			return workload.QueryLoad{}, err
		} else {
			__antithesis_instrumentation__.Notify(698524)
		}
	} else {
		__antithesis_instrumentation__.Notify(698525)
	}
	__antithesis_instrumentation__.Notify(698508)

	if w.reg == nil {
		__antithesis_instrumentation__.Notify(698526)
		w.reg = reg
		w.txCounters = setupTPCCMetrics(reg.Registerer())
	} else {
		__antithesis_instrumentation__.Notify(698527)
	}
	__antithesis_instrumentation__.Notify(698509)

	sqlDatabase, err := workload.SanitizeUrls(w, w.dbOverride, urls)
	if err != nil {
		__antithesis_instrumentation__.Notify(698528)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(698529)
	}
	__antithesis_instrumentation__.Notify(698510)
	parsedURL, err := url.Parse(urls[0])
	if err != nil {
		__antithesis_instrumentation__.Notify(698530)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(698531)
	}
	__antithesis_instrumentation__.Notify(698511)

	w.usePostgres = parsedURL.Port() == "5432"

	cfg := workload.MultiConnPoolCfg{
		MaxTotalConnections: (w.numConns + len(urls) - 1) / len(urls),

		MaxConnsPerPool: 50,
	}
	fmt.Printf("Initializing %d connections...\n", w.numConns)

	dbs := make([]*workload.MultiConnPool, len(urls))
	var g errgroup.Group
	for i := range urls {
		__antithesis_instrumentation__.Notify(698532)
		i := i
		g.Go(func() error {
			__antithesis_instrumentation__.Notify(698533)
			var err error
			dbs[i], err = workload.NewMultiConnPool(ctx, cfg, urls[i])
			return err
		})
	}
	__antithesis_instrumentation__.Notify(698512)
	if err := g.Wait(); err != nil {
		__antithesis_instrumentation__.Notify(698534)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(698535)
	}
	__antithesis_instrumentation__.Notify(698513)
	var partitionDBs [][]*workload.MultiConnPool
	if w.clientPartitions > 0 {
		__antithesis_instrumentation__.Notify(698536)

		partitionDBs = make([][]*workload.MultiConnPool, w.clientPartitions)
	} else {
		__antithesis_instrumentation__.Notify(698537)

		partitionDBs = make([][]*workload.MultiConnPool, w.partitions)
	}
	__antithesis_instrumentation__.Notify(698514)

	if len(w.affinityPartitions) == 1 {
		__antithesis_instrumentation__.Notify(698538)

		partitionDBs[w.affinityPartitions[0]] = dbs

	} else {
		__antithesis_instrumentation__.Notify(698539)

		for i, db := range dbs {
			__antithesis_instrumentation__.Notify(698541)
			p := i % w.partitions
			partitionDBs[p] = append(partitionDBs[p], db)
		}
		__antithesis_instrumentation__.Notify(698540)
		for i := range partitionDBs {
			__antithesis_instrumentation__.Notify(698542)

			if partitionDBs[i] == nil {
				__antithesis_instrumentation__.Notify(698543)
				partitionDBs[i] = dbs
			} else {
				__antithesis_instrumentation__.Notify(698544)
			}
		}
	}
	__antithesis_instrumentation__.Notify(698515)

	fmt.Printf("Initializing %d idle connections...\n", w.idleConns)
	var conns []*pgx.Conn
	for i := 0; i < w.idleConns; i++ {
		__antithesis_instrumentation__.Notify(698545)
		for _, url := range urls {
			__antithesis_instrumentation__.Notify(698546)
			connConfig, err := pgx.ParseConfig(url)
			if err != nil {
				__antithesis_instrumentation__.Notify(698549)
				return workload.QueryLoad{}, err
			} else {
				__antithesis_instrumentation__.Notify(698550)
			}
			__antithesis_instrumentation__.Notify(698547)
			conn, err := pgx.ConnectConfig(ctx, connConfig)
			if err != nil {
				__antithesis_instrumentation__.Notify(698551)
				return workload.QueryLoad{}, err
			} else {
				__antithesis_instrumentation__.Notify(698552)
			}
			__antithesis_instrumentation__.Notify(698548)
			conns = append(conns, conn)
		}
	}
	__antithesis_instrumentation__.Notify(698516)
	fmt.Printf("Initializing %d workers and preparing statements...\n", w.workers)
	ql := workload.QueryLoad{SQLDatabase: sqlDatabase}
	ql.WorkerFns = make([]func(context.Context) error, 0, w.workers)
	var group errgroup.Group

	isMyPart := func(p int) bool {
		__antithesis_instrumentation__.Notify(698553)
		for _, ap := range w.affinityPartitions {
			__antithesis_instrumentation__.Notify(698555)
			if p == ap {
				__antithesis_instrumentation__.Notify(698556)
				return true
			} else {
				__antithesis_instrumentation__.Notify(698557)
			}
		}
		__antithesis_instrumentation__.Notify(698554)

		return len(w.affinityPartitions) == 0
	}
	__antithesis_instrumentation__.Notify(698517)

	sem := make(chan struct{}, 100)
	for workerIdx := 0; workerIdx < w.workers; workerIdx++ {
		__antithesis_instrumentation__.Notify(698558)
		workerIdx := workerIdx
		var warehouse int
		var p int
		if len(w.multiRegionCfg.regions) == 0 {
			__antithesis_instrumentation__.Notify(698561)
			warehouse = w.wPart.totalElems[workerIdx%len(w.wPart.totalElems)]
			p = w.wPart.partElemsMap[warehouse]
		} else {
			__antithesis_instrumentation__.Notify(698562)

			warehouse = w.wMRPart.totalElems[workerIdx%len(w.wMRPart.totalElems)]
			p = w.wMRPart.partElemsMap[warehouse]
		}
		__antithesis_instrumentation__.Notify(698559)

		if !isMyPart(p) {
			__antithesis_instrumentation__.Notify(698563)
			continue
		} else {
			__antithesis_instrumentation__.Notify(698564)
		}
		__antithesis_instrumentation__.Notify(698560)
		dbs := partitionDBs[p]
		db := dbs[warehouse%len(dbs)]

		ql.WorkerFns = append(ql.WorkerFns, nil)
		idx := len(ql.WorkerFns) - 1
		sem <- struct{}{}
		group.Go(func() error {
			__antithesis_instrumentation__.Notify(698565)
			worker, err := newWorker(ctx, w, db, reg.GetHandle(), w.txCounters, warehouse)
			if err == nil {
				__antithesis_instrumentation__.Notify(698567)
				ql.WorkerFns[idx] = worker.run
			} else {
				__antithesis_instrumentation__.Notify(698568)
			}
			__antithesis_instrumentation__.Notify(698566)
			<-sem
			return err
		})
	}
	__antithesis_instrumentation__.Notify(698518)
	if err := group.Wait(); err != nil {
		__antithesis_instrumentation__.Notify(698569)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(698570)
	}
	__antithesis_instrumentation__.Notify(698519)

	for _, tx := range allTxs {
		__antithesis_instrumentation__.Notify(698571)
		reg.GetHandle().Get(tx.name)
	}
	__antithesis_instrumentation__.Notify(698520)

	ql.Close = func(context context.Context) {
		__antithesis_instrumentation__.Notify(698572)
		for _, conn := range conns {
			__antithesis_instrumentation__.Notify(698573)
			if err := conn.Close(ctx); err != nil {
				__antithesis_instrumentation__.Notify(698574)
				log.Warningf(ctx, "%v", err)
			} else {
				__antithesis_instrumentation__.Notify(698575)
			}
		}
	}
	__antithesis_instrumentation__.Notify(698521)
	return ql, nil
}

func (w *tpcc) partitionAndScatter(urls []string) error {
	__antithesis_instrumentation__.Notify(698576)
	db, err := gosql.Open(`cockroach`, strings.Join(urls, ` `))
	if err != nil {
		__antithesis_instrumentation__.Notify(698578)
		return err
	} else {
		__antithesis_instrumentation__.Notify(698579)
	}
	__antithesis_instrumentation__.Notify(698577)
	defer db.Close()
	return w.partitionAndScatterWithDB(db)
}

func (w *tpcc) partitionAndScatterWithDB(db *gosql.DB) error {
	__antithesis_instrumentation__.Notify(698580)
	if w.partitions > 1 {
		__antithesis_instrumentation__.Notify(698583)

		if parts, err := partitionCount(db); err != nil {
			__antithesis_instrumentation__.Notify(698584)
			return errors.Wrapf(err, "could not determine if tables are partitioned")
		} else {
			__antithesis_instrumentation__.Notify(698585)
			if parts == 0 {
				__antithesis_instrumentation__.Notify(698586)
				if err := partitionTables(db, w.zoneCfg, w.wPart, w.replicateStaticColumns); err != nil {
					__antithesis_instrumentation__.Notify(698587)
					return errors.Wrapf(err, "could not partition tables")
				} else {
					__antithesis_instrumentation__.Notify(698588)
				}
			} else {
				__antithesis_instrumentation__.Notify(698589)
				if parts != w.partitions {
					__antithesis_instrumentation__.Notify(698590)
					return errors.Errorf("tables are not partitioned %d way(s). "+
						"Pass the --partitions flag to 'workload init' or 'workload fixtures'.", w.partitions)
				} else {
					__antithesis_instrumentation__.Notify(698591)
				}
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(698592)
	}
	__antithesis_instrumentation__.Notify(698581)

	if w.scatter {
		__antithesis_instrumentation__.Notify(698593)
		if err := scatterRanges(db); err != nil {
			__antithesis_instrumentation__.Notify(698594)
			return errors.Wrapf(err, "could not scatter ranges")
		} else {
			__antithesis_instrumentation__.Notify(698595)
		}
	} else {
		__antithesis_instrumentation__.Notify(698596)
	}
	__antithesis_instrumentation__.Notify(698582)

	return nil
}
