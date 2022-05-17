package roachmartccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	gosql "database/sql"
	"fmt"
	"math/rand"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/util/randutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"github.com/cockroachdb/cockroach/pkg/workload/histogram"
	"github.com/cockroachdb/errors"
	"github.com/spf13/pflag"
)

var zones = []string{"us-central1-b", "us-west1-b", "europe-west2-b"}

var usersSchema = func() string {
	__antithesis_instrumentation__.Notify(27842)
	var buf bytes.Buffer
	buf.WriteString(`(
	zone STRING,
	email STRING,
	address STRING,
	PRIMARY KEY (zone, email)
) PARTITION BY LIST (zone) (`)
	for i, z := range zones {
		__antithesis_instrumentation__.Notify(27844)
		if i > 0 {
			__antithesis_instrumentation__.Notify(27846)
			buf.WriteByte(',')
		} else {
			__antithesis_instrumentation__.Notify(27847)
		}
		__antithesis_instrumentation__.Notify(27845)
		buf.WriteString(fmt.Sprintf("\n\tPARTITION %[1]q VALUES IN ('%[1]s')", z))
	}
	__antithesis_instrumentation__.Notify(27843)
	buf.WriteString("\n)")
	return buf.String()
}()

const (
	ordersSchema = `(
		user_zone STRING,
		user_email STRING,
		id INT,
		fulfilled BOOL,
		PRIMARY KEY (user_zone, user_email, id),
		FOREIGN KEY (user_zone, user_email) REFERENCES users
	)`

	defaultUsers  = 10000
	defaultOrders = 100000

	zoneLocationsStmt = `UPSERT INTO system.locations VALUES
		('zone', 'us-east1-b', 33.0641249, -80.0433347),
		('zone', 'us-west1-b', 45.6319052, -121.2010282),
		('zone', 'europe-west2-b', 51.509865, 0)
	`
)

type roachmart struct {
	flags     workload.Flags
	connFlags *workload.ConnFlags

	seed          int64
	partition     bool
	localZone     string
	localPercent  int
	users, orders int
}

func init() {
	workload.Register(roachmartMeta)
}

var roachmartMeta = workload.Meta{
	Name:        `roachmart`,
	Description: `Roachmart models a geo-distributed online storefront with users and orders`,
	Version:     `1.0.0`,
	New: func() workload.Generator {
		__antithesis_instrumentation__.Notify(27848)
		g := &roachmart{}
		g.flags.FlagSet = pflag.NewFlagSet(`roachmart`, pflag.ContinueOnError)
		g.flags.Meta = map[string]workload.FlagMeta{
			`local-zone`:    {RuntimeOnly: true},
			`local-percent`: {RuntimeOnly: true},
		}
		g.flags.Int64Var(&g.seed, `seed`, 1, `Key hash seed.`)
		g.flags.BoolVar(&g.partition, `partition`, true, `Whether to apply zone configs to the partitions of the users table.`)
		g.flags.StringVar(&g.localZone, `local-zone`, ``, `The zone in which this load generator is running.`)
		g.flags.IntVar(&g.localPercent, `local-percent`, 50, `Percent (0-100) of operations that operate on local data.`)
		g.flags.IntVar(&g.users, `users`, defaultUsers, `Initial number of accounts in users table.`)
		g.flags.IntVar(&g.orders, `orders`, defaultOrders, `Initial number of orders in orders table.`)
		g.connFlags = workload.NewConnFlags(&g.flags)
		return g
	},
}

func (m *roachmart) Meta() workload.Meta {
	__antithesis_instrumentation__.Notify(27849)
	return roachmartMeta
}

func (m *roachmart) Flags() workload.Flags {
	__antithesis_instrumentation__.Notify(27850)
	return m.flags
}

func (m *roachmart) Hooks() workload.Hooks {
	__antithesis_instrumentation__.Notify(27851)
	return workload.Hooks{
		Validate: func() error {
			__antithesis_instrumentation__.Notify(27852)
			if m.localZone == "" {
				__antithesis_instrumentation__.Notify(27856)
				return errors.New("local zone must be specified")
			} else {
				__antithesis_instrumentation__.Notify(27857)
			}
			__antithesis_instrumentation__.Notify(27853)
			found := false
			for _, z := range zones {
				__antithesis_instrumentation__.Notify(27858)
				if z == m.localZone {
					__antithesis_instrumentation__.Notify(27859)
					found = true
					break
				} else {
					__antithesis_instrumentation__.Notify(27860)
				}
			}
			__antithesis_instrumentation__.Notify(27854)
			if !found {
				__antithesis_instrumentation__.Notify(27861)
				return fmt.Errorf("unknown zone %q (options: %s)", m.localZone, zones)
			} else {
				__antithesis_instrumentation__.Notify(27862)
			}
			__antithesis_instrumentation__.Notify(27855)
			return nil
		},

		PreLoad: func(db *gosql.DB) error {
			__antithesis_instrumentation__.Notify(27863)
			if _, err := db.Exec(zoneLocationsStmt); err != nil {
				__antithesis_instrumentation__.Notify(27867)
				return err
			} else {
				__antithesis_instrumentation__.Notify(27868)
			}
			__antithesis_instrumentation__.Notify(27864)
			if !m.partition {
				__antithesis_instrumentation__.Notify(27869)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(27870)
			}
			__antithesis_instrumentation__.Notify(27865)
			for _, z := range zones {
				__antithesis_instrumentation__.Notify(27871)

				makeStmt := func(s string) string {
					__antithesis_instrumentation__.Notify(27874)
					return fmt.Sprintf(s, fmt.Sprintf("%q", z), fmt.Sprintf("'constraints: [+zone=%s]'", z))
				}
				__antithesis_instrumentation__.Notify(27872)
				stmt := makeStmt("ALTER PARTITION %[1]s OF TABLE users CONFIGURE ZONE = %[2]s")
				_, err := db.Exec(stmt)
				if err != nil && func() bool {
					__antithesis_instrumentation__.Notify(27875)
					return strings.Contains(err.Error(), "syntax error") == true
				}() == true {
					__antithesis_instrumentation__.Notify(27876)
					stmt = makeStmt("ALTER PARTITION %[1]s OF TABLE users EXPERIMENTAL CONFIGURE ZONE %[2]s")
					_, err = db.Exec(stmt)
				} else {
					__antithesis_instrumentation__.Notify(27877)
				}
				__antithesis_instrumentation__.Notify(27873)
				if err != nil {
					__antithesis_instrumentation__.Notify(27878)
					return err
				} else {
					__antithesis_instrumentation__.Notify(27879)
				}
			}
			__antithesis_instrumentation__.Notify(27866)
			return nil
		},
	}
}

func (m *roachmart) Tables() []workload.Table {
	__antithesis_instrumentation__.Notify(27880)
	users := workload.Table{
		Name:   `users`,
		Schema: usersSchema,
		InitialRows: workload.Tuples(
			m.users,
			func(rowIdx int) []interface{} {
				__antithesis_instrumentation__.Notify(27883)
				rng := rand.New(rand.NewSource(m.seed + int64(rowIdx)))
				const emailTemplate = `user-%d@roachmart.example`
				return []interface{}{
					zones[rowIdx%3],
					fmt.Sprintf(emailTemplate, rowIdx),
					string(randutil.RandBytes(rng, 64)),
				}
			},
		),
	}
	__antithesis_instrumentation__.Notify(27881)
	orders := workload.Table{
		Name:   `orders`,
		Schema: ordersSchema,
		InitialRows: workload.Tuples(
			m.orders,
			func(rowIdx int) []interface{} {
				__antithesis_instrumentation__.Notify(27884)
				user := users.InitialRows.BatchRows(rowIdx % m.users)[0]
				zone, email := user[0], user[1]
				return []interface{}{
					zone,
					email,
					rowIdx,
					[]string{`f`, `t`}[rowIdx%2],
				}
			},
		),
	}
	__antithesis_instrumentation__.Notify(27882)
	return []workload.Table{users, orders}
}

func (m *roachmart) Ops(
	ctx context.Context, urls []string, reg *histogram.Registry,
) (workload.QueryLoad, error) {
	__antithesis_instrumentation__.Notify(27885)
	sqlDatabase, err := workload.SanitizeUrls(m, m.connFlags.DBOverride, urls)
	if err != nil {
		__antithesis_instrumentation__.Notify(27889)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(27890)
	}
	__antithesis_instrumentation__.Notify(27886)
	db, err := gosql.Open(`cockroach`, strings.Join(urls, ` `))
	if err != nil {
		__antithesis_instrumentation__.Notify(27891)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(27892)
	}
	__antithesis_instrumentation__.Notify(27887)

	db.SetMaxOpenConns(m.connFlags.Concurrency + 1)
	db.SetMaxIdleConns(m.connFlags.Concurrency + 1)

	ql := workload.QueryLoad{SQLDatabase: sqlDatabase}

	const query = `SELECT * FROM orders WHERE user_zone = $1 AND user_email = $2`
	for i := 0; i < m.connFlags.Concurrency; i++ {
		__antithesis_instrumentation__.Notify(27893)
		rng := rand.New(rand.NewSource(m.seed))
		usersTable := m.Tables()[0]
		hists := reg.GetHandle()
		workerFn := func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(27895)
			wantLocal := rng.Intn(100) < m.localPercent

			var zone, email interface{}
			for i := rng.Int(); ; i++ {
				__antithesis_instrumentation__.Notify(27898)
				user := usersTable.InitialRows.BatchRows(i % m.users)[0]
				zone, email = user[0], user[1]
				userLocal := zone == m.localZone
				if userLocal == wantLocal {
					__antithesis_instrumentation__.Notify(27899)
					break
				} else {
					__antithesis_instrumentation__.Notify(27900)
				}
			}
			__antithesis_instrumentation__.Notify(27896)
			start := timeutil.Now()
			_, err := db.Exec(query, zone, email)
			if wantLocal {
				__antithesis_instrumentation__.Notify(27901)
				hists.Get(`local`).Record(timeutil.Since(start))
			} else {
				__antithesis_instrumentation__.Notify(27902)
				hists.Get(`remote`).Record(timeutil.Since(start))
			}
			__antithesis_instrumentation__.Notify(27897)
			return err
		}
		__antithesis_instrumentation__.Notify(27894)
		ql.WorkerFns = append(ql.WorkerFns, workerFn)
	}
	__antithesis_instrumentation__.Notify(27888)
	return ql, nil
}
