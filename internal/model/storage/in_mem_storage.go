package storage

import (
	"fmt"
	"max.ks1230/project-base/internal/entity/currency"
	"max.ks1230/project-base/internal/entity/user"
	"sync"
	"time"
)

type InMemStorage struct {
	users      map[int64]user.Record
	rates      map[string]currency.Rate
	ratesMutex sync.RWMutex
}

func NewInMemStorage() *InMemStorage {
	u := make(map[int64]user.Record)
	r := make(map[string]currency.Rate)
	return &InMemStorage{u, r, sync.RWMutex{}}
}

func (s *InMemStorage) GetUserByID(id int64) (user.Record, error) {
	u, ok := s.users[id]
	if !ok {
		return user.Record{}, nil
	}
	return u, nil
}

func (s *InMemStorage) SaveUserByID(id int64, rec user.Record) error {
	s.users[id] = rec
	return nil
}

func (s *InMemStorage) SetCurrencyForUser(id int64, curr string) error {
	u, ok := s.users[id]
	if !ok {
		u = user.Record{}
	}
	u.SetPreferredCurrency(curr)
	s.users[id] = u
	return nil
}

func (s *InMemStorage) GetRate(name string) (currency.Rate, error) {
	s.ratesMutex.RLock()
	defer s.ratesMutex.RUnlock()

	r, ok := s.rates[name]
	if !ok {
		return currency.Rate{}, fmt.Errorf("rate %s not found", name)
	}
	if !r.Set {
		return currency.Rate{}, fmt.Errorf("rate %s is not set yet", name)
	}
	return r, nil
}

func (s *InMemStorage) NewRate(name string) error {
	s.ratesMutex.Lock()
	defer s.ratesMutex.Unlock()

	s.rates[name] = currency.Rate{
		Name: name,
	}
	return nil
}

func (s *InMemStorage) UpdateRateValue(name string, val float64) error {
	s.ratesMutex.Lock()
	defer s.ratesMutex.Unlock()

	r, ok := s.rates[name]
	if !ok {
		return fmt.Errorf("rate %s not found", name)
	}

	r.BaseRate = val
	r.Set = true
	r.UpdatedAt = time.Now()

	s.rates[name] = r
	return nil
}
