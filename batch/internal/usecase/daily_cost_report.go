package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tamaco489/cost_explorer/batch/internal/configuration"
	"github.com/tamaco489/cost_explorer/batch/internal/library/debug_log"
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

	fd := j.dailyCostExplorerService.NewDailyReportDateFormatter(j.execTimeJST)
	if configuration.Get().Logging == "on" {
		debug_log.FormatDateForDailyReportLogs(ctx, fd)
	}

	// ************************* 2. AWS 利用コストの算出 *************************
	yesterdayCost, err := j.dailyCostExplorerService.GetYesterdayCost(ctx, fd.Yesterday, fd.EndDate)
	if err != nil {
		return fmt.Errorf("failed to get yesterday cost: %w", err)
	}

	actualCost, err := j.dailyCostExplorerService.GetActualCost(ctx, fd.StartDate, fd.EndDate)
	if err != nil {
		return fmt.Errorf("failed to get actual cost: %w", err)
	}

	forecastCost, err := j.dailyCostExplorerService.GetForecastCost(ctx, actualCost, fd.CurrentDay, fd.DaysInMonth)
	if err != nil {
		return fmt.Errorf("failed to get forecast cost: %w", err)
	}

	if configuration.Get().Logging == "on" {
		debug_log.DailyUsageCostLogs(ctx, yesterdayCost, actualCost, forecastCost)
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
		debug_log.ExchangeRatesResponseLogs(ctx, ratesResponse)
	}

	// ************************* 4. 取得した為替レートを利用して、利用コストをUSDからJPYに変換 *************************
	costUsage := newDailyCostUsage(yesterdayCost, actualCost, forecastCost)
	jpyUsage, err := costUsage.calcDailyCostInJPY(ratesResponse)
	if err != nil {
		return err
	}

	if configuration.Get().Logging == "on" {
		debug_log.ParseJPYCostLogs(ctx, jpyUsage.yesterdayCost, jpyUsage.actualCost, jpyUsage.forecastCost)
	}

	// ************************* 5. Slackにメッセージを送信する *************************
	message := jpyUsage.genSlackMessage()
	sc := slack.NewSlackClient(configuration.Get().Slack.DailyWebHookURL, configuration.Get().ServiceName)
	if err := sc.SendMessage(ctx, slack.DailyReportTitle.String(), message); err != nil {
		return fmt.Errorf("failed to send slack message: %w", err)
	}

	return nil
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
