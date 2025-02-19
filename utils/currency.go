package utils

const (
	USD = "USD"
	EUR = "EUR"
	CAD = "CAD"
)

func IsCurrencySupported(currency string) bool {
	switch currency {
	case USD, EUR, CAD:
		return true
	}
	return false
}
