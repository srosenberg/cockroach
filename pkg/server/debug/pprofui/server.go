package pprofui

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	runtimepprof "runtime/pprof"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/errors"
	"github.com/google/pprof/driver"
	"github.com/google/pprof/profile"
	"github.com/spf13/pflag"
)

type Profiler interface {
	Profile(
		ctx context.Context, req *serverpb.ProfileRequest,
	) (*serverpb.JSONResponse, error)
}

const (
	ProfileConcurrency = 2

	ProfileExpiry = 2 * time.Second
)

type Server struct {
	storage  Storage
	profiler Profiler
}

func NewServer(storage Storage, profiler Profiler) *Server {
	__antithesis_instrumentation__.Notify(190333)
	s := &Server{
		storage:  storage,
		profiler: profiler,
	}

	return s
}

func (s *Server) parsePath(reqPath string) (profType string, id string, remainingPath string) {
	__antithesis_instrumentation__.Notify(190334)
	parts := strings.Split(path.Clean(reqPath), "/")
	if parts[0] == "" {
		__antithesis_instrumentation__.Notify(190336)

		parts = parts[1:]
	} else {
		__antithesis_instrumentation__.Notify(190337)
	}
	__antithesis_instrumentation__.Notify(190335)
	switch len(parts) {
	case 0:
		__antithesis_instrumentation__.Notify(190338)
		return "", "", "/"
	case 1:
		__antithesis_instrumentation__.Notify(190339)
		return parts[0], "", "/"
	default:
		__antithesis_instrumentation__.Notify(190340)
		return parts[0], parts[1], "/" + strings.Join(parts[2:], "/")
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	__antithesis_instrumentation__.Notify(190341)
	profileName, id, remainingPath := s.parsePath(r.URL.Path)

	if profileName == "" {
		__antithesis_instrumentation__.Notify(190346)

		var names []string
		for _, p := range runtimepprof.Profiles() {
			__antithesis_instrumentation__.Notify(190348)
			names = append(names, p.Name())
		}
		__antithesis_instrumentation__.Notify(190347)
		sort.Strings(names)
		msg := fmt.Sprintf("Try %s for one of %s", path.Join(r.RequestURI, "<profileName>"), strings.Join(names, ", "))
		http.Error(w, msg, http.StatusNotFound)
		return
	} else {
		__antithesis_instrumentation__.Notify(190349)
	}
	__antithesis_instrumentation__.Notify(190342)

	if id != "" {
		__antithesis_instrumentation__.Notify(190350)

		if err := s.storage.Get(id, func(io.Reader) error { __antithesis_instrumentation__.Notify(190356); return nil }); err != nil {
			__antithesis_instrumentation__.Notify(190357)
			msg := fmt.Sprintf("profile for id %s not found: %s", id, err)
			http.Error(w, msg, http.StatusNotFound)
			return
		} else {
			__antithesis_instrumentation__.Notify(190358)
		}
		__antithesis_instrumentation__.Notify(190351)

		if r.URL.Query().Get("download") != "" {
			__antithesis_instrumentation__.Notify(190359)

			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s_%s.pb.gz", profileName, id))
			w.Header().Set("Content-Type", "application/octet-stream")
			if err := s.storage.Get(id, func(r io.Reader) error {
				__antithesis_instrumentation__.Notify(190361)
				_, err := io.Copy(w, r)
				return err
			}); err != nil {
				__antithesis_instrumentation__.Notify(190362)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			} else {
				__antithesis_instrumentation__.Notify(190363)
			}
			__antithesis_instrumentation__.Notify(190360)
			return
		} else {
			__antithesis_instrumentation__.Notify(190364)
		}
		__antithesis_instrumentation__.Notify(190352)

		server := func(args *driver.HTTPServerArgs) error {
			__antithesis_instrumentation__.Notify(190365)
			handler, ok := args.Handlers[remainingPath]
			if !ok {
				__antithesis_instrumentation__.Notify(190367)
				return errors.Errorf("unknown endpoint %s", remainingPath)
			} else {
				__antithesis_instrumentation__.Notify(190368)
			}
			__antithesis_instrumentation__.Notify(190366)
			handler.ServeHTTP(w, r)
			return nil
		}
		__antithesis_instrumentation__.Notify(190353)

		storageFetcher := func(_ string, _, _ time.Duration) (*profile.Profile, string, error) {
			__antithesis_instrumentation__.Notify(190369)
			var p *profile.Profile
			if err := s.storage.Get(id, func(reader io.Reader) error {
				__antithesis_instrumentation__.Notify(190371)
				var err error
				p, err = profile.Parse(reader)
				return err
			}); err != nil {
				__antithesis_instrumentation__.Notify(190372)
				return nil, "", err
			} else {
				__antithesis_instrumentation__.Notify(190373)
			}
			__antithesis_instrumentation__.Notify(190370)
			return p, "", nil
		}
		__antithesis_instrumentation__.Notify(190354)

		if err := driver.PProf(&driver.Options{
			Flagset: &pprofFlags{
				FlagSet: pflag.NewFlagSet("pprof", pflag.ExitOnError),
				args: []string{
					"--symbolize", "none",
					"--http", "localhost:0",
					"",
				},
			},
			UI:         &fakeUI{},
			Fetch:      fetcherFn(storageFetcher),
			HTTPServer: server,
		}); err != nil {
			__antithesis_instrumentation__.Notify(190374)
			_, _ = w.Write([]byte(err.Error()))
		} else {
			__antithesis_instrumentation__.Notify(190375)
		}
		__antithesis_instrumentation__.Notify(190355)

		return
	} else {
		__antithesis_instrumentation__.Notify(190376)
	}
	__antithesis_instrumentation__.Notify(190343)

	id = s.storage.ID()

	if err := s.storage.Store(id, func(w io.Writer) error {
		__antithesis_instrumentation__.Notify(190377)
		req, err := http.NewRequest("GET", "/unused", bytes.NewReader(nil))
		if err != nil {
			__antithesis_instrumentation__.Notify(190386)
			return err
		} else {
			__antithesis_instrumentation__.Notify(190387)
		}
		__antithesis_instrumentation__.Notify(190378)

		profileType, ok := serverpb.ProfileRequest_Type_value[strings.ToUpper(profileName)]
		if !ok && func() bool {
			__antithesis_instrumentation__.Notify(190388)
			return profileName != "profile" == true
		}() == true {
			__antithesis_instrumentation__.Notify(190389)
			return errors.Newf("unknown profile name: %s", profileName)
		} else {
			__antithesis_instrumentation__.Notify(190390)
		}
		__antithesis_instrumentation__.Notify(190379)

		if profileName == "profile" {
			__antithesis_instrumentation__.Notify(190391)
			profileType = int32(serverpb.ProfileRequest_CPU)
		} else {
			__antithesis_instrumentation__.Notify(190392)
		}
		__antithesis_instrumentation__.Notify(190380)
		var resp *serverpb.JSONResponse
		profileReq := &serverpb.ProfileRequest{
			NodeId: "local",
			Type:   serverpb.ProfileRequest_Type(profileType),
		}

		_ = r.ParseForm()
		req.Form = r.Form

		if r.Form.Get("seconds") != "" {
			__antithesis_instrumentation__.Notify(190393)
			sec, err := strconv.ParseInt(r.Form.Get("seconds"), 10, 32)
			if err != nil {
				__antithesis_instrumentation__.Notify(190395)
				return err
			} else {
				__antithesis_instrumentation__.Notify(190396)
			}
			__antithesis_instrumentation__.Notify(190394)
			profileReq.Seconds = int32(sec)
		} else {
			__antithesis_instrumentation__.Notify(190397)
		}
		__antithesis_instrumentation__.Notify(190381)
		if r.Form.Get("node") != "" {
			__antithesis_instrumentation__.Notify(190398)
			profileReq.NodeId = r.Form.Get("node")
		} else {
			__antithesis_instrumentation__.Notify(190399)
		}
		__antithesis_instrumentation__.Notify(190382)
		if r.Form.Get("labels") != "" {
			__antithesis_instrumentation__.Notify(190400)
			labels, err := strconv.ParseBool(r.Form.Get("labels"))
			if err != nil {
				__antithesis_instrumentation__.Notify(190402)
				return err
			} else {
				__antithesis_instrumentation__.Notify(190403)
			}
			__antithesis_instrumentation__.Notify(190401)
			profileReq.Labels = labels
		} else {
			__antithesis_instrumentation__.Notify(190404)
		}
		__antithesis_instrumentation__.Notify(190383)
		resp, err = s.profiler.Profile(r.Context(), profileReq)
		if err != nil {
			__antithesis_instrumentation__.Notify(190405)
			return err
		} else {
			__antithesis_instrumentation__.Notify(190406)
		}
		__antithesis_instrumentation__.Notify(190384)
		_, err = w.Write(resp.Data)
		if err != nil {
			__antithesis_instrumentation__.Notify(190407)
			return err
		} else {
			__antithesis_instrumentation__.Notify(190408)
		}
		__antithesis_instrumentation__.Notify(190385)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(190409)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		__antithesis_instrumentation__.Notify(190410)
	}
	__antithesis_instrumentation__.Notify(190344)

	origURL, err := url.Parse(r.RequestURI)
	if err != nil {
		__antithesis_instrumentation__.Notify(190411)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		__antithesis_instrumentation__.Notify(190412)
	}
	__antithesis_instrumentation__.Notify(190345)

	isGoPProf := strings.Contains(r.Header.Get("User-Agent"), "Go-http-client")
	origURL.Path = path.Join(origURL.Path, id, "flamegraph")
	if !isGoPProf {
		__antithesis_instrumentation__.Notify(190413)
		http.Redirect(w, r, origURL.String(), http.StatusTemporaryRedirect)
	} else {
		__antithesis_instrumentation__.Notify(190414)
		_ = s.storage.Get(id, func(r io.Reader) error {
			__antithesis_instrumentation__.Notify(190415)
			_, err := io.Copy(w, r)
			return err
		})
	}
}

type fetcherFn func(_ string, _, _ time.Duration) (*profile.Profile, string, error)

func (f fetcherFn) Fetch(s string, d, t time.Duration) (*profile.Profile, string, error) {
	__antithesis_instrumentation__.Notify(190416)
	return f(s, d, t)
}
