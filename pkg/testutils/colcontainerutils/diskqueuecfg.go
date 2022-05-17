package colcontainerutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/sql/colcontainer"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/fs"
	"github.com/cockroachdb/cockroach/pkg/testutils"
)

const inMemDirName = "testing"

func NewTestingDiskQueueCfg(t testing.TB, inMem bool) (colcontainer.DiskQueueCfg, func()) {
	__antithesis_instrumentation__.Notify(644030)
	t.Helper()

	var (
		cfg       colcontainer.DiskQueueCfg
		cleanup   func()
		testingFS fs.FS
		path      string
	)

	if inMem {
		__antithesis_instrumentation__.Notify(644034)
		ngn := storage.NewDefaultInMemForTesting()
		testingFS = ngn.(fs.FS)
		if err := testingFS.MkdirAll(inMemDirName); err != nil {
			__antithesis_instrumentation__.Notify(644036)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(644037)
		}
		__antithesis_instrumentation__.Notify(644035)
		path = inMemDirName
		cleanup = ngn.Close
	} else {
		__antithesis_instrumentation__.Notify(644038)
		tempPath, dirCleanup := testutils.TempDir(t)
		path = tempPath
		ngn, err := storage.Open(
			context.Background(),
			storage.Filesystem(tempPath),
			storage.CacheSize(0))
		if err != nil {
			__antithesis_instrumentation__.Notify(644040)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(644041)
		}
		__antithesis_instrumentation__.Notify(644039)
		testingFS = ngn
		cleanup = func() {
			__antithesis_instrumentation__.Notify(644042)
			ngn.Close()
			dirCleanup()
		}
	}
	__antithesis_instrumentation__.Notify(644031)
	cfg.FS = testingFS
	cfg.GetPather = colcontainer.GetPatherFunc(func(context.Context) string { __antithesis_instrumentation__.Notify(644043); return path })
	__antithesis_instrumentation__.Notify(644032)
	if err := cfg.EnsureDefaults(); err != nil {
		__antithesis_instrumentation__.Notify(644044)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(644045)
	}
	__antithesis_instrumentation__.Notify(644033)

	return cfg, cleanup
}
