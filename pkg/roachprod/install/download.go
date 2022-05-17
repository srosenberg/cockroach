package install

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	_ "embed"
	"fmt"
	"net/url"
	"path"
	"path/filepath"

	"github.com/cockroachdb/cockroach/pkg/roachprod/logger"
)

const (
	gcsCacheBaseURL = "https://storage.googleapis.com/cockroach-fixtures/tools/"
)

var downloadScript string

func Download(
	ctx context.Context,
	l *logger.Logger,
	c *SyncedCluster,
	sourceURLStr string,
	sha string,
	dest string,
) error {
	__antithesis_instrumentation__.Notify(181497)

	sourceURL, err := url.Parse(sourceURLStr)
	if err != nil {
		__antithesis_instrumentation__.Notify(181504)
		return err
	} else {
		__antithesis_instrumentation__.Notify(181505)
	}
	__antithesis_instrumentation__.Notify(181498)

	basename := path.Base(sourceURL.Path)

	cacheBasename := fmt.Sprintf("%s-%s", sha, basename)

	gcsCacheURL, err := url.Parse(path.Join(gcsCacheBaseURL, cacheBasename))
	if err != nil {
		__antithesis_instrumentation__.Notify(181506)
		return err
	} else {
		__antithesis_instrumentation__.Notify(181507)
	}
	__antithesis_instrumentation__.Notify(181499)

	if dest == "" {
		__antithesis_instrumentation__.Notify(181508)
		dest = path.Join("./", basename)
	} else {
		__antithesis_instrumentation__.Notify(181509)
	}
	__antithesis_instrumentation__.Notify(181500)

	downloadNodes := c.Nodes
	if c.IsLocal() {
		__antithesis_instrumentation__.Notify(181510)
		downloadNodes = downloadNodes[:1]
	} else {
		__antithesis_instrumentation__.Notify(181511)
	}
	__antithesis_instrumentation__.Notify(181501)

	downloadCmd := fmt.Sprintf(downloadScript,
		sourceURL.String(),
		gcsCacheURL.String(),
		sha,
		dest,
	)
	if err := c.Run(ctx, l, l.Stdout, l.Stderr,
		downloadNodes,
		fmt.Sprintf("downloading %s", basename),
		downloadCmd,
	); err != nil {
		__antithesis_instrumentation__.Notify(181512)
		return err
	} else {
		__antithesis_instrumentation__.Notify(181513)
	}
	__antithesis_instrumentation__.Notify(181502)

	if c.IsLocal() && func() bool {
		__antithesis_instrumentation__.Notify(181514)
		return !filepath.IsAbs(dest) == true
	}() == true {
		__antithesis_instrumentation__.Notify(181515)
		src := filepath.Join(c.localVMDir(downloadNodes[0]), dest)
		cpCmd := fmt.Sprintf(`cp "%s" "%s"`, src, dest)
		return c.Run(ctx, l, l.Stdout, l.Stderr, c.Nodes[1:], "copying to remaining nodes", cpCmd)
	} else {
		__antithesis_instrumentation__.Notify(181516)
	}
	__antithesis_instrumentation__.Notify(181503)

	return nil
}
