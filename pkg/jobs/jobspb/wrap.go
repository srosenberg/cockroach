package jobspb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/protoreflect"
	"github.com/cockroachdb/errors"
	"github.com/gogo/protobuf/jsonpb"
)

type JobID = catpb.JobID

const InvalidJobID = catpb.InvalidJobID

type Details interface{}

var (
	_ Details = BackupDetails{}
	_ Details = RestoreDetails{}
	_ Details = SchemaChangeDetails{}
	_ Details = ChangefeedDetails{}
	_ Details = CreateStatsDetails{}
	_ Details = SchemaChangeGCDetails{}
	_ Details = StreamIngestionDetails{}
	_ Details = NewSchemaChangeDetails{}
	_ Details = MigrationDetails{}
	_ Details = AutoSpanConfigReconciliationDetails{}
	_ Details = ImportDetails{}
	_ Details = StreamReplicationDetails{}
	_ Details = RowLevelTTLDetails{}
)

type ProgressDetails interface{}

var (
	_ ProgressDetails = BackupProgress{}
	_ ProgressDetails = RestoreProgress{}
	_ ProgressDetails = SchemaChangeProgress{}
	_ ProgressDetails = ChangefeedProgress{}
	_ ProgressDetails = CreateStatsProgress{}
	_ ProgressDetails = SchemaChangeGCProgress{}
	_ ProgressDetails = StreamIngestionProgress{}
	_ ProgressDetails = NewSchemaChangeProgress{}
	_ ProgressDetails = MigrationProgress{}
	_ ProgressDetails = AutoSpanConfigReconciliationDetails{}
	_ ProgressDetails = StreamReplicationProgress{}
	_ ProgressDetails = RowLevelTTLProgress{}
)

func (p *Payload) Type() Type {
	__antithesis_instrumentation__.Notify(84031)
	return DetailsType(p.Details)
}

var _ base.SQLInstanceID

const AutoStatsName = "__auto__"

const ImportStatsName = "__import__"

var AutomaticJobTypes = [...]Type{
	TypeAutoCreateStats,
	TypeAutoSpanConfigReconciliation,
	TypeAutoSQLStatsCompaction,
}

func DetailsType(d isPayload_Details) Type {
	__antithesis_instrumentation__.Notify(84032)
	switch d := d.(type) {
	case *Payload_Backup:
		__antithesis_instrumentation__.Notify(84033)
		return TypeBackup
	case *Payload_Restore:
		__antithesis_instrumentation__.Notify(84034)
		return TypeRestore
	case *Payload_SchemaChange:
		__antithesis_instrumentation__.Notify(84035)
		return TypeSchemaChange
	case *Payload_Import:
		__antithesis_instrumentation__.Notify(84036)
		return TypeImport
	case *Payload_Changefeed:
		__antithesis_instrumentation__.Notify(84037)
		return TypeChangefeed
	case *Payload_CreateStats:
		__antithesis_instrumentation__.Notify(84038)
		createStatsName := d.CreateStats.Name
		if createStatsName == AutoStatsName {
			__antithesis_instrumentation__.Notify(84050)
			return TypeAutoCreateStats
		} else {
			__antithesis_instrumentation__.Notify(84051)
		}
		__antithesis_instrumentation__.Notify(84039)
		return TypeCreateStats
	case *Payload_SchemaChangeGC:
		__antithesis_instrumentation__.Notify(84040)
		return TypeSchemaChangeGC
	case *Payload_TypeSchemaChange:
		__antithesis_instrumentation__.Notify(84041)
		return TypeTypeSchemaChange
	case *Payload_StreamIngestion:
		__antithesis_instrumentation__.Notify(84042)
		return TypeStreamIngestion
	case *Payload_NewSchemaChange:
		__antithesis_instrumentation__.Notify(84043)
		return TypeNewSchemaChange
	case *Payload_Migration:
		__antithesis_instrumentation__.Notify(84044)
		return TypeMigration
	case *Payload_AutoSpanConfigReconciliation:
		__antithesis_instrumentation__.Notify(84045)
		return TypeAutoSpanConfigReconciliation
	case *Payload_AutoSQLStatsCompaction:
		__antithesis_instrumentation__.Notify(84046)
		return TypeAutoSQLStatsCompaction
	case *Payload_StreamReplication:
		__antithesis_instrumentation__.Notify(84047)
		return TypeStreamReplication
	case *Payload_RowLevelTTL:
		__antithesis_instrumentation__.Notify(84048)
		return TypeRowLevelTTL
	default:
		__antithesis_instrumentation__.Notify(84049)
		panic(errors.AssertionFailedf("Payload.Type called on a payload with an unknown details type: %T", d))
	}
}

func WrapProgressDetails(details ProgressDetails) interface {
	isProgress_Details
} {
	__antithesis_instrumentation__.Notify(84052)
	switch d := details.(type) {
	case BackupProgress:
		__antithesis_instrumentation__.Notify(84053)
		return &Progress_Backup{Backup: &d}
	case RestoreProgress:
		__antithesis_instrumentation__.Notify(84054)
		return &Progress_Restore{Restore: &d}
	case SchemaChangeProgress:
		__antithesis_instrumentation__.Notify(84055)
		return &Progress_SchemaChange{SchemaChange: &d}
	case ImportProgress:
		__antithesis_instrumentation__.Notify(84056)
		return &Progress_Import{Import: &d}
	case ChangefeedProgress:
		__antithesis_instrumentation__.Notify(84057)
		return &Progress_Changefeed{Changefeed: &d}
	case CreateStatsProgress:
		__antithesis_instrumentation__.Notify(84058)
		return &Progress_CreateStats{CreateStats: &d}
	case SchemaChangeGCProgress:
		__antithesis_instrumentation__.Notify(84059)
		return &Progress_SchemaChangeGC{SchemaChangeGC: &d}
	case TypeSchemaChangeProgress:
		__antithesis_instrumentation__.Notify(84060)
		return &Progress_TypeSchemaChange{TypeSchemaChange: &d}
	case StreamIngestionProgress:
		__antithesis_instrumentation__.Notify(84061)
		return &Progress_StreamIngest{StreamIngest: &d}
	case NewSchemaChangeProgress:
		__antithesis_instrumentation__.Notify(84062)
		return &Progress_NewSchemaChange{NewSchemaChange: &d}
	case MigrationProgress:
		__antithesis_instrumentation__.Notify(84063)
		return &Progress_Migration{Migration: &d}
	case AutoSpanConfigReconciliationProgress:
		__antithesis_instrumentation__.Notify(84064)
		return &Progress_AutoSpanConfigReconciliation{AutoSpanConfigReconciliation: &d}
	case AutoSQLStatsCompactionProgress:
		__antithesis_instrumentation__.Notify(84065)
		return &Progress_AutoSQLStatsCompaction{AutoSQLStatsCompaction: &d}
	case StreamReplicationProgress:
		__antithesis_instrumentation__.Notify(84066)
		return &Progress_StreamReplication{StreamReplication: &d}
	case RowLevelTTLProgress:
		__antithesis_instrumentation__.Notify(84067)
		return &Progress_RowLevelTTL{RowLevelTTL: &d}
	default:
		__antithesis_instrumentation__.Notify(84068)
		panic(errors.AssertionFailedf("WrapProgressDetails: unknown details type %T", d))
	}
}

func (p *Payload) UnwrapDetails() Details {
	__antithesis_instrumentation__.Notify(84069)
	switch d := p.Details.(type) {
	case *Payload_Backup:
		__antithesis_instrumentation__.Notify(84070)
		return *d.Backup
	case *Payload_Restore:
		__antithesis_instrumentation__.Notify(84071)
		return *d.Restore
	case *Payload_SchemaChange:
		__antithesis_instrumentation__.Notify(84072)
		return *d.SchemaChange
	case *Payload_Import:
		__antithesis_instrumentation__.Notify(84073)
		return *d.Import
	case *Payload_Changefeed:
		__antithesis_instrumentation__.Notify(84074)
		return *d.Changefeed
	case *Payload_CreateStats:
		__antithesis_instrumentation__.Notify(84075)
		return *d.CreateStats
	case *Payload_SchemaChangeGC:
		__antithesis_instrumentation__.Notify(84076)
		return *d.SchemaChangeGC
	case *Payload_TypeSchemaChange:
		__antithesis_instrumentation__.Notify(84077)
		return *d.TypeSchemaChange
	case *Payload_StreamIngestion:
		__antithesis_instrumentation__.Notify(84078)
		return *d.StreamIngestion
	case *Payload_NewSchemaChange:
		__antithesis_instrumentation__.Notify(84079)
		return *d.NewSchemaChange
	case *Payload_Migration:
		__antithesis_instrumentation__.Notify(84080)
		return *d.Migration
	case *Payload_AutoSpanConfigReconciliation:
		__antithesis_instrumentation__.Notify(84081)
		return *d.AutoSpanConfigReconciliation
	case *Payload_AutoSQLStatsCompaction:
		__antithesis_instrumentation__.Notify(84082)
		return *d.AutoSQLStatsCompaction
	case *Payload_StreamReplication:
		__antithesis_instrumentation__.Notify(84083)
		return *d.StreamReplication
	case *Payload_RowLevelTTL:
		__antithesis_instrumentation__.Notify(84084)
		return *d.RowLevelTTL
	default:
		__antithesis_instrumentation__.Notify(84085)
		return nil
	}
}

func (p *Progress) UnwrapDetails() ProgressDetails {
	__antithesis_instrumentation__.Notify(84086)
	switch d := p.Details.(type) {
	case *Progress_Backup:
		__antithesis_instrumentation__.Notify(84087)
		return *d.Backup
	case *Progress_Restore:
		__antithesis_instrumentation__.Notify(84088)
		return *d.Restore
	case *Progress_SchemaChange:
		__antithesis_instrumentation__.Notify(84089)
		return *d.SchemaChange
	case *Progress_Import:
		__antithesis_instrumentation__.Notify(84090)
		return *d.Import
	case *Progress_Changefeed:
		__antithesis_instrumentation__.Notify(84091)
		return *d.Changefeed
	case *Progress_CreateStats:
		__antithesis_instrumentation__.Notify(84092)
		return *d.CreateStats
	case *Progress_SchemaChangeGC:
		__antithesis_instrumentation__.Notify(84093)
		return *d.SchemaChangeGC
	case *Progress_TypeSchemaChange:
		__antithesis_instrumentation__.Notify(84094)
		return *d.TypeSchemaChange
	case *Progress_StreamIngest:
		__antithesis_instrumentation__.Notify(84095)
		return *d.StreamIngest
	case *Progress_NewSchemaChange:
		__antithesis_instrumentation__.Notify(84096)
		return *d.NewSchemaChange
	case *Progress_Migration:
		__antithesis_instrumentation__.Notify(84097)
		return *d.Migration
	case *Progress_AutoSpanConfigReconciliation:
		__antithesis_instrumentation__.Notify(84098)
		return *d.AutoSpanConfigReconciliation
	case *Progress_AutoSQLStatsCompaction:
		__antithesis_instrumentation__.Notify(84099)
		return *d.AutoSQLStatsCompaction
	case *Progress_StreamReplication:
		__antithesis_instrumentation__.Notify(84100)
		return *d.StreamReplication
	case *Progress_RowLevelTTL:
		__antithesis_instrumentation__.Notify(84101)
		return *d.RowLevelTTL
	default:
		__antithesis_instrumentation__.Notify(84102)
		return nil
	}
}

func (t Type) String() string {
	__antithesis_instrumentation__.Notify(84103)

	return strings.Replace(Type_name[int32(t)], "_", " ", -1)
}

func WrapPayloadDetails(details Details) interface {
	isPayload_Details
} {
	__antithesis_instrumentation__.Notify(84104)
	switch d := details.(type) {
	case BackupDetails:
		__antithesis_instrumentation__.Notify(84105)
		return &Payload_Backup{Backup: &d}
	case RestoreDetails:
		__antithesis_instrumentation__.Notify(84106)
		return &Payload_Restore{Restore: &d}
	case SchemaChangeDetails:
		__antithesis_instrumentation__.Notify(84107)
		return &Payload_SchemaChange{SchemaChange: &d}
	case ImportDetails:
		__antithesis_instrumentation__.Notify(84108)
		return &Payload_Import{Import: &d}
	case ChangefeedDetails:
		__antithesis_instrumentation__.Notify(84109)
		return &Payload_Changefeed{Changefeed: &d}
	case CreateStatsDetails:
		__antithesis_instrumentation__.Notify(84110)
		return &Payload_CreateStats{CreateStats: &d}
	case SchemaChangeGCDetails:
		__antithesis_instrumentation__.Notify(84111)
		return &Payload_SchemaChangeGC{SchemaChangeGC: &d}
	case TypeSchemaChangeDetails:
		__antithesis_instrumentation__.Notify(84112)
		return &Payload_TypeSchemaChange{TypeSchemaChange: &d}
	case StreamIngestionDetails:
		__antithesis_instrumentation__.Notify(84113)
		return &Payload_StreamIngestion{StreamIngestion: &d}
	case NewSchemaChangeDetails:
		__antithesis_instrumentation__.Notify(84114)
		return &Payload_NewSchemaChange{NewSchemaChange: &d}
	case MigrationDetails:
		__antithesis_instrumentation__.Notify(84115)
		return &Payload_Migration{Migration: &d}
	case AutoSpanConfigReconciliationDetails:
		__antithesis_instrumentation__.Notify(84116)
		return &Payload_AutoSpanConfigReconciliation{AutoSpanConfigReconciliation: &d}
	case AutoSQLStatsCompactionDetails:
		__antithesis_instrumentation__.Notify(84117)
		return &Payload_AutoSQLStatsCompaction{AutoSQLStatsCompaction: &d}
	case StreamReplicationDetails:
		__antithesis_instrumentation__.Notify(84118)
		return &Payload_StreamReplication{StreamReplication: &d}
	case RowLevelTTLDetails:
		__antithesis_instrumentation__.Notify(84119)
		return &Payload_RowLevelTTL{RowLevelTTL: &d}
	default:
		__antithesis_instrumentation__.Notify(84120)
		panic(errors.AssertionFailedf("jobs.WrapPayloadDetails: unknown details type %T", d))
	}
}

type ChangefeedTargets map[descpb.ID]ChangefeedTargetTable

type SchemaChangeDetailsFormatVersion uint32

const (
	BaseFormatVersion SchemaChangeDetailsFormatVersion = iota

	JobResumerFormatVersion

	DatabaseJobFormatVersion

	_ = BaseFormatVersion
)

func (Type) SafeValue() { __antithesis_instrumentation__.Notify(84121) }

const NumJobTypes = 17

func (m ChangefeedDetails) MarshalJSONPB(marshaller *jsonpb.Marshaler) ([]byte, error) {
	__antithesis_instrumentation__.Notify(84122)
	if protoreflect.ShouldRedact(marshaller) {
		__antithesis_instrumentation__.Notify(84124)
		var err error
		m.SinkURI, err = cloud.SanitizeExternalStorageURI(m.SinkURI, nil)
		if err != nil {
			__antithesis_instrumentation__.Notify(84125)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(84126)
		}
	} else {
		__antithesis_instrumentation__.Notify(84127)
	}
	__antithesis_instrumentation__.Notify(84123)
	return json.Marshal(m)
}

type DescRewriteMap map[descpb.ID]*DescriptorRewrite

func init() {
	if len(Type_name) != NumJobTypes {
		panic(fmt.Errorf("NumJobTypes (%d) does not match generated job type name map length (%d)",
			NumJobTypes, len(Type_name)))
	}

	protoreflect.RegisterShorthands((*Progress)(nil), "progress")
	protoreflect.RegisterShorthands((*Payload)(nil), "payload")
	protoreflect.RegisterShorthands((*ScheduleDetails)(nil), "schedule", "schedule_details")
	protoreflect.RegisterShorthands((*ExecutionArguments)(nil), "exec_args", "execution_args", "schedule_args")
}
