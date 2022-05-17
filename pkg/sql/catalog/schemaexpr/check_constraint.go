package schemaexpr

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

type CheckConstraintBuilder struct {
	ctx       context.Context
	tableName tree.TableName
	desc      catalog.TableDescriptor
	semaCtx   *tree.SemaContext

	inUseNames map[string]struct{}
}

func MakeCheckConstraintBuilder(
	ctx context.Context,
	tableName tree.TableName,
	desc catalog.TableDescriptor,
	semaCtx *tree.SemaContext,
) CheckConstraintBuilder {
	__antithesis_instrumentation__.Notify(267780)
	return CheckConstraintBuilder{
		ctx:        ctx,
		tableName:  tableName,
		desc:       desc,
		semaCtx:    semaCtx,
		inUseNames: make(map[string]struct{}),
	}
}

func (b *CheckConstraintBuilder) MarkNameInUse(name string) {
	__antithesis_instrumentation__.Notify(267781)
	b.inUseNames[name] = struct{}{}
}

func (b *CheckConstraintBuilder) Build(
	c *tree.CheckConstraintTableDef,
) (*descpb.TableDescriptor_CheckConstraint, error) {
	__antithesis_instrumentation__.Notify(267782)
	name := string(c.Name)

	if name == "" {
		__antithesis_instrumentation__.Notify(267785)
		var err error
		name, err = b.generateUniqueName(c.Expr)
		if err != nil {
			__antithesis_instrumentation__.Notify(267786)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(267787)
		}
	} else {
		__antithesis_instrumentation__.Notify(267788)
	}
	__antithesis_instrumentation__.Notify(267783)

	expr, _, colIDs, err := DequalifyAndValidateExpr(
		b.ctx,
		b.desc,
		c.Expr,
		types.Bool,
		"CHECK",
		b.semaCtx,
		tree.VolatilityVolatile,
		&b.tableName,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(267789)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(267790)
	}
	__antithesis_instrumentation__.Notify(267784)
	constraintID := b.desc.TableDesc().GetNextConstraintID()
	b.desc.TableDesc().NextConstraintID++
	return &descpb.TableDescriptor_CheckConstraint{
		Expr:         expr,
		Name:         name,
		ColumnIDs:    colIDs.Ordered(),
		Hidden:       c.Hidden,
		ConstraintID: constraintID,
	}, nil
}

func (b *CheckConstraintBuilder) nameInUse(name string) bool {
	__antithesis_instrumentation__.Notify(267791)
	_, ok := b.inUseNames[name]
	return ok
}

func (b *CheckConstraintBuilder) generateUniqueName(expr tree.Expr) (string, error) {
	__antithesis_instrumentation__.Notify(267792)
	name, err := b.DefaultName(expr)
	if err != nil {
		__antithesis_instrumentation__.Notify(267795)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(267796)
	}
	__antithesis_instrumentation__.Notify(267793)

	if b.nameInUse(name) {
		__antithesis_instrumentation__.Notify(267797)
		i := 1
		for {
			__antithesis_instrumentation__.Notify(267798)
			numberedName := fmt.Sprintf("%s%d", name, i)
			if !b.nameInUse(numberedName) {
				__antithesis_instrumentation__.Notify(267800)
				name = numberedName
				break
			} else {
				__antithesis_instrumentation__.Notify(267801)
			}
			__antithesis_instrumentation__.Notify(267799)
			i++
		}
	} else {
		__antithesis_instrumentation__.Notify(267802)
	}
	__antithesis_instrumentation__.Notify(267794)

	b.MarkNameInUse(name)

	return name, nil
}

func (b *CheckConstraintBuilder) DefaultName(expr tree.Expr) (string, error) {
	__antithesis_instrumentation__.Notify(267803)
	var nameBuf bytes.Buffer
	nameBuf.WriteString("check")

	err := iterColDescriptors(b.desc, expr, func(c catalog.Column) error {
		__antithesis_instrumentation__.Notify(267806)
		nameBuf.WriteByte('_')
		nameBuf.WriteString(c.GetName())
		return nil
	})
	__antithesis_instrumentation__.Notify(267804)
	if err != nil {
		__antithesis_instrumentation__.Notify(267807)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(267808)
	}
	__antithesis_instrumentation__.Notify(267805)

	return nameBuf.String(), nil
}
