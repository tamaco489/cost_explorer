package service

import (
	"fmt"

	"github.com/tamaco489/cost_explorer/batch/internal/library/calc"
	"github.com/tamaco489/cost_explorer/batch/internal/library/exchange_rates"
)

// calcDailyCostInJPY: Open Exchange Rates APIのレスポンスから1$あたりの円を取得し、そのレートを使用して利用コストをUSDからJPYに変換
func (dcu *DailyCostUsage) CalcDailyCostInJPY(res *exchange_rates.ExchangeRatesResponse) (*DailyCostUsage, error) {
	rate, ok := res.Rates[exchange_rates.JPY.String()]
	if !ok {
		return nil, fmt.Errorf("JPY exchange rate not found in the response: %+v", res.Rates)
	}

	// calc.RoundUpToTwoDecimalPlaces のエラーをチェック
	yesterdayCost, err := calc.RoundUpToTwoDecimalPlaces(dcu.YesterdayCost * rate)
	if err != nil {
		return nil, fmt.Errorf("error rounding up YesterdayCost: %v", err)
	}

	actualCost, err := calc.RoundUpToTwoDecimalPlaces(dcu.ActualCost * rate)
	if err != nil {
		return nil, fmt.Errorf("error rounding up ActualCost: %v", err)
	}

	forecastCost, err := calc.RoundUpToTwoDecimalPlaces(dcu.ForecastCost * rate)
	if err != nil {
		return nil, fmt.Errorf("error rounding up ForecastCost: %v", err)
	}

	return &DailyCostUsage{
		YesterdayCost: yesterdayCost, // 昨日利用したコスト
		ActualCost:    actualCost,    // 本日時点で利用した総コスト
		ForecastCost:  forecastCost,  // 残り日数を考慮した今月の利用コスト
	}, nil
}

// calcWeeklyCostInJPY: Open Exchange Rates APIのレスポンスから1$あたりの円を取得し、そのレートを使用して利用コストをUSDからJPYに変換
func (wcu *WeeklyCostUsage) CalcWeeklyCostInJPY(res *exchange_rates.ExchangeRatesResponse) (*WeeklyCostUsage, error) {
	rate, ok := res.Rates[exchange_rates.JPY.String()]
	if !ok {
		return nil, fmt.Errorf("JPY exchange rate not found in the response: %+v", res.Rates)
	}

	// calc.RoundUpToTwoDecimalPlaces のエラーをチェック
	lastWeekCost, err := calc.RoundUpToTwoDecimalPlaces(wcu.LastWeekCost * rate)
	if err != nil {
		return nil, fmt.Errorf("error rounding up LastWeekCost: %v", err)
	}

	weekBeforeLastCost, err := calc.RoundUpToTwoDecimalPlaces(wcu.WeekBeforeLastCost * rate)
	if err != nil {
		return nil, fmt.Errorf("error rounding up WeekBeforeLastCost: %v", err)
	}

	return &WeeklyCostUsage{
		LastWeekCost:       lastWeekCost,         // 先週利用したコスト
		WeekBeforeLastCost: weekBeforeLastCost,   // 先々週利用した総コスト
		PercentageChange:   wcu.PercentageChange, // 先週と先々週のコスト増減（%）
	}, nil
}
