// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.9
// source: api/grpc/report-result.proto

package apiv1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// ReportAcceptorClient is the client API for ReportAcceptor service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ReportAcceptorClient interface {
	AcceptReport(ctx context.Context, in *ReportResult, opts ...grpc.CallOption) (*OperationStatus, error)
}

type reportAcceptorClient struct {
	cc grpc.ClientConnInterface
}

func NewReportAcceptorClient(cc grpc.ClientConnInterface) ReportAcceptorClient {
	return &reportAcceptorClient{cc}
}

func (c *reportAcceptorClient) AcceptReport(ctx context.Context, in *ReportResult, opts ...grpc.CallOption) (*OperationStatus, error) {
	out := new(OperationStatus)
	err := c.cc.Invoke(ctx, "/report.ReportAcceptor/AcceptReport", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ReportAcceptorServer is the server API for ReportAcceptor service.
// All implementations must embed UnimplementedReportAcceptorServer
// for forward compatibility
type ReportAcceptorServer interface {
	AcceptReport(context.Context, *ReportResult) (*OperationStatus, error)
	mustEmbedUnimplementedReportAcceptorServer()
}

// UnimplementedReportAcceptorServer must be embedded to have forward compatible implementations.
type UnimplementedReportAcceptorServer struct {
}

func (UnimplementedReportAcceptorServer) AcceptReport(context.Context, *ReportResult) (*OperationStatus, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AcceptReport not implemented")
}
func (UnimplementedReportAcceptorServer) mustEmbedUnimplementedReportAcceptorServer() {}

// UnsafeReportAcceptorServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ReportAcceptorServer will
// result in compilation errors.
type UnsafeReportAcceptorServer interface {
	mustEmbedUnimplementedReportAcceptorServer()
}

func RegisterReportAcceptorServer(s grpc.ServiceRegistrar, srv ReportAcceptorServer) {
	s.RegisterService(&ReportAcceptor_ServiceDesc, srv)
}

func _ReportAcceptor_AcceptReport_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReportResult)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReportAcceptorServer).AcceptReport(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/report.ReportAcceptor/AcceptReport",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReportAcceptorServer).AcceptReport(ctx, req.(*ReportResult))
	}
	return interceptor(ctx, in, info, handler)
}

// ReportAcceptor_ServiceDesc is the grpc.ServiceDesc for ReportAcceptor service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ReportAcceptor_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "report.ReportAcceptor",
	HandlerType: (*ReportAcceptorServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AcceptReport",
			Handler:    _ReportAcceptor_AcceptReport_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/grpc/report-result.proto",
}
