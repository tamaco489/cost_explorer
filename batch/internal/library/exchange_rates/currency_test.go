package exchange_rates_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tamaco489/cost_explorer/batch/internal/library/exchange_rates"
)

func TestExchangeRatesCurrencyCode_Valid(t *testing.T) {
	tests := map[string]struct {
		input    exchange_rates.ExchangeRatesCurrencyCode
		expected bool
	}{
		"USDは有効な通貨コード": {
			input:    exchange_rates.USD,
			expected: true,
		},
		"EURは無効な通貨コード": {
			input:    exchange_rates.EUR,
			expected: false,
		},
		"JPYは無効な通貨コード": {
			input:    exchange_rates.JPY,
			expected: false,
		},
		"空文字は無効な通貨コード": {
			input:    exchange_rates.ExchangeRatesCurrencyCode(""),
			expected: false,
		},
		"他の通貨コードも無効": {
			input:    exchange_rates.ExchangeRatesCurrencyCode("GBP"),
			expected: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := tt.input.Valid()
			assert.Equal(t, result, tt.expected)
		})
	}
}
