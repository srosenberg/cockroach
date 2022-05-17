package config

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"os/user"
	"regexp"

	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

var (
	Binary = "cockroach"

	SlackToken string

	OSUser *user.User

	Quiet = true

	MaxConcurrency = 32

	CockroachDevLicense = envutil.EnvOrDefaultString("COCKROACH_DEV_LICENSE", "")
)

func init() {
	var err error
	OSUser, err = user.Current()
	if err != nil {
		log.Fatalf(context.Background(), "Unable to determine OS user: %v", err)
	}
}

const (
	DefaultDebugDir = "${HOME}/.roachprod/debug"

	EmailDomain = "@cockroachlabs.com"

	Local = "local"

	ClustersDir = "${HOME}/.roachprod/clusters"

	SharedUser = "ubuntu"

	MemoryMax = "95%"

	DefaultSQLPort = 26257

	DefaultAdminUIPort = 26258

	DefaultNumFilesLimit = 65 << 13
)

func IsLocalClusterName(clusterName string) bool {
	__antithesis_instrumentation__.Notify(180405)
	return localClusterRegex.MatchString(clusterName)
}

var localClusterRegex = regexp.MustCompile(`^local(|-[a-zA-Z0-9\-]+)$`)
