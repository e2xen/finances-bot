package user

import "time"

type ExpenseRecord struct {
	Amount   float64
	Category string
	Created  time.Time
}

type Record struct {
	Expenses []ExpenseRecord
}
