package storage

import (
	"max.ks1230/project-base/internal/model/user"
)

type InMemStorage struct {
	userMap map[int64]user.Record
}

func NewInMemStorage() *InMemStorage {
	s := make(map[int64]user.Record)
	return &InMemStorage{s}
}

func (s *InMemStorage) GetByID(id int64) (user.Record, error) {
	u, ok := s.userMap[id]
	if !ok {
		return user.Record{}, nil
	}
	return u, nil
}

func (s *InMemStorage) SaveByID(id int64, rec user.Record) error {
	s.userMap[id] = rec
	return nil
}
