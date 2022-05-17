package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"regexp"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/spec"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/errors"
)

var djangoReleaseTagRegex = regexp.MustCompile(`^(?P<major>\d+)\.(?P<minor>\d+)(\.(?P<point>\d+))?$`)
var djangoCockroachDBReleaseTagRegex = regexp.MustCompile(`^(?P<major>\d+)\.(?P<minor>\d+)$`)

var djangoSupportedTag = "cockroach-4.0.x"
var djangoCockroachDBSupportedTag = "4.0.1"

func registerDjango(r registry.Registry) {
	__antithesis_instrumentation__.Notify(47341)
	runDjango := func(
		ctx context.Context,
		t test.Test,
		c cluster.Cluster,
	) {
		__antithesis_instrumentation__.Notify(47343)
		if c.IsLocal() {
			__antithesis_instrumentation__.Notify(47364)
			t.Fatal("cannot be run in local mode")
		} else {
			__antithesis_instrumentation__.Notify(47365)
		}
		__antithesis_instrumentation__.Notify(47344)
		node := c.Node(1)
		t.Status("setting up cockroach")
		c.Put(ctx, t.Cockroach(), "./cockroach", c.All())
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.All())

		version, err := fetchCockroachVersion(ctx, t.L(), c, node[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(47366)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(47367)
		}
		__antithesis_instrumentation__.Notify(47345)

		err = alterZoneConfigAndClusterSettings(ctx, t, version, c, node[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(47368)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(47369)
		}
		__antithesis_instrumentation__.Notify(47346)

		t.Status("cloning django and installing prerequisites")

		if err := repeatRunE(
			ctx, t, c, node, "update apt-get",
			`
				sudo add-apt-repository ppa:deadsnakes/ppa &&
				sudo apt-get -qq update`,
		); err != nil {
			__antithesis_instrumentation__.Notify(47370)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(47371)
		}
		__antithesis_instrumentation__.Notify(47347)

		if err := repeatRunE(
			ctx,
			t,
			c,
			node,
			"install dependencies",
			`sudo apt-get -qq install make python3.8 libpq-dev python3.8-dev gcc python3-virtualenv python3-setuptools python-setuptools build-essential python3.8-distutils python3-apt libmemcached-dev`,
		); err != nil {
			__antithesis_instrumentation__.Notify(47372)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(47373)
		}
		__antithesis_instrumentation__.Notify(47348)

		if err := repeatRunE(
			ctx, t, c, node, "set python3.8 as default", `
    		sudo update-alternatives --install /usr/bin/python3 python3 /usr/bin/python3.5 1
    		sudo update-alternatives --install /usr/bin/python3 python3 /usr/bin/python3.8 2
    		sudo update-alternatives --config python3`,
		); err != nil {
			__antithesis_instrumentation__.Notify(47374)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(47375)
		}
		__antithesis_instrumentation__.Notify(47349)

		if err := repeatRunE(
			ctx, t, c, node, "install pip",
			`curl https://bootstrap.pypa.io/get-pip.py | sudo -H python3.8`,
		); err != nil {
			__antithesis_instrumentation__.Notify(47376)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(47377)
		}
		__antithesis_instrumentation__.Notify(47350)

		if err := repeatRunE(
			ctx, t, c, node, "create virtualenv",
			`virtualenv venv &&
				source venv/bin/activate`,
		); err != nil {
			__antithesis_instrumentation__.Notify(47378)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(47379)
		}
		__antithesis_instrumentation__.Notify(47351)

		if err := repeatRunE(
			ctx,
			t,
			c,
			node,
			"install pytest",
			`pip3 install pytest pytest-xdist psycopg2`,
		); err != nil {
			__antithesis_instrumentation__.Notify(47380)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(47381)
		}
		__antithesis_instrumentation__.Notify(47352)

		if err := repeatRunE(
			ctx, t, c, node, "remove old django", `rm -rf /mnt/data1/django`,
		); err != nil {
			__antithesis_instrumentation__.Notify(47382)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(47383)
		}
		__antithesis_instrumentation__.Notify(47353)

		djangoLatestTag, err := repeatGetLatestTag(
			ctx, t, "django", "django", djangoReleaseTagRegex,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(47384)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(47385)
		}
		__antithesis_instrumentation__.Notify(47354)
		t.L().Printf("Latest Django release is %s.", djangoLatestTag)
		t.L().Printf("Supported Django release is %s.", djangoSupportedTag)

		if err := repeatGitCloneE(
			ctx,
			t,
			c,
			"https://github.com/timgraham/django/",
			"/mnt/data1/django",
			djangoSupportedTag,
			node,
		); err != nil {
			__antithesis_instrumentation__.Notify(47386)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(47387)
		}
		__antithesis_instrumentation__.Notify(47355)

		djangoCockroachDBLatestTag, err := repeatGetLatestTag(
			ctx, t, "cockroachdb", "django-cockroachdb", djangoCockroachDBReleaseTagRegex,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(47388)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(47389)
		}
		__antithesis_instrumentation__.Notify(47356)
		t.L().Printf("Latest django-cockroachdb release is %s.", djangoCockroachDBLatestTag)
		t.L().Printf("Supported django-cockroachdb release is %s.", djangoCockroachDBSupportedTag)

		if err := repeatGitCloneE(
			ctx,
			t,
			c,
			"https://github.com/cockroachdb/django-cockroachdb",
			"/mnt/data1/django/tests/django-cockroachdb",
			djangoCockroachDBSupportedTag,
			node,
		); err != nil {
			__antithesis_instrumentation__.Notify(47390)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(47391)
		}
		__antithesis_instrumentation__.Notify(47357)

		if err := repeatRunE(
			ctx, t, c, node, "install django's dependencies", `
				cd /mnt/data1/django/tests &&
				pip3 install -e .. &&
				pip3 install -r requirements/py3.txt &&
				pip3 install -r requirements/postgres.txt`,
		); err != nil {
			__antithesis_instrumentation__.Notify(47392)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(47393)
		}
		__antithesis_instrumentation__.Notify(47358)

		if err := repeatRunE(
			ctx, t, c, node, "install django-cockroachdb", `
					cd /mnt/data1/django/tests/django-cockroachdb/ &&
					pip3 install .`,
		); err != nil {
			__antithesis_instrumentation__.Notify(47394)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(47395)
		}
		__antithesis_instrumentation__.Notify(47359)

		if err := repeatRunE(
			ctx, t, c, node, "configuring tests to use cockroach",
			fmt.Sprintf(
				"echo \"%s\" > /mnt/data1/django/tests/cockroach_settings.py",
				cockroachDjangoSettings,
			),
		); err != nil {
			__antithesis_instrumentation__.Notify(47396)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(47397)
		}
		__antithesis_instrumentation__.Notify(47360)

		blocklistName, expectedFailureList, ignoredlistName, ignoredlist := djangoBlocklists.getLists(version)
		if expectedFailureList == nil {
			__antithesis_instrumentation__.Notify(47398)
			t.Fatalf("No django blocklist defined for cockroach version %s", version)
		} else {
			__antithesis_instrumentation__.Notify(47399)
		}
		__antithesis_instrumentation__.Notify(47361)
		if ignoredlist == nil {
			__antithesis_instrumentation__.Notify(47400)
			t.Fatalf("No django ignorelist defined for cockroach version %s", version)
		} else {
			__antithesis_instrumentation__.Notify(47401)
		}
		__antithesis_instrumentation__.Notify(47362)
		t.L().Printf("Running cockroach version %s, using blocklist %s, using ignoredlist %s",
			version, blocklistName, ignoredlistName)

		var fullTestResults []byte
		for _, testName := range enabledDjangoTests {
			__antithesis_instrumentation__.Notify(47402)
			t.Status("Running django test app ", testName)
			result, err := c.RunWithDetailsSingleNode(ctx, t.L(), node, fmt.Sprintf(djangoRunTestCmd, testName))

			if err != nil {
				__antithesis_instrumentation__.Notify(47404)

				commandError := (*install.NonZeroExitCode)(nil)
				if !errors.As(err, &commandError) {
					__antithesis_instrumentation__.Notify(47405)
					t.Fatal(err)
				} else {
					__antithesis_instrumentation__.Notify(47406)
				}
			} else {
				__antithesis_instrumentation__.Notify(47407)
			}
			__antithesis_instrumentation__.Notify(47403)

			rawResults := []byte(result.Stdout + result.Stderr)

			fullTestResults = append(fullTestResults, rawResults...)
			t.L().Printf("Test results for app %s: %s", testName, rawResults)
			t.L().Printf("Test stdout for app %s:", testName)
			if err := c.RunE(
				ctx, node, fmt.Sprintf("cd /mnt/data1/django/tests && cat %s.stdout", testName),
			); err != nil {
				__antithesis_instrumentation__.Notify(47408)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(47409)
			}
		}
		__antithesis_instrumentation__.Notify(47363)
		t.Status("collating test results")

		results := newORMTestsResults()
		results.parsePythonUnitTestOutput(fullTestResults, expectedFailureList, ignoredlist)
		results.summarizeAll(
			t, "django", blocklistName, expectedFailureList, version, djangoSupportedTag,
		)
	}
	__antithesis_instrumentation__.Notify(47342)

	r.Add(registry.TestSpec{
		Name:    "django",
		Owner:   registry.OwnerSQLExperience,
		Cluster: r.MakeClusterSpec(1, spec.CPU(16)),
		Tags:    []string{`default`, `orm`},
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(47410)
			runDjango(ctx, t, c)
		},
	})
}

const djangoRunTestCmd = `
cd /mnt/data1/django/tests &&
RUNNING_COCKROACH_BACKEND_TESTS=1 python3 runtests.py %[1]s --settings cockroach_settings --parallel 1 -v 2 > %[1]s.stdout
`

const cockroachDjangoSettings = `
from django.test.runner import DiscoverRunner


DATABASES = {
    'default': {
        'ENGINE': 'django_cockroachdb',
        'NAME': 'django_tests',
        'USER': 'root',
        'PASSWORD': '',
        'HOST': 'localhost',
        'PORT': 26257,
				'DISABLE_COCKROACHDB_TELEMETRY': True,
    },
    'other': {
        'ENGINE': 'django_cockroachdb',
        'NAME': 'django_tests2',
        'USER': 'root',
        'PASSWORD': '',
        'HOST': 'localhost',
        'PORT': 26257,
				'DISABLE_COCKROACHDB_TELEMETRY': True,
    },
}
SECRET_KEY = 'django_tests_secret_key'
PASSWORD_HASHERS = [
    'django.contrib.auth.hashers.MD5PasswordHasher',
]
TEST_RUNNER = '.cockroach_settings.NonDescribingDiscoverRunner'
DEFAULT_AUTO_FIELD = 'django.db.models.AutoField'

class NonDescribingDiscoverRunner(DiscoverRunner):
    def get_test_runner_kwargs(self):
        return {
            'failfast': self.failfast,
            'resultclass': self.get_resultclass(),
            'verbosity': self.verbosity,
            'descriptions': False,
        }

USE_TZ = False
`
