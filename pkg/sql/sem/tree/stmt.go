package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"strings"
)

type StatementReturnType int

type StatementType int

const (
	Ack StatementReturnType = iota

	DDL

	RowsAffected

	Rows

	CopyIn

	Unknown
)

const (
	TypeDDL StatementType = iota

	TypeDML

	TypeDCL

	TypeTCL
)

type Statement interface {
	fmt.Stringer
	NodeFormatter

	StatementReturnType() StatementReturnType

	StatementType() StatementType

	StatementTag() string
}

type canModifySchema interface {
	modifiesSchema() bool
}

func CanModifySchema(stmt Statement) bool {
	__antithesis_instrumentation__.Notify(613698)
	if stmt.StatementReturnType() == DDL {
		__antithesis_instrumentation__.Notify(613700)
		return true
	} else {
		__antithesis_instrumentation__.Notify(613701)
	}
	__antithesis_instrumentation__.Notify(613699)
	scm, ok := stmt.(canModifySchema)
	return ok && func() bool {
		__antithesis_instrumentation__.Notify(613702)
		return scm.modifiesSchema() == true
	}() == true
}

func CanWriteData(stmt Statement) bool {
	__antithesis_instrumentation__.Notify(613703)
	switch stmt.(type) {

	case *Insert, *Delete, *Update, *Truncate:
		__antithesis_instrumentation__.Notify(613705)
		return true

	case *CopyFrom, *Import, *Restore:
		__antithesis_instrumentation__.Notify(613706)
		return true

	case *Split, *Unsplit, *Relocate, *RelocateRange, *Scatter:
		__antithesis_instrumentation__.Notify(613707)
		return true
	}
	__antithesis_instrumentation__.Notify(613704)
	return false
}

type HiddenFromShowQueries interface {
	hiddenFromShowQueries()
}

type ObserverStatement interface {
	observerStatement()
}

type CCLOnlyStatement interface {
	cclOnlyStatement()
}

var _ CCLOnlyStatement = &AlterBackup{}
var _ CCLOnlyStatement = &Backup{}
var _ CCLOnlyStatement = &ShowBackup{}
var _ CCLOnlyStatement = &Restore{}
var _ CCLOnlyStatement = &CreateChangefeed{}
var _ CCLOnlyStatement = &AlterChangefeed{}
var _ CCLOnlyStatement = &Import{}
var _ CCLOnlyStatement = &Export{}
var _ CCLOnlyStatement = &ScheduledBackup{}
var _ CCLOnlyStatement = &StreamIngestion{}
var _ CCLOnlyStatement = &ReplicationStream{}

func (*AlterChangefeed) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613708)
	return Rows
}

func (*AlterChangefeed) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613709)
	return TypeDML
}

func (*AlterChangefeed) StatementTag() string {
	__antithesis_instrumentation__.Notify(613710)
	return `ALTER CHANGEFEED`
}

func (*AlterChangefeed) cclOnlyStatement() { __antithesis_instrumentation__.Notify(613711) }

func (*AlterBackup) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613712)
	return Rows
}

func (*AlterBackup) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613713)
	return TypeDML
}

func (*AlterBackup) StatementTag() string {
	__antithesis_instrumentation__.Notify(613714)
	return "ALTER BACKUP"
}

func (*AlterBackup) cclOnlyStatement() { __antithesis_instrumentation__.Notify(613715) }

func (*AlterDatabaseOwner) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613716)
	return DDL
}

func (*AlterDatabaseOwner) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613717)
	return TypeDCL
}

func (*AlterDatabaseOwner) StatementTag() string {
	__antithesis_instrumentation__.Notify(613718)
	return "ALTER DATABASE OWNER"
}

func (*AlterDatabaseOwner) hiddenFromShowQueries() { __antithesis_instrumentation__.Notify(613719) }

func (*AlterDatabaseAddRegion) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613720)
	return DDL
}

func (*AlterDatabaseAddRegion) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613721)
	return TypeDDL
}

func (*AlterDatabaseAddRegion) StatementTag() string {
	__antithesis_instrumentation__.Notify(613722)
	return "ALTER DATABASE ADD REGION"
}

func (*AlterDatabaseAddRegion) hiddenFromShowQueries() { __antithesis_instrumentation__.Notify(613723) }

func (*AlterDatabaseDropRegion) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613724)
	return DDL
}

func (*AlterDatabaseDropRegion) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613725)
	return TypeDDL
}

func (*AlterDatabaseDropRegion) StatementTag() string {
	__antithesis_instrumentation__.Notify(613726)
	return "ALTER DATABASE DROP REGION"
}

func (*AlterDatabaseDropRegion) hiddenFromShowQueries() {
	__antithesis_instrumentation__.Notify(613727)
}

func (*AlterDatabasePrimaryRegion) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613728)
	return DDL
}

func (*AlterDatabasePrimaryRegion) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613729)
	return TypeDDL
}

func (*AlterDatabasePrimaryRegion) StatementTag() string {
	__antithesis_instrumentation__.Notify(613730)
	return "ALTER DATABASE PRIMARY REGION"
}

func (*AlterDatabasePrimaryRegion) hiddenFromShowQueries() {
	__antithesis_instrumentation__.Notify(613731)
}

func (*AlterDatabaseSurvivalGoal) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613732)
	return DDL
}

func (*AlterDatabaseSurvivalGoal) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613733)
	return TypeDDL
}

func (*AlterDatabaseSurvivalGoal) StatementTag() string {
	__antithesis_instrumentation__.Notify(613734)
	return "ALTER DATABASE SURVIVE"
}

func (*AlterDatabaseSurvivalGoal) hiddenFromShowQueries() {
	__antithesis_instrumentation__.Notify(613735)
}

func (*AlterDatabasePlacement) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613736)
	return DDL
}

func (*AlterDatabasePlacement) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613737)
	return TypeDDL
}

func (*AlterDatabasePlacement) StatementTag() string {
	__antithesis_instrumentation__.Notify(613738)
	return "ALTER DATABASE PLACEMENT"
}

func (*AlterDatabasePlacement) hiddenFromShowQueries() { __antithesis_instrumentation__.Notify(613739) }

func (*AlterDatabaseAddSuperRegion) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613740)
	return DDL
}

func (*AlterDatabaseAddSuperRegion) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613741)
	return TypeDDL
}

func (*AlterDatabaseAddSuperRegion) StatementTag() string {
	__antithesis_instrumentation__.Notify(613742)
	return "ALTER DATABASE ADD SUPER REGION"
}

func (*AlterDatabaseAddSuperRegion) hiddenFromShowQueries() {
	__antithesis_instrumentation__.Notify(613743)
}

func (*AlterDatabaseDropSuperRegion) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613744)
	return DDL
}

func (*AlterDatabaseDropSuperRegion) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613745)
	return TypeDDL
}

func (*AlterDatabaseDropSuperRegion) StatementTag() string {
	__antithesis_instrumentation__.Notify(613746)
	return "ALTER DATABASE DROP SUPER REGION"
}

func (*AlterDatabaseDropSuperRegion) hiddenFromShowQueries() {
	__antithesis_instrumentation__.Notify(613747)
}

func (*AlterDatabaseAlterSuperRegion) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613748)
	return DDL
}

func (*AlterDatabaseAlterSuperRegion) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613749)
	return TypeDDL
}

func (*AlterDatabaseAlterSuperRegion) StatementTag() string {
	__antithesis_instrumentation__.Notify(613750)
	return "ALTER DATABASE ALTER SUPER REGION"
}

func (*AlterDatabaseAlterSuperRegion) hiddenFromShowQueries() {
	__antithesis_instrumentation__.Notify(613751)
}

func (*AlterDefaultPrivileges) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613752)
	return DDL
}

func (*AlterDefaultPrivileges) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613753)
	return TypeDCL
}

func (*AlterDefaultPrivileges) StatementTag() string {
	__antithesis_instrumentation__.Notify(613754)
	return "ALTER DEFAULT PRIVILEGES"
}

func (*AlterIndex) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613755)
	return DDL
}

func (*AlterIndex) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613756)
	return TypeDDL
}

func (*AlterIndex) StatementTag() string {
	__antithesis_instrumentation__.Notify(613757)
	return "ALTER INDEX"
}

func (*AlterIndex) hiddenFromShowQueries() { __antithesis_instrumentation__.Notify(613758) }

func (*AlterTable) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613759)
	return DDL
}

func (*AlterTable) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613760)
	return TypeDDL
}

func (*AlterTable) StatementTag() string {
	__antithesis_instrumentation__.Notify(613761)
	return "ALTER TABLE"
}

func (*AlterTable) hiddenFromShowQueries() { __antithesis_instrumentation__.Notify(613762) }

func (*AlterTableLocality) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613763)
	return DDL
}

func (*AlterTableLocality) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613764)
	return TypeDDL
}

func (*AlterTableLocality) StatementTag() string {
	__antithesis_instrumentation__.Notify(613765)
	return "ALTER TABLE SET LOCALITY"
}

func (*AlterTableLocality) hiddenFromShowQueries() { __antithesis_instrumentation__.Notify(613766) }

func (*AlterTableOwner) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613767)
	return DDL
}

func (*AlterTableOwner) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613768)
	return TypeDCL
}

func (*AlterTableOwner) StatementTag() string {
	__antithesis_instrumentation__.Notify(613769)
	return "ALTER TABLE OWNER"
}

func (*AlterTableOwner) hiddenFromShowQueries() { __antithesis_instrumentation__.Notify(613770) }

func (*AlterTableSetSchema) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613771)
	return DDL
}

func (*AlterTableSetSchema) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613772)
	return TypeDDL
}

func (*AlterTableSetSchema) StatementTag() string {
	__antithesis_instrumentation__.Notify(613773)
	return "ALTER TABLE SET SCHEMA"
}

func (*AlterTableSetSchema) hiddenFromShowQueries() { __antithesis_instrumentation__.Notify(613774) }

func (*AlterSchema) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613775)
	return DDL
}

func (*AlterSchema) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613776)
	return TypeDDL
}

func (*AlterSchema) StatementTag() string {
	__antithesis_instrumentation__.Notify(613777)
	return "ALTER SCHEMA"
}

func (*AlterSchema) hiddenFromShowQueries() { __antithesis_instrumentation__.Notify(613778) }

func (*AlterTenantSetClusterSetting) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613779)
	return Ack
}

func (*AlterTenantSetClusterSetting) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613780)
	return TypeDCL
}

func (*AlterTenantSetClusterSetting) StatementTag() string {
	__antithesis_instrumentation__.Notify(613781)
	return "ALTER TENANT SET CLUSTER SETTING"
}

func (*AlterType) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613782)
	return DDL
}

func (*AlterType) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613783)
	return TypeDDL
}

func (*AlterType) StatementTag() string {
	__antithesis_instrumentation__.Notify(613784)
	return "ALTER TYPE"
}

func (*AlterType) hiddenFromShowQueries() { __antithesis_instrumentation__.Notify(613785) }

func (*AlterSequence) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613786)
	return DDL
}

func (*AlterSequence) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613787)
	return TypeDDL
}

func (*AlterSequence) StatementTag() string {
	__antithesis_instrumentation__.Notify(613788)
	return "ALTER SEQUENCE"
}

func (*AlterRole) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613789)
	return Ack
}

func (*AlterRole) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613790)
	return TypeDDL
}

func (*AlterRole) StatementTag() string {
	__antithesis_instrumentation__.Notify(613791)
	return "ALTER ROLE"
}

func (*AlterRole) hiddenFromShowQueries() { __antithesis_instrumentation__.Notify(613792) }

func (*AlterRoleSet) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613793)
	return Ack
}

func (*AlterRoleSet) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613794)
	return TypeDDL
}

func (*AlterRoleSet) StatementTag() string {
	__antithesis_instrumentation__.Notify(613795)
	return "ALTER ROLE"
}

func (*AlterRoleSet) hiddenFromShowQueries() { __antithesis_instrumentation__.Notify(613796) }

func (*Analyze) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613797)
	return DDL
}

func (*Analyze) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613798)
	return TypeDDL
}

func (*Analyze) StatementTag() string {
	__antithesis_instrumentation__.Notify(613799)
	return "ANALYZE"
}

func (*Backup) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613800)
	return Rows
}

func (*Backup) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613801)
	return TypeDML
}

func (*Backup) StatementTag() string { __antithesis_instrumentation__.Notify(613802); return "BACKUP" }

func (*Backup) cclOnlyStatement() { __antithesis_instrumentation__.Notify(613803) }

func (*Backup) hiddenFromShowQueries() { __antithesis_instrumentation__.Notify(613804) }

func (*ScheduledBackup) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613805)
	return Rows
}

func (*ScheduledBackup) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613806)
	return TypeDML
}

func (*ScheduledBackup) StatementTag() string {
	__antithesis_instrumentation__.Notify(613807)
	return "SCHEDULED BACKUP"
}

func (*ScheduledBackup) cclOnlyStatement() { __antithesis_instrumentation__.Notify(613808) }

func (*ScheduledBackup) hiddenFromShowQueries() { __antithesis_instrumentation__.Notify(613809) }

func (*BeginTransaction) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613810)
	return Ack
}

func (*BeginTransaction) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613811)
	return TypeTCL
}

func (*BeginTransaction) StatementTag() string {
	__antithesis_instrumentation__.Notify(613812)
	return "BEGIN"
}

func (*ControlJobs) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613813)
	return RowsAffected
}

func (*ControlJobs) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613814)
	return TypeTCL
}

func (n *ControlJobs) StatementTag() string {
	__antithesis_instrumentation__.Notify(613815)
	return fmt.Sprintf("%s JOBS", JobCommandToStatement[n.Command])
}

func (*ControlSchedules) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613816)
	return RowsAffected
}

func (*ControlSchedules) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613817)
	return TypeTCL
}

func (n *ControlSchedules) StatementTag() string {
	__antithesis_instrumentation__.Notify(613818)
	return fmt.Sprintf("%s SCHEDULES", n.Command)
}

func (*ControlJobsForSchedules) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613819)
	return RowsAffected
}

func (*ControlJobsForSchedules) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613820)
	return TypeTCL
}

func (n *ControlJobsForSchedules) StatementTag() string {
	__antithesis_instrumentation__.Notify(613821)
	return fmt.Sprintf("%s JOBS FOR SCHEDULES", JobCommandToStatement[n.Command])
}

func (*ControlJobsOfType) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613822)
	return RowsAffected
}

func (*ControlJobsOfType) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613823)
	return TypeTCL
}

func (n *ControlJobsOfType) StatementTag() string {
	__antithesis_instrumentation__.Notify(613824)
	return fmt.Sprintf("%s ALL %s JOBS", JobCommandToStatement[n.Command], strings.ToUpper(n.Type))
}

func (*CancelQueries) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613825)
	return RowsAffected
}

func (*CancelQueries) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613826)
	return TypeDML
}

func (*CancelQueries) StatementTag() string {
	__antithesis_instrumentation__.Notify(613827)
	return "CANCEL QUERIES"
}

func (*CancelSessions) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613828)
	return RowsAffected
}

func (*CancelSessions) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613829)
	return TypeDML
}

func (*CancelSessions) StatementTag() string {
	__antithesis_instrumentation__.Notify(613830)
	return "CANCEL SESSIONS"
}

func (*CannedOptPlan) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613831)
	return Rows
}

func (*CannedOptPlan) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613832)
	return TypeDML
}

func (*CannedOptPlan) StatementTag() string {
	__antithesis_instrumentation__.Notify(613833)
	return "PREPARE AS OPT PLAN"
}

func (*CloseCursor) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613834)
	return Ack
}

func (*CloseCursor) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613835)
	return TypeDCL
}

func (*CloseCursor) StatementTag() string {
	__antithesis_instrumentation__.Notify(613836)
	return "CLOSE"
}

func (*CommentOnColumn) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613837)
	return DDL
}

func (*CommentOnColumn) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613838)
	return TypeDDL
}

func (*CommentOnColumn) StatementTag() string {
	__antithesis_instrumentation__.Notify(613839)
	return "COMMENT ON COLUMN"
}

func (*CommentOnConstraint) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613840)
	return DDL
}

func (*CommentOnConstraint) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613841)
	return TypeDDL
}

func (*CommentOnConstraint) StatementTag() string {
	__antithesis_instrumentation__.Notify(613842)
	return "COMMENT ON CONSTRAINT"
}

func (*CommentOnDatabase) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613843)
	return DDL
}

func (*CommentOnDatabase) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613844)
	return TypeDDL
}

func (*CommentOnDatabase) StatementTag() string {
	__antithesis_instrumentation__.Notify(613845)
	return "COMMENT ON DATABASE"
}

func (*CommentOnSchema) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613846)
	return DDL
}

func (*CommentOnSchema) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613847)
	return TypeDDL
}

func (*CommentOnSchema) StatementTag() string {
	__antithesis_instrumentation__.Notify(613848)
	return "COMMENT ON SCHEMA"
}

func (*CommentOnIndex) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613849)
	return DDL
}

func (*CommentOnIndex) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613850)
	return TypeDDL
}

func (*CommentOnIndex) StatementTag() string {
	__antithesis_instrumentation__.Notify(613851)
	return "COMMENT ON INDEX"
}

func (*CommentOnTable) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613852)
	return DDL
}

func (*CommentOnTable) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613853)
	return TypeDDL
}

func (*CommentOnTable) StatementTag() string {
	__antithesis_instrumentation__.Notify(613854)
	return "COMMENT ON TABLE"
}

func (*CommitTransaction) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613855)
	return Ack
}

func (*CommitTransaction) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613856)
	return TypeTCL
}

func (*CommitTransaction) StatementTag() string {
	__antithesis_instrumentation__.Notify(613857)
	return "COMMIT"
}

func (*CopyFrom) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613858)
	return CopyIn
}

func (*CopyFrom) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613859)
	return TypeDML
}

func (*CopyFrom) StatementTag() string { __antithesis_instrumentation__.Notify(613860); return "COPY" }

func (*CreateChangefeed) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613861)
	return Rows
}

func (*CreateChangefeed) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613862)
	return TypeDDL
}

func (n *CreateChangefeed) StatementTag() string {
	__antithesis_instrumentation__.Notify(613863)
	if n.SinkURI == nil {
		__antithesis_instrumentation__.Notify(613865)
		return "EXPERIMENTAL CHANGEFEED"
	} else {
		__antithesis_instrumentation__.Notify(613866)
	}
	__antithesis_instrumentation__.Notify(613864)
	return "CREATE CHANGEFEED"
}

func (*CreateChangefeed) cclOnlyStatement() { __antithesis_instrumentation__.Notify(613867) }

func (*CreateDatabase) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613868)
	return DDL
}

func (*CreateDatabase) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613869)
	return TypeDDL
}

func (*CreateDatabase) StatementTag() string {
	__antithesis_instrumentation__.Notify(613870)
	return "CREATE DATABASE"
}

func (*CreateExtension) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613871)
	return Ack
}

func (*CreateExtension) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613872)
	return TypeDDL
}

func (*CreateExtension) StatementTag() string {
	__antithesis_instrumentation__.Notify(613873)
	return "CREATE EXTENSION"
}

func (*CreateIndex) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613874)
	return DDL
}

func (*CreateIndex) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613875)
	return TypeDDL
}

func (*CreateIndex) StatementTag() string {
	__antithesis_instrumentation__.Notify(613876)
	return "CREATE INDEX"
}

func (n *CreateSchema) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613877)
	return DDL
}

func (*CreateSchema) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613878)
	return TypeDDL
}

func (n *CreateSchema) StatementTag() string {
	__antithesis_instrumentation__.Notify(613879)
	return "CREATE SCHEMA"
}

func (*CreateSchema) modifiesSchema() bool {
	__antithesis_instrumentation__.Notify(613880)
	return true
}

func (n *CreateTable) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613881)
	return DDL
}

func (*CreateTable) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613882)
	return TypeDDL
}

func (n *CreateTable) StatementTag() string {
	__antithesis_instrumentation__.Notify(613883)
	if n.As() {
		__antithesis_instrumentation__.Notify(613885)
		return "CREATE TABLE AS"
	} else {
		__antithesis_instrumentation__.Notify(613886)
	}
	__antithesis_instrumentation__.Notify(613884)
	return "CREATE TABLE"
}

func (*CreateTable) modifiesSchema() bool { __antithesis_instrumentation__.Notify(613887); return true }

func (*CreateType) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613888)
	return DDL
}

func (*CreateType) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613889)
	return TypeDDL
}

func (*CreateType) StatementTag() string {
	__antithesis_instrumentation__.Notify(613890)
	return "CREATE TYPE"
}

func (*CreateType) modifiesSchema() bool { __antithesis_instrumentation__.Notify(613891); return true }

func (*CreateRole) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613892)
	return Ack
}

func (*CreateRole) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613893)
	return TypeDDL
}

func (*CreateRole) StatementTag() string {
	__antithesis_instrumentation__.Notify(613894)
	return "CREATE ROLE"
}

func (*CreateRole) hiddenFromShowQueries() { __antithesis_instrumentation__.Notify(613895) }

func (*CreateView) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613896)
	return DDL
}

func (*CreateView) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613897)
	return TypeDDL
}

func (*CreateView) StatementTag() string {
	__antithesis_instrumentation__.Notify(613898)
	return "CREATE VIEW"
}

func (*CreateSequence) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613899)
	return DDL
}

func (*CreateSequence) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613900)
	return TypeDDL
}

func (*CreateSequence) StatementTag() string {
	__antithesis_instrumentation__.Notify(613901)
	return "CREATE SEQUENCE"
}

func (*CreateStats) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613902)
	return DDL
}

func (*CreateStats) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613903)
	return TypeDDL
}

func (*CreateStats) StatementTag() string {
	__antithesis_instrumentation__.Notify(613904)
	return "CREATE STATISTICS"
}

func (*Deallocate) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613905)
	return Ack
}

func (*Deallocate) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613906)
	return TypeTCL
}

func (n *Deallocate) StatementTag() string {
	__antithesis_instrumentation__.Notify(613907)

	if n.Name == "" {
		__antithesis_instrumentation__.Notify(613909)
		return "DEALLOCATE ALL"
	} else {
		__antithesis_instrumentation__.Notify(613910)
	}
	__antithesis_instrumentation__.Notify(613908)
	return "DEALLOCATE"
}

func (*Discard) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613911)
	return Ack
}

func (*Discard) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613912)
	return TypeTCL
}

func (*Discard) StatementTag() string {
	__antithesis_instrumentation__.Notify(613913)
	return "DISCARD"
}

func (n *DeclareCursor) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613914)
	return Ack
}

func (*DeclareCursor) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613915)
	return TypeDCL
}

func (*DeclareCursor) StatementTag() string {
	__antithesis_instrumentation__.Notify(613916)
	return "DECLARE CURSOR"
}

func (n *Delete) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613917)
	return n.Returning.statementReturnType()
}

func (*Delete) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613918)
	return TypeDML
}

func (*Delete) StatementTag() string { __antithesis_instrumentation__.Notify(613919); return "DELETE" }

func (*DropDatabase) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613920)
	return DDL
}

func (*DropDatabase) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613921)
	return TypeDDL
}

func (*DropDatabase) StatementTag() string {
	__antithesis_instrumentation__.Notify(613922)
	return "DROP DATABASE"
}

func (*DropIndex) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613923)
	return DDL
}

func (*DropIndex) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613924)
	return TypeDDL
}

func (*DropIndex) StatementTag() string {
	__antithesis_instrumentation__.Notify(613925)
	return "DROP INDEX"
}

func (*DropTable) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613926)
	return DDL
}

func (*DropTable) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613927)
	return TypeDDL
}

func (*DropTable) StatementTag() string {
	__antithesis_instrumentation__.Notify(613928)
	return "DROP TABLE"
}

func (*DropView) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613929)
	return DDL
}

func (*DropView) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613930)
	return TypeDDL
}

func (*DropView) StatementTag() string {
	__antithesis_instrumentation__.Notify(613931)
	return "DROP VIEW"
}

func (*DropSequence) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613932)
	return DDL
}

func (*DropSequence) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613933)
	return TypeDDL
}

func (*DropSequence) StatementTag() string {
	__antithesis_instrumentation__.Notify(613934)
	return "DROP SEQUENCE"
}

func (*DropRole) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613935)
	return Ack
}

func (*DropRole) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613936)
	return TypeDDL
}

func (*DropRole) StatementTag() string {
	__antithesis_instrumentation__.Notify(613937)
	return "DROP ROLE"
}

func (*DropRole) cclOnlyStatement() { __antithesis_instrumentation__.Notify(613938) }

func (*DropRole) hiddenFromShowQueries() { __antithesis_instrumentation__.Notify(613939) }

func (*DropType) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613940)
	return DDL
}

func (*DropType) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613941)
	return TypeDDL
}

func (*DropType) StatementTag() string {
	__antithesis_instrumentation__.Notify(613942)
	return "DROP TYPE"
}

func (*DropSchema) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613943)
	return DDL
}

func (*DropSchema) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613944)
	return TypeDDL
}

func (*DropSchema) StatementTag() string {
	__antithesis_instrumentation__.Notify(613945)
	return "DROP SCHEMA"
}

func (*Execute) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613946)
	return Unknown
}

func (*Execute) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613947)
	return TypeTCL
}

func (*Execute) StatementTag() string {
	__antithesis_instrumentation__.Notify(613948)
	return "EXECUTE"
}

func (*Explain) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613949)
	return Rows
}

func (*Explain) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613950)
	return TypeDML
}

func (*Explain) StatementTag() string {
	__antithesis_instrumentation__.Notify(613951)
	return "EXPLAIN"
}

func (*ExplainAnalyze) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613952)
	return Rows
}

func (*ExplainAnalyze) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613953)
	return TypeDML
}

func (*ExplainAnalyze) StatementTag() string {
	__antithesis_instrumentation__.Notify(613954)
	return "EXPLAIN ANALYZE"
}

func (*Export) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613955)
	return Rows
}

func (*Export) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613956)
	return TypeDML
}

func (*Export) cclOnlyStatement() { __antithesis_instrumentation__.Notify(613957) }

func (*Export) StatementTag() string { __antithesis_instrumentation__.Notify(613958); return "EXPORT" }

func (*Grant) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613959)
	return DDL
}

func (n *FetchCursor) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613960)
	return Rows
}

func (*FetchCursor) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613961)
	return TypeDML
}

func (*FetchCursor) StatementTag() string {
	__antithesis_instrumentation__.Notify(613962)
	return "FETCH"
}

func (n *MoveCursor) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613963)
	return RowsAffected
}

func (*MoveCursor) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613964)
	return TypeDML
}

func (*MoveCursor) StatementTag() string {
	__antithesis_instrumentation__.Notify(613965)
	return "MOVE"
}

func (*Grant) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613966)
	return TypeDCL
}

func (*Grant) StatementTag() string { __antithesis_instrumentation__.Notify(613967); return "GRANT" }

func (*GrantRole) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613968)
	return DDL
}

func (*GrantRole) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613969)
	return TypeDCL
}

func (*GrantRole) StatementTag() string {
	__antithesis_instrumentation__.Notify(613970)
	return "GRANT"
}

func (n *Insert) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613971)
	return n.Returning.statementReturnType()
}

func (*Insert) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613972)
	return TypeDML
}

func (*Insert) StatementTag() string { __antithesis_instrumentation__.Notify(613973); return "INSERT" }

func (*Import) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613974)
	return Rows
}

func (*Import) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613975)
	return TypeDML
}

func (*Import) StatementTag() string { __antithesis_instrumentation__.Notify(613976); return "IMPORT" }

func (*Import) cclOnlyStatement() { __antithesis_instrumentation__.Notify(613977) }

func (*ParenSelect) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613978)
	return Rows
}

func (*ParenSelect) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613979)
	return TypeDML
}

func (*ParenSelect) StatementTag() string {
	__antithesis_instrumentation__.Notify(613980)
	return "SELECT"
}

func (*Prepare) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613981)
	return Ack
}

func (*Prepare) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613982)
	return TypeTCL
}

func (*Prepare) StatementTag() string {
	__antithesis_instrumentation__.Notify(613983)
	return "PREPARE"
}

func (*ReassignOwnedBy) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613984)
	return DDL
}

func (*ReassignOwnedBy) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613985)
	return TypeDCL
}

func (*ReassignOwnedBy) StatementTag() string {
	__antithesis_instrumentation__.Notify(613986)
	return "REASSIGN OWNED BY"
}

func (*DropOwnedBy) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613987)
	return DDL
}

func (*DropOwnedBy) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613988)
	return TypeDCL
}

func (*DropOwnedBy) StatementTag() string {
	__antithesis_instrumentation__.Notify(613989)
	return "DROP OWNED BY"
}

func (*RefreshMaterializedView) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613990)
	return DDL
}

func (*RefreshMaterializedView) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613991)
	return TypeDDL
}

func (*RefreshMaterializedView) StatementTag() string {
	__antithesis_instrumentation__.Notify(613992)
	return "REFRESH MATERIALIZED VIEW"
}

func (*ReleaseSavepoint) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613993)
	return Ack
}

func (*ReleaseSavepoint) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613994)
	return TypeTCL
}

func (*ReleaseSavepoint) StatementTag() string {
	__antithesis_instrumentation__.Notify(613995)
	return "RELEASE"
}

func (*RenameColumn) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613996)
	return DDL
}

func (*RenameColumn) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(613997)
	return TypeDDL
}

func (*RenameColumn) StatementTag() string {
	__antithesis_instrumentation__.Notify(613998)
	return "ALTER TABLE"
}

func (*RenameDatabase) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(613999)
	return DDL
}

func (*RenameDatabase) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614000)
	return TypeDDL
}

func (*RenameDatabase) StatementTag() string {
	__antithesis_instrumentation__.Notify(614001)
	return "ALTER DATABASE"
}

func (*ReparentDatabase) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614002)
	return DDL
}

func (*ReparentDatabase) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614003)
	return TypeDDL
}

func (*ReparentDatabase) StatementTag() string {
	__antithesis_instrumentation__.Notify(614004)
	return "CONVERT TO SCHEMA"
}

func (*RenameIndex) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614005)
	return DDL
}

func (*RenameIndex) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614006)
	return TypeDDL
}

func (*RenameIndex) StatementTag() string {
	__antithesis_instrumentation__.Notify(614007)
	return "ALTER INDEX"
}

func (*RenameTable) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614008)
	return DDL
}

func (*RenameTable) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614009)
	return TypeDDL
}

func (n *RenameTable) StatementTag() string {
	__antithesis_instrumentation__.Notify(614010)
	if n.IsView {
		__antithesis_instrumentation__.Notify(614012)
		return "ALTER VIEW"
	} else {
		__antithesis_instrumentation__.Notify(614013)
		if n.IsSequence {
			__antithesis_instrumentation__.Notify(614014)
			return "ALTER SEQUENCE"
		} else {
			__antithesis_instrumentation__.Notify(614015)
		}
	}
	__antithesis_instrumentation__.Notify(614011)
	return "ALTER TABLE"
}

func (*Relocate) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614016)
	return Rows
}

func (*Relocate) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614017)
	return TypeDML
}

func (n *Relocate) StatementTag() string {
	__antithesis_instrumentation__.Notify(614018)
	name := "RELOCATE TABLE "
	if n.TableOrIndex.Index != "" {
		__antithesis_instrumentation__.Notify(614020)
		name = "RELOCATE INDEX "
	} else {
		__antithesis_instrumentation__.Notify(614021)
	}
	__antithesis_instrumentation__.Notify(614019)
	return name + n.SubjectReplicas.String()
}

func (*RelocateRange) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614022)
	return Rows
}

func (*RelocateRange) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614023)
	return TypeDML
}

func (n *RelocateRange) StatementTag() string {
	__antithesis_instrumentation__.Notify(614024)
	return "RELOCATE RANGE " + n.SubjectReplicas.String()
}

func (*ReplicationStream) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614025)
	return Rows
}

func (*ReplicationStream) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614026)
	return TypeDML
}

func (*ReplicationStream) StatementTag() string {
	__antithesis_instrumentation__.Notify(614027)
	return "CREATE REPLICATION STREAM"
}

func (*ReplicationStream) cclOnlyStatement() { __antithesis_instrumentation__.Notify(614028) }

func (*ReplicationStream) hiddenFromShowQueries() { __antithesis_instrumentation__.Notify(614029) }

func (*Restore) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614030)
	return Rows
}

func (*Restore) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614031)
	return TypeDML
}

func (*Restore) StatementTag() string {
	__antithesis_instrumentation__.Notify(614032)
	return "RESTORE"
}

func (*Restore) cclOnlyStatement() { __antithesis_instrumentation__.Notify(614033) }

func (*Restore) hiddenFromShowQueries() { __antithesis_instrumentation__.Notify(614034) }

func (*Revoke) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614035)
	return DDL
}

func (*Revoke) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614036)
	return TypeDCL
}

func (*Revoke) StatementTag() string { __antithesis_instrumentation__.Notify(614037); return "REVOKE" }

func (*RevokeRole) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614038)
	return DDL
}

func (*RevokeRole) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614039)
	return TypeDCL
}

func (*RevokeRole) StatementTag() string {
	__antithesis_instrumentation__.Notify(614040)
	return "REVOKE"
}

func (*RollbackToSavepoint) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614041)
	return Ack
}

func (*RollbackToSavepoint) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614042)
	return TypeTCL
}

func (*RollbackToSavepoint) StatementTag() string {
	__antithesis_instrumentation__.Notify(614043)
	return "ROLLBACK"
}

func (*RollbackTransaction) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614044)
	return Ack
}

func (*RollbackTransaction) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614045)
	return TypeTCL
}

func (*RollbackTransaction) StatementTag() string {
	__antithesis_instrumentation__.Notify(614046)
	return "ROLLBACK"
}

func (*Savepoint) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614047)
	return Ack
}

func (*Savepoint) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614048)
	return TypeTCL
}

func (*Savepoint) StatementTag() string {
	__antithesis_instrumentation__.Notify(614049)
	return "SAVEPOINT"
}

func (*Scatter) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614050)
	return Rows
}

func (*Scatter) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614051)
	return TypeDML
}

func (*Scatter) StatementTag() string {
	__antithesis_instrumentation__.Notify(614052)
	return "SCATTER"
}

func (*Scrub) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614053)
	return Rows
}

func (*Scrub) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614054)
	return TypeDML
}

func (n *Scrub) StatementTag() string { __antithesis_instrumentation__.Notify(614055); return "SCRUB" }

func (*Select) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614056)
	return Rows
}

func (*Select) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614057)
	return TypeDML
}

func (*Select) StatementTag() string { __antithesis_instrumentation__.Notify(614058); return "SELECT" }

func (*SelectClause) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614059)
	return Rows
}

func (*SelectClause) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614060)
	return TypeDML
}

func (*SelectClause) StatementTag() string {
	__antithesis_instrumentation__.Notify(614061)
	return "SELECT"
}

func (*SetVar) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614062)
	return Ack
}

func (*SetVar) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614063)
	return TypeDCL
}

func (n *SetVar) StatementTag() string {
	__antithesis_instrumentation__.Notify(614064)
	if n.Reset || func() bool {
		__antithesis_instrumentation__.Notify(614066)
		return n.ResetAll == true
	}() == true {
		__antithesis_instrumentation__.Notify(614067)
		return "RESET"
	} else {
		__antithesis_instrumentation__.Notify(614068)
	}
	__antithesis_instrumentation__.Notify(614065)
	return "SET"
}

func (*SetClusterSetting) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614069)
	return Ack
}

func (*SetClusterSetting) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614070)
	return TypeDCL
}

func (*SetClusterSetting) StatementTag() string {
	__antithesis_instrumentation__.Notify(614071)
	return "SET CLUSTER SETTING"
}

func (*SetTransaction) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614072)
	return Ack
}

func (*SetTransaction) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614073)
	return TypeDCL
}

func (*SetTransaction) StatementTag() string {
	__antithesis_instrumentation__.Notify(614074)
	return "SET TRANSACTION"
}

func (*SetTracing) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614075)
	return Ack
}

func (*SetTracing) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614076)
	return TypeDCL
}

func (*SetTracing) StatementTag() string {
	__antithesis_instrumentation__.Notify(614077)
	return "SET TRACING"
}

func (*SetTracing) observerStatement() { __antithesis_instrumentation__.Notify(614078) }

func (*SetZoneConfig) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614079)
	return RowsAffected
}

func (*SetZoneConfig) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614080)
	return TypeDCL
}

func (*SetZoneConfig) StatementTag() string {
	__antithesis_instrumentation__.Notify(614081)
	return "CONFIGURE ZONE"
}

func (*SetSessionAuthorizationDefault) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614082)
	return Ack
}

func (*SetSessionAuthorizationDefault) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614083)
	return TypeDCL
}

func (*SetSessionAuthorizationDefault) StatementTag() string {
	__antithesis_instrumentation__.Notify(614084)
	return "SET"
}

func (*SetSessionCharacteristics) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614085)
	return Ack
}

func (*SetSessionCharacteristics) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614086)
	return TypeDCL
}

func (*SetSessionCharacteristics) StatementTag() string {
	__antithesis_instrumentation__.Notify(614087)
	return "SET"
}

func (*ShowVar) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614088)
	return Rows
}

func (*ShowVar) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614089)
	return TypeDML
}

func (*ShowVar) StatementTag() string { __antithesis_instrumentation__.Notify(614090); return "SHOW" }

func (*ShowClusterSetting) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614091)
	return Rows
}

func (*ShowClusterSetting) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614092)
	return TypeDML
}

func (*ShowClusterSetting) StatementTag() string {
	__antithesis_instrumentation__.Notify(614093)
	return "SHOW"
}

func (*ShowClusterSettingList) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614094)
	return Rows
}

func (*ShowClusterSettingList) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614095)
	return TypeDML
}

func (*ShowClusterSettingList) StatementTag() string {
	__antithesis_instrumentation__.Notify(614096)
	return "SHOW"
}

func (*ShowTenantClusterSetting) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614097)
	return Rows
}

func (*ShowTenantClusterSetting) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614098)
	return TypeDML
}

func (*ShowTenantClusterSetting) StatementTag() string {
	__antithesis_instrumentation__.Notify(614099)
	return "SHOW"
}

func (*ShowTenantClusterSettingList) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614100)
	return Rows
}

func (*ShowTenantClusterSettingList) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614101)
	return TypeDML
}

func (*ShowTenantClusterSettingList) StatementTag() string {
	__antithesis_instrumentation__.Notify(614102)
	return "SHOW"
}

func (*ShowColumns) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614103)
	return Rows
}

func (*ShowColumns) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614104)
	return TypeDML
}

func (*ShowColumns) StatementTag() string {
	__antithesis_instrumentation__.Notify(614105)
	return "SHOW COLUMNS"
}

func (*ShowCreate) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614106)
	return Rows
}

func (*ShowCreate) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614107)
	return TypeDML
}

func (*ShowCreate) StatementTag() string {
	__antithesis_instrumentation__.Notify(614108)
	return "SHOW CREATE"
}

func (*ShowCreateAllSchemas) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614109)
	return Rows
}

func (*ShowCreateAllSchemas) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614110)
	return TypeDML
}

func (*ShowCreateAllSchemas) StatementTag() string {
	__antithesis_instrumentation__.Notify(614111)
	return "SHOW CREATE ALL SCHEMAS"
}

func (*ShowCreateAllTables) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614112)
	return Rows
}

func (*ShowCreateAllTables) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614113)
	return TypeDML
}

func (*ShowCreateAllTables) StatementTag() string {
	__antithesis_instrumentation__.Notify(614114)
	return "SHOW CREATE ALL TABLES"
}

func (*ShowCreateAllTypes) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614115)
	return Rows
}

func (*ShowCreateAllTypes) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614116)
	return TypeDML
}

func (*ShowCreateAllTypes) StatementTag() string {
	__antithesis_instrumentation__.Notify(614117)
	return "SHOW CREATE ALL TYPES"
}

func (*ShowCreateSchedules) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614118)
	return Rows
}

func (*ShowCreateSchedules) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614119)
	return TypeDML
}

func (*ShowCreateSchedules) StatementTag() string {
	__antithesis_instrumentation__.Notify(614120)
	return "SHOW CREATE SCHEDULES"
}

func (*ShowBackup) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614121)
	return Rows
}

func (*ShowBackup) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614122)
	return TypeDML
}

func (*ShowBackup) StatementTag() string {
	__antithesis_instrumentation__.Notify(614123)
	return "SHOW BACKUP"
}

func (*ShowBackup) cclOnlyStatement() { __antithesis_instrumentation__.Notify(614124) }

func (*ShowDatabases) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614125)
	return Rows
}

func (*ShowDatabases) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614126)
	return TypeDML
}

func (*ShowDatabases) StatementTag() string {
	__antithesis_instrumentation__.Notify(614127)
	return "SHOW DATABASES"
}

func (*ShowEnums) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614128)
	return Rows
}

func (*ShowEnums) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614129)
	return TypeDML
}

func (*ShowEnums) StatementTag() string {
	__antithesis_instrumentation__.Notify(614130)
	return "SHOW ENUMS"
}

func (*ShowTypes) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614131)
	return Rows
}

func (*ShowTypes) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614132)
	return TypeDML
}

func (*ShowTypes) StatementTag() string {
	__antithesis_instrumentation__.Notify(614133)
	return "SHOW TYPES"
}

func (*ShowTraceForSession) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614134)
	return Rows
}

func (*ShowTraceForSession) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614135)
	return TypeDML
}

func (*ShowTraceForSession) StatementTag() string {
	__antithesis_instrumentation__.Notify(614136)
	return "SHOW TRACE FOR SESSION"
}

func (*ShowGrants) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614137)
	return Rows
}

func (*ShowGrants) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614138)
	return TypeDML
}

func (*ShowGrants) StatementTag() string {
	__antithesis_instrumentation__.Notify(614139)
	return "SHOW GRANTS"
}

func (*ShowDatabaseIndexes) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614140)
	return Rows
}

func (*ShowDatabaseIndexes) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614141)
	return TypeDML
}

func (*ShowDatabaseIndexes) StatementTag() string {
	__antithesis_instrumentation__.Notify(614142)
	return "SHOW INDEXES FROM DATABASE"
}

func (*ShowIndexes) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614143)
	return Rows
}

func (*ShowIndexes) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614144)
	return TypeDML
}

func (*ShowIndexes) StatementTag() string {
	__antithesis_instrumentation__.Notify(614145)
	return "SHOW INDEXES FROM TABLE"
}

func (*ShowPartitions) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614146)
	return Rows
}

func (*ShowPartitions) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614147)
	return TypeDML
}

func (*ShowPartitions) StatementTag() string {
	__antithesis_instrumentation__.Notify(614148)
	return "SHOW PARTITIONS"
}

func (*ShowQueries) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614149)
	return Rows
}

func (*ShowQueries) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614150)
	return TypeDML
}

func (*ShowQueries) StatementTag() string {
	__antithesis_instrumentation__.Notify(614151)
	return "SHOW STATEMENTS"
}

func (*ShowJobs) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614152)
	return Rows
}

func (*ShowJobs) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614153)
	return TypeDML
}

func (*ShowJobs) StatementTag() string {
	__antithesis_instrumentation__.Notify(614154)
	return "SHOW JOBS"
}

func (*ShowChangefeedJobs) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614155)
	return Rows
}

func (*ShowChangefeedJobs) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614156)
	return TypeDML
}

func (*ShowChangefeedJobs) StatementTag() string {
	__antithesis_instrumentation__.Notify(614157)
	return "SHOW CHANGEFEED JOBS"
}

func (*ShowRoleGrants) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614158)
	return Rows
}

func (*ShowRoleGrants) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614159)
	return TypeDML
}

func (*ShowRoleGrants) StatementTag() string {
	__antithesis_instrumentation__.Notify(614160)
	return "SHOW GRANTS ON ROLE"
}

func (*ShowSessions) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614161)
	return Rows
}

func (*ShowSessions) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614162)
	return TypeDML
}

func (*ShowSessions) StatementTag() string {
	__antithesis_instrumentation__.Notify(614163)
	return "SHOW SESSIONS"
}

func (*ShowTableStats) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614164)
	return Rows
}

func (*ShowTableStats) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614165)
	return TypeDML
}

func (*ShowTableStats) StatementTag() string {
	__antithesis_instrumentation__.Notify(614166)
	return "SHOW STATISTICS"
}

func (*ShowHistogram) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614167)
	return Rows
}

func (*ShowHistogram) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614168)
	return TypeDML
}

func (*ShowHistogram) StatementTag() string {
	__antithesis_instrumentation__.Notify(614169)
	return "SHOW HISTOGRAM"
}

func (*ShowSchedules) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614170)
	return Rows
}

func (*ShowSchedules) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614171)
	return TypeDML
}

func (*ShowSchedules) StatementTag() string {
	__antithesis_instrumentation__.Notify(614172)
	return "SHOW SCHEDULES"
}

func (*ShowSyntax) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614173)
	return Rows
}

func (*ShowSyntax) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614174)
	return TypeDML
}

func (*ShowSyntax) StatementTag() string {
	__antithesis_instrumentation__.Notify(614175)
	return "SHOW SYNTAX"
}

func (*ShowSyntax) observerStatement() { __antithesis_instrumentation__.Notify(614176) }

func (*ShowTransactionStatus) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614177)
	return Rows
}

func (*ShowTransactionStatus) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614178)
	return TypeDML
}

func (*ShowTransactionStatus) StatementTag() string {
	__antithesis_instrumentation__.Notify(614179)
	return "SHOW TRANSACTION STATUS"
}

func (*ShowTransactionStatus) observerStatement() { __antithesis_instrumentation__.Notify(614180) }

func (*ShowTransferState) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614181)
	return Rows
}

func (*ShowTransferState) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614182)
	return TypeDML
}

func (*ShowTransferState) StatementTag() string {
	__antithesis_instrumentation__.Notify(614183)
	return "SHOW TRANSFER STATE"
}

func (*ShowTransferState) observerStatement() { __antithesis_instrumentation__.Notify(614184) }

func (*ShowSavepointStatus) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614185)
	return Rows
}

func (*ShowSavepointStatus) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614186)
	return TypeDML
}

func (*ShowSavepointStatus) StatementTag() string {
	__antithesis_instrumentation__.Notify(614187)
	return "SHOW SAVEPOINT STATUS"
}

func (*ShowSavepointStatus) observerStatement() { __antithesis_instrumentation__.Notify(614188) }

func (*ShowLastQueryStatistics) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614189)
	return Rows
}

func (*ShowLastQueryStatistics) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614190)
	return TypeDML
}

func (*ShowLastQueryStatistics) StatementTag() string {
	__antithesis_instrumentation__.Notify(614191)
	return "SHOW LAST QUERY STATISTICS"
}

func (*ShowLastQueryStatistics) observerStatement() { __antithesis_instrumentation__.Notify(614192) }

func (*ShowUsers) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614193)
	return Rows
}

func (*ShowUsers) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614194)
	return TypeDML
}

func (*ShowUsers) StatementTag() string {
	__antithesis_instrumentation__.Notify(614195)
	return "SHOW USERS"
}

func (*ShowFullTableScans) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614196)
	return Rows
}

func (*ShowFullTableScans) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614197)
	return TypeDML
}

func (*ShowFullTableScans) StatementTag() string {
	__antithesis_instrumentation__.Notify(614198)
	return "SHOW FULL TABLE SCANS"
}

func (*ShowRoles) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614199)
	return Rows
}

func (*ShowRoles) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614200)
	return TypeDML
}

func (*ShowRoles) StatementTag() string {
	__antithesis_instrumentation__.Notify(614201)
	return "SHOW ROLES"
}

func (*ShowZoneConfig) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614202)
	return Rows
}

func (*ShowZoneConfig) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614203)
	return TypeDML
}

func (*ShowZoneConfig) StatementTag() string {
	__antithesis_instrumentation__.Notify(614204)
	return "SHOW ZONE CONFIGURATION"
}

func (*ShowRanges) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614205)
	return Rows
}

func (*ShowRanges) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614206)
	return TypeDML
}

func (*ShowRanges) StatementTag() string {
	__antithesis_instrumentation__.Notify(614207)
	return "SHOW RANGES"
}

func (*ShowRangeForRow) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614208)
	return Rows
}

func (*ShowRangeForRow) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614209)
	return TypeDML
}

func (*ShowRangeForRow) StatementTag() string {
	__antithesis_instrumentation__.Notify(614210)
	return "SHOW RANGE FOR ROW"
}

func (*ShowSurvivalGoal) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614211)
	return Rows
}

func (*ShowSurvivalGoal) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614212)
	return TypeDML
}

func (*ShowSurvivalGoal) StatementTag() string {
	__antithesis_instrumentation__.Notify(614213)
	return "SHOW SURVIVAL GOAL"
}

func (*ShowRegions) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614214)
	return Rows
}

func (*ShowRegions) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614215)
	return TypeDML
}

func (*ShowRegions) StatementTag() string {
	__antithesis_instrumentation__.Notify(614216)
	return "SHOW REGIONS"
}

func (*ShowFingerprints) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614217)
	return Rows
}

func (*ShowFingerprints) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614218)
	return TypeDML
}

func (*ShowFingerprints) StatementTag() string {
	__antithesis_instrumentation__.Notify(614219)
	return "SHOW EXPERIMENTAL_FINGERPRINTS"
}

func (*ShowConstraints) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614220)
	return Rows
}

func (*ShowConstraints) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614221)
	return TypeDML
}

func (*ShowConstraints) StatementTag() string {
	__antithesis_instrumentation__.Notify(614222)
	return "SHOW CONSTRAINTS"
}

func (*ShowTables) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614223)
	return Rows
}

func (*ShowTables) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614224)
	return TypeDML
}

func (*ShowTables) StatementTag() string {
	__antithesis_instrumentation__.Notify(614225)
	return "SHOW TABLES"
}

func (*ShowTransactions) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614226)
	return Rows
}

func (*ShowTransactions) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614227)
	return TypeDML
}

func (*ShowTransactions) StatementTag() string {
	__antithesis_instrumentation__.Notify(614228)
	return "SHOW TRANSACTIONS"
}

func (*ShowSchemas) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614229)
	return Rows
}

func (*ShowSchemas) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614230)
	return TypeDML
}

func (*ShowSchemas) StatementTag() string {
	__antithesis_instrumentation__.Notify(614231)
	return "SHOW SCHEMAS"
}

func (*ShowSequences) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614232)
	return Rows
}

func (*ShowSequences) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614233)
	return TypeDML
}

func (*ShowSequences) StatementTag() string {
	__antithesis_instrumentation__.Notify(614234)
	return "SHOW SCHEMAS"
}

func (*ShowDefaultPrivileges) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614235)
	return Rows
}

func (*ShowDefaultPrivileges) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614236)
	return TypeDML
}

func (*ShowDefaultPrivileges) StatementTag() string {
	__antithesis_instrumentation__.Notify(614237)
	return "SHOW DEFAULT PRIVILEGES"
}

func (*ShowCompletions) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614238)
	return Rows
}

func (*ShowCompletions) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614239)
	return TypeDML
}

func (*ShowCompletions) StatementTag() string {
	__antithesis_instrumentation__.Notify(614240)
	return "SHOW COMPLETIONS"
}

func (*ShowCompletions) observerStatement() { __antithesis_instrumentation__.Notify(614241) }

func (*ShowCompletions) hiddenFromShowQueries() { __antithesis_instrumentation__.Notify(614242) }

func (*Split) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614243)
	return Rows
}

func (*Split) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614244)
	return TypeDML
}

func (*Split) StatementTag() string { __antithesis_instrumentation__.Notify(614245); return "SPLIT" }

func (*StreamIngestion) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614246)
	return Rows
}

func (*StreamIngestion) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614247)
	return TypeDML
}

func (*StreamIngestion) StatementTag() string {
	__antithesis_instrumentation__.Notify(614248)
	return "RESTORE FROM REPLICATION STREAM"
}

func (*StreamIngestion) cclOnlyStatement() { __antithesis_instrumentation__.Notify(614249) }

func (*Unsplit) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614250)
	return Rows
}

func (*Unsplit) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614251)
	return TypeDML
}

func (*Unsplit) StatementTag() string {
	__antithesis_instrumentation__.Notify(614252)
	return "UNSPLIT"
}

func (*Truncate) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614253)
	return Ack
}

func (*Truncate) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614254)
	return TypeDDL
}

func (*Truncate) StatementTag() string {
	__antithesis_instrumentation__.Notify(614255)
	return "TRUNCATE"
}

func (*Truncate) modifiesSchema() bool { __antithesis_instrumentation__.Notify(614256); return true }

func (n *Update) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614257)
	return n.Returning.statementReturnType()
}

func (*Update) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614258)
	return TypeDML
}

func (*Update) StatementTag() string { __antithesis_instrumentation__.Notify(614259); return "UPDATE" }

func (*UnionClause) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614260)
	return Rows
}

func (*UnionClause) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614261)
	return TypeDML
}

func (*UnionClause) StatementTag() string {
	__antithesis_instrumentation__.Notify(614262)
	return "UNION"
}

func (*ValuesClause) StatementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(614263)
	return Rows
}

func (*ValuesClause) StatementType() StatementType {
	__antithesis_instrumentation__.Notify(614264)
	return TypeDML
}

func (*ValuesClause) StatementTag() string {
	__antithesis_instrumentation__.Notify(614265)
	return "VALUES"
}

func (n *AlterChangefeed) String() string {
	__antithesis_instrumentation__.Notify(614266)
	return AsString(n)
}
func (n *AlterChangefeedCmds) String() string {
	__antithesis_instrumentation__.Notify(614267)
	return AsString(n)
}
func (n *AlterBackup) String() string {
	__antithesis_instrumentation__.Notify(614268)
	return AsString(n)
}
func (n *AlterIndex) String() string {
	__antithesis_instrumentation__.Notify(614269)
	return AsString(n)
}
func (n *AlterDatabaseOwner) String() string {
	__antithesis_instrumentation__.Notify(614270)
	return AsString(n)
}
func (n *AlterDatabaseAddRegion) String() string {
	__antithesis_instrumentation__.Notify(614271)
	return AsString(n)
}
func (n *AlterDatabaseDropRegion) String() string {
	__antithesis_instrumentation__.Notify(614272)
	return AsString(n)
}
func (n *AlterDatabaseSurvivalGoal) String() string {
	__antithesis_instrumentation__.Notify(614273)
	return AsString(n)
}
func (n *AlterDatabasePlacement) String() string {
	__antithesis_instrumentation__.Notify(614274)
	return AsString(n)
}
func (n *AlterDatabasePrimaryRegion) String() string {
	__antithesis_instrumentation__.Notify(614275)
	return AsString(n)
}
func (n *AlterDatabaseAddSuperRegion) String() string {
	__antithesis_instrumentation__.Notify(614276)
	return AsString(n)
}
func (n *AlterDatabaseDropSuperRegion) String() string {
	__antithesis_instrumentation__.Notify(614277)
	return AsString(n)
}
func (n *AlterDatabaseAlterSuperRegion) String() string {
	__antithesis_instrumentation__.Notify(614278)
	return AsString(n)
}
func (n *AlterDefaultPrivileges) String() string {
	__antithesis_instrumentation__.Notify(614279)
	return AsString(n)
}
func (n *AlterSchema) String() string {
	__antithesis_instrumentation__.Notify(614280)
	return AsString(n)
}
func (n *AlterTable) String() string {
	__antithesis_instrumentation__.Notify(614281)
	return AsString(n)
}
func (n *AlterTableCmds) String() string {
	__antithesis_instrumentation__.Notify(614282)
	return AsString(n)
}
func (n *AlterTableAddColumn) String() string {
	__antithesis_instrumentation__.Notify(614283)
	return AsString(n)
}
func (n *AlterTableAddConstraint) String() string {
	__antithesis_instrumentation__.Notify(614284)
	return AsString(n)
}
func (n *AlterTableAlterColumnType) String() string {
	__antithesis_instrumentation__.Notify(614285)
	return AsString(n)
}
func (n *AlterTableDropColumn) String() string {
	__antithesis_instrumentation__.Notify(614286)
	return AsString(n)
}
func (n *AlterTableDropConstraint) String() string {
	__antithesis_instrumentation__.Notify(614287)
	return AsString(n)
}
func (n *AlterTableDropNotNull) String() string {
	__antithesis_instrumentation__.Notify(614288)
	return AsString(n)
}
func (n *AlterTableDropStored) String() string {
	__antithesis_instrumentation__.Notify(614289)
	return AsString(n)
}
func (n *AlterTableLocality) String() string {
	__antithesis_instrumentation__.Notify(614290)
	return AsString(n)
}
func (n *AlterTableSetDefault) String() string {
	__antithesis_instrumentation__.Notify(614291)
	return AsString(n)
}
func (n *AlterTableSetVisible) String() string {
	__antithesis_instrumentation__.Notify(614292)
	return AsString(n)
}
func (n *AlterTableSetNotNull) String() string {
	__antithesis_instrumentation__.Notify(614293)
	return AsString(n)
}
func (n *AlterTableOwner) String() string {
	__antithesis_instrumentation__.Notify(614294)
	return AsString(n)
}
func (n *AlterTableSetSchema) String() string {
	__antithesis_instrumentation__.Notify(614295)
	return AsString(n)
}
func (n *AlterTenantSetClusterSetting) String() string {
	__antithesis_instrumentation__.Notify(614296)
	return AsString(n)
}
func (n *AlterType) String() string {
	__antithesis_instrumentation__.Notify(614297)
	return AsString(n)
}
func (n *AlterRole) String() string {
	__antithesis_instrumentation__.Notify(614298)
	return AsString(n)
}
func (n *AlterRoleSet) String() string {
	__antithesis_instrumentation__.Notify(614299)
	return AsString(n)
}
func (n *AlterSequence) String() string {
	__antithesis_instrumentation__.Notify(614300)
	return AsString(n)
}
func (n *Analyze) String() string { __antithesis_instrumentation__.Notify(614301); return AsString(n) }
func (n *Backup) String() string  { __antithesis_instrumentation__.Notify(614302); return AsString(n) }
func (n *BeginTransaction) String() string {
	__antithesis_instrumentation__.Notify(614303)
	return AsString(n)
}
func (n *ControlJobs) String() string {
	__antithesis_instrumentation__.Notify(614304)
	return AsString(n)
}
func (n *ControlSchedules) String() string {
	__antithesis_instrumentation__.Notify(614305)
	return AsString(n)
}
func (n *ControlJobsForSchedules) String() string {
	__antithesis_instrumentation__.Notify(614306)
	return AsString(n)
}
func (n *ControlJobsOfType) String() string {
	__antithesis_instrumentation__.Notify(614307)
	return AsString(n)
}
func (n *CancelQueries) String() string {
	__antithesis_instrumentation__.Notify(614308)
	return AsString(n)
}
func (n *CancelSessions) String() string {
	__antithesis_instrumentation__.Notify(614309)
	return AsString(n)
}
func (n *CannedOptPlan) String() string {
	__antithesis_instrumentation__.Notify(614310)
	return AsString(n)
}
func (n *CloseCursor) String() string {
	__antithesis_instrumentation__.Notify(614311)
	return AsString(n)
}
func (n *CommentOnColumn) String() string {
	__antithesis_instrumentation__.Notify(614312)
	return AsString(n)
}
func (n *CommentOnConstraint) String() string {
	__antithesis_instrumentation__.Notify(614313)
	return AsString(n)
}
func (n *CommentOnDatabase) String() string {
	__antithesis_instrumentation__.Notify(614314)
	return AsString(n)
}
func (n *CommentOnSchema) String() string {
	__antithesis_instrumentation__.Notify(614315)
	return AsString(n)
}
func (n *CommentOnIndex) String() string {
	__antithesis_instrumentation__.Notify(614316)
	return AsString(n)
}
func (n *CommentOnTable) String() string {
	__antithesis_instrumentation__.Notify(614317)
	return AsString(n)
}
func (n *CommitTransaction) String() string {
	__antithesis_instrumentation__.Notify(614318)
	return AsString(n)
}
func (n *CopyFrom) String() string { __antithesis_instrumentation__.Notify(614319); return AsString(n) }
func (n *CreateChangefeed) String() string {
	__antithesis_instrumentation__.Notify(614320)
	return AsString(n)
}
func (n *CreateDatabase) String() string {
	__antithesis_instrumentation__.Notify(614321)
	return AsString(n)
}
func (n *CreateExtension) String() string {
	__antithesis_instrumentation__.Notify(614322)
	return AsString(n)
}
func (n *CreateIndex) String() string {
	__antithesis_instrumentation__.Notify(614323)
	return AsString(n)
}
func (n *CreateRole) String() string {
	__antithesis_instrumentation__.Notify(614324)
	return AsString(n)
}
func (n *CreateTable) String() string {
	__antithesis_instrumentation__.Notify(614325)
	return AsString(n)
}
func (n *CreateSchema) String() string {
	__antithesis_instrumentation__.Notify(614326)
	return AsString(n)
}
func (n *CreateSequence) String() string {
	__antithesis_instrumentation__.Notify(614327)
	return AsString(n)
}
func (n *CreateStats) String() string {
	__antithesis_instrumentation__.Notify(614328)
	return AsString(n)
}
func (n *CreateView) String() string {
	__antithesis_instrumentation__.Notify(614329)
	return AsString(n)
}
func (n *Deallocate) String() string {
	__antithesis_instrumentation__.Notify(614330)
	return AsString(n)
}
func (n *Delete) String() string { __antithesis_instrumentation__.Notify(614331); return AsString(n) }
func (n *DeclareCursor) String() string {
	__antithesis_instrumentation__.Notify(614332)
	return AsString(n)
}
func (n *DropDatabase) String() string {
	__antithesis_instrumentation__.Notify(614333)
	return AsString(n)
}
func (n *DropIndex) String() string {
	__antithesis_instrumentation__.Notify(614334)
	return AsString(n)
}
func (n *DropOwnedBy) String() string {
	__antithesis_instrumentation__.Notify(614335)
	return AsString(n)
}
func (n *DropSchema) String() string {
	__antithesis_instrumentation__.Notify(614336)
	return AsString(n)
}
func (n *DropSequence) String() string {
	__antithesis_instrumentation__.Notify(614337)
	return AsString(n)
}
func (n *DropTable) String() string {
	__antithesis_instrumentation__.Notify(614338)
	return AsString(n)
}
func (n *DropType) String() string { __antithesis_instrumentation__.Notify(614339); return AsString(n) }
func (n *DropView) String() string { __antithesis_instrumentation__.Notify(614340); return AsString(n) }
func (n *DropRole) String() string { __antithesis_instrumentation__.Notify(614341); return AsString(n) }
func (n *Execute) String() string  { __antithesis_instrumentation__.Notify(614342); return AsString(n) }
func (n *Explain) String() string  { __antithesis_instrumentation__.Notify(614343); return AsString(n) }
func (n *ExplainAnalyze) String() string {
	__antithesis_instrumentation__.Notify(614344)
	return AsString(n)
}
func (n *Export) String() string { __antithesis_instrumentation__.Notify(614345); return AsString(n) }
func (n *FetchCursor) String() string {
	__antithesis_instrumentation__.Notify(614346)
	return AsString(n)
}
func (n *Grant) String() string { __antithesis_instrumentation__.Notify(614347); return AsString(n) }
func (n *GrantRole) String() string {
	__antithesis_instrumentation__.Notify(614348)
	return AsString(n)
}
func (n *MoveCursor) String() string {
	__antithesis_instrumentation__.Notify(614349)
	return AsString(n)
}
func (n *Insert) String() string { __antithesis_instrumentation__.Notify(614350); return AsString(n) }
func (n *Import) String() string { __antithesis_instrumentation__.Notify(614351); return AsString(n) }
func (n *ParenSelect) String() string {
	__antithesis_instrumentation__.Notify(614352)
	return AsString(n)
}
func (n *Prepare) String() string { __antithesis_instrumentation__.Notify(614353); return AsString(n) }
func (n *ReassignOwnedBy) String() string {
	__antithesis_instrumentation__.Notify(614354)
	return AsString(n)
}
func (n *ReleaseSavepoint) String() string {
	__antithesis_instrumentation__.Notify(614355)
	return AsString(n)
}
func (n *Relocate) String() string { __antithesis_instrumentation__.Notify(614356); return AsString(n) }
func (n *RelocateRange) String() string {
	__antithesis_instrumentation__.Notify(614357)
	return AsString(n)
}
func (n *RefreshMaterializedView) String() string {
	__antithesis_instrumentation__.Notify(614358)
	return AsString(n)
}
func (n *RenameColumn) String() string {
	__antithesis_instrumentation__.Notify(614359)
	return AsString(n)
}
func (n *RenameDatabase) String() string {
	__antithesis_instrumentation__.Notify(614360)
	return AsString(n)
}
func (n *ReparentDatabase) String() string {
	__antithesis_instrumentation__.Notify(614361)
	return AsString(n)
}
func (n *ReplicationStream) String() string {
	__antithesis_instrumentation__.Notify(614362)
	return AsString(n)
}
func (n *RenameIndex) String() string {
	__antithesis_instrumentation__.Notify(614363)
	return AsString(n)
}
func (n *RenameTable) String() string {
	__antithesis_instrumentation__.Notify(614364)
	return AsString(n)
}
func (n *Restore) String() string { __antithesis_instrumentation__.Notify(614365); return AsString(n) }
func (n *Revoke) String() string  { __antithesis_instrumentation__.Notify(614366); return AsString(n) }
func (n *RevokeRole) String() string {
	__antithesis_instrumentation__.Notify(614367)
	return AsString(n)
}
func (n *RollbackToSavepoint) String() string {
	__antithesis_instrumentation__.Notify(614368)
	return AsString(n)
}
func (n *RollbackTransaction) String() string {
	__antithesis_instrumentation__.Notify(614369)
	return AsString(n)
}
func (n *Savepoint) String() string {
	__antithesis_instrumentation__.Notify(614370)
	return AsString(n)
}
func (n *Scatter) String() string { __antithesis_instrumentation__.Notify(614371); return AsString(n) }
func (n *ScheduledBackup) String() string {
	__antithesis_instrumentation__.Notify(614372)
	return AsString(n)
}
func (n *Scrub) String() string  { __antithesis_instrumentation__.Notify(614373); return AsString(n) }
func (n *Select) String() string { __antithesis_instrumentation__.Notify(614374); return AsString(n) }
func (n *SelectClause) String() string {
	__antithesis_instrumentation__.Notify(614375)
	return AsString(n)
}
func (n *SetClusterSetting) String() string {
	__antithesis_instrumentation__.Notify(614376)
	return AsString(n)
}
func (n *SetZoneConfig) String() string {
	__antithesis_instrumentation__.Notify(614377)
	return AsString(n)
}
func (n *SetSessionAuthorizationDefault) String() string {
	__antithesis_instrumentation__.Notify(614378)
	return AsString(n)
}
func (n *SetSessionCharacteristics) String() string {
	__antithesis_instrumentation__.Notify(614379)
	return AsString(n)
}
func (n *SetTransaction) String() string {
	__antithesis_instrumentation__.Notify(614380)
	return AsString(n)
}
func (n *SetTracing) String() string {
	__antithesis_instrumentation__.Notify(614381)
	return AsString(n)
}
func (n *SetVar) String() string { __antithesis_instrumentation__.Notify(614382); return AsString(n) }
func (n *ShowBackup) String() string {
	__antithesis_instrumentation__.Notify(614383)
	return AsString(n)
}
func (n *ShowClusterSetting) String() string {
	__antithesis_instrumentation__.Notify(614384)
	return AsString(n)
}
func (n *ShowClusterSettingList) String() string {
	__antithesis_instrumentation__.Notify(614385)
	return AsString(n)
}
func (n *ShowTenantClusterSetting) String() string {
	__antithesis_instrumentation__.Notify(614386)
	return AsString(n)
}
func (n *ShowTenantClusterSettingList) String() string {
	__antithesis_instrumentation__.Notify(614387)
	return AsString(n)
}
func (n *ShowColumns) String() string {
	__antithesis_instrumentation__.Notify(614388)
	return AsString(n)
}
func (n *ShowConstraints) String() string {
	__antithesis_instrumentation__.Notify(614389)
	return AsString(n)
}
func (n *ShowCreate) String() string {
	__antithesis_instrumentation__.Notify(614390)
	return AsString(n)
}
func (node *ShowCreateAllSchemas) String() string {
	__antithesis_instrumentation__.Notify(614391)
	return AsString(node)
}
func (node *ShowCreateAllTables) String() string {
	__antithesis_instrumentation__.Notify(614392)
	return AsString(node)
}
func (node *ShowCreateAllTypes) String() string {
	__antithesis_instrumentation__.Notify(614393)
	return AsString(node)
}
func (n *ShowCreateSchedules) String() string {
	__antithesis_instrumentation__.Notify(614394)
	return AsString(n)
}
func (n *ShowDatabases) String() string {
	__antithesis_instrumentation__.Notify(614395)
	return AsString(n)
}
func (n *ShowDatabaseIndexes) String() string {
	__antithesis_instrumentation__.Notify(614396)
	return AsString(n)
}
func (n *ShowEnums) String() string {
	__antithesis_instrumentation__.Notify(614397)
	return AsString(n)
}
func (n *ShowFullTableScans) String() string {
	__antithesis_instrumentation__.Notify(614398)
	return AsString(n)
}
func (n *ShowGrants) String() string {
	__antithesis_instrumentation__.Notify(614399)
	return AsString(n)
}
func (n *ShowHistogram) String() string {
	__antithesis_instrumentation__.Notify(614400)
	return AsString(n)
}
func (n *ShowSchedules) String() string {
	__antithesis_instrumentation__.Notify(614401)
	return AsString(n)
}
func (n *ShowIndexes) String() string {
	__antithesis_instrumentation__.Notify(614402)
	return AsString(n)
}
func (n *ShowJobs) String() string { __antithesis_instrumentation__.Notify(614403); return AsString(n) }
func (n *ShowChangefeedJobs) String() string {
	__antithesis_instrumentation__.Notify(614404)
	return AsString(n)
}
func (n *ShowLastQueryStatistics) String() string {
	__antithesis_instrumentation__.Notify(614405)
	return AsString(n)
}
func (n *ShowPartitions) String() string {
	__antithesis_instrumentation__.Notify(614406)
	return AsString(n)
}
func (n *ShowQueries) String() string {
	__antithesis_instrumentation__.Notify(614407)
	return AsString(n)
}
func (n *ShowRanges) String() string {
	__antithesis_instrumentation__.Notify(614408)
	return AsString(n)
}
func (n *ShowRangeForRow) String() string {
	__antithesis_instrumentation__.Notify(614409)
	return AsString(n)
}
func (n *ShowRegions) String() string {
	__antithesis_instrumentation__.Notify(614410)
	return AsString(n)
}
func (n *ShowRoleGrants) String() string {
	__antithesis_instrumentation__.Notify(614411)
	return AsString(n)
}
func (n *ShowRoles) String() string {
	__antithesis_instrumentation__.Notify(614412)
	return AsString(n)
}
func (n *ShowSavepointStatus) String() string {
	__antithesis_instrumentation__.Notify(614413)
	return AsString(n)
}
func (n *ShowSchemas) String() string {
	__antithesis_instrumentation__.Notify(614414)
	return AsString(n)
}
func (n *ShowSequences) String() string {
	__antithesis_instrumentation__.Notify(614415)
	return AsString(n)
}
func (n *ShowSessions) String() string {
	__antithesis_instrumentation__.Notify(614416)
	return AsString(n)
}
func (n *ShowSurvivalGoal) String() string {
	__antithesis_instrumentation__.Notify(614417)
	return AsString(n)
}
func (n *ShowSyntax) String() string {
	__antithesis_instrumentation__.Notify(614418)
	return AsString(n)
}
func (n *ShowTableStats) String() string {
	__antithesis_instrumentation__.Notify(614419)
	return AsString(n)
}
func (n *ShowTables) String() string {
	__antithesis_instrumentation__.Notify(614420)
	return AsString(n)
}
func (n *ShowTypes) String() string {
	__antithesis_instrumentation__.Notify(614421)
	return AsString(n)
}
func (n *ShowTraceForSession) String() string {
	__antithesis_instrumentation__.Notify(614422)
	return AsString(n)
}
func (n *ShowTransactionStatus) String() string {
	__antithesis_instrumentation__.Notify(614423)
	return AsString(n)
}
func (n *ShowTransactions) String() string {
	__antithesis_instrumentation__.Notify(614424)
	return AsString(n)
}
func (n *ShowTransferState) String() string {
	__antithesis_instrumentation__.Notify(614425)
	return AsString(n)
}
func (n *ShowUsers) String() string {
	__antithesis_instrumentation__.Notify(614426)
	return AsString(n)
}
func (n *ShowVar) String() string { __antithesis_instrumentation__.Notify(614427); return AsString(n) }
func (n *ShowZoneConfig) String() string {
	__antithesis_instrumentation__.Notify(614428)
	return AsString(n)
}
func (n *ShowFingerprints) String() string {
	__antithesis_instrumentation__.Notify(614429)
	return AsString(n)
}
func (n *ShowDefaultPrivileges) String() string {
	__antithesis_instrumentation__.Notify(614430)
	return AsString(n)
}
func (n *ShowCompletions) String() string {
	__antithesis_instrumentation__.Notify(614431)
	return AsString(n)
}
func (n *Split) String() string { __antithesis_instrumentation__.Notify(614432); return AsString(n) }
func (n *StreamIngestion) String() string {
	__antithesis_instrumentation__.Notify(614433)
	return AsString(n)
}
func (n *Unsplit) String() string  { __antithesis_instrumentation__.Notify(614434); return AsString(n) }
func (n *Truncate) String() string { __antithesis_instrumentation__.Notify(614435); return AsString(n) }
func (n *UnionClause) String() string {
	__antithesis_instrumentation__.Notify(614436)
	return AsString(n)
}
func (n *Update) String() string { __antithesis_instrumentation__.Notify(614437); return AsString(n) }
func (n *ValuesClause) String() string {
	__antithesis_instrumentation__.Notify(614438)
	return AsString(n)
}
