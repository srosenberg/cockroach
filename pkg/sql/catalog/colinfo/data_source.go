package colinfo

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

type DataSourceInfo struct {
	SourceColumns ResultColumns

	SourceAlias tree.TableName
}

func (src *DataSourceInfo) String() string {
	__antithesis_instrumentation__.Notify(251125)
	var buf bytes.Buffer
	for i := range src.SourceColumns {
		__antithesis_instrumentation__.Notify(251129)
		if i > 0 {
			__antithesis_instrumentation__.Notify(251131)
			buf.WriteByte('\t')
		} else {
			__antithesis_instrumentation__.Notify(251132)
		}
		__antithesis_instrumentation__.Notify(251130)
		fmt.Fprintf(&buf, "%d", i)
	}
	__antithesis_instrumentation__.Notify(251126)
	buf.WriteString("\toutput column positions\n")
	for i, c := range src.SourceColumns {
		__antithesis_instrumentation__.Notify(251133)
		if i > 0 {
			__antithesis_instrumentation__.Notify(251136)
			buf.WriteByte('\t')
		} else {
			__antithesis_instrumentation__.Notify(251137)
		}
		__antithesis_instrumentation__.Notify(251134)
		if c.Hidden {
			__antithesis_instrumentation__.Notify(251138)
			buf.WriteByte('*')
		} else {
			__antithesis_instrumentation__.Notify(251139)
		}
		__antithesis_instrumentation__.Notify(251135)
		buf.WriteString(c.Name)
	}
	__antithesis_instrumentation__.Notify(251127)
	buf.WriteString("\toutput column names\n")
	if src.SourceAlias == descpb.AnonymousTable {
		__antithesis_instrumentation__.Notify(251140)
		buf.WriteString("\t<anonymous table>\n")
	} else {
		__antithesis_instrumentation__.Notify(251141)
		fmt.Fprintf(&buf, "\t'%s'\n", src.SourceAlias.String())
	}
	__antithesis_instrumentation__.Notify(251128)
	return buf.String()
}

func NewSourceInfoForSingleTable(tn tree.TableName, columns ResultColumns) *DataSourceInfo {
	__antithesis_instrumentation__.Notify(251142)
	if tn.ObjectName != "" && func() bool {
		__antithesis_instrumentation__.Notify(251144)
		return tn.SchemaName != "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(251145)

		tn.ExplicitCatalog = true
		tn.ExplicitSchema = true
	} else {
		__antithesis_instrumentation__.Notify(251146)
	}
	__antithesis_instrumentation__.Notify(251143)
	return &DataSourceInfo{
		SourceColumns: columns,
		SourceAlias:   tn,
	}
}

type varFormatter struct {
	TableName  tree.TableName
	ColumnName tree.Name
}

func (c *varFormatter) Format(ctx *tree.FmtCtx) {
	__antithesis_instrumentation__.Notify(251147)
	if ctx.HasFlags(tree.FmtShowTableAliases) && func() bool {
		__antithesis_instrumentation__.Notify(251149)
		return c.TableName.ObjectName != "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(251150)

		if c.TableName.SchemaName != "" {
			__antithesis_instrumentation__.Notify(251152)
			if c.TableName.CatalogName != "" {
				__antithesis_instrumentation__.Notify(251154)
				ctx.FormatNode(&c.TableName.CatalogName)
				ctx.WriteByte('.')
			} else {
				__antithesis_instrumentation__.Notify(251155)
			}
			__antithesis_instrumentation__.Notify(251153)
			ctx.FormatNode(&c.TableName.SchemaName)
			ctx.WriteByte('.')
		} else {
			__antithesis_instrumentation__.Notify(251156)
		}
		__antithesis_instrumentation__.Notify(251151)

		ctx.FormatNode(&c.TableName.ObjectName)
		ctx.WriteByte('.')
	} else {
		__antithesis_instrumentation__.Notify(251157)
	}
	__antithesis_instrumentation__.Notify(251148)
	ctx.FormatNode(&c.ColumnName)
}

func (src *DataSourceInfo) NodeFormatter(colIdx int) tree.NodeFormatter {
	__antithesis_instrumentation__.Notify(251158)
	return &varFormatter{
		TableName:  src.SourceAlias,
		ColumnName: tree.Name(src.SourceColumns[colIdx].Name),
	}
}
