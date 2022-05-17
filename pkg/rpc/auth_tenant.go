package rpc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"strconv"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
	"google.golang.org/grpc"
)

type tenantAuthorizer struct {
	tenantID roachpb.TenantID
}

func tenantFromCommonName(commonName string) (roachpb.TenantID, error) {
	__antithesis_instrumentation__.Notify(184039)
	tenID, err := strconv.ParseUint(commonName, 10, 64)
	if err != nil {
		__antithesis_instrumentation__.Notify(184042)
		return roachpb.TenantID{}, authErrorf("could not parse tenant ID from Common Name (CN): %s", err)
	} else {
		__antithesis_instrumentation__.Notify(184043)
	}
	__antithesis_instrumentation__.Notify(184040)
	if tenID < roachpb.MinTenantID.ToUint64() || func() bool {
		__antithesis_instrumentation__.Notify(184044)
		return tenID > roachpb.MaxTenantID.ToUint64() == true
	}() == true {
		__antithesis_instrumentation__.Notify(184045)
		return roachpb.TenantID{}, authErrorf("invalid tenant ID %d in Common Name (CN)", tenID)
	} else {
		__antithesis_instrumentation__.Notify(184046)
	}
	__antithesis_instrumentation__.Notify(184041)
	return roachpb.MakeTenantID(tenID), nil
}

func (a tenantAuthorizer) authorize(
	tenID roachpb.TenantID, fullMethod string, req interface{},
) error {
	__antithesis_instrumentation__.Notify(184047)
	switch fullMethod {
	case "/cockroach.roachpb.Internal/Batch":
		__antithesis_instrumentation__.Notify(184048)
		return a.authBatch(tenID, req.(*roachpb.BatchRequest))

	case "/cockroach.roachpb.Internal/RangeLookup":
		__antithesis_instrumentation__.Notify(184049)
		return a.authRangeLookup(tenID, req.(*roachpb.RangeLookupRequest))

	case "/cockroach.roachpb.Internal/RangeFeed":
		__antithesis_instrumentation__.Notify(184050)
		return a.authRangeFeed(tenID, req.(*roachpb.RangeFeedRequest))

	case "/cockroach.roachpb.Internal/GossipSubscription":
		__antithesis_instrumentation__.Notify(184051)
		return a.authGossipSubscription(tenID, req.(*roachpb.GossipSubscriptionRequest))

	case "/cockroach.roachpb.Internal/TokenBucket":
		__antithesis_instrumentation__.Notify(184052)
		return a.authTokenBucket(tenID, req.(*roachpb.TokenBucketRequest))

	case "/cockroach.roachpb.Internal/TenantSettings":
		__antithesis_instrumentation__.Notify(184053)
		return a.authTenantSettings(tenID, req.(*roachpb.TenantSettingsRequest))

	case "/cockroach.rpc.Heartbeat/Ping":
		__antithesis_instrumentation__.Notify(184054)
		return nil

	case "/cockroach.server.serverpb.Status/Regions":
		__antithesis_instrumentation__.Notify(184055)
		return nil

	case "/cockroach.server.serverpb.Status/Statements":
		__antithesis_instrumentation__.Notify(184056)
		return a.authTenant(tenID)

	case "/cockroach.server.serverpb.Status/CombinedStatementStats":
		__antithesis_instrumentation__.Notify(184057)
		return a.authTenant(tenID)

	case "/cockroach.server.serverpb.Status/ResetSQLStats":
		__antithesis_instrumentation__.Notify(184058)
		return a.authTenant(tenID)

	case "/cockroach.server.serverpb.Status/ListContentionEvents":
		__antithesis_instrumentation__.Notify(184059)
		return a.authTenant(tenID)

	case "/cockroach.server.serverpb.Status/ListLocalContentionEvents":
		__antithesis_instrumentation__.Notify(184060)
		return a.authTenant(tenID)

	case "/cockroach.server.serverpb.Status/ListSessions":
		__antithesis_instrumentation__.Notify(184061)
		return a.authTenant(tenID)

	case "/cockroach.server.serverpb.Status/ListLocalSessions":
		__antithesis_instrumentation__.Notify(184062)
		return a.authTenant(tenID)

	case "/cockroach.server.serverpb.Status/IndexUsageStatistics":
		__antithesis_instrumentation__.Notify(184063)
		return a.authTenant(tenID)

	case "/cockroach.server.serverpb.Status/CancelSession":
		__antithesis_instrumentation__.Notify(184064)
		return a.authTenant(tenID)

	case "/cockroach.server.serverpb.Status/CancelLocalSession":
		__antithesis_instrumentation__.Notify(184065)
		return a.authTenant(tenID)

	case "/cockroach.server.serverpb.Status/CancelQuery":
		__antithesis_instrumentation__.Notify(184066)
		return a.authTenant(tenID)

	case "/cockroach.server.serverpb.Status/TenantRanges":
		__antithesis_instrumentation__.Notify(184067)
		return a.authTenantRanges(tenID)

	case "/cockroach.server.serverpb.Status/CancelLocalQuery":
		__antithesis_instrumentation__.Notify(184068)
		return a.authTenant(tenID)

	case "/cockroach.server.serverpb.Status/TransactionContentionEvents":
		__antithesis_instrumentation__.Notify(184069)
		return a.authTenant(tenID)

	case "/cockroach.roachpb.Internal/GetSpanConfigs":
		__antithesis_instrumentation__.Notify(184070)
		return a.authGetSpanConfigs(tenID, req.(*roachpb.GetSpanConfigsRequest))

	case "/cockroach.roachpb.Internal/GetAllSystemSpanConfigsThatApply":
		__antithesis_instrumentation__.Notify(184071)
		return a.authGetAllSystemSpanConfigsThatApply(tenID, req.(*roachpb.GetAllSystemSpanConfigsThatApplyRequest))

	case "/cockroach.roachpb.Internal/UpdateSpanConfigs":
		__antithesis_instrumentation__.Notify(184072)
		return a.authUpdateSpanConfigs(tenID, req.(*roachpb.UpdateSpanConfigsRequest))

	default:
		__antithesis_instrumentation__.Notify(184073)
		return authErrorf("unknown method %q", fullMethod)
	}
}

func (a tenantAuthorizer) authBatch(tenID roachpb.TenantID, args *roachpb.BatchRequest) error {
	__antithesis_instrumentation__.Notify(184074)

	for _, ru := range args.Requests {
		__antithesis_instrumentation__.Notify(184078)
		if !reqAllowed(ru.GetInner()) {
			__antithesis_instrumentation__.Notify(184079)
			return authErrorf("request [%s] not permitted", args.Summary())
		} else {
			__antithesis_instrumentation__.Notify(184080)
		}
	}
	__antithesis_instrumentation__.Notify(184075)

	rSpan, err := keys.Range(args.Requests)
	if err != nil {
		__antithesis_instrumentation__.Notify(184081)
		return authError(err.Error())
	} else {
		__antithesis_instrumentation__.Notify(184082)
	}
	__antithesis_instrumentation__.Notify(184076)
	tenSpan := tenantPrefix(tenID)
	if !tenSpan.ContainsKeyRange(rSpan.Key, rSpan.EndKey) {
		__antithesis_instrumentation__.Notify(184083)
		return authErrorf("requested key span %s not fully contained in tenant keyspace %s", rSpan, tenSpan)
	} else {
		__antithesis_instrumentation__.Notify(184084)
	}
	__antithesis_instrumentation__.Notify(184077)
	return nil
}

var reqMethodAllowlist = [...]bool{
	roachpb.Get:            true,
	roachpb.Put:            true,
	roachpb.ConditionalPut: true,
	roachpb.Increment:      true,
	roachpb.Delete:         true,
	roachpb.DeleteRange:    true,
	roachpb.ClearRange:     true,
	roachpb.RevertRange:    true,
	roachpb.Scan:           true,
	roachpb.ReverseScan:    true,
	roachpb.EndTxn:         true,
	roachpb.AdminSplit:     true,
	roachpb.HeartbeatTxn:   true,
	roachpb.QueryTxn:       true,
	roachpb.QueryIntent:    true,
	roachpb.QueryLocks:     true,
	roachpb.InitPut:        true,
	roachpb.Export:         true,
	roachpb.AdminScatter:   true,
	roachpb.AddSSTable:     true,
	roachpb.Refresh:        true,
	roachpb.RefreshRange:   true,
}

func reqAllowed(r roachpb.Request) bool {
	__antithesis_instrumentation__.Notify(184085)
	m := int(r.Method())
	return m < len(reqMethodAllowlist) && func() bool {
		__antithesis_instrumentation__.Notify(184086)
		return reqMethodAllowlist[m] == true
	}() == true
}

func (a tenantAuthorizer) authRangeLookup(
	tenID roachpb.TenantID, args *roachpb.RangeLookupRequest,
) error {
	__antithesis_instrumentation__.Notify(184087)
	tenSpan := tenantPrefix(tenID)
	if !tenSpan.ContainsKey(args.Key) {
		__antithesis_instrumentation__.Notify(184089)
		return authErrorf("requested key %s not fully contained in tenant keyspace %s", args.Key, tenSpan)
	} else {
		__antithesis_instrumentation__.Notify(184090)
	}
	__antithesis_instrumentation__.Notify(184088)
	return nil
}

func (a tenantAuthorizer) authRangeFeed(
	tenID roachpb.TenantID, args *roachpb.RangeFeedRequest,
) error {
	__antithesis_instrumentation__.Notify(184091)
	rSpan, err := keys.SpanAddr(args.Span)
	if err != nil {
		__antithesis_instrumentation__.Notify(184094)
		return authError(err.Error())
	} else {
		__antithesis_instrumentation__.Notify(184095)
	}
	__antithesis_instrumentation__.Notify(184092)
	tenSpan := tenantPrefix(tenID)
	if !tenSpan.ContainsKeyRange(rSpan.Key, rSpan.EndKey) {
		__antithesis_instrumentation__.Notify(184096)
		return authErrorf("requested key span %s not fully contained in tenant keyspace %s", rSpan, tenSpan)
	} else {
		__antithesis_instrumentation__.Notify(184097)
	}
	__antithesis_instrumentation__.Notify(184093)
	return nil
}

func (a tenantAuthorizer) authGossipSubscription(
	tenID roachpb.TenantID, args *roachpb.GossipSubscriptionRequest,
) error {
	__antithesis_instrumentation__.Notify(184098)
	for _, pat := range args.Patterns {
		__antithesis_instrumentation__.Notify(184100)
		allowed := false
		for _, allow := range gossipSubscriptionPatternAllowlist {
			__antithesis_instrumentation__.Notify(184102)
			if pat == allow {
				__antithesis_instrumentation__.Notify(184103)
				allowed = true
				break
			} else {
				__antithesis_instrumentation__.Notify(184104)
			}
		}
		__antithesis_instrumentation__.Notify(184101)
		if !allowed {
			__antithesis_instrumentation__.Notify(184105)
			return authErrorf("requested pattern %q not permitted", pat)
		} else {
			__antithesis_instrumentation__.Notify(184106)
		}
	}
	__antithesis_instrumentation__.Notify(184099)
	return nil
}

func (a tenantAuthorizer) authTenant(id roachpb.TenantID) error {
	__antithesis_instrumentation__.Notify(184107)
	if a.tenantID != id {
		__antithesis_instrumentation__.Notify(184109)
		return authErrorf("request from tenant %s not permitted on tenant %s", id, a.tenantID)
	} else {
		__antithesis_instrumentation__.Notify(184110)
	}
	__antithesis_instrumentation__.Notify(184108)
	return nil
}

var gossipSubscriptionPatternAllowlist = []string{
	"cluster-id",
	"node:.*",
	"system-db",
}

func (a tenantAuthorizer) authTenantRanges(tenID roachpb.TenantID) error {
	__antithesis_instrumentation__.Notify(184111)
	if !tenID.IsSet() {
		__antithesis_instrumentation__.Notify(184113)
		return authErrorf("tenant ranges request with unspecified tenant not permitted.")
	} else {
		__antithesis_instrumentation__.Notify(184114)
	}
	__antithesis_instrumentation__.Notify(184112)
	return nil
}

func (a tenantAuthorizer) authTokenBucket(
	tenID roachpb.TenantID, args *roachpb.TokenBucketRequest,
) error {
	__antithesis_instrumentation__.Notify(184115)
	if args.TenantID == 0 {
		__antithesis_instrumentation__.Notify(184118)
		return authErrorf("token bucket request with unspecified tenant not permitted")
	} else {
		__antithesis_instrumentation__.Notify(184119)
	}
	__antithesis_instrumentation__.Notify(184116)
	if argTenant := roachpb.MakeTenantID(args.TenantID); argTenant != tenID {
		__antithesis_instrumentation__.Notify(184120)
		return authErrorf("token bucket request for tenant %s not permitted", argTenant)
	} else {
		__antithesis_instrumentation__.Notify(184121)
	}
	__antithesis_instrumentation__.Notify(184117)
	return nil
}

func (a tenantAuthorizer) authTenantSettings(
	tenID roachpb.TenantID, args *roachpb.TenantSettingsRequest,
) error {
	__antithesis_instrumentation__.Notify(184122)
	if !args.TenantID.IsSet() {
		__antithesis_instrumentation__.Notify(184125)
		return authErrorf("tenant settings request with unspecified tenant not permitted")
	} else {
		__antithesis_instrumentation__.Notify(184126)
	}
	__antithesis_instrumentation__.Notify(184123)
	if args.TenantID != tenID {
		__antithesis_instrumentation__.Notify(184127)
		return authErrorf("tenant settings request for tenant %s not permitted", args.TenantID)
	} else {
		__antithesis_instrumentation__.Notify(184128)
	}
	__antithesis_instrumentation__.Notify(184124)
	return nil
}

func (a tenantAuthorizer) authGetAllSystemSpanConfigsThatApply(
	tenID roachpb.TenantID, args *roachpb.GetAllSystemSpanConfigsThatApplyRequest,
) error {
	__antithesis_instrumentation__.Notify(184129)
	if !args.TenantID.IsSet() {
		__antithesis_instrumentation__.Notify(184132)
		return authErrorf(
			"GetAllSystemSpanConfigsThatApply request with unspecified tenant not permitted",
		)
	} else {
		__antithesis_instrumentation__.Notify(184133)
	}
	__antithesis_instrumentation__.Notify(184130)
	if args.TenantID != tenID {
		__antithesis_instrumentation__.Notify(184134)
		return authErrorf(
			"GetAllSystemSpanConfigsThatApply request for tenant %s not permitted", args.TenantID,
		)
	} else {
		__antithesis_instrumentation__.Notify(184135)
	}
	__antithesis_instrumentation__.Notify(184131)
	return nil
}

func (a tenantAuthorizer) authGetSpanConfigs(
	tenID roachpb.TenantID, args *roachpb.GetSpanConfigsRequest,
) error {
	__antithesis_instrumentation__.Notify(184136)
	for _, target := range args.Targets {
		__antithesis_instrumentation__.Notify(184138)
		if err := validateSpanConfigTarget(tenID, target); err != nil {
			__antithesis_instrumentation__.Notify(184139)
			return err
		} else {
			__antithesis_instrumentation__.Notify(184140)
		}
	}
	__antithesis_instrumentation__.Notify(184137)
	return nil
}

func (a tenantAuthorizer) authUpdateSpanConfigs(
	tenID roachpb.TenantID, args *roachpb.UpdateSpanConfigsRequest,
) error {
	__antithesis_instrumentation__.Notify(184141)
	for _, entry := range args.ToUpsert {
		__antithesis_instrumentation__.Notify(184144)
		if err := validateSpanConfigTarget(tenID, entry.Target); err != nil {
			__antithesis_instrumentation__.Notify(184145)
			return err
		} else {
			__antithesis_instrumentation__.Notify(184146)
		}
	}
	__antithesis_instrumentation__.Notify(184142)
	for _, target := range args.ToDelete {
		__antithesis_instrumentation__.Notify(184147)
		if err := validateSpanConfigTarget(tenID, target); err != nil {
			__antithesis_instrumentation__.Notify(184148)
			return err
		} else {
			__antithesis_instrumentation__.Notify(184149)
		}
	}
	__antithesis_instrumentation__.Notify(184143)

	return nil
}

func validateSpanConfigTarget(
	tenID roachpb.TenantID, spanConfigTarget roachpb.SpanConfigTarget,
) error {
	__antithesis_instrumentation__.Notify(184150)
	validateSystemTarget := func(target roachpb.SystemSpanConfigTarget) error {
		__antithesis_instrumentation__.Notify(184153)
		if target.SourceTenantID != tenID {
			__antithesis_instrumentation__.Notify(184158)
			return authErrorf("malformed source tenant field")
		} else {
			__antithesis_instrumentation__.Notify(184159)
		}
		__antithesis_instrumentation__.Notify(184154)

		if tenID == roachpb.SystemTenantID {
			__antithesis_instrumentation__.Notify(184160)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(184161)
		}
		__antithesis_instrumentation__.Notify(184155)

		if target.IsEntireKeyspaceTarget() {
			__antithesis_instrumentation__.Notify(184162)
			return authErrorf("secondary tenants cannot target the entire keyspace")
		} else {
			__antithesis_instrumentation__.Notify(184163)
		}
		__antithesis_instrumentation__.Notify(184156)

		if target.IsSpecificTenantKeyspaceTarget() && func() bool {
			__antithesis_instrumentation__.Notify(184164)
			return target.Type.GetSpecificTenantKeyspace().TenantID != target.SourceTenantID == true
		}() == true {
			__antithesis_instrumentation__.Notify(184165)
			return authErrorf(
				"secondary tenants cannot interact with system span configurations of other tenants",
			)
		} else {
			__antithesis_instrumentation__.Notify(184166)
		}
		__antithesis_instrumentation__.Notify(184157)

		return nil
	}
	__antithesis_instrumentation__.Notify(184151)

	validateSpan := func(sp roachpb.Span) error {
		__antithesis_instrumentation__.Notify(184167)
		tenSpan := tenantPrefix(tenID)
		rSpan, err := keys.SpanAddr(sp)
		if err != nil {
			__antithesis_instrumentation__.Notify(184170)
			return authError(err.Error())
		} else {
			__antithesis_instrumentation__.Notify(184171)
		}
		__antithesis_instrumentation__.Notify(184168)
		if !tenSpan.ContainsKeyRange(rSpan.Key, rSpan.EndKey) {
			__antithesis_instrumentation__.Notify(184172)
			return authErrorf("requested key span %s not fully contained in tenant keyspace %s", rSpan, tenSpan)
		} else {
			__antithesis_instrumentation__.Notify(184173)
		}
		__antithesis_instrumentation__.Notify(184169)
		return nil
	}
	__antithesis_instrumentation__.Notify(184152)

	switch spanConfigTarget.Union.(type) {
	case *roachpb.SpanConfigTarget_Span:
		__antithesis_instrumentation__.Notify(184174)
		return validateSpan(*spanConfigTarget.GetSpan())
	case *roachpb.SpanConfigTarget_SystemSpanConfigTarget:
		__antithesis_instrumentation__.Notify(184175)
		return validateSystemTarget(*spanConfigTarget.GetSystemSpanConfigTarget())
	default:
		__antithesis_instrumentation__.Notify(184176)
		return errors.AssertionFailedf("unknown span config target type")
	}
}

func contextWithTenant(ctx context.Context, tenID roachpb.TenantID) context.Context {
	__antithesis_instrumentation__.Notify(184177)
	ctx = roachpb.NewContextForTenant(ctx, tenID)
	ctx = logtags.AddTag(ctx, "tenant", tenID.String())
	return ctx
}

func tenantPrefix(tenID roachpb.TenantID) roachpb.RSpan {
	__antithesis_instrumentation__.Notify(184178)

	prefix := roachpb.RKey(keys.MakeTenantPrefix(tenID))
	return roachpb.RSpan{
		Key:    prefix,
		EndKey: prefix.PrefixEnd(),
	}
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx  context.Context
	recv func(interface{}) error
}

func (ss *wrappedServerStream) Context() context.Context {
	__antithesis_instrumentation__.Notify(184179)
	return ss.ctx
}

func (ss *wrappedServerStream) RecvMsg(m interface{}) error {
	__antithesis_instrumentation__.Notify(184180)
	return ss.recv(m)
}
