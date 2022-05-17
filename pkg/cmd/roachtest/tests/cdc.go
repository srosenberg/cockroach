package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	gosql "database/sql"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/url"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Shopify/sarama"
	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl"
	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/cdctest"
	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/changefeedbase"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/spec"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/roachprod/logger"
	"github.com/cockroachdb/cockroach/pkg/testutils/sqlutils"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/randutil"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/codahale/hdrhistogram"
)

type workloadType string

const (
	tpccWorkloadType   workloadType = "tpcc"
	ledgerWorkloadType workloadType = "ledger"
)

type sinkType int32

const (
	cloudStorageSink sinkType = iota + 1
	webhookSink
	pubsubSink
)

type cdcTestArgs struct {
	workloadType       workloadType
	tpccWarehouseCount int
	workloadDuration   string
	initialScan        bool
	kafkaChaos         bool
	crdbChaos          bool
	whichSink          sinkType
	sinkURI            string

	preStartStatements []string

	targetInitialScanLatency time.Duration
	targetSteadyLatency      time.Duration
	targetTxnPerSecond       float64
}

func cdcClusterSettings(t test.Test, db *sqlutils.SQLRunner) {
	__antithesis_instrumentation__.Notify(46047)

	db.Exec(t, "SET CLUSTER SETTING kv.rangefeed.enabled = true")
	randomlyRun(t, db, "SET CLUSTER SETTING kv.rangefeed.catchup_scan_iterator_optimization.enabled = false")
}

const randomSettingPercent = 0.50

var rng, _ = randutil.NewTestRand()

func randomlyRun(t test.Test, db *sqlutils.SQLRunner, query string) {
	__antithesis_instrumentation__.Notify(46048)
	if rng.Float64() < randomSettingPercent {
		__antithesis_instrumentation__.Notify(46049)
		db.Exec(t, query)
		t.L().Printf("setting non-default cluster setting: %s", query)
	} else {
		__antithesis_instrumentation__.Notify(46050)
	}

}

func cdcBasicTest(ctx context.Context, t test.Test, c cluster.Cluster, args cdcTestArgs) {
	__antithesis_instrumentation__.Notify(46051)
	crdbNodes := c.Range(1, c.Spec().NodeCount-1)
	workloadNode := c.Node(c.Spec().NodeCount)
	kafkaNode := c.Node(c.Spec().NodeCount)
	c.Put(ctx, t.Cockroach(), "./cockroach")
	c.Put(ctx, t.DeprecatedWorkload(), "./workload", workloadNode)
	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), crdbNodes)

	db := c.Conn(ctx, t.L(), 1)
	defer stopFeeds(db)
	tdb := sqlutils.MakeSQLRunner(db)
	cdcClusterSettings(t, tdb)
	kafka := kafkaManager{
		t:     t,
		c:     c,
		nodes: kafkaNode,
	}

	var sinkURI string
	if args.sinkURI != "" {
		__antithesis_instrumentation__.Notify(46058)
		sinkURI = args.sinkURI
	} else {
		__antithesis_instrumentation__.Notify(46059)
		if args.whichSink == cloudStorageSink {
			__antithesis_instrumentation__.Notify(46060)
			ts := timeutil.Now().Format(`20060102150405`)

			sinkURI = `experimental-gs://cockroach-tmp/roachtest/` + ts + "?AUTH=implicit"
		} else {
			__antithesis_instrumentation__.Notify(46061)
			if args.whichSink == webhookSink {
				__antithesis_instrumentation__.Notify(46062)

				cert, certEncoded, err := cdctest.NewCACertBase64Encoded()
				if err != nil {
					__antithesis_instrumentation__.Notify(46066)
					t.Fatal(err)
				} else {
					__antithesis_instrumentation__.Notify(46067)
				}
				__antithesis_instrumentation__.Notify(46063)
				sinkDest, err := cdctest.StartMockWebhookSink(cert)
				if err != nil {
					__antithesis_instrumentation__.Notify(46068)
					t.Fatal(err)
				} else {
					__antithesis_instrumentation__.Notify(46069)
				}
				__antithesis_instrumentation__.Notify(46064)

				sinkDestHost, err := url.Parse(sinkDest.URL())
				if err != nil {
					__antithesis_instrumentation__.Notify(46070)
					t.Fatal(err)
				} else {
					__antithesis_instrumentation__.Notify(46071)
				}
				__antithesis_instrumentation__.Notify(46065)

				params := sinkDestHost.Query()
				params.Set(changefeedbase.SinkParamCACert, certEncoded)
				sinkDestHost.RawQuery = params.Encode()

				sinkURI = fmt.Sprintf("webhook-%s", sinkDestHost.String())
			} else {
				__antithesis_instrumentation__.Notify(46072)
				if args.whichSink == pubsubSink {
					__antithesis_instrumentation__.Notify(46073)
					sinkURI = changefeedccl.GcpScheme + `://cockroach-ephemeral` + "?AUTH=implicit&topic_name=pubsubSink-roachtest&region=us-east1"
				} else {
					__antithesis_instrumentation__.Notify(46074)
					t.Status("installing kafka")
					kafka.install(ctx)
					kafka.start(ctx)
					sinkURI = kafka.sinkURL(ctx)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(46052)

	m := c.NewMonitor(ctx, crdbNodes)
	workloadCompleteCh := make(chan struct{}, 1)

	workloadStart := timeutil.Now()
	if args.workloadType == tpccWorkloadType {
		__antithesis_instrumentation__.Notify(46075)
		t.Status("installing TPCC")
		tpcc := tpccWorkload{
			sqlNodes:           crdbNodes,
			workloadNodes:      workloadNode,
			tpccWarehouseCount: args.tpccWarehouseCount,

			tolerateErrors: args.crdbChaos,
		}

		tpcc.install(ctx, c)

		time.Sleep(2 * time.Second)
		t.Status("initiating workload")
		m.Go(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(46076)
			defer func() { __antithesis_instrumentation__.Notify(46078); close(workloadCompleteCh) }()
			__antithesis_instrumentation__.Notify(46077)
			tpcc.run(ctx, c, args.workloadDuration)
			return nil
		})
	} else {
		__antithesis_instrumentation__.Notify(46079)
		t.Status("installing Ledger Workload")
		lw := ledgerWorkload{
			sqlNodes:      crdbNodes,
			workloadNodes: workloadNode,
		}
		lw.install(ctx, c)

		t.Status("initiating workload")
		m.Go(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(46080)
			defer func() { __antithesis_instrumentation__.Notify(46082); close(workloadCompleteCh) }()
			__antithesis_instrumentation__.Notify(46081)
			lw.run(ctx, c, args.workloadDuration)
			return nil
		})
	}
	__antithesis_instrumentation__.Notify(46053)

	changefeedLogger, err := t.L().ChildLogger("changefeed")
	if err != nil {
		__antithesis_instrumentation__.Notify(46083)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(46084)
	}
	__antithesis_instrumentation__.Notify(46054)
	defer changefeedLogger.Close()
	verifier := makeLatencyVerifier(
		args.targetInitialScanLatency,
		args.targetSteadyLatency,
		changefeedLogger,
		t.Status,
		args.crdbChaos,
	)
	defer verifier.maybeLogLatencyHist()

	m.Go(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(46085)

		if _, err := db.Exec(
			`SET CLUSTER SETTING kv.closed_timestamp.target_duration='10s'`,
		); err != nil {
			__antithesis_instrumentation__.Notify(46092)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(46093)
		}
		__antithesis_instrumentation__.Notify(46086)

		if _, err := db.Exec(
			`SET CLUSTER SETTING changefeed.slow_span_log_threshold='30s'`,
		); err != nil {
			__antithesis_instrumentation__.Notify(46094)

			t.L().Printf("failed to set cluster setting: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(46095)
		}
		__antithesis_instrumentation__.Notify(46087)

		for _, stmt := range args.preStartStatements {
			__antithesis_instrumentation__.Notify(46096)
			_, err := db.ExecContext(ctx, stmt)
			if err != nil {
				__antithesis_instrumentation__.Notify(46097)
				t.Fatalf("failed pre-start statement %q: %s", stmt, err.Error())
			} else {
				__antithesis_instrumentation__.Notify(46098)
			}
		}
		__antithesis_instrumentation__.Notify(46088)

		var targets string
		if args.workloadType == tpccWorkloadType {
			__antithesis_instrumentation__.Notify(46099)
			targets = `tpcc.warehouse, tpcc.district, tpcc.customer, tpcc.history,
			tpcc.order, tpcc.new_order, tpcc.item, tpcc.stock,
			tpcc.order_line`
		} else {
			__antithesis_instrumentation__.Notify(46100)
			targets = `ledger.customer, ledger.transaction, ledger.entry, ledger.session`
		}
		__antithesis_instrumentation__.Notify(46089)

		jobID, err := createChangefeed(db, targets, sinkURI, args)
		if err != nil {
			__antithesis_instrumentation__.Notify(46101)
			return err
		} else {
			__antithesis_instrumentation__.Notify(46102)
		}
		__antithesis_instrumentation__.Notify(46090)

		info, err := getChangefeedInfo(db, jobID)
		if err != nil {
			__antithesis_instrumentation__.Notify(46103)
			return err
		} else {
			__antithesis_instrumentation__.Notify(46104)
		}
		__antithesis_instrumentation__.Notify(46091)
		verifier.statementTime = info.statementTime
		changefeedLogger.Printf("started changefeed at (%d) %s\n",
			verifier.statementTime.UnixNano(), verifier.statementTime)
		t.Status("watching changefeed")
		return verifier.pollLatency(ctx, db, jobID, time.Second, workloadCompleteCh)
	})
	__antithesis_instrumentation__.Notify(46055)

	if args.kafkaChaos {
		__antithesis_instrumentation__.Notify(46105)
		m.Go(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(46106)
			period, downTime := 2*time.Minute, 20*time.Second
			return kafka.chaosLoop(ctx, period, downTime, workloadCompleteCh)
		})
	} else {
		__antithesis_instrumentation__.Notify(46107)
	}
	__antithesis_instrumentation__.Notify(46056)

	if args.crdbChaos {
		__antithesis_instrumentation__.Notify(46108)
		chaosDuration, err := time.ParseDuration(args.workloadDuration)
		if err != nil {
			__antithesis_instrumentation__.Notify(46110)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(46111)
		}
		__antithesis_instrumentation__.Notify(46109)
		ch := Chaos{
			Timer:   Periodic{Period: 2 * time.Minute, DownTime: 20 * time.Second},
			Target:  crdbNodes.RandNode,
			Stopper: time.After(chaosDuration),
		}
		m.Go(ch.Runner(c, t, m))
	} else {
		__antithesis_instrumentation__.Notify(46112)
	}
	__antithesis_instrumentation__.Notify(46057)
	m.Wait()

	verifier.assertValid(t)
	workloadEnd := timeutil.Now()
	if args.targetTxnPerSecond > 0.0 {
		__antithesis_instrumentation__.Notify(46113)
		verifyTxnPerSecond(
			ctx, c, t, crdbNodes.RandNode(), workloadStart, workloadEnd, args.targetTxnPerSecond, 0.05,
		)
	} else {
		__antithesis_instrumentation__.Notify(46114)
	}
}

func runCDCBank(ctx context.Context, t test.Test, c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(46115)

	c.Run(ctx, c.All(), `mkdir -p logs`)

	crdbNodes, workloadNode, kafkaNode := c.Range(1, c.Spec().NodeCount-1), c.Node(c.Spec().NodeCount), c.Node(c.Spec().NodeCount)
	c.Put(ctx, t.Cockroach(), "./cockroach", crdbNodes)
	c.Put(ctx, t.DeprecatedWorkload(), "./workload", workloadNode)
	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), crdbNodes)
	kafka := kafkaManager{
		t:     t,
		c:     c,
		nodes: kafkaNode,
	}
	kafka.install(ctx)
	if !c.IsLocal() {
		__antithesis_instrumentation__.Notify(46125)

		c.Run(ctx, kafka.nodes, `echo "advertised.listeners=PLAINTEXT://`+kafka.consumerURL(ctx)+`" >> `+
			filepath.Join(kafka.configDir(), "server.properties"))
	} else {
		__antithesis_instrumentation__.Notify(46126)
	}
	__antithesis_instrumentation__.Notify(46116)
	kafka.start(ctx, "kafka")
	defer kafka.stop(ctx)

	t.Status("creating kafka topic")
	if err := kafka.createTopic(ctx, "bank"); err != nil {
		__antithesis_instrumentation__.Notify(46127)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(46128)
	}
	__antithesis_instrumentation__.Notify(46117)

	c.Run(ctx, workloadNode, `./workload init bank {pgurl:1}`)
	db := c.Conn(ctx, t.L(), 1)
	defer stopFeeds(db)

	tdb := sqlutils.MakeSQLRunner(db)
	cdcClusterSettings(t, tdb)
	tdb.Exec(t, `SET CLUSTER SETTING changefeed.experimental_poll_interval = '10ms'`)
	tdb.Exec(t, `SET CLUSTER SETTING kv.closed_timestamp.target_duration = '1s'`)

	withDiff := t.IsBuildVersion("v20.1.0")
	var opts = []string{`updated`, `resolved`}
	if withDiff {
		__antithesis_instrumentation__.Notify(46129)
		opts = append(opts, `diff`)
	} else {
		__antithesis_instrumentation__.Notify(46130)
	}
	__antithesis_instrumentation__.Notify(46118)
	var jobID string
	if err := db.QueryRow(
		`CREATE CHANGEFEED FOR bank.bank INTO $1 WITH `+strings.Join(opts, `, `), kafka.sinkURL(ctx),
	).Scan(&jobID); err != nil {
		__antithesis_instrumentation__.Notify(46131)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(46132)
	}
	__antithesis_instrumentation__.Notify(46119)

	tc, err := kafka.consumer(ctx, "bank")
	if err != nil {
		__antithesis_instrumentation__.Notify(46133)
		t.Fatal(errors.Wrap(err, "could not create kafka consumer"))
	} else {
		__antithesis_instrumentation__.Notify(46134)
	}
	__antithesis_instrumentation__.Notify(46120)
	defer tc.Close()

	l, err := t.L().ChildLogger(`changefeed`)
	if err != nil {
		__antithesis_instrumentation__.Notify(46135)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(46136)
	}
	__antithesis_instrumentation__.Notify(46121)
	defer l.Close()

	t.Status("running workload")
	workloadCtx, workloadCancel := context.WithCancel(ctx)
	defer workloadCancel()

	m := c.NewMonitor(workloadCtx, crdbNodes)
	var doneAtomic int64
	messageBuf := make(chan *sarama.ConsumerMessage, 4096)
	const requestedResolved = 100

	m.Go(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(46137)
		err := c.RunE(ctx, workloadNode, `./workload run bank {pgurl:1} --max-rate=10`)
		if atomic.LoadInt64(&doneAtomic) > 0 {
			__antithesis_instrumentation__.Notify(46139)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(46140)
		}
		__antithesis_instrumentation__.Notify(46138)
		return errors.Wrap(err, "workload failed")
	})
	__antithesis_instrumentation__.Notify(46122)
	m.Go(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(46141)
		defer workloadCancel()
		defer func() { __antithesis_instrumentation__.Notify(46144); close(messageBuf) }()
		__antithesis_instrumentation__.Notify(46142)
		v := cdctest.MakeCountValidator(cdctest.NoOpValidator)
		for {
			__antithesis_instrumentation__.Notify(46145)
			m := tc.Next(ctx)
			if m == nil {
				__antithesis_instrumentation__.Notify(46148)
				return fmt.Errorf("unexpected end of changefeed")
			} else {
				__antithesis_instrumentation__.Notify(46149)
			}
			__antithesis_instrumentation__.Notify(46146)
			messageBuf <- m
			updated, resolved, err := cdctest.ParseJSONValueTimestamps(m.Value)
			if err != nil {
				__antithesis_instrumentation__.Notify(46150)
				return err
			} else {
				__antithesis_instrumentation__.Notify(46151)
			}
			__antithesis_instrumentation__.Notify(46147)

			partitionStr := strconv.Itoa(int(m.Partition))
			if len(m.Key) > 0 {
				__antithesis_instrumentation__.Notify(46152)
				if err := v.NoteRow(partitionStr, string(m.Key), string(m.Value), updated); err != nil {
					__antithesis_instrumentation__.Notify(46153)
					return err
				} else {
					__antithesis_instrumentation__.Notify(46154)
				}
			} else {
				__antithesis_instrumentation__.Notify(46155)
				if err := v.NoteResolved(partitionStr, resolved); err != nil {
					__antithesis_instrumentation__.Notify(46157)
					return err
				} else {
					__antithesis_instrumentation__.Notify(46158)
				}
				__antithesis_instrumentation__.Notify(46156)
				l.Printf("%d of %d resolved timestamps received from kafka, latest is %s behind realtime, %s beind realtime when sent to kafka",
					v.NumResolvedWithRows, requestedResolved, timeutil.Since(resolved.GoTime()), m.Timestamp.Sub(resolved.GoTime()))
				if v.NumResolvedWithRows >= requestedResolved {
					__antithesis_instrumentation__.Notify(46159)
					atomic.StoreInt64(&doneAtomic, 1)
					break
				} else {
					__antithesis_instrumentation__.Notify(46160)
				}
			}
		}
		__antithesis_instrumentation__.Notify(46143)
		return nil
	})
	__antithesis_instrumentation__.Notify(46123)
	m.Go(func(context.Context) error {
		__antithesis_instrumentation__.Notify(46161)
		if _, err := db.Exec(
			`CREATE TABLE fprint (id INT PRIMARY KEY, balance INT, payload STRING)`,
		); err != nil {
			__antithesis_instrumentation__.Notify(46167)
			return errors.Wrap(err, "CREATE TABLE failed")
		} else {
			__antithesis_instrumentation__.Notify(46168)
		}
		__antithesis_instrumentation__.Notify(46162)

		fprintV, err := cdctest.NewFingerprintValidator(db, `bank.bank`, `fprint`, tc.partitions, 0)
		if err != nil {
			__antithesis_instrumentation__.Notify(46169)
			return errors.Wrap(err, "error creating validator")
		} else {
			__antithesis_instrumentation__.Notify(46170)
		}
		__antithesis_instrumentation__.Notify(46163)
		validators := cdctest.Validators{
			cdctest.NewOrderValidator(`bank`),
			fprintV,
		}
		if withDiff {
			__antithesis_instrumentation__.Notify(46171)
			baV, err := cdctest.NewBeforeAfterValidator(db, `bank.bank`)
			if err != nil {
				__antithesis_instrumentation__.Notify(46173)
				return err
			} else {
				__antithesis_instrumentation__.Notify(46174)
			}
			__antithesis_instrumentation__.Notify(46172)
			validators = append(validators, baV)
		} else {
			__antithesis_instrumentation__.Notify(46175)
		}
		__antithesis_instrumentation__.Notify(46164)
		v := cdctest.MakeCountValidator(validators)
		for {
			__antithesis_instrumentation__.Notify(46176)
			m, ok := <-messageBuf
			if !ok {
				__antithesis_instrumentation__.Notify(46179)
				break
			} else {
				__antithesis_instrumentation__.Notify(46180)
			}
			__antithesis_instrumentation__.Notify(46177)
			updated, resolved, err := cdctest.ParseJSONValueTimestamps(m.Value)
			if err != nil {
				__antithesis_instrumentation__.Notify(46181)
				return err
			} else {
				__antithesis_instrumentation__.Notify(46182)
			}
			__antithesis_instrumentation__.Notify(46178)

			partitionStr := strconv.Itoa(int(m.Partition))
			if len(m.Key) > 0 {
				__antithesis_instrumentation__.Notify(46183)
				if err := v.NoteRow(partitionStr, string(m.Key), string(m.Value), updated); err != nil {
					__antithesis_instrumentation__.Notify(46184)
					return err
				} else {
					__antithesis_instrumentation__.Notify(46185)
				}
			} else {
				__antithesis_instrumentation__.Notify(46186)
				if err := v.NoteResolved(partitionStr, resolved); err != nil {
					__antithesis_instrumentation__.Notify(46188)
					return err
				} else {
					__antithesis_instrumentation__.Notify(46189)
				}
				__antithesis_instrumentation__.Notify(46187)
				l.Printf("%d of %d resolved timestamps validated, latest is %s behind realtime",
					v.NumResolvedWithRows, requestedResolved, timeutil.Since(resolved.GoTime()))

			}
		}
		__antithesis_instrumentation__.Notify(46165)
		if failures := v.Failures(); len(failures) > 0 {
			__antithesis_instrumentation__.Notify(46190)
			return errors.Newf("validator failures:\n%s", strings.Join(failures, "\n"))
		} else {
			__antithesis_instrumentation__.Notify(46191)
		}
		__antithesis_instrumentation__.Notify(46166)
		return nil
	})
	__antithesis_instrumentation__.Notify(46124)
	m.Wait()
}

func runCDCSchemaRegistry(ctx context.Context, t test.Test, c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(46192)

	crdbNodes, kafkaNode := c.Node(1), c.Node(1)
	c.Put(ctx, t.Cockroach(), "./cockroach", crdbNodes)
	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), crdbNodes)
	kafka := kafkaManager{
		t:     t,
		c:     c,
		nodes: kafkaNode,
	}
	kafka.install(ctx)
	kafka.start(ctx)
	defer kafka.stop(ctx)

	db := c.Conn(ctx, t.L(), 1)
	defer stopFeeds(db)

	cdcClusterSettings(t, sqlutils.MakeSQLRunner(db))

	if _, err := db.Exec(`CREATE TABLE foo (a INT PRIMARY KEY)`); err != nil {
		__antithesis_instrumentation__.Notify(46207)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(46208)
	}
	__antithesis_instrumentation__.Notify(46193)

	withDiff := t.IsBuildVersion("v20.1.0")
	var opts = []string{`updated`, `resolved`, `format=experimental_avro`, `confluent_schema_registry=$2`}
	if withDiff {
		__antithesis_instrumentation__.Notify(46209)
		opts = append(opts, `diff`)
	} else {
		__antithesis_instrumentation__.Notify(46210)
	}
	__antithesis_instrumentation__.Notify(46194)
	var jobID string
	if err := db.QueryRow(
		`CREATE CHANGEFEED FOR foo INTO $1 WITH `+strings.Join(opts, `, `),
		kafka.sinkURL(ctx), kafka.schemaRegistryURL(ctx),
	).Scan(&jobID); err != nil {
		__antithesis_instrumentation__.Notify(46211)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(46212)
	}
	__antithesis_instrumentation__.Notify(46195)

	if _, err := db.Exec(`INSERT INTO foo VALUES (1)`); err != nil {
		__antithesis_instrumentation__.Notify(46213)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(46214)
	}
	__antithesis_instrumentation__.Notify(46196)
	if _, err := db.Exec(`ALTER TABLE foo ADD COLUMN b STRING`); err != nil {
		__antithesis_instrumentation__.Notify(46215)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(46216)
	}
	__antithesis_instrumentation__.Notify(46197)
	if _, err := db.Exec(`INSERT INTO foo VALUES (2, '2')`); err != nil {
		__antithesis_instrumentation__.Notify(46217)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(46218)
	}
	__antithesis_instrumentation__.Notify(46198)
	if _, err := db.Exec(`ALTER TABLE foo ADD COLUMN c INT`); err != nil {
		__antithesis_instrumentation__.Notify(46219)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(46220)
	}
	__antithesis_instrumentation__.Notify(46199)
	if _, err := db.Exec(`INSERT INTO foo VALUES (3, '3', 3)`); err != nil {
		__antithesis_instrumentation__.Notify(46221)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(46222)
	}
	__antithesis_instrumentation__.Notify(46200)
	if _, err := db.Exec(`ALTER TABLE foo DROP COLUMN b`); err != nil {
		__antithesis_instrumentation__.Notify(46223)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(46224)
	}
	__antithesis_instrumentation__.Notify(46201)
	if _, err := db.Exec(`INSERT INTO foo VALUES (4, 4)`); err != nil {
		__antithesis_instrumentation__.Notify(46225)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(46226)
	}
	__antithesis_instrumentation__.Notify(46202)

	updatedRE := regexp.MustCompile(`"updated":\{"string":"[^"]+"\}`)
	updatedMap := make(map[string]struct{})
	var resolved []string
	pagesFetched := 0
	pageSize := 14

	for len(updatedMap) < 10 && func() bool {
		__antithesis_instrumentation__.Notify(46227)
		return pagesFetched < 5 == true
	}() == true {
		__antithesis_instrumentation__.Notify(46228)
		result, err := c.RunWithDetailsSingleNode(ctx, t.L(), kafkaNode,
			kafka.makeCommand("kafka-avro-console-consumer",
				fmt.Sprintf("--offset=%d", pagesFetched*pageSize),
				"--partition=0",
				"--topic=foo",
				fmt.Sprintf("--max-messages=%d", pageSize),
				"--bootstrap-server=localhost:9092"))
		t.L().Printf("\n%s\n", result.Stdout+result.Stderr)
		if err != nil {
			__antithesis_instrumentation__.Notify(46230)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(46231)
		}
		__antithesis_instrumentation__.Notify(46229)
		pagesFetched++

		for _, line := range strings.Split(result.Stdout, "\n") {
			__antithesis_instrumentation__.Notify(46232)
			if strings.Contains(line, `"updated"`) {
				__antithesis_instrumentation__.Notify(46233)
				line = updatedRE.ReplaceAllString(line, `"updated":{"string":""}`)
				updatedMap[line] = struct{}{}
			} else {
				__antithesis_instrumentation__.Notify(46234)
				if strings.Contains(line, `"resolved"`) {
					__antithesis_instrumentation__.Notify(46235)
					resolved = append(resolved, line)
				} else {
					__antithesis_instrumentation__.Notify(46236)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(46203)

	updated := make([]string, 0, len(updatedMap))
	for u := range updatedMap {
		__antithesis_instrumentation__.Notify(46237)
		updated = append(updated, u)
	}
	__antithesis_instrumentation__.Notify(46204)
	sort.Strings(updated)

	var expected []string
	if withDiff {
		__antithesis_instrumentation__.Notify(46238)
		expected = []string{
			`{"before":null,"after":{"foo":{"a":{"long":1}}},"updated":{"string":""}}`,
			`{"before":null,"after":{"foo":{"a":{"long":2},"b":{"string":"2"}}},"updated":{"string":""}}`,
			`{"before":null,"after":{"foo":{"a":{"long":3},"b":{"string":"3"},"c":{"long":3}}},"updated":{"string":""}}`,
			`{"before":null,"after":{"foo":{"a":{"long":4},"c":{"long":4}}},"updated":{"string":""}}`,
			`{"before":{"foo_before":{"a":{"long":1},"b":null,"c":null}},"after":{"foo":{"a":{"long":1},"c":null}},"updated":{"string":""}}`,
			`{"before":{"foo_before":{"a":{"long":1},"c":null}},"after":{"foo":{"a":{"long":1},"c":null}},"updated":{"string":""}}`,
			`{"before":{"foo_before":{"a":{"long":2},"b":{"string":"2"},"c":null}},"after":{"foo":{"a":{"long":2},"c":null}},"updated":{"string":""}}`,
			`{"before":{"foo_before":{"a":{"long":2},"c":null}},"after":{"foo":{"a":{"long":2},"c":null}},"updated":{"string":""}}`,
			`{"before":{"foo_before":{"a":{"long":3},"b":{"string":"3"},"c":{"long":3}}},"after":{"foo":{"a":{"long":3},"c":{"long":3}}},"updated":{"string":""}}`,
			`{"before":{"foo_before":{"a":{"long":3},"c":{"long":3}}},"after":{"foo":{"a":{"long":3},"c":{"long":3}}},"updated":{"string":""}}`,
		}
	} else {
		__antithesis_instrumentation__.Notify(46239)
		expected = []string{
			`{"updated":{"string":""},"after":{"foo":{"a":{"long":1},"c":null}}}`,
			`{"updated":{"string":""},"after":{"foo":{"a":{"long":1}}}}`,
			`{"updated":{"string":""},"after":{"foo":{"a":{"long":2},"b":{"string":"2"}}}}`,
			`{"updated":{"string":""},"after":{"foo":{"a":{"long":2},"c":null}}}`,
			`{"updated":{"string":""},"after":{"foo":{"a":{"long":3},"b":{"string":"3"},"c":{"long":3}}}}`,
			`{"updated":{"string":""},"after":{"foo":{"a":{"long":3},"c":{"long":3}}}}`,
			`{"updated":{"string":""},"after":{"foo":{"a":{"long":4},"c":{"long":4}}}}`,
		}
	}
	__antithesis_instrumentation__.Notify(46205)
	if strings.Join(expected, "\n") != strings.Join(updated, "\n") {
		__antithesis_instrumentation__.Notify(46240)
		t.Fatalf("expected\n%s\n\ngot\n%s\n\n",
			strings.Join(expected, "\n"), strings.Join(updated, "\n"))
	} else {
		__antithesis_instrumentation__.Notify(46241)
	}
	__antithesis_instrumentation__.Notify(46206)

	if len(resolved) == 0 {
		__antithesis_instrumentation__.Notify(46242)
		t.Fatal(`expected at least 1 resolved timestamp`)
	} else {
		__antithesis_instrumentation__.Notify(46243)
	}
}

func runCDCKafkaAuth(ctx context.Context, t test.Test, c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(46244)
	lastCrdbNode := c.Spec().NodeCount - 1
	if lastCrdbNode == 0 {
		__antithesis_instrumentation__.Notify(46246)
		lastCrdbNode = 1
	} else {
		__antithesis_instrumentation__.Notify(46247)
	}
	__antithesis_instrumentation__.Notify(46245)

	crdbNodes, kafkaNode := c.Range(1, lastCrdbNode), c.Node(c.Spec().NodeCount)
	c.Put(ctx, t.Cockroach(), "./cockroach", crdbNodes)
	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), crdbNodes)

	kafka := kafkaManager{
		t:     t,
		c:     c,
		nodes: kafkaNode,
	}
	kafka.install(ctx)
	testCerts := kafka.configureAuth(ctx)
	kafka.start(ctx, "kafka")
	kafka.addSCRAMUsers(ctx)
	defer kafka.stop(ctx)

	db := c.Conn(ctx, t.L(), 1)
	defer stopFeeds(db)

	tdb := sqlutils.MakeSQLRunner(db)
	cdcClusterSettings(t, tdb)
	tdb.Exec(t, `SET CLUSTER SETTING changefeed.experimental_poll_interval = '10ms'`)
	tdb.Exec(t, `SET CLUSTER SETTING kv.closed_timestamp.target_duration = '1s'`)

	tdb.Exec(t, `CREATE TABLE auth_test_table (a INT PRIMARY KEY)`)

	caCert := testCerts.CACertBase64()
	saslURL := kafka.sinkURLSASL(ctx)
	feeds := []struct {
		desc     string
		queryArg string
	}{
		{
			"create changefeed with insecure TLS transport and no auth",
			fmt.Sprintf("%s?tls_enabled=true&insecure_tls_skip_verify=true", kafka.sinkURLTLS(ctx)),
		},
		{
			"create changefeed with TLS transport and no auth",
			fmt.Sprintf("%s?tls_enabled=true&ca_cert=%s", kafka.sinkURLTLS(ctx), testCerts.CACertBase64()),
		},
		{
			"create changefeed with TLS transport and SASL/PLAIN (default mechanism)",
			fmt.Sprintf("%s?tls_enabled=true&ca_cert=%s&sasl_enabled=true&sasl_user=plain&sasl_password=plain-secret", saslURL, caCert),
		},
		{
			"create changefeed with TLS transport and SASL/PLAIN (explicit mechanism)",
			fmt.Sprintf("%s?tls_enabled=true&ca_cert=%s&sasl_enabled=true&sasl_user=plain&sasl_password=plain-secret&sasl_mechanism=PLAIN", saslURL, caCert),
		},
		{
			"create changefeed with TLS transport and SASL/SCRAM-SHA-256",
			fmt.Sprintf("%s?tls_enabled=true&ca_cert=%s&sasl_enabled=true&sasl_user=scram256&sasl_password=scram256-secret&sasl_mechanism=SCRAM-SHA-256", saslURL, caCert),
		},
		{
			"create changefeed with TLS transport and SASL/SCRAM-SHA-512",
			fmt.Sprintf("%s?tls_enabled=true&ca_cert=%s&sasl_enabled=true&sasl_user=scram512&sasl_password=scram512-secret&sasl_mechanism=SCRAM-SHA-512", saslURL, caCert),
		},
	}

	var jobID int
	for _, f := range feeds {
		__antithesis_instrumentation__.Notify(46248)
		t.Status(f.desc)
		row := db.QueryRow(`CREATE CHANGEFEED FOR auth_test_table INTO $1`, f.queryArg)
		if err := row.Scan(&jobID); err != nil {
			__antithesis_instrumentation__.Notify(46249)
			t.Fatalf("%s: %s", f.desc, err.Error())
		} else {
			__antithesis_instrumentation__.Notify(46250)
		}
	}
}

func registerCDC(r registry.Registry) {
	__antithesis_instrumentation__.Notify(46251)
	r.Add(registry.TestSpec{
		Name:            "cdc/tpcc-1000",
		Owner:           registry.OwnerCDC,
		Cluster:         r.MakeClusterSpec(4, spec.CPU(16)),
		RequiresLicense: true,
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(46261)
			cdcBasicTest(ctx, t, c, cdcTestArgs{
				workloadType:             tpccWorkloadType,
				tpccWarehouseCount:       1000,
				workloadDuration:         "120m",
				targetInitialScanLatency: 3 * time.Minute,
				targetSteadyLatency:      10 * time.Minute,
			})
		},
	})
	__antithesis_instrumentation__.Notify(46252)
	r.Add(registry.TestSpec{
		Name:            "cdc/tpcc-1000/sink=null",
		Owner:           registry.OwnerCDC,
		Cluster:         r.MakeClusterSpec(4, spec.CPU(16)),
		Tags:            []string{"manual"},
		RequiresLicense: true,
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(46262)
			cdcBasicTest(ctx, t, c, cdcTestArgs{
				workloadType:             tpccWorkloadType,
				tpccWarehouseCount:       1000,
				workloadDuration:         "120m",
				targetInitialScanLatency: 3 * time.Minute,
				targetSteadyLatency:      10 * time.Minute,
				sinkURI:                  "null://",
			})
		},
	})
	__antithesis_instrumentation__.Notify(46253)
	r.Add(registry.TestSpec{
		Name:            "cdc/initial-scan",
		Owner:           registry.OwnerCDC,
		Cluster:         r.MakeClusterSpec(4, spec.CPU(16)),
		RequiresLicense: true,
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(46263)
			cdcBasicTest(ctx, t, c, cdcTestArgs{
				workloadType:             tpccWorkloadType,
				tpccWarehouseCount:       100,
				workloadDuration:         "30m",
				initialScan:              true,
				targetInitialScanLatency: 30 * time.Minute,
				targetSteadyLatency:      time.Minute,
			})
		},
	})
	__antithesis_instrumentation__.Notify(46254)
	r.Add(registry.TestSpec{
		Name:            "cdc/sink-chaos",
		Owner:           `cdc`,
		Cluster:         r.MakeClusterSpec(4, spec.CPU(16)),
		RequiresLicense: true,
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(46264)
			cdcBasicTest(ctx, t, c, cdcTestArgs{
				workloadType:             tpccWorkloadType,
				tpccWarehouseCount:       100,
				workloadDuration:         "30m",
				kafkaChaos:               true,
				targetInitialScanLatency: 3 * time.Minute,
				targetSteadyLatency:      5 * time.Minute,
			})
		},
	})
	__antithesis_instrumentation__.Notify(46255)
	r.Add(registry.TestSpec{
		Name:            "cdc/crdb-chaos",
		Owner:           `cdc`,
		Cluster:         r.MakeClusterSpec(4, spec.CPU(16)),
		RequiresLicense: true,
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(46265)
			cdcBasicTest(ctx, t, c, cdcTestArgs{
				workloadType:             tpccWorkloadType,
				tpccWarehouseCount:       100,
				workloadDuration:         "30m",
				crdbChaos:                true,
				targetInitialScanLatency: 3 * time.Minute,

				targetSteadyLatency: 5 * time.Minute,
			})
		},
	})
	__antithesis_instrumentation__.Notify(46256)
	r.Add(registry.TestSpec{
		Name:  "cdc/ledger",
		Owner: `cdc`,

		Cluster:         r.MakeClusterSpec(4, spec.CPU(16)),
		RequiresLicense: true,
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(46266)
			cdcBasicTest(ctx, t, c, cdcTestArgs{
				workloadType: ledgerWorkloadType,

				workloadDuration:         "28m",
				initialScan:              true,
				targetInitialScanLatency: 10 * time.Minute,
				targetSteadyLatency:      time.Minute,
				targetTxnPerSecond:       575,
				preStartStatements:       []string{"ALTER DATABASE ledger CONFIGURE ZONE USING range_max_bytes = 805306368, range_min_bytes = 134217728"},
			})
		},
	})
	__antithesis_instrumentation__.Notify(46257)
	r.Add(registry.TestSpec{
		Name:            "cdc/cloud-sink-gcs/rangefeed=true",
		Owner:           `cdc`,
		Cluster:         r.MakeClusterSpec(4, spec.CPU(16)),
		RequiresLicense: true,
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(46267)
			cdcBasicTest(ctx, t, c, cdcTestArgs{
				workloadType: tpccWorkloadType,

				tpccWarehouseCount:       50,
				workloadDuration:         "30m",
				initialScan:              true,
				whichSink:                cloudStorageSink,
				targetInitialScanLatency: 30 * time.Minute,
				targetSteadyLatency:      time.Minute,
			})
		},
	})
	__antithesis_instrumentation__.Notify(46258)

	r.Add(registry.TestSpec{
		Name:            "cdc/kafka-auth",
		Owner:           `cdc`,
		Cluster:         r.MakeClusterSpec(1),
		RequiresLicense: true,
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(46268)
			runCDCKafkaAuth(ctx, t, c)
		},
	})
	__antithesis_instrumentation__.Notify(46259)
	r.Add(registry.TestSpec{
		Name:            "cdc/bank",
		Skip:            "#72904",
		Owner:           `cdc`,
		Cluster:         r.MakeClusterSpec(4),
		RequiresLicense: true,
		Timeout:         30 * time.Minute,
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(46269)
			runCDCBank(ctx, t, c)
		},
	})
	__antithesis_instrumentation__.Notify(46260)
	r.Add(registry.TestSpec{
		Name:            "cdc/schemareg",
		Owner:           `cdc`,
		Cluster:         r.MakeClusterSpec(1),
		RequiresLicense: true,
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(46270)
			runCDCSchemaRegistry(ctx, t, c)
		},
	})
}

const (
	certLifetime = 30 * 24 * time.Hour
	keyLength    = 2048

	keystorePassword = "storepassword"
)

type testCerts struct {
	CACert    string
	CAKey     string
	KafkaCert string
	KafkaKey  string
}

func (t *testCerts) CACertBase64() string {
	__antithesis_instrumentation__.Notify(46271)
	return base64.StdEncoding.EncodeToString([]byte(t.CACert))
}

func makeTestCerts(kafkaNodeIP string) (*testCerts, error) {
	__antithesis_instrumentation__.Notify(46272)
	CAKey, err := rsa.GenerateKey(rand.Reader, keyLength)
	if err != nil {
		__antithesis_instrumentation__.Notify(46281)
		return nil, errors.Wrap(err, "CA private key")
	} else {
		__antithesis_instrumentation__.Notify(46282)
	}
	__antithesis_instrumentation__.Notify(46273)

	KafkaKey, err := rsa.GenerateKey(rand.Reader, keyLength)
	if err != nil {
		__antithesis_instrumentation__.Notify(46283)
		return nil, errors.Wrap(err, "kafka private key")
	} else {
		__antithesis_instrumentation__.Notify(46284)
	}
	__antithesis_instrumentation__.Notify(46274)

	CACert, CACertSpec, err := generateCACert(CAKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(46285)
		return nil, errors.Wrap(err, "CA cert gen")
	} else {
		__antithesis_instrumentation__.Notify(46286)
	}
	__antithesis_instrumentation__.Notify(46275)

	KafkaCert, err := generateKafkaCert(kafkaNodeIP, KafkaKey, CACertSpec, CAKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(46287)
		return nil, errors.Wrap(err, "kafka cert gen")
	} else {
		__antithesis_instrumentation__.Notify(46288)
	}
	__antithesis_instrumentation__.Notify(46276)

	CAKeyPEM, err := pemEncodePrivateKey(CAKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(46289)
		return nil, errors.Wrap(err, "pem encode CA key")
	} else {
		__antithesis_instrumentation__.Notify(46290)
	}
	__antithesis_instrumentation__.Notify(46277)

	CACertPEM, err := pemEncodeCert(CACert)
	if err != nil {
		__antithesis_instrumentation__.Notify(46291)
		return nil, errors.Wrap(err, "pem encode CA cert")
	} else {
		__antithesis_instrumentation__.Notify(46292)
	}
	__antithesis_instrumentation__.Notify(46278)

	KafkaKeyPEM, err := pemEncodePrivateKey(KafkaKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(46293)
		return nil, errors.Wrap(err, "pem encode kafka key")
	} else {
		__antithesis_instrumentation__.Notify(46294)
	}
	__antithesis_instrumentation__.Notify(46279)

	KafkaCertPEM, err := pemEncodeCert(KafkaCert)
	if err != nil {
		__antithesis_instrumentation__.Notify(46295)
		return nil, errors.Wrap(err, "pem encode kafka cert")
	} else {
		__antithesis_instrumentation__.Notify(46296)
	}
	__antithesis_instrumentation__.Notify(46280)

	return &testCerts{
		CACert:    CACertPEM,
		CAKey:     CAKeyPEM,
		KafkaCert: KafkaCertPEM,
		KafkaKey:  KafkaKeyPEM,
	}, nil
}

func generateKafkaCert(
	kafkaIP string, priv *rsa.PrivateKey, CACert *x509.Certificate, CAKey *rsa.PrivateKey,
) ([]byte, error) {
	__antithesis_instrumentation__.Notify(46297)
	ip := net.ParseIP(kafkaIP)
	if ip == nil {
		__antithesis_instrumentation__.Notify(46300)
		return nil, fmt.Errorf("invalid IP address: %s", kafkaIP)
	} else {
		__antithesis_instrumentation__.Notify(46301)
	}
	__antithesis_instrumentation__.Notify(46298)

	serial, err := randomSerial()
	if err != nil {
		__antithesis_instrumentation__.Notify(46302)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(46303)
	}
	__antithesis_instrumentation__.Notify(46299)

	certSpec := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Country:            []string{"US"},
			Organization:       []string{"Cockroach Labs"},
			OrganizationalUnit: []string{"Engineering"},
			CommonName:         "kafka-node",
		},
		NotBefore:   timeutil.Now(),
		NotAfter:    timeutil.Now().Add(certLifetime),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDataEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageKeyAgreement,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		DNSNames:    []string{"localhost"},
		IPAddresses: []net.IP{ip},
	}

	return x509.CreateCertificate(rand.Reader, certSpec, CACert, &priv.PublicKey, CAKey)
}

func generateCACert(priv *rsa.PrivateKey) ([]byte, *x509.Certificate, error) {
	__antithesis_instrumentation__.Notify(46304)
	serial, err := randomSerial()
	if err != nil {
		__antithesis_instrumentation__.Notify(46306)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(46307)
	}
	__antithesis_instrumentation__.Notify(46305)

	certSpec := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Country:            []string{"US"},
			Organization:       []string{"Cockroach Labs"},
			OrganizationalUnit: []string{"Engineering"},
			CommonName:         "Roachtest Temporary Insecure CA",
		},
		NotBefore:             timeutil.Now(),
		NotAfter:              timeutil.Now().Add(certLifetime),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		IsCA:                  true,
		BasicConstraintsValid: true,
		MaxPathLenZero:        true,
	}
	cert, err := x509.CreateCertificate(rand.Reader, certSpec, certSpec, &priv.PublicKey, priv)
	return cert, certSpec, err
}

func pemEncode(dataType string, data []byte) (string, error) {
	__antithesis_instrumentation__.Notify(46308)
	ret := new(strings.Builder)
	err := pem.Encode(ret, &pem.Block{Type: dataType, Bytes: data})
	if err != nil {
		__antithesis_instrumentation__.Notify(46310)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(46311)
	}
	__antithesis_instrumentation__.Notify(46309)

	return ret.String(), nil
}

func pemEncodePrivateKey(key *rsa.PrivateKey) (string, error) {
	__antithesis_instrumentation__.Notify(46312)
	return pemEncode("RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(key))
}

func pemEncodeCert(cert []byte) (string, error) {
	__antithesis_instrumentation__.Notify(46313)
	return pemEncode("CERTIFICATE", cert)
}

func randomSerial() (*big.Int, error) {
	__antithesis_instrumentation__.Notify(46314)
	limit := new(big.Int).Lsh(big.NewInt(1), 128)
	ret, err := rand.Int(rand.Reader, limit)
	if err != nil {
		__antithesis_instrumentation__.Notify(46316)
		return nil, errors.Wrap(err, "generate random serial")
	} else {
		__antithesis_instrumentation__.Notify(46317)
	}
	__antithesis_instrumentation__.Notify(46315)
	return ret, nil
}

const (
	confluentDownloadURL = "https://storage.googleapis.com/cockroach-fixtures/tools/confluent-community-6.1.0.tar.gz"
	confluentSHA256      = "53b0e2f08c4cfc55087fa5c9120a614ef04d306db6ec3bcd7710f89f05355355"
	confluentInstallBase = "confluent-6.1.0"

	confluentCLIVersion         = "1.26.0"
	confluentCLIDownloadURLBase = "https://s3-us-west-2.amazonaws.com/confluent.cloud/confluent-cli/archives"
)

var confluentDownloadScript = fmt.Sprintf(`#!/usr/bin/env bash
set -euo pipefail

CONFLUENT_URL="%s"
CONFLUENT_SHA256="%s"
CONFLUENT_INSTALL_BASE="%s"

CONFLUENT_CLI_VERSION="%s"
CONFLUENT_CLI_URL_BASE="%s"


CONFLUENT_CLI_TAR_PATH="/tmp/confluent-cli-$CONFLUENT_CLI_VERSION.tar.gz"
CONFLUENT_TAR_PATH=/tmp/confluent.tar.gz

CONFLUENT_DIR="$1"

os() {
  uname -s | tr '[:upper:]' '[:lower:]'
}

arch() {
  local arch
  arch=$(uname -m)
  case "$arch" in
    x86_64)
      echo "amd64"
      ;;
    *)
      echo "$arch"
      ;;
  esac
}

checkFile() {
  local file_name="${1}"
  local expected_shasum="${2}"

  local actual_shasum=""
  if command -v sha256sum > /dev/null 2>&1; then
    actual_shasum=$(sha256sum "$file_name" | cut -f1 -d' ')
  elif command -v shasum > /dev/null 2>&1; then
    actual_shasum=$(shasum -a 256 "$file_name" | cut -f1 -d' ')
  else
    echo "sha256sum or shasum not found" >&2
    return 1
  fi

  if [[ "$actual_shasum" == "$expected_shasum" ]]; then
     return 0
  else
    return 1
  fi
}

download() {
  URL="$1"
  OUTPUT_FILE="$2"
  for i in $(seq 1 5); do
    if curl --retry 3 --retry-delay 1 --fail --show-error -o "$OUTPUT_FILE" "$URL"; then
      break
    fi
    sleep 15;
  done
}

PLATFORM="$(os)/$(arch)"
case "$PLATFORM" in
    linux/amd64)
      CONFLUENT_CLI_URL="${CONFLUENT_CLI_URL_BASE}/${CONFLUENT_CLI_VERSION}/confluent_v${CONFLUENT_CLI_VERSION}_linux_amd64.tar.gz"
      ;;
    darwin/amd64)
      CONFLUENT_CLI_URL="${CONFLUENT_CLI_URL_BASE}/${CONFLUENT_CLI_VERSION}/confluent_v${CONFLUENT_CLI_VERSION}_darwin_amd64.tar.gz"
      ;;
    *)
      echo "We don't know how to install the confluent CLI for \"${PLATFORM}\""
      exit 1
      ;;
esac

if ! [[ -f "$CONFLUENT_TAR_PATH" ]] || ! checkFile "$CONFLUENT_TAR_PATH" "$CONFLUENT_SHA256"; then
  download "$CONFLUENT_URL" "$CONFLUENT_TAR_PATH"
fi

tar xvf "$CONFLUENT_TAR_PATH" -C "$CONFLUENT_DIR"

if ! [[ -f "$CONFLUENT_DIR/bin/confluent" ]]; then
  if ! [[ -f "$CONFLUENT_CLI_TAR_PATH" ]]; then
    download "$CONFLUENT_CLI_URL" "$CONFLUENT_CLI_TAR_PATH"
  fi
  tar xvf "$CONFLUENT_CLI_TAR_PATH" -C "$CONFLUENT_DIR/$CONFLUENT_INSTALL_BASE/bin/" --strip-components=1 confluent/confluent
fi
`, confluentDownloadURL, confluentSHA256, confluentInstallBase, confluentCLIVersion, confluentCLIDownloadURLBase)

const (
	kafkaJAASConfig = `
KafkaServer {
   org.apache.kafka.common.security.plain.PlainLoginModule required
   username="admin"
   password="admin-secret"
   user_admin="admin-secret"
   user_plain="plain-secret";

   org.apache.kafka.common.security.scram.ScramLoginModule required
   username="admin"
   password="admin-secret";
};
`

	kafkaConfigTmpl = `
ssl.truststore.location=%s
ssl.truststore.password=%s

ssl.keystore.location=%s
ssl.keystore.password=%s

listeners=PLAINTEXT://:9092,SSL://:9093,SASL_SSL://:9094

sasl.enabled.mechanisms=SCRAM-SHA-256,SCRAM-SHA-512,PLAIN
sasl.mechanism.inter.broker.protocol=PLAIN
inter.broker.listener.name=SASL_SSL

# The following is from the confluent-4.0 default configuration file.
num.network.threads=3
num.io.threads=8
socket.send.buffer.bytes=102400
socket.receive.buffer.bytes=102400
socket.request.max.bytes=104857600
num.partitions=1
num.recovery.threads.per.data.dir=1

offsets.topic.replication.factor=1
transaction.state.log.replication.factor=1
transaction.state.log.min.isr=1

log.retention.hours=168
log.segment.bytes=1073741824
log.retention.check.interval.ms=300000
zookeeper.connect=localhost:2181
zookeeper.connection.timeout.ms=6000
confluent.support.metrics.enable=false
confluent.support.customer.id=anonymous
`
)

type kafkaManager struct {
	t     test.Test
	c     cluster.Cluster
	nodes option.NodeListOption
}

func (k kafkaManager) basePath() string {
	__antithesis_instrumentation__.Notify(46318)
	if k.c.IsLocal() {
		__antithesis_instrumentation__.Notify(46320)
		return `/tmp/confluent`
	} else {
		__antithesis_instrumentation__.Notify(46321)
	}
	__antithesis_instrumentation__.Notify(46319)
	return `/mnt/data1/confluent`
}

func (k kafkaManager) confluentHome() string {
	__antithesis_instrumentation__.Notify(46322)
	return filepath.Join(k.basePath(), confluentInstallBase)
}

func (k kafkaManager) configDir() string {
	__antithesis_instrumentation__.Notify(46323)
	return filepath.Join(k.basePath(), confluentInstallBase, "etc/kafka")
}

func (k kafkaManager) binDir() string {
	__antithesis_instrumentation__.Notify(46324)
	return filepath.Join(k.basePath(), confluentInstallBase, "bin")
}

func (k kafkaManager) confluentBin() string {
	__antithesis_instrumentation__.Notify(46325)
	return filepath.Join(k.binDir(), "confluent")
}

func (k kafkaManager) serverJAASConfig() string {
	__antithesis_instrumentation__.Notify(46326)
	return filepath.Join(k.configDir(), "server_jaas.conf")
}

func (k kafkaManager) install(ctx context.Context) {
	__antithesis_instrumentation__.Notify(46327)
	k.t.Status("installing kafka")
	folder := k.basePath()

	k.c.Run(ctx, k.nodes, `mkdir -p `+folder)

	downloadScriptPath := filepath.Join(folder, "install.sh")
	err := k.c.PutString(ctx, confluentDownloadScript, downloadScriptPath, 0700, k.nodes)
	if err != nil {
		__antithesis_instrumentation__.Notify(46329)
		k.t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(46330)
	}
	__antithesis_instrumentation__.Notify(46328)
	k.c.Run(ctx, k.nodes, downloadScriptPath, folder)
	if !k.c.IsLocal() {
		__antithesis_instrumentation__.Notify(46331)
		k.c.Run(ctx, k.nodes, `mkdir -p logs`)
		if err := k.installJRE(ctx); err != nil {
			__antithesis_instrumentation__.Notify(46332)
			k.t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(46333)
		}
	} else {
		__antithesis_instrumentation__.Notify(46334)
	}
}

func (k kafkaManager) installJRE(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(46335)
	retryOpts := retry.Options{
		InitialBackoff: 1 * time.Minute,
		MaxBackoff:     5 * time.Minute,
	}
	return retry.WithMaxAttempts(ctx, retryOpts, 3, func() error {
		__antithesis_instrumentation__.Notify(46336)
		err := k.c.RunE(ctx, k.nodes, `sudo apt-get -q update 2>&1 > logs/apt-get-update.log`)
		if err != nil {
			__antithesis_instrumentation__.Notify(46338)
			return err
		} else {
			__antithesis_instrumentation__.Notify(46339)
		}
		__antithesis_instrumentation__.Notify(46337)
		return k.c.RunE(ctx, k.nodes, `sudo DEBIAN_FRONTEND=noninteractive apt-get -yq --no-install-recommends install openssl default-jre 2>&1 > logs/apt-get-install.log`)
	})
}

func (k kafkaManager) configureAuth(ctx context.Context) *testCerts {
	__antithesis_instrumentation__.Notify(46340)
	k.t.Status("generating TLS certificates")
	ips, err := k.c.InternalIP(ctx, k.t.L(), k.nodes)
	if err != nil {
		__antithesis_instrumentation__.Notify(46343)
		k.t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(46344)
	}
	__antithesis_instrumentation__.Notify(46341)
	kafkaIP := ips[0]

	testCerts, err := makeTestCerts(kafkaIP)
	if err != nil {
		__antithesis_instrumentation__.Notify(46345)
		k.t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(46346)
	}
	__antithesis_instrumentation__.Notify(46342)

	configDir := k.configDir()

	truststorePath := filepath.Join(configDir, "kafka.truststore.jks")

	keystorePath := filepath.Join(configDir, "kafka.keystore.jks")

	caKeyPath := filepath.Join(configDir, "ca.key")
	caCertPath := filepath.Join(configDir, "ca.crt")

	kafkaKeyPath := filepath.Join(configDir, "kafka.key")
	kafkaCertPath := filepath.Join(configDir, "kafka.crt")
	kafkaBundlePath := filepath.Join(configDir, "kafka.p12")

	kafkaConfigPath := filepath.Join(configDir, "server.properties")
	kafkaJAASPath := filepath.Join(configDir, "server_jaas.conf")

	k.t.Status("writing kafka configuration files")
	kafkaConfig := fmt.Sprintf(kafkaConfigTmpl,
		truststorePath,
		keystorePassword,
		keystorePath,
		keystorePassword,
	)

	k.PutConfigContent(ctx, testCerts.KafkaKey, kafkaKeyPath)
	k.PutConfigContent(ctx, testCerts.KafkaCert, kafkaCertPath)
	k.PutConfigContent(ctx, testCerts.CAKey, caKeyPath)
	k.PutConfigContent(ctx, testCerts.CACert, caCertPath)
	k.PutConfigContent(ctx, kafkaConfig, kafkaConfigPath)
	k.PutConfigContent(ctx, kafkaJAASConfig, kafkaJAASPath)

	k.t.Status("constructing java keystores")

	k.c.Run(ctx, k.nodes,
		fmt.Sprintf("openssl pkcs12 -export -in %s -inkey %s -name kafka -out %s -password pass:%s",
			kafkaCertPath,
			kafkaKeyPath,
			kafkaBundlePath,
			keystorePassword))

	k.c.Run(ctx, k.nodes, fmt.Sprintf("rm -f %s", keystorePath))
	k.c.Run(ctx, k.nodes, fmt.Sprintf("rm -f %s", truststorePath))

	k.c.Run(ctx, k.nodes,
		fmt.Sprintf("keytool -importkeystore -deststorepass %s -destkeystore %s -srckeystore %s -srcstoretype PKCS12 -srcstorepass %s -alias kafka",
			keystorePassword,
			keystorePath,
			kafkaBundlePath,
			keystorePassword))
	k.c.Run(ctx, k.nodes,
		fmt.Sprintf("keytool -keystore %s -alias CAroot -importcert -file %s -no-prompt -storepass %s",
			truststorePath,
			caCertPath,
			keystorePassword))
	k.c.Run(ctx, k.nodes,
		fmt.Sprintf("keytool -keystore %s -alias CAroot -importcert -file %s -no-prompt -storepass %s",
			keystorePath,
			caCertPath,
			keystorePassword))

	return testCerts
}

func (k kafkaManager) PutConfigContent(ctx context.Context, data string, path string) {
	__antithesis_instrumentation__.Notify(46347)
	err := k.c.PutString(ctx, data, path, 0600, k.nodes)
	if err != nil {
		__antithesis_instrumentation__.Notify(46348)
		k.t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(46349)
	}
}

func (k kafkaManager) addSCRAMUsers(ctx context.Context) {
	__antithesis_instrumentation__.Notify(46350)
	k.t.Status("adding entries for SASL/SCRAM users")
	k.c.Run(ctx, k.nodes, filepath.Join(k.binDir(), "kafka-configs"),
		"--zookeeper", "localhost:2181",
		"--alter",
		"--add-config", "SCRAM-SHA-512=[password=scram512-secret]",
		"--entity-type", "users",
		"--entity-name", "scram512")

	k.c.Run(ctx, k.nodes, filepath.Join(k.binDir(), "kafka-configs"),
		"--zookeeper", "localhost:2181",
		"--alter",
		"--add-config", "SCRAM-SHA-256=[password=scram256-secret]",
		"--entity-type", "users",
		"--entity-name", "scram256")
}

func (k kafkaManager) start(ctx context.Context, services ...string) {
	__antithesis_instrumentation__.Notify(46351)

	k.c.Run(ctx, k.nodes, k.makeCommand("confluent", "local destroy || true"))
	k.restart(ctx, services...)
}

var kafkaServices = map[string][]string{
	"zookeeper":       {"zookeeper"},
	"kafka":           {"zookeeper", "kafka"},
	"schema-registry": {"zookeeper", "kafka", "schema-registry"},
}

func (k kafkaManager) kafkaServicesForTargets(targets []string) []string {
	__antithesis_instrumentation__.Notify(46352)
	var services []string
	for _, tgt := range targets {
		__antithesis_instrumentation__.Notify(46354)
		if s, ok := kafkaServices[tgt]; ok {
			__antithesis_instrumentation__.Notify(46355)
			services = append(services, s...)
		} else {
			__antithesis_instrumentation__.Notify(46356)
			k.t.Fatalf("unknown kafka start target %q", tgt)
		}
	}
	__antithesis_instrumentation__.Notify(46353)
	return services
}

func (k kafkaManager) restart(ctx context.Context, targetServices ...string) {
	__antithesis_instrumentation__.Notify(46357)
	var services []string
	if len(targetServices) == 0 {
		__antithesis_instrumentation__.Notify(46359)
		services = kafkaServices["schema-registry"]
	} else {
		__antithesis_instrumentation__.Notify(46360)
		services = k.kafkaServicesForTargets(targetServices)
	}
	__antithesis_instrumentation__.Notify(46358)

	k.c.Run(ctx, k.nodes, "touch", k.serverJAASConfig())
	for _, svcName := range services {
		__antithesis_instrumentation__.Notify(46361)

		opts := fmt.Sprintf("-Djava.security.auth.login.config=%s -Dkafka.logs.dir=%s",
			k.serverJAASConfig(),
			fmt.Sprintf("logs/%s", svcName),
		)
		startCmd := fmt.Sprintf(
			"CONFLUENT_CURRENT=%s CONFLUENT_HOME=%s KAFKA_OPTS='%s' %s local services %s start",
			k.basePath(),
			k.confluentHome(),
			opts,
			k.confluentBin(),
			svcName)
		k.c.Run(ctx, k.nodes, startCmd)
	}

}

func (k kafkaManager) makeCommand(exe string, args ...string) string {
	__antithesis_instrumentation__.Notify(46362)
	cmdPath := filepath.Join(k.binDir(), exe)
	return fmt.Sprintf("CONFLUENT_CURRENT=%s CONFLUENT_HOME=%s %s %s",
		k.basePath(),
		k.confluentHome(),
		cmdPath, strings.Join(args, " "))
}

func (k kafkaManager) stop(ctx context.Context) {
	__antithesis_instrumentation__.Notify(46363)
	k.c.Run(ctx, k.nodes, fmt.Sprintf("rm -f %s", k.serverJAASConfig()))
	k.c.Run(ctx, k.nodes, k.makeCommand("confluent", "local services stop"))
}

func (k kafkaManager) chaosLoop(
	ctx context.Context, period, downTime time.Duration, stopper chan struct{},
) error {
	__antithesis_instrumentation__.Notify(46364)
	t := time.NewTicker(period)
	for {
		__antithesis_instrumentation__.Notify(46365)
		select {
		case <-stopper:
			__antithesis_instrumentation__.Notify(46368)
			return nil
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(46369)
			return ctx.Err()
		case <-t.C:
			__antithesis_instrumentation__.Notify(46370)
		}
		__antithesis_instrumentation__.Notify(46366)

		k.stop(ctx)

		select {
		case <-stopper:
			__antithesis_instrumentation__.Notify(46371)
			return nil
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(46372)
			return ctx.Err()
		case <-time.After(downTime):
			__antithesis_instrumentation__.Notify(46373)
		}
		__antithesis_instrumentation__.Notify(46367)

		k.restart(ctx)
	}
}

func (k kafkaManager) sinkURL(ctx context.Context) string {
	__antithesis_instrumentation__.Notify(46374)
	ips, err := k.c.InternalIP(ctx, k.t.L(), k.nodes)
	if err != nil {
		__antithesis_instrumentation__.Notify(46376)
		k.t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(46377)
	}
	__antithesis_instrumentation__.Notify(46375)
	return `kafka://` + ips[0] + `:9092`
}

func (k kafkaManager) sinkURLTLS(ctx context.Context) string {
	__antithesis_instrumentation__.Notify(46378)
	ips, err := k.c.InternalIP(ctx, k.t.L(), k.nodes)
	if err != nil {
		__antithesis_instrumentation__.Notify(46380)
		k.t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(46381)
	}
	__antithesis_instrumentation__.Notify(46379)
	return `kafka://` + ips[0] + `:9093`
}

func (k kafkaManager) sinkURLSASL(ctx context.Context) string {
	__antithesis_instrumentation__.Notify(46382)
	ips, err := k.c.InternalIP(ctx, k.t.L(), k.nodes)
	if err != nil {
		__antithesis_instrumentation__.Notify(46384)
		k.t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(46385)
	}
	__antithesis_instrumentation__.Notify(46383)
	return `kafka://` + ips[0] + `:9094`
}

func (k kafkaManager) consumerURL(ctx context.Context) string {
	__antithesis_instrumentation__.Notify(46386)
	ips, err := k.c.ExternalIP(ctx, k.t.L(), k.nodes)
	if err != nil {
		__antithesis_instrumentation__.Notify(46388)
		k.t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(46389)
	}
	__antithesis_instrumentation__.Notify(46387)
	return ips[0] + `:9092`
}

func (k kafkaManager) schemaRegistryURL(ctx context.Context) string {
	__antithesis_instrumentation__.Notify(46390)
	ips, err := k.c.InternalIP(ctx, k.t.L(), k.nodes)
	if err != nil {
		__antithesis_instrumentation__.Notify(46392)
		k.t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(46393)
	}
	__antithesis_instrumentation__.Notify(46391)
	return `http://` + ips[0] + `:8081`
}

func (k kafkaManager) createTopic(ctx context.Context, topic string) error {
	__antithesis_instrumentation__.Notify(46394)
	kafkaAddrs := []string{k.consumerURL(ctx)}
	config := sarama.NewConfig()
	admin, err := sarama.NewClusterAdmin(kafkaAddrs, config)
	if err != nil {
		__antithesis_instrumentation__.Notify(46396)
		return errors.Wrap(err, "admin client")
	} else {
		__antithesis_instrumentation__.Notify(46397)
	}
	__antithesis_instrumentation__.Notify(46395)
	return admin.CreateTopic(topic, &sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1,
	}, false)
}

func (k kafkaManager) consumer(ctx context.Context, topic string) (*topicConsumer, error) {
	__antithesis_instrumentation__.Notify(46398)
	kafkaAddrs := []string{k.consumerURL(ctx)}
	config := sarama.NewConfig()

	config.Consumer.Fetch.Default = 1000012
	consumer, err := sarama.NewConsumer(kafkaAddrs, config)
	if err != nil {
		__antithesis_instrumentation__.Notify(46401)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(46402)
	}
	__antithesis_instrumentation__.Notify(46399)
	tc, err := makeTopicConsumer(consumer, topic)
	if err != nil {
		__antithesis_instrumentation__.Notify(46403)
		_ = consumer.Close()
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(46404)
	}
	__antithesis_instrumentation__.Notify(46400)
	return tc, nil
}

type tpccWorkload struct {
	workloadNodes      option.NodeListOption
	sqlNodes           option.NodeListOption
	tpccWarehouseCount int
	tolerateErrors     bool
}

func (tw *tpccWorkload) install(ctx context.Context, c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(46405)

	c.Run(ctx, tw.workloadNodes, fmt.Sprintf(
		`./cockroach workload fixtures import tpcc --warehouses=%d --checks=false {pgurl%s}`,
		tw.tpccWarehouseCount,
		tw.sqlNodes.RandNode(),
	))
}

func (tw *tpccWorkload) run(ctx context.Context, c cluster.Cluster, workloadDuration string) {
	__antithesis_instrumentation__.Notify(46406)
	tolerateErrors := ""
	if tw.tolerateErrors {
		__antithesis_instrumentation__.Notify(46408)
		tolerateErrors = "--tolerate-errors"
	} else {
		__antithesis_instrumentation__.Notify(46409)
	}
	__antithesis_instrumentation__.Notify(46407)
	c.Run(ctx, tw.workloadNodes, fmt.Sprintf(
		`./workload run tpcc --warehouses=%d --duration=%s %s {pgurl%s} `,
		tw.tpccWarehouseCount, workloadDuration, tolerateErrors, tw.sqlNodes,
	))
}

type ledgerWorkload struct {
	workloadNodes option.NodeListOption
	sqlNodes      option.NodeListOption
}

func (lw *ledgerWorkload) install(ctx context.Context, c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(46410)
	c.Run(ctx, lw.workloadNodes.RandNode(), fmt.Sprintf(
		`./workload init ledger {pgurl%s}`,
		lw.sqlNodes.RandNode(),
	))
}

func (lw *ledgerWorkload) run(ctx context.Context, c cluster.Cluster, workloadDuration string) {
	__antithesis_instrumentation__.Notify(46411)
	c.Run(ctx, lw.workloadNodes, fmt.Sprintf(
		`./workload run ledger --mix=balance=0,withdrawal=50,deposit=50,reversal=0 {pgurl%s} --duration=%s`,
		lw.sqlNodes,
		workloadDuration,
	))
}

type latencyVerifier struct {
	statementTime            time.Time
	targetSteadyLatency      time.Duration
	targetInitialScanLatency time.Duration
	tolerateErrors           bool
	logger                   *logger.Logger
	setTestStatus            func(...interface{})

	initialScanLatency   time.Duration
	maxSeenSteadyLatency time.Duration
	maxSeenSteadyEveryN  log.EveryN
	latencyBecameSteady  bool

	latencyHist *hdrhistogram.Histogram
}

func makeLatencyVerifier(
	targetInitialScanLatency time.Duration,
	targetSteadyLatency time.Duration,
	l *logger.Logger,
	setTestStatus func(...interface{}),
	tolerateErrors bool,
) *latencyVerifier {
	__antithesis_instrumentation__.Notify(46412)
	const sigFigs, minLatency, maxLatency = 1, 100 * time.Microsecond, 100 * time.Second
	hist := hdrhistogram.New(minLatency.Nanoseconds(), maxLatency.Nanoseconds(), sigFigs)
	return &latencyVerifier{
		targetInitialScanLatency: targetInitialScanLatency,
		targetSteadyLatency:      targetSteadyLatency,
		logger:                   l,
		setTestStatus:            setTestStatus,
		latencyHist:              hist,
		tolerateErrors:           tolerateErrors,
		maxSeenSteadyEveryN:      log.Every(10 * time.Second),
	}
}

func (lv *latencyVerifier) noteHighwater(highwaterTime time.Time) {
	__antithesis_instrumentation__.Notify(46413)
	if highwaterTime.Before(lv.statementTime) {
		__antithesis_instrumentation__.Notify(46420)
		return
	} else {
		__antithesis_instrumentation__.Notify(46421)
	}
	__antithesis_instrumentation__.Notify(46414)
	if lv.initialScanLatency == 0 {
		__antithesis_instrumentation__.Notify(46422)
		lv.initialScanLatency = timeutil.Since(lv.statementTime)
		lv.logger.Printf("initial scan completed: latency %s\n", lv.initialScanLatency)
		return
	} else {
		__antithesis_instrumentation__.Notify(46423)
	}
	__antithesis_instrumentation__.Notify(46415)

	latency := timeutil.Since(highwaterTime)
	if latency < lv.targetSteadyLatency/2 {
		__antithesis_instrumentation__.Notify(46424)
		lv.latencyBecameSteady = true
	} else {
		__antithesis_instrumentation__.Notify(46425)
	}
	__antithesis_instrumentation__.Notify(46416)
	if !lv.latencyBecameSteady {
		__antithesis_instrumentation__.Notify(46426)

		if lv.maxSeenSteadyEveryN.ShouldLog() {
			__antithesis_instrumentation__.Notify(46428)
			lv.setTestStatus(fmt.Sprintf(
				"watching changefeed: end-to-end latency %s not yet below target steady latency %s",
				latency.Truncate(time.Millisecond), lv.targetSteadyLatency.Truncate(time.Millisecond)))
		} else {
			__antithesis_instrumentation__.Notify(46429)
		}
		__antithesis_instrumentation__.Notify(46427)
		return
	} else {
		__antithesis_instrumentation__.Notify(46430)
	}
	__antithesis_instrumentation__.Notify(46417)
	if err := lv.latencyHist.RecordValue(latency.Nanoseconds()); err != nil {
		__antithesis_instrumentation__.Notify(46431)
		lv.logger.Printf("could not record value %s: %s\n", latency, err)
	} else {
		__antithesis_instrumentation__.Notify(46432)
	}
	__antithesis_instrumentation__.Notify(46418)
	if latency > lv.maxSeenSteadyLatency {
		__antithesis_instrumentation__.Notify(46433)
		lv.maxSeenSteadyLatency = latency
	} else {
		__antithesis_instrumentation__.Notify(46434)
	}
	__antithesis_instrumentation__.Notify(46419)
	if lv.maxSeenSteadyEveryN.ShouldLog() {
		__antithesis_instrumentation__.Notify(46435)
		lv.setTestStatus(fmt.Sprintf(
			"watching changefeed: end-to-end steady latency %s; max steady latency so far %s",
			latency.Truncate(time.Millisecond), lv.maxSeenSteadyLatency.Truncate(time.Millisecond)))
	} else {
		__antithesis_instrumentation__.Notify(46436)
	}
}

func (lv *latencyVerifier) pollLatency(
	ctx context.Context, db *gosql.DB, jobID int, interval time.Duration, stopper chan struct{},
) error {
	__antithesis_instrumentation__.Notify(46437)
	for {
		__antithesis_instrumentation__.Notify(46438)
		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(46442)
			return ctx.Err()
		case <-stopper:
			__antithesis_instrumentation__.Notify(46443)
			return nil
		case <-time.After(time.Second):
			__antithesis_instrumentation__.Notify(46444)
		}
		__antithesis_instrumentation__.Notify(46439)

		info, err := getChangefeedInfo(db, jobID)
		if err != nil {
			__antithesis_instrumentation__.Notify(46445)
			if lv.tolerateErrors {
				__antithesis_instrumentation__.Notify(46447)
				lv.logger.Printf("error getting changefeed info: %s", err)
				continue
			} else {
				__antithesis_instrumentation__.Notify(46448)
			}
			__antithesis_instrumentation__.Notify(46446)
			return err
		} else {
			__antithesis_instrumentation__.Notify(46449)
		}
		__antithesis_instrumentation__.Notify(46440)
		if info.status != `running` {
			__antithesis_instrumentation__.Notify(46450)
			lv.logger.Printf("unexpected status: %s, error: %s", info.status, info.errMsg)
			return errors.Errorf(`unexpected status: %s`, info.status)
		} else {
			__antithesis_instrumentation__.Notify(46451)
		}
		__antithesis_instrumentation__.Notify(46441)
		lv.noteHighwater(info.highwaterTime)
	}
}

func (lv *latencyVerifier) assertValid(t test.Test) {
	__antithesis_instrumentation__.Notify(46452)
	if lv.initialScanLatency == 0 {
		__antithesis_instrumentation__.Notify(46456)
		t.Fatalf("initial scan did not complete")
	} else {
		__antithesis_instrumentation__.Notify(46457)
	}
	__antithesis_instrumentation__.Notify(46453)
	if lv.initialScanLatency > lv.targetInitialScanLatency {
		__antithesis_instrumentation__.Notify(46458)
		t.Fatalf("initial scan latency was more than target: %s vs %s",
			lv.initialScanLatency, lv.targetInitialScanLatency)
	} else {
		__antithesis_instrumentation__.Notify(46459)
	}
	__antithesis_instrumentation__.Notify(46454)
	if !lv.latencyBecameSteady {
		__antithesis_instrumentation__.Notify(46460)
		t.Fatalf("latency never dropped to acceptable steady level: %s", lv.targetSteadyLatency)
	} else {
		__antithesis_instrumentation__.Notify(46461)
	}
	__antithesis_instrumentation__.Notify(46455)
	if lv.maxSeenSteadyLatency > lv.targetSteadyLatency {
		__antithesis_instrumentation__.Notify(46462)
		t.Fatalf("max latency was more than allowed: %s vs %s",
			lv.maxSeenSteadyLatency, lv.targetSteadyLatency)
	} else {
		__antithesis_instrumentation__.Notify(46463)
	}
}

func (lv *latencyVerifier) maybeLogLatencyHist() {
	__antithesis_instrumentation__.Notify(46464)
	if lv.latencyHist == nil {
		__antithesis_instrumentation__.Notify(46466)
		return
	} else {
		__antithesis_instrumentation__.Notify(46467)
	}
	__antithesis_instrumentation__.Notify(46465)
	lv.logger.Printf(
		"changefeed end-to-end __avg(ms)__p50(ms)__p75(ms)__p90(ms)__p95(ms)__p99(ms)_pMax(ms)\n")
	lv.logger.Printf("changefeed end-to-end  %8.1f %8.1f %8.1f %8.1f %8.1f %8.1f %8.1f\n",
		time.Duration(lv.latencyHist.Mean()).Seconds()*1000,
		time.Duration(lv.latencyHist.ValueAtQuantile(50)).Seconds()*1000,
		time.Duration(lv.latencyHist.ValueAtQuantile(75)).Seconds()*1000,
		time.Duration(lv.latencyHist.ValueAtQuantile(90)).Seconds()*1000,
		time.Duration(lv.latencyHist.ValueAtQuantile(95)).Seconds()*1000,
		time.Duration(lv.latencyHist.ValueAtQuantile(99)).Seconds()*1000,
		time.Duration(lv.latencyHist.ValueAtQuantile(100)).Seconds()*1000,
	)
}

func createChangefeed(db *gosql.DB, targets, sinkURL string, args cdcTestArgs) (int, error) {
	__antithesis_instrumentation__.Notify(46468)
	var jobID int
	createStmt := fmt.Sprintf(`CREATE CHANGEFEED FOR %s INTO $1`, targets)
	extraArgs := []interface{}{sinkURL}
	if args.whichSink == cloudStorageSink || func() bool {
		__antithesis_instrumentation__.Notify(46472)
		return args.whichSink == webhookSink == true
	}() == true {
		__antithesis_instrumentation__.Notify(46473)
		createStmt += ` WITH resolved='10s', envelope=wrapped, min_checkpoint_frequency='10s'`
	} else {
		__antithesis_instrumentation__.Notify(46474)
		createStmt += ` WITH resolved,  min_checkpoint_frequency='10s'`
	}
	__antithesis_instrumentation__.Notify(46469)
	if !args.initialScan {
		__antithesis_instrumentation__.Notify(46475)
		createStmt += `, cursor='-1s'`
	} else {
		__antithesis_instrumentation__.Notify(46476)
	}
	__antithesis_instrumentation__.Notify(46470)
	if err := db.QueryRow(createStmt, extraArgs...).Scan(&jobID); err != nil {
		__antithesis_instrumentation__.Notify(46477)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(46478)
	}
	__antithesis_instrumentation__.Notify(46471)
	return jobID, nil
}

type changefeedInfo struct {
	status        string
	errMsg        string
	statementTime time.Time
	highwaterTime time.Time
}

func getChangefeedInfo(db *gosql.DB, jobID int) (changefeedInfo, error) {
	__antithesis_instrumentation__.Notify(46479)
	var status string
	var payloadBytes []byte
	var progressBytes []byte
	if err := db.QueryRow(
		`SELECT status, payload, progress FROM system.jobs WHERE id = $1`, jobID,
	).Scan(&status, &payloadBytes, &progressBytes); err != nil {
		__antithesis_instrumentation__.Notify(46484)
		return changefeedInfo{}, err
	} else {
		__antithesis_instrumentation__.Notify(46485)
	}
	__antithesis_instrumentation__.Notify(46480)
	var payload jobspb.Payload
	if err := protoutil.Unmarshal(payloadBytes, &payload); err != nil {
		__antithesis_instrumentation__.Notify(46486)
		return changefeedInfo{}, err
	} else {
		__antithesis_instrumentation__.Notify(46487)
	}
	__antithesis_instrumentation__.Notify(46481)
	var progress jobspb.Progress
	if err := protoutil.Unmarshal(progressBytes, &progress); err != nil {
		__antithesis_instrumentation__.Notify(46488)
		return changefeedInfo{}, err
	} else {
		__antithesis_instrumentation__.Notify(46489)
	}
	__antithesis_instrumentation__.Notify(46482)
	var highwaterTime time.Time
	highwater := progress.GetHighWater()
	if highwater != nil {
		__antithesis_instrumentation__.Notify(46490)
		highwaterTime = highwater.GoTime()
	} else {
		__antithesis_instrumentation__.Notify(46491)
	}
	__antithesis_instrumentation__.Notify(46483)
	return changefeedInfo{
		status:        status,
		errMsg:        payload.Error,
		statementTime: payload.GetChangefeed().StatementTime.GoTime(),
		highwaterTime: highwaterTime,
	}, nil
}

func stopFeeds(db *gosql.DB) {
	__antithesis_instrumentation__.Notify(46492)
	_, _ = db.Exec(`CANCEL JOBS (
			SELECT job_id FROM [SHOW JOBS] WHERE status = 'running'
		)`)
}

type topicConsumer struct {
	sarama.Consumer

	topic              string
	partitions         []string
	partitionConsumers []sarama.PartitionConsumer
}

func makeTopicConsumer(c sarama.Consumer, topic string) (*topicConsumer, error) {
	__antithesis_instrumentation__.Notify(46493)
	t := &topicConsumer{Consumer: c, topic: topic}
	partitions, err := t.Partitions(t.topic)
	if err != nil {
		__antithesis_instrumentation__.Notify(46496)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(46497)
	}
	__antithesis_instrumentation__.Notify(46494)
	for _, partition := range partitions {
		__antithesis_instrumentation__.Notify(46498)
		t.partitions = append(t.partitions, strconv.Itoa(int(partition)))
		pc, err := t.ConsumePartition(topic, partition, sarama.OffsetOldest)
		if err != nil {
			__antithesis_instrumentation__.Notify(46500)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(46501)
		}
		__antithesis_instrumentation__.Notify(46499)
		t.partitionConsumers = append(t.partitionConsumers, pc)
	}
	__antithesis_instrumentation__.Notify(46495)
	return t, nil
}

func (c *topicConsumer) tryNextMessage(ctx context.Context) *sarama.ConsumerMessage {
	__antithesis_instrumentation__.Notify(46502)
	for _, pc := range c.partitionConsumers {
		__antithesis_instrumentation__.Notify(46504)
		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(46505)
			return nil
		case m := <-pc.Messages():
			__antithesis_instrumentation__.Notify(46506)
			return m
		default:
			__antithesis_instrumentation__.Notify(46507)
		}
	}
	__antithesis_instrumentation__.Notify(46503)
	return nil
}

func (c *topicConsumer) Next(ctx context.Context) *sarama.ConsumerMessage {
	__antithesis_instrumentation__.Notify(46508)
	m := c.tryNextMessage(ctx)
	for ; m == nil; m = c.tryNextMessage(ctx) {
		__antithesis_instrumentation__.Notify(46510)
		if ctx.Err() != nil {
			__antithesis_instrumentation__.Notify(46511)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(46512)
		}
	}
	__antithesis_instrumentation__.Notify(46509)
	return m
}

func (c *topicConsumer) Close() {
	__antithesis_instrumentation__.Notify(46513)
	for _, pc := range c.partitionConsumers {
		__antithesis_instrumentation__.Notify(46515)
		pc.AsyncClose()

		for range pc.Messages() {
			__antithesis_instrumentation__.Notify(46517)
		}
		__antithesis_instrumentation__.Notify(46516)
		for range pc.Errors() {
			__antithesis_instrumentation__.Notify(46518)
		}
	}
	__antithesis_instrumentation__.Notify(46514)
	_ = c.Consumer.Close()
}
