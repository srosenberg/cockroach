package screl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"reflect"

	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/rel"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

func NodeString(n *Node) string {
	__antithesis_instrumentation__.Notify(595001)
	var v redact.StringBuilder
	if err := FormatNode(&v, n); err != nil {
		__antithesis_instrumentation__.Notify(595003)
		return fmt.Sprintf("failed for format node: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(595004)
	}
	__antithesis_instrumentation__.Notify(595002)
	return v.String()
}

func FormatNode(w redact.SafeWriter, e *Node) (err error) {
	__antithesis_instrumentation__.Notify(595005)
	w.SafeString("[[")
	if err := FormatElement(w, e.Element()); err != nil {
		__antithesis_instrumentation__.Notify(595007)
		return err
	} else {
		__antithesis_instrumentation__.Notify(595008)
	}
	__antithesis_instrumentation__.Notify(595006)
	w.SafeString(", ")
	w.SafeString(redact.SafeString(e.Target.TargetStatus.String()))
	w.SafeString("], ")
	w.SafeString(redact.SafeString(e.CurrentStatus.String()))
	w.SafeString("]")
	return nil
}

func ElementString(e scpb.Element) string {
	__antithesis_instrumentation__.Notify(595009)
	var v redact.StringBuilder
	if err := FormatElement(&v, e); err != nil {
		__antithesis_instrumentation__.Notify(595011)
		return fmt.Sprintf("failed for format element %T: %v", e, err)
	} else {
		__antithesis_instrumentation__.Notify(595012)
	}
	__antithesis_instrumentation__.Notify(595010)
	return v.String()
}

func FormatElement(w redact.SafeWriter, e scpb.Element) (err error) {
	__antithesis_instrumentation__.Notify(595013)
	if e == nil {
		__antithesis_instrumentation__.Notify(595016)
		return errors.Errorf("nil element")
	} else {
		__antithesis_instrumentation__.Notify(595017)
	}
	__antithesis_instrumentation__.Notify(595014)
	w.SafeString(redact.SafeString(reflect.TypeOf(e).Elem().Name()))
	w.SafeString(":{")
	var written int
	if err := Schema.IterateAttributes(e, func(attr rel.Attr, value interface{}) error {
		__antithesis_instrumentation__.Notify(595018)
		if written > 0 {
			__antithesis_instrumentation__.Notify(595021)
			w.SafeString(", ")
		} else {
			__antithesis_instrumentation__.Notify(595022)
		}
		__antithesis_instrumentation__.Notify(595019)
		written++

		if str, isStr := value.(string); isStr {
			__antithesis_instrumentation__.Notify(595023)
			value = tree.Name(str)
		} else {
			__antithesis_instrumentation__.Notify(595024)
		}
		__antithesis_instrumentation__.Notify(595020)
		w.SafeString(redact.SafeString(attr.String()))
		w.SafeString(": ")
		w.Printf("%v", value)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(595025)
		return err
	} else {
		__antithesis_instrumentation__.Notify(595026)
	}
	__antithesis_instrumentation__.Notify(595015)
	w.SafeRune('}')
	return nil
}
