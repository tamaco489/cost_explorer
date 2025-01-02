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

func DailyUsageCostLogs(ctx context.Context, yesterdayCost, actualCost, forecastCost float64) {
	slog.InfoContext(ctx, "[2] get daily cost usage",
		slog.Float64("yesterday", yesterdayCost), // 0.0217344233
		slog.Float64("actual", actualCost),       // 0.7277853673
		slog.Float64("forecast", forecastCost),   // 0.78
	)
}

func ExchangeRatesResponseLogs(ctx context.Context, r *exchange_rates.ExchangeRatesResponse) {
	slog.InfoContext(ctx, "[3]. get exchange rates api response",
		slog.Float64("JPY", r.Rates["JPY"]), // 157.35784932
	)
}

func ParseJPYCostLogs(ctx context.Context, yesterdayCostJPY, actualCostJPY, forecastCostJPY float64) {
	slog.InfoContext(ctx, "[4] parsed jpy cost",
		slog.Float64("yesterday", yesterdayCostJPY), // 3.4200821066984974
		slog.Float64("actual", actualCostJPY),       // 114.52274016489426
		slog.Float64("forecast", forecastCostJPY),   // 122.73912246960002
	)
}
