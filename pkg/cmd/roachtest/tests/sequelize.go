package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"regexp"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
)

var sequelizeCockroachDBReleaseTagRegex = regexp.MustCompile(`^v(?P<major>\d+)\.(?P<minor>\d+)\.(?P<point>\d+)$`)
var supportedSequelizeCockroachDBRelease = "v6.0.5"

func registerSequelize(r registry.Registry) {
	__antithesis_instrumentation__.Notify(50789)
	runSequelize := func(
		ctx context.Context,
		t test.Test,
		c cluster.Cluster,
	) {
		__antithesis_instrumentation__.Notify(50791)
		if c.IsLocal() {
			__antithesis_instrumentation__.Notify(50807)
			t.Fatal("cannot be run in local mode")
		} else {
			__antithesis_instrumentation__.Notify(50808)
		}
		__antithesis_instrumentation__.Notify(50792)
		node := c.Node(1)
		t.Status("setting up cockroach")
		c.Put(ctx, t.Cockroach(), "./cockroach", c.All())
		if err := c.PutLibraries(ctx, "./lib"); err != nil {
			__antithesis_instrumentation__.Notify(50809)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50810)
		}
		__antithesis_instrumentation__.Notify(50793)
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.All())

		version, err := fetchCockroachVersion(ctx, t.L(), c, node[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(50811)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50812)
		}
		__antithesis_instrumentation__.Notify(50794)

		if err := alterZoneConfigAndClusterSettings(ctx, t, version, c, node[0]); err != nil {
			__antithesis_instrumentation__.Notify(50813)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50814)
		}
		__antithesis_instrumentation__.Notify(50795)

		t.Status("create database used by tests")
		db, err := c.ConnE(ctx, t.L(), node[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(50815)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50816)
		}
		__antithesis_instrumentation__.Notify(50796)
		defer db.Close()

		if _, err := db.ExecContext(
			ctx,
			`CREATE DATABASE sequelize_test`,
		); err != nil {
			__antithesis_instrumentation__.Notify(50817)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50818)
		}
		__antithesis_instrumentation__.Notify(50797)

		t.Status("cloning sequelize-cockroachdb and installing prerequisites")
		latestTag, err := repeatGetLatestTag(ctx, t, "cockroachdb", "sequelize-cockroachdb", sequelizeCockroachDBReleaseTagRegex)
		if err != nil {
			__antithesis_instrumentation__.Notify(50819)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50820)
		}
		__antithesis_instrumentation__.Notify(50798)
		t.L().Printf("Latest sequelize-cockroachdb release is %s.", latestTag)
		t.L().Printf("Supported sequelize-cockroachdb release is %s.", supportedSequelizeCockroachDBRelease)

		if err := repeatRunE(
			ctx, t, c, node, "update apt-get", `sudo apt-get -qq update`,
		); err != nil {
			__antithesis_instrumentation__.Notify(50821)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50822)
		}
		__antithesis_instrumentation__.Notify(50799)

		if err := repeatRunE(
			ctx,
			t,
			c,
			node,
			"install dependencies",
			`sudo apt-get -qq install make python3 libpq-dev python-dev gcc g++ `+
				`software-properties-common build-essential`,
		); err != nil {
			__antithesis_instrumentation__.Notify(50823)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50824)
		}
		__antithesis_instrumentation__.Notify(50800)

		if err := repeatRunE(
			ctx,
			t,
			c,
			node,
			"add nodesource repository",
			`sudo apt install ca-certificates && curl -sL https://deb.nodesource.com/setup_12.x | sudo -E bash -`,
		); err != nil {
			__antithesis_instrumentation__.Notify(50825)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50826)
		}
		__antithesis_instrumentation__.Notify(50801)

		if err := repeatRunE(
			ctx, t, c, node, "install nodejs and npm", `sudo apt-get -qq install nodejs`,
		); err != nil {
			__antithesis_instrumentation__.Notify(50827)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50828)
		}
		__antithesis_instrumentation__.Notify(50802)

		if err := repeatRunE(
			ctx, t, c, node, "update npm", `sudo npm i -g npm`,
		); err != nil {
			__antithesis_instrumentation__.Notify(50829)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50830)
		}
		__antithesis_instrumentation__.Notify(50803)

		if err := repeatRunE(
			ctx, t, c, node, "remove old sequelize", `sudo rm -rf /mnt/data1/sequelize`,
		); err != nil {
			__antithesis_instrumentation__.Notify(50831)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50832)
		}
		__antithesis_instrumentation__.Notify(50804)

		if err := repeatGitCloneE(
			ctx,
			t,
			c,
			"https://github.com/cockroachdb/sequelize-cockroachdb.git",
			"/mnt/data1/sequelize",
			supportedSequelizeCockroachDBRelease,
			node,
		); err != nil {
			__antithesis_instrumentation__.Notify(50833)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50834)
		}
		__antithesis_instrumentation__.Notify(50805)

		if err := repeatRunE(
			ctx, t, c, node, "install dependencies", `cd /mnt/data1/sequelize && sudo npm i`,
		); err != nil {
			__antithesis_instrumentation__.Notify(50835)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50836)
		}
		__antithesis_instrumentation__.Notify(50806)

		t.Status("running Sequelize test suite")
		result, err := c.RunWithDetailsSingleNode(ctx, t.L(), node,
			fmt.Sprintf(`cd /mnt/data1/sequelize/ && npm test --crdb_version=%s`, version),
		)
		rawResultsStr := result.Stdout + result.Stderr
		t.L().Printf("Test Results: %s", rawResultsStr)
		if err != nil {
			__antithesis_instrumentation__.Notify(50837)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50838)
		}
	}
	__antithesis_instrumentation__.Notify(50790)

	r.Add(registry.TestSpec{
		Name:    "sequelize",
		Owner:   registry.OwnerSQLExperience,
		Cluster: r.MakeClusterSpec(1),
		Tags:    []string{`default`, `orm`},
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(50839)
			runSequelize(ctx, t, c)
		},
	})
}
