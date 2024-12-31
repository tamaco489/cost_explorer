package usecase

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/tamaco489/cost_explorer/batch/internal/configuration"
	"github.com/tamaco489/cost_explorer/batch/internal/library/slack"
)

func (j *Job) WeeklyCostReport(ctx context.Context) error {

	lastWeekCost, err := j.getLastWeekCost(ctx)
	if err != nil {
		return fmt.Errorf("failed to get last week cost: %w", err)
	}

	weekBeforeLastCost, err := j.getWeekBeforeLastCost(ctx)
	if err != nil {
		return fmt.Errorf("failed to get week before last cost: %w", err)
	}

	percentageChange := calculatePercentageChange(lastWeekCost, weekBeforeLastCost)

	report := newWeeklyCostReport(lastWeekCost, weekBeforeLastCost, percentageChange)
	message := report.genSlackMessage()

	const title = "weekly-cost-report"
	sc := slack.NewSlackClient(configuration.Get().Slack.WeeklyWebHookURL, configuration.Get().ServiceName)
	if err := sc.SendMessage(ctx, title, message); err != nil {
		return fmt.Errorf("failed to send slack message: %w", err)
	}

	return nil
}

// getLastWeekCost: 先週の利用コストを取得する
func (j *Job) getLastWeekCost(ctx context.Context) (string, error) {
	// 先週の開始日と終了日を計算
	endDate := j.execTime.AddDate(0, 0, -7).Format("2006-01-02")
	startDate := j.execTime.AddDate(0, 0, -13).Format("2006-01-02")

	output, err := j.costExplorerClient.GetCostAndUsage(ctx, &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: &startDate,
			End:   &endDate,
		},
		Metrics:     []string{"UnblendedCost"},
		Granularity: types.GranularityDaily,
	})
	if err != nil {
		return "", err
	}

	totalCost := 0.0
	for _, result := range output.ResultsByTime {
		if cost, ok := result.Total["UnblendedCost"]; ok {
			costFloat, err := strconv.ParseFloat(*cost.Amount, 64)
			if err != nil {
				return "", err
			}
			totalCost += costFloat
		}
	}

	return fmt.Sprintf("%.2f", totalCost), nil
}

// getWeekBeforeLastCost: 先々週の利用コストを取得する
func (j *Job) getWeekBeforeLastCost(ctx context.Context) (string, error) {
	// 先々週の開始日と終了日を計算
	endDate := j.execTime.AddDate(0, 0, -14).Format("2006-01-02")
	startDate := j.execTime.AddDate(0, 0, -20).Format("2006-01-02")

	output, err := j.costExplorerClient.GetCostAndUsage(ctx, &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: &startDate,
			End:   &endDate,
		},
		Metrics:     []string{"UnblendedCost"},
		Granularity: types.GranularityDaily,
	})
	if err != nil {
		return "", err
	}

	totalCost := 0.0
	for _, result := range output.ResultsByTime {
		if cost, ok := result.Total["UnblendedCost"]; ok {
			costFloat, err := strconv.ParseFloat(*cost.Amount, 64)
			if err != nil {
				return "", err
			}
			totalCost += costFloat
		}
	}

	return fmt.Sprintf("%.2f", totalCost), nil
}

// calculatePercentageChange: コストの増減率を計算する
func calculatePercentageChange(lastWeekCost, weekBeforeLastCost string) string {
	lastWeek, err := parseCost(lastWeekCost)
	if err != nil {
		log.Printf("failed to parse last week cost: %v", err)
		return "0.0"
	}

	weekBeforeLast, err := parseCost(weekBeforeLastCost)
	if err != nil {
		log.Printf("failed to parse week before last cost: %v", err)
		return "0.0"
	}

	if weekBeforeLast == 0 {
		return "0.0"
	}

	change := ((lastWeek - weekBeforeLast) / weekBeforeLast) * 100
	return fmt.Sprintf("%.2f", change)
}

// newWeeklyCostReport: 週次コストレポートを作成するためのコンストラクタ関数
func newWeeklyCostReport(lastWeekCost, weekBeforeLastCost, percentageChange string) weeklyCostReport {
	return weeklyCostReport{
		lastWeekCost:       lastWeekCost,
		weekBeforeLastCost: weekBeforeLastCost,
		percentageChange:   percentageChange,
	}
}

// weeklyCostReport: 週次利用コストレポート向けのメッセージを生成するための構造体
type weeklyCostReport struct {
	lastWeekCost       string
	weekBeforeLastCost string
	percentageChange   string
}

// genSlackMessage: 週次利用コストレポートのメッセージを生成する
func (r weeklyCostReport) genSlackMessage() slack.Attachment {
	return slack.Attachment{
		Pretext: fmt.Sprintf(`
• 先週の利用コスト: %s USD
• 先々週の利用コスト: %s USD
• 先週と先々週のコスト増減: %s %%`,
			r.lastWeekCost, r.weekBeforeLastCost, r.percentageChange,
		),
	}
}
