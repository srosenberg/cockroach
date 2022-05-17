package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/featureflag"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/exec"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/errors"
)

type exportNode struct {
	optColumnsSlot

	source planNode

	destination string

	fileNamePattern string
	format          roachpb.IOFileFormat
	chunkRows       int
	chunkSize       int64
	colNames        []string
}

func (e *exportNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(491355)
	panic("exportNode cannot be run in local mode")
}

func (e *exportNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(491356)
	panic("exportNode cannot be run in local mode")
}

func (e *exportNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(491357)
	panic("exportNode cannot be run in local mode")
}

func (e *exportNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(491358)
	e.source.Close(ctx)
}

const (
	exportOptionDelimiter   = "delimiter"
	exportOptionNullAs      = "nullas"
	exportOptionChunkRows   = "chunk_rows"
	exportOptionChunkSize   = "chunk_size"
	exportOptionFileName    = "filename"
	exportOptionCompression = "compression"

	exportChunkSizeDefault = int64(32 << 20)
	exportChunkRowsDefault = 100000

	exportFilePatternPart = "%part%"
	exportGzipCodec       = "gzip"
	exportSnappyCodec     = "snappy"
	csvSuffix             = "csv"
	parquetSuffix         = "parquet"
)

var exportOptionExpectValues = map[string]KVStringOptValidate{
	exportOptionChunkRows:   KVStringOptRequireValue,
	exportOptionDelimiter:   KVStringOptRequireValue,
	exportOptionFileName:    KVStringOptRequireValue,
	exportOptionNullAs:      KVStringOptRequireValue,
	exportOptionCompression: KVStringOptRequireValue,
	exportOptionChunkSize:   KVStringOptRequireValue,
}

var featureExportEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"feature.export.enabled",
	"set to true to enable exports, false to disable; default is true",
	featureflag.FeatureFlagEnabledDefault,
).WithPublic()

func (ef *execFactory) ConstructExport(
	input exec.Node,
	fileName tree.TypedExpr,
	fileSuffix string,
	options []exec.KVOption,
	notNullCols exec.NodeColumnOrdinalSet,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(491359)
	fileSuffix = strings.ToLower(fileSuffix)

	if !featureExportEnabled.Get(&ef.planner.ExecCfg().Settings.SV) {
		__antithesis_instrumentation__.Notify(491374)
		return nil, pgerror.Newf(
			pgcode.OperatorIntervention,
			"feature EXPORT was disabled by the database administrator",
		)
	} else {
		__antithesis_instrumentation__.Notify(491375)
	}
	__antithesis_instrumentation__.Notify(491360)

	if err := featureflag.CheckEnabled(
		ef.planner.EvalContext().Context,
		ef.planner.execCfg,
		featureExportEnabled,
		"EXPORT",
	); err != nil {
		__antithesis_instrumentation__.Notify(491376)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(491377)
	}
	__antithesis_instrumentation__.Notify(491361)

	if !ef.planner.ExtendedEvalContext().TxnIsSingleStmt {
		__antithesis_instrumentation__.Notify(491378)
		return nil, errors.Errorf("EXPORT cannot be used inside a multi-statement transaction")
	} else {
		__antithesis_instrumentation__.Notify(491379)
	}
	__antithesis_instrumentation__.Notify(491362)

	if fileSuffix != csvSuffix && func() bool {
		__antithesis_instrumentation__.Notify(491380)
		return fileSuffix != parquetSuffix == true
	}() == true {
		__antithesis_instrumentation__.Notify(491381)
		return nil, errors.Errorf("unsupported export format: %q", fileSuffix)
	} else {
		__antithesis_instrumentation__.Notify(491382)
	}
	__antithesis_instrumentation__.Notify(491363)

	destinationDatum, err := fileName.Eval(ef.planner.EvalContext())
	if err != nil {
		__antithesis_instrumentation__.Notify(491383)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(491384)
	}
	__antithesis_instrumentation__.Notify(491364)

	destination, ok := destinationDatum.(*tree.DString)
	if !ok {
		__antithesis_instrumentation__.Notify(491385)
		return nil, errors.Errorf("expected string value for the file location")
	} else {
		__antithesis_instrumentation__.Notify(491386)
	}
	__antithesis_instrumentation__.Notify(491365)
	admin, err := ef.planner.HasAdminRole(ef.planner.EvalContext().Context)
	if err != nil {
		__antithesis_instrumentation__.Notify(491387)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(491388)
	}
	__antithesis_instrumentation__.Notify(491366)
	if !admin && func() bool {
		__antithesis_instrumentation__.Notify(491389)
		return !ef.planner.ExecCfg().ExternalIODirConfig.EnableNonAdminImplicitAndArbitraryOutbound == true
	}() == true {
		__antithesis_instrumentation__.Notify(491390)
		conf, err := cloud.ExternalStorageConfFromURI(string(*destination), ef.planner.User())
		if err != nil {
			__antithesis_instrumentation__.Notify(491392)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(491393)
		}
		__antithesis_instrumentation__.Notify(491391)
		if !conf.AccessIsWithExplicitAuth() {
			__antithesis_instrumentation__.Notify(491394)
			panic(pgerror.Newf(
				pgcode.InsufficientPrivilege,
				"only users with the admin role are allowed to EXPORT to the specified URI"))
		} else {
			__antithesis_instrumentation__.Notify(491395)
		}
	} else {
		__antithesis_instrumentation__.Notify(491396)
	}
	__antithesis_instrumentation__.Notify(491367)
	optVals, err := evalStringOptions(ef.planner.EvalContext(), options, exportOptionExpectValues)
	if err != nil {
		__antithesis_instrumentation__.Notify(491397)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(491398)
	}
	__antithesis_instrumentation__.Notify(491368)

	cols := planColumns(input.(planNode))
	colNames := make([]string, len(cols))
	colNullability := make([]bool, len(cols))
	for i, col := range cols {
		__antithesis_instrumentation__.Notify(491399)
		colNames[i] = col.Name
		colNullability[i] = !notNullCols.Contains(i)
	}
	__antithesis_instrumentation__.Notify(491369)

	format := roachpb.IOFileFormat{}
	switch fileSuffix {
	case csvSuffix:
		__antithesis_instrumentation__.Notify(491400)
		csvOpts := roachpb.CSVOptions{}
		if override, ok := optVals[exportOptionDelimiter]; ok {
			__antithesis_instrumentation__.Notify(491405)
			csvOpts.Comma, err = util.GetSingleRune(override)
			if err != nil {
				__antithesis_instrumentation__.Notify(491406)
				return nil, pgerror.New(pgcode.InvalidParameterValue, "invalid delimiter")
			} else {
				__antithesis_instrumentation__.Notify(491407)
			}
		} else {
			__antithesis_instrumentation__.Notify(491408)
		}
		__antithesis_instrumentation__.Notify(491401)
		if override, ok := optVals[exportOptionNullAs]; ok {
			__antithesis_instrumentation__.Notify(491409)
			csvOpts.NullEncoding = &override
		} else {
			__antithesis_instrumentation__.Notify(491410)
		}
		__antithesis_instrumentation__.Notify(491402)
		format.Format = roachpb.IOFileFormat_CSV
		format.Csv = csvOpts
	case parquetSuffix:
		__antithesis_instrumentation__.Notify(491403)
		parquetOpts := roachpb.ParquetOptions{
			ColNullability: colNullability,
		}
		format.Format = roachpb.IOFileFormat_Parquet
		format.Parquet = parquetOpts
	default:
		__antithesis_instrumentation__.Notify(491404)
	}
	__antithesis_instrumentation__.Notify(491370)

	chunkRows := exportChunkRowsDefault
	if override, ok := optVals[exportOptionChunkRows]; ok {
		__antithesis_instrumentation__.Notify(491411)
		chunkRows, err = strconv.Atoi(override)
		if err != nil {
			__antithesis_instrumentation__.Notify(491413)
			return nil, pgerror.WithCandidateCode(err, pgcode.InvalidParameterValue)
		} else {
			__antithesis_instrumentation__.Notify(491414)
		}
		__antithesis_instrumentation__.Notify(491412)
		if chunkRows < 1 {
			__antithesis_instrumentation__.Notify(491415)
			return nil, pgerror.New(pgcode.InvalidParameterValue, "invalid csv chunk rows")
		} else {
			__antithesis_instrumentation__.Notify(491416)
		}
	} else {
		__antithesis_instrumentation__.Notify(491417)
	}
	__antithesis_instrumentation__.Notify(491371)

	chunkSize := exportChunkSizeDefault
	if override, ok := optVals[exportOptionChunkSize]; ok {
		__antithesis_instrumentation__.Notify(491418)
		chunkSize, err = humanizeutil.ParseBytes(override)
		if err != nil {
			__antithesis_instrumentation__.Notify(491420)
			return nil, pgerror.WithCandidateCode(err, pgcode.InvalidParameterValue)
		} else {
			__antithesis_instrumentation__.Notify(491421)
		}
		__antithesis_instrumentation__.Notify(491419)
		if chunkSize < 1 {
			__antithesis_instrumentation__.Notify(491422)
			return nil, pgerror.New(pgcode.InvalidParameterValue, "invalid csv chunk size")
		} else {
			__antithesis_instrumentation__.Notify(491423)
		}
	} else {
		__antithesis_instrumentation__.Notify(491424)
	}
	__antithesis_instrumentation__.Notify(491372)

	var codec roachpb.IOFileFormat_Compression
	if name, ok := optVals[exportOptionCompression]; ok && func() bool {
		__antithesis_instrumentation__.Notify(491425)
		return len(name) != 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(491426)
		switch {
		case strings.EqualFold(name, exportGzipCodec):
			__antithesis_instrumentation__.Notify(491428)
			codec = roachpb.IOFileFormat_Gzip
		case strings.EqualFold(name, exportSnappyCodec) && func() bool {
			__antithesis_instrumentation__.Notify(491431)
			return fileSuffix == parquetSuffix == true
		}() == true:
			__antithesis_instrumentation__.Notify(491429)
			codec = roachpb.IOFileFormat_Snappy
		default:
			__antithesis_instrumentation__.Notify(491430)
			return nil, pgerror.Newf(pgcode.InvalidParameterValue,
				"unsupported compression codec %s for %s file format", name, fileSuffix)
		}
		__antithesis_instrumentation__.Notify(491427)
		format.Compression = codec
	} else {
		__antithesis_instrumentation__.Notify(491432)
	}
	__antithesis_instrumentation__.Notify(491373)

	exportID := ef.planner.stmt.QueryID.String()
	exportFilePattern := exportFilePatternPart + "." + fileSuffix
	namePattern := fmt.Sprintf("export%s-%s", exportID, exportFilePattern)
	return &exportNode{
		source:          input.(planNode),
		destination:     string(*destination),
		fileNamePattern: namePattern,
		format:          format,
		chunkRows:       chunkRows,
		chunkSize:       chunkSize,
		colNames:        colNames,
	}, nil
}
