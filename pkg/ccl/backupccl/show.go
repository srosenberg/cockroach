package backupccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/ccl/storageccl"
	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descbuilder"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgnotice"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/protoreflect"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/json"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
)

func checkShowBackupURIPrivileges(ctx context.Context, p sql.PlanHookState, uri string) error {
	__antithesis_instrumentation__.Notify(12794)
	conf, err := cloud.ExternalStorageConfFromURI(uri, p.User())
	if err != nil {
		__antithesis_instrumentation__.Notify(12800)
		return err
	} else {
		__antithesis_instrumentation__.Notify(12801)
	}
	__antithesis_instrumentation__.Notify(12795)
	if conf.AccessIsWithExplicitAuth() {
		__antithesis_instrumentation__.Notify(12802)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(12803)
	}
	__antithesis_instrumentation__.Notify(12796)
	if p.ExecCfg().ExternalIODirConfig.EnableNonAdminImplicitAndArbitraryOutbound {
		__antithesis_instrumentation__.Notify(12804)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(12805)
	}
	__antithesis_instrumentation__.Notify(12797)
	hasAdmin, err := p.HasAdminRole(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(12806)
		return err
	} else {
		__antithesis_instrumentation__.Notify(12807)
	}
	__antithesis_instrumentation__.Notify(12798)
	if !hasAdmin {
		__antithesis_instrumentation__.Notify(12808)
		return pgerror.Newf(
			pgcode.InsufficientPrivilege,
			"only users with the admin role are allowed to SHOW BACKUP from the specified %s URI",
			conf.Provider.String())
	} else {
		__antithesis_instrumentation__.Notify(12809)
	}
	__antithesis_instrumentation__.Notify(12799)
	return nil
}

type backupInfoReader interface {
	showBackup(
		context.Context,
		*mon.BoundAccount,
		cloud.ExternalStorage,
		cloud.ExternalStorage,
		*jobspb.BackupEncryptionOptions,
		[]string,
		chan<- tree.Datums,
	) error
	header() colinfo.ResultColumns
}

type manifestInfoReader struct {
	shower backupShower
}

var _ backupInfoReader = manifestInfoReader{}

func (m manifestInfoReader) header() colinfo.ResultColumns {
	__antithesis_instrumentation__.Notify(12810)
	return m.shower.header
}

func (m manifestInfoReader) showBackup(
	ctx context.Context,
	mem *mon.BoundAccount,
	store cloud.ExternalStorage,
	incStore cloud.ExternalStorage,
	enc *jobspb.BackupEncryptionOptions,
	incPaths []string,
	resultsCh chan<- tree.Datums,
) error {
	__antithesis_instrumentation__.Notify(12811)
	var memSize int64
	defer func() {
		__antithesis_instrumentation__.Notify(12818)
		mem.Shrink(ctx, memSize)
	}()
	__antithesis_instrumentation__.Notify(12812)

	var err error
	manifests := make([]BackupManifest, len(incPaths)+1)
	manifests[0], memSize, err = ReadBackupManifestFromStore(ctx, mem, store, enc)

	if err != nil {
		__antithesis_instrumentation__.Notify(12819)
		if errors.Is(err, cloud.ErrFileDoesNotExist) {
			__antithesis_instrumentation__.Notify(12821)
			latestFileExists, errLatestFile := checkForLatestFileInCollection(ctx, store)

			if errLatestFile == nil && func() bool {
				__antithesis_instrumentation__.Notify(12823)
				return latestFileExists == true
			}() == true {
				__antithesis_instrumentation__.Notify(12824)
				return errors.WithHintf(err, "The specified path is the root of a backup collection. "+
					"Use SHOW BACKUPS IN with this path to list all the backup subdirectories in the"+
					" collection. SHOW BACKUP can be used with any of these subdirectories to inspect a"+
					" backup.")
			} else {
				__antithesis_instrumentation__.Notify(12825)
			}
			__antithesis_instrumentation__.Notify(12822)
			return errors.CombineErrors(err, errLatestFile)
		} else {
			__antithesis_instrumentation__.Notify(12826)
		}
		__antithesis_instrumentation__.Notify(12820)
		return err
	} else {
		__antithesis_instrumentation__.Notify(12827)
	}
	__antithesis_instrumentation__.Notify(12813)

	for i := range incPaths {
		__antithesis_instrumentation__.Notify(12828)
		m, sz, err := readBackupManifest(ctx, mem, incStore, incPaths[i], enc)
		if err != nil {
			__antithesis_instrumentation__.Notify(12830)
			return err
		} else {
			__antithesis_instrumentation__.Notify(12831)
		}
		__antithesis_instrumentation__.Notify(12829)
		memSize += sz

		m.DeprecatedStatistics = nil
		manifests[i+1] = m
	}
	__antithesis_instrumentation__.Notify(12814)

	err = maybeUpgradeDescriptorsInBackupManifests(manifests, true)
	if err != nil {
		__antithesis_instrumentation__.Notify(12832)
		return err
	} else {
		__antithesis_instrumentation__.Notify(12833)
	}
	__antithesis_instrumentation__.Notify(12815)

	datums, err := m.shower.fn(manifests)
	if err != nil {
		__antithesis_instrumentation__.Notify(12834)
		return err
	} else {
		__antithesis_instrumentation__.Notify(12835)
	}
	__antithesis_instrumentation__.Notify(12816)

	for _, row := range datums {
		__antithesis_instrumentation__.Notify(12836)
		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(12837)
			return ctx.Err()
		case resultsCh <- row:
			__antithesis_instrumentation__.Notify(12838)
		}
	}
	__antithesis_instrumentation__.Notify(12817)
	return nil
}

type metadataSSTInfoReader struct{}

var _ backupInfoReader = manifestInfoReader{}

func (m metadataSSTInfoReader) header() colinfo.ResultColumns {
	__antithesis_instrumentation__.Notify(12839)
	return colinfo.ResultColumns{
		{Name: "file", Typ: types.String},
		{Name: "key", Typ: types.String},
		{Name: "detail", Typ: types.Jsonb},
	}
}

func (m metadataSSTInfoReader) showBackup(
	ctx context.Context,
	mem *mon.BoundAccount,
	store cloud.ExternalStorage,
	incStore cloud.ExternalStorage,
	enc *jobspb.BackupEncryptionOptions,
	incPaths []string,
	resultsCh chan<- tree.Datums,
) error {
	__antithesis_instrumentation__.Notify(12840)
	filename := metadataSSTName
	push := func(_, readable string, value json.JSON) error {
		__antithesis_instrumentation__.Notify(12844)
		val := tree.DNull
		if value != nil {
			__antithesis_instrumentation__.Notify(12846)
			val = tree.NewDJSON(value)
		} else {
			__antithesis_instrumentation__.Notify(12847)
		}
		__antithesis_instrumentation__.Notify(12845)
		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(12848)
			return ctx.Err()
		case resultsCh <- []tree.Datum{tree.NewDString(filename), tree.NewDString(readable), val}:
			__antithesis_instrumentation__.Notify(12849)
			return nil
		}
	}
	__antithesis_instrumentation__.Notify(12841)

	if err := DebugDumpMetadataSST(ctx, store, filename, enc, push); err != nil {
		__antithesis_instrumentation__.Notify(12850)
		return err
	} else {
		__antithesis_instrumentation__.Notify(12851)
	}
	__antithesis_instrumentation__.Notify(12842)

	for _, i := range incPaths {
		__antithesis_instrumentation__.Notify(12852)
		filename = strings.TrimSuffix(i, backupManifestName) + metadataSSTName
		if err := DebugDumpMetadataSST(ctx, incStore, filename, enc, push); err != nil {
			__antithesis_instrumentation__.Notify(12853)
			return err
		} else {
			__antithesis_instrumentation__.Notify(12854)
		}
	}
	__antithesis_instrumentation__.Notify(12843)
	return nil
}

func showBackupPlanHook(
	ctx context.Context, stmt tree.Statement, p sql.PlanHookState,
) (sql.PlanHookRowFn, colinfo.ResultColumns, []sql.PlanNode, bool, error) {
	__antithesis_instrumentation__.Notify(12855)
	backup, ok := stmt.(*tree.ShowBackup)
	if !ok {
		__antithesis_instrumentation__.Notify(12864)
		return nil, nil, nil, false, nil
	} else {
		__antithesis_instrumentation__.Notify(12865)
	}
	__antithesis_instrumentation__.Notify(12856)

	if backup.Path == nil && func() bool {
		__antithesis_instrumentation__.Notify(12866)
		return backup.InCollection != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(12867)
		return showBackupsInCollectionPlanHook(ctx, backup, p)
	} else {
		__antithesis_instrumentation__.Notify(12868)
	}
	__antithesis_instrumentation__.Notify(12857)

	toFn, err := p.TypeAsString(ctx, backup.Path, "SHOW BACKUP")
	if err != nil {
		__antithesis_instrumentation__.Notify(12869)
		return nil, nil, nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(12870)
	}
	__antithesis_instrumentation__.Notify(12858)

	var inColFn func() (string, error)
	if backup.InCollection != nil {
		__antithesis_instrumentation__.Notify(12871)
		inColFn, err = p.TypeAsString(ctx, backup.InCollection, "SHOW BACKUP")
		if err != nil {
			__antithesis_instrumentation__.Notify(12872)
			return nil, nil, nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(12873)
		}
	} else {
		__antithesis_instrumentation__.Notify(12874)
	}
	__antithesis_instrumentation__.Notify(12859)

	expected := map[string]sql.KVStringOptValidate{
		backupOptEncPassphrase:    sql.KVStringOptRequireValue,
		backupOptEncKMS:           sql.KVStringOptRequireValue,
		backupOptWithPrivileges:   sql.KVStringOptRequireNoValue,
		backupOptAsJSON:           sql.KVStringOptRequireNoValue,
		backupOptWithDebugIDs:     sql.KVStringOptRequireNoValue,
		backupOptIncStorage:       sql.KVStringOptRequireValue,
		backupOptDebugMetadataSST: sql.KVStringOptRequireNoValue,
		backupOptEncDir:           sql.KVStringOptRequireValue,
	}
	optsFn, err := p.TypeAsStringOpts(ctx, backup.Options, expected)
	if err != nil {
		__antithesis_instrumentation__.Notify(12875)
		return nil, nil, nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(12876)
	}
	__antithesis_instrumentation__.Notify(12860)
	opts, err := optsFn()
	if err != nil {
		__antithesis_instrumentation__.Notify(12877)
		return nil, nil, nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(12878)
	}
	__antithesis_instrumentation__.Notify(12861)

	var infoReader backupInfoReader
	if _, dumpSST := opts[backupOptDebugMetadataSST]; dumpSST {
		__antithesis_instrumentation__.Notify(12879)
		infoReader = metadataSSTInfoReader{}
	} else {
		__antithesis_instrumentation__.Notify(12880)
		if _, asJSON := opts[backupOptAsJSON]; asJSON {
			__antithesis_instrumentation__.Notify(12881)
			infoReader = manifestInfoReader{shower: jsonShower}
		} else {
			__antithesis_instrumentation__.Notify(12882)
			var shower backupShower
			switch backup.Details {
			case tree.BackupRangeDetails:
				__antithesis_instrumentation__.Notify(12884)
				shower = backupShowerRanges
			case tree.BackupFileDetails:
				__antithesis_instrumentation__.Notify(12885)
				shower = backupShowerFiles
			case tree.BackupSchemaDetails:
				__antithesis_instrumentation__.Notify(12886)
				shower = backupShowerDefault(ctx, p, true, opts)
			default:
				__antithesis_instrumentation__.Notify(12887)
				shower = backupShowerDefault(ctx, p, false, opts)
			}
			__antithesis_instrumentation__.Notify(12883)
			infoReader = manifestInfoReader{shower}
		}
	}
	__antithesis_instrumentation__.Notify(12862)

	fn := func(ctx context.Context, _ []sql.PlanNode, resultsCh chan<- tree.Datums) error {
		__antithesis_instrumentation__.Notify(12888)

		ctx, span := tracing.ChildSpan(ctx, stmt.StatementTag())
		defer span.Finish()

		dest, err := toFn()
		if err != nil {
			__antithesis_instrumentation__.Notify(12900)
			return err
		} else {
			__antithesis_instrumentation__.Notify(12901)
		}
		__antithesis_instrumentation__.Notify(12889)

		var subdir string

		if inColFn != nil {
			__antithesis_instrumentation__.Notify(12902)
			subdir = dest
			dest, err = inColFn()
			if err != nil {
				__antithesis_instrumentation__.Notify(12903)
				return err
			} else {
				__antithesis_instrumentation__.Notify(12904)
			}
		} else {
			__antithesis_instrumentation__.Notify(12905)

			p.BufferClientNotice(ctx,
				pgnotice.Newf("The `SHOW BACKUP` syntax without the `IN` keyword will be removed in a"+
					" future release. Please switch over to using `SHOW BACKUP FROM <backup> IN"+
					" <collection>` to view metadata on a backup collection: %s."+
					" Also note that backups created using the `BACKUP TO` syntax may not be showable or"+
					" restoreable in the next major version release. Use `BACKUP INTO` instead.",
					"https://www.cockroachlabs.com/docs/stable/show-backup.html"))
		}
		__antithesis_instrumentation__.Notify(12890)
		if err := checkShowBackupURIPrivileges(ctx, p, dest); err != nil {
			__antithesis_instrumentation__.Notify(12906)
			return err
		} else {
			__antithesis_instrumentation__.Notify(12907)
		}
		__antithesis_instrumentation__.Notify(12891)

		fullyResolvedDest := dest
		if subdir != "" {
			__antithesis_instrumentation__.Notify(12908)
			if strings.EqualFold(subdir, latestFileName) {
				__antithesis_instrumentation__.Notify(12911)
				subdir, err = readLatestFile(ctx, dest, p.ExecCfg().DistSQLSrv.ExternalStorageFromURI, p.User())
				if err != nil {
					__antithesis_instrumentation__.Notify(12912)
					return errors.Wrap(err, "read LATEST path")
				} else {
					__antithesis_instrumentation__.Notify(12913)
				}
			} else {
				__antithesis_instrumentation__.Notify(12914)
			}
			__antithesis_instrumentation__.Notify(12909)
			parsed, err := url.Parse(dest)
			if err != nil {
				__antithesis_instrumentation__.Notify(12915)
				return err
			} else {
				__antithesis_instrumentation__.Notify(12916)
			}
			__antithesis_instrumentation__.Notify(12910)
			initialPath := parsed.Path
			parsed.Path = JoinURLPath(initialPath, subdir)
			fullyResolvedDest = parsed.String()
		} else {
			__antithesis_instrumentation__.Notify(12917)
		}
		__antithesis_instrumentation__.Notify(12892)

		store, err := p.ExecCfg().DistSQLSrv.ExternalStorageFromURI(ctx, fullyResolvedDest, p.User())
		if err != nil {
			__antithesis_instrumentation__.Notify(12918)
			return errors.Wrapf(err, "make storage")
		} else {
			__antithesis_instrumentation__.Notify(12919)
		}
		__antithesis_instrumentation__.Notify(12893)
		defer store.Close()

		encStore := store
		if encDir, ok := opts[backupOptEncDir]; ok {
			__antithesis_instrumentation__.Notify(12920)
			encStore, err = p.ExecCfg().DistSQLSrv.ExternalStorageFromURI(ctx, encDir, p.User())
			if err != nil {
				__antithesis_instrumentation__.Notify(12922)
				return errors.Wrap(err, "make storage")
			} else {
				__antithesis_instrumentation__.Notify(12923)
			}
			__antithesis_instrumentation__.Notify(12921)
			defer encStore.Close()
		} else {
			__antithesis_instrumentation__.Notify(12924)
		}
		__antithesis_instrumentation__.Notify(12894)
		var encryption *jobspb.BackupEncryptionOptions
		showEncErr := `If you are running SHOW BACKUP exclusively on an incremental backup, 
you must pass the 'encryption_info_dir' parameter that points to the directory of your full backup`
		if passphrase, ok := opts[backupOptEncPassphrase]; ok {
			__antithesis_instrumentation__.Notify(12925)
			opts, err := readEncryptionOptions(ctx, encStore)
			if errors.Is(err, errEncryptionInfoRead) {
				__antithesis_instrumentation__.Notify(12928)
				return errors.WithHint(err, showEncErr)
			} else {
				__antithesis_instrumentation__.Notify(12929)
			}
			__antithesis_instrumentation__.Notify(12926)
			if err != nil {
				__antithesis_instrumentation__.Notify(12930)
				return err
			} else {
				__antithesis_instrumentation__.Notify(12931)
			}
			__antithesis_instrumentation__.Notify(12927)
			encryptionKey := storageccl.GenerateKey([]byte(passphrase), opts[0].Salt)
			encryption = &jobspb.BackupEncryptionOptions{
				Mode: jobspb.EncryptionMode_Passphrase,
				Key:  encryptionKey,
			}
		} else {
			__antithesis_instrumentation__.Notify(12932)
			if kms, ok := opts[backupOptEncKMS]; ok {
				__antithesis_instrumentation__.Notify(12933)
				opts, err := readEncryptionOptions(ctx, encStore)
				if errors.Is(err, errEncryptionInfoRead) {
					__antithesis_instrumentation__.Notify(12938)
					return errors.WithHint(err, showEncErr)
				} else {
					__antithesis_instrumentation__.Notify(12939)
				}
				__antithesis_instrumentation__.Notify(12934)
				if err != nil {
					__antithesis_instrumentation__.Notify(12940)
					return err
				} else {
					__antithesis_instrumentation__.Notify(12941)
				}
				__antithesis_instrumentation__.Notify(12935)

				env := &backupKMSEnv{p.ExecCfg().Settings, &p.ExecCfg().ExternalIODirConfig}
				var defaultKMSInfo *jobspb.BackupEncryptionOptions_KMSInfo
				for _, encFile := range opts {
					__antithesis_instrumentation__.Notify(12942)
					defaultKMSInfo, err = validateKMSURIsAgainstFullBackup([]string{kms},
						newEncryptedDataKeyMapFromProtoMap(encFile.EncryptedDataKeyByKMSMasterKeyID), env)
					if err == nil {
						__antithesis_instrumentation__.Notify(12943)
						break
					} else {
						__antithesis_instrumentation__.Notify(12944)
					}
				}
				__antithesis_instrumentation__.Notify(12936)
				if err != nil {
					__antithesis_instrumentation__.Notify(12945)
					return err
				} else {
					__antithesis_instrumentation__.Notify(12946)
				}
				__antithesis_instrumentation__.Notify(12937)
				encryption = &jobspb.BackupEncryptionOptions{
					Mode:    jobspb.EncryptionMode_KMS,
					KMSInfo: defaultKMSInfo,
				}
			} else {
				__antithesis_instrumentation__.Notify(12947)
			}
		}
		__antithesis_instrumentation__.Notify(12895)
		explicitIncPaths := make([]string, 0)
		explicitIncPath := opts[backupOptIncStorage]
		if len(explicitIncPath) > 0 {
			__antithesis_instrumentation__.Notify(12948)
			explicitIncPaths = append(explicitIncPaths, explicitIncPath)
		} else {
			__antithesis_instrumentation__.Notify(12949)
		}
		__antithesis_instrumentation__.Notify(12896)

		collection, computedSubdir := CollectionAndSubdir(dest, subdir)
		incLocations, err := resolveIncrementalsBackupLocation(
			ctx,
			p.User(),
			p.ExecCfg(),
			explicitIncPaths,
			[]string{collection},
			computedSubdir,
		)
		var incPaths []string
		var incStore cloud.ExternalStorage
		if err != nil {
			__antithesis_instrumentation__.Notify(12950)
			if errors.Is(err, cloud.ErrListingUnsupported) {
				__antithesis_instrumentation__.Notify(12951)

				log.Warningf(
					ctx, "storage sink %v does not support listing, only showing the base backup", explicitIncPaths)
			} else {
				__antithesis_instrumentation__.Notify(12952)
				return err
			}
		} else {
			__antithesis_instrumentation__.Notify(12953)
			if len(incLocations) > 0 {
				__antithesis_instrumentation__.Notify(12954)

				incStore, err = p.ExecCfg().DistSQLSrv.ExternalStorageFromURI(ctx, incLocations[0], p.User())
				if err != nil {
					__antithesis_instrumentation__.Notify(12956)
					return err
				} else {
					__antithesis_instrumentation__.Notify(12957)
				}
				__antithesis_instrumentation__.Notify(12955)
				incPaths, err = FindPriorBackups(ctx, incStore, IncludeManifest)
				if err != nil {
					__antithesis_instrumentation__.Notify(12958)
					return errors.Wrapf(err, "make incremental storage")
				} else {
					__antithesis_instrumentation__.Notify(12959)
				}
			} else {
				__antithesis_instrumentation__.Notify(12960)
			}
		}
		__antithesis_instrumentation__.Notify(12897)
		mem := p.ExecCfg().RootMemoryMonitor.MakeBoundAccount()
		defer mem.Close(ctx)

		if err := infoReader.showBackup(ctx, &mem, store, incStore, encryption, incPaths, resultsCh); err != nil {
			__antithesis_instrumentation__.Notify(12961)
			return err
		} else {
			__antithesis_instrumentation__.Notify(12962)
		}
		__antithesis_instrumentation__.Notify(12898)
		if backup.InCollection == nil {
			__antithesis_instrumentation__.Notify(12963)
			telemetry.Count("show-backup.deprecated-subdir-syntax")
		} else {
			__antithesis_instrumentation__.Notify(12964)
			telemetry.Count("show-backup.collection")
		}
		__antithesis_instrumentation__.Notify(12899)
		return nil
	}
	__antithesis_instrumentation__.Notify(12863)
	return fn, infoReader.header(), nil, false, nil
}

type backupShower struct {
	header colinfo.ResultColumns

	fn func([]BackupManifest) ([]tree.Datums, error)
}

func backupShowerHeaders(showSchemas bool, opts map[string]string) colinfo.ResultColumns {
	__antithesis_instrumentation__.Notify(12965)
	baseHeaders := colinfo.ResultColumns{
		{Name: "database_name", Typ: types.String},
		{Name: "parent_schema_name", Typ: types.String},
		{Name: "object_name", Typ: types.String},
		{Name: "object_type", Typ: types.String},
		{Name: "backup_type", Typ: types.String},
		{Name: "start_time", Typ: types.Timestamp},
		{Name: "end_time", Typ: types.Timestamp},
		{Name: "size_bytes", Typ: types.Int},
		{Name: "rows", Typ: types.Int},
		{Name: "is_full_cluster", Typ: types.Bool},
	}
	if showSchemas {
		__antithesis_instrumentation__.Notify(12969)
		baseHeaders = append(baseHeaders, colinfo.ResultColumn{Name: "create_statement", Typ: types.String})
	} else {
		__antithesis_instrumentation__.Notify(12970)
	}
	__antithesis_instrumentation__.Notify(12966)
	if _, shouldShowPrivleges := opts[backupOptWithPrivileges]; shouldShowPrivleges {
		__antithesis_instrumentation__.Notify(12971)
		baseHeaders = append(baseHeaders, colinfo.ResultColumn{Name: "privileges", Typ: types.String})
		baseHeaders = append(baseHeaders, colinfo.ResultColumn{Name: "owner", Typ: types.String})
	} else {
		__antithesis_instrumentation__.Notify(12972)
	}
	__antithesis_instrumentation__.Notify(12967)

	if _, shouldShowIDs := opts[backupOptWithDebugIDs]; shouldShowIDs {
		__antithesis_instrumentation__.Notify(12973)
		baseHeaders = append(
			colinfo.ResultColumns{
				baseHeaders[0],
				{Name: "database_id", Typ: types.Int},
				baseHeaders[1],
				{Name: "parent_schema_id", Typ: types.Int},
				baseHeaders[2],
				{Name: "object_id", Typ: types.Int},
			},
			baseHeaders[3:]...,
		)
	} else {
		__antithesis_instrumentation__.Notify(12974)
	}
	__antithesis_instrumentation__.Notify(12968)
	return baseHeaders
}

func backupShowerDefault(
	ctx context.Context, p sql.PlanHookState, showSchemas bool, opts map[string]string,
) backupShower {
	__antithesis_instrumentation__.Notify(12975)
	return backupShower{
		header: backupShowerHeaders(showSchemas, opts),
		fn: func(manifests []BackupManifest) ([]tree.Datums, error) {
			__antithesis_instrumentation__.Notify(12976)
			var rows []tree.Datums
			for _, manifest := range manifests {
				__antithesis_instrumentation__.Notify(12978)

				dbIDToName := make(map[descpb.ID]string)
				schemaIDToName := make(map[descpb.ID]string)
				schemaIDToName[keys.PublicSchemaIDForBackup] = catconstants.PublicSchemaName
				for i := range manifest.Descriptors {
					__antithesis_instrumentation__.Notify(12985)
					_, db, _, schema := descpb.FromDescriptor(&manifest.Descriptors[i])
					if db != nil {
						__antithesis_instrumentation__.Notify(12986)
						if _, ok := dbIDToName[db.ID]; !ok {
							__antithesis_instrumentation__.Notify(12987)
							dbIDToName[db.ID] = db.Name
						} else {
							__antithesis_instrumentation__.Notify(12988)
						}
					} else {
						__antithesis_instrumentation__.Notify(12989)
						if schema != nil {
							__antithesis_instrumentation__.Notify(12990)
							if _, ok := schemaIDToName[schema.ID]; !ok {
								__antithesis_instrumentation__.Notify(12991)
								schemaIDToName[schema.ID] = schema.Name
							} else {
								__antithesis_instrumentation__.Notify(12992)
							}
						} else {
							__antithesis_instrumentation__.Notify(12993)
						}
					}
				}
				__antithesis_instrumentation__.Notify(12979)
				tableSizes, err := getTableSizes(manifest.Files)
				if err != nil {
					__antithesis_instrumentation__.Notify(12994)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(12995)
				}
				__antithesis_instrumentation__.Notify(12980)
				backupType := tree.NewDString("full")
				if manifest.isIncremental() {
					__antithesis_instrumentation__.Notify(12996)
					backupType = tree.NewDString("incremental")
				} else {
					__antithesis_instrumentation__.Notify(12997)
				}
				__antithesis_instrumentation__.Notify(12981)
				start := tree.DNull
				end, err := tree.MakeDTimestamp(timeutil.Unix(0, manifest.EndTime.WallTime), time.Nanosecond)
				if err != nil {
					__antithesis_instrumentation__.Notify(12998)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(12999)
				}
				__antithesis_instrumentation__.Notify(12982)
				if manifest.StartTime.WallTime != 0 {
					__antithesis_instrumentation__.Notify(13000)
					start, err = tree.MakeDTimestamp(timeutil.Unix(0, manifest.StartTime.WallTime), time.Nanosecond)
					if err != nil {
						__antithesis_instrumentation__.Notify(13001)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(13002)
					}
				} else {
					__antithesis_instrumentation__.Notify(13003)
				}
				__antithesis_instrumentation__.Notify(12983)
				var row tree.Datums
				for i := range manifest.Descriptors {
					__antithesis_instrumentation__.Notify(13004)
					descriptor := &manifest.Descriptors[i]

					var dbName string
					var parentSchemaName string
					var descriptorType string

					var dbID descpb.ID
					var parentSchemaID descpb.ID

					createStmtDatum := tree.DNull
					dataSizeDatum := tree.DNull
					rowCountDatum := tree.DNull

					desc := descbuilder.NewBuilder(descriptor).BuildExistingMutable()

					descriptorName := desc.GetName()
					switch desc := desc.(type) {
					case catalog.DatabaseDescriptor:
						__antithesis_instrumentation__.Notify(13009)
						descriptorType = "database"
					case catalog.SchemaDescriptor:
						__antithesis_instrumentation__.Notify(13010)
						descriptorType = "schema"
						dbName = dbIDToName[desc.GetParentID()]
						dbID = desc.GetParentID()
					case catalog.TypeDescriptor:
						__antithesis_instrumentation__.Notify(13011)
						descriptorType = "type"
						dbName = dbIDToName[desc.GetParentID()]
						dbID = desc.GetParentID()
						parentSchemaName = schemaIDToName[desc.GetParentSchemaID()]
						parentSchemaID = desc.GetParentSchemaID()
					case catalog.TableDescriptor:
						__antithesis_instrumentation__.Notify(13012)
						descriptorType = "table"
						dbName = dbIDToName[desc.GetParentID()]
						dbID = desc.GetParentID()
						parentSchemaName = schemaIDToName[desc.GetParentSchemaID()]
						parentSchemaID = desc.GetParentSchemaID()
						tableSize := tableSizes[desc.GetID()]
						dataSizeDatum = tree.NewDInt(tree.DInt(tableSize.DataSize))
						rowCountDatum = tree.NewDInt(tree.DInt(tableSize.Rows))

						displayOptions := sql.ShowCreateDisplayOptions{
							FKDisplayMode:  sql.OmitMissingFKClausesFromCreate,
							IgnoreComments: true,
						}
						createStmt, err := p.ShowCreate(ctx, dbName, manifest.Descriptors,
							tabledesc.NewBuilder(desc.TableDesc()).BuildImmutableTable(), displayOptions)
						if err != nil {
							__antithesis_instrumentation__.Notify(13015)

							log.Errorf(ctx, "error while generating create statement: %+v", err)
						} else {
							__antithesis_instrumentation__.Notify(13016)
						}
						__antithesis_instrumentation__.Notify(13013)
						createStmtDatum = nullIfEmpty(createStmt)
					default:
						__antithesis_instrumentation__.Notify(13014)
						descriptorType = "unknown"
					}
					__antithesis_instrumentation__.Notify(13005)

					row = tree.Datums{
						nullIfEmpty(dbName),
						nullIfEmpty(parentSchemaName),
						tree.NewDString(descriptorName),
						tree.NewDString(descriptorType),
						backupType,
						start,
						end,
						dataSizeDatum,
						rowCountDatum,
						tree.MakeDBool(manifest.DescriptorCoverage == tree.AllDescriptors),
					}
					if showSchemas {
						__antithesis_instrumentation__.Notify(13017)
						row = append(row, createStmtDatum)
					} else {
						__antithesis_instrumentation__.Notify(13018)
					}
					__antithesis_instrumentation__.Notify(13006)
					if _, shouldShowPrivileges := opts[backupOptWithPrivileges]; shouldShowPrivileges {
						__antithesis_instrumentation__.Notify(13019)
						row = append(row, tree.NewDString(showPrivileges(descriptor)))
						owner := desc.GetPrivileges().Owner().SQLIdentifier()
						row = append(row, tree.NewDString(owner))
					} else {
						__antithesis_instrumentation__.Notify(13020)
					}
					__antithesis_instrumentation__.Notify(13007)
					if _, shouldShowIDs := opts[backupOptWithDebugIDs]; shouldShowIDs {
						__antithesis_instrumentation__.Notify(13021)

						row = append(
							tree.Datums{
								row[0],
								nullIfZero(dbID),
								row[1],
								nullIfZero(parentSchemaID),
								row[2],
								nullIfZero(desc.GetID()),
							},
							row[3:]...,
						)
					} else {
						__antithesis_instrumentation__.Notify(13022)
					}
					__antithesis_instrumentation__.Notify(13008)
					rows = append(rows, row)
				}
				__antithesis_instrumentation__.Notify(12984)
				for _, t := range manifest.GetTenants() {
					__antithesis_instrumentation__.Notify(13023)
					row := tree.Datums{
						tree.DNull,
						tree.DNull,
						tree.NewDString(roachpb.MakeTenantID(t.ID).String()),
						tree.NewDString("TENANT"),
						backupType,
						start,
						end,
						tree.DNull,
						tree.DNull,
						tree.DNull,
					}
					if showSchemas {
						__antithesis_instrumentation__.Notify(13027)
						row = append(row, tree.DNull)
					} else {
						__antithesis_instrumentation__.Notify(13028)
					}
					__antithesis_instrumentation__.Notify(13024)
					if _, shouldShowPrivileges := opts[backupOptWithPrivileges]; shouldShowPrivileges {
						__antithesis_instrumentation__.Notify(13029)
						row = append(row, tree.DNull)
					} else {
						__antithesis_instrumentation__.Notify(13030)
					}
					__antithesis_instrumentation__.Notify(13025)
					if _, shouldShowIDs := opts[backupOptWithDebugIDs]; shouldShowIDs {
						__antithesis_instrumentation__.Notify(13031)

						row = append(
							tree.Datums{
								row[0],
								tree.DNull,
								row[1],
								tree.DNull,
								row[2],
								tree.NewDInt(tree.DInt(t.ID)),
							},
							row[3:]...,
						)
					} else {
						__antithesis_instrumentation__.Notify(13032)
					}
					__antithesis_instrumentation__.Notify(13026)
					rows = append(rows, row)
				}
			}
			__antithesis_instrumentation__.Notify(12977)
			return rows, nil
		},
	}
}

func getTableSizes(files []BackupManifest_File) (map[descpb.ID]roachpb.RowCount, error) {
	__antithesis_instrumentation__.Notify(13033)
	tableSizes := make(map[descpb.ID]roachpb.RowCount)
	if len(files) == 0 {
		__antithesis_instrumentation__.Notify(13037)
		return tableSizes, nil
	} else {
		__antithesis_instrumentation__.Notify(13038)
	}
	__antithesis_instrumentation__.Notify(13034)
	_, tenantID, err := keys.DecodeTenantPrefix(files[0].Span.Key)
	if err != nil {
		__antithesis_instrumentation__.Notify(13039)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(13040)
	}
	__antithesis_instrumentation__.Notify(13035)
	showCodec := keys.MakeSQLCodec(tenantID)

	for _, file := range files {
		__antithesis_instrumentation__.Notify(13041)

		_, tableID, err := showCodec.DecodeTablePrefix(file.Span.Key)
		if err != nil {
			__antithesis_instrumentation__.Notify(13043)
			continue
		} else {
			__antithesis_instrumentation__.Notify(13044)
		}
		__antithesis_instrumentation__.Notify(13042)
		s := tableSizes[descpb.ID(tableID)]
		s.Add(file.EntryCounts)
		tableSizes[descpb.ID(tableID)] = s
	}
	__antithesis_instrumentation__.Notify(13036)
	return tableSizes, nil
}

func nullIfEmpty(s string) tree.Datum {
	__antithesis_instrumentation__.Notify(13045)
	if s == "" {
		__antithesis_instrumentation__.Notify(13047)
		return tree.DNull
	} else {
		__antithesis_instrumentation__.Notify(13048)
	}
	__antithesis_instrumentation__.Notify(13046)
	return tree.NewDString(s)
}

func nullIfZero(i descpb.ID) tree.Datum {
	__antithesis_instrumentation__.Notify(13049)
	if i == 0 {
		__antithesis_instrumentation__.Notify(13051)
		return tree.DNull
	} else {
		__antithesis_instrumentation__.Notify(13052)
	}
	__antithesis_instrumentation__.Notify(13050)
	return tree.NewDInt(tree.DInt(i))
}

func showPrivileges(descriptor *descpb.Descriptor) string {
	__antithesis_instrumentation__.Notify(13053)
	var privStringBuilder strings.Builder

	b := descbuilder.NewBuilder(descriptor)
	if b == nil {
		__antithesis_instrumentation__.Notify(13058)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(13059)
	}
	__antithesis_instrumentation__.Notify(13054)
	var objectType privilege.ObjectType
	switch b.DescriptorType() {
	case catalog.Database:
		__antithesis_instrumentation__.Notify(13060)
		objectType = privilege.Database
	case catalog.Table:
		__antithesis_instrumentation__.Notify(13061)
		objectType = privilege.Table
	case catalog.Type:
		__antithesis_instrumentation__.Notify(13062)
		objectType = privilege.Type
	case catalog.Schema:
		__antithesis_instrumentation__.Notify(13063)
		objectType = privilege.Schema
	default:
		__antithesis_instrumentation__.Notify(13064)
		return ""
	}
	__antithesis_instrumentation__.Notify(13055)
	privDesc := b.BuildImmutable().GetPrivileges()
	if privDesc == nil {
		__antithesis_instrumentation__.Notify(13065)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(13066)
	}
	__antithesis_instrumentation__.Notify(13056)
	for _, userPriv := range privDesc.Show(objectType) {
		__antithesis_instrumentation__.Notify(13067)
		privs := userPriv.Privileges
		if len(privs) == 0 {
			__antithesis_instrumentation__.Notify(13070)
			continue
		} else {
			__antithesis_instrumentation__.Notify(13071)
		}
		__antithesis_instrumentation__.Notify(13068)
		privStringBuilder.WriteString("GRANT ")

		for j, priv := range privs {
			__antithesis_instrumentation__.Notify(13072)
			if j != 0 {
				__antithesis_instrumentation__.Notify(13074)
				privStringBuilder.WriteString(", ")
			} else {
				__antithesis_instrumentation__.Notify(13075)
			}
			__antithesis_instrumentation__.Notify(13073)
			privStringBuilder.WriteString(priv.Kind.String())
		}
		__antithesis_instrumentation__.Notify(13069)
		privStringBuilder.WriteString(" ON ")
		privStringBuilder.WriteString(descpb.GetDescriptorName(descriptor))
		privStringBuilder.WriteString(" TO ")
		privStringBuilder.WriteString(userPriv.User.SQLIdentifier())
		privStringBuilder.WriteString("; ")
	}
	__antithesis_instrumentation__.Notify(13057)

	return privStringBuilder.String()
}

var backupShowerRanges = backupShower{
	header: colinfo.ResultColumns{
		{Name: "start_pretty", Typ: types.String},
		{Name: "end_pretty", Typ: types.String},
		{Name: "start_key", Typ: types.Bytes},
		{Name: "end_key", Typ: types.Bytes},
	},

	fn: func(manifests []BackupManifest) (rows []tree.Datums, err error) {
		__antithesis_instrumentation__.Notify(13076)
		for _, manifest := range manifests {
			__antithesis_instrumentation__.Notify(13078)
			for _, span := range manifest.Spans {
				__antithesis_instrumentation__.Notify(13079)
				rows = append(rows, tree.Datums{
					tree.NewDString(span.Key.String()),
					tree.NewDString(span.EndKey.String()),
					tree.NewDBytes(tree.DBytes(span.Key)),
					tree.NewDBytes(tree.DBytes(span.EndKey)),
				})
			}
		}
		__antithesis_instrumentation__.Notify(13077)
		return rows, nil
	},
}

var backupShowerFiles = backupShower{
	header: colinfo.ResultColumns{
		{Name: "path", Typ: types.String},
		{Name: "start_pretty", Typ: types.String},
		{Name: "end_pretty", Typ: types.String},
		{Name: "start_key", Typ: types.Bytes},
		{Name: "end_key", Typ: types.Bytes},
		{Name: "size_bytes", Typ: types.Int},
		{Name: "rows", Typ: types.Int},
	},

	fn: func(manifests []BackupManifest) (rows []tree.Datums, err error) {
		__antithesis_instrumentation__.Notify(13080)
		for _, manifest := range manifests {
			__antithesis_instrumentation__.Notify(13082)
			for _, file := range manifest.Files {
				__antithesis_instrumentation__.Notify(13083)
				rows = append(rows, tree.Datums{
					tree.NewDString(file.Path),
					tree.NewDString(file.Span.Key.String()),
					tree.NewDString(file.Span.EndKey.String()),
					tree.NewDBytes(tree.DBytes(file.Span.Key)),
					tree.NewDBytes(tree.DBytes(file.Span.EndKey)),
					tree.NewDInt(tree.DInt(file.EntryCounts.DataSize)),
					tree.NewDInt(tree.DInt(file.EntryCounts.Rows)),
				})
			}
		}
		__antithesis_instrumentation__.Notify(13081)
		return rows, nil
	},
}

var jsonShower = backupShower{
	header: colinfo.ResultColumns{
		{Name: "manifest", Typ: types.Jsonb},
	},

	fn: func(manifests []BackupManifest) ([]tree.Datums, error) {
		__antithesis_instrumentation__.Notify(13084)
		rows := make([]tree.Datums, len(manifests))
		for i, manifest := range manifests {
			__antithesis_instrumentation__.Notify(13086)
			j, err := protoreflect.MessageToJSON(
				&manifest, protoreflect.FmtFlags{EmitDefaults: true, EmitRedacted: true})
			if err != nil {
				__antithesis_instrumentation__.Notify(13088)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(13089)
			}
			__antithesis_instrumentation__.Notify(13087)
			rows[i] = tree.Datums{tree.NewDJSON(j)}
		}
		__antithesis_instrumentation__.Notify(13085)
		return rows, nil
	},
}

func showBackupsInCollectionPlanHook(
	ctx context.Context, backup *tree.ShowBackup, p sql.PlanHookState,
) (sql.PlanHookRowFn, colinfo.ResultColumns, []sql.PlanNode, bool, error) {
	__antithesis_instrumentation__.Notify(13090)

	collectionFn, err := p.TypeAsString(ctx, backup.InCollection, "SHOW BACKUPS")
	if err != nil {
		__antithesis_instrumentation__.Notify(13093)
		return nil, nil, nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(13094)
	}
	__antithesis_instrumentation__.Notify(13091)

	fn := func(ctx context.Context, _ []sql.PlanNode, resultsCh chan<- tree.Datums) error {
		__antithesis_instrumentation__.Notify(13095)
		ctx, span := tracing.ChildSpan(ctx, backup.StatementTag())
		defer span.Finish()

		collection, err := collectionFn()
		if err != nil {
			__antithesis_instrumentation__.Notify(13101)
			return err
		} else {
			__antithesis_instrumentation__.Notify(13102)
		}
		__antithesis_instrumentation__.Notify(13096)

		if err := checkShowBackupURIPrivileges(ctx, p, collection); err != nil {
			__antithesis_instrumentation__.Notify(13103)
			return err
		} else {
			__antithesis_instrumentation__.Notify(13104)
		}
		__antithesis_instrumentation__.Notify(13097)

		store, err := p.ExecCfg().DistSQLSrv.ExternalStorageFromURI(ctx, collection, p.User())
		if err != nil {
			__antithesis_instrumentation__.Notify(13105)
			return errors.Wrapf(err, "connect to external storage")
		} else {
			__antithesis_instrumentation__.Notify(13106)
		}
		__antithesis_instrumentation__.Notify(13098)
		defer store.Close()
		res, err := ListFullBackupsInCollection(ctx, store)
		if err != nil {
			__antithesis_instrumentation__.Notify(13107)
			return err
		} else {
			__antithesis_instrumentation__.Notify(13108)
		}
		__antithesis_instrumentation__.Notify(13099)
		for _, i := range res {
			__antithesis_instrumentation__.Notify(13109)
			resultsCh <- tree.Datums{tree.NewDString(i)}
		}
		__antithesis_instrumentation__.Notify(13100)
		return nil
	}
	__antithesis_instrumentation__.Notify(13092)
	return fn, colinfo.ResultColumns{{Name: "path", Typ: types.String}}, nil, false, nil
}

func init() {
	sql.AddPlanHook("show backup", showBackupPlanHook)
}
