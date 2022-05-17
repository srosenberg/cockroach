package kvnemesis

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
)

func (op Operation) Result() *Result {
	__antithesis_instrumentation__.Notify(90432)
	switch o := op.GetValue().(type) {
	case *GetOperation:
		__antithesis_instrumentation__.Notify(90433)
		return &o.Result
	case *PutOperation:
		__antithesis_instrumentation__.Notify(90434)
		return &o.Result
	case *ScanOperation:
		__antithesis_instrumentation__.Notify(90435)
		return &o.Result
	case *DeleteOperation:
		__antithesis_instrumentation__.Notify(90436)
		return &o.Result
	case *DeleteRangeOperation:
		__antithesis_instrumentation__.Notify(90437)
		return &o.Result
	case *SplitOperation:
		__antithesis_instrumentation__.Notify(90438)
		return &o.Result
	case *MergeOperation:
		__antithesis_instrumentation__.Notify(90439)
		return &o.Result
	case *ChangeReplicasOperation:
		__antithesis_instrumentation__.Notify(90440)
		return &o.Result
	case *TransferLeaseOperation:
		__antithesis_instrumentation__.Notify(90441)
		return &o.Result
	case *ChangeZoneOperation:
		__antithesis_instrumentation__.Notify(90442)
		return &o.Result
	case *BatchOperation:
		__antithesis_instrumentation__.Notify(90443)
		return &o.Result
	case *ClosureTxnOperation:
		__antithesis_instrumentation__.Notify(90444)
		return &o.Result
	default:
		__antithesis_instrumentation__.Notify(90445)
		panic(errors.AssertionFailedf(`unknown operation: %T %v`, o, o))
	}
}

type steps []Step

func (s steps) After() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(90446)
	var ts hlc.Timestamp
	for _, step := range s {
		__antithesis_instrumentation__.Notify(90448)
		ts.Forward(step.After)
	}
	__antithesis_instrumentation__.Notify(90447)
	return ts
}

func (s Step) String() string {
	__antithesis_instrumentation__.Notify(90449)
	var fctx formatCtx
	var buf strings.Builder
	s.format(&buf, fctx)
	return buf.String()
}

type formatCtx struct {
	receiver string
	indent   string
}

func (s Step) format(w *strings.Builder, fctx formatCtx) {
	__antithesis_instrumentation__.Notify(90450)
	if fctx.receiver != `` {
		__antithesis_instrumentation__.Notify(90452)
		panic(`cannot specify receiver in Step.format fctx`)
	} else {
		__antithesis_instrumentation__.Notify(90453)
	}
	__antithesis_instrumentation__.Notify(90451)
	fctx.receiver = fmt.Sprintf(`db%d`, s.DBID)
	w.WriteString("\n")
	w.WriteString(fctx.indent)
	s.Op.format(w, fctx)
}

func formatOps(w *strings.Builder, fctx formatCtx, ops []Operation) {
	__antithesis_instrumentation__.Notify(90454)
	for _, op := range ops {
		__antithesis_instrumentation__.Notify(90455)
		w.WriteString("\n")
		w.WriteString(fctx.indent)
		op.format(w, fctx)
	}
}

func (op Operation) String() string {
	__antithesis_instrumentation__.Notify(90456)
	fctx := formatCtx{receiver: `x`, indent: ``}
	var buf strings.Builder
	op.format(&buf, fctx)
	return buf.String()
}

func (op Operation) format(w *strings.Builder, fctx formatCtx) {
	__antithesis_instrumentation__.Notify(90457)
	switch o := op.GetValue().(type) {
	case *GetOperation:
		__antithesis_instrumentation__.Notify(90458)
		o.format(w, fctx)
	case *PutOperation:
		__antithesis_instrumentation__.Notify(90459)
		o.format(w, fctx)
	case *ScanOperation:
		__antithesis_instrumentation__.Notify(90460)
		o.format(w, fctx)
	case *DeleteOperation:
		__antithesis_instrumentation__.Notify(90461)
		o.format(w, fctx)
	case *DeleteRangeOperation:
		__antithesis_instrumentation__.Notify(90462)
		o.format(w, fctx)
	case *SplitOperation:
		__antithesis_instrumentation__.Notify(90463)
		o.format(w, fctx)
	case *MergeOperation:
		__antithesis_instrumentation__.Notify(90464)
		o.format(w, fctx)
	case *ChangeReplicasOperation:
		__antithesis_instrumentation__.Notify(90465)
		o.format(w, fctx)
	case *TransferLeaseOperation:
		__antithesis_instrumentation__.Notify(90466)
		o.format(w, fctx)
	case *ChangeZoneOperation:
		__antithesis_instrumentation__.Notify(90467)
		o.format(w, fctx)
	case *BatchOperation:
		__antithesis_instrumentation__.Notify(90468)
		newFctx := fctx
		newFctx.indent = fctx.indent + `  `
		newFctx.receiver = `b`
		w.WriteString(`{`)
		o.format(w, newFctx)
		w.WriteString("\n")
		w.WriteString(newFctx.indent)
		w.WriteString(fctx.receiver)
		w.WriteString(`.Run(ctx, b)`)
		o.Result.format(w)
		w.WriteString("\n")
		w.WriteString(fctx.indent)
		w.WriteString(`}`)
	case *ClosureTxnOperation:
		__antithesis_instrumentation__.Notify(90469)
		txnName := `txn` + o.TxnID
		newFctx := fctx
		newFctx.indent = fctx.indent + `  `
		newFctx.receiver = txnName
		w.WriteString(fctx.receiver)
		fmt.Fprintf(w, `.Txn(ctx, func(ctx context.Context, %s *kv.Txn) error {`, txnName)
		formatOps(w, newFctx, o.Ops)
		if o.CommitInBatch != nil {
			__antithesis_instrumentation__.Notify(90473)
			newFctx.receiver = `b`
			o.CommitInBatch.format(w, newFctx)
			newFctx.receiver = txnName
			w.WriteString("\n")
			w.WriteString(newFctx.indent)
			w.WriteString(newFctx.receiver)
			w.WriteString(`.CommitInBatch(ctx, b)`)
			o.CommitInBatch.Result.format(w)
		} else {
			__antithesis_instrumentation__.Notify(90474)
		}
		__antithesis_instrumentation__.Notify(90470)
		w.WriteString("\n")
		w.WriteString(newFctx.indent)
		switch o.Type {
		case ClosureTxnType_Commit:
			__antithesis_instrumentation__.Notify(90475)
			w.WriteString(`return nil`)
		case ClosureTxnType_Rollback:
			__antithesis_instrumentation__.Notify(90476)
			w.WriteString(`return errors.New("rollback")`)
		default:
			__antithesis_instrumentation__.Notify(90477)
			panic(errors.AssertionFailedf(`unknown closure txn type: %s`, o.Type))
		}
		__antithesis_instrumentation__.Notify(90471)
		w.WriteString("\n")
		w.WriteString(fctx.indent)
		w.WriteString(`})`)
		o.Result.format(w)
		if o.Txn != nil {
			__antithesis_instrumentation__.Notify(90478)
			fmt.Fprintf(w, ` txnpb:(%s)`, o.Txn)
		} else {
			__antithesis_instrumentation__.Notify(90479)
		}
	default:
		__antithesis_instrumentation__.Notify(90472)
		fmt.Fprintf(w, "%v", op.GetValue())
	}
}

func (op GetOperation) format(w *strings.Builder, fctx formatCtx) {
	__antithesis_instrumentation__.Notify(90480)
	methodName := `Get`
	if op.ForUpdate {
		__antithesis_instrumentation__.Notify(90482)
		methodName = `GetForUpdate`
	} else {
		__antithesis_instrumentation__.Notify(90483)
	}
	__antithesis_instrumentation__.Notify(90481)
	fmt.Fprintf(w, `%s.%s(ctx, %s)`, fctx.receiver, methodName, roachpb.Key(op.Key))
	switch op.Result.Type {
	case ResultType_Error:
		__antithesis_instrumentation__.Notify(90484)
		err := errors.DecodeError(context.TODO(), *op.Result.Err)
		fmt.Fprintf(w, ` // (nil, %s)`, err.Error())
	case ResultType_Value:
		__antithesis_instrumentation__.Notify(90485)
		v := `nil`
		if len(op.Result.Value) > 0 {
			__antithesis_instrumentation__.Notify(90488)
			v = `"` + mustGetStringValue(op.Result.Value) + `"`
		} else {
			__antithesis_instrumentation__.Notify(90489)
		}
		__antithesis_instrumentation__.Notify(90486)
		fmt.Fprintf(w, ` // (%s, nil)`, v)
	default:
		__antithesis_instrumentation__.Notify(90487)
	}
}

func (op PutOperation) format(w *strings.Builder, fctx formatCtx) {
	__antithesis_instrumentation__.Notify(90490)
	fmt.Fprintf(w, `%s.Put(ctx, %s, %s)`, fctx.receiver, roachpb.Key(op.Key), op.Value)
	op.Result.format(w)
}

func (op ScanOperation) format(w *strings.Builder, fctx formatCtx) {
	__antithesis_instrumentation__.Notify(90491)
	methodName := `Scan`
	if op.ForUpdate {
		__antithesis_instrumentation__.Notify(90495)
		methodName = `ScanForUpdate`
	} else {
		__antithesis_instrumentation__.Notify(90496)
	}
	__antithesis_instrumentation__.Notify(90492)
	if op.Reverse {
		__antithesis_instrumentation__.Notify(90497)
		methodName = `Reverse` + methodName
	} else {
		__antithesis_instrumentation__.Notify(90498)
	}
	__antithesis_instrumentation__.Notify(90493)

	maxRowsArg := `, 0`
	if fctx.receiver == `b` {
		__antithesis_instrumentation__.Notify(90499)
		maxRowsArg = ``
	} else {
		__antithesis_instrumentation__.Notify(90500)
	}
	__antithesis_instrumentation__.Notify(90494)
	fmt.Fprintf(w, `%s.%s(ctx, %s, %s%s)`, fctx.receiver, methodName, roachpb.Key(op.Key), roachpb.Key(op.EndKey), maxRowsArg)
	switch op.Result.Type {
	case ResultType_Error:
		__antithesis_instrumentation__.Notify(90501)
		err := errors.DecodeError(context.TODO(), *op.Result.Err)
		fmt.Fprintf(w, ` // (nil, %s)`, err.Error())
	case ResultType_Values:
		__antithesis_instrumentation__.Notify(90502)
		var kvs strings.Builder
		for i, kv := range op.Result.Values {
			__antithesis_instrumentation__.Notify(90505)
			if i > 0 {
				__antithesis_instrumentation__.Notify(90507)
				kvs.WriteString(`, `)
			} else {
				__antithesis_instrumentation__.Notify(90508)
			}
			__antithesis_instrumentation__.Notify(90506)
			kvs.WriteByte('"')
			kvs.WriteString(string(kv.Key))
			kvs.WriteString(`":"`)
			kvs.WriteString(mustGetStringValue(kv.Value))
			kvs.WriteByte('"')
		}
		__antithesis_instrumentation__.Notify(90503)
		fmt.Fprintf(w, ` // ([%s], nil)`, kvs.String())
	default:
		__antithesis_instrumentation__.Notify(90504)
	}
}

func (op DeleteOperation) format(w *strings.Builder, fctx formatCtx) {
	__antithesis_instrumentation__.Notify(90509)
	fmt.Fprintf(w, `%s.Del(ctx, %s)`, fctx.receiver, roachpb.Key(op.Key))
	op.Result.format(w)
}

func (op DeleteRangeOperation) format(w *strings.Builder, fctx formatCtx) {
	__antithesis_instrumentation__.Notify(90510)
	fmt.Fprintf(w, `%s.DelRange(ctx, %s, %s, true)`, fctx.receiver, roachpb.Key(op.Key), roachpb.Key(op.EndKey))
	switch op.Result.Type {
	case ResultType_Error:
		__antithesis_instrumentation__.Notify(90511)
		err := errors.DecodeError(context.TODO(), *op.Result.Err)
		fmt.Fprintf(w, ` // (nil, %s)`, err.Error())
	case ResultType_Keys:
		__antithesis_instrumentation__.Notify(90512)
		var keysW strings.Builder
		for i, key := range op.Result.Keys {
			__antithesis_instrumentation__.Notify(90515)
			if i > 0 {
				__antithesis_instrumentation__.Notify(90517)
				keysW.WriteString(`, `)
			} else {
				__antithesis_instrumentation__.Notify(90518)
			}
			__antithesis_instrumentation__.Notify(90516)
			keysW.WriteByte('"')
			keysW.WriteString(string(key))
			keysW.WriteString(`"`)
		}
		__antithesis_instrumentation__.Notify(90513)
		fmt.Fprintf(w, ` // ([%s], nil)`, keysW.String())
	default:
		__antithesis_instrumentation__.Notify(90514)
	}
}

func (op SplitOperation) format(w *strings.Builder, fctx formatCtx) {
	__antithesis_instrumentation__.Notify(90519)
	fmt.Fprintf(w, `%s.AdminSplit(ctx, %s)`, fctx.receiver, roachpb.Key(op.Key))
	op.Result.format(w)
}

func (op MergeOperation) format(w *strings.Builder, fctx formatCtx) {
	__antithesis_instrumentation__.Notify(90520)
	fmt.Fprintf(w, `%s.AdminMerge(ctx, %s)`, fctx.receiver, roachpb.Key(op.Key))
	op.Result.format(w)
}

func (op BatchOperation) format(w *strings.Builder, fctx formatCtx) {
	__antithesis_instrumentation__.Notify(90521)
	w.WriteString("\n")
	w.WriteString(fctx.indent)
	w.WriteString(`b := &Batch{}`)
	formatOps(w, fctx, op.Ops)
}

func (op ChangeReplicasOperation) format(w *strings.Builder, fctx formatCtx) {
	__antithesis_instrumentation__.Notify(90522)
	fmt.Fprintf(w, `%s.AdminChangeReplicas(ctx, %s, %s)`, fctx.receiver, roachpb.Key(op.Key), op.Changes)
	op.Result.format(w)
}

func (op TransferLeaseOperation) format(w *strings.Builder, fctx formatCtx) {
	__antithesis_instrumentation__.Notify(90523)
	fmt.Fprintf(w, `%s.TransferLeaseOperation(ctx, %s, %d)`, fctx.receiver, roachpb.Key(op.Key), op.Target)
	op.Result.format(w)
}

func (op ChangeZoneOperation) format(w *strings.Builder, fctx formatCtx) {
	__antithesis_instrumentation__.Notify(90524)
	fmt.Fprintf(w, `env.UpdateZoneConfig(ctx, %s)`, op.Type)
	op.Result.format(w)
}

func (r Result) format(w *strings.Builder) {
	__antithesis_instrumentation__.Notify(90525)
	switch r.Type {
	case ResultType_NoError:
		__antithesis_instrumentation__.Notify(90526)
		fmt.Fprintf(w, ` // nil`)
	case ResultType_Error:
		__antithesis_instrumentation__.Notify(90527)
		err := errors.DecodeError(context.TODO(), *r.Err)
		fmt.Fprintf(w, ` // %s`, err.Error())
	default:
		__antithesis_instrumentation__.Notify(90528)
	}
}
