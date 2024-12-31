package usecase

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/tamaco489/cost_explorer/batch/internal/configuration"
	"github.com/tamaco489/cost_explorer/batch/internal/library/slack"
)

func (j *Job) DailyCostReport(ctx context.Context) error {

	yesterdayCost, err := j.getYesterdayCost(ctx)
	if err != nil {
		return fmt.Errorf("failed to get yesterday cost: %w", err)
	}

	actualCost, err := j.getActualCost(ctx)
	if err != nil {
		return fmt.Errorf("failed to get actual cost: %w", err)
	}

	forecastCost, err := j.getForecastCost(actualCost)
	if err != nil {
		return fmt.Errorf("failed to get forecast cost: %w", err)
	}

	report := newDailyCostReport(yesterdayCost, actualCost, forecastCost)
	message := report.genSlackMessage()

	const title = "daily-cost-report"
	sc := slack.NewSlackClient(configuration.Get().Slack.DailyWebHookURL, configuration.Get().ServiceName)
	if err := sc.SendMessage(ctx, title, message); err != nil {
		return fmt.Errorf("failed to send slack message: %w", err)
	}

	return nil
}

// getYesterdayCost: 昨日の利用コストを取得する
func (j *Job) getYesterdayCost(ctx context.Context) (string, error) {
	// 昨日の開始日と終了日を計算
	yesterday := j.execTime.AddDate(0, 0, -1).Format("2006-01-02")
	endDate := j.execTime.Format("2006-01-02")

	log.Println("[1] getYesterdayCost |", "昨日:", yesterday, "現在日:", endDate)

	output, err := j.costExplorerClient.GetCostAndUsage(ctx, &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: &yesterday,
			End:   &endDate,
		},
		Metrics:     []string{"UnblendedCost"},
		Granularity: types.GranularityDaily,
	})
	if err != nil {
		return "", err
	}

	if len(output.ResultsByTime) > 0 && len(output.ResultsByTime[0].Total) > 0 {
		if cost, ok := output.ResultsByTime[0].Total["UnblendedCost"]; ok {
			return *cost.Amount, nil
		}
	}

	return "0.0", nil
}

// getActualCost: 本日時点での今月の利用コストを取得する
func (j *Job) getActualCost(ctx context.Context) (string, error) {
	// 今月の開始日と現在日
	startDate := time.Date(j.execTime.Year(), j.execTime.Month(), 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	endDate := j.execTime.Format("2006-01-02")

	log.Println("[2] getCost |", "今月の開始日:", startDate, "現在日:", endDate)

	output, err := j.costExplorerClient.GetCostAndUsage(ctx, &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: &startDate,
			End:   &endDate,
		},
		Metrics:     []string{"UnblendedCost"},
		Granularity: types.GranularityMonthly,
	})
	if err != nil {
		return "", err
	}

	if len(output.ResultsByTime) > 0 && len(output.ResultsByTime[0].Total) > 0 {
		if cost, ok := output.ResultsByTime[0].Total["UnblendedCost"]; ok {
			return *cost.Amount, nil
		}
	}

	return "0.0", nil
}

// getForecastCost: 今月の利用コストの予測値を算出する
func (j *Job) getForecastCost(actualCost string) (string, error) {
	// 今月の総日数
	currentYear, currentMonth, _ := j.execTime.Date()
	daysInMonth := time.Date(currentYear, currentMonth+1, 0, 0, 0, 0, 0, time.UTC).Day()

	// 今日までの日数
	currentDay := j.execTime.Day()

	// コスト計算
	actualCostFloat, err := parseCost(actualCost)
	if err != nil {
		return "", err
	}

	// 1日あたりの平均コスト
	averageCostPerDay := actualCostFloat / float64(currentDay)

	// 予測コスト
	forecastCost := averageCostPerDay * float64(daysInMonth)

	log.Println("[3] getForecastCost |", "今月の総日数:", daysInMonth, "今日までの日数:", currentDay)
	log.Println("[4] getForecastCost |", "1日あたりの平均コスト:", averageCostPerDay, "予測コスト:", forecastCost)

	return fmt.Sprintf("%.2f", forecastCost), nil
}

// parseCost: 文字列を float64 に変換する
func parseCost(cost string) (float64, error) {
	return strconv.ParseFloat(cost, 64)
}

// dailyCostReport: 日次利用コストレポート向けのメッセージを生成するための構造体
type dailyCostReport struct {
	yesterdayCost string
	actualCost    string
	forecastCost  string
}

// newDailyCostReport: 日次利用コストレポートを作成するためのコンストラクタ関数
func newDailyCostReport(yesterdayCost, actualCost, forecastCost string) dailyCostReport {
	return dailyCostReport{
		yesterdayCost: yesterdayCost,
		actualCost:    actualCost,
		forecastCost:  forecastCost,
	}
}

// genSlackMessage: 日次利用コストレポートのメッセージを生成する
func (r dailyCostReport) genSlackMessage() slack.Attachment {
	return slack.Attachment{
		Pretext: fmt.Sprintf(`
• 昨日の利用コスト: %s USD
• 本日時点での今月の利用コスト: %s USD
• 今月の利用コストの予測値: %s USD
`, r.yesterdayCost, r.actualCost, r.forecastCost,
		),
	}
}
