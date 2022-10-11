package user

import (
	"time"
)

type ExpenseRecord struct {
	Amount   float64
	Category string
	Created  time.Time
}

type Record struct {
	Expenses          []ExpenseRecord
	preferredCurrency string
}

func (r *Record) PreferredCurrency(def string) string {
	if r.preferredCurrency != "" {
		return r.preferredCurrency
	}
	return def
}

func (r *Record) SetPreferredCurrency(curr string) {
	r.preferredCurrency = curr
}
