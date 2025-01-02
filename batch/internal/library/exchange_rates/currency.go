package exchange_rates

// 通貨コードの共通の型
type ExchangeRatesCurrencyCode string

const (
	USD ExchangeRatesCurrencyCode = "USD"
	EUR ExchangeRatesCurrencyCode = "EUR"
	JPY ExchangeRatesCurrencyCode = "JPY"
	GBP ExchangeRatesCurrencyCode = "GBP"
	AUD ExchangeRatesCurrencyCode = "AUD"
)

// String: 通貨コード型を文字列型に変換
func (ecc ExchangeRatesCurrencyCode) String() string {
	return string(ecc)
}

// Valid: 指定された基軸通貨が正しいかを検証
//
// 現状のプランでは、USD以外は基軸通貨として指定できないため、無効な通貨が指定されている場合はエラーとして扱う
func (ecc ExchangeRatesCurrencyCode) Valid() bool {
	if ecc.String() == "" {
		return false
	}

	if ecc.String() != USD.String() {
		return false
	}

	return true
}
