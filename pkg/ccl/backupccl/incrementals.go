package backupccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/errors"
)

const (
	DefaultIncrementalsSubdir = "incrementals"
	incBackupSubdirGlob       = "/[0-9]*/[0-9]*.[0-9][0-9]/"

	listingDelimDataSlash = "data/"

	URLSeparator = '/'
)

var backupSubdirRE = regexp.MustCompile(`(.*)/([0-9]{4}/[0-9]{2}/[0-9]{2}-[0-9]{6}.[0-9]{2}/?)$`)

func CollectionAndSubdir(path string, subdir string) (string, string) {
	__antithesis_instrumentation__.Notify(9686)
	if subdir != "" {
		__antithesis_instrumentation__.Notify(9689)
		return path, subdir
	} else {
		__antithesis_instrumentation__.Notify(9690)
	}
	__antithesis_instrumentation__.Notify(9687)

	matchResult := backupSubdirRE.FindStringSubmatch(path)
	if matchResult == nil {
		__antithesis_instrumentation__.Notify(9691)
		return path, subdir
	} else {
		__antithesis_instrumentation__.Notify(9692)
	}
	__antithesis_instrumentation__.Notify(9688)
	return matchResult[1], matchResult[2]
}

func FindPriorBackups(
	ctx context.Context, store cloud.ExternalStorage, includeManifest bool,
) ([]string, error) {
	__antithesis_instrumentation__.Notify(9693)
	var prev []string
	if err := store.List(ctx, "", listingDelimDataSlash, func(p string) error {
		__antithesis_instrumentation__.Notify(9695)
		if ok, err := path.Match(incBackupSubdirGlob+backupManifestName, p); err != nil {
			__antithesis_instrumentation__.Notify(9698)
			return err
		} else {
			__antithesis_instrumentation__.Notify(9699)
			if ok {
				__antithesis_instrumentation__.Notify(9700)
				if !includeManifest {
					__antithesis_instrumentation__.Notify(9702)
					p = strings.TrimSuffix(p, "/"+backupManifestName)
				} else {
					__antithesis_instrumentation__.Notify(9703)
				}
				__antithesis_instrumentation__.Notify(9701)
				prev = append(prev, p)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(9704)
			}
		}
		__antithesis_instrumentation__.Notify(9696)
		if ok, err := path.Match(incBackupSubdirGlob+backupOldManifestName, p); err != nil {
			__antithesis_instrumentation__.Notify(9705)
			return err
		} else {
			__antithesis_instrumentation__.Notify(9706)
			if ok {
				__antithesis_instrumentation__.Notify(9707)
				if !includeManifest {
					__antithesis_instrumentation__.Notify(9709)
					p = strings.TrimSuffix(p, "/"+backupOldManifestName)
				} else {
					__antithesis_instrumentation__.Notify(9710)
				}
				__antithesis_instrumentation__.Notify(9708)
				prev = append(prev, p)
			} else {
				__antithesis_instrumentation__.Notify(9711)
			}
		}
		__antithesis_instrumentation__.Notify(9697)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(9712)
		return nil, errors.Wrap(err, "reading previous backup layers")
	} else {
		__antithesis_instrumentation__.Notify(9713)
	}
	__antithesis_instrumentation__.Notify(9694)
	sort.Strings(prev)
	return prev, nil
}

func appendPaths(uris []string, tailDir ...string) ([]string, error) {
	__antithesis_instrumentation__.Notify(9714)
	retval := make([]string, len(uris))
	for i, uri := range uris {
		__antithesis_instrumentation__.Notify(9716)
		parsed, err := url.Parse(uri)
		if err != nil {
			__antithesis_instrumentation__.Notify(9718)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(9719)
		}
		__antithesis_instrumentation__.Notify(9717)
		joinArgs := append([]string{parsed.Path}, tailDir...)
		parsed.Path = JoinURLPath(joinArgs...)
		retval[i] = parsed.String()
	}
	__antithesis_instrumentation__.Notify(9715)
	return retval, nil
}

func JoinURLPath(args ...string) string {
	__antithesis_instrumentation__.Notify(9720)
	argsCopy := make([]string, 0)
	for _, arg := range args {
		__antithesis_instrumentation__.Notify(9725)
		if len(arg) == 0 {
			__antithesis_instrumentation__.Notify(9727)
			continue
		} else {
			__antithesis_instrumentation__.Notify(9728)
		}
		__antithesis_instrumentation__.Notify(9726)

		argsCopy = append(argsCopy, arg)
	}
	__antithesis_instrumentation__.Notify(9721)
	if len(argsCopy) == 0 {
		__antithesis_instrumentation__.Notify(9729)
		return path.Join(argsCopy...)
	} else {
		__antithesis_instrumentation__.Notify(9730)
	}
	__antithesis_instrumentation__.Notify(9722)

	isAbs := false
	if argsCopy[0][0] == URLSeparator {
		__antithesis_instrumentation__.Notify(9731)
		isAbs = true
		argsCopy[0] = argsCopy[0][1:]
	} else {
		__antithesis_instrumentation__.Notify(9732)
	}
	__antithesis_instrumentation__.Notify(9723)
	joined := path.Join(argsCopy...)
	if isAbs {
		__antithesis_instrumentation__.Notify(9733)
		joined = string(URLSeparator) + joined
	} else {
		__antithesis_instrumentation__.Notify(9734)
	}
	__antithesis_instrumentation__.Notify(9724)
	return joined
}

func backupsFromLocation(
	ctx context.Context, user security.SQLUsername, execCfg *sql.ExecutorConfig, loc string,
) ([]string, error) {
	__antithesis_instrumentation__.Notify(9735)
	mkStore := execCfg.DistSQLSrv.ExternalStorageFromURI
	store, err := mkStore(ctx, loc, user)
	if err != nil {
		__antithesis_instrumentation__.Notify(9737)
		return nil, errors.Wrapf(err, "failed to open backup storage location")
	} else {
		__antithesis_instrumentation__.Notify(9738)
	}
	__antithesis_instrumentation__.Notify(9736)
	defer store.Close()
	prev, err := FindPriorBackups(ctx, store, false)
	return prev, err
}

func resolveIncrementalsBackupLocation(
	ctx context.Context,
	user security.SQLUsername,
	execCfg *sql.ExecutorConfig,
	explicitIncrementalCollections []string,
	fullBackupCollections []string,
	subdir string,
) ([]string, error) {
	__antithesis_instrumentation__.Notify(9739)
	if len(explicitIncrementalCollections) > 0 {
		__antithesis_instrumentation__.Notify(9747)
		incPaths, err := appendPaths(explicitIncrementalCollections, subdir)
		if err != nil {
			__antithesis_instrumentation__.Notify(9750)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(9751)
		}
		__antithesis_instrumentation__.Notify(9748)

		_, err = backupsFromLocation(ctx, user, execCfg, incPaths[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(9752)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(9753)
		}
		__antithesis_instrumentation__.Notify(9749)
		return incPaths, nil
	} else {
		__antithesis_instrumentation__.Notify(9754)
	}
	__antithesis_instrumentation__.Notify(9740)

	resolvedIncrementalsBackupLocationOld, err := appendPaths(fullBackupCollections, subdir)
	if err != nil {
		__antithesis_instrumentation__.Notify(9755)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(9756)
	}
	__antithesis_instrumentation__.Notify(9741)

	prevOld, err := backupsFromLocation(ctx, user, execCfg, resolvedIncrementalsBackupLocationOld[0])
	if err != nil {
		__antithesis_instrumentation__.Notify(9757)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(9758)
	}
	__antithesis_instrumentation__.Notify(9742)

	resolvedIncrementalsBackupLocation, err := appendPaths(fullBackupCollections, DefaultIncrementalsSubdir, subdir)
	if err != nil {
		__antithesis_instrumentation__.Notify(9759)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(9760)
	}
	__antithesis_instrumentation__.Notify(9743)

	prev, err := backupsFromLocation(ctx, user, execCfg, resolvedIncrementalsBackupLocation[0])
	if err != nil {
		__antithesis_instrumentation__.Notify(9761)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(9762)
	}
	__antithesis_instrumentation__.Notify(9744)

	if len(prevOld) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(9763)
		return len(prev) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(9764)
		return nil, errors.New(
			"Incremental layers found in both old and new default locations. " +
				"Please choose a location manually with the `incremental_location` parameter.")
	} else {
		__antithesis_instrumentation__.Notify(9765)
	}
	__antithesis_instrumentation__.Notify(9745)

	if len(prevOld) > 0 || func() bool {
		__antithesis_instrumentation__.Notify(9766)
		return !execCfg.Settings.Version.IsActive(ctx, clusterversion.IncrementalBackupSubdir) == true
	}() == true {
		__antithesis_instrumentation__.Notify(9767)
		return resolvedIncrementalsBackupLocationOld, nil
	} else {
		__antithesis_instrumentation__.Notify(9768)
	}
	__antithesis_instrumentation__.Notify(9746)

	return resolvedIncrementalsBackupLocation, nil
}
