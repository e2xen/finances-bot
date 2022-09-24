package storage

import (
	"time"
)

type ExpenseRecord struct {
	Amount   float64
	Category string
	Created  time.Time
}

type UserRecord struct {
	Expenses []ExpenseRecord
}

type UserStorage interface {
	GetById(userID int64) (UserRecord, error)
	SaveById(userID int64, rec UserRecord) error
}

type InMemStorage struct {
	userMap map[int64]UserRecord
}

func NewInMemStorage() *InMemStorage {
	s := make(map[int64]UserRecord)
	return &InMemStorage{s}
}

func (s *InMemStorage) GetById(id int64) (UserRecord, error) {
	user, ok := s.userMap[id]
	if !ok {
		return UserRecord{}, nil
	}
	return user, nil
}

func (s *InMemStorage) SaveById(id int64, rec UserRecord) error {
	s.userMap[id] = rec
	return nil
}
