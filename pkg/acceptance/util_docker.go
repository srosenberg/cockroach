package acceptance

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/acceptance/cluster"
	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/build/bazel"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/containerd/containerd/platforms"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
)

func defaultContainerConfig() container.Config {
	__antithesis_instrumentation__.Notify(1222)
	return container.Config{
		Image: acceptanceImage,
		Env: []string{
			fmt.Sprintf("PGUSER=%s", security.RootUser),
			fmt.Sprintf("PGPORT=%s", base.DefaultPort),
			"PGSSLCERT=/certs/client.root.crt",
			"PGSSLKEY=/certs/client.root.key",
		},
		Entrypoint: []string{"autouseradd", "-u", "roach", "-C", "/home/roach", "--"},
	}
}

func testDockerFail(ctx context.Context, t *testing.T, name string, cmd []string) {
	__antithesis_instrumentation__.Notify(1223)
	containerConfig := defaultContainerConfig()
	containerConfig.Cmd = cmd
	if err := testDockerSingleNode(ctx, t, name, containerConfig); err == nil {
		__antithesis_instrumentation__.Notify(1224)
		t.Error("expected failure")
	} else {
		__antithesis_instrumentation__.Notify(1225)
	}
}

func testDockerSuccess(ctx context.Context, t *testing.T, name string, cmd []string) {
	__antithesis_instrumentation__.Notify(1226)
	containerConfig := defaultContainerConfig()
	containerConfig.Cmd = cmd
	if err := testDockerSingleNode(ctx, t, name, containerConfig); err != nil {
		__antithesis_instrumentation__.Notify(1227)
		t.Error(err)
	} else {
		__antithesis_instrumentation__.Notify(1228)
	}
}

const (
	acceptanceImage = "docker.io/cockroachdb/acceptance:20200303-091324"
)

func testDocker(
	ctx context.Context, t *testing.T, num int, name string, containerConfig container.Config,
) error {
	__antithesis_instrumentation__.Notify(1229)
	var err error
	RunDocker(t, func(t *testing.T) {
		__antithesis_instrumentation__.Notify(1231)
		cfg := cluster.TestConfig{
			Name:     name,
			Duration: *flagDuration,
		}
		for i := 0; i < num; i++ {
			__antithesis_instrumentation__.Notify(1237)
			cfg.Nodes = append(cfg.Nodes, cluster.NodeConfig{Stores: []cluster.StoreConfig{{}}})
		}
		__antithesis_instrumentation__.Notify(1232)
		l := StartCluster(ctx, t, cfg).(*cluster.DockerCluster)
		defer l.AssertAndStop(ctx, t)

		if len(l.Nodes) > 0 {
			__antithesis_instrumentation__.Notify(1238)
			containerConfig.Env = append(containerConfig.Env, "PGHOST="+l.Hostname(0))
		} else {
			__antithesis_instrumentation__.Notify(1239)
		}
		__antithesis_instrumentation__.Notify(1233)
		var pwd string
		pwd, err = os.Getwd()
		if err != nil {
			__antithesis_instrumentation__.Notify(1240)
			return
		} else {
			__antithesis_instrumentation__.Notify(1241)
		}
		__antithesis_instrumentation__.Notify(1234)
		testdataDir := filepath.Join(pwd, "testdata")
		if bazel.BuiltWithBazel() {
			__antithesis_instrumentation__.Notify(1242)
			testdataDir, err = ioutil.TempDir("", "")
			if err != nil {
				__antithesis_instrumentation__.Notify(1245)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(1246)
			}
			__antithesis_instrumentation__.Notify(1243)

			err = copyRunfiles("testdata", testdataDir)
			if err != nil {
				__antithesis_instrumentation__.Notify(1247)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(1248)
			}
			__antithesis_instrumentation__.Notify(1244)
			defer func() {
				__antithesis_instrumentation__.Notify(1249)
				_ = os.RemoveAll(testdataDir)
			}()
		} else {
			__antithesis_instrumentation__.Notify(1250)
		}
		__antithesis_instrumentation__.Notify(1235)
		hostConfig := container.HostConfig{
			NetworkMode: "host",
			Binds:       []string{testdataDir + ":/mnt/data"},
		}
		if bazel.BuiltWithBazel() {
			__antithesis_instrumentation__.Notify(1251)
			interactivetestsDir, err := ioutil.TempDir("", "")
			if err != nil {
				__antithesis_instrumentation__.Notify(1255)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(1256)
			}
			__antithesis_instrumentation__.Notify(1252)

			err = copyRunfiles("../cli/interactive_tests", interactivetestsDir)
			if err != nil {
				__antithesis_instrumentation__.Notify(1257)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(1258)
			}
			__antithesis_instrumentation__.Notify(1253)
			defer func() {
				__antithesis_instrumentation__.Notify(1259)
				_ = os.RemoveAll(interactivetestsDir)
			}()
			__antithesis_instrumentation__.Notify(1254)
			hostConfig.Binds = append(hostConfig.Binds, interactivetestsDir+":/mnt/interactive_tests")
		} else {
			__antithesis_instrumentation__.Notify(1260)
		}
		__antithesis_instrumentation__.Notify(1236)
		err = l.OneShot(
			ctx, acceptanceImage, types.ImagePullOptions{}, containerConfig, hostConfig,
			platforms.DefaultSpec(), "docker-"+name,
		)
		preserveLogs := err != nil
		l.Cleanup(ctx, preserveLogs)
	})
	__antithesis_instrumentation__.Notify(1230)
	return err
}

func copyRunfiles(source, destination string) error {
	__antithesis_instrumentation__.Notify(1261)
	return filepath.WalkDir(source, func(path string, dirEntry os.DirEntry, walkErr error) error {
		__antithesis_instrumentation__.Notify(1262)
		if walkErr != nil {
			__antithesis_instrumentation__.Notify(1267)
			return walkErr
		} else {
			__antithesis_instrumentation__.Notify(1268)
		}
		__antithesis_instrumentation__.Notify(1263)
		relPath := strings.Replace(path, source, "", 1)
		if relPath == "" {
			__antithesis_instrumentation__.Notify(1269)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(1270)
		}
		__antithesis_instrumentation__.Notify(1264)
		if dirEntry.IsDir() {
			__antithesis_instrumentation__.Notify(1271)
			return os.Mkdir(filepath.Join(destination, relPath), 0755)
		} else {
			__antithesis_instrumentation__.Notify(1272)
		}
		__antithesis_instrumentation__.Notify(1265)
		data, err := ioutil.ReadFile(filepath.Join(source, relPath))
		if err != nil {
			__antithesis_instrumentation__.Notify(1273)
			return err
		} else {
			__antithesis_instrumentation__.Notify(1274)
		}
		__antithesis_instrumentation__.Notify(1266)
		return ioutil.WriteFile(filepath.Join(destination, relPath), data, 0755)
	})
}

func testDockerSingleNode(
	ctx context.Context, t *testing.T, name string, containerConfig container.Config,
) error {
	__antithesis_instrumentation__.Notify(1275)
	return testDocker(ctx, t, 1, name, containerConfig)
}

func testDockerOneShot(
	ctx context.Context, t *testing.T, name string, containerConfig container.Config,
) error {
	__antithesis_instrumentation__.Notify(1276)
	return testDocker(ctx, t, 0, name, containerConfig)
}
