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

func SendReport(ctx context.Context, addr string, report *pb.ReportResult) error {
	logger.Info("SendReport - start", zap.Int64("userID", report.GetUserID()))
	defer logger.Info("SendReport - end")

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return errors.Wrap(err, "send report")
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			logger.Error("failed to close grpc connection", zap.Error(err))
		}
	}()

	client := pb.NewReportAcceptorClient(conn)
	_, err = client.AcceptReport(ctx, report)
	return err
}
