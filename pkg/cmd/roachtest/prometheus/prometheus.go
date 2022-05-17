package prometheus

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/roachprod/logger"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"gopkg.in/yaml.v2"
)

type Client interface {
	Query(ctx context.Context, query string, ts time.Time) (model.Value, promv1.Warnings, error)
}

type ScrapeNode struct {
	Nodes option.NodeListOption
	Port  int
}

type ScrapeConfig struct {
	JobName     string
	MetricsPath string
	ScrapeNodes []ScrapeNode
}

type Config struct {
	PrometheusNode option.NodeListOption
	ScrapeConfigs  []ScrapeConfig
}

type Cluster interface {
	ExternalIP(context.Context, *logger.Logger, option.NodeListOption) ([]string, error)
	Get(ctx context.Context, l *logger.Logger, src, dest string, opts ...option.Option) error
	RunE(ctx context.Context, node option.NodeListOption, args ...string) error
	PutString(
		ctx context.Context, content, dest string, mode os.FileMode, opts ...option.Option,
	) error
}

type Prometheus struct {
	Config
}

func Init(
	ctx context.Context,
	cfg Config,
	c Cluster,
	l *logger.Logger,
	repeatFunc func(context.Context, option.NodeListOption, string, ...string) error,
) (*Prometheus, error) {
	__antithesis_instrumentation__.Notify(44332)
	if err := c.RunE(
		ctx,
		cfg.PrometheusNode,
		"sudo systemctl stop prometheus || echo 'no prometheus is running'",
	); err != nil {
		__antithesis_instrumentation__.Notify(44338)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(44339)
	}
	__antithesis_instrumentation__.Notify(44333)

	if err := repeatFunc(
		ctx,
		cfg.PrometheusNode,
		"download prometheus",
		`rm -rf /tmp/prometheus && mkdir /tmp/prometheus && cd /tmp/prometheus &&
			curl -fsSL https://storage.googleapis.com/cockroach-fixtures/prometheus/prometheus-2.27.1.linux-amd64.tar.gz | tar zxv --strip-components=1`,
	); err != nil {
		__antithesis_instrumentation__.Notify(44340)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(44341)
	}
	__antithesis_instrumentation__.Notify(44334)

	yamlCfg, err := makeYAMLConfig(
		ctx,
		l,
		c,
		cfg.ScrapeConfigs,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(44342)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(44343)
	}
	__antithesis_instrumentation__.Notify(44335)

	if err := c.PutString(
		ctx,
		yamlCfg,
		"/tmp/prometheus/prometheus.yml",
		0644,
		cfg.PrometheusNode,
	); err != nil {
		__antithesis_instrumentation__.Notify(44344)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(44345)
	}
	__antithesis_instrumentation__.Notify(44336)

	if err := c.RunE(
		ctx,
		cfg.PrometheusNode,
		`cd /tmp/prometheus &&
sudo systemd-run --unit prometheus --same-dir \
	./prometheus --config.file=prometheus.yml --storage.tsdb.path=data/ --web.enable-admin-api`,
	); err != nil {
		__antithesis_instrumentation__.Notify(44346)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(44347)
	}
	__antithesis_instrumentation__.Notify(44337)
	return &Prometheus{Config: cfg}, nil
}

func (pm *Prometheus) Snapshot(ctx context.Context, c Cluster, l *logger.Logger, dir string) error {
	__antithesis_instrumentation__.Notify(44348)
	if err := c.RunE(
		ctx,
		pm.PrometheusNode,
		`curl -XPOST http://localhost:9090/api/v1/admin/tsdb/snapshot &&
	cd /tmp/prometheus && tar cvf prometheus-snapshot.tar.gz data/snapshots`,
	); err != nil {
		__antithesis_instrumentation__.Notify(44351)
		return err
	} else {
		__antithesis_instrumentation__.Notify(44352)
	}
	__antithesis_instrumentation__.Notify(44349)
	if err := os.WriteFile(filepath.Join(dir, "prometheus-docker-run.sh"), []byte(`#!/bin/sh
set -eu

tar xf prometheus-snapshot.tar.gz
snapdir=$(find data/snapshots -mindepth 1 -maxdepth 1 -type d)
promyml=$(mktemp)
chmod -R o+rw "${snapdir}" "${promyml}"

cat <<EOF > "${promyml}"
global:
  scrape_interval: 10s
  scrape_timeout: 5s
EOF

set -x
# Worked as of v2.33.1 so you can hard-code that if necessary.
docker run --privileged -p 9090:9090 \
    -v "${promyml}:/etc/prometheus/prometheus.yml" -v "${PWD}/${snapdir}:/prometheus" \
    prom/prometheus:latest \
    --config.file=/etc/prometheus/prometheus.yml \
    --storage.tsdb.path=/prometheus \
    --web.enable-admin-api
`), 0755); err != nil {
		__antithesis_instrumentation__.Notify(44353)
		return err
	} else {
		__antithesis_instrumentation__.Notify(44354)
	}
	__antithesis_instrumentation__.Notify(44350)

	return c.Get(
		ctx,
		l,
		"/tmp/prometheus/prometheus-snapshot.tar.gz",
		dir,
		pm.PrometheusNode,
	)
}

const (
	DefaultScrapeInterval = 10 * time.Second

	DefaultScrapeTimeout = 5 * time.Second
)

func makeYAMLConfig(
	ctx context.Context, l *logger.Logger, c Cluster, scrapeConfigs []ScrapeConfig,
) (string, error) {
	__antithesis_instrumentation__.Notify(44355)
	type yamlStaticConfig struct {
		Targets []string
	}

	type yamlScrapeConfig struct {
		JobName       string             `yaml:"job_name"`
		StaticConfigs []yamlStaticConfig `yaml:"static_configs"`
		MetricsPath   string             `yaml:"metrics_path"`
	}

	type yamlConfig struct {
		Global struct {
			ScrapeInterval string `yaml:"scrape_interval"`
			ScrapeTimeout  string `yaml:"scrape_timeout"`
		}
		ScrapeConfigs []yamlScrapeConfig `yaml:"scrape_configs"`
	}

	cfg := yamlConfig{}
	cfg.Global.ScrapeInterval = DefaultScrapeInterval.String()
	cfg.Global.ScrapeTimeout = DefaultScrapeTimeout.String()

	for _, scrapeConfig := range scrapeConfigs {
		__antithesis_instrumentation__.Notify(44357)
		var targets []string
		for _, scrapeNode := range scrapeConfig.ScrapeNodes {
			__antithesis_instrumentation__.Notify(44359)
			ips, err := c.ExternalIP(ctx, l, scrapeNode.Nodes)
			if err != nil {
				__antithesis_instrumentation__.Notify(44361)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(44362)
			}
			__antithesis_instrumentation__.Notify(44360)
			for _, ip := range ips {
				__antithesis_instrumentation__.Notify(44363)
				targets = append(targets, fmt.Sprintf("%s:%d", ip, scrapeNode.Port))
			}
		}
		__antithesis_instrumentation__.Notify(44358)

		cfg.ScrapeConfigs = append(
			cfg.ScrapeConfigs,
			yamlScrapeConfig{
				JobName:     scrapeConfig.JobName,
				MetricsPath: scrapeConfig.MetricsPath,
				StaticConfigs: []yamlStaticConfig{
					{
						Targets: targets,
					},
				},
			},
		)
	}
	__antithesis_instrumentation__.Notify(44356)

	ret, err := yaml.Marshal(&cfg)
	return string(ret), err
}

func MakeWorkloadScrapeConfig(jobName string, scrapeNodes []ScrapeNode) ScrapeConfig {
	__antithesis_instrumentation__.Notify(44364)
	return ScrapeConfig{
		JobName:     jobName,
		MetricsPath: "/",
		ScrapeNodes: scrapeNodes,
	}
}

func MakeInsecureCockroachScrapeConfig(jobName string, nodes option.NodeListOption) ScrapeConfig {
	__antithesis_instrumentation__.Notify(44365)
	return ScrapeConfig{
		JobName:     jobName,
		MetricsPath: "/_status/vars",
		ScrapeNodes: []ScrapeNode{
			{
				Nodes: nodes,
				Port:  26258,
			},
		},
	}
}
