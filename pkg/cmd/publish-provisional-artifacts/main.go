package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"flag"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/cockroachdb/cockroach/pkg/release"
	"github.com/cockroachdb/cockroach/pkg/util/version"
	"github.com/kr/pretty"
)

const (
	awsAccessKeyIDKey      = "AWS_ACCESS_KEY_ID"
	awsSecretAccessKeyKey  = "AWS_SECRET_ACCESS_KEY"
	teamcityBuildBranchKey = "TC_BUILD_BRANCH"
)

var provisionalReleasePrefixRE = regexp.MustCompile(`^provisional_[0-9]{12}_`)

type s3I interface {
	GetObject(*s3.GetObjectInput) (*s3.GetObjectOutput, error)
	PutObject(*s3.PutObjectInput) (*s3.PutObjectOutput, error)
}

func makeS3() (s3I, error) {
	__antithesis_instrumentation__.Notify(41662)
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(41664)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(41665)
	}
	__antithesis_instrumentation__.Notify(41663)
	return s3.New(sess), nil
}

var isReleaseF = flag.Bool("release", false, "build in release mode instead of bleeding-edge mode")
var destBucket = flag.String("bucket", "", "override default bucket")
var doProvisionalF = flag.Bool("provisional", false, "publish provisional binaries")
var doBlessF = flag.Bool("bless", false, "bless provisional binaries")

var (
	latestStr = "latest"
)

func main() {
	__antithesis_instrumentation__.Notify(41666)
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if _, ok := os.LookupEnv(awsAccessKeyIDKey); !ok {
		__antithesis_instrumentation__.Notify(41674)
		log.Fatalf("AWS access key ID environment variable %s is not set", awsAccessKeyIDKey)
	} else {
		__antithesis_instrumentation__.Notify(41675)
	}
	__antithesis_instrumentation__.Notify(41667)
	if _, ok := os.LookupEnv(awsSecretAccessKeyKey); !ok {
		__antithesis_instrumentation__.Notify(41676)
		log.Fatalf("AWS secret access key environment variable %s is not set", awsSecretAccessKeyKey)
	} else {
		__antithesis_instrumentation__.Notify(41677)
	}
	__antithesis_instrumentation__.Notify(41668)
	s3, err := makeS3()
	if err != nil {
		__antithesis_instrumentation__.Notify(41678)
		log.Fatalf("Creating AWS S3 session: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(41679)
	}
	__antithesis_instrumentation__.Notify(41669)
	branch, ok := os.LookupEnv(teamcityBuildBranchKey)
	if !ok {
		__antithesis_instrumentation__.Notify(41680)
		log.Fatalf("VCS branch environment variable %s is not set", teamcityBuildBranchKey)
	} else {
		__antithesis_instrumentation__.Notify(41681)
	}
	__antithesis_instrumentation__.Notify(41670)
	pkg, err := os.Getwd()
	if err != nil {
		__antithesis_instrumentation__.Notify(41682)
		log.Fatalf("unable to locate CRDB directory: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(41683)
	}
	__antithesis_instrumentation__.Notify(41671)

	_, err = os.Stat(filepath.Join(pkg, "WORKSPACE"))
	if err != nil {
		__antithesis_instrumentation__.Notify(41684)
		log.Fatalf("unable to locate CRDB directory: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(41685)
	}
	__antithesis_instrumentation__.Notify(41672)

	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = pkg
	log.Printf("%s %s", cmd.Env, cmd.Args)
	shaOut, err := cmd.Output()
	if err != nil {
		__antithesis_instrumentation__.Notify(41686)
		log.Fatalf("%s: out=%q err=%s", cmd.Args, shaOut, err)
	} else {
		__antithesis_instrumentation__.Notify(41687)
	}
	__antithesis_instrumentation__.Notify(41673)

	run(s3, runFlags{
		doProvisional: *doProvisionalF,
		doBless:       *doBlessF,
		isRelease:     *isReleaseF,
		branch:        branch,
		pkgDir:        pkg,
		sha:           string(bytes.TrimSpace(shaOut)),
	}, release.ExecFn{})
}

type runFlags struct {
	doProvisional, doBless bool
	isRelease              bool
	branch, sha            string
	pkgDir                 string
}

func run(svc s3I, flags runFlags, execFn release.ExecFn) {
	__antithesis_instrumentation__.Notify(41688)

	if !flags.isRelease {
		__antithesis_instrumentation__.Notify(41694)
		flags.doProvisional = true
		flags.doBless = false
	} else {
		__antithesis_instrumentation__.Notify(41695)
	}
	__antithesis_instrumentation__.Notify(41689)

	var versionStr string
	var updateLatest bool
	if flags.isRelease {
		__antithesis_instrumentation__.Notify(41696)

		versionStr = provisionalReleasePrefixRE.ReplaceAllLiteralString(flags.branch, "")

		ver, err := version.Parse(versionStr)
		if err != nil {
			__antithesis_instrumentation__.Notify(41698)
			log.Fatalf("refusing to build release with invalid version name '%s' (err: %s)",
				versionStr, err)
		} else {
			__antithesis_instrumentation__.Notify(41699)
		}
		__antithesis_instrumentation__.Notify(41697)

		if ver.PreRelease() == "" {
			__antithesis_instrumentation__.Notify(41700)

			updateLatest = true
		} else {
			__antithesis_instrumentation__.Notify(41701)
		}
	} else {
		__antithesis_instrumentation__.Notify(41702)
		versionStr = flags.sha
		updateLatest = true
	}
	__antithesis_instrumentation__.Notify(41690)

	var bucketName string
	if len(*destBucket) > 0 {
		__antithesis_instrumentation__.Notify(41703)
		bucketName = *destBucket
	} else {
		__antithesis_instrumentation__.Notify(41704)
		if flags.isRelease {
			__antithesis_instrumentation__.Notify(41705)
			bucketName = "binaries.cockroachdb.com"
		} else {
			__antithesis_instrumentation__.Notify(41706)
			bucketName = "cockroach"
		}
	}
	__antithesis_instrumentation__.Notify(41691)
	log.Printf("Using S3 bucket: %s", bucketName)

	var cockroachBuildOpts []opts
	for _, platform := range []release.Platform{release.PlatformLinux, release.PlatformMacOS, release.PlatformWindows} {
		__antithesis_instrumentation__.Notify(41707)
		var o opts
		o.Platform = platform
		o.PkgDir = flags.pkgDir
		o.Branch = flags.branch
		o.VersionStr = versionStr
		o.BucketName = bucketName
		o.AbsolutePath = filepath.Join(flags.pkgDir, "cockroach"+release.SuffixFromPlatform(platform))
		o.CockroachSQLAbsolutePath = filepath.Join(flags.pkgDir, "cockroach-sql"+release.SuffixFromPlatform(platform))
		cockroachBuildOpts = append(cockroachBuildOpts, o)
	}
	__antithesis_instrumentation__.Notify(41692)

	if flags.doProvisional {
		__antithesis_instrumentation__.Notify(41708)
		for _, o := range cockroachBuildOpts {
			__antithesis_instrumentation__.Notify(41709)
			buildCockroach(flags, o, execFn)

			if !flags.isRelease {
				__antithesis_instrumentation__.Notify(41710)
				putNonRelease(svc, o)
			} else {
				__antithesis_instrumentation__.Notify(41711)
				putRelease(svc, o)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(41712)
	}
	__antithesis_instrumentation__.Notify(41693)
	if flags.doBless {
		__antithesis_instrumentation__.Notify(41713)
		if !flags.isRelease {
			__antithesis_instrumentation__.Notify(41715)
			log.Fatal("cannot bless non-release versions")
		} else {
			__antithesis_instrumentation__.Notify(41716)
		}
		__antithesis_instrumentation__.Notify(41714)
		if updateLatest {
			__antithesis_instrumentation__.Notify(41717)
			for _, o := range cockroachBuildOpts {
				__antithesis_instrumentation__.Notify(41718)
				markLatestRelease(svc, o)
			}
		} else {
			__antithesis_instrumentation__.Notify(41719)
		}
	} else {
		__antithesis_instrumentation__.Notify(41720)
	}
}

func buildCockroach(flags runFlags, o opts, execFn release.ExecFn) {
	__antithesis_instrumentation__.Notify(41721)
	log.Printf("building cockroach %s", pretty.Sprint(o))
	defer func() {
		__antithesis_instrumentation__.Notify(41724)
		log.Printf("done building cockroach: %s", pretty.Sprint(o))
	}()
	__antithesis_instrumentation__.Notify(41722)

	var buildOpts release.BuildOptions
	buildOpts.ExecFn = execFn
	if flags.isRelease {
		__antithesis_instrumentation__.Notify(41725)
		buildOpts.Release = true
		buildOpts.BuildTag = o.VersionStr
	} else {
		__antithesis_instrumentation__.Notify(41726)
	}
	__antithesis_instrumentation__.Notify(41723)

	if err := release.MakeRelease(o.Platform, buildOpts, o.PkgDir); err != nil {
		__antithesis_instrumentation__.Notify(41727)
		log.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(41728)
	}
}

type opts struct {
	VersionStr string
	Branch     string

	Platform release.Platform

	AbsolutePath             string
	CockroachSQLAbsolutePath string
	BucketName               string
	PkgDir                   string
}

func putNonRelease(svc s3I, o opts) {
	__antithesis_instrumentation__.Notify(41729)
	release.PutNonRelease(
		svc,
		release.PutNonReleaseOptions{
			Branch:     o.Branch,
			BucketName: o.BucketName,
			Files: append(
				[]release.NonReleaseFile{
					release.MakeCRDBBinaryNonReleaseFile(o.AbsolutePath, o.VersionStr),
					release.MakeCRDBBinaryNonReleaseFile(o.CockroachSQLAbsolutePath, o.VersionStr),
				},
				release.MakeCRDBLibraryNonReleaseFiles(o.PkgDir, o.Platform, o.VersionStr)...,
			),
		},
	)
}

func s3KeyRelease(o opts) (string, string) {
	__antithesis_instrumentation__.Notify(41730)
	return release.S3KeyRelease(o.Platform, o.VersionStr, "cockroach")
}

func putRelease(svc s3I, o opts) {
	__antithesis_instrumentation__.Notify(41731)
	release.PutRelease(svc, release.PutReleaseOptions{
		BucketName: o.BucketName,
		NoCache:    false,
		Platform:   o.Platform,
		VersionStr: o.VersionStr,
		Files: append(
			[]release.ArchiveFile{release.MakeCRDBBinaryArchiveFile(o.AbsolutePath, "cockroach")},
			release.MakeCRDBLibraryArchiveFiles(o.PkgDir, o.Platform)...,
		),
		ExtraFiles: []release.ArchiveFile{release.MakeCRDBBinaryArchiveFile(o.CockroachSQLAbsolutePath, "cockroach-sql")},
	})
}

func markLatestRelease(svc s3I, o opts) {
	__antithesis_instrumentation__.Notify(41732)
	markLatestReleaseWithSuffix(svc, o, "")
	markLatestReleaseWithSuffix(svc, o, release.ChecksumSuffix)
}

func markLatestReleaseWithSuffix(svc s3I, o opts, suffix string) {
	__antithesis_instrumentation__.Notify(41733)
	_, keyRelease := s3KeyRelease(o)
	keyRelease += suffix
	log.Printf("Downloading from %s/%s", o.BucketName, keyRelease)
	binary, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: &o.BucketName,
		Key:    &keyRelease,
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(41736)
		log.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(41737)
	}
	__antithesis_instrumentation__.Notify(41734)
	defer binary.Body.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, binary.Body); err != nil {
		__antithesis_instrumentation__.Notify(41738)
		log.Fatalf("downloading %s/%s: %s", o.BucketName, keyRelease, err)
	} else {
		__antithesis_instrumentation__.Notify(41739)
	}
	__antithesis_instrumentation__.Notify(41735)

	oLatest := o
	oLatest.VersionStr = latestStr
	_, keyLatest := s3KeyRelease(oLatest)
	keyLatest += suffix
	log.Printf("Uploading to s3://%s/%s", o.BucketName, keyLatest)
	putObjectInput := s3.PutObjectInput{
		Bucket:       &o.BucketName,
		Key:          &keyLatest,
		Body:         bytes.NewReader(buf.Bytes()),
		CacheControl: &release.NoCache,
	}
	if _, err := svc.PutObject(&putObjectInput); err != nil {
		__antithesis_instrumentation__.Notify(41740)
		log.Fatalf("s3 upload %s: %s", keyLatest, err)
	} else {
		__antithesis_instrumentation__.Notify(41741)
	}
}
