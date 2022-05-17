package backupccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"crypto"
	cryptorand "crypto/rand"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/build"
	"github.com/cockroachdb/cockroach/pkg/ccl/backupccl/backupresolver"
	"github.com/cockroachdb/cockroach/pkg/ccl/storageccl"
	"github.com/cockroachdb/cockroach/pkg/ccl/utilccl"
	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/featureflag"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobsprotectedts"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts/ptpb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/scheduledjobs"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descbuilder"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgnotice"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/interval"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	pbtypes "github.com/gogo/protobuf/types"
)

const (
	backupOptRevisionHistory  = "revision_history"
	backupOptEncPassphrase    = "encryption_passphrase"
	backupOptEncKMS           = "kms"
	backupOptWithPrivileges   = "privileges"
	backupOptAsJSON           = "as_json"
	backupOptWithDebugIDs     = "debug_ids"
	backupOptIncStorage       = "incremental_location"
	localityURLParam          = "COCKROACH_LOCALITY"
	defaultLocalityValue      = "default"
	backupOptDebugMetadataSST = "debug_dump_metadata_sst"
	backupOptEncDir           = "encryption_info_dir"
)

type tableAndIndex struct {
	tableID descpb.ID
	indexID descpb.IndexID
}

type backupKMSEnv struct {
	settings *cluster.Settings
	conf     *base.ExternalIODirConfig
}

var _ cloud.KMSEnv = &backupKMSEnv{}

var featureBackupEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"feature.backup.enabled",
	"set to true to enable backups, false to disable; default is true",
	featureflag.FeatureFlagEnabledDefault,
).WithPublic()

func (p *backupKMSEnv) ClusterSettings() *cluster.Settings {
	__antithesis_instrumentation__.Notify(7857)
	return p.settings
}

func (p *backupKMSEnv) KMSConfig() *base.ExternalIODirConfig {
	__antithesis_instrumentation__.Notify(7858)
	return p.conf
}

type (
	plaintextMasterKeyID string
	hashedMasterKeyID    string
	encryptedDataKeyMap  struct {
		m map[hashedMasterKeyID][]byte
	}
)

var featureFullBackupUserSubdir = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"bulkio.backup.deprecated_full_backup_with_subdir.enabled",
	"when true, a backup command with a user specified subdirectory will create a full backup at"+
		" the subdirectory if no backup already exists at that subdirectory.",
	false,
).WithPublic()

func newEncryptedDataKeyMap() *encryptedDataKeyMap {
	__antithesis_instrumentation__.Notify(7859)
	return &encryptedDataKeyMap{make(map[hashedMasterKeyID][]byte)}
}

func newEncryptedDataKeyMapFromProtoMap(protoDataKeyMap map[string][]byte) *encryptedDataKeyMap {
	__antithesis_instrumentation__.Notify(7860)
	encMap := &encryptedDataKeyMap{make(map[hashedMasterKeyID][]byte)}
	for k, v := range protoDataKeyMap {
		__antithesis_instrumentation__.Notify(7862)
		encMap.m[hashedMasterKeyID(k)] = v
	}
	__antithesis_instrumentation__.Notify(7861)

	return encMap
}

func (e *encryptedDataKeyMap) addEncryptedDataKey(
	masterKeyID plaintextMasterKeyID, encryptedDataKey []byte,
) {
	__antithesis_instrumentation__.Notify(7863)

	hasher := crypto.SHA256.New()
	hasher.Write([]byte(masterKeyID))
	hash := hasher.Sum(nil)
	e.m[hashedMasterKeyID(hash)] = encryptedDataKey
}

func (e *encryptedDataKeyMap) getEncryptedDataKey(
	masterKeyID plaintextMasterKeyID,
) ([]byte, error) {
	__antithesis_instrumentation__.Notify(7864)

	hasher := crypto.SHA256.New()
	hasher.Write([]byte(masterKeyID))
	hash := hasher.Sum(nil)
	var encDataKey []byte
	var ok bool
	if encDataKey, ok = e.m[hashedMasterKeyID(hash)]; !ok {
		__antithesis_instrumentation__.Notify(7866)
		return nil, errors.New("could not find an entry in the encryptedDataKeyMap")
	} else {
		__antithesis_instrumentation__.Notify(7867)
	}
	__antithesis_instrumentation__.Notify(7865)

	return encDataKey, nil
}

func (e *encryptedDataKeyMap) rangeOverMap(fn func(masterKeyID hashedMasterKeyID, dataKey []byte)) {
	__antithesis_instrumentation__.Notify(7868)
	for k, v := range e.m {
		__antithesis_instrumentation__.Notify(7869)
		fn(k, v)
	}
}

func getPublicIndexTableSpans(
	table catalog.TableDescriptor, added map[tableAndIndex]bool, codec keys.SQLCodec,
) ([]roachpb.Span, error) {
	__antithesis_instrumentation__.Notify(7870)
	publicIndexSpans := make([]roachpb.Span, 0)
	if err := catalog.ForEachActiveIndex(table, func(idx catalog.Index) error {
		__antithesis_instrumentation__.Notify(7872)
		key := tableAndIndex{tableID: table.GetID(), indexID: idx.GetID()}
		if added[key] {
			__antithesis_instrumentation__.Notify(7874)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(7875)
		}
		__antithesis_instrumentation__.Notify(7873)
		added[key] = true
		publicIndexSpans = append(publicIndexSpans, table.IndexSpan(codec, idx.GetID()))
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(7876)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(7877)
	}
	__antithesis_instrumentation__.Notify(7871)

	return publicIndexSpans, nil
}

func spansForAllTableIndexes(
	execCfg *sql.ExecutorConfig,
	tables []catalog.TableDescriptor,
	revs []BackupManifest_DescriptorRevision,
) ([]roachpb.Span, error) {
	__antithesis_instrumentation__.Notify(7878)

	added := make(map[tableAndIndex]bool, len(tables))
	sstIntervalTree := interval.NewTree(interval.ExclusiveOverlapper)
	var publicIndexSpans []roachpb.Span
	var err error

	for _, table := range tables {
		__antithesis_instrumentation__.Notify(7883)
		publicIndexSpans, err = getPublicIndexTableSpans(table, added, execCfg.Codec)
		if err != nil {
			__antithesis_instrumentation__.Notify(7885)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(7886)
		}
		__antithesis_instrumentation__.Notify(7884)

		for _, indexSpan := range publicIndexSpans {
			__antithesis_instrumentation__.Notify(7887)
			if err := sstIntervalTree.Insert(intervalSpan(indexSpan), false); err != nil {
				__antithesis_instrumentation__.Notify(7888)
				panic(errors.NewAssertionErrorWithWrappedErrf(err, "IndexSpan"))
			} else {
				__antithesis_instrumentation__.Notify(7889)
			}
		}
	}
	__antithesis_instrumentation__.Notify(7879)

	for _, rev := range revs {
		__antithesis_instrumentation__.Notify(7890)

		rawTbl, _, _, _ := descpb.FromDescriptor(rev.Desc)
		if rawTbl != nil && func() bool {
			__antithesis_instrumentation__.Notify(7891)
			return rawTbl.Public() == true
		}() == true {
			__antithesis_instrumentation__.Notify(7892)
			tbl := tabledesc.NewBuilder(rawTbl).BuildImmutableTable()
			revSpans, err := getPublicIndexTableSpans(tbl, added, execCfg.Codec)
			if err != nil {
				__antithesis_instrumentation__.Notify(7894)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(7895)
			}
			__antithesis_instrumentation__.Notify(7893)

			publicIndexSpans = append(publicIndexSpans, revSpans...)
			for _, indexSpan := range publicIndexSpans {
				__antithesis_instrumentation__.Notify(7896)
				if err := sstIntervalTree.Insert(intervalSpan(indexSpan), false); err != nil {
					__antithesis_instrumentation__.Notify(7897)
					panic(errors.NewAssertionErrorWithWrappedErrf(err, "IndexSpan"))
				} else {
					__antithesis_instrumentation__.Notify(7898)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(7899)
		}
	}
	__antithesis_instrumentation__.Notify(7880)

	var spans []roachpb.Span
	_ = sstIntervalTree.Do(func(r interval.Interface) bool {
		__antithesis_instrumentation__.Notify(7900)
		spans = append(spans, roachpb.Span{
			Key:    roachpb.Key(r.Range().Start),
			EndKey: roachpb.Key(r.Range().End),
		})
		return false
	})
	__antithesis_instrumentation__.Notify(7881)

	mergedSpans, _ := roachpb.MergeSpans(&spans)

	knobs := execCfg.BackupRestoreTestingKnobs
	if knobs != nil && func() bool {
		__antithesis_instrumentation__.Notify(7901)
		return knobs.CaptureResolvedTableDescSpans != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(7902)
		knobs.CaptureResolvedTableDescSpans(mergedSpans)
	} else {
		__antithesis_instrumentation__.Notify(7903)
	}
	__antithesis_instrumentation__.Notify(7882)

	return mergedSpans, nil
}

func getLocalityAndBaseURI(uri, appendPath string) (string, string, error) {
	__antithesis_instrumentation__.Notify(7904)
	parsedURI, err := url.Parse(uri)
	if err != nil {
		__antithesis_instrumentation__.Notify(7906)
		return "", "", err
	} else {
		__antithesis_instrumentation__.Notify(7907)
	}
	__antithesis_instrumentation__.Notify(7905)
	q := parsedURI.Query()
	localityKV := q.Get(localityURLParam)

	q.Del(localityURLParam)
	parsedURI.RawQuery = q.Encode()

	parsedURI.Path = JoinURLPath(parsedURI.Path, appendPath)

	baseURI := parsedURI.String()
	return localityKV, baseURI, nil
}

func getURIsByLocalityKV(
	to []string, appendPath string,
) (defaultURI string, urisByLocalityKV map[string]string, err error) {
	__antithesis_instrumentation__.Notify(7908)
	urisByLocalityKV = make(map[string]string)
	if len(to) == 1 {
		__antithesis_instrumentation__.Notify(7912)
		localityKV, baseURI, err := getLocalityAndBaseURI(to[0], appendPath)
		if err != nil {
			__antithesis_instrumentation__.Notify(7915)
			return "", nil, err
		} else {
			__antithesis_instrumentation__.Notify(7916)
		}
		__antithesis_instrumentation__.Notify(7913)
		if localityKV != "" && func() bool {
			__antithesis_instrumentation__.Notify(7917)
			return localityKV != defaultLocalityValue == true
		}() == true {
			__antithesis_instrumentation__.Notify(7918)
			return "", nil, errors.Errorf("%s %s is invalid for a single BACKUP location",
				localityURLParam, localityKV)
		} else {
			__antithesis_instrumentation__.Notify(7919)
		}
		__antithesis_instrumentation__.Notify(7914)
		return baseURI, urisByLocalityKV, nil
	} else {
		__antithesis_instrumentation__.Notify(7920)
	}
	__antithesis_instrumentation__.Notify(7909)

	for _, uri := range to {
		__antithesis_instrumentation__.Notify(7921)
		localityKV, baseURI, err := getLocalityAndBaseURI(uri, appendPath)
		if err != nil {
			__antithesis_instrumentation__.Notify(7924)
			return "", nil, err
		} else {
			__antithesis_instrumentation__.Notify(7925)
		}
		__antithesis_instrumentation__.Notify(7922)
		if localityKV == "" {
			__antithesis_instrumentation__.Notify(7926)
			return "", nil, errors.Errorf(
				"multiple URLs are provided for partitioned BACKUP, but %s is not specified",
				localityURLParam,
			)
		} else {
			__antithesis_instrumentation__.Notify(7927)
		}
		__antithesis_instrumentation__.Notify(7923)
		if localityKV == defaultLocalityValue {
			__antithesis_instrumentation__.Notify(7928)
			if defaultURI != "" {
				__antithesis_instrumentation__.Notify(7930)
				return "", nil, errors.Errorf("multiple default URLs provided for partition backup")
			} else {
				__antithesis_instrumentation__.Notify(7931)
			}
			__antithesis_instrumentation__.Notify(7929)
			defaultURI = baseURI
		} else {
			__antithesis_instrumentation__.Notify(7932)
			kv := roachpb.Tier{}
			if err := kv.FromString(localityKV); err != nil {
				__antithesis_instrumentation__.Notify(7935)
				return "", nil, errors.Wrap(err, "failed to parse backup locality")
			} else {
				__antithesis_instrumentation__.Notify(7936)
			}
			__antithesis_instrumentation__.Notify(7933)
			if _, ok := urisByLocalityKV[localityKV]; ok {
				__antithesis_instrumentation__.Notify(7937)
				return "", nil, errors.Errorf("duplicate URIs for locality %s", localityKV)
			} else {
				__antithesis_instrumentation__.Notify(7938)
			}
			__antithesis_instrumentation__.Notify(7934)
			urisByLocalityKV[localityKV] = baseURI
		}
	}
	__antithesis_instrumentation__.Notify(7910)
	if defaultURI == "" {
		__antithesis_instrumentation__.Notify(7939)
		return "", nil, errors.Errorf("no default URL provided for partitioned backup")
	} else {
		__antithesis_instrumentation__.Notify(7940)
	}
	__antithesis_instrumentation__.Notify(7911)
	return defaultURI, urisByLocalityKV, nil
}

func resolveOptionsForBackupJobDescription(
	opts tree.BackupOptions, kmsURIs []string, incrementalStorage []string,
) (tree.BackupOptions, error) {
	__antithesis_instrumentation__.Notify(7941)
	if opts.IsDefault() {
		__antithesis_instrumentation__.Notify(7946)
		return opts, nil
	} else {
		__antithesis_instrumentation__.Notify(7947)
	}
	__antithesis_instrumentation__.Notify(7942)

	newOpts := tree.BackupOptions{
		CaptureRevisionHistory: opts.CaptureRevisionHistory,
		Detached:               opts.Detached,
	}

	if opts.EncryptionPassphrase != nil {
		__antithesis_instrumentation__.Notify(7948)
		newOpts.EncryptionPassphrase = tree.NewDString("redacted")
	} else {
		__antithesis_instrumentation__.Notify(7949)
	}
	__antithesis_instrumentation__.Notify(7943)

	var err error

	newOpts.EncryptionKMSURI, err = sanitizeURIList(kmsURIs)
	if err != nil {
		__antithesis_instrumentation__.Notify(7950)
		return tree.BackupOptions{}, err
	} else {
		__antithesis_instrumentation__.Notify(7951)
	}
	__antithesis_instrumentation__.Notify(7944)

	newOpts.IncrementalStorage, err = sanitizeURIList(incrementalStorage)
	if err != nil {
		__antithesis_instrumentation__.Notify(7952)
		return tree.BackupOptions{}, err
	} else {
		__antithesis_instrumentation__.Notify(7953)
	}
	__antithesis_instrumentation__.Notify(7945)

	return newOpts, nil
}

func GetRedactedBackupNode(
	backup *tree.Backup,
	to []string,
	incrementalFrom []string,
	kmsURIs []string,
	resolvedSubdir string,
	incrementalStorage []string,
	hasBeenPlanned bool,
) (*tree.Backup, error) {
	__antithesis_instrumentation__.Notify(7954)
	b := &tree.Backup{
		AsOf:    backup.AsOf,
		Targets: backup.Targets,
		Nested:  backup.Nested,
	}

	if b.Nested && func() bool {
		__antithesis_instrumentation__.Notify(7959)
		return hasBeenPlanned == true
	}() == true {
		__antithesis_instrumentation__.Notify(7960)
		b.Subdir = tree.NewDString(resolvedSubdir)
	} else {
		__antithesis_instrumentation__.Notify(7961)
	}
	__antithesis_instrumentation__.Notify(7955)

	var err error
	b.To, err = sanitizeURIList(to)
	if err != nil {
		__antithesis_instrumentation__.Notify(7962)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(7963)
	}
	__antithesis_instrumentation__.Notify(7956)

	b.IncrementalFrom, err = sanitizeURIList(incrementalFrom)
	if err != nil {
		__antithesis_instrumentation__.Notify(7964)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(7965)
	}
	__antithesis_instrumentation__.Notify(7957)

	resolvedOpts, err := resolveOptionsForBackupJobDescription(backup.Options, kmsURIs,
		incrementalStorage)
	if err != nil {
		__antithesis_instrumentation__.Notify(7966)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(7967)
	}
	__antithesis_instrumentation__.Notify(7958)
	b.Options = resolvedOpts
	return b, nil
}

func sanitizeURIList(uris []string) ([]tree.Expr, error) {
	__antithesis_instrumentation__.Notify(7968)
	var sanitizedURIs []tree.Expr
	for _, uri := range uris {
		__antithesis_instrumentation__.Notify(7970)
		sanitizedURI, err := cloud.SanitizeExternalStorageURI(uri, nil)
		if err != nil {
			__antithesis_instrumentation__.Notify(7972)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(7973)
		}
		__antithesis_instrumentation__.Notify(7971)
		sanitizedURIs = append(sanitizedURIs, tree.NewDString(sanitizedURI))
	}
	__antithesis_instrumentation__.Notify(7969)
	return sanitizedURIs, nil
}

func backupJobDescription(
	p sql.PlanHookState,
	backup *tree.Backup,
	to []string,
	incrementalFrom []string,
	kmsURIs []string,
	resolvedSubdir string,
	incrementalStorage []string,
) (string, error) {
	__antithesis_instrumentation__.Notify(7974)
	b, err := GetRedactedBackupNode(backup, to, incrementalFrom, kmsURIs,
		resolvedSubdir, incrementalStorage, true)
	if err != nil {
		__antithesis_instrumentation__.Notify(7976)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(7977)
	}
	__antithesis_instrumentation__.Notify(7975)

	ann := p.ExtendedEvalContext().Annotations
	return tree.AsStringWithFQNames(b, ann), nil
}

func validateKMSURIsAgainstFullBackup(
	kmsURIs []string, kmsMasterKeyIDToDataKey *encryptedDataKeyMap, kmsEnv cloud.KMSEnv,
) (*jobspb.BackupEncryptionOptions_KMSInfo, error) {
	__antithesis_instrumentation__.Notify(7978)
	var defaultKMSInfo *jobspb.BackupEncryptionOptions_KMSInfo
	for _, kmsURI := range kmsURIs {
		__antithesis_instrumentation__.Notify(7980)
		kms, err := cloud.KMSFromURI(kmsURI, kmsEnv)
		if err != nil {
			__antithesis_instrumentation__.Notify(7985)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(7986)
		}
		__antithesis_instrumentation__.Notify(7981)

		defer func() {
			__antithesis_instrumentation__.Notify(7987)
			_ = kms.Close()
		}()
		__antithesis_instrumentation__.Notify(7982)

		id, err := kms.MasterKeyID()
		if err != nil {
			__antithesis_instrumentation__.Notify(7988)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(7989)
		}
		__antithesis_instrumentation__.Notify(7983)

		encryptedDataKey, err := kmsMasterKeyIDToDataKey.getEncryptedDataKey(plaintextMasterKeyID(id))
		if err != nil {
			__antithesis_instrumentation__.Notify(7990)
			return nil,
				errors.Wrap(err,
					"one of the provided URIs was not used when encrypting the base BACKUP")
		} else {
			__antithesis_instrumentation__.Notify(7991)
		}
		__antithesis_instrumentation__.Notify(7984)

		if defaultKMSInfo == nil {
			__antithesis_instrumentation__.Notify(7992)
			defaultKMSInfo = &jobspb.BackupEncryptionOptions_KMSInfo{
				Uri:              kmsURI,
				EncryptedDataKey: encryptedDataKey,
			}
		} else {
			__antithesis_instrumentation__.Notify(7993)
		}
	}
	__antithesis_instrumentation__.Notify(7979)

	return defaultKMSInfo, nil
}

type annotatedBackupStatement struct {
	*tree.Backup
	*jobs.CreatedByInfo
}

func getBackupStatement(stmt tree.Statement) *annotatedBackupStatement {
	__antithesis_instrumentation__.Notify(7994)
	switch backup := stmt.(type) {
	case *annotatedBackupStatement:
		__antithesis_instrumentation__.Notify(7995)
		return backup
	case *tree.Backup:
		__antithesis_instrumentation__.Notify(7996)
		return &annotatedBackupStatement{Backup: backup}
	default:
		__antithesis_instrumentation__.Notify(7997)
		return nil
	}
}

func checkPrivilegesForBackup(
	ctx context.Context,
	backupStmt *annotatedBackupStatement,
	p sql.PlanHookState,
	targetDescs []catalog.Descriptor,
	to []string,
) error {
	__antithesis_instrumentation__.Notify(7998)
	hasAdmin, err := p.HasAdminRole(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(8006)
		return err
	} else {
		__antithesis_instrumentation__.Notify(8007)
	}
	__antithesis_instrumentation__.Notify(7999)
	if hasAdmin {
		__antithesis_instrumentation__.Notify(8008)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(8009)
	}
	__antithesis_instrumentation__.Notify(8000)

	if backupStmt.Coverage() == tree.AllDescriptors {
		__antithesis_instrumentation__.Notify(8010)
		return pgerror.Newf(
			pgcode.InsufficientPrivilege,
			"only users with the admin role are allowed to perform full cluster backups")
	} else {
		__antithesis_instrumentation__.Notify(8011)
	}
	__antithesis_instrumentation__.Notify(8001)

	if backupStmt.Targets != nil && func() bool {
		__antithesis_instrumentation__.Notify(8012)
		return backupStmt.Targets.TenantID.IsSet() == true
	}() == true {
		__antithesis_instrumentation__.Notify(8013)
		return pgerror.Newf(
			pgcode.InsufficientPrivilege,
			"only users with the admin role can perform BACKUP TENANT")
	} else {
		__antithesis_instrumentation__.Notify(8014)
	}
	__antithesis_instrumentation__.Notify(8002)
	for _, desc := range targetDescs {
		__antithesis_instrumentation__.Notify(8015)
		switch desc := desc.(type) {
		case catalog.DatabaseDescriptor:
			__antithesis_instrumentation__.Notify(8016)
			if connectErr := p.CheckPrivilege(ctx, desc, privilege.CONNECT); connectErr != nil {
				__antithesis_instrumentation__.Notify(8019)

				if selectErr := p.CheckPrivilege(ctx, desc, privilege.SELECT); selectErr != nil {
					__antithesis_instrumentation__.Notify(8020)

					return connectErr
				} else {
					__antithesis_instrumentation__.Notify(8021)
				}
			} else {
				__antithesis_instrumentation__.Notify(8022)
			}
		case catalog.TableDescriptor:
			__antithesis_instrumentation__.Notify(8017)
			if err := p.CheckPrivilege(ctx, desc, privilege.SELECT); err != nil {
				__antithesis_instrumentation__.Notify(8023)
				return err
			} else {
				__antithesis_instrumentation__.Notify(8024)
			}
		case catalog.TypeDescriptor, catalog.SchemaDescriptor:
			__antithesis_instrumentation__.Notify(8018)
			if err := p.CheckPrivilege(ctx, desc, privilege.USAGE); err != nil {
				__antithesis_instrumentation__.Notify(8025)
				return err
			} else {
				__antithesis_instrumentation__.Notify(8026)
			}
		}
	}
	__antithesis_instrumentation__.Notify(8003)
	if p.ExecCfg().ExternalIODirConfig.EnableNonAdminImplicitAndArbitraryOutbound {
		__antithesis_instrumentation__.Notify(8027)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(8028)
	}
	__antithesis_instrumentation__.Notify(8004)

	for _, uri := range to {
		__antithesis_instrumentation__.Notify(8029)
		conf, err := cloud.ExternalStorageConfFromURI(uri, p.User())
		if err != nil {
			__antithesis_instrumentation__.Notify(8031)
			return err
		} else {
			__antithesis_instrumentation__.Notify(8032)
		}
		__antithesis_instrumentation__.Notify(8030)
		if !conf.AccessIsWithExplicitAuth() {
			__antithesis_instrumentation__.Notify(8033)
			return pgerror.Newf(
				pgcode.InsufficientPrivilege,
				"only users with the admin role are allowed to BACKUP to the specified %s URI",
				conf.Provider.String())
		} else {
			__antithesis_instrumentation__.Notify(8034)
		}
	}
	__antithesis_instrumentation__.Notify(8005)

	return nil
}

func requireEnterprise(execCfg *sql.ExecutorConfig, feature string) error {
	__antithesis_instrumentation__.Notify(8035)
	if err := utilccl.CheckEnterpriseEnabled(
		execCfg.Settings, execCfg.LogicalClusterID(), execCfg.Organization(),
		fmt.Sprintf("BACKUP with %s", feature),
	); err != nil {
		__antithesis_instrumentation__.Notify(8037)
		return err
	} else {
		__antithesis_instrumentation__.Notify(8038)
	}
	__antithesis_instrumentation__.Notify(8036)
	return nil
}

func backupPlanHook(
	ctx context.Context, stmt tree.Statement, p sql.PlanHookState,
) (sql.PlanHookRowFn, colinfo.ResultColumns, []sql.PlanNode, bool, error) {
	__antithesis_instrumentation__.Notify(8039)
	backupStmt := getBackupStatement(stmt)
	if backupStmt == nil {
		__antithesis_instrumentation__.Notify(8052)
		return nil, nil, nil, false, nil
	} else {
		__antithesis_instrumentation__.Notify(8053)
	}
	__antithesis_instrumentation__.Notify(8040)

	if err := featureflag.CheckEnabled(
		ctx,
		p.ExecCfg(),
		featureBackupEnabled,
		"BACKUP",
	); err != nil {
		__antithesis_instrumentation__.Notify(8054)
		return nil, nil, nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(8055)
	}
	__antithesis_instrumentation__.Notify(8041)

	if !backupStmt.Nested {
		__antithesis_instrumentation__.Notify(8056)
		p.BufferClientNotice(ctx,
			pgnotice.Newf("The `BACKUP TO` syntax will be removed in a future release, please"+
				" switch over to using `BACKUP INTO` to create a backup collection: %s. "+
				"Backups created using the `BACKUP TO` syntax may not be restoreable in the next major version release.",
				"https://www.cockroachlabs.com/docs/stable/backup.html#considerations"))
	} else {
		__antithesis_instrumentation__.Notify(8057)
	}
	__antithesis_instrumentation__.Notify(8042)

	var err error
	subdirFn := func() (string, error) { __antithesis_instrumentation__.Notify(8058); return "", nil }
	__antithesis_instrumentation__.Notify(8043)
	if backupStmt.Subdir != nil {
		__antithesis_instrumentation__.Notify(8059)
		subdirFn, err = p.TypeAsString(ctx, backupStmt.Subdir, "BACKUP")
		if err != nil {
			__antithesis_instrumentation__.Notify(8060)
			return nil, nil, nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(8061)
		}
	} else {
		__antithesis_instrumentation__.Notify(8062)
	}
	__antithesis_instrumentation__.Notify(8044)

	toFn, err := p.TypeAsStringArray(ctx, tree.Exprs(backupStmt.To), "BACKUP")
	if err != nil {
		__antithesis_instrumentation__.Notify(8063)
		return nil, nil, nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(8064)
	}
	__antithesis_instrumentation__.Notify(8045)
	incrementalFromFn, err := p.TypeAsStringArray(ctx, backupStmt.IncrementalFrom, "BACKUP")
	if err != nil {
		__antithesis_instrumentation__.Notify(8065)
		return nil, nil, nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(8066)
	}
	__antithesis_instrumentation__.Notify(8046)

	incToFn, err := p.TypeAsStringArray(ctx, tree.Exprs(backupStmt.Options.IncrementalStorage),
		"BACKUP")
	if err != nil {
		__antithesis_instrumentation__.Notify(8067)
		return nil, nil, nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(8068)
	}
	__antithesis_instrumentation__.Notify(8047)

	encryptionParams := jobspb.BackupEncryptionOptions{Mode: jobspb.EncryptionMode_None}

	var pwFn func() (string, error)
	if backupStmt.Options.EncryptionPassphrase != nil {
		__antithesis_instrumentation__.Notify(8069)
		fn, err := p.TypeAsString(ctx, backupStmt.Options.EncryptionPassphrase, "BACKUP")
		if err != nil {
			__antithesis_instrumentation__.Notify(8071)
			return nil, nil, nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(8072)
		}
		__antithesis_instrumentation__.Notify(8070)
		pwFn = fn
		encryptionParams.Mode = jobspb.EncryptionMode_Passphrase
	} else {
		__antithesis_instrumentation__.Notify(8073)
	}
	__antithesis_instrumentation__.Notify(8048)

	var kmsFn func() ([]string, error)
	if backupStmt.Options.EncryptionKMSURI != nil {
		__antithesis_instrumentation__.Notify(8074)
		if encryptionParams.Mode != jobspb.EncryptionMode_None {
			__antithesis_instrumentation__.Notify(8078)
			return nil, nil, nil, false,
				errors.New("cannot have both encryption_passphrase and kms option set")
		} else {
			__antithesis_instrumentation__.Notify(8079)
		}
		__antithesis_instrumentation__.Notify(8075)
		fn, err := p.TypeAsStringArray(ctx, tree.Exprs(backupStmt.Options.EncryptionKMSURI),
			"BACKUP")
		if err != nil {
			__antithesis_instrumentation__.Notify(8080)
			return nil, nil, nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(8081)
		}
		__antithesis_instrumentation__.Notify(8076)
		kmsFn = func() ([]string, error) {
			__antithesis_instrumentation__.Notify(8082)
			res, err := fn()
			if err == nil {
				__antithesis_instrumentation__.Notify(8084)
				return res, nil
			} else {
				__antithesis_instrumentation__.Notify(8085)
			}
			__antithesis_instrumentation__.Notify(8083)
			return nil, err
		}
		__antithesis_instrumentation__.Notify(8077)
		encryptionParams.Mode = jobspb.EncryptionMode_KMS
	} else {
		__antithesis_instrumentation__.Notify(8086)
	}
	__antithesis_instrumentation__.Notify(8049)

	fn := func(ctx context.Context, _ []sql.PlanNode, resultsCh chan<- tree.Datums) error {
		__antithesis_instrumentation__.Notify(8087)

		ctx, span := tracing.ChildSpan(ctx, stmt.StatementTag())
		defer span.Finish()

		if !(p.ExtendedEvalContext().TxnIsSingleStmt || func() bool {
			__antithesis_instrumentation__.Notify(8116)
			return backupStmt.Options.Detached == true
		}() == true) {
			__antithesis_instrumentation__.Notify(8117)
			return errors.Errorf("BACKUP cannot be used inside a multi-statement transaction without DETACHED option")
		} else {
			__antithesis_instrumentation__.Notify(8118)
		}
		__antithesis_instrumentation__.Notify(8088)

		subdir, err := subdirFn()
		if err != nil {
			__antithesis_instrumentation__.Notify(8119)
			return err
		} else {
			__antithesis_instrumentation__.Notify(8120)
		}
		__antithesis_instrumentation__.Notify(8089)

		to, err := toFn()
		if err != nil {
			__antithesis_instrumentation__.Notify(8121)
			return err
		} else {
			__antithesis_instrumentation__.Notify(8122)
		}
		__antithesis_instrumentation__.Notify(8090)
		if len(to) > 1 {
			__antithesis_instrumentation__.Notify(8123)
			if err := requireEnterprise(p.ExecCfg(), "partitioned destinations"); err != nil {
				__antithesis_instrumentation__.Notify(8124)
				return err
			} else {
				__antithesis_instrumentation__.Notify(8125)
			}
		} else {
			__antithesis_instrumentation__.Notify(8126)
		}
		__antithesis_instrumentation__.Notify(8091)

		incrementalFrom, err := incrementalFromFn()
		if err != nil {
			__antithesis_instrumentation__.Notify(8127)
			return err
		} else {
			__antithesis_instrumentation__.Notify(8128)
		}
		__antithesis_instrumentation__.Notify(8092)

		incrementalStorage, err := incToFn()
		if err != nil {
			__antithesis_instrumentation__.Notify(8129)
			return err
		} else {
			__antithesis_instrumentation__.Notify(8130)
		}
		__antithesis_instrumentation__.Notify(8093)
		if !backupStmt.Nested && func() bool {
			__antithesis_instrumentation__.Notify(8131)
			return len(incrementalStorage) > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(8132)
			return errors.New("incremental_location option not supported with `BACKUP TO` syntax")
		} else {
			__antithesis_instrumentation__.Notify(8133)
		}
		__antithesis_instrumentation__.Notify(8094)
		if len(incrementalStorage) > 0 && func() bool {
			__antithesis_instrumentation__.Notify(8134)
			return (len(incrementalStorage) != len(to)) == true
		}() == true {
			__antithesis_instrumentation__.Notify(8135)
			return errors.New("the incremental_location option must contain the same number of locality" +
				" aware URIs as the full backup destination")
		} else {
			__antithesis_instrumentation__.Notify(8136)
		}
		__antithesis_instrumentation__.Notify(8095)

		endTime := p.ExecCfg().Clock.Now()
		if backupStmt.AsOf.Expr != nil {
			__antithesis_instrumentation__.Notify(8137)
			asOf, err := p.EvalAsOfTimestamp(ctx, backupStmt.AsOf)
			if err != nil {
				__antithesis_instrumentation__.Notify(8139)
				return err
			} else {
				__antithesis_instrumentation__.Notify(8140)
			}
			__antithesis_instrumentation__.Notify(8138)
			endTime = asOf.Timestamp
		} else {
			__antithesis_instrumentation__.Notify(8141)
		}
		__antithesis_instrumentation__.Notify(8096)

		switch encryptionParams.Mode {
		case jobspb.EncryptionMode_Passphrase:
			__antithesis_instrumentation__.Notify(8142)
			pw, err := pwFn()
			if err != nil {
				__antithesis_instrumentation__.Notify(8148)
				return err
			} else {
				__antithesis_instrumentation__.Notify(8149)
			}
			__antithesis_instrumentation__.Notify(8143)
			if err := requireEnterprise(p.ExecCfg(), "encryption"); err != nil {
				__antithesis_instrumentation__.Notify(8150)
				return err
			} else {
				__antithesis_instrumentation__.Notify(8151)
			}
			__antithesis_instrumentation__.Notify(8144)
			encryptionParams.RawPassphrae = pw
		case jobspb.EncryptionMode_KMS:
			__antithesis_instrumentation__.Notify(8145)
			encryptionParams.RawKmsUris, err = kmsFn()
			if err != nil {
				__antithesis_instrumentation__.Notify(8152)
				return err
			} else {
				__antithesis_instrumentation__.Notify(8153)
			}
			__antithesis_instrumentation__.Notify(8146)
			if err := requireEnterprise(p.ExecCfg(), "encryption"); err != nil {
				__antithesis_instrumentation__.Notify(8154)
				return err
			} else {
				__antithesis_instrumentation__.Notify(8155)
			}
		default:
			__antithesis_instrumentation__.Notify(8147)
		}
		__antithesis_instrumentation__.Notify(8097)

		var revisionHistory bool
		if backupStmt.Options.CaptureRevisionHistory {
			__antithesis_instrumentation__.Notify(8156)
			if err := requireEnterprise(p.ExecCfg(), "revision_history"); err != nil {
				__antithesis_instrumentation__.Notify(8158)
				return err
			} else {
				__antithesis_instrumentation__.Notify(8159)
			}
			__antithesis_instrumentation__.Notify(8157)
			revisionHistory = true
		} else {
			__antithesis_instrumentation__.Notify(8160)
		}
		__antithesis_instrumentation__.Notify(8098)

		var targetDescs []catalog.Descriptor
		var completeDBs []descpb.ID

		switch backupStmt.Coverage() {
		case tree.RequestedDescriptors:
			__antithesis_instrumentation__.Notify(8161)
			var err error
			targetDescs, completeDBs, err = backupresolver.ResolveTargetsToDescriptors(ctx, p, endTime, backupStmt.Targets)
			if err != nil {
				__antithesis_instrumentation__.Notify(8164)
				return errors.Wrap(err, "failed to resolve targets specified in the BACKUP stmt")
			} else {
				__antithesis_instrumentation__.Notify(8165)
			}
		case tree.AllDescriptors:
			__antithesis_instrumentation__.Notify(8162)
			var err error
			targetDescs, completeDBs, err = fullClusterTargetsBackup(ctx, p.ExecCfg(), endTime)
			if err != nil {
				__antithesis_instrumentation__.Notify(8166)
				return err
			} else {
				__antithesis_instrumentation__.Notify(8167)
			}
		default:
			__antithesis_instrumentation__.Notify(8163)
			return errors.AssertionFailedf("unexpected descriptor coverage %v", backupStmt.Coverage())
		}
		__antithesis_instrumentation__.Notify(8099)

		err = checkPrivilegesForBackup(ctx, backupStmt, p, targetDescs, to)
		if err != nil {
			__antithesis_instrumentation__.Notify(8168)
			return err
		} else {
			__antithesis_instrumentation__.Notify(8169)
		}
		__antithesis_instrumentation__.Notify(8100)

		initialDetails := jobspb.BackupDetails{
			Destination:         jobspb.BackupDetails_Destination{To: to, IncrementalStorage: incrementalStorage},
			EndTime:             endTime,
			RevisionHistory:     revisionHistory,
			IncrementalFrom:     incrementalFrom,
			FullCluster:         backupStmt.Coverage() == tree.AllDescriptors,
			ResolvedCompleteDbs: completeDBs,
			EncryptionOptions:   &encryptionParams,
		}
		if backupStmt.CreatedByInfo != nil && func() bool {
			__antithesis_instrumentation__.Notify(8170)
			return backupStmt.CreatedByInfo.Name == jobs.CreatedByScheduledJobs == true
		}() == true {
			__antithesis_instrumentation__.Notify(8171)
			initialDetails.ScheduleID = backupStmt.CreatedByInfo.ID
		} else {
			__antithesis_instrumentation__.Notify(8172)
		}
		__antithesis_instrumentation__.Notify(8101)

		if !initialDetails.FullCluster {
			__antithesis_instrumentation__.Notify(8173)
			descriptorProtos := make([]descpb.Descriptor, 0, len(targetDescs))
			for _, desc := range targetDescs {
				__antithesis_instrumentation__.Notify(8175)
				descriptorProtos = append(descriptorProtos, *desc.DescriptorProto())
			}
			__antithesis_instrumentation__.Notify(8174)
			initialDetails.ResolvedTargets = descriptorProtos
		} else {
			__antithesis_instrumentation__.Notify(8176)
		}
		__antithesis_instrumentation__.Notify(8102)

		if backupStmt.Nested {
			__antithesis_instrumentation__.Notify(8177)
			if backupStmt.AppendToLatest {
				__antithesis_instrumentation__.Notify(8178)
				initialDetails.Destination.Subdir = latestFileName
				initialDetails.Destination.Exists = true

			} else {
				__antithesis_instrumentation__.Notify(8179)
				if subdir != "" {
					__antithesis_instrumentation__.Notify(8180)
					initialDetails.Destination.Subdir = "/" + strings.TrimPrefix(subdir, "/")
					initialDetails.Destination.Exists = true
				} else {
					__antithesis_instrumentation__.Notify(8181)
					initialDetails.Destination.Subdir = endTime.GoTime().Format(DateBasedIntoFolderName)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(8182)
		}
		__antithesis_instrumentation__.Notify(8103)

		if backupStmt.Targets != nil && func() bool {
			__antithesis_instrumentation__.Notify(8183)
			return backupStmt.Targets.TenantID.IsSet() == true
		}() == true {
			__antithesis_instrumentation__.Notify(8184)
			if !p.ExecCfg().Codec.ForSystemTenant() {
				__antithesis_instrumentation__.Notify(8186)
				return pgerror.Newf(pgcode.InsufficientPrivilege, "only the system tenant can backup other tenants")
			} else {
				__antithesis_instrumentation__.Notify(8187)
			}
			__antithesis_instrumentation__.Notify(8185)
			initialDetails.SpecificTenantIds = []roachpb.TenantID{backupStmt.Targets.TenantID.TenantID}
		} else {
			__antithesis_instrumentation__.Notify(8188)
		}
		__antithesis_instrumentation__.Notify(8104)

		jobID := p.ExecCfg().JobRegistry.MakeJobID()

		if p.ExecCfg().Settings.Version.IsActive(ctx, clusterversion.BackupResolutionInJob) {
			__antithesis_instrumentation__.Notify(8189)
			description, err := backupJobDescription(p,
				backupStmt.Backup, to, incrementalFrom,
				encryptionParams.RawKmsUris,
				initialDetails.Destination.Subdir,
				initialDetails.Destination.IncrementalStorage,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(8196)
				return err
			} else {
				__antithesis_instrumentation__.Notify(8197)
			}
			__antithesis_instrumentation__.Notify(8190)
			jr := jobs.Record{
				Description: description,
				Details:     initialDetails,
				Progress:    jobspb.BackupProgress{},
				CreatedBy:   backupStmt.CreatedByInfo,
				Username:    p.User(),
				DescriptorIDs: func() (sqlDescIDs []descpb.ID) {
					__antithesis_instrumentation__.Notify(8198)
					for i := range targetDescs {
						__antithesis_instrumentation__.Notify(8200)
						sqlDescIDs = append(sqlDescIDs, targetDescs[i].GetID())
					}
					__antithesis_instrumentation__.Notify(8199)
					return sqlDescIDs
				}(),
			}
			__antithesis_instrumentation__.Notify(8191)
			plannerTxn := p.ExtendedEvalContext().Txn

			if backupStmt.Options.Detached {
				__antithesis_instrumentation__.Notify(8201)

				_, err := p.ExecCfg().JobRegistry.CreateAdoptableJobWithTxn(
					ctx, jr, jobID, plannerTxn)
				if err != nil {
					__antithesis_instrumentation__.Notify(8203)
					return err
				} else {
					__antithesis_instrumentation__.Notify(8204)
				}
				__antithesis_instrumentation__.Notify(8202)
				resultsCh <- tree.Datums{tree.NewDInt(tree.DInt(jobID))}
				return nil
			} else {
				__antithesis_instrumentation__.Notify(8205)
			}
			__antithesis_instrumentation__.Notify(8192)
			var sj *jobs.StartableJob
			if err := func() (err error) {
				__antithesis_instrumentation__.Notify(8206)
				defer func() {
					__antithesis_instrumentation__.Notify(8209)
					if err == nil || func() bool {
						__antithesis_instrumentation__.Notify(8211)
						return sj == nil == true
					}() == true {
						__antithesis_instrumentation__.Notify(8212)
						return
					} else {
						__antithesis_instrumentation__.Notify(8213)
					}
					__antithesis_instrumentation__.Notify(8210)
					if cleanupErr := sj.CleanupOnRollback(ctx); cleanupErr != nil {
						__antithesis_instrumentation__.Notify(8214)
						log.Errorf(ctx, "failed to cleanup job: %v", cleanupErr)
					} else {
						__antithesis_instrumentation__.Notify(8215)
					}
				}()
				__antithesis_instrumentation__.Notify(8207)
				if err := p.ExecCfg().JobRegistry.CreateStartableJobWithTxn(ctx, &sj, jobID, plannerTxn, jr); err != nil {
					__antithesis_instrumentation__.Notify(8216)
					return err
				} else {
					__antithesis_instrumentation__.Notify(8217)
				}
				__antithesis_instrumentation__.Notify(8208)

				return plannerTxn.Commit(ctx)
			}(); err != nil {
				__antithesis_instrumentation__.Notify(8218)
				return err
			} else {
				__antithesis_instrumentation__.Notify(8219)
			}
			__antithesis_instrumentation__.Notify(8193)
			if err := sj.Start(ctx); err != nil {
				__antithesis_instrumentation__.Notify(8220)
				return err
			} else {
				__antithesis_instrumentation__.Notify(8221)
			}
			__antithesis_instrumentation__.Notify(8194)
			if err := sj.AwaitCompletion(ctx); err != nil {
				__antithesis_instrumentation__.Notify(8222)
				return err
			} else {
				__antithesis_instrumentation__.Notify(8223)
			}
			__antithesis_instrumentation__.Notify(8195)
			return sj.ReportExecutionResults(ctx, resultsCh)
		} else {
			__antithesis_instrumentation__.Notify(8224)
		}
		__antithesis_instrumentation__.Notify(8105)

		backupDetails, backupManifest, err := getBackupDetailAndManifest(
			ctx, p.ExecCfg(), p.ExtendedEvalContext().Txn, initialDetails, p.User(),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(8225)
			return err
		} else {
			__antithesis_instrumentation__.Notify(8226)
		}
		__antithesis_instrumentation__.Notify(8106)

		description, err := backupJobDescription(p, backupStmt.Backup, to, incrementalFrom, encryptionParams.RawKmsUris, backupDetails.Destination.Subdir, initialDetails.Destination.IncrementalStorage)
		if err != nil {
			__antithesis_instrumentation__.Notify(8227)
			return err
		} else {
			__antithesis_instrumentation__.Notify(8228)
		}
		__antithesis_instrumentation__.Notify(8107)

		plannerTxn := p.ExtendedEvalContext().Txn

		if err := planSchedulePTSChaining(ctx, p.ExecCfg(), plannerTxn, &backupDetails, backupStmt.CreatedByInfo); err != nil {
			__antithesis_instrumentation__.Notify(8229)
			return err
		} else {
			__antithesis_instrumentation__.Notify(8230)
		}
		__antithesis_instrumentation__.Notify(8108)

		if p.ExecCfg().Settings.Version.IsActive(ctx, clusterversion.EnableProtectedTimestampsForTenant) {
			__antithesis_instrumentation__.Notify(8231)
			protectedtsID := uuid.MakeV4()
			backupDetails.ProtectedTimestampRecord = &protectedtsID
		} else {
			__antithesis_instrumentation__.Notify(8232)
			if len(backupManifest.Spans) > 0 && func() bool {
				__antithesis_instrumentation__.Notify(8233)
				return p.ExecCfg().Codec.ForSystemTenant() == true
			}() == true {
				__antithesis_instrumentation__.Notify(8234)
				protectedtsID := uuid.MakeV4()
				backupDetails.ProtectedTimestampRecord = &protectedtsID
			} else {
				__antithesis_instrumentation__.Notify(8235)
			}
		}
		__antithesis_instrumentation__.Notify(8109)

		jr := jobs.Record{
			Description: description,
			Username:    p.User(),

			DescriptorIDs: func() (sqlDescIDs []descpb.ID) {
				__antithesis_instrumentation__.Notify(8236)
				for i := range backupManifest.Descriptors {
					__antithesis_instrumentation__.Notify(8238)
					sqlDescIDs = append(sqlDescIDs,
						descpb.GetDescriptorID(&backupManifest.Descriptors[i]))
				}
				__antithesis_instrumentation__.Notify(8237)
				return sqlDescIDs
			}(),
			Details:   backupDetails,
			Progress:  jobspb.BackupProgress{},
			CreatedBy: backupStmt.CreatedByInfo,
		}
		__antithesis_instrumentation__.Notify(8110)

		lic := utilccl.CheckEnterpriseEnabled(
			p.ExecCfg().Settings, p.ExecCfg().LogicalClusterID(), p.ExecCfg().Organization(), "",
		) != nil

		if backupDetails.ProtectedTimestampRecord != nil {
			__antithesis_instrumentation__.Notify(8239)
			if err := protectTimestampForBackup(
				ctx, p.ExecCfg(), plannerTxn, jobID, backupManifest, backupDetails,
			); err != nil {
				__antithesis_instrumentation__.Notify(8240)
				return err
			} else {
				__antithesis_instrumentation__.Notify(8241)
			}
		} else {
			__antithesis_instrumentation__.Notify(8242)
		}
		__antithesis_instrumentation__.Notify(8111)

		if backupStmt.Options.Detached {
			__antithesis_instrumentation__.Notify(8243)

			_, err := p.ExecCfg().JobRegistry.CreateAdoptableJobWithTxn(
				ctx, jr, jobID, plannerTxn)
			if err != nil {
				__antithesis_instrumentation__.Notify(8246)
				return err
			} else {
				__antithesis_instrumentation__.Notify(8247)
			}
			__antithesis_instrumentation__.Notify(8244)

			if err := writeBackupManifestCheckpoint(
				ctx, backupDetails.URI, backupDetails.EncryptionOptions, &backupManifest, p.ExecCfg(), p.User(),
			); err != nil {
				__antithesis_instrumentation__.Notify(8248)
				return err
			} else {
				__antithesis_instrumentation__.Notify(8249)
			}
			__antithesis_instrumentation__.Notify(8245)

			resultsCh <- tree.Datums{tree.NewDInt(tree.DInt(jobID))}
			collectTelemetry(backupManifest, initialDetails, backupDetails, lic)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(8250)
		}
		__antithesis_instrumentation__.Notify(8112)

		var sj *jobs.StartableJob
		if err := func() (err error) {
			__antithesis_instrumentation__.Notify(8251)
			defer func() {
				__antithesis_instrumentation__.Notify(8255)
				if err == nil || func() bool {
					__antithesis_instrumentation__.Notify(8257)
					return sj == nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(8258)
					return
				} else {
					__antithesis_instrumentation__.Notify(8259)
				}
				__antithesis_instrumentation__.Notify(8256)
				if cleanupErr := sj.CleanupOnRollback(ctx); cleanupErr != nil {
					__antithesis_instrumentation__.Notify(8260)
					log.Errorf(ctx, "failed to cleanup job: %v", cleanupErr)
				} else {
					__antithesis_instrumentation__.Notify(8261)
				}
			}()
			__antithesis_instrumentation__.Notify(8252)
			if err := p.ExecCfg().JobRegistry.CreateStartableJobWithTxn(ctx, &sj, jobID, plannerTxn, jr); err != nil {
				__antithesis_instrumentation__.Notify(8262)
				return err
			} else {
				__antithesis_instrumentation__.Notify(8263)
			}
			__antithesis_instrumentation__.Notify(8253)

			if err := writeBackupManifestCheckpoint(
				ctx, backupDetails.URI, backupDetails.EncryptionOptions, &backupManifest, p.ExecCfg(), p.User(),
			); err != nil {
				__antithesis_instrumentation__.Notify(8264)
				return err
			} else {
				__antithesis_instrumentation__.Notify(8265)
			}
			__antithesis_instrumentation__.Notify(8254)

			return plannerTxn.Commit(ctx)
		}(); err != nil {
			__antithesis_instrumentation__.Notify(8266)
			return err
		} else {
			__antithesis_instrumentation__.Notify(8267)
		}
		__antithesis_instrumentation__.Notify(8113)

		collectTelemetry(backupManifest, initialDetails, backupDetails, lic)
		if err := sj.Start(ctx); err != nil {
			__antithesis_instrumentation__.Notify(8268)
			return err
		} else {
			__antithesis_instrumentation__.Notify(8269)
		}
		__antithesis_instrumentation__.Notify(8114)
		if err := sj.AwaitCompletion(ctx); err != nil {
			__antithesis_instrumentation__.Notify(8270)
			return err
		} else {
			__antithesis_instrumentation__.Notify(8271)
		}
		__antithesis_instrumentation__.Notify(8115)
		return sj.ReportExecutionResults(ctx, resultsCh)
	}
	__antithesis_instrumentation__.Notify(8050)

	if backupStmt.Options.Detached {
		__antithesis_instrumentation__.Notify(8272)
		return fn, jobs.DetachedJobExecutionResultHeader, nil, false, nil
	} else {
		__antithesis_instrumentation__.Notify(8273)
	}
	__antithesis_instrumentation__.Notify(8051)
	return fn, jobs.BulkJobExecutionResultHeader, nil, false, nil
}

func collectTelemetry(
	backupManifest BackupManifest, initialDetails, backupDetails jobspb.BackupDetails, licensed bool,
) {
	__antithesis_instrumentation__.Notify(8274)

	sourceSuffix := ".manual"
	if backupDetails.ScheduleID != 0 {
		__antithesis_instrumentation__.Notify(8284)
		sourceSuffix = ".scheduled"
	} else {
		__antithesis_instrumentation__.Notify(8285)
	}
	__antithesis_instrumentation__.Notify(8275)

	countSource := func(feature string) {
		__antithesis_instrumentation__.Notify(8286)
		telemetry.Count(feature + sourceSuffix)
	}
	__antithesis_instrumentation__.Notify(8276)

	countSource("backup.total.started")
	if backupManifest.isIncremental() || func() bool {
		__antithesis_instrumentation__.Notify(8287)
		return backupDetails.EncryptionOptions != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(8288)
		countSource("backup.using-enterprise-features")
	} else {
		__antithesis_instrumentation__.Notify(8289)
	}
	__antithesis_instrumentation__.Notify(8277)
	if licensed {
		__antithesis_instrumentation__.Notify(8290)
		countSource("backup.licensed")
	} else {
		__antithesis_instrumentation__.Notify(8291)
		countSource("backup.free")
	}
	__antithesis_instrumentation__.Notify(8278)
	if backupDetails.StartTime.IsEmpty() {
		__antithesis_instrumentation__.Notify(8292)
		countSource("backup.span.full")
	} else {
		__antithesis_instrumentation__.Notify(8293)
		countSource("backup.span.incremental")
		telemetry.CountBucketed("backup.incremental-span-sec",
			int64(backupDetails.EndTime.GoTime().Sub(backupDetails.StartTime.GoTime()).Seconds()))
		if len(initialDetails.IncrementalFrom) == 0 {
			__antithesis_instrumentation__.Notify(8294)
			countSource("backup.auto-incremental")
		} else {
			__antithesis_instrumentation__.Notify(8295)
		}
	}
	__antithesis_instrumentation__.Notify(8279)
	if len(backupDetails.URIsByLocalityKV) > 1 {
		__antithesis_instrumentation__.Notify(8296)
		countSource("backup.partitioned")
	} else {
		__antithesis_instrumentation__.Notify(8297)
	}
	__antithesis_instrumentation__.Notify(8280)
	if backupManifest.MVCCFilter == MVCCFilter_All {
		__antithesis_instrumentation__.Notify(8298)
		countSource("backup.revision-history")
	} else {
		__antithesis_instrumentation__.Notify(8299)
	}
	__antithesis_instrumentation__.Notify(8281)
	if backupDetails.EncryptionOptions != nil {
		__antithesis_instrumentation__.Notify(8300)
		countSource("backup.encrypted")
		switch backupDetails.EncryptionOptions.Mode {
		case jobspb.EncryptionMode_Passphrase:
			__antithesis_instrumentation__.Notify(8301)
			countSource("backup.encryption.passphrase")
		case jobspb.EncryptionMode_KMS:
			__antithesis_instrumentation__.Notify(8302)
			countSource("backup.encryption.kms")
		default:
			__antithesis_instrumentation__.Notify(8303)
		}
	} else {
		__antithesis_instrumentation__.Notify(8304)
	}
	__antithesis_instrumentation__.Notify(8282)
	if backupDetails.CollectionURI != "" {
		__antithesis_instrumentation__.Notify(8305)
		countSource("backup.nested")
		timeBaseSubdir := true
		if _, err := time.Parse(DateBasedIntoFolderName,
			initialDetails.Destination.Subdir); err != nil {
			__antithesis_instrumentation__.Notify(8307)
			timeBaseSubdir = false
		} else {
			__antithesis_instrumentation__.Notify(8308)
		}
		__antithesis_instrumentation__.Notify(8306)
		if backupDetails.StartTime.IsEmpty() {
			__antithesis_instrumentation__.Notify(8309)
			if !timeBaseSubdir {
				__antithesis_instrumentation__.Notify(8310)
				countSource("backup.deprecated-full-nontime-subdir")
			} else {
				__antithesis_instrumentation__.Notify(8311)
				if initialDetails.Destination.Exists {
					__antithesis_instrumentation__.Notify(8312)
					countSource("backup.deprecated-full-time-subdir")
				} else {
					__antithesis_instrumentation__.Notify(8313)
					countSource("backup.full-no-subdir")
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(8314)
			if initialDetails.Destination.Subdir == latestFileName {
				__antithesis_instrumentation__.Notify(8315)
				countSource("backup.incremental-latest-subdir")
			} else {
				__antithesis_instrumentation__.Notify(8316)
				if !timeBaseSubdir {
					__antithesis_instrumentation__.Notify(8317)
					countSource("backup.deprecated-incremental-nontime-subdir")
				} else {
					__antithesis_instrumentation__.Notify(8318)
					countSource("backup.incremental-explicit-subdir")
				}
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(8319)
		countSource("backup.deprecated-non-collection")
	}
	__antithesis_instrumentation__.Notify(8283)
	if backupManifest.DescriptorCoverage == tree.AllDescriptors {
		__antithesis_instrumentation__.Notify(8320)
		countSource("backup.targets.full_cluster")
	} else {
		__antithesis_instrumentation__.Notify(8321)
	}
}

func getScheduledBackupExecutionArgsFromSchedule(
	ctx context.Context,
	env scheduledjobs.JobSchedulerEnv,
	txn *kv.Txn,
	ie *sql.InternalExecutor,
	scheduleID int64,
) (*jobs.ScheduledJob, *ScheduledBackupExecutionArgs, error) {
	__antithesis_instrumentation__.Notify(8322)

	sj, err := jobs.LoadScheduledJob(ctx, env, scheduleID, ie, txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(8325)
		return nil, nil, errors.Wrapf(err, "failed to load scheduled job %d", scheduleID)
	} else {
		__antithesis_instrumentation__.Notify(8326)
	}
	__antithesis_instrumentation__.Notify(8323)

	args := &ScheduledBackupExecutionArgs{}
	if err := pbtypes.UnmarshalAny(sj.ExecutionArgs().Args, args); err != nil {
		__antithesis_instrumentation__.Notify(8327)
		return nil, nil, errors.Wrap(err, "un-marshaling args")
	} else {
		__antithesis_instrumentation__.Notify(8328)
	}
	__antithesis_instrumentation__.Notify(8324)

	return sj, args, nil
}

func writeBackupManifestCheckpoint(
	ctx context.Context,
	storageURI string,
	encryption *jobspb.BackupEncryptionOptions,
	desc *BackupManifest,
	execCfg *sql.ExecutorConfig,
	user security.SQLUsername,
) error {
	__antithesis_instrumentation__.Notify(8329)
	defaultStore, err := execCfg.DistSQLSrv.ExternalStorageFromURI(ctx, storageURI, user)
	if err != nil {
		__antithesis_instrumentation__.Notify(8339)
		return err
	} else {
		__antithesis_instrumentation__.Notify(8340)
	}
	__antithesis_instrumentation__.Notify(8330)
	defer defaultStore.Close()

	sort.Sort(BackupFileDescriptors(desc.Files))

	descBuf, err := protoutil.Marshal(desc)
	if err != nil {
		__antithesis_instrumentation__.Notify(8341)
		return err
	} else {
		__antithesis_instrumentation__.Notify(8342)
	}
	__antithesis_instrumentation__.Notify(8331)

	descBuf, err = compressData(descBuf)
	if err != nil {
		__antithesis_instrumentation__.Notify(8343)
		return errors.Wrap(err, "compressing backup manifest")
	} else {
		__antithesis_instrumentation__.Notify(8344)
	}
	__antithesis_instrumentation__.Notify(8332)

	if encryption != nil {
		__antithesis_instrumentation__.Notify(8345)
		encryptionKey, err := getEncryptionKey(ctx, encryption, execCfg.Settings, defaultStore.ExternalIOConf())
		if err != nil {
			__antithesis_instrumentation__.Notify(8347)
			return err
		} else {
			__antithesis_instrumentation__.Notify(8348)
		}
		__antithesis_instrumentation__.Notify(8346)
		descBuf, err = storageccl.EncryptFile(descBuf, encryptionKey)
		if err != nil {
			__antithesis_instrumentation__.Notify(8349)
			return err
		} else {
			__antithesis_instrumentation__.Notify(8350)
		}
	} else {
		__antithesis_instrumentation__.Notify(8351)
	}
	__antithesis_instrumentation__.Notify(8333)

	if !execCfg.Settings.Version.IsActive(ctx, clusterversion.BackupDoesNotOverwriteLatestAndCheckpoint) {
		__antithesis_instrumentation__.Notify(8352)

		err = cloud.WriteFile(ctx, defaultStore, backupManifestCheckpointName, bytes.NewReader(descBuf))
		if err != nil {
			__antithesis_instrumentation__.Notify(8355)
			return err
		} else {
			__antithesis_instrumentation__.Notify(8356)
		}
		__antithesis_instrumentation__.Notify(8353)

		checksum, err := getChecksum(descBuf)
		if err != nil {
			__antithesis_instrumentation__.Notify(8357)
			return err
		} else {
			__antithesis_instrumentation__.Notify(8358)
		}
		__antithesis_instrumentation__.Notify(8354)

		return cloud.WriteFile(ctx, defaultStore, backupManifestCheckpointName+backupManifestChecksumSuffix, bytes.NewReader(checksum))
	} else {
		__antithesis_instrumentation__.Notify(8359)
	}
	__antithesis_instrumentation__.Notify(8334)

	filename := newTimestampedCheckpointFileName()

	if defaultStore.Conf().Provider == roachpb.ExternalStorageProvider_http {
		__antithesis_instrumentation__.Notify(8360)

		if r, err := defaultStore.ReadFile(ctx, backupProgressDirectory+"/"+backupManifestCheckpointName); err != nil {
			__antithesis_instrumentation__.Notify(8361)

			filename = backupManifestCheckpointName
		} else {
			__antithesis_instrumentation__.Notify(8362)
			err = r.Close(ctx)
			if err != nil {
				__antithesis_instrumentation__.Notify(8363)
				return err
			} else {
				__antithesis_instrumentation__.Notify(8364)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(8365)
	}
	__antithesis_instrumentation__.Notify(8335)

	err = cloud.WriteFile(ctx, defaultStore, backupProgressDirectory+"/"+filename, bytes.NewReader(descBuf))
	if err != nil {
		__antithesis_instrumentation__.Notify(8366)
		return errors.Wrap(err, "calculating checksum")
	} else {
		__antithesis_instrumentation__.Notify(8367)
	}
	__antithesis_instrumentation__.Notify(8336)

	checksum, err := getChecksum(descBuf)
	if err != nil {
		__antithesis_instrumentation__.Notify(8368)
		return errors.Wrap(err, "calculating checksum")
	} else {
		__antithesis_instrumentation__.Notify(8369)
	}
	__antithesis_instrumentation__.Notify(8337)

	err = cloud.WriteFile(ctx, defaultStore, backupProgressDirectory+"/"+filename+backupManifestChecksumSuffix, bytes.NewReader(checksum))
	if err != nil {
		__antithesis_instrumentation__.Notify(8370)
		return err
	} else {
		__antithesis_instrumentation__.Notify(8371)
	}
	__antithesis_instrumentation__.Notify(8338)

	return nil
}

func planSchedulePTSChaining(
	ctx context.Context,
	execCfg *sql.ExecutorConfig,
	txn *kv.Txn,
	backupDetails *jobspb.BackupDetails,
	createdBy *jobs.CreatedByInfo,
) error {
	__antithesis_instrumentation__.Notify(8372)
	env := scheduledjobs.ProdJobSchedulerEnv
	if knobs, ok := execCfg.DistSQLSrv.TestingKnobs.JobsTestingKnobs.(*jobs.TestingKnobs); ok {
		__antithesis_instrumentation__.Notify(8378)
		if knobs.JobSchedulerEnv != nil {
			__antithesis_instrumentation__.Notify(8379)
			env = knobs.JobSchedulerEnv
		} else {
			__antithesis_instrumentation__.Notify(8380)
		}
	} else {
		__antithesis_instrumentation__.Notify(8381)
	}
	__antithesis_instrumentation__.Notify(8373)

	if createdBy == nil || func() bool {
		__antithesis_instrumentation__.Notify(8382)
		return createdBy.Name != jobs.CreatedByScheduledJobs == true
	}() == true {
		__antithesis_instrumentation__.Notify(8383)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(8384)
	}
	__antithesis_instrumentation__.Notify(8374)

	_, args, err := getScheduledBackupExecutionArgsFromSchedule(
		ctx, env, txn, execCfg.InternalExecutor, createdBy.ID)
	if err != nil {
		__antithesis_instrumentation__.Notify(8385)
		return err
	} else {
		__antithesis_instrumentation__.Notify(8386)
	}
	__antithesis_instrumentation__.Notify(8375)
	if !args.ChainProtectedTimestampRecords {
		__antithesis_instrumentation__.Notify(8387)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(8388)
	}
	__antithesis_instrumentation__.Notify(8376)

	if args.BackupType == ScheduledBackupExecutionArgs_FULL {
		__antithesis_instrumentation__.Notify(8389)

		if args.DependentScheduleID == 0 {
			__antithesis_instrumentation__.Notify(8392)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(8393)
		}
		__antithesis_instrumentation__.Notify(8390)

		_, incArgs, err := getScheduledBackupExecutionArgsFromSchedule(
			ctx, env, txn, execCfg.InternalExecutor, args.DependentScheduleID)
		if err != nil {
			__antithesis_instrumentation__.Notify(8394)

			if jobs.HasScheduledJobNotFoundError(err) {
				__antithesis_instrumentation__.Notify(8396)
				log.Warningf(ctx, "could not find dependent schedule with id %d",
					args.DependentScheduleID)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(8397)
			}
			__antithesis_instrumentation__.Notify(8395)
			return err
		} else {
			__antithesis_instrumentation__.Notify(8398)
		}
		__antithesis_instrumentation__.Notify(8391)
		backupDetails.SchedulePTSChainingRecord = &jobspb.SchedulePTSChainingRecord{
			ProtectedTimestampRecord: incArgs.ProtectedTimestampRecord,
			Action:                   jobspb.SchedulePTSChainingRecord_RELEASE,
		}
	} else {
		__antithesis_instrumentation__.Notify(8399)

		backupDetails.SchedulePTSChainingRecord = &jobspb.SchedulePTSChainingRecord{
			ProtectedTimestampRecord: args.ProtectedTimestampRecord,
			Action:                   jobspb.SchedulePTSChainingRecord_UPDATE,
		}
	}
	__antithesis_instrumentation__.Notify(8377)
	return nil
}

func getReintroducedSpans(
	ctx context.Context,
	execCfg *sql.ExecutorConfig,
	prevBackups []BackupManifest,
	tables []catalog.TableDescriptor,
	revs []BackupManifest_DescriptorRevision,
	endTime hlc.Timestamp,
) ([]roachpb.Span, error) {
	__antithesis_instrumentation__.Notify(8400)
	reintroducedTables := make(map[descpb.ID]struct{})

	offlineInLastBackup := make(map[descpb.ID]struct{})
	lastBackup := prevBackups[len(prevBackups)-1]
	for _, desc := range lastBackup.Descriptors {
		__antithesis_instrumentation__.Notify(8406)

		if table, _, _, _ := descpb.FromDescriptor(&desc); table != nil && func() bool {
			__antithesis_instrumentation__.Notify(8407)
			return table.Offline() == true
		}() == true {
			__antithesis_instrumentation__.Notify(8408)
			offlineInLastBackup[table.GetID()] = struct{}{}
		} else {
			__antithesis_instrumentation__.Notify(8409)
		}
	}
	__antithesis_instrumentation__.Notify(8401)

	tablesToReinclude := make([]catalog.TableDescriptor, 0)
	for _, desc := range tables {
		__antithesis_instrumentation__.Notify(8410)
		if _, wasOffline := offlineInLastBackup[desc.GetID()]; wasOffline && func() bool {
			__antithesis_instrumentation__.Notify(8411)
			return desc.Public() == true
		}() == true {
			__antithesis_instrumentation__.Notify(8412)
			tablesToReinclude = append(tablesToReinclude, desc)
			reintroducedTables[desc.GetID()] = struct{}{}
		} else {
			__antithesis_instrumentation__.Notify(8413)
		}
	}
	__antithesis_instrumentation__.Notify(8402)

	for _, rev := range revs {
		__antithesis_instrumentation__.Notify(8414)
		rawTable, _, _, _ := descpb.FromDescriptor(rev.Desc)
		if rawTable == nil {
			__antithesis_instrumentation__.Notify(8416)
			continue
		} else {
			__antithesis_instrumentation__.Notify(8417)
		}
		__antithesis_instrumentation__.Notify(8415)
		table := tabledesc.NewBuilder(rawTable).BuildImmutableTable()
		if _, wasOffline := offlineInLastBackup[table.GetID()]; wasOffline && func() bool {
			__antithesis_instrumentation__.Notify(8418)
			return table.Public() == true
		}() == true {
			__antithesis_instrumentation__.Notify(8419)
			tablesToReinclude = append(tablesToReinclude, table)
			reintroducedTables[table.GetID()] = struct{}{}
		} else {
			__antithesis_instrumentation__.Notify(8420)
		}
	}
	__antithesis_instrumentation__.Notify(8403)

	allRevs := make([]BackupManifest_DescriptorRevision, 0, len(revs))
	for _, rev := range revs {
		__antithesis_instrumentation__.Notify(8421)
		rawTable, _, _, _ := descpb.FromDescriptor(rev.Desc)
		if rawTable == nil {
			__antithesis_instrumentation__.Notify(8423)
			continue
		} else {
			__antithesis_instrumentation__.Notify(8424)
		}
		__antithesis_instrumentation__.Notify(8422)
		if _, ok := reintroducedTables[rawTable.GetID()]; ok {
			__antithesis_instrumentation__.Notify(8425)
			allRevs = append(allRevs, rev)
		} else {
			__antithesis_instrumentation__.Notify(8426)
		}
	}
	__antithesis_instrumentation__.Notify(8404)

	tableSpans, err := spansForAllTableIndexes(execCfg, tablesToReinclude, allRevs)
	if err != nil {
		__antithesis_instrumentation__.Notify(8427)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(8428)
	}
	__antithesis_instrumentation__.Notify(8405)
	return tableSpans, nil
}

func makeNewEncryptionOptions(
	ctx context.Context, encryptionParams jobspb.BackupEncryptionOptions, kmsEnv cloud.KMSEnv,
) (*jobspb.BackupEncryptionOptions, *jobspb.EncryptionInfo, error) {
	__antithesis_instrumentation__.Notify(8429)
	var encryptionOptions *jobspb.BackupEncryptionOptions
	var encryptionInfo *jobspb.EncryptionInfo
	switch encryptionParams.Mode {
	case jobspb.EncryptionMode_Passphrase:
		__antithesis_instrumentation__.Notify(8431)
		salt, err := storageccl.GenerateSalt()
		if err != nil {
			__antithesis_instrumentation__.Notify(8438)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(8439)
		}
		__antithesis_instrumentation__.Notify(8432)

		encryptionInfo = &jobspb.EncryptionInfo{Salt: salt}
		encryptionOptions = &jobspb.BackupEncryptionOptions{
			Mode: jobspb.EncryptionMode_Passphrase,
			Key:  storageccl.GenerateKey([]byte(encryptionParams.RawPassphrae), salt),
		}
	case jobspb.EncryptionMode_KMS:
		__antithesis_instrumentation__.Notify(8433)

		plaintextDataKey := make([]byte, 32)
		_, err := cryptorand.Read(plaintextDataKey)
		if err != nil {
			__antithesis_instrumentation__.Notify(8440)
			return nil, nil, errors.Wrap(err, "failed to generate DataKey")
		} else {
			__antithesis_instrumentation__.Notify(8441)
		}
		__antithesis_instrumentation__.Notify(8434)

		encryptedDataKeyByKMSMasterKeyID, defaultKMSInfo, err :=
			getEncryptedDataKeyByKMSMasterKeyID(ctx, encryptionParams.RawKmsUris, plaintextDataKey, kmsEnv)
		if err != nil {
			__antithesis_instrumentation__.Notify(8442)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(8443)
		}
		__antithesis_instrumentation__.Notify(8435)

		encryptedDataKeyMapForProto := make(map[string][]byte)
		encryptedDataKeyByKMSMasterKeyID.rangeOverMap(
			func(masterKeyID hashedMasterKeyID, dataKey []byte) {
				__antithesis_instrumentation__.Notify(8444)
				encryptedDataKeyMapForProto[string(masterKeyID)] = dataKey
			})
		__antithesis_instrumentation__.Notify(8436)

		encryptionInfo = &jobspb.EncryptionInfo{EncryptedDataKeyByKMSMasterKeyID: encryptedDataKeyMapForProto}
		encryptionOptions = &jobspb.BackupEncryptionOptions{
			Mode:    jobspb.EncryptionMode_KMS,
			KMSInfo: defaultKMSInfo,
		}
	default:
		__antithesis_instrumentation__.Notify(8437)
	}
	__antithesis_instrumentation__.Notify(8430)
	return encryptionOptions, encryptionInfo, nil
}

func getProtectedTimestampTargetForBackup(backupManifest BackupManifest) *ptpb.Target {
	__antithesis_instrumentation__.Notify(8445)
	if backupManifest.DescriptorCoverage == tree.AllDescriptors {
		__antithesis_instrumentation__.Notify(8450)
		return ptpb.MakeClusterTarget()
	} else {
		__antithesis_instrumentation__.Notify(8451)
	}
	__antithesis_instrumentation__.Notify(8446)

	if len(backupManifest.Tenants) > 0 {
		__antithesis_instrumentation__.Notify(8452)
		tenantID := make([]roachpb.TenantID, 0)
		for _, tenant := range backupManifest.Tenants {
			__antithesis_instrumentation__.Notify(8454)
			tenantID = append(tenantID, roachpb.MakeTenantID(tenant.ID))
		}
		__antithesis_instrumentation__.Notify(8453)
		return ptpb.MakeTenantsTarget(tenantID)
	} else {
		__antithesis_instrumentation__.Notify(8455)
	}
	__antithesis_instrumentation__.Notify(8447)

	if len(backupManifest.CompleteDbs) > 0 {
		__antithesis_instrumentation__.Notify(8456)
		return ptpb.MakeSchemaObjectsTarget(backupManifest.CompleteDbs)
	} else {
		__antithesis_instrumentation__.Notify(8457)
	}
	__antithesis_instrumentation__.Notify(8448)

	tableIDs := make(descpb.IDs, 0)
	for _, desc := range backupManifest.Descriptors {
		__antithesis_instrumentation__.Notify(8458)
		t, _, _, _ := descpb.FromDescriptorWithMVCCTimestamp(&desc, hlc.Timestamp{})
		if t != nil {
			__antithesis_instrumentation__.Notify(8459)
			tableIDs = append(tableIDs, t.GetID())
		} else {
			__antithesis_instrumentation__.Notify(8460)
		}
	}
	__antithesis_instrumentation__.Notify(8449)
	return ptpb.MakeSchemaObjectsTarget(tableIDs)
}

func protectTimestampForBackup(
	ctx context.Context,
	execCfg *sql.ExecutorConfig,
	txn *kv.Txn,
	jobID jobspb.JobID,
	backupManifest BackupManifest,
	backupDetails jobspb.BackupDetails,
) error {
	__antithesis_instrumentation__.Notify(8461)
	tsToProtect := backupManifest.EndTime
	if !backupManifest.StartTime.IsEmpty() {
		__antithesis_instrumentation__.Notify(8463)
		tsToProtect = backupManifest.StartTime
	} else {
		__antithesis_instrumentation__.Notify(8464)
	}
	__antithesis_instrumentation__.Notify(8462)

	target := getProtectedTimestampTargetForBackup(backupManifest)

	target.IgnoreIfExcludedFromBackup = true
	rec := jobsprotectedts.MakeRecord(*backupDetails.ProtectedTimestampRecord, int64(jobID),
		tsToProtect, backupManifest.Spans, jobsprotectedts.Jobs, target)
	return execCfg.ProtectedTimestampProvider.Protect(ctx, txn, rec)
}

func getEncryptedDataKeyFromURI(
	ctx context.Context, plaintextDataKey []byte, kmsURI string, kmsEnv cloud.KMSEnv,
) (string, []byte, error) {
	__antithesis_instrumentation__.Notify(8465)
	kms, err := cloud.KMSFromURI(kmsURI, kmsEnv)
	if err != nil {
		__antithesis_instrumentation__.Notify(8471)
		return "", nil, err
	} else {
		__antithesis_instrumentation__.Notify(8472)
	}
	__antithesis_instrumentation__.Notify(8466)

	defer func() {
		__antithesis_instrumentation__.Notify(8473)
		_ = kms.Close()
	}()
	__antithesis_instrumentation__.Notify(8467)

	kmsURL, err := url.ParseRequestURI(kmsURI)
	if err != nil {
		__antithesis_instrumentation__.Notify(8474)
		return "", nil, errors.Wrap(err, "cannot parse KMSURI")
	} else {
		__antithesis_instrumentation__.Notify(8475)
	}
	__antithesis_instrumentation__.Notify(8468)
	encryptedDataKey, err := kms.Encrypt(ctx, plaintextDataKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(8476)
		return "", nil, errors.Wrapf(err, "failed to encrypt data key for KMS scheme %s",
			kmsURL.Scheme)
	} else {
		__antithesis_instrumentation__.Notify(8477)
	}
	__antithesis_instrumentation__.Notify(8469)

	masterKeyID, err := kms.MasterKeyID()
	if err != nil {
		__antithesis_instrumentation__.Notify(8478)
		return "", nil, errors.Wrapf(err, "failed to get master key ID for KMS scheme %s",
			kmsURL.Scheme)
	} else {
		__antithesis_instrumentation__.Notify(8479)
	}
	__antithesis_instrumentation__.Notify(8470)

	return masterKeyID, encryptedDataKey, nil
}

func getEncryptedDataKeyByKMSMasterKeyID(
	ctx context.Context, kmsURIs []string, plaintextDataKey []byte, kmsEnv cloud.KMSEnv,
) (*encryptedDataKeyMap, *jobspb.BackupEncryptionOptions_KMSInfo, error) {
	__antithesis_instrumentation__.Notify(8480)
	encryptedDataKeyByKMSMasterKeyID := newEncryptedDataKeyMap()

	var kmsInfo *jobspb.BackupEncryptionOptions_KMSInfo
	for _, kmsURI := range kmsURIs {
		__antithesis_instrumentation__.Notify(8482)
		masterKeyID, encryptedDataKey, err := getEncryptedDataKeyFromURI(ctx,
			plaintextDataKey, kmsURI, kmsEnv)
		if err != nil {
			__antithesis_instrumentation__.Notify(8485)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(8486)
		}
		__antithesis_instrumentation__.Notify(8483)

		if kmsInfo == nil {
			__antithesis_instrumentation__.Notify(8487)
			kmsInfo = &jobspb.BackupEncryptionOptions_KMSInfo{
				Uri:              kmsURI,
				EncryptedDataKey: encryptedDataKey,
			}
		} else {
			__antithesis_instrumentation__.Notify(8488)
		}
		__antithesis_instrumentation__.Notify(8484)

		encryptedDataKeyByKMSMasterKeyID.addEncryptedDataKey(plaintextMasterKeyID(masterKeyID),
			encryptedDataKey)
	}
	__antithesis_instrumentation__.Notify(8481)

	return encryptedDataKeyByKMSMasterKeyID, kmsInfo, nil
}

func checkForNewCompleteDatabases(
	targetDescs []catalog.Descriptor, curDBs []descpb.ID, prevDBs map[descpb.ID]struct{},
) error {
	__antithesis_instrumentation__.Notify(8489)
	for _, dbID := range curDBs {
		__antithesis_instrumentation__.Notify(8491)
		if _, inPrevious := prevDBs[dbID]; !inPrevious {
			__antithesis_instrumentation__.Notify(8492)

			violatingDatabase := strconv.Itoa(int(dbID))
			for _, desc := range targetDescs {
				__antithesis_instrumentation__.Notify(8494)
				if desc.GetID() == dbID {
					__antithesis_instrumentation__.Notify(8495)
					violatingDatabase = desc.GetName()
					break
				} else {
					__antithesis_instrumentation__.Notify(8496)
				}
			}
			__antithesis_instrumentation__.Notify(8493)
			return errors.Errorf("previous backup does not contain the complete database %q",
				violatingDatabase)
		} else {
			__antithesis_instrumentation__.Notify(8497)
		}
	}
	__antithesis_instrumentation__.Notify(8490)
	return nil
}

func checkForNewTables(
	ctx context.Context,
	codec keys.SQLCodec,
	db *kv.DB,
	targetDescs []catalog.Descriptor,
	tablesInPrev map[descpb.ID]struct{},
	dbsInPrev map[descpb.ID]struct{},
	priorIDs map[descpb.ID]descpb.ID,
	startTime hlc.Timestamp,
	endTime hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(8498)
	for _, d := range targetDescs {
		__antithesis_instrumentation__.Notify(8500)
		t, ok := d.(catalog.TableDescriptor)
		if !ok {
			__antithesis_instrumentation__.Notify(8505)
			continue
		} else {
			__antithesis_instrumentation__.Notify(8506)
		}
		__antithesis_instrumentation__.Notify(8501)

		if _, ok := tablesInPrev[t.GetID()]; ok {
			__antithesis_instrumentation__.Notify(8507)
			continue
		} else {
			__antithesis_instrumentation__.Notify(8508)
		}
		__antithesis_instrumentation__.Notify(8502)

		if _, ok := dbsInPrev[t.GetParentID()]; ok {
			__antithesis_instrumentation__.Notify(8509)
			continue
		} else {
			__antithesis_instrumentation__.Notify(8510)
		}
		__antithesis_instrumentation__.Notify(8503)

		if replacement := t.GetReplacementOf(); replacement.ID != descpb.InvalidID {
			__antithesis_instrumentation__.Notify(8511)

			if priorIDs == nil {
				__antithesis_instrumentation__.Notify(8514)
				priorIDs = make(map[descpb.ID]descpb.ID)
				_, err := getAllDescChanges(ctx, codec, db, startTime, endTime, priorIDs)
				if err != nil {
					__antithesis_instrumentation__.Notify(8515)
					return err
				} else {
					__antithesis_instrumentation__.Notify(8516)
				}
			} else {
				__antithesis_instrumentation__.Notify(8517)
			}
			__antithesis_instrumentation__.Notify(8512)
			found := false
			for was := replacement.ID; was != descpb.InvalidID && func() bool {
				__antithesis_instrumentation__.Notify(8518)
				return !found == true
			}() == true; was = priorIDs[was] {
				__antithesis_instrumentation__.Notify(8519)
				_, found = tablesInPrev[was]
			}
			__antithesis_instrumentation__.Notify(8513)
			if found {
				__antithesis_instrumentation__.Notify(8520)
				continue
			} else {
				__antithesis_instrumentation__.Notify(8521)
			}
		} else {
			__antithesis_instrumentation__.Notify(8522)
		}
		__antithesis_instrumentation__.Notify(8504)
		return errors.Errorf("previous backup does not contain table %q", t.GetName())
	}
	__antithesis_instrumentation__.Notify(8499)
	return nil
}

func getBackupDetailAndManifest(
	ctx context.Context,
	execCfg *sql.ExecutorConfig,
	txn *kv.Txn,
	initialDetails jobspb.BackupDetails,
	user security.SQLUsername,
) (jobspb.BackupDetails, BackupManifest, error) {
	__antithesis_instrumentation__.Notify(8523)
	makeCloudStorage := execCfg.DistSQLSrv.ExternalStorageFromURI

	mvccFilter := MVCCFilter_Latest
	if initialDetails.RevisionHistory {
		__antithesis_instrumentation__.Notify(8543)
		mvccFilter = MVCCFilter_All
	} else {
		__antithesis_instrumentation__.Notify(8544)
	}
	__antithesis_instrumentation__.Notify(8524)
	endTime := initialDetails.EndTime
	var targetDescs []catalog.Descriptor
	var descriptorProtos []descpb.Descriptor
	if initialDetails.FullCluster {
		__antithesis_instrumentation__.Notify(8545)
		var err error
		targetDescs, _, err = fullClusterTargetsBackup(ctx, execCfg, endTime)
		if err != nil {
			__antithesis_instrumentation__.Notify(8547)
			return jobspb.BackupDetails{}, BackupManifest{}, err
		} else {
			__antithesis_instrumentation__.Notify(8548)
		}
		__antithesis_instrumentation__.Notify(8546)
		descriptorProtos = make([]descpb.Descriptor, len(targetDescs))
		for i, desc := range targetDescs {
			__antithesis_instrumentation__.Notify(8549)
			descriptorProtos[i] = *desc.DescriptorProto()
		}
	} else {
		__antithesis_instrumentation__.Notify(8550)
		descriptorProtos = initialDetails.ResolvedTargets
		targetDescs = make([]catalog.Descriptor, len(descriptorProtos))
		for i := range descriptorProtos {
			__antithesis_instrumentation__.Notify(8551)
			targetDescs[i] = descbuilder.NewBuilder(&descriptorProtos[i]).BuildExistingMutable()
		}
	}
	__antithesis_instrumentation__.Notify(8525)

	collectionURI, defaultURI, resolvedSubdir, urisByLocalityKV, prevs, err :=
		resolveDest(ctx, user, initialDetails.Destination, endTime, initialDetails.IncrementalFrom, execCfg)
	if err != nil {
		__antithesis_instrumentation__.Notify(8552)
		return jobspb.BackupDetails{}, BackupManifest{}, err
	} else {
		__antithesis_instrumentation__.Notify(8553)
	}
	__antithesis_instrumentation__.Notify(8526)

	defaultStore, err := execCfg.DistSQLSrv.ExternalStorageFromURI(ctx, defaultURI, user)
	if err != nil {
		__antithesis_instrumentation__.Notify(8554)
		return jobspb.BackupDetails{}, BackupManifest{}, err
	} else {
		__antithesis_instrumentation__.Notify(8555)
	}
	__antithesis_instrumentation__.Notify(8527)
	defer defaultStore.Close()

	if err := checkForPreviousBackup(ctx, defaultStore, defaultURI); err != nil {
		__antithesis_instrumentation__.Notify(8556)
		return jobspb.BackupDetails{}, BackupManifest{}, err
	} else {
		__antithesis_instrumentation__.Notify(8557)
	}
	__antithesis_instrumentation__.Notify(8528)

	kmsEnv := &backupKMSEnv{settings: execCfg.Settings, conf: &execCfg.ExternalIODirConfig}

	mem := execCfg.RootMemoryMonitor.MakeBoundAccount()
	defer mem.Close(ctx)

	prevBackups, encryptionOptions, memSize, err := fetchPreviousBackups(ctx, &mem, user,
		makeCloudStorage, prevs, *initialDetails.EncryptionOptions, kmsEnv)
	if err != nil {
		__antithesis_instrumentation__.Notify(8558)
		return jobspb.BackupDetails{}, BackupManifest{}, err
	} else {
		__antithesis_instrumentation__.Notify(8559)
	}
	__antithesis_instrumentation__.Notify(8529)
	defer func() {
		__antithesis_instrumentation__.Notify(8560)
		mem.Shrink(ctx, memSize)
	}()
	__antithesis_instrumentation__.Notify(8530)

	if len(prevBackups) > 0 {
		__antithesis_instrumentation__.Notify(8561)
		baseManifest := prevBackups[0]
		if baseManifest.DescriptorCoverage == tree.AllDescriptors && func() bool {
			__antithesis_instrumentation__.Notify(8562)
			return !initialDetails.FullCluster == true
		}() == true {
			__antithesis_instrumentation__.Notify(8563)
			return jobspb.BackupDetails{}, BackupManifest{}, errors.Errorf("cannot append a backup of specific tables or databases to a cluster backup")
		} else {
			__antithesis_instrumentation__.Notify(8564)
		}
	} else {
		__antithesis_instrumentation__.Notify(8565)
	}
	__antithesis_instrumentation__.Notify(8531)

	var startTime hlc.Timestamp
	if len(prevBackups) > 0 {
		__antithesis_instrumentation__.Notify(8566)
		if err := requireEnterprise(execCfg, "incremental"); err != nil {
			__antithesis_instrumentation__.Notify(8569)
			return jobspb.BackupDetails{}, BackupManifest{}, err
		} else {
			__antithesis_instrumentation__.Notify(8570)
		}
		__antithesis_instrumentation__.Notify(8567)
		lastEndTime := prevBackups[len(prevBackups)-1].EndTime
		if lastEndTime.Compare(initialDetails.EndTime) > 0 {
			__antithesis_instrumentation__.Notify(8571)
			return jobspb.BackupDetails{}, BackupManifest{},
				errors.Newf("`AS OF SYSTEM TIME` %s must be greater than "+
					"the previous backup's end time of %s.",
					initialDetails.EndTime.GoTime(), lastEndTime.GoTime())
		} else {
			__antithesis_instrumentation__.Notify(8572)
		}
		__antithesis_instrumentation__.Notify(8568)
		startTime = prevBackups[len(prevBackups)-1].EndTime
	} else {
		__antithesis_instrumentation__.Notify(8573)
	}
	__antithesis_instrumentation__.Notify(8532)

	var tables []catalog.TableDescriptor
	statsFiles := make(map[descpb.ID]string)
	for _, desc := range targetDescs {
		__antithesis_instrumentation__.Notify(8574)
		switch desc := desc.(type) {
		case catalog.TableDescriptor:
			__antithesis_instrumentation__.Notify(8575)
			tables = append(tables, desc)

			statsFiles[desc.GetID()] = backupStatisticsFileName
		}
	}
	__antithesis_instrumentation__.Notify(8533)

	clusterID := execCfg.LogicalClusterID()
	for i := range prevBackups {
		__antithesis_instrumentation__.Notify(8576)

		if fromCluster := prevBackups[i].ClusterID; !fromCluster.Equal(clusterID) {
			__antithesis_instrumentation__.Notify(8577)
			return jobspb.BackupDetails{}, BackupManifest{}, errors.Newf("previous BACKUP belongs to cluster %s", fromCluster.String())
		} else {
			__antithesis_instrumentation__.Notify(8578)
		}
	}
	__antithesis_instrumentation__.Notify(8534)

	var newSpans roachpb.Spans

	var priorIDs map[descpb.ID]descpb.ID

	var revs []BackupManifest_DescriptorRevision
	if mvccFilter == MVCCFilter_All {
		__antithesis_instrumentation__.Notify(8579)
		priorIDs = make(map[descpb.ID]descpb.ID)
		revs, err = getRelevantDescChanges(ctx, execCfg, startTime, endTime, targetDescs,
			initialDetails.ResolvedCompleteDbs, priorIDs, initialDetails.FullCluster)
		if err != nil {
			__antithesis_instrumentation__.Notify(8580)
			return jobspb.BackupDetails{}, BackupManifest{}, err
		} else {
			__antithesis_instrumentation__.Notify(8581)
		}
	} else {
		__antithesis_instrumentation__.Notify(8582)
	}
	__antithesis_instrumentation__.Notify(8535)

	var spans []roachpb.Span
	var tenants []descpb.TenantInfoWithUsage

	if initialDetails.FullCluster && func() bool {
		__antithesis_instrumentation__.Notify(8583)
		return execCfg.Codec.ForSystemTenant() == true
	}() == true {
		__antithesis_instrumentation__.Notify(8584)

		tenants, err = retrieveAllTenantsMetadata(
			ctx, execCfg.InternalExecutor, txn,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(8585)
			return jobspb.BackupDetails{}, BackupManifest{}, err
		} else {
			__antithesis_instrumentation__.Notify(8586)
		}
	} else {
		__antithesis_instrumentation__.Notify(8587)
		if len(initialDetails.SpecificTenantIds) > 0 {
			__antithesis_instrumentation__.Notify(8588)
			for _, id := range initialDetails.SpecificTenantIds {
				__antithesis_instrumentation__.Notify(8589)
				tenantInfo, err := retrieveSingleTenantMetadata(
					ctx, execCfg.InternalExecutor, txn, id,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(8591)
					return jobspb.BackupDetails{}, BackupManifest{}, err
				} else {
					__antithesis_instrumentation__.Notify(8592)
				}
				__antithesis_instrumentation__.Notify(8590)
				tenants = append(tenants, tenantInfo)
			}
		} else {
			__antithesis_instrumentation__.Notify(8593)
		}
	}
	__antithesis_instrumentation__.Notify(8536)

	if len(tenants) > 0 {
		__antithesis_instrumentation__.Notify(8594)
		if initialDetails.RevisionHistory {
			__antithesis_instrumentation__.Notify(8596)
			return jobspb.BackupDetails{}, BackupManifest{}, errors.UnimplementedError(
				errors.IssueLink{IssueURL: "https://github.com/cockroachdb/cockroach/issues/47896"},
				"can not backup tenants with revision history",
			)
		} else {
			__antithesis_instrumentation__.Notify(8597)
		}
		__antithesis_instrumentation__.Notify(8595)
		for i := range tenants {
			__antithesis_instrumentation__.Notify(8598)
			prefix := keys.MakeTenantPrefix(roachpb.MakeTenantID(tenants[i].ID))
			spans = append(spans, roachpb.Span{Key: prefix, EndKey: prefix.PrefixEnd()})
		}
	} else {
		__antithesis_instrumentation__.Notify(8599)
	}
	__antithesis_instrumentation__.Notify(8537)

	tableSpans, err := spansForAllTableIndexes(execCfg, tables, revs)
	if err != nil {
		__antithesis_instrumentation__.Notify(8600)
		return jobspb.BackupDetails{}, BackupManifest{}, err
	} else {
		__antithesis_instrumentation__.Notify(8601)
	}
	__antithesis_instrumentation__.Notify(8538)
	spans = append(spans, tableSpans...)

	if len(prevBackups) > 0 {
		__antithesis_instrumentation__.Notify(8602)
		tablesInPrev := make(map[descpb.ID]struct{})
		dbsInPrev := make(map[descpb.ID]struct{})
		rawDescs := prevBackups[len(prevBackups)-1].Descriptors
		for i := range rawDescs {
			__antithesis_instrumentation__.Notify(8607)
			if t, _, _, _ := descpb.FromDescriptor(&rawDescs[i]); t != nil {
				__antithesis_instrumentation__.Notify(8608)
				tablesInPrev[t.ID] = struct{}{}
			} else {
				__antithesis_instrumentation__.Notify(8609)
			}
		}
		__antithesis_instrumentation__.Notify(8603)
		for _, d := range prevBackups[len(prevBackups)-1].CompleteDbs {
			__antithesis_instrumentation__.Notify(8610)
			dbsInPrev[d] = struct{}{}
		}
		__antithesis_instrumentation__.Notify(8604)

		if !initialDetails.FullCluster {
			__antithesis_instrumentation__.Notify(8611)
			if err := checkForNewTables(ctx, execCfg.Codec, execCfg.DB, targetDescs, tablesInPrev, dbsInPrev, priorIDs, startTime, endTime); err != nil {
				__antithesis_instrumentation__.Notify(8613)
				return jobspb.BackupDetails{}, BackupManifest{}, err
			} else {
				__antithesis_instrumentation__.Notify(8614)
			}
			__antithesis_instrumentation__.Notify(8612)

			if err := checkForNewCompleteDatabases(targetDescs, initialDetails.ResolvedCompleteDbs, dbsInPrev); err != nil {
				__antithesis_instrumentation__.Notify(8615)
				return jobspb.BackupDetails{}, BackupManifest{}, err
			} else {
				__antithesis_instrumentation__.Notify(8616)
			}
		} else {
			__antithesis_instrumentation__.Notify(8617)
		}
		__antithesis_instrumentation__.Notify(8605)

		newSpans = filterSpans(spans, prevBackups[len(prevBackups)-1].Spans)

		tableSpans, err := getReintroducedSpans(ctx, execCfg, prevBackups, tables, revs, endTime)
		if err != nil {
			__antithesis_instrumentation__.Notify(8618)
			return jobspb.BackupDetails{}, BackupManifest{}, err
		} else {
			__antithesis_instrumentation__.Notify(8619)
		}
		__antithesis_instrumentation__.Notify(8606)
		newSpans = append(newSpans, tableSpans...)
	} else {
		__antithesis_instrumentation__.Notify(8620)
	}
	__antithesis_instrumentation__.Notify(8539)

	coverage := tree.RequestedDescriptors
	if initialDetails.FullCluster {
		__antithesis_instrumentation__.Notify(8621)
		coverage = tree.AllDescriptors
	} else {
		__antithesis_instrumentation__.Notify(8622)
	}
	__antithesis_instrumentation__.Notify(8540)

	backupManifest := BackupManifest{
		StartTime:           startTime,
		EndTime:             endTime,
		MVCCFilter:          mvccFilter,
		Descriptors:         descriptorProtos,
		Tenants:             tenants,
		DescriptorChanges:   revs,
		CompleteDbs:         initialDetails.ResolvedCompleteDbs,
		Spans:               spans,
		IntroducedSpans:     newSpans,
		FormatVersion:       BackupFormatDescriptorTrackingVersion,
		BuildInfo:           build.GetInfo(),
		ClusterVersion:      execCfg.Settings.Version.ActiveVersion(ctx).Version,
		ClusterID:           execCfg.LogicalClusterID(),
		StatisticsFilenames: statsFiles,
		DescriptorCoverage:  coverage,
	}

	if err := checkCoverage(ctx, spans, append(prevBackups, backupManifest)); err != nil {
		__antithesis_instrumentation__.Notify(8623)
		return jobspb.BackupDetails{}, BackupManifest{}, errors.Wrap(err, "new backup would not cover expected time")
	} else {
		__antithesis_instrumentation__.Notify(8624)
	}
	__antithesis_instrumentation__.Notify(8541)

	var encryptionInfo *jobspb.EncryptionInfo
	if encryptionOptions == nil {
		__antithesis_instrumentation__.Notify(8625)
		encryptionOptions, encryptionInfo, err = makeNewEncryptionOptions(ctx, *initialDetails.EncryptionOptions, kmsEnv)
		if err != nil {
			__antithesis_instrumentation__.Notify(8626)
			return jobspb.BackupDetails{}, BackupManifest{}, err
		} else {
			__antithesis_instrumentation__.Notify(8627)
		}
	} else {
		__antithesis_instrumentation__.Notify(8628)
	}
	__antithesis_instrumentation__.Notify(8542)

	return jobspb.BackupDetails{
		Destination:       jobspb.BackupDetails_Destination{Subdir: resolvedSubdir},
		StartTime:         startTime,
		EndTime:           endTime,
		URI:               defaultURI,
		URIsByLocalityKV:  urisByLocalityKV,
		EncryptionOptions: encryptionOptions,
		EncryptionInfo:    encryptionInfo,
		CollectionURI:     collectionURI,
	}, backupManifest, nil
}

func init() {
	sql.AddPlanHook("backup", backupPlanHook)
}
