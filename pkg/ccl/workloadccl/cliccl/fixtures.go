package cliccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/ccl/workloadccl"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/workload"
	workloadcli "github.com/cockroachdb/cockroach/pkg/workload/cli"
	"github.com/cockroachdb/cockroach/pkg/workload/workloadsql"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var defaultConfig = workloadccl.FixtureConfig{
	StorageProvider: "gs",
	AuthParams:      "AUTH=implicit",
	Bucket:          `cockroach-fixtures`,
	Basename:        `workload`,
}

func config() workloadccl.FixtureConfig {
	__antithesis_instrumentation__.Notify(27502)
	config := defaultConfig
	if len(*providerOverride) > 0 {
		__antithesis_instrumentation__.Notify(27507)
		config.StorageProvider = *providerOverride
	} else {
		__antithesis_instrumentation__.Notify(27508)
	}
	__antithesis_instrumentation__.Notify(27503)
	if len(*bucketOverride) > 0 {
		__antithesis_instrumentation__.Notify(27509)
		config.Bucket = *bucketOverride
	} else {
		__antithesis_instrumentation__.Notify(27510)
	}
	__antithesis_instrumentation__.Notify(27504)
	if len(*prefixOverride) > 0 {
		__antithesis_instrumentation__.Notify(27511)
		config.Basename = *prefixOverride
	} else {
		__antithesis_instrumentation__.Notify(27512)
	}
	__antithesis_instrumentation__.Notify(27505)
	if len(*authParamsOverride) > 0 {
		__antithesis_instrumentation__.Notify(27513)
		config.AuthParams = *authParamsOverride
	} else {
		__antithesis_instrumentation__.Notify(27514)
	}
	__antithesis_instrumentation__.Notify(27506)
	config.CSVServerURL = *fixturesMakeImportCSVServerURL
	config.TableStats = *fixturesMakeTableStats
	return config
}

var fixturesCmd = workloadcli.SetCmdDefaults(&cobra.Command{
	Use:   `fixtures`,
	Short: `tools for quickly synthesizing and loading large datasets`,
})
var fixturesListCmd = workloadcli.SetCmdDefaults(&cobra.Command{
	Use:   `list`,
	Short: `list all fixtures stored on GCS`,
	Run:   workloadcli.HandleErrs(fixturesList),
})
var fixturesMakeCmd = workloadcli.SetCmdDefaults(&cobra.Command{
	Use:   `make`,
	Short: `IMPORT a fixture and then store a BACKUP of it on GCS`,
})
var fixturesLoadCmd = workloadcli.SetCmdDefaults(&cobra.Command{
	Use:   `load`,
	Short: `load a fixture into a running cluster. An enterprise license is required.`,
})
var fixturesImportCmd = workloadcli.SetCmdDefaults(&cobra.Command{
	Use:   `import`,
	Short: `import a fixture into a running cluster. An enterprise license is NOT required.`,
})
var fixturesURLCmd = workloadcli.SetCmdDefaults(&cobra.Command{
	Use:   `url`,
	Short: `generate the GCS URL for a fixture`,
})

var fixturesLoadImportShared = pflag.NewFlagSet(`load/import`, pflag.ContinueOnError)
var fixturesMakeImportShared = pflag.NewFlagSet(`load/import`, pflag.ContinueOnError)

var fixturesMakeImportCSVServerURL = fixturesMakeImportShared.String(
	`csv-server`, ``,
	`Skip saving CSVs to cloud storage, instead get them from a 'csv-server' running at this url`)

var fixturesMakeOnlyTable = fixturesMakeCmd.PersistentFlags().String(
	`only-tables`, ``,
	`Only load the tables with the given comma-separated names`)

var fixturesMakeFilesPerNode = fixturesMakeCmd.PersistentFlags().Int(
	`files-per-node`, 1,
	`number of file URLs to generate per node when using csv-server`)

var fixturesMakeTableStats = fixturesMakeCmd.PersistentFlags().Bool(
	`table-stats`, true,
	`generate full table statistics for all tables`)

var fixturesImportFilesPerNode = fixturesImportCmd.PersistentFlags().Int(
	`files-per-node`, 1,
	`number of file URLs to generate per node`)

var fixturesRunChecks = fixturesLoadImportShared.Bool(
	`checks`, true, `Run validity checks on the loaded fixture`)

var fixturesImportInjectStats = fixturesImportCmd.PersistentFlags().Bool(
	`inject-stats`, true, `Inject pre-calculated statistics if they are available`)

var bucketOverride, prefixOverride, providerOverride, authParamsOverride *string

func init() {
	bucketOverride = fixturesCmd.PersistentFlags().String(`bucket-override`, ``, ``)
	prefixOverride = fixturesCmd.PersistentFlags().String(`prefix-override`, ``, ``)
	authParamsOverride = fixturesCmd.PersistentFlags().String(
		`auth-params-override`, ``,
		`Override authentication parameters needed to access fixture; Cloud specific`)
	providerOverride = fixturesCmd.PersistentFlags().String(
		`provider-override`, ``,
		`Override storage provider type (Default: gcs; Also available s3 and azure`)
	_ = fixturesCmd.PersistentFlags().MarkHidden(`bucket-override`)
	_ = fixturesCmd.PersistentFlags().MarkHidden(`prefix-override`)
}

func init() {
	workloadcli.AddSubCmd(func(userFacing bool) *cobra.Command {
		for _, meta := range workload.Registered() {
			gen := meta.New()
			var genFlags *pflag.FlagSet
			if f, ok := gen.(workload.Flagser); ok {
				genFlags = f.Flags().FlagSet

				for flagName, meta := range f.Flags().Meta {
					if meta.RuntimeOnly || meta.CheckConsistencyOnly {
						_ = genFlags.MarkHidden(flagName)
					}
				}
			}

			genMakeCmd := workloadcli.SetCmdDefaults(&cobra.Command{
				Use:  meta.Name + ` [CRDB URI]`,
				Args: cobra.RangeArgs(0, 1),
			})
			genMakeCmd.Flags().AddFlagSet(genFlags)
			genMakeCmd.Flags().AddFlagSet(fixturesMakeImportShared)
			genMakeCmd.Run = workloadcli.CmdHelper(gen, fixturesMake)
			fixturesMakeCmd.AddCommand(genMakeCmd)

			genLoadCmd := workloadcli.SetCmdDefaults(&cobra.Command{
				Use:  meta.Name + ` [CRDB URI]`,
				Args: cobra.RangeArgs(0, 1),
			})
			genLoadCmd.Flags().AddFlagSet(genFlags)
			genLoadCmd.Flags().AddFlagSet(fixturesLoadImportShared)
			genLoadCmd.Run = workloadcli.CmdHelper(gen, fixturesLoad)
			fixturesLoadCmd.AddCommand(genLoadCmd)

			genImportCmd := workloadcli.SetCmdDefaults(&cobra.Command{
				Use:  meta.Name + ` [CRDB URI]`,
				Args: cobra.RangeArgs(0, 1),
			})
			genImportCmd.Flags().AddFlagSet(genFlags)
			genImportCmd.Flags().AddFlagSet(fixturesLoadImportShared)
			genImportCmd.Flags().AddFlagSet(fixturesMakeImportShared)
			genImportCmd.Run = workloadcli.CmdHelper(gen, fixturesImport)
			fixturesImportCmd.AddCommand(genImportCmd)

			genURLCmd := workloadcli.SetCmdDefaults(&cobra.Command{
				Use:  meta.Name,
				Args: cobra.NoArgs,
			})
			genURLCmd.Flags().AddFlagSet(genFlags)
			genURLCmd.Run = fixturesURL(gen)
			fixturesURLCmd.AddCommand(genURLCmd)
		}
		fixturesCmd.AddCommand(fixturesListCmd)
		fixturesCmd.AddCommand(fixturesMakeCmd)
		fixturesCmd.AddCommand(fixturesLoadCmd)
		fixturesCmd.AddCommand(fixturesImportCmd)
		fixturesCmd.AddCommand(fixturesURLCmd)
		return fixturesCmd
	})
}

func fixturesList(_ *cobra.Command, _ []string) error {
	__antithesis_instrumentation__.Notify(27515)
	ctx := context.Background()
	es, err := workloadccl.GetStorage(ctx, config())
	if err != nil {
		__antithesis_instrumentation__.Notify(27520)
		return err
	} else {
		__antithesis_instrumentation__.Notify(27521)
	}
	__antithesis_instrumentation__.Notify(27516)
	defer func() { __antithesis_instrumentation__.Notify(27522); _ = es.Close() }()
	__antithesis_instrumentation__.Notify(27517)
	fixtures, err := workloadccl.ListFixtures(ctx, es, config())
	if err != nil {
		__antithesis_instrumentation__.Notify(27523)
		return err
	} else {
		__antithesis_instrumentation__.Notify(27524)
	}
	__antithesis_instrumentation__.Notify(27518)
	for _, fixture := range fixtures {
		__antithesis_instrumentation__.Notify(27525)
		fmt.Println(fixture)
	}
	__antithesis_instrumentation__.Notify(27519)
	return nil
}

type filteringGenerator struct {
	gen    workload.Generator
	filter map[string]struct{}
}

func (f filteringGenerator) Meta() workload.Meta {
	__antithesis_instrumentation__.Notify(27526)
	return f.gen.Meta()
}

func (f filteringGenerator) Tables() []workload.Table {
	__antithesis_instrumentation__.Notify(27527)
	ret := make([]workload.Table, 0)
	for _, t := range f.gen.Tables() {
		__antithesis_instrumentation__.Notify(27529)
		if _, ok := f.filter[t.Name]; ok {
			__antithesis_instrumentation__.Notify(27530)
			ret = append(ret, t)
		} else {
			__antithesis_instrumentation__.Notify(27531)
		}
	}
	__antithesis_instrumentation__.Notify(27528)
	return ret
}

func fixturesMake(gen workload.Generator, urls []string, _ string) error {
	__antithesis_instrumentation__.Notify(27532)
	ctx := context.Background()
	gcs, err := workloadccl.GetStorage(ctx, config())
	if err != nil {
		__antithesis_instrumentation__.Notify(27539)
		return err
	} else {
		__antithesis_instrumentation__.Notify(27540)
	}
	__antithesis_instrumentation__.Notify(27533)
	defer func() { __antithesis_instrumentation__.Notify(27541); _ = gcs.Close() }()
	__antithesis_instrumentation__.Notify(27534)

	sqlDB, err := gosql.Open(`cockroach`, strings.Join(urls, ` `))
	if err != nil {
		__antithesis_instrumentation__.Notify(27542)
		return err
	} else {
		__antithesis_instrumentation__.Notify(27543)
	}
	__antithesis_instrumentation__.Notify(27535)
	if *fixturesMakeOnlyTable != "" {
		__antithesis_instrumentation__.Notify(27544)
		tableNames := strings.Split(*fixturesMakeOnlyTable, ",")
		if len(tableNames) == 0 {
			__antithesis_instrumentation__.Notify(27547)
			return errors.New("no table names specified")
		} else {
			__antithesis_instrumentation__.Notify(27548)
		}
		__antithesis_instrumentation__.Notify(27545)
		filter := make(map[string]struct{}, len(tableNames))
		for _, tableName := range tableNames {
			__antithesis_instrumentation__.Notify(27549)
			filter[tableName] = struct{}{}
		}
		__antithesis_instrumentation__.Notify(27546)
		gen = filteringGenerator{
			gen:    gen,
			filter: filter,
		}
	} else {
		__antithesis_instrumentation__.Notify(27550)
	}
	__antithesis_instrumentation__.Notify(27536)
	filesPerNode := *fixturesMakeFilesPerNode
	fixture, err := workloadccl.MakeFixture(ctx, sqlDB, gcs, config(), gen, filesPerNode)
	if err != nil {
		__antithesis_instrumentation__.Notify(27551)
		return err
	} else {
		__antithesis_instrumentation__.Notify(27552)
	}
	__antithesis_instrumentation__.Notify(27537)
	for _, table := range fixture.Tables {
		__antithesis_instrumentation__.Notify(27553)
		log.Infof(ctx, `stored backup %s`, table.BackupURI)
	}
	__antithesis_instrumentation__.Notify(27538)
	return nil
}

type restoreDataLoader struct {
	fixture  workloadccl.Fixture
	database string
}

func (l restoreDataLoader) InitialDataLoad(
	ctx context.Context, db *gosql.DB, gen workload.Generator,
) (int64, error) {
	__antithesis_instrumentation__.Notify(27554)
	log.Infof(ctx, "starting restore of %d tables", len(gen.Tables()))
	start := timeutil.Now()
	bytes, err := workloadccl.RestoreFixture(ctx, db, l.fixture, l.database, true)
	if err != nil {
		__antithesis_instrumentation__.Notify(27556)
		return 0, errors.Wrap(err, `restoring fixture`)
	} else {
		__antithesis_instrumentation__.Notify(27557)
	}
	__antithesis_instrumentation__.Notify(27555)
	elapsed := timeutil.Since(start)
	log.Infof(ctx, "restored %s bytes in %d tables (took %s, %s)",
		humanizeutil.IBytes(bytes), len(gen.Tables()), elapsed, humanizeutil.DataRate(bytes, elapsed))
	return bytes, nil
}

func fixturesLoad(gen workload.Generator, urls []string, dbName string) error {
	__antithesis_instrumentation__.Notify(27558)
	ctx := context.Background()
	gcs, err := workloadccl.GetStorage(ctx, config())
	if err != nil {
		__antithesis_instrumentation__.Notify(27566)
		return err
	} else {
		__antithesis_instrumentation__.Notify(27567)
	}
	__antithesis_instrumentation__.Notify(27559)
	defer func() { __antithesis_instrumentation__.Notify(27568); _ = gcs.Close() }()
	__antithesis_instrumentation__.Notify(27560)

	sqlDB, err := gosql.Open(`cockroach`, strings.Join(urls, ` `))
	if err != nil {
		__antithesis_instrumentation__.Notify(27569)
		return err
	} else {
		__antithesis_instrumentation__.Notify(27570)
	}
	__antithesis_instrumentation__.Notify(27561)
	if _, err := sqlDB.Exec(`CREATE DATABASE IF NOT EXISTS ` + dbName); err != nil {
		__antithesis_instrumentation__.Notify(27571)
		return err
	} else {
		__antithesis_instrumentation__.Notify(27572)
	}
	__antithesis_instrumentation__.Notify(27562)

	fixture, err := workloadccl.GetFixture(ctx, gcs, config(), gen)
	if err != nil {
		__antithesis_instrumentation__.Notify(27573)
		return errors.Wrap(err, `finding fixture`)
	} else {
		__antithesis_instrumentation__.Notify(27574)
	}
	__antithesis_instrumentation__.Notify(27563)

	l := restoreDataLoader{fixture: fixture, database: dbName}
	if _, err := workloadsql.Setup(ctx, sqlDB, gen, l); err != nil {
		__antithesis_instrumentation__.Notify(27575)
		return err
	} else {
		__antithesis_instrumentation__.Notify(27576)
	}
	__antithesis_instrumentation__.Notify(27564)

	if hooks, ok := gen.(workload.Hookser); *fixturesRunChecks && func() bool {
		__antithesis_instrumentation__.Notify(27577)
		return ok == true
	}() == true {
		__antithesis_instrumentation__.Notify(27578)
		if consistencyCheckFn := hooks.Hooks().CheckConsistency; consistencyCheckFn != nil {
			__antithesis_instrumentation__.Notify(27579)
			log.Info(ctx, "fixture is imported; now running consistency checks (ctrl-c to abort)")
			if err := consistencyCheckFn(ctx, sqlDB); err != nil {
				__antithesis_instrumentation__.Notify(27580)
				return err
			} else {
				__antithesis_instrumentation__.Notify(27581)
			}
		} else {
			__antithesis_instrumentation__.Notify(27582)
		}
	} else {
		__antithesis_instrumentation__.Notify(27583)
	}
	__antithesis_instrumentation__.Notify(27565)

	return nil
}

func fixturesImport(gen workload.Generator, urls []string, dbName string) error {
	__antithesis_instrumentation__.Notify(27584)
	ctx := context.Background()
	sqlDB, err := gosql.Open(`cockroach`, strings.Join(urls, ` `))
	if err != nil {
		__antithesis_instrumentation__.Notify(27589)
		return err
	} else {
		__antithesis_instrumentation__.Notify(27590)
	}
	__antithesis_instrumentation__.Notify(27585)
	if _, err := sqlDB.Exec(`CREATE DATABASE IF NOT EXISTS ` + dbName); err != nil {
		__antithesis_instrumentation__.Notify(27591)
		return err
	} else {
		__antithesis_instrumentation__.Notify(27592)
	}
	__antithesis_instrumentation__.Notify(27586)

	l := workloadccl.ImportDataLoader{
		FilesPerNode: *fixturesImportFilesPerNode,
		InjectStats:  *fixturesImportInjectStats,
		CSVServer:    *fixturesMakeImportCSVServerURL,
	}
	if _, err := workloadsql.Setup(ctx, sqlDB, gen, l); err != nil {
		__antithesis_instrumentation__.Notify(27593)
		return err
	} else {
		__antithesis_instrumentation__.Notify(27594)
	}
	__antithesis_instrumentation__.Notify(27587)

	if hooks, ok := gen.(workload.Hookser); *fixturesRunChecks && func() bool {
		__antithesis_instrumentation__.Notify(27595)
		return ok == true
	}() == true {
		__antithesis_instrumentation__.Notify(27596)
		if consistencyCheckFn := hooks.Hooks().CheckConsistency; consistencyCheckFn != nil {
			__antithesis_instrumentation__.Notify(27597)
			log.Info(ctx, "fixture is restored; now running consistency checks (ctrl-c to abort)")
			if err := consistencyCheckFn(ctx, sqlDB); err != nil {
				__antithesis_instrumentation__.Notify(27598)
				return err
			} else {
				__antithesis_instrumentation__.Notify(27599)
			}
		} else {
			__antithesis_instrumentation__.Notify(27600)
		}
	} else {
		__antithesis_instrumentation__.Notify(27601)
	}
	__antithesis_instrumentation__.Notify(27588)

	return nil
}

func fixturesURL(gen workload.Generator) func(*cobra.Command, []string) {
	__antithesis_instrumentation__.Notify(27602)
	return workloadcli.HandleErrs(func(*cobra.Command, []string) error {
		__antithesis_instrumentation__.Notify(27603)
		if h, ok := gen.(workload.Hookser); ok {
			__antithesis_instrumentation__.Notify(27605)
			if err := h.Hooks().Validate(); err != nil {
				__antithesis_instrumentation__.Notify(27606)
				return err
			} else {
				__antithesis_instrumentation__.Notify(27607)
			}
		} else {
			__antithesis_instrumentation__.Notify(27608)
		}
		__antithesis_instrumentation__.Notify(27604)

		fmt.Println(workloadccl.FixtureURL(config(), gen))
		return nil
	})
}
