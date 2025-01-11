package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tamaco489/cost_explorer/batch/internal/configuration"
	"github.com/tamaco489/cost_explorer/batch/internal/library/debug_log"
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

	ratesResponse, err := j.exchangeRatesClient.GetExchangeRates(ctx, pxr.BaseCurrencyCode, pxr.ExchangeCurrencyCodes)
	if err != nil {
		return fmt.Errorf("failed to get exchange rates: %w", err)
	}

	if configuration.Get().Logging == "on" {
		debug_log.ExchangeRatesResponseLogs(ctx, ratesResponse)
	}

	// ************************* 4. 取得した為替レートを利用して、利用コストをUSDからJPYに変換 *************************
	costUsage := j.dailyCostExplorerService.NewDailyCostUsage(yesterdayCost, actualCost, forecastCost)
	jpyUsage, err := costUsage.CalcDailyCostInJPY(ratesResponse)
	if err != nil {
		return err
	}

	if configuration.Get().Logging == "on" {
		debug_log.DailyParseJPYCostLogs(ctx, jpyUsage.YesterdayCost, jpyUsage.ActualCost, jpyUsage.ForecastCost)
	}

	// ************************* 5. Slackにメッセージを送信する *************************
	message := jpyUsage.GenDailySlackMessage()
	sc := slack.NewSlackClient(configuration.Get().Slack.DailyWebHookURL, configuration.Get().ServiceName)
	if err := sc.SendMessage(ctx, slack.DailyReportTitle.String(), message); err != nil {
		return fmt.Errorf("failed to send slack message: %w", err)
	}

	return nil
}
