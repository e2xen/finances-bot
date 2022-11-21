package reports

import (
	"context"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "max.ks1230/project-base/api/grpc"
	"max.ks1230/project-base/internal/logger"
)

type Sender struct {
	conn   *grpc.ClientConn
	client pb.ReportAcceptorClient
}

func NewSender(addr string) (*Sender, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, errors.Wrap(err, "cannot initiate new connection")
	}
	client := pb.NewReportAcceptorClient(conn)
	return &Sender{conn, client}, nil
}

func (s *Sender) Close() {
	err := s.conn.Close()
	if err != nil {
		logger.Error("failed to close grpc connection", zap.Error(err))
	}
}

func (s *Sender) SendReport(ctx context.Context, report *pb.ReportResult) error {
	logger.Info("SendReport - start", zap.Int64("userID", report.GetUserID()))
	defer logger.Info("SendReport - end")

	_, err := s.client.AcceptReport(ctx, report)
	return err
}
