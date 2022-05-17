package importer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/featureflag"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/resolver"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemadesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemaexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgnotice"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/ioctx"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
)

const (
	csvDelimiter    = "delimiter"
	csvComment      = "comment"
	csvNullIf       = "nullif"
	csvSkip         = "skip"
	csvRowLimit     = "row_limit"
	csvStrictQuotes = "strict_quotes"

	mysqlOutfileRowSep   = "rows_terminated_by"
	mysqlOutfileFieldSep = "fields_terminated_by"
	mysqlOutfileEnclose  = "fields_enclosed_by"
	mysqlOutfileEscape   = "fields_escaped_by"

	importOptionSSTSize          = "sstsize"
	importOptionDecompress       = "decompress"
	importOptionOversample       = "oversample"
	importOptionSkipFKs          = "skip_foreign_keys"
	importOptionDisableGlobMatch = "disable_glob_matching"
	importOptionSaveRejected     = "experimental_save_rejected"
	importOptionDetached         = "detached"

	pgCopyDelimiter = "delimiter"
	pgCopyNull      = "nullif"

	optMaxRowSize = "max_row_size"

	avroStrict = "strict_validation"

	avroBinRecords  = "data_as_binary_records"
	avroJSONRecords = "data_as_json_records"

	avroRecordsSeparatedBy = "records_terminated_by"

	avroSchema    = "schema"
	avroSchemaURI = "schema_uri"

	pgDumpIgnoreAllUnsupported     = "ignore_unsupported_statements"
	pgDumpIgnoreShuntFileDest      = "log_ignored_statements"
	pgDumpUnsupportedSchemaStmtLog = "unsupported_schema_stmts"
	pgDumpUnsupportedDataStmtLog   = "unsupported_data_stmts"

	runningStatusImportBundleParseSchema jobs.RunningStatus = "parsing schema on Import Bundle"
)

var importOptionExpectValues = map[string]sql.KVStringOptValidate{
	csvDelimiter:    sql.KVStringOptRequireValue,
	csvComment:      sql.KVStringOptRequireValue,
	csvNullIf:       sql.KVStringOptRequireValue,
	csvSkip:         sql.KVStringOptRequireValue,
	csvRowLimit:     sql.KVStringOptRequireValue,
	csvStrictQuotes: sql.KVStringOptRequireNoValue,

	mysqlOutfileRowSep:   sql.KVStringOptRequireValue,
	mysqlOutfileFieldSep: sql.KVStringOptRequireValue,
	mysqlOutfileEnclose:  sql.KVStringOptRequireValue,
	mysqlOutfileEscape:   sql.KVStringOptRequireValue,

	importOptionSSTSize:      sql.KVStringOptRequireValue,
	importOptionDecompress:   sql.KVStringOptRequireValue,
	importOptionOversample:   sql.KVStringOptRequireValue,
	importOptionSaveRejected: sql.KVStringOptRequireNoValue,

	importOptionSkipFKs:          sql.KVStringOptRequireNoValue,
	importOptionDisableGlobMatch: sql.KVStringOptRequireNoValue,
	importOptionDetached:         sql.KVStringOptRequireNoValue,

	optMaxRowSize: sql.KVStringOptRequireValue,

	avroStrict:             sql.KVStringOptRequireNoValue,
	avroSchema:             sql.KVStringOptRequireValue,
	avroSchemaURI:          sql.KVStringOptRequireValue,
	avroRecordsSeparatedBy: sql.KVStringOptRequireValue,
	avroBinRecords:         sql.KVStringOptRequireNoValue,
	avroJSONRecords:        sql.KVStringOptRequireNoValue,

	pgDumpIgnoreAllUnsupported: sql.KVStringOptRequireNoValue,
	pgDumpIgnoreShuntFileDest:  sql.KVStringOptRequireValue,
}

var pgDumpMaxLoggedStmts = 1024

func testingSetMaxLogIgnoredImportStatements(maxLogSize int) (cleanup func()) {
	__antithesis_instrumentation__.Notify(494001)
	prevLogSize := pgDumpMaxLoggedStmts
	pgDumpMaxLoggedStmts = maxLogSize
	return func() {
		__antithesis_instrumentation__.Notify(494002)
		pgDumpMaxLoggedStmts = prevLogSize
	}
}

func makeStringSet(opts ...string) map[string]struct{} {
	__antithesis_instrumentation__.Notify(494003)
	res := make(map[string]struct{}, len(opts))
	for _, opt := range opts {
		__antithesis_instrumentation__.Notify(494005)
		res[opt] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(494004)
	return res
}

var allowedCommonOptions = makeStringSet(
	importOptionSSTSize, importOptionDecompress, importOptionOversample,
	importOptionSaveRejected, importOptionDisableGlobMatch, importOptionDetached)

var avroAllowedOptions = makeStringSet(
	avroStrict, avroBinRecords, avroJSONRecords,
	avroRecordsSeparatedBy, avroSchema, avroSchemaURI, optMaxRowSize, csvRowLimit,
)

var csvAllowedOptions = makeStringSet(
	csvDelimiter, csvComment, csvNullIf, csvSkip, csvStrictQuotes, csvRowLimit,
)

var mysqlOutAllowedOptions = makeStringSet(
	mysqlOutfileRowSep, mysqlOutfileFieldSep, mysqlOutfileEnclose,
	mysqlOutfileEscape, csvNullIf, csvSkip, csvRowLimit,
)

var (
	mysqlDumpAllowedOptions = makeStringSet(importOptionSkipFKs, csvRowLimit)
	pgCopyAllowedOptions    = makeStringSet(pgCopyDelimiter, pgCopyNull, optMaxRowSize)
	pgDumpAllowedOptions    = makeStringSet(optMaxRowSize, importOptionSkipFKs, csvRowLimit,
		pgDumpIgnoreAllUnsupported, pgDumpIgnoreShuntFileDest)
)

var importIntoRequiredPrivileges = []privilege.Kind{privilege.INSERT, privilege.DROP}

var allowedIntoFormats = map[string]struct{}{
	"CSV":       {},
	"AVRO":      {},
	"DELIMITED": {},
	"PGCOPY":    {},
}

var featureImportEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"feature.import.enabled",
	"set to true to enable imports, false to disable; default is true",
	featureflag.FeatureFlagEnabledDefault,
).WithPublic()

func validateFormatOptions(
	format string, specified map[string]string, formatAllowed map[string]struct{},
) error {
	__antithesis_instrumentation__.Notify(494006)
	for opt := range specified {
		__antithesis_instrumentation__.Notify(494008)
		if _, ok := formatAllowed[opt]; !ok {
			__antithesis_instrumentation__.Notify(494009)
			if _, ok = allowedCommonOptions[opt]; !ok {
				__antithesis_instrumentation__.Notify(494010)
				return errors.Errorf(
					"invalid option %q specified for %s import format", opt, format)
			} else {
				__antithesis_instrumentation__.Notify(494011)
			}
		} else {
			__antithesis_instrumentation__.Notify(494012)
		}
	}
	__antithesis_instrumentation__.Notify(494007)
	return nil
}

func importJobDescription(
	p sql.PlanHookState, orig *tree.Import, files []string, opts map[string]string,
) (string, error) {
	__antithesis_instrumentation__.Notify(494013)
	stmt := *orig
	stmt.Files = nil
	for _, file := range files {
		__antithesis_instrumentation__.Notify(494017)
		clean, err := cloud.SanitizeExternalStorageURI(file, nil)
		if err != nil {
			__antithesis_instrumentation__.Notify(494019)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(494020)
		}
		__antithesis_instrumentation__.Notify(494018)
		stmt.Files = append(stmt.Files, tree.NewDString(clean))
	}
	__antithesis_instrumentation__.Notify(494014)
	stmt.Options = nil
	for k, v := range opts {
		__antithesis_instrumentation__.Notify(494021)
		opt := tree.KVOption{Key: tree.Name(k)}
		val := importOptionExpectValues[k] == sql.KVStringOptRequireValue
		val = val || func() bool {
			__antithesis_instrumentation__.Notify(494023)
			return (importOptionExpectValues[k] == sql.KVStringOptAny && func() bool {
				__antithesis_instrumentation__.Notify(494024)
				return len(v) > 0 == true
			}() == true) == true
		}() == true
		if val {
			__antithesis_instrumentation__.Notify(494025)
			opt.Value = tree.NewDString(v)
		} else {
			__antithesis_instrumentation__.Notify(494026)
		}
		__antithesis_instrumentation__.Notify(494022)
		stmt.Options = append(stmt.Options, opt)
	}
	__antithesis_instrumentation__.Notify(494015)
	sort.Slice(stmt.Options, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(494027)
		return stmt.Options[i].Key < stmt.Options[j].Key
	})
	__antithesis_instrumentation__.Notify(494016)
	ann := p.ExtendedEvalContext().Annotations
	return tree.AsStringWithFQNames(&stmt, ann), nil
}

func ensureRequiredPrivileges(
	ctx context.Context,
	requiredPrivileges []privilege.Kind,
	p sql.PlanHookState,
	desc *tabledesc.Mutable,
) error {
	__antithesis_instrumentation__.Notify(494028)
	for _, priv := range requiredPrivileges {
		__antithesis_instrumentation__.Notify(494030)
		err := p.CheckPrivilege(ctx, desc, priv)
		if err != nil {
			__antithesis_instrumentation__.Notify(494031)
			return err
		} else {
			__antithesis_instrumentation__.Notify(494032)
		}
	}
	__antithesis_instrumentation__.Notify(494029)

	return nil
}

func addToFileFormatTelemetry(fileFormat, state string) {
	__antithesis_instrumentation__.Notify(494033)
	telemetry.Count(fmt.Sprintf("%s.%s.%s", "import", strings.ToLower(fileFormat), state))
}

func resolveUDTsUsedByImportInto(
	ctx context.Context, p sql.PlanHookState, table *tabledesc.Mutable,
) ([]catalog.TypeDescriptor, error) {
	__antithesis_instrumentation__.Notify(494034)
	typeDescs := make([]catalog.TypeDescriptor, 0)
	var dbDesc catalog.DatabaseDescriptor
	err := sql.DescsTxn(ctx, p.ExecCfg(), func(
		ctx context.Context, txn *kv.Txn, descriptors *descs.Collection,
	) (err error) {
		__antithesis_instrumentation__.Notify(494036)
		_, dbDesc, err = descriptors.GetImmutableDatabaseByID(ctx, txn, table.GetParentID(),
			tree.DatabaseLookupFlags{
				Required:    true,
				AvoidLeased: true,
			})
		if err != nil {
			__antithesis_instrumentation__.Notify(494041)
			return err
		} else {
			__antithesis_instrumentation__.Notify(494042)
		}
		__antithesis_instrumentation__.Notify(494037)
		typeIDs, _, err := table.GetAllReferencedTypeIDs(dbDesc,
			func(id descpb.ID) (catalog.TypeDescriptor, error) {
				__antithesis_instrumentation__.Notify(494043)
				immutDesc, err := descriptors.GetImmutableTypeByID(ctx, txn, id, tree.ObjectLookupFlags{
					CommonLookupFlags: tree.CommonLookupFlags{
						Required:    true,
						AvoidLeased: true,
					},
				})
				if err != nil {
					__antithesis_instrumentation__.Notify(494045)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(494046)
				}
				__antithesis_instrumentation__.Notify(494044)
				return immutDesc, nil
			})
		__antithesis_instrumentation__.Notify(494038)
		if err != nil {
			__antithesis_instrumentation__.Notify(494047)
			return errors.Wrap(err, "resolving type descriptors")
		} else {
			__antithesis_instrumentation__.Notify(494048)
		}
		__antithesis_instrumentation__.Notify(494039)

		for _, typeID := range typeIDs {
			__antithesis_instrumentation__.Notify(494049)
			immutDesc, err := descriptors.GetImmutableTypeByID(ctx, txn, typeID, tree.ObjectLookupFlags{
				CommonLookupFlags: tree.CommonLookupFlags{
					Required:    true,
					AvoidLeased: true,
				},
			})
			if err != nil {
				__antithesis_instrumentation__.Notify(494051)
				return err
			} else {
				__antithesis_instrumentation__.Notify(494052)
			}
			__antithesis_instrumentation__.Notify(494050)
			typeDescs = append(typeDescs, immutDesc)
		}
		__antithesis_instrumentation__.Notify(494040)
		return err
	})
	__antithesis_instrumentation__.Notify(494035)
	return typeDescs, err
}

func importPlanHook(
	ctx context.Context, stmt tree.Statement, p sql.PlanHookState,
) (sql.PlanHookRowFn, colinfo.ResultColumns, []sql.PlanNode, bool, error) {
	__antithesis_instrumentation__.Notify(494053)
	importStmt, ok := stmt.(*tree.Import)
	if !ok {
		__antithesis_instrumentation__.Notify(494062)
		return nil, nil, nil, false, nil
	} else {
		__antithesis_instrumentation__.Notify(494063)
	}
	__antithesis_instrumentation__.Notify(494054)

	if !importStmt.Bundle && func() bool {
		__antithesis_instrumentation__.Notify(494064)
		return !importStmt.Into == true
	}() == true {
		__antithesis_instrumentation__.Notify(494065)
		p.BufferClientNotice(ctx, pgnotice.Newf("IMPORT TABLE has been deprecated in 21.2, and will be removed in a future version."+
			" Instead, use CREATE TABLE with the desired schema, and IMPORT INTO the newly created table."))
	} else {
		__antithesis_instrumentation__.Notify(494066)
	}
	__antithesis_instrumentation__.Notify(494055)

	addToFileFormatTelemetry(importStmt.FileFormat, "attempted")

	if err := featureflag.CheckEnabled(
		ctx,
		p.ExecCfg(),
		featureImportEnabled,
		"IMPORT",
	); err != nil {
		__antithesis_instrumentation__.Notify(494067)
		return nil, nil, nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(494068)
	}
	__antithesis_instrumentation__.Notify(494056)

	filesFn, err := p.TypeAsStringArray(ctx, importStmt.Files, "IMPORT")
	if err != nil {
		__antithesis_instrumentation__.Notify(494069)
		return nil, nil, nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(494070)
	}
	__antithesis_instrumentation__.Notify(494057)

	optsFn, err := p.TypeAsStringOpts(ctx, importStmt.Options, importOptionExpectValues)
	if err != nil {
		__antithesis_instrumentation__.Notify(494071)
		return nil, nil, nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(494072)
	}
	__antithesis_instrumentation__.Notify(494058)

	opts, optsErr := optsFn()

	var isDetached bool
	if _, ok := opts[importOptionDetached]; ok {
		__antithesis_instrumentation__.Notify(494073)
		isDetached = true
	} else {
		__antithesis_instrumentation__.Notify(494074)
	}
	__antithesis_instrumentation__.Notify(494059)

	fn := func(ctx context.Context, _ []sql.PlanNode, resultsCh chan<- tree.Datums) error {
		__antithesis_instrumentation__.Notify(494075)

		ctx, span := tracing.ChildSpan(ctx, importStmt.StatementTag())
		defer span.Finish()

		if !(p.ExtendedEvalContext().TxnIsSingleStmt || func() bool {
			__antithesis_instrumentation__.Notify(494097)
			return isDetached == true
		}() == true) {
			__antithesis_instrumentation__.Notify(494098)
			return errors.Errorf("IMPORT cannot be used inside a multi-statement transaction without DETACHED option")
		} else {
			__antithesis_instrumentation__.Notify(494099)
		}
		__antithesis_instrumentation__.Notify(494076)

		if optsErr != nil {
			__antithesis_instrumentation__.Notify(494100)
			return optsErr
		} else {
			__antithesis_instrumentation__.Notify(494101)
		}
		__antithesis_instrumentation__.Notify(494077)

		filenamePatterns, err := filesFn()
		if err != nil {
			__antithesis_instrumentation__.Notify(494102)
			return err
		} else {
			__antithesis_instrumentation__.Notify(494103)
		}
		__antithesis_instrumentation__.Notify(494078)

		if !p.ExecCfg().ExternalIODirConfig.EnableNonAdminImplicitAndArbitraryOutbound {
			__antithesis_instrumentation__.Notify(494104)
			for _, file := range filenamePatterns {
				__antithesis_instrumentation__.Notify(494105)
				conf, err := cloud.ExternalStorageConfFromURI(file, p.User())
				if err != nil {
					__antithesis_instrumentation__.Notify(494107)

					if _, workloadErr := parseWorkloadConfig(file); workloadErr == nil {
						__antithesis_instrumentation__.Notify(494109)
						continue
					} else {
						__antithesis_instrumentation__.Notify(494110)
					}
					__antithesis_instrumentation__.Notify(494108)
					return err
				} else {
					__antithesis_instrumentation__.Notify(494111)
				}
				__antithesis_instrumentation__.Notify(494106)
				if !conf.AccessIsWithExplicitAuth() {
					__antithesis_instrumentation__.Notify(494112)
					err := p.RequireAdminRole(ctx,
						fmt.Sprintf("IMPORT from the specified %s URI", conf.Provider.String()))
					if err != nil {
						__antithesis_instrumentation__.Notify(494113)
						return err
					} else {
						__antithesis_instrumentation__.Notify(494114)
					}
				} else {
					__antithesis_instrumentation__.Notify(494115)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(494116)
		}
		__antithesis_instrumentation__.Notify(494079)

		var files []string
		if _, ok := opts[importOptionDisableGlobMatch]; ok {
			__antithesis_instrumentation__.Notify(494117)
			files = filenamePatterns
		} else {
			__antithesis_instrumentation__.Notify(494118)
			for _, file := range filenamePatterns {
				__antithesis_instrumentation__.Notify(494119)
				uri, err := url.Parse(file)
				if err != nil {
					__antithesis_instrumentation__.Notify(494122)
					return err
				} else {
					__antithesis_instrumentation__.Notify(494123)
				}
				__antithesis_instrumentation__.Notify(494120)
				if strings.Contains(uri.Scheme, "workload") || func() bool {
					__antithesis_instrumentation__.Notify(494124)
					return strings.HasPrefix(uri.Scheme, "http") == true
				}() == true {
					__antithesis_instrumentation__.Notify(494125)
					files = append(files, file)
					continue
				} else {
					__antithesis_instrumentation__.Notify(494126)
				}
				__antithesis_instrumentation__.Notify(494121)
				prefix := cloud.GetPrefixBeforeWildcard(uri.Path)
				if len(prefix) < len(uri.Path) {
					__antithesis_instrumentation__.Notify(494127)
					pattern := uri.Path[len(prefix):]
					uri.Path = prefix
					s, err := p.ExecCfg().DistSQLSrv.ExternalStorageFromURI(ctx, uri.String(), p.User())
					if err != nil {
						__antithesis_instrumentation__.Notify(494131)
						return err
					} else {
						__antithesis_instrumentation__.Notify(494132)
					}
					__antithesis_instrumentation__.Notify(494128)
					var expandedFiles []string
					if err := s.List(ctx, "", "", func(s string) error {
						__antithesis_instrumentation__.Notify(494133)
						ok, err := path.Match(pattern, s)
						if ok {
							__antithesis_instrumentation__.Notify(494135)
							uri.Path = prefix + s
							expandedFiles = append(expandedFiles, uri.String())
						} else {
							__antithesis_instrumentation__.Notify(494136)
						}
						__antithesis_instrumentation__.Notify(494134)
						return err
					}); err != nil {
						__antithesis_instrumentation__.Notify(494137)
						return err
					} else {
						__antithesis_instrumentation__.Notify(494138)
					}
					__antithesis_instrumentation__.Notify(494129)
					if len(expandedFiles) < 1 {
						__antithesis_instrumentation__.Notify(494139)
						return errors.Errorf(`no files matched %q in prefix %q in uri provided: %q`, pattern, prefix, file)
					} else {
						__antithesis_instrumentation__.Notify(494140)
					}
					__antithesis_instrumentation__.Notify(494130)
					files = append(files, expandedFiles...)
				} else {
					__antithesis_instrumentation__.Notify(494141)
					files = append(files, file)
				}
			}
		}
		__antithesis_instrumentation__.Notify(494080)

		if importStmt.Bundle && func() bool {
			__antithesis_instrumentation__.Notify(494142)
			return len(files) != 1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(494143)
			return pgerror.New(pgcode.FeatureNotSupported, "SQL dump files must be imported individually")
		} else {
			__antithesis_instrumentation__.Notify(494144)
		}
		__antithesis_instrumentation__.Notify(494081)

		table := importStmt.Table
		var db catalog.DatabaseDescriptor
		var sc catalog.SchemaDescriptor
		if table != nil {
			__antithesis_instrumentation__.Notify(494145)

			un := table.ToUnresolvedObjectName()
			found, prefix, resPrefix, err := resolver.ResolveTarget(ctx,
				un, p, p.SessionData().Database, p.SessionData().SearchPath)
			if err != nil {
				__antithesis_instrumentation__.Notify(494149)
				return pgerror.Wrap(err, pgcode.UndefinedTable,
					"resolving target import name")
			} else {
				__antithesis_instrumentation__.Notify(494150)
			}
			__antithesis_instrumentation__.Notify(494146)
			if !found {
				__antithesis_instrumentation__.Notify(494151)

				return pgerror.Newf(pgcode.UndefinedObject,
					"database does not exist: %q", table)
			} else {
				__antithesis_instrumentation__.Notify(494152)
			}
			__antithesis_instrumentation__.Notify(494147)
			table.ObjectNamePrefix = prefix
			db = resPrefix.Database
			sc = resPrefix.Schema

			if !importStmt.Into {
				__antithesis_instrumentation__.Notify(494153)
				if err := p.CheckPrivilege(ctx, db, privilege.CREATE); err != nil {
					__antithesis_instrumentation__.Notify(494154)
					return err
				} else {
					__antithesis_instrumentation__.Notify(494155)
				}
			} else {
				__antithesis_instrumentation__.Notify(494156)
			}
			__antithesis_instrumentation__.Notify(494148)

			switch sc.SchemaKind() {
			case catalog.SchemaVirtual:
				__antithesis_instrumentation__.Notify(494157)
				return pgerror.Newf(pgcode.InvalidSchemaName,
					"cannot import into schema %q", table.SchemaName)
			default:
				__antithesis_instrumentation__.Notify(494158)
			}
		} else {
			__antithesis_instrumentation__.Notify(494159)

			txn := p.ExtendedEvalContext().Txn
			db, err = p.Accessor().GetDatabaseDesc(ctx, txn, p.SessionData().Database, tree.DatabaseLookupFlags{
				AvoidLeased: true,
				Required:    true,
			})
			if err != nil {
				__antithesis_instrumentation__.Notify(494162)
				return pgerror.Wrap(err, pgcode.UndefinedObject,
					"could not resolve current database")
			} else {
				__antithesis_instrumentation__.Notify(494163)
			}
			__antithesis_instrumentation__.Notify(494160)

			if !importStmt.Into {
				__antithesis_instrumentation__.Notify(494164)
				if err := p.CheckPrivilege(ctx, db, privilege.CREATE); err != nil {
					__antithesis_instrumentation__.Notify(494165)
					return err
				} else {
					__antithesis_instrumentation__.Notify(494166)
				}
			} else {
				__antithesis_instrumentation__.Notify(494167)
			}
			__antithesis_instrumentation__.Notify(494161)
			sc = schemadesc.GetPublicSchema()
		}
		__antithesis_instrumentation__.Notify(494082)

		format := roachpb.IOFileFormat{}
		switch importStmt.FileFormat {
		case "CSV":
			__antithesis_instrumentation__.Notify(494168)
			if err = validateFormatOptions(importStmt.FileFormat, opts, csvAllowedOptions); err != nil {
				__antithesis_instrumentation__.Notify(494200)
				return err
			} else {
				__antithesis_instrumentation__.Notify(494201)
			}
			__antithesis_instrumentation__.Notify(494169)
			format.Format = roachpb.IOFileFormat_CSV

			format.Csv.Comma = ','
			if override, ok := opts[csvDelimiter]; ok {
				__antithesis_instrumentation__.Notify(494202)
				comma, err := util.GetSingleRune(override)
				if err != nil {
					__antithesis_instrumentation__.Notify(494204)
					return pgerror.Wrap(err, pgcode.Syntax, "invalid comma value")
				} else {
					__antithesis_instrumentation__.Notify(494205)
				}
				__antithesis_instrumentation__.Notify(494203)
				format.Csv.Comma = comma
			} else {
				__antithesis_instrumentation__.Notify(494206)
			}
			__antithesis_instrumentation__.Notify(494170)

			if override, ok := opts[csvComment]; ok {
				__antithesis_instrumentation__.Notify(494207)
				comment, err := util.GetSingleRune(override)
				if err != nil {
					__antithesis_instrumentation__.Notify(494209)
					return pgerror.Wrap(err, pgcode.Syntax, "invalid comment value")
				} else {
					__antithesis_instrumentation__.Notify(494210)
				}
				__antithesis_instrumentation__.Notify(494208)
				format.Csv.Comment = comment
			} else {
				__antithesis_instrumentation__.Notify(494211)
			}
			__antithesis_instrumentation__.Notify(494171)

			if override, ok := opts[csvNullIf]; ok {
				__antithesis_instrumentation__.Notify(494212)
				format.Csv.NullEncoding = &override
			} else {
				__antithesis_instrumentation__.Notify(494213)
			}
			__antithesis_instrumentation__.Notify(494172)

			if override, ok := opts[csvSkip]; ok {
				__antithesis_instrumentation__.Notify(494214)
				skip, err := strconv.Atoi(override)
				if err != nil {
					__antithesis_instrumentation__.Notify(494217)
					return pgerror.Wrapf(err, pgcode.Syntax, "invalid %s value", csvSkip)
				} else {
					__antithesis_instrumentation__.Notify(494218)
				}
				__antithesis_instrumentation__.Notify(494215)
				if skip < 0 {
					__antithesis_instrumentation__.Notify(494219)
					return pgerror.Newf(pgcode.Syntax, "%s must be >= 0", csvSkip)
				} else {
					__antithesis_instrumentation__.Notify(494220)
				}
				__antithesis_instrumentation__.Notify(494216)
				format.Csv.Skip = uint32(skip)
			} else {
				__antithesis_instrumentation__.Notify(494221)
			}
			__antithesis_instrumentation__.Notify(494173)
			if _, ok := opts[csvStrictQuotes]; ok {
				__antithesis_instrumentation__.Notify(494222)
				format.Csv.StrictQuotes = true
			} else {
				__antithesis_instrumentation__.Notify(494223)
			}
			__antithesis_instrumentation__.Notify(494174)
			if _, ok := opts[importOptionSaveRejected]; ok {
				__antithesis_instrumentation__.Notify(494224)
				format.SaveRejected = true
			} else {
				__antithesis_instrumentation__.Notify(494225)
			}
			__antithesis_instrumentation__.Notify(494175)
			if override, ok := opts[csvRowLimit]; ok {
				__antithesis_instrumentation__.Notify(494226)
				rowLimit, err := strconv.Atoi(override)
				if err != nil {
					__antithesis_instrumentation__.Notify(494229)
					return pgerror.Wrapf(err, pgcode.Syntax, "invalid numeric %s value", csvRowLimit)
				} else {
					__antithesis_instrumentation__.Notify(494230)
				}
				__antithesis_instrumentation__.Notify(494227)
				if rowLimit <= 0 {
					__antithesis_instrumentation__.Notify(494231)
					return pgerror.Newf(pgcode.Syntax, "%s must be > 0", csvRowLimit)
				} else {
					__antithesis_instrumentation__.Notify(494232)
				}
				__antithesis_instrumentation__.Notify(494228)
				format.Csv.RowLimit = int64(rowLimit)
			} else {
				__antithesis_instrumentation__.Notify(494233)
			}
		case "DELIMITED":
			__antithesis_instrumentation__.Notify(494176)
			if err = validateFormatOptions(importStmt.FileFormat, opts, mysqlOutAllowedOptions); err != nil {
				__antithesis_instrumentation__.Notify(494234)
				return err
			} else {
				__antithesis_instrumentation__.Notify(494235)
			}
			__antithesis_instrumentation__.Notify(494177)
			format.Format = roachpb.IOFileFormat_MysqlOutfile
			format.MysqlOut = roachpb.MySQLOutfileOptions{
				RowSeparator:   '\n',
				FieldSeparator: '\t',
			}
			if override, ok := opts[mysqlOutfileRowSep]; ok {
				__antithesis_instrumentation__.Notify(494236)
				c, err := util.GetSingleRune(override)
				if err != nil {
					__antithesis_instrumentation__.Notify(494238)
					return pgerror.Wrapf(err, pgcode.Syntax,
						"invalid %q value", mysqlOutfileRowSep)
				} else {
					__antithesis_instrumentation__.Notify(494239)
				}
				__antithesis_instrumentation__.Notify(494237)
				format.MysqlOut.RowSeparator = c
			} else {
				__antithesis_instrumentation__.Notify(494240)
			}
			__antithesis_instrumentation__.Notify(494178)

			if override, ok := opts[mysqlOutfileFieldSep]; ok {
				__antithesis_instrumentation__.Notify(494241)
				c, err := util.GetSingleRune(override)
				if err != nil {
					__antithesis_instrumentation__.Notify(494243)
					return pgerror.Wrapf(err, pgcode.Syntax, "invalid %q value", mysqlOutfileFieldSep)
				} else {
					__antithesis_instrumentation__.Notify(494244)
				}
				__antithesis_instrumentation__.Notify(494242)
				format.MysqlOut.FieldSeparator = c
			} else {
				__antithesis_instrumentation__.Notify(494245)
			}
			__antithesis_instrumentation__.Notify(494179)

			if override, ok := opts[mysqlOutfileEnclose]; ok {
				__antithesis_instrumentation__.Notify(494246)
				c, err := util.GetSingleRune(override)
				if err != nil {
					__antithesis_instrumentation__.Notify(494248)
					return pgerror.Wrapf(err, pgcode.Syntax, "invalid %q value", mysqlOutfileRowSep)
				} else {
					__antithesis_instrumentation__.Notify(494249)
				}
				__antithesis_instrumentation__.Notify(494247)
				format.MysqlOut.Enclose = roachpb.MySQLOutfileOptions_Always
				format.MysqlOut.Encloser = c
			} else {
				__antithesis_instrumentation__.Notify(494250)
			}
			__antithesis_instrumentation__.Notify(494180)

			if override, ok := opts[mysqlOutfileEscape]; ok {
				__antithesis_instrumentation__.Notify(494251)
				c, err := util.GetSingleRune(override)
				if err != nil {
					__antithesis_instrumentation__.Notify(494253)
					return pgerror.Wrapf(err, pgcode.Syntax, "invalid %q value", mysqlOutfileRowSep)
				} else {
					__antithesis_instrumentation__.Notify(494254)
				}
				__antithesis_instrumentation__.Notify(494252)
				format.MysqlOut.HasEscape = true
				format.MysqlOut.Escape = c
			} else {
				__antithesis_instrumentation__.Notify(494255)
			}
			__antithesis_instrumentation__.Notify(494181)
			if override, ok := opts[csvSkip]; ok {
				__antithesis_instrumentation__.Notify(494256)
				skip, err := strconv.Atoi(override)
				if err != nil {
					__antithesis_instrumentation__.Notify(494259)
					return pgerror.Wrapf(err, pgcode.Syntax, "invalid %s value", csvSkip)
				} else {
					__antithesis_instrumentation__.Notify(494260)
				}
				__antithesis_instrumentation__.Notify(494257)
				if skip < 0 {
					__antithesis_instrumentation__.Notify(494261)
					return pgerror.Newf(pgcode.Syntax, "%s must be >= 0", csvSkip)
				} else {
					__antithesis_instrumentation__.Notify(494262)
				}
				__antithesis_instrumentation__.Notify(494258)
				format.MysqlOut.Skip = uint32(skip)
			} else {
				__antithesis_instrumentation__.Notify(494263)
			}
			__antithesis_instrumentation__.Notify(494182)
			if override, ok := opts[csvNullIf]; ok {
				__antithesis_instrumentation__.Notify(494264)
				format.MysqlOut.NullEncoding = &override
			} else {
				__antithesis_instrumentation__.Notify(494265)
			}
			__antithesis_instrumentation__.Notify(494183)
			if _, ok := opts[importOptionSaveRejected]; ok {
				__antithesis_instrumentation__.Notify(494266)
				format.SaveRejected = true
			} else {
				__antithesis_instrumentation__.Notify(494267)
			}
			__antithesis_instrumentation__.Notify(494184)
			if override, ok := opts[csvRowLimit]; ok {
				__antithesis_instrumentation__.Notify(494268)
				rowLimit, err := strconv.Atoi(override)
				if err != nil {
					__antithesis_instrumentation__.Notify(494271)
					return pgerror.Wrapf(err, pgcode.Syntax, "invalid numeric %s value", csvRowLimit)
				} else {
					__antithesis_instrumentation__.Notify(494272)
				}
				__antithesis_instrumentation__.Notify(494269)
				if rowLimit <= 0 {
					__antithesis_instrumentation__.Notify(494273)
					return pgerror.Newf(pgcode.Syntax, "%s must be > 0", csvRowLimit)
				} else {
					__antithesis_instrumentation__.Notify(494274)
				}
				__antithesis_instrumentation__.Notify(494270)
				format.MysqlOut.RowLimit = int64(rowLimit)
			} else {
				__antithesis_instrumentation__.Notify(494275)
			}
		case "MYSQLDUMP":
			__antithesis_instrumentation__.Notify(494185)
			if err = validateFormatOptions(importStmt.FileFormat, opts, mysqlDumpAllowedOptions); err != nil {
				__antithesis_instrumentation__.Notify(494276)
				return err
			} else {
				__antithesis_instrumentation__.Notify(494277)
			}
			__antithesis_instrumentation__.Notify(494186)
			format.Format = roachpb.IOFileFormat_Mysqldump
			if override, ok := opts[csvRowLimit]; ok {
				__antithesis_instrumentation__.Notify(494278)
				rowLimit, err := strconv.Atoi(override)
				if err != nil {
					__antithesis_instrumentation__.Notify(494281)
					return pgerror.Wrapf(err, pgcode.Syntax, "invalid numeric %s value", csvRowLimit)
				} else {
					__antithesis_instrumentation__.Notify(494282)
				}
				__antithesis_instrumentation__.Notify(494279)
				if rowLimit <= 0 {
					__antithesis_instrumentation__.Notify(494283)
					return pgerror.Newf(pgcode.Syntax, "%s must be > 0", csvRowLimit)
				} else {
					__antithesis_instrumentation__.Notify(494284)
				}
				__antithesis_instrumentation__.Notify(494280)
				format.MysqlDump.RowLimit = int64(rowLimit)
			} else {
				__antithesis_instrumentation__.Notify(494285)
			}
		case "PGCOPY":
			__antithesis_instrumentation__.Notify(494187)
			if err = validateFormatOptions(importStmt.FileFormat, opts, pgCopyAllowedOptions); err != nil {
				__antithesis_instrumentation__.Notify(494286)
				return err
			} else {
				__antithesis_instrumentation__.Notify(494287)
			}
			__antithesis_instrumentation__.Notify(494188)
			format.Format = roachpb.IOFileFormat_PgCopy
			format.PgCopy = roachpb.PgCopyOptions{
				Delimiter: '\t',
				Null:      `\N`,
			}
			if override, ok := opts[pgCopyDelimiter]; ok {
				__antithesis_instrumentation__.Notify(494288)
				c, err := util.GetSingleRune(override)
				if err != nil {
					__antithesis_instrumentation__.Notify(494290)
					return pgerror.Wrapf(err, pgcode.Syntax, "invalid %q value", pgCopyDelimiter)
				} else {
					__antithesis_instrumentation__.Notify(494291)
				}
				__antithesis_instrumentation__.Notify(494289)
				format.PgCopy.Delimiter = c
			} else {
				__antithesis_instrumentation__.Notify(494292)
			}
			__antithesis_instrumentation__.Notify(494189)
			if override, ok := opts[pgCopyNull]; ok {
				__antithesis_instrumentation__.Notify(494293)
				format.PgCopy.Null = override
			} else {
				__antithesis_instrumentation__.Notify(494294)
			}
			__antithesis_instrumentation__.Notify(494190)
			maxRowSize := int32(defaultScanBuffer)
			if override, ok := opts[optMaxRowSize]; ok {
				__antithesis_instrumentation__.Notify(494295)
				sz, err := humanizeutil.ParseBytes(override)
				if err != nil {
					__antithesis_instrumentation__.Notify(494298)
					return err
				} else {
					__antithesis_instrumentation__.Notify(494299)
				}
				__antithesis_instrumentation__.Notify(494296)
				if sz < 1 || func() bool {
					__antithesis_instrumentation__.Notify(494300)
					return sz > math.MaxInt32 == true
				}() == true {
					__antithesis_instrumentation__.Notify(494301)
					return errors.Errorf("%d out of range: %d", maxRowSize, sz)
				} else {
					__antithesis_instrumentation__.Notify(494302)
				}
				__antithesis_instrumentation__.Notify(494297)
				maxRowSize = int32(sz)
			} else {
				__antithesis_instrumentation__.Notify(494303)
			}
			__antithesis_instrumentation__.Notify(494191)
			format.PgCopy.MaxRowSize = maxRowSize
		case "PGDUMP":
			__antithesis_instrumentation__.Notify(494192)
			if err = validateFormatOptions(importStmt.FileFormat, opts, pgDumpAllowedOptions); err != nil {
				__antithesis_instrumentation__.Notify(494304)
				return err
			} else {
				__antithesis_instrumentation__.Notify(494305)
			}
			__antithesis_instrumentation__.Notify(494193)
			format.Format = roachpb.IOFileFormat_PgDump
			maxRowSize := int32(defaultScanBuffer)
			if override, ok := opts[optMaxRowSize]; ok {
				__antithesis_instrumentation__.Notify(494306)
				sz, err := humanizeutil.ParseBytes(override)
				if err != nil {
					__antithesis_instrumentation__.Notify(494309)
					return err
				} else {
					__antithesis_instrumentation__.Notify(494310)
				}
				__antithesis_instrumentation__.Notify(494307)
				if sz < 1 || func() bool {
					__antithesis_instrumentation__.Notify(494311)
					return sz > math.MaxInt32 == true
				}() == true {
					__antithesis_instrumentation__.Notify(494312)
					return errors.Errorf("%d out of range: %d", maxRowSize, sz)
				} else {
					__antithesis_instrumentation__.Notify(494313)
				}
				__antithesis_instrumentation__.Notify(494308)
				maxRowSize = int32(sz)
			} else {
				__antithesis_instrumentation__.Notify(494314)
			}
			__antithesis_instrumentation__.Notify(494194)
			format.PgDump.MaxRowSize = maxRowSize
			if _, ok := opts[pgDumpIgnoreAllUnsupported]; ok {
				__antithesis_instrumentation__.Notify(494315)
				format.PgDump.IgnoreUnsupported = true
			} else {
				__antithesis_instrumentation__.Notify(494316)
			}
			__antithesis_instrumentation__.Notify(494195)

			if dest, ok := opts[pgDumpIgnoreShuntFileDest]; ok {
				__antithesis_instrumentation__.Notify(494317)
				if !format.PgDump.IgnoreUnsupported {
					__antithesis_instrumentation__.Notify(494319)
					return errors.New("cannot log unsupported PGDUMP stmts without `ignore_unsupported_statements` option")
				} else {
					__antithesis_instrumentation__.Notify(494320)
				}
				__antithesis_instrumentation__.Notify(494318)
				format.PgDump.IgnoreUnsupportedLog = dest
			} else {
				__antithesis_instrumentation__.Notify(494321)
			}
			__antithesis_instrumentation__.Notify(494196)

			if override, ok := opts[csvRowLimit]; ok {
				__antithesis_instrumentation__.Notify(494322)
				rowLimit, err := strconv.Atoi(override)
				if err != nil {
					__antithesis_instrumentation__.Notify(494325)
					return pgerror.Wrapf(err, pgcode.Syntax, "invalid numeric %s value", csvRowLimit)
				} else {
					__antithesis_instrumentation__.Notify(494326)
				}
				__antithesis_instrumentation__.Notify(494323)
				if rowLimit <= 0 {
					__antithesis_instrumentation__.Notify(494327)
					return pgerror.Newf(pgcode.Syntax, "%s must be > 0", csvRowLimit)
				} else {
					__antithesis_instrumentation__.Notify(494328)
				}
				__antithesis_instrumentation__.Notify(494324)
				format.PgDump.RowLimit = int64(rowLimit)
			} else {
				__antithesis_instrumentation__.Notify(494329)
			}
		case "AVRO":
			__antithesis_instrumentation__.Notify(494197)
			if err = validateFormatOptions(importStmt.FileFormat, opts, avroAllowedOptions); err != nil {
				__antithesis_instrumentation__.Notify(494330)
				return err
			} else {
				__antithesis_instrumentation__.Notify(494331)
			}
			__antithesis_instrumentation__.Notify(494198)
			err := parseAvroOptions(ctx, opts, p, &format)
			if err != nil {
				__antithesis_instrumentation__.Notify(494332)
				return err
			} else {
				__antithesis_instrumentation__.Notify(494333)
			}
		default:
			__antithesis_instrumentation__.Notify(494199)
			return unimplemented.Newf("import.format", "unsupported import format: %q", importStmt.FileFormat)
		}
		__antithesis_instrumentation__.Notify(494083)

		var sstSize int64
		if override, ok := opts[importOptionSSTSize]; ok {
			__antithesis_instrumentation__.Notify(494334)
			sz, err := humanizeutil.ParseBytes(override)
			if err != nil {
				__antithesis_instrumentation__.Notify(494336)
				return err
			} else {
				__antithesis_instrumentation__.Notify(494337)
			}
			__antithesis_instrumentation__.Notify(494335)
			sstSize = sz
		} else {
			__antithesis_instrumentation__.Notify(494338)
		}
		__antithesis_instrumentation__.Notify(494084)
		var oversample int64
		if override, ok := opts[importOptionOversample]; ok {
			__antithesis_instrumentation__.Notify(494339)
			os, err := strconv.ParseInt(override, 10, 64)
			if err != nil {
				__antithesis_instrumentation__.Notify(494341)
				return err
			} else {
				__antithesis_instrumentation__.Notify(494342)
			}
			__antithesis_instrumentation__.Notify(494340)
			oversample = os
		} else {
			__antithesis_instrumentation__.Notify(494343)
		}
		__antithesis_instrumentation__.Notify(494085)

		var skipFKs bool
		if _, ok := opts[importOptionSkipFKs]; ok {
			__antithesis_instrumentation__.Notify(494344)
			skipFKs = true
		} else {
			__antithesis_instrumentation__.Notify(494345)
		}
		__antithesis_instrumentation__.Notify(494086)

		if override, ok := opts[importOptionDecompress]; ok {
			__antithesis_instrumentation__.Notify(494346)
			found := false
			for name, value := range roachpb.IOFileFormat_Compression_value {
				__antithesis_instrumentation__.Notify(494348)
				if strings.EqualFold(name, override) {
					__antithesis_instrumentation__.Notify(494349)
					format.Compression = roachpb.IOFileFormat_Compression(value)
					found = true
					break
				} else {
					__antithesis_instrumentation__.Notify(494350)
				}
			}
			__antithesis_instrumentation__.Notify(494347)
			if !found {
				__antithesis_instrumentation__.Notify(494351)
				return unimplemented.Newf("import.compression", "unsupported compression value: %q", override)
			} else {
				__antithesis_instrumentation__.Notify(494352)
			}
		} else {
			__antithesis_instrumentation__.Notify(494353)
		}
		__antithesis_instrumentation__.Notify(494087)

		var tableDetails []jobspb.ImportDetails_Table
		var typeDetails []jobspb.ImportDetails_Type
		jobDesc, err := importJobDescription(p, importStmt, filenamePatterns, opts)
		if err != nil {
			__antithesis_instrumentation__.Notify(494354)
			return err
		} else {
			__antithesis_instrumentation__.Notify(494355)
		}
		__antithesis_instrumentation__.Notify(494088)

		if importStmt.Into {
			__antithesis_instrumentation__.Notify(494356)
			if _, ok := allowedIntoFormats[importStmt.FileFormat]; !ok {
				__antithesis_instrumentation__.Notify(494362)
				return errors.Newf(
					"%s file format is currently unsupported by IMPORT INTO",
					importStmt.FileFormat)
			} else {
				__antithesis_instrumentation__.Notify(494363)
			}
			__antithesis_instrumentation__.Notify(494357)
			_, found, err := p.ResolveMutableTableDescriptor(ctx, table, true, tree.ResolveRequireTableDesc)
			if err != nil {
				__antithesis_instrumentation__.Notify(494364)
				return err
			} else {
				__antithesis_instrumentation__.Notify(494365)
			}
			__antithesis_instrumentation__.Notify(494358)

			err = ensureRequiredPrivileges(ctx, importIntoRequiredPrivileges, p, found)
			if err != nil {
				__antithesis_instrumentation__.Notify(494366)
				return err
			} else {
				__antithesis_instrumentation__.Notify(494367)
			}
			__antithesis_instrumentation__.Notify(494359)

			var intoCols []string
			isTargetCol := make(map[string]bool)
			for _, name := range importStmt.IntoCols {
				__antithesis_instrumentation__.Notify(494368)
				active, err := tabledesc.FindPublicColumnsWithNames(found, tree.NameList{name})
				if err != nil {
					__antithesis_instrumentation__.Notify(494370)
					return errors.Wrap(err, "verifying target columns")
				} else {
					__antithesis_instrumentation__.Notify(494371)
				}
				__antithesis_instrumentation__.Notify(494369)

				isTargetCol[active[0].GetName()] = true
				intoCols = append(intoCols, active[0].GetName())
			}
			__antithesis_instrumentation__.Notify(494360)

			if len(isTargetCol) != 0 {
				__antithesis_instrumentation__.Notify(494372)
				for _, col := range found.VisibleColumns() {
					__antithesis_instrumentation__.Notify(494373)
					if !(isTargetCol[col.GetName()] || func() bool {
						__antithesis_instrumentation__.Notify(494375)
						return col.IsNullable() == true
					}() == true || func() bool {
						__antithesis_instrumentation__.Notify(494376)
						return col.HasDefault() == true
					}() == true || func() bool {
						__antithesis_instrumentation__.Notify(494377)
						return col.IsComputed() == true
					}() == true) {
						__antithesis_instrumentation__.Notify(494378)
						return errors.Newf(
							"all non-target columns in IMPORT INTO must be nullable "+
								"or have default expressions, or have computed expressions"+
								" but violated by column %q",
							col.GetName(),
						)
					} else {
						__antithesis_instrumentation__.Notify(494379)
					}
					__antithesis_instrumentation__.Notify(494374)
					if isTargetCol[col.GetName()] && func() bool {
						__antithesis_instrumentation__.Notify(494380)
						return col.IsComputed() == true
					}() == true {
						__antithesis_instrumentation__.Notify(494381)
						return schemaexpr.CannotWriteToComputedColError(col.GetName())
					} else {
						__antithesis_instrumentation__.Notify(494382)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(494383)
			}

			{
				__antithesis_instrumentation__.Notify(494384)

				typeDescs, err := resolveUDTsUsedByImportInto(ctx, p, found)
				if err != nil {
					__antithesis_instrumentation__.Notify(494387)
					return errors.Wrap(err, "resolving UDTs used by table being imported into")
				} else {
					__antithesis_instrumentation__.Notify(494388)
				}
				__antithesis_instrumentation__.Notify(494385)
				if len(typeDescs) > 0 {
					__antithesis_instrumentation__.Notify(494389)
					typeDetails = make([]jobspb.ImportDetails_Type, 0, len(typeDescs))
				} else {
					__antithesis_instrumentation__.Notify(494390)
				}
				__antithesis_instrumentation__.Notify(494386)
				for _, typeDesc := range typeDescs {
					__antithesis_instrumentation__.Notify(494391)
					typeDetails = append(typeDetails, jobspb.ImportDetails_Type{Desc: typeDesc.TypeDesc()})
				}
			}
			__antithesis_instrumentation__.Notify(494361)

			tableDetails = []jobspb.ImportDetails_Table{{Desc: &found.TableDescriptor, IsNew: false, TargetCols: intoCols}}
		} else {
			__antithesis_instrumentation__.Notify(494392)
			if importStmt.Bundle {
				__antithesis_instrumentation__.Notify(494393)

				if table != nil {
					__antithesis_instrumentation__.Notify(494395)
					tableDetails = make([]jobspb.ImportDetails_Table, 1)
					tableName := table.ObjectName.String()

					if format.Format == roachpb.IOFileFormat_PgDump {
						__antithesis_instrumentation__.Notify(494397)
						if table.Schema() == "" {
							__antithesis_instrumentation__.Notify(494399)
							return errors.Newf("expected schema for target table %s to be resolved",
								tableName)
						} else {
							__antithesis_instrumentation__.Notify(494400)
						}
						__antithesis_instrumentation__.Notify(494398)
						tableName = fmt.Sprintf("%s.%s", table.SchemaName.String(),
							table.ObjectName.String())
					} else {
						__antithesis_instrumentation__.Notify(494401)
					}
					__antithesis_instrumentation__.Notify(494396)
					tableDetails[0] = jobspb.ImportDetails_Table{
						Name:  tableName,
						IsNew: true,
					}
				} else {
					__antithesis_instrumentation__.Notify(494402)
				}
				__antithesis_instrumentation__.Notify(494394)

				publicSchemaID := db.GetSchemaID(tree.PublicSchema)
				if sc.GetID() != publicSchemaID && func() bool {
					__antithesis_instrumentation__.Notify(494403)
					return sc.GetID() != keys.PublicSchemaID == true
				}() == true {
					__antithesis_instrumentation__.Notify(494404)
					err := errors.New("cannot use IMPORT with a user defined schema")
					hint := errors.WithHint(err, "create the table with CREATE TABLE and use IMPORT INTO instead")
					return hint
				} else {
					__antithesis_instrumentation__.Notify(494405)
				}
			} else {
				__antithesis_instrumentation__.Notify(494406)
			}
		}
		__antithesis_instrumentation__.Notify(494089)

		var databasePrimaryRegion catpb.RegionName
		if db.IsMultiRegion() {
			__antithesis_instrumentation__.Notify(494407)
			if err := sql.DescsTxn(ctx, p.ExecCfg(), func(ctx context.Context, txn *kv.Txn,
				descsCol *descs.Collection) error {
				__antithesis_instrumentation__.Notify(494408)
				regionConfig, err := sql.SynthesizeRegionConfig(ctx, txn, db.GetID(), descsCol)
				if err != nil {
					__antithesis_instrumentation__.Notify(494410)
					return err
				} else {
					__antithesis_instrumentation__.Notify(494411)
				}
				__antithesis_instrumentation__.Notify(494409)
				databasePrimaryRegion = regionConfig.PrimaryRegion()
				return nil
			}); err != nil {
				__antithesis_instrumentation__.Notify(494412)
				return errors.Wrap(err, "failed to resolve region config for multi region database")
			} else {
				__antithesis_instrumentation__.Notify(494413)
			}
		} else {
			__antithesis_instrumentation__.Notify(494414)
		}
		__antithesis_instrumentation__.Notify(494090)

		telemetry.CountBucketed("import.files", int64(len(files)))

		for _, file := range files {
			__antithesis_instrumentation__.Notify(494415)
			uri, err := url.Parse(file)

			if err != nil {
				__antithesis_instrumentation__.Notify(494417)
				log.Warningf(ctx, "failed to collect file specific import telemetry for %s", uri)
				continue
			} else {
				__antithesis_instrumentation__.Notify(494418)
			}
			__antithesis_instrumentation__.Notify(494416)

			if uri.Scheme == "userfile" {
				__antithesis_instrumentation__.Notify(494419)
				telemetry.Count("import.storage.userfile")
				break
			} else {
				__antithesis_instrumentation__.Notify(494420)
			}
		}
		__antithesis_instrumentation__.Notify(494091)
		if importStmt.Into {
			__antithesis_instrumentation__.Notify(494421)
			telemetry.Count("import.into")
		} else {
			__antithesis_instrumentation__.Notify(494422)
		}
		__antithesis_instrumentation__.Notify(494092)

		importDetails := jobspb.ImportDetails{
			URIs:                  files,
			Format:                format,
			ParentID:              db.GetID(),
			Tables:                tableDetails,
			Types:                 typeDetails,
			SSTSize:               sstSize,
			Oversample:            oversample,
			SkipFKs:               skipFKs,
			ParseBundleSchema:     importStmt.Bundle,
			DefaultIntSize:        p.SessionData().DefaultIntSize,
			DatabasePrimaryRegion: databasePrimaryRegion,
		}

		jr := jobs.Record{
			Description: jobDesc,
			Username:    p.User(),
			Details:     importDetails,
			Progress:    jobspb.ImportProgress{},
		}

		if isDetached {
			__antithesis_instrumentation__.Notify(494423)

			jobID := p.ExecCfg().JobRegistry.MakeJobID()
			_, err := p.ExecCfg().JobRegistry.CreateAdoptableJobWithTxn(
				ctx, jr, jobID, p.ExtendedEvalContext().Txn)
			if err != nil {
				__antithesis_instrumentation__.Notify(494425)
				return err
			} else {
				__antithesis_instrumentation__.Notify(494426)
			}
			__antithesis_instrumentation__.Notify(494424)

			addToFileFormatTelemetry(format.Format.String(), "started")
			resultsCh <- tree.Datums{tree.NewDInt(tree.DInt(jobID))}
			return nil
		} else {
			__antithesis_instrumentation__.Notify(494427)
		}
		__antithesis_instrumentation__.Notify(494093)

		plannerTxn := p.ExtendedEvalContext().Txn

		var sj *jobs.StartableJob
		if err := func() (err error) {
			__antithesis_instrumentation__.Notify(494428)
			defer func() {
				__antithesis_instrumentation__.Notify(494431)
				if err == nil || func() bool {
					__antithesis_instrumentation__.Notify(494433)
					return sj == nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(494434)
					return
				} else {
					__antithesis_instrumentation__.Notify(494435)
				}
				__antithesis_instrumentation__.Notify(494432)
				if cleanupErr := sj.CleanupOnRollback(ctx); cleanupErr != nil {
					__antithesis_instrumentation__.Notify(494436)
					log.Errorf(ctx, "failed to cleanup job: %v", cleanupErr)
				} else {
					__antithesis_instrumentation__.Notify(494437)
				}
			}()
			__antithesis_instrumentation__.Notify(494429)
			jobID := p.ExecCfg().JobRegistry.MakeJobID()
			if err := p.ExecCfg().JobRegistry.CreateStartableJobWithTxn(ctx, &sj, jobID, plannerTxn, jr); err != nil {
				__antithesis_instrumentation__.Notify(494438)
				return err
			} else {
				__antithesis_instrumentation__.Notify(494439)
			}
			__antithesis_instrumentation__.Notify(494430)

			return plannerTxn.Commit(ctx)
		}(); err != nil {
			__antithesis_instrumentation__.Notify(494440)
			return err
		} else {
			__antithesis_instrumentation__.Notify(494441)
		}
		__antithesis_instrumentation__.Notify(494094)

		if err := sj.Start(ctx); err != nil {
			__antithesis_instrumentation__.Notify(494442)
			return err
		} else {
			__antithesis_instrumentation__.Notify(494443)
		}
		__antithesis_instrumentation__.Notify(494095)
		addToFileFormatTelemetry(format.Format.String(), "started")
		if err := sj.AwaitCompletion(ctx); err != nil {
			__antithesis_instrumentation__.Notify(494444)
			return err
		} else {
			__antithesis_instrumentation__.Notify(494445)
		}
		__antithesis_instrumentation__.Notify(494096)
		return sj.ReportExecutionResults(ctx, resultsCh)
	}
	__antithesis_instrumentation__.Notify(494060)

	if isDetached {
		__antithesis_instrumentation__.Notify(494446)
		return fn, jobs.DetachedJobExecutionResultHeader, nil, false, nil
	} else {
		__antithesis_instrumentation__.Notify(494447)
	}
	__antithesis_instrumentation__.Notify(494061)
	return fn, jobs.BulkJobExecutionResultHeader, nil, false, nil
}

func parseAvroOptions(
	ctx context.Context, opts map[string]string, p sql.PlanHookState, format *roachpb.IOFileFormat,
) error {
	__antithesis_instrumentation__.Notify(494448)
	format.Format = roachpb.IOFileFormat_Avro

	format.Avro.Format = roachpb.AvroOptions_OCF
	_, format.Avro.StrictMode = opts[avroStrict]

	_, haveBinRecs := opts[avroBinRecords]
	_, haveJSONRecs := opts[avroJSONRecords]

	if haveBinRecs && func() bool {
		__antithesis_instrumentation__.Notify(494452)
		return haveJSONRecs == true
	}() == true {
		__antithesis_instrumentation__.Notify(494453)
		return errors.Errorf("only one of the %s or %s options can be set", avroBinRecords, avroJSONRecords)
	} else {
		__antithesis_instrumentation__.Notify(494454)
	}
	__antithesis_instrumentation__.Notify(494449)

	if override, ok := opts[csvRowLimit]; ok {
		__antithesis_instrumentation__.Notify(494455)
		rowLimit, err := strconv.Atoi(override)
		if err != nil {
			__antithesis_instrumentation__.Notify(494458)
			return pgerror.Wrapf(err, pgcode.Syntax, "invalid numeric %s value", csvRowLimit)
		} else {
			__antithesis_instrumentation__.Notify(494459)
		}
		__antithesis_instrumentation__.Notify(494456)
		if rowLimit <= 0 {
			__antithesis_instrumentation__.Notify(494460)
			return pgerror.Newf(pgcode.Syntax, "%s must be > 0", csvRowLimit)
		} else {
			__antithesis_instrumentation__.Notify(494461)
		}
		__antithesis_instrumentation__.Notify(494457)
		format.Avro.RowLimit = int64(rowLimit)
	} else {
		__antithesis_instrumentation__.Notify(494462)
	}
	__antithesis_instrumentation__.Notify(494450)

	if haveBinRecs || func() bool {
		__antithesis_instrumentation__.Notify(494463)
		return haveJSONRecs == true
	}() == true {
		__antithesis_instrumentation__.Notify(494464)

		if haveBinRecs {
			__antithesis_instrumentation__.Notify(494468)
			format.Avro.Format = roachpb.AvroOptions_BIN_RECORDS
		} else {
			__antithesis_instrumentation__.Notify(494469)
			format.Avro.Format = roachpb.AvroOptions_JSON_RECORDS
		}
		__antithesis_instrumentation__.Notify(494465)

		format.Avro.RecordSeparator = '\n'
		if override, ok := opts[avroRecordsSeparatedBy]; ok {
			__antithesis_instrumentation__.Notify(494470)
			c, err := util.GetSingleRune(override)
			if err != nil {
				__antithesis_instrumentation__.Notify(494472)
				return pgerror.Wrapf(err, pgcode.Syntax,
					"invalid %q value", avroRecordsSeparatedBy)
			} else {
				__antithesis_instrumentation__.Notify(494473)
			}
			__antithesis_instrumentation__.Notify(494471)
			format.Avro.RecordSeparator = c
		} else {
			__antithesis_instrumentation__.Notify(494474)
		}
		__antithesis_instrumentation__.Notify(494466)

		format.Avro.SchemaJSON = opts[avroSchema]

		if len(format.Avro.SchemaJSON) == 0 {
			__antithesis_instrumentation__.Notify(494475)

			uri, ok := opts[avroSchemaURI]
			if !ok {
				__antithesis_instrumentation__.Notify(494480)
				return errors.Errorf(
					"either %s or %s option must be set when importing avro record files", avroSchema, avroSchemaURI)
			} else {
				__antithesis_instrumentation__.Notify(494481)
			}
			__antithesis_instrumentation__.Notify(494476)

			store, err := p.ExecCfg().DistSQLSrv.ExternalStorageFromURI(ctx, uri, p.User())
			if err != nil {
				__antithesis_instrumentation__.Notify(494482)
				return err
			} else {
				__antithesis_instrumentation__.Notify(494483)
			}
			__antithesis_instrumentation__.Notify(494477)
			defer store.Close()

			raw, err := store.ReadFile(ctx, "")
			if err != nil {
				__antithesis_instrumentation__.Notify(494484)
				return err
			} else {
				__antithesis_instrumentation__.Notify(494485)
			}
			__antithesis_instrumentation__.Notify(494478)
			defer raw.Close(ctx)
			schemaBytes, err := ioctx.ReadAll(ctx, raw)
			if err != nil {
				__antithesis_instrumentation__.Notify(494486)
				return err
			} else {
				__antithesis_instrumentation__.Notify(494487)
			}
			__antithesis_instrumentation__.Notify(494479)
			format.Avro.SchemaJSON = string(schemaBytes)
		} else {
			__antithesis_instrumentation__.Notify(494488)
		}
		__antithesis_instrumentation__.Notify(494467)

		if override, ok := opts[optMaxRowSize]; ok {
			__antithesis_instrumentation__.Notify(494489)
			sz, err := humanizeutil.ParseBytes(override)
			if err != nil {
				__antithesis_instrumentation__.Notify(494492)
				return err
			} else {
				__antithesis_instrumentation__.Notify(494493)
			}
			__antithesis_instrumentation__.Notify(494490)
			if sz < 1 || func() bool {
				__antithesis_instrumentation__.Notify(494494)
				return sz > math.MaxInt32 == true
			}() == true {
				__antithesis_instrumentation__.Notify(494495)
				return errors.Errorf("%s out of range: %d", override, sz)
			} else {
				__antithesis_instrumentation__.Notify(494496)
			}
			__antithesis_instrumentation__.Notify(494491)
			format.Avro.MaxRecordSize = int32(sz)
		} else {
			__antithesis_instrumentation__.Notify(494497)
		}
	} else {
		__antithesis_instrumentation__.Notify(494498)
	}
	__antithesis_instrumentation__.Notify(494451)
	return nil
}

type loggerKind int

const (
	schemaParsing loggerKind = iota
	dataIngestion
)

type unsupportedStmtLogger struct {
	ctx   context.Context
	user  security.SQLUsername
	jobID int64

	ignoreUnsupported        bool
	ignoreUnsupportedLogDest string
	externalStorage          cloud.ExternalStorageFactory

	logBuffer       *bytes.Buffer
	numIgnoredStmts int

	flushCount int

	loggerType loggerKind
}

func makeUnsupportedStmtLogger(
	ctx context.Context,
	user security.SQLUsername,
	jobID int64,
	ignoreUnsupported bool,
	unsupportedLogDest string,
	loggerType loggerKind,
	externalStorage cloud.ExternalStorageFactory,
) *unsupportedStmtLogger {
	__antithesis_instrumentation__.Notify(494499)
	return &unsupportedStmtLogger{
		ctx:                      ctx,
		user:                     user,
		jobID:                    jobID,
		ignoreUnsupported:        ignoreUnsupported,
		ignoreUnsupportedLogDest: unsupportedLogDest,
		loggerType:               loggerType,
		logBuffer:                new(bytes.Buffer),
		externalStorage:          externalStorage,
	}
}

func (u *unsupportedStmtLogger) log(logLine string, isParseError bool) error {
	__antithesis_instrumentation__.Notify(494500)

	skipLoggingParseErr := isParseError && func() bool {
		__antithesis_instrumentation__.Notify(494504)
		return u.loggerType == dataIngestion == true
	}() == true
	if u.ignoreUnsupportedLogDest == "" || func() bool {
		__antithesis_instrumentation__.Notify(494505)
		return skipLoggingParseErr == true
	}() == true {
		__antithesis_instrumentation__.Notify(494506)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(494507)
	}
	__antithesis_instrumentation__.Notify(494501)

	if u.numIgnoredStmts >= pgDumpMaxLoggedStmts {
		__antithesis_instrumentation__.Notify(494508)
		err := u.flush()
		if err != nil {
			__antithesis_instrumentation__.Notify(494509)
			return err
		} else {
			__antithesis_instrumentation__.Notify(494510)
		}
	} else {
		__antithesis_instrumentation__.Notify(494511)
	}
	__antithesis_instrumentation__.Notify(494502)

	if isParseError {
		__antithesis_instrumentation__.Notify(494512)
		logLine = fmt.Sprintf("%s: could not be parsed\n", logLine)
	} else {
		__antithesis_instrumentation__.Notify(494513)
		logLine = fmt.Sprintf("%s: unsupported by IMPORT\n", logLine)
	}
	__antithesis_instrumentation__.Notify(494503)
	u.logBuffer.Write([]byte(logLine))
	u.numIgnoredStmts++
	return nil
}

func (u *unsupportedStmtLogger) flush() error {
	__antithesis_instrumentation__.Notify(494514)
	if u.ignoreUnsupportedLogDest == "" {
		__antithesis_instrumentation__.Notify(494520)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(494521)
	}
	__antithesis_instrumentation__.Notify(494515)

	conf, err := cloud.ExternalStorageConfFromURI(u.ignoreUnsupportedLogDest, u.user)
	if err != nil {
		__antithesis_instrumentation__.Notify(494522)
		return errors.Wrap(err, "failed to log unsupported stmts during IMPORT PGDUMP")
	} else {
		__antithesis_instrumentation__.Notify(494523)
	}
	__antithesis_instrumentation__.Notify(494516)
	var s cloud.ExternalStorage
	if s, err = u.externalStorage(u.ctx, conf); err != nil {
		__antithesis_instrumentation__.Notify(494524)
		return errors.New("failed to log unsupported stmts during IMPORT PGDUMP")
	} else {
		__antithesis_instrumentation__.Notify(494525)
	}
	__antithesis_instrumentation__.Notify(494517)
	defer s.Close()

	logFileName := fmt.Sprintf("import%d", u.jobID)
	if u.loggerType == dataIngestion {
		__antithesis_instrumentation__.Notify(494526)
		logFileName = path.Join(logFileName, pgDumpUnsupportedDataStmtLog, fmt.Sprintf("%d.log", u.flushCount))
	} else {
		__antithesis_instrumentation__.Notify(494527)
		logFileName = path.Join(logFileName, pgDumpUnsupportedSchemaStmtLog, fmt.Sprintf("%d.log", u.flushCount))
	}
	__antithesis_instrumentation__.Notify(494518)
	err = cloud.WriteFile(u.ctx, s, logFileName, bytes.NewReader(u.logBuffer.Bytes()))
	if err != nil {
		__antithesis_instrumentation__.Notify(494528)
		return errors.Wrap(err, "failed to log unsupported stmts to log during IMPORT PGDUMP")
	} else {
		__antithesis_instrumentation__.Notify(494529)
	}
	__antithesis_instrumentation__.Notify(494519)
	u.flushCount++
	u.numIgnoredStmts = 0
	u.logBuffer.Truncate(0)
	return nil
}

func init() {
	sql.AddPlanHook("import", importPlanHook)
}
