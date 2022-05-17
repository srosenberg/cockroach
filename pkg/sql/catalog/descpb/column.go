package descpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/errors"
)

func (desc *ColumnDescriptor) HasNullDefault() bool {
	__antithesis_instrumentation__.Notify(251478)
	if !desc.HasDefault() {
		__antithesis_instrumentation__.Notify(251481)
		return false
	} else {
		__antithesis_instrumentation__.Notify(251482)
	}
	__antithesis_instrumentation__.Notify(251479)
	defaultExpr, err := parser.ParseExpr(*desc.DefaultExpr)
	if err != nil {
		__antithesis_instrumentation__.Notify(251483)
		panic(errors.NewAssertionErrorWithWrappedErrf(err,
			"failed to parse default expression %s", *desc.DefaultExpr))
	} else {
		__antithesis_instrumentation__.Notify(251484)
	}
	__antithesis_instrumentation__.Notify(251480)
	return defaultExpr == tree.DNull
}

func (desc *ColumnDescriptor) HasDefault() bool {
	__antithesis_instrumentation__.Notify(251485)
	return desc.DefaultExpr != nil
}

func (desc *ColumnDescriptor) HasOnUpdate() bool {
	__antithesis_instrumentation__.Notify(251486)
	return desc.OnUpdateExpr != nil
}

func (desc *ColumnDescriptor) IsComputed() bool {
	__antithesis_instrumentation__.Notify(251487)
	return desc.ComputeExpr != nil
}

func (desc *ColumnDescriptor) ColName() tree.Name {
	__antithesis_instrumentation__.Notify(251488)
	return tree.Name(desc.Name)
}

func (desc *ColumnDescriptor) CheckCanBeOutboundFKRef() error {
	__antithesis_instrumentation__.Notify(251489)
	if desc.Inaccessible {
		__antithesis_instrumentation__.Notify(251493)
		return pgerror.Newf(
			pgcode.UndefinedColumn,
			"column %q is inaccessible and cannot reference a foreign key",
			desc.Name,
		)
	} else {
		__antithesis_instrumentation__.Notify(251494)
	}
	__antithesis_instrumentation__.Notify(251490)
	if desc.Virtual {
		__antithesis_instrumentation__.Notify(251495)
		return unimplemented.NewWithIssuef(
			59671, "virtual column %q cannot reference a foreign key",
			desc.Name,
		)
	} else {
		__antithesis_instrumentation__.Notify(251496)
	}
	__antithesis_instrumentation__.Notify(251491)
	if desc.IsComputed() {
		__antithesis_instrumentation__.Notify(251497)
		return unimplemented.NewWithIssuef(
			46672, "computed column %q cannot reference a foreign key",
			desc.Name,
		)
	} else {
		__antithesis_instrumentation__.Notify(251498)
	}
	__antithesis_instrumentation__.Notify(251492)
	return nil
}

func (desc *ColumnDescriptor) CheckCanBeInboundFKRef() error {
	__antithesis_instrumentation__.Notify(251499)
	if desc.Inaccessible {
		__antithesis_instrumentation__.Notify(251502)
		return pgerror.Newf(
			pgcode.UndefinedColumn,
			"column %q is inaccessible and cannot be referenced by a foreign key",
			desc.Name,
		)
	} else {
		__antithesis_instrumentation__.Notify(251503)
	}
	__antithesis_instrumentation__.Notify(251500)
	if desc.Virtual {
		__antithesis_instrumentation__.Notify(251504)
		return unimplemented.NewWithIssuef(
			59671, "virtual column %q cannot be referenced by a foreign key",
			desc.Name,
		)
	} else {
		__antithesis_instrumentation__.Notify(251505)
	}
	__antithesis_instrumentation__.Notify(251501)
	return nil
}

func (desc ColumnDescriptor) GetPGAttributeNum() uint32 {
	__antithesis_instrumentation__.Notify(251506)
	if desc.PGAttributeNum != 0 {
		__antithesis_instrumentation__.Notify(251508)
		return desc.PGAttributeNum
	} else {
		__antithesis_instrumentation__.Notify(251509)
	}
	__antithesis_instrumentation__.Notify(251507)

	return uint32(desc.ID)
}

func (desc *ColumnDescriptor) SQLStringNotHumanReadable() string {
	__antithesis_instrumentation__.Notify(251510)
	f := tree.NewFmtCtx(tree.FmtSimple)
	f.FormatNameP(&desc.Name)
	f.WriteByte(' ')
	f.WriteString(desc.Type.SQLString())
	if desc.Nullable {
		__antithesis_instrumentation__.Notify(251514)
		f.WriteString(" NULL")
	} else {
		__antithesis_instrumentation__.Notify(251515)
		f.WriteString(" NOT NULL")
	}
	__antithesis_instrumentation__.Notify(251511)
	if desc.DefaultExpr != nil {
		__antithesis_instrumentation__.Notify(251516)
		f.WriteString(" DEFAULT ")
		f.WriteString(*desc.DefaultExpr)
	} else {
		__antithesis_instrumentation__.Notify(251517)
	}
	__antithesis_instrumentation__.Notify(251512)
	if desc.IsComputed() {
		__antithesis_instrumentation__.Notify(251518)
		f.WriteString(" AS (")
		f.WriteString(*desc.ComputeExpr)
		f.WriteString(") STORED")
	} else {
		__antithesis_instrumentation__.Notify(251519)
	}
	__antithesis_instrumentation__.Notify(251513)
	return f.CloseAndGetString()
}
