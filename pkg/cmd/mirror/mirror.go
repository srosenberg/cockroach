package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/cockroachdb/cockroach/pkg/build/bazel"
	"github.com/cockroachdb/cockroach/pkg/build/starlarkutil"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/googleapi"
)

const gcpBucket = "cockroach-godeps"

type downloadedModule struct {
	Path    string `json:"Path"`
	Sum     string `json:"Sum"`
	Version string `json:"Version"`
	Zip     string `json:"Zip"`
}

type listedModule struct {
	Path    string        `json:"Path"`
	Version string        `json:"Version"`
	Replace *listedModule `json:"Replace,omitempty"`
}

func canMirror() bool {
	__antithesis_instrumentation__.Notify(41370)
	return envutil.EnvOrDefaultBool("COCKROACH_BAZEL_CAN_MIRROR", false)
}

func formatSubURL(path, version string) string {
	__antithesis_instrumentation__.Notify(41371)
	return fmt.Sprintf("gomod/%s/%s-%s.zip", path, modulePathToBazelRepoName(path), version)
}

func formatURL(path, version string) string {
	__antithesis_instrumentation__.Notify(41372)
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s",
		gcpBucket, formatSubURL(path, version))
}

func getSha256OfFile(path string) (string, error) {
	__antithesis_instrumentation__.Notify(41373)
	f, err := os.Open(path)
	if err != nil {
		__antithesis_instrumentation__.Notify(41376)
		return "", fmt.Errorf("failed to open %s: %w", path, err)
	} else {
		__antithesis_instrumentation__.Notify(41377)
	}
	__antithesis_instrumentation__.Notify(41374)
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		__antithesis_instrumentation__.Notify(41378)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(41379)
	}
	__antithesis_instrumentation__.Notify(41375)
	return hex.EncodeToString(h.Sum(nil)), nil
}

func copyFile(src, dst string) error {
	__antithesis_instrumentation__.Notify(41380)
	in, err := os.Open(src)
	if err != nil {
		__antithesis_instrumentation__.Notify(41383)
		return err
	} else {
		__antithesis_instrumentation__.Notify(41384)
	}
	__antithesis_instrumentation__.Notify(41381)
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		__antithesis_instrumentation__.Notify(41385)
		return err
	} else {
		__antithesis_instrumentation__.Notify(41386)
	}
	__antithesis_instrumentation__.Notify(41382)
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func uploadFile(ctx context.Context, client *storage.Client, localPath, remotePath string) error {
	__antithesis_instrumentation__.Notify(41387)
	in, err := os.Open(localPath)
	if err != nil {
		__antithesis_instrumentation__.Notify(41391)
		return fmt.Errorf("failed to open %s: %w", localPath, err)
	} else {
		__antithesis_instrumentation__.Notify(41392)
	}
	__antithesis_instrumentation__.Notify(41388)
	defer in.Close()
	out := client.Bucket(gcpBucket).Object(remotePath).If(storage.Conditions{DoesNotExist: true}).NewWriter(ctx)
	if _, err := io.Copy(out, in); err != nil {
		__antithesis_instrumentation__.Notify(41393)
		return err
	} else {
		__antithesis_instrumentation__.Notify(41394)
	}
	__antithesis_instrumentation__.Notify(41389)
	if err := out.Close(); err != nil {
		__antithesis_instrumentation__.Notify(41395)
		var gerr *googleapi.Error
		if errors.As(err, &gerr) {
			__antithesis_instrumentation__.Notify(41397)
			if gerr.Code == http.StatusPreconditionFailed {
				__antithesis_instrumentation__.Notify(41399)

				return nil
			} else {
				__antithesis_instrumentation__.Notify(41400)
			}
			__antithesis_instrumentation__.Notify(41398)
			return gerr
		} else {
			__antithesis_instrumentation__.Notify(41401)
		}
		__antithesis_instrumentation__.Notify(41396)
		return err
	} else {
		__antithesis_instrumentation__.Notify(41402)
	}
	__antithesis_instrumentation__.Notify(41390)
	return nil
}

func createTmpDir() (tmpdir string, err error) {
	__antithesis_instrumentation__.Notify(41403)
	tmpdir, err = bazel.NewTmpDir("gomirror")
	if err != nil {
		__antithesis_instrumentation__.Notify(41408)
		return
	} else {
		__antithesis_instrumentation__.Notify(41409)
	}
	__antithesis_instrumentation__.Notify(41404)
	gomod, err := bazel.Runfile("go.mod")
	if err != nil {
		__antithesis_instrumentation__.Notify(41410)
		return
	} else {
		__antithesis_instrumentation__.Notify(41411)
	}
	__antithesis_instrumentation__.Notify(41405)
	gosum, err := bazel.Runfile("go.sum")
	if err != nil {
		__antithesis_instrumentation__.Notify(41412)
		return
	} else {
		__antithesis_instrumentation__.Notify(41413)
	}
	__antithesis_instrumentation__.Notify(41406)
	err = copyFile(gomod, filepath.Join(tmpdir, "go.mod"))
	if err != nil {
		__antithesis_instrumentation__.Notify(41414)
		return
	} else {
		__antithesis_instrumentation__.Notify(41415)
	}
	__antithesis_instrumentation__.Notify(41407)
	err = copyFile(gosum, filepath.Join(tmpdir, "go.sum"))
	return
}

func downloadZips(tmpdir string) (map[string]downloadedModule, error) {
	__antithesis_instrumentation__.Notify(41416)
	gobin, err := bazel.Runfile("bin/go")
	if err != nil {
		__antithesis_instrumentation__.Notify(41420)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(41421)
	}
	__antithesis_instrumentation__.Notify(41417)
	cmd := exec.Command(gobin, "mod", "download", "-json")
	cmd.Dir = tmpdir
	jsonBytes, err := cmd.Output()
	if err != nil {
		__antithesis_instrumentation__.Notify(41422)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(41423)
	}
	__antithesis_instrumentation__.Notify(41418)
	var jsonBuilder strings.Builder
	ret := make(map[string]downloadedModule)
	for _, line := range strings.Split(string(jsonBytes), "\n") {
		__antithesis_instrumentation__.Notify(41424)
		jsonBuilder.WriteString(line)
		if strings.HasPrefix(line, "}") {
			__antithesis_instrumentation__.Notify(41425)
			var mod downloadedModule
			if err := json.Unmarshal([]byte(jsonBuilder.String()), &mod); err != nil {
				__antithesis_instrumentation__.Notify(41427)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(41428)
			}
			__antithesis_instrumentation__.Notify(41426)
			ret[mod.Path] = mod
			jsonBuilder.Reset()
		} else {
			__antithesis_instrumentation__.Notify(41429)
		}
	}
	__antithesis_instrumentation__.Notify(41419)
	return ret, nil
}

func listAllModules(tmpdir string) (map[string]listedModule, error) {
	__antithesis_instrumentation__.Notify(41430)
	gobin, err := bazel.Runfile("bin/go")
	if err != nil {
		__antithesis_instrumentation__.Notify(41434)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(41435)
	}
	__antithesis_instrumentation__.Notify(41431)
	cmd := exec.Command(gobin, "list", "-mod=readonly", "-m", "-json", "all")
	cmd.Dir = tmpdir
	jsonBytes, err := cmd.Output()
	if err != nil {
		__antithesis_instrumentation__.Notify(41436)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(41437)
	}
	__antithesis_instrumentation__.Notify(41432)
	ret := make(map[string]listedModule)
	var jsonBuilder strings.Builder
	for _, line := range strings.Split(string(jsonBytes), "\n") {
		__antithesis_instrumentation__.Notify(41438)
		jsonBuilder.WriteString(line)
		if strings.HasPrefix(line, "}") {
			__antithesis_instrumentation__.Notify(41439)
			var mod listedModule
			if err := json.Unmarshal([]byte(jsonBuilder.String()), &mod); err != nil {
				__antithesis_instrumentation__.Notify(41442)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(41443)
			}
			__antithesis_instrumentation__.Notify(41440)
			jsonBuilder.Reset()

			if mod.Path == "github.com/cockroachdb/cockroach" {
				__antithesis_instrumentation__.Notify(41444)
				continue
			} else {
				__antithesis_instrumentation__.Notify(41445)
			}
			__antithesis_instrumentation__.Notify(41441)
			ret[mod.Path] = mod
		} else {
			__antithesis_instrumentation__.Notify(41446)
		}
	}
	__antithesis_instrumentation__.Notify(41433)
	return ret, nil
}

func getExistingMirrors() (map[string]starlarkutil.DownloadableArtifact, error) {
	__antithesis_instrumentation__.Notify(41447)
	depsbzl, err := bazel.Runfile("DEPS.bzl")
	if err != nil {
		__antithesis_instrumentation__.Notify(41449)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(41450)
	}
	__antithesis_instrumentation__.Notify(41448)
	return starlarkutil.ListArtifactsInDepsBzl(depsbzl)
}

func mungeBazelRepoNameComponent(component string) string {
	__antithesis_instrumentation__.Notify(41451)
	component = strings.ReplaceAll(component, "-", "_")
	component = strings.ReplaceAll(component, ".", "_")
	return strings.ToLower(component)
}

func modulePathToBazelRepoName(mod string) string {
	__antithesis_instrumentation__.Notify(41452)
	components := strings.Split(mod, "/")
	head := strings.Split(components[0], ".")
	for i, j := 0, len(head)-1; i < j; i, j = i+1, j-1 {
		__antithesis_instrumentation__.Notify(41455)
		head[i], head[j] = mungeBazelRepoNameComponent(head[j]), mungeBazelRepoNameComponent(head[i])
	}
	__antithesis_instrumentation__.Notify(41453)
	for index, component := range components {
		__antithesis_instrumentation__.Notify(41456)
		if index == 0 {
			__antithesis_instrumentation__.Notify(41458)
			continue
		} else {
			__antithesis_instrumentation__.Notify(41459)
		}
		__antithesis_instrumentation__.Notify(41457)
		components[index] = mungeBazelRepoNameComponent(component)
	}
	__antithesis_instrumentation__.Notify(41454)
	return strings.Join(append(head, components[1:]...), "_")
}

func dumpPatchArgsForRepo(repoName string) error {
	__antithesis_instrumentation__.Notify(41460)
	runfiles, err := bazel.RunfilesPath()
	if err != nil {
		__antithesis_instrumentation__.Notify(41463)
		return err
	} else {
		__antithesis_instrumentation__.Notify(41464)
	}
	__antithesis_instrumentation__.Notify(41461)
	candidate := filepath.Join(runfiles, "build", "patches", repoName+".patch")
	if _, err := os.Stat(candidate); err == nil {
		__antithesis_instrumentation__.Notify(41465)
		fmt.Printf(`        patch_args = ["-p1"],
        patches = [
            "@cockroach//build/patches:%s.patch",
        ],
`, repoName)
	} else {
		__antithesis_instrumentation__.Notify(41466)
		if !os.IsNotExist(err) {
			__antithesis_instrumentation__.Notify(41467)
			return err
		} else {
			__antithesis_instrumentation__.Notify(41468)
		}
	}
	__antithesis_instrumentation__.Notify(41462)
	return nil
}

func buildFileProtoModeForRepo(repoName string) string {
	__antithesis_instrumentation__.Notify(41469)
	if repoName == "com_github_prometheus_client_model" {
		__antithesis_instrumentation__.Notify(41471)
		return "package"
	} else {
		__antithesis_instrumentation__.Notify(41472)
	}
	__antithesis_instrumentation__.Notify(41470)
	return "disable_global"
}

func dumpBuildDirectivesForRepo(repoName string) {
	__antithesis_instrumentation__.Notify(41473)
	if repoName == "com_github_cockroachdb_pebble" {
		__antithesis_instrumentation__.Notify(41474)
		fmt.Printf(`        build_directives = ["gazelle:build_tags invariants"],
`)
	} else {
		__antithesis_instrumentation__.Notify(41475)
	}
}

func dumpBuildNamingConventionArgsForRepo(repoName string) {
	__antithesis_instrumentation__.Notify(41476)
	if repoName == "com_github_envoyproxy_protoc_gen_validate" || func() bool {
		__antithesis_instrumentation__.Notify(41477)
		return repoName == "com_github_grpc_ecosystem_grpc_gateway" == true
	}() == true {
		__antithesis_instrumentation__.Notify(41478)
		fmt.Printf("        build_naming_convention = \"go_default_library\",\n")
	} else {
		__antithesis_instrumentation__.Notify(41479)
	}
}

func dumpNewDepsBzl(
	listed map[string]listedModule,
	downloaded map[string]downloadedModule,
	existingMirrors map[string]starlarkutil.DownloadableArtifact,
) error {
	__antithesis_instrumentation__.Notify(41480)
	var sorted []string
	repoNameToModPath := make(map[string]string)
	for _, mod := range listed {
		__antithesis_instrumentation__.Notify(41486)
		repoName := modulePathToBazelRepoName(mod.Path)
		sorted = append(sorted, repoName)
		repoNameToModPath[repoName] = mod.Path
	}
	__antithesis_instrumentation__.Notify(41481)
	sort.Strings(sorted)

	ctx := context.Background()
	var client *storage.Client
	if canMirror() {
		__antithesis_instrumentation__.Notify(41487)
		var err error
		client, err = storage.NewClient(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(41488)
			return err
		} else {
			__antithesis_instrumentation__.Notify(41489)
		}
	} else {
		__antithesis_instrumentation__.Notify(41490)
	}
	__antithesis_instrumentation__.Notify(41482)
	g, ctx := errgroup.WithContext(ctx)

	fmt.Println(`load("@bazel_gazelle//:deps.bzl", "go_repository")

# PRO-TIP: You can inject temorary changes to any of these dependencies by
# by pointing to an alternate remote to clone from. Delete the ` + "`sha256`" + `,
# ` + "`strip_prefix`, and ` + `urls`" + ` parameters, and add ` + "`vcs = \"git\"`" + ` as well as a
# custom ` + "`remote` and `commit`" + `. For example:
#     go_repository(
#        name = "com_github_cockroachdb_sentry_go",
#        build_file_proto_mode = "disable_global",
#        importpath = "github.com/cockroachdb/sentry-go",
#        vcs = "git",
#        remote = "https://github.com/rickystewart/sentry-go",  # Custom fork.
#        commit = "6c8e10aca9672de108063d4953399bd331b54037",  # Custom commit.
#    )
# The ` + "`remote` " + `can be EITHER a URL, or an absolute local path to a clone, such
# as ` + "`/Users/ricky/go/src/github.com/cockroachdb/sentry-go`" + `. Bazel will clone
# from the remote and check out the commit you specify.

def go_deps():
    # NOTE: We ensure that we pin to these specific dependencies by calling
    # this function FIRST, before calls to pull in dependencies for
    # third-party libraries (e.g. rules_go, gazelle, etc.)`)
	for _, repoName := range sorted {
		__antithesis_instrumentation__.Notify(41491)
		path := repoNameToModPath[repoName]
		mod := listed[path]
		replaced := &mod
		if mod.Replace != nil {
			__antithesis_instrumentation__.Notify(41495)
			replaced = mod.Replace
		} else {
			__antithesis_instrumentation__.Notify(41496)
		}
		__antithesis_instrumentation__.Notify(41492)
		fmt.Printf(`    go_repository(
        name = "%s",
`, repoName)
		dumpBuildDirectivesForRepo(repoName)
		fmt.Printf(`        build_file_proto_mode = "%s",
`, buildFileProtoModeForRepo(repoName))
		dumpBuildNamingConventionArgsForRepo(repoName)
		expectedURL := formatURL(replaced.Path, replaced.Version)
		fmt.Printf("        importpath = \"%s\",\n", mod.Path)
		if err := dumpPatchArgsForRepo(repoName); err != nil {
			__antithesis_instrumentation__.Notify(41497)
			return err
		} else {
			__antithesis_instrumentation__.Notify(41498)
		}
		__antithesis_instrumentation__.Notify(41493)
		oldMirror, ok := existingMirrors[repoName]
		if ok && func() bool {
			__antithesis_instrumentation__.Notify(41499)
			return oldMirror.URL == expectedURL == true
		}() == true {
			__antithesis_instrumentation__.Notify(41500)

			fmt.Printf(`        sha256 = "%s",
        strip_prefix = "%s@%s",
        urls = [
            "%s",
        ],
`, oldMirror.Sha256, replaced.Path, replaced.Version, oldMirror.URL)
		} else {
			__antithesis_instrumentation__.Notify(41501)
			if canMirror() {
				__antithesis_instrumentation__.Notify(41502)

				d := downloaded[replaced.Path]
				sha, err := getSha256OfFile(d.Zip)
				if err != nil {
					__antithesis_instrumentation__.Notify(41504)
					return fmt.Errorf("could not get zip for %v: %w", *replaced, err)
				} else {
					__antithesis_instrumentation__.Notify(41505)
				}
				__antithesis_instrumentation__.Notify(41503)
				fmt.Printf(`        sha256 = "%s",
        strip_prefix = "%s@%s",
        urls = [
            "%s",
        ],
`, sha, replaced.Path, replaced.Version, expectedURL)
				g.Go(func() error {
					__antithesis_instrumentation__.Notify(41506)
					return uploadFile(ctx, client, d.Zip, formatSubURL(replaced.Path, replaced.Version))
				})
			} else {
				__antithesis_instrumentation__.Notify(41507)

				d := downloaded[replaced.Path]
				if mod.Replace != nil {
					__antithesis_instrumentation__.Notify(41509)
					fmt.Printf("        replace = \"%s\",\n", replaced.Path)
				} else {
					__antithesis_instrumentation__.Notify(41510)
				}
				__antithesis_instrumentation__.Notify(41508)

				fmt.Printf(`        # TODO: mirror this repo (to fix, run `+"`./dev generate bazel --mirror`)"+`
        sum = "%s",
        version = "%s",
`, d.Sum, d.Version)
			}
		}
		__antithesis_instrumentation__.Notify(41494)
		fmt.Println("    )")
	}
	__antithesis_instrumentation__.Notify(41483)

	if err := g.Wait(); err != nil {
		__antithesis_instrumentation__.Notify(41511)
		return err
	} else {
		__antithesis_instrumentation__.Notify(41512)
	}
	__antithesis_instrumentation__.Notify(41484)
	if client == nil {
		__antithesis_instrumentation__.Notify(41513)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(41514)
	}
	__antithesis_instrumentation__.Notify(41485)
	return client.Close()
}

func mirror() error {
	__antithesis_instrumentation__.Notify(41515)
	tmpdir, err := createTmpDir()
	if err != nil {
		__antithesis_instrumentation__.Notify(41521)
		return err
	} else {
		__antithesis_instrumentation__.Notify(41522)
	}
	__antithesis_instrumentation__.Notify(41516)
	defer func() {
		__antithesis_instrumentation__.Notify(41523)
		err := os.RemoveAll(tmpdir)
		if err != nil {
			__antithesis_instrumentation__.Notify(41524)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(41525)
		}
	}()
	__antithesis_instrumentation__.Notify(41517)
	downloaded, err := downloadZips(tmpdir)
	if err != nil {
		__antithesis_instrumentation__.Notify(41526)
		return err
	} else {
		__antithesis_instrumentation__.Notify(41527)
	}
	__antithesis_instrumentation__.Notify(41518)
	listed, err := listAllModules(tmpdir)
	if err != nil {
		__antithesis_instrumentation__.Notify(41528)
		return err
	} else {
		__antithesis_instrumentation__.Notify(41529)
	}
	__antithesis_instrumentation__.Notify(41519)
	existingMirrors, err := getExistingMirrors()
	if err != nil {
		__antithesis_instrumentation__.Notify(41530)
		return err
	} else {
		__antithesis_instrumentation__.Notify(41531)
	}
	__antithesis_instrumentation__.Notify(41520)

	return dumpNewDepsBzl(listed, downloaded, existingMirrors)
}

func main() {
	__antithesis_instrumentation__.Notify(41532)
	if err := mirror(); err != nil {
		__antithesis_instrumentation__.Notify(41533)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(41534)
	}
}
