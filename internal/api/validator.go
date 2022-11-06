package api

import "github.com/go-playground/validator/v10"

const (
	USD = "USD"
	EUR = "EUR"
	BRL = "BRL"
)

var validCurrency validator.Func = func(fl validator.FieldLevel) bool {
	if currency, ok := fl.Field().Interface().(string); ok {
		return isSupportedCurrency(currency)
	}
	return false
}

func isSupportedCurrency(currency string) bool {
	switch currency {
	case USD, EUR, BRL:
		return true
	}
	return false
}
