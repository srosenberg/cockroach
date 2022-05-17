package rand

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	gosql "database/sql"
	"database/sql/driver"
	"fmt"
	"math/rand"
	"reflect"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/randgen"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"github.com/cockroachdb/cockroach/pkg/workload/histogram"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq"
	"github.com/lib/pq/oid"
	"github.com/spf13/pflag"
)

type random struct {
	flags     workload.Flags
	connFlags *workload.ConnFlags

	batchSize int

	seed int64

	tableName string

	tables     int
	method     string
	primaryKey string
	nullPct    int
}

func init() {
	workload.Register(randMeta)
}

var randMeta = workload.Meta{
	Name:        `rand`,
	Description: `random writes to table`,
	Version:     `1.0.0`,
	New: func() workload.Generator {
		__antithesis_instrumentation__.Notify(695576)
		g := &random{}
		g.flags.FlagSet = pflag.NewFlagSet(`rand`, pflag.ContinueOnError)
		g.flags.Meta = map[string]workload.FlagMeta{
			`batch`: {RuntimeOnly: true},
		}
		g.flags.IntVar(&g.tables, `tables`, 1, `Number of tables to create`)
		g.flags.StringVar(&g.tableName, `table`, ``, `Table to write to`)
		g.flags.IntVar(&g.batchSize, `batch`, 1, `Number of rows to insert in a single SQL statement`)
		g.flags.StringVar(&g.method, `method`, `upsert`, `Choice of DML name: insert, upsert, ioc-update (insert on conflict update), ioc-nothing (insert on conflict no nothing)`)
		g.flags.Int64Var(&g.seed, `seed`, 1, `Key hash seed.`)
		g.flags.StringVar(&g.primaryKey, `primary-key`, ``, `ioc-update and ioc-nothing require primary key`)
		g.flags.IntVar(&g.nullPct, `null-percent`, 5, `Percent random nulls`)
		g.connFlags = workload.NewConnFlags(&g.flags)
		return g
	},
}

func (*random) Meta() workload.Meta { __antithesis_instrumentation__.Notify(695577); return randMeta }

func (w *random) Flags() workload.Flags {
	__antithesis_instrumentation__.Notify(695578)
	return w.flags
}

func (w *random) Hooks() workload.Hooks {
	__antithesis_instrumentation__.Notify(695579)
	return workload.Hooks{}
}

func (w *random) Tables() []workload.Table {
	__antithesis_instrumentation__.Notify(695580)
	tables := make([]workload.Table, w.tables)
	rng := rand.New(rand.NewSource(w.seed))
	for i := 0; i < w.tables; i++ {
		__antithesis_instrumentation__.Notify(695582)
		createTable := randgen.RandCreateTable(rng, "table", rng.Int())
		ctx := tree.NewFmtCtx(tree.FmtParsable)
		createTable.FormatBody(ctx)
		tables[i] = workload.Table{
			Name:   createTable.Table.String(),
			Schema: ctx.CloseAndGetString(),
		}
	}
	__antithesis_instrumentation__.Notify(695581)
	return tables
}

type col struct {
	name          string
	dataType      *types.T
	dataPrecision int
	dataScale     int
	cdefault      gosql.NullString
	isNullable    bool
	isComputed    bool
}

func (w *random) Ops(
	ctx context.Context, urls []string, reg *histogram.Registry,
) (ql workload.QueryLoad, retErr error) {
	__antithesis_instrumentation__.Notify(695583)
	sqlDatabase, err := workload.SanitizeUrls(w, w.connFlags.DBOverride, urls)
	if err != nil {
		__antithesis_instrumentation__.Notify(695601)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(695602)
	}
	__antithesis_instrumentation__.Notify(695584)
	db, err := gosql.Open(`cockroach`, strings.Join(urls, ` `))
	if err != nil {
		__antithesis_instrumentation__.Notify(695603)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(695604)
	}
	__antithesis_instrumentation__.Notify(695585)

	db.SetMaxOpenConns(w.connFlags.Concurrency + 1)
	db.SetMaxIdleConns(w.connFlags.Concurrency + 1)

	tableName := w.tableName
	if tableName == "" {
		__antithesis_instrumentation__.Notify(695605)
		tableName = w.Tables()[0].Name
	} else {
		__antithesis_instrumentation__.Notify(695606)
	}
	__antithesis_instrumentation__.Notify(695586)

	var relid int
	if err := db.QueryRow(fmt.Sprintf("SELECT '%s'::REGCLASS::OID", tableName)).Scan(&relid); err != nil {
		__antithesis_instrumentation__.Notify(695607)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(695608)
	}
	__antithesis_instrumentation__.Notify(695587)

	rows, err := db.Query(
		`
SELECT attname, atttypid, adsrc, NOT attnotnull, attgenerated != ''
FROM pg_catalog.pg_attribute
LEFT JOIN pg_catalog.pg_attrdef
ON attrelid=adrelid AND attnum=adnum
WHERE attrelid=$1`, relid)
	if err != nil {
		__antithesis_instrumentation__.Notify(695609)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(695610)
	}
	__antithesis_instrumentation__.Notify(695588)
	defer func() {
		__antithesis_instrumentation__.Notify(695611)
		retErr = errors.CombineErrors(retErr, rows.Close())
	}()
	__antithesis_instrumentation__.Notify(695589)
	var cols []col
	var numCols = 0

	for rows.Next() {
		__antithesis_instrumentation__.Notify(695612)
		var c col
		c.dataPrecision = 0
		c.dataScale = 0

		var typOid int
		if err := rows.Scan(&c.name, &typOid, &c.cdefault, &c.isNullable, &c.isComputed); err != nil {
			__antithesis_instrumentation__.Notify(695616)
			return workload.QueryLoad{}, err
		} else {
			__antithesis_instrumentation__.Notify(695617)
		}
		__antithesis_instrumentation__.Notify(695613)
		datumType := types.OidToType[oid.Oid(typOid)]
		c.dataType = datumType
		if c.cdefault.String == "unique_rowid()" {
			__antithesis_instrumentation__.Notify(695618)
			continue
		} else {
			__antithesis_instrumentation__.Notify(695619)
		}
		__antithesis_instrumentation__.Notify(695614)
		if strings.HasPrefix(c.cdefault.String, "uuid_v4()") {
			__antithesis_instrumentation__.Notify(695620)
			continue
		} else {
			__antithesis_instrumentation__.Notify(695621)
		}
		__antithesis_instrumentation__.Notify(695615)
		cols = append(cols, c)
		numCols++
	}
	__antithesis_instrumentation__.Notify(695590)

	if numCols == 0 {
		__antithesis_instrumentation__.Notify(695622)
		return workload.QueryLoad{}, errors.New("no columns detected")
	} else {
		__antithesis_instrumentation__.Notify(695623)
	}
	__antithesis_instrumentation__.Notify(695591)

	if err = rows.Err(); err != nil {
		__antithesis_instrumentation__.Notify(695624)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(695625)
	}
	__antithesis_instrumentation__.Notify(695592)

	if strings.HasPrefix(w.method, "ioc") && func() bool {
		__antithesis_instrumentation__.Notify(695626)
		return w.primaryKey == "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(695627)
		rows, err := db.Query(
			`
SELECT a.attname
FROM   pg_index i
JOIN   pg_attribute a ON a.attrelid = i.indrelid
                      AND a.attnum = ANY(i.indkey)
WHERE  i.indrelid = $1
AND    i.indisprimary`, relid)
		if err != nil {
			__antithesis_instrumentation__.Notify(695631)
			return workload.QueryLoad{}, err
		} else {
			__antithesis_instrumentation__.Notify(695632)
		}
		__antithesis_instrumentation__.Notify(695628)
		defer func() {
			__antithesis_instrumentation__.Notify(695633)
			retErr = errors.CombineErrors(retErr, rows.Close())
		}()
		__antithesis_instrumentation__.Notify(695629)
		for rows.Next() {
			__antithesis_instrumentation__.Notify(695634)
			var colname string

			if err := rows.Scan(&colname); err != nil {
				__antithesis_instrumentation__.Notify(695636)
				return workload.QueryLoad{}, err
			} else {
				__antithesis_instrumentation__.Notify(695637)
			}
			__antithesis_instrumentation__.Notify(695635)
			if w.primaryKey != "" {
				__antithesis_instrumentation__.Notify(695638)
				w.primaryKey += "," + colname
			} else {
				__antithesis_instrumentation__.Notify(695639)
				w.primaryKey += colname
			}
		}
		__antithesis_instrumentation__.Notify(695630)
		if err = rows.Err(); err != nil {
			__antithesis_instrumentation__.Notify(695640)
			return workload.QueryLoad{}, err
		} else {
			__antithesis_instrumentation__.Notify(695641)
		}
	} else {
		__antithesis_instrumentation__.Notify(695642)
	}
	__antithesis_instrumentation__.Notify(695593)

	if strings.HasPrefix(w.method, "ioc") && func() bool {
		__antithesis_instrumentation__.Notify(695643)
		return w.primaryKey == "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(695644)
		err := errors.New(
			"insert on conflict requires primary key to be specified via -primary if the table does " +
				"not have primary key")
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(695645)
	}
	__antithesis_instrumentation__.Notify(695594)

	var dmlMethod string
	var dmlSuffix bytes.Buffer
	var buf bytes.Buffer
	switch w.method {
	case "insert":
		__antithesis_instrumentation__.Notify(695646)
		dmlMethod = "insert"
		dmlSuffix.WriteString("")
	case "upsert":
		__antithesis_instrumentation__.Notify(695647)
		dmlMethod = "upsert"
		dmlSuffix.WriteString("")
	case "ioc-nothing":
		__antithesis_instrumentation__.Notify(695648)
		dmlMethod = "insert"
		dmlSuffix.WriteString(fmt.Sprintf(" on conflict (%s) do nothing", w.primaryKey))
	case "ioc-update":
		__antithesis_instrumentation__.Notify(695649)
		dmlMethod = "insert"
		dmlSuffix.WriteString(fmt.Sprintf(" on conflict (%s) do update set ", w.primaryKey))
		for i, c := range cols {
			__antithesis_instrumentation__.Notify(695651)
			if i > 0 {
				__antithesis_instrumentation__.Notify(695653)
				dmlSuffix.WriteString(",")
			} else {
				__antithesis_instrumentation__.Notify(695654)
			}
			__antithesis_instrumentation__.Notify(695652)
			dmlSuffix.WriteString(fmt.Sprintf("%s=EXCLUDED.%s", c.name, c.name))
		}
	default:
		__antithesis_instrumentation__.Notify(695650)
		return workload.QueryLoad{}, errors.Errorf("%s DML method not valid", w.primaryKey)
	}
	__antithesis_instrumentation__.Notify(695595)

	var nonComputedCols []col
	for _, c := range cols {
		__antithesis_instrumentation__.Notify(695655)
		if !c.isComputed {
			__antithesis_instrumentation__.Notify(695656)
			nonComputedCols = append(nonComputedCols, c)
		} else {
			__antithesis_instrumentation__.Notify(695657)
		}
	}
	__antithesis_instrumentation__.Notify(695596)

	fmt.Fprintf(&buf, `%s INTO %s.%s (`, dmlMethod, sqlDatabase, tableName)
	for i, c := range nonComputedCols {
		__antithesis_instrumentation__.Notify(695658)
		if i > 0 {
			__antithesis_instrumentation__.Notify(695660)
			buf.WriteString(",")
		} else {
			__antithesis_instrumentation__.Notify(695661)
		}
		__antithesis_instrumentation__.Notify(695659)
		buf.WriteString(c.name)
	}
	__antithesis_instrumentation__.Notify(695597)
	buf.WriteString(`) VALUES `)

	nCols := len(nonComputedCols)
	for i := 0; i < w.batchSize; i++ {
		__antithesis_instrumentation__.Notify(695662)
		if i > 0 {
			__antithesis_instrumentation__.Notify(695665)
			buf.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(695666)
		}
		__antithesis_instrumentation__.Notify(695663)
		buf.WriteString("(")
		for j := range nonComputedCols {
			__antithesis_instrumentation__.Notify(695667)
			if j > 0 {
				__antithesis_instrumentation__.Notify(695669)
				buf.WriteString(", ")
			} else {
				__antithesis_instrumentation__.Notify(695670)
			}
			__antithesis_instrumentation__.Notify(695668)
			fmt.Fprintf(&buf, `$%d`, 1+j+(nCols*i))
		}
		__antithesis_instrumentation__.Notify(695664)
		buf.WriteString(")")
	}
	__antithesis_instrumentation__.Notify(695598)

	buf.WriteString(dmlSuffix.String())

	writeStmt, err := db.Prepare(buf.String())
	if err != nil {
		__antithesis_instrumentation__.Notify(695671)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(695672)
	}
	__antithesis_instrumentation__.Notify(695599)

	ql = workload.QueryLoad{SQLDatabase: sqlDatabase}

	for i := 0; i < w.connFlags.Concurrency; i++ {
		__antithesis_instrumentation__.Notify(695673)
		op := randOp{
			config:    w,
			hists:     reg.GetHandle(),
			db:        db,
			cols:      nonComputedCols,
			rng:       rand.New(rand.NewSource(w.seed + int64(i))),
			writeStmt: writeStmt,
		}
		ql.WorkerFns = append(ql.WorkerFns, op.run)
	}
	__antithesis_instrumentation__.Notify(695600)
	return ql, nil
}

type randOp struct {
	config    *random
	hists     *histogram.Histograms
	db        *gosql.DB
	cols      []col
	rng       *rand.Rand
	writeStmt *gosql.Stmt
}

func DatumToGoSQL(d tree.Datum) (interface{}, error) {
	__antithesis_instrumentation__.Notify(695674)
	d = tree.UnwrapDatum(nil, d)
	if d == tree.DNull {
		__antithesis_instrumentation__.Notify(695677)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(695678)
	}
	__antithesis_instrumentation__.Notify(695675)
	switch d := d.(type) {
	case *tree.DBool:
		__antithesis_instrumentation__.Notify(695679)
		return bool(*d), nil
	case *tree.DString:
		__antithesis_instrumentation__.Notify(695680)
		return string(*d), nil
	case *tree.DBytes:
		__antithesis_instrumentation__.Notify(695681)
		return string(*d), nil
	case *tree.DDate, *tree.DTime:
		__antithesis_instrumentation__.Notify(695682)
		return tree.AsStringWithFlags(d, tree.FmtBareStrings), nil
	case *tree.DTimestamp:
		__antithesis_instrumentation__.Notify(695683)
		return d.Time, nil
	case *tree.DTimestampTZ:
		__antithesis_instrumentation__.Notify(695684)
		return d.Time, nil
	case *tree.DInterval:
		__antithesis_instrumentation__.Notify(695685)
		return d.Duration.String(), nil
	case *tree.DBitArray:
		__antithesis_instrumentation__.Notify(695686)
		return tree.AsStringWithFlags(d, tree.FmtBareStrings), nil
	case *tree.DInt:
		__antithesis_instrumentation__.Notify(695687)
		return int64(*d), nil
	case *tree.DOid:
		__antithesis_instrumentation__.Notify(695688)
		return int(d.DInt), nil
	case *tree.DFloat:
		__antithesis_instrumentation__.Notify(695689)
		return float64(*d), nil
	case *tree.DDecimal:
		__antithesis_instrumentation__.Notify(695690)
		return d.Float64()
	case *tree.DArray:
		__antithesis_instrumentation__.Notify(695691)
		arr := make([]interface{}, len(d.Array))
		for i := range d.Array {
			__antithesis_instrumentation__.Notify(695696)
			elt, err := DatumToGoSQL(d.Array[i])
			if err != nil {
				__antithesis_instrumentation__.Notify(695699)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(695700)
			}
			__antithesis_instrumentation__.Notify(695697)
			if elt == nil {
				__antithesis_instrumentation__.Notify(695701)
				elt = nullVal{}
			} else {
				__antithesis_instrumentation__.Notify(695702)
			}
			__antithesis_instrumentation__.Notify(695698)
			arr[i] = elt
		}
		__antithesis_instrumentation__.Notify(695692)
		return pq.Array(arr), nil
	case *tree.DUuid:
		__antithesis_instrumentation__.Notify(695693)
		return d.UUID, nil
	case *tree.DIPAddr:
		__antithesis_instrumentation__.Notify(695694)
		return d.IPAddr.String(), nil
	case *tree.DJSON:
		__antithesis_instrumentation__.Notify(695695)
		return d.JSON.String(), nil
	}
	__antithesis_instrumentation__.Notify(695676)
	return nil, errors.Errorf("unhandled datum type: %s", reflect.TypeOf(d))
}

type nullVal struct {
}

func (nullVal) Value() (driver.Value, error) {
	__antithesis_instrumentation__.Notify(695703)
	return nil, nil
}

func (o *randOp) run(ctx context.Context) (err error) {
	__antithesis_instrumentation__.Notify(695704)
	params := make([]interface{}, len(o.cols)*o.config.batchSize)
	k := 0
	for j := 0; j < o.config.batchSize; j++ {
		__antithesis_instrumentation__.Notify(695706)
		for _, c := range o.cols {
			__antithesis_instrumentation__.Notify(695707)
			nullPct := 0
			if c.isNullable && func() bool {
				__antithesis_instrumentation__.Notify(695710)
				return o.config.nullPct > 0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(695711)
				nullPct = 100 / o.config.nullPct
			} else {
				__antithesis_instrumentation__.Notify(695712)
			}
			__antithesis_instrumentation__.Notify(695708)
			d := randgen.RandDatumWithNullChance(o.rng, c.dataType, nullPct)
			params[k], err = DatumToGoSQL(d)
			if err != nil {
				__antithesis_instrumentation__.Notify(695713)
				return err
			} else {
				__antithesis_instrumentation__.Notify(695714)
			}
			__antithesis_instrumentation__.Notify(695709)
			k++
		}
	}
	__antithesis_instrumentation__.Notify(695705)
	start := timeutil.Now()
	_, err = o.writeStmt.ExecContext(ctx, params...)
	o.hists.Get(`write`).Record(timeutil.Since(start))
	return err
}
