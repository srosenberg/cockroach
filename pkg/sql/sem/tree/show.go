package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
)

type ShowVar struct {
	Name string
}

func (node *ShowVar) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613467)
	ctx.WriteString("SHOW ")

	ctx.WithFlags(ctx.flags & ^FmtAnonymize & ^FmtMarkRedactionNode, func() {
		__antithesis_instrumentation__.Notify(613468)
		ctx.FormatNameP(&node.Name)
	})
}

type ShowClusterSetting struct {
	Name string
}

func (node *ShowClusterSetting) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613469)
	ctx.WriteString("SHOW CLUSTER SETTING ")

	ctx.WithFlags(ctx.flags & ^FmtAnonymize & ^FmtMarkRedactionNode, func() {
		__antithesis_instrumentation__.Notify(613470)
		ctx.FormatNameP(&node.Name)
	})
}

type ShowClusterSettingList struct {
	All bool
}

func (node *ShowClusterSettingList) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613471)
	ctx.WriteString("SHOW ")
	qual := "PUBLIC"
	if node.All {
		__antithesis_instrumentation__.Notify(613473)
		qual = "ALL"
	} else {
		__antithesis_instrumentation__.Notify(613474)
	}
	__antithesis_instrumentation__.Notify(613472)
	ctx.WriteString(qual)
	ctx.WriteString(" CLUSTER SETTINGS")
}

type ShowBackupDetails int

const (
	BackupDefaultDetails ShowBackupDetails = iota

	BackupRangeDetails

	BackupFileDetails

	BackupSchemaDetails
)

type ShowBackup struct {
	Path         Expr
	InCollection Expr
	From         bool
	Details      ShowBackupDetails
	Options      KVOptions
}

func (node *ShowBackup) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613475)
	if node.InCollection != nil && func() bool {
		__antithesis_instrumentation__.Notify(613480)
		return node.Path == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(613481)
		ctx.WriteString("SHOW BACKUPS IN ")
		ctx.FormatNode(node.InCollection)
		return
	} else {
		__antithesis_instrumentation__.Notify(613482)
	}
	__antithesis_instrumentation__.Notify(613476)
	ctx.WriteString("SHOW BACKUP ")

	switch node.Details {
	case BackupRangeDetails:
		__antithesis_instrumentation__.Notify(613483)
		ctx.WriteString("RANGES ")
	case BackupFileDetails:
		__antithesis_instrumentation__.Notify(613484)
		ctx.WriteString("FILES ")
	case BackupSchemaDetails:
		__antithesis_instrumentation__.Notify(613485)
		ctx.WriteString("SCHEMAS ")
	default:
		__antithesis_instrumentation__.Notify(613486)
	}
	__antithesis_instrumentation__.Notify(613477)

	if node.From {
		__antithesis_instrumentation__.Notify(613487)
		ctx.WriteString("FROM ")
	} else {
		__antithesis_instrumentation__.Notify(613488)
	}
	__antithesis_instrumentation__.Notify(613478)

	ctx.FormatNode(node.Path)
	if node.InCollection != nil {
		__antithesis_instrumentation__.Notify(613489)
		ctx.WriteString(" IN ")
		ctx.FormatNode(node.InCollection)
	} else {
		__antithesis_instrumentation__.Notify(613490)
	}
	__antithesis_instrumentation__.Notify(613479)
	if len(node.Options) > 0 {
		__antithesis_instrumentation__.Notify(613491)
		ctx.WriteString(" WITH ")
		ctx.FormatNode(&node.Options)
	} else {
		__antithesis_instrumentation__.Notify(613492)
	}
}

type ShowColumns struct {
	Table       *UnresolvedObjectName
	WithComment bool
}

func (node *ShowColumns) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613493)
	ctx.WriteString("SHOW COLUMNS FROM ")
	ctx.FormatNode(node.Table)

	if node.WithComment {
		__antithesis_instrumentation__.Notify(613494)
		ctx.WriteString(" WITH COMMENT")
	} else {
		__antithesis_instrumentation__.Notify(613495)
	}
}

type ShowDatabases struct {
	WithComment bool
}

func (node *ShowDatabases) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613496)
	ctx.WriteString("SHOW DATABASES")

	if node.WithComment {
		__antithesis_instrumentation__.Notify(613497)
		ctx.WriteString(" WITH COMMENT")
	} else {
		__antithesis_instrumentation__.Notify(613498)
	}
}

type ShowEnums struct {
	ObjectNamePrefix
}

func (node *ShowEnums) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613499)
	ctx.WriteString("SHOW ENUMS")
}

type ShowTypes struct{}

func (node *ShowTypes) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613500)
	ctx.WriteString("SHOW TYPES")
}

type ShowTraceType string

const (
	ShowTraceRaw     ShowTraceType = "TRACE"
	ShowTraceKV      ShowTraceType = "KV TRACE"
	ShowTraceReplica ShowTraceType = "EXPERIMENTAL_REPLICA TRACE"
)

type ShowTraceForSession struct {
	TraceType ShowTraceType
	Compact   bool
}

func (node *ShowTraceForSession) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613501)
	ctx.WriteString("SHOW ")
	if node.Compact {
		__antithesis_instrumentation__.Notify(613503)
		ctx.WriteString("COMPACT ")
	} else {
		__antithesis_instrumentation__.Notify(613504)
	}
	__antithesis_instrumentation__.Notify(613502)
	ctx.WriteString(string(node.TraceType))
	ctx.WriteString(" FOR SESSION")
}

type ShowIndexes struct {
	Table       *UnresolvedObjectName
	WithComment bool
}

func (node *ShowIndexes) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613505)
	ctx.WriteString("SHOW INDEXES FROM ")
	ctx.FormatNode(node.Table)

	if node.WithComment {
		__antithesis_instrumentation__.Notify(613506)
		ctx.WriteString(" WITH COMMENT")
	} else {
		__antithesis_instrumentation__.Notify(613507)
	}
}

type ShowDatabaseIndexes struct {
	Database    Name
	WithComment bool
}

func (node *ShowDatabaseIndexes) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613508)
	ctx.WriteString("SHOW INDEXES FROM DATABASE ")
	ctx.FormatNode(&node.Database)

	if node.WithComment {
		__antithesis_instrumentation__.Notify(613509)
		ctx.WriteString(" WITH COMMENT")
	} else {
		__antithesis_instrumentation__.Notify(613510)
	}
}

type ShowQueries struct {
	All     bool
	Cluster bool
}

func (node *ShowQueries) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613511)
	ctx.WriteString("SHOW ")
	if node.All {
		__antithesis_instrumentation__.Notify(613513)
		ctx.WriteString("ALL ")
	} else {
		__antithesis_instrumentation__.Notify(613514)
	}
	__antithesis_instrumentation__.Notify(613512)
	if node.Cluster {
		__antithesis_instrumentation__.Notify(613515)
		ctx.WriteString("CLUSTER STATEMENTS")
	} else {
		__antithesis_instrumentation__.Notify(613516)
		ctx.WriteString("LOCAL STATEMENTS")
	}
}

type ShowJobs struct {
	Jobs *Select

	Automatic bool

	Block bool

	Schedules *Select
}

func (node *ShowJobs) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613517)
	ctx.WriteString("SHOW ")
	if node.Automatic {
		__antithesis_instrumentation__.Notify(613521)
		ctx.WriteString("AUTOMATIC ")
	} else {
		__antithesis_instrumentation__.Notify(613522)
	}
	__antithesis_instrumentation__.Notify(613518)
	ctx.WriteString("JOBS")
	if node.Block {
		__antithesis_instrumentation__.Notify(613523)
		ctx.WriteString(" WHEN COMPLETE")
	} else {
		__antithesis_instrumentation__.Notify(613524)
	}
	__antithesis_instrumentation__.Notify(613519)
	if node.Jobs != nil {
		__antithesis_instrumentation__.Notify(613525)
		ctx.WriteString(" ")
		ctx.FormatNode(node.Jobs)
	} else {
		__antithesis_instrumentation__.Notify(613526)
	}
	__antithesis_instrumentation__.Notify(613520)
	if node.Schedules != nil {
		__antithesis_instrumentation__.Notify(613527)
		ctx.WriteString(" FOR SCHEDULES ")
		ctx.FormatNode(node.Schedules)
	} else {
		__antithesis_instrumentation__.Notify(613528)
	}
}

type ShowChangefeedJobs struct {
	Jobs *Select
}

func (node *ShowChangefeedJobs) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613529)
	ctx.WriteString("SHOW CHANGEFEED JOBS")
	if node.Jobs != nil {
		__antithesis_instrumentation__.Notify(613530)
		ctx.WriteString(" ")
		ctx.FormatNode(node.Jobs)
	} else {
		__antithesis_instrumentation__.Notify(613531)
	}
}

type ShowSurvivalGoal struct {
	DatabaseName Name
}

func (node *ShowSurvivalGoal) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613532)
	ctx.WriteString("SHOW SURVIVAL GOAL FROM DATABASE")
	if node.DatabaseName != "" {
		__antithesis_instrumentation__.Notify(613533)
		ctx.WriteString(" ")
		ctx.FormatNode(&node.DatabaseName)
	} else {
		__antithesis_instrumentation__.Notify(613534)
	}
}

type ShowRegionsFrom int

const (
	ShowRegionsFromCluster ShowRegionsFrom = iota

	ShowRegionsFromDatabase

	ShowRegionsFromAllDatabases

	ShowRegionsFromDefault

	ShowSuperRegionsFromDatabase
)

type ShowRegions struct {
	ShowRegionsFrom ShowRegionsFrom
	DatabaseName    Name
}

func (node *ShowRegions) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613535)
	if node.ShowRegionsFrom == ShowSuperRegionsFromDatabase {
		__antithesis_instrumentation__.Notify(613537)
		ctx.WriteString("SHOW SUPER REGIONS")
	} else {
		__antithesis_instrumentation__.Notify(613538)
		ctx.WriteString("SHOW REGIONS")
	}
	__antithesis_instrumentation__.Notify(613536)
	switch node.ShowRegionsFrom {
	case ShowRegionsFromDefault:
		__antithesis_instrumentation__.Notify(613539)
	case ShowRegionsFromAllDatabases:
		__antithesis_instrumentation__.Notify(613540)
		ctx.WriteString(" FROM ALL DATABASES")
	case ShowRegionsFromDatabase, ShowSuperRegionsFromDatabase:
		__antithesis_instrumentation__.Notify(613541)
		ctx.WriteString(" FROM DATABASE")
		if node.DatabaseName != "" {
			__antithesis_instrumentation__.Notify(613544)
			ctx.WriteString(" ")
			ctx.FormatNode(&node.DatabaseName)
		} else {
			__antithesis_instrumentation__.Notify(613545)
		}
	case ShowRegionsFromCluster:
		__antithesis_instrumentation__.Notify(613542)
		ctx.WriteString(" FROM CLUSTER")
	default:
		__antithesis_instrumentation__.Notify(613543)
		panic(fmt.Sprintf("unknown ShowRegionsFrom: %v", node.ShowRegionsFrom))
	}
}

type ShowSessions struct {
	All     bool
	Cluster bool
}

func (node *ShowSessions) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613546)
	ctx.WriteString("SHOW ")
	if node.All {
		__antithesis_instrumentation__.Notify(613548)
		ctx.WriteString("ALL ")
	} else {
		__antithesis_instrumentation__.Notify(613549)
	}
	__antithesis_instrumentation__.Notify(613547)
	if node.Cluster {
		__antithesis_instrumentation__.Notify(613550)
		ctx.WriteString("CLUSTER SESSIONS")
	} else {
		__antithesis_instrumentation__.Notify(613551)
		ctx.WriteString("LOCAL SESSIONS")
	}
}

type ShowSchemas struct {
	Database Name
}

func (node *ShowSchemas) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613552)
	ctx.WriteString("SHOW SCHEMAS")
	if node.Database != "" {
		__antithesis_instrumentation__.Notify(613553)
		ctx.WriteString(" FROM ")
		ctx.FormatNode(&node.Database)
	} else {
		__antithesis_instrumentation__.Notify(613554)
	}
}

type ShowSequences struct {
	Database Name
}

func (node *ShowSequences) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613555)
	ctx.WriteString("SHOW SEQUENCES")
	if node.Database != "" {
		__antithesis_instrumentation__.Notify(613556)
		ctx.WriteString(" FROM ")
		ctx.FormatNode(&node.Database)
	} else {
		__antithesis_instrumentation__.Notify(613557)
	}
}

type ShowTables struct {
	ObjectNamePrefix
	WithComment bool
}

func (node *ShowTables) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613558)
	ctx.WriteString("SHOW TABLES")
	if node.ExplicitSchema {
		__antithesis_instrumentation__.Notify(613560)
		ctx.WriteString(" FROM ")
		ctx.FormatNode(&node.ObjectNamePrefix)
	} else {
		__antithesis_instrumentation__.Notify(613561)
	}
	__antithesis_instrumentation__.Notify(613559)

	if node.WithComment {
		__antithesis_instrumentation__.Notify(613562)
		ctx.WriteString(" WITH COMMENT")
	} else {
		__antithesis_instrumentation__.Notify(613563)
	}
}

type ShowTransactions struct {
	All     bool
	Cluster bool
}

func (node *ShowTransactions) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613564)
	ctx.WriteString("SHOW ")
	if node.All {
		__antithesis_instrumentation__.Notify(613566)
		ctx.WriteString("ALL ")
	} else {
		__antithesis_instrumentation__.Notify(613567)
	}
	__antithesis_instrumentation__.Notify(613565)
	if node.Cluster {
		__antithesis_instrumentation__.Notify(613568)
		ctx.WriteString("CLUSTER TRANSACTIONS")
	} else {
		__antithesis_instrumentation__.Notify(613569)
		ctx.WriteString("LOCAL TRANSACTIONS")
	}
}

type ShowConstraints struct {
	Table       *UnresolvedObjectName
	WithComment bool
}

func (node *ShowConstraints) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613570)
	ctx.WriteString("SHOW CONSTRAINTS FROM ")
	ctx.FormatNode(node.Table)

	if node.WithComment {
		__antithesis_instrumentation__.Notify(613571)
		ctx.WriteString(" WITH COMMENT")
	} else {
		__antithesis_instrumentation__.Notify(613572)
	}
}

type ShowGrants struct {
	Targets  *TargetList
	Grantees RoleSpecList
}

func (node *ShowGrants) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613573)
	ctx.WriteString("SHOW GRANTS")
	if node.Targets != nil {
		__antithesis_instrumentation__.Notify(613575)
		ctx.WriteString(" ON ")
		ctx.FormatNode(node.Targets)
	} else {
		__antithesis_instrumentation__.Notify(613576)
	}
	__antithesis_instrumentation__.Notify(613574)
	if node.Grantees != nil {
		__antithesis_instrumentation__.Notify(613577)
		ctx.WriteString(" FOR ")
		ctx.FormatNode(&node.Grantees)
	} else {
		__antithesis_instrumentation__.Notify(613578)
	}
}

type ShowRoleGrants struct {
	Roles    RoleSpecList
	Grantees RoleSpecList
}

func (node *ShowRoleGrants) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613579)
	ctx.WriteString("SHOW GRANTS ON ROLE")
	if node.Roles != nil {
		__antithesis_instrumentation__.Notify(613581)
		ctx.WriteString(" ")
		ctx.FormatNode(&node.Roles)
	} else {
		__antithesis_instrumentation__.Notify(613582)
	}
	__antithesis_instrumentation__.Notify(613580)
	if node.Grantees != nil {
		__antithesis_instrumentation__.Notify(613583)
		ctx.WriteString(" FOR ")
		ctx.FormatNode(&node.Grantees)
	} else {
		__antithesis_instrumentation__.Notify(613584)
	}
}

type ShowCreateMode int

const (
	ShowCreateModeTable ShowCreateMode = iota

	ShowCreateModeView

	ShowCreateModeSequence

	ShowCreateModeDatabase
)

type ShowCreate struct {
	Mode ShowCreateMode
	Name *UnresolvedObjectName
}

func (node *ShowCreate) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613585)
	ctx.WriteString("SHOW CREATE ")

	switch node.Mode {
	case ShowCreateModeDatabase:
		__antithesis_instrumentation__.Notify(613587)
		ctx.WriteString("DATABASE ")
	default:
		__antithesis_instrumentation__.Notify(613588)
	}
	__antithesis_instrumentation__.Notify(613586)
	ctx.FormatNode(node.Name)
}

type ShowCreateAllSchemas struct{}

func (node *ShowCreateAllSchemas) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613589)
	ctx.WriteString("SHOW CREATE ALL SCHEMAS")
}

type ShowCreateAllTables struct{}

func (node *ShowCreateAllTables) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613590)
	ctx.WriteString("SHOW CREATE ALL TABLES")
}

type ShowCreateAllTypes struct{}

func (node *ShowCreateAllTypes) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613591)
	ctx.WriteString("SHOW CREATE ALL TYPES")
}

type ShowCreateSchedules struct {
	ScheduleID Expr
}

func (node *ShowCreateSchedules) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613592)
	if node.ScheduleID != nil {
		__antithesis_instrumentation__.Notify(613594)
		ctx.WriteString("SHOW CREATE SCHEDULE ")
		ctx.FormatNode(node.ScheduleID)
		return
	} else {
		__antithesis_instrumentation__.Notify(613595)
	}
	__antithesis_instrumentation__.Notify(613593)
	ctx.Printf("SHOW CREATE ALL SCHEDULES")
}

type ShowSyntax struct {
	Statement string
}

func (node *ShowSyntax) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613596)
	ctx.WriteString("SHOW SYNTAX ")
	if ctx.flags.HasFlags(FmtAnonymize) || func() bool {
		__antithesis_instrumentation__.Notify(613597)
		return ctx.flags.HasFlags(FmtHideConstants) == true
	}() == true {
		__antithesis_instrumentation__.Notify(613598)
		ctx.WriteString("'_'")
	} else {
		__antithesis_instrumentation__.Notify(613599)
		ctx.WriteString(lexbase.EscapeSQLString(node.Statement))
	}
}

type ShowTransactionStatus struct {
}

func (node *ShowTransactionStatus) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613600)
	ctx.WriteString("SHOW TRANSACTION STATUS")
}

type ShowLastQueryStatistics struct {
	Columns NameList
}

var ShowLastQueryStatisticsDefaultColumns = NameList([]Name{
	"parse_latency",
	"plan_latency",
	"exec_latency",
	"service_latency",
	"post_commit_jobs_latency",
})

func (node *ShowLastQueryStatistics) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613601)
	ctx.WriteString("SHOW LAST QUERY STATISTICS RETURNING ")

	ctx.WithFlags(ctx.flags & ^FmtAnonymize & ^FmtMarkRedactionNode, func() {
		__antithesis_instrumentation__.Notify(613602)
		ctx.FormatNode(&node.Columns)
	})
}

type ShowFullTableScans struct {
}

func (node *ShowFullTableScans) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613603)
	ctx.WriteString("SHOW FULL TABLE SCANS")
}

type ShowSavepointStatus struct {
}

func (node *ShowSavepointStatus) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613604)
	ctx.WriteString("SHOW SAVEPOINT STATUS")
}

type ShowUsers struct {
}

func (node *ShowUsers) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613605)
	ctx.WriteString("SHOW USERS")
}

type ShowRoles struct {
}

func (node *ShowRoles) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613606)
	ctx.WriteString("SHOW ROLES")
}

type ShowRanges struct {
	TableOrIndex TableIndexName
	DatabaseName Name
}

func (node *ShowRanges) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613607)
	ctx.WriteString("SHOW RANGES FROM ")
	if node.DatabaseName != "" {
		__antithesis_instrumentation__.Notify(613608)
		ctx.WriteString("DATABASE ")
		ctx.FormatNode(&node.DatabaseName)
	} else {
		__antithesis_instrumentation__.Notify(613609)
		if node.TableOrIndex.Index != "" {
			__antithesis_instrumentation__.Notify(613610)
			ctx.WriteString("INDEX ")
			ctx.FormatNode(&node.TableOrIndex)
		} else {
			__antithesis_instrumentation__.Notify(613611)
			ctx.WriteString("TABLE ")
			ctx.FormatNode(&node.TableOrIndex)
		}
	}
}

type ShowRangeForRow struct {
	TableOrIndex TableIndexName
	Row          Exprs
}

func (node *ShowRangeForRow) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613612)
	ctx.WriteString("SHOW RANGE FROM ")
	if node.TableOrIndex.Index != "" {
		__antithesis_instrumentation__.Notify(613614)
		ctx.WriteString("INDEX ")
	} else {
		__antithesis_instrumentation__.Notify(613615)
		ctx.WriteString("TABLE ")
	}
	__antithesis_instrumentation__.Notify(613613)
	ctx.FormatNode(&node.TableOrIndex)
	ctx.WriteString(" FOR ROW (")
	ctx.FormatNode(&node.Row)
	ctx.WriteString(")")
}

type ShowFingerprints struct {
	Table *UnresolvedObjectName
}

func (node *ShowFingerprints) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613616)
	ctx.WriteString("SHOW EXPERIMENTAL_FINGERPRINTS FROM TABLE ")
	ctx.FormatNode(node.Table)
}

type ShowTableStats struct {
	Table     *UnresolvedObjectName
	UsingJSON bool
}

func (node *ShowTableStats) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613617)
	ctx.WriteString("SHOW STATISTICS ")
	if node.UsingJSON {
		__antithesis_instrumentation__.Notify(613619)
		ctx.WriteString("USING JSON ")
	} else {
		__antithesis_instrumentation__.Notify(613620)
	}
	__antithesis_instrumentation__.Notify(613618)
	ctx.WriteString("FOR TABLE ")
	ctx.FormatNode(node.Table)
}

type ShowHistogram struct {
	HistogramID int64
}

func (node *ShowHistogram) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613621)
	ctx.Printf("SHOW HISTOGRAM %d", node.HistogramID)
}

type ShowPartitions struct {
	IsDB     bool
	Database Name

	IsIndex bool
	Index   TableIndexName

	IsTable bool
	Table   *UnresolvedObjectName
}

func (node *ShowPartitions) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613622)
	if node.IsDB {
		__antithesis_instrumentation__.Notify(613623)
		ctx.Printf("SHOW PARTITIONS FROM DATABASE ")
		ctx.FormatNode(&node.Database)
	} else {
		__antithesis_instrumentation__.Notify(613624)
		if node.IsIndex {
			__antithesis_instrumentation__.Notify(613625)
			ctx.Printf("SHOW PARTITIONS FROM INDEX ")
			ctx.FormatNode(&node.Index)
		} else {
			__antithesis_instrumentation__.Notify(613626)
			ctx.Printf("SHOW PARTITIONS FROM TABLE ")
			ctx.FormatNode(node.Table)
		}
	}
}

type ScheduledJobExecutorType int

const (
	InvalidExecutor ScheduledJobExecutorType = iota

	ScheduledBackupExecutor

	ScheduledSQLStatsCompactionExecutor

	ScheduledRowLevelTTLExecutor
)

var scheduleExecutorInternalNames = map[ScheduledJobExecutorType]string{
	InvalidExecutor:                     "unknown-executor",
	ScheduledBackupExecutor:             "scheduled-backup-executor",
	ScheduledSQLStatsCompactionExecutor: "scheduled-sql-stats-compaction-executor",
	ScheduledRowLevelTTLExecutor:        "scheduled-row-level-ttl-executor",
}

func (t ScheduledJobExecutorType) InternalName() string {
	__antithesis_instrumentation__.Notify(613627)
	return scheduleExecutorInternalNames[t]
}

func (t ScheduledJobExecutorType) UserName() string {
	__antithesis_instrumentation__.Notify(613628)
	switch t {
	case ScheduledBackupExecutor:
		__antithesis_instrumentation__.Notify(613630)
		return "BACKUP"
	case ScheduledSQLStatsCompactionExecutor:
		__antithesis_instrumentation__.Notify(613631)
		return "SQL STATISTICS"
	case ScheduledRowLevelTTLExecutor:
		__antithesis_instrumentation__.Notify(613632)
		return "ROW LEVEL TTL"
	default:
		__antithesis_instrumentation__.Notify(613633)
	}
	__antithesis_instrumentation__.Notify(613629)
	return "unsupported-executor"
}

type ScheduleState int

const (
	SpecifiedSchedules ScheduleState = iota

	ActiveSchedules

	PausedSchedules
)

func (s ScheduleState) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613634)
	switch s {
	case ActiveSchedules:
		__antithesis_instrumentation__.Notify(613635)
		ctx.WriteString("RUNNING")
	case PausedSchedules:
		__antithesis_instrumentation__.Notify(613636)
		ctx.WriteString("PAUSED")
	default:
		__antithesis_instrumentation__.Notify(613637)

	}
}

type ShowSchedules struct {
	WhichSchedules ScheduleState
	ExecutorType   ScheduledJobExecutorType
	ScheduleID     Expr
}

var _ Statement = &ShowSchedules{}

func (n *ShowSchedules) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613638)
	if n.ScheduleID != nil {
		__antithesis_instrumentation__.Notify(613641)
		ctx.WriteString("SHOW SCHEDULE ")
		ctx.FormatNode(n.ScheduleID)
		return
	} else {
		__antithesis_instrumentation__.Notify(613642)
	}
	__antithesis_instrumentation__.Notify(613639)
	ctx.Printf("SHOW")

	if n.WhichSchedules != SpecifiedSchedules {
		__antithesis_instrumentation__.Notify(613643)
		ctx.WriteString(" ")
		ctx.FormatNode(&n.WhichSchedules)
	} else {
		__antithesis_instrumentation__.Notify(613644)
	}
	__antithesis_instrumentation__.Notify(613640)

	ctx.Printf(" SCHEDULES")

	if n.ExecutorType != InvalidExecutor {
		__antithesis_instrumentation__.Notify(613645)

		ctx.Printf(" FOR %s", n.ExecutorType.UserName())
	} else {
		__antithesis_instrumentation__.Notify(613646)
	}
}

type ShowDefaultPrivileges struct {
	Roles       RoleSpecList
	ForAllRoles bool

	Schema Name
}

var _ Statement = &ShowDefaultPrivileges{}

func (n *ShowDefaultPrivileges) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613647)
	ctx.WriteString("SHOW DEFAULT PRIVILEGES ")
	if len(n.Roles) > 0 {
		__antithesis_instrumentation__.Notify(613649)
		ctx.WriteString("FOR ROLE ")
		for i, role := range n.Roles {
			__antithesis_instrumentation__.Notify(613651)
			if i > 0 {
				__antithesis_instrumentation__.Notify(613653)
				ctx.WriteString(", ")
			} else {
				__antithesis_instrumentation__.Notify(613654)
			}
			__antithesis_instrumentation__.Notify(613652)
			ctx.FormatNode(&role)
		}
		__antithesis_instrumentation__.Notify(613650)
		ctx.WriteString(" ")
	} else {
		__antithesis_instrumentation__.Notify(613655)
		if n.ForAllRoles {
			__antithesis_instrumentation__.Notify(613656)
			ctx.WriteString("FOR ALL ROLES ")
		} else {
			__antithesis_instrumentation__.Notify(613657)
		}
	}
	__antithesis_instrumentation__.Notify(613648)
	if n.Schema != Name("") {
		__antithesis_instrumentation__.Notify(613658)
		ctx.WriteString("IN SCHEMA ")
		ctx.FormatNode(&n.Schema)
	} else {
		__antithesis_instrumentation__.Notify(613659)
	}
}

type ShowTransferState struct {
	TransferKey *StrVal
}

func (node *ShowTransferState) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613660)
	ctx.WriteString("SHOW TRANSFER STATE")
	if node.TransferKey != nil {
		__antithesis_instrumentation__.Notify(613661)
		ctx.WriteString(" WITH ")
		ctx.FormatNode(node.TransferKey)
	} else {
		__antithesis_instrumentation__.Notify(613662)
	}
}

type ShowCompletions struct {
	Statement *StrVal
	Offset    *NumVal
}

func (s ShowCompletions) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613663)
	ctx.WriteString("SHOW COMPLETIONS AT OFFSET ")
	s.Offset.Format(ctx)
	ctx.WriteString(" FOR ")
	ctx.FormatNode(s.Statement)
}

var _ Statement = &ShowCompletions{}
