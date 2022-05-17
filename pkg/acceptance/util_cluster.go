package acceptance

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/cockroachdb/cockroach/pkg/acceptance/cluster"
	"github.com/cockroachdb/cockroach/pkg/testutils"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/errors"
)

const (
	dockerTest = "runMode=docker"
)

var stopper = stop.NewStopper()

func RunDocker(t *testing.T, testee func(t *testing.T)) {
	__antithesis_instrumentation__.Notify(1157)
	t.Run(dockerTest, testee)
}

var reStripTestEnumeration = regexp.MustCompile(`#\d+$`)

func StartCluster(ctx context.Context, t *testing.T, cfg cluster.TestConfig) (c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(1158)
	var completed bool
	defer func() {
		__antithesis_instrumentation__.Notify(1164)
		if !completed && func() bool {
			__antithesis_instrumentation__.Notify(1165)
			return c != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(1166)
			c.AssertAndStop(ctx, t)
		} else {
			__antithesis_instrumentation__.Notify(1167)
		}
	}()
	__antithesis_instrumentation__.Notify(1159)

	parts := strings.Split(t.Name(), "/")
	if len(parts) < 2 {
		__antithesis_instrumentation__.Notify(1168)
		t.Fatal("must invoke RunDocker")
	} else {
		__antithesis_instrumentation__.Notify(1169)
	}
	__antithesis_instrumentation__.Notify(1160)

	var runMode string
	for _, part := range parts[1:] {
		__antithesis_instrumentation__.Notify(1170)
		part = reStripTestEnumeration.ReplaceAllLiteralString(part, "")
		switch part {
		case dockerTest:
			__antithesis_instrumentation__.Notify(1171)
			if runMode != "" {
				__antithesis_instrumentation__.Notify(1174)
				t.Fatalf("test has more than one run mode: %s and %s", runMode, part)
			} else {
				__antithesis_instrumentation__.Notify(1175)
			}
			__antithesis_instrumentation__.Notify(1172)
			runMode = part
		default:
			__antithesis_instrumentation__.Notify(1173)
		}
	}
	__antithesis_instrumentation__.Notify(1161)

	switch runMode {
	case dockerTest:
		__antithesis_instrumentation__.Notify(1176)
		logDir := *flagLogDir
		if logDir != "" {
			__antithesis_instrumentation__.Notify(1179)
			logDir = filepath.Join(logDir, filepath.Clean(t.Name()))
		} else {
			__antithesis_instrumentation__.Notify(1180)
		}
		__antithesis_instrumentation__.Notify(1177)
		l := cluster.CreateDocker(ctx, cfg, logDir, stopper)
		l.Start(ctx)
		c = l

	default:
		__antithesis_instrumentation__.Notify(1178)
		t.Fatalf("unable to run in mode %q, use RunDocker", runMode)
	}
	__antithesis_instrumentation__.Notify(1162)

	if !cfg.NoWait && func() bool {
		__antithesis_instrumentation__.Notify(1181)
		return cfg.InitMode != cluster.INIT_NONE == true
	}() == true {
		__antithesis_instrumentation__.Notify(1182)
		wantedReplicas := 3
		if numNodes := c.NumNodes(); numNodes < wantedReplicas {
			__antithesis_instrumentation__.Notify(1185)
			wantedReplicas = numNodes
		} else {
			__antithesis_instrumentation__.Notify(1186)
		}
		__antithesis_instrumentation__.Notify(1183)

		if wantedReplicas > 0 {
			__antithesis_instrumentation__.Notify(1187)
			log.Infof(ctx, "waiting for first range to have %d replicas", wantedReplicas)

			testutils.SucceedsSoon(t, func() error {
				__antithesis_instrumentation__.Notify(1188)
				select {
				case <-stopper.ShouldQuiesce():
					__antithesis_instrumentation__.Notify(1194)
					t.Fatal("interrupted")
				case <-time.After(time.Second):
					__antithesis_instrumentation__.Notify(1195)
				}
				__antithesis_instrumentation__.Notify(1189)

				db, err := c.NewDB(ctx, 0)
				if err != nil {
					__antithesis_instrumentation__.Notify(1196)
					t.Fatal(err)
				} else {
					__antithesis_instrumentation__.Notify(1197)
				}
				__antithesis_instrumentation__.Notify(1190)
				rows, err := db.Query(`SELECT array_length(replicas, 1) FROM crdb_internal.ranges LIMIT 1`)
				if err != nil {
					__antithesis_instrumentation__.Notify(1198)

					if testutils.IsError(err, "(table|relation) \"crdb_internal.ranges\" does not exist") {
						__antithesis_instrumentation__.Notify(1200)
						return nil
					} else {
						__antithesis_instrumentation__.Notify(1201)
					}
					__antithesis_instrumentation__.Notify(1199)
					t.Fatal(err)
				} else {
					__antithesis_instrumentation__.Notify(1202)
				}
				__antithesis_instrumentation__.Notify(1191)
				defer rows.Close()
				var foundReplicas int
				if rows.Next() {
					__antithesis_instrumentation__.Notify(1203)
					if err = rows.Scan(&foundReplicas); err != nil {
						__antithesis_instrumentation__.Notify(1205)
						t.Fatalf("unable to scan for length of replicas array: %s", err)
					} else {
						__antithesis_instrumentation__.Notify(1206)
					}
					__antithesis_instrumentation__.Notify(1204)
					if log.V(1) {
						__antithesis_instrumentation__.Notify(1207)
						log.Infof(ctx, "found %d replicas", foundReplicas)
					} else {
						__antithesis_instrumentation__.Notify(1208)
					}
				} else {
					__antithesis_instrumentation__.Notify(1209)
					return errors.Errorf("no ranges listed")
				}
				__antithesis_instrumentation__.Notify(1192)

				if foundReplicas < wantedReplicas {
					__antithesis_instrumentation__.Notify(1210)
					return errors.Errorf("expected %d replicas, only found %d", wantedReplicas, foundReplicas)
				} else {
					__antithesis_instrumentation__.Notify(1211)
				}
				__antithesis_instrumentation__.Notify(1193)
				return nil
			})
		} else {
			__antithesis_instrumentation__.Notify(1212)
		}
		__antithesis_instrumentation__.Notify(1184)

		for i := 0; i < c.NumNodes(); i++ {
			__antithesis_instrumentation__.Notify(1213)
			testutils.SucceedsSoon(t, func() error {
				__antithesis_instrumentation__.Notify(1214)
				db, err := c.NewDB(ctx, i)
				if err != nil {
					__antithesis_instrumentation__.Notify(1217)
					return err
				} else {
					__antithesis_instrumentation__.Notify(1218)
				}
				__antithesis_instrumentation__.Notify(1215)
				if _, err := db.Exec("SHOW DATABASES"); err != nil {
					__antithesis_instrumentation__.Notify(1219)
					return err
				} else {
					__antithesis_instrumentation__.Notify(1220)
				}
				__antithesis_instrumentation__.Notify(1216)
				return nil
			})
		}
	} else {
		__antithesis_instrumentation__.Notify(1221)
	}
	__antithesis_instrumentation__.Notify(1163)

	completed = true
	return c
}
