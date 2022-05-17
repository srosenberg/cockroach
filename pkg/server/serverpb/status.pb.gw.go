/*
Package serverpb is a reverse proxy.

It translates gRPC into RESTful JSON APIs.
*/
package serverpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"io"
	"net/http"

	"github.com/golang/protobuf/descriptor"
	"github.com/golang/protobuf/proto"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/utilities"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var _ codes.Code
var _ io.Reader
var _ status.Status
var _ = runtime.String
var _ = utilities.NewDoubleArray
var _ = descriptor.ForMessage
var _ = metadata.Join

func request_Status_Certificates_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233424)
	var protoReq CertificatesRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233427)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233428)
	}
	__antithesis_instrumentation__.Notify(233425)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233429)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233430)
	}
	__antithesis_instrumentation__.Notify(233426)

	msg, err := client.Certificates(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_Certificates_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233431)
	var protoReq CertificatesRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233434)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233435)
	}
	__antithesis_instrumentation__.Notify(233432)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233436)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233437)
	}
	__antithesis_instrumentation__.Notify(233433)

	msg, err := server.Certificates(ctx, &protoReq)
	return msg, metadata, err

}

func request_Status_Details_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233438)
	var protoReq DetailsRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233441)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233442)
	}
	__antithesis_instrumentation__.Notify(233439)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233443)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233444)
	}
	__antithesis_instrumentation__.Notify(233440)

	msg, err := client.Details(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_Details_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233445)
	var protoReq DetailsRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233448)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233449)
	}
	__antithesis_instrumentation__.Notify(233446)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233450)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233451)
	}
	__antithesis_instrumentation__.Notify(233447)

	msg, err := server.Details(ctx, &protoReq)
	return msg, metadata, err

}

func request_Status_Nodes_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233452)
	var protoReq NodesRequest
	var metadata runtime.ServerMetadata

	msg, err := client.Nodes(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_Nodes_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233453)
	var protoReq NodesRequest
	var metadata runtime.ServerMetadata

	msg, err := server.Nodes(ctx, &protoReq)
	return msg, metadata, err

}

func request_Status_Node_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233454)
	var protoReq NodeRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233457)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233458)
	}
	__antithesis_instrumentation__.Notify(233455)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233459)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233460)
	}
	__antithesis_instrumentation__.Notify(233456)

	msg, err := client.Node(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_Node_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233461)
	var protoReq NodeRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233464)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233465)
	}
	__antithesis_instrumentation__.Notify(233462)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233466)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233467)
	}
	__antithesis_instrumentation__.Notify(233463)

	msg, err := server.Node(ctx, &protoReq)
	return msg, metadata, err

}

func request_Status_NodesUI_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233468)
	var protoReq NodesRequest
	var metadata runtime.ServerMetadata

	msg, err := client.NodesUI(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_NodesUI_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233469)
	var protoReq NodesRequest
	var metadata runtime.ServerMetadata

	msg, err := server.NodesUI(ctx, &protoReq)
	return msg, metadata, err

}

func request_Status_NodeUI_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233470)
	var protoReq NodeRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233473)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233474)
	}
	__antithesis_instrumentation__.Notify(233471)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233475)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233476)
	}
	__antithesis_instrumentation__.Notify(233472)

	msg, err := client.NodeUI(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_NodeUI_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233477)
	var protoReq NodeRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233480)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233481)
	}
	__antithesis_instrumentation__.Notify(233478)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233482)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233483)
	}
	__antithesis_instrumentation__.Notify(233479)

	msg, err := server.NodeUI(ctx, &protoReq)
	return msg, metadata, err

}

var (
	filter_Status_RaftDebug_0 = &utilities.DoubleArray{Encoding: map[string]int{}, Base: []int(nil), Check: []int(nil)}
)

func request_Status_RaftDebug_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233484)
	var protoReq RaftDebugRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(233487)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233488)
	}
	__antithesis_instrumentation__.Notify(233485)
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_Status_RaftDebug_0); err != nil {
		__antithesis_instrumentation__.Notify(233489)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233490)
	}
	__antithesis_instrumentation__.Notify(233486)

	msg, err := client.RaftDebug(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_RaftDebug_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233491)
	var protoReq RaftDebugRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(233494)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233495)
	}
	__antithesis_instrumentation__.Notify(233492)
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_Status_RaftDebug_0); err != nil {
		__antithesis_instrumentation__.Notify(233496)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233497)
	}
	__antithesis_instrumentation__.Notify(233493)

	msg, err := server.RaftDebug(ctx, &protoReq)
	return msg, metadata, err

}

var (
	filter_Status_Ranges_0 = &utilities.DoubleArray{Encoding: map[string]int{"node_id": 0}, Base: []int{1, 1, 0}, Check: []int{0, 1, 2}}
)

func request_Status_Ranges_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233498)
	var protoReq RangesRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233503)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233504)
	}
	__antithesis_instrumentation__.Notify(233499)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233505)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233506)
	}
	__antithesis_instrumentation__.Notify(233500)

	if err := req.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(233507)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233508)
	}
	__antithesis_instrumentation__.Notify(233501)
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_Status_Ranges_0); err != nil {
		__antithesis_instrumentation__.Notify(233509)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233510)
	}
	__antithesis_instrumentation__.Notify(233502)

	msg, err := client.Ranges(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_Ranges_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233511)
	var protoReq RangesRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233516)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233517)
	}
	__antithesis_instrumentation__.Notify(233512)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233518)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233519)
	}
	__antithesis_instrumentation__.Notify(233513)

	if err := req.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(233520)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233521)
	}
	__antithesis_instrumentation__.Notify(233514)
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_Status_Ranges_0); err != nil {
		__antithesis_instrumentation__.Notify(233522)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233523)
	}
	__antithesis_instrumentation__.Notify(233515)

	msg, err := server.Ranges(ctx, &protoReq)
	return msg, metadata, err

}

func request_Status_TenantRanges_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233524)
	var protoReq TenantRangesRequest
	var metadata runtime.ServerMetadata

	msg, err := client.TenantRanges(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_TenantRanges_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233525)
	var protoReq TenantRangesRequest
	var metadata runtime.ServerMetadata

	msg, err := server.TenantRanges(ctx, &protoReq)
	return msg, metadata, err

}

func request_Status_Gossip_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233526)
	var protoReq GossipRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233529)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233530)
	}
	__antithesis_instrumentation__.Notify(233527)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233531)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233532)
	}
	__antithesis_instrumentation__.Notify(233528)

	msg, err := client.Gossip(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_Gossip_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233533)
	var protoReq GossipRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233536)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233537)
	}
	__antithesis_instrumentation__.Notify(233534)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233538)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233539)
	}
	__antithesis_instrumentation__.Notify(233535)

	msg, err := server.Gossip(ctx, &protoReq)
	return msg, metadata, err

}

func request_Status_EngineStats_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233540)
	var protoReq EngineStatsRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233543)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233544)
	}
	__antithesis_instrumentation__.Notify(233541)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233545)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233546)
	}
	__antithesis_instrumentation__.Notify(233542)

	msg, err := client.EngineStats(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_EngineStats_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233547)
	var protoReq EngineStatsRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233550)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233551)
	}
	__antithesis_instrumentation__.Notify(233548)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233552)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233553)
	}
	__antithesis_instrumentation__.Notify(233549)

	msg, err := server.EngineStats(ctx, &protoReq)
	return msg, metadata, err

}

var (
	filter_Status_Allocator_0 = &utilities.DoubleArray{Encoding: map[string]int{"node_id": 0}, Base: []int{1, 1, 0}, Check: []int{0, 1, 2}}
)

func request_Status_Allocator_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233554)
	var protoReq AllocatorRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233559)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233560)
	}
	__antithesis_instrumentation__.Notify(233555)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233561)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233562)
	}
	__antithesis_instrumentation__.Notify(233556)

	if err := req.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(233563)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233564)
	}
	__antithesis_instrumentation__.Notify(233557)
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_Status_Allocator_0); err != nil {
		__antithesis_instrumentation__.Notify(233565)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233566)
	}
	__antithesis_instrumentation__.Notify(233558)

	msg, err := client.Allocator(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_Allocator_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233567)
	var protoReq AllocatorRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233572)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233573)
	}
	__antithesis_instrumentation__.Notify(233568)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233574)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233575)
	}
	__antithesis_instrumentation__.Notify(233569)

	if err := req.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(233576)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233577)
	}
	__antithesis_instrumentation__.Notify(233570)
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_Status_Allocator_0); err != nil {
		__antithesis_instrumentation__.Notify(233578)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233579)
	}
	__antithesis_instrumentation__.Notify(233571)

	msg, err := server.Allocator(ctx, &protoReq)
	return msg, metadata, err

}

func request_Status_AllocatorRange_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233580)
	var protoReq AllocatorRangeRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["range_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233583)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "range_id")
	} else {
		__antithesis_instrumentation__.Notify(233584)
	}
	__antithesis_instrumentation__.Notify(233581)

	protoReq.RangeId, err = runtime.Int64(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233585)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "range_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233586)
	}
	__antithesis_instrumentation__.Notify(233582)

	msg, err := client.AllocatorRange(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_AllocatorRange_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233587)
	var protoReq AllocatorRangeRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["range_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233590)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "range_id")
	} else {
		__antithesis_instrumentation__.Notify(233591)
	}
	__antithesis_instrumentation__.Notify(233588)

	protoReq.RangeId, err = runtime.Int64(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233592)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "range_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233593)
	}
	__antithesis_instrumentation__.Notify(233589)

	msg, err := server.AllocatorRange(ctx, &protoReq)
	return msg, metadata, err

}

var (
	filter_Status_ListSessions_0 = &utilities.DoubleArray{Encoding: map[string]int{}, Base: []int(nil), Check: []int(nil)}
)

func request_Status_ListSessions_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233594)
	var protoReq ListSessionsRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(233597)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233598)
	}
	__antithesis_instrumentation__.Notify(233595)
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_Status_ListSessions_0); err != nil {
		__antithesis_instrumentation__.Notify(233599)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233600)
	}
	__antithesis_instrumentation__.Notify(233596)

	msg, err := client.ListSessions(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_ListSessions_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233601)
	var protoReq ListSessionsRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(233604)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233605)
	}
	__antithesis_instrumentation__.Notify(233602)
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_Status_ListSessions_0); err != nil {
		__antithesis_instrumentation__.Notify(233606)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233607)
	}
	__antithesis_instrumentation__.Notify(233603)

	msg, err := server.ListSessions(ctx, &protoReq)
	return msg, metadata, err

}

var (
	filter_Status_ListLocalSessions_0 = &utilities.DoubleArray{Encoding: map[string]int{}, Base: []int(nil), Check: []int(nil)}
)

func request_Status_ListLocalSessions_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233608)
	var protoReq ListSessionsRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(233611)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233612)
	}
	__antithesis_instrumentation__.Notify(233609)
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_Status_ListLocalSessions_0); err != nil {
		__antithesis_instrumentation__.Notify(233613)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233614)
	}
	__antithesis_instrumentation__.Notify(233610)

	msg, err := client.ListLocalSessions(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_ListLocalSessions_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233615)
	var protoReq ListSessionsRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(233618)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233619)
	}
	__antithesis_instrumentation__.Notify(233616)
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_Status_ListLocalSessions_0); err != nil {
		__antithesis_instrumentation__.Notify(233620)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233621)
	}
	__antithesis_instrumentation__.Notify(233617)

	msg, err := server.ListLocalSessions(ctx, &protoReq)
	return msg, metadata, err

}

func request_Status_CancelQuery_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233622)
	var protoReq CancelQueryRequest
	var metadata runtime.ServerMetadata

	newReader, berr := utilities.IOReaderFactory(req.Body)
	if berr != nil {
		__antithesis_instrumentation__.Notify(233627)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", berr)
	} else {
		__antithesis_instrumentation__.Notify(233628)
	}
	__antithesis_instrumentation__.Notify(233623)
	if err := marshaler.NewDecoder(newReader()).Decode(&protoReq); err != nil && func() bool {
		__antithesis_instrumentation__.Notify(233629)
		return err != io.EOF == true
	}() == true {
		__antithesis_instrumentation__.Notify(233630)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233631)
	}
	__antithesis_instrumentation__.Notify(233624)

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233632)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233633)
	}
	__antithesis_instrumentation__.Notify(233625)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233634)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233635)
	}
	__antithesis_instrumentation__.Notify(233626)

	msg, err := client.CancelQuery(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_CancelQuery_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233636)
	var protoReq CancelQueryRequest
	var metadata runtime.ServerMetadata

	newReader, berr := utilities.IOReaderFactory(req.Body)
	if berr != nil {
		__antithesis_instrumentation__.Notify(233641)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", berr)
	} else {
		__antithesis_instrumentation__.Notify(233642)
	}
	__antithesis_instrumentation__.Notify(233637)
	if err := marshaler.NewDecoder(newReader()).Decode(&protoReq); err != nil && func() bool {
		__antithesis_instrumentation__.Notify(233643)
		return err != io.EOF == true
	}() == true {
		__antithesis_instrumentation__.Notify(233644)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233645)
	}
	__antithesis_instrumentation__.Notify(233638)

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233646)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233647)
	}
	__antithesis_instrumentation__.Notify(233639)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233648)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233649)
	}
	__antithesis_instrumentation__.Notify(233640)

	msg, err := server.CancelQuery(ctx, &protoReq)
	return msg, metadata, err

}

func request_Status_CancelLocalQuery_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233650)
	var protoReq CancelQueryRequest
	var metadata runtime.ServerMetadata

	newReader, berr := utilities.IOReaderFactory(req.Body)
	if berr != nil {
		__antithesis_instrumentation__.Notify(233653)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", berr)
	} else {
		__antithesis_instrumentation__.Notify(233654)
	}
	__antithesis_instrumentation__.Notify(233651)
	if err := marshaler.NewDecoder(newReader()).Decode(&protoReq); err != nil && func() bool {
		__antithesis_instrumentation__.Notify(233655)
		return err != io.EOF == true
	}() == true {
		__antithesis_instrumentation__.Notify(233656)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233657)
	}
	__antithesis_instrumentation__.Notify(233652)

	msg, err := client.CancelLocalQuery(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_CancelLocalQuery_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233658)
	var protoReq CancelQueryRequest
	var metadata runtime.ServerMetadata

	newReader, berr := utilities.IOReaderFactory(req.Body)
	if berr != nil {
		__antithesis_instrumentation__.Notify(233661)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", berr)
	} else {
		__antithesis_instrumentation__.Notify(233662)
	}
	__antithesis_instrumentation__.Notify(233659)
	if err := marshaler.NewDecoder(newReader()).Decode(&protoReq); err != nil && func() bool {
		__antithesis_instrumentation__.Notify(233663)
		return err != io.EOF == true
	}() == true {
		__antithesis_instrumentation__.Notify(233664)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233665)
	}
	__antithesis_instrumentation__.Notify(233660)

	msg, err := server.CancelLocalQuery(ctx, &protoReq)
	return msg, metadata, err

}

func request_Status_ListContentionEvents_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233666)
	var protoReq ListContentionEventsRequest
	var metadata runtime.ServerMetadata

	msg, err := client.ListContentionEvents(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_ListContentionEvents_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233667)
	var protoReq ListContentionEventsRequest
	var metadata runtime.ServerMetadata

	msg, err := server.ListContentionEvents(ctx, &protoReq)
	return msg, metadata, err

}

func request_Status_ListLocalContentionEvents_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233668)
	var protoReq ListContentionEventsRequest
	var metadata runtime.ServerMetadata

	msg, err := client.ListLocalContentionEvents(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_ListLocalContentionEvents_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233669)
	var protoReq ListContentionEventsRequest
	var metadata runtime.ServerMetadata

	msg, err := server.ListLocalContentionEvents(ctx, &protoReq)
	return msg, metadata, err

}

func request_Status_ListDistSQLFlows_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233670)
	var protoReq ListDistSQLFlowsRequest
	var metadata runtime.ServerMetadata

	msg, err := client.ListDistSQLFlows(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_ListDistSQLFlows_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233671)
	var protoReq ListDistSQLFlowsRequest
	var metadata runtime.ServerMetadata

	msg, err := server.ListDistSQLFlows(ctx, &protoReq)
	return msg, metadata, err

}

func request_Status_ListLocalDistSQLFlows_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233672)
	var protoReq ListDistSQLFlowsRequest
	var metadata runtime.ServerMetadata

	msg, err := client.ListLocalDistSQLFlows(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_ListLocalDistSQLFlows_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233673)
	var protoReq ListDistSQLFlowsRequest
	var metadata runtime.ServerMetadata

	msg, err := server.ListLocalDistSQLFlows(ctx, &protoReq)
	return msg, metadata, err

}

func request_Status_CancelSession_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233674)
	var protoReq CancelSessionRequest
	var metadata runtime.ServerMetadata

	newReader, berr := utilities.IOReaderFactory(req.Body)
	if berr != nil {
		__antithesis_instrumentation__.Notify(233679)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", berr)
	} else {
		__antithesis_instrumentation__.Notify(233680)
	}
	__antithesis_instrumentation__.Notify(233675)
	if err := marshaler.NewDecoder(newReader()).Decode(&protoReq); err != nil && func() bool {
		__antithesis_instrumentation__.Notify(233681)
		return err != io.EOF == true
	}() == true {
		__antithesis_instrumentation__.Notify(233682)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233683)
	}
	__antithesis_instrumentation__.Notify(233676)

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233684)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233685)
	}
	__antithesis_instrumentation__.Notify(233677)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233686)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233687)
	}
	__antithesis_instrumentation__.Notify(233678)

	msg, err := client.CancelSession(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_CancelSession_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233688)
	var protoReq CancelSessionRequest
	var metadata runtime.ServerMetadata

	newReader, berr := utilities.IOReaderFactory(req.Body)
	if berr != nil {
		__antithesis_instrumentation__.Notify(233693)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", berr)
	} else {
		__antithesis_instrumentation__.Notify(233694)
	}
	__antithesis_instrumentation__.Notify(233689)
	if err := marshaler.NewDecoder(newReader()).Decode(&protoReq); err != nil && func() bool {
		__antithesis_instrumentation__.Notify(233695)
		return err != io.EOF == true
	}() == true {
		__antithesis_instrumentation__.Notify(233696)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233697)
	}
	__antithesis_instrumentation__.Notify(233690)

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233698)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233699)
	}
	__antithesis_instrumentation__.Notify(233691)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233700)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233701)
	}
	__antithesis_instrumentation__.Notify(233692)

	msg, err := server.CancelSession(ctx, &protoReq)
	return msg, metadata, err

}

func request_Status_CancelLocalSession_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233702)
	var protoReq CancelSessionRequest
	var metadata runtime.ServerMetadata

	newReader, berr := utilities.IOReaderFactory(req.Body)
	if berr != nil {
		__antithesis_instrumentation__.Notify(233705)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", berr)
	} else {
		__antithesis_instrumentation__.Notify(233706)
	}
	__antithesis_instrumentation__.Notify(233703)
	if err := marshaler.NewDecoder(newReader()).Decode(&protoReq); err != nil && func() bool {
		__antithesis_instrumentation__.Notify(233707)
		return err != io.EOF == true
	}() == true {
		__antithesis_instrumentation__.Notify(233708)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233709)
	}
	__antithesis_instrumentation__.Notify(233704)

	msg, err := client.CancelLocalSession(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_CancelLocalSession_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233710)
	var protoReq CancelSessionRequest
	var metadata runtime.ServerMetadata

	newReader, berr := utilities.IOReaderFactory(req.Body)
	if berr != nil {
		__antithesis_instrumentation__.Notify(233713)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", berr)
	} else {
		__antithesis_instrumentation__.Notify(233714)
	}
	__antithesis_instrumentation__.Notify(233711)
	if err := marshaler.NewDecoder(newReader()).Decode(&protoReq); err != nil && func() bool {
		__antithesis_instrumentation__.Notify(233715)
		return err != io.EOF == true
	}() == true {
		__antithesis_instrumentation__.Notify(233716)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233717)
	}
	__antithesis_instrumentation__.Notify(233712)

	msg, err := server.CancelLocalSession(ctx, &protoReq)
	return msg, metadata, err

}

func request_Status_SpanStats_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233718)
	var protoReq SpanStatsRequest
	var metadata runtime.ServerMetadata

	newReader, berr := utilities.IOReaderFactory(req.Body)
	if berr != nil {
		__antithesis_instrumentation__.Notify(233721)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", berr)
	} else {
		__antithesis_instrumentation__.Notify(233722)
	}
	__antithesis_instrumentation__.Notify(233719)
	if err := marshaler.NewDecoder(newReader()).Decode(&protoReq); err != nil && func() bool {
		__antithesis_instrumentation__.Notify(233723)
		return err != io.EOF == true
	}() == true {
		__antithesis_instrumentation__.Notify(233724)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233725)
	}
	__antithesis_instrumentation__.Notify(233720)

	msg, err := client.SpanStats(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_SpanStats_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233726)
	var protoReq SpanStatsRequest
	var metadata runtime.ServerMetadata

	newReader, berr := utilities.IOReaderFactory(req.Body)
	if berr != nil {
		__antithesis_instrumentation__.Notify(233729)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", berr)
	} else {
		__antithesis_instrumentation__.Notify(233730)
	}
	__antithesis_instrumentation__.Notify(233727)
	if err := marshaler.NewDecoder(newReader()).Decode(&protoReq); err != nil && func() bool {
		__antithesis_instrumentation__.Notify(233731)
		return err != io.EOF == true
	}() == true {
		__antithesis_instrumentation__.Notify(233732)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233733)
	}
	__antithesis_instrumentation__.Notify(233728)

	msg, err := server.SpanStats(ctx, &protoReq)
	return msg, metadata, err

}

var (
	filter_Status_Stacks_0 = &utilities.DoubleArray{Encoding: map[string]int{"node_id": 0}, Base: []int{1, 1, 0}, Check: []int{0, 1, 2}}
)

func request_Status_Stacks_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233734)
	var protoReq StacksRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233739)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233740)
	}
	__antithesis_instrumentation__.Notify(233735)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233741)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233742)
	}
	__antithesis_instrumentation__.Notify(233736)

	if err := req.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(233743)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233744)
	}
	__antithesis_instrumentation__.Notify(233737)
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_Status_Stacks_0); err != nil {
		__antithesis_instrumentation__.Notify(233745)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233746)
	}
	__antithesis_instrumentation__.Notify(233738)

	msg, err := client.Stacks(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_Stacks_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233747)
	var protoReq StacksRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233752)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233753)
	}
	__antithesis_instrumentation__.Notify(233748)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233754)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233755)
	}
	__antithesis_instrumentation__.Notify(233749)

	if err := req.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(233756)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233757)
	}
	__antithesis_instrumentation__.Notify(233750)
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_Status_Stacks_0); err != nil {
		__antithesis_instrumentation__.Notify(233758)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233759)
	}
	__antithesis_instrumentation__.Notify(233751)

	msg, err := server.Stacks(ctx, &protoReq)
	return msg, metadata, err

}

var (
	filter_Status_Profile_0 = &utilities.DoubleArray{Encoding: map[string]int{"node_id": 0}, Base: []int{1, 1, 0}, Check: []int{0, 1, 2}}
)

func request_Status_Profile_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233760)
	var protoReq ProfileRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233765)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233766)
	}
	__antithesis_instrumentation__.Notify(233761)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233767)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233768)
	}
	__antithesis_instrumentation__.Notify(233762)

	if err := req.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(233769)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233770)
	}
	__antithesis_instrumentation__.Notify(233763)
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_Status_Profile_0); err != nil {
		__antithesis_instrumentation__.Notify(233771)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233772)
	}
	__antithesis_instrumentation__.Notify(233764)

	msg, err := client.Profile(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_Profile_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233773)
	var protoReq ProfileRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233778)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233779)
	}
	__antithesis_instrumentation__.Notify(233774)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233780)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233781)
	}
	__antithesis_instrumentation__.Notify(233775)

	if err := req.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(233782)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233783)
	}
	__antithesis_instrumentation__.Notify(233776)
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_Status_Profile_0); err != nil {
		__antithesis_instrumentation__.Notify(233784)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233785)
	}
	__antithesis_instrumentation__.Notify(233777)

	msg, err := server.Profile(ctx, &protoReq)
	return msg, metadata, err

}

func request_Status_Metrics_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233786)
	var protoReq MetricsRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233789)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233790)
	}
	__antithesis_instrumentation__.Notify(233787)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233791)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233792)
	}
	__antithesis_instrumentation__.Notify(233788)

	msg, err := client.Metrics(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_Metrics_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233793)
	var protoReq MetricsRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233796)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233797)
	}
	__antithesis_instrumentation__.Notify(233794)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233798)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233799)
	}
	__antithesis_instrumentation__.Notify(233795)

	msg, err := server.Metrics(ctx, &protoReq)
	return msg, metadata, err

}

var (
	filter_Status_GetFiles_0 = &utilities.DoubleArray{Encoding: map[string]int{"node_id": 0}, Base: []int{1, 1, 0}, Check: []int{0, 1, 2}}
)

func request_Status_GetFiles_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233800)
	var protoReq GetFilesRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233805)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233806)
	}
	__antithesis_instrumentation__.Notify(233801)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233807)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233808)
	}
	__antithesis_instrumentation__.Notify(233802)

	if err := req.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(233809)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233810)
	}
	__antithesis_instrumentation__.Notify(233803)
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_Status_GetFiles_0); err != nil {
		__antithesis_instrumentation__.Notify(233811)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233812)
	}
	__antithesis_instrumentation__.Notify(233804)

	msg, err := client.GetFiles(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_GetFiles_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233813)
	var protoReq GetFilesRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233818)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233819)
	}
	__antithesis_instrumentation__.Notify(233814)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233820)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233821)
	}
	__antithesis_instrumentation__.Notify(233815)

	if err := req.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(233822)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233823)
	}
	__antithesis_instrumentation__.Notify(233816)
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_Status_GetFiles_0); err != nil {
		__antithesis_instrumentation__.Notify(233824)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233825)
	}
	__antithesis_instrumentation__.Notify(233817)

	msg, err := server.GetFiles(ctx, &protoReq)
	return msg, metadata, err

}

func request_Status_LogFilesList_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233826)
	var protoReq LogFilesListRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233829)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233830)
	}
	__antithesis_instrumentation__.Notify(233827)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233831)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233832)
	}
	__antithesis_instrumentation__.Notify(233828)

	msg, err := client.LogFilesList(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_LogFilesList_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233833)
	var protoReq LogFilesListRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233836)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233837)
	}
	__antithesis_instrumentation__.Notify(233834)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233838)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233839)
	}
	__antithesis_instrumentation__.Notify(233835)

	msg, err := server.LogFilesList(ctx, &protoReq)
	return msg, metadata, err

}

var (
	filter_Status_LogFile_0 = &utilities.DoubleArray{Encoding: map[string]int{"node_id": 0, "file": 1}, Base: []int{1, 1, 2, 0, 0}, Check: []int{0, 1, 1, 2, 3}}
)

func request_Status_LogFile_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233840)
	var protoReq LogFileRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233847)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233848)
	}
	__antithesis_instrumentation__.Notify(233841)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233849)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233850)
	}
	__antithesis_instrumentation__.Notify(233842)

	val, ok = pathParams["file"]
	if !ok {
		__antithesis_instrumentation__.Notify(233851)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "file")
	} else {
		__antithesis_instrumentation__.Notify(233852)
	}
	__antithesis_instrumentation__.Notify(233843)

	protoReq.File, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233853)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "file", err)
	} else {
		__antithesis_instrumentation__.Notify(233854)
	}
	__antithesis_instrumentation__.Notify(233844)

	if err := req.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(233855)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233856)
	}
	__antithesis_instrumentation__.Notify(233845)
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_Status_LogFile_0); err != nil {
		__antithesis_instrumentation__.Notify(233857)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233858)
	}
	__antithesis_instrumentation__.Notify(233846)

	msg, err := client.LogFile(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_LogFile_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233859)
	var protoReq LogFileRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233866)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233867)
	}
	__antithesis_instrumentation__.Notify(233860)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233868)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233869)
	}
	__antithesis_instrumentation__.Notify(233861)

	val, ok = pathParams["file"]
	if !ok {
		__antithesis_instrumentation__.Notify(233870)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "file")
	} else {
		__antithesis_instrumentation__.Notify(233871)
	}
	__antithesis_instrumentation__.Notify(233862)

	protoReq.File, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233872)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "file", err)
	} else {
		__antithesis_instrumentation__.Notify(233873)
	}
	__antithesis_instrumentation__.Notify(233863)

	if err := req.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(233874)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233875)
	}
	__antithesis_instrumentation__.Notify(233864)
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_Status_LogFile_0); err != nil {
		__antithesis_instrumentation__.Notify(233876)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233877)
	}
	__antithesis_instrumentation__.Notify(233865)

	msg, err := server.LogFile(ctx, &protoReq)
	return msg, metadata, err

}

var (
	filter_Status_Logs_0 = &utilities.DoubleArray{Encoding: map[string]int{"node_id": 0}, Base: []int{1, 1, 0}, Check: []int{0, 1, 2}}
)

func request_Status_Logs_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233878)
	var protoReq LogsRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233883)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233884)
	}
	__antithesis_instrumentation__.Notify(233879)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233885)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233886)
	}
	__antithesis_instrumentation__.Notify(233880)

	if err := req.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(233887)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233888)
	}
	__antithesis_instrumentation__.Notify(233881)
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_Status_Logs_0); err != nil {
		__antithesis_instrumentation__.Notify(233889)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233890)
	}
	__antithesis_instrumentation__.Notify(233882)

	msg, err := client.Logs(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_Logs_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233891)
	var protoReq LogsRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233896)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233897)
	}
	__antithesis_instrumentation__.Notify(233892)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233898)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233899)
	}
	__antithesis_instrumentation__.Notify(233893)

	if err := req.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(233900)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233901)
	}
	__antithesis_instrumentation__.Notify(233894)
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_Status_Logs_0); err != nil {
		__antithesis_instrumentation__.Notify(233902)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233903)
	}
	__antithesis_instrumentation__.Notify(233895)

	msg, err := server.Logs(ctx, &protoReq)
	return msg, metadata, err

}

var (
	filter_Status_ProblemRanges_0 = &utilities.DoubleArray{Encoding: map[string]int{}, Base: []int(nil), Check: []int(nil)}
)

func request_Status_ProblemRanges_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233904)
	var protoReq ProblemRangesRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(233907)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233908)
	}
	__antithesis_instrumentation__.Notify(233905)
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_Status_ProblemRanges_0); err != nil {
		__antithesis_instrumentation__.Notify(233909)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233910)
	}
	__antithesis_instrumentation__.Notify(233906)

	msg, err := client.ProblemRanges(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_ProblemRanges_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233911)
	var protoReq ProblemRangesRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(233914)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233915)
	}
	__antithesis_instrumentation__.Notify(233912)
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_Status_ProblemRanges_0); err != nil {
		__antithesis_instrumentation__.Notify(233916)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233917)
	}
	__antithesis_instrumentation__.Notify(233913)

	msg, err := server.ProblemRanges(ctx, &protoReq)
	return msg, metadata, err

}

var (
	filter_Status_HotRanges_0 = &utilities.DoubleArray{Encoding: map[string]int{}, Base: []int(nil), Check: []int(nil)}
)

func request_Status_HotRanges_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233918)
	var protoReq HotRangesRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(233921)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233922)
	}
	__antithesis_instrumentation__.Notify(233919)
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_Status_HotRanges_0); err != nil {
		__antithesis_instrumentation__.Notify(233923)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233924)
	}
	__antithesis_instrumentation__.Notify(233920)

	msg, err := client.HotRanges(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_HotRanges_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233925)
	var protoReq HotRangesRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(233928)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233929)
	}
	__antithesis_instrumentation__.Notify(233926)
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_Status_HotRanges_0); err != nil {
		__antithesis_instrumentation__.Notify(233930)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233931)
	}
	__antithesis_instrumentation__.Notify(233927)

	msg, err := server.HotRanges(ctx, &protoReq)
	return msg, metadata, err

}

func request_Status_HotRangesV2_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233932)
	var protoReq HotRangesRequest
	var metadata runtime.ServerMetadata

	newReader, berr := utilities.IOReaderFactory(req.Body)
	if berr != nil {
		__antithesis_instrumentation__.Notify(233935)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", berr)
	} else {
		__antithesis_instrumentation__.Notify(233936)
	}
	__antithesis_instrumentation__.Notify(233933)
	if err := marshaler.NewDecoder(newReader()).Decode(&protoReq); err != nil && func() bool {
		__antithesis_instrumentation__.Notify(233937)
		return err != io.EOF == true
	}() == true {
		__antithesis_instrumentation__.Notify(233938)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233939)
	}
	__antithesis_instrumentation__.Notify(233934)

	msg, err := client.HotRangesV2(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_HotRangesV2_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233940)
	var protoReq HotRangesRequest
	var metadata runtime.ServerMetadata

	newReader, berr := utilities.IOReaderFactory(req.Body)
	if berr != nil {
		__antithesis_instrumentation__.Notify(233943)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", berr)
	} else {
		__antithesis_instrumentation__.Notify(233944)
	}
	__antithesis_instrumentation__.Notify(233941)
	if err := marshaler.NewDecoder(newReader()).Decode(&protoReq); err != nil && func() bool {
		__antithesis_instrumentation__.Notify(233945)
		return err != io.EOF == true
	}() == true {
		__antithesis_instrumentation__.Notify(233946)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233947)
	}
	__antithesis_instrumentation__.Notify(233942)

	msg, err := server.HotRangesV2(ctx, &protoReq)
	return msg, metadata, err

}

func request_Status_Range_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233948)
	var protoReq RangeRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["range_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233951)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "range_id")
	} else {
		__antithesis_instrumentation__.Notify(233952)
	}
	__antithesis_instrumentation__.Notify(233949)

	protoReq.RangeId, err = runtime.Int64(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233953)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "range_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233954)
	}
	__antithesis_instrumentation__.Notify(233950)

	msg, err := client.Range(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_Range_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233955)
	var protoReq RangeRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["range_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233958)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "range_id")
	} else {
		__antithesis_instrumentation__.Notify(233959)
	}
	__antithesis_instrumentation__.Notify(233956)

	protoReq.RangeId, err = runtime.Int64(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233960)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "range_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233961)
	}
	__antithesis_instrumentation__.Notify(233957)

	msg, err := server.Range(ctx, &protoReq)
	return msg, metadata, err

}

func request_Status_Diagnostics_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233962)
	var protoReq DiagnosticsRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233965)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233966)
	}
	__antithesis_instrumentation__.Notify(233963)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233967)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233968)
	}
	__antithesis_instrumentation__.Notify(233964)

	msg, err := client.Diagnostics(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_Diagnostics_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233969)
	var protoReq DiagnosticsRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233972)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233973)
	}
	__antithesis_instrumentation__.Notify(233970)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233974)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233975)
	}
	__antithesis_instrumentation__.Notify(233971)

	msg, err := server.Diagnostics(ctx, &protoReq)
	return msg, metadata, err

}

func request_Status_Stores_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233976)
	var protoReq StoresRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233979)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233980)
	}
	__antithesis_instrumentation__.Notify(233977)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233981)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233982)
	}
	__antithesis_instrumentation__.Notify(233978)

	msg, err := client.Stores(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_Stores_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233983)
	var protoReq StoresRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(233986)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(233987)
	}
	__antithesis_instrumentation__.Notify(233984)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(233988)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(233989)
	}
	__antithesis_instrumentation__.Notify(233985)

	msg, err := server.Stores(ctx, &protoReq)
	return msg, metadata, err

}

var (
	filter_Status_Statements_0 = &utilities.DoubleArray{Encoding: map[string]int{}, Base: []int(nil), Check: []int(nil)}
)

func request_Status_Statements_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233990)
	var protoReq StatementsRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(233993)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233994)
	}
	__antithesis_instrumentation__.Notify(233991)
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_Status_Statements_0); err != nil {
		__antithesis_instrumentation__.Notify(233995)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(233996)
	}
	__antithesis_instrumentation__.Notify(233992)

	msg, err := client.Statements(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_Statements_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(233997)
	var protoReq StatementsRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(234000)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(234001)
	}
	__antithesis_instrumentation__.Notify(233998)
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_Status_Statements_0); err != nil {
		__antithesis_instrumentation__.Notify(234002)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(234003)
	}
	__antithesis_instrumentation__.Notify(233999)

	msg, err := server.Statements(ctx, &protoReq)
	return msg, metadata, err

}

var (
	filter_Status_CombinedStatementStats_0 = &utilities.DoubleArray{Encoding: map[string]int{}, Base: []int(nil), Check: []int(nil)}
)

func request_Status_CombinedStatementStats_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(234004)
	var protoReq CombinedStatementsStatsRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(234007)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(234008)
	}
	__antithesis_instrumentation__.Notify(234005)
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_Status_CombinedStatementStats_0); err != nil {
		__antithesis_instrumentation__.Notify(234009)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(234010)
	}
	__antithesis_instrumentation__.Notify(234006)

	msg, err := client.CombinedStatementStats(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_CombinedStatementStats_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(234011)
	var protoReq CombinedStatementsStatsRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(234014)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(234015)
	}
	__antithesis_instrumentation__.Notify(234012)
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_Status_CombinedStatementStats_0); err != nil {
		__antithesis_instrumentation__.Notify(234016)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(234017)
	}
	__antithesis_instrumentation__.Notify(234013)

	msg, err := server.CombinedStatementStats(ctx, &protoReq)
	return msg, metadata, err

}

var (
	filter_Status_StatementDetails_0 = &utilities.DoubleArray{Encoding: map[string]int{"fingerprint_id": 0}, Base: []int{1, 1, 0}, Check: []int{0, 1, 2}}
)

func request_Status_StatementDetails_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(234018)
	var protoReq StatementDetailsRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["fingerprint_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(234023)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "fingerprint_id")
	} else {
		__antithesis_instrumentation__.Notify(234024)
	}
	__antithesis_instrumentation__.Notify(234019)

	protoReq.FingerprintId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(234025)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "fingerprint_id", err)
	} else {
		__antithesis_instrumentation__.Notify(234026)
	}
	__antithesis_instrumentation__.Notify(234020)

	if err := req.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(234027)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(234028)
	}
	__antithesis_instrumentation__.Notify(234021)
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_Status_StatementDetails_0); err != nil {
		__antithesis_instrumentation__.Notify(234029)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(234030)
	}
	__antithesis_instrumentation__.Notify(234022)

	msg, err := client.StatementDetails(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_StatementDetails_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(234031)
	var protoReq StatementDetailsRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["fingerprint_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(234036)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "fingerprint_id")
	} else {
		__antithesis_instrumentation__.Notify(234037)
	}
	__antithesis_instrumentation__.Notify(234032)

	protoReq.FingerprintId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(234038)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "fingerprint_id", err)
	} else {
		__antithesis_instrumentation__.Notify(234039)
	}
	__antithesis_instrumentation__.Notify(234033)

	if err := req.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(234040)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(234041)
	}
	__antithesis_instrumentation__.Notify(234034)
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_Status_StatementDetails_0); err != nil {
		__antithesis_instrumentation__.Notify(234042)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(234043)
	}
	__antithesis_instrumentation__.Notify(234035)

	msg, err := server.StatementDetails(ctx, &protoReq)
	return msg, metadata, err

}

func request_Status_CreateStatementDiagnosticsReport_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(234044)
	var protoReq CreateStatementDiagnosticsReportRequest
	var metadata runtime.ServerMetadata

	newReader, berr := utilities.IOReaderFactory(req.Body)
	if berr != nil {
		__antithesis_instrumentation__.Notify(234047)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", berr)
	} else {
		__antithesis_instrumentation__.Notify(234048)
	}
	__antithesis_instrumentation__.Notify(234045)
	if err := marshaler.NewDecoder(newReader()).Decode(&protoReq); err != nil && func() bool {
		__antithesis_instrumentation__.Notify(234049)
		return err != io.EOF == true
	}() == true {
		__antithesis_instrumentation__.Notify(234050)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(234051)
	}
	__antithesis_instrumentation__.Notify(234046)

	msg, err := client.CreateStatementDiagnosticsReport(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_CreateStatementDiagnosticsReport_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(234052)
	var protoReq CreateStatementDiagnosticsReportRequest
	var metadata runtime.ServerMetadata

	newReader, berr := utilities.IOReaderFactory(req.Body)
	if berr != nil {
		__antithesis_instrumentation__.Notify(234055)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", berr)
	} else {
		__antithesis_instrumentation__.Notify(234056)
	}
	__antithesis_instrumentation__.Notify(234053)
	if err := marshaler.NewDecoder(newReader()).Decode(&protoReq); err != nil && func() bool {
		__antithesis_instrumentation__.Notify(234057)
		return err != io.EOF == true
	}() == true {
		__antithesis_instrumentation__.Notify(234058)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(234059)
	}
	__antithesis_instrumentation__.Notify(234054)

	msg, err := server.CreateStatementDiagnosticsReport(ctx, &protoReq)
	return msg, metadata, err

}

func request_Status_CancelStatementDiagnosticsReport_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(234060)
	var protoReq CancelStatementDiagnosticsReportRequest
	var metadata runtime.ServerMetadata

	newReader, berr := utilities.IOReaderFactory(req.Body)
	if berr != nil {
		__antithesis_instrumentation__.Notify(234063)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", berr)
	} else {
		__antithesis_instrumentation__.Notify(234064)
	}
	__antithesis_instrumentation__.Notify(234061)
	if err := marshaler.NewDecoder(newReader()).Decode(&protoReq); err != nil && func() bool {
		__antithesis_instrumentation__.Notify(234065)
		return err != io.EOF == true
	}() == true {
		__antithesis_instrumentation__.Notify(234066)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(234067)
	}
	__antithesis_instrumentation__.Notify(234062)

	msg, err := client.CancelStatementDiagnosticsReport(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_CancelStatementDiagnosticsReport_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(234068)
	var protoReq CancelStatementDiagnosticsReportRequest
	var metadata runtime.ServerMetadata

	newReader, berr := utilities.IOReaderFactory(req.Body)
	if berr != nil {
		__antithesis_instrumentation__.Notify(234071)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", berr)
	} else {
		__antithesis_instrumentation__.Notify(234072)
	}
	__antithesis_instrumentation__.Notify(234069)
	if err := marshaler.NewDecoder(newReader()).Decode(&protoReq); err != nil && func() bool {
		__antithesis_instrumentation__.Notify(234073)
		return err != io.EOF == true
	}() == true {
		__antithesis_instrumentation__.Notify(234074)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(234075)
	}
	__antithesis_instrumentation__.Notify(234070)

	msg, err := server.CancelStatementDiagnosticsReport(ctx, &protoReq)
	return msg, metadata, err

}

func request_Status_StatementDiagnosticsRequests_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(234076)
	var protoReq StatementDiagnosticsReportsRequest
	var metadata runtime.ServerMetadata

	msg, err := client.StatementDiagnosticsRequests(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_StatementDiagnosticsRequests_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(234077)
	var protoReq StatementDiagnosticsReportsRequest
	var metadata runtime.ServerMetadata

	msg, err := server.StatementDiagnosticsRequests(ctx, &protoReq)
	return msg, metadata, err

}

func request_Status_StatementDiagnostics_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(234078)
	var protoReq StatementDiagnosticsRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["statement_diagnostics_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(234081)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "statement_diagnostics_id")
	} else {
		__antithesis_instrumentation__.Notify(234082)
	}
	__antithesis_instrumentation__.Notify(234079)

	protoReq.StatementDiagnosticsId, err = runtime.Int64(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(234083)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "statement_diagnostics_id", err)
	} else {
		__antithesis_instrumentation__.Notify(234084)
	}
	__antithesis_instrumentation__.Notify(234080)

	msg, err := client.StatementDiagnostics(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_StatementDiagnostics_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(234085)
	var protoReq StatementDiagnosticsRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["statement_diagnostics_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(234088)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "statement_diagnostics_id")
	} else {
		__antithesis_instrumentation__.Notify(234089)
	}
	__antithesis_instrumentation__.Notify(234086)

	protoReq.StatementDiagnosticsId, err = runtime.Int64(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(234090)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "statement_diagnostics_id", err)
	} else {
		__antithesis_instrumentation__.Notify(234091)
	}
	__antithesis_instrumentation__.Notify(234087)

	msg, err := server.StatementDiagnostics(ctx, &protoReq)
	return msg, metadata, err

}

func request_Status_JobRegistryStatus_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(234092)
	var protoReq JobRegistryStatusRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(234095)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(234096)
	}
	__antithesis_instrumentation__.Notify(234093)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(234097)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(234098)
	}
	__antithesis_instrumentation__.Notify(234094)

	msg, err := client.JobRegistryStatus(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_JobRegistryStatus_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(234099)
	var protoReq JobRegistryStatusRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["node_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(234102)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "node_id")
	} else {
		__antithesis_instrumentation__.Notify(234103)
	}
	__antithesis_instrumentation__.Notify(234100)

	protoReq.NodeId, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(234104)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "node_id", err)
	} else {
		__antithesis_instrumentation__.Notify(234105)
	}
	__antithesis_instrumentation__.Notify(234101)

	msg, err := server.JobRegistryStatus(ctx, &protoReq)
	return msg, metadata, err

}

func request_Status_JobStatus_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(234106)
	var protoReq JobStatusRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["job_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(234109)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "job_id")
	} else {
		__antithesis_instrumentation__.Notify(234110)
	}
	__antithesis_instrumentation__.Notify(234107)

	protoReq.JobId, err = runtime.Int64(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(234111)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "job_id", err)
	} else {
		__antithesis_instrumentation__.Notify(234112)
	}
	__antithesis_instrumentation__.Notify(234108)

	msg, err := client.JobStatus(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_JobStatus_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(234113)
	var protoReq JobStatusRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["job_id"]
	if !ok {
		__antithesis_instrumentation__.Notify(234116)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "job_id")
	} else {
		__antithesis_instrumentation__.Notify(234117)
	}
	__antithesis_instrumentation__.Notify(234114)

	protoReq.JobId, err = runtime.Int64(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(234118)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "job_id", err)
	} else {
		__antithesis_instrumentation__.Notify(234119)
	}
	__antithesis_instrumentation__.Notify(234115)

	msg, err := server.JobStatus(ctx, &protoReq)
	return msg, metadata, err

}

func request_Status_ResetSQLStats_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(234120)
	var protoReq ResetSQLStatsRequest
	var metadata runtime.ServerMetadata

	newReader, berr := utilities.IOReaderFactory(req.Body)
	if berr != nil {
		__antithesis_instrumentation__.Notify(234123)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", berr)
	} else {
		__antithesis_instrumentation__.Notify(234124)
	}
	__antithesis_instrumentation__.Notify(234121)
	if err := marshaler.NewDecoder(newReader()).Decode(&protoReq); err != nil && func() bool {
		__antithesis_instrumentation__.Notify(234125)
		return err != io.EOF == true
	}() == true {
		__antithesis_instrumentation__.Notify(234126)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(234127)
	}
	__antithesis_instrumentation__.Notify(234122)

	msg, err := client.ResetSQLStats(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_ResetSQLStats_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(234128)
	var protoReq ResetSQLStatsRequest
	var metadata runtime.ServerMetadata

	newReader, berr := utilities.IOReaderFactory(req.Body)
	if berr != nil {
		__antithesis_instrumentation__.Notify(234131)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", berr)
	} else {
		__antithesis_instrumentation__.Notify(234132)
	}
	__antithesis_instrumentation__.Notify(234129)
	if err := marshaler.NewDecoder(newReader()).Decode(&protoReq); err != nil && func() bool {
		__antithesis_instrumentation__.Notify(234133)
		return err != io.EOF == true
	}() == true {
		__antithesis_instrumentation__.Notify(234134)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(234135)
	}
	__antithesis_instrumentation__.Notify(234130)

	msg, err := server.ResetSQLStats(ctx, &protoReq)
	return msg, metadata, err

}

var (
	filter_Status_IndexUsageStatistics_0 = &utilities.DoubleArray{Encoding: map[string]int{}, Base: []int(nil), Check: []int(nil)}
)

func request_Status_IndexUsageStatistics_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(234136)
	var protoReq IndexUsageStatisticsRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(234139)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(234140)
	}
	__antithesis_instrumentation__.Notify(234137)
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_Status_IndexUsageStatistics_0); err != nil {
		__antithesis_instrumentation__.Notify(234141)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(234142)
	}
	__antithesis_instrumentation__.Notify(234138)

	msg, err := client.IndexUsageStatistics(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_IndexUsageStatistics_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(234143)
	var protoReq IndexUsageStatisticsRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(234146)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(234147)
	}
	__antithesis_instrumentation__.Notify(234144)
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_Status_IndexUsageStatistics_0); err != nil {
		__antithesis_instrumentation__.Notify(234148)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(234149)
	}
	__antithesis_instrumentation__.Notify(234145)

	msg, err := server.IndexUsageStatistics(ctx, &protoReq)
	return msg, metadata, err

}

func request_Status_ResetIndexUsageStats_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(234150)
	var protoReq ResetIndexUsageStatsRequest
	var metadata runtime.ServerMetadata

	newReader, berr := utilities.IOReaderFactory(req.Body)
	if berr != nil {
		__antithesis_instrumentation__.Notify(234153)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", berr)
	} else {
		__antithesis_instrumentation__.Notify(234154)
	}
	__antithesis_instrumentation__.Notify(234151)
	if err := marshaler.NewDecoder(newReader()).Decode(&protoReq); err != nil && func() bool {
		__antithesis_instrumentation__.Notify(234155)
		return err != io.EOF == true
	}() == true {
		__antithesis_instrumentation__.Notify(234156)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(234157)
	}
	__antithesis_instrumentation__.Notify(234152)

	msg, err := client.ResetIndexUsageStats(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_ResetIndexUsageStats_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(234158)
	var protoReq ResetIndexUsageStatsRequest
	var metadata runtime.ServerMetadata

	newReader, berr := utilities.IOReaderFactory(req.Body)
	if berr != nil {
		__antithesis_instrumentation__.Notify(234161)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", berr)
	} else {
		__antithesis_instrumentation__.Notify(234162)
	}
	__antithesis_instrumentation__.Notify(234159)
	if err := marshaler.NewDecoder(newReader()).Decode(&protoReq); err != nil && func() bool {
		__antithesis_instrumentation__.Notify(234163)
		return err != io.EOF == true
	}() == true {
		__antithesis_instrumentation__.Notify(234164)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(234165)
	}
	__antithesis_instrumentation__.Notify(234160)

	msg, err := server.ResetIndexUsageStats(ctx, &protoReq)
	return msg, metadata, err

}

func request_Status_TableIndexStats_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(234166)
	var protoReq TableIndexStatsRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["database"]
	if !ok {
		__antithesis_instrumentation__.Notify(234171)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "database")
	} else {
		__antithesis_instrumentation__.Notify(234172)
	}
	__antithesis_instrumentation__.Notify(234167)

	protoReq.Database, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(234173)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "database", err)
	} else {
		__antithesis_instrumentation__.Notify(234174)
	}
	__antithesis_instrumentation__.Notify(234168)

	val, ok = pathParams["table"]
	if !ok {
		__antithesis_instrumentation__.Notify(234175)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "table")
	} else {
		__antithesis_instrumentation__.Notify(234176)
	}
	__antithesis_instrumentation__.Notify(234169)

	protoReq.Table, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(234177)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "table", err)
	} else {
		__antithesis_instrumentation__.Notify(234178)
	}
	__antithesis_instrumentation__.Notify(234170)

	msg, err := client.TableIndexStats(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_TableIndexStats_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(234179)
	var protoReq TableIndexStatsRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["database"]
	if !ok {
		__antithesis_instrumentation__.Notify(234184)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "database")
	} else {
		__antithesis_instrumentation__.Notify(234185)
	}
	__antithesis_instrumentation__.Notify(234180)

	protoReq.Database, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(234186)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "database", err)
	} else {
		__antithesis_instrumentation__.Notify(234187)
	}
	__antithesis_instrumentation__.Notify(234181)

	val, ok = pathParams["table"]
	if !ok {
		__antithesis_instrumentation__.Notify(234188)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "table")
	} else {
		__antithesis_instrumentation__.Notify(234189)
	}
	__antithesis_instrumentation__.Notify(234182)

	protoReq.Table, err = runtime.String(val)

	if err != nil {
		__antithesis_instrumentation__.Notify(234190)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "table", err)
	} else {
		__antithesis_instrumentation__.Notify(234191)
	}
	__antithesis_instrumentation__.Notify(234183)

	msg, err := server.TableIndexStats(ctx, &protoReq)
	return msg, metadata, err

}

func request_Status_UserSQLRoles_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(234192)
	var protoReq UserSQLRolesRequest
	var metadata runtime.ServerMetadata

	msg, err := client.UserSQLRoles(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_UserSQLRoles_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(234193)
	var protoReq UserSQLRolesRequest
	var metadata runtime.ServerMetadata

	msg, err := server.UserSQLRoles(ctx, &protoReq)
	return msg, metadata, err

}

var (
	filter_Status_TransactionContentionEvents_0 = &utilities.DoubleArray{Encoding: map[string]int{}, Base: []int(nil), Check: []int(nil)}
)

func request_Status_TransactionContentionEvents_0(ctx context.Context, marshaler runtime.Marshaler, client StatusClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(234194)
	var protoReq TransactionContentionEventsRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(234197)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(234198)
	}
	__antithesis_instrumentation__.Notify(234195)
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_Status_TransactionContentionEvents_0); err != nil {
		__antithesis_instrumentation__.Notify(234199)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(234200)
	}
	__antithesis_instrumentation__.Notify(234196)

	msg, err := client.TransactionContentionEvents(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_Status_TransactionContentionEvents_0(ctx context.Context, marshaler runtime.Marshaler, server StatusServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	__antithesis_instrumentation__.Notify(234201)
	var protoReq TransactionContentionEventsRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(234204)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(234205)
	}
	__antithesis_instrumentation__.Notify(234202)
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_Status_TransactionContentionEvents_0); err != nil {
		__antithesis_instrumentation__.Notify(234206)
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(234207)
	}
	__antithesis_instrumentation__.Notify(234203)

	msg, err := server.TransactionContentionEvents(ctx, &protoReq)
	return msg, metadata, err

}

func RegisterStatusHandlerServer(ctx context.Context, mux *runtime.ServeMux, server StatusServer) error {
	__antithesis_instrumentation__.Notify(234208)

	mux.Handle("GET", pattern_Status_Certificates_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234261)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234264)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234265)
		}
		__antithesis_instrumentation__.Notify(234262)
		resp, md, err := local_request_Status_Certificates_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234266)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234267)
		}
		__antithesis_instrumentation__.Notify(234263)

		forward_Status_Certificates_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234209)

	mux.Handle("GET", pattern_Status_Details_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234268)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234271)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234272)
		}
		__antithesis_instrumentation__.Notify(234269)
		resp, md, err := local_request_Status_Details_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234273)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234274)
		}
		__antithesis_instrumentation__.Notify(234270)

		forward_Status_Details_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234210)

	mux.Handle("GET", pattern_Status_Nodes_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234275)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234278)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234279)
		}
		__antithesis_instrumentation__.Notify(234276)
		resp, md, err := local_request_Status_Nodes_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234280)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234281)
		}
		__antithesis_instrumentation__.Notify(234277)

		forward_Status_Nodes_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234211)

	mux.Handle("GET", pattern_Status_Node_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234282)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234285)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234286)
		}
		__antithesis_instrumentation__.Notify(234283)
		resp, md, err := local_request_Status_Node_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234287)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234288)
		}
		__antithesis_instrumentation__.Notify(234284)

		forward_Status_Node_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234212)

	mux.Handle("GET", pattern_Status_NodesUI_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234289)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234292)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234293)
		}
		__antithesis_instrumentation__.Notify(234290)
		resp, md, err := local_request_Status_NodesUI_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234294)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234295)
		}
		__antithesis_instrumentation__.Notify(234291)

		forward_Status_NodesUI_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234213)

	mux.Handle("GET", pattern_Status_NodeUI_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234296)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234299)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234300)
		}
		__antithesis_instrumentation__.Notify(234297)
		resp, md, err := local_request_Status_NodeUI_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234301)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234302)
		}
		__antithesis_instrumentation__.Notify(234298)

		forward_Status_NodeUI_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234214)

	mux.Handle("GET", pattern_Status_RaftDebug_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234303)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234306)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234307)
		}
		__antithesis_instrumentation__.Notify(234304)
		resp, md, err := local_request_Status_RaftDebug_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234308)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234309)
		}
		__antithesis_instrumentation__.Notify(234305)

		forward_Status_RaftDebug_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234215)

	mux.Handle("GET", pattern_Status_Ranges_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234310)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234313)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234314)
		}
		__antithesis_instrumentation__.Notify(234311)
		resp, md, err := local_request_Status_Ranges_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234315)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234316)
		}
		__antithesis_instrumentation__.Notify(234312)

		forward_Status_Ranges_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234216)

	mux.Handle("GET", pattern_Status_TenantRanges_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234317)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234320)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234321)
		}
		__antithesis_instrumentation__.Notify(234318)
		resp, md, err := local_request_Status_TenantRanges_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234322)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234323)
		}
		__antithesis_instrumentation__.Notify(234319)

		forward_Status_TenantRanges_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234217)

	mux.Handle("GET", pattern_Status_Gossip_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234324)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234327)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234328)
		}
		__antithesis_instrumentation__.Notify(234325)
		resp, md, err := local_request_Status_Gossip_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234329)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234330)
		}
		__antithesis_instrumentation__.Notify(234326)

		forward_Status_Gossip_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234218)

	mux.Handle("GET", pattern_Status_EngineStats_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234331)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234334)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234335)
		}
		__antithesis_instrumentation__.Notify(234332)
		resp, md, err := local_request_Status_EngineStats_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234336)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234337)
		}
		__antithesis_instrumentation__.Notify(234333)

		forward_Status_EngineStats_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234219)

	mux.Handle("GET", pattern_Status_Allocator_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234338)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234341)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234342)
		}
		__antithesis_instrumentation__.Notify(234339)
		resp, md, err := local_request_Status_Allocator_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234343)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234344)
		}
		__antithesis_instrumentation__.Notify(234340)

		forward_Status_Allocator_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234220)

	mux.Handle("GET", pattern_Status_AllocatorRange_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234345)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234348)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234349)
		}
		__antithesis_instrumentation__.Notify(234346)
		resp, md, err := local_request_Status_AllocatorRange_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234350)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234351)
		}
		__antithesis_instrumentation__.Notify(234347)

		forward_Status_AllocatorRange_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234221)

	mux.Handle("GET", pattern_Status_ListSessions_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234352)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234355)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234356)
		}
		__antithesis_instrumentation__.Notify(234353)
		resp, md, err := local_request_Status_ListSessions_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234357)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234358)
		}
		__antithesis_instrumentation__.Notify(234354)

		forward_Status_ListSessions_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234222)

	mux.Handle("GET", pattern_Status_ListLocalSessions_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234359)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234362)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234363)
		}
		__antithesis_instrumentation__.Notify(234360)
		resp, md, err := local_request_Status_ListLocalSessions_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234364)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234365)
		}
		__antithesis_instrumentation__.Notify(234361)

		forward_Status_ListLocalSessions_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234223)

	mux.Handle("POST", pattern_Status_CancelQuery_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234366)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234369)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234370)
		}
		__antithesis_instrumentation__.Notify(234367)
		resp, md, err := local_request_Status_CancelQuery_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234371)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234372)
		}
		__antithesis_instrumentation__.Notify(234368)

		forward_Status_CancelQuery_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234224)

	mux.Handle("POST", pattern_Status_CancelLocalQuery_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234373)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234376)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234377)
		}
		__antithesis_instrumentation__.Notify(234374)
		resp, md, err := local_request_Status_CancelLocalQuery_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234378)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234379)
		}
		__antithesis_instrumentation__.Notify(234375)

		forward_Status_CancelLocalQuery_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234225)

	mux.Handle("GET", pattern_Status_ListContentionEvents_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234380)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234383)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234384)
		}
		__antithesis_instrumentation__.Notify(234381)
		resp, md, err := local_request_Status_ListContentionEvents_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234385)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234386)
		}
		__antithesis_instrumentation__.Notify(234382)

		forward_Status_ListContentionEvents_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234226)

	mux.Handle("GET", pattern_Status_ListLocalContentionEvents_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234387)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234390)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234391)
		}
		__antithesis_instrumentation__.Notify(234388)
		resp, md, err := local_request_Status_ListLocalContentionEvents_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234392)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234393)
		}
		__antithesis_instrumentation__.Notify(234389)

		forward_Status_ListLocalContentionEvents_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234227)

	mux.Handle("GET", pattern_Status_ListDistSQLFlows_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234394)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234397)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234398)
		}
		__antithesis_instrumentation__.Notify(234395)
		resp, md, err := local_request_Status_ListDistSQLFlows_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234399)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234400)
		}
		__antithesis_instrumentation__.Notify(234396)

		forward_Status_ListDistSQLFlows_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234228)

	mux.Handle("GET", pattern_Status_ListLocalDistSQLFlows_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234401)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234404)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234405)
		}
		__antithesis_instrumentation__.Notify(234402)
		resp, md, err := local_request_Status_ListLocalDistSQLFlows_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234406)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234407)
		}
		__antithesis_instrumentation__.Notify(234403)

		forward_Status_ListLocalDistSQLFlows_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234229)

	mux.Handle("POST", pattern_Status_CancelSession_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234408)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234411)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234412)
		}
		__antithesis_instrumentation__.Notify(234409)
		resp, md, err := local_request_Status_CancelSession_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234413)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234414)
		}
		__antithesis_instrumentation__.Notify(234410)

		forward_Status_CancelSession_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234230)

	mux.Handle("POST", pattern_Status_CancelLocalSession_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234415)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234418)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234419)
		}
		__antithesis_instrumentation__.Notify(234416)
		resp, md, err := local_request_Status_CancelLocalSession_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234420)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234421)
		}
		__antithesis_instrumentation__.Notify(234417)

		forward_Status_CancelLocalSession_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234231)

	mux.Handle("POST", pattern_Status_SpanStats_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234422)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234425)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234426)
		}
		__antithesis_instrumentation__.Notify(234423)
		resp, md, err := local_request_Status_SpanStats_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234427)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234428)
		}
		__antithesis_instrumentation__.Notify(234424)

		forward_Status_SpanStats_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234232)

	mux.Handle("GET", pattern_Status_Stacks_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234429)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234432)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234433)
		}
		__antithesis_instrumentation__.Notify(234430)
		resp, md, err := local_request_Status_Stacks_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234434)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234435)
		}
		__antithesis_instrumentation__.Notify(234431)

		forward_Status_Stacks_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234233)

	mux.Handle("GET", pattern_Status_Profile_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234436)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234439)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234440)
		}
		__antithesis_instrumentation__.Notify(234437)
		resp, md, err := local_request_Status_Profile_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234441)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234442)
		}
		__antithesis_instrumentation__.Notify(234438)

		forward_Status_Profile_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234234)

	mux.Handle("GET", pattern_Status_Metrics_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234443)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234446)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234447)
		}
		__antithesis_instrumentation__.Notify(234444)
		resp, md, err := local_request_Status_Metrics_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234448)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234449)
		}
		__antithesis_instrumentation__.Notify(234445)

		forward_Status_Metrics_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234235)

	mux.Handle("GET", pattern_Status_GetFiles_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234450)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234453)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234454)
		}
		__antithesis_instrumentation__.Notify(234451)
		resp, md, err := local_request_Status_GetFiles_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234455)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234456)
		}
		__antithesis_instrumentation__.Notify(234452)

		forward_Status_GetFiles_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234236)

	mux.Handle("GET", pattern_Status_LogFilesList_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234457)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234460)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234461)
		}
		__antithesis_instrumentation__.Notify(234458)
		resp, md, err := local_request_Status_LogFilesList_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234462)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234463)
		}
		__antithesis_instrumentation__.Notify(234459)

		forward_Status_LogFilesList_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234237)

	mux.Handle("GET", pattern_Status_LogFile_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234464)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234467)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234468)
		}
		__antithesis_instrumentation__.Notify(234465)
		resp, md, err := local_request_Status_LogFile_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234469)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234470)
		}
		__antithesis_instrumentation__.Notify(234466)

		forward_Status_LogFile_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234238)

	mux.Handle("GET", pattern_Status_Logs_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234471)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234474)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234475)
		}
		__antithesis_instrumentation__.Notify(234472)
		resp, md, err := local_request_Status_Logs_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234476)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234477)
		}
		__antithesis_instrumentation__.Notify(234473)

		forward_Status_Logs_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234239)

	mux.Handle("GET", pattern_Status_ProblemRanges_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234478)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234481)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234482)
		}
		__antithesis_instrumentation__.Notify(234479)
		resp, md, err := local_request_Status_ProblemRanges_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234483)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234484)
		}
		__antithesis_instrumentation__.Notify(234480)

		forward_Status_ProblemRanges_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234240)

	mux.Handle("GET", pattern_Status_HotRanges_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234485)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234488)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234489)
		}
		__antithesis_instrumentation__.Notify(234486)
		resp, md, err := local_request_Status_HotRanges_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234490)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234491)
		}
		__antithesis_instrumentation__.Notify(234487)

		forward_Status_HotRanges_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234241)

	mux.Handle("POST", pattern_Status_HotRangesV2_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234492)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234495)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234496)
		}
		__antithesis_instrumentation__.Notify(234493)
		resp, md, err := local_request_Status_HotRangesV2_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234497)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234498)
		}
		__antithesis_instrumentation__.Notify(234494)

		forward_Status_HotRangesV2_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234242)

	mux.Handle("GET", pattern_Status_Range_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234499)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234502)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234503)
		}
		__antithesis_instrumentation__.Notify(234500)
		resp, md, err := local_request_Status_Range_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234504)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234505)
		}
		__antithesis_instrumentation__.Notify(234501)

		forward_Status_Range_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234243)

	mux.Handle("GET", pattern_Status_Diagnostics_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234506)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234509)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234510)
		}
		__antithesis_instrumentation__.Notify(234507)
		resp, md, err := local_request_Status_Diagnostics_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234511)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234512)
		}
		__antithesis_instrumentation__.Notify(234508)

		forward_Status_Diagnostics_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234244)

	mux.Handle("GET", pattern_Status_Stores_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234513)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234516)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234517)
		}
		__antithesis_instrumentation__.Notify(234514)
		resp, md, err := local_request_Status_Stores_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234518)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234519)
		}
		__antithesis_instrumentation__.Notify(234515)

		forward_Status_Stores_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234245)

	mux.Handle("GET", pattern_Status_Statements_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234520)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234523)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234524)
		}
		__antithesis_instrumentation__.Notify(234521)
		resp, md, err := local_request_Status_Statements_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234525)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234526)
		}
		__antithesis_instrumentation__.Notify(234522)

		forward_Status_Statements_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234246)

	mux.Handle("GET", pattern_Status_CombinedStatementStats_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234527)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234530)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234531)
		}
		__antithesis_instrumentation__.Notify(234528)
		resp, md, err := local_request_Status_CombinedStatementStats_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234532)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234533)
		}
		__antithesis_instrumentation__.Notify(234529)

		forward_Status_CombinedStatementStats_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234247)

	mux.Handle("GET", pattern_Status_StatementDetails_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234534)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234537)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234538)
		}
		__antithesis_instrumentation__.Notify(234535)
		resp, md, err := local_request_Status_StatementDetails_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234539)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234540)
		}
		__antithesis_instrumentation__.Notify(234536)

		forward_Status_StatementDetails_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234248)

	mux.Handle("POST", pattern_Status_CreateStatementDiagnosticsReport_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234541)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234544)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234545)
		}
		__antithesis_instrumentation__.Notify(234542)
		resp, md, err := local_request_Status_CreateStatementDiagnosticsReport_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234546)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234547)
		}
		__antithesis_instrumentation__.Notify(234543)

		forward_Status_CreateStatementDiagnosticsReport_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234249)

	mux.Handle("POST", pattern_Status_CancelStatementDiagnosticsReport_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234548)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234551)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234552)
		}
		__antithesis_instrumentation__.Notify(234549)
		resp, md, err := local_request_Status_CancelStatementDiagnosticsReport_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234553)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234554)
		}
		__antithesis_instrumentation__.Notify(234550)

		forward_Status_CancelStatementDiagnosticsReport_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234250)

	mux.Handle("GET", pattern_Status_StatementDiagnosticsRequests_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234555)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234558)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234559)
		}
		__antithesis_instrumentation__.Notify(234556)
		resp, md, err := local_request_Status_StatementDiagnosticsRequests_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234560)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234561)
		}
		__antithesis_instrumentation__.Notify(234557)

		forward_Status_StatementDiagnosticsRequests_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234251)

	mux.Handle("GET", pattern_Status_StatementDiagnostics_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234562)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234565)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234566)
		}
		__antithesis_instrumentation__.Notify(234563)
		resp, md, err := local_request_Status_StatementDiagnostics_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234567)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234568)
		}
		__antithesis_instrumentation__.Notify(234564)

		forward_Status_StatementDiagnostics_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234252)

	mux.Handle("GET", pattern_Status_JobRegistryStatus_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234569)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234572)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234573)
		}
		__antithesis_instrumentation__.Notify(234570)
		resp, md, err := local_request_Status_JobRegistryStatus_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234574)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234575)
		}
		__antithesis_instrumentation__.Notify(234571)

		forward_Status_JobRegistryStatus_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234253)

	mux.Handle("GET", pattern_Status_JobStatus_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234576)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234579)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234580)
		}
		__antithesis_instrumentation__.Notify(234577)
		resp, md, err := local_request_Status_JobStatus_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234581)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234582)
		}
		__antithesis_instrumentation__.Notify(234578)

		forward_Status_JobStatus_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234254)

	mux.Handle("POST", pattern_Status_ResetSQLStats_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234583)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234586)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234587)
		}
		__antithesis_instrumentation__.Notify(234584)
		resp, md, err := local_request_Status_ResetSQLStats_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234588)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234589)
		}
		__antithesis_instrumentation__.Notify(234585)

		forward_Status_ResetSQLStats_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234255)

	mux.Handle("GET", pattern_Status_IndexUsageStatistics_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234590)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234593)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234594)
		}
		__antithesis_instrumentation__.Notify(234591)
		resp, md, err := local_request_Status_IndexUsageStatistics_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234595)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234596)
		}
		__antithesis_instrumentation__.Notify(234592)

		forward_Status_IndexUsageStatistics_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234256)

	mux.Handle("POST", pattern_Status_ResetIndexUsageStats_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234597)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234600)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234601)
		}
		__antithesis_instrumentation__.Notify(234598)
		resp, md, err := local_request_Status_ResetIndexUsageStats_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234602)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234603)
		}
		__antithesis_instrumentation__.Notify(234599)

		forward_Status_ResetIndexUsageStats_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234257)

	mux.Handle("GET", pattern_Status_TableIndexStats_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234604)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234607)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234608)
		}
		__antithesis_instrumentation__.Notify(234605)
		resp, md, err := local_request_Status_TableIndexStats_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234609)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234610)
		}
		__antithesis_instrumentation__.Notify(234606)

		forward_Status_TableIndexStats_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234258)

	mux.Handle("GET", pattern_Status_UserSQLRoles_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234611)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234614)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234615)
		}
		__antithesis_instrumentation__.Notify(234612)
		resp, md, err := local_request_Status_UserSQLRoles_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234616)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234617)
		}
		__antithesis_instrumentation__.Notify(234613)

		forward_Status_UserSQLRoles_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234259)

	mux.Handle("GET", pattern_Status_TransactionContentionEvents_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234618)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234621)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234622)
		}
		__antithesis_instrumentation__.Notify(234619)
		resp, md, err := local_request_Status_TransactionContentionEvents_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234623)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234624)
		}
		__antithesis_instrumentation__.Notify(234620)

		forward_Status_TransactionContentionEvents_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234260)

	return nil
}

func RegisterStatusHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error) {
	__antithesis_instrumentation__.Notify(234625)
	conn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		__antithesis_instrumentation__.Notify(234628)
		return err
	} else {
		__antithesis_instrumentation__.Notify(234629)
	}
	__antithesis_instrumentation__.Notify(234626)
	defer func() {
		__antithesis_instrumentation__.Notify(234630)
		if err != nil {
			__antithesis_instrumentation__.Notify(234632)
			if cerr := conn.Close(); cerr != nil {
				__antithesis_instrumentation__.Notify(234634)
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			} else {
				__antithesis_instrumentation__.Notify(234635)
			}
			__antithesis_instrumentation__.Notify(234633)
			return
		} else {
			__antithesis_instrumentation__.Notify(234636)
		}
		__antithesis_instrumentation__.Notify(234631)
		go func() {
			__antithesis_instrumentation__.Notify(234637)
			<-ctx.Done()
			if cerr := conn.Close(); cerr != nil {
				__antithesis_instrumentation__.Notify(234638)
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			} else {
				__antithesis_instrumentation__.Notify(234639)
			}
		}()
	}()
	__antithesis_instrumentation__.Notify(234627)

	return RegisterStatusHandler(ctx, mux, conn)
}

func RegisterStatusHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	__antithesis_instrumentation__.Notify(234640)
	return RegisterStatusHandlerClient(ctx, mux, NewStatusClient(conn))
}

func RegisterStatusHandlerClient(ctx context.Context, mux *runtime.ServeMux, client StatusClient) error {
	__antithesis_instrumentation__.Notify(234641)

	mux.Handle("GET", pattern_Status_Certificates_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234694)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234697)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234698)
		}
		__antithesis_instrumentation__.Notify(234695)
		resp, md, err := request_Status_Certificates_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234699)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234700)
		}
		__antithesis_instrumentation__.Notify(234696)

		forward_Status_Certificates_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234642)

	mux.Handle("GET", pattern_Status_Details_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234701)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234704)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234705)
		}
		__antithesis_instrumentation__.Notify(234702)
		resp, md, err := request_Status_Details_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234706)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234707)
		}
		__antithesis_instrumentation__.Notify(234703)

		forward_Status_Details_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234643)

	mux.Handle("GET", pattern_Status_Nodes_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234708)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234711)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234712)
		}
		__antithesis_instrumentation__.Notify(234709)
		resp, md, err := request_Status_Nodes_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234713)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234714)
		}
		__antithesis_instrumentation__.Notify(234710)

		forward_Status_Nodes_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234644)

	mux.Handle("GET", pattern_Status_Node_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234715)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234718)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234719)
		}
		__antithesis_instrumentation__.Notify(234716)
		resp, md, err := request_Status_Node_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234720)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234721)
		}
		__antithesis_instrumentation__.Notify(234717)

		forward_Status_Node_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234645)

	mux.Handle("GET", pattern_Status_NodesUI_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234722)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234725)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234726)
		}
		__antithesis_instrumentation__.Notify(234723)
		resp, md, err := request_Status_NodesUI_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234727)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234728)
		}
		__antithesis_instrumentation__.Notify(234724)

		forward_Status_NodesUI_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234646)

	mux.Handle("GET", pattern_Status_NodeUI_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234729)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234732)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234733)
		}
		__antithesis_instrumentation__.Notify(234730)
		resp, md, err := request_Status_NodeUI_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234734)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234735)
		}
		__antithesis_instrumentation__.Notify(234731)

		forward_Status_NodeUI_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234647)

	mux.Handle("GET", pattern_Status_RaftDebug_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234736)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234739)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234740)
		}
		__antithesis_instrumentation__.Notify(234737)
		resp, md, err := request_Status_RaftDebug_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234741)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234742)
		}
		__antithesis_instrumentation__.Notify(234738)

		forward_Status_RaftDebug_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234648)

	mux.Handle("GET", pattern_Status_Ranges_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234743)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234746)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234747)
		}
		__antithesis_instrumentation__.Notify(234744)
		resp, md, err := request_Status_Ranges_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234748)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234749)
		}
		__antithesis_instrumentation__.Notify(234745)

		forward_Status_Ranges_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234649)

	mux.Handle("GET", pattern_Status_TenantRanges_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234750)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234753)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234754)
		}
		__antithesis_instrumentation__.Notify(234751)
		resp, md, err := request_Status_TenantRanges_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234755)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234756)
		}
		__antithesis_instrumentation__.Notify(234752)

		forward_Status_TenantRanges_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234650)

	mux.Handle("GET", pattern_Status_Gossip_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234757)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234760)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234761)
		}
		__antithesis_instrumentation__.Notify(234758)
		resp, md, err := request_Status_Gossip_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234762)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234763)
		}
		__antithesis_instrumentation__.Notify(234759)

		forward_Status_Gossip_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234651)

	mux.Handle("GET", pattern_Status_EngineStats_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234764)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234767)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234768)
		}
		__antithesis_instrumentation__.Notify(234765)
		resp, md, err := request_Status_EngineStats_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234769)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234770)
		}
		__antithesis_instrumentation__.Notify(234766)

		forward_Status_EngineStats_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234652)

	mux.Handle("GET", pattern_Status_Allocator_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234771)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234774)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234775)
		}
		__antithesis_instrumentation__.Notify(234772)
		resp, md, err := request_Status_Allocator_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234776)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234777)
		}
		__antithesis_instrumentation__.Notify(234773)

		forward_Status_Allocator_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234653)

	mux.Handle("GET", pattern_Status_AllocatorRange_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234778)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234781)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234782)
		}
		__antithesis_instrumentation__.Notify(234779)
		resp, md, err := request_Status_AllocatorRange_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234783)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234784)
		}
		__antithesis_instrumentation__.Notify(234780)

		forward_Status_AllocatorRange_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234654)

	mux.Handle("GET", pattern_Status_ListSessions_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234785)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234788)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234789)
		}
		__antithesis_instrumentation__.Notify(234786)
		resp, md, err := request_Status_ListSessions_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234790)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234791)
		}
		__antithesis_instrumentation__.Notify(234787)

		forward_Status_ListSessions_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234655)

	mux.Handle("GET", pattern_Status_ListLocalSessions_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234792)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234795)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234796)
		}
		__antithesis_instrumentation__.Notify(234793)
		resp, md, err := request_Status_ListLocalSessions_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234797)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234798)
		}
		__antithesis_instrumentation__.Notify(234794)

		forward_Status_ListLocalSessions_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234656)

	mux.Handle("POST", pattern_Status_CancelQuery_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234799)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234802)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234803)
		}
		__antithesis_instrumentation__.Notify(234800)
		resp, md, err := request_Status_CancelQuery_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234804)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234805)
		}
		__antithesis_instrumentation__.Notify(234801)

		forward_Status_CancelQuery_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234657)

	mux.Handle("POST", pattern_Status_CancelLocalQuery_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234806)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234809)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234810)
		}
		__antithesis_instrumentation__.Notify(234807)
		resp, md, err := request_Status_CancelLocalQuery_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234811)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234812)
		}
		__antithesis_instrumentation__.Notify(234808)

		forward_Status_CancelLocalQuery_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234658)

	mux.Handle("GET", pattern_Status_ListContentionEvents_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234813)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234816)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234817)
		}
		__antithesis_instrumentation__.Notify(234814)
		resp, md, err := request_Status_ListContentionEvents_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234818)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234819)
		}
		__antithesis_instrumentation__.Notify(234815)

		forward_Status_ListContentionEvents_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234659)

	mux.Handle("GET", pattern_Status_ListLocalContentionEvents_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234820)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234823)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234824)
		}
		__antithesis_instrumentation__.Notify(234821)
		resp, md, err := request_Status_ListLocalContentionEvents_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234825)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234826)
		}
		__antithesis_instrumentation__.Notify(234822)

		forward_Status_ListLocalContentionEvents_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234660)

	mux.Handle("GET", pattern_Status_ListDistSQLFlows_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234827)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234830)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234831)
		}
		__antithesis_instrumentation__.Notify(234828)
		resp, md, err := request_Status_ListDistSQLFlows_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234832)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234833)
		}
		__antithesis_instrumentation__.Notify(234829)

		forward_Status_ListDistSQLFlows_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234661)

	mux.Handle("GET", pattern_Status_ListLocalDistSQLFlows_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234834)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234837)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234838)
		}
		__antithesis_instrumentation__.Notify(234835)
		resp, md, err := request_Status_ListLocalDistSQLFlows_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234839)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234840)
		}
		__antithesis_instrumentation__.Notify(234836)

		forward_Status_ListLocalDistSQLFlows_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234662)

	mux.Handle("POST", pattern_Status_CancelSession_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234841)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234844)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234845)
		}
		__antithesis_instrumentation__.Notify(234842)
		resp, md, err := request_Status_CancelSession_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234846)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234847)
		}
		__antithesis_instrumentation__.Notify(234843)

		forward_Status_CancelSession_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234663)

	mux.Handle("POST", pattern_Status_CancelLocalSession_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234848)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234851)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234852)
		}
		__antithesis_instrumentation__.Notify(234849)
		resp, md, err := request_Status_CancelLocalSession_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234853)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234854)
		}
		__antithesis_instrumentation__.Notify(234850)

		forward_Status_CancelLocalSession_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234664)

	mux.Handle("POST", pattern_Status_SpanStats_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234855)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234858)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234859)
		}
		__antithesis_instrumentation__.Notify(234856)
		resp, md, err := request_Status_SpanStats_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234860)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234861)
		}
		__antithesis_instrumentation__.Notify(234857)

		forward_Status_SpanStats_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234665)

	mux.Handle("GET", pattern_Status_Stacks_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234862)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234865)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234866)
		}
		__antithesis_instrumentation__.Notify(234863)
		resp, md, err := request_Status_Stacks_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234867)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234868)
		}
		__antithesis_instrumentation__.Notify(234864)

		forward_Status_Stacks_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234666)

	mux.Handle("GET", pattern_Status_Profile_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234869)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234872)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234873)
		}
		__antithesis_instrumentation__.Notify(234870)
		resp, md, err := request_Status_Profile_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234874)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234875)
		}
		__antithesis_instrumentation__.Notify(234871)

		forward_Status_Profile_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234667)

	mux.Handle("GET", pattern_Status_Metrics_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234876)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234879)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234880)
		}
		__antithesis_instrumentation__.Notify(234877)
		resp, md, err := request_Status_Metrics_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234881)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234882)
		}
		__antithesis_instrumentation__.Notify(234878)

		forward_Status_Metrics_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234668)

	mux.Handle("GET", pattern_Status_GetFiles_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234883)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234886)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234887)
		}
		__antithesis_instrumentation__.Notify(234884)
		resp, md, err := request_Status_GetFiles_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234888)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234889)
		}
		__antithesis_instrumentation__.Notify(234885)

		forward_Status_GetFiles_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234669)

	mux.Handle("GET", pattern_Status_LogFilesList_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234890)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234893)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234894)
		}
		__antithesis_instrumentation__.Notify(234891)
		resp, md, err := request_Status_LogFilesList_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234895)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234896)
		}
		__antithesis_instrumentation__.Notify(234892)

		forward_Status_LogFilesList_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234670)

	mux.Handle("GET", pattern_Status_LogFile_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234897)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234900)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234901)
		}
		__antithesis_instrumentation__.Notify(234898)
		resp, md, err := request_Status_LogFile_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234902)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234903)
		}
		__antithesis_instrumentation__.Notify(234899)

		forward_Status_LogFile_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234671)

	mux.Handle("GET", pattern_Status_Logs_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234904)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234907)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234908)
		}
		__antithesis_instrumentation__.Notify(234905)
		resp, md, err := request_Status_Logs_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234909)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234910)
		}
		__antithesis_instrumentation__.Notify(234906)

		forward_Status_Logs_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234672)

	mux.Handle("GET", pattern_Status_ProblemRanges_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234911)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234914)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234915)
		}
		__antithesis_instrumentation__.Notify(234912)
		resp, md, err := request_Status_ProblemRanges_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234916)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234917)
		}
		__antithesis_instrumentation__.Notify(234913)

		forward_Status_ProblemRanges_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234673)

	mux.Handle("GET", pattern_Status_HotRanges_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234918)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234921)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234922)
		}
		__antithesis_instrumentation__.Notify(234919)
		resp, md, err := request_Status_HotRanges_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234923)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234924)
		}
		__antithesis_instrumentation__.Notify(234920)

		forward_Status_HotRanges_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234674)

	mux.Handle("POST", pattern_Status_HotRangesV2_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234925)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234928)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234929)
		}
		__antithesis_instrumentation__.Notify(234926)
		resp, md, err := request_Status_HotRangesV2_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234930)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234931)
		}
		__antithesis_instrumentation__.Notify(234927)

		forward_Status_HotRangesV2_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234675)

	mux.Handle("GET", pattern_Status_Range_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234932)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234935)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234936)
		}
		__antithesis_instrumentation__.Notify(234933)
		resp, md, err := request_Status_Range_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234937)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234938)
		}
		__antithesis_instrumentation__.Notify(234934)

		forward_Status_Range_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234676)

	mux.Handle("GET", pattern_Status_Diagnostics_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234939)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234942)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234943)
		}
		__antithesis_instrumentation__.Notify(234940)
		resp, md, err := request_Status_Diagnostics_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234944)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234945)
		}
		__antithesis_instrumentation__.Notify(234941)

		forward_Status_Diagnostics_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234677)

	mux.Handle("GET", pattern_Status_Stores_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234946)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234949)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234950)
		}
		__antithesis_instrumentation__.Notify(234947)
		resp, md, err := request_Status_Stores_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234951)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234952)
		}
		__antithesis_instrumentation__.Notify(234948)

		forward_Status_Stores_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234678)

	mux.Handle("GET", pattern_Status_Statements_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234953)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234956)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234957)
		}
		__antithesis_instrumentation__.Notify(234954)
		resp, md, err := request_Status_Statements_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234958)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234959)
		}
		__antithesis_instrumentation__.Notify(234955)

		forward_Status_Statements_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234679)

	mux.Handle("GET", pattern_Status_CombinedStatementStats_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234960)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234963)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234964)
		}
		__antithesis_instrumentation__.Notify(234961)
		resp, md, err := request_Status_CombinedStatementStats_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234965)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234966)
		}
		__antithesis_instrumentation__.Notify(234962)

		forward_Status_CombinedStatementStats_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234680)

	mux.Handle("GET", pattern_Status_StatementDetails_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234967)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234970)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234971)
		}
		__antithesis_instrumentation__.Notify(234968)
		resp, md, err := request_Status_StatementDetails_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234972)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234973)
		}
		__antithesis_instrumentation__.Notify(234969)

		forward_Status_StatementDetails_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234681)

	mux.Handle("POST", pattern_Status_CreateStatementDiagnosticsReport_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234974)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234977)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234978)
		}
		__antithesis_instrumentation__.Notify(234975)
		resp, md, err := request_Status_CreateStatementDiagnosticsReport_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234979)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234980)
		}
		__antithesis_instrumentation__.Notify(234976)

		forward_Status_CreateStatementDiagnosticsReport_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234682)

	mux.Handle("POST", pattern_Status_CancelStatementDiagnosticsReport_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234981)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234984)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234985)
		}
		__antithesis_instrumentation__.Notify(234982)
		resp, md, err := request_Status_CancelStatementDiagnosticsReport_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234986)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234987)
		}
		__antithesis_instrumentation__.Notify(234983)

		forward_Status_CancelStatementDiagnosticsReport_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234683)

	mux.Handle("GET", pattern_Status_StatementDiagnosticsRequests_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234988)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234991)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234992)
		}
		__antithesis_instrumentation__.Notify(234989)
		resp, md, err := request_Status_StatementDiagnosticsRequests_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(234993)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234994)
		}
		__antithesis_instrumentation__.Notify(234990)

		forward_Status_StatementDiagnosticsRequests_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234684)

	mux.Handle("GET", pattern_Status_StatementDiagnostics_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(234995)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(234998)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(234999)
		}
		__antithesis_instrumentation__.Notify(234996)
		resp, md, err := request_Status_StatementDiagnostics_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(235000)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(235001)
		}
		__antithesis_instrumentation__.Notify(234997)

		forward_Status_StatementDiagnostics_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234685)

	mux.Handle("GET", pattern_Status_JobRegistryStatus_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(235002)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(235005)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(235006)
		}
		__antithesis_instrumentation__.Notify(235003)
		resp, md, err := request_Status_JobRegistryStatus_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(235007)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(235008)
		}
		__antithesis_instrumentation__.Notify(235004)

		forward_Status_JobRegistryStatus_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234686)

	mux.Handle("GET", pattern_Status_JobStatus_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(235009)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(235012)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(235013)
		}
		__antithesis_instrumentation__.Notify(235010)
		resp, md, err := request_Status_JobStatus_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(235014)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(235015)
		}
		__antithesis_instrumentation__.Notify(235011)

		forward_Status_JobStatus_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234687)

	mux.Handle("POST", pattern_Status_ResetSQLStats_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(235016)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(235019)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(235020)
		}
		__antithesis_instrumentation__.Notify(235017)
		resp, md, err := request_Status_ResetSQLStats_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(235021)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(235022)
		}
		__antithesis_instrumentation__.Notify(235018)

		forward_Status_ResetSQLStats_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234688)

	mux.Handle("GET", pattern_Status_IndexUsageStatistics_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(235023)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(235026)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(235027)
		}
		__antithesis_instrumentation__.Notify(235024)
		resp, md, err := request_Status_IndexUsageStatistics_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(235028)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(235029)
		}
		__antithesis_instrumentation__.Notify(235025)

		forward_Status_IndexUsageStatistics_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234689)

	mux.Handle("POST", pattern_Status_ResetIndexUsageStats_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(235030)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(235033)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(235034)
		}
		__antithesis_instrumentation__.Notify(235031)
		resp, md, err := request_Status_ResetIndexUsageStats_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(235035)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(235036)
		}
		__antithesis_instrumentation__.Notify(235032)

		forward_Status_ResetIndexUsageStats_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234690)

	mux.Handle("GET", pattern_Status_TableIndexStats_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(235037)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(235040)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(235041)
		}
		__antithesis_instrumentation__.Notify(235038)
		resp, md, err := request_Status_TableIndexStats_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(235042)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(235043)
		}
		__antithesis_instrumentation__.Notify(235039)

		forward_Status_TableIndexStats_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234691)

	mux.Handle("GET", pattern_Status_UserSQLRoles_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(235044)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(235047)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(235048)
		}
		__antithesis_instrumentation__.Notify(235045)
		resp, md, err := request_Status_UserSQLRoles_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(235049)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(235050)
		}
		__antithesis_instrumentation__.Notify(235046)

		forward_Status_UserSQLRoles_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234692)

	mux.Handle("GET", pattern_Status_TransactionContentionEvents_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		__antithesis_instrumentation__.Notify(235051)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(235054)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(235055)
		}
		__antithesis_instrumentation__.Notify(235052)
		resp, md, err := request_Status_TransactionContentionEvents_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			__antithesis_instrumentation__.Notify(235056)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(235057)
		}
		__antithesis_instrumentation__.Notify(235053)

		forward_Status_TransactionContentionEvents_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	__antithesis_instrumentation__.Notify(234693)

	return nil
}

var (
	pattern_Status_Certificates_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 1, 0, 4, 1, 5, 2}, []string{"_status", "certificates", "node_id"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_Details_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 1, 0, 4, 1, 5, 2}, []string{"_status", "details", "node_id"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_Nodes_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"_status", "nodes"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_Node_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 1, 0, 4, 1, 5, 2}, []string{"_status", "nodes", "node_id"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_NodesUI_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"_status", "nodes_ui"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_NodeUI_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 1, 0, 4, 1, 5, 2}, []string{"_status", "nodes_ui", "node_id"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_RaftDebug_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"_status", "raft"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_Ranges_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 1, 0, 4, 1, 5, 2}, []string{"_status", "ranges", "node_id"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_TenantRanges_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"_status", "tenant_ranges"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_Gossip_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 1, 0, 4, 1, 5, 2}, []string{"_status", "gossip", "node_id"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_EngineStats_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 1, 0, 4, 1, 5, 2}, []string{"_status", "enginestats", "node_id"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_Allocator_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2, 1, 0, 4, 1, 5, 3}, []string{"_status", "allocator", "node", "node_id"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_AllocatorRange_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2, 1, 0, 4, 1, 5, 3}, []string{"_status", "allocator", "range", "range_id"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_ListSessions_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"_status", "sessions"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_ListLocalSessions_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"_status", "local_sessions"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_CancelQuery_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 1, 0, 4, 1, 5, 2}, []string{"_status", "cancel_query", "node_id"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_CancelLocalQuery_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"_status", "cancel_local_query"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_ListContentionEvents_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"_status", "contention_events"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_ListLocalContentionEvents_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"_status", "local_contention_events"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_ListDistSQLFlows_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"_status", "distsql_flows"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_ListLocalDistSQLFlows_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"_status", "local_distsql_flows"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_CancelSession_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 1, 0, 4, 1, 5, 2}, []string{"_status", "cancel_session", "node_id"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_CancelLocalSession_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"_status", "cancel_local_session"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_SpanStats_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"_status", "span"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_Stacks_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 1, 0, 4, 1, 5, 2}, []string{"_status", "stacks", "node_id"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_Profile_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 1, 0, 4, 1, 5, 2}, []string{"_status", "profile", "node_id"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_Metrics_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 1, 0, 4, 1, 5, 2}, []string{"_status", "metrics", "node_id"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_GetFiles_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 1, 0, 4, 1, 5, 2}, []string{"_status", "files", "node_id"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_LogFilesList_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 1, 0, 4, 1, 5, 2}, []string{"_status", "logfiles", "node_id"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_LogFile_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 1, 0, 4, 1, 5, 2, 1, 0, 4, 1, 5, 3}, []string{"_status", "logfiles", "node_id", "file"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_Logs_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 1, 0, 4, 1, 5, 2}, []string{"_status", "logs", "node_id"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_ProblemRanges_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"_status", "problemranges"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_HotRanges_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"_status", "hotranges"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_HotRangesV2_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2}, []string{"_status", "v2", "hotranges"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_Range_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 1, 0, 4, 1, 5, 2}, []string{"_status", "range", "range_id"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_Diagnostics_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 1, 0, 4, 1, 5, 2}, []string{"_status", "diagnostics", "node_id"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_Stores_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 1, 0, 4, 1, 5, 2}, []string{"_status", "stores", "node_id"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_Statements_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"_status", "statements"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_CombinedStatementStats_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"_status", "combinedstmts"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_StatementDetails_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 1, 0, 4, 1, 5, 2}, []string{"_status", "stmtdetails", "fingerprint_id"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_CreateStatementDiagnosticsReport_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"_status", "stmtdiagreports"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_CancelStatementDiagnosticsReport_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2}, []string{"_status", "stmtdiagreports", "cancel"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_StatementDiagnosticsRequests_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"_status", "stmtdiagreports"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_StatementDiagnostics_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 1, 0, 4, 1, 5, 2}, []string{"_status", "stmtdiag", "statement_diagnostics_id"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_JobRegistryStatus_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 1, 0, 4, 1, 5, 2}, []string{"_status", "job_registry", "node_id"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_JobStatus_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 1, 0, 4, 1, 5, 2}, []string{"_status", "job", "job_id"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_ResetSQLStats_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"_status", "resetsqlstats"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_IndexUsageStatistics_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"_status", "indexusagestatistics"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_ResetIndexUsageStats_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"_status", "resetindexusagestats"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_TableIndexStats_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 1, 0, 4, 1, 5, 2, 2, 3, 1, 0, 4, 1, 5, 4, 2, 5}, []string{"_status", "databases", "database", "tables", "table", "indexstats"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_UserSQLRoles_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"_status", "sqlroles"}, "", runtime.AssumeColonVerbOpt(true)))

	pattern_Status_TransactionContentionEvents_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"_status", "transactioncontentionevents"}, "", runtime.AssumeColonVerbOpt(true)))
)

var (
	forward_Status_Certificates_0 = runtime.ForwardResponseMessage

	forward_Status_Details_0 = runtime.ForwardResponseMessage

	forward_Status_Nodes_0 = runtime.ForwardResponseMessage

	forward_Status_Node_0 = runtime.ForwardResponseMessage

	forward_Status_NodesUI_0 = runtime.ForwardResponseMessage

	forward_Status_NodeUI_0 = runtime.ForwardResponseMessage

	forward_Status_RaftDebug_0 = runtime.ForwardResponseMessage

	forward_Status_Ranges_0 = runtime.ForwardResponseMessage

	forward_Status_TenantRanges_0 = runtime.ForwardResponseMessage

	forward_Status_Gossip_0 = runtime.ForwardResponseMessage

	forward_Status_EngineStats_0 = runtime.ForwardResponseMessage

	forward_Status_Allocator_0 = runtime.ForwardResponseMessage

	forward_Status_AllocatorRange_0 = runtime.ForwardResponseMessage

	forward_Status_ListSessions_0 = runtime.ForwardResponseMessage

	forward_Status_ListLocalSessions_0 = runtime.ForwardResponseMessage

	forward_Status_CancelQuery_0 = runtime.ForwardResponseMessage

	forward_Status_CancelLocalQuery_0 = runtime.ForwardResponseMessage

	forward_Status_ListContentionEvents_0 = runtime.ForwardResponseMessage

	forward_Status_ListLocalContentionEvents_0 = runtime.ForwardResponseMessage

	forward_Status_ListDistSQLFlows_0 = runtime.ForwardResponseMessage

	forward_Status_ListLocalDistSQLFlows_0 = runtime.ForwardResponseMessage

	forward_Status_CancelSession_0 = runtime.ForwardResponseMessage

	forward_Status_CancelLocalSession_0 = runtime.ForwardResponseMessage

	forward_Status_SpanStats_0 = runtime.ForwardResponseMessage

	forward_Status_Stacks_0 = runtime.ForwardResponseMessage

	forward_Status_Profile_0 = runtime.ForwardResponseMessage

	forward_Status_Metrics_0 = runtime.ForwardResponseMessage

	forward_Status_GetFiles_0 = runtime.ForwardResponseMessage

	forward_Status_LogFilesList_0 = runtime.ForwardResponseMessage

	forward_Status_LogFile_0 = runtime.ForwardResponseMessage

	forward_Status_Logs_0 = runtime.ForwardResponseMessage

	forward_Status_ProblemRanges_0 = runtime.ForwardResponseMessage

	forward_Status_HotRanges_0 = runtime.ForwardResponseMessage

	forward_Status_HotRangesV2_0 = runtime.ForwardResponseMessage

	forward_Status_Range_0 = runtime.ForwardResponseMessage

	forward_Status_Diagnostics_0 = runtime.ForwardResponseMessage

	forward_Status_Stores_0 = runtime.ForwardResponseMessage

	forward_Status_Statements_0 = runtime.ForwardResponseMessage

	forward_Status_CombinedStatementStats_0 = runtime.ForwardResponseMessage

	forward_Status_StatementDetails_0 = runtime.ForwardResponseMessage

	forward_Status_CreateStatementDiagnosticsReport_0 = runtime.ForwardResponseMessage

	forward_Status_CancelStatementDiagnosticsReport_0 = runtime.ForwardResponseMessage

	forward_Status_StatementDiagnosticsRequests_0 = runtime.ForwardResponseMessage

	forward_Status_StatementDiagnostics_0 = runtime.ForwardResponseMessage

	forward_Status_JobRegistryStatus_0 = runtime.ForwardResponseMessage

	forward_Status_JobStatus_0 = runtime.ForwardResponseMessage

	forward_Status_ResetSQLStats_0 = runtime.ForwardResponseMessage

	forward_Status_IndexUsageStatistics_0 = runtime.ForwardResponseMessage

	forward_Status_ResetIndexUsageStats_0 = runtime.ForwardResponseMessage

	forward_Status_TableIndexStats_0 = runtime.ForwardResponseMessage

	forward_Status_UserSQLRoles_0 = runtime.ForwardResponseMessage

	forward_Status_TransactionContentionEvents_0 = runtime.ForwardResponseMessage
)
