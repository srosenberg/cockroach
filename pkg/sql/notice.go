package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgnotice"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

var NoticesEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.notices.enabled",
	"enable notices in the server/client protocol being sent",
	true,
).WithPublic()

type noticeSender interface {
	BufferNotice(pgnotice.Notice)
}

func (p *planner) BufferClientNotice(ctx context.Context, notice pgnotice.Notice) {
	__antithesis_instrumentation__.Notify(501847)
	if log.V(2) {
		__antithesis_instrumentation__.Notify(501851)
		log.Infof(ctx, "buffered notice: %+v", notice)
	} else {
		__antithesis_instrumentation__.Notify(501852)
	}
	__antithesis_instrumentation__.Notify(501848)
	noticeSeverity, ok := pgnotice.ParseDisplaySeverity(pgerror.GetSeverity(notice))
	if !ok {
		__antithesis_instrumentation__.Notify(501853)
		noticeSeverity = pgnotice.DisplaySeverityNotice
	} else {
		__antithesis_instrumentation__.Notify(501854)
	}
	__antithesis_instrumentation__.Notify(501849)
	if p.noticeSender == nil || func() bool {
		__antithesis_instrumentation__.Notify(501855)
		return noticeSeverity > pgnotice.DisplaySeverity(p.SessionData().NoticeDisplaySeverity) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(501856)
		return !NoticesEnabled.Get(&p.execCfg.Settings.SV) == true
	}() == true {
		__antithesis_instrumentation__.Notify(501857)

		return
	} else {
		__antithesis_instrumentation__.Notify(501858)
	}
	__antithesis_instrumentation__.Notify(501850)
	p.noticeSender.BufferNotice(notice)
}
