package calc

import "math"

// RoundUpToTwoDecimalPlaces: float64 の値を小数点以下2桁で切り上げる
func RoundUpToTwoDecimalPlaces(value float64) float64 {
	factor := math.Pow(10, 2)
	return math.Ceil(value*factor) / factor
}
