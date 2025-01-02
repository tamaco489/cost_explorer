package service

import (
	"fmt"

	"github.com/tamaco489/cost_explorer/batch/internal/library/slack"
)

// WeeklyCostUsage: 週次レポートに必要な要素を含む構造体
type WeeklyCostUsage struct {
	LastWeekCost       float64
	WeekBeforeLastCost float64
	PercentageChange   float64
}

// newWeeklyCostUsage: WeeklyCostUsage のコンストラクタ
func (wcs *WeeklyCostExplorerService) NewWeeklyCostUsage(lastWeekCost, weekBeforeLastCost, percentageChange float64) *WeeklyCostUsage {
	return &WeeklyCostUsage{
		LastWeekCost:       lastWeekCost,
		WeekBeforeLastCost: weekBeforeLastCost,
		PercentageChange:   percentageChange,
	}
}

// genSlackMessage: 週次利用コストレポートのメッセージを生成
func (wcu *WeeklyCostUsage) GenWeeklySlackMessage() slack.Attachment {
	return slack.Attachment{
		Pretext: fmt.Sprintf(`
• 先週の利用コスト: %.2f 円
• 先々週の利用コスト: %.2f 円
• 先々週のコストに対する先週のコスト: %.2f %%`,
			wcu.LastWeekCost, wcu.WeekBeforeLastCost, wcu.PercentageChange,
		),
	}
}
