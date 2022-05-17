package cliccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/ccl/baseccl"
	"github.com/cockroachdb/cockroach/pkg/ccl/cliccl/cliflagsccl"
	"github.com/cockroachdb/cockroach/pkg/ccl/storageccl/engineccl/enginepbccl"
	"github.com/cockroachdb/cockroach/pkg/cli"
	"github.com/cockroachdb/cockroach/pkg/cli/clierrorplus"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/oserror"
	"github.com/spf13/cobra"
)

const (
	plaintextKeyID       = "plain"
	keyRegistryFilename  = "COCKROACHDB_DATA_KEYS"
	fileRegistryFilename = "COCKROACHDB_REGISTRY"
)

var encryptionStatusOpts struct {
	activeStoreIDOnly bool
}

func init() {
	encryptionStatusCmd := &cobra.Command{
		Use:   "encryption-status <directory>",
		Short: "show encryption status of a store",
		Long: `
Shows encryption status of the store located in 'directory'.
Encryption keys must be specified in the '--enterprise-encryption' flag.

Displays all store and data keys as well as files encrypted with each.
Specifying --active-store-key-id-only prints the key ID of the active store key
and exits.
`,
		Args: cobra.ExactArgs(1),
		RunE: clierrorplus.MaybeDecorateError(runEncryptionStatus),
	}

	encryptionActiveKeyCmd := &cobra.Command{
		Use:   "encryption-active-key <directory>",
		Short: "return ID of the active store key",
		Long: `
Display the algorithm and key ID of the active store key for existing data directory 'directory'.
Does not require knowing the key.

Some sample outputs:
Plaintext:            # encryption not enabled
AES128_CTR:be235...   # AES-128 encryption with store key ID
`,
		Args: cobra.ExactArgs(1),
		RunE: clierrorplus.MaybeDecorateError(runEncryptionActiveKey),
	}

	cli.DebugCmd.AddCommand(encryptionStatusCmd)
	cli.DebugCmd.AddCommand(encryptionActiveKeyCmd)

	f := encryptionStatusCmd.Flags()
	cli.VarFlag(f, &storeEncryptionSpecs, cliflagsccl.EnterpriseEncryption)

	f.BoolVar(&encryptionStatusOpts.activeStoreIDOnly, "active-store-key-id-only", false,
		"print active store key ID and exit")

	for _, cmd := range cli.DebugCommandsRequiringEncryption {

		cli.VarFlag(cmd.Flags(), &storeEncryptionSpecs, cliflagsccl.EnterpriseEncryption)
	}

	cli.VarFlag(cli.DebugPebbleCmd.PersistentFlags(),
		&storeEncryptionSpecs, cliflagsccl.EnterpriseEncryption)

	cli.PopulateStorageConfigHook = fillEncryptionOptionsForStore
}

func fillEncryptionOptionsForStore(cfg *base.StorageConfig) error {
	__antithesis_instrumentation__.Notify(19070)
	opts, err := baseccl.EncryptionOptionsForStore(cfg.Dir, storeEncryptionSpecs)
	if err != nil {
		__antithesis_instrumentation__.Notify(19073)
		return err
	} else {
		__antithesis_instrumentation__.Notify(19074)
	}
	__antithesis_instrumentation__.Notify(19071)

	if opts != nil {
		__antithesis_instrumentation__.Notify(19075)
		cfg.EncryptionOptions = opts
		cfg.UseFileRegistry = true
	} else {
		__antithesis_instrumentation__.Notify(19076)
	}
	__antithesis_instrumentation__.Notify(19072)
	return nil
}

type keyInfoByAge []*enginepbccl.KeyInfo

func (ki keyInfoByAge) Len() int { __antithesis_instrumentation__.Notify(19077); return len(ki) }
func (ki keyInfoByAge) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(19078)
	ki[i], ki[j] = ki[j], ki[i]
}
func (ki keyInfoByAge) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(19079)
	return ki[i].CreationTime < ki[j].CreationTime
}

type JSONTime time.Time

func (t JSONTime) MarshalJSON() ([]byte, error) {
	__antithesis_instrumentation__.Notify(19080)
	return []byte(fmt.Sprintf("\"%s\"", time.Time(t).String())), nil
}

type PrettyDataKey struct {
	ID      string
	Active  bool `json:",omitempty"`
	Exposed bool `json:",omitempty"`
	Created JSONTime
	Files   []string `json:",omitempty"`
}

type PrettyStoreKey struct {
	ID       string
	Active   bool `json:",omitempty"`
	Type     string
	Created  JSONTime
	Source   string
	Files    []string        `json:",omitempty"`
	DataKeys []PrettyDataKey `json:",omitempty"`
}

func runEncryptionStatus(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(19081)
	stopper := stop.NewStopper()
	defer stopper.Stop(context.Background())

	dir := args[0]

	db, err := cli.OpenExistingStore(dir, stopper, true, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(19095)
		return err
	} else {
		__antithesis_instrumentation__.Notify(19096)
	}
	__antithesis_instrumentation__.Notify(19082)

	registries, err := db.GetEncryptionRegistries()
	if err != nil {
		__antithesis_instrumentation__.Notify(19097)
		return err
	} else {
		__antithesis_instrumentation__.Notify(19098)
	}
	__antithesis_instrumentation__.Notify(19083)

	if len(registries.KeyRegistry) == 0 {
		__antithesis_instrumentation__.Notify(19099)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(19100)
	}
	__antithesis_instrumentation__.Notify(19084)

	var fileRegistry enginepb.FileRegistry
	if err := protoutil.Unmarshal(registries.FileRegistry, &fileRegistry); err != nil {
		__antithesis_instrumentation__.Notify(19101)
		return err
	} else {
		__antithesis_instrumentation__.Notify(19102)
	}
	__antithesis_instrumentation__.Notify(19085)

	var keyRegistry enginepbccl.DataKeysRegistry
	if err := protoutil.Unmarshal(registries.KeyRegistry, &keyRegistry); err != nil {
		__antithesis_instrumentation__.Notify(19103)
		return err
	} else {
		__antithesis_instrumentation__.Notify(19104)
	}
	__antithesis_instrumentation__.Notify(19086)

	if encryptionStatusOpts.activeStoreIDOnly {
		__antithesis_instrumentation__.Notify(19105)
		fmt.Println(keyRegistry.ActiveStoreKeyId)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(19106)
	}
	__antithesis_instrumentation__.Notify(19087)

	fileKeyMap := make(map[string][]string)

	for name, entry := range fileRegistry.Files {
		__antithesis_instrumentation__.Notify(19107)
		keyID := plaintextKeyID

		if entry.EnvType != enginepb.EnvType_Plaintext && func() bool {
			__antithesis_instrumentation__.Notify(19109)
			return len(entry.EncryptionSettings) > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(19110)
			var setting enginepbccl.EncryptionSettings
			if err := protoutil.Unmarshal(entry.EncryptionSettings, &setting); err != nil {
				__antithesis_instrumentation__.Notify(19112)
				fmt.Fprintf(os.Stderr, "could not unmarshal encryption settings for file %s: %v", name, err)
				continue
			} else {
				__antithesis_instrumentation__.Notify(19113)
			}
			__antithesis_instrumentation__.Notify(19111)
			keyID = setting.KeyId
		} else {
			__antithesis_instrumentation__.Notify(19114)
		}
		__antithesis_instrumentation__.Notify(19108)

		fileKeyMap[keyID] = append(fileKeyMap[keyID], name)
	}
	__antithesis_instrumentation__.Notify(19088)

	childKeyMap := make(map[string]keyInfoByAge)

	for _, dataKey := range keyRegistry.DataKeys {
		__antithesis_instrumentation__.Notify(19115)
		info := dataKey.Info
		parentKey := plaintextKeyID
		if len(info.ParentKeyId) > 0 {
			__antithesis_instrumentation__.Notify(19117)
			parentKey = info.ParentKeyId
		} else {
			__antithesis_instrumentation__.Notify(19118)
		}
		__antithesis_instrumentation__.Notify(19116)
		childKeyMap[parentKey] = append(childKeyMap[parentKey], info)
	}
	__antithesis_instrumentation__.Notify(19089)

	storeKeyList := make(keyInfoByAge, 0)
	for _, ki := range keyRegistry.StoreKeys {
		__antithesis_instrumentation__.Notify(19119)
		storeKeyList = append(storeKeyList, ki)
	}
	__antithesis_instrumentation__.Notify(19090)

	storeKeys := make([]PrettyStoreKey, 0, len(storeKeyList))
	sort.Sort(storeKeyList)
	for _, storeKey := range storeKeyList {
		__antithesis_instrumentation__.Notify(19120)
		storeNode := PrettyStoreKey{
			ID:      storeKey.KeyId,
			Active:  (storeKey.KeyId == keyRegistry.ActiveStoreKeyId),
			Type:    storeKey.EncryptionType.String(),
			Created: JSONTime(timeutil.Unix(storeKey.CreationTime, 0)),
			Source:  storeKey.Source,
		}

		if files, ok := fileKeyMap[storeKey.KeyId]; ok {
			__antithesis_instrumentation__.Notify(19123)
			sort.Strings(files)
			storeNode.Files = files
			delete(fileKeyMap, storeKey.KeyId)
		} else {
			__antithesis_instrumentation__.Notify(19124)
		}
		__antithesis_instrumentation__.Notify(19121)

		if children, ok := childKeyMap[storeKey.KeyId]; ok {
			__antithesis_instrumentation__.Notify(19125)
			storeNode.DataKeys = make([]PrettyDataKey, 0, len(children))

			sort.Sort(children)
			for _, c := range children {
				__antithesis_instrumentation__.Notify(19127)
				dataNode := PrettyDataKey{
					ID:      c.KeyId,
					Active:  (c.KeyId == keyRegistry.ActiveDataKeyId),
					Exposed: c.WasExposed,
					Created: JSONTime(timeutil.Unix(c.CreationTime, 0)),
				}
				files, ok := fileKeyMap[c.KeyId]
				if ok {
					__antithesis_instrumentation__.Notify(19129)
					sort.Strings(files)
					dataNode.Files = files
					delete(fileKeyMap, c.KeyId)
				} else {
					__antithesis_instrumentation__.Notify(19130)
				}
				__antithesis_instrumentation__.Notify(19128)
				storeNode.DataKeys = append(storeNode.DataKeys, dataNode)
			}
			__antithesis_instrumentation__.Notify(19126)
			delete(childKeyMap, storeKey.KeyId)
		} else {
			__antithesis_instrumentation__.Notify(19131)
		}
		__antithesis_instrumentation__.Notify(19122)
		storeKeys = append(storeKeys, storeNode)
	}
	__antithesis_instrumentation__.Notify(19091)

	j, err := json.MarshalIndent(storeKeys, "", "  ")
	if err != nil {
		__antithesis_instrumentation__.Notify(19132)
		return err
	} else {
		__antithesis_instrumentation__.Notify(19133)
	}
	__antithesis_instrumentation__.Notify(19092)
	fmt.Printf("%s\n", j)

	if len(fileKeyMap) > 0 {
		__antithesis_instrumentation__.Notify(19134)
		fmt.Fprintf(os.Stderr, "WARNING: could not find key info for some files: %+v\n", fileKeyMap)
	} else {
		__antithesis_instrumentation__.Notify(19135)
	}
	__antithesis_instrumentation__.Notify(19093)
	if len(childKeyMap) > 0 {
		__antithesis_instrumentation__.Notify(19136)
		fmt.Fprintf(os.Stderr, "WARNING: could not find parent key info for some data keys: %+v\n", childKeyMap)
	} else {
		__antithesis_instrumentation__.Notify(19137)
	}
	__antithesis_instrumentation__.Notify(19094)

	return nil
}

func runEncryptionActiveKey(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(19138)
	keyType, keyID, err := getActiveEncryptionkey(args[0])
	if err != nil {
		__antithesis_instrumentation__.Notify(19140)
		return err
	} else {
		__antithesis_instrumentation__.Notify(19141)
	}
	__antithesis_instrumentation__.Notify(19139)

	fmt.Printf("%s:%s\n", keyType, keyID)
	return nil
}

func getActiveEncryptionkey(dir string) (string, string, error) {
	__antithesis_instrumentation__.Notify(19142)
	registryFile := filepath.Join(dir, fileRegistryFilename)

	if _, err := os.Stat(dir); err != nil {
		__antithesis_instrumentation__.Notify(19149)
		return "", "", errors.Wrapf(err, "data directory %s does not exist", dir)
	} else {
		__antithesis_instrumentation__.Notify(19150)
	}
	__antithesis_instrumentation__.Notify(19143)

	contents, err := ioutil.ReadFile(registryFile)
	if err != nil {
		__antithesis_instrumentation__.Notify(19151)
		if oserror.IsNotExist(err) {
			__antithesis_instrumentation__.Notify(19153)
			return enginepbccl.EncryptionType_Plaintext.String(), "", nil
		} else {
			__antithesis_instrumentation__.Notify(19154)
		}
		__antithesis_instrumentation__.Notify(19152)
		return "", "", errors.Wrapf(err, "could not open registry file %s", registryFile)
	} else {
		__antithesis_instrumentation__.Notify(19155)
	}
	__antithesis_instrumentation__.Notify(19144)

	var fileRegistry enginepb.FileRegistry
	if err := protoutil.Unmarshal(contents, &fileRegistry); err != nil {
		__antithesis_instrumentation__.Notify(19156)
		return "", "", err
	} else {
		__antithesis_instrumentation__.Notify(19157)
	}
	__antithesis_instrumentation__.Notify(19145)

	entry, ok := fileRegistry.Files[keyRegistryFilename]
	if !ok {
		__antithesis_instrumentation__.Notify(19158)
		return "", "", fmt.Errorf("key registry file %s was not found in the file registry", keyRegistryFilename)
	} else {
		__antithesis_instrumentation__.Notify(19159)
	}
	__antithesis_instrumentation__.Notify(19146)

	if entry.EnvType == enginepb.EnvType_Plaintext || func() bool {
		__antithesis_instrumentation__.Notify(19160)
		return len(entry.EncryptionSettings) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(19161)

		return enginepbccl.EncryptionType_Plaintext.String(), "", nil
	} else {
		__antithesis_instrumentation__.Notify(19162)
	}
	__antithesis_instrumentation__.Notify(19147)

	var setting enginepbccl.EncryptionSettings
	if err := protoutil.Unmarshal(entry.EncryptionSettings, &setting); err != nil {
		__antithesis_instrumentation__.Notify(19163)
		return "", "", errors.Wrapf(err, "could not unmarshal encryption settings for %s", keyRegistryFilename)
	} else {
		__antithesis_instrumentation__.Notify(19164)
	}
	__antithesis_instrumentation__.Notify(19148)

	return setting.EncryptionType.String(), setting.KeyId, nil
}
