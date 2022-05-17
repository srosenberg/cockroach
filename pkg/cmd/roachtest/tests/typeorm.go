package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
)

var typeORMReleaseTagRegex = regexp.MustCompile(`^(?P<major>\d+)\.(?P<minor>\d+)\.(?P<point>\d+)$`)
var supportedTypeORMRelease = "0.3.5"

func registerTypeORM(r registry.Registry) {
	__antithesis_instrumentation__.Notify(52138)
	runTypeORM := func(
		ctx context.Context,
		t test.Test,
		c cluster.Cluster,
	) {
		__antithesis_instrumentation__.Notify(52140)
		if c.IsLocal() {
			__antithesis_instrumentation__.Notify(52156)
			t.Fatal("cannot be run in local mode")
		} else {
			__antithesis_instrumentation__.Notify(52157)
		}
		__antithesis_instrumentation__.Notify(52141)
		node := c.Node(1)
		t.Status("setting up cockroach")
		c.Put(ctx, t.Cockroach(), "./cockroach", c.All())
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.All())

		cockroachVersion, err := fetchCockroachVersion(ctx, t.L(), c, node[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(52158)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(52159)
		}
		__antithesis_instrumentation__.Notify(52142)

		if err := alterZoneConfigAndClusterSettings(ctx, t, cockroachVersion, c, node[0]); err != nil {
			__antithesis_instrumentation__.Notify(52160)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(52161)
		}
		__antithesis_instrumentation__.Notify(52143)

		t.Status("cloning TypeORM and installing prerequisites")
		latestTag, err := repeatGetLatestTag(ctx, t, "typeorm", "typeorm", typeORMReleaseTagRegex)
		if err != nil {
			__antithesis_instrumentation__.Notify(52162)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(52163)
		}
		__antithesis_instrumentation__.Notify(52144)
		t.L().Printf("Latest TypeORM release is %s.", latestTag)
		t.L().Printf("Supported TypeORM release is %s.", supportedTypeORMRelease)

		if err := repeatRunE(
			ctx, t, c, node, "purge apt-get",
			`sudo apt-get purge -y command-not-found`,
		); err != nil {
			__antithesis_instrumentation__.Notify(52164)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(52165)
		}
		__antithesis_instrumentation__.Notify(52145)

		if err := repeatRunE(
			ctx, t, c, node, "update apt-get", `sudo apt-get update`,
		); err != nil {
			__antithesis_instrumentation__.Notify(52166)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(52167)
		}
		__antithesis_instrumentation__.Notify(52146)

		if err := repeatRunE(
			ctx,
			t,
			c,
			node,
			"install dependencies",
			`sudo apt-get install -y make python3 libpq-dev python-dev gcc g++ `+
				`software-properties-common build-essential`,
		); err != nil {
			__antithesis_instrumentation__.Notify(52168)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(52169)
		}
		__antithesis_instrumentation__.Notify(52147)

		if err := repeatRunE(
			ctx,
			t,
			c,
			node,
			"add nodesource repository",
			`sudo apt install ca-certificates && curl -fsSL https://deb.nodesource.com/setup_14.x | sudo -E bash -`,
		); err != nil {
			__antithesis_instrumentation__.Notify(52170)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(52171)
		}
		__antithesis_instrumentation__.Notify(52148)

		if err := repeatRunE(
			ctx, t, c, node, "install nodejs and npm", `sudo apt-get install -y nodejs`,
		); err != nil {
			__antithesis_instrumentation__.Notify(52172)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(52173)
		}
		__antithesis_instrumentation__.Notify(52149)

		if err := repeatRunE(
			ctx, t, c, node, "update npm", `sudo npm i -g npm`,
		); err != nil {
			__antithesis_instrumentation__.Notify(52174)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(52175)
		}
		__antithesis_instrumentation__.Notify(52150)

		if err := repeatRunE(
			ctx, t, c, node, "remove old TypeORM", `sudo rm -rf /mnt/data1/typeorm`,
		); err != nil {
			__antithesis_instrumentation__.Notify(52176)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(52177)
		}
		__antithesis_instrumentation__.Notify(52151)

		if err := repeatGitCloneE(
			ctx,
			t,
			c,
			"https://github.com/typeorm/typeorm.git",
			"/mnt/data1/typeorm",
			supportedTypeORMRelease,
			node,
		); err != nil {
			__antithesis_instrumentation__.Notify(52178)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(52179)
		}
		__antithesis_instrumentation__.Notify(52152)

		if err := repeatRunE(
			ctx,
			t,
			c,
			node,
			"configuring tests for cockroach only",
			fmt.Sprintf("echo '%s' > /mnt/data1/typeorm/ormconfig.json", typeORMConfigJSON),
		); err != nil {
			__antithesis_instrumentation__.Notify(52180)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(52181)
		}
		__antithesis_instrumentation__.Notify(52153)

		if err := repeatRunE(
			ctx,
			t,
			c,
			node,
			"patch TypeORM test script to run all tests even on failure",
			`sed -i 's/--bail //' /mnt/data1/typeorm/package.json`,
		); err != nil {
			__antithesis_instrumentation__.Notify(52182)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(52183)
		}
		__antithesis_instrumentation__.Notify(52154)

		if err := repeatRunE(
			ctx,
			t,
			c,
			node,
			"building TypeORM",
			`cd /mnt/data1/typeorm/ && npm install`,
		); err != nil {
			__antithesis_instrumentation__.Notify(52184)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(52185)
		}
		__antithesis_instrumentation__.Notify(52155)

		t.Status("running TypeORM test suite - approx 12 mins")
		result, err := c.RunWithDetailsSingleNode(ctx, t.L(), node,
			`cd /mnt/data1/typeorm/ && npm test`,
		)
		rawResults := result.Stdout + result.Stderr
		t.L().Printf("Test Results: %s", rawResults)
		if err != nil {
			__antithesis_instrumentation__.Notify(52186)
			if strings.Contains(rawResults, "1 failing") && func() bool {
				__antithesis_instrumentation__.Notify(52188)
				return strings.Contains(rawResults, "Error: Cannot find connection better-sqlite3 because its not defined in any orm configuration files.") == true
			}() == true {
				__antithesis_instrumentation__.Notify(52189)
				err = nil
			} else {
				__antithesis_instrumentation__.Notify(52190)
			}
			__antithesis_instrumentation__.Notify(52187)
			if err != nil {
				__antithesis_instrumentation__.Notify(52191)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(52192)
			}
		} else {
			__antithesis_instrumentation__.Notify(52193)
		}
	}
	__antithesis_instrumentation__.Notify(52139)

	r.Add(registry.TestSpec{
		Name:    "typeorm",
		Owner:   registry.OwnerSQLExperience,
		Cluster: r.MakeClusterSpec(1),
		Tags:    []string{`default`, `orm`},
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(52194)
			runTypeORM(ctx, t, c)
		},
	})
}

const typeORMConfigJSON = `
[
  {
    "skip": true,
    "name": "mysql",
    "type": "mysql",
    "host": "localhost",
    "port": 3306,
    "username": "root",
    "password": "admin",
    "database": "test",
    "logging": false
  },
  {
    "skip": true,
    "name": "mariadb",
    "type": "mariadb",
    "host": "localhost",
    "port": 3307,
    "username": "root",
    "password": "admin",
    "database": "test",
    "logging": false
  },
  {
    "skip": true,
    "name": "sqlite",
    "type": "sqlite",
    "database": "temp/sqlitedb.db",
    "logging": false
  },
  {
    "skip": true,
    "name": "postgres",
    "type": "postgres",
    "host": "localhost",
    "port": 5432,
    "username": "test",
    "password": "test",
    "database": "test",
    "logging": false
  },
  {
    "skip": true,
    "name": "mssql",
    "type": "mssql",
    "host": "localhost",
    "username": "sa",
    "password": "Admin12345",
    "database": "tempdb",
    "logging": false
  },
  {
    "skip": true,
    "name": "oracle",
    "type": "oracle",
    "host": "localhost",
    "username": "system",
    "password": "oracle",
    "port": 1521,
    "sid": "xe.oracle.docker",
    "logging": false
  },
  {
    "skip": false,
    "name": "cockroachdb",
    "type": "cockroachdb",
    "host": "localhost",
    "port": 26257,
    "username": "root",
    "password": "",
    "database": "defaultdb"
  },
  {
    "skip": true,
    "disabledIfNotEnabledImplicitly": true,
    "name": "mongodb",
    "type": "mongodb",
    "database": "test",
    "logging": false,
    "useNewUrlParser": true
  }
]
`
