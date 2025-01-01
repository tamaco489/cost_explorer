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

	// ************************* 1. 実行日時からコスト算出に必要な各基準日を取得 (月初の場合は処理をスキップ) *************************
	if j.execTime.Day() == 1 {
		slog.InfoContext(ctx, "no processing is performed at the beginning of the month")
		return nil
	}

	fd := newFormattedDateForDailyReport(j.execTime)

	if configuration.Get().Logging == "on" {
		fd.formattedDateLogs(ctx)
	}

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

	if configuration.Get().Logging == "on" {
		getUsageCostLogs(ctx, yesterdayCost, actualCost, forecastCost)
	}

	// ************************* 3. Open Exchange Rates API を使用して、為替レートを取得 *************************
	pxr, err := j.exchangeRatesClient.PrepareExchangeRates()
	if err != nil {
		return err
	}

	ratesResponse, err := j.exchangeRatesClient.GetExchangeRates(pxr.BaseCurrencyCode, pxr.ExchangeCurrencyCodes)
	if err != nil {
		return fmt.Errorf("failed to get exchange rates: %w", err)
	}

	if configuration.Get().Logging == "on" {
		exchangeRatesResponseLogs(ctx, ratesResponse)
	}

	// ************************* 4. 取得した為替レートを利用して、利用コストをUSDからJPYに変換 *************************
	costUsage := newDailyCostUsage(yesterdayCost, actualCost, forecastCost)
	parseUsage, err := costUsage.parseDailyCostUsage()
	if err != nil {
		return err
	}

	jpyUsage, err := parseUsage.calcDailyCostInJPY(ratesResponse)
	if err != nil {
		return err
	}

	if configuration.Get().Logging == "on" {
		parseJPYCostLogs(ctx, jpyUsage.yesterdayCost, jpyUsage.actualCost, jpyUsage.forecastCost)
	}

	// ************************* 5. Slackにメッセージを送信する *************************
	message := jpyUsage.genSlackMessage()
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
//
// 実行日時からコスト算出に必要な各基準日を取得
//
// yesterday: 昨日の日付 (string)
//
// startDate: 今月の開始日付 (string)
//
// endDate: 今月の終了日付 (string)
//
// currentDay: 今日までの日数 (int)
//
// daysInMonth: 今月の総日数 (int)
func newFormattedDateForDailyReport(execTime time.Time) formattedDateForDailyReport {
	currentYear, currentMonth, _ := execTime.Date()
	daysInMonth := time.Date(currentYear, currentMonth+1, 0, 0, 0, 0, 0, time.UTC).Day()

	return formattedDateForDailyReport{
		yesterday:   execTime.AddDate(0, 0, -1).Format("2006-01-02"),
		startDate:   time.Date(execTime.Year(), execTime.Month(), 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02"),
		endDate:     execTime.Format("2006-01-02"),
		currentDay:  execTime.Day(),
		daysInMonth: daysInMonth,
	}
}

// getYesterdayCost: 昨日の利用コストを取得
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

// getActualCost: 本日時点での今月の利用コストを取得
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

// getForecastCost: 今月の利用コストの予測値を算出
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

// dailyCostUsage: 日次レポートに必要な要素を含む構造体。
//
// AWS Cost Explorer SDKの戻り値が全てstring型のため、各パラメータは全てstring型として定義
type dailyCostUsage struct {
	yesterdayCost string
	actualCost    string
	forecastCost  string
}

// newDailyCostUsage: dailyCostUsage のコンストラクタ
func newDailyCostUsage(yesterdayCost, actualCost, forecastCost string) *dailyCostUsage {
	return &dailyCostUsage{
		yesterdayCost: yesterdayCost,
		actualCost:    actualCost,
		forecastCost:  forecastCost,
	}
}

// parseDailyCostUsage: dailyCostUsage の各パラメータをfloat64型にした構造体
//
// 為替レートに計算において、小数点を考慮する必要があるため、float型として定義
type parseDailyCostUsage struct {
	yesterdayCost float64
	actualCost    float64
	forecastCost  float64
}

// parseDailyCostUsage: 利用コストをfloat64型に変換
func (dcu *dailyCostUsage) parseDailyCostUsage() (*parseDailyCostUsage, error) {
	yesterdayCost, err := strconv.ParseFloat(dcu.yesterdayCost, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse yesterdayCost to float64: %w", err)
	}

	actualCost, err := strconv.ParseFloat(dcu.actualCost, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse actualCost to float64: %w", err)
	}

	forecastCost, err := strconv.ParseFloat(dcu.forecastCost, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse forecast cost: %w", err)
	}

	return &parseDailyCostUsage{
		yesterdayCost: yesterdayCost,
		actualCost:    actualCost,
		forecastCost:  forecastCost,
	}, nil
}

// calcDailyCostInJPY: Open Exchange Rates APIのレスポンスから1$あたりの円を取得し、そのレートを使用して利用コストをUSDからJPYに変換
func (pdcu *parseDailyCostUsage) calcDailyCostInJPY(res *exchange_rates.ExchangeRatesResponse) (*dailyCostUsage, error) {
	rate, ok := res.Rates[exchange_rates.JPY.String()]
	if !ok {
		return nil, fmt.Errorf("JPY exchange rate not found in the response: %+v", res.Rates)
	}

	return &dailyCostUsage{
		yesterdayCost: fmt.Sprintf("%.2f", roundUpToTwoDecimalPlaces(pdcu.yesterdayCost*rate)), // 昨日利用したコスト
		actualCost:    fmt.Sprintf("%.2f", roundUpToTwoDecimalPlaces(pdcu.actualCost*rate)),    // 本日時点で利用した総コスト
		forecastCost:  fmt.Sprintf("%.2f", roundUpToTwoDecimalPlaces(pdcu.forecastCost*rate)),  // 残り日数を考慮した今月の利用コスト
	}, nil
}

// genSlackMessage: 日次利用コストレポートのメッセージを生成
func (r dailyCostUsage) genSlackMessage() slack.Attachment {
	return slack.Attachment{
		Pretext: fmt.Sprintf(`
• 昨日の利用コスト: %s 円
• 本日時点での今月の利用コスト: %s 円
• 今月の利用コストの予測値: %s 円
`, r.yesterdayCost, r.actualCost, r.forecastCost,
		),
	}
}

// NOTE: debug用途のログ
func (fd formattedDateForDailyReport) formattedDateLogs(ctx context.Context) {
	slog.InfoContext(ctx, "[1]. formatted date:",
		slog.String("昨日の日付", fd.yesterday),   // 2024-12-28
		slog.String("今月の開始日付", fd.startDate), // 2024-12-01
		slog.String("今月の終了日付", fd.endDate),   // 2024-12-29
		slog.Int("今日までの日数", fd.currentDay),   // 29
		slog.Int("今月の総日数", fd.daysInMonth),   // 31
	)
}

func getUsageCostLogs(ctx context.Context, yesterdayCost, actualCost, forecastCost string) {
	slog.InfoContext(ctx, "[2] get usage cost",
		slog.String("yesterday", yesterdayCost), // 0.0217344233
		slog.String("actual", actualCost),       // 0.7277853673
		slog.String("forecast", forecastCost),   // 0.78
	)
}

func exchangeRatesResponseLogs(ctx context.Context, r *exchange_rates.ExchangeRatesResponse) {
	slog.InfoContext(ctx, "[3]. get exchange rates api response",
		slog.Float64("EUR", r.Rates["EUR"]), // 0.966185
		slog.Float64("JPY", r.Rates["JPY"]), // 157.35784932
	)
}

func parseJPYCostLogs(ctx context.Context, yesterdayCostJPY, actualCostJPY, forecastCostJPY string) {
	slog.InfoContext(ctx, "[4] parsed jpy cost",
		slog.String("yesterday", yesterdayCostJPY), // 3.4200821066984974
		slog.String("actual", actualCostJPY),       // 114.52274016489426
		slog.String("forecast", forecastCostJPY),   // 122.73912246960002
	)
}
