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

func (j *Job) WeeklyCostReport(ctx context.Context) error {

	// ************************* 1. 実行日時からコスト算出に必要な各基準日を取得 *************************
	slog.InfoContext(ctx, "WeeklyCostReport",
		slog.String("date (jst)", j.execTimeJST.Format("2006-01-02 15:04:05 MST")),
	)

	fd := j.weeklyCostExplorerService.NewWeeklyReportDateFormatter(j.execTimeJST)
	if configuration.Get().Logging == "on" {
		debug_log.FormatDateForWeeklyReportLogs(ctx, fd)
	}

	// ************************* 2. AWS 利用コストの算出 *************************
	lastWeekCost, err := j.weeklyCostExplorerService.GetLastWeekCost(ctx, fd.LastWeekStartDate, fd.LastWeekEndDate)
	if err != nil {
		return fmt.Errorf("failed to get last week cost: %w", err)
	}

	weekBeforeLastCost, err := j.weeklyCostExplorerService.GetWeekBeforeLastCost(ctx, fd.WeekBeforeLastStartDate, fd.WeekBeforeLastEndDate)
	if err != nil {
		return fmt.Errorf("failed to get week before last cost: %w", err)
	}

	percentageChange, err := j.weeklyCostExplorerService.CalcPercentageChange(ctx, lastWeekCost, weekBeforeLastCost)
	if err != nil {
		return err
	}

	if configuration.Get().Logging == "on" {
		debug_log.WeeklyUsageCostLogs(ctx, lastWeekCost, weekBeforeLastCost, percentageChange)
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
	costUsage := newWeeklyCostUsage(lastWeekCost, weekBeforeLastCost, percentageChange)
	jpyUsage, err := costUsage.calcWeeklyCostInJPY(ratesResponse)
	if err != nil {
		return err
	}

	if configuration.Get().Logging == "on" {
		debug_log.WeeklyParseJPYCostLogs(ctx, jpyUsage.lastWeekCost, jpyUsage.weekBeforeLastCost)
	}

	// note: 値の受け渡しを行う
	jpyUsage.percentageChange = costUsage.percentageChange

	// ************************* 5. Slackにメッセージを送信する *************************
	message := jpyUsage.genSlackMessage()
	sc := slack.NewSlackClient(configuration.Get().Slack.WeeklyWebHookURL, configuration.Get().ServiceName)
	if err := sc.SendMessage(ctx, slack.WeeklyReportTitle.String(), message); err != nil {
		return fmt.Errorf("failed to send slack message: %w", err)
	}

	return nil
}

// weeklyCostUsage: 週次レポートに必要な要素を含む構造体
type weeklyCostUsage struct {
	lastWeekCost       float64
	weekBeforeLastCost float64
	percentageChange   float64
}

// newWeeklyCostUsage: weeklyCostUsage のコンストラクタ
func newWeeklyCostUsage(lastWeekCost, weekBeforeLastCost, percentageChange float64) *weeklyCostUsage {
	return &weeklyCostUsage{
		lastWeekCost:       lastWeekCost,
		weekBeforeLastCost: weekBeforeLastCost,
		percentageChange:   percentageChange,
	}
}

// calcWeeklyCostInJPY: Open Exchange Rates APIのレスポンスから1$あたりの円を取得し、そのレートを使用して利用コストをUSDからJPYに変換
func (wcu *weeklyCostUsage) calcWeeklyCostInJPY(res *exchange_rates.ExchangeRatesResponse) (*weeklyCostUsage, error) {
	rate, ok := res.Rates[exchange_rates.JPY.String()]
	if !ok {
		return nil, fmt.Errorf("JPY exchange rate not found in the response: %+v", res.Rates)
	}

	return &weeklyCostUsage{
		lastWeekCost:       roundUpToTwoDecimalPlaces(wcu.lastWeekCost * rate),       // 先週利用したコスト
		weekBeforeLastCost: roundUpToTwoDecimalPlaces(wcu.weekBeforeLastCost * rate), // 先々週利用した総コスト
		percentageChange:   wcu.percentageChange,                                     // 先週と先々週のコスト増減（%）
	}, nil
}

// genSlackMessage: 週次利用コストレポートのメッセージを生成する
func (wcu *weeklyCostUsage) genSlackMessage() slack.Attachment {
	return slack.Attachment{
		Pretext: fmt.Sprintf(`
• 先週の利用コスト: %.2f 円
• 先々週の利用コスト: %.2f 円
• 先週と先々週のコスト増減: %.2f %%`,
			wcu.lastWeekCost, wcu.weekBeforeLastCost, wcu.percentageChange,
		),
	}
}
