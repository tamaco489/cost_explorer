package usecase

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
)

func (j *Job) DailyCostReport(ctx context.Context) error {
	// 昨日の利用コスト
	yesterdayCost, err := j.getYesterdayCost(ctx)
	if err != nil {
		return fmt.Errorf("failed to get yesterday cost: %w", err)
	}

	// 本日時点での今月の利用コスト
	actualCost, err := j.getActualCost(ctx)
	if err != nil {
		return fmt.Errorf("failed to get actual cost: %w", err)
	}

	// 今月の利用コストの予測値
	forecastCost, err := j.getForecastCost(actualCost)
	if err != nil {
		return fmt.Errorf("failed to get forecast cost: %w", err)
	}

	log.Printf("昨日の利用コスト: %s USD\n", yesterdayCost)
	log.Printf("本日時点での今月の利用コスト: %s USD\n", actualCost)
	log.Printf("今月の利用コストの予測値: %s USD\n", forecastCost)

	return nil
}

// 昨日の利用コストを取得する
func (j *Job) getYesterdayCost(ctx context.Context) (string, error) {
	// 昨日の開始日と終了日を計算
	yesterday := j.execTime.AddDate(0, 0, -1).Format("2006-01-02")
	endDate := j.execTime.Format("2006-01-02")

	log.Println("[1] getYesterdayCost |", "昨日:", yesterday, "現在日:", endDate)

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

// 本日時点での今月の利用コストを取得する
func (j *Job) getActualCost(ctx context.Context) (string, error) {
	// 今月の開始日と現在日
	startDate := time.Date(j.execTime.Year(), j.execTime.Month(), 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	endDate := j.execTime.Format("2006-01-02")

	log.Println("[2] getCost |", "今月の開始日:", startDate, "現在日:", endDate)

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

// 今月の利用コストの予測値を算出する
func (j *Job) getForecastCost(actualCost string) (string, error) {
	// 今月の総日数
	currentYear, currentMonth, _ := j.execTime.Date()
	daysInMonth := time.Date(currentYear, currentMonth+1, 0, 0, 0, 0, 0, time.UTC).Day()

	// 今日までの日数
	currentDay := j.execTime.Day()

	// コスト計算
	actualCostFloat, err := parseCost(actualCost)
	if err != nil {
		return "", err
	}

	// 1日あたりの平均コスト
	averageCostPerDay := actualCostFloat / float64(currentDay)

	// 予測コスト
	forecastCost := averageCostPerDay * float64(daysInMonth)

	log.Println("[3] getForecastCost |", "今月の総日数:", daysInMonth, "今日までの日数:", currentDay)
	log.Println("[4] getForecastCost |", "1日あたりの平均コスト:", averageCostPerDay, "予測コスト:", forecastCost)

	return fmt.Sprintf("%.2f", forecastCost), nil
}

// parseCost はコスト文字列を float64 に変換する
func parseCost(cost string) (float64, error) {
	return strconv.ParseFloat(cost, 64)
}
