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
	expenses   map[int64][]user.ExpenseRecord
	rates      map[string]currency.Rate
	ratesMutex sync.RWMutex
}

func NewInMemStorage() *InMemStorage {
	u := make(map[int64]user.Record)
	e := make(map[int64][]user.ExpenseRecord)
	r := make(map[string]currency.Rate)
	return &InMemStorage{u, e, r, sync.RWMutex{}}
}

func (s *InMemStorage) GetUserByID(id int64) (user.Record, error) {
	u, _ := s.users[id]
	return u, nil
}

func (s *InMemStorage) SaveExpense(userID int64, record user.ExpenseRecord) error {
	exps, _ := s.expenses[userID]
	exps = append(exps, record)
	s.expenses[userID] = exps
	return nil
}

func (s *InMemStorage) GetUserExpenses(userID int64) ([]user.ExpenseRecord, error) {
	exps, ok := s.expenses[userID]
	if !ok {
		return make([]user.ExpenseRecord, 0), nil
	}
	return exps, nil
}

func (s *InMemStorage) SaveUserByID(id int64, rec user.Record) error {
	s.users[id] = rec
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
