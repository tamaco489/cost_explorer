package service

import (
	"fmt"

	"github.com/tamaco489/cost_explorer/batch/internal/library/slack"
)

// DailyCostUsage: 日次レポートに必要な要素を含む構造体。
type DailyCostUsage struct {
	YesterdayCost float64
	ActualCost    float64
	ForecastCost  float64
}

// NewDailyCostUsage: DailyCostUsage のコンストラクタ
func (dcs *DailyCostExplorerService) NewDailyCostUsage(yesterdayCost, actualCost, forecastCost float64) *DailyCostUsage {
	return &DailyCostUsage{
		YesterdayCost: yesterdayCost,
		ActualCost:    actualCost,
		ForecastCost:  forecastCost,
	}
}

// genSlackMessage: 日次利用コストレポートのメッセージを生成
func (dcu DailyCostUsage) GenDailySlackMessage() slack.Attachment {
	return slack.Attachment{
		Pretext: fmt.Sprintf(`
• 昨日の利用コスト: %.2f 円
• 本日時点での今月の利用コスト: %.2f 円
• 今月の利用コストの予測値: %.2f 円
`, dcu.YesterdayCost, dcu.ActualCost, dcu.ForecastCost,
		),
	}
}
