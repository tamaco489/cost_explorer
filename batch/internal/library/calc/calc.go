package calc

import (
	"errors"
	"math"
)

// RoundUpToTwoDecimalPlaces: float64 の値を小数点以下2桁で切り上げる
//
// 負の値が入力された場合はエラーを返す
func RoundUpToTwoDecimalPlaces(value float64) (float64, error) {
	// 負の値が入力された場合はエラーを返す
	if value < 0 {
		return 0, errors.New("negative values are not allowed")
	}

	// 小数点以下2桁で切り上げる
	factor := math.Pow(10, 2)
	result := math.Ceil(value*factor) / factor

	return result, nil
}
