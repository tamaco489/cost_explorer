package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/tamaco489/cost_explorer/batch/internal/configuration"
	"github.com/tamaco489/cost_explorer/batch/internal/library/exchange_rates"
	"github.com/tamaco489/cost_explorer/batch/internal/library/slack"
)

func (j *Job) DailyCostReport(ctx context.Context) error {

	// ************************* 1. 実行日時からコスト算出に必要な各基準日を取得 (月初の場合は処理をスキップ) *************************
	slog.InfoContext(ctx, "DailyCostReport",
		slog.String("date (jst)", j.execTimeJST.Format("2006-01-02 15:04:05 MST")),
	)

	if j.execTimeJST.Day() == 1 {
		slog.InfoContext(ctx, "no processing is performed at the beginning of the month",
			slog.String("date (jst)", j.execTimeJST.Format("2006-01-02 15:04:05 MST")),
		)
		return nil
	}

	fd := newFormatDateForDailyReport(j.execTimeJST)

	if configuration.Get().Logging == "on" {
		fd.formattedDateLogs(ctx)
	}

	// ************************* 2. AWS 利用コストの算出 *************************
	yesterdayCost, err := j.dailyCostExplorerService.GetYesterdayCost(ctx, fd.yesterday, fd.endDate)
	if err != nil {
		return fmt.Errorf("failed to get yesterday cost: %w", err)
	}

	actualCost, err := j.dailyCostExplorerService.GetActualCost(ctx, fd.startDate, fd.endDate)
	if err != nil {
		return fmt.Errorf("failed to get actual cost: %w", err)
	}

	forecastCost, err := j.dailyCostExplorerService.GetForecastCost(ctx, actualCost, fd.currentDay, fd.daysInMonth)
	if err != nil {
		return fmt.Errorf("failed to get forecast cost: %w", err)
	}

	if configuration.Get().Logging == "on" {
		fd.getDailyUsageCostLogs(ctx, yesterdayCost, actualCost, forecastCost)
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
		fd.exchangeRatesResponseLogs(ctx, ratesResponse)
	}

	// ************************* 4. 取得した為替レートを利用して、利用コストをUSDからJPYに変換 *************************
	costUsage := newDailyCostUsage(yesterdayCost, actualCost, forecastCost)
	jpyUsage, err := costUsage.calcDailyCostInJPY(ratesResponse)
	if err != nil {
		return err
	}

	if configuration.Get().Logging == "on" {
		fd.parseJPYCostLogs(ctx, jpyUsage.yesterdayCost, jpyUsage.actualCost, jpyUsage.forecastCost)
	}

	// ************************* 5. Slackにメッセージを送信する *************************
	message := jpyUsage.genSlackMessage()
	sc := slack.NewSlackClient(configuration.Get().Slack.DailyWebHookURL, configuration.Get().ServiceName)
	if err := sc.SendMessage(ctx, slack.DailyReportTitle.String(), message); err != nil {
		return fmt.Errorf("failed to send slack message: %w", err)
	}

	return nil
}

// formatDateForDailyReport: 日次コストレポートのための日時情報を保持する構造体
type formatDateForDailyReport struct {
	yesterday   string // 昨日の日付
	startDate   string // 今月の開始日付
	endDate     string // 今月の終了日付
	currentDay  int    // 今日までの日数
	daysInMonth int    // 今月の総日数
}

// newFormatDateForDailyReport: formatDateForDailyReport のコンストラクタ
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
func newFormatDateForDailyReport(execTime time.Time) formatDateForDailyReport {
	currentYear, currentMonth, _ := execTime.Date()
	daysInMonth := time.Date(currentYear, currentMonth+1, 0, 0, 0, 0, 0, time.UTC).Day()

	return formatDateForDailyReport{
		yesterday:   execTime.AddDate(0, 0, -1).Format("2006-01-02"),
		startDate:   time.Date(execTime.Year(), execTime.Month(), 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02"),
		endDate:     execTime.Format("2006-01-02"),
		currentDay:  execTime.Day(),
		daysInMonth: daysInMonth,
	}
}

// dailyCostUsage: 日次レポートに必要な要素を含む構造体。
type dailyCostUsage struct {
	yesterdayCost float64
	actualCost    float64
	forecastCost  float64
}

// newDailyCostUsage: dailyCostUsage のコンストラクタ
func newDailyCostUsage(yesterdayCost, actualCost, forecastCost float64) *dailyCostUsage {
	return &dailyCostUsage{
		yesterdayCost: yesterdayCost,
		actualCost:    actualCost,
		forecastCost:  forecastCost,
	}
}

// calcDailyCostInJPY: Open Exchange Rates APIのレスポンスから1$あたりの円を取得し、そのレートを使用して利用コストをUSDからJPYに変換
func (pdcu *dailyCostUsage) calcDailyCostInJPY(res *exchange_rates.ExchangeRatesResponse) (*dailyCostUsage, error) {
	rate, ok := res.Rates[exchange_rates.JPY.String()]
	if !ok {
		return nil, fmt.Errorf("JPY exchange rate not found in the response: %+v", res.Rates)
	}

	return &dailyCostUsage{
		yesterdayCost: roundUpToTwoDecimalPlaces(pdcu.yesterdayCost * rate), // 昨日利用したコスト
		actualCost:    roundUpToTwoDecimalPlaces(pdcu.actualCost * rate),    // 本日時点で利用した総コスト
		forecastCost:  roundUpToTwoDecimalPlaces(pdcu.forecastCost * rate),  // 残り日数を考慮した今月の利用コスト
	}, nil
}

// genSlackMessage: 日次利用コストレポートのメッセージを生成
func (r dailyCostUsage) genSlackMessage() slack.Attachment {
	return slack.Attachment{
		Pretext: fmt.Sprintf(`
• 昨日の利用コスト: %.2f 円
• 本日時点での今月の利用コスト: %.2f 円
• 今月の利用コストの予測値: %.2f 円
`, r.yesterdayCost, r.actualCost, r.forecastCost,
		),
	}
}

// NOTE: debug用途のログ
func (fd formatDateForDailyReport) formattedDateLogs(ctx context.Context) {
	slog.InfoContext(ctx, "[1]. formatted date",
		slog.String("昨日の日付", fd.yesterday),   // 2024-12-28
		slog.String("今月の開始日付", fd.startDate), // 2024-12-01
		slog.String("今月の終了日付", fd.endDate),   // 2024-12-29
		slog.Int("今日までの日数", fd.currentDay),   // 29
		slog.Int("今月の総日数", fd.daysInMonth),   // 31
	)
}

func (fd formatDateForDailyReport) getDailyUsageCostLogs(ctx context.Context, yesterdayCost, actualCost, forecastCost float64) {
	slog.InfoContext(ctx, "[2] get daily cost usage",
		slog.Float64("yesterday", yesterdayCost), // 0.0217344233
		slog.Float64("actual", actualCost),       // 0.7277853673
		slog.Float64("forecast", forecastCost),   // 0.78
	)
}

func (fd formatDateForDailyReport) exchangeRatesResponseLogs(ctx context.Context, r *exchange_rates.ExchangeRatesResponse) {
	slog.InfoContext(ctx, "[3]. get exchange rates api response",
		slog.Float64("JPY", r.Rates["JPY"]), // 157.35784932
	)
}

func (fd formatDateForDailyReport) parseJPYCostLogs(ctx context.Context, yesterdayCostJPY, actualCostJPY, forecastCostJPY float64) {
	slog.InfoContext(ctx, "[4] parsed jpy cost",
		slog.Float64("yesterday", yesterdayCostJPY), // 3.4200821066984974
		slog.Float64("actual", actualCostJPY),       // 114.52274016489426
		slog.Float64("forecast", forecastCostJPY),   // 122.73912246960002
	)
}
