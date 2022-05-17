package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/spec"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
)

var jepsenNemeses = []struct {
	name, config string
}{
	{"majority-ring", "--nemesis majority-ring"},
	{"split", "--nemesis split"},
	{"start-kill-2", "--nemesis start-kill-2"},
	{"start-stop-2", "--nemesis start-stop-2"},
	{"strobe-skews", "--nemesis strobe-skews"},

	{"majority-ring-start-kill-2", "--nemesis majority-ring --nemesis2 start-kill-2"},
	{"parts-start-kill-2", "--nemesis parts --nemesis2 start-kill-2"},
}

func initJepsen(ctx context.Context, t test.Test, c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(48708)

	if c.IsLocal() {
		__antithesis_instrumentation__.Notify(48718)
		t.Fatal("local execution not supported")
	} else {
		__antithesis_instrumentation__.Notify(48719)
	}
	__antithesis_instrumentation__.Notify(48709)

	if c.IsLocal() {
		__antithesis_instrumentation__.Notify(48720)

		return
	} else {
		__antithesis_instrumentation__.Notify(48721)
	}
	__antithesis_instrumentation__.Notify(48710)

	controller := c.Node(c.Spec().NodeCount)
	workers := c.Range(1, c.Spec().NodeCount-1)

	if err := c.GitClone(
		ctx, t.L(),
		"https://github.com/cockroachdb/jepsen", "/mnt/data1/jepsen", "tc-nightly", controller,
	); err != nil {
		__antithesis_instrumentation__.Notify(48722)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(48723)
	}
	__antithesis_instrumentation__.Notify(48711)

	if err := c.RunE(ctx, c.Node(1), "test -e jepsen_initialized"); err == nil {
		__antithesis_instrumentation__.Notify(48724)
		t.L().Printf("cluster already initialized\n")
		return
	} else {
		__antithesis_instrumentation__.Notify(48725)
	}
	__antithesis_instrumentation__.Notify(48712)
	t.L().Printf("initializing cluster\n")
	t.Status("initializing cluster")

	c.Run(ctx, c.All(), "mkdir", "-p", "logs")

	c.Run(ctx, c.All(), "sh", "-c", `"sudo apt-get -y update > logs/apt-upgrade.log 2>&1"`)
	c.Run(ctx, c.All(), "sh", "-c", `"sudo DEBIAN_FRONTEND=noninteractive apt-get -y upgrade -o Dpkg::Options::='--force-confold' -o DPkg::options::='--force-confdef' > logs/apt-upgrade.log 2>&1"`)

	c.Put(ctx, t.Cockroach(), "./cockroach", c.All())

	c.Run(ctx, c.All(), "tar --transform s,^,cockroach/, -c -z -f cockroach.tgz cockroach")

	if result, err := c.RunWithDetailsSingleNode(
		ctx, t.L(), controller, "sh", "-c",
		`"sudo DEBIAN_FRONTEND=noninteractive apt-get -qqy install openjdk-8-jre openjdk-8-jre-headless libjna-java gnuplot > /dev/null 2>&1"`,
	); err != nil {
		__antithesis_instrumentation__.Notify(48726)
		if result.RemoteExitStatus == "100" {
			__antithesis_instrumentation__.Notify(48728)
			t.Skip("apt-get failure (#31944)", result.Stdout+result.Stderr)
		} else {
			__antithesis_instrumentation__.Notify(48729)
		}
		__antithesis_instrumentation__.Notify(48727)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(48730)
	}
	__antithesis_instrumentation__.Notify(48713)

	c.Run(ctx, controller, "test -x lein || (curl -o lein https://raw.githubusercontent.com/technomancy/leiningen/stable/bin/lein && chmod +x lein)")

	tempDir, err := ioutil.TempDir("", "jepsen")
	if err != nil {
		__antithesis_instrumentation__.Notify(48731)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(48732)
	}
	__antithesis_instrumentation__.Notify(48714)
	c.Run(ctx, controller, "sh", "-c", `"test -f .ssh/id_rsa || ssh-keygen -f .ssh/id_rsa -t rsa -m pem -N ''"`)

	c.Run(ctx, controller, "sh", "-c", `"ssh-keygen -p -f .ssh/id_rsa -m pem -P '' -N ''"`)

	pubSSHKey := filepath.Join(tempDir, "id_rsa.pub")
	if err := c.Get(ctx, t.L(), ".ssh/id_rsa.pub", pubSSHKey, controller); err != nil {
		__antithesis_instrumentation__.Notify(48733)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(48734)
	}
	__antithesis_instrumentation__.Notify(48715)

	c.Put(ctx, pubSSHKey, "controller_id_rsa.pub", workers)
	c.Run(ctx, workers, "sh", "-c", `"cat controller_id_rsa.pub >> .ssh/authorized_keys"`)

	ips, err := c.InternalIP(ctx, t.L(), workers)
	if err != nil {
		__antithesis_instrumentation__.Notify(48735)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(48736)
	}
	__antithesis_instrumentation__.Notify(48716)
	for _, ip := range ips {
		__antithesis_instrumentation__.Notify(48737)
		c.Run(ctx, controller, "sh", "-c", fmt.Sprintf(`"ssh-keyscan -t rsa %s >> .ssh/known_hosts"`, ip))
	}
	__antithesis_instrumentation__.Notify(48717)

	t.L().Printf("cluster initialization complete\n")
	c.Run(ctx, c.Node(1), "touch jepsen_initialized")
}

func runJepsen(ctx context.Context, t test.Test, c cluster.Cluster, testName, nemesis string) {
	__antithesis_instrumentation__.Notify(48738)
	initJepsen(ctx, t, c)

	controller := c.Node(c.Spec().NodeCount)

	var nodeFlags []string
	ips, err := c.InternalIP(ctx, t.L(), c.Range(1, c.Spec().NodeCount-1))
	if err != nil {
		__antithesis_instrumentation__.Notify(48747)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(48748)
	}
	__antithesis_instrumentation__.Notify(48739)
	for _, ip := range ips {
		__antithesis_instrumentation__.Notify(48749)
		nodeFlags = append(nodeFlags, "-n "+ip)
	}
	__antithesis_instrumentation__.Notify(48740)
	nodesStr := strings.Join(nodeFlags, " ")

	runE := func(c cluster.Cluster, ctx context.Context, node option.NodeListOption, args ...string) error {
		__antithesis_instrumentation__.Notify(48750)
		if c.IsLocal() {
			__antithesis_instrumentation__.Notify(48752)

			t.L().Printf("> %s\n", strings.Join(args, " "))
			return nil
		} else {
			__antithesis_instrumentation__.Notify(48753)
		}
		__antithesis_instrumentation__.Notify(48751)
		return c.RunE(ctx, node, args...)
	}
	__antithesis_instrumentation__.Notify(48741)

	run := func(c cluster.Cluster, ctx context.Context, node option.NodeListOption, args ...string) {
		__antithesis_instrumentation__.Notify(48754)
		err := runE(c, ctx, node, args...)
		if err != nil {
			__antithesis_instrumentation__.Notify(48755)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48756)
		}
	}
	__antithesis_instrumentation__.Notify(48742)

	t.Status("running")
	run(c, ctx, controller, "rm -f /mnt/data1/jepsen/cockroachdb/store/latest")

	{
		__antithesis_instrumentation__.Notify(48757)
		err := runE(c, ctx, controller, "bash", "-e", "-c",
			`"cd /mnt/data1/jepsen/jepsen && ~/lein install && rm -f /mnt/data1/jepsen/cockroachdb/invoke.log"`)
		if err != nil {
			__antithesis_instrumentation__.Notify(48758)

			r := regexp.MustCompile("Could not transfer artifact|Failed to read artifact descriptor for")
			match := r.FindStringSubmatch(fmt.Sprintf("%+v", err))
			if match != nil {
				__antithesis_instrumentation__.Notify(48760)
				t.L().PrintfCtx(ctx, "failure installing deps (\"%s\")\nfull err: %+v",
					match, err)
				t.Skipf("failure installing deps (\"%s\"); in the past it's been transient", match)
			} else {
				__antithesis_instrumentation__.Notify(48761)
			}
			__antithesis_instrumentation__.Notify(48759)
			t.Fatalf("error installing Jepsen deps: %+v", err)
		} else {
			__antithesis_instrumentation__.Notify(48762)
		}
	}
	__antithesis_instrumentation__.Notify(48743)

	errCh := make(chan error, 1)
	go func() {
		__antithesis_instrumentation__.Notify(48763)
		errCh <- runE(c, ctx, controller, "bash", "-e", "-c", fmt.Sprintf(`"\
cd /mnt/data1/jepsen/cockroachdb && set -eo pipefail && \
 ~/lein run test \
   --tarball file://${PWD}/cockroach.tgz \
   --username ${USER} \
   --ssh-private-key ~/.ssh/id_rsa \
   --os ubuntu \
   --time-limit 300 \
   --concurrency 30 \
   --recovery-time 25 \
   --test-count 1 \
   %s \
   --test %s %s \
> invoke.log 2>&1 \
"`, nodesStr, testName, nemesis))
	}()
	__antithesis_instrumentation__.Notify(48744)

	outputDir := t.ArtifactsDir()
	if err := os.MkdirAll(outputDir, 0777); err != nil {
		__antithesis_instrumentation__.Notify(48764)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(48765)
	}
	__antithesis_instrumentation__.Notify(48745)
	var testErr error
	select {
	case testErr = <-errCh:
		__antithesis_instrumentation__.Notify(48766)
		if testErr == nil {
			__antithesis_instrumentation__.Notify(48768)
			t.L().Printf("passed, grabbing minimal logs")
		} else {
			__antithesis_instrumentation__.Notify(48769)
			t.L().Printf("failed: %s", testErr)
		}

	case <-time.After(40 * time.Minute):
		__antithesis_instrumentation__.Notify(48767)

		run(c, ctx, controller, "pkill -QUIT java")
		time.Sleep(10 * time.Second)
		run(c, ctx, controller, "pkill java")
		t.L().Printf("timed out")
		testErr = fmt.Errorf("timed out")
	}
	__antithesis_instrumentation__.Notify(48746)

	if testErr != nil {
		__antithesis_instrumentation__.Notify(48770)
		t.L().Printf("grabbing artifacts from controller. Tail of controller log:")
		run(c, ctx, controller, "tail -n 100 /mnt/data1/jepsen/cockroachdb/invoke.log")

		ignoreErr := false
		if err := runE(c, ctx, controller,
			`grep -E "(Oh jeez, I'm sorry, Jepsen broke. Here's why|Caused by)" /mnt/data1/jepsen/cockroachdb/invoke.log -A1 | grep `+
				`-e BrokenBarrierException `+
				`-e InterruptedException `+
				`-e ArrayIndexOutOfBoundsException `+
				`-e NullPointerException `+

				`-e "clojure.lang.ExceptionInfo: clj-ssh scp failure" `+

				`-e "Everything looks good" `+
				`-e "RuntimeException: Connection to"`,
		); err == nil {
			__antithesis_instrumentation__.Notify(48774)
			t.L().Printf("Recognized BrokenBarrier or other known exceptions (see grep output above). " +
				"Ignoring it and considering the test successful. " +
				"See #30527 or #26082 for some of the ignored exceptions.")
			ignoreErr = true
		} else {
			__antithesis_instrumentation__.Notify(48775)
		}
		__antithesis_instrumentation__.Notify(48771)

		if result, err := c.RunWithDetailsSingleNode(
			ctx, t.L(), controller,

			"tar -chj --ignore-failed-read -C /mnt/data1/jepsen/cockroachdb -f- store/latest invoke.log",
		); err != nil {
			__antithesis_instrumentation__.Notify(48776)
			t.L().Printf("failed to retrieve jepsen artifacts and invoke.log: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(48777)
			if err := ioutil.WriteFile(filepath.Join(outputDir, "failure-logs.tbz"), []byte(result.Stdout), 0666); err != nil {
				__antithesis_instrumentation__.Notify(48778)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(48779)
				t.L().Printf("downloaded jepsen logs in failure-logs.tbz")
			}
		}
		__antithesis_instrumentation__.Notify(48772)
		if ignoreErr {
			__antithesis_instrumentation__.Notify(48780)
			t.Skip("recognized known error", testErr.Error())
		} else {
			__antithesis_instrumentation__.Notify(48781)
		}
		__antithesis_instrumentation__.Notify(48773)
		t.Fatal(testErr)
	} else {
		__antithesis_instrumentation__.Notify(48782)
		collectFiles := []string{
			"test.fressian", "results.edn", "latency-quantiles.png", "latency-raw.png", "rate.png",
		}
		anyFailed := false
		for _, file := range collectFiles {
			__antithesis_instrumentation__.Notify(48784)
			if err := c.Get(
				ctx, t.L(),
				"/mnt/data1/jepsen/cockroachdb/store/latest/"+file,
				filepath.Join(outputDir, file),
				controller,
			); err != nil {
				__antithesis_instrumentation__.Notify(48785)
				anyFailed = true
				t.L().Printf("failed to retrieve %s: %s", file, err)
			} else {
				__antithesis_instrumentation__.Notify(48786)
			}
		}
		__antithesis_instrumentation__.Notify(48783)
		if anyFailed {
			__antithesis_instrumentation__.Notify(48787)

			if err := c.Get(ctx, t.L(),
				"/mnt/data1/jepsen/cockroachdb/invoke.log",
				filepath.Join(outputDir, "invoke.log"),
				controller,
			); err != nil {
				__antithesis_instrumentation__.Notify(48788)
				t.L().Printf("failed to retrieve invoke.log: %s", err)
			} else {
				__antithesis_instrumentation__.Notify(48789)
			}
		} else {
			__antithesis_instrumentation__.Notify(48790)
		}
	}
}

func RegisterJepsen(r registry.Registry) {
	__antithesis_instrumentation__.Notify(48791)

	tests := []string{
		"bank",
		"bank-multitable",
		"g2",
		"monotonic",
		"register",
		"sequential",
		"sets",
		"multi-register",
	}
	for _, testName := range tests {
		__antithesis_instrumentation__.Notify(48792)
		testName := testName
		for _, nemesis := range jepsenNemeses {
			__antithesis_instrumentation__.Notify(48793)
			nemesis := nemesis
			s := registry.TestSpec{
				Name: fmt.Sprintf("jepsen/%s/%s", testName, nemesis.name),

				Owner: registry.OwnerKV,

				Cluster: r.MakeClusterSpec(6, spec.ReuseTagged("jepsen")),
				Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
					__antithesis_instrumentation__.Notify(48795)
					runJepsen(ctx, t, c, testName, nemesis.config)
				},
			}
			__antithesis_instrumentation__.Notify(48794)
			r.Add(s)
		}
	}
}
