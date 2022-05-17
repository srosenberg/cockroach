package typedesc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/redact"
)

func (desc *immutable) SafeMessage() string {
	__antithesis_instrumentation__.Notify(271836)
	return formatSafeType("typedesc.immutable", desc)
}

func (desc *Mutable) SafeMessage() string {
	__antithesis_instrumentation__.Notify(271837)
	return formatSafeType("typedesc.Mutable", desc)
}

func formatSafeType(typeName string, desc catalog.TypeDescriptor) string {
	__antithesis_instrumentation__.Notify(271838)
	var buf redact.StringBuilder
	buf.Printf(typeName + ": {")
	formatSafeTypeProperties(&buf, desc)
	buf.Printf("}")
	return buf.String()
}

func formatSafeTypeProperties(w *redact.StringBuilder, desc catalog.TypeDescriptor) {
	__antithesis_instrumentation__.Notify(271839)
	catalog.FormatSafeDescriptorProperties(w, desc)
	td := desc.TypeDesc()
	w.Printf(", Kind: %s", td.Kind)
	if len(td.EnumMembers) > 0 {
		__antithesis_instrumentation__.Notify(271844)
		w.Printf(", NumEnumMembers: %d", len(td.EnumMembers))
	} else {
		__antithesis_instrumentation__.Notify(271845)
	}
	__antithesis_instrumentation__.Notify(271840)
	if td.Alias != nil {
		__antithesis_instrumentation__.Notify(271846)
		w.Printf(", Alias: %d", td.Alias.Oid())
	} else {
		__antithesis_instrumentation__.Notify(271847)
	}
	__antithesis_instrumentation__.Notify(271841)
	if td.ArrayTypeID != 0 {
		__antithesis_instrumentation__.Notify(271848)
		w.Printf(", ArrayTypeID: %d", td.ArrayTypeID)
	} else {
		__antithesis_instrumentation__.Notify(271849)
	}
	__antithesis_instrumentation__.Notify(271842)
	for i := range td.ReferencingDescriptorIDs {
		__antithesis_instrumentation__.Notify(271850)
		w.Printf(", ")
		if i == 0 {
			__antithesis_instrumentation__.Notify(271852)
			w.Printf("ReferencingDescriptorIDs: [")
		} else {
			__antithesis_instrumentation__.Notify(271853)
		}
		__antithesis_instrumentation__.Notify(271851)
		w.Printf("%d", td.ReferencingDescriptorIDs[i])
	}
	__antithesis_instrumentation__.Notify(271843)
	if len(td.ReferencingDescriptorIDs) > 0 {
		__antithesis_instrumentation__.Notify(271854)
		w.Printf("]")
	} else {
		__antithesis_instrumentation__.Notify(271855)
	}
}
