package baseccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/ccl/cliccl/cliflagsccl"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
	"github.com/spf13/pflag"
)

const DefaultRotationPeriod = time.Hour * 24 * 7

const plaintextFieldValue = "plain"

type StoreEncryptionSpec struct {
	Path           string
	KeyPath        string
	OldKeyPath     string
	RotationPeriod time.Duration
}

func (es StoreEncryptionSpec) toEncryptionOptions() ([]byte, error) {
	__antithesis_instrumentation__.Notify(14008)
	opts := EncryptionOptions{
		KeySource: EncryptionKeySource_KeyFiles,
		KeyFiles: &EncryptionKeyFiles{
			CurrentKey: es.KeyPath,
			OldKey:     es.OldKeyPath,
		},
		DataKeyRotationPeriod: int64(es.RotationPeriod / time.Second),
	}

	return protoutil.Marshal(&opts)
}

func (es StoreEncryptionSpec) String() string {
	__antithesis_instrumentation__.Notify(14009)

	return fmt.Sprintf("path=%s,key=%s,old-key=%s,rotation-period=%s",
		es.Path, es.KeyPath, es.OldKeyPath, es.RotationPeriod)
}

func NewStoreEncryptionSpec(value string) (StoreEncryptionSpec, error) {
	__antithesis_instrumentation__.Notify(14010)
	const pathField = "path"
	var es StoreEncryptionSpec
	es.RotationPeriod = DefaultRotationPeriod

	used := make(map[string]struct{})
	for _, split := range strings.Split(value, ",") {
		__antithesis_instrumentation__.Notify(14015)
		if len(split) == 0 {
			__antithesis_instrumentation__.Notify(14021)
			continue
		} else {
			__antithesis_instrumentation__.Notify(14022)
		}
		__antithesis_instrumentation__.Notify(14016)
		subSplits := strings.SplitN(split, "=", 2)
		if len(subSplits) == 1 {
			__antithesis_instrumentation__.Notify(14023)
			return StoreEncryptionSpec{}, fmt.Errorf("field not in the form <key>=<value>: %s", split)
		} else {
			__antithesis_instrumentation__.Notify(14024)
		}
		__antithesis_instrumentation__.Notify(14017)
		field := strings.ToLower(subSplits[0])
		value := subSplits[1]
		if _, ok := used[field]; ok {
			__antithesis_instrumentation__.Notify(14025)
			return StoreEncryptionSpec{}, fmt.Errorf("%s field was used twice in encryption definition", field)
		} else {
			__antithesis_instrumentation__.Notify(14026)
		}
		__antithesis_instrumentation__.Notify(14018)
		used[field] = struct{}{}

		if len(field) == 0 {
			__antithesis_instrumentation__.Notify(14027)
			return StoreEncryptionSpec{}, fmt.Errorf("empty field")
		} else {
			__antithesis_instrumentation__.Notify(14028)
		}
		__antithesis_instrumentation__.Notify(14019)
		if len(value) == 0 {
			__antithesis_instrumentation__.Notify(14029)
			return StoreEncryptionSpec{}, fmt.Errorf("no value specified for %s", field)
		} else {
			__antithesis_instrumentation__.Notify(14030)
		}
		__antithesis_instrumentation__.Notify(14020)

		switch field {
		case pathField:
			__antithesis_instrumentation__.Notify(14031)
			var err error
			es.Path, err = base.GetAbsoluteStorePath(pathField, value)
			if err != nil {
				__antithesis_instrumentation__.Notify(14036)
				return StoreEncryptionSpec{}, err
			} else {
				__antithesis_instrumentation__.Notify(14037)
			}
		case "key":
			__antithesis_instrumentation__.Notify(14032)
			if value == plaintextFieldValue {
				__antithesis_instrumentation__.Notify(14038)
				es.KeyPath = plaintextFieldValue
			} else {
				__antithesis_instrumentation__.Notify(14039)
				var err error
				es.KeyPath, err = base.GetAbsoluteStorePath("key", value)
				if err != nil {
					__antithesis_instrumentation__.Notify(14040)
					return StoreEncryptionSpec{}, err
				} else {
					__antithesis_instrumentation__.Notify(14041)
				}
			}
		case "old-key":
			__antithesis_instrumentation__.Notify(14033)
			if value == plaintextFieldValue {
				__antithesis_instrumentation__.Notify(14042)
				es.OldKeyPath = plaintextFieldValue
			} else {
				__antithesis_instrumentation__.Notify(14043)
				var err error
				es.OldKeyPath, err = base.GetAbsoluteStorePath("old-key", value)
				if err != nil {
					__antithesis_instrumentation__.Notify(14044)
					return StoreEncryptionSpec{}, err
				} else {
					__antithesis_instrumentation__.Notify(14045)
				}
			}
		case "rotation-period":
			__antithesis_instrumentation__.Notify(14034)
			var err error
			es.RotationPeriod, err = time.ParseDuration(value)
			if err != nil {
				__antithesis_instrumentation__.Notify(14046)
				return StoreEncryptionSpec{}, errors.Wrapf(err, "could not parse rotation-duration value: %s", value)
			} else {
				__antithesis_instrumentation__.Notify(14047)
			}
		default:
			__antithesis_instrumentation__.Notify(14035)
			return StoreEncryptionSpec{}, fmt.Errorf("%s is not a valid enterprise-encryption field", field)
		}
	}
	__antithesis_instrumentation__.Notify(14011)

	if es.Path == "" {
		__antithesis_instrumentation__.Notify(14048)
		return StoreEncryptionSpec{}, fmt.Errorf("no path specified")
	} else {
		__antithesis_instrumentation__.Notify(14049)
	}
	__antithesis_instrumentation__.Notify(14012)
	if es.KeyPath == "" {
		__antithesis_instrumentation__.Notify(14050)
		return StoreEncryptionSpec{}, fmt.Errorf("no key specified")
	} else {
		__antithesis_instrumentation__.Notify(14051)
	}
	__antithesis_instrumentation__.Notify(14013)
	if es.OldKeyPath == "" {
		__antithesis_instrumentation__.Notify(14052)
		return StoreEncryptionSpec{}, fmt.Errorf("no old-key specified")
	} else {
		__antithesis_instrumentation__.Notify(14053)
	}
	__antithesis_instrumentation__.Notify(14014)

	return es, nil
}

type StoreEncryptionSpecList struct {
	Specs []StoreEncryptionSpec
}

var _ pflag.Value = &StoreEncryptionSpecList{}

func (encl StoreEncryptionSpecList) String() string {
	__antithesis_instrumentation__.Notify(14054)
	var buffer bytes.Buffer
	for _, ss := range encl.Specs {
		__antithesis_instrumentation__.Notify(14057)
		fmt.Fprintf(&buffer, "--%s=%s ", cliflagsccl.EnterpriseEncryption.Name, ss)
	}
	__antithesis_instrumentation__.Notify(14055)

	if l := buffer.Len(); l > 0 {
		__antithesis_instrumentation__.Notify(14058)
		buffer.Truncate(l - 1)
	} else {
		__antithesis_instrumentation__.Notify(14059)
	}
	__antithesis_instrumentation__.Notify(14056)
	return buffer.String()
}

func (encl *StoreEncryptionSpecList) Type() string {
	__antithesis_instrumentation__.Notify(14060)
	return "StoreEncryptionSpec"
}

func (encl *StoreEncryptionSpecList) Set(value string) error {
	__antithesis_instrumentation__.Notify(14061)
	spec, err := NewStoreEncryptionSpec(value)
	if err != nil {
		__antithesis_instrumentation__.Notify(14064)
		return err
	} else {
		__antithesis_instrumentation__.Notify(14065)
	}
	__antithesis_instrumentation__.Notify(14062)
	if encl.Specs == nil {
		__antithesis_instrumentation__.Notify(14066)
		encl.Specs = []StoreEncryptionSpec{spec}
	} else {
		__antithesis_instrumentation__.Notify(14067)
		encl.Specs = append(encl.Specs, spec)
	}
	__antithesis_instrumentation__.Notify(14063)
	return nil
}

func PopulateStoreSpecWithEncryption(
	storeSpecs base.StoreSpecList, encryptionSpecs StoreEncryptionSpecList,
) error {
	__antithesis_instrumentation__.Notify(14068)
	for _, es := range encryptionSpecs.Specs {
		__antithesis_instrumentation__.Notify(14070)
		var found bool
		for i := range storeSpecs.Specs {
			__antithesis_instrumentation__.Notify(14072)
			if storeSpecs.Specs[i].Path != es.Path {
				__antithesis_instrumentation__.Notify(14076)
				continue
			} else {
				__antithesis_instrumentation__.Notify(14077)
			}
			__antithesis_instrumentation__.Notify(14073)

			if storeSpecs.Specs[i].UseFileRegistry {
				__antithesis_instrumentation__.Notify(14078)
				return fmt.Errorf("store with path %s already has an encryption setting",
					storeSpecs.Specs[i].Path)
			} else {
				__antithesis_instrumentation__.Notify(14079)
			}
			__antithesis_instrumentation__.Notify(14074)

			storeSpecs.Specs[i].UseFileRegistry = true
			opts, err := es.toEncryptionOptions()
			if err != nil {
				__antithesis_instrumentation__.Notify(14080)
				return err
			} else {
				__antithesis_instrumentation__.Notify(14081)
			}
			__antithesis_instrumentation__.Notify(14075)
			storeSpecs.Specs[i].EncryptionOptions = opts
			found = true
			break
		}
		__antithesis_instrumentation__.Notify(14071)
		if !found {
			__antithesis_instrumentation__.Notify(14082)
			return fmt.Errorf("no store with path %s found for encryption setting: %v", es.Path, es)
		} else {
			__antithesis_instrumentation__.Notify(14083)
		}
	}
	__antithesis_instrumentation__.Notify(14069)
	return nil
}

func EncryptionOptionsForStore(
	dir string, encryptionSpecs StoreEncryptionSpecList,
) ([]byte, error) {
	__antithesis_instrumentation__.Notify(14084)

	path, err := filepath.Abs(dir)
	if err != nil {
		__antithesis_instrumentation__.Notify(14087)
		return nil, errors.Wrapf(err, "could not find absolute path for %s ", dir)
	} else {
		__antithesis_instrumentation__.Notify(14088)
	}
	__antithesis_instrumentation__.Notify(14085)

	for _, es := range encryptionSpecs.Specs {
		__antithesis_instrumentation__.Notify(14089)
		if es.Path == path {
			__antithesis_instrumentation__.Notify(14090)
			return es.toEncryptionOptions()
		} else {
			__antithesis_instrumentation__.Notify(14091)
		}
	}
	__antithesis_instrumentation__.Notify(14086)

	return nil, nil
}
