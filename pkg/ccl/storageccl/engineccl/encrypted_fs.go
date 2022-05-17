package engineccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/ccl/baseccl"
	"github.com/cockroachdb/cockroach/pkg/ccl/storageccl/engineccl/enginepbccl"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/pebble/vfs"
)

type encryptedFile struct {
	vfs.File
	mu struct {
		syncutil.Mutex
		rOffset int64
		wOffset int64
	}
	stream FileStream
}

func (f *encryptedFile) Write(p []byte) (n int, err error) {
	__antithesis_instrumentation__.Notify(23510)
	f.mu.Lock()
	defer f.mu.Unlock()
	f.stream.Encrypt(f.mu.wOffset, p)
	n, err = f.File.Write(p)
	f.mu.wOffset += int64(n)
	return n, err
}

func (f *encryptedFile) Read(p []byte) (n int, err error) {
	__antithesis_instrumentation__.Notify(23511)
	f.mu.Lock()
	defer f.mu.Unlock()
	n, err = f.ReadAt(p, f.mu.rOffset)
	f.mu.rOffset += int64(n)
	return n, err
}

func (f *encryptedFile) ReadAt(p []byte, off int64) (n int, err error) {
	__antithesis_instrumentation__.Notify(23512)
	n, err = f.File.ReadAt(p, off)
	if n > 0 {
		__antithesis_instrumentation__.Notify(23514)
		f.stream.Decrypt(off, p[:n])
	} else {
		__antithesis_instrumentation__.Notify(23515)
	}
	__antithesis_instrumentation__.Notify(23513)
	return n, err
}

type encryptedFS struct {
	vfs.FS
	fileRegistry  *storage.PebbleFileRegistry
	streamCreator *FileCipherStreamCreator
}

func (fs *encryptedFS) Create(name string) (vfs.File, error) {
	__antithesis_instrumentation__.Notify(23516)
	f, err := fs.FS.Create(name)
	if err != nil {
		__antithesis_instrumentation__.Notify(23520)
		return f, err
	} else {
		__antithesis_instrumentation__.Notify(23521)
	}
	__antithesis_instrumentation__.Notify(23517)

	settings, stream, err := fs.streamCreator.CreateNew(context.TODO())
	if err != nil {
		__antithesis_instrumentation__.Notify(23522)
		_ = f.Close()
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(23523)
	}
	__antithesis_instrumentation__.Notify(23518)

	if settings.EncryptionType == enginepbccl.EncryptionType_Plaintext {
		__antithesis_instrumentation__.Notify(23524)
		if err := fs.fileRegistry.MaybeDeleteEntry(name); err != nil {
			__antithesis_instrumentation__.Notify(23525)
			_ = f.Close()
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(23526)
		}
	} else {
		__antithesis_instrumentation__.Notify(23527)
		fproto := &enginepb.FileEntry{}
		fproto.EnvType = fs.streamCreator.envType
		if fproto.EncryptionSettings, err = protoutil.Marshal(settings); err != nil {
			__antithesis_instrumentation__.Notify(23529)
			_ = f.Close()
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(23530)
		}
		__antithesis_instrumentation__.Notify(23528)
		if err := fs.fileRegistry.SetFileEntry(name, fproto); err != nil {
			__antithesis_instrumentation__.Notify(23531)
			_ = f.Close()
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(23532)
		}
	}
	__antithesis_instrumentation__.Notify(23519)
	ef := &encryptedFile{File: f, stream: stream}
	return vfs.WithFd(f, ef), nil
}

func (fs *encryptedFS) Link(oldname, newname string) error {
	__antithesis_instrumentation__.Notify(23533)
	if err := fs.FS.Link(oldname, newname); err != nil {
		__antithesis_instrumentation__.Notify(23535)
		return err
	} else {
		__antithesis_instrumentation__.Notify(23536)
	}
	__antithesis_instrumentation__.Notify(23534)
	return fs.fileRegistry.MaybeLinkEntry(oldname, newname)
}

func (fs *encryptedFS) Open(name string, opts ...vfs.OpenOption) (vfs.File, error) {
	__antithesis_instrumentation__.Notify(23537)
	f, err := fs.FS.Open(name, opts...)
	if err != nil {
		__antithesis_instrumentation__.Notify(23541)
		return f, err
	} else {
		__antithesis_instrumentation__.Notify(23542)
	}
	__antithesis_instrumentation__.Notify(23538)
	fileEntry := fs.fileRegistry.GetFileEntry(name)
	var settings *enginepbccl.EncryptionSettings
	if fileEntry != nil {
		__antithesis_instrumentation__.Notify(23543)
		if fileEntry.EnvType != fs.streamCreator.envType {
			__antithesis_instrumentation__.Notify(23545)
			f.Close()
			return nil, fmt.Errorf("filename: %s has env %d not equal to FS env %d",
				name, fileEntry.EnvType, fs.streamCreator.envType)
		} else {
			__antithesis_instrumentation__.Notify(23546)
		}
		__antithesis_instrumentation__.Notify(23544)
		settings = &enginepbccl.EncryptionSettings{}
		if err := protoutil.Unmarshal(fileEntry.EncryptionSettings, settings); err != nil {
			__antithesis_instrumentation__.Notify(23547)
			f.Close()
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(23548)
		}
	} else {
		__antithesis_instrumentation__.Notify(23549)
	}
	__antithesis_instrumentation__.Notify(23539)
	stream, err := fs.streamCreator.CreateExisting(settings)
	if err != nil {
		__antithesis_instrumentation__.Notify(23550)
		f.Close()
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(23551)
	}
	__antithesis_instrumentation__.Notify(23540)
	ef := &encryptedFile{File: f, stream: stream}
	return vfs.WithFd(f, ef), nil
}

func (fs *encryptedFS) Remove(name string) error {
	__antithesis_instrumentation__.Notify(23552)
	if err := fs.FS.Remove(name); err != nil {
		__antithesis_instrumentation__.Notify(23554)
		return err
	} else {
		__antithesis_instrumentation__.Notify(23555)
	}
	__antithesis_instrumentation__.Notify(23553)
	return fs.fileRegistry.MaybeDeleteEntry(name)
}

func (fs *encryptedFS) Rename(oldname, newname string) error {
	__antithesis_instrumentation__.Notify(23556)

	if err := fs.fileRegistry.MaybeCopyEntry(oldname, newname); err != nil {
		__antithesis_instrumentation__.Notify(23559)
		return err
	} else {
		__antithesis_instrumentation__.Notify(23560)
	}
	__antithesis_instrumentation__.Notify(23557)

	if err := fs.FS.Rename(oldname, newname); err != nil {
		__antithesis_instrumentation__.Notify(23561)
		return err
	} else {
		__antithesis_instrumentation__.Notify(23562)
	}
	__antithesis_instrumentation__.Notify(23558)

	return fs.fileRegistry.MaybeDeleteEntry(oldname)
}

func (fs *encryptedFS) ReuseForWrite(oldname, newname string) (vfs.File, error) {
	__antithesis_instrumentation__.Notify(23563)

	if err := fs.Remove(oldname); err != nil {
		__antithesis_instrumentation__.Notify(23565)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(23566)
	}
	__antithesis_instrumentation__.Notify(23564)
	return fs.Create(newname)
}

type encryptionStatsHandler struct {
	storeKM *StoreKeyManager
	dataKM  *DataKeyManager
}

func (e *encryptionStatsHandler) GetEncryptionStatus() ([]byte, error) {
	__antithesis_instrumentation__.Notify(23567)
	var s enginepbccl.EncryptionStatus
	if e.storeKM.activeKey != nil {
		__antithesis_instrumentation__.Notify(23571)
		s.ActiveStoreKey = e.storeKM.activeKey.Info
	} else {
		__antithesis_instrumentation__.Notify(23572)
	}
	__antithesis_instrumentation__.Notify(23568)
	k, err := e.dataKM.ActiveKey(context.TODO())
	if err != nil {
		__antithesis_instrumentation__.Notify(23573)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(23574)
	}
	__antithesis_instrumentation__.Notify(23569)
	if k != nil {
		__antithesis_instrumentation__.Notify(23575)
		s.ActiveDataKey = k.Info
	} else {
		__antithesis_instrumentation__.Notify(23576)
	}
	__antithesis_instrumentation__.Notify(23570)
	return protoutil.Marshal(&s)
}

func (e *encryptionStatsHandler) GetDataKeysRegistry() ([]byte, error) {
	__antithesis_instrumentation__.Notify(23577)
	r := e.dataKM.getScrubbedRegistry()
	return protoutil.Marshal(r)
}

func (e *encryptionStatsHandler) GetActiveDataKeyID() (string, error) {
	__antithesis_instrumentation__.Notify(23578)
	k, err := e.dataKM.ActiveKey(context.TODO())
	if err != nil {
		__antithesis_instrumentation__.Notify(23581)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(23582)
	}
	__antithesis_instrumentation__.Notify(23579)
	if k != nil {
		__antithesis_instrumentation__.Notify(23583)
		return k.Info.KeyId, nil
	} else {
		__antithesis_instrumentation__.Notify(23584)
	}
	__antithesis_instrumentation__.Notify(23580)
	return "plain", nil
}

func (e *encryptionStatsHandler) GetActiveStoreKeyType() int32 {
	__antithesis_instrumentation__.Notify(23585)
	if e.storeKM.activeKey != nil {
		__antithesis_instrumentation__.Notify(23587)
		return int32(e.storeKM.activeKey.Info.EncryptionType)
	} else {
		__antithesis_instrumentation__.Notify(23588)
	}
	__antithesis_instrumentation__.Notify(23586)
	return int32(enginepbccl.EncryptionType_Plaintext)
}

func (e *encryptionStatsHandler) GetKeyIDFromSettings(settings []byte) (string, error) {
	__antithesis_instrumentation__.Notify(23589)
	var s enginepbccl.EncryptionSettings
	if err := protoutil.Unmarshal(settings, &s); err != nil {
		__antithesis_instrumentation__.Notify(23591)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(23592)
	}
	__antithesis_instrumentation__.Notify(23590)
	return s.KeyId, nil
}

func init() {
	storage.NewEncryptedEnvFunc = newEncryptedEnv
	storage.CanRegistryElideFunc = canRegistryElide
}

func newEncryptedEnv(
	fs vfs.FS, fr *storage.PebbleFileRegistry, dbDir string, readOnly bool, optionBytes []byte,
) (*storage.EncryptionEnv, error) {
	__antithesis_instrumentation__.Notify(23593)
	options := &baseccl.EncryptionOptions{}
	if err := protoutil.Unmarshal(optionBytes, options); err != nil {
		__antithesis_instrumentation__.Notify(23599)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(23600)
	}
	__antithesis_instrumentation__.Notify(23594)
	if options.KeySource != baseccl.EncryptionKeySource_KeyFiles {
		__antithesis_instrumentation__.Notify(23601)
		return nil, fmt.Errorf("unknown encryption key source: %d", options.KeySource)
	} else {
		__antithesis_instrumentation__.Notify(23602)
	}
	__antithesis_instrumentation__.Notify(23595)
	storeKeyManager := &StoreKeyManager{
		fs:                fs,
		activeKeyFilename: options.KeyFiles.CurrentKey,
		oldKeyFilename:    options.KeyFiles.OldKey,
	}
	if err := storeKeyManager.Load(context.TODO()); err != nil {
		__antithesis_instrumentation__.Notify(23603)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(23604)
	}
	__antithesis_instrumentation__.Notify(23596)
	storeFS := &encryptedFS{
		FS:           fs,
		fileRegistry: fr,
		streamCreator: &FileCipherStreamCreator{
			envType:    enginepb.EnvType_Store,
			keyManager: storeKeyManager,
		},
	}
	dataKeyManager := &DataKeyManager{
		fs:             storeFS,
		dbDir:          dbDir,
		rotationPeriod: options.DataKeyRotationPeriod,
		readOnly:       readOnly,
	}
	if err := dataKeyManager.Load(context.TODO()); err != nil {
		__antithesis_instrumentation__.Notify(23605)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(23606)
	}
	__antithesis_instrumentation__.Notify(23597)
	dataFS := &encryptedFS{
		FS:           fs,
		fileRegistry: fr,
		streamCreator: &FileCipherStreamCreator{
			envType:    enginepb.EnvType_Data,
			keyManager: dataKeyManager,
		},
	}

	if !readOnly {
		__antithesis_instrumentation__.Notify(23607)
		key, err := storeKeyManager.ActiveKey(context.TODO())
		if err != nil {
			__antithesis_instrumentation__.Notify(23609)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(23610)
		}
		__antithesis_instrumentation__.Notify(23608)
		if err := dataKeyManager.SetActiveStoreKeyInfo(context.TODO(), key.Info); err != nil {
			__antithesis_instrumentation__.Notify(23611)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(23612)
		}
	} else {
		__antithesis_instrumentation__.Notify(23613)
	}
	__antithesis_instrumentation__.Notify(23598)

	return &storage.EncryptionEnv{
		Closer: dataKeyManager,
		FS:     dataFS,
		StatsHandler: &encryptionStatsHandler{
			storeKM: storeKeyManager,
			dataKM:  dataKeyManager,
		},
	}, nil
}

func canRegistryElide(entry *enginepb.FileEntry) bool {
	__antithesis_instrumentation__.Notify(23614)
	if entry == nil {
		__antithesis_instrumentation__.Notify(23617)
		return true
	} else {
		__antithesis_instrumentation__.Notify(23618)
	}
	__antithesis_instrumentation__.Notify(23615)
	settings := &enginepbccl.EncryptionSettings{}
	if err := protoutil.Unmarshal(entry.EncryptionSettings, settings); err != nil {
		__antithesis_instrumentation__.Notify(23619)
		return false
	} else {
		__antithesis_instrumentation__.Notify(23620)
	}
	__antithesis_instrumentation__.Notify(23616)
	return settings.EncryptionType == enginepbccl.EncryptionType_Plaintext
}
