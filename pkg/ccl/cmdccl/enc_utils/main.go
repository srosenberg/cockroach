package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"crypto/aes"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/cockroachdb/cockroach/pkg/ccl/storageccl/engineccl/enginepbccl"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
)

const fileRegistryPath = "COCKROACHDB_REGISTRY"
const keyRegistryPath = "COCKROACHDB_DATA_KEYS"
const currentPath = "CURRENT"
const optionsPathGlob = "OPTIONS-*"

var dbDir = flag.String("db-dir", "", "path to the db directory")
var storeKeyPath = flag.String("store-key", "", "path to the active store key")

type fileEntry struct {
	envType  enginepb.EnvType
	settings enginepbccl.EncryptionSettings
}

func (f fileEntry) String() string {
	__antithesis_instrumentation__.Notify(19434)
	ret := fmt.Sprintf(" env type: %d, %s\n",
		f.envType, f.settings.EncryptionType)
	if f.settings.EncryptionType != enginepbccl.EncryptionType_Plaintext {
		__antithesis_instrumentation__.Notify(19436)
		ret += fmt.Sprintf("  keyID: %s\n  nonce: % x\n  counter: %d\n",
			f.settings.KeyId,
			f.settings.Nonce,
			f.settings.Counter)
	} else {
		__antithesis_instrumentation__.Notify(19437)
	}
	__antithesis_instrumentation__.Notify(19435)
	return ret
}

type keyEntry struct {
	encryptionType enginepbccl.EncryptionType
	rawKey         []byte
}

func (k keyEntry) String() string {
	__antithesis_instrumentation__.Notify(19438)
	return fmt.Sprintf("%s len: %d", k.encryptionType, len(k.rawKey))
}

var fileRegistry = map[string]fileEntry{}
var keyRegistry = map[string]keyEntry{}

func loadFileRegistry() {
	__antithesis_instrumentation__.Notify(19439)
	data, err := ioutil.ReadFile(filepath.Join(*dbDir, fileRegistryPath))
	if err != nil {
		__antithesis_instrumentation__.Notify(19442)
		log.Fatalf(context.Background(), "could not read %s: %v", fileRegistryPath, err)
	} else {
		__antithesis_instrumentation__.Notify(19443)
	}
	__antithesis_instrumentation__.Notify(19440)

	var reg enginepb.FileRegistry
	if err := protoutil.Unmarshal(data, &reg); err != nil {
		__antithesis_instrumentation__.Notify(19444)
		log.Fatalf(context.Background(), "could not unmarshal %s: %v", fileRegistryPath, err)
	} else {
		__antithesis_instrumentation__.Notify(19445)
	}
	__antithesis_instrumentation__.Notify(19441)

	log.Infof(context.Background(), "file registry version: %s", reg.Version)
	log.Infof(context.Background(), "file registry contains %d entries", len(reg.Files))
	for name, entry := range reg.Files {
		__antithesis_instrumentation__.Notify(19446)
		var encSettings enginepbccl.EncryptionSettings
		settings := entry.EncryptionSettings
		if err := protoutil.Unmarshal(settings, &encSettings); err != nil {
			__antithesis_instrumentation__.Notify(19448)
			log.Fatalf(context.Background(), "could not unmarshal encryption setting for %s: %v", name, err)
		} else {
			__antithesis_instrumentation__.Notify(19449)
		}
		__antithesis_instrumentation__.Notify(19447)

		fileRegistry[name] = fileEntry{entry.EnvType, encSettings}

		log.Infof(context.Background(), "  %-30s level: %-8s type: %-12s keyID: %s", name, entry.EnvType, encSettings.EncryptionType, encSettings.KeyId[:8])
	}
}

func loadStoreKey() {
	__antithesis_instrumentation__.Notify(19450)
	if len(*storeKeyPath) == 0 || func() bool {
		__antithesis_instrumentation__.Notify(19454)
		return *storeKeyPath == "plain" == true
	}() == true {
		__antithesis_instrumentation__.Notify(19455)
		log.Infof(context.Background(), "no store key specified")
		return
	} else {
		__antithesis_instrumentation__.Notify(19456)
	}
	__antithesis_instrumentation__.Notify(19451)

	data, err := ioutil.ReadFile(*storeKeyPath)
	if err != nil {
		__antithesis_instrumentation__.Notify(19457)
		log.Fatalf(context.Background(), "could not read %s: %v", *storeKeyPath, err)
	} else {
		__antithesis_instrumentation__.Notify(19458)
	}
	__antithesis_instrumentation__.Notify(19452)

	var k keyEntry
	switch len(data) {
	case 48:
		__antithesis_instrumentation__.Notify(19459)
		k.encryptionType = enginepbccl.EncryptionType_AES128_CTR
	case 56:
		__antithesis_instrumentation__.Notify(19460)
		k.encryptionType = enginepbccl.EncryptionType_AES192_CTR
	case 64:
		__antithesis_instrumentation__.Notify(19461)
		k.encryptionType = enginepbccl.EncryptionType_AES256_CTR
	default:
		__antithesis_instrumentation__.Notify(19462)
		log.Fatalf(context.Background(), "wrong key length %d, want 32 bytes + AES length", len(data))
	}
	__antithesis_instrumentation__.Notify(19453)

	id := hex.EncodeToString(data[0:32])

	k.rawKey = data[32:]

	keyRegistry[id] = k

	log.Infof(context.Background(), "store key: %s", k)
}

func loadKeyRegistry() {
	__antithesis_instrumentation__.Notify(19463)
	data, err := readFile(keyRegistryPath)
	if err != nil {
		__antithesis_instrumentation__.Notify(19468)
		log.Fatalf(context.Background(), "could not read %s: %v", keyRegistryPath, err)
	} else {
		__antithesis_instrumentation__.Notify(19469)
	}
	__antithesis_instrumentation__.Notify(19464)

	var reg enginepbccl.DataKeysRegistry
	if err := protoutil.Unmarshal(data, &reg); err != nil {
		__antithesis_instrumentation__.Notify(19470)
		log.Fatalf(context.Background(), "could not unmarshal %s: %v", keyRegistryPath, err)
	} else {
		__antithesis_instrumentation__.Notify(19471)
	}
	__antithesis_instrumentation__.Notify(19465)

	log.Infof(context.Background(), "key registry contains %d store keys(s) and %d data key(s)",
		len(reg.StoreKeys), len(reg.DataKeys))
	for _, e := range reg.StoreKeys {
		__antithesis_instrumentation__.Notify(19472)
		log.Infof(context.Background(), "  store key: type: %-12s %v", e.EncryptionType, e)
	}
	__antithesis_instrumentation__.Notify(19466)
	for _, e := range reg.DataKeys {
		__antithesis_instrumentation__.Notify(19473)
		log.Infof(context.Background(), "  data  key: type: %-12s %v", e.Info.EncryptionType, e.Info)
	}
	__antithesis_instrumentation__.Notify(19467)
	for k, e := range reg.DataKeys {
		__antithesis_instrumentation__.Notify(19474)
		keyRegistry[k] = keyEntry{e.Info.EncryptionType, e.Key}
	}
}

func loadCurrent() {
	__antithesis_instrumentation__.Notify(19475)
	data, err := readFile(currentPath)
	if err != nil {
		__antithesis_instrumentation__.Notify(19477)
		log.Fatalf(context.Background(), "could not read %s: %v", currentPath, err)
	} else {
		__antithesis_instrumentation__.Notify(19478)
	}
	__antithesis_instrumentation__.Notify(19476)

	log.Infof(context.Background(), "current: %s", string(data))
}

func loadOptions() {
	__antithesis_instrumentation__.Notify(19479)
	absGlob := filepath.Join(*dbDir, optionsPathGlob)
	paths, err := filepath.Glob(absGlob)
	if err != nil {
		__antithesis_instrumentation__.Notify(19481)
		log.Fatalf(context.Background(), "problem finding files matching %s: %v", absGlob, err)
	} else {
		__antithesis_instrumentation__.Notify(19482)
	}
	__antithesis_instrumentation__.Notify(19480)

	for _, f := range paths {
		__antithesis_instrumentation__.Notify(19483)
		fname := filepath.Base(f)
		data, err := readFile(fname)
		if err != nil {
			__antithesis_instrumentation__.Notify(19485)
			log.Fatalf(context.Background(), "could not read %s: %v", fname, err)
		} else {
			__antithesis_instrumentation__.Notify(19486)
		}
		__antithesis_instrumentation__.Notify(19484)
		log.Infof(context.Background(), "options file: %s starts with: %s", fname, string(data[:100]))
	}
}

func readFile(filename string) ([]byte, error) {
	__antithesis_instrumentation__.Notify(19487)
	if len(filename) == 0 {
		__antithesis_instrumentation__.Notify(19496)
		return nil, errors.Errorf("filename is empty")
	} else {
		__antithesis_instrumentation__.Notify(19497)
	}
	__antithesis_instrumentation__.Notify(19488)

	absPath := filename
	if filename[0] != '/' {
		__antithesis_instrumentation__.Notify(19498)
		absPath = filepath.Join(*dbDir, filename)
	} else {
		__antithesis_instrumentation__.Notify(19499)
	}
	__antithesis_instrumentation__.Notify(19489)

	data, err := ioutil.ReadFile(absPath)
	if err != nil {
		__antithesis_instrumentation__.Notify(19500)
		return nil, errors.Wrapf(err, "could not read %s", absPath)
	} else {
		__antithesis_instrumentation__.Notify(19501)
	}
	__antithesis_instrumentation__.Notify(19490)

	reg, ok := fileRegistry[filename]
	if !ok || func() bool {
		__antithesis_instrumentation__.Notify(19502)
		return reg.settings.EncryptionType == enginepbccl.EncryptionType_Plaintext == true
	}() == true {
		__antithesis_instrumentation__.Notify(19503)

		log.Infof(context.Background(), "reading plaintext %s", absPath)
		return data, nil
	} else {
		__antithesis_instrumentation__.Notify(19504)
	}
	__antithesis_instrumentation__.Notify(19491)

	key, ok := keyRegistry[reg.settings.KeyId]
	if !ok {
		__antithesis_instrumentation__.Notify(19505)
		return nil, errors.Errorf("could not find key %s for file %s", reg.settings.KeyId, absPath)
	} else {
		__antithesis_instrumentation__.Notify(19506)
	}
	__antithesis_instrumentation__.Notify(19492)
	log.Infof(context.Background(), "decrypting %s with %s key %s...", filename, reg.settings.EncryptionType, reg.settings.KeyId[:8])

	cipher, err := aes.NewCipher(key.rawKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(19507)
		return nil, errors.Wrapf(err, "could not build AES cipher for file %s", absPath)
	} else {
		__antithesis_instrumentation__.Notify(19508)
	}
	__antithesis_instrumentation__.Notify(19493)

	size := len(data)
	counter := reg.settings.Counter
	nonce := reg.settings.Nonce
	if len(nonce) != 12 {
		__antithesis_instrumentation__.Notify(19509)
		log.Fatalf(context.Background(), "nonce has wrong length: %d, expected 12", len(nonce))
	} else {
		__antithesis_instrumentation__.Notify(19510)
	}
	__antithesis_instrumentation__.Notify(19494)

	iv := make([]byte, aes.BlockSize)
	for offset := 0; offset < size; offset += aes.BlockSize {
		__antithesis_instrumentation__.Notify(19511)

		copy(iv, nonce)

		binary.BigEndian.PutUint32(iv[12:], counter)

		counter++

		cipher.Encrypt(iv, iv)

		for i := 0; i < aes.BlockSize; i++ {
			__antithesis_instrumentation__.Notify(19512)
			pos := offset + i
			if pos >= size {
				__antithesis_instrumentation__.Notify(19514)

				break
			} else {
				__antithesis_instrumentation__.Notify(19515)
			}
			__antithesis_instrumentation__.Notify(19513)
			data[pos] = data[pos] ^ iv[i]
		}
	}
	__antithesis_instrumentation__.Notify(19495)

	return data, nil
}

func main() {
	__antithesis_instrumentation__.Notify(19516)
	flag.Parse()
	loadStoreKey()
	loadFileRegistry()
	loadKeyRegistry()
	loadCurrent()
	loadOptions()
}
