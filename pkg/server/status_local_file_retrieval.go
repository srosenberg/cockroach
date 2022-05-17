package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/server/debug"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func profileLocal(
	ctx context.Context, req *serverpb.ProfileRequest, st *cluster.Settings,
) (*serverpb.JSONResponse, error) {
	__antithesis_instrumentation__.Notify(238006)
	switch req.Type {
	case serverpb.ProfileRequest_CPU:
		__antithesis_instrumentation__.Notify(238007)
		var buf bytes.Buffer
		profileType := cluster.CPUProfileDefault
		if req.Labels {
			__antithesis_instrumentation__.Notify(238014)
			profileType = cluster.CPUProfileWithLabels
		} else {
			__antithesis_instrumentation__.Notify(238015)
		}
		__antithesis_instrumentation__.Notify(238008)
		if err := debug.CPUProfileDo(st, profileType, func() error {
			__antithesis_instrumentation__.Notify(238016)
			duration := 30 * time.Second
			if req.Seconds != 0 {
				__antithesis_instrumentation__.Notify(238019)
				duration = time.Duration(req.Seconds) * time.Second
			} else {
				__antithesis_instrumentation__.Notify(238020)
			}
			__antithesis_instrumentation__.Notify(238017)
			if err := pprof.StartCPUProfile(&buf); err != nil {
				__antithesis_instrumentation__.Notify(238021)

				return serverError(ctx, err)
			} else {
				__antithesis_instrumentation__.Notify(238022)
			}
			__antithesis_instrumentation__.Notify(238018)
			defer pprof.StopCPUProfile()
			select {
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(238023)
				return ctx.Err()
			case <-time.After(duration):
				__antithesis_instrumentation__.Notify(238024)
				return nil
			}
		}); err != nil {
			__antithesis_instrumentation__.Notify(238025)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(238026)
		}
		__antithesis_instrumentation__.Notify(238009)
		return &serverpb.JSONResponse{Data: buf.Bytes()}, nil
	default:
		__antithesis_instrumentation__.Notify(238010)
		name, ok := serverpb.ProfileRequest_Type_name[int32(req.Type)]
		if !ok {
			__antithesis_instrumentation__.Notify(238027)
			return nil, status.Errorf(codes.InvalidArgument, "unknown profile: %d", req.Type)
		} else {
			__antithesis_instrumentation__.Notify(238028)
		}
		__antithesis_instrumentation__.Notify(238011)
		name = strings.ToLower(name)
		p := pprof.Lookup(name)
		if p == nil {
			__antithesis_instrumentation__.Notify(238029)
			return nil, status.Errorf(codes.InvalidArgument, "unable to find profile: %s", name)
		} else {
			__antithesis_instrumentation__.Notify(238030)
		}
		__antithesis_instrumentation__.Notify(238012)
		var buf bytes.Buffer
		if err := p.WriteTo(&buf, 0); err != nil {
			__antithesis_instrumentation__.Notify(238031)
			return nil, status.Errorf(codes.Internal, err.Error())
		} else {
			__antithesis_instrumentation__.Notify(238032)
		}
		__antithesis_instrumentation__.Notify(238013)
		return &serverpb.JSONResponse{Data: buf.Bytes()}, nil
	}
}

func stacksLocal(req *serverpb.StacksRequest) (*serverpb.JSONResponse, error) {
	__antithesis_instrumentation__.Notify(238033)
	var stackType int
	switch req.Type {
	case serverpb.StacksType_GOROUTINE_STACKS:
		__antithesis_instrumentation__.Notify(238036)
		stackType = 2
	case serverpb.StacksType_GOROUTINE_STACKS_DEBUG_1:
		__antithesis_instrumentation__.Notify(238037)
		stackType = 1
	default:
		__antithesis_instrumentation__.Notify(238038)
		return nil, status.Errorf(codes.InvalidArgument, "unknown stacks type: %s", req.Type)
	}
	__antithesis_instrumentation__.Notify(238034)

	var buf bytes.Buffer
	if err := pprof.Lookup("goroutine").WriteTo(&buf, stackType); err != nil {
		__antithesis_instrumentation__.Notify(238039)
		return nil, status.Errorf(codes.Unknown, "failed to write goroutine stack: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(238040)
	}
	__antithesis_instrumentation__.Notify(238035)
	return &serverpb.JSONResponse{Data: buf.Bytes()}, nil
}

func getLocalFiles(
	req *serverpb.GetFilesRequest, heapProfileDirName string, goroutineDumpDirName string,
) (*serverpb.GetFilesResponse, error) {
	__antithesis_instrumentation__.Notify(238041)
	var dir string
	switch req.Type {

	case serverpb.FileType_HEAP:
		__antithesis_instrumentation__.Notify(238045)
		dir = heapProfileDirName
	case serverpb.FileType_GOROUTINES:
		__antithesis_instrumentation__.Notify(238046)
		dir = goroutineDumpDirName
	default:
		__antithesis_instrumentation__.Notify(238047)
		return nil, status.Errorf(codes.InvalidArgument, "unknown file type: %s", req.Type)
	}
	__antithesis_instrumentation__.Notify(238042)
	if dir == "" {
		__antithesis_instrumentation__.Notify(238048)
		return nil, status.Errorf(codes.Unimplemented, "dump directory not configured: %s", req.Type)
	} else {
		__antithesis_instrumentation__.Notify(238049)
	}
	__antithesis_instrumentation__.Notify(238043)
	var resp serverpb.GetFilesResponse
	for _, pattern := range req.Patterns {
		__antithesis_instrumentation__.Notify(238050)
		if err := checkFilePattern(pattern); err != nil {
			__antithesis_instrumentation__.Notify(238053)
			return nil, status.Errorf(codes.InvalidArgument, err.Error())
		} else {
			__antithesis_instrumentation__.Notify(238054)
		}
		__antithesis_instrumentation__.Notify(238051)
		filepaths, err := filepath.Glob(filepath.Join(dir, pattern))
		if err != nil {
			__antithesis_instrumentation__.Notify(238055)
			return nil, status.Errorf(codes.InvalidArgument, "bad pattern: %s", pattern)
		} else {
			__antithesis_instrumentation__.Notify(238056)
		}
		__antithesis_instrumentation__.Notify(238052)

		for _, path := range filepaths {
			__antithesis_instrumentation__.Notify(238057)
			fileinfo, _ := os.Stat(path)
			var contents []byte
			if !req.ListOnly {
				__antithesis_instrumentation__.Notify(238059)
				contents, err = ioutil.ReadFile(path)
				if err != nil {
					__antithesis_instrumentation__.Notify(238060)
					return nil, status.Errorf(codes.Internal, err.Error())
				} else {
					__antithesis_instrumentation__.Notify(238061)
				}
			} else {
				__antithesis_instrumentation__.Notify(238062)
			}
			__antithesis_instrumentation__.Notify(238058)
			resp.Files = append(resp.Files,
				&serverpb.File{Name: fileinfo.Name(), FileSize: fileinfo.Size(), Contents: contents})
		}
	}
	__antithesis_instrumentation__.Notify(238044)
	return &resp, nil
}
