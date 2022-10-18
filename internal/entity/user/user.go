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
	preferredCurrency string
	MonthLimit        float64
}

func (r *Record) PreferredCurrencyOrDefault(def string) string {
	if r.preferredCurrency != "" {
		return r.preferredCurrency
	}
	return def
}

func (r *Record) PreferredCurrency() string {
	return r.preferredCurrency
}

func (r *Record) SetPreferredCurrency(curr string) {
	r.preferredCurrency = curr
}
