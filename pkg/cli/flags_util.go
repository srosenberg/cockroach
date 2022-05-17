package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	gohex "encoding/hex"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/status"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/keysutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/pebble/vfs"
	humanize "github.com/dustin/go-humanize"
	"github.com/spf13/pflag"
)

type localityList []roachpb.LocalityAddress

var _ pflag.Value = &localityList{}

func (l *localityList) Type() string {
	__antithesis_instrumentation__.Notify(32833)
	return "localityList"
}

func (l *localityList) String() string {
	__antithesis_instrumentation__.Notify(32834)
	string := ""
	for _, loc := range []roachpb.LocalityAddress(*l) {
		__antithesis_instrumentation__.Notify(32836)
		string += loc.LocalityTier.Key + "=" + loc.LocalityTier.Value + "@" + loc.Address.String() + ","
	}
	__antithesis_instrumentation__.Notify(32835)

	return string
}

func (l *localityList) Set(value string) error {
	__antithesis_instrumentation__.Notify(32837)
	*l = []roachpb.LocalityAddress{}

	values := strings.Split(value, ",")

	for _, value := range values {
		__antithesis_instrumentation__.Notify(32839)
		split := strings.Split(value, "@")
		if len(split) != 2 {
			__antithesis_instrumentation__.Notify(32842)
			return fmt.Errorf("invalid value for --locality-advertise-address: %s", l)
		} else {
			__antithesis_instrumentation__.Notify(32843)
		}
		__antithesis_instrumentation__.Notify(32840)

		tierSplit := strings.Split(split[0], "=")
		if len(tierSplit) != 2 {
			__antithesis_instrumentation__.Notify(32844)
			return fmt.Errorf("invalid value for --locality-advertise-address: %s", l)
		} else {
			__antithesis_instrumentation__.Notify(32845)
		}
		__antithesis_instrumentation__.Notify(32841)

		tier := roachpb.Tier{}
		tier.Key = tierSplit[0]
		tier.Value = tierSplit[1]

		locAddress := roachpb.LocalityAddress{}
		locAddress.LocalityTier = tier
		locAddress.Address = util.MakeUnresolvedAddr("tcp", split[1])

		*l = append(*l, locAddress)
	}
	__antithesis_instrumentation__.Notify(32838)

	return nil
}

type dumpMode int

const (
	dumpBoth dumpMode = iota
	dumpSchemaOnly
	dumpDataOnly
)

func (m *dumpMode) Type() string { __antithesis_instrumentation__.Notify(32846); return "string" }

func (m *dumpMode) String() string {
	__antithesis_instrumentation__.Notify(32847)
	switch *m {
	case dumpBoth:
		__antithesis_instrumentation__.Notify(32849)
		return "both"
	case dumpSchemaOnly:
		__antithesis_instrumentation__.Notify(32850)
		return "schema"
	case dumpDataOnly:
		__antithesis_instrumentation__.Notify(32851)
		return "data"
	default:
		__antithesis_instrumentation__.Notify(32852)
	}
	__antithesis_instrumentation__.Notify(32848)
	return ""
}

func (m *dumpMode) Set(s string) error {
	__antithesis_instrumentation__.Notify(32853)
	switch s {
	case "both":
		__antithesis_instrumentation__.Notify(32855)
		*m = dumpBoth
	case "schema":
		__antithesis_instrumentation__.Notify(32856)
		*m = dumpSchemaOnly
	case "data":
		__antithesis_instrumentation__.Notify(32857)
		*m = dumpDataOnly
	default:
		__antithesis_instrumentation__.Notify(32858)
		return fmt.Errorf("invalid value for --dump-mode: %s", s)
	}
	__antithesis_instrumentation__.Notify(32854)
	return nil
}

type mvccKey storage.MVCCKey

func (k *mvccKey) Type() string {
	__antithesis_instrumentation__.Notify(32859)
	return "engine.MVCCKey"
}

func (k *mvccKey) String() string {
	__antithesis_instrumentation__.Notify(32860)
	return storage.MVCCKey(*k).String()
}

func (k *mvccKey) Set(value string) error {
	__antithesis_instrumentation__.Notify(32861)
	var typ keyType
	var keyStr string
	i := strings.IndexByte(value, ':')
	if i == -1 {
		__antithesis_instrumentation__.Notify(32864)
		keyStr = value
	} else {
		__antithesis_instrumentation__.Notify(32865)
		var err error
		typ, err = parseKeyType(value[:i])
		if err != nil {
			__antithesis_instrumentation__.Notify(32867)
			return err
		} else {
			__antithesis_instrumentation__.Notify(32868)
		}
		__antithesis_instrumentation__.Notify(32866)
		keyStr = value[i+1:]
	}
	__antithesis_instrumentation__.Notify(32862)

	switch typ {
	case hex:
		__antithesis_instrumentation__.Notify(32869)
		b, err := gohex.DecodeString(keyStr)
		if err != nil {
			__antithesis_instrumentation__.Notify(32879)
			return err
		} else {
			__antithesis_instrumentation__.Notify(32880)
		}
		__antithesis_instrumentation__.Notify(32870)
		newK, err := storage.DecodeMVCCKey(b)
		if err != nil {
			__antithesis_instrumentation__.Notify(32881)
			encoded := gohex.EncodeToString(storage.EncodeMVCCKey(storage.MakeMVCCMetadataKey(roachpb.Key(b))))
			return errors.Wrapf(err, "perhaps this is just a hex-encoded key; you need an "+
				"encoded MVCCKey (i.e. with a timestamp component); here's one with a zero timestamp: %s",
				encoded)
		} else {
			__antithesis_instrumentation__.Notify(32882)
		}
		__antithesis_instrumentation__.Notify(32871)
		*k = mvccKey(newK)
	case raw:
		__antithesis_instrumentation__.Notify(32872)
		unquoted, err := unquoteArg(keyStr)
		if err != nil {
			__antithesis_instrumentation__.Notify(32883)
			return err
		} else {
			__antithesis_instrumentation__.Notify(32884)
		}
		__antithesis_instrumentation__.Notify(32873)
		*k = mvccKey(storage.MakeMVCCMetadataKey(roachpb.Key(unquoted)))
	case human:
		__antithesis_instrumentation__.Notify(32874)
		scanner := keysutil.MakePrettyScanner(nil)
		key, err := scanner.Scan(keyStr)
		if err != nil {
			__antithesis_instrumentation__.Notify(32885)
			return err
		} else {
			__antithesis_instrumentation__.Notify(32886)
		}
		__antithesis_instrumentation__.Notify(32875)
		*k = mvccKey(storage.MakeMVCCMetadataKey(key))
	case rangeID:
		__antithesis_instrumentation__.Notify(32876)
		fromID, err := parseRangeID(keyStr)
		if err != nil {
			__antithesis_instrumentation__.Notify(32887)
			return err
		} else {
			__antithesis_instrumentation__.Notify(32888)
		}
		__antithesis_instrumentation__.Notify(32877)
		*k = mvccKey(storage.MakeMVCCMetadataKey(keys.MakeRangeIDPrefix(fromID)))
	default:
		__antithesis_instrumentation__.Notify(32878)
		return fmt.Errorf("unknown key type %s", typ)
	}
	__antithesis_instrumentation__.Notify(32863)

	return nil
}

func unquoteArg(arg string) (string, error) {
	__antithesis_instrumentation__.Notify(32889)
	s, err := strconv.Unquote(`"` + arg + `"`)
	if err != nil {
		__antithesis_instrumentation__.Notify(32891)
		return "", errors.Wrapf(err, "invalid argument %q", arg)
	} else {
		__antithesis_instrumentation__.Notify(32892)
	}
	__antithesis_instrumentation__.Notify(32890)
	return s, nil
}

type keyType int

const (
	raw keyType = iota
	human
	rangeID
	hex
)

var _keyTypes []string

func keyTypes() []string {
	__antithesis_instrumentation__.Notify(32893)
	if _keyTypes == nil {
		__antithesis_instrumentation__.Notify(32895)
		for i := 0; i+1 < len(_keyType_index); i++ {
			__antithesis_instrumentation__.Notify(32896)
			_keyTypes = append(_keyTypes, _keyType_name[_keyType_index[i]:_keyType_index[i+1]])
		}
	} else {
		__antithesis_instrumentation__.Notify(32897)
	}
	__antithesis_instrumentation__.Notify(32894)
	return _keyTypes
}

func parseKeyType(value string) (keyType, error) {
	__antithesis_instrumentation__.Notify(32898)
	for i, typ := range keyTypes() {
		__antithesis_instrumentation__.Notify(32900)
		if strings.EqualFold(value, typ) {
			__antithesis_instrumentation__.Notify(32901)
			return keyType(i), nil
		} else {
			__antithesis_instrumentation__.Notify(32902)
		}
	}
	__antithesis_instrumentation__.Notify(32899)
	return 0, fmt.Errorf("unknown key type '%s'", value)
}

type nodeDecommissionWaitType int

const (
	nodeDecommissionWaitAll nodeDecommissionWaitType = iota
	nodeDecommissionWaitNone
)

func (s *nodeDecommissionWaitType) Type() string {
	__antithesis_instrumentation__.Notify(32903)
	return "string"
}

func (s *nodeDecommissionWaitType) String() string {
	__antithesis_instrumentation__.Notify(32904)
	switch *s {
	case nodeDecommissionWaitAll:
		__antithesis_instrumentation__.Notify(32905)
		return "all"
	case nodeDecommissionWaitNone:
		__antithesis_instrumentation__.Notify(32906)
		return "none"
	default:
		__antithesis_instrumentation__.Notify(32907)
		panic("unexpected node decommission wait type (possible values: all, none)")
	}
}

func (s *nodeDecommissionWaitType) Set(value string) error {
	__antithesis_instrumentation__.Notify(32908)
	switch value {
	case "all":
		__antithesis_instrumentation__.Notify(32910)
		*s = nodeDecommissionWaitAll
	case "none":
		__antithesis_instrumentation__.Notify(32911)
		*s = nodeDecommissionWaitNone
	default:
		__antithesis_instrumentation__.Notify(32912)
		return fmt.Errorf("invalid node decommission parameter: %s "+
			"(possible values: all, none)", value)
	}
	__antithesis_instrumentation__.Notify(32909)
	return nil
}

type bytesOrPercentageValue struct {
	bval *humanizeutil.BytesValue

	origVal string

	percentResolver percentResolverFunc
}
type percentResolverFunc func(percent int) (int64, error)

func memoryPercentResolver(percent int) (int64, error) {
	__antithesis_instrumentation__.Notify(32913)
	sizeBytes, _, err := status.GetTotalMemoryWithoutLogging()
	if err != nil {
		__antithesis_instrumentation__.Notify(32915)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(32916)
	}
	__antithesis_instrumentation__.Notify(32914)
	return (sizeBytes * int64(percent)) / 100, nil
}

func diskPercentResolverFactory(dir string) (percentResolverFunc, error) {
	__antithesis_instrumentation__.Notify(32917)
	du, err := vfs.Default.GetDiskUsage(dir)
	if err != nil {
		__antithesis_instrumentation__.Notify(32920)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(32921)
	}
	__antithesis_instrumentation__.Notify(32918)
	if du.TotalBytes > math.MaxInt64 {
		__antithesis_instrumentation__.Notify(32922)
		return nil, fmt.Errorf("unsupported disk size %s, max supported size is %s",
			humanize.IBytes(du.TotalBytes), humanizeutil.IBytes(math.MaxInt64))
	} else {
		__antithesis_instrumentation__.Notify(32923)
	}
	__antithesis_instrumentation__.Notify(32919)
	deviceCapacity := int64(du.TotalBytes)

	return func(percent int) (int64, error) {
		__antithesis_instrumentation__.Notify(32924)
		return (deviceCapacity * int64(percent)) / 100, nil
	}, nil
}

func newBytesOrPercentageValue(
	v *int64, percentResolver func(percent int) (int64, error),
) *bytesOrPercentageValue {
	__antithesis_instrumentation__.Notify(32925)
	return &bytesOrPercentageValue{
		bval:            humanizeutil.NewBytesValue(v),
		percentResolver: percentResolver,
	}
}

var fractionRE = regexp.MustCompile(`^0?\.[0-9]+$`)

func (b *bytesOrPercentageValue) Set(s string) error {
	__antithesis_instrumentation__.Notify(32926)
	b.origVal = s
	if strings.HasSuffix(s, "%") || func() bool {
		__antithesis_instrumentation__.Notify(32928)
		return fractionRE.MatchString(s) == true
	}() == true {
		__antithesis_instrumentation__.Notify(32929)
		multiplier := 100.0
		if s[len(s)-1] == '%' {
			__antithesis_instrumentation__.Notify(32935)

			multiplier = 1.0
			s = s[:len(s)-1]
		} else {
			__antithesis_instrumentation__.Notify(32936)
		}
		__antithesis_instrumentation__.Notify(32930)

		frac, err := strconv.ParseFloat(s, 32)
		if err != nil {
			__antithesis_instrumentation__.Notify(32937)
			return err
		} else {
			__antithesis_instrumentation__.Notify(32938)
		}
		__antithesis_instrumentation__.Notify(32931)
		percent := int(frac * multiplier)
		if percent < 1 || func() bool {
			__antithesis_instrumentation__.Notify(32939)
			return percent > 99 == true
		}() == true {
			__antithesis_instrumentation__.Notify(32940)
			return fmt.Errorf("percentage %d%% out of range 1%% - 99%%", percent)
		} else {
			__antithesis_instrumentation__.Notify(32941)
		}
		__antithesis_instrumentation__.Notify(32932)

		if b.percentResolver == nil {
			__antithesis_instrumentation__.Notify(32942)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(32943)
		}
		__antithesis_instrumentation__.Notify(32933)

		absVal, err := b.percentResolver(percent)
		if err != nil {
			__antithesis_instrumentation__.Notify(32944)
			return err
		} else {
			__antithesis_instrumentation__.Notify(32945)
		}
		__antithesis_instrumentation__.Notify(32934)
		s = fmt.Sprint(absVal)
	} else {
		__antithesis_instrumentation__.Notify(32946)
	}
	__antithesis_instrumentation__.Notify(32927)
	return b.bval.Set(s)
}

func (b *bytesOrPercentageValue) Resolve(v *int64, percentResolver percentResolverFunc) error {
	__antithesis_instrumentation__.Notify(32947)

	if b.origVal == "" {
		__antithesis_instrumentation__.Notify(32949)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(32950)
	}
	__antithesis_instrumentation__.Notify(32948)
	b.percentResolver = percentResolver
	b.bval = humanizeutil.NewBytesValue(v)
	return b.Set(b.origVal)
}

func (b *bytesOrPercentageValue) Type() string {
	__antithesis_instrumentation__.Notify(32951)
	return b.bval.Type()
}

func (b *bytesOrPercentageValue) String() string {
	__antithesis_instrumentation__.Notify(32952)
	return b.bval.String()
}

func (b *bytesOrPercentageValue) IsSet() bool {
	__antithesis_instrumentation__.Notify(32953)
	return b.bval.IsSet()
}
