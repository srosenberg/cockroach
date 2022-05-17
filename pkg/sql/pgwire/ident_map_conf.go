package pgwire

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/identmap"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

const serverIdentityMapSetting = "server.identity_map.configuration"

var connIdentityMapConf = func() *settings.StringSetting {
	__antithesis_instrumentation__.Notify(560189)
	s := settings.RegisterValidatedStringSetting(
		settings.TenantWritable,
		serverIdentityMapSetting,
		"system-identity to database-username mappings",
		"",
		func(values *settings.Values, s string) error {
			__antithesis_instrumentation__.Notify(560191)
			_, err := identmap.From(strings.NewReader(s))
			return err
		},
	)
	__antithesis_instrumentation__.Notify(560190)
	s.SetVisibility(settings.Public)
	return s
}()

func loadLocalIdentityMapUponRemoteSettingChange(
	ctx context.Context, server *Server, st *cluster.Settings,
) {
	__antithesis_instrumentation__.Notify(560192)
	val := connIdentityMapConf.Get(&st.SV)
	idMap, err := identmap.From(strings.NewReader(val))
	if err != nil {
		__antithesis_instrumentation__.Notify(560194)
		log.Ops.Warningf(ctx, "invalid %s: %v", serverIdentityMapSetting, err)
		idMap = identmap.Empty()
	} else {
		__antithesis_instrumentation__.Notify(560195)
	}
	__antithesis_instrumentation__.Notify(560193)

	server.auth.Lock()
	defer server.auth.Unlock()
	server.auth.identityMap = idMap
}
