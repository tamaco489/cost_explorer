package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tamaco489/cost_explorer/batch/internal/configuration"
	"github.com/tamaco489/cost_explorer/batch/internal/library/debug_log"
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

	ratesResponse, err := j.exchangeRatesClient.GetExchangeRates(ctx, pxr.BaseCurrencyCode, pxr.ExchangeCurrencyCodes)
	if err != nil {
		return fmt.Errorf("failed to get exchange rates: %w", err)
	}

	if configuration.Get().Logging == "on" {
		debug_log.ExchangeRatesResponseLogs(ctx, ratesResponse)
	}

	// ************************* 4. 取得した為替レートを利用して、利用コストをUSDからJPYに変換 *************************
	costUsage := j.weeklyCostExplorerService.NewWeeklyCostUsage(lastWeekCost, weekBeforeLastCost, percentageChange)
	jpyUsage, err := costUsage.CalcWeeklyCostInJPY(ratesResponse)
	if err != nil {
		return err
	}

	if configuration.Get().Logging == "on" {
		debug_log.WeeklyParseJPYCostLogs(ctx, jpyUsage.LastWeekCost, jpyUsage.WeekBeforeLastCost)
	}

	// ************************* 5. Slackにメッセージを送信する *************************
	message := jpyUsage.GenWeeklySlackMessage()
	sc := slack.NewSlackClient(configuration.Get().Slack.WeeklyWebHookURL, configuration.Get().ServiceName)
	if err := sc.SendMessage(ctx, slack.WeeklyReportTitle.String(), message); err != nil {
		return fmt.Errorf("failed to send slack message: %w", err)
	}

	return nil
}
