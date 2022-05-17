// testfilter is a utility to manipulate JSON streams in [test2json] format.
// Standard input is read and each line starting with `{` and ending with `}`
// parsed (and expected to parse successfully). Lines not matching this pattern
// are classified as output not related to the test and, depending on the args
// passed to `testfilter`, are passed through or removed. The arguments available
// are `--mode=(strip|omit|convert)`, where:
//
// strip: omit output for non-failing tests, pass everything else through. In
//   particular, non-test output and tests that never terminate are passed through.
// omit: print only failing tests. Note that test2json does not close scopes for
//   tests that are running in parallel (in the same package) with a "foreground"
//   test that panics, so it will pass through *only* the one foreground test.
//   Note also that package scopes are omitted; test2json does not reliably close
//   them on panic/Exit anyway.
// convert:
//   no filtering is performed, but any test2json input is translated back into
//   its pure Go test framework text representation. This is useful for output
//   intended for human eyes.
//
// [test2json]: https://golang.org/cmd/test2json/
package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
)

const modeUsage = `strip:
  omit output for non-failing tests, but print run/pass/skip events for all tests
omit:
  only emit failing tests
convert:
  don't perform any filtering, simply convert the json back to original test format'
`

type modeT byte

const (
	modeStrip modeT = iota
	modeOmit
	modeConvert
)

func (m *modeT) Set(s string) error {
	__antithesis_instrumentation__.Notify(53094)
	switch s {
	case "strip":
		__antithesis_instrumentation__.Notify(53096)
		*m = modeStrip
	case "omit":
		__antithesis_instrumentation__.Notify(53097)
		*m = modeOmit
	case "convert":
		__antithesis_instrumentation__.Notify(53098)
		*m = modeConvert
	default:
		__antithesis_instrumentation__.Notify(53099)
		return errors.New("unsupported mode")
	}
	__antithesis_instrumentation__.Notify(53095)
	return nil
}

func (m *modeT) String() string {
	__antithesis_instrumentation__.Notify(53100)
	switch *m {
	case modeStrip:
		__antithesis_instrumentation__.Notify(53101)
		return "strip"
	case modeOmit:
		__antithesis_instrumentation__.Notify(53102)
		return "omit"
	case modeConvert:
		__antithesis_instrumentation__.Notify(53103)
		return "convert"
	default:
		__antithesis_instrumentation__.Notify(53104)
		return "unknown"
	}
}

var modeVar = modeStrip

func init() {
	flag.Var(&modeVar, "mode", modeUsage)
}

type testEvent struct {
	Time    time.Time
	Action  string
	Package string
	Test    string
	Elapsed float64
	Output  string
}

func (t *testEvent) json() string {
	__antithesis_instrumentation__.Notify(53105)
	j, err := json.Marshal(t)
	if err != nil {
		__antithesis_instrumentation__.Notify(53107)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(53108)
	}
	__antithesis_instrumentation__.Notify(53106)
	return string(j)
}

func main() {
	__antithesis_instrumentation__.Notify(53109)
	flag.Parse()
	if err := filter(os.Stdin, os.Stdout, modeVar); err != nil {
		__antithesis_instrumentation__.Notify(53110)
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	} else {
		__antithesis_instrumentation__.Notify(53111)
	}
}

const packageLevelTestName = "PackageLevel"

type tup struct {
	pkg  string
	test string
}

type ent struct {
	first, last string
	strings.Builder

	numActiveTests int
	numTestsFailed int
}

func filter(in io.Reader, out io.Writer, mode modeT) error {
	__antithesis_instrumentation__.Notify(53112)
	scanner := bufio.NewScanner(in)
	m := map[tup]*ent{}
	ev := &testEvent{}
	var n int
	var passFailLine string
	for scanner.Scan() {
		__antithesis_instrumentation__.Notify(53116)
		line := scanner.Text()
		if len(line) <= 2 || func() bool {
			__antithesis_instrumentation__.Notify(53121)
			return line[0] != '{' == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(53122)
			return line[len(line)-1] != '}' == true
		}() == true {
			__antithesis_instrumentation__.Notify(53123)

			if passFailLine == "" && func() bool {
				__antithesis_instrumentation__.Notify(53126)
				return (strings.Contains(line, "PASS") || func() bool {
					__antithesis_instrumentation__.Notify(53127)
					return strings.Contains(line, "FAIL") == true
				}() == true) == true
			}() == true {
				__antithesis_instrumentation__.Notify(53128)
				passFailLine = line
			} else {
				__antithesis_instrumentation__.Notify(53129)
			}
			__antithesis_instrumentation__.Notify(53124)
			if mode != modeOmit {
				__antithesis_instrumentation__.Notify(53130)
				fmt.Fprintln(out, line)
			} else {
				__antithesis_instrumentation__.Notify(53131)
			}
			__antithesis_instrumentation__.Notify(53125)
			continue
		} else {
			__antithesis_instrumentation__.Notify(53132)
		}
		__antithesis_instrumentation__.Notify(53117)
		*ev = testEvent{}
		if err := json.Unmarshal([]byte(line), ev); err != nil {
			__antithesis_instrumentation__.Notify(53133)
			return err
		} else {
			__antithesis_instrumentation__.Notify(53134)
		}
		__antithesis_instrumentation__.Notify(53118)
		n++

		if mode == modeConvert {
			__antithesis_instrumentation__.Notify(53135)
			if ev.Action == "output" {
				__antithesis_instrumentation__.Notify(53137)
				fmt.Fprint(out, ev.Output)
			} else {
				__antithesis_instrumentation__.Notify(53138)
			}
			__antithesis_instrumentation__.Notify(53136)
			continue
		} else {
			__antithesis_instrumentation__.Notify(53139)
		}
		__antithesis_instrumentation__.Notify(53119)

		if ev.Test == "" {
			__antithesis_instrumentation__.Notify(53140)

			ev.Test = packageLevelTestName
			pkey := tup{ev.Package, ev.Test}

			switch ev.Action {
			case "fail":
				__antithesis_instrumentation__.Notify(53144)
				buf := m[pkey]

				hasOpenSubTests := buf != nil && func() bool {
					__antithesis_instrumentation__.Notify(53146)
					return buf.numActiveTests > 0 == true
				}() == true

				hadSomeFailedTests := buf != nil && func() bool {
					__antithesis_instrumentation__.Notify(53147)
					return buf.numTestsFailed > 0 == true
				}() == true

				if !hasOpenSubTests && func() bool {
					__antithesis_instrumentation__.Notify(53148)
					return hadSomeFailedTests == true
				}() == true {
					__antithesis_instrumentation__.Notify(53149)
					delete(m, pkey)
					continue
				} else {
					__antithesis_instrumentation__.Notify(53150)
				}
			default:
				__antithesis_instrumentation__.Notify(53145)
			}
			__antithesis_instrumentation__.Notify(53141)

			if err := ensurePackageEntry(m, out, ev, pkey, mode); err != nil {
				__antithesis_instrumentation__.Notify(53151)
				return err
			} else {
				__antithesis_instrumentation__.Notify(53152)
			}
			__antithesis_instrumentation__.Notify(53142)

			const helpMessage = `
Check full_output.txt in artifacts for stray panics or other errors that broke
the test process. Note that you might be looking at this message from a parent
CI job. To reliably get to the "correct" job, click the drop-down next to
"PackageLevel" above, then "Show in build log", and then navigate to the
artifacts tab. See:

https://user-images.githubusercontent.com/5076964/110923167-e2ab4780-8320-11eb-8fba-99da632aa814.png
https://user-images.githubusercontent.com/5076964/110923299-08d0e780-8321-11eb-91af-f4eedcf8bacb.png

for details.
`
			if ev.Action != "output" {
				__antithesis_instrumentation__.Notify(53153)

				var testReport strings.Builder
				for key := range m {
					__antithesis_instrumentation__.Notify(53155)
					if key.pkg == ev.Package && func() bool {
						__antithesis_instrumentation__.Notify(53156)
						return key.test != ev.Test == true
					}() == true {
						__antithesis_instrumentation__.Notify(53157)

						if strings.Contains(key.test, "/") {
							__antithesis_instrumentation__.Notify(53158)

							delete(m, key)
						} else {
							__antithesis_instrumentation__.Notify(53159)

							testReport.WriteString("\n" + key.test)

							syntheticSkipEv := testEvent{
								Time:    ev.Time,
								Action:  "skip",
								Package: ev.Package,
								Test:    key.test,
								Elapsed: 0,
								Output:  "unfinished due to package-level failure" + helpMessage,
							}

							syntheticLine := syntheticSkipEv.json()
							if err := processTestEvent(m, out, &syntheticSkipEv, syntheticLine, mode); err != nil {
								__antithesis_instrumentation__.Notify(53160)
								return err
							} else {
								__antithesis_instrumentation__.Notify(53161)
							}
						}
					} else {
						__antithesis_instrumentation__.Notify(53162)
					}
				}
				__antithesis_instrumentation__.Notify(53154)

				if ev.Action == "fail" {
					__antithesis_instrumentation__.Notify(53163)
					ev.Output += helpMessage
					if testReport.Len() > 0 {
						__antithesis_instrumentation__.Notify(53164)
						ev.Output += "\nThe following tests have not completed and could be the cause of the failure:" + testReport.String()
					} else {
						__antithesis_instrumentation__.Notify(53165)
					}
				} else {
					__antithesis_instrumentation__.Notify(53166)
				}
			} else {
				__antithesis_instrumentation__.Notify(53167)
			}
			__antithesis_instrumentation__.Notify(53143)

			line = ev.json()
		} else {
			__antithesis_instrumentation__.Notify(53168)
		}
		__antithesis_instrumentation__.Notify(53120)

		if err := processTestEvent(m, out, ev, line, mode); err != nil {
			__antithesis_instrumentation__.Notify(53169)
			return err
		} else {
			__antithesis_instrumentation__.Notify(53170)
		}
	}
	__antithesis_instrumentation__.Notify(53113)

	if mode == modeStrip {
		__antithesis_instrumentation__.Notify(53171)
		for key := range m {
			__antithesis_instrumentation__.Notify(53172)
			buf := m[key]

			if key.test == packageLevelTestName {
				__antithesis_instrumentation__.Notify(53174)
				continue
			} else {
				__antithesis_instrumentation__.Notify(53175)
			}
			__antithesis_instrumentation__.Notify(53173)
			fmt.Fprintln(out, buf.String())
		}
	} else {
		__antithesis_instrumentation__.Notify(53176)
	}
	__antithesis_instrumentation__.Notify(53114)

	if mode != modeConvert && func() bool {
		__antithesis_instrumentation__.Notify(53177)
		return n == 0 == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(53178)
		return passFailLine != "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(53179)

		return fmt.Errorf("not a single test was parsed, but detected test output: %s", passFailLine)
	} else {
		__antithesis_instrumentation__.Notify(53180)
	}
	__antithesis_instrumentation__.Notify(53115)
	return nil
}

func ensurePackageEntry(m map[tup]*ent, out io.Writer, ev *testEvent, pkey tup, mode modeT) error {
	__antithesis_instrumentation__.Notify(53181)
	if buf := m[pkey]; buf != nil {
		__antithesis_instrumentation__.Notify(53183)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(53184)
	}
	__antithesis_instrumentation__.Notify(53182)

	packageEvent := *ev
	packageEvent.Test = packageLevelTestName
	packageEvent.Action = "run"
	packageEvent.Output = ""
	packageLine := packageEvent.json()
	return processTestEvent(m, out, &packageEvent, packageLine, mode)
}

func processTestEvent(m map[tup]*ent, out io.Writer, ev *testEvent, line string, mode modeT) error {
	__antithesis_instrumentation__.Notify(53185)

	pkey := tup{ev.Package, packageLevelTestName}

	key := tup{ev.Package, ev.Test}

	if ev.Test != packageLevelTestName {
		__antithesis_instrumentation__.Notify(53190)
		if err := ensurePackageEntry(m, out, ev, pkey, mode); err != nil {
			__antithesis_instrumentation__.Notify(53191)
			return err
		} else {
			__antithesis_instrumentation__.Notify(53192)
		}
	} else {
		__antithesis_instrumentation__.Notify(53193)
	}
	__antithesis_instrumentation__.Notify(53186)

	buf := m[key]
	if buf == nil {
		__antithesis_instrumentation__.Notify(53194)
		buf = &ent{first: line}
		m[key] = buf
		if key != pkey {
			__antithesis_instrumentation__.Notify(53195)

			m[pkey].numActiveTests++
		} else {
			__antithesis_instrumentation__.Notify(53196)
		}
	} else {
		__antithesis_instrumentation__.Notify(53197)
	}
	__antithesis_instrumentation__.Notify(53187)
	if _, err := fmt.Fprintln(buf, line); err != nil {
		__antithesis_instrumentation__.Notify(53198)
		return err
	} else {
		__antithesis_instrumentation__.Notify(53199)
	}
	__antithesis_instrumentation__.Notify(53188)
	switch ev.Action {
	case "pass", "skip", "fail":
		__antithesis_instrumentation__.Notify(53200)
		buf.last = line
		if ev.Action == "fail" {
			__antithesis_instrumentation__.Notify(53204)
			fmt.Fprint(out, buf.String())
		} else {
			__antithesis_instrumentation__.Notify(53205)
			if mode == modeStrip {
				__antithesis_instrumentation__.Notify(53206)

				fmt.Fprintln(out, buf.first)
				fmt.Fprintln(out, buf.last)
			} else {
				__antithesis_instrumentation__.Notify(53207)
			}
		}
		__antithesis_instrumentation__.Notify(53201)

		delete(m, key)
		if key != pkey {
			__antithesis_instrumentation__.Notify(53208)
			m[pkey].numActiveTests--
			if ev.Action == "fail" {
				__antithesis_instrumentation__.Notify(53209)
				m[pkey].numTestsFailed++
			} else {
				__antithesis_instrumentation__.Notify(53210)
			}
		} else {
			__antithesis_instrumentation__.Notify(53211)
		}

	case "run", "pause", "cont", "bench", "output":
		__antithesis_instrumentation__.Notify(53202)
	default:
		__antithesis_instrumentation__.Notify(53203)

		return fmt.Errorf("unknown input: %s", line)
	}
	__antithesis_instrumentation__.Notify(53189)
	return nil
}
