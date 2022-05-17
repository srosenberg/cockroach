package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/errors"
)

type alterSequenceNode struct {
	n       *tree.AlterSequence
	seqDesc *tabledesc.Mutable
}

func (p *planner) AlterSequence(ctx context.Context, n *tree.AlterSequence) (planNode, error) {
	__antithesis_instrumentation__.Notify(243861)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"ALTER SEQUENCE",
	); err != nil {
		__antithesis_instrumentation__.Notify(243866)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(243867)
	}
	__antithesis_instrumentation__.Notify(243862)

	_, seqDesc, err := p.ResolveMutableTableDescriptorEx(
		ctx, n.Name, !n.IfExists, tree.ResolveRequireSequenceDesc,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(243868)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(243869)
	}
	__antithesis_instrumentation__.Notify(243863)
	if seqDesc == nil {
		__antithesis_instrumentation__.Notify(243870)
		return newZeroNode(nil), nil
	} else {
		__antithesis_instrumentation__.Notify(243871)
	}
	__antithesis_instrumentation__.Notify(243864)

	if err := p.CheckPrivilege(ctx, seqDesc, privilege.CREATE); err != nil {
		__antithesis_instrumentation__.Notify(243872)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(243873)
	}
	__antithesis_instrumentation__.Notify(243865)

	return &alterSequenceNode{n: n, seqDesc: seqDesc}, nil
}

func (n *alterSequenceNode) ReadingOwnWrites() { __antithesis_instrumentation__.Notify(243874) }

func (n *alterSequenceNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(243875)
	telemetry.Inc(sqltelemetry.SchemaChangeAlterCounter("sequence"))
	desc := n.seqDesc

	oldMinValue := desc.SequenceOpts.MinValue
	oldMaxValue := desc.SequenceOpts.MaxValue

	existingType := types.Int
	if desc.GetSequenceOpts().AsIntegerType != "" {
		__antithesis_instrumentation__.Notify(243881)
		switch desc.GetSequenceOpts().AsIntegerType {
		case types.Int2.SQLString():
			__antithesis_instrumentation__.Notify(243882)
			existingType = types.Int2
		case types.Int4.SQLString():
			__antithesis_instrumentation__.Notify(243883)
			existingType = types.Int4
		case types.Int.SQLString():
			__antithesis_instrumentation__.Notify(243884)

		default:
			__antithesis_instrumentation__.Notify(243885)
			return errors.AssertionFailedf("sequence has unexpected type %s", desc.GetSequenceOpts().AsIntegerType)
		}
	} else {
		__antithesis_instrumentation__.Notify(243886)
	}
	__antithesis_instrumentation__.Notify(243876)
	if err := assignSequenceOptions(
		params.ctx,
		params.p,
		desc.SequenceOpts,
		n.n.Options,
		false,
		desc.GetID(),
		desc.ParentID,
		existingType,
	); err != nil {
		__antithesis_instrumentation__.Notify(243887)
		return err
	} else {
		__antithesis_instrumentation__.Notify(243888)
	}
	__antithesis_instrumentation__.Notify(243877)
	opts := desc.SequenceOpts
	seqValueKey := params.p.ExecCfg().Codec.SequenceKey(uint32(desc.ID))

	getSequenceValue := func() (int64, error) {
		__antithesis_instrumentation__.Notify(243889)
		kv, err := params.p.txn.Get(params.ctx, seqValueKey)
		if err != nil {
			__antithesis_instrumentation__.Notify(243891)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(243892)
		}
		__antithesis_instrumentation__.Notify(243890)
		return kv.ValueInt(), nil
	}
	__antithesis_instrumentation__.Notify(243878)

	if opts.Increment < 0 && func() bool {
		__antithesis_instrumentation__.Notify(243893)
		return desc.SequenceOpts.MinValue < oldMinValue == true
	}() == true {
		__antithesis_instrumentation__.Notify(243894)
		sequenceVal, err := getSequenceValue()
		if err != nil {
			__antithesis_instrumentation__.Notify(243896)
			return err
		} else {
			__antithesis_instrumentation__.Notify(243897)
		}
		__antithesis_instrumentation__.Notify(243895)

		if sequenceVal < oldMinValue {
			__antithesis_instrumentation__.Notify(243898)
			err := params.p.txn.Put(params.ctx, seqValueKey, oldMinValue)
			if err != nil {
				__antithesis_instrumentation__.Notify(243899)
				return err
			} else {
				__antithesis_instrumentation__.Notify(243900)
			}
		} else {
			__antithesis_instrumentation__.Notify(243901)
		}
	} else {
		__antithesis_instrumentation__.Notify(243902)
		if opts.Increment > 0 && func() bool {
			__antithesis_instrumentation__.Notify(243903)
			return desc.SequenceOpts.MaxValue > oldMaxValue == true
		}() == true {
			__antithesis_instrumentation__.Notify(243904)
			sequenceVal, err := getSequenceValue()
			if err != nil {
				__antithesis_instrumentation__.Notify(243906)
				return err
			} else {
				__antithesis_instrumentation__.Notify(243907)
			}
			__antithesis_instrumentation__.Notify(243905)

			if sequenceVal > oldMaxValue {
				__antithesis_instrumentation__.Notify(243908)
				err := params.p.txn.Put(params.ctx, seqValueKey, oldMaxValue)
				if err != nil {
					__antithesis_instrumentation__.Notify(243909)
					return err
				} else {
					__antithesis_instrumentation__.Notify(243910)
				}
			} else {
				__antithesis_instrumentation__.Notify(243911)
			}
		} else {
			__antithesis_instrumentation__.Notify(243912)
		}
	}
	__antithesis_instrumentation__.Notify(243879)

	if err := params.p.writeSchemaChange(
		params.ctx, n.seqDesc, descpb.InvalidMutationID, tree.AsStringWithFQNames(n.n, params.Ann()),
	); err != nil {
		__antithesis_instrumentation__.Notify(243913)
		return err
	} else {
		__antithesis_instrumentation__.Notify(243914)
	}
	__antithesis_instrumentation__.Notify(243880)

	return params.p.logEvent(params.ctx,
		n.seqDesc.ID,
		&eventpb.AlterSequence{
			SequenceName: params.p.ResolvedName(n.n.Name).FQString(),
		})
}

func (n *alterSequenceNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(243915)
	return false, nil
}
func (n *alterSequenceNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(243916)
	return tree.Datums{}
}
func (n *alterSequenceNode) Close(context.Context) { __antithesis_instrumentation__.Notify(243917) }
