// reduce reduces SQL passed from the input file using cockroach demo. The input
// file is simplified such that -contains argument is present as an error during
// SQL execution. Run `make bin/reduce` to compile the reduce program.
package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/cmd/reduce/reduce"
	"github.com/cockroachdb/cockroach/pkg/cmd/reduce/reduce/reducesql"
	"github.com/cockroachdb/errors"
)

var (
	goroutines = func() int {
		__antithesis_instrumentation__.Notify(41742)

		n := (runtime.GOMAXPROCS(0) + 2) / 3
		if n < 1 {
			__antithesis_instrumentation__.Notify(41744)
			n = 1
		} else {
			__antithesis_instrumentation__.Notify(41745)
		}
		__antithesis_instrumentation__.Notify(41743)
		return n
	}()
	flags           = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	binary          = flags.String("binary", "./cockroach", "path to cockroach binary")
	file            = flags.String("file", "", "the path to a file containing SQL queries to reduce; required")
	verbose         = flags.Bool("v", false, "print progress to standard output and the original test case output if it is not interesting")
	contains        = flags.String("contains", "", "error regex to search for")
	unknown         = flags.Bool("unknown", false, "print unknown types during walk")
	workers         = flags.Int("goroutines", goroutines, "number of worker goroutines (defaults to NumCPU/3")
	chunkReductions = flags.Int("chunk", 0, "number of consecutive chunk reduction failures allowed before halting chunk reduction (default 0)")
	tlp             = flags.Bool("tlp", false, "last two non-empty lines in file are equivalent queries returning different results")
)

const description = `
The reduce utility attempts to simplify SQL that produces an error in
CockroachDB. The problematic SQL, specified via -file flag, is
repeatedly reduced as long as it produces an error in the CockroachDB
demo that matches the provided -contains regex.

An alternative mode of operation is enabled by specifying -tlp option:
in such case the last two queries in the file must be equivalent and
produce different results.

The following options are available:

`

func usage() {
	__antithesis_instrumentation__.Notify(41746)
	fmt.Fprintf(flags.Output(), "Usage of %s:\n", os.Args[0])
	fmt.Fprint(flags.Output(), description)
	flags.PrintDefaults()
	os.Exit(1)
}

func main() {
	__antithesis_instrumentation__.Notify(41747)
	flags.Usage = usage
	if err := flags.Parse(os.Args[1:]); err != nil {
		__antithesis_instrumentation__.Notify(41752)
		usage()
	} else {
		__antithesis_instrumentation__.Notify(41753)
	}
	__antithesis_instrumentation__.Notify(41748)
	if *file == "" {
		__antithesis_instrumentation__.Notify(41754)
		fmt.Printf("%s: -file must be provided\n\n", os.Args[0])
		usage()
	} else {
		__antithesis_instrumentation__.Notify(41755)
	}
	__antithesis_instrumentation__.Notify(41749)
	if *contains == "" && func() bool {
		__antithesis_instrumentation__.Notify(41756)
		return !*tlp == true
	}() == true {
		__antithesis_instrumentation__.Notify(41757)
		fmt.Printf("%s: either -contains must be provided or -tlp flag specified\n\n", os.Args[0])
		usage()
	} else {
		__antithesis_instrumentation__.Notify(41758)
	}
	__antithesis_instrumentation__.Notify(41750)
	reducesql.LogUnknown = *unknown
	out, err := reduceSQL(*binary, *contains, file, *workers, *verbose, *chunkReductions, *tlp)
	if err != nil {
		__antithesis_instrumentation__.Notify(41759)
		log.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(41760)
	}
	__antithesis_instrumentation__.Notify(41751)
	fmt.Println(out)
}

func reduceSQL(
	binary, contains string, file *string, workers int, verbose bool, chunkReductions int, tlp bool,
) (string, error) {
	__antithesis_instrumentation__.Notify(41761)
	const tlpFailureError = "TLP_FAILURE"
	if tlp {
		__antithesis_instrumentation__.Notify(41770)
		contains = tlpFailureError
	} else {
		__antithesis_instrumentation__.Notify(41771)
	}
	__antithesis_instrumentation__.Notify(41762)
	containsRE, err := regexp.Compile(contains)
	if err != nil {
		__antithesis_instrumentation__.Notify(41772)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(41773)
	}
	__antithesis_instrumentation__.Notify(41763)
	input, err := ioutil.ReadFile(*file)
	if err != nil {
		__antithesis_instrumentation__.Notify(41774)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(41775)
	}
	__antithesis_instrumentation__.Notify(41764)

	inputString := string(input)
	var tlpCheck string

	if tlp {
		__antithesis_instrumentation__.Notify(41776)
		lines := strings.Split(string(input), "\n")
		lineIdx := len(lines) - 1

		findPreviousQuery := func() string {
			__antithesis_instrumentation__.Notify(41778)

			for lines[lineIdx] == "" {
				__antithesis_instrumentation__.Notify(41781)
				lineIdx--
			}
			__antithesis_instrumentation__.Notify(41779)
			lastQueryLineIdx := lineIdx

			for lines[lineIdx] != "" {
				__antithesis_instrumentation__.Notify(41782)
				lineIdx--
			}
			__antithesis_instrumentation__.Notify(41780)

			query := strings.Join(lines[lineIdx+1:lastQueryLineIdx+1], " ")

			return query[:len(query)-1]
		}
		__antithesis_instrumentation__.Notify(41777)
		partitioned := findPreviousQuery()
		unpartitioned := findPreviousQuery()
		inputString = strings.Join(lines[:lineIdx], "\n")

		tlpCheck = fmt.Sprintf(`
SELECT CASE
  WHEN
    (SELECT count(*) FROM ((%[1]s) EXCEPT ALL (%[2]s))) != 0
  OR
    (SELECT count(*) FROM ((%[2]s) EXCEPT ALL (%[1]s))) != 0
  THEN
    crdb_internal.force_error('', '%[3]s')
  END;`, unpartitioned, partitioned, tlpFailureError)
	} else {
		__antithesis_instrumentation__.Notify(41783)
	}
	__antithesis_instrumentation__.Notify(41765)

	inputSQL, err := reducesql.Pretty(inputString)
	if err != nil {
		__antithesis_instrumentation__.Notify(41784)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(41785)
	}
	__antithesis_instrumentation__.Notify(41766)

	var logger *log.Logger
	if verbose {
		__antithesis_instrumentation__.Notify(41786)
		logger = log.New(os.Stderr, "", 0)
		logger.Printf("input SQL pretty printed, %d bytes -> %d bytes\n", len(input), len(inputSQL))
		if tlp {
			__antithesis_instrumentation__.Notify(41787)
			prettyTLPCheck, err := reducesql.Pretty(tlpCheck)
			if err != nil {
				__antithesis_instrumentation__.Notify(41789)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(41790)
			}
			__antithesis_instrumentation__.Notify(41788)
			logger.Printf("\nTLP check query:\n%s\n\n", prettyTLPCheck)
		} else {
			__antithesis_instrumentation__.Notify(41791)
		}
	} else {
		__antithesis_instrumentation__.Notify(41792)
	}
	__antithesis_instrumentation__.Notify(41767)

	var chunkReducer reduce.ChunkReducer
	if chunkReductions > 0 {
		__antithesis_instrumentation__.Notify(41793)
		chunkReducer = reducesql.NewSQLChunkReducer(chunkReductions)
	} else {
		__antithesis_instrumentation__.Notify(41794)
	}
	__antithesis_instrumentation__.Notify(41768)

	isInteresting := func(ctx context.Context, sql string) (interesting bool, logOriginalHint func()) {
		__antithesis_instrumentation__.Notify(41795)

		cmd := exec.CommandContext(ctx, binary,
			"demo",
			"--empty",
			"--disable-demo-license",
			"--set=errexit=false",
		)
		cmd.Env = []string{"COCKROACH_SKIP_ENABLING_DIAGNOSTIC_REPORTING", "true"}
		if !strings.HasSuffix(sql, ";") {
			__antithesis_instrumentation__.Notify(41799)
			sql += ";"
		} else {
			__antithesis_instrumentation__.Notify(41800)
		}
		__antithesis_instrumentation__.Notify(41796)

		sql += tlpCheck
		cmd.Stdin = strings.NewReader(sql)
		out, err := cmd.CombinedOutput()
		switch {
		case errors.HasType(err, (*exec.Error)(nil)):
			__antithesis_instrumentation__.Notify(41801)
			if errors.Is(err, exec.ErrNotFound) {
				__antithesis_instrumentation__.Notify(41804)
				log.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(41805)
			}
		case errors.HasType(err, (*os.PathError)(nil)):
			__antithesis_instrumentation__.Notify(41802)
			log.Fatal(err)
		default:
			__antithesis_instrumentation__.Notify(41803)
		}
		__antithesis_instrumentation__.Notify(41797)
		if verbose {
			__antithesis_instrumentation__.Notify(41806)
			logOriginalHint = func() {
				__antithesis_instrumentation__.Notify(41807)
				logger.Printf("output did not match regex %s:\n\n%s", contains, string(out))
			}
		} else {
			__antithesis_instrumentation__.Notify(41808)
		}
		__antithesis_instrumentation__.Notify(41798)
		return containsRE.Match(out), logOriginalHint
	}
	__antithesis_instrumentation__.Notify(41769)

	out, err := reduce.Reduce(
		logger,
		inputSQL,
		isInteresting,
		workers,
		reduce.ModeInteresting,
		chunkReducer,
		reducesql.SQLPasses...,
	)
	return out, err
}
