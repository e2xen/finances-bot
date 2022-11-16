package reports

import (
	"context"
	"fmt"
	"net"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"max.ks1230/project-base/internal/logger"

	pb "max.ks1230/project-base/api/grpc"
)

type reportAcceptor interface {
	AcceptReport(ctx context.Context, report *pb.ReportResult) error
}

type AcceptorServer struct {
	pb.UnimplementedReportAcceptorServer
	acceptor reportAcceptor
	server   *grpc.Server
	lis      net.Listener
}

func NewServer(port int, acceptor reportAcceptor) (*AcceptorServer, error) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, errors.Wrap(err, "cannot create server")
	}

	rpcServer := grpc.NewServer()
	service := &AcceptorServer{
		acceptor: acceptor,
		server:   rpcServer,
		lis:      lis,
	}
	pb.RegisterReportAcceptorServer(rpcServer, service)
	return service, nil
}

func (s *AcceptorServer) Serve() {
	logger.Info("gRPC server listening", zap.Any("addr", s.lis.Addr()))
	err := s.server.Serve(s.lis)
	if err != nil {
		logger.Error("failed to serve gPRC", zap.Error(err))
	}
}

func (s *AcceptorServer) Shutdown() {
	s.server.GracefulStop()
	logger.Info("grpc server stopped")
}

func (s *AcceptorServer) AcceptReport(ctx context.Context, in *pb.ReportResult) (*pb.OperationStatus, error) {
	err := s.acceptor.AcceptReport(ctx, in)
	if err != nil {
		errMes := err.Error()
		return &pb.OperationStatus{Success: false, Error: &errMes}, err
	}
	return &pb.OperationStatus{Success: true}, nil
}
