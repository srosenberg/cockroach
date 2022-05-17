package diagnostics

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/build"
	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/diagnostics/diagnosticspb"
	"github.com/cockroachdb/cockroach/pkg/server/status/statuspb"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/util/httputil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/logcrash"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/mitchellh/reflectwalk"
	"google.golang.org/protobuf/proto"
)

type NodeStatusGenerator interface {
	GenerateNodeStatus(ctx context.Context) *statuspb.NodeStatus
}

var reportFrequency = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"diagnostics.reporting.interval",
	"interval at which diagnostics data should be reported",
	time.Hour,
	settings.NonNegativeDuration,
).WithPublic()

type Reporter struct {
	StartTime  time.Time
	AmbientCtx *log.AmbientContext
	Config     *base.Config
	Settings   *cluster.Settings

	StorageClusterID func() uuid.UUID
	TenantID         roachpb.TenantID

	LogicalClusterID func() uuid.UUID

	SQLInstanceID func() base.SQLInstanceID
	SQLServer     *sql.Server
	InternalExec  *sql.InternalExecutor
	DB            *kv.DB
	Recorder      NodeStatusGenerator

	Locality roachpb.Locality

	TestingKnobs *TestingKnobs
}

func (r *Reporter) PeriodicallyReportDiagnostics(ctx context.Context, stopper *stop.Stopper) {
	__antithesis_instrumentation__.Notify(193052)
	_ = stopper.RunAsyncTaskEx(ctx, stop.TaskOpts{TaskName: "diagnostics", SpanOpt: stop.SterileRootSpan}, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(193053)
		defer logcrash.RecoverAndReportNonfatalPanic(ctx, &r.Settings.SV)
		nextReport := r.StartTime

		var timer timeutil.Timer
		defer timer.Stop()
		for {
			__antithesis_instrumentation__.Notify(193054)

			if logcrash.DiagnosticsReportingEnabled.Get(&r.Settings.SV) {
				__antithesis_instrumentation__.Notify(193056)
				r.ReportDiagnostics(ctx)
			} else {
				__antithesis_instrumentation__.Notify(193057)
			}
			__antithesis_instrumentation__.Notify(193055)

			nextReport = nextReport.Add(reportFrequency.Get(&r.Settings.SV))

			timer.Reset(addJitter(nextReport.Sub(timeutil.Now())))
			select {
			case <-stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(193058)
				return
			case <-timer.C:
				__antithesis_instrumentation__.Notify(193059)
				timer.Read = true
			}
		}
	})
}

func (r *Reporter) ReportDiagnostics(ctx context.Context) {
	__antithesis_instrumentation__.Notify(193060)
	ctx, span := r.AmbientCtx.AnnotateCtxWithSpan(ctx, "usageReport")
	defer span.Finish()

	report := r.CreateReport(ctx, telemetry.ResetCounts)

	url := r.buildReportingURL(report)
	if url == nil {
		__antithesis_instrumentation__.Notify(193065)
		return
	} else {
		__antithesis_instrumentation__.Notify(193066)
	}
	__antithesis_instrumentation__.Notify(193061)

	b, err := protoutil.Marshal(report)
	if err != nil {
		__antithesis_instrumentation__.Notify(193067)
		log.Warningf(ctx, "%v", err)
		return
	} else {
		__antithesis_instrumentation__.Notify(193068)
	}
	__antithesis_instrumentation__.Notify(193062)

	res, err := httputil.Post(
		ctx, url.String(), "application/x-protobuf", bytes.NewReader(b),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(193069)
		if log.V(2) {
			__antithesis_instrumentation__.Notify(193071)

			log.Warningf(ctx, "failed to report node usage metrics: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(193072)
		}
		__antithesis_instrumentation__.Notify(193070)
		return
	} else {
		__antithesis_instrumentation__.Notify(193073)
	}
	__antithesis_instrumentation__.Notify(193063)
	defer res.Body.Close()
	b, err = ioutil.ReadAll(res.Body)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(193074)
		return res.StatusCode != http.StatusOK == true
	}() == true {
		__antithesis_instrumentation__.Notify(193075)
		log.Warningf(ctx, "failed to report node usage metrics: status: %s, body: %s, "+
			"error: %v", res.Status, b, err)
		return
	} else {
		__antithesis_instrumentation__.Notify(193076)
	}
	__antithesis_instrumentation__.Notify(193064)
	r.SQLServer.GetReportedSQLStatsController().ResetLocalSQLStats(ctx)
}

func (r *Reporter) CreateReport(
	ctx context.Context, reset telemetry.ResetCounters,
) *diagnosticspb.DiagnosticReport {
	__antithesis_instrumentation__.Notify(193077)
	info := diagnosticspb.DiagnosticReport{}
	secret := sql.ClusterSecret.Get(&r.Settings.SV)
	uptime := int64(timeutil.Since(r.StartTime).Seconds())

	r.populateEnvironment(ctx, secret, &info.Env)

	r.populateSQLInfo(uptime, &info.SQL)

	if r.TenantID == roachpb.SystemTenantID {
		__antithesis_instrumentation__.Notify(193083)
		r.populateNodeInfo(ctx, uptime, &info)
	} else {
		__antithesis_instrumentation__.Notify(193084)
	}
	__antithesis_instrumentation__.Notify(193078)

	schema, err := r.collectSchemaInfo(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(193085)
		log.Warningf(ctx, "error collecting schema info for diagnostic report: %+v", err)
		schema = nil
	} else {
		__antithesis_instrumentation__.Notify(193086)
	}
	__antithesis_instrumentation__.Notify(193079)
	info.Schema = schema

	info.FeatureUsage = telemetry.GetFeatureCounts(telemetry.Quantized, reset)

	if it, err := r.InternalExec.QueryIteratorEx(
		ctx, "read-setting", nil,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		"SELECT name FROM system.settings",
	); err != nil {
		__antithesis_instrumentation__.Notify(193087)
		log.Warningf(ctx, "failed to read settings: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(193088)
		info.AlteredSettings = make(map[string]string)
		var ok bool
		for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
			__antithesis_instrumentation__.Notify(193090)
			row := it.Cur()
			name := string(tree.MustBeDString(row[0]))
			info.AlteredSettings[name] = settings.RedactedValue(
				name, &r.Settings.SV, r.TenantID == roachpb.SystemTenantID,
			)
		}
		__antithesis_instrumentation__.Notify(193089)
		if err != nil {
			__antithesis_instrumentation__.Notify(193091)

			log.Warningf(ctx, "failed to read settings: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(193092)
		}
	}
	__antithesis_instrumentation__.Notify(193080)

	if it, err := r.InternalExec.QueryIteratorEx(
		ctx,
		"read-zone-configs",
		nil,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		"SELECT id, config FROM system.zones",
	); err != nil {
		__antithesis_instrumentation__.Notify(193093)
		log.Warningf(ctx, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(193094)
		info.ZoneConfigs = make(map[int64]zonepb.ZoneConfig)
		var ok bool
		for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
			__antithesis_instrumentation__.Notify(193096)
			row := it.Cur()
			id := int64(tree.MustBeDInt(row[0]))
			var zone zonepb.ZoneConfig
			if bytes, ok := row[1].(*tree.DBytes); !ok {
				__antithesis_instrumentation__.Notify(193098)
				continue
			} else {
				__antithesis_instrumentation__.Notify(193099)
				if err := protoutil.Unmarshal([]byte(*bytes), &zone); err != nil {
					__antithesis_instrumentation__.Notify(193100)
					log.Warningf(ctx, "unable to parse zone config %d: %v", id, err)
					continue
				} else {
					__antithesis_instrumentation__.Notify(193101)
				}
			}
			__antithesis_instrumentation__.Notify(193097)
			var anonymizedZone zonepb.ZoneConfig
			anonymizeZoneConfig(&anonymizedZone, zone, secret)
			info.ZoneConfigs[id] = anonymizedZone
		}
		__antithesis_instrumentation__.Notify(193095)
		if err != nil {
			__antithesis_instrumentation__.Notify(193102)

			log.Warningf(ctx, "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(193103)
		}
	}
	__antithesis_instrumentation__.Notify(193081)

	info.SqlStats, err = r.SQLServer.GetScrubbedReportingStats(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(193104)
		if log.V(2) {
			__antithesis_instrumentation__.Notify(193105)
			log.Warningf(ctx, "unexpected error encountered when getting scrubbed reporting stats: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(193106)
		}
	} else {
		__antithesis_instrumentation__.Notify(193107)
	}
	__antithesis_instrumentation__.Notify(193082)

	return &info
}

func (r *Reporter) populateEnvironment(
	ctx context.Context, secret string, env *diagnosticspb.Environment,
) {
	__antithesis_instrumentation__.Notify(193108)
	env.Build = build.GetInfo()
	env.LicenseType = getLicenseType(ctx, r.Settings)
	populateHardwareInfo(ctx, env)

	for _, tier := range r.Locality.Tiers {
		__antithesis_instrumentation__.Notify(193109)
		env.Locality.Tiers = append(env.Locality.Tiers, roachpb.Tier{
			Key:   sql.HashForReporting(secret, tier.Key),
			Value: sql.HashForReporting(secret, tier.Value),
		})
	}
}

func (r *Reporter) populateNodeInfo(
	ctx context.Context, uptime int64, info *diagnosticspb.DiagnosticReport,
) {
	__antithesis_instrumentation__.Notify(193110)
	n := r.Recorder.GenerateNodeStatus(ctx)
	info.Node.NodeID = n.Desc.NodeID
	info.Node.Uptime = uptime

	info.Stores = make([]diagnosticspb.StoreInfo, len(n.StoreStatuses))
	for i, r := range n.StoreStatuses {
		__antithesis_instrumentation__.Notify(193111)
		info.Stores[i].NodeID = r.Desc.Node.NodeID
		info.Stores[i].StoreID = r.Desc.StoreID
		info.Stores[i].KeyCount = int64(r.Metrics["keycount"])
		info.Stores[i].Capacity = int64(r.Metrics["capacity"])
		info.Stores[i].Available = int64(r.Metrics["capacity.available"])
		info.Stores[i].Used = int64(r.Metrics["capacity.used"])
		info.Node.KeyCount += info.Stores[i].KeyCount
		info.Stores[i].RangeCount = int64(r.Metrics["replicas"])
		info.Node.RangeCount += info.Stores[i].RangeCount
		bytes := int64(r.Metrics["sysbytes"] + r.Metrics["intentbytes"] + r.Metrics["valbytes"] + r.Metrics["keybytes"])
		info.Stores[i].Bytes = bytes
		info.Node.Bytes += bytes
		info.Stores[i].EncryptionAlgorithm = int64(r.Metrics["rocksdb.encryption.algorithm"])
	}
}

func (r *Reporter) populateSQLInfo(uptime int64, sql *diagnosticspb.SQLInstanceInfo) {
	__antithesis_instrumentation__.Notify(193112)
	sql.SQLInstanceID = r.SQLInstanceID()
	sql.Uptime = uptime
}

func (r *Reporter) collectSchemaInfo(ctx context.Context) ([]descpb.TableDescriptor, error) {
	__antithesis_instrumentation__.Notify(193113)
	startKey := keys.MakeSQLCodec(r.TenantID).TablePrefix(keys.DescriptorTableID)
	endKey := startKey.PrefixEnd()
	kvs, err := r.DB.Scan(ctx, startKey, endKey, 0)
	if err != nil {
		__antithesis_instrumentation__.Notify(193116)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(193117)
	}
	__antithesis_instrumentation__.Notify(193114)
	tables := make([]descpb.TableDescriptor, 0, len(kvs))
	redactor := stringRedactor{}
	for _, kv := range kvs {
		__antithesis_instrumentation__.Notify(193118)
		var desc descpb.Descriptor
		if err := kv.ValueProto(&desc); err != nil {
			__antithesis_instrumentation__.Notify(193120)
			return nil, errors.Wrapf(err, "%s: unable to unmarshal SQL descriptor", kv.Key)
		} else {
			__antithesis_instrumentation__.Notify(193121)
		}
		__antithesis_instrumentation__.Notify(193119)
		t, _, _, _ := descpb.FromDescriptorWithMVCCTimestamp(&desc, kv.Value.Timestamp)
		if t != nil && func() bool {
			__antithesis_instrumentation__.Notify(193122)
			return t.ParentID != keys.SystemDatabaseID == true
		}() == true {
			__antithesis_instrumentation__.Notify(193123)
			if err := reflectwalk.Walk(t, redactor); err != nil {
				__antithesis_instrumentation__.Notify(193125)
				panic(err)
			} else {
				__antithesis_instrumentation__.Notify(193126)
			}
			__antithesis_instrumentation__.Notify(193124)
			tables = append(tables, *t)
		} else {
			__antithesis_instrumentation__.Notify(193127)
		}
	}
	__antithesis_instrumentation__.Notify(193115)
	return tables, nil
}

func (r *Reporter) buildReportingURL(report *diagnosticspb.DiagnosticReport) *url.URL {
	__antithesis_instrumentation__.Notify(193128)
	clusterInfo := ClusterInfo{
		StorageClusterID: r.StorageClusterID(),
		LogicalClusterID: r.LogicalClusterID(),
		TenantID:         r.TenantID,
		IsInsecure:       r.Config.Insecure,
		IsInternal:       sql.ClusterIsInternal(&r.Settings.SV),
	}

	url := reportingURL
	if r.TestingKnobs != nil && func() bool {
		__antithesis_instrumentation__.Notify(193130)
		return r.TestingKnobs.OverrideReportingURL != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(193131)
		url = *r.TestingKnobs.OverrideReportingURL
	} else {
		__antithesis_instrumentation__.Notify(193132)
	}
	__antithesis_instrumentation__.Notify(193129)
	return addInfoToURL(url, &clusterInfo, &report.Env, report.Node.NodeID, &report.SQL)
}

func getLicenseType(ctx context.Context, settings *cluster.Settings) string {
	__antithesis_instrumentation__.Notify(193133)
	licenseType, err := base.LicenseType(settings)
	if err != nil {
		__antithesis_instrumentation__.Notify(193135)
		log.Errorf(ctx, "error retrieving license type: %s", err)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(193136)
	}
	__antithesis_instrumentation__.Notify(193134)
	return licenseType
}

func anonymizeZoneConfig(dst *zonepb.ZoneConfig, src zonepb.ZoneConfig, secret string) {
	__antithesis_instrumentation__.Notify(193137)
	if src.RangeMinBytes != nil {
		__antithesis_instrumentation__.Notify(193145)
		dst.RangeMinBytes = proto.Int64(*src.RangeMinBytes)
	} else {
		__antithesis_instrumentation__.Notify(193146)
	}
	__antithesis_instrumentation__.Notify(193138)
	if src.RangeMaxBytes != nil {
		__antithesis_instrumentation__.Notify(193147)
		dst.RangeMaxBytes = proto.Int64(*src.RangeMaxBytes)
	} else {
		__antithesis_instrumentation__.Notify(193148)
	}
	__antithesis_instrumentation__.Notify(193139)
	if src.GC != nil {
		__antithesis_instrumentation__.Notify(193149)
		dst.GC = &zonepb.GCPolicy{TTLSeconds: src.GC.TTLSeconds}
	} else {
		__antithesis_instrumentation__.Notify(193150)
	}
	__antithesis_instrumentation__.Notify(193140)
	if src.NumReplicas != nil {
		__antithesis_instrumentation__.Notify(193151)
		dst.NumReplicas = proto.Int32(*src.NumReplicas)
	} else {
		__antithesis_instrumentation__.Notify(193152)
	}
	__antithesis_instrumentation__.Notify(193141)
	dst.Constraints = make([]zonepb.ConstraintsConjunction, len(src.Constraints))
	dst.InheritedConstraints = src.InheritedConstraints
	for i := range src.Constraints {
		__antithesis_instrumentation__.Notify(193153)
		dst.Constraints[i].NumReplicas = src.Constraints[i].NumReplicas
		dst.Constraints[i].Constraints = make([]zonepb.Constraint, len(src.Constraints[i].Constraints))
		for j := range src.Constraints[i].Constraints {
			__antithesis_instrumentation__.Notify(193154)
			dst.Constraints[i].Constraints[j].Type = src.Constraints[i].Constraints[j].Type
			if key := src.Constraints[i].Constraints[j].Key; key != "" {
				__antithesis_instrumentation__.Notify(193156)
				dst.Constraints[i].Constraints[j].Key = sql.HashForReporting(secret, key)
			} else {
				__antithesis_instrumentation__.Notify(193157)
			}
			__antithesis_instrumentation__.Notify(193155)
			if val := src.Constraints[i].Constraints[j].Value; val != "" {
				__antithesis_instrumentation__.Notify(193158)
				dst.Constraints[i].Constraints[j].Value = sql.HashForReporting(secret, val)
			} else {
				__antithesis_instrumentation__.Notify(193159)
			}
		}
	}
	__antithesis_instrumentation__.Notify(193142)
	dst.VoterConstraints = make([]zonepb.ConstraintsConjunction, len(src.VoterConstraints))
	dst.NullVoterConstraintsIsEmpty = src.NullVoterConstraintsIsEmpty
	for i := range src.VoterConstraints {
		__antithesis_instrumentation__.Notify(193160)
		dst.VoterConstraints[i].NumReplicas = src.VoterConstraints[i].NumReplicas
		dst.VoterConstraints[i].Constraints = make([]zonepb.Constraint, len(src.VoterConstraints[i].Constraints))
		for j := range src.VoterConstraints[i].Constraints {
			__antithesis_instrumentation__.Notify(193161)
			dst.VoterConstraints[i].Constraints[j].Type = src.VoterConstraints[i].Constraints[j].Type
			if key := src.VoterConstraints[i].Constraints[j].Key; key != "" {
				__antithesis_instrumentation__.Notify(193163)
				dst.VoterConstraints[i].Constraints[j].Key = sql.HashForReporting(secret, key)
			} else {
				__antithesis_instrumentation__.Notify(193164)
			}
			__antithesis_instrumentation__.Notify(193162)
			if val := src.VoterConstraints[i].Constraints[j].Value; val != "" {
				__antithesis_instrumentation__.Notify(193165)
				dst.VoterConstraints[i].Constraints[j].Value = sql.HashForReporting(secret, val)
			} else {
				__antithesis_instrumentation__.Notify(193166)
			}
		}
	}
	__antithesis_instrumentation__.Notify(193143)
	dst.LeasePreferences = make([]zonepb.LeasePreference, len(src.LeasePreferences))
	dst.InheritedLeasePreferences = src.InheritedLeasePreferences
	for i := range src.LeasePreferences {
		__antithesis_instrumentation__.Notify(193167)
		dst.LeasePreferences[i].Constraints = make([]zonepb.Constraint, len(src.LeasePreferences[i].Constraints))
		for j := range src.LeasePreferences[i].Constraints {
			__antithesis_instrumentation__.Notify(193168)
			dst.LeasePreferences[i].Constraints[j].Type = src.LeasePreferences[i].Constraints[j].Type
			if key := src.LeasePreferences[i].Constraints[j].Key; key != "" {
				__antithesis_instrumentation__.Notify(193170)
				dst.LeasePreferences[i].Constraints[j].Key = sql.HashForReporting(secret, key)
			} else {
				__antithesis_instrumentation__.Notify(193171)
			}
			__antithesis_instrumentation__.Notify(193169)
			if val := src.LeasePreferences[i].Constraints[j].Value; val != "" {
				__antithesis_instrumentation__.Notify(193172)
				dst.LeasePreferences[i].Constraints[j].Value = sql.HashForReporting(secret, val)
			} else {
				__antithesis_instrumentation__.Notify(193173)
			}
		}
	}
	__antithesis_instrumentation__.Notify(193144)
	dst.Subzones = make([]zonepb.Subzone, len(src.Subzones))
	for i := range src.Subzones {
		__antithesis_instrumentation__.Notify(193174)
		dst.Subzones[i].IndexID = src.Subzones[i].IndexID
		dst.Subzones[i].PartitionName = sql.HashForReporting(secret, src.Subzones[i].PartitionName)
		anonymizeZoneConfig(&dst.Subzones[i].Config, src.Subzones[i].Config, secret)
	}
}

type stringRedactor struct{}

func (stringRedactor) Primitive(v reflect.Value) error {
	__antithesis_instrumentation__.Notify(193175)
	if v.Kind() == reflect.String && func() bool {
		__antithesis_instrumentation__.Notify(193177)
		return v.String() != "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(193178)
		v.Set(reflect.ValueOf("_").Convert(v.Type()))
	} else {
		__antithesis_instrumentation__.Notify(193179)
	}
	__antithesis_instrumentation__.Notify(193176)
	return nil
}
