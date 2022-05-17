package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

type zipper struct {
	syncutil.Mutex

	f *os.File
	z *zip.Writer
}

func newZipper(f *os.File) *zipper {
	__antithesis_instrumentation__.Notify(35350)
	return &zipper{
		f: f,
		z: zip.NewWriter(f),
	}
}

func (z *zipper) close() error {
	__antithesis_instrumentation__.Notify(35351)
	z.Lock()
	defer z.Unlock()

	err1 := z.z.Close()
	err2 := z.f.Close()
	return errors.CombineErrors(err1, err2)
}

func (z *zipper) createLocked(name string, mtime time.Time) (io.Writer, error) {
	__antithesis_instrumentation__.Notify(35352)
	if mtime.IsZero() {
		__antithesis_instrumentation__.Notify(35354)
		mtime = timeutil.Now()
	} else {
		__antithesis_instrumentation__.Notify(35355)
	}
	__antithesis_instrumentation__.Notify(35353)
	return z.z.CreateHeader(&zip.FileHeader{
		Name:     name,
		Method:   zip.Deflate,
		Modified: mtime,
	})
}

func (z *zipper) createRaw(s *zipReporter, name string, b []byte) error {
	__antithesis_instrumentation__.Notify(35356)
	z.Lock()
	defer z.Unlock()

	s.progress("writing binary output: %s", name)
	w, err := z.createLocked(name, time.Time{})
	if err != nil {
		__antithesis_instrumentation__.Notify(35358)
		return s.fail(err)
	} else {
		__antithesis_instrumentation__.Notify(35359)
	}
	__antithesis_instrumentation__.Notify(35357)
	_, err = w.Write(b)
	return s.result(err)
}

func (z *zipper) createJSON(s *zipReporter, name string, m interface{}) (err error) {
	__antithesis_instrumentation__.Notify(35360)
	if !strings.HasSuffix(name, ".json") {
		__antithesis_instrumentation__.Notify(35363)
		return s.fail(errors.Errorf("%s does not have .json suffix", name))
	} else {
		__antithesis_instrumentation__.Notify(35364)
	}
	__antithesis_instrumentation__.Notify(35361)
	s.progress("converting to JSON")
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		__antithesis_instrumentation__.Notify(35365)
		return s.fail(err)
	} else {
		__antithesis_instrumentation__.Notify(35366)
	}
	__antithesis_instrumentation__.Notify(35362)
	return z.createRaw(s, name, b)
}

func (z *zipper) createError(s *zipReporter, name string, e error) error {
	__antithesis_instrumentation__.Notify(35367)
	z.Lock()
	defer z.Unlock()

	s.shout("last request failed: %v", e)
	out := name + ".err.txt"
	s.progress("creating error output: %s", out)
	w, err := z.createLocked(out, time.Time{})
	if err != nil {
		__antithesis_instrumentation__.Notify(35369)
		return s.fail(err)
	} else {
		__antithesis_instrumentation__.Notify(35370)
	}
	__antithesis_instrumentation__.Notify(35368)
	fmt.Fprintf(w, "%+v\n", e)
	s.done()
	return nil
}

func (z *zipper) createJSONOrError(s *zipReporter, name string, m interface{}, e error) error {
	__antithesis_instrumentation__.Notify(35371)
	if e != nil {
		__antithesis_instrumentation__.Notify(35373)
		return z.createError(s, name, e)
	} else {
		__antithesis_instrumentation__.Notify(35374)
	}
	__antithesis_instrumentation__.Notify(35372)
	return z.createJSON(s, name, m)
}

func (z *zipper) createRawOrError(s *zipReporter, name string, b []byte, e error) error {
	__antithesis_instrumentation__.Notify(35375)
	if filepath.Ext(name) == "" {
		__antithesis_instrumentation__.Notify(35378)
		return errors.Errorf("%s has no extension", name)
	} else {
		__antithesis_instrumentation__.Notify(35379)
	}
	__antithesis_instrumentation__.Notify(35376)
	if e != nil {
		__antithesis_instrumentation__.Notify(35380)
		return z.createError(s, name, e)
	} else {
		__antithesis_instrumentation__.Notify(35381)
	}
	__antithesis_instrumentation__.Notify(35377)
	return z.createRaw(s, name, b)
}

type nodeSelection struct {
	inclusive     rangeSelection
	exclusive     rangeSelection
	includedCache map[int]struct{}
	excludedCache map[int]struct{}
}

func (n *nodeSelection) isIncluded(nodeID roachpb.NodeID) bool {
	__antithesis_instrumentation__.Notify(35382)

	if n.includedCache == nil {
		__antithesis_instrumentation__.Notify(35387)
		n.includedCache = n.inclusive.items()
	} else {
		__antithesis_instrumentation__.Notify(35388)
	}
	__antithesis_instrumentation__.Notify(35383)
	if n.excludedCache == nil {
		__antithesis_instrumentation__.Notify(35389)
		n.excludedCache = n.exclusive.items()
	} else {
		__antithesis_instrumentation__.Notify(35390)
	}
	__antithesis_instrumentation__.Notify(35384)

	isIncluded := true
	if len(n.includedCache) > 0 {
		__antithesis_instrumentation__.Notify(35391)
		_, isIncluded = n.includedCache[int(nodeID)]
	} else {
		__antithesis_instrumentation__.Notify(35392)
	}
	__antithesis_instrumentation__.Notify(35385)

	if _, excluded := n.excludedCache[int(nodeID)]; excluded {
		__antithesis_instrumentation__.Notify(35393)
		isIncluded = false
	} else {
		__antithesis_instrumentation__.Notify(35394)
	}
	__antithesis_instrumentation__.Notify(35386)
	return isIncluded
}

type rangeSelection struct {
	input  string
	ranges []vrange
}

type vrange struct {
	a, b int
}

func (r *rangeSelection) String() string {
	__antithesis_instrumentation__.Notify(35395)
	return r.input
}

func (r *rangeSelection) Type() string {
	__antithesis_instrumentation__.Notify(35396)
	return "a-b,c,d-e,..."
}

func (r *rangeSelection) Set(v string) error {
	__antithesis_instrumentation__.Notify(35397)
	r.input = v
	for _, rs := range strings.Split(v, ",") {
		__antithesis_instrumentation__.Notify(35399)
		var thisRange vrange
		if strings.Contains(rs, "-") {
			__antithesis_instrumentation__.Notify(35401)
			ab := strings.SplitN(rs, "-", 2)
			a, err := strconv.Atoi(ab[0])
			if err != nil {
				__antithesis_instrumentation__.Notify(35405)
				return err
			} else {
				__antithesis_instrumentation__.Notify(35406)
			}
			__antithesis_instrumentation__.Notify(35402)
			b, err := strconv.Atoi(ab[1])
			if err != nil {
				__antithesis_instrumentation__.Notify(35407)
				return err
			} else {
				__antithesis_instrumentation__.Notify(35408)
			}
			__antithesis_instrumentation__.Notify(35403)
			if b < a {
				__antithesis_instrumentation__.Notify(35409)
				return errors.New("invalid range")
			} else {
				__antithesis_instrumentation__.Notify(35410)
			}
			__antithesis_instrumentation__.Notify(35404)
			thisRange = vrange{a, b}
		} else {
			__antithesis_instrumentation__.Notify(35411)
			a, err := strconv.Atoi(rs)
			if err != nil {
				__antithesis_instrumentation__.Notify(35413)
				return err
			} else {
				__antithesis_instrumentation__.Notify(35414)
			}
			__antithesis_instrumentation__.Notify(35412)
			thisRange = vrange{a, a}
		}
		__antithesis_instrumentation__.Notify(35400)
		r.ranges = append(r.ranges, thisRange)
	}
	__antithesis_instrumentation__.Notify(35398)
	return nil
}

func (r *rangeSelection) items() map[int]struct{} {
	__antithesis_instrumentation__.Notify(35415)
	s := map[int]struct{}{}
	for _, vr := range r.ranges {
		__antithesis_instrumentation__.Notify(35417)
		for i := vr.a; i <= vr.b; i++ {
			__antithesis_instrumentation__.Notify(35418)
			s[i] = struct{}{}
		}
	}
	__antithesis_instrumentation__.Notify(35416)
	return s
}

type fileSelection struct {
	includePatterns []string
	excludePatterns []string
	startTimestamp  timestampValue
	endTimestamp    timestampValue
}

func (fs *fileSelection) validate() error {
	__antithesis_instrumentation__.Notify(35419)
	for _, p := range append(fs.includePatterns, fs.excludePatterns...) {
		__antithesis_instrumentation__.Notify(35421)
		if _, err := filepath.Match(p, ""); err != nil {
			__antithesis_instrumentation__.Notify(35422)
			return err
		} else {
			__antithesis_instrumentation__.Notify(35423)
		}
	}
	__antithesis_instrumentation__.Notify(35420)
	return nil
}

func (fs *fileSelection) retrievalPatterns() []string {
	__antithesis_instrumentation__.Notify(35424)
	if len(fs.includePatterns) == 0 {
		__antithesis_instrumentation__.Notify(35426)

		return []string{"*"}
	} else {
		__antithesis_instrumentation__.Notify(35427)
	}
	__antithesis_instrumentation__.Notify(35425)
	return fs.includePatterns
}

func (fs *fileSelection) isIncluded(filename string, ctime, mtime time.Time) bool {
	__antithesis_instrumentation__.Notify(35428)

	included := false
	for _, p := range fs.retrievalPatterns() {
		__antithesis_instrumentation__.Notify(35435)
		if matched, _ := filepath.Match(p, filename); matched {
			__antithesis_instrumentation__.Notify(35436)
			included = true
			break
		} else {
			__antithesis_instrumentation__.Notify(35437)
		}
	}
	__antithesis_instrumentation__.Notify(35429)
	if !included {
		__antithesis_instrumentation__.Notify(35438)
		return false
	} else {
		__antithesis_instrumentation__.Notify(35439)
	}
	__antithesis_instrumentation__.Notify(35430)

	for _, p := range fs.excludePatterns {
		__antithesis_instrumentation__.Notify(35440)
		if matched, _ := filepath.Match(p, filename); matched {
			__antithesis_instrumentation__.Notify(35441)
			included = false
			break
		} else {
			__antithesis_instrumentation__.Notify(35442)
		}
	}
	__antithesis_instrumentation__.Notify(35431)
	if !included {
		__antithesis_instrumentation__.Notify(35443)
		return false
	} else {
		__antithesis_instrumentation__.Notify(35444)
	}
	__antithesis_instrumentation__.Notify(35432)

	if mtime.Before(time.Time(fs.startTimestamp)) {
		__antithesis_instrumentation__.Notify(35445)
		return false
	} else {
		__antithesis_instrumentation__.Notify(35446)
	}
	__antithesis_instrumentation__.Notify(35433)

	if (*time.Time)(&fs.endTimestamp).Before(ctime) {
		__antithesis_instrumentation__.Notify(35447)
		return false
	} else {
		__antithesis_instrumentation__.Notify(35448)
	}
	__antithesis_instrumentation__.Notify(35434)
	return true
}

var zipReportingMu syncutil.Mutex

type zipReporter struct {
	prefix string

	flowing bool

	newline bool

	inItem bool
}

func (zc *zipContext) newZipReporter(format string, args ...interface{}) *zipReporter {
	__antithesis_instrumentation__.Notify(35449)
	return &zipReporter{
		flowing: zc.concurrency == 1,
		prefix:  "[" + fmt.Sprintf(format, args...) + "]",
		newline: true,
		inItem:  false,
	}
}

func (z *zipReporter) withPrefix(format string, args ...interface{}) *zipReporter {
	__antithesis_instrumentation__.Notify(35450)
	zipReportingMu.Lock()
	defer zipReportingMu.Unlock()

	if z.inItem {
		__antithesis_instrumentation__.Notify(35452)
		panic(errors.AssertionFailedf("can't use withPrefix() under start()"))
	} else {
		__antithesis_instrumentation__.Notify(35453)
	}
	__antithesis_instrumentation__.Notify(35451)

	z.completeprevLocked()
	return &zipReporter{
		prefix:  z.prefix + " [" + fmt.Sprintf(format, args...) + "]",
		flowing: z.flowing,
		newline: z.newline,
	}
}

func (z *zipReporter) start(format string, args ...interface{}) *zipReporter {
	__antithesis_instrumentation__.Notify(35454)
	zipReportingMu.Lock()
	defer zipReportingMu.Unlock()

	if z.inItem {
		__antithesis_instrumentation__.Notify(35456)
		panic(errors.AssertionFailedf("can't use start() under start()"))
	} else {
		__antithesis_instrumentation__.Notify(35457)
	}
	__antithesis_instrumentation__.Notify(35455)

	z.completeprevLocked()
	msg := z.prefix + " " + fmt.Sprintf(format, args...)
	nz := &zipReporter{
		prefix:  msg,
		flowing: z.flowing,
		inItem:  true,
	}
	fmt.Print(msg + "...")
	nz.flowLocked()
	return nz
}

func (z *zipReporter) flowLocked() {
	__antithesis_instrumentation__.Notify(35458)
	if !z.flowing {
		__antithesis_instrumentation__.Notify(35459)

		fmt.Println()
	} else {
		__antithesis_instrumentation__.Notify(35460)
		z.newline = false
	}
}

func (z *zipReporter) resumeLocked() {
	__antithesis_instrumentation__.Notify(35461)
	if !z.flowing || func() bool {
		__antithesis_instrumentation__.Notify(35463)
		return z.newline == true
	}() == true {
		__antithesis_instrumentation__.Notify(35464)
		fmt.Print(z.prefix + ":")
	} else {
		__antithesis_instrumentation__.Notify(35465)
	}
	__antithesis_instrumentation__.Notify(35462)
	if z.flowing {
		__antithesis_instrumentation__.Notify(35466)
		z.newline = false
	} else {
		__antithesis_instrumentation__.Notify(35467)
	}
}

func (z *zipReporter) completeprevLocked() {
	__antithesis_instrumentation__.Notify(35468)
	if z.flowing && func() bool {
		__antithesis_instrumentation__.Notify(35469)
		return !z.newline == true
	}() == true {
		__antithesis_instrumentation__.Notify(35470)
		fmt.Println()
		z.newline = true
	} else {
		__antithesis_instrumentation__.Notify(35471)
	}
}

func (z *zipReporter) endlLocked() {
	__antithesis_instrumentation__.Notify(35472)
	fmt.Println()
	if z.flowing {
		__antithesis_instrumentation__.Notify(35473)
		z.newline = true
	} else {
		__antithesis_instrumentation__.Notify(35474)
	}
}

func (z *zipReporter) info(format string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(35475)
	zipReportingMu.Lock()
	defer zipReportingMu.Unlock()

	z.completeprevLocked()
	fmt.Print(z.prefix)
	fmt.Print(" ")
	fmt.Printf(format, args...)
	z.endlLocked()
}

func (z *zipReporter) progress(format string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(35476)
	zipReportingMu.Lock()
	defer zipReportingMu.Unlock()

	if !z.inItem {
		__antithesis_instrumentation__.Notify(35478)
		panic(errors.AssertionFailedf("can't use progress() without start()"))
	} else {
		__antithesis_instrumentation__.Notify(35479)
	}
	__antithesis_instrumentation__.Notify(35477)

	z.resumeLocked()
	fmt.Print(" ")
	fmt.Printf(format, args...)
	fmt.Print("...")
	z.flowLocked()
}

func (z *zipReporter) shout(format string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(35480)
	zipReportingMu.Lock()
	defer zipReportingMu.Unlock()

	z.completeprevLocked()
	fmt.Print(z.prefix + ": ")
	fmt.Printf(format, args...)
	z.endlLocked()
}

func (z *zipReporter) done() {
	__antithesis_instrumentation__.Notify(35481)
	zipReportingMu.Lock()
	defer zipReportingMu.Unlock()

	if !z.inItem {
		__antithesis_instrumentation__.Notify(35483)
		panic(errors.AssertionFailedf("can't use done() without start()"))
	} else {
		__antithesis_instrumentation__.Notify(35484)
	}
	__antithesis_instrumentation__.Notify(35482)
	z.resumeLocked()
	fmt.Print(" done")
	z.endlLocked()
	z.inItem = false
}

func (z *zipReporter) fail(err error) error {
	__antithesis_instrumentation__.Notify(35485)
	zipReportingMu.Lock()
	defer zipReportingMu.Unlock()

	if !z.inItem {
		__antithesis_instrumentation__.Notify(35487)
		panic(errors.AssertionFailedf("can't use fail() without start()"))
	} else {
		__antithesis_instrumentation__.Notify(35488)
	}
	__antithesis_instrumentation__.Notify(35486)

	z.resumeLocked()
	fmt.Print(" error:", err)
	z.endlLocked()
	z.inItem = false
	return err
}

func (z *zipReporter) result(err error) error {
	__antithesis_instrumentation__.Notify(35489)
	if err == nil {
		__antithesis_instrumentation__.Notify(35491)
		z.done()
		return nil
	} else {
		__antithesis_instrumentation__.Notify(35492)
	}
	__antithesis_instrumentation__.Notify(35490)
	return z.fail(err)
}

type timestampValue time.Time

func (t *timestampValue) Type() string {
	__antithesis_instrumentation__.Notify(35493)
	return "YYYY-MM-DD [HH:MM[:SS]]"
}

func (t *timestampValue) String() string {
	__antithesis_instrumentation__.Notify(35494)
	return (*time.Time)(t).Format("2006-01-02 15:04:05")
}

func (t *timestampValue) Set(v string) error {
	__antithesis_instrumentation__.Notify(35495)
	v = strings.TrimSpace(v)
	var tm time.Time
	var err error
	if len(v) <= len("YYYY-MM-DD") {
		__antithesis_instrumentation__.Notify(35498)
		tm, err = time.ParseInLocation("2006-01-02", v, time.UTC)
	} else {
		__antithesis_instrumentation__.Notify(35499)
		if len(v) <= len("YYYY-MM-DD HH:MM") {
			__antithesis_instrumentation__.Notify(35500)
			tm, err = time.ParseInLocation("2006-01-02 15:04", v, time.UTC)
		} else {
			__antithesis_instrumentation__.Notify(35501)
			tm, err = time.ParseInLocation("2006-01-02 15:04:05", v, time.UTC)
		}
	}
	__antithesis_instrumentation__.Notify(35496)
	if err != nil {
		__antithesis_instrumentation__.Notify(35502)
		return err
	} else {
		__antithesis_instrumentation__.Notify(35503)
	}
	__antithesis_instrumentation__.Notify(35497)
	*t = timestampValue(tm)
	return nil
}
