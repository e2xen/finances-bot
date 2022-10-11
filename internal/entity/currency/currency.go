package currency

import "time"

const (
	RUB = "RUB"
	USD = "USD"
	EUR = "EUR"
	CNY = "CNY"
)

var Currencies = []string{RUB, USD, EUR, CNY}

type Rate struct {
	Name      string
	BaseRate  float64
	Set       bool
	UpdatedAt time.Time
}
