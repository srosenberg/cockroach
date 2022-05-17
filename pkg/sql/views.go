package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
)

type planDependencyInfo struct {
	desc catalog.TableDescriptor

	deps []descpb.TableDescriptor_Reference
}

type planDependencies map[descpb.ID]planDependencyInfo

func (d planDependencies) String() string {
	__antithesis_instrumentation__.Notify(632399)
	var buf bytes.Buffer
	for id, deps := range d {
		__antithesis_instrumentation__.Notify(632401)
		name := deps.desc.GetName()
		fmt.Fprintf(&buf, "%d (%q):", id, tree.ErrNameStringP(&name))
		for _, dep := range deps.deps {
			__antithesis_instrumentation__.Notify(632403)
			buf.WriteString(" [")
			if dep.IndexID != 0 {
				__antithesis_instrumentation__.Notify(632405)
				fmt.Fprintf(&buf, "idx: %d ", dep.IndexID)
			} else {
				__antithesis_instrumentation__.Notify(632406)
			}
			__antithesis_instrumentation__.Notify(632404)
			fmt.Fprintf(&buf, "cols: %v]", dep.ColumnIDs)
		}
		__antithesis_instrumentation__.Notify(632402)
		buf.WriteByte('\n')
	}
	__antithesis_instrumentation__.Notify(632400)
	return buf.String()
}

type typeDependencies map[descpb.ID]struct{}

func checkViewMatchesMaterialized(
	desc catalog.TableDescriptor, requireView, wantMaterialized bool,
) error {
	__antithesis_instrumentation__.Notify(632407)
	if !requireView {
		__antithesis_instrumentation__.Notify(632412)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(632413)
	}
	__antithesis_instrumentation__.Notify(632408)
	if !desc.IsView() {
		__antithesis_instrumentation__.Notify(632414)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(632415)
	}
	__antithesis_instrumentation__.Notify(632409)
	isMaterialized := desc.MaterializedView()
	if isMaterialized && func() bool {
		__antithesis_instrumentation__.Notify(632416)
		return !wantMaterialized == true
	}() == true {
		__antithesis_instrumentation__.Notify(632417)
		err := pgerror.Newf(pgcode.WrongObjectType, "%q is a materialized view", desc.GetName())
		return errors.WithHint(err, "use the corresponding MATERIALIZED VIEW command")
	} else {
		__antithesis_instrumentation__.Notify(632418)
	}
	__antithesis_instrumentation__.Notify(632410)
	if !isMaterialized && func() bool {
		__antithesis_instrumentation__.Notify(632419)
		return wantMaterialized == true
	}() == true {
		__antithesis_instrumentation__.Notify(632420)
		return pgerror.Newf(pgcode.WrongObjectType, "%q is not a materialized view", desc.GetName())
	} else {
		__antithesis_instrumentation__.Notify(632421)
	}
	__antithesis_instrumentation__.Notify(632411)
	return nil
}
