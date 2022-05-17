package cliccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"encoding/csv"
	gohex "encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/blobs"
	"github.com/cockroachdb/cockroach/pkg/ccl/backupccl"
	"github.com/cockroachdb/cockroach/pkg/ccl/storageccl"
	"github.com/cockroachdb/cockroach/pkg/cli"
	"github.com/cockroachdb/cockroach/pkg/cli/clierrorplus"
	"github.com/cockroachdb/cockroach/pkg/cli/cliflags"
	"github.com/cockroachdb/cockroach/pkg/cli/clisqlexec"
	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/cloud/nodelocal"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil/pgdate"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

const (
	backupOptRevisionHistory = "revision_history"
)

type key struct {
	rawByte []byte
	typ     string
}

func (k *key) String() string {
	__antithesis_instrumentation__.Notify(19165)
	return string(k.rawByte)
}

func (k *key) Type() string {
	__antithesis_instrumentation__.Notify(19166)
	return k.typ
}

func (k *key) setType(v string) (string, error) {
	__antithesis_instrumentation__.Notify(19167)
	i := strings.IndexByte(v, ':')
	if i == -1 {
		__antithesis_instrumentation__.Notify(19169)
		return "", errors.Newf("no format specified in start key %s", v)
	} else {
		__antithesis_instrumentation__.Notify(19170)
	}
	__antithesis_instrumentation__.Notify(19168)
	k.typ = v[:i]
	return v[i+1:], nil
}

func (k *key) Set(v string) error {
	__antithesis_instrumentation__.Notify(19171)
	v, err := k.setType(v)
	if err != nil {
		__antithesis_instrumentation__.Notify(19174)
		return err
	} else {
		__antithesis_instrumentation__.Notify(19175)
	}
	__antithesis_instrumentation__.Notify(19172)
	switch k.typ {
	case "hex":
		__antithesis_instrumentation__.Notify(19176)
		b, err := gohex.DecodeString(v)
		if err != nil {
			__antithesis_instrumentation__.Notify(19183)
			return err
		} else {
			__antithesis_instrumentation__.Notify(19184)
		}
		__antithesis_instrumentation__.Notify(19177)
		k.rawByte = b
	case "raw":
		__antithesis_instrumentation__.Notify(19178)
		s, err := strconv.Unquote(`"` + v + `"`)
		if err != nil {
			__antithesis_instrumentation__.Notify(19185)
			return errors.Wrapf(err, "invalid argument %q", v)
		} else {
			__antithesis_instrumentation__.Notify(19186)
		}
		__antithesis_instrumentation__.Notify(19179)
		k.rawByte = []byte(s)
	case "bytekey":
		__antithesis_instrumentation__.Notify(19180)
		s, err := strconv.Unquote(`"` + v + `"`)
		if err != nil {
			__antithesis_instrumentation__.Notify(19187)
			return errors.Wrapf(err, "invalid argument %q", v)
		} else {
			__antithesis_instrumentation__.Notify(19188)
		}
		__antithesis_instrumentation__.Notify(19181)
		k.rawByte = []byte(s)
	default:
		__antithesis_instrumentation__.Notify(19182)
	}
	__antithesis_instrumentation__.Notify(19173)
	return nil
}

var debugBackupArgs struct {
	externalIODir string

	exportTableName string
	readTime        string
	destination     string
	format          string
	nullas          string
	maxRows         int
	startKey        key
	withRevisions   bool

	rowCount int
}

func setDebugContextDefault() {
	__antithesis_instrumentation__.Notify(19189)
	debugBackupArgs.externalIODir = ""
	debugBackupArgs.exportTableName = ""
	debugBackupArgs.readTime = ""
	debugBackupArgs.destination = ""
	debugBackupArgs.format = "csv"
	debugBackupArgs.nullas = "null"
	debugBackupArgs.maxRows = 0
	debugBackupArgs.startKey = key{}
	debugBackupArgs.rowCount = 0
	debugBackupArgs.withRevisions = false
}

func init() {

	showCmd := &cobra.Command{
		Use:   "show <backup_path>",
		Short: "show backup summary",
		Long:  "Shows summary of meta information about a SQL backup.",
		Args:  cobra.ExactArgs(1),
		RunE:  clierrorplus.MaybeDecorateError(runShowCmd),
	}

	listBackupsCmd := &cobra.Command{
		Use:   "list-backups <collection_path>",
		Short: "show backups in collection",
		Long:  "Shows full backup paths in a backup collection.",
		Args:  cobra.ExactArgs(1),
		RunE:  clierrorplus.MaybeDecorateError(runListBackupsCmd),
	}

	listIncrementalCmd := &cobra.Command{
		Use:   "list-incremental <backup_path>",
		Short: "show incremental backups",
		Long:  "Shows incremental chain of a SQL backup.",
		Args:  cobra.ExactArgs(1),
		RunE:  clierrorplus.MaybeDecorateError(runListIncrementalCmd),
	}

	exportDataCmd := &cobra.Command{
		Use:   "export <backup_path>",
		Short: "export table data from a backup",
		Long:  "export table data from a backup, requires specifying --table to export data from",
		Args:  cobra.MinimumNArgs(1),
		RunE:  clierrorplus.MaybeDecorateError(runExportDataCmd),
	}

	backupCmds := &cobra.Command{
		Use:   "backup [command]",
		Short: "debug backups",
		Long:  "Shows information about a SQL backup.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.UsageAndErr(cmd, args)
		},

		Hidden: true,
	}

	backupFlags := backupCmds.Flags()
	backupFlags.StringVarP(
		&debugBackupArgs.externalIODir,
		cliflags.ExternalIODir.Name,
		cliflags.ExternalIODir.Shorthand,
		"",
		cliflags.ExternalIODir.Usage())

	exportDataCmd.Flags().StringVarP(
		&debugBackupArgs.exportTableName,
		cliflags.ExportTableTarget.Name,
		cliflags.ExportTableTarget.Shorthand,
		"",
		cliflags.ExportTableTarget.Usage())

	exportDataCmd.Flags().StringVarP(
		&debugBackupArgs.readTime,
		cliflags.ReadTime.Name,
		cliflags.ReadTime.Shorthand,
		"",
		cliflags.ReadTime.Usage())

	exportDataCmd.Flags().StringVarP(
		&debugBackupArgs.destination,
		cliflags.ExportDestination.Name,
		cliflags.ExportDestination.Shorthand,
		"",
		cliflags.ExportDestination.Usage())

	exportDataCmd.Flags().StringVarP(
		&debugBackupArgs.format,
		cliflags.ExportTableFormat.Name,
		cliflags.ExportTableFormat.Shorthand,
		"csv",
		cliflags.ExportTableFormat.Usage())

	exportDataCmd.Flags().StringVarP(
		&debugBackupArgs.nullas,
		cliflags.ExportCSVNullas.Name,
		cliflags.ExportCSVNullas.Shorthand,
		"null",
		cliflags.ExportCSVNullas.Usage())

	exportDataCmd.Flags().IntVar(
		&debugBackupArgs.maxRows,
		cliflags.MaxRows.Name,
		0,
		cliflags.MaxRows.Usage())

	exportDataCmd.Flags().Var(
		&debugBackupArgs.startKey,
		cliflags.StartKey.Name,
		cliflags.StartKey.Usage())

	exportDataCmd.Flags().BoolVar(
		&debugBackupArgs.withRevisions,
		cliflags.ExportRevisions.Name,
		false,
		cliflags.ExportRevisions.Usage())

	exportDataCmd.Flags().StringVarP(
		&debugBackupArgs.readTime,
		cliflags.ExportRevisionsUpTo.Name,
		cliflags.ExportRevisionsUpTo.Shorthand,
		"",
		cliflags.ExportRevisionsUpTo.Usage())

	backupSubCmds := []*cobra.Command{
		showCmd,
		listBackupsCmd,
		listIncrementalCmd,
		exportDataCmd,
	}

	for _, cmd := range backupSubCmds {
		backupCmds.AddCommand(cmd)
		cmd.Flags().AddFlagSet(backupFlags)
	}
	cli.DebugCmd.AddCommand(backupCmds)
}

func newBlobFactory(ctx context.Context, dialing roachpb.NodeID) (blobs.BlobClient, error) {
	__antithesis_instrumentation__.Notify(19190)
	if dialing != 0 {
		__antithesis_instrumentation__.Notify(19193)
		return nil, errors.Errorf("accessing node %d during nodelocal access is unsupported for CLI inspection; only local access is supported with nodelocal://self", dialing)
	} else {
		__antithesis_instrumentation__.Notify(19194)
	}
	__antithesis_instrumentation__.Notify(19191)
	if debugBackupArgs.externalIODir == "" {
		__antithesis_instrumentation__.Notify(19195)
		debugBackupArgs.externalIODir = filepath.Join(server.DefaultStorePath, "extern")
	} else {
		__antithesis_instrumentation__.Notify(19196)
	}
	__antithesis_instrumentation__.Notify(19192)
	return blobs.NewLocalClient(debugBackupArgs.externalIODir)
}

func externalStorageFromURIFactory(
	ctx context.Context, uri string, user security.SQLUsername,
) (cloud.ExternalStorage, error) {
	__antithesis_instrumentation__.Notify(19197)
	defaultSettings := &cluster.Settings{}
	defaultSettings.SV.Init(ctx, nil)
	return cloud.ExternalStorageFromURI(ctx, uri, base.ExternalIODirConfig{},
		defaultSettings, newBlobFactory, user, nil, nil, nil)
}

func getManifestFromURI(ctx context.Context, path string) (backupccl.BackupManifest, error) {
	__antithesis_instrumentation__.Notify(19198)

	if !strings.Contains(path, "://") {
		__antithesis_instrumentation__.Notify(19201)
		path = nodelocal.MakeLocalStorageURI(path)
	} else {
		__antithesis_instrumentation__.Notify(19202)
	}
	__antithesis_instrumentation__.Notify(19199)

	backupManifest, _, err := backupccl.ReadBackupManifestFromURI(ctx, nil, path, security.RootUserName(),
		externalStorageFromURIFactory, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(19203)
		return backupccl.BackupManifest{}, err
	} else {
		__antithesis_instrumentation__.Notify(19204)
	}
	__antithesis_instrumentation__.Notify(19200)
	return backupManifest, nil
}

func runShowCmd(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(19205)

	path := args[0]
	ctx := context.Background()
	desc, err := getManifestFromURI(ctx, path)
	if err != nil {
		__antithesis_instrumentation__.Notify(19208)
		return errors.Wrapf(err, "fetching backup manifest")
	} else {
		__antithesis_instrumentation__.Notify(19209)
	}
	__antithesis_instrumentation__.Notify(19206)

	var meta = backupMetaDisplayMsg(desc)
	jsonBytes, err := json.MarshalIndent(meta, "", "\t")
	if err != nil {
		__antithesis_instrumentation__.Notify(19210)
		return errors.Wrapf(err, "marshall backup manifest")
	} else {
		__antithesis_instrumentation__.Notify(19211)
	}
	__antithesis_instrumentation__.Notify(19207)
	s := string(jsonBytes)
	fmt.Println(s)
	return nil
}

func runListBackupsCmd(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(19212)

	path := args[0]
	if !strings.Contains(path, "://") {
		__antithesis_instrumentation__.Notify(19217)
		path = nodelocal.MakeLocalStorageURI(path)
	} else {
		__antithesis_instrumentation__.Notify(19218)
	}
	__antithesis_instrumentation__.Notify(19213)
	ctx := context.Background()
	store, err := externalStorageFromURIFactory(ctx, path, security.RootUserName())
	if err != nil {
		__antithesis_instrumentation__.Notify(19219)
		return errors.Wrapf(err, "connect to external storage")
	} else {
		__antithesis_instrumentation__.Notify(19220)
	}
	__antithesis_instrumentation__.Notify(19214)
	defer store.Close()

	backupPaths, err := backupccl.ListFullBackupsInCollection(ctx, store)
	if err != nil {
		__antithesis_instrumentation__.Notify(19221)
		return errors.Wrapf(err, "list full backups in collection")
	} else {
		__antithesis_instrumentation__.Notify(19222)
	}
	__antithesis_instrumentation__.Notify(19215)

	cols := []string{"path"}
	rows := make([][]string, 0)
	for _, backupPath := range backupPaths {
		__antithesis_instrumentation__.Notify(19223)
		rows = append(rows, []string{"." + backupPath})
	}
	__antithesis_instrumentation__.Notify(19216)
	rowSliceIter := clisqlexec.NewRowSliceIter(rows, "l")
	return cli.PrintQueryOutput(os.Stdout, cols, rowSliceIter)
}

func runListIncrementalCmd(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(19224)

	path := args[0]
	if !strings.Contains(path, "://") {
		__antithesis_instrumentation__.Notify(19234)
		path = nodelocal.MakeLocalStorageURI(path)
	} else {
		__antithesis_instrumentation__.Notify(19235)
	}
	__antithesis_instrumentation__.Notify(19225)

	basepath, subdir := backupccl.CollectionAndSubdir(path, "")

	uri, err := url.Parse(basepath)
	if err != nil {
		__antithesis_instrumentation__.Notify(19236)
		return err
	} else {
		__antithesis_instrumentation__.Notify(19237)
	}
	__antithesis_instrumentation__.Notify(19226)

	ctx := context.Background()

	priorPaths := []string{backupccl.JoinURLPath(
		strings.TrimSuffix(
			uri.Path, string(backupccl.URLSeparator)+backupccl.DefaultIncrementalsSubdir),
		subdir)}

	oldIncURI := *uri
	oldIncURI.Path = backupccl.JoinURLPath(oldIncURI.Path, subdir)
	baseStore, err := externalStorageFromURIFactory(ctx, oldIncURI.String(), security.RootUserName())
	if err != nil {
		__antithesis_instrumentation__.Notify(19238)
		return errors.Wrapf(err, "connect to external storage")
	} else {
		__antithesis_instrumentation__.Notify(19239)
	}
	__antithesis_instrumentation__.Notify(19227)
	defer baseStore.Close()

	oldIncPaths, err := backupccl.FindPriorBackups(ctx, baseStore, backupccl.OmitManifest)
	if err != nil {
		__antithesis_instrumentation__.Notify(19240)
		return err
	} else {
		__antithesis_instrumentation__.Notify(19241)
	}
	__antithesis_instrumentation__.Notify(19228)
	for _, path := range oldIncPaths {
		__antithesis_instrumentation__.Notify(19242)
		priorPaths = append(priorPaths, backupccl.JoinURLPath(oldIncURI.Path, path))
	}
	__antithesis_instrumentation__.Notify(19229)

	newIncURI := *uri
	newIncURI.Path = backupccl.JoinURLPath(newIncURI.Path, backupccl.DefaultIncrementalsSubdir, subdir)
	incStore, err := externalStorageFromURIFactory(ctx, newIncURI.String(), security.RootUserName())
	if err != nil {
		__antithesis_instrumentation__.Notify(19243)
		return errors.Wrapf(err, "connect to external storage")
	} else {
		__antithesis_instrumentation__.Notify(19244)
	}
	__antithesis_instrumentation__.Notify(19230)
	defer incStore.Close()

	newIncPaths, err := backupccl.FindPriorBackups(ctx, incStore, backupccl.OmitManifest)
	if err != nil {
		__antithesis_instrumentation__.Notify(19245)
		return err
	} else {
		__antithesis_instrumentation__.Notify(19246)
	}
	__antithesis_instrumentation__.Notify(19231)
	for _, path := range newIncPaths {
		__antithesis_instrumentation__.Notify(19247)
		priorPaths = append(priorPaths, backupccl.JoinURLPath(newIncURI.Path, path))
	}
	__antithesis_instrumentation__.Notify(19232)

	stores := make([]cloud.ExternalStorage, len(priorPaths))
	rows := make([][]string, 0)
	for i, path := range priorPaths {
		__antithesis_instrumentation__.Notify(19248)
		uri.Path = path
		stores[i], err = externalStorageFromURIFactory(ctx, uri.String(), security.RootUserName())
		if err != nil {
			__antithesis_instrumentation__.Notify(19252)
			return errors.Wrapf(err, "connect to external storage")
		} else {
			__antithesis_instrumentation__.Notify(19253)
		}
		__antithesis_instrumentation__.Notify(19249)
		defer stores[i].Close()
		manifest, _, err := backupccl.ReadBackupManifestFromStore(ctx, nil, stores[i], nil)
		if err != nil {
			__antithesis_instrumentation__.Notify(19254)
			return err
		} else {
			__antithesis_instrumentation__.Notify(19255)
		}
		__antithesis_instrumentation__.Notify(19250)
		startTime := manifest.StartTime.GoTime().Format(time.RFC3339)
		endTime := manifest.EndTime.GoTime().Format(time.RFC3339)
		if i == 0 {
			__antithesis_instrumentation__.Notify(19256)
			startTime = "-"
		} else {
			__antithesis_instrumentation__.Notify(19257)
		}
		__antithesis_instrumentation__.Notify(19251)
		newRow := []string{uri.Path, startTime, endTime}
		rows = append(rows, newRow)
	}
	__antithesis_instrumentation__.Notify(19233)
	cols := []string{"path", "start time", "end time"}
	rowSliceIter := clisqlexec.NewRowSliceIter(rows, "lll")
	return cli.PrintQueryOutput(os.Stdout, cols, rowSliceIter)
}

func runExportDataCmd(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(19258)
	if debugBackupArgs.exportTableName == "" {
		__antithesis_instrumentation__.Notify(19265)
		return errors.New("export data requires table name specified by --table flag")
	} else {
		__antithesis_instrumentation__.Notify(19266)
	}
	__antithesis_instrumentation__.Notify(19259)
	fullyQualifiedTableName := strings.ToLower(debugBackupArgs.exportTableName)
	manifestPaths := args
	ctx := context.Background()
	manifests := make([]backupccl.BackupManifest, 0, len(manifestPaths))
	for _, path := range manifestPaths {
		__antithesis_instrumentation__.Notify(19267)
		manifest, err := getManifestFromURI(ctx, path)
		if err != nil {
			__antithesis_instrumentation__.Notify(19269)
			return errors.Wrapf(err, "fetching backup manifests from %s", path)
		} else {
			__antithesis_instrumentation__.Notify(19270)
		}
		__antithesis_instrumentation__.Notify(19268)
		manifests = append(manifests, manifest)
	}
	__antithesis_instrumentation__.Notify(19260)

	if debugBackupArgs.withRevisions && func() bool {
		__antithesis_instrumentation__.Notify(19271)
		return manifests[0].MVCCFilter != backupccl.MVCCFilter_All == true
	}() == true {
		__antithesis_instrumentation__.Notify(19272)
		return errors.WithHintf(
			errors.Newf("invalid flag: %s", cliflags.ExportRevisions.Name),
			"requires backup created with %q", backupOptRevisionHistory,
		)
	} else {
		__antithesis_instrumentation__.Notify(19273)
	}
	__antithesis_instrumentation__.Notify(19261)

	endTime, err := evalAsOfTimestamp(debugBackupArgs.readTime, manifests)
	if err != nil {
		__antithesis_instrumentation__.Notify(19274)
		return errors.Wrapf(err, "eval as of timestamp %s", debugBackupArgs.readTime)
	} else {
		__antithesis_instrumentation__.Notify(19275)
	}
	__antithesis_instrumentation__.Notify(19262)

	codec := keys.TODOSQLCodec
	entry, err := backupccl.MakeBackupTableEntry(
		ctx,
		fullyQualifiedTableName,
		manifests,
		endTime,
		security.RootUserName(),
		codec,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(19276)
		return errors.Wrapf(err, "fetching entry")
	} else {
		__antithesis_instrumentation__.Notify(19277)
	}
	__antithesis_instrumentation__.Notify(19263)

	if err = showData(ctx, entry, endTime, codec); err != nil {
		__antithesis_instrumentation__.Notify(19278)
		return errors.Wrapf(err, "show data")
	} else {
		__antithesis_instrumentation__.Notify(19279)
	}
	__antithesis_instrumentation__.Notify(19264)
	return nil
}

func evalAsOfTimestamp(
	readTime string, manifests []backupccl.BackupManifest,
) (hlc.Timestamp, error) {
	__antithesis_instrumentation__.Notify(19280)
	if readTime == "" {
		__antithesis_instrumentation__.Notify(19284)
		return manifests[len(manifests)-1].EndTime, nil
	} else {
		__antithesis_instrumentation__.Notify(19285)
	}
	__antithesis_instrumentation__.Notify(19281)
	var err error

	if ts, _, err := pgdate.ParseTimestampWithoutTimezone(timeutil.Now(), pgdate.DateStyle{Order: pgdate.Order_MDY}, readTime); err == nil {
		__antithesis_instrumentation__.Notify(19286)
		readTS := hlc.Timestamp{WallTime: ts.UnixNano()}
		return readTS, nil
	} else {
		__antithesis_instrumentation__.Notify(19287)
	}
	__antithesis_instrumentation__.Notify(19282)

	if dec, _, err := apd.NewFromString(readTime); err == nil {
		__antithesis_instrumentation__.Notify(19288)
		if readTS, err := tree.DecimalToHLC(dec); err == nil {
			__antithesis_instrumentation__.Notify(19289)
			return readTS, nil
		} else {
			__antithesis_instrumentation__.Notify(19290)
		}
	} else {
		__antithesis_instrumentation__.Notify(19291)
	}
	__antithesis_instrumentation__.Notify(19283)
	err = errors.Newf("value %s is neither timestamp nor decimal", readTime)
	return hlc.Timestamp{}, err
}

func showData(
	ctx context.Context, entry backupccl.BackupTableEntry, endTime hlc.Timestamp, codec keys.SQLCodec,
) error {
	__antithesis_instrumentation__.Notify(19292)

	buf := bytes.NewBuffer([]byte{})
	var writer *csv.Writer
	if debugBackupArgs.format != "csv" {
		__antithesis_instrumentation__.Notify(19299)
		return errors.Newf("only exporting to csv format is supported")
	} else {
		__antithesis_instrumentation__.Notify(19300)
	}
	__antithesis_instrumentation__.Notify(19293)
	if debugBackupArgs.destination == "" {
		__antithesis_instrumentation__.Notify(19301)
		writer = csv.NewWriter(os.Stdout)
	} else {
		__antithesis_instrumentation__.Notify(19302)
		writer = csv.NewWriter(buf)
	}
	__antithesis_instrumentation__.Notify(19294)

	rf, err := makeRowFetcher(ctx, entry, codec)
	if err != nil {
		__antithesis_instrumentation__.Notify(19303)
		return errors.Wrapf(err, "make row fetcher")
	} else {
		__antithesis_instrumentation__.Notify(19304)
	}
	__antithesis_instrumentation__.Notify(19295)
	defer rf.Close(ctx)

	if debugBackupArgs.withRevisions {
		__antithesis_instrumentation__.Notify(19305)
		startT := entry.LastSchemaChangeTime.GoTime().UTC()
		endT := endTime.GoTime().UTC()
		fmt.Fprintf(os.Stderr, "DETECTED SCHEMA CHANGE AT %s, ONLY SHOWING UPDATES IN RANGE [%s, %s]\n", startT, startT, endT)
	} else {
		__antithesis_instrumentation__.Notify(19306)
	}
	__antithesis_instrumentation__.Notify(19296)

	for _, files := range entry.Files {
		__antithesis_instrumentation__.Notify(19307)
		if err := processEntryFiles(ctx, rf, files, entry.Span, entry.LastSchemaChangeTime, endTime, writer); err != nil {
			__antithesis_instrumentation__.Notify(19309)
			return err
		} else {
			__antithesis_instrumentation__.Notify(19310)
		}
		__antithesis_instrumentation__.Notify(19308)
		if debugBackupArgs.maxRows != 0 && func() bool {
			__antithesis_instrumentation__.Notify(19311)
			return debugBackupArgs.rowCount >= debugBackupArgs.maxRows == true
		}() == true {
			__antithesis_instrumentation__.Notify(19312)
			break
		} else {
			__antithesis_instrumentation__.Notify(19313)
		}
	}
	__antithesis_instrumentation__.Notify(19297)

	if debugBackupArgs.destination != "" {
		__antithesis_instrumentation__.Notify(19314)
		dir, file := filepath.Split(debugBackupArgs.destination)
		store, err := externalStorageFromURIFactory(ctx, dir, security.RootUserName())
		if err != nil {
			__antithesis_instrumentation__.Notify(19317)
			return errors.Wrapf(err, "unable to open store to write files: %s", debugBackupArgs.destination)
		} else {
			__antithesis_instrumentation__.Notify(19318)
		}
		__antithesis_instrumentation__.Notify(19315)
		if err = cloud.WriteFile(ctx, store, file, bytes.NewReader(buf.Bytes())); err != nil {
			__antithesis_instrumentation__.Notify(19319)
			_ = store.Close()
			return err
		} else {
			__antithesis_instrumentation__.Notify(19320)
		}
		__antithesis_instrumentation__.Notify(19316)
		return store.Close()
	} else {
		__antithesis_instrumentation__.Notify(19321)
	}
	__antithesis_instrumentation__.Notify(19298)
	return nil
}

func makeIters(
	ctx context.Context, files backupccl.EntryFiles,
) ([]storage.SimpleMVCCIterator, func() error, error) {
	__antithesis_instrumentation__.Notify(19322)
	iters := make([]storage.SimpleMVCCIterator, len(files))
	dirStorage := make([]cloud.ExternalStorage, len(files))
	for i, file := range files {
		__antithesis_instrumentation__.Notify(19325)
		var err error
		clusterSettings := cluster.MakeClusterSettings()
		dirStorage[i], err = cloud.MakeExternalStorage(ctx, file.Dir, base.ExternalIODirConfig{},
			clusterSettings, newBlobFactory, nil, nil, nil)
		if err != nil {
			__antithesis_instrumentation__.Notify(19327)
			return nil, nil, errors.Wrapf(err, "making external storage")
		} else {
			__antithesis_instrumentation__.Notify(19328)
		}
		__antithesis_instrumentation__.Notify(19326)

		iters[i], err = storageccl.ExternalSSTReader(ctx, dirStorage[i], file.Path, nil)
		if err != nil {
			__antithesis_instrumentation__.Notify(19329)
			return nil, nil, errors.Wrapf(err, "fetching sst reader")
		} else {
			__antithesis_instrumentation__.Notify(19330)
		}
	}
	__antithesis_instrumentation__.Notify(19323)

	cleanup := func() error {
		__antithesis_instrumentation__.Notify(19331)
		for _, iter := range iters {
			__antithesis_instrumentation__.Notify(19334)
			iter.Close()
		}
		__antithesis_instrumentation__.Notify(19332)
		for _, dir := range dirStorage {
			__antithesis_instrumentation__.Notify(19335)
			if err := dir.Close(); err != nil {
				__antithesis_instrumentation__.Notify(19336)
				return err
			} else {
				__antithesis_instrumentation__.Notify(19337)
			}
		}
		__antithesis_instrumentation__.Notify(19333)
		return nil
	}
	__antithesis_instrumentation__.Notify(19324)
	return iters, cleanup, nil
}

func makeRowFetcher(
	ctx context.Context, entry backupccl.BackupTableEntry, codec keys.SQLCodec,
) (row.Fetcher, error) {
	__antithesis_instrumentation__.Notify(19338)
	colIDs := entry.Desc.PublicColumnIDs()
	if debugBackupArgs.withRevisions {
		__antithesis_instrumentation__.Notify(19342)
		colIDs = append(colIDs, colinfo.MVCCTimestampColumnID)
	} else {
		__antithesis_instrumentation__.Notify(19343)
	}
	__antithesis_instrumentation__.Notify(19339)

	var spec descpb.IndexFetchSpec
	if err := rowenc.InitIndexFetchSpec(&spec, codec, entry.Desc, entry.Desc.GetPrimaryIndex(), colIDs); err != nil {
		__antithesis_instrumentation__.Notify(19344)
		return row.Fetcher{}, err
	} else {
		__antithesis_instrumentation__.Notify(19345)
	}
	__antithesis_instrumentation__.Notify(19340)

	var rf row.Fetcher
	if err := rf.Init(
		ctx,
		false,
		descpb.ScanLockingStrength_FOR_NONE,
		descpb.ScanLockingWaitPolicy_BLOCK,
		0,
		&tree.DatumAlloc{},
		nil,
		&spec,
	); err != nil {
		__antithesis_instrumentation__.Notify(19346)
		return rf, err
	} else {
		__antithesis_instrumentation__.Notify(19347)
	}
	__antithesis_instrumentation__.Notify(19341)
	return rf, nil
}

func processEntryFiles(
	ctx context.Context,
	rf row.Fetcher,
	files backupccl.EntryFiles,
	span roachpb.Span,
	startTime hlc.Timestamp,
	endTime hlc.Timestamp,
	writer *csv.Writer,
) (err error) {
	__antithesis_instrumentation__.Notify(19348)

	iters, cleanup, err := makeIters(ctx, files)
	defer func() {
		__antithesis_instrumentation__.Notify(19354)
		if cleanupErr := cleanup(); err == nil {
			__antithesis_instrumentation__.Notify(19355)
			err = cleanupErr
		} else {
			__antithesis_instrumentation__.Notify(19356)
		}
	}()
	__antithesis_instrumentation__.Notify(19349)
	if err != nil {
		__antithesis_instrumentation__.Notify(19357)
		return errors.Wrapf(err, "make iters")
	} else {
		__antithesis_instrumentation__.Notify(19358)
	}
	__antithesis_instrumentation__.Notify(19350)

	iter := storage.MakeMultiIterator(iters)
	defer iter.Close()

	startKeyMVCC, endKeyMVCC := storage.MVCCKey{Key: span.Key}, storage.MVCCKey{Key: span.EndKey}
	if len(debugBackupArgs.startKey.rawByte) != 0 {
		__antithesis_instrumentation__.Notify(19359)
		if debugBackupArgs.startKey.typ == "bytekey" {
			__antithesis_instrumentation__.Notify(19360)
			startKeyMVCC.Key = append(startKeyMVCC.Key, debugBackupArgs.startKey.rawByte...)
		} else {
			__antithesis_instrumentation__.Notify(19361)
			startKeyMVCC.Key = roachpb.Key(debugBackupArgs.startKey.rawByte)
		}
	} else {
		__antithesis_instrumentation__.Notify(19362)
	}
	__antithesis_instrumentation__.Notify(19351)
	kvFetcher := row.MakeBackupSSTKVFetcher(startKeyMVCC, endKeyMVCC, iter, startTime, endTime, debugBackupArgs.withRevisions)

	if err := rf.StartScanFrom(ctx, &kvFetcher, false); err != nil {
		__antithesis_instrumentation__.Notify(19363)
		return errors.Wrapf(err, "row fetcher starts scan")
	} else {
		__antithesis_instrumentation__.Notify(19364)
	}
	__antithesis_instrumentation__.Notify(19352)

	for {
		__antithesis_instrumentation__.Notify(19365)
		datums, err := rf.NextRowDecoded(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(19370)
			return errors.Wrapf(err, "decode row")
		} else {
			__antithesis_instrumentation__.Notify(19371)
		}
		__antithesis_instrumentation__.Notify(19366)
		if datums == nil {
			__antithesis_instrumentation__.Notify(19372)
			break
		} else {
			__antithesis_instrumentation__.Notify(19373)
		}
		__antithesis_instrumentation__.Notify(19367)
		rowDisplay := make([]string, datums.Len())
		for i, datum := range datums {
			__antithesis_instrumentation__.Notify(19374)

			if debugBackupArgs.withRevisions && func() bool {
				__antithesis_instrumentation__.Notify(19376)
				return i == datums.Len()-1 == true
			}() == true {
				__antithesis_instrumentation__.Notify(19377)
				approx, err := tree.DecimalToInexactDTimestamp(datum.(*tree.DDecimal))
				if err != nil {
					__antithesis_instrumentation__.Notify(19379)
					return errors.Wrapf(err, "convert datum %s to mvcc timestamp", datum)
				} else {
					__antithesis_instrumentation__.Notify(19380)
				}
				__antithesis_instrumentation__.Notify(19378)
				rowDisplay[i] = approx.UTC().String()
				break
			} else {
				__antithesis_instrumentation__.Notify(19381)
			}
			__antithesis_instrumentation__.Notify(19375)

			if datum == tree.DNull {
				__antithesis_instrumentation__.Notify(19382)
				rowDisplay[i] = debugBackupArgs.nullas
			} else {
				__antithesis_instrumentation__.Notify(19383)
				rowDisplay[i] = datum.String()
			}
		}
		__antithesis_instrumentation__.Notify(19368)
		if err := writer.Write(rowDisplay); err != nil {
			__antithesis_instrumentation__.Notify(19384)
			return err
		} else {
			__antithesis_instrumentation__.Notify(19385)
		}
		__antithesis_instrumentation__.Notify(19369)
		writer.Flush()

		if debugBackupArgs.maxRows != 0 {
			__antithesis_instrumentation__.Notify(19386)
			debugBackupArgs.rowCount++
			if debugBackupArgs.rowCount >= debugBackupArgs.maxRows {
				__antithesis_instrumentation__.Notify(19387)
				break
			} else {
				__antithesis_instrumentation__.Notify(19388)
			}
		} else {
			__antithesis_instrumentation__.Notify(19389)
		}
	}
	__antithesis_instrumentation__.Notify(19353)
	return nil
}

type backupMetaDisplayMsg backupccl.BackupManifest
type backupFileDisplayMsg backupccl.BackupManifest_File

func (f backupFileDisplayMsg) MarshalJSON() ([]byte, error) {
	__antithesis_instrumentation__.Notify(19390)
	fileDisplayMsg := struct {
		Path         string
		Span         string
		DataSize     string
		IndexEntries int64
		Rows         int64
	}{
		Path:         f.Path,
		Span:         fmt.Sprint(f.Span),
		DataSize:     string(humanizeutil.IBytes(f.EntryCounts.DataSize)),
		IndexEntries: f.EntryCounts.IndexEntries,
		Rows:         f.EntryCounts.Rows,
	}
	return json.Marshal(fileDisplayMsg)
}

func (b backupMetaDisplayMsg) MarshalJSON() ([]byte, error) {
	__antithesis_instrumentation__.Notify(19391)

	fileMsg := make([]backupFileDisplayMsg, len(b.Files))
	for i, file := range b.Files {
		__antithesis_instrumentation__.Notify(19394)
		fileMsg[i] = backupFileDisplayMsg(file)
	}
	__antithesis_instrumentation__.Notify(19392)

	displayMsg := struct {
		StartTime           string
		EndTime             string
		DataSize            string
		Rows                int64
		IndexEntries        int64
		FormatVersion       uint32
		ClusterID           uuid.UUID
		NodeID              roachpb.NodeID
		BuildInfo           string
		Files               []backupFileDisplayMsg
		Spans               string
		DatabaseDescriptors map[descpb.ID]string
		TableDescriptors    map[descpb.ID]string
		TypeDescriptors     map[descpb.ID]string
		SchemaDescriptors   map[descpb.ID]string
	}{
		StartTime:           timeutil.Unix(0, b.StartTime.WallTime).Format(time.RFC3339),
		EndTime:             timeutil.Unix(0, b.EndTime.WallTime).Format(time.RFC3339),
		DataSize:            string(humanizeutil.IBytes(b.EntryCounts.DataSize)),
		Rows:                b.EntryCounts.Rows,
		IndexEntries:        b.EntryCounts.IndexEntries,
		FormatVersion:       b.FormatVersion,
		ClusterID:           b.ClusterID,
		NodeID:              b.NodeID,
		BuildInfo:           b.BuildInfo.Short(),
		Files:               fileMsg,
		Spans:               fmt.Sprint(b.Spans),
		DatabaseDescriptors: make(map[descpb.ID]string),
		TableDescriptors:    make(map[descpb.ID]string),
		TypeDescriptors:     make(map[descpb.ID]string),
		SchemaDescriptors:   make(map[descpb.ID]string),
	}

	dbIDToName := make(map[descpb.ID]string)
	schemaIDToFullyQualifiedName := make(map[descpb.ID]string)
	schemaIDToFullyQualifiedName[keys.PublicSchemaIDForBackup] = catconstants.PublicSchemaName
	typeIDToFullyQualifiedName := make(map[descpb.ID]string)
	tableIDToFullyQualifiedName := make(map[descpb.ID]string)

	for i := range b.Descriptors {
		__antithesis_instrumentation__.Notify(19395)
		d := &b.Descriptors[i]
		id := descpb.GetDescriptorID(d)
		tableDesc, databaseDesc, typeDesc, schemaDesc := descpb.FromDescriptor(d)
		if databaseDesc != nil {
			__antithesis_instrumentation__.Notify(19396)
			dbIDToName[id] = descpb.GetDescriptorName(d)
		} else {
			__antithesis_instrumentation__.Notify(19397)
			if schemaDesc != nil {
				__antithesis_instrumentation__.Notify(19398)
				dbName := dbIDToName[schemaDesc.GetParentID()]
				schemaName := descpb.GetDescriptorName(d)
				schemaIDToFullyQualifiedName[id] = dbName + "." + schemaName
			} else {
				__antithesis_instrumentation__.Notify(19399)
				if typeDesc != nil {
					__antithesis_instrumentation__.Notify(19400)
					parentSchema := schemaIDToFullyQualifiedName[typeDesc.GetParentSchemaID()]
					if parentSchema == catconstants.PublicSchemaName {
						__antithesis_instrumentation__.Notify(19402)
						parentSchema = dbIDToName[typeDesc.GetParentID()] + "." + parentSchema
					} else {
						__antithesis_instrumentation__.Notify(19403)
					}
					__antithesis_instrumentation__.Notify(19401)
					typeName := descpb.GetDescriptorName(d)
					typeIDToFullyQualifiedName[id] = parentSchema + "." + typeName
				} else {
					__antithesis_instrumentation__.Notify(19404)
					if tableDesc != nil {
						__antithesis_instrumentation__.Notify(19405)
						tbDesc := tabledesc.NewBuilder(tableDesc).BuildImmutable()
						parentSchema := schemaIDToFullyQualifiedName[tbDesc.GetParentSchemaID()]
						if parentSchema == catconstants.PublicSchemaName {
							__antithesis_instrumentation__.Notify(19407)
							parentSchema = dbIDToName[tableDesc.GetParentID()] + "." + parentSchema
						} else {
							__antithesis_instrumentation__.Notify(19408)
						}
						__antithesis_instrumentation__.Notify(19406)
						tableName := descpb.GetDescriptorName(d)
						tableIDToFullyQualifiedName[id] = parentSchema + "." + tableName
					} else {
						__antithesis_instrumentation__.Notify(19409)
					}
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(19393)
	displayMsg.DatabaseDescriptors = dbIDToName
	displayMsg.TableDescriptors = tableIDToFullyQualifiedName
	displayMsg.SchemaDescriptors = schemaIDToFullyQualifiedName
	displayMsg.TypeDescriptors = typeIDToFullyQualifiedName

	return json.Marshal(displayMsg)
}
