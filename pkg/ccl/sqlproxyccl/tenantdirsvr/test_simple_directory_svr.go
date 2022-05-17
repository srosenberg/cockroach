package tenantdirsvr

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/ccl/sqlproxyccl/tenant"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/gogo/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type TestSimpleDirectoryServer struct {
	podAddr string

	mu struct {
		syncutil.Mutex

		deleted map[roachpb.TenantID]struct{}
	}
}

var _ tenant.DirectoryServer = &TestSimpleDirectoryServer{}

func NewTestSimpleDirectoryServer(podAddr string) (tenant.DirectoryServer, *grpc.Server) {
	__antithesis_instrumentation__.Notify(23289)
	dir := &TestSimpleDirectoryServer{podAddr: podAddr}
	dir.mu.deleted = make(map[roachpb.TenantID]struct{})
	grpcServer := grpc.NewServer()
	tenant.RegisterDirectoryServer(grpcServer, dir)
	return dir, grpcServer
}

func (d *TestSimpleDirectoryServer) ListPods(
	ctx context.Context, req *tenant.ListPodsRequest,
) (*tenant.ListPodsResponse, error) {
	__antithesis_instrumentation__.Notify(23290)
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.mu.deleted[roachpb.MakeTenantID(req.TenantID)]; ok {
		__antithesis_instrumentation__.Notify(23292)
		return &tenant.ListPodsResponse{}, nil
	} else {
		__antithesis_instrumentation__.Notify(23293)
	}
	__antithesis_instrumentation__.Notify(23291)
	return &tenant.ListPodsResponse{
		Pods: []*tenant.Pod{
			{
				TenantID:       req.TenantID,
				Addr:           d.podAddr,
				State:          tenant.RUNNING,
				StateTimestamp: timeutil.Now(),
			},
		},
	}, nil
}

func (d *TestSimpleDirectoryServer) WatchPods(
	req *tenant.WatchPodsRequest, server tenant.Directory_WatchPodsServer,
) error {
	__antithesis_instrumentation__.Notify(23294)
	return nil
}

func (d *TestSimpleDirectoryServer) EnsurePod(
	ctx context.Context, req *tenant.EnsurePodRequest,
) (*tenant.EnsurePodResponse, error) {
	__antithesis_instrumentation__.Notify(23295)
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.mu.deleted[roachpb.MakeTenantID(req.TenantID)]; ok {
		__antithesis_instrumentation__.Notify(23297)
		return nil, status.Errorf(codes.NotFound, "tenant has been deleted")
	} else {
		__antithesis_instrumentation__.Notify(23298)
	}
	__antithesis_instrumentation__.Notify(23296)
	return &tenant.EnsurePodResponse{}, nil
}

func (d *TestSimpleDirectoryServer) GetTenant(
	ctx context.Context, req *tenant.GetTenantRequest,
) (*tenant.GetTenantResponse, error) {
	__antithesis_instrumentation__.Notify(23299)
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.mu.deleted[roachpb.MakeTenantID(req.TenantID)]; ok {
		__antithesis_instrumentation__.Notify(23301)
		return nil, status.Errorf(codes.NotFound, "tenant has been deleted")
	} else {
		__antithesis_instrumentation__.Notify(23302)
	}
	__antithesis_instrumentation__.Notify(23300)

	return &tenant.GetTenantResponse{}, nil
}

func (d *TestSimpleDirectoryServer) DeleteTenant(tenantID roachpb.TenantID) {
	__antithesis_instrumentation__.Notify(23303)
	d.mu.Lock()
	defer d.mu.Unlock()
	d.mu.deleted[tenantID] = struct{}{}
}
