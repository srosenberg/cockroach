package querycache

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/lib/pq/oid"
)

type PrepareMetadata struct {
	parser.Statement

	StatementNoConstants string

	tree.PlaceholderTypesInfo

	Columns colinfo.ResultColumns

	InferredTypes []oid.Oid
}

func (pm *PrepareMetadata) MemoryEstimate() int64 {
	__antithesis_instrumentation__.Notify(563686)
	res := int64(unsafe.Sizeof(*pm))
	res += int64(len(pm.SQL))

	res += 2 * int64(len(pm.SQL))

	res += int64(len(pm.StatementNoConstants))

	res += int64(len(pm.TypeHints)+len(pm.Types)) *
		int64(unsafe.Sizeof(tree.PlaceholderIdx(0))+unsafe.Sizeof((*types.T)(nil)))

	res += int64(len(pm.Columns)) * int64(unsafe.Sizeof(colinfo.ResultColumn{}))
	res += int64(len(pm.InferredTypes)) * int64(unsafe.Sizeof(oid.Oid(0)))

	return res
}
