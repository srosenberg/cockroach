// smithtest is a tool to execute sqlsmith tests on cockroach demo
// instances. Failures are tracked, de-duplicated, reduced. Issues are
// prefilled for GitHub.
package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"bytes"
	"context"
	gosql "database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/internal/sqlsmith"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/google/go-github/github"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/lib/pq"
	"github.com/pkg/browser"
)

var (
	flags     = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	cockroach = flags.String("cockroach", "./cockroach", "path to cockroach binary")
	reduce    = flags.String("reduce", "./bin/reduce", "path to reduce binary")
	num       = flags.Int("num", 1, "number of parallel testers")
)

func usage() {
	__antithesis_instrumentation__.Notify(52947)
	fmt.Fprintf(flags.Output(), "Usage of %s:\n", os.Args[0])
	flags.PrintDefaults()
	os.Exit(1)
}

func main() {
	__antithesis_instrumentation__.Notify(52948)
	if err := flags.Parse(os.Args[1:]); err != nil {
		__antithesis_instrumentation__.Notify(52951)
		usage()
	} else {
		__antithesis_instrumentation__.Notify(52952)
	}
	__antithesis_instrumentation__.Notify(52949)

	ctx := context.Background()
	setup := WorkerSetup{
		cockroach: *cockroach,
		reduce:    *reduce,
		github:    github.NewClient(nil),
	}
	rand.Seed(timeutil.Now().UnixNano())

	setup.populateGitHubIssues(ctx)

	fmt.Println("running...")

	g := ctxgroup.WithContext(ctx)
	for i := 0; i < *num; i++ {
		__antithesis_instrumentation__.Notify(52953)
		g.GoCtx(setup.work)
	}
	__antithesis_instrumentation__.Notify(52950)
	if err := g.Wait(); err != nil {
		__antithesis_instrumentation__.Notify(52954)
		log.Fatalf("%+v", err)
	} else {
		__antithesis_instrumentation__.Notify(52955)
	}
}

type WorkerSetup struct {
	cockroach, reduce string
	github            *github.Client
}

func (s WorkerSetup) populateGitHubIssues(ctx context.Context) {
	__antithesis_instrumentation__.Notify(52956)
	var opts github.SearchOptions
	for {
		__antithesis_instrumentation__.Notify(52957)
		results, _, err := s.github.Search.Issues(ctx, "repo:cockroachdb/cockroach type:issue state:open label:C-bug label:O-sqlsmith", &opts)
		if err != nil {
			__antithesis_instrumentation__.Notify(52961)
			log.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(52962)
		}
		__antithesis_instrumentation__.Notify(52958)
		for _, issue := range results.Issues {
			__antithesis_instrumentation__.Notify(52963)
			title := filterIssueTitle(issue.GetTitle())
			seenIssues[title] = true
			fmt.Println("pre populate", title)
		}
		__antithesis_instrumentation__.Notify(52959)
		if results.GetIncompleteResults() {
			__antithesis_instrumentation__.Notify(52964)
			opts.Page++
			continue
		} else {
			__antithesis_instrumentation__.Notify(52965)
		}
		__antithesis_instrumentation__.Notify(52960)
		return
	}
}

func (s WorkerSetup) work(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(52966)
	rnd := rand.New(rand.NewSource(rand.Int63()))
	for {
		__antithesis_instrumentation__.Notify(52967)
		if err := s.run(ctx, rnd); err != nil {
			__antithesis_instrumentation__.Notify(52968)
			return err
		} else {
			__antithesis_instrumentation__.Notify(52969)
		}
	}
}

var (
	lock syncutil.RWMutex

	seenIssues = map[string]bool{}

	connRE         = regexp.MustCompile(`(?m)^sql:\s*(postgresql://.*)$`)
	panicRE        = regexp.MustCompile(`(?m)^(panic: .*?)( \[recovered\])?$`)
	stackRE        = regexp.MustCompile(`panic: .*\n\ngoroutine \d+ \[running\]:\n(?s:(.*))$`)
	fatalRE        = regexp.MustCompile(`(?m)^(fatal error: .*?)$`)
	runtimeStackRE = regexp.MustCompile(`goroutine \d+ \[running\]:\n(?s:(.*?))\n\n`)
)

func (s WorkerSetup) run(ctx context.Context, rnd *rand.Rand) error {
	__antithesis_instrumentation__.Notify(52970)

	done := timeutil.Now().Add(time.Minute)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	cmd := exec.CommandContext(ctx, s.cockroach,
		"start-single-node",
		"--port", "0",
		"--http-port", "0",
		"--insecure",
		"--store=type=mem,size=1GB",
		"--logtostderr",
	)

	var pgdb *pgx.Conn
	var db *gosql.DB
	var output bytes.Buffer

	stderr, err := cmd.StderrPipe()
	if err != nil {
		__antithesis_instrumentation__.Notify(52978)
		return err
	} else {
		__antithesis_instrumentation__.Notify(52979)
	}
	__antithesis_instrumentation__.Notify(52971)
	if err := cmd.Start(); err != nil {
		__antithesis_instrumentation__.Notify(52980)
		return errors.Wrap(err, "start")
	} else {
		__antithesis_instrumentation__.Notify(52981)
	}
	__antithesis_instrumentation__.Notify(52972)

	scanner := bufio.NewScanner(io.TeeReader(stderr, &output))
	for scanner.Scan() {
		__antithesis_instrumentation__.Notify(52982)
		line := scanner.Text()
		if match := connRE.FindStringSubmatch(line); match != nil {
			__antithesis_instrumentation__.Notify(52983)
			config, err := pgx.ParseConfig(match[1])
			if err != nil {
				__antithesis_instrumentation__.Notify(52987)
				return errors.Wrap(err, "parse uri")
			} else {
				__antithesis_instrumentation__.Notify(52988)
			}
			__antithesis_instrumentation__.Notify(52984)
			pgdb, err = pgx.ConnectConfig(ctx, config)
			if err != nil {
				__antithesis_instrumentation__.Notify(52989)
				return errors.Wrap(err, "connect")
			} else {
				__antithesis_instrumentation__.Notify(52990)
			}
			__antithesis_instrumentation__.Notify(52985)

			connector, err := pq.NewConnector(match[1])
			if err != nil {
				__antithesis_instrumentation__.Notify(52991)
				return errors.Wrap(err, "connector error")
			} else {
				__antithesis_instrumentation__.Notify(52992)
			}
			__antithesis_instrumentation__.Notify(52986)
			db = gosql.OpenDB(connector)
			fmt.Println("connected to", match[1])
			break
		} else {
			__antithesis_instrumentation__.Notify(52993)
		}
	}
	__antithesis_instrumentation__.Notify(52973)
	if err := scanner.Err(); err != nil {
		__antithesis_instrumentation__.Notify(52994)
		fmt.Println(output.String())
		return errors.Wrap(err, "scanner error")
	} else {
		__antithesis_instrumentation__.Notify(52995)
	}
	__antithesis_instrumentation__.Notify(52974)
	if db == nil {
		__antithesis_instrumentation__.Notify(52996)
		fmt.Println(output.String())
		return errors.New("no DB address found")
	} else {
		__antithesis_instrumentation__.Notify(52997)
	}
	__antithesis_instrumentation__.Notify(52975)
	fmt.Println("worker started")

	initSQL := sqlsmith.Setups[sqlsmith.RandSetup(rnd)](rnd)
	for _, stmt := range initSQL {
		__antithesis_instrumentation__.Notify(52998)
		if _, err := pgdb.Exec(ctx, stmt); err != nil {
			__antithesis_instrumentation__.Notify(52999)
			return errors.Wrap(err, "init")
		} else {
			__antithesis_instrumentation__.Notify(53000)
		}
	}
	__antithesis_instrumentation__.Notify(52976)

	setting := sqlsmith.Settings[sqlsmith.RandSetting(rnd)](rnd)
	opts := append([]sqlsmith.SmitherOption{
		sqlsmith.DisableMutations(),
	}, setting.Options...)
	smither, err := sqlsmith.NewSmither(db, rnd, opts...)
	if err != nil {
		__antithesis_instrumentation__.Notify(53001)
		return errors.Wrap(err, "new smither")
	} else {
		__antithesis_instrumentation__.Notify(53002)
	}
	__antithesis_instrumentation__.Notify(52977)
	for {
		__antithesis_instrumentation__.Notify(53003)
		if timeutil.Now().After(done) {
			__antithesis_instrumentation__.Notify(53008)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(53009)
		}
		__antithesis_instrumentation__.Notify(53004)

		lock.RLock()
		stmt := smither.Generate()
		done := make(chan struct{}, 1)
		go func() {
			__antithesis_instrumentation__.Notify(53010)
			_, err = pgdb.Exec(ctx, stmt)
			done <- struct{}{}
		}()
		__antithesis_instrumentation__.Notify(53005)

		select {
		case <-time.After(10 * time.Second):
			__antithesis_instrumentation__.Notify(53011)
			fmt.Printf("TIMEOUT:\n%s\n", stmt)
			lock.RUnlock()
			return nil
		case <-done:
			__antithesis_instrumentation__.Notify(53012)
		}
		__antithesis_instrumentation__.Notify(53006)
		lock.RUnlock()
		if err != nil {
			__antithesis_instrumentation__.Notify(53013)
			if strings.Contains(err.Error(), "internal error") {
				__antithesis_instrumentation__.Notify(53014)

				return s.failure(ctx, initSQL, stmt, err)
			} else {
				__antithesis_instrumentation__.Notify(53015)
			}

		} else {
			__antithesis_instrumentation__.Notify(53016)
		}
		__antithesis_instrumentation__.Notify(53007)

		if err := db.PingContext(ctx); err != nil {
			__antithesis_instrumentation__.Notify(53017)
			input := fmt.Sprintf("%s; %s;", initSQL, stmt)
			out, _ := exec.CommandContext(ctx, s.cockroach, "demo", "--no-example-database", "-e", input).CombinedOutput()
			var pqerr pq.Error
			if match := stackRE.FindStringSubmatch(string(out)); match != nil {
				__antithesis_instrumentation__.Notify(53022)
				pqerr.Detail = strings.TrimSpace(match[1])
			} else {
				__antithesis_instrumentation__.Notify(53023)
			}
			__antithesis_instrumentation__.Notify(53018)
			if match := panicRE.FindStringSubmatch(string(out)); match != nil {
				__antithesis_instrumentation__.Notify(53024)

				pqerr.Message = match[1]
				return s.failure(ctx, initSQL, stmt, &pqerr)
			} else {
				__antithesis_instrumentation__.Notify(53025)
			}
			__antithesis_instrumentation__.Notify(53019)

			if match := runtimeStackRE.FindStringSubmatch(string(out)); match != nil {
				__antithesis_instrumentation__.Notify(53026)
				pqerr.Detail = strings.TrimSpace(match[1])
			} else {
				__antithesis_instrumentation__.Notify(53027)
			}
			__antithesis_instrumentation__.Notify(53020)
			if match := fatalRE.FindStringSubmatch(string(out)); match != nil {
				__antithesis_instrumentation__.Notify(53028)

				pqerr.Message = match[1]
				return s.failure(ctx, initSQL, stmt, &pqerr)
			} else {
				__antithesis_instrumentation__.Notify(53029)
			}
			__antithesis_instrumentation__.Notify(53021)

			fmt.Printf("output:\n%s\n", out)
			fmt.Printf("Ping stmt:\n%s;\n", stmt)
			return err
		} else {
			__antithesis_instrumentation__.Notify(53030)
		}
	}
}

func (s WorkerSetup) failure(ctx context.Context, initSQL []string, stmt string, err error) error {
	__antithesis_instrumentation__.Notify(53031)
	var message, stack string
	var pqerr pgconn.PgError
	if errors.As(err, &pqerr) {
		__antithesis_instrumentation__.Notify(53040)
		stack = pqerr.Detail
		message = pqerr.Message
	} else {
		__antithesis_instrumentation__.Notify(53041)
		message = err.Error()
	}
	__antithesis_instrumentation__.Notify(53032)
	filteredMessage := filterIssueTitle(regexp.QuoteMeta(message))
	message = fmt.Sprintf("sql: %s", message)

	lock.Lock()

	defer lock.Unlock()
	sqlFilteredMessage := fmt.Sprintf("sql: %s", filteredMessage)
	alreadySeen := seenIssues[sqlFilteredMessage]
	if !alreadySeen {
		__antithesis_instrumentation__.Notify(53042)
		seenIssues[sqlFilteredMessage] = true
	} else {
		__antithesis_instrumentation__.Notify(53043)
	}
	__antithesis_instrumentation__.Notify(53033)
	if alreadySeen {
		__antithesis_instrumentation__.Notify(53044)
		fmt.Println("already found", message)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(53045)
	}
	__antithesis_instrumentation__.Notify(53034)
	fmt.Println("found", message)
	input := fmt.Sprintf("%s\n\n%s;", strings.Join(initSQL, "\n"), stmt)
	fmt.Printf("SQL:\n%s\n\n", input)

	cmd := exec.CommandContext(ctx, s.reduce, "-v", "-contains", filteredMessage)
	cmd.Stdin = strings.NewReader(input)
	cmd.Stderr = os.Stderr
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		__antithesis_instrumentation__.Notify(53046)
		fmt.Println(input)
		return err
	} else {
		__antithesis_instrumentation__.Notify(53047)
	}
	__antithesis_instrumentation__.Notify(53035)

	makeBody := func() string {
		__antithesis_instrumentation__.Notify(53048)
		return fmt.Sprintf("```\n%s\n```\n\n```\n%s\n```", strings.TrimSpace(out.String()), strings.TrimSpace(stack))
	}
	__antithesis_instrumentation__.Notify(53036)
	query := url.Values{
		"title":  []string{message},
		"labels": []string{"C-bug,O-sqlsmith"},
		"body":   []string{makeBody()},
	}
	url := url.URL{
		Scheme:   "https",
		Host:     "github.com",
		Path:     "/cockroachdb/cockroach/issues/new",
		RawQuery: query.Encode(),
	}
	const max = 8000

	for len(url.String()) > max {
		__antithesis_instrumentation__.Notify(53049)
		last := strings.LastIndex(stack, "\n")
		if last < 0 {
			__antithesis_instrumentation__.Notify(53051)
			break
		} else {
			__antithesis_instrumentation__.Notify(53052)
		}
		__antithesis_instrumentation__.Notify(53050)
		stack = stack[:last]
		query["body"][0] = makeBody()
		url.RawQuery = query.Encode()
	}
	__antithesis_instrumentation__.Notify(53037)
	if len(url.String()) > max {
		__antithesis_instrumentation__.Notify(53053)
		fmt.Println(stmt)
		return errors.New("request could not be shortened to max length")
	} else {
		__antithesis_instrumentation__.Notify(53054)
	}
	__antithesis_instrumentation__.Notify(53038)

	if err := browser.OpenURL(url.String()); err != nil {
		__antithesis_instrumentation__.Notify(53055)
		return err
	} else {
		__antithesis_instrumentation__.Notify(53056)
	}
	__antithesis_instrumentation__.Notify(53039)

	return nil
}

func filterIssueTitle(s string) string {
	__antithesis_instrumentation__.Notify(53057)
	for _, reS := range []string{
		`given: .*, expected .*`,
		`Datum is .*, not .*`,
		`expected .*, found .*`,
		`\d+`,
		`\*tree\.D\w+`,
	} {
		__antithesis_instrumentation__.Notify(53059)
		re := regexp.MustCompile(reS)
		s = re.ReplaceAllString(s, reS)
	}
	__antithesis_instrumentation__.Notify(53058)
	return s
}
