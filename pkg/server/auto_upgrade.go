package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness/livenesspb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/errors"
)

func (s *Server) startAttemptUpgrade(ctx context.Context) {
	__antithesis_instrumentation__.Notify(189603)
	ctx, cancel := s.stopper.WithCancelOnQuiesce(ctx)
	if err := s.stopper.RunAsyncTask(ctx, "auto-upgrade", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(189604)
		defer cancel()
		retryOpts := retry.Options{
			InitialBackoff: time.Second,
			MaxBackoff:     30 * time.Second,
			Multiplier:     2,
			Closer:         s.stopper.ShouldQuiesce(),
		}

		if k := s.cfg.TestingKnobs.Server; k != nil {
			__antithesis_instrumentation__.Notify(189606)
			upgradeTestingKnobs := k.(*TestingKnobs)
			if disableCh := upgradeTestingKnobs.DisableAutomaticVersionUpgrade; disableCh != nil {
				__antithesis_instrumentation__.Notify(189607)
				log.Infof(ctx, "auto upgrade disabled by testing")
				select {
				case <-disableCh:
					__antithesis_instrumentation__.Notify(189608)
					log.Infof(ctx, "auto upgrade no longer disabled by testing")
				case <-s.stopper.ShouldQuiesce():
					__antithesis_instrumentation__.Notify(189609)
					return
				}
			} else {
				__antithesis_instrumentation__.Notify(189610)
			}
		} else {
			__antithesis_instrumentation__.Notify(189611)
		}
		__antithesis_instrumentation__.Notify(189605)

		for r := retry.StartWithCtx(ctx, retryOpts); r.Next(); {
			__antithesis_instrumentation__.Notify(189612)

			if quit, err := s.upgradeStatus(ctx); err != nil {
				__antithesis_instrumentation__.Notify(189614)
				log.Infof(ctx, "failed attempt to upgrade cluster version, error: %s", err)
				continue
			} else {
				__antithesis_instrumentation__.Notify(189615)
				if quit {
					__antithesis_instrumentation__.Notify(189616)
					log.Info(ctx, "no need to upgrade, cluster already at the newest version")
					return
				} else {
					__antithesis_instrumentation__.Notify(189617)
				}
			}
			__antithesis_instrumentation__.Notify(189613)

			upgradeRetryOpts := retry.Options{
				InitialBackoff: 5 * time.Second,
				MaxBackoff:     10 * time.Second,
				Multiplier:     2,
				Closer:         s.stopper.ShouldQuiesce(),
			}

			for ur := retry.StartWithCtx(ctx, upgradeRetryOpts); ur.Next(); {
				__antithesis_instrumentation__.Notify(189618)
				if _, err := s.sqlServer.internalExecutor.ExecEx(
					ctx, "set-version", nil,
					sessiondata.InternalExecutorOverride{User: security.RootUserName()},
					"SET CLUSTER SETTING version = crdb_internal.node_executable_version();",
				); err != nil {
					__antithesis_instrumentation__.Notify(189619)
					log.Infof(ctx, "error when finalizing cluster version upgrade: %s", err)
				} else {
					__antithesis_instrumentation__.Notify(189620)
					log.Info(ctx, "successfully upgraded cluster version")
					return
				}
			}
		}
	}); err != nil {
		__antithesis_instrumentation__.Notify(189621)
		cancel()
		log.Infof(ctx, "failed attempt to upgrade cluster version, error: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(189622)
	}
}

func (s *Server) upgradeStatus(ctx context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(189623)

	clusterVersion, err := s.clusterVersion(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(189632)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(189633)
	}
	__antithesis_instrumentation__.Notify(189624)

	nodesWithLiveness, err := s.status.nodesStatusWithLiveness(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(189634)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(189635)
	}
	__antithesis_instrumentation__.Notify(189625)

	var newVersion string
	var notRunningErr error
	for nodeID, st := range nodesWithLiveness {
		__antithesis_instrumentation__.Notify(189636)
		if st.livenessStatus != livenesspb.NodeLivenessStatus_LIVE && func() bool {
			__antithesis_instrumentation__.Notify(189638)
			return st.livenessStatus != livenesspb.NodeLivenessStatus_DECOMMISSIONING == true
		}() == true {
			__antithesis_instrumentation__.Notify(189639)

			if notRunningErr == nil {
				__antithesis_instrumentation__.Notify(189641)
				notRunningErr = errors.Errorf("node %d not running (%s), cannot determine version", nodeID, st.livenessStatus)
			} else {
				__antithesis_instrumentation__.Notify(189642)
			}
			__antithesis_instrumentation__.Notify(189640)
			continue
		} else {
			__antithesis_instrumentation__.Notify(189643)
		}
		__antithesis_instrumentation__.Notify(189637)

		version := st.NodeStatus.Desc.ServerVersion.String()
		if newVersion == "" {
			__antithesis_instrumentation__.Notify(189644)
			newVersion = version
		} else {
			__antithesis_instrumentation__.Notify(189645)
			if version != newVersion {
				__antithesis_instrumentation__.Notify(189646)
				return false, errors.Newf("not all nodes are running the latest version yet (saw %s and %s)", newVersion, version)
			} else {
				__antithesis_instrumentation__.Notify(189647)
			}
		}
	}
	__antithesis_instrumentation__.Notify(189626)

	if newVersion == "" {
		__antithesis_instrumentation__.Notify(189648)
		return false, errors.Errorf("no live nodes found")
	} else {
		__antithesis_instrumentation__.Notify(189649)
	}
	__antithesis_instrumentation__.Notify(189627)

	if newVersion == clusterVersion {
		__antithesis_instrumentation__.Notify(189650)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(189651)
	}
	__antithesis_instrumentation__.Notify(189628)

	if notRunningErr != nil {
		__antithesis_instrumentation__.Notify(189652)
		return false, notRunningErr
	} else {
		__antithesis_instrumentation__.Notify(189653)
	}
	__antithesis_instrumentation__.Notify(189629)

	row, err := s.sqlServer.internalExecutor.QueryRowEx(
		ctx, "read-downgrade", nil,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		"SELECT value FROM system.settings WHERE name = 'cluster.preserve_downgrade_option';",
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(189654)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(189655)
	}
	__antithesis_instrumentation__.Notify(189630)

	if row != nil {
		__antithesis_instrumentation__.Notify(189656)
		downgradeVersion := string(tree.MustBeDString(row[0]))

		if clusterVersion == downgradeVersion {
			__antithesis_instrumentation__.Notify(189657)
			return false, errors.Errorf("auto upgrade is disabled for current version: %s", clusterVersion)
		} else {
			__antithesis_instrumentation__.Notify(189658)
		}
	} else {
		__antithesis_instrumentation__.Notify(189659)
	}
	__antithesis_instrumentation__.Notify(189631)

	return false, nil
}

func (s *Server) clusterVersion(ctx context.Context) (string, error) {
	__antithesis_instrumentation__.Notify(189660)
	row, err := s.sqlServer.internalExecutor.QueryRowEx(
		ctx, "show-version", nil,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		"SHOW CLUSTER SETTING version;",
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(189663)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(189664)
	}
	__antithesis_instrumentation__.Notify(189661)
	if row == nil {
		__antithesis_instrumentation__.Notify(189665)
		return "", errors.New("cluster version is not set")
	} else {
		__antithesis_instrumentation__.Notify(189666)
	}
	__antithesis_instrumentation__.Notify(189662)
	clusterVersion := string(tree.MustBeDString(row[0]))

	return clusterVersion, nil
}
