package pgwire

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/hba"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/identmap"
	"github.com/cockroachdb/errors"
)

type RoleMapper = func(
	ctx context.Context,
	systemIdentity security.SQLUsername,
) ([]security.SQLUsername, error)

func UseProvidedIdentity(
	_ context.Context, id security.SQLUsername,
) ([]security.SQLUsername, error) {
	__antithesis_instrumentation__.Notify(561455)
	return []security.SQLUsername{id}, nil
}

var _ RoleMapper = UseProvidedIdentity

func HbaMapper(hbaEntry *hba.Entry, identMap *identmap.Conf) RoleMapper {
	__antithesis_instrumentation__.Notify(561456)
	mapName := hbaEntry.GetOption("map")
	if mapName == "" {
		__antithesis_instrumentation__.Notify(561458)
		return UseProvidedIdentity
	} else {
		__antithesis_instrumentation__.Notify(561459)
	}
	__antithesis_instrumentation__.Notify(561457)
	return func(_ context.Context, id security.SQLUsername) ([]security.SQLUsername, error) {
		__antithesis_instrumentation__.Notify(561460)
		users, err := identMap.Map(mapName, id.Normalized())
		if err != nil {
			__antithesis_instrumentation__.Notify(561463)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(561464)
		}
		__antithesis_instrumentation__.Notify(561461)
		for _, user := range users {
			__antithesis_instrumentation__.Notify(561465)
			if user.IsRootUser() || func() bool {
				__antithesis_instrumentation__.Notify(561466)
				return user.IsReserved() == true
			}() == true {
				__antithesis_instrumentation__.Notify(561467)
				return nil, errors.Newf("system identity %q mapped to reserved database role %q",
					id.Normalized(), user.Normalized())
			} else {
				__antithesis_instrumentation__.Notify(561468)
			}
		}
		__antithesis_instrumentation__.Notify(561462)
		return users, nil
	}
}
