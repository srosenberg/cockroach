package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/sql/roleoption"
)

func (s *baseStatusServer) UserSQLRoles(
	ctx context.Context, req *serverpb.UserSQLRolesRequest,
) (_ *serverpb.UserSQLRolesResponse, retErr error) {
	__antithesis_instrumentation__.Notify(239528)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	username, isAdmin, err := s.privilegeChecker.getUserAndRole(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(239531)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(239532)
	}
	__antithesis_instrumentation__.Notify(239529)

	var resp serverpb.UserSQLRolesResponse
	if !isAdmin {
		__antithesis_instrumentation__.Notify(239533)
		for name := range roleoption.ByName {
			__antithesis_instrumentation__.Notify(239534)
			hasRole, err := s.privilegeChecker.hasRoleOption(ctx, username, roleoption.ByName[name])
			if err != nil {
				__antithesis_instrumentation__.Notify(239536)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(239537)
			}
			__antithesis_instrumentation__.Notify(239535)
			if hasRole {
				__antithesis_instrumentation__.Notify(239538)
				resp.Roles = append(resp.Roles, name)
			} else {
				__antithesis_instrumentation__.Notify(239539)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(239540)
		resp.Roles = append(resp.Roles, "ADMIN")
	}
	__antithesis_instrumentation__.Notify(239530)
	return &resp, nil
}
