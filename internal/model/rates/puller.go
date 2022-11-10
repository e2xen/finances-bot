package rates

import (
	"context"
	"fmt"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	"go.uber.org/zap"
	"max.ks1230/project-base/internal/logger"

	"github.com/pkg/errors"
	"max.ks1230/project-base/internal/entity/currency"
	"max.ks1230/project-base/internal/utils"
)

type ratesStorage interface {
	NewRate(ctx context.Context, name string) error
	UpdateRateValue(ctx context.Context, name string, val float64) error
}

type ratesProvider interface {
	GetRates(ctx context.Context, base string, relatives []string) (map[string]float64, error)
}

type config interface {
	BaseCurrency() string
	PullingDelayMinutes() int64
}

type Puller struct {
	storage      ratesStorage
	provider     ratesProvider
	baseCurrency string
	pullingDelay int64
}

func NewPuller(storage ratesStorage, provider ratesProvider, config config) (*Puller, error) {
	p := &Puller{
		storage:      storage,
		provider:     provider,
		baseCurrency: config.BaseCurrency(),
		pullingDelay: config.PullingDelayMinutes(),
	}
	err := p.initStorage()
	if err != nil {
		return nil, errors.Wrap(err, "cannot init storage")
	}
	return p, nil
}

func (p *Puller) initStorage() error {
	ctx := context.Background()

	if !utils.Contains(currency.Currencies, p.baseCurrency) {
		return fmt.Errorf("unknown currency %s", p.baseCurrency)
	}

	for _, curr := range currency.Currencies {
		err := p.storage.NewRate(ctx, curr)
		if err != nil {
			return errors.New("cannot save currency to storage")
		}
	}

	err := p.storage.UpdateRateValue(ctx, p.baseCurrency, 1)
	if err != nil {
		return errors.New("cannot update currency")
	}
	return nil
}

func (p *Puller) Pull(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(p.pullingDelay) * time.Minute)
	firstTick := make(chan struct{}, 1)
	firstTick <- struct{}{}

	logger.Info("Start pulling rates")
	for {
		select {
		case <-ctx.Done():
			logger.Info("Stop pulling rates")
			return
		// fake first tick to pull rates immediately
		case <-firstTick:
			p.pullOnce(ctx)
		case <-ticker.C:
			p.pullOnce(ctx)
		}
	}
}

func (p *Puller) pullOnce(ctx context.Context) {
	logger.Info("Pulling current rates...")

	span, ctx := opentracing.StartSpanFromContext(ctx, "pullRates")
	defer span.Finish()

	relatives := p.nonBaseCurrencies()
	pulledRates, err := p.provider.GetRates(ctx, p.baseCurrency, relatives)
	if err != nil {
		logger.Error("cannot get rates", zap.Error(err))
		return
	}

	for name, rate := range pulledRates {
		p.updateRate(ctx, name, rate)
	}

	logger.Info("Successfully pulled current rates")
}

func (p *Puller) updateRate(ctx context.Context, name string, rate float64) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "updateRate")
	defer span.Finish()
	span.SetTag("rate", name)

	err := p.storage.UpdateRateValue(ctx, name, rate)
	if err == nil {
		logger.Info("successfully saved rate", zap.String("rate", name))
	} else {
		ext.Error.Set(span, true)
		logger.Error("failed to save rate", zap.Error(err), zap.String("rate", name))
	}
}

func (p *Puller) nonBaseCurrencies() []string {
	var relatives []string
	for _, curr := range currency.Currencies {
		if curr != p.baseCurrency {
			relatives = append(relatives, curr)
		}
	}
	return relatives
}
