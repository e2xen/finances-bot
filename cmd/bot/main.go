package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"max.ks1230/project-base/internal/model/reports"

	"max.ks1230/project-base/internal/clients/cache"

	"max.ks1230/project-base/internal/clients/kafka"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	jconfig "github.com/uber/jaeger-client-go/config"
	"go.uber.org/zap"

	"max.ks1230/project-base/internal/clients/fixer"
	"max.ks1230/project-base/internal/clients/tg"
	"max.ks1230/project-base/internal/config"
	"max.ks1230/project-base/internal/logger"
	"max.ks1230/project-base/internal/model/messages"
	"max.ks1230/project-base/internal/model/rates"
	"max.ks1230/project-base/internal/model/storage"
)

const (
	serviceName            = "finances-route-bot"
	httpPort               = 80
	grpcPort               = 8080
	shutdownTimeoutSeconds = 2
)

func main() {
	logger.Info("App init - start")

	conf, err := config.New()
	if err != nil {
		logger.Fatal("failed to init config:", zap.Error(err))
	}

	tgClient, err := tg.New(conf.Telegram())
	if err != nil {
		logger.Fatal("failed to init client:", zap.Error(err))
	}

	fixerClient := fixer.New(conf.Fixer())

	userStorage, err := storage.NewPostgresStorage(conf.Postgres())
	if err != nil {
		logger.Fatal("failed to init postgres:", zap.Error(err))
	}

	reportCache, err := cache.NewMemcache(conf.Memcached())
	if err != nil {
		logger.Fatal("failed to init memcache:", zap.Error(err))
	}

	producer, err := kafka.NewProducer(conf.Kafka())
	if err != nil {
		logger.Fatal("failed to init producer:", zap.Error(err))
	}
	defer producer.Close()

	msgService := messages.NewService(conf.App(), tgClient, userStorage, reportCache, producer)

	reportAcceptor, err := reports.NewServer(grpcPort, msgService)
	if err != nil {
		logger.Fatal("failed to init grpc server:", zap.Error(err))
	}

	ratesPuller, err := rates.NewPuller(userStorage, fixerClient, conf.App())
	if err != nil {
		logger.Fatal("failed to init puller:", zap.Error(err))
	}

	logger.Info("App init - end")

	ctx, cancel := context.WithCancel(context.Background())
	cancel = startHTTPServer(cancel)
	defer cancel()
	cancelOnSignals(cancel,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	go ratesPuller.Pull(ctx)
	go reportAcceptor.Serve()
	defer reportAcceptor.Shutdown()

	tgClient.ListenUpdates(ctx, msgService)
}

func cancelOnSignals(cancel context.CancelFunc, signals ...os.Signal) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, signals...)
	go func() {
		<-sigChan
		cancel()
	}()
}

func startHTTPServer(cancelParent context.CancelFunc) context.CancelFunc {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", httpPort),
		Handler: promhttp.Handler(),
	}

	go func() {
		logger.Info("starting http server", zap.Int("port", httpPort))
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Fatal("error starting http server", zap.Error(err))
		}
		logger.Info("http server stopped")
	}()

	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeoutSeconds*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			logger.Error("error shutting down http server", zap.Error(err))
		}
		cancelParent()
	}
}

func init() {
	cfg := jconfig.Configuration{
		Sampler: &jconfig.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
	}

	_, err := cfg.InitGlobalTracer(serviceName)
	if err != nil {
		logger.Fatal("cannot init tracing", zap.Error(err))
	}
}
