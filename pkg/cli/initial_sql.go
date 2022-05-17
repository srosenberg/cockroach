package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver"
	"github.com/cockroachdb/cockroach/pkg/server"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

func runInitialSQL(
	ctx context.Context, s *server.Server, startSingleNode bool, adminUser, adminPassword string,
) error {
	__antithesis_instrumentation__.Notify(33218)
	newCluster := s.InitialStart() && func() bool {
		__antithesis_instrumentation__.Notify(33222)
		return s.NodeID() == kvserver.FirstNodeID == true
	}() == true
	if !newCluster {
		__antithesis_instrumentation__.Notify(33223)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(33224)
	}
	__antithesis_instrumentation__.Notify(33219)

	if startSingleNode {
		__antithesis_instrumentation__.Notify(33225)

		if err := cliDisableReplication(ctx, s); err != nil {
			__antithesis_instrumentation__.Notify(33227)
			log.Ops.Errorf(ctx, "could not disable replication: %v", err)
			return err
		} else {
			__antithesis_instrumentation__.Notify(33228)
		}
		__antithesis_instrumentation__.Notify(33226)
		log.Ops.Infof(ctx, "Replication was disabled for this cluster.\n"+
			"When/if adding nodes in the future, update zone configurations to increase the replication factor.")
	} else {
		__antithesis_instrumentation__.Notify(33229)
	}
	__antithesis_instrumentation__.Notify(33220)

	if adminUser != "" && func() bool {
		__antithesis_instrumentation__.Notify(33230)
		return !s.Insecure() == true
	}() == true {
		__antithesis_instrumentation__.Notify(33231)
		if err := createAdminUser(ctx, s, adminUser, adminPassword); err != nil {
			__antithesis_instrumentation__.Notify(33232)
			return err
		} else {
			__antithesis_instrumentation__.Notify(33233)
		}
	} else {
		__antithesis_instrumentation__.Notify(33234)
	}
	__antithesis_instrumentation__.Notify(33221)

	return nil
}

func createAdminUser(ctx context.Context, s *server.Server, adminUser, adminPassword string) error {
	__antithesis_instrumentation__.Notify(33235)
	return s.RunLocalSQL(ctx,
		func(ctx context.Context, ie *sql.InternalExecutor) error {
			__antithesis_instrumentation__.Notify(33236)
			_, err := ie.Exec(
				ctx, "admin-user", nil,
				fmt.Sprintf("CREATE USER %s WITH PASSWORD $1", adminUser),
				adminPassword,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(33238)
				return err
			} else {
				__antithesis_instrumentation__.Notify(33239)
			}
			__antithesis_instrumentation__.Notify(33237)

			_, err = ie.Exec(ctx, "admin-user", nil, fmt.Sprintf("GRANT admin TO %s", tree.Name(adminUser)))
			return err
		})
}

func cliDisableReplication(ctx context.Context, s *server.Server) error {
	__antithesis_instrumentation__.Notify(33240)
	return s.RunLocalSQL(ctx,
		func(ctx context.Context, ie *sql.InternalExecutor) (retErr error) {
			__antithesis_instrumentation__.Notify(33241)
			it, err := ie.QueryIterator(ctx, "get-zones", nil,
				"SELECT target FROM crdb_internal.zones")
			if err != nil {
				__antithesis_instrumentation__.Notify(33245)
				return err
			} else {
				__antithesis_instrumentation__.Notify(33246)
			}
			__antithesis_instrumentation__.Notify(33242)

			defer func() {
				__antithesis_instrumentation__.Notify(33247)
				retErr = errors.CombineErrors(retErr, it.Close())
			}()
			__antithesis_instrumentation__.Notify(33243)

			var ok bool
			for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
				__antithesis_instrumentation__.Notify(33248)
				zone := string(*it.Cur()[0].(*tree.DString))
				if _, err := ie.Exec(ctx, "set-zone", nil,
					fmt.Sprintf("ALTER %s CONFIGURE ZONE USING num_replicas = 1", zone)); err != nil {
					__antithesis_instrumentation__.Notify(33249)
					return err
				} else {
					__antithesis_instrumentation__.Notify(33250)
				}
			}
			__antithesis_instrumentation__.Notify(33244)
			return err
		})
}
