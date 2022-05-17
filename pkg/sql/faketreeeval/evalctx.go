// Package faketreeeval provides fake implementations of tree eval interfaces.
package faketreeeval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgnotice"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/roleoption"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"
)

type DummySequenceOperators struct{}

var _ tree.SequenceOperators = &DummySequenceOperators{}

var errSequenceOperators = unimplemented.NewWithIssue(42508,
	"cannot evaluate scalar expressions containing sequence operations in this context")

func (so *DummySequenceOperators) GetSerialSequenceNameFromColumn(
	ctx context.Context, tn *tree.TableName, columnName tree.Name,
) (*tree.TableName, error) {
	__antithesis_instrumentation__.Notify(491433)
	return nil, errors.WithStack(errSequenceOperators)
}

func (so *DummySequenceOperators) ParseQualifiedTableName(sql string) (*tree.TableName, error) {
	__antithesis_instrumentation__.Notify(491434)
	return nil, errors.WithStack(errSequenceOperators)
}

func (so *DummySequenceOperators) ResolveTableName(
	ctx context.Context, tn *tree.TableName,
) (tree.ID, error) {
	__antithesis_instrumentation__.Notify(491435)
	return 0, errors.WithStack(errSequenceOperators)
}

func (so *DummySequenceOperators) SchemaExists(
	ctx context.Context, dbName, scName string,
) (bool, error) {
	__antithesis_instrumentation__.Notify(491436)
	return false, errors.WithStack(errSequenceOperators)
}

func (so *DummySequenceOperators) IsTableVisible(
	ctx context.Context, curDB string, searchPath sessiondata.SearchPath, tableID oid.Oid,
) (bool, bool, error) {
	__antithesis_instrumentation__.Notify(491437)
	return false, false, errors.WithStack(errSequenceOperators)
}

func (so *DummySequenceOperators) IsTypeVisible(
	ctx context.Context, curDB string, searchPath sessiondata.SearchPath, typeID oid.Oid,
) (bool, bool, error) {
	__antithesis_instrumentation__.Notify(491438)
	return false, false, errors.WithStack(errEvalPlanner)
}

func (so *DummySequenceOperators) HasAnyPrivilege(
	ctx context.Context,
	specifier tree.HasPrivilegeSpecifier,
	user security.SQLUsername,
	privs []privilege.Privilege,
) (tree.HasAnyPrivilegeResult, error) {
	__antithesis_instrumentation__.Notify(491439)
	return tree.HasNoPrivilege, errors.WithStack(errEvalPlanner)
}

func (so *DummySequenceOperators) IncrementSequenceByID(
	ctx context.Context, seqID int64,
) (int64, error) {
	__antithesis_instrumentation__.Notify(491440)
	return 0, errors.WithStack(errSequenceOperators)
}

func (so *DummySequenceOperators) GetLatestValueInSessionForSequenceByID(
	ctx context.Context, seqID int64,
) (int64, error) {
	__antithesis_instrumentation__.Notify(491441)
	return 0, errors.WithStack(errSequenceOperators)
}

func (so *DummySequenceOperators) SetSequenceValueByID(
	ctx context.Context, seqID uint32, newVal int64, isCalled bool,
) error {
	__antithesis_instrumentation__.Notify(491442)
	return errors.WithStack(errSequenceOperators)
}

type DummyRegionOperator struct{}

var _ tree.RegionOperator = &DummyRegionOperator{}

var errRegionOperator = unimplemented.NewWithIssue(42508,
	"cannot evaluate scalar expressions containing region operations in this context")

func (so *DummyRegionOperator) CurrentDatabaseRegionConfig(
	_ context.Context,
) (tree.DatabaseRegionConfig, error) {
	__antithesis_instrumentation__.Notify(491443)
	return nil, errors.WithStack(errRegionOperator)
}

func (so *DummyRegionOperator) ValidateAllMultiRegionZoneConfigsInCurrentDatabase(
	_ context.Context,
) error {
	__antithesis_instrumentation__.Notify(491444)
	return errors.WithStack(errRegionOperator)
}

func (so *DummyRegionOperator) ResetMultiRegionZoneConfigsForTable(
	_ context.Context, id int64,
) error {
	__antithesis_instrumentation__.Notify(491445)
	return errors.WithStack(errRegionOperator)
}

func (so *DummyRegionOperator) ResetMultiRegionZoneConfigsForDatabase(
	_ context.Context, id int64,
) error {
	__antithesis_instrumentation__.Notify(491446)
	return errors.WithStack(errRegionOperator)
}

type DummyEvalPlanner struct{}

func (ep *DummyEvalPlanner) ResolveOIDFromString(
	ctx context.Context, resultType *types.T, toResolve *tree.DString,
) (*tree.DOid, error) {
	__antithesis_instrumentation__.Notify(491447)
	return nil, errors.WithStack(errEvalPlanner)
}

func (ep *DummyEvalPlanner) ResolveOIDFromOID(
	ctx context.Context, resultType *types.T, toResolve *tree.DOid,
) (*tree.DOid, error) {
	__antithesis_instrumentation__.Notify(491448)
	return nil, errors.WithStack(errEvalPlanner)
}

func (ep *DummyEvalPlanner) UnsafeUpsertDescriptor(
	ctx context.Context, descID int64, encodedDescriptor []byte, force bool,
) error {
	__antithesis_instrumentation__.Notify(491449)
	return errors.WithStack(errEvalPlanner)
}

func (ep *DummyEvalPlanner) GetImmutableTableInterfaceByID(
	ctx context.Context, id int,
) (interface{}, error) {
	__antithesis_instrumentation__.Notify(491450)
	return nil, errors.WithStack(errEvalPlanner)
}

func (ep *DummyEvalPlanner) UnsafeDeleteDescriptor(
	ctx context.Context, descID int64, force bool,
) error {
	__antithesis_instrumentation__.Notify(491451)
	return errors.WithStack(errEvalPlanner)
}

func (ep *DummyEvalPlanner) ForceDeleteTableData(ctx context.Context, descID int64) error {
	__antithesis_instrumentation__.Notify(491452)
	return errors.WithStack(errEvalPlanner)
}

func (ep *DummyEvalPlanner) UnsafeUpsertNamespaceEntry(
	ctx context.Context, parentID, parentSchemaID int64, name string, descID int64, force bool,
) error {
	__antithesis_instrumentation__.Notify(491453)
	return errors.WithStack(errEvalPlanner)
}

func (ep *DummyEvalPlanner) UnsafeDeleteNamespaceEntry(
	ctx context.Context, parentID, parentSchemaID int64, name string, descID int64, force bool,
) error {
	__antithesis_instrumentation__.Notify(491454)
	return errors.WithStack(errEvalPlanner)
}

func (ep *DummyEvalPlanner) UserHasAdminRole(
	ctx context.Context, user security.SQLUsername,
) (bool, error) {
	__antithesis_instrumentation__.Notify(491455)
	return false, errors.WithStack(errEvalPlanner)
}

func (ep *DummyEvalPlanner) MemberOfWithAdminOption(
	ctx context.Context, member security.SQLUsername,
) (map[security.SQLUsername]bool, error) {
	__antithesis_instrumentation__.Notify(491456)
	return nil, errors.WithStack(errEvalPlanner)
}

func (*DummyEvalPlanner) ExternalReadFile(ctx context.Context, uri string) ([]byte, error) {
	__antithesis_instrumentation__.Notify(491457)
	return nil, errors.WithStack(errEvalPlanner)
}

func (*DummyEvalPlanner) ExternalWriteFile(ctx context.Context, uri string, content []byte) error {
	__antithesis_instrumentation__.Notify(491458)
	return errors.WithStack(errEvalPlanner)
}

func (*DummyEvalPlanner) DecodeGist(gist string) ([]string, error) {
	__antithesis_instrumentation__.Notify(491459)
	return nil, errors.WithStack(errEvalPlanner)
}

func (*DummyEvalPlanner) SerializeSessionState() (*tree.DBytes, error) {
	__antithesis_instrumentation__.Notify(491460)
	return nil, errors.WithStack(errEvalPlanner)
}

func (*DummyEvalPlanner) DeserializeSessionState(token *tree.DBytes) (*tree.DBool, error) {
	__antithesis_instrumentation__.Notify(491461)
	return nil, errors.WithStack(errEvalPlanner)
}

func (*DummyEvalPlanner) CreateSessionRevivalToken() (*tree.DBytes, error) {
	__antithesis_instrumentation__.Notify(491462)
	return nil, errors.WithStack(errEvalPlanner)
}

func (*DummyEvalPlanner) ValidateSessionRevivalToken(token *tree.DBytes) (*tree.DBool, error) {
	__antithesis_instrumentation__.Notify(491463)
	return nil, errors.WithStack(errEvalPlanner)
}

func (*DummyEvalPlanner) RevalidateUniqueConstraintsInCurrentDB(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(491464)
	return errors.WithStack(errEvalPlanner)
}

func (*DummyEvalPlanner) RevalidateUniqueConstraintsInTable(
	ctx context.Context, tableID int,
) error {
	__antithesis_instrumentation__.Notify(491465)
	return errors.WithStack(errEvalPlanner)
}

func (*DummyEvalPlanner) RevalidateUniqueConstraint(
	ctx context.Context, tableID int, constraintName string,
) error {
	__antithesis_instrumentation__.Notify(491466)
	return errors.WithStack(errEvalPlanner)
}

func (*DummyEvalPlanner) ValidateTTLScheduledJobsInCurrentDB(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(491467)
	return errors.WithStack(errEvalPlanner)
}

func (*DummyEvalPlanner) RepairTTLScheduledJobForTable(ctx context.Context, tableID int64) error {
	__antithesis_instrumentation__.Notify(491468)
	return errors.WithStack(errEvalPlanner)
}

func (*DummyEvalPlanner) ExecutorConfig() interface{} {
	__antithesis_instrumentation__.Notify(491469)
	return nil
}

var _ tree.EvalPlanner = &DummyEvalPlanner{}

var errEvalPlanner = pgerror.New(pgcode.ScalarOperationCannotRunWithoutFullSessionContext,
	"cannot evaluate scalar expressions using table lookups in this context")

func (ep *DummyEvalPlanner) CurrentDatabaseRegionConfig(
	_ context.Context,
) (tree.DatabaseRegionConfig, error) {
	__antithesis_instrumentation__.Notify(491470)
	return nil, errors.WithStack(errEvalPlanner)
}

func (ep *DummyEvalPlanner) ResetMultiRegionZoneConfigsForTable(_ context.Context, _ int64) error {
	__antithesis_instrumentation__.Notify(491471)
	return errors.WithStack(errEvalPlanner)
}

func (ep *DummyEvalPlanner) ResetMultiRegionZoneConfigsForDatabase(
	_ context.Context, _ int64,
) error {
	__antithesis_instrumentation__.Notify(491472)
	return errors.WithStack(errEvalPlanner)
}

func (ep *DummyEvalPlanner) ValidateAllMultiRegionZoneConfigsInCurrentDatabase(
	_ context.Context,
) error {
	__antithesis_instrumentation__.Notify(491473)
	return errors.WithStack(errEvalPlanner)
}

func (ep *DummyEvalPlanner) ParseQualifiedTableName(sql string) (*tree.TableName, error) {
	__antithesis_instrumentation__.Notify(491474)
	return parser.ParseQualifiedTableName(sql)
}

func (ep *DummyEvalPlanner) SchemaExists(ctx context.Context, dbName, scName string) (bool, error) {
	__antithesis_instrumentation__.Notify(491475)
	return false, errors.WithStack(errEvalPlanner)
}

func (ep *DummyEvalPlanner) IsTableVisible(
	ctx context.Context, curDB string, searchPath sessiondata.SearchPath, tableID oid.Oid,
) (bool, bool, error) {
	__antithesis_instrumentation__.Notify(491476)
	return false, false, errors.WithStack(errEvalPlanner)
}

func (ep *DummyEvalPlanner) IsTypeVisible(
	ctx context.Context, curDB string, searchPath sessiondata.SearchPath, typeID oid.Oid,
) (bool, bool, error) {
	__antithesis_instrumentation__.Notify(491477)
	return false, false, errors.WithStack(errEvalPlanner)
}

func (ep *DummyEvalPlanner) HasAnyPrivilege(
	ctx context.Context,
	specifier tree.HasPrivilegeSpecifier,
	user security.SQLUsername,
	privs []privilege.Privilege,
) (tree.HasAnyPrivilegeResult, error) {
	__antithesis_instrumentation__.Notify(491478)
	return tree.HasNoPrivilege, errors.WithStack(errEvalPlanner)
}

func (ep *DummyEvalPlanner) ResolveTableName(
	ctx context.Context, tn *tree.TableName,
) (tree.ID, error) {
	__antithesis_instrumentation__.Notify(491479)
	return 0, errors.WithStack(errEvalPlanner)
}

func (ep *DummyEvalPlanner) GetTypeFromValidSQLSyntax(sql string) (*types.T, error) {
	__antithesis_instrumentation__.Notify(491480)
	return nil, errors.WithStack(errEvalPlanner)
}

func (ep *DummyEvalPlanner) EvalSubquery(expr *tree.Subquery) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(491481)
	return nil, errors.WithStack(errEvalPlanner)
}

func (ep *DummyEvalPlanner) ResolveTypeByOID(_ context.Context, _ oid.Oid) (*types.T, error) {
	__antithesis_instrumentation__.Notify(491482)
	return nil, errors.WithStack(errEvalPlanner)
}

func (ep *DummyEvalPlanner) ResolveType(
	_ context.Context, _ *tree.UnresolvedObjectName,
) (*types.T, error) {
	__antithesis_instrumentation__.Notify(491483)
	return nil, errors.WithStack(errEvalPlanner)
}

func (ep *DummyEvalPlanner) QueryRowEx(
	ctx context.Context,
	opName string,
	txn *kv.Txn,
	session sessiondata.InternalExecutorOverride,
	stmt string,
	qargs ...interface{},
) (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(491484)
	return nil, errors.WithStack(errEvalPlanner)
}

func (ep *DummyEvalPlanner) QueryIteratorEx(
	ctx context.Context,
	opName string,
	txn *kv.Txn,
	session sessiondata.InternalExecutorOverride,
	stmt string,
	qargs ...interface{},
) (tree.InternalRows, error) {
	__antithesis_instrumentation__.Notify(491485)
	return nil, errors.WithStack(errEvalPlanner)
}

type DummyPrivilegedAccessor struct{}

var _ tree.PrivilegedAccessor = &DummyPrivilegedAccessor{}

var errEvalPrivileged = pgerror.New(pgcode.ScalarOperationCannotRunWithoutFullSessionContext,
	"cannot evaluate privileged expressions in this context")

func (ep *DummyPrivilegedAccessor) LookupNamespaceID(
	ctx context.Context, parentID int64, parentSchemaID int64, name string,
) (tree.DInt, bool, error) {
	__antithesis_instrumentation__.Notify(491486)
	return 0, false, errors.WithStack(errEvalPrivileged)
}

func (ep *DummyPrivilegedAccessor) LookupZoneConfigByNamespaceID(
	ctx context.Context, id int64,
) (tree.DBytes, bool, error) {
	__antithesis_instrumentation__.Notify(491487)
	return "", false, errors.WithStack(errEvalPrivileged)
}

type DummySessionAccessor struct{}

var _ tree.EvalSessionAccessor = &DummySessionAccessor{}

var errEvalSessionVar = pgerror.New(pgcode.ScalarOperationCannotRunWithoutFullSessionContext,
	"cannot evaluate scalar expressions that access session variables in this context")

func (ep *DummySessionAccessor) GetSessionVar(
	_ context.Context, _ string, _ bool,
) (bool, string, error) {
	__antithesis_instrumentation__.Notify(491488)
	return false, "", errors.WithStack(errEvalSessionVar)
}

func (ep *DummySessionAccessor) SetSessionVar(
	ctx context.Context, settingName, newValue string, isLocal bool,
) error {
	__antithesis_instrumentation__.Notify(491489)
	return errors.WithStack(errEvalSessionVar)
}

func (ep *DummySessionAccessor) HasAdminRole(_ context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(491490)
	return false, errors.WithStack(errEvalSessionVar)
}

func (ep *DummySessionAccessor) HasRoleOption(
	ctx context.Context, roleOption roleoption.Option,
) (bool, error) {
	__antithesis_instrumentation__.Notify(491491)
	return false, errors.WithStack(errEvalSessionVar)
}

type DummyClientNoticeSender struct{}

var _ tree.ClientNoticeSender = &DummyClientNoticeSender{}

func (c *DummyClientNoticeSender) BufferClientNotice(context.Context, pgnotice.Notice) {
	__antithesis_instrumentation__.Notify(491492)
}

type DummyTenantOperator struct{}

var _ tree.TenantOperator = &DummyTenantOperator{}

var errEvalTenant = pgerror.New(pgcode.ScalarOperationCannotRunWithoutFullSessionContext,
	"cannot evaluate tenant operation in this context")

func (c *DummyTenantOperator) CreateTenant(_ context.Context, _ uint64) error {
	__antithesis_instrumentation__.Notify(491493)
	return errors.WithStack(errEvalTenant)
}

func (c *DummyTenantOperator) DestroyTenant(
	ctx context.Context, tenantID uint64, synchronous bool,
) error {
	__antithesis_instrumentation__.Notify(491494)
	return errors.WithStack(errEvalTenant)
}

func (c *DummyTenantOperator) GCTenant(_ context.Context, _ uint64) error {
	__antithesis_instrumentation__.Notify(491495)
	return errors.WithStack(errEvalTenant)
}

func (c *DummyTenantOperator) UpdateTenantResourceLimits(
	_ context.Context,
	tenantID uint64,
	availableRU float64,
	refillRate float64,
	maxBurstRU float64,
	asOf time.Time,
	asOfConsumedRequestUnits float64,
) error {
	__antithesis_instrumentation__.Notify(491496)
	return errors.WithStack(errEvalTenant)
}

type DummyPreparedStatementState struct{}

var _ tree.PreparedStatementState = (*DummyPreparedStatementState)(nil)

func (ps *DummyPreparedStatementState) HasActivePortals() bool {
	__antithesis_instrumentation__.Notify(491497)
	return false
}

func (ps *DummyPreparedStatementState) MigratablePreparedStatements() []sessiondatapb.MigratableSession_PreparedStatement {
	__antithesis_instrumentation__.Notify(491498)
	return nil
}

func (ps *DummyPreparedStatementState) HasPortal(_ string) bool {
	__antithesis_instrumentation__.Notify(491499)
	return false
}
