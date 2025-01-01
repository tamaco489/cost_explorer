package usecase

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/tamaco489/cost_explorer/batch/internal/configuration"
	"github.com/tamaco489/cost_explorer/batch/internal/library/exchange_rates"
	"github.com/tamaco489/cost_explorer/batch/internal/library/slack"
)

func (j *Job) WeeklyCostReport(ctx context.Context) error {

	// ************************* 1. 実行日時からコスト算出に必要な各基準日を取得 *************************
	fd := newFormattedDateForWeeklyReport(j.execTime)

	if configuration.Get().Logging == "on" {
		fd.formattedDateLogs(ctx)
	}

	// ************************* 2. AWS 利用コストの算出 *************************
	lastWeekCost, err := j.getLastWeekCost(ctx, fd.lastWeekStartDate, fd.lastWeekEndDate)
	if err != nil {
		return fmt.Errorf("failed to get last week cost: %w", err)
	}

	weekBeforeLastCost, err := j.getWeekBeforeLastCost(ctx, fd.weekBeforeLastStartDate, fd.weekBeforeLastEndDate)
	if err != nil {
		return fmt.Errorf("failed to get week before last cost: %w", err)
	}

	percentageChange := j.calcPercentageChange(lastWeekCost, weekBeforeLastCost)

	if configuration.Get().Logging == "on" {
		fd.getWeeklyUsageCostLogs(ctx, lastWeekCost, weekBeforeLastCost, percentageChange)
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
	costUsage := newWeeklyCostUsage(lastWeekCost, weekBeforeLastCost, percentageChange)
	parseUsage, err := costUsage.parseWeeklyCostUsage()
	if err != nil {
		return err
	}

	jpyUsage, err := parseUsage.calcWeeklyCostInJPY(ratesResponse)
	if err != nil {
		return err
	}

	if configuration.Get().Logging == "on" {
		fd.parseJPYCostLogs(ctx, jpyUsage.lastWeekCost, jpyUsage.weekBeforeLastCost)
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

// formattedDateForWeeklyReport: 週次コストレポートのための日時情報を保持する構造体
type formattedDateForWeeklyReport struct {
	lastWeekStartDate       string // 先週の開始日付
	lastWeekEndDate         string // 先週の終了日付
	weekBeforeLastStartDate string // 先々週の開始日付
	weekBeforeLastEndDate   string // 先々週の終了日付
}

// newFormattedDateForWeeklyReport: formattedDateForWeeklyReport のコンストラクタ
//
// lastWeekStartDate: 先週の開始日付 (string)
//
// lastWeekEndDate: 先週の終了日付 (string)
//
// weekBeforeLastStartDate: 先々週の開始日付 (string)
//
// weekBeforeLastEndDate: 先々週の終了日付 (string)
func newFormattedDateForWeeklyReport(execTime time.Time) formattedDateForWeeklyReport {
	return formattedDateForWeeklyReport{
		lastWeekStartDate:       execTime.AddDate(0, 0, -13).Format("2006-01-02"),
		lastWeekEndDate:         execTime.AddDate(0, 0, -7).Format("2006-01-02"),
		weekBeforeLastStartDate: execTime.AddDate(0, 0, -20).Format("2006-01-02"),
		weekBeforeLastEndDate:   execTime.AddDate(0, 0, -14).Format("2006-01-02"),
	}
}

// getLastWeekCost: 先週の利用コストを取得
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

// getWeekBeforeLastCost: 先々週の利用コストを取得
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

// calcPercentageChange: コストの増減率を算出
func (j *Job) calcPercentageChange(lastWeekCost, weekBeforeLastCost string) string {

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

// weeklyCostUsage: 週次レポートに必要な要素を含む構造体
//
// AWS Cost Explorer SDKの戻り値が全てstring型のため、各パラメータは全てstring型として定義
type weeklyCostUsage struct {
	lastWeekCost       string
	weekBeforeLastCost string
	percentageChange   string
}

// newWeeklyCostUsage: weeklyCostUsage のコンストラクタ
func newWeeklyCostUsage(lastWeekCost, weekBeforeLastCost, percentageChange string) *weeklyCostUsage {
	return &weeklyCostUsage{
		lastWeekCost:       lastWeekCost,
		weekBeforeLastCost: weekBeforeLastCost,
		percentageChange:   percentageChange,
	}
}

// parseWeeklyCostUsage: weeklyCostUsage の各パラメータをfloat64型にした構造体
//
// 為替レートに計算において、小数点を考慮する必要があるため、float型として定義
type parseWeeklyCostUsage struct {
	lastWeekCost       float64
	weekBeforeLastCost float64
	percentageChange   float64
}

// parseWeeklyCostUsage: 利用コストをfloat64型に変換
func (wcu *weeklyCostUsage) parseWeeklyCostUsage() (*parseWeeklyCostUsage, error) {
	lastWeekCost, err := strconv.ParseFloat(wcu.lastWeekCost, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse last week cost to float64: %w", err)
	}

	weekBeforeLastCost, err := strconv.ParseFloat(wcu.weekBeforeLastCost, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse week before last cost to float64: %w", err)
	}

	percentageChange, err := strconv.ParseFloat(wcu.percentageChange, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse percentage change: %w", err)
	}

	return &parseWeeklyCostUsage{
		lastWeekCost:       lastWeekCost,
		weekBeforeLastCost: weekBeforeLastCost,
		percentageChange:   percentageChange,
	}, nil
}

// calcWeeklyCostInJPY: Open Exchange Rates APIのレスポンスから1$あたりの円を取得し、そのレートを使用して利用コストをUSDからJPYに変換
func (pwcu *parseWeeklyCostUsage) calcWeeklyCostInJPY(res *exchange_rates.ExchangeRatesResponse) (*weeklyCostUsage, error) {
	rate, ok := res.Rates[exchange_rates.JPY.String()]
	if !ok {
		return nil, fmt.Errorf("JPY exchange rate not found in the response: %+v", res.Rates)
	}

	return &weeklyCostUsage{
		lastWeekCost:       fmt.Sprintf("%.2f", roundUpToTwoDecimalPlaces(pwcu.lastWeekCost*rate)),       // 先週利用したコスト
		weekBeforeLastCost: fmt.Sprintf("%.2f", roundUpToTwoDecimalPlaces(pwcu.weekBeforeLastCost*rate)), // 先々週利用した総コスト
		percentageChange:   fmt.Sprintf("%.2f", pwcu.percentageChange),                                   // 先週と先々週のコスト増減（%）
	}, nil
}

// genSlackMessage: 週次利用コストレポートのメッセージを生成する
func (r weeklyCostUsage) genSlackMessage() slack.Attachment {
	return slack.Attachment{
		Pretext: fmt.Sprintf(`
• 先週の利用コスト: %s 円
• 先々週の利用コスト: %s 円
• 先週と先々週のコスト増減: %s %%`,
			r.lastWeekCost, r.weekBeforeLastCost, r.percentageChange,
		),
	}
}

// NOTE: debug用途のログ
func (fd formattedDateForWeeklyReport) formattedDateLogs(ctx context.Context) {
	slog.InfoContext(ctx, "[1] formatted date",
		slog.String("先週の開始日付", fd.lastWeekStartDate),        // 2024-12-19
		slog.String("先週の終了日付", fd.lastWeekEndDate),          // 2024-12-25
		slog.String("先々週の開始日付", fd.weekBeforeLastStartDate), // 2024-12-12
		slog.String("先々週の終了日付", fd.weekBeforeLastEndDate),   // 2024-12-18
	)
}

func (fd formattedDateForWeeklyReport) getWeeklyUsageCostLogs(ctx context.Context, lastWeekCost, weekBeforeLastCost, percentageChange string) {
	slog.InfoContext(ctx, "[2] get usage cost",
		slog.String("last week cost", lastWeekCost),                              // 0.03
		slog.String("week before last cost", weekBeforeLastCost),                 // 0.03
		slog.String("percentage change", fmt.Sprintf("%s %%", percentageChange)), // 0.00 %
	)
}

func (fd formattedDateForWeeklyReport) exchangeRatesResponseLogs(ctx context.Context, r *exchange_rates.ExchangeRatesResponse) {
	slog.InfoContext(ctx, "[3]. get exchange rates api response",
		slog.Float64("JPY", r.Rates["JPY"]), // 157.35784932
	)
}

func (fd formattedDateForWeeklyReport) parseJPYCostLogs(ctx context.Context, lastWeekCost, weekBeforeLastCost string) {
	slog.InfoContext(ctx, "[4] parsed jpy cost",
		slog.String("last week cost", lastWeekCost),              // 4.73
		slog.String("week before last cost", weekBeforeLastCost), // 4.73
	)
}
