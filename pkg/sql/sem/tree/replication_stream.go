package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/errors"

type ReplicationOptions struct {
	Cursor   Expr
	Detached bool
}

var _ NodeFormatter = &ReplicationOptions{}

func (o *ReplicationOptions) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(612902)
	var addSep bool
	maybeAddSep := func() {
		__antithesis_instrumentation__.Notify(612905)
		if addSep {
			__antithesis_instrumentation__.Notify(612907)
			ctx.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(612908)
		}
		__antithesis_instrumentation__.Notify(612906)
		addSep = true
	}
	__antithesis_instrumentation__.Notify(612903)
	if o.Cursor != nil {
		__antithesis_instrumentation__.Notify(612909)
		ctx.WriteString("CURSOR=")
		o.Cursor.Format(ctx)
		addSep = true
	} else {
		__antithesis_instrumentation__.Notify(612910)
	}
	__antithesis_instrumentation__.Notify(612904)

	if o.Detached {
		__antithesis_instrumentation__.Notify(612911)
		maybeAddSep()
		ctx.WriteString("DETACHED")
	} else {
		__antithesis_instrumentation__.Notify(612912)
	}
}

func (o *ReplicationOptions) CombineWith(other *ReplicationOptions) error {
	__antithesis_instrumentation__.Notify(612913)
	if o.Cursor != nil {
		__antithesis_instrumentation__.Notify(612916)
		if other.Cursor != nil {
			__antithesis_instrumentation__.Notify(612917)
			return errors.New("CURSOR option specified multiple times")
		} else {
			__antithesis_instrumentation__.Notify(612918)
		}
	} else {
		__antithesis_instrumentation__.Notify(612919)
		o.Cursor = other.Cursor
	}
	__antithesis_instrumentation__.Notify(612914)

	if o.Detached {
		__antithesis_instrumentation__.Notify(612920)
		if other.Detached {
			__antithesis_instrumentation__.Notify(612921)
			return errors.New("detached option specified multiple times")
		} else {
			__antithesis_instrumentation__.Notify(612922)
		}
	} else {
		__antithesis_instrumentation__.Notify(612923)
		o.Detached = other.Detached
	}
	__antithesis_instrumentation__.Notify(612915)

	return nil
}

func (o ReplicationOptions) IsDefault() bool {
	__antithesis_instrumentation__.Notify(612924)
	options := ReplicationOptions{}
	return o.Cursor == options.Cursor && func() bool {
		__antithesis_instrumentation__.Notify(612925)
		return o.Detached == options.Detached == true
	}() == true
}

type ReplicationStream struct {
	Targets TargetList
	SinkURI Expr
	Options ReplicationOptions
}

var _ Statement = &ReplicationStream{}

func (n *ReplicationStream) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(612926)
	ctx.WriteString("CREATE REPLICATION STREAM FOR ")
	ctx.FormatNode(&n.Targets)

	if n.SinkURI != nil {
		__antithesis_instrumentation__.Notify(612928)
		ctx.WriteString(" INTO ")
		ctx.FormatNode(n.SinkURI)
	} else {
		__antithesis_instrumentation__.Notify(612929)
	}
	__antithesis_instrumentation__.Notify(612927)
	if !n.Options.IsDefault() {
		__antithesis_instrumentation__.Notify(612930)
		ctx.WriteString(" WITH ")
		ctx.FormatNode(&n.Options)
	} else {
		__antithesis_instrumentation__.Notify(612931)
	}
}
