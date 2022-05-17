package debug

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/channel"
	"github.com/cockroachdb/cockroach/pkg/util/log/logpb"
	"github.com/cockroachdb/cockroach/pkg/util/log/severity"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

type regexpAsString struct {
	re *regexp.Regexp
}

func (r regexpAsString) String() string {
	__antithesis_instrumentation__.Notify(190237)
	if r.re == nil {
		__antithesis_instrumentation__.Notify(190239)
		return ".*"
	} else {
		__antithesis_instrumentation__.Notify(190240)
	}
	__antithesis_instrumentation__.Notify(190238)
	return r.re.String()
}

func (r *regexpAsString) UnmarshalJSON(data []byte) error {
	__antithesis_instrumentation__.Notify(190241)
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		__antithesis_instrumentation__.Notify(190243)
		return err
	} else {
		__antithesis_instrumentation__.Notify(190244)
	}
	__antithesis_instrumentation__.Notify(190242)
	var err error
	(*r).re, err = regexp.Compile(s)
	return err
}

type intAsString int

func (i *intAsString) UnmarshalJSON(data []byte) error {
	__antithesis_instrumentation__.Notify(190245)
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		__antithesis_instrumentation__.Notify(190247)
		return err
	} else {
		__antithesis_instrumentation__.Notify(190248)
	}
	__antithesis_instrumentation__.Notify(190246)
	var err error
	*(*int)(i), err = strconv.Atoi(s)
	return err
}

type durationAsString time.Duration

func (d *durationAsString) UnmarshalJSON(data []byte) error {
	__antithesis_instrumentation__.Notify(190249)
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		__antithesis_instrumentation__.Notify(190251)
		return err
	} else {
		__antithesis_instrumentation__.Notify(190252)
	}
	__antithesis_instrumentation__.Notify(190250)
	var err error
	*(*time.Duration)(d), err = time.ParseDuration(s)
	return err
}

func (d durationAsString) String() string {
	__antithesis_instrumentation__.Notify(190253)
	return time.Duration(d).String()
}

const (
	logSpyDefaultDuration = durationAsString(5 * time.Second)
	logSpyDefaultCount    = 1000
	logSpyChanCap         = 4096
)

type logSpyOptions struct {
	Count          intAsString
	Grep           regexpAsString
	Flatten        intAsString
	vmoduleOptions `json:",inline"`
}

func logSpyOptionsFromValues(values url.Values) (logSpyOptions, error) {
	__antithesis_instrumentation__.Notify(190254)
	rawValues := map[string]string{}
	for k, vals := range values {
		__antithesis_instrumentation__.Notify(190259)
		if len(vals) > 0 {
			__antithesis_instrumentation__.Notify(190260)
			rawValues[k] = vals[0]
		} else {
			__antithesis_instrumentation__.Notify(190261)
		}
	}
	__antithesis_instrumentation__.Notify(190255)
	data, err := json.Marshal(rawValues)
	if err != nil {
		__antithesis_instrumentation__.Notify(190262)
		return logSpyOptions{}, err
	} else {
		__antithesis_instrumentation__.Notify(190263)
	}
	__antithesis_instrumentation__.Notify(190256)
	var opts logSpyOptions
	if err := json.Unmarshal(data, &opts); err != nil {
		__antithesis_instrumentation__.Notify(190264)
		return logSpyOptions{}, err
	} else {
		__antithesis_instrumentation__.Notify(190265)
	}
	__antithesis_instrumentation__.Notify(190257)
	if opts.Count == 0 {
		__antithesis_instrumentation__.Notify(190266)
		opts.Count = logSpyDefaultCount
	} else {
		__antithesis_instrumentation__.Notify(190267)
	}
	__antithesis_instrumentation__.Notify(190258)
	opts.vmoduleOptions.setDefaults(values)
	return opts, nil
}

type logSpy struct {
	vsrv         *vmoduleServer
	setIntercept func(ctx context.Context, f log.Interceptor) func()
}

func (spy *logSpy) handleDebugLogSpy(w http.ResponseWriter, r *http.Request) {
	__antithesis_instrumentation__.Notify(190268)
	opts, err := logSpyOptionsFromValues(r.URL.Query())
	if err != nil {
		__antithesis_instrumentation__.Notify(190270)
		http.Error(w, "while parsing options: "+err.Error(), http.StatusInternalServerError)
		return
	} else {
		__antithesis_instrumentation__.Notify(190271)
	}
	__antithesis_instrumentation__.Notify(190269)

	w.Header().Add("Content-type", "text/plain; charset=UTF-8")
	ctx := r.Context()
	if err := spy.run(ctx, w, opts); err != nil {
		__antithesis_instrumentation__.Notify(190272)

		log.Infof(ctx, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(190273)
	}
}

func (spy *logSpy) run(ctx context.Context, w io.Writer, opts logSpyOptions) (err error) {
	__antithesis_instrumentation__.Notify(190274)
	ctx, cancel := context.WithTimeout(ctx, time.Duration(opts.Duration))
	defer cancel()

	interceptor := newLogSpyInterceptor(opts)

	defer func() {
		__antithesis_instrumentation__.Notify(190277)
		if dropped := atomic.LoadInt32(&interceptor.countDropped); dropped > 0 {
			__antithesis_instrumentation__.Notify(190278)
			entry := log.MakeLegacyEntry(
				ctx, severity.WARNING, channel.DEV,
				0, true,
				"%d messages were dropped", redact.Safe(dropped))
			err = errors.CombineErrors(err, interceptor.outputEntry(w, entry))
		} else {
			__antithesis_instrumentation__.Notify(190279)
		}
	}()
	__antithesis_instrumentation__.Notify(190275)

	cleanup := spy.setIntercept(ctx, interceptor)
	defer cleanup()

	log.Infof(ctx, "intercepting logs with options: %+v", opts)

	prevVModule := log.GetVModule()
	if opts.hasVModule {
		__antithesis_instrumentation__.Notify(190280)
		if err := spy.vsrv.lockVModule(ctx); err != nil {
			__antithesis_instrumentation__.Notify(190283)
			fmt.Fprintf(w, "error: %v", err)
			return err
		} else {
			__antithesis_instrumentation__.Notify(190284)
		}
		__antithesis_instrumentation__.Notify(190281)

		log.Infof(ctx, "previous vmodule configuration: %s", prevVModule)

		if err := log.SetVModule(opts.VModule); err != nil {
			__antithesis_instrumentation__.Notify(190285)
			fmt.Fprintf(w, "error: %v", err)
			return err
		} else {
			__antithesis_instrumentation__.Notify(190286)
		}
		__antithesis_instrumentation__.Notify(190282)
		log.Infof(ctx, "new vmodule configuration (previous will be restored when logspy session completes): %s", redact.SafeString(opts.VModule))
		defer func() {
			__antithesis_instrumentation__.Notify(190287)

			err := log.SetVModule(prevVModule)

			log.Infof(ctx, "restoring vmodule configuration (%q): %v", redact.SafeString(prevVModule), err)
			spy.vsrv.unlockVModule(ctx)
		}()
	} else {
		__antithesis_instrumentation__.Notify(190288)
		log.Infof(ctx, "current vmodule setting: %s", redact.SafeString(prevVModule))
	}
	__antithesis_instrumentation__.Notify(190276)

	const flushInterval = time.Second
	var flushTimer timeutil.Timer
	defer flushTimer.Stop()
	flushTimer.Reset(flushInterval)

	numReportedEntries := 0
	for {
		__antithesis_instrumentation__.Notify(190289)
		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(190290)
			err := ctx.Err()
			if errors.Is(err, context.DeadlineExceeded) {
				__antithesis_instrumentation__.Notify(190295)

				return nil
			} else {
				__antithesis_instrumentation__.Notify(190296)
			}
			__antithesis_instrumentation__.Notify(190291)
			return err

		case jsonEntry := <-interceptor.jsonEntries:
			__antithesis_instrumentation__.Notify(190292)
			if err := interceptor.outputJSONEntry(w, jsonEntry); err != nil {
				__antithesis_instrumentation__.Notify(190297)
				return errors.Wrapf(err, "while writing entry %s", jsonEntry)
			} else {
				__antithesis_instrumentation__.Notify(190298)
			}
			__antithesis_instrumentation__.Notify(190293)
			numReportedEntries++
			if numReportedEntries >= int(opts.Count) {
				__antithesis_instrumentation__.Notify(190299)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(190300)
			}

		case <-flushTimer.C:
			__antithesis_instrumentation__.Notify(190294)
			flushTimer.Read = true
			flushTimer.Reset(flushInterval)
			if flusher, ok := w.(http.Flusher); ok {
				__antithesis_instrumentation__.Notify(190301)
				flusher.Flush()
			} else {
				__antithesis_instrumentation__.Notify(190302)
			}
		}
	}
}

type logSpyInterceptor struct {
	opts         logSpyOptions
	countDropped int32
	jsonEntries  chan []byte
}

func newLogSpyInterceptor(opts logSpyOptions) *logSpyInterceptor {
	__antithesis_instrumentation__.Notify(190303)
	return &logSpyInterceptor{
		opts:        opts,
		jsonEntries: make(chan []byte, logSpyChanCap),
	}
}

func (i *logSpyInterceptor) Intercept(jsonEntry []byte) {
	__antithesis_instrumentation__.Notify(190304)
	if re := i.opts.Grep.re; re != nil {
		__antithesis_instrumentation__.Notify(190306)
		switch {
		case re.Match(jsonEntry):
			__antithesis_instrumentation__.Notify(190307)
		default:
			__antithesis_instrumentation__.Notify(190308)
			return
		}
	} else {
		__antithesis_instrumentation__.Notify(190309)
	}
	__antithesis_instrumentation__.Notify(190305)

	jsonCopy := make([]byte, len(jsonEntry))
	copy(jsonCopy, jsonEntry)

	select {
	case i.jsonEntries <- jsonCopy:
		__antithesis_instrumentation__.Notify(190310)
	default:
		__antithesis_instrumentation__.Notify(190311)

		atomic.AddInt32(&i.countDropped, 1)
	}
}

func (i *logSpyInterceptor) outputEntry(w io.Writer, entry logpb.Entry) error {
	__antithesis_instrumentation__.Notify(190312)
	if i.opts.Flatten > 0 {
		__antithesis_instrumentation__.Notify(190314)
		return log.FormatLegacyEntry(entry, w)
	} else {
		__antithesis_instrumentation__.Notify(190315)
	}
	__antithesis_instrumentation__.Notify(190313)
	j, _ := json.Marshal(entry)
	return i.outputJSONEntry(w, j)
}

func (i *logSpyInterceptor) outputJSONEntry(w io.Writer, jsonEntry []byte) error {
	__antithesis_instrumentation__.Notify(190316)
	if i.opts.Flatten == 0 {
		__antithesis_instrumentation__.Notify(190319)
		_, err1 := w.Write(jsonEntry)
		_, err2 := w.Write([]byte("\n"))
		return errors.CombineErrors(err1, err2)
	} else {
		__antithesis_instrumentation__.Notify(190320)
	}
	__antithesis_instrumentation__.Notify(190317)
	var legacyEntry logpb.Entry
	if err := json.Unmarshal(jsonEntry, &legacyEntry); err != nil {
		__antithesis_instrumentation__.Notify(190321)
		return errors.NewAssertionErrorWithWrappedErrf(err, "interceptor API does not seem to provide valid Entry payloads")
	} else {
		__antithesis_instrumentation__.Notify(190322)
	}
	__antithesis_instrumentation__.Notify(190318)
	return i.outputEntry(w, legacyEntry)
}
