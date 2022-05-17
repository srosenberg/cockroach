package install

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/roachprod/logger"
	"github.com/cockroachdb/cockroach/pkg/util/httputil"
	"github.com/cockroachdb/errors"
)

const (
	edgeBinaryServer    = "https://edge-binaries.cockroachdb.com"
	releaseBinaryServer = "https://s3.amazonaws.com/binaries.cockroachdb.com/"
)

type archInfo struct {
	DebugArchitecture string

	ReleaseArchitecture string

	LibraryExtension string
}

var (
	linuxArchInfo = archInfo{
		DebugArchitecture:   "linux-gnu-amd64",
		ReleaseArchitecture: "linux-amd64",
		LibraryExtension:    ".so",
	}
	darwinArchInfo = archInfo{
		DebugArchitecture:   "darwin-amd64",
		ReleaseArchitecture: "darwin-10.9-amd64",
		LibraryExtension:    ".dylib",
	}
	windowsArchInfo = archInfo{
		DebugArchitecture:   "windows-amd64",
		ReleaseArchitecture: "windows-6.2-amd64",
		LibraryExtension:    ".dll",
	}

	crdbLibraries = []string{"libgeos", "libgeos_c"}
)

func archInfoForOS(os string) (archInfo, error) {
	__antithesis_instrumentation__.Notify(181746)
	switch os {
	case "linux":
		__antithesis_instrumentation__.Notify(181747)
		return linuxArchInfo, nil
	case "darwin":
		__antithesis_instrumentation__.Notify(181748)
		return darwinArchInfo, nil
	case "windows":
		__antithesis_instrumentation__.Notify(181749)
		return windowsArchInfo, nil
	default:
		__antithesis_instrumentation__.Notify(181750)
		return archInfo{}, errors.Errorf("no release architecture information for %q", os)
	}
}

func getEdgeURL(urlPathBase, SHA, arch string, ext string) (*url.URL, error) {
	__antithesis_instrumentation__.Notify(181751)
	edgeBinaryLocation, err := url.Parse(edgeBinaryServer)
	if err != nil {
		__antithesis_instrumentation__.Notify(181755)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(181756)
	}
	__antithesis_instrumentation__.Notify(181752)
	edgeBinaryLocation.Path = urlPathBase

	if len(arch) > 0 {
		__antithesis_instrumentation__.Notify(181757)
		edgeBinaryLocation.Path += "." + arch
	} else {
		__antithesis_instrumentation__.Notify(181758)
	}
	__antithesis_instrumentation__.Notify(181753)

	if len(SHA) > 0 {
		__antithesis_instrumentation__.Notify(181759)
		edgeBinaryLocation.Path += "." + SHA + ext
	} else {
		__antithesis_instrumentation__.Notify(181760)
		edgeBinaryLocation.Path += ext + ".LATEST"

		resp, err := httputil.Head(context.TODO(), edgeBinaryLocation.String())
		if err != nil {
			__antithesis_instrumentation__.Notify(181762)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(181763)
		}
		__antithesis_instrumentation__.Notify(181761)
		defer resp.Body.Close()
		edgeBinaryLocation = resp.Request.URL
	}
	__antithesis_instrumentation__.Notify(181754)

	return edgeBinaryLocation, nil
}

func shaFromEdgeURL(u *url.URL) string {
	__antithesis_instrumentation__.Notify(181764)
	urlSplit := strings.Split(u.Path, ".")
	return urlSplit[len(urlSplit)-1]
}

func cockroachReleaseURL(version string, arch string) (*url.URL, error) {
	__antithesis_instrumentation__.Notify(181765)
	binURL, err := url.Parse(releaseBinaryServer)
	if err != nil {
		__antithesis_instrumentation__.Notify(181767)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(181768)
	}
	__antithesis_instrumentation__.Notify(181766)
	binURL.Path += fmt.Sprintf("cockroach-%s.%s.tgz", version, arch)
	return binURL, nil
}

func StageApplication(
	ctx context.Context,
	l *logger.Logger,
	c *SyncedCluster,
	applicationName string,
	version string,
	os string,
	destDir string,
) error {
	__antithesis_instrumentation__.Notify(181769)
	archInfo, err := archInfoForOS(os)
	if err != nil {
		__antithesis_instrumentation__.Notify(181771)
		return err
	} else {
		__antithesis_instrumentation__.Notify(181772)
	}
	__antithesis_instrumentation__.Notify(181770)

	switch applicationName {
	case "cockroach":
		__antithesis_instrumentation__.Notify(181773)
		sha, err := StageRemoteBinary(
			ctx, l, c, applicationName, "cockroach/cockroach", version, archInfo.DebugArchitecture, destDir,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(181779)
			return err
		} else {
			__antithesis_instrumentation__.Notify(181780)
		}
		__antithesis_instrumentation__.Notify(181774)

		for _, library := range crdbLibraries {
			__antithesis_instrumentation__.Notify(181781)
			if err := StageOptionalRemoteLibrary(
				ctx,
				l,
				c,
				library,
				fmt.Sprintf("cockroach/lib/%s", library),
				sha,
				archInfo.DebugArchitecture,
				archInfo.LibraryExtension,
				destDir,
			); err != nil {
				__antithesis_instrumentation__.Notify(181782)
				return err
			} else {
				__antithesis_instrumentation__.Notify(181783)
			}
		}
		__antithesis_instrumentation__.Notify(181775)
		return nil
	case "workload":
		__antithesis_instrumentation__.Notify(181776)
		_, err := StageRemoteBinary(
			ctx, l, c, applicationName, "cockroach/workload", version, "", destDir,
		)
		return err
	case "release":
		__antithesis_instrumentation__.Notify(181777)
		return StageCockroachRelease(ctx, l, c, version, archInfo.ReleaseArchitecture, destDir)
	default:
		__antithesis_instrumentation__.Notify(181778)
		return fmt.Errorf("unknown application %s", applicationName)
	}
}

func URLsForApplication(application string, version string, os string) ([]*url.URL, error) {
	__antithesis_instrumentation__.Notify(181784)
	archInfo, err := archInfoForOS(os)
	if err != nil {
		__antithesis_instrumentation__.Notify(181786)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(181787)
	}
	__antithesis_instrumentation__.Notify(181785)

	switch application {
	case "cockroach":
		__antithesis_instrumentation__.Notify(181788)
		u, err := getEdgeURL("cockroach/cockroach", version, archInfo.DebugArchitecture, "")
		if err != nil {
			__antithesis_instrumentation__.Notify(181796)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(181797)
		}
		__antithesis_instrumentation__.Notify(181789)

		urls := []*url.URL{u}
		sha := shaFromEdgeURL(u)
		for _, library := range crdbLibraries {
			__antithesis_instrumentation__.Notify(181798)
			u, err := getEdgeURL(
				fmt.Sprintf("cockroach/lib/%s", library),
				sha,
				archInfo.DebugArchitecture,
				archInfo.LibraryExtension,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(181800)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(181801)
			}
			__antithesis_instrumentation__.Notify(181799)
			urls = append(urls, u)
		}
		__antithesis_instrumentation__.Notify(181790)
		return urls, nil
	case "workload":
		__antithesis_instrumentation__.Notify(181791)
		u, err := getEdgeURL("cockroach/workload", version, "", "")
		if err != nil {
			__antithesis_instrumentation__.Notify(181802)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(181803)
		}
		__antithesis_instrumentation__.Notify(181792)
		return []*url.URL{u}, nil
	case "release":
		__antithesis_instrumentation__.Notify(181793)
		u, err := cockroachReleaseURL(version, archInfo.ReleaseArchitecture)
		if err != nil {
			__antithesis_instrumentation__.Notify(181804)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(181805)
		}
		__antithesis_instrumentation__.Notify(181794)
		return []*url.URL{u}, nil
	default:
		__antithesis_instrumentation__.Notify(181795)
		return nil, fmt.Errorf("unknown application %s", application)
	}
}

func StageRemoteBinary(
	ctx context.Context,
	l *logger.Logger,
	c *SyncedCluster,
	applicationName, urlPathBase, SHA, arch, dir string,
) (string, error) {
	__antithesis_instrumentation__.Notify(181806)
	binURL, err := getEdgeURL(urlPathBase, SHA, arch, "")
	if err != nil {
		__antithesis_instrumentation__.Notify(181808)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(181809)
	}
	__antithesis_instrumentation__.Notify(181807)
	l.Printf("Resolved binary url for %s: %s", applicationName, binURL)
	target := filepath.Join(dir, applicationName)
	cmdStr := fmt.Sprintf(
		`curl -sfSL -o "%s" "%s" && chmod 755 %s`, target, binURL, target,
	)
	return shaFromEdgeURL(binURL), c.Run(
		ctx, l, l.Stdout, l.Stderr, c.Nodes, fmt.Sprintf("staging binary (%s)", applicationName), cmdStr,
	)
}

func StageOptionalRemoteLibrary(
	ctx context.Context,
	l *logger.Logger,
	c *SyncedCluster,
	libraryName, urlPathBase, SHA, arch, ext, dir string,
) error {
	__antithesis_instrumentation__.Notify(181810)
	url, err := getEdgeURL(urlPathBase, SHA, arch, ext)
	if err != nil {
		__antithesis_instrumentation__.Notify(181812)
		return err
	} else {
		__antithesis_instrumentation__.Notify(181813)
	}
	__antithesis_instrumentation__.Notify(181811)
	libDir := filepath.Join(dir, "lib")
	target := filepath.Join(libDir, libraryName+ext)
	l.Printf("Resolved library url for %s: %s", libraryName, url)
	cmdStr := fmt.Sprintf(
		`mkdir -p "%s" && \
curl -sfSL -o "%s" "%s" 2>/dev/null || echo 'optional library %s not found; continuing...'`,
		libDir,
		target,
		url,
		libraryName+ext,
	)
	return c.Run(
		ctx, l, l.Stdout, l.Stderr, c.Nodes, fmt.Sprintf("staging library (%s)", libraryName), cmdStr,
	)
}

func StageCockroachRelease(
	ctx context.Context, l *logger.Logger, c *SyncedCluster, version, arch, dir string,
) error {
	__antithesis_instrumentation__.Notify(181814)
	if len(version) == 0 {
		__antithesis_instrumentation__.Notify(181817)
		return fmt.Errorf(
			"release application cannot be staged without specifying a specific version",
		)
	} else {
		__antithesis_instrumentation__.Notify(181818)
	}
	__antithesis_instrumentation__.Notify(181815)
	binURL, err := cockroachReleaseURL(version, arch)
	if err != nil {
		__antithesis_instrumentation__.Notify(181819)
		return err
	} else {
		__antithesis_instrumentation__.Notify(181820)
	}
	__antithesis_instrumentation__.Notify(181816)
	l.Printf("Resolved release url for cockroach version %s: %s", version, binURL)

	cmdStr := fmt.Sprintf(`
tmpdir="$(mktemp -d /tmp/cockroach-release.XXX)" && \
dir=%s && \
curl -f -s -S -o- %s | tar xfz - -C "${tmpdir}" --strip-components 1 && \
mv ${tmpdir}/cockroach ${dir}/cockroach && \
mkdir -p ${dir}/lib && \
if [ -d ${tmpdir}/lib ]; then mv ${tmpdir}/lib/* ${dir}/lib; fi && \
chmod 755 ${dir}/cockroach
`, dir, binURL)
	return c.Run(
		ctx, l, l.Stdout, l.Stderr, c.Nodes, "staging cockroach release binary", cmdStr,
	)
}
