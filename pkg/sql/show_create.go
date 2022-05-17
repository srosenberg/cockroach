package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catformat"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemaexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
)

type shouldOmitFKClausesFromCreate int

const (
	_ shouldOmitFKClausesFromCreate = iota

	OmitFKClausesFromCreate

	IncludeFkClausesInCreate

	OmitMissingFKClausesFromCreate
)

type ShowCreateDisplayOptions struct {
	FKDisplayMode shouldOmitFKClausesFromCreate

	IgnoreComments bool
}

func ShowCreateTable(
	ctx context.Context,
	p PlanHookState,
	tn *tree.TableName,
	dbPrefix string,
	desc catalog.TableDescriptor,
	lCtx simpleSchemaResolver,
	displayOptions ShowCreateDisplayOptions,
) (string, error) {
	__antithesis_instrumentation__.Notify(622786)
	a := &tree.DatumAlloc{}

	f := p.ExtendedEvalContext().FmtCtx(tree.FmtSimple)
	f.WriteString("CREATE ")
	if desc.IsTemporary() {
		__antithesis_instrumentation__.Notify(622797)
		f.WriteString("TEMP ")
	} else {
		__antithesis_instrumentation__.Notify(622798)
	}
	__antithesis_instrumentation__.Notify(622787)
	f.WriteString("TABLE ")
	f.FormatNode(tn)
	f.WriteString(" (")

	for i, col := range desc.AccessibleColumns() {
		__antithesis_instrumentation__.Notify(622799)
		if i != 0 {
			__antithesis_instrumentation__.Notify(622802)
			f.WriteString(",")
		} else {
			__antithesis_instrumentation__.Notify(622803)
		}
		__antithesis_instrumentation__.Notify(622800)
		f.WriteString("\n\t")
		colstr, err := schemaexpr.FormatColumnForDisplay(
			ctx, desc, col, &p.RunParams(ctx).p.semaCtx, p.RunParams(ctx).p.SessionData(),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(622804)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(622805)
		}
		__antithesis_instrumentation__.Notify(622801)
		f.WriteString(colstr)
	}
	__antithesis_instrumentation__.Notify(622788)

	if desc.IsPhysicalTable() {
		__antithesis_instrumentation__.Notify(622806)
		f.WriteString(",\n\tCONSTRAINT ")
		formatQuoteNames(&f.Buffer, desc.GetPrimaryIndex().GetName())
		f.WriteString(" ")
		f.WriteString(tabledesc.PrimaryKeyString(desc))
	} else {
		__antithesis_instrumentation__.Notify(622807)
	}
	__antithesis_instrumentation__.Notify(622789)

	if displayOptions.FKDisplayMode != OmitFKClausesFromCreate {
		__antithesis_instrumentation__.Notify(622808)
		if err := desc.ForeachOutboundFK(func(fk *descpb.ForeignKeyConstraint) error {
			__antithesis_instrumentation__.Notify(622809)
			fkCtx := tree.NewFmtCtx(tree.FmtSimple)
			fkCtx.WriteString(",\n\tCONSTRAINT ")
			fkCtx.FormatNameP(&fk.Name)
			fkCtx.WriteString(" ")

			if err := showForeignKeyConstraint(
				&fkCtx.Buffer,
				dbPrefix,
				desc,
				fk,
				lCtx,
				sessiondata.EmptySearchPath,
			); err != nil {
				__antithesis_instrumentation__.Notify(622811)
				if displayOptions.FKDisplayMode == OmitMissingFKClausesFromCreate {
					__antithesis_instrumentation__.Notify(622813)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(622814)
				}
				__antithesis_instrumentation__.Notify(622812)

				return err
			} else {
				__antithesis_instrumentation__.Notify(622815)
			}
			__antithesis_instrumentation__.Notify(622810)
			f.WriteString(fkCtx.String())
			return nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(622816)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(622817)
		}
	} else {
		__antithesis_instrumentation__.Notify(622818)
	}
	__antithesis_instrumentation__.Notify(622790)
	for _, idx := range desc.PublicNonPrimaryIndexes() {
		__antithesis_instrumentation__.Notify(622819)

		var partitionBuf bytes.Buffer
		if err := ShowCreatePartitioning(
			a, p.ExecCfg().Codec, desc, idx, idx.GetPartitioning(), &partitionBuf, 1, 0,
		); err != nil {
			__antithesis_instrumentation__.Notify(622822)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(622823)
		}
		__antithesis_instrumentation__.Notify(622820)

		f.WriteString(",\n\t")
		idxStr, err := catformat.IndexForDisplay(
			ctx,
			desc,
			&descpb.AnonymousTable,
			idx,
			partitionBuf.String(),
			tree.FmtSimple,
			p.RunParams(ctx).p.SemaCtx(),
			p.RunParams(ctx).p.SessionData(),
			catformat.IndexDisplayDefOnly,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(622824)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(622825)
		}
		__antithesis_instrumentation__.Notify(622821)
		f.WriteString(idxStr)
	}
	__antithesis_instrumentation__.Notify(622791)

	showFamilyClause(desc, f)
	if err := showConstraintClause(ctx, desc, &p.RunParams(ctx).p.semaCtx, p.RunParams(ctx).p.SessionData(), f); err != nil {
		__antithesis_instrumentation__.Notify(622826)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(622827)
	}
	__antithesis_instrumentation__.Notify(622792)

	if err := ShowCreatePartitioning(
		a, p.ExecCfg().Codec, desc, desc.GetPrimaryIndex(), desc.GetPrimaryIndex().GetPartitioning(), &f.Buffer, 0, 0,
	); err != nil {
		__antithesis_instrumentation__.Notify(622828)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(622829)
	}
	__antithesis_instrumentation__.Notify(622793)

	if storageParams := desc.GetStorageParams(true); len(storageParams) > 0 {
		__antithesis_instrumentation__.Notify(622830)
		f.Buffer.WriteString(` WITH (`)
		f.Buffer.WriteString(strings.Join(storageParams, ", "))
		f.Buffer.WriteString(`)`)
	} else {
		__antithesis_instrumentation__.Notify(622831)
	}
	__antithesis_instrumentation__.Notify(622794)

	if err := showCreateLocality(desc, f); err != nil {
		__antithesis_instrumentation__.Notify(622832)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(622833)
	}
	__antithesis_instrumentation__.Notify(622795)

	if !displayOptions.IgnoreComments {
		__antithesis_instrumentation__.Notify(622834)
		if err := showComments(tn, desc, selectComment(ctx, p, desc.GetID()), &f.Buffer); err != nil {
			__antithesis_instrumentation__.Notify(622835)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(622836)
		}
	} else {
		__antithesis_instrumentation__.Notify(622837)
	}
	__antithesis_instrumentation__.Notify(622796)

	return f.CloseAndGetString(), nil
}

func formatQuoteNames(buf *bytes.Buffer, names ...string) {
	__antithesis_instrumentation__.Notify(622838)
	f := tree.NewFmtCtx(tree.FmtSimple)
	for i := range names {
		__antithesis_instrumentation__.Notify(622840)
		if i > 0 {
			__antithesis_instrumentation__.Notify(622842)
			f.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(622843)
		}
		__antithesis_instrumentation__.Notify(622841)
		f.FormatNameP(&names[i])
	}
	__antithesis_instrumentation__.Notify(622839)
	buf.WriteString(f.CloseAndGetString())
}

func (p *planner) ShowCreate(
	ctx context.Context,
	dbPrefix string,
	allDescs []descpb.Descriptor,
	desc catalog.TableDescriptor,
	displayOptions ShowCreateDisplayOptions,
) (string, error) {
	__antithesis_instrumentation__.Notify(622844)
	var stmt string
	var err error
	tn := tree.MakeUnqualifiedTableName(tree.Name(desc.GetName()))
	if desc.IsView() {
		__antithesis_instrumentation__.Notify(622846)
		stmt, err = ShowCreateView(ctx, &p.RunParams(ctx).p.semaCtx, p.RunParams(ctx).p.SessionData(), &tn, desc)
	} else {
		__antithesis_instrumentation__.Notify(622847)
		if desc.IsSequence() {
			__antithesis_instrumentation__.Notify(622848)
			stmt, err = ShowCreateSequence(ctx, &tn, desc)
		} else {
			__antithesis_instrumentation__.Notify(622849)
			lCtx, lErr := newInternalLookupCtxFromDescriptorProtos(
				ctx, allDescs, nil,
			)
			if lErr != nil {
				__antithesis_instrumentation__.Notify(622852)
				return "", lErr
			} else {
				__antithesis_instrumentation__.Notify(622853)
			}
			__antithesis_instrumentation__.Notify(622850)

			desc, err = lCtx.getTableByID(desc.GetID())
			if err != nil {
				__antithesis_instrumentation__.Notify(622854)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(622855)
			}
			__antithesis_instrumentation__.Notify(622851)
			stmt, err = ShowCreateTable(ctx, p, &tn, dbPrefix, desc, lCtx, displayOptions)
		}
	}
	__antithesis_instrumentation__.Notify(622845)

	return stmt, err
}
