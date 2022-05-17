package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/roleoption"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/errors"
)

type createDatabaseNode struct {
	n *tree.CreateDatabase
}

func (p *planner) CreateDatabase(ctx context.Context, n *tree.CreateDatabase) (planNode, error) {
	__antithesis_instrumentation__.Notify(462846)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"CREATE DATABASE",
	); err != nil {
		__antithesis_instrumentation__.Notify(462858)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(462859)
	}
	__antithesis_instrumentation__.Notify(462847)

	if n.Name == "" {
		__antithesis_instrumentation__.Notify(462860)
		return nil, errEmptyDatabaseName
	} else {
		__antithesis_instrumentation__.Notify(462861)
	}
	__antithesis_instrumentation__.Notify(462848)

	if tmpl := n.Template; tmpl != "" {
		__antithesis_instrumentation__.Notify(462862)

		if !strings.EqualFold(tmpl, "template0") {
			__antithesis_instrumentation__.Notify(462863)
			return nil, unimplemented.NewWithIssuef(10151,
				"unsupported template: %s", tmpl)
		} else {
			__antithesis_instrumentation__.Notify(462864)
		}
	} else {
		__antithesis_instrumentation__.Notify(462865)
	}
	__antithesis_instrumentation__.Notify(462849)

	if enc := n.Encoding; enc != "" {
		__antithesis_instrumentation__.Notify(462866)

		if !(strings.EqualFold(enc, "UTF8") || func() bool {
			__antithesis_instrumentation__.Notify(462867)
			return strings.EqualFold(enc, "UTF-8") == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(462868)
			return strings.EqualFold(enc, "UNICODE") == true
		}() == true) {
			__antithesis_instrumentation__.Notify(462869)
			return nil, unimplemented.NewWithIssueDetailf(35882, "create.db.encoding",
				"unsupported encoding: %s", enc)
		} else {
			__antithesis_instrumentation__.Notify(462870)
		}
	} else {
		__antithesis_instrumentation__.Notify(462871)
	}
	__antithesis_instrumentation__.Notify(462850)

	if col := n.Collate; col != "" {
		__antithesis_instrumentation__.Notify(462872)

		if col != "C" && func() bool {
			__antithesis_instrumentation__.Notify(462873)
			return col != "C.UTF-8" == true
		}() == true {
			__antithesis_instrumentation__.Notify(462874)
			return nil, unimplemented.NewWithIssueDetailf(16618, "create.db.collation",
				"unsupported collation: %s", col)
		} else {
			__antithesis_instrumentation__.Notify(462875)
		}
	} else {
		__antithesis_instrumentation__.Notify(462876)
	}
	__antithesis_instrumentation__.Notify(462851)

	if ctype := n.CType; ctype != "" {
		__antithesis_instrumentation__.Notify(462877)

		if ctype != "C" && func() bool {
			__antithesis_instrumentation__.Notify(462878)
			return ctype != "C.UTF-8" == true
		}() == true {
			__antithesis_instrumentation__.Notify(462879)
			return nil, unimplemented.NewWithIssueDetailf(35882, "create.db.classification",
				"unsupported character classification: %s", ctype)
		} else {
			__antithesis_instrumentation__.Notify(462880)
		}
	} else {
		__antithesis_instrumentation__.Notify(462881)
	}
	__antithesis_instrumentation__.Notify(462852)

	if n.ConnectionLimit != -1 {
		__antithesis_instrumentation__.Notify(462882)
		return nil, unimplemented.NewWithIssueDetailf(
			54241,
			"create.db.connection_limit",
			"only connection limit -1 is supported, got: %d",
			n.ConnectionLimit,
		)
	} else {
		__antithesis_instrumentation__.Notify(462883)
	}
	__antithesis_instrumentation__.Notify(462853)

	if n.SurvivalGoal != tree.SurvivalGoalDefault && func() bool {
		__antithesis_instrumentation__.Notify(462884)
		return n.PrimaryRegion == tree.PrimaryRegionNotSpecifiedName == true
	}() == true {
		__antithesis_instrumentation__.Notify(462885)
		return nil, pgerror.New(
			pgcode.InvalidDatabaseDefinition,
			"PRIMARY REGION must be specified when using SURVIVE",
		)
	} else {
		__antithesis_instrumentation__.Notify(462886)
	}
	__antithesis_instrumentation__.Notify(462854)

	if n.Placement != tree.DataPlacementUnspecified {
		__antithesis_instrumentation__.Notify(462887)
		if !p.EvalContext().SessionData().PlacementEnabled {
			__antithesis_instrumentation__.Notify(462890)
			return nil, errors.WithHint(pgerror.New(
				pgcode.FeatureNotSupported,
				"PLACEMENT requires that the session setting enable_multiregion_placement_policy "+
					"is enabled",
			),
				"to use PLACEMENT, enable the session setting with SET"+
					" enable_multiregion_placement_policy = true or enable the cluster setting"+
					" sql.defaults.multiregion_placement_policy.enabled",
			)
		} else {
			__antithesis_instrumentation__.Notify(462891)
		}
		__antithesis_instrumentation__.Notify(462888)

		if n.PrimaryRegion == tree.PrimaryRegionNotSpecifiedName {
			__antithesis_instrumentation__.Notify(462892)
			return nil, pgerror.New(
				pgcode.InvalidDatabaseDefinition,
				"PRIMARY REGION must be specified when using PLACEMENT",
			)
		} else {
			__antithesis_instrumentation__.Notify(462893)
		}
		__antithesis_instrumentation__.Notify(462889)

		if n.Placement == tree.DataPlacementRestricted && func() bool {
			__antithesis_instrumentation__.Notify(462894)
			return n.SurvivalGoal == tree.SurvivalGoalRegionFailure == true
		}() == true {
			__antithesis_instrumentation__.Notify(462895)
			return nil, pgerror.New(
				pgcode.InvalidDatabaseDefinition,
				"PLACEMENT RESTRICTED can only be used with SURVIVE ZONE FAILURE",
			)
		} else {
			__antithesis_instrumentation__.Notify(462896)
		}
	} else {
		__antithesis_instrumentation__.Notify(462897)
	}
	__antithesis_instrumentation__.Notify(462855)

	hasCreateDB, err := p.HasRoleOption(ctx, roleoption.CREATEDB)
	if err != nil {
		__antithesis_instrumentation__.Notify(462898)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(462899)
	}
	__antithesis_instrumentation__.Notify(462856)
	if !hasCreateDB {
		__antithesis_instrumentation__.Notify(462900)
		return nil, pgerror.New(
			pgcode.InsufficientPrivilege,
			"permission denied to create database",
		)
	} else {
		__antithesis_instrumentation__.Notify(462901)
	}
	__antithesis_instrumentation__.Notify(462857)

	return &createDatabaseNode{n: n}, nil
}

func (n *createDatabaseNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(462902)
	telemetry.Inc(sqltelemetry.SchemaChangeCreateCounter("database"))

	desc, created, err := params.p.createDatabase(
		params.ctx, n.n, tree.AsStringWithFQNames(n.n, params.Ann()))
	if err != nil {
		__antithesis_instrumentation__.Notify(462905)
		return err
	} else {
		__antithesis_instrumentation__.Notify(462906)
	}
	__antithesis_instrumentation__.Notify(462903)
	if created {
		__antithesis_instrumentation__.Notify(462907)

		if err := params.p.logEvent(params.ctx, desc.GetID(),
			&eventpb.CreateDatabase{
				DatabaseName: n.n.Name.String(),
			}); err != nil {
			__antithesis_instrumentation__.Notify(462908)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462909)
		}
	} else {
		__antithesis_instrumentation__.Notify(462910)
	}
	__antithesis_instrumentation__.Notify(462904)
	return nil
}

func (*createDatabaseNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(462911)
	return false, nil
}
func (*createDatabaseNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(462912)
	return tree.Datums{}
}
func (*createDatabaseNode) Close(context.Context) { __antithesis_instrumentation__.Notify(462913) }

func (*createDatabaseNode) ReadingOwnWrites() { __antithesis_instrumentation__.Notify(462914) }
