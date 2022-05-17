package engineccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/cockroachdb/cockroach/pkg/ccl/storageccl/engineccl/enginepbccl"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/oserror"
	"github.com/cockroachdb/pebble/vfs"
	"github.com/cockroachdb/pebble/vfs/atomicfs"
	"github.com/gogo/protobuf/proto"
)

const (
	storeFileNamePlain = "plain"

	plainKeyID = "plain"

	keyIDLength = 32

	keyRegistryFilename = "COCKROACHDB_DATA_KEYS"

	keysRegistryMarkerName = "datakeys"
)

type registryFormat string

const (
	registryFormatMonolith registryFormat = "monolith"
)

type PebbleKeyManager interface {
	ActiveKey(ctx context.Context) (*enginepbccl.SecretKey, error)

	GetKey(id string) (*enginepbccl.SecretKey, error)
}

var _ PebbleKeyManager = &StoreKeyManager{}
var _ PebbleKeyManager = &DataKeyManager{}

var kmTimeNow = time.Now

type StoreKeyManager struct {
	fs                vfs.FS
	activeKeyFilename string
	oldKeyFilename    string

	activeKey *enginepbccl.SecretKey
	oldKey    *enginepbccl.SecretKey
}

func (m *StoreKeyManager) Load(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(24695)
	var err error
	m.activeKey, err = loadKeyFromFile(m.fs, m.activeKeyFilename)
	if err != nil {
		__antithesis_instrumentation__.Notify(24698)
		return err
	} else {
		__antithesis_instrumentation__.Notify(24699)
	}
	__antithesis_instrumentation__.Notify(24696)
	m.oldKey, err = loadKeyFromFile(m.fs, m.oldKeyFilename)
	if err != nil {
		__antithesis_instrumentation__.Notify(24700)
		return err
	} else {
		__antithesis_instrumentation__.Notify(24701)
	}
	__antithesis_instrumentation__.Notify(24697)
	log.Infof(ctx, "loaded active store key: %s, old store key: %s",
		proto.CompactTextString(m.activeKey.Info), proto.CompactTextString(m.oldKey.Info))
	return nil
}

func (m *StoreKeyManager) ActiveKey(ctx context.Context) (*enginepbccl.SecretKey, error) {
	__antithesis_instrumentation__.Notify(24702)
	return m.activeKey, nil
}

func (m *StoreKeyManager) GetKey(id string) (*enginepbccl.SecretKey, error) {
	__antithesis_instrumentation__.Notify(24703)
	if m.activeKey.Info.KeyId == id {
		__antithesis_instrumentation__.Notify(24706)
		return m.activeKey, nil
	} else {
		__antithesis_instrumentation__.Notify(24707)
	}
	__antithesis_instrumentation__.Notify(24704)
	if m.oldKey.Info.KeyId == id {
		__antithesis_instrumentation__.Notify(24708)
		return m.oldKey, nil
	} else {
		__antithesis_instrumentation__.Notify(24709)
	}
	__antithesis_instrumentation__.Notify(24705)
	return nil, fmt.Errorf("store key ID %s was not found", id)
}

func loadKeyFromFile(fs vfs.FS, filename string) (*enginepbccl.SecretKey, error) {
	__antithesis_instrumentation__.Notify(24710)
	now := kmTimeNow().Unix()
	key := &enginepbccl.SecretKey{}
	key.Info = &enginepbccl.KeyInfo{}
	if filename == storeFileNamePlain {
		__antithesis_instrumentation__.Notify(24715)
		key.Info.EncryptionType = enginepbccl.EncryptionType_Plaintext
		key.Info.KeyId = plainKeyID
		key.Info.CreationTime = now
		key.Info.Source = storeFileNamePlain
		return key, nil
	} else {
		__antithesis_instrumentation__.Notify(24716)
	}
	__antithesis_instrumentation__.Notify(24711)

	f, err := fs.Open(filename)
	if err != nil {
		__antithesis_instrumentation__.Notify(24717)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(24718)
	}
	__antithesis_instrumentation__.Notify(24712)
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		__antithesis_instrumentation__.Notify(24719)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(24720)
	}
	__antithesis_instrumentation__.Notify(24713)

	keyLength := len(b) - keyIDLength
	switch keyLength {
	case 16:
		__antithesis_instrumentation__.Notify(24721)
		key.Info.EncryptionType = enginepbccl.EncryptionType_AES128_CTR
	case 24:
		__antithesis_instrumentation__.Notify(24722)
		key.Info.EncryptionType = enginepbccl.EncryptionType_AES192_CTR
	case 32:
		__antithesis_instrumentation__.Notify(24723)
		key.Info.EncryptionType = enginepbccl.EncryptionType_AES256_CTR
	default:
		__antithesis_instrumentation__.Notify(24724)
		return nil, fmt.Errorf("store key of unsupported length: %d", keyLength)
	}
	__antithesis_instrumentation__.Notify(24714)
	key.Key = b[keyIDLength:]

	key.Info.KeyId = hex.EncodeToString(b[:keyIDLength])
	key.Info.CreationTime = now
	key.Info.Source = filename

	return key, nil
}

type DataKeyManager struct {
	fs             vfs.FS
	dbDir          string
	rotationPeriod int64
	readOnly       bool

	mu struct {
		syncutil.Mutex

		keyRegistry *enginepbccl.DataKeysRegistry

		activeKey *enginepbccl.SecretKey

		rotationEnabled bool

		marker *atomicfs.Marker

		filename string
	}
}

func makeRegistryProto() *enginepbccl.DataKeysRegistry {
	__antithesis_instrumentation__.Notify(24725)
	return &enginepbccl.DataKeysRegistry{
		StoreKeys: make(map[string]*enginepbccl.KeyInfo),
		DataKeys:  make(map[string]*enginepbccl.SecretKey),
	}
}

func (m *DataKeyManager) Close() error {
	__antithesis_instrumentation__.Notify(24726)
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.mu.marker.Close()
}

func (m *DataKeyManager) Load(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(24727)
	marker, filename, err := atomicfs.LocateMarker(m.fs, m.dbDir, keysRegistryMarkerName)
	if err != nil {
		__antithesis_instrumentation__.Notify(24732)
		return err
	} else {
		__antithesis_instrumentation__.Notify(24733)
	}
	__antithesis_instrumentation__.Notify(24728)

	m.mu.Lock()
	defer m.mu.Unlock()
	m.mu.marker = marker
	if oserror.IsNotExist(err) {
		__antithesis_instrumentation__.Notify(24734)

		m.mu.keyRegistry = makeRegistryProto()
		return nil
	} else {
		__antithesis_instrumentation__.Notify(24735)
	}
	__antithesis_instrumentation__.Notify(24729)

	m.mu.filename = filename
	m.mu.keyRegistry = makeRegistryProto()
	if filename != "" {
		__antithesis_instrumentation__.Notify(24736)
		f, err := m.fs.Open(m.fs.PathJoin(m.dbDir, filename))
		if err != nil {
			__antithesis_instrumentation__.Notify(24740)
			return err
		} else {
			__antithesis_instrumentation__.Notify(24741)
		}
		__antithesis_instrumentation__.Notify(24737)
		defer f.Close()
		b, err := ioutil.ReadAll(f)
		if err != nil {
			__antithesis_instrumentation__.Notify(24742)
			return err
		} else {
			__antithesis_instrumentation__.Notify(24743)
		}
		__antithesis_instrumentation__.Notify(24738)
		if err = protoutil.Unmarshal(b, m.mu.keyRegistry); err != nil {
			__antithesis_instrumentation__.Notify(24744)
			return err
		} else {
			__antithesis_instrumentation__.Notify(24745)
		}
		__antithesis_instrumentation__.Notify(24739)
		if err = validateRegistry(m.mu.keyRegistry); err != nil {
			__antithesis_instrumentation__.Notify(24746)
			return err
		} else {
			__antithesis_instrumentation__.Notify(24747)
		}
	} else {
		__antithesis_instrumentation__.Notify(24748)
	}
	__antithesis_instrumentation__.Notify(24730)

	if m.mu.keyRegistry.ActiveDataKeyId != "" {
		__antithesis_instrumentation__.Notify(24749)
		key, found := m.mu.keyRegistry.DataKeys[m.mu.keyRegistry.ActiveDataKeyId]
		if !found {
			__antithesis_instrumentation__.Notify(24751)

			panic("unexpected inconsistent DataKeysRegistry")
		} else {
			__antithesis_instrumentation__.Notify(24752)
		}
		__antithesis_instrumentation__.Notify(24750)
		m.mu.activeKey = key
		log.Infof(ctx, "loaded active data key: %s", m.mu.activeKey.Info.String())
	} else {
		__antithesis_instrumentation__.Notify(24753)
		log.Infof(ctx, "no active data key yet")
	}
	__antithesis_instrumentation__.Notify(24731)
	return nil
}

func (m *DataKeyManager) ActiveKey(ctx context.Context) (*enginepbccl.SecretKey, error) {
	__antithesis_instrumentation__.Notify(24754)
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.mu.rotationEnabled {
		__antithesis_instrumentation__.Notify(24756)
		now := kmTimeNow().Unix()
		if now-m.mu.activeKey.Info.CreationTime > m.rotationPeriod {
			__antithesis_instrumentation__.Notify(24757)
			keyRegistry := makeRegistryProto()
			proto.Merge(keyRegistry, m.mu.keyRegistry)
			if err := m.rotateDataKeyAndWrite(ctx, keyRegistry); err != nil {
				__antithesis_instrumentation__.Notify(24758)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(24759)
			}
		} else {
			__antithesis_instrumentation__.Notify(24760)
		}
	} else {
		__antithesis_instrumentation__.Notify(24761)
	}
	__antithesis_instrumentation__.Notify(24755)
	return m.mu.activeKey, nil
}

func (m *DataKeyManager) GetKey(id string) (*enginepbccl.SecretKey, error) {
	__antithesis_instrumentation__.Notify(24762)
	m.mu.Lock()
	defer m.mu.Unlock()
	key, found := m.mu.keyRegistry.DataKeys[id]
	if !found {
		__antithesis_instrumentation__.Notify(24764)
		return nil, fmt.Errorf("key %s is not found", id)
	} else {
		__antithesis_instrumentation__.Notify(24765)
	}
	__antithesis_instrumentation__.Notify(24763)
	return key, nil
}

func (m *DataKeyManager) SetActiveStoreKeyInfo(
	ctx context.Context, storeKeyInfo *enginepbccl.KeyInfo,
) error {
	__antithesis_instrumentation__.Notify(24766)
	if m.readOnly {
		__antithesis_instrumentation__.Notify(24772)
		return errors.New("read only")
	} else {
		__antithesis_instrumentation__.Notify(24773)
	}
	__antithesis_instrumentation__.Notify(24767)
	m.mu.Lock()
	defer m.mu.Unlock()

	m.mu.rotationEnabled = true
	prevActiveStoreKey, found := m.mu.keyRegistry.StoreKeys[m.mu.keyRegistry.ActiveStoreKeyId]
	if found && func() bool {
		__antithesis_instrumentation__.Notify(24774)
		return prevActiveStoreKey.KeyId == storeKeyInfo.KeyId == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(24775)
		return m.mu.activeKey != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(24776)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(24777)
	}
	__antithesis_instrumentation__.Notify(24768)

	if storeKeyInfo.EncryptionType != enginepbccl.EncryptionType_Plaintext {
		__antithesis_instrumentation__.Notify(24778)
		if _, found := m.mu.keyRegistry.StoreKeys[storeKeyInfo.KeyId]; found {
			__antithesis_instrumentation__.Notify(24779)
			return fmt.Errorf("new active store key ID %s already exists as an inactive key -- this"+
				"is really dangerous", storeKeyInfo.KeyId)
		} else {
			__antithesis_instrumentation__.Notify(24780)
		}
	} else {
		__antithesis_instrumentation__.Notify(24781)
	}
	__antithesis_instrumentation__.Notify(24769)

	keyRegistry := makeRegistryProto()
	proto.Merge(keyRegistry, m.mu.keyRegistry)
	keyRegistry.StoreKeys[storeKeyInfo.KeyId] = storeKeyInfo
	keyRegistry.ActiveStoreKeyId = storeKeyInfo.KeyId
	if storeKeyInfo.EncryptionType == enginepbccl.EncryptionType_Plaintext {
		__antithesis_instrumentation__.Notify(24782)

		for _, key := range keyRegistry.DataKeys {
			__antithesis_instrumentation__.Notify(24783)
			key.Info.WasExposed = true
		}
	} else {
		__antithesis_instrumentation__.Notify(24784)
	}
	__antithesis_instrumentation__.Notify(24770)
	if err := m.rotateDataKeyAndWrite(ctx, keyRegistry); err != nil {
		__antithesis_instrumentation__.Notify(24785)
		return err
	} else {
		__antithesis_instrumentation__.Notify(24786)
	}
	__antithesis_instrumentation__.Notify(24771)
	m.mu.rotationEnabled = true
	return nil
}

func (m *DataKeyManager) getScrubbedRegistry() *enginepbccl.DataKeysRegistry {
	__antithesis_instrumentation__.Notify(24787)
	m.mu.Lock()
	defer m.mu.Unlock()
	r := makeRegistryProto()
	proto.Merge(r, m.mu.keyRegistry)
	for _, v := range r.DataKeys {
		__antithesis_instrumentation__.Notify(24789)
		v.Key = nil
	}
	__antithesis_instrumentation__.Notify(24788)
	return r
}

func validateRegistry(keyRegistry *enginepbccl.DataKeysRegistry) error {
	__antithesis_instrumentation__.Notify(24790)
	if keyRegistry.ActiveStoreKeyId != "" && func() bool {
		__antithesis_instrumentation__.Notify(24793)
		return keyRegistry.StoreKeys[keyRegistry.ActiveStoreKeyId] == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(24794)
		return fmt.Errorf("active store key %s not found", keyRegistry.ActiveStoreKeyId)
	} else {
		__antithesis_instrumentation__.Notify(24795)
	}
	__antithesis_instrumentation__.Notify(24791)
	if keyRegistry.ActiveDataKeyId != "" && func() bool {
		__antithesis_instrumentation__.Notify(24796)
		return keyRegistry.DataKeys[keyRegistry.ActiveDataKeyId] == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(24797)
		return fmt.Errorf("active data key %s not found", keyRegistry.ActiveDataKeyId)
	} else {
		__antithesis_instrumentation__.Notify(24798)
	}
	__antithesis_instrumentation__.Notify(24792)
	return nil
}

func generateAndSetNewDataKey(
	ctx context.Context, keyRegistry *enginepbccl.DataKeysRegistry,
) (*enginepbccl.SecretKey, error) {
	__antithesis_instrumentation__.Notify(24799)
	activeStoreKey := keyRegistry.StoreKeys[keyRegistry.ActiveStoreKeyId]
	if activeStoreKey == nil {
		__antithesis_instrumentation__.Notify(24802)
		panic("expected registry with active store key")
	} else {
		__antithesis_instrumentation__.Notify(24803)
	}
	__antithesis_instrumentation__.Notify(24800)
	key := &enginepbccl.SecretKey{}
	key.Info = &enginepbccl.KeyInfo{}
	key.Info.EncryptionType = activeStoreKey.EncryptionType
	key.Info.CreationTime = kmTimeNow().Unix()
	key.Info.Source = "data key manager"
	key.Info.ParentKeyId = activeStoreKey.KeyId

	if activeStoreKey.EncryptionType == enginepbccl.EncryptionType_Plaintext {
		__antithesis_instrumentation__.Notify(24804)
		key.Info.KeyId = plainKeyID
		key.Info.WasExposed = true
	} else {
		__antithesis_instrumentation__.Notify(24805)
		var keyLength int
		switch activeStoreKey.EncryptionType {
		case enginepbccl.EncryptionType_AES128_CTR:
			__antithesis_instrumentation__.Notify(24811)
			keyLength = 16
		case enginepbccl.EncryptionType_AES192_CTR:
			__antithesis_instrumentation__.Notify(24812)
			keyLength = 24
		case enginepbccl.EncryptionType_AES256_CTR:
			__antithesis_instrumentation__.Notify(24813)
			keyLength = 32
		default:
			__antithesis_instrumentation__.Notify(24814)
			return nil, fmt.Errorf("unknown encryption type %d for key ID %s",
				activeStoreKey.EncryptionType, activeStoreKey.KeyId)
		}
		__antithesis_instrumentation__.Notify(24806)
		key.Key = make([]byte, keyLength)
		n, err := rand.Read(key.Key)
		if err != nil {
			__antithesis_instrumentation__.Notify(24815)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(24816)
		}
		__antithesis_instrumentation__.Notify(24807)
		if n != keyLength {
			__antithesis_instrumentation__.Notify(24817)
			log.Fatalf(ctx, "rand.Read returned no error but fewer bytes %d than promised %d", n, keyLength)
		} else {
			__antithesis_instrumentation__.Notify(24818)
		}
		__antithesis_instrumentation__.Notify(24808)
		keyID := make([]byte, keyIDLength)
		if n, err = rand.Read(keyID); err != nil {
			__antithesis_instrumentation__.Notify(24819)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(24820)
		}
		__antithesis_instrumentation__.Notify(24809)
		if n != keyIDLength {
			__antithesis_instrumentation__.Notify(24821)
			log.Fatalf(ctx, "rand.Read returned no error but fewer bytes %d than promised %d", n, keyIDLength)
		} else {
			__antithesis_instrumentation__.Notify(24822)
		}
		__antithesis_instrumentation__.Notify(24810)

		key.Info.KeyId = hex.EncodeToString(keyID)
		key.Info.WasExposed = false
	}
	__antithesis_instrumentation__.Notify(24801)
	keyRegistry.DataKeys[key.Info.KeyId] = key
	keyRegistry.ActiveDataKeyId = key.Info.KeyId
	return key, nil
}

func (m *DataKeyManager) rotateDataKeyAndWrite(
	ctx context.Context, keyRegistry *enginepbccl.DataKeysRegistry,
) (err error) {
	__antithesis_instrumentation__.Notify(24823)
	defer func() {
		__antithesis_instrumentation__.Notify(24833)
		if err != nil {
			__antithesis_instrumentation__.Notify(24834)
			log.Infof(ctx, "error while attempting to rotate data key: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(24835)
			log.Infof(ctx, "rotated to new active data key: %s", proto.CompactTextString(m.mu.activeKey.Info))
		}
	}()
	__antithesis_instrumentation__.Notify(24824)

	var newKey *enginepbccl.SecretKey
	if newKey, err = generateAndSetNewDataKey(ctx, keyRegistry); err != nil {
		__antithesis_instrumentation__.Notify(24836)
		return
	} else {
		__antithesis_instrumentation__.Notify(24837)
	}
	__antithesis_instrumentation__.Notify(24825)
	if err = validateRegistry(keyRegistry); err != nil {
		__antithesis_instrumentation__.Notify(24838)
		return
	} else {
		__antithesis_instrumentation__.Notify(24839)
	}
	__antithesis_instrumentation__.Notify(24826)
	bytes, err := protoutil.Marshal(keyRegistry)
	if err != nil {
		__antithesis_instrumentation__.Notify(24840)
		return err
	} else {
		__antithesis_instrumentation__.Notify(24841)
	}
	__antithesis_instrumentation__.Notify(24827)

	filename := fmt.Sprintf("%s_%06d_%s", keyRegistryFilename, m.mu.marker.NextIter(), registryFormatMonolith)
	f, err := m.fs.Create(m.fs.PathJoin(m.dbDir, filename))
	if err != nil {
		__antithesis_instrumentation__.Notify(24842)
		return err
	} else {
		__antithesis_instrumentation__.Notify(24843)
	}
	__antithesis_instrumentation__.Notify(24828)
	defer f.Close()
	if _, err := f.Write(bytes); err != nil {
		__antithesis_instrumentation__.Notify(24844)
		return err
	} else {
		__antithesis_instrumentation__.Notify(24845)
	}
	__antithesis_instrumentation__.Notify(24829)
	if err := f.Sync(); err != nil {
		__antithesis_instrumentation__.Notify(24846)
		return err
	} else {
		__antithesis_instrumentation__.Notify(24847)
	}
	__antithesis_instrumentation__.Notify(24830)

	if err := m.mu.marker.Move(filename); err != nil {
		__antithesis_instrumentation__.Notify(24848)
		return err
	} else {
		__antithesis_instrumentation__.Notify(24849)
	}
	__antithesis_instrumentation__.Notify(24831)

	prevFilename := m.mu.filename
	m.mu.filename = filename
	m.mu.keyRegistry = keyRegistry
	m.mu.activeKey = newKey

	if prevFilename != "" {
		__antithesis_instrumentation__.Notify(24850)
		path := m.fs.PathJoin(m.dbDir, prevFilename)
		if err := m.fs.Remove(path); err != nil && func() bool {
			__antithesis_instrumentation__.Notify(24851)
			return !oserror.IsNotExist(err) == true
		}() == true {
			__antithesis_instrumentation__.Notify(24852)
			return err
		} else {
			__antithesis_instrumentation__.Notify(24853)
		}
	} else {
		__antithesis_instrumentation__.Notify(24854)
	}
	__antithesis_instrumentation__.Notify(24832)
	return err
}
