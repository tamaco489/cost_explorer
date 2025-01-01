package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/tamaco489/cost_explorer/batch/internal/configuration"
	"github.com/tamaco489/cost_explorer/batch/internal/library/slack"
)

func (j *Job) DailyCostReport(ctx context.Context) error {

	if j.execTime.Day() == 1 {
		slog.InfoContext(ctx, "no processing is performed at the beginning of the month")
		return nil
	}

	// 日時情報オブジェクトの生成
	fmtDate := newDailyCostReportDateFormate(j.execTime)

	slog.InfoContext(ctx, "Generated ReportDateInfo",
		slog.String("startDate", fmtDate.StartDate),
		slog.String("yesterday", fmtDate.Yesterday),
		slog.String("endDate", fmtDate.EndDate),
		slog.Int("currentDay", fmtDate.CurrentDay),
		slog.Int("daysInMonth", fmtDate.DaysInMonth),
	)

	yesterdayCost, err := j.getYesterdayCost(ctx, fmtDate.Yesterday, fmtDate.EndDate)
	if err != nil {
		return fmt.Errorf("failed to get yesterday cost: %w", err)
	}

	actualCost, err := j.getActualCost(ctx, fmtDate.StartDate, fmtDate.EndDate)
	if err != nil {
		return fmt.Errorf("failed to get actual cost: %w", err)
	}

	forecastCost, err := j.getForecastCost(actualCost, fmtDate.CurrentDay, fmtDate.DaysInMonth)
	if err != nil {
		return fmt.Errorf("failed to get forecast cost: %w", err)
	}

	report := newDailySlackReport(yesterdayCost, actualCost, forecastCost)
	message := report.genSlackMessage()

	sc := slack.NewSlackClient(configuration.Get().Slack.DailyWebHookURL, configuration.Get().ServiceName)
	if err := sc.SendMessage(ctx, slack.DailyCostReportTitle.String(), message); err != nil {
		return fmt.Errorf("failed to send slack message: %w", err)
	}

	return nil
}

// dailyCostReportDateFormate: 日次コストレポートのための日時情報を保持する構造体
type dailyCostReportDateFormate struct {
	StartDate   string // 今月の開始日付
	Yesterday   string // 昨日の日付
	EndDate     string // 今月の終了日付
	CurrentDay  int    // 今日までの日数
	DaysInMonth int    // 今月の総日数
}

// newDailyCostReportDateFormate: ReportDateInfoのコンストラクタ
func newDailyCostReportDateFormate(execTime time.Time) dailyCostReportDateFormate {
	currentYear, currentMonth, _ := execTime.Date()
	daysInMonth := time.Date(currentYear, currentMonth+1, 0, 0, 0, 0, 0, time.UTC).Day()

	return dailyCostReportDateFormate{
		StartDate:   time.Date(execTime.Year(), execTime.Month(), 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02"),
		Yesterday:   execTime.AddDate(0, 0, -1).Format("2006-01-02"),
		EndDate:     execTime.Format("2006-01-02"),
		CurrentDay:  execTime.Day(),
		DaysInMonth: daysInMonth,
	}
}

// getYesterdayCost: 昨日の利用コストを取得する
func (j *Job) getYesterdayCost(ctx context.Context, yesterday, endDate string) (string, error) {

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
func (j *Job) getActualCost(ctx context.Context, startDate, endDate string) (string, error) {

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
func (j *Job) getForecastCost(actualCost string, currentDay, daysInMonth int) (string, error) {

	// コスト計算
	actualCostFloat, err := j.parseCost(actualCost)
	if err != nil {
		return "", err
	}

	// 1日あたりの平均コスト
	averageCostPerDay := actualCostFloat / float64(currentDay)

	// 予測コスト
	forecastCost := averageCostPerDay * float64(daysInMonth)

	return fmt.Sprintf("%.2f", forecastCost), nil
}

// dailySlackReport: 日次利用コストレポート向けのメッセージを生成するための構造体
type dailySlackReport struct {
	yesterdayCost string
	actualCost    string
	forecastCost  string
}

// newDailySlackReport: 日次利用コストレポートを作成するためのコンストラクタ関数
func newDailySlackReport(yesterdayCost, actualCost, forecastCost string) dailySlackReport {
	return dailySlackReport{
		yesterdayCost: yesterdayCost,
		actualCost:    actualCost,
		forecastCost:  forecastCost,
	}
}

// genSlackMessage: 日次利用コストレポートのメッセージを生成する
func (r dailySlackReport) genSlackMessage() slack.Attachment {
	return slack.Attachment{
		Pretext: fmt.Sprintf(`
• 昨日の利用コスト: %s USD
• 本日時点での今月の利用コスト: %s USD
• 今月の利用コストの予測値: %s USD
`, r.yesterdayCost, r.actualCost, r.forecastCost,
		),
	}
}
