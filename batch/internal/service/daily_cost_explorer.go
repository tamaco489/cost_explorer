package service

import (
	"context"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"

	cost_explorer "github.com/aws/aws-sdk-go-v2/service/costexplorer"
)

type IDailyCostExplorerClient interface {
	GetYesterdayCost(ctx context.Context, yesterday, endDate string) (float64, error)
	GetActualCost(ctx context.Context, startDate, endDate string) (float64, error)
	GetForecastCost(ctx context.Context, actualCost float64, currentDay, daysInMonth int) (float64, error)
}

var _ IDailyCostExplorerClient = (*DailyCostExplorerService)(nil)

type DailyCostExplorerService struct {
	client *cost_explorer.Client
}

func NewDailyCostExplorerService(client *cost_explorer.Client) *DailyCostExplorerService {
	return &DailyCostExplorerService{client: client}
}

// GetYesterdayCost: 昨日の利用コストを取得
func (s *DailyCostExplorerService) GetYesterdayCost(ctx context.Context, yesterday, endDate string) (float64, error) {

	input := &cost_explorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: &yesterday,
			End:   &endDate,
		},
		Granularity: types.GranularityDaily,
		Metrics:     []string{"UnblendedCost"},
	}

	output, err := s.client.GetCostAndUsage(ctx, input)
	if err != nil {
		return 0, err
	}

	if len(output.ResultsByTime) > 0 && len(output.ResultsByTime[0].Total) > 0 {
		cost, ok := output.ResultsByTime[0].Total["UnblendedCost"]
		if ok && cost.Amount != nil {
			return s.parseFloat(*cost.Amount)
		}
	}

	return 0, nil
}

// GetActualCost: 本日時点での今月の利用コストを取得
func (s *DailyCostExplorerService) GetActualCost(ctx context.Context, startDate, endDate string) (float64, error) {
	input := &cost_explorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: &startDate,
			End:   &endDate,
		},
		Granularity: types.GranularityDaily,
		Metrics:     []string{"UnblendedCost"},
	}

	output, err := s.client.GetCostAndUsage(ctx, input)
	if err != nil {
		return 0, err
	}

	totalCost := 0.0
	for _, result := range output.ResultsByTime {
		if cost, ok := result.Total["UnblendedCost"]; ok && cost.Amount != nil {
			amount, err := s.parseFloat(*cost.Amount)
			if err != nil {
				return 0, err
			}
			totalCost += amount
		}
	}

	return totalCost, nil
}

// GetForecastCost: 今月の利用コストの予測値を算出
func (s *DailyCostExplorerService) GetForecastCost(ctx context.Context, actualCost float64, currentDay, daysInMonth int) (float64, error) {

	// 1日あたりの平均コスト
	averageCostPerDay := actualCost / float64(currentDay)

	// 予測コスト
	forecastCost := averageCostPerDay * float64(daysInMonth)

	return forecastCost, nil
}

func (s *DailyCostExplorerService) parseFloat(value string) (float64, error) {
	return strconv.ParseFloat(value, 64)
}
