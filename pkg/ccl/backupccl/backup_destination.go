package backupccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/ccl/storageccl"
	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/featureflag"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/ioctx"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

func fetchPreviousBackups(
	ctx context.Context,
	mem *mon.BoundAccount,
	user security.SQLUsername,
	makeCloudStorage cloud.ExternalStorageFromURIFactory,
	prevBackupURIs []string,
	encryptionParams jobspb.BackupEncryptionOptions,
	kmsEnv cloud.KMSEnv,
) ([]BackupManifest, *jobspb.BackupEncryptionOptions, int64, error) {
	__antithesis_instrumentation__.Notify(6887)
	if len(prevBackupURIs) == 0 {
		__antithesis_instrumentation__.Notify(6891)
		return nil, nil, 0, nil
	} else {
		__antithesis_instrumentation__.Notify(6892)
	}
	__antithesis_instrumentation__.Notify(6888)

	baseBackup := prevBackupURIs[0]
	encryptionOptions, err := getEncryptionFromBase(ctx, user, makeCloudStorage, baseBackup,
		encryptionParams, kmsEnv)
	if err != nil {
		__antithesis_instrumentation__.Notify(6893)
		return nil, nil, 0, err
	} else {
		__antithesis_instrumentation__.Notify(6894)
	}
	__antithesis_instrumentation__.Notify(6889)
	prevBackups, size, err := getBackupManifests(ctx, mem, user, makeCloudStorage, prevBackupURIs,
		encryptionOptions)
	if err != nil {
		__antithesis_instrumentation__.Notify(6895)
		return nil, nil, 0, err
	} else {
		__antithesis_instrumentation__.Notify(6896)
	}
	__antithesis_instrumentation__.Notify(6890)

	return prevBackups, encryptionOptions, size, nil
}

func resolveDest(
	ctx context.Context,
	user security.SQLUsername,
	dest jobspb.BackupDetails_Destination,
	endTime hlc.Timestamp,
	incrementalFrom []string,
	execCfg *sql.ExecutorConfig,
) (
	collectionURI string,
	plannedBackupDefaultURI string,

	chosenSuffix string,
	urisByLocalityKV map[string]string,
	prevBackupURIs []string,
	err error,
) {
	__antithesis_instrumentation__.Notify(6897)
	makeCloudStorage := execCfg.DistSQLSrv.ExternalStorageFromURI

	defaultURI, _, err := getURIsByLocalityKV(dest.To, "")
	if err != nil {
		__antithesis_instrumentation__.Notify(6911)
		return "", "", "", nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(6912)
	}
	__antithesis_instrumentation__.Notify(6898)

	chosenSuffix = dest.Subdir

	if chosenSuffix != "" {
		__antithesis_instrumentation__.Notify(6913)

		collectionURI = defaultURI

		if chosenSuffix == latestFileName {
			__antithesis_instrumentation__.Notify(6914)
			latest, err := readLatestFile(ctx, defaultURI, makeCloudStorage, user)
			if err != nil {
				__antithesis_instrumentation__.Notify(6916)
				return "", "", "", nil, nil, err
			} else {
				__antithesis_instrumentation__.Notify(6917)
			}
			__antithesis_instrumentation__.Notify(6915)
			chosenSuffix = latest
		} else {
			__antithesis_instrumentation__.Notify(6918)
		}
	} else {
		__antithesis_instrumentation__.Notify(6919)
	}
	__antithesis_instrumentation__.Notify(6899)

	plannedBackupDefaultURI, urisByLocalityKV, err = getURIsByLocalityKV(dest.To, chosenSuffix)
	if err != nil {
		__antithesis_instrumentation__.Notify(6920)
		return "", "", "", nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(6921)
	}
	__antithesis_instrumentation__.Notify(6900)

	if len(incrementalFrom) != 0 {
		__antithesis_instrumentation__.Notify(6922)

		prevBackupURIs = incrementalFrom
		return collectionURI, plannedBackupDefaultURI, chosenSuffix, urisByLocalityKV, prevBackupURIs, nil
	} else {
		__antithesis_instrumentation__.Notify(6923)
	}
	__antithesis_instrumentation__.Notify(6901)

	defaultStore, err := makeCloudStorage(ctx, plannedBackupDefaultURI, user)
	if err != nil {
		__antithesis_instrumentation__.Notify(6924)
		return "", "", "", nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(6925)
	}
	__antithesis_instrumentation__.Notify(6902)
	defer defaultStore.Close()
	exists, err := containsManifest(ctx, defaultStore)
	if err != nil {
		__antithesis_instrumentation__.Notify(6926)
		return "", "", "", nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(6927)
	}
	__antithesis_instrumentation__.Notify(6903)
	if exists && func() bool {
		__antithesis_instrumentation__.Notify(6928)
		return !dest.Exists == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(6929)
		return chosenSuffix != "" == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(6930)
		return execCfg.Settings.Version.IsActive(ctx,
			clusterversion.Start22_1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(6931)

		return "",
			"",
			"",
			nil,
			nil,
			errors.Newf("A full backup already exists in %s. "+
				"Consider running an incremental backup to this full backup via `BACKUP INTO '%s' IN '%s'`",
				plannedBackupDefaultURI, chosenSuffix, dest.To[0])

	} else {
		__antithesis_instrumentation__.Notify(6932)
		if !exists {
			__antithesis_instrumentation__.Notify(6933)
			if dest.Exists {
				__antithesis_instrumentation__.Notify(6935)

				if err := featureflag.CheckEnabled(
					ctx,
					execCfg,
					featureFullBackupUserSubdir,
					"'Full Backup with user defined subdirectory'",
				); err != nil {
					__antithesis_instrumentation__.Notify(6936)
					return "", "", "", nil, nil, errors.Wrapf(err,
						"The full backup cannot get written to '%s', a user defined subdirectory. "+
							"To take a full backup, remove the subdirectory from the backup command, "+
							"(i.e. run 'BACKUP ... INTO <collectionURI>'). "+
							"Or, to take a full backup at a specific subdirectory, "+
							"enable the deprecated syntax by switching the 'bulkio.backup."+
							"deprecated_full_backup_with_subdir.enable' cluster setting to true; "+
							"however, note this deprecated syntax will not be available in a future release.", chosenSuffix)
				} else {
					__antithesis_instrumentation__.Notify(6937)
				}
			} else {
				__antithesis_instrumentation__.Notify(6938)
			}
			__antithesis_instrumentation__.Notify(6934)

			return collectionURI, plannedBackupDefaultURI, chosenSuffix, urisByLocalityKV, prevBackupURIs, nil
		} else {
			__antithesis_instrumentation__.Notify(6939)
		}
	}
	__antithesis_instrumentation__.Notify(6904)

	fullyResolvedIncrementalsLocation, err := resolveIncrementalsBackupLocation(
		ctx,
		user,
		execCfg,
		dest.IncrementalStorage,
		dest.To,
		chosenSuffix)
	if err != nil {
		__antithesis_instrumentation__.Notify(6940)
		return "", "", "", nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(6941)
	}
	__antithesis_instrumentation__.Notify(6905)

	priorsDefaultURI, _, err := getURIsByLocalityKV(fullyResolvedIncrementalsLocation, "")
	if err != nil {
		__antithesis_instrumentation__.Notify(6942)
		return "", "", "", nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(6943)
	}
	__antithesis_instrumentation__.Notify(6906)
	incrementalStore, err := makeCloudStorage(ctx, priorsDefaultURI, user)
	if err != nil {
		__antithesis_instrumentation__.Notify(6944)
		return "", "", "", nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(6945)
	}
	__antithesis_instrumentation__.Notify(6907)
	defer incrementalStore.Close()

	priors, err := FindPriorBackups(ctx, incrementalStore, OmitManifest)
	if err != nil {
		__antithesis_instrumentation__.Notify(6946)
		return "", "", "", nil, nil, errors.Wrap(err, "adjusting backup destination to append new layer to existing backup")
	} else {
		__antithesis_instrumentation__.Notify(6947)
	}
	__antithesis_instrumentation__.Notify(6908)

	for _, prior := range priors {
		__antithesis_instrumentation__.Notify(6948)
		priorURI, err := url.Parse(priorsDefaultURI)
		if err != nil {
			__antithesis_instrumentation__.Notify(6950)
			return "", "", "", nil, nil, errors.Wrapf(err, "parsing default backup location %s",
				priorsDefaultURI)
		} else {
			__antithesis_instrumentation__.Notify(6951)
		}
		__antithesis_instrumentation__.Notify(6949)
		priorURI.Path = JoinURLPath(priorURI.Path, prior)
		prevBackupURIs = append(prevBackupURIs, priorURI.String())
	}
	__antithesis_instrumentation__.Notify(6909)
	prevBackupURIs = append([]string{plannedBackupDefaultURI}, prevBackupURIs...)

	partName := endTime.GoTime().Format(DateBasedIncFolderName)
	defaultIncrementalsURI, urisByLocalityKV, err := getURIsByLocalityKV(fullyResolvedIncrementalsLocation, partName)
	if err != nil {
		__antithesis_instrumentation__.Notify(6952)
		return "", "", "", nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(6953)
	}
	__antithesis_instrumentation__.Notify(6910)
	return collectionURI, defaultIncrementalsURI, chosenSuffix, urisByLocalityKV, prevBackupURIs, nil
}

func getBackupManifests(
	ctx context.Context,
	mem *mon.BoundAccount,
	user security.SQLUsername,
	makeCloudStorage cloud.ExternalStorageFromURIFactory,
	backupURIs []string,
	encryption *jobspb.BackupEncryptionOptions,
) ([]BackupManifest, int64, error) {
	__antithesis_instrumentation__.Notify(6954)
	manifests := make([]BackupManifest, len(backupURIs))
	if len(backupURIs) == 0 {
		__antithesis_instrumentation__.Notify(6958)
		return manifests, 0, nil
	} else {
		__antithesis_instrumentation__.Notify(6959)
	}
	__antithesis_instrumentation__.Notify(6955)

	memMu := struct {
		syncutil.Mutex
		total int64
		mem   *mon.BoundAccount
	}{}
	memMu.mem = mem

	g := ctxgroup.WithContext(ctx)
	for i := range backupURIs {
		__antithesis_instrumentation__.Notify(6960)
		i := i

		subMem := mem.Monitor().MakeBoundAccount()
		g.GoCtx(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(6961)
			defer subMem.Close(ctx)

			uri := backupURIs[i]
			desc, size, err := ReadBackupManifestFromURI(
				ctx, &subMem, uri, user, makeCloudStorage, encryption,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(6964)
				return errors.Wrapf(err, "failed to read backup from %q",
					RedactURIForErrorMessage(uri))
			} else {
				__antithesis_instrumentation__.Notify(6965)
			}
			__antithesis_instrumentation__.Notify(6962)

			memMu.Lock()
			err = memMu.mem.Grow(ctx, size)

			if err == nil {
				__antithesis_instrumentation__.Notify(6966)
				memMu.total += size
				manifests[i] = desc
			} else {
				__antithesis_instrumentation__.Notify(6967)
			}
			__antithesis_instrumentation__.Notify(6963)
			subMem.Shrink(ctx, size)
			memMu.Unlock()

			return err
		})
	}
	__antithesis_instrumentation__.Notify(6956)

	if err := g.Wait(); err != nil {
		__antithesis_instrumentation__.Notify(6968)
		mem.Shrink(ctx, memMu.total)
		return nil, 0, err
	} else {
		__antithesis_instrumentation__.Notify(6969)
	}
	__antithesis_instrumentation__.Notify(6957)

	return manifests, memMu.total, nil
}

func getEncryptionFromBase(
	ctx context.Context,
	user security.SQLUsername,
	makeCloudStorage cloud.ExternalStorageFromURIFactory,
	baseBackupURI string,
	encryptionParams jobspb.BackupEncryptionOptions,
	kmsEnv cloud.KMSEnv,
) (*jobspb.BackupEncryptionOptions, error) {
	__antithesis_instrumentation__.Notify(6970)
	var encryptionOptions *jobspb.BackupEncryptionOptions
	if encryptionParams.Mode != jobspb.EncryptionMode_None {
		__antithesis_instrumentation__.Notify(6972)
		exportStore, err := makeCloudStorage(ctx, baseBackupURI, user)
		if err != nil {
			__antithesis_instrumentation__.Notify(6975)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(6976)
		}
		__antithesis_instrumentation__.Notify(6973)
		defer exportStore.Close()
		opts, err := readEncryptionOptions(ctx, exportStore)
		if err != nil {
			__antithesis_instrumentation__.Notify(6977)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(6978)
		}
		__antithesis_instrumentation__.Notify(6974)

		switch encryptionParams.Mode {
		case jobspb.EncryptionMode_Passphrase:
			__antithesis_instrumentation__.Notify(6979)
			encryptionOptions = &jobspb.BackupEncryptionOptions{
				Mode: jobspb.EncryptionMode_Passphrase,
				Key:  storageccl.GenerateKey([]byte(encryptionParams.RawPassphrae), opts[0].Salt),
			}
		case jobspb.EncryptionMode_KMS:
			__antithesis_instrumentation__.Notify(6980)
			var defaultKMSInfo *jobspb.BackupEncryptionOptions_KMSInfo
			for _, encFile := range opts {
				__antithesis_instrumentation__.Notify(6984)
				defaultKMSInfo, err = validateKMSURIsAgainstFullBackup(encryptionParams.RawKmsUris,
					newEncryptedDataKeyMapFromProtoMap(encFile.EncryptedDataKeyByKMSMasterKeyID), kmsEnv)
				if err == nil {
					__antithesis_instrumentation__.Notify(6985)
					break
				} else {
					__antithesis_instrumentation__.Notify(6986)
				}
			}
			__antithesis_instrumentation__.Notify(6981)
			if err != nil {
				__antithesis_instrumentation__.Notify(6987)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(6988)
			}
			__antithesis_instrumentation__.Notify(6982)
			encryptionOptions = &jobspb.BackupEncryptionOptions{
				Mode:    jobspb.EncryptionMode_KMS,
				KMSInfo: defaultKMSInfo}
		default:
			__antithesis_instrumentation__.Notify(6983)
		}
	} else {
		__antithesis_instrumentation__.Notify(6989)
	}
	__antithesis_instrumentation__.Notify(6971)
	return encryptionOptions, nil
}

func readLatestFile(
	ctx context.Context,
	collectionURI string,
	makeCloudStorage cloud.ExternalStorageFromURIFactory,
	user security.SQLUsername,
) (string, error) {
	__antithesis_instrumentation__.Notify(6990)
	collection, err := makeCloudStorage(ctx, collectionURI, user)
	if err != nil {
		__antithesis_instrumentation__.Notify(6995)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(6996)
	}
	__antithesis_instrumentation__.Notify(6991)
	defer collection.Close()

	latestFile, err := findLatestFile(ctx, collection)

	if err != nil {
		__antithesis_instrumentation__.Notify(6997)
		if errors.Is(err, cloud.ErrFileDoesNotExist) {
			__antithesis_instrumentation__.Notify(6999)
			return "", pgerror.Wrapf(err, pgcode.UndefinedFile, "path does not contain a completed latest backup")
		} else {
			__antithesis_instrumentation__.Notify(7000)
		}
		__antithesis_instrumentation__.Notify(6998)
		return "", pgerror.WithCandidateCode(err, pgcode.Io)
	} else {
		__antithesis_instrumentation__.Notify(7001)
	}
	__antithesis_instrumentation__.Notify(6992)
	latest, err := ioctx.ReadAll(ctx, latestFile)
	if err != nil {
		__antithesis_instrumentation__.Notify(7002)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(7003)
	}
	__antithesis_instrumentation__.Notify(6993)
	if len(latest) == 0 {
		__antithesis_instrumentation__.Notify(7004)
		return "", errors.Errorf("malformed LATEST file")
	} else {
		__antithesis_instrumentation__.Notify(7005)
	}
	__antithesis_instrumentation__.Notify(6994)
	return string(latest), nil
}

func findLatestFile(
	ctx context.Context, exportStore cloud.ExternalStorage,
) (ioctx.ReadCloserCtx, error) {
	__antithesis_instrumentation__.Notify(7006)
	var latestFile string
	var latestFileFound bool

	err := exportStore.List(ctx, latestHistoryDirectory, "", func(p string) error {
		__antithesis_instrumentation__.Notify(7011)
		p = strings.TrimPrefix(p, "/")
		latestFile = p
		latestFileFound = true

		return cloud.ErrListingDone
	})
	__antithesis_instrumentation__.Notify(7007)

	if errors.Is(err, cloud.ErrListingUnsupported) {
		__antithesis_instrumentation__.Notify(7012)
		r, err := exportStore.ReadFile(ctx, latestHistoryDirectory+"/"+latestFileName)
		if err == nil {
			__antithesis_instrumentation__.Notify(7013)
			return r, nil
		} else {
			__antithesis_instrumentation__.Notify(7014)
		}
	} else {
		__antithesis_instrumentation__.Notify(7015)
		if err != nil && func() bool {
			__antithesis_instrumentation__.Notify(7016)
			return !errors.Is(err, cloud.ErrListingDone) == true
		}() == true {
			__antithesis_instrumentation__.Notify(7017)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(7018)
		}
	}
	__antithesis_instrumentation__.Notify(7008)

	if latestFileFound {
		__antithesis_instrumentation__.Notify(7019)
		return exportStore.ReadFile(ctx, latestHistoryDirectory+"/"+latestFile)
	} else {
		__antithesis_instrumentation__.Notify(7020)
	}
	__antithesis_instrumentation__.Notify(7009)

	r, err := exportStore.ReadFile(ctx, latestFileName)
	if err != nil {
		__antithesis_instrumentation__.Notify(7021)
		return nil, errors.Wrap(err, "LATEST file could not be read in base or metadata directory")
	} else {
		__antithesis_instrumentation__.Notify(7022)
	}
	__antithesis_instrumentation__.Notify(7010)
	return r, nil
}

func writeNewLatestFile(
	ctx context.Context, settings *cluster.Settings, exportStore cloud.ExternalStorage, suffix string,
) error {
	__antithesis_instrumentation__.Notify(7023)

	if !settings.Version.IsActive(ctx, clusterversion.BackupDoesNotOverwriteLatestAndCheckpoint) {
		__antithesis_instrumentation__.Notify(7026)
		return cloud.WriteFile(ctx, exportStore, latestFileName, strings.NewReader(suffix))
	} else {
		__antithesis_instrumentation__.Notify(7027)
	}
	__antithesis_instrumentation__.Notify(7024)

	if exportStore.Conf().Provider == roachpb.ExternalStorageProvider_http {
		__antithesis_instrumentation__.Notify(7028)
		return cloud.WriteFile(ctx, exportStore, latestFileName, strings.NewReader(suffix))
	} else {
		__antithesis_instrumentation__.Notify(7029)
	}
	__antithesis_instrumentation__.Notify(7025)

	return cloud.WriteFile(ctx, exportStore, newTimestampedLatestFileName(), strings.NewReader(suffix))
}

func newTimestampedLatestFileName() string {
	__antithesis_instrumentation__.Notify(7030)
	var buffer []byte
	buffer = encoding.EncodeStringDescending(buffer, timeutil.Now().String())
	return fmt.Sprintf("%s/%s-%s", latestHistoryDirectory, latestFileName, hex.EncodeToString(buffer))
}
