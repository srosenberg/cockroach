package backupccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/ccl/storageccl"
	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descbuilder"
	descpb "github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/ioctx"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

const (
	backupManifestName = "BACKUP_MANIFEST"

	backupOldManifestName = "BACKUP"

	backupManifestChecksumSuffix = "-CHECKSUM"

	backupPartitionDescriptorPrefix = "BACKUP_PART"

	backupManifestCheckpointName = "BACKUP-CHECKPOINT"

	backupStatisticsFileName = "BACKUP-STATISTICS"

	backupEncryptionInfoFile = "ENCRYPTION-INFO"
)

const (
	BackupFormatDescriptorTrackingVersion uint32 = 1

	backupProgressDirectory = "progress"

	DateBasedIncFolderName = "/20060102/150405.00"

	DateBasedIntoFolderName = "/2006/01/02-150405.00"

	latestFileName = "LATEST"

	latestHistoryDirectory = backupMetadataDirectory + "/" + "latest"

	backupMetadataDirectory = "metadata"
)

var backupPathRE = regexp.MustCompile("^/?[^\\/]+/[^\\/]+/[^\\/]+/" + backupManifestName + "$")

var writeMetadataSST = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"kv.bulkio.write_metadata_sst.enabled",
	"write experimental new format BACKUP metadata file",
	true,
)
var errEncryptionInfoRead = errors.New(`ENCRYPTION-INFO not found`)

func isGZipped(dat []byte) bool {
	__antithesis_instrumentation__.Notify(9843)
	gzipPrefix := []byte("\x1F\x8B\x08")
	return bytes.HasPrefix(dat, gzipPrefix)
}

type BackupFileDescriptors []BackupManifest_File

func (r BackupFileDescriptors) Len() int { __antithesis_instrumentation__.Notify(9844); return len(r) }
func (r BackupFileDescriptors) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(9845)
	r[i], r[j] = r[j], r[i]
}
func (r BackupFileDescriptors) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(9846)
	if cmp := bytes.Compare(r[i].Span.Key, r[j].Span.Key); cmp != 0 {
		__antithesis_instrumentation__.Notify(9848)
		return cmp < 0
	} else {
		__antithesis_instrumentation__.Notify(9849)
	}
	__antithesis_instrumentation__.Notify(9847)
	return bytes.Compare(r[i].Span.EndKey, r[j].Span.EndKey) < 0
}

func (m *BackupManifest) isIncremental() bool {
	__antithesis_instrumentation__.Notify(9850)
	return !m.StartTime.IsEmpty()
}

func ReadBackupManifestFromURI(
	ctx context.Context,
	mem *mon.BoundAccount,
	uri string,
	user security.SQLUsername,
	makeExternalStorageFromURI cloud.ExternalStorageFromURIFactory,
	encryption *jobspb.BackupEncryptionOptions,
) (BackupManifest, int64, error) {
	__antithesis_instrumentation__.Notify(9851)
	exportStore, err := makeExternalStorageFromURI(ctx, uri, user)

	if err != nil {
		__antithesis_instrumentation__.Notify(9853)
		return BackupManifest{}, 0, err
	} else {
		__antithesis_instrumentation__.Notify(9854)
	}
	__antithesis_instrumentation__.Notify(9852)
	defer exportStore.Close()
	return ReadBackupManifestFromStore(ctx, mem, exportStore, encryption)
}

func ReadBackupManifestFromStore(
	ctx context.Context,
	mem *mon.BoundAccount,
	exportStore cloud.ExternalStorage,
	encryption *jobspb.BackupEncryptionOptions,
) (BackupManifest, int64, error) {
	__antithesis_instrumentation__.Notify(9855)
	backupManifest, memSize, err := readBackupManifest(ctx, mem, exportStore, backupManifestName,
		encryption)
	if err != nil {
		__antithesis_instrumentation__.Notify(9857)
		oldManifest, newMemSize, newErr := readBackupManifest(ctx, mem, exportStore, backupOldManifestName,
			encryption)
		if newErr != nil {
			__antithesis_instrumentation__.Notify(9859)
			return BackupManifest{}, 0, err
		} else {
			__antithesis_instrumentation__.Notify(9860)
		}
		__antithesis_instrumentation__.Notify(9858)
		backupManifest = oldManifest
		memSize = newMemSize
	} else {
		__antithesis_instrumentation__.Notify(9861)
	}
	__antithesis_instrumentation__.Notify(9856)
	backupManifest.Dir = exportStore.Conf()

	return backupManifest, memSize, nil
}

func containsManifest(ctx context.Context, exportStore cloud.ExternalStorage) (bool, error) {
	__antithesis_instrumentation__.Notify(9862)
	r, err := exportStore.ReadFile(ctx, backupManifestName)
	if err != nil {
		__antithesis_instrumentation__.Notify(9864)
		if errors.Is(err, cloud.ErrFileDoesNotExist) {
			__antithesis_instrumentation__.Notify(9866)
			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(9867)
		}
		__antithesis_instrumentation__.Notify(9865)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(9868)
	}
	__antithesis_instrumentation__.Notify(9863)
	r.Close(ctx)
	return true, nil
}

func compressData(descBuf []byte) ([]byte, error) {
	__antithesis_instrumentation__.Notify(9869)
	gzipBuf := bytes.NewBuffer([]byte{})
	gz := gzip.NewWriter(gzipBuf)
	if _, err := gz.Write(descBuf); err != nil {
		__antithesis_instrumentation__.Notify(9872)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(9873)
	}
	__antithesis_instrumentation__.Notify(9870)
	if err := gz.Close(); err != nil {
		__antithesis_instrumentation__.Notify(9874)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(9875)
	}
	__antithesis_instrumentation__.Notify(9871)
	return gzipBuf.Bytes(), nil
}

func decompressData(ctx context.Context, mem *mon.BoundAccount, descBytes []byte) ([]byte, error) {
	__antithesis_instrumentation__.Notify(9876)
	r, err := gzip.NewReader(bytes.NewBuffer(descBytes))
	if err != nil {
		__antithesis_instrumentation__.Notify(9878)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(9879)
	}
	__antithesis_instrumentation__.Notify(9877)
	defer r.Close()
	return mon.ReadAll(ctx, ioctx.ReaderAdapter(r), mem)
}

func readBackupCheckpointManifest(
	ctx context.Context,
	mem *mon.BoundAccount,
	exportStore cloud.ExternalStorage,
	filename string,
	encryption *jobspb.BackupEncryptionOptions,
) (BackupManifest, int64, error) {
	__antithesis_instrumentation__.Notify(9880)
	checkpointFile, err := readLatestCheckpointFile(ctx, exportStore, filename)
	if err != nil {
		__antithesis_instrumentation__.Notify(9883)
		return BackupManifest{}, 0, err
	} else {
		__antithesis_instrumentation__.Notify(9884)
	}
	__antithesis_instrumentation__.Notify(9881)
	defer checkpointFile.Close(ctx)

	checksumFile, err := readLatestCheckpointFile(ctx, exportStore, filename+backupManifestChecksumSuffix)
	if err != nil {
		__antithesis_instrumentation__.Notify(9885)
		if !errors.Is(err, cloud.ErrFileDoesNotExist) {
			__antithesis_instrumentation__.Notify(9887)
			return BackupManifest{}, 0, err
		} else {
			__antithesis_instrumentation__.Notify(9888)
		}
		__antithesis_instrumentation__.Notify(9886)

		return readManifest(ctx, mem, exportStore, encryption, checkpointFile, nil)
	} else {
		__antithesis_instrumentation__.Notify(9889)
	}
	__antithesis_instrumentation__.Notify(9882)
	defer checksumFile.Close(ctx)
	return readManifest(ctx, mem, exportStore, encryption, checkpointFile, checksumFile)
}

func readBackupManifest(
	ctx context.Context,
	mem *mon.BoundAccount,
	exportStore cloud.ExternalStorage,
	filename string,
	encryption *jobspb.BackupEncryptionOptions,
) (BackupManifest, int64, error) {
	__antithesis_instrumentation__.Notify(9890)
	manifestFile, err := exportStore.ReadFile(ctx, filename)
	if err != nil {
		__antithesis_instrumentation__.Notify(9893)
		return BackupManifest{}, 0, err
	} else {
		__antithesis_instrumentation__.Notify(9894)
	}
	__antithesis_instrumentation__.Notify(9891)
	defer manifestFile.Close(ctx)

	checksumFile, err := exportStore.ReadFile(ctx, filename+backupManifestChecksumSuffix)
	if err != nil {
		__antithesis_instrumentation__.Notify(9895)
		if !errors.Is(err, cloud.ErrFileDoesNotExist) {
			__antithesis_instrumentation__.Notify(9897)
			return BackupManifest{}, 0, err
		} else {
			__antithesis_instrumentation__.Notify(9898)
		}
		__antithesis_instrumentation__.Notify(9896)

		return readManifest(ctx, mem, exportStore, encryption, manifestFile, nil)
	} else {
		__antithesis_instrumentation__.Notify(9899)
	}
	__antithesis_instrumentation__.Notify(9892)
	defer checksumFile.Close(ctx)
	return readManifest(ctx, mem, exportStore, encryption, manifestFile, checksumFile)
}

func readManifest(
	ctx context.Context,
	mem *mon.BoundAccount,
	exportStore cloud.ExternalStorage,
	encryption *jobspb.BackupEncryptionOptions,
	manifestReader ioctx.ReadCloserCtx,
	checksumReader ioctx.ReadCloserCtx,
) (BackupManifest, int64, error) {
	__antithesis_instrumentation__.Notify(9900)
	descBytes, err := mon.ReadAll(ctx, manifestReader, mem)
	if err != nil {
		__antithesis_instrumentation__.Notify(9909)
		return BackupManifest{}, 0, err
	} else {
		__antithesis_instrumentation__.Notify(9910)
	}
	__antithesis_instrumentation__.Notify(9901)
	defer func() {
		__antithesis_instrumentation__.Notify(9911)
		mem.Shrink(ctx, int64(cap(descBytes)))
	}()
	__antithesis_instrumentation__.Notify(9902)
	if checksumReader != nil {
		__antithesis_instrumentation__.Notify(9912)

		checksumFileData, err := ioctx.ReadAll(ctx, checksumReader)
		if err != nil {
			__antithesis_instrumentation__.Notify(9915)
			return BackupManifest{}, 0, errors.Wrap(err, "reading checksum file")
		} else {
			__antithesis_instrumentation__.Notify(9916)
		}
		__antithesis_instrumentation__.Notify(9913)
		checksum, err := getChecksum(descBytes)
		if err != nil {
			__antithesis_instrumentation__.Notify(9917)
			return BackupManifest{}, 0, errors.Wrap(err, "calculating checksum of manifest")
		} else {
			__antithesis_instrumentation__.Notify(9918)
		}
		__antithesis_instrumentation__.Notify(9914)
		if !bytes.Equal(checksumFileData, checksum) {
			__antithesis_instrumentation__.Notify(9919)
			return BackupManifest{}, 0, errors.Newf("checksum mismatch; expected %s, got %s",
				hex.EncodeToString(checksumFileData), hex.EncodeToString(checksum))
		} else {
			__antithesis_instrumentation__.Notify(9920)
		}
	} else {
		__antithesis_instrumentation__.Notify(9921)
	}
	__antithesis_instrumentation__.Notify(9903)

	var encryptionKey []byte
	if encryption != nil {
		__antithesis_instrumentation__.Notify(9922)
		encryptionKey, err = getEncryptionKey(ctx, encryption, exportStore.Settings(),
			exportStore.ExternalIOConf())
		if err != nil {
			__antithesis_instrumentation__.Notify(9925)
			return BackupManifest{}, 0, err
		} else {
			__antithesis_instrumentation__.Notify(9926)
		}
		__antithesis_instrumentation__.Notify(9923)
		plaintextBytes, err := storageccl.DecryptFile(ctx, descBytes, encryptionKey, mem)
		if err != nil {
			__antithesis_instrumentation__.Notify(9927)
			return BackupManifest{}, 0, err
		} else {
			__antithesis_instrumentation__.Notify(9928)
		}
		__antithesis_instrumentation__.Notify(9924)
		mem.Shrink(ctx, int64(cap(descBytes)))
		descBytes = plaintextBytes
	} else {
		__antithesis_instrumentation__.Notify(9929)
	}
	__antithesis_instrumentation__.Notify(9904)

	if isGZipped(descBytes) {
		__antithesis_instrumentation__.Notify(9930)
		decompressedBytes, err := decompressData(ctx, mem, descBytes)
		if err != nil {
			__antithesis_instrumentation__.Notify(9932)
			return BackupManifest{}, 0, errors.Wrap(
				err, "decompressing backup manifest")
		} else {
			__antithesis_instrumentation__.Notify(9933)
		}
		__antithesis_instrumentation__.Notify(9931)

		mem.Shrink(ctx, int64(cap(descBytes)))
		descBytes = decompressedBytes
	} else {
		__antithesis_instrumentation__.Notify(9934)
	}
	__antithesis_instrumentation__.Notify(9905)

	approxMemSize := int64(len(descBytes))
	if err := mem.Grow(ctx, approxMemSize); err != nil {
		__antithesis_instrumentation__.Notify(9935)
		return BackupManifest{}, 0, err
	} else {
		__antithesis_instrumentation__.Notify(9936)
	}
	__antithesis_instrumentation__.Notify(9906)

	var backupManifest BackupManifest
	if err := protoutil.Unmarshal(descBytes, &backupManifest); err != nil {
		__antithesis_instrumentation__.Notify(9937)
		mem.Shrink(ctx, approxMemSize)
		if encryption == nil && func() bool {
			__antithesis_instrumentation__.Notify(9939)
			return storageccl.AppearsEncrypted(descBytes) == true
		}() == true {
			__antithesis_instrumentation__.Notify(9940)
			return BackupManifest{}, 0, errors.Wrapf(
				err, "file appears encrypted -- try specifying one of \"%s\" or \"%s\"",
				backupOptEncPassphrase, backupOptEncKMS)
		} else {
			__antithesis_instrumentation__.Notify(9941)
		}
		__antithesis_instrumentation__.Notify(9938)
		return BackupManifest{}, 0, err
	} else {
		__antithesis_instrumentation__.Notify(9942)
	}
	__antithesis_instrumentation__.Notify(9907)
	for _, d := range backupManifest.Descriptors {
		__antithesis_instrumentation__.Notify(9943)

		if t := d.GetTable(); t == nil {
			__antithesis_instrumentation__.Notify(9944)
			continue
		} else {
			__antithesis_instrumentation__.Notify(9945)
			if t.Version == 1 && func() bool {
				__antithesis_instrumentation__.Notify(9946)
				return t.ModificationTime.IsEmpty() == true
			}() == true {
				__antithesis_instrumentation__.Notify(9947)
				t.ModificationTime = hlc.Timestamp{WallTime: 1}
			} else {
				__antithesis_instrumentation__.Notify(9948)
			}
		}
	}
	__antithesis_instrumentation__.Notify(9908)
	return backupManifest, approxMemSize, nil
}

func readBackupPartitionDescriptor(
	ctx context.Context,
	mem *mon.BoundAccount,
	exportStore cloud.ExternalStorage,
	filename string,
	encryption *jobspb.BackupEncryptionOptions,
) (BackupPartitionDescriptor, int64, error) {
	__antithesis_instrumentation__.Notify(9949)
	r, err := exportStore.ReadFile(ctx, filename)
	if err != nil {
		__antithesis_instrumentation__.Notify(9957)
		return BackupPartitionDescriptor{}, 0, err
	} else {
		__antithesis_instrumentation__.Notify(9958)
	}
	__antithesis_instrumentation__.Notify(9950)
	defer r.Close(ctx)
	descBytes, err := mon.ReadAll(ctx, r, mem)
	if err != nil {
		__antithesis_instrumentation__.Notify(9959)
		return BackupPartitionDescriptor{}, 0, err
	} else {
		__antithesis_instrumentation__.Notify(9960)
	}
	__antithesis_instrumentation__.Notify(9951)
	defer func() {
		__antithesis_instrumentation__.Notify(9961)
		mem.Shrink(ctx, int64(cap(descBytes)))
	}()
	__antithesis_instrumentation__.Notify(9952)

	if encryption != nil {
		__antithesis_instrumentation__.Notify(9962)
		encryptionKey, err := getEncryptionKey(ctx, encryption, exportStore.Settings(),
			exportStore.ExternalIOConf())
		if err != nil {
			__antithesis_instrumentation__.Notify(9965)
			return BackupPartitionDescriptor{}, 0, err
		} else {
			__antithesis_instrumentation__.Notify(9966)
		}
		__antithesis_instrumentation__.Notify(9963)
		plaintextData, err := storageccl.DecryptFile(ctx, descBytes, encryptionKey, mem)
		if err != nil {
			__antithesis_instrumentation__.Notify(9967)
			return BackupPartitionDescriptor{}, 0, err
		} else {
			__antithesis_instrumentation__.Notify(9968)
		}
		__antithesis_instrumentation__.Notify(9964)
		mem.Shrink(ctx, int64(cap(descBytes)))
		descBytes = plaintextData
	} else {
		__antithesis_instrumentation__.Notify(9969)
	}
	__antithesis_instrumentation__.Notify(9953)

	if isGZipped(descBytes) {
		__antithesis_instrumentation__.Notify(9970)
		decompressedData, err := decompressData(ctx, mem, descBytes)
		if err != nil {
			__antithesis_instrumentation__.Notify(9972)
			return BackupPartitionDescriptor{}, 0, errors.Wrap(
				err, "decompressing backup partition descriptor")
		} else {
			__antithesis_instrumentation__.Notify(9973)
		}
		__antithesis_instrumentation__.Notify(9971)
		mem.Shrink(ctx, int64(cap(descBytes)))
		descBytes = decompressedData
	} else {
		__antithesis_instrumentation__.Notify(9974)
	}
	__antithesis_instrumentation__.Notify(9954)

	memSize := int64(len(descBytes))

	if err := mem.Grow(ctx, memSize); err != nil {
		__antithesis_instrumentation__.Notify(9975)
		return BackupPartitionDescriptor{}, 0, err
	} else {
		__antithesis_instrumentation__.Notify(9976)
	}
	__antithesis_instrumentation__.Notify(9955)

	var backupManifest BackupPartitionDescriptor
	if err := protoutil.Unmarshal(descBytes, &backupManifest); err != nil {
		__antithesis_instrumentation__.Notify(9977)
		mem.Shrink(ctx, memSize)
		return BackupPartitionDescriptor{}, 0, err
	} else {
		__antithesis_instrumentation__.Notify(9978)
	}
	__antithesis_instrumentation__.Notify(9956)

	return backupManifest, memSize, err
}

func readTableStatistics(
	ctx context.Context,
	exportStore cloud.ExternalStorage,
	filename string,
	encryption *jobspb.BackupEncryptionOptions,
) (*StatsTable, error) {
	__antithesis_instrumentation__.Notify(9979)
	r, err := exportStore.ReadFile(ctx, filename)
	if err != nil {
		__antithesis_instrumentation__.Notify(9984)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(9985)
	}
	__antithesis_instrumentation__.Notify(9980)
	defer r.Close(ctx)
	statsBytes, err := ioctx.ReadAll(ctx, r)
	if err != nil {
		__antithesis_instrumentation__.Notify(9986)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(9987)
	}
	__antithesis_instrumentation__.Notify(9981)
	if encryption != nil {
		__antithesis_instrumentation__.Notify(9988)
		encryptionKey, err := getEncryptionKey(ctx, encryption, exportStore.Settings(),
			exportStore.ExternalIOConf())
		if err != nil {
			__antithesis_instrumentation__.Notify(9990)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(9991)
		}
		__antithesis_instrumentation__.Notify(9989)
		statsBytes, err = storageccl.DecryptFile(ctx, statsBytes, encryptionKey, nil)
		if err != nil {
			__antithesis_instrumentation__.Notify(9992)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(9993)
		}
	} else {
		__antithesis_instrumentation__.Notify(9994)
	}
	__antithesis_instrumentation__.Notify(9982)
	var tableStats StatsTable
	if err := protoutil.Unmarshal(statsBytes, &tableStats); err != nil {
		__antithesis_instrumentation__.Notify(9995)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(9996)
	}
	__antithesis_instrumentation__.Notify(9983)
	return &tableStats, err
}

func writeBackupManifest(
	ctx context.Context,
	settings *cluster.Settings,
	exportStore cloud.ExternalStorage,
	filename string,
	encryption *jobspb.BackupEncryptionOptions,
	desc *BackupManifest,
) error {
	__antithesis_instrumentation__.Notify(9997)
	sort.Sort(BackupFileDescriptors(desc.Files))

	descBuf, err := protoutil.Marshal(desc)
	if err != nil {
		__antithesis_instrumentation__.Notify(10004)
		return err
	} else {
		__antithesis_instrumentation__.Notify(10005)
	}
	__antithesis_instrumentation__.Notify(9998)

	descBuf, err = compressData(descBuf)
	if err != nil {
		__antithesis_instrumentation__.Notify(10006)
		return errors.Wrap(err, "compressing backup manifest")
	} else {
		__antithesis_instrumentation__.Notify(10007)
	}
	__antithesis_instrumentation__.Notify(9999)

	if encryption != nil {
		__antithesis_instrumentation__.Notify(10008)
		encryptionKey, err := getEncryptionKey(ctx, encryption, settings, exportStore.ExternalIOConf())
		if err != nil {
			__antithesis_instrumentation__.Notify(10010)
			return err
		} else {
			__antithesis_instrumentation__.Notify(10011)
		}
		__antithesis_instrumentation__.Notify(10009)
		descBuf, err = storageccl.EncryptFile(descBuf, encryptionKey)
		if err != nil {
			__antithesis_instrumentation__.Notify(10012)
			return err
		} else {
			__antithesis_instrumentation__.Notify(10013)
		}
	} else {
		__antithesis_instrumentation__.Notify(10014)
	}
	__antithesis_instrumentation__.Notify(10000)

	if err := cloud.WriteFile(ctx, exportStore, filename, bytes.NewReader(descBuf)); err != nil {
		__antithesis_instrumentation__.Notify(10015)
		return err
	} else {
		__antithesis_instrumentation__.Notify(10016)
	}
	__antithesis_instrumentation__.Notify(10001)

	checksum, err := getChecksum(descBuf)
	if err != nil {
		__antithesis_instrumentation__.Notify(10017)
		return errors.Wrap(err, "calculating checksum")
	} else {
		__antithesis_instrumentation__.Notify(10018)
	}
	__antithesis_instrumentation__.Notify(10002)

	if err := cloud.WriteFile(ctx, exportStore, filename+backupManifestChecksumSuffix, bytes.NewReader(checksum)); err != nil {
		__antithesis_instrumentation__.Notify(10019)
		return errors.Wrap(err, "writing manifest checksum")
	} else {
		__antithesis_instrumentation__.Notify(10020)
	}
	__antithesis_instrumentation__.Notify(10003)

	return nil
}

func getChecksum(data []byte) ([]byte, error) {
	__antithesis_instrumentation__.Notify(10021)
	const checksumSizeBytes = 4
	hash := sha256.New()
	if _, err := hash.Write(data); err != nil {
		__antithesis_instrumentation__.Notify(10023)
		return nil, errors.Wrap(err,
			`"It never returns an error." -- https://golang.org/pkg/hash`)
	} else {
		__antithesis_instrumentation__.Notify(10024)
	}
	__antithesis_instrumentation__.Notify(10022)
	return hash.Sum(nil)[:checksumSizeBytes], nil
}

func getEncryptionKey(
	ctx context.Context,
	encryption *jobspb.BackupEncryptionOptions,
	settings *cluster.Settings,
	ioConf base.ExternalIODirConfig,
) ([]byte, error) {
	__antithesis_instrumentation__.Notify(10025)
	if encryption == nil {
		__antithesis_instrumentation__.Notify(10028)
		return nil, errors.New("FileEncryptionOptions is nil when retrieving encryption key")
	} else {
		__antithesis_instrumentation__.Notify(10029)
	}
	__antithesis_instrumentation__.Notify(10026)
	switch encryption.Mode {
	case jobspb.EncryptionMode_Passphrase:
		__antithesis_instrumentation__.Notify(10030)
		return encryption.Key, nil
	case jobspb.EncryptionMode_KMS:
		__antithesis_instrumentation__.Notify(10031)

		kms, err := cloud.KMSFromURI(encryption.KMSInfo.Uri, &backupKMSEnv{
			settings: settings,
			conf:     &ioConf,
		})
		if err != nil {
			__antithesis_instrumentation__.Notify(10036)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(10037)
		}
		__antithesis_instrumentation__.Notify(10032)

		defer func() {
			__antithesis_instrumentation__.Notify(10038)
			_ = kms.Close()
		}()
		__antithesis_instrumentation__.Notify(10033)

		plaintextDataKey, err := kms.Decrypt(ctx, encryption.KMSInfo.EncryptedDataKey)
		if err != nil {
			__antithesis_instrumentation__.Notify(10039)
			return nil, errors.Wrap(err, "failed to decrypt data key")
		} else {
			__antithesis_instrumentation__.Notify(10040)
		}
		__antithesis_instrumentation__.Notify(10034)

		return plaintextDataKey, nil
	default:
		__antithesis_instrumentation__.Notify(10035)
	}
	__antithesis_instrumentation__.Notify(10027)

	return nil, errors.New("invalid encryption mode")
}

func writeBackupPartitionDescriptor(
	ctx context.Context,
	exportStore cloud.ExternalStorage,
	filename string,
	encryption *jobspb.BackupEncryptionOptions,
	desc *BackupPartitionDescriptor,
) error {
	__antithesis_instrumentation__.Notify(10041)
	descBuf, err := protoutil.Marshal(desc)
	if err != nil {
		__antithesis_instrumentation__.Notify(10045)
		return err
	} else {
		__antithesis_instrumentation__.Notify(10046)
	}
	__antithesis_instrumentation__.Notify(10042)
	descBuf, err = compressData(descBuf)
	if err != nil {
		__antithesis_instrumentation__.Notify(10047)
		return errors.Wrap(err, "compressing backup partition descriptor")
	} else {
		__antithesis_instrumentation__.Notify(10048)
	}
	__antithesis_instrumentation__.Notify(10043)
	if encryption != nil {
		__antithesis_instrumentation__.Notify(10049)
		encryptionKey, err := getEncryptionKey(ctx, encryption, exportStore.Settings(),
			exportStore.ExternalIOConf())
		if err != nil {
			__antithesis_instrumentation__.Notify(10051)
			return err
		} else {
			__antithesis_instrumentation__.Notify(10052)
		}
		__antithesis_instrumentation__.Notify(10050)
		descBuf, err = storageccl.EncryptFile(descBuf, encryptionKey)
		if err != nil {
			__antithesis_instrumentation__.Notify(10053)
			return err
		} else {
			__antithesis_instrumentation__.Notify(10054)
		}
	} else {
		__antithesis_instrumentation__.Notify(10055)
	}
	__antithesis_instrumentation__.Notify(10044)

	return cloud.WriteFile(ctx, exportStore, filename, bytes.NewReader(descBuf))
}

func writeTableStatistics(
	ctx context.Context,
	exportStore cloud.ExternalStorage,
	filename string,
	encryption *jobspb.BackupEncryptionOptions,
	stats *StatsTable,
) error {
	__antithesis_instrumentation__.Notify(10056)
	statsBuf, err := protoutil.Marshal(stats)
	if err != nil {
		__antithesis_instrumentation__.Notify(10059)
		return err
	} else {
		__antithesis_instrumentation__.Notify(10060)
	}
	__antithesis_instrumentation__.Notify(10057)
	if encryption != nil {
		__antithesis_instrumentation__.Notify(10061)
		encryptionKey, err := getEncryptionKey(ctx, encryption, exportStore.Settings(),
			exportStore.ExternalIOConf())
		if err != nil {
			__antithesis_instrumentation__.Notify(10063)
			return err
		} else {
			__antithesis_instrumentation__.Notify(10064)
		}
		__antithesis_instrumentation__.Notify(10062)
		statsBuf, err = storageccl.EncryptFile(statsBuf, encryptionKey)
		if err != nil {
			__antithesis_instrumentation__.Notify(10065)
			return err
		} else {
			__antithesis_instrumentation__.Notify(10066)
		}
	} else {
		__antithesis_instrumentation__.Notify(10067)
	}
	__antithesis_instrumentation__.Notify(10058)
	return cloud.WriteFile(ctx, exportStore, filename, bytes.NewReader(statsBuf))
}

func loadBackupManifests(
	ctx context.Context,
	mem *mon.BoundAccount,
	uris []string,
	user security.SQLUsername,
	makeExternalStorageFromURI cloud.ExternalStorageFromURIFactory,
	encryption *jobspb.BackupEncryptionOptions,
) ([]BackupManifest, int64, error) {
	__antithesis_instrumentation__.Notify(10068)
	backupManifests := make([]BackupManifest, len(uris))
	var reserved int64
	defer func() {
		__antithesis_instrumentation__.Notify(10072)
		if reserved != 0 {
			__antithesis_instrumentation__.Notify(10073)
			mem.Shrink(ctx, reserved)
		} else {
			__antithesis_instrumentation__.Notify(10074)
		}
	}()
	__antithesis_instrumentation__.Notify(10069)
	for i, uri := range uris {
		__antithesis_instrumentation__.Notify(10075)
		desc, memSize, err := ReadBackupManifestFromURI(ctx, mem, uri, user, makeExternalStorageFromURI,
			encryption)
		if err != nil {
			__antithesis_instrumentation__.Notify(10077)
			return nil, 0, errors.Wrapf(err, "failed to read backup descriptor")
		} else {
			__antithesis_instrumentation__.Notify(10078)
		}
		__antithesis_instrumentation__.Notify(10076)
		reserved += memSize
		backupManifests[i] = desc
	}
	__antithesis_instrumentation__.Notify(10070)
	if len(backupManifests) == 0 {
		__antithesis_instrumentation__.Notify(10079)
		return nil, 0, errors.Newf("no backups found")
	} else {
		__antithesis_instrumentation__.Notify(10080)
	}
	__antithesis_instrumentation__.Notify(10071)
	memSize := reserved
	reserved = 0

	return backupManifests, memSize, nil
}

func getLocalityInfo(
	ctx context.Context,
	stores []cloud.ExternalStorage,
	uris []string,
	mainBackupManifest BackupManifest,
	encryption *jobspb.BackupEncryptionOptions,
	prefix string,
) (jobspb.RestoreDetails_BackupLocalityInfo, error) {
	__antithesis_instrumentation__.Notify(10081)
	var info jobspb.RestoreDetails_BackupLocalityInfo

	urisByOrigLocality := make(map[string]string)
	for _, filename := range mainBackupManifest.PartitionDescriptorFilenames {
		__antithesis_instrumentation__.Notify(10083)
		if prefix != "" {
			__antithesis_instrumentation__.Notify(10086)
			filename = path.Join(prefix, filename)
		} else {
			__antithesis_instrumentation__.Notify(10087)
		}
		__antithesis_instrumentation__.Notify(10084)
		found := false
		for i, store := range stores {
			__antithesis_instrumentation__.Notify(10088)
			if desc, _, err := readBackupPartitionDescriptor(ctx, nil, store, filename, encryption); err == nil {
				__antithesis_instrumentation__.Notify(10089)
				if desc.BackupID != mainBackupManifest.ID {
					__antithesis_instrumentation__.Notify(10093)
					return info, errors.Errorf(
						"expected backup part to have backup ID %s, found %s",
						mainBackupManifest.ID, desc.BackupID,
					)
				} else {
					__antithesis_instrumentation__.Notify(10094)
				}
				__antithesis_instrumentation__.Notify(10090)
				origLocalityKV := desc.LocalityKV
				kv := roachpb.Tier{}
				if err := kv.FromString(origLocalityKV); err != nil {
					__antithesis_instrumentation__.Notify(10095)
					return info, errors.Wrapf(err, "reading backup manifest from %s",
						RedactURIForErrorMessage(uris[i]))
				} else {
					__antithesis_instrumentation__.Notify(10096)
				}
				__antithesis_instrumentation__.Notify(10091)
				if _, ok := urisByOrigLocality[origLocalityKV]; ok {
					__antithesis_instrumentation__.Notify(10097)
					return info, errors.Errorf("duplicate locality %s found in backup", origLocalityKV)
				} else {
					__antithesis_instrumentation__.Notify(10098)
				}
				__antithesis_instrumentation__.Notify(10092)
				urisByOrigLocality[origLocalityKV] = uris[i]
				found = true
				break
			} else {
				__antithesis_instrumentation__.Notify(10099)
			}
		}
		__antithesis_instrumentation__.Notify(10085)
		if !found {
			__antithesis_instrumentation__.Notify(10100)
			return info, errors.Errorf("expected manifest %s not found in backup locations", filename)
		} else {
			__antithesis_instrumentation__.Notify(10101)
		}
	}
	__antithesis_instrumentation__.Notify(10082)
	info.URIsByOriginalLocalityKV = urisByOrigLocality
	return info, nil
}

const (
	IncludeManifest = true

	OmitManifest = false
)

func checkForLatestFileInCollection(
	ctx context.Context, store cloud.ExternalStorage,
) (bool, error) {
	__antithesis_instrumentation__.Notify(10102)
	r, err := findLatestFile(ctx, store)
	if err != nil {
		__antithesis_instrumentation__.Notify(10105)
		if !errors.Is(err, cloud.ErrFileDoesNotExist) {
			__antithesis_instrumentation__.Notify(10107)
			return false, pgerror.WithCandidateCode(err, pgcode.Io)
		} else {
			__antithesis_instrumentation__.Notify(10108)
		}
		__antithesis_instrumentation__.Notify(10106)

		r, err = store.ReadFile(ctx, latestFileName)
	} else {
		__antithesis_instrumentation__.Notify(10109)
	}
	__antithesis_instrumentation__.Notify(10103)
	if err != nil {
		__antithesis_instrumentation__.Notify(10110)
		if errors.Is(err, cloud.ErrFileDoesNotExist) {
			__antithesis_instrumentation__.Notify(10112)
			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(10113)
		}
		__antithesis_instrumentation__.Notify(10111)
		return false, pgerror.WithCandidateCode(err, pgcode.Io)
	} else {
		__antithesis_instrumentation__.Notify(10114)
	}
	__antithesis_instrumentation__.Notify(10104)
	r.Close(ctx)
	return true, nil
}

func resolveBackupManifestsExplicitIncrementals(
	ctx context.Context,
	mem *mon.BoundAccount,
	mkStore cloud.ExternalStorageFromURIFactory,
	from [][]string,
	endTime hlc.Timestamp,
	encryption *jobspb.BackupEncryptionOptions,
	user security.SQLUsername,
) (
	defaultURIs []string,

	mainBackupManifests []BackupManifest,
	localityInfo []jobspb.RestoreDetails_BackupLocalityInfo,
	reservedMemSize int64,
	_ error,
) {
	__antithesis_instrumentation__.Notify(10115)

	var ownedMemSize int64
	defer func() {
		__antithesis_instrumentation__.Notify(10119)
		if ownedMemSize != 0 {
			__antithesis_instrumentation__.Notify(10120)
			mem.Shrink(ctx, ownedMemSize)
		} else {
			__antithesis_instrumentation__.Notify(10121)
		}
	}()
	__antithesis_instrumentation__.Notify(10116)

	defaultURIs = make([]string, len(from))
	localityInfo = make([]jobspb.RestoreDetails_BackupLocalityInfo, len(from))
	mainBackupManifests = make([]BackupManifest, len(from))

	var err error
	for i, uris := range from {
		__antithesis_instrumentation__.Notify(10122)

		defaultURIs[i] = uris[0]

		stores := make([]cloud.ExternalStorage, len(uris))
		for j := range uris {
			__antithesis_instrumentation__.Notify(10125)
			stores[j], err = mkStore(ctx, uris[j], user)
			if err != nil {
				__antithesis_instrumentation__.Notify(10127)
				return nil, nil, nil, 0, errors.Wrapf(err, "export configuration")
			} else {
				__antithesis_instrumentation__.Notify(10128)
			}
			__antithesis_instrumentation__.Notify(10126)
			defer stores[j].Close()
		}
		__antithesis_instrumentation__.Notify(10123)

		var memSize int64
		mainBackupManifests[i], memSize, err = ReadBackupManifestFromStore(ctx, mem, stores[0], encryption)
		if err != nil {
			__antithesis_instrumentation__.Notify(10129)
			return nil, nil, nil, 0, err
		} else {
			__antithesis_instrumentation__.Notify(10130)
		}
		__antithesis_instrumentation__.Notify(10124)
		ownedMemSize += memSize

		if len(uris) > 1 {
			__antithesis_instrumentation__.Notify(10131)
			localityInfo[i], err = getLocalityInfo(
				ctx, stores, uris, mainBackupManifests[i], encryption, "",
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(10132)
				return nil, nil, nil, 0, err
			} else {
				__antithesis_instrumentation__.Notify(10133)
			}
		} else {
			__antithesis_instrumentation__.Notify(10134)
		}
	}
	__antithesis_instrumentation__.Notify(10117)

	totalMemSize := ownedMemSize
	ownedMemSize = 0

	validatedDefaultURIs, validatedMainBackupManifests, validatedLocalityInfo, err := validateEndTimeAndTruncate(
		defaultURIs, mainBackupManifests, localityInfo, endTime)

	if err != nil {
		__antithesis_instrumentation__.Notify(10135)
		return nil, nil, nil, 0, err
	} else {
		__antithesis_instrumentation__.Notify(10136)
	}
	__antithesis_instrumentation__.Notify(10118)
	return validatedDefaultURIs, validatedMainBackupManifests, validatedLocalityInfo, totalMemSize, nil
}

func resolveBackupManifests(
	ctx context.Context,
	mem *mon.BoundAccount,
	baseStores []cloud.ExternalStorage,
	mkStore cloud.ExternalStorageFromURIFactory,
	fullyResolvedBaseDirectory []string,
	fullyResolvedIncrementalsDirectory []string,
	endTime hlc.Timestamp,
	encryption *jobspb.BackupEncryptionOptions,
	user security.SQLUsername,
) (
	defaultURIs []string,

	mainBackupManifests []BackupManifest,
	localityInfo []jobspb.RestoreDetails_BackupLocalityInfo,
	reservedMemSize int64,
	_ error,
) {
	__antithesis_instrumentation__.Notify(10137)
	var ownedMemSize int64
	defer func() {
		__antithesis_instrumentation__.Notify(10145)
		if ownedMemSize != 0 {
			__antithesis_instrumentation__.Notify(10146)
			mem.Shrink(ctx, ownedMemSize)
		} else {
			__antithesis_instrumentation__.Notify(10147)
		}
	}()
	__antithesis_instrumentation__.Notify(10138)
	baseManifest, memSize, err := ReadBackupManifestFromStore(ctx, mem, baseStores[0], encryption)
	if err != nil {
		__antithesis_instrumentation__.Notify(10148)
		return nil, nil, nil, 0, err
	} else {
		__antithesis_instrumentation__.Notify(10149)
	}
	__antithesis_instrumentation__.Notify(10139)
	ownedMemSize += memSize

	incStores := make([]cloud.ExternalStorage, len(fullyResolvedIncrementalsDirectory))
	for i := range fullyResolvedIncrementalsDirectory {
		__antithesis_instrumentation__.Notify(10150)
		store, err := mkStore(ctx, fullyResolvedIncrementalsDirectory[i], user)
		if err != nil {
			__antithesis_instrumentation__.Notify(10152)
			return nil, nil, nil, 0, errors.Wrapf(err, "failed to open backup storage location")
		} else {
			__antithesis_instrumentation__.Notify(10153)
		}
		__antithesis_instrumentation__.Notify(10151)
		defer store.Close()
		incStores[i] = store
	}
	__antithesis_instrumentation__.Notify(10140)

	var prev []string
	if len(incStores) > 0 {
		__antithesis_instrumentation__.Notify(10154)
		prev, err = FindPriorBackups(ctx, incStores[0], IncludeManifest)
		if err != nil {
			__antithesis_instrumentation__.Notify(10155)
			return nil, nil, nil, 0, err
		} else {
			__antithesis_instrumentation__.Notify(10156)
		}
	} else {
		__antithesis_instrumentation__.Notify(10157)
	}
	__antithesis_instrumentation__.Notify(10141)
	numLayers := len(prev) + 1

	defaultURIs = make([]string, numLayers)
	mainBackupManifests = make([]BackupManifest, numLayers)
	localityInfo = make([]jobspb.RestoreDetails_BackupLocalityInfo, numLayers)

	defaultURIs[0] = fullyResolvedBaseDirectory[0]
	mainBackupManifests[0] = baseManifest
	localityInfo[0], err = getLocalityInfo(
		ctx, baseStores, fullyResolvedBaseDirectory, baseManifest, encryption, "",
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(10158)
		return nil, nil, nil, 0, err
	} else {
		__antithesis_instrumentation__.Notify(10159)
	}
	__antithesis_instrumentation__.Notify(10142)

	if numLayers > 1 {
		__antithesis_instrumentation__.Notify(10160)
		numPartitions := len(fullyResolvedIncrementalsDirectory)

		baseURIs := make([]*url.URL, numPartitions)
		for i := range fullyResolvedIncrementalsDirectory {
			__antithesis_instrumentation__.Notify(10162)
			baseURIs[i], err = url.Parse(fullyResolvedIncrementalsDirectory[i])
			if err != nil {
				__antithesis_instrumentation__.Notify(10163)
				return nil, nil, nil, 0, err
			} else {
				__antithesis_instrumentation__.Notify(10164)
			}
		}
		__antithesis_instrumentation__.Notify(10161)

		for i := range prev {
			__antithesis_instrumentation__.Notify(10165)
			defaultManifestForLayer, memSize, err := readBackupManifest(ctx, mem, incStores[0], prev[i], encryption)
			if err != nil {
				__antithesis_instrumentation__.Notify(10168)
				return nil, nil, nil, 0, err
			} else {
				__antithesis_instrumentation__.Notify(10169)
			}
			__antithesis_instrumentation__.Notify(10166)
			ownedMemSize += memSize
			mainBackupManifests[i+1] = defaultManifestForLayer

			incSubDir := path.Dir(prev[i])
			partitionURIs := make([]string, numPartitions)
			for j := range baseURIs {
				__antithesis_instrumentation__.Notify(10170)
				u := *baseURIs[j]
				u.Path = JoinURLPath(u.Path, incSubDir)
				partitionURIs[j] = u.String()
			}
			__antithesis_instrumentation__.Notify(10167)
			defaultURIs[i+1] = partitionURIs[0]
			localityInfo[i+1], err = getLocalityInfo(ctx, incStores, partitionURIs, defaultManifestForLayer, encryption, incSubDir)
			if err != nil {
				__antithesis_instrumentation__.Notify(10171)
				return nil, nil, nil, 0, err
			} else {
				__antithesis_instrumentation__.Notify(10172)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(10173)
	}
	__antithesis_instrumentation__.Notify(10143)

	totalMemSize := ownedMemSize
	ownedMemSize = 0

	validatedDefaultURIs, validatedMainBackupManifests, validatedLocalityInfo, err := validateEndTimeAndTruncate(
		defaultURIs, mainBackupManifests, localityInfo, endTime)

	if err != nil {
		__antithesis_instrumentation__.Notify(10174)
		return nil, nil, nil, 0, err
	} else {
		__antithesis_instrumentation__.Notify(10175)
	}
	__antithesis_instrumentation__.Notify(10144)
	return validatedDefaultURIs, validatedMainBackupManifests, validatedLocalityInfo, totalMemSize, nil
}

func validateEndTimeAndTruncate(
	defaultURIs []string,
	mainBackupManifests []BackupManifest,
	localityInfo []jobspb.RestoreDetails_BackupLocalityInfo,
	endTime hlc.Timestamp,
) ([]string, []BackupManifest, []jobspb.RestoreDetails_BackupLocalityInfo, error) {
	__antithesis_instrumentation__.Notify(10176)

	if endTime.IsEmpty() {
		__antithesis_instrumentation__.Notify(10179)
		return defaultURIs, mainBackupManifests, localityInfo, nil
	} else {
		__antithesis_instrumentation__.Notify(10180)
	}
	__antithesis_instrumentation__.Notify(10177)
	for i, b := range mainBackupManifests {
		__antithesis_instrumentation__.Notify(10181)

		if !(b.StartTime.Less(endTime) && func() bool {
			__antithesis_instrumentation__.Notify(10184)
			return endTime.LessEq(b.EndTime) == true
		}() == true) {
			__antithesis_instrumentation__.Notify(10185)
			continue
		} else {
			__antithesis_instrumentation__.Notify(10186)
		}
		__antithesis_instrumentation__.Notify(10182)

		if !endTime.Equal(b.EndTime) {
			__antithesis_instrumentation__.Notify(10187)
			if b.MVCCFilter != MVCCFilter_All {
				__antithesis_instrumentation__.Notify(10189)
				const errPrefix = "invalid RESTORE timestamp: restoring to arbitrary time requires that BACKUP for requested time be created with '%s' option."
				if i == 0 {
					__antithesis_instrumentation__.Notify(10191)
					return nil, nil, nil, errors.Errorf(
						errPrefix+" nearest backup time is %s", backupOptRevisionHistory,
						timeutil.Unix(0, b.EndTime.WallTime).UTC(),
					)
				} else {
					__antithesis_instrumentation__.Notify(10192)
				}
				__antithesis_instrumentation__.Notify(10190)
				return nil, nil, nil, errors.Errorf(
					errPrefix+" nearest BACKUP times are %s or %s",
					backupOptRevisionHistory,
					timeutil.Unix(0, mainBackupManifests[i-1].EndTime.WallTime).UTC(),
					timeutil.Unix(0, b.EndTime.WallTime).UTC(),
				)
			} else {
				__antithesis_instrumentation__.Notify(10193)
			}
			__antithesis_instrumentation__.Notify(10188)

			if endTime.LessEq(b.RevisionStartTime) {
				__antithesis_instrumentation__.Notify(10194)
				return nil, nil, nil, errors.Errorf(
					"invalid RESTORE timestamp: BACKUP for requested time only has revision history"+
						" from %v", timeutil.Unix(0, b.RevisionStartTime.WallTime).UTC(),
				)
			} else {
				__antithesis_instrumentation__.Notify(10195)
			}
		} else {
			__antithesis_instrumentation__.Notify(10196)
		}
		__antithesis_instrumentation__.Notify(10183)
		return defaultURIs[:i+1], mainBackupManifests[:i+1], localityInfo[:i+1], nil

	}
	__antithesis_instrumentation__.Notify(10178)

	return nil, nil, nil, errors.Errorf(
		"invalid RESTORE timestamp: supplied backups do not cover requested time",
	)
}

func getBackupIndexAtTime(backupManifests []BackupManifest, asOf hlc.Timestamp) (int, error) {
	__antithesis_instrumentation__.Notify(10197)
	if len(backupManifests) == 0 {
		__antithesis_instrumentation__.Notify(10201)
		return -1, errors.New("expected a nonempty backup manifest list, got an empty list")
	} else {
		__antithesis_instrumentation__.Notify(10202)
	}
	__antithesis_instrumentation__.Notify(10198)
	backupManifestIndex := len(backupManifests) - 1
	if asOf.IsEmpty() {
		__antithesis_instrumentation__.Notify(10203)
		return backupManifestIndex, nil
	} else {
		__antithesis_instrumentation__.Notify(10204)
	}
	__antithesis_instrumentation__.Notify(10199)
	for ind, b := range backupManifests {
		__antithesis_instrumentation__.Notify(10205)
		if asOf.Less(b.StartTime) {
			__antithesis_instrumentation__.Notify(10207)
			break
		} else {
			__antithesis_instrumentation__.Notify(10208)
		}
		__antithesis_instrumentation__.Notify(10206)
		backupManifestIndex = ind
	}
	__antithesis_instrumentation__.Notify(10200)
	return backupManifestIndex, nil
}

func loadSQLDescsFromBackupsAtTime(
	backupManifests []BackupManifest, asOf hlc.Timestamp,
) ([]catalog.Descriptor, BackupManifest) {
	__antithesis_instrumentation__.Notify(10209)
	lastBackupManifest := backupManifests[len(backupManifests)-1]

	unwrapDescriptors := func(raw []descpb.Descriptor) []catalog.Descriptor {
		__antithesis_instrumentation__.Notify(10216)
		ret := make([]catalog.Descriptor, 0, len(raw))
		for i := range raw {
			__antithesis_instrumentation__.Notify(10218)
			ret = append(ret, descbuilder.NewBuilder(&raw[i]).BuildExistingMutable())
		}
		__antithesis_instrumentation__.Notify(10217)
		return ret
	}
	__antithesis_instrumentation__.Notify(10210)
	if asOf.IsEmpty() {
		__antithesis_instrumentation__.Notify(10219)
		if lastBackupManifest.DescriptorCoverage != tree.AllDescriptors {
			__antithesis_instrumentation__.Notify(10221)
			return unwrapDescriptors(lastBackupManifest.Descriptors), lastBackupManifest
		} else {
			__antithesis_instrumentation__.Notify(10222)
		}
		__antithesis_instrumentation__.Notify(10220)

		asOf = lastBackupManifest.EndTime
	} else {
		__antithesis_instrumentation__.Notify(10223)
	}
	__antithesis_instrumentation__.Notify(10211)

	for _, b := range backupManifests {
		__antithesis_instrumentation__.Notify(10224)
		if asOf.Less(b.StartTime) {
			__antithesis_instrumentation__.Notify(10226)
			break
		} else {
			__antithesis_instrumentation__.Notify(10227)
		}
		__antithesis_instrumentation__.Notify(10225)
		lastBackupManifest = b
	}
	__antithesis_instrumentation__.Notify(10212)
	if len(lastBackupManifest.DescriptorChanges) == 0 {
		__antithesis_instrumentation__.Notify(10228)
		return unwrapDescriptors(lastBackupManifest.Descriptors), lastBackupManifest
	} else {
		__antithesis_instrumentation__.Notify(10229)
	}
	__antithesis_instrumentation__.Notify(10213)

	byID := make(map[descpb.ID]*descpb.Descriptor, len(lastBackupManifest.Descriptors))
	for _, rev := range lastBackupManifest.DescriptorChanges {
		__antithesis_instrumentation__.Notify(10230)
		if asOf.Less(rev.Time) {
			__antithesis_instrumentation__.Notify(10232)
			break
		} else {
			__antithesis_instrumentation__.Notify(10233)
		}
		__antithesis_instrumentation__.Notify(10231)
		if rev.Desc == nil {
			__antithesis_instrumentation__.Notify(10234)
			delete(byID, rev.ID)
		} else {
			__antithesis_instrumentation__.Notify(10235)
			byID[rev.ID] = rev.Desc
		}
	}
	__antithesis_instrumentation__.Notify(10214)

	allDescs := make([]catalog.Descriptor, 0, len(byID))
	for _, raw := range byID {
		__antithesis_instrumentation__.Notify(10236)

		desc := descbuilder.NewBuilder(raw).BuildExistingMutable()
		var isObject bool
		switch d := desc.(type) {
		case catalog.TableDescriptor:
			__antithesis_instrumentation__.Notify(10239)

			if d.GetState() == descpb.DescriptorState_DROP {
				__antithesis_instrumentation__.Notify(10242)
				continue
			} else {
				__antithesis_instrumentation__.Notify(10243)
			}
			__antithesis_instrumentation__.Notify(10240)
			isObject = true
		case catalog.TypeDescriptor, catalog.SchemaDescriptor:
			__antithesis_instrumentation__.Notify(10241)
			isObject = true
		}
		__antithesis_instrumentation__.Notify(10237)
		if isObject && func() bool {
			__antithesis_instrumentation__.Notify(10244)
			return byID[desc.GetParentID()] == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(10245)
			continue
		} else {
			__antithesis_instrumentation__.Notify(10246)
		}
		__antithesis_instrumentation__.Notify(10238)
		allDescs = append(allDescs, desc)
	}
	__antithesis_instrumentation__.Notify(10215)
	return allDescs, lastBackupManifest
}

func sanitizeLocalityKV(kv string) string {
	__antithesis_instrumentation__.Notify(10247)
	sanitizedKV := make([]byte, len(kv))
	for i := 0; i < len(kv); i++ {
		__antithesis_instrumentation__.Notify(10249)
		if (kv[i] >= 'a' && func() bool {
			__antithesis_instrumentation__.Notify(10250)
			return kv[i] <= 'z' == true
		}() == true) || func() bool {
			__antithesis_instrumentation__.Notify(10251)
			return (kv[i] >= 'A' && func() bool {
				__antithesis_instrumentation__.Notify(10252)
				return kv[i] <= 'Z' == true
			}() == true) == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(10253)
			return (kv[i] >= '0' && func() bool {
				__antithesis_instrumentation__.Notify(10254)
				return kv[i] <= '9' == true
			}() == true) == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(10255)
			return kv[i] == '-' == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(10256)
			return kv[i] == '=' == true
		}() == true {
			__antithesis_instrumentation__.Notify(10257)
			sanitizedKV[i] = kv[i]
		} else {
			__antithesis_instrumentation__.Notify(10258)
			sanitizedKV[i] = '_'
		}
	}
	__antithesis_instrumentation__.Notify(10248)
	return string(sanitizedKV)
}

func readEncryptionOptions(
	ctx context.Context, src cloud.ExternalStorage,
) ([]jobspb.EncryptionInfo, error) {
	__antithesis_instrumentation__.Notify(10259)
	const encryptionReadErrorMsg = `could not find or read encryption information`

	files, err := getEncryptionInfoFiles(ctx, src)
	if err != nil {
		__antithesis_instrumentation__.Notify(10262)
		return nil, errors.Mark(errors.Wrap(err, encryptionReadErrorMsg), errEncryptionInfoRead)
	} else {
		__antithesis_instrumentation__.Notify(10263)
	}
	__antithesis_instrumentation__.Notify(10260)
	var encInfo []jobspb.EncryptionInfo

	for i := len(files) - 1; i >= 0; i-- {
		__antithesis_instrumentation__.Notify(10264)
		r, err := src.ReadFile(ctx, files[i])
		if err != nil {
			__antithesis_instrumentation__.Notify(10268)
			return nil, errors.Wrap(err, encryptionReadErrorMsg)
		} else {
			__antithesis_instrumentation__.Notify(10269)
		}
		__antithesis_instrumentation__.Notify(10265)
		defer r.Close(ctx)

		encInfoBytes, err := ioctx.ReadAll(ctx, r)
		if err != nil {
			__antithesis_instrumentation__.Notify(10270)
			return nil, errors.Wrap(err, encryptionReadErrorMsg)
		} else {
			__antithesis_instrumentation__.Notify(10271)
		}
		__antithesis_instrumentation__.Notify(10266)
		var currentEncInfo jobspb.EncryptionInfo
		if err := protoutil.Unmarshal(encInfoBytes, &currentEncInfo); err != nil {
			__antithesis_instrumentation__.Notify(10272)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(10273)
		}
		__antithesis_instrumentation__.Notify(10267)
		encInfo = append(encInfo, currentEncInfo)
	}
	__antithesis_instrumentation__.Notify(10261)
	return encInfo, nil
}

func getEncryptionInfoFiles(ctx context.Context, dest cloud.ExternalStorage) ([]string, error) {
	__antithesis_instrumentation__.Notify(10274)
	var files []string

	err := dest.List(ctx, "", "", func(p string) error {
		__antithesis_instrumentation__.Notify(10277)
		paths := strings.Split(p, "/")
		p = paths[len(paths)-1]
		if match := strings.HasPrefix(p, backupEncryptionInfoFile); match {
			__antithesis_instrumentation__.Notify(10279)
			files = append(files, p)
		} else {
			__antithesis_instrumentation__.Notify(10280)
		}
		__antithesis_instrumentation__.Notify(10278)

		return nil
	})
	__antithesis_instrumentation__.Notify(10275)
	if len(files) < 1 {
		__antithesis_instrumentation__.Notify(10281)
		return nil, errors.New("no ENCRYPTION-INFO files found")
	} else {
		__antithesis_instrumentation__.Notify(10282)
	}
	__antithesis_instrumentation__.Notify(10276)

	return files, err
}

func writeEncryptionInfoIfNotExists(
	ctx context.Context, opts *jobspb.EncryptionInfo, dest cloud.ExternalStorage,
) error {
	__antithesis_instrumentation__.Notify(10283)
	r, err := dest.ReadFile(ctx, backupEncryptionInfoFile)
	if err == nil {
		__antithesis_instrumentation__.Notify(10288)
		r.Close(ctx)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(10289)
	}
	__antithesis_instrumentation__.Notify(10284)

	if !errors.Is(err, cloud.ErrFileDoesNotExist) {
		__antithesis_instrumentation__.Notify(10290)
		return errors.Wrapf(err,
			"returned an unexpected error when checking for the existence of %s file",
			backupEncryptionInfoFile)
	} else {
		__antithesis_instrumentation__.Notify(10291)
	}
	__antithesis_instrumentation__.Notify(10285)

	buf, err := protoutil.Marshal(opts)
	if err != nil {
		__antithesis_instrumentation__.Notify(10292)
		return err
	} else {
		__antithesis_instrumentation__.Notify(10293)
	}
	__antithesis_instrumentation__.Notify(10286)
	if err := cloud.WriteFile(ctx, dest, backupEncryptionInfoFile, bytes.NewReader(buf)); err != nil {
		__antithesis_instrumentation__.Notify(10294)
		return err
	} else {
		__antithesis_instrumentation__.Notify(10295)
	}
	__antithesis_instrumentation__.Notify(10287)
	return nil
}

func RedactURIForErrorMessage(uri string) string {
	__antithesis_instrumentation__.Notify(10296)
	redactedURI, err := cloud.SanitizeExternalStorageURI(uri, []string{})
	if err != nil {
		__antithesis_instrumentation__.Notify(10298)
		return "<uri_failed_to_redact>"
	} else {
		__antithesis_instrumentation__.Notify(10299)
	}
	__antithesis_instrumentation__.Notify(10297)
	return redactedURI
}

func checkForPreviousBackup(
	ctx context.Context, exportStore cloud.ExternalStorage, defaultURI string,
) error {
	__antithesis_instrumentation__.Notify(10300)
	redactedURI := RedactURIForErrorMessage(defaultURI)
	r, err := exportStore.ReadFile(ctx, backupManifestName)
	if err == nil {
		__antithesis_instrumentation__.Notify(10305)
		r.Close(ctx)
		return pgerror.Newf(pgcode.FileAlreadyExists,
			"%s already contains a %s file",
			redactedURI, backupManifestName)
	} else {
		__antithesis_instrumentation__.Notify(10306)
	}
	__antithesis_instrumentation__.Notify(10301)

	if !errors.Is(err, cloud.ErrFileDoesNotExist) {
		__antithesis_instrumentation__.Notify(10307)
		return errors.Wrapf(err,
			"%s returned an unexpected error when checking for the existence of %s file",
			redactedURI, backupManifestName)
	} else {
		__antithesis_instrumentation__.Notify(10308)
	}
	__antithesis_instrumentation__.Notify(10302)

	r, err = readLatestCheckpointFile(ctx, exportStore, backupManifestCheckpointName)
	if err == nil {
		__antithesis_instrumentation__.Notify(10309)
		r.Close(ctx)
		return pgerror.Newf(pgcode.FileAlreadyExists,
			"%s already contains a %s file (is another operation already in progress?)",
			redactedURI, backupManifestCheckpointName)
	} else {
		__antithesis_instrumentation__.Notify(10310)
	}
	__antithesis_instrumentation__.Notify(10303)

	if !errors.Is(err, cloud.ErrFileDoesNotExist) {
		__antithesis_instrumentation__.Notify(10311)
		return errors.Wrapf(err,
			"%s returned an unexpected error when checking for the existence of %s file",
			redactedURI, backupManifestCheckpointName)
	} else {
		__antithesis_instrumentation__.Notify(10312)
	}
	__antithesis_instrumentation__.Notify(10304)

	return nil
}

func tempCheckpointFileNameForJob(jobID jobspb.JobID) string {
	__antithesis_instrumentation__.Notify(10313)
	return fmt.Sprintf("%s-%d", backupManifestCheckpointName, jobID)
}

func ListFullBackupsInCollection(
	ctx context.Context, store cloud.ExternalStorage,
) ([]string, error) {
	__antithesis_instrumentation__.Notify(10314)
	var backupPaths []string
	if err := store.List(ctx, "", listingDelimDataSlash, func(f string) error {
		__antithesis_instrumentation__.Notify(10317)
		if backupPathRE.MatchString(f) {
			__antithesis_instrumentation__.Notify(10319)
			backupPaths = append(backupPaths, f)
		} else {
			__antithesis_instrumentation__.Notify(10320)
		}
		__antithesis_instrumentation__.Notify(10318)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(10321)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(10322)
	}
	__antithesis_instrumentation__.Notify(10315)
	for i, backupPath := range backupPaths {
		__antithesis_instrumentation__.Notify(10323)
		backupPaths[i] = strings.TrimSuffix(backupPath, "/"+backupManifestName)
	}
	__antithesis_instrumentation__.Notify(10316)
	return backupPaths, nil
}

func readLatestCheckpointFile(
	ctx context.Context, exportStore cloud.ExternalStorage, filename string,
) (ioctx.ReadCloserCtx, error) {
	__antithesis_instrumentation__.Notify(10324)
	filename = strings.TrimPrefix(filename, "/")

	var checkpoint string
	var checkpointFound bool
	var r ioctx.ReadCloserCtx
	var err error

	err = exportStore.List(ctx, backupProgressDirectory, "", func(p string) error {
		__antithesis_instrumentation__.Notify(10329)

		p = strings.TrimPrefix(p, "/")
		checkpoint = strings.TrimSuffix(p, backupManifestChecksumSuffix)
		checkpointFound = true

		return cloud.ErrListingDone
	})
	__antithesis_instrumentation__.Notify(10325)

	if errors.Is(err, cloud.ErrListingUnsupported) {
		__antithesis_instrumentation__.Notify(10330)
		r, err = exportStore.ReadFile(ctx, backupProgressDirectory+"/"+filename)

		if err == nil {
			__antithesis_instrumentation__.Notify(10331)
			return r, nil
		} else {
			__antithesis_instrumentation__.Notify(10332)
		}
	} else {
		__antithesis_instrumentation__.Notify(10333)
		if err != nil && func() bool {
			__antithesis_instrumentation__.Notify(10334)
			return !errors.Is(err, cloud.ErrListingDone) == true
		}() == true {
			__antithesis_instrumentation__.Notify(10335)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(10336)
		}
	}
	__antithesis_instrumentation__.Notify(10326)

	if checkpointFound {
		__antithesis_instrumentation__.Notify(10337)
		if strings.HasSuffix(filename, backupManifestChecksumSuffix) {
			__antithesis_instrumentation__.Notify(10339)
			return exportStore.ReadFile(ctx, backupProgressDirectory+"/"+checkpoint+backupManifestChecksumSuffix)
		} else {
			__antithesis_instrumentation__.Notify(10340)
		}
		__antithesis_instrumentation__.Notify(10338)
		return exportStore.ReadFile(ctx, backupProgressDirectory+"/"+checkpoint)
	} else {
		__antithesis_instrumentation__.Notify(10341)
	}
	__antithesis_instrumentation__.Notify(10327)

	r, err = exportStore.ReadFile(ctx, filename)

	if err != nil {
		__antithesis_instrumentation__.Notify(10342)
		return nil, errors.Wrapf(err, "%s could not be read in the base or progress directory", filename)
	} else {
		__antithesis_instrumentation__.Notify(10343)
	}
	__antithesis_instrumentation__.Notify(10328)
	return r, nil

}

func newTimestampedCheckpointFileName() string {
	__antithesis_instrumentation__.Notify(10344)
	var buffer []byte
	buffer = encoding.EncodeStringDescending(buffer, timeutil.Now().String())
	return fmt.Sprintf("%s-%s", backupManifestCheckpointName, hex.EncodeToString(buffer))
}
