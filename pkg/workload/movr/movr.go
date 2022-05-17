package movr

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	gosql "database/sql"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"github.com/cockroachdb/cockroach/pkg/workload/faker"
	"github.com/cockroachdb/errors"
	"github.com/spf13/pflag"
	"golang.org/x/exp/rand"
)

func (g *movr) maybeFormatWithCity(fmtString string) string {
	__antithesis_instrumentation__.Notify(694752)
	var posOne, posTwo string
	if !g.multiRegion {
		__antithesis_instrumentation__.Notify(694754)
		posOne = "city, "
		posTwo = "vehicle_city, "
	} else {
		__antithesis_instrumentation__.Notify(694755)
	}
	__antithesis_instrumentation__.Notify(694753)
	return fmt.Sprintf(fmtString, posOne, posTwo)
}

const (
	TablesUsersIdx                    = 0
	TablesVehiclesIdx                 = 1
	TablesRidesIdx                    = 2
	TablesVehicleLocationHistoriesIdx = 3
	TablesPromoCodesIdx               = 4
	TablesUserPromoCodesIdx           = 5
)

func (g *movr) movrUsersSchema() string {
	__antithesis_instrumentation__.Notify(694756)
	return g.maybeFormatWithCity(`(
  id UUID NOT NULL,
  city VARCHAR NOT NULL,
  name VARCHAR NULL,
  address VARCHAR NULL,
  credit_card VARCHAR NULL,
  PRIMARY KEY (%[1]sid ASC)
)`)
}

const (
	usersIDIdx   = 0
	usersCityIdx = 1
)

func (g *movr) movrVehiclesSchema() string {
	__antithesis_instrumentation__.Notify(694757)
	return g.maybeFormatWithCity(`(
  id UUID NOT NULL,
  city VARCHAR NOT NULL,
  type VARCHAR NULL,
  owner_id UUID NULL,
  creation_time TIMESTAMP NULL,
  status VARCHAR NULL,
  current_location VARCHAR NULL,
  ext JSONB NULL,
  PRIMARY KEY (%[1]sid ASC),
  INDEX vehicles_auto_index_fk_city_ref_users (%[1]sowner_id ASC)
)`,
	)
}

const (
	vehiclesIDIdx   = 0
	vehiclesCityIdx = 1
)

func (g *movr) movrRidesSchema() string {
	__antithesis_instrumentation__.Notify(694758)
	return g.maybeFormatWithCity(`(
  id UUID NOT NULL,
  city VARCHAR NOT NULL,
  vehicle_city VARCHAR NULL,
  rider_id UUID NULL,
  vehicle_id UUID NULL,
  start_address VARCHAR NULL,
  end_address VARCHAR NULL,
  start_time TIMESTAMP NULL,
  end_time TIMESTAMP NULL,
  revenue DECIMAL(10,2) NULL,
  PRIMARY KEY (%[1]sid ASC),
  INDEX rides_auto_index_fk_city_ref_users (%[1]srider_id ASC),
  INDEX rides_auto_index_fk_vehicle_city_ref_vehicles (%[2]svehicle_id ASC),
  CONSTRAINT check_vehicle_city_city CHECK (vehicle_city = city)
)`)
}

const (
	ridesIDIdx   = 0
	ridesCityIdx = 1
)

func (g *movr) movrVehicleLocationHistoriesSchema() string {
	__antithesis_instrumentation__.Notify(694759)
	return g.maybeFormatWithCity(`(
  city VARCHAR NOT NULL,
  ride_id UUID NOT NULL,
  "timestamp" TIMESTAMP NOT NULL,
  lat FLOAT8 NULL,
  long FLOAT8 NULL,
  PRIMARY KEY (%[1]sride_id ASC, "timestamp" ASC)
)`)
}

const movrPromoCodesSchema = `(
  code VARCHAR NOT NULL,
  description VARCHAR NULL,
  creation_time TIMESTAMP NULL,
  expiration_time TIMESTAMP NULL,
  rules JSONB NULL,
  PRIMARY KEY (code ASC)
)`

func (g *movr) movrUserPromoCodesSchema() string {
	__antithesis_instrumentation__.Notify(694760)
	return g.maybeFormatWithCity(`(
  city VARCHAR NOT NULL,
  user_id UUID NOT NULL,
  code VARCHAR NOT NULL,
  "timestamp" TIMESTAMP NULL,
  usage_count INT NULL,
  PRIMARY KEY (%[1]suser_id ASC, code ASC)
)`)
}

var rbrTables = []string{
	"users",
	"vehicles",
	"rides",
	"vehicle_location_histories",
	"user_promo_codes",
}

var globalTables = []string{
	"promo_codes",
}

const (
	timestampFormat = "2006-01-02 15:04:05.999999-07:00"
)

var cities = []struct {
	city   string
	region string
}{
	{city: "new york", region: "us-east1"},
	{city: "boston", region: "us-east1"},
	{city: "washington dc", region: "us-east1"},
	{city: "seattle", region: "us-west1"},
	{city: "san francisco", region: "us-west1"},
	{city: "los angeles", region: "us-west1"},
	{city: "amsterdam", region: "europe-west1"},
	{city: "paris", region: "europe-west1"},
	{city: "rome", region: "europe-west1"},
}

type movr struct {
	flags     workload.Flags
	connFlags *workload.ConnFlags

	seed                              uint64
	users, vehicles, rides, histories cityDistributor
	numPromoCodes                     int
	numUserPromoCodes                 int
	ranges                            int

	multiRegion           bool
	inferCRDBRegionColumn bool
	survivalGoal          string

	creationTime time.Time

	fakerOnce sync.Once
	faker     faker.Faker
}

func init() {
	workload.Register(movrMeta)
}

var movrMeta = workload.Meta{
	Name:         `movr`,
	Description:  `MovR is a fictional vehicle sharing company`,
	Version:      `1.0.0`,
	PublicFacing: true,
	New: func() workload.Generator {
		__antithesis_instrumentation__.Notify(694761)
		g := &movr{}
		g.flags.FlagSet = pflag.NewFlagSet(`movr`, pflag.ContinueOnError)
		g.flags.Uint64Var(&g.seed, `seed`, 1, `Key hash seed.`)
		g.flags.IntVar(&g.users.numRows, `num-users`, 50, `Initial number of users.`)
		g.flags.IntVar(&g.vehicles.numRows, `num-vehicles`, 15, `Initial number of vehicles.`)
		g.flags.IntVar(&g.rides.numRows, `num-rides`, 500, `Initial number of rides.`)
		g.flags.BoolVar(
			&g.multiRegion,
			`multi-region`,
			false,
			`Whether to use the multi-region configuration for partitions.`,
		)
		g.flags.BoolVar(
			&g.inferCRDBRegionColumn,
			`infer-crdb-region-column`,
			true,
			`Whether to infer crdb_region for multi-region REGIONAL BY ROW columns using the city column.
Otherwise defaults to the gateway_region.`,
		)
		g.flags.StringVar(
			&g.survivalGoal,
			"survive",
			"az",
			"Survival goal for multi-region workloads. Supported goals: az, region.",
		)
		g.flags.IntVar(&g.histories.numRows, `num-histories`, 1000,
			`Initial number of ride location histories.`)
		g.flags.IntVar(&g.numPromoCodes, `num-promo-codes`, 1000, `Initial number of promo codes.`)
		g.flags.IntVar(&g.numUserPromoCodes, `num-user-promos`, 5, `Initial number of promo codes in use.`)
		g.flags.IntVar(&g.ranges, `num-ranges`, 9, `Initial number of ranges to break the tables into`)
		g.connFlags = workload.NewConnFlags(&g.flags)
		g.creationTime = time.Date(2019, 1, 2, 3, 4, 5, 6, time.UTC)
		return g
	},
}

func (*movr) Meta() workload.Meta { __antithesis_instrumentation__.Notify(694762); return movrMeta }

func (g *movr) Flags() workload.Flags { __antithesis_instrumentation__.Notify(694763); return g.flags }

func (g *movr) Hooks() workload.Hooks {
	__antithesis_instrumentation__.Notify(694764)
	return workload.Hooks{
		Validate: func() error {
			__antithesis_instrumentation__.Notify(694765)

			if g.users.numRows < len(cities) {
				__antithesis_instrumentation__.Notify(694770)
				return errors.Errorf(`at least %d users are required`, len(cities))
			} else {
				__antithesis_instrumentation__.Notify(694771)
			}
			__antithesis_instrumentation__.Notify(694766)
			if g.vehicles.numRows < len(cities) {
				__antithesis_instrumentation__.Notify(694772)
				return errors.Errorf(`at least %d vehicles are required`, len(cities))
			} else {
				__antithesis_instrumentation__.Notify(694773)
			}
			__antithesis_instrumentation__.Notify(694767)
			if g.rides.numRows < len(cities) {
				__antithesis_instrumentation__.Notify(694774)
				return errors.Errorf(`at least %d rides are required`, len(cities))
			} else {
				__antithesis_instrumentation__.Notify(694775)
			}
			__antithesis_instrumentation__.Notify(694768)
			if g.histories.numRows < len(cities) {
				__antithesis_instrumentation__.Notify(694776)
				return errors.Errorf(`at least %d histories are required`, len(cities))
			} else {
				__antithesis_instrumentation__.Notify(694777)
			}
			__antithesis_instrumentation__.Notify(694769)
			return nil
		},
		PostLoad: func(db *gosql.DB) error {
			__antithesis_instrumentation__.Notify(694778)
			fkStmts := []string{
				g.maybeFormatWithCity(
					`ALTER TABLE vehicles ADD FOREIGN KEY
						(%[1]sowner_id) REFERENCES users (%[1]sid)`,
				),
				g.maybeFormatWithCity(
					`ALTER TABLE rides ADD FOREIGN KEY
						(%[1]srider_id) REFERENCES users (%[1]sid)`,
				),
				g.maybeFormatWithCity(
					`ALTER TABLE rides ADD FOREIGN KEY
						(%[2]svehicle_id) REFERENCES vehicles (%[1]sid)`,
				),
				g.maybeFormatWithCity(
					`ALTER TABLE vehicle_location_histories ADD FOREIGN KEY
						(%[1]sride_id) REFERENCES rides (%[1]sid)`,
				),
				g.maybeFormatWithCity(
					`ALTER TABLE user_promo_codes ADD FOREIGN KEY
						(%[1]suser_id) REFERENCES users (%[1]sid)`,
				),
			}

			for _, fkStmt := range fkStmts {
				__antithesis_instrumentation__.Notify(694780)
				if _, err := db.Exec(fkStmt); err != nil {
					__antithesis_instrumentation__.Notify(694781)

					const duplicateFKErr = "columns cannot be used by multiple foreign key constraints"
					if !strings.Contains(err.Error(), duplicateFKErr) {
						__antithesis_instrumentation__.Notify(694782)
						return err
					} else {
						__antithesis_instrumentation__.Notify(694783)
					}
				} else {
					__antithesis_instrumentation__.Notify(694784)
				}
			}
			__antithesis_instrumentation__.Notify(694779)
			return nil
		},

		Partition: func(db *gosql.DB) error {
			__antithesis_instrumentation__.Notify(694785)
			if g.multiRegion {
				__antithesis_instrumentation__.Notify(694788)
				var survivalGoal string
				switch g.survivalGoal {
				case "az":
					__antithesis_instrumentation__.Notify(694793)
					survivalGoal = "ZONE"
				case "region":
					__antithesis_instrumentation__.Notify(694794)
					survivalGoal = "REGION"
				default:
					__antithesis_instrumentation__.Notify(694795)
					return errors.Errorf("unsupported survival goal: %s", g.survivalGoal)
				}
				__antithesis_instrumentation__.Notify(694789)
				qs := fmt.Sprintf(
					`
ALTER DATABASE %[1]s SET PRIMARY REGION "us-east1";
ALTER DATABASE %[1]s ADD REGION "us-west1";
ALTER DATABASE %[1]s ADD REGION "europe-west1";
ALTER DATABASE %[1]s SURVIVE %s FAILURE
`,
					g.Meta().Name,
					survivalGoal,
				)
				for _, q := range strings.Split(qs, ";") {
					__antithesis_instrumentation__.Notify(694796)
					if _, err := db.Exec(q); err != nil {
						__antithesis_instrumentation__.Notify(694797)
						return err
					} else {
						__antithesis_instrumentation__.Notify(694798)
					}
				}
				__antithesis_instrumentation__.Notify(694790)
				for _, rbrTable := range rbrTables {
					__antithesis_instrumentation__.Notify(694799)
					if g.inferCRDBRegionColumn {
						__antithesis_instrumentation__.Notify(694801)
						regionToCities := make(map[string][]string)
						for _, city := range cities {
							__antithesis_instrumentation__.Notify(694804)
							regionToCities[city.region] = append(regionToCities[city.region], city.city)
						}
						__antithesis_instrumentation__.Notify(694802)
						cityClauses := make([]string, 0, len(regionToCities))
						for region, cities := range regionToCities {
							__antithesis_instrumentation__.Notify(694805)
							cityClauses = append(
								cityClauses,
								fmt.Sprintf(
									`WHEN city IN (%s) THEN '%s'`,
									`'`+strings.Join(cities, `', '`)+`'`,
									region,
								),
							)
						}
						__antithesis_instrumentation__.Notify(694803)
						sort.Strings(cityClauses)
						if _, err := db.Exec(
							fmt.Sprintf(
								`ALTER TABLE %s ADD COLUMN crdb_region crdb_internal_region NOT NULL AS (
								CASE
									%s
									ELSE 'us-east1'
							END) STORED`,
								rbrTable,
								strings.Join(cityClauses, "\n"),
							),
						); err != nil {
							__antithesis_instrumentation__.Notify(694806)
							return err
						} else {
							__antithesis_instrumentation__.Notify(694807)
						}
					} else {
						__antithesis_instrumentation__.Notify(694808)
					}
					__antithesis_instrumentation__.Notify(694800)
					if _, err := db.Exec(
						fmt.Sprintf(
							`ALTER TABLE %s SET LOCALITY REGIONAL BY ROW`,
							rbrTable,
						),
					); err != nil {
						__antithesis_instrumentation__.Notify(694809)
						return err
					} else {
						__antithesis_instrumentation__.Notify(694810)
					}
				}
				__antithesis_instrumentation__.Notify(694791)
				for _, globalTable := range globalTables {
					__antithesis_instrumentation__.Notify(694811)
					if _, err := db.Exec(
						fmt.Sprintf(
							`ALTER TABLE %s SET LOCALITY GLOBAL`,
							globalTable,
						),
					); err != nil {
						__antithesis_instrumentation__.Notify(694812)
						return err
					} else {
						__antithesis_instrumentation__.Notify(694813)
					}
				}
				__antithesis_instrumentation__.Notify(694792)

				return nil
			} else {
				__antithesis_instrumentation__.Notify(694814)
			}
			__antithesis_instrumentation__.Notify(694786)

			qs := []string{

				`ALTER TABLE users PARTITION BY LIST (city) (
					PARTITION us_west VALUES IN ('seattle', 'san francisco', 'los angeles'),
					PARTITION us_east VALUES IN ('new york', 'boston', 'washington dc'),
					PARTITION europe_west VALUES IN ('amsterdam', 'paris', 'rome')
				);`,
				`ALTER TABLE vehicles PARTITION BY LIST (city) (
					PARTITION us_west VALUES IN ('seattle', 'san francisco', 'los angeles'),
					PARTITION us_east VALUES IN ('new york', 'boston', 'washington dc'),
					PARTITION europe_west VALUES IN ('amsterdam', 'paris', 'rome')
				);`,
				`ALTER INDEX vehicles_auto_index_fk_city_ref_users PARTITION BY LIST (city) (
					PARTITION us_west VALUES IN ('seattle', 'san francisco', 'los angeles'),
					PARTITION us_east VALUES IN ('new york', 'boston', 'washington dc'),
					PARTITION europe_west VALUES IN ('amsterdam', 'paris', 'rome')
				);`,
				`ALTER TABLE rides PARTITION BY LIST (city) (
					PARTITION us_west VALUES IN ('seattle', 'san francisco', 'los angeles'),
					PARTITION us_east VALUES IN ('new york', 'boston', 'washington dc'),
					PARTITION europe_west VALUES IN ('amsterdam', 'paris', 'rome')
				);`,
				`ALTER INDEX rides_auto_index_fk_city_ref_users PARTITION BY LIST (city) (
					PARTITION us_west VALUES IN ('seattle', 'san francisco', 'los angeles'),
					PARTITION us_east VALUES IN ('new york', 'boston', 'washington dc'),
					PARTITION europe_west VALUES IN ('amsterdam', 'paris', 'rome')
				);`,
				`ALTER INDEX rides_auto_index_fk_vehicle_city_ref_vehicles PARTITION BY LIST (vehicle_city) (
					PARTITION us_west VALUES IN ('seattle', 'san francisco', 'los angeles'),
					PARTITION us_east VALUES IN ('new york', 'boston', 'washington dc'),
					PARTITION europe_west VALUES IN ('amsterdam', 'paris', 'rome')
				);`,
				`ALTER TABLE user_promo_codes PARTITION BY LIST (city) (
					PARTITION us_west VALUES IN ('seattle', 'san francisco', 'los angeles'),
					PARTITION us_east VALUES IN ('new york', 'boston', 'washington dc'),
					PARTITION europe_west VALUES IN ('amsterdam', 'paris', 'rome')
				);`,
				`ALTER TABLE vehicle_location_histories PARTITION BY LIST (city) (
					PARTITION us_west VALUES IN ('seattle', 'san francisco', 'los angeles'),
					PARTITION us_east VALUES IN ('new york', 'boston', 'washington dc'),
					PARTITION europe_west VALUES IN ('amsterdam', 'paris', 'rome')
				);`,

				`ALTER PARTITION us_west OF INDEX users@* CONFIGURE ZONE USING CONSTRAINTS='["+region=us-west1"]';`,
				`ALTER PARTITION us_east OF INDEX users@* CONFIGURE ZONE USING CONSTRAINTS='["+region=us-east1"]';`,
				`ALTER PARTITION europe_west OF INDEX users@* CONFIGURE ZONE USING CONSTRAINTS='["+region=europe-west1"]';`,

				`ALTER PARTITION us_west OF INDEX vehicles@* CONFIGURE ZONE USING CONSTRAINTS='["+region=us-west1"]';`,
				`ALTER PARTITION us_east OF INDEX vehicles@* CONFIGURE ZONE USING CONSTRAINTS='["+region=us-east1"]';`,
				`ALTER PARTITION europe_west OF INDEX vehicles@* CONFIGURE ZONE USING CONSTRAINTS='["+region=europe-west1"]';`,

				`ALTER PARTITION us_west OF INDEX rides@* CONFIGURE ZONE USING CONSTRAINTS='["+region=us-west1"]';`,
				`ALTER PARTITION us_east OF INDEX rides@* CONFIGURE ZONE USING CONSTRAINTS='["+region=us-east1"]';`,
				`ALTER PARTITION europe_west OF INDEX rides@* CONFIGURE ZONE USING CONSTRAINTS='["+region=europe-west1"]';`,

				`ALTER PARTITION us_west OF INDEX user_promo_codes@* CONFIGURE ZONE USING CONSTRAINTS='["+region=us-west1"]';`,
				`ALTER PARTITION us_east OF INDEX user_promo_codes@* CONFIGURE ZONE USING CONSTRAINTS='["+region=us-east1"]';`,
				`ALTER PARTITION europe_west OF INDEX user_promo_codes@* CONFIGURE ZONE USING CONSTRAINTS='["+region=europe-west1"]';`,

				`ALTER PARTITION us_west OF INDEX vehicle_location_histories@* CONFIGURE ZONE USING CONSTRAINTS='["+region=us-west1"]';`,
				`ALTER PARTITION us_east OF INDEX vehicle_location_histories@* CONFIGURE ZONE USING CONSTRAINTS='["+region=us-east1"]';`,
				`ALTER PARTITION europe_west OF INDEX vehicle_location_histories@* CONFIGURE ZONE USING CONSTRAINTS='["+region=europe-west1"]';`,

				`CREATE INDEX promo_codes_idx_us_west ON promo_codes (code) STORING (description, creation_time,expiration_time, rules);`,
				`CREATE INDEX promo_codes_idx_europe_west ON promo_codes (code) STORING (description, creation_time, expiration_time, rules);`,

				`ALTER TABLE promo_codes CONFIGURE ZONE USING num_replicas = 3,
					constraints = '{"+region=us-east1": 1}',
					lease_preferences = '[[+region=us-east1]]';`,
				`ALTER INDEX promo_codes@promo_codes_idx_us_west CONFIGURE ZONE USING
					num_replicas = 3,
					constraints = '{"+region=us-west1": 1}',
					lease_preferences = '[[+region=us-west1]]';`,
				`ALTER INDEX promo_codes@promo_codes_idx_europe_west CONFIGURE ZONE USING
					num_replicas = 3,
					constraints = '{"+region=europe-west1": 1}',
					lease_preferences = '[[+region=europe-west1]]';`,
			}
			for _, q := range qs {
				__antithesis_instrumentation__.Notify(694815)
				if _, err := db.Exec(q); err != nil {
					__antithesis_instrumentation__.Notify(694816)
					return err
				} else {
					__antithesis_instrumentation__.Notify(694817)
				}
			}
			__antithesis_instrumentation__.Notify(694787)
			return nil
		},
	}
}

func (g *movr) Tables() []workload.Table {
	__antithesis_instrumentation__.Notify(694818)
	g.fakerOnce.Do(func() {
		__antithesis_instrumentation__.Notify(694823)
		g.faker = faker.NewFaker()
	})
	__antithesis_instrumentation__.Notify(694819)
	tables := make([]workload.Table, 6)
	tables[TablesUsersIdx] = workload.Table{
		Name:   `users`,
		Schema: g.movrUsersSchema(),
		InitialRows: workload.Tuples(
			g.users.numRows,
			g.movrUsersInitialRow,
		),
		Splits: workload.Tuples(
			g.ranges-1,
			func(splitIdx int) []interface{} {
				__antithesis_instrumentation__.Notify(694824)
				row := g.movrUsersInitialRow((splitIdx + 1) * (g.users.numRows / g.ranges))
				if g.multiRegion {
					__antithesis_instrumentation__.Notify(694826)
					return []interface{}{row[usersIDIdx]}
				} else {
					__antithesis_instrumentation__.Notify(694827)
				}
				__antithesis_instrumentation__.Notify(694825)

				return []interface{}{row[usersCityIdx], row[usersIDIdx]}
			},
		),
	}
	__antithesis_instrumentation__.Notify(694820)
	tables[TablesVehiclesIdx] = workload.Table{
		Name:   `vehicles`,
		Schema: g.movrVehiclesSchema(),
		InitialRows: workload.Tuples(
			g.vehicles.numRows,
			g.movrVehiclesInitialRow,
		),
		Splits: workload.Tuples(
			g.ranges-1,
			func(splitIdx int) []interface{} {
				__antithesis_instrumentation__.Notify(694828)
				row := g.movrVehiclesInitialRow((splitIdx + 1) * (g.vehicles.numRows / g.ranges))
				if g.multiRegion {
					__antithesis_instrumentation__.Notify(694830)
					return []interface{}{row[vehiclesIDIdx]}
				} else {
					__antithesis_instrumentation__.Notify(694831)
				}
				__antithesis_instrumentation__.Notify(694829)

				return []interface{}{row[vehiclesCityIdx], row[vehiclesIDIdx]}
			},
		),
	}
	__antithesis_instrumentation__.Notify(694821)
	tables[TablesRidesIdx] = workload.Table{
		Name:   `rides`,
		Schema: g.movrRidesSchema(),
		InitialRows: workload.Tuples(
			g.rides.numRows,
			g.movrRidesInitialRow,
		),
		Splits: workload.Tuples(
			g.ranges-1,
			func(splitIdx int) []interface{} {
				__antithesis_instrumentation__.Notify(694832)
				row := g.movrRidesInitialRow((splitIdx + 1) * (g.rides.numRows / g.ranges))
				if g.multiRegion {
					__antithesis_instrumentation__.Notify(694834)
					return []interface{}{row[ridesIDIdx]}
				} else {
					__antithesis_instrumentation__.Notify(694835)
				}
				__antithesis_instrumentation__.Notify(694833)

				return []interface{}{row[ridesCityIdx], row[ridesIDIdx]}
			},
		),
	}
	__antithesis_instrumentation__.Notify(694822)
	tables[TablesVehicleLocationHistoriesIdx] = workload.Table{
		Name:   `vehicle_location_histories`,
		Schema: g.movrVehicleLocationHistoriesSchema(),
		InitialRows: workload.Tuples(
			g.histories.numRows,
			g.movrVehicleLocationHistoriesInitialRow,
		),
	}
	tables[TablesPromoCodesIdx] = workload.Table{
		Name:   `promo_codes`,
		Schema: movrPromoCodesSchema,
		InitialRows: workload.Tuples(
			g.numPromoCodes,
			g.movrPromoCodesInitialRow,
		),
	}
	tables[TablesUserPromoCodesIdx] = workload.Table{
		Name:   `user_promo_codes`,
		Schema: g.movrUserPromoCodesSchema(),
		InitialRows: workload.Tuples(
			g.numUserPromoCodes,
			g.movrUserPromoCodesInitialRow,
		),
	}
	return tables
}

type cityDistributor struct {
	numRows int
}

func (d cityDistributor) cityForRow(rowIdx int) int {
	__antithesis_instrumentation__.Notify(694836)
	if d.numRows < len(cities) {
		__antithesis_instrumentation__.Notify(694838)
		panic(errors.Errorf(`a minimum of %d rows are required got %d`, len(cities), d.numRows))
	} else {
		__antithesis_instrumentation__.Notify(694839)
	}
	__antithesis_instrumentation__.Notify(694837)
	numPerCity := float64(d.numRows) / float64(len(cities))
	cityIdx := int(float64(rowIdx) / numPerCity)
	return cityIdx
}

func (d cityDistributor) rowsForCity(cityIdx int) (min, max int) {
	__antithesis_instrumentation__.Notify(694840)
	if d.numRows < len(cities) {
		__antithesis_instrumentation__.Notify(694844)
		panic(errors.Errorf(`a minimum of %d rows are required got %d`, len(cities), d.numRows))
	} else {
		__antithesis_instrumentation__.Notify(694845)
	}
	__antithesis_instrumentation__.Notify(694841)
	numPerCity := float64(d.numRows) / float64(len(cities))
	min = int(math.Ceil(float64(cityIdx) * numPerCity))
	max = int(math.Ceil(float64(cityIdx+1) * numPerCity))
	if min >= d.numRows {
		__antithesis_instrumentation__.Notify(694846)
		min = d.numRows
	} else {
		__antithesis_instrumentation__.Notify(694847)
	}
	__antithesis_instrumentation__.Notify(694842)
	if max >= d.numRows {
		__antithesis_instrumentation__.Notify(694848)
		max = d.numRows
	} else {
		__antithesis_instrumentation__.Notify(694849)
	}
	__antithesis_instrumentation__.Notify(694843)
	return min, max
}

func (d cityDistributor) randRowInCity(rng *rand.Rand, cityIdx int) int {
	__antithesis_instrumentation__.Notify(694850)
	min, max := d.rowsForCity(cityIdx)
	return min + rng.Intn(max-min)
}

func (g *movr) movrUsersInitialRow(rowIdx int) []interface{} {
	__antithesis_instrumentation__.Notify(694851)
	rng := rand.New(rand.NewSource(g.seed + uint64(rowIdx)))
	cityIdx := g.users.cityForRow(rowIdx)
	city := cities[cityIdx]

	var id uuid.UUID
	id.DeterministicV4(uint64(rowIdx), uint64(g.users.numRows))

	return []interface{}{
		id.String(),
		city.city,
		g.faker.Name(rng),
		g.faker.StreetAddress(rng),
		randCreditCard(rng),
	}
}

func (g *movr) movrVehiclesInitialRow(rowIdx int) []interface{} {
	__antithesis_instrumentation__.Notify(694852)
	rng := rand.New(rand.NewSource(g.seed + uint64(rowIdx)))
	cityIdx := g.vehicles.cityForRow(rowIdx)
	city := cities[cityIdx]

	var id uuid.UUID
	id.DeterministicV4(uint64(rowIdx), uint64(g.vehicles.numRows))

	vehicleType := randVehicleType(rng)
	ownerRowIdx := g.users.randRowInCity(rng, cityIdx)
	ownerID := g.movrUsersInitialRow(ownerRowIdx)[0]

	return []interface{}{
		id.String(),
		city.city,
		vehicleType,
		ownerID,
		g.creationTime.Format(timestampFormat),
		randVehicleStatus(rng),
		g.faker.StreetAddress(rng),
		randVehicleMetadata(rng, vehicleType),
	}
}

func (g *movr) movrRidesInitialRow(rowIdx int) []interface{} {
	__antithesis_instrumentation__.Notify(694853)
	rng := rand.New(rand.NewSource(g.seed + uint64(rowIdx)))
	cityIdx := g.rides.cityForRow(rowIdx)
	city := cities[cityIdx]

	var id uuid.UUID
	id.DeterministicV4(uint64(rowIdx), uint64(g.rides.numRows))

	riderRowIdx := g.users.randRowInCity(rng, cityIdx)
	riderID := g.movrUsersInitialRow(riderRowIdx)[0]
	vehicleRowIdx := g.vehicles.randRowInCity(rng, cityIdx)
	vehicleID := g.movrVehiclesInitialRow(vehicleRowIdx)[0]
	startTime := g.creationTime.Add(-time.Duration(rng.Intn(30)) * 24 * time.Hour)
	endTime := startTime.Add(time.Duration(rng.Intn(60)) * time.Hour)

	return []interface{}{
		id.String(),
		city.city,
		city.city,
		riderID,
		vehicleID,
		g.faker.StreetAddress(rng),
		g.faker.StreetAddress(rng),
		startTime.Format(timestampFormat),
		endTime.Format(timestampFormat),
		rng.Intn(100),
	}
}

func (g *movr) movrVehicleLocationHistoriesInitialRow(rowIdx int) []interface{} {
	__antithesis_instrumentation__.Notify(694854)
	rng := rand.New(rand.NewSource(g.seed + uint64(rowIdx)))
	cityIdx := g.histories.cityForRow(rowIdx)
	city := cities[cityIdx]

	rideRowIdx := g.rides.randRowInCity(rng, cityIdx)
	rideID := g.movrRidesInitialRow(rideRowIdx)[0]
	time := g.creationTime.Add(time.Duration(rowIdx) * time.Millisecond)
	lat, long := randLatLong(rng)

	return []interface{}{
		city.city,
		rideID,
		time.Format(timestampFormat),
		lat,
		long,
	}
}

func (g *movr) movrPromoCodesInitialRow(rowIdx int) []interface{} {
	__antithesis_instrumentation__.Notify(694855)
	rng := rand.New(rand.NewSource(g.seed + uint64(rowIdx)))
	code := strings.ToLower(strings.Join(g.faker.Words(rng, 3), `_`))
	code = fmt.Sprintf("%d_%s", rowIdx, code)
	description := g.faker.Paragraph(rng)
	expirationTime := g.creationTime.Add(time.Duration(rng.Intn(30)) * 24 * time.Hour)

	creationTime := expirationTime.Add(-time.Duration(rng.Intn(30)) * 24 * time.Hour)
	const rulesJSON = `{"type": "percent_discount", "value": "10%"}`

	return []interface{}{
		code,
		description,
		creationTime,
		expirationTime,
		rulesJSON,
	}
}

func (g *movr) movrUserPromoCodesInitialRow(rowIdx int) []interface{} {
	__antithesis_instrumentation__.Notify(694856)
	rng := rand.New(rand.NewSource(g.seed + uint64(rowIdx)))

	var id uuid.UUID
	id.DeterministicV4(uint64(rowIdx), uint64(g.users.numRows))
	cityIdx := g.users.cityForRow(rowIdx)
	city := cities[cityIdx]
	code := strings.ToLower(strings.Join(g.faker.Words(rng, 3), `_`))
	code = fmt.Sprintf("%d_%s", rowIdx, code)
	time := g.creationTime.Add(time.Duration(rowIdx) * time.Millisecond)
	return []interface{}{
		city.city,
		id.String(),
		code,
		time.Format(timestampFormat),
		rng.Intn(20),
	}
}
