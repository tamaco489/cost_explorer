package debug_log

import (
	"context"
	"log/slog"

	"github.com/tamaco489/cost_explorer/batch/internal/library/exchange_rates"
	"github.com/tamaco489/cost_explorer/batch/internal/service"
)

// NOTE: debug用途のログ
func FormatDateForDailyReportLogs(ctx context.Context, fd service.DailyReportDateFormatter) {
	slog.InfoContext(ctx, "[1]. formatted date",
		slog.String("昨日の日付", fd.Yesterday),   // 2024-12-28
		slog.String("今月の開始日付", fd.StartDate), // 2024-12-01
		slog.String("今月の終了日付", fd.EndDate),   // 2024-12-29
		slog.Int("今日までの日数", fd.CurrentDay),   // 29
		slog.Int("今月の総日数", fd.DaysInMonth),   // 31
	)
}

func FormatDateForWeeklyReportLogs(ctx context.Context, wd service.WeeklyReportDateFormatter) {
	slog.InfoContext(ctx, "[1] formatted date",
		slog.String("先週の開始日付", wd.LastWeekStartDate),        // 2024-12-19
		slog.String("先週の終了日付", wd.LastWeekEndDate),          // 2024-12-25
		slog.String("先々週の開始日付", wd.WeekBeforeLastStartDate), // 2024-12-12
		slog.String("先々週の終了日付", wd.WeekBeforeLastEndDate),   // 2024-12-18
	)
}

func DailyUsageCostLogs(ctx context.Context, yesterdayCost, actualCost, forecastCost float64) {
	slog.InfoContext(ctx, "[2] get daily cost usage",
		slog.Float64("yesterday", yesterdayCost), // 0.0217344233
		slog.Float64("actual", actualCost),       // 0.7277853673
		slog.Float64("forecast", forecastCost),   // 0.78
	)
}

func WeeklyUsageCostLogs(ctx context.Context, lastWeekCost, weekBeforeLastCost, percentageChange float64) {
	slog.InfoContext(ctx, "[2] get weekly usage cost",
		slog.Float64("last week cost", lastWeekCost),              // 0.027573460300000005
		slog.Float64("week before last cost", weekBeforeLastCost), // 0.0291809323
		slog.Float64("percentage change", percentageChange),       // -5.508638255536459
	)
}

func ExchangeRatesResponseLogs(ctx context.Context, r *exchange_rates.ExchangeRatesResponse) {
	slog.InfoContext(ctx, "[3]. get exchange rates api response",
		slog.Float64("JPY", r.Rates["JPY"]), // 157.35784932
	)
}

func DailyParseJPYCostLogs(ctx context.Context, yesterdayCostJPY, actualCostJPY, forecastCostJPY float64) {
	slog.InfoContext(ctx, "[4] parsed jpy cost",
		slog.Float64("yesterday", yesterdayCostJPY), // 3.4200821066984974
		slog.Float64("actual", actualCostJPY),       // 114.52274016489426
		slog.Float64("forecast", forecastCostJPY),   // 122.73912246960002
	)
}

func WeeklyParseJPYCostLogs(ctx context.Context, lastWeekCost, weekBeforeLastCost float64) {
	slog.InfoContext(ctx, "[4] parsed jpy cost",
		slog.Float64("last week cost", lastWeekCost),              // 4.73
		slog.Float64("week before last cost", weekBeforeLastCost), // 4.73
	)
}
