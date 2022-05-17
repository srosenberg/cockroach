package spanconfigtestutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts/ptpb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/datadriven"
	"github.com/stretchr/testify/require"
)

var spanRe = regexp.MustCompile(`^\[(\w+),\s??(\w+)\)$`)

var systemTargetRe = regexp.MustCompile(
	`^{(entire-keyspace)|(source=(\d*),\s??((target=(\d*))|all-tenant-keyspace-targets-set))}$`,
)

var configRe = regexp.MustCompile(`^(FALLBACK)|(^\w)$`)

func ParseSpan(t testing.TB, sp string) roachpb.Span {
	__antithesis_instrumentation__.Notify(241795)
	if !spanRe.MatchString(sp) {
		__antithesis_instrumentation__.Notify(241797)
		t.Fatalf("expected %s to match span regex", sp)
	} else {
		__antithesis_instrumentation__.Notify(241798)
	}
	__antithesis_instrumentation__.Notify(241796)

	matches := spanRe.FindStringSubmatch(sp)
	start, end := matches[1], matches[2]
	return roachpb.Span{
		Key:    roachpb.Key(start),
		EndKey: roachpb.Key(end),
	}
}

func parseSystemTarget(t testing.TB, systemTarget string) spanconfig.SystemTarget {
	__antithesis_instrumentation__.Notify(241799)
	if !systemTargetRe.MatchString(systemTarget) {
		__antithesis_instrumentation__.Notify(241803)
		t.Fatalf("expected %s to match system target regex", systemTargetRe)
	} else {
		__antithesis_instrumentation__.Notify(241804)
	}
	__antithesis_instrumentation__.Notify(241800)
	matches := systemTargetRe.FindStringSubmatch(systemTarget)

	if matches[1] == "entire-keyspace" {
		__antithesis_instrumentation__.Notify(241805)
		return spanconfig.MakeEntireKeyspaceTarget()
	} else {
		__antithesis_instrumentation__.Notify(241806)
	}
	__antithesis_instrumentation__.Notify(241801)

	sourceID, err := strconv.Atoi(matches[3])
	require.NoError(t, err)
	if matches[4] == "all-tenant-keyspace-targets-set" {
		__antithesis_instrumentation__.Notify(241807)
		return spanconfig.MakeAllTenantKeyspaceTargetsSet(roachpb.MakeTenantID(uint64(sourceID)))
	} else {
		__antithesis_instrumentation__.Notify(241808)
	}
	__antithesis_instrumentation__.Notify(241802)
	targetID, err := strconv.Atoi(matches[6])
	require.NoError(t, err)
	target, err := spanconfig.MakeTenantKeyspaceTarget(
		roachpb.MakeTenantID(uint64(sourceID)), roachpb.MakeTenantID(uint64(targetID)),
	)
	require.NoError(t, err)
	return target
}

func ParseTarget(t testing.TB, target string) spanconfig.Target {
	__antithesis_instrumentation__.Notify(241809)
	switch {
	case spanRe.MatchString(target):
		__antithesis_instrumentation__.Notify(241811)
		return spanconfig.MakeTargetFromSpan(ParseSpan(t, target))
	case systemTargetRe.MatchString(target):
		__antithesis_instrumentation__.Notify(241812)
		return spanconfig.MakeTargetFromSystemTarget(parseSystemTarget(t, target))
	default:
		__antithesis_instrumentation__.Notify(241813)
		t.Fatalf("expected %s to match span or system target regex", target)
	}
	__antithesis_instrumentation__.Notify(241810)
	panic("unreachable")
}

func ParseConfig(t testing.TB, conf string) roachpb.SpanConfig {
	__antithesis_instrumentation__.Notify(241814)
	if !configRe.MatchString(conf) {
		__antithesis_instrumentation__.Notify(241817)
		t.Fatalf("expected %s to match config regex", conf)
	} else {
		__antithesis_instrumentation__.Notify(241818)
	}
	__antithesis_instrumentation__.Notify(241815)
	matches := configRe.FindStringSubmatch(conf)

	var ts int64
	if matches[1] == "FALLBACK" {
		__antithesis_instrumentation__.Notify(241819)
		ts = -1
	} else {
		__antithesis_instrumentation__.Notify(241820)
		ts = int64(matches[2][0])
	}
	__antithesis_instrumentation__.Notify(241816)
	return roachpb.SpanConfig{
		GCPolicy: roachpb.GCPolicy{
			ProtectionPolicies: []roachpb.ProtectionPolicy{
				{
					ProtectedTimestamp: hlc.Timestamp{
						WallTime: ts,
					},
				},
			},
		},
	}
}

func ParseSpanConfigRecord(t testing.TB, conf string) spanconfig.Record {
	__antithesis_instrumentation__.Notify(241821)
	parts := strings.Split(conf, ":")
	if len(parts) != 2 {
		__antithesis_instrumentation__.Notify(241823)
		t.Fatalf("expected single %q separator", ":")
	} else {
		__antithesis_instrumentation__.Notify(241824)
	}
	__antithesis_instrumentation__.Notify(241822)
	record, err := spanconfig.MakeRecord(ParseTarget(t, parts[0]),
		ParseConfig(t, parts[1]))
	require.NoError(t, err)
	return record
}

func ParseKVAccessorGetArguments(t testing.TB, input string) []spanconfig.Target {
	__antithesis_instrumentation__.Notify(241825)
	var targets []spanconfig.Target
	for _, line := range strings.Split(input, "\n") {
		__antithesis_instrumentation__.Notify(241827)
		line = strings.TrimSpace(line)
		if line == "" {
			__antithesis_instrumentation__.Notify(241830)
			continue
		} else {
			__antithesis_instrumentation__.Notify(241831)
		}
		__antithesis_instrumentation__.Notify(241828)

		const spanPrefix = "span "
		const systemTargetPrefix = "system-target "
		switch {
		case strings.HasPrefix(line, spanPrefix):
			__antithesis_instrumentation__.Notify(241832)
			line = strings.TrimPrefix(line, spanPrefix)
		case strings.HasPrefix(line, systemTargetPrefix):
			__antithesis_instrumentation__.Notify(241833)
			line = strings.TrimPrefix(line, systemTargetPrefix)
		default:
			__antithesis_instrumentation__.Notify(241834)
			t.Fatalf(
				"malformed line %q, expected to find %q or %q prefix",
				line,
				spanPrefix,
				systemTargetPrefix,
			)
		}
		__antithesis_instrumentation__.Notify(241829)
		targets = append(targets, ParseTarget(t, line))
	}
	__antithesis_instrumentation__.Notify(241826)
	return targets
}

func ParseKVAccessorUpdateArguments(
	t testing.TB, input string,
) ([]spanconfig.Target, []spanconfig.Record) {
	__antithesis_instrumentation__.Notify(241835)
	var toDelete []spanconfig.Target
	var toUpsert []spanconfig.Record
	for _, line := range strings.Split(input, "\n") {
		__antithesis_instrumentation__.Notify(241837)
		line = strings.TrimSpace(line)
		if line == "" {
			__antithesis_instrumentation__.Notify(241839)
			continue
		} else {
			__antithesis_instrumentation__.Notify(241840)
		}
		__antithesis_instrumentation__.Notify(241838)

		const upsertPrefix, deletePrefix = "upsert ", "delete "
		switch {
		case strings.HasPrefix(line, deletePrefix):
			__antithesis_instrumentation__.Notify(241841)
			line = strings.TrimPrefix(line, line[:len(deletePrefix)])
			toDelete = append(toDelete, ParseTarget(t, line))
		case strings.HasPrefix(line, upsertPrefix):
			__antithesis_instrumentation__.Notify(241842)
			line = strings.TrimPrefix(line, line[:len(upsertPrefix)])
			toUpsert = append(toUpsert, ParseSpanConfigRecord(t, line))
		default:
			__antithesis_instrumentation__.Notify(241843)
			t.Fatalf("malformed line %q, expected to find prefix %q or %q",
				line, upsertPrefix, deletePrefix)
		}
	}
	__antithesis_instrumentation__.Notify(241836)
	return toDelete, toUpsert
}

func ParseStoreApplyArguments(t testing.TB, input string) (updates []spanconfig.Update) {
	__antithesis_instrumentation__.Notify(241844)
	for _, line := range strings.Split(input, "\n") {
		__antithesis_instrumentation__.Notify(241846)
		line = strings.TrimSpace(line)
		if line == "" {
			__antithesis_instrumentation__.Notify(241848)
			continue
		} else {
			__antithesis_instrumentation__.Notify(241849)
		}
		__antithesis_instrumentation__.Notify(241847)

		const setPrefix, deletePrefix = "set ", "delete "
		switch {
		case strings.HasPrefix(line, deletePrefix):
			__antithesis_instrumentation__.Notify(241850)
			line = strings.TrimPrefix(line, line[:len(deletePrefix)])
			del, err := spanconfig.Deletion(ParseTarget(t, line))
			require.NoError(t, err)
			updates = append(updates, del)
		case strings.HasPrefix(line, setPrefix):
			__antithesis_instrumentation__.Notify(241851)
			line = strings.TrimPrefix(line, line[:len(setPrefix)])
			entry := ParseSpanConfigRecord(t, line)
			updates = append(updates, spanconfig.Update(entry))
		default:
			__antithesis_instrumentation__.Notify(241852)
			t.Fatalf("malformed line %q, expected to find prefix %q or %q",
				line, setPrefix, deletePrefix)
		}
	}
	__antithesis_instrumentation__.Notify(241845)
	return updates
}

func PrintSpan(sp roachpb.Span) string {
	__antithesis_instrumentation__.Notify(241853)
	s := []string{
		sp.Key.String(),
		sp.EndKey.String(),
	}
	for i := range s {
		__antithesis_instrumentation__.Notify(241855)

		if strings.Contains(s[i], "\"") {
			__antithesis_instrumentation__.Notify(241856)
			var err error
			s[i], err = strconv.Unquote(s[i])
			if err != nil {
				__antithesis_instrumentation__.Notify(241857)
				panic(err)
			} else {
				__antithesis_instrumentation__.Notify(241858)
			}
		} else {
			__antithesis_instrumentation__.Notify(241859)
		}
	}
	__antithesis_instrumentation__.Notify(241854)
	return fmt.Sprintf("[%s,%s)", s[0], s[1])
}

func PrintTarget(t testing.TB, target spanconfig.Target) string {
	__antithesis_instrumentation__.Notify(241860)
	switch {
	case target.IsSpanTarget():
		__antithesis_instrumentation__.Notify(241862)
		return PrintSpan(target.GetSpan())
	case target.IsSystemTarget():
		__antithesis_instrumentation__.Notify(241863)
		return target.GetSystemTarget().String()
	default:
		__antithesis_instrumentation__.Notify(241864)
		t.Fatalf("unknown target type")
	}
	__antithesis_instrumentation__.Notify(241861)
	panic("unreachable")
}

func PrintSpanConfig(config roachpb.SpanConfig) string {
	__antithesis_instrumentation__.Notify(241865)

	conf := make([]string, 0, len(config.GCPolicy.ProtectionPolicies)*2)
	for i, policy := range config.GCPolicy.ProtectionPolicies {
		__antithesis_instrumentation__.Notify(241867)
		if i > 0 {
			__antithesis_instrumentation__.Notify(241869)
			conf = append(conf, "+")
		} else {
			__antithesis_instrumentation__.Notify(241870)
		}
		__antithesis_instrumentation__.Notify(241868)

		if policy.ProtectedTimestamp.WallTime == -1 {
			__antithesis_instrumentation__.Notify(241871)
			conf = append(conf, "FALLBACK")
		} else {
			__antithesis_instrumentation__.Notify(241872)
			conf = append(conf, fmt.Sprintf("%c", policy.ProtectedTimestamp.WallTime))
		}
	}
	__antithesis_instrumentation__.Notify(241866)
	return strings.Join(conf, "")
}

func PrintSpanConfigRecord(t testing.TB, record spanconfig.Record) string {
	__antithesis_instrumentation__.Notify(241873)
	return fmt.Sprintf("%s:%s", PrintTarget(t, record.GetTarget()), PrintSpanConfig(record.GetConfig()))
}

func PrintSystemSpanConfigDiffedAgainstDefault(conf roachpb.SpanConfig) string {
	__antithesis_instrumentation__.Notify(241874)
	if conf.Equal(roachpb.TestingDefaultSystemSpanConfiguration()) {
		__antithesis_instrumentation__.Notify(241877)
		return "default system span config"
	} else {
		__antithesis_instrumentation__.Notify(241878)
	}
	__antithesis_instrumentation__.Notify(241875)

	var diffs []string
	defaultSystemTargetConf := roachpb.TestingDefaultSystemSpanConfiguration()
	if !reflect.DeepEqual(conf.GCPolicy.ProtectionPolicies,
		defaultSystemTargetConf.GCPolicy.ProtectionPolicies) {
		__antithesis_instrumentation__.Notify(241879)
		sort.Slice(conf.GCPolicy.ProtectionPolicies, func(i, j int) bool {
			__antithesis_instrumentation__.Notify(241882)
			lhs := conf.GCPolicy.ProtectionPolicies[i].ProtectedTimestamp
			rhs := conf.GCPolicy.ProtectionPolicies[j].ProtectedTimestamp
			return lhs.Less(rhs)
		})
		__antithesis_instrumentation__.Notify(241880)
		protectionPolicies := make([]string, 0, len(conf.GCPolicy.ProtectionPolicies))
		for _, pp := range conf.GCPolicy.ProtectionPolicies {
			__antithesis_instrumentation__.Notify(241883)
			protectionPolicies = append(protectionPolicies, pp.String())
		}
		__antithesis_instrumentation__.Notify(241881)
		diffs = append(diffs, fmt.Sprintf("protection_policies=[%s]", strings.Join(protectionPolicies, " ")))
	} else {
		__antithesis_instrumentation__.Notify(241884)
	}
	__antithesis_instrumentation__.Notify(241876)
	return strings.Join(diffs, " ")
}

func PrintSpanConfigDiffedAgainstDefaults(conf roachpb.SpanConfig) string {
	__antithesis_instrumentation__.Notify(241885)
	if conf.Equal(roachpb.TestingDefaultSpanConfig()) {
		__antithesis_instrumentation__.Notify(241903)
		return "range default"
	} else {
		__antithesis_instrumentation__.Notify(241904)
	}
	__antithesis_instrumentation__.Notify(241886)
	if conf.Equal(roachpb.TestingSystemSpanConfig()) {
		__antithesis_instrumentation__.Notify(241905)
		return "range system"
	} else {
		__antithesis_instrumentation__.Notify(241906)
	}
	__antithesis_instrumentation__.Notify(241887)
	if conf.Equal(roachpb.TestingDatabaseSystemSpanConfig(true)) {
		__antithesis_instrumentation__.Notify(241907)
		return "database system (host)"
	} else {
		__antithesis_instrumentation__.Notify(241908)
	}
	__antithesis_instrumentation__.Notify(241888)
	if conf.Equal(roachpb.TestingDatabaseSystemSpanConfig(false)) {
		__antithesis_instrumentation__.Notify(241909)
		return "database system (tenant)"
	} else {
		__antithesis_instrumentation__.Notify(241910)
	}
	__antithesis_instrumentation__.Notify(241889)

	defaultConf := roachpb.TestingDefaultSpanConfig()
	var diffs []string
	if conf.RangeMaxBytes != defaultConf.RangeMaxBytes {
		__antithesis_instrumentation__.Notify(241911)
		diffs = append(diffs, fmt.Sprintf("range_max_bytes=%d", conf.RangeMaxBytes))
	} else {
		__antithesis_instrumentation__.Notify(241912)
	}
	__antithesis_instrumentation__.Notify(241890)
	if conf.RangeMinBytes != defaultConf.RangeMinBytes {
		__antithesis_instrumentation__.Notify(241913)
		diffs = append(diffs, fmt.Sprintf("range_min_bytes=%d", conf.RangeMinBytes))
	} else {
		__antithesis_instrumentation__.Notify(241914)
	}
	__antithesis_instrumentation__.Notify(241891)
	if conf.GCPolicy.TTLSeconds != defaultConf.GCPolicy.TTLSeconds {
		__antithesis_instrumentation__.Notify(241915)
		diffs = append(diffs, fmt.Sprintf("ttl_seconds=%d", conf.GCPolicy.TTLSeconds))
	} else {
		__antithesis_instrumentation__.Notify(241916)
	}
	__antithesis_instrumentation__.Notify(241892)
	if conf.GCPolicy.IgnoreStrictEnforcement != defaultConf.GCPolicy.IgnoreStrictEnforcement {
		__antithesis_instrumentation__.Notify(241917)
		diffs = append(diffs, fmt.Sprintf("ignore_strict_gc=%t", conf.GCPolicy.IgnoreStrictEnforcement))
	} else {
		__antithesis_instrumentation__.Notify(241918)
	}
	__antithesis_instrumentation__.Notify(241893)
	if conf.GlobalReads != defaultConf.GlobalReads {
		__antithesis_instrumentation__.Notify(241919)
		diffs = append(diffs, fmt.Sprintf("global_reads=%v", conf.GlobalReads))
	} else {
		__antithesis_instrumentation__.Notify(241920)
	}
	__antithesis_instrumentation__.Notify(241894)
	if conf.NumReplicas != defaultConf.NumReplicas {
		__antithesis_instrumentation__.Notify(241921)
		diffs = append(diffs, fmt.Sprintf("num_replicas=%d", conf.NumReplicas))
	} else {
		__antithesis_instrumentation__.Notify(241922)
	}
	__antithesis_instrumentation__.Notify(241895)
	if conf.NumVoters != defaultConf.NumVoters {
		__antithesis_instrumentation__.Notify(241923)
		diffs = append(diffs, fmt.Sprintf("num_voters=%d", conf.NumVoters))
	} else {
		__antithesis_instrumentation__.Notify(241924)
	}
	__antithesis_instrumentation__.Notify(241896)
	if conf.RangefeedEnabled != defaultConf.RangefeedEnabled {
		__antithesis_instrumentation__.Notify(241925)
		diffs = append(diffs, fmt.Sprintf("rangefeed_enabled=%t", conf.RangefeedEnabled))
	} else {
		__antithesis_instrumentation__.Notify(241926)
	}
	__antithesis_instrumentation__.Notify(241897)
	if !reflect.DeepEqual(conf.Constraints, defaultConf.Constraints) {
		__antithesis_instrumentation__.Notify(241927)
		diffs = append(diffs, fmt.Sprintf("constraints=%v", conf.Constraints))
	} else {
		__antithesis_instrumentation__.Notify(241928)
	}
	__antithesis_instrumentation__.Notify(241898)
	if !reflect.DeepEqual(conf.VoterConstraints, defaultConf.VoterConstraints) {
		__antithesis_instrumentation__.Notify(241929)
		diffs = append(diffs, fmt.Sprintf("voter_constraints=%v", conf.VoterConstraints))
	} else {
		__antithesis_instrumentation__.Notify(241930)
	}
	__antithesis_instrumentation__.Notify(241899)
	if !reflect.DeepEqual(conf.LeasePreferences, defaultConf.LeasePreferences) {
		__antithesis_instrumentation__.Notify(241931)
		diffs = append(diffs, fmt.Sprintf("lease_preferences=%v", conf.VoterConstraints))
	} else {
		__antithesis_instrumentation__.Notify(241932)
	}
	__antithesis_instrumentation__.Notify(241900)
	if !reflect.DeepEqual(conf.GCPolicy.ProtectionPolicies, defaultConf.GCPolicy.ProtectionPolicies) {
		__antithesis_instrumentation__.Notify(241933)
		sort.Slice(conf.GCPolicy.ProtectionPolicies, func(i, j int) bool {
			__antithesis_instrumentation__.Notify(241936)
			lhs := conf.GCPolicy.ProtectionPolicies[i].ProtectedTimestamp
			rhs := conf.GCPolicy.ProtectionPolicies[j].ProtectedTimestamp
			return lhs.Less(rhs)
		})
		__antithesis_instrumentation__.Notify(241934)
		protectionPolicies := make([]string, 0, len(conf.GCPolicy.ProtectionPolicies))
		for _, pp := range conf.GCPolicy.ProtectionPolicies {
			__antithesis_instrumentation__.Notify(241937)
			protectionPolicies = append(protectionPolicies, pp.String())
		}
		__antithesis_instrumentation__.Notify(241935)
		diffs = append(diffs, fmt.Sprintf("protection_policies=[%s]", strings.Join(protectionPolicies, " ")))
	} else {
		__antithesis_instrumentation__.Notify(241938)
	}
	__antithesis_instrumentation__.Notify(241901)
	if conf.ExcludeDataFromBackup != defaultConf.ExcludeDataFromBackup {
		__antithesis_instrumentation__.Notify(241939)
		diffs = append(diffs, fmt.Sprintf("exclude_data_from_backup=%v", conf.ExcludeDataFromBackup))
	} else {
		__antithesis_instrumentation__.Notify(241940)
	}
	__antithesis_instrumentation__.Notify(241902)

	return strings.Join(diffs, " ")
}

func MaybeLimitAndOffset(
	t *testing.T, d *datadriven.TestData, separator string, lines []string,
) string {
	__antithesis_instrumentation__.Notify(241941)
	var offset, limit int
	if d.HasArg("offset") {
		__antithesis_instrumentation__.Notify(241946)
		d.ScanArgs(t, "offset", &offset)
		require.True(t, offset >= 0)
		require.Truef(t, offset <= len(lines),
			"offset (%d) larger than number of lines (%d)", offset, len(lines))
	} else {
		__antithesis_instrumentation__.Notify(241947)
	}
	__antithesis_instrumentation__.Notify(241942)
	if d.HasArg("limit") {
		__antithesis_instrumentation__.Notify(241948)
		d.ScanArgs(t, "limit", &limit)
		require.True(t, limit >= 0)
	} else {
		__antithesis_instrumentation__.Notify(241949)
		limit = len(lines)
	}
	__antithesis_instrumentation__.Notify(241943)

	var output strings.Builder
	if offset > 0 && func() bool {
		__antithesis_instrumentation__.Notify(241950)
		return len(lines) > 0 == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(241951)
		return separator != "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(241952)
		output.WriteString(fmt.Sprintf("%s\n", separator))
	} else {
		__antithesis_instrumentation__.Notify(241953)
	}
	__antithesis_instrumentation__.Notify(241944)
	lines = lines[offset:]
	for i, line := range lines {
		__antithesis_instrumentation__.Notify(241954)
		if i == limit {
			__antithesis_instrumentation__.Notify(241956)
			if separator != "" {
				__antithesis_instrumentation__.Notify(241958)
				output.WriteString(fmt.Sprintf("%s\n", separator))
			} else {
				__antithesis_instrumentation__.Notify(241959)
			}
			__antithesis_instrumentation__.Notify(241957)
			break
		} else {
			__antithesis_instrumentation__.Notify(241960)
		}
		__antithesis_instrumentation__.Notify(241955)
		output.WriteString(fmt.Sprintf("%s\n", line))
	}
	__antithesis_instrumentation__.Notify(241945)

	return strings.TrimSpace(output.String())
}

type SplitPoint struct {
	RKey   roachpb.RKey
	Config roachpb.SpanConfig
}

type SplitPoints []SplitPoint

func (rs SplitPoints) String() string {
	__antithesis_instrumentation__.Notify(241961)
	var output strings.Builder
	for _, c := range rs {
		__antithesis_instrumentation__.Notify(241963)
		output.WriteString(fmt.Sprintf("%-42s %s\n", c.RKey.String(),
			PrintSpanConfigDiffedAgainstDefaults(c.Config)))
	}
	__antithesis_instrumentation__.Notify(241962)
	return output.String()
}

func GetSplitPoints(ctx context.Context, t testing.TB, reader spanconfig.StoreReader) SplitPoints {
	__antithesis_instrumentation__.Notify(241964)
	var splitPoints []SplitPoint
	splitKey := roachpb.RKeyMin
	for {
		__antithesis_instrumentation__.Notify(241966)
		splitKeyConf, err := reader.GetSpanConfigForKey(ctx, splitKey)
		require.NoError(t, err)

		splitPoints = append(splitPoints, SplitPoint{
			RKey:   splitKey,
			Config: splitKeyConf,
		})

		if !reader.NeedsSplit(ctx, splitKey, roachpb.RKeyMax) {
			__antithesis_instrumentation__.Notify(241968)
			break
		} else {
			__antithesis_instrumentation__.Notify(241969)
		}
		__antithesis_instrumentation__.Notify(241967)
		splitKey = reader.ComputeSplitKey(ctx, splitKey, roachpb.RKeyMax)
	}
	__antithesis_instrumentation__.Notify(241965)

	return splitPoints
}

func ParseProtectionTarget(t testing.TB, input string) *ptpb.Target {
	__antithesis_instrumentation__.Notify(241970)
	line := strings.Split(input, "\n")
	if len(line) != 1 {
		__antithesis_instrumentation__.Notify(241973)
		t.Fatal("only one target must be specified per protectedts operation")
	} else {
		__antithesis_instrumentation__.Notify(241974)
	}
	__antithesis_instrumentation__.Notify(241971)
	target := line[0]

	const clusterPrefix, tenantPrefix, schemaObjectPrefix = "cluster", "tenants", "descs"
	switch {
	case strings.HasPrefix(target, clusterPrefix):
		__antithesis_instrumentation__.Notify(241975)
		return ptpb.MakeClusterTarget()
	case strings.HasPrefix(target, tenantPrefix):
		__antithesis_instrumentation__.Notify(241976)
		target = strings.TrimPrefix(target, target[:len(tenantPrefix)+1])
		tenantIDs := strings.Split(target, ",")
		ids := make([]roachpb.TenantID, 0, len(tenantIDs))
		for _, tenID := range tenantIDs {
			__antithesis_instrumentation__.Notify(241981)
			id, err := strconv.Atoi(tenID)
			require.NoError(t, err)
			ids = append(ids, roachpb.MakeTenantID(uint64(id)))
		}
		__antithesis_instrumentation__.Notify(241977)
		return ptpb.MakeTenantsTarget(ids)
	case strings.HasPrefix(target, schemaObjectPrefix):
		__antithesis_instrumentation__.Notify(241978)
		target = strings.TrimPrefix(target, target[:len(schemaObjectPrefix)+1])
		schemaObjectIDs := strings.Split(target, ",")
		ids := make([]descpb.ID, 0, len(schemaObjectIDs))
		for _, tenID := range schemaObjectIDs {
			__antithesis_instrumentation__.Notify(241982)
			id, err := strconv.Atoi(tenID)
			require.NoError(t, err)
			ids = append(ids, descpb.ID(id))
		}
		__antithesis_instrumentation__.Notify(241979)
		return ptpb.MakeSchemaObjectsTarget(ids)
	default:
		__antithesis_instrumentation__.Notify(241980)
		t.Fatalf("malformed line %q, expected to find prefix %q, %q or %q", target, tenantPrefix,
			schemaObjectPrefix, clusterPrefix)
	}
	__antithesis_instrumentation__.Notify(241972)
	return nil
}
