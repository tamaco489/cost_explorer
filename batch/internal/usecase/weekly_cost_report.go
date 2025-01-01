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

func (j *Job) WeeklyCostReport(ctx context.Context) error {

	// NOTE: 検証用途として一時的に日付を書き換える
	// j.execTime = time.Date(2024, 12, 29, 0, 0, 0, 0, time.UTC)

	fd := newFormattedDateForWeeklyReport(j.execTime)

	lastWeekCost, err := j.getLastWeekCost(ctx, fd.lastWeekStartDate, fd.lastWeekEndDate)
	if err != nil {
		return fmt.Errorf("failed to get last week cost: %w", err)
	}

	weekBeforeLastCost, err := j.getWeekBeforeLastCost(ctx, fd.weekBeforeLastStartDate, fd.weekBeforeLastEndDate)
	if err != nil {
		return fmt.Errorf("failed to get week before last cost: %w", err)
	}

	percentageChange := j.calculatePercentageChange(lastWeekCost, weekBeforeLastCost)

	report := newWeeklySlackReport(lastWeekCost, weekBeforeLastCost, percentageChange)
	message := report.genSlackMessage()

	sc := slack.NewSlackClient(configuration.Get().Slack.WeeklyWebHookURL, configuration.Get().ServiceName)
	if err := sc.SendMessage(ctx, slack.WeeklyCostReportTitle.String(), message); err != nil {
		return fmt.Errorf("failed to send slack message: %w", err)
	}

	return nil
}

// formattedDateForWeeklyReport: 週次コストレポートのための日時情報を保持する構造体
type formattedDateForWeeklyReport struct {
	lastWeekStartDate       string // 先週の開始日付
	lastWeekEndDate         string // 先週の終了日付
	weekBeforeLastStartDate string // 先々週の開始日付
	weekBeforeLastEndDate   string // 先々週の終了日付
}

// newFormattedDateForWeeklyReport: formattedDateForWeeklyReport のコンストラクタ
func newFormattedDateForWeeklyReport(execTime time.Time) formattedDateForWeeklyReport {
	return formattedDateForWeeklyReport{
		lastWeekStartDate:       execTime.AddDate(0, 0, -13).Format("2006-01-02"),
		lastWeekEndDate:         execTime.AddDate(0, 0, -7).Format("2006-01-02"),
		weekBeforeLastStartDate: execTime.AddDate(0, 0, -20).Format("2006-01-02"),
		weekBeforeLastEndDate:   execTime.AddDate(0, 0, -14).Format("2006-01-02"),
	}
}

// getLastWeekCost: 先週の利用コストを取得する
func (j *Job) getLastWeekCost(ctx context.Context, lastWeekStartDate, lastWeekEndDate string) (string, error) {

	output, err := j.costExplorerClient.GetCostAndUsage(ctx, &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: &lastWeekStartDate,
			End:   &lastWeekEndDate,
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
func (j *Job) getWeekBeforeLastCost(ctx context.Context, weekBeforeLastStartDate, weekBeforeLastEndDate string) (string, error) {

	output, err := j.costExplorerClient.GetCostAndUsage(ctx, &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: &weekBeforeLastStartDate,
			End:   &weekBeforeLastEndDate,
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
func (j *Job) calculatePercentageChange(lastWeekCost, weekBeforeLastCost string) string {

	lastWeek, err := j.parseCost(lastWeekCost)
	if err != nil {
		log.Printf("failed to parse last week cost: %v", err)
		return "0.0"
	}

	weekBeforeLast, err := j.parseCost(weekBeforeLastCost)
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

// weeklyCostReport: 週次利用コストレポート向けのメッセージを生成するための構造体
type weeklySlackReport struct {
	lastWeekCost       string
	weekBeforeLastCost string
	percentageChange   string
}

// newWeeklySlackReport: 週次コストレポートを作成するためのコンストラクタ関数
func newWeeklySlackReport(lastWeekCost, weekBeforeLastCost, percentageChange string) weeklySlackReport {
	return weeklySlackReport{
		lastWeekCost:       lastWeekCost,
		weekBeforeLastCost: weekBeforeLastCost,
		percentageChange:   percentageChange,
	}
}

// genSlackMessage: 週次利用コストレポートのメッセージを生成する
func (r weeklySlackReport) genSlackMessage() slack.Attachment {
	return slack.Attachment{
		Pretext: fmt.Sprintf(`
• 先週の利用コスト: %s USD
• 先々週の利用コスト: %s USD
• 先週と先々週のコスト増減: %s %%`,
			r.lastWeekCost, r.weekBeforeLastCost, r.percentageChange,
		),
	}
}
