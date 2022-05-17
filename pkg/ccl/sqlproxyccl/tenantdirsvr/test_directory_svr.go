package tenantdirsvr

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"container/list"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/ccl/sqlproxyccl/tenant"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/logtags"
	"google.golang.org/grpc"
)

var _ tenant.DirectoryServer = (*TestDirectoryServer)(nil)

type Process struct {
	Stopper  *stop.Stopper
	Cmd      *exec.Cmd
	SQL      net.Addr
	FakeLoad float32
}

func NewSubStopper(parentStopper *stop.Stopper) *stop.Stopper {
	__antithesis_instrumentation__.Notify(23159)
	mu := &syncutil.Mutex{}
	var subStopper *stop.Stopper
	parentStopper.AddCloser(stop.CloserFn(func() {
		__antithesis_instrumentation__.Notify(23162)
		mu.Lock()
		defer mu.Unlock()
		if subStopper == nil {
			__antithesis_instrumentation__.Notify(23164)
			subStopper = stop.NewStopper(stop.WithTracer(parentStopper.Tracer()))
		} else {
			__antithesis_instrumentation__.Notify(23165)
		}
		__antithesis_instrumentation__.Notify(23163)
		subStopper.Stop(context.Background())
	}))
	__antithesis_instrumentation__.Notify(23160)
	mu.Lock()
	defer mu.Unlock()
	if subStopper == nil {
		__antithesis_instrumentation__.Notify(23166)
		subStopper = stop.NewStopper(stop.WithTracer(parentStopper.Tracer()))
	} else {
		__antithesis_instrumentation__.Notify(23167)
	}
	__antithesis_instrumentation__.Notify(23161)
	return subStopper
}

type TestDirectoryServer struct {
	args                []string
	stopper             *stop.Stopper
	grpcServer          *grpc.Server
	cockroachExecutable string

	TenantStarterFunc func(ctx context.Context, tenantID uint64) (*Process, error)

	proc struct {
		syncutil.RWMutex
		processByAddrByTenantID map[uint64]map[net.Addr]*Process
	}
	listen struct {
		syncutil.RWMutex
		eventListeners *list.List
	}
}

func New(stopper *stop.Stopper, args ...string) (*TestDirectoryServer, error) {
	__antithesis_instrumentation__.Notify(23168)

	cockroachExecutable, err := os.Executable()
	if err != nil {
		__antithesis_instrumentation__.Notify(23170)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(23171)
	}
	__antithesis_instrumentation__.Notify(23169)
	dir := &TestDirectoryServer{
		args:                args,
		grpcServer:          grpc.NewServer(),
		stopper:             stopper,
		cockroachExecutable: cockroachExecutable,
	}
	dir.TenantStarterFunc = dir.startTenantLocked
	dir.proc.processByAddrByTenantID = map[uint64]map[net.Addr]*Process{}
	dir.listen.eventListeners = list.New()
	stopper.AddCloser(stop.CloserFn(dir.grpcServer.GracefulStop))
	tenant.RegisterDirectoryServer(dir.grpcServer, dir)
	return dir, nil
}

func (s *TestDirectoryServer) Stopper() *stop.Stopper {
	__antithesis_instrumentation__.Notify(23172)
	return s.stopper
}

func (s *TestDirectoryServer) Get(id roachpb.TenantID) (result map[net.Addr]*Process) {
	__antithesis_instrumentation__.Notify(23173)
	result = make(map[net.Addr]*Process)
	s.proc.RLock()
	defer s.proc.RUnlock()
	processes, ok := s.proc.processByAddrByTenantID[id.ToUint64()]
	if ok {
		__antithesis_instrumentation__.Notify(23175)
		for k, v := range processes {
			__antithesis_instrumentation__.Notify(23176)
			result[k] = v
		}
	} else {
		__antithesis_instrumentation__.Notify(23177)
	}
	__antithesis_instrumentation__.Notify(23174)
	return
}

func (s *TestDirectoryServer) StartTenant(ctx context.Context, id roachpb.TenantID) error {
	__antithesis_instrumentation__.Notify(23178)
	select {
	case <-s.stopper.ShouldQuiesce():
		__antithesis_instrumentation__.Notify(23182)
		return context.Canceled
	default:
		__antithesis_instrumentation__.Notify(23183)
	}
	__antithesis_instrumentation__.Notify(23179)

	ctx = logtags.AddTag(ctx, "tenant", id)

	s.proc.Lock()
	defer s.proc.Unlock()

	process, err := s.TenantStarterFunc(ctx, id.ToUint64())
	if err != nil {
		__antithesis_instrumentation__.Notify(23184)
		return err
	} else {
		__antithesis_instrumentation__.Notify(23185)
	}
	__antithesis_instrumentation__.Notify(23180)

	s.registerInstanceLocked(id.ToUint64(), process)
	process.Stopper.AddCloser(stop.CloserFn(func() {
		__antithesis_instrumentation__.Notify(23186)
		s.deregisterInstance(id.ToUint64(), process.SQL)
	}))
	__antithesis_instrumentation__.Notify(23181)

	return nil
}

func (s *TestDirectoryServer) SetFakeLoad(id roachpb.TenantID, addr net.Addr, fakeLoad float32) {
	__antithesis_instrumentation__.Notify(23187)
	s.proc.RLock()
	defer s.proc.RUnlock()

	if processes, ok := s.proc.processByAddrByTenantID[id.ToUint64()]; ok {
		__antithesis_instrumentation__.Notify(23189)
		if process, ok := processes[addr]; ok {
			__antithesis_instrumentation__.Notify(23190)
			process.FakeLoad = fakeLoad
		} else {
			__antithesis_instrumentation__.Notify(23191)
		}
	} else {
		__antithesis_instrumentation__.Notify(23192)
	}
	__antithesis_instrumentation__.Notify(23188)

	s.listen.RLock()
	defer s.listen.RUnlock()
	s.notifyEventListenersLocked(&tenant.WatchPodsResponse{
		Pod: &tenant.Pod{
			Addr:           addr.String(),
			TenantID:       id.ToUint64(),
			Load:           fakeLoad,
			State:          tenant.UNKNOWN,
			StateTimestamp: timeutil.Now(),
		},
	})
}

func (s *TestDirectoryServer) GetTenant(
	_ context.Context, _ *tenant.GetTenantRequest,
) (*tenant.GetTenantResponse, error) {
	__antithesis_instrumentation__.Notify(23193)
	return &tenant.GetTenantResponse{
		ClusterName: "tenant-cluster",
	}, nil
}

func (s *TestDirectoryServer) ListPods(
	ctx context.Context, req *tenant.ListPodsRequest,
) (*tenant.ListPodsResponse, error) {
	__antithesis_instrumentation__.Notify(23194)
	ctx = logtags.AddTag(ctx, "tenant", req.TenantID)
	s.proc.RLock()
	defer s.proc.RUnlock()
	return s.listLocked(ctx, req)
}

func (s *TestDirectoryServer) WatchPods(
	_ *tenant.WatchPodsRequest, server tenant.Directory_WatchPodsServer,
) error {
	__antithesis_instrumentation__.Notify(23195)
	select {
	case <-s.stopper.ShouldQuiesce():
		__antithesis_instrumentation__.Notify(23198)
		return context.Canceled
	default:
		__antithesis_instrumentation__.Notify(23199)
	}
	__antithesis_instrumentation__.Notify(23196)

	c := make(chan *tenant.WatchPodsResponse, 10)
	s.listen.Lock()
	elem := s.listen.eventListeners.PushBack(c)
	s.listen.Unlock()
	err := s.stopper.RunTask(context.Background(), "watch-pods-server",
		func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(23200)
		out:
			for {
				__antithesis_instrumentation__.Notify(23201)
				select {
				case e, ok := <-c:
					__antithesis_instrumentation__.Notify(23202)
					if !ok {
						__antithesis_instrumentation__.Notify(23205)
						break out
					} else {
						__antithesis_instrumentation__.Notify(23206)
					}
					__antithesis_instrumentation__.Notify(23203)
					if err := server.Send(e); err != nil {
						__antithesis_instrumentation__.Notify(23207)
						s.listen.Lock()
						s.listen.eventListeners.Remove(elem)
						close(c)
						s.listen.Unlock()
						break out
					} else {
						__antithesis_instrumentation__.Notify(23208)
					}
				case <-s.stopper.ShouldQuiesce():
					__antithesis_instrumentation__.Notify(23204)
					s.listen.Lock()
					s.listen.eventListeners.Remove(elem)
					close(c)
					s.listen.Unlock()
					break out
				}
			}
		})
	__antithesis_instrumentation__.Notify(23197)
	return err
}

func (s *TestDirectoryServer) Drain() {
	__antithesis_instrumentation__.Notify(23209)
	s.proc.RLock()
	defer s.proc.RUnlock()

	for tenantID, processByAddr := range s.proc.processByAddrByTenantID {
		__antithesis_instrumentation__.Notify(23210)
		for addr := range processByAddr {
			__antithesis_instrumentation__.Notify(23211)
			s.listen.RLock()
			defer s.listen.RUnlock()
			s.notifyEventListenersLocked(&tenant.WatchPodsResponse{
				Pod: &tenant.Pod{
					TenantID:       tenantID,
					Addr:           addr.String(),
					State:          tenant.DRAINING,
					StateTimestamp: timeutil.Now(),
				},
			})
		}
	}
}

func (s *TestDirectoryServer) notifyEventListenersLocked(req *tenant.WatchPodsResponse) {
	__antithesis_instrumentation__.Notify(23212)
	for e := s.listen.eventListeners.Front(); e != nil; {
		__antithesis_instrumentation__.Notify(23213)
		select {
		case e.Value.(chan *tenant.WatchPodsResponse) <- req:
			__antithesis_instrumentation__.Notify(23214)
			e = e.Next()
		default:
			__antithesis_instrumentation__.Notify(23215)

			eToClose := e
			e = e.Next()
			close(eToClose.Value.(chan *tenant.WatchPodsResponse))
			s.listen.eventListeners.Remove(eToClose)
		}
	}
}

func (s *TestDirectoryServer) EnsurePod(
	ctx context.Context, req *tenant.EnsurePodRequest,
) (*tenant.EnsurePodResponse, error) {
	__antithesis_instrumentation__.Notify(23216)
	select {
	case <-s.stopper.ShouldQuiesce():
		__antithesis_instrumentation__.Notify(23220)
		return nil, context.Canceled
	default:
		__antithesis_instrumentation__.Notify(23221)
	}
	__antithesis_instrumentation__.Notify(23217)

	ctx = logtags.AddTag(ctx, "tenant", req.TenantID)

	s.proc.Lock()
	defer s.proc.Unlock()

	lst, err := s.listLocked(ctx, &tenant.ListPodsRequest{TenantID: req.TenantID})
	if err != nil {
		__antithesis_instrumentation__.Notify(23222)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(23223)
	}
	__antithesis_instrumentation__.Notify(23218)
	if len(lst.Pods) == 0 {
		__antithesis_instrumentation__.Notify(23224)
		process, err := s.TenantStarterFunc(ctx, req.TenantID)
		if err != nil {
			__antithesis_instrumentation__.Notify(23226)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(23227)
		}
		__antithesis_instrumentation__.Notify(23225)
		s.registerInstanceLocked(req.TenantID, process)
		process.Stopper.AddCloser(stop.CloserFn(func() {
			__antithesis_instrumentation__.Notify(23228)
			s.deregisterInstance(req.TenantID, process.SQL)
		}))
	} else {
		__antithesis_instrumentation__.Notify(23229)
	}
	__antithesis_instrumentation__.Notify(23219)

	return &tenant.EnsurePodResponse{}, nil
}

func (s *TestDirectoryServer) Serve(listener net.Listener) error {
	__antithesis_instrumentation__.Notify(23230)
	return s.grpcServer.Serve(listener)
}

func (s *TestDirectoryServer) listLocked(
	_ context.Context, req *tenant.ListPodsRequest,
) (*tenant.ListPodsResponse, error) {
	__antithesis_instrumentation__.Notify(23231)
	processByAddr, ok := s.proc.processByAddrByTenantID[req.TenantID]
	if !ok {
		__antithesis_instrumentation__.Notify(23234)
		return &tenant.ListPodsResponse{}, nil
	} else {
		__antithesis_instrumentation__.Notify(23235)
	}
	__antithesis_instrumentation__.Notify(23232)
	resp := tenant.ListPodsResponse{}
	for addr, proc := range processByAddr {
		__antithesis_instrumentation__.Notify(23236)
		resp.Pods = append(resp.Pods, &tenant.Pod{
			TenantID:       req.TenantID,
			Addr:           addr.String(),
			State:          tenant.RUNNING,
			Load:           proc.FakeLoad,
			StateTimestamp: timeutil.Now(),
		})
	}
	__antithesis_instrumentation__.Notify(23233)
	return &resp, nil
}

func (s *TestDirectoryServer) registerInstanceLocked(tenantID uint64, process *Process) {
	__antithesis_instrumentation__.Notify(23237)
	processByAddr, ok := s.proc.processByAddrByTenantID[tenantID]
	if !ok {
		__antithesis_instrumentation__.Notify(23239)
		processByAddr = map[net.Addr]*Process{}
		s.proc.processByAddrByTenantID[tenantID] = processByAddr
	} else {
		__antithesis_instrumentation__.Notify(23240)
	}
	__antithesis_instrumentation__.Notify(23238)
	processByAddr[process.SQL] = process

	s.listen.RLock()
	defer s.listen.RUnlock()
	s.notifyEventListenersLocked(&tenant.WatchPodsResponse{
		Pod: &tenant.Pod{
			TenantID:       tenantID,
			Addr:           process.SQL.String(),
			State:          tenant.RUNNING,
			Load:           process.FakeLoad,
			StateTimestamp: timeutil.Now(),
		},
	})
}

func (s *TestDirectoryServer) deregisterInstance(tenantID uint64, sql net.Addr) {
	__antithesis_instrumentation__.Notify(23241)
	s.proc.Lock()
	defer s.proc.Unlock()
	processByAddr, ok := s.proc.processByAddrByTenantID[tenantID]
	if !ok {
		__antithesis_instrumentation__.Notify(23243)
		return
	} else {
		__antithesis_instrumentation__.Notify(23244)
	}
	__antithesis_instrumentation__.Notify(23242)

	if _, ok = processByAddr[sql]; ok {
		__antithesis_instrumentation__.Notify(23245)
		delete(processByAddr, sql)

		s.listen.RLock()
		defer s.listen.RUnlock()
		s.notifyEventListenersLocked(&tenant.WatchPodsResponse{
			Pod: &tenant.Pod{
				TenantID:       tenantID,
				Addr:           sql.String(),
				State:          tenant.DELETING,
				StateTimestamp: timeutil.Now(),
			},
		})
	} else {
		__antithesis_instrumentation__.Notify(23246)
	}
}

type writerFunc func(p []byte) (int, error)

func (wf writerFunc) Write(p []byte) (int, error) {
	__antithesis_instrumentation__.Notify(23247)
	return wf(p)
}

func (s *TestDirectoryServer) startTenantLocked(
	ctx context.Context, tenantID uint64,
) (*Process, error) {
	__antithesis_instrumentation__.Notify(23248)

	sqlListener, err := net.Listen("tcp", "")
	if err != nil {
		__antithesis_instrumentation__.Notify(23260)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(23261)
	}
	__antithesis_instrumentation__.Notify(23249)
	httpListener, err := net.Listen("tcp", "")
	if err != nil {
		__antithesis_instrumentation__.Notify(23262)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(23263)
	}
	__antithesis_instrumentation__.Notify(23250)
	process := &Process{SQL: sqlListener.Addr(), FakeLoad: 0.01}
	args := s.args
	if len(args) == 0 {
		__antithesis_instrumentation__.Notify(23264)
		args = append(args,
			s.cockroachExecutable, "mt", "start-sql", "--kv-addrs=:26257", "--insecure",
		)
	} else {
		__antithesis_instrumentation__.Notify(23265)
	}
	__antithesis_instrumentation__.Notify(23251)
	args = append(args,
		fmt.Sprintf("--sql-addr=%s", sqlListener.Addr().String()),
		fmt.Sprintf("--http-addr=%s", httpListener.Addr().String()),
		fmt.Sprintf("--tenant-id=%d", tenantID),
	)
	if err = sqlListener.Close(); err != nil {
		__antithesis_instrumentation__.Notify(23266)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(23267)
	}
	__antithesis_instrumentation__.Notify(23252)
	if err = httpListener.Close(); err != nil {
		__antithesis_instrumentation__.Notify(23268)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(23269)
	}
	__antithesis_instrumentation__.Notify(23253)

	c := exec.Command(args[0], args[1:]...)
	process.Cmd = c
	c.Env = append(os.Environ(), "COCKROACH_TRUST_CLIENT_PROVIDED_SQL_REMOTE_ADDR=true")
	var f writerFunc = func(p []byte) (int, error) {
		__antithesis_instrumentation__.Notify(23270)
		sc := bufio.NewScanner(strings.NewReader(string(p)))
		for sc.Scan() {
			__antithesis_instrumentation__.Notify(23272)
			log.Infof(ctx, "%s", sc.Text())
		}
		__antithesis_instrumentation__.Notify(23271)
		return len(p), nil
	}
	__antithesis_instrumentation__.Notify(23254)
	c.Stdout = f
	c.Stderr = f
	err = c.Start()
	if err != nil {
		__antithesis_instrumentation__.Notify(23273)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(23274)
	}
	__antithesis_instrumentation__.Notify(23255)
	process.Stopper = NewSubStopper(s.stopper)
	process.Stopper.AddCloser(stop.CloserFn(func() {
		__antithesis_instrumentation__.Notify(23275)
		_ = c.Process.Kill()
		s.deregisterInstance(tenantID, process.SQL)
	}))
	__antithesis_instrumentation__.Notify(23256)
	err = s.stopper.RunAsyncTask(ctx, "cmd-wait", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(23276)
		if err := c.Wait(); err != nil {
			__antithesis_instrumentation__.Notify(23278)
			log.Infof(ctx, "finished %s with err %s", process.Cmd.Args, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(23279)
		}
		__antithesis_instrumentation__.Notify(23277)
		log.Infof(ctx, "finished %s with success", process.Cmd.Args)
		process.Stopper.Stop(ctx)
	})
	__antithesis_instrumentation__.Notify(23257)
	if err != nil {
		__antithesis_instrumentation__.Notify(23280)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(23281)
	}
	__antithesis_instrumentation__.Notify(23258)

	start := timeutil.Now()
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transport}
	for {
		__antithesis_instrumentation__.Notify(23282)
		time.Sleep(300 * time.Millisecond)
		resp, err := client.Get(fmt.Sprintf("https://%s/health", httpListener.Addr().String()))
		waitTime := timeutil.Since(start)
		if err == nil {
			__antithesis_instrumentation__.Notify(23285)
			resp.Body.Close()
			log.Infof(ctx, "tenant is healthy")
			break
		} else {
			__antithesis_instrumentation__.Notify(23286)
		}
		__antithesis_instrumentation__.Notify(23283)
		if waitTime > 5*time.Second {
			__antithesis_instrumentation__.Notify(23287)
			log.Infof(ctx, "waited more than 5 sec for the tenant to get healthy and it still isn't")
			break
		} else {
			__antithesis_instrumentation__.Notify(23288)
		}
		__antithesis_instrumentation__.Notify(23284)
		log.Infof(ctx, "waiting %s for healthy tenant: %s", waitTime, err)
	}
	__antithesis_instrumentation__.Notify(23259)

	time.Sleep(time.Second)
	return process, nil
}
