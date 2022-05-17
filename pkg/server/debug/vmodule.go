package debug

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

type vmoduleOptions struct {
	hasVModule bool
	VModule    string
	Duration   durationAsString
}

func loadVmoduleOptionsFromValues(values url.Values) (vmoduleOptions, error) {
	__antithesis_instrumentation__.Notify(190510)
	rawValues := map[string]string{}
	for k, vals := range values {
		__antithesis_instrumentation__.Notify(190514)
		if len(vals) > 0 {
			__antithesis_instrumentation__.Notify(190515)
			rawValues[k] = vals[0]
		} else {
			__antithesis_instrumentation__.Notify(190516)
		}
	}
	__antithesis_instrumentation__.Notify(190511)
	data, err := json.Marshal(rawValues)
	if err != nil {
		__antithesis_instrumentation__.Notify(190517)
		return vmoduleOptions{}, err
	} else {
		__antithesis_instrumentation__.Notify(190518)
	}
	__antithesis_instrumentation__.Notify(190512)
	var opts vmoduleOptions
	if err := json.Unmarshal(data, &opts); err != nil {
		__antithesis_instrumentation__.Notify(190519)
		return vmoduleOptions{}, err
	} else {
		__antithesis_instrumentation__.Notify(190520)
	}
	__antithesis_instrumentation__.Notify(190513)

	opts.setDefaults(values)
	return opts, nil
}

func (opts *vmoduleOptions) setDefaults(values url.Values) {
	__antithesis_instrumentation__.Notify(190521)
	if opts.Duration == 0 {
		__antithesis_instrumentation__.Notify(190523)
		opts.Duration = logSpyDefaultDuration
	} else {
		__antithesis_instrumentation__.Notify(190524)
	}
	__antithesis_instrumentation__.Notify(190522)

	_, opts.hasVModule = values["vmodule"]
}

type vmoduleServer struct {
	lock uint32
}

func (s *vmoduleServer) lockVModule(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(190525)
	if swapped := atomic.CompareAndSwapUint32(&s.lock, 0, 1); !swapped {
		__antithesis_instrumentation__.Notify(190527)
		return errors.New("another in-flight HTTP request is already managing vmodule")
	} else {
		__antithesis_instrumentation__.Notify(190528)
	}
	__antithesis_instrumentation__.Notify(190526)
	return nil
}

func (s *vmoduleServer) unlockVModule(ctx context.Context) {
	__antithesis_instrumentation__.Notify(190529)
	atomic.StoreUint32(&s.lock, 0)
}

func (s *vmoduleServer) vmoduleHandleDebug(w http.ResponseWriter, r *http.Request) {
	__antithesis_instrumentation__.Notify(190530)
	opts, err := loadVmoduleOptionsFromValues(r.URL.Query())
	if err != nil {
		__antithesis_instrumentation__.Notify(190532)
		http.Error(w, "while parsing options: "+err.Error(), http.StatusInternalServerError)
		return
	} else {
		__antithesis_instrumentation__.Notify(190533)
	}
	__antithesis_instrumentation__.Notify(190531)

	w.Header().Add("Content-type", "text/plain; charset=UTF-8")
	ctx := r.Context()
	if err := s.vmoduleHandleDebugInternal(ctx, w, opts); err != nil {
		__antithesis_instrumentation__.Notify(190534)

		log.Infof(ctx, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(190535)
	}
}

func (s *vmoduleServer) vmoduleHandleDebugInternal(
	ctx context.Context, w http.ResponseWriter, opts vmoduleOptions,
) error {
	__antithesis_instrumentation__.Notify(190536)
	prevSettings := log.GetVModule()

	_, err := w.Write([]byte("previous vmodule configuration: " + prevSettings + "\n"))
	if err != nil {
		__antithesis_instrumentation__.Notify(190545)
		return err
	} else {
		__antithesis_instrumentation__.Notify(190546)
	}
	__antithesis_instrumentation__.Notify(190537)
	if !opts.hasVModule {
		__antithesis_instrumentation__.Notify(190547)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(190548)
	}
	__antithesis_instrumentation__.Notify(190538)

	if err := s.lockVModule(ctx); err != nil {
		__antithesis_instrumentation__.Notify(190549)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(190550)
	}
	__antithesis_instrumentation__.Notify(190539)

	if err := log.SetVModule(opts.VModule); err != nil {
		__antithesis_instrumentation__.Notify(190551)
		s.unlockVModule(ctx)
		http.Error(w, "setting vmodule: "+err.Error(), http.StatusInternalServerError)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(190552)
	}
	__antithesis_instrumentation__.Notify(190540)

	_, err = w.Write([]byte("new vmodule configuration: " + opts.VModule + "\n"))
	if err != nil {
		__antithesis_instrumentation__.Notify(190553)
		s.unlockVModule(ctx)
		return err
	} else {
		__antithesis_instrumentation__.Notify(190554)
	}
	__antithesis_instrumentation__.Notify(190541)

	log.Infof(ctx, "configured vmodule: %q", redact.SafeString(opts.VModule))

	if opts.Duration <= 0 {
		__antithesis_instrumentation__.Notify(190555)
		s.unlockVModule(ctx)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(190556)
	}
	__antithesis_instrumentation__.Notify(190542)

	_, err = w.Write([]byte(fmt.Sprintf("will restore previous vmodule config after %s\n", opts.Duration)))
	if err != nil {
		__antithesis_instrumentation__.Notify(190557)
		s.unlockVModule(ctx)
		return err
	} else {
		__antithesis_instrumentation__.Notify(190558)
	}
	__antithesis_instrumentation__.Notify(190543)

	go func() {
		__antithesis_instrumentation__.Notify(190559)

		time.Sleep(time.Duration(opts.Duration))

		err := log.SetVModule(prevSettings)

		log.Infof(context.Background(), "restoring vmodule configuration (%q): %v", redact.SafeString(prevSettings), err)

		s.unlockVModule(context.Background())
	}()
	__antithesis_instrumentation__.Notify(190544)

	return nil
}
