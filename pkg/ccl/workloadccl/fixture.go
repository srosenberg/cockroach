package workloadccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	gosql "database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"github.com/cockroachdb/errors"
	"github.com/spf13/pflag"
)

func init() {
	workload.ImportDataLoader = ImportDataLoader{}
}

type FixtureConfig struct {
	StorageProvider string

	Bucket string

	Basename string

	AuthParams string

	CSVServerURL string

	TableStats bool
}

func (s FixtureConfig) ObjectPathToURI(folders ...string) string {
	__antithesis_instrumentation__.Notify(27609)
	path := filepath.Join(folders...)
	return fmt.Sprintf("%s://%s/%s/%s?%s", s.StorageProvider, s.Bucket, s.Basename, path, s.AuthParams)
}

type Fixture struct {
	Config    FixtureConfig
	Generator workload.Generator
	Tables    []FixtureTable
}

type FixtureTable struct {
	TableName string
	BackupURI string
}

func serializeOptions(gen workload.Generator) string {
	__antithesis_instrumentation__.Notify(27610)
	f, ok := gen.(workload.Flagser)
	if !ok {
		__antithesis_instrumentation__.Notify(27613)
		return ``
	} else {
		__antithesis_instrumentation__.Notify(27614)
	}
	__antithesis_instrumentation__.Notify(27611)

	var buf bytes.Buffer
	flags := f.Flags()
	flags.VisitAll(func(f *pflag.Flag) {
		__antithesis_instrumentation__.Notify(27615)
		if flags.Meta != nil && func() bool {
			__antithesis_instrumentation__.Notify(27618)
			return flags.Meta[f.Name].RuntimeOnly == true
		}() == true {
			__antithesis_instrumentation__.Notify(27619)
			return
		} else {
			__antithesis_instrumentation__.Notify(27620)
		}
		__antithesis_instrumentation__.Notify(27616)
		if buf.Len() > 0 {
			__antithesis_instrumentation__.Notify(27621)
			buf.WriteString(`,`)
		} else {
			__antithesis_instrumentation__.Notify(27622)
		}
		__antithesis_instrumentation__.Notify(27617)
		fmt.Fprintf(&buf, `%s=%s`, url.PathEscape(f.Name), url.PathEscape(f.Value.String()))
	})
	__antithesis_instrumentation__.Notify(27612)
	return buf.String()
}

func generatorToStorageFolder(gen workload.Generator) string {
	__antithesis_instrumentation__.Notify(27623)
	meta := gen.Meta()
	return filepath.Join(meta.Name,
		fmt.Sprintf(`version=%s,%s`, meta.Version, serializeOptions(gen)))
}

func FixtureURL(config FixtureConfig, gen workload.Generator) string {
	__antithesis_instrumentation__.Notify(27624)
	return config.ObjectPathToURI(generatorToStorageFolder(gen))
}

func GetFixture(
	ctx context.Context, es cloud.ExternalStorage, config FixtureConfig, gen workload.Generator,
) (Fixture, error) {
	__antithesis_instrumentation__.Notify(27625)
	var fixture Fixture
	fixtureFolder := generatorToStorageFolder(gen)
	fixture = Fixture{Config: config, Generator: gen}
	for _, table := range gen.Tables() {
		__antithesis_instrumentation__.Notify(27627)
		tableFolder := filepath.Join(fixtureFolder, table.Name)
		var files []string
		if err := listDir(ctx, es, func(s string) error {
			__antithesis_instrumentation__.Notify(27630)
			files = append(files, s)
			return nil
		}, tableFolder); err != nil {
			__antithesis_instrumentation__.Notify(27631)
			return Fixture{}, err
		} else {
			__antithesis_instrumentation__.Notify(27632)
		}
		__antithesis_instrumentation__.Notify(27628)

		if len(files) == 0 {
			__antithesis_instrumentation__.Notify(27633)
			return Fixture{}, errors.Newf(
				"expected non zero files for table %q in fixture folder %q", table.Name, tableFolder)
		} else {
			__antithesis_instrumentation__.Notify(27634)
		}
		__antithesis_instrumentation__.Notify(27629)

		fixture.Tables = append(fixture.Tables, FixtureTable{
			TableName: table.Name,
			BackupURI: config.ObjectPathToURI(tableFolder),
		})
	}
	__antithesis_instrumentation__.Notify(27626)
	return fixture, nil
}

func csvServerPaths(
	csvServerURL string, gen workload.Generator, table workload.Table, numNodes int,
) []string {
	__antithesis_instrumentation__.Notify(27635)
	if table.InitialRows.FillBatch == nil {
		__antithesis_instrumentation__.Notify(27639)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(27640)
	}
	__antithesis_instrumentation__.Notify(27636)

	numFiles := numNodes
	rowStep := table.InitialRows.NumBatches / numFiles
	if rowStep == 0 {
		__antithesis_instrumentation__.Notify(27641)
		rowStep = 1
	} else {
		__antithesis_instrumentation__.Notify(27642)
	}
	__antithesis_instrumentation__.Notify(27637)

	var paths []string
	for rowIdx := 0; ; {
		__antithesis_instrumentation__.Notify(27643)
		chunkRowStart, chunkRowEnd := rowIdx, rowIdx+rowStep
		if chunkRowEnd > table.InitialRows.NumBatches {
			__antithesis_instrumentation__.Notify(27646)
			chunkRowEnd = table.InitialRows.NumBatches
		} else {
			__antithesis_instrumentation__.Notify(27647)
		}
		__antithesis_instrumentation__.Notify(27644)

		params := url.Values{
			`row-start`: []string{strconv.Itoa(chunkRowStart)},
			`row-end`:   []string{strconv.Itoa(chunkRowEnd)},
			`version`:   []string{gen.Meta().Version},
		}
		if f, ok := gen.(workload.Flagser); ok {
			__antithesis_instrumentation__.Notify(27648)
			flags := f.Flags()
			flags.VisitAll(func(f *pflag.Flag) {
				__antithesis_instrumentation__.Notify(27649)
				if flags.Meta[f.Name].RuntimeOnly {
					__antithesis_instrumentation__.Notify(27651)
					return
				} else {
					__antithesis_instrumentation__.Notify(27652)
				}
				__antithesis_instrumentation__.Notify(27650)
				params[f.Name] = append(params[f.Name], f.Value.String())
			})
		} else {
			__antithesis_instrumentation__.Notify(27653)
		}
		__antithesis_instrumentation__.Notify(27645)
		path := fmt.Sprintf(`%s/csv/%s/%s?%s`,
			csvServerURL, gen.Meta().Name, table.Name, params.Encode())
		paths = append(paths, path)

		rowIdx = chunkRowEnd
		if rowIdx >= table.InitialRows.NumBatches {
			__antithesis_instrumentation__.Notify(27654)
			break
		} else {
			__antithesis_instrumentation__.Notify(27655)
		}
	}
	__antithesis_instrumentation__.Notify(27638)
	return paths
}

const numNodesQuery = `SELECT count(node_id) FROM "".crdb_internal.gossip_liveness`

func MakeFixture(
	ctx context.Context,
	sqlDB *gosql.DB,
	es cloud.ExternalStorage,
	config FixtureConfig,
	gen workload.Generator,
	filesPerNode int,
) (Fixture, error) {
	__antithesis_instrumentation__.Notify(27656)
	for _, t := range gen.Tables() {
		__antithesis_instrumentation__.Notify(27664)
		if t.InitialRows.FillBatch == nil {
			__antithesis_instrumentation__.Notify(27665)
			return Fixture{}, errors.Errorf(
				`make fixture is not supported for workload %s`, gen.Meta().Name,
			)
		} else {
			__antithesis_instrumentation__.Notify(27666)
		}
	}
	__antithesis_instrumentation__.Notify(27657)

	fixtureFolder := generatorToStorageFolder(gen)
	if _, err := GetFixture(ctx, es, config, gen); err == nil {
		__antithesis_instrumentation__.Notify(27667)
		return Fixture{}, errors.Errorf(
			`fixture %s already exists`, config.ObjectPathToURI(fixtureFolder))
	} else {
		__antithesis_instrumentation__.Notify(27668)
	}
	__antithesis_instrumentation__.Notify(27658)

	dbName := gen.Meta().Name
	if _, err := sqlDB.Exec(`CREATE DATABASE IF NOT EXISTS ` + dbName); err != nil {
		__antithesis_instrumentation__.Notify(27669)
		return Fixture{}, err
	} else {
		__antithesis_instrumentation__.Notify(27670)
	}
	__antithesis_instrumentation__.Notify(27659)
	l := ImportDataLoader{
		FilesPerNode: filesPerNode,
		dbName:       dbName,
	}

	if _, err := l.InitialDataLoad(ctx, sqlDB, gen); err != nil {
		__antithesis_instrumentation__.Notify(27671)
		return Fixture{}, err
	} else {
		__antithesis_instrumentation__.Notify(27672)
	}
	__antithesis_instrumentation__.Notify(27660)

	if config.TableStats {
		__antithesis_instrumentation__.Notify(27673)

		_, err := sqlDB.Exec("DELETE FROM system.table_statistics WHERE true")
		if err != nil {
			__antithesis_instrumentation__.Notify(27676)
			return Fixture{}, errors.Wrapf(err, "while deleting table statistics")
		} else {
			__antithesis_instrumentation__.Notify(27677)
		}
		__antithesis_instrumentation__.Notify(27674)
		g := ctxgroup.WithContext(ctx)
		for _, t := range gen.Tables() {
			__antithesis_instrumentation__.Notify(27678)
			t := t
			g.Go(func() error {
				__antithesis_instrumentation__.Notify(27679)
				log.Infof(ctx, "Creating table stats for %s", t.Name)
				_, err := sqlDB.Exec(fmt.Sprintf(
					`CREATE STATISTICS pre_backup FROM "%s"."%s"`, dbName, t.Name,
				))
				return err
			})
		}
		__antithesis_instrumentation__.Notify(27675)
		if err := g.Wait(); err != nil {
			__antithesis_instrumentation__.Notify(27680)
			return Fixture{}, err
		} else {
			__antithesis_instrumentation__.Notify(27681)
		}
	} else {
		__antithesis_instrumentation__.Notify(27682)
	}
	__antithesis_instrumentation__.Notify(27661)

	g := ctxgroup.WithContext(ctx)
	for _, t := range gen.Tables() {
		__antithesis_instrumentation__.Notify(27683)
		t := t
		g.Go(func() error {
			__antithesis_instrumentation__.Notify(27684)
			q := fmt.Sprintf(`BACKUP "%s"."%s" TO $1`, dbName, t.Name)
			output := config.ObjectPathToURI(filepath.Join(fixtureFolder, t.Name))
			log.Infof(ctx, "Backing %s up to %q...", t.Name, output)
			_, err := sqlDB.Exec(q, output)
			return err
		})
	}
	__antithesis_instrumentation__.Notify(27662)
	if err := g.Wait(); err != nil {
		__antithesis_instrumentation__.Notify(27685)
		return Fixture{}, err
	} else {
		__antithesis_instrumentation__.Notify(27686)
	}
	__antithesis_instrumentation__.Notify(27663)

	return GetFixture(ctx, es, config, gen)
}

type ImportDataLoader struct {
	FilesPerNode int
	InjectStats  bool
	CSVServer    string
	dbName       string
}

func (l ImportDataLoader) InitialDataLoad(
	ctx context.Context, db *gosql.DB, gen workload.Generator,
) (int64, error) {
	__antithesis_instrumentation__.Notify(27687)
	if l.FilesPerNode == 0 {
		__antithesis_instrumentation__.Notify(27690)
		l.FilesPerNode = 1
	} else {
		__antithesis_instrumentation__.Notify(27691)
	}
	__antithesis_instrumentation__.Notify(27688)

	log.Infof(ctx, "starting import of %d tables", len(gen.Tables()))
	start := timeutil.Now()
	bytes, err := ImportFixture(
		ctx, db, gen, l.dbName, l.FilesPerNode, l.InjectStats, l.CSVServer)
	if err != nil {
		__antithesis_instrumentation__.Notify(27692)
		return 0, errors.Wrap(err, `importing fixture`)
	} else {
		__antithesis_instrumentation__.Notify(27693)
	}
	__antithesis_instrumentation__.Notify(27689)
	elapsed := timeutil.Since(start)
	log.Infof(ctx, "imported %s bytes in %d tables (took %s, %s)",
		humanizeutil.IBytes(bytes), len(gen.Tables()), elapsed, humanizeutil.DataRate(bytes, elapsed))

	return bytes, nil
}

func ImportFixture(
	ctx context.Context,
	sqlDB *gosql.DB,
	gen workload.Generator,
	dbName string,
	filesPerNode int,
	injectStats bool,
	csvServer string,
) (int64, error) {
	__antithesis_instrumentation__.Notify(27694)
	for _, t := range gen.Tables() {
		__antithesis_instrumentation__.Notify(27702)
		if t.InitialRows.FillBatch == nil {
			__antithesis_instrumentation__.Notify(27703)
			return 0, errors.Errorf(
				`import fixture is not supported for workload %s`, gen.Meta().Name,
			)
		} else {
			__antithesis_instrumentation__.Notify(27704)
		}
	}
	__antithesis_instrumentation__.Notify(27695)

	var numNodes int
	if err := sqlDB.QueryRow(numNodesQuery).Scan(&numNodes); err != nil {
		__antithesis_instrumentation__.Notify(27705)
		if strings.Contains(err.Error(), "operation is unsupported in multi-tenancy mode") {
			__antithesis_instrumentation__.Notify(27706)

			numNodes = 1
		} else {
			__antithesis_instrumentation__.Notify(27707)
			return 0, err
		}
	} else {
		__antithesis_instrumentation__.Notify(27708)
	}
	__antithesis_instrumentation__.Notify(27696)

	var bytesAtomic int64
	g := ctxgroup.WithContext(ctx)
	tables := gen.Tables()
	if injectStats && func() bool {
		__antithesis_instrumentation__.Notify(27709)
		return tablesHaveStats(tables) == true
	}() == true {
		__antithesis_instrumentation__.Notify(27710)

		enableFn := disableAutoStats(ctx, sqlDB)
		defer enableFn()
	} else {
		__antithesis_instrumentation__.Notify(27711)
	}
	__antithesis_instrumentation__.Notify(27697)

	pathPrefix := csvServer
	if pathPrefix == `` {
		__antithesis_instrumentation__.Notify(27712)
		pathPrefix = `workload://`
	} else {
		__antithesis_instrumentation__.Notify(27713)
	}
	__antithesis_instrumentation__.Notify(27698)

	for _, table := range tables {
		__antithesis_instrumentation__.Notify(27714)
		err := createFixtureTable(sqlDB, dbName, table)
		if err != nil {
			__antithesis_instrumentation__.Notify(27715)
			return 0, errors.Wrapf(err, `creating table %s`, table.Name)
		} else {
			__antithesis_instrumentation__.Notify(27716)
		}
	}
	__antithesis_instrumentation__.Notify(27699)

	for _, t := range tables {
		__antithesis_instrumentation__.Notify(27717)
		table := t
		paths := csvServerPaths(pathPrefix, gen, table, numNodes*filesPerNode)
		g.GoCtx(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(27718)
			tableBytes, err := importFixtureTable(
				ctx, sqlDB, dbName, table, paths, ``, injectStats)
			atomic.AddInt64(&bytesAtomic, tableBytes)
			return errors.Wrapf(err, `importing table %s`, table.Name)
		})
	}
	__antithesis_instrumentation__.Notify(27700)
	if err := g.Wait(); err != nil {
		__antithesis_instrumentation__.Notify(27719)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(27720)
	}
	__antithesis_instrumentation__.Notify(27701)
	return atomic.LoadInt64(&bytesAtomic), nil
}

func createFixtureTable(sqlDB *gosql.DB, dbName string, table workload.Table) error {
	__antithesis_instrumentation__.Notify(27721)
	qualifiedTableName := makeQualifiedTableName(dbName, &table)
	createTable := fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS %s %s`,
		qualifiedTableName,
		table.Schema)
	_, err := sqlDB.Exec(createTable)
	return err
}

func importFixtureTable(
	ctx context.Context,
	sqlDB *gosql.DB,
	dbName string,
	table workload.Table,
	paths []string,
	output string,
	injectStats bool,
) (int64, error) {
	__antithesis_instrumentation__.Notify(27722)
	start := timeutil.Now()
	var buf bytes.Buffer
	var params []interface{}

	qualifiedTableName := makeQualifiedTableName(dbName, &table)
	fmt.Fprintf(&buf, `IMPORT INTO %s CSV DATA (`, qualifiedTableName)

	for _, path := range paths {
		__antithesis_instrumentation__.Notify(27730)
		params = append(params, path)
		if len(params) != 1 {
			__antithesis_instrumentation__.Notify(27732)
			buf.WriteString(`,`)
		} else {
			__antithesis_instrumentation__.Notify(27733)
		}
		__antithesis_instrumentation__.Notify(27731)
		fmt.Fprintf(&buf, `$%d`, len(params))
	}
	__antithesis_instrumentation__.Notify(27723)
	buf.WriteString(`) WITH nullif='NULL'`)
	if len(output) > 0 {
		__antithesis_instrumentation__.Notify(27734)
		params = append(params, output)
		fmt.Fprintf(&buf, `, transform=$%d`, len(params))
	} else {
		__antithesis_instrumentation__.Notify(27735)
	}
	__antithesis_instrumentation__.Notify(27724)
	var rows, index, tableBytes int64
	var discard driver.Value
	res, err := sqlDB.Query(buf.String(), params...)
	if err != nil {
		__antithesis_instrumentation__.Notify(27736)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(27737)
	}
	__antithesis_instrumentation__.Notify(27725)
	defer res.Close()
	if !res.Next() {
		__antithesis_instrumentation__.Notify(27738)
		if err := res.Err(); err != nil {
			__antithesis_instrumentation__.Notify(27740)
			return 0, errors.Wrap(err, "unexpected error during import")
		} else {
			__antithesis_instrumentation__.Notify(27741)
		}
		__antithesis_instrumentation__.Notify(27739)
		return 0, gosql.ErrNoRows
	} else {
		__antithesis_instrumentation__.Notify(27742)
	}
	__antithesis_instrumentation__.Notify(27726)
	resCols, err := res.Columns()
	if err != nil {
		__antithesis_instrumentation__.Notify(27743)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(27744)
	}
	__antithesis_instrumentation__.Notify(27727)
	if len(resCols) == 7 {
		__antithesis_instrumentation__.Notify(27745)
		if err := res.Scan(
			&discard, &discard, &discard, &rows, &index, &discard, &tableBytes,
		); err != nil {
			__antithesis_instrumentation__.Notify(27746)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(27747)
		}
	} else {
		__antithesis_instrumentation__.Notify(27748)
		if err := res.Scan(
			&discard, &discard, &discard, &rows, &index, &tableBytes,
		); err != nil {
			__antithesis_instrumentation__.Notify(27749)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(27750)
		}
	}
	__antithesis_instrumentation__.Notify(27728)
	elapsed := timeutil.Since(start)
	log.Infof(ctx, `imported %s in %s table (%d rows, %d index entries, took %s, %s)`,
		humanizeutil.IBytes(tableBytes), table.Name, rows, index, elapsed,
		humanizeutil.DataRate(tableBytes, elapsed))

	if injectStats && func() bool {
		__antithesis_instrumentation__.Notify(27751)
		return len(table.Stats) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(27752)
		if err := injectStatistics(qualifiedTableName, &table, sqlDB); err != nil {
			__antithesis_instrumentation__.Notify(27753)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(27754)
		}
	} else {
		__antithesis_instrumentation__.Notify(27755)
	}
	__antithesis_instrumentation__.Notify(27729)

	return tableBytes, nil
}

func tablesHaveStats(tables []workload.Table) bool {
	__antithesis_instrumentation__.Notify(27756)
	for _, t := range tables {
		__antithesis_instrumentation__.Notify(27758)
		if len(t.Stats) > 0 {
			__antithesis_instrumentation__.Notify(27759)
			return true
		} else {
			__antithesis_instrumentation__.Notify(27760)
		}
	}
	__antithesis_instrumentation__.Notify(27757)
	return false
}

func disableAutoStats(ctx context.Context, sqlDB *gosql.DB) func() {
	__antithesis_instrumentation__.Notify(27761)
	var autoStatsEnabled bool
	err := sqlDB.QueryRow(
		`SHOW CLUSTER SETTING sql.stats.automatic_collection.enabled`,
	).Scan(&autoStatsEnabled)
	if err != nil {
		__antithesis_instrumentation__.Notify(27764)
		log.Warningf(ctx, "error retrieving automatic stats cluster setting: %v", err)
		return func() { __antithesis_instrumentation__.Notify(27765) }
	} else {
		__antithesis_instrumentation__.Notify(27766)
	}
	__antithesis_instrumentation__.Notify(27762)

	if autoStatsEnabled {
		__antithesis_instrumentation__.Notify(27767)
		_, err = sqlDB.Exec(
			`SET CLUSTER SETTING sql.stats.automatic_collection.enabled=false`,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(27769)
			log.Warningf(ctx, "error disabling automatic stats: %v", err)
			return func() { __antithesis_instrumentation__.Notify(27770) }
		} else {
			__antithesis_instrumentation__.Notify(27771)
		}
		__antithesis_instrumentation__.Notify(27768)
		return func() {
			__antithesis_instrumentation__.Notify(27772)
			_, err := sqlDB.Exec(
				`SET CLUSTER SETTING sql.stats.automatic_collection.enabled=true`,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(27773)
				log.Warningf(ctx, "error enabling automatic stats: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(27774)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(27775)
	}
	__antithesis_instrumentation__.Notify(27763)

	return func() { __antithesis_instrumentation__.Notify(27776) }
}

func injectStatistics(qualifiedTableName string, table *workload.Table, sqlDB *gosql.DB) error {
	__antithesis_instrumentation__.Notify(27777)
	var encoded []byte
	encoded, err := json.Marshal(table.Stats)
	if err != nil {
		__antithesis_instrumentation__.Notify(27780)
		return err
	} else {
		__antithesis_instrumentation__.Notify(27781)
	}
	__antithesis_instrumentation__.Notify(27778)
	if _, err := sqlDB.Exec(
		fmt.Sprintf(`ALTER TABLE %s INJECT STATISTICS '%s'`, qualifiedTableName, encoded),
	); err != nil {
		__antithesis_instrumentation__.Notify(27782)
		if strings.Contains(err.Error(), "syntax error") {
			__antithesis_instrumentation__.Notify(27784)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(27785)
		}
		__antithesis_instrumentation__.Notify(27783)
		return err
	} else {
		__antithesis_instrumentation__.Notify(27786)
	}
	__antithesis_instrumentation__.Notify(27779)
	return nil
}

func makeQualifiedTableName(dbName string, table *workload.Table) string {
	__antithesis_instrumentation__.Notify(27787)
	if dbName == "" {
		__antithesis_instrumentation__.Notify(27789)
		return fmt.Sprintf(`"%s"`, table.Name)
	} else {
		__antithesis_instrumentation__.Notify(27790)
	}
	__antithesis_instrumentation__.Notify(27788)
	return fmt.Sprintf(`"%s"."%s"`, dbName, table.Name)
}

func RestoreFixture(
	ctx context.Context, sqlDB *gosql.DB, fixture Fixture, database string, injectStats bool,
) (int64, error) {
	__antithesis_instrumentation__.Notify(27791)
	var bytesAtomic int64
	g := ctxgroup.WithContext(ctx)
	genName := fixture.Generator.Meta().Name
	tables := fixture.Generator.Tables()
	if injectStats && func() bool {
		__antithesis_instrumentation__.Notify(27796)
		return tablesHaveStats(tables) == true
	}() == true {
		__antithesis_instrumentation__.Notify(27797)

		enableFn := disableAutoStats(ctx, sqlDB)
		defer enableFn()
	} else {
		__antithesis_instrumentation__.Notify(27798)
	}
	__antithesis_instrumentation__.Notify(27792)
	for _, table := range fixture.Tables {
		__antithesis_instrumentation__.Notify(27799)
		table := table
		g.GoCtx(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(27800)
			start := timeutil.Now()
			restoreStmt := fmt.Sprintf(`RESTORE %s.%s FROM $1 WITH into_db=$2`, genName, table.TableName)
			log.Infof(ctx, "Restoring from %s", table.BackupURI)
			var rows, index, tableBytes int64
			var discard interface{}
			res, err := sqlDB.Query(restoreStmt, table.BackupURI, database)
			if err != nil {
				__antithesis_instrumentation__.Notify(27805)
				return errors.Wrapf(err, "restore: %s", table.BackupURI)
			} else {
				__antithesis_instrumentation__.Notify(27806)
			}
			__antithesis_instrumentation__.Notify(27801)
			defer res.Close()
			if !res.Next() {
				__antithesis_instrumentation__.Notify(27807)
				if err := res.Err(); err != nil {
					__antithesis_instrumentation__.Notify(27809)
					return errors.Wrap(err, "unexpected error during restore")
				} else {
					__antithesis_instrumentation__.Notify(27810)
				}
				__antithesis_instrumentation__.Notify(27808)
				return gosql.ErrNoRows
			} else {
				__antithesis_instrumentation__.Notify(27811)
			}
			__antithesis_instrumentation__.Notify(27802)
			resCols, err := res.Columns()
			if err != nil {
				__antithesis_instrumentation__.Notify(27812)
				return err
			} else {
				__antithesis_instrumentation__.Notify(27813)
			}
			__antithesis_instrumentation__.Notify(27803)
			if len(resCols) == 7 {
				__antithesis_instrumentation__.Notify(27814)
				if err := res.Scan(
					&discard, &discard, &discard, &rows, &index, &discard, &tableBytes,
				); err != nil {
					__antithesis_instrumentation__.Notify(27815)
					return err
				} else {
					__antithesis_instrumentation__.Notify(27816)
				}
			} else {
				__antithesis_instrumentation__.Notify(27817)
				if err := res.Scan(
					&discard, &discard, &discard, &rows, &index, &tableBytes,
				); err != nil {
					__antithesis_instrumentation__.Notify(27818)
					return err
				} else {
					__antithesis_instrumentation__.Notify(27819)
				}
			}
			__antithesis_instrumentation__.Notify(27804)
			atomic.AddInt64(&bytesAtomic, tableBytes)
			elapsed := timeutil.Since(start)
			log.Infof(ctx, `loaded %s table %s in %s (%d rows, %d index entries, %s)`,
				humanizeutil.IBytes(tableBytes), table.TableName, elapsed, rows, index,
				humanizeutil.IBytes(int64(float64(tableBytes)/elapsed.Seconds())))
			return nil
		})
	}
	__antithesis_instrumentation__.Notify(27793)
	if err := g.Wait(); err != nil {
		__antithesis_instrumentation__.Notify(27820)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(27821)
	}
	__antithesis_instrumentation__.Notify(27794)
	if injectStats {
		__antithesis_instrumentation__.Notify(27822)
		for i := range tables {
			__antithesis_instrumentation__.Notify(27823)
			t := &tables[i]
			if len(t.Stats) > 0 {
				__antithesis_instrumentation__.Notify(27824)
				qualifiedTableName := makeQualifiedTableName(genName, t)
				if err := injectStatistics(qualifiedTableName, t, sqlDB); err != nil {
					__antithesis_instrumentation__.Notify(27825)
					return 0, err
				} else {
					__antithesis_instrumentation__.Notify(27826)
				}
			} else {
				__antithesis_instrumentation__.Notify(27827)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(27828)
	}
	__antithesis_instrumentation__.Notify(27795)
	return atomic.LoadInt64(&bytesAtomic), nil
}

func listDir(
	ctx context.Context, es cloud.ExternalStorage, lsFn cloud.ListingFn, dir string,
) error {
	__antithesis_instrumentation__.Notify(27829)

	if dir[len(dir)-1] != '/' {
		__antithesis_instrumentation__.Notify(27832)
		dir = dir + "/"
	} else {
		__antithesis_instrumentation__.Notify(27833)
	}
	__antithesis_instrumentation__.Notify(27830)
	if log.V(1) {
		__antithesis_instrumentation__.Notify(27834)
		log.Infof(ctx, "Listing %s", dir)
	} else {
		__antithesis_instrumentation__.Notify(27835)
	}
	__antithesis_instrumentation__.Notify(27831)
	return es.List(ctx, dir, "/", lsFn)
}

func ListFixtures(
	ctx context.Context, es cloud.ExternalStorage, config FixtureConfig,
) ([]string, error) {
	__antithesis_instrumentation__.Notify(27836)
	var fixtures []string
	for _, gen := range workload.Registered() {
		__antithesis_instrumentation__.Notify(27838)
		if err := listDir(ctx, es, func(s string) error {
			__antithesis_instrumentation__.Notify(27839)

			fixtures = append(fixtures, filepath.Join(gen.Name, s))
			return nil
		}, gen.Name); err != nil {
			__antithesis_instrumentation__.Notify(27840)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(27841)
		}
	}
	__antithesis_instrumentation__.Notify(27837)
	return fixtures, nil
}
