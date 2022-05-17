package urlcheck

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/ghemawat/stream"
)

const maxConcurrentRequests = 20

const timeoutRetries = 3

const URLRE = `https?://[^ \t\n/$.?#].[^ \t\n"<]*`

var re = regexp.MustCompile(URLRE)

var ignored = []string{

	"http://%s",
	"http://127.0.0.1",
	"http://\"",
	"http://HOST:PORT/",
	"http://localhost",
	"http://s3.amazonaws.com/cockroach/cockroach/cockroach.$(",
	"https://127.0.0.1",
	"https://\"",
	"https://binaries.cockroachdb.com/cockroach-${BETA_TAG}",
	"https://edge-binaries.cockroachdb.com/cockroach/cockroach.linux-gnu-amd64.${var.cockroach_sha}",
	"https://edge-binaries.cockroachdb.com/examples-go/block_writer.${var.block_writer_sha}",
	"https://edge-binaries.cockroachdb.com/examples-go/photos.${var.photos_sha}",
	"https://binaries.cockroachdb.com/cockroach-${FORWARD_REFERENCE_VERSION}.linux-amd64.tgz",
	"https://binaries.cockroachdb.com/cockroach-${BIDIRECTIONAL_REFERENCE_VERSION}.linux-amd64.tgz",
	"https://github.com/cockroachdb/cockroach/commits/%s",
	"https://github.com/cockroachdb/cockroach/issues/%d",
	"https://ignored:ignored@ignored/ignored",
	"https://localhost",
	"https://myroach:8080",
	"https://storage.googleapis.com/golang/go${GOVERSION}",
	"https://%s.blob.core.windows.net/%s",
	"https://roachfixtureseastus.blob.core.windows.net/$container",
	"https://roachfixtureswestus.blob.core.windows.net/$container",

	"http://instance-data/latest/meta-data/public-ipv4",
	"http://metadata/",

	"https://registry.yarnpkg.com/",

	"http://www.bohemiancoding.com/sketch/ns",
	"http://www.w3.org/",
	"https://www.w3.org/",
	"http://maven.apache.org/POM/4.0.0",

	"http://ignored:ignored@errors.cockroachdb.com/",
	"https://cockroachdb.github.io/distsqlplan/",
	"https://ignored:ignored@errors.cockroachdb.com/",
	"https://register.cockroachdb.com",
	"https://www.googleapis.com/auth",

	"https://console.aws.amazon.com/",
	"https://github.com/cockroachlabs/registration/",
	"https://api.github.com/repos/cockroachdb/cockroach/issues",
	"https://index.docker.io/v1/",
}

func chompUnbalanced(left, right rune, s string) string {
	__antithesis_instrumentation__.Notify(53242)
	depth := 0
	for i, c := range s {
		__antithesis_instrumentation__.Notify(53244)
		switch c {
		case left:
			__antithesis_instrumentation__.Notify(53245)
			depth++
		case right:
			__antithesis_instrumentation__.Notify(53246)
			if depth == 0 {
				__antithesis_instrumentation__.Notify(53249)
				return s[:i]
			} else {
				__antithesis_instrumentation__.Notify(53250)
			}
			__antithesis_instrumentation__.Notify(53247)
			depth--
		default:
			__antithesis_instrumentation__.Notify(53248)
		}
	}
	__antithesis_instrumentation__.Notify(53243)
	return s
}

func checkURL(client *http.Client, url string) error {
	__antithesis_instrumentation__.Notify(53251)
	resp, err := client.Head(url)
	if err != nil {
		__antithesis_instrumentation__.Notify(53258)
		return err
	} else {
		__antithesis_instrumentation__.Notify(53259)
	}
	__antithesis_instrumentation__.Notify(53252)
	if err := resp.Body.Close(); err != nil {
		__antithesis_instrumentation__.Notify(53260)
		return err
	} else {
		__antithesis_instrumentation__.Notify(53261)
	}
	__antithesis_instrumentation__.Notify(53253)

	if resp.StatusCode >= 200 && func() bool {
		__antithesis_instrumentation__.Notify(53262)
		return resp.StatusCode < 300 == true
	}() == true {
		__antithesis_instrumentation__.Notify(53263)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(53264)
	}
	__antithesis_instrumentation__.Notify(53254)

	resp, err = client.Get(url)
	if err != nil {
		__antithesis_instrumentation__.Notify(53265)
		return err
	} else {
		__antithesis_instrumentation__.Notify(53266)
	}
	__antithesis_instrumentation__.Notify(53255)
	if err := resp.Body.Close(); err != nil {
		__antithesis_instrumentation__.Notify(53267)
		return err
	} else {
		__antithesis_instrumentation__.Notify(53268)
	}
	__antithesis_instrumentation__.Notify(53256)

	if resp.StatusCode >= 200 && func() bool {
		__antithesis_instrumentation__.Notify(53269)
		return resp.StatusCode < 300 == true
	}() == true {
		__antithesis_instrumentation__.Notify(53270)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(53271)
	}
	__antithesis_instrumentation__.Notify(53257)

	return errors.Newf("%s", errors.Safe(resp.Status))
}

func checkURLWithRetries(client *http.Client, url string) error {
	__antithesis_instrumentation__.Notify(53272)
	for i := 0; i < timeoutRetries; i++ {
		__antithesis_instrumentation__.Notify(53274)
		err := checkURL(client, url)
		if netErr := (net.Error)(nil); errors.As(err, &netErr) && func() bool {
			__antithesis_instrumentation__.Notify(53276)
			return netErr.Timeout() == true
		}() == true {
			__antithesis_instrumentation__.Notify(53277)

			time.Sleep((1 << uint(i)) * time.Second)
			continue
		} else {
			__antithesis_instrumentation__.Notify(53278)
		}
		__antithesis_instrumentation__.Notify(53275)
		return err
	}
	__antithesis_instrumentation__.Notify(53273)
	return fmt.Errorf("timed out on %d separate tries, giving up", timeoutRetries)
}

func CheckURLsFromGrepOutput(cmd *exec.Cmd) error {
	__antithesis_instrumentation__.Notify(53279)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		__antithesis_instrumentation__.Notify(53284)
		log.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(53285)
	}
	__antithesis_instrumentation__.Notify(53280)
	filter := stream.ReadLines(stdout)
	stderr := new(bytes.Buffer)
	cmd.Stderr = stderr

	if err := cmd.Start(); err != nil {
		__antithesis_instrumentation__.Notify(53286)
		log.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(53287)
	}
	__antithesis_instrumentation__.Notify(53281)

	uniqueURLs, err := getURLs(filter)
	if err != nil {
		__antithesis_instrumentation__.Notify(53288)
		log.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(53289)
	}
	__antithesis_instrumentation__.Notify(53282)

	if err := cmd.Wait(); err != nil {
		__antithesis_instrumentation__.Notify(53290)
		log.Fatalf("err=%s, stderr=%s", err, stderr.String())
	} else {
		__antithesis_instrumentation__.Notify(53291)
	}
	__antithesis_instrumentation__.Notify(53283)
	return checkURLs(uniqueURLs)
}

func getURLs(filter stream.Filter) (map[string][]string, error) {
	__antithesis_instrumentation__.Notify(53292)
	uniqueURLs := map[string][]string{}

	if err := stream.ForEach(filter, func(s string) {
		__antithesis_instrumentation__.Notify(53294)
	outer:
		for _, match := range re.FindAllString(s, -1) {
			__antithesis_instrumentation__.Notify(53295)

			match = chompUnbalanced('(', ')', match)
			match = chompUnbalanced('[', ']', match)

			match = strings.TrimRight(match, ".,;\\\">`]")

			n := strings.LastIndexByte(match, '#')
			if n != -1 {
				__antithesis_instrumentation__.Notify(53299)
				match = match[:n]
			} else {
				__antithesis_instrumentation__.Notify(53300)
			}
			__antithesis_instrumentation__.Notify(53296)

			if len(s) > 150 {
				__antithesis_instrumentation__.Notify(53301)
				s = s[:150] + "..."
			} else {
				__antithesis_instrumentation__.Notify(53302)
			}
			__antithesis_instrumentation__.Notify(53297)
			for _, ig := range ignored {
				__antithesis_instrumentation__.Notify(53303)
				if strings.HasPrefix(match, ig) {
					__antithesis_instrumentation__.Notify(53304)
					continue outer
				} else {
					__antithesis_instrumentation__.Notify(53305)
				}
			}
			__antithesis_instrumentation__.Notify(53298)
			uniqueURLs[match] = append(uniqueURLs[match], s)
		}
	}); err != nil {
		__antithesis_instrumentation__.Notify(53306)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(53307)
	}
	__antithesis_instrumentation__.Notify(53293)

	return uniqueURLs, nil
}

func checkURLs(uniqueURLs map[string][]string) error {
	__antithesis_instrumentation__.Notify(53308)
	sem := make(chan struct{}, maxConcurrentRequests)
	errChan := make(chan error, len(uniqueURLs))

	client := &http.Client{
		Transport: &http.Transport{

			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: time.Minute,
	}

	for url, locs := range uniqueURLs {
		__antithesis_instrumentation__.Notify(53312)
		sem <- struct{}{}
		go func(url string, locs []string) {
			__antithesis_instrumentation__.Notify(53313)
			defer func() { __antithesis_instrumentation__.Notify(53315); <-sem }()
			__antithesis_instrumentation__.Notify(53314)
			log.Printf("Checking %s...", url)
			if err := checkURLWithRetries(client, url); err != nil {
				__antithesis_instrumentation__.Notify(53316)
				var buf bytes.Buffer
				fmt.Fprintf(&buf, "%s : %s\n", url, err)
				for _, loc := range locs {
					__antithesis_instrumentation__.Notify(53318)
					fmt.Fprintln(&buf, "    ", loc)
				}
				__antithesis_instrumentation__.Notify(53317)
				errChan <- errors.Newf("%s", buf.String())
			} else {
				__antithesis_instrumentation__.Notify(53319)
				errChan <- nil
			}
		}(url, locs)
	}
	__antithesis_instrumentation__.Notify(53309)

	var errs []error
	for i := 0; i < len(uniqueURLs); i++ {
		__antithesis_instrumentation__.Notify(53320)
		if err := <-errChan; err != nil {
			__antithesis_instrumentation__.Notify(53321)
			errs = append(errs, err)
		} else {
			__antithesis_instrumentation__.Notify(53322)
		}
	}
	__antithesis_instrumentation__.Notify(53310)
	if len(errs) > 0 {
		__antithesis_instrumentation__.Notify(53323)
		var buf bytes.Buffer
		for _, err := range errs {
			__antithesis_instrumentation__.Notify(53325)
			fmt.Fprintln(&buf, err)
		}
		__antithesis_instrumentation__.Notify(53324)
		fmt.Fprintf(&buf, "%d errors\n", len(errs))
		return errors.Newf("%s", buf.String())
	} else {
		__antithesis_instrumentation__.Notify(53326)
	}
	__antithesis_instrumentation__.Notify(53311)
	return nil
}
