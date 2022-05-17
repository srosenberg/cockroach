// cmp-sql connects to postgres and cockroach servers and compares the
// results of SQL statements. Statements support both random generation
// and random placeholder values. It can thus be used to do correctness or
// compatibility testing.
//
// To use, start a cockroach and postgres server with SSL disabled. cmp-sql
// will connect to both, generate some random SQL, and print an error when
// difference results are returned. Currently it tests LIKE, binary operators,
// and unary operators. cmp-sql runs a loop: 1) choose an Input, 2) generate a
// random SQL string from from the Input, 3) generate random placeholders, 4)
// execute the SQL + placeholders and compare results.
//
// The Inputs slice determines what SQL is generated. cmp-sql will repeatedly
// generate new kinds of input SQL. The `sql` property of an Input is a
// function that returns a SQL string with possible placeholders. The `args`
// func slice generates the placeholder arguments.
package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"regexp"
	"strings"

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/cockroach/pkg/util/randutil"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

var (
	pgAddr = flag.String("pg", "localhost:5432", "postgres address")
	pgUser = flag.String("pg-user", "postgres", "postgres user")
	crAddr = flag.String("cr", "localhost:26257", "cockroach address")
	crUser = flag.String("cr-user", "root", "cockroach user")
	rng, _ = randutil.NewPseudoRand()
)

func main() {
	__antithesis_instrumentation__.Notify(37877)
	flag.Parse()
	ctx := context.Background()

	var confs []*pgx.ConnConfig
	for addr, user := range map[string]string{
		*pgAddr: *pgUser,
		*crAddr: *crUser,
	} {
		__antithesis_instrumentation__.Notify(37881)
		conf, err := pgx.ParseConfig(fmt.Sprintf("postgresql://%s@%s?sslmode=disable", user, addr))
		if err != nil {
			__antithesis_instrumentation__.Notify(37883)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(37884)
		}
		__antithesis_instrumentation__.Notify(37882)
		confs = append(confs, conf)
	}
	__antithesis_instrumentation__.Notify(37878)
	dbs := make([]*pgx.Conn, len(confs))
	for i, conf := range confs {
		__antithesis_instrumentation__.Notify(37885)
		db, err := pgx.ConnectConfig(ctx, conf)
		if err != nil {
			__antithesis_instrumentation__.Notify(37887)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(37888)
		}
		__antithesis_instrumentation__.Notify(37886)
		dbs[i] = db
	}
	__antithesis_instrumentation__.Notify(37879)

	seen := make(map[string]bool)
	sawErr := func(err error) string {
		__antithesis_instrumentation__.Notify(37889)
		s := reduceErr(err)

		res := fmt.Sprintf("ERR: %s", s)
		if !seen[res] {
			__antithesis_instrumentation__.Notify(37891)
			fmt.Print(err, "\n\n")
			seen[res] = true
		} else {
			__antithesis_instrumentation__.Notify(37892)
		}
		__antithesis_instrumentation__.Notify(37890)
		return res
	}
	__antithesis_instrumentation__.Notify(37880)
	for {
		__antithesis_instrumentation__.Notify(37893)
	Loop:
		for _, input := range Inputs {
			__antithesis_instrumentation__.Notify(37894)
			results := map[int]string{}
			sql, args, repro := input.Generate()
			for i, db := range dbs {
				__antithesis_instrumentation__.Notify(37895)
				var res, full string
				if rows, err := db.Query(ctx, sql, args...); err != nil {
					__antithesis_instrumentation__.Notify(37899)
					res = sawErr(err)
					full = err.Error()
				} else {
					__antithesis_instrumentation__.Notify(37900)
					if rows.Next() {
						__antithesis_instrumentation__.Notify(37903)
						vals, err := rows.Values()
						if err != nil {
							__antithesis_instrumentation__.Notify(37904)
							panic(err)
						} else {
							__antithesis_instrumentation__.Notify(37905)
							if len(vals) != 1 {
								__antithesis_instrumentation__.Notify(37906)
								panic(fmt.Errorf("expected 1 val, got %v", vals))
							} else {
								__antithesis_instrumentation__.Notify(37907)
								switch v := vals[0].(type) {
								case *pgtype.Numeric:
									__antithesis_instrumentation__.Notify(37909)
									b, err := v.EncodeText(nil, nil)
									if err != nil {
										__antithesis_instrumentation__.Notify(37913)
										panic(err)
									} else {
										__antithesis_instrumentation__.Notify(37914)
									}
									__antithesis_instrumentation__.Notify(37910)

									var d apd.Decimal
									if _, _, err := d.SetString(string(b)); err != nil {
										__antithesis_instrumentation__.Notify(37915)
										panic(err)
									} else {
										__antithesis_instrumentation__.Notify(37916)
									}
									__antithesis_instrumentation__.Notify(37911)
									d.Reduce(&d)
									res = d.String()
								default:
									__antithesis_instrumentation__.Notify(37912)
									res = fmt.Sprint(v)
								}
								__antithesis_instrumentation__.Notify(37908)
								full = res
							}
						}
					} else {
						__antithesis_instrumentation__.Notify(37917)
					}
					__antithesis_instrumentation__.Notify(37901)
					rows.Close()
					if err := rows.Err(); err != nil {
						__antithesis_instrumentation__.Notify(37918)
						res = sawErr(err)
						full = err.Error()
					} else {
						__antithesis_instrumentation__.Notify(37919)
					}
					__antithesis_instrumentation__.Notify(37902)
					if res == "" {
						__antithesis_instrumentation__.Notify(37920)
						panic("empty")
					} else {
						__antithesis_instrumentation__.Notify(37921)
					}
				}
				__antithesis_instrumentation__.Notify(37896)

				if err := db.Ping(ctx); err != nil {
					__antithesis_instrumentation__.Notify(37922)
					fmt.Print("CRASHER:\n", repro)
					panic(fmt.Errorf("%v is down", confs[i].Port))
				} else {
					__antithesis_instrumentation__.Notify(37923)
				}
				__antithesis_instrumentation__.Notify(37897)

				for vi, v := range results {
					__antithesis_instrumentation__.Notify(37924)
					if verr, reserr := strings.HasPrefix(v, "ERR"), strings.HasPrefix(res, "ERR"); verr && func() bool {
						__antithesis_instrumentation__.Notify(37926)
						return reserr == true
					}() == true {
						__antithesis_instrumentation__.Notify(37927)
						continue
					} else {
						__antithesis_instrumentation__.Notify(37928)
						if input.ignoreIfEitherError && func() bool {
							__antithesis_instrumentation__.Notify(37929)
							return (verr || func() bool {
								__antithesis_instrumentation__.Notify(37930)
								return reserr == true
							}() == true) == true
						}() == true {
							__antithesis_instrumentation__.Notify(37931)
							continue
						} else {
							__antithesis_instrumentation__.Notify(37932)
						}
					}
					__antithesis_instrumentation__.Notify(37925)
					if v != res {
						__antithesis_instrumentation__.Notify(37933)
						mismatch := fmt.Sprintf("%v: got %s\n%v: saw %s\n",
							confs[i].Port,
							full,
							confs[vi].Port,
							v,
						)
						if !seen[mismatch] {
							__antithesis_instrumentation__.Notify(37935)
							seen[mismatch] = true
							fmt.Print("MISMATCH:\n", mismatch)
							fmt.Println(repro)
						} else {
							__antithesis_instrumentation__.Notify(37936)
						}
						__antithesis_instrumentation__.Notify(37934)
						continue Loop
					} else {
						__antithesis_instrumentation__.Notify(37937)
					}
				}
				__antithesis_instrumentation__.Notify(37898)
				results[i] = res
			}
		}
	}
}

var reduceErrRE = regexp.MustCompile(` *(ERROR)?[ :]*([A-Za-z ]+?) +`)

func reduceErr(err error) string {
	__antithesis_instrumentation__.Notify(37938)
	match := reduceErrRE.FindStringSubmatch(err.Error())
	if match == nil {
		__antithesis_instrumentation__.Notify(37940)
		return err.Error()
	} else {
		__antithesis_instrumentation__.Notify(37941)
	}
	__antithesis_instrumentation__.Notify(37939)
	return match[2]
}

type Input struct {
	sql  func() string
	args []func() interface{}

	ignoreIfEitherError bool
}

func (i Input) Generate() (sql string, args []interface{}, repro string) {
	__antithesis_instrumentation__.Notify(37942)
	sql = i.sql()
	args = make([]interface{}, len(i.args))
	for i, fn := range i.args {
		__antithesis_instrumentation__.Notify(37945)
		args[i] = fn()
	}
	__antithesis_instrumentation__.Notify(37943)
	var b strings.Builder
	fmt.Fprintf(&b, "PREPARE a AS %s;\n", sql)
	b.WriteString("EXECUTE a (")
	for i, fn := range i.args {
		__antithesis_instrumentation__.Notify(37946)
		if i > 0 {
			__antithesis_instrumentation__.Notify(37948)
			b.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(37949)
		}
		__antithesis_instrumentation__.Notify(37947)
		arg := fn()
		switch arg := arg.(type) {
		case int, int64, float64:
			__antithesis_instrumentation__.Notify(37950)
			fmt.Fprint(&b, arg)
		case string:
			__antithesis_instrumentation__.Notify(37951)
			s := fmt.Sprintf("%q", arg)
			fmt.Fprintf(&b, "e'%s'", s[1:len(s)-1])
		default:
			__antithesis_instrumentation__.Notify(37952)
			panic(fmt.Errorf("unknown type: %T", arg))
		}
	}
	__antithesis_instrumentation__.Notify(37944)
	b.WriteString(");\n")
	return sql, args, b.String()
}

var Inputs = []Input{
	{
		sql:  pass("SELECT $1 LIKE $2"),
		args: twoLike,
	},
	{
		sql:  pass("SELECT $1 LIKE $2 ESCAPE $3"),
		args: threeLike,
	},
	{
		sql: fromSlices(
			"SELECT $1::%s %s $2::%s %s $3::%s",
			numTyps,
			binaryNumOps,
			numTyps,
			binaryNumOps,
			numTyps,
		),
		args:                threeNum,
		ignoreIfEitherError: true,
	},
	{
		sql: fromSlices(
			"SELECT %s($1::%s)",
			unaryNumOps,
			numTyps,
		),
		args: oneNum,
	},
}

var (
	twoLike   = []func() interface{}{likeArg(5), likeArg(5)}
	threeLike = []func() interface{}{likeArg(5), likeArg(5), likeArg(3)}
	oneNum    = []func() interface{}{num}
	threeNum  = []func() interface{}{num, num, num}

	binaryNumOps = []string{
		"-",
		"+",
		"^",
		"*",
		"/",
		"//",
		"%",
		"<<",
		">>",
		"&",
		"#",
		"|",
	}
	unaryNumOps = []string{
		"-",
		"~",
	}
	numTyps = []string{
		"int8",
		"float8",
		"decimal",
	}
)

func pass(s string) func() string {
	__antithesis_instrumentation__.Notify(37953)
	return func() string {
		__antithesis_instrumentation__.Notify(37954)
		return s
	}
}

func fromSlices(s string, args ...[]string) func() string {
	__antithesis_instrumentation__.Notify(37955)
	return func() string {
		__antithesis_instrumentation__.Notify(37956)
		gen := make([]interface{}, len(args))
		for i, arg := range args {
			__antithesis_instrumentation__.Notify(37958)
			gen[i] = arg[rand.Intn(len(arg))]
		}
		__antithesis_instrumentation__.Notify(37957)
		return fmt.Sprintf(s, gen...)
	}
}

func num() interface{} {
	__antithesis_instrumentation__.Notify(37959)
	switch rand.Intn(6) {
	case 1:
		__antithesis_instrumentation__.Notify(37960)
		return 1
	case 2:
		__antithesis_instrumentation__.Notify(37961)
		return 2
	case 3:
		__antithesis_instrumentation__.Notify(37962)
		return -1
	case 4:
		__antithesis_instrumentation__.Notify(37963)
		return rand.Int() / (rand.Intn(10) + 1)
	case 5:
		__antithesis_instrumentation__.Notify(37964)
		return rand.NormFloat64()
	default:
		__antithesis_instrumentation__.Notify(37965)
		return 0
	}
}

func likeArg(n int) func() interface{} {
	__antithesis_instrumentation__.Notify(37966)
	return func() interface{} {
		__antithesis_instrumentation__.Notify(37967)
		p := make([]byte, rng.Intn(n))
		for i := range p {
			__antithesis_instrumentation__.Notify(37969)
			switch rand.Intn(4) {
			case 0:
				__antithesis_instrumentation__.Notify(37970)
				p[i] = '_'
			case 1:
				__antithesis_instrumentation__.Notify(37971)
				p[i] = '%'
			default:
				__antithesis_instrumentation__.Notify(37972)
				p[i] = byte(1 + rng.Intn(127))
			}
		}
		__antithesis_instrumentation__.Notify(37968)
		return string(p)
	}
}
