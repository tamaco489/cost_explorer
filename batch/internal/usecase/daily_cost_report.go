package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
)

func (j *Job) DailyCostReport(ctx context.Context) error {

	log.Println("DailyCostReport job started...")

	// 昨日の利用コストを取得
	yesterdayCost, err := j.getYesterdayCost(ctx)
	if err != nil {
		return fmt.Errorf("failed to get yesterday cost: %w", err)
	}
	log.Printf("昨日の利用コスト: %s USD\n", yesterdayCost)

	// 今月の利用コスト
	actualCost, err := j.getCost(ctx)
	if err != nil {
		return fmt.Errorf("failed to get actual cost: %w", err)
	}

	log.Printf("今月の利用コスト（本日までの実績）: %s USD\n", actualCost)

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

// 今月の実績を取得する
func (j *Job) getCost(ctx context.Context) (string, error) {

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
