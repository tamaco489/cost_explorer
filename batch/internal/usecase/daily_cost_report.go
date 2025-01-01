package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/tamaco489/cost_explorer/batch/internal/configuration"
	"github.com/tamaco489/cost_explorer/batch/internal/library/exchange_rates"
	"github.com/tamaco489/cost_explorer/batch/internal/library/slack"
)

func (j *Job) DailyCostReport(ctx context.Context) error {

	// ************************* 1. 月初かどうかの判定 *************************
	// NOTE: 検証用途として一時的に日付を書き換える
	j.execTime = time.Date(2024, 12, 29, 0, 0, 0, 0, time.UTC)

	// 月初の場合は処理をスキップ
	if j.execTime.Day() == 1 {
		slog.InfoContext(ctx, "no processing is performed at the beginning of the month")
		return nil
	}

	fd := newFormattedDateForDailyReport(j.execTime)

	// NOTE: debug log
	fd.formattedDateLogs(ctx)

	// ************************* 2. AWS 利用コストの算出 *************************
	yesterdayCost, err := j.getYesterdayCost(ctx, fd.yesterday, fd.endDate)
	if err != nil {
		return fmt.Errorf("failed to get yesterday cost: %w", err)
	}

	actualCost, err := j.getActualCost(ctx, fd.startDate, fd.endDate)
	if err != nil {
		return fmt.Errorf("failed to get actual cost: %w", err)
	}

	forecastCost, err := j.getForecastCost(actualCost, fd.currentDay, fd.daysInMonth)
	if err != nil {
		return fmt.Errorf("failed to get forecast cost: %w", err)
	}

	// NOTE: debug log
	getUsageCostLogs(ctx, yesterdayCost, actualCost, forecastCost)

	// ************************* 3. Open Exchange Rate API を使用して、現時点の為替レートを取得 *************************
	pxr, err := j.exchangeRatesClient.PrepareExchangeRates()
	if err != nil {
		return err
	}

	ratesResponse, err := j.exchangeRatesClient.GetExchangeRates(pxr.BaseCurrencyCode, pxr.ExchangeCurrencyCodes)
	if err != nil {
		return fmt.Errorf("failed to get exchange rates: %w", err)
	}

	// NOTE: debug log
	exchangeRatesResponseLogs(ctx, ratesResponse)

	// ************************* 4. 取得した為替レートを利用して、利用コストをUSDからJPYに変換 *************************
	yesterdayCostFloat, err := strconv.ParseFloat(yesterdayCost, 64)
	if err != nil {
		return fmt.Errorf("failed to parse yesterdayCost to float64: %w", err)
	}

	actualCostFloat, err := strconv.ParseFloat(actualCost, 64)
	if err != nil {
		return fmt.Errorf("failed to parse actualCost to float64: %w", err)
	}

	forecastCostFloat, err := strconv.ParseFloat(forecastCost, 64)
	if err != nil {
		return fmt.Errorf("failed to parse forecast cost: %w", err)
	}

	// 1$あたりの円を取得
	usdToJpyRate, ok := ratesResponse.Rates[exchange_rates.JPY.String()]
	if !ok {
		return fmt.Errorf("JPY exchange rate not found in the response: %+v", ratesResponse.Rates)
	}

	// 小数点第二位を切り上げ
	yesterdayCostJPY := roundUpToTwoDecimalPlaces(yesterdayCostFloat * usdToJpyRate) // 昨日利用したコスト
	actualCostJPY := roundUpToTwoDecimalPlaces(actualCostFloat * usdToJpyRate)       // 本日時点で利用した総コスト
	forecastCostJPY := roundUpToTwoDecimalPlaces(forecastCostFloat * usdToJpyRate)   // 残り日数を考慮した今月の利用コスト

	// NOTE: debug log
	parsedJPYCostLogs(ctx, yesterdayCostJPY, actualCostJPY, forecastCostJPY)

	// ************************* 5. Slackにメッセージを送信する *************************
	report := newDailySlackReport(yesterdayCostJPY, actualCostJPY, forecastCostJPY)
	message := report.genSlackMessage()

	sc := slack.NewSlackClient(configuration.Get().Slack.DailyWebHookURL, configuration.Get().ServiceName)
	if err := sc.SendMessage(ctx, slack.DailyReportTitle.String(), message); err != nil {
		return fmt.Errorf("failed to send slack message: %w", err)
	}

	return nil
}

// formattedDateForDailyReport: 日次コストレポートのための日時情報を保持する構造体
type formattedDateForDailyReport struct {
	yesterday   string // 昨日の日付
	startDate   string // 今月の開始日付
	endDate     string // 今月の終了日付
	currentDay  int    // 今日までの日数
	daysInMonth int    // 今月の総日数
}

// newFormattedDateForDailyReport: formattedDateForDailyReport のコンストラクタ
func newFormattedDateForDailyReport(execTime time.Time) formattedDateForDailyReport {
	currentYear, currentMonth, _ := execTime.Date()
	daysInMonth := time.Date(currentYear, currentMonth+1, 0, 0, 0, 0, 0, time.UTC).Day()

	return formattedDateForDailyReport{
		startDate:   time.Date(execTime.Year(), execTime.Month(), 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02"),
		yesterday:   execTime.AddDate(0, 0, -1).Format("2006-01-02"),
		endDate:     execTime.Format("2006-01-02"),
		currentDay:  execTime.Day(),
		daysInMonth: daysInMonth,
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
func newDailySlackReport(yesterdayCost, actualCost, forecastCost float64) dailySlackReport {
	return dailySlackReport{
		yesterdayCost: fmt.Sprintf("%.2f", yesterdayCost),
		actualCost:    fmt.Sprintf("%.2f", actualCost),
		forecastCost:  fmt.Sprintf("%.2f", forecastCost),
	}
}

// genSlackMessage: 日次利用コストレポートのメッセージを生成する
func (r dailySlackReport) genSlackMessage() slack.Attachment {
	return slack.Attachment{
		Pretext: fmt.Sprintf(`
• 昨日の利用コスト: %s 円
• 本日時点での今月の利用コスト: %s 円
• 今月の利用コストの予測値: %s 円
`, r.yesterdayCost, r.actualCost, r.forecastCost,
		),
	}
}

// NOTE: debug用のログ
func (fd formattedDateForDailyReport) formattedDateLogs(ctx context.Context) {
	slog.InfoContext(ctx, "[1]. outputFormattedDateList:",
		slog.String("昨日の日付", fd.yesterday),   // 2024-12-28
		slog.String("今月の開始日付", fd.startDate), // 2024-12-01
		slog.String("今月の終了日付", fd.endDate),   // 2024-12-29
		slog.Int("今日までの日数", fd.currentDay),   // 29
		slog.Int("今月の総日数", fd.daysInMonth),   // 31
	)
}

// NOTE: debug用のログ
func getUsageCostLogs(ctx context.Context, yesterdayCost, actualCost, forecastCost string) {
	slog.InfoContext(ctx, "[2] get usage cost",
		slog.String("yesterday", yesterdayCost), // 0.0217344233
		slog.String("actual", actualCost),       // 0.7277853673
		slog.String("forecast", forecastCost),   // 0.78
	)
}

// NOTE: debug用のログ
func exchangeRatesResponseLogs(ctx context.Context, r *exchange_rates.ExchangeRatesResponse) {
	slog.InfoContext(ctx, "[3]. get exchange rates api response",
		slog.Float64("EUR", r.Rates["EUR"]), // 0.966185
		slog.Float64("JPY", r.Rates["JPY"]), // 157.35784932
	)
}

// NOTE: debug用のログ
func parsedJPYCostLogs(ctx context.Context, yesterdayCostJPY, actualCostJPY, forecastCostJPY float64) {
	slog.InfoContext(ctx, "[4] parsed jpy cost",
		slog.Float64("yesterday", yesterdayCostJPY), // 3.4200821066984974
		slog.Float64("actual", actualCostJPY),       // 114.52274016489426
		slog.Float64("forecast", forecastCostJPY),   // 122.73912246960002
	)
}
