package server

import (
	"github.com/go-playground/validator/v10"
	"github.com/simplebank/internal/testutils"
)

var validCurrency validator.Func = func(fieldLevel validator.FieldLevel) bool {
	if currency, ok := fieldLevel.Field().Interface().(string); ok {
		return testutils.IsSupportedCurrency(currency)
	}
	return false
}
