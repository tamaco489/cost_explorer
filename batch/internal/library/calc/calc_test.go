package calc_test

import (
	"errors"
	"testing"

	"github.com/go-playground/assert"
	"github.com/tamaco489/cost_explorer/batch/internal/library/calc"
)

func TestRoundUpToTwoDecimalPlaces(t *testing.T) {

	tests := map[string]struct {
		input    float64
		expected float64
		err      error
	}{
		"小数点以下2桁の値はそのまま": {
			input:    123.45,
			expected: 123.45,
			err:      nil,
		},
		"小数点以下3桁目を切り上げる": {
			input:    123.451,
			expected: 123.46,
			err:      nil,
		},
		"小数部分がない場合でも正確に処理する": {
			input:    123.0,
			expected: 123.00,
			err:      nil,
		},
		"大きな数値を正確に切り上げる": {
			input:    123456.789,
			expected: 123456.79,
			err:      nil,
		},
		"小さな数値を正確に切り上げる": {
			input:    0.004,
			expected: 0.01,
			err:      nil,
		},
		"負の数を正確に切り上げる": {
			input:    -123.456,
			expected: 0,
			err:      errors.New("negative values are not allowed"),
		},
		"ゼロをそのまま処理する": {
			input:    0.0,
			expected: 0.00,
			err:      nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := calc.RoundUpToTwoDecimalPlaces(tt.input)
			assert.Equal(t, result, tt.expected)
			if tt.err != nil {
				assert.Equal(t, err, tt.err)
			}
		})
	}
}
