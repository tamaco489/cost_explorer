package service

import (
	"context"
	"fmt"
	"strconv"

	cost_explorer "github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
)

type WeeklyCostExplorerService struct {
	client *cost_explorer.Client
}

func NewWeeklyCostExplorerService(client *cost_explorer.Client) *WeeklyCostExplorerService {
	return &WeeklyCostExplorerService{client: client}
}

// getLastWeekCost: 先週の利用コストを取得
func (s *WeeklyCostExplorerService) GetLastWeekCost(ctx context.Context, lastWeekStartDate, lastWeekEndDate string) (float64, error) {

	output, err := s.client.GetCostAndUsage(ctx, &cost_explorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: &lastWeekStartDate,
			End:   &lastWeekEndDate,
		},
		Metrics:     []string{"UnblendedCost"},
		Granularity: types.GranularityDaily,
	})
	if err != nil {
		return 0, err
	}

	totalCost := 0.0
	for _, result := range output.ResultsByTime {
		if cost, ok := result.Total["UnblendedCost"]; ok {
			amount, err := s.parseFloat(*cost.Amount)
			if err != nil {
				return 0, err
			}
			totalCost += amount
		}
	}

	return totalCost, nil
}

// getWeekBeforeLastCost: 先々週の利用コストを取得
func (s *WeeklyCostExplorerService) GetWeekBeforeLastCost(ctx context.Context, weekBeforeLastStartDate, weekBeforeLastEndDate string) (float64, error) {

	output, err := s.client.GetCostAndUsage(ctx, &cost_explorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: &weekBeforeLastStartDate,
			End:   &weekBeforeLastEndDate,
		},
		Metrics:     []string{"UnblendedCost"},
		Granularity: types.GranularityDaily,
	})
	if err != nil {
		return 0, err
	}

	totalCost := 0.0
	for _, result := range output.ResultsByTime {
		if cost, ok := result.Total["UnblendedCost"]; ok {
			amount, err := strconv.ParseFloat(*cost.Amount, 64)
			if err != nil {
				return 0, err
			}
			totalCost += amount
		}
	}

	return totalCost, nil
}

// calcPercentageChange: コストの増減率を算出
func (s *WeeklyCostExplorerService) CalcPercentageChange(ctx context.Context, lastWeekCost, weekBeforeLastCost float64) (float64, error) {

	if weekBeforeLastCost == 0 {
		return 0, fmt.Errorf("week before last cost is zero")
	}

	change := ((lastWeekCost - weekBeforeLastCost) / weekBeforeLastCost) * 100

	return change, nil
}

func (s *WeeklyCostExplorerService) parseFloat(value string) (float64, error) {
	return strconv.ParseFloat(value, 64)
}
