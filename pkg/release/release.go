// Package release contains utilities for assisting with the release process.
// This is intended for use for the release commands.
package release

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"mime"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3"
	bazelutil "github.com/cockroachdb/cockroach/pkg/build/util"
	"github.com/cockroachdb/errors"
)

var (
	linuxStaticLibsRe = func() *regexp.Regexp {
		__antithesis_instrumentation__.Notify(128791)
		libs := strings.Join([]string{
			regexp.QuoteMeta("linux-vdso.so."),
			regexp.QuoteMeta("librt.so."),
			regexp.QuoteMeta("libpthread.so."),
			regexp.QuoteMeta("libdl.so."),
			regexp.QuoteMeta("libm.so."),
			regexp.QuoteMeta("libc.so."),
			regexp.QuoteMeta("libresolv.so."),
			strings.Replace(regexp.QuoteMeta("ld-linux-ARCH.so."), "ARCH", ".*", -1),
		}, "|")
		return regexp.MustCompile(libs)
	}()
	osVersionRe = regexp.MustCompile(`\d+(\.\d+)*-`)
)

var (
	NoCache = "no-cache"
)

type Platform int

const (
	PlatformLinux Platform = iota

	PlatformLinuxArm

	PlatformMacOS

	PlatformWindows
)

const ChecksumSuffix = ".sha256sum"

type ExecFn struct {
	MockExecFn func(*exec.Cmd) ([]byte, error)
}

func (e ExecFn) Run(cmd *exec.Cmd) ([]byte, error) {
	__antithesis_instrumentation__.Notify(128792)
	if cmd.Stdout != nil {
		__antithesis_instrumentation__.Notify(128795)
		return nil, errors.New("exec: Stdout already set")
	} else {
		__antithesis_instrumentation__.Notify(128796)
	}
	__antithesis_instrumentation__.Notify(128793)
	var stdout bytes.Buffer
	cmd.Stdout = io.MultiWriter(&stdout, os.Stdout)
	if e.MockExecFn == nil {
		__antithesis_instrumentation__.Notify(128797)
		err := cmd.Run()
		return stdout.Bytes(), err
	} else {
		__antithesis_instrumentation__.Notify(128798)
	}
	__antithesis_instrumentation__.Notify(128794)
	return e.MockExecFn(cmd)
}

type BuildOptions struct {
	Release bool

	BuildTag string

	ExecFn ExecFn
}

func SuffixFromPlatform(platform Platform) string {
	__antithesis_instrumentation__.Notify(128799)
	switch platform {
	case PlatformLinux:
		__antithesis_instrumentation__.Notify(128800)
		return ".linux-2.6.32-gnu-amd64"
	case PlatformLinuxArm:
		__antithesis_instrumentation__.Notify(128801)
		return ".linux-3.7.10-gnu-aarch64"
	case PlatformMacOS:
		__antithesis_instrumentation__.Notify(128802)

		return ".darwin-10.9-amd64"
	case PlatformWindows:
		__antithesis_instrumentation__.Notify(128803)
		return ".windows-6.2-amd64.exe"
	default:
		__antithesis_instrumentation__.Notify(128804)
		panic(errors.Newf("unknown platform %d", platform))
	}
}

func CrossConfigFromPlatform(platform Platform) string {
	__antithesis_instrumentation__.Notify(128805)
	switch platform {
	case PlatformLinux:
		__antithesis_instrumentation__.Notify(128806)
		return "crosslinuxbase"
	case PlatformLinuxArm:
		__antithesis_instrumentation__.Notify(128807)
		return "crosslinuxarmbase"
	case PlatformMacOS:
		__antithesis_instrumentation__.Notify(128808)
		return "crossmacosbase"
	case PlatformWindows:
		__antithesis_instrumentation__.Notify(128809)
		return "crosswindowsbase"
	default:
		__antithesis_instrumentation__.Notify(128810)
		panic(errors.Newf("unknown platform %d", platform))
	}
}

func TargetTripleFromPlatform(platform Platform) string {
	__antithesis_instrumentation__.Notify(128811)
	switch platform {
	case PlatformLinux:
		__antithesis_instrumentation__.Notify(128812)
		return "x86_64-pc-linux-gnu"
	case PlatformLinuxArm:
		__antithesis_instrumentation__.Notify(128813)
		return "aarch64-unknown-linux-gnu"
	case PlatformMacOS:
		__antithesis_instrumentation__.Notify(128814)
		return "x86_64-apple-darwin19"
	case PlatformWindows:
		__antithesis_instrumentation__.Notify(128815)
		return "x86_64-w64-mingw32"
	default:
		__antithesis_instrumentation__.Notify(128816)
		panic(errors.Newf("unknown platform %d", platform))
	}
}

func SharedLibraryExtensionFromPlatform(platform Platform) string {
	__antithesis_instrumentation__.Notify(128817)
	switch platform {
	case PlatformLinux, PlatformLinuxArm:
		__antithesis_instrumentation__.Notify(128818)
		return ".so"
	case PlatformWindows:
		__antithesis_instrumentation__.Notify(128819)
		return ".dll"
	case PlatformMacOS:
		__antithesis_instrumentation__.Notify(128820)
		return ".dylib"
	default:
		__antithesis_instrumentation__.Notify(128821)
		panic(errors.Newf("unknown platform %d", platform))
	}
}

func MakeWorkload(pkgDir string) error {
	__antithesis_instrumentation__.Notify(128822)

	cmd := exec.Command("bazel", "build", "//pkg/cmd/workload", "--config=crosslinux", "--config=ci")
	cmd.Dir = pkgDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Printf("%s", cmd.Args)
	err := cmd.Run()
	if err != nil {
		__antithesis_instrumentation__.Notify(128825)
		return err
	} else {
		__antithesis_instrumentation__.Notify(128826)
	}
	__antithesis_instrumentation__.Notify(128823)
	bazelBin, err := getPathToBazelBin(ExecFn{}, pkgDir, []string{"--config=crosslinux", "--config=ci"})
	if err != nil {
		__antithesis_instrumentation__.Notify(128827)
		return err
	} else {
		__antithesis_instrumentation__.Notify(128828)
	}
	__antithesis_instrumentation__.Notify(128824)
	return stageBinary("//pkg/cmd/workload", PlatformLinux, bazelBin, filepath.Join(pkgDir, "bin"))
}

func MakeRelease(platform Platform, opts BuildOptions, pkgDir string) error {
	__antithesis_instrumentation__.Notify(128829)
	buildArgs := []string{"build", "//pkg/cmd/cockroach", "//c-deps:libgeos", "//pkg/cmd/cockroach-sql"}
	targetTriple := TargetTripleFromPlatform(platform)
	if opts.Release {
		__antithesis_instrumentation__.Notify(128837)
		if opts.BuildTag == "" {
			__antithesis_instrumentation__.Notify(128839)
			return errors.Newf("must set BuildTag if Release is set")
		} else {
			__antithesis_instrumentation__.Notify(128840)
		}
		__antithesis_instrumentation__.Notify(128838)
		buildArgs = append(buildArgs, fmt.Sprintf("--workspace_status_command=./build/bazelutil/stamp.sh %s official-binary %s release", targetTriple, opts.BuildTag))
	} else {
		__antithesis_instrumentation__.Notify(128841)
		if opts.BuildTag != "" {
			__antithesis_instrumentation__.Notify(128843)
			return errors.Newf("cannot set BuildTag if Release is not set")
		} else {
			__antithesis_instrumentation__.Notify(128844)
		}
		__antithesis_instrumentation__.Notify(128842)
		buildArgs = append(buildArgs, fmt.Sprintf("--workspace_status_command=./build/bazelutil/stamp.sh %s official-binary", targetTriple))
	}
	__antithesis_instrumentation__.Notify(128830)
	configs := []string{"-c", "opt", "--config=ci", "--config=with_ui", fmt.Sprintf("--config=%s", CrossConfigFromPlatform(platform))}
	buildArgs = append(buildArgs, configs...)
	cmd := exec.Command("bazel", buildArgs...)
	cmd.Dir = pkgDir
	cmd.Stderr = os.Stderr
	log.Printf("%s", cmd.Args)
	stdoutBytes, err := opts.ExecFn.Run(cmd)
	if err != nil {
		__antithesis_instrumentation__.Notify(128845)
		return errors.Wrapf(err, "failed to run %s: %s", cmd.Args, string(stdoutBytes))
	} else {
		__antithesis_instrumentation__.Notify(128846)
	}
	__antithesis_instrumentation__.Notify(128831)

	bazelBin, err := getPathToBazelBin(opts.ExecFn, pkgDir, configs)
	if err != nil {
		__antithesis_instrumentation__.Notify(128847)
		return err
	} else {
		__antithesis_instrumentation__.Notify(128848)
	}
	__antithesis_instrumentation__.Notify(128832)
	if err := stageBinary("//pkg/cmd/cockroach", platform, bazelBin, pkgDir); err != nil {
		__antithesis_instrumentation__.Notify(128849)
		return err
	} else {
		__antithesis_instrumentation__.Notify(128850)
	}
	__antithesis_instrumentation__.Notify(128833)

	if err := stageBinary("//pkg/cmd/cockroach-sql", platform, bazelBin, pkgDir); err != nil {
		__antithesis_instrumentation__.Notify(128851)
		return err
	} else {
		__antithesis_instrumentation__.Notify(128852)
	}
	__antithesis_instrumentation__.Notify(128834)
	if err := stageLibraries(platform, bazelBin, filepath.Join(pkgDir, "lib")); err != nil {
		__antithesis_instrumentation__.Notify(128853)
		return err
	} else {
		__antithesis_instrumentation__.Notify(128854)
	}
	__antithesis_instrumentation__.Notify(128835)

	if platform == PlatformLinux {
		__antithesis_instrumentation__.Notify(128855)
		suffix := SuffixFromPlatform(platform)
		binaryName := "./cockroach" + suffix

		cmd := exec.Command(binaryName, "version")
		cmd.Dir = pkgDir
		cmd.Env = append(cmd.Env, "MALLOC_CONF=prof:true")
		cmd.Stderr = os.Stderr
		log.Printf("%s %s", cmd.Env, cmd.Args)
		stdoutBytes, err := opts.ExecFn.Run(cmd)
		if err != nil {
			__antithesis_instrumentation__.Notify(128859)
			return errors.Wrapf(err, "%s %s: %s", cmd.Env, cmd.Args, string(stdoutBytes))
		} else {
			__antithesis_instrumentation__.Notify(128860)
		}
		__antithesis_instrumentation__.Notify(128856)

		cmd = exec.Command("ldd", binaryName)
		cmd.Dir = pkgDir
		cmd.Stderr = os.Stderr
		log.Printf("%s %s", cmd.Env, cmd.Args)
		stdoutBytes, err = opts.ExecFn.Run(cmd)
		if err != nil {
			__antithesis_instrumentation__.Notify(128861)
			log.Fatalf("%s %s: out=%s err=%v", cmd.Env, cmd.Args, string(stdoutBytes), err)
		} else {
			__antithesis_instrumentation__.Notify(128862)
		}
		__antithesis_instrumentation__.Notify(128857)
		scanner := bufio.NewScanner(bytes.NewReader(stdoutBytes))
		for scanner.Scan() {
			__antithesis_instrumentation__.Notify(128863)
			if line := scanner.Text(); !linuxStaticLibsRe.MatchString(line) {
				__antithesis_instrumentation__.Notify(128864)
				return errors.Newf("%s is not properly statically linked:\n%s", binaryName, line)
			} else {
				__antithesis_instrumentation__.Notify(128865)
			}
		}
		__antithesis_instrumentation__.Notify(128858)
		if err := scanner.Err(); err != nil {
			__antithesis_instrumentation__.Notify(128866)
			return err
		} else {
			__antithesis_instrumentation__.Notify(128867)
		}
	} else {
		__antithesis_instrumentation__.Notify(128868)
	}
	__antithesis_instrumentation__.Notify(128836)
	return nil
}

func TrimDotExe(name string) (string, bool) {
	__antithesis_instrumentation__.Notify(128869)
	const dotExe = ".exe"
	return strings.TrimSuffix(name, dotExe), strings.HasSuffix(name, dotExe)
}

type S3Putter interface {
	PutObject(*s3.PutObjectInput) (*s3.PutObjectOutput, error)
}

func S3KeyRelease(platform Platform, versionStr string, binaryPrefix string) (string, string) {
	__antithesis_instrumentation__.Notify(128870)
	suffix := SuffixFromPlatform(platform)
	targetSuffix, hasExe := TrimDotExe(suffix)

	if platform == PlatformLinux {
		__antithesis_instrumentation__.Notify(128873)
		targetSuffix = strings.Replace(targetSuffix, "gnu-", "", -1)
		targetSuffix = osVersionRe.ReplaceAllLiteralString(targetSuffix, "")
	} else {
		__antithesis_instrumentation__.Notify(128874)
	}
	__antithesis_instrumentation__.Notify(128871)

	archiveBase := fmt.Sprintf("%s-%s", binaryPrefix, versionStr)
	targetArchiveBase := archiveBase + targetSuffix
	if hasExe {
		__antithesis_instrumentation__.Notify(128875)
		return targetArchiveBase, targetArchiveBase + ".zip"
	} else {
		__antithesis_instrumentation__.Notify(128876)
	}
	__antithesis_instrumentation__.Notify(128872)
	return targetArchiveBase, targetArchiveBase + ".tgz"
}

type NonReleaseFile struct {
	S3FileName string

	S3FilePath string

	S3RedirectPathPrefix string

	LocalAbsolutePath string
}

func MakeCRDBBinaryNonReleaseFile(localAbsolutePath string, versionStr string) NonReleaseFile {
	__antithesis_instrumentation__.Notify(128877)
	base := filepath.Base(localAbsolutePath)
	remoteName, hasExe := TrimDotExe(base)

	remoteName = osVersionRe.ReplaceAllLiteralString(remoteName, "")

	fileName := fmt.Sprintf("%s.%s", remoteName, versionStr)
	if hasExe {
		__antithesis_instrumentation__.Notify(128879)
		fileName += ".exe"
	} else {
		__antithesis_instrumentation__.Notify(128880)
	}
	__antithesis_instrumentation__.Notify(128878)

	return NonReleaseFile{
		S3FileName:           fileName,
		S3FilePath:           fileName,
		S3RedirectPathPrefix: remoteName,
		LocalAbsolutePath:    localAbsolutePath,
	}
}

var CRDBSharedLibraries = []string{"libgeos", "libgeos_c"}

func MakeCRDBLibraryNonReleaseFiles(
	localAbsoluteBasePath string, platform Platform, versionStr string,
) []NonReleaseFile {
	__antithesis_instrumentation__.Notify(128881)
	files := []NonReleaseFile{}
	suffix := SuffixFromPlatform(platform)
	ext := SharedLibraryExtensionFromPlatform(platform)
	for _, localFileName := range CRDBSharedLibraries {
		__antithesis_instrumentation__.Notify(128883)
		remoteFileNameBase, _ := TrimDotExe(osVersionRe.ReplaceAllLiteralString(fmt.Sprintf("%s%s", localFileName, suffix), ""))
		remoteFileName := fmt.Sprintf("%s.%s", remoteFileNameBase, versionStr)

		files = append(
			files,
			NonReleaseFile{
				S3FileName:           fmt.Sprintf("%s%s", remoteFileName, ext),
				S3FilePath:           fmt.Sprintf("lib/%s%s", remoteFileName, ext),
				S3RedirectPathPrefix: fmt.Sprintf("lib/%s%s", remoteFileNameBase, ext),
				LocalAbsolutePath:    filepath.Join(localAbsoluteBasePath, "lib", localFileName+ext),
			},
		)
	}
	__antithesis_instrumentation__.Notify(128882)
	return files
}

type PutNonReleaseOptions struct {
	Branch string

	BucketName string

	Files []NonReleaseFile
}

func PutNonRelease(svc S3Putter, o PutNonReleaseOptions) {
	__antithesis_instrumentation__.Notify(128884)
	const repoName = "cockroach"
	for _, f := range o.Files {
		__antithesis_instrumentation__.Notify(128885)
		disposition := mime.FormatMediaType("attachment", map[string]string{
			"filename": f.S3FileName,
		})

		fileToUpload, err := os.Open(f.LocalAbsolutePath)
		if err != nil {
			__antithesis_instrumentation__.Notify(128890)
			log.Fatalf("failed to open %s: %s", f.LocalAbsolutePath, err)
		} else {
			__antithesis_instrumentation__.Notify(128891)
		}
		__antithesis_instrumentation__.Notify(128886)
		defer func() {
			__antithesis_instrumentation__.Notify(128892)
			_ = fileToUpload.Close()
		}()
		__antithesis_instrumentation__.Notify(128887)

		versionKey := fmt.Sprintf("/%s/%s", repoName, f.S3FilePath)
		log.Printf("Uploading to s3://%s%s", o.BucketName, versionKey)
		if _, err := svc.PutObject(&s3.PutObjectInput{
			Bucket:             &o.BucketName,
			ContentDisposition: &disposition,
			Key:                &versionKey,
			Body:               fileToUpload,
		}); err != nil {
			__antithesis_instrumentation__.Notify(128893)
			log.Fatalf("s3 upload %s: %s", versionKey, err)
		} else {
			__antithesis_instrumentation__.Notify(128894)
		}
		__antithesis_instrumentation__.Notify(128888)

		latestSuffix := o.Branch
		if latestSuffix == "master" {
			__antithesis_instrumentation__.Notify(128895)
			latestSuffix = "LATEST"
		} else {
			__antithesis_instrumentation__.Notify(128896)
		}
		__antithesis_instrumentation__.Notify(128889)
		latestKey := fmt.Sprintf("%s/%s.%s", repoName, f.S3RedirectPathPrefix, latestSuffix)
		if _, err := svc.PutObject(&s3.PutObjectInput{
			Bucket:                  &o.BucketName,
			CacheControl:            &NoCache,
			Key:                     &latestKey,
			WebsiteRedirectLocation: &versionKey,
		}); err != nil {
			__antithesis_instrumentation__.Notify(128897)
			log.Fatalf("s3 redirect to %s: %s", versionKey, err)
		} else {
			__antithesis_instrumentation__.Notify(128898)
		}
	}
}

type ArchiveFile struct {
	LocalAbsolutePath string

	ArchiveFilePath string
}

func MakeCRDBBinaryArchiveFile(localAbsolutePath string, path string) ArchiveFile {
	__antithesis_instrumentation__.Notify(128899)
	base := filepath.Base(localAbsolutePath)
	_, hasExe := TrimDotExe(base)
	if hasExe {
		__antithesis_instrumentation__.Notify(128901)
		path += ".exe"
	} else {
		__antithesis_instrumentation__.Notify(128902)
	}
	__antithesis_instrumentation__.Notify(128900)
	return ArchiveFile{
		LocalAbsolutePath: localAbsolutePath,
		ArchiveFilePath:   path,
	}
}

func MakeCRDBLibraryArchiveFiles(pkgDir string, platform Platform) []ArchiveFile {
	__antithesis_instrumentation__.Notify(128903)
	files := []ArchiveFile{}
	ext := SharedLibraryExtensionFromPlatform(platform)
	for _, lib := range CRDBSharedLibraries {
		__antithesis_instrumentation__.Notify(128905)
		localFileName := lib + ext
		files = append(
			files,
			ArchiveFile{
				LocalAbsolutePath: filepath.Join(pkgDir, "lib", localFileName),
				ArchiveFilePath:   "lib/" + localFileName,
			},
		)
	}
	__antithesis_instrumentation__.Notify(128904)
	return files
}

type PutReleaseOptions struct {
	BucketName string

	NoCache bool

	Platform Platform

	VersionStr string

	Files      []ArchiveFile
	ExtraFiles []ArchiveFile
}

func PutRelease(svc S3Putter, o PutReleaseOptions) {
	__antithesis_instrumentation__.Notify(128906)
	targetArchiveBase, targetArchive := S3KeyRelease(o.Platform, o.VersionStr, "cockroach")
	var body bytes.Buffer

	if strings.HasSuffix(targetArchive, ".zip") {
		__antithesis_instrumentation__.Notify(128912)
		zw := zip.NewWriter(&body)

		for _, f := range o.Files {
			__antithesis_instrumentation__.Notify(128914)
			file, err := os.Open(f.LocalAbsolutePath)
			if err != nil {
				__antithesis_instrumentation__.Notify(128920)
				log.Fatalf("failed to open file: %s", f.LocalAbsolutePath)
			} else {
				__antithesis_instrumentation__.Notify(128921)
			}
			__antithesis_instrumentation__.Notify(128915)
			defer func() { __antithesis_instrumentation__.Notify(128922); _ = file.Close() }()
			__antithesis_instrumentation__.Notify(128916)

			stat, err := file.Stat()
			if err != nil {
				__antithesis_instrumentation__.Notify(128923)
				log.Fatalf("failed to stat: %s", f.LocalAbsolutePath)
			} else {
				__antithesis_instrumentation__.Notify(128924)
			}
			__antithesis_instrumentation__.Notify(128917)

			zipHeader, err := zip.FileInfoHeader(stat)
			if err != nil {
				__antithesis_instrumentation__.Notify(128925)
				log.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(128926)
			}
			__antithesis_instrumentation__.Notify(128918)
			zipHeader.Name = filepath.Join(targetArchiveBase, f.ArchiveFilePath)
			zipHeader.Method = zip.Deflate

			zfw, err := zw.CreateHeader(zipHeader)
			if err != nil {
				__antithesis_instrumentation__.Notify(128927)
				log.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(128928)
			}
			__antithesis_instrumentation__.Notify(128919)
			if _, err := io.Copy(zfw, file); err != nil {
				__antithesis_instrumentation__.Notify(128929)
				log.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(128930)
			}
		}
		__antithesis_instrumentation__.Notify(128913)
		if err := zw.Close(); err != nil {
			__antithesis_instrumentation__.Notify(128931)
			log.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(128932)
		}
	} else {
		__antithesis_instrumentation__.Notify(128933)
		gzw := gzip.NewWriter(&body)
		tw := tar.NewWriter(gzw)
		for _, f := range o.Files {
			__antithesis_instrumentation__.Notify(128936)

			file, err := os.Open(f.LocalAbsolutePath)
			if err != nil {
				__antithesis_instrumentation__.Notify(128942)
				log.Fatalf("failed to open file: %s", f.LocalAbsolutePath)
			} else {
				__antithesis_instrumentation__.Notify(128943)
			}
			__antithesis_instrumentation__.Notify(128937)
			defer func() { __antithesis_instrumentation__.Notify(128944); _ = file.Close() }()
			__antithesis_instrumentation__.Notify(128938)

			stat, err := file.Stat()
			if err != nil {
				__antithesis_instrumentation__.Notify(128945)
				log.Fatalf("failed to stat: %s", f.LocalAbsolutePath)
			} else {
				__antithesis_instrumentation__.Notify(128946)
			}
			__antithesis_instrumentation__.Notify(128939)

			tarHeader, err := tar.FileInfoHeader(stat, "")
			if err != nil {
				__antithesis_instrumentation__.Notify(128947)
				log.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(128948)
			}
			__antithesis_instrumentation__.Notify(128940)
			tarHeader.Name = filepath.Join(targetArchiveBase, f.ArchiveFilePath)
			if err := tw.WriteHeader(tarHeader); err != nil {
				__antithesis_instrumentation__.Notify(128949)
				log.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(128950)
			}
			__antithesis_instrumentation__.Notify(128941)

			if _, err := io.Copy(tw, file); err != nil {
				__antithesis_instrumentation__.Notify(128951)
				log.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(128952)
			}
		}
		__antithesis_instrumentation__.Notify(128934)
		if err := tw.Close(); err != nil {
			__antithesis_instrumentation__.Notify(128953)
			log.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(128954)
		}
		__antithesis_instrumentation__.Notify(128935)
		if err := gzw.Close(); err != nil {
			__antithesis_instrumentation__.Notify(128955)
			log.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(128956)
		}
	}
	__antithesis_instrumentation__.Notify(128907)

	log.Printf("Uploading to s3://%s/%s", o.BucketName, targetArchive)
	putObjectInput := s3.PutObjectInput{
		Bucket: &o.BucketName,
		Key:    &targetArchive,
		Body:   bytes.NewReader(body.Bytes()),
	}
	if o.NoCache {
		__antithesis_instrumentation__.Notify(128957)
		putObjectInput.CacheControl = &NoCache
	} else {
		__antithesis_instrumentation__.Notify(128958)
	}
	__antithesis_instrumentation__.Notify(128908)
	if _, err := svc.PutObject(&putObjectInput); err != nil {
		__antithesis_instrumentation__.Notify(128959)
		log.Fatalf("s3 upload %s: %s", targetArchive, err)
	} else {
		__antithesis_instrumentation__.Notify(128960)
	}
	__antithesis_instrumentation__.Notify(128909)

	checksumContents := fmt.Sprintf("%x %s\n", sha256.Sum256(body.Bytes()),
		filepath.Base(targetArchive))
	targetChecksum := targetArchive + ChecksumSuffix
	log.Printf("Uploading to s3://%s/%s", o.BucketName, targetChecksum)
	putObjectInputChecksum := s3.PutObjectInput{
		Bucket: &o.BucketName,
		Key:    &targetChecksum,
		Body:   strings.NewReader(checksumContents),
	}
	if o.NoCache {
		__antithesis_instrumentation__.Notify(128961)
		putObjectInputChecksum.CacheControl = &NoCache
	} else {
		__antithesis_instrumentation__.Notify(128962)
	}
	__antithesis_instrumentation__.Notify(128910)
	if _, err := svc.PutObject(&putObjectInputChecksum); err != nil {
		__antithesis_instrumentation__.Notify(128963)
		log.Fatalf("s3 upload %s: %s", targetChecksum, err)
	} else {
		__antithesis_instrumentation__.Notify(128964)
	}
	__antithesis_instrumentation__.Notify(128911)
	for _, f := range o.ExtraFiles {
		__antithesis_instrumentation__.Notify(128965)
		keyBase, hasExe := TrimDotExe(f.ArchiveFilePath)
		targetKey, _ := S3KeyRelease(o.Platform, o.VersionStr, keyBase)
		if hasExe {
			__antithesis_instrumentation__.Notify(128969)
			targetKey += ".exe"
		} else {
			__antithesis_instrumentation__.Notify(128970)
		}
		__antithesis_instrumentation__.Notify(128966)
		log.Printf("Uploading to s3://%s/%s", o.BucketName, targetKey)
		handle, err := os.Open(f.LocalAbsolutePath)
		if err != nil {
			__antithesis_instrumentation__.Notify(128971)
			log.Fatalf("failed to open %s: %s", f.LocalAbsolutePath, err)
		} else {
			__antithesis_instrumentation__.Notify(128972)
		}
		__antithesis_instrumentation__.Notify(128967)
		putObjectInput := s3.PutObjectInput{
			Bucket: &o.BucketName,
			Key:    &targetKey,
			Body:   handle,
		}
		if o.NoCache {
			__antithesis_instrumentation__.Notify(128973)
			putObjectInput.CacheControl = &NoCache
		} else {
			__antithesis_instrumentation__.Notify(128974)
		}
		__antithesis_instrumentation__.Notify(128968)
		if _, err := svc.PutObject(&putObjectInput); err != nil {
			__antithesis_instrumentation__.Notify(128975)
			log.Fatalf("s3 upload %s: %s", targetKey, err)
		} else {
			__antithesis_instrumentation__.Notify(128976)
		}
	}
}

func getPathToBazelBin(execFn ExecFn, pkgDir string, configArgs []string) (string, error) {
	__antithesis_instrumentation__.Notify(128977)
	args := []string{"info", "bazel-bin"}
	args = append(args, configArgs...)
	cmd := exec.Command("bazel", args...)
	cmd.Dir = pkgDir
	cmd.Stderr = os.Stderr
	stdoutBytes, err := execFn.Run(cmd)
	if err != nil {
		__antithesis_instrumentation__.Notify(128979)
		return "", errors.Wrapf(err, "failed to run %s: %s", cmd.Args, string(stdoutBytes))
	} else {
		__antithesis_instrumentation__.Notify(128980)
	}
	__antithesis_instrumentation__.Notify(128978)
	return strings.TrimSpace(string(stdoutBytes)), nil
}

func stageBinary(target string, platform Platform, bazelBin string, dir string) error {
	__antithesis_instrumentation__.Notify(128981)
	if err := os.MkdirAll(dir, 0755); err != nil {
		__antithesis_instrumentation__.Notify(128985)
		return err
	} else {
		__antithesis_instrumentation__.Notify(128986)
	}
	__antithesis_instrumentation__.Notify(128982)
	rel := bazelutil.OutputOfBinaryRule(target, platform == PlatformWindows)
	src := filepath.Join(bazelBin, rel)
	dstBase, _ := TrimDotExe(filepath.Base(rel))
	suffix := SuffixFromPlatform(platform)
	dstBase = dstBase + suffix
	dst := filepath.Join(dir, dstBase)
	srcF, err := os.Open(src)
	if err != nil {
		__antithesis_instrumentation__.Notify(128987)
		return err
	} else {
		__antithesis_instrumentation__.Notify(128988)
	}
	__antithesis_instrumentation__.Notify(128983)
	defer closeFileOrPanic(srcF)
	dstF, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		__antithesis_instrumentation__.Notify(128989)
		return err
	} else {
		__antithesis_instrumentation__.Notify(128990)
	}
	__antithesis_instrumentation__.Notify(128984)
	defer closeFileOrPanic(dstF)
	_, err = io.Copy(dstF, srcF)
	return err
}

func stageLibraries(platform Platform, bazelBin string, dir string) error {
	__antithesis_instrumentation__.Notify(128991)
	if err := os.MkdirAll(dir, 0755); err != nil {
		__antithesis_instrumentation__.Notify(128994)
		return err
	} else {
		__antithesis_instrumentation__.Notify(128995)
	}
	__antithesis_instrumentation__.Notify(128992)
	ext := SharedLibraryExtensionFromPlatform(platform)
	for _, lib := range CRDBSharedLibraries {
		__antithesis_instrumentation__.Notify(128996)
		libDir := "lib"
		if platform == PlatformWindows {
			__antithesis_instrumentation__.Notify(129000)

			libDir = "bin"
		} else {
			__antithesis_instrumentation__.Notify(129001)
		}
		__antithesis_instrumentation__.Notify(128997)
		src := filepath.Join(bazelBin, "c-deps", "libgeos", libDir, lib+ext)
		srcF, err := os.Open(src)
		if err != nil {
			__antithesis_instrumentation__.Notify(129002)
			return err
		} else {
			__antithesis_instrumentation__.Notify(129003)
		}
		__antithesis_instrumentation__.Notify(128998)
		defer closeFileOrPanic(srcF)
		dst := filepath.Join(dir, filepath.Base(src))
		dstF, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			__antithesis_instrumentation__.Notify(129004)
			return err
		} else {
			__antithesis_instrumentation__.Notify(129005)
		}
		__antithesis_instrumentation__.Notify(128999)
		defer closeFileOrPanic(dstF)
		_, err = io.Copy(dstF, srcF)
		if err != nil {
			__antithesis_instrumentation__.Notify(129006)
			return err
		} else {
			__antithesis_instrumentation__.Notify(129007)
		}
	}
	__antithesis_instrumentation__.Notify(128993)
	return nil
}

func closeFileOrPanic(f io.Closer) {
	__antithesis_instrumentation__.Notify(129008)
	err := f.Close()
	if err != nil {
		__antithesis_instrumentation__.Notify(129009)
		panic(errors.Wrapf(err, "could not close file"))
	} else {
		__antithesis_instrumentation__.Notify(129010)
	}
}
