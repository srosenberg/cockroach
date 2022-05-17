package cluster

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/cockroachdb/cockroach/pkg/security"
)

const certsDir = ".localcluster.certs"

var absCertsDir string

const keyLen = 2048

func AbsCertsDir() string {
	__antithesis_instrumentation__.Notify(8)
	return absCertsDir
}

func GenerateCerts(ctx context.Context) func() {
	__antithesis_instrumentation__.Notify(9)
	var err error

	absCertsDir, err = filepath.Abs(certsDir)
	if err != nil {
		__antithesis_instrumentation__.Notify(12)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(13)
	}
	__antithesis_instrumentation__.Notify(10)
	maybePanic(os.RemoveAll(certsDir))

	maybePanic(security.CreateCAPair(
		certsDir, filepath.Join(certsDir, security.EmbeddedCAKey),
		keyLen, 96*time.Hour, false, false))

	maybePanic(security.CreateClientPair(
		certsDir, filepath.Join(certsDir, security.EmbeddedCAKey),
		2048, 48*time.Hour, false, security.RootUserName(), true))

	maybePanic(security.CreateClientPair(
		certsDir, filepath.Join(certsDir, security.EmbeddedCAKey),
		1024, 48*time.Hour, false, security.TestUserName(), true))

	maybePanic(security.CreateNodePair(
		certsDir, filepath.Join(certsDir, security.EmbeddedCAKey),
		keyLen, 48*time.Hour, false, []string{"localhost", "cockroach"}))

	{
		__antithesis_instrumentation__.Notify(14)
		execCmd("openssl", "pkcs12", "-export", "-password", "pass:",
			"-in", filepath.Join(certsDir, "client.root.crt"),
			"-inkey", filepath.Join(certsDir, "client.root.key"),
			"-out", filepath.Join(certsDir, "client.root.pk12"))
	}
	__antithesis_instrumentation__.Notify(11)

	return func() { __antithesis_instrumentation__.Notify(15); _ = os.RemoveAll(certsDir) }
}

var _ = GenerateCerts

func execCmd(args ...string) {
	__antithesis_instrumentation__.Notify(16)
	cmd := exec.Command(args[0], args[1:]...)
	if out, err := cmd.CombinedOutput(); err != nil {
		__antithesis_instrumentation__.Notify(17)
		panic(fmt.Sprintf("error: %s: %s\nout: %s\n", args, err, out))
	} else {
		__antithesis_instrumentation__.Notify(18)
	}
}
