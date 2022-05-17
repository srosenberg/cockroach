package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"context"
	"database/sql/driver"
	hx "encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	apd "github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/cockroach/pkg/cli/clierror"
	"github.com/cockroachdb/cockroach/pkg/cli/clierrorplus"
	"github.com/cockroachdb/cockroach/pkg/cli/clisqlclient"
	"github.com/cockroachdb/cockroach/pkg/cli/exit"
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/doctor"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/builtins"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq"
	"github.com/spf13/cobra"
)

var debugDoctorCmd = &cobra.Command{
	Use:   "doctor [command]",
	Short: "run a cockroach doctor tool command",
	Long: `
Run the doctor tool to recreate or examine cockroach system table contents. 
System tables are queried either from a live cluster or from an unzipped 
debug.zip.
`,
}

var doctorExamineCmd = &cobra.Command{
	Use:   "examine [cluster|zipdir]",
	Short: "examine system tables for inconsistencies",
	Long: `
Run the doctor tool to examine the system table contents and perform validation
checks. System tables are queried either from a live cluster or from an unzipped 
debug.zip.
`,
}

var doctorRecreateCmd = &cobra.Command{
	Use:   "recreate [cluster|zipdir]",
	Short: "prints SQL that tries to recreate system table state",
	Long: `
Run the doctor tool to examine system tables and generate SQL statements that,
when run on an empty cluster, recreate that state as closely as possible. System
tables are queried either from a live cluster or from an unzipped debug.zip.
`,
}

type doctorFn = func(
	version *clusterversion.ClusterVersion,
	descTable doctor.DescriptorTable,
	namespaceTable doctor.NamespaceTable,
	jobsTable doctor.JobsTable,
	out io.Writer,
) (err error)

func makeZipDirCommand(fn doctorFn) *cobra.Command {
	__antithesis_instrumentation__.Notify(32474)
	return &cobra.Command{
		Use:   "zipdir <debug_zip_dir> [version]",
		Short: "run doctor tool on data from an unzipped debug.zip",
		Long: `
Run the doctor tool on system data from an unzipped debug.zip. This command
requires the path of the unzipped 'debug' directory as its argument. A version
can be optionally specified, which will be used enable / disable validation
that may not exist on downlevel versions.
`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			__antithesis_instrumentation__.Notify(32475)
			descs, ns, jobs, err := fromZipDir(args[0])
			var version *clusterversion.ClusterVersion
			if len(args) == 2 {
				__antithesis_instrumentation__.Notify(32478)
				version = &clusterversion.ClusterVersion{
					Version: roachpb.MustParseVersion(args[1]),
				}
			} else {
				__antithesis_instrumentation__.Notify(32479)
			}
			__antithesis_instrumentation__.Notify(32476)
			if err != nil {
				__antithesis_instrumentation__.Notify(32480)
				return err
			} else {
				__antithesis_instrumentation__.Notify(32481)
			}
			__antithesis_instrumentation__.Notify(32477)
			return fn(version, descs, ns, jobs, os.Stdout)
		},
	}
}

func makeClusterCommand(fn doctorFn) *cobra.Command {
	__antithesis_instrumentation__.Notify(32482)
	return &cobra.Command{
		Use:   "cluster --url=<cluster connection string>",
		Short: "run doctor tool on live cockroach cluster",
		Long: `
Run the doctor tool system data from a live cluster specified by --url.
`,
		Args: cobra.NoArgs,
		RunE: clierrorplus.MaybeDecorateError(
			func(cmd *cobra.Command, args []string) (resErr error) {
				__antithesis_instrumentation__.Notify(32483)
				sqlConn, err := makeSQLClient("cockroach doctor", useSystemDb)
				if err != nil {
					__antithesis_instrumentation__.Notify(32487)
					return errors.Wrap(err, "could not establish connection to cluster")
				} else {
					__antithesis_instrumentation__.Notify(32488)
				}
				__antithesis_instrumentation__.Notify(32484)
				defer func() {
					__antithesis_instrumentation__.Notify(32489)
					resErr = errors.CombineErrors(resErr, sqlConn.Close())
				}()
				__antithesis_instrumentation__.Notify(32485)
				descs, ns, jobs, err := fromCluster(sqlConn, cliCtx.cmdTimeout)
				if err != nil {
					__antithesis_instrumentation__.Notify(32490)
					return err
				} else {
					__antithesis_instrumentation__.Notify(32491)
				}
				__antithesis_instrumentation__.Notify(32486)
				return fn(nil, descs, ns, jobs, os.Stdout)
			}),
	}
}

func deprecateCommand(cmd *cobra.Command) *cobra.Command {
	__antithesis_instrumentation__.Notify(32492)
	cmd.Hidden = true
	cmd.Deprecated = fmt.Sprintf("use 'doctor examine %s' instead.", cmd.Name())
	return cmd
}

var doctorExamineClusterCmd = makeClusterCommand(runDoctorExamine)
var doctorExamineZipDirCmd = makeZipDirCommand(runDoctorExamine)
var doctorExamineFallbackClusterCmd = deprecateCommand(makeClusterCommand(runDoctorExamine))
var doctorExamineFallbackZipDirCmd = deprecateCommand(makeZipDirCommand(runDoctorExamine))
var doctorRecreateClusterCmd = makeClusterCommand(runDoctorRecreate)
var doctorRecreateZipDirCmd = makeZipDirCommand(runDoctorRecreate)

func runDoctorRecreate(
	_ *clusterversion.ClusterVersion,
	descTable doctor.DescriptorTable,
	namespaceTable doctor.NamespaceTable,
	jobsTable doctor.JobsTable,
	out io.Writer,
) (err error) {
	__antithesis_instrumentation__.Notify(32493)
	return doctor.DumpSQL(out, descTable, namespaceTable)
}

func runDoctorExamine(
	version *clusterversion.ClusterVersion,
	descTable doctor.DescriptorTable,
	namespaceTable doctor.NamespaceTable,
	jobsTable doctor.JobsTable,
	out io.Writer,
) (err error) {
	__antithesis_instrumentation__.Notify(32494)
	if version == nil {
		__antithesis_instrumentation__.Notify(32498)
		version = &clusterversion.ClusterVersion{
			Version: clusterversion.DoctorBinaryVersion,
		}
	} else {
		__antithesis_instrumentation__.Notify(32499)
	}
	__antithesis_instrumentation__.Notify(32495)
	var valid bool
	valid, err = doctor.Examine(
		context.Background(),
		*version,
		descTable,
		namespaceTable,
		jobsTable,
		debugCtx.verbose,
		out)
	if err != nil {
		__antithesis_instrumentation__.Notify(32500)
		return err
	} else {
		__antithesis_instrumentation__.Notify(32501)
	}
	__antithesis_instrumentation__.Notify(32496)
	if !valid {
		__antithesis_instrumentation__.Notify(32502)
		return clierror.NewError(errors.New("validation failed"),
			exit.DoctorValidationFailed())
	} else {
		__antithesis_instrumentation__.Notify(32503)
	}
	__antithesis_instrumentation__.Notify(32497)
	fmt.Fprintln(out, "No problems found!")
	return nil
}

func fromCluster(
	sqlConn clisqlclient.Conn, timeout time.Duration,
) (
	descTable doctor.DescriptorTable,
	namespaceTable doctor.NamespaceTable,
	jobsTable doctor.JobsTable,
	retErr error,
) {
	__antithesis_instrumentation__.Notify(32504)
	ctx := context.Background()
	if timeout != 0 {
		__antithesis_instrumentation__.Notify(32511)
		if err := sqlConn.Exec(ctx,
			`SET statement_timeout = $1`, timeout.String()); err != nil {
			__antithesis_instrumentation__.Notify(32512)
			return nil, nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(32513)
		}
	} else {
		__antithesis_instrumentation__.Notify(32514)
	}
	__antithesis_instrumentation__.Notify(32505)
	stmt := `
SELECT id, descriptor, crdb_internal_mvcc_timestamp AS mod_time_logical
FROM system.descriptor ORDER BY id`
	checkColumnExistsStmt := "SELECT crdb_internal_mvcc_timestamp FROM system.descriptor LIMIT 1"
	_, err := sqlConn.QueryRow(ctx, checkColumnExistsStmt)

	if pqErr := (*pq.Error)(nil); errors.As(err, &pqErr) {
		__antithesis_instrumentation__.Notify(32515)
		if pgcode.MakeCode(string(pqErr.Code)) == pgcode.UndefinedColumn {
			__antithesis_instrumentation__.Notify(32516)
			stmt = `
SELECT id, descriptor, NULL AS mod_time_logical
FROM system.descriptor ORDER BY id`
		} else {
			__antithesis_instrumentation__.Notify(32517)
		}
	} else {
		__antithesis_instrumentation__.Notify(32518)
		if err != nil {
			__antithesis_instrumentation__.Notify(32519)
			return nil, nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(32520)
		}
	}
	__antithesis_instrumentation__.Notify(32506)
	descTable = make([]doctor.DescriptorTableRow, 0)

	if err := selectRowsMap(sqlConn, stmt, make([]driver.Value, 3), func(vals []driver.Value) error {
		__antithesis_instrumentation__.Notify(32521)
		var row doctor.DescriptorTableRow
		if id, ok := vals[0].(int64); ok {
			__antithesis_instrumentation__.Notify(32525)
			row.ID = id
		} else {
			__antithesis_instrumentation__.Notify(32526)
			return errors.Errorf("unexpected value: %T of %v", vals[0], vals[0])
		}
		__antithesis_instrumentation__.Notify(32522)
		if descBytes, ok := vals[1].([]byte); ok {
			__antithesis_instrumentation__.Notify(32527)
			row.DescBytes = descBytes
		} else {
			__antithesis_instrumentation__.Notify(32528)
			return errors.Errorf("unexpected value: %T of %v", vals[1], vals[1])
		}
		__antithesis_instrumentation__.Notify(32523)
		if vals[2] == nil {
			__antithesis_instrumentation__.Notify(32529)
			row.ModTime = hlc.Timestamp{WallTime: timeutil.Now().UnixNano()}
		} else {
			__antithesis_instrumentation__.Notify(32530)
			if mt, ok := vals[2].([]byte); ok {
				__antithesis_instrumentation__.Notify(32531)
				decimal, _, err := apd.NewFromString(string(mt))
				if err != nil {
					__antithesis_instrumentation__.Notify(32534)
					return err
				} else {
					__antithesis_instrumentation__.Notify(32535)
				}
				__antithesis_instrumentation__.Notify(32532)
				ts, err := tree.DecimalToHLC(decimal)
				if err != nil {
					__antithesis_instrumentation__.Notify(32536)
					return err
				} else {
					__antithesis_instrumentation__.Notify(32537)
				}
				__antithesis_instrumentation__.Notify(32533)
				row.ModTime = ts
			} else {
				__antithesis_instrumentation__.Notify(32538)
				return errors.Errorf("unexpected value: %T of %v", vals[2], vals[2])
			}
		}
		__antithesis_instrumentation__.Notify(32524)
		descTable = append(descTable, row)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(32539)
		return nil, nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(32540)
	}
	__antithesis_instrumentation__.Notify(32507)

	stmt = `SELECT "parentID", "parentSchemaID", name, id FROM system.namespace`

	checkColumnExistsStmt = `SELECT "parentSchemaID" FROM system.namespace LIMIT 1`
	_, err = sqlConn.QueryRow(ctx, checkColumnExistsStmt)

	if pqErr := (*pq.Error)(nil); errors.As(err, &pqErr) {
		__antithesis_instrumentation__.Notify(32541)
		if pgcode.MakeCode(string(pqErr.Code)) == pgcode.UndefinedColumn {
			__antithesis_instrumentation__.Notify(32542)
			stmt = `
SELECT "parentID", CASE WHEN "parentID" = 0 THEN 0 ELSE 29 END AS "parentSchemaID", name, id
FROM system.namespace`
		} else {
			__antithesis_instrumentation__.Notify(32543)
		}
	} else {
		__antithesis_instrumentation__.Notify(32544)
		if err != nil {
			__antithesis_instrumentation__.Notify(32545)
			return nil, nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(32546)
		}
	}
	__antithesis_instrumentation__.Notify(32508)

	namespaceTable = make([]doctor.NamespaceTableRow, 0)
	if err := selectRowsMap(sqlConn, stmt, make([]driver.Value, 4), func(vals []driver.Value) error {
		__antithesis_instrumentation__.Notify(32547)
		var row doctor.NamespaceTableRow
		if parentID, ok := vals[0].(int64); ok {
			__antithesis_instrumentation__.Notify(32552)
			row.ParentID = descpb.ID(parentID)
		} else {
			__antithesis_instrumentation__.Notify(32553)
			return errors.Errorf("unexpected value: %T of %v", vals[0], vals[0])
		}
		__antithesis_instrumentation__.Notify(32548)
		if schemaID, ok := vals[1].(int64); ok {
			__antithesis_instrumentation__.Notify(32554)
			row.ParentSchemaID = descpb.ID(schemaID)
		} else {
			__antithesis_instrumentation__.Notify(32555)
			return errors.Errorf("unexpected value: %T of %v", vals[0], vals[0])
		}
		__antithesis_instrumentation__.Notify(32549)
		if name, ok := vals[2].(string); ok {
			__antithesis_instrumentation__.Notify(32556)
			row.Name = name
		} else {
			__antithesis_instrumentation__.Notify(32557)
			return errors.Errorf("unexpected value: %T of %v", vals[1], vals[1])
		}
		__antithesis_instrumentation__.Notify(32550)
		if id, ok := vals[3].(int64); ok {
			__antithesis_instrumentation__.Notify(32558)
			row.ID = id
		} else {
			__antithesis_instrumentation__.Notify(32559)
			return errors.Errorf("unexpected value: %T of %v", vals[0], vals[0])
		}
		__antithesis_instrumentation__.Notify(32551)
		namespaceTable = append(namespaceTable, row)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(32560)
		return nil, nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(32561)
	}
	__antithesis_instrumentation__.Notify(32509)

	stmt = `SELECT id, status, payload, progress FROM system.jobs`
	jobsTable = make(doctor.JobsTable, 0)

	if err := selectRowsMap(sqlConn, stmt, make([]driver.Value, 4), func(vals []driver.Value) error {
		__antithesis_instrumentation__.Notify(32562)
		md := jobs.JobMetadata{}
		md.ID = jobspb.JobID(vals[0].(int64))
		md.Status = jobs.Status(vals[1].(string))
		md.Payload = &jobspb.Payload{}
		if err := protoutil.Unmarshal(vals[2].([]byte), md.Payload); err != nil {
			__antithesis_instrumentation__.Notify(32566)
			return err
		} else {
			__antithesis_instrumentation__.Notify(32567)
		}
		__antithesis_instrumentation__.Notify(32563)
		md.Progress = &jobspb.Progress{}

		progressBytes, ok := vals[3].([]byte)
		if !ok {
			__antithesis_instrumentation__.Notify(32568)
			return errors.Errorf("unexpected NULL progress on job row: %v", md)
		} else {
			__antithesis_instrumentation__.Notify(32569)
		}
		__antithesis_instrumentation__.Notify(32564)
		if err := protoutil.Unmarshal(progressBytes, md.Progress); err != nil {
			__antithesis_instrumentation__.Notify(32570)
			return err
		} else {
			__antithesis_instrumentation__.Notify(32571)
		}
		__antithesis_instrumentation__.Notify(32565)
		jobsTable = append(jobsTable, md)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(32572)
		return nil, nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(32573)
	}
	__antithesis_instrumentation__.Notify(32510)

	return descTable, namespaceTable, jobsTable, nil
}

func fromZipDir(
	zipDirPath string,
) (
	descTable doctor.DescriptorTable,
	namespaceTable doctor.NamespaceTable,
	jobsTable doctor.JobsTable,
	retErr error,
) {
	__antithesis_instrumentation__.Notify(32574)

	_ = builtins.AllBuiltinNames

	descTable = make(doctor.DescriptorTable, 0)
	if err := slurp(zipDirPath, "system.descriptor.txt", func(row string) error {
		__antithesis_instrumentation__.Notify(32579)
		fields := strings.Fields(row)
		last := len(fields) - 1
		i, err := strconv.Atoi(fields[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(32582)
			return errors.Wrapf(err, "failed to parse descriptor id %s", fields[0])
		} else {
			__antithesis_instrumentation__.Notify(32583)
		}
		__antithesis_instrumentation__.Notify(32580)

		descBytes, err := hx.DecodeString(fields[last])
		if err != nil {
			__antithesis_instrumentation__.Notify(32584)
			return errors.Wrapf(err, "failed to decode hex descriptor %d", i)
		} else {
			__antithesis_instrumentation__.Notify(32585)
		}
		__antithesis_instrumentation__.Notify(32581)
		ts := hlc.Timestamp{WallTime: timeutil.Now().UnixNano()}
		descTable = append(descTable, doctor.DescriptorTableRow{ID: int64(i), DescBytes: descBytes, ModTime: ts})
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(32586)
		return nil, nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(32587)
	}
	__antithesis_instrumentation__.Notify(32575)

	namespaceFileName := "system.namespace2.txt"
	if _, err := os.Stat(namespaceFileName); err != nil {
		__antithesis_instrumentation__.Notify(32588)
		if !errors.Is(err, os.ErrNotExist) {
			__antithesis_instrumentation__.Notify(32590)

			return nil, nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(32591)
		}
		__antithesis_instrumentation__.Notify(32589)
		namespaceFileName = "system.namespace.txt"
	} else {
		__antithesis_instrumentation__.Notify(32592)
	}
	__antithesis_instrumentation__.Notify(32576)

	namespaceTable = make(doctor.NamespaceTable, 0)
	if err := slurp(zipDirPath, namespaceFileName, func(row string) error {
		__antithesis_instrumentation__.Notify(32593)
		fields := strings.Fields(row)
		parID, err := strconv.Atoi(fields[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(32597)
			return errors.Wrapf(err, "failed to parse parent id %s", fields[0])
		} else {
			__antithesis_instrumentation__.Notify(32598)
		}
		__antithesis_instrumentation__.Notify(32594)
		parSchemaID, err := strconv.Atoi(fields[1])
		if err != nil {
			__antithesis_instrumentation__.Notify(32599)
			return errors.Wrapf(err, "failed to parse parent schema id %s", fields[1])
		} else {
			__antithesis_instrumentation__.Notify(32600)
		}
		__antithesis_instrumentation__.Notify(32595)
		id, err := strconv.Atoi(fields[3])
		if err != nil {
			__antithesis_instrumentation__.Notify(32601)
			if fields[3] == "NULL" {
				__antithesis_instrumentation__.Notify(32602)
				id = int(descpb.InvalidID)
			} else {
				__antithesis_instrumentation__.Notify(32603)
				return errors.Wrapf(err, "failed to parse id %s", fields[3])
			}
		} else {
			__antithesis_instrumentation__.Notify(32604)
		}
		__antithesis_instrumentation__.Notify(32596)

		namespaceTable = append(namespaceTable, doctor.NamespaceTableRow{
			NameInfo: descpb.NameInfo{
				ParentID: descpb.ID(parID), ParentSchemaID: descpb.ID(parSchemaID), Name: fields[2],
			},
			ID: int64(id),
		})
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(32605)
		return nil, nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(32606)
	}
	__antithesis_instrumentation__.Notify(32577)

	jobsTable = make(doctor.JobsTable, 0)
	if err := slurp(zipDirPath, "system.jobs.txt", func(row string) error {
		__antithesis_instrumentation__.Notify(32607)
		fields := strings.Fields(row)
		md := jobs.JobMetadata{}
		md.Status = jobs.Status(fields[1])

		id, err := strconv.Atoi(fields[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(32613)
			return errors.Wrapf(err, "failed to parse job id %s", fields[0])
		} else {
			__antithesis_instrumentation__.Notify(32614)
		}
		__antithesis_instrumentation__.Notify(32608)
		md.ID = jobspb.JobID(id)

		last := len(fields) - 1
		payloadBytes, err := hx.DecodeString(fields[last-1])
		if err != nil {
			__antithesis_instrumentation__.Notify(32615)
			return errors.Wrapf(err, "job %d: failed to decode hex payload", id)
		} else {
			__antithesis_instrumentation__.Notify(32616)
		}
		__antithesis_instrumentation__.Notify(32609)
		md.Payload = &jobspb.Payload{}
		if err := protoutil.Unmarshal(payloadBytes, md.Payload); err != nil {
			__antithesis_instrumentation__.Notify(32617)
			return errors.Wrap(err, "failed unmarshalling job payload")
		} else {
			__antithesis_instrumentation__.Notify(32618)
		}
		__antithesis_instrumentation__.Notify(32610)
		progressBytes, err := hx.DecodeString(fields[last])
		if err != nil {
			__antithesis_instrumentation__.Notify(32619)
			return errors.Wrapf(err, "job %d: failed to decode hex progress", id)
		} else {
			__antithesis_instrumentation__.Notify(32620)
		}
		__antithesis_instrumentation__.Notify(32611)
		md.Progress = &jobspb.Progress{}
		if err := protoutil.Unmarshal(progressBytes, md.Progress); err != nil {
			__antithesis_instrumentation__.Notify(32621)
			return errors.Wrap(err, "failed unmarshalling job progress")
		} else {
			__antithesis_instrumentation__.Notify(32622)
		}
		__antithesis_instrumentation__.Notify(32612)

		jobsTable = append(jobsTable, md)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(32623)
		return nil, nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(32624)
	}
	__antithesis_instrumentation__.Notify(32578)

	return descTable, namespaceTable, jobsTable, nil
}

func slurp(zipDirPath string, fileName string, tableMapFn func(row string) error) error {
	__antithesis_instrumentation__.Notify(32625)
	filePath := path.Join(zipDirPath, fileName)

	_, err := os.Stat(filePath + ".err.txt")
	if err == nil {
		__antithesis_instrumentation__.Notify(32629)

		fmt.Printf("WARNING: errors occurred during the production of %s, contents may be missing or incomplete.\n", fileName)
	} else {
		__antithesis_instrumentation__.Notify(32630)
		if !errors.Is(err, os.ErrNotExist) {
			__antithesis_instrumentation__.Notify(32631)

			return err
		} else {
			__antithesis_instrumentation__.Notify(32632)
		}
	}
	__antithesis_instrumentation__.Notify(32626)

	f, err := os.Open(filePath)
	if err != nil {
		__antithesis_instrumentation__.Notify(32633)
		return err
	} else {
		__antithesis_instrumentation__.Notify(32634)
	}
	__antithesis_instrumentation__.Notify(32627)
	defer f.Close()
	if debugCtx.verbose {
		__antithesis_instrumentation__.Notify(32635)
		fmt.Println("reading " + filePath)
	} else {
		__antithesis_instrumentation__.Notify(32636)
	}
	__antithesis_instrumentation__.Notify(32628)
	return tableMap(f, tableMapFn)
}

func tableMap(in io.Reader, fn func(string) error) error {
	__antithesis_instrumentation__.Notify(32637)
	firstLine := true
	sc := bufio.NewScanner(in)

	sc.Buffer(make([]byte, 64*1024), 50*1024*1024)
	for sc.Scan() {
		__antithesis_instrumentation__.Notify(32639)
		if firstLine {
			__antithesis_instrumentation__.Notify(32641)
			firstLine = false
			continue
		} else {
			__antithesis_instrumentation__.Notify(32642)
		}
		__antithesis_instrumentation__.Notify(32640)
		if err := fn(sc.Text()); err != nil {
			__antithesis_instrumentation__.Notify(32643)
			return err
		} else {
			__antithesis_instrumentation__.Notify(32644)
		}
	}
	__antithesis_instrumentation__.Notify(32638)
	return sc.Err()
}

func selectRowsMap(
	conn clisqlclient.Conn, stmt string, vals []driver.Value, fn func([]driver.Value) error,
) error {
	__antithesis_instrumentation__.Notify(32645)
	rows, err := conn.Query(context.Background(), stmt)
	if err != nil {
		__antithesis_instrumentation__.Notify(32648)
		return errors.Wrapf(err, "query '%s'", stmt)
	} else {
		__antithesis_instrumentation__.Notify(32649)
	}
	__antithesis_instrumentation__.Notify(32646)
	for {
		__antithesis_instrumentation__.Notify(32650)
		var err error
		if err = rows.Next(vals); err == io.EOF {
			__antithesis_instrumentation__.Notify(32653)
			break
		} else {
			__antithesis_instrumentation__.Notify(32654)
		}
		__antithesis_instrumentation__.Notify(32651)
		if err != nil {
			__antithesis_instrumentation__.Notify(32655)
			return err
		} else {
			__antithesis_instrumentation__.Notify(32656)
		}
		__antithesis_instrumentation__.Notify(32652)
		if err := fn(vals); err != nil {
			__antithesis_instrumentation__.Notify(32657)
			return err
		} else {
			__antithesis_instrumentation__.Notify(32658)
		}
	}
	__antithesis_instrumentation__.Notify(32647)
	return nil
}
