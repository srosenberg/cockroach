package keys

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

var PrettyPrintTimeseriesKey func(key roachpb.Key) string

type DictEntry struct {
	Name   string
	prefix roachpb.Key

	ppFunc func(valDirs []encoding.Direction, key roachpb.Key) string

	PSFunc KeyParserFunc
}

type KeyParserFunc func(input string) (string, roachpb.Key)

func parseUnsupported(_ string) (string, roachpb.Key) {
	__antithesis_instrumentation__.Notify(85542)
	panic(&ErrUglifyUnsupported{})
}

type KeyComprehensionTable []struct {
	Name    string
	start   roachpb.Key
	end     roachpb.Key
	Entries []DictEntry
}

var (
	ConstKeyDict = []struct {
		Name  string
		Value roachpb.Key
	}{
		{"/Max", MaxKey},
		{"/Min", MinKey},
		{"/Meta1/Max", Meta1KeyMax},
		{"/Meta2/Max", Meta2KeyMax},
	}

	KeyDict = KeyComprehensionTable{
		{Name: "/Local", start: LocalPrefix, end: LocalMax, Entries: []DictEntry{
			{Name: "/Store", prefix: roachpb.Key(LocalStorePrefix),
				ppFunc: localStoreKeyPrint, PSFunc: localStoreKeyParse},
			{Name: "/RangeID", prefix: roachpb.Key(LocalRangeIDPrefix),
				ppFunc: localRangeIDKeyPrint, PSFunc: localRangeIDKeyParse},
			{Name: "/Range", prefix: LocalRangePrefix, ppFunc: localRangeKeyPrint,
				PSFunc: parseUnsupported},
			{Name: "/Lock", prefix: LocalRangeLockTablePrefix, ppFunc: localRangeLockTablePrint,
				PSFunc: parseUnsupported},
		}},
		{Name: "/Meta1", start: Meta1Prefix, end: Meta1KeyMax, Entries: []DictEntry{
			{Name: "", prefix: Meta1Prefix, ppFunc: print,
				PSFunc: func(input string) (string, roachpb.Key) {
					__antithesis_instrumentation__.Notify(85543)
					input = mustShiftSlash(input)
					unq, err := strconv.Unquote(input)
					if err != nil {
						__antithesis_instrumentation__.Notify(85546)
						panic(err)
					} else {
						__antithesis_instrumentation__.Notify(85547)
					}
					__antithesis_instrumentation__.Notify(85544)
					if len(unq) == 0 {
						__antithesis_instrumentation__.Notify(85548)
						return "", Meta1Prefix
					} else {
						__antithesis_instrumentation__.Notify(85549)
					}
					__antithesis_instrumentation__.Notify(85545)
					return "", RangeMetaKey(RangeMetaKey(MustAddr(
						roachpb.Key(unq)))).AsRawKey()
				},
			}},
		},
		{Name: "/Meta2", start: Meta2Prefix, end: Meta2KeyMax, Entries: []DictEntry{
			{Name: "", prefix: Meta2Prefix, ppFunc: print,
				PSFunc: func(input string) (string, roachpb.Key) {
					__antithesis_instrumentation__.Notify(85550)
					input = mustShiftSlash(input)
					unq, err := strconv.Unquote(input)
					if err != nil {
						__antithesis_instrumentation__.Notify(85553)
						panic(&ErrUglifyUnsupported{err})
					} else {
						__antithesis_instrumentation__.Notify(85554)
					}
					__antithesis_instrumentation__.Notify(85551)
					if len(unq) == 0 {
						__antithesis_instrumentation__.Notify(85555)
						return "", Meta2Prefix
					} else {
						__antithesis_instrumentation__.Notify(85556)
					}
					__antithesis_instrumentation__.Notify(85552)
					return "", RangeMetaKey(MustAddr(roachpb.Key(unq))).AsRawKey()
				},
			}},
		},
		{Name: "/System", start: SystemPrefix, end: SystemMax, Entries: []DictEntry{
			{Name: "/NodeLiveness", prefix: NodeLivenessPrefix,
				ppFunc: decodeKeyPrint,
				PSFunc: parseUnsupported,
			},
			{Name: "/NodeLivenessMax", prefix: NodeLivenessKeyMax,
				ppFunc: decodeKeyPrint,
				PSFunc: parseUnsupported,
			},
			{Name: "/StatusNode", prefix: StatusNodePrefix,
				ppFunc: decodeKeyPrint,
				PSFunc: parseUnsupported,
			},
			{Name: "/tsd", prefix: TimeseriesPrefix,
				ppFunc: timeseriesKeyPrint,
				PSFunc: parseUnsupported,
			},
			{Name: "/SystemSpanConfigKeys", prefix: SystemSpanConfigPrefix,
				ppFunc: decodeKeyPrint,
				PSFunc: parseUnsupported,
			},
		}},
		{Name: "/NamespaceTable", start: NamespaceTableMin, end: NamespaceTableMax, Entries: []DictEntry{
			{Name: "", prefix: nil, ppFunc: decodeKeyPrint, PSFunc: parseUnsupported},
		}},
		{Name: "/Table", start: TableDataMin, end: TableDataMax, Entries: []DictEntry{
			{Name: "", prefix: nil, ppFunc: decodeKeyPrint, PSFunc: tableKeyParse},
		}},
		{Name: "/Tenant", start: TenantTableDataMin, end: TenantTableDataMax, Entries: []DictEntry{
			{Name: "", prefix: nil, ppFunc: tenantKeyPrint, PSFunc: tenantKeyParse},
		}},
	}

	keyOfKeyDict = []struct {
		name   string
		prefix []byte
	}{
		{name: "/Meta2", prefix: Meta2Prefix},
		{name: "/Meta1", prefix: Meta1Prefix},
	}

	rangeIDSuffixDict = []struct {
		name   string
		suffix []byte
		ppFunc func(key roachpb.Key) string
		psFunc func(rangeID roachpb.RangeID, input string) (string, roachpb.Key)
	}{
		{name: "AbortSpan", suffix: LocalAbortSpanSuffix, ppFunc: abortSpanKeyPrint, psFunc: abortSpanKeyParse},
		{name: "RangeTombstone", suffix: LocalRangeTombstoneSuffix},
		{name: "RaftHardState", suffix: LocalRaftHardStateSuffix},
		{name: "RangeAppliedState", suffix: LocalRangeAppliedStateSuffix},
		{name: "RaftLog", suffix: LocalRaftLogSuffix,
			ppFunc: raftLogKeyPrint,
			psFunc: raftLogKeyParse,
		},
		{name: "RaftTruncatedState", suffix: LocalRaftTruncatedStateSuffix},
		{name: "RangeLastReplicaGCTimestamp", suffix: LocalRangeLastReplicaGCTimestampSuffix},
		{name: "RangeLease", suffix: LocalRangeLeaseSuffix},
		{name: "RangePriorReadSummary", suffix: LocalRangePriorReadSummarySuffix},
		{name: "RangeStats", suffix: LocalRangeStatsLegacySuffix},
		{name: "RangeGCThreshold", suffix: LocalRangeGCThresholdSuffix},
		{name: "RangeVersion", suffix: LocalRangeVersionSuffix},
	}

	rangeSuffixDict = []struct {
		name   string
		suffix []byte
		atEnd  bool
	}{
		{name: "RangeDescriptor", suffix: LocalRangeDescriptorSuffix, atEnd: true},
		{name: "Transaction", suffix: LocalTransactionSuffix, atEnd: false},
		{name: "QueueLastProcessed", suffix: LocalQueueLastProcessedSuffix, atEnd: false},
		{name: "RangeProbe", suffix: LocalRangeProbeSuffix, atEnd: true},
	}
)

var constSubKeyDict = []struct {
	name string
	key  roachpb.RKey
}{
	{"/storeIdent", localStoreIdentSuffix},
	{"/gossipBootstrap", localStoreGossipSuffix},
	{"/clusterVersion", localStoreClusterVersionSuffix},
	{"/nodeTombstone", localStoreNodeTombstoneSuffix},
	{"/cachedSettings", localStoreCachedSettingsSuffix},
	{"/lossOfQuorumRecovery/applied", localStoreUnsafeReplicaRecoverySuffix},
}

func nodeTombstoneKeyPrint(key roachpb.Key) string {
	__antithesis_instrumentation__.Notify(85557)
	nodeID, err := DecodeNodeTombstoneKey(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(85559)
		return fmt.Sprintf("<invalid: %s>", err)
	} else {
		__antithesis_instrumentation__.Notify(85560)
	}
	__antithesis_instrumentation__.Notify(85558)
	return fmt.Sprint("n", nodeID)
}

func cachedSettingsKeyPrint(key roachpb.Key) string {
	__antithesis_instrumentation__.Notify(85561)
	settingKey, err := DecodeStoreCachedSettingsKey(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(85563)
		return fmt.Sprintf("<invalid: %s>", err)
	} else {
		__antithesis_instrumentation__.Notify(85564)
	}
	__antithesis_instrumentation__.Notify(85562)
	return settingKey.String()
}

func localStoreKeyPrint(_ []encoding.Direction, key roachpb.Key) string {
	__antithesis_instrumentation__.Notify(85565)
	for _, v := range constSubKeyDict {
		__antithesis_instrumentation__.Notify(85567)
		if bytes.HasPrefix(key, v.key) {
			__antithesis_instrumentation__.Notify(85568)
			if v.key.Equal(localStoreNodeTombstoneSuffix) {
				__antithesis_instrumentation__.Notify(85570)
				return v.name + "/" + nodeTombstoneKeyPrint(
					append(roachpb.Key(nil), append(LocalStorePrefix, key...)...),
				)
			} else {
				__antithesis_instrumentation__.Notify(85571)
				if v.key.Equal(localStoreCachedSettingsSuffix) {
					__antithesis_instrumentation__.Notify(85572)
					return v.name + "/" + cachedSettingsKeyPrint(
						append(roachpb.Key(nil), append(LocalStorePrefix, key...)...),
					)
				} else {
					__antithesis_instrumentation__.Notify(85573)
					if v.key.Equal(localStoreUnsafeReplicaRecoverySuffix) {
						__antithesis_instrumentation__.Notify(85574)
						return v.name + "/" + lossOfQuorumRecoveryEntryKeyPrint(
							append(roachpb.Key(nil), append(LocalStorePrefix, key...)...),
						)
					} else {
						__antithesis_instrumentation__.Notify(85575)
					}
				}
			}
			__antithesis_instrumentation__.Notify(85569)
			return v.name
		} else {
			__antithesis_instrumentation__.Notify(85576)
		}
	}
	__antithesis_instrumentation__.Notify(85566)

	return fmt.Sprintf("%q", []byte(key))
}

func lossOfQuorumRecoveryEntryKeyPrint(key roachpb.Key) string {
	__antithesis_instrumentation__.Notify(85577)
	entryID, err := DecodeStoreUnsafeReplicaRecoveryKey(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(85579)
		return fmt.Sprintf("<invalid: %s>", err)
	} else {
		__antithesis_instrumentation__.Notify(85580)
	}
	__antithesis_instrumentation__.Notify(85578)
	return entryID.String()
}

func localStoreKeyParse(input string) (remainder string, output roachpb.Key) {
	__antithesis_instrumentation__.Notify(85581)
	for _, s := range constSubKeyDict {
		__antithesis_instrumentation__.Notify(85584)
		if strings.HasPrefix(input, s.name) {
			__antithesis_instrumentation__.Notify(85585)
			switch {
			case
				s.key.Equal(localStoreNodeTombstoneSuffix),
				s.key.Equal(localStoreCachedSettingsSuffix):
				__antithesis_instrumentation__.Notify(85587)
				panic(&ErrUglifyUnsupported{errors.Errorf("cannot parse local store key with suffix %s", s.key)})
			case s.key.Equal(localStoreUnsafeReplicaRecoverySuffix):
				__antithesis_instrumentation__.Notify(85588)
				recordIDString := input[len(localStoreUnsafeReplicaRecoverySuffix):]
				recordUUID, err := uuid.FromString(recordIDString)
				if err != nil {
					__antithesis_instrumentation__.Notify(85591)
					panic(&ErrUglifyUnsupported{errors.Errorf("cannot parse local store key with suffix %s", s.key)})
				} else {
					__antithesis_instrumentation__.Notify(85592)
				}
				__antithesis_instrumentation__.Notify(85589)
				output = StoreUnsafeReplicaRecoveryKey(recordUUID)
			default:
				__antithesis_instrumentation__.Notify(85590)
				output = MakeStoreKey(s.key, nil)
			}
			__antithesis_instrumentation__.Notify(85586)
			return
		} else {
			__antithesis_instrumentation__.Notify(85593)
		}
	}
	__antithesis_instrumentation__.Notify(85582)
	input = mustShiftSlash(input)
	slashPos := strings.IndexByte(input, '/')
	if slashPos < 0 {
		__antithesis_instrumentation__.Notify(85594)
		slashPos = len(input)
	} else {
		__antithesis_instrumentation__.Notify(85595)
	}
	__antithesis_instrumentation__.Notify(85583)
	remainder = input[slashPos:]
	output = roachpb.Key(input[:slashPos])
	return
}

const strTable = "/Table/"
const strSystemConfigSpan = "SystemConfigSpan"
const strSystemConfigSpanStart = "Start"

func tenantKeyParse(input string) (remainder string, output roachpb.Key) {
	__antithesis_instrumentation__.Notify(85596)
	input = mustShiftSlash(input)
	slashPos := strings.Index(input, "/")
	if slashPos < 0 {
		__antithesis_instrumentation__.Notify(85600)
		slashPos = len(input)
	} else {
		__antithesis_instrumentation__.Notify(85601)
	}
	__antithesis_instrumentation__.Notify(85597)
	remainder = input[slashPos:]
	tenantIDStr := input[:slashPos]
	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 64)
	if err != nil {
		__antithesis_instrumentation__.Notify(85602)
		panic(&ErrUglifyUnsupported{err})
	} else {
		__antithesis_instrumentation__.Notify(85603)
	}
	__antithesis_instrumentation__.Notify(85598)
	output = MakeTenantPrefix(roachpb.MakeTenantID(tenantID))
	if strings.HasPrefix(remainder, strTable) {
		__antithesis_instrumentation__.Notify(85604)
		var indexKey roachpb.Key
		remainder = remainder[len(strTable)-1:]
		remainder, indexKey = tableKeyParse(remainder)
		output = append(output, indexKey...)
	} else {
		__antithesis_instrumentation__.Notify(85605)
	}
	__antithesis_instrumentation__.Notify(85599)
	return remainder, output
}

func tableKeyParse(input string) (remainder string, output roachpb.Key) {
	__antithesis_instrumentation__.Notify(85606)
	input = mustShiftSlash(input)
	slashPos := strings.Index(input, "/")
	if slashPos < 0 {
		__antithesis_instrumentation__.Notify(85611)
		slashPos = len(input)
	} else {
		__antithesis_instrumentation__.Notify(85612)
	}
	__antithesis_instrumentation__.Notify(85607)
	remainder = input[slashPos:]
	tableIDStr := input[:slashPos]
	if tableIDStr == strSystemConfigSpan {
		__antithesis_instrumentation__.Notify(85613)
		if remainder[1:] == strSystemConfigSpanStart {
			__antithesis_instrumentation__.Notify(85615)
			remainder = ""
		} else {
			__antithesis_instrumentation__.Notify(85616)
		}
		__antithesis_instrumentation__.Notify(85614)
		output = SystemConfigSpan.Key
		return
	} else {
		__antithesis_instrumentation__.Notify(85617)
	}
	__antithesis_instrumentation__.Notify(85608)
	tableID, err := strconv.ParseUint(tableIDStr, 10, 32)
	if err != nil {
		__antithesis_instrumentation__.Notify(85618)
		panic(&ErrUglifyUnsupported{err})
	} else {
		__antithesis_instrumentation__.Notify(85619)
	}
	__antithesis_instrumentation__.Notify(85609)
	output = encoding.EncodeUvarintAscending(nil, tableID)
	if remainder != "" {
		__antithesis_instrumentation__.Notify(85620)
		var indexKey roachpb.Key
		remainder, indexKey = tableIndexParse(remainder)
		output = append(output, indexKey...)
	} else {
		__antithesis_instrumentation__.Notify(85621)
	}
	__antithesis_instrumentation__.Notify(85610)
	return remainder, output
}

func tableIndexParse(input string) (string, roachpb.Key) {
	__antithesis_instrumentation__.Notify(85622)
	input = mustShiftSlash(input)
	slashPos := strings.Index(input, "/")
	if slashPos < 0 {
		__antithesis_instrumentation__.Notify(85625)

		slashPos = len(input)
	} else {
		__antithesis_instrumentation__.Notify(85626)
	}
	__antithesis_instrumentation__.Notify(85623)
	remainder := input[slashPos:]
	indexIDStr := input[:slashPos]
	indexID, err := strconv.ParseUint(indexIDStr, 10, 32)
	if err != nil {
		__antithesis_instrumentation__.Notify(85627)
		panic(&ErrUglifyUnsupported{err})
	} else {
		__antithesis_instrumentation__.Notify(85628)
	}
	__antithesis_instrumentation__.Notify(85624)
	output := encoding.EncodeUvarintAscending(nil, indexID)
	return remainder, output
}

const strLogIndex = "/logIndex:"

func raftLogKeyParse(rangeID roachpb.RangeID, input string) (string, roachpb.Key) {
	__antithesis_instrumentation__.Notify(85629)
	if !strings.HasPrefix(input, strLogIndex) {
		__antithesis_instrumentation__.Notify(85632)
		panic("expected log index")
	} else {
		__antithesis_instrumentation__.Notify(85633)
	}
	__antithesis_instrumentation__.Notify(85630)
	input = input[len(strLogIndex):]
	index, err := strconv.ParseUint(input, 10, 64)
	if err != nil {
		__antithesis_instrumentation__.Notify(85634)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(85635)
	}
	__antithesis_instrumentation__.Notify(85631)
	return "", RaftLogKey(rangeID, index)
}

func raftLogKeyPrint(key roachpb.Key) string {
	__antithesis_instrumentation__.Notify(85636)
	var logIndex uint64
	var err error
	key, logIndex, err = encoding.DecodeUint64Ascending(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(85638)
		return fmt.Sprintf("/err<%v:%q>", err, []byte(key))
	} else {
		__antithesis_instrumentation__.Notify(85639)
	}
	__antithesis_instrumentation__.Notify(85637)

	return fmt.Sprintf("%s%d", strLogIndex, logIndex)
}

func mustShiftSlash(in string) string {
	__antithesis_instrumentation__.Notify(85640)
	slash, out := mustShift(in)
	if slash != "/" {
		__antithesis_instrumentation__.Notify(85642)
		panic("expected /: " + in)
	} else {
		__antithesis_instrumentation__.Notify(85643)
	}
	__antithesis_instrumentation__.Notify(85641)
	return out
}

func mustShift(in string) (first, remainder string) {
	__antithesis_instrumentation__.Notify(85644)
	if len(in) == 0 {
		__antithesis_instrumentation__.Notify(85646)
		panic("premature end of string")
	} else {
		__antithesis_instrumentation__.Notify(85647)
	}
	__antithesis_instrumentation__.Notify(85645)
	return in[:1], in[1:]
}

func localRangeIDKeyParse(input string) (remainder string, key roachpb.Key) {
	__antithesis_instrumentation__.Notify(85648)
	var rangeID int64
	var err error
	input = mustShiftSlash(input)
	if endPos := strings.IndexByte(input, '/'); endPos > 0 {
		__antithesis_instrumentation__.Notify(85654)
		rangeID, err = strconv.ParseInt(input[:endPos], 10, 64)
		if err != nil {
			__antithesis_instrumentation__.Notify(85656)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(85657)
		}
		__antithesis_instrumentation__.Notify(85655)
		input = input[endPos:]
	} else {
		__antithesis_instrumentation__.Notify(85658)
		panic(errors.Errorf("illegal RangeID: %q", input))
	}
	__antithesis_instrumentation__.Notify(85649)
	input = mustShiftSlash(input)
	var infix string
	infix, input = mustShift(input)
	var replicated bool
	switch {
	case bytes.Equal(localRangeIDUnreplicatedInfix, []byte(infix)):
		__antithesis_instrumentation__.Notify(85659)
	case bytes.Equal(LocalRangeIDReplicatedInfix, []byte(infix)):
		__antithesis_instrumentation__.Notify(85660)
		replicated = true
	default:
		__antithesis_instrumentation__.Notify(85661)
		panic(errors.Errorf("invalid infix: %q", infix))
	}
	__antithesis_instrumentation__.Notify(85650)

	input = mustShiftSlash(input)

	var suffix roachpb.RKey
	for _, s := range rangeIDSuffixDict {
		__antithesis_instrumentation__.Notify(85662)
		if strings.HasPrefix(input, s.name) {
			__antithesis_instrumentation__.Notify(85663)
			input = input[len(s.name):]
			if s.psFunc != nil {
				__antithesis_instrumentation__.Notify(85665)
				remainder, key = s.psFunc(roachpb.RangeID(rangeID), input)
				return
			} else {
				__antithesis_instrumentation__.Notify(85666)
			}
			__antithesis_instrumentation__.Notify(85664)
			suffix = roachpb.RKey(s.suffix)
			break
		} else {
			__antithesis_instrumentation__.Notify(85667)
		}
	}
	__antithesis_instrumentation__.Notify(85651)
	maker := makeRangeIDUnreplicatedKey
	if replicated {
		__antithesis_instrumentation__.Notify(85668)
		maker = makeRangeIDReplicatedKey
	} else {
		__antithesis_instrumentation__.Notify(85669)
	}
	__antithesis_instrumentation__.Notify(85652)
	if suffix != nil {
		__antithesis_instrumentation__.Notify(85670)
		if input != "" {
			__antithesis_instrumentation__.Notify(85672)
			panic(&ErrUglifyUnsupported{errors.New("nontrivial detail")})
		} else {
			__antithesis_instrumentation__.Notify(85673)
		}
		__antithesis_instrumentation__.Notify(85671)
		var detail roachpb.RKey

		remainder = ""
		key = maker(roachpb.RangeID(rangeID), suffix, detail)
		return
	} else {
		__antithesis_instrumentation__.Notify(85674)
	}
	__antithesis_instrumentation__.Notify(85653)
	panic(&ErrUglifyUnsupported{errors.New("unhandled general range key")})
}

func localRangeIDKeyPrint(valDirs []encoding.Direction, key roachpb.Key) string {
	__antithesis_instrumentation__.Notify(85675)
	var buf bytes.Buffer
	if encoding.PeekType(key) != encoding.Int {
		__antithesis_instrumentation__.Notify(85681)
		return fmt.Sprintf("/err<%q>", []byte(key))
	} else {
		__antithesis_instrumentation__.Notify(85682)
	}
	__antithesis_instrumentation__.Notify(85676)

	key, i, err := encoding.DecodeVarintAscending(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(85683)
		return fmt.Sprintf("/err<%v:%q>", err, []byte(key))
	} else {
		__antithesis_instrumentation__.Notify(85684)
	}
	__antithesis_instrumentation__.Notify(85677)

	fmt.Fprintf(&buf, "/%d", i)

	if len(key) != 0 {
		__antithesis_instrumentation__.Notify(85685)
		fmt.Fprintf(&buf, "/%s", string(key[0]))
		key = key[1:]
	} else {
		__antithesis_instrumentation__.Notify(85686)
	}
	__antithesis_instrumentation__.Notify(85678)

	hasSuffix := false
	for _, s := range rangeIDSuffixDict {
		__antithesis_instrumentation__.Notify(85687)
		if bytes.HasPrefix(key, s.suffix) {
			__antithesis_instrumentation__.Notify(85688)
			fmt.Fprintf(&buf, "/%s", s.name)
			key = key[len(s.suffix):]
			if s.ppFunc != nil && func() bool {
				__antithesis_instrumentation__.Notify(85690)
				return len(key) != 0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(85691)
				fmt.Fprintf(&buf, "%s", s.ppFunc(key))
				return buf.String()
			} else {
				__antithesis_instrumentation__.Notify(85692)
			}
			__antithesis_instrumentation__.Notify(85689)
			hasSuffix = true
			break
		} else {
			__antithesis_instrumentation__.Notify(85693)
		}
	}
	__antithesis_instrumentation__.Notify(85679)

	if hasSuffix {
		__antithesis_instrumentation__.Notify(85694)
		fmt.Fprintf(&buf, "%s", decodeKeyPrint(valDirs, key))
	} else {
		__antithesis_instrumentation__.Notify(85695)
		fmt.Fprintf(&buf, "%q", []byte(key))
	}
	__antithesis_instrumentation__.Notify(85680)

	return buf.String()
}

func localRangeKeyPrint(valDirs []encoding.Direction, key roachpb.Key) string {
	__antithesis_instrumentation__.Notify(85696)
	var buf bytes.Buffer

	for _, s := range rangeSuffixDict {
		__antithesis_instrumentation__.Notify(85699)
		if s.atEnd {
			__antithesis_instrumentation__.Notify(85700)
			if bytes.HasSuffix(key, s.suffix) {
				__antithesis_instrumentation__.Notify(85701)
				key = key[:len(key)-len(s.suffix)]
				_, decodedKey, err := encoding.DecodeBytesAscending([]byte(key), nil)
				if err != nil {
					__antithesis_instrumentation__.Notify(85703)
					fmt.Fprintf(&buf, "%s/%s", decodeKeyPrint(valDirs, key), s.name)
				} else {
					__antithesis_instrumentation__.Notify(85704)
					fmt.Fprintf(&buf, "%s/%s", roachpb.Key(decodedKey), s.name)
				}
				__antithesis_instrumentation__.Notify(85702)
				return buf.String()
			} else {
				__antithesis_instrumentation__.Notify(85705)
			}
		} else {
			__antithesis_instrumentation__.Notify(85706)
			begin := bytes.Index(key, s.suffix)
			if begin > 0 {
				__antithesis_instrumentation__.Notify(85707)
				addrKey := key[:begin]
				_, decodedAddrKey, err := encoding.DecodeBytesAscending([]byte(addrKey), nil)
				if err != nil {
					__antithesis_instrumentation__.Notify(85710)
					fmt.Fprintf(&buf, "%s/%s", decodeKeyPrint(valDirs, addrKey), s.name)
				} else {
					__antithesis_instrumentation__.Notify(85711)
					fmt.Fprintf(&buf, "%s/%s", roachpb.Key(decodedAddrKey), s.name)
				}
				__antithesis_instrumentation__.Notify(85708)
				if bytes.Equal(s.suffix, LocalTransactionSuffix) {
					__antithesis_instrumentation__.Notify(85712)
					txnID, err := uuid.FromBytes(key[(begin + len(s.suffix)):])
					if err != nil {
						__antithesis_instrumentation__.Notify(85714)
						return fmt.Sprintf("/%q/err:%v", key, err)
					} else {
						__antithesis_instrumentation__.Notify(85715)
					}
					__antithesis_instrumentation__.Notify(85713)
					fmt.Fprintf(&buf, "/%q", txnID)
				} else {
					__antithesis_instrumentation__.Notify(85716)
					id := key[(begin + len(s.suffix)):]
					fmt.Fprintf(&buf, "/%q", []byte(id))
				}
				__antithesis_instrumentation__.Notify(85709)
				return buf.String()
			} else {
				__antithesis_instrumentation__.Notify(85717)
			}
		}
	}
	__antithesis_instrumentation__.Notify(85697)

	_, decodedKey, err := encoding.DecodeBytesAscending([]byte(key), nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(85718)
		fmt.Fprintf(&buf, "%s", decodeKeyPrint(valDirs, key))
	} else {
		__antithesis_instrumentation__.Notify(85719)
		fmt.Fprintf(&buf, "%s", roachpb.Key(decodedKey))
	}
	__antithesis_instrumentation__.Notify(85698)

	return buf.String()
}

var lockTablePrintLockedKey func(valDirs []encoding.Direction, key roachpb.Key, quoteRawKeys bool) string

func localRangeLockTablePrint(valDirs []encoding.Direction, key roachpb.Key) string {
	__antithesis_instrumentation__.Notify(85720)
	var buf bytes.Buffer
	if !bytes.HasPrefix(key, LockTableSingleKeyInfix) {
		__antithesis_instrumentation__.Notify(85723)
		fmt.Fprintf(&buf, "/\"%x\"", key)
		return buf.String()
	} else {
		__antithesis_instrumentation__.Notify(85724)
	}
	__antithesis_instrumentation__.Notify(85721)
	buf.WriteString("/Intent")
	key = key[len(LockTableSingleKeyInfix):]
	b, lockedKey, err := encoding.DecodeBytesAscending(key, nil)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(85725)
		return len(b) != 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(85726)
		fmt.Fprintf(&buf, "/\"%x\"", key)
		return buf.String()
	} else {
		__antithesis_instrumentation__.Notify(85727)
	}
	__antithesis_instrumentation__.Notify(85722)
	buf.WriteString(lockTablePrintLockedKey(valDirs, lockedKey, true))
	return buf.String()
}

type ErrUglifyUnsupported struct {
	Wrapped error
}

func (euu *ErrUglifyUnsupported) Error() string {
	__antithesis_instrumentation__.Notify(85728)
	return fmt.Sprintf("unsupported pretty key: %v", euu.Wrapped)
}

func abortSpanKeyParse(rangeID roachpb.RangeID, input string) (string, roachpb.Key) {
	__antithesis_instrumentation__.Notify(85729)
	var err error
	input = mustShiftSlash(input)
	_, input = mustShift(input[:len(input)-1])
	if len(input) != len(uuid.UUID{}.String()) {
		__antithesis_instrumentation__.Notify(85732)
		panic(&ErrUglifyUnsupported{errors.New("txn id not available")})
	} else {
		__antithesis_instrumentation__.Notify(85733)
	}
	__antithesis_instrumentation__.Notify(85730)
	id, err := uuid.FromString(input)
	if err != nil {
		__antithesis_instrumentation__.Notify(85734)
		panic(&ErrUglifyUnsupported{err})
	} else {
		__antithesis_instrumentation__.Notify(85735)
	}
	__antithesis_instrumentation__.Notify(85731)
	return "", AbortSpanKey(rangeID, id)
}

func abortSpanKeyPrint(key roachpb.Key) string {
	__antithesis_instrumentation__.Notify(85736)
	_, id, err := encoding.DecodeBytesAscending([]byte(key), nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(85739)
		return fmt.Sprintf("/%q/err:%v", key, err)
	} else {
		__antithesis_instrumentation__.Notify(85740)
	}
	__antithesis_instrumentation__.Notify(85737)

	txnID, err := uuid.FromBytes(id)
	if err != nil {
		__antithesis_instrumentation__.Notify(85741)
		return fmt.Sprintf("/%q/err:%v", key, err)
	} else {
		__antithesis_instrumentation__.Notify(85742)
	}
	__antithesis_instrumentation__.Notify(85738)

	return fmt.Sprintf("/%q", txnID)
}

func print(_ []encoding.Direction, key roachpb.Key) string {
	__antithesis_instrumentation__.Notify(85743)
	return fmt.Sprintf("/%q", []byte(key))
}

func decodeKeyPrint(valDirs []encoding.Direction, key roachpb.Key) string {
	__antithesis_instrumentation__.Notify(85744)
	if key.Equal(SystemConfigSpan.Key) {
		__antithesis_instrumentation__.Notify(85746)
		return "/SystemConfigSpan/Start"
	} else {
		__antithesis_instrumentation__.Notify(85747)
	}
	__antithesis_instrumentation__.Notify(85745)
	return encoding.PrettyPrintValue(valDirs, key, "/")
}

func timeseriesKeyPrint(_ []encoding.Direction, key roachpb.Key) string {
	__antithesis_instrumentation__.Notify(85748)
	return PrettyPrintTimeseriesKey(key)
}

func tenantKeyPrint(valDirs []encoding.Direction, key roachpb.Key) string {
	__antithesis_instrumentation__.Notify(85749)
	key, tID, err := DecodeTenantPrefix(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(85752)
		return fmt.Sprintf("/err:%v", err)
	} else {
		__antithesis_instrumentation__.Notify(85753)
	}
	__antithesis_instrumentation__.Notify(85750)
	if len(key) == 0 {
		__antithesis_instrumentation__.Notify(85754)
		return fmt.Sprintf("/%s", tID)
	} else {
		__antithesis_instrumentation__.Notify(85755)
	}
	__antithesis_instrumentation__.Notify(85751)
	return fmt.Sprintf("/%s%s", tID, key.StringWithDirs(valDirs, 0))
}

func prettyPrintInternal(valDirs []encoding.Direction, key roachpb.Key, quoteRawKeys bool) string {
	__antithesis_instrumentation__.Notify(85756)
	for _, k := range ConstKeyDict {
		__antithesis_instrumentation__.Notify(85760)
		if key.Equal(k.Value) {
			__antithesis_instrumentation__.Notify(85761)
			return k.Name
		} else {
			__antithesis_instrumentation__.Notify(85762)
		}
	}
	__antithesis_instrumentation__.Notify(85757)

	helper := func(key roachpb.Key) (string, bool) {
		__antithesis_instrumentation__.Notify(85763)
		var b strings.Builder
		for _, k := range KeyDict {
			__antithesis_instrumentation__.Notify(85766)
			if key.Compare(k.start) >= 0 && func() bool {
				__antithesis_instrumentation__.Notify(85767)
				return (k.end == nil || func() bool {
					__antithesis_instrumentation__.Notify(85768)
					return key.Compare(k.end) <= 0 == true
				}() == true) == true
			}() == true {
				__antithesis_instrumentation__.Notify(85769)
				b.WriteString(k.Name)
				if k.end != nil && func() bool {
					__antithesis_instrumentation__.Notify(85773)
					return k.end.Compare(key) == 0 == true
				}() == true {
					__antithesis_instrumentation__.Notify(85774)
					b.WriteString("/Max")
					return b.String(), true
				} else {
					__antithesis_instrumentation__.Notify(85775)
				}
				__antithesis_instrumentation__.Notify(85770)

				hasPrefix := false
				for _, e := range k.Entries {
					__antithesis_instrumentation__.Notify(85776)
					if bytes.HasPrefix(key, e.prefix) {
						__antithesis_instrumentation__.Notify(85777)
						hasPrefix = true
						key = key[len(e.prefix):]
						b.WriteString(e.Name)
						b.WriteString(e.ppFunc(valDirs, key))
						break
					} else {
						__antithesis_instrumentation__.Notify(85778)
					}
				}
				__antithesis_instrumentation__.Notify(85771)
				if !hasPrefix {
					__antithesis_instrumentation__.Notify(85779)
					key = key[len(k.start):]
					if quoteRawKeys {
						__antithesis_instrumentation__.Notify(85781)
						b.WriteByte('/')
						b.WriteByte('"')
					} else {
						__antithesis_instrumentation__.Notify(85782)
					}
					__antithesis_instrumentation__.Notify(85780)
					b.Write([]byte(key))
					if quoteRawKeys {
						__antithesis_instrumentation__.Notify(85783)
						b.WriteByte('"')
					} else {
						__antithesis_instrumentation__.Notify(85784)
					}
				} else {
					__antithesis_instrumentation__.Notify(85785)
				}
				__antithesis_instrumentation__.Notify(85772)

				return b.String(), true
			} else {
				__antithesis_instrumentation__.Notify(85786)
			}
		}
		__antithesis_instrumentation__.Notify(85764)

		if quoteRawKeys {
			__antithesis_instrumentation__.Notify(85787)
			return fmt.Sprintf("%q", []byte(key)), false
		} else {
			__antithesis_instrumentation__.Notify(85788)
		}
		__antithesis_instrumentation__.Notify(85765)
		return string(key), false
	}
	__antithesis_instrumentation__.Notify(85758)

	for _, k := range keyOfKeyDict {
		__antithesis_instrumentation__.Notify(85789)
		if bytes.HasPrefix(key, k.prefix) {
			__antithesis_instrumentation__.Notify(85790)
			key = key[len(k.prefix):]
			str, formatted := helper(key)
			if formatted {
				__antithesis_instrumentation__.Notify(85792)
				return k.name + str
			} else {
				__antithesis_instrumentation__.Notify(85793)
			}
			__antithesis_instrumentation__.Notify(85791)
			return k.name + "/" + str
		} else {
			__antithesis_instrumentation__.Notify(85794)
		}
	}
	__antithesis_instrumentation__.Notify(85759)
	str, _ := helper(key)
	return str
}

func PrettyPrint(valDirs []encoding.Direction, key roachpb.Key) string {
	__antithesis_instrumentation__.Notify(85795)
	return prettyPrintInternal(valDirs, key, true)
}

func init() {
	roachpb.PrettyPrintKey = PrettyPrint
	roachpb.PrettyPrintRange = PrettyPrintRange
	lockTablePrintLockedKey = prettyPrintInternal
}

func PrettyPrintRange(start, end roachpb.Key, maxChars int) string {
	__antithesis_instrumentation__.Notify(85796)
	var b bytes.Buffer
	if maxChars < 8 {
		__antithesis_instrumentation__.Notify(85802)
		maxChars = 8
	} else {
		__antithesis_instrumentation__.Notify(85803)
	}
	__antithesis_instrumentation__.Notify(85797)
	prettyStart := prettyPrintInternal(nil, start, false)
	if len(end) == 0 {
		__antithesis_instrumentation__.Notify(85804)
		if len(prettyStart) <= maxChars {
			__antithesis_instrumentation__.Notify(85806)
			return prettyStart
		} else {
			__antithesis_instrumentation__.Notify(85807)
		}
		__antithesis_instrumentation__.Notify(85805)
		copyEscape(&b, prettyStart[:maxChars-1])
		b.WriteRune('…')
		return b.String()
	} else {
		__antithesis_instrumentation__.Notify(85808)
	}
	__antithesis_instrumentation__.Notify(85798)
	prettyEnd := prettyPrintInternal(nil, end, false)
	i := 0

	for ; i < len(prettyStart) && func() bool {
		__antithesis_instrumentation__.Notify(85809)
		return i < len(prettyEnd) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(85810)
		return prettyStart[i] == prettyEnd[i] == true
	}() == true; i++ {
		__antithesis_instrumentation__.Notify(85811)
	}
	__antithesis_instrumentation__.Notify(85799)

	if i > maxChars-7 {
		__antithesis_instrumentation__.Notify(85812)
		if i > maxChars-1 {
			__antithesis_instrumentation__.Notify(85814)
			i = maxChars - 1
		} else {
			__antithesis_instrumentation__.Notify(85815)
		}
		__antithesis_instrumentation__.Notify(85813)
		copyEscape(&b, prettyStart[:i])
		b.WriteRune('…')
		return b.String()
	} else {
		__antithesis_instrumentation__.Notify(85816)
	}
	__antithesis_instrumentation__.Notify(85800)
	b.WriteString(prettyStart[:i])
	remaining := (maxChars - i - 3) / 2

	printTrunc := func(b *bytes.Buffer, what string, maxChars int) {
		__antithesis_instrumentation__.Notify(85817)
		if len(what) <= maxChars {
			__antithesis_instrumentation__.Notify(85818)
			copyEscape(b, what)
		} else {
			__antithesis_instrumentation__.Notify(85819)
			copyEscape(b, what[:maxChars-1])
			b.WriteRune('…')
		}
	}
	__antithesis_instrumentation__.Notify(85801)

	b.WriteByte('{')
	printTrunc(&b, prettyStart[i:], remaining)
	b.WriteByte('-')
	printTrunc(&b, prettyEnd[i:], remaining)
	b.WriteByte('}')

	return b.String()
}

func copyEscape(buf *bytes.Buffer, s string) {
	__antithesis_instrumentation__.Notify(85820)
	buf.Grow(len(s))

	k := 0
	for i := 0; i < len(s); i++ {
		__antithesis_instrumentation__.Notify(85822)
		c := s[i]
		if c < utf8.RuneSelf && func() bool {
			__antithesis_instrumentation__.Notify(85825)
			return strconv.IsPrint(rune(c)) == true
		}() == true {
			__antithesis_instrumentation__.Notify(85826)
			continue
		} else {
			__antithesis_instrumentation__.Notify(85827)
		}
		__antithesis_instrumentation__.Notify(85823)
		buf.WriteString(s[k:i])
		l, width := utf8.DecodeRuneInString(s[i:])
		if l == utf8.RuneError || func() bool {
			__antithesis_instrumentation__.Notify(85828)
			return l < 0x20 == true
		}() == true {
			__antithesis_instrumentation__.Notify(85829)
			const hex = "0123456789abcdef"
			buf.WriteByte('\\')
			buf.WriteByte('x')
			buf.WriteByte(hex[c>>4])
			buf.WriteByte(hex[c&0xf])
		} else {
			__antithesis_instrumentation__.Notify(85830)
			buf.WriteRune(l)
		}
		__antithesis_instrumentation__.Notify(85824)
		k = i + width
		i += width - 1
	}
	__antithesis_instrumentation__.Notify(85821)
	buf.WriteString(s[k:])
}
