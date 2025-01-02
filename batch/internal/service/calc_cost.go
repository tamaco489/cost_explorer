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

	return &DailyCostUsage{
		YesterdayCost: calc.RoundUpToTwoDecimalPlaces(dcu.YesterdayCost * rate), // 昨日利用したコスト
		ActualCost:    calc.RoundUpToTwoDecimalPlaces(dcu.ActualCost * rate),    // 本日時点で利用した総コスト
		ForecastCost:  calc.RoundUpToTwoDecimalPlaces(dcu.ForecastCost * rate),  // 残り日数を考慮した今月の利用コスト
	}, nil
}

// calcWeeklyCostInJPY: Open Exchange Rates APIのレスポンスから1$あたりの円を取得し、そのレートを使用して利用コストをUSDからJPYに変換
func (wcu *WeeklyCostUsage) CalcWeeklyCostInJPY(res *exchange_rates.ExchangeRatesResponse) (*WeeklyCostUsage, error) {
	rate, ok := res.Rates[exchange_rates.JPY.String()]
	if !ok {
		return nil, fmt.Errorf("JPY exchange rate not found in the response: %+v", res.Rates)
	}

	return &WeeklyCostUsage{
		LastWeekCost:       calc.RoundUpToTwoDecimalPlaces(wcu.LastWeekCost * rate),       // 先週利用したコスト
		WeekBeforeLastCost: calc.RoundUpToTwoDecimalPlaces(wcu.WeekBeforeLastCost * rate), // 先々週利用した総コスト
		PercentageChange:   wcu.PercentageChange,                                          // 先週と先々週のコスト増減（%）
	}, nil
}
