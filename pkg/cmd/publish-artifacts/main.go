package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/cockroachdb/cockroach/pkg/release"
	"github.com/kr/pretty"
)

const (
	awsAccessKeyIDKey      = "AWS_ACCESS_KEY_ID"
	awsSecretAccessKeyKey  = "AWS_SECRET_ACCESS_KEY"
	teamcityBuildBranchKey = "TC_BUILD_BRANCH"
)

type s3putter interface {
	PutObject(*s3.PutObjectInput) (*s3.PutObjectOutput, error)
}

var testableS3 = func() (s3putter, error) {
	__antithesis_instrumentation__.Notify(41625)
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(41627)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(41628)
	}
	__antithesis_instrumentation__.Notify(41626)
	return s3.New(sess), nil
}

var destBucket = flag.String("bucket", "", "override default bucket")

func main() {
	__antithesis_instrumentation__.Notify(41629)
	flag.Parse()

	if _, ok := os.LookupEnv(awsAccessKeyIDKey); !ok {
		__antithesis_instrumentation__.Notify(41638)
		log.Fatalf("AWS access key ID environment variable %s is not set", awsAccessKeyIDKey)
	} else {
		__antithesis_instrumentation__.Notify(41639)
	}
	__antithesis_instrumentation__.Notify(41630)
	if _, ok := os.LookupEnv(awsSecretAccessKeyKey); !ok {
		__antithesis_instrumentation__.Notify(41640)
		log.Fatalf("AWS secret access key environment variable %s is not set", awsSecretAccessKeyKey)
	} else {
		__antithesis_instrumentation__.Notify(41641)
	}
	__antithesis_instrumentation__.Notify(41631)

	branch, ok := os.LookupEnv(teamcityBuildBranchKey)
	if !ok {
		__antithesis_instrumentation__.Notify(41642)
		log.Fatalf("VCS branch environment variable %s is not set", teamcityBuildBranchKey)
	} else {
		__antithesis_instrumentation__.Notify(41643)
	}
	__antithesis_instrumentation__.Notify(41632)
	pkg, err := os.Getwd()
	if err != nil {
		__antithesis_instrumentation__.Notify(41644)
		log.Fatalf("unable to locate CRDB directory: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(41645)
	}
	__antithesis_instrumentation__.Notify(41633)

	_, err = os.Stat(filepath.Join(pkg, "WORKSPACE"))
	if err != nil {
		__antithesis_instrumentation__.Notify(41646)
		log.Fatalf("unable to locate CRDB directory: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(41647)
	}
	__antithesis_instrumentation__.Notify(41634)

	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = pkg
	log.Printf("%s %s", cmd.Env, cmd.Args)
	out, err := cmd.Output()
	if err != nil {
		__antithesis_instrumentation__.Notify(41648)
		log.Fatalf("%s: out=%q err=%s", cmd.Args, out, err)
	} else {
		__antithesis_instrumentation__.Notify(41649)
	}
	__antithesis_instrumentation__.Notify(41635)
	versionStr := string(bytes.TrimSpace(out))

	svc, err := testableS3()
	if err != nil {
		__antithesis_instrumentation__.Notify(41650)
		log.Fatalf("Creating AWS S3 session: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(41651)
	}
	__antithesis_instrumentation__.Notify(41636)

	var bucketName string
	if len(*destBucket) > 0 {
		__antithesis_instrumentation__.Notify(41652)
		bucketName = *destBucket
	} else {
		__antithesis_instrumentation__.Notify(41653)
		bucketName = "cockroach"
	}
	__antithesis_instrumentation__.Notify(41637)
	log.Printf("Using S3 bucket: %s", bucketName)

	releaseVersionStrs := []string{versionStr}

	for _, platform := range []release.Platform{release.PlatformLinux, release.PlatformMacOS, release.PlatformWindows} {
		__antithesis_instrumentation__.Notify(41654)
		var o opts
		o.Platform = platform
		o.ReleaseVersionStrs = releaseVersionStrs
		o.PkgDir = pkg
		o.Branch = branch
		o.VersionStr = versionStr
		o.BucketName = bucketName
		o.Branch = branch
		o.AbsolutePath = filepath.Join(pkg, "cockroach"+release.SuffixFromPlatform(platform))

		log.Printf("building %s", pretty.Sprint(o))

		buildOneCockroach(svc, o)
	}
}

func buildOneCockroach(svc s3putter, o opts) {
	__antithesis_instrumentation__.Notify(41655)
	log.Printf("building cockroach %s", pretty.Sprint(o))
	defer func() {
		__antithesis_instrumentation__.Notify(41658)
		log.Printf("done building cockroach: %s", pretty.Sprint(o))
	}()
	__antithesis_instrumentation__.Notify(41656)

	if err := release.MakeRelease(o.Platform, release.BuildOptions{}, o.PkgDir); err != nil {
		__antithesis_instrumentation__.Notify(41659)
		log.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(41660)
	}
	__antithesis_instrumentation__.Notify(41657)

	putNonRelease(svc, o, release.MakeCRDBLibraryNonReleaseFiles(o.PkgDir, o.Platform, o.VersionStr)...)
}

type opts struct {
	VersionStr         string
	Branch             string
	ReleaseVersionStrs []string

	Platform release.Platform

	BucketName   string
	AbsolutePath string
	PkgDir       string
}

func putNonRelease(svc s3putter, o opts, additionalNonReleaseFiles ...release.NonReleaseFile) {
	__antithesis_instrumentation__.Notify(41661)
	release.PutNonRelease(
		svc,
		release.PutNonReleaseOptions{
			Branch:     o.Branch,
			BucketName: o.BucketName,
			Files: append(
				[]release.NonReleaseFile{release.MakeCRDBBinaryNonReleaseFile(o.AbsolutePath, o.VersionStr)},
				additionalNonReleaseFiles...,
			),
		},
	)
}
