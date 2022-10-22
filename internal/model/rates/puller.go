package rates

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"max.ks1230/project-base/internal/entity/currency"
	"max.ks1230/project-base/internal/utils"
	"time"
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

	log.Println("Start pulling rates")
	for {
		select {
		case <-ctx.Done():
			log.Println("Stop pulling rates")
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
	log.Println("Pulling rates...")

	relatives := p.nonBaseCurrencies()
	pulledRates, err := p.provider.GetRates(ctx, p.baseCurrency, relatives)
	if err != nil {
		log.Println(errors.Wrap(err, "cannot get rates").Error())
		return
	}

	for name, rate := range pulledRates {
		err = p.storage.UpdateRateValue(ctx, name, rate)
		if err == nil {
			log.Printf("successfully saved rate %s\n", name)
		} else {
			log.Println(errors.Wrap(err, fmt.Sprintf("failed to save rate %s", name)).Error())
		}
	}

	log.Println("Pulled rates")
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
