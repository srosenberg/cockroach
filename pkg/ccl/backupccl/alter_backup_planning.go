package backupccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/featureflag"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
)

func alterBackupPlanHook(
	ctx context.Context, stmt tree.Statement, p sql.PlanHookState,
) (sql.PlanHookRowFn, colinfo.ResultColumns, []sql.PlanNode, bool, error) {
	__antithesis_instrumentation__.Notify(4042)
	alterBackupStmt, ok := stmt.(*tree.AlterBackup)
	if !ok {
		__antithesis_instrumentation__.Notify(4050)
		return nil, nil, nil, false, nil
	} else {
		__antithesis_instrumentation__.Notify(4051)
	}
	__antithesis_instrumentation__.Notify(4043)

	if err := featureflag.CheckEnabled(
		ctx,
		p.ExecCfg(),
		featureBackupEnabled,
		"ALTER BACKUP",
	); err != nil {
		__antithesis_instrumentation__.Notify(4052)
		return nil, nil, nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(4053)
	}
	__antithesis_instrumentation__.Notify(4044)

	fromFn, err := p.TypeAsString(ctx, alterBackupStmt.Backup, "ALTER BACKUP")
	if err != nil {
		__antithesis_instrumentation__.Notify(4054)
		return nil, nil, nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(4055)
	}
	__antithesis_instrumentation__.Notify(4045)

	subdirFn := func() (string, error) { __antithesis_instrumentation__.Notify(4056); return "", nil }
	__antithesis_instrumentation__.Notify(4046)
	if alterBackupStmt.Subdir != nil {
		__antithesis_instrumentation__.Notify(4057)
		subdirFn, err = p.TypeAsString(ctx, alterBackupStmt.Subdir, "ALTER BACKUP")
		if err != nil {
			__antithesis_instrumentation__.Notify(4058)
			return nil, nil, nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(4059)
		}
	} else {
		__antithesis_instrumentation__.Notify(4060)
	}
	__antithesis_instrumentation__.Notify(4047)

	var newKmsFn func() ([]string, error)
	var oldKmsFn func() ([]string, error)

	for _, cmd := range alterBackupStmt.Cmds {
		__antithesis_instrumentation__.Notify(4061)
		switch v := cmd.(type) {
		case *tree.AlterBackupKMS:
			__antithesis_instrumentation__.Notify(4062)
			newKmsFn, err = p.TypeAsStringArray(ctx, tree.Exprs(v.KMSInfo.NewKMSURI), "ALTER BACKUP")
			if err != nil {
				__antithesis_instrumentation__.Notify(4064)
				return nil, nil, nil, false, err
			} else {
				__antithesis_instrumentation__.Notify(4065)
			}
			__antithesis_instrumentation__.Notify(4063)
			oldKmsFn, err = p.TypeAsStringArray(ctx, tree.Exprs(v.KMSInfo.OldKMSURI), "ALTER BACKUP")
			if err != nil {
				__antithesis_instrumentation__.Notify(4066)
				return nil, nil, nil, false, err
			} else {
				__antithesis_instrumentation__.Notify(4067)
			}
		}
	}
	__antithesis_instrumentation__.Notify(4048)

	fn := func(ctx context.Context, _ []sql.PlanNode, resultsCh chan<- tree.Datums) error {
		__antithesis_instrumentation__.Notify(4068)
		backup, err := fromFn()
		if err != nil {
			__antithesis_instrumentation__.Notify(4074)
			return err
		} else {
			__antithesis_instrumentation__.Notify(4075)
		}
		__antithesis_instrumentation__.Notify(4069)

		subdir, err := subdirFn()
		if err != nil {
			__antithesis_instrumentation__.Notify(4076)
			return err
		} else {
			__antithesis_instrumentation__.Notify(4077)
		}
		__antithesis_instrumentation__.Notify(4070)

		if subdir != "" {
			__antithesis_instrumentation__.Notify(4078)
			if strings.EqualFold(subdir, "LATEST") {
				__antithesis_instrumentation__.Notify(4081)

				latest, err := readLatestFile(ctx, backup, p.ExecCfg().DistSQLSrv.ExternalStorageFromURI, p.User())
				if err != nil {
					__antithesis_instrumentation__.Notify(4083)
					return err
				} else {
					__antithesis_instrumentation__.Notify(4084)
				}
				__antithesis_instrumentation__.Notify(4082)
				subdir = latest
			} else {
				__antithesis_instrumentation__.Notify(4085)
			}
			__antithesis_instrumentation__.Notify(4079)

			appendPaths := func(uri string, tailDir string) (string, error) {
				__antithesis_instrumentation__.Notify(4086)
				parsed, err := url.Parse(uri)
				if err != nil {
					__antithesis_instrumentation__.Notify(4088)
					return uri, err
				} else {
					__antithesis_instrumentation__.Notify(4089)
				}
				__antithesis_instrumentation__.Notify(4087)
				parsed.Path = path.Join(parsed.Path, tailDir)
				uri = parsed.String()
				return uri, nil
			}
			__antithesis_instrumentation__.Notify(4080)

			if backup, err = appendPaths(backup, subdir); err != nil {
				__antithesis_instrumentation__.Notify(4090)
				return err
			} else {
				__antithesis_instrumentation__.Notify(4091)
			}
		} else {
			__antithesis_instrumentation__.Notify(4092)
		}
		__antithesis_instrumentation__.Notify(4071)

		var newKms []string
		newKms, err = newKmsFn()
		if err != nil {
			__antithesis_instrumentation__.Notify(4093)
			return err
		} else {
			__antithesis_instrumentation__.Notify(4094)
		}
		__antithesis_instrumentation__.Notify(4072)

		var oldKms []string
		oldKms, err = oldKmsFn()
		if err != nil {
			__antithesis_instrumentation__.Notify(4095)
			return err
		} else {
			__antithesis_instrumentation__.Notify(4096)
		}
		__antithesis_instrumentation__.Notify(4073)

		return doAlterBackupPlan(ctx, alterBackupStmt, p, backup, newKms, oldKms)
	}
	__antithesis_instrumentation__.Notify(4049)

	return fn, nil, nil, false, nil
}

func doAlterBackupPlan(
	ctx context.Context,
	alterBackupStmt *tree.AlterBackup,
	p sql.PlanHookState,
	backup string,
	newKms []string,
	oldKms []string,
) error {
	__antithesis_instrumentation__.Notify(4097)
	if len(backup) < 1 {
		__antithesis_instrumentation__.Notify(4106)
		return errors.New("invalid base backup specified")
	} else {
		__antithesis_instrumentation__.Notify(4107)
	}
	__antithesis_instrumentation__.Notify(4098)

	baseStore, err := p.ExecCfg().DistSQLSrv.ExternalStorageFromURI(ctx, backup, p.User())
	if err != nil {
		__antithesis_instrumentation__.Notify(4108)
		return errors.Wrapf(err, "failed to open backup storage location")
	} else {
		__antithesis_instrumentation__.Notify(4109)
	}
	__antithesis_instrumentation__.Notify(4099)
	defer baseStore.Close()

	opts, err := readEncryptionOptions(ctx, baseStore)
	if err != nil {
		__antithesis_instrumentation__.Notify(4110)
		return err
	} else {
		__antithesis_instrumentation__.Notify(4111)
	}
	__antithesis_instrumentation__.Notify(4100)

	ioConf := baseStore.ExternalIOConf()

	var defaultKMSInfo *jobspb.BackupEncryptionOptions_KMSInfo
	oldKMSFound := false
	for _, old := range oldKms {
		__antithesis_instrumentation__.Notify(4112)
		for _, encFile := range opts {
			__antithesis_instrumentation__.Notify(4114)
			defaultKMSInfo, err = validateKMSURIsAgainstFullBackup([]string{old},
				newEncryptedDataKeyMapFromProtoMap(encFile.EncryptedDataKeyByKMSMasterKeyID), &backupKMSEnv{
					baseStore.Settings(),
					&ioConf,
				})

			if err == nil {
				__antithesis_instrumentation__.Notify(4115)
				oldKMSFound = true
				break
			} else {
				__antithesis_instrumentation__.Notify(4116)
			}
		}
		__antithesis_instrumentation__.Notify(4113)
		if oldKMSFound {
			__antithesis_instrumentation__.Notify(4117)
			break
		} else {
			__antithesis_instrumentation__.Notify(4118)
		}
	}
	__antithesis_instrumentation__.Notify(4101)
	if !oldKMSFound {
		__antithesis_instrumentation__.Notify(4119)
		return errors.New("no key in OLD_KMS matches a key that was previously used to encrypt the backup")
	} else {
		__antithesis_instrumentation__.Notify(4120)
	}
	__antithesis_instrumentation__.Notify(4102)

	encryption := &jobspb.BackupEncryptionOptions{
		Mode:    jobspb.EncryptionMode_KMS,
		KMSInfo: defaultKMSInfo}

	var plaintextDataKey []byte
	plaintextDataKey, err = getEncryptionKey(ctx, encryption, baseStore.Settings(),
		baseStore.ExternalIOConf())
	if err != nil {
		__antithesis_instrumentation__.Notify(4121)
		return err
	} else {
		__antithesis_instrumentation__.Notify(4122)
	}
	__antithesis_instrumentation__.Notify(4103)

	kmsEnv := &backupKMSEnv{settings: p.ExecCfg().Settings, conf: &p.ExecCfg().ExternalIODirConfig}

	encryptedDataKeyByKMSMasterKeyID := newEncryptedDataKeyMap()

	for _, kmsURI := range newKms {
		__antithesis_instrumentation__.Notify(4123)
		masterKeyID, encryptedDataKey, err := getEncryptedDataKeyFromURI(ctx,
			plaintextDataKey, kmsURI, kmsEnv)
		if err != nil {
			__antithesis_instrumentation__.Notify(4125)
			return errors.Wrap(err, "failed to encrypt data key when adding new KMS")
		} else {
			__antithesis_instrumentation__.Notify(4126)
		}
		__antithesis_instrumentation__.Notify(4124)

		encryptedDataKeyByKMSMasterKeyID.addEncryptedDataKey(plaintextMasterKeyID(masterKeyID),
			encryptedDataKey)
	}
	__antithesis_instrumentation__.Notify(4104)

	encryptedDataKeyMapForProto := make(map[string][]byte)
	encryptedDataKeyByKMSMasterKeyID.rangeOverMap(
		func(masterKeyID hashedMasterKeyID, dataKey []byte) {
			__antithesis_instrumentation__.Notify(4127)
			encryptedDataKeyMapForProto[string(masterKeyID)] = dataKey
		})
	__antithesis_instrumentation__.Notify(4105)

	encryptionInfo := &jobspb.EncryptionInfo{EncryptedDataKeyByKMSMasterKeyID: encryptedDataKeyMapForProto}

	return writeNewEncryptionInfoToBackup(ctx, encryptionInfo, baseStore, len(opts))
}

func writeNewEncryptionInfoToBackup(
	ctx context.Context, opts *jobspb.EncryptionInfo, dest cloud.ExternalStorage, numFiles int,
) error {
	__antithesis_instrumentation__.Notify(4128)

	newEncryptionInfoFile := fmt.Sprintf("%s-%d", backupEncryptionInfoFile, numFiles+1)

	buf, err := protoutil.Marshal(opts)
	if err != nil {
		__antithesis_instrumentation__.Notify(4130)
		return err
	} else {
		__antithesis_instrumentation__.Notify(4131)
	}
	__antithesis_instrumentation__.Notify(4129)
	return cloud.WriteFile(ctx, dest, newEncryptionInfoFile, bytes.NewReader(buf))
}

func init() {
	sql.AddPlanHook("alter backup", alterBackupPlanHook)
}
