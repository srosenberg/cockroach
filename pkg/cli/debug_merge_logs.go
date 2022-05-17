package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"bytes"
	"container/heap"
	"context"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/logpb"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/ttycolor"
	"golang.org/x/sync/errgroup"
)

type logStream interface {
	fileInfo() *fileInfo
	peek() (logpb.Entry, bool)
	pop() (logpb.Entry, bool)
	error() error
}

func writeLogStream(
	s logStream, out io.Writer, filter *regexp.Regexp, keepRedactable bool, cp ttycolor.Profile,
) error {
	__antithesis_instrumentation__.Notify(31203)
	const chanSize = 1 << 16
	const maxWriteBufSize = 1 << 18

	type entryInfo struct {
		logpb.Entry
		*fileInfo
	}
	render := func(ei entryInfo, w io.Writer) (err error) {
		__antithesis_instrumentation__.Notify(31208)

		err = log.FormatLegacyEntryPrefix(ei.prefix, w, cp)
		if err != nil {
			__antithesis_instrumentation__.Notify(31210)
			return err
		} else {
			__antithesis_instrumentation__.Notify(31211)
		}
		__antithesis_instrumentation__.Notify(31209)
		return log.FormatLegacyEntryWithOptionalColors(ei.Entry, w, cp)
	}
	__antithesis_instrumentation__.Notify(31204)

	g, ctx := errgroup.WithContext(context.Background())
	entryChan := make(chan entryInfo, chanSize)
	writeChan := make(chan *bytes.Buffer)
	read := func() error {
		__antithesis_instrumentation__.Notify(31212)
		defer close(entryChan)
		for e, ok := s.peek(); ok; e, ok = s.peek() {
			__antithesis_instrumentation__.Notify(31214)
			select {
			case entryChan <- entryInfo{Entry: e, fileInfo: s.fileInfo()}:
				__antithesis_instrumentation__.Notify(31216)
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(31217)
				return nil
			}
			__antithesis_instrumentation__.Notify(31215)
			s.pop()
		}
		__antithesis_instrumentation__.Notify(31213)
		return s.error()
	}
	__antithesis_instrumentation__.Notify(31205)
	bufferWrites := func() error {
		__antithesis_instrumentation__.Notify(31218)
		defer close(writeChan)
		writing, pending := &bytes.Buffer{}, &bytes.Buffer{}
		for {
			__antithesis_instrumentation__.Notify(31219)
			send, recv := writeChan, entryChan
			var scratch []byte
			if pending.Len() == 0 {
				__antithesis_instrumentation__.Notify(31221)
				send = nil
				if recv == nil {
					__antithesis_instrumentation__.Notify(31222)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(31223)
				}
			} else {
				__antithesis_instrumentation__.Notify(31224)
				if pending.Len() > maxWriteBufSize {
					__antithesis_instrumentation__.Notify(31225)
					recv = nil
				} else {
					__antithesis_instrumentation__.Notify(31226)
				}
			}
			__antithesis_instrumentation__.Notify(31220)
			select {
			case ei, open := <-recv:
				__antithesis_instrumentation__.Notify(31227)
				if !open {
					__antithesis_instrumentation__.Notify(31232)
					entryChan = nil
					break
				} else {
					__antithesis_instrumentation__.Notify(31233)
				}
				__antithesis_instrumentation__.Notify(31228)
				startLen := pending.Len()
				if err := render(ei, pending); err != nil {
					__antithesis_instrumentation__.Notify(31234)
					return err
				} else {
					__antithesis_instrumentation__.Notify(31235)
				}
				__antithesis_instrumentation__.Notify(31229)
				if filter != nil {
					__antithesis_instrumentation__.Notify(31236)
					matches := filter.FindSubmatch(pending.Bytes()[startLen:])
					if matches == nil {
						__antithesis_instrumentation__.Notify(31237)

						pending.Truncate(startLen)
					} else {
						__antithesis_instrumentation__.Notify(31238)
						if len(matches) > 1 {
							__antithesis_instrumentation__.Notify(31239)

							scratch = scratch[:0]
							for _, b := range matches[1:] {
								__antithesis_instrumentation__.Notify(31241)
								scratch = append(scratch, b...)
							}
							__antithesis_instrumentation__.Notify(31240)
							pending.Truncate(startLen)
							_, _ = pending.Write(scratch)
							pending.WriteByte('\n')
						} else {
							__antithesis_instrumentation__.Notify(31242)
						}
					}
				} else {
					__antithesis_instrumentation__.Notify(31243)
				}
			case send <- pending:
				__antithesis_instrumentation__.Notify(31230)
				writing.Reset()
				pending, writing = writing, pending
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(31231)
				return nil
			}
		}
	}
	__antithesis_instrumentation__.Notify(31206)
	write := func() error {
		__antithesis_instrumentation__.Notify(31244)
		for buf := range writeChan {
			__antithesis_instrumentation__.Notify(31246)
			if _, err := out.Write(buf.Bytes()); err != nil {
				__antithesis_instrumentation__.Notify(31247)
				return err
			} else {
				__antithesis_instrumentation__.Notify(31248)
			}
		}
		__antithesis_instrumentation__.Notify(31245)
		return nil
	}
	__antithesis_instrumentation__.Notify(31207)
	g.Go(read)
	g.Go(bufferWrites)
	g.Go(write)
	return g.Wait()
}

type mergedStream []logStream

func newMergedStreamFromPatterns(
	ctx context.Context,
	patterns []string,
	filePattern, programFilter *regexp.Regexp,
	from, to time.Time,
	editMode log.EditSensitiveData,
	format string,
	prefixer filePrefixer,
) (logStream, error) {
	__antithesis_instrumentation__.Notify(31249)
	paths, err := expandPatterns(patterns)
	if err != nil {
		__antithesis_instrumentation__.Notify(31252)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(31253)
	}
	__antithesis_instrumentation__.Notify(31250)
	files, err := findLogFiles(paths, filePattern, programFilter,
		groupIndex(filePattern, "program"), to)
	if err != nil {
		__antithesis_instrumentation__.Notify(31254)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(31255)
	}
	__antithesis_instrumentation__.Notify(31251)

	prefixer.PopulatePrefixes(files)
	return newMergedStream(ctx, files, from, to, editMode, format)
}

func groupIndex(re *regexp.Regexp, groupName string) int {
	__antithesis_instrumentation__.Notify(31256)
	for i, n := range re.SubexpNames() {
		__antithesis_instrumentation__.Notify(31258)
		if n == groupName {
			__antithesis_instrumentation__.Notify(31259)
			return i
		} else {
			__antithesis_instrumentation__.Notify(31260)
		}
	}
	__antithesis_instrumentation__.Notify(31257)
	return -1
}

func newMergedStream(
	ctx context.Context,
	files []fileInfo,
	from, to time.Time,
	editMode log.EditSensitiveData,
	format string,
) (*mergedStream, error) {
	__antithesis_instrumentation__.Notify(31261)

	const maxConcurrentFiles = 256
	sem := make(chan struct{}, maxConcurrentFiles)
	res := make(mergedStream, len(files))
	g, _ := errgroup.WithContext(ctx)
	createFileStream := func(i int) func() error {
		__antithesis_instrumentation__.Notify(31266)
		return func() error {
			__antithesis_instrumentation__.Notify(31267)
			sem <- struct{}{}
			defer func() { __antithesis_instrumentation__.Notify(31270); <-sem }()
			__antithesis_instrumentation__.Notify(31268)
			s, err := newFileLogStream(files[i], from, to, editMode, format)
			if s != nil {
				__antithesis_instrumentation__.Notify(31271)
				res[i] = s
			} else {
				__antithesis_instrumentation__.Notify(31272)
			}
			__antithesis_instrumentation__.Notify(31269)
			return err
		}
	}
	__antithesis_instrumentation__.Notify(31262)
	for i := range files {
		__antithesis_instrumentation__.Notify(31273)
		g.Go(createFileStream(i))
	}
	__antithesis_instrumentation__.Notify(31263)
	if err := g.Wait(); err != nil {
		__antithesis_instrumentation__.Notify(31274)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(31275)
	}
	__antithesis_instrumentation__.Notify(31264)
	g, ctx = errgroup.WithContext(ctx)
	filtered := res[:0]
	for _, s := range res {
		__antithesis_instrumentation__.Notify(31276)
		if s != nil {
			__antithesis_instrumentation__.Notify(31277)
			s = newBufferedLogStream(ctx, g, s)
			filtered = append(filtered, s)
		} else {
			__antithesis_instrumentation__.Notify(31278)
		}
	}
	__antithesis_instrumentation__.Notify(31265)
	res = filtered
	heap.Init(&res)
	return &res, nil
}

func (l mergedStream) Len() int { __antithesis_instrumentation__.Notify(31279); return len(l) }
func (l mergedStream) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(31280)
	l[i], l[j] = l[j], l[i]
}
func (l mergedStream) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(31281)
	ie, iok := l[i].peek()
	je, jok := l[j].peek()
	if iok && func() bool {
		__antithesis_instrumentation__.Notify(31283)
		return jok == true
	}() == true {
		__antithesis_instrumentation__.Notify(31284)
		return ie.Time < je.Time
	} else {
		__antithesis_instrumentation__.Notify(31285)
	}
	__antithesis_instrumentation__.Notify(31282)
	return !iok && func() bool {
		__antithesis_instrumentation__.Notify(31286)
		return jok == true
	}() == true
}

func (l *mergedStream) Push(s interface{}) {
	__antithesis_instrumentation__.Notify(31287)
	*l = append(*l, s.(logStream))
}

func (l *mergedStream) Pop() (v interface{}) {
	__antithesis_instrumentation__.Notify(31288)
	n := len(*l) - 1
	v = (*l)[n]
	*l = (*l)[:n]
	return
}

func (l *mergedStream) peek() (logpb.Entry, bool) {
	__antithesis_instrumentation__.Notify(31289)
	if len(*l) == 0 {
		__antithesis_instrumentation__.Notify(31291)
		return logpb.Entry{}, false
	} else {
		__antithesis_instrumentation__.Notify(31292)
	}
	__antithesis_instrumentation__.Notify(31290)
	return (*l)[0].peek()
}

func (l *mergedStream) pop() (logpb.Entry, bool) {
	__antithesis_instrumentation__.Notify(31293)
	e, ok := l.peek()
	if !ok {
		__antithesis_instrumentation__.Notify(31296)
		return logpb.Entry{}, false
	} else {
		__antithesis_instrumentation__.Notify(31297)
	}
	__antithesis_instrumentation__.Notify(31294)
	s := (*l)[0]
	s.pop()
	if _, stillOk := s.peek(); stillOk {
		__antithesis_instrumentation__.Notify(31298)
		heap.Push(l, heap.Pop(l))
	} else {
		__antithesis_instrumentation__.Notify(31299)
		if err := s.error(); err != nil && func() bool {
			__antithesis_instrumentation__.Notify(31300)
			return err != io.EOF == true
		}() == true {
			__antithesis_instrumentation__.Notify(31301)
			return logpb.Entry{}, false
		} else {
			__antithesis_instrumentation__.Notify(31302)
			heap.Pop(l)
		}
	}
	__antithesis_instrumentation__.Notify(31295)
	return e, true
}

func (l *mergedStream) fileInfo() *fileInfo {
	__antithesis_instrumentation__.Notify(31303)
	if len(*l) == 0 {
		__antithesis_instrumentation__.Notify(31305)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(31306)
	}
	__antithesis_instrumentation__.Notify(31304)
	return (*l)[0].fileInfo()
}

func (l *mergedStream) error() error {
	__antithesis_instrumentation__.Notify(31307)
	if len(*l) == 0 {
		__antithesis_instrumentation__.Notify(31309)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(31310)
	}
	__antithesis_instrumentation__.Notify(31308)
	return (*l)[0].error()
}

func expandPatterns(patterns []string) ([]string, error) {
	__antithesis_instrumentation__.Notify(31311)
	var paths []string
	for _, p := range patterns {
		__antithesis_instrumentation__.Notify(31313)
		matches, err := filepath.Glob(p)
		if err != nil {
			__antithesis_instrumentation__.Notify(31315)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(31316)
		}
		__antithesis_instrumentation__.Notify(31314)
		paths = append(paths, matches...)
	}
	__antithesis_instrumentation__.Notify(31312)
	return removeDuplicates(paths), nil
}

func removeDuplicates(strings []string) (filtered []string) {
	__antithesis_instrumentation__.Notify(31317)
	filtered = strings[:0]
	prev := ""
	for _, s := range strings {
		__antithesis_instrumentation__.Notify(31319)
		if s == prev {
			__antithesis_instrumentation__.Notify(31321)
			continue
		} else {
			__antithesis_instrumentation__.Notify(31322)
		}
		__antithesis_instrumentation__.Notify(31320)
		filtered = append(filtered, s)
		prev = s
	}
	__antithesis_instrumentation__.Notify(31318)
	return filtered
}

func getLogFileInfo(path string, filePattern *regexp.Regexp) (fileInfo, bool) {
	__antithesis_instrumentation__.Notify(31323)
	if matches := filePattern.FindStringSubmatchIndex(path); matches != nil {
		__antithesis_instrumentation__.Notify(31325)
		return fileInfo{path: path, matches: matches, pattern: filePattern}, true
	} else {
		__antithesis_instrumentation__.Notify(31326)
	}
	__antithesis_instrumentation__.Notify(31324)
	return fileInfo{}, false
}

type fileInfo struct {
	path    string
	prefix  []byte
	pattern *regexp.Regexp
	matches []int
}

func findLogFiles(
	paths []string, filePattern, programFilter *regexp.Regexp, programGroup int, to time.Time,
) ([]fileInfo, error) {
	__antithesis_instrumentation__.Notify(31327)
	if programGroup == 0 || func() bool {
		__antithesis_instrumentation__.Notify(31330)
		return programFilter == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(31331)
		programGroup = 0
	} else {
		__antithesis_instrumentation__.Notify(31332)
	}
	__antithesis_instrumentation__.Notify(31328)
	to = to.Truncate(time.Second)
	var files []fileInfo
	for _, p := range paths {
		__antithesis_instrumentation__.Notify(31333)

		if err := filepath.Walk(p, func(p string, info os.FileInfo, err error) error {
			__antithesis_instrumentation__.Notify(31334)
			if err != nil {
				__antithesis_instrumentation__.Notify(31339)
				return err
			} else {
				__antithesis_instrumentation__.Notify(31340)
			}
			__antithesis_instrumentation__.Notify(31335)
			if info.IsDir() {
				__antithesis_instrumentation__.Notify(31341)

				return nil
			} else {
				__antithesis_instrumentation__.Notify(31342)
			}
			__antithesis_instrumentation__.Notify(31336)

			fi, ok := getLogFileInfo(p, filePattern)
			if !ok {
				__antithesis_instrumentation__.Notify(31343)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(31344)
			}
			__antithesis_instrumentation__.Notify(31337)
			if programGroup > 0 {
				__antithesis_instrumentation__.Notify(31345)
				program := fi.path[fi.matches[2*programGroup]:fi.matches[2*programGroup+1]]
				if !programFilter.MatchString(program) {
					__antithesis_instrumentation__.Notify(31346)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(31347)
				}
			} else {
				__antithesis_instrumentation__.Notify(31348)
			}
			__antithesis_instrumentation__.Notify(31338)
			files = append(files, fi)
			return nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(31349)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(31350)
		}

	}
	__antithesis_instrumentation__.Notify(31329)
	return files, nil
}

func newBufferedLogStream(ctx context.Context, g *errgroup.Group, s logStream) logStream {
	__antithesis_instrumentation__.Notify(31351)
	bs := &bufferedLogStream{ctx: ctx, logStream: s, g: g}

	bs.e, bs.ok = s.peek()
	bs.read = true
	s.pop()
	return bs
}

type bufferedLogStream struct {
	logStream
	runOnce sync.Once
	e       logpb.Entry
	read    bool
	ok      bool
	c       chan logpb.Entry
	ctx     context.Context
	g       *errgroup.Group
}

func (bs *bufferedLogStream) run() {
	__antithesis_instrumentation__.Notify(31352)
	const readChanSize = 512
	bs.c = make(chan logpb.Entry, readChanSize)
	bs.g.Go(func() error {
		__antithesis_instrumentation__.Notify(31353)
		defer close(bs.c)
		for {
			__antithesis_instrumentation__.Notify(31354)
			e, ok := bs.logStream.pop()
			if !ok {
				__antithesis_instrumentation__.Notify(31356)
				if err := bs.error(); err != io.EOF {
					__antithesis_instrumentation__.Notify(31358)
					return err
				} else {
					__antithesis_instrumentation__.Notify(31359)
				}
				__antithesis_instrumentation__.Notify(31357)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(31360)
			}
			__antithesis_instrumentation__.Notify(31355)
			select {
			case bs.c <- e:
				__antithesis_instrumentation__.Notify(31361)
			case <-bs.ctx.Done():
				__antithesis_instrumentation__.Notify(31362)
				return nil
			}
		}
	})
}

func (bs *bufferedLogStream) peek() (logpb.Entry, bool) {
	__antithesis_instrumentation__.Notify(31363)
	if bs.ok && func() bool {
		__antithesis_instrumentation__.Notify(31366)
		return !bs.read == true
	}() == true {
		__antithesis_instrumentation__.Notify(31367)
		if bs.c == nil {
			__antithesis_instrumentation__.Notify(31369)
			bs.runOnce.Do(bs.run)
		} else {
			__antithesis_instrumentation__.Notify(31370)
		}
		__antithesis_instrumentation__.Notify(31368)
		bs.e, bs.ok = <-bs.c
		bs.read = true
	} else {
		__antithesis_instrumentation__.Notify(31371)
	}
	__antithesis_instrumentation__.Notify(31364)
	if !bs.ok {
		__antithesis_instrumentation__.Notify(31372)
		return logpb.Entry{}, false
	} else {
		__antithesis_instrumentation__.Notify(31373)
	}
	__antithesis_instrumentation__.Notify(31365)
	return bs.e, true
}

func (bs *bufferedLogStream) pop() (logpb.Entry, bool) {
	__antithesis_instrumentation__.Notify(31374)
	e, ok := bs.peek()
	bs.read = false
	return e, ok
}

type fileLogStream struct {
	from, to time.Time
	prevTime int64
	fi       fileInfo
	f        *os.File
	d        log.EntryDecoder
	read     bool
	editMode log.EditSensitiveData
	format   string

	e   logpb.Entry
	err error
}

func newFileLogStream(
	fi fileInfo, from, to time.Time, editMode log.EditSensitiveData, format string,
) (logStream, error) {
	__antithesis_instrumentation__.Notify(31375)
	s := &fileLogStream{
		fi:       fi,
		from:     from,
		to:       to,
		editMode: editMode,
		format:   format,
	}
	if _, ok := s.peek(); !ok {
		__antithesis_instrumentation__.Notify(31377)
		if err := s.error(); err != io.EOF {
			__antithesis_instrumentation__.Notify(31379)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(31380)
		}
		__antithesis_instrumentation__.Notify(31378)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(31381)
	}
	__antithesis_instrumentation__.Notify(31376)
	s.close()
	return s, nil
}

func (s *fileLogStream) close() {
	__antithesis_instrumentation__.Notify(31382)
	s.f.Close()
	s.f = nil
	s.d = nil
}

func (s *fileLogStream) open() bool {
	__antithesis_instrumentation__.Notify(31383)
	const readBufSize = 1024
	if s.f, s.err = os.Open(s.fi.path); s.err != nil {
		__antithesis_instrumentation__.Notify(31388)
		return false
	} else {
		__antithesis_instrumentation__.Notify(31389)
	}
	__antithesis_instrumentation__.Notify(31384)
	if s.format == "" {
		__antithesis_instrumentation__.Notify(31390)
		if _, s.format, s.err = log.ReadFormatFromLogFile(s.f); s.err != nil {
			__antithesis_instrumentation__.Notify(31392)
			return false
		} else {
			__antithesis_instrumentation__.Notify(31393)
		}
		__antithesis_instrumentation__.Notify(31391)
		if _, s.err = s.f.Seek(0, io.SeekStart); s.err != nil {
			__antithesis_instrumentation__.Notify(31394)
			return false
		} else {
			__antithesis_instrumentation__.Notify(31395)
		}
	} else {
		__antithesis_instrumentation__.Notify(31396)
	}
	__antithesis_instrumentation__.Notify(31385)
	if s.err = seekToFirstAfterFrom(s.f, s.from, s.editMode, s.format); s.err != nil {
		__antithesis_instrumentation__.Notify(31397)
		return false
	} else {
		__antithesis_instrumentation__.Notify(31398)
	}
	__antithesis_instrumentation__.Notify(31386)
	if s.d, s.err = log.NewEntryDecoderWithFormat(bufio.NewReaderSize(s.f, readBufSize), s.editMode, s.format); s.err != nil {
		__antithesis_instrumentation__.Notify(31399)
		return false
	} else {
		__antithesis_instrumentation__.Notify(31400)
	}
	__antithesis_instrumentation__.Notify(31387)
	return true
}

func (s *fileLogStream) peek() (logpb.Entry, bool) {
	__antithesis_instrumentation__.Notify(31401)
	for !s.read && func() bool {
		__antithesis_instrumentation__.Notify(31403)
		return s.err == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(31404)
		justOpened := false
		if s.d == nil {
			__antithesis_instrumentation__.Notify(31409)
			if !s.open() {
				__antithesis_instrumentation__.Notify(31411)
				return logpb.Entry{}, false
			} else {
				__antithesis_instrumentation__.Notify(31412)
			}
			__antithesis_instrumentation__.Notify(31410)
			justOpened = true
		} else {
			__antithesis_instrumentation__.Notify(31413)
		}
		__antithesis_instrumentation__.Notify(31405)
		var e logpb.Entry
		if s.err = s.d.Decode(&e); s.err != nil {
			__antithesis_instrumentation__.Notify(31414)
			s.close()
			s.e = logpb.Entry{}
			break
		} else {
			__antithesis_instrumentation__.Notify(31415)
		}
		__antithesis_instrumentation__.Notify(31406)

		if justOpened && func() bool {
			__antithesis_instrumentation__.Notify(31416)
			return e == s.e == true
		}() == true {
			__antithesis_instrumentation__.Notify(31417)
			continue
		} else {
			__antithesis_instrumentation__.Notify(31418)
		}
		__antithesis_instrumentation__.Notify(31407)
		s.e = e
		if s.e.Time < s.prevTime {
			__antithesis_instrumentation__.Notify(31419)
			s.e.Time = s.prevTime
		} else {
			__antithesis_instrumentation__.Notify(31420)
			s.prevTime = s.e.Time
		}
		__antithesis_instrumentation__.Notify(31408)
		afterTo := !s.to.IsZero() && func() bool {
			__antithesis_instrumentation__.Notify(31421)
			return s.e.Time > s.to.UnixNano() == true
		}() == true
		if afterTo {
			__antithesis_instrumentation__.Notify(31422)
			s.close()
			s.e = logpb.Entry{}
			s.err = io.EOF
		} else {
			__antithesis_instrumentation__.Notify(31423)
			beforeFrom := !s.from.IsZero() && func() bool {
				__antithesis_instrumentation__.Notify(31424)
				return s.e.Time < s.from.UnixNano() == true
			}() == true
			s.read = !beforeFrom
		}
	}
	__antithesis_instrumentation__.Notify(31402)
	return s.e, s.err == nil
}

func (s *fileLogStream) pop() (e logpb.Entry, ok bool) {
	__antithesis_instrumentation__.Notify(31425)
	if e, ok = s.peek(); !ok {
		__antithesis_instrumentation__.Notify(31427)
		return
	} else {
		__antithesis_instrumentation__.Notify(31428)
	}
	__antithesis_instrumentation__.Notify(31426)
	s.read = false
	return e, ok
}

func (s *fileLogStream) fileInfo() *fileInfo {
	__antithesis_instrumentation__.Notify(31429)
	return &s.fi
}
func (s *fileLogStream) error() error { __antithesis_instrumentation__.Notify(31430); return s.err }

func seekToFirstAfterFrom(
	f *os.File, from time.Time, editMode log.EditSensitiveData, format string,
) (err error) {
	__antithesis_instrumentation__.Notify(31431)
	if from.IsZero() {
		__antithesis_instrumentation__.Notify(31439)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(31440)
	}
	__antithesis_instrumentation__.Notify(31432)
	fi, err := f.Stat()
	if err != nil {
		__antithesis_instrumentation__.Notify(31441)
		return err
	} else {
		__antithesis_instrumentation__.Notify(31442)
	}
	__antithesis_instrumentation__.Notify(31433)
	size := fi.Size()
	defer func() {
		__antithesis_instrumentation__.Notify(31443)
		if r := recover(); r != nil {
			__antithesis_instrumentation__.Notify(31444)
			err = r.(error)
		} else {
			__antithesis_instrumentation__.Notify(31445)
		}
	}()
	__antithesis_instrumentation__.Notify(31434)
	offset := sort.Search(int(size), func(i int) bool {
		__antithesis_instrumentation__.Notify(31446)
		if _, err := f.Seek(int64(i), io.SeekStart); err != nil {
			__antithesis_instrumentation__.Notify(31450)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(31451)
		}
		__antithesis_instrumentation__.Notify(31447)
		var e logpb.Entry
		d, err := log.NewEntryDecoderWithFormat(f, editMode, format)
		if err != nil {
			__antithesis_instrumentation__.Notify(31452)
			panic(errors.WithMessagef(err, "error while processing file %s", f.Name()))
		} else {
			__antithesis_instrumentation__.Notify(31453)
		}
		__antithesis_instrumentation__.Notify(31448)
		if err := d.Decode(&e); err != nil {
			__antithesis_instrumentation__.Notify(31454)
			if err == io.EOF {
				__antithesis_instrumentation__.Notify(31456)
				return true
			} else {
				__antithesis_instrumentation__.Notify(31457)
			}
			__antithesis_instrumentation__.Notify(31455)
			panic(errors.WithMessagef(err, "error while processing file %s", f.Name()))
		} else {
			__antithesis_instrumentation__.Notify(31458)
		}
		__antithesis_instrumentation__.Notify(31449)
		return e.Time >= from.UnixNano()
	})
	__antithesis_instrumentation__.Notify(31435)
	if _, err := f.Seek(int64(offset), io.SeekStart); err != nil {
		__antithesis_instrumentation__.Notify(31459)
		return err
	} else {
		__antithesis_instrumentation__.Notify(31460)
	}
	__antithesis_instrumentation__.Notify(31436)
	var e logpb.Entry
	d, err := log.NewEntryDecoderWithFormat(f, editMode, format)
	if err != nil {
		__antithesis_instrumentation__.Notify(31461)
		return err
	} else {
		__antithesis_instrumentation__.Notify(31462)
	}
	__antithesis_instrumentation__.Notify(31437)
	if err := d.Decode(&e); err != nil {
		__antithesis_instrumentation__.Notify(31463)
		return err
	} else {
		__antithesis_instrumentation__.Notify(31464)
	}
	__antithesis_instrumentation__.Notify(31438)
	_, err = f.Seek(int64(offset), io.SeekStart)
	return err
}

type forceColor int

const (
	forceColorAuto forceColor = iota
	forceColorOn
	forceColorOff
)

func (c *forceColor) Type() string {
	__antithesis_instrumentation__.Notify(31465)
	return "<true/false/auto>"
}

func (c *forceColor) String() string {
	__antithesis_instrumentation__.Notify(31466)
	switch *c {
	case forceColorAuto:
		__antithesis_instrumentation__.Notify(31467)
		return "auto"
	case forceColorOn:
		__antithesis_instrumentation__.Notify(31468)
		return "true"
	case forceColorOff:
		__antithesis_instrumentation__.Notify(31469)
		return "false"
	default:
		__antithesis_instrumentation__.Notify(31470)
		panic(errors.AssertionFailedf("unknown value: %v", int(*c)))
	}
}

func (c *forceColor) Set(v string) error {
	__antithesis_instrumentation__.Notify(31471)
	switch v {
	case "on", "true":
		__antithesis_instrumentation__.Notify(31473)
		*c = forceColorOn
	case "off", "false":
		__antithesis_instrumentation__.Notify(31474)
		*c = forceColorOff
	case "auto":
		__antithesis_instrumentation__.Notify(31475)
		*c = forceColorAuto
	default:
		__antithesis_instrumentation__.Notify(31476)
		return errors.Newf("unknown value: %v (supported: true/false/auto)", v)
	}
	__antithesis_instrumentation__.Notify(31472)
	return nil
}
